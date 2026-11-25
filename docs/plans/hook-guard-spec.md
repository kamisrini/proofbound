# Spec: an agent cannot switch off its own guard

**Status:** draft 3, 2026-11-25, built. The owner said "go" to draft 1. Draft 1 was reviewed once by a
non-author (§6: NEEDS_WORK, 1 HIGH, 5 MED, 2 LOW, all folded in). The O-4 wording changed (§4) and
needs the owner's approval.
**Implements:** a new change intent → `BR-write-perimeter-i1xmq8`, new obligation O-4 (it retires O-3).
**Contract row:** CLI-INV-69 (amended).

## 0. The gap, in one line

While a task is active, the write-time perimeter lets an agent Edit or Write `.claude/settings.json`,
the file that holds the perimeter hook itself.

**Why it is a gap:**
- `.claude/settings.json` is on the housekeeping list (CLI-INV-33), and the perimeter allows every
  housekeeping path. So an agent refused at `src/pricing.sh` may remove the hook, or add
  `"disableAllHooks": true`, and then try again.
- `.claude/settings.local.json` is not housekeeping, but it is usually outside the area, so the
  perimeter refuses it only by accident. It is also usually gitignored, so no commit, C6 or reviewer
  ever sees it. The same `disableAllHooks` line there switches off every project hook.
- What already holds: the Audit room's C6 lists every commit that changes `.claude/settings.json`, and
  C10 still judges every committed line. The gap is the moment of writing, which is the perimeter's
  whole purpose.

## 1. Behaviour after the fix

1. **The guarded files.** Three paths:
   - `.claude/settings.json` and `.claude/settings.local.json`: the hooks, and `disableAllHooks`;
   - `.proofbound/task`: the task itself. An agent that writes
     another change's name (one with no area) or garbage into it switches the perimeter off as surely
     (§6 F1).
   An edit's path **matches** a guarded file when, either as given or after symlinks are resolved
   (a link to a file not created yet included), it names the file without regard to letter case
   (§6 F3, F4), or when it is the same existing file (catching hard links and the other names Windows
   has for one file: a trailing dot or space, a `:$DATA` stream, a short name; §7 F2).
2. **Where.** A guarded path counts at these bases: the root of the repository that holds the edited
   file, the root of a repository the path as given names (a linked `.claude`), the session's working
   directory (`cwd`), Claude Code's project directory (`CLAUDE_PROJECT_DIR`: settings load from where
   the session started, while `cwd` follows a later `cd`; §7 F1), and the roots of the repositories
   holding those two folders (§6 F2). The edited file need not be in a repository.
3. **When.** A task counts as active for the guard while `.proofbound/task` **exists**, whatever it
   holds, at the root of any of those repositories. A corrupt or unreadable task file therefore keeps the
   guard on (fails closed). The
   area check is unchanged: a corrupt task file still allows code edits there, as today.
4. **The order.** The guard runs right after the tool is known to be an Edit, Write, MultiEdit or
   NotebookEdit with a path, **before** the not-in-a-repository, housekeeping, no-task and no-area
   returns (§6 F5).
5. **The refusal** reads: `proofbound: <file> holds the guard on this agent's work (the hooks, or the
   active task <claim>). An agent may not change it while a task is active. Ask the person running the
   task to change it.` The claim is the task file's first line, or `(unreadable)` when it cannot be
   read or is empty.
6. **With no task active,** all three files stay editable, so setup works as today.
7. **Nothing else changes.**
   - `.claude/settings.json` stays housekeeping for `check` (CLI-INV-33) and `hook commit-msg`.
   - C6 already lists commits that change it, so the Audit room needs no change.
   - A refused edit never reaches PostToolUse, so provenance is unchanged.

## 2. What it does not close (stated, not fixed)

These go on the live demo page and in the docs, so no one reads the perimeter as a lock:
- **Shell writes.** The perimeter sees only Claude Code's file tools. An agent allowed to run
  `sed -i`, `echo >` or `rm` through Bash can still change any guarded file. C6 sees the committed
  settings file; nothing sees the local one or the task file.
- **The task commands.** An agent allowed to run `proofbound task stop`, or `task start` for a change
  with no area, ends or widens its own task. The live demo allows only `proofbound task show`.
- **User-level settings.** `~/.claude/settings.json` is outside every repository.
- **A session started above the repository.** The hook sees a task only in the edited file's
  repository or the session folder's repository. A session started in an outer folder, whose own
  settings run the hook, may edit those settings while a nested repository's task is active.
- **The binary.** The installed hook ends in `|| true`, so if `proofbound` stops resolving on PATH
  the guard is silently gone. Changing PATH through settings needs a guarded file, and a file tool
  cannot set the executable bit, so the agent cannot plant a fake binary that way (§6 F8).
- **When an edit takes effect.** Claude Code snapshots hooks at session start and asks for review in
  `/hooks` before a changed hook applies; whether `disableAllHooks` is treated the same is unverified
  (§6 F7). Either way, the change would apply in the next session, while the task is still active.
- **Recorded from the code review (§7), not fixed:** a path whose `..` is cleaned as text before an
  existing symlink is resolved, for a guard file that does not exist yet; two to four extra `git`
  processes on every edit; no guard when `git` itself fails (missing, or a repository git refuses);
  an edit of a nested fixture repository's `.claude/settings.json` is refused though Claude Code
  never loads it.
- **The developer.** A developer can always switch the hooks off. They are guard rails for the agent,
  not a control on the person. Trust is decided by the pull-request check, which runs in CI.
- **A real lock** would need Claude Code's managed (organisation) settings: the owner's open question.

## 3. Tests (planned)

- **cli, `perimeterDecision`.** Every refusal asserts the new reason, including the file and the claim
  (§6 F6):
  - a task is active: Edit, Write, MultiEdit and NotebookEdit of each of the three files are refused;
  - each file is refused also when the spec declares no area, and when the area is `**`;
  - `.CLAUDE/Settings.Local.json` is refused;
  - a `.claude` directory that is a symlink, to a folder inside or outside the repository, is refused
    by the path before resolution; a link elsewhere that points at `.claude/settings.json` is refused
    by the path after;
  - the session started in a subfolder `sub/` of the repository with the task: `sub/.claude/settings.json`
    (the file Claude Code loads for that session) is refused; the same path with the session started
    at the root is judged by the area as today;
  - a corrupt task file (`x`) keeps the guard on;
  - with no task file: all three are allowed; `.claude/commands/x.md` is always allowed.
- **Unchanged:** `TestHousekeepingPath` still reads `.claude/settings.json` as housekeeping.
- **Mutation calibration:** positive controls (each must fail a test): drop the local settings file
  from the list; drop the task file; compare case-sensitively; check only the resolved path; move the
  guard after the housekeeping return. Neutral control: a comment edit must survive.

## 4. The requirement rewording (needs the owner's approval)

O-3 says the hook allows every housekeeping path, `.claude/settings.json` and `.proofbound/` included,
so the fix contradicts it. O-3 is retired and O-4 replaces it: O-3's first sentence unchanged, plus
this (the two earlier drafts were each found AMBIGUOUS, §6, §7):

> Despite that, while a file .proofbound/task exists, whatever it contains, at the root of the
> repository that holds the edited file, of a repository whose root the path as given names, or of the
> repository holding the session's working directory or Claude Code's project directory, a Claude Code
> agent's Edit, Write, MultiEdit or NotebookEdit is always refused when its path, either as given or
> after symlinks are resolved, matches .claude/settings.json, .claude/settings.local.json or
> .proofbound/task, compared without regard to letter case, at any of those roots, in the session's
> working directory or in the project directory, or is the same existing file as one of them, with a
> reason naming the file and the first line, trimmed of surrounding white space, of one such task file,
> or (unreadable) when that file cannot be read or that trimmed line is empty.
> Init installs the hook, and C10 judges every commit afterwards.

## 5. Decisions (defaults taken; the owner may change them)

1. **D1: guard both settings files and the task file** (recommended). The local file and the task
   file are the quieter ways, because git never sees them.
2. **D2: guard only while a task is active** (recommended). Outside a task, the perimeter allows
   everything already, and refusing the settings file there would block the normal setup.
3. **D3: no attempt to parse Bash commands.** Guessing which shell command writes which file is
   unreliable, and a guess would read as a lock. The gap is stated (§2) instead.

## 6. Review and amendments (owner-visible)

Round 1, one non-author, on draft 1: **NEEDS_WORK**. Every finding was folded into draft 2:
- **F1 HIGH.** The agent could Write `.proofbound/task` (housekeeping, gitignored): a corrupt or
  other claim ends the perimeter. Amendment: the task file is guarded, and the guard is on while the
  file exists (§1.1, §1.3).
- **F2 MED.** Claude Code loads settings from the session's directory, not the edited file's
  repository. Amendment: three bases (§1.2).
- **F3 MED.** Case variants. Amendment: compared without regard to case (§1.1).
- **F4 MED.** A symlinked `.claude` escaped the resolved-path check. Amendment: both paths (§1.1).
- **F5 MED.** The guard's place was unstated. Amendment: §1.4.
- **F6 MED.** Deny tests could pass with half the check removed. Amendment: reasons asserted, one
  mutant per file (§3).
- **F7, F8 LOW.** When an edit takes effect, and the `|| true` PATH gap: stated in §2.
- **O-4 AMBIGUOUS** (which file, what counts as active, what an edit is): reworded with the
  reviewer's fix, plus the session directory (§4).

## 7. Code review (owner-visible)

`hook-guard-review-20261125`: **ACCEPTABLE** (no HIGH; F1, F2 MED fixed; F3-F8 LOW recorded, two of
them also fixed because they were cheap and found real gaps).
- **F1 MED, fixed.** `cwd` follows a `cd`; Claude Code's project directory is now a base.
- **F2 MED, fixed.** Other Windows names for one file: a same-file check, and the Win32 name
  trimming for files not created yet.
- **F6 LOW, fixed.** Tests that could not fail: a linked `.claude` seen from outside, a linked
  `.claude` inside the repository, an unreadable task file. Writing them found a real gap: a link to
  a guard file not created yet was not followed. The guard now follows it.
- **F3, F4, F5, F8 LOW, recorded** (§2). **F7 LOW, fixed:** §1.2 and O-4 now name the root the path
  names.
- **Found while fixing, recorded:** the area check itself does not follow a link to a file not
  created yet, so a Write through such a link inside the area can create a file outside it. C10
  still judges the commit.
- **O-4 round 2: AMBIGUOUS** (either form or both; what names a corrupt task). Reworded (§4).
- **O-4 round 3** (owner approved a third round): AMBIGUOUS on an empty or blank task file and which
  of several task files the reason names. The reviewer's replacement was taken verbatim and confirmed
  by the same reviewer (§4).
