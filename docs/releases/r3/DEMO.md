# Proofbound tour: everything in v0.3.2, and why it matters

**Proofbound proves that someone other than the author checked AI-written code against its spec,
promise by promise, and hands you a signed file that anyone can verify from a clone of your
repository.** It works with OpenSpec and GitHub Spec Kit, beside any coding agent, and runs entirely on
your machine and in your CI.

This page walks through every feature on one small project, using the real output of
[`tour.sh`](tour.sh). To type every step yourself instead, follow [DEMO-MANUAL.md](DEMO-MANUAL.md).
New here? [Start here](README.md) has install and the quickstart. **Using coding agents?** The
[agent demo](DEMO-AGENT.md) is the shorter path: one agent's change that goes off-script, and
everything Proofbound shows about it (the line outside the spec's area, the test that notices nothing,
which agent session wrote each line), through to a warranty on the redo.

## Why it matters

Coding agents now write code faster than people can read it. Teams answer with specs (OpenSpec, Spec
Kit) and reviews, but three questions still have no good answer:

1. **Did anyone other than the author check this change against what the spec promised?** "Tests
   passed" and "approved" say nothing about which promise was checked, or by whom.
2. **Can I show that to someone who doesn't trust me?** A customer, an auditor or a buyer cannot read
   your CI logs or your review tool's database.
3. **Is it still true?** Specs change after code ships; nobody notices which promise moved.

Proofbound answers all three with a record bound to git:

| Who | What they get |
|---|---|
| Engineering lead | A pull-request check that blocks spec-claiming changes until an independent reviewer signs off on every promise |
| Reviewer (a person or a review agent) | A draft listing each promise to judge, and a signature that proves it was them |
| Spec owner | Cannot approve their own change, so the record is credible to everyone else |
| Security and compliance | A reproducible record of who verified what, against which exact spec revision |
| Customer or auditor | One signed file per change that they verify themselves from a clone: no account, no database, no trust in your servers |

## Run it (about 2 minutes)

```bash
bash docs/releases/r3/tour.sh                                        # builds the CLI with Go
PROOFBOUND_BIN=$(command -v proofbound) bash docs/releases/r3/tour.sh   # or use an installed binary
```

It creates a throwaway Spec Kit project, `shop`, and runs thirteen chapters. The demo plays both the
spec owner and the reviewer, and says so; in real use the reviewer is a different person or a review
agent with its own key. Output below is trimmed where marked (`…`).

---

## 1. Set up a project: `proofbound init speckit`

```
$ proofbound init speckit
wrote .proofbound/sdd-profile.json
appended .gitignore
appended AGENTS.md
wrote .claude/agents/proofbound-reviewer.md
wrote .git/hooks/commit-msg
wrote .git/hooks/prepare-commit-msg
wrote .github/workflows/proofbound.yml
```

One command adds a profile naming the spec owner (from `git config user.name`), two commit-message hooks
(one adds the spec claim for a coding agent, one refuses code that claims no spec, chapter 13),
instructions for coding agents (`AGENTS.md`), a ready-made reviewer agent, and the GitHub workflow that
runs the pull-request check.

**Why it matters:** adoption is one command in an existing repository. Nothing to host, no account.

## 2. A spec with promises, and a commit that claims it

```
- **FR-001**: System MUST export a report as CSV.
- **FR-002**: System MUST start every export with a header row.

$ git commit -m "implement CSV export" -m "Intent: sdd:001-csv-export"
```

Each requirement in the spec is a promise (an *obligation*). The commit says which spec it implements
with one trailer line, `Intent: sdd:<feature>`. Agents learn the convention from `AGENTS.md`.

**Why it matters:** the claim is explicit and machine-checkable. A commit that says nothing is never
mistaken for verified work, and a mistyped claim fails the check instead of slipping through.

## 3. The record: bound to the exact spec revision, every promise UNVERIFIED

```
$ proofbound sync all
sdd listed=2 appended=2 existing=0
intent appended=2 git appended=3 checks appended=0 sessions appended=0 reviews appended=0

$ proofbound report intent sdd:001-csv-export
intent=CI-001-csv-export-184wxu … state=IMPLEMENTED_UNVERIFIED spec=unreviewed
  target=…#O-FR-001 … review=UNREVIEWED outcome=UNVERIFIED statement="System MUST export a report as CSV."
  target=…#O-FR-002 … review=UNREVIEWED outcome=UNVERIFIED statement="System MUST start every export with a header row."
  commit=0c5c80d374d2…
```

`sync all` reads git history, every spec revision and every review into an append-only event ledger.
The claim binds to the spec **as it was at that commit**, identified by its content hash.

**Why it matters:** "implemented" is not "verified". Proofbound tracks the difference per promise, and
nobody can later swap in a friendlier version of the spec.

## 4. The pull-request check, before review

```
$ proofbound check <base>
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)
  0c5c80d374d2 sdd:001-csv-export NO REVIEW  no readiness gate gates/ready-0c5c80d374d2.yaml; a reviewer runs `proofbound review draft 0c5c80d374d2`
(exit 1)
```

This is exactly what the GitHub Action runs on every pull request: install, `sync all`, `check` against
the PR's base. It fails the PR and names the next command.

**Why it matters:** the rule is enforced where work merges, not remembered by people.

## 5. Identities: each person declares a signing key

```
$ proofbound keys init --name spec-owner
key 5000cd017a655baa
private …/config/proofbound/signing-key (never commit it)
public .proofbound/keys/spec-owner.pub (commit it: it declares the key for this repository)

$ PROOFBOUND_KEY=…/reviewer.key proofbound keys init --name demo-reviewer
key 0877ffc28ca972b9
```

Private keys stay on each person's machine. Committing the public half is how the repository declares
"this key speaks for demo-reviewer".

**Why it matters:** a signature, not a name typed into a file, identifies the reviewer. Adding a
reviewer is a visible, reviewable commit.

## 6. No one grades their own work

```
$ proofbound review draft 0c5c80d374d2…
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
provenance src/export.sh: 0 by a session, 4 with no recorded agent edit (lines 1, 2, 3, 4), 0 common
provenance warning: no session logs here, so nothing can be traced (they are local to the machine the agent ran on)

$ proofbound review seal …/draft.json          # the spec owner names themselves as reviewer
proofbound: seal: the reviewer "spec-owner" is the declared owner of BR-001-csv-export-9s0gqo; a different reviewer must review it
(exit 1)
```

In this tour a person wrote the code in a shell, so no agent session recorded it, and the draft says
exactly that. With Claude Code, the same lines read "by Claude Code session …"
([agent demo, step 8](DEMO-AGENT.md#step-8--the-review-draft-a-line-outside-the-area-and-a-test-that-notices-nothing)).

**Why it matters:** independence is enforced by the tool, not by policy. This is what makes the record
credible to someone outside the team.

## 7. Independent review, signed: the gate turns PASS

```
$ PROOFBOUND_KEY=…/reviewer.key proofbound review seal …/draft.json
wrote docs/verification/verdicts/RR-0c5c80d374d2-1.md
wrote docs/verification/verdicts/RR-0c5c80d374d2-1.md.sig
wrote docs/verification/verdicts/OV-0c5c80d374d2-1.md
wrote docs/verification/verdicts/OV-0c5c80d374d2-1.md.sig
wrote docs/verification/verdicts/PROV-0c5c80d374d2.json
wrote gates/ready-0c5c80d374d2.yaml
signed as demo-reviewer (key 0877ffc28ca972b9)

$ proofbound gates canary
gate=ready-0c5c80d374d2 state=PASS … reason="2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE"
```

The reviewer judges two things: is each requirement **verifiable** as written (`RR-…`), and does the
commit **satisfy** each promise (`OV-…`). The review files, their signatures and a readiness gate are
committed with the code.

**Why it matters:** the evidence lives in your repository, in plain files, signed. Reviews are no
longer trapped in a SaaS tool or a chat thread.

## 8. The pull-request check, after review

```
$ proofbound check <base>
Proofbound: WARRANTED (1 of 1 claiming commits verified)
  0c5c80d374d2 sdd:001-csv-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
```

With `--markdown` the same result becomes the table the Action posts as the job summary:

| Commit | Claims | State | Why |
|---|---|---|---|
| `0c5c80d374d2` | `sdd:001-csv-export` | **PASS** | 2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE |

The check runs on a fresh CI machine with an empty database and gets the same answer (fixed in v0.3.1).

**Why it matters:** "warranted" becomes a merge requirement that anyone on the team can read.

## 9. A portable signed warranty

```
$ proofbound warranty export 0c5c80d374d2… -o warranty.json
warranty warranty.json commit=0c5c80d374d2… change=sdd:001-csv-export key=5000cd017a655baa obligations=2

$ proofbound warranty verify warranty.json                      # no repository: never a pass
UNVERIFIED key=5000cd017a655baa: signature only; pass --repo to check the claims          (exit 2)

$ proofbound warranty verify warranty.json --repo buyer-clone --trust-key 5000cd017a655baa
VALID key=5000cd017a655baa commit=0c5c80d374d2… change=sdd:001-csv-export obligations=2 reviewers=demo-reviewer:0877ffc28ca972b9

$ proofbound warranty verify tampered.json --repo buyer-clone       # one promise edited in the file
INVALID: signature                                                                         (exit 1)

$ proofbound warranty verify warranty.json --repo buyer-clone --trust-key 0000000000000000
INVALID: fingerprint mismatch: key 5000cd017a655baa, expected 0000000000000000             (exit 1)
```

The warranty is one JSON file: the commit, every promise word for word, the review outcomes, and
signatures. `verify` needs only the file and a clone; it re-reads the spec from git itself.

**Why it matters:** this is the proof you hand to someone who does not trust you. A customer or an
auditor checks it in seconds, with no access to your systems.

## 10. Release proof: one bundle per release

```
$ proofbound warranty bundle <base> -o release-warranties
warranty release-warranties/warranty-0c5c80d374d2.json … VALID reviewers=demo-reviewer:0877ffc28ca972b9
not warranted 45ce3fba83c4 sdd:002-totals-row NO REVIEW  …
bundle release-warranties (bdc4eac50884..45ce3fba83c4): 1 of 2 claims verified (1 VALID, 0 UNREVIEWED); VERIFY.md, index.json, badge.svg
(exit 3)

$ ls release-warranties
VERIFY.md  badge.svg  index.json  warranty-0c5c80d374d2.json
```

`VERIFY.md` is a page anyone can read: what is covered, what is not ("4 commit(s) in this range claim
none and are not covered"), a table per commit, and the exact steps to verify each file. It says up
front that it and `index.json` are unsigned conveniences and only the warranty files are proof. The
badge reads "1 of 2 claims verified" for your README. Exit 3 means "written, but not everything is
VALID".

**Why it matters:** every release can ship with its proof attached, honest about its gaps.

## 11. Drift: the spec changed after it was verified

```
$ proofbound report drift
drift change=CI-001-csv-export-184wxu state=STALE verified_commit=0c5c80d374d2 … verified_revision=46d1df8f current_revision=00aeaa1e
  changed O-FR-002 "System MUST start every export with a header row." -> "System MUST start every export with a header row naming each column in lower case."
summary warranted=1 current=0 stale=1 other=0
```

The spec owner tightened one promise. The old commit is still warranted for what the spec said then,
but Proofbound flags exactly which promise moved, so someone re-verifies.

**Why it matters:** you know when a promise you sold is no longer the promise you verified.

## 12. The record is reproducible

```
$ proofbound rebuild
$ proofbound check <base>
Proofbound: NOT WARRANTED (1 of 2 claiming commits verified)
  0c5c80d374d2 sdd:001-csv-export PASS  …
  45ce3fba83c4 sdd:002-totals-row NO REVIEW  …
```

`rebuild` drops every derived view and replays the append-only ledger; the answers are the same. On a
fresh clone, `sync all` rebuilds the same answers from git alone, which is what CI does.

**Why it matters:** the record is derived from git, not a database someone could edit, so it holds up
to scrutiny.

## 13. Agents: the claim is added for them, and unclaimed code fails the check

```
$ git commit -m "add a greeting"      # code, no claim, no task
proofbound: this commit changes code (src/hello.sh) but claims no spec.
  Run `proofbound task start <change>` and commit again: the claim is added for you.
  Specs, docs, gates, review records and other Markdown need no claim; a coding agent's instructions (CLAUDE.md, AGENTS.md, .claude/) do.
$ proofbound task start 002-totals-row
task sdd:002-totals-row: commits that change code now end with `Intent: sdd:002-totals-row`; `proofbound task stop` when done
$ git commit -m "add a greeting"      # the same commit, with a task active
$ git log -1 --format=%B
add a greeting

Intent: sdd:002-totals-row
$ proofbound task stop
no task
$ git commit --no-verify -m "sneak in a change"      # skips the hooks: no claim
$ proofbound check <before these commits>
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified; 1 commit changes code without claiming a spec)
  72c2a430b618 sdd:002-totals-row NO REVIEW  …
  4a79ce286a4a - UNCOVERED  claims no spec but changes src/hello.sh; run `proofbound task start <change>` …
```

A coding agent runs `proofbound task start <change>` once; from then on every commit that changes
code carries the claim without anyone typing it. A code commit with no claim is refused at commit
time, and if someone skips the hooks, the pull-request check lists it as `UNCOVERED` and fails.
Commits that only touch specs, docs, `gates/`, review records or Markdown never need a claim.

**Why it matters:** the check cannot be passed by leaving the label off. Every line of code in a
warranted pull request belongs to a spec someone reviewed, and agents do not have to remember a rule
to get there.

---

## Use it on your project

Install and set up with the [quickstart](../mvp/QUICKSTART.md). Before an agent writes code for a
change: `proofbound task start <change>` (chapter 13). The signing steps:

```bash
proofbound keys init                       # private key in ~/.config/proofbound (never in the repo)
git add .gitignore .proofbound/keys && git commit -m "declare warranty key"   # keys need no claim
proofbound warranty export <commit>        # only for a commit whose ready-<commit12> gate is PASS
```

**The reviewer signs with their own key.** Either way makes the warranty VALID instead of UNREVIEWED:

```bash
PROOFBOUND_KEY=~/.config/proofbound/reviewer-key proofbound keys init --name <reviewer>   # once; commit the .pub
# A: declare the key before sealing; `review seal` then writes a .sig beside each review file (chapter 7)
# B: after export, cosign the warranty file
PROOFBOUND_KEY=~/.config/proofbound/reviewer-key proofbound warranty cosign warranty-<commit12>.json
```

Option B, from UNREVIEWED to VALID, is shown end to end by
`bash docs/releases/r3/demo-warranty.sh speckit` (also: a key the repository never declared is refused
at export).

Whoever receives `warranty-<commit12>.json`:

```bash
git clone <the project's public URL> project      # from a source you trust, not from the sender
proofbound warranty verify warranty-<commit12>.json --repo project [--trust-key <fingerprint>]
```

## What VALID means

`VALID` (exit 0) means all of the following hold:

1. The file was signed by a key the repository commits under `.proofbound/keys/`.
2. Every promise in it is word for word what the spec said at that commit. `verify` re-reads the spec
   from git itself.
3. Every obligation was reviewed SATISFIED, and every requirement VERIFIABLE, by someone other than the
   spec's owner.
4. The review files are committed, byte for byte.
5. Each reviewer's own declared key signed: a `.sig` beside each of their review files, or a
   cosignature in the file. That key is not the exporter's, an owner's, or another reviewer's.

`UNVERIFIED` (exit 2) means no repository was given. `UNREVIEWED` (exit 3) means everything else holds
but a named reviewer has not signed. `INVALID` (exit 1) names the first check that failed.

## Caveats

- **The repository is the trust anchor.** Anyone who can commit to it can declare a key, so clone from
  a source you trust, and pin `--trust-key` when you know the fingerprint. There is no rotation,
  revocation or transparency log yet.
- **A reviewer signature proves "the key declared as alice signed", not "the human alice signed".**
  Declaring a key is a commit, so forging a reviewer takes a visible key declaration instead of a text
  edit. Reviewer names are compared by key name (`Alice` and `alice` are one person).
- **Proofbound records what reviewers decided.** It runs your tests only when `.proofbound/tests.json`
  declares them, and then only in `review draft` and `seal` on the reviewer's machine, never in CI (see
  [the agent demo, step 8](DEMO-AGENT.md#step-8--the-review-draft-a-line-outside-the-area-and-a-test-that-notices-nothing)).
- **Only commits that claim a spec are covered.** The bundle says how many are not.
- **A warranty holds only while its evidence is in the repository.** Deleting or editing a cited review
  file makes it INVALID.
- **A warranty is about its commit.** Later spec changes do not invalidate it; `proofbound report
  drift` says whether the spec has moved since.
- **Verdicts sealed with v0.3.0 must be sealed again** to pass in CI (see the
  [release notes](RELEASE-NOTES.md)).
- **Plan and decision:** [`R3-warranty-export-spec.md`](../../plans/R3-warranty-export-spec.md),
  [`VD-r3-warranty-first-s59ktv`](../../decisions/VD-r3-warranty-first-s59ktv.md).
