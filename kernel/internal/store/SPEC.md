# SPEC — `internal/store`

**Status:** authored before implementation (Build Law 6) · P1 Task 3 · 2026-08-08
**Authority:** [docs/plans/P1-flight-recorder-plan.md](../../../docs/plans/P1-flight-recorder-plan.md) § Architecture / § Ledger rules · [VD-stack-go-fid9mi](../../../docs/decisions/VD-stack-go-fid9mi.md) (blessed dependency set) · [internal/core/SPEC.md](../core/SPEC.md) (the consumed envelope) · [docs/design/continuity-chain.md](../../../docs/design/continuity-chain.md)
**Lock rule:** § 2 is the interface lock. Changing a signature is a reviewed diff to THIS FILE first, then code — never the reverse (`/vera-review` enforces).
**Re-confirmation:** Task 3's DoD requires re-confirming the plan's DB/locking design against real embedded-postgres behavior. § 4 is that re-confirmation — every number and every behavior there was **measured on 2026-08-08**, not assumed. Three plan statements did not survive contact and are corrected in place (§ 4 F1, F2, F6).
**Review hardening (2026-08-09):** an adversarial pass demonstrated that a live lock holder never re-checked that it still held the lock, and that the read seam could not be used from inside its own callback. Both were measured, not argued (§ 4 F14, F15); the lock was rebuilt on an ownership token with an atomic publish and a compare-and-swap takeover, and INV-33 … INV-38 were added. Every change in that pass is recorded here first.

**Lock replaced with `flock(2)` (2026-08-09, round 3).** The rebuilt lock failed review too, in the same place and for the same underlying reason. Measured (§ 4 F14b): detection latency equalled `HeartbeatInterval` — a cached loss flag meant that at the SHIPPED 60s default a lock lost to a `git clean` stayed invisible for up to a minute, and two processes appended interleaved to one ledger for 2.6 seconds with no error on either side; the takeover compare-and-swap did not mutually exclude, because `os.Link` shares the source inode so every reclaim marker was born already stale; and a constant `newNonce` passed the whole suite. The common cause is not any of those three bugs: it is that **a regular file was being asked to answer "does anybody else hold this ledger", and every answer a regular file can give is an inference about some past instant.**

**Round 4 — four defects the flock rebuild left standing (2026-08-09).** The exclusion
mechanism itself survived adversarial review: the unlink attack is closed at the shipped
default, and `flock` withstood the whole battery (F16-F21). Four things around it did not.
**(F1)** nothing related `LockPath` to `DataDir` — two Stores on ONE data directory with two
lock paths both acquired and both appended, 40 interleaved rows, no race required; the lock
is now DERIVED from the data directory and a disagreeing `LockPath` is `ErrConfig`
(INV-42). **(F2)** INV-41 claimed zero operations succeed after a loss, and 8 did, the last
4.012s later; `Append` now re-verifies inside the critical section, and the invariant is
rewritten to the bound it can actually prove — one write round trip, measured in F22 —
because *a spec that overclaims is itself the defect*. **(F3)** a dispossessed Store stopped
the postmaster the new legitimate holder was using; it no longer stops a server it does not
own (INV-43). **(F4)** three mutations survived the suite and a gate was green only 6 runs
in 7; two are closed with new proving tests, the third is stated under INV-41, and the flaky
gate is diagnosed in **F24** — a closed listening socket is not released while a sibling
`fork`/`exec` is in flight, reproduced at 4.2% under fork pressure, which is `F21`'s
mechanism on a different file descriptor. Two unchosen numbers found while hunting it are
measured and replaced in F23.

So the exclusion mechanism is no longer hand-rolled. It is an exclusive `flock(2)` held for the life of the `Store` (§ 4 F16–F21, measured 2026-08-09). The kernel releases it when the holder dies for ANY reason, which deletes staleness, takeover, reclaim markers, pid liveness *for the lock*, the ownership nonce, and the heartbeat — together with the invariants that existed only to make those safe (INV-22, INV-24, INV-38 RETIRED; INV-23, INV-25, INV-27, INV-37 AMENDED; INV-39, INV-40, INV-41 ADDED). The one thing `flock` does not cover — the lock lives on the inode, so unlinking the PATH admits a second holder — is closed by an identity check at the operation choke point rather than on a timer. Interface changes are in § 2 and are a deletion in every case: `Config.StaleLockAfter`, `Config.HeartbeatInterval`, `LockInfo.TookOverFrom`, `LockedError.Alive`, and the lock record's `nonce`.

---

## 1. Purpose

`store` is **the only package in the kernel that opens a database.** It owns the append-only
event ledger (`events`), the sync-run journal (`sync_runs`), the single-data-dir lock, the
embedded Postgres lifecycle, the one `*pgxpool.Pool`, and the goose-migrated LEDGER schema.
Everything above it — connectors, projections, `cmd/vera` — receives a `*Store` injected and
never imports a driver.

Its read surface is the **P1→P2 seam**: the P2 gates engine is designed to be just another
`ReadEvents` consumer, not a new reader of the database.

**store does NOT own:**

| Not owned | Home | Why not here |
|---|---|---|
| The envelope, canonical JSON, `content_sha`, ULID identity, kinds | `internal/core` | store persists what core constructs; it never re-derives a hash and never mints an id. |
| Payload meaning (commit fields, witness v1 JSON, session metadata) | each `internal/connector/*` SPEC | `payload` is opaque JSON to store; it is stored and returned verbatim (INV-7). |
| Projection tables, reducers, `[superseded]` marking, the week report | `internal/projections` | derived state. store lends a transaction; projections own the DDL and the rows. **Projection DDL never enters the migration stream** (INV-19). |
| What to ingest, cursors as correctness | each connector | `sync_runs.cursor_json` is observability only; the UNIQUE index **is** the seen-set. |
| Deciding *when* to sync, CLI flags, the repo root | `cmd/vera` | `Config.Root` is injected; store never guesses a path from the working directory. |
| Wall-clock reading | `cmd/vera` (via `Config.Now`) | the clock is injected so `sync_runs` timestamps are deterministic in tests. (It was originally injected to age a lock; lock staleness no longer exists.) |

Per the CLAUDE.md one-home table, `kernel/internal/<pkg>/SPEC.md` is the single home of a
package contract; code cites the spec, never restates it.

---

## 2. Interface

The complete exported surface of `package store`. Nothing else is exported (INV-14).

### 2.1 Configuration and lifecycle

```go
// Config is the injected configuration. Root is REQUIRED; every other path
// defaults beneath it. store never infers a location from the process's working
// directory — the composition root supplies it.
type Config struct {
	// Root is the VERA state directory (conventionally "<repo>/.vera"). Created if absent.
	Root string

	// DataDir is the Postgres data directory (default <Root>/db). It and
	// RuntimeDir MUST NOT contain each other IN EITHER DIRECTION:
	// embedded-postgres deletes RuntimeDir on EVERY start, so a ledger nested
	// under it is destroyed silently (§ 4, F2) — and a RuntimeDir nested under
	// the ledger takes ledger files with it when it goes. Open refuses both
	// (INV-27).
	DataDir string

	// RuntimeDir is scratch that embedded-postgres wipes on every start
	// (default <Root>/pgrun). Nothing durable may live here — not the ledger,
	// and not the lock.
	RuntimeDir string

	// BinariesDir holds the extracted Postgres binaries and PERSISTS across runs
	// (default <Root>/pgbin). Keeping it OUT of RuntimeDir is what makes a warm
	// Open cost ~0.21s instead of ~5.4s (§ 4, F1).
	BinariesDir string

	// LockPath is OPTIONAL, and it is NOT a choice. The lock is a DERIVED property
	// of the data directory it protects: it lives BESIDE DataDir, named after it
	// (<DataDir>.lock), so one data directory has exactly one lock. Supplying this
	// field is an ASSERTION about where that lock will be — a value naming any
	// other file is refused with ErrConfig, never honoured (INV-42).
	//
	// It survives as an assertion rather than being deleted because the composition
	// root spells its paths explicitly, and a spelling that disagrees with the data
	// directory must fail loudly instead of being silently redirected.
	//
	// store holds an exclusive flock(2) on it for the life of the Store. The file's
	// CONTENT is informational and its EXISTENCE means nothing: the file outlives
	// every release on purpose (INV-25). It MUST NOT be inside RuntimeDir: a lock
	// the next start deletes is not a lock (INV-27).
	LockPath string

	// Port for the embedded server. Zero means the DataDir-derived default
	// (§ 4, F10 as amended by F23 — a private band BELOW every platform's
	// ephemeral floor) so two repositories on one machine never collide. Ignored
	// when DatabaseURL is set.
	Port uint16

	// DatabaseURL, when non-empty, points at an already-running Postgres: no
	// embedded server is started and NO data-dir lock is taken (INV-26). Single-writer
	// discipline then belongs to the operator — see § 6.
	DatabaseURL string

	// There is NO lock-timing configuration. The lock is an exclusive flock the
	// kernel releases when its holder dies, so there is no staleness window to
	// size and no heartbeat to keep shorter than it. StaleLockAfter and
	// HeartbeatInterval were deleted with the machinery they tuned, and INV-14
	// keeps them deleted: a reintroduced knob fails the pinned-surface scan until
	// this section is amended.

	// MaxConns caps the pool (default 4). Exposed as a plain int — never a pgx
	// type — so tests can pin it to 1 and prove ReadEvents leaks no connection (INV-13).
	MaxConns int

	// Now is optional; nil means time.Now. It stamps sync_runs.started_at /
	// finished_at and the lock file's informational acquired_at. It is no longer
	// load-bearing for the lock: it was injected so lock-STALENESS tests could
	// advance a clock, and staleness no longer exists.
	Now func() time.Time

	// Logger is optional; nil means a discard logger.
	Logger *slog.Logger
}

// Open takes the ledger lock, starts (or adopts, § 4 F4) Postgres, applies the
// ledger migrations, and returns the single Store. There is no way to obtain a
// Store without holding the lock — "every command that opens the ledger takes the
// lock" is therefore structural, not a convention. The caller MUST Close.
func Open(ctx context.Context, cfg Config) (*Store, error)

type Store struct{ /* unexported */ }

// Close closes the pool, settles ownership, stops the Postgres server, and
// releases the lock — in that order, best effort, errors joined. Close is
// idempotent; after it, every method returns ErrClosed.
//
// It stops the server whether this process started it or adopted an orphan, which
// is what makes a crashed run self-healing (§ 4, F4) — WITH ONE EXCEPTION. A Store
// whose lock was lost does NOT stop the server: the server belongs to whichever
// process holds the lock now, and stopping it kills the ledger out from under the
// legitimate holder (INV-43). Such a Store releases only what is still its own,
// the pool and the descriptor, and reports the loss. The server it leaves running
// is picked up by the next Open through ordinary adoption (INV-28).
//
// Releasing means CLOSING THE DESCRIPTOR the kernel associates the flock with. The
// lock FILE is deliberately left on disk: unlinking it is the two-holder window
// (INV-41), so presence of the file never means the ledger is held.
//
// Close returns an error wrapping ErrLockLost when the file at LockPath stopped
// being the file this process locked before it ran. "Released the lock" is a lie
// when the lock has belonged to somebody else since before Close (INV-37). Close
// DETECTS that itself — it must, because it cannot decide about the server
// otherwise — rather than relying on an earlier operation to have noticed.
func (s *Store) Close() error

// Lock reports the lock this Store holds. The zero LockInfo when DatabaseURL was set.
func (s *Store) Lock() LockInfo

type LockInfo struct {
	Path       string
	PID        int
	AcquiredAt time.Time
}
```

**The lock itself.** The lock is an **exclusive `flock(2)`** on `LockPath`, taken
`LOCK_EX|LOCK_NB` at `Open` and held until `Close` closes the descriptor. That is the
entire exclusion mechanism. Everything below the flock is a label on it.

- **The kernel is the arbiter.** A second acquirer fails immediately with
  `EWOULDBLOCK` (§ 4, F16) — no pid probing, no liveness heuristic, no age threshold,
  and nothing to decide.
- **`LOCK_NB` is load-bearing.** Without it acquisition BLOCKS for the holder's
  remaining lifetime (§ 4, F20), so a busy ledger would hang a command instead of
  reporting who has it.
- **The kernel releases on death, however the holder dies** — `kill -9` included
  (§ 4, F17). This is why there is no staleness rule to configure: a lock file left
  by a crash is just a file.
- **The lock is per open-file-description**, so a second `open`+`flock` inside the
  SAME process is refused too (§ 4, F19). In-process exclusion is free.
- **The lock lives on the INODE, not the path.** Unlinking `LockPath` leaves this
  process holding a lock on a file nothing can reach while another process creates
  and locks a new one there — two holders (§ 4, F18). `os.SameFile(fstat(held fd),
  stat(LockPath))` is checked before EVERY ledger operation (INV-41), and the
  descriptor is never released by unlinking (INV-25).

**The lock record (informational only).** `LockPath` holds one line of JSON:

```json
{"pid":41207,"acquired_at":1786312045}
```

- It exists so an operator — and `LockedError` — can NAME the holder. `flock` refuses
  without saying who won, so this is the only source for that name.
- **Nothing in the exclusion decision reads it.** A record that is missing, torn,
  hand-edited, or stale costs the holder's name in one error message and never
  correctness. There is no ownership token, because ownership is not something this
  file is asked to prove any more.
- It is written **in place** (truncate + write), never write-temp-then-rename: a
  rename replaces the inode, and the inode is the lock. A concurrent reader can
  therefore observe a torn record and report an unidentified holder — an
  informational degradation, accepted deliberately.
- The previous design's *read-once* discipline (take bytes and mtime through a single
  descriptor, because a `ReadFile` plus a separate `Stat` could pair one process's
  content with another's mtime) is **moot and gone**: mtime is no longer read at all,
  and content decides nothing.

Changing the record's shape is a diff to this section first, but it is no longer a
change to the locking algorithm — only to what an operator gets told.

### 2.2 Appending — only through a sync run

```go
// Record is a ledger row: the ledger-assigned seq plus the envelope.
type Record struct {
	Seq   int64
	Event core.Event
}

// Sync is one connector pass. It is the ONLY append surface: every event in the
// ledger belongs to a sync run, and events_appended is DERIVED from the appends
// that actually inserted — never a number the caller reports (INV-6).
type Sync struct{ /* unexported */ }

// BeginSync writes a sync_runs row with started_at set and finished_at NULL. A row
// that keeps finished_at NULL is an honest record of a crashed pass.
func (s *Store) BeginSync(ctx context.Context, connector string) (*Sync, error)

// Append inserts e when (source, native_id, content_sha) is new.
//
//	inserted=true  -> Record is the row just written.
//	inserted=false -> the identical event was already present; Record is the
//	                  EXISTING row — its seq and its original event_id, never the
//	                  rejected candidate's (INV-2).
//
// A revision (same source+native_id, new content_sha) is a NEW row at a higher seq
// (INV-3). e.Validate() runs BEFORE the round trip: a cheap guard ahead of an
// expensive one (INV-5).
//
// Appends are SERIALISED across every goroutine and every Sync on one Store
// (INV-36). § 6's SinceSeq argument requires a single writer; the data-dir lock
// delivers that between processes, this delivers it within one. Concurrent callers
// are delayed, never refused. A finished run refuses every further append.
func (sy *Sync) Append(ctx context.Context, e core.Event) (Record, bool, error)

// Appended is the running count of inserted=true appends on this run.
func (sy *Sync) Appended() int64

// Finish closes the run: finished_at, events_appended (from Appended), cursor_json
// (observability only — never read back to decide what to ingest), and the text of
// cause when non-nil. Idempotent; the first call wins.
//
// The run is marked finished only AFTER the journal row is written, so a failed
// write is a failed Finish and a retry actually retries (INV-34).
func (sy *Sync) Finish(ctx context.Context, cursor json.RawMessage, cause error) error
```

### 2.3 Reading — the P1→P2 seam

```go
// Filter selects an ordered slice of the ledger. A zero field means "any".
// Exactly one shape of range is offered: bounds are EXCLUSIVE lower bounds.
type Filter struct {
	Source        core.Source // exact match; "" = any
	Kind          core.Kind   // exact match; "" = any
	SinceSeq      int64       // EXCLUSIVE: seq > SinceSeq. 0 = from genesis.
	OccurredAfter time.Time   // EXCLUSIVE: occurred_at > t. Zero = any.
	Limit         int         // 0 = unlimited
}

// ReadEvents streams every matching row in ascending seq order, calling yield once
// per row. yield returning ErrStopIteration ends the walk and ReadEvents returns
// nil. Any other non-nil error ends the walk and is returned wrapped. A negative
// Limit is ErrConfig, never "unlimited".
//
// The walk is PAGED: it materialises at most one bounded page (512 rows), and the
// pooled connection is released BEFORE ANY yield runs (INV-13). ReadEvents is
// therefore RE-ENTRANT — yield may call WithTx or ReadEvents again, which is the
// shape every projection reducer has (INV-35).
//
// Because pages are separate statements the walk is not a frozen snapshot: an event
// appended during the walk may be yielded by a later page. On an append-only ledger
// read in seq order that only ever means "you saw more".
func (s *Store) ReadEvents(ctx context.Context, f Filter, yield func(Record) error) error
```

One iteration shape is pinned deliberately. A second spelling (returning a slice, or an
`iter.Seq2`) would be a second contract to keep honest — the Task 2 lesson that a helper
accepting several spellings of one contract is a hole applies to the interface itself.

**Paging is a mechanism, not a contract.** The page size is not configurable and is not
observable: the rows a caller receives, and their order, are identical at any page size.
It exists because holding a pooled connection across `yield` makes the obvious consumer —
a reducer that writes derived state as it walks — deadlock: immediately at `MaxConns: 1`,
and unpredictably under load otherwise. Bounded memory is the side benefit; re-entrancy is
the point.

### 2.4 The projection handle

```go
// Tx is the projection-side handle: projections own their tables, store owns the
// connection. It runs under the least-privilege vera_projection role, so the ledger
// tables are readable and PHYSICALLY unwritable through it (§ 4, F7; INV-18).
type Tx struct{ /* unexported */ }

func (tx *Tx) Exec(ctx context.Context, sql string, args ...any) (rowsAffected int64, err error)
func (tx *Tx) Query(ctx context.Context, sql string, args ...any) (*Rows, error)
func (tx *Tx) QueryRow(ctx context.Context, sql string, args ...any) *Row

type Rows struct{ /* unexported */ }

func (r *Rows) Next() bool
func (r *Rows) Scan(dest ...any) error
func (r *Rows) Err() error
func (r *Rows) Close()

type Row struct{ /* unexported */ }

func (r *Row) Scan(dest ...any) error

// WithTx runs fn in one transaction as vera_projection: commit on nil, rollback on
// error, rollback-and-repanic on panic. This is how projections create and write
// derived state, and it is the ONLY generic SQL path store exposes.
func (s *Store) WithTx(ctx context.Context, fn func(context.Context, *Tx) error) error
```

`Tx` / `Rows` / `Row` are concrete types with unexported fields rather than interfaces:
they exist so that **no pgx type crosses the package boundary** (INV-15). Declaring an
interface at the provider would be the smell the repo's Go conventions warn about; a
concrete handle is the honest shape for "store owns the connection, you own the SQL".

`WithTx` is the reason the database-level role guard exists. Without it, the escape hatch
would let `projections` delete from `events` and "append-only" would be a claim, not a
property. The role makes Postgres — not a code review — refuse (§ 4, F7).

### 2.5 Errors

```go
var (
	ErrConfig        = errors.New("store: invalid config")
	ErrLocked        = errors.New("store: ledger is locked by another process")
	ErrClosed        = errors.New("store: store is closed")
	ErrLockLost      = errors.New("store: the ledger lock was lost")
	ErrStopIteration = errors.New("store: stop iteration")
	ErrMigrate       = errors.New("store: migration failed")
	ErrLedgerWrite   = errors.New("store: ledger tables are append-only")
	ErrKindConflict  = errors.New("store: same subject and payload with a different kind")
)

// LockedError names the holder so the CLI can say something true. PID and Since
// come from the INFORMATIONAL record (§ 2.1); an unreadable record leaves PID 0 and
// the refusal stands regardless. There is no Alive field: a flock dies with its
// holder. NOTE: it does NOT follow that a held lock is held by a LIVE process —
// F27 measured exactly that inference being false (a forked child's inherited
// descriptor holds a lock nobody owns). The absence of an Alive field survives that
// correction; the justification for it does not, so it is not restated here.
// Unwrap returns ErrLocked.
type LockedError struct {
	Path  string
	PID   int
	Since time.Time
}

func (e *LockedError) Error() string
func (e *LockedError) Unwrap() error
```

An invalid envelope surfaces as `core`'s error: `errors.Is(err, core.ErrInvalidEvent)` holds
and store adds no sentinel of its own for it (one home per datum).

`ErrLocked` and `ErrLockLost` are opposite ends of the same fact. `ErrLocked` is *somebody
else has it, you may not start*. It does NOT follow that the somebody is alive — F27
measured that inference false, and `acquireLock` re-probes before naming a holder.
`ErrLockLost` is *you had it and no longer do, and everything you were doing stops here* —
returned by `BeginSync`, `Append`, `ReadEvents`, and `WithTx` on a Store whose `LockPath`
no longer names the file it locked, and by `Close` when the same is true at release time.
A caller distinguishes them with `errors.Is`; neither is ever a message match.

Note what no message says any more: *remove the lock file by hand*. That was the escape
hatch from a lock file a crash could leave behind; under `flock` a dead holder's lock is
already gone, and deleting the file is the one action that can actually produce two
writers (INV-41).

### 2.6 The ledger schema (goose, `store/migrations/*.sql`, `embed.FS`)

Pinned objects — the from-empty test asserts exactly this set plus `goose_db_version` (INV-30):

| Object | Shape | Why |
|---|---|---|
| `events.seq` | `BIGSERIAL PRIMARY KEY` | **the sole replay order.** Gaps are expected and harmless (§ 4, F3). |
| `events.event_id` | `TEXT NOT NULL UNIQUE` | identity (ULID), never ordering. UNIQUE so a duplicated id is loud (INV-4). |
| `events.payload` | **`json`, not `jsonb`** | `json` preserves the canonical bytes byte-for-byte, so the stored `content_sha` still verifies against the stored payload; `jsonb` re-spaces them and breaks that (§ 4, F5). |
| `events.occurred_at` / `recorded_at` | `TIMESTAMPTZ NOT NULL` | round-trips exactly at core's microsecond-UTC normalization (§ 4, F9). |
| `events_idempotency` | `UNIQUE (source, native_id, content_sha)` | the idempotency key. A repeat with a NEW `content_sha` is a legitimate revision row. |
| `events_source_kind_seq` | `INDEX (source, kind, seq)` | serves every `Filter` shape in seq order. |
| `events_occurred_at` | `INDEX (occurred_at)` | serves `OccurredAfter`. |
| `sync_runs` | `(id, connector, cursor_json, started_at, finished_at, events_appended, error)` | observability, never correctness. |
| role `vera_projection` | `NOLOGIN`; `SELECT` on the ledger, `CREATE` on the schema | the `WithTx` guard (§ 4, F7). |

Migration SQL is generate-once / append-only; a landed migration is never hand-edited. That
bypass gets its row in [docs/gates.md](../../../docs/gates.md) "Known accepted bypasses" in
the Task 3 commit.

---

## 3. Invariants

Each is a testable statement; the mapping to test names is § 5. Every one is expected to be
**mutation-proven** — break the implementation, watch the named test fail, restore (the Task 2
lesson: a green suite proves nothing on its own).

Invariant numbers are **stable identifiers assigned in the order they were added**, and this
section is grouped thematically, so numbering is not contiguous within a group. A number is
never reused and never renumbered: it is cited from code comments, tests, and review notes.

**Append and idempotency**

1. **INV-1 — First append inserts.** A new event returns `inserted=true`, `Record.Seq > 0`,
   and exactly one matching row exists.
2. **INV-2 — Re-ingest is a no-op that returns the ORIGINAL row — the row for THAT
   content, not the newest one for the subject.** Appending the identical event again
   returns `inserted=false`, leaves the row count unchanged, and the returned `Record`
   carries the matching row's `seq` and `event_id` — never the rejected candidate's, and
   never a sibling revision's. Re-appending a revision that has since been SUPERSEDED
   returns the superseded row itself.
   *This is the fallback SELECT: `ON CONFLICT … DO NOTHING … RETURNING` returns no row at
   all (§ 4, F3), so without the fallback the caller gets a zero Record. The
   superseded-revision half is what pins the lookup to the full idempotency tuple: with
   one row per subject, a fallback matching on `(source, native_id)` alone is
   indistinguishable from the correct one — measured, `ORDER BY seq DESC LIMIT 1` with no
   `content_sha` predicate passed the entire suite.*
3. **INV-3 — Revision semantics.** Same `(source, native_id)` with a new `content_sha`
   inserts a NEW row at a strictly higher `seq`; the earlier row is byte-unchanged and both
   remain readable.
4. **INV-4 — The conflict target is the idempotency index only.** An event whose `event_id`
   duplicates an existing row but whose idempotency tuple is new is REJECTED with an error
   (`23505` surfaced), never silently absorbed. *A bare `ON CONFLICT DO NOTHING` swallows it
   — measured, § 4 F6.*
5. **INV-5 — Validate before the round trip.** A hand-built invalid `core.Event` (e.g. an
   unregistered kind) is refused with `core.ErrInvalidEvent`, nothing is written, and **no
   sequence value is consumed** — the next successful append's `seq` is the previous plus one.
6. **INV-6 — `events_appended` is derived, not reported; it NEVER exceeds the rows the run
   inserted; and a finished run accepts nothing further.** No caller-supplied count exists.
   Once `Finish` has written the count, `Append` on the run is refused — an accepted append
   would make the number a lie about a ledger that has moved on, with nothing in the journal
   to say so.
   **The bound is one-directional, and that is a decision, not an oversight (§ 4, F32).**
   The count equals the number of appends the CLIENT OBSERVED as inserted. It can be LOWER
   than the rows this run actually put in the ledger: a context deadline expiring after
   PostgreSQL commits but before pgx returns leaves a row whose fate the caller never
   learned (measured, 202 of 600 deadline trials). It is never HIGHER. Recovering the exact
   count is not possible with what the ledger records — see F32 — so the journal is
   deliberately allowed to understate real work and forbidden from inventing any. An earlier
   version of this invariant claimed the equality; it was false in one direction, then the
   repair made it false in the other, which was worse.
7. **INV-7 — Payload bytes survive the ledger.** The payload read back is byte-identical to
   the canonical bytes written, and `sha256` of it equals the stored `content_sha`.
8. **INV-8 — Timestamps survive the ledger, and every `Record` is normalised to UTC.**
   `occurred_at` and `recorded_at` read back are `Time.Equal` to what was written, and the
   `Record` returned by `Append` and by `ReadEvents` carries them with `Location() ==
   time.UTC`. *Compared with `Equal`, never `==` or `DeepEqual`: pgx returns a `Local`-zone
   time for the same instant (§ 4, F9) — which is also why normalising matters, or the
   machine's timezone leaks into every report built on a Record.*
9. **INV-9 — One kind per subject (core's delegated constraint, closed).** A re-append whose
   `(source, native_id, content_sha)` matches an existing row but whose `kind` differs returns
   `ErrKindConflict` instead of a silent `inserted=false`. *core's SPEC § 6 hands store this
   hole because `content_sha` covers the payload only; the fallback SELECT already reads the
   existing row, so the check is free and the index stays exactly as the panel verified.*

**Ordering and the read seam**

10. **INV-10 — `seq` is the sole replay order.** `ReadEvents` yields strictly increasing
    `Seq`, and a ledger containing gaps (from absorbed conflicts and rolled-back
    transactions) still replays in strictly increasing order — **including when the
    table's PHYSICAL row order disagrees with `seq`** (§ 4, F12). Storage layout is not
    part of the contract; `ORDER BY seq` is.
11. **INV-11 — Order is independent of identity and of wall-clock.** Events appended with
    DESCENDING `event_id`s and DESCENDING `occurred_at` values still replay in append order.
    *This is the anti-trap test: it fails the moment anyone sorts by ULID or by timestamp.*
12. **INV-12 — Filter semantics, one row per field.** `Source` and `Kind` are exact matches;
    `SinceSeq` is EXCLUSIVE; `OccurredAfter` is EXCLUSIVE; `Limit` caps the yield count; the
    zero `Filter` returns everything from genesis; a NEGATIVE `Limit` is `ErrConfig`.
    *Negative is refused rather than clamped because `0` already means unlimited: passed
    through it reaches SQL as `LIMIT -1`, which PostgreSQL reads as no limit at all — a
    caller whose arithmetic underflowed asks for less than nothing and silently gets
    everything.*
13. **INV-13 — Streaming, early exit, no leak.** `yield` returning `ErrStopIteration` ends the
    walk and `ReadEvents` returns nil; any other error aborts and is returned wrapped
    (`errors.Is` finds it). With `MaxConns: 1`, a second `ReadEvents` after an early exit
    succeeds within a short deadline — proving the connection went back to the pool.

**The append-only surface (mechanically enforced, not asserted in prose)**

14. **INV-14 — Exported surface is exactly the pinned set.** A `go/parser` scan of store's
    non-test files finds exported identifiers equal to the § 2 list and nothing else — in
    particular nothing matching `Update|Delete|Remove|Truncate|Purge|Rewrite` and no exported
    way to set `Seq`. A new export is a spec diff first.
15. **INV-15 — No driver escapes.** The same scan finds no exported signature (parameter,
    result, or field) mentioning `pgx`, `pgxpool`, `pgconn`, `database/sql`, `goose`, or
    `embeddedpostgres`. Connectors, projections and cli therefore cannot import a driver
    through store's surface.
16. **INV-16 — No mutating SQL against `events`, anywhere.** A scan of every `.go` and `.sql`
    file in the package finds no `UPDATE`, `DELETE`, `TRUNCATE`, or `DROP` naming `events`;
    `sync_runs` permits exactly ONE `UPDATE` (the `Finish` statement, matched literally) and
    no `DELETE` or `TRUNCATE`.

**The projection handle**

17. **INV-17 — Transaction discipline.** `WithTx` commits when `fn` returns nil, rolls back
    when it returns an error (the error is returned wrapped), and rolls back then re-panics
    on a panic — leaving no open transaction behind in any case.
18. **INV-18 — The ledger is read-only to projections, enforced by Postgres.** Inside
    `WithTx`: `INSERT` / `UPDATE` / `DELETE` / `TRUNCATE` against `events` or `sync_runs` fail
    with `ErrLedgerWrite`; `SELECT` on them succeeds; `CREATE TABLE` of a projection table
    succeeds. *Measured: `42501 permission denied` (§ 4, F7).*
19. **INV-19 — No projection DDL in the migration stream.** The embedded migrations contain
    only the § 2.6 objects; after `Open` on an empty Root the database holds exactly those
    tables plus `goose_db_version` — no projection table exists until `projections` creates it.

**The single-data-dir lock**

20. **INV-20 — Exclusivity.** With a `Store` open, a second `Open` on the same Root fails with
    `ErrLocked` and a `LockedError` naming the holder's pid, and the data directory is
    unmodified (file set and mtimes unchanged).
21. **INV-21 — A second OS process exits non-zero and touches nothing.** The Task 3 DoD
    stated literally: a subprocess running a real ledger command against a held Root exits
    non-zero, prints the holder, and leaves the data directory byte-identical.
22. **INV-22 — RETIRED (2026-08-09, round 3).** Stale takeover required both an aged
    lock and a dead holder, with a compare-and-swap on the observed identity. There is
    no takeover any more: the kernel releases a dead holder's lock (INV-40), so there is
    nothing stale to find and nothing to swap. *Its own proving test was also the
    clearest instance of the defect class this spec now names explicitly: the marker it
    planted with `os.WriteFile` was a shape the code — which used `os.Link`, sharing the
    source inode — never created, so the test passed while the CAS did not mutually
    exclude.*
23. **INV-23 — `EPERM` means ALIVE.** A liveness probe that returns `EPERM` (the pid exists
    but belongs to another user) counts as alive; only `ESRCH` counts as dead.
    *Measured, § 4 F8.* **AMENDED (round 3):** the lock no longer probes a pid for
    anything, so this is re-pointed at its one surviving consumer — postmaster adoption
    (INV-28). Reading an unsignalable postmaster as dead would start a SECOND server
    against a data directory another one is serving. The invariant is unchanged; only
    the surface that exercises it moved.
24. **INV-24 — RETIRED (2026-08-09, round 3).** The heartbeat refreshed the lock's mtime
    so a long operation was not judged stale. Nothing judges staleness now, so nothing
    needs refreshing; the heartbeat goroutine is deleted rather than kept as decoration.
    *It was also the mechanism behind the round-2 failure: ownership was re-read only on
    the beat, so detection latency equalled `HeartbeatInterval` — 60s as shipped.*
25. **INV-25 — Close releases, and releases ONLY what is ours.** After `Close`, another
    process can take the lock and a fresh `Open` succeeds; a second `Close` is a no-op
    and every other method returns `ErrClosed`. **AMENDED (round 3):** release is
    closing the descriptor, and the lock FILE deliberately REMAINS on disk. Unlinking on
    release is a self-inflicted two-holder window — between dropping the lock and
    unlinking the path another process can open and lock that inode, and the unlink then
    hands the path to a third. So "released" is proven by a competitor being able to
    acquire, never by the file's absence. A file that is no longer ours is likewise never
    deleted and never reported as a clean release (INV-37). **AMENDED (F4 round):** release
    REPORTS a loss; it does not detect one. Teardown settles ownership FIRST because it has
    to — it cannot decide whether this process still owns the postmaster (INV-43) without
    the answer — so a second identity check inside release could never fire, and a check
    that can never fire is a check no mutation can prove. There is ONE detector, sited
    where the answer is needed.
26. **INV-26 — `DatabaseURL` mode takes no data-dir lock and starts no server.** No lock file
    is created and no Postgres child process is spawned.

**Embedded server and configuration guards**

27. **INV-27 — The wipe-trap guards are bidirectional and cover the lock.** `Open` returns
    `ErrConfig` when `DataDir` and `RuntimeDir` are nested IN EITHER DIRECTION, when
    `LockPath` is inside `RuntimeDir`, and for an empty `Root`. **AMENDED (F1 round):**
    deriving the lock from `DataDir` (INV-42) narrowed the `LockPath`-in-`RuntimeDir`
    clause without retiring it. A canonical lock UNDER `RuntimeDir` would require
    `DataDir` under `RuntimeDir`, which the first clause already refuses; what remains
    reachable is a `RuntimeDir` that IS the lock file's path (`RuntimeDir=<root>/db.lock`
    beside `DataDir=<root>/db`), which is neither nesting and which `RemoveAll` would take
    on every start. *That case is proven — deleting the clause fails
    `TestOpen_RejectsUnsafeConfig/RuntimeDir_IS_the_derived_lock_path`.* *This turns the measured
    silent-wipe trap (§ 4, F2) into a mechanical refusal. The guard is bidirectional
    because the one-directional version permitted `RuntimeDir` under `DataDir`, which is
    the same `RemoveAll` taking ledger files on its way out; and it covers `LockPath`
    because a lock the next start deletes lets a second process take a lock of its own
    and adopt this ledger's postmaster — the two-writer failure reached by configuration
    instead of by a race.* **AMENDED (round 3):** the
    `HeartbeatInterval >= StaleLockAfter` clause is gone with both fields — there is no
    lock timing left to get wrong.
28. **INV-28 — Orphan adoption and its limit.** When a live postmaster already serves this
    `DataDir`, `Open` adopts it instead of starting a second one, and `Close` shuts it down
    (self-healing after a crash) — unless this Store lost its lock, in which case the server
    is not its to stop (INV-43). When the port is busy but no live `postmaster.pid` belongs
    to this `DataDir`, `Open` fails with a message naming the port — it never adopts a foreign
    process. **AMENDED (F4 round):** "busy" now means *stays* busy. A single refused bind is
    not evidence about a port: a closed listening socket is still held by any sibling sitting
    between `fork()` and `exec()`, at a measured 4.2% under fork pressure (§ 4, F24), so both
    the probe and the start ride that window out before refusing. A port a foreign process
    genuinely holds never comes free, so the refusal is unchanged in substance — only in
    patience.
29. **INV-29 — Two roots coexist.** Two `Store`s on different Roots are open simultaneously,
    because the default port is derived from `DataDir` (§ 4, F10).

**Migrations and lifecycle**

30. **INV-30 — Migrations apply from empty.** `Open` against a fresh temp Root creates exactly
    the § 2.6 object set plus `goose_db_version`, over the store's own pool, with no goose
    package-level global touched.
31. **INV-31 — Re-open applies nothing and preserves everything.** A second `Open` on the same
    Root applies zero migrations and every previously appended row is still readable at its
    original `seq`. *This is the positive proof of the F2 trap: the naive configuration
    silently wipes here.*
32. **INV-32 — Migration failure is loud and leaves no Store; and a failed Open never
    strands the lock.** Two claims, proven at two different failure sites, because store
    exposes no seam that makes a MIGRATION fail in embedded mode:
    **(a)** a migration error surfaces as `ErrMigrate` and `Open` returns a nil `*Store` —
    proven in `DatabaseURL` mode, where a test can pre-poison the target database, and
    where **no lock is taken at all**, so this half says nothing about the lock;
    **(b)** an Open that fails AFTER the lock is taken releases it — proven in embedded
    mode at two post-lock failure sites (an occupied port, INV-32's own test; a foreign
    data-dir version, INV-33's).
    *The gap is stated rather than papered over: (b) is a PROXY for the migration step. No
    test proves lock release for a migration failure specifically. Closing it needs either
    a pinned DSN for the embedded server or an unexported test hook — a spec diff, not a
    test workaround.*

**Hardening added at review (2026-08-09).** Numbers continue the sequence; the grouping
above is thematic (see the note at the top of this section).

33. **INV-33 — A data directory written by another PostgreSQL major is refused.** `Open`
    returns `ErrConfig` naming the on-disk version, before `initdb` can run, and releases
    the lock. *embedded-postgres answers a `PG_VERSION` mismatch by wiping the directory
    and re-initialising it, so a dependency bump that moves the default major would
    silently empty the ledger and report success — the § 4 F2 amnesia trap arriving
    through the dependency graph, where no diff says "delete the ledger".*
34. **INV-34 — `Finish` reports success only after the journal row is written.** A refused
    write is a failed `Finish`, and a retry after the obstacle clears actually writes:
    `finished_at` is non-NULL and `events_appended` is the derived count. *Flag-then-write
    turns a failed write into a silent success — the retry finds the flag set, returns
    nil, and the run stays unfinished forever while every report says it completed.*
35. **INV-35 — The read seam is re-entrant, and paging is invisible.** `yield` may call
    `WithTx` or `ReadEvents` again — at `MaxConns: 1`, within a bounded deadline. A walk
    LONGER than one page yields every matching row exactly once in seq order, for an
    unbounded walk, for a mid-ledger `SinceSeq`, and for `Limit` values on either side of
    and exactly at a page boundary. *A three-row fixture cannot see a keyset off by one;
    only a fixture larger than a page can. Holding the connection across `yield` deadlocks
    the one consumer that matters — a reducer writing derived state as it walks.*
36. **INV-36 — One appender at a time within a process.** With `events` held under an
    `ACCESS EXCLUSIVE` lock so every INSERT that reaches PostgreSQL parks visibly, at most
    ONE append is in flight at a time however many goroutines are appending; all of them
    then complete, each at its own `seq`. *§ 6 rests `SinceSeq`'s soundness on single-writer
    discipline — measured, `seq` assignment order is not commit order, so of two
    overlapping appends the one holding `seq=1` can commit after the one holding `seq=2`
    and a reader that stored cursor 2 never sees row 1. The data-dir lock enforces that
    between processes; nothing enforced it within one, so the claim the read seam depends
    on rested on a discipline no code asserted.*
37. **INV-37 — Losing the lock poisons the Store.** When the file at `LockPath` stops
    being the file this process locked, the lock is LOST: every subsequent `BeginSync`,
    `Append`, `ReadEvents`, and `WithTx` fails with `ErrLockLost`, `Close` reports the
    loss instead of success, and a file that is not ours is never deleted.
    *Measured (§ 4, F14): with the holder merely logging a warning, a `git clean -xdf` under
    a running command let a second process take the lock 1.0 s later, adopt the first's
    postmaster, and STOP THE DATABASE the first was still appending to — whose own Close
    then reported success.* **AMENDED (round 3):** the trigger is file IDENTITY
    (`os.SameFile`), not the record's bytes — overwriting the informational record in
    place is NOT a loss, because the flock is unchanged and the ledger is still ours.
    *A byte comparison would poison a working Store for a scribble.*
    **AMENDED (F4 round):** `Close` DETECTS the loss itself rather than reporting a flag
    some earlier operation happened to set. Every other lock test appends or reads before
    closing, so the teardown-side check was never the detector in any test that ran and
    deleting it survived the whole suite. It is now proven by a Store that is opened,
    dispossessed, and closed with NO operation in between — which is also a real sequence:
    a `git clean` under an idle command.
38. **INV-38 — RETIRED (2026-08-09, round 3).** An unreadable or empty lock record used
    to need an age rule: it named no holder, calling it alive made it immortal, calling
    it dead let a starting process be robbed. The record is informational now — an
    unreadable one costs a name in an error message. Nothing about it can wedge or
    release a lock, so there is no rule to state.

**Round 3 — the flock rebuild (2026-08-09).** Numbers continue the sequence.

39. **INV-39 — The lock is an exclusive, NON-BLOCKING flock on `LockPath`.** While a
    Store is open, an outside `open`+`flock(LOCK_EX|LOCK_NB)` on that exact path fails
    with `EWOULDBLOCK`, and a second `Open` — in this process or another — is refused
    with `ErrLocked` promptly rather than waiting. *Three one-line mistakes break this
    and each is caught separately: locking a different descriptor leaves `LockPath`
    takeable; `LOCK_SH` admits two holders; dropping `LOCK_NB` makes a busy ledger hang a
    command instead of reporting who holds it (§ 4, F20). The same-process half is free
    rather than engineered — flock belongs to the open file description (§ 4, F19) — but
    it is asserted, because it is the guarantee `appendMu` alone cannot give across two
    `Store`s in one process.*
40. **INV-40 — The kernel releases the lock when the holder dies, however it dies.** A
    holder killed with `SIGKILL` — running no cleanup at all — leaves its lock file on
    disk, fresh, naming a now-dead pid; the next `Open` succeeds anyway. *This is the
    invariant that pays for the deletions: staleness, takeover, reclaim markers, the
    ownership nonce and the lock's pid probe all existed to answer "did the process that
    wrote this file die?", and the kernel answers it by construction (§ 4, F17). The test
    asserts the file is present, fresh, and names a dead pid precisely so that any
    reintroduced age-or-liveness rule fails it. It also proves the postmaster does not
    inherit the lock descriptor: `pg_ctl` daemonizes, so a descriptor crossing `exec`
    would leave the flock held by a server nobody reaps and the ledger permanently
    unopenable (§ 4, F21).*
41. **INV-41 — Identity is checked at the OPERATION choke point AND again inside the append
    critical section, so a lost lock is caught before the next operation and an in-flight
    one is bounded by a single write round trip.** `flock` lives on the inode, so removing
    or replacing `LockPath` admits a second legitimate holder (§ 4, F18). Two checks, both
    `os.SameFile(fstat(held fd), stat(LockPath))` and both uncached:
    - `Store.conn` — the single path to the database for `BeginSync`, `Append`,
      `ReadEvents` and `WithTx` — runs it before an operation may START.
    - `Sync.Append` runs it AGAIN after taking `appendMu` and immediately before issuing
      the INSERT.

    What follows, and what is claimed: **no ledger operation that reaches an identity
    check after the loss succeeds.** The first `Append`, `BeginSync`, `ReadEvents` or
    `WithTx` attempted after the loss fails with `ErrLockLost`, and an append that was
    merely QUEUED behind `appendMu` when the lock went is refused rather than admitted.

    **The residual, stated rather than hidden: this is not zero and cannot be made zero.**
    No transaction spans the filesystem lock and the PostgreSQL write, so an append whose
    INSERT is already in flight cannot be recalled — at most ONE append, the one between
    its own re-check and its commit, can land after the loss. That window is one write
    round trip — bounded in COUNT (at most one), **not in duration**. Uncontended that trip is fast (p50 **142.75µs**, p99 **223.7µs**, max **3.19ms** over 400 appends), but an ordinary table-level lock (`VACUUM FULL`, `ALTER TABLE`, `CLUSTER`, a conflicting transaction) stretches it to SECONDS: measured **4.059s** with the ledger held under `ACCESS EXCLUSIVE`. Do not read the microsecond figures as the bound; the bound is one append
    (§ 4, F22).

    *The previous wording — "the count of operations that succeed after the loss is ZERO"
    — was measured FALSE: 8 appends succeeded, the last of them 4.012s after the loss,
    interleaved with the new legitimate holder's rows. The cause was structural. `Append`
    ran its only ownership check in `conn()` BEFORE `appendMu`, so every appender already
    past that point completed regardless, and the exposure was the whole queue plus any
    PostgreSQL wait — measured growing with depth: 394µs at 4 concurrent appenders, 822µs
    at 8, 1.42ms at 16, 6.04ms at 64, and unbounded whenever the queue is held open, which
    is what produced the 4.012s (§ 4, F22). A spec that overclaims is itself the defect:
    the absolute was unreachable, so it hid the real bound instead of establishing it.*

    *This also remains round 2's finding answered directly: there, ownership was re-read
    only on the heartbeat tick and every operation between ticks trusted a cached flag, so
    at the shipped 60s default a lost lock was invisible for up to a minute and two
    processes appended interleaved for 2.6 seconds (§ 4, F14b). The check is two syscalls
    at 1.5µs against a database round trip three orders of magnitude larger (§ 4, F18), so
    there is no cost that would justify a TTL — and a TTL is precisely the property that
    failed.*

    *`WithTx` is deliberately NOT re-checked mid-transaction, and the claim above is scoped
    accordingly: a projection transaction runs as `vera_projection`, which the database
    refuses to let write the ledger at all (INV-18), so its exposure is derived state on a
    connection it already holds — not a second writer.*
    **The acquisition-side half of this guard is NOT proven by a test, and is stated
    rather than claimed:** `acquireLock` re-checks the same identity between its `open`
    and its `flock`, because the path can be replaced in that window and an exclusive
    lock on an orphan inode is indistinguishable from holding the ledger. Forcing that
    interleaving needs a seam in the acquire path that exists only for the test — a spec
    diff, not a test workaround. Removing the check is a mutation that SURVIVES the
    suite. Its residual exposure is bounded: with the check absent, the first `conn()`
    still catches it, so only server start and migrations would run on a lock that was
    never really held.
    **Re-confirmed 2026-08-09 (F1-F4 round).** A mutation sweep over the lock-safety
    surface after this round's changes re-ran it and it still survives — it is the ONLY
    survivor left there. The `noteLost` memoisation is NOT among the mutants a test catches: round 7 proved it SURVIVED, and round 8 confirmed the code was unreachable and deleted (see § 5 and `lock.go`'s noteLost precondition note). The window is two adjacent syscalls, so
    the only ways to close it are a production-only test hook or a stochastic churn test —
    and a stochastic test is the wrong trade in a suite whose gate has to be green every
    time. It stays stated.

42. **INV-42 — The lock is DERIVED from the data directory it protects, so two Stores
    cannot disagree about which lock guards a ledger.** `LockPath` resolves to
    `<DataDir>.lock` — beside the data directory, named after it. `LockPath` supplied in
    `Config` is an ASSERTION, compared by `samePath`, which resolves symlinked ANCESTORS
    but deliberately NOT the final component (INV-46). It fails closed: a spelling that
    reaches the canonical lock by another route — a differing case on a case-insensitive
    filesystem, or a symlink pointing AT the canonical lock — is refused, because the
    caller must NAME the derived lock rather than merely arrive at it. Resolving the leaf
    was worse than either rule: the verdict then depended on whether the canonical lock
    happened to exist yet (refused before first `Open`, accepted after), so one
    configuration got two answers from filesystem state the caller does not control;
    because two spellings of one path are one lock); anything that does not resolve to the
    canonical path is `ErrConfig` at `Open`, before any lock is taken, any server started
    or adopted, and any directory created. *Measured defect: `resolve` related `DataDir` to
    `RuntimeDir`, and `LockPath` to `RuntimeDir`, but NOTHING related `LockPath` to
    `DataDir`. Two Stores pointed at ONE data directory with two DIFFERENT lock paths both
    acquired, both adopted the same postmaster, and both appended — 40 interleaved rows,
    no race required. That is the two-writer failure INV-27 already names, reached by
    configuration instead of by a wipe.* The refusal is `ErrConfig` and not `ErrLocked` on
    purpose: the configuration is wrong, and "somebody else holds it" would invite the
    operator to try a third path. *The field is kept rather than deleted so the mistake
    stays expressible and therefore provable: with the guard removed the proving test
    opens a second Store on one data directory and says so.*

43. **INV-43 — A Store that LOST its lock does not stop the server, because the server is
    no longer its to stop.** Teardown settles ownership before it touches anything: a
    dispossessed Store closes its pool and its descriptor, reports `ErrLockLost`, and
    leaves the postmaster to whichever process holds the lock now. A holder's `Close`
    still stops the server it started or adopted (INV-28) — the asymmetry IS the
    invariant. *Measured: A lost the lock, B took over and appended 90 rows, A's `Close`
    stopped the postmaster, and B's next 274 operations failed with "terminating
    connection due to administrator command" (SQLSTATE 57P01) — while B's own `Close`
    returned nil, because B holds the lock and has no way to know its database was shot
    from under it. That is round 1's catastrophe with the roles reversed.* The server left
    running is not a leak: the next `Open` adopts it (INV-28), which is exactly how the
    suite cleans up after these tests.

44. **INV-44 — Every comparison between two configured paths compares FILES, not
    spellings.** `normalize` resolves the longest existing prefix through symlinks and
    rejoins the rest; `lockPathFor`, `samePath`, `within` and `enclosingCluster` all go
    through it. A lexical comparison answers about the SPELLING, and these guards must
    answer about the directory: with `PG_VERSION` in `outer/` and `db -> outer/inner`,
    the lexical `enclosingCluster` refused `outer/inner` and accepted `db` — the same
    directory (§ 4, F26/F29).
45. **INV-45 — The pre-initdb port guard rides out a transiently-bound port** rather than
    returning on its first probe, and it probes BOTH loopbacks: a server bound only to
    `[::1]` is a real squatter, and an IPv4-only probe would let initdb run against a
    port in use (§ 4, F24).
46. **INV-46 — `samePath` resolves ancestors but NOT the leaf, so a LockPath verdict is
    TOTAL** — identical whether or not the canonical lock exists yet. Resolving the leaf
    made the answer depend on filesystem state the caller does not control (refused
    before the first `Open`, accepted after). Two spellings of one path are one lock, so
    a lock named through a symlinked parent is accepted; a DIFFERENT file that merely
    points at the canonical lock is refused (§ 4, F26).
47. **INV-47 — EWOULDBLOCK from `flock` does not imply a live foreign holder, and
    `acquireLock` re-probes before reporting one.** The lock belongs to the open file
    description and `fork()` duplicates descriptors, so a child forked while this
    process held the lock keeps the description alive past our own release: the kernel
    refuses a lock nobody logically holds. The re-probe is bounded by
    `lockContendedWindow`; it can only turn a false "locked" into a correct acquire,
    never hand the lock to a second writer, because acquisition is still granted by
    `flock` and by nothing else (§ 4, F27).
48. **INV-48 — A row is never counted twice, and never counted by a run that did not insert
    it.** The run mutex is held across the check, the INSERT and the increment, so an
    `Append` racing a `Finish` resolves one way or the other rather than both (§ 4, F28).
    On the error path the run counts NOTHING: it cannot tell whether its own INSERT created
    a row that is now present, so it declines to guess (§ 4, F32). Together with INV-6 this
    gives the guarantee that matters — the journal may understate, never overstate.
49. **INV-49 — A lost lock is reported at ERROR level.** The level is asserted by a test,
    not left to review, because the regression is a one-word edit and the original
    design's failure mode was exactly a complaint nobody saw (§ 4, F27).
50. **INV-50 — Teardown does not run under the lifecycle mutex.** `Close` marks the Store
    closed and detaches its pool, server and lock under `s.mu`, then releases them outside
    it. Holding the mutex across `pool.Close()` deadlocks against any callback that
    re-enters the Store while holding a connection, permanently and without a timeout
    (§ 4, F31). The detach is what makes the release safe: nothing can reach a resource
    that is no longer on the Store.
51. **INV-51 — A projection rollback does not depend on the caller's context.**
    `WithTx`'s deferred rollback runs on `context.WithoutCancel(ctx)`, so it still
    completes when the caller's context is exactly what failed. The guarantee is narrow and
    worth stating precisely: pgxpool destroys a connection returned mid-transaction, so a
    missed rollback does not poison the pool — what this invariant buys is that the
    rollback SUCCEEDS instead of failing on a dead context (§ 4, F29's note on asserting
    the right consequence).

---

## 4. Measured platform facts (the Task 3 re-confirmation)

**Environment:** darwin/arm64 (macOS 26.5.2 / Darwin 25.5.0), Go 1.26.2,
`embedded-postgres v1.34.0` (PostgreSQL **18.3.0**, its default), `pgx/v5 v5.10.0`,
`goose/v3 v3.27.3`. F1-F13 measured 2026-08-08; F14/F15 during the 2026-08-09 review;
**F14b and F16-F21 on 2026-08-09** for the flock rebuild. Every one was taken with a
scratch program outside the repo. A differing result on another platform is a fact to
re-measure, not a number to copy — and F16-F21 are POSIX-family behaviour this package
already depends on elsewhere (`syscall.Kill` in the postmaster probe), so `flock` adds no
portability constraint that was not already present.

**F1 — Startup cost is 26× worse in the obvious configuration, and the plan's "~1–2s per
command" was wrong in both directions.**

| Configuration | Open (start) | Close (stop) | Notes |
|---|---|---|---|
| cold, first ever | 12.36s | 0.13s | downloads a 33 MB `.txz` from `repo1.maven.org`, extracts 144 MB, runs `initdb` |
| fresh data dir, binaries already extracted | 3.6–4.0s | 0.15s | `initdb` + `createdb` |
| **library default `BinariesPath`** (= `RuntimePath`) | **5.43 / 5.50 / 5.43s** | 0.13s | `Start()` `RemoveAll`s `RuntimePath` then re-extracts 144 MB **every single start** |
| **persistent `BinariesDir` outside `RuntimeDir`** | **0.203 / 0.203 / 0.211 / 0.212s** (first 0.369s) | 0.130–0.134s | extraction skipped: `Start()` skips it when `<binariesPath>/bin/pg_ctl` exists |

Opening the pgx pool and running the first query adds 0.018–0.021s. So a warm command pays
**~0.35s** for open+close, versus ~5.6s in the naive configuration. `vera verify` syncs twice
and rebuilds; at 5.6s per open the plan's own revisit trigger (`make check` > ~30s) fires on
startup alone. Hence `BinariesDir` is a first-class Config field, not an optimisation.

**F2 — The library's default data path silently destroys the ledger on every start.**
`Start()` runs `os.RemoveAll(runtimePath)` unconditionally, and when `DataPath` is unset it has
already defaulted to `runtimePath/data`. Measured: wrote a table, stopped, started again — the
table was **gone, with no error**; `initdb` had re-run. A flight recorder configured this way
is permanently amnesiac and never says so. `DataDir` must live outside `RuntimeDir`, and
INV-27 makes `Open` refuse anything else.

**F3 — `ON CONFLICT DO NOTHING … RETURNING` really returns no row, and BIGSERIAL burns values.**
Measured with `pgx`: on conflict `QueryRow(...).Scan` returns an error satisfying
`errors.Is(err, pgx.ErrNoRows)` — so **the fallback SELECT is required**, and it returns the
ORIGINAL row (`seq=1`, `event_id=e1`), not the rejected candidate. Sequence behavior across
five appends (insert, duplicate, revision, duplicate, new subject):

```
seq=1  first insert
       duplicate            -> no row (sequence value 2 consumed and discarded)
seq=3  revision (new sha)
       duplicate            -> no row (value 4 consumed and discarded)
seq=5  fresh subject
```

A rolled-back transaction burns one too (took 6, the next append got 7). A window-function
check over the committed rows confirmed `seq` is still **strictly increasing** — gaps are
expected, order is what matters, and nothing may treat `seq` as a count.

**F4 — A crashed command leaves a LIVE Postgres behind: the lock cannot be the only guard.**
`pg_ctl start` daemonizes; the postmaster is **not** a child of the Go process. `kill -9` on
the CLI left `postgres` listening and fully serving queries, still holding the data directory
with a live `postmaster.pid`. So the next run can meet a *stale lock file* and a *live server*
at the same time — which is exactly why `Open` adopts an orphan that belongs to this `DataDir`
and `Close` shuts it down (INV-28), rather than failing forever until someone runs `kill` by
hand.

**F5 — `json` preserves the canonical bytes; `jsonb` destroys them.**

```
wrote      : {"a":1,"n":1.5,"z":"café x","zz":[3,1,2]}
json  back : {"a":1,"n":1.5,"z":"café x","zz":[3,1,2]}     byte-identical, sha matches
jsonb back : {"a": 1, "n": 1.5, "z": "café x", "zz": [3, 1, 2]}   NOT identical, sha differs
```

The plan assumed `jsonb` and worked around it ("never re-derive `content_sha` from stored
JSONB"). Choosing `json` is strictly stronger: the ledger stays self-verifying — the stored
payload still hashes to the stored `content_sha` (INV-7). core's INV-23 is unaffected: core
still never re-derives.

**F6 — A bare `ON CONFLICT DO NOTHING` silently swallows FOREIGN unique violations.** With a
`UNIQUE(event_id)` alongside the idempotency index:

| Statement | Duplicate `event_id`, new subject |
|---|---|
| `ON CONFLICT DO NOTHING` (the plan's literal wording) | **absorbed silently — no row, no error** |
| `ON CONFLICT (source, native_id, content_sha) DO NOTHING` | `23505`, `constraint=…_event_id_key` — loud |

The targeted form is pinned (INV-4). This is the plan's third correction.

**F7 — Postgres can make the ledger append-only for projections.** Inside a transaction
running `SET LOCAL ROLE vera_projection`:

```
SELECT from ledger            ALLOWED
DELETE / UPDATE / INSERT / TRUNCATE   BLOCKED  42501 permission denied
DROP TABLE                    BLOCKED  must be owner of table
CREATE own projection table   ALLOWED
```

The session user (superuser) is unaffected outside `SET LOCAL ROLE`, so `Append` still works.
This is what turns "append-only" from a claim about the Go surface into a property of the
database (INV-18).

**F8 — Liveness probe semantics.** `syscall.Kill(pid, 0)`: `nil` ⇒ alive; `ESRCH` ⇒ dead;
**`EPERM` ⇒ ALIVE but owned by another user** (measured against pid 1). `os.FindProcess`
always succeeds on unix and is **not** a liveness probe. Only `ESRCH` may be read as death
(INV-23). *This measurement originally justified the lock's takeover rule; after the round-3
flock rebuild the lock probes no pid at all, and the surviving consumer is postmaster
adoption (INV-28) — where reading an unsignalable postmaster as dead would start a second
server against a live data directory.*

**F9 — `timestamptz` round-trips exactly, but the zone changes.** `2026-08-08T12:34:56.123456Z`
returned identically — core's microsecond-UTC normalization (its INV-19) is exactly the right
precision. pgx returns the value with `Location() == Local`, so tests must compare with
`Time.Equal`; `==` or `reflect.DeepEqual` on the struct fails on a value that is the same
instant (INV-8).

**F10 — Collision behavior of a second process, all three variants — every one failed CLOSED,
none corrupted anything.**

| Second process | Outcome | Latency |
|---|---|---|
| same data dir, **same port** | `process already listening on port 55432` — raised by `ensurePortAvailable`, the FIRST statement of `Start()`, before any filesystem touch | 3 ms |
| same data dir, different port, separate runtime dir | Postgres itself refuses: `FATAL: lock file "postmaster.pid" already exists / HINT: Is another postmaster (PID …) running` — first server stayed healthy | 216 ms |
| same data dir, different port, **shared runtime dir** | same refusal, but only after the second process **wiped and re-extracted the running server's binaries** (the live postmaster survived on its open inode) | 5.3 s |

Two conclusions. First, the plan's every-command lock is belt to Postgres's braces, not the
only thing standing between us and corruption — good. Second, the same-port error message
("process already listening") is *wrong* when the collision is actually a **different
repository** using the same fixed port: two independent Roots were verified to run
simultaneously on distinct ports, so the default port is derived from the absolute `DataDir`,
stable across runs so two repositories never collide run after run (INV-29).

**AMENDED by F23 (2026-08-09, F4 round):** the derivation was
`49152 + fnv32a(dataDir) % 16000` — the IANA dynamic range — and that choice was wrong for
the reason F23 measures: the "dynamic" range is the range the OS allocates its own client
source ports from. It is now `10000 + fnv32a(dataDir) % 10000`. The claim that stability is
what "makes adoption work" was also wrong and is withdrawn: `livePostmaster` reads the port
out of the orphan's `postmaster.pid` (F4), so adoption finds a server wherever it is
actually serving. The derived port is only ever used to START one.

**F11 — goose applies from empty over the store's own pool with no package globals.**
`goose.NewProvider(goose.DialectPostgres, stdlib.OpenDBFromPool(pool), sub)` applied the
ledger migration from empty in 71–76 ms and re-ran as a 0 ms no-op, creating
`events, sync_runs, goose_db_version`. The instance-scoped provider is pinned over
`goose.SetBaseFS` / `goose.SetDialect`, which are package-level mutable state. pgx documents
that closing the `*sql.DB` from `OpenDBFromPool` does **not** close the pool, so the migration
handle is safe to close while the store keeps running.

**F12 — A small append-only table scans in insertion order, so "no `ORDER BY` at all" is
INVISIBLE to an ordering test.** Measured during Task 3 mutation-proving: deleting the
`ORDER BY seq` clause from `ReadEvents` entirely left **every** ordering assertion green —
both the strictly-increasing test (INV-10) and the ULID/wall-clock anti-trap (INV-11) — because
PostgreSQL returns a never-updated three-row heap in the order the rows were written, which is
seq order. Sorting by the WRONG column is caught; sorting by NOTHING was not. The fix is to make
physical order disagree: `CLUSTER events USING events_idempotency` rewrites the heap in
`(source, native_id, content_sha)` order, so a fixture whose `native_id`s descend as `seq`
ascends ends up physically reversed. INV-10's test now does this and carries a vacuity guard that
fails if the CLUSTER ever stops biting. **Lesson, generalised:** an ordering assertion over data
that is already in the right order proves the absence of a wrong sort, never the presence of a
right one.

**F13 — `information_schema` is PRIVILEGE-FILTERED; `pg_catalog` is not.** Measured the same way:
adding a projection table to the migration stream *without* a matching `GRANT` left INV-19 and
INV-30 green, because both listed tables via `information_schema.tables` through `WithTx` — which
runs as `vera_projection`, and `information_schema` shows a role only the objects it holds a
privilege on. The invisible table was exactly the thing the invariant forbids. Both tests now read
`pg_catalog.pg_tables`, which is not privilege-filtered. This is also why the migration grants
`SELECT ON goose_db_version` to `vera_projection`: without it the version table is invisible from
the projection role's vantage point. That grant is now belt-and-braces for the schema-listing
tests rather than load-bearing, and is retained because a projection inspecting its own schema
should still see a truthful picture.

**F14 — The file lock was racy at its root: a live holder never re-checked that it still
held the lock, and takeover was unlink-then-create.** Measured during the 2026-08-09
adversarial review, on the original implementation.

**SUPERSEDED — read the diagnosis, do not implement the remedy.** The three properties this
finding prescribed (an ownership token, an atomic publish, a compare-and-swap takeover) were
built, and F14b records them failing review in turn. They are retained because the FAILURE
below is still the canonical description of what two writers on one ledger looks like; the
remedy is now `flock` (F16–F21), which removes the questions rather than answering them
better.


| Step | Observation |
|---|---|
| Process A acquires the lock | `t=973.838` |
| `.vera/sync.lock` removed (it is gitignored — `git clean -xdf` does exactly this) | the heartbeat's `Chtimes` failed, logged `Warn` to a **discard-by-default** logger, and **continued** |
| Process B acquires the same lock | `t=974.821` — the path was free, so nothing refused it |
| B adopts A's postmaster (correctly: it serves this DataDir, § 4 F4) and later Closes | **the server A was still appending to stopped** |
| A's next read | `SQLSTATE 57P01` (admin shutdown) |
| A's `Close` | **reported success** |

Two writers, one ledger, no complaint — against a `SinceSeq` contract whose soundness is
single-writer (§ 6). Three properties were derived from it at the time. **All three are
superseded** — each is listed with what actually happened to it:

1. **Ownership must be a token, not a pid.** Pids are recycled; a nonce minted per
   acquisition is the only thing that answers "is this record still *my* claim". The
   heartbeat re-reads it every beat and the Store is poisoned when it does not match
   (INV-37) — a lock holder that merely assumes is the whole failure above.
   *Superseded: a constant nonce passed the whole suite (F14b), and flock needs no token
   because the kernel, not a file, records who holds the lock.*
2. **The record must never exist half-written.** `O_CREAT|O_EXCL` then write leaves a
   zero-byte lock if the process dies between the two, and a zero-byte record parses to
   pid 0 — which a `pid <= 0 ⇒ alive` probe calls alive **forever**, wedging the ledger
   until someone deletes a file by hand. Write-temp-then-`os.Link` publishes the complete
   record atomically (`link` fails with `EEXIST`, so exclusivity is preserved), and pid 0
   is never a live process (INV-38).
   *Superseded: the record is informational now, so there is nothing to publish atomically
   and a half-written one costs a name in an error message.*
3. **Takeover must be a compare-and-swap.** Probe-then-`os.Remove`-then-create lets a
   loser, preempted between the probe and the remove, delete the winner's *fresh* lock. It
   was not reproduced in 72 racing processes (the window is microseconds) — which is the
   argument for fixing it structurally rather than testing for it: `os.Link` onto a marker
   named after the OBSERVED identity gives exactly one winner per identity, and a
   bytes-plus-inode re-check before the unlink proves the file is still the one that was
   judged (INV-22).
   *Superseded, and it did not work: `os.Link` shares the source inode, so every marker was
   born already stale and a live competitor's marker was deleted on sight (F14b). There is
   no takeover at all now — the kernel releases a dead holder's lock (F17).*

**F14b — the rebuilt lock failed the same way, and the proving tests could not have
caught it.** Measured during the 2026-08-09 round-3 review, on the nonce/CAS/heartbeat
implementation:

| Claim | What was measured |
|---|---|
| "a live holder re-verifies ownership" | It re-verified on the HEARTBEAT TICK only; `conn()` consulted a cached flag. Lock file removed at +0.735s, second process acquired at +0.838s, and BOTH processes appended interleaved to one ledger for **2.6 s** with no error on either side. The proving test pinned `HeartbeatInterval` to **25ms** against a shipped default of **60s** — 1/2400 — so the real exposure was never measured. |
| "takeover is a compare-and-swap" | `reclaim()` won its marker with `os.Link(path, claim)`. A hard link SHARES the source inode, so the marker's mtime IS the stale lock's mtime and every marker is born already stale. Same fixtures, two constructions: a marker made with `os.Link` (what the code does) → Open RECLAIMS; a marker made with `os.WriteFile` (what the test planted) → Open refuses. **The test proved a shape the code never created.** |
| "the nonce is an ownership token" | `newNonce` returning a fixed constant passed the entire suite: the test only proved that a DIFFERENT hardcoded nonce was rejected, never that two acquisitions differ. |

Nine of sixteen independent mutations survived, clustered on the safety half of the lock.
The lesson generalises past this package, and both halves are now spec rules: **a test
must exercise the SHIPPED default configuration**, and **a proving test must construct
the artifact the way the code constructs it.**

**F15 — `ReadEvents` held a pooled connection across every `yield`, so the read seam was
not re-entrant.** Measured the same day: calling `WithTx` from inside `yield` at
`MaxConns: 1` never returned (deadline exceeded); at the default `MaxConns` it hung at a
depth and load that varied per run, which is worse — it passes in a test and stops a
rebuild in production. The consumer this forbids is the obvious one: a projection reducer
walking the ledger from genesis and writing derived state as it goes. The fix is to page
the walk (512 rows) and release the connection **before** any `yield` runs; the rows and
their order are unchanged at any page size (INV-35).

**F16 — a second process is refused immediately, and touches nothing.** Measured
darwin/arm64, Go 1.26.2, 2026-08-09, with a scratch program outside the repo. Five
consecutive attempts against a held lock: `flock(LOCK_EX|LOCK_NB)` failed with
`EWOULDBLOCK` in **115–142µs** every time, and a sibling data directory's file was
byte- and mtime-identical afterwards. `EWOULDBLOCK` and `EAGAIN` are the same errno here,
so `errors.Is(err, syscall.EWOULDBLOCK)` is the correct discriminator.

**F17 — `kill -9` releases the lock; there is nothing stale to detect.** A holder was
SIGKILLed and the lock became takeable **105µs** later, with its lock file still on disk
and still naming the dead pid. In the real suite (a full `Open`, with the killed holder's
postmaster still serving) reacquisition took **39 ms** with the lock file **6.3 s** old.
No age threshold, no pid probe, no takeover: this single property is what retires INV-22,
INV-24 and INV-38.

**F18 — the lock lives on the INODE, so the PATH is the one remaining hole — and it costs
1.5µs to close.** With A holding the lock, unlinking the path and letting B create a new
file there produced **two legitimate flock holders** on different inodes (40867463 and
40867464). A's `os.SameFile(fstat(held fd), stat(path))` returned false — detected — and
the same check with the path removed entirely returns `ErrNotExist`. Cost over 20 000
iterations: **1.494µs per call**. That is roughly 0.1% of a trivial database round trip,
which is why the check runs on every operation with **no TTL cache**: a cache would
reintroduce the exact "lost but still reads as held" window that failed in F14b.

**F19 — a second `open`+`flock` in the SAME process is refused** (`EWOULDBLOCK`, 193µs),
because the lock belongs to the open file description. Re-locking a descriptor already
held is a no-op, and a `dup` of it shares the lock — so in-process protection is free, but
only for genuinely separate `open` calls.

**F20 — without `LOCK_NB`, acquisition blocks for the holder's remaining lifetime.**
Measured: **1.202 s** against a holder scheduled to live 1.2 s, then succeeded. A CLI that
did this would hang on a busy ledger instead of naming who has it. Also measured:
`LOCK_EX` succeeds on an `O_RDONLY` descriptor, so read-only opens are not a loophole.

**F21 — a released flock stays briefly visible to others while a sibling `fork`/`exec` is
in flight, and O_CLOEXEC is what stops it becoming permanent.** A forked child holds a
duplicate of every open file description until it `exec`s, so a lock released in that
window is still held by the child's copy for microseconds. Measured under continuous
`fork`/`exec` pressure: **3 of 3000** clean releases were transiently refused, worst case
**500µs**. Consequences, both real:

- *Benign:* a spurious `ErrLocked` is possible if another process attempts acquisition
  inside that window. It can never produce two holders — it only delays one.
- *Load-bearing:* this suite forks constantly (`initdb`, `pg_ctl`, helper subprocesses),
  so "released" assertions must tolerate a bounded window; they poll for up to 2 s and
  still fail hard on a genuinely unreleased lock, which no `exec` will ever free.
- *The failure this rules out:* `pg_ctl` daemonizes, so if the lock descriptor crossed
  `exec` the postmaster would hold the flock after our process exited and the ledger would
  be permanently unopenable. Go opens files `O_CLOEXEC` by default; INV-40's test is where
  a regression would surface, because it reacquires while the killed holder's postmaster
  is still running.

**F22 — the residual window of a lost lock is one write round trip, and the queue it
replaced was unbounded.** Measured 2026-08-09 against a real embedded server, through
`Sync.Append` itself.

| What | p50 | p99 | max |
|---|---|---|---|
| One appender, no contention (`n=400`) — the FAST case, **not the bound** | 142.75µs | 223.7µs | 3.19ms |
| One appender, ledger under `ACCESS EXCLUSIVE` — **the duration is unbounded** | — | — | **4.059s** |
| 4 concurrent appenders | 230µs | 310µs | 394µs |
| 8 concurrent appenders | 443µs | 714µs | 822µs |
| 16 concurrent appenders | 739µs | 1.35ms | 1.42ms |
| 64 concurrent appenders | 3.09ms | 5.95ms | 6.04ms |

Read the second block as the exposure BEFORE the fix. With the only ownership check in
`conn()`, an append's answer dated from the top of its call, so its total call time is
exactly how stale that answer was when its INSERT ran — and it grows with queue depth
without limit. Held open (a connection parked, a slow statement, a lock wait), it is
whatever the wait is: the review measured **8 appends landing after the loss, the last
4.012s later**, interleaved with the new holder's rows.

Re-checking inside the critical section moves the exposure from "the queue" to "one write",
and that write cannot be made zero: nothing transactional spans the filesystem lock and the
PostgreSQL INSERT. The bound is what INV-41 now claims, in place of an absolute that was
never reachable.

**F23 — the two numbers a green suite was resting on, and neither was chosen.** Measured
2026-08-09 while diagnosing an intermittently red `make check`.

1. **The derived port range WAS the OS's ephemeral range.** `sysctl
   net.inet.ip.portrange.first` on this machine returns **49152** — byte-identical to the
   old `portFloor`. Linux's default `ip_local_port_range` (32768-60999) overlaps it too. So
   the ledger's LISTENING port was drawn from the pool the kernel hands out to every
   outbound connection on the machine. Sampled during a suite run: **32-118** local ports in
   49152-65152 in use at a time, and **0.03-0.07%** of a random 4000-port sample in that
   range refused a bind (only *listening* sockets conflict — `SO_REUSEADDR` means an
   established connection on a port does not block a new listener, which is why the failure
   rate is far below the occupancy). Small, load-dependent, and entirely avoidable: the band
   is now **10000-19999**, below both platforms' ephemeral floors and disjoint from the
   20000-32000 band the test harness pins from.
2. **`embedded-postgres`'s default `StartTimeout` is 15s; a cold `Open` measured 7.8-8.9s**
   (12 samples, under this repo's own test parallelism). That is a **1.7x** margin on an
   otherwise-idle machine — the tightest budget in the package by an order of magnitude, and
   one nobody picked: it is the library's default, while the test harness had already had to
   give its OWN cluster 2 minutes for the same workload. `startServer` now sets it
   explicitly to 2 minutes. Erring long is the right direction: a flight recorder that waits
   is recoverable; one that abandons a half-initialised data directory is the F2 amnesia
   trap arriving by another road.

**F24 — a closed listening socket is NOT released while a sibling `fork`/`exec` is in
flight. This is the flaky gate, caught in the act and then reproduced on demand.** It is
`F21` — measured there for `flock` — arriving on a different kind of file descriptor.

The red run was captured with a diagnostic that re-probed the port and dumped `lsof`:

```
--- FAIL: TestLock_LossIsCaughtBeforeTheNextOperation/the_lock_file_replaced_by_a_different_file
    ... starting postgres ... on port 30286: process already listening on port 30286
    DIAG port=30286 claimedByHarness=true
    lsof:                      <- EMPTY. nothing was listening.
--- FAIL: TestOpen_AdoptsOrphanServer/never_adopts_a_foreign_process_on_the_port
    occupying port 22542: listen tcp 127.0.0.1:22542: bind: address already in use
```

A port the harness had just bind-probed as free, refused microseconds later, with nothing
listening on it. The mechanism is fd inheritance: a forked child holds a duplicate of every
open file description until it `exec`s, so a listening socket closed HERE stays bound
THERE for the width of that window. This suite forks continuously — `initdb`, `pg_ctl`,
helper subprocesses — so every probe-then-use of a port sits in it.

Reproduced deliberately, outside the suite: bind → close → rebind the same port in a loop,
with six sibling goroutines doing nothing but `fork`/`exec` of `/usr/bin/true`:

| Attempts | EADDRINUSE | Rate | Worst transient |
|---|---|---|---|
| 54 214 | 2 279 | **4.20%** | **2.63ms** |

Percent-level, on a port nothing was listening on. That is the whole explanation for a gate
that was green six times in seven: a suite run performs ~30 probe-then-use sequences, each
exposed at a few percent whenever a sibling happens to be mid-`fork`.

The fix is to stop believing a single refusal, in every place that binds:
`ensurePortFree` retries EADDRINUSE for `portProbeWindow` (2s) before reporting a port busy;
`startServer` retries `Start()` up to `startAttempts` when the library reports the port busy
AND our own re-probe says it is free; the harness's `portIsBindable` and `listenRetrying` do
the same. None of it is slack for a genuinely occupied port, which never becomes free — the
loud refusal still happens, one window later. Same reasoning, and the same budget, as
`releaseVisibilityBudget` for the flock case.

**Honest status of the earlier hypotheses.** The two facts in F23 are real and were fixed on
their own merits, but neither was this flake: the derived-port band never appears in the
captured failures, and no start ever came near the 15s budget (measured 7.8-8.9s). They were
found while hunting, not by catching. F24 is the one that was caught — twice, in different
tests, with the port measurably free both times.

---

**F25 — a symlinked FINAL component derived a second lock for one data directory; two
Stores both acquired.** The lock path was derived LEXICALLY from `DataDir`, but the
postmaster-adoption check resolves symlinks — so `db` and a symlink pointing at it agreed
about the cluster and disagreed about the lock. Measured: two Stores opened one data
directory via the two spellings, both acquired their "exclusive" lock, and the ledger took
**40 interleaved rows**. Closed by deriving the lock from the RESOLVED data directory
(`lockPathFor` → `normalize`); proven by `internal_test.go::TestLock_SymlinkedDataDirLeafDerivesOneLock`.
This finding's fix is what exposed F26 below. (This definition was restored 2026-08-09:
five citations referenced F25 while no definition existed — found independently by the
round-7 adjudication (M10) and by `scripts/invariant-lint.sh` on its first run.)

**(F27) EWOULDBLOCK did not mean what this package claimed it meant — measured.** The
lock's design note said `LOCK_EX|LOCK_NB` leaves "nothing to probe and nothing to decide",
and `acquireLock` mapped every EWOULDBLOCK to "another live process holds the ledger,
necessarily alive". Both were false. `flock` belongs to the OPEN FILE DESCRIPTION, and
`fork()` duplicates descriptors, so a child forked while this process held the lock keeps
that description alive after our own `release()` closed our copy — until the child `exec`s
(O_CLOEXEC) or exits. In that window the kernel refuses a lock NOBODY logically holds, and
`heldBy` then read the pid out of the lock file and named THIS process as the holder.
Self-inflicted, because the store forks constantly (initdb, pg_ctl). *Measured (this paragraph is the SINGLE home for these figures — round-8 finding 8 found them
restated in four places with three values): 112 of 300 acquires refused on a path nothing else
ever locked, with an accidental fixture that forked without exec'ing; 6 of 300 with a real exec,
which is why the deterministic `dup(2)` proof exists. A separate 83-sample run measured the
phantom's clearance: p50 289µs, p90 1.00ms, p99 1.24ms, max 1.83ms — every one cleared. Rates
are machine- and load-dependent; the round-8 reviewer measured different ones on their host,
which is a property of the measurement, not a contradiction.* Closed by a bounded
re-probe (INV-47) sized at 250ms, ~135x the observed maximum. The direction matters: the
residual error was a false "locked", never a false "acquired", because `flock` still grants
the lock and nothing else does — so retrying cannot produce two writers, it can only stop
refusing a free ledger. Proven deterministically by `dup(2)`, which creates the same
one-description-two-descriptors condition on demand: the statistical fork test caught the
unfixed code only 6 times in 300 with a real `exec`, and a 2%-sensitivity detector is one
faster machine away from proving nothing.

**(F28) A run counted rows it had not finished writing — measured.** `Append` took the run
mutex to read `finished`, RELEASED it across the INSERT, and re-took it to increment. A
`Finish` arriving in that gap wrote the run's durable `events_appended` without the
in-flight row, which then landed anyway: a row in `events` that its own run says it never
appended, falsifying INV-6 and `Finish`'s own promise that a racing pair "resolves one way
or the other rather than both". *Measured at the time: 48 of 300 trials — on a fixture that opened a store per trial and no longer ships. On today's fixture the same mutant surfaces ~1-3 per 300 trials; the detector's power now comes from trial COUNT, and its measured kill rate lives next to the test rather than here.* Closed by holding the mutex
across check → INSERT → increment (INV-48), which costs no throughput because `appendMu`
already serializes every append on the Store; the only thing that now waits is a concurrent
`Finish`, which is the promised semantics. Lock order is `appendMu` then `sy.mu` in Append
and `sy.mu` alone in Finish, so there is no cycle.

**(F30) The counter was tied to the client's observation, not to the row's fate — measured.**
F28 closed the interleaving it was shown (a `Finish` landing in the released-mutex gap) and left
INV-6/INV-48 still false by a second, more reachable route: a context deadline expiring AFTER
PostgreSQL commits but BEFORE pgx returns leaves the row durably in `events` while `sy.appended`
is never incremented. *Measured by the round-8 reviewer: 202 of 600 deadline trials left an
uncounted row — 232 rows against a durable `events_appended` of 30. Independently reproduced here
at 137 of 200 with the fix removed.* Closed by confirming the row's fate before reporting failure:
on a non-`ErrNoRows` error the append re-reads by idempotency key on a context DETACHED from the
caller's (the caller's is usually exactly what died) and counts a row that is actually there. The
error is still returned — the client's observation really was a failure — but the ledger's
accounting no longer disagrees with the ledger's contents. Safe against a retry: the retry takes
the idempotency branch, which does not increment.

**The lesson is bigger than the fix.** Closing the reproduction you were HANDED is not the same as
making the invariant true. F28's race was real and is fixed; the invariant was still false. When a
fix answers a specific reproduction, the remaining question is what OTHER routes reach the same
violated state — here, any path that returns an error after a successful commit.

**(F32) The repair for an undercount produced an unbounded overcount, and the invariant
was wrong rather than the code — measured.** F30 recorded that a context deadline landing
after the commit leaves a row uncounted (202 of 600 trials). The repair re-read the row by
idempotency key on the error path and counted it if present. That asks the wrong question:
"a row with this key exists" is not "THIS call inserted a row", and the two differ every
time a retry meets a row an earlier attempt — or another run — inserted. *Measured: a
connector retry loop against one already-present event produced 24 durable appends against
1 row, and a run that inserted nothing claimed an append.* Unbounded in retries.

It is not repairable with what the ledger records. `event_id` does not discriminate: a
caller retrying the same `core.Event` reuses its ULID, so two runs appending one event carry
one id. And `events` deliberately holds no run reference. After an errored INSERT, whether
this call created the row is simply not recoverable.

So INV-6 was corrected instead of the code: the count is bounded ABOVE by the rows the run
inserted and may fall below them, and the error path counts nothing. Between two available
wrong answers, the one that understates real work beats the one that fabricates work —
a journal that overstates is not merely imprecise, it is misleading.

**Why nine rounds did not catch it.** Every assertion about this counter was ONE-SIDED:
both race detectors asserted only `counted < added`, and INV-6's cited test exercised only
the success and conflict branches, never the error branch. An equality invariant was tested
exclusively as an inequality, so an overcount was invisible by construction. The general
form is worth keeping: **when an invariant is an equality, a test that checks one direction
does not test it** — and the untested direction is where the next defect lives.

**(F31) The lifecycle mutex was held across teardown, and Close deadlocked against its
own caller — measured, pre-existing.** `Close` held `s.mu` across `unwind`, and `unwind`
calls `pool.Close()`, which waits for every checked-out connection to be returned. A
`WithTx` callback holding that connection and re-entering the Store blocked in `conn()`
waiting for `s.mu`. Neither side takes a context, so nothing timed out: *measured, `Close()
did not return within 8s`, the callback never returned, and the run died on the test
binary's own 90s timeout.* `conn`'s doc comment asserted the opposite — that a racing
`Close` is "resolved by pgx, not by blocking every read behind the lifecycle mutex" — which
was false when written and is true only now. Closed by settling `closed` and DETACHING the
resources under the lock, then releasing them outside it (INV-50); after the detach,
`usable()` returns `ErrClosed` and `Lock()` returns the zero `LockInfo`, so no method can
observe a half-torn-down resource. Not introduced by round 7 — it predates the remediation
and no round before 8 looked at teardown re-entrancy.

**(F29) A mutation sweep believed its own tests, and the survivor count was wrong twice.**
§ 5 claimed ONE known-surviving mutant after the F1-F4 round; an instrumented sweep found
EIGHT. The instrument is the difference: calibration controls (a positive control that MUST
die — deleting an `ORDER BY seq` the suite pins — and a comment-only control that MUST
survive) run before any count is believed, and path-guard mutations re-run under a
realpath'd `$TMPDIR`. Three guards had been "proven" only because macOS hands back a
symlinked `t.TempDir()`. Of the eight: five are now killed by named tests
(`enclosingCluster` wiring, dead-pid adoption, the IPv6 probe, the lost-lock level,
`normalizeParent`), one was DEAD CODE removed rather than tested (the unreachable
`noteLost` memoisation), and two are disclosed as surviving in § 5 with a cost argument
rather than left implied.

**(F26) A fix to one guard silently opened another — measured.** Deriving the lock from the
RESOLVED data directory (F25) closed the two-writer hole, and in the same stroke stopped the
`LockPath`-inside-`RuntimeDir` guard from firing: `within` was a lexical prefix test, so it
compared a resolved `/private/var/...` lock against an unresolved `/var/...` RuntimeDir and
found no containment. The guard returned "safe" for a lock file sitting in the directory
embedded-postgres wipes on every start. Two of the three tests that caught it were asserting
on the caller's SPELLING rather than on file identity, so they failed against the CORRECT
behaviour — a test that pins a spelling cannot referee a change of representation. Closed by
routing every path comparison through one `normalize` (INV-44) and re-stating those assertions
as file-identity claims; mutation-proven by reverting `within` to the lexical form, which re-opens the guard and fails
`TestOpen_RefusesLockInsideRuntimeDirReachedBySymlink`. That test builds its OWN symlink: the
first proof of this fix relied on macOS handing back `/var/folders/...` for `/private/var/...`,
so the divergence was an accident of the platform and the same mutant SURVIVED under a `$TMPDIR`
that is a real path — a guard proven on one operating system only. The same lexical blindness had
always let a symlinked route evade the DataDir/RuntimeDir nesting guard, so this closes a
latent hole that predates F25.

## 5. Invariant table

Format: `| INV-<n> | <statement> | <test file>::<TestName> |` — the pinned rule and its
rationale live in `.claude/commands/vera-spec.md` § 5 (single home; do not restate it here).
Most invariants own one row; several own more (see the note below the table). The third cell names a real Go test function in this package — `scripts/invariant-lint.sh` fails the build if any citation here, or any `F<n>` reference anywhere in this document, does not resolve.
**Citation resolution IS enforced today** by `scripts/invariant-lint.sh` — BLOCKING, inside
`make check` (docs/gates.md): every `<file>.go::<Test…>` citation in this table and every `F<n>` reference
anywhere in this document must resolve, or the build fails. It has caught four broken citations
that authors introduced *while fixing other citations*.

**What it does NOT do, and this is the honest status:** it guarantees a citation RESOLVES, never
that the named test PROVES its claim. That semantic half stays with adversarial review, and it
has been got wrong three times (a row citing a test that asserts something weaker; a row attached
to an invariant about a different subject). P1 Task 9 adds the remaining mechanical half: every
`internal/*` package must have a SPEC whose table names at least one existing test.

| Invariant | Statement | Proving test |
|---|---|---|
| INV-1 | A new event appends with inserted=true and a positive seq | append_test.go::TestAppend_FirstInsert |
| INV-2 | Re-appending an identical event returns inserted=false and the original row | append_test.go::TestAppend_IdempotentReturnsExistingRow |
| INV-3 | A new content_sha for the same subject appends a revision at a higher seq | append_test.go::TestAppend_RevisionAppendsNewRow |
| INV-4 | A duplicate event_id with a new idempotency tuple is rejected, never absorbed | append_test.go::TestAppend_ForeignUniqueViolationIsLoud |
| INV-5 | Append validates before the round trip and consumes no sequence value | append_test.go::TestAppend_ValidatesBeforeInsert |
| INV-6 | events_appended counts inserted=true appends and is never caller-reported | append_test.go::TestSync_EventsAppendedIsDerived |
| INV-7 | Payload bytes round-trip byte-identically and still hash to content_sha | append_test.go::TestAppend_PayloadBytesSurviveRoundTrip |
| INV-8 | occurred_at and recorded_at round-trip to the same instant | append_test.go::TestAppend_TimestampsSurviveRoundTrip |
| INV-9 | Same subject and payload with a different kind returns ErrKindConflict | append_test.go::TestAppend_KindConflictIsNotAbsorbed |
| INV-10 | ReadEvents yields strictly increasing seq across a ledger containing gaps, even when physical row order disagrees | read_test.go::TestReadEvents_SeqIsTotalOrderDespiteGaps |
| INV-11 | Replay order ignores event_id and occurred_at | read_test.go::TestReadEvents_OrderIndependentOfIDAndTime |
| INV-12 | Filter matches source and kind exactly and treats SinceSeq and OccurredAfter as exclusive | read_test.go::TestReadEvents_FilterSemantics |
| INV-13 | Early exit returns nil, other yield errors propagate, and no connection leaks | read_test.go::TestReadEvents_EarlyExitReleasesConnection |
| INV-14 | The package's public handles expose the pinned safe surface | surface_test.go::TestSurface_NilAndClosedHandles |
| INV-15 | No exported signature mentions a database driver or migration library | surface_test.go::TestNoDriverEscapesTheSurface |
| INV-16 | No mutating SQL targets events, and sync_runs has exactly one UPDATE | surface_test.go::TestNoMutatingLedgerSQL |
| INV-17 | WithTx commits on success, rolls back on error, and rolls back then repanics | projection_test.go::TestWithTx_TransactionDiscipline |
| INV-18 | Ledger writes through a projection transaction are refused by the database | projection_test.go::TestWithTx_LedgerIsReadOnlyToProjections |
| INV-19 | The migration stream creates no projection table | projection_test.go::TestMigrations_ContainNoProjectionDDL |
| INV-20 | A second Open on the same root fails with ErrLocked and leaves the data dir untouched | lock_test.go::TestLock_SecondOpenIsRefused |
| INV-21 | A second OS process running a ledger command exits non-zero and touches nothing | lock_test.go::TestLock_SecondProcessExitsNonZero |
| INV-22 | RETIRED (round 3) — no takeover exists; the kernel releases a dead holder's lock | — |
| INV-23 | An EPERM liveness probe counts as alive, so a postmaster we cannot signal is never treated as dead | embedded_test.go::TestOpen_EPERMPostmasterCountsAsAlive |
| INV-24 | RETIRED (round 3) — no staleness rule, so nothing needs a heartbeat to stay fresh | — |
| INV-25 | Close releases the lock (a competitor can take it), leaves the lock file, is idempotent, and later calls return ErrClosed | lock_test.go::TestClose_ReleasesLockAndIsIdempotent |
| INV-26 | DatabaseURL mode takes no data-dir lock and starts no server | lock_test.go::TestOpen_DatabaseURLTakesNoLock |
| INV-27 | Open refuses a data dir nested in the runtime dir and other bad config | embedded_test.go::TestOpen_RejectsUnsafeConfig |
| INV-28 | Open adopts a live server for this data dir and never adopts a foreign one | embedded_test.go::TestOpen_AdoptsOrphanServer |
| INV-29 | Two stores on different roots are open at the same time | embedded_test.go::TestOpen_TwoRootsCoexist |
| INV-30 | Migrations apply from empty and create exactly the pinned ledger objects | migrate_test.go::TestOpen_MigratesFromEmpty |
| INV-31 | A second Open applies no migration and preserves every row | migrate_test.go::TestOpen_SecondOpenPreservesRows |
| INV-32 | A migration failure returns ErrMigrate and no Store, and a post-lock Open failure strands no lock | migrate_test.go::TestOpen_MigrationFailureReleasesLock |
| INV-33 | A data directory written by another PostgreSQL major is refused before initdb | embedded_test.go::TestOpen_RefusesAForeignDataDirVersion |
| INV-34 | Finish reports success only after the journal row is written | append_test.go::TestSync_FinishReportsSuccessOnlyAfterTheWrite |
| INV-35 | yield may use the Store, and a walk longer than one page yields every row exactly once | read_test.go::TestReadEvents_YieldMayUseTheStore |
| INV-35 | Paging is invisible across page boundaries, for SinceSeq and for Limit | read_test.go::TestReadEvents_PagingIsInvisible |
| INV-36 | At most one append is in flight at a time within one process | append_test.go::TestAppend_IsSerialisedWithinTheProcess |
| INV-37 | A lost lock poisons the Store and is never released as if it were ours | lock_test.go::TestLock_LossIsCaughtBeforeTheNextOperation |
| INV-37 | Close detects the loss itself, with no operation in between to have noticed it | lock_test.go::TestClose_DetectsALostLockWithNoInterveningOperation |
| INV-38 | RETIRED (round 3) — the lock record is informational, so an unreadable one wedges nothing | — |
| INV-39 | The lock is an exclusive, non-blocking flock on LockPath, refused promptly to a second opener in any process | lock_test.go::TestLock_IsAnExclusiveNonBlockingFlock |
| INV-39 | A second Open in this process is refused, naming the holder from the informational record | lock_test.go::TestLock_SecondOpenIsRefused |
| INV-40 | The kernel releases the lock when the holder is SIGKILLed, so a fresh young lock file naming a dead pid does not block the next Open | lock_test.go::TestLock_KernelReleasesTheLockWhenTheHolderIsKilled |
| INV-41 | A lock lost to an unlinked or replaced path is caught before the NEXT operation — including when the path is simply GONE (the ENOENT branch) | lock_test.go::TestLock_LossIsCaughtBeforeTheNextOperation |
| INV-41 | An append merely QUEUED behind appendMu when the lock is lost is refused; at most the one already in flight lands | lock_test.go::TestLock_InFlightAppendIsTheOnlyResidual |
| INV-42 | The lock is derived from DataDir: a second Store on one data directory with a different LockPath is refused with ErrConfig | lock_test.go::TestLock_IsBoundToTheDataDirectory |
| INV-42 | A LockPath that is not this DataDir's lock, and a RuntimeDir that IS the derived lock path, are both refused | embedded_test.go::TestOpen_RejectsUnsafeConfig |
| INV-44 | Every comparison between two configured paths compares FILES, not spellings: all go through `normalize`, which resolves the longest existing prefix through symlinks | internal_test.go::TestLock_SymlinkedDataDirLeafDerivesOneLock |
| INV-44 | The lock-inside-RuntimeDir guard fires even when the two paths reach the same directory by different routes (resolved lock vs unresolved RuntimeDir) | embedded_test.go::TestOpen_RefusesLockInsideRuntimeDirReachedBySymlink |
| INV-45 | The pre-initdb port guard retries a transiently-bound port rather than returning on the first probe | internal_test.go::TestEnsurePortFree_RidesOutATransientBind |
| INV-44 | The enclosing-cluster guard sees through a symlinked route: one directory yields one verdict however it is spelled | internal_test.go::TestEnclosingCluster_SeesThroughASymlinkedRoute |
| INV-44 | The root contains every absolute path, and is contained by none of them | internal_test.go::TestWithin_TreatsTheRootAsAContainer |
| INV-46 | `samePath` resolves ancestors but NOT the leaf, so a LockPath verdict is total — identical whether or not the canonical lock exists yet | internal_test.go::TestSamePath_VerdictDoesNotDependOnWhetherTheLockExistsYet |
| INV-28 | A port held by a stranger is refused by the store's OWN message, BEFORE initdb, leaving no data directory behind | embedded_test.go::TestOpen_RefusesABusyPortByName |
| INV-43 | A Store that lost its lock does not stop the server the current holder is using | lock_test.go::TestClose_PoisonedStoreLeavesTheServerToItsNewHolder |
| INV-44 | The enclosing-cluster guard is WIRED into resolve: a DataDir inside another cluster's data directory is refused | internal_test.go::TestResolve_RefusesADataDirInsideAnotherCluster |
| INV-45 | A postmaster.pid naming a DEAD process is not adopted | internal_test.go::TestLivePostmaster_DoesNotAdoptADeadPID |
| INV-45 | The port probe checks BOTH loopbacks, so an IPv6-only squatter is seen | internal_test.go::TestProbePort_ChecksBothLoopbacks |
| INV-46 | The derived lock named through a symlinked PARENT is accepted — two spellings of one path are one lock | internal_test.go::TestSamePath_AcceptsTheDerivedLockSpelledThroughASymlinkedParent |
| INV-47 | acquireLock rides out a phantom holder (a duplicated open file description no live logic owns) instead of reporting a foreign holder — deterministic, via dup(2) | internal_test.go::TestAcquireLock_RidesOutAPhantomHolder |
| INV-47 | Acquires on an uncontended lock are not refused while this process forks | internal_test.go::TestAcquireLock_IsNotRefusedByThisProcessOwnFork |
| INV-48 | An Append racing a Finish resolves one way or the other: a row that lands is counted by the run that inserted it, and never by a run that did not | append_test.go::TestAppend_RacingFinishCannotLoseTheRowFromRunAccounting |
| INV-49 | A lost lock is reported at ERROR level | internal_test.go::TestNoteLost_ReportsAtErrorLevel |
| INV-50 | Close returns while a WithTx callback re-enters the Store, instead of deadlocking against its own teardown | surface_test.go::TestClose_DoesNotDeadlockAgainstAReenteringCallback |
| INV-6 | A retry loop against an already-present event never inflates the run's count | append_test.go::TestAppend_ARetryLoopDoesNotOvercount |
| INV-6 | A run that inserted nothing counts nothing, even when the row is present from another run | append_test.go::TestAppend_ARunThatInsertedNothingCountsNothing |
| INV-6 | The count never exceeds the rows the run inserted, whatever the deadline does; the undercount is measured, not asserted away | append_test.go::TestAppend_ADeadlineNeverOVERcountsAndMeasuresTheDisclosedGap |
| INV-16 | The SQL scan reads EVERY non-test .go file in the package, in or out of this build — a build-excluded file still ships SQL, and reading everything is the only direction that cannot hide it | surface_test.go::TestPackageGoFiles_ReadsEveryFileInOrOutOfTheBuild |
| INV-51 | A projection rollback runs on a context detached from the caller's, so it succeeds when the caller's context is what died | projection_test.go::TestWithTx_RollsBackEvenWhenTheCallersContextIsCancelled |
| INV-16 | The mutating-SQL scan judges a STATEMENT, folding concatenated literals and named consts (incl. consts that are themselves concatenations) | surface_test.go::TestFoldStringConcat_SeesAStatementBuiltFromParts |

Some invariants own more than one row, and the count is deliberately NOT enumerated here: an enumeration is a hand-maintained copy of something the table already states, and every such copy in this project has gone stale (round-7 M12 found the suite duration in four homes with four values; a prose count of "five invariants own two rows" was wrong by the time it was read). `grep -c '^| INV-44 |' SPEC.md` is the answer. Where an invariant does own several rows, the extra rows
proves a mechanism the first row's test cannot reach — a walk longer than one page; a loss
that only `Close` is present to notice; the whole-`Open` path as distinct from the raw
primitive; the queue behind `appendMu` as distinct from the operation choke point; and the
configuration refusal as distinct from the two-Store attack. Splitting the test is honest;
splitting the invariant would not be, because each pair is one property. INV-37 and INV-41
still SHARE `TestLock_LossIsCaughtBeforeTheNextOperation` for the opposite reason: they are
two statements about one event (a lock lost because the path was taken away), so one fixture
proves both and duplicating it would prove nothing twice.

**Mutations known to SURVIVE this table — ONE.** This count has now been wrong four times
(§ 4, F29 records the first three), so the number is stated with what would falsify it:
`grep -n "SURVIVOR:" SPEC.md` lists them, and any mutation of this package that the suite
does not kill belongs in that list or in a test.

What changed at round 9. Two entries previously listed here were WRONG to be listed:

- **`defer rows.Close()` in the paged read — KILLED**, and its disclosure was false on both
  claims. "Needs a row that scans in PostgreSQL and fails in Go — a type-level
  impossibility" was wrong (`event_id` is TEXT in the schema and a ULID in Go; one UPDATE
  on the shipped schema produces exactly that row), and "bounded by the pool's own
  lifetime, which Close reclaims" was wrong in the worst direction: `pool.Close()` WAITS
  for the leaked connection, so the mutant wedges teardown — measured 20s+ against
  microseconds. Now killed by
  `read_test.go::TestReadEvents_AMidPageScanFailureDoesNotWedgeClose`. **A cost argument
  is a claim and needs the same evidence as any other**; this one had none and was wrong.
- **`rows.Err()` after the page loop — was an UNDISCLOSED survivor**, found at round 9 and
  now listed below rather than left implied.

SURVIVOR: **deleting `acquireLock`'s identity confirmation between its `open` and its
`flock`.** The window is two adjacent syscalls, so proving it needs a production-only seam
or a stochastic churn test. Residual is bounded: the first `conn()` catches it (INV-41).
Retained deliberately.

SURVIVOR-CANDIDATE (owed, not accepted): **deleting the `rows.Err()` guard** after the page
loop. Shipped code is correct; the guard is simply unproven. Losing it makes `readPage`
return a truncated page with a nil error, which `ReadEvents` reads as end-of-ledger — a
silent short read of the ledger, which is not a residual anyone should accept. It is not
killed here only because the session closed; it is the first item of Task 3's residual list.

Three rows read RETIRED. They are kept rather than deleted because invariant numbers are
never reused (see the note at the top of § 3): a future reader meeting "INV-22" in a
review note or a git comment needs to be able to find out what happened to it, and "there
is no takeover any more" is the answer.

Tests are written from this table before implementation, failing rather than fake-passing
(no stub assertions). Where a test needs a fixture ledger it builds one in `t.TempDir()` at
runtime — nothing is committed.

---

## 6. Non-goals

A reviewer rejects these on sight:

- **No update, no delete, no rewrite of events — ever.** Not "discouraged": absent from the
  surface (INV-14), absent from the SQL (INV-16), and refused by the database for the one
  generic path that exists (INV-18). A retraction is a new event, not an edit.
- **No second append path.** `Sync.Append` is the only one; a `Store.Append` convenience would
  let `events_appended` drift from the rows that exist.
- **No cursor-as-correctness.** Nothing reads `sync_runs.cursor_json` to decide what to
  ingest. The UNIQUE index is the seen-set; a watermark would reintroduce every amend/rebase
  bug the plan's connector design removes.
- **No concurrent writers.** P1 is single-writer, and that is load-bearing: measured, `seq`
  assignment order is **not** commit order — a transaction holding `seq=1` can commit after
  one holding `seq=2`, so a reader that stored cursor `2` would never see row `1` (verified:
  one row permanently missed). `SinceSeq` is sound only because exactly one append is in
  flight. That is enforced on **two** levels, and both are needed: the data-dir lock between
  processes (INV-20, INV-37), and a mutex inside `Store` between goroutines (INV-36). The
  in-process half is not a convenience — without it the guarantee the read seam rests on was
  a discipline no code asserted, and two goroutines appending concurrently drew no complaint.
  Concurrent callers are serialised, never refused. If genuinely concurrent appenders ever
  arrive, `SinceSeq` needs a visibility-aware watermark (`pg_snapshot_xmin` / logical
  decoding) — that is a spec diff here, not a quiet change. The `DatabaseURL` escape hatch
  takes no data-dir lock (it protects a directory, not a database), so it gets the
  in-process guarantee only; single-writer discipline ACROSS processes there belongs to the
  operator. Between processes the lock is now the kernel's `flock`, so the between-processes
  half holds without any inference about whether another holder is still alive (INV-39,
  INV-40) — and the one case it cannot see, the lock PATH being deleted underneath a live
  holder, fails the ledger closed before the next operation rather than silently admitting a
  second writer (INV-41). One append can still land in that transition and the spec says so
  rather than claiming otherwise: the write already in flight when the lock goes cannot be
  recalled, because no transaction spans the filesystem lock and the PostgreSQL INSERT. The
  bound is one write round trip (§ 4, F22). The other way two writers were reachable —
  agreeing on a data directory while disagreeing about its lock — is refused at `Open`
  (INV-42).
- **No re-derivation of `content_sha` on the write path**, and no exported integrity verifier
  in P1. INV-7 proves the ledger *can* be re-verified byte-exactly; exposing a `Verify` walk
  is a P2 concern (the gates engine reads the same seam).
- **No payload schema knowledge.** store never parses a commit sha or a witness field out of
  `payload`; it never indexes into JSON. That is each connector's SPEC.
- **No projection tables in migrations**, no `projection_version` handling here, no reducer
  logic, no report queries. store lends `WithTx`; projections own everything inside it.
- **No multi-value filters, no descending reads, no pagination tokens, no aggregates.** One
  `Filter`, one direction, one iteration shape. P2 widens it with a spec diff.
- **No `iter.Seq2`, no slice-returning read.** One iteration shape, pinned (§ 2.3).
- **No pgx, goose, or embedded-postgres type in an exported signature** (INV-15). No
  `*pgxpool.Pool` accessor "just for tests".
- **No daemon, no server management commands, no connection to a remote Postgres by default.**
  `Close` always stops the server it is talking to; there is no `vera db start`.
- **No hand-rolled mutual exclusion.** The ledger lock is `flock(2)` and nothing else. No
  staleness window, no takeover, no reclaim marker, no ownership nonce, no pid-liveness
  probe for the lock, and no heartbeat — every one of those was deleted in round 3 after
  two hand-built designs failed adversarial review by admitting two writers (§ 4, F14 and
  F14b). Reintroducing any of them is a spec diff here first, and INV-14's pinned surface
  fails until § 2 is amended.
- **No lock-timing configuration.** There is nothing to tune, deliberately: a knob is how
  round 2 came to prove its detection latency at 1/2400 of the value users ran.
- **No new dependency.** § 7 is the whole list; anything else needs a `/vera-decide` record
  first (Build Law 8).

**Known limitation, stated rather than hidden.** Adoption (INV-28) trusts
`<DataDir>/postmaster.pid` plus a liveness probe to decide that a listening server is *ours*.
A different Postgres started by hand against the same data directory would be adopted, because
that is indistinguishable from our own orphan — and it is also, correctly, the same server
serving the same ledger. What is *not* adopted is a process on the port with no live
`postmaster.pid` for this data directory (INV-28, second half).

---

## 7. Dependencies

**Standard library** — `context`, `database/sql` (only to hand goose a handle; never
exported, INV-15), `embed`, `encoding/json`, `errors`, `fmt`, `hash/fnv` (the derived
default port), `io/fs`, `log/slog`, `net`, `os`, `path/filepath`, `strconv`, `strings`,
`sync`, `syscall` (**`Flock` — the ledger lock itself** — plus `Kill` for the postmaster
liveness probe and the errno comparisons), `time`; `go/parser` + `go/ast` in tests only.

`crypto/rand` is **no longer used**: it minted the lock's ownership nonce, and there is no
ownership token any more — the kernel holds the lock, so nothing has to be proven about
who wrote a file. `bytes` and `io` went with the record-comparison and read-once machinery
for the same reason. Note that `syscall` moved from a supporting role to a load-bearing
one: `syscall.Flock` IS the exclusion mechanism, which also pins this package to unix —
a constraint it already carried through `syscall.Kill` and embedded-postgres.

**Internal** — `vera/kernel/internal/core` (the envelope; store adds `seq` and never anything
core forbade).

**Blessed external dependencies (VD-stack-go-fid9mi) — the three store is permitted to add,
and no others:**

| Module | Version verified | Used for |
|---|---|---|
| `github.com/jackc/pgx/v5` | v5.10.0 | the single `*pgxpool.Pool`, `pgconn.PgError` codes, `stdlib.OpenDBFromPool` for goose |
| `github.com/fergusstrange/embedded-postgres` | v1.34.0 | the dev/test Postgres child process (PostgreSQL 18.3.0) |
| `github.com/pressly/goose/v3` | v3.27.3 | ledger migrations over `embed.FS` via the instance-scoped `NewProvider` |

**That is the entire list.** goose pulls its own transitive modules into `go.sum`; a new
*direct* dependency requires a `VD-` record before `go get` (Build Law 8), and INV-15's scan
plus review are what keep the surface honest.
