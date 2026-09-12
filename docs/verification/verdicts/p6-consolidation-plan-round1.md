# Adjudication — P6 consolidation plan, round 1

**Verdict:** NEEDS_WORK
**Date:** 2026-09-12
**Reviewer:** build-machine non-author adjudicator
**Reviewed artifact:** `docs/plans/P6-CONSOLIDATION-PLAN-draft1.md`
**Artifact SHA-256:** `4564e0f0aafd81d88c0d5fc0322bcb5640e61508b07db918c911911c38f3a764`
**Repository baseline:** `1e5e96cd25c78f2628e4e9f80fe312c42a240203`

## Scope and mechanical probes

This round adjudicates the plan, not P1–P5 implementation acceptance. It checked the plan's census
classes and task DoDs against the build-machine checkout. The following read-only probes were used:

- `git status --short --branch`, `git log`, `git rev-parse HEAD`, and per-package `git log -1`.
- Makefile target/dependency enumeration and `rg --files` over scripts, gates, package SPECs,
  verification artifacts, and connector surfaces.
- Exact-path existence checks for every mechanism named in `docs/gates.md`.
- Registry/source inspection in `internal/core`, route inspection in projections, and P5-kind test
  discovery across projections, gates, and twin.
- Package and mutation/verdict evidence enumeration.
- P5 measurement-name search across committed verification and journal artifacts.
- Bare `PATH=/home/thamm/go/bin:$PATH make check` and bare `make verify`, both under host execution
  because the Snap Go toolchain cannot run inside the restricted sandbox. Both exited 0.

The green commands do not close this verdict's findings. The current `make check` invokes only the
mechanisms wired into its actual dependency graph; it cannot prove absent mechanisms named by the
registry.

## Census-category adjudication

| Category | Real? | Complete and mechanically closeable as drafted? | Repository evidence |
|---|---:|---:|---|
| C1 — Doc truth | Yes | No | `README.md` still says P5 is in progress and P4 is current; historical plans contain intentionally frozen baseline claims, but the draft gives no grammar for distinguishing those from live claims. |
| C2 — Gate estate | Yes | No | The registry names seven absent blocking scripts, four absent `.claude` hooks, and five absent Make targets. `make short` runs only `hooks-test`. The category is necessary, but “every check named anywhere” has no closed extraction scope or syntax. |
| C3 — Acceptance coverage | Yes | No | There are 15 internal packages with shipped non-test Go files. P5 mutation evidence covers eight relevant packages; older verdicts bind older commits. No artifact proves every current package tree is simultaneously under the acceptance bar. |
| C4 — Connector reality | Yes | Partly | Sessions has synthetic-only acceptance; GitHub has real but deliberately narrow acceptance against `github/docs` at `d9f41c7`. The category is closeable after minimum live-corpus evidence and the narrowness inventory are defined. |
| C5 — Event-universe re-proof | Yes | Partly | Current `make verify` passes and P5 projections have intent-specific rebuild tests. Twin tests contain no P5-kind/source matrix or P5-specific replay vectors. “Full event universe” needs a generated source-kind route matrix with consume/ignore/reject expectations. |
| C6 — Intent self-coverage | Yes | No | The repository has one active native requirement with two reviewed obligations, but “since P5 acceptance” has no exact commit anchor and behavior-changing is not yet a closed path predicate. |
| C7 — Measurement debt | Yes | No | The committed estate contains the P5 measurement requirements but no complete measurement/falsifier result table. Human-time figures cannot universally cite ledger event IDs, so the required provenance form is impossible for every row. |

The seven categories cover the known debt families, but they are not a complete mechanical census.
Fresh-clone reproducibility, declared platform support, and acceptance-artifact/reference integrity
can fail without necessarily appearing in C1–C7. A consolidation phase needs an explicit category
for those surfaces; “a gap discovered outside every category enters the census” has no detector.

## Findings

### F1 — HIGH: the census has no closed machine grammar

C1's “every mechanism claim,” C2's “every check named anywhere,” and the catch-all unknown-gap rule
cannot be implemented as complete loops from the draft. The terms have no finite source set, token
grammar, canonical manifest, or historical/live classification rule. This is already material:
`docs/gates.md` references absent paths and targets while bare `make check` remains green.

Draft 2 must pin the live source files, machine-recognizable target/path/gate/test forms, Make DAG
checks, historical/frozen exclusions, and structured census row schema. It must add a clean-clone,
platform-claim, and artifact-integrity category rather than claim unknown gaps are mechanically found.

### F2 — HIGH: authorization and Task 0 order contradict each other

The status and Authorization boundary require the semantic VD and `ROADMAP.md` amendment before
Task 0. The Task 0 Census DoD requires the roadmap amendment “in the same change.” Both sequences
cannot be followed. Draft 2 must make receipt/adjudication/folding/ratification/VD+roadmap a preflight,
then begin Task 0; Task 0 must not own the already-completed roadmap amendment.

### F3 — HIGH: Task 8 leaks width into a no-new-capability phase

A production snapshot-class provider for a mutable upstream is a new source class and runtime
capability. P5 explicitly deferred it until a real feed; Draft 1 simultaneously declares no new
capability and exempts this new provider from its own non-goal. Founder input (c) is therefore a real
scope decision, not an implementation switch. Under “depth before width,” the coherent default is
to skip it and keep the contract pinned for P7+. Designating a feed requires the founder to
explicitly amend the no-width doctrine in the ratification record.

### F4 — HIGH: package acceptance is sequenced before later code changes

Task 2 requires verdicts on current code, but Tasks 3–7 can subsequently change checks, connectors,
core, projections, twin, gates, CLI, and their tests. A commit-bound verdict from Task 2 is then no
longer a verdict on the final P6 surface. Task 9's census rerun would reopen C3 after claiming Task 2
closed it. Draft 2 must move final package mutation/verdict acceptance after the last code-changing
task, define the package set from `go list`, and bind each result to a package tree object plus one
frozen final implementation commit.

### F5 — MED: the applicability rule is an example, not a closed rule

“Kernel code and gate definitions” omits behavior-affecting surfaces already present in `Makefile`,
`scripts/`, `tools/`, migrations, PowerShell entry points, and potential workflow definitions. The
draft also omits diff basis, root/merge commits, renames, deletions, and precedence. Falsifier 4 adds
a founder-set misclassification fraction that is absent from the four-input Authorization boundary.

Draft 2 must propose exact include patterns and Git diff semantics. The false-positive/false-negative
tolerance must travel with founder input (b), not become a hidden fifth input.

### F6 — MED: ambient witness capture has no named choke point or command

Task 3 requires the witnessed gate to become the default at “the repo's own choke point,” but does
not identify that point or its invocation. Interpreted as `make check`, it would conflict with the
accepted invariant that the plain gate never depends on Proofbound. Interpreted as
`make delivery-enforce`, the mechanism already exists and is explicit rather than ambient.

Draft 2 must preserve plain `make check` independence and either name an actual unskippable boundary
with a proving test or remove the ambient claim. The accepted P5 scope supports the latter: retain
`make delivery-enforce` as Proofbound's explicit controlled boundary without claiming more.

### F7 — MED: several DoDs contain unbounded judgments

“Inner loop is complete,” “real telemetry corpus,” “full event universe,” “evaluated honestly,” and
“deep-complete” lack a pinned count or result schema. Task 7 additionally requires every figure to
cite event IDs although human authoring minutes and unavailable historical measurements may have no
ledger event. Draft 2 must define minimum corpus counts, a source-kind route matrix, exact measurement
rows and outcomes (`measured`, `not-recoverable`, `not-applicable`), and allowed provenance types.

### F8 — MED: deferral can preserve a broken promise while claiming completion

The census can classify a live P1–P5 promise `defer-to-P7` and still close P6 with zero
`close-in-P6` rows. That is not “fully complete P1–P5.” A deferral or wontfix closes a row only when
a decision record explicitly supersedes, narrows, or retires the original promise and updates every
live authority in the same change. Otherwise it remains open.

### F9 — MED: advisory closure overreaches and conflicts with the registry law

The measurements target zero advisory rows and Task 9 requires zero “unresolved advisory expiries,”
but the constitution permits dated, owned, unexpired advisories. Zero advisories is not the ratified
rule and can turn consolidation into gold-plating. Draft 2 must require zero expired advisories and
explicit disposition of advisories due at P6 start; future valid advisories may remain. The legacy
identity alias is due at P6 start and must be handled explicitly.

### F10 — LOW: P5-kind replay vectors need compatibility wording

Task 4 says twin replay proof carries pinned vectors for the P5 kinds. The existing
`vera.replay.v1` identity and pinned vectors are frozen. Draft 2 must require additional distinct
fixtures/vectors without editing or renaming any frozen vector, and classify every source-kind pair
as consumed, ignored, or rejected rather than assume every mechanism consumes every event.

## Required disposition

Draft 2 should retain the consolidation doctrine and C1–C7 intent, add the missing reproducibility
category, and fold F1–F10. After that, stop for the four founder inputs. No implementation task is
authorized by this verdict.
