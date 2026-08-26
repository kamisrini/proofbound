# SPEC — `internal/twin`

**Status:** authored before implementation · P4 twin spike · 2026-08-26
**Authority:** [ROADMAP.md](../../../ROADMAP.md) § P4 · [CLAUDE.md](../../../CLAUDE.md) Law 6

## Purpose

`twin` constructs a bounded, ephemeral candidate event stream. It never writes the
production ledger or projection tables. The first spike deliberately injects the
projection function; a PostgreSQL fork adapter is a later implementation behind this
contract.

## Interface

```go
type Candidate struct { Seq int64; Event core.Event }
type Request struct { ThroughSeq int64; Candidate []Candidate }
type Result struct {
    ThroughSeq int64
    SourceEvents int
    CandidateEvents int
    Baseline projections.Snapshot
    Candidate projections.Snapshot
    Verdict Verdict
}
type Verdict string
const VerdictPreserved Verdict = "preserved"
const VerdictChanged Verdict = "changed"
func Replay(context.Context, *store.Store, Request, projections.Snapshot,
    func(context.Context, []store.Record) (projections.Snapshot, error)) (Result, error)
```

## Invariants

1. **TWIN-INV-1 — Production preservation.** Replay only reads the source store.
2. **TWIN-INV-2 — Bounded source.** No source event after `ThroughSeq` reaches the projector.
3. **TWIN-INV-3 — Ordered replay.** Candidate sequences strictly follow the selected prefix.
4. **TWIN-INV-4 — Immutable source.** Source records are copied before candidate assembly.
5. **TWIN-INV-6 — Deterministic result.** Equal inputs and projector output yield equal verdicts.
6. **TWIN-INV-7 — Fail closed.** Invalid requests and projector errors return no successful result.

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
