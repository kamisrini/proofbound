# P7 replay evidence requirement — proposal for owner approval

**Status:** HISTORICAL PROPOSAL, owner-approved for formal P5 records on 2026-10-14.
The authoritative records are
[BD-replay-evidence-7a91c4](../intent/records/business-decisions/BD-replay-evidence-7a91c4/0001.md)
and [BR-replay-evidence-6e3d2b](../intent/records/requirements/BR-replay-evidence-6e3d2b/0001.md).
This planning file is not itself a P5 record or satisfaction claim.

## Proposed business outcome

An operator may rely on a counterfactual replay result only when an independent reviewer can
reproduce its claimed behavior from the exact recorded source, authored candidate inputs, accepted
projector revision, baseline, and replay proof. Producing the result must leave the recorded source
history unchanged.

The first application is the accepted `ReplayIsolated` behavior-lock subject in
[VD-p7-behavior-lock-subject-2026-10-06](../decisions/VD-p7-behavior-lock-subject-2026-10-06.md).
It does not establish a claim about production traffic, an external snapshot provider, or
regeneration.

## Proposed requirement

**One obligation, `O-1`:** A `ReplayIsolated` result presented as a verified behavior-lock claim
must be independently reproducible from the exact observed ledger prefix, ordered authored
candidate events, accepted projector revision, baseline snapshot, candidate snapshot, complete
sequence map, and replay proof; the source ledger must remain unchanged. Missing, stale, or
conflicting evidence must prevent that verified claim.

This is one outcome: a replay claim is trustworthy only if its evidence can be reproduced without
altering the recorded history. The requirement is independent of the chosen lock-file layout and
does not assert that current P4 code already satisfies the new provenance or committed-proof
conditions.

## Proposed acceptance examples

| Example | Expected conclusion |
|---|---|
| Exact committed source, candidate, projector, baseline, result, and proof evidence reproduce; source ledger is byte-for-byte unchanged | An independent reviewer may evaluate `O-1` for the exact requirement revision and reviewed commit. A green claim still needs that reviewer's positive verdict. |
| Baseline provenance is absent or points to another prefix/projector revision | `UNVERIFIED`; no verified behavior-lock claim. |
| Proof bytes are missing, or any proof digest disagrees with the lock | `UNVERIFIED`; no verified behavior-lock claim. |
| A candidate is relabeled as an observed production event, or the source ledger changes | Reject the claim; the obligation is not satisfied. |
| The requirement revision or independent review evidence is absent | Derived `UNVERIFIED`; do not infer satisfaction from a passing replay test. |

The [P7 fixture matrix](P7-replay-lock-fixture-matrix-draft1.md) holds the detailed cases. Its
`UNDECLARED` case remains a negative control even after a real requirement is authorized.

## P5 record sequence after approval

1. Record an exact P5 business decision authorizing this outcome, with a named declared owner and
   approver. The repository-native provider records a declaration; it does not authenticate
   approval by itself.
2. After owner approval, record one active P5 requirement revision with `O-1` and an exact
   `authorized_by` reference to that business decision's committed revision digest. If a
   `proposed` revision is recorded first, activation creates a new immutable revision and digest;
   a review of the earlier proposal cannot be reused for the active revision.
3. Obtain a fresh non-author requirement review binding the exact active revision and `O-1`.
   Later, a change intent and independent obligation verdict must bind that exact revision,
   reviewed commit, and evidence before anyone reports `SATISFIED`.
4. Bind a real lock to the exact provider-scoped tuple only after these records exist:
   `record_kind=requirement`, registered `source=intent.<provider>`, `record_id`,
   `artifact_sha256`, and `obligation_id=O-1`.

The proposed owner, approver, record IDs, exact artifact digests, and reviewer are deliberately
unset. They must come from the actual authorization and committed evidence; they cannot be inferred
from this planning document.

## Owner decision outcome

The owner approved the proposed business outcome and `O-1` wording, then selected the existing
declared roles: `proofbound-founder` for business decision owner and approver, and
`proofbound-maintainer` for requirement owner. The formal P5 records above carry the exact
declarations and revision binding. Their independent review and any satisfaction verdict remain
pending; this approval does not authorize P7 implementation.
