#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

scripts/p6-task5-review.sh --check

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git clone -q --no-local . "$tmp/repo"
fixture=$tmp/repo
cp scripts/p6-task5-review.sh "$fixture/scripts/p6-task5-review.sh"
cp docs/verification/p6-requirement-review.md "$fixture/docs/verification/p6-requirement-review.md"
cp "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md" "$tmp/review.md"

expect_fail() {
  local label=$1
  shift
  if "$@" >/dev/null 2>&1; then
    echo "p6-task5-review: $label was accepted" >&2
    exit 1
  fi
}

expect_fail "missing exact review" env P6_TASK5_ROOT="$fixture" bash -c \
  'rm "$P6_TASK5_ROOT/docs/verification/verdicts/p5-requirement-review-round1.md" && "$P6_TASK5_ROOT/scripts/p6-task5-review.sh" --check'

cp "$tmp/review.md" "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md"
sed -i 's/"declared_reviewer":"proofbound-independent-verifier-round1"/"declared_reviewer":"proofbound-maintainer"/' "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md"
expect_fail "author and reviewer equality" env P6_TASK5_ROOT="$fixture" "$fixture/scripts/p6-task5-review.sh" --check
cp "$tmp/review.md" "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md"

sed -i 's/"outcome":"VERIFIABLE"/"outcome":"AMBIGUOUS"/' "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md"
expect_fail "non-verifiable outcome" env P6_TASK5_ROOT="$fixture" "$fixture/scripts/p6-task5-review.sh" --check
cp "$tmp/review.md" "$fixture/docs/verification/verdicts/p5-requirement-review-round1.md"

printf '\n' >> "$fixture/docs/verification/p6-requirement-review.md"
expect_fail "stale generated packet" env P6_TASK5_ROOT="$fixture" "$fixture/scripts/p6-task5-review.sh" --check

echo 'p6-task5-review: hostile closure checks passed'
