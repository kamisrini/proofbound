#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT; mkdir -p "$tmp/kernel/internal/x"
printf '# SPEC\n| X-INV-1 | claim | x_test.go::TestClaim |\n' >"$tmp/kernel/internal/x/SPEC.md"; printf 'package x\n' >"$tmp/kernel/internal/x/x_test.go"
if INVARIANT_LINT_ROOT=$tmp scripts/invariant-lint.sh >/dev/null 2>&1; then exit 1; fi
printf 'package x\nfunc TestClaim(t *testing.T) {}\n' >"$tmp/kernel/internal/x/x_test.go"; INVARIANT_LINT_ROOT=$tmp scripts/invariant-lint.sh
