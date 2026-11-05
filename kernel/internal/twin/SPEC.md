# SPEC — `internal/twin`

**Status:** implemented P4 slice plus adopted P7 verified-archive preflight/staging API/SPEC
boundary; runtime acceptance pending · 2026-10-24
**Authority:** [ROADMAP.md](../../../ROADMAP.md) § P4 · [CLAUDE.md](../../../CLAUDE.md) Law 6 ·
[P7-SPEC-ADOPT-1](../../../docs/intent/records/change-intents/CI-p7-spec-adopt-1a2b3c/0001.md) ·
[P7 companion boundary decision](../../../docs/decisions/VD-p7-replay-lock-companion-api-boundary-2026-10-23.md)

## Purpose

`twin` constructs a bounded, ephemeral candidate event stream. It never writes the
production ledger or projection tables. `Replay` injects a projector for fast tests;
`ReplayIsolated` runs the real projector in a temporary embedded store.

The authorized P7 non-runtime manifest preparation contract is implemented by
[`internal/twin/manifest`](manifest/SPEC.md); it declares source/toolchain/build identities only
and does not retrieve, build, execute, or validate runtime inputs.

## Interface

```go
type Candidate struct { Seq int64; Event core.Event }
type Request struct { ThroughSeq int64; Candidate []Candidate }
type SequenceBinding struct { EventID core.EventID; Requested, Effective int64 }
type Proof struct { Schema string; ThroughSeq int64; SourceDigest, CandidateDigest string; BaselineSnapshotDigest, CandidateSnapshotDigest string }
type VerifiedArchiveEntry struct { Path, Mode, SHA256 string }
type VerifiedPostgresInput struct {
    ArchiveBytes []byte
    ArchiveSHA256, PostgresBinarySHA256 string
    Inventory []VerifiedArchiveEntry
}
type Forecast struct { ID string; Probability float64; Outcome bool }
type Calibration struct { Count int; BrierScore, MeanProbability, ObservedRate float64 }
type Result struct {
    ThroughSeq int64
    SourceEvents int
    CandidateEvents int
    Baseline projections.Snapshot
    Candidate projections.Snapshot
    Verdict Verdict
    SequenceMap []SequenceBinding
    Proof Proof
}
type Verdict string
const VerdictPreserved Verdict = "preserved"
const VerdictChanged Verdict = "changed"
func Replay(context.Context, *store.Store, Request, projections.Snapshot,
    func(context.Context, []store.Record) (projections.Snapshot, error)) (Result, error)
func ReplayIsolated(context.Context, *store.Store, Request, projections.Snapshot) (Result, error)
func ReplayIsolatedWithVerifiedPostgres(context.Context, *store.Store, Request,
    projections.Snapshot, VerifiedPostgresInput) (Result, error)
func Calibrate([]Forecast) (Calibration, error)
```

The adopted P7 companion consumes exact caller-supplied `.txz` bytes and a complete regular-file
inventory. It hashes the bytes, safely extracts only regular files into a fresh replay root,
requires executable `bin/pg_ctl` and `bin/postgres`, verifies the staged `bin/postgres` digest
before store open, and passes the staged directory as `store.Config.BinariesDir`. It accepts no
URL, path-only input, ambient cache, or network fallback. While the real postmaster is live it now
observes the launched executable image at `/proc/<pid>/exe` and binds it to the declared digest
(TWIN-INV-18, the single sole-producer witness on the `ReplayIsolatedWithVerifiedPostgres` path),
closing the launched-process gap the staging re-hash could not. The remaining P7 B1 unit-3 runtime
witnesses — no-fetch isolation, whole-ledger preservation, cleanup absence, and real-proof
determinism (TWIN-INV-19..22) — and the case-manifest/behavior-lock authoring over the real proof
are the in-progress increments of that unit.

## Invariants

1. **TWIN-INV-1 — Production preservation.** Replay only reads the source store.
2. **TWIN-INV-2 — Bounded source.** No source event after `ThroughSeq` reaches the projector.
3. **TWIN-INV-3 — Ordered replay.** Candidate sequences strictly follow the selected prefix.
4. **TWIN-INV-4 — Immutable source.** Source records are copied before candidate assembly.
5. **TWIN-INV-6 — Deterministic result.** Equal inputs and projector output yield equal verdicts.
6. **TWIN-INV-7 — Fail closed.** Invalid requests and projector errors return no successful result.
7. **TWIN-INV-8 — Ephemeral isolation.** `ReplayIsolated` projects only in a temporary store and removes it on return.
8. **TWIN-INV-9 — Bound proof.** Every successful replay result carries a deterministic `vera.replay.v1` digest for its source, candidates, and snapshots.
9. **TWIN-INV-10 — Honest calibration.** Calibration rejects empty, malformed, duplicate, or out-of-range feed records and computes the Brier score over the supplied outcomes.
10. **TWIN-INV-11 — Complete route proof.** The frozen replay proof binds every registered
    source-kind pair without rewriting its identity.
11. **TWIN-INV-12 — Exact archive binding.** The companion consumes archive bytes whose SHA-256
    equals the declared identity.
12. **TWIN-INV-13 — Complete safe staging.** Extracted regular-file paths, modes, and digests
    equal the reviewed inventory; unsafe paths, links, and unsupported file types fail closed.
13. **TWIN-INV-14 — Executable binding.** The adopted preflight/staging seam requires executable
    `bin/pg_ctl` and `bin/postgres`, and the staged `bin/postgres` bytes equal the declared digest.
    This is a staging (pre-launch) check only; the runtime fact that the *launched* `postgres` bytes
    equal the declared digest is established by TWIN-INV-18, not by this on-disk re-hash.
14. **TWIN-INV-15 — Offline provisioning.** The companion has no URL, download, or ambient-cache
    fallback route.
15. **TWIN-INV-16 — Binary-directory correctness.** Verified binaries are supplied through
    `BinariesDir`, never `RuntimeDir`.
16. **TWIN-INV-17 — Cleanup closure.** The owned temporary root is removed on success and every
    failure path.
17. **TWIN-INV-18 — Launched-process witness.** On the `ReplayIsolatedWithVerifiedPostgres` path,
    while the real postmaster is live, the running executable image resolved from `/proc/<pid>/exe`
    (postmaster PID located under the replay-owned temporary root; the `(deleted)` marker rejected)
    is hashed and must equal the declared `PostgresBinarySHA256`, or the replay fails with a distinct
    launched-process mismatch error (separate from the TWIN-INV-14 staging re-hash error). The single
    `witnessLaunchedProcess` function is the sole producer of the resolved digest surfaced on the
    result for `proof_binding` evidence — never a literal. Its discriminating power is calibrated by a
    direct unit test against a live process running a byte-divergent executable (unit-3 spec §5/§5.1).
    *(TWIN-INV-5 is a retired tombstone; numbers are never reused.)*

## Non-goals

This slice does not fork PostgreSQL, execute arbitrary code, persist verdicts, or score
predictions. Those require a separate decision and acceptance evidence.

## Proving tests

| Invariant | Statement | Proving test |
|---|---|---|
| TWIN-INV-1/2/3/4 | replay bounds, ordering, and preservation | replay_test.go::TestReplayBoundsAndOrdersStream |
| TWIN-INV-6 | deterministic result | replay_test.go::TestReplayIsDeterministic |
| TWIN-INV-7 | invalid requests fail closed | replay_test.go::TestReplayRejectsDecreasingCandidates |
| projector failure | projector errors fail closed | replay_test.go::TestReplayFailsClosedOnProjectorError |
| TWIN-INV-8 | isolated replay projects candidates without changing the source | replay_test.go::TestReplayIsolatedProjectsCandidateAndPreservesSource |
| TWIN-INV-8 | isolated replay preserves gapped multi-event sequence identity | replay_test.go::TestReplayIsolatedSupportsMultipleGappedCandidates |
| TWIN-INV-9 | successful replay includes deterministic proof metadata | replay_test.go::TestReplayIsDeterministic |
| TWIN-INV-10 | calibration validates records and computes the Brier score | calibration_test.go::TestCalibrateComputesBrierScoreAndRates |
| TWIN-INV-11 | frozen proof covers every registered source-kind pair | replay_test.go::TestReplayProofCoversEveryRegisteredEventPair |
| TWIN-INV-12/13/14 | exact archive digest, safe extraction, inventory, required executables, and staged executable binding | verified_archive_test.go::TestStageVerifiedPostgresBindsArchiveInventoryAndExecutable |
| TWIN-INV-12/13/14 | digest, unsafe inventory, wrong executable, and missing-required-file rejection | verified_archive_test.go::TestStageVerifiedPostgresRejectsDigestAndInventoryFailures |
| TWIN-INV-13/14 | inventory ordering, duplicate/mode/required-file rejection | verified_archive_test.go::TestValidateVerifiedPostgresInputRejectsInventoryShape |
| TWIN-INV-13/14 | unsafe archive paths fail closed | verified_archive_test.go::TestNormalizeArchivePathRejectsUnsafeForms |
| TWIN-INV-13/14 | staged-file drift fails closed | verified_archive_test.go::TestVerifyStagedPostgresRejectsUnexpectedFileAndModeDrift |
| TWIN-INV-17 | cancellation is rejected before staging begins | verified_archive_test.go::TestStageVerifiedPostgresHonorsCancellationBeforeExtraction |
| P7 preflight | companion validates its verified input before touching the source store | verified_archive_test.go::TestReplayIsolatedWithVerifiedPostgresValidatesBeforeStore |
