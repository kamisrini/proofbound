---
schema: vera.verdict.v1
verdict_id: write-perimeter-review-20261124
status: NEEDS_WORK
reviewed_commit: 4a12f5be1f5df8087c28bee076dca73bd7058272
findings:
  - finding_id: write-perimeter-review-20261124-F1
    severity: HIGH
  - finding_id: write-perimeter-review-20261124-F2
    severity: MED
  - finding_id: write-perimeter-review-20261124-F3
    severity: MED
  - finding_id: write-perimeter-review-20261124-F4
    severity: MED
  - finding_id: write-perimeter-review-20261124-F5
    severity: LOW
  - finding_id: write-perimeter-review-20261124-F6
    severity: LOW
  - finding_id: write-perimeter-review-20261124-F7
    severity: LOW
  - finding_id: write-perimeter-review-20261124-F8
    severity: LOW
artifact_path: docs/verification/verdicts/write-perimeter-review-20261124.md
artifact_sha: ab97bc9128aa0440562c866b59b93ee20c7bb37f898bbf2bb4ff08b3d7559db9
---
## Provenance
Non-author subagent review (Law 9), the single code round (Law 11), of the staged diff on top of HEAD 4a12f5b, exported with git checkout-index to the session scratchpad (staged/); mutations on a copy (review-wp/). No repository file was modified; `make check` was not run; port 55432 was not used; only `go test ./internal/cli -run 'TestPerimeter|TestInitWritesPerimeterHook|TestProbe'` ran, one at a time. Old-binary behaviour was checked by building HEAD's kernel (pb-old) and piping hook JSON into `pb-old hook pre-tool-use`: usage, exit 2. The Claude Code hook contract was checked from knowledge of the hooks reference (https://docs.claude.com/en/docs/claude-code/hooks), not fetched this session: the settings shape hooks.PreToolUse[].matcher plus hooks[].{type:"command",command} is right; the input has session_id, transcript_path, cwd, hook_event_name, tool_name, tool_input (file_path for Edit/Write/MultiEdit, notebook_path for NotebookEdit); for PreToolUse exit 2 blocks and sends stderr to the model, any other non-zero exit is non-blocking; alternatively exit 0 with hookSpecificOutput.permissionDecision "deny" plus permissionDecisionReason. The four tools cover the built-in file writers; Bash and MCP writers are not covered, as spec §2 says. Unstaged working-tree edits (audit.go, check.go, perimeter.go, perimeter_test.go, agent-demo.sh) were not the target and are mentioned only for F3. Calibration first: positive control (perimeterDecision always allows) killed; neutral (comment whitespace) survived. 18 mutants: 11 killed, 7 survived; 3 equivalent on Linux (directory-only symlink resolution, unresolved root, dropping whole-path EvalSymlinks while the directory is resolved); 4 real gaps: perimeterExit = 1, dropping the relative-path join, dropping the empty-cwd guard, and the new-directory symlink case found by a probe.

## Status
NEEDS_WORK: one HIGH, three MED. The core design is sound: the area is read at HEAD, an invalid area fails closed, symlinks are resolved for existing paths, housekeeping is exempt, and the refusal names the file, task, area and next step. The spec is honest that this is not a trust boundary; one fair addition: specs/ is housekeeping, so an agent can widen and commit its own scope.paths without a claim (C6/C10 catch that later).

## Findings
- **F1 (HIGH).** perimeter.go claudeSettings and runPerimeterHook; cli.go:226-231. Claude Code reads exit 2 as block, but the CLI also exits 2 for a usage error and Go exits 2 on a panic. Verified: with a committed, shared .claude/settings.json, a collaborator, CI agent or PATH-shadowed install (DEF-INSTALL-PATH-SHADOW) running an older proofbound has every Edit/Write/MultiEdit/NotebookEdit blocked and the agent shown usage text; the same on any panic or a future rename. In practice "the agent cannot edit files", pushing it to write through Bash, which nothing inspects. Fix: decide via stdout JSON (hookSpecificOutput permissionDecision "deny", exit 0); configure `proofbound hook pre-tool-use || true`; test that a non-decision failure does not block.
- **F2 (MED).** perimeter.go perimeterDecision, rev-parse --show-toplevel from call.Cwd. The repository comes from the session cwd, not the target. Verified: a session started from a parent directory (~/projects) or after a Bash cd (the Bash cwd persists and is reported as cwd) allows every out-of-area edit silently; the reverse, cwd in repo A and the file in nested repo B, judges the file against A's area though C10 never sees it. Fix: root from the target's nearest existing ancestor; cwd only to resolve a relative path.
- **F3 (MED).** check.go housekeepingPath (staged). init now writes .claude/settings.json but the staged change does not make it housekeeping: committing init's output is refused as unclaimed code and check lists it UNCOVERED, contradicting CLI-INV-33 ("the files init writes"). The unstaged tree already fixes it (claudeSettingsPath housekeeping, listed by C6 via trustPaths, with a test). Fix: stage it with this change; CLI-INV-33/69 should name .claude/settings.json.
- **F4 (MED).** perimeter_test.go "through the command" compares with the constant perimeterExit, not 2: `const perimeterExit = 1` survives, and exit 1 does not block in Claude Code, so the central contract can break with every test green. Fix: assert the literal and unmarshal the settings JSON's exact hook entry.
- **F5 (LOW).** perimeter_test.go. The relative path is tested only allowing, so dropping filepath.Join(call.Cwd, target) survives (Rel errors, allowed); the empty-cwd guard is untested and its mutant survives. Fix: a refused relative path outside the area, and an input with no cwd.
- **F6 (LOW).** perimeter.go resolveFile/resolveDir resolve only the immediate parent of a not-yet-existing file. Verified: with src/a/up -> src, writing src/a/up/newdir/x.go is allowed but lands in src/newdir/. Fix: walk to the nearest existing ancestor, EvalSymlinks it, re-append the rest.
- **F7 (LOW).** perimeter.go false refusals (safe direction, friction): gitignored files in the repository are refused though C10 never judges them (verified); case-variant paths (SRC/A/x.go) on case-insensitive file systems are refused (verified); a directory path is refused (harmless). Fix: allow on `git check-ignore -q`; optionally canonicalise case on darwin/windows.
- **F8 (LOW).** HEAD-only area, spec §1: when the agent writes spec, scope.paths and code in one uncommitted tree there is no area at HEAD and everything is allowed silently; by design and honest in the spec, but the agents block does not say scope.paths must be committed first. Fix: one sentence there, optionally a verbose-mode stderr note when a working-tree scope.paths exists but HEAD has none.
