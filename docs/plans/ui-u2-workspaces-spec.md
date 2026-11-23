# Spec: U2, the other four workspaces (Now, Review bench, Promise map, Agent surface)

**Status:** draft 1, 2026-11-22, reviewed together with the code by one non-author (owner asked for the
whole set in one round: "All four, fully interactive").
**Implements:** `CI-ui-workspaces-u2-p3k7d4` → `BR-ui-workspaces-m6q2v8` O-1..O-4.
**Builds on:** [U1](ui-u1-audit-room-spec.md) (server, guards, signed acts) and [the plan](ui-local-plan.md)
section 2. **Contract rows:** CLI-INV-45..49.

Every workspace is a page of the same `proofbound ui` server, under the same guards (U1 section 4),
computed from the same audit of the same period, so two people never see two truths. Each has a JSON
twin at the same address (`Accept: application/json` or `format=json`).

## 1. Promise map (`/promises`, O-3)

`computePromises` reads the ledger once. For each spec change (the latest `intent.sdd` change intent
per id), for each obligation its change intent targets:
- **statement and owner** from the targeted requirement revision;
- **state**, first match: `PROVEN` (a v2 verdict SATISFIED for this obligation of this requirement
  revision), `NOT KEPT` (NOT_SATISFIED), `INCONCLUSIVE`, `CHANGED AFTER PROOF` (SATISFIED only for an
  older revision of the same requirement), `CLAIMED, NOT REVIEWED` (a ledgered commit binds the change),
  `NOT CLAIMED`;
- **unclear**: the latest requirement review of this revision says AMBIGUOUS, UNTESTABLE or
  CONTRADICTORY; shown with its finding and reviewer, whatever the state;
- the reviewer of the deciding verdict, and the commits that claim the change.

## 2. Now (`/now`, O-1)

Attention items, each with a kind, a sentence, why it needs a person, the exact command, and a link
to the workspace that acts on it:
- a claiming commit with NO REVIEW → Review bench (`proofbound review draft <commit12>`);
- a BLOCKED or UNBOUND claiming commit → its reason;
- an UNCOVERED commit → `proofbound task start <change>` before committing code, or an outcome in the
  Audit room;
- a promise `CHANGED AFTER PROOF`, `NOT KEPT` or unclear → Promise map.
**Agents at work:** each worktree (`git worktree list --porcelain`) with its active task
(`.proofbound/task`), branch and last commit. **Since you last looked:** the page remembers the last
HEAD it showed in the browser's storage and marks newer commits; nothing is stored on the server.

## 3. Review bench (`/review`, `/review?commit=<sha>`, O-2)

**Queue:** claiming commits in the period whose state is NO REVIEW. A commit is shown only if the
viewer (the local key's declared name) is not, by key name, its author, an agent named on it, its
verified signer, or an owner of a requirement it targets; the page says how many were hidden and why.
**One commit:** the ledger draft (`ledgerDraft`: every requirement-review and verdict obligation, with
its statement), the commit's change (`git show --stat --patch`, first 1,500 lines), and any local
draft's decisions. Per obligation the reviewer chooses SATISFIED, NOT_SATISFIED or INCONCLUSIVE; per
requirement obligation VERIFIABLE, AMBIGUOUS, UNTESTABLE or CONTRADICTORY with a finding.
**Acts** (POST, U1 guards): `/act/review-draft` saves the decisions to
`.proofbound/drafts/<sha>/draft.json` (the local draft, never committed); `/act/review-preview` lists
the files sealing would write and the decisions; `/act/review-seal` re-checks eligibility, writes the
draft, runs the existing `reviewSeal` (same validation, same signing, all-or-nothing), and commits only
the files it wrote with message `review <commit12>`; if the commit fails, those files are removed. The
seal's own refusals (an owner reviewing, an undecided obligation, a key mismatch) reach the page as
written.

## 4. Agent surface (`/agent`, `/agent/next`, O-4)

`/agent` documents the endpoints for a person. `GET /agent/next[?task=sdd:<change>]` (default: this
worktree's task) returns JSON only: the task, its promises (section 1), the claiming commits with their
`check` states, the period's uncovered commits, `blocking` items (each with `you_may`: an agent may
never review a change it wrote), and `next` steps, each a command or `stop_and_ask_spec_owner`.
**Non-browser clients** may send `Authorization: Bearer <agent token>` on GET requests instead of the
cookie; the agent token is a second token, separate from the person's, and only reads. The server writes `.proofbound/ui.json` (`{"url","token","pid"}`, holding the agent token, mode 0600, gitignored with
the rest of `.proofbound/`) while it runs, so an agent of the same user can find it, and removes it on
exit. POST still needs the cookie and the page's Origin: agents read, people act.

## 5. Not in U2

Time rail and as-of replay, in-browser signature checks, hosted mode, an MCP tool server.

## 6. Tests the rows derive

`TestPromiseStates` (the state rule on built events, no store), `TestNowItems` (attention items from a
report and promises), `TestBenchEligibility` (who may review), `TestUISealFlow` (draft, preview and seal
through the server against a stubbed ledger draft, committing only the sealed files),
`TestAgentNext` (the next-step rule and the bearer token), `TestPromiseLatestVerdictWins`.

## 7. Review (`ui-u2-review-20261122`, NEEDS_WORK: F1..F4 MED, F5..F7 LOW)

F1: the agent's bearer token is now a second, read-only token (what `ui.json` holds); it never opens
the person's page. F2: the exported SSH key is created 0600 exclusively. F3: `keys git-sign` sets the
signer per worktree (`extensions.worktreeConfig`, `git config --worktree`). F4: on earlier revisions
only the latest verdict counts. LOW adopted: F5 (GPG signatures can never verify as a declared key),
F6 (an existing export must be a regular 0600 file holding this key). LOW recorded: F5's "a key
declared under two names gets one principal", F7 (`you_may` is always false today; the HTTP seal test
stubs `reviewSeal`, which the demo-repo acceptance run exercises for real).
