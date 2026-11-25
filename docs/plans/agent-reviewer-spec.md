# Spec: an agent reviews first, then the person decides (D4)

**Status:** draft 1, 2026-11-25, **not built**. The owner's design (2026-11-25, while running the live
demo): "agent reviewer first and then have human reviewer next with findings, assessment from agent
reviewer; human reviewer should have ability to review the whole thing even then." Held as
`ui-reviewer-spec.md` D4, released by the owner ("yes to both") with one choice made: **the agent is
any command the project configures, by default Claude Code headless** (`claude -p`).
**Implements (proposed):** a new obligation on `BR-audit-room-ui-r7t3w1` (§8), via a new change
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
  `claude -p --output-format json --tools "" --no-session-persistence --strict-mcp-config
  --json-schema <schema> --max-budget-usd <budget>` (budget default 1.00).
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
- Not the agent that wrote the code reviewing itself: if the agent's reported model/session is the
  commit's recorded authoring session, the page says so in the summary (it cannot be prevented for an
  arbitrary command; it is disclosed).
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

## 8. Requirement (proposed wording for O-9)

"Agent reviewer first: when a project's parent commit enables it and the reviewer starts proofbound ui
with --agent-review, the reviewer can, with a click and never on page load, send a commit waiting for
review to the configured agent command (by default Claude Code headless with no tools), and then sees
the agent's answer beside each question it covers without any answer being selected for them; the
seal records the agent's answer and every question where the reviewer's answer differs, signed by the
reviewer, and the readiness check's result is the same with or without that record."

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
