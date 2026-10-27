# P7 evidence-manifest skeleton — independent implementation review packet

**Review type:** independent non-author review of the authorized non-runtime skeleton slice
**Status:** `ACCEPT`
**Intent:** `P7-NONRUNTIME-EVIDENCE-MANIFEST-SKELETON-1`
**Change intent:** [`CI-p7-evidence-manifest-skeleton-1a2b3c`](../intent/records/change-intents/CI-p7-evidence-manifest-skeleton-1a2b3c/0001.md)
**Reviewed revision:** `c9daba2d4afdef4a2966e54813da81c1b56363db`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision add only a pure deterministic evidence-manifest skeleton over the
accepted projector manifest, source/toolchain inventory, committed-evidence resolutions, and
closure result; preserve `UNVERIFIED` and `PENDING`; reject missing, extra, stale, mismatched,
contradictory, and non-canonical inputs; and make no runtime or external-artifact claims?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-evidence-manifest-skeleton-1a2b3c/0001.md` | `1b643addd447d55bd703c2ea0f869e59869fd3dda30c097a768631111db20f4e` |
| `kernel/internal/twin/manifest/SPEC.md` | `8a1b5525f6de8a91c06d778c766f427ffff9f8193b30be166fe8e72e186c5a72` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/inventory.go` | `6873d8831c68e76a0dc9258a7be6eafb02516df7231ea860beeb67fb1fe48524` |
| `kernel/internal/twin/manifest/evidence.go` | `c2f5c38128f43aa17861352bf6b746b88d9d75b8eb77c00b35df4ba9b069d56d` |
| `kernel/internal/twin/manifest/skeleton.go` | `5bac76b646c2fca4c593c703b707ccd0718b4ce6076d25767b842e25dc8e8c44` |
| `kernel/internal/twin/manifest/manifest_test.go` | `4ef89230294b6ba5b80c63089001f6e64f407496b98c797573e6203390865794` |
| `docs/verification/p6-event-route-matrix.md` | `14f7fb4d824b155ece5a113e38f2fa410b5cd87c4e0faccdb82e78ebd7eacdb8` |

## Required reviewer checks

1. Verify the reviewed revision and every frozen corpus digest above.
2. Confirm the skeleton derives exact manifest and inventory digests from canonical inputs and
   preserves the closure’s exact sorted field rows.
3. Confirm `UNVERIFIED` and `PENDING` states survive assembly and cannot be converted into a
   completeness claim by absence.
4. Confirm supplied closure, stale revision, manifest digest, inventory digest, field coverage,
   duplicate/extra/missing fields, contradictory resolutions, and unknown JSON members fail closed.
5. Confirm canonical JSON is required for parsing and serialization is deterministic.
6. Confirm the implementation performs no filesystem, Git, network, external-artifact, toolchain,
   binary, runtime, PostgreSQL, or provider/event-kind operation.
7. Re-run synthetic manifest and intent tests, `scripts/deferrals.sh --check`,
   `scripts/p6-artifact-integrity.sh --check`, `scripts/p6-route-matrix.sh --check`,
   `git diff --check`, and the pinned unrestricted `make check`.
8. Return `ACCEPT`, `AMEND`, or `REJECT` with findings, naming the exact reviewed revision.

## Verification already run by the author

- `go test ./internal/twin/manifest ./internal/connector/intent/...` with Go 1.26.7 and an
  isolated writable cache: PASS.
- `scripts/p6-route-matrix.sh --write && scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS before commit.

## Explicit non-claims

This packet does not accept complete projector/build evidence, reproducible build validity,
external-artifact provenance, toolchain downloads, binary builds, runtime fixtures, PostgreSQL
execution or approval, `DEF-P7-001`, L12/L17, behavior-lock acceptance, clean-clone proof, or a
runtime verdict. The pinned PostgreSQL JAR remains unapproved.
