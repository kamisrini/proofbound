# P6 Task 2 connector-reality SPEC

## Scope

This task re-proves the existing connector estate and Proofbound's explicit delivery boundary. It
adds no connector, event kind, provider, ambient capture, or external enforcement claim. The exact
C4 universe is the 12 rows in `docs/plans/p6-census-rows.tsv`.

## Connector classification

Every production package below `kernel/internal/connector/` is classified exactly `live`,
`synthetic`, or `none` at one committed code revision. `live` requires a dated external or genuine
local-corpus run; `synthetic` requires a discriminating fixture test; `none` is not acceptable for a
C4 `close-in-P6` row. Classification cites exact artifacts and tests rather than package presence.

| ID | Assertion | Evidence required |
|---|---|---|
| T2-INV-1 | The nine discovered production connector packages are each classified once | closure checker compares C4 connector rows with tracked non-test connector packages |
| T2-INV-2 | Sessions accepts a genuine quiescent JSONL file without retaining message content | explicit `liveacceptance` Go test reports total/valid/skipped lines, one appended metadata-only event, then zero appended on replay |
| T2-INV-3 | GitHub retains owner/repository allowlisting, the 200-record v1 bound, and workflow/deployment-only scope | package SPEC plus exact hostile tests and retained P3 live artifact |
| T2-INV-4 | Multiple configured GitHub repositories retain owner/repository identity and cannot collide on equal upstream IDs | a two-repository package test proves distinct native IDs and payload identities |
| T2-INV-5 | GitHub report joins by exact commit and renders missing tested/deployed data explicitly | exact projection/report proving tests |
| T2-INV-6 | The controlled delivery boundary refreshes every promoted witness, then ingests all sources, then enforces | shell self-test asserts the complete ordered command log and fails on omission/reordering |
| T2-INV-7 | Plain `make check` remains independent of the Proofbound executable and ledger | Make-DAG checker rejects witnessed, sync, gate-enforce, or `proofbound` dependencies in the `check` recipe |

## Sessions live acceptance

The test receives one founder-owned source path through `PROOFBOUND_SESSIONS_LIVE_FILE`. It refuses
a missing, empty, or non-quiescent file (mtime newer than ten minutes). It copies the bytes only to a
Go temporary directory shaped like the connector's input directory; neither source bytes nor copied
bytes enter Git. It counts nonempty JSONL lines and valid JSON objects, and reports invalid/skipped
lines without printing source content. Acceptance requires at least one valid line, parse coverage
at least 50%, exactly one accepted `session.observed` event, metadata-only payload keys, absence of
known content-bearing keys/values, and a second sync with `appended=0` and `existing=1`.

The current lawful corpus is the migrated Windows Claude projects directory. Its path is machine
evidence, not a repository default. Normal tests do not silently claim live acceptance when the
environment variable is absent; the live test exists only under the explicit `liveacceptance` build
tag.

## GitHub retained narrowness

P6 explicitly retains P3's read-only public `github/docs` acceptance rather than rerunning a
mutable network feed. The retained artifact records 100 workflow runs plus 100 deployments,
zero-append replay, exact commit rendering, and explicit missing-side output. Current-code tests
must independently retain the configuration, bound, transport secrecy, multi-repository identity,
exact-commit join, and missing-data semantics. This is evidence retention, not a claim that the 2026
GitHub dataset is unchanged.

## Closure artifact

`docs/verification/p6-connector-reality.md` records the exact implementation commit, all 12 C4
rows, classifications, commands, counts, and test citations. A checker rejects a missing connector,
duplicate/missing classification, absent evidence path, incomplete GitHub narrowness inventory,
or incomplete delivery-order proof. Task 2 closes only when all C4 rows are closed,
`scripts/p6-census.sh --check` exits 0, the sessions live command exits 0, the connector and delivery
tests pass, and bare `make check` exits 0.

## Non-goals

- No session message-content persistence or committed live fixture.
- No new GitHub entity, pagination width, repository discovery, or network dependency in
  `make check`.
- No ambient witness capture and no claim about external consumer pipelines.
- No new product behavior unless a retained assertion fails and the smallest in-scope correction is
  required.
