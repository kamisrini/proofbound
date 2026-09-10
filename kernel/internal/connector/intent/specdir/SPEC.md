# `specdir` foreign-provider contract

## 1. Convention and boundary

The provider name is `specdir`; emitted source is `intent.specdir`. This P5 mapper supports exactly
one committed convention rooted at `specs/`:

```text
specs/<feature-slug>/requirements.md
specs/<feature-slug>/change-intent.md   # optional
```

The reader boundary and commit-tree semantics are the same as the native provider. No business
decision is synthesized. Requirements therefore emit an empty `authorized_by` list and reports must
render `authorization: undeclared`.

## 2. Requirement mapping

`requirements.md` is strict UTF-8 Markdown:

```text
# Requirement: <BR-id>
Status: active
Owner: <declared identity>

## Obligations
- [O-1] <non-empty one-line statement>
```

The exact source bytes are the artifact digest. Status is `proposed`, `active`, `superseded`, or
`retired`. Metadata lines are single, ordered, and unknown metadata is rejected. Obligation ids and
statements map without invention into `proofbound.requirement.v1`. Feature slug and record id are
independent provider-scoped identities but each must be unambiguous and stable.

## 3. Optional change-intent mapping

`change-intent.md` is:

```text
# Change Intent: <CI-id>
Status: accepted
Sponsor: <declared identity>

## Targets
- implements intent.specdir:<BR-id>@<64-lowercase-hex>#O-1[,O-2]
```

Relations are closed to `implements`, `modifies`, `repairs`, or `retires`. Targets are sorted and
deduplicated after parsing; the serialized canonical payload is stable across runs. `specdir` does
not author VD constraints or cross-provider supersession in this first mapping, though the family
model accepts both from other providers.

## 4. STOP fixture

`testdata/unmappable/conditional-obligation.md` expresses an obligation whose identity changes by
runtime branch and therefore has no stable obligation id. Mapping it returns `intent.ErrUnmappable`.
The mapper must not add a union type, generated id, condition field, or other canonical-schema change.

## 5. Invariants and test derivation

| Invariant | Statement | Proving test |
|---|---|---|
| SPD-INV-1 | Exact convention files map to canonical BR/CI payloads with stable digests | specdir_test.go::TestValidVectors |
| SPD-INV-2 | Requirement mapping fabricates no business decision or authorization | specdir_test.go::TestAuthorizationRemainsUndeclared |
| SPD-INV-3 | IDs, metadata order, obligations, targets, UTF-8, paths, and digests fail closed | specdir_test.go::TestHostileFixturesFailClosed |
| SPD-INV-4 | Canonical payload bytes are identical across repeated runs | specdir_test.go::TestCanonicalizationStability |
| SPD-INV-5 | Resolution reads the named commit tree and returns its observed digest | specdir_test.go::TestResolveUsesCommitTree |
| SPD-INV-6 | Conditional unstable identity returns STOP without changing the model | specdir_test.go::TestUnmappableFixtureStops |

