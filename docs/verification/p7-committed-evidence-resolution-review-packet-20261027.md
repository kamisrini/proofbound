# P7 committed-evidence resolution — independent implementation review packet

**Review type:** independent non-author review of the authorized evidence-resolution slice
**Status:** `PENDING`
**Intent:** `P7-NONRUNTIME-COMMITTED-EVIDENCE-RESOLUTION-1`
**Change intent:** [`CI-p7-committed-evidence-resolution-1a2b3c`](../intent/records/change-intents/CI-p7-committed-evidence-resolution-1a2b3c/0001.md)
**Reviewed revision:** `dd497f2ae419f0d38af759c86ffe0029fe012aab`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision add only an injected committed-evidence reader boundary that
verifies declared `BOUND` paths at their explicit revision and SHA-256 identity, preserves
`UNVERIFIED` fields without reader calls, and performs no filesystem, Git, network, external
artifact, binary, runtime, or PostgreSQL operation itself?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-committed-evidence-resolution-1a2b3c/0001.md` | `368515aa854b16ba3a739a32e7a986969d570fde732837bf55a2489d5859c609` |
| `kernel/internal/twin/manifest/SPEC.md` | `a7c71cc419d2aee278cc1ed85aedb0d285e5c4a4c5991ecff6934bf25646a428` |
| `kernel/internal/twin/manifest/inventory.go` | `6873d8831c68e76a0dc9258a7be6eafb02516df7231ea860beeb67fb1fe48524` |
| `kernel/internal/twin/manifest/evidence.go` | `8e7604aad69fe53eadb55b4067fe132bdbc751cb1c9b1a7110b1b34e2b024865` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/manifest_test.go` | `6add2342f1a60573c435db9dc842e8bf12c4c997d0a2ad9d65cefc9ad177d37c` |
| `docs/verification/p6-event-route-matrix.md` | `5d6065a1cffef3a7bddbd40706406f66fe14666a230d67fffb404f11325d508d` |

## Required reviewer checks

1. Verify the reviewed revision and every frozen corpus digest above.
2. Confirm `ResolveInventory` validates the inventory before resolution and requests only declared
   `BOUND` paths at their declared evidence revision.
3. Confirm returned bytes are hashed and compared to the declared SHA-256, with missing paths,
   unavailable revisions, unsafe paths, and mismatches failing closed.
4. Confirm `UNVERIFIED` fields are preserved and never passed to the reader.
5. Confirm the package does not invoke Git, read the filesystem, fetch the network, retrieve
   external artifacts, build or execute binaries, run PostgreSQL, or alter runtime behavior.
6. Confirm synthetic tests cover successful resolution, unverified preservation, missing paths,
   unavailable revisions, unsafe paths, and digest mismatches.
7. Re-run the listed checks and return `ACCEPT`, `AMEND`, or `REJECT` with findings.

## Verification already run by the author

- Manifest and intent-record tests: PASS.
- `scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS.

## Explicit non-claims

This packet does not accept complete projector/build evidence, external-artifact provenance,
toolchain or module retrieval, binary builds, runtime fixtures, PostgreSQL execution or approval,
`DEF-P7-001`, L12/L17, behavior-lock acceptance, clean-clone proof, or a runtime verdict. The
pinned PostgreSQL JAR remains unapproved.
