#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT; mkdir -p "$tmp/kernel/internal/x"
printf '# SPEC\n`--one --two`\n' >"$tmp/kernel/internal/x/SPEC.md"; printf 'package x\nvar x = "--one"\n' >"$tmp/kernel/internal/x/x.go"
if PRESCRIPTION_ROOT=$tmp scripts/prescription-lint.sh >/dev/null 2>&1; then exit 1; fi
printf 'package x\nvar x = []string{"--one", "--two"}\n' >"$tmp/kernel/internal/x/x.go"; PRESCRIPTION_ROOT=$tmp scripts/prescription-lint.sh
