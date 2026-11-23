# Plan: the Proofbound UI (local first, hosted later)

**Status:** U1 and U2 BUILT 2026-11-22 (all five workspaces; see the U1 and U2 specs). APPROVED direction, 2026-11-22 (owner: "go ahead lets do these"; section 8 decisions
1 to 3 made). Planning and design only: nothing is built; U0 (decision record, obligation, change
intent) comes before any code. Owner direction: "lets go with 2 and then 3 later. plan and design. remember we
are visionaries ... We dont want UI that is 2026, but rather 2030 UI."
**Scope:** option 2, a local UI served by the `proofbound` binary (`proofbound ui`), designed so the
hosted, multi-repository UI (option 3) is the same product in a different place, not a rewrite.
**Primary customer** (owner, 2026-11-22): compliance and auditors, "who want to keep track of agents
doing what they are supposed to, nothing outside the authorized perimeter". The Audit room is
therefore built first (section 7), and section 2 says what of the perimeter v1 can and cannot show.
**Visual design:** the private canvas "Proofbound Ledger UI concept" (one screen per workspace), linked from the
owner's session; this page is the text of record.

## 1. What a 2030 UI is (from the vision, not from today's dashboards)

The vision says what dies: "status fields and dashboards that quietly lie", "sprint planning,
standups, status decks", "documentation as an artifact" (vision-2028 section 5, vision-100x
section 4). A 2026 developer dashboard (repo list, a table of builds, charts of counts, a settings
sidebar) is exactly that furniture. We design from the axioms instead:

| Axiom (vision-100x section 5, vision-2028 section 2) | What it means on screen |
|---|---|
| **Attention is sacred** (axiom 8, Law 8: "the decision inbox is measured in minutes per day") | The UI opens on what needs a human, with an estimated number of minutes, and nothing else competes for the eye. An empty state reads "Nothing needs you", and that is the success state. |
| **The record is the reality** (axiom 1, Law 1: "status is not a writable field") | Every value on screen is a fold over the ledger, and says so: each number and state opens to the events it came from. The UI has no write path of its own. |
| **Behavior is born proven** / every claim carries its proof (vision-2028 section 5) | The primary object is the **promise** (an obligation), not a file, repo or commit. Each promise shows its proof chain: spec text at a revision, the claiming commit, the independent review, the signatures, the warranty. |
| **Time is navigable** (axiom 5, Wonder 2) | A time rail over the ledger: drag to any moment and the whole UI shows what was true then; compare two moments. |
| **Trust decays** (vision-2028 Trust Engine: "Done is not a terminal state") | Verification has an age. A promise whose spec moved after it was verified shows as STALE, not green. |
| **Trust composes / verifies locally** (Wonder 3: "its proof chain verifies locally in milliseconds") | Signatures are checked in the browser (WebCrypto Ed25519) against the repository's declared keys, so the page shows "verified here", not "the server says". The same code later lets anyone verify a hosted page without trusting us. |
| **Rendered on demand** (Wonder 1: "a UI synthesized for this user, this task, this moment") | v1 has a small set of fixed lenses; each is a composition of a few primitives (promise, chain, moment, attention item), so later lenses can be composed per person or per task, by them or by an agent. |
| **Agents first** (Proofbound is for agents coding) | Every view has a machine twin: the same URL with `Accept: application/json` returns the data the view rendered, and every suggested action is the exact CLI command, so an agent and a human act on the same surface. |
| **Meaning is the only handwriting** (axiom 3) | The UI writes nothing a person did not decide, and everything a person decides can be done in it, without a terminal. A human act (an exception outcome, a review, a countersignature) is previewed exactly as it will be written, confirmed, then signed on the person's own machine with their own key and written as the same record the CLI writes. |

## 2. The five users, and a UI for each one's work

Owner, 2026-11-22: these five are the main users, and each gets a UI specialised for the function
they perform. There are no external users yet, so each row is a hypothesis to test with the first
design partners.

One product, five **workspaces**. A workspace is a composition of the same primitives (promise,
chain, moment, attention item) for one job, not a permission level: locally, anyone can open any
workspace. Every workspace is read from the same ledger snapshot, so two people looking at the same
promise can never see different truths. Each one opens on the single question that person comes
with, and shows nothing that does not help answer it.

| # | User | Their function | The question they open with | Workspace | Where it runs |
|---|---|---|---|---|---|
| 1 | **Developer running coding agents** | Directs agents, merges their work | "What do my agents' changes need from me before they can merge?" | **Now**: attention items with minutes and the exact command; agents at work and the task each holds; promises at risk on the branch | local |
| 2 | **Reviewer** (a person, or a reviewer agent in another session) | Independently checks a change against its promises | "What is waiting for someone who did not write it, and exactly what am I checking?" | **Review bench**: a queue of changes this person is eligible to review (never their own); one promise at a time beside the code that claims to keep it; an outcome per promise | local |
| 3 | **Spec owner** | Decides what should be true | "Which of my promises hold, which drifted when I changed the words, and which did reviewers find unclear?" | **Promise map**: every promise they own, its state, drift, and the requirement-review findings (AMBIGUOUS, UNTESTABLE, CONTRADICTORY) that say which words to fix | local |
| 4 | **Compliance and auditors** (primary customer) (internal compliance, external auditors, a customer's security reviewer) | Establish, for a period or a release, that the controls held, without trusting the team | "For this release or period, did every change follow the rules, and can I check that myself?" | **Audit room**: pick a release or date range; each control is a question with an answer, a count, its exceptions and the events behind it; reproduce it from a clone; export a transcript | local first (on a clone, offline), then hosted |
| 5 | **Coding agents** | Write the code, claim the spec, ask what is next | "For my task, what is unproven and what is the next command?" | **Agent surface**: no pixels; the JSON twin of every workspace plus a task-shaped endpoint (`/agent/next?task=sdd:<change>`) returning open promises, their state and the next command; a tool server (MCP) later | local |

**What each workspace never shows:** the developer is not shown other people's review queues; the
reviewer is not shown changes they authored (Law 9: they are not eligible); the spec owner sees
code only as evidence under a promise; the auditor sees no "needs you" items, only answers and
evidence; the agent gets no prose it has to interpret, only typed states and commands.

**Audit room controls in v1** (each answerable from today's ledger, each re-computable):
- Every change that claims a spec was reviewed by someone who is neither its author nor its spec owner.
- No code entered the release without a spec claim (no UNCOVERED commits), or each exception is listed.
- Every warranty in the release verifies (VALID), checked in the auditor's browser against a clone.
- Every reviewer's signature is by a key the repository declared before the review.
- Which promises changed after they were verified (drift), and when.
- The answers are reproducible: rebuilt from git alone, the same answers (`rebuild`, `sync all` on a clone).
The transcript the auditor exports lists each control, its answer, its exceptions and the event
numbers, so another auditor can re-run it and get the same result (vision-2028: "audit = a query +
cryptographic transcript"). An auditor countersignature is reserved for later.

### The authorized perimeter: what the Audit room can show, and when

The controls above prove **authorization**: each change claimed a spec and someone else reviewed it.
The primary customer asks the harder question, **did the agent stay inside what was authorized?**
Three tiers, by what data they need.

**Tier 1, in the first Audit room (from data the ledger and git already hold):**

| View | The auditor's question | What it shows |
|---|---|---|
| **Every commit accounted for** | "Is this every change, or the ones you chose to show me?" | Each commit in the range classed as claimed and reviewed, claimed and unreviewed, housekeeping only (specs, docs, records), or UNCOVERED, with the counts summing to the range. Whether the ledger can also flag a rewritten history (a commit it ingested that is no longer reachable) is a spec question for U1. |
| **Changes to the rules themselves** | "Did an agent move its own fence?" | Every commit in the range that touches the rule surface: `specs/` (the wording of the authorization), `.proofbound/` (keys, records), `gates/`, and the CI workflow that runs the check. CLI-INV-33 deliberately exempts these paths from UNCOVERED, so today they are never an exception; here each one is listed with its author and whether it went through a reviewed claim, and one without independent review is an exception. |
| **Who did what** | "Did anyone grade their own work?" | A grid per change: author, reviewer, spec owner. Any change where two of the three are the same identity is flagged (Law 9 made visible). |
| **Authorized before acted** | "Was the permission there before the work?" | For each claiming commit, whether the bound spec revision existed, worded as reviewed, in a commit before the code. A spec written in the same commit as its code, or reworded after it, is authorization filled in after the fact and is listed. The exact rule is a spec question for U1. |

These four become controls 7 to 10 in the transcript, each with its answer, exceptions, event numbers
and the command to re-run it. Their re-run commands on the canvas (`proofbound audit rules|roles|order <base>`)
are proposed names, not existing commands; U1's spec settles them.

**Tier 1, later in the Audit room (same data, less urgent):** a **reproducible sample** (the auditor
tests N changes in depth; the sample is drawn from the range's hash, so a second auditor gets the
same N); **since my last visit** (bookmark an event number, see which control answers moved); and a
**framework mapping** in the transcript (SOC 2 CC8.1, ISO 27001 A.8.32, EU AI Act Articles 12 and 14),
worded "maps to", never "certifies", and checked by a compliance professional before it ships.

**Tier 2, the literal "nothing outside the perimeter" (needs kernel work first):**
- **Perimeter per change:** a spec declares the code it covers, and a hunk outside it is OUT OF SCOPE
  even when the commit claims the spec correctly. Depends on `DEF-CLAIMED-CODE-SCOPE-001` and option 1
  of `DEF-AGENT-NATIVE-INTENT-001`. Recommended as the next kernel milestone after the UI.
- **Exceptions with a disposition:** each exception gets a signed outcome from a person (risk
  accepted, fixed in commit X, reverted), or is pre-approved as a signed waiver from someone other
  than the author (`DEF-NO-INTENT-SELF-WAIVER-001`). "1 exception" becomes a closed loop.

**Tier 3, actions outside git (vision, later):** an agent-session record (commands run, secrets
touched, network calls, deploys). Until it exists Proofbound sees only what lands in the repository,
and the Audit room says so on screen: it never implies an agent stayed out of production.

### Interactive, for people who never open a terminal

Owner, 2026-11-22: "is there a way we can provide more value by making them interactive, so people
can interact without cli, especially non developers". Yes: the primary customer is not a developer.
Two kinds of interaction (section 8, decision 2, as revised):

**Explore (writes nothing; everywhere, from U1).** Pick a release or range; filter to exceptions;
click a control to see the commits and events behind its answer, and a commit to see its diff and
proof chain; draw the reproducible sample; compare two releases; drag the time rail; export the
transcript as a readable page, PDF or JSON.

**Act, signed, without the CLI.** `proofbound ui` runs on the person's own machine, as them, next to
their key, so a button can run the same code path as the CLI command. Every act shows a preview of
the exact record it will write and the key that will sign it; nothing is written until the person
confirms. The record is the same signed file the CLI writes, so the ledger, replay and every check
cannot tell the two apart.

| Workspace | Act | Record written | Step |
|---|---|---|---|
| Audit room | **Record an outcome on an exception**: accepted risk, fixed in commit X, or reverted, with a reason | a signed exception-outcome record (new type) | U1 |
| Audit room | **Countersign the transcript** | a signed auditor attestation (new type) | U1 |
| Review bench | Choose outcomes per promise, then **seal and sign** | the same verdict and `.sig` files as `review seal` | U3 |
| Now | **Hand this to my agent** | `.proofbound/task` (what `task start` writes) | U2 |
| Promise map | Reply to a finding / **propose new wording** | a proposed spec change on a branch; it changes the rules, so it needs non-author review like any rule change | later |

The two new record types need spec work in U0. A commit that carries only records signed by declared
keys counts as signed, not as an unreviewed rule change, in the "changes to the rules" view. Writing
safety: the existing loopback `Host` check, plus an `Origin` check and a per-session token on every
act, so no other web page can trigger one.

**Without git (U1).** `proofbound ui --bundle <file>` opens the evidence bundle the pull-request
check already produces, so an auditor with no clone and no CLI can open it in a browser and have the
page verify the signatures itself. The hosted option 3 replaces this later.

## 3. What v1 shows (only data Proofbound records today)

No screen shows a capability that does not exist. Today's ledger holds: sdd changes and their
obligations, claiming commits and their bound spec revision, UNCOVERED commits, readiness gates
(PASS, BLOCKED, NO REVIEW), v2 verdicts and requirement reviews with reviewer identities and `.sig`
files, warranties (VALID, UNREVIEWED, INVALID), drift (CURRENT, STALE), declared keys, the active
`proofbound task`, and every event with its sequence number and time.

The lenses below are the building blocks the workspaces in section 2 are composed from.

**Lens 1: Now (the developer's home).**
- One generated sentence: "2 things need you, about 6 minutes. Everything else is proven or waiting
  on machines." Derived from the ledger, never written by anyone.
- **Attention** (left): only items a person must act on, each with why it is theirs, an estimate in
  minutes, and one action as the exact command (copy, or "hand to my agent"). v1 sources: NO REVIEW
  (needs a reviewer who is not the author), BLOCKED, STALE drift (re-review), UNCOVERED commits
  (claim or revert), UNREVIEWED warranties (cosign).
- **The promise field** (centre): every promise grouped by change, each with its state, who verified
  it, how long ago, and a mark when the spec moved since. No files, no commit list.
- **Since you last looked** (right): what changed, derived from events after the viewer's last visit
  (kept in browser storage, per viewer).
- **Agents at work** (top): active `proofbound task` claims per worktree, so a person sees which
  change each agent is working on.

**Lens 2: Proof chain (one promise).** The chain as a vertical sequence, each link with its hash and
event number: spec text at revision -> claiming commit -> independent review (reviewer, outcome,
finding) -> signatures (each verified in the browser) -> warranty -> later events (drift). Each link
says what would break it.

**Lens 3: Moment (time).** The time rail at the bottom of every lens; dragging it sets "as of event N".
v1 computes an as-of view by replaying the ledger through N in a disposable store (the existing
`twin replay --through <seq>` machinery), cached per N; "compare with now" lists what differs.

**Lens 4: Pull request.** `proofbound check <base>` as a page: WARRANTED / NOT WARRANTED, each
commit's state, UNCOVERED commits, drift. The same data the GitHub Action prints.

**Ask (off by default).** A free-text "ask the record" box would spend model tokens, so it ships
behind an opt-in flag that is off by default (owner rule, 2026-11-22). Without it, the same box is a
deterministic finder over promises, changes, commits and events.

## 4. How it runs (boring kernel, 2030 surface)

- `proofbound ui [--port N] [--open]` serves on `127.0.0.1` only, prints its URL, and stops on
  Ctrl-C. Go standard library only (`net/http`, `html/template`, `embed`): no new dependency, so no
  decision record is needed for a library (Law 8). Browser code is plain modern JavaScript and CSS
  embedded in the binary: no build step, no npm, no framework, no CDN, works offline.
- **The ledger lock (real constraint, found in planning).** The embedded store takes an exclusive
  lock per repository (`store.acquireLock`). A UI that held it open would block `sync`, `check` and
  every other command in other terminals. So the UI never holds the store between refreshes: it
  opens the store, syncs, reads a snapshot into memory, and closes it, then serves the snapshot.
  It refreshes when git HEAD, the specs or the review files change (polled every few seconds), and
  pushes the new snapshot to open pages (server-sent events), so the page is live without a refresh
  button. What a concurrent command sees during the seconds of a refresh is a spec question (wait or
  fail with a clear message).
- **Local-only safety.** The only write endpoints are the signed acts in section 2, each behind a
  preview and a confirmation, an `Origin` check and a per-session token. Refuse requests whose `Host`
  header is not the loopback address and port (DNS-rebinding guard). Every untrusted string (commit messages,
  spec text, reviewer findings) is escaped by `html/template`; no inline script built from data.
- **Machine twin.** Each lens at `/now`, `/promise/<id>`, `/moment/<seq>`, `/check/<base>` returns
  HTML for a browser and JSON for `Accept: application/json`, from one handler, so the two cannot
  drift (Law 2).
- **Proof stays honest.** The page labels what it verified itself (signatures, in the browser) and
  what it reports from the ledger. It never shows green for something it has not checked.

## 5. Option 3 later: the same UI, hosted

The hosted, multi-repository UI is the same lenses over many ledgers. What v1 must not foreclose:
- Views read a snapshot, not the database, so a hosted service can serve snapshots that each
  repository's CI pushes, without holding anyone's database.
- Browser-side signature checks mean a hosted page can be verified without trusting the host.
- The JSON twin is the hosted API.
Needs before it starts: real design partners asking, hosting, authentication, tenancy, and a
decision record. Commercial terms stay out of this repository.

## 6. Not in v1 (reserved, so the layout has room)

The decision inbox of Governor v1, twin futures ("show me the quarter where we delay X"), ambiguity
adjudication ("which of these two futures do you mean?"), value theses and the morning settlement
report across a fleet. Each is a frozen organ (`VD-vertical-mvp-first-ai3dhn`); the Attention lens is
designed as the inbox's first, honest version, and the time rail as the twin's.

## 7. Build steps (each ends in one command the owner can run, Law 11)

| Step | Delivers | The command |
|---|---|---|
| U0 | A decision record: the UI direction above, stack (stdlib + embedded vanilla JS), local-only and read-only, the lock strategy; a requirement obligation and change intent | — |
| U1 | `proofbound ui` with the **Audit room** (primary customer): release or range scope, controls 1 to 10 (section 2, including the four perimeter views), click-through evidence, exceptions with signed outcomes, sample, transcript export and countersign, `--bundle`, Proof chain, browser signature check, the JSON twin of each | `proofbound ui --open` on the tour's `shop` project, open its release range |
| U2 | The **Now** workspace (developer) and the **Agent surface** (`/agent/next`), live refresh | `proofbound ui --open` while an agent holds a `task start` |
| U3 | **Review bench** (reviewer, seals and signs in the browser), **Promise map** (spec owner), time rail and as-of replay | review and sign a promise in the bench, no terminal |
| U4 | Tour chapter and DEMO-MANUAL step per workspace; release | next release |

Each step: a spec section first (Law 6, a `SPEC.md` for a new `internal/ui` package), one spec review,
one code review (Law 11), tests that derive from the spec, `make check` green.

## 8. Decisions the owner made before U0

1. **Release.** DECIDED 2026-11-22, then changed the same day: folded into v0.3.2 (no one outside had it
   before the 2026-11-30 launch), not v0.3.3. Originally: the UI ships as the next release (v0.3.3); v0.3.2 stays as
   published.
2. **Interactive, signed acts without the CLI.** DECIDED 2026-11-22, replacing the earlier
   "read-only except the bench's local draft": exploring writes nothing; each act in section 2 is
   previewed, confirmed and signed locally with the person's own key, writing the same record the CLI
   writes. Exception outcomes and transcript countersignature in U1; bench sealing in U3; spec wording
   later.
3. **Order of workspaces:** DECIDED 2026-11-22: compliance and auditors are the primary customer,
   so the Audit room is U1 (section 7).
4. **Ask (natural-language query):** confirm it stays behind an off-by-default flag, and out of v1.

## 9. Risks

- **Scope gravity.** A UI invites width (settings, filters, charts). The test for every element:
  does it spend attention only where a person is needed, and does it carry its proof? If not, it is
  2026 furniture.
- **A browser that can sign.** Signed acts make the local server a signing service; a missed origin
  or token check would let another page sign as the user. The spec states the threat model, and the
  code review tests it with a forged cross-origin request.
- **Overclaiming the perimeter.** Tier 1 shows authorization and its paper trail, not that every
  line stayed in scope (Tier 2) or that the agent did nothing outside git (Tier 3). The Audit room
  must state both limits on screen; a compliance product that overstates its coverage is worse than
  none.
- **Lock contention** (section 4) is the one technical unknown; the spec must settle it first.
- **Browser Ed25519** is in current Chrome, Firefox and Safari; older browsers fall back to "reported
  by the ledger, not verified here", labelled as such.
- **Meta-tax** is 74%; the UI is product work and should lower it, provided reviews stay within the
  Law 11 budget.
