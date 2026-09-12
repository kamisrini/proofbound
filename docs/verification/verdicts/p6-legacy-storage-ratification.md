# Founder ratification — P6 private legacy-storage identity

**Date:** 2026-09-12  
**Founder statement on receipt:** “go with recommendation”  
**Applies to:** the P6 Task 1 contradiction recorded in `notes/state.md` at commit `2d35faa`.

This is a **declared-authority** record. It records the founder's choice in the execution session;
it does not claim cryptographically authenticated approval.

## Ratified decision

Retain the private `vera-v1` embedded-PostgreSQL identity indefinitely for an already-migrated
`.proofbound` database. Classify it as frozen migration compatibility, not a live product alias.
New databases continue to use Proofbound-only private identities.

Remove the externally callable `vera` executable and all shipped `VERA_*` environment fallbacks in
P6. Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, committed
history, and documentary evidence byte-for-byte.

## Reason

The only local ledger on this machine is an existing migrated cluster marked `vera-v1`. Removing
the private database/user compatibility would make that ledger inaccessible; an in-place
PostgreSQL role/database migration adds avoidable data risk and does not improve the public product
identity. The private identity is therefore frozen at the storage boundary while external aliases
complete their already-ratified one-phase removal.
