# VD-p6-consolidation-2026-09-12: consolidate and deeply complete P1–P5 before adding width

**Status:** Accepted
**Date:** 2026-09-12

## Authority and provenance

This decision adopts
[`P6-CONSOLIDATION-PLAN-draft2.md`](../plans/P6-CONSOLIDATION-PLAN-draft2.md), bound by
SHA-256 `43a0409da3bf627e2dac23ebad5039ad09640056ce66f6779bf54b900fd60de3` in the founder's
declared-authority
[`p6-founder-ratification.md`](../verification/verdicts/p6-founder-ratification.md). It also cites
the build-machine non-author
[`p6-consolidation-plan-round1.md`](../verification/verdicts/p6-consolidation-plan-round1.md), whose
NEEDS_WORK findings were folded into draft 2 before ratification.

The ratification is a recorded declared-authority claim, not authenticated approval. The plan
adjudication establishes plan adequacy after its findings were folded; it does not accept P6's
future implementation. Implementation acceptance still requires the final mutation sweep and a
non-author verdict on the frozen current code.

## Decision

P6 is a consolidation and deep-completion phase for the promises already made in P1–P5. It adds no
new product capability. A mechanically generated C1–C8 census is the closed work queue. Every live
promise must either be satisfied by cited mechanical evidence or be explicitly superseded,
narrowed, or retired by a decision that updates every live authority. A deferral label by itself
does not close debt.

The task order in the ratified plan is binding. Work proceeds spec-first from the census through
documentation and gate truth, connector reality, event-universe and reproducibility proof, intent
applicability, requirement reviews, P5 measurements, the conditional-provider decision, final
package acceptance, and a consolidation verdict. Commits provide durability, not acceptance.

## Intent-applicability rule

For P6 intent coverage and enforcement, a commit is behavior-changing exactly when its changed-path
set intersects:

```text
Makefile
check-windows.ps1
setup-windows.ps1
CLAUDE.md
.github/workflows/**
gates/**
scripts/**
tools/**
kernel/**
docs/gates.md
docs/allowed-skips.txt
docs/allowed-survivors.txt
docs/decisions/INDEX.md
docs/invariants.lock
docs/laws.lock
```

The only exclusions within included directory classes are `kernel/**/SPEC.md` and
`kernel/**/testdata/**`. Test Go files remain included. Other documentation, notes, vision files,
`README.md`, `ROADMAP.md`, `.gitignore`, and `LICENSE` are excluded unless exact-listed above.

Changed paths are content-independent. Root commits use all paths; ordinary commits diff their
first parent; merges use the union of diffs against every parent; renames and copies test old and
new paths; deletions test the deleted path. Any match is applicable. An applicable commit passes
only when the Git connector resolves at least one valid exact `Intent:` claim from that commit.

Historical coverage begins immediately after `f426ca8`. Any known false negative blocks
graduation. A false-positive rate greater than 10% across the complete post-anchor history fires
redesign before enforcement. The implementation may not silently widen, narrow, or reinterpret
this predicate.

## Scope and effort boundary

The conditional snapshot provider is skipped in P6: no real export feed is designated and the
no-width doctrine remains intact. Its pinned contract is preserved for separate P7+ authorization.
P6 must not add behavior locks, regeneration, a production-shaped twin, signing or portable
warranties, a governor, marketplace features, an end-user UI, new event kinds, or other vision-ladder
capability.

The stop-early ceiling is 40 active execution hours after the Task 0 census commit, excluding time
waiting for founder input or independent verification. If the plan's stop-early falsifier fires at
that ceiling, P6 records the remaining rows without claiming deep completion; P7 planning still
requires an explicit founder decision.

## Acceptance boundary

P6 closes only when the census has zero unclassified rows, zero open `close-in-P6` rows, and no live
promise hidden by defer/wontfix; every production package is mutation-green and non-author accepted
against one frozen final implementation commit; the event-route and clean-clone proofs pass; no
expired advisory remains; a round-C non-author verdict is ACCEPTABLE and committed on receipt; and
bare `make check` plus `proofbound verify` are green.

P7+ capability planning begins only after that closure. Frozen wire identities
`vera.witness.v1`, `vera.verdict.v1`, and `vera.replay.v1`, pinned vectors, and historical artifacts
remain unchanged.
