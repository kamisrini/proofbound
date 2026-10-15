# VD-p7-external-replay-prerequisites-2026-10-15: public binary setup outside Git

**Status:** OWNER DIRECTION RECORDED; round-1 non-author review AMEND; fresh review pending.
See the [round-1 verdict](../verification/verdicts/p7-external-replay-prerequisites-round1-20261015.md).
This does not replace the accepted P7 evidence contract yet or authorize code.
**Date:** 2026-10-15

## Decision requested by owner

Do not commit executable or runtime binary payloads for this P7 route: the public Go toolchain
archive, the embedded-Postgres distribution JAR and nested `txz`, or the locally built
projector executable. The reviewed Go module `.mod` and `.zip` source artifacts remain
Git-tracked; a source archive is not excluded solely because its container is a `.zip`.
Source code, configuration, READMEs, manifests, lock/fixture evidence, and review records
are also eligible for Git. If a purported source artifact contains an executable/runtime
binary payload, stop for a separate review; do not silently commit that payload. The proposed
setup route obtains public binary distributions ahead of replay. Never commit a Postgres
data directory or ledger database.

This changes the binary portion of the Git-tracked build/runtime byte route in the previously
accepted
[P7 plan's evidence-controls exit criterion](../plans/P7-behavior-locks-regeneration-plan-draft1.md),
[addendum](../plans/P7-replay-lock-spec-addendum-draft1.md),
[fixture matrix](../plans/P7-replay-lock-fixture-matrix-draft1.md), and accepted
[first-slice scope](../plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md).
Those exact reviewed artifacts remain historical evidence; their clean-clone/offline claims are
not silently reinterpreted. The proposed replacement is
[the external-prerequisites amendment](../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
and its [setup README](../verification/p7-replay-lock-prerequisites/README.md).

## Consequence and gate

A plain clone with empty caches and no network cannot execute L12 under this policy. The
replacement claim is narrower: start from a plain clean clone, obtain exact public binary
inputs in an explicit network-permitted preparation phase, verify all bytes against reviewed
hashes, then replay with egress disabled and loopback available. This proposed prepared-offline
control replaces only the P7 plan's clean-clone/replay-from-committed-artifacts subcriterion;
the original empty-cache, no-network claim is not eligible for `PASS`. Public availability
is not a substitute for a matching digest or an attempt-level zero-fetch witness. Missing,
changed, or unavailable inputs yield `UNVERIFIED`, not an accepted replay result.

The P7 `ReplayIsolated` subject, frozen `vera.replay.v1` proof, authored/observed boundary,
source-ledger preservation, and no-new-provider/feed/event/writable-truth exclusions do not
change. The old L12 clean-clone-offline claim and accepted first-slice provisioning details
must be amended and freshly reviewed before any implementation. No runtime fixture, L12/L17
pass, requirement satisfaction, or implementation verdict is accepted by this owner direction.
