# P7 Phase-A module corpus inventory — 2026-10-20

**Status:** `UNVERIFIED`; source-only inventory, not a clean-clone proof.

**Probe revision:** `eacc574d8cf5104084653dc372cd00e098e6d23c`

This inventory follows the authorized non-PostgreSQL Phase-A build probe. It measures the exact
selected Go module graph needed by `kernel/go.mod` without downloading or inspecting the deferred
PostgreSQL distribution. The downloaded source archives remain in the local Go module cache and
were not copied into Git.

## Retrieval

Executed from `kernel/`:

```text
GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH \
GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org \
go mod download all
```

The retrieval completed successfully. A preceding `GOPROXY=off go mod download -json all` probe
reported 63 missing module lookups, confirming that the pre-retrieval local cache could not prove
a clean-clone corpus.

## Inventory result

| Measure | Result |
|---|---:|
| Selected module objects | 77 |
| Required `.mod` artifacts | 77 |
| Required `.zip` artifacts | 77 |
| Total source artifacts | 154 |
| Total local artifact bytes | 171,215,534 |
| Retrieval scope | Go module source only |
| PostgreSQL binary retrieval | none |

Each selected module now has a local `.mod` and `.zip` path under the Go module download cache.
The cache is not an evidence revision: it is not Git-tracked, is not available to a plain clone,
and was not independently reviewed as the final content-addressed corpus.

## Remaining decision and boundary

The accepted P7 evidence contract requires the complete `.mod`/`.zip` set to be Git-tracked under
`docs/verification/p7-replay-lock-fixtures/bytes/sha256/<digest>`. Adding the measured 171 MB
corpus is a material repository expansion. Until that owner decision is made, the projector/build
manifest remains incomplete and all clean-clone/runtime claims remain `UNVERIFIED`.

This inventory does not activate `DEF-P7-001`, validate a PostgreSQL binary, establish licensing
or provenance, or create a runtime case manifest. No source archive is being claimed as accepted
evidence merely because it is present in the local cache.
