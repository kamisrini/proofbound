// Package twin provides bounded, non-mutating counterfactual event replay.
package twin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type Candidate struct {
	Seq   int64
	Event core.Event
}

type Request struct {
	ThroughSeq int64
	Candidate  []Candidate
}

type Verdict string

const (
	VerdictPreserved Verdict = "preserved"
	VerdictChanged   Verdict = "changed"
)

type Result struct {
	ThroughSeq      int64
	SourceEvents    int
	CandidateEvents int
	Baseline        projections.Snapshot
	Candidate       projections.Snapshot
	Verdict         Verdict
	SequenceMap     []SequenceBinding
}

type SequenceBinding struct {
	EventID   core.EventID
	Requested int64
	Effective int64
}

type Project func(context.Context, []store.Record) (projections.Snapshot, error)

func Replay(ctx context.Context, s *store.Store, req Request, baseline projections.Snapshot, project Project) (Result, error) {
	if s == nil || project == nil || req.ThroughSeq < 0 {
		return Result{}, fmt.Errorf("twin: invalid replay request")
	}
	if err := validateCandidates(req); err != nil {
		return Result{}, err
	}
	records, err := readPrefix(ctx, s, req.ThroughSeq)
	if err != nil {
		return Result{}, fmt.Errorf("twin: read prefix: %w", err)
	}
	combined := append([]store.Record(nil), records...)
	for _, c := range req.Candidate {
		combined = append(combined, store.Record{Seq: c.Seq, Event: c.Event})
	}
	candidate, err := project(ctx, combined)
	if err != nil {
		return Result{}, fmt.Errorf("twin: project candidate: %w", err)
	}
	verdict := VerdictPreserved
	if err := projections.CompareSnapshots(baseline, candidate); err != nil {
		if !errors.Is(err, projections.ErrSnapshotMismatch) {
			return Result{}, fmt.Errorf("twin: compare snapshots: %w", err)
		}
		verdict = VerdictChanged
	}
	return Result{ThroughSeq: req.ThroughSeq, SourceEvents: len(records), CandidateEvents: len(req.Candidate), Baseline: baseline, Candidate: candidate, Verdict: verdict, SequenceMap: sequenceMap(records, req.Candidate)}, nil
}

// ReplayIsolated runs the real projector against a temporary embedded store.
// The source store is read-only from this path and the temporary store is
// removed on every return path.
func ReplayIsolated(ctx context.Context, s *store.Store, req Request, baseline projections.Snapshot) (result Result, err error) {
	if s == nil || req.ThroughSeq < 0 {
		return Result{}, fmt.Errorf("twin: invalid replay request")
	}
	if err := validateCandidates(req); err != nil {
		return Result{}, err
	}
	records, err := readPrefix(ctx, s, req.ThroughSeq)
	if err != nil {
		return Result{}, err
	}
	root, err := os.MkdirTemp("", "vera-twin-")
	if err != nil {
		return Result{}, fmt.Errorf("twin: create temporary root: %w", err)
	}
	var candidateStore *store.Store
	defer func() {
		if candidateStore != nil {
			err = errors.Join(err, candidateStore.Close())
		}
		err = errors.Join(err, os.RemoveAll(root))
	}()
	port, err := temporaryPort()
	if err != nil {
		return Result{}, err
	}
	candidateStore, err = store.Open(ctx, store.Config{Root: root, Port: port, AllowReplayImport: true})
	if err != nil {
		return Result{}, fmt.Errorf("twin: open temporary store: %w", err)
	}
	stream := make([]store.Record, 0, len(records)+len(req.Candidate))
	stream = append(stream, records...)
	for _, c := range req.Candidate {
		stream = append(stream, store.Record{Seq: c.Seq, Event: c.Event})
	}
	if err := candidateStore.ImportReplayRecords(ctx, stream); err != nil {
		return Result{}, fmt.Errorf("twin: import temporary stream: %w", err)
	}
	projector := projections.New()
	if err := projector.Apply(ctx, candidateStore); err != nil {
		return Result{}, fmt.Errorf("twin: project isolated stream: %w", err)
	}
	candidate, err := projector.Snapshot(ctx, candidateStore)
	if err != nil {
		return Result{}, fmt.Errorf("twin: snapshot isolated stream: %w", err)
	}
	verdict := VerdictPreserved
	if err := projections.CompareSnapshots(baseline, candidate); err != nil {
		if !errors.Is(err, projections.ErrSnapshotMismatch) {
			return Result{}, fmt.Errorf("twin: compare snapshots: %w", err)
		}
		verdict = VerdictChanged
	}
	return Result{ThroughSeq: req.ThroughSeq, SourceEvents: len(records), CandidateEvents: len(req.Candidate), Baseline: baseline, Candidate: candidate, Verdict: verdict, SequenceMap: sequenceMap(records, req.Candidate)}, nil
}

func readPrefix(ctx context.Context, s *store.Store, through int64) ([]store.Record, error) {
	records := make([]store.Record, 0)
	err := s.ReadEvents(ctx, store.Filter{}, func(r store.Record) error {
		if r.Seq > through {
			return store.ErrStopIteration
		}
		if err := r.Event.Validate(); err != nil {
			return fmt.Errorf("twin: source seq %d: %w", r.Seq, err)
		}
		records = append(records, r)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("twin: read prefix: %w", err)
	}
	return records, nil
}

func sequenceMap(records []store.Record, candidates []Candidate) []SequenceBinding {
	result := make([]SequenceBinding, 0, len(records)+len(candidates))
	for _, r := range records {
		result = append(result, SequenceBinding{EventID: r.Event.ID, Requested: r.Seq, Effective: r.Seq})
	}
	for _, c := range candidates {
		result = append(result, SequenceBinding{EventID: c.Event.ID, Requested: c.Seq, Effective: c.Seq})
	}
	return result
}

func temporaryPort() (uint16, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("twin: reserve temporary port: %w", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, fmt.Errorf("twin: release temporary port: %w", err)
	}
	return uint16(port), nil
}

func validateCandidates(req Request) error {
	last := req.ThroughSeq
	for i, c := range req.Candidate {
		if c.Seq <= last {
			return fmt.Errorf("twin: candidate %d sequence must follow through_seq", i)
		}
		if err := c.Event.Validate(); err != nil {
			return fmt.Errorf("twin: candidate %d: %w", i, err)
		}
		last = c.Seq
	}
	return nil
}
