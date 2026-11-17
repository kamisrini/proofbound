# `intent.sdd` provider fixtures (R2 M1) — DRAFT, written before the provider

Test data only; no code. Written against `docs/plans/R2-sdd-provider-spec-draft3.md` and extended for
drafts 4-7 (`docs/plans/R2-sdd-provider-spec-draft7.md` is current; invariant numbers below are draft 7's) while those drafts are under
non-author review, so **the profile shape here is provisional** and must be updated
if the reviewed contract changes it. These become the M1 tests' inputs once
`kernel/internal/connector/intent/sdd/` gains its code and its `SPEC.md`.

## Format

`openspec-history/` and `speckit-history/` are **commit sequences**: each `NN-label/` holds the full
repository tree for one commit (`tree/`) and its commit message (`MESSAGE`). A step may also carry
`PARENTS` (whitespace-separated earlier step labels, first parent first; empty = root commit; absent =
the previous step) and `REF` (the branch it is written to; absent = `main`). A test helper replays
them in order into a temporary git repository (one commit per step), so every lifecycle test runs
over a history — never a single snapshot, which is how drafts 1 and 2 passed their own tests while
real histories broke them. `branches/` is a commit **graph** (draft 4 §13). `unmappable/` and
`fenced/` hold single trees.

## Steps and what they exercise

| Fixture step | Reproduces (real-history shape) | Invariants (draft 7) |
|---|---|---|
| openspec 01-baseline | main spec present but never ingested | INV-30 |
| openspec 02-draft-proposal-only | change with zero requirement entries (38 of 135 real changes at some state) | INV-28, INV-59 (Resolve → ErrNoChangeIntent) |
| openspec 03-deltas-added | change with two ADDED requirements | INV-9, INV-44 |
| openspec 04-implement | the implementing commit a verdict binds | INV-31, INV-32 |
| openspec 05-within-change-edit | requirement removed + statement changed within a change (83 removals in 17 changes / 39 statement changes in 9, provider walk) | INV-44, INV-5 |
| openspec 06-archive | normal archive: dated move + main-spec merge | INV-20, INV-32 (archived earlier ⇒ not resolvable later) |
| openspec 07-squash-merged | change arriving already archived with its code (12 / 108 real on the provider walk) | INV-32 (archived by this commit ⇒ resolvable), INV-31 |
| openspec 08-rework-active | RENAMED + MODIFIED (on the new name) + REMOVED in one delta | INV-9, INV-43 (no collision: section prefixes) |
| openspec 09-archive-skip-specs | `--skip-specs` archive: main spec not updated | INV-20 |
| openspec 10-main-spec-deleted | a main spec deleted outright | INV-20, INV-30 |
| speckit 01-feature-spec | FR / SC ids; no tasks yet ⇒ no change intent | INV-10, INV-42 |
| speckit 02-tasks-added | sibling appears ⇒ feature-level change intent | INV-10 |
| speckit 03-in-place-edit | `FR-001` edited in place + status Draft→Approved ⇒ new record | INV-44, INV-42 |
| unmappable/duplicate-obligation | two names mapping to one obligation id | INV-43, INV-46 |
| unmappable/reused-name | one name both active and archived | INV-45 |
| branches 01-root-with-archive | root commit already carrying an archived change | INV-32 (archived by the root ⇒ resolves) |
| branches 02-06 | edit / claim (`Intent: sdd:add-x`) / revert on `feature`, merged `--no-ff`, TREESAME to 02 on `openspec/` | INV-31 (a path-limited HEAD walk misses 03-05) |
| branches 07-08 | add-x archived on `side`, then merged | INV-32 (the merge does not archive; 07 does) |
| branches 09-unmerged-claim | `Intent: sdd:add-y` on `wip`, never merged | INV-31 (on the `--all` walk only) |
| branches 10-13 | add-z archived leaving its delta byte-identical; a claim during the leftover; leftover removed (OpenSpec `7ced2a87` .. `df12368f`) | INV-20, INV-45, INV-35(a), INV-46 |
| branches 14-stale-tip | a ref tip (`stale`) left on the leftover state (OpenSpec: 4 remote tips) | INV-35(b) |
| fenced/improve-onboarding | a column-0 `## TL;DR` inside a fence, followed by two more requirements (`improve-init-onboarding` shape) — masked 3 obligations, unmasked 1 | INV-39 (discriminating) |
| fenced/document-renames | an indented RENAMED example inside a fence (`adopt-delta-based-changes` shape) — 1 obligation either way | INV-39 (neutral control) |
| fenced/fence-marker, fence-length, fence-info-backtick | §4.2 fence precision: a `~~~` line inside a backtick fence, a shorter closer, a backtick in a backtick opener's info string — each kills a naive toggle mask | INV-39 |
| head-on-leftover | HEAD (`main`) itself ends on the dual-presence state | INV-35(c) |
| expected-skips.json | the exact Skip set of every fixture | INV-46 |
| expected-obligations.json | per fenced case: obligations under OpenSpec's mask, a naive toggle, and no mask | INV-39 |

## Re-deriving the evidence

`derive/history_probe.py` is the probe that measured draft 3 §0 through draft 7 §0 (it implements draft 7 §4.2 —
OpenSpec's fence, section and requirement rules — and prints every count draft 7 cites; `--pins` regenerates `expected-obligations.json`) over the full history of
`Fission-AI/OpenSpec` (run: `python3 -I derive/history_probe.py <openspec-clone>`). It is kept so the
fixture rows can be checked against a fresh measurement; it is documentation, not part of the build.
