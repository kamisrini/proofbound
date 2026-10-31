#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
bad=0
# (1) every decision file resolves in the generated index
while IFS= read -r file; do
 id=$(basename "$file" .md)
 rg -F -q "$id" docs/decisions/INDEX.md || { echo "decision missing from index: $id" >&2; bad=1; }
done < <(find docs/decisions -maxdepth 1 -type f -name 'VD-*.md' | sort)
# (2) Law 4: every ADVISORY row in the gate registry is dated and unexpired
scripts/advisory-expiry.sh || bad=1
exit "$bad"
