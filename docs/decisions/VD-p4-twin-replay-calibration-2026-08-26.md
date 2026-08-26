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
The next P4 slice must add an isolated PostgreSQL implementation and tests for cleanup,
production preservation, cancellation, and proof-bearing replay results before any
prediction-ledger surface is accepted.
