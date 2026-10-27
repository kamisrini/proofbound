# P7 projector/build manifest declaration — independent implementation review packet

**Review type:** independent non-author review of the authorized non-runtime declaration slice
**Status:** `PENDING`
**Intent:** `P7-NONRUNTIME-PROJECTOR-MANIFEST-DECLARATION-1`
**Change intent:** [`CI-p7-projector-manifest-declaration-1a2b3c`](../intent/records/change-intents/CI-p7-projector-manifest-declaration-1a2b3c/0001.md)
**Reviewed revision:** `ded357462fa6dfdad6d0a5348e989e75aadb6f78`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision declare a concrete projector/build manifest by reading only
caller-supplied bytes from the existing committed-evidence reader at one exact revision, bind
sorted source paths and accepted module identities deterministically, and preserve unavailable
toolchain/archive and completeness evidence as explicit `UNVERIFIED` or `PENDING` without making
runtime, binary, PostgreSQL, complete, or reproducible-build claims?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-projector-manifest-declaration-1a2b3c/0001.md` | `897c609dd1577b7425c1b15a0d610a562b0d7310c9007f12bdd30a3f84341804` |
| `kernel/internal/twin/manifest/SPEC.md` | `49fc920c54e75bf212e7c85d9cb182f24fc643210e851d3d95c875a9a03cf2ae` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/inventory.go` | `6873d8831c68e76a0dc9258a7be6eafb02516df7231ea860beeb67fb1fe48524` |
| `kernel/internal/twin/manifest/evidence.go` | `c2f5c38128f43aa17861352bf6b746b88d9d75b8eb77c00b35df4ba9b069d56d` |
| `kernel/internal/twin/manifest/skeleton.go` | `5bac76b646c2fca4c593c703b707ccd0718b4ce6076d25767b842e25dc8e8c44` |
| `kernel/internal/twin/manifest/declaration.go` | `d56410145cf6e83b8563f4bea748e7a70d5204e03f865ab3f6cdae373a389202` |
| `kernel/internal/twin/manifest/manifest_test.go` | `9d699ecc1146a170a3b9edbb9ecdca93257a9ac5b0b4c862768c0e21f39e763d` |
| `docs/verification/p6-event-route-matrix.md` | `f425538c65b105b722dff0ffa9807e7bfe602a27b1341c51151db07ce2bb7966` |

## Required reviewer checks

1. Verify the reviewed revision and every frozen corpus digest above.
2. Confirm source paths and module metadata/archive paths require exact sorted, unique, safe
   committed declarations at the supplied evidence revision.
3. Confirm the implementation reads only through `CommittedEvidenceReader`, computes source and
   module SHA-256 identities, and carries exact source/projector revisions and build settings.
4. Confirm unavailable toolchain/archive fields are explicit `UNVERIFIED`, other unestablished
   completeness fields are explicit `PENDING`, and neither state can become complete by absence.
5. Confirm missing bytes, unavailable revisions, duplicate/unsafe/stale inputs, contradictory
   evidence, unknown fields, and non-canonical JSON fail closed.
6. Confirm no filesystem, Git, network, external-artifact, toolchain-download, binary, runtime,
   PostgreSQL, provider, event-kind, writable-truth, or behavior operation occurs.
7. Re-run synthetic and committed-byte manifest tests, intent-record tests,
   `scripts/deferrals.sh --check`, `scripts/p6-artifact-integrity.sh --check`,
   `scripts/p6-route-matrix.sh --check`, `git diff --check`, and the pinned unrestricted
   `make check`.
8. Return `ACCEPT`, `AMEND`, or `REJECT` with findings, naming the exact reviewed revision.

## Verification already run by the author

- `go test ./internal/twin/manifest ./internal/connector/intent/...` with Go 1.26.7 and an
  isolated writable cache: PASS.
- Tests cover committed-byte binding and digest derivation, canonical round-trip serialization,
  explicit incomplete states, invalid revisions, duplicate/unsafe declarations, and missing
  committed bytes.
- `git diff --check`: PASS.
- `scripts/p6-route-matrix.sh --write && scripts/p6-route-matrix.sh --check`: PASS.

## Explicit non-claims

This packet does not accept a complete projector/build manifest, complete source/toolchain
binding, reproducible build validity, toolchain or module downloads, external-artifact provenance,
binary builds or execution, runtime fixtures, PostgreSQL execution or approval, `DEF-P7-001`,
L12/L17, behavior-lock acceptance, clean-clone proof, or any runtime verdict. The pinned
PostgreSQL JAR remains unapproved.
