// Package twin provides bounded, non-mutating counterfactual event replay.
package twin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	// launch carries the TWIN-INV-18 launched-process observation, populated only on the
	// ReplayIsolatedWithVerifiedPostgres path. Its resolvedSHA256 is produced solely by
	// witnessLaunchedProcess (unit-3 spec §5.1 N8) — never a literal.
	launch launchWitness
}

// launchWitness is the TWIN-INV-18 observation record. method is always "proc-exe".
type launchWitness struct {
	method         string
	resolvedPID    int
	resolvedSHA256 string
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
	makeTempTwinRoot   = func() (string, error) { return os.MkdirTemp("", "proofbound-twin-") }
	removeTempTwinRoot = os.RemoveAll
	closeTwinStore     = func(s *store.Store) error { return s.Close() }
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
	return replayRecords(ctx, req, records, baseline, project)
}

func replayRecords(ctx context.Context, req Request, records []store.Record, baseline projections.Snapshot, project Project) (Result, error) {
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
	return replayIsolated(ctx, s, req, baseline, nil, "")
}

// ReplayIsolatedWithVerifiedPostgres runs the real projector against a temporary
// embedded store provisioned from exact caller-supplied PostgreSQL archive bytes.
// It never accepts a path, URL, ambient cache, or network fallback.
//
// While the real postmaster is live it also computes the TWIN-INV-18 launched-process witness:
// the running executable image at /proc/<pid>/exe must equal input.PostgresBinarySHA256, or the
// replay fails with errLaunchedProcessDigestMismatch. The resolved digest is surfaced on
// Result.launch for the proof_binding evidence and is produced solely by witnessLaunchedProcess
// (unit-3 spec §5.1 N8).
func ReplayIsolatedWithVerifiedPostgres(ctx context.Context, s *store.Store, req Request, baseline projections.Snapshot, input VerifiedPostgresInput) (result Result, err error) {
	if err := validateVerifiedPostgresInput(input); err != nil {
		return Result{}, err
	}
	return replayIsolated(ctx, s, req, baseline, func(root string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("twin: stage verified postgres archive: %w", err)
		}
		return stageVerifiedPostgres(ctx, root, input)
	}, input.PostgresBinarySHA256)
}

func replayIsolated(ctx context.Context, s *store.Store, req Request, baseline projections.Snapshot, prepare func(string) (string, error), expectPostgresSHA256 string) (result Result, err error) {
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
			err = errors.Join(err, closeTwinStore(candidateStore))
		}
		err = errors.Join(err, removeTempTwinRoot(root))
	}()
	port, err := temporaryPort()
	if err != nil {
		return Result{}, err
	}
	var binariesDir string
	if prepare != nil {
		binariesDir, err = prepare(root)
		if err != nil {
			return Result{}, fmt.Errorf("twin: prepare verified postgres archive: %w", err)
		}
	}
	candidateStore, err = store.Open(ctx, store.Config{Root: root, Port: port, BinariesDir: binariesDir, AllowReplayImport: true})
	if err != nil {
		return Result{}, fmt.Errorf("twin: open temporary store: %w", err)
	}
	var witness launchWitness
	if expectPostgresSHA256 != "" {
		// TWIN-INV-18: with the real postmaster live, witness its launched executable image.
		pid, pidErr := findPostmasterPID(root)
		if pidErr != nil {
			return Result{}, fmt.Errorf("twin: locate launched postmaster: %w", pidErr)
		}
		resolved, wErr := witnessLaunchedProcess(pid, expectPostgresSHA256)
		if wErr != nil {
			return Result{}, fmt.Errorf("twin: launched process witness: %w", wErr)
		}
		witness = launchWitness{method: "proc-exe", resolvedPID: pid, resolvedSHA256: resolved}
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
	out := makeResult(req, records, baseline, candidate, verdict)
	out.launch = witness
	return out, nil
}

// findPostmasterPID locates the single postmaster.pid under the replay-owned temporary root and
// returns the postmaster PID from its first line. The temporary root is created and owned by this
// package (makeTempTwinRoot), so exactly one embedded cluster lives under it. This avoids coupling
// twin to the store package's data-directory defaulting convention (Law 2: that datum's home is
// package store).
func findPostmasterPID(root string) (int, error) {
	var pidFile string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "postmaster.pid" {
			pidFile = p
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walk temporary root: %w", err)
	}
	if pidFile == "" {
		return 0, errors.New("postmaster.pid not found under temporary root")
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", pidFile, err)
	}
	line := data
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		line = data[:i]
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(line)))
	if err != nil {
		return 0, fmt.Errorf("parse pid from %s: %w", pidFile, err)
	}
	if pid <= 0 {
		return 0, fmt.Errorf("non-positive pid %d in %s", pid, pidFile)
	}
	return pid, nil
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
