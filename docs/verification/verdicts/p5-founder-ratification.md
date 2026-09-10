# Founder ratification — P5 Intent Provenance (Proofbound)

**Date:** 2026-09-09 · **Recorded by:** the independent verifier session that produced round-1 adjudication `p5-adjudication-round1.md` · **Applies to:** the P5 plan `P5-PROOFBOUND-INTENT-PROVENANCE-v3.md` · **Standing goals unchanged:** the product chain and goal section of the plan govern; this record authorizes scope and tradeoffs only.

Under the plan's own semantic rule 1, this is a **declared-authority** record: it observes that the founder stated these decisions in the planning session; it does not claim authenticated approval. Once P5 lands, a decision of exactly this kind is what the system itself will record — this file is the bootstrap instance.

## Ratified decisions

1. **Intent providers:** exactly two real providers in P5 — the repo-native records provider and the `specdir` file-native foreign mapper. API-backed providers against mutable upstreams are contract-pinned and deferred to a real feed. A synthetic test-double provider exists in fixtures only, never configurable in a release build.
2. **Identity:** the product renames to Proofbound now, as Task 0.5, under the plan's frozen/live split — `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1` and all pinned vectors frozen forever; history verbatim; everything live renames while zero external consumers exist.
3. **Adoption posture:** partial intent chains are first-class — a source that authors only part of the chain yields honest gaps (`authorization: undeclared`); records are never fabricated to complete a chain.
4. **Requirement review, soft-gate form (semantic rule 8):** a non-author verifiability review per requirement revision — closed per-obligation outcomes, verifiability only, never merit; absence or a non-verifiable outcome caps the chain below green and blocks nothing; provider-independent; the hard form applies only to Proofbound's own P5 requirements; rubber-stamping is falsifier 7.

## Rationale anchor

The product's premise is that intent is the scarce human input in a world where implementation is cheap; a warranty chain whose first link is unexamined warrants the wrong layer. The no-self-grading law has no carve-out for requirements, and the soft form preserves the adoption posture: gaps are visible, nothing is blocked, and green means every link — including the spec itself — was independently examined.

## What the semantic VD must cite

This record and the round-1 adjudication, both committed on receipt before the VD is minted.
