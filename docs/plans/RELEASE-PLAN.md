# Release plan — R1 (Trust Record) · R2 (SDD Warranty Connector) · R3 (the organs)

**What this is:** the maintained, derived view of the three-release line. The owner drew the first
line on 2026-11-10 (R1 vs "everything else") and redrew it on 2026-11-16: accept R1 as shipped,
insert the **SDD Warranty Connector** as R2, and move the former R2 scope (the revenue organs +
turnkey redistribution) to R3. The authority for the *line itself* is
[`VD-release-line-r2-sdd-t5j34p`](../decisions/VD-release-line-r2-sdd-t5j34p.md) (which supersedes
`VD-release-line-kwdh4v`); the technical authorities underneath remain [ROADMAP.md](../../ROADMAP.md),
the vision build order (vision-2028 §9), and
[`product-completion-plan-2026-11-05.md`](product-completion-plan-2026-11-05.md). This file does not
restate rationale — it tracks *which cut each item is in*.

## Release 1 — "the Trust Record" (records + verifies)

**Scope — already deep-complete and non-author accepted (ships as-is):**

| Capability | Phase | State |
|---|---|---|
| Flight Recorder — ledger + git/check-runner/agent-session connectors + projections + `proofbound verify` | P1 | ✅ done |
| Gates-as-data — definitions evaluated against the ledger, canary→enforce | P2 | ✅ done |
| First external connector — read-only GitHub Actions/deployments | P3 | ✅ done |
| Intent provenance — BD/BR/CI chain + governed change intents | P5 | ✅ done |
| Behavior locks + real-postgres runtime replay | P7 B1 | ✅ accepted |

**R1 wrap-up — the hardening gaps that make it *usable by someone else* (NOT new features):**

| # | Item | Effort | Status | In R1? |
|---|---|---|---|---|
| R1-1 | `DEF-KERNELCHECK-FLAKE-001` — flaky embedded-postgres test | ~0 (done) | **FIXED + non-author accepted (r1+r2); awaiting owner final close.** Closing also repoints the `deferrals.test.sh` `--open` fixture. | ✅ required |
| R1-2 | Clean-clone proof — fresh clone builds + `make check` green on a clean machine | medium | **PROVEN 2026-11-12** ([`r1-2-fresh-clone-release1.md`](../verification/r1-2-fresh-clone-release1.md)): cold-cache build + `make check` green; live surface (migrate→sync→report/rebuild) green; two findings — (A) bootstrap order undocumented, (B) `verify` red on fresh clone via R1-4. | ✅ required |
| R1-3 | `DEF-ROUTE-MATRIX-PRECOMMIT` — staged-aware derived-pin freshness (pre-commit blindspot) | small | **DONE 2026-11-12** (`a5a085b`, `CI-staged-pin-guard-7b3c9d`): route-matrix + `p6-task2-close` `--check` fail closed on pending pinned-path changes (seam-tested); `p6-census` excluded by design. Deferral RESOLVED. | ✅ required |
| R1-4 | `DEF-DELIVERY-ENFORCE-VERDICTS` — bring the pre-schema verdicts to `vera.verdict.v1` so `proofbound verify` / `sync all` run on a fresh clone | medium (repo-wide, mechanical) | **DONE 2026-11-12** (`421938c`): 15 pre-schema verdicts normalized (status synonyms + structured findings; originals kept verbatim in-body). Proven on a clean ledger: `sync reviews` malformed=0, `sync all` exit 0, `verify` exit 0. `DEF-DELIVERY-ENFORCE-VERDICTS` RESOLVED. | ✅ required (owner, 2026-11-12) |

**R1 is shippable when:** R1-1..R1-4 are closed, R1-2 proven, and the owner accepts R1 as shipped.
**R1-1..R1-4 are all DONE** ([`r1-2-fresh-clone-release1.md`](../verification/r1-2-fresh-clone-release1.md)
is the clean-clone proof; the corp-readiness probe additionally restored `gates enforce`).
**✅ Release 1 ACCEPTED as shipped by the owner on 2026-11-16** (`VD-release-line-r2-sdd-t5j34p`).
Follow-ups: tag the release, add a demo script, and run a single-dev corp pilot.

## Release 2 — the SDD Warranty Connector

**Built; awaiting owner acceptance** (notes: [`docs/releases/r2/RELEASE-NOTES.md`](../releases/r2/RELEASE-NOTES.md),
demos: [`docs/releases/r2/DEMO.md`](../releases/r2/DEMO.md)). One connector lets Proofbound **warrant
the output of spec-driven-development tools** (OpenSpec and Spec Kit). It ingests their specs as
intents from git history, binds each tenant commit's `Intent: sdd:<name>` claim to the exact spec
revision, applies the existing readiness gate (a non-owner SATISFIED verdict plus a VERIFIABLE review
for every claimed obligation), and reports **promise drift**. It **carries all of R1**: the SDD tool
is a new *source* feeding the same Trust Record, and nothing bypasses the intent → verdict → gate
path. The build plan (re-sequenced 2026-11-19 under Law 11) is
[`sdd-warranty-connector-plan.md`](sdd-warranty-connector-plan.md).

| Item | Status | Where |
|---|---|---|
| M0: mapping spike, provider SPEC, intent home (`BD`/`BR-sdd-warranty-*`) | ✅ accepted 2026-11-17 | `kernel/internal/connector/intent/sdd/SPEC.md`; BR rev 0003 |
| M1: `intent.sdd` provider core, OpenSpec (`sync intent sdd`) | ✅ accepted 2026-11-18 | `CI-sdd-warranty-m1-6op6vl` |
| M3: commit binding + readiness gate on sdd chains | ✅ accepted 2026-11-19 | `CI-sdd-warranty-m3-xaqocp`; verdicts `r2-m3-*` |
| Real-repo demo on OpenSpec's own repository; clean output, gate reasons, lookup by name | ✅ accepted 2026-11-19 | `docs/releases/r2/demo-openspec.sh`; `r2-demo-polish-review` |
| Starter kit + reviewer workflow (`init`, `review draft`/`seal`) | ✅ accepted 2026-11-19 | `CI-starter-kit-68z7m8`; `r2-starter-kit-*` |
| M4: promise drift (`report drift`) | ✅ accepted 2026-11-19 | `CI-sdd-warranty-m4-*`; BR rev 0005 O-14; `r2-m4-drift-*` |
| Spec Kit (old M2 Spec Kit mapping + M5, merged) | ✅ accepted 2026-11-19 | `CI-sdd-warranty-speckit-eiis5s`; `r2-speckit-review` |
| Full-loop demos for both kits | ✅ 2026-11-19 | `docs/releases/r2/demo-loop.sh` |
| Release notes + owner acceptance | ⏳ awaiting the owner | tag `v0.2.0` after acceptance (pushes held) |

**Not in R2** (stated in the release notes): CI check runs as verdict evidence (the evidence is the
reviewed commit's git event), a drift *gate* (drift is a report), Spec Kit status lines, and parsing
OpenSpec main specs or `design.md`.

## Release 3 — the revenue organs + turnkey redistribution (everything else)

Marked R3 by default: anything not in the R1 or R2 sets above is Release 3. (This is the former
Release 2 scope, unchanged — only the label moved from R2 to R3 on 2026-11-16.)

| Item | Vision step | Status | Notes |
|---|---|---|---|
| **Witnessed regeneration — FULL loop** (step 3 regenerator / step 4 candidate verification) | #2b | gate-only slice CI-complete; full loop not started | reopens the §5.1 regenerator-boundary decision; `DEF-P7-REGENLOOP-001` |
| **Living Twin as a product** — prediction ledger, public Brier calibration, differential-verdict surface | #3 | primitives exist (`ReplayIsolated`, `Calibrate`); product surface not started | snapshot provider skipped (Gate 1) — needs owner un-skip |
| **Governor v1** — policy envelopes, scoped credentials, human decision inbox | #4 | not started | new capability; needs owner scope + change intent |
| **Trust Engine** — portable signed warranties, verifier identities + signing, verifier marketplace | #5 | not started | the business model; depends on #2b–#4 + verifier signing |
| `DEF-P7-DP-TWINROOT-001` — heavy full-package `internal/twin` detector-power sweep | #2b input | OPEN, deferred | only needed as input to the #2b full loop |
| **Embedded-DB redistribution** — shippable embedded-DB release, clean-clone-with-DB, `L12`/`L17` | release/hardening | gated | binary-exclusion law unchanged; `DEF-P7-001` resolved the *evidence* deferral only, NOT a release authorization; needs owner go/no-go |
| Broader connectors / event kinds; verifier identities + signing (deferred since P1) | cross-cutting | not started | under existing no-hand-authored-facts + feed rules |

**Boundary rule:** **R1 records and verifies** on this repo plus one external connector, against a
user-supplied Postgres. **R2 warrants the output of an external SDD tool**: it ingests specs as intents
and binds, gates and drift-checks them, but adds no new runtime organ. **R3
regenerates, predicts, governs, issues portable warranties, or redistributes a database.** Each of
those last four crosses into Release 3.
