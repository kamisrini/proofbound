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

echo 'ok intent-delivery-applicability'
