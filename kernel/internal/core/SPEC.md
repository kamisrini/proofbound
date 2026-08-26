# SPEC — `internal/core`

**Status:** authored before implementation (Build Law 6) · P1 Task 2 · 2026-08-08
**Authority:** [docs/plans/P1-flight-recorder-plan.md](../../../docs/plans/P1-flight-recorder-plan.md) § Architecture · [VD-stack-go-fid9mi](../../../docs/decisions/VD-stack-go-fid9mi.md) (blessed dependency set) · [docs/design/continuity-chain.md](../../../docs/design/continuity-chain.md) (this file is the behavior lock hop)
**Lock rule:** § 2 is the interface lock. Changing a signature is a reviewed diff to THIS FILE first, then code — never the reverse (`/vera-review` enforces).

---

## 1. Purpose

`core` owns the **pure, I/O-free primitives every other kernel package depends on**: the event
envelope, RFC-8785 (JCS) canonical JSON and content hashing, ULID event identity, the kinds
registry, envelope validation, and the idempotency key. It is the bottom of the dependency
graph — `store`, the connectors, `projections`, and `cmd/vera` all import it; it imports none
of them.

**core does NOT own:**

| Not owned | Home | Why not here |
|---|---|---|
| `seq` — the replay order | `internal/store` (`events.seq BIGSERIAL`) | Order is a *ledger* fact, assigned at append. An in-memory envelope has no position. |
| Persistence, the `*pgxpool.Pool`, migrations, the data-dir lock (`.vera/db.lock` — derived from the data directory, store's SPEC § 2.1) | `internal/store` (the ONLY package that opens the DB) | core opens no database, no file, no socket. |
| The UNIQUE `(source, native_id, content_sha)` index and the append semantics | `internal/store` SPEC | core *computes* the key; the ledger *enforces* it. |
| Payload schemas (commit fields, witness v1 JSON, session metadata) | each `internal/connector/*` SPEC | core treats `payload` as opaque JSON. |
| Projection reducers, `[superseded]` marking, week report | `internal/projections` | derived state, rebuildable. |
| Wall-clock reading | the composition root (`cmd/vera`) | the clock and the entropy source are injected (§ 2.4). |

Per the CLAUDE.md one-home table, `kernel/internal/<pkg>/SPEC.md` is the single home of a
package contract; code cites the spec, never restates it.

---

## 2. Interface

The complete exported surface of `package core`. Nothing else is exported.

> **This block drifted once and nothing caught it (2026-08-11).** `KindReviewVerdict` and
> `SourceReviews` shipped in `kinds.go` while § 2 still listed three of each — code moved before
> spec, the one direction the lock rule forbids, and § 3 of this same file already said FOUR kinds,
> so the file disagreed with itself. `store` has had a pinned-surface test since Task 3; `core` did
> not, so the lock was prose only. It is now asserted by
> `speclock_test.go::TestExportedSurfaceMatchesTheSpec`, which parses THIS block and compares it to
> the package's real exports.

### 2.1 Kinds registry and sources

```go
// Kind is a registered event kind. The string values are the stable identifiers
// P2 gate definitions match on — adding, renaming, or removing one is a spec diff.
type Kind string

const (
	KindCommitRecorded   Kind = "commit.recorded"
	KindCheckRun         Kind = "check.run"
	KindSessionObserved  Kind = "session.observed"
	KindReviewVerdict    Kind = "review.verdict"
	KindGitHubWorkflow   Kind = "github.workflow_run"
	KindGitHubDeployment Kind = "github.deployment"
)

// Registered reports whether k is a member of the kinds registry.
func (k Kind) Registered() bool

// Kinds returns the registry in lexical order. The caller receives a copy.
func Kinds() []Kind

// Source names the connector that observed the event. Unlike Kind, Source is an
// OPEN set (P3 adds external connectors) validated by shape, not by membership.
type Source string

const (
	SourceGit      Source = "git"
	SourceChecks   Source = "checks"
	SourceSessions Source = "sessions"
	SourceReviews  Source = "reviews"
	SourceGitHub   Source = "github"
)

// WellFormed reports whether s matches ^[a-z][a-z0-9_]{0,31}$.
func (s Source) WellFormed() bool
```

`Registered` and `WellFormed` are deliberately named differently: kind membership is closed,
source shape is open. A single `Valid()` on both would hide that asymmetry.

### 2.2 Canonical JSON and content hashing

```go
// ContentSHALen is the length of a ContentSHA string (lowercase hex sha256).
const ContentSHALen = 64

// MaxPayloadBytes caps a canonical payload. Connectors summarize; the ledger is not a blob store.
const MaxPayloadBytes = 1 << 20

// Canonicalize returns the RFC 8785 (JCS) canonical form of raw via github.com/gowebpki/jcs,
// after rejecting numbers that cannot survive the transform (see § 3, INV-5).
// Any valid JSON value is accepted — object, array, string, number, bool, null.
// Errors satisfy errors.Is(err, ErrCanonicalJSON).
func Canonicalize(raw json.RawMessage) ([]byte, error)

// ContentSHA returns the lowercase hex sha256 of Canonicalize(raw).
// It is the ONLY blessed way to compute content_sha.
func ContentSHA(raw json.RawMessage) (string, error)
```

### 2.3 Event identity

```go
// EventID is a ULID. It is IDENTITY ONLY — never an ordering key (§ 3, INV-27).
// The zero value is invalid and is what Validate rejects as "missing".
type EventID [16]byte

// ParseEventID parses the canonical 26-character Crockford base32 form.
// Errors satisfy errors.Is(err, ErrInvalidEventID).
func ParseEventID(s string) (EventID, error)

func (id EventID) String() string            // canonical 26-character form
func (id EventID) IsZero() bool
func (id EventID) MarshalText() ([]byte, error)
func (id *EventID) UnmarshalText(b []byte) error
```

No `Compare`, `Less`, `Before`, `After`, or `Timestamp` method exists — see Non-goals.

### 2.4 The generator (the single construction path)

```go
// IDGeneratorConfig injects the two impure inputs core refuses to acquire itself.
type IDGeneratorConfig struct {
	// Entropy is REQUIRED. The composition root passes crypto/rand.Reader; tests pass a
	// deterministic reader. core wraps it in a locked monotonic reader internally.
	Entropy io.Reader
	// Now is optional; nil means time.Now.
	Now func() time.Time
}

// NewIDGenerator returns a generator safe for concurrent use.
// A nil Entropy is an error (errors.Is(err, ErrInvalidConfig)).
func NewIDGenerator(cfg IDGeneratorConfig) (*IDGenerator, error)

type IDGenerator struct{ /* unexported */ }

// NewEventID mints a ULID from the injected clock and entropy.
func (g *IDGenerator) NewEventID() (EventID, error)

// NewEventParams is everything a connector knows; the generator supplies the rest.
type NewEventParams struct {
	Source           Source
	NativeID         string
	Kind             Kind
	OccurredAt       time.Time
	Payload          json.RawMessage
	ConnectorVersion string
}

// NewEvent is the ONLY blessed way to build an Event. It mints the EventID, stamps
// RecordedAt from the injected clock, canonicalizes the payload, computes ContentSHA
// exactly once from those canonical bytes, and returns Validate's error if any.
func (g *IDGenerator) NewEvent(p NewEventParams) (Event, error)
```

### 2.5 The envelope

```go
// Event is the ledger envelope. It carries NO seq — replay order belongs to store.
type Event struct {
	ID               EventID         `json:"event_id"`
	Source           Source          `json:"source"`
	NativeID         string          `json:"native_id"`
	Kind             Kind            `json:"kind"`
	OccurredAt       time.Time       `json:"occurred_at"`
	RecordedAt       time.Time       `json:"recorded_at"`
	Payload          json.RawMessage `json:"payload"`
	ContentSHA       string          `json:"content_sha"`
	ConnectorVersion string          `json:"connector_version"`
}

// Validate checks the SHAPE of every field (§ 3, INV-20/21). It reports ALL failures via
// errors.Join; every returned error satisfies errors.Is(err, ErrInvalidEvent).
// It does NOT re-derive ContentSHA from Payload — see INV-23.
func (e Event) Validate() error

// IdempotencyKey is the ledger's uniqueness tuple: (source, native_id, content_sha).
type IdempotencyKey struct {
	Source     Source
	NativeID   string
	ContentSHA string
}

func (e Event) IdempotencyKey() IdempotencyKey

// String renders "source/native_id/content_sha" for logs and test failures.
// Diagnostic only — there is no parser, and it is not a wire format.
func (k IdempotencyKey) String() string

// SameSubject reports whether both keys describe the same (source, native_id).
func (k IdempotencyKey) SameSubject(other IdempotencyKey) bool

// IsRevisionOf reports whether k is a REVISION of other: same subject, different
// content. Equal keys are re-ingests, not revisions.
func (k IdempotencyKey) IsRevisionOf(other IdempotencyKey) bool
```

### 2.6 Errors

```go
var (
	ErrInvalidEvent   = errors.New("core: invalid event")
	ErrCanonicalJSON  = errors.New("core: canonical json")
	ErrUnsafeNumber   = errors.New("core: json number outside exact integer range")
	ErrInvalidEventID = errors.New("core: invalid event id")
	ErrInvalidConfig  = errors.New("core: invalid config")
)

// ValidationError names the offending field. Unwrap returns ErrInvalidEvent.
// Field is ALWAYS the wire name (the struct's JSON tag: event_id, source, native_id,
// kind, occurred_at, recorded_at, payload, content_sha, connector_version) — never the
// Go field name. Errors cross process boundaries into logs, API responses and the
// ledger itself, so a Go-side rename must not change what a caller branches on (INV-30).
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string
func (e *ValidationError) Unwrap() error
```

`Canonicalize` reports an unsafe number with both sentinels (`fmt.Errorf("%w: ...: %w",
ErrCanonicalJSON, ErrUnsafeNumber)`), so a caller can test the general or the specific case.

### 2.7 Field rules (what Validate enforces)

| Field | Rule |
|---|---|
| `ID` | not the zero ULID |
| `Source` | `WellFormed()` — `^[a-z][a-z0-9_]{0,31}$` |
| `NativeID` | 1–512 bytes, valid UTF-8, no ASCII control characters, no leading/trailing space |
| `Kind` | `Registered()` |
| `OccurredAt` | non-zero |
| `RecordedAt` | non-zero |
| `Payload` | valid JSON **object** (`{…}`), non-empty, ≤ `MaxPayloadBytes` |
| `ContentSHA` | exactly `^[0-9a-f]{64}$` (lowercase) |
| `ConnectorVersion` | 1–64 bytes, no whitespace or control characters |

**Deliberately NOT enforced:** `RecordedAt >= OccurredAt`. Git committer dates come from other
machines' clocks; rejecting skew would drop true history. The recorder records what happened,
including clock skew.

---

## 3. Invariants

Each is a testable statement. The mapping to test names is § 5.

**Canonicalization and hashing**

1. **INV-1 — Order and whitespace independence.** Canonical bytes are identical for two inputs
   that differ only in object key order or insignificant whitespace.
2. **INV-2 — Idempotence.** `Canonicalize(Canonicalize(x)) == Canonicalize(x)`.
3. **INV-3 — Unicode fidelity.** Non-ASCII strings survive as UTF-8 unescaped; keys sort by
   UTF-16 code unit (`"Z" < "a" < "z" < "é"`), per RFC 8785 — not by Go byte order.
4. **INV-4 — Number normalization.** `-0` and `0` produce identical bytes; `1e2` and `100`
   produce identical bytes.
5. **INV-5 — Unsafe integers are rejected, never truncated.** An integer-valued literal with
   |value| > 2^53 (`9007199254740992`) is rejected with `ErrUnsafeNumber` + `ErrCanonicalJSON`;
   exactly 2^53 is accepted. *Verified 2026-08-08: `jcs.Transform` SILENTLY rewrites
   `9007199254740993` to `9007199254740992`. The guard is core's, not the library's — without
   it, two different payloads would share one content_sha.* Non-finite literals (`1e400`) are
   rejected too.
6. **INV-6 — Malformed input is rejected with no partial output.** Invalid JSON, trailing
   garbage, `NaN`, and duplicate object keys all return `ErrCanonicalJSON` and nil bytes.
7. **INV-7 — Structure preserved.** Nested objects canonicalize recursively; **array element
   order is preserved and never sorted**.
8. **INV-8 — Pinned vector (library-swap guard).** The § 4 fixture input yields exactly the
   pinned canonical bytes and the pinned sha256. A dependency swap or version bump that changes
   either fails here first.
9. **INV-9 — ContentSHA shape.** `ContentSHA` returns `ContentSHALen` lowercase hex characters
   and equals `sha256(Canonicalize(raw))`.
10. **INV-10 — ContentSHA discriminates.** Semantically equal payloads share a sha; any payload
    difference (including a single value change) produces a different sha.

**Identity**

11. **INV-11 — EventID text round-trip.** `ParseEventID(id.String()) == id`; `MarshalText` /
    `UnmarshalText` round-trip the same way.
12. **INV-12 — EventID parsing rejects malformed input.** Wrong length, non-Crockford
    characters, and empty string return `ErrInvalidEventID`.
13. **INV-13 — Generated ids are unique.** Two `NewEventID` calls within the same millisecond
    return different ids (monotonic entropy), and neither is the zero value.
14. **INV-14 — Pinned id vector.** With `Now` fixed to `2026-08-08T00:00:00Z` and an entropy
    reader emitting `0xAB`, the first minted id is exactly
    `01KZFAPQ00NENTQAXBNENTQAXB` — the ULID layout (48-bit ms + 80-bit entropy) is pinned.
15. **INV-15 — Entropy must be injected.** `NewIDGenerator` with a nil `Entropy` returns
    `ErrInvalidConfig`; core never acquires randomness on its own.

**Kinds registry**

16. **INV-16 — Closed registry.** `Registered()` is true for exactly the registry constants and
    false for anything else, including near-misses (`"commit.record"`, `"Commit.Recorded"`, `""`).
    The registry is FOUR kinds as of 2026-08-11 — do not restate the count in prose elsewhere;
    `Kinds()` is the single home and `len(core.Kinds())` is the answer.
17. **INV-17 — Stable strings, defensive copy.** The literal values are asserted verbatim (P2
    gates match them), `Kinds()` is lexically sorted, and mutating the returned slice does not
    affect a later call.

**Interface lock**

31. **INV-31 — § 2 IS the exported surface, mechanically.** The Kind and Source constants
    listed in § 2 are exactly those the package exports — checked by parsing both, not by
    review. § 2 drifted once (2026-08-11) because the lock was prose only while `store` had
    had a pinned-surface test since Task 3. Compares NAMES; a changed TYPE with an unchanged
    name is caught by INV-17's literal assertions instead.

**Why there is a fourth kind (added 2026-08-11).** `review.verdict` records that something was
FOUND, which no other kind does. The other three record that work happened (`commit.recorded`),
that checks ran (`check.run`), or that a session occurred (`session.observed`) — and across the
nine review rounds that built `internal/store`, `make check` was GREEN at every single commit
while adversarial review found one to three new defects per remediation cycle. A fix-induced
regression series derived from `check.run` alone would therefore read ZERO for the exact period it
is meant to describe. The kind is the minimum addition that makes the project's own most-cited
metric computable rather than aspirational (ROADMAP § P1).

**What core owns of it, and what it does not.** Core owns the kind's IDENTITY only. The payload —
finding id, severity, the commit under review, the commit that introduced the defect where known —
is opaque JSON here and is specified by whichever connector emits it, exactly as for the other
three (§ 1, "core treats `payload` as opaque JSON"). Core gains no knowledge of findings.

**Envelope**

18. **INV-18 — Hash computed once, from canonical bytes.** After `NewEvent`, `Event.Payload`
    equals `Canonicalize(input)` and `Event.ContentSHA` equals `ContentSHA(input)`; the stored
    payload is a fresh slice, so mutating the caller's input afterwards changes neither.
19. **INV-19 — Timestamps normalized.** `NewEvent` converts `OccurredAt` and `RecordedAt` to UTC
    and truncates both to microsecond precision (Postgres `timestamptz` resolution), so a
    round-trip through the ledger cannot alter them.
20. **INV-20 — Required fields.** `Validate` rejects a zero/empty value in every field of § 2.7,
    one table row per field, and accepts a fully populated event.
21. **INV-21 — Malformed values.** `Validate` rejects an unregistered kind, a malformed source,
    a non-object or oversized payload, an uppercase or short content_sha, and a native_id
    containing a control character.
22. **INV-22 — All failures reported.** An event with several bad fields yields a joined error
    naming every offending field, and `errors.Is(err, ErrInvalidEvent)` holds.
23. **INV-23 — Validate never re-derives content_sha.** An event whose payload no longer matches
    its content_sha still passes `Validate`. *Deliberate:* a payload read back from JSONB has
    been re-ordered by Postgres, so re-deriving would report false corruption. The hash is
    computed once at ingest (INV-18) and compared only against other hashes.
24. **INV-24 — No seq on the wire.** `Event` JSON round-trips all nine fields, and the marshalled
    object contains no `seq` key.

**Idempotency and revision**

25. **INV-25 — Key identity.** Two events built from the same `(source, native_id, payload)` have
    equal `IdempotencyKey`s regardless of `EventID`, `RecordedAt`, or payload key order.
26. **INV-26 — Revision semantics.** Same subject + different content_sha ⇒ `IsRevisionOf` true.
    Identical keys ⇒ false (a re-ingest, not a revision). Different source or native_id ⇒ false,
    and `SameSubject` false.

**Construction refuses invalid envelopes**

29. **INV-29 — `NewEvent` returns `Validate`'s error and builds nothing.** An unregistered kind,
    an empty or malformed `native_id`/`source`/`connector_version`, a zero `occurred_at`, or a
    non-object payload each make `NewEvent` fail; the returned `Event` is the zero value, not a
    partly-built envelope. *Added 2026-08-08 (reconciliation): § 2.4 already required this in
    prose ("returns Validate's error if any"), but no invariant pinned it, so the only construction
    path could have dropped its `Validate` call and let a connector mint an invalid event straight
    into the ledger with the whole suite still green.* Numbered 29 to keep 1–28 stable; it belongs
    beside INV-18/19 on the construction path.

**Package purity (mechanically enforced, not just asserted in prose)**

27. **INV-27 — No ordering surface.** A `go/parser` scan of core's non-test files finds no
    exported identifier named `Seq`, `Compare`, `Less`, `Sort`, `Order`, `Before`, or `After`,
    and no exported struct field named `Seq`. Ordering is store's job; core cannot even express it.
28. **INV-28 — No I/O, blessed imports only.** The same scan finds no import outside
    {standard library, `github.com/gowebpki/jcs`, `github.com/oklog/ulid/v2`} and none of
    `os`, `os/exec`, `net/…`, `database/sql`, `path/filepath`, `syscall`, `io/fs`, `embed`,
    `log`, `crypto/rand`. (`io` is allowed — `io.Reader` is the injected entropy seam.)

---

## 4. Pinned vectors

Copy these verbatim into `vector_test.go`. They were produced on 2026-08-08 against
`gowebpki/jcs v1.0.1` and `oklog/ulid/v2 v2.1.2`; a differing result means the library changed
behavior and the change must be reviewed, not absorbed.

**Canonical JSON + content_sha (INV-8).** Input (note the deliberately shuffled keys, unicode,
nested object, and array):

```json
{"subject":"feat: café ☕ ledger","sha":"9f2b7c1d","files":[{"path":"b.go","added":3},{"path":"a.go","added":10}],"stats":{"insertions":13,"deletions":0},"vd_ids":["VD-<slug>-????"],"signed":false,"parent":null}
```

Canonical bytes:

```json
{"files":[{"added":3,"path":"b.go"},{"added":10,"path":"a.go"}],"parent":null,"sha":"9f2b7c1d","signed":false,"stats":{"deletions":0,"insertions":13},"subject":"feat: café ☕ ledger","vd_ids":["VD-<slug>-????"]}
```

`content_sha = d5ebe7c20d9be166eaf39bb22e07263a0d4fd46a5c137ec7d1ea474a554522d4`

The `vd_ids` entry uses the plan's placeholder form `VD-<slug>-????` deliberately: a real
decision id inside a **pinned** fixture would couple the hash to a filename, so renaming a
decision would force a fixture edit and a pinned-value change. The placeholder does not match
link-lint's `VD-[a-z0-9-]+-[a-z0-9]{6}` pattern, so it is inert to the ref scan.

The same document with every key order reversed canonicalizes to those identical bytes (INV-1);
note that `files` keeps `b.go` before `a.go` — arrays are not sorted (INV-7).

**Event id (INV-14).** `Now = 2026-08-08T00:00:00Z` (ULID ms `1786147200000`), entropy reader
emitting `0xAB` forever ⇒ first id `01KZFAPQ00NENTQAXBNENTQAXB`.

> **Trap, verified 2026-08-08 — do not use an all-zero entropy reader in tests.**
> `ulid.Monotonic` derives its per-call increment from the entropy source, so a reader emitting
> `0x00` produces the *same id twice* in one millisecond and INV-13 would fail for the wrong
> reason. Use a non-zero constant reader (`0xAB`) for deterministic tests; the monotonic
> increment is then deterministic and strictly increasing.

**Number boundary (INV-5).** `9007199254740992` accepted → `{"n":9007199254740992}`;
`9007199254740993` rejected; `1e17` rejected (integer-valued, > 2^53); `0.1` accepted → `{"n":0.1}`;
`-0` accepted → `{"n":0}`; `1e400` rejected.

---

## 5. Invariant table

Format: `| INV-<n> | <statement> | <test file>::<TestName> |` — the pinned rule and its
rationale live in `.claude/commands/vera-spec.md` § 5 (single home; do not restate it here).
One row per invariant; the third cell names a real Go test function in this package.
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
| INV-1 | Canonical bytes are independent of object key order and insignificant whitespace | core_test.go::TestCanonicalize_KeyOrderAndWhitespaceIndependent |
| INV-2 | Canonicalization is idempotent | canonical_test.go::TestCanonicalize_Idempotent |
| INV-3 | Non-ASCII strings survive as UTF-8 and keys sort by UTF-16 code unit | canonical_test.go::TestCanonicalize_Unicode |
| INV-4 | -0 equals 0 and exponent forms normalize to their plain form | canonical_test.go::TestCanonicalize_NumberNormalization |
| INV-5 | Integer literals beyond 2^53 and non-finite literals are rejected, never truncated | canonical_test.go::TestCanonicalize_RejectsUnsafeNumbers |
| INV-6 | Invalid JSON, trailing garbage, and duplicate keys are rejected with no output | canonical_test.go::TestCanonicalize_RejectsMalformedJSON |
| INV-7 | Nested structures canonicalize recursively and array order is preserved | canonical_test.go::TestCanonicalize_NestingAndArrayOrder |
| INV-8 | The pinned input yields the pinned canonical bytes and the pinned sha256 | vector_test.go::TestPinnedVector_CanonicalBytesAndSHA |
| INV-9 | ContentSHA is 64 lowercase hex characters equal to sha256 of the canonical bytes | canonical_test.go::TestContentSHA_Shape |
| INV-10 | Equal payloads share a content_sha and any difference changes it | canonical_test.go::TestContentSHA_DistinguishesPayloads |
| INV-11 | EventID round-trips through String, ParseEventID, MarshalText and UnmarshalText | id_test.go::TestEventID_TextRoundTrip |
| INV-12 | Malformed event id text is rejected with ErrInvalidEventID | id_test.go::TestEventID_ParseRejectsMalformed |
| INV-13 | Ids minted within one millisecond are distinct and never the zero value | id_test.go::TestIDGenerator_UniqueWithinMillisecond |
| INV-14 | Fixed clock plus fixed entropy yields the pinned event id | vector_test.go::TestPinnedVector_EventID |
| INV-15 | NewIDGenerator requires injected entropy and errors on nil | id_test.go::TestIDGenerator_RequiresEntropy |
| INV-16 | Kind.Registered is true for exactly the registry members | kinds_test.go::TestKind_RegisteredOnlyForRegistryMembers |
| INV-17 | Kind literal strings are stable and Kinds returns a sorted defensive copy | kinds_test.go::TestKinds_StableStringsAndDefensiveCopy |
| INV-31 | SPEC § 2's const blocks name exactly the package's exported Kind and Source constants | speclock_test.go::TestExportedSurfaceMatchesTheSpec |
| INV-18 | NewEvent stores canonical payload bytes and computes content_sha exactly once | event_test.go::TestNewEvent_CanonicalPayloadAndSingleHash |
| INV-19 | NewEvent normalizes both timestamps to UTC at microsecond precision | event_test.go::TestNewEvent_NormalizesTimestamps |
| INV-20 | Validate rejects a zero or empty value in every required field | event_test.go::TestValidate_RejectsEmptyRequiredFields |
| INV-21 | Validate rejects malformed kind, source, payload, content_sha and native_id values | event_test.go::TestValidate_RejectsMalformedValues |
| INV-22 | Validate reports every offending field and its error satisfies errors.Is ErrInvalidEvent | event_test.go::TestValidate_JoinsAllFailures |
| INV-23 | Validate does not re-derive content_sha from the payload | event_test.go::TestValidate_DoesNotRederiveContentSHA |
| INV-24 | Event JSON round-trips all nine fields and emits no seq key | event_test.go::TestEvent_JSONRoundTripHasNoSeq |
| INV-25 | Identical source, native_id and payload yield an identical idempotency key | idempotency_test.go::TestIdempotencyKey_EqualForIdenticalContent |
| INV-26 | Same subject with a different content_sha is a revision; an identical key is not | idempotency_test.go::TestIdempotencyKey_RevisionSemantics |
| INV-27 | The package exports no ordering symbol and no Seq field | surface_test.go::TestNoOrderingSurface |
| INV-28 | The package imports only stdlib plus the two blessed dependencies and no I/O package | surface_test.go::TestOnlyBlessedImports |
| INV-29 | NewEvent returns Validate's error and returns the zero Event, never a partial one | event_test.go::TestNewEvent_RejectsInvalidEnvelope |
| INV-30 | ValidationError.Field is always the JSON-tag (wire) name, never the Go field name | event_test.go::TestValidationError_FieldUsesWireNames |

Tests are written from this table before implementation, failing rather than fake-passing
(no stub assertions).

---

## 6. Non-goals

A reviewer rejects these on sight:

- **No `seq`, no ordering.** No `Seq` field, no `Compare`/`Less`/`Sort`, no "sort events by
  event_id". ULIDs are lexically sortable — that is a trap, not a feature. Replay order is
  `store.events.seq`, always (INV-27).
- **No time-based ordering.** `occurred_at` is evidence about the world, not a position;
  connectors ingest clock-skewed timestamps and the recorder keeps them.
- **No I/O of any kind.** No file reads, no `os`, no network, no `database/sql`, no pgx, no
  goose, no logging. The clock and the entropy source are injected (INV-28).
- **No DB types.** No `pgtype`, no `sql.NullString`, no struct tags for a driver. The envelope
  is plain Go plus `encoding/json`.
- **No hand-rolled canonicalization.** JCS comes from `gowebpki/jcs`. A hand-written key sorter
  "because it's simple" is rejected — RFC 8785's number and UTF-16 sorting rules are where it
  would silently go wrong.
- **No re-derivation of content_sha from stored payloads** (INV-23), and no second hashing
  helper that would let a caller do it accidentally.
- **No payload schema knowledge.** core never parses a commit sha, a witness field, or a session
  count out of `payload`. That belongs to each connector's SPEC.
- **No connector or store constructors, no config file loading, no CLI flags.**
- **No new dependency.** § 7 is the whole list. Anything else needs a `/vera-decide` record
  first (Build Law 8).

**Known limitation, stated rather than hidden:** RFC 8785 defines JSON numbers as IEEE-754
doubles, so a non-integer literal such as `1.0000000000000000005` canonicalizes to `1` with no
error. **Underflow is the same class and is equally unguarded, verified 2026-08-08:
`strconv.ParseFloat("1e-400", 64)` returns `0` with a NIL error (it is finite, so the non-finite
guard never fires), so `{"n":1e-400}` canonicalizes to `{"n":0}` and shares a content_sha with
`{"n":0}`.** core guards the *integer* range (INV-5) because that is where silent collisions
occur at values a connector plausibly emits; connectors MUST NOT put high-precision decimals or
subnormal magnitudes in payloads. If either becomes a real risk, the fix is a spec diff here
(widen INV-5 to reject a non-zero literal that rounds to zero), not a quiet change in a connector.

**Delegated constraint (for the store SPEC, Task 3):** `content_sha` covers the payload only —
not `kind`. The ledger key is `(source, native_id, content_sha)`, so a connector that emitted
two different kinds for one `(source, native_id)` with an identical payload would have the
second silently absorbed as a duplicate. No P1 connector does this (each owns one kind), but
store's SPEC should either restate the one-kind-per-subject obligation or widen the index.

---

## 7. Dependencies

**Standard library** — `encoding/json`, `crypto/sha256`, `encoding/hex`, `errors`, `fmt`, `io`,
`math/big` (exact integer classification for INV-5), `strconv`, `strings`, `unicode/utf8`,
`time`, `sync`, `regexp` or hand-rolled shape checks; `go/parser` + `go/ast` in tests only.

**Blessed external dependencies (VD-stack-go-fid9mi), both already in the P1 set:**

| Module | Version verified | Used for |
|---|---|---|
| `github.com/gowebpki/jcs` | v1.0.1 | `jcs.Transform` — RFC 8785 canonicalization |
| `github.com/oklog/ulid/v2` | v2.1.2 | ULID minting (`New`, `Timestamp`, `Monotonic`, `LockedMonotonicReader`) and parsing |

**That is the entire list.** core adds no third dependency; any addition requires a `VD-` record
before `go get` (Build Law 8), and INV-28 fails the build if one appears without one.
