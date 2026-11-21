# Proofbound by hand: every step, typed yourself, and why it matters

This is the [tour](DEMO.md), but **you type every command**. Each step tells you what you are about to
do, gives you a block to paste, shows what you should see, explains what just happened, and says why it
matters to a business. It takes about 20–30 minutes.

**How to read the blocks.** A block marked `bash` is safe to paste into your terminal as it is. A block
marked `text` is output, for comparison only; never paste it. Your commit hashes and key fingerprints
will differ from the ones shown; everything else should match.

**Two hats.** You play two people: the **spec owner** (who writes the spec and the code) and an
independent **reviewer**. In real use the reviewer is a different person, or a review agent with its
own key. The demo keeps their private keys in separate files so you can see that they are separate.

## Before you start

You need Proofbound v0.3.1 on your PATH, plus `git` and `jq`. Port 55432 must be free (Proofbound
starts a private database there for each command and stops it afterwards).

```bash
proofbound version    # proofbound v0.3.1
```

If that says `command not found`, install it with the one-line command in the [README](README.md), then
run `export PATH="$HOME/.local/bin:$PATH"; hash -r`.

---

## Step 0a — Create an empty project (once)

**What you're doing.** Making a brand-new git repository called `shop` in a demo folder of its own. The
`git config` lines set *this repository's* author to `spec-owner`; your global git settings are not
touched.

```bash
mkdir -p "$HOME/proofbound-manual-demo/shop"
cd "$HOME/proofbound-manual-demo/shop"
git init -q -b main
git config user.name spec-owner
git config user.email spec-owner@example.invalid
```

Nothing is printed.

## Step 0b — Prepare this terminal (repeat in every new terminal)

**What you're doing.** Setting a few shell variables the later steps use. `XDG_CONFIG_HOME` makes the
spec owner's private key go into the demo folder instead of your real `~/.config`, so the demo cannot
touch any key you already have. `base` and `csv` are two commit hashes the later steps refer to; on a
first run they are empty, and they fill in as you go.

**If you open a new terminal halfway through, paste this block again**, and it picks up where you left
off.

```bash
export W="$HOME/proofbound-manual-demo"
export XDG_CONFIG_HOME="$W/config"
unset DATABASE_URL
cd "$W/shop"
base=$(git rev-list --max-parents=0 HEAD 2>/dev/null)
csv=$(git log --format=%H -1 --grep='^Intent: sdd:001-csv-export$' 2>/dev/null)
```

**Use this terminal for the demo only**, because `XDG_CONFIG_HOME` now points at the demo folder.

---

## Step 1 — Set the project up: `proofbound init speckit`

**What you're doing.** Adding Proofbound to the project in one command, for a project that writes its
specs with [GitHub Spec Kit](https://github.com/github/spec-kit) (`init openspec` exists too).

```bash
proofbound init speckit
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github
PROOFBOUND_NO_INTENT=1 git commit -q -m "add proofbound"
base=$(git rev-parse HEAD)
git ls-files
```

**What you should see:**

```text
wrote …/shop/.proofbound/sdd-profile.json
appended …/shop/.gitignore
appended …/shop/AGENTS.md
wrote …/shop/.claude/agents/proofbound-reviewer.md
wrote …/shop/.git/hooks/commit-msg
wrote …/shop/.github/workflows/proofbound.yml
.claude/agents/proofbound-reviewer.md
.github/workflows/proofbound.yml
.gitignore
.proofbound/sdd-profile.json
AGENTS.md
```

**What just happened.** `init` wrote six things:

| File | What it is for |
|---|---|
| `.proofbound/sdd-profile.json` | Names the spec owner (your `git config user.name`: `spec-owner`) and where specs live |
| `.gitignore` | Keeps Proofbound's local database and drafts out of git, but lets the profile and public keys in |
| `AGENTS.md` | Tells any coding agent (Claude, Copilot, Cursor…) the commit rule in step 3 |
| `.claude/agents/proofbound-reviewer.md` | A ready-made reviewer agent for Claude Code |
| `.git/hooks/commit-msg` | Refuses a commit that does not say which spec it implements (next step shows it) |
| `.github/workflows/proofbound.yml` | The pull-request check, for when this project is on GitHub |

`PROOFBOUND_NO_INTENT=1` says "this commit implements no spec". `base` remembers this commit: later,
the pull-request check compares everything after it, as if it were the base of a pull request.

**Business value.** Adoption is one command in an existing repository: no server to host, no account,
nothing to buy before you see it work.

## Step 1b — See the commit rule fire

**What you're doing.** Trying to commit without saying which spec the commit implements.

```bash
git commit --allow-empty -q -m "try a commit with no claim"
```

**What you should see** (and no commit is made):

```text
proofbound: end the message with 'Intent: sdd:<change-name>', or commit with PROOFBOUND_NO_INTENT=1
```

**What just happened.** The hook `init` installed blocks commits that make no statement. Every commit
either claims a spec (`Intent: sdd:<name>`) or says explicitly that it implements none.

**Business value.** Nobody, human or agent, can quietly skip the question "which promise does this code
keep?". Silence is never mistaken for a claim.

---

## Step 2 — Write a spec with two promises

**What you're doing.** Writing a small Spec Kit feature, `001-csv-export`. Each `FR-…` line is a
requirement; Proofbound treats each one as a **promise** (it calls them *obligations*). Spec Kit counts a
feature once its `tasks.md` exists, so we write one.

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
git add specs
PROOFBOUND_NO_INTENT=1 git commit -q -m "spec: 001-csv-export"
cat specs/001-csv-export/spec.md
```

**What just happened.** The spec is committed on its own. Writing a spec is not implementing it, so this
commit also claims nothing.

**Business value.** The spec is the contract. Everything Proofbound does later is measured against
these two sentences, word for word, so there is no arguing afterwards about what was promised.

## Step 3 — Write the code, and a commit that claims the spec

**What you're doing.** Writing a tiny "export" script and committing it with one extra line in the
message: `Intent: sdd:001-csv-export`. That line is the **claim**: "this commit implements that spec".

```bash
mkdir -p src
cat > src/export.sh <<'EOF'
#!/bin/sh
# Prints the report as CSV: a header row, then one row per item.
printf 'item,count\n'
printf 'apples,3\npears,5\n'
EOF
git add src
git commit -q -m "implement CSV export" -m "Intent: sdd:001-csv-export"
csv=$(git rev-parse HEAD)
git log --oneline
```

**What you should see:**

```text
e620a48 implement CSV export
4483fc6 spec: 001-csv-export
8050d06 add proofbound
```

**What just happened.** The hook accepted the commit because it makes a claim. `csv` remembers this
commit's hash; steps 7–10 use it.

**Business value.** The claim is explicit and machine-checkable. A coding agent learns to write it from
`AGENTS.md`, so AI-written changes declare what they claim to do.

---

## Step 4 — Record everything, and see each promise UNVERIFIED

**What you're doing.** `sync all` reads git history, every spec revision and every review into
Proofbound's record. `report intent` then asks: what is the state of this spec's promises?

```bash
proofbound sync all
proofbound report intent sdd:001-csv-export
```

**What you should see** (the first run may take longer while the embedded database is downloaded, about
15 MB, once per machine):

```text
sdd listed=2 appended=2 existing=0
intent appended=2 git appended=3 checks appended=0 sessions appended=0 reviews appended=0
intent=CI-001-csv-export-184wxu … state=IMPLEMENTED_UNVERIFIED spec=unreviewed …
  target=…BR-001-csv-export-9s0gqo@c0245f15…#O-FR-001 … review=UNREVIEWED outcome=UNVERIFIED statement="System MUST export a report as CSV." …
  target=…BR-001-csv-export-9s0gqo@c0245f15…#O-FR-002 … review=UNREVIEWED outcome=UNVERIFIED statement="System MUST start every export with a header row." …
  commit=e620a48ca48b… …
```

**What just happened.** The claim is now bound to the spec **as it was at that commit**, identified by
the content hash after the `@`. Each promise gets its own row, and both say `UNVERIFIED`: the code
claims to implement them, but nobody independent has checked.

**Business value.** "Implemented" is not "verified", and Proofbound tracks the difference promise by
promise. Nobody can later swap in a friendlier version of the spec, because the claim points at the
exact text.

## Step 5 — The pull-request check, before anyone reviews

**What you're doing.** Running the same check the GitHub Action runs on every pull request: is every
commit since `base` that claims a spec verified?

```bash
proofbound check "$base"
```

**What you should see** (exit code 1, which fails a pull request):

```text
Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)
  e620a48ca48b sdd:001-csv-export NO REVIEW  no readiness gate gates/ready-e620a48ca48b.yaml; a reviewer runs `proofbound review draft e620a48ca48b`
Drift:
  summary warranted=0 current=0 stale=0 other=0
proofbound: check: not warranted
```

**What just happened.** The check found one claiming commit, saw that nobody has reviewed it, and named
the exact next command.

**Business value.** The rule is enforced where work merges, not remembered by people. A reviewer is
told exactly what to do next.

---

## Step 6 — Each person declares a signing key

**What you're doing.** Creating two key pairs: one for the spec owner, one for the reviewer. Each
private key stays on that person's machine. The public half is committed to the repository.

```bash
proofbound keys init --name spec-owner
PROOFBOUND_KEY="$W/reviewer.key" proofbound keys init --name demo-reviewer
git add .gitignore .proofbound/keys
PROOFBOUND_NO_INTENT=1 git commit -q -m "declare the spec owner's and the reviewer's keys"
git ls-files .proofbound/keys
proofbound sync all
```

**What you should see:**

```text
key 499259f09a9fb7cd
private …/config/proofbound/signing-key (never commit it)
public …/shop/.proofbound/keys/spec-owner.pub (commit it: it declares the key for this repository)
key 7a7df12f53e632bd
private …/reviewer.key (never commit it)
public …/shop/.proofbound/keys/demo-reviewer.pub (commit it: it declares the key for this repository)
.proofbound/keys/demo-reviewer.pub
.proofbound/keys/spec-owner.pub
```

(`unchanged …/.gitignore` lines are fine; `init` already prepared it.) **Note the first fingerprint**
(`499259f0…` here): it is the spec owner's, and it signs the warranty in step 9.

**What just happened.** `PROOFBOUND_KEY` chose where the reviewer's private key lives, so the two keys
are separate files, as they would be on two people's machines. Committing a `.pub` file is how the
repository declares "this key speaks for demo-reviewer".

**Business value.** A signature, not a name typed into a file, identifies who reviewed. Adding a
reviewer is itself a visible commit that can be reviewed.

## Step 7 — Get the review draft, and look inside it

**What you're doing.** Asking Proofbound for a review **draft** of the claiming commit: a file listing
every promise, each with an empty outcome for the reviewer to fill in.

```bash
proofbound review draft "$csv"
draft=".proofbound/drafts/$csv/draft.json"
jq '{declared_reviewer, review: [.reviews[].obligations[] | {obligation_id, statement, outcome}], verdict: [.verdicts[].obligations[] | {obligation_id, statement, outcome}]}' "$draft"
```

**What you should see:**

```text
obligation CI-001-csv-export-184wxu O-FR-001: System MUST export a report as CSV.
obligation CI-001-csv-export-184wxu O-FR-002: System MUST start every export with a header row.
draft …/shop/.proofbound/drafts/e620a48c…/draft.json
{
  "declared_reviewer": "",
  "review": [
    { "obligation_id": "O-FR-001", "statement": "System MUST export a report as CSV.", "outcome": "" },
    { "obligation_id": "O-FR-002", "statement": "System MUST start every export with a header row.", "outcome": "" }
  ],
  "verdict": [
    { "obligation_id": "O-FR-001", "statement": "System MUST export a report as CSV.", "outcome": "" },
    { "obligation_id": "O-FR-002", "statement": "System MUST start every export with a header row.", "outcome": "" }
  ]
}
```

**What just happened.** The reviewer must answer two different questions for every promise:

- **review** (the requirement review): is the promise *testable as written*? `VERIFIABLE`, or
  `AMBIGUOUS` / `UNTESTABLE` / `CONTRADICTORY`.
- **verdict** (the obligation verdict): does *this commit* keep the promise? `SATISFIED`, or
  `NOT_SATISFIED` / `INCONCLUSIVE`.

`declared_reviewer` is who is answering. The draft is a working file, kept out of git.

**Business value.** Review becomes a checklist of the actual promises instead of a thumbs-up on a diff.
A vague requirement gets caught as vague, not quietly "passed".

## Step 8 — Try to review your own work (and be refused)

**What you're doing.** Filling in the draft as the **spec owner** (everything `VERIFIABLE` and
`SATISFIED`) and trying to seal it. The small `decide` helper edits the draft with `jq`; you could also
open the JSON in an editor and fill in the same fields by hand.

```bash
decide() { jq --arg who "$1" '.declared_reviewer = $who | .reviews[].obligations[].outcome = "VERIFIABLE" | .verdicts[].obligations[].outcome = "SATISFIED"' "$draft" > "$draft.tmp" && mv "$draft.tmp" "$draft"; }
decide spec-owner
proofbound review seal "$draft"
```

**What you should see** (exit code 1):

```text
proofbound: seal: the reviewer "spec-owner" is the declared owner of BR-001-csv-export-9s0gqo; a different reviewer must review it
```

**What just happened.** `seal` checked who is reviewing against the spec's declared owner (from the
profile in step 1) and refused.

**Business value.** Independence is enforced by the tool, not by a policy document. That is what makes
the record believable to someone outside the team.

## Step 9 — The independent reviewer checks the code and signs

**What you're doing.** Now wearing the reviewer's hat: running the code to check both promises, filling
in the draft as `demo-reviewer`, and sealing it with the reviewer's own key. Then committing the review
and checking the gate.

```bash
sh src/export.sh
decide demo-reviewer
PROOFBOUND_KEY="$W/reviewer.key" proofbound review seal "$draft"
git add docs/verification/verdicts gates
PROOFBOUND_NO_INTENT=1 git commit -q -m "review $(git rev-parse --short=12 "$csv")"
proofbound sync all
proofbound gates canary
```

**What you should see:**

```text
item,count
apples,3
pears,5
wrote docs/verification/verdicts/RR-e620a48ca48b-1.md
wrote docs/verification/verdicts/RR-e620a48ca48b-1.md.sig
wrote docs/verification/verdicts/OV-e620a48ca48b-1.md
wrote docs/verification/verdicts/OV-e620a48ca48b-1.md.sig
wrote gates/ready-e620a48ca48b.yaml
signed as demo-reviewer (key 7a7df12f53e632bd); …
…
gate=ready-e620a48ca48b state=PASS … reason="2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE"
```

**What just happened.** The output shows a header row and CSV rows, so both promises hold. `seal` wrote
two review files, `RR-…` (requirement review) and `OV-…` (obligation verdict), each with a `.sig`
signature by the reviewer's key, plus a readiness **gate** for the commit. They are committed with the
code. The gate is `PASS` because every promise is `SATISFIED` and every requirement `VERIFIABLE`. Open
`docs/verification/verdicts/OV-*.md` to see what a review looks like.

**Business value.** The evidence lives in your own repository, as plain signed files. Reviews are no
longer stuck in a SaaS tool or a chat thread, and they survive any vendor.

## Step 10 — The pull-request check, after review

**What you're doing.** Running the same check as step 5, then the Markdown version the GitHub Action
posts on the pull request.

```bash
proofbound check "$base"
proofbound check "$base" --markdown
```

**What you should see** (exit code 0, which passes a pull request):

```text
Proofbound: WARRANTED (1 of 1 claiming commits verified)
  e620a48ca48b sdd:001-csv-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
Drift:
  drift change=CI-001-csv-export-184wxu state=CURRENT …
  summary warranted=1 current=1 stale=0 other=0
### Proofbound: WARRANTED (1 of 1 claiming commits verified)

| Commit | Claims | State | Why |
|---|---|---|---|
| `e620a48ca48b` | `sdd:001-csv-export` | **PASS** | 2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE |
…
```

**What just happened.** Step 5's NOT WARRANTED is now WARRANTED, because of the signed review you
committed and nothing else. `CURRENT` means the spec has not changed since it was verified.

**Business value.** "Warranted" becomes a merge requirement anyone on the team can read, and a fresh CI
machine with an empty database reaches the same answer from git alone.

---

## Step 11 — Export a signed warranty, and verify it like a stranger

**What you're doing.** Exporting a **warranty** for the reviewed commit: one signed JSON file holding the
commit, every promise word for word, the review outcomes and the signatures. Then verifying it the way
a customer or auditor would.

```bash
proofbound warranty export "$csv" -o "$W/warranty.json"
jq '{commit: .body.commit, change: .body.change_name, promises: [.body.requirements[].obligations[].statement]}' "$W/warranty.json"
proofbound warranty verify "$W/warranty.json"
```

**What you should see** (the last command exits 2):

```text
warranty …/warranty.json commit=e620a48c… change=sdd:001-csv-export key=499259f09a9fb7cd obligations=2
{
  "commit": "e620a48c…",
  "change": "001-csv-export",
  "promises": [
    "System MUST export a report as CSV.",
    "System MUST start every export with a header row."
  ]
}
UNVERIFIED key=499259f09a9fb7cd: signature only; pass --repo to check the claims
```

**What just happened.** The file is signed by the spec owner's key (step 6). Verifying it **without the
repository** only checks the signature, so it is never a pass: `UNVERIFIED`.

Now pretend to be the buyer. They clone the project themselves and verify against their clone. Then
they pin the fingerprint they got out of band (on a call, from the project's website), and see what
happens when someone tampers with the file or uses the wrong key:

```bash
git clone -q "$W/shop" "$W/buyer-clone"
fpr=$(proofbound warranty verify "$W/warranty.json" --repo "$W/buyer-clone" | sed -n 's/.*key=\([0-9a-f]*\).*/\1/p')
proofbound warranty verify "$W/warranty.json" --repo "$W/buyer-clone" --trust-key "$fpr"
sed 's/System MUST start every export with a header row./System MAY start an export with a header row./' "$W/warranty.json" > "$W/tampered.json"
proofbound warranty verify "$W/tampered.json" --repo "$W/buyer-clone"
proofbound warranty verify "$W/warranty.json" --repo "$W/buyer-clone" --trust-key 0000000000000000
```

**What you should see:**

```text
VALID key=499259f09a9fb7cd commit=e620a48c… change=sdd:001-csv-export obligations=2 reviewers=demo-reviewer:7a7df12f53e632bd
INVALID: signature
INVALID: fingerprint mismatch: key 499259f09a9fb7cd, expected 0000000000000000
```

**What just happened.** `VALID` means: the key is declared in the repository; every promise in the file
is word for word what the spec said at that commit (`verify` re-reads it from git); every promise was
reviewed by someone other than the spec owner; the review files are committed byte for byte; and the
reviewer's own key signed them. Changing one word (`MUST` to `MAY`) breaks the signature. Pinning the
wrong fingerprint is refused. ([What VALID means](DEMO.md#what-valid-means), in full.)

**Business value.** This is the proof you hand to someone who does not trust you. A customer or auditor
checks it in seconds with a clone and one command: no account, no access to your systems, no trust in
your servers.

## Step 12 — Attach the proof to a release

**What you're doing.** Adding a second feature, `002-totals-row`, and implementing it, but **not**
reviewing it. Then building the proof bundle for a release covering everything since `base`.

```bash
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
git add specs
PROOFBOUND_NO_INTENT=1 git commit -q -m "spec: 002-totals-row"
cat >> src/export.sh <<'EOF'
printf 'total,8\n'
EOF
git add src
git commit -q -m "add a totals row" -m "Intent: sdd:002-totals-row"
proofbound sync all
proofbound warranty bundle "$base" -o "$W/release-warranties"
ls "$W/release-warranties"
```

**What you should see** (exit code 3: "written, but not everything is VALID"):

```text
warranty …/release-warranties/warranty-e620a48ca48b.json … claims=sdd:001-csv-export VALID reviewers=demo-reviewer:7a7df12f53e632bd
not warranted ef9f36d9f27e sdd:002-totals-row NO REVIEW  …
bundle …/release-warranties (8050d0657b71..ef9f36d9f27e): 1 of 2 claims verified (1 VALID, 0 UNREVIEWED); VERIFY.md, index.json, badge.svg
proofbound: bundle: written, but not every claiming commit is VALID
VERIFY.md
badge.svg
index.json
warranty-e620a48ca48b.json
```

Open `$W/release-warranties/VERIFY.md` in VS Code's preview (`code "$W/release-warranties/VERIFY.md"`)
and look at `badge.svg` too.

**What just happened.** The bundle holds one signed warranty per verified commit, plus three unsigned
conveniences: `VERIFY.md` (a page telling a buyer how to check each file, which also says which commits
are not covered and why), `index.json`, and `badge.svg` ("1 of 2 claims verified", for a README). The
unreviewed totals-row change is listed honestly as NO REVIEW, not hidden.

**Business value.** Every release can ship with its proof attached, honest about its gaps. A buyer sees
what is covered and what is not before they trust either.

## Step 13 — Drift: the spec changes after it was verified

**What you're doing.** As the spec owner, tightening promise FR-002 of the *already verified* feature.
Then asking Proofbound what moved.

```bash
sed -i 's/start every export with a header row\./start every export with a header row naming each column in lower case./' specs/001-csv-export/spec.md
git diff
git add specs
PROOFBOUND_NO_INTENT=1 git commit -q -m "spec: tighten the header-row promise"
proofbound sync all
proofbound report drift
```

**What you should see:**

```text
drift change=CI-001-csv-export-184wxu state=STALE verified_commit=e620a48ca48b … verified_revision=46d1df8f current_revision=00aeaa1e
  changed O-FR-002 "System MUST start every export with a header row." -> "System MUST start every export with a header row naming each column in lower case."
summary warranted=1 current=0 stale=1 other=0
```

**What just happened.** The old commit is still warranted for what the spec said *then* (its warranty
from step 11 still verifies). But Proofbound marks the feature `STALE` and names exactly which promise
changed, old wording and new, so someone re-verifies.

**Business value.** You find out when a promise you sold is no longer the promise you verified, before
a customer does.

## Step 14 — The record is reproducible

**What you're doing.** Throwing away every derived view of the record and rebuilding it from the
append-only log, then asking the same question as before.

```bash
proofbound rebuild
proofbound check "$base"
```

**What you should see** (`rebuild` prints nothing; `check` exits 1 because the totals row is unreviewed):

```text
Proofbound: NOT WARRANTED (1 of 2 claiming commits verified)
  e620a48ca48b sdd:001-csv-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
  ef9f36d9f27e sdd:002-totals-row NO REVIEW  …
Drift:
  drift change=CI-001-csv-export-184wxu state=STALE …
    changed O-FR-002 "System MUST start every export with a header row." -> "System MUST start every export with a header row naming each column in lower case."
  summary warranted=1 current=0 stale=1 other=0
proofbound: check: not warranted
```

**What just happened.** Same answers after a full rebuild. Every view is derived from git and the
event log, not stored as editable state.

**Business value.** The record is not a database someone could quietly edit; anyone can rebuild it and
get the same answer, so it holds up to scrutiny.

---

## Try it yourself: review the totals row

You now know every move. To make step 14 say WARRANTED, review `002-totals-row` as `demo-reviewer`,
reusing steps 7–10: `proofbound review draft <its commit>` (from `git log --oneline`), set `draft` to its
`draft.json`, `decide demo-reviewer`, seal with `PROOFBOUND_KEY="$W/reviewer.key"`, commit, `sync all`,
`check "$base"`. (If you open a new terminal, paste step 0b and the `decide` line from step 8 again.)

## Clean up

```bash
cd ~ && rm -rf "$HOME/proofbound-manual-demo"
```

Close the terminal too, so `XDG_CONFIG_HOME` goes back to normal.

## Where next

- [The tour](DEMO.md): the same story run by a script, with the full caveats.
- [The quickstart](../mvp/QUICKSTART.md): use Proofbound on your own repository.
- [Start here](README.md): install, every command, and release notes.
