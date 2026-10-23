# P7 replay-lock companion API/SPEC boundary — review packet

**Review type:** independent non-author documentary API/SPEC review
**Status:** REVIEW REQUEST — no verdict recorded
**Review date:** 2026-10-23
**Scope:** resolve the five open API/SPEC questions while preserving all runtime and provenance
gates

## Review corpus

The reviewer must bind the verdict to the exact full commit containing this packet and the
following files. Record SHA-256 values before issuing the verdict; any changed byte requires a new
packet and fresh review.

| Item | Path |
|---|---|
| boundary revision | `docs/plans/P7-replay-lock-twin-api-amendment-draft2.md` |
| recommendation | `docs/decisions/VD-p7-replay-lock-companion-api-boundary-2026-10-23.md` |
| accepted planning baseline | `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` |
| accepted fixture baseline | `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` |
| live package contract | `kernel/internal/twin/SPEC.md` |
| accepted implementation boundary | `docs/verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md` |
| current state/roadmap | `notes/state.md`, `ROADMAP.md` |

## Questions for the reviewer

1. Does draft 2 resolve inventory ownership by selecting one in-memory `Inventory` value without
   creating a second runtime truth?
2. Is the exact `.txz` xz/tar format and regular-file-only, no-link extraction policy precise
   enough to fail closed?
3. Does it correctly distinguish canonical staged paths from the not-yet-proven observation of
   the executable actually launched by the daemonized store layer?
4. Does the temporary-root contract assign ownership and cleanup across cancellation, partial
   extraction, store-open failure, projector failure, success, and cleanup failure?
5. Are the proposed names and future live SPEC wording stable enough for a separately authorized
   implementation/API adoption?
6. Does the packet preserve the existing gates: no runtime fixtures, no binary validation, no
   L12/L17, no behavior-lock acceptance, no PostgreSQL approval, and dormant `DEF-P7-001`?

## Acceptance bar

Return `ACCEPT` only if all six questions are answered without weakening exact-byte binding,
safe extraction, offline provisioning, source preservation, or cleanup closure. Return `AMEND`
with exact findings and required edits if any contract wording is incomplete. Return `REJECT` if
the boundary should not proceed.

An `ACCEPT` applies only to this documentary boundary. It does not adopt the live SPEC, authorize
runtime fixtures, validate PostgreSQL, close L12/L17, accept a behavior lock, or establish an
independent runtime verdict. The reviewer must be non-author and must identify the reviewed full
commit and this packet's SHA-256.

## Evidence boundaries retained

- `DEF-P7-002` is already resolved under its owner-approved source-corpus scope and is not reopened.
- `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant; the pinned Linux/amd64 PostgreSQL 16.4.0
  Maven JAR is not approved.
- The preflight/staging seam and synthetic archive controls remain accepted only within their
  recorded scope.
- The complete projector/build manifest, source/toolchain binding, runtime fixture corpus,
  no-fetch and cleanup witnesses, L12/L17 result, behavior-lock acceptance, and independent
  runtime verdict remain absent.

## Verdict slot

No verdict yet. The authoring session must not fill this section with an acceptance decision.
