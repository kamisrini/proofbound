---
schema: vera.verdict.v1
verdict_id: mvp-release-v032-review-20261122
status: ACCEPTABLE
reviewed_commit: 1ec78505dd62c25b48bbf06b10746653a9d1d76a
findings:
  - finding_id: mvp-release-v032-review-20261122-F1
    severity: LOW
  - finding_id: mvp-release-v032-review-20261122-F2
    severity: LOW
  - finding_id: mvp-release-v032-review-20261122-F3
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-release-v032-review-20261122.md
artifact_sha: 5161764e79919c6fb8d3441d2db3abf60f467d4dd31af7066501097713374288
---
## Provenance
Separate Sonnet reviewer agent, one round, on the uncommitted change on top of reviewed_commit. Ran `git diff -- install.sh scripts/tests/install.test.sh docs/` and `bash scripts/tests/install.test.sh` (prints `install.test: ok`). Skimmed the doc edits against `housekeepingPath` in kernel/internal/cli/check.go (list matches) and the hook text in kernel/internal/cli/task.go. No make check, no Go tests.

## Status
ACCEPTABLE: no HIGH or MED findings.

## Findings
- **F1 (LOW).** scripts/tests/install.test.sh:60 `! grep -q "warning:" "$tmp/out"` cannot fail: under `set -e` a leading `!` is exempt from errexit, so a spurious warning would pass. Fix: `if grep -q "warning:" "$tmp/out"; then exit 1; fi`.
- **F2 (LOW).** install.sh:75 runs `"$found" version` with no timeout; a shadowing binary that hangs on `version` would hang the installer after the install has already succeeded. Errors are handled (falls back to "unknown version"). Fix: accept, or guard with `timeout 5` when available.
- **F3 (LOW).** RELEASE-NOTES.md v0.3.2 says the claim must be "on the message's last lines"; task.go:209 says it must be in the final trailer block, last line after a blank line. Fix: reword to "in the final trailer block". Also: the warning only fires when `$dest` is spelled identically on PATH (a trailing slash or `~` spelling takes the "not on PATH" note branch instead, never a false warning); POSIX portability is fine (`command -v`, `case`, `$(...)` only).
