#!/usr/bin/env bash
set -euo pipefail

TOOL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "$TOOL_ROOT/../.." && pwd)"

MODE="${MODE:-local}"
INPUT=""
DATASET=""
OUT_DIR="${OUT_DIR:-$REPO_ROOT/build/ripex/ru}"
RESOLVER="${RESOLVER:-}"
CONCURRENCY="${CONCURRENCY:-16}"
TIMEOUT="${TIMEOUT:-5s}"
SSH_TARGET="${SSH_TARGET:-}"
SSH_WORKDIR="${SSH_WORKDIR:-$REPO_ROOT}"
GO_BIN="${GO_BIN:-}"
SSH_GO_BIN="${SSH_GO_BIN:-go}"

usage() {
  cat <<'EOF'
Usage: tools/ripex/scripts/resolve-hosts.sh --input PATH --dataset NAME [options]

Options:
  --input PATH          Newline-delimited host list to resolve
  --dataset NAME        Dataset name for generated artifacts
  --out-dir PATH        Output directory (default: build/ripex/ru)
  --mode local|ssh      Resolve locally or through ssh (default: local)
  --resolver HOST:PORT  Optional DNS resolver to use in local mode or pass to remote resolve
  --concurrency N       Max parallel DNS queries (default: 16)
  --timeout DURATION    Per-host DNS timeout (default: 5s)
  --ssh-target HOST     SSH target for ssh mode, e.g. user@host
  --ssh-workdir PATH    Repo checkout path on the remote host
  --go-bin PATH         Local Go binary to use for local resolution
  --ssh-go-bin PATH     Remote Go binary to use for ssh mode (default: go)
EOF
}

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

run_local() {
  local go_cmd
  go_cmd="$(find_go)"

  local args=(
    run ./cmd/ripex resolve
    --input "$INPUT"
    --dataset "$DATASET"
    --out-dir "$OUT_DIR"
    --concurrency "$CONCURRENCY"
    --timeout "$TIMEOUT"
  )
  if [[ -n "$RESOLVER" ]]; then
    args+=(--resolver "$RESOLVER")
  fi

  (
    cd "$TOOL_ROOT"
    "$go_cmd" "${args[@]}"
  )
}

q() {
  printf '%q' "$1"
}

run_ssh() {
  if [[ -z "$SSH_TARGET" ]]; then
    printf 'ssh mode requires --ssh-target user@host\n' >&2
    exit 1
  fi

  local remote_tmp remote_input remote_out remote_cmd
  remote_tmp="$(ssh "$SSH_TARGET" "mktemp -d")"
  remote_input="$remote_tmp/input.lst"
  remote_out="$remote_tmp/out"
  trap 'ssh "$SSH_TARGET" "rm -rf $(q "$remote_tmp")" >/dev/null 2>&1 || true' EXIT

  ssh "$SSH_TARGET" "mkdir -p $(q "$remote_out")"
  scp "$INPUT" "$SSH_TARGET:$remote_input"

  remote_cmd="cd $(q "$SSH_WORKDIR")/tools/ripex && $(q "$SSH_GO_BIN") run ./cmd/ripex resolve --input $(q "$remote_input") --dataset $(q "$DATASET") --out-dir $(q "$remote_out") --concurrency $(q "$CONCURRENCY") --timeout $(q "$TIMEOUT")"
  if [[ -n "$RESOLVER" ]]; then
    remote_cmd+=" --resolver $(q "$RESOLVER")"
  fi
  ssh "$SSH_TARGET" "$remote_cmd"

  mkdir -p "$OUT_DIR"
  scp "$SSH_TARGET:$remote_out/$DATASET.jsonl" "$OUT_DIR/$DATASET.jsonl"
  scp "$SSH_TARGET:$remote_out/$DATASET.csv" "$OUT_DIR/$DATASET.csv"
  scp "$SSH_TARGET:$remote_out/$DATASET.prefixes.txt" "$OUT_DIR/$DATASET.prefixes.txt"
}

if [[ $# -eq 0 ]]; then
  usage
  exit 1
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input)
      INPUT="${2:-}"
      shift 2
      ;;
    --dataset)
      DATASET="${2:-}"
      shift 2
      ;;
    --out-dir)
      OUT_DIR="${2:-}"
      shift 2
      ;;
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --resolver)
      RESOLVER="${2:-}"
      shift 2
      ;;
    --concurrency)
      CONCURRENCY="${2:-}"
      shift 2
      ;;
    --timeout)
      TIMEOUT="${2:-}"
      shift 2
      ;;
    --ssh-target)
      SSH_TARGET="${2:-}"
      shift 2
      ;;
    --ssh-workdir)
      SSH_WORKDIR="${2:-}"
      shift 2
      ;;
    --go-bin)
      GO_BIN="${2:-}"
      shift 2
      ;;
    --ssh-go-bin)
      SSH_GO_BIN="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$INPUT" ]]; then
  printf '--input is required\n' >&2
  exit 1
fi
if [[ -z "$DATASET" ]]; then
  printf '--dataset is required\n' >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

case "$MODE" in
  local)
    run_local
    ;;
  ssh)
    run_ssh
    ;;
  *)
    printf 'unsupported mode %q; use local or ssh\n' "$MODE" >&2
    exit 1
    ;;
esac
