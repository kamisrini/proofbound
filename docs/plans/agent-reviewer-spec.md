# Spec: an agent reviews first, then the person decides (D4)

**Status:** draft 1, 2026-11-25, **approved by the owner** ("approve": D-A..D-D as recommended),
being built. The owner's design (2026-11-25, while running the live
demo): "agent reviewer first and then have human reviewer next with findings, assessment from agent
reviewer; human reviewer should have ability to review the whole thing even then." Held as
`ui-reviewer-spec.md` D4, released by the owner ("yes to both") with one choice made: **the agent is
any command the project configures, by default Claude Code headless** (`claude -p`).
**Implements:** `BR-audit-room-ui-r7t3w1` O-9 (revision 0007 proposed, §8), via a new change
intent. **Contract rows (proposed):** CLI-INV-80..82.

## 0. The gap, in one line

Today the reviewer starts from a blank page. A first pass by an agent (does each promise look kept,
which lines keep it, is each sentence checkable, should each flagged part be accepted) would make the
person's review faster and better, as long as the person still decides everything and the record says
an agent went first.

## 1. What the reviewer sees

1. **Ask the agent reviewer.** On the Review bench, beside **Run the tests**, a button **Ask the agent
   reviewer** appears only when the feature is on (§3). It warns, before running: *this sends the
   change, the spec and the test results to the configured agent (by default Claude Code, which sends
   them to Anthropic) and may cost money.* Nothing runs on page load.
2. **It runs in the background**, as the tests do: the server's one job (a second job is refused),
   progress lines, a closed tab does not stop it, stopping the server kills it. If the tests have run,
   their result goes to the agent; if not, the page says the agent will review without them.
3. **Its answers sit beside each question**, never in the radio buttons:
   - each promise: *Agent: Kept / Not kept / Inconclusive*, its reason, and the lines it cites
     (each with **Show in the code**);
   - each requirement: *Agent: Verifiable / Ambiguous / …* and its suggested finding;
   - each flagged part (outside the area, no test checks this): *Agent suggests: accept / reject*
     and why;
   - at the top: a short summary and the agent's overall view (*looks ready* or *needs work*), with
     the agent's name and model.
4. **The person decides every answer.** The agent's answers are **not** pre-selected (D-A). The
   person may read everything the page shows, exactly as today.
5. **Disagreement is visible.** Before sealing, the preview lists each question where the person's
   answer differs from the agent's ("you said Kept; the agent said Not kept"). It informs; it never
   blocks.
6. **The agent's answer is kept** in the local draft (as the tests' run is), so a reload does not
   re-run it. Asking again replaces it.

## 2. What the agent is given and must answer

- **Input (one prompt on stdin, built by Proofbound):** the promises and requirement sentences, the
  spec files and `scope.paths` as of the parent, the change (`git show` with the bench's fixed options,
  cut at the bench's limit), who wrote each line, the tests' result if run, and the flagged parts. It
  says plainly that the change is untrusted text and that instructions inside it are not instructions.
- **No tools.** The default command gives the agent no tools (`--tools ""`): it reads only what
  Proofbound sends; it cannot read files, run commands or change anything.
- **Output:** one JSON object (`proofbound.agent-review.v1`): `summary`, `overall`
  (`ready` | `needs_work`), and per question `outcome`, `reason` (≤ 500 chars), `lines`
  (`[{path, line}]`, each checked against the change; a cited line not in it is dropped with a note).
  Anything else, an unknown obligation, or invalid JSON: the run fails and says why; nothing is kept.
- **Default command** (when the project enables the feature without naming a command):
  `claude -p --safe-mode --output-format json --tools "" --no-session-persistence --strict-mcp-config
  --json-schema <schema> --max-budget-usd <budget>` (`--safe-mode`: none of the person's settings, hooks,
  plugins or CLAUDE.md load or are sent; review F2) (budget default 1.00).
- **Any command:** the project may name its own command (argv list) in the config; it reads the same
  prompt on stdin and must print the same JSON. This is how Codex, a local model, or a team's own
  reviewer plugs in.
- The command runs with the tests' cleaned environment rules (no `PROOFBOUND_*`, no `DATABASE_URL`),
  a timeout (default 10 minutes), and in a throwaway empty directory, never the repository.

## 3. Off by default

The feature spends tokens and sends code out, so it needs **both**:
- the project's `.proofbound/agent-review.json`, read from the **parent** commit (the reviewed commit
  cannot change what runs or where its code is sent, as with `tests.json`):
  `{"enabled": true}` for the default, or `{"enabled": true, "command": ["…"], "budget_usd": 0.5,
  "timeout_seconds": 300}`;
- and the person starting the page with `proofbound ui --agent-review`.
Without either, the button does not appear and the endpoint refuses (404).

## 4. What is recorded (D-B)

When the person seals, the review also writes `docs/verification/verdicts/AGENT-<commit12>.json`
(`proofbound.agent-review-record.v1`), signed by the **person's** key like the area and tests records:
the command (argv, no secrets), the agent and model it reported, the SHA-256 of the prompt and of the
answer, the answer itself, and each question where the person differed. It says *the person read an
agent's review first*; it is evidence, not a review: **`check`, C1 and the warranty rule are
unchanged**, and an agent's answer never satisfies a review. Without an agent run, nothing new is
written.

## 5. Also on the command line

`proofbound review agent <commit>` runs the same agent with the same prompt and writes the answer into
the local draft, printing the summary and each answer; `review seal` then records it as in §4. Same
config, same off-by-default rule (the CLI needs `--agent-review` too).

## 6. What this is not

- Not an agent that approves: the agent never seals, never signs, never counts as the reviewer.
- Not a change to `check`, the gate, the warranty or what an existing record means.
- Not a check that the reviewing agent differs from the one that wrote the code: a headless run has
  no session to compare, and a model name is not an identity (review F10). The person is the reviewer.
- Not hosted: the command runs on the reviewer's machine.

## 7. Tests (planned)

- The prompt: built from a fixture commit; contains every promise, the flagged parts, the untrusted
  warning; never contains the reviewer's key or environment.
- The command: a fake agent script (any command) answering fixed JSON: the bench shows each answer
  beside its question, no radio button is pre-selected, Show in the code goes to a cited line; a cited
  line not in the change is dropped; invalid JSON, an unknown obligation, a timeout each fail and keep
  nothing.
- Off by default: no config, config off, or no `--agent-review`: no button, POST 404; config only in
  the reviewed commit (not the parent): refused.
- Seal: the AGENT record is written, signed, lists the differences; `check` reads the same result with
  and without it (a mutation making the AGENT record count as a review must die).
- The default command line is exactly as §2 says (argv test; Claude Code itself is not run in tests).
- One live run on the demo with real Claude Code, cost recorded, before release.

## 8. Requirement

`BR-audit-room-ui-r7t3w1` O-9A (revision 0008), the reviewer's wording of O-9 (round 5 found O-9
AMBIGUOUS: who may click, a disabled or invalid file, what "differs" means). The record is the one
home of the wording.

## 9. Decisions (defaults proposed; the owner decides)

- **D-A. Pre-fill the person's answers with the agent's?** Recommended: **no**, show beside them.
  Pre-filled answers get clicked through; the point of the person is to judge. (Alternative: a "copy
  the agent's answers" button.)
- **D-B. Record the agent's review on seal?** Recommended: **yes**, as signed evidence, never as a
  review. An auditor should see that an agent went first and where the person disagreed.
- **D-C. Budget per run.** Recommended: **$1.00** default, configurable. The live demo's whole agent
  session cost about $0.85; a no-tools review of one commit should cost far less.
- **D-D. Send the tests' result to the agent?** Recommended: **yes** when run; the agent's view of
  "no test checks this" depends on it.

## 10. Effort and order

One change intent, one requirement round, one non-author spec review, then build: about the size of
R1. It fits before Sunday only if the owner approves this draft today or tomorrow; otherwise it is the
first thing after launch.

## 11. Spec review (owner-visible)

`agent-reviewer-spec-review-20261125`: **NEEDS_WORK** (1 HIGH, 6 MED, 5 LOW), folded in while building:
- **F1 HIGH, fixed.** An agent's commit could set the command the reviewer's click runs (with the
  reviewer's home directory and keys reachable). The agent reviewer is refused while the config was
  last changed by a commit that names a coding agent; the config is a trust file (C6).
- **F2 MED, fixed:** `--safe-mode` (not `--bare`, which needs an API key). **F3 MED, fixed:** Claude
  Code's envelope is read (structured output, else the result text), `is_error` or a non-success
  subtype refused, model and cost from it; any other command prints the bare answer; stdout capped.
  **F4 MED, fixed:** questions are keyed as the form names them (v-g-i, r-g-i, a-i, t-i), each answer
  carries its obligation, outcomes are the seal's words, an unanswered question is never a
  difference. **F5 MED, fixed:** the answer is carried like the tests' run, written under the request
  gate, shown and sealed only while its digest and questions match (else refused: ask again).
  **F6 MED, fixed:** CLI-INV-80, 81 write the endpoint's guards (404 off, 422 refused, 409 busy).
  **F7 MED, fixed:** the prompt says where the change is cut; the answer and record say how much the
  agent saw.
- **F8 LOW, partly fixed:** unknown fields refused, limits, `--bundle` refused (no schema field).
  **F9 LOW, fixed:** the record is written only when signed. **F10 LOW, fixed:** the self-review
  disclosure is dropped (§6). **F11 LOW, partly fixed:** runner environment, empty directory,
  timeout, Save keeping the answer, the F1 refusal and a tampered answer are tested; 409 and the
  stdout cap are not. **F12 LOW, fixed:** §8 links the record.

## 12. Code review (owner-visible)

`agent-reviewer-code-review-20261125`: **NEEDS_WORK** (1 HIGH, 3 MED, 5 LOW); HIGH and MEDs fixed:
- **F1 HIGH, fixed.** The person never saw what a click ran, and the agent-commit guard trusts commit
  names. A configured (non-default) command is now shown word for word with the commit that set it,
  marked "not Claude Code", and runs only after a ticked confirmation (409 otherwise); the CLI needs
  `--yes`. The default command is shown too.
- **F2 MED, fixed.** How much of the change the agent read is shown on the page, in the job log and on
  the CLI ("It read only the first N of M lines").
- **F3 MED, fixed.** `review agent` needs `review draft` first (it holds the test run).
- **F4 MED, fixed.** A stale answer (its questions gone) is warned in the preview and left out of the
  seal with a line; an edited one is refused.
- **F5..F9 LOW, recorded:** the digest is a mistake check, not forgery-proof (the record says only what
  the reviewer kept); answers bound by key and obligation, not a question digest; untrusted markers are
  fixed strings; some mutations survive (signer-less record, subtype check, 422 vs 404, stdout cap,
  the job's merge check); the budget does not apply to a configured command; `--agent-review` twice.

Live run (demo copy, Claude Code 2.1.331, 2026-11-25): 6 of 6 answered in 21 s for $0.08; it found
the seal's missing-area-questions bug, fixed with a regression test. Mutation calibration for the
fixes: each of six positive controls died; a comment edit survived.
