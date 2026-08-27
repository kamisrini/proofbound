# SPEC — `internal/twin`

**Status:** implemented slice · P4 twin spike · 2026-08-27
**Authority:** [ROADMAP.md](../../../ROADMAP.md) § P4 · [CLAUDE.md](../../../CLAUDE.md) Law 6

## Purpose

`twin` constructs a bounded, ephemeral candidate event stream. It never writes the
production ledger or projection tables. `Replay` injects a projector for fast tests;
`ReplayIsolated` runs the real projector in a temporary embedded store.

## Interface

```go
type Candidate struct { Seq int64; Event core.Event }
type Request struct { ThroughSeq int64; Candidate []Candidate }
type SequenceBinding struct { EventID core.EventID; Requested, Effective int64 }
type Proof struct { Schema string; ThroughSeq int64; SourceDigest, CandidateDigest string; BaselineSnapshotDigest, CandidateSnapshotDigest string }
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
func Calibrate([]Forecast) (Calibration, error)
```

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
