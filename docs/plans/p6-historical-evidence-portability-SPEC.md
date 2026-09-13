# P6 historical-evidence portability SPEC

## Purpose and boundary

This is a migration-only seam for the exact ledger chain needed by the two events cited by the
accepted P5 obligation verdict. It repairs source portability for a fresh clone; it is not a
connector, a new event source, or a relaxation of projection integrity. Existing wire identities,
event kinds, pinned vectors, and ordinary sync/verify/report behavior remain unchanged.

The committed archive is `docs/verification/p6-historical-evidence.jsonl`. It contains exactly five
JSON-lines records: the three exact P5 intent prerequisites at sequences 1227–1229, followed by
sequence 1670 with event ID `01M28TPW9C8R7ND19MNDCJ9GDG` and sequence 1834 with event ID
`01M29HMPE5V977AR3VMW47DVDE`, the two evidence events cited by the P5 verdict. The prerequisite
records are the minimal transitive closure required for the cited commit's intent reference to be
valid in ledger order. All records are original envelopes, including sequence, event identity,
timestamps, payload, content hash, and connector version.

## Operator contract

The only entry point is the explicit command:

```text
go run ./cmd/proofbound migrate historical-evidence
```

It resolves the archive from the repository root and imports it before the first `sync all` on a
fresh local ledger. The command refuses any non-empty ledger, validates the complete fixed archive,
and writes through the store's append-only replay seam. It is not invoked by `sync`, `verify`,
`rebuild`, reports, or gates. A second invocation is refused because the ledger is no longer empty.

## Archive validation

Parsing is strict: exactly five non-empty lines, one object per line, no unknown envelope/event
fields, valid canonical event payloads, valid event IDs, ascending fixed sequence numbers, and the
fixed source/kind/native-ID/content-hash/connector identity for each cited event. The implementation
also binds each complete line to its expected SHA-256, so changing any byte, including an envelope
field or payload representation, is rejected. Missing, additional, reordered, or duplicate records
are rejected.

The target ledger must be empty before import. Import is atomic, append-only, and uses the exact
stored sequence numbers. No event is re-minted and no existing event is rewritten. A dangling
reference remains invalid unless its exact cited envelope is present.

## Invariants and tests

| ID | Invariant | Proving test |
|---|---|---|
| HE-INV-1 | The committed archive has exactly the five-record cited-chain closure and exact line digests | `migration_test.go::TestLoadCommittedArchive` |
| HE-INV-2 | Unknown fields, malformed JSON, altered bytes, missing/additional/reordered/duplicate records fail closed | `migration_test.go::TestParseRejectsArchiveMutations` |
| HE-INV-3 | Import refuses a non-empty ledger and never changes it | `migration_test.go::TestImportRefusesNonEmptyLedger` |
| HE-INV-4 | Import of the exact archive preserves sequence and event identity and is atomic | `migration_test.go::TestImportExactArchive` |
| HE-INV-5 | The CLI exposes only the explicit migration command; ordinary command parsing does not invoke it | `cli_test.go::TestParseHistoricalEvidenceMigration` |

## Acceptance evidence

Fresh Linux and native Windows records run the migration command before `sync all`, then record both
sync passes, `proofbound verify`, rebuild equality, and self-hosted reports. The archive's purpose is
limited to closing the previously observed historical dangling reference; it does not waive any
other C8 requirement.

## Non-goals

- No broad ledger export/import facility.
- No import of arbitrary user-supplied envelopes, current connector output, or other historical IDs.
- No schema, route, event-kind, wire-identity, vector, provider, platform, or product-capability change.
- No weakening of failed-closed dangling-reference validation.
