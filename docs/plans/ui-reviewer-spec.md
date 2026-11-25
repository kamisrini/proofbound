# Spec: a reviewer does the whole review in the browser

**Status:** draft 2, 2026-11-25, built (R1 and R2). Owner asked (2026-11-25, while walking
`DEMO-AGENT-LIVE.md`): "reviewers should be able to do their activities in the UI, and see any
opportunities for streamlining", then "lets do both r1 and r2". Draft 1 was reviewed once by a
non-author (§7: NEEDS_WORK, 4 HIGH, 6 MED, 2 LOW, all folded in). The owner's hold, the
agent-reviewer-first flow (D4), is **not** in this spec.
**Implements:** one change intent for both parts (built and reviewed in one session) →
`BR-audit-room-ui-r7t3w1` new obligations O-8 (R1), O-7 (R2.1, R2.2) and O-5 (R2.3), via
`CI-ui-reviewer-b8n4q2`. Requirement reviews: round 2 found O-3 and O-4 AMBIGUOUS (replaced by O-6,
O-7), round 3 found O-6 AMBIGUOUS in one clause (replaced by O-8), round 4 (owner-approved, past the
Law 11 budget) found all VERIFIABLE; revision 0006 is active.
`BR-ui-workspaces-m6q2v8` is not touched: its own requirement review is a dormant deferral
(`DEF-UI-WORKSPACES-REQ-REVIEW-001`). **Contract
rows:** CLI-INV-47 and 67 amended (R1); CLI-INV-75, 76 (R1), 77, 78 (R2) new; W-INV-13 (the verify page's check, in the warranty package); CLI-INV-28, 30
untouched (the browser's refusal deliberately differs, R2.2).

## 0. The gap, in one line

In the live demo a reviewer types about twenty commands for steps 8, 9 and 12. The Review bench
(`/review`) already covers promises and out-of-area lines, but **refuses any commit in a project that
declares tests** (CLI-INV-67), so for an agent-first project, which is every project the demo sets up,
the reviewer cannot use the browser at all.

**Why it refused:** the bench rebuilds the draft from the record on every save, and the record has no
test results, so one Save wiped the tests out of `draft.json` (tests-can-fail review F3). The fix must
carry them through, not only lift the refusal (§7 F1).

**What the bench already does** (no change): queues the commits waiting for review that this reviewer
may review (never their own, nor an agent's commit they are named on); shows each promise, the
requirement wording, the out-of-area lines and the change; saves a draft; previews what sealing
writes; seals with the local key and commits only the written files. The UI syncs the record on
every page load, so the reviewer never runs `sync`.

## 1. R1: the reviewer's whole job in the browser

1. **Tests in the bench.** The 422 refusal goes.
   - **Run.** The page shows **Run the tests** with the CLI's warning: this runs the commit's own code,
     as you, with your files; read the change first. Nothing runs on page load. When
     `.proofbound/tests.json` is invalid at the parent, the page shows why and offers no Run.
   - **The endpoint.** Run is `POST /act/review-tests` and never a GET. Before starting anything it
     applies every check `serveReviewAct` applies: loopback host, cookie and Origin; the commit is a
     claiming commit of this period waiting for review; `reviewerFor` passes. An agent's bearer
     request never starts it.
   - **The job.** One job at a time per server (a second Run or Seal while one runs is refused, saying
     so). It runs under the server's own context, not the request's: closing the tab does not stop it;
     stopping the server cancels it (its process group killed, its temporary checkout removed). It
     holds **no** request lock and no open store while the tests run. Status is
     `GET /review/job?commit=`, read-only: running (with the result lines so far), done, failed.
   - **Progress.** `commitTests` gains a progress callback that reports each run as it ends, in the
     words `review draft` prints (`tests caught undoing …`). This changes shared code; the CLI prints
     the same lines as today, earlier.
   - **The result is kept.** When the run ends, it is written into the local `draft.json` `tests`
     (creating the draft from the record if there is none), keeping any decision that matches a hunk by
     path and lines. A reload shows it without running again.
   - **Decisions.** Each part the tests did not catch (not caught, not run, undo failed) gets
     **Accept** with a one-line reason, or **Reject**, as a new decision part `"tests"` addressed by path
     and lines, checked like `"area"`. `undecided` lists every one. The whole-change result, a failing
     or unreliable baseline are shown, not decided.
   - **Saving keeps them.** Save and Seal take `tests` from the local `draft.json` and replace only each
     hunk's decision and reason from the request. `mergeDecisions` carries `tests` too, so the page
     shows a saved run.
   - **Seal re-runs the tests**, as `review seal` does. In order: run the tests (no lock, no store);
     then take the lock, open the store, rebuild the truth from the record, match the decisions the
     reviewer submitted with the click, and write; refused and nothing written if a part not caught in
     the new run is undecided. `reviewSeal` takes its fresh test run from the caller, so the CLI runs
     its tests before opening the store too: the order changes, the outputs do not. A seal whose tab
     closed still finishes and commits; the next load shows its result.
   - **The tests' environment** drops `PROOFBOUND_*` and `DATABASE_URL` (as it drops `GIT_*` today), in
     the CLI and the bench: the commit's code must not learn the reviewer's key path or the record.
2. **Who wrote each line.** The bench shows the provenance summary per file (by which agent session,
   with no recorded agent edit, common) and its warnings. Read-only; seal matches again, as today.
3. **The flags on the code itself.** In "The change", each flagged added line carries its flags beside
   it: *outside the area*, *tests did not notice* (*not tested* since R3, §10), *no recorded agent edit*. The demo's pricing line
   reads all three. Past the 1,500-line cut, the page lists the flagged lines it could not show.
4. **The spec beside the promises.** Above the promises: the claimed spec's Markdown files and
   `scope.paths`, read from the commit's **parent** (amended after code review F4: that is what the area
   check enforces), with the path they came from, and a mark when the commit itself also changes one.
5. **The queue says what needs a decision.** Each entry: promises, lines outside the area, tests
   (not run yet / N parts not caught / all caught, from the local draft), and the agents named on it.
6. **Who you sign as.** The header names the reviewer and key fingerprint the seal will use, and says
   to compare it with the `.pub` file under `.proofbound/keys/` for that name. When the local key cannot
   review (none, undeclared, the author), it says why and the command to start the UI with a
   reviewer's key: `PROOFBOUND_KEY=<file> proofbound ui`.
7. **Preview lists every file** seal writes: RR, OV, AREA, `TESTS-<c12>.json`, `PROV-<c12>.json`
   (only when signing) and the gate; and warns when a failing or unreliable baseline or a whole change
   not caught will make seal record NEEDS_WORK.
8. **The result after sealing.** The page shows that commit's check result from a fresh sync: PASS
   or BLOCKED with the gate's reason (the line `check` prints), and a link to the Audit room.
9. **Reviewed, needs work.** Below the queue, read-only: commits this period whose current review is
   NEEDS_WORK, with the reviewer and what was rejected.

**R1 does not change** what seal writes, signs or decides; the independence rule; the record. The
bench calls the same draft, tests, area, provenance and seal code the CLI calls.

## 2. R2: the flow on one page, and the warranty

1. **The Change page** (`/change?claim=sdd:<change>`), read-only: one change as the flow the demo
   walks: spec → its commits (with agents and sessions) → reviews (outcome, reviewer) → check →
   warranty. Each step shows its state; the viewer sees the one action that is theirs now ("Review
   this" → the bench; "Issue the warranty" → R2.2), as a link. It reads only what the Audit room and the
   bench already compute.
2. **Issue the warranty.** A button on the Audit room, for a period whose check is fully WARRANTED:
   - writes the bundle into `docs/warranties/<head12>/` (a fresh folder per HEAD, so a second issue at
     a new HEAD never meets CLI-INV-30's non-empty refusal; at the same HEAD it says it was issued);
   - signs as `warranty bundle` does, with the key the UI holds, named on the button ("Issue as
     <name>, key <fp>");
   - shows what was written and commits only that folder, "release warranties at <head12>". The
     committed bundle describes HEAD's parent, the commit it names; the page says so.
   - When anything is not warranted, the browser writes nothing and lists the commits. This differs
     from the CLI, which writes a bundle that carries no warranty (CLI-INV-28); the browser only
     issues a release that is whole.
3. **Verify with no install** (D5). A page published with each Proofbound release, never inside a
   bundle (a bundle must not vouch for itself, §7 F3): `proofbound-verify.html`. The auditor opens it,
   chooses the warranty files, and **types or pastes the trust key fingerprint** they got from the
   vendor by another channel; the page never reads a key from the files. The check is the Go code of
   `pbwarranty` (parse, signature, rule) compiled to WebAssembly and embedded in the page, so there is
   one home for the rule (Law 2) and no Web Crypto dependency. It says plainly what it cannot check
   without the repository (`--repo`) and names `proofbound ui --bundle` for that. Build: the Go
   toolchain's own `GOOS=js GOARCH=wasm` (no new dependency); `release-build.sh` adds the page as a
   release asset.

## 3. What this is not

- Not a hosted, multi-user service. The UI stays a local server for one person and one key.
- Not task start/stop, init, keys or running an agent from the browser.
- Not a way to run anything but the project's declared tests, on a click, as the person who clicked.
- Not a lock on the reviewer: they can still overturn any flag; the flags only say what to look at.
- Not the agent-reviewer-first flow (owner's hold, D4).

## 4. Tests (planned)

R1:
- a project with tests: Run → progress lines → result written into draft.json; reload shows it without
  running again; Save after Run keeps `tests`; a not-caught part needs accept-with-reason or reject
  before preview; seal re-runs and writes `TESTS-<c12>.json`; seal refuses, writing nothing, when the
  second run finds an undecided part.
- the CLI's `review seal` of a bench-saved draft writes the same TESTS record as the bench's seal.
- a GET, an agent's bearer request, and a commit outside the period start nothing.
- one job at a time; a page loads while a seal's tests run; a closed tab's seal still commits.
- the tests do not see `PROOFBOUND_KEY` or `DATABASE_URL`.
- an invalid tests.json shows its reason and no Run.
- flags: the demo's pricing line carries all three; the agent's lines carry none; flagged lines past
  the cut are listed.
- spec text and area come from the commit, not the worktree; queue counts; signer line; preview lists
  TESTS and PROV; the post-seal line equals `check`'s line; the needs-work list shows the step-9
  commit.
- end to end: the live demo's steps 8–9 and 12 through the bench's HTTP calls give the same verdict
  files, gate result and warranty as the CLI steps.

R2:
- Change page: the demo's two attempts appear as one change with two commits, one BLOCKED with its
  reason, one WARRANTED; the action shown matches the viewer (author: none; reviewer: Review this).
- Issue: refused with the list when one commit is not warranted (nothing written); written and
  committed under `docs/warranties/<head12>/` when all are; the same HEAD twice says it was issued;
  the written bundle equals `warranty bundle` at the same base.
- Verify page: the WebAssembly build passes the same vectors as `pbwarranty`'s Go tests (VALID,
  UNREVIEWED, INVALID, a wrong trust key, an altered byte); it never reads a key from the files.

Mutation calibration (positive controls, each must fail a test): the bench seals without re-running
the tests; a not-caught part left undecided is sealed; Save drops `tests`; the pricing line loses its
area flag; Issue writes a partial release; the verify page takes the key from the file. Neutral
control: a comment edit survives.

## 5. Decisions (defaults taken; the owner may change them)

1. **D1: R1 and R2 both before launch** (owner's choice). Built in order R1, R2.1, R2.2, R2.3; each is
   gated and folded into v0.3.2 when done, so if Sunday comes first, what is finished ships and the
   rest follows. The reviewer recommended R2 after launch (§7 F12); the owner chose both.
2. **D2: tests run only on a click, as a background job** (recommended).
3. **D3: the browser's seal re-runs the tests, as the CLI does** (recommended).
4. **D4: the held flow.** The owner's design, recorded so R1 fits it, not built: a reviewer agent
   reviews first and records findings and an assessment; a person then reviews the whole change with
   those in hand, may overturn or add to any of them, and signs the verdict that counts; the record
   keeps both. R1's per-line flags and decisions are where those findings would appear.
5. **D5: the verify page is Proofbound's, not the bundle's, and runs the Go rule as WebAssembly**
   (recommended over draft 1's verify.html inside the bundle, which the reviewer found let a bundle
   vouch for itself and copied the rule into JavaScript).

## 6. Not decided here

- Hosting the verify page on a public site (commercial, separate).

## 7. Review and amendments (owner-visible)

Round 1, one non-author, on draft 1: **NEEDS_WORK**. Every finding was folded into draft 2:
- **F1 HIGH.** Save dropped the tests (the old refusal's reason), and nothing carried them. Amendment:
  the `"tests"` part, Save/Seal keep `tests`, `mergeDecisions` carries it (§1.1).
- **F2 HIGH.** Seal held the lock and an open store through the second test run, so every page
  returned 503 for up to hours. Amendment: tests first, then lock and store; `reviewSeal` reordered for
  the CLI too (§1.1).
- **F3 HIGH.** verify.html inside the bundle let the bundle vouch for itself and copied the rule into
  JavaScript. Amendment: D5, §2.3.
- **F4 HIGH.** R2 had no contract or tests. Amendment: §4 R2, contract rows named in the header.
- **F5 MED.** The Run endpoint was unstated. Amendment: POST under `/act/`, every bench check first,
  no bearer (§1.1).
- **F6 MED.** Job lifecycle. Amendment: server context, cancelled on shutdown, status endpoint, a
  closed tab's seal still commits (§1.1).
- **F7 MED.** The tests saw the reviewer's key path. Amendment: the environment drops `PROOFBOUND_*`
  and `DATABASE_URL`, in the CLI too (§1.1); §3.
- **F8 MED.** Line-by-line progress did not exist. Amendment: a progress callback (§1.1).
- **F9 MED.** Issue: folder, second issue, key, the parent commit, partial bundles. Amendment: §2.2.
- **F10 MED.** Preview missed TESTS and PROV. Amendment: §1.7.
- **F11 LOW.** Invalid tests.json, "reload shows it unrun", flags past the cut, what to compare the
  fingerprint with: folded (§1.1, §1.3, §1.6, §4).
- **F12 LOW (scope).** R2 after launch recommended; the owner chose both (D1).

## 8. Code review of R1 (owner-visible)

`ui-reviewer-r1-review-20261125`: **NEEDS_WORK** (no HIGH, five MED), all MED fixed before commit:
- **F1 MED, fixed.** Run was offered for an invalid tests file. The page now shows the reason and no
  Run; the endpoint refuses (`TestBenchInvalidTestsFile`).
- **F2 MED, fixed.** The tests and seal endpoints did not check the commit was waiting for review: a
  reviewed commit's code could be run with a click. Every bench action now requires NO REVIEW.
- **F3 MED, fixed.** The flags depended on the reviewer's git settings (no prefix, colour, renames).
  `git show` now runs with fixed options; a flagged path the patch does not show is listed.
- **F4 MED, fixed.** The spec text came from the parent while §1.4 said the commit. §1.4 amended (the
  parent is what the area check enforces), and a commit that changes its own spec is marked.
- **F5 MED, fixed.** Missing tests (reload, page loads during a job, reviewed commit, invalid tests
  file, the real seal refusing a part only the second run finds, flags past the cut) added; a test
  that could not fail (a decision on a caught part) now isolates its one reason.
- **F7, F8, F9 LOW, fixed** (cheap): the tests job's draft merge takes the lock and writes atomically;
  a seal that has started finishes even if the server stops; buttons refused by the server stay off.
- **F6, F10, F11 LOW, recorded.** The CLI now runs the tests before it checks `declared_reviewer`
  (wasteful, not wrong); the job's status is readable with the agent's read token (summary lines
  only); the queue computes the area per commit on every load, and the needs-work list shows the gate
  line rather than the reviewer's name.

Mutation calibration (stated before the count): positive controls, each died: the seal skips the
second test run; an undecided part not caught is sealed; Save drops the tests; the pricing line loses
its area flag; a reviewed commit can run its tests; the tests see `PROOFBOUND_KEY`. Neutral: a comment
edit survived. R2: Issue writes a partial release, Issue is offered to a reviewer, the verify page
takes the key from the file: each died; a comment edit survived.

Tried for real on a copy of the live demo, through the bench's own HTTP calls: the step-9 commit
(agent work plus the hand-edited price) ran its tests in the background, showed all three flags on the
price line, was rejected and sealed: BLOCKED, NEEDS_WORK, 2 of 2 SATISFIED, exactly as `check` reads.
The step-12 commit sealed PASS. Issued as the spec owner (the bench refused the reviewer's key: a
reviewer may not issue), committed `docs/warranties/<head12>/`; `warranty verify --repo` from a fresh
clone read VALID, and the verify page's check read SIGNATURE ONLY with that key, INVALID with another.

## 9. Code review of R2 (owner-visible)

`ui-reviewer-r2-review-20261125`: **NEEDS_WORK** (one HIGH, three MED), all fixed before commit:
- **F1 HIGH, fixed.** Issue checked only the bundle's entries, which cover claiming commits, so a
  period with code that claims no spec (NOT WARRANTED by `check`) could still be issued. Issue now
  requires the period's check to read WARRANTED: every C1 line PASS and no uncovered-code exception.
- **F2 MED, fixed.** The Changes page offered Issue from one change's last commit; Issue covers the
  whole period. It now offers Issue only when the viewer may issue the period's warranty (BR O-7),
  says what holds the release back otherwise, and offers Review this for any waiting commit.
- **F3 MED, fixed.** "Issued" could be spoofed by an uncommitted index or a commit subject. It is now
  read from HEAD's tree, a release commit must add exactly its folder, and the dirty check runs first.
- **F4 MED, partly fixed.** Added: the author sees no action; the release build must produce and list
  the verify page. Recorded: Issue's bundle is compared with `warranty bundle` only through the shared
  `bundleCompute` (a fixture needs the record store), and the WebAssembly glue and page script are not
  run by any test here (no node or browser on this machine): the owner opens the page once by hand.
- **F5, F7 LOW, fixed:** HEAD moving during Issue is refused; the page's SIGNATURE ONLY wording says the
  review account is the file's own; the issue message says which commit it warrants; an unreadable
  file shows a line; the loader check ignores case. **F6 LOW, recorded:** the WebAssembly links the
  database driver through the connectors (9 MB, 3 MB compressed).

Mutation calibration for these fixes: Issue ignoring uncovered code, Review this only for the last
commit, a subject alone counting as a release: each died; a comment edit survived.

Found by the gate after commit `b553d2d`: the verify page's check lived in the command, where the
package-spec lint cannot cite its test. It moved into the warranty package as `PageCheck` (W-INV-13,
`page_test.go::TestPageCheck`), a pure addition that calls the existing `Parse` and `Verify`; the
rule is unchanged. CLI-INV-79 is retired. The key-from-file mutation, re-run there, died.

## 10. R3: the code beside the questions, in plain words (owner "yes", 2026-11-25)

Found by the owner walking the bench: the code sat far below the questions ("where is the code to
check on this screen?"), and the questions assumed a developer ("what is this section supposed to
be?"). R3 changes only the page: what the reviewer decides, what is sealed and what `check` reads are
untouched.

1. **Side by side.** With a commit open on a wide screen (≥ 1280 px), the bench shows the decisions in
   one column and the evidence in another, in this order: the change (flagged lines marked), who wrote
   each line, the spec. The evidence column stays in view while the decisions scroll (sticky, its
   own scroll). The queue narrows. Below 1280 px it stays one column, with a "Read the change" link
   at the top of the decisions. The page widens to 1600 px for the bench only.
2. **Show in the code.** Each *Outside the area* and *No test checks this* box has a **Show in the
   code** link that scrolls the change to that part's first added line (or the first line of the
   file's section, `diff --git`, for a whole-file part or one with no added line) and highlights it. Added lines carry their path and new line number for this.
   A part the change view does not show (past the 1,500-line cut) says so instead.
3. **Plain words.** One line under each question, saying what each answer means:
   - *Does the code keep this promise?* Read the change beside it. Kept: the code does what this
     sentence says. Not kept: it does not, and the review will need work. Inconclusive: you cannot tell
     from what is here.
   - *Are the words clear enough to verify?* This is about the sentence, not the code: could anyone
     check it and get a clear yes or no? If not, say what is unclear in the finding; the spec owner
     rewords it.
   - *Outside the area:* the spec lists the files this change may touch, and this one is not on it.
     Accept only if the change was needed, and say why.
   - *The tests* (owner, 2026-11-25: "test explanation is not really clear, make it simpler"): the
     heading is "Do the tests check this code?"; the line under it: Proofbound removes each piece of
     new code, one at a time, and runs the project's tests; if the tests still pass without a piece,
     no test checks that piece. The run is summed up in a sentence ("The tests pass on this change.
     With all the new code removed, they fail, so they do test it."), a part caught reads
     "Tested: …", a part not caught is a box titled "No test checks this" asking "Is that OK?", and
     the command is shown last, small. The diff flag *tests did not notice* becomes *not tested*,
     and the queue says "N part(s) untested". Only words change; the run and the decisions do not.
   - The finding field asks for a finding when the answer is not Verifiable (still optional, as the
     record allows).
4. **Promise to lines is not claimed.** Nothing records which lines keep which promise, so the page
   does not pretend to link them. The agent reviewer (D4) is where per-promise line citations would
   come from.

**Not in R3:** any change to the endpoints, the draft, the seal or the records.

**Tests:** `bench_test.go::TestBenchLayout`: the rendered bench has the decisions and evidence
columns, each plain-words line, a Show in the code link per area and tests box whose path and line
match an added diff line carrying the same path and number, and a whole-file part linking to its file
header. Contract row: CLI-INV-76 amended (the anchors and the link). Mutation calibration: a link to
the wrong line and a missing hint must die; a CSS colour change must survive.

## 11. Code review of R3 (owner-visible)

`ui-reviewer-r3-review-20261125`: **NEEDS_WORK** (two MED, four LOW); the MEDs fixed before commit:
- **F1 MED, fixed.** A whole-file part (binary, empty, mode only, deleted) has no `+++` line, so its
  link found nothing and the page said, falsely, that it was past the end of the view. Line 0 now
  sits on each file's `diff --git` line (quoted paths read too); a part with no added line lands on
  its file; the fallback says only that the file is not in the view.
- **F2 MED, fixed.** The tests' sentence claimed runs that did not happen (baseline fail, timeout or
  unreliable; whole change not run or not undone). It is now one sentence per result
  (`testsSummary`); "Tested:" shows only when the tests pass reliably; a part not run or not undone is
  titled "Not checked"; the queue says "they do not pass reliably" instead of "every part tested".
- **F3, F4, F6 LOW, recorded / partly fixed:** removed lines land on the file's first line; the JS
  link and the sticky column are not run by any test here (no browser): the owner checks them once
  by hand; F6's labels fixed with F2. **F5 LOW, fixed:** the old wording in this spec.

Mutation calibration for the fixes: a quoted path dropped, no anchor on `diff --git`, *unreliable*
read as pass, a whole change not caught read as not tried: each died; a CSS outline width survived.
