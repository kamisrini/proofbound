# P6 Task 1 mechanism-truth SPEC

## Scope

This specification closes C1/C2 rows from the P6 census without adding product capability. It
restores build-law mechanisms promised by `CLAUDE.md` and `docs/gates.md`, makes their self-tests
discriminating, reconciles user-facing command documentation, and explicitly retires claims that
were optional or superseded. It does not alter frozen wire identities or pinned vectors.

## Gate contract

All blocking mechanisms below run from the repository root, fail closed on malformed configuration,
and have a hostile fixture in `scripts/tests/`. `make hooks-test` runs every shell self-test. Bare
`make check` includes every blocking repository-wide mechanism but remains independent of the
Proofbound binary and ledger.

| ID | Mechanism | Required behavior | Proving test |
|---|---|---|---|
| T1-INV-1 | commit cadence | Dirty work fails when HEAD age exceeds the configured minute window; clean work passes; malformed thresholds fail | commit-cadence.test.sh::all-cases |
| T1-INV-2 | cleanroom | Configured literal patterns are searched in tracked content; absent/unreadable pattern input reports INERT without claiming clean; matches fail | cleanroom-lint.test.sh::all-cases |
| T1-INV-3 | lesson recurrence | A `LESSON: class=<id>` occurring twice requires one existing compiled-response path or an explicit RETIRED record | lesson-recurrence.test.sh::all-cases |
| T1-INV-4 | figure provenance | Numeric claims in the bounded `notes/state.md` Blockers section require a date, command, sample/trial count, named round, or explicit not-reproducible marker | figure-provenance.test.sh::all-cases |
| T1-INV-5 | skip declarations | Every emitted Go test or subtest skip exactly matches a nonempty allowlist entry with a reason; empty/test-free logs fail | skip-lint.test.sh::all-cases |
| T1-INV-6 | SPEC prescriptions | Every backtick-quoted sequence of at least two CLI flags in a SPEC appears as Go string literals in shipped non-test code below that SPEC, unless marked `[retracted]` | prescription-lint.test.sh::all-cases |
| T1-INV-7 | state freshness | More than the configured number of commits since `notes/state.md` last changed fails; an edit to that note alone satisfies the in-progress exemption | state-freshness.test.sh::all-cases |
| T1-INV-8 | citation resolution | Every `file_test.go::TestName` citation in a current package SPEC resolves to that exact file and function; stale citations fail | invariant-lint.test.sh::all-cases |
| T1-INV-9 | Claude hooks | secret/force-push/generated-write guards fail closed on missing `jq`; generated edits block; Markdown feedback runs link lint; stop feedback reports dirty/stale state | claude-hooks.test.sh::all-cases |
| T1-INV-10 | generated laws lock | regeneration is deterministic, nonempty, contiguous, and detects inserted, deleted, reordered, or retitled laws | law-citation-lint.test.sh::all-cases |
| T1-INV-11 | short loop | `make short` runs shell self-tests and `go test ./... -short`; bare `make check` contains no short-test bypass | make-contract.test.sh::short-and-full |
| T1-INV-12 | operational helpers | `make meta-tax`, `make backup`, `make state`, and `make wrap-verify` invoke tested scripts; backup destination is explicit/validated and never used by `make check` | operational-tools.test.sh::all-cases |

The exact invariant resolver is intentionally strict. Its first production run is allowed to stay
red while stale citations are updated to real current test functions. A SPEC invariant with no
current proving test is not renamed into green: implementation/test evidence must be restored or the
invariant must be explicitly retired while preserving its permanent ID.

## Documentation and decision reconciliation

- `README.md` states P5 accepted and P6 active, and documents every public operator Make target.
- Every `gates/*.yaml` path is named exactly in `docs/gates.md`.
- Every gate-registry row names an existing mechanism, a resolving self-test, its actual Make
  boundary, and a valid tier/owner/expiry combination.
- The P1 `make vera` text is recorded as an unexercised optional example, not a promised target.
- The local-only backup decision is superseded to reflect the configured Git remote while retaining
  explicit bundle generation as a manual fallback.
- P5's one-phase compatibility allowance is executed at P6 start: live `vera` CLI and `VERA_*`
  aliases are removed only after hostile tests prove Proofbound surfaces remain. Frozen
  `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and historical artifacts
  remain byte-exact.

## Task 1 acceptance artifact

`docs/verification/p6-c1-c2-closure.md` is generated from a checker, not accepted by existence
alone. It lists every C1/C2 row, its probe command, exit status, and exact evidence. Task 1 closes
only when:

```text
scripts/tests/*.test.sh all pass
scripts/p6-census.sh --check exits 0 after regeneration
all C1 and C2 rows are closed
zero expired advisories remain
bare make check exits 0
```

The artifact checker rejects a missing census row, missing mechanism/self-test, stale Make-DAG
claim, unresolved gate path, or evidence bound only to prose.

## Non-goals

- No intent provider, event kind, product report, gate semantic, user UI, or P7+ capability.
- No automatic network push or backup during `make check`.
- No broad cleanroom pattern corpus committed to the repository.
- No claim that a shell linter proves semantic truth beyond its stated lexical boundary.
