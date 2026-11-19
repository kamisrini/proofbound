---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round5-20261119
status: NEEDS_WORK
reviewed_commit: 7f29f853d3e4019210e25dc9725d5d31bbc888f9
findings:
  - finding_id: r2-m3-spec-review-round5-20261119-F1
    severity: MED
  - finding_id: r2-m3-spec-review-round5-20261119-F2
    severity: MED
  - finding_id: r2-m3-spec-review-round5-20261119-F3
    severity: MED
  - finding_id: r2-m3-spec-review-round5-20261119-F4
    severity: LOW
  - finding_id: r2-m3-spec-review-round5-20261119-F5
    severity: LOW
  - finding_id: r2-m3-spec-review-round5-20261119-F6
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round5-20261119.md
artifact_sha: 2b47de8fc29ecf6b98c10ef660d6a476b4c8f339ae8adb22676daf27253a15c6
---

# R2 M3 git-binding SPEC draft 5: non-author review, round 5

**Committing session's note (mechanical, not the reviewer's words):** the reviewer returned the
`findings` list in flow style (`- {finding_id: …, severity: …}`), as the committing session had asked.
The reviewer pointed out that the reviews connector's v1 parser rejects that form, so it is committed
in block style with the same ids and severities. No other change was made. Absolute paths were already
redacted by the reviewer to `<repo>` and `<scratch>`.

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**: it did not write drafts 1-5, the round 1-4 verdicts, the sdd SPEC,
  or any code it read. It has no authoring context from the session that wrote them (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft5.md` at
  `7f29f853d3e4019210e25dc9725d5d31bbc888f9`.
- **HEAD moved during the review**, to `a72a8858fef2caf03427136bf7d9a2e98dc377a3`.
  `git diff --stat 7f29f85 a72a885` touches only `notes/state.md`. The draft and all code are
  identical, so the findings hold at both commits.
- **Read at HEAD:**
  - `projections/projections.go`: `Apply`, `Rebuild`, `reduce` (the commit route and its dangling
    check), `reduceIntent`, `requireExactRecord`, `reduceReview`, `reduceObligationVerdict`,
    `reduceRequirementReview`, `commitPayload` and `validate`;
  - `projections/SPEC.md` §3;
  - `connector/reviews/reviews.go` (`Deps`, `New`, `Sync`, the cursor, `ArtifactSHA`), `v2.go`
    (`ParseObligationVerdict`, `ParseRequirementReview`, `decodeExactOrdered`, `validateReference`,
    `ArtifactSHA256`), `connector/reviews/SPEC.md`, and the `reviews_test.go` `Deps` literals;
  - `cli/cli.go`: `openStore`, `stateRoot`, `newIDs`, `syncGit`, `commitIntentResolver`,
    `committedVerdictReader`, `syncReviewsOnStore`, `runCommand` (the gates branch, the `sync git` /
    `sync all` order), `resolveIntentGateHeadScope`, `verify`, `requireNoSecondSyncAppends`;
  - `cli/SPEC.md` CLI-INV-7/8;
  - `gates/gates.go`: `Definition`, `LoadDir`, `Parse`, `validate`, `evaluateIntentRule`,
    `latestIntentProof`;
  - `store/store.go` (`Filter`, `ReadEvents`, the append idempotency key, `importRecords`) and
    `store/migrations/001_ledger.sql`;
  - `core/core.go` `ParseEventID`;
  - `connector/intent/intent.go` `Reference`;
  - `connector/intent/sdd/{build.go, ids.go}` (ID and obligation-ID derivation, the target relation);
  - `sdd/SPEC.md` §14;
  - `docs/deferrals.tsv` (the `DEF-INTENT-GIT-ORDERING-001` row);
  - `scripts/delivery-enforce.sh`, `scripts/p6-task2-close.sh`, `scripts/p6-task5-review.sh`,
    `docs/releases/r1/seed-demo-app.sh`;
  - `docs/verification/p6-historical-evidence.jsonl`;
  - every committed file under `docs/verification/verdicts/`.
- **Executed** (with `$HOME/go/bin` first on PATH and GOCACHE under `<scratch>`):
  - `go test ./internal/connector/reviews/ ./internal/gates/ -count=1`: both pass.
  - A scratch test, added through `go test -overlay` and not written into the repository, that runs
    the reviews connector's own `wireCandidate`, `artifactSchema` and the three parsers over every
    file `git ls-tree HEAD -- docs/verification/verdicts` lists, in sorted order. Result:
    - `p5-obligation-verdict-round1.md` (v2), `p5-requirement-review-round1.md` and
      `p7-replay-requirement-review-round1-20261014.md` all **parse**;
    - the first parse failures are `r2-m3-spec-review-round{1..4}-20261118.md`
      (`unknown finding field "{finding_id"`), which sort after them.
  - The overlay run used `GOFLAGS=-mod=mod`, which rewrote `kernel/go.mod`. The reviewer restored it
    with `git checkout -- kernel/go.mod`. The final `git status` shows only the two permanent
    untracked paths.

## Round 4, closure

| Round-4 finding | Verdict | Evidence |
|---|---|---|
| F1 MED: row 8 cannot yield PASS | **Closed, with LOW residue (F4)** | See the §7.1 trace below the table |
| F2 MED: verdict-before-binding wedge | **Partly closed** | The hold closes the claim and requirement prerequisites correctly. But evidence ids are not order-only (F1), the hold opens a fail-open path on in-place verdict revision (F2), and the real `LedgerView` has no derivation outside `sdd` (F3) |
| F3 LOW: two-claim kill cannot die through the gate | **Closed** | Under correct code X3 has no claim row, so its verdict is held and x3 reads BLOCKED. Under partial binding the claim is ledgered, the verdict appends and x3 reads PASS. With the hold disabled, Apply fails and x1 reads BLOCKED. Provided the X1 and X3 verdicts have distinct `verdict_id`s (F4) |
| F4 LOW: revisit trigger has no home | **Closed (as planned sites)** | §5 carries it into the git SPEC and a new sdd SPEC §14 M2 entry condition. §14 has no M2 list today, so it is a new item, as stated |

**How the §7.1 procedure was traced.** Taken step by step against the code:
- **The verdict cannot be ingested before the commit it reviews.** The order (`sync intent sdd` →
  `sync git`, commit the artifacts as X4, `sync git` → `sync reviews`) puts the verdict at a higher
  sequence number than X1's commit event.
- **Gate definitions are read from the working tree.** `runCommand` calls
  `gates.LoadDir(<root>/gates)`, so uncommitted files work. JSON is what `Parse` requires. An
  exact-sha `scope_commit` passes `validate`, and `resolveIntentGateHeadScope` leaves it alone.
- **Canary output matches the assertions.** The line is `gate=… state=PASS|BLOCKED …`.
- **One requirement review is enough for both X1 and X3.** The readiness query's `reviewed` test
  joins `requirement_reviews_view` on the requirement revision and obligation only, never on the
  commit.
- **The gate requires a review.** `reviewed == targets` is part of PASS.

## Answers to the attacks

### 1. Can §7.1 author valid artifacts from what step 3 reads back?

Not from the list as written, but it can from the payloads it reads.
- **What the v2 verdict needs.** Its front matter is one compact JSON line with exactly 11 keys in
  order. That includes `declared_reviewer`, which must not be blank, and `findings`, where `[]` is
  accepted. There is no field linking the verdict to a requirement review.
- **What step 3 leaves out:**
  - **Obligation ids.** sdd derives them through `obligationID`, and the review must cover every one.
  - **Relations.** They must be `evaluates` for the verdict and `reviews` for the requirement review,
    but sdd change-intent targets carry `implements`.
- **Distinct `verdict_id`s matter** (F4).

### 2. D5 soundness

- **The prerequisite list is complete against both reducers:**
  - `reduceObligationVerdict` looks up `commits_view`, `commit_intents_view`, `requirements_view`,
    `intent_targets_view` and `events`.
  - `reduceRequirementReview` looks up `requirements_view` and `requirement_obligations_view`. The
    second is filled in the same `reduceIntent` call as the first.
  - No other reducer reads review prerequisites.
- **"Item 4 is present whenever item 2 is" is true.** The commit route inserts a claim row only after
  `change_intents_view` holds the revision, and `intent_targets_view` is written in the same reduce.
- **The sequence argument is sound for the claim and requirement checks.** The snapshot precedes the
  append, the ledger is append-only, and the views are insert-only or monotone.
- **It is not sound for evidence** (F1).
- **The interactions asked about are all safe:**
  - **Idempotency:** the key is `(source, native_id, content_sha)`, so a late append is an ordinary
    first append.
  - **Cursor:** it is only recorded, so leaving a held path out of it is harmless.
  - **`sync_runs`:** a hold is not an error, so the run finishes normally.
  - **R-INV-4/5:** they are not violated.
  - **`verify`:** its second pass sees the same ledger, because a reviews event's id cannot be cited
    before it exists. So a hold in pass 1 stays a hold in pass 2, and the second-sync-appends check
    holds.
  - **`sync all` order:** reviews run after git, so the lazy view sees that run's git events.
- **A held-then-appended verdict can still fail Apply** only on content, or when the git event
  carrying the claim itself dangles (the records/specdir deferral).

### 3. M3-INV-5

- It is true for `sdd` claims under the listed commands.
- As written ("any tenant") it is false for tenants with `records` / `specdir` claims under
  `sync git` before `sync intent`, which is `DEF-INTENT-GIT-ORDERING-001`. Its derivation also does
  not run green (F1, F5).

### 4. M3-INV-8's two-claim kill

It dies through the gate on one ledger, with the X1 PASS intact (closure table).

### 5. The `LedgerView` sketch

- **Feasible.** `ReadEvents(Filter{})` with `Limit` 0 pages every event in `seq` order, 256 per page.
  `latestIntentProof` already does the same scan.
- **Digests come from decoded payloads.** `commitPayload` is unexported in `projections` (F6).
- **Event-id comparison semantics are unstated.** `ulid.ParseStrict` accepts lower case, but
  `events.event_id` is `TEXT` holding canonical upper case (F5).

### 6. Sites

- `reviews.New` / `Deps` are constructed only in `cli.go:455` and in `reviews_test.go` (lines 69,
  607, 619, 639).
- `connectorgit.Result` is used only in `cli.go` (`syncGit`, `verify`).
- The site list misses reviews SPEC text that the hold falsifies (F6).

### 7. Law walk

See below.

## Findings

**F1 (MED): §7.2, the derivation for M3-INV-5, cannot run green on a correct implementation, because
evidence event ids are local to one ledger and not "order-only".**
- **The construction.** §7.2 reuses "§7.1's tenant after step 4", including X4's committed verdict,
  "on a fresh state root".
- **Why the id is never found.** That verdict's evidence is X1's git event id as minted in §7.1's
  ledger. `newIDs` uses `crypto/rand`, so X1's event on the fresh ledger gets a different ULID. No sync
  can ever append an event with the cited id.
- **What happens in step 2.** After `sync intent sdd` and `sync git`, the claim and the requirement
  are ledgered, but `EventLedgered` is false. The verdict is reported `held evidence-not-ledgered` and
  is not appended, so `gates canary` on x1 reads BLOCKED. Step 2 asserts "both appended" and
  `state=PASS`.
- **It cannot be fixed by re-authoring.** The verdict must already be committed before step 1 for
  step 1 to exercise the hold. At that moment no event exists on the fresh ledger for it to cite.
- **The root cause is in §1.** It says items 1, 2, 3 and 5 "can be absent only because of order: a
  later sync can supply them". For item 5 this is false on every ledger other than the one the
  evidence was minted on. The only exception is ids that `migrate historical-evidence` imports
  verbatim.
- **This is the round-4 F1 class again.** A prescribed positive control is red on correct code, so
  the derivation is not executable (Law 6).

**F2 (MED): the hold can keep a commit warranted on a verdict revision that the committed artifact has
withdrawn, which contradicts §4.6's claim that "trust stays fail-closed".**
- **Why the old revision survives.** `obligation_verdicts_view` is upserted on
  `(verdict_id, requirement…, obligation_id)`. Only a later revision's event overwrites `outcome`.
- **How it happens.** Take rev1 of a verdict, SATISFIED, appended and projected. The reviewer revises
  the same `verdict_id` in place to NOT_SATISFIED / NEEDS_WORK, citing an evidence id that is not on
  this ledger. The hold then:
  - reports rev2 `held evidence-not-ledgered` and does not append it;
  - leaves rev1's SATISFIED row in place, so `evaluateIntentRule` keeps returning PASS;
  - does not stop `delivery-enforce.sh`, which runs `sync all` (printing the held line) and then
    `gates enforce` (passing).
- **Before M3**, the same ledger would append rev2. Apply then fails or overwrites the outcome, and
  either way the commit reads BLOCKED.
- **It is reachable from correct data.** This repository's pattern does it:
  - `p5-obligation-verdict-round1.md` cites ids from `docs/verification/p6-historical-evidence.jsonl`,
    which every ledger can import with the same ids.
  - A revision that adds a locally minted id is held on every other ledger, permanently.
- **How the draft misses it.** §4.6's argument ("a held verdict is not in the ledger, so … a commit it
  would warrant stays unwarranted") covers only verdicts that would *add* a warrant, not revisions
  that would *remove* one. No §7 row covers revision.

**F3 (MED): the CLI's real `LedgerView` has no derivation outside `sdd`, and the premise used to
dismiss D5's effect on this repository is false.**
- **The premise is wrong.** §1 states "This repository holds no v2 verdict and no requirement review
  (… 0 files). So D5 changes nothing in this repository's own ledger today." At HEAD there are three
  committed wire artifacts of those schemas:
  - `p5-obligation-verdict-round1.md` (`intent.records`, ACCEPTABLE, historical-archive evidence);
  - `p5-requirement-review-round1.md`;
  - `p7-replay-requirement-review-round1-20261014.md`.
- **The hold runs on every sync of this repository.** The overlay scan shows all three parse and sort
  before the first malformed file, so every `sync reviews`, `sync all` and `verify` here would
  consult the hold (and pay the O(ledger) read).
- **Live evidence depends on it.** `p6-connector-reality.md` C4-008 classifies `reviews` as live on
  the strength of that verdict.
- **No test exercises the real view outside `sdd`:**
  - Rows 5 and 8 use only `sdd` claims and `intent.sdd` requirements.
  - Row 11 uses a fake view.
  - The only tests that drive a records v2 verdict through projections are `//go:build integration`,
    which `make check` does not run.
- **What would slip through.** Each of these mutants passes every §7 row, yet holds this repository's
  own live verdict and every `records` / `specdir` tenant's verdict:
  - a view that collects only `sdd` refs;
  - a view that collects only `intent.sdd` requirements;
  - a view that reads events with a source filter, and so misses historically imported evidence.
- **The claim it leaves unproven.** §0c says the hold "closes the class for every provider, `records`
  included" (Law 6; Proofbound's first tenant is Proofbound).

**F4 (LOW): §7.1 under-specifies the artifacts in ways that can turn a kill cell silent or a control
red.**
- **Missing from step 3's read-back:**
  - obligation ids;
  - the requirement revision's id (sdd's `requirementID` embeds the digest);
  - the required relations (`evaluates` / `reviews`, against the `implements` that sdd change-intent
    targets carry).
- **"The same for X3" does not require a distinct `verdict_id`.** The `ON CONFLICT … DO UPDATE` on
  `obligation_verdicts_view` does not update `commit_sha`. So if the X3 verdict reuses X1's id, then
  under the partial-binding mutant X3's row keeps `commit_sha = X1`, x3 stays BLOCKED, and the mutant
  survives.
- **One review covers X3 under that mutant only if X2 and X3 leave `add-auth`'s files unchanged.** The
  draft does not say they must.
- **Required gate fields are not listed.** `description` and `expires` are required by `validate` but
  not in step 5.
- **Why LOW.** The calibration rule would expose most of these before the test is believed.

**F5 (LOW): M3-INV-5 claims more than the design delivers, and `EventLedgered`'s comparison semantics
are unstated.**
- **It overreaches on providers.** "Any tenant" includes `records` / `specdir` claims, where
  `sync git` before `sync intent records` still makes rebuild fail ("dangling intent revision"). The
  carve-out for `DEF-INTENT-GIT-ORDERING-001` appears in §4.6's prose but not in the invariant.
- **"Well-formed in content" is undefined.** Several artifacts pass the parser yet make every rebuild
  fail:
  - a reviewer equal to the owner;
  - a lower-case evidence ULID, which `ulid.ParseStrict` accepts but the reducer compares as `TEXT`
    against canonical upper case.
- **`EventLedgered`'s semantics decide between two outcomes:**
  - If the view normalises ids through `ParseEventID`, a lower-case id passes the hold, appends, and
    wedges the projection.
  - If it compares exact strings, the verdict is held forever, with a cause that misleads.
- **The derivation covers one interleaving**, not "any".

**F6 (LOW): the hold falsifies reviews SPEC text that §5 does not list, and the sketch implies a second
decoder.**
- **SPEC text the hold makes false:**
  - The reviews SPEC Purpose says "each valid schema-bearing artifact emits one … event".
  - R-INV-11 says "dangling or future evidence fails projection".
  - The requirement-review contract says "projection validation rejects a dangling or wrong revision
    digest".
  - Through the connector, each of these becomes a hold. §5 lists only R-INV-13, R-INV-5, R-INV-6 and
    the Interface.
- **The sketch implies a second decoder.** It decodes git refs "with the payload type that
  projections validate", but `commitPayload` is unexported, and §2 says projections change only
  validation. So either a second decoder of the git payload exists (Law 2) or there is an unlisted
  export.
- **One claim has no test row.** §4.6 says "a sync that sees none never reads the ledger", but row 6
  covers only `L`.

## Other observations (no finding)

- **Held lines for already-ledgered artifacts.** An artifact that an earlier kernel already appended,
  whose prerequisites are absent, will now print `held …` rather than counting as `existing`. That is
  misleading but harmless.
- **The repository's red `delivery-enforce`.** The malformed v1 files are this review series' own
  verdicts (rounds 1-4, flow-style `{finding_id: …}` findings, which `parseFindings` rejects). A
  round-5 verdict committed in the same form will add a fifth.
- **The accepted cost is understated.** On any ledger other than the authoring one, every SATISFIED v2
  verdict whose evidence was not imported historically is held permanently. Before M3 it wedged the
  projection, so that is still an improvement, but the "accepted cost" paragraph describes only digest
  typos.

## Law walk

- **Law 1.** The `@generated` pins are listed for regeneration. Compliant.
- **Law 2.**
  - D5's rule lives in the reviews SPEC (R-INV-13), but conflicting statements in the same SPEC are not
    amended (F6).
  - A second git-payload decoder is implied (F6).
- **Law 6.**
  - M3-INV-5's derivation is not executable (F1).
  - The real `LedgerView` has no derivation outside `sdd` (F3).
  - §4.6's trust claim is refuted (F2).
  - The SPEC precedes code, and the calibration rule is stated.
- **Law 7.** The ordering half of the wedge lesson is now compiled for claims and requirements.
  Evidence ordering and revisions are not (F1, F2).
- **Law 8.** No new dependency. Compliant.
- **Law 9.** This is a non-author re-check. The calibration rule would catch F4, but not F1's red
  control before code exists.
- **Law 10.** Not engaged by a plan document.

## Verdict

**NEEDS_WORK.**
- **What draft 5 closes:**
  - Round-4 F1's procedure: §7.1 can run green, with LOW gaps.
  - F3: the two-claim kill now dies through the gate.
  - F4: the revisit trigger has homes future authors read.
- **D5 is the right shape for claims and requirements.** Its prerequisite list matches the reducers,
  and it is safe with idempotency, the cursor, `sync all` order and `verify`.
- **Three MED findings remain:**
  - **F1:** evidence ids are ledger-local, so §7.2's PASS control is red on correct code.
  - **F2:** holding an in-place verdict revision keeps a withdrawn SATISFIED outcome warranting a
    commit.
  - **F3:** the real `LedgerView` has no test outside `sdd`, and the premise that this repository has
    no v2 artifacts is false.
- F4-F6 are LOW.
