# VD-p7-readiness-manifest-2026-10-17: case/evidence manifest shape

**Status:** ACCEPTED FOR DOCUMENTARY PLANNING — round-1 non-author review on 2026-10-19;
implementation separately gated

**Date:** 2026-10-17

## Context

The P7 planning artifacts require a case manifest that binds one bounded `ReplayIsolated` case to
its reviewed contracts, inputs, commands, witnesses, outcomes, artifacts, and independent review.
The readiness-manifest draft proposed the needed fields but left the exact status and pending-review
representations open.

This is a planning decision only. It does not authorize a parser, runtime fixture, binary
validation, live API change, behavior-lock implementation, regeneration, or any provider.

## Decision

Use a UTF-8 RFC 8785 canonical JSON case manifest with this exact top-level member set:

`schema`, `case_id`, `subject`, `case_status`, `evidence_revision`, `contract_refs`, `inputs`,
`commands`, `witnesses`, `outcomes`, `artifacts`, and `review`.

The following rules are part of the planning shape:

1. No unknown, missing, or `null` members are permitted in an emitted manifest. A pending value is
   represented by an object with an explicit `status`, never by `null` or an omitted required field.
2. `case_status` is derived, not trusted: `FAIL` if any applicable outcome is `FAIL`; `UNVERIFIED`
   if required evidence, review, or an applicable outcome is pending/unverified; otherwise `PASS`.
   A PostgreSQL-dependent case cannot reach `PASS` while `DEF-P7-001` remains dormant.
3. Every content reference uses a repository-relative `path` and lowercase SHA-256 `sha256`.
   Evidence files additionally record positive `size_bytes`. The path, bytes, and digest must bind
   to `evidence_revision`; a digest alone is insufficient.
4. `contract_refs` binds the exact lock addendum, fixture matrix, and authorized implementation
   scope. `inputs` binds the lock, source, candidate, setup-source, and projector-manifest
   evidence. These references identify evidence; they do not themselves establish acceptance.
5. Every command records exact argv, repository-relative working directory, declared network
   policy, environment digest, exit code, and stdout/stderr evidence references.
6. Witnesses are typed evidence references. The minimum planning witness roles are complete
   ledger-before and ledger-after bytes, network observation, and cleanup observation.
7. Each outcome is an object with `status` and evidence references. Allowed statuses are
   `PASS`, `FAIL`, `UNVERIFIED`, `UNDECLARED`, and `PENDING`; `UNDECLARED` is only valid for
   the intent-binding outcome, and `PENDING` is reserved for work not yet run or reviewed.
8. `review` always contains `author_commit` and an `independent_verdict` object. The latter has
   `status=PENDING` until a non-author verdict exists; an accepted verdict must include its
   content-addressed path and digest.
9. Arrays have deterministic order: contract names and outcome names are fixed by the schema;
   commands follow execution order; witnesses and artifacts follow declared role order.
10. The manifest is an evidence index, not the behavior lock, frozen replay proof, setup-source
    manifest, ledger event, wire identity, or writable truth.

## Consequences

- The shape can be reviewed and prepared without activating the deferred PostgreSQL provenance or
  licensing check.
- A future implementation must not emit a `PASS` by copying an authored status or by relying on a
  hash without retrievable bytes.
- Missing fixtures, missing witnesses, dormant prerequisites, and absent independent review remain
  `UNVERIFIED`/`PENDING`; they are not silently converted to failures or passes.
- The exact shape required fresh non-author review before a parser, runtime corpus, or P7
  implementation could be accepted.

## Fresh non-author review outcome — round 1 — 2026-10-19

The exact planning corpus at commit `334b33822cfe2cf35071ee936f260ed079b8a6c1` was reviewed by
Srini K as an independent non-author reviewer. The verdict is recorded in
[`p7-readiness-manifest-round1-20261019.md`](../verification/verdicts/p7-readiness-manifest-round1-20261019.md).

**Outcome: ACCEPT for documentary planning only.** No findings were reported. This does not
authorize a parser, runtime fixture, PostgreSQL validation, live API/SPEC change, behavior-lock
implementation, regeneration, provider, or external feed.

## Authority

The implementation planning companion is
[`P7-readiness-manifest-plan-draft1.md`](../plans/P7-readiness-manifest-plan-draft1.md). The
ReplayIsolated lock contract, fixture matrix, and external-prerequisites amendment remain the
authorities for their respective domains. This decision does not alter `DEF-P7-001`.
