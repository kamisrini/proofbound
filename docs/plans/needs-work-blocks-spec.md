# Spec: a NEEDS_WORK review blocks the commit

**Status:** draft 2, 2026-11-24, built. Draft 1 was reviewed once by a non-author (§6: NEEDS_WORK, 2 HIGH,
3 MED, 3 LOW, all folded in). The owner said "go" after the live agent run showed the defect.
**Implements:** `CI-needs-work-blocks-h4r8wq` → `BR-change-area-p59cyd` O-2 and
`BR-tests-can-fail-k2h06j` O-5. **Contract rows:** GATE-INV-17, P-INV-35, W-INV-3 (amended).

## 0. The defect, in one line

A reviewer who marks every promise SATISFIED but rejects a line outside the spec's area (C10), or a
part no test caught (C11), seals a verdict whose status is `NEEDS_WORK`. Yet the pull-request check
reads **WARRANTED** and `warranty bundle` issues a **VALID** warranty for that commit.

**Why:**
- The readiness gate, the warranty rule and the drift report count only each promise's outcome. They
  never read the verdict's own status, which is the only place the area and tests rejections land.
- A sealed NEEDS_WORK verdict carries no findings, so its status reaches no ledger view at all.
- The scripted agent demo also marks one promise NOT_SATISFIED, which alone blocks the check. So no
  test, demo or review covered "every promise met, something else rejected". The first live agent
  run did exactly that (2026-11-24).

## 1. Behaviour after the fix

1. **The ledger keeps each verdict's status.**
   - `obligation_verdicts_view` gains `verdict_status`, set from the v2 verdict's `status`.
   - Like `outcome`, it follows the latest revision of each verdict id. A reviewer who deletes a
     verdict file and seals again under the same id replaces it, as today.
2. **The readiness gate blocks it.** The `intent-delivery-readiness` rule returns BLOCKED for a
   commit when any of its current verdicts (the latest revision of each verdict id) has status
   NEEDS_WORK, whatever the counts. Its reason reads
   `the reviewer recorded NEEDS_WORK (<verdict ids>); <the usual counts>`.
   - `check`, and the GitHub Action, print BLOCKED with that reason, so the pull request is NOT
     WARRANTED.
   - The Audit room's C1 shows the commit as an exception.
   - `warranty export` and `warranty bundle` refuse it. They already refuse any commit whose gate is
     not PASS (W-INV-2).
3. **The warranty rule refuses it too.** `CheckRule` fails when any verdict in the body has status
   NEEDS_WORK.
   - So `warranty verify` reports INVALID for a warranty issued before this fix that cites a
     NEEDS_WORK verdict, with or without a repository.
   - `ui --bundle` shows it INVALID too, because it uses the same check. That includes bundles
     already published (decision D2).
4. **Reports stop calling it verified.**
   - `report drift` lists a change as warranted only when every outcome is SATISFIED **and** the
     verdict is not NEEDS_WORK.
   - `report intent` shows a SATISFIED outcome from a NEEDS_WORK verdict as `NEEDS_WORK`, so the
     change reads `IMPLEMENTED_UNVERIFIED`, not SATISFIED.
5. **Existing ledgers are migrated in place, atomically.** Inside `Ensure`'s single transaction,
   `ALTER TABLE obligation_verdicts_view ADD COLUMN IF NOT EXISTS verdict_status TEXT`, then the
   backfill:

   ```sql
   UPDATE obligation_verdicts_view ov SET verdict_status = e.payload->>'status'
   FROM events e WHERE ov.verdict_status IS NULL AND e.event_id = ov.event_id
   ```

   This runs before `Apply` reads its checkpoint, so an idle ledger with no new events is migrated
   too. There is no drop, no replay and no window for two processes. A row the backfill cannot fill
   stays NULL. The gate treats NULL as **not** ACCEPTABLE: such a commit is BLOCKED with reason
   `verdict status unknown; run proofbound rebuild`, so the migration fails closed.
6. **The snapshot includes it.** `Snapshot` selects `verdict_status` as its last column, so P-INV-5
   (rebuilt rows equal incremental rows) covers it, and the JSON column index stays where it is.

## 2. HEAD and ledger history

- **CI uses a fresh ledger.** The Action runs `sync all` into a fresh ledger built from the checkout,
  so it sees only the verdict files at HEAD. That is the authoritative pull-request answer.
- **A long-lived local ledger can keep more.** It may also hold verdicts that HEAD no longer has
  (deleted, or sealed on another branch). For SATISFIED rows this already fails open: an old
  SATISFIED counts locally. For NEEDS_WORK it fails closed: an old NEEDS_WORK blocks locally while CI
  passes.
- **The remedy for a confusing local result** is `proofbound rebuild` after syncing, or
  a fresh ledger. The BLOCKED reason names the verdict ids, so the stale one is visible.
- **The end-to-end test** asserts that a fresh ledger built from HEAD and the long-lived ledger agree
  when no verdict file was deleted.

## 3. What it is not

- **Not a change to sealing.** Seal already writes NEEDS_WORK on a reject, on a whole change not
  caught or not undone, and on a failing baseline. Only what reads the status changes.
- **No new event kind and no change to verdict files.** The status is already in every verdict.
- **Not a ban on re-review.** A reviewer may re-seal a commit, and the latest revision of a verdict id
  decides, as for outcomes (D1). Fixing the code in a new commit stays the advice.

## 4. Tests (planned)

- **gates:**
  - all obligations SATISFIED and VERIFIABLE, but the verdict is NEEDS_WORK: BLOCKED, and the reason
    names the verdict;
  - the same verdict ACCEPTABLE: PASS;
  - a NEEDS_WORK verdict beside SATISFIED rows from another verdict id: BLOCKED;
  - re-sealed under the same id (NEEDS_WORK, then ACCEPTABLE): PASS;
  - NULL status: BLOCKED with the rebuild reason.
- **projections:**
  - the column is set and follows the latest revision;
  - a view created without the column, with a NEEDS_WORK row and a checkpoint above 0, gains the
    status on the next `Apply` even with no new events;
  - the snapshot carries the column.
- **warranty:** a body re-signed with a NEEDS_WORK verdict fails `CheckRule`.
- **drift and intent report:** the live agent case is not listed as warranted, and reads
  `IMPLEMENTED_UNVERIFIED`.
- **End to end** (the sdd package): every promise SATISFIED plus a rejected out-of-area hunk.
  - `check` reads NOT WARRANTED / BLOCKED.
  - `warranty bundle` issues no warranty.
  - The redo reviewed ACCEPTABLE reads WARRANTED.
  - A fresh ledger agrees with the long-lived one.
- **Mutation calibration:**
  - positive control: dropping the NEEDS_WORK condition in the gate must fail a test;
  - neutral control: a comment edit must survive.

## 5. Decisions (defaults taken; the owner may change them)

1. **D1, the latest revision of each verdict id decides,** as for outcomes. Across different verdict
   ids for one commit, any current NEEDS_WORK blocks, even if another id says ACCEPTABLE. Draft 1's
   "a NEEDS_WORK stands for ever" could not be built consistently: a re-seal reuses the id, and CI
   sees only HEAD (review HIGH 1, 2).
2. **D2, warranties issued before the fix that cite a NEEDS_WORK verdict verify INVALID,** in
   `warranty verify` and `ui --bundle`, published bundles included. They were wrongly issued. None is
   known outside demos.
3. **D3, the scripted demo is unchanged.** It still marks FR-002 NOT_SATISFIED. The live demo page
   will show the case this fix covers.

## 6. Review and amendments (owner-visible)

Round 1, one non-author, on draft 1: **NEEDS_WORK**. Every finding was folded into draft 2:
- **HIGH 1.** The update rule was unstated, and D1 contradicted how a re-seal reuses a verdict id.
  Amendment: the latest revision decides (§1.1, D1).
- **HIGH 2.** A rule based on ledger history would make local and CI disagree. Amendment: §2, the
  CI-agreement test.
- **MED 1.** Drift listed a NEEDS_WORK verdict as warranted. Amendment: §1.4.
- **MED 2.** The intent report read SATISFIED. Amendment: §1.4.
- **MED 3.** The migration was non-atomic, and missed idle ledgers. Amendment: an in-place
  ALTER and backfill inside `Ensure`, with NULL failing closed (§1.5).
- **LOW 1–3:** snapshot column (§1.6); missing tests (§4); published bundles named (D2).

## 7. Code review (owner-visible)

`needs-work-blocks-review-20261124`: **ACCEPTABLE** (no HIGH; M1 and M2 fixed; L1–L6 recorded).

- **M1, fixed.** The agent answer (`/agent/next`) still said `done: every promise of the change is
  proven` for a commit BLOCKED by a NEEDS_WORK verdict. A BLOCKED commit now adds a `fix_code` step
  carrying the gate's reason (CLI-INV-48 amended; `TestAgentNext`). The Promises page still shows such
  a promise as proven, which is true of the promise itself.
- **M2, fixed.** The change intent's scope now names `kernel/internal/cli` (drift, the agent answer)
  and the in-place migration.
- **Recorded, not fixed (Law 11):**
  - **L1:** `report drift` groups by verdict id, so a commit with one NEEDS_WORK verdict and one
    ACCEPTABLE verdict under different ids still lists as warranted.
  - **L2:** "the latest revision" holds per obligation row. A re-seal under the same id that drops an
    obligation row leaves that row's old status.
  - **L3:** the backfill UPDATE runs on every `Ensure`. It is cheap once nothing is NULL, but it is not
    free.
  - **L4:** the fresh-versus-long-lived ledger test from §4 is not written (a `rebuild` stands in), and
    no test sends a pre-fix warranty through `warranty verify` end to end.
  - **L5:** the intent report maps a NULL status to NEEDS_WORK. That fails closed, and the
    behaviour is now stated.
  - **L6:** `agent-demo.sh` must be re-run. Done before release.
