# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 1)

**Status:** draft 1, 2026-11-18. **Not yet reviewed.** Written by the authoring session; it needs a
non-author review (Law 9) and the owner's decisions D1-D3 (§3) before any code lands (Law 6).
**Parent contract:** `kernel/internal/connector/intent/sdd/SPEC.md` (R2 M1, accepted). Its §6 defines
the `Resolve` outcomes, its §12 names the four M3 sites, and its §14 lists five M3 entry conditions.
This draft answers each entry condition in §2. **Requirement served:** `BR-sdd-warranty-i9miv7` O-8,
the readiness gate on `sdd` chains, through the lifecycle step 3 of SPEC §8: "commit with
`Intent: sdd:X` → git connector binds commit → CI-X@exact revision".

## 0. Evidence first (read from HEAD `b18e672`, not recalled)

| Fact | Where | Consequence for M3 |
|---|---|---|
| The git connector resolves every `Intent:` trailer before it appends the commit event, and **any** resolution error fails the whole sync | `connector/git/connector.go` `resolvedCommit` (called from `Sync` at :128); git SPEC INV-29 | One unresolvable `sdd` claim anywhere in history wedges every git sync forever (§14 entry condition 3) |
| The provider allowlist is `records` / `specdir`; the id must match `^CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`; the resolver must return `RecordID == id` | `connector.go:197-210`, `intentIDRE` at :21 | An `sdd:` trailer names a **native change name**, not a CI id, so the spoof check cannot be "returned id equals trailer id" for `sdd` |
| The CLI resolver switches on `records` / `specdir` and is built inside `syncGit` | `cli/cli.go:212-261` | `sdd` needs a third arm and a provider |
| Projection validation accepts only `records` / `specdir` refs with a CI-shaped id | `projections/projections.go:725-729` | Widening it to `sdd` is retroactively safe: it accepts more, rejects nothing it accepted before (SPEC §10 I) |
| A commit event whose ref has no `change_intents_view` row for `(intent.<provider>, id, digest)` **fails the projection** ("dangling intent revision") | `projections.go:218-226` | Projections run in ledger sequence order. If a git event carrying an `sdd` ref is appended **before** its change-intent revision, every later rebuild fails, and no later sync can repair it, because the commit event already has the lower sequence number. **This is a permanent-wedge path that M3 must close (§4.3)** |
| `sync all` syncs intent `all` (records + specdir, **not** sdd, SDD-INV-56) and then git; `sync git` syncs git alone | `cli.go:655-670` | Neither path ledgers `sdd` revisions before a git event that might cite them |
| The commit-intent projection routes a ref to source `"intent." + Provider` | `projections.go:219` | `sdd` maps to `intent.sdd` = `core.SourceIntentSDD` with no new routing |
| The trailer gate's value grammar is `^((records\|specdir):)?CI-…$` | `scripts/intent-delivery-applicability.sh:93` | A tenant `sdd:` trailer is refused by this repository's own gate today |
| All 147 change names in the reference OpenSpec history match `^[a-z0-9][a-z0-9-]*$`; the longest is 42 bytes | probe over `rev-list --all` of the reference clone | A strict trailer grammar (D1) excludes no real name |
| `Resolve` is total over tenant states; its non-revision outcomes are `ErrUnknownChange`, `ErrArchivedEarlier`, `ErrNoChangeIntent`, `*UnmappableError`, else a reader error | sdd SPEC §6, SDD-INV-59 | M3 can classify every claim outcome without reading tenant bytes itself |
| `Resolve(c, n)` returns only revisions that `Revisions` emits | SDD-INV-31 | A resolved `sdd` ref always has a revision that a `sync intent sdd` would ledger |

## 1. Purpose and boundary

M3 lets a tenant commit carry `Intent: sdd:<change-name>` and binds it, at that commit's own tree, to
the exact `intent.sdd` change-intent revision, so the existing readiness gate (O-8) can judge the
commit. M3 changes: the git connector's trailer handling, the CLI resolver, projection validation, and
the trailer gate script. It does **not** change the `sdd` provider (its production code stays
byte-identical to the accepted tree), the readiness gate's logic, or `records` / `specdir` behaviour.

## 2. The five M1 entry conditions, answered

1. **The `sdd:` trailer token and its decision record** → §3 D1 (owner decision; a VD lands with M3).
2. **Anti-spoofing against a native name, including the profile's `kit`** → §4.2. The git connector
   never reads tenant files itself: it receives `(provider, record id, digest)` from the resolver, and
   the resolver accepts an `sdd` revision only if it is a change intent from `intent.sdd` whose
   artifact path names exactly the claimed change. `kit` needs no separate check in M1-scope: the
   strict profile admits only `"kit": "openspec"` (TestM1IsOpenSpecOnly), so any resolved revision is
   OpenSpec's. M2 must revisit this when a second kit exists.
3. **Per-commit handling of an unresolvable claim** → §3 D2 and §4.4.
4. **The four M3 sites** → §5 (complete list, with the two this draft adds).
5. **The readiness gate on `sdd` chains** → §4.5: unchanged logic; §6 M3-INV-8 proves it applies.

## 3. Decisions for the owner

- **D1 — trailer token.** Recommended: `Intent: sdd:<change-name>` with `<change-name>` matching
  `^[a-z0-9][a-z0-9-]{1,99}$` (the native name a developer knows; every real name fits, §0).
  Alternative: `Intent: sdd:CI-<slug>-<suffix>`, the derived id. It keeps one id grammar everywhere but
  asks developers to compute an id they never see. Names outside the grammar cannot be claimed; the
  provider still ingests them.
- **D2 — an unresolvable `sdd` claim.** Recommended: **per-commit, fail-closed on trust, never on the
  sync.** The commit event is appended **without** that ref, and the claim is reported as a git
  `ClaimSkip{Commit, Value, Cause}` (not ledgered, like sdd Skips). The commit therefore carries no
  binding and can never be warranted through it. Alternative: keep INV-29's whole-sync failure for
  `sdd` too, which wedges every git sync on any historical bad claim (23 dual-presence commits exist
  in the reference history alone). `records` / `specdir` keep INV-29 unchanged either way.
- **D3 — ledger ordering.** Recommended: before the git sync appends any commit event that carries
  an `sdd` ref, the CLI ensures every `intent.sdd` revision is ledgered by running the `sdd` intent
  sync on the same store first — **only when** the git history has at least one `sdd:` trailer
  (§4.3). Alternative: make the projection tolerate a dangling `sdd` ref. That changes projection
  semantics for every source and hides real dangling refs, so it is not recommended.

## 4. Contract

### 4.1 Trailer grammar (git connector)

- A trailer value is `records:<CI-id>`, `specdir:<CI-id>`, bare `<CI-id>` (= `records`), or
  `sdd:<name>` with `<name>` per D1. Anything else is malformed and, as today, fails the whole sync
  (malformed values never leave a committed repository's own gate; this does not change INV-29).
- For `sdd`, the connector passes `<name>` to the resolver as the record argument.

### 4.2 Resolution and anti-spoofing (CLI resolver, `sdd` arm)

- `sdd.Provider.Resolve(ctx, commit, name)`; on a revision `r`, the resolver returns
  `IntentRef{Provider: "sdd", RecordID: r.NativeID, ArtifactSHA256: r.ArtifactSHA256}` **only if**
  `r.Source == intent.sdd`, `r.Kind == change_intent.recorded`, `r.NativeID` matches the CI grammar,
  and `r.ArtifactPath` is `<root>/<name>` or `<archive>/YYYY-MM-DD-<name>` for the profile in force.
  Otherwise it returns an error, which the connector treats as a spoof (whole-sync failure: a spoof is
  a code defect, not tenant data).
- For `sdd`, the connector's spoof check is: `ref.Provider == "sdd"`, `RecordID` matches the CI
  grammar, digest is 64 hex. (`RecordID == trailer value` cannot hold, §0 row 2.)

### 4.3 Ordering (D3)

- `syncGit` builds the `sdd` provider **lazily**, on the first `sdd:` claim it must resolve. A
  repository with no `sdd:` trailer never constructs it (SDD-INV-56's spirit extends to `sync git`).
- When the provider is first built, `syncGit` runs the `sdd` intent sync on the same store **before**
  returning the first `sdd` ref, so every revision a ref can name (SDD-INV-31) is ledgered with a lower
  sequence number than the commit event that cites it.
- If that intent sync fails, the git sync fails (an I/O or store failure, not tenant data).

### 4.4 Per-commit claim outcomes (D2)

| `Resolve` outcome | Result for the commit |
|---|---|
| a revision passing §4.2 | ref bound |
| `ErrUnknownChange`, `ErrArchivedEarlier`, `ErrNoChangeIntent`, `*UnmappableError` | no ref; `ClaimSkip` with cause `unknown-change`, `archived-earlier`, `no-change-intent`, or `unmappable:<cause>` |
| anything else (reader I/O, `ErrShallow`, cancellation) | the whole sync fails, as today |

Other trailers on the same commit are unaffected. The CLI prints `claim-skip <cause> <value> at
<commit>` lines after the counts line, quoted per `printableUnit`, and exits 0.

### 4.5 Readiness gate

Unchanged. A bound `sdd` ref is a `commit_intents_view` row with provider `sdd`; the gate reads it like
any other. A commit whose only claim was skipped has no row and is not warranted.

## 5. Sites (complete list)

| Site | Change |
|---|---|
| `connector/git/connector.go` `resolvedCommit` + allowlist + spoof check | §4.1, §4.2, §4.4 |
| `connector/git/SPEC.md` INV-29 | amended: `sdd` claims fail per commit (D2); others unchanged |
| `cli/cli.go` `commitIntentResolver` + `syncGit` | `sdd` arm (§4.2), lazy provider + ordering (§4.3), claim-skip output |
| `projections/projections.go:727` | accept provider `sdd` |
| `scripts/intent-delivery-applicability.sh:93` | accept `sdd:<name>` |
| `connector/git/connector.go` `Result` | gains `ClaimSkips` (new in this draft) |
| `scripts/p6-route-matrix.sh` / census rows, if the git connector's surface rows name providers | to be checked at build time (new in this draft) |

## 6. Invariants (proposed, M3-INV-n)

1. **M3-INV-1** — An `sdd:<name>` trailer at commit `c` binds exactly `sdd.Resolve(c, name)`'s revision.
2. **M3-INV-2** — A resolver revision whose source, kind, id grammar, or artifact path does not name
   the claimed change fails the sync (spoof).
3. **M3-INV-3** — Every tenant `Resolve` sentinel yields a `ClaimSkip` and no ref; the sync succeeds and
   other claims on the commit bind.
4. **M3-INV-4** — Reader errors and cancellation still fail the sync.
5. **M3-INV-5** — No commit event with an `sdd` ref is appended before the revision it names.
   Proven by a rebuild after `sync git` alone on a fresh store.
6. **M3-INV-6** — A repository without `sdd:` trailers never constructs the `sdd` provider during
   `sync git` (shallow clone: `sync git` behaves as before M3).
7. **M3-INV-7** — `records` / `specdir` behaviour is byte-identical: git SPEC INV-28..30 pass unchanged,
   and the pinned v2 payload vector is unchanged for commits without `sdd` refs.
8. **M3-INV-8** — The readiness gate passes an `sdd`-bound commit whose targets are SATISFIED with a
   non-owner VERIFIABLE review, and refuses one whose claim was skipped.
9. **M3-INV-9** — The trailer gate script accepts `sdd:<name>` per D1 and still refuses everything it
   refused before.

## 7. Proposed change intent (owner approves the wording)

`CI-sdd-warranty-m3-<suffix>`: implements `BR-sdd-warranty-i9miv7` O-8 (and O-10's "fails only a
claim on that state" for claims), at the sites in §5; non-goals as §8.

## 8. Non-goals

- Any change to the `sdd` provider; M2's Spec Kit; drift (M4).
- Fixing a **suspected, unverified** pre-existing `records` ordering hazard: a `sync git` run alone
  after a new records CI might also dangle. Not yet reproduced; the review should confirm or refute it
  before it is raised to the owner.
- Ledgering claim skips.
