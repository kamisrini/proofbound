# P7 source/toolchain inventory — independent implementation review packet

**Review type:** independent non-author review of the authorized inventory slice
**Status:** `ACCEPT`
**Intent:** `P7-NONRUNTIME-SOURCE-TOOLCHAIN-INVENTORY-1`
**Change intent:** [`CI-p7-source-toolchain-inventory-1a2b3c`](../intent/records/change-intents/CI-p7-source-toolchain-inventory-1a2b3c/0001.md)
**Reviewed revision:** `987ed9104f8a0c9fad5facd9ee81aa030f57253b`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision add only a machine-readable source/toolchain evidence inventory
and synthetic validation, with exact manifest-field coverage, explicit `BOUND`/`UNVERIFIED` states,
matching evidence revisions, and manifest-digest binding, without reading paths, resolving Git
trees, retrieving artifacts, building binaries, or performing runtime/PostgreSQL work?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-source-toolchain-inventory-1a2b3c/0001.md` | `f570d00ccaab8da8070dc8a84e3a680de9db4e397949bd8900e71ab0bc458bf4` |
| `kernel/internal/twin/manifest/SPEC.md` | `053c83b70ca6b51055086cababc5ec2dce0a0f207bbe1aef0d706f48bdd1a39d` |
| `kernel/internal/twin/manifest/inventory.go` | `6873d8831c68e76a0dc9258a7be6eafb02516df7231ea860beeb67fb1fe48524` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/manifest_test.go` | `8dd23b5e62bec950800242f99d8798433146296eb3e36148f836318aa75f34af` |
| `docs/verification/p6-event-route-matrix.md` | `d8dca1c6e46a844bd3abf6681a9b15d74770fe59ffb3f07fa14debb86fd58592` |

## Required reviewer checks

1. Verify the reviewed revision and every frozen corpus digest above.
2. Confirm the inventory requires the exact sorted manifest-field set and rejects missing, extra,
   duplicate, unknown, and non-canonical fields.
3. Confirm `BOUND` requires a safe relative evidence path, matching evidence revision, and a
   lowercase SHA-256 declaration; confirm `UNVERIFIED` requires all evidence fields to be empty.
4. Confirm the inventory manifest digest equals the validated canonical manifest digest and stale
   revisions or digest mismatches fail closed.
5. Confirm validation performs no filesystem, Git-tree, network, external-artifact, toolchain,
   executable, runtime, or PostgreSQL operation.
6. Confirm synthetic tests cover exact field coverage, explicit `UNVERIFIED`, stale revision,
   digest mismatch, unknown/non-canonical shape, and canonical round-trip.
7. Re-run the listed checks and return `ACCEPT`, `AMEND`, or `REJECT` with findings.

## Verification already run by the author

- Manifest and intent-record tests: PASS.
- `scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS.

## Explicit non-claims

This packet does not accept complete projector/build evidence, source/toolchain completeness,
runtime fixtures, binary provenance or licensing, PostgreSQL execution, `DEF-P7-001`, L12/L17,
behavior locks, clean-clone proof, or a runtime verdict. The pinned PostgreSQL JAR remains
unapproved.
