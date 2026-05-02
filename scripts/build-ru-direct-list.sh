#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_URL="${SOURCE_URL:-https://raw.githubusercontent.com/itdoginfo/allow-domains/refs/heads/main/Russia/outside-raw.lst}"
SOURCE_PATH="${SOURCE_PATH:-$ROOT_DIR/data/outside/russia-outside.lst}"
PROBE_SOURCE_PATH="${PROBE_SOURCE_PATH:-$ROOT_DIR/data/outside/ru-probe-hosts.lst}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/data/ripe/ru}"
DATASET="${DATASET:-ru_direct_domains_v4}"
PROBE_DATASET="${PROBE_DATASET:-ru_probe_hosts_v4}"
MODE="${MODE:-local}"
SSH_TARGET="${SSH_TARGET:-}"
SSH_WORKDIR="${SSH_WORKDIR:-$ROOT_DIR}"
RESOLVER="${RESOLVER:-}"
CONCURRENCY="${CONCURRENCY:-16}"
TIMEOUT="${TIMEOUT:-5s}"
GO_BIN="${GO_BIN:-}"
SSH_GO_BIN="${SSH_GO_BIN:-go}"

usage() {
  cat <<'EOF'
Usage: scripts/build-ru-direct-list.sh

Downloads the latest Russia/outside-raw.lst source list and produces:
  - data/ripe/ru/ru_direct_domains_v4.jsonl
  - data/ripe/ru/ru_direct_domains_v4.csv
  - data/ripe/ru/ru_direct_domains_v4.prefixes.txt
  - data/ripe/ru/ru_probe_hosts_v4.jsonl
  - data/ripe/ru/ru_probe_hosts_v4.csv
  - data/ripe/ru/ru_probe_hosts_v4.prefixes.txt

Environment overrides:
  SOURCE_URL   Upstream raw list URL
  SOURCE_PATH  Local staged source list path
  PROBE_SOURCE_PATH  Local curated VPN/IP-check host list
  OUT_DIR      Output directory for generated artifacts
  DATASET      Output dataset name
  PROBE_DATASET Output dataset name for the probe-host subset
  MODE         local or ssh (default: local)
  SSH_TARGET   SSH target for MODE=ssh, e.g. user@host
  SSH_WORKDIR  Repo checkout path on the remote host
  RESOLVER     Optional DNS resolver host:port, e.g. 1.1.1.1:53
  CONCURRENCY  Max parallel DNS queries
  TIMEOUT      Per-host DNS timeout
  GO_BIN       Explicit local Go binary path
  SSH_GO_BIN   Explicit remote Go binary path
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

download() {
  local url="$1"
  local dst="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dst"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO "$dst" "$url"
    return
  fi
  printf 'curl or wget is required to download %s\n' "$url" >&2
  exit 1
}

run_resolve() {
  local input_path="$1"
  local dataset="$2"
  local resolve_args=(
    --input "$input_path"
    --dataset "$dataset"
    --out-dir "$OUT_DIR"
    --mode "$MODE"
    --concurrency "$CONCURRENCY"
    --timeout "$TIMEOUT"
  )
  if [[ -n "$RESOLVER" ]]; then
    resolve_args+=(--resolver "$RESOLVER")
  fi
  if [[ -n "$SSH_TARGET" ]]; then
    resolve_args+=(--ssh-target "$SSH_TARGET")
  fi
  if [[ -n "$SSH_WORKDIR" ]]; then
    resolve_args+=(--ssh-workdir "$SSH_WORKDIR")
  fi
  if [[ -n "$GO_BIN" ]]; then
    resolve_args+=(--go-bin "$GO_BIN")
  fi
  if [[ -n "$SSH_GO_BIN" ]]; then
    resolve_args+=(--ssh-go-bin "$SSH_GO_BIN")
  fi

  bash "$ROOT_DIR/scripts/resolve-hosts.sh" "${resolve_args[@]}"
}

mkdir -p "$(dirname "$SOURCE_PATH")" "$OUT_DIR"
tmp="$(mktemp)"
merged="$(mktemp)"
trap 'rm -f "$tmp" "$merged"' EXIT

printf 'Downloading source list from %s\n' "$SOURCE_URL"
download "$SOURCE_URL" "$tmp"
mv "$tmp" "$SOURCE_PATH"

if [[ ! -f "$PROBE_SOURCE_PATH" ]]; then
  printf 'missing probe host source list: %s\n' "$PROBE_SOURCE_PATH" >&2
  exit 1
fi

cat "$SOURCE_PATH" "$PROBE_SOURCE_PATH" > "$merged"

printf 'Resolving direct RU domains from %s + %s\n' "$SOURCE_PATH" "$PROBE_SOURCE_PATH"
run_resolve "$merged" "$DATASET"

printf 'Resolving VPN/IP-check probe hosts from %s\n' "$PROBE_SOURCE_PATH"
run_resolve "$PROBE_SOURCE_PATH" "$PROBE_DATASET"

printf 'Ready:\n'
printf '  %s/%s.prefixes.txt\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.csv\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.jsonl\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.prefixes.txt\n' "$OUT_DIR" "$PROBE_DATASET"
printf '  %s/%s.csv\n' "$OUT_DIR" "$PROBE_DATASET"
printf '  %s/%s.jsonl\n' "$OUT_DIR" "$PROBE_DATASET"
