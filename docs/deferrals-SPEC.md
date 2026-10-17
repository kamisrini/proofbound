# Deferral register specification

`docs/deferrals.tsv` is the canonical, tracked register for work that is intentionally left open
for later action. Narrative files may summarize a row, but they do not replace it.

## Schema

Every data row has these tab-separated fields, in this order:

| Field | Meaning |
|---|---|
| `id` | Stable identifier, formatted `DEF-...`; never reused |
| `status` | Lifecycle state: `OPEN`, `RESOLVED`, `REJECTED`, or `RETIRED` |
| `hold_type` | Human-facing reason: `DEFERRED`, `POSTPONED`, `AWAITING_OWNER`, `PENDING_PREREQUISITE`, `AWAITING_REVIEW`, `BLOCKED`, or `NONE` |
| `owner` | Person or role responsible for the next decision/action |
| `created` | ISO date the row was opened |
| `phase` | Roadmap phase, such as `P7` |
| `subject` | Short stable name for the deferred item |
| `reason` | Why the item is open |
| `review_trigger` | Event or date that should cause reconsideration |
| `unblock_condition` | Evidence or decision required to resolve the item |
| `evidence` | Comma-separated repository-relative paths supporting the row |

`status` is deliberately canonical for queries. Synonyms belong in `hold_type` and prose, so
`POSTPONED`, `AWAITING_OWNER`, and `PENDING_PREREQUISITE` can all be reported as open without
creating competing meanings for the lifecycle state.

## Lifecycle

- `OPEN`: unresolved and still relevant; this includes postponed or blocked work.
- `RESOLVED`: the unblock condition was met; add the closing evidence and date in the reason or
  supporting record rather than deleting the row.
- `REJECTED`: evaluated and deliberately not accepted.
- `RETIRED`: no longer applicable or superseded; preserve the rationale.

Never use `RESOLVED` for an item that was merely postponed. Never use `REJECTED` when the check was
not performed. An owner deferral is `status=OPEN` with an appropriate `hold_type`.

## Process

1. Add a row before or at the time work is postponed.
2. Give it a stable ID, owner, review trigger, and unblock condition.
3. Record the decision in the appropriate journal/state/decision document and link those paths in
   `evidence`.
4. Run `scripts/deferrals.sh --check` to validate the register.
5. Query with `scripts/deferrals.sh --open`, `--status STATUS`, or `--hold-type HOLD_TYPE`.
6. On completion, update the same row to `RESOLVED`, `REJECTED`, or `RETIRED`; do not delete it.

## Dormant-until-ready policy

An `OPEN` deferral is a dormant record, not an active work queue. The agent must skip it and must
not repeatedly ask, remind, or pressure the owner about it. Resume or work an item only when the
owner explicitly says they are ready, asks to resume it, or names it in a work request. Until then,
preserve its row and evidence, leave its lifecycle status `OPEN`, and continue with other authorized
work. The owner may query the register at any time without activating any row.

The ignored `notes/overall-status.md` may summarize this register for local coordination, but it is
not the source of truth.
