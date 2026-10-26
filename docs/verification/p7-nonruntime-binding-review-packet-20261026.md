# P7 non-runtime binding — independent implementation review packet

**Review type:** independent non-author review of the authorized binding slice
**Status:** `ACCEPT`
**Intent:** `P7-NONRUNTIME-BINDING-1`
**Change intent:** [`CI-p7-nonruntime-binding-1a2b3c`](../intent/records/change-intents/CI-p7-nonruntime-binding-1a2b3c/0001.md)
**Reviewed revision:** `385274060b62e0dd912b80bd946b494b45277e05`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision add only deterministic binding of caller-supplied source and
module bytes into the accepted manifest contract, without filesystem, network, external-artifact,
toolchain, executable, runtime, or PostgreSQL operations?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-nonruntime-binding-1a2b3c/0001.md` | `6c36765064b56bde715d4965c201aa262d04a2599be9a9f5f49feb3c2dcd88a4` |
| `kernel/internal/twin/manifest/SPEC.md` | `a6162c6979b7807440d607abca4230f546c6edc279dc9436407390c08cd4de6e` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/manifest_test.go` | `e28bfb6af2facc2ac410fcfd10506d64e40316c222df0500e7b153514c0a8b24` |
| `docs/verification/p6-event-route-matrix.md` | `f1088e424b348f213a37a94ef423397741d60cd36488828eca477e4961319e86` |

## Required reviewer checks

1. Verify the reviewed revision and every corpus digest above.
2. Confirm `Bind` hashes only caller-supplied bytes, returns deterministic sorted identities, and
   does not read paths, access the network, execute bytes, or retain input buffers.
3. Confirm duplicate source/module identities, unsafe modes, invalid revisions, and malformed
   declarations fail closed.
4. Confirm the new tests cover byte identity, deterministic ordering, duplicates, and invalid input.
5. Confirm no runtime fixture, external artifact, binary, PostgreSQL, L12/L17, or behavior-lock
   claim was added.
6. Re-run the listed checks and return `ACCEPT`, `AMEND`, or `REJECT` with findings.

## Verification already run by the author

- Manifest, binding, and intent-record tests: PASS.
- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check`: PASS, `0 issues`.
- `scripts/deferrals.sh --check`: PASS (`rows=2 open=1`).
- `scripts/p6-artifact-integrity.sh --check`: PASS.
- `scripts/p6-census.sh --check`: PASS.
- `scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS.

## Explicit non-claims

This packet does not accept runtime fixtures, binary provenance or licensing, PostgreSQL execution,
L12/L17, behavior locks, clean-clone proof, or a runtime verdict. `DEF-P7-001` remains dormant.
