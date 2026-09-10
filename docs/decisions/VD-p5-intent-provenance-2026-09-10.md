# VD-p5-intent-provenance-2026-09-10: adopt the P5 Proofbound intent-provenance semantics

**Status:** Accepted
**Date:** 2026-09-10

## Authority and provenance

This decision adopts the ratified P5 plan
[`P5-PROOFBOUND-INTENT-PROVENANCE-v3.md`](../plans/P5-PROOFBOUND-INTENT-PROVENANCE-v3.md).
It cites the founder's declared-authority decisions in
[`p5-founder-ratification.md`](../verification/verdicts/p5-founder-ratification.md) and the independent
semantic review in
[`p5-adjudication-round1.md`](../verification/verdicts/p5-adjudication-round1.md). The exact exhibit
reviewed in that adjudication is preserved beside it and its bound digest was verified locally.
Machine-local receipt and gate evidence is attached in
[`p5-ratification-baseline.md`](../verification/p5-ratification-baseline.md).

The ratification is a recorded declared-authority claim, not authenticated approval. Committing this
decision makes it durable; only the later package acceptance evidence and a non-author verdict can
make the implementation accepted.

## Decision

P5 builds a provider-independent, ledger-native provenance chain:

```text
business-decision revision -> requirement revision and obligations
  -> change-intent revision -> exact commit claim
  -> independent per-obligation verdict -> observed deployment
```

The canonical record types are business decisions (`BD`), behavioral requirements (`BR`), change
intents (`CI`), commit claims, requirement reviews, and obligation verdicts. A committed record is an
authored claim observed by Proofbound, not automatically a fact. Repo-native owners, approvers,
sponsors, authors, and reviewers are declared identities. `approval.observed` remains reserved for a
future authenticated feed.

Every semantic relation is typed and binds the target's provider-scoped source, record kind, stable
record ID, and exact artifact SHA-256. The closed P5 relations are `authorized_by`, `implements`,
`modifies`, `repairs`, `retires`, `constrained_by`, `claims`, `evaluates`, and `supersedes` where the
plan permits it. Artifact SHA-256 binds immutable provider revision bytes; the separate canonical
payload digest binds RFC 8785-normalized ledger payload bytes. Neither comparison may substitute for
the other.

Requirement obligations have append-only stable IDs within a lineage. A changed meaning gets a new
ID; retirement leaves a tombstone. Authored obligation verdict outcomes are `SATISFIED`,
`NOT_SATISFIED`, or `INCONCLUSIVE`; absence is the derived, never-stored state `UNVERIFIED`.
Verdicts own outcomes, cite evidence events no later than their observation, and cannot launder
builder-authored evidence into satisfaction. Aggregate green fails closed on any missing, mismatched,
inconclusive, contradictory, or non-satisfied component.

Requirement review follows ratified semantic rule 8. A declared non-author reviews verifiability,
not business merit, with per-obligation outcomes `VERIFIABLE`, `AMBIGUOUS`, `UNTESTABLE`, or
`CONTRADICTORY`. Equal declared author and reviewer identities fail closed. For general users the
rule is a soft derived gate: absent or non-verifiable review caps the chain below green but blocks
nothing. The hard form applies only to Proofbound's own P5 self-hosting requirements.

The intent-provider contract is source-neutral. P5 implements exactly two real file-native
providers: `intent.records` and `intent.specdir`; a synthetic provider exists only in conformance
fixtures, and mutable API-backed providers wait for a real feed. Providers emit replay-sufficient
canonical payloads. Missing source segments remain explicit partial-chain gaps and are never
fabricated. An unqualified `Intent:` trailer resolves in its commit's own tree; a qualified trailer
uses its named provider's point-in-time rule and records the resolved digest.

Accepted native revisions are immutable append-only files that link predecessors. Projections are
disposable relational views rebuilt from ledger events. No graph or second writable truth is
introduced. Integrity gates begin in canary; the delivery-readiness gate graduates only at the
actual boundary Proofbound controls, `make delivery-enforce`.

## Compatibility and identity

The product's live identity becomes Proofbound in Task 0.5. The wire names `vera.witness.v1`,
`vera.verdict.v1`, and `vera.replay.v1`, their pinned byte-exact vectors, committed decisions,
verdicts, journals, and Git history remain frozen. Live CLI, module/import, environment, runtime-path,
and documentation surfaces rename, with `vera` CLI and `VERA_*` environment aliases retained for one
phase under a dated deprecation advisory. Old commit and verdict payloads retain their exact meaning;
they gain no retroactive intent or obligation satisfaction.

## Threat model and falsifiers

The design fails closed against wrong-revision joins, missing or future evidence, provider-prefix
spoofing, cross-provider ID collision and sync-order races, canonicalization instability, upstream
mutation between resolution and ingest, obligation removal/reuse, self-review, and projections that
cannot reproduce the ledger row set. The conformance suite must distinguish artifact and payload
digests independently.

Stop and re-ratify if the `specdir` mapper requires a canonical-schema change. Revisit or simplify
when any plan falsifier fires, including unrecoverable old meaning, subjective obligation verdicts,
unusable applicability boundaries, authoring cost exceeding evidence cost for three changes, or
rubber-stamp requirement reviews whose near-zero findings coexist with downstream inconclusive
verdicts.

## Consequences

P5 proceeds in the ratified order: identity migration, frozen provider/spec contracts and invariant-
derived tests, providers, commit binding, projections/reports, verdicts/reviews, deployment joins,
integrity gates, then self-hosted acceptance. Each coherent implementation slice is committed for
durability. No package is accepted until its calibrated mutation sweep is green and a non-author
independent verdict is committed on receipt.
