#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

script=$PWD/scripts/p6-task8-package-acceptance.sh
bash "$script" --check
bash "$script" --package internal/cli | rg -q '^PASS internal/cli '
if bash "$script" --package internal/specfirst >/dev/null 2>&1; then
  echo 'test-only package was accepted' >&2
  exit 1
fi
echo 'ok package-set-and-current-verdict'

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git clone -q --shared "$PWD" "$tmp/repo"

rehash_verdict() {
  local path=$1 zero=0000000000000000000000000000000000000000000000000000000000000000 digest
  digest=$(sed -E "s/^artifact_sha: [0-9a-f]{64}$/artifact_sha: $zero/" "$path" | sha256sum | awk '{print $1}')
  sed -i "s/^artifact_sha: .*/artifact_sha: $digest/" "$path"
}

awk -F '\t' 'BEGIN { OFS = FS } $1 == "internal/cli" { $4 = $4 - 1 } { print }' \
  "$tmp/repo/docs/verification/p6-task8-package-results-e4c8e77.tsv" \
  >"$tmp/modified.tsv"
mv "$tmp/modified.tsv" "$tmp/repo/docs/verification/p6-task8-package-results-e4c8e77.tsv"
if bash "$tmp/repo/scripts/p6-task8-package-acceptance.sh" --check --root "$tmp/repo" >/dev/null 2>&1; then
  echo 'mutation count mismatch was accepted' >&2
  exit 1
fi
echo 'ok mutation-count-mismatch'

git -C "$tmp/repo" checkout -q -- docs/verification/p6-task8-package-results-e4c8e77.tsv
awk -F '\t' 'BEGIN { OFS = FS } $1 == "internal/cli" { print $1, $2, $3, $4, $5, $6, $7; next } { print }' \
  "$tmp/repo/docs/verification/p6-task8-package-results-e4c8e77.tsv" \
  >"$tmp/modified.tsv"
mv "$tmp/modified.tsv" "$tmp/repo/docs/verification/p6-task8-package-results-e4c8e77.tsv"
if bash "$tmp/repo/scripts/p6-task8-package-acceptance.sh" --check --root "$tmp/repo" >/dev/null 2>&1; then
  echo 'manifest row with a missing field was accepted' >&2
  exit 1
fi
echo 'ok manifest-row-width'

git -C "$tmp/repo" checkout -q -- docs/verification/p6-task8-package-results-e4c8e77.tsv
sed -i "s/^reviewed_commit: .*/reviewed_commit: 0000000000000000000000000000000000000000/" \
  "$tmp/repo/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md"
rehash_verdict "$tmp/repo/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md"
if bash "$tmp/repo/scripts/p6-task8-package-acceptance.sh" --check --root "$tmp/repo" >/dev/null 2>&1; then
  echo 'verdict bound to the wrong implementation commit was accepted' >&2
  exit 1
fi
echo 'ok verdict-commit-binding'

git -C "$tmp/repo" checkout -q -- docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md
sed -i 's/^reviewer_id: .*/reviewer_id: VERA Maintainer/' \
  "$tmp/repo/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md"
rehash_verdict "$tmp/repo/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md"
if bash "$tmp/repo/scripts/p6-task8-package-acceptance.sh" --check --root "$tmp/repo" >/dev/null 2>&1; then
  echo 'author-authored verdict was accepted' >&2
  exit 1
fi
echo 'ok non-author-identity'

git -C "$tmp/repo" checkout -q -- docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md
sed -i 's/^artifact_sha: .*/artifact_sha: ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff/' \
  "$tmp/repo/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md"
if bash "$tmp/repo/scripts/p6-task8-package-acceptance.sh" --check --root "$tmp/repo" >/dev/null 2>&1; then
  echo 'verdict with a bad self-digest was accepted' >&2
  exit 1
fi
echo 'ok verdict-self-digest'

git -C "$tmp/repo" checkout -q -- docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md
sed -i 's#docs/verification/p6-task8-package-results-e4c8e77.tsv,docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md#—#' \
  "$tmp/repo/docs/plans/p6-census-rows.tsv"
census=$(bash "$tmp/repo/scripts/p6-census.sh" --render --root "$tmp/repo")
[[ $census == *'| C3-001 | C3 | package:internal/cli |'*'| open | close-in-P6 | — | — |'* ]] || {
  echo 'package row closed without listing packet and verdict evidence' >&2
  exit 1
}
echo 'ok census-evidence-linkage'
