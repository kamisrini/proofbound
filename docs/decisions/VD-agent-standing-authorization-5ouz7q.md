# VD-agent-standing-authorization-5ouz7q: owner grants a standing authorization for a product push

**Status:** Accepted
**Date:** 2026-10-31
**Context:** The owner (Srinivas Kamisetty) is the founder and wants to finish building Proofbound to
sell to companies, and asked why the build keeps stopping. The honest answer: the stops are
self-imposed by this repo's own constitution — P7 and every `OPEN` deferral are dormant until the
owner explicitly authorizes them, governed-path changes expect an owner-approved change intent per
change, and no actor grades its own work (Law 9). With the substantive product work (P7) all behind
those gates and `/proofbound-next` designed to end in a recommendation, the owner was the bottleneck
for essentially all forward motion, and meta-tax sat at 71% (scaffolding over product). The owner
chose to turn the per-step gating off.
**Decision:** A standing authorization is in force until the owner revokes it:
- The agent MAY self-authorize governed-path commits under the owner's name (`Authorized-By: Srinivas
  Kamisetty`) without a fresh per-commit change-intent exchange, for work within an agreed product
  push.
- The agent MAY work across the normally-dormant P7 and `OPEN`-deferral gates as part of that push,
  rather than stopping at each. The current push direction is: make P7 actually run and become
  demonstrable (owner's 2026-10-31 instruction).
- The agent STILL STOPS at genuine decision forks and irreversible/outward-facing actions: shipping
  or redistributing a third-party binary, approving a license/provenance go/no-go, deleting owner
  data, publishing, or any step whose cost the owner alone should weigh. [[VD-fix-discipline-0e0tnz]]
- Law 9 is UNCHANGED: the owner remains the non-author acceptor, and adversarial rounds run via
  dispatched non-author reviewers so verdicts do not block on the owner
  ([[VD-review-ceremony-tiers-v2b6kp]], [[VD-verification-asymmetry-2dyjnd]]).
**Consequences:** Forward motion stops waiting on per-step permission, which is what the owner wants.
The cost, stated plainly: the owner sees less of each governed change before it lands, so a wrong
call is caught later (at acceptance or by a reviewer) rather than at the gate. The standing stops
above are the backstop, and the durable record (journal, `notes/state.md`, verdicts) remains the
audit trail. This does not weaken any gate or law — `make check`, the hooks, and Law 9 still apply;
it changes only who must say "go" before each governed commit and dormant-gate step.
**Revisit when:** the owner revokes or narrows it; the product push's direction changes; a governed
change lands that the owner would have stopped at the gate (evidence the stop list is too narrow); or
a session uses it to justify an irreversible/outward action that should have been a fork.
