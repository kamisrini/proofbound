// Package twin provides bounded, non-mutating counterfactual event replay.
package twin

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net"
	"os"
	"sort"

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
	Proof           Proof
}

type Proof struct {
	Schema                  string
	ThroughSeq              int64
	SourceDigest            string
	CandidateDigest         string
	BaselineSnapshotDigest  string
	CandidateSnapshotDigest string
}

type SequenceBinding struct {
	EventID   core.EventID
	Requested int64
	Effective int64
}

type Project func(context.Context, []store.Record) (projections.Snapshot, error)

var (
	makeTempTwinRoot   = func() (string, error) { return os.MkdirTemp("", "vera-twin-") }
	removeTempTwinRoot = os.RemoveAll
)

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
	return makeResult(req, records, baseline, candidate, verdict), nil
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
	root, err := makeTempTwinRoot()
	if err != nil {
		return Result{}, fmt.Errorf("twin: create temporary root: %w", err)
	}
	var candidateStore *store.Store
	defer func() {
		if candidateStore != nil {
			err = errors.Join(err, candidateStore.Close())
		}
		err = errors.Join(err, removeTempTwinRoot(root))
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
	return makeResult(req, records, baseline, candidate, verdict), nil
}

func makeResult(req Request, records []store.Record, baseline, candidate projections.Snapshot, verdict Verdict) Result {
	return Result{
		ThroughSeq: req.ThroughSeq, SourceEvents: len(records), CandidateEvents: len(req.Candidate),
		Baseline: baseline, Candidate: candidate, Verdict: verdict,
		SequenceMap: sequenceMap(records, req.Candidate),
		Proof: Proof{
			Schema: "vera.replay.v1", ThroughSeq: req.ThroughSeq,
			SourceDigest: digestRecords(records), CandidateDigest: digestRecords(candidateRecords(req.Candidate)),
			BaselineSnapshotDigest: digestSnapshot(baseline), CandidateSnapshotDigest: digestSnapshot(candidate),
		},
	}
}

func candidateRecords(candidates []Candidate) []store.Record {
	result := make([]store.Record, 0, len(candidates))
	for _, c := range candidates {
		result = append(result, store.Record{Seq: c.Seq, Event: c.Event})
	}
	return result
}

func digestRecords(records []store.Record) string {
	h := sha256.New()
	var buf [8]byte
	for _, r := range records {
		binary.BigEndian.PutUint64(buf[:], uint64(r.Seq))
		_, _ = h.Write(buf[:])
		for _, part := range []string{r.Event.ID.String(), string(r.Event.Source), r.Event.NativeID, string(r.Event.Kind), r.Event.ContentSHA, string(r.Event.Payload)} {
			binary.BigEndian.PutUint64(buf[:], uint64(len(part)))
			_, _ = h.Write(buf[:])
			_, _ = h.Write([]byte(part))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func digestSnapshot(snapshot projections.Snapshot) string {
	h := sha256.New()
	var names []string
	for name := range snapshot.Tables {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf [8]byte
	for _, name := range names {
		writeDigestPart(h, buf[:], name)
		rows := append([]string(nil), snapshot.Tables[name]...)
		sort.Strings(rows)
		for _, row := range rows {
			writeDigestPart(h, buf[:], row)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writeDigestPart(h hash.Hash, buf []byte, value string) {
	binary.BigEndian.PutUint64(buf, uint64(len(value)))
	_, _ = h.Write(buf)
	_, _ = h.Write([]byte(value))
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
