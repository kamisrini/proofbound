# P6 Task 5 — requirement-review completion SPEC

## Scope

Task 5 closes the C6 requirement-review obligation for the providers configured by Proofbound's
production CLI. It audits the committed requirement revisions currently visible to those providers
and binds every active obligation to an exact-revision, non-author requirement-review artifact.
This is a consolidation audit: it adds no provider, event kind, review outcome, or enforcement
capability.

The configured provider set is exactly `records` and `specdir`, as wired by
`kernel/internal/cli/cli.go`. The audit reads committed paths only. Native records are found under
`docs/intent/records/requirements/<BR-id>/<revision>.md`; a configured provider with no active
committed requirement artifact is reported with an explicit zero count rather than silently omitted.

## Closure contract

The checker is `scripts/p6-task5-review.sh` and its generated result is
`docs/verification/p6-requirement-review.md`.

For every active requirement revision, the checker proves:

1. the artifact path, schema, ID, owner, status, and self-digest are internally consistent;
2. every active obligation appears in the requirement artifact;
3. exactly one current review artifact binds the requirement's source, ID, and artifact digest;
4. the review artifact's own path and self-digest resolve, its declared reviewer is distinct from
   the requirement owner, and every active obligation is marked `VERIFIABLE`; and
5. the exact self-hosted chain is present and renders `state=SATISFIED` in the intent-coverage
   packet.

The checker fails closed on a missing provider wiring, malformed active requirement, duplicate
matching review, wrong revision digest, reviewer/owner equality, non-`VERIFIABLE` outcome, missing
obligation, stale closure artifact, or missing chain proof. Documentary verdicts remain distinct
from schema-bearing requirement reviews; the existing reviews connector and projection tests own
their wire and event semantics.

## Invariants and test derivation

| ID | Statement | Proving test |
|---|---|---|
| T5-INV-1 | The configured provider set is explicit and every active requirement revision is enumerated | p6-task5-review.test.sh::provider-and-revision-census |
| T5-INV-2 | Active requirements expose exact active obligation IDs and a valid normalized artifact digest | p6-task5-review.test.sh::requirement-digest-and-obligation-census |
| T5-INV-3 | Each active revision has exactly one exact-revision review with a distinct declared reviewer | p6-task5-review.test.sh::missing-review-and-author-reviewer |
| T5-INV-4 | Every active obligation is covered by a `VERIFIABLE` review outcome | p6-task5-review.test.sh::non-verifiable-and-missing-obligation |
| T5-INV-5 | The current self-hosted intent chain is evidenced by the exact intent-coverage packet | p6-task5-review.test.sh::chain-proof |
| T5-INV-6 | The generated closure packet is deterministic and stale output fails closed | p6-task5-review.test.sh::generated-freshness |

## Acceptance

The Task 5 closure packet is accepted only when all of the following pass:

```text
scripts/tests/p6-task5-review.test.sh
scripts/p6-task5-review.sh --check
go test ./internal/connector/reviews ./internal/projections -run 'RequirementReview|Intent' -count=1
scripts/p6-census.sh --check
bare make check
```

The packet records the active requirement and obligation counts, exact revision and review
references, chain evidence, and these command names. A zero active-requirement result is valid only
when the configured providers have been scanned and the packet explicitly says so.

## Non-goals

- No new provider, requirement, obligation, review schema, event kind, or gate.
- No review authentication claim; reviewer independence is declared and exact owner/reviewer
  equality is rejected as required by the existing contract.
- No retroactive intent claim for historical commits and no change to the ratified applicability
  rule.
- No mutation sweep or final package acceptance; those remain later P6 Tasks 6 and 8.
