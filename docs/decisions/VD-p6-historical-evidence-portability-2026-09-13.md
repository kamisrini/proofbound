# VD-p6-historical-evidence-portability-2026-09-13: strict migration-only historical evidence

**Status:** Accepted

**Date:** 2026-09-13

## Authority and provenance

This decision applies the founder's declared-authority receipt
[`p6-historical-evidence-ratification.md`](../verification/verdicts/p6-historical-evidence-ratification.md)
to the P6 Task 3 reproducibility blocker. It supplements the accepted P6 semantic decision
[`VD-p6-consolidation-2026-09-12`](VD-p6-consolidation-2026-09-12.md), which cites the non-author
plan adjudication [`p6-consolidation-plan-round1.md`](../verification/verdicts/p6-consolidation-plan-round1.md).

The receipt records the founder's exact statement “ratify option 1”; it is not cryptographically
authenticated and does not itself constitute a verification result.

## Decision

Proofbound may carry the two exact historical evidence envelopes cited by the accepted P5 obligation
verdict, together with the three exact intent-record envelopes that are their minimal transitive
ledger-order prerequisites, in a versioned, committed migration archive. The archive is imported
only by an explicit fresh-ledger migration command, after strict validation of its fixed record
count, sequence numbers, event IDs, event fields, and content hashes. It is not read by any
connector and cannot broaden the event universe.

The import is append-only and refuses a non-empty ledger. Ordinary sync, projection, verification,
and report behavior remains unchanged; in particular, missing evidence remains a failed-closed
condition and the archive does not make arbitrary dangling references valid.

## Consequence

The Task 3 SPEC and tests define the archive schema and close the exact source-recoverability gap.
The implementation may add only the narrow migration seam and its acceptance record. No event kind,
wire identity, pinned vector, provider, or product capability changes.
