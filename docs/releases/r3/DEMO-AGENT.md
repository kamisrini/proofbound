# The agent goes off-script: a hands-on demo

A coding agent is asked to build one small feature from a spec. It finishes in seconds, says the tests
pass, and they do. It also changed a price it was never asked to touch, left out one of the two things
the spec promised, and wrote a test that cannot fail. At agent speed, nobody reads every line, so this
is the change that gets merged.

This page shows what Proofbound does with it:
- it names the exact line outside the area the spec allowed;
- an independent reviewer catches the test that proves nothing, and signs the finding;
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

You need Proofbound v0.3.3 or later on your PATH, plus `git` and `jq`. Port 55432 must be free
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
- The tag `demo-base` marks where the agent's work will start: every check below asks "is everything
  since `demo-base` warranted?"

```bash
proofbound init speckit > /dev/null
git add -A .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github
git commit -qm "add proofbound"
proofbound keys init --name spec-owner > /dev/null
PROOFBOUND_KEY="$W/reviewer.key" proofbound keys init --name demo-reviewer > /dev/null
git add .gitignore .proofbound/keys
git commit -qm "declare the spec owner's and the reviewer's keys"
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

## Step 5 — The agent's edits (you paste them)

**What you're doing.** Playing the agent. It writes the export, writes a test, and, "while it was
there", raises the member discount from 10% to 25%. Look closely at all three:
- the export has no header row;
- the test only checks that the script runs;
- `src/pricing.sh` is not in the spec's area.

```bash
mkdir -p tests
printf '#!/bin/sh\nprintf "apples,3\\npears,5\\n"\n' > src/export.sh
printf '#!/bin/sh\nsh src/export.sh > /dev/null && echo "PASS: export works"\n' > tests/export_test.sh
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
  a14083681dee sdd:001-csv-export NO REVIEW  no readiness gate gates/ready-a14083681dee.yaml; a reviewer runs `proofbound review draft a14083681dee`
Drift:
  summary warranted=0 current=0 stale=0 other=0
proofbound: check: not warranted
```

**Why it matters.** "The agent says the tests pass" is not evidence. Nothing is warranted until someone
other than the agent and the spec owner has checked each promise.

---

## Act 2 — An independent reviewer

## Step 8 — The review draft names the line outside the area

**What you're doing.** The reviewer asks Proofbound for a review draft of the agent's commit. The draft
lists each promise to decide and, from git, every changed line outside the spec's area.

```bash
proofbound review draft attempt-1
```

```text
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
outside area src/pricing.sh line 3
decide each in out_of_area: "accept" with a reason, or "reject"
draft /home/you/projects/proofbound-agent-demo/shop/.proofbound/drafts/a14083681deea06c39ff32ad303ba8f2b3aa82ff/draft.json
```

**What just happened.** Proofbound found the discount change by itself: `src/pricing.sh line 3`. The
review cannot be sealed until the reviewer accepts that line, with a reason, or rejects it, so a change
outside the area cannot slip through unnoticed.

**Why it matters.** In a large agent diff, a one-line price change hides easily. Here it is named,
with its file and line, and it needs a signed decision.

## Step 9 — Does the agent's test prove anything?

**What you're doing.** The reviewer runs the test, then breaks the code on purpose (an export that
prints nothing) and runs it again. A test that still passes on broken code proves nothing. Then the
reviewer looks at the real output.

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

**What just happened.** The test passed on code that does nothing, so it cannot fail. The real output
has no header row, so FR-002 is not met.

**Be clear about who found this.** Proofbound does not detect a test that cannot fail; the reviewer did,
by breaking the code. What Proofbound guarantees is that the reviewer is someone other than the agent
and the spec owner, that their finding is signed, and that nothing is warranted until it is fixed.

## Step 10 — The reviewer records the findings and signs them

**What you're doing.** Filling in the draft:
- both requirements are clear enough to check (VERIFIABLE);
- FR-001 is SATISFIED and FR-002 is NOT_SATISFIED;
- the pricing line is rejected.

Then the reviewer seals it with their own key and commits the signed records.

```bash
draft=.proofbound/drafts/$(git rev-parse attempt-1)/draft.json
jq '.declared_reviewer = "demo-reviewer"
  | .reviews[].obligations[].outcome = "VERIFIABLE"
  | .verdicts[].obligations |= map(.outcome = (if .obligation_id == "O-FR-002" then "NOT_SATISFIED" else "SATISFIED" end))
  | .out_of_area |= map(.decision = "reject")' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"
PROOFBOUND_KEY="$W/reviewer.key" proofbound review seal "$draft"
git add docs/verification/verdicts gates
git commit -qm "review attempt-1: needs work"
sed -n 2p docs/verification/verdicts/OV-*.md | jq -r .status
```

```text
wrote docs/verification/verdicts/RR-a14083681dee-1.md
wrote docs/verification/verdicts/RR-a14083681dee-1.md.sig
wrote docs/verification/verdicts/OV-a14083681dee-1.md
wrote docs/verification/verdicts/OV-a14083681dee-1.md.sig
wrote docs/verification/verdicts/AREA-a14083681dee.json
wrote gates/ready-a14083681dee.yaml
signed as demo-reviewer (key 98d3a6c05b6d2203); commit .proofbound/keys/demo-reviewer.pub with the review if it is not committed yet
NEEDS_WORK
```

**What just happened.** The verdict is NEEDS_WORK, and `AREA-….json` records, signed, that the pricing
line was rejected and by whom.

## Step 11 — The check after review

```bash
proofbound sync all > /dev/null
proofbound check demo-base
```

```text
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)
  a14083681dee sdd:001-csv-export BLOCKED  1 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
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
- "5 of 7 controls held for this range. 2 exception(s), 0 with a signed outcome."
- **Unreviewed change · control C1**: the commit claims the spec but is BLOCKED.
- **Change outside its declared area · control C10**: "implement CSV export" changed lines outside its
  spec's area (not accepted): `src/pricing.sh line 3`.
- **Who did the work**: Claude, as an Agent, with one commit and two open exceptions.
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
not warranted a14083681dee sdd:001-csv-export BLOCKED  1 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
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
- the draft now reports that every changed line is inside the area;
- the break-it check now **fails** the test, so the test proves something;
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
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
area: every changed line is inside the declared area
draft /home/you/projects/proofbound-agent-demo/shop/.proofbound/drafts/b24e6d02a475cddb18434b03a2ade1c546c7d64b/draft.json
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
wrote docs/verification/verdicts/RR-b24e6d02a475-1.md
wrote docs/verification/verdicts/RR-b24e6d02a475-1.md.sig
wrote docs/verification/verdicts/OV-b24e6d02a475-1.md
wrote docs/verification/verdicts/OV-b24e6d02a475-1.md.sig
wrote gates/ready-b24e6d02a475.yaml
signed as demo-reviewer (key 98d3a6c05b6d2203); commit .proofbound/keys/demo-reviewer.pub with the review if it is not committed yet
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
  b24e6d02a475 sdd:001-csv-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
Drift:
  drift change=CI-001-csv-export-184wxu state=CURRENT verified_commit=b24e6d02a475 verdict=OV-b24e6d02a475-1 current_revision=994aee65
  summary warranted=1 current=1 stale=0 other=0
warranty /home/you/projects/proofbound-agent-demo/release-2/warranty-b24e6d02a475.json commit=b24e6d02a475cddb18434b03a2ade1c546c7d64b claims=sdd:001-csv-export VALID reviewers=demo-reviewer:98d3a6c05b6d2203
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
| changed a file outside the spec's area | named `src/pricing.sh line 3` from git, and refused to seal until it was decided | the reviewer, signed |
| shipped a test that cannot fail | made sure the reviewer was not the agent or the spec owner, signed their NOT_SATISFIED, and blocked the warranty | the reviewer (by breaking the code) |
| said "tests pass" | did not count it: NOT WARRANTED until reviewed | — |
| redid the work | WARRANTED, and a bundle any customer can verify against a clone | the reviewer, signed |

**What Proofbound does not do (yet).** It does not detect a test that cannot fail, and it does not run
your tests. Checking tests automatically by breaking the code is planned; today the reviewer does it.
Git authorship is self-declared unless commits are signed (C9).

**Where the first attempt went.** The branch `agent/csv-export` still holds the commit and its signed
NEEDS_WORK review: the record of what the agent tried, and why it was rejected.

## Clean up

```bash
cd "$HOME" && rm -rf "$HOME/projects/proofbound-agent-demo"
```

The same story, run by a script: `bash docs/releases/r3/agent-demo.sh`. The full feature tour:
[DEMO.md](DEMO.md), and by hand: [DEMO-MANUAL.md](DEMO-MANUAL.md).
