# R2 Spec Kit support — spec

**Status:** owner-approved scope 2026-11-19 ("go with recommendation"). Merges the old M2 (Spec Kit
mapping) and M5 (second kit) per the re-sequenced R2 plan. Law 11 budget: at most 2 spec review
rounds, 1 code review. Starting point: `R2-sdd-provider-spec-draft10.md` Appendix A, whose open items
(a Resolve rule, a change-intent id formula, feature-key collisions) are closed below.

**Payoff:** a [Spec Kit](https://github.com/github/spec-kit) project gets exactly what an OpenSpec
project gets, through the same commands: `proofbound init speckit`, `Intent: sdd:<feature>`
trailers, readiness gates, `review draft`/`seal`, and `report drift`.

## 1. Spec Kit's shape (what a project commits)

```
specs/001-export/spec.md      # the feature spec: "- **FR-001**: System MUST ..." lines
specs/001-export/plan.md      # the plan
specs/001-export/tasks.md     # the task list: the feature is planned for build
```

A **feature** is a directory directly under the root (`specs/`). Its native name is the directory
name (`001-export`). There is no archive: a feature edits its spec in place.

## 2. Profile (`kit: speckit`)

```json
{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"] },
  "changes": { "mode": "feature-sibling", "root": "specs", "sibling": "tasks.md", "status": "proposed" },
  "owners":   [ { "pattern": "specs/**", "owner": "..." } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "..." } ]
}
```

`root` follows the literal path grammar (§2.1). `sibling` is one literal file-name segment.
`requirements.paths` must be exactly `[<root>/*/spec.md]`: one level, so two features cannot share
a key (Appendix A's multi-`*` collision cannot arise). There is no status regex: a regex in a tenant
profile is attack surface, and requirement status follows OpenSpec's (`proposed` for live specs).
The OpenSpec profile is unchanged, and `kit` decides which shape is valid.

## 3. Mapping one state

- **Feature entries:** trees directly under `root` whose name is in the change-name domain. A
  symlink or gitlink entry is `non-regular-file`. Regular files under `root` are ignored.
- **Requirement record:** one per feature with a regular `spec.md` holding at least one obligation
  line. Its key is the feature name, and its digest is the manifest of `spec.md` plus the profile.
  Obligation lines (outside code fences, using OpenSpec's fence mask) match
  `^- \*\*((?:FR|SC|NFR)-[0-9]+)\*\*:? (.+)$`, with id `O-<FR-001>`. Causes: `statement-blank`,
  `obligation-collision` (a repeated id), plus the shared `invalid-utf8`, `nul-byte`,
  `non-regular-file`, `owner-uncovered`, `payload-too-large`, `slug-too-short`. A `spec.md` with no
  obligation line is not a record unit.
- **Change intent:** one per feature whose `sibling` exists as a regular file and whose record was
  emitted. It targets every obligation of that record, its digest is the manifest of every entry
  under the feature directory plus the profile, and its status is `changes.status`. A feature without
  its sibling has no change intent. A skipped record cascades (`cascade`).
- **Ids:** the same formulas as OpenSpec. `CI-<slug(name)>-<suffix>` and `BR-<slug(key)>-<suffix>`,
  with the suffix hashing the kit as well (`speckit`), so the two kits can never share an id.
  OpenSpec ids are unchanged byte for byte.

## 4. Resolve and StateRevision

`Resolve(commit, name)` / `StateRevision(commit, name)`: the feature named `name` in the commit's
state. Its change intent if it has one, `ErrNoChangeIntent` if the feature exists without it,
`ErrUnknownChange` otherwise. `ErrArchivedEarlier` never arises. Claims, the readiness gate,
`review`, and drift then work unchanged, because they only see revisions.

## 5. Invariants (sdd SPEC)

- **SDD-INV-64 retired** (M1 was OpenSpec-only; a Spec Kit profile is now valid).
- **SDD-INV-69** The Spec Kit fixture history maps exactly: step 01 gives one requirement
  `O-FR-001, O-FR-002, O-SC-001` and no change intent; step 02 adds the feature's change intent
  targeting all three; step 03 (FR-001 edited in place) gives a new requirement id and a new change
  intent revision, whose promise differs only in `O-FR-001`.
- **SDD-INV-70** The Spec Kit profile is strict: wrong `requirements.paths`, a non-literal `root`, a
  `sibling` with a slash, an `archive` key, or a speckit/openspec field mix is `profile-invalid`.
- **SDD-INV-71** Resolve on Spec Kit: a planned feature gives its revision, an unplanned one gives
  `ErrNoChangeIntent`, a missing one gives `ErrUnknownChange`; OpenSpec id vectors are unchanged.

`proofbound init speckit` (CLI-INV-18 amended) writes the Spec Kit profile instead of OpenSpec's.

**Found by the documentation walkthrough (2026-11-19), CLI-INV-22:** the P5 `specdir` provider is also
rooted at `specs/` and rejected every Spec Kit file, so `sync all` failed on Spec Kit projects. When
HEAD's committed profile declares Spec Kit at `specs/`, `sync intent all`, `sync all` and `verify` omit
`specdir`.

## 6. Non-goals

- Spec Kit's `plan.md`, `research.md` and `contracts/` are not parsed. They count toward the change
  intent's digest only, so they never cause drift, which compares promises.
- No status mapping from `**Status**:` lines.
