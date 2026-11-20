---
schema: vera.verdict.v1
verdict_id: r2-polish-review-20261119
status: NEEDS_WORK
reviewed_commit: 8cc56c73977c6140984a03938a5d3da8d9f51e0f
findings:
  - finding_id: r2-polish-review-20261119-F1
    severity: MED
  - finding_id: r2-polish-review-20261119-F2
    severity: LOW
  - finding_id: r2-polish-review-20261119-F3
    severity: LOW
  - finding_id: r2-polish-review-20261119-F4
    severity: LOW
artifact_path: docs/verification/verdicts/r2-polish-review-20261119.md
artifact_sha: 6196a584e2a66c331f66b09510b63b46fe0e91b2996494f2626aae6bec68569e
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet) that did not write the change, one
round under the Law 11 code budget. It reviewed the **staged diff** on top of `reviewed_commit` (the
`CI-r2-polish-wkudnk` change, before it was committed): `git diff --cached` of cli.go, starter.go,
their tests, the sdd e2e tests and the R2 docs. Read-only; it ran no tests (the author's four hand
mutants were taken as covering the unit tests). Recorded verbatim below, findings re-flowed.

## Status
NEEDS_WORK (one MED).

## Findings
**F1 (MED), kernel/internal/cli/cli.go (the new sdd step in `sync all`).** A structural failure of the
sdd step aborts the whole `sync all` before git, checks, sessions and reviews run. Those steps used to
be independent of spec-sync health, and that property is now broken. sdd `cli_test.go` shows `sync
intent sdd` erroring with "incomplete" on a shallow clone, so a shallow CI checkout that commits a
profile now makes `sync all` fail outright and ingest nothing. That also breaks
`scripts/delivery-enforce.sh`, which runs `sync all`. The same holds for an unreadable profile, or a
profile path that is a tree or a symlink (`hasSDDProfile` only tests that the rev resolves, never that
it is a blob). For `sync all` the failure should be joined after the other steps run, as the reviews
error is, or the SPEC should state that it is intentionally fatal. CLI-INV-23 and the amended
SDD-INV-56 say neither. This repo has no profile, so Proofbound's own gate is not affected.

**F2 (LOW), sdd m3_e2e_test.go TestM3InterleavingAThenRefusal.** It no longer exercises `sync all` for
three things: claim-skip, hold and refused lines printed before the reviews error returns; the
non-zero exit; and the combined summary line. No other test covers these end to end
(cli/m3_print_test.go unit-tests printHolds only; cli/sync_integration_test.go runs the happy path).

**F3 (LOW), kernel/internal/cli/cli.go `hasSDDProfile`.** It goes through `gitOutput`, which swallows
every git error and returns "", so an odd git failure silently skips spec sync (fail-open). The unit
test cannot tell "absent" from "git failed". It also has no tree or symlink case.

**F4 (LOW), kernel/internal/cli/starter.go seal path.** A missing `draft.json` path falls through to
the original read error, which names `<path>/draft.json`. This is cosmetic.

## What was checked
- `verify`, `sync intent all` and delivery-enforce on a repo without a profile are unchanged:
  `intentProviders` still builds sdd only for selection "sdd", and the syntax-tree pin on the literal
  "all" in verify is untouched.
- The summary count correctly adds sdd's appended count, and the output order is coherent.
- `rev-parse HEAD:<profile>` is the correct form for a blob.
- The seal guard accepts only a regular file named exactly `draft.json`.
- The sampled DEMO.md lines match the new behaviour.

## Author's disposition (not part of the verdict)
- **F1 fixed** in the same commit. The spec step's error is printed as `sdd error="..."`, the
  remaining steps still run, and the error is joined into the result (exit non-zero). CLI-INV-23 states
  this, and sdd `TestSyncAllSurvivesAFailingSpecSync` proves the printed, joined error on a shallow
  clone.
- **F2–F4 recorded, not fixed** (Law 11).
