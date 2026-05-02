#!/usr/bin/env bash
set -euo pipefail

TOOL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "$TOOL_ROOT/../.." && pwd)"
OUT_DIR="${OUT_DIR:-$REPO_ROOT/build/ripex/ru}"
OUT_FILE="${OUT_FILE:-$OUT_DIR/ru_all_v4.prefixes.txt}"
LISTS_DIR="${LISTS_DIR:-$REPO_ROOT/lists}"
RIPE_LISTS_DIR="${RIPE_LISTS_DIR:-$LISTS_DIR/ripe}"
DOMAINS_LISTS_DIR="${DOMAINS_LISTS_DIR:-$LISTS_DIR/domains}"
PROBE_LISTS_DIR="${PROBE_LISTS_DIR:-$LISTS_DIR/probe}"
WHITELIST_LISTS_DIR="${WHITELIST_LISTS_DIR:-$LISTS_DIR/whitelist}"
MERGED_LISTS_DIR="${MERGED_LISTS_DIR:-$LISTS_DIR/merged}"
PUBLIC_PREFIX_FILE="${PUBLIC_PREFIX_FILE:-$MERGED_LISTS_DIR/ru_all_v4.prefixes.txt}"
MANIFEST_FILE="${MANIFEST_FILE:-$OUT_DIR/manifest.json}"
PUBLIC_MANIFEST_FILE="${PUBLIC_MANIFEST_FILE:-$MERGED_LISTS_DIR/manifest.json}"
GO_BIN="${GO_BIN:-}"

usage() {
  cat <<'EOF'
Usage: tools/ripex/scripts/build-ru-router-list.sh

Builds a single router-facing RU prefix feed by merging and minimizing:
  - build/ripex/ru/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt
  - build/ripex/ru/ru_direct_domains_v4.prefixes.txt
  - build/ripex/ru/ru_probe_hosts_v4.prefixes.txt
  - build/ripex/ru/ru_wl_hosts_v4.prefixes.txt
  - lists/ripe/*.prefixes.txt
  - lists/domains/*.prefixes.txt
  - lists/probe/*.prefixes.txt
  - lists/whitelist/*.prefixes.txt
  - lists/merged/ru_all_v4.prefixes.txt
  - lists/merged/manifest.json

Environment overrides:
  OUT_DIR               Output directory for generated artifacts
  OUT_FILE              Output file path (default: build/ripex/ru/ru_all_v4.prefixes.txt)
  LISTS_DIR             Published list root (default: lists)
  MANIFEST_FILE         Source manifest path (default: build/ripex/ru/manifest.json)
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
mkdir -p "$RIPE_LISTS_DIR" "$DOMAINS_LISTS_DIR" "$PROBE_LISTS_DIR" "$WHITELIST_LISTS_DIR" "$MERGED_LISTS_DIR"

args=(
  run ./cmd/ripex merge-prefixes
  --out "$OUT_FILE"
)
for input in "${inputs[@]}"; do
  args+=(--input "$input")
done

(
  cd "$TOOL_ROOT"
  "$(find_go)" "${args[@]}"
)

cp "$OUT_DIR/ru_org_inetnum_v4.prefixes.txt" "$RIPE_LISTS_DIR/ru_org_inetnum_v4.prefixes.txt"
cp "$OUT_DIR/ru_as_route_v4.prefixes.txt" "$RIPE_LISTS_DIR/ru_as_route_v4.prefixes.txt"
cp "$OUT_DIR/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt" "$RIPE_LISTS_DIR/ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt"
cp "$OUT_DIR/ru_direct_domains_v4.prefixes.txt" "$DOMAINS_LISTS_DIR/ru_direct_domains_v4.prefixes.txt"
cp "$OUT_DIR/ru_probe_hosts_v4.prefixes.txt" "$PROBE_LISTS_DIR/ru_probe_hosts_v4.prefixes.txt"
cp "$OUT_DIR/ru_wl_hosts_v4.prefixes.txt" "$WHITELIST_LISTS_DIR/ru_wl_hosts_v4.prefixes.txt"
cp "$OUT_FILE" "$PUBLIC_PREFIX_FILE"
cp "$MANIFEST_FILE" "$PUBLIC_MANIFEST_FILE"

printf 'Ready:\n'
printf '  %s\n' "$OUT_FILE"
printf '  %s\n' "$PUBLIC_PREFIX_FILE"
printf '  %s\n' "$PUBLIC_MANIFEST_FILE"
