# `sdd` spec-driven-development provider contract

Status: **DRAFT 1 (R2 M0), not yet a live kernel contract.** It becomes
`kernel/internal/connector/intent/sdd/SPEC.md` in the same M1 commit as the code and the tests it
cites — a live kernel SPEC may not cite tests that do not exist (census row C2-036), so per this
repo's convention the spec-before-code draft lives here. Invariant ids below are already permanent
and are pinned in `docs/invariants.lock` when the file lands. Under non-author review in R2 M0 step 4. Requirement: `BR-sdd-warranty-i9miv7` (this package
discharges O-1..O-5; O-6 is `delivery-enforce`'s, O-7 is a projection's — see §8). Decisions:
`VD-release-line-r2-sdd-t5j34p`, `VD-sdd-accountable-identity-ypcx7x`. Evidence:
`docs/verification/r2-m0-mapping-spike.md`. Prior art: the sibling `specdir` provider, whose
boundary this package keeps.

## 1. Purpose and boundary

The provider name is `sdd`; emitted source is `intent.sdd`. It maps the committed artifacts of a
spec-driven-development kit (OpenSpec, Spec Kit, others) onto the **existing** canonical intent
model — `proofbound.requirement.v1` and `proofbound.change-intent.v1` — through one declarative
**profile** committed in the tenant repository. It is a third provider of the intent connector, not
a new connector, model, or event kind.

What differs from `specdir`: `specdir` fixes one markdown grammar and requires `Owner:` / `Sponsor:`
lines in it. `sdd` takes the grammar, the status vocabulary, and accountable identity from the
tenant's profile, because SDD kits record none of the three in a form Proofbound can rely on.

Three rules carry over unchanged from the intent family contract and are restated because this
provider is where they are most tempting to break:
- **No fabrication.** No business decision, authority, relation, obligation, status, or identity is
  invented. Governance documents (OpenSpec `config.yaml`, Spec Kit `constitution.md`) map to
  nothing; requirements carry an empty `authorized_by` and render `authorization: undeclared`.
- **Exact bytes.** `artifact_sha256` is the SHA-256 of the exact committed file bytes. No
  normalization: a reformatted spec is a new revision.
- **Commit tree.** The profile and every artifact are read through the `Reader` at one named tree.
  Nothing is read from the working tree.

## 2. Interface lock

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface {
    // Every committed file at tree whose path matches any pattern, plus the profile itself.
    ReadSDDArtifacts(ctx context.Context, tree string, patterns []string) ([]Artifact, error)
}
type Provider struct { /* unexported */ }

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                   // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) // at "HEAD"
func (p *Provider) Resolve(ctx context.Context, tree, recordID string) (intent.Revision, error)
```

The profile path is fixed: `.proofbound/sdd-profile.json` at the repository root. A tree without
it yields zero revisions (the tenant has not opted in); a tree with an invalid one fails the sync.

## 3. Profile (`proofbound.sdd-profile.v1`)

Strict JSON (stdlib only, Law 8): unknown fields, missing required fields, and unknown enum values
are rejected. Shape, shown for OpenSpec:

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
    "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"],
    "status": { "by_path": { "openspec/changes/archive/**": "accepted", "openspec/changes/**": "proposed" } },
    "sections": { "^## ADDED Requirements$": "implements", "^## MODIFIED Requirements$": "modifies", "^## REMOVED Requirements$": "retires" }
  },
  "identity": [
    { "pattern": "openspec/specs/**",   "owner": "team-platform" },
    { "pattern": "openspec/changes/**", "sponsor": "team-platform" }
  ]
}
```

Spec Kit uses the same shape with `"kit": "speckit"`; requirements under `specs/*/spec.md`; an
obligation **line** rule instead of a heading rule:
`"obligation": { "line": "^- \\*\\*(?P<id>(?:FR|SC|NFR)-[0-9]+)\\*\\*:? (?P<statement>.+)$" }`;
a status line rule (`"line": "^\\*\\*Status:?\\*\\*:? *(?P<status>.+?) *$"`) with an explicit `map`;
and `"mode": "feature-sibling"` for changes (§5.2).

Field rules:
- **Patterns** are globs: `*` matches within one path segment, `**` across segments.
- **Regexes** are Go RE2 with the named groups shown; a regex lacking a required group is rejected.
- **Status.** `map` is an explicit table from the exact observed string (after trimming) to a
  canonical status. `absent` applies only when the artifact has no status line and must be declared
  to be used. An observed string missing from `map`, or an absent status with no `absent`, is
  `intent.ErrUnmappable` — never guessed.
- **Identity** (`VD-sdd-accountable-identity-ypcx7x`). Entries are matched **in declared order;
  first match wins**. A requirement takes `owner`, a change intent takes `sponsor`. An artifact no
  entry covers — or an entry covering it without the needed field — is `intent.ErrUnmappable`. No
  default, git author, or placeholder is ever substituted.

## 4. Requirement mapping

One requirement record per matching artifact.
- **Record id** = `BR-<slug>-<suffix>`. `slug` is the artifact's directory name lower-cased to
  `[a-z0-9-]`, truncated to 40 characters; `suffix` is the first 6 characters of the SHA-256 of
  `"<kit>:<path>"`, mapped to `[a-z0-9]` by `byte % 36`. Deterministic across runs and trees.
- **Obligations** come from the profile's obligation rule:
  - `id` group present (Spec Kit) → `O-<id>` verbatim (`FR-001` → `O-FR-001`).
  - `name` group only (OpenSpec) → `O-` + the name upper-cased to `[A-Z0-9-]` (runs of other
    characters collapse to one `-`, trimmed). If longer than 32 characters after `O-`: the first 25,
    then `-`, then 6 upper-case hex characters of the SHA-256 of the full name.
  - **Never positional.** Inserting, deleting, or reordering one requirement does not change any
    other obligation's id.
  - Statement = the `statement` group, else the `name` group; must be non-empty.
  - Two obligations with the same id in one record → `intent.ErrUnmappable`. Zero obligations →
    `intent.ErrUnmappable` (the canonical model requires at least one).
  - Obligations are emitted sorted by id (the canonical model requires strictly increasing ids).
- `authorized_by` is `[]`; `supersedes` is `[]`.

## 5. Change-intent mapping

### 5.1 `delta-sections` (OpenSpec)

Each matching delta file `<change>/specs/<capability>/spec.md` is one change's view of one
capability. Within it, each section heading listed in `sections` opens a relation, and each
obligation heading under it names a requirement.

- **ADDED** requirements do not yet exist in the capability's main spec. The delta file is itself a
  real requirement document, so it maps to a **delta requirement record** (§4 rules; id from the
  delta path; status from `changes.status`; owner from `identity`) holding the ADDED and MODIFIED
  requirements as obligations. The change intent `implements` the delta record's ADDED obligations.
- **MODIFIED** → the change intent `modifies` the obligation of the same id in the capability's
  **main** requirement record at the same tree.
- **REMOVED** → the change intent `retires` the obligation of the same id in the main record.
- A `modifies` / `retires` target whose obligation is absent from the main record at that tree is
  `intent.ErrUnmappable` — never a dangling or guessed reference.
- One change intent per change directory; id `CI-<change-dir-slug>-<suffix>` (§4 derivation over the
  change directory path). Targets bind the exact artifact digest of each referenced record.

### 5.2 `feature-sibling` (Spec Kit)

One change intent per feature directory, emitted **only** when the profile's sibling file (e.g.
`tasks.md`) exists beside the spec. It `implements` **all** obligations of that feature's
requirement record. Per-task targeting is not done: Spec Kit tasks cite user stories, not
requirement ids, so a task→obligation link would be invented.

## 6. Resolution

`Resolve(tree, recordID)` reads the profile and artifacts at `tree` through the `Reader`, maps
them, and returns the single change intent with that id. Zero or more than one match, an invalid
profile, or a `withdrawn` / `superseded` status is an error. This is the seam the git connector
uses to bind a tenant commit's `Intent:` trailer (§8).

## 7. Invariants

IDs are permanent from this draft on, and are pinned by `docs/invariants.lock` once the file lands
in the kernel: new invariants append; none is inserted, reordered, reused, or silently retitled.

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

| Invariant | Statement | Proving test |
|---|---|---|
| SDD-INV-1 | The profile is strict JSON: unknown or missing fields, unknown modes, and regexes lacking required groups fail closed | sdd_test.go::TestProfileIsStrict |
| SDD-INV-2 | Only requirement and change-intent revisions are emitted; governance documents produce nothing and requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted |
| SDD-INV-3 | Owner and sponsor come only from the first matching profile identity entry at the same tree; an uncovered artifact is ErrUnmappable | sdd_test.go::TestIdentityOnlyFromProfile |
| SDD-INV-4 | Status comes only from the explicit map or a declared absent value; anything else is ErrUnmappable | sdd_test.go::TestStatusMapFailsClosed |
| SDD-INV-5 | Obligation ids derive from source ids or names and are never positional | sdd_test.go::TestObligationIDsAreNotPositional |
| SDD-INV-6 | Duplicate obligation ids or zero obligations in one record are ErrUnmappable | sdd_test.go::TestObligationCollisionAndEmptyFailClosed |
| SDD-INV-7 | Record ids are a deterministic function of kit and path | sdd_test.go::TestRecordIDsAreDeterministic |
| SDD-INV-8 | artifact_sha256 is the exact committed bytes; a whitespace-only edit is a new revision | sdd_test.go::TestDigestIsExactBytes |
| SDD-INV-9 | Delta sections map ADDED, MODIFIED, REMOVED to implements, modifies, retires; a missing target is ErrUnmappable | sdd_test.go::TestDeltaSectionsMapToClosedRelations |
| SDD-INV-10 | Feature-sibling mode emits one feature-level change intent only when the sibling exists | sdd_test.go::TestFeatureSiblingIsFeatureLevel |
| SDD-INV-11 | Resolution reads only the named tree and returns exactly one non-withdrawn intent | sdd_test.go::TestResolveUsesCommitTree |
| SDD-INV-12 | Canonical payload bytes are identical across repeated runs | sdd_test.go::TestCanonicalizationStability |
| SDD-INV-13 | The OpenSpec and Spec Kit reference fixtures run on one provider with only the profile differing | sdd_test.go::TestTwoKitsOneProvider |
| SDD-INV-14 | The unmappable fixture returns intent.ErrUnmappable without changing the canonical model | sdd_test.go::TestUnmappableFixtureStops |

## 8. Non-goals and boundaries owned elsewhere

- **Authoring** specs, decisions, or impact graphs — the SDD tool's job.
- **Business decisions from governance** — no kit records an approver (mapping spike).
- **YAML.** No kit file the mapping needs is YAML; the profile is JSON. A YAML dependency needs a VD.
- **Heuristic or model-based extraction.** Only the profile's explicit rules apply.
- **Per-task change intents** for Spec Kit (§5.2).
- **API / DB providers** — the escape hatch; not built in R2.
- **More than one profile per repository** in R2.
- **Owned elsewhere (named so the gap is not read as a bug):**
  - Registering `intent.sdd` as a `core.Source` and adding `sync intent sdd` to the CLI (M1).
  - The git connector and the Intent-trailer gate accepting an `sdd:` qualifier, extended through
    the shared trailer corpus so the two parsers stay in step (M3).
  - **How a tenant learns a derived change-intent id** to put in its trailer — a CLI listing (M3).
  - The traceability gate, `BR-sdd-warranty-i9miv7` O-6 — `delivery-enforce` (M3).
  - Drift reporting, O-7 — a projection over revisions this provider already emits (M4).

## 9. Test derivations

Every "never / only / always" above is a test:
- **Fixtures** live under this package's `testdata/`: an OpenSpec mini-repo (2 capabilities; one
  active and one archived change with ADDED, MODIFIED, and REMOVED sections) and a Spec Kit mini-repo
  (2 features, one with `tasks.md`), each with its profile; plus hostile variants.
- **Fail-closed tables** (INV-1, 3, 4, 6, 9): one row per rejection cause, each asserting
  `errors.Is(err, intent.ErrUnmappable)` or the profile error, and asserting **no** revisions were
  returned.
- **Not positional** (INV-5): map a spec, insert a requirement at the top, map again, and assert
  every pre-existing obligation id is unchanged.
- **Exact bytes** (INV-8): add one trailing space and assert a new digest and a new revision.
- **Tree isolation** (INV-11): two trees with different profiles; `Resolve` at each returns that
  tree's identity. A reader that fails on any non-named tree proves no other tree is touched.
- **One provider, two kits** (INV-13): the same constructor maps both fixtures; the test fails if the
  kit name is consulted anywhere except the profile's own field.
- **Calibration** (Law 9 corollary): each table states a positive control (a fixture that must be
  rejected) and a neutral control (a valid fixture that must map) before its count is believed.
- **Family conformance**: every emitted revision passes `intent.ValidateRevision`, and a whole
  fixture batch passes `Sync` against a test appender.
