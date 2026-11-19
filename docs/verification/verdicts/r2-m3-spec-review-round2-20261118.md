---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round2-20261118
status: NEEDS_WORK
reviewed_commit: 722240b557149284190b73bfd1a7afab397ea540
findings:
  - {finding_id: r2-m3-spec-review-round2-20261118-F1, severity: MED}
  - {finding_id: r2-m3-spec-review-round2-20261118-F2, severity: MED}
  - {finding_id: r2-m3-spec-review-round2-20261118-F3, severity: MED}
  - {finding_id: r2-m3-spec-review-round2-20261118-F4, severity: MED}
  - {finding_id: r2-m3-spec-review-round2-20261118-F5, severity: LOW}
  - {finding_id: r2-m3-spec-review-round2-20261118-F6, severity: LOW}
  - {finding_id: r2-m3-spec-review-round2-20261118-F7, severity: LOW}
  - {finding_id: r2-m3-spec-review-round2-20261118-F8, severity: LOW}
  - {finding_id: r2-m3-spec-review-round2-20261118-F9, severity: LOW}
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round2-20261118.md
artifact_sha: d218718b356e9306d3c82631660a45c85c79d16d51f6bb19ede5eab620240397
---

# R2 M3 git-binding SPEC draft 2: non-author review, round 2

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**. It did not write draft 1 or draft 2, the round-1 verdict, the sdd
  SPEC, or any code it read, and it has no authoring context from the session that wrote them (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft2.md` at HEAD
  `722240b557149284190b73bfd1a7afab397ea540`. HEAD was the same at the start and at the end of the
  review.
- **The draft's evidence base.** Its §1 says it was read from `82a92ec`. `git diff --stat 82a92ec
  722240b` shows that only the draft itself was added, so the code it cites is the code at HEAD.
- **`git status` was the same at the start and at the end:** only the two permanent untracked paths
  (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`). Nothing in the repository was
  modified.
- **Read at HEAD:**
  - `connector/git/connector.go` and its `SPEC.md` (§2, INV-20, INV-28..30, §5.1);
  - `connector/git/gitcmd/gitcmd.go` (`intentTrailers`, `trailerLineRE`) and `gitcmd_test.go`
    (`TestIntentTrailers_SharedCorpusKernelBinds`, `corpusResolver`);
  - `cli/cli.go` (`syncGit`, `commitIntentResolver`, `verify`) and `cli/SPEC.md` (CLI-INV-7/8);
  - `projections/projections.go` (`Apply`, `Rebuild`, `reduce`, `reduceIntent`,
    `reduceObligationVerdict`, `commitPayload.validate`, `Snapshot`) and `intent_report.go`;
  - `gates/gates.go` (`evaluateIntentRule`, `latestIntentProof`);
  - `store/store.go` (`Filter`, `ReadEvents`, `Open`) and `store/migrations/001_ledger.sql`;
  - `core/core.go` (`IdempotencyKey`);
  - `connector/intent/intent.go` (`ValidateRevision`);
  - `connector/intent/sdd/{SPEC.md, provider.go, build.go, cli_test.go}`;
  - `connector/reviews/{reviews.go, v2.go}`;
  - `scripts/intent-delivery-applicability.sh`, `scripts/kernel-check.sh`, `Makefile`,
    `scripts/tests/fixtures/intent-trailers/cases.tsv`, `docs/deferrals.tsv`.
- **Executed:** `go test ./internal/connector/intent/sdd/ -run
  'TestSyncIntentSDDPrintsSkips|TestIntentAllExcludesSDD' -count=1 -v`, with `$HOME/go/bin` first on
  PATH and GOCACHE under `<scratch>`. Both tests passed, and the embedded Postgres started and stopped
  in the log. I ran no other code.

## Round 1, closure

| Round-1 finding | Verdict | Evidence |
|---|---|---|
| F1 HIGH: a malformed `sdd:` trailer wedges every sync | **Closed** for its stated scope | §4.1 makes every `sdd:`-prefixed value a per-commit skip. The check has to run before `resolvedCommit`'s `strings.Split(raw, ":")` / `len(parts) != 2` check (`connector.go:189-193`), and "any value beginning `sdd:`" places it there, so `sdd:a:b` is covered. The values just outside the prefix still wedge the sync (new F9) |
| F2 MED: D3 races ref movement | **Closed** | The ledger check (§4.3) makes the binding depend only on the ledger, never on the refs. The attacks below confirm that every revision in `L` is reduced into `change_intents_view` before any commit event that cites it |
| F3 MED: D3 walks sdd history from `sync git` / `verify` | **Closed** | `sync git` only calls `Resolve`, which reads one state plus its parents' archive listings (`build.go:168-215`). It does not walk history. §5 lists the SDD-INV-56 text as a site. Under M3, `sync git` and `verify` still construct the provider on repositories that carry `sdd:` trailers, which goes further than the letter of "the provider runs only when selected by name". The draft calls this "unchanged in substance" and does not raise it with the owner. That is a wording point, not a harm |
| F4 MED: one skipped claim beside one bound claim passes the gate | **Closed** | Under D2 the commit carries no refs, so the readiness query (`gates.go:285`) counts `targets = 0` and blocks it. New F6 records a residual inconsistency |
| F5 MED: the path check needs profile values | **Closed** (buildable) | The last-segment check needs nothing from the profile. It is weaker than draft 1's check (new F5) |
| F6 MED: skip channel, types, test derivations | **Partly closed** | §4.4 defines the interface lock and §7 the derivation table. Two invariants are not provable inside `make check`, and one kill cell names a mutant that its test never executes (new F4) |
| F7 MED: the site list is incomplete | **Partly closed** | The listed sites are right. The shared corpus is listed, but the kernel-side oracle that consumes it is not, and that oracle cannot accept `sdd` rows (new F2) |
| F8 MED: the records/specdir hazard is unowned | **Partly closed** | It is raised to the owner as D4, which is an acceptable owner for it. D4 recommends "file a deferral now", but `docs/deferrals.tsv` at HEAD has no row for it, so until the owner decides, the only place the hazard is recorded is this plan |
| F9 LOW: D1 names that fail `slug-too-short` | **Closed** | §4.4 maps them to `unmappable:slug-too-short`. Also confirmed: `UnmappableError.Unwrap` returns `intent.ErrUnmappable` (`provider.go:32`) |
| F10 LOW: the "147 names" figure cannot be reproduced | **Partly closed** | The command is given and it is correct: the optional dated group and `grep -vx archive` handle files directly under `archive/`. The clone and its revision are still not identified, so the figure still cannot be reproduced from the repository. Not re-raised: only the grammar matters, as the draft says |
| F11 LOW: stale bindings across versions | **Partly closed** | Replacing the rows per commit removes stale bindings, but it opens a new permanent projection wedge through the verdict cross-check (new F1). It also leaves `Apply` and `Rebuild` disagreeing on stores built before M3 (new F8) |

## Answers to the attacks

### The ledger check (§4.3)

**Is reading `L` once per `sync git` safe?** Yes.
- `sync git` appends only `git/commit.recorded` events (`connector.go:134-149`), so nothing it does
  adds to `intent.sdd`.
- The ledger is append-only, so `L` only ever grows. A snapshot can therefore be stale only toward
  "smaller". A concurrent `sync intent sdd` in another process can only cause a spurious
  `not-ledgered` skip, which the next sync recovers.
- Every member of `L` was ledgered before the snapshot was read. Every commit event the same sync
  appends gets a higher `seq`. So M3-INV-5's "at a lower sequence number" holds by construction.

**Can a revision be in `L` yet absent from `change_intents_view` when the commit event is reduced?**
Not while the projection succeeds.
- `eventRoute` sends `intent.sdd` change intents to `reduceIntent` (`projections.go:300`).
- `reduceIntent` upserts every such event (`:362`). Nothing ever deletes rows from
  `change_intents_view`, and a superseding sdd event only adds rows, because `supersedes` is reference
  validation only.
- `Apply` reduces records strictly in `seq` order, inside one transaction, from the checkpoint
  (`:159-179`). `Rebuild` re-runs `Apply` from zero.
- So both paths reduce the change intent before the commit event that cites it.
- The only exception is a change-intent event whose own `reduceIntent` fails, for example
  `dangling or inactive obligation` (`:372`). That aborts the whole projection at the change intent's
  own `seq`, before the commit is reached. It is a pre-existing sdd-sync failure, not one M3
  introduces.

**Do the keys agree?** Yes.
- `ValidateRevision` requires `payload.intent_id == NativeID` and `payload.artifact_sha256 ==`
  the envelope digest (`intent.go:300`).
- `reduceIntent` keys `change_intents_view` on `(source, v.IntentID, v.ArtifactSHA256)`, and the
  dangling check queries `(intent.<provider>, RecordID, digest)` (`:221`).
- `L`'s `(native id, artifact_sha256)` is therefore the same key, provided the implementation takes
  the digest from the decoded payload. The draft does not say where the digest comes from; the event
  envelope has no digest column (`store.go:410`). This is minor.

### All-or-nothing (D2/§4.4)

- **Existing commits.** No existing ledger can hold a commit event with an `sdd` value: before M3,
  the allowlist at `connector.go:195` fails the sync. So D2 changes no existing commit's payload, and
  the pinned §5.1 vector (`IntentRefs: nil`) is unaffected.
- **INV-29 as written** ("every trailer … stores the returned provider, CI id, and exact artifact
  digest") is broken for a `records` / `specdir` trailer that sits beside a skipped `sdd` claim. Draft
  §4.1 nevertheless says those trailers "behave exactly as today" (new F6).
- **Can a tenant strip a valid records claim with a deliberately bad `sdd:` trailer?** Yes, but the
  author controls their own trailers, and the result fails closed: the commit becomes unwarranted.
  That is something the author could equally do by omitting the trailer. The residual cost is
  fidelity, not trust. A dropped claim is indistinguishable in the ledger from no claim, because
  `IntentTrailers` is `json:"-"` and `ClaimSkip` is not ledgered.

### `commit_intents_view` replacement (§4.5)

- **Incremental `Apply` and `Rebuild`.** For new events it is correct. `Apply` and `Rebuild` both see
  a commit's events in ascending `seq`, so "higher sequence than the current row" always holds, and
  the outcome matches the `WHERE seq <` upsert on `commits_view` (`:215`).
- **It is not retroactively neutral, and the draft argues retroactive safety only for the validate
  widening.**
  1. `reduceObligationVerdict` refuses, as an error that aborts the projection, any v2 verdict whose
     `(reviewed_commit, CI@digest)` is not in `commit_intents_view` at the verdict's `seq`
     (`projections.go:475-480`). Insert-only rows made that check monotone. Replacement rows do not
     (new F1).
  2. `Rebuild` re-derives the latest-only semantics, while an incremental `Apply` over a store built
     before M3 keeps union rows. `verify` compares the two snapshots (`cli.go:952-967`), and
     `commit_intents_view` is part of the snapshot (`projections.go:882`) (new F8).
- **Existing tests.** No non-tagged test touches `commit_intents_view`. Every test that does is
  `//go:build integration` (`projections/intent_test.go`, `projection_test.go`,
  `gates/gates_integration_test.go`). None of them asserts two events for one commit, so I found no
  existing test that changes.
- **The readiness gate for records commits.** Its meaning changes from "union of every event's refs"
  to "latest event's refs". In this repository a records commit's refs are deterministic within one
  binary version, so the change is visible only across versions. I did not inspect this repository's
  live ledger, because doing so would open and write its projection state.

### The anti-spoof last-segment check (§4.2)

- **It is sound against the resolver failures it targets:** wrong kind, wrong source, and a wrong
  change whose directory has another name. `buildChange` sets `ArtifactPath` to the change entry's
  path (`build.go:98`), and Resolve's archive match is `datedRE` with an exact capture
  (`provider.go:55`, `build.go:205`). A correct Resolve therefore always passes the check, so the
  check cannot create a false-spoof wedge on tenant data.
- **A dated prefix can collide with an active name.** `2024-01-01-foo` is itself a valid D1 name and
  a valid active change. The check accepts its revision for the claim `foo`, because the last segment
  matches the dated alternative. Nothing distinguishes a root parent from an archive parent, and
  nothing ties `NativeID` to the name (new F5).

### Laziness (M3-INV-6) together with the ledger read

- §4.2 makes the provider lazy, but §4.3 reads `L` "once per `sync git`", without condition.
- On a repository with no `sdd:` values, this adds one `ReadEvents` call per `sync git`, which means
  two per `verify`.
- That read is cheap: `events_source_kind_seq` (`001_ledger.sql:14`) serves it, and it returns zero
  rows.
- It is still new store work, and a new way to fail, on every git sync, and M3-INV-6 does not cover
  it (new F7).

### The test-derivation table (§7)

- **Confirmed:** `sdd/cli_test.go` has no build tag. It calls `cli.Run` with `DATABASE_URL=""`, which
  starts an embedded Postgres. It ran green here. `kernel-check.sh` runs `go test ./... -count=1 -v`
  with no tags, so these tests run inside `make check`.
- **Rows 1, 5 and 6** (cli parts) can be proved inside `make check`: `sync intent sdd`, `sync git`,
  `rebuild` and `intent check --commit` all exist as commands (`cli.go:37`).
- **Rows 2, 3, 4 and 7** are git unit tests with fakes; they need no store.
- **Row 9** runs over the corpus, but see new F2.
- **Rows 8 and 10 cannot run inside `make check` as written** (new F4).

## Findings

**F1 (MED): replacing the rows per commit (§4.5) opens a permanent projection wedge through the
v2-verdict cross-check, on the very path that F11 addresses.**
- **The cross-check.** `reduceObligationVerdict` returns an error, not a skip, when a verdict's
  `(reviewed_commit, provider, CI, digest)` has no row in `commit_intents_view`
  (`projections.go:475-480`, "v2 verdict does not match an exact commit intent claim").
- **Why it was safe before.** Today rows are never deleted, so once a binding has appeared, every
  later-sequenced verdict against it reduces.
- **What §4.5 changes.** Under §4.5 a later commit event deletes the binding. F11's scenario is
  exactly such an event: a version that reclassifies a bound claim as skipped, or drops one of two
  refs.
- **The failure sequence.**
  1. The commit binds at seq 10.
  2. A re-keyed event for the same commit, which no longer binds that revision, is appended at
     seq 20.
  3. A verdict naming the old binding (written from earlier `intent check` output, or first synced
     after the upgrade) is ledgered at seq 30.
  4. Every `Apply` and every `Rebuild` now aborts at seq 30.
- **Why it is permanent.** The ledger is append-only and the order is fixed, so this is the same
  permanent-wedge class as round-1 Q1 (sdd SPEC §7).
- **Impact.** The gates turn this into BLOCKED or "projection failed" (`gates.go:269-274`).
- **The draft does not cover it.** It argues retroactive safety only for the validate widening, and
  M3-INV-5 / M3-INV-10 do not mention verdict reduction.

**F2 (MED): the shared trailer corpus cannot carry the `sdd:` rows §5 adds, because the kernel-side
oracle reads "sync succeeded" as "binds".**
- **The kernel oracle.** `TestIntentTrailers_SharedCorpusKernelBinds` (`gitcmd_test.go:142-200`,
  non-tagged, so it runs in `make check`) computes `binds = syncErr == nil`, using `corpusResolver`.
  That resolver echoes `RecordID: id` (`:127-129`).
- **The gate oracle.** The gate test reads column 3 as the gate's exit status
  (`intent-delivery-applicability.test.sh:35`).
- **How the two disagree under M3:**
  - **A valid `sdd:<name>` row:** the gate passes (M3-INV-9). In the kernel test, `corpusResolver`
    returns the name as the `RecordID`, so §4.2's CI-shape check fails the sync, and the kernel reads
    "fail".
  - **An upper-case `sdd:Add-Auth` row:** the gate fails. In the kernel test it is a `malformed`
    skip, so the sync succeeds and the kernel reads "pass".
- **Consequence.** Neither value of column 3 satisfies both consumers. The corpus's documented
  meaning ("column 3 means the kernel binds a claim", `cases.tsv:6`) stops being true for `sdd`. Under
  D2 a successful sync no longer means "binds". The anti-drift property of
  DEF-INTENT-TRAILER-BLOCK-001 is lost for `sdd`.
- **Missing from §5.** Neither the oracle nor `corpusResolver` is a §5 site, so `make check` goes red
  as soon as the listed `cases.tsv` rows land, or the corpus semantics change without a contract.

**F3 (MED): widening the trailer gate (M3-INV-9) gives this repository a zero-cost bypass of its own
intent gate, because an `sdd:` claim that can never bind passes the gate and is then silently
skipped.**
- **The gate's contract.** `check_commit`'s contract is "0 when the commit … carries one [claim] the
  kernel binds" (`intent-delivery-applicability.sh:27`). `make check` runs it as
  `intent-trailer-gate` (`Makefile:76-77`).
- **What M3 makes it accept.** The gate accepts any D1-shaped `sdd:<name>`.
- **What the kernel does with that claim here.** This repository has no sdd profile, so `Resolve`
  returns `ErrUnknownChange` for every name (sdd SPEC §6, outcome 2). Under D2 that is a silent
  `unknown-change` skip, and the sync exits 0.
- **The bypass.** A governed commit can satisfy `make check` with `Intent: sdd:anything` and never
  name a real change intent.
- **What it replaces.** Today a bogus `records` id passes the regex but fails `sync git` loudly at
  resolution. The `sdd` route removes even that signal.
- **Not stated anywhere.** The draft does not state this effect on the repository's own enforcement
  (the CLAUDE.md premise: "if the way we build contradicts what we're building, the build is wrong").

**F4 (MED): two invariants in §7 cannot be proved inside `make check`, and one kill cell names a
mutant its test never executes.**
- **Row 8** ("gates unit over projected rows") and **row 10** ("projections unit: two events for one
  commit") both need a store.
- **Where the store tests are.** In both packages, every test that opens a store is
  `//go:build integration` (`gates_integration_test.go`, `projections/projection_test.go`,
  `intent_test.go`). The non-tagged tests (`gates_test.go`, `citation_test.go`, `report_test.go`,
  `placeholder_test.go`) never open a store.
- **Why `make check` misses them.** `kernel-check.sh:7` runs no tags.
- **The draft contradicts itself.** Its own §0 answer to F6 says that ledger-touching invariants are
  proved "through `cli.Run` on the embedded store", but rows 8 and 10 are not assigned there.
- **Row 8's kill cell is unreachable.** The cell is "gate passing on a partial binding". The gate is
  unchanged by M3, so the only mutant that produces a partial binding lives in the connector's
  all-or-nothing drop. A test over hand-inserted projected rows (the pattern at
  `gates_integration_test.go:203`) never runs the connector, so that mutant survives.
- **Effect.** Under Law 9's corollary, this row's calibration cannot be believed: M3-INV-8's
  two-claim case, which is round-1 F4's protection, would be guarded by a test that cannot fail.

**F5 (LOW): the last-segment anti-spoof check accepts an active change whose name is a dated form of
the claimed name.**
- **The collision.** For claim `sdd:foo`, a revision with `ArtifactPath` `<root>/2024-01-01-foo`
  passes §4.2. Its last segment matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}-foo$`. `2024-01-01-foo` is a
  legal D1 name and a legal active change.
- **Why it slips through.** The check does not know whether the parent segment is the root or the
  archive, and it never compares `NativeID` with the id derived from the name.
- **Impact.** A resolver defect that maps `foo` to that sibling is therefore not caught.
  `Resolve`'s exact-name lookup remains the primary guarantee (`build.go:192`), as the draft says, so
  this is defence-in-depth only.

**F6 (LOW): D2 breaks INV-29's text for `records` / `specdir` trailers, while §4.1 says they "behave
exactly as today".**
- **The conflict.** A valid `records:<CI>` beside a skipped `sdd` claim stores no ref. INV-29 says
  "every trailer … stores the returned provider, CI id, and exact artifact digest" (git SPEC
  `:403-405`).
- **What §5 does with INV-29.** §5 amends INV-29 only to say "`sdd` claims fail per commit; others
  unchanged", and M3-INV-7 is scoped to commits without `sdd` values. So the contract states two
  things that cannot both hold for a mixed commit.
- **Lost fidelity.** The ledger then shows the mixed commit as unclaimed, which is indistinguishable
  from a commit with no trailer.

**F7 (LOW): the ledger read is unconditional, which leaves the laziness promise partial and the
claim-skip output under-specified.**
- **Unconditional read.** §4.3 reads `L` on every `sync git`, and so twice per `verify`
  (`cli.go:905`, `:926`), including on repositories with no `sdd:` value. M3-INV-6 pins only the
  provider's construction.
- **Output.** §4.4 specifies the `claim-skip` lines for `sync git` only. It is silent on `sync all`
  (`cli.go:668`) and `verify`, both of which reach `syncGit`.

**F8 (LOW): the replacement semantics come with no migration, so the first `verify` on a store built
before M3 can fail `CompareSnapshots`.**
- **The divergence.** An incremental `Apply` keeps union rows for commits that get no new event,
  while `Rebuild` derives latest-only rows. `verify` compares the two (`cli.go:952-967`).
- **When it fails.** Any commit whose older event carried a ref its newer event lacks makes the
  comparison fail once. It self-heals, because verify's own rebuild leaves the store latest-only.
- **It also affects the gates.** `projection_version` stays `1` (`projections.go:154`), so nothing
  forces the rebuild. The readiness gate evaluates through `Apply` (`gates.go:269`), so it keeps
  seeing stale union rows on such stores until something rebuilds them.
- **Not checked.** I did not inspect whether this repository's ledger has such a commit.

**F9 (LOW): the F1 wedge class survives for near-miss values outside the exact `sdd:` prefix.**
- **What the trailer reader accepts.** `trailerLineRE` keeps the value's case
  (`gitcmd.go:278`).
- **Which values still fail the whole sync.**
  - `Intent: SDD:add-auth` and `Intent: Sdd:add-auth` fail with `unknown provider`
    (`connector.go:195-196`).
  - A free-text tenant value such as `Intent: add auth` fails with `malformed change intent`.
- **Why it matters now.** In a tenant history either one wedges every git sync permanently, which is
  the condition the sdd SPEC §7 forbids. M3 is what makes `sync git` worth running on such
  repositories.
- **Scope.** This is pre-existing INV-29 behaviour, and §4.1 keeps it deliberately. It is recorded
  because the draft treats the class as closed for `sdd`, while the boundary of the fix is the
  byte-exact prefix.

## Law walk

- **Law 1.** Both `@generated` pins are listed as sites (§5), and they fail closed. Compliant.
- **Law 2.**
  - The archive-name pattern gains a second home in the CLI's last-segment check, alongside
    `sdd.datedRE`. It is small and profile-free, so it is acceptable.
  - The D1 grammar will live in the connector, the gate script and the VD, as the CI regex already
    does, and the shared corpus is the drift defence. F2 shows that the defence does not extend to
    `sdd`.
  - The D4 hazard has no deferral row, so this plan is its only home (closure table, F8).
- **Law 6.** The draft is a SPEC ahead of code, now with an interface lock (§4.4) and a derivation
  table (§7). Two rows are not provable where the law needs them (F4).
- **Law 7.**
  - The ledger check compiles the round-1 wedge class into one enforcement point for `sdd`.
  - §4.5 adds a new instance of the same class through the verdict cross-check (F1). No invariant
    states the general rule: a reduce step must not depend on a row that a later event can delete.
  - `records` / `specdir` remain a recorded instance pending D4.
- **Law 8.** No new dependency. Compliant.
- **Law 9.** The draft asks for a non-author re-check, and this is one. The calibration statement in
  §7 is the right shape, but row 8's kill mutant lies outside its test's reach (F4), so that
  calibration would pass while proving nothing.

## Verdict

**NEEDS_WORK.**
- The ledger check (§4.3) is sound and closes round 1's HIGH and its race. All-or-nothing closes F4.
- Four MED findings remain:
  - a new permanent-wedge path that the F11 fix introduces (F1);
  - a shared-corpus oracle that cannot express D2 (F2);
  - a bypass of this repository's own intent gate (F3);
  - two invariants that `make check` cannot prove, one of them with an unreachable kill (F4).
