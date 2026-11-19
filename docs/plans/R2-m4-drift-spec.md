# R2 M4 — spec drift — spec

**Status:** owner-approved scope 2026-11-19 ("go"). Law 11 budget: at most 2 spec review rounds,
1 code review.

**Payoff:** `proofbound report drift` on a real OpenSpec project lists every warranted change and
says whether its warranty still covers what the change now promises, and if not, exactly which
obligations were added, removed or reworded since it was verified.

## Why drift compares promises, not digests (measured on OpenSpec's own history)

Requirement records are content-addressed, so an edit never "supersedes" a requirement id (why BR
O-9 was retired). Change intents keep one id per change and gain revisions, but a revision's digest
is a poor drift signal. A probe of all 104 change intents in `Fission-AI/OpenSpec` (pinned commit
`e6539de`) found:

| Revision step | Same promise | Different promise |
|---|---|---|
| Archive move (active → archive) | 60 | 13 |
| Other revisions | 59 | 22 |

Every archive move changes the digest and every requirement id, yet 60 of 73 promise exactly the
same thing. A digest rule would flag all 73; a promise rule flags the 13 that really rewrote
obligations, plus the 22 real edits. So:

**A change intent revision's promise** is the set of `(obligation_id, statement)` over its target
obligations, read from `intent_targets_view` joined to `requirement_obligations_view`.

## The report

For every sdd change intent with at least one **warranty**, meaning an obligation verdict
(projected, so already validated) whose outcomes are all `SATISFIED`:
- **current revision** = the change's revision in the state at HEAD, read from the repository by
  the sdd provider (`StateRevision`, SDD-INV-68), **never from ledger order**. A revert re-emits an
  earlier revision's exact bytes, which the ledger dedupes, so ledger order can name a revision the
  tree no longer has (review `r2-m4-drift-review-20261119` F1, HIGH; BR rev 0005 O-14 replaces O-13).
  If HEAD's revision is not ledgered yet the change is **UNSYNCED**, if the change is gone at HEAD it
  is **REMOVED**, and if it is unmappable at HEAD it is **UNKNOWN**, each with a reason;
- **CURRENT** if any warranty's revision has the same promise as the current revision;
- otherwise **STALE**, diffed against the latest warranty (highest seq): each obligation id is
  `added` (only in the current promise), `removed` (only in the verified one), or `changed` (same id,
  different statement).

```
drift change=<CI id> state=STALE verified_commit=<sha12> verdict=<id> verified_revision=<sha8> current_revision=<sha8>
  changed <obligation id> "<old statement>" -> "<new statement>"
  added <obligation id> "<statement>"
  removed <obligation id> "<statement>"
drift change=<CI id> state=CURRENT verified_commit=<sha12> verdict=<id> current_revision=<sha8>
drift change=<CI id> state=REMOVED|UNSYNCED|UNKNOWN verified_commit=<sha12> verdict=<id> reason="..."
summary warranted=<n> current=<n> stale=<n> other=<n>
```

Changes are listed in CI-id order, obligations in id order, and statements Go-quoted (`strconv.Quote`), so a statement can never break a line.
A STALE warranty clears when a new verdict (on any commit) covers a revision with the current
promise. Exit status is 0: drift is information. The readiness gate is unchanged, since a commit
is still warranted against the exact revision it claimed.

## Invariants (CLI SPEC)

- **CLI-INV-21** drift compares promises: a revision differing only in digest, requirement ids or
  non-promise files is CURRENT; an added, removed or reworded obligation is STALE with exactly that
  diff; a later all-SATISFIED verdict over the current promise clears it; a NEEDS_WORK verdict is
  not a warranty. Proven by a unit test of the pure diff, and end to end in the sdd package
  (verify, then edit a statement, then re-verify, then archive).

## Non-goals and limits

- Only `intent.sdd` changes. `records` and `specdir` intents are out of scope.
- No new event, projection table or gate rule (a drift gate is a later option).
- A warranty here is the verdict alone. The readiness gate also needs a VERIFIABLE requirement
  review, so CURRENT means "what was verified is still what is promised", not "the gate passes".
- Requirement-review drift is not reported. Reviews bind requirement revisions, and a reworded
  obligation already shows as `changed`.
