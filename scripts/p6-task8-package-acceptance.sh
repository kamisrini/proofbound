#!/usr/bin/env bash
set -euo pipefail

die() { echo "p6-task8-package-acceptance: $*" >&2; exit 1; }

mode=--check
package=''
root=''
while (($#)); do
  case $1 in
    --check) mode=--check ;;
    --package)
      (($# >= 2)) || die '--package requires a package path'
      mode=--package
      package=$2
      shift
      ;;
    --root)
      (($# >= 2)) || die '--root requires a directory'
      root=$2
      shift
      ;;
    *) die "unknown argument: $1" ;;
  esac
  shift
done

if [[ -z $root ]]; then root=$(git rev-parse --show-toplevel); fi
root=$(cd "$root" && pwd)
git -C "$root" rev-parse --git-dir >/dev/null 2>&1 || die "not a Git worktree: $root"

manifest=$root/docs/verification/p6-task8-package-results-e4c8e77.tsv
verdict=$root/docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md
expected_frozen=e4c8e77407699f7e089d5c1a2b3ce58df5871fbf
declare -A expected_candidates=(
  [internal/cli]=234
  [internal/connector/checks]=36
  [internal/connector/git]=34
  [internal/connector/git/gitcmd]=86
  [internal/connector/github]=58
  [internal/connector/intent]=93
  [internal/connector/intent/records]=52
  [internal/connector/intent/specdir]=47
  [internal/connector/reviews]=114
  [internal/connector/sessions]=62
  [internal/core]=58
  [internal/gates]=109
  [internal/projections]=404
  [internal/store]=133
  [internal/twin]=31
  [internal/migration]=33
)
tracked_path() {
  [[ -f $root/$1 ]] && git -C "$root" ls-files --error-unmatch -- "$1" >/dev/null 2>&1
}
tracked_path docs/verification/p6-task8-package-results-e4c8e77.tsv || die 'package result manifest must be tracked'
tracked_path docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md || die 'non-author verdict must be tracked'

frozen=$(awk -F '\t' '$1 == "# frozen-implementation" {print $2; exit}' "$manifest")
[[ $frozen =~ ^[0-9a-f]{40}$ ]] || die 'manifest has no exact frozen implementation commit'
[[ $frozen == "$expected_frozen" ]] || die "manifest is not bound to frozen Task 8 implementation $expected_frozen"
git -C "$root" cat-file -e "$frozen^{commit}" 2>/dev/null || die "frozen commit does not resolve: $frozen"
git -C "$root" cat-file -e "$frozen:kernel/internal" 2>/dev/null || die 'frozen commit has no kernel/internal tree'

status=$(awk '$1 == "status:" {print $2; exit}' "$verdict")
schema=$(awk '$1 == "schema:" {print $2; exit}' "$verdict")
reviewed=$(awk '$1 == "reviewed_commit:" {print $2; exit}' "$verdict")
reviewer=$(sed -n 's/^Reviewer identity: `\([^`]*\)`.*/\1/p' "$verdict" | head -1)
reviewed_author=$(sed -n 's/^Code author: `\([^`]*\)`.*/\1/p' "$verdict" | head -1)
reviewed_committer=$(sed -n 's/^Code committer: `\([^`]*\)`.*/\1/p' "$verdict" | head -1)
declared_path=$(awk '$1 == "artifact_path:" {print $2; exit}' "$verdict")
declared_digest=$(awk '$1 == "artifact_sha:" {print $2; exit}' "$verdict")
zero_digest=0000000000000000000000000000000000000000000000000000000000000000
actual_digest=$(sed -E "s/^artifact_sha: [0-9a-f]{64}$/artifact_sha: $zero_digest/" "$verdict" | sha256sum | awk '{print $1}')
actual_author=$(git -C "$root" show -s --format=%an "$frozen")
actual_committer=$(git -C "$root" show -s --format=%cn "$frozen")
actual_author_id=${actual_author//[[:space:]]/}
actual_committer_id=${actual_committer//[[:space:]]/}
reviewed_author_id=${reviewed_author//[[:space:]]/}
reviewed_committer_id=${reviewed_committer//[[:space:]]/}
[[ $schema == vera.verdict.v1 ]] || die 'verdict does not declare vera.verdict.v1'
[[ $status == ACCEPTABLE ]] || die "non-author verdict status is $status, not ACCEPTABLE"
[[ $reviewed == "$frozen" ]] || die 'verdict is not bound to the manifest frozen commit'
[[ $reviewer =~ ^[[:alnum:]_.-]+$ ]] || die 'reviewer identity is missing or malformed'
[[ $reviewer != "$actual_author_id" && $reviewer != "$actual_committer_id" && $reviewer != "$reviewed_author_id" && $reviewer != "$reviewed_committer_id" ]] || die 'reviewer identity equals the code author or committer'
[[ $reviewed_author == "$actual_author" ]] || die 'verdict code-author identity does not match the frozen commit'
[[ $reviewed_committer == "$actual_committer" ]] || die 'verdict code-committer identity does not match the frozen commit'
[[ $declared_path == docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md ]] || die 'verdict artifact path is not exact'
[[ $declared_digest =~ ^[0-9a-f]{64}$ && $actual_digest == "$declared_digest" ]] || die 'verdict self-digest is missing or invalid'

header='package tree candidates killed invalid survived calibration evidence'
actual_header=$(awk -F '\t' 'NR == 2 {gsub(/\t/, " "); print; exit}' "$manifest")
[[ $actual_header == "$header" ]] || die 'manifest header is invalid'
awk -F '\t' 'NR > 2 && $0 != "" && $1 !~ /^#/ && NF != 8 {exit 1}' "$manifest" || die 'manifest row does not have exactly eight fields'

declare -A tree_by_package=() candidate_by_package=() evidence_by_package=()
count=0
while IFS=$'\t' read -r pkg tree candidates killed invalid survived calibration evidence extra; do
  [[ -n $pkg ]] || continue
  [[ $pkg == \#* ]] && continue
  [[ $pkg != package ]] || continue
  [[ -z ${extra:-} ]] || die "manifest row has extra fields: $pkg"
  [[ $pkg == internal/* && $pkg != *..* ]] || die "invalid package path: $pkg"
  [[ -z ${tree_by_package[$pkg]+x} ]] || die "duplicate package row: $pkg"
  [[ -n ${expected_candidates[$pkg]+x} ]] || die "manifest includes unexpected package: $pkg"
  [[ $tree =~ ^[0-9a-f]{40}$ ]] || die "invalid package tree object: $pkg"
  actual_tree=$(git -C "$root" rev-parse "$frozen:kernel/$pkg" 2>/dev/null) || die "package tree is missing at frozen commit: $pkg"
  [[ $tree == "$actual_tree" ]] || die "package tree mismatch: $pkg"
  [[ $candidates =~ ^[0-9]+$ && $killed =~ ^[0-9]+$ && $invalid =~ ^[0-9]+$ && $survived =~ ^[0-9]+$ ]] || die "non-numeric mutation count: $pkg"
  [[ $candidates == "${expected_candidates[$pkg]}" ]] || die "candidate universe mismatch: $pkg"
  ((candidates > 0)) || die "package has no declared mutation candidates: $pkg"
  ((candidates == killed + invalid + survived)) || die "mutation counts do not add up: $pkg"
  ((invalid == 0 && survived == 0 && killed == candidates)) || die "package is below the zero-invalid, zero-survivor bar: $pkg"
  [[ $calibration == 'neutral=survived;invalid=invalid;lethal=killed' ]] || die "calibration controls are not declared exactly: $pkg"
  tracked_path "$evidence" || die "mutation source evidence is not tracked: $pkg: $evidence"
  rg -Fq "| $pkg | $tree | ACCEPTABLE |" "$verdict" || die "verdict does not bind package tree: $pkg"
  tree_by_package[$pkg]=$tree
  candidate_by_package[$pkg]=$candidates
  evidence_by_package[$pkg]=$evidence
  count=$((count + 1))
done <"$manifest"

mapfile -t expected_packages < <(
  git -C "$root" ls-tree -r --name-only "$frozen" -- kernel/internal |
    awk '$0 ~ /\.go$/ && $0 !~ /_test\.go$/ {sub(/^kernel\//, ""); sub(/\/[^/]+$/, ""); print}' |
    LC_ALL=C sort -u
)
[[ ${#expected_packages[@]} -eq 16 ]] || die "frozen production package count is ${#expected_packages[@]}, expected 16"
[[ ${#expected_candidates[@]} -eq ${#expected_packages[@]} ]] || die 'checker candidate registry does not cover the frozen package universe'
[[ $count -eq ${#expected_packages[@]} ]] || die "manifest package count is $count, expected ${#expected_packages[@]}"
for pkg in "${expected_packages[@]}"; do
  [[ -n ${tree_by_package[$pkg]+x} ]] || die "production package is missing from manifest: $pkg"
done
for pkg in "${!tree_by_package[@]}"; do
  [[ " ${expected_packages[*]} " == *" $pkg "* ]] || die "manifest includes non-production package: $pkg"
done

if [[ $mode == --package ]]; then
  [[ -n ${tree_by_package[$package]+x} ]] || die "package is not accepted: $package"
  printf 'PASS %s tree=%s candidates=%s verdict=ACCEPTABLE\n' "$package" "${tree_by_package[$package]}" "${candidate_by_package[$package]}"
else
  printf 'PASS %d production packages; frozen=%s; reviewer=%s\n' "$count" "$frozen" "$reviewer"
fi
