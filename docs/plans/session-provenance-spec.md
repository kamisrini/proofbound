# Spec: session-to-line provenance (which agent session wrote which committed lines)

**Status:** draft 2, 2026-11-23, **approved by the owner 2026-11-24** ("approved", the §6 defaults).
Built; amended in §9 where the code review met the design. Item 3 of the
agent-first list; the owner said "lets do next". **Would implement:** a new BR/CI pair (written at
build, like the earlier features). **Contract rows:** CLI-INV-70..72 (§7).

## 0. The problem, in one line

Today the only thing that says "an agent wrote this" is a `Co-Authored-By: Claude` line, which the
agent types itself. Nothing ties a committed line to the session that wrote it. Nothing shows the
lines that reached the commit some other way either, for example through a shell command that the
write-time perimeter never sees.

## 1. The idea: record each edit as it happens, match lines at review

### 1.1 Record

`proofbound init` adds a Claude Code `PostToolUse` hook beside the perimeter hook:
- The hook's matcher is `Edit|Write|MultiEdit`. NotebookEdit is left out (§8 F8), because a
  committed `.ipynb` holds JSON-encoded lines, and those never equal the raw source.
- The hook runs `proofbound hook post-tool-use || true`.
- PostToolUse fires only for a tool call that succeeded.

For each call, the hook appends records to `$(git rev-parse --git-common-dir)/proofbound/sessions/<session_id>.jsonl`:
- **Which repository:** the one that holds the file, found from the file as the perimeter does.
- **Why that folder (§8 F6):** it is untracked, shared by every worktree of the repository, and needs
  no `.gitignore` line, so an agent working in a worktree is still seen from the main checkout.

Each record holds:
- the session id;
- the time, from the hook's own clock (Claude Code's input has no time field, §8 F7);
- the tool;
- the repository-relative file;
- the active task, if any;
- **the sha256 of each line the tool added** (§8 F3). The added lines are found in this order:
  1. the `+` lines of `tool_response.structuredPatch`, when Claude Code sends one;
  2. otherwise, for Edit and MultiEdit, the lines of `new_string` minus the lines of `old_string`,
     counted as a multiset, so repeated context lines are not credited;
  3. otherwise, for Write, every line of `content`. Such a record is marked `whole_file`, and its
     matches are reported as the weaker "via a whole-file write".

Each line is normalised before hashing (§8 F11): a trailing `\r` and trailing whitespace are removed.

Records store hashes only, never file content and never a transcript. A large edit is split across
several records, each written in one `write` call of at most 3 KB, so parallel subagents in the same
session cannot interleave their records (§8 F14).

### 1.2 Match

At `review draft <commit>`, Proofbound takes every line the commit **adds** to a path that is not
housekeeping. Test files are included; docs, Markdown and records are not traced, and the draft says
so (§8 F12). Each line, normalised the same way, is looked up in that repository's session logs for
the same file. There is no time cut-off (§8 F5): the commit's date can be set by whoever commits, and
amending or rebasing keeps the old date. Each added line is then one of:

- **by session S.** One or more sessions recorded writing it; all of them are listed, with "via a
  whole-file write" where that applies.
- **no recorded agent edit** (§8 F1). No log records this line. The page says only what is known:
  that no recorded edit wrote it. It does not guess who did.
- **common line, not counted** (§8 F4). Either the line has three or fewer non-space characters, or
  the same text occurs more than once among the lines the commit adds to that file (`return err`,
  `} else {`). Such lines would match almost anything, so they are counted apart and never used as
  evidence.

**Warnings** in the draft and in the record:
- No session logs for this repository on this machine: "no session logs here, so nothing can be
  traced (they are local to the machine the agent ran on)".
- The commit names an agent (author or `Co-Authored-By`), but no line is by any session: "the
  edit-recording hook was probably not running for this commit".
- `.claude/settings.json` has no `post-tool-use` hook: "edits are not being recorded" (§8 F2).

`review draft` prints one summary line per file, for example
`provenance src/export.sh: 6 lines by Claude Code session 3f2a91c0, 0 with no recorded agent edit, 2 common`
or `provenance src/pricing.sh: 0 by a session, 1 with no recorded agent edit (line 1)`.

Lines moved with a shell `mv`, or rewritten by a format-on-save step, correctly show as "no recorded
agent edit". The page says so, so a reviewer does not read them as a person's work (§8 F11, F15).

### 1.3 Seal

`review seal` re-runs the match and never trusts the draft. It then writes
`docs/verification/verdicts/PROV-<commit12>.json`, signed by the reviewer like the AREA and TESTS
records. The record holds:
- the commit;
- for each file, its counts per session;
- the line numbers with no recorded agent edit;
- the warnings;
- the sha256 of each session log, **as read at seal**;
- the reviewer, the key and the time.

The log hashes are the reviewer's statement of what they read. A third party can re-check them only
while those local logs still exist (§8 F13).

The Review bench in `ui` computes the same record when it seals, since matching runs nothing and is
cheap (§8 F9). Commits sealed before this feature have no PROV record and stay "not traced": seal
refuses to write over existing artifacts, and this spec adds no command to add a record afterwards.

### 1.4 Audit room, control C12

C12 asks: "Each claimed change says which agent sessions wrote its lines." (§8 F10)

- **Not applicable:** no claimed commit in the range has a PROV record.
- **Held:** every PROV record in the range verifies against a declared key, and was signed by a
  reviewer who is neither the commit's author nor one of its agents (the same independence rule as
  C10).
- **Exception:** a PROV record fails verification, or was signed by the commit's own author or agent.
- **Never an exception for the counts.** People and generators legitimately write lines (§6
  decision 2). Commits without a record are counted as "not traced" and listed in the evidence.
- **The evidence lists,** for each commit: its sessions and their line counts, the lines with no
  recorded agent edit, and the warnings.
- **Other edits:** the `controlQuestions` comment changes from "ten questions" to the actual count,
  and C6's Basis wording names both hooks in `.claude/settings.json`.

## 1.5 Existing repositories (§8 F2)

`init` never edits an existing `.claude/settings.json`, including the one v0.3.2's `init` wrote with
only the perimeter hook. It now checks for each hook separately. For any hook that is missing, it
prints the exact JSON to add, where today it prints plain "exists". The draft warning in §1.2 says
the same thing at review time. The agent demo's repository is created fresh, so it gets both hooks.

## 2. What this makes visible

- **The demo's first attempt.** The agent changes `src/pricing.sh` with `sed -i` from the shell,
  around the perimeter. The draft says `src/pricing.sh: 1 with no recorded agent edit`. The edit
  the perimeter could not see now shows up by name.
- **A human fix-up on an agent branch.** It shows as lines with no recorded agent edit, so the
  reviewer knows which lines no recorded session produced.
- **The agent track record page (the next item)** reads the PROV records: per session, which
  commits, how many lines, and how they fared in review (C10 rejects, C11 not caught).

## 3. What it is not

- **Not proof against a hostile agent.** The agent runs on the same machine, so it could write to the
  session log or skip the hook. The record is the reviewer's signed statement of what the logs on
  their machine showed. It turns "the agent says it wrote this" into "these lines match these
  recorded edits, and these lines match none".
- **Not line authorship in the `git blame` sense.** A line counts as by a session if that session
  recorded adding the same text to the same file. An identical line typed by a person later still
  matches, and so does any line from a whole-file write.
- **Same machine only.** The logs are local. A reviewer on another machine gets the "no session logs
  here" warning, not a claim that a person wrote everything (§6 decision 1 is the alternative).
- **Claude Code only,** like the perimeter, and Edit/Write/MultiEdit only (not notebooks).

## 4. Non-goals

- No storing of file content, prompts or transcripts.
- No change to C10, C11 or the warranty schema. The warranty does not cite PROV yet, in the same
  way as AREA and TESTS (`DEF-WARRANTY-CITES-AREA-TESTS-001`).
- No `Agent-Session:` commit trailer: a trailer is self-declared, which is the problem this solves.
- No command to add PROV records for commits that are already sealed.

## 5. Tests (planned)

- **`TestSessionLog`:**
  - Edit, Write and MultiEdit each log only the lines they added: through `structuredPatch`, through
    the `new_string` minus `old_string` fallback (context lines not credited), and through a
    whole-file Write.
  - A tool that does not edit, or a path outside the repository, writes nothing.
  - A failed edit (a defensive case; Claude Code does not send it) writes nothing.
  - Bad input never fails the hook.
  - A worktree's log lands in the common git folder.
  - A large edit is split into records of at most 3 KB.
- **`TestProvenanceMatch`:**
  - by one session, by two sessions, no recorded agent edit, a common line (short or repeated);
  - a person's line followed by an agent's whole-file Write is reported "via a whole-file write";
  - the same text in another file does not count;
  - an amend after more edits still matches (no time cut-off);
  - CRLF and a missing final newline;
  - each of the three warnings.
- **`TestSealWritesProv`, `TestAuditC12`:**
  - the record is signed and verifies;
  - not applicable, held, and not traced;
  - a tampered record, or one signed by the commit's agent, is an exception;
  - the bench seal writes PROV.
- **`TestInitHooks`:** a settings file with only the perimeter hook makes init print the
  post-tool-use JSON. A settings file with both hooks prints "exists".
- **End to end in the sdd package:**
  - a Write and an Edit through the hook, then a shell-written line, then a commit;
  - draft, seal and audit;
  - with an older binary on PATH, the hook is skipped (`|| true`), and the draft gives the
    hook-not-running warning.
- **Acceptance:** the agent demo shows the pricing line as "no recorded agent edit".

## 6. Decisions for the owner (defaults marked)

1. **Where the evidence lives.**
   - **Default: local logs in the git folder, signed into a PROV record at seal.**
     - It is cheap.
     - It is private, because the logs hold hashes only.
     - It works for the common case, where the person supervising the agent reviews on the same
       machine.
   - Alternative: also commit the session logs, so that a reviewer on any machine can match lines.
     The cost is a file per session in the history, and the agent can still edit a log before it is
     committed.
2. **Is C12 a blocker or a disclosure?**
   - **Default: a disclosure.** Lines with no recorded agent edit are shown but never block. People
     fix things by hand, generators write files, and the matching has known false matches (common
     lines, whole-file writes).
   - Alternative: an exception when an agent commit has, inside its own area, lines with no recorded
     agent edit (likely shell writes). Each would need an accept-with-reason or a reject, as in C10.
3. **The demo** gets the shell edit described in §2, so the page shows the perimeter's blind spot by
   name.

**Size:** about 2 days, with one non-author review of the code (Law 11).

## 7. Contract rows (planned)

- **CLI-INV-70:** the `post-tool-use` hook, the session log, and init's per-hook check.
- **CLI-INV-71:** matching the added lines at draft, at seal and in the bench; the PROV record.
- **CLI-INV-72:** control C12.

## 8. Review

Round 1, one non-author, on draft 1: **NEEDS_WORK** (F1 and F2 HIGH, F3 to F10 MED, F11 to F15 LOW).
Every HIGH and MED finding is folded into draft 2 at the section marked. F11, F12, F14 and F15 were
cheap enough to fold in too. F13 is reflected in the wording of §1.3. The reviewer's facts about
Claude Code's hooks came from its own knowledge, so the build checks them against the documented
contract before coding.

- **F1 (HIGH):** "not from an agent edit" claimed more than is known, given a missing hook, an older
  binary or another worktree. It is renamed "no recorded agent edit", and the hook-not-running
  warning is added.
- **F2 (HIGH):** `init` could not add the hook to any v0.3.2 repository, and it said "exists" without
  a hint. Fixed with the per-hook check and the draft warning (§1.5).
- **F3 (MED):** context lines and whole-file writes were credited to the agent. Only added lines are
  hashed now, and whole-file writes are marked.
- **F4 (MED):** the three-character threshold was too low. Lines that repeat in the commit are now
  "common" as well.
- **F5 (MED):** the time cut-off rested on a date anyone can set, and broke on amend. It is dropped.
- **F6 (MED):** logs were per worktree. They now live in the git common folder.
- **F7 (MED):** Claude Code's input has no time field, and PostToolUse fires only on success. The hook
  now uses its own clock, and the failure test is marked defensive.
- **F8 (MED):** notebooks can never match. NotebookEdit is left out.
- **F9 (MED):** the bench sealed without PROV. It now writes PROV, and the lack of a record for
  commits sealed earlier is stated.
- **F10 (MED):** C12 was self-contradictory. It now has defined states, signature and independence
  rules, and no opt-in.

## 9. Code review and amendments (owner-visible)

`session-provenance-review-20261124`: NEEDS_WORK (F1 HIGH, F2..F7 MED, all fixed with regression
tests; F8, F9, F11 and F12 LOW fixed because they were one line each; F10 LOW recorded).

- **F1 (HIGH), fixed.** Seal wrote an unsigned PROV record when the reviewer had no declared key, and
  C12 then made every keyless seal an exception. **Amendment:** PROV is written only when the
  reviewer signs. Without a key the commit is "not traced".
- **F2 (MED), fixed.** Paths with a space (git adds a tab after them) or non-ASCII characters (git
  quotes them) were mis-read: the lines were signed as "no recorded agent edit", or the file was left
  out.
- **F3 (MED), fixed.** An added line starting with `++ ` was taken for a file header, and the rest of
  the file was lost. Hunks are now read to the length their header gives.
- **F4 (MED), fixed.** **Amendment to §1.4:** C12 is **listed**, never held, so a page with one traced
  commit and nine untraced ones no longer reads "held". The question now reads "Traced changes say
  which agent sessions wrote their lines, in records that verify." A record that does not verify,
  that judges another commit, or that the agent signed itself is still an exception.
- **F5 (MED), fixed.** An unreadable log or log folder now gives a warning, not a failed draft or
  seal. Very long diff lines no longer fail.
- **F6 (MED), fixed.** Tests were added for a record of another commit and for a path too long for
  one record.
- **F7 (MED), fixed.** DEMO-MANUAL.md and DEMO.md now show the PROV line that seal prints.
- **LOW, fixed:**
  - F8: the settings warning says "in this checkout".
  - F9: the hook never writes through a link planted at the log path.
  - F11: the Basis says a merge commit is compared with its first parent.
  - F12: a session id cannot start with a dot.
- **LOW, recorded:** F10. Input over 8 MB, and a path too long for one record, are skipped silently.
