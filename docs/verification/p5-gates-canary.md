# P5 intent gates — historical canary

Date: 2026-09-10  
Revision: `91cade9 feat: canary intent integrity and readiness gates`

Bare command from `kernel/`:

```text
go run ./cmd/proofbound gates canary
```

Semantic gate results against the migrated local ledger:

```text
gate=intent-delivery-readiness state=UNKNOWN seq=0 proof= would_block=false
gate=intent-reference-integrity state=PASS seq=887 proof=01M25RAKCM0BFW074G93Q7PXPN would_block=false
gate=intent-verdict-integrity state=PASS seq=280 proof=01M14QPNQV1GV90BGA41DEADJ9 would_block=false
```

`UNKNOWN` readiness is the honest historical result: the scoped HEAD had not yet been ingested as
an explicit self-hosting claim. It blocks nothing in canary. The integration acceptance at this
revision also exercised an intentionally dangling claim, which returned proof-bearing `BLOCKED`,
and a complete synthetic chain, which returned `PASS`:

```text
go test -tags=integration ./internal/gates -run '^TestIntent' -count=1
ok github.com/kamisrini/proofbound/kernel/internal/gates
```

The integrity definitions remain canary until non-author independent acceptance. Delivery readiness
remains canary for external consumers; Task 8 wires enforcement only into Proofbound's controlled
delivery boundary.
