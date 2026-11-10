# Release plan — Release 1 (Trust Record) vs Release 2 (the organs)

**What this is:** the maintained, derived view of the two-release line the owner drew on 2026-11-10.
The authority for the *line itself* is [`VD-release-line-kwdh4v`](../decisions/VD-release-line-kwdh4v.md);
the technical authorities underneath remain [ROADMAP.md](../../ROADMAP.md), the vision build order
(vision-2028 §9), and [`product-completion-plan-2026-11-05.md`](product-completion-plan-2026-11-05.md).
This file does not restate rationale — it tracks *which cut each item is in* and *what remains to ship
R1*. Update it as R1 wrap-up items close.

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
| R1-2 | Clean-clone proof — fresh clone builds + `make check` green on a clean machine | medium | not started | ✅ required |
| R1-3 | `DEF-ROUTE-MATRIX-PRECOMMIT` — staged-aware derived-pin freshness (pre-commit blindspot) | small | not started | ✅ required |
| R1-4 | `DEF-DELIVERY-ENFORCE-VERDICTS` — bring ~11 pre-existing verdicts to `vera.verdict.v1` so `make delivery-enforce` is green end-to-end | medium (repo-wide, mechanical) | not started | ⚠ recommended, owner-optional |

**R1 is shippable when:** R1-1..R1-3 are closed, R1-2 proven, the owner accepts R1 as shipped, and
(if chosen) R1-4 is green. `make check` is already green at HEAD; `make delivery-enforce` is NOT a
`make check` gate, so R1-4 does not block the local gate — only the end-to-end governance story.

## Release 2 — the revenue organs + turnkey redistribution (everything else)

Marked R2 by default: anything not in the R1 set above is Release 2.

| Item | Vision step | Status | Notes |
|---|---|---|---|
| **Witnessed regeneration — FULL loop** (step 3 regenerator / step 4 candidate verification) | #2b | gate-only slice CI-complete; full loop not started | reopens the §5.1 regenerator-boundary decision; `DEF-P7-REGENLOOP-001` |
| **Living Twin as a product** — prediction ledger, public Brier calibration, differential-verdict surface | #3 | primitives exist (`ReplayIsolated`, `Calibrate`); product surface not started | snapshot provider skipped (Gate 1) — needs owner un-skip |
| **Governor v1** — policy envelopes, scoped credentials, human decision inbox | #4 | not started | new capability; needs owner scope + change intent |
| **Trust Engine** — portable signed warranties, verifier identities + signing, verifier marketplace | #5 | not started | the business model; depends on #2b–#4 + verifier signing |
| `DEF-P7-DP-TWINROOT-001` — heavy full-package `internal/twin` detector-power sweep | #2b input | OPEN, deferred | only needed as input to the #2b full loop |
| **Embedded-DB redistribution** — shippable embedded-DB release, clean-clone-with-DB, `L12`/`L17` | release/hardening | gated | binary-exclusion law unchanged; `DEF-P7-001` resolved the *evidence* deferral only, NOT a release authorization; needs owner go/no-go |
| Broader connectors / event kinds; verifier identities + signing (deferred since P1) | cross-cutting | not started | under existing no-hand-authored-facts + feed rules |

**R1-vs-R2 boundary rule:** Release 1 ships a system that **records and verifies** on this repo plus
one external connector, run against a user-supplied Postgres. It does **not** regenerate, predict,
govern, issue portable warranties, or redistribute a database. Each of those crosses into Release 2.
