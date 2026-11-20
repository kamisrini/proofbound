---
schema: vera.verdict.v1
verdict_id: mvp-4-spec-review-round1-20261120
status: NEEDS_WORK
reviewed_commit: d9a9c74433ec7e882f9e316ddf2a6668a28597f2
findings:
  - finding_id: mvp-4-spec-review-round1-20261120-F1
    severity: MED
  - finding_id: mvp-4-spec-review-round1-20261120-F2
    severity: MED
  - finding_id: mvp-4-spec-review-round1-20261120-F3
    severity: MED
  - finding_id: mvp-4-spec-review-round1-20261120-F4
    severity: LOW
  - finding_id: mvp-4-spec-review-round1-20261120-F5
    severity: LOW
  - finding_id: mvp-4-spec-review-round1-20261120-F6
    severity: LOW
  - finding_id: mvp-4-spec-review-round1-20261120-F7
    severity: LOW
  - finding_id: mvp-4-spec-review-round1-20261120-F8
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-4-spec-review-round1-20261120.md
artifact_sha: 6924bd5d4e5ca68042352e9fdfb077237f17a18d1d2651defac8033e613d1ac2
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet), spec round 1 of 2 (Law 11) on
`docs/plans/mvp-4-shareable-proof-spec.md` draft 1, under `CI-mvp-share-proof-w7c4na`.

## Findings (answered in draft 2 §6)
- **F1 MED.** The badge and page overstated scope (caller-chosen base, no range on the badge).
- **F2 MED.** The exporter fingerprint "to pin" came from the same unsigned channel; reviewer pinning was
  not actionable; proven and named reviewers were conflated.
- **F3 MED.** Untrusted commit, gate and artifact text went into the page unescaped.
- **F4-F8 LOW.** Exit 1 overloaded; badge over check states; base/HEAD/dirty-tree semantics; loose
  invariants; staleness after cosign. Taken where cheap.
