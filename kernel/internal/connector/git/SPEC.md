# internal/connector/git — SPEC

Contract for the git connector. Written before implementation (Build Law 6). § 2 is the
interface lock: changing a signature is a reviewed diff to THIS FILE first, then code — never
the reverse. `core`'s § 2 drifted from its code once (2026-08-11) because the lock was prose
only; this package pins its surface mechanically from the start (INV-14).

---

## 1. Purpose

**Owns:** turning a git repository's history into `commit.recorded` events in the ledger. It
lists commits, builds one payload per commit, and appends each through a `*store.Sync`.

**Does NOT own:**

| Not owned | Home | Why not here |
|---|---|---|
| The ledger, the lock, `seq`, migrations | `internal/store` | the ONLY package that opens the DB |
| Event identity, canonical JSON, `content_sha`, validation | `internal/core` | the connector fills `NewEventParams` and lets core mint |
| What "already seen" means | the ledger's UNIQUE index | see INV-2; this connector holds no cursor |
| Projections, reports, `[superseded]` marking | `internal/projections` | derived state |
| Deciding WHEN to sync | `cmd/vera` | this package exposes one function and returns |
| **Running `git`** | `internal/connector/git/gitcmd` | see the split below — `os/exec` is BANNED here and the ban is tested (INV-15) |

**The two-package split, and why it is not gold-plating.** `package git` is pure: it takes a
`Repo` and an `Appender` and mints events. `package gitcmd` implements `Repo` by shelling out.
The reason is not taste, it is that the Task-4 acceptance criteria require REAL repositories —
amend, rebase of N, second-branch-and-switch, each built in `t.TempDir()` at test runtime. Those
cases must exercise real git, because whether an amend produces a new SHA is a property of git
and not of this code. So: `gitcmd` may use `os/exec`; `git` may not; and the rewrite tests live
in the external test package `git_test`, which imports `gitcmd`, builds a real repo, and runs a
real `Sync` against a FAKE `Appender` — real git, no database. `gitcmd` importing `git` for the
`Commit` type creates no cycle, because an external test package may import a package that
imports the package under test.

**The load-bearing design point, inherited and not re-litigated here** (P1 plan, "verified
design points that survived the review panel"): the events UNIQUE index
`(source, native_id, content_sha)` IS the seen-set. There is no date watermark and no stored
cursor driving what to read. `sync_runs.cursor_json` records the tip set for OBSERVABILITY only
and is never read back to decide anything (INV-9, INV-23).

**Which tests actually DISCRIMINATE that design point — corrected 2026-08-12.** This section
previously claimed the proof was INV-6/7/8: "a watermark is what breaks on amend, rebase and branch
switch … and every one of those is a required test here". That was **false as a proof claim**. A
reviewer installed a real date watermark in `Sync` and ran each cited test in isolation: all three
still PASSED. Their fixture pinned `GIT_COMMITTER_DATE` to one constant for every commit, so no date
filter could exclude anything. Only `TestSync_SecondSyncAppendsNothing` failed.

The discriminating test is **INV-18**: a commit dated BEFORE one already ingested is still
ingested, which every date-ordered cursor skips — inclusive or exclusive, package-level or held on
the Connector.

**INV-2 is NOT a discriminating test, and naming it here was a second false proof-claim** (corrected
2026-08-13, round 2 MED-2). A reviewer installed three real watermark shapes and measured each named
test in isolation: INV-2 passed under all three. The first correction to this section replaced one
wrong claim with a narrower wrong claim, which is worth stating because it is the failure mode of
correcting under time pressure.

INV-18 itself only became sufficient in the same repair. It had built a FRESH `Connector` for its
second `Sync`, wiping any state the Connector carries — so a watermark stored on the Connector and
advanced at end-of-`Sync` passed untouched. It now reuses ONE Connector across both syncs, and that
shape dies. A test that rebuilds the object under test between observations cannot see state that
object holds.

INV-6/7/8 remain true and worth having as behavioural statements about amend, rebase and
branch-switch; they are not the proof that a watermark is impossible. Their fixture now advances its
dates one minute per git invocation, so it can no longer hide a date filter by accident.

**Scope of "history".** Commits are listed with `--no-replace-objects` from
`--exclude=refs/stash --exclude=refs/notes/* --exclude=refs/replace/*` applied to `--all`. See
INV-19: bare `--all` reaches `refs/stash` and `refs/notes/*`, and a stash alone mints two commit
objects — but it also reaches HEAD, which a hand-rebuilt scope drops.

(`refs/replace/*` was added 2026-08-13 and this line did not mention it for one commit.
`prescription-lint` could not catch that: it verifies every flag a SPEC quotes EXISTS in the code,
never that the quote is COMPLETE. A stated limit of that gate, not a surprise.)
A shallow repository is REFUSED rather than partially ingested (INV-20), at construction AND at every
listing. Object replacement is neutralised so a `replace --graft` cannot rewrite what is recorded.
All of these are `gitcmd`'s decisions; [`gitcmd/SPEC.md`](gitcmd/SPEC.md) § 3 and § 5 state them with
their consequences.

## 2. Interface

The complete exported surface of `package git`. Nothing else is exported.

```go
// Repo is the git surface this connector needs, declared HERE at the consumer so a test can
// supply a fake without a real repository, and so the package never grows a dependency on a
// git library (Law 8: no new dependency without a decision record).
type Repo interface {
	// Commits returns every commit reachable from any ref in reverse Git date-order, with
	// parents before children. This is not a strict timestamp sort. Ordering is a convenience
	// for readable ledgers, NOT a correctness property:
	// dedupe is the ledger's index, so a reordered or repeated list changes nothing.
	Commits(ctx context.Context) ([]Commit, error)
	// Tips returns the current ref tips, for the observability cursor only.
	Tips(ctx context.Context) (map[string]string, error)
}

// Commit is one commit as this connector reads it. Field set is the payload contract.
type Commit struct {
	SHA            string    `json:"sha"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedAt    time.Time `json:"committed_at"`
	Subject        string    `json:"subject"`
	FilesTouched   []string  `json:"files_touched"`
	CitedDecisions []string  `json:"cited_decisions"`
}

// Appender is the store surface this connector needs — declared at the consumer, and
// deliberately NOT *store.Sync, so a test can drive the connector without a database.
type Appender interface {
	Append(ctx context.Context, e core.Event) (store.Record, bool, error)
}

// Deps documents every constructor input. Exported fields are required unless noted.
type Deps struct {
	Repo   Repo            // required
	IDs    *core.IDGenerator // required
	Logger *slog.Logger    // required
}

type Connector struct{ /* unexported */ }

func New(d *Deps) (*Connector, error)

// Version is the connector_version stamped on every event this package mints. It is a
// CONSTANT, not derived from the build, so a replayed ledger says which code shape wrote a
// row (core INV-25 requires a non-empty value).
const Version = "git/1"

// Result is what one Sync reports. Counts are DERIVED from the appends, never accumulated by
// the caller.
type Result struct {
	Listed   int             // commits the repo returned
	Appended int             // appends that INSERTED a row
	Existing int             // appends the idempotency index absorbed
	Cursor   json.RawMessage // the tip set, for sync_runs.cursor_json — observability only
}

// Sync lists every reachable commit and appends one event per commit. It is idempotent: a
// second Sync over unchanged history appends nothing (INV-2).
func (c *Connector) Sync(ctx context.Context, a Appender) (Result, error)
```

`FilesTouched` is the commit's tree delta from its first parent. For a root commit it is the full
tree. This gives merges one deterministic meaning instead of Git's mode-dependent empty output.
Paths that are not valid UTF-8 are refused because JSON strings cannot reversibly preserve their
Git byte identity.

**Not exported, deliberately:** the payload struct is marshalled from `Commit` itself, so
there is no second shape to keep in step (Law 2 — one home per datum). A caller that wants
the payload reads the ledger.

## 3. Invariants

Every invariant names a test in § 5. Numbers are permanent identifiers — new invariants
append, none is inserted or re-ordered (the Build Law lesson of 2026-08-11 applies to
invariant numbering too, and always did).

**Identity and idempotency**

1. **INV-1 — One event per commit, keyed by SHA.** `NativeID` is the commit SHA verbatim.
   `Kind` is `commit.recorded`, `Source` is `git`. Nothing else is minted per commit.
2. **INV-2 — A second Sync over unchanged history appends ZERO.** Not "few", zero. Dedupe is
   the ledger's UNIQUE index via the idempotent append, so this holds with no cursor, no
   watermark, and no state carried between runs.
3. **INV-3 — `Result` counts are exact and partition the listing.**
   `Appended + Existing == Listed` after a successful Sync, and `Appended` counts only appends
   that INSERTED. **Tested in BOTH directions** — never over, never under. (The store's own
   counter was false in one direction, then false in the other after the repair; an equality
   tested one way is not tested. VD-fix-discipline-0e0tnz rule 2.)
4. **INV-4 — Payload is the commit, canonically.** The payload is `Commit` marshalled to JSON;
   `content_sha` therefore changes if and only if a payload field changes. Field NAMES are part
   of the contract (they are the wire), pinned by the vector in INV-13.
5. **INV-5 — The same commit read twice yields the same `content_sha`.** No timestamp, no
   ordering, no map iteration order may enter the payload. `FilesTouched` and
   `CitedDecisions` are therefore SORTED and de-duplicated before marshalling, and an EMPTY
   slice normalises to `null` — never `[]` (pinned by the second vector in § 5.1).

   **Was FALSE, now held by a refusal (2026-08-12).** A shallow clone broke this. `--name-only`
   at a grafted boundary lists the whole tree rather than what that commit changed, so the same
   commit read before and after `--unshallow` produced two different payloads — and since
   `content_sha` is half the idempotency key, the ledger ended up holding TWO contradictory
   events for one SHA with nothing to say which was true. That is unfixable after the fact in
   an append-only store, so the repository is refused up front instead (INV-20). The
   empty-slice half was separately unpinned: flipping `null` to `[]` re-keys every commit that
   cites no decision, and the § 5.1 vector could not see it because its slices were non-empty.

**History rewrites — the reason there is no watermark**

6. **INV-6 — After `commit --amend`, a Sync appends exactly one event** (the new SHA), and a
   second Sync appends zero. The amended-away commit's event REMAINS in the ledger: the
   ledger is append-only and records what was observed, not what currently exists.
7. **INV-7 — After a rebase of N commits, a Sync appends exactly N events** (the N new SHAs),
   and a second Sync appends zero. A date watermark would append 0 here, which is the defect
   this design exists to avoid.
8. **INV-8 — After a commit on a second branch and a switch, a Sync appends exactly one**, and
   a second Sync appends zero. Commits are listed from ALL refs, so which branch is checked out
   never changes what is ingested.

**Cursor, ordering, failure**

9. **INV-9 — The cursor is written and never read.** `Result.Cursor` carries the tip set for
   `sync_runs.cursor_json`. No code path in this package reads a previous cursor to decide what
   to list. Asserted mechanically by a source scan, because "we didn't use it" is exactly the
   kind of claim that rots (INV-15).
10. **INV-10 — Listing order does not affect the outcome.** Syncing a shuffled commit list
    produces the same set of events and the same `Appended` count as sorted order. Ordering is
    presentation.
11. **INV-11 — A failed append aborts the Sync and reports the cause.** The error is returned
    wrapped, `Result` reports what had been appended before the failure, and the connector
    appends nothing further. Partial progress is real and is stated, not hidden.
12. **INV-12 — A malformed commit is refused by core, not smuggled past it.** The connector
    does not sanitise: an empty SHA, a control character in the SHA, or a zero `CommittedAt`
    surfaces as core's validation error naming the wire field.

    **Corrected 2026-08-12, during implementation.** This invariant originally named "a control
    character in a subject". That is not true and was never true: `core` guards control
    characters in `native_id` and `connector_version` — the fields that reach logs and CLI
    output verbatim — and treats payload content as opaque JSON, where NUL escapes legally as
    `\u0000`. git also refuses NUL in a commit message, so the input could not arise. The test
    written from this line failed against correct code, which is how it was caught. Recorded
    rather than silently reworded: an invariant that asserts more than the code promises is a
    defect in the SPEC, and the temptation on finding one is to "fix" the code to match it.

**Pinned vectors and the locks**

13. **INV-13 — Pinned payload vector.** A fixed `Commit` marshals to an exact byte string with
    an exact `content_sha`, both pinned in § 5.1. A dependency swap or a field rename fails
    HERE first, which is the point of pinning it (P1 plan: "pin any known-input→known-output
    vectors so a dependency swap fails here first").
14. **INV-14 — § 2 IS the exported surface, mechanically.** Checked by parsing this file and
    the package's real exports. `core` shipped exports absent from its own interface lock
    because the lock was prose; this package is born with the check.
15. **INV-15 — No cursor read, no direct DB, no git binary in this package.** A source scan
    refuses: any import of `database/sql`, `pgx`, or `os/exec` in a non-test file of THIS
    package, and any read of a previous cursor in this package OR in `gitcmd`.

    **Scope corrected 2026-08-12.** Both scans read only this directory, so `gitcmd` — the one
    package that could ever pass `--since` — was invisible to them. A reviewer planted
    `database/sql`, a `watermarkSince` variable and a `readCursor` call inside `gitcmd.Commits`,
    gofmt-clean, and `make check` stayed GREEN. INV-15 as WORDED stayed true, which is exactly
    what made the hole invisible; SPEC § 4's project-level non-goal was unguarded at the only
    site it would be violated. The purity ban stays scoped to this package (running git is
    `gitcmd`'s job); the DB ban and the no-cursor rule now cover both, and the no-cursor scan
    also refuses `--since` / `--after` / `--until` / `--before` / `--max-count` as literal
    arguments. The INV-15 row in § 5 also cited `TestPackagePurity` for the "no cursor read"
    clause, which that test never checked — now cited correctly.

**Appended 2026-08-12 after the first adversarial round.** Numbers are permanent identifiers, so
these append. Each names behaviour that already existed but had no witness, or a decision this
connector makes that was previously recorded nowhere.

16. **INV-16 — Decision ids are read from the commit BODY, sorted and de-duplicated.** (Was
    listed above; the sort was previously "detected" only by Go's map-iteration randomisation,
    which killed a sort-removal mutation in 3 runs out of 25. A coin flip is not an assertion.)
17. **INV-17 — An empty repository syncs to zero events and is not an error.**
18. **INV-18 — A commit dated BEFORE one already ingested is still ingested.** This is the
    discriminating test for the no-watermark design point (see § 1): every date-ordered cursor
    skips such a commit, so no implementation carrying one can pass. Real cases: a rebase
    preserving committer dates, a merged long-lived branch, a fetch of history written elsewhere.
19. **INV-19 — Stash and notes refs are not history; everything else reachable is.** Commits come
    from `--all` with `refs/stash` and `refs/notes/*` EXCLUDED. Bare `--all` reached those two:
    one `git stash push` mints TWO commit objects and `git notes add` a third, so a repository
    with one real commit produced four `commit.recorded` events — permanently, and growing with
    ordinary local workflow.

    A commit on a non-checked-out branch, reachable only from a tag, or sitting on a DETACHED
    HEAD is history and is still ingested. The detached-HEAD arm is not decoration: the first fix
    rebuilt the scope by hand as three flags [retracted] and silently lost every such commit,
    because `--all` includes HEAD and the rebuild did not. See `gitcmd/SPEC.md` G-INV-3/G-INV-4,
    which own this behaviour and its tests.
20. **INV-20 — A shallow repository is refused, not partially ingested.** `New` returns
    `ErrShallow`. See INV-5 for why partial ingestion cannot be corrected afterwards.
21. **INV-21 — No commit content can break the output framing.** NUL is the only separator: git
    refuses NUL in a commit message and POSIX forbids it in a filename, so no content can
    produce one. Records are delimited by a fixed FIELD COUNT, not by a byte that content might
    contain, and a stream that is not a whole number of records is a loud error rather than a
    silent mis-attribution of fields to the wrong commit.
22. **INV-22 — A broken repository is an error, never an empty one.** A corrupt ref store must
    not be byte-indistinguishable from a fresh repository — it previously was, and `Sync`
    reported complete success with zero events.
23. **INV-23 — The cursor is the repository's REAL tips.** INV-9 covers "never read"; this
    covers "actually written". Without it, `Result.Cursor`, the `Tips` call, and the whole
    mechanism were deletable — and fabricating the cursor's contents also passed.
24. **INV-24 — Every scalar comes from its own git placeholder.** Author and committer name and
    email are distinct fields; `occurred_at` is the COMMITTER date (`%cI`), not the author date;
    the subject is git's first line, not the whole body. Each of those was a surviving mutation,
    and each is a deterministic wrong value that nothing downstream could ever notice.
25. **INV-25 — A c-quoted path is decoded.** git quotes a path containing a newline, quote,
    backslash or non-ASCII byte — which is what stops such a path breaking the newline-delimited
    file block, so the quoting is kept and undone on read. Previously the payload recorded the
    quoted string verbatim, naming a path that does not exist.
26. **INV-26 — `Version`'s VALUE is pinned.** It is stamped on every row this package mints;
    changing it silently re-labels the provenance of every future event.
27. **INV-27 — Every dependency is required at construction.** `New` refuses a nil `Deps`,
    `Repo`, `IDs` or `Logger`, so a missing dependency fails at wiring time rather than as a nil
    dereference mid-sync.

## 4. Non-goals

A reviewer should reject, on sight:

- **A cursor, watermark, or `--since` date** that decides what to list. INV-2/6/7/8 exist to
  make its absence testable. This is a verified P1 design point; weakening it needs a new
  decision record, not a commit.
- **A git library dependency.** `Repo` is an interface at the consumer; the real implementation
  shells out and lives behind it. A new dependency needs a `VD-` record first (Law 8).
- **Deleting or updating events** when history is rewritten. The ledger is append-only; an
  amended-away commit's event stays (INV-6). "Cleaning up" is the opposite of the product.
- **A second payload struct** beside `Commit`. One home per datum (Law 2).
- **Sanitising commit data** to make core accept it. If core rejects a commit, that is the
  finding (INV-12).
- **Counting appends in the caller.** `Result` is derived (INV-3).
- **Reading `sync_runs`** for anything. It is a journal, not an input.

## 5. Invariant table

Most invariants own one row. The third cell names a real Go test function in this package —
`scripts/invariant-lint.sh` (BLOCKING, in `make check`) fails the build if a citation here does
not resolve. It guarantees resolution only; whether the named test PROVES its claim stays with
adversarial review, and that has been got wrong three times in this repo.

| Invariant | Statement | Proving test |
|---|---|---|
| INV-1 | One event per commit: native_id is the SHA, kind is commit.recorded, source is git | connector_test.go::TestSync_MintsOneCommitRecordedEventPerCommit |
| INV-2 | A second Sync over unchanged history appends ZERO | connector_test.go::TestSync_SecondSyncAppendsNothing |
| INV-3 | Appended + Existing == Listed, and Appended counts only inserts — asserted in BOTH directions | connector_test.go::TestSync_ResultCountsPartitionTheListingExactly |
| INV-4 | The payload is the Commit marshalled canonically; a changed field changes content_sha | connector_test.go::TestPayload_FieldChangeChangesTheContentSHA |
| INV-5 | The same commit read twice yields the same content_sha; slices are sorted and de-duplicated | connector_test.go::TestPayload_IsStableAcrossReads |
| INV-6 | After an amend, one event is appended and the amended-away event REMAINS | rewrite_test.go::TestSync_AmendAppendsTheNewShaAndKeepsTheOld |
| INV-7 | After a rebase of N commits, exactly N events are appended | rewrite_test.go::TestSync_RebaseAppendsEveryNewSha |
| INV-8 | After a commit on a second branch and a switch, exactly one event is appended | rewrite_test.go::TestSync_SecondBranchIsIngestedRegardlessOfCheckout |
| INV-9 | The cursor is written for observability and never read to decide what to list | surface_test.go::TestNoCursorIsEverRead |
| INV-10 | A shuffled commit list yields the same events and the same counts | connector_test.go::TestSync_ListingOrderDoesNotChangeTheOutcome |
| INV-11 | A failed append aborts the Sync, reports the cause, and states partial progress | connector_test.go::TestSync_AppendFailureAbortsAndReportsProgress |
| INV-12 | A malformed commit surfaces core's validation error rather than being sanitised | connector_test.go::TestSync_MalformedCommitIsRefusedByCore |
| INV-13 | The pinned payload vector marshals to the exact bytes and content_sha in § 5.1 | connector_test.go::TestPayload_PinnedVector |
| INV-14 | § 2's declared surface is exactly what the package exports | surface_test.go::TestExportedSurfaceMatchesTheSpec |
| INV-15a | No database/sql, pgx or os/exec import in any non-test file of THIS package | surface_test.go::TestPackagePurity |
| INV-15b | No cursor read and no date-limiting git argument, in this package OR gitcmd | surface_test.go::TestNoCursorIsEverRead |
| INV-15c | gitcmd does not open the database either | surface_test.go::TestGitcmdDoesNotReachTheDatabase |
| INV-18 | A commit dated before one already ingested is still ingested | rewrite_test.go::TestSync_ACommitOlderThanAnyAlreadySeenIsIngested |
| INV-23 | The cursor is the repository's real tips | connector_test.go::TestSync_CursorIsTheRepositoryTips |
| INV-26 | Version's value is pinned | connector_test.go::TestVersion_IsPinned |
| INV-27 | Every dependency is required at construction | connector_test.go::TestNew_RequiresEveryDependency |
| INV-4b | An empty slice normalises to null, pinned by a second vector | connector_test.go::TestPayload_EmptySlicesArePinnedToNull |
| INV-14b | § 2's struct FIELDS are the complete exported field set | surface_test.go::TestExportedStructFieldsMatchTheSpec |
| INV-16 | Decision ids are read from the commit BODY, sorted and de-duplicated | rewrite_test.go::TestSync_CitedDecisionsAreReadFromTheCommitBody |
| INV-17 | An empty repository syncs to zero events and is not an error | rewrite_test.go::TestSync_EmptyRepositoryIsNotAnError |

**INV-19 through INV-22, INV-24 and INV-25 are `gitcmd`'s and live in
[`gitcmd/SPEC.md`](gitcmd/SPEC.md)** under that package's own numbering — the ref scope, the shallow
refusal, the framing, the broken-versus-empty distinction, placeholder mapping and path unquoting.
They moved there for the reason § 6 already stated ("if it ever grows a decision of its own it earns
a SPEC"): those are decisions, not translation. One SPEC governs one package, which is also what
`invariant-lint` assumes — its citation grammar has no path separator, so a parent SPEC cannot cite
a child package's tests, and a SPEC that tried would be describing code it does not govern.

**INV-16 and INV-17 were APPENDED during implementation (2026-08-12), not inserted** — numbers are
permanent identifiers. Both name behaviour that was already written and tested but had no row, and
a behaviour with no row is one a future edit can drop with nothing noticing:

- **INV-16** — `cited_decisions` is populated by parsing the commit body for `VD-<slug>-<6>`,
  de-duplicated and sorted (INV-5's normalisation applies). The in-package payload tests set that
  field directly, so they prove the MARSHALLING and never the READING: without this row, a `gitcmd`
  that parsed no decision id at all would pass every other test in the package. The
  decision-citation trail is what the ledger exists to support, so it earns an invariant rather
  than remaining an incidental test.
- **INV-17** — a fresh repository has no commits and no refs, and git reports that through an exit
  code rather than through empty output. `vera sync` runs against whatever it is pointed at, so
  "no history yet" must be zero events, not a failure.

### 5.1 Pinned vector

Pinned 2026-08-12 from the first real run of `payloadOf`. The vector is a fixed `Commit` with
every field populated, an unsorted `FilesTouched` and a duplicated entry in `CitedDecisions`, so
it proves the normalisation of INV-5 as well as the bytes:

```
INPUT   Commit{
          SHA:            "9f2b7c1d8e5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c",
          AuthorName:     "A Author",    AuthorEmail:    "author@example.test",
          CommitterName:  "C Committer", CommitterEmail: "committer@example.test",
          CommittedAt:    2026-08-12T09:00:00Z,
          Subject:        "pin the vector",
          FilesTouched:   ["b.go","a.go","b.go"],
          CitedDecisions: ["VD-fixture-aaaaaa","VD-fixture-aaaaaa"],
        }

BYTES   {"author_email":"author@example.test","author_name":"A Author",
         "cited_decisions":["VD-fixture-aaaaaa"],"committed_at":"2026-08-12T09:00:00Z",
         "committer_email":"committer@example.test","committer_name":"C Committer",
         "files_touched":["a.go","b.go"],"sha":"9f2b7c1d8e5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c",
         "subject":"pin the vector"}

        ^ ONE line, no whitespace — the breaks above are for reading only.

SHA256  eaaf4869ccfa6491f2da707db289c4f3a72aa851a9d2c7db8dcb07e9af6be7ad
```

Read the bytes against the input: keys alphabetical (RFC 8785), `files_touched` sorted with the
duplicate gone, `cited_decisions` reduced to one. Those three differences ARE INV-5 — which is
why the vector uses deliberately messy input rather than tidy input.

To regenerate after an INTENDED change: blank both consts at the top of `payload_test.go`, run
`go test ./internal/connector/git/ -run TestPayload_PinnedVector`, and the failure prints the
exact lines to paste back into the test and into this section.

**It was deliberately not invented in advance.** Writing a plausible-looking hash here before the
code existed would have been a fabricated vector, and a fabricated vector is worse than none: it
"passes" the moment someone makes the code match it, and thereafter proves only that the code
equals itself. The consts sat empty and the test sat red until there was a real run to pin.

**The SECOND vector — empty slices.** Added 2026-08-12, because the first one could not see the
decision it most needed to pin:

```
INPUT   the same Commit, with FilesTouched and CitedDecisions both empty

BYTES   {"author_email":"author@example.test","author_name":"A Author",
         "cited_decisions":null,"committed_at":"2026-08-12T09:00:00Z",
         "committer_email":"committer@example.test","committer_name":"C Committer",
         "files_touched":null,"sha":"9f2b7c1d8e5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c",
         "subject":"pin the vector"}

SHA256  32074ae0364663ee0a8c9b4afb5fa5900f7c8c152cf0c3423568f9da67fdda31
```

An empty slice normalises to `null`, never `[]`. A nil slice and an empty slice must produce
identical bytes, or the same logical commit becomes two events depending on how its caller built the
slice. The first vector's slices are both NON-empty, so it never took this branch: changing
`return nil` to `return []string{}` survived the entire suite while this section still promised such
a change "fails HERE first". One vector cannot pin a branch it does not take.

**What a change here means.** These bytes feed `content_sha`, which is half the ledger's
idempotency key. Changing the field set, a `json` tag, or the normalisation re-keys every commit
event: existing rows keep their old `content_sha`, the next sync appends a SECOND event per
commit, and the ledger holds both. That is correct append-only behaviour rather than corruption —
but it is migration-shaped, and the commit message must say so. The empty-slice choice re-keys
every commit that cites no decision, which is nearly all of them.

## 6. Dependencies

- `internal/core` — event identity, canonical JSON, validation.
- `internal/store` — the `store.Record` type only, via the `Appender` interface at the consumer.
- Standard library: `context`, `encoding/json`, `errors`, `fmt`, `log/slog`, `reflect`, `sort`,
  `time`.
  (Corrected 2026-08-13 by reading the import block: this line listed `strings`, which the package
  does not import, and omitted `errors` and `fmt`, which it does. A dependency list written from
  memory is a claim like any other.)

**No new external dependency.** The real `Repo` implementation shells out to `git` and lives in
the sibling package `internal/connector/git/gitcmd` (§ 1). `os/exec` is banned from THIS package
by INV-15, so the ban is testable rather than aspirational. Any git library would need a `VD-`
record before `go get` (Law 8).

**`gitcmd` now has its own SPEC** — [`gitcmd/SPEC.md`](gitcmd/SPEC.md), with its own `G-INV-*`
numbering and its own tests.

This paragraph used to say the opposite: that `gitcmd` "owns no invariant of its own beyond faithful
translation … If it ever grows a decision of its own it earns a SPEC; today it would be a file with
nothing in it but a restatement." That was true when written and false within a day. The trigger it
named fired on 2026-08-12, when an adversarial round found FOUR HIGH defects there — every one of
them a decision the package was making silently and testing not at all: what counts as a ref
(`--all` ingesting stash and notes), what to do with a shallow clone (recording a boundary commit's
whole tree as its changed files), how to frame the output (a record separator that ordinary commit
content could contain), and whether a broken repository is distinguishable from an empty one.

The lesson is worth more than the correction. "Thin translation layer, no decisions of its own" was
the reason it shipped with no tests, and it was self-fulfilling: a package nobody tests is a package
whose decisions nobody notices it is making. `invariant-lint` assumes one SPEC per package — its
citation grammar has no path separator — so the split was also the only way its tests could be
cited at all.
