#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

scripts/p6-task6-results.sh --check

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git clone -q --no-local . "$tmp/repo"
fixture=$tmp/repo
cp scripts/p6-task6-results.sh "$fixture/scripts/p6-task6-results.sh"
cp docs/plans/p6-task6-results.tsv "$fixture/docs/plans/p6-task6-results.tsv"
cp docs/verification/p6-measurements-falsifiers.md "$fixture/docs/verification/p6-measurements-falsifiers.md"

expect_fail() {
  local label=$1
  shift
  if "$@" >/dev/null 2>&1; then
    echo "p6-task6-results: $label was accepted" >&2
    exit 1
  fi
}

sed -i '0,/^measurement\tfalse-or-missing-commit-intent-links\tmeasured\t/s//measurement\tfalse-or-missing-commit-intent-links\tINVALID\t/' "$fixture/docs/plans/p6-task6-results.tsv"
expect_fail 'invalid result enum' env P6_TASK6_ROOT="$fixture" "$fixture/scripts/p6-task6-results.sh" --check
cp docs/plans/p6-task6-results.tsv "$fixture/docs/plans/p6-task6-results.tsv"

sed -i '/^falsifier\tforeign-provider-schema-change\t/d' "$fixture/docs/plans/p6-task6-results.tsv"
expect_fail 'missing falsifier row' env P6_TASK6_ROOT="$fixture" "$fixture/scripts/p6-task6-results.sh" --check
cp docs/plans/p6-task6-results.tsv "$fixture/docs/plans/p6-task6-results.tsv"

printf '\n' >> "$fixture/docs/verification/p6-measurements-falsifiers.md"
expect_fail 'stale generated result artifact' env P6_TASK6_ROOT="$fixture" "$fixture/scripts/p6-task6-results.sh" --check

echo 'p6-task6-results: hostile result checks passed'
