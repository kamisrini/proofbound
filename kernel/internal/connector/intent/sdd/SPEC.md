# `sdd` spec-driven-development provider contract — DRAFT 10 — the M1 contract (change-scoped, history-native, per-state; OpenSpec)

Status: **DRAFT 10 — the M1 contract.** M0 is closed (owner, 2026-11-17). `BR-sdd-warranty-i9miv7`
revision 0003 is written.

Draft 10 is draft 9 plus the six M1 entry conditions from the targeted non-author re-check
(`docs/verification/verdicts/r2-m0-recheck-draft9-20261117.md`, F1-F6), answered in the table below.
It is copied **verbatim** into `kernel/internal/connector/intent/sdd/SPEC.md` in the M1 commit that
adds the code and tests it cites. It cannot land earlier: `invariant-lint` fails `make check` when a
kernel SPEC cites a test that does not exist. M1's foundational-tier non-author review checks these
six answers.

**Amended in M1 (2026-11-18), answering the M1 foundational review round 1**
(`docs/verification/verdicts/r2-m1-review-round1-20261117.md`):
- **F1:** §4.1's layout rule no longer treats a regular file at `specs` or directly under `specs/`
  as a cause. OpenSpec's reader ignores it, and real history has one (`no-changes.md`, §0).
- **F4:** §0 names git walk options in words, not as backticked literals.
- **F10:** §2's interface lock lists every exported name; §12 records how the census and connector
  universe were really registered.
- **F2 and F3:** test-only (SDD-INV-59, SDD-INV-52).

**Self-contained:** every definition the contract uses is in this file, and earlier drafts are cited
only as history. This draft becomes `kernel/internal/connector/intent/sdd/SPEC.md` in the same M1
commit as the code and tests it cites, together with every amendment listed in §12.

Scope (owner decision 2026-11-16, `VD-sdd-change-scoped-r2-k7q3wz`, unchanged): **R2 warrants OpenSpec
changes and Spec Kit features — the commits that implement a change, against the requirements that
change declares — read from the repository's history.** The main-spec lifecycle is not in R2.
Requirement: `BR-sdd-warranty-i9miv7` (revision proposed in §11).

**Owner decisions in force (2026-11-17).**
- `sync intent all` does **not** include `sdd`; the provider runs only when selected by name (§12,
  SDD-INV-56).
- The replay helper is committed (`4c8c838`), as an M1 precondition.
- **`VD-sdd-m1-openspec-first-2026-11-17`:**
  - **M1 maps OpenSpec only.** Spec Kit moves to M2, still within R2's subset. Its rules are parked,
    non-normative, in Appendix A.
  - **Profile paths are literal,** and every git path lookup uses literal pathspecs (§2.1, §3).

## 0. Evidence first

Measured on a full clone of `Fission-AI/OpenSpec` at `f085c850`. Every row is printed by
`testdata/derive/history_probe.py`. Since draft 6 the probe walks **the provider's own walk**
(unsimplified `rev-list --all`, date order) and parses with §4.2's rules. The two exceptions are
labelled.

| Observation (real history) | Count | Consequence |
|---|---|---|
| Commits on the walk / merges / roots | 1624 / 120 / 1 | The provider walks this set (§3) |
| Distinct SDD states | 338 | Mapping cost tracks states (§3) |
| Commits touching `openspec/changes` on the HEAD path-limited walk / on all refs with full history (labelled: other walks) | 256 / 451 | Path-limited walks miss states (§3) |
| Changes observed active / archived | 135 / 108 | — |
| **Commits carrying a change both active and archived** | **23** (1 name) | Unmappable for that change at that state only (§7) |
| …of which are ref tips | **4 remote branches**, 0 tags; `main` itself sat on this state for 13 consecutive first-parent commits, 8 of them non-merge (probe-printed; round 7's "9" was one too many) | Neither tip-fatal nor HEAD-fatal is acceptable (§10 D, SDD-INV-35) |
| Names archived twice at one commit | 0 | Still unmappable if it occurs (§5) |
| Changes archived without ever being observed active (squash-merge) | **12** / 108 (draft 5's 16 came from the path-limited walk) | A name resolves when archived **by** the commit (§6) |
| Changes with zero requirement entries at some commit | 38 | Absence, never an error (§5) |
| Requirement removals / statement changes between consecutive observations of a change (date order across all refs, so branch interleaving inflates both) | 83 (17 changes) / 39 (9 changes) | Content-addressed records (§4) |
| Requirements with no statement line (statement = name) | 4 | §4.2 fallback |
| Regular files directly under a change's `specs/` (`archive/2026-01-30-opencode-command-references/specs/no-changes.md`, "no spec files are needed") | 1 | Ignored, as OpenSpec ignores it: never a unit or cause (§4.1) |
| Regular files AT a change's `specs` (probe-printed; entries typed by mode, since `ls-tree` lists a symlink as type `blob` too, M1 re-check F4) | 0 | Ignored, as above (§4.1) |
| Symlinks or gitlinks at or directly under a change's `specs` (probe-printed) | 0 | `non-regular-file` record units (§4.1); pinned only by synthesized tests |
| Commits with a **regular file directly under `openspec/changes`** (`IMPLEMENTATION_ORDER.md`, `skills-only-migration.napkin.md`) | **1393** / 1624 | Never a change (§5, SDD-INV-66) |
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

**How each M1 entry condition (re-check of draft 9) is answered.** Answers to rounds 3-8 are in
drafts 4-9, §0.

| Finding | Answer |
|---|---|
| F1 symlink/gitlink inside a change directory has two outcomes | **Layout types are fixed (§4.1).** On a delta path, `<change>/specs` and `<change>/specs/<cap>` must be trees, and `spec.md` must be a regular file. A symlink or gitlink at `specs` or `specs/<cap>`, or a symlink, gitlink or tree at `spec.md`, is a **record unit** skipped with `non-regular-file`, so the change is `cascade`. A **regular file** at `specs` or directly under `specs/` is ignored, never a cause (M1 review F1; §4.1 is the normative rule). **Every other entry** in a change directory (`design.md` as a symlink, a gitlink anywhere else) is never a cause. It enters manifests as `"<mode> <oid>  <path>"` instead of a content hash (§4.3), so the digest is always computable. SDD-INV-54's statement is unchanged and now has one reading: a unit "reads" only what it parses |
| F2 parent archive check reads names only | `List` returns typed entries. §6 line 5 counts only parent archive entries that are change entries (tree, symlink, or gitlink) naming `n`, the §5 definition |
| F3 tree at a delta path | Covered by F1's layout rule: a tree at `…/spec.md` is a `non-regular-file` record unit, so its change is `cascade` |
| F4 `profile` digest reads roots from an invalid profile | For `profile-invalid`, the unit digest is the profile entry line alone. For a state-level `non-regular-file`, it is the profile entry line plus the offending root or archive entry line. One long-lived invalid profile is therefore one Skip |
| F5 INV-65's "caching Resolve by key" kill cannot fire | The kill is removed from INV-65's row. It is SDD-INV-59's (`branches` 08) |
| F6 drift | §7's `archive-name-invalid` row and §2's `Skip.Unit` comment now match §6. SDD-INV-62 is retired and re-issued as INV-67, without the pathspec claim. `fixture_test.go`'s message is fixed in the M1 commit |

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

**M1 implements the OpenSpec kit only** (`VD-sdd-m1-openspec-first-2026-11-17`). Every rule below is
OpenSpec's unless it is marked otherwise. The Spec Kit material is in Appendix A, which is not part of
M1's contract.

**Not wired into the git connector in M1.** M1 registers `intent.sdd` with the intent connector,
core, the CLI, projections, and gates (§12). The `sdd:` trailer token and everything that resolves a
tenant claim through the provider belong to M3 (§14).

## 2. Interface lock

```go
type Entry struct { Name string; Mode string; OID string } // Mode: "040000", "100644", "100755", "120000", "160000"
type Artifact struct { Path string; Bytes []byte; Regular bool; Mode string; OID string } // Regular: git mode 100644/100755; symlinks and gitlinks carry Regular=false and no Bytes (§7 non-regular-file)
type Reader interface {
    // Every commit the git connector ingests (§3), oldest first, under its read contract.
    Commits(ctx context.Context) ([]string, error)
    // The parents of commit, in order (empty for a root commit).
    Parents(ctx context.Context, commit string) ([]string, error)
    // The state key of commit (§3): "<mode> <oid>" of the profile path's tree entry and of each SDD
    // root's and changes.archive's entry, in a fixed order, "-" for an absent entry; "" when the
    // profile entry is absent. Equal keys ⇒ byte-identical SDD inputs AND equal object types.
    StateKey(ctx context.Context, commit string) (string, error)
    // Every tree entry at commit under the profile's SDD roots, plus the profile (paths as raw bytes;
    // the provider applies §7's invalid-utf8 and non-regular-file causes).
    Read(ctx context.Context, commit string) ([]Artifact, error)
    // The entries directly under dir at commit, each with its git mode (the parent archive check, §6).
    List(ctx context.Context, commit, dir string) ([]Entry, error)
}
type Provider struct { /* unexported */ }

var (
    ErrUnknownChange   = errors.New("sdd: no such change at this commit")                 // §6
    ErrArchivedEarlier = errors.New("sdd: change was archived before this commit")         // §6
    ErrNoChangeIntent  = errors.New("sdd: change exists at this commit but has no intent") // §6
)

// UnmappableError carries one §7 cause; it wraps intent.ErrUnmappable.
type UnmappableError struct{ Cause string }

const ProfilePath = ".proofbound/sdd-profile.json" // §2.1
const Kit = "openspec"                              // the only kit M1 maps; suffix hashes it (§4.3)

// GitReader is the production Reader over the git connector's gitcmd.Repo (§3).
type GitReader struct{ /* unexported */ }
func NewGitReader(repo *gitcmd.Repo) (*GitReader, error)

func New(reader Reader, observedAt time.Time) (*Provider, error)
func (p *Provider) Name() string                                             // "sdd"
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) // whole history (§3)
func (p *Provider) Skipped() []intent.Skip                                   // after Revisions (§7)
func (p *Provider) Resolve(ctx context.Context, commit, name string) (intent.Revision, error) // §6
func (p *Provider) StateRevision(ctx context.Context, commit, name string) (intent.Revision, error) // SDD-INV-68 (R2 M4)
```

`intent.Skip` is `struct { Commit, Unit, Cause string; Err error }` (intent SPEC §2 amendment, §12):
- `Commit` is the first commit, in `Commits` order, that carries the skipped state.
- `Unit` is `requirement <path>`, `change <name>`, `archive <entry>` (a tree, symlink, or gitlink under
  the archive that names no in-domain change, §6), or
  `profile` (a state-level cause).
- `Cause` is one value from §7's closed list.
- `Err` wraps `intent.ErrUnmappable`.

`Revisions` returns an error for anything that is not on that list (§7).

### 2.1 Profile — `proofbound.sdd-profile.v1`

Path `.proofbound/sdd-profile.json`. A commit without it contributes nothing and has the empty state
key. The profile is strict JSON: UTF-8, objects with exactly the fields below, no unknown, missing,
duplicate, or `null` fields, and no unknown enum values.

```json
{ "schema": "proofbound.sdd-profile.v1", "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": "team-x" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "team-x" } ] }
```

Field rules:
- **`kit`** must be `openspec` in M1, and `changes.mode` must be `delta-sections`. A profile with
  `kit: speckit` is `profile-invalid` until M2 (Appendix A).
- **`changes.status`** ∈ {`proposed`, `accepted`} is the status of every active change. OpenSpec records
  no approval, so the tenant declares what an active change means in its workflow; Proofbound never
  infers it.
- **Literal path grammar** (round-7 F3; `VD-sdd-m1-openspec-first-2026-11-17`).
  - A **path** is one or more segments joined by single `/`. Each segment matches `^[A-Za-z0-9._-]+$`
    and is neither `.` nor `..`. Paths have no leading or trailing `/` and no empty segment.
  - `changes.root` and `changes.archive` are paths, and `changes.archive` must be a **direct child** of
    `changes.root` (round-8 F3).
  - `requirements.paths` must be **exactly**
    `["<root>/*/specs/*/spec.md", "<archive>/*/specs/*/spec.md"]`, in that order (round-8 F4). Kit
    rules are code: the profile chooses the root names, not the delta layout.
  - Anything else, including a leading `:`, `..`, a leading `/`, an empty prefix, `./`, `//`, or a
    trailing `/`, is `profile-invalid`.
- **`owners` / `sponsors` patterns** use `*` (one segment) and `**` (any number of segments) over the same
  literal segment grammar. They are matched by the provider against path bytes and are never passed to
  git.
- **SDD roots.** The roots are `changes.root` and `changes.archive`. Every file a unit reads lies under
  them.
- **Kit rules are code, not profile.** The section, requirement, and rename rules (§4.2) are fixed by
  the contract.
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
  - A non-UTF-8 path **outside** the roots therefore never reaches `sdd`. Inside the roots, the
    `invalid-utf8` scope is §7's: only for a file a unit parses, or a name or path a payload carries.
  - The git connector's own failure on such paths is pre-existing git-connector behaviour, recorded in
    §14 and not inherited.
- **Literal pathspecs (round-7 F3, round-8 F5).** What keeps tenant strings from being read as
  pathspec magic is **the literal grammar** (§2.1): an empty or malformed path is `profile-invalid`
  before it reaches git.
  - Every path-taking git command the `sdd` reader runs also carries `--literal-pathspecs`. This is
    **defence in depth with no observable effect** on grammar-valid paths: no such path starts with
    `:`, and `ls-tree` and `cat-file` do not glob.
  - It is required, but it is not separately testable, and no invariant claims it is.
- **State-level object types (round-6 F3).** `StateKey`, `Read`, and `List` look paths up in the
  commit's tree, the equivalent of `git --literal-pathspecs ls-tree <c> -- <path>`. They never dereference `<c>:<path>`, so
  an entry of an unexpected type is a value, not a reader error.
  - **Profile path:**
    - absent → no profile (the state contributes nothing);
    - a regular blob → parsed;
    - a tree, symlink, or gitlink → `profile-invalid`.
  - **An SDD root, or `changes.archive`:**
    - absent → empty;
    - a tree → read;
    - a blob, symlink, or gitlink → the whole state is unmappable (`non-regular-file`,
      `Skip.Unit = "profile"`).
  - **The state key** (round-7 F5; SDD-INV-65) is the newline-joined list of `<mode> <oid>` for the
    profile path's entry, then `changes.root`, then `changes.archive`. An absent entry is `-`. With no profile entry, the key is `""`. A symlink or gitlink
    at the profile path therefore never shares a key with a regular profile.
- **No path limiting and no history simplification.** Every listed commit is visited.
- **Per distinct state.** The provider maps each distinct `StateKey` once and returns the union,
  de-duplicated by `(kind, record id, artifact digest)`.
- **SDD-INV-31.** For every commit `c` that `Commits` lists and every name `n`, a successful
  `Resolve(c, n)` is a member of `Revisions()`. Both are the same function of `c`'s state and its
  parents' archive listings, over the same commit set. No unit's mappability depends on another state.
- **Ordering of syncs** (INT-INV-9) and **idempotence** (INT-INV-5): unchanged.

## 4. Requirement records (content-addressed)

### 4.1 What is a requirement record

- **Layout types (re-check F1, F3).** On a delta path `<change>/specs/<cap>/spec.md`, `<change>/specs`
  and `<change>/specs/<cap>` must be trees, and `spec.md` must be a regular file (`100644` or
  `100755`).
  - If a symlink or gitlink occupies `specs` or `specs/<cap>`, or a symlink, gitlink, or tree
    occupies `spec.md`, the unit is **one record unit** at that position (`requirement <path of the
    position>`), skipped with `non-regular-file`.
  - A **regular file** at `specs`, or directly under `specs/` (OpenSpec's real `no-changes.md` note,
    §0), is ignored, as OpenSpec's reader ignores it: never a record unit and never a cause.
  - Every other entry in a change directory, of any type, is never a record unit and never a cause.
- **Record unit (round-8 F1).** Every delta path matched by `requirements.paths` (active
  `<root>/<name>/specs/<capability>/spec.md`, or archived
  `<archive>/YYYY-MM-DD-<name>/specs/<capability>/spec.md`) is a **record unit** when either:
  - it is **skipped before parsing**: `non-regular-file`, `invalid-utf8`, or `nul-byte`; or
  - it **parses** to at least one entry, **or** to at least one unpaired RENAMED line.
- A record unit is either emitted as a record or skipped with its cause.
- A matched path that parses to zero entries and no unpaired line is **not** a record unit. It yields
  no record and no Skip.

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
  statement is its name, and it **ends any open block**, like a header (round-6 F7).
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

**Obligation ids (OpenSpec).** `O-<SEC>-<np>`, with `SEC` ∈ {`ADD`, `MOD`, `REM`, `REN`}. `np` is
derived from the name:
1. **byte-wise, ASCII only** (round-6 F4): each byte `a`-`z` → `A`-`Z`; each byte in `A`-`Z0-9` is
   kept; **every other byte, including every byte of a multi-byte UTF-8 sequence, → `-`**; runs
   collapsed; trimmed. There is no Unicode case mapping: `expoſe` → `EXPO-E`, not `EXPOSE`. In Go,
   iterate bytes, never `strings.ToUpper`;
2. if empty, the record is unmappable (`empty-name-part`);
3. if longer than 28 bytes, the first 21 bytes, re-trimmed, + `-` + the first 6 upper-case hex
   characters of `SHA-256(name)`.

Never positional. Two entries with one id make the record unmappable (`obligation-collision`).

**Payload values.**
- Every obligation's `state` is `active`, including a REMOVED entry: the change's declared promise to
  remove is an active obligation of that change.
- `authorized_by` and `supersedes` are `[]`.
- Record status: active change → `proposed`; archived → `active`.
- Owner: the first match in `owners`; none → `owner-uncovered`.

### 4.3 Identity and digest

- **`slug(s)`:** byte-wise ASCII only, as `np` above: `A`-`Z` → `a`-`z`; `a`-`z0-9` kept; every other
  byte, including every byte of a multi-byte sequence, → `-`; runs collapsed; trimmed; truncated to 40
  bytes and re-trimmed. Shorter than 2 bytes → unmappable (`slug-too-short`).
- **`suffix(kind, key)`:** the first 6 bytes of `SHA-256(kit + "\x00" + kind + "\x00" + key)`, each
  byte `b` mapped to `"abcdefghijklmnopqrstuvwxyz0123456789"[b % 36]`. `kind` is `requirement` or
  `change-intent`.
- **Manifest:** one line per input entry, sorted by path bytes; `artifact_sha256 = SHA-256(manifest)`.
  - A regular file contributes `"<sha256-hex-of-exact-bytes>  <path>\n"`.
  - A symlink or gitlink that a change directory contains outside the layout positions contributes
    `"<mode> <oid>  <path>\n"` (re-check F1). Its bytes are not read; its git identity is pinned.
- **Requirement record:**
  - Inputs: the source file and the profile. `artifact_path` is the source path.
  - Key: `<name>/<capability>`.
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

## 5. Change intents — one outcome function

A **change entry** for name `n` at a state is an entry that is a **tree, symlink, or gitlink** (never a
regular file, round-8 F2), and is either:
- directly under `changes.root`, named `n`, and not `changes.archive` itself; or
- directly under `changes.archive`, and naming `n` (§6).

**Regular files directly under either directory are ignored:** they are not changes, not units, and
never a cause (SDD-INV-66). Real history has such a file at 1393 of 1624 commits.

The `sdd` provider computes **`changeOutcome(state, n)`** for every **in-domain** name (§6) that has at
least one change entry. **Both `Revisions` / `Skipped` and `Resolve` use
this one function** (round-7 F1), so they cannot disagree.

The **first** rule that applies decides:
1. **State-level cause** (`profile-invalid`; `non-regular-file` at a root or the archive, §3) → the
   whole state is skipped once, as unit `profile`, and every `n` resolves to `ErrUnmappable:<that cause>`.
2. **`name-reuse`:** at this state, `n` has an active entry and an archive entry, or two or more
   archive entries. This includes a name archived twice by a single commit, and the commit that
   re-archives a reused name.
3. **Name causes:** `invalid-utf8` (the name), or `slug-too-short`.
4. **`non-regular-file`:** a change entry of `n`, whether active or archived, is not a tree.
5. **No record units:** none of `n`'s delta paths is a record unit (§4.1). This is the absent outcome
   (SDD-INV-28): no change intent and no `Skip`, and `Resolve` gives `ErrNoChangeIntent`.
6. **`cascade`:** a record unit of `n` at this state is skipped, for any reason, including before
   parsing.
7. **`sponsor-uncovered`**, or **`payload-too-large`** (the canonical change-intent payload).
8. Otherwise, **a change intent**.

Rules 2-4, 6, and 7 yield one `Skip` (unit `change <n>`). A requirement record's own causes (§4, §7)
are decided per record, independently of rule 2: a reused name's records are still emitted.

- **Id:** `CI-<slug(n)>-<suffix("change-intent", n)>`. At a state, the intent `implements` every
  obligation of every requirement record of `n` at that state.
- **Payload values:**
  - `status`: active → `changes.status` (declared, §2.1); archived → `accepted` (the tenant archived
    it, which is OpenSpec's own acceptance act).
  - `declared_sponsor`: the first match in `sponsors`.
  - `constrained_by` and `supersedes` are `[]` (`DEF-INTENT-CI-REVISION-001`).
  - `artifact_path` is the change directory. The manifest covers every file in it plus the profile.

## 6. Resolution

**Archive names (round-4 F9).** An entry directly under `changes.archive` names change `n` exactly when
its name matches `^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$` and the capture equals `n` byte for byte. There
are no globs. The pattern is matched with Go RE2 semantics, in which `.` does not match `\n` (M1 re-check F3):
an archive entry whose capture would contain a newline does not match, so it is `archive-name-invalid`,
while the same name as an active change (where no pattern applies) maps. M1 keeps this asymmetry: it
only ever yields a Skip, never a wrong record, and real history has no such name.
- A tree, symlink, or gitlink entry that does not match, or whose capture is outside the name domain
  (`.`, `..`, or the archive's own name, round-8 F3), names no change. It is one unit,
  `archive <entry>`, skipped with `archive-name-invalid`, and its delta files are not read. Real history
  has none. A regular file there is ignored (SDD-INV-66).

**Name domain (round-6 F1).** A change name is one path segment: non-empty, with no `/` and no U+0000,
not `.` or `..`, and not the final segment of `changes.archive`, which is a direct child of the root
(§2.1). Any other `n` is `ErrUnknownChange`.

**Outcomes (SDD-INV-59; `Resolve` is total over tenant states).** For any `c` that `Commits` lists and
any `n`, the first matching line decides:
1. `n` is outside the name domain → **`ErrUnknownChange`**.
2. `c` has no profile → **`ErrUnknownChange`** (nothing opted in). `c`'s state has a state-level cause
   → **`ErrUnmappable:<cause>`** (§5 rule 1).
3. No change entry names `n` at `c` → **`ErrUnknownChange`**.
4. `changeOutcome` is a `Skip` (§5 rules 2-4, 6, 7) → **`ErrUnmappable:<cause>`**. This includes
   `name-reuse`, so a name archived twice by `c`, or re-archived after reuse, is `name-reuse`, exactly
   as `Skipped()` reports it.
5. `n` has exactly one change entry, it is an archive entry, and **some parent** of `c` has an archive
   entry that is a **change entry** (a tree, symlink, or gitlink, §5; a regular file never counts,
   re-check F2) naming `n` → **`ErrArchivedEarlier`**.
   - **A merge** whose side branch archived `n` gets this outcome; the side-branch commit resolves.
6. `changeOutcome` is absent (§5 rule 5) → **`ErrNoChangeIntent`**: a proposal-only change, or a
   zero-entry archive. 38 of 135 real changes were in this state at some time.
7. Otherwise → **the change intent's revision.** The intent is either active at `c`, or archived by
   `c` with no parent carrying it. **A root commit** archives what it carries; roots are real roots,
   because replace refs are ignored and grafts and shallow clones are refused (§3).

An error that is none of these comes **only** from the `Reader` (I/O, refs, `ErrShallow`) or from
cancellation. No tenant state reaches it.

M3 uses these outcomes to report a bad claim per commit instead of failing the sync (§14).

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
| a file a unit **parses** (a delta), or a name or path a payload **carries** (a record's `artifact_path`, the change directory name), that is not valid UTF-8. This is the only scope: other files and paths are never this cause | `invalid-utf8` |
| U+0000 in a file a unit parses: it would pass the ledger's `JSON` column and fail every projection `TEXT` insert. Paths cannot hold U+0000 | `nul-byte` |
| a non-tree change entry (§5 rule 4); a symlink, gitlink or tree at a delta-path layout position as §4.1 lists them (a regular file at `specs` or directly under it is ignored, not this cause); at state level, a non-tree root or archive (§3). Nothing else is this cause | `non-regular-file` |
| a tree, symlink, or gitlink directly under the archive with no date prefix, or whose capture is out of domain (§6). A regular file there is ignored | `archive-name-invalid` |
| `owners` / `sponsors` match nothing | `owner-uncovered`, `sponsor-uncovered` |
| *(M2, Spec Kit; unreachable in M1)* a status outside `map` / an id outside `obligationRE` / a blank statement | `status-unmapped` / `obligation-id-invalid` / `statement-blank` |
| `validateObligations`: duplicate id | `obligation-collision` |
| an empty name-part (§4.2) / a slug shorter than 2 (§4.3) | `empty-name-part`, `slug-too-short` |
| an unpaired RENAMED line | `unpaired-rename` |
| two keys minting one record id at one state | `record-id-collision` |
| a name both active and archived, or archived twice | `name-reuse` |
| `core.Event.Validate`: canonical payload over `core.MaxPayloadBytes` | `payload-too-large` |
| a change intent whose target record is skipped | `cascade` |

- **Other files** in a change directory, such as images or binaries, are neither parsed nor
  carried in a payload. They enter the unit's manifest as raw bytes and raw path bytes, and are never
  a cause (round-6 F3).
- Each unit with a listed cause contributes **no revision** and **one `Skip`**.
  - Skips are **de-duplicated by `(Unit, Cause, unit digest)`**. The **unit digest** (round-8 F6) is
    SHA-256 over sorted lines `"<mode> <oid> <path>\n"`, covering the profile path's entry plus:
    - **requirement:** the delta path's entry;
    - **change:** every entry, recursively, under every change entry of `n` (active and archived);
      a non-tree entry contributes its own line;
    - **archive `<entry>`:** every entry, recursively, under that entry;
    - **profile** (state-level), with `profile-invalid`: the profile entry line alone, so roots are
      never read from an invalid profile, and one long-lived invalid profile is one Skip;
    - **profile** (state-level), with `non-regular-file`: the profile entry line plus the offending
      root's or archive's entry line, so a symlinked root and, later, a blob archive are two Skips.
  - `Commit` is the first commit, in `Commits` order, that carries that triple.
  - So a name that stays archived twice across many states yields one `Skip`, not one per state. Everything else at every
  commit is still emitted.
- **Everything else fails `Revisions`** (SDD-INV-49):
  - a candidate that fails `ValidateRevision` for an unlisted reason;
  - a reader or ref failure, `ErrShallow`, or cancellation.

  None of these can be caused by tenant bytes in history. Each is fixed by fixing the provider or the
  clone.
- The provider checks the data causes, builds the candidate, then calls `ValidateRevision`. An error at
  that point is returned, never skipped.
- **Batch-safe through projection.** Every returned revision passes `ValidateRevision`, which after
  §12 also enforces the size bound. The provider's own `nul-byte` check keeps U+0000 out of `sdd`
  payloads. The batch has no duplicate `(source, id, digest)`
  and no dangling same-source reference. So `Sync` validates, orders, and appends it (SDD-INV-47), and
  a rebuild projects it (SDD-INV-52).
- **Failure moves to the claim.** Only a claim on an unmappable state fails, at `Resolve`.
- **Skips are derived, not ledgered.** They are recomputed from the refs on every sync and carried in
  the connector's `Result`. The CLI prints one line per Skip (`skip <cause> <unit> at <commit>`) after
  the counts line, and exits 0 (SDD-INV-60). A skip can only withhold trust.
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
10. **SDD-INV-10 — Feature-sibling change intents are feature-level only.** *(M2: Spec Kit; meaning unchanged, proven when M2 lands)*
11. **SDD-INV-11 — RETIRED (draft 4): resolution also reads parent archive listings; replaced by INV-33.**
12. **SDD-INV-12 — Canonical payload bytes are stable across runs.**
13. **SDD-INV-13 — Two kits run on one provider with only the profile differing.** *(M2: Spec Kit; meaning unchanged, proven when M2 lands)*
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
42. **SDD-INV-42 — A Spec Kit status comes only from the explicit map or the declared absent value; otherwise that record is unmappable.** *(M2)*
43. **SDD-INV-43 — Duplicate ids, empty name-parts, and unpaired renames make that record unmappable; zero entries yield no record.**
44. **SDD-INV-44 — Each distinct (key, manifest) yields exactly one requirement revision, with a content-addressed id.**
45. **SDD-INV-45 — A reused change name makes that change unmappable at that state only.**
46. **SDD-INV-46 — Only the closed list of data causes yields a Skip, each fixture's Skip set is exactly the pinned one, and dependants cascade.**
47. **SDD-INV-47 — Every returned batch syncs end to end: validation, ordering, and append.**
48. **SDD-INV-48 — Within one state, two keys minting one record id are both unmappable.**
49. **SDD-INV-49 — A failure outside the closed cause list fails Revisions; it is never a Skip.**
50. **SDD-INV-50 — RETIRED (draft 7): Resolve gains a fourth outcome and a name domain; replaced by INV-59.**
51. **SDD-INV-51 — The reader shares the git connector's read contract: replace refs ignored, shallow and graft refused, refs validated.**
52. **SDD-INV-52 — A sync containing intent.sdd events projects and rebuilds.**
53. **SDD-INV-53 — Every payload string is valid UTF-8 without U+0000, and every statement is non-blank; a violation skips its unit.**
54. **SDD-INV-54 — Only regular files are read under SDD roots; a symlink or gitlink a unit reads makes that unit unmappable.**
55. **SDD-INV-55 — A path outside the SDD roots, whatever its bytes, never affects the sdd provider.**
56. **SDD-INV-56 — `sdd` runs only when selected by name; `sync intent all` neither constructs nor walks it.**
57. **SDD-INV-57 — The §4.2 parser agrees with OpenSpec's reader on the pinned corpus.**
58. **SDD-INV-58 — A Spec Kit id outside the canonical obligation grammar makes that record unmappable.** *(M2)*
59. **SDD-INV-59 — Resolve is total over tenant states: a revision or one of four sentinels; any other error comes only from the reader.**
60. **SDD-INV-60 — `sync intent sdd` prints every Skip and exits 0.**
61. **SDD-INV-61 — Unexpected object types at a root, the archive, or the profile path are state-level skips, never reader errors; files a unit neither parses nor carries are never a cause.**
62. **SDD-INV-62 — RETIRED (draft 10): its literal-pathspec clause is untestable (§3); replaced by INV-67.**
63. **SDD-INV-63 — RETIRED (draft 9): "equal outcomes" was false for Resolve, which also reads parents; replaced by INV-65.**
64. **SDD-INV-64 — M1 maps OpenSpec only; a speckit profile is profile-invalid.**
65. **SDD-INV-65 — The state key carries the mode and object id of the profile entry and of every root; equal keys imply equal changeOutcome and equal Skips.**
66. **SDD-INV-66 — A regular file directly under changes.root or changes.archive is never a change, a unit, or a cause.**
67. **SDD-INV-67 — Profile paths follow the literal grammar; a malformed path is profile-invalid, never a reader error.**

Each proving test names **the wrong implementation it kills**. Where it matters, it also names **a
neutral change that must survive**. Retired invariants have no row.

| Invariant | Statement | Proving test | Kills / survives |
|---|---|---|---|
| SDD-INV-2 | No business decision is emitted; requirements render authorization undeclared | sdd_test.go::TestNoBusinessDecisionIsEmitted | kills: mapping `proposal.md` to a BD |
| SDD-INV-5 | Inserting, deleting, or reordering one entry changes no other obligation id | sdd_test.go::TestObligationIDsAreNotPositional | kills: index-based ids |
| SDD-INV-7 | Every vector in `testdata/id-vectors.json` (byte-wise ASCII mapping; non-ASCII inputs `expoſe`, `ıd check`, the Kelvin sign, `Café menu`, `ünïcode-only`; over-length; `!!!` → `empty-name-part`; `a` → `slug-too-short`) is reproduced exactly | sdd_test.go::TestRecordIDsAreDeterministic | kills: a suffix omitting `kind`; **`strings.ToUpper` / `ToLower`** (`EXPOSE` ≠ `EXPO-E`) |
| SDD-INV-9 | Every ADDED, MODIFIED, REMOVED, RENAMED entry is an `implements` target | sdd_test.go::TestDeltaSectionsMapToClosedRelations | kills: dropping REMOVED |
| SDD-INV-12 | Repeated runs over one history emit byte-identical payloads | sdd_test.go::TestCanonicalizationStability | kills: map-ordered obligations |
| SDD-INV-18 | The change manifest changes when any file of the change directory or the profile changes | sdd_test.go::TestChangeManifestDigest | kills: a manifest over delta files only |
| SDD-INV-20 | The leftover-after-archive state, a `--skip-specs` archive, and a deleted main spec all sync | sdd_test.go::TestArchiveNeverBreaksSync | kills: batch-fatal name reuse |
| SDD-INV-26 | (a) Changing the owner changes the digest. (b) Changing a non-manifest input (a main spec, another change's files, `src/`, git author or date) leaves every revision byte-identical | sdd_test.go::TestPayloadIsPinnedByManifest | kills: reading outside the manifest; survives: non-manifest edits |
| SDD-INV-28 | A change with only `proposal.md` yields no change intent and no error | sdd_test.go::TestEmptyChangeIsAbsent | kills: emitting a target-less CI |
| SDD-INV-30 | Files under the main-spec path never produce a revision | sdd_test.go::TestMainSpecsAreNotIngested | kills: ingesting `openspec/specs` |
| SDD-INV-31 | Over `branches`: `Commits` equals the git connector's list, and every successful `Resolve(c, n)` is in `Revisions()` | sdd_test.go::TestResolveIsMemberOfRevisions | kills: a HEAD-only walk; a path-limited walk (TREESAME) |
| SDD-INV-32 | Active; archived-by (single parent); side-branch archive then merge (the merge fails, the side commit resolves); root with archive (resolves); dual presence (unmappable); archived earlier; unknown | sdd_test.go::TestNativeNameResolution | kills: a first-parent-only archive check |
| SDD-INV-33 | A reader that fails on any other commit, and on any parent read other than an archive listing | sdd_test.go::TestResolveReadSet | kills: resolving from a cached HEAD state |
| SDD-INV-34 | The digest equals a manifest recomputed independently from `git cat-file` bytes; a one-byte edit changes it; a CRLF copy of a delta has a different digest but the same obligations | sdd_test.go::TestDigestIsManifestOfExactBytes | kills: digesting normalized bytes; survives: CRLF for parsing |
| SDD-INV-35 | Every data cause — `hostile()`'s set, the pinned record-id collision pair, `owner-uncovered`, `sponsor-uncovered` and `profile-invalid`, which need three profiles and so span three states — placed (a) at a non-tip commit fixed later, (b) at a non-HEAD ref tip, and (c) **at HEAD**, with each of the three states taking HEAD in its own variant, leaves `Revisions` without error and every other unit emitted — proven by clean units (`fresh-data`, and `good` with a statement only the narrow state carries) that exist **only** in states carrying Skips, so dropping a defect state's clean revisions fails. `branches` step 14 `stale` and `head-on-leftover` pin the same shapes over committed fixtures through SDD-INV-46 | sdd_test.go::TestHistoricalFailClosedIsRecoverable | kills: batch-fatal; tip-fatal; **HEAD-fatal** |
| SDD-INV-39 | Each `fenced/` case yields exactly its `openspec_mask` list. The test re-parses with a test-only naive-toggle mask (§13) and with no mask, and must reproduce the `naive_toggle` and `unmasked` lists, so every list is checked. Non-neutral cases differ; `document-renames` is the one neutral case | sdd_test.go::TestCodeFencesAreMasked | kills: unmasked; a naive toggle; a stale pin; survives: the neutral |
| SDD-INV-40 | Each §2.1 violation (unknown, missing, null, or duplicate field; unknown enum; `kit: speckit` in M1; kit/mode mismatch; a path outside the literal grammar; a glob outside the root; a blank or U+0000 owner/sponsor; non-UTF-8) skips that state with `profile-invalid` | sdd_test.go::TestProfileIsStrict | kills: a lenient decoder; a sync-fatal profile error |
| SDD-INV-41 | Owner and sponsor come from their list's first match at the same state; uncovered → `owner-uncovered` / `sponsor-uncovered` | sdd_test.go::TestIdentityOnlyFromProfile | kills: falling back to the git author |
| SDD-INV-43 | Colliding ids, empty name-parts, and unpaired renames are skipped with their cause. A RENAMED-only delta with one unpaired `FROM` is a skipped record unit (`unpaired-rename`), and its change is `cascade` (`resolve-edges` `rename-only`). A zero-entry file with no unpaired line yields no record and no skip | sdd_test.go::TestObligationDefectsArePerRecord | kills: last-wins; `O-ADD-`; treating unpaired-only as zero entries |
| SDD-INV-44 | Editing a delta yields a new record id and leaves the earlier record unchanged; each `(key, manifest)` has one revision | sdd_test.go::TestRequirementRecordsAreContentAddressed | kills: name-keyed BR ids |
| SDD-INV-45 | Name reuse skips that change at that state; at the next state the change maps again; its records are emitted throughout | sdd_test.go::TestReusedNameIsPerState | kills: picking the active copy; a sticky skip |
| SDD-INV-46 | Every fixture history's Skip set equals `testdata/expected-skips.json` exactly (units, causes, first steps); a CI targeting a skipped BR is skipped as `cascade` | sdd_test.go::TestSkipSetsArePinned | kills: a partial-skip bug (e.g. map-ordered obligations skipped instead of failing) |
| SDD-INV-47 | Every fixture history, the hostile ones included, passes `Sync` against a test appender with zero errors. A requirement payload of **exactly** `core.MaxPayloadBytes` is emitted and appended; one byte more is `payload-too-large` and never reaches append. Also proven by recheck_kill_test.go::TestExactBoundPayloadSyncs | sdd_test.go::TestBatchSyncsEndToEnd | kills: emitting an oversize payload (partial append) |
| SDD-INV-48 | A pinned colliding key pair at one state (found offline, in testdata) makes both units unmappable; the same pair at two states yields two revisions of one id | sdd_test.go::TestRecordIDCollisionIsPerState | kills: first-wins; a cross-state lookup |
| SDD-INV-49 | A provider mutant that emits unsorted obligations makes `Revisions` return an error, not a Skip | sdd_test.go::TestProviderDefectsFailNotSkip | kills: the draft-4 catch-all |
| SDD-INV-51 | Over a repository with a `git replace --graft` ref, `Commits` and `Parents` equal the git connector's (no false root); a shallow clone and an `info/grafts` file make `Revisions` fail with `ErrShallow` | gitcmd_test.go::TestSDDReaderSharesReadContract | kills: a bare `git rev-list` reader |
| SDD-INV-52 | `openspec-history`, plus a synthesized history carrying one unit for **each** of the 15 M1-reachable §7 causes (`profile-invalid`, `invalid-utf8`, `nul-byte`, `non-regular-file`, `owner-uncovered`, `sponsor-uncovered`, `obligation-collision`, `empty-name-part`, `slug-too-short`, `unpaired-rename`, `record-id-collision`, `name-reuse`, `archive-name-invalid`, `payload-too-large`, `cascade`), syncs and then **rebuilds projections** with zero errors. Requirement and change-intent rows exist for the clean units | sdd_projection_test.go::TestSDDEventsProject | kills: an unregistered route; a NUL statement reaching a `TEXT` column |
| SDD-INV-53 | **U+0000** on every M1 route that reaches a payload string (OpenSpec statement, name fallback, RENAMED `TO`) → `nul-byte`; in a profile owner or sponsor → `profile-invalid`. No OpenSpec route can produce a blank statement (§4.2); the Spec Kit `statement-blank` route is M2's (Appendix A) | sdd_test.go::TestPayloadStringsAreStoreSafe | kills: a NUL-tolerant provider |
| SDD-INV-54 | A symlinked delta, a tree at `…/spec.md`, a symlinked `specs/` and `specs/<cap>`, a symlinked change directory, a gitlink under the root, and a symlinked **dated archive entry** each skip their unit with `non-regular-file`. A symlinked `design.md` and a gitlink elsewhere in a change directory are **not** causes; they enter the manifest as `<mode> <oid>`, recomputed independently from git. A regular file at `specs`, or directly under `specs/`, is ignored: no unit, no cause, and the change is spec-less (`ErrNoChangeIntent`). A regular sibling is emitted. Also proven by survivors_test.go::TestRegularFileNamedSpecsIsIgnored, survivors_test.go::TestSymlinkManifestLineIsGitIdentity, survivors_test.go::TestSymlinkManifestLineFormatIsPinned | sdd_test.go::TestNonRegularFilesAreUnits | kills: following a symlink; a reader error on a gitlink; an active-only rule 4; hashing a symlink target as content |
| SDD-INV-55 | A history adding `caf\xe9.md` outside the roots syncs `sdd` with no error and no skip. The same bytes as a **non-parsed file inside a change directory** are not a cause either (INV-61); as a **change directory name** they skip that change with `invalid-utf8` | sdd_test.go::TestForeignPathsAreIgnored | kills: a reader that lists per-commit file diffs; an over-broad `invalid-utf8` |
| SDD-INV-56 | `sync intent all` on a shallow clone with no profile behaves exactly as before M1; only `sync intent sdd` constructs the provider. `proofbound verify` reaches intent sync only through `syncIntentOnStore(…, "all")`, whose provider set is pinned directly: `all` builds `records` and `specdir`, never `sdd` — outside any repository, and in a full repository where `sdd` itself is constructible. Every `syncIntentOnStore` call inside `verify` passes the literal `"all"` (checked on `cli.go`'s syntax tree, since running `verify` needs a ledger). Also proven by `TestAllSelectionNeverConstructsSDD`, `TestAllSelectionExcludesSDDWhereItCouldBeBuilt` and `TestVerifySyncsIntentWithAllOnly` in package `internal/cli` | cli_test.go::TestIntentAllExcludesSDD | kills: adding `sdd` to `all` |
| SDD-INV-57 | Round 5's divergence table (REMOVED bullet; `## ADDED  Requirements`; NBSP after `##`; a one-space `## ` line; `Requirementſ`; interleaved RENAMED) plus `expected-obligations.json` match OpenSpec's reader output, as pinned in testdata | sdd_test.go::TestParserMatchesOpenSpecReader | kills: `DELTA_HEADER` as section rule; a Go `\s`/`(?i)` translation |
| SDD-INV-59 | Every `(step, name)` row of `testdata/expected-resolve.json` yields its pinned outcome. The rows cover `resolve-edges` (archived twice by a root; re-propose / re-archive; a profile-invalid state; a short name; proposal-only; a zero-entry leftover; plain files under the root and the archive; rename-only; out-of-domain names) and `branches` (side-branch archive then merge; root). **Synthesized by the M1 test**, since the replay helper cannot commit symlinks or gitlinks: absent profile → `ErrUnknownChange`; symlinked root or blob archive → `ErrUnmappable:non-regular-file`; symlinked change directory → `ErrUnmappable:non-regular-file`; a change whose only delta is a symlink → `ErrUnmappable:cascade`. Over every listed commit and every name, `mapState`'s skips and `Resolve` agree on the cause. A reader failing on purpose — in `Read`, `Parents`, or `List` — yields none of the sentinels. A regular file in a parent's archive never archives a change. Also proven by survivors_test.go::TestParentsFailureIsNotASentinel, survivors_test.go::TestParentArchiveFileDoesNotArchive, recheck_kill_test.go::TestResolveParentListFailureIsAReaderError | sdd_test.go::TestResolveIsTotal | kills: archived-twice falling through; `ErrArchivedEarlier` on re-archive; `ErrNoChangeIntent` before `name-reuse` or `cascade`; plain files as changes |
| SDD-INV-60 | `proofbound sync intent sdd` over `unmappable/reused-name` prints `skip name-reuse change add-login at <sha>` and exits 0. A unit with a C0 or C1 control, DEL, U+2028/U+2029, or invalid UTF-8 prints Go-quoted on one line. Within a state, revisions and skips come out in a fixed order. Also proven by cli_test.go::TestSyncIntentSDDQuotesHostileUnits, `TestPrintableUnitQuotesEveryLineBreaker` in package `internal/cli`, survivors_test.go::TestOutputOrderIsDeterministic | cli_test.go::TestSyncIntentSDDPrintsSkips | kills: silent skips; a non-zero exit |
| SDD-INV-61 | A symlinked `openspec/changes`, a blob at `changes.archive`, and a gitlink root each skip the state with `non-regular-file`; a directory or symlink at the profile path is `profile-invalid`. `Revisions` never errors, and later clean states are emitted. These three are three Skips, counted as Skips (not as a unit-and-cause set). A binary image (non-UTF-8, with NULs) inside a change directory changes the manifest but is not a cause. Also proven by recheck_kill_test.go::TestStateLevelSkipsAreCounted | sdd_test.go::TestStateLevelObjectTypes | kills: dereferencing `<c>:<path>` (a reader error); an over-broad `invalid-utf8` |
| SDD-INV-64 | `speckit-history` under M1 yields exactly one `Skip` (`profile`, `profile-invalid`) and no revision | sdd_test.go::TestM1IsOpenSpecOnly | kills: a half-implemented Spec Kit path |
| SDD-INV-65 | (a) A state whose profile path is a symlink to text byte-identical to a valid profile has a different key from the regular-profile state, and is `profile-invalid` while the other maps. (b) Over every fixture, two commits with equal keys have equal `changeOutcome` for every name and equal Skips. `Resolve` may differ, because it also reads parents (`branches` 07 vs 08; the cache-by-key mutant is SDD-INV-59's kill). (c) Profiles that differ only in mode (regular vs executable) have different keys, so the oid-only key cannot pass on line count alone (round-1 F8). Also proven by survivors_test.go::TestStateKeyModeOnlyDifference | sdd_test.go::TestStateKeyCarriesModes | kills: an oid-only key |
| SDD-INV-66 | `resolve-edges` 07 carries `openspec/changes/IMPLEMENTATION_ORDER.md` and `openspec/changes/archive/README.md` (the real shape, at 1393 of 1624 OpenSpec commits). Neither is a change, unit, or Skip, and `Resolve` gives `ErrUnknownChange` for both names | sdd_test.go::TestPlainFilesAreNotChanges | kills: draft 8's "any object type" change entry |
| SDD-INV-67 | Profiles whose root or archive is `:(glob)openspec/changes`, `:!x`, `../x`, `/x`, `openspec/changes/`, `./openspec/changes`, `openspec//changes`, or `open spec/changes`; an archive that is not a direct child of the root; and `requirements.paths` other than the exact pair (for example `openspec/changes/*/spec.md`, or a glob with no `*`) are each `profile-invalid`. `Revisions` never errors | sdd_test.go::TestProfilePathsAreLiteral | kills: a lenient grammar (a malformed path reaching git, rc 128). |
| SDD-INV-68 | `StateRevision(commit, name)` returns exactly what `Resolve` returns, except that a change `Resolve` refuses with `ErrArchivedEarlier` yields its revision in that commit's own state (an archived path). Added 2026-11-19 for R2 M4 drift, which asks what a change promises at HEAD rather than whether a commit may claim it (review `r2-m4-drift-review-20261119` F1) | state_revision_test.go::TestStateRevisionAgreesWithResolveExceptArchivedEarlier |

## 10. Decisions this draft settles or withdraws

- **A, B, C:** unchanged from draft 3 (warranty across archive withdrawn; main-spec lifecycle
  withdrawn from R2; Spec Kit lifecycle answered by content addressing).
- **D: fail-closed scope.** Per unit, never per history, never per ref tip, never per HEAD (§0: four
  branch tips today, and `main` itself for 13 first-parent commits, 8 of them non-merge). SDD-INV-35(a-c) kill all three alternatives.
- **E: walk.** The git connector's walk under its read contract, without its per-commit file lists
  (§3).
- **F: data causes vs defects.** A closed list, derived from every check tenant bytes reach (§7).
- **G (owner, 2026-11-17): `all` excludes `sdd`.** Including it would make `sync all`, `verify`, and
  `delivery-enforce` walk full history in every repository, and would inherit `ErrShallow` in shallow
  CI checkouts that work today. Revisit at M2, once an OpenSpec repository has synced end to end.
- **H: the statement is Proofbound's,** the one labelled divergence from OpenSpec's reader (§4.2).
- **J (owner, 2026-11-17): M1 is OpenSpec-only, over a literal-path profile**
  (`VD-sdd-m1-openspec-first-2026-11-17`).
- **I: `ValidateRevision` changes are only ever retroactively safe.** The projector re-validates
  every historical intent event on rebuild (`projections.go:311`, `:327`, `:350`), so a new check is
  admissible there only if no ledgered event can fail it. The size bound qualifies, because append
  enforces it. A U+0000 rejection does not, so it lives in the `sdd` provider (round-6 F2).

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
O-1 / O-2 / O-10 (M1 for OpenSpec, M2 for Spec Kit), O-3 / O-4 (M1-M2, M5), O-8 (M3), O-12 (all).

## 12. Amendments that land with M1 — complete list (round-4 F3)

Found by `grep -rn 'SourceIntentSpecdir\|"specdir"\|intent\.specdir'` over `kernel/` and `scripts/`
(non-test) at `f6813bf`. Every site that registers a provider today is listed and assigned. The M1
review re-runs the grep.

**Kernel, M1:**

| Site | Change |
|---|---|
| `core/core.go:56` and `core/SPEC.md:75-83` (core's interface lock) | add `SourceIntentSDD Source = "intent.sdd"` in both |
| `connector/intent/intent.go:134` (`New`) | admit provider name `sdd` |
| `connector/intent/intent.go:215` (`ValidateRevision`) | admit `intent.sdd`. **For every provider:** enforce `len(payload) ≤ core.MaxPayloadBytes`. This is retroactively safe, because `core.Event.Validate` already rejects such payloads at append, so no ledgered event exceeds it (§10 I). It closes the partial-append path for `records` / `specdir`. **No** U+0000 rule here (round-6 F2) |
| `connector/intent` (`Result`) | additive `Skip { Commit, Unit, Cause string; Err error }`, optional `interface{ Skipped() []Skip }`, and `ProviderResult.Skipped` (omitted when empty) |
| `projections/projections.go:296` (`eventRoute`) | route `intent.sdd` intent kinds |
| `kernel/internal/gates/gates.go:304` (`intent-reference-integrity`) | include `intent.sdd` |
| `cli/cli.go:57-60` (command enum), `:92`-area parse, `:513-514` (dispatch), `:37` (usage) | add `sync intent sdd` |
| `cli/cli.go:366-382` (`syncIntentOnStore`) | construct the `sdd` provider **only when the selection is `sdd`** (SDD-INV-56) |
| `cli/cli.go:~528` (the `listed=… appended=… existing=…` print) | print one `skip …` line per `Skip` (SDD-INV-60) |
| `connector/git/gitcmd` | factor the shared listing out of `Commits`; add the `sdd` Reader on `Repo` (§3), every path-taking command with `--literal-pathspecs`. The git connector's behaviour is unchanged |

**Unchanged by M1, on purpose:** `cli.go:596` (`sync all`), `:835` and `:854` (`verify`),
`scripts/delivery-enforce.sh:38` (which runs `sync all`, reaching `"all"` through `cli.go:596`), and
`docs/releases/r1/seed-demo-app.sh:94`. All of them still mean `records` + `specdir`.

**Documentation:** `docs/releases/r1/USER-GUIDE.md:269` documents R1's
`sync intent {records|specdir|all}` and stays accurate for R1. `sync intent sdd` is documented in R2's
release documentation, at the R2 release. That is not an M1 code site.

**Scripts and generated artifacts, M1:**

| Site | Change |
|---|---|
| `scripts/p6-route-matrix.sh:19-24` | add the three `intent.sdd` routes, then regenerate the `@generated` route matrix |
| `docs/plans/p6-census-rows.tsv` | add C3-019 (the package), C4-013 (the connector), and C5-053..060 (the two `intent.sdd` kinds × four consumers). **As done:** `scripts/p6-census.sh` itself is deliberately **not** edited, because its last commit bounds the closed P6 C6 intent-commit window; editing it pulls every later commit into the P6 census |
| `scripts/p6-task2-close.sh` | add `intent/sdd` to the pinned connector universe, classified C4-013 synthetic (missed by draft 10, found by the gate during M1) |

**Intent SPEC text, M1:**
- **§1:** point at `sdd/SPEC.md`.
- **§2:** the `Skip` report.
- **§4:** `sdd`'s manifest digest and parent archive reads.
- **§6 and INT-INV-11:** for `sdd`, `ErrUnmappable` marks a unit and is reported as a `Skip`; the
  foreign-fixture STOP applies to `records` and `specdir`.
- **INT-INV-4:** its test gains an oversize-payload row that fails before any append, and an
  exact-bound row: a payload of exactly `core.MaxPayloadBytes` still validates. That row is the
  retroactive-safety argument, not a replay of a ledgered event: `core.Event` validation already
  rejected any payload over the same bound at append, so no event ledgered before M1 exceeds it
  (round-1 F8 corrected this row's wording).

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

  `expected-obligations.json` pins each case's obligations under three masks: OpenSpec's mask, the
  **naive toggle**, and no mask.
  - **Naive toggle** (round-6 F6): every line matching ``^[ \t]*(```|~~~)`` flips an in-fence flag and
    is itself masked, ignoring marker, length, and info string. It is the mutant SDD-INV-39 kills.
  - The file is regenerated by
    `python3 -I kernel/internal/connector/intent/sdd/testdata/derive/history_probe.py --pins kernel/internal/connector/intent/sdd/testdata`
    (its `generated_by` field), and never hand-edited.
- **`testdata/resolve-edges/`** (7 steps) and **`testdata/expected-resolve.json`** pin `Resolve` over
  every round-7 corner, plus round 8's plain files and rename-only delta (SDD-INV-59, 66). The outcomes were hand-derived from §6's precedence, and each row
  names the rule that decides it.
- **`testdata/id-vectors.json`** (`history_probe.py --vectors`, `@generated`) pins ids with non-ASCII
  inputs (SDD-INV-7).
- **`testdata/expected-skips.json`** pins each fixture's exact Skip set under M1 (SDD-INV-46),
  de-duplicated per §7. `speckit-history` is one `profile` / `profile-invalid` (SDD-INV-64).
- **Mechanised now (round-5 F6):** `TestPinnedExpectationsMatchFixtures` checks that:
  - every pinned skip names an existing step and unit (all four unit forms), and a cause from §7's
    closed list;
  - `expected-resolve.json` rows name existing steps and known outcomes;
  - both generated pins carry the `@generated by` marker, and `id-vectors.json` holds non-ASCII
    inputs;
  - every fenced case keeps its column-0 MODIFIED header and a header for each pinned obligation;
  - each case's `discriminates` flag matches its lists, and exactly one case is neutral;
  - HEAD of `head-on-leftover` is on the dual-presence state.

  Calibrated: renaming the MODIFIED header fails the test; a bogus skip unit fails it; a bogus cause
  fails it; a neutral edit passes. Pin *values* are checked by M1's INV-39 / INV-46 tests, which
  re-derive all three lists and the exact skip sets.
- **Byte-level hostile cases** (NUL, oversize, invalid UTF-8, blank statements, long ids, symlinks,
  gitlinks, a non-UTF-8 path) are **synthesized by the M1 tests**, not committed (SDD-INV-52..55).
  Committed binary-hostile bytes would trip repository hygiene for no gain.
- **Calibration:** every §9 row names its kill. M1 runs those mutants through
  `scripts/mutant-apply.sh` before its counts are believed.
- **Probe honesty:**
  - `derive/history_probe.py` implements §4.2, compares statements as exact bytes, walks the provider's
    walk, and prints every §0 count, including the 13/8 first-parent row.
  - `--pins` and `--vectors` need no git repository.

**M3 landed (2026-11-19, `CI-sdd-warranty-m3-xaqocp`, `VD-sdd-m3-git-binding-2026-11-19`).** The
contract is the M3 SPEC, `docs/plans/R2-sdd-m3-git-binding-spec-draft5.md`, as amended by drafts 6-8.
It changed these sites:
- the git connector: per-commit `sdd` claim skips (git SPEC INV-31/32);
- the CLI commit-claim resolver: the `sdd` arm, with the lazy ledger set `L` (CLI-INV-13/14);
- projection validation: provider `sdd` (P-INV-33);
- the reviews connector's hold and refusal (R-INV-13, CLI-INV-15/16).

This package's production code is unchanged.

**SDD-INV-56 in substance after M3:** `sync intent all`, `sync all`'s intent step, `verify`'s intent
step and delivery-enforce still never construct or walk `sdd` for **ingestion**. The commit-claim
resolver of `sync git`, `sync all` and `verify` constructs the provider lazily, only when a commit
carries an `sdd:` claim, and only calls `Resolve` at that commit. It never calls `Revisions`, so it
never walks history to ledger `sdd` revisions. A repository without `sdd:` claims constructs nothing
(CLI-INV-14).

## 14. Non-goals and boundaries owned elsewhere

- Main-spec ingestion, folding, tombstones, and carry-over (deferred with the main-spec lifecycle).
- Authoring or writing tenant SDD files; governance → business decisions; YAML; heuristic extraction;
  Spec Kit per-task intents; API / DB providers; more than one profile per repository; ledgering skips.
- **M3 entry conditions** (all answered when M3 landed, 2026-11-19: §12's M3 note):
  1. the `sdd:` trailer token and its decision record;
  2. anti-spoofing against a native name the git connector reads from the commit tree itself,
     including the profile's `kit`;
  3. **per-commit handling of an unresolvable claim.** Today the git connector fails the sync closed.
     A claim at one of the 23 dual-presence commits would wedge every git sync. M3 uses §6's sentinels
     to report it per commit, or proves the trailer gate refuses it first;
  4. the M3 sites in §12;
  5. the readiness gate on `sdd` chains.
- **M2 entry condition, added by M3:** bindings are monotone (`commit_intents_view` is insert-only),
  so a binding made under M1's resolution rules keeps warranting its commit even if M2's rules would
  not bind that claim. M2 must state in its own SPEC whether existing bindings stay, and ship a
  migration if they do not. The same holds for any other change that narrows `sdd` binding (a change
  to a `Resolve` outcome, or a stricter anti-spoofing check). This is the git SPEC's "accepted
  limitation of R2 M3" (`VD-sdd-m3-git-binding-2026-11-19`).
- **Recorded, not owned here:** the git connector itself (`gitcmd.Commits` → `files`) fails on any
  non-UTF-8 path anywhere in history. That is a pre-existing git-connector wedge, independent of `sdd`,
  and `sdd` does not inherit it (§3). It is raised to the owner as its own item.
- **Recorded, not owned here (round-7 F8):** a `records` or `specdir` statement, or any other value
  projected into a `TEXT` column, that carries U+0000 passes `ValidateRevision`, appends, and then fails
  every projection rebuild.
  - Draft 6's attempt to close this inside `ValidateRevision` was withdrawn (§10 I), because the check
    would be retroactive. No ledger in this repository is exposed today; round 6 grepped every record.
  - A safe fix needs an append-time-only check outside `ValidateRevision`. That is a P5/records matter,
    raised to the owner, and not an M1 site.
- **M4:** change-drift reporting (only if re-proposed with a ledger-evaluable definition) and the status
  export.

## Appendix A — Spec Kit (M2; NOT part of M1's contract, non-normative)

Parked here by `VD-sdd-m1-openspec-first-2026-11-17` so that M2 starts from what earlier drafts
established, together with its known open items. M2 writes the normative version and has it reviewed by
a non-author.

- **Profile (draft 7 shape):** `kit: speckit`; `requirements.paths` (for example `specs/*/spec.md`);
  `requirements.status` with `line` (named group `status`), `map`, and optional `absent`;
  `changes: { mode: feature-sibling, sibling, status }`. Literal path grammar as §2.1.
- **Records:** one per matched feature spec. Obligation lines match
  `^- \*\*((?:FR|SC|NFR)-[0-9]+)\*\*:? (.+)$`, giving id `O-<id>`.
  - Causes: `obligation-id-invalid`, `statement-blank`, `status-unmapped`.
- **Change intents:** one per feature directory that has its sibling; status `changes.status`.
- **Open items for M2 (round 7 F2):**
  - a `Resolve` rule for Spec Kit;
  - a change-intent id formula;
  - feature-key collisions under multi-`*` globs (`specs/a/x/` and `specs/b/x/`);
  - SDD-INV-10, 13, 42, and 58's proofs.
