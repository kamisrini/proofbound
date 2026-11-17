# `sdd` spec-driven-development provider contract — DRAFT 5 (change-scoped, history-native, per-state)

Status: **DRAFT 5 (R2 M0), not yet a live kernel contract.** Supersedes
[`R2-sdd-provider-spec-draft4.md`](R2-sdd-provider-spec-draft4.md) (reviewed NEEDS_WORK — 6 MED, 4 LOW,
no HIGH — in `docs/verification/verdicts/r2-m0-review-round4-20261117.md`) and drafts 1-3.
**Self-contained:** every definition the contract uses is in this file; earlier drafts are cited only
as history. Becomes `kernel/internal/connector/intent/sdd/SPEC.md` in the same M1 commit as the code and
tests it cites, together with every amendment listed in §12.

Scope (owner decision 2026-11-16, `VD-sdd-change-scoped-r2-k7q3wz`, unchanged): **R2 warrants OpenSpec
changes and Spec Kit features — the commits that implement a change, against the requirements that
change declares — read from the repository's history.** The main-spec lifecycle is not in R2.
Requirement: `BR-sdd-warranty-i9miv7` (revision proposed in §11).

## 0. Evidence first

Measured on a full clone of `Fission-AI/OpenSpec` at `f085c850`. Every count is printed by
`testdata/derive/history_probe.py`, which applies the parsing rules of §4.2 (OpenSpec's own fence,
section, and requirement rules). Round-4's reviewer reproduced each count independently.

| Observation (real history) | Count | Consequence |
|---|---|---|
| Commits on the git connector's walk (`rev-list --all`) | 1624 (120 merges, 1 root) | The provider walks exactly this set, under the git connector's read contract (§3) |
| Distinct SDD states over that walk | 338 | Mapping cost tracks states, not commits (§3) |
| Commits touching `openspec/changes`: HEAD path-limited / `--all --full-history` | 256 / 451 | Path-limited walks miss states; no path limiting is used (§3) |
| **Commits carrying a change both active and archived** (leftover after archive, from `7ced2a87`), on the provider's unsimplified `--all` walk | **23** (9 on the HEAD path-limited walk) | Unmappable for that change at that state only (§7) |
| …of which are **ref tips** | **4 remote branches** (`archive-changes`, `prepare-publish`, `update-openspec-agent-instructions`, `updating-agent-instructions`); 0 tags | A tip-fatal rule would fail every sync of the reference repository today (§10 D, SDD-INV-35) |
| Names archived twice at one commit | 0 | Still unmappable if it occurs (§5) |
| Changes archived without ever being observed active (squash-merge) | 16 / 108 | A name resolves when archived **by** the commit (§6) |
| Changes with zero requirement entries at some commit | 34 | Absence, never an error (§5) |
| Within-change requirement removals / statement changes | 47 (10 changes) / 10 (5 changes) | Content-addressed records (§4) |
| Requirements with no statement line | 4 | Statement falls back to the name (§4.2) |
| RENAMED `FROM` entries in a real delta section | 0 | RENAMED is tested on fixtures only (§4.2) |
| Distinct delta blobs / blobs whose obligations change under fence masking | 401 / **1** (`archive/2025-09-29-improve-init-onboarding/specs/cli-init/spec.md`: a column-0 `## TL;DR` inside a fence) | Unmasked, 2 of its requirements are silently dropped. That is under-mapping, so masking is required (§4.2, SDD-INV-39) |

**Draft 4's §0 corrected (round-4 F8).** "12 dual-presence commits" was a path-limited count; the
right count for the provider's walk is 23. The "47 fenced `FROM:` lines" row counted every `FROM:` line
in active deltas, fenced or not; it is removed. The fenced RENAMED example it described is **indented**,
so OpenSpec's column-0 section rule ignores it whether or not fences are masked. It is now the
*neutral* control of SDD-INV-39, not its motivation.

**How each round-4 finding is answered.**

| Finding | Answer |
|---|---|
| F1 payload size / invalid UTF-8 | Both are per-unit data causes (§7). `ValidateRevision` gains the append-time size bound for every provider (§12) |
| F2 catch-all masks provider bugs | Skips come only from a closed list of tenant-data causes. Any other failure fails `Revisions` (§7). Each fixture's exact Skip set is pinned (SDD-INV-46) |
| F3 incomplete amendment list | §12 lists every site that registers a provider today, found by grep and assigned to M1 or M3. M1's acceptance includes projecting `intent.sdd` events |
| F4 walk shares only the arguments | The production Reader is the git connector's own `gitcmd.Repo`: the same `run`, `refusePartial`, and `validateRefs` (§3, SDD-INV-51) |
| F5 INV-39 cannot discriminate | Rebuilt on the column-0 shape, with the indented example as the neutral control. The fence and section rules are pinned to OpenSpec's code (§4.2) |
| F6 tip-fatal survives; meaning flips | A fixture ref tip sits on the leftover state. INV-1/3/4/6/29 are retired and re-issued. Tombstone reasons are restored (§9) |
| F7 INV-38 vs INV-31/33 | Collisions are checked within one state only. A cross-state collision is two revisions of one id, bound by digest (§4.3) |
| F8 mis-scoped counts | Corrected above |
| F9 not self-contained | The profile schema, obligation `state`, archive-name rule, CRLF/BOM handling, and the README are fixed (§2.1, §4.2, §6) |
| F10 hand-offs | Resolve error sentinels (§6, SDD-INV-50). The fixture helper must be committed before M1 (§13) |

## 1. Purpose and boundary

Provider name `sdd`; source `intent.sdd`. It maps committed SDD artifacts onto the **existing**
canonical model — `proofbound.requirement.v1` and `proofbound.change-intent.v1` — through one
declarative profile committed in the tenant repository. A third provider of the intent connector: no
new connector, model, or event kind.

Family rules: **no fabrication** (no business decision, authority, relation, obligation, status, or
identity is invented; governance maps to nothing; requirements render `authorization: undeclared`);
**exact bytes** (digests cover committed bytes; statements are exact line bytes, §4.2); **commit tree**
(everything is read through the `Reader` at named commits). Proofbound **never writes** into tenant SDD
files, and in R2 the **main specs are not ingested**.

**Not wired into the git connector in M1.** M1 registers `intent.sdd` with the intent connector,
core, the CLI, projections, and gates (§12). The `sdd:` trailer token and everything that resolves a
tenant claim through the provider belong to M3 (§14).

## 2. Interface lock

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface {
    // Every commit the git connector ingests (§3), oldest first, under its read contract.
    Commits(ctx context.Context) ([]string, error)
    // The parents of commit, in order (empty for a root commit).
    Parents(ctx context.Context, commit string) ([]string, error)
    // The state key of commit: the object ids of the profile blob and of every SDD root tree the
    // profile names ("" when the profile is absent). Equal keys ⇒ byte-identical SDD inputs.
    StateKey(ctx context.Context, commit string) (string, error)
    // Every committed file at commit under the profile's SDD roots, plus the profile.
    Read(ctx context.Context, commit string) ([]Artifact, error)
    // Names of the entries directly under dir at commit (the parent archive check, §6).
    List(ctx context.Context, commit, dir string) ([]string, error)
}
type Provider struct { /* unexported */ }

var (
    ErrUnknownChange   = errors.New("sdd: no such change at this commit")         // §6
    ErrArchivedEarlier = errors.New("sdd: change was archived before this commit") // §6
)

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                             // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) // whole history (§3)
func (p *Provider) Skipped() []intent.Skip                                   // after Revisions (§7)
func (p *Provider) Resolve(ctx context.Context, commit, name string) (intent.Revision, error) // §6
```

`intent.Skip` is `struct { Commit, Unit, Cause string; Err error }` (intent SPEC §2 amendment, §12):
- `Commit` is the first commit, in `Commits` order, that carries the skipped state.
- `Unit` is `requirement <path>`, `change <name>`, or `profile`.
- `Cause` is one value from §7's closed list.
- `Err` wraps `intent.ErrUnmappable`.

`Revisions` returns an error for anything that is not on that list (§7).

### 2.1 Profile — `proofbound.sdd-profile.v1`

Path `.proofbound/sdd-profile.json`. A commit without it contributes nothing and has the empty state
key. The profile is strict JSON: UTF-8, objects with exactly the fields below, no unknown, missing,
duplicate, or `null` fields, and no unknown enum values. Globs: `*` matches within one path segment,
`**` across segments. Regexes are Go RE2.

```json
{ "schema": "proofbound.sdd-profile.v1", "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": "team-x" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "team-x" } ] }
```
```json
{ "schema": "proofbound.sdd-profile.v1", "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"],
                    "status": { "line": "^\\*\\*Status:?\\*\\*:? *(?P<status>.+?) *$", "map": { "Draft": "proposed" }, "absent": "proposed" } },
  "changes": { "mode": "feature-sibling", "sibling": "tasks.md" },
  "owners":   [ { "pattern": "specs/**", "owner": "team-x" } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "team-x" } ] }
```

Field rules:
- **`kit`** ∈ {`openspec`, `speckit`}, and fixes `changes.mode` (`delta-sections` / `feature-sibling`).
  A mismatch is a profile violation.
- **OpenSpec:** `requirements.paths` selects delta files, and every glob must lie under
  `changes.root`. `changes.status` ∈ {`proposed`, `accepted`} applies to every active change.
- **Spec Kit:** `requirements.status.line` must contain the named group `status`. `map` goes from the
  exact captured string to {`proposed`, `active`, `superseded`, `retired`}. `absent` is optional, with
  the same value set, and applies only to a file with no status line. `changes.sibling` is a file name.
- **Kit rules are code, not profile.** The section, requirement, rename (§4.2), and Spec Kit
  obligation-line rules are fixed by the contract, so no profile can make the parser fabricate.
- **`owners`** covers requirement records; **`sponsors`** covers change intents. Each is a non-empty
  list, matched against the record's path in declared order, first match wins
  (`VD-sdd-accountable-identity-ypcx7x`). Values are non-empty strings.

## 3. History-native ingestion — the git connector's walk, under its read contract

- **One reader contract (round-4 F4).** The production `Reader` is implemented on the git
  connector's `gitcmd.Repo`. Every git call goes through the same `run`:
  - `--no-replace-objects`, the same `-c` options, and the same scrubbed `GIT_*` environment;
  - `Commits` is the same function the git connector calls, so it refuses shallow repositories and
    grafts (`ErrShallow`) and validates refs before running the same `rev-list`.

  No argument list is restated anywhere (Law 2). A replace ref or graft therefore cannot change which
  commits are walked or what their parents are, so a fabricated root cannot occur (§6).
- **No path limiting and no history simplification.** Every listed commit is visited, merges and
  unmerged branches included.
- **Per distinct state.** The provider maps each distinct `StateKey` once (§4-§5) and returns the union,
  de-duplicated by `(kind, record id, artifact digest)`.
- **SDD-INV-31.** For every commit `c` that `Commits` lists and every name `n`, a successful
  `Resolve(c, n)` is a member of `Revisions()`. Both are the same function of `c`'s state and its
  parents' archive listings (§6), over the same commit set. No unit's mappability depends on any other
  state (§4.3).
- **Ordering of syncs** (INT-INV-9) and **idempotence** (INT-INV-5): unchanged from the intent family.

## 4. Requirement records (content-addressed)

### 4.1 What is a requirement record

- **OpenSpec:** one record per delta file matched by `requirements.paths` that has at least one
  requirement entry, whether active (`<root>/<name>/specs/<capability>/spec.md`) or archived
  (`<archive>/YYYY-MM-DD-<name>/specs/<capability>/spec.md`).
- **Spec Kit:** one record per matched feature spec with at least one obligation line.
- A file with zero entries yields no record.

### 4.2 Parsing (OpenSpec's rules, pinned to `src/core/parsers/` at `f085c850`)

- **Input.** The file must be valid UTF-8; otherwise the unit is unmappable (`invalid-utf8`). One
  leading U+FEFF is ignored. Lines are split on `\r\n`, `\r`, or `\n`, as OpenSpec does. Statements
  are the exact bytes of a line **without its terminator**. The digest still covers the file's exact
  bytes (§4.3).
- **Fence mask** (`code-fence.ts`):
  - An opener is a line matching `^\s*(`{3,}|~{3,})(.*)`, except a backtick opener whose remainder
    contains a backtick.
  - It is closed by the next line matching `^\s*(`{3,}|~{3,})\s*$` that uses the same marker
    character, at least as long as the opener.
  - Opener, body, and closer lines are masked. An unclosed fence masks through the end of the file.
  - Every rule below sees only unmasked lines.
- **Sections:**
  - A delta section starts at a line matching
    `^##\s+(ADDED|MODIFIED|REMOVED|RENAMED)\s+Requirements(?:[ \t]+#+)?\s*$` (case-insensitive).
  - It ends at the next line matching `^##\s+`, or at the end of the file.
  - Lines outside a delta section are ignored.
- **ADDED / MODIFIED / REMOVED entries:**
  - Each line matching `^###\s*Requirement:\s*(.+)\s*$` (case-insensitive) opens one entry. Its
    **name** is the capture, with a trailing `[ \t]+#+[ \t]*` removed, then trimmed.
  - The entry ends at the next requirement header or section boundary.
  - **Statement:** the first non-blank line of the entry before any line starting with `#`, as exact
    bytes. If there is none, the statement is the name.
- **RENAMED entries:** within a RENAMED section, lines matching
  ``^\s*[-*+]?\s*FROM:\s*`?###\s*Requirement:\s*(.+?)`?\s*$`` and the corresponding `TO:` line pair up
  in order. Each pair is one entry: the name is the `FROM` name, and the statement is the `TO` line's
  exact bytes. An unpaired line makes the record unmappable (`unpaired-rename`), as OpenSpec refuses
  it.
- **Spec Kit:**
  - Obligation lines match the fixed rule `^- \*\*(?P<id>(?:FR|SC|NFR)-[0-9]+)\*\*:? (?P<statement>.+)$`;
    id `O-<id>`, statement = the `statement` group's exact bytes.
  - Status comes from the first line matching `requirements.status.line`, then `map`; with no status
    line, `absent`. Anything else is unmappable (`status-unmapped`).
- **Obligation id (OpenSpec):** `O-<SEC>-<np>`, with `SEC` ∈ {`ADD`, `MOD`, `REM`, `REN`}. `np` is
  derived from the name:
  1. upper-case ASCII; every byte outside `[A-Z0-9]` → `-`; runs collapsed; leading and trailing `-`
     trimmed;
  2. if empty, the record is unmappable (`empty-name-part`);
  3. if longer than 28 bytes, the first 21 bytes, re-trimmed, + `-` + the first 6 upper-case hex
     characters of `SHA-256(name)`.

  Obligation ids are never positional. Two entries mapping to one id make the record unmappable
  (`obligation-collision`).
- **Payload values** (round-4 F9). Every obligation's `state` is `active`, a REMOVED entry included:
  the change's declared promise to remove is itself an active obligation of that change.
  `authorized_by` and `supersedes` are `[]`.
  - Record status: an active OpenSpec change → `proposed`; archived → `active`; Spec Kit → as above.
  - Owner: the first match in `owners`; none → unmappable (`owner-uncovered`).

### 4.3 Identity and digest

- **`slug(s)`:** lower-case ASCII; every byte outside `[a-z0-9]` → `-`; runs collapsed; trimmed;
  truncated to 40 bytes and re-trimmed. Shorter than 2 bytes → unmappable (`slug-too-short`).
- **`suffix(kind, key)`:** the first 6 bytes of `SHA-256(kit + "\x00" + kind + "\x00" + key)`, each
  byte `b` mapped to `"abcdefghijklmnopqrstuvwxyz0123456789"[b % 36]`. `kind` is `requirement` or
  `change-intent`.
- **Manifest:** one line `"<sha256-hex-of-exact-bytes>  <path>\n"` per input file, sorted by path
  bytes; `artifact_sha256 = SHA-256(manifest)`.
- **Requirement record:**
  - Inputs: the source file and the profile. `artifact_path` is the source path.
  - Key: `<name>/<capability>` (OpenSpec) or the feature directory name (Spec Kit).
  - Id: `BR-<slug(key)>-<suffix("requirement", key + "\x00" + artifact_sha256)>`.
  - Each distinct `(key, manifest)` yields exactly one revision (SDD-INV-44). Any byte change of the
    source file, its path, or the profile mints a new record. Intent SPEC §5 holds by construction.
- **The payload is a pure function of the manifest's bytes** (SDD-INV-26).
- **Collisions are checked within one state only (round-4 F7).** If two distinct keys at one state
  mint one record id, both units are unmappable (`record-id-collision`, SDD-INV-48). This is decidable
  from `c`'s state alone, so SDD-INV-31 and SDD-INV-33 hold.
  - Across states, two `(key, manifest)` pairs that share an id (probability about 2⁻³¹ per same-slug
    pair) are simply two revisions of one id.
  - Every reference binds `(source, id, digest)`, and projections key requirement rows the same way,
    so meaning stays unambiguous. This residual is stated, not hidden.

## 5. Change intents

- **OpenSpec:** one change intent per change **name**, id
  `CI-<slug(name)>-<suffix("change-intent", name)>`. At a state it `implements` every obligation of
  every requirement record of that change at that state.
- **Spec Kit:** one change intent per feature directory, only when `changes.sibling` exists there. It
  implements every obligation of the feature's record.
- **Payload values:**
  - `status`: active → `changes.status`; archived → `accepted`; Spec Kit → `accepted`. **New in draft
    5, for the reviewer to check:** in Spec Kit, `tasks.md` existing is the kit's only signal that
    work on a feature was committed to, and the intent is emitted only when it exists.
  - `declared_sponsor`: the first match in `sponsors`; none → unmappable (`sponsor-uncovered`).
  - `constrained_by` and `supersedes` are `[]` (`DEF-INTENT-CI-REVISION-001`).
  - `artifact_path` is the change directory. The manifest covers every file in it plus the profile.
- **No requirement records → no change intent** (SDD-INV-28).
- **Name reuse** — a name both active and archived, or archived twice, at one state — makes that change
  unmappable at that state (`name-reuse`, SDD-INV-45). Its requirement records are unaffected.
- **Cascade:** a change intent whose target record is unmappable at that state is unmappable
  (`cascade`).

## 6. Resolution

**Archive names (round-4 F9).** An entry directly under `changes.archive` names change `n` exactly when
its name matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$` and the capture equals `n` byte for byte. There
are no globs.

`Resolve(c, n)` returns the change intent of `n` at `c`'s state when:
1. **active:** `<root>/<n>/` exists at `c` and `n` is mappable there; or
2. **archived by `c`:** exactly one archive entry names `n` at `c`, and **no parent** of `c` has an
   archive entry naming `n`.
   - **A merge** whose side branch archived `n` does not archive it; the side-branch commit does.
   - **A root commit** archives what it carries. Roots are real roots, because replace refs are ignored
     and grafts and shallow clones are refused (§3).

Errors (SDD-INV-50):
- `ErrUnknownChange`: no active or archive entry names `n` at `c`.
- `ErrArchivedEarlier`: an archive entry names `n` at `c` and at some parent.
- An error wrapping `intent.ErrUnmappable`: `n` is unmappable at `c`.
- Anything else is a reader failure.

M3 needs this split to report a bad claim per commit instead of failing the sync (§14).

`Resolve` reads only `c`'s state and its parents' archive listings (SDD-INV-33). **Verdicts bind to the
change:** an `obligation-verdict.v2` names the commit, the exact change-intent revision, and its targets.

## 7. Fail-closed per state, never per history — a closed list of data causes

A batch-fatal rule over full history is a permanent wedge, because past commits cannot be edited. So:

- **Units:** a requirement record, a change intent at a state, or a whole state (the profile).
- **The closed list of tenant-data causes** (`Skip.Cause`):
  - `profile-invalid` (any §2.1 violation, including non-UTF-8)
  - `invalid-utf8`
  - `owner-uncovered`, `sponsor-uncovered`
  - `status-unmapped`
  - `empty-name-part`, `slug-too-short`
  - `obligation-collision`, `unpaired-rename`
  - `record-id-collision`
  - `name-reuse`
  - `payload-too-large` (the canonical payload exceeds `core.MaxPayloadBytes`, checked by the provider
    before it emits — round-4 F1)
  - `cascade`

  Each such unit contributes **no revision** and **one `Skip`**. Every other unit at every commit is
  still emitted.
- **Everything else fails `Revisions` (round-4 F2):** a candidate that fails `intent.ValidateRevision`
  for any reason not on the list (non-canonical JSON, unsorted obligations or references, registry or
  envelope mismatch), a reader or ref failure, `ErrShallow`, or context cancellation.
  - Those are provider or environment defects, not tenant content. A sync that fails on one is
    recovered by fixing the provider or the clone, never by editing history.
  - The provider never pre-validates in order to *hide* a failure: it checks the data causes, builds
    the candidate, then calls `ValidateRevision`, and an error at that point is returned.
- **Batch-safe.** Every returned revision passes `ValidateRevision`, which after §12 also enforces the
  append-time size bound. The batch has no duplicate `(source, id, digest)` and no dangling same-source
  reference. So `Sync` validates, orders, and appends every `sdd` batch end to end (SDD-INV-47).
- **Failure moves to the claim.** Only a claim on an unmappable state fails, at `Resolve`.
- **Skips are derived, not ledgered.** They are recomputed on every sync from the refs. The intent
  connector's `Result` carries them, and the CLI prints them and exits 0. A skip can only *withhold*
  trust; it never grants it.
- **The family's `ErrUnmappable` STOP** (intent SPEC §6, INT-INV-11) is amended for `sdd` (§12). For
  `sdd`, `ErrUnmappable` marks one unit that the canonical model and profile cannot represent, and it
  is reported as a `Skip`. The foreign-fixture STOP for `records` / `specdir` is unchanged.

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

IDs are permanent from draft 1 on (pinned in `docs/invariants.lock` once the file lands). New
invariants append and none is reused. **An invariant whose meaning changes is retired, its tombstone
keeps the reason, and it is re-issued under a new number.**

1. **SDD-INV-1 — RETIRED (draft 5): profile violations are now per-state, not sync-fatal; replaced by INV-40.**
2. **SDD-INV-2 — Only requirements and change intents are emitted; governance maps to nothing.**
3. **SDD-INV-3 — RETIRED (draft 5): uncovered identity is now per-unit; replaced by INV-41.**
4. **SDD-INV-4 — RETIRED (draft 5): an unmapped status is now per-unit; replaced by INV-42.**
5. **SDD-INV-5 — Obligation ids are never positional.**
6. **SDD-INV-6 — RETIRED (draft 5): obligation defects are now per-record; replaced by INV-43.**
7. **SDD-INV-7 — Record ids are deterministic.**
8. **SDD-INV-8 — RETIRED (draft 4): the digest is a manifest digest; replaced by INV-34.**
9. **SDD-INV-9 — Delta sections map onto the closed relation vocabulary.**
10. **SDD-INV-10 — Feature-sibling change intents are feature-level only.**
11. **SDD-INV-11 — RETIRED (draft 4): resolution also reads parent archive listings; replaced by INV-33.**
12. **SDD-INV-12 — Canonical payload bytes are stable across runs.**
13. **SDD-INV-13 — Two kits run on one provider with only the profile differing.**
14. **SDD-INV-14 — RETIRED (draft 4): unmappable content no longer stops the sync; replaced by INV-35/36.**
15. **SDD-INV-15 — RETIRED (draft 3): archived changes are now ingested.**
16. **SDD-INV-16 — RETIRED (draft 3): requirement records are content-addressed; change ids remain name-derived (INV-25).**
17. **SDD-INV-17 — RETIRED (draft 3): archived delta records are `active`.**
18. **SDD-INV-18 — A change's artifact digest covers the exact bytes of every file it reads.**
19. **SDD-INV-19 — RETIRED (draft 3): replaced by INV-27.**
20. **SDD-INV-20 — Archiving a change never makes a later sync fail.**
21. **SDD-INV-21 — RETIRED (draft 3): no fingerprinted obligation ids.**
22. **SDD-INV-22 — RETIRED (draft 3): no tombstones; records never change.**
23. **SDD-INV-23 — RETIRED (draft 3): no delta fold.**
24. **SDD-INV-24 — RETIRED (draft 4): the walk is no longer HEAD-scoped; replaced by INV-31.**
25. **SDD-INV-25 — RETIRED (draft 5): a cross-state id collision yields two revisions of one id; replaced by INV-44.**
26. **SDD-INV-26 — The payload is a pure function of the manifest bytes.**
27. **SDD-INV-27 — RETIRED (draft 4): "archived by" now means absent from every parent; replaced by INV-32.**
28. **SDD-INV-28 — A commit with no requirement records for a change yields no change intent, not an error.**
29. **SDD-INV-29 — RETIRED (draft 5): name reuse is now per change and state; replaced by INV-45.**
30. **SDD-INV-30 — Main specs are not ingested in R2.**
31. **SDD-INV-31 — Every commit the git connector lists contributes its state, and every successful Resolve is a member of Revisions.**
32. **SDD-INV-32 — A name resolves only to a change active at the commit, or archived by it and in none of its parents.**
33. **SDD-INV-33 — Resolution reads only the commit's state and its parents' archive listings.**
34. **SDD-INV-34 — The artifact digest is the SHA-256 of the manifest of exact committed bytes.**
35. **SDD-INV-35 — No tenant content at any commit makes Revisions fail.**
36. **SDD-INV-36 — RETIRED (draft 5): skips now come only from the closed cause list; replaced by INV-46.**
37. **SDD-INV-37 — RETIRED (draft 5): the guarantee now covers append, not only validation; replaced by INV-47.**
38. **SDD-INV-38 — RETIRED (draft 5): collisions are now per state; replaced by INV-48.**
39. **SDD-INV-39 — Nothing inside a code fence is read as a section, entry, or rename.**
40. **SDD-INV-40 — The profile is strict JSON; a violation makes that state unmappable, never the sync.**
41. **SDD-INV-41 — Accountable identity comes only from the profile at the same state; uncovered makes that unit unmappable.**
42. **SDD-INV-42 — A Spec Kit status comes only from the explicit map or the declared absent value; otherwise that record is unmappable.**
43. **SDD-INV-43 — Duplicate ids, empty name-parts, and unpaired renames make that record unmappable; zero entries yield no record.**
44. **SDD-INV-44 — Each distinct (key, manifest) yields exactly one requirement revision, with a content-addressed id.**
45. **SDD-INV-45 — A reused change name makes that change unmappable at that state only.**
46. **SDD-INV-46 — Only the closed list of data causes yields a Skip, each fixture's Skip set is exactly the pinned one, and dependants cascade.**
47. **SDD-INV-47 — Every returned batch syncs end to end: validation, ordering, and append.**
48. **SDD-INV-48 — Within one state, two keys minting one record id are both unmappable.**
49. **SDD-INV-49 — A failure outside the closed cause list fails Revisions; it is never a Skip.**
50. **SDD-INV-50 — Resolve distinguishes unknown, archived-earlier, and unmappable by sentinel.**
51. **SDD-INV-51 — The reader shares the git connector's read contract: replace refs ignored, shallow and graft refused, refs validated.**
52. **SDD-INV-52 — A sync containing intent.sdd events projects and rebuilds.**

Each proving test names **the wrong implementation it kills**. Where it matters, it also names **a
neutral change that must survive**. Retired invariants have no row.

| Invariant | Statement | Proving test | Kills / survives |
|---|---|---|---|
| SDD-INV-2 | No business decision is emitted; requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted | kills: mapping `proposal.md` to a BD |
| SDD-INV-5 | Inserting, deleting, or reordering one entry changes no other obligation id | sdd_test.go::TestObligationIDsAreNotPositional | kills: index-based ids |
| SDD-INV-7 | The §4.3 / §5 id vectors in testdata (including the `kit` and `kind` literals) match exactly | sdd_test.go::TestRecordIDsAreDeterministic | kills: a suffix omitting `kind` |
| SDD-INV-9 | Every ADDED, MODIFIED, REMOVED, RENAMED entry is an `implements` target | sdd_test.go::TestDeltaSectionsMapToClosedRelations | kills: dropping REMOVED |
| SDD-INV-10 | One feature-level change intent, only when the sibling exists | sdd_test.go::TestFeatureSiblingIsFeatureLevel | kills: per-task intents |
| SDD-INV-12 | Repeated runs over one history emit byte-identical payloads | sdd_test.go::TestCanonicalizationStability | kills: map-ordered obligations |
| SDD-INV-13 | (a) Both reference fixtures map through one constructor. (b) Spec Kit-shaped files in an `openspec`-kit repository are ignored, and flipping only `kit` changes the mapping | sdd_test.go::TestTwoKitsOneProvider | kills: kit sniffing from the tree |
| SDD-INV-18 | The change manifest changes when any file of the change directory or the profile changes | sdd_test.go::TestChangeManifestDigest | kills: a manifest over delta files only |
| SDD-INV-20 | The leftover-after-archive state, a `--skip-specs` archive, and a deleted main spec all sync | sdd_test.go::TestArchiveNeverBreaksSync | kills: batch-fatal name reuse |
| SDD-INV-26 | (a) Changing the owner changes the digest. (b) Changing a non-manifest input (a main spec, another change's files, `src/`, git author or date) leaves every revision byte-identical | sdd_test.go::TestPayloadIsPinnedByManifest | kills: reading outside the manifest; survives: non-manifest edits |
| SDD-INV-28 | A change with only `proposal.md` yields no change intent and no error | sdd_test.go::TestEmptyChangeIsAbsent | kills: emitting a target-less CI |
| SDD-INV-30 | Files under the main-spec path never produce a revision | sdd_test.go::TestMainSpecsAreNotIngested | kills: ingesting `openspec/specs` |
| SDD-INV-31 | Over `branches`: `Commits` equals the git connector's list, and every successful `Resolve(c, n)` is in `Revisions()` | sdd_test.go::TestResolveIsMemberOfRevisions | kills: a HEAD-only walk; a path-limited walk (TREESAME) |
| SDD-INV-32 | Active; archived-by (single parent); side-branch archive then merge (the merge fails, the side commit resolves); root with archive (resolves); dual presence (unmappable); archived earlier; unknown | sdd_test.go::TestNativeNameResolution | kills: a first-parent-only archive check |
| SDD-INV-33 | A reader that fails on any other commit, and on any parent read other than an archive listing | sdd_test.go::TestResolveReadSet | kills: resolving from a cached HEAD state |
| SDD-INV-34 | The digest equals a manifest recomputed independently from `git cat-file` bytes; a one-byte edit changes it; a CRLF copy of a delta has a different digest but the same obligations | sdd_test.go::TestDigestIsManifestOfExactBytes | kills: digesting normalized bytes; survives: CRLF for parsing |
| SDD-INV-35 | Every data cause placed (a) at a non-tip commit fixed later and (b) **at a ref tip** (`branches` step 14 `stale`) leaves `Revisions` without error and every other unit emitted | sdd_test.go::TestHistoricalFailClosedIsRecoverable | kills: batch-fatal; **tip-fatal** (§10 D) |
| SDD-INV-39 | `fenced/improve-onboarding` (column-0 `## TL;DR` inside a fence) yields all 3 MODIFIED obligations; `fenced/document-renames` (indented example) yields exactly its 1 under masked and unmasked parsing alike | sdd_test.go::TestCodeFencesAreMasked | kills: an unmasked parser (yields 1 of 3); survives: the indented neutral |
| SDD-INV-40 | Each §2.1 violation (unknown, missing, null, or duplicate field; unknown enum; regex lacking `status`; kit/mode mismatch; glob outside root; non-UTF-8) skips that state with `profile-invalid` | sdd_test.go::TestProfileIsStrict | kills: a lenient decoder; a sync-fatal profile error |
| SDD-INV-41 | Owner and sponsor come from their list's first match at the same state; uncovered → `owner-uncovered` / `sponsor-uncovered` | sdd_test.go::TestIdentityOnlyFromProfile | kills: falling back to the git author |
| SDD-INV-42 | A status outside `map` with no `absent` → `status-unmapped` | sdd_test.go::TestStatusMapFailsClosed | kills: defaulting to `proposed` |
| SDD-INV-43 | Colliding ids, empty name-parts, and unpaired renames are skipped with their cause; a zero-entry file yields no record and no skip | sdd_test.go::TestObligationDefectsArePerRecord | kills: last-wins; `O-ADD-` |
| SDD-INV-44 | Editing a delta yields a new record id and leaves the earlier record unchanged; each `(key, manifest)` has one revision | sdd_test.go::TestRequirementRecordsAreContentAddressed | kills: name-keyed BR ids |
| SDD-INV-45 | Name reuse skips that change at that state; at the next state the change maps again; its records are emitted throughout | sdd_test.go::TestReusedNameIsPerState | kills: picking the active copy; a sticky skip |
| SDD-INV-46 | Every fixture history's Skip set equals `testdata/expected-skips.json` exactly (units, causes, first steps); a CI targeting a skipped BR is skipped as `cascade` | sdd_test.go::TestSkipSetsArePinned | kills: a partial-skip bug (e.g. map-ordered obligations skipped instead of failing) |
| SDD-INV-47 | Every fixture history, the hostile ones and a just-under-limit payload included, passes `Sync` against a test appender with zero errors; a payload over the limit is `payload-too-large` and never reaches append | sdd_test.go::TestBatchSyncsEndToEnd | kills: emitting an oversize payload (partial append) |
| SDD-INV-48 | A pinned colliding key pair at one state (found offline, in testdata) makes both units unmappable; the same pair at two states yields two revisions of one id | sdd_test.go::TestRecordIDCollisionIsPerState | kills: first-wins; a cross-state lookup |
| SDD-INV-49 | A provider mutant that emits unsorted obligations makes `Revisions` return an error, not a Skip | sdd_test.go::TestProviderDefectsFailNotSkip | kills: the draft-4 catch-all |
| SDD-INV-50 | `errors.Is` distinguishes `ErrUnknownChange`, `ErrArchivedEarlier`, and `intent.ErrUnmappable`; a failing reader yields none of them | sdd_test.go::TestResolveErrorClasses | kills: plain `fmt.Errorf` errors |
| SDD-INV-51 | Over a repository with a `git replace --graft` ref, `Commits` and `Parents` equal the git connector's (no false root); a shallow clone and an `info/grafts` file make `Revisions` fail with `ErrShallow` | gitcmd_test.go::TestSDDReaderSharesReadContract | kills: a bare `git rev-list` reader |
| SDD-INV-52 | Syncing `openspec-history` and running a projection rebuild succeeds; requirement and change-intent rows exist for `intent.sdd` | sdd_projection_test.go::TestSDDEventsProject | kills: an unregistered projection route |

## 10. Decisions this draft settles or withdraws

- **A, B, C:** unchanged from draft 3 (warranty across archive withdrawn; main-spec lifecycle
  withdrawn from R2; Spec Kit lifecycle answered by content addressing).
- **D: fail-closed scope.** Per unit, never per history and never per ref tip. OpenSpec has four branch
  tips on an unmappable state today (§0), so a tip-fatal rule would fail the reference repository now.
  SDD-INV-35(b) kills that implementation.
- **E: walk.** The git connector's walk, under its full read contract, unsimplified, deduplicated by
  state (§3).
- **F (new): data causes vs defects.** Tenant data yields Skips from a closed list. Everything else
  fails the sync (§7). A skip can withhold trust but never hide a provider defect.

## 11. Proposed `BR-sdd-warranty-i9miv7` revision 0003 (owner approves the wording)

Unchanged from draft 4; round 4 found the wording sound. Retire **O-5** and **O-9** (statements kept).
Add:
- **O-10 — History-native exact binding:** every commit the git connector ingests contributes the
  revisions of its SDD state. Each revision's `artifact_sha256` is a manifest digest over the exact
  committed bytes that fully determine its payload. A native change name resolves at a commit to
  exactly one change-intent revision — the change active there, or archived by that commit and by none
  of its parents — and that revision is in the ingested batch. Content the profile cannot map is
  reported per state, and fails only a claim on that state.
- **O-12 — R2 scope:** main specs are not ingested; R2's warranty claims concern a change and the
  commits that implement it, never the state of a main spec.

**O-11 (change drift) is not proposed** (round-3 F3). O-1..O-4 and O-8 are unchanged. Milestones:
O-1 / O-2 / O-10 (M1), O-3 / O-4 (M1-M2, M5), O-8 (M3), O-12 (all).

## 12. Amendments that land with M1 — complete list (round-4 F3)

Found by `grep -rn 'SourceIntentSpecdir\|"specdir"\|intent\.specdir'` over `kernel/` and `scripts/`
(non-test) at `f6813bf`. Every site that registers a provider today is listed and assigned. The M1
review re-runs the grep.

**Kernel, M1:**

| Site | Change |
|---|---|
| `core/core.go:56` | add `SourceIntentSDD Source = "intent.sdd"` |
| `connector/intent/intent.go:134` (`New`) | admit provider name `sdd` |
| `connector/intent/intent.go:215` (`ValidateRevision`) | admit `intent.sdd`, and **enforce `len(canonical payload) ≤ core.MaxPayloadBytes`** for every provider — closes the latent partial-append (round-4 F1) for `records`/`specdir` too |
| `connector/intent` (`Result`) | additive `Skip` type, optional `interface{ Skipped() []Skip }`, `ProviderResult.Skipped` (omitted when empty) |
| `projections/projections.go:296` (`eventRoute`) | route `intent.sdd` intent kinds |
| `gates/gates.go:304` (`intent-reference-integrity` matcher) | include `intent.sdd` |
| `cli/cli.go:37, 90, 247, 514` | `sync intent sdd`; `all` includes it; usage string |
| `connector/git/gitcmd` | the production `sdd` Reader on `Repo` (§3); no git-connector behaviour change |

**Scripts and generated artifacts, M1:**

| Site | Change |
|---|---|
| `scripts/p6-route-matrix.sh:19-24` | add the three `intent.sdd` routes, then regenerate the `@generated` route matrix (memory: route-matrix staleness trap) |
| `scripts/p6-census.sh:233-235` | add the `intent.sdd` rows |

**Intent SPEC text, M1:**
- **§1:** point at `sdd/SPEC.md`.
- **§2:** the `Skip` report.
- **§4:** `sdd`'s manifest digest and parent archive reads.
- **§6 and INT-INV-11:** add "for `sdd`, `ErrUnmappable` marks a unit and is reported as a `Skip`
  (`sdd/SPEC.md` §7); the foreign-fixture STOP applies to `records` and `specdir`".
- **INT-INV-4:** its test gains an oversize-payload row that fails before any append.

**Owned by M3, not M1** (resolving tenant claims):
- `connector/git/connector.go:198`, the provider allowlist;
- `projections/projections.go:723`, the `IntentRef` provider check;
- `scripts/intent-delivery-applicability.sh:93`, the trailer id regex.

M1 must not touch these.

**M1 acceptance** includes SDD-INV-52: a sync containing `intent.sdd` events, followed by a projection
rebuild.

## 13. Test derivations and fixtures

- **The fixture helper is committed before M1 builds on it** (round-4 F10). The extended
  `fixture_test.go` (PARENTS/REF commit graphs, `requireProfiles`, the branch and fenced self-tests)
  is written and passing but uncommitted; it is a governed path and needs the owner's authorization.
- **Commit-graph fixture** `testdata/branches/` (14 steps):
  - a root commit with an archive;
  - an edit / claim / revert on `feature`, merged `--no-ff` (TREESAME);
  - a side-branch archive, then merge;
  - an unmerged claim on `wip`;
  - the leftover-after-archive state with a claim during it;
  - the leftover's removal;
  - **a `stale` ref tip left on the leftover state** (SDD-INV-35(b)).
- **Fenced fixture** `testdata/fenced/`: the discriminating column-0 case and the indented neutral
  control (SDD-INV-39). Verified with the probe's §4.2 parser: masked 3 obligations, unmasked 1; the
  neutral yields 1 either way.
- **Pinned Skip sets** `testdata/expected-skips.json` (SDD-INV-46):
  - `openspec-history`, `speckit-history`, and `fenced`: empty.
  - `branches`: exactly `change add-z` / `name-reuse`, first at step 11.
  - `unmappable/duplicate-obligation`: the record (`obligation-collision`) and its change (`cascade`).
  - `unmappable/reused-name`: `change add-login` / `name-reuse`.
- **Every lifecycle test runs over a whole history.** The fail-closed tables (INV-35, 40-43, 45, 48)
  have one row per cause, each placed at a non-tip commit *and* at a ref tip.
- **Calibration:** every §9 row names its kill. M1 runs those mutants through
  `scripts/mutant-apply.sh` before its counts are believed.
- **Probe honesty:** `testdata/derive/history_probe.py` implements §4.2 and prints every §0 count. It is
  re-run at M1 against a pinned OpenSpec commit.

## 14. Non-goals and boundaries owned elsewhere

- Main-spec ingestion, folding, tombstones, and carry-over (deferred with the main-spec lifecycle).
- Authoring or writing tenant SDD files; governance → business decisions; YAML; heuristic extraction;
  Spec Kit per-task intents; API / DB providers; more than one profile per repository; ledgering skips.
- **M3 entry conditions:**
  1. the `sdd:` trailer token and its decision record;
  2. anti-spoofing against a native name the git connector reads from the commit tree itself,
     including the profile's `kit`;
  3. **per-commit handling of an unresolvable claim.** Today the git connector fails the sync closed,
     which is the same wedge class: a claim at one of the 23 dual-presence commits would fail every git
     sync. M3 uses §6's sentinels to report such a claim per commit, or proves the trailer gate refuses
     it first;
  4. the three M3 sites in §12;
  5. the readiness gate on `sdd` chains.
- **M4:** change-drift reporting (only if re-proposed with a ledger-evaluable definition) and the status
  export.
