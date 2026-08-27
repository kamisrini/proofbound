# VD-p4-twin-replay-calibration

**Status:** Accepted (2026-08-26)
**Date:** 2026-08-26

## Decision

P4 begins with a bounded, non-mutating replay contract in `kernel/internal/twin`.
Replay selects a ledger prefix, validates and appends explicitly sequenced candidate
events, and compares the resulting projection snapshot with a supplied baseline. The
projection function is injected for this spike; a later slice may provide an isolated
PostgreSQL fork adapter behind the same boundary.

The first slice does not execute arbitrary code, write the production ledger, persist
verdicts, or score predictions. Prediction events and calibration follow only after
ephemeral database isolation and proof binding have their own acceptance evidence.

## Rationale

`projections.Projector.Rebuild` operates on live derived tables and is therefore not a
twin. A transaction-scoped or filesystem-copy shortcut would risk production mutation.
The injected projection seam proves ordering, prefix bounds, deterministic comparison,
candidate validation, and fail-closed projector errors without claiming database-fork
isolation prematurely.

## Acceptance boundary

The package contract and proving tests are in [kernel/internal/twin/SPEC.md](../../kernel/internal/twin/SPEC.md).
The isolated implementation, deterministic replay proof, and pure calibration boundary are now
present behind the package boundary. They are not accepted as a production capability until
independent review verifies cleanup on every failure path, production preservation, cancellation,
multi-event sequence gaps, proof binding, and calibration validation. Durable prediction events
remain deferred until a real prediction feed has its own decision and acceptance evidence.
