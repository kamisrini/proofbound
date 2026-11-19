# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 2)

**Status:** draft 2, 2026-11-18. It answers non-author round 1,
[`r2-m3-spec-review-round1-20261118`](../verification/verdicts/r2-m3-spec-review-round1-20261118.md)
(NEEDS_WORK: 1 HIGH, 7 MED, 3 LOW), over [draft 1](R2-sdd-m3-git-binding-spec-draft1.md). It needs a
non-author re-check (Law 9) and the owner's decisions D1-D4 (§3) before any code lands (Law 6).
**Parent contract:** `kernel/internal/connector/intent/sdd/SPEC.md` (R2 M1, accepted): §6 `Resolve`
outcomes, §12 M3 sites, §14 entry conditions. **Requirement served:** `BR-sdd-warranty-i9miv7` O-8
(readiness gate on `sdd` chains) and O-10's "fails only a claim on that state".

## 0. Round 1, answered

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
| F11 LOW — a claim outcome that changes across versions leaves stale bindings | §4.5: `commit_intents_view` holds the bindings of a commit's **latest** event only. A newer commit event replaces the commit's rows. This also makes the `not-ledgered` → bound recovery clean |

## 1. Evidence (read from HEAD `82a92ec`)

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
- **Changed:** git connector trailer handling, the CLI resolver, projection validation and
  `commit_intents_view` replacement, and the trailer gate script.
- **Unchanged:** the `sdd` provider (byte-identical production code), the readiness gate's logic, and
  `records` / `specdir` resolution unless the owner chooses D4.

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

- `records:<CI-id>`, `specdir:<CI-id>`, and bare `<CI-id>` behave exactly as today, including
  whole-sync failure on a malformed value (INV-29 unchanged).
- **Any value beginning `sdd:`** is an `sdd` claim:
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
- The connector's own check for `sdd` refs: `Provider == "sdd"`, `RecordID` CI-shaped, digest 64 hex.

### 4.3 The ledger check (D3')

- Once per `sync git`, the CLI reads the set `L` of `(native id, artifact_sha256)` for every ledgered
  `intent.sdd` `change_intent.recorded` event.
- An accepted revision binds only if `(r.NativeID, r.ArtifactSHA256) ∈ L`. Otherwise the claim is
  skipped with cause `not-ledgered`.
- **Consequence:** no commit event ever carries an `sdd` ref to an un-ledgered revision, so the
  projection's dangling check (`projections.go:218-226`) cannot fire for `sdd`, whatever the refs do
  between syncs (F2).
- **Recovery:** run `sync intent sdd`, then `sync git`. The commit's next event carries the ref (§4.5).

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

The CLI prints `claim-skip <cause> <value> at <commit>` after the counts line (value through
`printableUnit`), and exits 0.

### 4.5 Projections and the readiness gate

- `commitPayload.validate` accepts provider `sdd`. This widens validation, so it is retroactively safe.
- `commit_intents_view` is **replaced per commit**: when a commit event with a higher sequence than the
  commit's current row arrives, the commit's old intent rows are deleted before its refs are inserted
  (F11).
- The readiness gate is unchanged. A commit whose claim was skipped has no rows, so it is not
  warranted.

## 5. Sites (complete; F7)

| Site | Change |
|---|---|
| `connector/git/connector.go` `resolvedCommit`, allowlist, spoof check, `Result` | §4.1, §4.2, §4.4 |
| `connector/git/SPEC.md` §2 (`Result`, `IntentResolver` contract), INV-29 | `sdd` claims fail per commit; others unchanged; new INV rows (§6) |
| shared trailer fixture `scripts/tests/fixtures/intent-trailers/cases.tsv` (used by the git tests and the gate script test) | `sdd:` rows: valid, malformed, upper-case |
| `cli/cli.go` `commitIntentResolver`, `syncGit`, the claim-skip output | §4.2-§4.4 |
| `cli/SPEC.md` CLI-INV-7 / CLI-INV-8 | the `sdd` arm and the claim-skip lines |
| `projections/projections.go` validate (:727) and the commit-intent insert (:218-227) | §4.5 |
| `scripts/intent-delivery-applicability.sh` (:93) | accept `sdd:<name>` per D1 |
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
6. **M3-INV-6** — A history with no `sdd:` value never constructs the `sdd` provider during `sync git`.
7. **M3-INV-7** — `records` / `specdir` behaviour and the pinned v2 payload vector are unchanged for
   commits without `sdd` values.
8. **M3-INV-8** — The readiness gate passes a commit bound to a SATISFIED `sdd` target with a non-owner
   VERIFIABLE review. It refuses:
   - a commit whose only claim was skipped;
   - a commit with one skipped and one otherwise-satisfied claim.
9. **M3-INV-9** — The trailer gate accepts `sdd:<name>` per D1 and still refuses everything it refused
   before.
10. **M3-INV-10** — `commit_intents_view` holds only the latest event's refs for each commit.

## 7. Test derivations

| INV | Test (where) | Kills | Controls |
|---|---|---|---|
| 1 | git: fake resolver binds `sdd:add-auth` (unit); cli: real sdd tenant + embedded store via `cli.Run` | binding the trailer value instead of the revision id | positive: a `records` claim still binds |
| 2 | git unit with a resolver returning a wrong-segment / wrong-kind / wrong-source revision | dropping any one §4.2 clause (one case per clause) | neutral: a correct revision binds |
| 3 | git unit, one case per §4.4 row, two-claim commits | whole-sync failure; partial binding | a clean commit beside the skipped one is appended with its ref |
| 4 | git unit: resolver returns a plain error | reading every error as a skip | — |
| 5 | cli + embedded store: `sync git` → rebuild; `sync intent sdd` → `sync git` → rebuild; then delete the branch and repeat | skipping the ledger check | the bound commit's row exists after the second order |
| 6 | cli: an injected constructor counter over a history with no `sdd:` value | eager construction | counter is 1 once a single `sdd:` value is added |
| 7 | existing git INV-28..30 tests and §5.1 vector, unchanged | any drift in v2 bytes | — |
| 8 | gates unit over projected rows (single-claim pass, skipped refusal, two-claim refusal) | gate passing on a partial binding | — |
| 9 | `scripts/tests` over `cases.tsv` | regex widened beyond D1 | old refusal rows unchanged |
| 10 | projections unit: two events for one commit, the second with fewer refs | insert-only rows | — |

Every new test is calibrated before it is believed: a hand mutant per "kills" cell must make it fail,
and a neutral edit must leave it passing.

## 8. Proposed change intent (owner approves the wording)

`CI-sdd-warranty-m3-<suffix>`: implements `BR-sdd-warranty-i9miv7` O-8, and O-10 for claims, at the
§5 sites. Non-goals as in §9.

## 9. Non-goals

- Any change to the `sdd` provider; M2 (Spec Kit); M4 (drift).
- The records/specdir ordering hazard (D4), unless the owner folds it in.
- Ledgering claim skips.
