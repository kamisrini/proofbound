# P6 Task 6 — measurements and falsifiers SPEC

## Scope

Task 6 records the nine P5 measurements and seven P5 falsifiers against the evidence that actually
exists at P6. It does not manufacture historical elapsed time, review history, or ledger facts that
were never recorded. Results are maintained in `docs/plans/p6-task6-results.tsv` and rendered by
`scripts/p6-task6-results.sh` into `docs/verification/p6-measurements-falsifiers.md`.

This is an evidence and decision packet only. It adds no capability, provider, event kind, gate, or
new P5 measurement. The packet distinguishes a measured result from a datum that cannot be
recovered and from a falsifier whose required longitudinal sample is not yet available.

## Closed result vocabulary

Every measurement result is exactly `measured`, `not-recoverable`, or `not-applicable`. Every
falsifier result is exactly `false`, `fired`, or `indeterminate`. Each row has a nonempty value,
typed provenance, and a next action or explicit no-action statement. A `fired` falsifier must name
the required decision or action; an `indeterminate` falsifier must name the missing sample or datum
and the action needed to resolve it.

Allowed provenance is explicit and typed:

- `artifact:<tracked path>` for a repository artifact;
- `command:<command text>` for a repository command result;
- `session:<tracked dated journal>` for human-time or unavailable historical facts; and
- `event:<ledger event ID>` only when the ledger could have observed the datum.

The checker resolves every artifact and session path and validates event-ID shape. It does not claim
that a command string was run unless the generated packet separately cites the resulting artifact.

## Invariants and test derivation

| ID | Statement | Proving test |
|---|---|---|
| T6-INV-1 | Exactly the nine ratified measurements and seven falsifiers appear once | p6-task6-results.test.sh::complete-result-universe |
| T6-INV-2 | Result values use only the closed measurement/falsifier registries | p6-task6-results.test.sh::closed-result-vocabulary |
| T6-INV-3 | Every row has typed, resolving provenance and an action | p6-task6-results.test.sh::typed-provenance-and-action |
| T6-INV-4 | Non-recoverable and indeterminate claims state the missing datum and next action | p6-task6-results.test.sh::honest-uncertainty |
| T6-INV-5 | Generated output is deterministic and stale output fails closed | p6-task6-results.test.sh::generated-freshness |

## Acceptance

```text
scripts/tests/p6-task6-results.test.sh
scripts/p6-task6-results.sh --check
scripts/p6-census.sh --check
bare make check
```

Task 6 closes only when the generated packet contains exactly 9 measurement rows and 7 falsifier
rows, every row has allowed provenance and a result, every `fired` or `indeterminate` falsifier has
its required action, and all C7 census rows cite the packet. No row is closed by the phrase
“evaluated honestly” alone.

## Non-goals

- No retroactive ledger events, invented timings, or inferred human decisions.
- No new provider, feed, event kind, acceptance bar, or P7+ capability.
- No mutation sweep or final package verdict; those remain later P6 tasks.
