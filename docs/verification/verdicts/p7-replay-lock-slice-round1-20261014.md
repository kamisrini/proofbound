# P7 replay lock addendum and fixture planning review — round 1

**Reviewed commit:** `bf2c5ff387f6da7dcc1596ec4666685c5481881b`
**Reviewer identity:** fresh independent non-author context (`Fermat`)
**Reviewed artifacts:**

- `docs/intent/records/business-decisions/BD-replay-evidence-7a91c4/0001.md`
- `docs/intent/records/requirements/BR-replay-evidence-6e3d2b/0001.md`
- `docs/plans/P7-replay-lock-spec-addendum-draft1.md`
- `docs/plans/P7-replay-lock-fixture-matrix-draft1.md`
- `docs/plans/P7-replay-intent-requirement-proposal-draft1.md`

## VERDICT: AMEND

The exact P5 decision and requirement normalized self-digests reproduce, and the requirement's
`authorized_by` reference binds the exact decision revision. `O-1` is independently verifiable;
its separate P5 review is recorded in
[`p7-replay-requirement-review-round1-20261014.md`](p7-replay-requirement-review-round1-20261014.md).
The later P7 addendum and fixture plan preserve the accepted planning subject and exclusions, but
four defects block acceptance of this slice.

## Findings

1. **HIGH — Incomplete observed source binding.** The proposed `source_prefix.events` records
   event ID, sequence, and `content_sha`, while the reused twin `digestRecords` omits event times
   and connector version. Projection consumes event time. A change in an omitted field can leave
   the proposed source digest unchanged. Bind the complete observed event records without changing
   the frozen replay-proof semantics. See the addendum's `source_prefix` profile and
   `kernel/internal/twin/replay.go` `digestRecords`.
2. **MED — No whole-ledger preservation witness.** Fixtures select a prefix that excludes a later
   event, but do not define before/after evidence for the entire source ledger on success and
   failure. A matching prefix alone does not establish that recorded history stayed unchanged.
   See fixture cases L03–L06.
3. **MED — Evidence bytes are not fully pinned.** The projector manifest lacks an exact canonical
   byte encoding and complete build identity. The proof serialization is still explicitly open.
   Pin both and supply byte-exact compatibility fixtures before accepting an exact reproducibility
   profile. See the addendum's projector and proof evidence proposals.
4. **MED — Candidate payload integrity is unspecified.** The addendum calls for decoding a full
   event and running `core.Event.Validate`, but that method does not recompute `content_sha` from
   the payload. Require that independent comparison, or define precise trusted evidence for it.
   See the addendum's `candidates` profile and `kernel/internal/core/SPEC.md`.

## Disposition and limits

Amend the draft addendum and fixture matrix, then obtain another fresh non-author review of their
exact revised commit. The accepted round-3 P7 planning subject remains accepted; this verdict
does not overturn it. No fixture corpus, run result, independent `O-1` satisfaction verdict, or
implementation acceptance was established. Behavior-lock, regeneration, snapshot-provider, and
external-feed implementation remain gated.
