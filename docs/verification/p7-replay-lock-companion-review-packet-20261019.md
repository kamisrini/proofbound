# P7 replay lock companion boundary — documentary review packet

**Packet status:** REVIEW REQUEST — no verdict recorded

**Review type:** fresh non-author documentary planning review

**Requested outcome:** `ACCEPT`, `AMEND`, or `REJECT`

**Reviewed commit:** `9d22128aeef5bb5fe0ed79f36f69cecab382a158`

This packet asks an independent reviewer to assess the proposed companion API/SPEC boundary and
fixture plan for the P7 `ReplayIsolated` lock slice. It does not modify the live `twin` contract,
authorize implementation, create runtime fixtures, validate PostgreSQL, or reopen dormant
deferral `DEF-P7-001`.

## Immutable review corpus

| Role | Path | SHA-256 |
|---|---|---|
| proposed companion API/SPEC amendment | `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` | `198ea1ed3699a6468114d9dbb7613f0293a63276dcdc5e4edb70d65639dc7b57` |
| proposed companion fixture plan | `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` | `fc9c4742f1a9b1b12f2f19e3d4cae0642c0e6debd53f3f6f674c8e4ee1e91cad` |
| roadmap boundary | `ROADMAP.md` | `494190fc2baac7e6181960f35db94a9cfe0a7726c2e36f67103f3ad489372309` |
| accepted replay-lock addendum | `docs/plans/P7-replay-lock-spec-addendum-draft1.md` | `0ffc074b67170145fbd47a9fa034aa64ce5cbd4928bc53b83c7ecb639c89353a` |
| accepted external-prerequisites amendment | `docs/plans/P7-replay-lock-external-prerequisites-amendment-draft1.md` | `922e708551d80658fee045e38c6f1487e90bdb5aed1f0eecb39c6cc044811a0c` |
| accepted fixture matrix | `docs/plans/P7-replay-lock-fixture-matrix-draft1.md` | `93edb519bc5da643ae4b0cf365b879e2f8b0115ec66f1545b72cb6e10af4cef4` |

The review must bind its verdict to the full commit above and these exact bytes, not to a later
working tree or conversation summary. Any changed byte requires a new packet and fresh review.

## Review questions

1. Does the proposed companion boundary preserve the live P4 `ReplayIsolated` behavior and keep
   the new API explicitly separate until implementation is authorized?
2. Is the proposed verified-input shape precise enough to bind exact archive bytes, archive digest,
   complete extracted inventory, and the launched PostgreSQL executable without trusting a path,
   URL, ambient cache, or digest alone?
3. Are safe extraction, required `bin/pg_ctl`, server-executable binding, `BinariesDir` routing,
   no-fetch behavior, cleanup, cancellation, and complete source-ledger preservation specified
   strongly enough for an independent implementation review?
4. Are the unresolved API decisions clearly identified, especially inventory ownership, archive
   format/link policy, executable observation, and temporary-root ownership?
5. Does the fixture plan cover successful replay, malformed/unsafe archives, missing executables,
   cache-only fallback, digest substitution, projector failure after store open, cancellation,
   cleanup, network observation, setup-source binding, and result/proof byte comparison?
6. Are expected negative controls distinguished from actual failures, and are `PENDING`,
   `UNVERIFIED`, and `PASS` kept separate until the required evidence exists?
7. Does the packet preserve the accepted P7 addendum, external-prerequisites amendment, and
   readiness-manifest boundary without approving the pinned PostgreSQL input or changing L12/L17?
8. Are the non-goals and stop points sufficient to prevent documentary acceptance from being
   treated as live API adoption, runtime fixture acceptance, or implementation authorization?

## Acceptance bar

Return `ACCEPT` only if the proposed planning boundary and fixture coverage are coherent and
complete for a future implementation review, with no blocking, high, medium, or low findings.
Return `AMEND` with exact findings and required changes if any rule is ambiguous, incomplete, or
inconsistent. Return `REJECT` if this boundary should be abandoned.

The verdict must identify the reviewer as non-author, include the reviewed commit and this packet's
SHA-256, and state its scope. An `ACCEPT` accepts documentary planning only. It does not accept
runtime fixtures, L12/L17, the PostgreSQL provenance/licensing prerequisite, a live API/SPEC
change, a parser, behavior-lock implementation, regeneration, or any provider.

## Current gate state

- `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant under owner policy.
- No PostgreSQL input was downloaded, inspected, approved, or replaced for this packet.
- No parser, fixture corpus, runtime case, or P7 implementation was added.
- The full pinned-toolchain `make check` passed for the reviewed commit before packet creation;
  targeted link, index, deferral, artifact-integrity, and diff checks also passed.

## Verdict slot

No verdict yet. The authoring session must not fill this section with an acceptance decision.
