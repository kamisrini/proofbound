#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

bash scripts/p6-artifact-integrity.sh --check
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
cp docs/verification/verdicts/*.md "$tmp/"
sed -i 's/"artifact_sha256":"[0-9a-f]\{64\}"/"artifact_sha256":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"/' "$tmp/p5-requirement-review-round1.md"
if P6_INTEGRITY_VERDICTS="$tmp" bash scripts/p6-artifact-integrity.sh --check >/dev/null 2>&1; then
  echo 'corrupt verdict digest was accepted' >&2
  exit 1
fi
