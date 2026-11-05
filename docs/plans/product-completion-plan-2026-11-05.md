# Path to product completion — derived plan (2026-11-05)

**What this is:** a forward-looking plan for the work between today's state and the product the vision
defines. It is a **derived view** — the authorities are [vision-2028.md](../../vision-2028.md) (§3 the
kernel + four organs, §9 build order), [vision-100x.md](../../vision-100x.md) (the north star), and
[ROADMAP.md](../../ROADMAP.md) (the phase record). This file does not restate those; it sequences
what remains and names each item's gate. When an item starts, it gets its own package `SPEC.md`
(Law 6) and, for governed `kernel/*` work, an owner-approved change intent before any code.

**Completion bar (honest):** the full 2028 vision is a **kernel + four organs + a portable warranty
product**, de-risked over a multi-year arc, with the **2029 milestone "Proofbound builds Proofbound"**
(the control plane regenerating itself under its own warranty, meta-tax inside budget) as the real
done-line. This is years of scope. The foundation is solid; three of the four revenue-bearing organs
are largely unbuilt.

## Status against the vision build order (§9)

| # | Build-order step | Status |
|---|---|---|
| 1 | **Flight Recorder** — witness substrate + claim ledger + gates-as-data, self-hosted | ✅ done (P1–P3, deep-completed P6; this repo is its own first tenant) |
| — | **Intent provenance substrate** (BD/BR/CI chain, governed change intents) | ✅ done (P5) |
| 2 | **Behavior locks** (Intent Fabric v1) | 🟡 partial — lock + replay-verify contracts done; real-postgres runtime replay + provenance done (P7 B1) |
| 2b | **Witnessed regeneration / regeneration-as-repair** | ⬜ not started — the demonstrable wedge; gated on measured detector power |
| 3 | **Ephemeral replay universes + prediction ledger** (Twin v1) | 🟡 partial — `ReplayIsolated` + calibration primitive exist; no prediction ledger / public calibration / differential-verdict product surface; snapshot provider skipped (Gate 1) |
| 4 | **Governor v1** (autonomy ratchet + decision router) | ⬜ not started |
| 5 | **Trust Engine endgame** (verifier marketplace + portable signed warranties) | ⬜ not started — the business model lives here |
| — | **2029: Proofbound builds Proofbound** (self-hosting under own warranty) | ⬜ depends on 2b–5 |

## Remaining work, in the vision's de-risking order

### A. Finish #2b — witnessed regeneration (NEAREST, highest-leverage)
The wedge: when an invariant fails, **discard the implementation and re-derive it against the lock +
the new counterexample**, instead of patching (patching is the measured-expensive act — vision §9
item 2b). Builds directly on the behavior-lock work just completed.
- **Deliverables:** a regeneration loop (lock + counterexample → candidate implementation → verify
  against the lock); an aggregate **detector-power metric** (mutation/invariant kill-rate per
  package) that gates whether regeneration is trusted for a given lock; the "satisfy-and-reopen"
  guard (a regenerated impl must not pass a weak lock).
- **Gate (the point):** regeneration is only as good as the lock, so 2b unlocks **only where detector
  power is measured and high**. The deferred full-package mutation sweep + a detector-power report
  are prerequisites, not optional.
- **First step:** a `SPEC.md` for the regeneration mechanism + the detector-power metric (Law 6),
  then an owner-approved change intent before `kernel/*` code.

### B. #3 — Living Twin as a product (prediction ledger + calibration)
- **Deliverables:** a prediction ledger (differential replay verdicts recorded as witnessed claims),
  **public Brier calibration** published over outcomes (the `Calibrate` primitive exists — TWIN-INV-10),
  and a verdict surface (CLI/report) that presents "what would happen" with its track record.
- **Gate:** revisit the deliberately-skipped snapshot provider (Gate 1) only under the no-width
  doctrine + pinned provider contract; needs owner authorization to un-skip.

### C. #4 — Governor v1 (the programmable organization)
- **Deliverables:** policy envelopes (what agents may do unattended), scoped credentials, and the
  **human decision inbox** (minutes/day of genuinely-human decisions). Turns the autonomy ratchet
  into enforcement on the ledger.
- **Gate:** new capability — owner direction on scope + change intent.

### D. #5 — Trust Engine endgame (the business model)
- **Deliverables:** **portable, cryptographically signed warranties** (proof of what a system does,
  who verified it, who answers when it fails) that cross org boundaries; verifier **identities + signing**
  (deferred since P1); the verifier marketplace.
- **Gate:** the category move; depends on A–C and on verifier-identity/signing being built.

### E. Cross-cutting kernel breadth
- More **connectors** beyond git / check-runner / agent-sessions; more event kinds/providers under
  the existing no-hand-authored-facts + feed rules.
- **Verifier identities + signing** (deferred at P1) — a prerequisite for D.

## Release / hardening gaps (parallel, not organ-sequenced)
- **`DEF-P7-001` residuals (not closed by resolution):** actual **redistribution** of an embedded DB
  is *not* authorized; a shippable embedded-DB **release**, a **clean-clone proof**, and the
  **L12/L17** invariants are flagged in ROADMAP as still-gated downstream work. Binary-exclusion law
  unchanged.
- **`DEF-KERNELCHECK-FLAKE-001`** (OPEN) — make the flaky embedded-postgres test hermetic so full
  `make check` is deterministic.
- **Deferred full-package mutation sweep** over `internal/twin` — also a prerequisite input to A's
  detector-power gate, so doing it feeds two needs.
- Optional unit-3 LOW cleanups (verdict F1/F2, each needs a re-verify; unit-2 cosmetic license nits).

## Recommended immediate next action
Begin **A (#2b witnessed regeneration), spec-first**: write the regeneration + detector-power
`SPEC.md`, and (because the detector-power gate needs it) schedule the deferred full-package mutation
sweep as its first evidence input. That is the vision's own next de-risking step and the safest,
most-demonstrable wedge on top of the behavior-lock foundation just completed.
