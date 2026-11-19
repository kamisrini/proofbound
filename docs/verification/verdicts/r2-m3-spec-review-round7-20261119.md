---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round7-20261119
status: ACCEPTABLE
reviewed_commit: 26de649d9abbb6956cf8640254d75fc6d1580ddb
findings:
  - finding_id: r2-m3-spec-review-round7-20261119-F1
    severity: LOW
  - finding_id: r2-m3-spec-review-round7-20261119-F2
    severity: LOW
  - finding_id: r2-m3-spec-review-round7-20261119-F3
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round7-20261119.md
artifact_sha: c411b1d24125fd8641e509b0eb6133eba69d8aa5785aa31ca899515c78dbed41
---

# R2 M3 git-binding SPEC draft 7: non-author review, round 7

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). They are a **non-author**: they did not write drafts 1-7, any earlier verdict, the sdd SPEC,
  or any code they read (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft7.md` at
  `26de649d9abbb6956cf8640254d75fc6d1580ddb`, together with drafts 6 and 5, which it incorporates by
  reference.
- **HEAD moved during the review**, to `b0bd8883a4080c43579eda0b0aacf06fd4bce646` ("notes: D6 taken; M3
  draft 7 under round-7 review"). `git diff --stat 26de649 HEAD` shows only `notes/state.md` changed.
  The drafts and all code are identical at both commits.
- **Read at HEAD:**
  - drafts 5, 6 and 7, and the round-6 verdict;
  - `connector/reviews/{reviews.go, v2.go, SPEC.md}`;
  - `cli/cli.go`: `syncGit`, `committedVerdictReader`, `syncIntentOnStore`, `intentProviders`,
    `printableUnit`, `syncReviewsOnStore`, `runCommand` (the sync-reviews, migrate, gates, rebuild and
    sync-all branches), `resolveIntentGateHeadScope`, `verify`, `requireNoSecondSyncAppends`;
  - the `cli_test.go` structural and nil-store tests that touch `syncReviewsOnStore`;
  - `store/store.go`: `BeginSync`, `Sync.Append` (the `ON CONFLICT(source,native_id,content_sha)`
    dedupe), `Sync.Finish`;
  - `projections/projections.go`: `reduceReview`, `reduceObligationVerdict` (including the
    `obligation_verdicts_view` upsert), `reduceRequirementReview`, `commitPayload`, `decode`;
  - `gates/gates.go`: `evaluateIntentRule`, `latestIntentProof`;
  - `migration/migration.go` (the pinned `expected` table);
  - `connector/git/connector.go`: `Commit`, `IntentRef`;
  - `connector/intent/intent.go`: `Reference`;
  - `core/core.go`: `ContentSHA`, `Canonicalize`;
  - `scripts/delivery-enforce.sh`, `Makefile`;
  - `docs/verification/p6-historical-evidence.jsonl`.
- **Executed.** `$HOME/go/bin` was first on PATH and GOCACHE was under `<scratch>`. No
  `GOFLAGS=-mod=mod` was used.
  - Built the CLI to `<scratch>/proofbound`.
  - Built a scratch tenant at `<scratch>/tenant` with an OpenSpec sdd profile (owner `owner-a`), a
    change `add-auth` with one requirement, and the repository's five-record archive copied verbatim.
    Its only commit, X1, has no trailer: at HEAD the git connector cannot bind `sdd:` values.
  - Ran `migrate historical-evidence` (`imported=5`), then `sync all` (rc 0:
    `intent appended=0 git appended=1 checks appended=0 sessions appended=0 reviews appended=0`), then
    `rebuild` (rc 0), then `sync intent sdd` (rc 0), then `rebuild` (rc 0).
  - Edited the spec statement and committed again, then re-ran `sync intent sdd`
    (`listed=4 appended=2 existing=2`).
  - Dumped the ledger with a scratch `package cli` test that calls `openStore`, supplied through
    `go test -overlay` and never written into the repository. The dump shows:
    - the archive at seqs 1227/1228/1229/1670/1834;
    - the first new append at seq 1835;
    - `E` = `01M28TPW9C8R7ND19MNDCJ9GDG` at seq 1670;
    - after the spec edit, the change-intent id stays `CI-add-auth-xuuno2` and its digest changes;
    - the requirement id changes from `BR-add-auth-core-4ly4j8` to `BR-add-auth-core-5iy98u` with a
      new digest;
    - the obligation id stays `O-ADD-AUTH`.
  - Ran a second overlay test that runs the real `reviews` connector over this repository's committed
    verdicts with a fake appender: 115 listed, 74 wire artifacts, 41 documentary, **0 malformed**, no
    error. So at HEAD the repository's reviews sync processes every artifact, and a refusal here would
    not be hidden behind an earlier parse failure.
  - `git status --short` before and after showed only the two permanent untracked paths. No tracked
    file was rewritten. The embedded Postgres was stopped by the CLI after each command.

## Round-6 closure

| Round-6 item | Verdict | Evidence |
|---|---|---|
| F1 MED: unconditional revisions re-admit the wedge | **Closed** | Under D6 a later revision whose order-only prerequisites are missing is not appended (§4.6), so no review event reaches the ledger ahead of its claim or requirement. `store.Sync.Append` is not transactional and `Finish(ctx, cursor, cause)` only records the error (`store.go:294-305`). So "process the rest, then fail" leaves the other artifacts' appends in place and records the refusal on the `sync_runs` row. §7.3 derives it |
| F2 MED: one-record archive | **Closed** | Executed: the verbatim five-record archive imports into an empty tenant ledger. `sync all` and `rebuild` then succeed. `E` is seq 1670 and the next append is seq 1835, so `E` precedes every verdict (`reduceObligationVerdict`'s "from the future" check passes) |
| F3 LOW: identity and snapshot | **Closed in mechanism; one claim is false (F1 below)** | Identity `(schema, native id)` separates v1 and v2 (`reviews.go`: both kind `review.verdict`). Updating the set on append handles a same-identity pair whose path-first file appends. The reverse order is not handled as §0e claims (F1) |
| F4 LOW: artifact residue | **Closed** | §7.1 matches `validateReference` (`record_kind` `change_intent`, relation `evaluates`, empty `obligation_ids`), the outcome sort key in `ParseObligationVerdict`, and the review's `reviews` relation. §7.2 now writes `gates/x1.yaml` |
| Observation: undecodable git payload | **Closed** | §4.6 makes it a sync failure. See Other observations on decoder strictness |
| Observation: R-INV-5 note dropped | **Closed** for R-INV-5. The neighbouring R-INV-4 and Purpose text are not amended (F2) | `connector/reviews/SPEC.md` lines 7-9 and 123-124 |
| Round-6 observation: a revision that moves to a new requirement revision leaves old SATISFIED rows | **Not answered or recorded**; draft 7's own §7.3 reaches it (F3) | `reduceObligationVerdict` upsert |

## Answers to the attacks

### 1. Round-6 closure

See the table.

### 2. D6 mechanics

- **Identity.** `(schema, native id)` is right for the views D6 protects:
  - v2 rows key on `verdict_id` in `obligation_verdicts_view`;
  - requirement reviews key on `review_id`;
  - v1 never consults the view.

  A v2 artifact carrying a ledgered v1's id is a first revision, so it is held, not refused. That is
  correct, because no v2 outcome exists to keep alive.
- **Set update on append.** If the path-first file of a pair appends, the second file is classified
  identically in both of `verify`'s passes. The reverse order is not (F1).
- **The interface leaves one point open.** It lists only `Ledgered`, `CommitClaimed` and
  `RequirementLedgered`, so "the connector adds an identity to the set" must mean the connector keeps
  its own appended set and ORs it with the view. That is implementable, and row 11 tests it with a
  fake view, but the text could say so.
- **`verify`.**
  - A refusal in pass 1 fails `verify` before the idempotence passes, which is correct and loud.
  - A hold in pass 1 cannot become an append in pass 2. Between the passes `verify` re-runs only
    `intent all` (records and specdir; `sdd` is never selected, `intentProviders`) and `git`, and both
    are already idempotent. So `requireNoSecondSyncAppends` cannot trip on a hold.
- **`sync_runs`.** Finishing with the refusal error records `cursor_json`, `events_appended` and the
  error text. Nothing reads `sync_runs` (`grep`: written only in `store.go`). The draft does not say
  whether a refused path is in the cursor, but the cursor is diagnostic only.
- **`sync all` ordering.** Reviews is the last step (`runCommand`), so a refusal stops nothing after
  it. Intent, git, checks and sessions appends persist, and are correct.
  - Today the counts line is printed only after every step succeeds. §4.6 requires the CLI to print
    the counts, claim-skip, held and refused lines before returning the refusal error. §7.3 step 4
    asserts the claim-skip line, so the test forces that ordering.
- **delivery-enforce.** It runs under `set -euo pipefail`, and `sync all` precedes `gates enforce`, so
  a refusal stops it before enforcement. That is red, as claimed.
- **Can a review artifact still be appended ahead of its prerequisites?** No path was found. The cases:
  - **A first revision, then a same-identity file in the same sync.** The second file is refused (set
    updated).
  - **A refused revision on a later sync.** It is still refused, because the set only grows.
  - **A fresh ledger.** The revision is a first revision and is held, and no old outcome exists there.
  - **Partial recovery** (`sync intent sdd`, then `sync reviews` before `sync git`). The review
    appends, because its requirement is ledgered. The verdict is still refused on its claim, so
    nothing dangles.
  - **The `CommitClaimed` and `RequirementLedgered` snapshots.** A reviews sync appends only reviews
    events, so these snapshots cannot go stale in the unsafe direction. "Present" means the event is
    already at a lower seq.
- **Is the fail-open acceptance honest?** Yes. §3 states that a canary or an enforce run without a
  prior sync reads the old outcome. `make gates-enforce` (`Makefile:24-25`) is exactly such a
  standalone enforce path, and it is covered by that sentence.

### 3. §7.3

- **The sequence runs as described, by execution and code reading.**
  - A spec edit gives a new sdd requirement id (the digest is embedded) and a new change-intent digest
    under the same CI id. The obligation id is unchanged (`O-ADD-AUTH`), so "its new obligation ids"
    means the ids of the new revision, which may be the same strings.
  - At step 4, `L` holds only CI@d1, so X5 is `claim-skip not-ledgered`. The RR revision is refused
    `requirement-not-ledgered`. The OV revision is refused `claim-not-ledgered`, because the claim is
    checked first.
  - Nothing reviews-side is appended, so `rebuild` succeeds.
  - At step 5, the X5 event is re-emitted with its ref (new content, so a new event), and both
    revisions append.
  - The OV revision's rows key on the new `requirement_id`, so they are fresh inserts carrying
    `commit_sha` X5, not an upsert onto X1's rows.
  - The gate query joins X5's claim to CI@d2's targets, SATISFIED `obligation_verdicts_view` rows with
    `commit_sha` X5, and VERIFIABLE `requirement_reviews_view` rows on R'@d2. So x5 `state=PASS` is
    reachable.
- **This only works because the requirement revision changes.** A same-key revision that changes only
  `reviewed_commit` would not move `commit_sha` (F3).
- **Two residues (F2):**
  - The x5 gate is "Add `gates/x5.yaml`" with no fields, though the §7.1 step 5 pattern is evidently
    intended.
  - Step 2 does not say how the R'@d2 and CI@d2 references are computed **without** syncing `sdd`.
    That must be §7.2 step 2's `Revisions` approach. Reading them back from the ledger would ledger
    R'@d2 and break step 4's expected `requirement-not-ledgered` refusal.

### 4. §7.1 and §7.2 against `v2.go` and `migration.go`

- The five-record archive, `E`'s id and seq, and the post-import seq were confirmed by execution.
- `migrate` reads the archive from the working tree (`filepath.Join(root, migration.ArchivePath)`).
- Field order, relations, empty `obligation_ids`, outcome sort keys and the last-occurrence
  self-digest rule (`ArtifactSHA256`) all agree with `v2.go`.

### 5. Sites and laws

- **The reviews SPEC rows miss two sentences that D6 falsifies (F2).**
- **Construction sites are unchanged from round 6.** `cli_test.go:440` calls `syncReviewsOnStore` with
  a nil ledger and expects the store's error. If the CLI view's construction rejects a nil store
  before `beginSync`, that test's expected error changes. This is an implementation note, not a
  finding.
- **Law walk:** below.

## Findings

**F1 (LOW): same-identity classification depends on path order. In the reverse order, the first sync
is green and the next one, or `verify`'s second pass, is red. This contradicts §0e F3's claim that
`verify`'s two passes classify each file the same way, and §4.6's claim that such conflicts "surface
as a refusal".**
- **The case.** Take two files with one identity, `a.md` < `b.md` in path order. `a` lacks its claim or
  requirement; `b` has its prerequisites.
  - **Sync 1:** `a`'s identity is not in the set, so it is **held**. Then `b` appends and the set is
    updated. The exit is 0.
  - **Sync 2, or `verify` pass 2:** `a`'s identity is now ledgered, so `a` is **refused**, and the sync
    and `verify` fail.
- **The harm.**
  - There is no wedge, and the red arrives on the next sync.
  - But for one run a conflicting pair passes green with `b`'s outcome in the ledger, even if `a` is
    the withdrawing copy.
  - The contract's stated stability property is false.
  - Row 11 covers only the forward order ("the first appended and the second … refused").

**F2 (LOW): the site list and residual text are incomplete.**
- **Reviews SPEC text that D6 makes false is not listed for amendment:**
  - the Purpose sentence "each valid schema-bearing artifact emits one … event … changed content a
    revision". Draft 6 amended it only for a held first revision;
  - **R-INV-4**: "changed content with the same verdict ID produces a new event". A refused revision is
    changed content that produces no event.

  §5 lists only R-INV-5 and R-INV-13.
- **§4.6 is silent on three points:**
  - whether a refused path enters the cursor;
  - what the CLI prints, and what `Sync` returns, when a malformed artifact follows a refusal. Today
    `Sync` returns at the first malformed file, and the CLI prints nothing on error;
  - whether refused paths go through `printableUnit` as held paths do. Exposure is limited:
    `committedVerdictReader` lists without `-z`, so a path containing a control character is quoted by
    git and fails the read.
- **§7.3 under-specifies two steps:** it does not define `gates/x5.yaml`, and it does not say that
  step 2 must compute the references without syncing `sdd`.
- **Why LOW.** Nothing here is fail-open or a wedge, and the calibration rule would surface most of it.
  But the R-INV-4 and Purpose sentences would ship contradicting the code (Law 2).

**F3 (LOW): D6's rationale ("the withdrawn outcome cannot pass silently", §0e F1) is stated more
broadly than the system supports, and draft 7's own §7.3 falsifies it. The round-6 observation was
dropped without being recorded.**
- **The mechanism.** `obligation_verdicts_view`'s upsert is `ON CONFLICT(verdict_id,
  requirement_source, requirement_id, requirement_artifact_sha256, obligation_id) DO UPDATE SET
  outcome, evidence_event_ids, event_id, seq`. It never updates `commit_sha` or the intent columns.
- **In §7.3, after a correct recovery,** `OV-m3-x1` says X5. Its X1 rows, under the old requirement
  key, stay SATISFIED, so x1 keeps reading PASS with no output at all.
- **For a same-key revision that changes only `reviewed_commit`** (spec unchanged, so the same CI and
  requirement revision):
  - the new outcome overwrites the old commit's row while `commit_sha` stays the old commit;
  - a NOT_SATISFIED rev1 for X1 followed by a SATISFIED rev2 for X7 makes **X1** PASS, and X7 can
    never PASS;
  - that is a silent fail-open on correct, in-scope data.
- **A revert does the same.** Store dedupe on `(source, native_id, content_sha)` means reverting a file
  to rev1's exact bytes is "existing", so rev2's outcome stays in the views silently.
- **Scope.** All three are pre-existing: M3 neither causes nor worsens them, and records claims reach
  them today. But M3 makes them reachable for `sdd` commits. The draft's D6 justification claims a
  property that only refusal-induced staleness has. Round 6 raised the first case for the owner, and
  draft 7 neither answers it nor puts it in §9.

## Other observations (no finding)

- **Decoder strictness.** §4.6 says a git payload that "does not decode as `connectorgit.Commit`" fails
  the sync, justified by "projections could not reduce it either". That holds only if the view's
  decoder is no stricter than projections' `decode`, which applies `DisallowUnknownFields` plus
  `requireFields`. A plain `json.Unmarshal` into `Commit` is no stricter, so the argument stands. The
  draft could name the decoder.
- **Permanent red, by design.** A refusal whose prerequisite will never be ledgered (a digest typo, a
  claim on a permanently skipped commit) makes `sync all`, `verify` and delivery-enforce red until the
  artifact is fixed. The same typo in a first revision is only a held line. The asymmetry is the
  intended D6 behaviour, but §3's accepted cost names only the recoverable case.
- **Full history.** `sync intent sdd` walks full history (`intentProviders`), so intermediate
  requirement revisions are always ledgerable. No correct-data refusal is permanent on that account.
- **Malformed artifacts at HEAD.** This repository's reviews sync currently has 0 malformed artifacts
  (executed). The project memory note that delivery-enforce is red on malformed v1 verdicts is out of
  date for this path.

## Law walk

- **Law 1.** No generated artifact is hand-edited. The regeneration pins are carried from draft 5.
  Compliant.
- **Law 2.** No second decoder. The Purpose and R-INV-4 sentences would contradict the code (F2).
- **Law 3.** Not engaged: no hooks change.
- **Law 4.** No advisory tier is added. A refusal is blocking. Compliant.
- **Law 6.** M3-INV-12 is now true and derived: no wedge path found, and §7.3 runs as described. The
  two residues are F1's false stability claim and F3's over-broad rationale. The calibration rule is
  restated.
- **Law 7.** The wedge lesson is compiled for both revision kinds (M3-INV-11 and M3-INV-12, with rows
  11 and 12). Compliant.
- **Law 8.** No new dependency. Compliant.
- **Law 9.** This is a non-author re-check. Round 6's prescribed-but-unrun control (the archive) was
  executed this round and passes.
- **Law 10.** Not engaged by a plan document. HEAD's only move was a `notes/state.md` update.

## Verdict

**ACCEPTABLE.**
- **What D6 closes, as specified:**
  - it closes round-6 F1's wedge without reintroducing round-5's fail-open. A later revision is never
    ledgered ahead of its order-only prerequisites, and every sync-first enforcement path is red while
    one is refused;
  - round-6 F2 is closed, verified by execution;
  - F3 and F4 are closed in mechanism.
- **No reachable permanent projection wedge from correct in-scope data was found,** and no new silent
  fail-open was found.
- **Three LOW findings remain:**
  - path-order dependence of the conflict classification (F1);
  - incomplete reviews-SPEC sites and §7.3 residues (F2);
  - an over-broad "cannot pass silently" rationale over pre-existing upsert, revert and stale-row
    behaviour that draft 7's own §7.3 reaches (F3).

  These can be fixed in the SPEC text or recorded by the owner without another adversarial round.
