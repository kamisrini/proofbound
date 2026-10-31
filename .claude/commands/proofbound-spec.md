---
description: Write a kernel package SPEC.md before implementation (Law 6)
---

Write the contract for a kernel package BEFORE its implementation (Law 6: spec before code). Tell me
the package path if you have not already (e.g. `kernel/internal/<pkg>`). Tests derive from this spec;
`kernel/internal/specfirst` enforces that every internal package has a `SPEC.md` and a proof, so a
package without one cannot pass `make check`.

Study an existing contract first — `kernel/internal/gates/SPEC.md` or `kernel/internal/store/SPEC.md`
— and match its shape and voice. The spec has four jobs (CLAUDE.md Law 6):

1. **Interface** — the exported surface and the data/format it operates on, concretely. Show the
   shapes (envelopes, payloads, CLI forms) the way `gates/SPEC.md` shows its gate-definition JSON.
2. **Invariants** — a numbered table of the properties that must always hold, each in the form a
   test can check. Every row carries a permanent `INV-<N>` id and names its proving test as a
   machine-readable `file_test.go::TestName` token (or a retired `—`). `scripts/invariant-table-lint.sh`
   requires exactly that token shape in each proving cell.
3. **Non-goals** — what this package deliberately does NOT do, so a later reader does not read a gap
   as a bug.
4. **Test derivations** — how the tests fall out of the invariants. State every "cannot / never /
   always" as a test, because in this repo an untested absolute is a defect
   (`VD-verification-asymmetry-2dyjnd`).

Numbering is permanent (Law numbers' sibling rule): `INV-<N>` ids APPEND only — never insert,
reorder, reuse, or silently retitle. `scripts/spec-numbering-lint.sh` checks this against the
`@generated` `docs/invariants.lock`; regenerate the lock with `make invariants-lock` ONLY on a
deliberate append or retitle, and the lock diff is the review.

Watch `scripts/prescription-lint.sh`: a backtick-quoted sequence of TWO OR MORE CLI flags in a SPEC
must have each flag present as a Go string literal in shipped (non-test) code at or below this spec's
directory — so do not prescribe an invocation the code does not make. Mark a deliberately
not-yet-implemented sequence `[retracted]` rather than letting the gate guess intent.

Then confirm: `scripts/invariant-table-lint.sh`, `scripts/spec-numbering-lint.sh`, and (if you
appended invariants) `make invariants-lock` followed by a review of its diff. Run `make check` when
the package is ready for the `specfirst` proof to bind.

**Governed-path note:** `kernel/**/SPEC.md` is explicitly NOT a governed path
(`scripts/intent-applicability.sh`), so a spec-only commit is not governed. But `docs/invariants.lock`
IS governed and `@generated` — if you regenerate it, that commit expects an owner-approved change
intent.

End by printing the spec path, its `INV-*` ids, and whether the invariants lock changed.
