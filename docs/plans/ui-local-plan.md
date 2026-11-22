# Plan: the Proofbound UI (local first, hosted later)

**Status:** DRAFT for the owner, 2026-11-22. Planning and design only: nothing here is built or
authorized to build. Owner direction: "lets go with 2 and then 3 later. plan and design. remember we
are visionaries ... We dont want UI that is 2026, but rather 2030 UI."
**Scope:** option 2, a local UI served by the `proofbound` binary (`proofbound ui`), designed so the
hosted, multi-repository UI (option 3) is the same product in a different place, not a rewrite.
**Visual design:** the private canvas "Proofbound Ledger UI concept" (two screens), linked from the
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
| **Meaning is the only handwriting** (axiom 3) | The UI writes nothing a person did not decide. Human acts (review, sign) stay where the private keys are: the CLI or the agent's session. The UI hands them over as commands. |

## 2. What v1 shows (only data Proofbound records today)

No screen shows a capability that does not exist. Today's ledger holds: sdd changes and their
obligations, claiming commits and their bound spec revision, UNCOVERED commits, readiness gates
(PASS, BLOCKED, NO REVIEW), v2 verdicts and requirement reviews with reviewer identities and `.sig`
files, warranties (VALID, UNREVIEWED, INVALID), drift (CURRENT, STALE), declared keys, the active
`proofbound task`, and every event with its sequence number and time.

**Lens 1: Now (the home).**
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

## 3. How it runs (boring kernel, 2030 surface)

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
- **Read-only, local-only safety.** No write endpoints. Refuse requests whose `Host` header is not
  the loopback address and port (DNS-rebinding guard). Every untrusted string (commit messages,
  spec text, reviewer findings) is escaped by `html/template`; no inline script built from data.
- **Machine twin.** Each lens at `/now`, `/promise/<id>`, `/moment/<seq>`, `/check/<base>` returns
  HTML for a browser and JSON for `Accept: application/json`, from one handler, so the two cannot
  drift (Law 2).
- **Proof stays honest.** The page labels what it verified itself (signatures, in the browser) and
  what it reports from the ledger. It never shows green for something it has not checked.

## 4. Option 3 later: the same UI, hosted

The hosted, multi-repository UI is the same lenses over many ledgers. What v1 must not foreclose:
- Views read a snapshot, not the database, so a hosted service can serve snapshots that each
  repository's CI pushes, without holding anyone's database.
- Browser-side signature checks mean a hosted page can be verified without trusting the host.
- The JSON twin is the hosted API.
Needs before it starts: real design partners asking, hosting, authentication, tenancy, and a
decision record. Commercial terms stay out of this repository.

## 5. Not in v1 (reserved, so the layout has room)

The decision inbox of Governor v1, twin futures ("show me the quarter where we delay X"), ambiguity
adjudication ("which of these two futures do you mean?"), value theses and the morning settlement
report across a fleet. Each is a frozen organ (`VD-vertical-mvp-first-ai3dhn`); the Attention lens is
designed as the inbox's first, honest version, and the time rail as the twin's.

## 6. Build steps (each ends in one command the owner can run, Law 11)

| Step | Delivers | The command |
|---|---|---|
| U0 | A decision record: the UI direction above, stack (stdlib + embedded vanilla JS), local-only and read-only, the lock strategy; a requirement obligation and change intent | — |
| U1 | `proofbound ui` with Now and Proof chain lenses, live refresh, JSON twin, browser signature check | `proofbound ui --open` on the tour's `shop` project |
| U2 | Moment (time rail, as-of replay, compare) and the Pull-request lens | drag the rail to before the review; the PR lens for `<base>` |
| U3 | Tour chapter and DEMO-MANUAL step; release | next release |

Each step: a spec section first (Law 6, a `SPEC.md` for a new `internal/ui` package), one spec review,
one code review (Law 11), tests that derive from the spec, `make check` green.

## 7. Decisions the owner should make before U0

1. **Release.** v0.3.2 is published, so the UI ships as the next release (v0.3.3). Re-cutting v0.3.2
   is possible but not recommended.
2. **Read-only v1** (recommended), or should the UI also drive actions (start a task, draft a review)?
   Signing stays in the CLI either way, because private keys stay local.
3. **Lens scope for U1:** Now + Proof chain (recommended), with time travel in U2, or all four at once.
4. **Ask (natural-language query):** confirm it stays behind an off-by-default flag, and out of v1.

## 8. Risks

- **Scope gravity.** A UI invites width (settings, filters, charts). The test for every element:
  does it spend attention only where a person is needed, and does it carry its proof? If not, it is
  2026 furniture.
- **Lock contention** (section 3) is the one technical unknown; the spec must settle it first.
- **Browser Ed25519** is in current Chrome, Firefox and Safari; older browsers fall back to "reported
  by the ledger, not verified here", labelled as such.
- **Meta-tax** is 74%; the UI is product work and should lower it, provided reviews stay within the
  Law 11 budget.
