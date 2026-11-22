#!/usr/bin/env bash
# MVP step 4: the shareable proof a release would attach, built locally (nothing is uploaded anywhere).
#
#   bash docs/releases/mvp/demo-share.sh [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-share-demo
#
# It runs the Release 2 loop with a reviewer who has their own key (so `review seal` signs), declares
# the exporter's key, runs `proofbound warranty bundle`, shows VERIFY.md, index.json and the badge, and
# then checks the bundled warranty the way a stranger would: from a fresh clone, pinning the exporter's
# fingerprint. Needs git, go, jq and the embedded database's port 55432 free.
set -euo pipefail
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${1:-${TMPDIR:-/tmp}/proofbound-share-demo}
marker=.proofbound-share-demo
if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "demo-share: $work exists and is not a demo workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"
touch "$work/$marker"
say() { printf '\n== %s\n' "$*"; }

say "Release 2 loop, with a reviewer who signs (docs/releases/r2/demo-loop.sh)"
DEMO_REVIEWER_KEY="$work/reviewer-key" bash "$repo/docs/releases/r2/demo-loop.sh" openspec "$work/loop" >"$work/loop.log" 2>&1 || { cat "$work/loop.log"; exit 1; }
grep -E "state=PASS" "$work/loop.log"
pb() { "$work/loop/proofbound" "$@"; }
export PATH="$work/loop:$PATH"  # the git hooks `init` installs run `proofbound` from PATH: use this build
cd "$work/loop/project"
export XDG_CONFIG_HOME="$work/config"

say "the exporter declares their key"
pb keys init --name spec-owner
git add .gitignore .proofbound/keys
git commit -qm "declare the exporter's key"
pb sync all >/dev/null

base=$(git log --format=%H --grep='^add proofbound' -1)
say "proofbound warranty bundle <first commit> -o release-warranties"
pb warranty bundle "$base" -o "$work/release-warranties"
ls "$work/release-warranties"

say "VERIFY.md (what a buyer reads)"
sed -n 1,14p "$work/release-warranties/VERIFY.md"

say "index.json and badge.svg"
jq -c '{unsigned, non_claiming_commits, entries: [.entries[] | {commit: .commit[0:12], state, reviewers_proven}]}' "$work/release-warranties/index.json"
grep -o '<title>[^<]*</title>' "$work/release-warranties/badge.svg"

say "a stranger checks it: fresh clone, exporter fingerprint pinned (obtained out of band)"
git clone -q "$work/loop/project" "$work/stranger-clone"
fpr=$(jq -r .exporter_key "$work/release-warranties/index.json")
file=$(ls "$work/release-warranties"/warranty-*.json | head -1)
pb warranty verify "$file" --repo "$work/stranger-clone" --trust-key "$fpr"

cat <<MSG

== what this shows
- One command turns every verified commit in a range into signed files plus a page and a badge a
  release can attach; this demo builds it locally and uploads nothing.
- The page and index say they are unsigned; the stranger trusts only \`warranty verify\` on a clone.
Workdir: $work (delete it when done).
MSG
