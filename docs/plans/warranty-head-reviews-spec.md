# Spec: a warranty cites only reviews committed at HEAD

**Status:** draft 1, 2026-11-23, owner chose "fix it first" when the agent demo found the bug; reviewed
with the code by one non-author (§4). **Implements:** `CI-warranty-head-reviews-xa1ljk` →
`BR-warranty-head-reviews-gdjrrv` O-1 (requirement `proposed`). **Contract row:** CLI-INV-64.

## 0. The bug

An agent's branch is reviewed, rejected, and redone on a fresh branch from the spec. On the same
machine, `warranty export` and `warranty bundle` for the redo then fail:

```text
not warranted 8c3e13fe7d6e sdd:001-csv-export NOT EXPORTABLE  warranty: would not verify against this
repository: artifact docs/verification/verdicts/RR-45fc6d62ad9d-1.md is not committed at HEAD
```

The local ledger is append-only and keeps every review it ever synced, including the rejected
branch's requirement review. The warranty body cites every review of a targeted requirement, so it
cites one that the redo's history does not hold, and the export's own verification (W-INV-5) refuses
it. A fresh clone, and so CI and the GitHub Action, never synced that review and is not affected.

## 1. The rule

When a warranty is built, its requirement reviews are only those whose artifact is a regular file
(mode 100644 or 100755) committed at HEAD. Nothing else changes:
- Verdicts are already bound to the reviewed commit.
- The export still verifies the result exactly as a recipient would, so a requirement left with no
  VERIFIABLE review at HEAD still fails loudly ("not reviewed"); it is never silently warranted.
- The gate is unchanged.

## 2. Non-goals

- No change to the ledger, to `sync`, or to how the readiness gate counts reviews.
- No pruning of the ledger.

## 3. Tests

- `TestReviewsAtHead` (cli): a review committed at HEAD is kept. Dropped: one only in the working tree,
  one on another branch, one deleted at HEAD, a symlink, a directory, a missing path.
- `TestWarrantyIgnoresReviewsOffHead` (sdd, end to end through `cli.Run`): a first attempt is reviewed
  NOT_SATISFIED and abandoned; the redo from the spec is reviewed SATISFIED. Its warranty exports,
  cites only the redo's review, and verifies VALID against a fresh clone of the redo branch.
- **Calibration:** without the fix the end-to-end test fails with the bug's message (positive
  control); with the mode check removed, `TestReviewsAtHead` fails.

**Acceptance:** the agent demo's redo bundles VALID in the same repository that holds the rejected
branch.

## 4. Review

`warranty-head-reviews-review-20261123`: **ACCEPTABLE** (F1..F3 LOW, recorded per Law 11, not fixed).
- **F1:** the test does not pin the `100755` arm or the exact path compare (safe direction: a false
  refusal, never a false warranty).
- **F2:** one `git ls-tree` per review per commit, and a git error shows up as "not reviewed".
- **F3:** no committed test pins "a requirement with no review at HEAD still fails"; it was shown by
  the reviewer's mutation.
