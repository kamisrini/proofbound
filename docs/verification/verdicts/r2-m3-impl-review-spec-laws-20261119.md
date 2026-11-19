---
schema: vera.verdict.v1
verdict_id: r2-m3-impl-review-spec-laws-20261119
status: NEEDS_WORK
reviewed_commit: 2d67fc03c1900d9df7de57a841401d9a9d74c6fe
findings:
  - finding_id: r2-m3-impl-review-spec-laws-20261119-F1
    severity: MED
  - finding_id: r2-m3-impl-review-spec-laws-20261119-F2
    severity: LOW
  - finding_id: r2-m3-impl-review-spec-laws-20261119-F3
    severity: LOW
  - finding_id: r2-m3-impl-review-spec-laws-20261119-F4
    severity: LOW
  - finding_id: r2-m3-impl-review-spec-laws-20261119-F5
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-impl-review-spec-laws-20261119.md
artifact_sha: a7d0cb14a3da6132ae49e94d0ea4732eb3979fc144e139c18da4036c48fe9364
---

# R2 M3 implementation review: spec truth and Build Laws

## Provenance

- **Reviewer:** an independent, non-author agent session (Build Law 9). I wrote none of the code or
  SPECs under review. The lens was spec truth and Build Laws. Mutation adequacy and concurrency
  belong to a second reviewer.
- **HEAD:** `2d67fc03c1900d9df7de57a841401d9a9d74c6fe` at both start and end. It did not move.
- **Working tree:** at both start and end, the only entries were the two pre-existing untracked
  paths (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`). No tracked file was
  modified.
- **Range reviewed:** `9600441..2d67fc0`, i.e. `git log 260b563..2d67fc0` and
  `git diff 260b563 2d67fc0 -- kernel scripts`. Production code was read at the reviewed commit
  with `git show 2d67fc0:<path>`.
- **Read:**
  - CI `CI-sdd-warranty-m3-xaqocp/0001.md`;
  - M3 SPEC drafts 5, 6, 7 and 8;
  - diffs of the git, cli, projections, reviews and sdd SPECs;
  - `connector.go`, `cli.go`, `reviews.go`, `projections.go` (the commit reducer,
    `reduceObligationVerdict`, `reduceRequirementReview`), `gates.go` (the readiness query), and sdd
    `build.go` / `provider.go` (`Resolve`, `mapState`);
  - every M3 test file named in the brief;
  - `notes/state.md` and `notes/journal/2026-11-19.md`.
- **Executed** (with `$HOME/go/bin` first on PATH, GOCACHE under `<scratch>/impl-a`, `-count=1`, cli
  and sdd run one after the other):
  - `go test` on `connector/git/...`, `connector/reviews`, `projections`, `cli` (184.5s) and
    `connector/intent/sdd` (216.5s): all `ok`;
  - `scripts/gen-invariants-lock.sh` with `INVARIANTS_LOCK_OUTPUT=<scratch>/impl-a/inv.lock`:
    byte-identical to `docs/invariants.lock`;
  - `scripts/p6-route-matrix.sh --check` and `scripts/p6-task2-close.sh --check`: exit 0;
  - `scripts/state-freshness.sh`, `invariant-lint.sh`, `invariant-table-lint.sh`,
    `spec-numbering-lint.sh`, `law-citation-lint.sh`: exit 0;
  - `scripts/tests/intent-delivery-applicability.test.sh`: ok;
  - the two exclusion checks: `git diff --quiet 77cc6ba HEAD` over sdd production `*.go` exits 0,
    and over `scripts/intent-delivery-applicability.sh` exits 0.
- **Not run:** I did not run `make check` as a whole. I ran its relevant components individually,
  listed above.

## Answers to the attacks

### 1. Clause-to-code coverage

The code implements the contract faithfully. I found no divergence in behaviour.

**Git connector (`connector.go`):**
- `sddValueRE` is exactly D1's grammar.
- `isSDDClaim` folds only bytes 0-2 to ASCII lower case and requires byte 3 to be exactly `:`. That
  matches "first four bytes ASCII-case-fold to `sdd:`".
- A value that fails the grammar becomes a `malformed` skip.
- `ClaimUnboundError` unwraps to `ErrClaimUnbound`. An empty `Cause` fails the sync. The amended git
  SPEC §2 states this, so the code has a clause.
- The spoof re-check is `Provider=="sdd"`, a CI-grammar id and a 64-hex digest.
- All-or-nothing: when any skip occurs, `commit.IntentRefs = nil`.
- `records` / `specdir` trailers sort before `sdd:` and keep their whole-sync failure.
- The error message keeps the pre-normalisation sha.

**CLI resolver:**
- The `sdd` provider is built lazily (`newSDDClaimProvider`), and `L` is read lazily, after the §4.2
  checks.
- The cause table maps the three sentinels plus `UnmappableError` to `unmappable:<cause>`.
- The segment check is `path.Base(ArtifactPath)` equal to `name`, or `^YYYY-MM-DD-` followed by
  `name`.
- `L` is keyed on the payload's `intent_id` plus digest. The intent connector enforces
  `intent_id == NativeID` (`intent.go:301`), so this equals the SPEC's "(native id, digest)".

**`storeLedgerView`:**
- One empty-filter pass builds three sets.
- Git payloads are decoded as `connectorgit.Commit`, and a decode error fails the pass.
- A reviews payload with no `schema` fails the pass.
- Requirements come only from `intent.*` sources.
- It is loaded lazily, and only on the first query.

**Reviews connector:**
- `duplicateIdentities` runs before the main loop. Duplicates are refused without consulting
  prerequisites.
- Prerequisites are checked in order: claim, then each requirement.
- `Ledgered` is consulted only when a prerequisite is missing, which yields held versus refused.
- Neither held nor refused paths enter the cursor.
- `fail()` uses `errors.Join` of a malformed or view error with `refusedError`.
- After the loop, a refusal returns an error wrapping `ErrRevisionRefused`.
- v1 never touches the view.

**CLI output:**
- `sync reviews` and `sync all` print the counts line, then (for `sync all`) the claim-skip lines,
  then held, then refused lines before returning the reviews error. Values and paths go through
  `printableUnit`.
- `verify` calls neither printer.
- **Projections:** the provider set is widened to `sdd`, and nothing else in projections changed.

**Code with no clause:** none found beyond the empty-cause rule, which the git SPEC covers.

**Scope wording:** the CI's Allowed scope says "CLI-INV-7 and CLI-INV-8 amendments". The
implementation left those rows untouched and appended CLI-INV-13..16 instead. The substance is present
(see Other observations).

### 2. Do the proving tests discriminate?

| Item | Test(s) | Status |
|---|---|---|
| Row 1 | git `TestSync_SDDClaimBindsResolvedRevision`; sdd `TestM3ClaimBindsOnlyALedgeredRevision` | Proved |
| Row 2 | git `TestSync_SDDSpoofedRevisionFailsSync`; cli `TestSDDResolverChecksAndCauses` (one case per clause, including undated and dated-other segments) | Proved |
| Row 3 | git `TestSync_SDDTenantOutcomesAreClaimSkips`: every cause, `SDD:`/`Sdd:`/`sdd:Add-auth` malformed beside a resolving `sdd:add-auth`, and two-claim / mixed commits | Proved |
| Row 4 | git `TestSync_SDDResolverErrorsFailSync` | Proved |
| Row 5 | sdd §7.2 interleaving A and §7.3 (`TestM3InterleavingAThenRefusal`), interleaving B (`TestM3InterleavingB`) | Proved |
| Row 6 | cli `TestSDDClaimWorkIsLazy` | Proved; also sees eager construction inside `syncGit`, via the nil-ledger call |
| Row 7 | existing tests | Unchanged |
| Row 8 | sdd `TestM3ReadinessGateOnOneLedger` | Proved: x1 PASS, x2 and x3 BLOCKED, X3's verdict held |
| Row 9 | the corpus rows plus the oracle now meaning "a ref is bound" | Proved |
| Row 10 | — | **Absent (F1)** |
| Row 11 | reviews `hold_test.go` | Proved |
| Row 12 | cli `ledger_view_test.go` | Proved, except the draft-7 "undecodable git payload fails the view" case (F2) |

Further notes:
- **Check order.** Interleaving A's `sync all` holds the verdict as `claim-not-ledgered` while its
  requirement is also not ledgered. That discriminates the order of the prerequisite checks, which
  the unit cases do not.
- **Row 11 detail.** It covers each missing prerequisite (including a second requirement), revision
  refusal with a later artifact still appended, a v2 sharing a v1's id treated as a first revision,
  duplicates in both orders, malformed-after-refusal, v1 never consulting the view, and a neutral
  control.

### 3. Permanent wedge or silent fail-open from correct in-scope data

None found.
- **`sdd` git refs.** These are appended only when they are in `L`, which is read before any git
  append in that sync. So the commit reducer's dangling check passes.
- **v2 verdicts.** They are appended only when the exact `(sha, provider, id, digest)` ref is in some
  ledgered git event, and every cited requirement revision is ledgered. Items 1-3 of the reducer are
  therefore present at a lower seq.
- **Requirement reviews.** They are likewise gated on their requirement.
- **Evidence.** It is excluded by design. It is pre-existing and deferred.
- **`commit_intents_view`.** It is insert-or-update only; no statement deletes from it. The readiness
  gate reads `commit_intents_view`, not `commits_view.intent_refs`, so monotonicity holds in code
  today. It is untested, though (F1).
- **Fail-open.** A reviews payload with no `schema` cannot come from this connector:
  `Verdict.Schema` has existed since the connector's first commit, `2bc52b7`. So the identity set
  cannot silently miss a revision from correct data.

### 4. Scope and exclusions

- `sdd` production code is unchanged since `77cc6ba`, and `scripts/intent-delivery-applicability.sh`
  is unchanged.
- The changed paths match the allowed scope: git, cli, reviews and projections code and SPECs; the
  trailer fixtures and oracle; the sdd SPEC; the three pins; tests; and notes.
- The `docs/plans/R2-sdd-provider-spec-draft10.md` edit is the byte-identical mirror of the sdd
  SPEC, as the notes state.
- No ledgering of skips, holds or refusals was added. No BR status changed.

### 5. Law walk

See the Law walk section below.

### 6. Tests

All the touched packages pass: git, gitcmd, reviews, projections, cli and sdd.

## Findings

### F1 (MED): M3 SPEC §7 row 10 (M3-INV-10, monotone bindings) has no proving test

**What the contract requires.**
- Draft 5 §7 row 10 derives an untagged embedded-store test with this sequence: a change intent; a
  git event binding it; a second git event for the same commit *without* the ref; a v2 verdict
  citing the binding. Apply and Rebuild must then both succeed with the row present.
- Its "kills" cell is "deleting a commit's rows on a newer commit event".
- The CI's Allowed scope says "every test the M3 SPEC §7 derives (rows 1-12)". Its Acceptance
  boundary requires "the M3 SPEC §7 proving tests pass".

**What exists.**
- No such test is in the tree: `git grep` at HEAD for `M3-INV-10`, `monoton` and `row 10` finds
  only SPEC prose.
- No M3 test produces a newer git event without a ref for an already-bound sha. In the M3 tests the
  order is always no-ref first, then ref.
- The only projections tests that touch `commit_intents_view` (`intent_test.go`,
  `projection_test.go`) are `//go:build integration`, so they are outside `make check`.
- `notes/state.md` nevertheless records M3 as "Code complete" and lists no gap.

**Harm.**
- M3-INV-10 is the premise of the accepted cross-version limitation now written into the git SPEC
  and into sdd SPEC §14. Bindings are monotone because the reducer never deletes them.
- **Failure scenario:** a future change makes the commit reducer replace a commit's
  `commit_intents_view` rows on a newer event, for example a "clean-up" that deletes and re-inserts.
  Every M3 test and every untagged projections test would still pass.
- After that change, on a ledger where a later event for a bound commit lacks the ref, a verdict
  that projected once would fail Apply with "does not match an exact commit intent claim". That
  wedges the whole projection permanently from data that was correct when written.
- That is exactly the property row 10 was derived to pin, and the acceptance boundary is not met as
  stated.

### F2 (LOW): Row 12's "undecodable git payload fails the view" case (draft 7) is not tested, nor is the schema-less reviews payload path

**What the contract requires.** Draft 7 §7 "12 (view)" adds "a ledgered git payload that does not
decode fails the view", with the kill "skipping undecodable payloads". Draft 7 §4.6 "View errors"
also lists "a `reviews` payload without a `schema` string".

**What exists.** `readLedgerView` implements both error paths. No test exercises either: `git grep`
over `*_test.go` finds no such case.

**Failure scenario.**
- A mutant that skips a schema-less reviews event is not caught by any test.
- With that mutant, the identity set loses that event, so a later revision of it is classified as a
  first revision and held quietly instead of refused. That is the round-6 fail-open class.
- Reachability from correct data is nil (every connector-produced reviews payload carries `schema`),
  hence LOW.

### F3 (LOW): Draft 8's CLI half of "malformed after a refusal" is proved only at the connector level

**What the contract requires.** Draft 8 §7 row 11 states that for a malformed artifact after a
refusal, "the error joins both, and the CLI still prints the refused line".

**What exists.**
- `TestRefusalJoinsALaterMalformedError` (reviews) proves the join.
- `TestSyncReviewsPrintsHoldsAndFailsOnRefusal` (cli) proves printing before a *refusal* error only.
- No CLI test drives a malformed artifact after a refusal.
- The code (`runCommand` prints before `errors.Join(syncErr, …)`) is correct today.

**Failure scenario.** A refactor that prints only when `errors.Is(err, ErrRevisionRefused)` is the
sole error would drop the refused lines on that path, and nothing would catch it.

### F4 (LOW): Law 2, the CI-id and digest grammars gain another home, and intent payloads gain ad-hoc decoders

**Evidence.**
- `cli.go` declares `sddIntentIDRE` and `sddDigestRE` beside the git connector's `intentIDRE` /
  `digestRE`, projections' inline regexp, and the intent connector's `recordIDRE`.
- `readSDDLedgered` and `readLedgerView` decode intent change-intent and requirement payloads with
  anonymous structs, although `connectorintent` exports the producer types (`intent.go:77-96`).
  Draft 6 F6 resolved "no second decoder" only for git payloads.

**Harm.** These are drift seeds. If the CI grammar or a payload field name changes in its home, the
CLI copy silently diverges. Divergence here fails closed: either a sync failure, or a claim or
verdict that never binds or appends. So the harm is a misleading hold or skip, not a false PASS.

### F5 (LOW): The compiled Law 7 response (`port_home_test.go`) is real, but narrower than its LESSON line claims

**Evidence.**
- The detector's `store\.Config\{[^}\n]*\}` pattern misses a multi-line `store.Config{…}` literal,
  and it accepts any literal containing `Port:` (including `Port: 0`).
- In package `cli`, a bare `runCommand(..., "", ...)` that does not stub `openStore` opens the
  default port undetected. The test's own comment acknowledges this.
- The in-test calibration is only an inertness guard (`homeHits >= 2`), a positive control. The
  "seven probes" the journal line cites are not in the test, so they cannot be re-run.

**Failure scenario.** A future untagged cli test written with a multi-line `store.Config{}` without
`Port`, or an unstubbed `runCommand`, reintroduces the 2026-11-09 and 2026-11-19 port-contention
flake, and the compiled lesson stays green.

**Current state.** I checked every current untagged cli test that calls `runCommand`. Each one stubs
`openStore` or uses an ephemeral port.

## Other observations

- **CLI SPEC rows.** Draft 5 §5 and the CI name "CLI-INV-7 / CLI-INV-8" as the amended rows. The
  implementation appended CLI-INV-13..16 and left CLI-INV-7 ("select only records, specdir, or both")
  and CLI-INV-8 unedited. The new rows carry the substance. CLI-INV-7's wording was already stale
  since M1's `sync intent sdd`; that is not M3's doing.
- **Behaviour change in `sync reviews` / `sync all`.** On *any* reviews error, including malformed
  input, a counts line is now printed before the error. This is per draft 8, but it is visible to any
  script that treated "no stdout" as failure.
- **`verify` on a fresh ledger.** `verify` never syncs `sdd`, so on a fresh ledger every `sdd` claim
  is `not-ledgered` and every `sdd` verdict is held. That is consistent with SDD-INV-56 and the
  accepted cost, but `verify` therefore never exercises an `sdd` binding.
- **Commit trailers.** The M3 commits carry `Intent: CI-sdd-warranty-m3-xaqocp` and an authorisation
  trailer.

## Law walk

| Law | Result | Evidence |
|---|---|---|
| 1 | Pass | The lock regenerates byte-identically. The route-matrix and connector-reality pins pass `--check`. Both pin files carry `@generated` markers. |
| 2 | LOW | F4. Git payloads are decoded with the producer type, as draft 6 required. |
| 3 | No new hooks | The trailer-gate corpus self-test passes with the three new `sdd:` rows. |
| 4 | Not applicable | No advisory rows were added. |
| 5 | Not applicable | — |
| 6 | Pass on ordering | The M3 SPEC was accepted (round 7, `fe8d116`, an ancestor of `9600441`) before code. Package SPEC amendments land in the same commits as their code. The SPEC-derived test set is incomplete (F1, F2, F3). |
| 7 | Pass, with F5 | The LESSON is compiled into an enforcing untagged test that runs in `make check`. |
| 8 | Pass | No new dependency (`path` and `regexp` are stdlib). |
| 9 | Pass | Author-side calibration claims, such as "N mutants killed", are self-reported. This review is the non-author check, and mutation is the other reviewer's lens. |
| 10 | Pass | `notes/state.md` was touched at HEAD and `state-freshness.sh` passes. Its "Code complete" claim is contradicted by F1. |

## Verdict

**NEEDS_WORK.** The production code matches the contract (draft 5 §4.1-4.5 as amended, and draft 8
§4.6) clause for clause. I found no reachable wedge or fail-open from correct data, the exclusions
hold, and the generated pins are fresh.

One MED stands: the M3 SPEC's row 10 proving test for M3-INV-10 (monotone bindings), which the CI's
acceptance boundary requires, does not exist. Nothing in `make check` guards the reducer property
that the accepted cross-version limitation rests on.

There are four LOW findings: F2 and F3 are missing SPEC-derived cases, F4 is a Law 2 drift seed, and
F5 is detector coverage of the compiled lesson.
