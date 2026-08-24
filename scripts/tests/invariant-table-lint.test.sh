#!/usr/bin/env bash
set -euo pipefail
repo=$(git rev-parse --show-toplevel)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/kernel/pkg" "$tmp/scripts"
cp "$repo/scripts/invariant-table-lint.sh" "$tmp/scripts/"
chmod +x "$tmp/scripts/invariant-table-lint.sh"

lint() {
  SPEC_LINT_ROOT="$tmp" "$tmp/scripts/invariant-table-lint.sh"
}

printf '%s\n' '| INV-1 | claim | pkg_test.go::TestProof |' > "$tmp/kernel/pkg/SPEC.md"
lint

printf '%s\n' '| C-INV-1 | claim | pkg_test.go::TestProof |' > "$tmp/kernel/pkg/SPEC.md"
lint

printf '%s\n' '| INV-1 | claim | pkg_test.go::TestProof, TestOther |' > "$tmp/kernel/pkg/SPEC.md"
if lint 2>/dev/null; then
  echo 'multiple citations in one cell passed' >&2
  exit 1
fi

printf '%s\n' '| INV-1 | claim | |' > "$tmp/kernel/pkg/SPEC.md"
if lint 2>/dev/null; then
  echo 'empty citation passed' >&2
  exit 1
fi

printf '%s\n' '| INV-1 | retired | — |' > "$tmp/kernel/pkg/SPEC.md"
lint
