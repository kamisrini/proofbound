---
schema: vera.verdict.v1
verdict_id: change-area-review-20261123
status: NEEDS_WORK
reviewed_commit: dbe0154e9ac544b61585365f667969f9507ca02d
findings:
  - finding_id: change-area-review-20261123-F1
    severity: MED
  - finding_id: change-area-review-20261123-F2
    severity: MED
  - finding_id: change-area-review-20261123-F3
    severity: MED
  - finding_id: change-area-review-20261123-F4
    severity: MED
  - finding_id: change-area-review-20261123-F5
    severity: MED
  - finding_id: change-area-review-20261123-F6
    severity: LOW
  - finding_id: change-area-review-20261123-F7
    severity: LOW
  - finding_id: change-area-review-20261123-F8
    severity: LOW
  - finding_id: change-area-review-20261123-F9
    severity: LOW
  - finding_id: change-area-review-20261123-F10
    severity: LOW
artifact_path: docs/verification/verdicts/change-area-review-20261123.md
artifact_sha: 1a4fcb365bec8e51143664ad2b53bd13cf357f6603ac4b4961e4d2ca34944c73
---
## Provenance
Non-author review (Build Law 9), the one code round (Law 11), of the staged change (`git diff --cached`) on top of HEAD dbe0154e9ac544b61585365f667969f9507ca02d: the declared change area. Read docs/plans/change-area-spec.md, CLI-INV-52..55, area.go, and the starter.go, audit.go, acts.go, workspaces_http.go, ui.go, review.html and app.js edits, plus the helpers they use (verifyRecord, signRecord, declaredKeys, independence, housekeepingPath, claimLines, reviewSigner, mergeDecisions). `go vet ./internal/cli/` is clean and `go test ./internal/cli/ -run 'Area|C10|TestUISealFlow|TestAuditControls'` passes. Git behaviour behind F1 and F6 was confirmed with throwaway repositories under the session scratchpad. The sdd/e2e packages and `make check` were not run, so the spec's acceptance on the tour's `shop` repository and "`sync all` still accepts the change folder" were not checked.

## Status
NEEDS_WORK: five MED findings, no HIGH. Checked and sound:
- The area is read from the first parent, never from the commit. Paths are listed with `-z` and `--no-renames`, so quoting, spaces, newlines and renames cannot hide a path. The per-path diff uses `--literal-pathspecs`, `--no-ext-diff` and `--no-color`. A hunk header cannot be forged from file content under `-U0`, because content lines start with `+` or `-`.
- Seal rebuilds the list from git and takes only the decisions (matched by path and lines) from the editable draft. An undecided, empty, mistyped or reasonless entry, or a draft without the field, refuses before any file is written. Any `reject` turns every OV into `NEEDS_WORK`. The AREA record goes into the same all-or-nothing batch.
- C10 reads HELD only when: the HEAD record verifies (canonical form, schema fields only, the `by` name bound to the key HEAD declares and to its fingerprint); its `commit` equals the full SHA; every hunk git lists now is accepted with a valid reason; and the signer passes the C8 `independence` rule (author, agents, signer, spec owners). An unsigned record (`key: ""`, no signature) fails verification. A record cannot go stale, because hunks are a pure function of the commit and its parent.
- The bench escapes through html/template, and the JS uses `getAttribute` and `textContent` only. A decision for another range is refused by index plus path plus lines.
- Existing users without `scope.paths` see no change beyond C10 reading NONE: the draft gains `"out_of_area": []`; old drafts decode (the field is optional) and need no decision; `specInParent` keeps the same behaviour through `specDirsAt`.
- Each of the five tests would fail if the rule it names were broken (checked one by one). Coverage gaps are in F9.

## Findings
- **F1 (MED).** kernel/internal/cli/area.go:216 and :255. Both `git diff` calls honour `submodule.<name>.ignore` from `.gitmodules` and the local `diff.ignoreSubmodules`. Scenario: `.gitmodules` sets `ignore = all` for `lib/sub`. A claimed commit then bumps the `lib/sub` gitlink outside the area. `--name-status` prints nothing for it, so C10 reads "inside its area" (HELD) and the draft lists nothing. Confirmed in a scratch repository. The same option also hides the bump from `diff-tree`, so C5's uncoveredCommits has the same blind spot; that one predates this change. Fix: pass `--ignore-submodules=none` to both diff calls (and to the C5 `diff-tree`), and add a test case.
- **F2 (MED).** kernel/internal/cli/area.go:573-583 (`areaSetter`). It runs one `git log -1` across all of `res.Files`, so it finds only the most recent commit that touched any of them. Spec §4 asks whether "the area file of a claimed change" was last changed by an agent. Scenario: an agent commit widens one claim's file, for example `specs/cart/scope.paths` set to `**` next to the owner's narrow `openspec/changes/cart/scope.paths`, or the file of a second claimed change. A later human commit then edits the other file. `log -1` returns the human commit, so the agent's widening is not flagged and the union lets everything in. Fix: run the check once per file and raise the exception when any file's last in-period change was made by an agent. Add a two-file test.
- **F3 (MED).** kernel/internal/cli/area.go:156-159 (`specDirsAt`). Any archive entry that ends in `-<name>` counts as a spec folder of `<name>`, and its `scope.paths` is added to the union. Scenario: change `cart` is claimed, and the archive holds `2026-01-01-add-cart` with `scope.paths` set to `src/`. The `cart` area silently widens to all of `src/`, with no agent action and nothing shown to the reviewer. C7 already had this suffix rule, but there a false match only made a spec look present; here it widens what is allowed. Fix: match the OpenSpec archive form exactly (`^[0-9]{4}-[0-9]{2}-[0-9]{2}-<name>$`). Also consider showing `res.Files` (the area files used) in the draft output, the bench note and the C10 evidence.
- **F4 (MED).** kernel/internal/cli/workspaces_http.go (serveReviewAct decode, `http.MaxBytesReader(..., uiMaxBody)`) with ui.go:72 (`uiMaxBody = 16 << 10`). Each area decision costs about 150-250 bytes of JSON plus its reason, and the list has no total cap (only 50 per file). Scenario: an agent reformats two or three helper files outside the area, giving about 100 entries. The bench's draft, preview and seal POSTs exceed 16 KiB and come back "malformed request". Browser seal can never succeed, which §3 says the bench exists to prevent. Fix: give the review acts a larger limit sized to the list (for example 1 MiB), or cap the total entries per commit with one "and N more" entry, as is done per file. Add a test with more than 16 KiB of decisions.
- **F5 (MED).** kernel/internal/cli/area.go:506-510 (the not-declared branch) and :204-209. The "set by an agent" check runs only when the area is declared. Scenario: an agent deletes `openspec/changes/cart/scope.paths`, either in an unclaimed housekeeping commit (it is under `openspec/`) or inside its own earlier claimed commit. Its next claimed commit then reads "area not declared", which is evidence only. Removing the area is the widest possible widening, yet C10 stays NONE or HELD. Fix: when a claimed change has no `scope.paths` at the parent, run the per-file setter check (F2) on the candidate paths `<dir>/scope.paths` of every spec folder. `git log` with a pathspec also finds the deleting commit. Raise an `area` exception when an agent made that last change. Add a test.
- **F6 (LOW).** kernel/internal/cli/area.go:190. `git show <parent>:<dir>/scope.paths` also succeeds when `scope.paths` is a directory. It prints `tree <sha>:<path>`, a blank line and the entry names, which `parseArea` accepts as patterns. A directory `scope.paths/` holding a file named `**` therefore yields the area "everything" (confirmed in a scratch repository). Only the setter check (F2, F5) stands in the way. Fix: read `<parent>:<file>` only when `git cat-file -t` says `blob`; otherwise treat it as invalid.
- **F7 (LOW).** kernel/internal/cli/area.go:255. Hunk ranges depend on the local git config and version (`diff.algorithm`, `diff.indentHeuristic`, `diff.interHunkContext`). Decisions are keyed by the exact `lines` text, so an AREA record sealed on one machine can stop matching the ranges an audit computes on another (CI, or the as-of worktree under a different git). C10 then reads EXCEPTION. This fails safe but is noisy. Fix: pin `--diff-algorithm=myers --no-indent-heuristic --inter-hunk-context=0` in `pathHunks`.
- **F8 (LOW).** kernel/internal/cli/area.go:282-287. A new empty file has no text hunk and is listed as "mode change", which is wrong. Fix: pass the name-status `A` through and label it "file added (empty)"; keep "mode change" for real mode-only changes.
- **F9 (LOW).** kernel/internal/cli/area_test.go. §6 says TestAuditC10 covers the invalid-file case, and it does not. These are also untested: an unsigned AREA record in C10, a record whose `commit` names another commit, the all-inside HELD path, the 50-per-file cap and its "and N more hunks" entry (CLI-INV-53), the "binary file" and "mode change" labels, and `review draft` writing and printing `out_of_area` (CLI-INV-54, first clause). Fix: add these small cases, or narrow the CLI-INV rows to what is tested.
- **F10 (LOW).** kernel/internal/cli/area.go:66-100. Spec §1 says a line that "contains `..`" is invalid. The code rejects only a `..` segment, so `a..b` is accepted. That is harmless, because the matcher has no `..` meaning, and CLI-INV-52 states the segment rule. Also, a control character at the edge of a line (for example a trailing tab) is removed by `TrimSpace` before the check, so it is accepted. Fix: align the spec §1 wording with CLI-INV-52, and run the control-character check before trimming (or trim only spaces).
