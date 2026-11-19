# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 3)

**Status:** draft 3, 2026-11-18. It answers non-author round 2,
[`r2-m3-spec-review-round2-20261118`](../verification/verdicts/r2-m3-spec-review-round2-20261118.md)
(NEEDS_WORK: 4 MED, 5 LOW), over [draft 2](R2-sdd-m3-git-binding-spec-draft2.md); round 1's answers are
kept below. It needs a non-author re-check (Law 9) and the owner's decisions D1-D4 (§3) before any code
lands (Law 6). **Parent contract:** `kernel/internal/connector/intent/sdd/SPEC.md` (R2 M1, accepted).
**Requirement served:** `BR-sdd-warranty-i9miv7` O-8 and O-10 ("fails only a claim on that state").

## 0a. Round 2, answered

| Finding | Answer in draft 3 |
|---|---|
| F1 MED — replacing `commit_intents_view` rows wedges `reduceObligationVerdict` | **Replacement withdrawn.** `commit_intents_view` stays insert-only, as today. A binding, once made, persists (monotone), so a verdict citing it always projects. Round-1 F11 (a later version un-binding) is answered by monotonicity, stated as M3-INV-10, and its consequence is a non-goal (§9) |
| F2 MED — the shared corpus cannot carry `sdd:` rows | The kernel oracle `gitcmd_test.go::TestIntentTrailers_SharedCorpusKernelBinds` now counts a row as bound when **a ref is bound** (not when the sync succeeds). Its `corpusResolver` has no `sdd` arm, so every `sdd:` value is unbound. With this repository's gate unchanged (F3), the `sdd:` rows are gate `fail` / kernel unbound, which agree. The oracle is listed as a site (§5) |
| F3 MED — widening the gate lets `Intent: sdd:x` bypass this repository's `make check` | **The trailer gate script is not changed.** It governs Proofbound's own commits, and Proofbound is not an `sdd` tenant. Tenant repositories do not run it. M3-INV-9 now asserts the gate still **refuses** `sdd:` values |
| F4 MED — §7 rows 8 and 10 cannot run in `make check` | Row 10 is gone with F1. Row 8 is proved end to end through `cli.Run` on the embedded store (no build tag; round 2 confirmed this runs in `make check`): an sdd tenant fixture, `sync intent sdd`, `sync git`, then `intent check --commit`. The connector is in the path, so the partial-binding mutant is reachable |
| F5 LOW — the last-segment check cannot tell root from archive, and does not tie the native id to the name | **Stated as defence in depth.** The primary guarantee is `Resolve`'s exact-name semantics, pinned by SDD-INV-59's `expected-resolve.json`. Tying the id to the name needs `changeIntentID`, which is unexported, and exporting it changes the admitted sdd tree. The residual is recorded (§4.2), not hidden |
| F6 LOW — all-or-nothing contradicts INV-29 and "records exactly as today" | INV-29's text is amended in §5 to say this. §4.1 no longer claims "exactly as today" for a commit that also carries an `sdd` value. The ledger does not distinguish such a commit from an unclaimed one; that is accepted, because neither can be warranted |
| F7 LOW — the ledger read is eager, and the skip output is unspecified for `sync all` / `verify` | The set `L` is read **lazily**, on the first `sdd` claim. A history with no `sdd` value does neither the read nor the construction (M3-INV-6). `sync git` and `sync all` print `claim-skip` lines after their counts line. `verify` prints them in its git step's output. None changes an exit status |
| F8 LOW — no migration for the replacement rule | Moot: there is no replacement rule (F1) |
| F9 LOW — `SDD:x` / `Sdd:x` still wedge | The `sdd:` prefix is matched **ASCII-case-insensitively**. `SDD:x` is an `sdd` claim and, failing D1's lower-case grammar, a `malformed` skip. Free-text `Intent:` values are pre-existing INV-29 behaviour and go to the D4 deferral (§3) |
| Round-1 F8 partial — no deferral row | D4 recommends filing it; the row lands when the owner chooses (deferral rows are owner-dormant) |
| Round-1 F10 partial — clone unnamed | The survey ran on a full clone of `Fission-AI/OpenSpec` at `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`, the parent contract's reference revision |

## 0. Round 1, answered (draft 2)

| Finding | Answer in draft 2 |
|---|---|
| F1 HIGH — a malformed `sdd:` trailer in tenant history fails every git sync | Every value beginning `sdd:` is tenant data. A malformed one is a per-commit claim skip (`malformed`), never a sync failure (§4.1). Non-`sdd` values keep INV-29 unchanged |
| F2 MED — D3's "sync sdd first" misses revisions outside the sdd listing (refs moved) | D3 is **withdrawn**. Replaced by the **ledger check** (§4.3, D3'): a ref binds only if its exact revision is **already in the ledger**; otherwise the claim is skipped (`not-ledgered`). No commit event can ever name an un-ledgered revision, whatever the refs do |
| F3 MED — D3 made `sync git` / `sync all` / `verify` / `delivery-enforce` walk sdd history | Gone with D3. `sync git` never runs the `sdd` intent sync; it only **resolves** claims, on demand. SDD-INV-56 is unchanged, and `verify`'s second-sync-appends-nothing check sees no hidden appends |
| F4 MED — one skipped claim plus one bound claim passes the gate | **All-or-nothing per commit (D2):** if any `sdd` claim on a commit is skipped, the commit binds **no** refs at all, so it cannot be warranted. M3-INV-8 now has a two-claim case |
| F5 MED — the path check needs profile values sdd does not export | Replaced by a profile-free check on the revision's **last path segment**: it equals `<name>`, or matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}-<name>$` (§4.2). No sdd code change and no second profile parser. `Resolve`'s exact-name semantics remain the primary guarantee |
| F6 MED — skip channel, types, test derivations, untestable invariants | §4.4 defines `connectorgit.ErrClaimUnbound`, `ClaimSkip`, and `Result.ClaimSkips`. §7 is a test-derivation table with kills and controls. Ledger-touching invariants are proved through `cli.Run` on the embedded store, as `sdd/cli_test.go` already does inside `make check`. The vacuous shallow-clone clause is dropped |
| F7 MED — the site list was incomplete | §5 adds: `cli/SPEC.md` CLI-INV-7/8, git SPEC §2 (`Result`, resolver), the shared trailer fixture `cases.tsv`, the sdd SPEC §12/§14/SDD-INV-56 text, and the two generated pins |
| F8 MED — the records/specdir ordering hazard is real and unowned | **Owned elsewhere, raised to the owner as D4** (§3). M3's ledger check is the same mechanism, so extending it is cheap. But changing `records` / `specdir` from whole-sync failure changes git INV-29 for this repository's own enforcement, and that is the owner's call |
| F9 LOW — names matching D1 that fail `slug-too-short` | They resolve to `ErrUnmappable:slug-too-short` → claim skip `unmappable:slug-too-short` (§4.4). The trailer gate accepts them deliberately: the grammar stays a syntax check, and outcomes belong to resolution |
| F10 LOW — the "147 names" figure had no command and no reconciliation | §1 gives the command. It counts every distinct directory name ever under `openspec/changes/` (active, and archive captures), across all of history. The parent's 135 counts distinct change **names**, after archive captures are folded into their active names. Both meet the grammar; only the grammar check matters here |
| F11 LOW — a claim outcome that changes across versions leaves stale bindings | §4.5: `commit_intents_view` holds the bindings of a commit's **latest** event only. A newer commit event replaces the commit's rows. **Superseded in draft 3** (round-2 F1): bindings are insert-only and monotone (M3-INV-10) |

## 1. Evidence (read from HEAD `82a92ec`; re-checked at `1489027`)

As draft 1 §0, with these corrections:
- **A commit event's idempotency key is `(source, native id, content sha)`** (`core.Event.IdempotencyKey`).
  So a later sync that binds differently appends a **new** event for the same commit. `commits_view`
  upserts by sequence (`projections.go:215`). `commit_intents_view` only inserts (`:227`), which is the
  F11 hazard.
- **The ledger can answer "is this exact revision ledgered?"** through `store.ReadEvents` with
  `Filter{Source: intent.sdd, Kind: change_intent.recorded}`, read once per `sync git` into a set of
  `(native id, artifact_sha256)`.
- **Change-name survey** (F10):
  `git -C <clone> log --all --format= --name-only -- openspec/changes | sed -nE 's#^openspec/changes/(archive/[0-9]{4}-[0-9]{2}-[0-9]{2}-)?([^/]+)/.*#\2#p' | grep -vx archive | sort -u`
  gives 147 names, all matching `^[a-z0-9][a-z0-9-]*$`, the longest 42 bytes.

## 2. Purpose, boundary, entry conditions

M3 binds a tenant commit's `Intent: sdd:<name>` to the exact `intent.sdd` change-intent revision at
that commit, so the unchanged readiness gate (O-8) can judge it.
- **Changed:** git connector trailer handling, the CLI resolver, and projection validation (nothing
  else in projections).
- **Unchanged:**
  - the `sdd` provider (byte-identical production code);
  - the readiness gate's logic;
  - this repository's trailer gate script (round-2 F3);
  - `records` / `specdir` resolution, except D2's all-or-nothing on a commit that also carries an `sdd`
    value, and unless the owner chooses D4.

Entry conditions (sdd SPEC §14):
1. Token → D1, with a VD landing with M3.
2. Anti-spoofing → §4.2.
3. Per-commit handling → §4.1 and §4.4.
4. Sites → §5.
5. Readiness gate → §4.5 and M3-INV-8.

## 3. Owner decisions

- **D1 — token** (unchanged from draft 1). Recommended: `Intent: sdd:<name>`, with `<name>` matching
  `^[a-z0-9][a-z0-9-]{1,99}$`. Alternative: `sdd:CI-…`.
- **D2 — an unbindable `sdd` claim.** Recommended:
  - per commit, all-or-nothing: the commit's event carries no refs at all;
  - the claim is reported as a `ClaimSkip` (not ledgered);
  - the sync succeeds.

  Alternative: whole-sync failure, the wedge (F1).
- **D3' — ordering.** Recommended: the ledger check (§4.3). Alternatives:
  - withdrawn D3 (auto-sync sdd first), which F2 and F3 rule out;
  - a dangling-tolerant projection, which hides real dangling refs.
- **D4 — the pre-existing records/specdir hazard (F8).** Recommended: file a deferral now and fix it
  in its own change with the same ledger check, after M3 proves it on `sdd`. Alternative: fold it into
  M3, which widens M3 to change INV-29 for `records` / `specdir`.

## 4. Contract

### 4.1 Trailer values

- `records:<CI-id>`, `specdir:<CI-id>`, and bare `<CI-id>` resolve as today, including whole-sync
  failure on a malformed value. The one change is D2: on a commit that also carries an unbindable
  `sdd` claim, their refs are dropped too (INV-29 amended, §5).
- **Any value whose first four bytes ASCII-case-fold to `sdd:`** is an `sdd` claim (round-2 F9):
  - If the rest matches D1's grammar, it is resolved (§4.2-§4.4).
  - Otherwise it is a claim skip with cause `malformed`. It is never a sync failure.

### 4.2 Resolution and anti-spoofing (CLI resolver, `sdd` arm)

- The `sdd` provider is built **lazily**, on the first `sdd` claim. A history without `sdd:` values
  never constructs it.
- The resolver calls `sdd.Provider.Resolve(ctx, commit, name)`. A revision `r` is accepted only if
  **all** of these hold:
  - `r.Source == intent.sdd`;
  - `r.Kind == change_intent.recorded`;
  - `r.NativeID` matches `^CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`;
  - `r.ArtifactSHA256` is 64 hex;
  - the **last segment** of `r.ArtifactPath` equals `name`, or matches
    `^[0-9]{4}-[0-9]{2}-[0-9]{2}-` followed by `name`.
- A revision failing any of these is a **spoof**: a code defect, not tenant data, so the sync fails.
- **Residual (round-2 F5):** the segment check cannot tell an active change named `2024-01-01-foo`
  from an archived `foo`, and it does not tie `NativeID` to `name`. It is defence in depth. The
  guarantee is `Resolve`'s exact-name semantics (SDD-INV-59): `Resolve(c, "foo")` never returns the
  active `2024-01-01-foo`'s revision.
- The connector's own check for `sdd` refs: `Provider == "sdd"`, `RecordID` CI-shaped, digest 64 hex.

### 4.3 The ledger check (D3')

- Once per `sync git`, **lazily on the first `sdd` claim** (round-2 F7), the CLI reads the set `L` of `(native id, artifact_sha256)` for every ledgered
  `intent.sdd` `change_intent.recorded` event.
- An accepted revision binds only if `(r.NativeID, r.ArtifactSHA256) ∈ L`. Otherwise the claim is
  skipped with cause `not-ledgered`.
- **Consequence:** no commit event ever carries an `sdd` ref to an un-ledgered revision, so the
  projection's dangling check (`projections.go:218-226`) cannot fire for `sdd`, whatever the refs do
  between syncs (F2).
- **Recovery:** run `sync intent sdd`, then `sync git`. The commit's next event (a new content sha, so
  a new idempotency key) carries the ref.
- **Why `L` is safe as a snapshot** (round 2 confirmed): `sync git` appends only git events, and the
  ledger is append-only, so every member of `L` has a lower sequence number than any commit event the
  same sync appends. Every ledgered `intent.sdd` change intent has a `change_intents_view` row
  (`reduceIntent`), keyed exactly as `L` is (`ValidateRevision` ties payload id and digest to the
  envelope).

### 4.4 Outcomes and the skip channel

- The git package gains:
  - `var ErrClaimUnbound` and the type `ClaimUnboundError{Cause string}`, which `errors.Is`-matches
    it;
  - `ClaimSkip{Commit, Value, Cause string}`;
  - `Result.ClaimSkips []ClaimSkip`.
- The resolver returns a `*ClaimUnboundError` for every tenant outcome. The connector turns it into a
  `ClaimSkip`, and on any skip it drops **all** of that commit's refs (D2).
- Any other resolver error fails the sync, as today.

| Outcome | Cause |
|---|---|
| value fails D1's grammar | `malformed` |
| `ErrUnknownChange` | `unknown-change` |
| `ErrArchivedEarlier` | `archived-earlier` |
| `ErrNoChangeIntent` | `no-change-intent` |
| `*UnmappableError` | `unmappable:<cause>` (incl. `slug-too-short`, F9) |
| accepted, not in `L` | `not-ledgered` |
| reader I/O, `ErrShallow`, cancellation, spoof | sync fails |

`sync git` and `sync all` print `claim-skip <cause> <value> at <commit>` after their counts line (the
value through `printableUnit`); `verify` prints the same lines in its git step's output. No skip changes
an exit status.

### 4.5 Projections and the readiness gate

- `commitPayload.validate` accepts provider `sdd`. This widens validation, so it is retroactively safe.
- `commit_intents_view` stays **insert-only**, as today (round-2 F1). A binding, once projected, is
  never withdrawn by a later event for the same commit, so `reduceObligationVerdict`'s lookup can never
  lose a row it found before.
- The readiness gate is unchanged. A commit whose claim was skipped on every sync has no rows, so it
  is not warranted.

## 5. Sites (complete; F7)

| Site | Change |
|---|---|
| `connector/git/connector.go` `resolvedCommit`, allowlist, spoof check, `Result` | §4.1, §4.2, §4.4 |
| `connector/git/SPEC.md` §2 (`Result`, `IntentResolver` contract), INV-29 | `sdd` claims fail per commit; INV-29's text amended for D2's all-or-nothing (round-2 F6); new INV rows (§6) |
| shared trailer fixture `scripts/tests/fixtures/intent-trailers/cases.tsv` | `sdd:` rows (valid, malformed, upper-case), all gate `fail` |
| kernel corpus oracle `connector/git/gitcmd/gitcmd_test.go::TestIntentTrailers_SharedCorpusKernelBinds` | "binds" = a ref is bound, not "the sync succeeded" (round-2 F2) |
| `cli/cli.go` `commitIntentResolver`, `syncGit`, the claim-skip output | §4.2-§4.4 |
| `cli/SPEC.md` CLI-INV-7 / CLI-INV-8 | the `sdd` arm and the claim-skip lines |
| `projections/projections.go` validate (:727) | accept provider `sdd` (§4.5) |
| `scripts/intent-delivery-applicability.sh` | **unchanged** (round-2 F3) |
| `kernel/internal/connector/intent/sdd/SPEC.md` §12, §14, SDD-INV-56 text | record M3 as landed; SDD-INV-56 unchanged in substance |
| generated pins `docs/verification/p6-event-route-matrix.md`, `p6-connector-reality.md` | regenerate after the code commits (their pins follow `cli`/`projections`/`connector`) |
| `docs/decisions/VD-…` (D1) | new |

## 6. Invariants (M3-INV-n)

1. **M3-INV-1** — A well-formed `sdd:<name>` at `c`, accepted and ledgered, binds exactly
   `Resolve(c, name)`'s revision.
2. **M3-INV-2** — A resolver revision failing any §4.2 check fails the sync.
3. **M3-INV-3** — Every tenant outcome in §4.4's table yields one `ClaimSkip`, and the commit event
   carries **no** refs. Other commits are unaffected, and the sync succeeds.
4. **M3-INV-4** — Reader errors and cancellation fail the sync.
5. **M3-INV-5** — No ledgered commit event names an `sdd` revision that is not itself ledgered at a
   lower sequence number. A projection rebuild after any interleaving of `sync git` and
   `sync intent sdd` succeeds.
6. **M3-INV-6** — A history with no `sdd` value neither constructs the `sdd` provider nor reads the
   ledger set `L` during `sync git`.
7. **M3-INV-7** — `records` / `specdir` behaviour and the pinned v2 payload vector are unchanged for
   commits without `sdd` values.
8. **M3-INV-8** — The readiness gate passes a commit bound to a SATISFIED `sdd` target with a non-owner
   VERIFIABLE review. It refuses:
   - a commit whose only claim was skipped;
   - a commit with one skipped and one otherwise-satisfied claim.
9. **M3-INV-9** — This repository's trailer gate still refuses every `sdd:` value, and everything it
   refused before.
10. **M3-INV-10** — Bindings are monotone: a `commit_intents_view` row, once projected, survives every
    later event for its commit, and an obligation verdict citing it always projects.

## 7. Test derivations

| INV | Test (where) | Kills | Controls |
|---|---|---|---|
| 1 | git: fake resolver binds `sdd:add-auth` (unit); cli: real sdd tenant + embedded store via `cli.Run` | binding the trailer value instead of the revision id | positive: a `records` claim still binds |
| 2 | git unit with a resolver returning a wrong-segment / wrong-kind / wrong-source revision | dropping any one §4.2 clause (one case per clause) | neutral: a correct revision binds |
| 3 | git unit, one case per §4.4 row, two-claim commits | whole-sync failure; partial binding | a clean commit beside the skipped one is appended with its ref |
| 4 | git unit: resolver returns a plain error | reading every error as a skip | — |
| 5 | cli + embedded store: `sync git` → rebuild; `sync intent sdd` → `sync git` → rebuild; then delete the branch and repeat | skipping the ledger check | the bound commit's row exists after the second order |
| 6 | cli: injected counters for construction and for the `L` read, over a history with no `sdd` value | eager construction; eager ledger read | each counter is 1 once a single `sdd:` value is added |
| 7 | existing git INV-28..30 tests and §5.1 vector, unchanged | any drift in v2 bytes | — |
| 8 | cli + embedded store (`cli.Run`, no build tag): sdd tenant fixture → `sync intent sdd` → `sync git` → `intent check --commit` for a bound commit (reports the `sdd` target), a skipped-only commit (refused), and a two-claim commit with one skipped (refused) | keeping the bindable refs of a commit with a skipped claim | the bound commit's report names its `sdd` revision |
| 9 | `scripts/tests` and the kernel oracle over `cases.tsv` | widening the gate to `sdd:`; the oracle counting sync success as binding | old rows unchanged |
| 10 | a non-tagged test that opens the embedded store directly. It appends a change intent, a git event binding it, a second git event for the same commit **without** the ref (what a later kernel version could append; one kernel version cannot, because `L` only grows and `Resolve` is deterministic), and a v2 verdict citing the binding, then runs Apply and Rebuild | deleting a commit's rows on a newer commit event | both succeed and the row is present |

Every new test is calibrated before it is believed: a hand mutant per "kills" cell must make it fail,
and a neutral edit must leave it passing.

## 8. Proposed change intent (owner approves the wording)

`CI-sdd-warranty-m3-<suffix>`: implements `BR-sdd-warranty-i9miv7` O-8, and O-10 for claims, at the
§5 sites. Non-goals as in §9.

## 9. Non-goals

- Any change to the `sdd` provider; M2 (Spec Kit); M4 (drift).
- The records/specdir ordering hazard (D4), unless the owner folds it in.
- Ledgering claim skips.
- Withdrawing a binding when a later sync or kernel version would no longer bind it (bindings are
  monotone, M3-INV-10). Reporting such drift belongs with M4.
