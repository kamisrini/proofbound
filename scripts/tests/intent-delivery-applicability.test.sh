#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
repo=$tmp/repo
git init -q "$repo"
git -C "$repo" config user.email fixture@example.invalid
git -C "$repo" config user.name fixture
mkdir -p "$repo/kernel" "$repo/scripts"
cp scripts/intent-applicability.sh "$repo/scripts/"
printf 'x\n' >"$repo/kernel/main.go"
git -C "$repo" add .
git -C "$repo" commit -qm initial
missing=$(git -C "$repo" rev-parse HEAD)
if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$missing" >/dev/null 2>&1; then
  echo 'applicable commit without Intent trailer was accepted' >&2
  exit 1
fi

printf 'docs\n' >"$repo/README.md"
git -C "$repo" add .
git -C "$repo" commit -qm docs
non_applicable=$(git -C "$repo" rev-parse HEAD)
scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$non_applicable" >/dev/null

mkdir -p "$repo/scripts"
printf 'script\n' >"$repo/scripts/check.sh"
git -C "$repo" add .
git -C "$repo" commit -q -m applicable -m 'Intent: CI-fixture-item-acde12'
with_intent=$(git -C "$repo" rev-parse HEAD)
scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$with_intent" >/dev/null

# Shared corpus (DEF-INTENT-TRAILER-BLOCK-001): column 3 is "the kernel binds a claim". The same
# cases.tsv drives the kernel's end-to-end test (gitcmd_test.go
# TestIntentTrailers_SharedCorpusKernelBinds), so the gate and the kernel cannot disagree on any
# corpus input. Commits are built with hash-object from the exact .msg bytes: `git commit` and
# `git commit-tree` re-encode non-UTF-8 messages, so they are not byte-exact.
corpus=scripts/tests/fixtures/intent-trailers
raw_commit() { # $1 = message file; prints the new commit sha, parented on HEAD
  local tree parent
  git -C "$repo" add .
  tree=$(git -C "$repo" write-tree)
  parent=$(git -C "$repo" rev-parse HEAD)
  { printf 'tree %s\nparent %s\nauthor f <f@example.invalid> 1700000000 +0000\ncommitter f <f@example.invalid> 1700000000 +0000\n\n' "$tree" "$parent"; cat "$1"; } |
    git -C "$repo" hash-object -t commit -w --stdin
}
cases=0; passes=0
while IFS=$'\t' read -r name _ expect; do
  [[ -z $name || $name == \#* ]] && continue
  printf '%s\n' "$name" >"$repo/scripts/case-$name.sh"
  sha=$(raw_commit "$PWD/$corpus/$name.msg")
  git -C "$repo" update-ref HEAD "$sha"
  if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$sha" >/dev/null 2>&1; then got=pass; else got=fail; fi
  if [[ $got != "$expect" ]]; then
    echo "intent-delivery-applicability: corpus case $name: gate=$got, expected $expect" >&2
    exit 1
  fi
  [[ $got == pass ]] && passes=$((passes + 1))
  cases=$((cases + 1))
done <"$corpus/cases.tsv"
# Calibration: both outcomes must be exercised, or a constant answer would pass.
((cases >= 19 && passes >= 4 && cases - passes >= 10)) || { echo "intent-delivery-applicability: corpus too weak: cases=$cases passes=$passes" >&2; exit 1; }

# Diagnostics must name the real cause: body-paragraph (the a028984 shape) and CRLF.
for pair in "body-paragraph:OUTSIDE the final trailer block" "crlf:CRLF line endings"; do
  name=${pair%%:*}; want=${pair#*:}
  printf 'diag %s\n' "$name" >"$repo/scripts/diag-$name.sh"
  sha=$(raw_commit "$PWD/$corpus/$name.msg"); git -C "$repo" update-ref HEAD "$sha"
  if err=$(scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$sha" 2>&1); then
    echo "intent-delivery-applicability: $name was accepted" >&2; exit 1
  fi
  [[ $err == *"$want"* ]] || { echo "intent-delivery-applicability: wrong diagnostic for $name: $err" >&2; exit 1; }
done

# Replace refs must not change what the gate reads (gate review round 1, F3): replace a
# body-only commit with a correctly trailered one; the kernel ignores replace refs, so must the gate.
printf 'replaced\n' >"$repo/scripts/replaced.sh"
bad=$(raw_commit "$PWD/$corpus/body-paragraph.msg"); git -C "$repo" update-ref HEAD "$bad"
printf 'replacement\n' >"$repo/scripts/replacement.sh"
good=$(raw_commit "$PWD/$corpus/final-block.msg")
git -C "$repo" replace "$bad" "$good"
if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$bad" >/dev/null 2>&1; then
  echo 'intent-delivery-applicability: a replace ref made a body-only Intent commit pass' >&2; exit 1
fi
# ...including when the caller's environment points git elsewhere.
if GIT_DIR=/nonexistent scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$good" >/dev/null 2>&1; then :; else
  echo 'intent-delivery-applicability: an inherited GIT_DIR changed what the gate read' >&2; exit 1
fi

# Non-UTF-8 identity or path: the kernel rejects the whole commit (gate review round 2, F4).
printf 'id\n' >"$repo/scripts/latin1-author.sh"; git -C "$repo" add .
tree=$(git -C "$repo" write-tree); parent=$(git -C "$repo" rev-parse HEAD)
sha=$({ printf 'tree %s\nparent %s\nauthor J\351 <j@example.invalid> 1700000000 +0000\ncommitter f <f@example.invalid> 1700000000 +0000\n\n' "$tree" "$parent"; cat "$PWD/$corpus/final-block.msg"; } | git -C "$repo" hash-object -t commit -w --stdin)
git -C "$repo" update-ref HEAD "$sha"
if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$sha" >/dev/null 2>&1; then
  echo 'intent-delivery-applicability: a non-UTF-8 author name was accepted' >&2; exit 1
fi
printf 'p\n' >"$repo/scripts/path-"$'\351'".sh"
sha=$(raw_commit "$PWD/$corpus/final-block.msg"); git -C "$repo" update-ref HEAD "$sha"
if scripts/intent-delivery-applicability.sh --repo "$repo" --commit "$sha" >/dev/null 2>&1; then
  echo 'intent-delivery-applicability: a non-UTF-8 touched path was accepted' >&2; exit 1
fi
git -C "$repo" rm -q --cached "scripts/path-"$'\351'".sh"; rm "$repo/scripts/path-"$'\351'".sh"

# Default mode checks the whole unpushed range (gate review round 2, F2): a governed commit with a
# body-only Intent, buried under a non-governed commit, must fail the default run even though
# HEAD alone passes. Upstream = the last good commit.
printf 'base\n' >"$repo/scripts/range-base.sh"
base=$(raw_commit "$PWD/$corpus/final-block.msg"); git -C "$repo" update-ref HEAD "$base"
git -C "$repo" branch -q -f range-upstream "$base"
git -C "$repo" branch -q --set-upstream-to=range-upstream
printf 'buried\n' >"$repo/scripts/buried.sh"
buried=$(raw_commit "$PWD/$corpus/body-paragraph.msg"); git -C "$repo" update-ref HEAD "$buried"
printf 'docs on top\n' >"$repo/README.md"
top=$(raw_commit "$PWD/$corpus/prose-mention.msg"); git -C "$repo" update-ref HEAD "$top"
scripts/intent-delivery-applicability.sh --repo "$repo" --commit HEAD >/dev/null ||
  { echo 'intent-delivery-applicability: range probe is miscalibrated (HEAD should pass alone)' >&2; exit 1; }
if (cd "$repo" && "$OLDPWD/scripts/intent-delivery-applicability.sh") >/dev/null 2>&1; then
  echo 'intent-delivery-applicability: a governed commit buried under a non-governed one passed the default range' >&2; exit 1
fi
# Neutral control: once the buried commit is upstream ("pushed"), the default range excludes it.
git -C "$repo" branch -q -f range-upstream "$buried"
(cd "$repo" && "$OLDPWD/scripts/intent-delivery-applicability.sh") >/dev/null ||
  { echo 'intent-delivery-applicability: default range rechecked an already-pushed commit' >&2; exit 1; }

echo 'ok intent-delivery-applicability'
