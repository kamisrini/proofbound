# VERA — Current State

> THE resume note. Overwrite-in-place; never append; never a second copy. A fresh session reads this first.
> Kept live under **Law 10** — `make check` fails when HEAD moves >3 commits past its last update.
> `make state` prints the derived half; the judgement half below is hand-written.

**As of:** 2026-08-23 (session: store configuration/lock and initial pgx ledger API implemented; `make check` green; DB integration tests and embedded lifecycle remain next)

## Resume Prompt

Copy this verbatim into a fresh session started with `cd ~/vera-core && claude`:

```
Read notes/state.md FIRST, then notes/journal/2026-08-14.md (the retrospective), then 08-13 and 08-12. Do not start
work until you have.

Load-bearing facts:
- Task 4 (git connector) is BUILT, heavily remediated, and NOT YET ACCEPTED. Rounds 1 and 2
  both returned NEEDS_WORK (15 findings, then 21 against the remediation). Every HIGH and
  every MED from both rounds is now closed. Round 3 is owed and is the gate to acceptance.
- Nothing ships. gitcmd has no production caller and cmd/vera does not wire it, so every
  finding is latent. That bounds urgency; it does not bound the SPEC-truth question.
- THE CENTRAL FINDING, which outranks any fix list: hand-iterating fixes does NOT converge.
  My fix rate and my defect-introduction rate are roughly equal, and being MORE careful did
  not change it. What converges is MECHANISM — script+self-test classes have recurred zero
  times; prose-answered classes recurred 3-5x.
- Six gates now exist that did not this morning: skip-lint, prescription-lint,
  state-freshness (Law 10), index-check (extracted + self-tested), make mutants, and the
  refreshed link-lint. EACH caught a live defect on its first real run. EACH also had a bug
  its own self-test caught first.
- make mutants is GREEN for gitcmd (154 candidates, 120 killed, 17 declared). Do NOT read
  that as "gitcmd is proven": small operator set, never mutates _test.go, and it says
  nothing about whether a test asserts the RIGHT thing.
- Fixing the REPRODUCTION is not fixing the INVARIANT. Before closing anything: write the
  HARM in the finding's own words, then enumerate ROUTES to it.
- A verified FINDING does not make its proposed FIX verified. Round 2's remedy for the
  citation defect would have broken a real decision id; check remedies against real data.
- A "cannot/never/always" in a comment is a claim needing a test. Two of mine were false.
- Law numbers and SPEC invariant numbers are PERMANENT: APPEND, never insert, then regenerate
  the matching lock (make laws-lock / make invariants-lock) in the same commit. Both namespaces
  now have detectors; the lock DIFF is the review.
- Two product VDs landed 2026-08-14 and are the citable grounding for Task 5+ design:
  VD-verification-asymmetry-2dyjnd (mechanism beats care — measured) and
  VD-verdicts-are-artifacts-rl0rab (verdicts commit on receipt; evidence binds to content,
  not position).
- LESSON tags now use the six parent classes consolidated in the 08-14 journal — 18 singleton
  tags in two days had quietly defeated the recurrence counter.
- make short is the inner loop; make check is the gate and never takes -short.
- macOS trap: t.TempDir() is a SYMLINK. Path-guard mutations under TMPDIR=/private/tmp/vkreal.
- Run make check BARE. Never pipe it through a && chain — the pipe eats the exit code.

Pre-flight: `make check` BARE — expect GREEN. Then `make backup`.

Start by asking which: (A) round-3 verdict actions — if the verdict landed, work its fix
list and fire the three queued items in Blockers #1; if it was lost AGAIN, salvage from the
run journal as before; (B) mutation-gate the other packages (git, core, store — none swept);
(C) Task 5, the witness emitter, citing VD-verification-asymmetry-2dyjnd for its design
premises; (D) the owed comment-claims detector (see the 08-14 journal, claim-outruns-evidence).
```

## Worktree

- **Branch:** `main`
- **HEAD:** see `make state` — this note is committed alongside the work it describes
- **Remote:** NONE — the working tree and `~/Backups` bundles are the only copies. `make backup` after a session.
- **Uncommitted:** none expected; commit cadence is enforced at 90m

## What this session shipped

Sixteen commits. The gates are the durable output; the connector is the occasion that produced them.

**Gates built — all six caught a live defect on their first real run, and all six had a bug their
own self-test caught first:**

- `skip-lint` — a skipped test is not coverage. Caught a test that skipped 100% of the time,
  structurally, on every machine: its guard called `git branch --contains … --all`, which always
  prints the detached-HEAD pseudo-entry. `invariant-lint` verifies a cited test EXISTS, never that
  it RUNS.
- `prescription-lint` — a SPEC must not prescribe an invocation the code does not make. Caught the
  parent SPEC still prescribing a retracted, defective ref scope in two places.
- `state-freshness` + Law 10 — this note is live state. It has already fired on me twice.
- `index-check` — extracted from an inline Makefile recipe (which is why it had no self-test), and
  it no longer MUTATES the file it checks.
- `make mutants` — the mutation harness. Three calibration controls, because two of them both
  resolve through the compiler and could not tell "tests ran" from "tests silently did not".
- `link-lint` — gained the `VD-fixture-` namespace and its first self-test. It has now refused an
  id-shaped test fixture of mine four times in one session.

**Connector work:** both round-2 HIGH fabrication defects closed (paths, citations); all nine MED
items closed; two provably-dead guards deleted; the survivor triage taken to a green mutation gate.

## Where Task 4 stands

`make check` GREEN — 12 gate self-tests. `make mutants` GREEN for gitcmd. Every HIGH and MED from
rounds 1 and 2 closed, and both SPECs corrected against the findings.

None of that is acceptance. Rounds 1 and 2 each found real defects in code that was green and
self-swept, and this remediation touched `Tips`, `Commits`, both SPECs, added a new interface lock
and six gates — none of which an independent verifier has seen. **Round 3 decides.**

## Blockers / open, in priority order

1. **Round 3 landed: NEEDS_WORK — 2 HIGH, 5 MED. Closure 28 of 36, skip count 0, trend
   converging (R1: 15 found; R2: 21; R3: 12, most of them narrower).** The verdict is committed
   verbatim at `docs/verification/verdicts/round3-adjudication.md` (rounds 1–2 sit beside it,
   salvaged). Its fix list, in priority order:

   - **HIGH-1, and it is a DESIGN DECISION not a patch:** my non-greedy citation fix REGRESSED —
     it fabricates a TRUNCATED PREFIX of the real corpus id ending `-mjic4a`, cutting at `backup` —
     a six-char mid-slug segment; the exact mirror of the defect it fixed. (The fabricated form is
     quoted only inside the committed verdict: link-lint rightly refuses it anywhere else, including
     the first draft of this very note). Greedy fabricates one
     direction, lazy the other; NO regex disambiguates syntactically. My comment "Verified against
     every real suffix shape" was FALSE — I checked shapes I thought of, not the corpus
     programmatically. The disambiguation choice (resolve against the commit's own tree / tighten
     the id grammar / deterministic rule + honest limitation) is an operator call — asked.
   - **HIGH-2:** legacy `.git/info/grafts` reproduces the shallow/replace harm exactly (2 of 4
     commits vanish, whole-tree files) and nothing guards it. Fix: refuse like ErrShallow, both
     call sites, pinned fixture.
   - **MED:** Tips ⊉ Commits again by two routes (linked worktree detached HEADs; `refs/stashed`
     prefix-vs-exact mismatch with git's --exclude semantics); the false NUL premise survives in
     four sites OUTSIDE the corrected child SPEC; the withdrawn greedy narrative still ships as
     citedDecisions' doc comment; allowlist `396#136` "UNREACHABLE" is FALSE (a 20-line corrupt-HEAD
     fixture reaches it — write the test, delete the entry); § 5 "Both sides are tested" false for
     refs/changes//refs/bisect; gitcmd speclock lacks the struct-field half.
   - **LOW + gate residues:** `-n` missing from bannedArgs (needs exact-match arm); `git.go`
     payload-error `return res` still unwitnessed; `state-freshness`'s dirty-file escape never
     expires (one uncommitted byte = permanent green); `prescription-lint` success-shaped on a
     wrong root; `spec-numbering-lint` blind to § 4 table ids — the live `G-INV-5b` duplicate sits
     exactly there; two `excludedRefPrefixes` elements undefended (mutation harness cannot express
     slice-element deletion).

2. **P1 forward path (validated 2026-08-14 — full audit in the plan's Position section):**
   Tasks 0–3 DONE · Task 4 = blocker #1 above, then re-sweep + ROUND 4 (the acceptance candidate;
   the amended bar: mutants green + non-author ACCEPTABLE verdict) · Tasks 5–9 NOT STARTED
   (placeholders verified — no check-witnessed target, no spool schema anywhere) · Task 9 hard
   deadline 2026-09-18 (spec-first advisory expiry), ~5 weeks out — Task 4 must close this week.
   Task 5's design inputs are written: VD-verification-asymmetry-2dyjnd + VD-verdicts-are-artifacts-rl0rab.
3. **Retroactive mutation sweeps: `git`, `core`, `store`** (only gitcmd is gated). The store sweep
   inherits the `rows.Err()` survivor-candidate its own SPEC records as owed-not-accepted —
   validated real: the guard exists at read.go:120, nothing kills its deletion.
4. **Owed mechanisms, validated and ranked:** (a) the comment-claims detector — cannot/never/always
   in code comments has no witness, and claim-outruns-evidence is OVER Law 7's threshold under the
   honest taxonomy (4+ instances); (b) cleanroom-lint staged-content mode — it scans tracked
   content, so the verdict leak was caught one commit LATE; (c) the round-3 gate-hygiene residues
   (state-freshness dirty-note escape never expires; prescription-lint success-shaped on a wrong
   root; spec-numbering-lint blind to § 4 table ids — the live G-INV-5b dup sits there).
   VALIDATED STALE this pass, removed from the ledger: the gates scope-audit debt (superseded by
   round 3's gate-audit table); figure-provenance's narrow scope (by design per its registry row);
   index-check self-test + lesson-tag registry (shipped).
5. **Calendar (none urgent):** spec-first advisory expires 2026-09-18 = Task 9's deadline;
   meta-tax + backup advisories 2026-10-16 (backup's graduation = mechanize bundle freshness or
   delete the row). No expired advisories; 0.6's stop-check bomb was defused by graduation 08-08.
   No human-owned items pending — remote was settled local-only (bundles).

## Anything brittle

- Rounds 1 and 2's full verdicts are NOT in the repo — they live in workflow transcripts. Their fix
  lists and reasoning are in `notes/journal/2026-08-12.md` and `2026-08-13.md`.
- `docs/verification/task4-git{,-round2,-round3}.js` are committed and re-runnable.
- `make mutants` takes ~10 min per package and is NOT in `make check` — deliberately, since a gate
  that slow gets skipped. It is a package-acceptance gate, and `docs/gates.md` says so.
- `docs/allowed-survivors.txt` keys are `line#ordinal` and MOVE whenever the package changes.
  Re-sweep and refresh after any edit; a stale entry licenses a different site than its reason
  describes.
- No remote. `make backup` is the only off-machine copy.
