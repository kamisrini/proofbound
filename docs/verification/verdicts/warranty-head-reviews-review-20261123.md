---
schema: vera.verdict.v1
verdict_id: warranty-head-reviews-review-20261123
status: ACCEPTABLE
reviewed_commit: 7639357d497fe229b87daef7f73fd1ccca06d2ca
findings:
  - finding_id: warranty-head-reviews-review-20261123-F1
    severity: LOW
  - finding_id: warranty-head-reviews-review-20261123-F2
    severity: LOW
  - finding_id: warranty-head-reviews-review-20261123-F3
    severity: LOW
artifact_path: docs/verification/verdicts/warranty-head-reviews-review-20261123.md
artifact_sha: 7885d53177fd418bebbe49ff8572adbc2d422a5dca082cd79c5541fb58100cd1
---
## Provenance
Non-author review (Build Law 9), the one code round (Law 11), of the staged change (`git diff --cached`) on top of HEAD 7639357d497fe229b87daef7f73fd1ccca06d2ca. Read docs/plans/warranty-head-reviews-spec.md, the SPEC.md intro paragraph and CLI-INV-64 row, kernel/internal/cli/warranty.go (buildWarranty, warrantyBody, the new reviewsAtHead), gitOutput in starter.go, the other buildWarranty callers (bundle.go:134, audit.go:226 C3), kernel/internal/warranty/verify.go (verifyThrough8, blobAtHead, artifactsAtHead) and warranty.go (rule verdict and review coverage loops incl. "not reviewed VERIFIABLE"), W-INV-2 and W-INV-5, review seal's RR-<commit12>-<n> naming (starter.go:468), both new tests, and the BR/CI record shapes. All runs were in a scratchpad copy of kernel/ (`review-whr/`), go1.27.4, one go test invocation at a time; no repository file was modified.

TestReviewsAtHead mutations (cli). Calibration: positive control M1 (reviewsAtHead keeps everything) died; neutral control M3 (the redundant `strings.Contains(path, "\x00")` removed) survived. Then: M2 mode check removed died (the spec's calibration claim holds); M5 `-z` dropped died (trailing newline breaks the exact path compare); M4 exact path compare removed survived; M6 `100755` dropped survived; M7 `blob` type check removed survived (equivalent: 040000/160000 are never in the accepted mode set).

TestWarrantyIgnoresReviewsOffHead (sdd, ~20s/run): baseline ok; positive control (reviewsAtHead call removed) FAIL with the bug's exact message (`artifact docs/verification/verdicts/RR-762b29ca10f0-1.md is not committed at HEAD`); neutral (NUL check removed) ok; extra mutation (reviewsAtHead drops everything) FAIL with `obligation O-FR-001 of BR-001-export-b5loax is not reviewed VERIFIABLE`, showing by experiment that a requirement left without a review at HEAD still fails loudly.

Odd-path probe (throwaway test, deleted): committed files named with a space, tab, `"`, unicode, newline, leading `-`, `*`, `\` or `:(top)` were all kept (`-z` disables quoting; splitting on the first tab is correct because the metadata holds none). Dropped: `star?.md` (no glob match of `starA.md`; ls-tree disables wildcards), a `./`-prefixed path, a trailing-slash directory, an empty name.

Not run: `make check`, the full sdd/cli suites, the agent demo acceptance, and digest validation of the BR/CI records beyond shape. On shape: the CI target's artifact_sha256 matches the BR's self-digest (475a89a9…), IDs match the pattern, scope and exclusions agree with the diff, and the BR is `proposed` pending its requirement review.

## Status
ACCEPTABLE: three LOW findings, no HIGH or MED.
- Q1, can this weaken the guarantee? No. Filtering only removes entries from body.Reviews, and CheckRule gives a review only positive power (VERIFIABLE adds coverage; nothing in a review blocks), so removal can never turn a failing export VALID; at worst a requirement becomes "not reviewed VERIFIABLE", a loud refusal (demonstrated). Verdicts are untouched and stay bound to b.Commit. A same-path, different-content review at HEAD passes the filter (path and mode only) and is then refused by artifactsAtHead ("differs from the warranty"). The gate is evaluated from the full ledger before filtering and is unchanged. warranty export, warranty bundle and the Audit room's C3 all go through buildWarranty and get the same filter and self-verify. Filter and self-verify both use HEAD (self-verify requires HEAD to descend from the warranted commit), which holds for a detached HEAD or a worktree (`-C root` resolves that worktree's HEAD). The prior behaviour where a committed-then-deleted review blocked the export is gone; that review never had blocking power under the rule.
- Q2, path parsing: sound (probe above). A NUL cannot reach git through exec; a git failure returns "" and the review is dropped, the safe direction.
- Q3, tests: both calibration claims in spec §3 hold as written.
- Q4, spec and SPEC row: nothing false; the row names all three entry points, and the mode set matches what blobAtHead accepts once git normalises modes.

## Findings
- **F1 (LOW).** kernel/internal/cli/reviews_at_head_test.go:15-42. Two branches of reviewsAtHead are not pinned: the `f[0] == "100755"` arm (M6 survives) and the exact `path != r.ArtifactPath` compare (M4 survives). Both failure directions are safe (loud export refusal, or blobAtHead refuses): the cost is false refusals, not false warranties. Fix: add a committed `chmod +x` review that must be kept and a `./docs/...` path that must be dropped.
- **F2 (LOW).** kernel/internal/cli/warranty.go:432-444. One `git ls-tree` process per targeted review on every buildWarranty call; bundle and C3 call it per commit, so a long range spawns about commits × reviews processes. gitOutput swallows errors, so a git failure (cancelled context, broken repository) surfaces as the misleading "obligation … is not reviewed VERIFIABLE". Fix: one `git ls-tree -r -z HEAD -- docs/verification/verdicts` parsed into a set, returning the git error.
- **F3 (LOW).** docs/plans/warranty-head-reviews-spec.md §1/§3 and the CLI-INV-64 row. "A requirement left with no VERIFIABLE review at HEAD still fails loudly" is backed by reasoning and this review's manual mutation, but no committed test pins it; a later change that skipped the self-verify for bundles would not be caught here. Fix: in TestWarrantyIgnoresReviewsOffHead, export the first attempt's commit from the redo checkout, or export after `git rm` of the redo's review, and assert a non-zero exit with "not reviewed VERIFIABLE".
