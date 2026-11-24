# The agent goes off-script: a hands-on demo

A coding agent is asked to build one small feature from a spec. It finishes in seconds, says the tests
pass, and they do. It also changed a price it was never asked to touch, left out one of the two things
the spec promised, and wrote a test that cannot fail. At agent speed, nobody reads every line, so this
is the change that gets merged.

This page shows what Proofbound does with it:
- with Claude Code, it stops an edit outside the area the spec allowed, before it happens;
- when an edit gets through anyway, it names the exact line;
- it undoes each part of the change and shows that the agent's test notices none of them;
- an independent reviewer decides on each finding, and signs it;
- the pull-request check and the release bundle refuse to warrant the change;
- the redo earns a warranty that anyone can verify from a clone.

You type every command. It takes about 15 minutes.

**How to read the blocks.** A block marked `bash` is safe to paste into your terminal as it is. A block
marked `text` is output, for comparison only; never paste it. Your commit hashes and key fingerprints
will differ from the ones shown; everything else should match.

**Three hats.** You play three parts:
- **the agent:** you paste its edits. With Claude Code, the same edits and the same commit come from
  the agent itself, through the same hooks;
- **the spec owner,** who wrote the spec;
- **an independent reviewer:** in real use a different person, or a review agent with its own key.

Their private keys are kept in separate files so you can see that they are separate.

## Before you start

You need Proofbound v0.3.2 on your PATH, plus `git` and `jq`. If you installed it before 2026-11-24,
run the install command again: the last step needs a fix that was added to v0.3.2 on 2026-11-23. Port 55432 must be free
(Proofbound starts a private database there for each command and stops it afterwards).

```bash
proofbound version
```

If that says `command not found`, install it with the one-line command in the [README](README.md), then
run `export PATH="$HOME/.local/bin:$PATH"; hash -r`.

---

## Step 1 — A project that already has code

**What you're doing.** Making a small `shop` repository with one existing file, `src/pricing.sh`, which
prints the member discount (10%). It is committed before Proofbound is set up, as in any real project.
`XDG_CONFIG_HOME` keeps the spec owner's private key inside the demo folder, so the demo cannot touch any
key you already have.

```bash
export W="$HOME/projects/proofbound-agent-demo"
export XDG_CONFIG_HOME="$W/config"
unset DATABASE_URL
mkdir -p "$W/shop/src"
cd "$W/shop"
git init -q -b main
git config user.name spec-owner
git config user.email spec-owner@example.invalid
printf '#!/bin/sh\n# The member discount, in percent.\necho 10\n' > src/pricing.sh
git add src
git commit -qm "pricing: member discount"
sh src/pricing.sh
```

```text
10
```

**If you open a new terminal later,** paste this to pick up where you left off:

```bash
export W="$HOME/projects/proofbound-agent-demo"
export XDG_CONFIG_HOME="$W/config"
unset DATABASE_URL
cd "$W/shop"
```

## Step 2 — Set up Proofbound, and two keys

**What you're doing.**
- `init speckit` adds Proofbound's profile, a commit hook, a reviewer agent and the pull-request
  workflow.
- The spec owner and the reviewer each make a signing key. Only the public halves are committed.
- `.proofbound/tests.json` says how to run this project's tests. With it, Proofbound can undo each
  part of a change and check that a test notices.
- The tag `demo-base` marks where the agent's work will start: every check below asks "is everything
  since `demo-base` warranted?"

```bash
proofbound init speckit > /dev/null
git add -A .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github
git commit -qm "add proofbound"
proofbound keys init --name spec-owner > /dev/null
PROOFBOUND_KEY="$W/reviewer.key" proofbound keys init --name demo-reviewer > /dev/null
printf '{"schema":"proofbound.tests.v1","run":"sh tests/export_test.sh","paths":["tests/"],"timeout_seconds":60}\n' > .proofbound/tests.json
git add .gitignore .proofbound/keys .proofbound/tests.json
git commit -qm "declare the keys, and how to run the tests"
git tag demo-base
git ls-files .proofbound/keys
```

```text
.proofbound/keys/demo-reviewer.pub
.proofbound/keys/spec-owner.pub
```

## Step 3 — The spec, and the area it allows

**What you're doing.** Writing a spec with two promises: export a report as CSV (FR-001), and start
every export with a header row (FR-002). Next to it, `scope.paths` lists the only files this change may
touch: the export script and the tests.

```bash
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
git add specs
git commit -qm "spec: 001-csv-export"
```

Nothing is printed.

**Why it matters.** The spec is committed **before** any code, so the agent's permission exists before
its work, and the agent cannot widen its own area: the area is read from the commit before the agent's
commit.

---

## Act 1 — The agent works

## Step 4 — The agent starts its task

**What you're doing.** The agent works on a branch and tells Proofbound which spec it is working on.
From now on, every code commit it makes ends with `Intent: sdd:001-csv-export`, added by the hook, so
the agent cannot forget to say what it was doing.

```bash
git checkout -qb agent/csv-export
proofbound task start 001-csv-export
```

```text
task sdd:001-csv-export: commits that change code now end with `Intent: sdd:001-csv-export`; `proofbound task stop` when done
```

**With Claude Code, an edit outside the area is stopped before it happens.** `init` wrote
`.claude/settings.json`, a hook that Claude Code runs before every file edit. Here you hand that hook
exactly what Claude Code would send when the agent tries to edit `src/pricing.sh`, and then
`src/export.sh`:

```bash
cat .claude/settings.json
printf '{"cwd":"%s","tool_name":"Edit","tool_input":{"file_path":"%s/src/pricing.sh"}}' "$PWD" "$PWD" | proofbound hook pre-tool-use | jq -r '.hookSpecificOutput | .permissionDecision + ": " + .permissionDecisionReason'
printf '{"cwd":"%s","tool_name":"Write","tool_input":{"file_path":"%s/src/export.sh"}}' "$PWD" "$PWD" | proofbound hook pre-tool-use | wc -c
```

```text
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit|NotebookEdit",
        "hooks": [ { "type": "command", "command": "proofbound hook pre-tool-use || true" } ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [ { "type": "command", "command": "proofbound hook post-tool-use || true" } ]
      }
    ]
  }
}
deny: proofbound: src/pricing.sh is outside the area the spec for sdd:001-csv-export allows (specs/001-csv-export/scope.paths: src/export.sh tests/). Do not edit it for this task. If the change truly needs it, stop and ask the spec owner to widen scope.paths and commit that first; otherwise the reviewer must accept or reject every line outside the area (control C10).
0
```

**What just happened.**
- **The pricing edit:** the hook answers in Claude Code's format, with permission `deny` and the
  reason. Claude Code blocks the edit, and the agent reads the reason instead.
- **The export edit:** it is inside the area, so the hook prints nothing (`0` bytes) and the edit
  goes ahead.
- **The `|| true`:** if `proofbound` is missing, too old or broken, the hook stays out of the way
  rather than blocking every edit.
- **The second hook, `PostToolUse`:** after each edit the agent makes through Claude Code, it logs a
  hash of every line the edit added (never the content), in the repository's git folder. Review uses
  it in Step 8.

**Why it matters, and its limit.** Stopping an out-of-scope edit while the agent works is cheaper
than finding it in review. But the agent controls its own machine: it could change a file through a
shell command, or remove the hook. So this is early feedback, not the guarantee. In this demo you play
the agent in a shell, which no hook sees, so the next step makes the pricing edit anyway, and C10
catches it at review.

## Step 5 — The agent's edits (you paste them)

**What you're doing.** Playing the agent. It writes the export, writes a test, and, "while it was
there", raises the member discount from 10% to 25%. Look closely at all three:
- the export has no header row;
- the test runs the export but ignores what it prints, and always says PASS;
- `src/pricing.sh` is not in the spec's area.

The agent writes the two new files with Claude Code's Write tool, so its `PostToolUse` hook logs them.
The second command hands the hook what Claude Code sends after each Write. The pricing edit is a
shell `sed`, which no hook sees:

```bash
mkdir -p tests
printf '#!/bin/sh\nprintf "apples,3\\npears,5\\n"\n' > src/export.sh
printf '#!/bin/sh\nsh src/export.sh > /dev/null; echo "PASS: export works"\n' > tests/export_test.sh
for f in src/export.sh tests/export_test.sh; do jq -nc --arg cwd "$PWD" --arg f "$PWD/$f" --rawfile c "$f" '{session_id:"7f3c2a91-agent-session", cwd:$cwd, tool_name:"Write", tool_input:{file_path:$f, content:$c}, tool_response:{type:"create"}}' | proofbound hook post-tool-use; done
sed -i.bak 's/echo 10/echo 25/' src/pricing.sh && rm -f src/pricing.sh.bak
git status --short
```

```text
 M src/pricing.sh
?? src/export.sh
?? tests/
```

## Step 6 — The agent commits, and says the tests pass

**What you're doing.** The agent commits all three files in one commit, naming itself as co-author, as
Claude Code does. Then it runs its test.

```bash
git add src tests
git commit -qm "implement CSV export" -m "Co-Authored-By: Claude <noreply@anthropic.com>"
git tag attempt-1
proofbound task stop
git log -1 --format=%B
sh tests/export_test.sh
```

```text
no task
implement CSV export

Co-Authored-By: Claude <noreply@anthropic.com>
Intent: sdd:001-csv-export

PASS: export works
```

**What just happened.** The hook added the `Intent:` line. The commit claims the spec, and the test
passes. To a busy person, that looks done.

## Step 7 — The pull-request check, before anyone reviews

**What you're doing.** Running the check the GitHub Action runs on every pull request.

```bash
proofbound sync all > /dev/null
proofbound check demo-base
```

```text
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)
  9c887d3d5e67 sdd:001-csv-export NO REVIEW  no readiness gate gates/ready-9c887d3d5e67.yaml; a reviewer runs `proofbound review draft 9c887d3d5e67`
Drift:
  summary warranted=0 current=0 stale=0 other=0
proofbound: check: not warranted
```

**Why it matters.** "The agent says the tests pass" is not evidence. Nothing is warranted until someone
other than the agent and the spec owner has checked each promise.

---

## Act 2 — An independent reviewer

## Step 8 — The review draft: a line outside the area, and a test that notices nothing

**What you're doing.** The reviewer asks Proofbound for a review draft of the agent's commit. The
draft lists each promise to decide and, from git, every changed line outside the spec's area. Then,
because the project declares its tests, Proofbound undoes each part of the agent's change, one at a
time and then all at once, runs the tests each time, and reports whether they noticed. It does this
in a temporary copy, so your files are not touched. It runs the commit's own code, so do it only
where you would run that commit's tests anyway.

```bash
proofbound review draft attempt-1
```

```text
tests: running `sh tests/export_test.sh` with each hunk of code undone (this runs the commit's own code)
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
outside area src/pricing.sh line 3
decide each in out_of_area: "accept" with a reason, or "reject"
tests did NOT catch undoing src/export.sh lines 1-2
tests did NOT catch undoing src/pricing.sh line 3
tests did NOT catch undoing the whole change
decide each hunk not caught in tests.hunks: "accept" with a reason, or "reject"
provenance src/export.sh: 2 by Claude Code session 7f3c2a91, 0 with no recorded agent edit, 0 common
provenance src/pricing.sh: 0 by a session, 1 with no recorded agent edit (lines 3), 0 common
provenance tests/export_test.sh: 2 by Claude Code session 7f3c2a91, 0 with no recorded agent edit, 0 common
provenance: docs, Markdown and records are not traced; lines moved by a shell or rewritten by a formatter show as no recorded agent edit
draft /home/you/projects/proofbound-agent-demo/shop/.proofbound/drafts/9c887d3d5e6781f274895620b742e00494926261/draft.json
```

**What just happened.** Three things, all found by Proofbound on its own:
- **The discount change:** `src/pricing.sh line 3` is outside the spec's area.
- **The test:** with the export removed, the price change removed, or the whole change removed, the
  agent's test still passes. It notices nothing.
- **Where each line came from:** the export and the test were written by Claude Code session
  `7f3c2a91`, line for line. The pricing line has **no recorded agent edit**: it reached the commit
  around the hooks, the edit the perimeter could not see. Proofbound says only what the logs show.
  It does not guess who wrote the line.

The review cannot be sealed until the reviewer decides on each finding: accept it with a reason, or
reject it.

**Why it matters.** In a large agent diff, a one-line price change hides easily, and so does a test
that only looks like a test. Here both are named, with file and line, and each needs a signed
decision.

## Step 9 — Confirm it by hand

**What you're doing.** The reviewer checks Proofbound's finding by hand: run the test, break the code
on purpose (an export that prints nothing), and run it again. Then look at the real output.

```bash
sh tests/export_test.sh
cp src/export.sh "$W/export.sh.keep"
printf '#!/bin/sh\nexit 0\n' > src/export.sh
sh tests/export_test.sh
cp "$W/export.sh.keep" src/export.sh
sh src/export.sh
```

```text
PASS: export works
PASS: export works
apples,3
pears,5
```

**What just happened.** The test passed on code that does nothing, as Proofbound said. The real output
has no header row, so FR-002 is not met.

**Be clear about what was checked.** Proofbound undid each *part* of the change. A test can notice a
removed part and still miss a subtle bug inside it, so this is not full mutation testing, and the
reviewer still judges each promise. What Proofbound adds is that "the agent's tests pass" is no longer
taken on trust.

## Step 10 — The reviewer records the findings and signs them

**What you're doing.** Filling in the draft:
- both requirements are clear enough to check (VERIFIABLE);
- FR-001 is SATISFIED and FR-002 is NOT_SATISFIED;
- the pricing line is rejected;
- both hunks the tests did not catch are rejected.

Seal runs the tests and the provenance match again itself, and takes only the reviewer's decisions
from the draft.

Then the reviewer seals it with their own key and commits the signed records.

```bash
draft=.proofbound/drafts/$(git rev-parse attempt-1)/draft.json
jq '.declared_reviewer = "demo-reviewer"
  | .reviews[].obligations[].outcome = "VERIFIABLE"
  | .verdicts[].obligations |= map(.outcome = (if .obligation_id == "O-FR-002" then "NOT_SATISFIED" else "SATISFIED" end))
  | .out_of_area |= map(.decision = "reject")
  | .tests.hunks |= map(if .result == "not caught" then .decision = "reject" else . end)' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"
PROOFBOUND_KEY="$W/reviewer.key" proofbound review seal "$draft"
git add docs/verification/verdicts gates
git commit -qm "review attempt-1: needs work"
sed -n 2p docs/verification/verdicts/OV-*.md | jq -r .status
```

```text
tests: running `sh tests/export_test.sh` with each hunk of code undone (this runs the commit's own code)
wrote docs/verification/verdicts/RR-9c887d3d5e67-1.md
wrote docs/verification/verdicts/RR-9c887d3d5e67-1.md.sig
wrote docs/verification/verdicts/OV-9c887d3d5e67-1.md
wrote docs/verification/verdicts/OV-9c887d3d5e67-1.md.sig
wrote docs/verification/verdicts/AREA-9c887d3d5e67.json
wrote docs/verification/verdicts/TESTS-9c887d3d5e67.json
wrote docs/verification/verdicts/PROV-9c887d3d5e67.json
wrote gates/ready-9c887d3d5e67.yaml
signed as demo-reviewer (key 1f9834b2f85b4562); commit .proofbound/keys/demo-reviewer.pub with the review if it is not committed yet
NEEDS_WORK
```

**What just happened.** The verdict is NEEDS_WORK. Two signed records say what was decided and by
whom:
- `AREA-….json`: the pricing line was rejected;
- `TESTS-….json`: which parts the tests did not catch, and that each was rejected.
- `PROV-….json`: which session wrote which lines, and the pricing line with no recorded agent edit.

## Step 11 — The check after review

```bash
proofbound sync all > /dev/null
proofbound check demo-base
```

```text
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)
  9c887d3d5e67 sdd:001-csv-export BLOCKED  1 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
Drift:
  summary warranted=0 current=0 stale=0 other=0
proofbound: check: not warranted
```

**Why it matters.** The pull request stays red, with the reason, until the work is fixed.

## Step 12 — The Audit room: what an auditor sees

**What you're doing.** Opening the Audit room in your browser for everything since `demo-base`.
Proofbound prints a link; open it (or add `--open`). Press Ctrl-C in the terminal when you are done.

```bash
proofbound ui --base demo-base
```

```text
ui http://127.0.0.1:…/?k=… (Ctrl-C to stop; the link is your key to this page, do not share it)
```

**What to look for on the page:**
- "5 of 8 controls held for this range. 3 exception(s), 0 with a signed outcome."
- **Unreviewed change · control C1**: the commit claims the spec but is BLOCKED.
- **Change outside its declared area · control C10**: "implement CSV export" changed lines outside its
  spec's area (not accepted): `src/pricing.sh line 3`.
- **Change its tests do not catch · control C11**: the tests do not catch undoing the whole change,
  nor `src/export.sh` or `src/pricing.sh`.
- **Which agent sessions wrote its lines · control C12** (listed): 4 lines by session 7f3c2a91, and 1
  with no recorded agent edit in `src/pricing.sh`. The counts are disclosed and never make an
  exception, because people and generators write lines too. Only a record that does not verify, or
  that the agent signed itself, is an exception.
- **Who did the work**: Claude, as an Agent, with one commit and three open exceptions.
- **Maps to** under each control, for example C11 → NIST SSDF PW.8 (perform the testing and document
  the results).
- **Who reviewed each change**: demo-reviewer, independent (Yes), with both review signatures checked
  in your browser.

**Why it matters.** This is the one page an auditor or a team lead needs: which agent changed what,
whether it stayed inside what was authorized, who checked it, and what is still open. Every line links
to the git evidence behind it.

The Agent row comes from git's author and `Co-Authored-By` lines, which anyone can type. The page says
so. To prove who made a commit, sign commits with `proofbound keys git-sign` (control C9).

## Step 13 — The release bundle: no warranty

**What you're doing.** Building the release evidence a customer or auditor would receive, and opening
it as they would: from the folder alone, checked against a clone.

```bash
proofbound warranty bundle demo-base -o "$W/release-1"
```

```text
not warranted 9c887d3d5e67 sdd:001-csv-export BLOCKED  1 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
bundle /home/you/projects/proofbound-agent-demo/release-1 (456aada51336..3114d958afbb): 0 of 1 claims verified (0 VALID, 0 UNREVIEWED); VERIFY.md, index.json, badge.svg
proofbound: bundle: written, but not every claiming commit is VALID
```

```bash
proofbound ui --bundle "$W/release-1" --repo "$W/shop"
```

On the page: **NO WARRANTY · 1**, and "The unsigned index says this claiming commit has no warranty:
BLOCKED." Press Ctrl-C when done.

**Why it matters.** The agent's change cannot ship as warranted. The release evidence says so to anyone
who receives it, without them needing to trust you, your database or your CI.

---

## Act 3 — The redo

## Step 14 — Start again from the spec, test first

**What you're doing.**
- The rejected branch is left as it is, as the record of the first attempt.
- The agent starts a new branch from `main`, where the spec is, and writes the test first: the first
  line of the export must be the header row.
- It runs the test against an export without the header row, to see it fail.

```bash
git checkout -q main
git checkout -qb agent/csv-export-2
proofbound task start 001-csv-export
mkdir -p tests src
cat > tests/export_test.sh <<'EOF'
#!/bin/sh
# FR-002: the first line of every export is the header row.
first=$(sh src/export.sh | head -n 1)
[ "$first" = "item,count" ] || { echo "FAIL: first line is '$first', not the header row"; exit 1; }
echo "PASS: export starts with the header row"
EOF
printf '#!/bin/sh\nprintf "apples,3\\npears,5\\n"\n' > src/export.sh
sh tests/export_test.sh
```

```text
task sdd:001-csv-export: commits that change code now end with `Intent: sdd:001-csv-export`; `proofbound task stop` when done
FAIL: first line is 'apples,3', not the header row
```

## Step 15 — Implement it, inside the area, and commit

```bash
cat > src/export.sh <<'EOF'
#!/bin/sh
# Prints the report as CSV: a header row, then one row per item.
printf 'item,count\n'
printf 'apples,3\npears,5\n'
EOF
sh tests/export_test.sh
for f in src/export.sh tests/export_test.sh; do jq -nc --arg cwd "$PWD" --arg f "$PWD/$f" --rawfile c "$f" '{session_id:"2b9e04d6-agent-session", cwd:$cwd, tool_name:"Write", tool_input:{file_path:$f, content:$c}, tool_response:{type:"create"}}' | proofbound hook post-tool-use; done
git add src tests
git commit -qm "implement CSV export with a header row" -m "Co-Authored-By: Claude <noreply@anthropic.com>"
git tag attempt-2
proofbound task stop
```

```text
PASS: export starts with the header row
no task
```

## Step 16 — The reviewer checks the redo the same way

**What you're doing.** The same review as before:
- the draft now reports that every changed line is inside the area, and that the new lines were written by the redo's session;
- Proofbound reports that undoing the export makes the tests fail, so the test depends on the change;
- the reviewer's own break-it check agrees;
- the output has its header row.

Then the reviewer records every promise SATISFIED, signs, and commits.

```bash
proofbound sync all > /dev/null
proofbound review draft attempt-2
cp src/export.sh "$W/export.sh.keep"
printf '#!/bin/sh\nexit 0\n' > src/export.sh
sh tests/export_test.sh
cp "$W/export.sh.keep" src/export.sh
sh src/export.sh
```

```text
tests: running `sh tests/export_test.sh` with each hunk of code undone (this runs the commit's own code)
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
area: every changed line is inside the declared area
tests caught undoing src/export.sh lines 1-4
tests caught undoing the whole change
provenance src/export.sh: 4 by Claude Code session 2b9e04d6, 1 by Claude Code session 7f3c2a91, 0 with no recorded agent edit, 0 common
provenance tests/export_test.sh: 5 by Claude Code session 2b9e04d6, 1 by Claude Code session 7f3c2a91, 0 with no recorded agent edit, 0 common
provenance: docs, Markdown and records are not traced; lines moved by a shell or rewritten by a formatter show as no recorded agent edit
draft /home/you/projects/proofbound-agent-demo/shop/.proofbound/drafts/b1ecdc3ed0cb20bb8e7dcae9e5ef6be205449e5e/draft.json
FAIL: first line is '', not the header row
item,count
apples,3
pears,5
```

```bash
draft=.proofbound/drafts/$(git rev-parse attempt-2)/draft.json
jq '.declared_reviewer = "demo-reviewer" | .reviews[].obligations[].outcome = "VERIFIABLE" | .verdicts[].obligations[].outcome = "SATISFIED"' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"
PROOFBOUND_KEY="$W/reviewer.key" proofbound review seal "$draft"
git add docs/verification/verdicts gates
git commit -qm "review attempt-2"
```

```text
tests: running `sh tests/export_test.sh` with each hunk of code undone (this runs the commit's own code)
wrote docs/verification/verdicts/RR-b1ecdc3ed0cb-1.md
wrote docs/verification/verdicts/RR-b1ecdc3ed0cb-1.md.sig
wrote docs/verification/verdicts/OV-b1ecdc3ed0cb-1.md
wrote docs/verification/verdicts/OV-b1ecdc3ed0cb-1.md.sig
wrote docs/verification/verdicts/TESTS-b1ecdc3ed0cb.json
wrote docs/verification/verdicts/PROV-b1ecdc3ed0cb.json
wrote gates/ready-b1ecdc3ed0cb.yaml
signed as demo-reviewer (key 1f9834b2f85b4562); commit .proofbound/keys/demo-reviewer.pub with the review if it is not committed yet
```

## Step 17 — Warranted, and verifiable by anyone

**What you're doing.**
- Running the pull-request check again.
- Building the release bundle.
- Opening the bundle as a customer would, against their own fresh clone of the branch.

```bash
proofbound sync all > /dev/null
proofbound check demo-base
proofbound warranty bundle demo-base -o "$W/release-2"
git clone -q -b agent/csv-export-2 "$W/shop" "$W/buyer-clone"
```

```text
Proofbound: WARRANTED (1 of 1 claiming commits verified)
  b1ecdc3ed0cb sdd:001-csv-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
Drift:
  drift change=CI-001-csv-export-184wxu state=CURRENT verified_commit=b1ecdc3ed0cb verdict=OV-b1ecdc3ed0cb-1 current_revision=994aee65
  summary warranted=1 current=1 stale=0 other=0
warranty /home/you/projects/proofbound-agent-demo/release-2/warranty-b1ecdc3ed0cb.json commit=b1ecdc3ed0cb20bb8e7dcae9e5ef6be205449e5e claims=sdd:001-csv-export VALID reviewers=demo-reviewer:1f9834b2f85b4562
bundle /home/you/projects/proofbound-agent-demo/release-2 (456aada51336..b413e4295911): 1 of 1 claims verified (1 VALID, 0 UNREVIEWED); VERIFY.md, index.json, badge.svg
```

```bash
proofbound ui --bundle "$W/release-2" --repo "$W/buyer-clone"
```

On the page:
- **VALID · 1**;
- each promise with its verdict: "System MUST export a report as CSV. SATISFIED (demo-reviewer)", and
  the same for the header row;
- the requirement review by demo-reviewer.

Press Ctrl-C when done.

**Why it matters.** The customer gets a signed statement, per promise, of what the agent's change
does and who checked it, and checks it themselves against their own clone.

---

## What you just saw

| The agent… | What Proofbound did | Who decided |
|---|---|---|
| changed a file outside the spec's area | with Claude Code, refused the edit as it was made; when it got through anyway, named `src/pricing.sh line 3` from git, and refused to seal until it was decided | the reviewer, signed |
| shipped a test that cannot fail | undid each part of the change and showed the test noticed none of them; refused to seal until each was decided; blocked the warranty | the reviewer, signed |
| changed a file through the shell, around the hooks | named the line as having no recorded agent edit, next to the lines each Claude Code session wrote | the reviewer, signed (a disclosure) |
| said "tests pass" | did not count it: NOT WARRANTED until reviewed | — |
| redid the work | WARRANTED, and a bundle any customer can verify against a clone | the reviewer, signed |

**What Proofbound does not do.**
- **It undoes parts of a change, not every possible bug.** A test can notice a removed part and still
  miss a subtle bug inside it. This is not full mutation testing.
- **It runs your tests only where the reviewer runs `review draft` and `review seal`,** and only when
  `.proofbound/tests.json` declares them. It never runs them in CI; the check reads the signed record.
- **Git authorship is self-declared** unless commits are signed (C9).
- **Session provenance is what the logs on the agent's machine show.** The agent could write to its
  own log, and an identical line typed later still matches (the redo's `#!/bin/sh` matches the first
  session too). It turns "the agent says it wrote this" into "these lines match these recorded edits,
  and these match none".

**Where the first attempt went.** The branch `agent/csv-export` still holds the commit and its signed
NEEDS_WORK review: the record of what the agent tried, and why it was rejected.

## Clean up

```bash
cd "$HOME" && rm -rf "$HOME/projects/proofbound-agent-demo"
```

The same story, run by a script: `bash docs/releases/r3/agent-demo.sh`. The full feature tour:
[DEMO.md](DEMO.md), and by hand: [DEMO-MANUAL.md](DEMO-MANUAL.md).
