# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-13 (P6 Tasks 0–2 complete; Task 3 in progress with C8-001 still blocked)

## Resume — 2026-09-13

The founder ratified dual-platform execution on 2026-09-13: retain Linux and native Windows, using
Git Bash/MSYS2 behind the PowerShell entry point. The receipt, semantic VD, SPEC contract, and
roadmap update are committed. The exact next action is to run the native Windows prerequisite/setup
path from a fresh PowerShell process and complete C8-002 evidence; C8-001's historical-evidence
portability decision remains separate and must be resolved before Task 4.

If the founder explicitly authorizes pushing branch `main` to
`https://github.com/kamisrini/proofbound.git`, test write authentication by performing that push and
report its exact result; do not infer destination approval from read access.

## Branch and repository status

- Branch: `main`.
- The parent of this dated-resume layout update is
  `33bc32b`; `origin/main` remains `1d1112df2622abd54f07837335a7805c3c69145d`. Once this note is
  committed, local is 35 commits ahead and 0 behind.
  ahead and 0 behind.
- Push was not executed. A remote read succeeded, but write authentication was not reached because
  safety review requires explicit destination-specific approval to export these commits to
  `https://github.com/kamisrini/proofbound.git`. Do not report the branch as pushed.
- The only pre-existing untracked path is `docs/plans/P5-intent-provenance-plan.md`, SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`. It is not the accepted P5
  exhibit (SHA-256 `9d203c96243a73cad2cc119662a9b58ea04dc7e9b17d8e5ed3b1a418118d5ffa`)
  and remains untouched pending ownership resolution.

## Completed and verified

- P5 is accepted and must not be restarted. Its ratified plan, semantic VD, founder record,
  adjudication, independent obligation review, mutation evidence, and delivery-readiness evidence
  remain committed.
- P6 is authorized as consolidation/deep completion with no new capability. Task 0's closed scanner
  and 188-row C1–C8 census are durable. Task 7's snapshot provider is decision-skipped to P7+.
- P6 Task 1 is complete: all 34 C1 and 41 C2 rows are closed; operational/lexical gates, hooks,
  exact SPEC citations, documentation truth, and external legacy-alias removal are mechanically
  checked. The private migrated `vera-v1` database identity remains frozen by accepted decision.
- P6 Task 2 is complete: all 12 C4 rows are closed. Sessions has a genuine quiescent JSONL result;
  GitHub boundaries and two-repository identity are re-proven; the controlled delivery ordering is
  pinned while plain `make check` remains product-independent.
- P6 Task 3 completed portions: all 52 C5 route cells are closed by the generated 13-by-4 matrix
  (49 consume, 3 current-gate ignore, 0 reject/unclassified); C8 artifact integrity is green over
  23 schema-bearing and 6 documentary verdicts; the ignored vision assessment is preserved
  byte-identically under `docs/verification/`.
- The census is freshly regenerated at 188 total, 142 closed, 46 open, 0 unclassified:
  C1 34/34, C2 41/41, C3 0/15, C4 12/12, C5 52/52, C6 0/13, C7 0/16, C8 3/5 closed.

## Verification at wrap

- Bare `PATH=/home/thamm/go/bin:$PATH make check`: exit 0 on 2026-09-12; all shell and Go tests
  passed and `golangci-lint` reported `0 issues.` The printed `index stale; run make index` was the
  expected negative fixture, not the production index check. The cleanroom gate reported `INERT`
  because no external private pattern file was readable; it did not claim a cleanroom proof.
- `PATH=/home/thamm/go/bin:$PATH make verify`: exit 0 against the existing migrated local ledger.
- `scripts/p6-census.sh --check`, Task 1/2 closure checkers, route matrix checker, and artifact
  integrity checker: exit 0.
- `git fsck --no-dangling --no-progress`: exit 0.

## Open, blocked, and unverified

- **C8-001 / Task 3 blocker:** fresh detached Linux clone
  `a5383b947ecc304d49d6969aea83821bc828ab63` passed bare `make check`; first sync appended
  intent=3, git=173, reviews=23 and the second appended zero. `proofbound verify` then failed closed
  because P5's committed obligation verdict cites historical event
  `01M28TPW9C8R7ND19MNDCJ9GDG`, which is not source-recoverable from Git. A newly witnessed check
  did not repair the historical ID. Rebuild equality and fresh-store self-hosted reports remain
  unverified. Evidence: `docs/verification/p6-fresh-clone-linux.md`.
- **C8-002 / Task 3 blocker:** native Windows PowerShell passed the corpus check but cannot find
  native `make`; no Windows gate result exists. Evidence: `docs/verification/p6-windows-platform.md`.
- The dual-platform decision is now received and recorded in
  `docs/verification/verdicts/p6-dual-platform-ratification.md`, with semantic VD
  `docs/decisions/VD-p6-dual-platform-2026-09-13.md`. Native Windows acceptance is still open.
- One founder decision remains unreceived: authorize a strict migration-only archive of the exact
  cited historical event envelopes (recommended over relaxing dangling-evidence checks).
- The historical archive is only a recommendation: its exact envelopes have not been exported,
  validated, specified, implemented, or independently reviewed. Do not present it as decided.
- Native Windows currently has Git, Go, and `jq`; it lacks GNU Make, `golangci-lint`, and `rg` in the
  PowerShell PATH. Git Bash exists at `C:\Program Files\Git\bin\bash.exe`; the first installer
  attempt was stopped after the Go MSI made no progress. `setup-windows.ps1` now installs only missing
  packages, so the next attempt will not unnecessarily upgrade Go.
- Tasks 4–6 and 8–9 have not started. The 15 C3 package rows still require final-code calibrated
  mutation sweeps and non-author acceptance; 13 C6 intent/review rows and 16 C7 measurement/falsifier
  rows remain open; final round-C acceptance is absent.

## Institutionalized improvement

Task 3 repeatedly committed valid closure artifacts before binding them into the generated census,
leaving the production census checker red while component tests stayed green. The completed evidence
is now bound and `scripts/tests/p6-census-current.test.sh` runs the production
`scripts/p6-census.sh --check` inside every `make hooks-test` / bare `make check`. The gate registry
documents this lightweight freshness backstop.

## Standing cautions

- P6 authority is `docs/decisions/VD-p6-consolidation-2026-09-12.md` and ratified draft 2.
- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, and all pinned vectors.
- Do not relax evidence referential integrity or claim fresh-clone, native-Windows, package, or P6
  acceptance without the missing mechanical and non-author evidence.
- Use host execution for the Snap Go toolchain on this machine. Do not delete or commit the
  pre-existing untracked P5 draft without explicit ownership resolution.
