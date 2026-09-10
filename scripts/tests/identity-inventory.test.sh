#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/docs/decisions" "$tmp/kernel/cmd/vera"
printf 'historical VERA name\n' >"$tmp/docs/decisions/VD-old.md"
printf 'schema vera.witness.v1\n' >"$tmp/wire.txt"
printf 'VERA_OWNER fallback\n' >"$tmp/alias.txt"

output=$(bash scripts/identity-inventory.sh "$tmp")
[[ $output == *$'frozen-history\tdocs/decisions/VD-old.md:1'* ]] || exit 1
[[ $output == *$'frozen-wire\twire.txt:1'* ]] || exit 1
[[ $output == *$'deprecated-alias\talias.txt:1'* ]] || exit 1
[[ $output == *$'deprecated-alias\tkernel/cmd/vera\tpath component'* ]] || exit 1
[[ $output == *'unclassified=0'* ]] || exit 1

printf 'VERA is still the live product\n' >"$tmp/unclassified.txt"
if bash scripts/identity-inventory.sh "$tmp" >/dev/null; then
  echo 'unclassified identity occurrence was accepted' >&2
  exit 1
fi
