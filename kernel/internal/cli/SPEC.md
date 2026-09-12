# CLI composition contract

## 1. Boundary

This package is Proofbound's reusable command composition root. It parses commands, opens the
ledger, constructs connectors and projections, and renders command results. Executable packages
under `cmd/` are deliberately tiny; `cmd/proofbound` is the only live entry point.

The package owns no semantic event contract. Frozen wire identities remain owned by their connector
or replay SPECs.

P5 adds `proofbound sync intent {records|specdir|all}`. Its committed-tree reader uses only
`git ls-tree` and `git show` at an explicit tree. `sync all` ingests both intent providers before
Git so an explicit commit claim cannot outrun the exact record revision it names.
It also routes `proofbound report intent <id>`, `proofbound report requirement <id>`, and
`proofbound intent check --commit <sha>` to the proof-bearing projection API without interpreting
chain semantics in the CLI.

## 2. Interface lock

```go
func Run(ctx context.Context, program string, args []string, stdout, stderr io.Writer) int
```

`program` is retained for composition compatibility but does not select an identity. Unknown command forms return status 2 and
print usage for the live name. Runtime failures return status 1 and are prefixed with `proofbound:`.

## 3. Identity and compatibility

- Local state is rooted only at `<repository>/.proofbound`. The one-time migration is an explicit,
  fail-safe directory move documented in `docs/proofbound-identity-migration.md`; commands never
  merge two state directories or silently fall back to the old path.
- Live environment names use only `PROOFBOUND_*`; legacy names are ignored after their one-phase
  P5 window.
- `cmd/proofbound` is the only executable wrapper.
- `DATABASE_URL`, `GITHUB_TOKEN`, and standard tool environment names are not product-namespaced and
  are unchanged.
- Removal is governed by the dated advisory in `docs/gates.md`, not by an undocumented future edit.

## 4. Non-goals

- No interpretation or rewriting of frozen `vera.witness.v1`, `vera.verdict.v1`, or
  `vera.replay.v1` bytes.
- No automatic merge or deletion of local state directories.
- No compatibility promise beyond the single ratified phase.

## 5. Invariants and proving tests

| Invariant | Statement | Proving test |
|---|---|---|
| CLI-INV-1 | The live executable and all usage/runtime diagnostics name Proofbound | cli_test.go::TestRunUsesProofboundIdentity |
| CLI-INV-2 | Every command opens state and witness spool paths below `.proofbound` | cli_test.go::TestOpenStoreUsesProofboundState |
| CLI-INV-3 | Live environment values are read only from the Proofbound namespace | cli_test.go::TestProductEnvPrefersProofbound |
| CLI-INV-4 | Legacy environment names are ignored after P5 | cli_test.go::TestLegacyEnvironmentIsIgnored |
| CLI-INV-5 | The legacy executable wrapper is absent | cli_test.go::TestLegacyCommandWrapperIsAbsent |
| CLI-INV-6 | Frozen witness bytes remain accepted without reinterpretation | sync_integration_test.go::TestSyncChecksRebuildAndVerify |
| CLI-INV-7 | Intent commands select only records, specdir, or both and read committed trees | cli_test.go::TestParseIntentCommandsAndCommittedReader |
| CLI-INV-8 | Full sync orders all intent providers before commit ingestion | cli_test.go::TestSyncAllOrdersIntentBeforeGit |
| CLI-INV-9 | Intent report and check identifiers are passed unchanged to projections | cli_test.go::TestParseIntentReportAndCheckCommands |
| CLI-INV-10 | Commands outside a repository fail before opening product state | cli_test.go::TestRunStopsWhenRepositoryRootIsMissing |
