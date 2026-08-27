# VERA — Build Constitution

VERA is the trust layer for AI-built software: an event-sourced delivery record (Flight Recorder), intent fabric, living twin, and trust engine. Vision: [vision-2028.md](vision-2028.md). Roadmap: [ROADMAP.md](ROADMAP.md).

**★ THE NORTH STAR (VD-north-star-6io56h, ratified):** *software stops being an artifact you possess and becomes a promise the world keeps* — [vision-100x.md](vision-100x.md). The 100x sets direction, the 10x sets pace. Every phase review asks: **did we imagine hard enough?**

**This repo is built under VERA's own laws.** If the way we build contradicts what we're building, the build is wrong.

## The Build Laws (enforced, not hoped)

> **Law numbers are permanent identifiers. New laws APPEND; none is ever inserted, re-ordered or
> re-used, and a retired law leaves a numbered tombstone.** Learned 2026-08-11: a ninth law was
> inserted at position 8, which moved "small kernel, boring tech" to 9 and silently invalidated
> NINE existing "Law 8" citations across commands, SPECs, the genesis prompt and a decision
> record. Every one still resolved to a real law and every one named the wrong one — the
> stale-index class, which no citation can defend against. Invariant numbers already work this
> way (§ store SPEC: three rows read RETIRED because numbers are never reused); the laws did not,
> and now do. `scripts/law-citation-lint.sh` detects a renumber mechanically.

1. **Derived state or dead state.** Every generated artifact carries an `@generated` marker + its regeneration command, and is never hand-edited (hook-blocked). `make check` fails if regeneration produces a diff.
2. **One home per datum.** See the table below. A fact stored twice is a drift seed — link, never copy.
3. **Enforcement proves it fires.** Every hook has a self-test; `make hooks-test` runs them all and is part of `make check`. Blocking hooks fail CLOSED when a dependency is missing.
4. **Nothing advisory without an expiry.** Every warning-tier check is registered in [docs/gates.md](docs/gates.md) with an owner and an ISO expiry date: graduate to blocking or delete. A dateless or expired advisory row fails the build.
5. **Meta-tax is budgeted.** `make meta-tax` measures scaffolding commits vs product commits; the budget's single home is [docs/gates.md](docs/gates.md). Breach = simplify the scaffolding, never grow it.
6. **Spec before code.** Every kernel package gets a `SPEC.md` (interface, invariants, non-goals, test derivations) before implementation — `/vera-spec`. Tests derive from the spec.
7. **Lessons compile to enforcement.** Journal mistakes with a greppable `LESSON:` prefix. A lesson seen twice MUST become a hook case, gate row, spec invariant, or law amendment — or be explicitly retired. Prose lessons decay; compiled lessons hold.
8. **Small kernel, boring tech.** Go + Postgres (VD-stack-go-fid9mi). No new dependency without a decision record. Prereqs: bash, git, jq; Go ≥1.26 + golangci-lint at P1.

9. **No actor grades its own work.** The author of a change may fix it; only a verifier who did not write it may accept it. Added 2026-08-11 as the ninth law, and it is the one with the most evidence behind it: nine adversarial rounds on the first package, nine non-empty verdicts, and the author — with full context and explicit rules about not trusting himself — still shipped a test that could not fail *in the same session he was fixing tests that could not fail*. This was vision Law 2 for four days while doing more work than any law that was actually written down. Corollary: a verdict is only as good as its harness, so a mutation sweep states its calibration controls (a positive control that must die, a neutral one that must survive) before its count is believed.
10. **The resume note is live state.** `notes/state.md` is rewritten every few commits, not at session end — `make check` fails when HEAD has moved more than `STATE_FRESHNESS_COMMITS` (default 3) past its last update. Added 2026-08-12, after a session where HEAD advanced EIGHT commits and the note was touched zero times, so the one file a fresh session is told to read FIRST described a repository two days gone: a different task, a closed fix list, no mention of two adversarial rounds or the gates they produced. The tell was that I announced "state.md is stale" three times in that session — a check performed by hand is a check that must be compiled (Law 7). `make wrap-verify` already asserted HEAD touched the note, but ran only at `/vera-wrap`, so it could not fire mid-session and `docs/gates.md` records three commits that bypassed it; this is that check promoted into the gate. `scripts/gen-state.sh` prints the derived half so compliance is cheap, and deliberately owns no file — a partly-`@generated` note would either block the narrative edits it exists for (Law 1) or need a second home (Law 2).

## One home per datum

| Datum | Home | Everything else |
|---|---|---|
| Product vision | `vision-2028.md` / `vision-100x.md` (+ plain-english) | links only |
| Why we chose X | `docs/decisions/VD-*.md` | cite the ID |
| Current state / resume note | `notes/state.md` (overwrite-in-place, ONE file) | never a second status file |
| Session history | `notes/journal/YYYY-MM-DD.md` (append-only) | — |
| Quality gates + advisory registry | `docs/gates.md` | — |
| Plan | `ROADMAP.md` | — |
| Package contracts | `kernel/internal/<pkg>/SPEC.md` | code cites spec |
| Throwaway work | `notes/tmp/` (gitignored) | nothing load-bearing, ever |

**Rule of thumb: if losing a file would cost more than 30 minutes, it goes in git this session.**

## No knowledge graph (VD-no-graph-2aa4vz)

Files + git + grep ARE the record during bootstrap. Derived indexes only (`make index`). The kernel itself becomes the project's record once it exists — **VERA's first tenant is VERA** (self-hosting: ingest this repo's commits, check-runs, and agent sessions as the first connector).

## Session protocol

- **Start:** the SessionStart hook prints `notes/state.md` + git status. Run `/vera-next` to orient and pick the next task.
- **End:** use the single canonical EOD prompt in [docs/eod-prompt.md](docs/eod-prompt.md). It owns
  inspection, journal/state durability, `make check`, coherent commit/push, and bundle verification.
  The next action written into `notes/state.md` is agent handoff state, not a second user command.
  `/vera-wrap` is optional only when that command exists in the checkout.
- **Decisions during work:** `/vera-decide` immediately, not at session end.
- **Before merging a meaningful chunk:** `/vera-review` (adversarial pass against these laws + the spec).
- **Parallel sessions:** exactly one session owns `notes/` writes and `/vera-wrap` at a time; subagents never write `notes/state.md` or the journal — they return results to the owning session; parallel worktree sessions write journal fragments to `notes/tmp/` and the owner merges at wrap.
- Sessions run **from this directory** — hooks and commands only load here.

## Commands

```bash
make check        # the gate: hooks-test + link-lint + generated-freshness + kernel checks (when kernel exists)
make hooks-test   # prove every hook fires
make index        # regenerate docs/decisions/INDEX.md (@generated)
make meta-tax     # scaffolding vs product commit ratio
make backup       # git bundle to ~/Backups (interim off-machine story)
```

**Never pipe `make check` through anything in a `&&` chain** — the pipe swallows the exit code and a red gate reads as green.

## Hard limits on this file

CLAUDE.md stays under 120 lines. A rule corpus that outgrows the context budget stops being read, which is the same as not existing. Add depth in decisions and specs and LINK it — never inline it here.
