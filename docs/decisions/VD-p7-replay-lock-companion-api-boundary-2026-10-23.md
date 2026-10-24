# VD-p7-replay-lock-companion-api-boundary-2026-10-23

**Status:** OWNER-ACCEPTED DOCUMENTARY RECOMMENDATION; not an accepted live API decision.

## Recommendation

Adopt the companion boundary described in
[`P7-replay-lock-twin-api-amendment-draft2.md`](../plans/P7-replay-lock-twin-api-amendment-draft2.md)
for the next independent API/SPEC review:

- keep one complete in-memory `Inventory` in `VerifiedPostgresInput`;
- consume exactly one fully consumed xz-compressed POSIX `ustar` `.txz` stream, with canonical
  slash-path and regular-file-only extraction; reject path-only, URL, cache, and network inputs;
- allow only regular files with modes `100644`/`100755`, with no links or unsupported tar types;
- use replay-owned temporary storage and `BinariesDir`, with unconditional close-before-remove
  cleanup and explicit combined-error handling;
- retain the proposed names `VerifiedArchiveEntry`, `VerifiedPostgresInput`, and
  `ReplayIsolatedWithVerifiedPostgres`;
- distinguish preflight hashing of staged `bin/postgres` from future observation of the actually
  launched daemon executable.

This is the narrowest contract that answers the five open review questions without creating a
second inventory truth, weakening extraction safety, or overstating the current implementation.

## Boundaries

This recommendation does not edit the live `kernel/internal/twin/SPEC.md`, authorize runtime
fixtures, validate or approve PostgreSQL, perform binary validation, claim L12/L17, accept the
behavior lock, or reopen `DEF-P7-001`. The owner-approved and independently reviewed Go module
corpus under resolved `DEF-P7-002` is unrelated to this API boundary.

The current implementation may be reviewed only for its independently accepted preflight/staging
seam. Process-level executable observation, no-fetch/cleanup witnesses, runtime fixture results,
and independent runtime verdict remain future gates.

The current live SPEC's `TWIN-INV-14` launched-executable wording is future runtime wording for
this boundary, not runtime acceptance. Any correction or adoption of that wording requires a
separately authorized SPEC change and fresh review.

## Decision record

The owner acceptance is recorded in
[`p7-replay-lock-companion-api-owner-acceptance-20261023.md`](../verification/verdicts/p7-replay-lock-companion-api-owner-acceptance-20261023.md).
The independent non-author review packet remains open:
[`p7-replay-lock-companion-api-review-packet-20261023.md`](../verification/p7-replay-lock-companion-api-review-packet-20261023.md).
Until that independent review records `ACCEPT`, this document must not be cited as an accepted live
contract or implementation authorization.
