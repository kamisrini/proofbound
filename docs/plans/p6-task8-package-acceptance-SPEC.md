# P6 Task 8 — final package acceptance SPEC

## Scope

Task 8 is the final package acceptance sweep after all P6 code-changing tasks. It freezes one
implementation commit and evaluates the current production package universe against the existing
package acceptance bar: calibrated mutation evidence plus a non-author verdict. It adds no product
capability and does not reinterpret older verdicts as current acceptance.

The production package set is every package returned by `go list ./internal/...` that contains at
least one non-test `.go` file. At the current inventory this is 16 packages. `internal/specfirst`
is returned by `go list` but contains only `SPEC.md` and `_test.go`; it is explicitly classified as
`not-in-production-package-set` and receives no mutation requirement.

## Frozen package evidence

For each production package, the final packet must record:

- the one frozen implementation commit;
- the exact package tree object from `git rev-parse <commit>:kernel/<package-path>`;
- candidate, killed, invalid, and survived mutation counts from the calibrated `make mutants`
  route, with neutral/invalid/lethal controls identified; and
- an exact path to a committed non-author verdict that names the same commit and package tree
  object. A verdict bound to an earlier commit or only to a package name does not close the row.

The packet must separately list test-only packages and must not silently exclude a package because
it has no mutation candidates. An undeclared survivor, invalid calibration, missing tree object,
missing package, or missing exact non-author verdict keeps the package row open.

## Invariants and test derivation

| ID | Statement | Proving test |
|---|---|---|
| T8-INV-1 | The production package universe is exactly the current non-test Go package set | p6-task8-package-universe.test.sh::production-and-test-only-universe |
| T8-INV-2 | Test-only packages are explicitly classified and excluded from mutation acceptance | p6-task8-package-universe.test.sh::production-and-test-only-universe |
| T8-INV-3 | Every production package binds one frozen commit and exact tree object | final acceptance checker and package packet |
| T8-INV-4 | Every mutation result has calibrated counts with no undeclared survivors | final acceptance checker and mutation result packet |
| T8-INV-5 | Every package is covered by a committed non-author verdict for the same commit and tree | final acceptance checker and committed verdict artifact |

## Acceptance

Task 8 closes only after the final code-changing commit is frozen and all 16 production rows have
calibrated mutation results, zero undeclared survivors, exact tree-object bindings, and committed
non-author verdict coverage. Then, and only then, the C3 census rows may cite the packet and the
final package acceptance sweep may proceed to the Task 9 consolidation round.

The mutation sweep is deliberately not part of plain `make check`; package acceptance is an
explicit release gate. No author-authored verdict can satisfy the non-author requirement.

## Non-goals

- No new implementation capability, package, provider, event kind, or acceptance level.
- No reuse of a prior verdict bound to a superseded commit.
- No self-authored “non-author” verdict and no claim that mutation alone proves semantic correctness.
- No Task 9 close or P6 deep-complete claim until package acceptance and the final independent round
  are both committed.
