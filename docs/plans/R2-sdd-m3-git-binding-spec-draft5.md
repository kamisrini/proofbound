# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 5)

**Status:** draft 5, 2026-11-19. It answers non-author round 4,
[`r2-m3-spec-review-round4-20261118`](../verification/verdicts/r2-m3-spec-review-round4-20261118.md)
(NEEDS_WORK: 2 MED, 2 LOW), over [draft 4](R2-sdd-m3-git-binding-spec-draft4.md). Rounds 1-3 and their
answers are in draft 4 §0, §0a and §0b; they are not repeated here. **Owner decisions, 2026-11-19:**
D1-D5 all take the recommended option (§3). It still needs a non-author re-check (Law 9) and the
owner's approval of the change-intent wording (§8) before any code lands (Law 6). **Parent contract:**
`kernel/internal/connector/intent/sdd/SPEC.md` (R2 M1, accepted). **Requirement served:**
`BR-sdd-warranty-i9miv7` O-8 and O-10 ("fails only a claim on that state").

## 0c. Round 4, answered

| Finding | Answer in draft 5 |
|---|---|
| F1 MED — row 8 cannot yield PASS on a correct implementation (reviews before git; a static verdict; one gate file) | §7 row 8 is rewritten as an explicit procedure (§7.1). The test lives in package `cli`. It synthesises the tenant at run time, syncs **git before reviews**, reads the evidence event id and the ledgered digests back from the store, writes the verdict and review with their self-digests, commits them as a new HEAD, and writes **one exact-sha gate definition per gated commit** (JSON, despite the `.yaml` suffix). Canary output is matched as `state=PASS` / `state=BLOCKED` |
| F2 MED — a verdict ingested before the binding it cites wedges projections, and M3 makes that reachable from correct data | **D5 (owner, 2026-11-19): the reviews connector holds such a verdict** (§4.6). A v2 verdict or requirement review is appended only when every ledger prerequisite its reducer looks up is already ledgered; otherwise it is reported as `held` and picked up by a later sync. This closes the class for **every** provider, `records` included. M3-INV-5 and M3-INV-10 are restated, and M3-INV-11 is new (§6). §7 row 5 now includes the round-4 `sync all` scenario |
| F3 LOW — the two-claim kill cannot die through the gate | It can now. Under D5, the two-claim commit's verdict is **held** under correct code (D2 binds no ref, so its claim is not ledgered), and the commit is BLOCKED. Under the partial-binding mutant, the claim is ledgered, the verdict is appended and projects, and the commit reads PASS. So the cell dies in row 8, and row 3 keeps its unit-level kill too |
| F4 LOW — §9's revisit trigger has no home future authors read | §5 now carries the accepted limitation and its trigger into the sdd SPEC §14 (a new M2 entry condition) and into the git SPEC (beside the new INV rows). §9 links to them; the plan is no longer the only home |

**Owner decision D4 is filed** as `DEF-INTENT-GIT-ORDERING-001` in `docs/deferrals.tsv` (round-1 F8,
closed as owned).

## 1. Evidence

As draft 4 §1, plus these facts read at HEAD `92c6027` for D5:
- **What the v2 verdict reducer looks up** (`projections.go` `reduceObligationVerdict`):
  1. `commits_view` holds `reviewed_commit` ("reviewed commit is dangling");
  2. `commit_intents_view` holds the exact `(reviewed_commit, provider, id, digest)` claim ("does not
     match an exact commit intent claim");
  3. `requirements_view` holds every listed requirement revision ("requirement revision is dangling");
  4. `intent_targets_view` rows for the change intent (projected with the change intent itself);
  5. every evidence event id exists in `events` at a sequence number no higher than the verdict's
     ("dangling" / "from the future").

  Items 1, 2, 3 and 5 can be absent only because of **order**: a later sync can supply them. Item 4
  is present whenever item 2 is, because a claim binds only to a projected change intent (§4.3 for
  `sdd`; the dangling check for the others). Every other failure in the reducer (coverage, outcome
  values, ACCEPTABLE with a non-satisfied outcome) is about the verdict's **content**, and no later
  sync can repair it.
- **What the requirement-review reducer looks up** (`reduceRequirementReview`): `requirements_view`
  holds the exact requirement revision. That is also order-only. The rest is content.
- **v1 verdicts** (`vera.verdict.v1`) look up nothing in other views. They are unaffected.
- **The reviews connector** appends every parsed artifact unconditionally (`reviews.go` `Sync`). Its
  only dependencies are a reader, an id generator and a logger (`Deps`).
- **This repository holds no v2 verdict and no requirement review** (`grep -rl` over
  `docs/verification/verdicts` for both schemas: 0 files). So D5 changes nothing in this repository's
  own ledger today. Its red `delivery-enforce` run is caused by malformed v1 front matter, a parse
  failure that D5 does not touch.
- **The store has no lookup by event id.** `ReadEvents(Filter{Source, Kind, …})` is the only read.

## 2. Purpose, boundary, entry conditions

M3 binds a tenant commit's `Intent: sdd:<name>` to the exact `intent.sdd` change-intent revision at
that commit, so the unchanged readiness gate (O-8) can judge it, and it makes sure no verdict about
such a commit can be ledgered ahead of what it cites (D5).
- **Changed:**
  - git connector trailer handling;
  - the CLI resolver;
  - projection validation (nothing else in projections);
  - the reviews connector's append condition (D5);
  - the CLI's `sync reviews` / `sync all` output.
- **Unchanged:**
  - the `sdd` provider (byte-identical production code);
  - the readiness gate's logic;
  - every projection reducer;
  - this repository's trailer gate script;
  - `records` / `specdir` claim resolution, except D2's all-or-nothing on a commit that also carries
    an `sdd` value (their ordering hazard is `DEF-INTENT-GIT-ORDERING-001`).

Entry conditions (sdd SPEC §14): token → D1 (a VD lands with M3); anti-spoofing → §4.2; per-commit
handling → §4.1, §4.4; sites → §5; readiness gate → §4.5, M3-INV-8.

## 3. Owner decisions (all decided 2026-11-19: the recommended option)

- **D1 — token.** `Intent: sdd:<name>`, the whole value matching `^sdd:[a-z0-9][a-z0-9-]{1,99}$`.
- **D2 — an unbindable `sdd` claim.** Per commit, all-or-nothing: the commit's event carries no refs at
  all; the claim is reported as a `ClaimSkip` (not ledgered); the sync succeeds.
- **D3' — ordering of claims.** The ledger check (§4.3).
- **D4 — the pre-existing records/specdir claim hazard.** Deferred, filed as
  `DEF-INTENT-GIT-ORDERING-001`; fixed in its own change after M3.
- **D5 — ordering of verdicts and reviews.** The reviews connector holds a v2 verdict or requirement
  review until its order-only prerequisites are ledgered (§4.6). Rejected alternatives: a
  dangling-tolerant verdict projection (it hides real errors); keeping the hard failure (round 4
  showed correct data reaches it).

## 4. Contract

### 4.1 Trailer values

- `records:<CI-id>`, `specdir:<CI-id>`, and bare `<CI-id>` resolve as today, including whole-sync
  failure on a malformed value. The one change is D2: on a commit that also carries an unbindable
  `sdd` claim, their refs are dropped too (INV-29 amended, §5).
- **Any value whose first four bytes ASCII-case-fold to `sdd:`** is an `sdd` claim:
  - if the **whole value** matches D1's grammar (lower-case prefix included), it is resolved
    (§4.2-§4.4);
  - otherwise it is a claim skip with cause `malformed`, never a sync failure. `SDD:add-auth` is
    `malformed`.

### 4.2 Resolution and anti-spoofing (CLI resolver, `sdd` arm)

Unchanged from draft 4 §4.2: lazy construction on the first `sdd` claim; `Resolve(ctx, commit, name)`;
a revision is accepted only if its source is `intent.sdd`, its kind is `change_intent.recorded`, its
native id is CI-shaped, its digest is 64 hex, and the last segment of its artifact path is `name` or
`YYYY-MM-DD-name`. Any failure is a spoof, and the sync fails. The segment check is defence in depth;
the guarantee is `Resolve`'s exact-name semantics (SDD-INV-59). The connector re-checks `Provider ==
"sdd"`, a CI-shaped `RecordID` and a 64-hex digest.

### 4.3 The ledger check for claims (D3')

Unchanged from draft 4 §4.3. Once per `sync git`, lazily on the first `sdd` claim, the CLI reads the
set `L` of `(native id, artifact_sha256)` of every ledgered `intent.sdd` `change_intent.recorded`
event (the digest decoded from the payload). An accepted revision binds only if it is in `L`;
otherwise the claim is skipped as `not-ledgered`. **Recovery:** `sync intent sdd`, then `sync git`.
`L` is a safe snapshot: `sync git` appends only git events, and the ledger is append-only.

### 4.4 Claim outcomes and the skip channel

Unchanged from draft 4 §4.4:

| Outcome | Cause |
|---|---|
| value fails D1's grammar | `malformed` |
| `ErrUnknownChange` | `unknown-change` |
| `ErrArchivedEarlier` | `archived-earlier` |
| `ErrNoChangeIntent` | `no-change-intent` |
| `*UnmappableError` | `unmappable:<cause>` (incl. `slug-too-short`) |
| accepted, not in `L` | `not-ledgered` |
| reader I/O, `ErrShallow`, cancellation, spoof | sync fails |

Types: `connectorgit.ErrClaimUnbound`, `ClaimUnboundError{Cause}` (which `errors.Is`-matches it),
`ClaimSkip{Commit, Value, Cause}`, `Result.ClaimSkips`. On any skip, the connector drops **all** of
that commit's refs (D2). `sync git` and `sync all` print `claim-skip <cause> <value> at <commit>` after
their counts line, the value through `printableUnit`. `verify` prints none and discards both of its git
syncs' skips. No skip changes an exit status or an event.

### 4.5 Projections and the readiness gate

- `commitPayload.validate` accepts provider `sdd`. This widens validation, so it is retroactively safe.
- `commit_intents_view` stays insert-only. A binding, once projected, is never withdrawn.
- The readiness gate is unchanged. A commit with no projected claim rows is not warranted.

### 4.6 The reviews hold (D5)

- **New dependency.** `reviews.Deps` gains `Ledger LedgerView`, **required** (a nil or typed-nil value
  fails `New`, as the other dependencies do, R-INV-6):

  ```go
  type LedgerView interface {
      // CommitClaimed reports whether a ledgered git commit event for sha carries exactly this ref.
      CommitClaimed(ctx context.Context, sha, provider, recordID, digest string) (bool, error)
      // RequirementLedgered reports whether a requirement.recorded event of this source, id and digest is ledgered.
      RequirementLedgered(ctx context.Context, source core.Source, recordID, digest string) (bool, error)
      // EventLedgered reports whether an event with this id is ledgered.
      EventLedgered(ctx context.Context, eventID string) (bool, error)
  }
  ```
- **The hold rule.** After an artifact parses (the existing parse rules and their failures are
  unchanged), and before it is appended:
  - a `proofbound.obligation-verdict.v2` is appended only if `CommitClaimed(reviewed_commit,
    trimPrefix(change_intent.source, "intent."), change_intent.record_id, change_intent.artifact_sha256)`,
    `RequirementLedgered` for every listed requirement, and `EventLedgered` for every evidence id all
    hold;
  - a `proofbound.requirement-review.v1` is appended only if `RequirementLedgered` holds for its
    requirement;
  - a `vera.verdict.v1` is never held.

  Otherwise the artifact is **held**: not appended, not counted as malformed, and the sync continues.
  The checks run in the order listed, and the first one that fails names the cause.
- **Causes:** `claim-not-ledgered`, `requirement-not-ledgered`, `evidence-not-ledgered`.
- **Result.** `reviews.Result` gains `Held []Hold` with `Hold{Path, Cause string}`. A held path is not
  in the cursor. `sync reviews` and `sync all` print `held <cause> <path>` after their counts line,
  the path through `printableUnit`. A hold never changes an exit status. `verify` prints none (as for
  claim skips).
- **The CLI's `LedgerView`.** It is built over the same store, lazily, on the first v2 verdict or
  requirement review. One `ReadEvents` pass with an empty filter fills three sets:
  - every event id;
  - every `(commit sha, provider, record id, digest)` ref carried by a `git` `commit.recorded` event,
    decoded with the payload type that projections validate;
  - every `(source, requirement id, digest)` of an `intent.*` `requirement.recorded` event, with the
    digest decoded from the payload.

  The pass costs O(ledger) once per `sync reviews` that sees a v2 artifact. A sync that sees none
  never reads the ledger.
- **Why the hold is sufficient.** A held check is made against the ledger as it stands before the
  append. The ledger is append-only, so everything the check found has a lower sequence number than
  the artifact's own event. When the reducer reaches that event:
  - the claim row exists, because the git event that carries it was projected first, and it passes
    the dangling check (M3-INV-5 for `sdd`; for `records` / `specdir` that is
    `DEF-INTENT-GIT-ORDERING-001`'s scope);
  - `commits_view` holds the sha, because the same event put it there;
  - the requirement rows exist (`reduceIntent` projects every ledgered requirement revision);
  - every evidence event precedes the verdict.
- **Why the hold is fail-closed on trust.** A held verdict is not in the ledger, so no projection
  sees it, and a commit it would warrant stays unwarranted. Holding never makes a commit pass.
- **What the hold does not cover.** A verdict whose **content** is wrong (coverage, outcome values,
  status) is appended, as today, and its reducer still fails. That is a malformed tenant artifact.
  The parse rules catch what they can, and the rest is unchanged by M3.
- **The accepted cost.** A verdict that cites a claim that will never be ledgered (a typo in a
  digest, say) used to make Apply fail loudly. Now it is a `held` line printed on every sync, and the
  commit stays unwarranted. Trust stays fail-closed; only the diagnosis moves from an Apply error to
  a printed line.

## 5. Sites (complete)

| Site | Change |
|---|---|
| `connector/git/connector.go` `resolvedCommit`, allowlist, spoof check, `Result` | §4.1, §4.2, §4.4 |
| `connector/git/SPEC.md` §2 (`Result`, `IntentResolver` contract), INV-29, new INV rows | `sdd` claims fail per commit; INV-29 amended for D2; **the accepted cross-version limitation and its revisit trigger (§9) are stated beside the new INV rows** (round-4 F4) |
| `connector/reviews/reviews.go` `Deps`, `New`, `Sync`, `Result` | §4.6 |
| `connector/reviews/SPEC.md` | the hold rule as a new invariant R-INV-13; R-INV-5's "stops on malformed input" unchanged (held ≠ malformed); R-INV-6 lists `Ledger`; Interface section lists `LedgerView`, `Hold`, `Result.Held` |
| `connector/reviews/reviews_test.go` | existing `Deps` literals gain a `LedgerView` (a fake that admits everything keeps their behaviour) |
| shared trailer fixture `scripts/tests/fixtures/intent-trailers/cases.tsv` | `sdd:` rows (valid, malformed, upper-case), all gate `fail` |
| `connector/git/gitcmd/gitcmd_test.go::TestIntentTrailers_SharedCorpusKernelBinds` | "binds" = a ref is bound. `corpusResolver` models this repository: provider `sdd` → `ClaimUnboundError{unknown-change}`; its comment amended |
| `cli/cli.go` `commitIntentResolver`, `syncGit`, `syncReviewsOnStore`, the claim-skip and held output | §4.2-§4.4, §4.6 |
| `cli/SPEC.md` CLI-INV-7 / CLI-INV-8 | the `sdd` arm, the claim-skip lines, the held lines |
| `projections/projections.go` validate (:727) | accept provider `sdd` |
| `scripts/intent-delivery-applicability.sh` | **unchanged** |
| `kernel/internal/connector/intent/sdd/SPEC.md` §12, SDD-INV-56 text | record M3 as landed; SDD-INV-56 unchanged in substance |
| `kernel/internal/connector/intent/sdd/SPEC.md` §14 | **a new M2 entry condition**: M2 states whether bindings made under M1's rules stay, and ships a migration if not (§9; round-4 F4) |
| generated pins `docs/verification/p6-event-route-matrix.md`, `p6-connector-reality.md`, `p6-task2-close` | regenerate after the code commits |
| `docs/decisions/VD-…` (D1-D5) | new; lands with the M3 change intent (committing it regenerates the governed `INDEX.md`) |

## 6. Invariants (M3-INV-n)

1. **M3-INV-1** — A well-formed `sdd:<name>` at `c`, accepted and ledgered, binds exactly
   `Resolve(c, name)`'s revision.
2. **M3-INV-2** — A resolver revision failing any §4.2 check fails the sync.
3. **M3-INV-3** — Every tenant outcome in §4.4's table yields one `ClaimSkip`, and the commit event
   carries **no** refs. Other commits are unaffected, and the sync succeeds.
4. **M3-INV-4** — Reader errors and cancellation fail the sync.
5. **M3-INV-5** (restated) — No ledgered commit event names an `sdd` revision that is not itself
   ledgered at a lower sequence number. A projection rebuild succeeds after **any** interleaving of
   `sync git`, `sync intent sdd`, `sync reviews` and `sync all`, on any tenant whose committed review
   artifacts are well-formed in content.
6. **M3-INV-6** — A history with no `sdd` value neither constructs the `sdd` provider nor reads `L`
   during `sync git`.
7. **M3-INV-7** — `records` / `specdir` claim behaviour and the pinned v2 payload vector are unchanged
   for commits without `sdd` values.
8. **M3-INV-8** — On one ledger, the readiness gate passes a commit bound to an `sdd` change intent
   whose targets are SATISFIED with a non-owner VERIFIABLE review, and refuses a commit whose only
   claim was skipped, and a commit with one skipped and one bindable claim even when a SATISFIED
   verdict for the bindable claim is committed. (Under M3-INV-10's cross-version exception, an earlier
   kernel's binding may still warrant a commit; §9.)
9. **M3-INV-9** — This repository's trailer gate still refuses every `sdd:` value, and everything it
   refused before.
10. **M3-INV-10** (restated) — Bindings are monotone: a `commit_intents_view` row, once projected,
    survives every later event for its commit. So an obligation verdict that projected once keeps
    projecting on every rebuild.
11. **M3-INV-11** (new, D5) — No v2 verdict or requirement review is ledgered before its order-only
    prerequisites (§4.6). One whose prerequisites are missing is held, reported, and appended by the
    first later `sync reviews` that finds them ledgered. A hold never fails a sync or changes an exit
    status.

## 7. Test derivations

| INV | Test (where) | Kills | Controls |
|---|---|---|---|
| 1 | git: fake resolver binds `sdd:add-auth` (unit); cli: real sdd tenant + embedded store via `cli.Run` | binding the trailer value instead of the revision id | positive: a `records` claim still binds |
| 2 | git unit, resolver returns a wrong-segment / wrong-kind / wrong-source revision | dropping any one §4.2 clause (one case per clause) | neutral: a correct revision binds |
| 3 | git unit, one case per §4.4 row (incl. `SDD:add-auth` → `malformed` beside `sdd:add-auth` → resolved), two-claim commits | whole-sync failure; partial binding; case-folding D1's grammar | a clean commit beside the skipped one is appended with its ref |
| 4 | git unit: resolver returns a plain error | reading every error as a skip | — |
| 5 | cli package, embedded store, untagged (§7.2) | skipping the ledger check; disabling the hold | the bound commit's row and the verdict both project after recovery |
| 6 | cli: injected counters for construction and for the `L` read, over a history with no `sdd` value | eager construction; eager ledger read | each counter is 1 once a single `sdd:` value is added |
| 7 | existing git INV-28..30 tests and §5.1 vector, unchanged | any drift in v2 bytes | — |
| 8 | cli package, embedded store, untagged (§7.1) | the gate ignoring `sdd` rows (X1 must PASS); keeping the bindable ref of a two-claim commit (X3 must stay BLOCKED: its verdict would then append and project); disabling the hold (X3's verdict is appended, Apply fails, so X1 is BLOCKED) | X1 `state=PASS`, X2 and X3 `state=BLOCKED`, all on one ledger; the X3 verdict reported `held claim-not-ledgered` |
| 9 | `scripts/tests` and the kernel oracle over `cases.tsv` | widening the gate to `sdd:`; the oracle counting sync success as binding | old rows unchanged |
| 10 | untagged test that opens the embedded store directly: a change intent, a git event binding it, a second git event for the same commit **without** the ref (what a later kernel version could append), a v2 verdict citing the binding; then Apply and Rebuild | deleting a commit's rows on a newer commit event | both succeed and the row is present |
| 11 | reviews unit with a fake `LedgerView`: one case per prerequisite missing (claim, each requirement, each evidence id), each held with its cause and absent from the appender and the cursor; requirement review with its requirement missing; v1 never consults the view; the same artifact appended on a second sync once the fake admits it | dropping any one prerequisite check; counting a hold as malformed; failing the sync on a hold; holding a v1 verdict | neutral: an artifact whose prerequisites are all present appends exactly as before |

Every new test is calibrated before it is believed: a hand mutant per "kills" cell must make it fail,
and a neutral edit must leave it passing.

### 7.1 Row 8 procedure

The test is an internal test of package `cli` (no build tag), so it can call `runCommand` and also
open the same store with `openStore(ctx, root, "")` between commands to read event ids.
1. **Tenant.** In a temp dir, `git init` a repository with an sdd profile (OpenSpec, owner `owner-a`)
   and an active change `add-auth` with one requirement. Commit `X1` with `Intent: sdd:add-auth`.
   Commit `X2` with `Intent: sdd:nope` (no such change). Commit `X3` with both `Intent: sdd:add-auth`
   and `Intent: sdd:nope`. `X3` is HEAD.
2. **Ledger, part 1.** `sync intent sdd`, then `sync git`. Expect `claim-skip unknown-change sdd:nope`
   at X2 and at X3; X1 bound.
3. **Read back.** Open the store; read the `intent.sdd` change-intent and requirement events for
   `add-auth` (record ids and payload digests) and X1's and X3's `git` event ids.
4. **Artifacts.** Write, each with its self-digest (`ArtifactSHA256`'s zeroing rule):
   - a `proofbound.requirement-review.v1`, VERIFIABLE, reviewer `reviewer-b` (≠ `owner-a`);
   - a `proofbound.obligation-verdict.v2` for X1, SATISFIED, evidence = X1's git event id;
   - the same for X3, evidence = X3's git event id.

   Commit them under `docs/verification/verdicts/` as `X4` (no trailer). X4 is HEAD.
5. **Gates.** Write `gates/x1.yaml`, `gates/x2.yaml` and `gates/x3.yaml` in the working tree. Each is a
   **JSON** `proofbound.gate.v1` definition with a distinct `id`, rule `intent-delivery-readiness`,
   mode `canary`, and `scope_commit` = that commit's full sha.
6. **Ledger, part 2.** `sync git` (X4, no claims), then `sync reviews`. Expect the review and X1's
   verdict appended, and `held claim-not-ledgered docs/verification/verdicts/<x3 file>`.
7. **Judge.** `gates canary`. Assert `state=PASS` for x1 and `state=BLOCKED` for x2 and x3.

BLOCKED is meaningful only beside the PASS on the same ledger: a failed Apply makes every readiness
gate BLOCKED (`gates.go` `evaluateIntentRule`), so the x1 PASS is what proves the ledger projects.

### 7.2 Row 5 procedure

Same package and harness. The tenant is §7.1's after step 4 (X1 claims `sdd:add-auth`, X4 holds X1's
verdict and review), on a **fresh** state root:
1. `sync all`, which never syncs `sdd` (SDD-INV-56). Expect X1 `claim-skip not-ledgered`, the verdict
   `held claim-not-ledgered`, and the review `held requirement-not-ledgered`. Rebuild succeeds.
2. `sync intent sdd`, then `sync git` (X1 now bound), then `sync reviews` (both appended). Rebuild
   succeeds; `gates canary` on x1 reads `state=PASS`.
3. On another fresh root: `sync intent sdd` → `sync git` → rebuild; the bound row exists.

This is round 4's F2 scenario; without the hold, step 1 appends the verdict ahead of the binding and
step 2's rebuild fails.

## 8. Proposed change intent (owner approves the wording)

`CI-sdd-warranty-m3-<suffix>`: implements `BR-sdd-warranty-i9miv7` O-8, and O-10 for claims, at the
§5 sites, including the reviews hold (D5). Non-goals as in §9.

## 9. Non-goals

- Any change to the `sdd` provider; M2 (Spec Kit); M4 (drift).
- The records/specdir claim-ordering hazard: `DEF-INTENT-GIT-ORDERING-001` (D4).
- Ledgering claim skips or holds.
- Repairing a review artifact whose content is wrong (§4.6).
- Withdrawing a binding when a later kernel version would no longer bind it. Bindings are monotone
  (M3-INV-10). **Accepted limitation of M3.** Its home after M3 lands is the git SPEC (beside the new
  INV rows) and the sdd SPEC §14's M2 entry condition (§5). **Revisit trigger:** any change that narrows
  `sdd` binding (M2's Spec Kit rules, a `Resolve` outcome change, a stricter §4.2 check) states in its
  own SPEC whether existing bindings stay, and ships a migration if they do not. Within one kernel
  version it cannot arise: `L` only grows and `Resolve` is deterministic.
