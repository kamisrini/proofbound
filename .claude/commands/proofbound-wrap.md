---
description: Run the canonical end-of-day durability closeout
---

End the workday according to the Proofbound durability rules.

The canonical protocol is `docs/eod-prompt.md` — it is the single home of the closeout steps
(Law 2). Read it now and execute its procedure exactly, in order. Do not improvise a parallel
sequence and do not skip steps.

In summary, that protocol has you:

1. Inspect the real repository state — today's commits, tests, journals, notes, session artifacts.
   Do not rely on conversation memory. Summarize completed, open, blocked, and unverified work.
2. Append the factual entry to today's journal (`notes/journal/YYYY-MM-DD.md`), including a
   greppable `LESSON: class=<registered-class> response=<path>` line for any recurring mistake, so
   it compiles to enforcement rather than decaying as prose (Law 7).
3. Rewrite `notes/state.md` in place with the current date, `## Resume — YYYY-MM-DD` first, the
   exact next action and immediate blockers there, then branch status, completed/open/blocked/
   unverified work, and any institutionalized improvement. Also rewrite the gitignored
   `notes/overall-status.md`.
4. Update `docs/deferrals.tsv` for anything newly deferred/blocked/owner-pending and run
   `scripts/deferrals.sh --check`. Never delete closed rows; mark them resolved/rejected/retired.
5. Run the gate — bare `make check`. If generated-index or invariant-lock freshness fails,
   regenerate the derived artifact and rerun. Treat expected negative-test messages as non-failures.
   After adding documentary verdicts, run `scripts/p6-artifact-integrity.sh --check`.
6. Commit only coherent, completed, verified changes. Governed-path changes expect an
   owner-approved change intent and an `Intent:` trailer. Push only if authorized. Create and
   verify the dated git bundle (`make backup`).
7. Report: commit hash, push status, verification result, backup path, unresolved work, and
   tomorrow's first action.

If there are no meaningful changes, do not manufacture a commit — refresh state if needed and
report that the repository was already complete and clean.

**Report honestly.** If `make check` is red, say so with the failing output; if a step was skipped,
say which and why. A green claim requires the kernel half to have actually completed.
