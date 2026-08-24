#!/usr/bin/env bash
set -euo pipefail
repo=$(git rev-parse --show-toplevel)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/kernel/pkg" "$tmp/docs" "$tmp/scripts"
cp "$repo/scripts/gen-invariants-lock.sh" "$repo/scripts/spec-numbering-lint.sh" "$tmp/scripts/"
chmod +x "$tmp/scripts/"*.sh

write_spec() {
  printf '%s\n' "$1" > "$tmp/kernel/pkg/SPEC.md"
}
generate() {
  local stderr="$tmp/generator.stderr"
  SPEC_LINT_ROOT="$tmp" "$tmp/scripts/gen-invariants-lock.sh" 2>"$stderr"
  [ ! -s "$stderr" ] || { cat "$stderr" >&2; return 1; }
}
lint() {
  SPEC_LINT_ROOT="$tmp" "$tmp/scripts/spec-numbering-lint.sh"
}

write_spec '1. **INV-1 — First wrapped
title.** Detail.
2. **INV-2 — Second.** Detail.'
generate
lint
rg -q $'kernel/pkg/SPEC.md\tINV-1\tFirst wrapped title.' "$tmp/docs/invariants.lock"

write_spec '1. **INV-1 — Inserted.**
2. **INV-2 — First wrapped title.**
3. **INV-3 — Second.**'
if lint 2>/dev/null; then
  echo 'numbering insertion passed' >&2
  exit 1
fi

write_spec '1. **INV-1 — First wrapped title.**
2. **INV-2 — Second.**
3. **INV-3 — Appended.**'
if lint 2>/dev/null; then
  echo 'unlocked append passed' >&2
  exit 1
fi
generate
lint

write_spec '1. **INV-1 — One.**
2. **INV-1 — Duplicate.**'
if generate 2>/dev/null; then
  echo 'duplicate id passed' >&2
  exit 1
fi

: > "$tmp/docs/invariants.lock"
if lint 2>/dev/null; then
  echo 'empty lock passed' >&2
  exit 1
fi
