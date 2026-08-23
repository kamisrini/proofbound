# VERA — Verified Delivery Organism

> ★ **THE NORTH STAR** (ratified, VD-north-star-6io56h): *software stops being an artifact you possess and becomes a promise the world keeps.* Only intent, meaning, and memory persist. The 100x vision sets direction; the 10x vision sets pace; every phase review asks **"did we imagine hard enough?"** — Aim high. Go big.

**Product vision:** a 2028-era, category-creating software delivery system. The product is the *warranty*: machine-checkable proof of what software does, who verified it, and who answers when it fails.

## The elevator pitch

> By 2028, AI writes the world's software — and no one can prove any of it is safe. VERA is the trust layer for that world: it records what agents actually did, simulates every change before reality touches it, and ships every system with a machine-checkable warranty. Code became free; we sell certainty.

**Mission:** make trusting software a fact, not a faith. **North star (2030):** no software ships without its warranty.

## Contents

| File | What it is |
|---|---|
| [GENESIS-PROMPT.md](GENESIS-PROMPT.md) | **The master bootstrap prompt** — paste into a fresh Claude session on any machine to recreate this project, verified to the same bar. Doubles as disaster recovery |
| [vision-2028.md](vision-2028.md) | **The vision (10x — build now)** — the Eight Laws, the kernel + four organs architecture, a day-in-2028 narrative, the 10x table, moats, ten hard problems, build order |
| [vision-100x.md](vision-100x.md) | **The magic-wand document (100x — steer by)** — software as verified promises; the five wonders; the affirmative axioms; what dies; why P1 serves both rungs |
| [vision-plain-english.md](vision-plain-english.md) | The same vision in everyday words |
| [CLAUDE.md](CLAUDE.md) | **The Build Constitution** — the Nine Build Laws, one-home table, session protocol. Read first |
| [ROADMAP.md](ROADMAP.md) | Phases P0–P4, each with a mechanical Definition of Done |
| [docs/plans/P1-flight-recorder-plan.md](docs/plans/P1-flight-recorder-plan.md) | The verified P1 execution plan — self-contained for a cold session |
| [docs/design/architecture-2028.html](docs/design/architecture-2028.html) | The 2028 architecture diagram with build-phase tags |
| [docs/design/continuity-chain.md](docs/design/continuity-chain.md) | How intent stays bound to code — traceability by derivation, not citation |
| [docs/gates.md](docs/gates.md) | The gate registry — every check, its tier, its self-test, its expiry |
| [docs/decisions/](docs/decisions/) | Decision records (`VD-*`) — the why behind every structural choice |

## Working in this repo

Read [CLAUDE.md](CLAUDE.md) first. On Windows, run `./setup-windows.ps1 -InstallTools` in PowerShell, then `./check-windows.ps1`. Prereqs: Git, Bash, jq (Go ≥1.26 + golangci-lint from P1). The current migration is text-only; the executable scaffold is regenerated from the carried specs before `make check` becomes available. Sessions start from this directory; orient with `/vera-next`, end with `/vera-wrap`.

## Build order (each stage is a standalone business)

1. **The Flight Recorder** — witness substrate + claim ledger + gates-as-data ("what did our agents actually do — prove it")
2. Behavior locks + witnessed regeneration (code-as-cattle)
3. Ephemeral replay universes + prediction ledger (the twin, calibration-scored)
4. Autonomy ratchet + decision inbox (the governor)
5. Verifier marketplace + portable warranties (the category move)
