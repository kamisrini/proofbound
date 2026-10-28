# P7 projector/build evidence capture — 2026-10-28

**Intent:** `P7-NONRUNTIME-PROJECTOR-BUILD-EVIDENCE-1`
**Artifact:** `proofbound.p7.projector-manifest.v1`
**Evidence revision inspected:** `7eeeb1ba297d046ac68f2599ed38b1c032fdd72f`
**Status:** documentary capture complete; complete build evidence `UNVERIFIED`

This capture inventories existing records and read-only measurements at the stated revision.
The 2026-10-20 Phase-A probe is historical diagnostic evidence; its local output is not an
artifact of this evidence revision. No build or binary validation was performed in this slice.

## Existing projector/build command record

The [Phase-A probe](p7-replay-lock-fixtures/phase-a-build-probe-20261020.md) recorded these commands
from `kernel/` with the historical pinned environment:

```text
GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go version
GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go list -m all
GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go build -o /tmp/proofbound-p7-projector-probe ./cmd/proofbound
```

Its captured output is `go version go1.26.7 linux/amd64`, `PASS` for the locally resolved module
graph, and `PASS` for the local build. It measured an untracked temporary executable at
17,362,505 bytes with SHA-256
`4d992a5650620e46efc9f4a209cd0bd4c183ca7475f58a9c13fa229c29328222`.
Those are observations at probe revision `007d71bdb94b165f93ce696f01595276f8042a6d`, not
repeated results or a reviewed executable for the revision inspected here. The probe does not
record the full `go list -m all` stdout, effective build environment, or complete source file
set, so it cannot reproduce or bind a complete projector manifest on its own.

## Current read-only capture

Commands run from the repository root at the evidence revision above:

```text
git rev-parse HEAD
sha256sum kernel/go.mod kernel/go.sum docs/verification/p7-replay-lock-fixtures/phase-a-build-probe-20261020.md docs/verification/p7-replay-lock-fixtures/phase-a-module-corpus-inventory-20261020.md docs/verification/verdicts/p7-go-module-corpus-round1-20261023.md docs/verification/p7-replay-lock-fixtures/setup-sources.json
git ls-files 'docs/verification/p7-replay-lock-fixtures/bytes/sha256/*' | wc -l
git ls-files 'kernel/*' | wc -l
git ls-tree HEAD kernel/go.mod kernel/go.sum kernel/internal/projections/projections.go
```

Captured results:

| Input or measurement | Result |
|---|---|
| `git rev-parse HEAD` | `7eeeb1ba297d046ac68f2599ed38b1c032fdd72f` |
| `kernel/go.mod` SHA-256 | `0cefb499cbd9ef6cfe917ae58d3857c0456481d01c45161a6e358c396d9d1393` |
| `kernel/go.sum` SHA-256 | `3b5ef08294f1dbcb6a1acc9fc011410f7ebcb78becc0e8b6c0aeb191c776f1bf` |
| Phase-A probe note SHA-256 | `968bf4186ad9a22b810b1e03fcc39c0ac1a3a687b8b5a4be92ab6bbcb39341cf` |
| Phase-A module inventory note SHA-256 | `dbed84ed3aba1484c005316d9afb45c806a1804c2c15784231ee95864499a904` |
| Module corpus verdict SHA-256 | `214b7242a399834eb3b3db3eb5f08bb03182a267d807d25f1d4892d63f0e9459` |
| Candidate setup-source manifest SHA-256 | `e312bf28cea20d516e3b21704a53ba3ebc8ac84ff988eccb7da555317b900b33` |
| Git-tracked content-addressed source artifacts | 154 paths; accepted corpus is 77 `.mod` and 77 `.zip` files |
| Git-tracked `kernel/*` paths | 114; a count, not a validated build input list |
| `git ls-tree` modes | `100644` for `kernel/go.mod`, `kernel/go.sum`, and `kernel/internal/projections/projections.go` |

The [module corpus verdict](verdicts/p7-go-module-corpus-round1-20261023.md) accepts only the
source-archive corpus, not a projector build. The [manifest contract](../../kernel/internal/twin/manifest/SPEC.md)
defines source, toolchain, module, and build fields and preserves incomplete evidence states.
The [declaration verdict](verdicts/p7-projector-manifest-declaration-round1-20261027.md)
accepts the non-runtime declaration implementation, not a complete concrete manifest.
The [candidate setup-source manifest](p7-replay-lock-fixtures/setup-sources.json) names the Go
archive, but its [prerequisite status](p7-replay-lock-prerequisites/README.md) remains proposed;
its declared digest is not verified toolchain/archive evidence here. The earlier
[module inventory](p7-replay-lock-fixtures/phase-a-module-corpus-inventory-20261020.md)
records a historical fetch and local cache; the later corpus verdict is the applicable source
artifact acceptance.

## Discrepancies and pending bindings

- `UNVERIFIED`: the historical build used ambient Go installation and module cache paths;
  no verified archive, extracted tree inventory, or complete toolchain binding accompanies it.
- `PENDING`: no exact source-file corpus, accepted projector revision, full selected module
  graph mapping, and build settings have been assembled into one concrete manifest for the
  current evidence revision. The 114-path count does not establish which paths the build consumes.
- `UNVERIFIED`: the historical executable digest and output are untracked, not reproduced at
  this revision, and not independently reviewed as an executable artifact.
- `PENDING`: independent non-author review of this documentary capture, if sought, must bind
  its final bytes and exact reviewed revision. No independent runtime verdict is claimed.

The documentary capture is complete within the authorized slice. Complete projector/build
evidence, reproducible-build validity, source/toolchain completeness, and runtime acceptance
remain unestablished. `DEF-P7-001` stays dormant; the PostgreSQL 16.4.0 Maven JAR is unapproved.
