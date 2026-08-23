# VERA MIGRATION PROMPT — rebuild on a new machine from text only

**Purpose:** move VERA to a different machine carrying ONLY text artifacts (plans, SPECs, decisions,
verdicts, journals, registries, locks, prompts) — no source code, no scripts, no binaries — and
regenerate everything to the same quality bar there, then continue to completion.
**Why this works here and almost nowhere else:** this repo runs spec-first by law (Law 6), so the
code is DERIVED state. Its sources of truth — the SPECs' interface locks and invariant tables, the
pinned byte-exact vectors, the two numbering locks, the verdict corpus — all travel as text, and
three of them double as MECHANICAL fidelity proofs after regeneration. You are not losing the code;
you are re-deriving it from the same documents it was derived from the first time, with the
expensive lessons pre-paid.
**Relationship to GENESIS-PROMPT.md:** genesis bootstraps the project from NOTHING (scaffolding,
gates, plans). This prompt is the second stage layered on it: genesis first, then the carried
originals overwrite regenerated prose, then code re-derivation against the carried SPECs.

---

## STAGE 0 — build the kit ON THIS MACHINE (one command, text-only by construction)

From the repo root:

```
KIT=~/vera-migration-kit-$(date +%F) && mkdir -p "$KIT" && \
git ls-files | grep -E '\.(md|lock|txt)$' | tar -cf "$KIT/kit.tar" -T - && \
git log --reverse --format='%ad %h %s' --date=short > "$KIT/history.txt" && \
tar -tf "$KIT/kit.tar" > "$KIT/MANIFEST.txt" && \
tar -tf "$KIT/kit.tar" | grep -vE '\.(md|lock|txt)$' | wc -l   # MUST print 0
```

The final line is the audit: the kit contains nothing but `.md`, `.lock`, `.txt`. `history.txt` is
the commit narrative (derived text — git history itself does not travel; the journals and verdicts
carry the story, this carries the sequence). Transport the kit directory however you carry text.

**What is deliberately NOT in the kit:**
- Any `.go`, `.sh`, `.js`, `Makefile`, `.yml`, `go.mod/sum`, SQL — the constraint, and all derivable.
- The cleanroom pattern file — it lives OUTSIDE the repo by design (the guard must not carry the
  contraband). On the new machine either recreate it at its private path by your own means (never
  through this kit or the repo) or accept that `cleanroom-lint` runs loudly INERT until you do.
- `notes/tmp/` — nothing load-bearing lives there, by law.

## STAGE 1 — on the new machine: genesis

`mkdir -p ~/vera-core && cd ~/vera-core && claude --model opus`, then paste the ENTIRE
GENESIS-PROMPT.md (it is in the kit). It rebuilds scaffolding, hooks, gates (each gate is specified
there as a behavioral contract with self-tests), Makefile, commands, and skeleton docs — and it ends
with `make check` GREEN and audits passed. Genesis regenerates prose *faithfully but not
byte-identically*; that is what Stage 2 fixes.

## STAGE 2 — overlay the carried originals (originals WIN)

Unpack `kit.tar` over the repo root, letting every carried file overwrite its regenerated
counterpart — the originals are the record; regenerated prose was only ever a stand-in. Then
reconcile mechanically, not by eye:

1. `make index` (decision INDEX regenerates over the carried decisions).
2. `make laws-lock && git diff docs/laws.lock` — MUST be empty against the carried lock. If not,
   genesis drifted a law: fix CLAUDE.md to match the carried lock, never the reverse.
3. `make check` — the carried gates.md, registries and allowlists are now live against the
   regenerated gate scripts. Fix any mismatch in the SCRIPT (the carried registry is the record).
4. Commit: `chore: migrated text corpus over genesis scaffold — originals win`.

## STAGE 3 — re-derive the kernel, package by package, against the carried SPECs

Order: `core` → `store` → `connector/git` + `gitcmd`. For each package the carried SPEC is the
contract and the method is the house method:

1. **SPEC-side fixes FIRST.** The carried SPECs include statements the verdicts proved false that
   are still queued (see `notes/state.md` Blockers #1 — e.g. the parent SPEC's NUL premise at
   INV-21, the § 4 table's duplicated sub-id). Apply the round-3 fix list's SPEC-side items BEFORE
   implementing — it is cheaper to derive from a corrected spec than to fix derived code.
2. **Tests from the invariant table.** Every row names its test (`file.go::TestName`); write them
   failing-not-fake-passing, exactly as `/vera-spec` demands. The table IS the test plan — 122
   invariants across the four SPECs.
3. **Implement until green**, honoring the SPECs' § 2 interface locks verbatim (they pin names AND
   struct fields).
4. **Fidelity anchors — mechanical, run all of them:**
   - **Pinned vectors:** the SPECs carry byte-exact payloads and content_shas (including the
     empty-slice second vector). The rebuilt marshalling MUST reproduce them exactly. This is the
     strongest possible proof the re-derivation matches the original ledger semantics.
   - **`make invariants-lock && git diff docs/invariants.lock`** — MUST be empty against the
     carried lock (modulo the SPEC-side fixes from step 1, whose lock changes were reviewed then).
   - **Verdict regression probes:** the three carried adjudications enumerate every known defect
     with reproductions (the 0x1e framing break, path TrimSpace fabrication, `--all` scope, shallow
     and grafts rewrites, Tips coverage, the citation ambiguity, the watermark shapes…). Re-run
     their probes against the rebuilt code — the defect corpus is the cheapest acceptance suite
     this project owns.
5. **The acceptance bar (ROADMAP § P1, binding):** `make mutants` green for the package (survivors
   killed or declared — the carried `allowed-survivors.txt` reasons transfer, but the keys are
   `line#ordinal` and WILL differ in a rebuild: re-sweep and re-key, reusing the carried reasoning)
   AND a non-author adversarial round returning ACCEPTABLE, verdict committed on receipt.

## STAGE 4 — continue to completion

Resume exactly where this machine stopped, from the carried `notes/state.md`:
round-3 fix list (code-side items — including the founder-ruled tree-resolution semantics for
`cited_decisions` and the grafts refusal) → re-sweep → **round 4** (Task 4's acceptance candidate)
→ Tasks 5–9 per the plan's Position section, Task 9 before the 2026-09-18 spec-first expiry.
Task 5's design inputs are carried as decisions (VD-verification-asymmetry-2dyjnd,
VD-verdicts-are-artifacts-rl0rab) — cite them, don't reconstruct them.

## Honest costs and boundaries (read before deciding)

- **Rebuild effort:** Tasks 1–4 originally cost ~3 days INCLUDING discovering everything the kit
  now carries. The re-derivation pass is bounded by typing speed against complete SPECs — expect
  roughly a day of agent time for the three packages, plus the acceptance rounds you choose to run
  (a full adversarial round is ~0.5–1M subagent tokens; the verdict-probe regression suite is the
  cheap substitute for the first two packages if you accept round-count risk — `core` passed in one
  round originally, `store` took nine, so scope rounds accordingly).
- **Git history does not travel.** `history.txt` carries the sequence; journals and verdicts carry
  the narrative and the evidence. The new repo's ledger starts at its own genesis under the
  repo-local maintainer identity (public-bound rule from genesis applies unchanged).
- **Machine-specific figures reset:** suite durations, the embedded-postgres contention range, and
  every `line#ordinal` survivor key are properties of a machine or a build — remeasure, re-key.
- **The boundary of "text-only":** the SPECs quote Go signatures and the gates doc quotes command
  fragments — that is what design documents are; no executable file travels. If your policy is
  stricter than "no code files" (e.g. no code-like text at all), the SPECs cannot travel either and
  regeneration falls back to GENESIS-PROMPT alone — everything still rebuilds, but the fidelity
  anchors and ~3M tokens of adversarial findings are lost with it. Decide before packing.

## THE NEW-MACHINE PROMPT (paste verbatim as the first message, after placing the kit)

```
ultracode. This machine has ~/vera-migration-kit-<date>/ containing kit.tar (text-only corpus of an
existing project), MANIFEST.txt, and history.txt. Execute the migration protocol:

1. Read MIGRATION-PROMPT.md from the kit FIRST — it is the protocol; follow its stages exactly.
2. Stage 1: run GENESIS-PROMPT.md from the kit as if pasted fresh (bootstrap scaffolding + gates,
   make check GREEN).
3. Stage 2: overlay the kit over the repo (originals WIN), reconcile with make index /
   make laws-lock / make check — fix scripts to match carried records, never the reverse.
4. Stage 3: re-derive kernel packages core → store → git+gitcmd against the carried SPECs:
   SPEC-side fixes from notes/state.md Blockers #1 first, tests from the invariant tables, then
   implementation. Prove fidelity mechanically: pinned vectors reproduce byte-exact, regenerated
   invariants.lock diffs empty against the carried one, verdict-corpus regression probes pass.
   Package acceptance = make mutants green + a non-author adversarial verdict (ROADMAP § P1 bar).
5. Stage 4: resume from the carried notes/state.md — round-3 fix list, round 4, then Tasks 5-9.
   Task 9 lands before 2026-09-18.

Work autonomously; ask me only for genuine value tradeoffs. The carried state.md's Resume Prompt
facts are binding context. Cleanroom note: the private pattern file did not travel by design —
tell me once, then proceed with cleanroom-lint loudly INERT until I configure it.
```
