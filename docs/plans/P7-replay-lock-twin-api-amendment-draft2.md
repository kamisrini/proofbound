# P7 replay lock — companion API/SPEC boundary revision

**Status:** PROPOSED FOR THIRD-ROUND INDEPENDENT DOCUMENTARY REVIEW — 2026-10-24. This revision resolves
the five open questions from draft 1. It does not modify the live `twin` SPEC, authorize runtime
fixtures, validate PostgreSQL, approve `DEF-P7-001`, or claim L12/L17 or behavior-lock acceptance.

**Supersedes for review only:**
[`P7-replay-lock-twin-api-amendment-draft1.md`](P7-replay-lock-twin-api-amendment-draft1.md).
Draft 1 remains the historical planning record. This document is the proposed boundary to review;
the accepted fixture plan remains unchanged.

## Recommended boundary

Use an in-memory, caller-supplied inventory in `VerifiedPostgresInput`. Do not introduce a second
content-addressed inventory object into the API. The setup evidence and case manifest may bind the
inventory before the call, but the companion receives one complete value and verifies it against the
exact archive bytes it consumes.

The proposed names and signature are:

```go
type VerifiedArchiveEntry struct {
    Path   string // archive-relative slash path
    Mode   string // exactly 100644 or 100755
    SHA256 string // lowercase SHA-256 of the extracted regular file
}

type VerifiedPostgresInput struct {
    ArchiveBytes          []byte
    ArchiveSHA256         string
    PostgresBinarySHA256  string
    Inventory             []VerifiedArchiveEntry
}

func ReplayIsolatedWithVerifiedPostgres(
    context.Context,
    *store.Store,
    Request,
    projections.Snapshot,
    VerifiedPostgresInput,
) (Result, error)
```

These names remain implementation-scoped until a separately authorized live SPEC/API adoption.
No alternate path-only, URL, cache, or content-addressed-inventory parameter is accepted by this
boundary.

## Resolved contract questions

### 1. Inventory ownership

`Inventory` is the single API-level inventory value. It must be complete, sorted by path, duplicate
free, and contain only lowercase SHA-256 values and modes `100644` or `100755`. The caller may
derive it from a reviewed content-addressed evidence record before calling the companion, but that
evidence record is not a second runtime input and is not silently resolved by the companion.

The companion hashes the supplied `ArchiveBytes`, extracts them, and checks the extracted regular
files against `Inventory`. A digest, path, setup directory, or evidence-record identifier without
the exact bytes cannot satisfy the call.

### 2. Archive format and link policy

The archive is the exact caller-supplied `.txz`: exactly one xz stream whose decoder consumes every
input byte. Concatenated xz streams, xz stream padding, trailing bytes, and truncated or
checksum-invalid streams are rejected. The decompressed payload is one POSIX `ustar` tar stream. It
is an exact sequence of 512-byte blocks with POSIX `ustar\0` magic and version `00` on every entry
header, valid POSIX ASCII-octal numeric fields only (no base-256, signed, overflowing, or malformed
encodings), and a valid header checksum computed with the checksum field treated as spaces. PAX and
GNU extension headers, long-name/long-link records, sparse records, and unknown header types are
rejected. The reader must consume the entire decompressed payload: it requires exactly two
consecutive zero blocks as the final blocks, rejects a partial block, an extra end block, any
trailing bytes, or a concatenated tar payload, and accepts no bytes after the final end block. Header
metadata other than type, path, mode, and size is not trusted or used.

For each header, the POSIX `prefix` and `name` fields are combined with one slash when `prefix` is
nonempty, then validated as a valid UTF-8 slash path with no NUL, backslash, leading slash, empty
segment, `.` segment, or `..` segment. File paths have no trailing slash. Directory paths may have
exactly one trailing slash, which is removed for comparison; the root directory is not an entry.
After that normalization, every path is nonempty and unique across both files and directories, and
a path cannot be both a file and a directory. Directory entries have zero size. Implicit parent
directories are allowed only when no file path is equal to an implicit parent path; a regular file
that is an ancestor of another entry is rejected. Symlinks, hard links, devices, FIFOs, other tar
types, duplicate paths, absolute paths, and unsupported modes are rejected. There is no safe-link
allowance in this boundary.

Required regular files are `bin/pg_ctl` and `bin/postgres`, both mode `100755`. All other files
must be present in the declared inventory with matching mode and digest. Extra, missing, changed,
or non-regular entries fail closed before a successful replay result.

### 3. Executable path and observation

The canonical staged paths are `bin/pg_ctl` and `bin/postgres` beneath the replay-owned
`BinariesDir`. The preflight seam hashes the staged `bin/postgres` bytes against
`PostgresBinarySHA256` before `store.Open` and passes the staged directory through
`store.Config.BinariesDir`, never `RuntimeDir`.

The phrase “launched executable” is reserved for a future runtime witness. The current
preflight/staging acceptance does not prove that the daemonized PostgreSQL process actually
launched by the embedded-store layer executed those exact bytes. A future runtime acceptance must
observe the launched process and its children, bind the observed executable to the declared digest,
and record the witness. Until that exists, TWIN-INV-14 is only a proposed runtime requirement,
not a runtime acceptance claim.

### 4. Temporary-root ownership and cleanup

The companion creates a fresh temporary root and owns it for the entire call. It creates the
staging directory beneath that root, supplies it as `BinariesDir`, and lets the store derive
separate data and runtime paths beneath the same root subject to the store non-overlap rules. The
caller supplies no directory and receives no reusable path.

Cleanup is unconditional and covers archive validation/staging, cancellation, partial extraction,
store-open failure, import/projector/snapshot failure, comparison failure, success, and close
failure. The companion attempts every required store close before attempting root removal, even
when close fails. Cancellation does not suppress or interrupt close or root removal; cleanup runs
synchronously after cancellation. If both the primary operation and cleanup fail, the returned
error preserves both failures, with the primary failure retained as the primary cause. If cleanup
alone fails, the call returns that cleanup error and no successful result; a root that could not be
removed leaves the case unverified. Cleanup failure is never downgraded to a warning.
`ArchiveBytes` are consumed synchronously, never mutated, and not retained after return.

### 5. Names and live SPEC wording

Retain the names `VerifiedArchiveEntry`, `VerifiedPostgresInput`, and
`ReplayIsolatedWithVerifiedPostgres`. Do not add a new package, wire identity, proof schema, event
kind, or P7 lock evaluator.

The exact wording proposed for a future live SPEC amendment is:

> The companion consumes exact caller-supplied `.txz` bytes and one complete regular-file
> inventory. It hashes the supplied bytes, safely stages only regular files under a fresh
> replay-owned root, requires executable `bin/pg_ctl` and `bin/postgres`, verifies the staged
> `bin/postgres` digest before store open, and supplies that directory as `BinariesDir`. It
> accepts no URL, path-only input, ambient cache, or network fallback. Runtime acceptance must
> additionally observe the launched PostgreSQL executable and bind it to the declared digest;
> the preflight/staging seam alone does not establish that runtime fact.

No live SPEC edit is part of this revision. Adoption requires a separate implementation/API
authorization and a fresh review of the runtime evidence contract.

The reviewed live `kernel/internal/twin/SPEC.md` currently contains the proposed `TWIN-INV-14`
phrase that the companion “binds the launched” digest. For this documentary boundary, that phrase
is a future runtime requirement only and is not evidence that the current implementation observes
the launched process. It must not be cited as runtime acceptance. Correcting or adopting that live
SPEC wording requires a separately authorized SPEC change and fresh review.

## Unchanged gates and non-goals

The draft-1 required behavior, failure rules, P4 replay semantics, source-ledger preservation,
offline policy, and proposed P7 invariants remain in force except for the clarifications above.
The following remain outside this revision:

- real PostgreSQL runtime fixtures, process/network/cleanup witnesses, or binary validation;
- L12/L17 runtime results, behavior-lock acceptance, complete projector/build evidence, or an
  independent runtime verdict;
- publisher identity, detached-signature resolution, or license/notice coverage for the pinned
  PostgreSQL JAR; `DEF-P7-001` remains dormant and unapproved;
- any provider, external feed, downloaded artifact, runtime fixture file, or production code;
- approval of the nested binaries inside the already reviewed Go source corpus for any runtime use.

## Review stop point

The independent reviewer should accept this revision only if the five resolutions above are
precise, the current preflight/staging scope is not mistaken for runtime acceptance, and no path
exists to weaken exact-byte, safe-extraction, offline, source-preservation, or cleanup guarantees.
Otherwise return `AMEND` with exact findings. Do not fill a verdict in this planning document.
