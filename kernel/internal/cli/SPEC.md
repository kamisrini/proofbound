# CLI composition contract

## 1. Boundary

This package is Proofbound's reusable command composition root. It parses commands, opens the
ledger, constructs connectors and projections, and renders command results. Executable packages
under `cmd/` are deliberately tiny: `cmd/proofbound` is the live entry point and `cmd/vera` is a
one-phase compatibility alias only.

The package owns no semantic event contract. Frozen wire identities remain owned by their connector
or replay SPECs.

## 2. Interface lock

```go
func Run(ctx context.Context, program string, args []string, stdout, stderr io.Writer) int
```

`program` is either `proofbound` or the deprecated alias. Unknown command forms return status 2 and
print usage for the live name. Runtime failures return status 1 and are prefixed with `proofbound:`.

## 3. Identity and compatibility

- Local state is rooted only at `<repository>/.proofbound`. The one-time migration is an explicit,
  fail-safe directory move documented in `docs/proofbound-identity-migration.md`; commands never
  merge two state directories or silently fall back to the old path.
- Live environment names use `PROOFBOUND_*`. For one phase, each shipped legacy `VERA_*` name is
  honored only when its live equivalent is unset. The live name wins when both exist. Consuming a
  legacy name emits the dated deprecation advisory to stderr.
- `cmd/vera` remains callable through the same `Run` function for one phase and emits the same
  advisory before command execution. It must not fork a second implementation.
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
| CLI-INV-3 | A live environment value wins over its deprecated alias | cli_test.go::TestProductEnvPrefersProofbound |
| CLI-INV-4 | A deprecated environment alias remains functional and emits a dated advisory | cli_test.go::TestProductEnvLegacyAliasWarns |
| CLI-INV-5 | The deprecated executable alias delegates to the same command implementation and warns | cli_test.go::TestRunLegacyProgramWarns |
| CLI-INV-6 | Frozen witness bytes remain accepted without reinterpretation | sync_integration_test.go::TestSyncChecksRebuildAndVerify |

