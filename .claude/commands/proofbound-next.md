---
description: Orient from durable state and propose the next authorized task
---

Orient me and propose the next task. Do not start work yet — this command ends with a
recommendation, not an edit.

Follow the Proofbound session protocol (CLAUDE.md):

1. Read `notes/state.md`, especially the first `## Resume` section — the exact next action and
   immediate blockers live there. Treat it as agent handoff state.
2. Read `git status -sb` and the last few commits. Note whether the worktree is dirty and whether
   `notes/state.md` is stale (`scripts/state-freshness.sh`).
3. Determine the gate's honest state. Do NOT assume green from the state note — a prior session may
   have recorded PASS before a regression. If a fast check is cheap, note it; otherwise say the gate
   status is unverified this session and name `make check` as the way to confirm.
4. Respect the standing stops, which override any eagerness to make progress:
   - P7 and any `status=OPEN` deferral (`docs/deferrals.tsv`) are DORMANT until the owner explicitly
     authorizes them in a work request. Do not start, and do not nag the owner about open deferrals.
   - Governed-path changes (see `scripts/intent-applicability.sh`) expect an owner-approved change
     intent before they are committed.
   - No actor grades its own work (Law 9): anything I authored needs a non-author verdict to count
     as accepted.
5. Cross-reference `ROADMAP.md` for the current phase and its Definition of Done.

Then give me, concisely:
- Where we are (phase, gate state, dirty/stale flags, anything uncommitted and why).
- A short ranked list of what can be picked up next, each tagged **unblocked** or
  **needs-owner-authorization**, with the one real decision (if any) each carries.
- Your single recommended next action and why.
