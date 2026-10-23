# P7 replay-lock companion API/SPEC boundary — review packet

**Review type:** independent non-author documentary API/SPEC review
**Status:** REVIEW REQUEST — no verdict recorded
**Review date:** 2026-10-23
**Scope:** resolve the five open API/SPEC questions while preserving all runtime and provenance
gates
**Documentation freeze commit:** `b2982d0c8b35cdb7cb71e70bab1989c2d07c9b8a`

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

### Corpus digests at the documentation freeze

| Path | SHA-256 |
|---|---|
| `docs/plans/P7-replay-lock-twin-api-amendment-draft2.md` | `88ff99e1a3fd53e6a3553b5bad9d170eed9b5a969c2584c876196a38ffead2c0` |
| `docs/decisions/VD-p7-replay-lock-companion-api-boundary-2026-10-23.md` | `2a6f39d92048dbe4af498d2c065cb1b770544fbd17f9ab01bb0a60f76e15ebdc` |
| `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` | `c562eb6690d378a358a368d78135d471b6c5c139324793ed0dd63907f2946ed2` |
| `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` | `093179a9a69e90d90292607ad70c50eb8a671658cb21ea38b77bbc0df69f4b27` |
| `kernel/internal/twin/SPEC.md` | `fec7f8947807e03139fc0ddb79f5fbe2c88f1cd90da207446038727254437424` |
| `docs/verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md` | `86cc4c6b8b7471ea3a4f3be2968c4b510e08a103243d3e8284eed5ae15ac5504` |
| `ROADMAP.md` | `1dbe75e836c615872e462421e687b46baadde71f4b4fca52cd22b243e7a5d201` |
| `notes/state.md` | `0eaa109d541752ab885d5397d49fa2e6488be3bdf5e4d8c542845148b260ab2e` |

The packet's own SHA-256 must be recorded by the independent reviewer against the final packet
commit; it is intentionally not self-embedded.

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
