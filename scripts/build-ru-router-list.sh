#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/data/ripe/ru}"
OUT_FILE="${OUT_FILE:-$OUT_DIR/ru_all_v4.prefixes.txt}"
ASSETS_DIR="${ASSETS_DIR:-$ROOT_DIR/assets}"
ASSETS_PREFIX_FILE="${ASSETS_PREFIX_FILE:-$ASSETS_DIR/ru_all_v4.prefixes.txt}"
MANIFEST_FILE="${MANIFEST_FILE:-$OUT_DIR/manifest.json}"
ASSETS_MANIFEST_FILE="${ASSETS_MANIFEST_FILE:-$ASSETS_DIR/manifest.json}"
GO_BIN="${GO_BIN:-}"

usage() {
  cat <<'EOF'
Usage: scripts/build-ru-router-list.sh

Builds a single router-facing RU prefix feed by merging and minimizing:
  - data/ripe/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt
  - data/ripe/ru/ru_direct_domains_v4.prefixes.txt
  - data/ripe/ru/ru_probe_hosts_v4.prefixes.txt
  - data/ripe/ru/ru_wl_hosts_v4.prefixes.txt

Environment overrides:
  OUT_DIR               Output directory for generated artifacts
  OUT_FILE              Output file path (default: data/ripe/ru/ru_all_v4.prefixes.txt)
  ASSETS_DIR            Directory for tracked release assets (default: assets)
  ASSETS_PREFIX_FILE    Release prefix file path (default: assets/ru_all_v4.prefixes.txt)
  MANIFEST_FILE         Source manifest path (default: data/ripe/ru/manifest.json)
  ASSETS_MANIFEST_FILE  Release manifest path (default: assets/manifest.json)
  GO_BIN                Explicit Go binary path
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

find_go() {
  if [[ -n "$GO_BIN" ]]; then
    printf '%s\n' "$GO_BIN"
    return
  fi
  if command -v go >/dev/null 2>&1; then
    command -v go
    return
  fi
  if [[ -x /usr/local/go/bin/go ]]; then
    printf '%s\n' /usr/local/go/bin/go
    return
  fi
  printf 'go binary not found; set GO_BIN=/path/to/go\n' >&2
  exit 1
}

inputs=(
  "$OUT_DIR/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt"
  "$OUT_DIR/ru_direct_domains_v4.prefixes.txt"
  "$OUT_DIR/ru_probe_hosts_v4.prefixes.txt"
  "$OUT_DIR/ru_wl_hosts_v4.prefixes.txt"
)

for input in "${inputs[@]}"; do
  if [[ ! -f "$input" ]]; then
    printf 'missing input prefix file: %s\n' "$input" >&2
    exit 1
  fi
done

if [[ ! -f "$MANIFEST_FILE" ]]; then
  printf 'missing manifest file: %s\n' "$MANIFEST_FILE" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT_FILE")"
mkdir -p "$ASSETS_DIR"

args=(
  run ./cmd/ripex merge-prefixes
  --out "$OUT_FILE"
)
for input in "${inputs[@]}"; do
  args+=(--input "$input")
done

(
  cd "$ROOT_DIR"
  "$(find_go)" "${args[@]}"
)

cp "$OUT_FILE" "$ASSETS_PREFIX_FILE"
cp "$MANIFEST_FILE" "$ASSETS_MANIFEST_FILE"

printf 'Ready:\n'
printf '  %s\n' "$OUT_FILE"
printf '  %s\n' "$ASSETS_PREFIX_FILE"
printf '  %s\n' "$ASSETS_MANIFEST_FILE"
