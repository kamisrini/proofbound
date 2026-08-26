# `internal/specfirst` SPEC

The spec-first gate requires every kernel package to publish a `SPEC.md` with
at least one proving-test citation that resolves to a real test in that
package.

## Invariants

| Invariant | Proving test |
|---|---|
| SF-INV-1 | Every internal package has a SPEC and a resolvable proving test citation | specfirst_test.go::TestEveryInternalPackageHasSpecAndProof |

