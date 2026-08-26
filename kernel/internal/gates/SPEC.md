# `internal/gates` SPEC

The gates package evaluates versioned gate definitions against the append-only
event ledger. P2 begins in canary mode: a failing condition is reported as
`BLOCKED` with ledger proof, but no command is refused and no ledger row is
written by the evaluator.

## Definition format

Gate files use the JSON subset of YAML so the repository needs no new parser
dependency. They carry a `.yaml` extension and must contain:

```json
{
  "schema": "vera.gate.v1",
  "id": "make-check-success",
  "description": "The latest make check witness succeeded",
  "expires": "2026-10-16",
  "mode": "canary",
  "source": "checks",
  "kind": "check.run",
  "selector": {"field": "command", "equals": "make check"},
  "condition": {"field": "exit_code", "equals": 0}
}
```

The evaluator selects the highest-sequence event matching `source`, `kind`, and
the optional `selector`,
reads the named top-level JSON payload field, and compares it with `equals`.
`condition.all` is also supported for a non-empty list of field predicates; all
predicates must match the same event.
Missing events produce `UNKNOWN`; a matching value produces `PASS`; another
value produces `BLOCKED`. Every non-UNKNOWN result retains event ID and seq.

## Invariants

1. **GATE-INV-1 — Definitions require the closed v1 schema and an explicit canary or enforce mode.**
2. **GATE-INV-2 — The latest matching ledger event determines the result.**
3. **GATE-INV-3 — PASS and BLOCKED results retain event proof; no event is UNKNOWN.**
4. **GATE-INV-4 — Canary evaluation does not mutate the ledger.**
5. **GATE-INV-5 — Enforcement requires explicit mode promotion.**
6. **GATE-INV-6 — Enforcement rejects BLOCKED and UNKNOWN results.**
7. **GATE-INV-7 — Enforcement rejects an empty definition set.**
8. **GATE-INV-8 — Compound conditions are conjunctive.**
9. **GATE-INV-9 — Selectors isolate event streams.**

| Invariant | Statement | Proving test |
|---|---|---|
| GATE-INV-1 | Definitions require the closed v1 schema and an explicit canary or enforce mode | gates_test.go::TestLoadRejectsInvalidDefinitions |
| GATE-INV-2 | The latest matching ledger event determines the result | gates_integration_test.go::TestEvaluateUsesLatestMatchingEvent |
| GATE-INV-3 | PASS and BLOCKED results retain event proof; no event is UNKNOWN | gates_integration_test.go::TestEvaluateStatesAndProof |
| GATE-INV-4 | Canary evaluation does not mutate the ledger | gates_integration_test.go::TestEvaluateIsReadOnly |
| GATE-INV-5 | Enforcement requires explicit mode promotion | gates_test.go::TestEnforceRequiresPromotion |
| GATE-INV-6 | Enforcement rejects BLOCKED and UNKNOWN results | gates_test.go::TestEnforceRejectsNonPass |
| GATE-INV-7 | Enforcement rejects an empty definition set | gates_test.go::TestRequireDefinitions |
| GATE-INV-8 | Compound conditions are conjunctive | gates_test.go::TestEvaluatePayloadRequiresAllPredicates |
| GATE-INV-9 | Selectors isolate event streams | gates_integration_test.go::TestLoadedIndexGateMatchesCommandAndExitCode |
