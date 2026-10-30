#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
awk '/^check:/,/^[^[:space:]]/ {if ($0 ~ /-short/) exit 1}' Makefile
awk '/^short:/ {seen=1; if ($0 !~ /hooks-test/) exit 1} /go test \.\/\.\.\. -short/ {go=1} END {exit !(seen && go)}' Makefile
for target in backup meta-tax wrap-verify laws-lock state; do
  rg -q "^${target}:" Makefile || { echo "missing Make target: $target" >&2; exit 1; }
done

check_block=$(awk '/^check:/,/^[^[:space:]]/ {print}' Makefile)
for forbidden in proofbound check-witnessed delivery-enforce gates-enforce sync; do
  if [[ $check_block == *"$forbidden"* ]]; then
    echo "make check depends on product state through $forbidden" >&2
    exit 1
  fi
done

# hooks-test must fail closed on the FIRST failing self-test, not only the last.
# A loop that returns only its last command's status masks intermediate failures
# (the defect journaled 2026-10-17, class=detector-inert-or-scope-blind). This runs
# the Makefile's real recipe against hostile fixtures, with calibration controls:
# a fail-then-pass set that MUST be rejected, and an all-pass set that MUST survive.
recipe=$(awk '/^hooks-test:/{getline; sub(/^\t@?/,""); gsub(/\$\$/,"$"); print; exit}' Makefile)
[[ -n $recipe ]] || { echo 'could not extract the hooks-test recipe' >&2; exit 1; }
fixture=$(mktemp -d); trap 'rm -rf "$fixture"' EXIT
mkdir -p "$fixture/scripts/tests"
printf 'exit 1\n' >"$fixture/scripts/tests/aaa-intermediate.test.sh"
printf 'exit 0\n' >"$fixture/scripts/tests/zzz-last.test.sh"
if ( cd "$fixture" && eval "$recipe" ) >/dev/null 2>&1; then
  echo 'hooks-test loop is NOT fail-closed: an intermediate failure was masked' >&2
  exit 1
fi
printf 'exit 0\n' >"$fixture/scripts/tests/aaa-intermediate.test.sh"
if ! ( cd "$fixture" && eval "$recipe" ) >/dev/null 2>&1; then
  echo 'hooks-test loop rejects an all-passing self-test set' >&2
  exit 1
fi
