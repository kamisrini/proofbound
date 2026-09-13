#!/usr/bin/env bash
set -euo pipefail

repo=$(git rev-parse --show-toplevel)
commit=HEAD
while (($#)); do
  case $1 in
    --repo) (($# >= 2)) || { echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2; }; repo=$2; shift 2 ;;
    --commit) (($# >= 2)) || { echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2; }; commit=$2; shift 2 ;;
    *) echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2 ;;
  esac
done
repo=$(cd "$repo" && pwd)
git -C "$repo" cat-file -e "$commit^{commit}" 2>/dev/null || {
  printf 'intent-delivery-applicability: invalid commit %s\n' "$commit" >&2
  exit 1
}

result=$($repo/scripts/intent-applicability.sh --repo "$repo" --commit "$commit")
applicable=$(printf '%s\n' "$result" | sed -n 's/^applicable=//p')
if [[ $applicable != true ]]; then
  printf 'applicable=false commit=%s\n' "$(git -C "$repo" rev-parse "$commit")"
  exit 0
fi

if ! git -C "$repo" show -s --format=%B "$commit" | grep -Eq '^Intent:[[:space:]]+[^[:space:]]'; then
  printf 'intent-delivery-applicability: applicable commit %s has no explicit Intent trailer\n' "$(git -C "$repo" rev-parse "$commit")" >&2
  exit 1
fi
printf 'applicable=true commit=%s explicit-intent=true\n' "$(git -C "$repo" rev-parse "$commit")"
