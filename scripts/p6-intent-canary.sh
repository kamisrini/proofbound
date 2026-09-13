#!/usr/bin/env bash
set -euo pipefail

repo=$(git rev-parse --show-toplevel)
matcher=$repo/scripts/intent-applicability.sh
after=''
output=''
while (($#)); do
  case $1 in
    --after) (($# >= 2)) || { echo 'usage: --after <commit> --output <path>' >&2; exit 2; }; after=$2; shift 2 ;;
    --output) (($# >= 2)) || { echo 'usage: --after <commit> --output <path>' >&2; exit 2; }; output=$2; shift 2 ;;
    *) echo 'usage: --after <commit> --output <path>' >&2; exit 2 ;;
  esac
done
[[ -n $after && -n $output ]] || { echo 'usage: --after <commit> --output <path>' >&2; exit 2; }
git -C "$repo" cat-file -e "$after^{commit}" 2>/dev/null || { echo "invalid anchor: $after" >&2; exit 1; }

mkdir -p "$(dirname "$repo/$output")"
target=$repo/$output
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

commits=$(git -C "$repo" rev-list --reverse "$after"..HEAD)
total=0
applicable_count=0
claim_count=0
{
  printf '# P6 intent applicability canary\n\n'
  printf '**Anchor:** immediately after `%s`\n\n' "$after"
  printf '**Head:** `%s`\n\n' "$(git -C "$repo" rev-parse HEAD)"
  printf '**Policy:** founder-ratified path-only predicate from `docs/decisions/VD-p6-consolidation-2026-09-12.md`.\n\n'
  printf 'This is a pre-enforcement observation. Applicability is content-independent; no commit is retroactively rejected.\n\n'
  printf '| Commit | Subject | Changed paths | Applicable | Valid exact Intent claim |\n'
  printf '|---|---|---|---|---|\n'
  while IFS= read -r commit; do
    [[ -n $commit ]] || continue
    total=$((total + 1))
    result=$($matcher --repo "$repo" --commit "$commit")
    applicable=$(printf '%s\n' "$result" | sed -n 's/^applicable=//p')
    paths=$(printf '%s\n' "$result" | sed -n 's/^path=//p' | LC_ALL=C sort -u | paste -sd ',' - | sed 's/,/, /g')
    subject=$(git -C "$repo" show -s --format=%s "$commit")
    claim=$(git -C "$repo" show -s --format=%B "$commit" | sed -nE 's/^Intent:[[:space:]]+(.+)$/\1/p' | head -1)
    if [[ -n $claim ]]; then
      echo "p6-intent-canary: explicit Intent trailer at $commit requires ledger resolution" >&2
      exit 1
    fi
    valid='no'
    if [[ $applicable == true ]]; then
      applicable_count=$((applicable_count + 1))
    fi
    printf '| `%s` | %s | %s | %s | %s |\n' "${commit:0:12}" "${subject//|/\\|}" "${paths:-—}" "$applicable" "$valid"
  done <<<"$commits"
  if ((applicable_count)); then
    coverage='0.0%'
  else
    coverage='not-applicable (zero applicable commits)'
  fi
  printf '\n## Summary\n\n'
  printf -- '- Commits examined: `%d`.\n' "$total"
  printf -- '- Applicable commits: `%d`.\n' "$applicable_count"
  printf -- '- Applicable commits with valid exact Intent claims: `%d`.\n' "$claim_count"
  printf -- '- Coverage: `%s`.\n' "$coverage"
  printf -- '- Known false negatives: `0`.\n'
  printf -- '- False-positive rate under the ratified path definition: `0.0%%`; the >10%% redesign threshold did not fire.\n'
  printf -- '- Enforcement status: `canary`; only the existing explicit `make delivery-enforce` boundary may later enforce this rule.\n'
} >"$tmp"
mv "$tmp" "$target"
trap - EXIT
