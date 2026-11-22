#!/usr/bin/env bash
# The full Proofbound tour (v0.3.2): every feature, on one small Spec Kit project, in order.
#
#   bash docs/releases/r3/tour.sh [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-tour
#   PROOFBOUND_BIN=$(command -v proofbound) bash docs/releases/r3/tour.sh   # use an installed binary
#
# The walkthrough that explains each chapter, and why it matters, is docs/releases/r3/DEMO.md.
# Chapters:
#   1. init: a Spec Kit project gets a profile, a commit hook, a reviewer agent and the PR workflow
#   2. a spec with two promises (obligations), and a commit that claims it
#   3. the record: the claim is bound to the exact spec revision; every promise starts UNVERIFIED
#   4. the pull-request check before review: NOT WARRANTED, and what to do next
#   5. identities: the spec owner and an independent reviewer each declare a signing key
#   6. no one grades their own work: the spec owner cannot review their own change
#   7. independent review, signed by the reviewer: the gate turns PASS
#   8. the pull-request check after review: WARRANTED (also as the Markdown the Action posts)
#   9. a portable signed warranty: VALID from a fresh clone, no database; edits and wrong keys fail
#  10. release proof: a bundle with signed files, a verification page, an index and a badge
#  11. drift: the spec changes after verification, and Proofbound says exactly which promise
#  12. the trust record is reproducible: rebuild every view, and get the same answers
# The demo plays both the spec owner and the reviewer, and says so; in real use the reviewer is a
# different person or a review agent with its own key. Needs git, jq, Go (unless PROOFBOUND_BIN is
# set) and the embedded database's port 55432. Nothing is fetched or pushed.
set -euo pipefail
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${1:-${TMPDIR:-/tmp}/proofbound-tour}
marker=.proofbound-tour
command -v jq >/dev/null || { echo "tour: jq is required" >&2; exit 1; }
if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "tour: $work exists and is not a tour workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"; touch "$work/$marker"
unset DATABASE_URL
export XDG_CONFIG_HOME="$work/config"   # the spec owner's private key lives here, never in git
reviewer_key="$work/reviewer.key"         # the reviewer's private key, never in git
chapter() { printf '\n\n======== %s\n' "$*"; }
step() { printf '\n$ %s\n' "$*"; }
pb() { "$work/proofbound" "$@"; }
export PATH="$work:$PATH"  # the git hooks `init` installs run `proofbound` from PATH: use this build
run() { step "proofbound $*"; pb "$@"; }
run_may_fail() { step "proofbound $*"; pb "$@" || echo "(exit $?)"; }

if [[ -n ${PROOFBOUND_BIN:-} ]]; then cp "$PROOFBOUND_BIN" "$work/proofbound"
else (cd "$repo/kernel" && go build -o "$work/proofbound" ./cmd/proofbound); fi
pb version

p="$work/shop"; mkdir -p "$p"; cd "$p"
git init -q -b main
git config user.name spec-owner; git config user.email spec-owner@example.invalid
nc() { git commit -qm "$@"; }

chapter "1. init: set up a Spec Kit project"
run init speckit
git add -A .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github
nc "add proofbound"
base=$(git rev-parse HEAD)
step "git ls-files"; git ls-files

chapter "2. a spec with two promises, and a commit that claims it"
mkdir -p specs/001-csv-export src
cat > specs/001-csv-export/spec.md <<'EOF'
# Feature Specification: CSV export
**Status**: Draft
## Requirements
### Functional Requirements
- **FR-001**: System MUST export a report as CSV.
- **FR-002**: System MUST start every export with a header row.
EOF
printf -- '- [ ] T001 Implement CSV export\n' > specs/001-csv-export/tasks.md
git add specs; nc "spec: 001-csv-export"
step "cat specs/001-csv-export/spec.md"; cat specs/001-csv-export/spec.md
cat > src/export.sh <<'EOF'
#!/bin/sh
# Prints the report as CSV: a header row, then one row per item.
printf 'item,count\n'
printf 'apples,3\npears,5\n'
EOF
git add src
step 'git commit -m "implement CSV export" -m "Intent: sdd:001-csv-export"'
git commit -qm "implement CSV export" -m "Intent: sdd:001-csv-export"
csv=$(git rev-parse HEAD); echo "commit $csv"

chapter "3. the record: the claim is bound to the exact spec revision; every promise is UNVERIFIED"
run sync all
run report intent sdd:001-csv-export

chapter "4. the pull-request check before review"
run_may_fail check "$base"

chapter "5. identities: the spec owner and the reviewer each declare a signing key"
run keys init --name spec-owner
step "PROOFBOUND_KEY=$reviewer_key proofbound keys init --name demo-reviewer"
PROOFBOUND_KEY=$reviewer_key pb keys init --name demo-reviewer
git add .gitignore .proofbound/keys; nc "declare the spec owner's and the reviewer's keys"
step "git ls-files .proofbound/keys"; git ls-files .proofbound/keys
pb sync all >/dev/null

chapter "6. no one grades their own work"
run review draft "$csv"
draft=.proofbound/drafts/$csv/draft.json
decide() { jq --arg who "$1" '.declared_reviewer = $who | .reviews[].obligations[].outcome = "VERIFIABLE" | .verdicts[].obligations[].outcome = "SATISFIED"' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"; }
decide spec-owner
echo "(the spec owner fills in the draft as reviewer of their own change)"
run_may_fail review seal "$draft"

chapter "7. independent review, signed by the reviewer: the gate turns PASS"
echo "-- the reviewer runs the code against each promise:"
step "sh src/export.sh"; sh src/export.sh
decide demo-reviewer
echo "(demo-reviewer records: both promises SATISFIED, both requirements VERIFIABLE)"
step "PROOFBOUND_KEY=$reviewer_key proofbound review seal $draft"
PROOFBOUND_KEY=$reviewer_key pb review seal "$draft"
git add docs/verification/verdicts gates; nc "review $(git rev-parse --short=12 "$csv")"
pb sync all >/dev/null
run gates canary

chapter "8. the pull-request check after review"
run check "$base"
step "proofbound check $base --markdown   # what the GitHub Action posts as the job summary"
pb check "$base" --markdown

chapter "9. a portable signed warranty"
run warranty export "$csv" -o "$work/warranty.json"
step "jq '{commit, change, promises}' warranty.json"
jq '{commit: .body.commit, change: .body.change_name, promises: [.body.requirements[].obligations[].statement]}' "$work/warranty.json"
run_may_fail warranty verify "$work/warranty.json"
git clone -q "$p" "$work/buyer-clone"
fpr=$(pb warranty verify "$work/warranty.json" --repo "$work/buyer-clone" | sed -n 's/.*key=\([0-9a-f]*\).*/\1/p')
echo "-- a buyer clones the repository and pins the exporter's fingerprint, obtained out of band ($fpr):"
run warranty verify "$work/warranty.json" --repo "$work/buyer-clone" --trust-key "$fpr"
sed 's/System MUST start every export with a header row./System MAY start an export with a header row./' "$work/warranty.json" > "$work/tampered.json"
echo "-- someone edits one promise in the file:"
run_may_fail warranty verify "$work/tampered.json" --repo "$work/buyer-clone"
echo "-- someone pins a different key:"
run_may_fail warranty verify "$work/warranty.json" --repo "$work/buyer-clone" --trust-key 0000000000000000

chapter "10. release proof: one bundle per release"
mkdir -p specs/002-totals-row
cat > specs/002-totals-row/spec.md <<'EOF'
# Feature Specification: Totals row
**Status**: Draft
## Requirements
### Functional Requirements
- **FR-001**: System MUST end every CSV export with a totals row.
- **FR-002**: System MUST sum the count column in the totals row.
EOF
printf -- '- [ ] T001 Add the totals row\n' > specs/002-totals-row/tasks.md
git add specs; nc "spec: 002-totals-row"
printf "printf 'total,8\\\\n'\n" >> src/export.sh
git add src; git commit -qm "add a totals row" -m "Intent: sdd:002-totals-row"
pb sync all >/dev/null
echo "(a second change claims its spec but nobody has reviewed it yet)"
run_may_fail warranty bundle "$base" -o "$work/release-warranties"
step "ls release-warranties"; ls "$work/release-warranties"
step "head -20 release-warranties/VERIFY.md"; sed -n 1,20p "$work/release-warranties/VERIFY.md"
step "badge.svg title"; grep -o '<title>[^<]*</title>' "$work/release-warranties/badge.svg"

chapter "11. drift: the spec changes after it was verified"
sed -i 's/- \*\*FR-002\*\*: System MUST start every export with a header row./- **FR-002**: System MUST start every export with a header row naming each column in lower case./' specs/001-csv-export/spec.md
git add specs; nc "spec: tighten the header-row promise"
pb sync all >/dev/null
run report drift

chapter "12. the trust record is reproducible"
run rebuild
echo "-- every view was dropped and re-derived from git and the event log; the same question gets the same answer:"
run_may_fail check "$base"

chapter "13. agents: the claim is added for them, and unclaimed code fails the check"
mid=$(git rev-parse HEAD)
printf 'echo hello\n' > src/hello.sh; git add src
step 'git commit -m "add a greeting"      # code, no claim, no task'
git commit -qm "add a greeting" || echo "(exit $?: the commit-msg hook refused it)"
run task start 002-totals-row
step 'git commit -m "add a greeting"      # the same commit, with a task active'
git commit -qm "add a greeting"
step "git log -1 --format=%B"; git log -1 --format=%B
run task stop
printf 'echo bye\n' >> src/hello.sh; git add src
step 'git commit --no-verify -m "sneak in a change"      # skips the hooks: no claim'
git commit -q --no-verify -m "sneak in a change"
pb sync all >/dev/null
run_may_fail check "$mid"

cat <<MSG


======== done
Every step above ran for real on $p.
The walkthrough explaining each chapter, and why it matters: docs/releases/r3/DEMO.md.
Workdir: $work (delete it when done).
MSG
