# Spec: tests that can fail (undo each change and see if a test notices)

**Status:** draft 2, 2026-11-23, approved by the owner ("APPROVED", both §5 defaults); amended in §7 where the build met the code. **Implements:** `CI-tests-can-fail-2ag60t` → `BR-tests-can-fail-k2h06j` O-1. The owner picked it ("do the first one"):
item 5 of the agent-first list. **Gives evidence toward:** NIST SSDF PW.8.2 ("perform the testing,
and document the results"), produced as PO.3.3 describes ("tools generate artifacts of their support
of secure software development practices").

## 0. The problem, in one line

An agent's test can pass whatever the code does. The demo's reviewer caught that by breaking the code
by hand, and Proofbound could only say "the reviewer did it". After this change, Proofbound shows
whether a test notices, for each piece of the change.

## 1. The idea: undo each hunk

For a commit that claims a spec, take each hunk of **code** it changed (the same hunks C10 lists: not
specs, docs, records or Markdown, and not test files, §2). Then, one hunk at a time:
1. put that hunk back to how it was in the parent;
2. run the project's tests;
3. read the result:
   - **CAUGHT**: the tests fail;
   - **NOT CAUGHT**: the tests still pass. No test depends on that part of the change.

There is also one whole-change run: undo every code hunk at once. If the tests still pass, no test
checks this change at all.

**Why undoing hunks.**
- It works for any language: no parser and no mutation operators, only git and your test command.
- It asks exactly the question a reviewer asks: "if this part of the change weren't there, would
  anything notice?"
- On the demo's first attempt, undoing `src/export.sh` leaves `PASS: export works`, so the result is
  NOT CAUGHT. On the redo, undoing the header line makes the test fail, so it is CAUGHT.

**What this is not.** It is not full mutation testing. A test can catch "the hunk was removed" and
still miss a subtle bug inside the hunk. The page and the record say so.

## 2. Declaring how to test (opt-in)

One optional file, `.proofbound/tests.json` (§7 A1):

```json
{"schema":"proofbound.tests.v1","run":"sh tests/export_test.sh","paths":["tests/"],"timeout_seconds":300}
```

- **`run`** is the test command. Proofbound runs it with `sh -c` at the repository root, in a
  temporary git worktree, so the reviewer's checkout is untouched. The exit code decides: 0 is pass,
  anything else is fail. A timeout counts as fail, and the result says so.
- **`paths`** are the test files, which are never undone. The patterns are written as in `scope.paths`.
- **Without the field, nothing runs:** every result is NOT RUN, and the feature costs nothing.
- **The file is read from the commit's first parent,** as `scope.paths` is (§7 A2), so an agent
  cannot weaken its own test command or test paths in the commit being judged. Its own new tests are still run, because they are in the commit.
- **First, a baseline run** at the commit itself, with nothing undone. If it fails, the result is
  BASELINE FAILS and nothing else is judged.
- **Limits:** at most 20 hunks per commit; the rest are counted as NOT RUN, and the whole-change run
  still happens.

**Warning on the page and in `review draft`:** this runs the commit's own code. Run it only where you
would run that commit's tests anyway.

## 3. Where it shows

- **`review draft <commit>`** runs it when the profile declares tests, and prints one line per hunk:
  `tests caught undoing src/export.sh lines 3-3` or `tests did NOT catch undoing src/export.sh lines
  1-2`. It writes the result into the draft as `"tests": {...}`.
- **`review seal`:**
  - It re-runs the checks and does not trust the draft (the C10 F1 rule).
  - Every NOT CAUGHT hunk needs a decision, as for out-of-area lines: `accept` with a reason (for
    example "logging only") or `reject`.
  - Any `reject`, a whole-change NOT CAUGHT, or BASELINE FAILS makes the verdict NEEDS_WORK.
  - It writes `docs/verification/verdicts/TESTS-<commit12>.json`, signed like the area record:
    command, base revision, each hunk and its result, decisions, reviewer, key and time.
- **Audit room, control C11:** "Each claimed change's tests fail when its code is undone."
  - **Held** when every claimed commit has a record, every hunk is CAUGHT or accepted with a signed
    reason, and the whole-change run was caught.
  - **Exception** otherwise.
  - Commits in a project that declares no tests are counted ("not declared"), not flagged.
  - Maps to SOC 2 CC8.1, ISO A.8.32, SSDF PW.8 and PO.3, and EU AI Act Art. 17(1)(d).
- **Warranty:** not cited yet (§7 A3).

## 4. Non-goals

- No language-aware mutation operators, no coverage tool, no new dependency.
- Nothing runs in CI. The check only reads the signed record. Running arbitrary test commands in CI,
  for every commit, was rejected for cost and safety.
- No automatic test discovery: the project declares the command.

## 5. Decisions for the owner (defaults marked)

1. **Is C11 a blocker or a disclosure?**
   - **Default: a blocker per commit only when the project declares tests.** A NOT CAUGHT hunk with no
     signed decision makes the seal refuse, as C10 does.
   - Alternative: a disclosure only, never blocking.
2. **Where it runs.**
   - **Default: at `review draft` and again at `review seal`, on the reviewer's machine.**
   - Alternative: only a separate `proofbound review tests <commit>`, so drafting stays fast.
3. **The demo gets an act** where `review draft` prints "tests did NOT catch undoing src/export.sh".
   The page's line "Proofbound does not detect a test that cannot fail" changes to say what it now
   does detect, and what it still does not (§1).

**Size:** about 2–3 days, with one non-author review of the code.

## 6. Contract rows

- **CLI-INV-66:** finding and running the hunks.
- **CLI-INV-67:** seal.
- **CLI-INV-68:** C11.

## 7. Amendments made at build (owner-visible)

- **A1. A file of its own, not a profile field.** The sdd profile parser rejects unknown fields, so a
  profile field would have meant changing the spec connector for every project. The file is
  housekeeping (it needs no claim), and C6 lists every commit that changes it.
- **A2. Read from the commit's parent.** At `review draft` there is no range base. The parent is the
  rule `scope.paths` already follows, and C11 adds the F5 rule of C10: a declaration last changed by
  an agent commit inside the range is an exception.
- **A3. The warranty does not cite the TESTS record yet,** nor the AREA record. Citing them changes
  the warranty schema and its verification, so this is `DEF-WARRANTY-CITES-AREA-TESTS-001`.
- **A4. ISO A.8.29 dropped from C11:** it is *security* testing (framework-map review F2).
- **A5. The Review bench in `ui` does not show the tests result.** A seal from the bench in a project
  that declares tests is refused until the hunks are decided in `draft.json`. `review draft` remains
  the way to run the tests.

- **A6. Review fixes changed the runner (§8):**
  - Every run gets a fresh copy of the commit's files with **no `.git`**, not a git worktree.
  - The tests run with nothing undone both **first and last**. If the last run fails, the result is
    **unreliable** and counts as NEEDS_WORK.
  - Binary files, empty files and mode changes are undone as one whole-file unit.
  - **Every hunk that was not caught** (not caught, not run, undo failed) needs a signed decision.
  - A timeout kills the tests' whole process group.
  - The Review bench refuses to save or seal a commit whose project declares tests.

## 8. Review

`tests-can-fail-review-20261123`: NEEDS_WORK (F1 HIGH, F2..F5 MED, all fixed with regression tests;
F6..F9 LOW, recorded per Law 11).

**F1 (HIGH), fixed.** The judged commit's own test could report every hunk as "caught" without
exercising the code, in two ways:
- by asking git whether the worktree was dirty;
- by failing once a marker left by its first run existed.

Now each run gets a fresh copy with no `.git` (git's search stops above it), and the plain run is
repeated at the end: a test that only fails after it has run once reads **unreliable**. C11's Basis
says that a test which only compares file contents still notices every hunk, so the reviewer reads
the tests.

**F2 (MED), fixed.** Hunks that were not run (past the limit, binary, an empty file, a mode change,
or an undo that failed) were recorded with no decision and blocked C11 for ever. Now:
- whole-file changes are undone with their binary-safe patch;
- an undo that fails is its own result;
- every hunk that was not caught needs accept-with-reason or reject;
- C11 accepts an accepted one.

**F3 (MED), fixed.** The bench rebuilt `draft.json` from the ledger, which dropped the test
decisions, and would have run the tests inside an HTTP request. It now refuses to save or seal such a
commit and points to `review draft` and `review seal`.

**F4 (MED), fixed.** A timeout killed only `sh`. The tests now run in their own process group, and
the whole group is killed; on Windows, only the shell is killed.

**F5 (MED), fixed.** Each surviving mutant now has a test:
- the whole change not caught;
- C11 independence;
- the GIT_ variables removed;
- the draft lines;
- the hunk limit;
- the record's commit;
- C6 listing the declaration;
- `draft.json` carrying the result, end to end in the sdd package;
- accept without a reason.

**LOW findings, recorded:**
- **F6:** `PROOFBOUND_*` is still passed to the tests.
- **F7:** `review draft` runs the tests before refusing an existing `draft.json`.
- **F8:** §3 named a "base revision" field that the record does not have; a cancelled caller
  context reads as caught; the labels differ from C10's.
- **F9:** C11's EU AI Act wording.
