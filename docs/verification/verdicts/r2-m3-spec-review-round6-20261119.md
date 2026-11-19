---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round6-20261119
status: NEEDS_WORK
reviewed_commit: 6a4a0df57aab136860ac6b53f5d612d537ba711b
findings:
  - finding_id: r2-m3-spec-review-round6-20261119-F1
    severity: MED
  - finding_id: r2-m3-spec-review-round6-20261119-F2
    severity: MED
  - finding_id: r2-m3-spec-review-round6-20261119-F3
    severity: LOW
  - finding_id: r2-m3-spec-review-round6-20261119-F4
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round6-20261119.md
artifact_sha: fed64fc3964210d40a55161c811c0dd7bb5105f288a85ca33918dbd8c0abd7ce
---

# R2 M3 git-binding SPEC draft 6: non-author review, round 6

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**: it did not write drafts 1-6, any earlier verdict, the sdd SPEC, or
  any code it read (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft6.md` at
  `6a4a0df57aab136860ac6b53f5d612d537ba711b`, together with draft 5, which draft 6 incorporates by
  reference.
- **HEAD moved during the review**, to `e470f61615dae22bac7b2512761b8390daf22ad4`.
  `git diff --stat 6a4a0df HEAD` shows that only `notes/state.md` changed, so the draft and all code
  are the same at both commits.
- **Read at HEAD:**
  - draft 6, draft 5, and the round-5 verdict;
  - `connector/reviews/{reviews.go, v2.go, SPEC.md}`, plus the `Deps` construction sites in
    `reviews_test.go` and `cli.go`;
  - `connector/git/connector.go`: `Commit`, `IntentRef`, `Sync`, `normalizedCommit`,
    `resolvedCommit`;
  - `projections/projections.go`: `reduce` (the commit route), `eventRoute`, `reduceIntent`,
    `requireExactRecord`, `reduceReview`, `reduceObligationVerdict`, `reduceRequirementReview`,
    `commitPayload.validate`;
  - `connector/intent/intent.go`: `Reference`, `ValidateRevision`;
  - `connector/intent/sdd/{build.go, provider.go}`: `Revisions`, `buildRecord`, the change-intent
    targets, `build`;
  - `migration/migration.go` in full, plus `migration_test.go`;
  - `store/store.go`: `ImportHistoricalEvidenceRecords`, `importRecords`;
  - `cli/cli.go`: `openStore`, `openHistoricalEvidenceStore`, `syncGit`, `intentProviders`,
    `syncReviewsOnStore`, `committedVerdictReader`, `runCommand` (the migrate, gates and sync-all
    branches), `verify`, `requireNoSecondSyncAppends`;
  - `gates/gates.go`: `evaluateIntentRule`, `latestIntentProof`;
  - `docs/verification/p6-historical-evidence.jsonl`.
- **Executed** (with `$HOME/go/bin` first on PATH and GOCACHE under `<scratch>`). A scratch external
  test, `package migration_test`, was added through `go test -overlay` and was never written into the
  repository. It made two checks:
  - `migration.Parse` on a one-record archive built from each of the five committed archive lines,
    one at a time. All five are rejected with `historical evidence archive: got 1 records, want 5`.
  - The full committed archive, imported with `migration.Import` into an empty embedded store, then
    `projections.Rebuild`: it passes.
  - No `GOFLAGS=-mod=mod` was used. `git status` before and after shows only the two permanent
    untracked paths, and no tracked file was rewritten.

## Round-5 closure

| Round-5 finding | Verdict | Evidence |
|---|---|---|
| F1 MED: evidence ids are ledger-local; §7.2's PASS control is red | **Closed in the contract, but the new §7.2 cannot run (F2)** | §4.6 rule 4 drops evidence from the hold, and §1's "order-only" list is now items 1-3, which matches `reduceObligationVerdict`. But the portable-evidence step depends on a one-record archive, and `migration.Parse` rejects that (F2) |
| F2 MED: holding an in-place revision keeps a withdrawn SATISFIED outcome warranting a commit | **Closed for the fail-open; it opens a wedge (F1)** | Rule 1 appends any artifact whose `(kind, native id)` is already ledgered. The upsert on `(verdict_id, requirement…, obligation_id)` then overwrites the outcome, or Apply fails, so a withdrawn outcome can no longer survive because of the hold. Appending without a check, though, re-admits the ordering wedge for revisions, and M3-INV-5 denies that wedge |
| F3 MED: the real `LedgerView` is untested outside `sdd`; the "0 files" premise is false | **Closed** | §1 now lists the three committed artifacts. Row 12 exercises the real view over records, specdir and sdd requirements, git refs of each provider, and imported events, and counts reads. All of that is feasible on the embedded store. Its "imported event" part needs the full archive (F2) |
| F4 LOW: §7.1 under-specifies the artifacts | **Mostly closed; residue in F4** | Obligation ids, the requirement id and digest, distinct verdict ids, X2/X3 outside `openspec/`, and the gate's `description` and `expires` are all named now. The `change_intent` reference's kind and relation, and the empty `obligation_ids`, are not |
| F5 LOW: M3-INV-5 overreaches; event-id semantics are unstated | **Closed as worded; refuted on substance by F1** | The scope is now `sdd` claims, content-valid artifacts and the records/specdir carve-out. `EventLedgered` is gone. The restated promise is still false for revisions (F1) |
| F6 LOW: reviews SPEC text not listed; second decoder; lazy read untested | **Closed** | §5 lists the Purpose sentence, R-INV-11 and the requirement-review sentence. `connectorgit.Commit` is the type the producer marshals: `connector.go:132` calls `json.Marshal(commit)` on `Commit` (`:28`). Row 12 counts reads |

## Answers to the attacks

### 1. Round-5 closure

See the table.

### 2. The first-revision rule

- **What the rule gets right.** It closes round-5 F2's fail-open. A revision is always appended, so a
  same-key withdrawal overwrites the outcome or fails Apply. A held first revision adds no rows, so it
  adds no warrant.
- **Can a withdrawn outcome survive by gaming the rule?** Only through paths that exist before M3:
  - a revision that moves to a new requirement revision writes rows under new keys, so the old
    SATISFIED rows stay;
  - a withdrawal under a new `verdict_id` never removes another verdict's SATISFIED row, because the
    gate's query is `EXISTS … outcome='SATISFIED'`.

  The hold does not worsen either of these.
- **The identity is wrong in three ways (F3):**
  - v1 and v2 verdicts share the kind `review.verdict` (`reviews.go`:
    `kind := core.KindReviewVerdict` for both), and they share the `idRE` grammar;
  - two different files carrying the same id are treated as revisions of one artifact;
  - the `Ledgered` set is a snapshot taken once per sync, so it misses appends made earlier in the same
    sync.
- **Does appending a revision unconditionally reintroduce a reachable wedge?** **Yes (F1).**

### 3. Can §7.2 run?

- **Step 1 cannot (F2).** `migration.Parse` requires exactly the five pinned records. It checks each
  line's SHA-256, its seq, its event id, source, native id, kind and timestamps. So there is no "one
  record with a fixed event id `E`" of the test's choosing.
- **The full five-line archive does work.** Executed: copied verbatim into the tenant at
  `migration.ArchivePath` (read from the working tree, no commit needed), it imports into an empty
  ledger and rebuilds cleanly. Then `E` would have to be one of the pinned ids, for example the
  `checks` event `01M28TPW9C8R7ND19MNDCJ9GDG` at seq 1670. Later appends start above 1834 (`setval`),
  so the evidence precedes the verdict.
- **The `Import` emptiness check holds.** It reads `ReadEvents(Filter{})`, and
  `openHistoricalEvidenceStore` sets `AllowHistoricalEvidenceImport`. Running it first on a fresh root
  satisfies both.
- **Computing references through `Revisions` is sound:**
  - The digests are `manifest` over git mode, OID and path lines, so they are content-derived.
  - `observedAt` touches only `ObservedAt`, not the payload or the digest.
  - X4 adds files only under `docs/verification/`, so the `sdd` state key is unchanged and the
    revisions are identical.
- **The readiness gate needs:**
  - a git event for the scoped commit (`latestIntentProof`);
  - Apply to succeed;
  - claim, target, SATISFIED-verdict and VERIFIABLE-review rows.

  Nothing else. §7.2 never says to write `gates/x1.yaml` (F4).
- **Interleaving outputs are correct at every step, once the archive is fixed:**
  - **A:** after `sync all`, X1 is `claim-skip not-ledgered`, and the held causes are
    `claim-not-ledgered` for the verdict and `requirement-not-ledgered` for the review. Rebuild
    succeeds. `sync intent sdd` and `sync git` re-emit X1 with refs; the content differs, so it is a
    new event. Both reviews then append, and the gate reads PASS.
  - **B:** the review appends after `sync intent sdd`. The verdict stays held on its claim until
    `sync git`, then appends.

### 4. Row 12 and the `storeLedgerView` sketch

- **Decoding is right.** The git payload is exactly `json.Marshal(connectorgit.Commit)`.
  `IntentTrailers` is `json:"-"`. Projections decode the same bytes with `commitPayload`.
- **Requirement digests are uniform.** All three intent sources route to `reduceIntent`, which decodes
  `connectorintent.Requirement` and keys `requirements_view` by `(event source, payload
  requirement_id, payload artifact_sha256)`. So the payload's `artifact_sha256` is correct for records,
  specdir and sdd.
- **No other source emits `requirement.recorded`** (`eventRoute`).
- **The sketch is feasible.** One small gap: §4.6 does not say what the view does when a historical git
  payload fails to decode. The draft is silent on whether that fails the sync.

### 5. §7.1 against `v2.go`

- **Top-level key order is right for both schemas.** Verdict: `schema` … `artifact_sha256`, 11 keys.
  Review: 7 keys.
- **Relations and vocabularies are right:** `evaluates` and `reviews`, VERIFIABLE, SATISFIED,
  ACCEPTABLE.
- **Ids are fine.** `idRE` admits `OV-m3-x1` and `RR-m3-x`; no `OV-` or `RR-` prefix is required.
- **What is unnamed (F4):**
  - `change_intent` must carry `record_kind` `change_intent` and relation `evaluates`;
  - every verdict reference must have empty `obligation_ids`, but the sdd change-intent targets this
    procedure reads back carry `implements` and `obligation_ids` (`build.go:84`), so copying them fails
    `validateReference`;
  - obligation outcomes must be sorted and unique.
- **Why LOW.** The calibration rule would catch these.

### 6. Sites and laws

- **The `reviews.New` sites are complete:** only `cli.go:455` and `reviews_test.go` lines 69, 607, 619
  and 639 construct it.
- **R-INV-5 has dropped out of the list.** Draft 6 §5 replaces the reviews SPEC row and no longer
  mentions it (draft 5 recorded it as unchanged). This is harmless if read as "plus draft 5's row", but
  the replace-not-add wording leaves that unclear.
- **Law walk:** below.

## Findings

**F1 (MED): appending every revision unconditionally reintroduces, from correct data, the ordering
wedge D5 was adopted to close. That makes the restated M3-INV-5 false, and §4.6 "why the hold is
sufficient" covers only first revisions.**
- **The rule.** §4.6 rule 1 appends any artifact for which `Ledgered(kind, nativeID)` holds, "as before
  M3".
- **Why it wedges permanently.** Both review reducers check their prerequisites against the views as
  they stand at the event's sequence number:
  - `reduceRequirementReview`: "requirement review revision is dangling";
  - `reduceObligationVerdict`: "reviewed commit is dangling", "does not match an exact commit intent
    claim", "requirement revision is dangling".

  The ledger is append-only, so an event that is projected before its prerequisites fails every
  `Rebuild` from then on.
- **A reachable sequence, all inside M3-INV-5's scope** (an `sdd`-only tenant with content-valid
  artifacts):
  1. `RR-1` reviews requirement `R@d1` and `OV-1` warrants X1. Both are ledgered and the gate reads
     PASS.
  2. The author edits the spec and commits X5 with `Intent: sdd:add-auth`. sdd's requirement id embeds
     the digest, so this is a new requirement, `R'@d2`.
  3. The reviewer updates `RR-1` in place to review `R'@d2`. Optionally they also revise `OV-1` in
     place to `reviewed_commit` X5.
  4. CI runs `sync all` (delivery-enforce does). It never syncs `sdd` (`intentProviders`,
     SDD-INV-56).
  5. `sync git` reports X5 `claim-skip not-ledgered`.
  6. `sync reviews` finds `Ledgered(requirement.reviewed, RR-1)` true and appends `RR-1` rev2
     unconditionally. If `OV-1` rev2 is present it does the same, citing an unbound X5 claim.
  7. Apply fails, and keeps failing after `sync intent sdd` and `sync git`. The rev2 event sits at a
     lower sequence number than `R'@d2` and X5's binding, for good.
- **Why rev2 counts as content-valid.** By §6's own definition it "projects on a ledger that holds
  every prerequisite it cites". So M3-INV-5's "a projection rebuild succeeds after any interleaving" is
  refuted.
- **Before M3**, the same flow also wedged. So M3 does not create the wedge. But D5 exists to close
  exactly this class (round-4 F2), and draft 6 states an invariant that claims it is closed.
- **How the draft hides it.**
  - §4.6 frames the revision outcome as "fails Apply loudly", but a failed Apply here is a permanent
    wedge of the ledger, not a per-artifact error.
  - The "accepted cost" paragraph does not mention revisions.
  - Row 11's revision case asserts only that the revision is appended. No row tests a rebuild
    afterwards, so nothing derives the invariant for revisions.
- **This is the round-5 F2 / wedge trade-off resolved by assertion.** Holding revisions is fail-open;
  appending them wedges. The contract picks appending and then claims the wedge is absent (Law 6: an
  invariant whose derivation covers only half its domain).

**F2 (MED): §7.2, the derivation for M3-INV-5, and row 12's "historically imported event" prescribe a
one-record archive, which `migrate historical-evidence` rejects. So the positive control is red on
correct code for the third round running.**
- **What the draft prescribes.** §7.2 step 1 has the tenant carry
  `docs/verification/p6-historical-evidence.jsonl` with "**one** record … with a fixed event id `E`".
- **What the code requires.** `migration.Parse` (`migration.go`) fails unless
  `len(lines) == len(expected)`, which is 5. It then checks each line's SHA-256, seq, event id, source,
  native id, kind, connector version and timestamps against the pinned `expected` table.
- **Executed:** each of the five committed lines, used alone as a one-record archive, fails with
  `got 1 records, want 5`.
- **Consequences.**
  - Step 3.1 (interleaving A) and the first step of interleaving B both fail.
  - No archive with a test-chosen `E` can be imported at all.
- **What does run.** Executed: the full committed archive imports into an empty ledger and rebuilds.
  The procedure can therefore run with the five-line archive copied verbatim and `E` set to one of its
  pinned ids. The draft says neither.
- **Why MED.** Round-5 F1 was the same class: a prescribed derivation that cannot run green on a
  correct implementation (Law 6).

**F3 (LOW): the "first revision" identity `(kind, nativeID)` misclassifies artifacts. A first revision
can then be appended without the hold, and `verify` can fail on its idempotence pass.**
- **v1 and v2 verdicts share an identity space.** Both are emitted with kind `review.verdict`
  (`reviews.go` `Sync`), and both draw `verdict_id` from the same `idRE`. So a v2 artifact whose
  `verdict_id` equals a ledgered v1 verdict's id counts as a "revision" and is appended
  unconditionally. One example is a verdict file upgraded in place from v1 to v2.
- **Distinct files sharing an id are conflated.** A v2 verdict copied to a new file for another commit,
  keeping its id, is not a revision of one artifact, yet it is treated as one.
- **The snapshot does not see same-sync appends.** The set is filled once per sync and not updated on
  append. Take two files with the same `(kind, id)`, where the first appends and the second's
  prerequisites are missing:
  - pass 1 of `verify` holds the second file;
  - pass 2 sees the id ledgered and appends it;
  - `requireNoSecondSyncAppends` then fails `verify`;
  - if the prerequisites are still missing, Apply wedges as in F1.
- **Effect.** No fail-open, but a hold bypass and a wedge or verify failure on data the parser accepts.
  Row 11 has no case for any of these.

**F4 (LOW): residual under-specification in §7.1 and §7.2.**
- **The verdict's `change_intent` reference.** It must have `record_kind` `change_intent`, relation
  `evaluates` and no `obligation_ids` (`validateReference`). Step 3 reads the sdd change-intent
  payload, whose targets carry relation `implements` and `obligation_ids` (`sdd/build.go:84`). The
  draft names relations only for `requirements` and the review.
- **Obligation outcome ordering.** Outcomes must be sorted by
  `(source, requirement_id, artifact_sha256, obligation_id)`; the draft does not say so.
- **The x1 gate in §7.2.** §7.2 asserts `gates canary` on x1 but never writes `gates/x1.yaml` for its
  tenant.
- **Why LOW.** The calibration rule would surface each of these.

## Other observations (no finding)

- **Revisions that move to a new requirement revision.** A revision that changes its requirement
  revision writes new `obligation_verdicts_view` rows, so the old SATISFIED rows keep warranting the
  old commit. This is pre-existing upsert semantics, and M3 neither causes nor worsens it. The owner
  may want it on record beside the evidence-portability item in §9.
- **The `storeLedgerView` error path.** It should state what happens when a ledgered git payload does
  not decode as `connectorgit.Commit`: fail the sync, or skip the event.
- **The evidence-portability item.** It is correctly classed as pre-existing and is now raised to the
  owner, not hidden.

## Law walk

- **Law 1.** No generated artifact is hand-edited. The regeneration pins are carried over from draft 5.
  Compliant.
- **Law 2.**
  - There is no second git-payload decoder: the producer's own type is used.
  - The amended reviews SPEC sentences are listed.
  - Compliant.
- **Law 6.**
  - M3-INV-5 is refuted for revisions (F1).
  - Its derivation, and row 12's import step, cannot run as written (F2).
  - The SPEC still precedes code, and the calibration rule is stated.
- **Law 7.** The wedge lesson is now compiled for first revisions only. The revision half is asserted,
  not enforced (F1).
- **Law 8.** No new dependency. Compliant.
- **Law 9.** This is a non-author re-check. F2 is again a control the author never ran. Executing
  `migration.Parse` once would have shown it.
- **Law 10.** Not engaged by a plan document. HEAD's only move during the review was a `notes/state.md`
  update.

## Verdict

**NEEDS_WORK.**
- **What draft 6 closes:**
  - round-5 F2's fail-open: revisions now overwrite or fail, and never survive through the hold;
  - F3: the real view is derived across providers, and the premise is corrected;
  - F5 and F6 as worded;
  - most of F4.
- **The `LedgerView` decoding and the requirement-digest source are correct for all three
  providers.**
- **Two MED findings remain:**
  - **F1:** unconditional revisions re-admit the permanent ordering wedge from correct sdd data,
    contradicting M3-INV-5.
  - **F2:** §7.2's one-record archive is rejected by `migration.Parse`. The full five-line archive
    works.
- F3 and F4 are LOW.
