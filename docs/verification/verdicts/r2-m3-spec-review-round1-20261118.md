---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round1-20261118
status: NEEDS_WORK
reviewed_commit: 5ab705931dd48225764a95e7e209421a6feb9d0e
findings:
  - finding_id: r2-m3-spec-review-round1-20261118-F1
    severity: HIGH
  - finding_id: r2-m3-spec-review-round1-20261118-F2
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F3
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F4
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F5
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F6
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F7
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F8
    severity: MED
  - finding_id: r2-m3-spec-review-round1-20261118-F9
    severity: LOW
  - finding_id: r2-m3-spec-review-round1-20261118-F10
    severity: LOW
  - finding_id: r2-m3-spec-review-round1-20261118-F11
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round1-20261118.md
artifact_sha: b1126d98a2d275b7d7d631c323eb8ac9c43dece97abae2c13c446a6029ef3199
---

# R2 M3 git-binding SPEC draft 1: non-author review, round 1

**Format note (2026-11-19, owner-approved):** the front-matter `findings` list was converted from flow style (`- {finding_id: …, severity: …}`), which the reviews connector's v1 parser rejects, to block style. The ids, severities and body are unchanged; `artifact_sha` was recomputed.

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**: it did not write the draft, the sdd SPEC, or any of the code it
  reads, and it has no authoring context from the session that wrote them (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft1.md` at HEAD
  `5ab705931dd48225764a95e7e209421a6feb9d0e`.
- **HEAD moved during the review.** It was `5ab7059` at the start and `a791d2c`
  ("notes: M3 SPEC draft 1 under non-author review") at the end. `git diff --stat 5ab7059 a791d2c`
  shows that the only change is `notes/state.md`, which the owning session wrote. The draft and all
  the code read are unchanged, so this verdict applies to `5ab7059`.
- **`git status` was the same at the start and at the end:** only the two permanent untracked paths
  (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`).
- **Read at HEAD:** `connector/git/connector.go` and its `SPEC.md` (§4 INV-1..5, INV-28..30, §5.1,
  §6); `cli/cli.go` (`syncGit`, `commitIntentResolver`, `intentProviders`, the `sync git`/`sync all`
  dispatch, `verify`); `cli/SPEC.md`; `projections/projections.go` (`Apply`, `Rebuild`, `reduce`,
  `reduceIntent`, `reduceObligationVerdict`, `commitPayload.validate`); `gates/gates.go` (the
  readiness query and `latestIntentProof`); `connector/intent/sdd/{SPEC.md, build.go, provider.go,
  ids.go, profile.go}`; `connector/intent/records/records.go`; `connector/intent/specdir/specdir.go`;
  `store/store.go` (`BeginSync`, `Append`); `scripts/intent-delivery-applicability.sh`;
  `scripts/p6-route-matrix.sh`; `scripts/p6-task2-close.sh`; `docs/plans/p6-census-rows.tsv`;
  `scripts/tests/fixtures/intent-trailers/cases.tsv`; `Makefile`; and
  `docs/intent/records/requirements/BR-sdd-warranty-i9miv7/0003.md` (O-8, O-10).
- **Executed:** `go test ./internal/connector/git/... ./internal/connector/intent/sdd/ -count=1`, which
  passed (baseline only). I also ran a scratch probe under `<scratch>/reviewer-m3d1/slugprobe/`, a
  verbatim copy of `sdd/ids.go` with a driver, for Q5. I did not run Postgres or integration tests
  (Q1 below says why).
- **Nothing in the repository was modified.**

## Answers

### Q1: is the permanent wedge real?

**Yes. Confirmed by reading the code, not by execution.**

- `projections.go:218-226`: for every `IntentRef` on a `git/commit.recorded` event, `reduce` queries
  `change_intents_view` for `(intent.<provider>, id, digest)` and returns
  `dangling intent revision` if no row exists.
- `Apply` (`projections.go:147-179`) reads every record after the checkpoint in sequence order and
  reduces them all inside **one** transaction. Any error aborts the transaction, so the checkpoint
  (`projection_meta.last_seq`) never moves past the dangling event. `Rebuild` (`:189-197`) drops the
  views and calls the same `Apply`.
- The ledger is append-only, and `Append` is idempotent on `(source, native_id, content_sha)`
  (`store.go:261`). A later sync that ledgers the missing change intent gives it a **higher**
  sequence number, so the commit event is still reduced first and still fails.

So the wedge is not limited to rebuilds. Incremental `Apply` also fails, and so does everything that
projects: `report`, `verify`'s projection step, and the gates (which report "projection failed" as
BLOCKED). No sync can repair it.

**Reproduction.** I did not reproduce it by execution. Every path through `reduce` needs the
Postgres-backed store, the projections suite is `//go:build integration`, and the brief notes that
the suite fails at setup. No existing test asserts the commit-ref dangling error: `grep 'dangling'`
over `*_test.go` finds only the target and exact-record cases in `intent_test.go`.

**The §8 "suspected records ordering hazard": confirmed, and broader than §8 states.** See F8.
- `records.Revisions` reads `HEAD` only (`records.go:48`), while `records.Resolve` reads the commit's
  own tree (`records.go:58`). `specdir` has the same split (`specdir.go:51` and `:61`).
- `sync git` alone never ledgers intents (`cli.go:660-668`). So a commit citing a records CI that has
  not been through `sync intent` dangles on the next projection, permanently.
- `sync all` does not close the hazard either. Its intent step ledgers only the HEAD tree's
  revisions, while the git connector walks every ref. A commit on a non-HEAD branch citing a CI
  revision that exists only on that branch, or whose bytes at that commit differ from HEAD's, binds a
  digest that no sync ever ledgers.

### Q2: is D3's lazy ordering sufficient?

**No (F2, F3).** The draft builds on its §0 row 11: "`Resolve(c,n)` returns only revisions that
`Revisions` emits". SDD-INV-31 says something narrower: "for every commit `c` that **`Commits` lists**".

Under D3, the two listings happen at different times:
- The git connector lists commits first (`connector.go:112`).
- The sdd intent sync lists again later, when the first `sdd` claim triggers it.

Two orderings matter:
- **Refs advance between the walks.** sdd sees a superset, which is harmless. This also answers the
  question of a revision that only appears at a later commit: one that appears after sdd's snapshot
  is never one the git walk's commits can name, because the git walk listed earlier.
- **Refs move backwards between the walks** (branch deletion, force-push, `reset`). Commit `c` was
  listed by git but is not in sdd's walk. `Resolve` still succeeds, because it reads `c` by SHA
  (`build.go:172`) and nothing checks reachability. If `c`'s change-directory state is unique to
  `c` (an amended commit), its `(id, digest)` is never ledgered. The commit event is then appended
  with a dangling ref, which is the permanent wedge of Q1.

The other cases are safe:
- **Idempotency** holds: the intent sync is idempotent on `content_sha`.
- **Cancellation midway** is safe. A cancelled or failed sdd sync is a non-sentinel error, so the git
  sync fails before any commit carrying an sdd ref is appended, and the partial sdd appends are
  harmless.
- **Nested sync runs** are safe. `BeginSync` (`store.go:238`) only inserts a `sync_runs` row, with
  no lock, so an `intent.sdd` run inside an open `git` run does not conflict.

**A simpler correct design exists.**
- In the resolver, after a successful `Resolve`, require that the exact intent revision is already
  in the ledger. If it is missing, fail the sync, which is a retryable failure, never a dangling
  append. This makes the wedge impossible whatever the race, the provider, or the sync order. It
  also closes the confirmed records/specdir hazard (F8) with one mechanism.
- It needs no full-history walk inside `sync git`, so it also avoids F3.
- Appending only the change-intent revision that `Resolve` returned is **not** enough on its own.
  `reduceIntent` also requires every target requirement revision and active obligation to exist
  (`projections.go:357-373`).

### Q3: does D2 preserve the git connector's invariants and determinism?

**Within one code version, yes.** `Resolve` reads only `c`'s state and its parents' archive listings
(SDD-INV-33, `build.go:168-215`). Both are content-addressed, so the outcome of a claim at `c` does
not depend on ref movement, and the payload (which carries `intent_refs`, but never
`IntentTrailers`, `json:"-"`) is deterministic. INV-2, INV-3 and INV-5 hold.

**No already-appended event changes meaning.**
- Before M3, any `sdd:` trailer fails the sync at the allowlist (`connector.go:198`), so no ledger
  holds a commit event with an `sdd` claim.
- Widening `commitPayload.validate` (`projections.go:727`) only accepts more.

**Two gaps remain:**
- **Commit-level trust (F4).** A skipped claim leaves no trace in the ledger. A commit with one
  skipped claim and one bound claim passes the readiness gate on the bound claim alone.
- **Across code versions (F11).** A later binary that classifies a claim differently gives the same
  commit a different `content_sha`, which appends a second event for that SHA.
  `commit_intents_view` keeps rows from both events (`ON CONFLICT ... DO UPDATE`, never deleted,
  `projections.go:224`), so a binding a later version withdraws persists.

### Q4: is the anti-spoofing sound?

**What `Resolve` itself guarantees.** It looks the name up exactly:
- `st.changes[name]` (`build.go:192`);
- archive capture equality (`build.go:205`, `datedRE`);
- a reused name is `name-reuse`, which is unmappable.

So a tenant cannot make `Resolve(c, A)` return change B's revision. The CI id is derived from the
name (`ids.go:72-78`).

**Cross-state id collisions.** A tenant can grind a name that shares another change's slug (only the
first 40 characters count) and its 6-character suffix (about 2³¹ work). The two changes then share
one CI id. M1 accepts this: "two revisions of one id", sdd SPEC §4.3. Bindings, verdicts and the
readiness join are digest-exact (`gates.go:285`), so this cannot cross-satisfy a gate. Only id-keyed
views would merge the two changes. This is an observation, not an M3 finding.

**The check §4.2 adds** (the artifact path names the claimed change, active or archived, for the
profile in force) **cannot be built inside the draft's own boundary (F5).** An archive-path match
must also be anchored: an unanchored suffix match would accept `<archive>/2024-01-01-2024-02-02-foo`
for the claim `foo`, although that entry names `2024-02-02-foo` (sdd SPEC §6). The draft's prose does
not say "anchored".

**`kit`:** the claim is true. `profile.go:128` admits only `"kit": "openspec"`, and
`TestM1IsOpenSpecOnly` exists (`sdd_test.go:1000`).

**The connector-side check** for `sdd` (CI grammar plus a 64-hex digest) ties the returned id to
nothing the trailer said. Every name-to-revision guarantee therefore rests on the resolver.

### Q5: does the trailer grammar agree with the name domain and slug rules?

**Yes, a name can match D1 and still fail `changeIntentID`.** The probe printed:

```
"a-"   D1=true id="" err=slug-too-short
"0--"  D1=true id="" err=slug-too-short
"z---" D1=true id="" err=slug-too-short
```

The reason is that `asciiMap` trims trailing `-` (`ids.go:37`), and then `len < 2` fails
(`ids.go:57`).

Under D2 such a claim gives `ClaimSkip unmappable:slug-too-short`, so it is safe. The draft does not
say so, and its gate script would accept such a trailer although it can never bind (F9).

The opposite direction is the larger problem. Names in the sdd name domain but **outside** D1 (upper
case, `_`, `.`, more than 100 bytes) count as malformed under §4.1, and a malformed claim fails the
whole sync (F1).

### Q6: is the contract complete?

**Sites.** The site list is incomplete (F7). `grep -rln specdir kernel scripts` finds no further
**provider allowlist** beyond the four that §12 names. The sites the draft misses are:
- `cli/SPEC.md` CLI-INV-7 and CLI-INV-8. CLI-INV-8 says "Full sync orders all intent providers before
  commit ingestion", which D3 changes.
- The git SPEC interface lock. The `Result` struct is pinned at git SPEC §2 "Interface", lines 150-156, and the
  draft amends only INV-29.
- The shared trailer corpus `scripts/tests/fixtures/intent-trailers/cases.tsv`, which
  `gitcmd_test.go` and the gate's test both consume. The gate rule changes (M3-INV-9).
- The sdd SPEC's own §12 "Owned by M3" lines, §14, and SDD-INV-56, if D3 is adopted.
- Two `@generated` pins:
  - `p6-route-matrix.sh` pins `code_commit` over `projections`, `cli`, `twin` and `gates`;
  - `p6-task2-close.sh` pins it over `kernel/internal/connector` and `cli/cli.go`.

  M3 touches all of those paths. No route row changes: M3 adds no source or kind.

**The §5.1 pinned vector** carries no intent refs (`IntentRefs: nil`), so M3-INV-7's "vector unchanged"
holds trivially. It does not pin a commit with an `sdd` ref.

**Invariant testability.** The invariants are not paired with tests (F6):
- The draft has no test-derivation table, no named tests, no kill mutants, and no calibration
  controls.
- M3-INV-5 (rebuild) and M3-INV-8 (gate) need the Postgres path, which `make check` does not run:
  `Makefile:27` runs `go test ./... -short` with no `-tags integration`.
- M3-INV-5 as stated (a fresh store and a static repository) cannot fail under the F2 race.
- M3-INV-6's parenthetical ("shallow clone: `sync git` behaves as before M3") is vacuous. INV-20
  already refuses a shallow repository before any resolver runs.

### Q7: Law walk

- **Law 1:** the two `@generated` pins above go stale with M3 and are not listed as sites. They fail
  closed at `make check`, so the cost is churn, not silent drift (part of F7).
- **Law 2:**
  - Building §4.2's profile-relative check in the CLI would mean re-parsing the tenant profile
    outside `sdd`, giving profile semantics a second home (F5).
  - §0 row 9 (147 names) is a count with no home or command; the parent contract says 135 active and
    108 archived (F10).
- **Law 6:** the draft is a SPEC ahead of code, which is correct. It lacks an interface lock for the
  new connector/resolver surface and lacks test derivations (F6).
- **Law 7:** the permanent-wedge class has now appeared in sdd SPEC §7, in §14 (twice: non-UTF-8
  paths and U+0000), in this draft's §4.3, and in the confirmed records/specdir hazard (F8). The
  draft handles one instance locally and leaves the confirmed sibling unowned. It does not compile the
  class into an enforcement point. The ledger-membership check in Q2 would be one.
- **Law 8:** no new dependency. Compliant.
- **Law 9:** the draft states it is unreviewed and needs non-author review. Compliant.

## Findings

**F1 (HIGH): a single malformed `sdd:` trailer anywhere in a tenant's history permanently wedges
every git sync, and the draft's stated reason for allowing it is false for tenant repositories.**
- **The rule.** §4.1 keeps "anything else is malformed and, as today, fails the whole sync". That
  covers an `sdd:` value outside D1, such as `sdd:Add-Auth`, `sdd:my_change`, a name over 100
  bytes, or `sdd:a:b`.
- **The stated reason.** "Malformed values never leave a committed repository's own gate."
- **Why the reason fails.** The gate (`scripts/intent-delivery-applicability.sh`) is this
  repository's own script. It shells out to `$repo/scripts/intent-applicability.sh`, a
  Proofbound-internal file. Nothing makes a tenant repository run it, and commits that predate a
  tenant's adoption were never gated.
- **Why the wedge is permanent.** `Sync` re-lists all history on every run, and `resolvedCommit`
  returns an error at the same commit each time (`connector.go:128-131`, `:194-203`).
- **The contracts it breaks.** This is the batch-fatal rule over full history that sdd SPEC §7
  forbids ("a permanent wedge, because past commits cannot be edited"). sdd SPEC §14 entry condition
  3 allows the alternative only if the draft "proves the trailer gate refuses it first", and no such
  proof is offered.
- **Scope.** The sdd name domain (sdd SPEC §6) is much wider than D1. Every in-domain name outside D1
  is tenant data that M3 turns into a whole-sync failure rather than a per-commit outcome.

**F2 (MED): D3 does not guarantee M3-INV-5 or O-10's "that revision is in the ingested batch".**
- **The premise is narrower than the draft uses it.** §0 row 11 and §4.3 rest on SDD-INV-31, which
  holds only for commits in the sdd walk's **own** `Commits` listing.
- **The two walks are taken at different times.** Under D3 the git connector's listing
  (`connector.go:112`) comes first and the sdd sync's listing comes later.
- **The failure.** A ref deleted or rewound between them leaves a commit that git appends and sdd
  never visits. `Resolve` still returns its revision, because it reads the commit by SHA with no
  reachability check (`build.go:168-215`). That revision can then be absent from the ledger, which is
  the unrecoverable wedge of Q1.
- **The proof cannot catch it.** M3-INV-5's stated proof (a rebuild after `sync git` on a fresh
  store) runs on a static repository and cannot detect this race.

**F3 (MED): D3 contradicts an owner decision in force, and the draft presents it as an extension of
that decision.**
- **The decision.** The sdd SPEC header records the owner decision of 2026-11-17: "the provider
  runs only when selected by name". §12 says `sync all`, `verify` and `delivery-enforce` "still mean
  `records` + `specdir`", and its stated purpose (`cli.go:388-390`) is that they "never walk full
  history".
- **What D3 does instead.** On any repository with an `sdd:` trailer, D3 makes `sync git`, `sync all`
  (`cli.go:668`), both `syncGit` calls in `verify` (`cli.go:905`, `:926`), and therefore
  `delivery-enforce`, construct the sdd provider and walk its full history.
- **How the draft describes it.** It calls this extending "SDD-INV-56's spirit" and does not raise it
  as a change to an owner decision.
- **A side effect in `verify`.** The idempotence check (`requireNoSecondSyncAppends`,
  `cli.go:983-991`) counts only `connectorgit.Result.Appended`. A non-idempotent sdd append made
  inside the second `syncGit` would be invisible to it.

**F4 (MED): under D2, a commit can pass the readiness gate while one of its own claims was dropped,
and nothing in the ledger records the drop.**
- **No ledger record.** A skipped claim yields no ref. `IntentTrailers` is `json:"-"`
  (`connector.go:39`), and `ClaimSkip` is "not ledgered".
- **The gate cannot see the drop.** It judges `targets > 0 && satisfied == targets && reviewed ==
  targets` over the bound rows of `commit_intents_view` only (`gates.go:283-297`). A commit carrying
  `sdd:X` (skipped, for example `archived-earlier` or `name-reuse`) and a bound records or sdd claim
  Y with satisfied targets therefore passes as warranted.
- **The draft's claim is true only per claim.** "Fail-closed on trust ... can never be warranted
  through it" holds for the claim, not for the commit. Before M3, INV-29 made the same situation a
  loud failure.
- **The test cannot catch it.** M3-INV-8 tests only a commit whose single claim was skipped.

**F5 (MED): §4.2's anti-spoof check cannot be built under the boundaries the draft sets.**
- **What the check needs.** It requires `r.ArtifactPath` to equal `<root>/<name>` or
  `<archive>/YYYY-MM-DD-<name>` "for the profile in force". That means `changes.root` and
  `changes.archive` at commit `c`.
- **Where those values live.** They come only from the tenant profile at `c`. `sdd` does not export
  them: the `profile` type is unexported, and so are `changeIntentID` and `datedRE`.
- **Both ways to obtain them break a stated boundary.** §1 fixes the sdd provider's production code
  as "byte-identical", which rules out exporting them. §0 row 10 and §2 say M3 classifies claims
  "without reading tenant bytes itself", which rules out the CLI reading and parsing the profile.
  Parsing it in the CLI would also give profile semantics a second home (Law 2).
- **The fallback is weaker.** A profile-independent check on the basename alone is buildable, but it
  is not what §4.2 specifies.

**F6 (MED): as a Law 6 contract the draft is missing its interface lock and its test derivations.**
- **The skip channel is undefined.** The draft does not say how the CLI resolver signals a
  per-commit skip to the connector. `IntentResolver.Resolve` returns `(IntentRef, error)`, and
  package `git` imports only `core` and `store` (git SPEC §6), so it cannot recognise sdd's
  sentinels.
- **The new types are undefined.** It does not define the `ClaimSkip` type, the `Result.ClaimSkips`
  field (`Result` is part of git SPEC's interface lock), or the cause strings as a closed set.
- **No test derivations.** There is no table of tests, kill mutants, or calibration controls,
  although the parent sdd SPEC and the git SPEC both have them.
- **Two invariants are unguarded.** M3-INV-5 and M3-INV-8 can only be proved on the Postgres path,
  and `make check` does not run that path (`Makefile:27`, no `-tags integration`). Nothing the
  draft specifies would guard them.
- **One invariant is vacuous.** M3-INV-6's shallow-clone clause cannot discriminate, because git
  INV-20 refuses shallow repositories regardless of M3.

**F7 (MED): §5 "Sites (complete list)" is not complete.** It omits:
- `cli/SPEC.md` CLI-INV-7 and CLI-INV-8 (the provider selection and the "intent before git" ordering
  that D3 changes);
- git SPEC §2 `Result` and resolver contract, beyond INV-29;
- the shared trailer corpus `scripts/tests/fixtures/intent-trailers/cases.tsv`, consumed by
  `gitcmd_test.go` and `scripts/tests/intent-delivery-applicability.test.sh`;
- the sdd SPEC's §12, §14 and SDD-INV-56 text;
- the `@generated` `code_commit` pins in `p6-route-matrix.sh` (`routed_dirs` covers `projections` and
  `cli`) and in `p6-task2-close.sh` (`pinned_paths` covers `kernel/internal/connector` and
  `cli/cli.go`).

The draft's own placeholder row ("if the git connector's surface rows name providers") misidentifies
the coupling. The route matrix's rows are source:kind pairs, which M3 does not change, while its pin
covers paths M3 edits.

**F8 (MED): the records/specdir ordering hazard in §8 is real, can wedge projections permanently,
and is wider than §8 says. The draft leaves it unowned.**
- **The mismatch.** `records.Revisions` and `specdir.Revisions` read `HEAD` only (`records.go:48`,
  `specdir.go:51`). Their `Resolve` reads the claiming commit's tree (`records.go:58`,
  `specdir.go:61`).
- **When it bites.** A commit event can cite a revision no sync has ledgered in two situations:
  - after `sync git` alone, once a new CI is committed;
  - under `sync all`, for a commit on any non-HEAD ref that cites a branch-only revision, or a
    revision whose bytes at that commit differ from HEAD's.
- **The consequence** is the projection wedge of Q1.
- **The draft's handling.** It keeps this as an "unverified" non-goal while it widens the same
  projection check and builds an sdd-only ordering mechanism (D3) for the same class.

**F9 (LOW): D1 admits names that can never bind, and the draft does not say so.**
- **The names.** Names such as `a-`, `0--` and `z---` match `^[a-z0-9][a-z0-9-]{1,99}$`, but
  `changeIntentID` fails them with `slug-too-short` (scratch probe; `ids.go:37`, `:57`).
- **The consequence.** Under D2 they become `ClaimSkip unmappable:slug-too-short`, but the widened
  trailer gate (M3-INV-9) would accept them as well-formed claims.

**F10 (LOW): §0 row 9 is evidence that cannot be reproduced from the repository.**
- **What is missing.** It cites "147 change names" and "longest is 42 bytes" from "a probe over
  `rev-list --all` of the reference clone", with no command, no clone revision, and no artifact.
- **Why it matters.** The parent contract's own evidence counts 135 active and 108 archived changes
  (sdd SPEC §0 table). The two counts are not reconciled.
- **What depends on it.** D1's claim that "every real name fits" rests on this row.

**F11 (LOW): D2 has no stated policy for a claim's outcome changing across kernel versions.**
- **How a change re-keys commits.** A binary that classifies a claim differently changes the
  commit's payload and so its `content_sha`. The next sync appends a second event for the same SHA
  (git SPEC §5.1, "What a change here means").
- **Bound to skipped.** `commit_intents_view` rows are only ever inserted or updated, never deleted
  (`projections.go:224`). A binding that a later version withdraws therefore stays visible to the
  readiness gate.
- **Skipped to bound.** M2 (Spec Kit) will turn today's `profile-invalid` skips into bindings for Spec
  Kit repositories. That re-keys those commits, a migration-shaped change the draft does not mention.
