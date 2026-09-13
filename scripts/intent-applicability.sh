#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'usage: %s --repo <repository> --commit <commit>\n' "$0" >&2
  exit 2
}

repo=''
commit=''
while (($#)); do
  case $1 in
    --repo)
      (($# >= 2)) || usage
      repo=$2
      shift 2
      ;;
    --commit)
      (($# >= 2)) || usage
      commit=$2
      shift 2
      ;;
    *) usage ;;
  esac
done
[[ -n $repo && -n $commit ]] || usage
repo=$(cd "$repo" && pwd)
git -C "$repo" rev-parse --show-toplevel >/dev/null
git -C "$repo" cat-file -e "$commit^{commit}" 2>/dev/null || {
  printf 'intent-applicability: invalid commit %s\n' "$commit" >&2
  exit 1
}

matches() {
  local path=$1
  case $path in
    Makefile|check-windows.ps1|setup-windows.ps1|CLAUDE.md|docs/gates.md|docs/allowed-skips.txt|docs/allowed-survivors.txt|docs/decisions/INDEX.md|docs/invariants.lock|docs/laws.lock)
      return 0
      ;;
    .github/workflows/*|gates/*|scripts/*|tools/*)
      return 0
      ;;
    kernel/*)
      [[ $path =~ ^kernel/(.*/)?SPEC\.md$ ]] && return 1
      [[ $path =~ ^kernel/(.*/)?testdata(/|$) ]] && return 1
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

parents=$(git -C "$repo" rev-list --parents -n 1 "$commit")
read -r -a parent_words <<<"$parents"
parent_count=$((${#parent_words[@]} - 1))
diff_args=(--no-commit-id --name-status -r --find-renames --find-copies --find-copies-harder)
if ((parent_count == 0)); then
  diff_args+=(--root)
elif ((parent_count > 1)); then
  diff_args+=(-m)
fi

applicable=false
while IFS=$'\t' read -r status old_path new_path; do
  [[ -n ${status:-} ]] || continue
  case $status in
    R*|C*) paths=($old_path $new_path) ;;
    *) paths=($old_path) ;;
  esac
  for path in "${paths[@]}"; do
    [[ -n $path ]] || continue
    printf 'path=%s\n' "$path"
    if matches "$path"; then applicable=true; fi
  done
done < <(git -C "$repo" -c core.quotePath=false diff-tree "${diff_args[@]}" "$commit")

printf 'applicable=%s\n' "$applicable"
