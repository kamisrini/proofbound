#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
repo=$tmp/repo
git init -q "$repo"
git -C "$repo" config user.email fixture@example.invalid
git -C "$repo" config user.name fixture
mkdir -p "$repo/kernel" "$repo/scripts"
cp scripts/intent-applicability.sh "$repo/scripts/"
printf 'x\n' >"$repo/kernel/main.go"
git -C "$repo" add .
git -C "$repo" commit -qm initial
missing=$(git -C "$repo" rev-parse HEAD)
if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$missing" >/dev/null 2>&1; then
  echo 'applicable commit without Intent trailer was accepted' >&2
  exit 1
fi

printf 'docs\n' >"$repo/README.md"
git -C "$repo" add .
git -C "$repo" commit -qm docs
non_applicable=$(git -C "$repo" rev-parse HEAD)
scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$non_applicable" >/dev/null

mkdir -p "$repo/scripts"
printf 'script\n' >"$repo/scripts/check.sh"
git -C "$repo" add .
git -C "$repo" commit -q -m applicable -m 'Intent: CI-fixture-item-acde12'
with_intent=$(git -C "$repo" rev-parse HEAD)
scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$with_intent" >/dev/null

# Shared corpus (DEF-INTENT-TRAILER-BLOCK-001): the gate must accept exactly the commits whose
# Intent claim the kernel binds. The same cases.tsv drives the kernel's intentTrailers test, so a
# drift in either parser fails a test. Column 3 is the gate's expected verdict.
corpus=scripts/tests/fixtures/intent-trailers
cases=0
while IFS=$'\t' read -r name _ expect; do
  [[ -z $name || $name == \#* ]] && continue
  printf '%s\n' "$name" >"$repo/scripts/case-$name.sh"
  git -C "$repo" add .
  git -C "$repo" commit -q --cleanup=verbatim -F "$PWD/$corpus/$name.msg"
  sha=$(git -C "$repo" rev-parse HEAD)
  if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$sha" >/dev/null 2>&1; then got=pass; else got=fail; fi
  if [[ $got != "$expect" ]]; then
    echo "intent-delivery-applicability: corpus case $name: gate=$got, expected $expect" >&2
    exit 1
  fi
  cases=$((cases + 1))
done <"$corpus/cases.tsv"
# Calibration: a vacuous corpus must not read as a pass.
((cases >= 12)) || { echo "intent-delivery-applicability: corpus ran $cases cases, want >= 12" >&2; exit 1; }

# The diagnostic must name the real cause for the a028984 shape (body-paragraph Intent line).
git -C "$repo" commit -q --allow-empty --cleanup=verbatim -F "$PWD/$corpus/body-paragraph.msg"
printf 'x\n' >"$repo/scripts/diag.sh"; git -C "$repo" add .; git -C "$repo" commit -q --amend --no-edit --cleanup=verbatim
if err=$(scripts/intent-delivery-applicability.sh --repo "$repo" --commit HEAD 2>&1); then
  echo 'intent-delivery-applicability: body-paragraph Intent was accepted' >&2; exit 1
fi
[[ $err == *'OUTSIDE the final trailer block'* ]] || { echo "intent-delivery-applicability: wrong diagnostic: $err" >&2; exit 1; }

echo 'ok intent-delivery-applicability'
