# Proofbound — Full Value Tour (Release 1)

**Who this is for:** someone who wants to see *everything Release 1 does* on a realistic project —
not the 5-minute taste, but the whole trust surface — without reading Proofbound's own source.

**What it shows:** a permanent trust record, decisions bound to the commits that implement them, the
full **decision → requirement → change → commit** intent chain, independent **review verdicts**
(including a *red-verdict chain*: a fix that introduced a new defect), and a record that **rebuilds
identically** from scratch.

> **Run it (≈1 minute):**
> ```bash
> # build the binary first (see DEMO.md §1), then:
> bash docs/releases/r1/full-tour.sh                    # narrated
> # or the same thing without narration, leaving an app to explore:
> bash docs/releases/r1/seed-demo-app.sh
> ```
> This tour is just [`seed-demo-app.sh`](seed-demo-app.sh) run in `--tour` mode (one source of truth).
> It builds its **own** throwaway sample app (default `~/proofbound-demo-app`) and records the full
> trust surface into it — so it runs from **any directory** on **any computer** and doesn't touch
> your other projects. Run it on a Linux filesystem (WSL home), **not** a Windows `/mnt/d` path.
> First run downloads the embedded PostgreSQL once (needs internet). Set
> `PROOFBOUND_BIN=/path/to/proofbound` if it isn't on your PATH; pass a directory to choose where it
> seeds. The seeded app is left in place — explore it with `proofbound report week`, `report intent`,
> and friends.

The 5-minute [DEMO](DEMO.md) shows the foundation; this tour shows the whole thing. Proof IDs
(`proof=01M…`) are time-based and differ every run — that's expected.

---

## What the tour builds

A tiny invoice-tax library, through a *realistic* history: a business decision, a requirement it
must satisfy, a change that implements it, and two independent reviews that catch real defects —
then it reads the whole thing back as a durable record.

## 1. The permanent trust record

```
proofbound sync git
→ listed=1 appended=1 existing=0
```

Every commit becomes a permanent, content-addressed record with a **proof ID bound to its exact
code**. Re-running appends only what's new (`appended=0` the second time) — the record can't drift.
**Business value:** an honest, tamper-evident history of what was built, maintained for you.

## 2. The intent chain — *why this change exists, and what it must guarantee*

Three small records state the **why** and the **what**: a business **decision**, the **requirement**
it creates (with an obligation like *"tax must never be negative"*), and the **change intent** that
implements it. Each record is content-addressed, and each child cites its parent's exact digest — a
tamper-evident chain. The commit that does the work cites the change intent in a one-line trailer.

```
proofbound intent check --commit <impl>
→ commit=6e8d0f… intent=records:CI-clamp-negative-tax-8c9d0e@c7d300e… proof=01M…

proofbound report intent CI-clamp-negative-tax-8c9d0e
→ intent=CI-clamp-negative-tax-8c9d0e … state=IMPLEMENTED_UNVERIFIED
    target=…BR-tax-never-negative…#O-1 relation=implements review=UNREVIEWED outcome=UNVERIFIED
            statement="Computed tax must never be negative for any invoice line."
    commit=6e8d0f… deployment=missing
```

This is the strongest "**who approved this, and what was it required to do**" story — the entire
chain from a business decision down to the exact commit, with proof IDs at every link. Note it is
**honest**: `review=UNREVIEWED outcome=UNVERIFIED deployment=missing`. Proofbound distinguishes
*implemented* from *verified* and never pretends work is proven before it is.
**Business value:** you can always answer "why does this code exist and who owns that decision?" —
and you can trust the answer because it's bound to the code, not written in a wiki that rots.

## 3. Independent reviews — *what was found, not just what ran*

Ordinary CI tells you a check *ran*. Proofbound records what reviewers **found**. The tour files two
independent review verdicts: one catches a false "rounding" claim; the second catches a **new defect
that the fix itself introduced** (an overflow).

```
proofbound report week
→ reviews count=2
   - tax-rounding-truncates  status=NEEDS_WORK severity=MED  reviewed=…
   - tax-rounding-overflows  status=NEEDS_WORK severity=HIGH reviewed=…
  red_verdict_chains count=1
```

That `red_verdict_chains count=1` is the signal a green CI badge can never give you: **a fix
introduced a new defect.** This is Proofbound's whole reason for preferring *rebuild-from-the-rules*
over endless patching — on real work, patching introduced one-to-three new defects per fix.
**Business value:** you can see defects *being introduced by fixes*, over time — the failure pattern
that quietly erodes AI-built software.

## 4. Decisions bound to commits

A plain decision record (`VD-rounding-policy-…`) cited by the commit that implements it shows up on
that commit forever:

```
→ … fix(tax): overflow-safe half-up rounding   decisions=VD-rounding-policy-3f2a1b   proof=01M…
```

**Business value:** the *why* travels with the *what*, permanently — even if the commit message is
later rewritten or the decision file moves.

## 5. CI witnesses — tamper-evident proof that a check actually *ran*

The tour seeds two **witnesses** — a passing `make test` and a failing `make build` — bound to the
exact commit. A witness is produced *by the check*, never asserted by an agent: it carries the
command, the exit code, the timing, and a content hash of the output.

```
proofbound report week
→ checks count=2
   - 01HQRST9…ABC  pass  duration_ms=3000  proof=01M…
   - 01HQRST9…ABD  fail  duration_ms=1200  proof=01M…
```

**Business value:** a green CI badge is a *claim* that evaporates; a witness is **evidence** — that
a specific check ran, with a specific result, on a specific commit, permanently. Notice the **fail**
is recorded too: failures are first-class evidence, not something a passing re-run erases. (In a real
project, a one-line wrapper writes these as your tests/CI actually run; this tour hand-seeds them so
you can see the shape.)

## 6. The record is reproducible, not a snapshot

```
proofbound rebuild          # drop every projection, rebuild from the raw ledger
→ first commit's proof ID is IDENTICAL before and after
```

The entire trust record rebuilds to the exact same proof IDs. **Business value:** nothing is a
hand-maintained dashboard that can disagree with reality — every view is recomputable from the one
source of truth. If it ever needed reconciling, that would prove you had two truths.

---

## What a sample project does *not* show (by design)

Being honest about the edges:

- **`proofbound verify`** checks Proofbound's *own* delivery boundary. On an ordinary app repo it
  says `no make check witness found in spool` — **expected**. Your trust surface is
  `sync` + `report` + `rebuild`; `verify` is for Proofbound's self-hosting workflow.
- **`proofbound gates …`** evaluate declarative rules over the check witnesses. The tour now seeds
  witnesses (§5), so the missing piece is the gate *definitions* (`gates/*.yaml`) plus a
  canary→enforce promotion — a layer beyond this tour.
- **`proofbound sync github` / `report github`** need a real GitHub repo and a `GITHUB_TOKEN`.
- **`proofbound twin replay`** is the advanced replay engine (it spins up a private PostgreSQL) —
  the foundation under behavior-locks, beyond this tour.

---

*Start smaller: [5-minute demo](DEMO.md). Go deeper on concepts: [User Guide](USER-GUIDE.md). Scope
and caveats: [Release Notes](RELEASE-NOTES.md).*
