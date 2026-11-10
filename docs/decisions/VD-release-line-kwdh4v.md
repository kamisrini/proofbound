# VD-release-line-kwdh4v: Release 1 is the Trust Record; the revenue organs are Release 2

**Status:** Accepted
**Date:** 2026-11-10

**Context:** The repo is sequenced by phases (P0–P7) and by the vision's four organs, with no
"release" concept — so there is no defined line between "done in the repo" and "usable by someone
other than this repo." The owner asked to draw that line now: cut a usable **Release 1** from what is
already deep-complete and non-author accepted, and explicitly defer the rest to **Release 2**, rather
than wait for the full 2028 vision (kernel + four organs + portable warranty), which is a multi-year
arc. The forcing question: what is the smallest honest, coherent product we can call done, and what is
the wrap-up to make it usable?

**Decision:** **Release 1 = "the Trust Record."** Its scope is the trust-record foundation that is
already deep-complete: the Flight Recorder (event ledger + git/check-runner/agent-session connectors +
projections + `proofbound verify`, P1), gates-as-data (P2), the first external read-only GitHub
connector (P3), intent provenance (the BD/BR/CI chain and governed change intents, P5), and behavior
locks with real-postgres runtime replay (P7 B1). Release 1 **records and verifies**; it does not
regenerate, predict, govern, or issue portable warranties. The usability wrap-up for Release 1 is a
fixed hardening set, not new features: close `DEF-KERNELCHECK-FLAKE-001` (fix already non-author
accepted; owner final close outstanding), a clean-clone build+`make check` proof, and the
`DEF-ROUTE-MATRIX-PRECOMMIT` determinism fix; `DEF-DELIVERY-ENFORCE-VERDICTS` (end-to-end
governance-gate green) is **recommended-in but owner-optional** for R1. **Everything else is Release 2
and is marked so:** #2b witnessed regeneration's *full* loop (step 3 regenerator / step 4 candidate
verification — the gate-only slice is already CI-complete), #3 Twin-as-product (prediction ledger,
public Brier calibration, differential-verdict surface), #4 Governor, #5 Trust Engine (portable signed
warranties, verifier identities/signing), broader connectors/event kinds, `DEF-P7-DP-TWINROOT-001`
(heavy twin-root detector-power sweep), and any **redistribution** of an embedded DB / `L12`/`L17` /
shippable embedded-DB release (still gated on the binary-exclusion law and the owner's
provenance go/no-go; `DEF-P7-001` resolved only the evidence deferral, not a release authorization).
The derived, maintained scope split lives in
[`RELEASE-PLAN.md`](../plans/RELEASE-PLAN.md); this VD is the authority for the line itself.

**Consequences:** Makes easy — a shippable, demonstrable product boundary now, on top of a solid
foundation, with a short and mostly-already-accepted wrap-up list; a clear "no" to scope creep, since
anything not in the R1 set is R2 by default. Makes hard / gave up — Release 1 deliberately ships
without the three-plus revenue-bearing organs, so it is a *trust record*, not yet the warranty
business the vision defines; it also ships **without** redistributing an embedded database (users
supply their own Postgres via `DATABASE_URL`, per the binary-exclusion law), which narrows "usable"
to run-it-yourself rather than turnkey. The R1/R2 split is a product line laid over the phase record,
not a replacement for it — the ROADMAP phases and the vision build order remain the technical
authority, and this VD only groups them into two shippable cuts.

**Revisit when:** Release 1's wrap-up set is complete and the owner accepts R1 as shipped (then open a
Release 2 scope decision); OR the owner decides to pull a Release-2 item forward into R1 (e.g.
authorizes embedded-DB redistribution, or the #2b full loop); OR the vision build order changes such
that "Trust Record" is no longer the right R1 boundary.
