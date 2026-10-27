# P7 evidence closure evaluator — independent implementation review packet

**Review type:** independent non-author review of the authorized closure slice
**Status:** `PENDING`
**Intent:** `P7-NONRUNTIME-EVIDENCE-CLOSURE-1`
**Change intent:** [`CI-p7-evidence-closure-1a2b3c`](../intent/records/change-intents/CI-p7-evidence-closure-1a2b3c/0001.md)
**Reviewed revision:** `3e31c3a191a15113cd26e007caf848c0d2a9b79a`
**Owner:** `project-owner`

## Review question

Does the exact reviewed revision add only a pure non-runtime evidence-closure evaluator that
requires exact resolution coverage, classifies fields as `BOUND`, `UNVERIFIED`, or explicit
`PENDING`, rejects incomplete or contradictory inputs, and never turns missing evidence into an
overall `BOUND` result?

## Frozen review corpus

| Path | SHA-256 |
|---|---|
| `docs/intent/records/change-intents/CI-p7-evidence-closure-1a2b3c/0001.md` | `b925e5b5d8c1944eb9d58a0ea41d10a839229dc4ae42c037eb9096afe2a0f66c` |
| `kernel/internal/twin/manifest/SPEC.md` | `7bb3dca1be5097154fb18f0a4d75a81e8ee9a1cd5e6f7c03bf396ac26cf64672` |
| `kernel/internal/twin/manifest/inventory.go` | `6873d8831c68e76a0dc9258a7be6eafb02516df7231ea860beeb67fb1fe48524` |
| `kernel/internal/twin/manifest/evidence.go` | `c2f5c38128f43aa17861352bf6b746b88d9d75b8eb77c00b35df4ba9b069d56d` |
| `kernel/internal/twin/manifest/manifest.go` | `0f0f30c4e08beb08150121421c7af684b6da7060c57936b2e6d34f430e8abe00` |
| `kernel/internal/twin/manifest/manifest_test.go` | `8958c49541454f167fb6fc230879675d0bd5d7580b1b8803bb741b146cb7521e` |
| `docs/verification/p6-event-route-matrix.md` | `97ca5a18e0f804b96e45cee639af36949924538c72d21a3deb8ab1a34f200e3a` |

## Required reviewer checks

1. Verify the reviewed revision and every frozen corpus digest above.
2. Confirm exact sorted resolution coverage is required and missing, extra, duplicate, or
   malformed fields fail closed.
3. Confirm `BOUND` requires valid evidence identity, `UNVERIFIED` requires an explicit reason and
   no evidence declaration, and `PENDING` requires an explicit reason and no evidence declaration.
4. Confirm contradictory inventory/resolution states and stale revisions fail closed.
5. Confirm overall precedence is `PENDING`, then `UNVERIFIED`, then `BOUND`, so incomplete input
   cannot become complete by absence.
6. Confirm the evaluator performs no filesystem, Git, network, external-artifact, toolchain,
   binary, runtime, or PostgreSQL operation.
7. Confirm synthetic tests cover complete, unverified, pending, missing, extra, duplicate,
   contradictory, stale, and invalid-state inputs.
8. Re-run the listed checks and return `ACCEPT`, `AMEND`, or `REJECT` with findings.

## Verification already run by the author

- Manifest and intent-record tests: PASS.
- `scripts/p6-route-matrix.sh --check`: PASS.
- `git diff --check`: PASS.

## Explicit non-claims

This packet does not accept complete projector/build evidence, reproducible build validity,
external-artifact provenance, toolchain downloads, binary builds, runtime fixtures, PostgreSQL
execution or approval, `DEF-P7-001`, L12/L17, behavior-lock acceptance, clean-clone proof, or a
runtime verdict. The pinned PostgreSQL JAR remains unapproved.
