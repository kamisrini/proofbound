# Spec: the launch demo "the agent goes off-script"

**Status:** draft 1, 2026-11-23, accepted by the owner ("lets do it, go ahead"). Documentation only:
no kernel code (the one kernel bug it found is fixed under its own spec,
[warranty-head-reviews-spec.md](warranty-head-reviews-spec.md)).

## 0. Why

The README now leads with agents: at their speed and volume, nobody can read every line an agent
writes, so the question "did this change do what was asked, and who checked?" needs a record, not
memory. The tour shows every feature on a human-written change. This demo shows the same machinery
on the case the pitch is about: an agent's change that is not what its spec asked for.

## 1. The story (three acts, one Spec Kit project)

1. **The agent works.** A spec with two promises (CSV export; a header row) declares its area in
   `scope.paths` (`src/export.sh`, `tests/`). The agent, under `proofbound task start`, makes one
   commit that claims the spec, is co-authored by Claude, and:
   - edits `src/pricing.sh` (the member discount, 10 → 25), outside the area;
   - leaves out the header row;
   - ships a test that cannot fail.

   The agent says the tests pass, and they do. `check`: NOT WARRANTED (nobody has reviewed it).
2. **An independent reviewer.**
   - `review draft` names `src/pricing.sh line 3` as outside the area.
   - The reviewer breaks the code on purpose; the agent's test still passes, so it proves nothing.
   - The review is sealed: FR-002 NOT_SATISFIED and the pricing line rejected, so the verdict is
     NEEDS_WORK.
   - `check` reads BLOCKED. The Audit room shows C1 and C10 exceptions and Claude in "Who did the
     work". The release bundle reads NO WARRANTY.
3. **The redo, from the spec,** on a new branch:
   - test first, and the new test fails on the old code;
   - the change stays inside the area;
   - the reviewer's break-it check now fails the test;
   - the review is sealed: `check` WARRANTED, and the bundle VALID against a fresh clone.

   The rejected branch is kept, with its signed NEEDS_WORK review.

## 2. Honesty rules (the claims must match what ships)

- Proofbound does **not** detect a test that cannot fail. The reviewer finds it by breaking the code.
  What Proofbound adds is that the reviewer is someone other than the agent and the spec owner, their
  finding is signed, and nothing is warranted until it is fixed. Automatic mutation checks are named
  as future work, not claimed.
- The out-of-area line **is** found by Proofbound, from git, and cannot be waved through without a
  signed decision.
- Git authorship (the Claude co-author) is self-declared; the page says so, and points to C9 signing.
- The reader plays the agent by pasting its edits; the page says so and says that with Claude Code
  the same commit comes from the agent through the same hooks.

## 3. Deliverables

- `docs/releases/r3/agent-demo.sh`, which runs every step for real (like `tour.sh`).
- `docs/releases/r3/DEMO-AGENT.md`, the hands-on page in the DEMO-MANUAL style: paste-safe `bash`
  blocks, `text` output blocks taken from a real run, and for each step what it does, why, and its
  business value.
- Links from `docs/releases/r3/README.md`.

**Acceptance:** the script runs green from a clean workdir; every `bash` block on the page is run, in
order, in a fresh shell, and its output matches the `text` block (hashes and keys aside).
