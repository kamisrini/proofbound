# `sdd` spec-driven-development provider contract — DRAFT 6 (change-scoped, history-native, per-state)

Status: **DRAFT 6 (R2 M0), not yet a live kernel contract.** Supersedes
[`R2-sdd-provider-spec-draft5.md`](R2-sdd-provider-spec-draft5.md), which was reviewed NEEDS_WORK (3
MED, 4 LOW, no HIGH) in `docs/verification/verdicts/r2-m0-review-round5-20261117.md`, and drafts 1-4.
**Self-contained:** every definition the contract uses is in this file, and earlier drafts are cited
only as history. This draft becomes `kernel/internal/connector/intent/sdd/SPEC.md` in the same M1
commit as the code and tests it cites, together with every amendment listed in §12.

Scope (owner decision 2026-11-16, `VD-sdd-change-scoped-r2-k7q3wz`, unchanged): **R2 warrants OpenSpec
changes and Spec Kit features — the commits that implement a change, against the requirements that
change declares — read from the repository's history.** The main-spec lifecycle is not in R2.
Requirement: `BR-sdd-warranty-i9miv7` (revision proposed in §11).

**Owner decisions taken for this draft (2026-11-17).**
- `sync intent all` does **not** include `sdd`; the provider runs only when selected by name (§12,
  SDD-INV-56).
- The replay helper is committed with this draft, as an M1 precondition.

## 0. Evidence first

Measured on a full clone of `Fission-AI/OpenSpec` at `f085c850`. Every row is printed by
`testdata/derive/history_probe.py`. Since draft 6 the probe walks **the provider's own walk**
(unsimplified `rev-list --all`, date order) and parses with §4.2's rules. The two exceptions are
labelled.

| Observation (real history) | Count | Consequence |
|---|---|---|
| Commits on the walk / merges / roots | 1624 / 120 / 1 | The provider walks this set (§3) |
| Distinct SDD states | 338 | Mapping cost tracks states (§3) |
| Commits touching `openspec/changes` on HEAD path-limited / `--all --full-history` (labelled: other walks) | 256 / 451 | Path-limited walks miss states (§3) |
| Changes observed active / archived | 135 / 108 | — |
| **Commits carrying a change both active and archived** | **23** (1 name) | Unmappable for that change at that state only (§7) |
| …of which are ref tips | **4 remote branches**, 0 tags; on the HEAD walk, `main` itself sat on this state for 9 commits | Neither tip-fatal nor HEAD-fatal is acceptable (§10 D, SDD-INV-35) |
| Names archived twice at one commit | 0 | Still unmappable if it occurs (§5) |
| Changes archived without ever being observed active (squash-merge) | **12** / 108 (draft 5's 16 came from the path-limited walk) | A name resolves when archived **by** the commit (§6) |
| Changes with zero requirement entries at some commit | 38 | Absence, never an error (§5) |
| Requirement removals / statement changes between consecutive observations of a change (date order across all refs, so branch interleaving inflates both) | 83 (17 changes) / 39 (9 changes) | Content-addressed records (§4) |
| Requirements with no statement line (statement = name) | 4 | §4.2 fallback |
| RENAMED `FROM` entries in a real delta section | 0 | RENAMED is tested on fixtures only (§4.2) |
| Distinct delta blobs (labelled: all objects reachable on any ref) / blobs whose obligations change under fence masking | 401 / **1** (`improve-init-onboarding/specs/cli-init`: a column-0 `## TL;DR` inside a fence) | Masking is required (§4.2, SDD-INV-39) |

**Parser fidelity, measured.** The probe's §4.2 parser reproduces OpenSpec's reader on every row of
round 5's divergence table (§4.2):
- the REMOVED bullet form;
- `## ADDED  Requirements` with two spaces;
- an NBSP after `##`;
- a `## ` line with one trailing space;
- `## ADDED Requirementſ`;
- interleaved RENAMED lines.

**How each round-5 finding is answered.**

| Finding | Answer |
|---|---|
| F1 tenant bytes outside the list | The cause list is derived from every `ValidateRevision`, store, and projection check that tenant bytes can reach (§7). New causes: `statement-blank`, `obligation-id-invalid`, `nul-byte`, `non-regular-file`. The reader no longer reads per-commit file lists, so a non-UTF-8 path outside the SDD roots never touches `sdd` (§3) |
| F2 NUL passes to the projector | `nul-byte` is a per-unit cause, and `ValidateRevision` rejects U+0000 in any payload string for every provider (§12). SDD-INV-52 projects a history carrying every hostile byte |
| F3 not OpenSpec's reader | §4.2 is re-pinned to `parseDeltaSpec`, with JavaScript regex semantics spelled out as explicit classes. The one deliberate divergence (the statement) is labelled |
| F4 CLI rows; `all` | §12 lists `syncIntentOnStore` and the command enum. `cli.go:247` moves to M3. `all` excludes `sdd` (owner decision) |
| F5 Spec Kit status fabricated | `changes.status` is declared in the Spec Kit profile, as draft 2 had it |
| F6 self-tests weak | The pins are mechanised: skip units must exist, and fenced cases carry masked, naive, and unmasked lists. There are fence-precision fixtures, and a fixture where HEAD itself is bad |
| F7 drift | READMEs, plan, the §0 walk, probe statements (exact bytes), and the Spec Kit state roots (§2.1) are fixed |

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
type Artifact struct { Path string; Bytes []byte; Regular bool } // Regular: git mode 100644/100755; symlinks and gitlinks carry Regular=false and no Bytes (§7 non-regular-file)
type Reader interface {
    // Every commit the git connector ingests (§3), oldest first, under its read contract.
    Commits(ctx context.Context) ([]string, error)
    // The parents of commit, in order (empty for a root commit).
    Parents(ctx context.Context, commit string) ([]string, error)
    // The state key of commit: the object ids of the profile blob and of every SDD root tree the
    // profile names ("" when the profile is absent). Equal keys ⇒ byte-identical SDD inputs.
    StateKey(ctx context.Context, commit string) (string, error)
    // Every tree entry at commit under the profile's SDD roots, plus the profile (paths as raw bytes;
    // the provider applies §7's invalid-utf8 and non-regular-file causes).
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
duplicate, or `null` fields, and no unknown enum values. Globs: `*` matches within one path segment and
`**` matches across segments, **but `requirements.paths` may use `*` only**. Regexes are Go RE2.

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
  "changes": { "mode": "feature-sibling", "sibling": "tasks.md", "status": "proposed" },
  "owners":   [ { "pattern": "specs/**", "owner": "team-x" } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "team-x" } ] }
```

Field rules:
- **`kit`** ∈ {`openspec`, `speckit`}, and fixes `changes.mode` (`delta-sections` / `feature-sibling`).
  A mismatch is a profile violation.
- **`changes.status`** ∈ {`proposed`, `accepted`}, required for both kits (round-5 F5). It is the status
  of every active OpenSpec change and of every Spec Kit feature with its sibling. Neither kit records
  an approval, so the tenant declares what one means in its workflow; Proofbound never infers it.
- **OpenSpec:** every `requirements.paths` glob must lie under `changes.root`.
- **Spec Kit:**
  - `requirements.status.line` must contain the named group `status`;
  - `map` goes from the exact captured string to {`proposed`, `active`, `superseded`, `retired`};
  - `absent` is optional, takes the same values, and applies only to a file with no status line;
  - `changes.sibling` is a plain file name.
- **SDD roots, which define the state key (round-5 F7).** The roots are the longest wildcard-free
  directory prefix of each `requirements.paths` glob, plus, for OpenSpec, `changes.root` and
  `changes.archive`. Every file a unit reads (a change directory, a feature directory) lies under a
  root, so an edit to any of them changes the state key. Spec Kit's feature directory is the directory
  of the matched `spec.md`, and its key is that directory's name.
- **Kit rules are code, not profile.** The section, requirement, rename, and Spec Kit obligation-line
  rules (§4.2) are fixed by the contract.
- **`owners`** covers requirement records; **`sponsors`** covers change intents.
  - Each list is non-empty and is matched against the record's path in declared order; the first
    match wins (`VD-sdd-accountable-identity-ypcx7x`).
  - A value must be non-blank (§4.2's blank) and free of U+0000. Otherwise the profile is invalid.

## 3. History-native ingestion — the git connector's walk, under its read contract

- **One read contract (round-4 F4), without the per-commit file lists (round-5 F1).** The production
  `Reader` is implemented on the git connector's `gitcmd.Repo`. Every git call goes through the same
  `run`: `--no-replace-objects`, the same `-c` options, and the same scrubbed environment.
  - `Commits` runs the same `refusePartial` (`ErrShallow` on a shallow or grafted repository) and
    `validateRefs`, then the same `rev-list`, through one function factored out of `gitcmd.Commits`
    (Law 2).
  - Unlike `gitcmd.Commits`, it does **not** call `readCommit`/`files`, which fail on any non-UTF-8
    path in any commit's diff. The provider reads only `ls-tree` and `cat-file` under the SDD roots.
  - A non-UTF-8 path **outside** the roots therefore never reaches `sdd`. **Inside** the roots, it
    makes its unit unmappable (`invalid-utf8`).
  - The git connector's own failure on such paths is pre-existing git-connector behaviour, recorded in
    §14 and not inherited.
- **No path limiting and no history simplification.** Every listed commit is visited.
- **Per distinct state.** The provider maps each distinct `StateKey` once and returns the union,
  de-duplicated by `(kind, record id, artifact digest)`.
- **SDD-INV-31.** For every commit `c` that `Commits` lists and every name `n`, a successful
  `Resolve(c, n)` is a member of `Revisions()`. Both are the same function of `c`'s state and its
  parents' archive listings, over the same commit set. No unit's mappability depends on another state.
- **Ordering of syncs** (INT-INV-9) and **idempotence** (INT-INV-5): unchanged.

## 4. Requirement records (content-addressed)

### 4.1 What is a requirement record

- **OpenSpec:** one record per delta file matched by `requirements.paths` that has at least one
  requirement entry, whether active (`<root>/<name>/specs/<capability>/spec.md`) or archived
  (`<archive>/YYYY-MM-DD-<name>/specs/<capability>/spec.md`).
- **Spec Kit:** one record per matched feature spec with at least one obligation line.
- A file with zero entries yields no record.

### 4.2 Parsing — OpenSpec's delta reader, pinned to `parseDeltaSpec` at `f085c850`

**Source of truth (round-5 F3).** `src/core/parsers/requirement-blocks.ts` `parseDeltaSpec` and its
helpers (`splitTopLevelSections`, `getSectionsCaseInsensitive`, `parseRequirementBlocksFromSection`,
`parseRemovedNames`, `parseRenamedPairs`, `normalizeRequirementName`), plus `code-fence.ts`.
`spec-structure.ts`'s `DELTA_HEADER` is a main-spec diagnostic and is **not** used.

**Regex dialect.** OpenSpec's regexes are ECMAScript, without the `u` flag. Go RE2 differs, so every
rule below is written with explicit classes:
- **`WS`** — JavaScript `\s`: `[\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`.
- **`DOT`** — JavaScript `.`: `[^\n\r\x{2028}\x{2029}]`.
- **Case-insensitive matching** applies only to the literal `Requirement:` in the requirement header,
  and folds **ASCII letters only**. This matches JavaScript's non-unicode `/i`, where no non-ASCII code
  point matches an ASCII letter, so `ſ` does not match `s`.
- **Blank:** a line is blank when every code point is in `WS` or is Go `unicode.IsSpace`. Using this
  union, a chosen statement always passes `ValidateRevision`'s `TrimSpace` check.
- **`trim`:** removes leading and trailing `WS` (JavaScript `String.prototype.trim`).

**Input.**
- The file must be valid UTF-8 (`invalid-utf8`) and free of U+0000 (`nul-byte`).
- One leading U+FEFF is removed.
- `\r\n` and lone `\r` become `\n`, and lines are split on `\n`.
- Statements are the exact bytes of one line **without its terminator**, while the digest covers the
  file's exact bytes (§4.3).

**Fence mask** (`code-fence.ts`). Masked lines are invisible to every rule below.
- An **opener** matches `^WS*(`{3,}|~{3,})(DOT*)`, unless it is a backtick opener whose captured
  remainder contains a backtick.
- The fence **closes** at the first later line matching `^WS*(`{3,}|~{3,})WS*$` that uses the same
  marker character and is at least as long.
- Opener, body, and closer are masked. An unclosed fence masks to the end of the file.

**Sections** (`splitTopLevelSections`).
- Every unmasked line matching `^##WS+(DOT+)$` starts a section, which runs to the next such line or
  the end of the file. A `## ` line with one trailing space therefore starts nothing.
- The section's title is the capture, with a trailing `[ \t]+#+[ \t]*` removed, then `trim`med.
- The section is a delta section of kind ADDED, MODIFIED, REMOVED, or RENAMED exactly when its title
  is ASCII and its ASCII lower-case form equals `added requirements`, `modified requirements`,
  `removed requirements`, or `renamed requirements`. Repeated sections of one kind all apply.
- Anything else is ignored. For example, `## ADDED  Requirements` with two spaces is ignored, as
  OpenSpec's reader ignores it.

**ADDED / MODIFIED / REMOVED blocks** (`parseRequirementBlocksFromSection`, `parseRemovedNames`).
- **Header:** an unmasked line matching `^###WS*Requirement:WS*(DOT+)WS*$`, with `Requirement:`
  case-insensitive as above, opens an entry.
- **Name:** `normalizeRequirementName`, i.e. the capture with a trailing `[ \t]+#+[ \t]*` removed, then
  `trim`.
- **End of the block:** the next header, or an unmasked line matching `^##WS+`.
- **REMOVED bullet form:** in REMOVED sections, an unmasked line matching
  ``^WS*[-*+]WS*`?###WS*Requirement:WS*(DOT+?)`?WS*$`` (case-sensitive) is also an entry. Its
  statement is its name.
- **Statement — the one deliberate divergence.** OpenSpec has no single-line statement; this is
  Proofbound's. It is the first unmasked, non-blank line of the block that comes before any line
  starting with `#`. If there is none, it is the name.

**RENAMED pairs** (`parseRenamedPairs`).
- Lines are read in order. `FROM` matches
  ``^WS*[-*+]?WS*FROM:WS*`?###WS*Requirement:WS*(DOT+?)`?WS*$``, and `TO` is the same with `TO:`;
  both are case-sensitive.
- A `FROM` sets the pending name. A second `FROM` before a `TO` *displaces* the first, which becomes
  unpaired. A `TO` pairs with the pending `FROM`, or is unpaired if none is pending. A `FROM` still
  pending at the end of the section is unpaired.
- Each pair is one entry: the name is the `FROM` name, and the statement is the `TO` line's exact bytes.
- **Any unpaired line makes the record unmappable** (`unpaired-rename`), as OpenSpec refuses the delta.

**Spec Kit** (no upstream parser, so these rules are Proofbound's):
- **Obligation lines** match `^- \*\*((?:FR|SC|NFR)-[0-9]+)\*\*:? (.+)$`, giving id `O-<id>` and the
  capture's exact bytes as the statement.
  - An id that does not match the canonical `^O-[A-Z0-9][A-Z0-9-]{0,31}$` makes the record
    unmappable (`obligation-id-invalid`); for example, an id with 30 digits.
  - A blank statement does too (`statement-blank`).
- **Status** comes from the first line matching `requirements.status.line`, then `map`. With no status
  line, it is `absent`. Anything else is `status-unmapped`.

**Obligation ids (OpenSpec).** `O-<SEC>-<np>`, with `SEC` ∈ {`ADD`, `MOD`, `REM`, `REN`}. `np` is
derived from the name:
1. upper-case ASCII; every byte outside `[A-Z0-9]` → `-`; runs collapsed; trimmed;
2. if empty, the record is unmappable (`empty-name-part`);
3. if longer than 28 bytes, the first 21 bytes, re-trimmed, + `-` + the first 6 upper-case hex
   characters of `SHA-256(name)`.

Never positional. Two entries with one id make the record unmappable (`obligation-collision`).

**Payload values.**
- Every obligation's `state` is `active`, including a REMOVED entry: the change's declared promise to
  remove is an active obligation of that change.
- `authorized_by` and `supersedes` are `[]`.
- Record status: active OpenSpec change → `proposed`; archived → `active`; Spec Kit → as above.
- Owner: the first match in `owners`; none → `owner-uncovered`.

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
  - `status`: active OpenSpec change, or Spec Kit feature with its sibling → `changes.status`
    (declared, §2.1); archived OpenSpec change → `accepted` (the tenant archived it: OpenSpec's own
    acceptance act).
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

## 7. Fail-closed per state, never per history — a derived, closed list of data causes

A batch-fatal rule over full history is a permanent wedge, because past commits cannot be edited.

**Units:** a requirement record, a change intent at a state, or a whole state (the profile).

**How the list is derived (round-5 F1).** Every check that a tenant byte can reach between the tree
and a projected row is listed below with its cause. Each check is one of:
- §2.1 profile rules;
- §4.2 parsing rules;
- `ValidateRevision` (`intent.go`);
- `core.Event.Validate`;
- the ledger's `JSON` column;
- the projection's `TEXT` columns.

Checks that only provider-built structure can fail are excluded and fail the sync, because no tenant
byte controls them. Those are: canonical JSON, sorting, the source/kind registry, the envelope
binding, reference shape, and the record-id regex (the §4.3 slug always yields `[a-z0-9-]`).

| Check that tenant bytes reach | Cause |
|---|---|
| any §2.1 violation (including a blank owner or sponsor, U+0000, or non-UTF-8 in the profile) | `profile-invalid` |
| a source file or path under an SDD root that is not valid UTF-8 | `invalid-utf8` |
| U+0000 in a file a unit parses (it would pass the ledger's `JSON` column and fail every projection `TEXT` insert) | `nul-byte` |
| a symlink (mode `120000`) or gitlink (`160000`) where a unit reads a file or directory | `non-regular-file` |
| `owners` / `sponsors` match nothing | `owner-uncovered`, `sponsor-uncovered` |
| a Spec Kit status outside `map` with no `absent` | `status-unmapped` |
| `validateObligations`: id fails `obligationRE` (Spec Kit) | `obligation-id-invalid` |
| `validateObligations`: blank statement (Spec Kit) | `statement-blank` |
| `validateObligations`: duplicate id | `obligation-collision` |
| an empty name-part (§4.2) / a slug shorter than 2 (§4.3) | `empty-name-part`, `slug-too-short` |
| an unpaired RENAMED line | `unpaired-rename` |
| two keys minting one record id at one state | `record-id-collision` |
| a name both active and archived, or archived twice | `name-reuse` |
| `core.Event.Validate`: canonical payload over `core.MaxPayloadBytes` | `payload-too-large` |
| a change intent whose target record is skipped | `cascade` |

- Each unit with a listed cause contributes **no revision** and **one `Skip`**. Everything else at every
  commit is still emitted.
- **Everything else fails `Revisions`** (SDD-INV-49):
  - a candidate that fails `ValidateRevision` for an unlisted reason;
  - a reader or ref failure, `ErrShallow`, or cancellation.

  None of these can be caused by tenant bytes in history. Each is fixed by fixing the provider or the
  clone.
- The provider checks the data causes, builds the candidate, then calls `ValidateRevision`. An error at
  that point is returned, never skipped.
- **Batch-safe through projection.** Every returned revision passes `ValidateRevision`, which after
  §12 also enforces the size bound and rejects U+0000. The batch has no duplicate `(source, id, digest)`
  and no dangling same-source reference. So `Sync` validates, orders, and appends it (SDD-INV-47), and
  a rebuild projects it (SDD-INV-52).
- **Failure moves to the claim.** Only a claim on an unmappable state fails, at `Resolve`.
- **Skips are derived, not ledgered.** They are recomputed from the refs on every sync and carried in
  the connector's `Result`. The CLI prints them and exits 0. A skip can only withhold trust.
- **The family's `ErrUnmappable` STOP** is amended for `sdd` (§12): it marks one unit and is reported as
  a `Skip`. The foreign-fixture STOP for `records` / `specdir` is unchanged.

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
53. **SDD-INV-53 — Every payload string is valid UTF-8 without U+0000, and every statement is non-blank; a violation skips its unit.**
54. **SDD-INV-54 — Only regular files are read under SDD roots; a symlink or gitlink a unit reads makes that unit unmappable.**
55. **SDD-INV-55 — A path outside the SDD roots, whatever its bytes, never affects the sdd provider.**
56. **SDD-INV-56 — `sdd` runs only when selected by name; `sync intent all` neither constructs nor walks it.**
57. **SDD-INV-57 — The §4.2 parser agrees with OpenSpec's reader on the pinned corpus.**
58. **SDD-INV-58 — A Spec Kit id outside the canonical obligation grammar makes that record unmappable.**

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
| SDD-INV-35 | Every data cause placed (a) at a non-tip commit fixed later, (b) at a non-HEAD ref tip (`branches` step 14 `stale`), and (c) **at HEAD** (`head-on-leftover`) leaves `Revisions` without error and every other unit emitted | sdd_test.go::TestHistoricalFailClosedIsRecoverable | kills: batch-fatal; tip-fatal; **HEAD-fatal** |
| SDD-INV-39 | Each `fenced/` case yields exactly its `openspec_mask` list in `expected-obligations.json`; the column-0 and three precision cases each differ from the naive-toggle or unmasked list; `document-renames` is the one neutral case | sdd_test.go::TestCodeFencesAreMasked | kills: unmasked; a naive toggle (each precision case); survives: the neutral |
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
| SDD-INV-52 | `openspec-history`, plus a synthesized history carrying one unit per §7 cause (NUL, oversize, invalid UTF-8, blank statement, long id, symlink, gitlink), syncs and then **rebuilds projections** with zero errors; requirement and change-intent rows exist for the clean units | sdd_projection_test.go::TestSDDEventsProject | kills: an unregistered route; a NUL statement reaching a `TEXT` column |
| SDD-INV-53 | One synthesized unit per string route (OpenSpec statement, name fallback, RENAMED `TO`, Spec Kit statement, profile owner) carrying U+0000 or a blank statement is skipped with `nul-byte` / `statement-blank` / `profile-invalid` | sdd_test.go::TestPayloadStringsAreStoreSafe | kills: a NUL-tolerant provider |
| SDD-INV-54 | A symlinked delta, a symlinked change directory, and a gitlink under the root each skip their unit with `non-regular-file`; a regular sibling is emitted | sdd_test.go::TestNonRegularFilesAreUnits | kills: following a symlink; a reader error on a gitlink |
| SDD-INV-55 | A repository whose history adds `caf\xe9.md` outside the roots syncs `sdd` with no error and no skip; the same name inside a change directory skips that change with `invalid-utf8` | sdd_test.go::TestForeignPathsAreIgnored | kills: a reader that lists per-commit file diffs |
| SDD-INV-56 | `sync intent all` and `proofbound verify` on a shallow clone with no profile behave exactly as before M1; only `sync intent sdd` constructs the provider | cli_test.go::TestIntentAllExcludesSDD | kills: adding `sdd` to `all` |
| SDD-INV-57 | Round 5's divergence table (REMOVED bullet; `## ADDED  Requirements`; NBSP after `##`; a one-space `## ` line; `Requirementſ`; interleaved RENAMED) plus `expected-obligations.json` match OpenSpec's reader output, as pinned in testdata | sdd_test.go::TestParserMatchesOpenSpecReader | kills: `DELTA_HEADER` as section rule; a Go `\s`/`(?i)` translation |
| SDD-INV-58 | `- **FR-<30 digits>**: x` skips its record with `obligation-id-invalid`; `FR-001` beside it in another feature is emitted | sdd_test.go::TestSpecKitIDsFitTheGrammar | kills: emitting an id `ValidateRevision` rejects (a sync failure) |

## 10. Decisions this draft settles or withdraws

- **A, B, C:** unchanged from draft 3 (warranty across archive withdrawn; main-spec lifecycle
  withdrawn from R2; Spec Kit lifecycle answered by content addressing).
- **D: fail-closed scope.** Per unit, never per history, never per ref tip, never per HEAD (§0: four
  branch tips today, and `main` itself for 9 commits). SDD-INV-35(a-c) kill all three alternatives.
- **E: walk.** The git connector's walk under its read contract, without its per-commit file lists
  (§3).
- **F: data causes vs defects.** A closed list, derived from every check tenant bytes reach (§7).
- **G (owner, 2026-11-17): `all` excludes `sdd`.** Including it would make `sync all`, `verify`, and
  `delivery-enforce` walk full history in every repository, and would inherit `ErrShallow` in shallow
  CI checkouts that work today. Revisit at M2, once an OpenSpec repository has synced end to end.
- **H: the statement is Proofbound's,** the one labelled divergence from OpenSpec's reader (§4.2).

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
| `connector/intent/intent.go:215` (`ValidateRevision`) | admit `intent.sdd`. **For every provider:** enforce `len(payload) ≤ core.MaxPayloadBytes`, and reject U+0000 in any payload string. This closes the latent partial-append and projection-wedge paths for `records` / `specdir` too |
| `connector/intent` (`Result`) | additive `Skip { Commit, Unit, Cause string; Err error }`, optional `interface{ Skipped() []Skip }`, and `ProviderResult.Skipped` (omitted when empty) |
| `projections/projections.go:296` (`eventRoute`) | route `intent.sdd` intent kinds |
| `gates/gates.go:304` (`intent-reference-integrity`) | include `intent.sdd` |
| `cli/cli.go:57-60` (command enum), `:92`-area parse, `:513-514` (dispatch), `:37` (usage) | add `sync intent sdd` |
| `cli/cli.go:366-382` (`syncIntentOnStore`) | construct the `sdd` provider **only when the selection is `sdd`** (SDD-INV-56) |
| `connector/git/gitcmd` | factor the shared listing out of `Commits`; add the `sdd` Reader on `Repo` (§3). The git connector's behaviour is unchanged |

**Unchanged by M1, on purpose:** `cli.go:596` (`sync all`), `:835` and `:854` (`verify`),
`scripts/delivery-enforce.sh:38`, and `docs/releases/r1/seed-demo-app.sh:94`. They call
`sync intent all`, which still means `records` + `specdir`.

**Scripts and generated artifacts, M1:**

| Site | Change |
|---|---|
| `scripts/p6-route-matrix.sh:19-24` | add the three `intent.sdd` routes, then regenerate the `@generated` route matrix |
| `scripts/p6-census.sh:233-235` | add the `intent.sdd` rows |

**Intent SPEC text, M1:**
- **§1:** point at `sdd/SPEC.md`.
- **§2:** the `Skip` report.
- **§3:** add "payload strings carry no U+0000".
- **§4:** `sdd`'s manifest digest and parent archive reads.
- **§6 and INT-INV-11:** for `sdd`, `ErrUnmappable` marks a unit and is reported as a `Skip`; the
  foreign-fixture STOP applies to `records` and `specdir`.
- **INT-INV-4:** its test gains an oversize-payload row and a U+0000 row, each failing before any
  append.

**Owned by M3, not M1** (resolving tenant claims):
- `connector/git/connector.go:198`, the provider allowlist;
- `cli/cli.go:247`, the `commitIntentResolver.Resolve` provider switch (draft 5 wrongly listed it as M1);
- `projections/projections.go:723`, the `IntentRef` provider check;
- `scripts/intent-delivery-applicability.sh:93`, the trailer id regex.

M1 must not touch these.

**M1 acceptance** includes SDD-INV-52: a sync containing `intent.sdd` events, followed by a projection
rebuild.

## 13. Test derivations and fixtures

- **The replay helper** (`fixture_test.go`) is committed with this draft, on the owner's authorization
  2026-11-17. It provides PARENTS/REF commit graphs, `requireProfiles`, and the self-tests below.
- **`testdata/branches/`** (14 steps): a root with an archive; an edit / claim / revert on `feature`,
  merged `--no-ff`; a side-branch archive and merge; an unmerged claim on `wip`; the leftover state with
  a claim, and its removal; a `stale` ref tip on the leftover state.
- **`testdata/head-on-leftover/`** (2 steps): HEAD itself ends on the dual-presence state
  (SDD-INV-35(c)).
- **`testdata/fenced/`** (5 cases):
  - `improve-onboarding`: column-0 `## TL;DR` inside a fence;
  - `fence-marker`: a `~~~` line inside a backtick fence;
  - `fence-length`: a shorter closer;
  - `fence-info-backtick`: a backtick in a backtick opener's info string;
  - `document-renames`: the neutral indented example.

  `expected-obligations.json` pins each case's obligations under OpenSpec's mask, a naive toggle, and
  no mask. The lists were computed by the probe's §4.2 parser, and every non-neutral case differs.
- **`testdata/expected-skips.json`** pins each fixture's exact Skip set (SDD-INV-46).
- **Mechanised now (round-5 F6):** `TestPinnedExpectationsMatchFixtures` checks that:
  - every pinned skip names an existing step and unit;
  - every fenced case keeps its column-0 MODIFIED header and a header for each pinned obligation;
  - each case's `discriminates` flag matches its lists, and exactly one case is neutral;
  - HEAD of `head-on-leftover` is on the dual-presence state.

  Calibrated: renaming the MODIFIED header fails the test; a bogus skip unit fails it; a neutral edit
  passes.
- **Byte-level hostile cases** (NUL, oversize, invalid UTF-8, blank statements, long ids, symlinks,
  gitlinks, a non-UTF-8 path) are **synthesized by the M1 tests**, not committed (SDD-INV-52..55).
  Committed binary-hostile bytes would trip repository hygiene for no gain.
- **Calibration:** every §9 row names its kill. M1 runs those mutants through
  `scripts/mutant-apply.sh` before its counts are believed.
- **Probe honesty:** `derive/history_probe.py` implements §4.2, compares statements as exact bytes, walks
  the provider's walk, and prints every §0 count.

## 14. Non-goals and boundaries owned elsewhere

- Main-spec ingestion, folding, tombstones, and carry-over (deferred with the main-spec lifecycle).
- Authoring or writing tenant SDD files; governance → business decisions; YAML; heuristic extraction;
  Spec Kit per-task intents; API / DB providers; more than one profile per repository; ledgering skips.
- **M3 entry conditions:**
  1. the `sdd:` trailer token and its decision record;
  2. anti-spoofing against a native name the git connector reads from the commit tree itself,
     including the profile's `kit`;
  3. **per-commit handling of an unresolvable claim.** Today the git connector fails the sync closed.
     A claim at one of the 23 dual-presence commits would wedge every git sync. M3 uses §6's sentinels
     to report it per commit, or proves the trailer gate refuses it first;
  4. the M3 sites in §12;
  5. the readiness gate on `sdd` chains.
- **Recorded, not owned here:** the git connector itself (`gitcmd.Commits` → `files`) fails on any
  non-UTF-8 path anywhere in history. That is a pre-existing git-connector wedge, independent of `sdd`,
  and `sdd` does not inherit it (§3). It is raised to the owner as its own item.
- **M4:** change-drift reporting (only if re-proposed with a ledger-evaluable definition) and the status
  export.
