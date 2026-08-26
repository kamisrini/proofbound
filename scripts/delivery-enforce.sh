#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
mkdir -p .vera
lock=.vera/delivery.lock
if ! mkdir "$lock" 2>/dev/null; then
  pid_file="$lock/pid"
  pid=''
  if [[ -f $pid_file ]]; then
    read -r pid <"$pid_file" || true
  fi
  if [[ $pid =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null; then
    printf 'delivery-enforce: another workflow is running (pid %s)\n' "$pid" >&2
    exit 1
  fi
  rm -f "$pid_file"
  rmdir "$lock" 2>/dev/null || {
    printf 'delivery-enforce: stale lock cannot be reclaimed\n' >&2
    exit 1
  }
  mkdir "$lock"
fi
printf '%s\n' "$$" >"$lock/pid"
cleanup() {
  rm -f "$lock/pid"
  rmdir "$lock" 2>/dev/null || true
}
trap cleanup EXIT

make index-check-witnessed
make law-citation-witnessed
make spec-numbering-witnessed
make invariant-table-witnessed
make link-witnessed
make kernel-check-witnessed
VERA_CHECK_TARGET=check make check-witnessed
(cd kernel && go run ./cmd/vera sync checks)
(cd kernel && go run ./cmd/vera gates enforce)
