# internal/migration — SPEC

This package owns the one explicit P6 historical-evidence migration. Its only input is the committed
`docs/verification/p6-historical-evidence.jsonl` archive; it is not a connector and does not widen
the event universe. The governing contract is
[`p6-historical-evidence-portability-SPEC.md`](../../../docs/plans/p6-historical-evidence-portability-SPEC.md).

| Invariant | Statement | Proving test |
|---|---|---|
| HE-INV-1 | The archive has exactly the five-record fixed chain closure and exact line digests | migration_test.go::TestLoadCommittedArchive |
| HE-INV-2 | Altered, missing, additional, reordered, duplicate, malformed, and unknown-field records fail closed | migration_test.go::TestParseRejectsArchiveMutations |
| HE-INV-3 | A non-empty target ledger is refused without mutation | migration_test.go::TestImportRefusesNonEmptyLedger |
| HE-INV-4 | Exact import preserves all fixed sequences and event identities | migration_test.go::TestImportExactArchive |

The migration command is routed by the CLI's `TestParseCommand` table; normal sync and verification
commands do not invoke this package.
