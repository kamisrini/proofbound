# VD-sdd-m1-openspec-first-2026-11-17: R2 M1 maps OpenSpec only, over a literal-path profile

**Status:** Accepted
**Date:** 2026-11-17
**Context:** The `intent.sdd` provider contract (R2 M0) went through seven non-author review rounds;
the verdicts are under `docs/verification/verdicts/r2-m0-review-round*`. The findings shrank from 2 HIGH
(round 3) to 2 MED (round 6). Then they rose again to 3 MED in round 7, because each review reached a
corner the contract had not specified yet:
- Spec Kit has no `Resolve` rule and no change-intent id formula;
- tenant profile path strings reach git's pathspec parser, where `:(glob)`, `../` and an empty prefix
  make git fail, which wedges the sync permanently.

The forcing question: keep specifying every kit and every path form before any code exists, or narrow
what M1 must get right.

**Decision:** The owner chose to narrow it, on 2026-11-17.
1. **M1 implements the OpenSpec kit only.** The profile's `kit` must be `openspec`; `speckit` is a
   profile violation until M2.
   - Spec Kit mapping moves to **M2**, still within R2's "Spec Kit subset" (`VD-sdd-change-scoped-r2-k7q3wz`
     is unchanged): its record and change-intent rules, its `Resolve` rule, and its id formula.
   - The Spec Kit fixtures stay committed, as M2's inputs.
2. **Profile paths are literal.**
   - Every path in the profile must match `^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`, with no `.` or `..`
     segment. That covers `changes.root`, `changes.archive`, and the literal prefix of each
     `requirements.paths` glob. `requirements.paths` may use `*` only, as a whole path segment.
   - Anything else is `profile-invalid`.
   - Every git path lookup the `sdd` reader makes runs with literal pathspecs (`--literal-pathspecs`,
     or `GIT_LITERAL_PATHSPECS=1`), so no tenant string is ever interpreted as pathspec magic.

**Consequences:**
- **Easier:** M1's contract covers one kit, against the reference repository it was measured on.
  Pathspec handling is closed by construction instead of by enumerating git's magic forms.
- **Harder:** Spec Kit tenants cannot be warranted until M2. M2 needs its own spec section and its own
  non-author review.
- **Given up:** shipping both kits in one milestone. Some tenant layouts are also excluded, namely
  paths with spaces or other characters outside the literal grammar; those tenants get a
  `profile-invalid` skip.

**Revisit when:**
- M2 starts, at which point Spec Kit is specified against its own fixtures and the profile `kit`
  enum widens;
- or a real tenant repository needs a path outside the literal grammar.
