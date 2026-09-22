# Proofbound Roadmap

Phases are sequential; each has a mechanical Definition of Done. A phase is not done because it feels done — it is done when its DoD checks pass. Update `notes/state.md` as phases move.

**★ Every phase climbs toward the north star** (VD-north-star-6io56h: software as verified promises — 100x sets direction, 10x sets pace). Each phase review asks: *did we build what we planned?* AND *did we imagine hard enough?*

## P0 — Scaffolding ✅ when `make check` is green

Repo structure, CLAUDE.md constitution, hooks (self-tested), gates registry, decision records, session protocol, meta-tax metric.

**DoD:** `make check` green · `make hooks-test` proves every hook fires · founding decisions recorded · state.md + journal established · initial commit.

## P1 — Flight Recorder kernel, self-hosted (target: +6 weeks)

The event ledger + witness substrate, with **this repo as the first tenant**: ingest Proofbound's own git commits, `make check` runs (as witnesses), and agent-session telemetry. Zero external dependencies. Full execution plan: [docs/plans/P1-flight-recorder-plan.md](docs/plans/P1-flight-recorder-plan.md).

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

**P6 final consolidation review 2026-09-22:** P1 is deep-complete under the P6 round-C verdict and
the final zero-open census; evidence is recorded in
[`p6-task9-consolidation.md`](docs/verification/p6-task9-consolidation.md).

## P2 — Gates as data ✅ (target: +4 weeks)

**Status 2026-08-26:** P2 is complete. The gate set and delivery boundary are complete: `gates/make-check-success.yaml`, the
ledger-backed `proofbound gates canary` command, proof-bearing PASS/BLOCKED/UNKNOWN results, and an
explicit `proofbound gates enforce` path. All seven current definitions are promoted to `mode: enforce`
after PASS canary evidence; enforcement is explicit and fails closed. Gate definitions also carry
an ISO expiry date, and enforcement rejects expired definitions. The canary→enforce bad-witness
sequence is proven in `docs/verification/p2-gate-evidence.md`. The runtime P0 checks are represented
by dedicated witnessed gates; `hooks-test` remains explicitly retained as a mechanism self-test.

**P6 final consolidation review 2026-09-22:** P2 is deep-complete under the same round-C verdict
and zero-open census; evidence is recorded in
[`p6-task9-consolidation.md`](docs/verification/p6-task9-consolidation.md).

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

**P6 final consolidation review 2026-09-22:** P3 is deep-complete under the same round-C verdict
and zero-open census; evidence is recorded in
[`p6-task9-consolidation.md`](docs/verification/p6-task9-consolidation.md).

## P4 — Twin spike (target: +6 weeks)

**Accepted 2026-08-27:** `kernel/internal/twin` provides bounded replay, a
temporary embedded-PostgreSQL projection path, deterministic `vera.replay.v1` proof metadata,
and validated in-memory forecast calibration. Isolated replay preserves the source ledger and
projects only in a disposable store. Independent acceptance is recorded in
[`p4-current-round1.md`](docs/verification/verdicts/p4-current-round1.md); durable prediction
events remain a later feed-backed extension.
Decision and acceptance boundary: [VD-p4-twin-replay-calibration-2026-08-26](docs/decisions/VD-p4-twin-replay-calibration-2026-08-26.md).

**P6 final consolidation review 2026-09-22:** P4 is deep-complete under the same round-C verdict
and zero-open census; evidence is recorded in
[`p6-task9-consolidation.md`](docs/verification/p6-task9-consolidation.md).

## P5 — Intent provenance (ratified 2026-09-10)

P5 builds the smallest honest, proof-bearing chain from a declared business decision through exact
requirement obligations, change intent, commit claim, independent verdict, and observed deployment.
The ratified execution plan is
[`P5-PROOFBOUND-INTENT-PROVENANCE-v3.md`](docs/plans/P5-PROOFBOUND-INTENT-PROVENANCE-v3.md);
the semantic decision is
[`VD-p5-intent-provenance-2026-09-10`](docs/decisions/VD-p5-intent-provenance-2026-09-10.md).

**P5 DoD, in required order:**

1. Ratification artifacts, reviewed exhibit, local baseline, semantic VD, roadmap, and stale README
   status are durable and mutually linked.
2. Live product identity migrates to Proofbound while frozen `vera.witness.v1`, `vera.verdict.v1`,
   `vera.replay.v1`, pinned vectors, and historical artifacts remain byte-compatible; the complete
   case-insensitive identity inventory has zero unclassified hits and records dated alias removal.
3. Spec-first contracts freeze the canonical BD/BR/CI model, requirement review and verdict schemas,
   both digest meanings, provider interface, native-record and `specdir` mappings, hostile fixtures,
   conformance tests, and the unmappable-provider STOP control before implementation.
4. `intent.records` and `intent.specdir` ingest immutable revisions idempotently; partial chains stay
   honest; replay from an empty projection store reconstructs all relation and obligation meaning.
5. Explicit commit trailers bind exact CI revisions with correct commit-tree/provider point-in-time
   resolution; missing, withdrawn, ambiguous, wrong-digest, spoofed, and unordered references fail
   closed while historical citations remain readable.
6. Disposable projections and reports preserve event proof and distinguish missing, contradictory,
   superseded, inconclusive, and unverified components. Non-verifiable or absent requirement review
   caps green without blocking general users.
7. Verdict v2, requirement-review v1, and deployment joins bind exact commits, revisions,
   obligations, prior evidence events, and environments; v1 artifacts retain their original meaning.
8. Reference- and verdict-integrity gates prove bad chains BLOCK and good chains PASS before
   promotion. Delivery readiness graduates only in Proofbound's own `make delivery-enforce` boundary.
9. Proofbound self-hosts P5 with a complete BD/BR/CI chain, non-author all-`VERIFIABLE` requirement
   review, exact commit and evidence binding, v2 verdict, report, and a demonstrated bad-chain block.
   Bare `make check` and `proofbound verify` pass; calibrated package mutation sweeps are green; and
   a non-author independent package verdict is committed on receipt.

**P6 final consolidation review 2026-09-22:** P5 is deep-complete under the same round-C verdict
and zero-open census; evidence is recorded in
[`p6-task9-consolidation.md`](docs/verification/p6-task9-consolidation.md).

## P6 — Consolidation: deep completion of P1–P5 (ratified 2026-09-12)

Depth before width. P6 closes a mechanically generated C1–C8 census covering live documentation,
gate/build-law estate, current package acceptance, connector reality, the full source-kind event
universe, Proofbound intent self-coverage, P5 measurements, and clean-clone/platform/artifact
integrity. Every live P1–P5 promise is either evidenced or explicitly superseded, narrowed, or
retired; deferral alone is not completion. The ratified plan is
[`P6-CONSOLIDATION-PLAN-draft2.md`](docs/plans/P6-CONSOLIDATION-PLAN-draft2.md), and the semantic
authority is
[`VD-p6-consolidation-2026-09-12`](docs/decisions/VD-p6-consolidation-2026-09-12.md).

**DoD:** committed census with zero unclassified rows and zero open `close-in-P6` rows; no live
promise hidden by defer/wontfix; all production packages mutation-green and non-author accepted on
one frozen final implementation commit; zero expired advisories; full event-route and clean-clone
proof; round-C non-author verdict ACCEPTABLE and committed; bare `make check` and
`proofbound verify` green. P7+ capability planning begins only after P6 closes.

**Status 2026-09-22:** Task 9 final consolidation is complete. The round-C verdict is ACCEPTABLE,
the final census is 327/327 closed with zero unclassified rows, and P1–P5 are annotated
deep-complete with evidence above. Kernel lint remains blocked by the documented Go 1.26-built
`golangci-lint` versus Go 1.27 compatibility mismatch and is not claimed as passed.

**Ratified boundaries:** no new capability; the conditional snapshot provider is skipped and its
contract remains pinned for P7+; intent applicability uses the closed path rule in the semantic VD,
with zero known false negatives and a greater-than-10% false-positive redesign trigger; the
stop-early ceiling is 40 active execution hours after Task 0, excluding founder/verifier wait time.
P6 retains both Linux and native Windows execution for the existing tooling under
[`VD-p6-dual-platform-2026-09-13`](docs/decisions/VD-p6-dual-platform-2026-09-13.md); native Windows
acceptance remains an evidence requirement, not an assumed green result.

**Standing rules across all phases:** meta-tax within budget (docs/gates.md) · no new primitive without a feed · no hand-authored fact rows · Go ≥1.26 + golangci-lint installed at P1 start (VD-stack-go-fid9mi).
