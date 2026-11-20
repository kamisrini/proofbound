---
schema: vera.verdict.v1
verdict_id: mvp-evidence-spec-review-round2-20261120
status: ACCEPTABLE
reviewed_commit: 98b33a7826c0d41044cec495e69c793a46186512
findings: []
artifact_path: docs/verification/verdicts/mvp-evidence-spec-review-round2-20261120.md
artifact_sha: d719b6f8c3b3fc536e3826ae6869403ee688bbfa1f45384834f26e856aec01cd
---
## Provenance
Non-author reviewer agent, spec round 2 of 2 (Law 11), on docs/plans/mvp-evidence-portability-spec.md (draft 2, untracked) at HEAD 98b33a7. Read-only: round-1 verdict, reviews/v2.go (ObligationOutcome, ParseObligationVerdict evidence loop), cli/starter.go (draft ~342, draft decoder ~383, seal ~460-476), projections.go (reduceObligationVerdict evidence loop ~525, view upsert), action.yml (version input), reviews/SPEC.md R-INV-10..13. Ran only grep/sed; no tests, no make check.

## Status
ACCEPTABLE. No new HIGH or MED defect in draft 2. All four round-1 MEDs are answered and match the code.

## Round 1 dispositions checked
- **F1 answered.** The nested outcome struct (v2.go:25) has no order check, and draft 2 drops the order claim. `omitempty` plus byte-identical re-marshal is stated, and `[]` is rejected explicitly. The only reader of `evidence_event_ids` outside parse, seal and projection is the view, so the field change is contained.
- **F2 answered.** The ref is pinned to `git:commit.recorded:<reviewed_commit>`. Legacy ids are stated as unrestricted and pre-existing, so no false self-grading implication remains.
- **F3 answered.** §2.5 names the failure mode (a v0.3.0 binary rejects the unknown field and sync fails closed). action.yml defaults `version` to latest, so the unpinned Action upgrades by itself. The release-notes requirement is stated.
- **F4 answered.** The success-path seal test is added for CLI-INV-20. R-INV-14 enumerates its rejection cases. CLI-INV-31 is in a non-integration-tagged file with a calibration control. Residual: the spec does not spell out the legacy-id negative control or the clone mechanics. These are test-design details, recorded as LOW and not fixed (Law 11).
- **F5, F6 taken.** The query is pinned and the view union is deduped. F7 needed no change.

## Findings
None at HIGH or MED.

LOW, recorded only: (a) the CLI-INV-31 test description could state that the clone copies the committed verdict, readiness gate and signature, and could add a legacy-event-id verdict negative control that yields `dangling`. (b) The spec does not say that the `SATISFIED` evidence check in parse must count refs, as opposed to the current `len(EvidenceEventIDs)==0` check. §2.1 says it implicitly.
