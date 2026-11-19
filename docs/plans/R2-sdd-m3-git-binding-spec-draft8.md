# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 8)

**Status:** draft 8, 2026-11-19. It fixes the three LOW findings of non-author round 7,
[`r2-m3-spec-review-round7-20261119`](../verification/verdicts/r2-m3-spec-review-round7-20261119.md)
(**ACCEPTABLE**, 3 LOW), over [draft 7](R2-sdd-m3-git-binding-spec-draft7.md). Round 7 judged that
these "can be fixed in the SPEC text or recorded by the owner without another adversarial round".
Draft 8 restates only what changes. Drafts 7, 6 and 5 stand otherwise, and their union is the
contract. Owner decisions D1-D6: as in draft 7 §3.

**Next step:** the owner approves the change-intent wording (§8). Then code lands under that intent.
The implementation gets its own non-author review (Law 9).

## 0f. Round 7, answered

| Finding | Answer in draft 8 |
|---|---|
| F1 LOW — same-identity classification depends on path order | **Duplicate identities are refused as a group, whatever the order** (§4.6). Before processing, the connector counts the v2 and requirement-review artifacts listed in this sync by `(schema, native id)`. Every artifact whose identity appears more than once is refused with cause `duplicate-identity`, so the sync fails naming all of them. Classification then depends only on the ledger snapshot, which is the same in both of `verify`'s passes. The "update the set on append" rule is no longer needed and is dropped. This repository has no duplicate v2 or review ids today (3 artifacts, 3 ids) |
| F2 LOW — reviews SPEC sentences, cursor, malformed-after-refusal, quoting, §7.3 residue | §5 now also amends the reviews SPEC **Purpose** sentence and **R-INV-4**. Refused paths are **not** in the cursor. A malformed artifact still stops the sync at once, and the returned error joins the malformed error with every refusal recorded so far. On **any** error, the CLI prints the counts, held and refused lines it has before it returns. Refused paths go through `printableUnit`, as held paths do. §7.3 now defines `gates/x5.yaml` and states how step 2 computes its references |
| F3 LOW — the "cannot pass silently" rationale over-claims | It is narrowed: **a refused revision** cannot pass silently on any path that syncs before it enforces. Three **pre-existing** stale-outcome paths are recorded in §9, raised to the owner, and not claimed to be closed. M3 neither causes nor worsens them: the `obligation_verdicts_view` upsert never updates `commit_sha`, a revision moving to a new requirement revision leaves the old rows, and reverting to identical bytes is deduplicated as "existing" |
| Observations — the decoder; permanent red | The view decodes git payloads with `encoding/json.Unmarshal` into `connectorgit.Commit`, which is no stricter than projections' `decode`. The accepted cost (draft 7 §3) gains: a refused revision whose prerequisite will **never** be ledgered keeps `sync all`, `verify` and delivery-enforce red until the artifact is fixed. For a first revision, the same defect is only a held line. This asymmetry is intended |

## 4.6 The reviews hold and refusal: changes to draft 7

- **Identity set.** It is the ledger snapshot of `(schema, native id)` for `reviews` events, read in
  the view's single pass. The connector does not add to it during the sync.
- **Duplicates first.** Among this sync's listed v2 and requirement-review artifacts, any
  `(schema, native id)` that occurs more than once makes **every** artifact with that identity
  `refused duplicate-identity`. Prerequisites are not consulted for them.
- **Then the rest,** classified as in draft 7: present → appended; missing and not in the snapshot →
  held; missing and in the snapshot → refused.
- **Cursor:** neither held nor refused paths are included.
- **Errors.**
  - A malformed artifact stops the sync immediately, as today (R-INV-5). The returned error is
    `errors.Join` of the malformed error and, if any refusal was recorded, the
    `ErrRevisionRefused` error.
  - The CLI prints the counts line, the held lines and the refused lines it has (each path through
    `printableUnit`) before returning **any** error from the reviews step.
- **Rationale, narrowed.** A refused revision is never appended, and every sync that sees it fails.
  So it cannot pass silently on a sync-first enforcement path. **This is not a claim** that stale
  outcomes in general are impossible: see §9.

## 5. Sites: changes to draft 7

| Site | Change |
|---|---|
| `connector/reviews/SPEC.md` Purpose | "each valid schema-bearing artifact emits one … event; changed content a revision" gains "unless it is held or refused (R-INV-13)" |
| `connector/reviews/SPEC.md` R-INV-4 | "changed content with the same verdict ID produces a new event" gains "once its order-only prerequisites are ledgered; until then it is refused (R-INV-13)" |
| `connector/reviews/SPEC.md` R-INV-13 | covers the hold, the refusal and `duplicate-identity` |
| `cli/cli_test.go` nil-store test (`syncReviewsOnStore` with a nil ledger) | keeps expecting the store's error. The view is built lazily after `beginSync`, so a nil store still fails at `beginSync` first |

## 6. Invariants: change to draft 7

12. **M3-INV-12** — A later revision (its `(schema, native id)` is in the ledger snapshot) is never
    ledgered before its prerequisites.
    - It is refused: the remaining artifacts are still processed, and then the sync fails, naming it.
    - Every artifact whose identity is listed more than once in one sync is refused
      (`duplicate-identity`), whatever the path order.
    - With prerequisites present and a unique identity, a later revision is appended as before M3.
    - v1 verdicts are never checked.

## 7. Test derivations: changes to draft 7

- **Row 11** replaces its "two files with one identity" case with two cases:
  - such a pair in **both** path orders; both files are `refused duplicate-identity`, and the sync
    fails in each order;
  - a malformed artifact after a refusal; the error joins both, and the CLI still prints the refused
    line.

  It drops the "snapshot updated on append" kill and adds "classification depending on path order".
- **§7.3 step 2:** the R'@d2 and CI@d2 references are computed **without syncing `sdd`**, by running
  the `sdd` provider's `Revisions` over the tenant (as in draft 6 §7.2 step 2). Reading them from the
  ledger would ledger R'@d2 and defeat step 4's `requirement-not-ledgered` refusal.
- **§7.3 step 3:** `gates/x5.yaml` is a JSON `proofbound.gate.v1` with:
  - a distinct `id`;
  - a `description`;
  - `expires` `2099-01-01`;
  - `mode` `canary`;
  - `rule` `intent-delivery-readiness`;
  - `scope_commit` X5's full sha.

## 8. Proposed change intent (owner approves the wording)

`CI-sdd-warranty-m3-<suffix>`: implements `BR-sdd-warranty-i9miv7` O-8, and O-10 for claims. Three
parts:
- the git connector binds `Intent: sdd:<name>` to the exact ledgered `intent.sdd` change-intent
  revision, with per-commit, all-or-nothing claim skips;
- the reviews connector holds first revisions and refuses later revisions whose order-only
  prerequisites are not yet ledgered;
- projections accept provider `sdd`.

The sites are draft 5 §5 as amended by drafts 6-8. The non-goals are §9.

## 9. Non-goals: additions to draft 6

**Pre-existing stale outcomes** (raised to the owner, not filed). The `obligation_verdicts_view`
upsert keys on `(verdict_id, requirement…, obligation_id)` and never updates `commit_sha` or the
intent columns. So:
- a revision that changes only `reviewed_commit` moves its outcome onto the **old** commit's row;
- a revision that moves to a new requirement revision leaves the old rows' outcomes in force;
- reverting an artifact to byte-identical earlier content is deduplicated as "existing", so the later
  outcome stays.

All three predate M3 and are reachable with `records` claims today. M3 makes them reachable for
`sdd` commits too, and does not change them.
