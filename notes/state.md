# VERA — Current State

> THE resume note. Overwrite-in-place; never append; never a second copy. A fresh session reads this first.
> Kept live under **Law 10** — `make check` fails when HEAD moves >3 commits past its last update.
> The judgement below is hand-written; the imported `make state` target is currently absent.

**As of:** 2026-08-24 (store accepted and pushed; Task 4 Git connector implemented with 77/77 combined mutants killed; independent non-author verdict still required)

## Resume Prompt

Copy this verbatim into a fresh session started in this repository:

```
Read `CLAUDE.md` explicitly, then `notes/state.md` and `notes/journal/2026-08-24.md`. Do not start work until you have.

Load-bearing facts:
- The GitHub remote is `git@github.com:kamisrini/proofbound.git`; branch `main` is pushed and clean.
- The P1 Go scaffold/core/store work is committed. Store supports migrations, embedded/external DB
  configuration, append/read, transactions, and Docker-backed integration tests.
- The mutation harness is calibrated and supports `MUTANT_TEST_TAGS=integration` with `DATABASE_URL`.
- The post-`dc94712` database-aware store acceptance sweep found 107 candidates, 107 killed,
  0 invalid, and 0 survivors. Calibration passed and no store allowlist entries were needed.
- Store mutation acceptance is recorded in `notes/journal/2026-08-24.md` and pushed at `89cfa30`.
- Task 4 Round 1 received a committed `NEEDS_WORK` verdict. All six reported routes are remediated
  author-side; gitcmd is 70/70 killed and the combined connector sweep is 88/88 killed, with no
  invalids or survivors. It is NOT accepted until a non-author verifies the frozen remediation and
  returns an ACCEPTABLE verdict committed under `docs/verification/verdicts/`.
- THE CENTRAL FINDING, which outranks any fix list: hand-iterating fixes does NOT converge.
  My fix rate and my defect-introduction rate are roughly equal, and being MORE careful did
  not change it. What converges is MECHANISM — script+self-test classes have recurred zero
  times; prose-answered classes recurred 3-5x.
- The imported corpus DESCRIBES skip-lint, prescription-lint, state-freshness, invariant-lint,
  spec-numbering-lint and other gates that are absent from this checkout. Current `make check`
  runs only the scripts actually present plus kernel build/test/lint. Do not cite historical gate
  claims as current enforcement.
- Current author sweeps are GREEN: gitcmd 60/60 killed; combined connector target 77/77 killed,
  with no invalids or survivors. Do NOT read that as "git is proven": the operator set is small,
  never mutates `_test.go`, and says nothing about whether a test asserts the RIGHT thing.
- Fixing the REPRODUCTION is not fixing the INVARIANT. Before closing anything: write the
  HARM in the finding's own words, then enumerate ROUTES to it.
- A verified FINDING does not make its proposed FIX verified. Round 2's remedy for the
  citation defect would have broken a real decision id; check remedies against real data.
- A "cannot/never/always" in a comment is a claim needing a test. Two of mine were false.
- Law numbers and SPEC invariant numbers are PERMANENT: APPEND, never insert. The imported locks
  exist, but `make invariants-lock` and its detector are absent in this checkout; restore the
  generator+self-test before changing either namespace.
- Two product VDs landed 2026-08-14 and are the citable grounding for Task 5+ design:
  VD-verification-asymmetry-2dyjnd (mechanism beats care — measured) and
  VD-verdicts-are-artifacts-rl0rab (verdicts commit on receipt; evidence binds to content,
  not position).
- LESSON tags now use the six parent classes consolidated in the 08-14 journal — 18 singleton
  tags in two days had quietly defeated the recurrence counter.
- `make short` is intended to be the inner loop, but currently runs no Go tests; do not rely on it
  until its target and self-test are repaired. `make check` is the gate and never takes `-short`.
- macOS trap: t.TempDir() is a SYMLINK. Path-guard mutations under TMPDIR=/private/tmp/vkreal.
- Run make check BARE. Never pipe it through a && chain — the pipe eats the exit code.

Pre-flight: `make check` BARE — expect GREEN. `make backup` is currently missing; create a manual
`git bundle --all` until the target and its self-test are restored.

Run a non-author adversarial review of Task 4 against both connector SPECs and the harm/routes in
`notes/journal/2026-08-24.md`. Commit the verdict on receipt. Fix findings, re-sweep, and repeat until
ACCEPTABLE. Resolve the missing `make backup` mechanism and misleading `make short` target before
relying on either command as evidence.
```

## Worktree

- **Branch:** `main`
- **HEAD:** use `git log -1 --oneline`; the imported `make state` target is absent
- **Remote:** `origin` → `git@github.com:kamisrini/proofbound.git`; push after each coherent commit and keep a local bundle backup.
- **Uncommitted:** none expected; commit cadence is enforced at 90m

## What this session shipped

This repository migration began from a documentation corpus. Store is now accepted; the Git
connector implementation is author-complete and awaits independent verification.

**Migration warning:** the imported docs describe a richer enforcement suite than this repository
currently contains. Present mechanisms are the scripts visible under `scripts/`, their three
self-tests, kernel build/test/lint, and the calibrated mutation harness. Missing targets/scripts are
mechanism debt, not silently inherited evidence.

**Store work:** append/read and transaction paths are implemented. The post-`dc94712` database-aware
mutation sweep is accepted at 107 candidates, 107 killed, 0 invalid, and 0 survivors; calibration
passed and the allowed-survivor ledger needed no store entry. Full reasoning and routes are in
`notes/journal/2026-08-24.md`.

## Where Task 4 stands

The pure connector and gitcmd adapter are implemented from restored package-specific SPECs. Round 1
returned `NEEDS_WORK` for missing-object refs, non-UTF-8 path collapse, nested decision citations,
empty merge file sets, overstated ordering prose, and typed-nil dependencies. Those routes now have
code/spec remedies and discriminating tests. gitcmd is 70/70 killed; combined parent+child is 88/88
killed. No survivor was declared.

None of that is acceptance. The frozen remediation still needs the non-author Round 2 verdict.

## Blockers / open, in priority order

1. **Git connector remains unaccepted:** current Round 1 is committed as `NEEDS_WORK`; its six routes
   are remediated and author gates are green, but a non-author must review the frozen remediation
   and commit the Round 2 ACCEPTABLE or NEEDS_WORK verdict.

2. **P1 forward path (validated 2026-08-14 — full audit in the plan's Position section):**
   Tasks 0–3 DONE · Task 4 = implemented, blocker #1 is its independent acceptance gate;
   the amended bar: mutants green + non-author ACCEPTABLE verdict) · Tasks 5–9 NOT STARTED
   (placeholders verified — no check-witnessed target, no spool schema anywhere) · Task 9 hard
   deadline 2026-09-18 (spec-first advisory expiry), ~5 weeks out — Task 4 must close this week.
   Task 5's design inputs are written: VD-verification-asymmetry-2dyjnd + VD-verdicts-are-artifacts-rl0rab.
3. **Retroactive mutation sweeps:** store is accepted; current Git implementation is green
   author-side; `core` remains queued.
4. **Owed mechanisms, validated and ranked:** restore the mechanisms the imported constitution and
   gates registry claim are live but this checkout lacks: `make backup` + self-test; a real
   `make short`; state-freshness; invariant citation/numbering generators and self-tests; skip-lint;
   prescription-lint. Then reassess the historical comment-claims and cleanroom-lint debts against
   the code that actually exists here.
5. **Calendar (none urgent):** spec-first advisory expires 2026-09-18 = Task 9's deadline;
   meta-tax + backup advisories 2026-10-16 (backup's graduation = mechanize bundle freshness or
   delete the row). No expired advisories; 0.6's stop-check bomb was defused by graduation 08-08.
   No human-owned items pending — remote was settled local-only (bundles).

## Anything brittle

- The imported 2026-08-12 through 2026-08-14 journals/specs describe a prior implementation whose
  code and verdict artifacts were never imported. Treat their findings as design lessons, not as a
  verdict on the current implementation.
- `make mutants` takes ~10 min per package and is NOT in `make check` — deliberately, since a gate
  that slow gets skipped. It is a package-acceptance gate, and `docs/gates.md` says so.
- `docs/allowed-survivors.txt` keys are `line#ordinal` and MOVE whenever the package changes.
  Re-sweep and refresh after any edit; a stale entry licenses a different site than its reason
  describes.
- Keep the pushed GitHub repository and a dated `git bundle` as the two recovery copies.
