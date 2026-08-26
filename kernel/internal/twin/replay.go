// Package twin provides bounded, non-mutating counterfactual event replay.
package twin

import (
	"context"
	"errors"
	"fmt"

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
}

type Project func(context.Context, []store.Record) (projections.Snapshot, error)

func Replay(ctx context.Context, s *store.Store, req Request, baseline projections.Snapshot, project Project) (Result, error) {
	if s == nil || project == nil || req.ThroughSeq < 0 {
		return Result{}, fmt.Errorf("twin: invalid replay request")
	}
	if err := validateCandidates(req); err != nil {
		return Result{}, err
	}
	records := make([]store.Record, 0)
	err := s.ReadEvents(ctx, store.Filter{}, func(r store.Record) error {
		if r.Seq > req.ThroughSeq {
			return store.ErrStopIteration
		}
		if err := r.Event.Validate(); err != nil {
			return fmt.Errorf("twin: source seq %d: %w", r.Seq, err)
		}
		records = append(records, r)
		return nil
	})
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
	return Result{ThroughSeq: req.ThroughSeq, SourceEvents: len(records), CandidateEvents: len(req.Candidate), Baseline: baseline, Candidate: candidate, Verdict: verdict}, nil
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
