# Canonical EOD prompt

Use this one prompt at the end of a workday. It owns the complete durability closeout; do not
create a separate manual resume-command sequence for the user.

```text
End the workday according to the Proofbound durability rules.

Inspect the current repository state, today’s commits, tests, journals, notes, and session
artifacts. Do not rely on conversation memory. Summarize completed, open, blocked, and unverified
work. Look for recurring process mistakes and institutionalize only lightweight, repeatable fixes.

Append the factual entry to today’s journal, including LESSON when applicable. Rewrite
notes/state.md in place with the current date, branch status, completed work, open blockers,
unverified assumptions, the exact next action, and any institutionalized improvement. The exact
next action is durable agent state; the user must not be asked to run a separate resume prompt.
Per Build Law 10, put `## Resume — YYYY-MM-DD` first after `As of`, with the exact next action and
immediate blockers there rather than buried later in the file.
Also rewrite the ignored local file `notes/overall-status.md` with the current roadmap phase status,
capability progress, built and unbuilt components, open blockers, verification health, and next
authorized action. Keep it internal and do not commit it.
For every newly postponed, deferred, blocked, or owner-pending item, add or update one row in
`docs/deferrals.tsv` with a stable `DEF-*` ID, canonical lifecycle status, human `hold_type`, owner,
review trigger, unblock condition, and evidence paths. Run `scripts/deferrals.sh --check`; use
`status=OPEN` with a hold type such as `POSTPONED`, `AWAITING_OWNER`, or `PENDING_PREREQUISITE`.
Do not delete closed rows; mark them `RESOLVED`, `REJECTED`, or `RETIRED` with closing evidence.

Run the appropriate verification gate, especially bare `make check`; if generated-index or
invariant-lock freshness fails, regenerate the derived artifact and rerun the gate. Treat expected
negative-test messages separately from actual failures. Commit only coherent completed changes; never commit
unfinished or unverified implementation. Push if authorized and authenticated. Create and verify
the dated Git bundle fallback. Confirm the final repository is clean and report the commit hash,
push status, verification result, backup path, unresolved work, and tomorrow’s first action.
After adding documentary verdicts, check `scripts/p6-artifact-integrity.sh --check`; if its
generated packet is stale, run `--write` and recheck it before the closeout commit.

If there are no meaningful changes, do not manufacture a commit; still refresh state if needed and
report that the repository was already complete and clean.
```

The generated next-action text in `notes/state.md` is for the next agent’s inspection and handoff;
it is not an additional user command.
