# `sdd` spec-driven-development provider contract — DRAFT 3 (change-scoped, history-native)

Status: **DRAFT 3 (R2 M0), not yet a live kernel contract.** Supersedes
[`R2-sdd-provider-spec-draft2.md`](R2-sdd-provider-spec-draft2.md) (reviewed NEEDS_WORK in
`docs/verification/verdicts/r2-m0-review-round2-20261116.md`) and draft 1. Becomes
`kernel/internal/connector/intent/sdd/SPEC.md` in the same M1 commit as the code and tests it cites.

Scope (owner decision 2026-11-16, recorded in `VD-sdd-change-scoped-r2-k7q3wz`): **R2 warrants
OpenSpec changes and Spec Kit features — the commits that implement a change, against the
requirements that change declares — read from the repository's history.** The main-spec lifecycle
(main specs as living requirement records, folding archived deltas, tombstones, carry-over after
archive) is **not** in R2. Requirement: `BR-sdd-warranty-i9miv7` (revision proposed in §10).

## 0. Evidence first

Draft 3 was checked against the full history of `Fission-AI/OpenSpec` (1052 commits; 256 touching
`openspec/changes`; 111 changes observed active, 108 archived) **before** it was written, by a
probe that reads the change folders at every one of those commits. What it found drives the design:

| Observation (real history) | Count | Consequence in this draft |
|---|---|---|
| Changes archived without ever being observed active — they arrive in one commit with their code (squash-merge) | 16 / 108 | A trailer resolves to an active change **or one archived by that same commit** (§6) |
| Changes with no requirement entries at some commit (early drafts) | 34 | No change intent at that commit: absence, never an error (§5) |
| Requirements removed from an active change's delta between commits | 47 events, 10 changes | **Content-addressed** requirement records: one revision each (§4) |
| Requirement statements changed within an active change | 10 events, 5 changes | Same: a changed delta is a new record, never an in-place edit |
| MODIFIED / REMOVED targets missing from the main spec at the same commit | 1044 events, 6 changes | Main specs are not targeted at all in R2 (§1) |
| Change names reused | 0 | Name-derived change ids are safe; a reuse fails closed (§4) |
| Requirements with no statement paragraph | 4 | Statement falls back to the declared requirement name (§4.2) |

How each round-2 finding is answered: F1 (fold rejects real repos), F3 (fingerprint undefined),
F4 (A2 unsatisfiable), and F7 (mid-change flows) no longer arise — the fold, fingerprinted ids,
carry-over, and main-spec targets are all gone. F2 (tombstones need absent text) and the §5 conflict
dissolve: records never change, so nothing disappears inside one. F5 (changes missed between syncs)
→ history-native ingestion (§3). F6 (digest does not pin payload) → the manifest covers every byte
the payload depends on, the profile included (§4.3). F8 (anti-spoofing) → §6, owned by M3. F9 / F10
→ §11 and the BR proposal in §10.

## 1. Purpose and boundary

Provider name `sdd`; source `intent.sdd`. It maps committed SDD artifacts onto the **existing**
canonical model — `proofbound.requirement.v1` and `proofbound.change-intent.v1` — through one
declarative profile committed in the tenant repository. A third provider of the intent connector:
no new connector, model, or event kind.

Unchanged family rules: **no fabrication** (no business decision, authority, relation, obligation,
status, or identity is invented; governance maps to nothing; requirements render
`authorization: undeclared`); **exact bytes** (no normalization); **commit tree** (everything is
read through the `Reader` at named commits). Specific to this provider: Proofbound **never writes**
into tenant SDD files; and in R2 the **main specs are not ingested** — a change's own delta files
are the requirements it is warranted against.

## 2. Interface lock

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface {
    // Commits reachable from HEAD that touch any pattern, oldest first, topologically ordered.
    Commits(ctx context.Context, patterns []string) ([]string, error)
    // Every committed file at commit matching any pattern, plus the profile.
    Read(ctx context.Context, commit string, patterns []string) ([]Artifact, error)
    // The first parent of commit, or ok=false for a root commit.
    FirstParent(ctx context.Context, commit string) (parent string, ok bool, err error)
}
type Provider struct { /* unexported */ }

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                             // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) // whole history
// Resolve maps a native change name at commit to that commit's change-intent revision.
func (p *Provider) Resolve(ctx context.Context, commit, name string) (intent.Revision, error)
```

Profile path: `.proofbound/sdd-profile.json`. A commit without it contributes nothing (the tenant had
not opted in yet). An invalid profile at a commit fails the sync.

## 3. History-native ingestion

`Revisions` walks `Commits(profile + SDD patterns)` — every commit reachable from HEAD that touches
the profile or a change folder, merged side branches included — maps the tree at each one (§4-§5),
and returns the union, de-duplicated by `(kind, record id, artifact digest)`. Consequences:
- A change opened and archived between two syncs is still ingested (round-2 F5).
- Every commit whose trailer the git connector will resolve (§6) has its exact revision in the batch.
- Re-running is idempotent: unchanged history yields no new revisions (intent connector INV-5).
- The cost is bounded by commits touching SDD paths, not by repository size.

## 4. Requirement records (content-addressed)

### 4.1 What is a requirement record

- **OpenSpec:** one record per delta file `<change-dir>/specs/<capability>/spec.md` that has at least
  one requirement entry, wherever the change directory sits (active under `changes/`, or under
  `changes/archive/YYYY-MM-DD-<name>/`).
- **Spec Kit:** one record per feature spec `specs/<feature>/spec.md` with at least one obligation.
- A file with zero entries yields no record (the canonical model requires ≥ 1 obligation).

### 4.2 Obligations

- **OpenSpec:** each requirement entry under `ADDED`, `MODIFIED`, `REMOVED`, or `RENAMED` is one
  obligation — the change's declared promise. Id: `O-<SEC>-<name-part>`, `SEC` ∈ {`ADD`, `MOD`,
  `REM`, `REN`}; `name-part` is the requirement name upper-cased ASCII, every byte outside `[A-Z0-9]`
  → `-`, runs collapsed, trimmed; if longer than 28 bytes, the first 21 (re-trimmed) + `-` + 6
  upper-case hex of `SHA-256(name)`. (For `RENAMED`, the name is the `FROM` name.) Total ≤ 32 bytes
  after `O-`, as the canonical regex requires. **Statement:** the first non-empty line after the
  requirement heading, before any further heading (for ADDED / MODIFIED the SHALL sentence; for
  REMOVED the `Reason` line; for RENAMED the `TO` line), exact bytes; if there is none, the declared
  requirement name. Never positional.
- **Spec Kit:** `- **FR-001**` / `SC-` / `NFR-` lines → `O-FR-001`, statement = the line's statement
  group.
- Two entries mapping to one obligation id in a record → `intent.ErrUnmappable`.
- Status: an OpenSpec record under an active change → `proposed`; under the archive → `active`
  (the tenant merged it). Spec Kit → the profile's explicit status map, else its declared absent
  value, else `intent.ErrUnmappable`.
- Owner from the profile's `owners` list, first match (`VD-sdd-accountable-identity-ypcx7x`).

### 4.3 Identity and digest — content-addressed

- **Manifest:** one line `"<sha256-of-exact-bytes>  <path>\n"` for the record's source file **and
  the profile file**, sorted by path. `artifact_sha256 = SHA-256(manifest)`; `artifact_path` is the
  source file's path.
- **The payload is a pure function of the manifest's bytes** (round-2 F6). Owner, status, and
  obligations derive only from the source file, its path, and the profile.
- **Record id:** `BR-<slug(key)>-<suffix(key + "\x00" + artifact_sha256)>`, where `key` is
  `<change-name>/<capability>` (OpenSpec) or the feature directory name (Spec Kit), and `slug` /
  `suffix` are as in draft 2 §4.1. A record therefore has **exactly one revision**: any byte change in
  the delta file, its path (archiving moves it), or the profile mints a new record. Intent SPEC §5
  holds by construction — no obligation can disappear or change meaning inside a record that never
  changes.

## 5. Change intents

- **OpenSpec:** one change intent per change **name** (the directory name, minus the archive's
  `YYYY-MM-DD-` prefix). Id `CI-<slug(name)>-<suffix(name)>`: stable across the change's life, so
  trailers and verdicts name one intent. At each commit it `implements` every obligation of every
  requirement record of that change at that commit.
- **Spec Kit (`feature-sibling`):** one change intent per feature directory, emitted only when the
  declared sibling (`tasks.md`) exists; it implements every obligation of the feature's record.
- **Revisions:** the intent's payload changes as its records change, so successive commits yield
  successive revisions of one id (as `specdir`'s do). `supersedes` stays `[]` — the canonical model
  cannot express a change-intent self-supersession (`DEF-INTENT-CI-REVISION-001`) and intent SPEC §5
  constrains requirement obligations, not change-intent targets.
- **Artifact:** manifest over every file in the change directory at that commit **plus the profile**;
  `artifact_path` = the change directory. Pins sponsor, status, and targets (round-2 F6).
- **Status:** active change → the profile's declared `changes.status`; archived → `accepted`.
  Sponsor from the profile's `sponsors` list.
- **No requirement records at a commit → no change intent at that commit.** Absence, not an error.
- **Name reuse** — a name present both active and archived, or archived twice — is
  `intent.ErrUnmappable` for that commit.

## 6. Linkage and lifecycle

```
 OpenSpec (tenant repo)                     Proofbound
 1 propose changes/X/      ── sync ───────▶ BR records for X's deltas (proposed) + CI-X
 2                                          requirement-review.v1 on those records ⇒ VERIFIABLE   [R1]
 3 implement; commit with  ───────────────▶ git connector binds commit → CI-X@exact revision
     Intent: sdd:X                          (resolved at that commit)                            [M3]
 4 tests / CI run          ───────────────▶ check + workflow witness events                       [R1]
 5                                          obligation-verdict.v2 on that commit ⇒ SATISFIED      [R1]
 6                                          intent-delivery-readiness ⇒ PASS = warranted           [R1]
 7 edit X's deltas         ── sync ───────▶ new BR records (new ids); CI-X's next revision targets them
                                            ⇒ a verdict bound to the earlier revision no longer
                                              covers the current one: change drift (M4)
 8 openspec archive X      ── sync ───────▶ archived-location records (active); CI-X revision accepted
 ◀── feedback (read-only): reports, the readiness gate result, a status export (M4)
```

- **Resolve(commit, name)** returns the change intent of `name` at `commit` if the change is
  **active** there, or was **archived by that commit** (its archive directory is absent from the
  commit's first parent). Anything else is an error. The second case is the squash-merge workflow
  (16 / 108 real changes), where the implementing commit is also the archiving commit.
- **Verdicts bind to the change, not the main spec.** An `obligation-verdict.v2` names the commit,
  the exact change-intent revision, and that revision's targets — exactly what projections accept
  today. No carry-over step exists, because R2 makes no claim about main specs (round-2 F4).
- **Trailer token (M3, its own decision record).** `Intent: sdd:<name>`. Needed, and named here so
  M3 is not surprised: the git connector's anti-spoofing check must compare the trailer token to the
  resolved revision's **native name** rather than its record id; `IntentRef` gains that name as an
  additive, omitted-when-empty field so git-connector INV-30's pinned payloads stay byte-identical;
  and the shell trailer gate, the kernel `intentIDRE` check, and the shared trailer corpus all
  extend together. Round-2 F8's point stands — the provider must not vouch for itself — so the
  native name the connector compares comes from the commit's own tree read by the connector, not
  from the provider's answer.

## 7. Resolution invariants of history

A tenant commit's claim can only be bound if the batch contains that commit's revision; §3 ensures
it. A claim on a change with no requirement records at that commit (34 real early-draft states)
fails resolution, and the git connector fails the sync closed, as it does today for any unresolvable
claim. The tenant's own trailer gate is where such a commit should be refused first.

## 8. Invariants

IDs are permanent from draft 1 on (pinned in `docs/invariants.lock` once the file lands): new
invariants append; none is reused. Retired invariants keep their number with `RETIRED` and the
draft that retired them.

1. **SDD-INV-1 — The profile is strict JSON and fails closed.**
2. **SDD-INV-2 — Only requirements and change intents are emitted; governance maps to nothing.**
3. **SDD-INV-3 — Accountable identity comes only from the profile at the same tree.**
4. **SDD-INV-4 — Status comes only from the explicit map or a declared absent value.**
5. **SDD-INV-5 — Obligation ids are never positional.**
6. **SDD-INV-6 — Duplicate or missing obligations fail closed.**
7. **SDD-INV-7 — Record ids are deterministic.**
8. **SDD-INV-8 — The artifact digest is the exact committed bytes.**
9. **SDD-INV-9 — Delta sections map onto the closed relation vocabulary.**
10. **SDD-INV-10 — Feature-sibling change intents are feature-level only.**
11. **SDD-INV-11 — Resolution reads only the named tree and returns exactly one intent.**
12. **SDD-INV-12 — Canonical payload bytes are stable across runs.**
13. **SDD-INV-13 — Two kits run on one provider with only the profile differing.**
14. **SDD-INV-14 — The unmappable fixture stops without changing the model.**
15. **SDD-INV-15 — RETIRED (draft 3): archived changes are now ingested (§3, §5).**
16. **SDD-INV-16 — RETIRED (draft 3): requirement records are content-addressed (§4.3); change ids remain name-derived (INV-25).**
17. **SDD-INV-17 — RETIRED (draft 3): archived delta records are `active` (§4.2).**
18. **SDD-INV-18 — A change's artifact digest covers the exact bytes of every file it reads.**
19. **SDD-INV-19 — RETIRED (draft 3): replaced by INV-27.**
20. **SDD-INV-20 — Archiving a change never makes a later sync fail.**
21. **SDD-INV-21 — RETIRED (draft 3): no fingerprinted obligation ids.**
22. **SDD-INV-22 — RETIRED (draft 3): no tombstones; records never change.**
23. **SDD-INV-23 — RETIRED (draft 3): no delta fold.**
24. **SDD-INV-24 — Every commit reachable from HEAD that touches SDD paths contributes its revisions.**
25. **SDD-INV-25 — A requirement record has exactly one revision; its id is content-addressed.**
26. **SDD-INV-26 — The payload is a pure function of the manifest bytes.**
27. **SDD-INV-27 — A name resolves only to a change active at the commit or archived by it.**
28. **SDD-INV-28 — A commit with no requirement records for a change yields no change intent, not an error.**
29. **SDD-INV-29 — A reused change name is unmappable.**
30. **SDD-INV-30 — Main specs are not ingested in R2.**

| Invariant | Statement | Proving test |
|---|---|---|
| SDD-INV-1 | Unknown or missing fields, unknown modes, and regexes lacking required groups are rejected | sdd_test.go::TestProfileIsStrict |
| SDD-INV-2 | No business decision is emitted; requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted |
| SDD-INV-3 | Owner and sponsor come only from their list's first match at the same commit; uncovered is ErrUnmappable | sdd_test.go::TestIdentityOnlyFromProfile |
| SDD-INV-4 | A Spec Kit status not in the explicit map and with no declared absent value is ErrUnmappable | sdd_test.go::TestStatusMapFailsClosed |
| SDD-INV-5 | Inserting, deleting, or reordering one requirement entry changes no other obligation id | sdd_test.go::TestObligationIDsAreNotPositional |
| SDD-INV-6 | Colliding obligation ids are ErrUnmappable; a zero-entry file yields no record | sdd_test.go::TestObligationCollisionAndEmptyFailClosed |
| SDD-INV-7 | The §4.3 / §5 id vectors in testdata match exactly, across runs and commits | sdd_test.go::TestRecordIDsAreDeterministic |
| SDD-INV-8 | A one-byte edit to any manifest file changes the artifact digest | sdd_test.go::TestDigestIsExactBytes |
| SDD-INV-9 | Every ADDED, MODIFIED, REMOVED, RENAMED entry is an `implements` target of the change intent | sdd_test.go::TestDeltaSectionsMapToClosedRelations |
| SDD-INV-10 | One feature-level change intent, only when the sibling exists | sdd_test.go::TestFeatureSiblingIsFeatureLevel |
| SDD-INV-11 | A reader that fails on any other commit proves Resolve reads only the named commit and its first parent | sdd_test.go::TestResolveUsesCommitTree |
| SDD-INV-12 | Repeated runs over one history emit byte-identical payloads | sdd_test.go::TestCanonicalizationStability |
| SDD-INV-13 | Both reference fixtures map through one constructor, the kit read only from the profile | sdd_test.go::TestTwoKitsOneProvider |
| SDD-INV-14 | The unmappable fixture returns intent.ErrUnmappable and no revisions | sdd_test.go::TestUnmappableFixtureStops |
| SDD-INV-15 | RETIRED (draft 3) | — |
| SDD-INV-16 | RETIRED (draft 3) | — |
| SDD-INV-17 | RETIRED (draft 3) | — |
| SDD-INV-18 | The change manifest changes when any file of the change or the profile changes | sdd_test.go::TestChangeManifestDigest |
| SDD-INV-19 | RETIRED (draft 3) | — |
| SDD-INV-20 | A history containing an archive commit, a `--skip-specs` archive, and a deleted main spec syncs | sdd_test.go::TestArchiveNeverBreaksSync |
| SDD-INV-21 | RETIRED (draft 3) | — |
| SDD-INV-22 | RETIRED (draft 3) | — |
| SDD-INV-23 | RETIRED (draft 3) | — |
| SDD-INV-24 | A change opened and archived between two commits of the walk is still emitted | sdd_test.go::TestHistoryNativeIngestion |
| SDD-INV-25 | Editing a delta file yields a new record id and leaves the earlier record unchanged | sdd_test.go::TestRequirementRecordsAreContentAddressed |
| SDD-INV-26 | Mutating the profile's owner without other change yields a different artifact digest | sdd_test.go::TestPayloadIsPinnedByManifest |
| SDD-INV-27 | Active, archived-by-this-commit, archived-earlier, and unknown names resolve, resolve, fail, fail | sdd_test.go::TestNativeNameResolution |
| SDD-INV-28 | A change with only a proposal.md at a commit yields no change intent and no error | sdd_test.go::TestEmptyChangeIsAbsent |
| SDD-INV-29 | A name both active and archived, or archived twice, is ErrUnmappable | sdd_test.go::TestReusedNameIsUnmappable |
| SDD-INV-30 | Files under the main-spec path never produce a revision | sdd_test.go::TestMainSpecsAreNotIngested |

## 9. Decisions this draft settles or withdraws

- **A (warranty across archive): withdrawn.** R2 makes no main-spec claim, so there is nothing to
  carry over; the change's own verdicts remain valid history.
- **B (main-spec obligation lifecycle): withdrawn from R2** and deferred with the main-spec lifecycle
  (`VD-sdd-change-scoped-r2-k7q3wz`). Round 2 showed the fold cannot hold on real repositories.
- **C (Spec Kit lifecycle): answered** by content addressing — an in-place edit of `FR-001` yields a
  new record; the earlier one is untouched, so §5 holds without any history of edits.

## 10. Proposed `BR-sdd-warranty-i9miv7` revision 0003 (owner approves the wording)

Retire (statements kept): **O-5** (exact binding — its `Resolve(tree, id)` and single-file wording
are stale) and **O-9** (drift — defined against requirement digests a later revision supersedes,
which content-addressed records never do). Add:
- **O-10 — History-native exact binding:** every commit reachable from HEAD that touches the SDD paths
  contributes its revisions; each revision's `artifact_sha256` is a manifest digest over the exact
  committed bytes that fully determine its payload; a native change name resolves at a commit to
  exactly one change-intent revision — the change active there or archived by that commit.
- **O-11 — Change drift:** when a change intent's current revision targets requirement records that
  no `obligation-verdict.v2` covers, the change is reported stale until a new verdict covers them.
- **O-12 — R2 scope:** main specs are not ingested; R2's warranty claims concern a change and the
  commits that implement it, never the state of a main spec.

O-1..O-4 and O-8 unchanged. Milestones: O-1 / O-2 / O-10 (M1), O-3 / O-4 (M1-M2, M5), O-8 (M3),
O-11 (M4), O-12 (all).

## 11. Non-goals and boundaries owned elsewhere

- Main-spec ingestion, folding, tombstones, carry-over (deferred with the main-spec lifecycle).
- Authoring or writing tenant SDD files; governance → business decisions; YAML; heuristic extraction;
  Spec Kit per-task intents; API / DB providers; more than one profile per repository.
- **Owned elsewhere:** registering `intent.sdd` + `sync intent sdd` (M1); the trailer token, the
  git-connector anti-spoofing adaptation, the trailer gate, and the shared corpus (M3, own decision
  record); the readiness gate on `sdd` chains (M3); change-drift reporting and the status export (M4).

## 12. Test derivations

- **Real-history fixtures:** a small OpenSpec repository built as a **commit sequence** under
  `testdata/` reproducing every row of §0 — a change opened, edited (requirement removed, statement
  changed), and archived; a squash-merged change archived in its first commit; an early draft with
  only `proposal.md`; a `--skip-specs` archive; a deleted main spec — plus a Spec Kit sequence with
  an in-place `FR-001` edit. The probe that measured §0 is kept beside them so the fixtures can be
  re-derived.
- **Every lifecycle test runs over the whole commit sequence**, not a single tree — the gap that let
  drafts 1 and 2 pass their own tests while real histories broke them.
- **Fail-closed tables** (INV-1, 3, 4, 6, 29): one row per cause, asserting the error and that no
  revisions were returned. **Calibration** (Law 9 corollary): each table states a positive control
  and a neutral control before its count is believed; mutation runs go through
  `scripts/mutant-apply.sh`.
- **Family conformance:** every revision passes `intent.ValidateRevision`, and each fixture history
  passes `Sync` against a test appender.
