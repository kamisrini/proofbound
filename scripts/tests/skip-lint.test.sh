#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
printf 'ok package/x\n' >"$tmp/log"; printf '# none\n' >"$tmp/allow"; SKIP_LINT_ALLOWLIST=$tmp/allow scripts/skip-lint.sh "$tmp/log"
printf '%s\n' '--- SKIP: TestNeedsThing (0.00s)' >>"$tmp/log"
if SKIP_LINT_ALLOWLIST=$tmp/allow scripts/skip-lint.sh "$tmp/log" >/dev/null 2>&1; then exit 1; fi
printf 'TestNeedsThing  # environment lacks thing\n' >"$tmp/allow"; SKIP_LINT_ALLOWLIST=$tmp/allow scripts/skip-lint.sh "$tmp/log"
if SKIP_LINT_ALLOWLIST=$tmp/allow scripts/skip-lint.sh /dev/null >/dev/null 2>&1; then exit 1; fi
