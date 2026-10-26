# P7 Proposed Next Non-Runtime Slice — Projector/Build Manifest Preparation

**Status:** DRAFT FOR OWNER AUTHORIZATION — NOT AUTHORIZED FOR IMPLEMENTATION

**Purpose:** select and bound the next eligible P7 slice after the completed
`P7-SPEC-ADOPT-1` slice. This proposal addresses the missing projector/build-manifest and
source/toolchain-binding evidence without opening the deferred runtime or binary gates.

## Proposed boundary

Prepare the documentary and validation contract for a canonical projector/build manifest and its
source/toolchain binding. The future authorized slice may define or implement manifest-shape
validation and evidence binding for exact source revisions, Go/toolchain identity, module inputs,
build settings, and projector revision. It must preserve the existing projector and twin behavior;
it must not add a provider, event kind, writable truth, or runtime behavior.

This proposal is limited to preparation and non-runtime evidence. It does not establish that the
manifest is complete, that a build is reproducible, or that any runtime input is approved.

## Deliverables if separately authorized

1. A canonical manifest contract bound to an exact evidence revision, with explicit fields for:
   source revision, projector revision, Go/toolchain identity, selected module bytes, build flags,
   platform, and relevant environment settings.
2. A source/toolchain binding plan that identifies each required input, its content identity, its
   retrieval/provenance source, and the evidence needed to establish completeness.
3. A non-runtime case/evidence index or manifest skeleton whose incomplete outcomes remain
   `UNVERIFIED` or `PENDING`, never `PASS` by absence.
4. Deterministic validation tests for schema shape, exact digest binding, missing/extra input
   rejection, and stale-revision rejection, using synthetic or already accepted repository bytes
   only.
5. A review packet recording exact scope, exclusions, commands, changed files, and the evidence
   required for a later independent non-author verdict.

## Explicit exclusions

- No runtime fixture corpus or launched-process observation.
- No PostgreSQL execution, PostgreSQL binary validation, or approval of the pinned PostgreSQL
  16.4.0 Maven JAR.
- No activation, resolution, or status change for `DEF-P7-001`.
- No L12/L17 result, behavior-lock acceptance, clean-clone proof, or independent runtime verdict.
- No network fetch, external artifact retrieval, binary build, or executable validation in this
  preparation slice.
- No case-manifest parser, lock evaluator, regeneration implementation, provider, feed, schema
  event, or new writable projection unless a later authorization explicitly adds it.

## Acceptance evidence required for this slice

- Exact proposal/authorization revision and native self-digest, verified with the repository’s
  zero-field normalization rule before staging.
- A machine-readable field and invariant inventory with tests proving shape and exact-revision
  binding only.
- A mapping from each proposed manifest field to its source document or declared `UNVERIFIED`
  state; no field may be inferred from filenames, mutable names, or ambient caches.
- Independent non-author review of the exact implementation commit and review packet.
- `scripts/deferrals.sh --check`, `scripts/p6-artifact-integrity.sh --check` when documentary
  verdicts change, `git diff --check`, and the pinned unrestricted `make check`.

## Authorization gate

This document is a proposal only. No implementation, external input retrieval, binary operation,
runtime fixture, or acceptance claim may begin until the owner records a separate authorization
that names this exact scope and preserves all exclusions above. If the owner wants a different
slice, amend this proposal and obtain fresh review before implementation.
