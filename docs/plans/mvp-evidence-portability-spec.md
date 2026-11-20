# MVP fix: portable verdict evidence — spec

**Status:** ACCEPTED at draft 2 (round 1 NEEDS_WORK: F1-F4 MED answered, F5-F6 taken, F7 no change; round 2 ACCEPTABLE, 2 LOW recorded). Spec gets at most 2 non-author rounds, code 1 (Law 11). **Authorized:** the owner
named `DEF-VERDICT-EVIDENCE-PORTABILITY-001` on 2026-11-20 ("go with recommendation"). **Intent:**
`CI-mvp-evidence-portable-6dq3ve` → `BR-vertical-mvp-xc31gp` O-2 (the pull-request check works in CI).

## 1. Payoff

The v0.3.0 Action fails on every reviewed commit (`kamisrini/proofbound-action-test` PR #1):

```
v2 verdict evidence 01K… is dangling
```

A sealed verdict cites its evidence by event id, and event ids are random per ledger, so a CI runner
that builds a fresh ledger with `proofbound sync all` can never resolve them. After this fix, a commit
reviewed and sealed on the reviewer's laptop is `PASS` in `proofbound check` on a fresh clone with an
empty ledger, so PR #1 goes green.

## 2. Change

**Portable reference.** The only portable evidence reference is the reviewed commit itself, written
`git:commit.recorded:<reviewed_commit>` (40 lowercase hex). Git commit SHAs are content addresses, so it
means the same thing on every ledger. It is pinned to the verdict's own `reviewed_commit` so no other
event can be cited as evidence (round 1 F2): a verdict can only say "this commit, as the reviewer read
it". `content_sha` is deliberately not part of it: it hashes the connector's payload shape, which
changes with connector versions (`git/2`), and pinning it would make old verdicts dangle after an
upgrade.

1. **Schema (`proofbound.obligation-verdict.v2`, additive).** An obligation outcome keeps its fields and
   may add `evidence_refs`. The nested decoder rejects unknown fields but does not check key order, so
   no order is claimed (F1). `evidence_refs` is absent, `null` (treated as absent) or exactly
   `["git:commit.recorded:<reviewed_commit>"]`; `[]`, any other value, a duplicate and a second entry
   are rejected. A `SATISFIED` outcome needs at least one entry across `evidence_event_ids` and
   `evidence_refs`. The Go field is `omitempty`, so a verdict without it re-marshals to the same payload
   bytes and no existing ledger re-appends an event (F1).
2. **Seal.** `review draft` and `review seal` cite the reviewed commit in `evidence_refs` and write
   `evidence_event_ids: []`. The draft's evidence field is renamed to match and is still rebuilt from
   the ledger, never taken from the editable draft (CLI-INV-20). A draft written by v0.3.0 is refused by
   the strict draft decoder; rerun `review draft`.
3. **Projection.** The reference resolves with `SELECT seq,event_id FROM events WHERE source='git' AND
   kind='commit.recorded' AND native_id=$1 ORDER BY seq LIMIT 1` (several events may share a native id
   because the store dedupes on content, F5). None → `v2 verdict evidence for commit <sha12> is
   dangling`; its sequence after the verdict's → `… is from the future` (R-INV-11).
   `obligation_verdicts_view.evidence_event_ids` keeps its meaning ("events on this ledger"): the sorted,
   de-duplicated union of the legacy ids and the resolved id (F6).
4. **Legacy ids keep their meaning:** they still resolve by event id only, and remain unrestricted as
   before (pre-existing; not widened). A verdict sealed by v0.3.0 still projects on the ledger that
   sealed it and still dangles elsewhere; the remedy is to seal again with v0.3.1.
5. **Version skew (F3).** A v0.3.0 binary rejects `evidence_refs` as an unknown field and the sync
   fails closed with that error. The Action installs the latest release unless pinned, so it picks up
   v0.3.1 by itself; the release notes say that a pinned Action and every reviewer must upgrade together.

## 3. Invariants

| ID | Statement | Proving test |
|---|---|---|
| R-INV-11 (amended) | Every evidence id and the resolved evidence reference name an event at or before the verdict; dangling or future evidence fails projection | projections `TestObligationVerdictProjectionValidation` (new cases: resolved ref, ref with no git event, ref whose git event is after the verdict) |
| R-INV-14 (new) | `evidence_refs` is absent/null or exactly the reviewed commit's reference; `[]`, another commit, a malformed string, a duplicate or a second entry are rejected; SATISFIED with neither list is rejected; a verdict without it re-marshals byte-identically | reviews `TestObligationVerdictEvidenceRefs` (one case per rejection + the payload-identity case) |
| CLI-INV-20 (amended) | seal writes the reviewed commit as the only evidence reference and an empty `evidence_event_ids` | cli `TestReviewSealCitesTheReviewedCommit` (new, positive path) |
| CLI-INV-31 (new) | A commit reviewed and sealed in one clone is `PASS` in `proofbound check` run in a second `git clone` with its own empty ledger after `sync all` | sdd `TestPRCheckOnAFreshLedger` (non-integration-tagged file; calibrated by the code review: with resolution stubbed out to legacy-only, the test must fail) |

## 4. Non-goals

Rewriting or migrating legacy verdicts; changing event identity or `content_sha`; evidence other than
the reviewed commit; holding verdicts for evidence (R-INV-13 still never consults evidence).
