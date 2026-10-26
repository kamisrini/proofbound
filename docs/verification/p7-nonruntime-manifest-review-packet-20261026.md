# P7 non-runtime manifest — independent implementation review packet

**Review type:** independent non-author review of the authorized non-runtime slice
**Status:** `ACCEPT`
**Intent:** `P7-NONRUNTIME-MANIFEST-1`
**Change intent:** [`CI-p7-nonruntime-manifest-1a2b3c`](../intent/records/change-intents/CI-p7-nonruntime-manifest-1a2b3c/0001.md)
**Reviewed revision:** `66cd3441ab4952eb0c50d16a2e39071020035492`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision implement only the authorized documentary projector/build-manifest
and source/toolchain-binding preparation contract, with strict shape/revision/digest validation and
synthetic tests, while preserving every explicit runtime and binary exclusion?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-nonruntime-manifest-1a2b3c/0001.md` | `dc597e6cbed9f68b6be86bda12ffe63b13b58d3ba819e37274094d3f1fcfb0e4` |
| `docs/plans/P7-next-nonruntime-slice-proposal-draft1.md` | `70b220f4aea04e0e3d2bd2d72d5474f8fbc85f370f979e8bde74815461d1e233` |
| `docs/decisions/VD-p7-nonruntime-manifest-2026-10-26.md` | `1c47755aefacc8da4e1d31e0ec36db0f5af290e2254337b94211b4d512ca310a` |
| `kernel/internal/twin/SPEC.md` | `21497867765876e34c03de121875a08c1dd69ada664e5661efe10ebb499aedb4` |
| `kernel/internal/twin/manifest/SPEC.md` | `4524aa9ee3f7f3f6d48319229a81d9cc68633dc35eb4bb7cd9ee0c1ce676ea72` |
| `kernel/internal/twin/manifest/manifest.go` | `3b363b642ec86458ce735d83f69fef9cd5c5716d6a9636ed1c852ea20eee890f` |
| `kernel/internal/twin/manifest/manifest_test.go` | `89101f2cd2389dbb47c44c2ac6a1e703b32c97abdc69e227c7cda755ae62bf7a` |

## Required reviewer checks

1. Verify the reviewed revision and every corpus digest above.
2. Confirm `Parse` rejects unknown, missing, null, duplicate, malformed, and non-canonical input.
3. Confirm source paths, file modes, module ordering, revisions, tags, and all SHA-256 values fail
   closed when malformed, stale, duplicated, or reordered.
4. Confirm `Digest` hashes only validated canonical bytes and changes when an exact revision changes.
5. Confirm no code path performs network access, external artifact retrieval, binary build or
   validation, PostgreSQL execution, runtime fixture execution, or runtime acceptance.
6. Confirm the generated decision index, P6 census classification, and route matrix remain
   consistent with the new package.
7. Re-run the listed checks and return `ACCEPT`, `AMEND`, or `REJECT` with findings.

## Verification already run by the author

- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go test ./...`: PASS.
- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check`: PASS, `0 issues`.
- `scripts/deferrals.sh --check`: PASS (`rows=2 open=1`).
- `scripts/p6-artifact-integrity.sh --check`: PASS.
- `scripts/p6-census.sh --check`: PASS.
- `scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS.

The hostile package-universe test reports `expected 16 production packages, got 17` while proving
that the frozen P6 package acceptance set rejects an unapproved package. The live census separately
classifies `internal/twin/manifest` as `defer-to-P7`; this diagnostic is not a failure.

## Explicit non-claims

This packet does not accept runtime fixtures, binary provenance or licensing, PostgreSQL execution,
L12/L17, behavior locks, clean-clone proof, or a runtime verdict. `DEF-P7-001` remains dormant.
