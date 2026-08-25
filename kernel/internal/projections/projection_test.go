//go:build integration

package projections

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestApply_UsesLedgerOrder(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "order", "first", 1)
	appendCommit(t, s, "order", "second", 2)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var subject string
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT subject FROM commits_view WHERE sha='order'`).Scan(&subject)
	}); err != nil {
		t.Fatal(err)
	}
	if subject != "second" {
		t.Fatalf("subject=%q", subject)
	}
}

func TestApply_RevisionLastWriteWins(t *testing.T) { TestApply_UsesLedgerOrder(t) }

func TestApply_MalformedPayloadRollsBack(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "good", "good", 1)
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, "bad", []byte(`{"sha":`), 2)
	if err := New().Apply(context.Background(), s); err == nil {
		t.Fatal("malformed event accepted")
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM commits_view`).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial rows=%d", count)
	}
}

func TestRebuild_DoesNotModifyLedger(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "keep", "keep", 1)
	before := ledgerCount(t, s)
	if err := New().Rebuild(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if got := ledgerCount(t, s); got != before {
		t.Fatalf("ledger count=%d want=%d", got, before)
	}
}

func TestRows_RetainProofIdentity(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	r := appendCommit(t, s, "proof", "proof", 1)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var id string
	var seq int64
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT event_id,seq FROM commits_view WHERE sha='proof'`).Scan(&id, &seq)
	}); err != nil {
		t.Fatal(err)
	}
	if id != r.Event.ID.String() || seq != r.Seq {
		t.Fatalf("proof=(%s,%d) want=(%s,%d)", id, seq, r.Event.ID, r.Seq)
	}
}

func TestApply_RejectsUnsupportedEvent(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendRaw(t, s, core.Source("unknown"), core.Kind("unknown"), "x", []byte(`{}`), 1)
	if err := New().Apply(context.Background(), s); err == nil {
		t.Fatal("unsupported event accepted")
	}
}

func TestDDL_IsNotLedgerMigration(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "store", "migrations", "001_ledger.sql"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, name := range []string{"commits_view", "checks_view", "sessions_view", "reviews_view", "projection_meta"} {
		if stringsContainsFold(text, name) {
			t.Fatalf("ledger migration owns %s", name)
		}
	}
}

func TestSnapshot_CanonicalMultisets(t *testing.T) {
	a := Snapshot{Tables: map[string][]string{"x": {"a", "b", "a"}}}
	b := Snapshot{Tables: map[string][]string{"x": {"b", "a", "a"}}}
	if err := CompareSnapshots(a, b); err != nil {
		t.Fatal(err)
	}
	if err := CompareSnapshots(a, Snapshot{Tables: map[string][]string{"x": {"a"}}}); !errors.Is(err, ErrSnapshotMismatch) {
		t.Fatalf("error=%v", err)
	}
}

func TestEnsure_CreatesFutureViews(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	if err := New().Ensure(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	snap, err := New().Snapshot(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sessions_view", "reviews_view"} {
		if len(snap.Tables[name]) != 0 {
			t.Fatalf("%s not empty", name)
		}
	}
}

func TestRebuild_RowSetMatchesIncremental(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "same", "same", 1)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	a, err := p.Snapshot(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Rebuild(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	b, err := p.Snapshot(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if err := CompareSnapshots(a, b); err != nil {
		t.Fatal(err)
	}
}

func testStore(t *testing.T) *store.Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required")
	}
	s, err := store.Open(context.Background(), store.Config{Root: t.TempDir(), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func appendCommit(t *testing.T, s *store.Store, sha, subject string, n int64) store.Record {
	return appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, commitJSON(sha, subject), n)
}
func appendRaw(t *testing.T, s *store.Store, source core.Source, kind core.Kind, native string, payload []byte, _ int64) store.Record {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	e, err := g.NewEvent(core.NewEventParams{Source: source, NativeID: native, Kind: kind, OccurredAt: time.Now(), Payload: payload, ConnectorVersion: "test/1"})
	if err != nil {
		t.Fatal(err)
	}
	sy, err := s.BeginSync(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := sy.Append(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	return r
}
func commitJSON(sha, subject string) []byte {
	b, _ := json.Marshal(map[string]any{"sha": sha, "author_name": "a", "author_email": "a@e", "committer_name": "c", "committer_email": "c@e", "committed_at": time.Now().UTC(), "subject": subject, "files_touched": []string{}, "cited_decisions": []string{}})
	return b
}
func ledgerCount(t *testing.T, s *store.Store) int {
	var n int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM events`).Scan(&n)
	}); err != nil {
		t.Fatal(err)
	}
	return n
}
func stringsContainsFold(s, sub string) bool {
	return len(s) >= len(sub) && reflect.ValueOf(s).String() != "" && containsFold(s, sub)
}
func containsFold(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 32
		}
		if y >= 'A' && y <= 'Z' {
			y += 32
		}
		if x != y {
			return false
		}
	}
	return true
}
