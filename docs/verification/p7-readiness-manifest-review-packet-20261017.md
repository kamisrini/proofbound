# P7 readiness-manifest planning review packet

**Packet status:** REVIEW REQUEST — no verdict recorded

**Review type:** fresh non-author documentary planning review

**Requested outcome:** `ACCEPT`, `AMEND`, or `REJECT`

**Reviewed commit:** `334b33822cfe2cf35071ee936f260ed079b8a6c1`

This packet asks an independent reviewer to evaluate the owner-approved planning shape for the
future P7 case/evidence manifest. It does not authorize a parser, runtime fixture, PostgreSQL
validation, behavior-lock implementation, regeneration, or any provider.

## Immutable review corpus

| Role | Path | SHA-256 |
|---|---|---|
| manifest decision | `docs/decisions/VD-p7-readiness-manifest-2026-10-17.md` | `a8774fa303fe07f3b779ca23c1339c26868fe655e0b7945d1971dd126cbaa641` |
| manifest planning draft | `docs/plans/P7-readiness-manifest-plan-draft1.md` | `855f19a9440ddc9e24bd7bb56e2acb63be32548a24d9b876509933adb85b6049` |
| lock contract | `docs/plans/P7-replay-lock-spec-addendum-draft1.md` | `0ffc074b67170145fbd47a9fa034aa64ce5cbd4928bc53b83c7ecb639c89353a` |
| fixture matrix | `docs/plans/P7-replay-lock-fixture-matrix-draft1.md` | `93edb519bc5da643ae4b0cf365b879e2f8b0115ec66f1545b72cb6e10af4cef4` |
| external-prerequisites amendment | `docs/plans/P7-replay-lock-external-prerequisites-amendment-draft1.md` | `922e708551d80658fee045e38c6f1487e90bdb5aed1f0eecb39c6cc044811a0c` |
| first-slice scope | `docs/plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md` | `26e02341ca27d176f5c4457fc1a492103555957da8f1cec5efc2f6c88d33949d` |
| roadmap boundary | `ROADMAP.md` | `482bc0c65ea9b3ea1df93d4f1c537ef07beffe1832d3a86d63dcdc07e761db00` |

The review must bind its verdict to the full commit above, not to a later working tree or a
conversation summary. A changed byte requires a new packet and fresh review.

## Review questions

1. Does the exact top-level member set provide a sufficient evidence index for one bounded
   `ReplayIsolated` case without becoming a second lock, ledger, proof, or writable truth?
2. Are the no-unknown, no-missing, no-`null`, RFC 8785, lowercase-SHA-256, repository-relative
   path, and deterministic-order rules precise enough for a future parser and independent verifier?
3. Is the derived `case_status` rule fail-closed, especially for missing evidence, pending review,
   `UNDECLARED` intent, and a dormant PostgreSQL prerequisite?
4. Are the `PASS`, `FAIL`, `UNVERIFIED`, `UNDECLARED`, and `PENDING` outcome states scoped clearly
   enough that no outcome can be upgraded by wording alone?
5. Do the command, output, ledger, network, cleanup, artifact, and review references preserve the
   evidence required by the accepted lock contract and fixture matrix?
6. Does the manifest remain compatible with the external-prerequisites amendment without
   approving or revalidating the pinned PostgreSQL input?
7. Are the planning boundary, non-goals, and implementation stop points explicit enough to prevent
   a future agent from treating this packet as implementation authorization?

## Acceptance bar

Return `ACCEPT` only if the planning shape is coherent and complete for a future implementation
review, with no blocking, high, medium, or low findings. Return `AMEND` with exact findings and
required changes if any rule is ambiguous, incomplete, or inconsistent. Return `REJECT` if the
shape or boundary should be abandoned. The verdict must identify the reviewer as non-author and
include the reviewed commit and packet artifact digest.

An `ACCEPT` here accepts documentary planning only. It does not accept runtime fixtures, L12/L17,
the PostgreSQL provenance/licensing prerequisite, a live API/SPEC change, or implementation.

## Current gate state

- `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant.
- No PostgreSQL input was downloaded, inspected, approved, or replaced for this packet.
- No parser, fixture corpus, runtime case, or P7 implementation was added.
- The repository's targeted documentation, link, index, and deferral checks passed before packet
  creation.

## Verdict slot

No verdict yet. The authoring session must not fill this section with an acceptance decision.
