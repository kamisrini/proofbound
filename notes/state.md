# VERA — Current State

> THE resume note. Overwrite-in-place; never append; never a second copy. A fresh session reads this first.
> Kept live under **Law 10** — `make check` fails when HEAD moves >3 commits past its last update.
> The judgement below is hand-written; the imported `make state` target is currently absent.

**As of:** 2026-08-26 (P1 closed; P2 gates-as-data in progress)

## Resume Prompt

Copy this verbatim into a fresh session started in this repository:

```
Read `CLAUDE.md` explicitly, then `notes/state.md` and `notes/journal/2026-08-26.md`.

Load-bearing facts:
- The GitHub remote is `git@github.com:kamisrini/proofbound.git`; branch `main` is pushed and clean.
- The P1 Go scaffold/core/store work is committed. Store supports migrations, embedded/external DB
  configuration, append/read, transactions, and Docker-backed integration tests.
- PostgreSQL test infrastructure is available locally: Docker image `postgres:16-alpine` was used
  successfully on 2026-08-25. Start a disposable container named `proofbound-task6-postgres` with
  host port `55433` mapped to 5432 and use
  `postgres://postgres:postgres@127.0.0.1:55433/vera?sslmode=disable`; Docker inspection/run may
  require escalated host access. Check this image before claiming PostgreSQL is unavailable.
- The mutation harness is calibrated and supports `MUTANT_TEST_TAGS=integration` with `DATABASE_URL`.
- The post-`dc94712` database-aware store acceptance sweep found 107 candidates, 107 killed,
  0 invalid, and 0 survivors. Calibration passed and no store allowlist entries were needed.
- Store mutation acceptance is recorded in `notes/journal/2026-08-24.md` and pushed at `89cfa30`.
- Task 4 is ACCEPTED under Law 9. The Round 4 verdict independently accepted frozen `d794ff7` with
  no HIGH, MED, or LOW findings and is committed verbatim at
  `docs/verification/verdicts/task4-current-round4.md`. Author evidence is gitcmd 77/77 killed and
  combined parent/child 95/95 killed, with no invalids or survivors.
- Task 5 Rounds 1–6 returned NEEDS_WORK and are committed verbatim under
  `docs/verification/verdicts/task5-current-round{1,2,3,4,5,6}.md`. Round 1 remediation rejects null evidence and
  typed-nil appenders, binds gate/Git to one root, strips all `GIT_*`, fails repository observation
  before the gate, exercises the real Make target at mode 0644, and JSON-escapes control bytes.
  Round 2 remediation passes the sanitized environment into a real Git-using gate and refuses NUL
  or invalid UTF-8 before the gate/witness and before Go decoding/content identity.
  Round 3 remediation makes hash, serialization, publication, and capture cleanup failures loud.
  `make check-witnessed` emits strict
  `vera.witness.v1` evidence without invoking VERA; `vera sync checks` ingests and deduplicates it.
  Post-remediation connector mutation is 37/37 killed; DB-aware CLI mutation is 17/17 killed, with
  no invalids or survivors. Task 5 Round 6 remediation is committed at `517470e`; Round 7 independently
  returned ACCEPTABLE and is committed at `docs/verification/verdicts/task5-current-round7.md`.
- Task 8 `vera report week` is accepted under Law 9. Frozen remediation is `327219c`; the independent
  Round 1 verdict is committed verbatim at `docs/verification/verdicts/task8-current-round1.md`. Evidence
  is in `docs/verification/task8-final-evidence.md`.
- Task 9 closed P1: review verdict artifacts use `vera.verdict.v1` metadata, `vera sync reviews` ingests
  committed artifacts, `reviews_view` retains finding/event proof, and `vera report week` exposes the
  ledger-ordered red-verdict/change/next-verdict chain. The spec-first graduation is a blocking Go test
  under `make check` at `kernel/internal/specfirst`.
- P2 first slice is landed: `gates/make-check-success.yaml` and `vera gates canary` evaluate the
  latest `check.run` payload against the ledger and retain event proof. Gate definitions carry an
  ISO expiry date; `vera gates enforce` requires explicit `mode: enforce` promotion, rejects expired
  definitions, and fails closed on BLOCKED/UNKNOWN. Remaining P0 checks are not yet migrated and
  the initial gate remains canary-only.
- The next P2 migration adds `make index-check-witnessed` and a distinct `index-check-success` gate;
  its compound condition requires both the witnessed command and `exit_code: 0`.
- The current P2 migration adds `make law-citation-witnessed` and `law-citation-success` with the
  same command-plus-exit-code proof requirement.
- The next P2 migration adds `make spec-numbering-witnessed` and `spec-numbering-success` with the
  same command-plus-exit-code proof requirement.
- The current P2 migration adds `make invariant-table-witnessed` and `invariant-table-success` with
  the same command-plus-exit-code proof requirement.
- The tagged mutation mechanism now serializes integration packages and uses a tested 30-second
  ceiling. Its earlier concurrent form let packages contaminate one database; its fixed 10-second
  ceiling could misclassify timeout as a kill. Both mechanism routes have a self-test.
- THE CENTRAL FINDING, which outranks any fix list: hand-iterating fixes does NOT converge.
  My fix rate and my defect-introduction rate are roughly equal, and being MORE careful did
  not change it. What converges is MECHANISM — script+self-test classes have recurred zero
  times; prose-answered classes recurred 3-5x.
- The imported corpus DESCRIBES skip-lint, prescription-lint, state-freshness, invariant-lint,
  spec-numbering-lint and other gates that are absent from this checkout. Current `make check`
  runs only the scripts actually present plus kernel build/test/lint. Do not cite historical gate
  claims as current enforcement.
- Accepted Task 4 author sweeps are GREEN: gitcmd 77/77 killed; combined connector target 95/95
  killed, with no invalids or survivors. Do NOT read that as "git is proven": the operator set is small,
  never mutates `_test.go`, and says nothing about whether a test asserts the RIGHT thing.
- Fixing the REPRODUCTION is not fixing the INVARIANT. Before closing anything: write the
  HARM in the finding's own words, then enumerate ROUTES to it.
- A verified FINDING does not make its proposed FIX verified. Round 2's remedy for the
  citation defect would have broken a real decision id; check remedies against real data.
- A "cannot/never/always" in a comment is a claim needing a test. Two of mine were false.
- Law numbers and SPEC invariant numbers are PERMANENT: APPEND, never insert. `make invariants-lock`,
  its blocking drift detector, and its self-test are restored. Proving-test cells now have a
  blocking shape check, but historical file/test resolution remains migration debt.
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

Next-session sequence: run `git status --short --branch && git log -1 --oneline`, then reread
`CLAUDE.md`, `notes/state.md`, and `notes/journal/2026-08-26.md`. Continue P2 from the ROADMAP P2
status and migrate the next P0 check only after adding its gate and canary test. `make check` BARE is
the verification gate. `make backup` is currently missing;
create a manual `git bundle --all` until the target and its self-test are restored.
```

## Worktree

- **Branch:** `main`
- **HEAD:** use `git log -1 --oneline`; the imported `make state` target is absent
- **Remote:** `origin` → `git@github.com:kamisrini/proofbound.git`; push after each coherent commit and keep a local bundle backup.
- **Uncommitted:** none expected; commit cadence is enforced at 90m

## What this session shipped

This repository migration began from a documentation corpus. P1 Tasks 0–9 are accepted and P1 is
closed under Law 9.

**Migration warning:** the imported docs describe a richer enforcement suite than this repository
currently contains. Present mechanisms are the scripts visible under `scripts/`, their visible
self-tests, kernel build/test/lint, and the calibrated mutation harness. Missing targets/scripts are
mechanism debt, not silently inherited evidence.

**Store work:** append/read and transaction paths are implemented. The post-`dc94712` database-aware
mutation sweep is accepted at 107 candidates, 107 killed, 0 invalid, and 0 survivors; calibration
passed and the allowed-survivor ledger needed no store entry. Full reasoning and routes are in
`notes/journal/2026-08-24.md`.

## Accepted work

The pure connector and gitcmd adapter are implemented from restored package-specific SPECs. Rounds
1–2 are closed. Round 3 found wrong-repository ingestion through inherited Git environment,
configuration-dependent gitlink paths, and an annotated-tag object accepted as detached HEAD; those
routes now have code/spec remedies and real-adapter tests. G-INV-19 through G-INV-21 were appended.
gitcmd is 77/77 killed; combined parent+child is 95/95 killed. No survivor was declared.

Round 4 independently closed the full route classes and returned ACCEPTABLE. Task 4 is complete.

Task 5 is accepted by its Round 7 verdict. Task 6 is accepted by its Round 6 verdict, with
83/83 mutants killed and no invalids or survivors. Task 7 is accepted by its Round 1 verdict,
with 62/62 session mutants and 99/99 projection mutants killed and no invalids or survivors.
The final Task 7 evidence also includes race, PostgreSQL integration, fresh-schema verification,
and the full `make check` gate.

Task 8 is accepted by its independent Round 1 verdict. Its final remediation adds fail-closed proof
validation to the weekly report; the final evidence includes PostgreSQL report integration, lint, and
the full `make check` gate.

## Blockers / open, in priority order

1. **P1 closed:** Tasks 5–9 are accepted under Law 9. Task 9’s review connector, projection chain,
   full verifier evidence, and blocking spec-first gate are recorded in the 2026-08-26 journal.

2. **P1 forward path (validated 2026-08-14 — full audit in the plan's Position section):**
   Tasks 0–9 DONE · P1 closed 2026-08-26. The planned `/vera-wrap` step-3 amendment cannot be
   applied because `.claude/commands/vera-wrap.md` is absent from this checkout.
3. **Retroactive mutation sweeps:** store is accepted; current Git implementation is green
   author-side; `core` remains queued.
4. **Owed mechanisms, validated and ranked:** restore the mechanisms the imported constitution and
   gates registry claim are live but this checkout lacks: `make backup` + self-test; a real
   `make short`; state-freshness; full invariant citation resolution; skip-lint; prescription-lint.
   Invariant numbering and proving-test-cell shape are now blocking and self-tested. Then reassess
   the historical comment-claims and cleanroom-lint debts against the code that actually exists.
5. **Calendar (none urgent):** spec-first graduated to blocking on 2026-08-26;
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

## End-of-day 2026-08-26 — Task 9 / P1 close

- Branch `main` is clean and pushed to `origin`; the final durability commit is the current HEAD.
- Today’s completed work is Task 9 implementation, review-artifact migration, spec-first graduation,
  clean-database verification, independent acceptance, P1 evidence, journal, and state updates.
- Unverified assumptions: no real Claude session JSONL corpus was present in the expected session-artifact
  directory; Task 7’s ingestion behavior was verified with synthetic fixtures only, as designed.
- Next exact first action: run `git status --short --branch && git log -1 --oneline`, reread
  `CLAUDE.md`, this file, and `notes/journal/2026-08-26.md`; then inspect the P2 plan.
- Institutionalized lightweight improvements: when adding a SPEC invariant, use the existing
  `file_test.go::TestName` citation form, append the invariant number, and run `make invariants-lock`
  before `make check`; for PostgreSQL-backed mutation runs, confirm the disposable Docker image and
  `DATABASE_URL` first, and inspect the retained mutant log when calibration fails. The EOD protocol
  now requires the verbatim Resume Prompt itself to contain the complete next-session sequence;
  tomorrow’s instructions may not exist only in a separate closeout section.
