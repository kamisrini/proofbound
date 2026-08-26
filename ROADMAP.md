# VERA Roadmap

Phases are sequential; each has a mechanical Definition of Done. A phase is not done because it feels done — it is done when its DoD checks pass. Update `notes/state.md` as phases move.

**★ Every phase climbs toward the north star** (VD-north-star-6io56h: software as verified promises — 100x sets direction, 10x sets pace). Each phase review asks: *did we build what we planned?* AND *did we imagine hard enough?*

## P0 — Scaffolding ✅ when `make check` is green

Repo structure, CLAUDE.md constitution, hooks (self-tested), gates registry, decision records, session protocol, meta-tax metric.

**DoD:** `make check` green · `make hooks-test` proves every hook fires · founding decisions recorded · state.md + journal established · initial commit.

## P1 — Flight Recorder kernel, self-hosted (target: +6 weeks)

The event ledger + witness substrate, with **this repo as the first tenant**: ingest VERA's own git commits, `make check` runs (as witnesses), and agent-session telemetry. Zero external dependencies. Full execution plan: [docs/plans/P1-flight-recorder-plan.md](docs/plans/P1-flight-recorder-plan.md).

**P1 preflight:** Go ≥1.26 pinned via the `go.mod` toolchain directive; golangci-lint installed; `kernel/go.mod` is the module the Makefile builds/tests; test DB = embedded-postgres per VD-stack-go-fid9mi (daemon-free, `DATABASE_URL` escape hatch).

- `kernel/` single Go module (Go + Postgres/pgx — VD-stack-go-fid9mi)
- `events` append-only table; idempotency on (source, native_id, content_sha)
- Connectors: git (commits→events), check-runner (make check emits witness JSON, content-digested; signing deferred to P2 verifier identities), agent-sessions
- One projection + one view: "what happened this week, with proof" — replaces the journal's factual half
- **DoD:** re-ingest twice → zero new events · drop projections + replay → row-set-identical projections · witness event for every `make check-witnessed` run since Task 5 landed · `make check` extended to kernel build/test/lint · spec exists for every package

**Scope added 2026-08-11 (vision pass).** The kinds registry grows a fourth kind, **`review.verdict`** —
finding id, severity, the commit under review, and the commit that introduced the defect *where known*.
Without it the ledger records that checks RAN, not that anything was FOUND: every defect in Task 3 was found
by adversarial review while `make check` was green, so a fix→check→break series derived from `check.run`
alone reads zero for the week that measured one to three new defects per remediation cycle. Added to the P1
DoD: a projection exposing that chain — a change following a red verdict and preceding another.
**Stated limit, deliberately:** attributing a fix to a specific finding is an AUTHORED claim (this repo does
it by commit-subject convention), and Build Law 1 refuses authored state. The series is measurable; the
attribution is not, and the projection must say so on its face. **Falsifier:** if the series shows no signal
distinguishable from noise across three packages, the metric is n=1 folklore and gets demoted from "probably
the commercial one".

**Package acceptance bar (added 2026-08-14, Law 9 + VD-verification-asymmetry-2dyjnd):** a kernel
package is DONE only when `make mutants` is green for it (all survivors declared) AND a non-author
adversarial verdict says ACCEPTABLE, committed under `docs/verification/verdicts/`. Author-green
predicts nothing — measured three times.

**Position 2026-08-26:** P1 Tasks 0–9 are DONE. Task 9 closed the review-verdict connector and
ledger-ordered red-verdict chain, graduated spec-first enforcement to a blocking `make check`
test, and recorded the full verifier run. Task status detail: the plan's Position section.

## P2 — Gates as data ✅ (target: +4 weeks)

**Status 2026-08-26:** P2 is complete. The gate set and delivery boundary are complete: `gates/make-check-success.yaml`, the
ledger-backed `vera gates canary` command, proof-bearing PASS/BLOCKED/UNKNOWN results, and an
explicit `vera gates enforce` path. All seven current definitions are promoted to `mode: enforce`
after PASS canary evidence; enforcement is explicit and fails closed. Gate definitions also carry
an ISO expiry date, and enforcement rejects expired definitions. The canary→enforce bad-witness
sequence is proven in `docs/verification/p2-gate-evidence.md`. The runtime P0 checks are represented
by dedicated witnessed gates; `hooks-test` remains explicitly retained as a mechanism self-test.

- Gate definitions in `gates/*.yaml` evaluated by the kernel against the ledger (replaces parts of Makefile checks)
- Canary evaluation against historical events before a gate can block
- Advisory-expiry enforcement moves from docs/gates.md prose into the engine
- **DoD:** every P0 Makefile check re-expressed as a data gate OR explicitly retained with reason · one gate demonstrably blocked a real bad change in canary-then-enforce sequence

## P3 — First external connector ✅ (target: +4 weeks)

**Status 2026-08-26:** P3 is complete. The first connector decision selected a narrow, read-only
GitHub Actions and deployments slice for one organization with an explicit repository allowlist.
The connector, joined deployed-where / tested-what projection, and CLI sync/report surface are
implemented; live acceptance is recorded in `docs/verification/p3-github-live-acceptance.md`.

- **DoD:** the deployed-where / tested-what view running on real external data; cold sync < 10 min; freshness rendered on every surface

## P4 — Twin spike (target: +6 weeks)

**In progress 2026-08-26:** the first bounded replay contract is implemented in
`kernel/internal/twin`: it selects a ledger prefix, validates sequenced candidates, and
compares injected projection snapshots without mutating production state. PostgreSQL fork
isolation, proof-bearing verdict persistence, and prediction-ledger calibration remain open.
Decision and acceptance boundary: [VD-p4-twin-replay-calibration-2026-08-26](docs/decisions/VD-p4-twin-replay-calibration-2026-08-26.md).

**Standing rules across all phases:** meta-tax within budget (docs/gates.md) · no new primitive without a feed · no hand-authored fact rows · Go ≥1.26 + golangci-lint installed at P1 start (VD-stack-go-fid9mi).
