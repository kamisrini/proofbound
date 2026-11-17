# `intent.sdd` provider fixtures (R2 M1) — DRAFT, written before the provider

Test data only; no code. Written against `docs/plans/R2-sdd-provider-spec-draft3.md` and extended for
`docs/plans/R2-sdd-provider-spec-draft4.md` (round-3 fixture conditions) while those drafts are under
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

| Fixture step | Reproduces (draft 3 §0 row) | Invariants |
|---|---|---|
| openspec 01-baseline | main spec present but never ingested | INV-30 |
| openspec 02-draft-proposal-only | change with zero requirement entries (34 real) | INV-28 |
| openspec 03-deltas-added | change with two ADDED requirements | INV-9, INV-25 |
| openspec 04-implement | the implementing commit a verdict binds | INV-24, INV-27 |
| openspec 05-within-change-edit | requirement removed + statement changed within a change (47 / 10 real) | INV-25, INV-5 |
| openspec 06-archive | normal archive: dated move + main-spec merge | INV-20, INV-27 (archived earlier ⇒ not resolvable later) |
| openspec 07-squash-merged | change arriving already archived with its code (16 / 108 real) | INV-27 (archived by this commit ⇒ resolvable), INV-24 |
| openspec 08-rework-active | RENAMED + MODIFIED (on the new name) + REMOVED in one delta | INV-9, INV-6 (no collision: section prefixes) |
| openspec 09-archive-skip-specs | `--skip-specs` archive: main spec not updated | INV-20 |
| openspec 10-main-spec-deleted | a main spec deleted outright | INV-20, INV-30 |
| speckit 01-feature-spec | FR / SC ids; no tasks yet ⇒ no change intent | INV-10, INV-4 |
| speckit 02-tasks-added | sibling appears ⇒ feature-level change intent | INV-10 |
| speckit 03-in-place-edit | `FR-001` edited in place + status Draft→Approved ⇒ new record | INV-25, INV-4 |
| unmappable/duplicate-obligation | two names mapping to one obligation id | INV-6, INV-14 |
| unmappable/reused-name | one name both active and archived | INV-29 |
| branches 01-root-with-archive | root commit already carrying an archived change | INV-32 (archived by the root ⇒ resolves) |
| branches 02-06 | edit / claim (`Intent: sdd:add-x`) / revert on `feature`, merged `--no-ff`, TREESAME to 02 on `openspec/` | INV-31 (a path-limited HEAD walk misses 03-05) |
| branches 07-08 | add-x archived on `side`, then merged | INV-32 (the merge does not archive; 07 does) |
| branches 09-unmerged-claim | `Intent: sdd:add-y` on `wip`, never merged | INV-31 (on the `--all` walk only) |
| branches 10-13 | add-z archived leaving its delta byte-identical; a claim during the leftover; leftover removed (OpenSpec `7ced2a87` .. `df12368f`) | INV-20, INV-29, INV-35, INV-36 |
| fenced | a RENAMED example inside a code fence of a MODIFIED requirement (`adopt-delta-based-changes` shape) | INV-39 |

## Re-deriving the evidence

`derive/history_probe.py` is the probe that measured draft 3 §0 and draft 4 §0 (it prints every count
draft 4 cites; code fences are masked as OpenSpec masks them) over the full history of
`Fission-AI/OpenSpec` (run: `python3 -I derive/history_probe.py <openspec-clone>`). It is kept so the
fixture rows can be checked against a fresh measurement; it is documentation, not part of the build.
