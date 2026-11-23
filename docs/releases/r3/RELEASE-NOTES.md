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
  since the last tag). It answers ten controls, each with the commits behind its answer and the
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
  - **Promise map** (spec owner): every promise with its state (proven, changed after proof, not
    kept, claimed but not reviewed, not claimed) and any reviewer's finding that the words are unclear.
  - **Agent surface**: every page as JSON, and `/agent/next` gives a coding agent its task's
    promises, what blocks it, whether it may act, and the next command. Agents read with the token in
    `.proofbound/ui.json`; only people act.
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
- **New demo: [the agent goes off-script](DEMO-AGENT.md).** By hand, about 15 minutes: an agent's
  change edits a file outside its spec's area and ships a test that cannot fail; the line is named,
  the reviewer's finding is signed, and nothing is warranted until the redo.

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
