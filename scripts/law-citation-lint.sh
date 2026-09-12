#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
[ -s docs/laws.lock ] || exit 1
expected=1
while IFS=$'\t' read -r n _; do
 [ "$n" = "$expected" ] || { echo 'law numbering drift' >&2; exit 1; }
 expected=$((expected+1))
done < <(awk -F '\t' '/^[0-9]+\t/{print}' docs/laws.lock)
[ "$(rg -c '^([[:space:]]*)[0-9]+\. \*\*' CLAUDE.md)" -eq $((expected-1)) ]
