# VD-p6-private-storage-compat-2026-09-12: freeze migrated private storage identity, retire external aliases

**Status:** Accepted
**Date:** 2026-09-12

## Authority

This decision applies the founder's declared-authority
[`p6-legacy-storage-ratification.md`](../verification/verdicts/p6-legacy-storage-ratification.md) to
the P6 consolidation authority in
[`VD-p6-consolidation-2026-09-12.md`](VD-p6-consolidation-2026-09-12.md). It resolves the actual
machine-state contradiction recorded at `2d35faa`: the only local ledger is a migrated embedded
cluster carrying the private `vera-v1` identity.

## Decision

An existing `.proofbound` embedded-PostgreSQL cluster marked `vera-v1` remains readable
indefinitely. Its internal database, role, password, and marker are frozen migration compatibility,
not callable product identity. New stores continue to initialize only with the Proofbound private
identity. No automatic rewrite, merge, deletion, or in-place role/database migration is permitted.

P6 removes every external legacy product alias: the `vera` command wrapper, `VERA_*` environment
fallbacks, their advisories, and public compatibility documentation. Frozen wire identities
`vera.witness.v1`, `vera.verdict.v1`, and `vera.replay.v1`, pinned vectors, Git history, verdicts,
decisions, and journal evidence remain unchanged.

## Consequences

The identity inventory classifies the private store identity as frozen compatibility and requires
zero live external aliases. Tests must discriminate a new Proofbound store from a migrated legacy
store and prove that public legacy command/environment routes no longer work. This decision
supersedes only the removal sentence for private embedded storage in
`docs/proofbound-identity-migration.md`; the P5 one-phase removal rule remains binding for external
surfaces.
