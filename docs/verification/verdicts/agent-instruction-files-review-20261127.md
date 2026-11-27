---
schema: vera.verdict.v1
verdict_id: agent-instruction-files-review-20261127
status: NEEDS_WORK
reviewed_commit: 59e8be1579fa5f26936b13cbe002491e822c7d7d
findings:
  - finding_id: agent-instruction-files-review-20261127-F1
    severity: MED
  - finding_id: agent-instruction-files-review-20261127-F2
    severity: MED
  - finding_id: agent-instruction-files-review-20261127-F3
    severity: MED
  - finding_id: agent-instruction-files-review-20261127-F4
    severity: MED
  - finding_id: agent-instruction-files-review-20261127-F5
    severity: LOW
  - finding_id: agent-instruction-files-review-20261127-F6
    severity: LOW
  - finding_id: agent-instruction-files-review-20261127-F7
    severity: LOW
  - finding_id: agent-instruction-files-review-20261127-F8
    severity: LOW
artifact_path: docs/verification/verdicts/agent-instruction-files-review-20261127.md
artifact_sha: 76de24b53395ceb4207ec30bbd91f7aa5e9aefd11fb556e288de0dd25f778705
---
## Provenance
Non-author subagent review (Law 9), the single code round (Law 11), of the uncommitted working-tree change on top of HEAD 59e8be1 for `CI-agent-instruction-files-q8v3n1` (kernel/internal/cli: SPEC.md CLI-INV-33/69 amendments and CLI-INV-83/84, check.go, task.go, perimeter.go, check_test.go, perimeter_test.go), read against `BR-write-perimeter-i1xmq8` 0007 O-5 (diffed against 0006 O-4) and the change intent's scope and exclusions. Other callers of `housekeepingPath` (area.go C10, tests.go C11, provenance.go), `runHook` (prepare-commit-msg and commit-msg), starter.go (`init`) and audit.go `trustPaths` (C6) were read. Mutations and probes ran on a copy of `kernel/` in the session scratchpad; no repository file other than this verdict was modified; `make check`, the sdd package and anything starting embedded Postgres were not run; only `go test ./internal/cli -run 'TestHousekeepingPath|TestAgentInstructionFilesAreCode|TestPerimeter'` (and throwaway `TestProbeAIF*` probes in the copy, since deleted) ran. Git behaviour of the no-HEAD `stagedDiff` query and of `ls-tree <base> -- .proofbound/sdd-profile.json` was checked by hand in a scratch repository: both list the profile as intended.

Calibration first. Positive control (`agentInstructionPath` always returns false): killed (TestHousekeepingPath). Neutral controls (rename `baseProfile`; rename `lower` in `trustConfigPath`): both survived. Two first-attempt mutants (M1, M11) failed to compile ("declared and not used") and were re-made to compile before counting.

| # | Mutation | Result |
|---|---|---|
| M0 | positive control: agentInstructionPath returns false | killed |
| N1 | neutral: rename baseProfile | survived |
| N2 | neutral: rename local in trustConfigPath | survived |
| M1 | drop the `!profileAtBase()` guard | killed |
| M2 | stagedCode setup query always false | killed |
| M3 | drop `gates/` from trustConfigPath | killed |
| M4 | drop `.github/workflows/proofbound.yml` from trustConfigPath | killed |
| M5 | trust refusal needs `currentTask` to succeed (not fail closed) | killed |
| M6 | setupCommit without `--diff-filter=A` (a modified profile counts as setup) | **survived** |
| M7 | withoutSetupFiles drops every path, not only instruction files | killed |
| M8 | agentInstructionPath case-sensitive | killed |
| M9 | drop the `.claude/settings.json` exclusion | killed |
| M10 | trustConfigPath case-sensitive | killed |
| M11 | stagedDiff drops the pathspec (any staged added file opens the setup exception) | **survived** |
| M12 | no-HEAD stagedDiff puts the empty tree after `--` (initial setup commit refused) | **survived** |
| M13 | setupCommit without `--root` | survived (equivalent unless a root commit is in the range) |
| M14 | trust check placed after the housekeeping return | killed |
| M15 | trust refusal only when Stat succeeds (fail open on a stat error) | survived (equivalent on Linux test fixtures) |
| M16 | drop the setup exception in check | killed |

17 mutants: 12 killed, 5 survived (2 equivalent in practice, 3 real test gaps: F4).

## Status
NEEDS_WORK: four MED, four LOW, no HIGH. The core is right and well guarded: instruction files are classified before the housekeeping prefixes, case-insensitively, `.claude/settings.json` stays housekeeping, the trust refusal runs before housekeeping and the area and fails closed on whatever the task file holds, `stagedDiff` builds correct arguments in the HEAD and no-HEAD cases and keeps `GIT_INDEX_FILE`, a merge is never a setup commit, and a range whose base has the profile cannot reopen the exception (M1 killed). The defects are a consumer the spec does not account for (C11), a folder gap against the change's own "whatever its folder" claim, user-facing text that now contradicts the rule, and test gaps on the setup query.

## Findings
- **F1 (MED).** tests.go:133 (`testsUnits`), CLI-INV-83/68. C11 also calls `housekeepingPath`, so instruction files are now "code" to the declared-tests control, and CLI-INV-83's list of consumers (check, both commit hooks, perimeter, C10, provenance) leaves it out. Verified by probe: a claimed commit changing only `CLAUDE.md` yields one C11 unit (`CLAUDE.md line 1`). No test can ever catch an undone instruction-file hunk, so in a repository with `.proofbound/tests.json` every claimed `CLAUDE.md`/`AGENTS.md`/`.claude/**` hunk becomes "not caught" and a C11 exception unless the reviewer signs a reason for each one, and `review draft` reruns the suite once per such hunk. Fix: in `testsUnits` skip `agentInstructionPath` paths (C10 already shows them to the reviewer), or deliberately keep them and say so in CLI-INV-83 and CLI-INV-68; either way add a test.
- **F2 (MED).** check.go:322 (`agentInstructionPath`). `.claude/`, `.cursor/` and `.github/instructions/` match only at the repository root, so `packages/api/.claude/commands/x.md` and `packages/api/.claude/skills/s/SKILL.md` are still housekeeping (verified). Claude Code reads `.claude/` from the directory a session starts in (and nested skills in a monorepo), and Cursor supports nested `.cursor/rules`; CLI-INV-83's own justification ("what steers an agent is code, whatever its folder") holds for `CLAUDE.md` but not for these. An unclaimed commit can rewrite a subpackage's agent commands with no claim, and a mid-task agent can edit them outside its area. Fix: match a `.claude/` or `.cursor/` path component anywhere (`lower == ".claude/..." || strings.Contains("/"+lower, "/.claude/")`), keeping only the root `.claude/settings.json` excluded; add the nested cases to TestAgentInstructionFilesAreCode; amend the CLI-INV-83 wording.
- **F3 (MED).** task.go:259; docs/releases/r3/RELEASE-NOTES.md:177; DEMO.md:280; DEMO-MANUAL.md:136. The commit-msg refusal now prints "this commit changes code (AGENTS.md) but claims no spec ... Markdown need no claim", contradicting itself, and the r3 release notes (v0.3.2, being republished with this change rolled in) still promise that commits touching only "Markdown ... still need no claim", with no word that a pull request changing `CLAUDE.md`, `AGENTS.md` or `.claude/` now reads NOT WARRANTED. An adopter upgrading gets a newly red check with no notice. Fix: reword the hook text ("Specs, docs, gates, review records and Markdown other than agent instructions (CLAUDE.md, AGENTS.md, .claude/) need no claim"), update the two demo transcripts, and add a behaviour-change bullet to the release notes.
- **F4 (MED).** check_test.go TestAgentInstructionFilesAreCode. The setup exception's central condition is under-tested: M11 (the staged query loses its pathspec, so staging ANY new file, e.g. `docs/x.md`, exempts every instruction file in the commit) survives; M12 (the no-HEAD argument order, i.e. `init`'s commit in a fresh repository refused by the hook) survives; M6 (check treats a commit that merely modifies the profile as the setup commit) survives. Fix: add staged cases "new housekeeping file + changed AGENTS.md, no profile" (AGENTS.md listed) and "fresh repository, first commit stages profile + AGENTS.md" (nothing listed); add a check case where, in a range whose base lacks the profile, a later commit modifies the profile and AGENTS.md (listed).
- **F5 (LOW).** check.go:249/335, task.go:147. The exception reopens wherever the profile is absent from the immediate comparison point, not only at real setup. Verified by probe: (a) in a range whose base has no profile, commits init -> `git rm` profile -> re-add profile + `CLAUDE.md` "ignore the area": the third commit is not listed; (b) in the commit hook, a housekeeping commit deleting the profile followed by one re-adding it with an `AGENTS.md` change: `stagedCode` returns nothing. (b) is caught by `check` when the base has the profile, and the hook is bypassable with `--no-verify` anyway, so this is residual, not a hole. Fix (cheap, also closes the "profile deleted by an earlier merged commit" residual risk CLI-INV-83 records): a setup commit is one whose ancestry never held the profile (`git rev-list -1 <sha>^@ -- .proofbound/sdd-profile.json` empty; in the hook, the same against HEAD).
- **F6 (LOW).** check.go:462 (`withoutSetupFiles`). The setup commit may change every agent instruction file, not just the two `init` writes: a setup commit can also add or rewrite `CLAUDE.md`, `.claude/commands/**` or `.claude/settings.local.json` unclaimed. The spec states this breadth, so it is a design note. Fix if wanted: exempt only `AGENTS.md` and `.claude/agents/proofbound-reviewer.md`.
- **F7 (LOW).** perimeter.go:67/106 (`trustConfigPath` on `fileRepo`'s rel). Two paths around CLI-INV-84 that `guardRefusal` already handles for the guard files: a dangling symlink into `.proofbound/` (verified: with task active, `src/a/w.json -> ../../.proofbound/waivers/new.json`, a Write is allowed as in-area while an existing-target link is refused), and on Windows a trailing dot/space or `:$DATA` suffix (`.github/workflows/proofbound.yml.`) misses the exact compare. The hook is guidance (the agent has Bash), so LOW. Fix: reuse the guard's hand-followed link chain and `windowsName` before `trustConfigPath`.
- **F8 (LOW).** check.go:309 (`agentInstructionPath`). Other agents' instruction files remain housekeeping (verified): `AGENT.md`, `CONVENTIONS.md`, `.github/prompts/*.prompt.md`, `.github/agents/*.md`, `.github/chatmodes/*.md`, `.clinerules/*.md`, `.windsurf/rules/*.md`, `.gemini/styleguide.md`, `.kiro/steering/*.md`. The owner chose a targeted fix, so record, not fix: a deferral row naming the list would keep the gap visible.
