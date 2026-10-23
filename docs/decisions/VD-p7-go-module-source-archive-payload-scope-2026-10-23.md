# VD-p7-go-module-source-archive-payload-scope-2026-10-23

**Status:** OWNER-APPROVED SCOPE DECISION; independent final corpus review remains required.
**Date:** 2026-10-23

## Decision

For `DEF-P7-002`, executable, library, and test-data members nested inside the required
third-party Go module source `.zip` artifacts are treated as contents of those immutable source
archives, not as standalone runtime-binary inputs. They are out of scope of the standalone
binary-exclusion boundary established by
[`VD-p7-external-replay-prerequisites-2026-10-15`](VD-p7-external-replay-prerequisites-2026-10-15.md).

This decision is limited to the selected Go module source corpus. It does not approve a
PostgreSQL distribution, Go toolchain archive, projector executable, runtime fixture, L12/L17
result, binary provenance, licensing, or any runtime use of the nested members. The nested
members must not be extracted for execution or used as a projector, toolchain, database, or
runtime input.

The owner reviewed and approved this scope on 2026-10-23. The complete committed corpus still
requires independent non-author review before `DEF-P7-002` can be closed.

## Reviewed nested members

These are the nine members identified in the selected 77-module, 154-artifact corpus:

| Module | Archive member | Size (bytes) |
|---|---|---:|
| `golang.org/x/sys v0.33.0` | `windows/testdata/ev-signed-file.exe` | 18,152 |
| `golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d` | `go/gccgoexportdata/testdata/long.a` | 802 |
| same | `go/gccgoexportdata/testdata/short.a` | 710 |
| same | `go/internal/gccgoimporter/testdata/libimportsar.a` | 9,302 |
| same | `internal/gcimporter/testdata/versions/test_go1.16_i.a` | 3,464 |
| same | `internal/gcimporter/testdata/versions/test_go1.17_i.a` | 3,638 |
| same | `internal/gcimporter/testdata/versions/test_go1.18.5_i.a` | 4,162 |
| same | `internal/gcimporter/testdata/versions/test_go1.19_i.a` | 4,026 |
| same | `internal/gcimporter/testdata/versions/test_go1.20_u.a` | 4,288 |

## Binding and review boundary

The parent `.mod` and `.zip` files remain subject to the accepted content-addressed evidence
route at `docs/verification/p7-replay-lock-fixtures/bytes/sha256/<digest>`. This decision does
not change the required 77-module graph, 154 artifacts, or measured total of 171,215,534 bytes.
The final Git tree, exact artifact hashes, complete count/size inventory, and this scope decision
must be reviewed by a non-author verifier before the deferral is closed.
