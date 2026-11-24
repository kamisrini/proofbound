# Spec: the write-time perimeter (refuse an agent's edit outside its spec's area)

**Status:** draft 1, 2026-11-23. The owner said "LETS DO IT" to option 5 of
`DEF-AGENT-NATIVE-INTENT-001`. Reviewed with the code by one non-author (§5). **Implements:**
`CI-write-perimeter-rt7y7g` → `BR-write-perimeter-i1xmq8` O-1. **Contract row:** CLI-INV-69.

## 0. Why

C10 finds the lines an agent changed outside its spec's area, but only after the fact: in review,
once the work is done. At agent speed, the cheaper moment is when the agent tries to write. Stopped
there, the agent hears why and can stay inside the area, or stop and ask.

## 1. The hook

`proofbound init` writes `.claude/settings.json`, but only when none exists. An existing file is never
edited: init prints how to add the hook instead. The file holds a Claude Code `PreToolUse` hook for
`Edit|Write|MultiEdit|NotebookEdit` that runs `proofbound hook pre-tool-use`.

The command reads Claude Code's JSON from stdin (`cwd`, `tool_name`, `tool_input.file_path` or
`notebook_path`), then:

1. **Refuses,** by exiting 0 with Claude Code's PreToolUse JSON (`permissionDecision: "deny"` and
   the reason), when all of these hold (§5 F1):
   - a task is active (`proofbound task start <change>`);
   - that change's spec has a `scope.paths` at **HEAD** (committed, so the agent cannot widen its
     own area in the working tree);
   - the file, after symlinks are resolved, is inside the repository that holds it (found from the
     file's folder, not from the session's cwd: §5 F2), is not housekeeping (CLI-INV-33), and matches
     no pattern in the area.

   The installed command is `proofbound hook pre-tool-use || true`, so a missing, old or broken
   binary never blocks an edit. `.claude/settings.json` is housekeeping, so it needs no claim, and C6
   lists every change to it.
2. **Also refuses** every code edit when that `scope.paths` is invalid (fail closed, as in C10).
3. **Allows** everything it cannot judge: no repository, no task, no area declared, a path outside
   the repository, a tool that is not an edit, or input it cannot read.

**The reason names:**
- the file;
- the task;
- the area file and its patterns;
- what to do: do not edit the file for this task; if the change truly needs it, ask the spec owner
  to widen `scope.paths` and commit that first; otherwise the reviewer decides every line outside the
  area (C10).

## 2. What it is not

- **Not a trust boundary.** The agent controls its own environment: it could edit settings, remove
  the hook, or write files through a shell command. That is why C10 still judges every committed line.
  This hook is early feedback, in line with the principle recorded in the deferral.
- **Not a shell guard.** The `Bash` tool can write files, and is not inspected.
- **Claude Code only.** Other harnesses (Cursor, Copilot, Codex) are not covered.

## 3. Non-goals

- No change to C10 or to any check.
- No editing of an existing `.claude/settings.json`.
- No policy for edits without a task. The commit hook already asks for a task.

## 4. Tests

- **`TestPerimeter`.** Allowed: inside the area, a relative path, a test path in the area,
  housekeeping, a tool that is not an edit, outside the repository, malformed input. Refused: Write,
  MultiEdit and NotebookEdit outside the area, a `..` path back out, a symlink inside the area that
  points outside it. Also checked:
  - an uncommitted wider area does not count;
  - no task, or a spec without an area, is allowed;
  - an invalid area refuses;
  - through `run`, the command exits 2 with the reason, or 0 silently.
- **`TestInitWritesPerimeterHook`.** init writes the file when absent, and an existing file is left
  untouched.

**Acceptance:**
- In the agent demo, the agent's pricing edit is refused by the hook, given the JSON Claude Code sends.
- **Not done here: a live Claude Code session.** The hook format is Claude Code's documented
  PreToolUse contract: JSON on stdout with `permissionDecision: "deny"` blocks the call and gives the
  agent the reason. The owner tries it in a real session.

## 5. Review

`write-perimeter-review-20261124`: NEEDS_WORK (F1 HIGH, F2..F4 MED, all fixed; F5..F8 LOW recorded
per Law 11; F6 fixed with F2, since the same function was rewritten).

**F1 (HIGH), fixed.** Exit 2 blocks in Claude Code, but it is also an older binary's usage error and a
Go panic, so a shared `.claude/settings.json` with an older `proofbound` would have blocked every
edit. The hook now refuses with the documented PreToolUse JSON on stdout and exits 0, and the
installed command ends in `|| true`. A test runs that command against a stub `proofbound` that exits
2, and the edit is allowed.

**F2 (MED), fixed.** The repository was found from the session's cwd, so a session started above the
repository, or after a `cd`, allowed everything. It is now found from the file's nearest existing
folder.

**F3 (MED), fixed.** `.claude/settings.json` needed a claim to commit, which the acceptance run also
hit. It is now housekeeping, C6 lists it, and there is a test.

**F4 (MED), fixed.** The command test compared against the code's own constant. It now checks the
exact JSON and the exact installed command.

**LOW findings:**
- **F5, recorded:** a relative path is tested only in the allow direction, and the empty-cwd guard is
  untested.
- **F6, fixed with F2:** a new folder behind a symlink.
- **F7, recorded:** false refusals for gitignored files and for case variants on case-insensitive
  file systems.
- **F8, recorded:** the agents block does not say that `scope.paths` must be committed before the
  hook takes effect.
