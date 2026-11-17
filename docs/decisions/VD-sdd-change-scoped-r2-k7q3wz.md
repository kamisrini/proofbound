# VD-sdd-change-scoped-r2-k7q3wz: Release 2 warrants changes, read history-natively; the main-spec lifecycle is deferred

**Status:** Accepted
**Date:** 2026-11-16

**Context:** Release 2's SDD Warranty Connector (`VD-release-line-r2-sdd-t5j34p`) was first designed
to map an SDD kit's main specs as living requirement records and to follow their lifecycle across
OpenSpec's archive step. Two non-author review rounds rejected both designs
(`docs/verification/verdicts/r2-m0-review-round1-20261116.md`, `…round2-20261116.md`). Round 2 ran
the second design against the real history of the OpenSpec repository and found it fails there: the
archived-delta fold disagrees with the main specs (77 of 358 final requirement states differ), the
old text a tombstone needs is absent from the tree (29 of 113 cases), changes opened and archived
between syncs never reach the ledger, and a warranty cannot be carried to the main spec under the
existing verdict rules. The shared root cause: an SDD kit's lifecycle lives in git history, and both
designs reconstructed it from one snapshot. The forcing question: what is the smallest honest R2
claim the evidence supports?

**Decision:** (owner, 2026-11-16) **Release 2 warrants changes, not main specs.** An OpenSpec change
(a Spec Kit feature) is warranted against the requirements it itself declares in its delta files,
together with the commits that implement it. The provider reads the repository **history-natively**
— every commit reachable from HEAD that touches SDD paths — so no change is missed, including those
that arrive already archived in one squash-merged commit (16 of 108 in OpenSpec's history).
Requirement records are content-addressed (one revision each), which keeps intent SPEC §5 without
tombstones. **Main specs are not ingested in R2**; their lifecycle (folding archived deltas,
tombstones, carrying a warranty across archive) is deferred to a later release. The contract is
`docs/plans/R2-sdd-provider-spec-draft3.md`; its §0 evidence was measured on the real history before
it was written.

**Consequences:** Makes easy — every R2 claim is one the evidence supports: "this commit was verified
against the requirements this change declared"; verdicts use the existing `obligation-verdict.v2`
rules unchanged; Spec Kit's in-place edits stop conflicting with §5. Makes hard / gave up — R2 says
nothing about whether a main spec as a whole is satisfied, and no warranty follows a requirement
after its change is archived; a reader wanting "is the system's current spec warranted" must wait for
the main-spec lifecycle release. History-native ingestion costs a walk over SDD-touching commits on
each sync.

**Revisit when:** the main-spec lifecycle release is scoped (it reopens this deferral); OR an SDD kit
records its own requirement history in-tree in a form a provider can read without reconstruction; OR
history-native ingestion proves too slow on a real tenant repository.
