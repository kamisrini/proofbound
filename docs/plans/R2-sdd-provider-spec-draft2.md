# `sdd` spec-driven-development provider contract — DRAFT 2

Status: **DRAFT 2 (R2 M0), not yet a live kernel contract.** Supersedes
[`R2-sdd-provider-spec-draft1.md`](R2-sdd-provider-spec-draft1.md), which is kept unchanged as the
artifact reviewed in `docs/verification/verdicts/r2-m0-review-round1-20261116.md` (NEEDS_WORK). It
becomes `kernel/internal/connector/intent/sdd/SPEC.md` in the same M1 commit as the code and tests it
cites (census row C2-036). **Two decisions are open (§9) and must be settled before M1.**

Requirement: `BR-sdd-warranty-i9miv7` (O-6 to be restated — §10). Decisions:
`VD-release-line-r2-sdd-t5j34p`, `VD-sdd-accountable-identity-ypcx7x`. Evidence:
`docs/verification/r2-m0-mapping-spike.md`. Prior art: the sibling `specdir` provider.

## 0. What changed from draft 1, and why

The round-1 review showed draft 1 was designed against a **snapshot** of OpenSpec, never its
**lifecycle**: `openspec archive` moves `changes/<name>/` to `changes/archive/YYYY-MM-DD-<name>/`
and merges the deltas into the main specs. Every HIGH finding sat on that transition.

| Finding | Draft 2 resolution |
|---|---|
| F1 delta record status `accepted` is not a requirement status | Delta records are always `proposed` (§5.2); change status is separate (§3) |
| F2 archived REMOVED change → permanently unsyncable; archived intents re-bind every sync | **Only active changes map to change intents** (§5.1); targets resolve at the same tree, where an active change's targets exist |
| F3 path-derived ids change on archive | Ids derive from the **native name**, never the path (§4.1) |
| F4 one requirement in two records | Delta = proposed home, main = accepted home; the handover is decision A (§9) |
| F5 one CI over N files has no defined digest | A change's artifact is a **manifest** of its files' exact digests (§5.1) |
| F6 reference profile rejects valid changes; REMOVED-only → zero obligations; RENAMED unhandled | Separate `owners` / `sponsors` lists (§3); a delta record is emitted only when it has obligations; RENAMED mapped (§5.3) |
| F7 unordered `by_path` status precedence | Removed; change status is one declared value (§3) |
| F8 obligations vanish / change meaning in place, against intent SPEC §5 | **Open decision B** (§9) |
| F9 id derivation underspecified | Fully specified, with collision checks (§4.1) |
| F10 cross-document contradictions | Fixed in the plan, VD, and this draft (separate commit) |
| F11 weak obligations | O-6 restated to the existing readiness gate (§10); BR revision comes back to the owner |
| *(owner question)* how artifacts tie both ways | New §6, linkage and lifecycle |

## 1. Purpose and boundary

Provider name `sdd`; emitted source `intent.sdd`. It maps the committed artifacts of a
spec-driven-development kit onto the **existing** canonical model — `proofbound.requirement.v1` and
`proofbound.change-intent.v1` — through one declarative profile committed in the tenant repository.
It is a third provider of the intent connector: no new connector, model, or event kind.

Carried over unchanged from the family contract: **no fabrication** (no business decision,
authority, relation, obligation, status, or identity is invented; governance maps to nothing and
requirements render `authorization: undeclared`); **exact bytes** (no normalization; a reformatted
spec is a new revision); **commit tree** (everything is read through the `Reader` at one named tree).
And one boundary specific to this provider: **Proofbound never writes into the tenant's SDD files.**

## 2. Interface lock

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface {
    // Every committed file at tree matching any pattern, plus the profile itself.
    ReadSDDArtifacts(ctx context.Context, tree string, patterns []string) ([]Artifact, error)
}
type Provider struct { /* unexported */ }

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                              // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error)  // at "HEAD"
// Resolve maps a native change name (the trailer token) at tree to its change intent.
func (p *Provider) Resolve(ctx context.Context, tree, token string) (intent.Revision, error)
```

Profile path: `.proofbound/sdd-profile.json` at the repository root. No profile → zero revisions (the
tenant has not opted in). Invalid profile → the sync fails.

## 3. Profile (`proofbound.sdd-profile.v1`)

Strict JSON, stdlib only (Law 8): unknown or missing fields and unknown enum values are rejected.
OpenSpec:

```json
{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "openspec",
  "requirements": {
    "paths": ["openspec/specs/*/spec.md"],
    "status": { "line": null, "absent": "active", "map": {} },
    "obligation": { "heading": "^### Requirement: (?P<name>.+)$" }
  },
  "changes": {
    "mode": "delta-sections",
    "root": "openspec/changes",
    "archive": "openspec/changes/archive",
    "status": "accepted",
    "sections": {
      "^## ADDED Requirements$": "added", "^## MODIFIED Requirements$": "modified",
      "^## REMOVED Requirements$": "removed", "^## RENAMED Requirements$": "renamed"
    },
    "renamed": { "from": "^- FROM: `### Requirement: (?P<name>.+)`$", "to": "^- TO: `### Requirement: (?P<name>.+)`$" }
  },
  "owners":   [ { "pattern": "openspec/**", "owner": "team-platform" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "team-platform" } ]
}
```

Spec Kit uses `"kit": "speckit"`, requirements under `specs/*/spec.md`, an obligation **line** rule
(`"line": "^- \\*\\*(?P<id>(?:FR|SC|NFR)-[0-9]+)\\*\\*:? (?P<statement>.+)$"`), a status line rule
with an explicit `map`, and `"mode": "feature-sibling"` with a `"sibling": "tasks.md"` field.

Field rules:
- **Patterns** are globs: `*` within one segment, `**` across segments. **Regexes** are Go RE2 and
  must contain the named groups shown, or the profile is rejected.
- **Requirement status:** `map` is an explicit table from the exact observed string (trimmed) to a
  canonical status; `absent` applies only to an artifact with no status line and must be declared to
  be used. Anything else is `intent.ErrUnmappable` — never guessed.
- **Change status:** one declared value (`proposed` or `accepted`) for every *active* change. The
  kit records no approval, so the tenant declares what an active change means in its workflow.
- **Identity** (`VD-sdd-accountable-identity-ypcx7x`): `owners` covers requirement records (main and
  delta); `sponsors` covers change intents. Each list is matched **in declared order, first match
  wins**. An artifact its list does not cover is `intent.ErrUnmappable`; no default, git author, or
  placeholder is ever substituted. (Separate lists remove draft 1's first-match trap, F6.)

## 4. Requirement mapping (main specs)

One requirement record per matching main-spec artifact; `authorized_by` and `supersedes` are `[]`.

### 4.1 Identity derivation (exact)

- **`slug(s)`**: lower-case ASCII; every byte outside `[a-z0-9]` becomes `-`; runs of `-` collapse;
  leading and trailing `-` are trimmed; truncate to 40 bytes and re-trim. A result shorter than 2
  bytes is `intent.ErrUnmappable`.
- **`suffix(key)`**: the first 6 bytes of `SHA-256(kit + "\x00" + kind + "\x00" + key)`, each mapped
  to `"abcdefghijklmnopqrstuvwxyz0123456789"[b % 36]`.
- **Native keys** (names, never paths, so archive cannot change them): a main requirement →
  the capability directory name; a delta requirement → `<change-name>/<capability>`; a change →
  the change name. An archived directory's `YYYY-MM-DD-` prefix is not part of the name.
- **Record ids**: `BR-<slug(key)>-<suffix(key)>` and `CI-<slug(key)>-<suffix(key)>`.
- **Obligation ids**: an `id` group → `O-<id>` verbatim (`FR-001` → `O-FR-001`). A `name` group →
  `O-` + the name upper-cased ASCII with every byte outside `[A-Z0-9]` becoming `-`, runs collapsed,
  trimmed. Empty → `intent.ErrUnmappable`. Longer than 32 bytes after `O-` → the first 25 bytes
  (re-trimmed) + `-` + 6 upper-case hex characters of `SHA-256(name)`. **Never positional.** (The
  spike's "31" was wrong; the canonical regex allows 32.)
- **Collisions fail closed**: two obligation names mapping to one id in a record, or two artifacts
  mapping to one record id in a batch, is `intent.ErrUnmappable` — checked by the provider before it
  returns, because the connector's duplicate check includes the digest and would not catch it.
- Obligations are emitted sorted by id (the canonical model requires strictly increasing ids); a
  record with zero obligations is `intent.ErrUnmappable`.

## 5. Change mapping

### 5.1 `delta-sections` (OpenSpec) — active changes only

A change is **active** while its directory is directly under `changes.root` and not under
`changes.archive`. **Only active changes produce change intents and delta records.** An archived
change is finished: its effect is in the main specs, and the ledger already holds every revision it
had while active (the record is event-sourced; leaving HEAD is not deletion).

- **One change intent per active change**, id from the change name (§4.1), status from
  `changes.status`, sponsor from `sponsors`.
- **Artifact:** `artifact_path` is the change directory. `artifact_sha256` is the SHA-256 of a
  **manifest**: one line `"<sha256-of-exact-file-bytes>  <path>\n"` per file the mapping reads in that
  directory, sorted by path. Every file still contributes its exact bytes; any byte change in any of
  them changes the digest (F5).
- **Targets** resolve at the same tree. Because the change is active at that tree, a MODIFIED,
  REMOVED, or RENAMED target exists in the main spec there unless someone edited the main spec
  outside the workflow — which is `intent.ErrUnmappable`, correctly (F2).

### 5.2 Delta requirement records

Each active change's delta file `<change>/specs/<capability>/spec.md` that contains ADDED or
MODIFIED requirements maps to one **delta requirement record**: id from `<change-name>/<capability>`;
obligations = its ADDED and MODIFIED requirements (§4.1 rules); status **always `proposed`** (it is a
proposed requirement by construction — F1); owner from `owners`. A delta file with only REMOVED or
RENAMED sections produces no delta record (F6).

### 5.3 Section → relation

| Delta section | Change-intent relation |
|---|---|
| ADDED | `implements` the delta record's obligation |
| MODIFIED | `implements` the delta record's obligation (new statement) **and** `modifies` the main record's obligation of the same name |
| REMOVED | `retires` the main record's obligation of that name |
| RENAMED (FROM → TO) | `modifies` the main record's FROM obligation (the TO name enters the main record on archive) |

### 5.4 `feature-sibling` (Spec Kit)

One change intent per feature directory, emitted only when the declared sibling (e.g. `tasks.md`)
exists; it `implements` **all** obligations of that feature's requirement record. Per-task targeting
is not done: Spec Kit tasks cite user stories, not requirement ids. Spec Kit has no archive step, so
§5.1's lifecycle does not arise.

## 6. Linkage and lifecycle

How artifacts tie from the SDD kit into Proofbound and back. Most of the middle is R1 machinery; this
section names what is reused, what is new, and what is open.

```
 OpenSpec (tenant repo)                         Proofbound
 1 propose changes/X/        ── sync ─────────▶  CI-X (status per profile) + delta BR-X (proposed)
 2                                                requirement-review.v1 by a non-owner ⇒ VERIFIABLE   [R1]
 3 implement; commits carry  ─────────────────▶  git connector binds commit → CI-X@exact digest
     Intent: sdd:X                                (native-name token, resolved at the commit's tree) [NEW]
 4 tests / CI run            ─────────────────▶  check + workflow witness events (evidence)        [R1]
 5                                                obligation-verdict.v2 by a verifier ⇒ SATISFIED     [R1]
 6                                                intent-delivery-readiness gate ⇒ PASS = warranted   [R1]
 7 openspec archive X        ── sync ─────────▶  main BR new revision; CI-X leaves HEAD (ledger keeps it)
                                                  — carrying the warranty across: OPEN DECISION A
 8 later spec edit           ── sync ─────────▶  new digest ⇒ verdicts bound to the old one are stale
 ◀── feedback (read-only): reports, the readiness gate as a PR check, a status export agents can read
```

- **Commit → intent (new).** Tenants write `Intent: sdd:<change-name>` — the name an agent already
  works in — never a derived hash. `Resolve(tree, token)` returns the active change of that name at
  the commit's own tree; zero or more than one match, or an archived-only name, is an error. This
  needs the git connector and the Intent-trailer gate to accept the `sdd` qualifier with a native
  token, and the connector's anti-spoofing rule (resolved record id must equal the trailer token,
  git-connector INV-29) adapted for `sdd`: the token must be the resolved change's native name, and
  both the token and the resolved id are recorded. Both parsers extend through the shared trailer
  corpus. Owned by M3; needs its own decision record.
- **Evidence → obligation (reused).** Satisfaction is an `obligation-verdict.v2` naming the commit,
  the exact change-intent revision, every targeted obligation, and evidence event ids. OpenSpec's
  WHEN/THEN scenarios are where a verifier looks for evidence, but **the verdict is the proof**: the
  provider never infers a test → obligation link from test or scenario names.
- **Requirement review (reused).** `requirement-review.v1`, reviewer ≠ the profile-declared owner.
- **Gate (reused).** `intent-delivery-readiness` applies to `intent.sdd` chains unchanged.
- **Feedback (read-only).** Reports and the gate result go back to the tenant. A machine-readable
  status export for agents is M3/M4 scope. Proofbound never writes into `openspec/` or `specs/`.

## 7. Resolution

`Resolve(tree, token)` reads the profile and artifacts at `tree` through the `Reader`, maps them, and
returns the single **active** change intent whose native name is `token`. An invalid profile, no
match, or more than one match is an error. A withdrawn or superseded status is an error.

## 8. Invariants

IDs are permanent from draft 1 on (pinned by `docs/invariants.lock` when the file lands): new
invariants append; none is inserted, reordered, reused, or silently retitled. INV-1..14 keep their
draft-1 ids and titles; INV-15..20 are appended by draft 2.

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
15. **SDD-INV-15 — Archived changes produce no change intent and no delta record.**
16. **SDD-INV-16 — Record ids survive archive: they derive from native names, never paths.**
17. **SDD-INV-17 — A delta requirement record is always proposed.**
18. **SDD-INV-18 — A change's artifact digest covers the exact bytes of every file it reads.**
19. **SDD-INV-19 — A native-name token resolves to exactly one active change at the commit's tree.**
20. **SDD-INV-20 — Archiving a change never makes a later sync fail.**

| Invariant | Statement | Proving test |
|---|---|---|
| SDD-INV-1 | Unknown or missing fields, unknown modes, and regexes lacking required groups are rejected | sdd_test.go::TestProfileIsStrict |
| SDD-INV-2 | No business decision is emitted; requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted |
| SDD-INV-3 | Owner and sponsor come only from their own list's first match at the same tree; uncovered is ErrUnmappable | sdd_test.go::TestIdentityOnlyFromProfile |
| SDD-INV-4 | An unmapped status string, or an absent status with no declared absent value, is ErrUnmappable | sdd_test.go::TestStatusMapFailsClosed |
| SDD-INV-5 | Inserting, deleting, or reordering one requirement changes no other obligation id | sdd_test.go::TestObligationIDsAreNotPositional |
| SDD-INV-6 | Colliding obligation or record ids, and zero-obligation records, are ErrUnmappable | sdd_test.go::TestObligationCollisionAndEmptyFailClosed |
| SDD-INV-7 | The same native key yields the same id across runs and trees; the §4.1 vectors match exactly | sdd_test.go::TestRecordIDsAreDeterministic |
| SDD-INV-8 | A one-byte edit to any read file changes the artifact digest and yields a new revision | sdd_test.go::TestDigestIsExactBytes |
| SDD-INV-9 | ADDED, MODIFIED, REMOVED, RENAMED map per §5.3; a missing main target is ErrUnmappable | sdd_test.go::TestDeltaSectionsMapToClosedRelations |
| SDD-INV-10 | One feature-level change intent, only when the sibling exists | sdd_test.go::TestFeatureSiblingIsFeatureLevel |
| SDD-INV-11 | A reader that fails on any other tree proves only the named tree is read | sdd_test.go::TestResolveUsesCommitTree |
| SDD-INV-12 | Repeated runs over one tree emit byte-identical payloads | sdd_test.go::TestCanonicalizationStability |
| SDD-INV-13 | Both reference fixtures map through one constructor; the kit name is read only from the profile | sdd_test.go::TestTwoKitsOneProvider |
| SDD-INV-14 | The unmappable fixture returns intent.ErrUnmappable and no revisions | sdd_test.go::TestUnmappableFixtureStops |
| SDD-INV-15 | A tree with only archived changes emits main records and nothing else | sdd_test.go::TestArchivedChangesEmitNothing |
| SDD-INV-16 | A change's and its delta records' ids are identical before and after a simulated archive move | sdd_test.go::TestIdsSurviveArchive |
| SDD-INV-17 | Every delta record is proposed regardless of the declared change status | sdd_test.go::TestDeltaRecordIsProposed |
| SDD-INV-18 | The manifest digest changes when any one file in the change changes, and only then | sdd_test.go::TestChangeManifestDigest |
| SDD-INV-19 | A token resolves to one active change; unknown, archived-only, or ambiguous tokens are errors | sdd_test.go::TestNativeTokenResolution |
| SDD-INV-20 | Syncing the tree after an archive commit succeeds, including for a REMOVED-only change | sdd_test.go::TestArchiveNeverBreaksSync |

## 9. Open decisions (must be settled before M1)

### A. Carrying a warranty across archive (step 7; F4)

Verdicts bind to an active change's delta record and change intent. After archive the requirement
lives in the main record, which has no verdict, so a satisfied requirement would read UNVERIFIED.
- **A1 — Derived carry-over.** A report joins an archived delta obligation to the main obligation with
  the same capability, obligation id, and statement digest. Computed at report time, never stored;
  any statement edit breaks the join. Convenient; the join is an inference the report makes.
- **A2 — Re-attest at archive (author's recommendation).** A verifier writes an
  `obligation-verdict.v2` for the main record, citing the earlier verdict's events as evidence. No new
  mechanism, no inference; costs one verifier step per archived change. A1 may still be shown as a
  non-authoritative *carry-over candidate* that counts for nothing.

### B. Obligation lifecycle vs intent SPEC §5 (F8)

§5: *active obligation ids cannot disappear or change statement in place; meaning change uses a new
id and retirement leaves a tombstone.* OpenSpec edits requirements in place (MODIFIED) and deletes
them (REMOVED). The family invariant is not weakened (option B3 below is listed only to reject it).
- **B1 — Deltas are the lifecycle (author's recommendation).** OpenSpec's archived deltas are the
  tenant's own declarations of each change, kept in the repository and ordered by their date prefix.
  The main record's obligations are rebuilt by folding them: ADDED adds; REMOVED leaves a `retired`
  tombstone (declared by the tenant, not invented); MODIFIED retires the old statement and adds a new
  obligation id carrying a statement fingerprint. The fold must equal the current main spec, or the
  sync fails closed — a main spec hand-edited outside the workflow is detected, not guessed around.
  Faithful to §5 and to OpenSpec; the most complex option, and strict for tenants who hand-edit.
- **B2 — Statement-fingerprinted ids only.** `O-<NAME>-<6 hex of statement>`: a meaning change is
  automatically a new id. Simple, but a REMOVED requirement still disappears without a tombstone, so
  §5 is still broken for retirement.
- **B3 — Amend §5 for mutable sources.** Rejected: it weakens a family invariant every provider and
  report relies on.

## 10. `BR-sdd-warranty-i9miv7` changes this draft implies (owner approves the wording)

- **O-6** currently says `delivery-enforce` refuses a change whose intent targets no requirement
  obligation — nearly vacuous, because canonical validation already rejects such an intent (F11).
  Proposed restatement: *the existing `intent-delivery-readiness` gate applies to `intent.sdd`
  chains unchanged, and a tenant commit is warranted only when every exact target is SATISFIED with
  VERIFIABLE requirement review.*
- O-7's "until re-verified" gains a definition (a new `obligation-verdict.v2` bound to the new
  revision), and O-2 / O-5 may be split; drafted in the BR revision, not here.

## 11. Non-goals and boundaries owned elsewhere

- Authoring specs, decisions, or impact graphs; writing into tenant SDD files.
- Business decisions from governance; YAML parsing; heuristic or model-based extraction.
- Per-task change intents for Spec Kit; API / DB providers; more than one profile per repository.
- **Owned elsewhere:** registering `intent.sdd` and `sync intent sdd` (M1); the git connector, the
  Intent-trailer gate, and the shared trailer corpus accepting `sdd:<native-name>`, with the INV-29
  adaptation and its own decision record (M3); the readiness gate on `sdd` chains (M3); drift
  reporting and the status export (M4).

## 12. Test derivations

- **Fixtures** under the package's `testdata/`: an OpenSpec mini-repo as a **sequence of trees** —
  T0 (two capabilities), T1 (an active change with ADDED, MODIFIED, REMOVED, RENAMED), T2 (that change
  archived and merged) — plus a Spec Kit mini-repo and hostile variants.
- **Lifecycle tests run across the tree sequence** (INV-15, 16, 20): draft 1 had none, which is why
  its tests could pass while archive broke everything.
- **§4.1 vectors** (INV-7): fixed name → id pairs computed once and committed, so the derivation
  cannot drift silently.
- **Fail-closed tables** (INV-1, 3, 4, 6, 9, 19): one row per rejection cause, each asserting the
  error and that **no** revisions were returned.
- **Calibration** (Law 9 corollary): each table states a positive control that must be rejected and
  a neutral control that must map before its count is believed.
- **Family conformance:** every revision passes `intent.ValidateRevision`; each fixture tree passes
  `Sync` against a test appender.
