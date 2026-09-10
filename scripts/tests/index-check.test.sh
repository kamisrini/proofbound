#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
cp docs/decisions/INDEX.md "$tmp/index"
printf '\n' >> docs/decisions/INDEX.md
if scripts/index-check.sh; then exit 1; fi
mv "$tmp/index" docs/decisions/INDEX.md
scripts/index-check.sh

if PROOFBOUND_CHECK_TARGET=-n bash kernel/scripts/check-witness.sh >/dev/null 2>&1; then
  echo 'option-like witness target was accepted' >&2
  exit 1
fi

set +e
legacy_output=$(VERA_CHECK_TARGET=-n bash kernel/scripts/check-witness.sh 2>&1 >/dev/null)
legacy_status=$?
set -e
if [[ $legacy_status -eq 0 ]] || [[ $legacy_output != *'deprecated VERA_CHECK_TARGET alias used'* ]] || [[ $legacy_output != *'2026-12-31'* ]]; then
  echo 'legacy witness-target alias did not fail safely with its dated advisory' >&2
  exit 1
fi
