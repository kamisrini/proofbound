#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

script=$PWD/scripts/intent-applicability.sh
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

repo=$tmp/repo
git init -q "$repo"
git -C "$repo" config user.email fixture@example.invalid
git -C "$repo" config user.name fixture

commit_file() {
  local path=$1 body=${2:-x}
  mkdir -p "$(dirname "$repo/$path")"
  printf '%s\n' "$body" >"$repo/$path"
  git -C "$repo" add -A
  git -C "$repo" commit -qm "$path"
  git -C "$repo" rev-parse HEAD
}

assert_result() {
  local want=$1 commit=$2
  local output
  output=$("$script" --repo "$repo" --commit "$commit")
  printf '%s\n' "$output" | rg -q "^applicable=$want$"
}

assert_result true "$(commit_file kernel/main.go)"
assert_result false "$(commit_file kernel/internal/SPEC.md spec-only-content)"
assert_result false "$(commit_file kernel/testdata/vector.json vector-only-content)"
assert_result true "$(commit_file docs/gates.md)"
assert_result false "$(commit_file README.md readme-only-content)"

# Matching depends on changed paths, not content or the working tree.
commit=$(commit_file scripts/check.sh original)
printf 'changed after commit\n' >"$repo/scripts/check.sh"
assert_result true "$commit"

# Deletions remain applicable when the deleted path is included.
rm "$repo/scripts/check.sh"
git -C "$repo" add -A
git -C "$repo" commit -qm delete
assert_result true "$(git -C "$repo" rev-parse HEAD)"

# Renames and copies test both old and new paths.
printf 'rename\n' >"$repo/README.md"
git -C "$repo" add -A
git -C "$repo" commit -qm readme-reset
git -C "$repo" mv README.md kernel/renamed.go
git -C "$repo" commit -qm rename
assert_result true "$(git -C "$repo" rev-parse HEAD)"

cp "$repo/kernel/renamed.go" "$repo/ROADMAP.md"
git -C "$repo" add -A
git -C "$repo" commit -qm copy
assert_result true "$(git -C "$repo" rev-parse HEAD)"

if "$script" --repo "$repo" --commit deadbeef >/dev/null 2>&1; then
  echo 'invalid commit was accepted' >&2
  exit 1
fi

echo 'ok intent-applicability invariants'
