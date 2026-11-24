#!/usr/bin/env bash
# "The agent goes off-script": what Proofbound shows when a coding agent's change is not what its
# spec asked for, and how the redo earns a warranty.
#
#   bash docs/releases/r3/agent-demo.sh [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-agent-demo
#   PROOFBOUND_BIN=$(command -v proofbound) bash docs/releases/r3/agent-demo.sh   # an installed binary
#
# The walkthrough you type by hand, with what each step means, is docs/releases/r3/DEMO-AGENT.md.
# Acts:
#   0. a small shop project, a spec with two promises and a declared area (scope.paths)
#   1. the agent works: one commit that claims the spec, edits a file outside its area, and ships a
#      test that cannot fail; the check says NOT WARRANTED
#   2. an independent reviewer: the out-of-area line is named, the test is broken on purpose and still
#      passes, the review is sealed NEEDS_WORK; the Audit room and the release bundle show why
#   3. the redo from the spec: a test that fails first, a change inside its area, a signed review,
#      WARRANTED, and a bundle that verifies VALID against a clone
# The demo plays the agent (it pastes the agent's edits), the spec owner and the reviewer, and says so.
# Needs git, jq, curl, Go (unless PROOFBOUND_BIN is set) and port 55432. Nothing is fetched or pushed.
set -euo pipefail
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${1:-${TMPDIR:-/tmp}/proofbound-agent-demo}
marker=.proofbound-agent-demo
for tool in jq curl; do command -v "$tool" >/dev/null || { echo "agent-demo: $tool is required" >&2; exit 1; }; done
if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "agent-demo: $work exists and is not a demo workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"; touch "$work/$marker"
unset DATABASE_URL
export XDG_CONFIG_HOME="$work/config"   # the spec owner's private key lives here, never in git
reviewer_key="$work/reviewer.key"         # the reviewer's private key, never in git
act() { printf '\n\n======== %s\n' "$*"; }
step() { printf '\n$ %s\n' "$*"; }
pb() { "$work/proofbound" "$@"; }
export PATH="$work:$PATH"  # the git hooks `init` installs run `proofbound` from PATH: use this build
run() { step "proofbound $*"; pb "$@"; }
# agent_wrote <session> <file>...: what Claude Code's PostToolUse hook sends after the agent's Write of
# each file; `proofbound hook post-tool-use` logs a hash of each line (session-provenance-spec.md).
agent_wrote() {
  local session=$1 f; shift
  for f in "$@"; do
    jq -nc --arg s "$session" --arg cwd "$PWD" --arg f "$PWD/$f" --rawfile c "$f" \
      '{session_id:$s, cwd:$cwd, hook_event_name:"PostToolUse", tool_name:"Write", tool_input:{file_path:$f, content:$c}, tool_response:{type:"create"}}' |
      pb hook post-tool-use
  done
}
run_may_fail() { step "proofbound $*"; pb "$@" || echo "(exit $?)"; }
# page <port> <args...>: start `proofbound ui` on a port, fetch the page it links to, print its text.
page() {
  local port=$1; shift
  "$work/proofbound" ui --port "$port" "$@" > "$work/ui-$port.log" 2>&1 & local pid=$!
  local url="" i
  for i in $(seq 100); do url=$(sed -n 's/^ui \(http[^ ]*\).*/\1/p' "$work/ui-$port.log"); [[ -n $url ]] && break; sleep 0.1; done
  [[ -n $url ]] || { cat "$work/ui-$port.log"; kill "$pid"; return 1; }
  curl -s -c "$work/cookies-$port" -b "$work/cookies-$port" -L "$url" | sed 's/<[^>]*>/ /g; s/&#34;/"/g; s/&#39;/'"'"'/g; s/&lt;/</g; s/&gt;/>/g; s/  */ /g; s/^ //' | grep -v '^ *$' > "$work/page-$port.txt"
  kill "$pid"; wait "$pid" 2>/dev/null || true
}

if [[ -n ${PROOFBOUND_BIN:-} ]]; then cp "$PROOFBOUND_BIN" "$work/proofbound"
else (cd "$repo/kernel" && go build -o "$work/proofbound" ./cmd/proofbound); fi
pb version

act "0. the project, its spec and the area the spec allows"
p="$work/shop"; mkdir -p "$p/src"; cd "$p"
git init -q -b main
git config user.name spec-owner; git config user.email spec-owner@example.invalid
cat > src/pricing.sh <<'EOF'
#!/bin/sh
# The member discount, in percent.
echo 10
EOF
git add src; git commit -qm "pricing: member discount"
pb init speckit >/dev/null
git add -A .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github; git commit -qm "add proofbound"
pb keys init --name spec-owner >/dev/null
PROOFBOUND_KEY=$reviewer_key pb keys init --name demo-reviewer >/dev/null
printf '{"schema":"proofbound.tests.v1","run":"sh tests/export_test.sh","paths":["tests/"],"timeout_seconds":60}\n' > .proofbound/tests.json
git add .gitignore .proofbound/keys .proofbound/tests.json; git commit -qm "declare the keys, and how to run the tests"
base=$(git rev-parse HEAD)
mkdir -p specs/001-csv-export
cat > specs/001-csv-export/spec.md <<'EOF'
# Feature Specification: CSV export
**Status**: Draft
## Requirements
### Functional Requirements
- **FR-001**: System MUST export a report as CSV.
- **FR-002**: System MUST start every export with a header row.
EOF
printf -- '- [ ] T001 Implement CSV export\n' > specs/001-csv-export/tasks.md
printf '# The only files this change may touch.\nsrc/export.sh\ntests/\n' > specs/001-csv-export/scope.paths
git add specs; git commit -qm "spec: 001-csv-export"
step "cat specs/001-csv-export/scope.paths"; cat specs/001-csv-export/scope.paths
step "cat .proofbound/tests.json"; cat .proofbound/tests.json

act "1. the agent works"
git checkout -qb agent/csv-export
run task start 001-csv-export
echo "-- with Claude Code, the hook init wrote stops the agent's pricing edit before it happens:"
step "cat .claude/settings.json"; cat .claude/settings.json
step "echo '{\"cwd\":...,\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\".../src/pricing.sh\"}}' | proofbound hook pre-tool-use | jq ..."
printf '{"cwd":"%s","tool_name":"Edit","tool_input":{"file_path":"%s/src/pricing.sh"}}' "$p" "$p" | pb hook pre-tool-use | jq -r '.hookSpecificOutput | .permissionDecision + ": " + .permissionDecisionReason'
echo "(Claude Code blocks the edit and shows the agent this reason)"
printf '{"cwd":"%s","tool_name":"Write","tool_input":{"file_path":"%s/src/export.sh"}}' "$p" "$p" | pb hook pre-tool-use | wc -c | sed 's/^ *0$/(no output: src\/export.sh is inside the area, so the edit goes ahead)/' 
echo "-- this demo plays the agent in a shell, which no hook sees, so the pricing edit still lands; C10 catches it at review."
echo "-- the agent's two Writes go through Claude Code, whose PostToolUse hook logs a hash of each line; the pricing edit is a shell sed, which no hook sees"
mkdir -p tests
cat > src/export.sh <<'EOF'
#!/bin/sh
printf 'apples,3\npears,5\n'
EOF
cat > tests/export_test.sh <<'EOF'
#!/bin/sh
sh src/export.sh > /dev/null; echo "PASS: export works"
EOF
agent_wrote 7f3c2a91-agent-session src/export.sh tests/export_test.sh
step "sed -i 's/echo 10/echo 25/' src/pricing.sh   # the agent, through the shell"
sed -i.bak 's/echo 10/echo 25/' src/pricing.sh && rm -f src/pricing.sh.bak
step "git status --short"; git status --short
git add src tests
step 'git commit -m "implement CSV export" -m "Co-Authored-By: Claude <noreply@anthropic.com>"'
git commit -qm "implement CSV export" -m "Co-Authored-By: Claude <noreply@anthropic.com>"
step "git log -1 --format=%B"; git log -1 --format=%B
run task stop
agent=$(git rev-parse HEAD)
echo "-- the agent reports: tests pass"
step "sh tests/export_test.sh"; sh tests/export_test.sh
pb sync all >/dev/null
run_may_fail check "$base"

act "2. an independent reviewer"
run review draft "$agent"
echo "-- Proofbound already undid each hunk; the reviewer can confirm by hand: break the code, run the test:"
step "sh tests/export_test.sh"; sh tests/export_test.sh
cp src/export.sh "$work/export.sh.keep"
step "printf '#!/bin/sh\nexit 0\n' > src/export.sh   # an export that prints nothing"
printf '#!/bin/sh\nexit 0\n' > src/export.sh
step "sh tests/export_test.sh"; sh tests/export_test.sh
cp "$work/export.sh.keep" src/export.sh
echo "-- the test passes on code that does nothing: it cannot fail. And the real output has no header row:"
step "sh src/export.sh"; sh src/export.sh
draft=.proofbound/drafts/$agent/draft.json
jq '.declared_reviewer = "demo-reviewer"
  | .reviews[].obligations[].outcome = "VERIFIABLE"
  | .verdicts[].obligations |= map(.outcome = (if .obligation_id == "O-FR-002" then "NOT_SATISFIED" else "SATISFIED" end))
  | .out_of_area |= map(.decision = "reject")
  | .tests.hunks |= map(if .result == "not caught" then .decision = "reject" else . end)' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"
echo "(demo-reviewer records: FR-001 SATISFIED, FR-002 NOT_SATISFIED, the pricing line and both uncaught hunks rejected)"
step "PROOFBOUND_KEY=$reviewer_key proofbound review seal $draft"
PROOFBOUND_KEY=$reviewer_key pb review seal "$draft"
step "sed -n 2p docs/verification/verdicts/OV-*.md | jq -r .status"; sed -n 2p docs/verification/verdicts/OV-"${agent:0:12}"-1.md | jq -r .status
git add docs/verification/verdicts gates
git commit -qm "review $(git rev-parse --short=12 "$agent"): needs work"
pb sync all >/dev/null
run_may_fail check "$base"
step "proofbound ui --base $base   # the Audit room, as text"
page 47801 --base "$base"
grep -E '^[0-9]+ of [0-9]+ controls held|· control C|changed lines outside' "$work/page-47801.txt"
grep -A1 -E '^Claude ?$' "$work/page-47801.txt" | head -2
run_may_fail warranty bundle "$base" -o "$work/release-1"
step "proofbound ui --bundle release-1 --repo shop   # the bundle page, as text"
page 47802 --bundle "$work/release-1" --repo "$p"
grep -E '^NO WARRANTY|^VALID|BLOCKED|obligations verified' "$work/page-47802.txt"

act "3. the redo, from the spec"
git checkout -q main
git checkout -qb agent/csv-export-2
run task start 001-csv-export
mkdir -p tests
cat > tests/export_test.sh <<'EOF'
#!/bin/sh
# FR-002: the first line of every export is the header row.
first=$(sh src/export.sh | head -n 1)
[ "$first" = "item,count" ] || { echo "FAIL: first line is '$first', not the header row"; exit 1; }
echo "PASS: export starts with the header row"
EOF
cat > src/export.sh <<'EOF'
#!/bin/sh
printf 'apples,3\npears,5\n'
EOF
echo "-- test first: the new test fails on the old code"
step "sh tests/export_test.sh"; sh tests/export_test.sh || echo "(exit $?)"
cat > src/export.sh <<'EOF'
#!/bin/sh
# Prints the report as CSV: a header row, then one row per item.
printf 'item,count\n'
printf 'apples,3\npears,5\n'
EOF
step "sh tests/export_test.sh"; sh tests/export_test.sh
agent_wrote 2b9e04d6-agent-session src/export.sh tests/export_test.sh
git add src tests
step 'git commit -m "implement CSV export with a header row" -m "Co-Authored-By: Claude <noreply@anthropic.com>"'
git commit -qm "implement CSV export with a header row" -m "Co-Authored-By: Claude <noreply@anthropic.com>"
run task stop
redo=$(git rev-parse HEAD)
pb sync all >/dev/null
run review draft "$redo"
echo "-- the reviewer breaks the code on purpose: this time the test catches it"
cp src/export.sh "$work/export.sh.keep"
step "printf '#!/bin/sh\nexit 0\n' > src/export.sh"; printf '#!/bin/sh\nexit 0\n' > src/export.sh
step "sh tests/export_test.sh"; sh tests/export_test.sh || echo "(exit $?)"
cp "$work/export.sh.keep" src/export.sh
step "sh src/export.sh"; sh src/export.sh
draft=.proofbound/drafts/$redo/draft.json
jq '.declared_reviewer = "demo-reviewer" | .reviews[].obligations[].outcome = "VERIFIABLE" | .verdicts[].obligations[].outcome = "SATISFIED"' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"
step "PROOFBOUND_KEY=$reviewer_key proofbound review seal $draft"
PROOFBOUND_KEY=$reviewer_key pb review seal "$draft"
git add docs/verification/verdicts gates
git commit -qm "review $(git rev-parse --short=12 "$redo")"
pb sync all >/dev/null
run check "$base"
run warranty bundle "$base" -o "$work/release-2"
git clone -q -b agent/csv-export-2 "$p" "$work/buyer-clone"
step "proofbound ui --bundle release-2 --repo buyer-clone   # the bundle page, as text"
page 47803 --bundle "$work/release-2" --repo "$work/buyer-clone"
grep -E '^VALID|^NO WARRANTY|^Requirement|MUST' "$work/page-47803.txt"

cat <<MSG


======== done
Every step above ran for real on $p.
The rejected attempt stays on branch agent/csv-export, with its signed NEEDS_WORK review.
The walkthrough to type by hand: docs/releases/r3/DEMO-AGENT.md.
Workdir: $work (delete it when done).
MSG
