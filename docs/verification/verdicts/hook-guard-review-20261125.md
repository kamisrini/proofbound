---
schema: vera.verdict.v1
verdict_id: hook-guard-review-20261125
status: ACCEPTABLE
reviewed_commit: a07981008aaf7761101782148dc60718d5176b96
findings:
  - finding_id: hook-guard-review-20261125-F1
    severity: MED
  - finding_id: hook-guard-review-20261125-F2
    severity: MED
  - finding_id: hook-guard-review-20261125-F3
    severity: LOW
  - finding_id: hook-guard-review-20261125-F4
    severity: LOW
  - finding_id: hook-guard-review-20261125-F5
    severity: LOW
  - finding_id: hook-guard-review-20261125-F6
    severity: LOW
  - finding_id: hook-guard-review-20261125-F7
    severity: LOW
  - finding_id: hook-guard-review-20261125-F8
    severity: LOW
artifact_path: docs/verification/verdicts/hook-guard-review-20261125.md
artifact_sha: 6be7c249e81ec6af11fdafd1d8ff8bff2c01b8d238e74e0a479e970b68cca828
---
## Provenance
Non-author review (Law 9), the single code round (Law 11), of the uncommitted diff on top of HEAD: kernel/internal/cli/perimeter.go (guardRefusal, called first in perimeterDecision), perimeter_test.go and the CLI-INV-69 amendment, against docs/plans/hook-guard-spec.md draft 2 and BR-write-perimeter O-4. The reviewer ran `go test -run TestPerimeter` and tried the built binary by hand on a scratch repository. Every case §1 and O-4 name held: four tools, three files, case variants, linked `.claude`, a link to settings.json, the session subfolder, a relative path, a corrupt task file, no task; an unreadable task file fails closed. The author's mutation check, calibration stated first: positive control (the guard call removed) died; neutral control (a comment edit) survived; after the fixes, thirteen further mutants died (drop each guarded file, case-sensitive compare, given-only and resolved-only match, guard after the housekeeping return, no cwd base, no project-directory base, no same-file check, no named-root base, unreadable task allowed, no hand-followed link).

## Status
ACCEPTABLE: no HIGH. F1 and F2 fixed before commit; F6 and F7 fixed (cheap, and F6 found a real gap); F3, F4, F5 and F8 recorded (Law 11, spec §2).

## Findings
- **F1 (MED), fixed.** The guard trusted the hook's cwd, which follows a later `cd`, while Claude Code loads settings from where the session started. `CLAUDE_PROJECT_DIR` and the root of its repository are now bases (perimeter_test.go, the project directory after a cd).
- **F2 (MED), fixed.** Native Windows has other names for one file (a trailing dot or space, `::$DATA`, a short name). A same-file check against each existing guard file, and Win32 name trimming for one not created yet (windowsName; a hard link stands in for the alias in the test).
- **F3 (LOW), recorded.** `..` is cleaned as text before an existing symlink is resolved; the same-file check covers an existing guard file, not one not created yet.
- **F4 (LOW), recorded.** Two to four extra `git rev-parse` processes on every edit.
- **F5 (LOW), recorded.** When git itself fails (missing, or a repository it refuses), no root is found and the guard is off.
- **F6 (LOW), fixed.** Tests that could not fail: added a linked `.claude` seen from a cwd outside the repository, a linked `.claude` inside it, and a task file that is a directory. Writing them found that a link to a guard file not created yet was not followed (resolvePath falls back to the link's own path); the guard now follows such a chain by hand. The area check has the same blind spot for code files; recorded in the spec (§7), C10 judges the commit.
- **F7 (LOW), fixed.** The contract row named a base (the root the path names) that §1.2 and O-4 did not; both now name it.
- **F8 (LOW), recorded.** An edit of a nested fixture repository's `.claude/settings.json` is refused though Claude Code never loads it; it only refuses more.
