---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round4-20261118
status: NEEDS_WORK
reviewed_commit: 9c913e6047e00ccd911a25d959d77044114564f9
findings:
  - finding_id: r2-m3-spec-review-round4-20261118-F1
    severity: MED
  - finding_id: r2-m3-spec-review-round4-20261118-F2
    severity: MED
  - finding_id: r2-m3-spec-review-round4-20261118-F3
    severity: LOW
  - finding_id: r2-m3-spec-review-round4-20261118-F4
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round4-20261118.md
artifact_sha: bf1156e60cf5ef2ce287ab3e3cd459c345c9bcc124576e6d2228f311e67e8a44
---

# R2 M3 git-binding SPEC draft 4: non-author review, round 4

**Format note (2026-11-19, owner-approved):** the front-matter `findings` list was converted from flow style (`- {finding_id: …, severity: …}`), which the reviews connector's v1 parser rejects, to block style. The ids, severities and body are unchanged; `artifact_sha` was recomputed.

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**: it did not write drafts 1-4, the round-1 to round-3 verdicts, the
  sdd SPEC, or any code it read, and it has no authoring context from the session that wrote them
  (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft4.md` at HEAD
  `9c913e6047e00ccd911a25d959d77044114564f9`.
- **HEAD did not move.** It was `9c913e6` at the start and at the end of the review.
- **Evidence base.** `git diff --stat 36012cf 9c913e6` touches only the draft 4 file, the round-3
  verdict, `notes/state.md` and one `@generated` integrity pin. No kernel or script file has changed
  since round 3, so the code read here is the code at HEAD.
- **`git status` was the same at the start and at the end:** only the two permanent untracked paths
  (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`). Nothing in the repository was
  modified.
- **Read at HEAD:**
  - `cli/cli.go`: `parseCommand`, `committedVerdictReader`, `syncReviewsOnStore`, the
    `gates canary|enforce` branch, `resolveIntentGateHeadScope`, `gateDefinitionsForCommand`, the
    `sync git` / `sync all` order, `verify`'s step order, `syncIntentOnStore`, `intentProviders`,
    `openStore`, `newIDs`;
  - `gates/gates.go`: `LoadDir`, `Parse`, `validate`, `Evaluate`, `evaluateIntentRule`,
    `latestIntentProof`;
  - `projections/projections.go`: `Apply`, `Rebuild`, the commit reducer, `reduceIntent`,
    `reduceObligationVerdict`, `reduceRequirementReview`, `commitPayload.validate`;
  - `connector/reviews/reviews.go` (`Sync`) and `v2.go` (`ParseObligationVerdict`,
    `ParseRequirementReview`, `validateReference`, `ArtifactSHA256`);
  - `connector/git/connector.go` (`resolvedCommit`), `connector/git/gitcmd/gitcmd.go`
    (`trailerLineRE`, `intentTrailers`), `gitcmd_test.go` (`corpusResolver`,
    `TestIntentTrailers_SharedCorpusKernelBinds`);
  - `connector/intent/sdd/{build.go, provider.go}` (`Resolve`, `inDomain`, requirement and change
    payloads), `connector/intent/sdd/cli_test.go` (the untagged `cli.Run` harness);
  - `core/core.go` (`IDGenerator`, `SourceIntentSDD`);
  - `scripts/tests/fixtures/intent-trailers/cases.tsv`.
- **Executed** (with `$HOME/go/bin` first on PATH and GOCACHE under `<scratch>`):
  - `go test ./internal/connector/git/gitcmd/ -run TestIntentTrailers_SharedCorpus -count=1`: passes.
  - `go test ./internal/connector/intent/sdd/ -run 'TestSyncIntentSDDPrintsSkips|TestIntentAllExcludesSDD' -count=1`:
    passes. This confirms again that `cli.Run` on the embedded store runs untagged, in `make check`.

## Round 3, closure

| Round-3 finding | Verdict | Evidence |
|---|---|---|
| F1 MED: INV-8's pass half has no test | **Partly closed** | The vehicle is now the right one: `gates canary` runs `evaluateIntentRule`'s readiness query (`gates.go:285`) on the embedded store through `cli.Run`, untagged. But row 8 as written cannot produce **pass** for a correct implementation: its command order, its static verdict fixture and its single gate file each defeat it (new F1). Its two-claim kill cell cannot die through the gate (new F3) |
| F2 LOW: cross-version residual unowned | **Closed** | §9 owns the limitation, gives a revisit trigger, and §6 INV-8 now carries INV-10's exception. Where the trigger lives after M3 lands is a separate point (new F4) |
| F3 LOW: `SDD:add-auth` binds or skips | **Closed** | §4.1 now applies the grammar to the whole value with a lower-case prefix, §0b agrees, and §7 row 3 pins `SDD:add-auth` → `malformed` beside `sdd:add-auth` → resolved |
| F4 LOW: two proofs mis-described | **Closed** | §0a's row-10 line now says "re-specified". §5 states that `corpusResolver` returns `ClaimUnboundError{unknown-change}` for provider `sdd`, and that its comment is amended. Checked against the 32 rows below |
| F5 LOW: `verify` has no output channel | **Closed** | §4.4: `verify` prints no skip lines and discards both git syncs' skips. This matches `verify(ctx, root, ledger, projector, ids)`, which takes no writer (`cli.go:901`). No signature change is needed |

## Earlier partials, closure

| Finding | Verdict | Evidence |
|---|---|---|
| Round-1 F6 (derivations) | **Partly closed** | Rows 1-7, 9 and 10 derive from their invariants. Row 8 does not run as written (F1), and one of its kill cells cannot die (F3) |
| Round-1 F8 (records/specdir hazard unowned) | **Partly closed**, not re-raised | `docs/deferrals.tsv` still has no row. D4 is with the owner, and deferral rows are owner-dormant (CLAUDE.md, session protocol). This is an owner decision, not a spec defect |
| Round-2 F4 / F7 / F9 | **Closed** | Row 10 is feasible untagged (round 3). `L` is lazy. The case-folded prefix removes the wedge, and F3 above fixes the reading |

## Answers to the attacks

### Does `gates canary` load `<root>/gates/*.yaml` and evaluate readiness on the embedded store, untagged?

Yes.
- `run` → `gates.LoadDir(filepath.Join(root, "gates"))` (`cli.go:621`) reads every `*.yaml` in the
  **working tree** (not HEAD), so gate files need not be committed.
- **The format is JSON.** `Parse` canonicalises JSON and decodes it with `DisallowUnknownFields`.
  The `.yaml` name is only a suffix filter. The draft does not say this, but `cli_test.go:585`
  shows the shape.
- `evaluateGate` defaults to `gates.Evaluate` (`cli.go:188`). `openStore` uses `stateRoot(root)`, so
  the ledger persists across successive `cli.Run` calls in one tenant root. `sdd/cli_test.go` is the
  untagged precedent.

### Does `sync reviews` ingest committed verdict and review files from the fixture's HEAD?

Yes, from HEAD only.
- `committedVerdictReader` runs `git ls-tree -r HEAD -- docs/verification/verdicts` and then
  `git show HEAD:<path>` (`cli.go:281-303`).
- `reviews.Sync` dispatches `proofbound.obligation-verdict.v2` and `proofbound.requirement-review.v1`
  (`reviews.go:141-157`).
- **A file must be committed at HEAD to be read.** Since a verdict names its `reviewed_commit` by sha,
  it can only be committed in a **later** commit than the one it reviews (F1).

### Can a VERIFIABLE requirement review and a SATISFIED v2 verdict be authored against `intent.sdd` records?

Yes, as far as the schemas go: nothing in them is specific to `intent.records`.
- `validateReference` (`v2.go:283`) checks `record_kind`, `relation`, a well-formed source, an id
  matching `idRE`, a 64-hex digest, and no obligation ids. `intent.sdd` is well-formed.
- `reduceRequirementReview` looks up `requirements_view` by `(source, id, digest)` and requires the
  reviewer to differ from `declared_owner`. sdd requirements carry an owner and `active` obligations
  (`build.go:33, :51`).
- The gate's join uses `'intent.'||ci.provider`, which gives `intent.sdd` = `core.SourceIntentSDD`.
  sdd change intents target `intent.sdd` requirements (`build.go:84`), so `intent_targets_view`
  joins.
- **But three schema facts constrain the fixture, and the draft states none of them** (F1):
  1. A `SATISFIED` outcome needs at least one `evidence_event_ids` entry (`v2.go`, "satisfied outcome
     requires evidence"). The reducer requires each id to be in `events` at a sequence number no
     higher than the verdict's own (`projections.go`, "evidence … is from the future"). Event ids
     are random ULIDs (`newIDs` uses `crypto/rand`), so the id is known only after a sync has run.
  2. Both artifacts carry a self-digest (`ArtifactSHA256` zeroes the last `artifact_sha256` value).
  3. The verdict names `reviewed_commit`, the `sdd` change-intent revision and the requirement
     digests, and each of these exists only after the fixture is committed and synced.

  So the verdict file cannot be a static file in the fixture. It must be generated during the test
  and committed as a new HEAD.

### Does `resolveIntentGateHeadScope` or `scope_commit` interfere with a non-HEAD commit?

No. It only rewrites `scope_commit: "HEAD"` (`cli.go:750-770`). An exact 40-hex `scope_commit`
passes `validate` and is used as is. `latestIntentProof` matches a git `commit.recorded` event whose
native id is the sha.
- **HEAD scope cannot be used for row 8, though.** At canary time HEAD is the verdict commit, not the
  bound commit (previous attack). One definition has one `scope_commit`, and `LoadDir` rejects
  duplicate ids. So three verdicts (pass, skipped-only, two-claim) need three definitions with
  distinct ids, each holding a sha known only at run time. The draft names one file,
  `gates/intent-delivery-readiness.yaml` (F1).

### Is "blocked" what canary reports for a commit with no targets?

Yes, with two conditions.
- `targets = 0` fails `targets > 0 && …`, giving `state=BLOCKED would_block=true`.
- **Condition 1:** a commit event for the sha must exist. Otherwise `latest == nil` gives
  `UNKNOWN`. A skipped commit still has its event under D2, so this holds.
- **Condition 2:** `Apply` must succeed. If `Apply` fails and a proof event exists, every
  readiness gate reports BLOCKED (`gates.go:266-272`). So a BLOCKED assertion holds vacuously on a
  wedged ledger. Only a PASS on the **same** ledger shows that a BLOCKED is meaningful. Row 8
  does put all three commits on one ledger.

### `corpusResolver` under M3: do the 32 rows hold?

Yes.
- No existing value case-folds to `sdd:`. The values are bare `CI-…`, `records:`, `specdir:`,
  `github:`, `not-an-id`, `records:specdir:…` and `records:…:extra`.
- For every non-`sdd` value, `resolvedCommit` (`connector.go:180-209`) behaves exactly as today,
  and `corpusResolver` echoes a well-formed ref. So "a ref is bound" and "sync succeeded with
  trailers" agree row for row, as round 3 showed.
- The test passes at HEAD.
- The new `sdd:` rows are unbound under both a `malformed` skip and an `unknown-change` skip, and
  the unchanged gate refuses them.

### The whole-value grammar against the sdd name domain

- `^sdd:[a-z0-9][a-z0-9-]{1,99}$` accepts a strict subset of `inDomain` (`provider.go:135`), which
  allows any name without `/` or NUL.
- So a tenant change named with upper case, `_` or `.`, or with a single character, can never be
  claimed. It is a permanent `malformed` skip, never a wedge.
- The 147-name OpenSpec survey all fits, and D1 is the owner's choice. So this is a stated
  restriction, not a defect.
- `trailerLineRE` (`(.+)`) passes any such value through to the connector, so the skip is reached.

### Discarding skips in `verify`

Sound. `verify` has no writer, skips are not events, and its idempotence check compares append
counts, which skips do not change. A skip computed twice and discarded twice has no observable
effect.

## Findings

**F1 (MED): §7 row 8, as specified, cannot produce PASS for the verified commit on a correct
implementation, so M3-INV-8's pass half still has no test that can run green.**
- **Command order.** Row 8 runs `sync intent sdd` → `sync reviews` → `sync git` → `gates canary`.
  - `sync reviews` therefore appends the verdict at a lower sequence number than the commit event
    it reviews.
  - `Apply` reduces in sequence order. `reduceObligationVerdict` first requires
    `commits_view` to hold `reviewed_commit` ("v2 verdict reviewed commit is dangling"), and then
    requires an exact `commit_intents_view` row ("does not match an exact commit intent claim").
    Neither row exists yet at the verdict's sequence number.
  - `Apply` fails, so `evaluateIntentRule` reports BLOCKED for **every** readiness gate, the
    verified commit's included.
  - The ledger is append-only and `Rebuild` replays in the same order, so that ledger stays wedged
    for good. The order every existing caller uses is the reverse: `sync all` and `verify` sync git
    before reviews (`cli.go`, `sync all` branch; `verify` steps at `:902-914`).
- **A static fixture.** "The sdd tenant fixture carries … committed verdict and review files"
  reads as files checked into testdata. A SATISFIED verdict cannot be static:
  - it must name `reviewed_commit`;
  - it must name a random-ULID evidence event id that is already in the ledger at a lower sequence
    number;
  - it must name the ledgered revision digests and its own self-digest.

  The harness has to write it and commit it as a new HEAD after the first git sync. The draft does
  not say this.
- **One gate file.** One definition has one `scope_commit`. HEAD scope names the verdict commit,
  not the bound one. Three commits need three exact-sha definitions written at run time (attacks
  above).
- **The effect.** An implementer who follows row 8 literally gets a red pass control on correct
  code, and cannot tell that apart from an M3 defect. Under Law 6 the derivation for the invariant
  O-8 needs is still not executable as written. This is the third round in which INV-8's
  derivation has been the open MED. The vehicle is now right; the procedure is not.

**F2 (MED): M3 makes a permanent projection wedge reachable from correct tenant data under ordinary
command orders. M3-INV-5, M3-INV-10's "an obligation verdict citing it always projects", and
§4.3's recovery are each false once `sync reviews` interleaves.**
- **Before M3**, this ordering was unreachable:
  - An `sdd:` trailer failed `sync git` and `sync all` as a whole, so no later `sync reviews` step
    in `sync all` ran.
  - `records` and `specdir` claims are resolved by `sync all`'s own intent step, or the git sync
    fails.
  - So a verdict could never be ingested ahead of the binding it cites.
- **Under M3**, a claim that is `not-ledgered` is a skip, and the sync **succeeds**.
  - `sync all` never syncs `sdd` (SDD-INV-56, `intentProviders`), and it syncs reviews **after**
    git in the same run.
  - Take a tenant that commits change revision R at commit X with `Intent: sdd:<name>`, and later
    commits a correct SATISFIED verdict for X.
  - If that tenant runs `sync all` before `sync intent sdd` has ledgered R, the ledger gets:
    1. X's event with **no** refs (`not-ledgered`; D2 also drops any `records` ref on X);
    2. then the verdict.
  - At the next `Apply`, the verdict fails "does not match an exact commit intent claim".
  - §4.3's recovery ("run `sync intent sdd`, then `sync git`") appends X's bound event at a
    **higher** sequence number. Replay still reduces the verdict first, so the projection never
    recovers.
  - Every readiness gate, report and `verify` on that ledger fails from then on.
- **The same holds on a fresh ledger per run** (a CI pipeline built on `sync all`, as
  `delivery-enforce.sh` is): there, every `sdd` claim is always `not-ledgered`, so any committed
  sdd verdict wedges every run.
- **D2 widens it to `records`.** A correct `records` verdict on a commit that also carries a
  skipped `sdd` claim wedges in the same way.
- **What the draft claims.**
  - M3-INV-5 claims a rebuild succeeds "after any interleaving of `sync git` and
    `sync intent sdd`", and its derivation (row 5) never includes `sync reviews`.
  - M3-INV-10's "always projects" holds only when the binding precedes the verdict.
  - Round 3 accepted `sync all`'s exclusion of `sdd` on the strength of §4.3's recovery. That
    recovery does not hold once a verdict is in the ledger.
- **The class.** This is the round-2 F1 wedge class ("a reduce step depends on a row the ledger
  does not hold at that sequence"), reached from the other direction. Round 2 compiled the
  deletion half into M3-INV-10, but not this ordering half.

**F3 (LOW): row 8's kill cell "keeping the bindable refs of a commit with a skipped claim (the
two-claim case must be blocked)" cannot die through the gate on row 8's ledger.**
- **Under the mutant**, the two-claim commit keeps its bindable ref, so `targets > 0`.
- **It still reads BLOCKED** unless a SATISFIED verdict and a VERIFIABLE review cover that commit.
  - So the gate reads BLOCKED under both the mutant and the correct code, unless the fixture
    supplies a SATISFIED verdict citing the two-claim commit's bindable ref.
  - Under correct code that verdict cannot project: the ref is dropped, giving "does not match an
    exact commit intent claim".
  - That fails `Apply` and turns the verified commit's PASS into BLOCKED.
- **On one ledger**, the kill is therefore either unreachable (no verdict) or incompatible with the
  pass control (with a verdict).
- **The calibration rule fails.** §7's rule that "a hand mutant per kills cell must make it fail"
  cannot be met for this cell.
- **Coverage elsewhere.** The mutant itself is killed elsewhere: row 3's git-unit two-claim cases
  ("partial binding"), and draft 3's `intent check --commit` binding check, which draft 4 dropped
  from row 8.
- **INV-8's clause is vacuous on a projectable ledger.** INV-8's two-claim refusal ("one skipped
  and one otherwise-satisfied claim") describes a commit whose "otherwise-satisfied" verdict cannot
  project at all. So on a projectable ledger that state does not exist, and on a ledger that holds
  the verdict, the refusal comes from F2's global wedge, not from the gate's logic.

**F4 (LOW): §9's revisit trigger binds future SPECs (M2, a `Resolve` change, a stricter §4.2) but
has no home they will read.**
- **Where it lives.** The trigger lives only in this plan document.
- **What §5 lists.** §5's site list updates the sdd SPEC's §12, §14 and SDD-INV-56 to "record M3
  as landed". It does not carry the accepted limitation or the trigger into the sdd SPEC's M2 entry
  conditions (§14), nor into the git SPEC's new INV rows.
- **The effect.** The M2 author reads the sdd SPEC, not a superseded draft of a plan. So the
  obligation, "state whether existing bindings stay, and ship a migration if not", can be missed.
  That is the drift that round-3 F2 asked to close (Law 2: the fact needs exactly one home, and
  that home has to be where the reader is).

## Other observations (no finding)

- **Output casing.** Row 8 says "pass" and "blocked"; canary prints `state=PASS` and
  `state=BLOCKED`.
- **The gate file is JSON**, despite its `.yaml` name. An implementer copying the name may write
  YAML, which `Parse` rejects.
- **`L`'s digest** comes from the decoded payload (the envelope has no digest column). There is one
  choice, as round 3 noted.

## Law walk

- **Law 1.** The `@generated` pins (route matrix, connector reality) are listed for regeneration
  after the code commits. Compliant.
- **Law 2.**
  - D1's grammar lives in the connector and the VD; the trailer gate script is unchanged.
  - The cross-version limitation now has an owner, but its trigger has no home future authors will
    read (F4).
  - D4 still lives only in this plan, pending the owner.
- **Law 6.**
  - The SPEC precedes code and has an interface lock, a derivation table and a calibration
    statement.
  - INV-8's derivation is not executable as written (F1), and one of its kill cells cannot die
    (F3).
  - INV-5 and INV-10 claim properties that F2 refutes.
- **Law 7.** The deletion half of the verdict-wedge lesson is compiled (INV-10). The ordering half,
  a verdict reduced before the binding it cites, is not, and M3 is what makes it reachable (F2).
- **Law 8.** No new dependency. Compliant.
- **Law 9.**
  - The draft asks for a non-author re-check, and this is one.
  - The calibration rule is stated. Applying it to row 8 would expose F1 (the pass control is red
    on correct code) and F3 (a kill cell survives). Both are found here, before any test is
    believed.

## Verdict

**NEEDS_WORK.**
- Draft 4 closes round 3's four LOW findings. It also moves INV-8's proof onto the right vehicle:
  `gates canary` on the embedded store, untagged, ingesting committed sdd-sourced verdicts and
  reviews, all of which this review confirmed against the code.
- Two MED findings remain:
  - **F1:** row 8's prescribed procedure (reviews before git, a static verdict fixture, one gate
    file) cannot yield PASS on a correct implementation.
  - **F2:** M3's success-with-skip makes a permanent projection wedge reachable from correct tenant
    data under `sync all`. This falsifies M3-INV-5, M3-INV-10 and §4.3's recovery.
- F3 and F4 are LOW and may remain under ACCEPTABLE.
