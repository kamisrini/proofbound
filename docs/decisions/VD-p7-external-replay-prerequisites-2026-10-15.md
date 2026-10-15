# VD-p7-external-replay-prerequisites-2026-10-15: public binary setup outside Git

**Status:** OWNER DIRECTION RECORDED; proposed contract amendment awaits fresh non-author
review. This does not replace the accepted P7 evidence contract yet or authorize code.
**Date:** 2026-10-15

## Decision requested by owner

Keep public third-party Go toolchain and embedded-Postgres binary archives, downloaded Go
module `.mod`/`.zip` artifacts, and the locally built executable out of Git. Commit authored
code, small manifests, lock/fixture evidence, and review records only. Document prerequisite
setup ahead of the replay test. Never commit a Postgres data directory or ledger database.

This reverses the Git-tracked build/runtime byte route in the previously accepted
[P7 addendum](../plans/P7-replay-lock-spec-addendum-draft1.md),
[fixture matrix](../plans/P7-replay-lock-fixture-matrix-draft1.md), and accepted
[first-slice scope](../plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md).
Those exact reviewed artifacts remain historical evidence; their clean-clone/offline claims are
not silently reinterpreted. The proposed replacement is
[the external-prerequisites amendment](../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
and its [setup README](../verification/p7-replay-lock-prerequisites/README.md).

## Consequence and gate

A plain clone with empty caches and no network cannot execute L12 under this policy. The
replacement claim is narrower: prepare exact public inputs in an explicit network-permitted
phase, verify all bytes against reviewed hashes, then run replay with egress disabled and
loopback available. Public availability is not a substitute for a matching digest or an
attempt-level zero-fetch witness. Missing, changed, or unavailable inputs yield `UNVERIFIED`;
they are not an accepted replay result.

The P7 `ReplayIsolated` subject, frozen `vera.replay.v1` proof, authored/observed boundary,
source-ledger preservation, and no-new-provider/feed/event/writable-truth exclusions do not
change. The old L12 clean-clone-offline claim and accepted first-slice provisioning details
must be amended and freshly reviewed before any implementation. No runtime fixture, L12/L17
pass, requirement satisfaction, or implementation verdict is accepted by this owner direction.
