#!/usr/bin/env bash
set -euo pipefail

# meta-tax is an ON-DEMAND DIAGNOSTIC of scaffolding vs product commits, not a
# mechanical gate (Law 5, amended 2026-10-31; DEF-METATAX-001). A cumulative
# lifetime ratio is dominated by immutable bootstrap history and cannot reflect
# recent discipline, so by default this reports a TRAILING WINDOW of the last
# META_TAX_WINDOW commits. META_TAX_SINCE pins an explicit baseline instead; a
# history shorter than the window falls back to the whole history.

repo=${META_TAX_ROOT:-$(git rev-parse --show-toplevel)}
window=${META_TAX_WINDOW:-50}
[[ $window =~ ^[0-9]+$ ]] || { echo 'meta-tax: invalid META_TAX_WINDOW' >&2; exit 1; }

if [[ -n ${META_TAX_SINCE:-} ]]; then
  git -C "$repo" cat-file -e "${META_TAX_SINCE}^{commit}" 2>/dev/null || { echo 'meta-tax: invalid META_TAX_SINCE commit' >&2; exit 1; }
  range=$(git -C "$repo" rev-list --reverse "${META_TAX_SINCE}^..HEAD" 2>/dev/null || git -C "$repo" rev-list --reverse HEAD)
elif base=$(git -C "$repo" rev-parse --verify --quiet "HEAD~${window}^{commit}"); then
  range=$(git -C "$repo" rev-list --reverse "${base}..HEAD")
else
  range=$(git -C "$repo" rev-list --reverse HEAD)
fi

scaffold=0
product=0
while IFS= read -r sha; do
  [[ -n $sha ]] || continue
  if git -C "$repo" diff-tree --root --no-commit-id --name-only -r "$sha" | rg -q '^(kernel/.*\.(go|sql)|tools/.*\.go)$'; then
    product=$((product + 1))
  else
    scaffold=$((scaffold + 1))
  fi
done <<<"$range"

total=$((scaffold + product))
percent=0
((total == 0)) || percent=$((scaffold * 100 / total))
printf 'scaffolding=%d product=%d total=%d meta_tax_percent=%d\n' "$scaffold" "$product" "$total" "$percent"

# META_TAX_MAX_PERCENT remains available for a human-driven spot check; it is not
# wired into make check (Law 5 is a periodic review, not a blocking gate).
if [[ -n ${META_TAX_MAX_PERCENT:-} ]]; then
  [[ $META_TAX_MAX_PERCENT =~ ^[0-9]+$ ]] || { echo 'meta-tax: invalid META_TAX_MAX_PERCENT' >&2; exit 1; }
  ((percent <= META_TAX_MAX_PERCENT)) || exit 1
fi
