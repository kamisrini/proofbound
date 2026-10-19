# P7 replay lock — proposed `twin` API/SPEC amendment

**Status:** ACCEPTED FOR DOCUMENTARY PLANNING — round-1 non-author review on 2026-10-19;
implementation separately gated. This remains a proposed companion contract and does not modify the live
[`internal/twin` SPEC](../../kernel/internal/twin/SPEC.md), add product code, authorize
runtime fixtures, validate PostgreSQL, or reopen dormant deferral `DEF-P7-001`.

**Authority:** Read with the accepted [first implementation-slice proposal](P7-replay-lock-first-implementation-slice-proposal-draft1.md),
[replay-lock addendum](P7-replay-lock-spec-addendum-draft1.md),
[external-prerequisites amendment](P7-replay-lock-external-prerequisites-amendment-draft1.md),
and [readiness-manifest plan](P7-readiness-manifest-plan-draft1.md). The accepted planning
documents remain documentary only; this draft makes the proposed companion boundary precise
enough for a separate API/SPEC review.

## Current live contract remains unchanged

The live package continues to expose the accepted P4 API:

```go
func ReplayIsolated(
    context.Context,
    *store.Store,
    Request,
    projections.Snapshot,
) (Result, error)
```

It continues to use the existing store configuration and projector path. It does not claim to
verify a third-party PostgreSQL archive, prove binary provenance, emit committed proof bytes, or
evaluate a P7 behavior lock. No caller may infer those claims from this proposed amendment.

## Proposed companion boundary

The proposed P7-only companion receives exact PostgreSQL archive bytes prepared by the separately
reviewed Phase-A setup route. It receives bytes and declared identities, never a URL, download
instruction, ambient cache, pre-staged directory, or runtime path.

Illustrative API shape for review:

```go
type VerifiedArchiveEntry struct {
    Path   string // archive-relative slash path
    Mode   string // exactly 100644 or 100755
    SHA256 string // lowercase SHA-256 of the extracted regular file
}

type VerifiedPostgresInput struct {
    ArchiveBytes       []byte
    ArchiveSHA256      string
    PostgresBinarySHA256 string
    Inventory          []VerifiedArchiveEntry
}

func ReplayIsolatedWithVerifiedPostgres(
    context.Context,
    *store.Store,
    Request,
    projections.Snapshot,
    VerifiedPostgresInput,
) (Result, error)
```

The names and exact Go signature are proposed, not live. The separate review must confirm whether
the inventory belongs in this value or in a content-addressed evidence object resolved before the
call. Either choice must preserve the same byte and identity checks below. The API must not accept
both alternatives, silently infer missing inventory, or weaken a failed check to a warning.

## Required companion behavior

Before a successful result, the companion must:

1. Require a nonempty archive byte slice, lowercase SHA-256 values, and a complete, sorted,
   duplicate-free inventory. Reject path traversal, absolute paths, unsafe links, unsupported
   file types, invalid modes, and malformed digests.
2. Hash the exact supplied archive bytes and require equality with `ArchiveSHA256`. A digest,
   path, or setup directory without the bytes is insufficient.
3. Extract into a fresh temporary root owned by this replay. Verify the complete extracted set,
   modes, regular-file digests, safe link targets if links are permitted by the reviewed archive
   contract, and required `bin/pg_ctl` and server executable before `store.Open`.
4. Require the actual launched server executable to hash to `PostgresBinarySHA256`. The value is
   a binding to the executable used by the replay, not merely a declaration copied from setup.
5. Pass the staged binary directory as `store.Config.BinariesDir`. Never pass it as
   `RuntimeDir`, because store startup clears the runtime directory.
6. Open only the temporary isolated store, apply the existing P4 replay path, and preserve the
   existing `Result`, `Proof`, snapshot, sequence-map, and verdict meanings. This amendment does
   not add a lock evaluator or change the `vera.replay.v1` proof encoding.
7. Close and remove the entire temporary root, including staged binaries, on success, context
   cancellation, archive failure, store-open failure, projector failure, and any later error.
8. Make no DNS, non-loopback connection, download, cache fallback, or remote lookup. A failed
   network request is not evidence of zero attempts; the fixture must observe the process and its
   children under the declared network policy.

The source store remains read-only. The companion must not mutate the production ledger or
projection tables. The complete ledger before and after every success or failure call must remain
byte-for-byte equal under the fixture's exclusive-writer boundary, including records beyond
`ThroughSeq`.

## Failure and ownership rules

| Condition | Required result | Ownership rule |
|---|---|---|
| Archive digest mismatch | Error; no successful `Result` | Caller retains its input bytes; companion must not use them |
| Missing, extra, unsafe, or changed extracted entry | Error before successful replay | Temporary extraction root is removed |
| Missing `bin/pg_ctl` or server executable | Error before `store.Open` | A populated ambient cache cannot satisfy the request |
| Launched server digest mismatch | Error; result is unverified | The fixture records the actual executable identity |
| Network/cache attempt | Error or unverified case; never a pass | Witness covers replay and child processes |
| Projector/store/context error | Error; no successful `Result` | Source ledger must still match before/after bytes |
| Cleanup failure | Error and unverified case | Preserve cleanup evidence; do not claim isolation |
| Valid input and successful replay | Existing `Result` shape only | Temporary root is removed before return |

The companion may copy the supplied bytes before extraction; it must not mutate the caller's byte
slice. The API must define whether the input is retained during the call or copied immediately,
and tests must cover cancellation during extraction and after store open. No input path or URL may
be retained as a later fallback.

## Proposed invariant additions

These are proposed P7 invariants, not yet live `TWIN-INV-*` entries:

- **P7-TWIN-1 — Exact archive binding:** consumed archive bytes equal the declared digest.
- **P7-TWIN-2 — Complete safe staging:** extracted paths, modes, digests, and required executables
  equal the reviewed inventory before store open.
- **P7-TWIN-3 — Executable binding:** the launched PostgreSQL executable equals the declared digest.
- **P7-TWIN-4 — Offline provisioning:** the companion has no network or ambient-cache route.
- **P7-TWIN-5 — Binary-directory correctness:** verified binaries are supplied through `BinariesDir`,
  never `RuntimeDir`.
- **P7-TWIN-6 — Cleanup closure:** every success and failure path removes the owned temporary root,
  or the case is unverified.
- **P7-TWIN-7 — Source preservation:** complete before/after ledger bytes are equal under an
  exclusive-writer boundary, including records after `ThroughSeq`.

## Review questions and stop points

The separate non-author API/SPEC review must decide:

1. Whether the companion receives an in-memory `Inventory` as shown or a reviewed,
   content-addressed inventory artifact.
2. The exact archive format and safe-link policy for the embedded-Postgres distribution.
3. The exact server executable path and how the launched executable is observed and hashed.
4. The temporary-root ownership and cleanup contract under cancellation and partial extraction.
5. The final exported/unexported API names and the exact live `twin` SPEC wording.

Stop before implementation if any answer requires a URL, path-only trust, ambient cache, network
fallback, `RuntimeDir`, a digest without bytes, or a weaker cleanup/ledger witness. The deferred
third-party provenance/licensing go/no-go remains outside this document; this draft neither
approves nor rejects the pinned PostgreSQL input.

## Explicit non-goals

This draft does not change the live API, implement a parser, create runtime fixtures, regenerate
snapshots, add a provider or external feed, add an event kind, evaluate a lock, claim L12/L17,
approve a third-party binary, or reopen `DEF-P7-001`.

## Review outcome

Round-1 non-author documentary review by Srini K accepted this planning boundary with no findings.
See [`p7-replay-lock-companion-round1-20261019.md`](../verification/verdicts/p7-replay-lock-companion-round1-20261019.md).
The acceptance is limited to documentary planning. The separately authorized preflight/staging
implementation slice was independently accepted with no findings by Srini K in
[`p7-replay-lock-companion-implementation-round1-20261019.md`](../verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md).
Runtime fixtures, PostgreSQL execution, behavior-lock evidence, and third-party binary
provenance/licensing remain separately gated.
