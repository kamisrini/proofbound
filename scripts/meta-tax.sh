#!/usr/bin/env bash
set -euo pipefail

repo=${META_TAX_ROOT:-$(git rev-parse --show-toplevel)}
since=${META_TAX_SINCE:-$(git -C "$repo" rev-list --max-parents=0 HEAD | tail -1)}
git -C "$repo" cat-file -e "$since^{commit}" 2>/dev/null || { echo 'meta-tax: invalid META_TAX_SINCE commit' >&2; exit 1; }
scaffold=0
product=0
while IFS= read -r sha; do
  [[ -n $sha ]] || continue
  if git -C "$repo" diff-tree --root --no-commit-id --name-only -r "$sha" | rg -q '^(kernel/.*\.(go|sql)|tools/.*\.go)$'; then
    product=$((product + 1))
  else
    scaffold=$((scaffold + 1))
  fi
done < <(git -C "$repo" rev-list --reverse "$since^..HEAD" 2>/dev/null || git -C "$repo" rev-list --reverse HEAD)
total=$((scaffold + product))
percent=0
((total == 0)) || percent=$((scaffold * 100 / total))
printf 'scaffolding=%d product=%d total=%d meta_tax_percent=%d\n' "$scaffold" "$product" "$total" "$percent"
if [[ -n ${META_TAX_MAX_PERCENT:-} ]]; then
  [[ $META_TAX_MAX_PERCENT =~ ^[0-9]+$ ]] || { echo 'meta-tax: invalid META_TAX_MAX_PERCENT' >&2; exit 1; }
  ((percent <= META_TAX_MAX_PERCENT)) || exit 1
fi
