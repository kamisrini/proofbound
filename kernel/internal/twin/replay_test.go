package twin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestReplayBoundsAndOrdersStream(t *testing.T) {
	s := testStore(t)
	appendEvent(t, s, 1)
	appendEvent(t, s, 2)
	before := eventSeqs(t, s)
	baseline := projections.Snapshot{Tables: map[string][]string{"x": {"base"}}}
	var got []int64
	result, err := Replay(context.Background(), s, Request{ThroughSeq: 1, Candidate: []Candidate{{Seq: 3, Event: event(3)}}}, baseline, func(_ context.Context, records []store.Record) (projections.Snapshot, error) {
		for _, r := range records {
			got = append(got, r.Seq)
		}
		return baseline, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != VerdictPreserved || len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("got result=%+v stream=%v", result, got)
	}
	if after := eventSeqs(t, s); !reflect.DeepEqual(before, after) {
		t.Fatalf("replay changed production ledger: before=%v after=%v", before, after)
	}
}

func TestReplayIsDeterministic(t *testing.T) {
	s := testStore(t)
	appendEvent(t, s, 1)
	b := projections.Snapshot{Tables: map[string][]string{"x": {"base"}}}
	project := func(_ context.Context, _ []store.Record) (projections.Snapshot, error) { return b, nil }
	a, err := Replay(context.Background(), s, Request{ThroughSeq: 1}, b, project)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Replay(context.Background(), s, Request{ThroughSeq: 1}, b, project)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("not deterministic: %+v vs %+v", a, c)
	}
	if a.Proof.Schema != "vera.replay.v1" || a.Proof.SourceDigest == "" || a.Proof.BaselineSnapshotDigest == "" {
		t.Fatalf("missing replay proof: %+v", a.Proof)
	}
}

func TestReplayDetectsChangedProjection(t *testing.T) {
	s := testStore(t)
	baseline := projections.Snapshot{Tables: map[string][]string{"x": {"base"}}}
	result, err := Replay(context.Background(), s, Request{ThroughSeq: 0}, baseline, func(_ context.Context, _ []store.Record) (projections.Snapshot, error) {
		return projections.Snapshot{Tables: map[string][]string{"x": {"candidate"}}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != VerdictChanged {
		t.Fatalf("got verdict %q", result.Verdict)
	}
}

func TestReplayRejectsInvalidCandidate(t *testing.T) {
	s := testStore(t)
	_, err := Replay(context.Background(), s, Request{ThroughSeq: 2, Candidate: []Candidate{{Seq: 2, Event: event(2)}}}, projections.Snapshot{}, func(context.Context, []store.Record) (projections.Snapshot, error) {
		return projections.Snapshot{}, nil
	})
	if err == nil {
		t.Fatal("expected sequence validation error")
	}
}

func TestReplayFailsClosedOnProjectorError(t *testing.T) {
	s := testStore(t)
	want := errors.New("boom")
	_, err := Replay(context.Background(), s, Request{ThroughSeq: 0}, projections.Snapshot{}, func(context.Context, []store.Record) (projections.Snapshot, error) {
		return projections.Snapshot{}, want
	})
	if err == nil || !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestReplayIsolatedProjectsCandidateAndPreservesSource(t *testing.T) {
	s := testStore(t)
	result, err := ReplayIsolated(context.Background(), s, Request{ThroughSeq: 0, Candidate: []Candidate{{Seq: 10, Event: githubDeploymentEvent(10)}}}, emptySnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != VerdictChanged || result.CandidateEvents != 1 || len(result.SequenceMap) != 1 || result.SequenceMap[0].Effective != 10 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if got := eventSeqs(t, s); len(got) != 0 {
		t.Fatalf("source ledger changed: %v", got)
	}
}

func TestReplayIsolatedSupportsMultipleGappedCandidates(t *testing.T) {
	s := testStore(t)
	result, err := ReplayIsolated(context.Background(), s, Request{ThroughSeq: 0, Candidate: []Candidate{
		{Seq: 10, Event: githubDeploymentEvent(10)}, {Seq: 20, Event: githubDeploymentEvent(20)},
	}}, emptySnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if result.CandidateEvents != 2 || len(result.SequenceMap) != 2 || result.SequenceMap[1].Effective != 20 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestReplayIsolatedRejectsCanceledContext(t *testing.T) {
	s := testStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReplayIsolated(ctx, s, Request{ThroughSeq: 0}, emptySnapshot()); err == nil {
		t.Fatal("expected canceled replay to fail closed")
	}
}

func TestReplayIsolatedFailsClosedOnProjectionError(t *testing.T) {
	s := testStore(t)
	bad := githubDeploymentEvent(30)
	bad.Payload = json.RawMessage(`{"repository":"github/docs"}`)
	h := sha256.Sum256(bad.Payload)
	bad.ContentSHA = hex.EncodeToString(h[:])
	if _, err := ReplayIsolated(context.Background(), s, Request{ThroughSeq: 0, Candidate: []Candidate{{Seq: 10, Event: bad}}}, emptySnapshot()); err == nil {
		t.Fatal("expected projection failure")
	}
	if got := eventSeqs(t, s); len(got) != 0 {
		t.Fatalf("source ledger changed: %v", got)
	}
}

func TestReplayRejectsDecreasingCandidates(t *testing.T) {
	s := testStore(t)
	_, err := Replay(context.Background(), s, Request{ThroughSeq: 1, Candidate: []Candidate{{Seq: 3, Event: event(3)}, {Seq: 3, Event: event(4)}}}, projections.Snapshot{}, func(context.Context, []store.Record) (projections.Snapshot, error) {
		return projections.Snapshot{}, nil
	})
	if err == nil {
		t.Fatal("expected candidate ordering error")
	}
}

func event(seq int64) core.Event {
	p := json.RawMessage(`{"seq":1}`)
	h := sha256.Sum256(p)
	return core.Event{ID: core.EventID{byte(seq)}, Source: core.SourceGit, NativeID: "n-" + strconv.FormatInt(seq, 10), Kind: core.KindCheckRun, OccurredAt: time.Unix(seq, 0), RecordedAt: time.Unix(seq, 0), Payload: p, ContentSHA: hex.EncodeToString(h[:]), ConnectorVersion: "test"}
}

func githubDeploymentEvent(id int) core.Event {
	p := json.RawMessage(fmt.Sprintf(`{"repository":"github/docs","deployment_id":%d,"environment":"preview","sha":"077df53b0a1887aa6e941814e384fba00bceae8b","url":"https://example.test/deploy/%d","created_at":"2026-08-26T00:00:00Z","updated_at":"2026-08-26T00:01:00Z"}`, id, id))
	h := sha256.Sum256(p)
	return core.Event{ID: core.EventID{byte(id)}, Source: core.SourceGitHub, NativeID: fmt.Sprintf("deployment-%d", id), Kind: core.KindGitHubDeployment, OccurredAt: time.Unix(1, 0), RecordedAt: time.Unix(1, 0), Payload: p, ContentSHA: hex.EncodeToString(h[:]), ConnectorVersion: "test"}
}

func emptySnapshot() projections.Snapshot {
	return projections.Snapshot{Tables: map[string][]string{"commits_view": {}, "checks_view": {}, "sessions_view": {}, "reviews_view": {}, "github_delivery_view": {}}}
}

func testStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(context.Background(), store.Config{Root: t.TempDir()})
	if err != nil {
		t.Skipf("embedded postgres unavailable: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
func appendEvent(t *testing.T, s *store.Store, seq int64) {
	t.Helper()
	sy, err := s.BeginSync(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = sy.Append(context.Background(), event(seq))
	if err != nil {
		t.Fatal(err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func eventSeqs(t *testing.T, s *store.Store) []int64 {
	t.Helper()
	var seqs []int64
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(r store.Record) error { seqs = append(seqs, r.Seq); return nil }); err != nil {
		t.Fatal(err)
	}
	return seqs
}
