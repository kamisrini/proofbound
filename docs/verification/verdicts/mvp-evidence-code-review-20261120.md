---
schema: vera.verdict.v1
verdict_id: mvp-evidence-code-review-20261120
status: NEEDS_WORK
reviewed_commit: 98b33a7826c0d41044cec495e69c793a46186512
findings:
  - finding_id: mvp-evidence-code-review-20261120-F1
    severity: MED
  - finding_id: mvp-evidence-code-review-20261120-F2
    severity: LOW
  - finding_id: mvp-evidence-code-review-20261120-F3
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-evidence-code-review-20261120.md
artifact_sha: 6b33da9ef079ec175022b1954d1027ba7807b102924e07528aa1921994d9179c
---
## Provenance
Non-author reviewer agent, code round 1 (Law 11). Reviewed the uncommitted working tree on top of 98b33a7 against docs/plans/mvp-evidence-portability-spec.md. Ran: `go vet` on the four packages (clean); `go test ./internal/connector/reviews/ ./internal/cli/` (ok); `go test ./internal/connector/intent/sdd/ -run 'TestPRCheckOnAFreshLedger|TestPRCheckEndToEnd'` (ok, 65s). One hand mutation (projection ref loop iterates `EvidenceRefs[:0]`, i.e. resolution skipped), reverted exactly from a saved copy; `git diff --stat` matches the original.

## Status
NEEDS_WORK: one MED. The parser, seal, back-compat and view logic read correct; the projection's ref resolution has no test that can fail.

## Findings
**F1 (MED) Projection ref resolution is unproven; the spec's calibration claim is false.** With the reference loop in `reduceObligationVerdict` skipped entirely, `TestPRCheckOnAFreshLedger` still passes: resolution failing open (no dangling check, no future check, no event id in the view) is invisible, because the e2e control only exercises the legacy-id dangling path and the positive case only asserts `PASS`. Spec section 3 R-INV-11 names projections `TestObligationVerdictProjectionValidation` new cases (resolved ref, no git event, git event after verdict) as the proving test; no projections test file was changed. The integration harness is broken at HEAD (known), but the dangling and future branches therefore have zero coverage anywhere. Fix: add the three cases (even if the harness cannot run them yet), and/or strengthen the e2e positive to assert the view's `evidence_event_ids` is non-empty and equals clone B's local git event id for x; add a control that deletes/omits the git event (e.g. check against a base that excludes the commit) so dangling-ref fails.

**F2 (LOW)** Parser rule is correct (absent/null pass via `!= nil`, `[]` rejected, one exact entry), but reviews tests were not mutation-checked by me beyond reading; the ref-mismatch branch in the projection (`ref != CommitEvidenceRef(...)`) is unreachable after parsing and untested. Recorded, not blocking.

**F3 (LOW)** Back-compat reads correct: `omitempty` keeps legacy payload bytes; legacy rows keep exact view bytes (`evidenceIDs = outcome.EvidenceEventIDs` before sort/compact, which sorts the original slice in place, harmless since it is the same sorted-or-not data already marshalled previously only if already sorted; legacy view bytes could differ if a legacy verdict listed ids unsorted, though the parser enforces ordering via the `previous` check). Seal output matches spec (refs only, `evidence_event_ids: []`).
