# VD-sdd-m3-git-binding-2026-11-19: How tenant commits bind to sdd change intents, and how review artifacts wait for what they cite

**Status:** Accepted
**Date:** 2026-11-19

**Context:** R2 M3 lets a tenant commit carry `Intent: sdd:<change-name>` so the unchanged
intent-delivery-readiness gate (`BR-sdd-warranty-i9miv7` O-8) can judge it. Seven non-author review
rounds over the M3 SPEC (`docs/plans/R2-sdd-m3-git-binding-spec-draft5.md`, as amended by drafts 6-8;
ACCEPTABLE at round 7, `r2-m3-spec-review-round7-20261119`) showed two hazards:
- A claim, verdict or review that reaches the ledger ahead of the record it cites makes every later
  projection rebuild fail permanently, because the ledger is append-only and reduced in sequence order.
- Under M3, "claim not yet bound" is a normal state for correct tenant data.

**Decision:** The owner took the recommended option for each of six decisions on 2026-11-19.
- **D1 — token:** the whole trailer value matches `^sdd:[a-z0-9][a-z0-9-]{1,99}$`. A value whose first
  four bytes case-fold to `sdd:` but do not match that grammar is a `malformed` claim skip, never a sync
  failure.
- **D2 — unbindable claims:** handled per commit, all-or-nothing. If any `sdd` claim on a commit cannot
  bind, the commit event carries no refs at all, a `ClaimSkip` is reported (not ledgered), and the sync
  succeeds.
- **D3' — claim ordering:** a claim binds only if its exact `(change-intent id, digest)` is already
  ledgered. Otherwise it is skipped as `not-ledgered`. Recovery is `sync intent sdd`, then `sync git`.
- **D4 — the pre-existing `records`/`specdir` claim-ordering hazard:** deferred
  (`DEF-INTENT-GIT-ORDERING-001`), and fixed after M3 with the same mechanism.
- **D5 — first revisions of review artifacts:** the reviews connector **holds** the first revision of a
  v2 verdict or requirement review until its order-only prerequisites are ledgered: its exact commit
  claim and its requirement revisions. Evidence ids are not prerequisites, because they are local to
  one ledger.
- **D6 — later revisions:** a later revision whose prerequisites are missing is **refused loudly**. It
  is not appended, the sync processes the rest, and then the sync fails, naming it. Every artifact
  whose `(schema, native id)` is listed more than once in one sync is refused as `duplicate-identity`.

**Consequences:**
- **What this makes easy:**
  - No correct-data path appends a claim or a review artifact ahead of what it cites, so rebuilds stay
    green.
  - A held first revision adds no warrant.
  - A refused revision makes every sync-first enforcement path red, so it cannot pass silently there.
  - The mechanism is the same for claims and reviews, and it is ready to extend to
    `DEF-INTENT-GIT-ORDERING-001`.
- **What this makes hard:**
  - On a long-lived ledger, `sync all` and delivery-enforce stay red after an in-place revision until
    `sync intent sdd` and `sync git` have run.
  - A refused revision whose prerequisite will never be ledgered stays red until the artifact is fixed.
  - A canary or enforce run without a prior sync still reads the old outcome.
- **What we gave up:**
  - A dangling-tolerant projection (rejected: it hides real errors).
  - Holding revisions quietly (rejected: it is fail-open).
  - Appending revisions unconditionally (rejected: it is a permanent wedge).
  - Bindings are monotone, so a binding made by an earlier kernel keeps warranting its commit.
- **Not addressed here:** pre-existing stale verdict outcomes (`DEF-VERDICT-STALE-OUTCOME-001`) and
  evidence portability (`DEF-VERDICT-EVIDENCE-PORTABILITY-001`).

**Revisit when:** any change narrows `sdd` binding (M2's Spec Kit rules, a `Resolve` outcome change, a
stricter anti-spoofing check), which must state whether existing bindings stay; or
`DEF-INTENT-GIT-ORDERING-001` or `DEF-VERDICT-STALE-OUTCOME-001` is taken up; or a tenant reports a
refusal that the documented recovery does not clear.
