#!/usr/bin/env bash
set -euo pipefail

TOOL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "$TOOL_ROOT/../.." && pwd)"
SOURCE_URL="${SOURCE_URL:-https://raw.githubusercontent.com/hxehex/russia-mobile-internet-whitelist/main/whitelist.txt}"
OUT_DIR="${OUT_DIR:-$REPO_ROOT/build/ripex/ru}"
DATASET="${DATASET:-ru_wl_hosts_v4}"
PUBLIC_DIR="${PUBLIC_DIR:-$REPO_ROOT/lists/whitelist}"
MODE="${MODE:-local}"
SSH_TARGET="${SSH_TARGET:-}"
SSH_WORKDIR="${SSH_WORKDIR:-$REPO_ROOT}"
RESOLVER="${RESOLVER:-}"
CONCURRENCY="${CONCURRENCY:-16}"
TIMEOUT="${TIMEOUT:-5s}"
GO_BIN="${GO_BIN:-}"
SSH_GO_BIN="${SSH_GO_BIN:-go}"

usage() {
  cat <<'EOF'
Usage: tools/ripex/scripts/build-wl-hosts.sh

Downloads hxehex/russia-mobile-internet-whitelist whitelist.txt and
resolves it into a separate WL-host dataset:
  - build/ripex/ru/ru_wl_hosts_v4.jsonl
  - build/ripex/ru/ru_wl_hosts_v4.csv
  - build/ripex/ru/ru_wl_hosts_v4.prefixes.txt
  - lists/whitelist/ru_wl_hosts_v4.prefixes.txt

Environment overrides:
  SOURCE_URL   Upstream raw whitelist URL
  OUT_DIR      Output directory for generated artifacts
  DATASET      Output dataset name
  PUBLIC_DIR   Published WL-host list directory
  MODE         local or ssh (default: local)
  SSH_TARGET   SSH target for MODE=ssh, e.g. user@host
  SSH_WORKDIR  Repo checkout path on the remote host
  RESOLVER     Optional DNS resolver host:port
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
  local resolve_args=(
    --input "$input_path"
    --dataset "$DATASET"
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

  bash "$TOOL_ROOT/scripts/resolve-hosts.sh" "${resolve_args[@]}"
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

source_path="$tmpdir/whitelist.txt"
printf 'Downloading WL host list from %s\n' "$SOURCE_URL"
download "$SOURCE_URL" "$source_path"

printf 'Resolving WL hosts into %s/%s.*\n' "$OUT_DIR" "$DATASET"
run_resolve "$source_path"
mkdir -p "$PUBLIC_DIR"
cp "$OUT_DIR/$DATASET.prefixes.txt" "$PUBLIC_DIR/$DATASET.prefixes.txt"

printf 'Ready:\n'
printf '  %s/%s.prefixes.txt\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.csv\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.jsonl\n' "$OUT_DIR" "$DATASET"
printf '  %s/%s.prefixes.txt\n' "$PUBLIC_DIR" "$DATASET"
