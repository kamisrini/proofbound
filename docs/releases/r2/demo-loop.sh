#!/usr/bin/env bash
# R2 full warranty loop on a tiny project, for either kit:
#
#   bash docs/releases/r2/demo-loop.sh openspec [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-<kit>-loop
#   bash docs/releases/r2/demo-loop.sh speckit  [workdir]
#
# The walkthrough is docs/releases/r2/DEMO.md. In about two minutes it:
#   1. creates a project and runs `proofbound init <kit>`;
#   2. writes a spec (an OpenSpec change, or a Spec Kit feature) with two obligations;
#   3. implements it with a test, in a commit that claims the spec (`Intent: sdd:<name>`);
#   4. shows the claim bound and every obligation UNVERIFIED;
#   5. reviews it: runs the feature's test, records the outcomes, seals; the gate turns PASS;
#   6. rewords one obligation in the spec: `report drift` flags exactly that obligation.
# The demo plays the reviewer in step 5, and says so. In real use, a person or the
# proofbound-reviewer agent who did not write the commit makes those decisions.
# Needs git, go and jq; nothing is fetched or pushed. Needs the embedded database's port 55432 free.
set -euo pipefail

kit=${1:-}
[[ $kit == openspec || $kit == speckit ]] || { echo "usage: $0 {openspec|speckit} [workdir]" >&2; exit 2; }
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${2:-${TMPDIR:-/tmp}/proofbound-$kit-loop}
marker=.proofbound-loop-demo
command -v jq >/dev/null || { echo "demo-loop: jq is required" >&2; exit 1; }

if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "demo-loop: $work exists and is not a demo workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"
touch "$work/$marker"
unset DATABASE_URL

say() { printf '\n== %s\n' "$*"; }
pb() { "$work/proofbound" "$@"; }

say "building the proofbound CLI"
(cd "$repo/kernel" && go build -o "$work/proofbound" ./cmd/proofbound)

project="$work/project"
mkdir -p "$project"
cd "$project"
git init -q -b main
git config user.name spec-owner
git config user.email spec-owner@example.invalid

say "proofbound init $kit (the spec owner is 'spec-owner', from git config)"
pb init "$kit"
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude
PROOFBOUND_NO_INTENT=1 git commit -qm "add proofbound"

if [[ $kit == openspec ]]; then
  name=add-csv-export
  spec=openspec/changes/$name/specs/export/spec.md
  mkdir -p "$(dirname "$spec")"
  printf '## Why\nUsers need reports outside the app.\n' >"openspec/changes/$name/proposal.md"
  printf -- '- [ ] Implement CSV export\n' >"openspec/changes/$name/tasks.md"
  cat >"$spec" <<'EOF'
## ADDED Requirements

### Requirement: CSV export
The system SHALL export a report as CSV.

### Requirement: Header row
The system SHALL start every export with a header row.
EOF
  reword() { sed -i 's/The system SHALL start every export with a header row./The system SHALL start every export with a header row naming each column in lower case./' "$spec"; }
else
  name=001-csv-export
  spec=specs/$name/spec.md
  mkdir -p "specs/$name"
  cat >"$spec" <<'EOF'
# Feature Specification: CSV export

**Status**: Draft

## Requirements

### Functional Requirements

- **FR-001**: System MUST export a report as CSV.
- **FR-002**: System MUST start every export with a header row.
EOF
  printf -- '- [ ] T001 Implement CSV export\n' >"specs/$name/tasks.md"
  reword() { sed -i 's/- \*\*FR-002\*\*: System MUST start every export with a header row./- **FR-002**: System MUST start every export with a header row naming each column in lower case./' "$spec"; }
fi

say "the spec: $spec"
cat "$spec"
git add -A "${spec%%/*}"
PROOFBOUND_NO_INTENT=1 git commit -qm "spec: $name"

say "implement it, with a test, in a commit that claims the spec (Intent: sdd:$name)"
mkdir -p src
cat >src/export.sh <<'EOF'
#!/bin/sh
# Prints the report as CSV: a header row, then one row per item.
printf 'Item,Count\n'
printf 'apples,3\npears,5\n'
EOF
cat >src/export_test.sh <<'EOF'
#!/bin/sh
out=$(sh "$(dirname "$0")/export.sh")
[ "$(printf '%s\n' "$out" | head -1)" = "Item,Count" ] || { echo "FAIL: header row"; exit 1; }
[ "$(printf '%s\n' "$out" | wc -l)" -eq 3 ] || { echo "FAIL: CSV rows"; exit 1; }
echo "ok: CSV export with a header row"
EOF
git add src
git commit -qm "implement CSV export" -m "Intent: sdd:$name"
commit=$(git rev-parse HEAD)
echo "commit $commit"

say "sync: the spec's history, then the commits"
pb sync intent sdd
pb sync git

say "the claim is bound; every obligation is UNVERIFIED"
pb report intent "sdd:$name"

say "review (the demo plays the reviewer; in real use, someone who did not write the commit)"
pb review draft "$commit"
echo "-- the reviewer runs the feature's test:"
sh src/export_test.sh
draft=.proofbound/drafts/$commit/draft.json
jq '.declared_reviewer = "demo-reviewer"
    | .reviews[].obligations[].outcome = "VERIFIABLE"
    | .verdicts[].obligations[].outcome = "SATISFIED"' "$draft" >"$draft.tmp" && mv "$draft.tmp" "$draft"
echo "-- outcomes recorded: every obligation SATISFIED, every requirement VERIFIABLE, by demo-reviewer"
pb review seal "$(dirname "$draft")"
git add docs/verification/verdicts gates
PROOFBOUND_NO_INTENT=1 git commit -qm "review $commit"
pb sync git
pb sync reviews

say "the readiness gate for the reviewed commit"
pb gates canary

say "drift: the verified spec is current"
pb report drift

say "the spec owner rewords one obligation"
reword
git add -A "${spec%%/*}"
PROOFBOUND_NO_INTENT=1 git commit -qm "spec: reword the header-row obligation"
pb sync intent sdd
pb report drift

cat <<EOF

== what this shows ($kit)
- The commit claiming \`sdd:$name\` bound to the spec's exact revision; its obligations started
  UNVERIFIED.
- After a review by someone other than the spec owner, the commit's gate said PASS, with the counts.
- When the spec was reworded, \`report drift\` flagged exactly the reworded obligation: the warranty
  no longer covers what the spec promises, until someone re-verifies.
Workdir: $work (delete it when done).
EOF
