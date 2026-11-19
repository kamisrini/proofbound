---
schema: vera.verdict.v1
verdict_id: r2-m3-spec-review-round3-20261118
status: NEEDS_WORK
reviewed_commit: 36012cfae917f526bf48f4313af55eb379607020
findings:
  - {finding_id: r2-m3-spec-review-round3-20261118-F1, severity: MED}
  - {finding_id: r2-m3-spec-review-round3-20261118-F2, severity: LOW}
  - {finding_id: r2-m3-spec-review-round3-20261118-F3, severity: LOW}
  - {finding_id: r2-m3-spec-review-round3-20261118-F4, severity: LOW}
  - {finding_id: r2-m3-spec-review-round3-20261118-F5, severity: LOW}
artifact_path: docs/verification/verdicts/r2-m3-spec-review-round3-20261118.md
artifact_sha: 43c58f8613d79412a97c952252000a8b5ce4d52cf56a774d7018c63622d3b5af
---

# R2 M3 git-binding SPEC draft 3: non-author review, round 3

## Provenance

- **Reviewer:** an independent adversarial reviewer agent (a Claude model instance spawned for this
  review). It is a **non-author**. It did not write drafts 1-3, the round-1 or round-2 verdicts, the
  sdd SPEC, or any code it read, and it has no authoring context from the session that wrote them
  (Law 9).
- **Under review:** `<repo>/docs/plans/R2-sdd-m3-git-binding-spec-draft3.md` at HEAD
  `36012cfae917f526bf48f4313af55eb379607020`.
- **HEAD moved during the review.** It was `36012cf` at the start and `fc899e5` ("notes: M3 SPEC
  draft 3 under round-3 review") at the end. `git diff --stat 36012cf fc899e5` shows that the only
  change is `notes/state.md`, which the owning session wrote. The draft and all the code read are
  unchanged, so this verdict applies to `36012cf`.
- **Evidence base.** `git diff --stat 722240b 36012cf` touches only the draft, the round-2 verdict,
  `notes/state.md` and one `@generated` integrity pin. No kernel or script file changed since
  round 2, so the code both rounds read is the code at HEAD.
- **`git status` was the same at the start and at the end:** only the two permanent untracked paths
  (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`). Nothing in the repository was
  modified.
- **Read at HEAD:**
  - `connector/git/connector.go` (`Sync`, `resolvedCommit`, the spoof check), `connector/git/SPEC.md`
    (INV-28..30);
  - `connector/git/gitcmd/gitcmd.go` (`trailerLineRE`, `intentTrailers`) and `gitcmd_test.go`
    (`corpusResolver`, `corpusAppender`, `TestIntentTrailers_SharedCorpusKernelBinds`);
  - `cli/cli.go` (`syncGit`, `commitIntentResolver`, `intentProviders`, `runCommand` for `sync git` /
    `sync all` / `intent check`, `resolveIntentGateHeadScope`, `verify`), `cli/SPEC.md`,
    `cli/cli_test.go`, `cli/twin_replay_test.go`;
  - `projections/projections.go` (commit reduce at `:205-229`, `reduceObligationVerdict`,
    `reduceRequirementReview`, `commitPayload.validate`, `Snapshot`), `projections/intent_report.go`
    (`ReportIntent`, `CheckIntent`), `projections/report.go` (`readCommitReportRows`);
  - `gates/gates.go` (`evaluateIntentRule`, `latestIntentProof`);
  - `store/store.go` (`BeginSync`, `Sync.Append`);
  - `connector/intent/intent.go` (`Revision`), `connector/intent/sdd/{provider.go, build.go}`
    (`Resolve`, `inDomain`, `datedRE`, `buildChange`), `connector/reviews/v2.go`;
  - `scripts/intent-delivery-applicability.sh`, its test, `scripts/delivery-enforce.sh`,
    `scripts/kernel-check.sh`, `Makefile`, `scripts/tests/fixtures/intent-trailers/cases.tsv`,
    `docs/deferrals.tsv`.
- **Executed** (with `$HOME/go/bin` first on PATH and GOCACHE under `<scratch>`):
  - `go test ./internal/connector/git/gitcmd/ -run TestIntentTrailers_SharedCorpus -count=1 -v`:
    both corpus tests pass at HEAD.
  - `go test ./internal/cli/ -run TestTwinReplay -count=1`: passes. This is the non-tagged precedent
    for a test that opens the embedded store and appends hand-built events.
  - `git log --all --format=%B | grep -i '^Intent:'`, filtered for any value outside
    `((records|specdir):)?CI-…`: no output. This repository's history holds no `sdd:` or other
    non-standard `Intent` value.

## Round 2, closure

| Round-2 finding | Verdict | Evidence |
|---|---|---|
| F1 MED: replacing rows wedges `reduceObligationVerdict` | **Closed** | Draft 3 keeps `commit_intents_view` as it is at HEAD. The upsert at `projections.go:227` only inserts a row or moves its `event_id`/`seq` forward, and nothing deletes from the table. `reduceObligationVerdict`'s lookup (`:475`) ignores `seq`, so a row it found once is found at every later `seq`, in `Apply` and in `Rebuild` alike. Stated as M3-INV-10 |
| F2 MED: the shared corpus cannot carry `sdd:` rows | **Closed** | With "binds = a ref is bound", every `sdd:` row is unbound in the kernel oracle and refused by the unchanged gate, so the two agree. No existing row changes (attack below). The mechanism is stated inaccurately (new F4) |
| F3 MED: widening the gate opens a bypass | **Closed** | The gate's value regex is unchanged: `^((records\|specdir):)?CI-…$` (`intent-delivery-applicability.sh`). It refuses every `sdd:` value, including a mixed `records` + `sdd` commit. M3-INV-9 pins this |
| F4 MED: rows 8 and 10 cannot run in `make check` | **Partly closed** | Both rows can now run without a build tag (attacks below). Row 8's vehicle, `intent check --commit`, proves the binding, but it does not evaluate readiness, so M3-INV-8 as stated is still untested (new F1). §0a says row 10 is gone, while §7 still specifies it (new F4) |
| F5 LOW: the last-segment check is weak | **Closed** | The residual is recorded in §4.2. `buildChange` sets `ArtifactPath` to the change entry's directory path (`build.go:98`), so a correct `Resolve` always passes the check |
| F6 LOW: D2 contradicts INV-29 | **Closed** | §4.1 no longer claims "exactly as today" for a mixed commit, and §5 amends INV-29's text |
| F7 LOW: eager `L` read, unspecified output | **Partly closed** | `L` is now lazy (§4.3, M3-INV-6). `verify` has no output channel at HEAD, so "prints them in its git step's output" names an output that does not exist (new F5) |
| F8 LOW: no migration for replacement | **Closed** (moot) | There is no replacement |
| F9 LOW: `SDD:x` / `Sdd:x` still wedge | **Partly closed** | The case-folded prefix removes the wedge. Whether `SDD:add-auth` binds or is skipped is ambiguous (new F3) |

## Round 1 partials, closure

| Round-1 finding | Verdict | Evidence |
|---|---|---|
| F6: interface lock and derivations | **Partly closed** | The interface lock (§4.4) and the table (§7) exist. One invariant's derivation proves a different property (new F1) |
| F7: site list | **Closed** | §5 now lists the kernel oracle. The `verify` signature change implied by §4.4 sits inside the listed `cli/cli.go` site |
| F8: records/specdir hazard unowned | **Partly closed**, not re-raised | `docs/deferrals.tsv` at HEAD still has no row for it, so this plan and the verdicts are its only home. D4 is pending with the owner, and draft 3 says the row lands when the owner chooses, which is an owner decision rather than a spec defect |
| F10: survey not reproducible | **Closed** | §0a names the clone and its revision (`f085c850…`, the parent contract's reference) |
| F11: stale bindings across versions | **Closed** as a stated non-goal | Monotonicity (M3-INV-10) and §9. The pointer to M4 as the owner of the residual does not hold (new F2) |

## Answers to the attacks

### Monotone bindings (M3-INV-10)

- **Does insert-only plus `WHERE seq <` make a binding survive a later event without the ref?** Yes.
  - The commit reducer runs an upsert only for refs present in the event (`projections.go:218-229`).
    A later event without the ref touches no row of that key.
  - Nothing deletes from `commit_intents_view`; only `Rebuild` drops it (`:191`).
  - `Rebuild` replays every event, so it reproduces the union. `Apply` and `Rebuild` therefore agree,
    and `verify`'s `CompareSnapshots` stays green.
  - A row's `event_id`/`seq` move to the latest event that carries the ref, and stay put when a
    later event does not. The proof link always names a real event.
  - "Insert-only" in §4.5 is shorthand for "never deleted". The table is upserted on `event_id`/`seq`.
    No reader depends on the difference.
- **`commits_view` drops `intent_refs`.** Its upsert overwrites `intent_refs` with the latest
  event's value (`:215`). So after a later unbound event, `commits_view.intent_refs` is `null` while
  `commit_intents_view` still holds the row.
- **Does any reader take `intent_refs` from `commits_view`?** No. The only reference outside the
  reducer is `Snapshot` (`:872`), which compares and does not decide.
  - These read `commit_intents_view`:
    - the readiness gate (`gates.go:285`);
    - `intent check` (`intent_report.go:304`);
    - `report intent` (`:123`);
    - the verdict cross-check (`projections.go:475`).
  - `report week` reads `commits_view` but not `intent_refs` (`report.go:215`).
  - So no reader hides a binding. The two views simply disagree about the commit's refs. That is a
    fidelity point only; no gate or report reads the stale column.
- **Can a commit stay warranted after its claim became unbindable?**
  - **Within one kernel version, no.** `Resolve` is a function of the commit's content and its
    parents' archive listings (`build.go:168-215`). The connector's checks are deterministic, and `L`
    only grows. So a claim can move from skipped to bound, never back. Under D2 every event of a
    commit carries either all of its refs or none, so the union equals the full set.
  - **Across kernel versions, yes.** Suppose version A binds `{records:A, sdd:X}`, and version B
    classifies `sdd:X` as skipped and so appends an event with no refs. Both rows survive, and the
    gate's `targets` query (`gates.go:285`) still counts them. The commit stays warranted, although
    under version B it carries a skipped claim.
  - The same applies to a binding that a later spoof or `Resolve` fix would refuse: the wrong
    revision keeps warranting the commit, and verdicts citing it keep projecting.
  - Is that acceptable as stated? It is accepted in §9, and it is the price of closing round-2 F1
    without a projection-version migration. Two things are still off: M3-INV-8 states its refusal
    without this qualification, and the residual is handed to an owner that does not cover it
    (new F2).

### Gate unchanged (F3)

- **`make check`** runs `intent-trailer-gate` (`Makefile:76-77`) and `kernel-check`. The gate still
  refuses every `sdd:` value, and the kernel tests are unaffected for this repository, whose history
  holds no `sdd:` value (executed scan above). M3-INV-7 holds trivially here.
- **`delivery-enforce`** runs `sync all`, then the trailer gate, then `gates enforce`
  (`delivery-enforce.sh`).
  - Today an `Intent: sdd:x` commit makes `sync all` fail on `unknown provider`.
  - Under M3, `sync all` succeeds with an `unknown-change` `claim-skip` line (this repository has no
    sdd profile), and the trailer gate then refuses.
  - The outcome is red either way. Only the step that turns red moves.
- **Commits the gate never inspects.**
  - The gate inspects only applicable commits in `@{upstream}..HEAD`, and returns 0 early for
    non-applicable ones.
  - For an `sdd:` value on such a commit, M3 turns today's whole-sync failure into an exit-0
    `claim-skip` line.
  - A non-applicable commit needs no claim, and an applicable one outside the window is still
    unwarranted (no rows), so the change loses no protection. It removes a wedge. Not a finding.

### The corpus oracle change (F2)

- **No existing row flips.** In `resolvedCommit` (`connector.go:180-209`), every trailer either
  appends one ref or returns an error, and `corpusResolver` always returns a well-formed ref for a
  value that passes the connector's checks. So for a single-commit fixture,
  "sync succeeded and the commit had trailers" implies "at least one ref is bound":
  - every current `pass` row binds;
  - every current `fail` row has no trailers, fails `repo.Commits` (UTF-8 rows), or fails `Sync`.

  Both definitions give the same answer on every one of the 32 rows.
- **The kill is reachable.** A malformed or upper-case `sdd:` row is a skip, so `Sync` succeeds with
  no refs. Under the old definition the oracle would read `pass` against the gate's `fail`. So the
  "oracle counting sync success as binding" mutant in §7 row 9 dies on those rows.
- **Implementation note.** `corpusAppender` discards events today, so the new definition needs it to
  keep the appended payload. That is an implementation detail; "a ref is bound" fixes the meaning.

### `intent check --commit` as the proof vehicle for M3-INV-8

- **Ledger, projections, embedded store, no tag?** Yes.
  - `CheckIntent` (`intent_report.go:295-334`) runs `Apply`, then reads `commit_intents_view` joined
    to `change_intents_view`.
  - `cli.Run` with `DATABASE_URL` unset opens the embedded store.
  - `sdd/cli_test.go` and `cli/twin_replay_test.go` are untagged precedents.
- **Does it evaluate readiness?** No.
  - It prints one `commit=… intent=<provider>:<id>@<digest>` line per binding, or `intent=none`.
  - It exits 0 in both cases. It never "refuses".
  - It reads neither `intent_targets_view`, nor `obligation_verdicts_view`, nor
    `requirement_reviews_view`.
- **Can a bound sdd commit "report its sdd target" without a verdict and review chain?**
  - It can report its **binding** (`intent=sdd:CI-…@<digest>`). It cannot report a **target**
    (obligation) or a readiness result.
  - The gate's pass needs a SATISFIED obligation verdict and a VERIFIABLE requirement review per
    target (`gates.go:285-290`), so it cannot be shown without that chain.
  - Through `cli.Run` the chain is constructible without a tag:
    1. commit verdict files under the fixture's `docs/verification/verdicts/`
       (`committedVerdictReader`, `cli.go:282`);
    2. `sync reviews`;
    3. `gates canary` with a gate file whose `scope_commit` names the commit (the pattern in
       `cli_test.go:585`).
  - The two refusal legs need only the commit event: `gates canary` reports `blocked` when
    `targets = 0`.
  - Draft 3 uses none of this (new F1).

### The row-10 test (direct embedded-store appends)

- **Feasible without the integration tag.** `store.Open` with an empty `DatabaseURL` starts the
  embedded Postgres. `BeginSync` and `Sync.Append` accept any envelope that passes `Event.Validate`
  (`store.go:238-286`). `cli/twin_replay_test.go:143-166` does exactly this in a non-tagged test.
- **Its cost.** The change intent's `reduceIntent` needs its target obligations to exist, and the
  verdict's reducer needs requirement revisions and targets (`projections.go:460-500`). So the test
  must also append a valid requirement. Taking the change intent and requirement from a real
  `sync intent sdd` over the tenant fixture is cheaper than hand-building them.
- **Its home.** The projections package's free-port helper is integration-tagged
  (`projections/freeport_test.go`), so a non-tagged test there needs its own helper. The draft does
  not name the package.

### Case-insensitive prefix (F9) and D1's grammar

- **No collision with records or specdir.**
  - A value that case-folds to `sdd:` cannot also start with `records:`, `specdir:` or `CI-`.
  - A value such as `records:sdd:x` does not start with `sdd:`. It takes the INV-29 path and fails
    the whole sync, as today: three parts.
  - `sdd:`, `sdd:a:b` and `sdd:CI-…` all fail D1 (empty, colon, upper case) and become `malformed`
    skips.
- **One reading is ambiguous** (new F3): an upper-case prefix with a valid lower-case name.

## Findings

**F1 (MED): M3-INV-8 states a readiness-gate property, but its only derivation (§7 row 8) runs a
command that does not evaluate readiness. So the property M3 exists to serve, O-8, has no test
that can fail on it.**
- **What INV-8 says.** "The readiness gate passes a commit bound to a SATISFIED `sdd` target with a
  non-owner VERIFIABLE review. It refuses" the skipped-only and the two-claim commit.
- **What row 8 runs.** `intent check --commit`. That command lists `commit_intents_view` rows joined
  to `change_intents_view` (`intent_report.go:304`) and exits 0 whether or not any row exists. It
  never consults targets, verdicts or reviews.
  - Row 8's "(refused)" therefore means "printed `intent=none`".
  - Its positive control, "reports the `sdd` target", is a binding line, not a target.
- **What it does prove.** Row 8's kill, keeping a mixed commit's bindable refs, is now reachable and
  dies. The binding half of round-2 F4 is closed.
- **What it does not prove.** The **passing** half of INV-8 is not derived anywhere: that the
  unchanged gate's query (`gates.go:285`) joins an `intent.sdd` binding through `intent_targets_view`
  to an obligation verdict and a requirement review and returns PASS.
  - No test in `make check` runs that query: every readiness test that opens a store is
    `//go:build integration`.
  - A defect anywhere on the sdd side of that chain would leave every sdd commit permanently
    unwarranted while every M3 test passed. Examples: requirement owner mapping, target rows, the
    verdict's source string.
  - That is exactly O-8 failing.
- **It is provable where the law needs it.** The attack above shows that the chain can be driven
  through `cli.Run` on the embedded store without a tag (`sync reviews`, `gates canary` with a
  scoped gate file). The two refusal legs are cheap: `gates canary` reports `blocked` with
  `targets = 0`.
- **Effect.** Under Law 6 the test derivation does not derive from the invariant it is listed for.
  Under Law 9's corollary, a green row 8 would be believed as proof of a gate property it never
  exercises. Round 2 raised this invariant's derivation as MED, and draft 3 changes the vehicle but
  not the gap.

**F2 (LOW): the monotone residual is handed to an owner that does not cover it, and M3-INV-8's
refusal is stated without the cross-version exception that M3-INV-10 creates.**
- **What survives.** Under M3-INV-10, a binding made by one kernel version keeps warranting its
  commit after a later version would skip or refuse it. This includes a binding produced by a
  `Resolve` or spoof-check defect that is later fixed. The gate counts the surviving rows
  (`gates.go:285`).
- **Where §9 sends it.** §9 says "Reporting such drift belongs with M4".
- **What M4 is.** The parent contract defines M4 as "change-drift reporting (only if re-proposed with
  a ledger-evaluable definition) and the status export", and records that "change drift is NOT in
  R2's BR" (sdd SPEC `:554`, `:849`). That is spec-versus-code drift, conditional and out of
  requirement. It is not binding drift across kernel versions.
- **The result.** The residual has no real home: no deferral row, and no milestone whose scope
  includes it (Law 2).
- **The unqualified invariant.** M3-INV-8 says the gate "refuses … a commit with one skipped and one
  otherwise-satisfied claim". On a ledger where an earlier version bound both claims, it does not.
  The two invariants are consistent only if INV-8 is read as "on a ledger where no event ever bound
  the claim", and the draft does not say so.

**F3 (LOW): whether an upper-case `sdd:` prefix with a valid name binds or is skipped is ambiguous,
and nothing would pin either reading.**
- **§4.1's reading: it binds.** §4.1 says a value whose first four bytes case-fold to `sdd:` is an
  `sdd` claim, and "if the rest matches D1's grammar, it is resolved". `SDD:add-auth`'s rest,
  `add-auth`, matches `^[a-z0-9][a-z0-9-]{1,99}$`, so by §4.1 it resolves and binds.
- **§0a's reading: it is skipped.** §0a's F9 answer says "`SDD:x` is an `sdd` claim and, failing
  D1's lower-case grammar, a `malformed` skip". That implies the prefix's case is part of the grammar
  D1 checks. Its example cannot settle the question, because `x` fails D1 on length alone.
- **D1 itself** fixes the token as `Intent: sdd:<name>`.
- **The tests cannot discriminate.** The `cases.tsv` "upper-case" row (§5) is unbound under both
  readings in the kernel oracle: a `malformed` skip, or a `corpusResolver` echo that fails the spoof
  check. No §7 row pins the outcome.
- **The effect.** Two conforming implementations would bind and skip the same tenant trailer
  respectively. Neither reading wedges a sync, so this is fidelity, not safety.

**F4 (LOW): the draft misdescribes two of its own proof mechanisms.**
- **Row 10.** §0a's F4 answer says "Row 10 is gone with F1", but §7 still specifies row 10, as
  M3-INV-10's only derivation, with direct embedded-store appends. The row is feasible (attack
  above). The statement that it is gone is wrong, and a reader who trusts §0a would not build it.
- **The corpus resolver.** §0a's F2 answer says "`corpusResolver` has no `sdd` arm, so every `sdd:`
  value is unbound".
  - `corpusResolver` has no arms at all. It echoes `Provider: provider, RecordID: id` for any
    provider (`gitcmd_test.go:127-129`).
  - So for a well-formed `sdd:<name>` it returns `RecordID: <name>`. The connector's sdd spoof check
    (§4.2, "`RecordID` CI-shaped") rejects that, which makes the oracle's sync **fail**, rather than
    leaving an unbound claim.
  - The rows still agree with the gate, but for a different reason than the one stated.
  - The test's own comment ("a Sync failure can only come from the kernel's own grammar, framing, or
    value rule", `:124-125`) becomes false for those rows, and no site entry says it changes.

**F5 (LOW): §4.4 specifies `claim-skip` output for `verify` on a channel that does not exist.**
- **No output channel.** `verify(ctx, root, ledger, projector, ids)` (`cli.go:901`) takes no
  writer, and prints nothing on success.
- **No step output.** Its git sync runs inside `verifyStep`, which writes only an optional trace to
  stderr. So "its git step's output" names no existing output.
- **Two syncs.** `verify` runs `syncGit` twice ("initial" and "idempotence"), and the draft does not
  say whether the skips print once or twice.
- **Effect.** Meeting §4.4 means changing `verify`'s signature and its observable output, which no
  CLI SPEC row covers (`cli/SPEC.md` has no `verify` output row). An implementer must invent that
  contract.

## Other observations (no finding)

- **Where `L`'s digest comes from.** §4.3 still does not say whether `L`'s `artifact_sha256` is taken
  from the decoded payload; the event envelope has no digest column. Round 2 called this minor, and
  `ValidateRevision` makes the payload the only source, so an implementer has one choice.
- **M3-INV-5's second sentence** ("a projection rebuild after any interleaving … succeeds") is
  unconditional, but a tenant that also uses `records` claims remains exposed to the D4 hazard. Read
  in context it is scoped to `sdd`.
- **`sync all` and sdd.** `sync all` excludes `sdd` (`TestIntentAllExcludesSDD`), so on a fresh sdd
  tenant every `sdd` claim is `not-ledgered` until `sync intent sdd` runs. §4.3 states the recovery.
  That is a deliberate consequence of withdrawing D3.

## Law walk

- **Law 1.** The two `@generated` pins are listed for regeneration after the code commits (§5).
  Compliant.
- **Law 2.**
  - Draft 3 removes the D1 grammar from the gate script, so it has fewer homes than draft 2: the
    connector and the VD.
  - The dated-archive pattern still has a second home in the CLI's segment check (accepted in
    round 2).
  - The cross-version binding residual has no real home (F2), and D4 still lives only in this plan
    pending the owner.
- **Law 6.** A SPEC ahead of code, with an interface lock, a derivation table and a calibration
  statement. One invariant's derivation does not exercise the invariant (F1).
- **Law 7.** Round 2's general rule ("a reduce step must not depend on a row a later event can
  delete") is now compiled as M3-INV-10, with a derivation (row 10). This is the right compilation
  of that wedge class.
- **Law 8.** No new dependency. Compliant.
- **Law 9.** The draft asks for a non-author re-check, and this is one. The calibration requirement
  is stated. Row 8's calibration would pass while proving nothing about the gate property it is
  listed for (F1).

## Verdict

**NEEDS_WORK.**
- Draft 3 closes three of round 2's four MED findings:
  - the verdict wedge, by keeping the bindings monotone;
  - the corpus oracle, with no existing row flipping;
  - the gate bypass, by leaving the gate unchanged.
- The ledger check, D2 and the monotone projection hold together within one kernel version.
- One MED remains (F1): M3-INV-8, the readiness property that O-8 needs, is still derived from a
  command that never evaluates readiness, although the readiness chain can be driven inside
  `make check`.
- The four LOW findings may remain under ACCEPTABLE.
