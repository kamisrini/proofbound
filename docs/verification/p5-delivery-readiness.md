# P5 controlled delivery-readiness evidence

**Captured:** 2026-09-11  
**Controlled commit:** `01f41d608a7d7cd2c99c538555c2ef4891fa405c`  
**Expected result:** BLOCKED while the self-hosted P5 obligations have no accepted v2 verdict and
no all-VERIFIABLE non-author requirement review.

This is a negative control at Proofbound's own delivery boundary. It does not claim that a failed
check witness or a malformed reference caused the stop. The exact P5 BD/BR/CI chain and commit
claim were already ingestible; the readiness join deliberately lacked the two acceptance artifacts
that Task 8 had not yet received.

## Wiring under test

`gates/intent-delivery-readiness.yaml` is explicitly `mode: enforce` and scopes the frozen
reviewed implementation commit `c29bb3b78899f54588f9cbad6e6ecf50333d787c`.
`scripts/delivery-enforce.sh` refreshes every required witness, runs `proofbound sync all`, and only
then runs `proofbound gates enforce`. The CLI resolves `HEAD` to an exact commit, selects only
promoted definitions for enforcement, and makes readiness PASS only when every exact target has
both a `SATISFIED` v2 obligation verdict and a `VERIFIABLE` requirement review.

The script self-test separately proves that a gate failure propagates as a non-zero delivery result,
that complete-chain sync precedes enforcement, and that the serialization lock is cleaned after the
block:

```text
scripts/tests/delivery-enforce.test.sh
```

The database integration test
`kernel/internal/gates/gates_integration_test.go::TestIntentDeliveryReadiness` proves both sides of
the semantic join: the incomplete controlled commit is BLOCKED, then the same commit becomes PASS
only after exact satisfaction and review rows exist.

## Real boundary reproduction

From the repository root, with the self-hosted chain intentionally incomplete:

```text
PATH=/home/thamm/go/bin:$PATH make delivery-enforce
```

The target refreshed the witnessed checks (including the complete bare aggregate test/lint run),
synced all local providers, and returned non-zero at enforcement. A direct repeat of the final
command against the same ledger produced:

```text
gate=index-check-success state=PASS seq=1664 proof=01M28TPW8NPX4RG3061QKKT59Q would_block=false
gate=intent-delivery-readiness state=BLOCKED seq=1509 proof=01M25Y911D9TJ823DZPVF75DH1 would_block=true
gate=invariant-table-success state=PASS seq=1667 proof=01M28TPW9203E27KA27XFGX41B would_block=false
gate=kernel-check-success state=PASS seq=1669 proof=01M28TPW99PBYATTFS2CV5HJT9 would_block=false
gate=law-citation-success state=PASS seq=1665 proof=01M28TPW8V4Y38CPQDFHVS5BFR would_block=false
gate=link-success state=PASS seq=1668 proof=01M28TPW95H0NJCQKYN85D2A0F would_block=false
gate=make-check-success state=PASS seq=1670 proof=01M28TPW9C8R7ND19MNDCJ9GDG would_block=false
gate=spec-numbering-success state=PASS seq=1666 proof=01M28TPW8ZA5T6JV8R3AJGH2G1 would_block=false
proofbound: gate intent-delivery-readiness is BLOCKED
exit status 1
```

Thus every ordinary promoted delivery prerequisite was green, while the readiness rule alone
stopped the controlled boundary and retained the exact commit-event proof. External-consumer
posture is unchanged: promotion is claimed only for Proofbound's own `make delivery-enforce`.

After the independent requirement review and obligation verdict were committed verbatim, the
same `PATH=/home/thamm/go/bin:$PATH make delivery-enforce` boundary was rerun. It returned zero;
`intent-delivery-readiness` was `PASS` at sequence 1834 with proof event
`01M29HMPE5V977AR3VMW47DVDE`, and index-check, invariant-table, kernel-check, law-citation,
link, make-check, and spec-numbering all returned `PASS`. Readiness is therefore bound to the
exact reviewed implementation commit rather than the later artifact-receipt tip.
