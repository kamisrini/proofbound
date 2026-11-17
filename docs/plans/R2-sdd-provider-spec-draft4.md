# `sdd` spec-driven-development provider contract — DRAFT 4 (change-scoped, history-native, per-state)

Status: **DRAFT 4 (R2 M0), not yet a live kernel contract.** Supersedes
[`R2-sdd-provider-spec-draft3.md`](R2-sdd-provider-spec-draft3.md) (reviewed NEEDS_WORK, 2 HIGH, in
`docs/verification/verdicts/r2-m0-review-round3-20261116.md`) and drafts 1-2. **Self-contained:** no
definition is inherited from a superseded draft. Becomes `kernel/internal/connector/intent/sdd/SPEC.md`
in the same M1 commit as the code and tests it cites, together with the intent SPEC §1, §2 and §4
amendments named in §12.

Scope (owner decision 2026-11-16, `VD-sdd-change-scoped-r2-k7q3wz`, unchanged): **R2 warrants OpenSpec
changes and Spec Kit features — the commits that implement a change, against the requirements that
change declares — read from the repository's history.** The main-spec lifecycle is not in R2.
Requirement: `BR-sdd-warranty-i9miv7` (revision proposed in §11).

## 0. Evidence first

Measured on a fresh full clone of `Fission-AI/OpenSpec` at `f085c850` (2026-11-17; one commit past the
reviewer's `8bd4b355`), with the round-3 reviewer's independent counts beside the author's. **Each count
below is printed by the probe that claims it** (`testdata/derive/`, §13); a count the probe does not
print is not evidence.

| Observation (real history) | Count | Consequence in this draft |
|---|---|---|
| Commits on all refs (`rev-list --all`, the git connector's walk) | 1624 (120 merges, 1 root) | The provider walks exactly this set (§3) |
| Distinct SDD states over those commits (tree id of `openspec/changes` + profile blob) | 338 | Mapping is per distinct state, so cost tracks states, not commits (§3) |
| Commits touching `openspec/changes`: HEAD path-limited / `--all --full-history` | 256 / 451 | Path-limited walks miss states; draft 4 uses no path limiting at all (§3) |
| **Change name present both active and archived** (leftover after archive, `7ced2a87`..`df12368f`) | **1 name; 9 commits on the HEAD path-limited walk, 12 on `--all --full-history`** (draft 3 claimed 0 — round-3 F1); archived twice: 0 | Unmappable **for that state only**; never a sync failure (§7) |
| Release tags pointing at one of those 12 commits | 0 | (recorded so the "no tip is wedged" reading is checkable) |
| Changes archived without ever being observed active (squash-merge) | 16 / 108 | A name resolves when archived **by** the commit (§6) |
| Changes with zero requirement entries at some commit | 34 | Absence, never an error (§5) |
| Within-change requirement removals / statement changes | 47 (10 changes) / 10 (5 changes) | Content-addressed records (§4) |
| Requirements with no statement line | 4 | Statement falls back to the name (§4.2) |
| RENAMED entries outside code fences | 0 | RENAMED follows OpenSpec's parser format, tested on fixtures only (§4.2) |
| `- FROM:` lines and a `## RENAMED Requirements` header **inside a code fence** (a documentation example in a MODIFIED requirement of `adopt-delta-based-changes`) | 47 lines over the walk | Fenced code is masked before any section or entry is read, as OpenSpec does (§4.2, SDD-INV-39) |

**How each round-3 finding is answered.** F1 (permanent wedge) → §7: every fail-closed condition is
scoped to one state and reported; only a claim on that state fails, at resolution. F2 (walk ≠ git
connector's walk) → §3: the walk *is* the git connector's commit list, with no path limiting, and
SDD-INV-31 asserts `Resolve(c, n) ∈ Revisions()` for every listed commit. F3 (O-11 unsatisfiable) →
O-11 is dropped from the BR proposal (§11). F4 (tests that cannot fail; titles that changed meaning) →
§9: the drifted invariants are retired and re-issued under new numbers, and each proving test names
the wrong implementation it kills. Beyond round 3: re-running the probe found fenced documentation
examples that an unmasked parser would read as renames (§4.2, SDD-INV-39). F5 (drift) → §4.3 defines `slug`/`suffix` here, empty name-parts
are unmappable, RENAMED is read as `FROM`/`TO` pairs, and §12 names the intent SPEC amendments.

## 1. Purpose and boundary

Provider name `sdd`; source `intent.sdd`. It maps committed SDD artifacts onto the **existing**
canonical model — `proofbound.requirement.v1` and `proofbound.change-intent.v1` — through one
declarative profile committed in the tenant repository. A third provider of the intent connector: no
new connector, model, or event kind.

Family rules: **no fabrication** (no business decision, authority, relation, obligation, status, or
identity is invented; governance maps to nothing; requirements render `authorization: undeclared`);
**exact bytes** (no normalization); **commit tree** (everything is read through the `Reader` at named
commits). Specific to this provider: Proofbound **never writes** into tenant SDD files, and in R2 the
**main specs are not ingested**.

**Not wired into the git connector in M1** (round-3 condition 1). M1 registers `intent.sdd` with the
intent connector only. The `sdd:` trailer token, its anti-spoofing check, and how the git connector
treats a claim that does not resolve are M3's (§6, §14).

## 2. Interface lock

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface {
    // Every commit the git connector ingests (§3), oldest first. One home: the argument list is the
    // git connector's, imported, never restated.
    Commits(ctx context.Context) ([]string, error)
    // The parents of commit, in order (empty for a root commit).
    Parents(ctx context.Context, commit string) ([]string, error)
    // The state key of commit: the git object ids of the profile blob and of every SDD root tree
    // the profile names, or the empty key when the profile is absent. Two commits with equal keys
    // have byte-identical SDD inputs.
    StateKey(ctx context.Context, commit string) (string, error)
    // Every committed file at commit under the profile's SDD roots, plus the profile.
    Read(ctx context.Context, commit string) ([]Artifact, error)
    // Names of the entries directly under one directory at commit (for the parent archive check, §6).
    List(ctx context.Context, commit, dir string) ([]string, error)
}
type Skip struct {
    Commit string // the first commit, in Commits order, that carries the skipped state
    Unit   string // "requirement <path>", "change <name>", or "profile"
    Err    error  // wraps intent.ErrUnmappable
}
type Provider struct { /* unexported */ }

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                             // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) // whole history (§3)
func (p *Provider) Skipped() []intent.Skip                                   // after Revisions (§7)
// Resolve maps a native change name at commit to that commit's change-intent revision (§6).
func (p *Provider) Resolve(ctx context.Context, commit, name string) (intent.Revision, error)
```

`Skip` lives in package `intent` (intent SPEC §2 amendment, §12); it is shown here for reading.
`Revisions` returns an error **only** for reader failures, context cancellation, or a broken
provider invariant (§7) — never for tenant content.

Profile path: `.proofbound/sdd-profile.json`. A commit without it contributes nothing (the tenant had
not opted in yet) and has the empty state key.

## 3. History-native ingestion — the git connector's walk, per distinct state

- **The commit set.** `Commits` returns exactly the commits the git connector ingests: today
  `git rev-list --reverse --date-order --exclude=refs/stash --exclude=refs/notes/* --exclude=refs/replace/* --all`
  (`gitcmd/gitcmd.go`). That argument list moves to one exported home in `gitcmd` and both connectors
  use it (Law 2). **No path limiting and no history simplification**: every listed commit is visited,
  merges and unmerged branches included. This closes round-3 F2 by construction rather than by
  choosing the right simplification mode.
- **Per distinct state.** For each commit, `StateKey` names its SDD input by git object ids. The
  provider maps each distinct key once (§4-§5) and returns the union of mapped revisions, de-duplicated
  by `(kind, record id, artifact digest)`. On OpenSpec: 1624 commits, 338 states.
- **The guarantee M3 relies on (SDD-INV-31).** For every commit `c` that `Commits` lists and every
  name `n`, if `Resolve(c, n)` succeeds its result is a member of `Revisions()`. It holds because both
  are the same function of `c`'s state, over the same commit set.
- **Ordering of syncs.** The guarantee is over one ref snapshot. Refs that move between the intent
  sync and the git sync are the existing provider-ordering rule (intent INT-INV-9: revisions are
  ingested before commit consumers); M3 owns what the git connector does when a claim's revision is not
  yet in the ledger (§14).
- **Idempotence.** Unchanged refs yield no new revisions (intent INT-INV-5).

## 4. Requirement records (content-addressed)

### 4.1 What is a requirement record

- **OpenSpec:** one record per delta file `<change-dir>/specs/<capability>/spec.md` that has at least
  one requirement entry, wherever the change directory sits (active under `changes/`, or under
  `changes/archive/YYYY-MM-DD-<name>/`).
- **Spec Kit:** one record per feature spec `specs/<feature>/spec.md` with at least one obligation.
- A file with zero entries yields no record (the canonical model requires ≥ 1 obligation).

### 4.2 Obligations

- **Code fences are masked first.** Every line from an opening ```` ``` ```` or `~~~` fence through its
  closing fence is invisible to the rules below, as OpenSpec's `buildCodeFenceMask` makes it. Real
  OpenSpec history contains a fenced `## RENAMED Requirements` example (§0); read unmasked, it would
  mint rename obligations nobody declared — fabrication.
- **Sections.** OpenSpec delta sections are matched with OpenSpec's own header rule
  (`## ADDED|MODIFIED|REMOVED|RENAMED Requirements`, case-insensitive, as in
  `src/core/parsers/spec-structure.ts`).
- **ADDED / MODIFIED / REMOVED.** Each `### Requirement: <name>` block is one obligation.
  **Statement:** the first non-empty line after the heading, before any further heading (for ADDED /
  MODIFIED the SHALL sentence, for REMOVED the `**Reason**` line), exact bytes without the line
  terminator; if there is none, `<name>`.
- **RENAMED.** OpenSpec writes renames as line pairs, not blocks:
  `- FROM: \`### Requirement: <old>\`` followed by `- TO: \`### Requirement: <new>\``. Each pair is one
  obligation; its **name** is `<old>` and its **statement** is the exact bytes of the `TO` line. An
  unpaired `FROM` or `TO` (OpenSpec's own parser refuses these) makes the record state unmappable.
- **Spec Kit:** `- **FR-001**` / `SC-` / `NFR-` lines → `O-FR-001`, statement = the line's statement
  group.
- **Obligation id (OpenSpec):** `O-<SEC>-<np>`, `SEC` ∈ {`ADD`, `MOD`, `REM`, `REN`}, where
  `np = namepart(name)`:
  1. upper-case ASCII; every byte outside `[A-Z0-9]` → `-`; runs of `-` collapsed; leading and
     trailing `-` trimmed;
  2. **empty → the record state is unmappable** (round-3 F5; `O-ADD-` would otherwise pass the
     canonical regex);
  3. longer than 28 bytes → its first 21 bytes, re-trimmed, + `-` + the first 6 upper-case hex
     characters of `SHA-256(name)`.

  `SEC` + `-` + `np` is at most 32 bytes, as the canonical obligation regex requires. Never positional.
- **Collisions:** two entries mapping to one obligation id in a record make the record state
  unmappable.
- **Status:** OpenSpec record under an active change → `proposed`; under the archive → `active`. Spec
  Kit → the profile's explicit status map, else its declared absent value, else unmappable.
- **Owner:** the first match in the profile's `owners` list (`VD-sdd-accountable-identity-ypcx7x`);
  none → unmappable.

### 4.3 Identity and digest

- **`slug(s)`:** lower-case ASCII; every byte outside `[a-z0-9]` → `-`; runs collapsed; leading and
  trailing `-` trimmed; truncated to 40 bytes and re-trimmed. Shorter than 2 bytes → unmappable.
- **`suffix(kind, key)`:** the first 6 bytes of `SHA-256(kit + "\x00" + kind + "\x00" + key)`, each
  byte `b` mapped to `"abcdefghijklmnopqrstuvwxyz0123456789"[b % 36]`. `kit` is the profile's `kit`
  value as committed (`openspec` or `speckit`); `kind` is the literal `requirement` or `change-intent`.
- **Manifest:** one line `"<sha256-hex-of-exact-bytes>  <path>\n"` per input file, sorted by path
  bytes. `artifact_sha256 = SHA-256(manifest)`.
- **Requirement record:** inputs = the delta (or feature spec) file **and the profile**;
  `artifact_path` = the source file's path; key = `<change-name>/<capability>` (OpenSpec, change name
  without the archive's `YYYY-MM-DD-` prefix) or the feature directory name (Spec Kit). Record id:
  `BR-<slug(key)>-<suffix("requirement", key + "\x00" + artifact_sha256)>`.
- **A record has exactly one revision.** Any byte change of the source file, its path (archiving
  moves it), or the profile mints a new record. Intent SPEC §5 holds by construction.
- **The payload is a pure function of the manifest's bytes.** Owner, status, and obligations derive
  only from the manifest's files and their paths (SDD-INV-26).
- **Record-id collision** — two distinct `(key, artifact_sha256)` pairs minting one record id — makes
  **both** states unmappable (SDD-INV-38). Probability is about 2⁻³¹ per pair that also shares a slug.
  Residual, stated rather than hidden: if the colliding state appears after the first was appended,
  the first stays in the ledger (append-only) but no longer resolves.

## 5. Change intents

- **OpenSpec:** one change intent per change **name**. Id
  `CI-<slug(name)>-<suffix("change-intent", name)>`: stable across the change's life. At a state it
  `implements` every obligation of every requirement record of that change at that state.
- **Spec Kit (`feature-sibling`):** one change intent per feature directory, only when the declared
  sibling (`tasks.md`) exists; it implements every obligation of the feature's record.
- **Revisions:** successive states yield successive revisions of one id. `supersedes` stays `[]`
  (`DEF-INTENT-CI-REVISION-001`).
- **Artifact:** manifest over every file in the change directory at that state **plus the profile**;
  `artifact_path` = the change directory.
- **Status:** active → the profile's declared `changes.status`; archived → `accepted`. Sponsor: the
  first match in the profile's `sponsors` list; none → unmappable.
- **No requirement records → no change intent.** Absence, not an error (SDD-INV-28).
- **Name reuse** (a name both active and archived, or archived twice, at one state) → that change's
  state is unmappable (SDD-INV-29). Its requirement records are unaffected: their paths differ, so
  they are distinct records.
- **Cascade:** a change-intent state is unmappable if any requirement record state it would target is
  unmappable. Without this, the intent connector's batch check would reject a dangling target and the
  whole sync would fail (SDD-INV-36).

## 6. Resolution

`Resolve(c, n)` returns the change intent of `n` at `c`'s state when:

1. **active:** `changes/<n>/` exists at `c` and `n` is mappable there; or
2. **archived by `c`:** exactly one `changes/archive/*-<n>/` exists at `c`, and **no parent** of `c`
   has any `changes/archive/*-<n>/`. This is the squash-merge workflow (16 / 108 real changes).
   - **A merge** whose side branch archived `n` does *not* archive it, because one parent already has
     the directory. The archiving commit is the side-branch commit (OpenSpec `98d90cc0` merging after
     `ccaa5ad0`).
   - **A root commit** has no parents, so an archive present at the root is archived by the root.
     That is consistent with the rule and is tested.

Every other case returns an error: unknown name, archived earlier, or **unmappable at `c`** (wrapping
`intent.ErrUnmappable`, e.g. the 12 OpenSpec dual-presence commits). `Resolve` reads only `c`'s state
and the archive listings of `c`'s parents (SDD-INV-33).

**Verdicts bind to the change.** An `obligation-verdict.v2` names the commit, the exact change-intent
revision, and that revision's targets — what projections accept today. There is no carry-over step.

## 7. Fail-closed per state, never per history

History-native ingestion turns any batch-fatal rule into a permanent wedge, because past commits
cannot be edited (round-3 F1). So:

- **Units.** A requirement record state (its manifest), a change-intent state (`(name, state key)`),
  or a whole state (an invalid profile makes every unit at that state unmappable).
- **Every content condition is per unit:** invalid profile, uncovered owner or sponsor, obligation-id
  collision, empty name-part, slug shorter than 2, unpaired RENAMED line, unmapped Spec Kit status,
  name reuse, record-id collision, cascade, and a candidate revision that fails
  `intent.ValidateRevision`. The unit contributes **no revision** and **one `Skip`**, keyed by its
  first commit in walk order. Every other unit at every commit is still emitted.
- **Batch-safe by construction.** Every returned revision has already passed
  `intent.ValidateRevision`, and the batch has no duplicate `(source, id, digest)` and no dangling
  same-source reference. So tenant content can never trip the intent connector's batch check
  (INT-INV-4 holds unchanged). If the batch check does fire on `sdd` revisions, that is a provider bug
  and fails the sync, as it should (SDD-INV-37).
- **Failure moves to the claim.** Only a claim on an unmappable state fails, at `Resolve`.
- **Skips are derived, not ledgered.** They are a deterministic function of the refs, so they are
  recomputed on every sync, and the intent connector's `Result` reports them (intent SPEC §2
  amendment). The CLI prints them and exits 0. A non-zero exit on a historical state would be the
  wedge again, now in the scheduler. R2 adds no event kind for skips.
- **The unmappable fixture.** `testdata/unmappable/*` now yields its skips and the batch of everything
  else. It no longer yields a sync error, so SDD-INV-14 ("stops") is retired. The intent family's own
  STOP (INT-INV-11) concerns the `records`/`specdir` foreign fixture and is unchanged.

## 8. Lifecycle (reference)

```
 OpenSpec (tenant repo)                     Proofbound
 1 propose changes/X/      ── sync ───────▶ BR records for X's deltas (proposed) + CI-X
 2                                          requirement-review.v1 on those records ⇒ VERIFIABLE   [R1]
 3 implement; commit with  ───────────────▶ git connector binds commit → CI-X@exact revision      [M3]
     Intent: sdd:X                          (resolved at that commit; branch or main)
 4 tests / CI run          ───────────────▶ check + workflow witness events                       [R1]
 5                                          obligation-verdict.v2 on that commit ⇒ SATISFIED      [R1]
 6                                          intent-delivery-readiness ⇒ PASS = warranted           [R1]
 7 edit X's deltas         ── sync ───────▶ new BR records; CI-X's next revision targets them
 8 openspec archive X      ── sync ───────▶ archived-location records (active); CI-X accepted
 ◀── feedback (read-only): reports, readiness result, status export (M4; change drift is NOT in R2's BR)
```

## 9. Invariants

IDs are permanent from draft 1 on (pinned in `docs/invariants.lock` once the file lands): new
invariants append, none is reused, and **an invariant whose meaning changes is retired and re-issued
under a new number** (round-3 F4).

1. **SDD-INV-1 — The profile is strict JSON and fails closed.**
2. **SDD-INV-2 — Only requirements and change intents are emitted; governance maps to nothing.**
3. **SDD-INV-3 — Accountable identity comes only from the profile at the same tree.**
4. **SDD-INV-4 — Status comes only from the explicit map or a declared absent value.**
5. **SDD-INV-5 — Obligation ids are never positional.**
6. **SDD-INV-6 — Duplicate or missing obligations fail closed.**
7. **SDD-INV-7 — Record ids are deterministic.**
8. **SDD-INV-8 — RETIRED (draft 4): the digest is a manifest digest; replaced by INV-34.**
9. **SDD-INV-9 — Delta sections map onto the closed relation vocabulary.**
10. **SDD-INV-10 — Feature-sibling change intents are feature-level only.**
11. **SDD-INV-11 — RETIRED (draft 4): resolution also reads parent archive listings; replaced by INV-33.**
12. **SDD-INV-12 — Canonical payload bytes are stable across runs.**
13. **SDD-INV-13 — Two kits run on one provider with only the profile differing.**
14. **SDD-INV-14 — RETIRED (draft 4): unmappable content no longer stops the sync; replaced by INV-35/36.**
15. **SDD-INV-15 — RETIRED (draft 3).**
16. **SDD-INV-16 — RETIRED (draft 3).**
17. **SDD-INV-17 — RETIRED (draft 3).**
18. **SDD-INV-18 — A change's artifact digest covers the exact bytes of every file it reads.**
19. **SDD-INV-19 — RETIRED (draft 3).**
20. **SDD-INV-20 — Archiving a change never makes a later sync fail.**
21. **SDD-INV-21 — RETIRED (draft 3).**
22. **SDD-INV-22 — RETIRED (draft 3).**
23. **SDD-INV-23 — RETIRED (draft 3).**
24. **SDD-INV-24 — RETIRED (draft 4): the walk is no longer HEAD-scoped; replaced by INV-31.**
25. **SDD-INV-25 — A requirement record has exactly one revision; its id is content-addressed.**
26. **SDD-INV-26 — The payload is a pure function of the manifest bytes.**
27. **SDD-INV-27 — RETIRED (draft 4): "archived by" now means absent from every parent; replaced by INV-32.**
28. **SDD-INV-28 — A commit with no requirement records for a change yields no change intent, not an error.**
29. **SDD-INV-29 — A reused change name is unmappable.**
30. **SDD-INV-30 — Main specs are not ingested in R2.**
31. **SDD-INV-31 — Every commit the git connector lists contributes its state, and every successful Resolve is a member of Revisions.**
32. **SDD-INV-32 — A name resolves only to a change active at the commit, or archived by it and in none of its parents.**
33. **SDD-INV-33 — Resolution reads only the commit's state and its parents' archive listings.**
34. **SDD-INV-34 — The artifact digest is the SHA-256 of the manifest of exact committed bytes.**
35. **SDD-INV-35 — No tenant content at any commit makes Revisions fail.**
36. **SDD-INV-36 — An unmappable unit yields one Skip and no revision, and its dependants cascade.**
37. **SDD-INV-37 — Every returned batch passes the intent connector's batch validation.**
38. **SDD-INV-38 — Two states minting one record id are both unmappable.**
39. **SDD-INV-39 — Nothing inside a code fence is read as a section, entry, or rename.**

Each proving test states **the wrong implementation it kills** (the positive control, Law 9
corollary) and, where it matters, **a neutral change that must survive**.

| Invariant | Statement | Proving test | Kills / survives |
|---|---|---|---|
| SDD-INV-1 | Unknown, missing, or null fields, unknown modes, and regexes lacking required groups make that state unmappable | sdd_test.go::TestProfileIsStrict | kills: a lenient JSON decoder |
| SDD-INV-2 | No business decision is emitted; requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted | kills: mapping `proposal.md` to a BD |
| SDD-INV-3 | Owner and sponsor come only from their list's first match at the same state; uncovered is unmappable | sdd_test.go::TestIdentityOnlyFromProfile | kills: falling back to the git author |
| SDD-INV-4 | A Spec Kit status not in the map and with no declared absent value is unmappable | sdd_test.go::TestStatusMapFailsClosed | kills: defaulting to `proposed` |
| SDD-INV-5 | Inserting, deleting, or reordering one entry changes no other obligation id | sdd_test.go::TestObligationIDsAreNotPositional | kills: index-based ids |
| SDD-INV-6 | Colliding ids, empty name-parts, and unpaired RENAMED lines are unmappable; a zero-entry file yields no record | sdd_test.go::TestObligationCollisionAndEmptyFailClosed | kills: last-wins on collision; `O-ADD-` |
| SDD-INV-7 | The §4.3 / §5 id vectors in testdata (including `kit` and `kind` literals) match exactly | sdd_test.go::TestRecordIDsAreDeterministic | kills: a suffix omitting `kind` |
| SDD-INV-9 | Every ADDED, MODIFIED, REMOVED, RENAMED entry is an `implements` target | sdd_test.go::TestDeltaSectionsMapToClosedRelations | kills: dropping REMOVED |
| SDD-INV-10 | One feature-level change intent, only when the sibling exists | sdd_test.go::TestFeatureSiblingIsFeatureLevel | kills: per-task intents |
| SDD-INV-12 | Repeated runs over one history emit byte-identical payloads | sdd_test.go::TestCanonicalizationStability | kills: map-ordered obligations |
| SDD-INV-13 | (a) Both reference fixtures map through one constructor. (b) Speckit-shaped files planted in an `openspec`-kit repository are ignored, and flipping only the profile's `kit` changes the mapping | sdd_test.go::TestTwoKitsOneProvider | kills: kit sniffing from the tree |
| SDD-INV-18 | The change manifest changes when any file of the change directory or the profile changes | sdd_test.go::TestChangeManifestDigest | kills: a manifest over delta files only |
| SDD-INV-20 | The leftover-after-archive state (real `7ced2a87` shape), a `--skip-specs` archive, and a deleted main spec all sync | sdd_test.go::TestArchiveNeverBreaksSync | kills: batch-fatal name reuse |
| SDD-INV-25 | Editing a delta file yields a new record id and leaves the earlier record unchanged | sdd_test.go::TestRequirementRecordsAreContentAddressed | kills: name-keyed BR ids |
| SDD-INV-26 | (a) Changing the profile's owner changes the digest. (b) Changing a non-manifest input (a main spec, another change's files, `src/`, the git author or date) leaves every revision byte-identical | sdd_test.go::TestPayloadIsPinnedByManifest | kills: reading outside the manifest; survives: non-manifest edits |
| SDD-INV-28 | A change with only `proposal.md` yields no change intent and no error | sdd_test.go::TestEmptyChangeIsAbsent | kills: emitting a target-less CI |
| SDD-INV-29 | A name both active and archived, or archived twice, at one state is unmappable for that change at that state only | sdd_test.go::TestReusedNameIsUnmappable | kills: picking the active copy |
| SDD-INV-30 | Files under the main-spec path never produce a revision | sdd_test.go::TestMainSpecsAreNotIngested | kills: ingesting `openspec/specs` |
| SDD-INV-31 | Over the `branches` fixture (unmerged-branch claim, TREESAME `--no-ff` merge, root commit): `Commits` equals the git connector's list, and for every listed `c` and every name `n` where `Resolve(c, n)` succeeds, the result is in `Revisions()` | sdd_test.go::TestResolveIsMemberOfRevisions | kills: a HEAD-only walk; a path-limited walk (TREESAME) |
| SDD-INV-32 | Active, archived-by (single parent), archived-on-side-branch-then-merged (merge fails, side commit resolves), root-with-archive (resolves), dual presence (ErrUnmappable), archived earlier (fails), unknown (fails) | sdd_test.go::TestNativeNameResolution | kills: a first-parent-only archive check |
| SDD-INV-33 | A reader that fails on any other commit, and on any parent read other than an archive listing, proves Resolve's reads | sdd_test.go::TestResolveReadSet | kills: resolving from a cached HEAD state |
| SDD-INV-34 | The digest equals a manifest recomputed independently from `git cat-file` bytes; a one-byte edit of any manifest file changes it | sdd_test.go::TestDigestIsManifestOfExactBytes | kills: normalizing line endings |
| SDD-INV-35 | Each §7 cause, placed at a non-tip commit and fixed later, leaves `Revisions` without error and every other unit emitted | sdd_test.go::TestHistoricalFailClosedIsRecoverable | kills: a batch-fatal implementation (round-3 F1) |
| SDD-INV-36 | Each cause yields exactly one Skip wrapping `intent.ErrUnmappable` and no revision; a CI targeting a skipped BR is skipped too; Resolve on that state fails | sdd_test.go::TestUnmappableUnitsSkipAndCascade | kills: a missing cascade (a dangling target fails the batch) |
| SDD-INV-37 | Every fixture history's batch passes `Sync` against a test appender; the hostile fixtures included | sdd_test.go::TestBatchPassesIntentValidation | kills: returning a revision that fails ValidateRevision |
| SDD-INV-38 | A pinned colliding key pair (found offline by search, recorded in testdata) makes both states unmappable | sdd_test.go::TestRecordIDCollisionFailsClosed | kills: first-wins on collision |
| SDD-INV-39 | A delta whose MODIFIED requirement contains the real fenced RENAMED example yields exactly its MODIFIED obligation; removing the fence lines yields a REN obligation | sdd_test.go::TestCodeFencesAreMasked | kills: an unmasked parser; the fence-removed variant proves the case discriminates |

Retired invariants have no row; their tombstones are the numbered list above.

## 10. Decisions this draft settles or withdraws

- **A (warranty across archive): withdrawn** (draft 3); unchanged.
- **B (main-spec obligation lifecycle): withdrawn from R2** (draft 3); unchanged.
- **C (Spec Kit lifecycle): answered** by content addressing (draft 3); unchanged.
- **D (new): fail-closed scope.** Per unit, never per history (§7). The alternative — fail only when
  the bad state sits at a ref tip — was rejected because a tag at a bad commit could then wedge the
  sync, and deleting a release tag is effectively a history rewrite. OpenSpec has no such tag today
  (§0), but the rule must not depend on that.
- **E (new): walk.** The git connector's walk, unsimplified, deduplicated by state key (§3). Path
  limiting was rejected because no simplification mode equals "the commits the git connector binds".

## 11. Proposed `BR-sdd-warranty-i9miv7` revision 0003 (owner approves the wording)

Retire (statements kept): **O-5** (exact binding — stale `Resolve(tree, id)` and single-file wording)
and **O-9** (drift against requirement digests that content-addressed records never supersede). Add:

- **O-10 — History-native exact binding:** every commit the git connector ingests contributes the
  revisions of its SDD state. Each revision's `artifact_sha256` is a manifest digest over the exact
  committed bytes that fully determine its payload. A native change name resolves at a commit to
  exactly one change-intent revision — the change active there, or archived by that commit and by
  none of its parents — and that revision is in the ingested batch. Content the profile cannot map
  is reported per state, and fails only a claim on that state.
- **O-12 — R2 scope:** main specs are not ingested; R2's warranty claims concern a change and the
  commits that implement it, never the state of a main spec.

**O-11 (change drift) is not proposed** (round-3 F3). As drafted, it re-flags every separately-archived
change permanently (92 / 108 on OpenSpec), and "current revision" is not observable from the ledger.
It returns only with a definition that is ledger-evaluable and clearable. It is M4's to propose; it is
not a deferral row.

O-1..O-4 and O-8 unchanged. Milestones: O-1 / O-2 / O-10 (M1), O-3 / O-4 (M1-M2, M5), O-8 (M3),
O-12 (all).

## 12. Amendments that land with M1 (named, not yet made)

- **Intent SPEC §1:** point at `sdd/SPEC.md` (replacing the draft link).
- **Intent SPEC §2:** additive `type Skip struct { Commit, Unit string; Err error }`. An optional
  provider interface `interface{ Skipped() []Skip }`. `ProviderResult` gains `Skipped []Skip`
  (omitted when empty). No change for `records` / `specdir`.
- **Intent SPEC §4:** `artifact_sha256` for `sdd` is the §4.3 manifest digest. Committed-file
  providers resolve at the commit's own tree, **and `sdd` additionally reads its parents' archive
  listings** (§6).
- **git connector (`gitcmd`):** the `rev-list` argument list gets one exported home, shared by the
  `sdd` reader. No behaviour change. M1 must not add `sdd` to the git connector's provider allowlist.

## 13. Test derivations and fixtures

- **Commit-graph fixtures.** `testdata/openspec-history/` and `speckit-history/` stay linear. A new
  `testdata/branches/` adds a commit **graph** (the replay helper gains optional per-step `PARENTS` and
  `REF` files). It covers:
  - a root commit carrying an archived change;
  - an unmerged branch whose commit claims a change;
  - an edit-claim-revert on a branch merged `--no-ff` (TREESAME to the base for the SDD paths);
  - a change archived on a side branch, then merged;
  - the leftover-after-archive state (`7ced2a87`: the active delta left byte-identical to its
    archived copy, removed later as in `df12368f`).
- **Fenced-example fixture** (`testdata/unmappable/` sibling `fenced/`): the `adopt-delta-based-changes`
  shape — a MODIFIED requirement whose scenario quotes a RENAMED block inside a fence (INV-39).
- **Every lifecycle test runs over a whole history**, never a single tree.
- **Fail-closed tables** (INV-1, 3, 4, 6, 29, 35, 36, 38): one row per cause, each placed at a non-tip
  commit followed by a fixing commit. Each asserts the Skip, the missing revision, every other
  revision present, and Resolve's error on that state.
- **Calibration** (Law 9 corollary): every row in §9 names the implementation it kills. M1 runs those
  mutants through `scripts/mutant-apply.sh` before its counts are believed.
- **Probe honesty:** `testdata/derive/` prints every §0 count it is cited for. Draft 3's probe computed
  name reuse as `{…} & set()`; draft 4's probe computes it by listing active and archived names per
  state and prints the count. Each count is re-run at M1 against a pinned OpenSpec commit.
- **Family conformance:** every revision passes `intent.ValidateRevision`, and each fixture history
  passes `Sync` against a test appender (INV-37).

## 14. Non-goals and boundaries owned elsewhere

- Main-spec ingestion, folding, tombstones, carry-over (deferred with the main-spec lifecycle).
- Authoring or writing tenant SDD files; governance → business decisions; YAML; heuristic extraction;
  Spec Kit per-task intents; API / DB providers; more than one profile per repository; ledgering
  skips.
- **M3 entry conditions** (owned there, named here so M3 is not surprised):
  1. the `sdd:` trailer token and its decision record;
  2. anti-spoofing that compares the token to a native name the git connector reads from the commit
     tree itself, including the profile's `kit`, because the change-intent suffix hashes it;
  3. **what the git connector does with a claim that does not resolve.** Today it fails the sync
     closed, which is round-3 F1's wedge class moved to the git connector: a tenant commit claiming
     `sdd:X` at one of the 12 dual-presence commits would wedge every git sync. M3 must make this
     per-commit and reported, or prove the trailer gate refuses such commits first;
  4. the readiness gate on `sdd` chains.
- **M4:** change-drift reporting (if ever re-proposed, per §11) and the status export.
