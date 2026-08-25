//go:build integration

package projections

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		return tx.QueryRow(ctx, `SELECT subject FROM commits_view WHERE sha LIKE '%order'`).Scan(&subject)
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
	var checkpoint int64
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM commits_view`).Scan(&count); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT last_seq FROM projection_meta WHERE projection_name='default'`).Scan(&checkpoint)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 || checkpoint != 0 {
		t.Fatalf("partial state rows=%d checkpoint=%d", count, checkpoint)
	}
}

func TestRebuild_DoesNotModifyLedger(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "keep", "keep", 1)
	before := ledgerIdentity(t, s)
	if err := New().Rebuild(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if got := ledgerIdentity(t, s); !reflect.DeepEqual(got, before) {
		t.Fatalf("ledger changed: got=%v want=%v", got, before)
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
		return tx.QueryRow(ctx, `SELECT event_id,seq FROM commits_view WHERE sha LIKE '%proof'`).Scan(&id, &seq)
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

func TestApply_RejectsDeferredSessionAndReviewEvents(t *testing.T) {
	for _, tc := range []struct {
		source core.Source
		kind   core.Kind
	}{
		{core.SourceSessions, core.KindSessionObserved},
		{core.SourceReviews, core.KindReviewVerdict},
	} {
		t.Run(string(tc.source), func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			appendRaw(t, s, tc.source, tc.kind, "deferred", []byte(`{}`), 1)
			if err := New().Apply(context.Background(), s); err == nil {
				t.Fatal("deferred event accepted")
			}
		})
	}
}

func TestDecode_RejectsTrailingGarbageAndInvalidPayloads(t *testing.T) {
	valid := commitJSON("ignored", "subject")
	var payload commitPayload
	if err := decode(append(valid, []byte("garbage")...), &payload); err == nil {
		t.Fatal("trailing garbage accepted")
	}
	if err := decode([]byte(`{"sha":"not-a-hash"}`), &payload); err == nil {
		t.Fatal("incomplete commit accepted")
	}
}

func TestEnsure_MetadataIsUniqueAndVersioned(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	if err := New().Ensure(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var count, version int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*), max(projection_version) FROM projection_meta WHERE projection_name='default'`).Scan(&count, &version)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 || version != 1 {
		t.Fatalf("metadata=(%d,%d)", count, version)
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
	return appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, commitJSON(shaFor(sha), subject), n)
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
func shaFor(label string) string { return strings.Repeat("0", 40-len(label)) + label }
func ledgerCount(t *testing.T, s *store.Store) int {
	var n int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM events`).Scan(&n)
	}); err != nil {
		t.Fatal(err)
	}
	return n
}

func ledgerIdentity(t *testing.T, s *store.Store) []string {
	t.Helper()
	var identity []string
	err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT seq,event_id,source,native_id,kind,content_sha FROM events ORDER BY seq`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var seq int64
			var eventID, source, nativeID, kind, contentSHA string
			if err := rows.Scan(&seq, &eventID, &source, &nativeID, &kind, &contentSHA); err != nil {
				return err
			}
			identity = append(identity, fmt.Sprintf("%d|%s|%s|%s|%s|%s", seq, eventID, source, nativeID, kind, contentSHA))
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	return identity
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
