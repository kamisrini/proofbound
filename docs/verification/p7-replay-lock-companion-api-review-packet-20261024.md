# P7 replay-lock companion API/SPEC boundary — follow-up review packet

**Review type:** independent non-author documentary API/SPEC follow-up review
**Status:** REVIEW REQUEST — no verdict recorded
**Review date:** 2026-10-24
**Scope:** resolve the findings from the round-1 documentary review while preserving all runtime
and provenance gates

The round-1 verdict was `AMEND` because the archive grammar, cleanup/error semantics, and the
staged-versus-launched wording in the live SPEC needed explicit documentary treatment. This packet
binds the amended boundary to a new exact corpus. It does not edit the live SPEC.

## Review corpus

The reviewer must bind the verdict to the exact full commit containing this packet and the following
files. Record SHA-256 values before issuing the verdict; any changed byte requires a new packet and
fresh review. The packet's own SHA-256 must be recorded in the verdict against that final commit;
it is intentionally not self-embedded.

| Item | Path |
|---|---|
| amended boundary revision | `docs/plans/P7-replay-lock-twin-api-amendment-draft2.md` |
| recommendation | `docs/decisions/VD-p7-replay-lock-companion-api-boundary-2026-10-23.md` |
| accepted planning baseline | `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` |
| accepted fixture baseline | `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` |
| live package contract | `kernel/internal/twin/SPEC.md` |
| accepted implementation boundary | `docs/verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md` |
| round-1 documentary verdict | `docs/verification/verdicts/p7-replay-lock-companion-api-round1-20261024.md` |
| current state/roadmap | `notes/state.md`, `ROADMAP.md` |

### Corpus digests

These values are the exact SHA-256 values of the amended corpus in the review commit.

| Path | SHA-256 |
|---|---|
| `docs/plans/P7-replay-lock-twin-api-amendment-draft2.md` | `abf8c043cdc1ab9c160c2b76171826742465636a94ad3b19cdfe13f43a73c749` |
| `docs/decisions/VD-p7-replay-lock-companion-api-boundary-2026-10-23.md` | `bb2720fdbb0f6140249bfd60dc86577013a0ddfad2e7186d03cefb931515c366` |
| `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` | `c562eb6690d378a358a368d78135d471b6c5c139324793ed0dd63907f2946ed2` |
| `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` | `093179a9a69e90d90292607ad70c50eb8a671658cb21ea38b77bbc0df69f4b27` |
| `kernel/internal/twin/SPEC.md` | `fec7f8947807e03139fc0ddb79f5fbe2c88f1cd90da207446038727254437424` |
| `docs/verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md` | `86cc4c6b8b7471ea3a4f3be2968c4b510e08a103243d3e8284eed5ae15ac5504` |
| `docs/verification/verdicts/p7-replay-lock-companion-api-round1-20261024.md` | `9b96f8658b3299182003faba6fd89354b8906f26b012549df3342080a2df22c5` |
| `ROADMAP.md` | `584380805fe5f4bd20e0a3157d4fda532e51cd7aeec3d37c7b68fe4bbaade2bc` |
| `notes/state.md` | `9925b6939ad28130ee6581c486e6e8d68662045b5638824850f6c1cb4a6d5882` |

## Questions for the reviewer

1. Does the amended boundary still select one complete in-memory `Inventory` value without creating
   a second runtime truth?
2. Does the exact one-stream xz plus POSIX `ustar` contract, including byte consumption, header
   restrictions, canonical path grammar, duplicate/collision rules, and regular-file-only policy,
   fail closed without weakening the archive boundary?
3. Does the boundary clearly distinguish staged `bin/postgres` hashing from the not-yet-proven
   observation of the executable launched by the daemonized store layer, including the explicit
   treatment of the current live SPEC's `TWIN-INV-14` wording?
4. Does the temporary-root contract cover ownership, close-before-remove ordering, cancellation,
   close failure, cleanup failure, and combined primary/cleanup error semantics?
5. Are the proposed names and future live-SPEC wording stable enough for separately authorized
   implementation/API adoption after these clarifications?
6. Does the packet preserve the existing gates: no runtime fixtures, no binary validation, no
   L12/L17, no behavior-lock acceptance, no PostgreSQL approval, and dormant `DEF-P7-001`?

## Acceptance bar

Return `ACCEPT` only if all six questions are answered without weakening exact-byte binding, safe
extraction, offline provisioning, source preservation, or cleanup closure. Return `AMEND` with
exact findings and required edits if any contract wording remains incomplete. Return `REJECT` if the
boundary should not proceed.

An `ACCEPT` applies only to this documentary boundary. It does not adopt or edit the live SPEC,
authorize runtime fixtures, validate PostgreSQL, close L12/L17, accept a behavior lock, or establish
an independent runtime verdict. The reviewer must be non-author and must identify the reviewed full
commit and this packet's SHA-256.

## Evidence boundaries retained

- `DEF-P7-002` is already resolved under its owner-approved source-corpus scope and is not reopened.
- `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant; the pinned Linux/amd64 PostgreSQL 16.4.0
  Maven JAR is not approved.
- The preflight/staging seam and synthetic archive controls remain accepted only within their
  recorded scope.
- The complete projector/build manifest, source/toolchain binding, runtime fixture corpus,
  no-fetch and cleanup witnesses, L12/L17 result, behavior-lock acceptance, clean-clone proof, and
  independent runtime verdict remain absent.

## Verdict slot

No verdict yet. The authoring session must not fill this section with an acceptance decision.
