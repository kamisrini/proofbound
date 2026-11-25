# Proofbound v0.3 release notes (v0.3.0, v0.3.1, v0.3.2)

**Release 3, first public cut: install it, gate pull requests, and hand over signed warranties.**
Tag `v0.3.0` · Decisions [`VD-r3-warranty-first-s59ktv`](../../decisions/VD-r3-warranty-first-s59ktv.md)
and [`VD-vertical-mvp-first-ai3dhn`](../../decisions/VD-vertical-mvp-first-ai3dhn.md) · Scope in
[`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md).

**Status:** R3 slice 1 accepted by the owner 2026-11-20. MVP steps 1 to 4 built 2026-11-20, each with
its spec and code reviewed by a non-author (Law 11 budget). Published by the owner's decision of
2026-11-20.

## v0.3.2 (2026-11-23): a UI for compliance, reviewers, spec owners and agents; unclaimed code fails the check; agents get the claim added for them; line-level scope, signed waivers, and release evidence in a browser

- **`proofbound ui`** opens the Audit room in your browser, served by the binary on `127.0.0.1` only.
  Run it in a repository, open the link it prints, and pick the range to audit (by default, everything
  since the last tag). It answers twelve controls, each with the commits behind its answer and the
  terminal command that gives the same answer:
  1. Every commit that claims a spec passed its readiness gate.
  2. No commit changed code without claiming a spec.
  3. Every passing claim carries a warranty that verifies.
  4. Promises that changed after they were verified are disclosed.
  5. Every commit in the range is accounted for.
  6. Changes to who may sign (declared keys, the sdd profile), to the check's workflow, or to an
     existing gate are listed for review.
  7. Each claimed spec existed before the code that claims it.
  8. No change was reviewed by its own author, agent or spec owner.
  9. Commits are signed by a key the repository declares (a disclosure).
  10. Each agent change stayed inside the area its spec declares (see line-level scope below).
  11. Each claimed change's tests fail when its code is undone (see tests that can fail below).
  12. Traced changes say which agent sessions wrote their lines, in records that verify.
- **Agent oversight.** The page shows who did the work: every agent (from the commit author or a
  `Co-Authored-By` line naming Claude, Copilot, Cursor, Codex and similar) and every person, with
  their commits, claims, code without a spec and open exceptions; who reviewed each change, and
  whether that reviewer was independent of its author, agent and spec owner (control 8); which
  commits are signed by a key the repository declares (control 9); and the Claude Code sessions in
  the period (times and counts, never content) with the commits made while each ran.
- **`proofbound keys git-sign`.** Run once in a repository and your declared Proofbound key signs your
  git commits, so the Audit room can prove who (or which agent) made each one instead of trusting the
  author name.
- **Fixed:** the sessions connector read no real Claude Code logs (their lines exceed 64 KiB) and did
  not count Claude Code's file edits; both work now.
- **Four more workspaces in the same `proofbound ui`:**
  - **Now** (developer): what needs a person, each with the exact command; which task each
    worktree's agent holds; commits since your last visit.
  - **Review bench** (reviewer): only the changes you may review (never your own, your agent's or
    your spec's); each promise beside the change; record an outcome per promise and **seal and sign
    the review in the browser**, writing and committing exactly what `proofbound review seal` writes.
    The whole review, with no terminal: **Run the tests** (in the background, with progress; they
    undo each part of the change), decide each part the tests did not notice, and see every added
    line marked *outside the area*, *tests did not notice* or *no recorded agent edit*, beside the
    spec it claims. Sealing runs the tests again. After sealing, the page shows the check result.
  - **Changes**: each spec change as one flow, from the agent's commits to reviews, the check and
    the warranty, with your one next step ("Review this", "Issue the warranty").
  - **Promise map** (spec owner): every promise with its state (proven, changed after proof, not
    kept, claimed but not reviewed, not claimed) and any reviewer's finding that the words are unclear.
  - **Agent surface**: every page as JSON, and `/agent/next` gives a coding agent its task's
    promises, what blocks it, whether it may act, and the next command. Agents read with the token in
    `.proofbound/ui.json`; only people act.
- **Issue the warranty from the Audit room.** When every change of the period is warranted, one
  click previews and then writes the bundle into `docs/warranties/<head>/` and commits it. A partial
  release is refused, and a reviewer of a change cannot issue its warranty.
- **Check a warranty with nothing installed.** The release carries `proofbound-verify.html`: open
  it in a browser, type the vendor's key fingerprint (got by another channel, never from the bundle),
  choose the warranty files. It runs Proofbound's own verification code in the page, offline, and
  answers SIGNATURE ONLY or INVALID; `warranty verify --repo` remains the full check.
- **Time rail.** Pick any commit of the period and the Audit room shows the audit exactly as it
  stood then, computed from a checkout of that commit with its own record (nothing later can leak
  in), and lists which controls and exceptions differ from now.
- **Proof in your browser.** Every signed outcome, countersignature and reviewer signature the page
  shows is verified again by your browser itself (Ed25519), so you do not have to take the server's
  word for it.
- **Close exceptions without a terminal.** For each exception, choose accepted risk, fixed in a later
  commit, or reverted, and give a reason. The page shows the exact record first; "Sign and record"
  signs it with your own key on your machine and commits it under `.proofbound/outcomes/`. Only a key
  declared before the range began can sign for it, and nobody can close an exception on their own
  commit.
- **Countersign the transcript.** Download the audit as JSON, print it or save it as PDF, and sign it
  as the auditor: the attestation is committed under `.proofbound/attestations/`.
- **Safe by construction:** the link the command prints is the only way in (a strict cookie, never
  written into the page); other web pages cannot read or act, and another local program cannot
  without that link; the server holds the ledger only for the seconds it takes to answer, and a
  second request while one runs is told the ledger is busy.
- **Decision and contract:** [`VD-audit-room-ui-q2v6x3`](../../decisions/VD-audit-room-ui-q2v6x3.md),
  [U1 spec](../../plans/ui-u1-audit-room-spec.md) (CLI-INV-37..41), reviewed by a non-author twice
  (spec and code; every HIGH and MED finding fixed).
- **Line-level scope.** A spec folder may list the files its change may touch in a `scope.paths`
  file (paths and patterns such as `src/cart/` or `src/**/*_test.go`, read from the commit's parent so
  a commit cannot widen its own area). `proofbound review draft` lists every line range a claiming
  commit changed outside that area; `review seal` (and the Review bench) refuses until the reviewer
  accepts each with a reason or rejects it, keeps the decisions in a signed record, and a rejection
  makes the review NEEDS_WORK. Control 10 flags anything no independent reviewer accepted, an invalid
  area, and an area an agent changed. A spec without `scope.paths` is counted, not flagged. Spec:
  [change-area-spec.md](../../plans/change-area-spec.md).
- **Signed exception waivers.** For the rare commit that must change code without a spec (a hotfix,
  vendored code), `proofbound waiver sign <id> --path <pattern> --reason <text> --expires <date>`
  writes a signed waiver that lasts at most 90 days. A commit is WAIVED only when the waiver was
  committed before it, is signed by a key declared at the start of the range, by someone who is not
  the commit's author, co-author or signer, and has not expired. `proofbound check` lists it ("1 commit
  waived") and still passes; the release bundle, its badge and the Audit room show it; the commit hook
  lets it through. **Optional:** commit `.proofbound/waivers/policy.json` with
  `{"schema":"proofbound.waiver-policy.v1","require_signed_commits":true}` and waivers cover only
  commits signed by a declared key (`proofbound keys git-sign`), since a git author name is
  self-declared. Specs: [signed-waiver-spec.md](../../plans/signed-waiver-spec.md),
  [waiver-signed-commits-spec.md](../../plans/waiver-signed-commits-spec.md).
- **Release evidence in a browser, no clone.** `proofbound ui --bundle <folder>` opens the folder
  `proofbound warranty bundle` wrote as a read-only page, anywhere, with no repository or database.
  Each warranty shows its promises, verdicts, reviews and reviewers from the signed file, labelled
  with how far it was checked: SIGNATURE ONLY without a clone, VALID / UNREVIEWED / INVALID with
  `--repo <clone>`, MISSING or ALTERED when it disagrees with the (unsigned) index; `--trust-key`
  pins the exporter's key, and the browser re-verifies each signature. Spec:
  [ui-bundle-spec.md](../../plans/ui-bundle-spec.md).
- **Write-time perimeter for Claude Code.** `proofbound init` now writes `.claude/settings.json` (only
  when there is none) with a PreToolUse hook: while a task is active, an agent's Edit, Write,
  MultiEdit or NotebookEdit of a file outside the spec's committed `scope.paths` is refused before it
  happens, and the agent is told why and what to do. It is early feedback, not the guarantee: the
  agent controls its own machine (a shell command, or removing the hook, gets past it), so control
  C10 still judges every committed line. The settings file needs no claim to commit, and C6 lists
  every change to it. In an existing repository, run `proofbound init` again, or add the hook to your
  own `.claude/settings.json` (init prints how). While a task is active, the hook also refuses an
  agent's edit of the files that hold its own guard: `.claude/settings.json`,
  `.claude/settings.local.json` and `.proofbound/task` (any letter case, through any link). Spec:
  [hook-guard-spec.md](../../plans/hook-guard-spec.md).
- **Tests that can fail.** Commit `.proofbound/tests.json` with your test command and test folders,
  and `review draft` undoes each hunk of code a claiming commit changed, one at a time and then all
  at once, runs your tests in a temporary worktree, and prints whether they noticed ("tests did NOT
  catch undoing src/export.sh lines 1-2"). `review seal` runs them again and needs a signed decision
  (accept with a reason, or reject) for every hunk the tests did not catch; a reject, an uncaught
  whole change or a failing test run makes the review NEEDS_WORK. The signed `TESTS-` record feeds
  control C11 in the Audit room, which runs nothing itself. The declaration is read from the
  commit's parent, so a commit cannot weaken its own tests. It is opt-in, never runs in CI, and runs
  the commit's own code, so use it where you would run that commit's tests anyway. Undoing a hunk is
  not full mutation testing: a test can notice a removed hunk and still miss a subtle bug in it. Run
  `proofbound init` again in an existing repository so `.gitignore` stops hiding `.proofbound/tests.json`.
- **Session-to-line provenance.** `proofbound init` adds a second Claude Code hook (`PostToolUse`)
  that logs a hash of every line an agent's Edit, Write or MultiEdit adds (never the content), in the
  repository's git folder, so agents in worktrees are seen too. `review draft` then matches every line
  the commit adds: written by which session, with **no recorded agent edit** (a shell command, a
  person, a formatter, or an agent whose hook was not running), or too common to count (`}`,
  repeated lines). `review seal` matches again and writes a signed `PROV-` record, and control C12 in
  the Audit room discloses it; the counts never block. It is the reviewer's signed statement of what
  the logs on the agent's machine showed, not proof against a hostile agent, which could write to its
  own log. In a repository set up with an earlier v0.3.2, run `proofbound init` again: it prints the
  exact hook to add to your `.claude/settings.json`.
- **Review outcomes by agent.** `proofbound report agents <base>` (and a new section of the Audit
  room) shows, per coding agent, how its changes fared when someone checked them: ACCEPTABLE,
  NEEDS_WORK, reviewed but not independent, or not reviewed; the hunks outside the spec's area and the
  parts no test caught, with how each was decided; the lines a recorded session wrote and the lines
  no recorded agent edit wrote; and the open exceptions. It also lists the sessions in the provenance
  records. The counts come only from signed records the controls C10, C11 and C12 trust, never from
  what an agent says. They are counts, not a score, and the output says what they do not mean.
  `--json` gives the same for a script or an agent.
- **What each control maps to.** Every Audit room control names the criteria it gives evidence
  toward in SOC 2 (CC5.1, CC6.1, CC8.1), ISO/IEC 27001:2022 Annex A (for example A.8.32 change
  management, A.5.3 segregation of duties), NIST SSDF v1.1 (PS.1, PS.2 partial, PS.3, PW.7, PO.3) and,
  for providers of high-risk AI systems, the EU AI Act (Art. 17(1)(a)–(d) and (k), Art. 11 with
  Annex IV). Each entry has a one-line reason, on the card and in the transcript (`maps_to`); the
  whole mapping, with versions and caveats, is in
  [FRAMEWORK-MAP.md](../../compliance/FRAMEWORK-MAP.md). It is worded "maps to", meaning evidence
  toward a criterion, never compliance: whether a criterion is met is the auditor's judgment. EU AI
  Act Art. 12 and Art. 14 are deliberately not mapped, and the page says why.
- **Not yet:** anything an agent did outside the repository, which the Audit room says on the page it
  cannot see.

- **Stricter pull-request check (behaviour change).** `proofbound check` now lists every commit that
  claims no spec but changes code as `UNCOVERED`, and the pull request is **NOT WARRANTED**. Before,
  such commits were ignored, so one reviewed commit plus an unclaimed code commit read WARRANTED.
  Commits that only touch specs, `gates/`, `.proofbound/`, `docs/`, Markdown, `.gitignore` or the
  Proofbound workflow still need no claim. **Some pull requests that passed on v0.3.1 will fail.**
  Make the check a required status check, or a red check can still be merged.
- **`proofbound task start <change>`.** Run it once before writing code for a change: every commit
  that changes code then ends with `Intent: sdd:<change>` without anyone typing it. `proofbound task`
  shows the active task, `proofbound task stop` clears it. The change's spec folder must exist first.
- **New hooks from `init`.** `prepare-commit-msg` adds the claim; `commit-msg` refuses a code commit
  with no claim, a mistyped claim, or a claim outside the message's final trailer block (put it on the last line, after a blank line). `PROOFBOUND_NO_INTENT=1`
  is retired: housekeeping commits need no claim, and code commits cannot opt out.
- **Action needed for repositories set up before v0.3.2:** `init` never replaces an existing hook, so
  remove `.git/hooks/commit-msg` and run `proofbound init openspec` (or `speckit`) again to get both
  new hooks. The old hook still works, but keeps the retired opt-out.
- **`init` now stops ignoring `.proofbound/waivers/`.** Run `proofbound init` again in an existing
  repository to add the line, or commit a waiver with `git add -f`.
- **Fixed:** the reviewer agent `init` writes has valid YAML front matter (it was rejected by VS Code
  and strict parsers). `install.sh` now warns when an older `proofbound` earlier on PATH would run
  instead of the new one. The quickstart points to a from-scratch path for an empty folder.
- **Fixed:** after a reviewed branch was abandoned and the work redone on a new branch, `warranty
  export` and `warranty bundle` for the redo failed on the same machine ("… is not committed at
  HEAD"), because the warranty cited the abandoned branch's review. A warranty now cites only the
  reviews committed at HEAD. Fresh clones, and so the GitHub Action, were never affected.
- **Fixed:** a review that met every promise but rejected something else (a line outside the area, or
  a part no test caught) sealed NEEDS_WORK, yet the pull-request check read WARRANTED and the bundle
  issued a VALID warranty. A NEEDS_WORK review now blocks the commit, no warranty is issued, and a
  warranty issued before the fix that cites one verifies INVALID. Existing ledgers are migrated on
  first use. Found by the first live agent run. Spec:
  [needs-work-blocks-spec.md](../../plans/needs-work-blocks-spec.md).
- **New demo: [the agent goes off-script](DEMO-AGENT.md).** By hand, about 15 minutes: an agent's
  change edits a file outside its spec's area and ships a test that cannot fail; the line is named,
  the reviewer's finding is signed, and nothing is warranted until the redo.
- **New demo: [a real agent, live](DEMO-AGENT-LIVE.md).** Claude Code on your own account: the hooks
  refuse its out-of-area edit and its attempt to remove them; a person's hand edit is named by C10,
  C11 and C12; NEEDS_WORK blocks the pull request though every promise is met.

## v0.3.1 (patch, 2026-11-20): the pull-request check works in CI

- **Fixed:** the GitHub Action failed on every reviewed commit with `v2 verdict evidence … is
  dangling`. `review seal` now cites the reviewed commit by its SHA (`evidence_refs:
  ["git:commit.recorded:<sha>"]`), which every machine resolves, instead of an event id that existed
  only in the reviewer's local database (`DEF-VERDICT-EVIDENCE-PORTABILITY-001`, spec
  [`mvp-evidence-portability`](../../plans/mvp-evidence-portability-spec.md)).
- **Action needed:** verdicts sealed with v0.3.0 still work locally but still fail in CI. Seal them
  again with v0.3.1: remove that commit's `docs/verification/verdicts/OV-<commit12>-*` and
  `RR-<commit12>-*` files (with any `.sig`) and `gates/ready-<commit12>.yaml`, run `proofbound sync
  all`, then `proofbound review draft <commit>` and `review seal`. Verified on the first live Action
  run: after a re-seal the pull-request check reports WARRANTED in CI. A draft written by
  v0.3.0 is refused; run `review draft` again.
- **Upgrade together:** v0.3.0 cannot read a verdict sealed by v0.3.1. The Action installs the
  latest release unless its `version` input is pinned; reviewers should upgrade too.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
proofbound version
```

Linux and macOS (amd64, arm64); Windows users download `proofbound_windows_amd64.tar.gz`. The script
checks the archive against `SHA256SUMS` and never uses `sudo`. Quickstart:
[`docs/releases/mvp/QUICKSTART.md`](../mvp/QUICKSTART.md).

## What's new since v0.2.0

| What | Command | Where |
|---|---|---|
| Portable signed warranty: a file anyone can verify against a clone, no database | `keys init`, `warranty export`, `warranty verify` | [R3 demo](DEMO.md) |
| Reviewer proof: VALID only when each named reviewer's own declared key signed (at seal, or by cosigning) | `review seal` (signs), `warranty cosign`; verify exit 3 = UNREVIEWED | [`mvp-3` spec](../../plans/mvp-3-reviewer-signed-spec.md) |
| Release archives and a checksum-verified installer | `install.sh`, `proofbound version` | [`mvp-1` spec](../../plans/mvp-1-install-spec.md) |
| Pull-request check and GitHub Action; `init` writes the workflow | `proofbound check <base>`, `uses: kamisrini/proofbound@main` | [`mvp-2` spec](../../plans/mvp-2-pr-check-spec.md) |
| Release bundle: signed files, a verification page, an index and a badge | `proofbound warranty bundle <base>` | [`mvp-4` spec](../../plans/mvp-4-shareable-proof-spec.md) |

## Changes in behaviour

- `warranty verify` now returns **UNREVIEWED (exit 3)** for a warranty whose reviewer has not signed;
  warranties made before reviewer signatures existed verify UNREVIEWED until the reviewer cosigns.
- Reviewer and owner names are compared by key name: `Alice` and `alice` are one person.
- `proofbound check` lists a mistyped claim (for example `intent: SDD:x`) as UNBOUND instead of ignoring it.

## Known issue (v0.3.0 only; fixed in v0.3.1)

- **The GitHub Action fails on reviewed commits in CI.** A sealed verdict cites its evidence by an
  event id that exists only in the reviewer's local database, so `proofbound sync all` on a fresh CI
  runner stops with `v2 verdict evidence … is dangling` (`DEF-VERDICT-EVIDENCE-PORTABILITY-001`, found
  by the first live Action run, 2026-11-20). `proofbound check` and warranties work locally, and
  `warranty verify` is unaffected (it needs no database). Fixed in v0.3.1 (above).

## Caveats

- The repository is the trust anchor: anyone who can commit to it can declare a key. Clone from a
  source you trust and pin the exporter's fingerprint, obtained out of band.
- A reviewer signature proves the key declared for that name signed, not who holds it.
- The embedded database downloads a PostgreSQL build from Maven Central on first use; its published
  signature did not verify when we checked (`DEF-P7-001`). Set `DATABASE_URL` to use your own.
- Release archives are checksummed, not signed. Pin the action to a tag once you rely on it.
- Only OpenSpec and GitHub Spec Kit are supported.
