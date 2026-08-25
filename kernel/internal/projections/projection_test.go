//go:build integration

package projections

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	"github.com/oklog/ulid/v2"
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
		return tx.QueryRow(ctx, `SELECT subject FROM commits_view WHERE sha=$1`, shaFor("order")).Scan(&subject)
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
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, "bad", []byte(`{"sha":"bad"}`), 2)
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
		return tx.QueryRow(ctx, `SELECT event_id,seq FROM commits_view WHERE sha=$1`, shaFor("proof")).Scan(&id, &seq)
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
	appendRaw(t, s, core.Source("unknown"), core.KindCommitRecorded, "x", commitJSON(shaFor("unsupported"), "unsupported"), 1)
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

func TestApply_CheckRun(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	runID := ulid.MustParse("01ARZ3NDEKTSV4RRFFQ69G5FAV").String()
	payload, err := json.Marshal(map[string]any{
		"schema": "vera.witness.v1", "run_id": runID, "command": "make check", "exit_code": 0,
		"started_at": "2026-08-25T12:00:00Z", "finished_at": "2026-08-25T12:00:01Z", "duration_ms": 1000,
		"output_sha256": strings.Repeat("a", 64), "git_sha": strings.Repeat("b", 40), "git_dirty": false,
		"tool_versions": map[string]string{"go": "go1.26", "golangci_lint": "v2", "make": "GNU Make 4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	appendRaw(t, s, core.SourceChecks, core.KindCheckRun, "check", payload, 1)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM checks_view WHERE run_id=$1`, runID).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("checks_view rows=%d", count)
	}
}

func TestSupportedEventMatrix(t *testing.T) {
	cases := []struct {
		source core.Source
		kind   core.Kind
		want   bool
	}{
		{core.SourceGit, core.KindCommitRecorded, true},
		{core.SourceChecks, core.KindCheckRun, true},
		{core.SourceGit, core.KindCheckRun, false},
		{core.SourceChecks, core.KindCommitRecorded, false},
		{core.SourceSessions, core.KindSessionObserved, false},
		{core.SourceReviews, core.KindReviewVerdict, false},
	}
	for _, tc := range cases {
		if got := supported(tc.source, tc.kind); got != tc.want {
			t.Errorf("supported(%s,%s)=%v want %v", tc.source, tc.kind, got, tc.want)
		}
		wantRoute := routeUnsupported
		if tc.source == core.SourceGit && tc.kind == core.KindCommitRecorded {
			wantRoute = routeCommit
		}
		if tc.source == core.SourceChecks && tc.kind == core.KindCheckRun {
			wantRoute = routeCheck
		}
		if got := eventRoute(tc.source, tc.kind); got != wantRoute {
			t.Errorf("eventRoute(%s,%s)=%v want %v", tc.source, tc.kind, got, wantRoute)
		}
	}
}

func TestCommitValidationRejectsUnsafeArrays(t *testing.T) {
	for _, files := range [][]string{{""}, {"unsafe\x00path"}} {
		payload := commitPayload{SHA: shaFor("files"), AuthorName: "Author", AuthorEmail: "author@example.test", CommitterName: "Committer", CommitterEmail: "committer@example.test", CommittedAt: time.Now(), Subject: "subject", FilesTouched: files}
		if err := payload.validate(); err == nil {
			t.Fatalf("files %q accepted", files)
		}
	}
	valid := commitPayload{SHA: shaFor("citation"), AuthorName: "Author", AuthorEmail: "author@example.test", CommitterName: "Committer", CommitterEmail: "committer@example.test", CommittedAt: time.Now(), Subject: "subject", CitedDecisions: []string{"not-a-decision"}}
	if err := valid.validate(); err == nil {
		t.Fatal("invalid citation accepted")
	}
}

func TestCommitValidationRejectsEachScalar(t *testing.T) {
	base := commitPayload{SHA: shaFor("scalar"), AuthorName: "Author", AuthorEmail: "author@example.test", CommitterName: "Committer", CommitterEmail: "committer@example.test", CommittedAt: time.Now(), Subject: "subject"}
	cases := []struct {
		name string
		edit func(*commitPayload)
	}{
		{"sha", func(v *commitPayload) { v.SHA = "bad" }},
		{"author name", func(v *commitPayload) { v.AuthorName = " " }},
		{"author email", func(v *commitPayload) { v.AuthorEmail = "bad" }},
		{"committer name", func(v *commitPayload) { v.CommitterName = " " }},
		{"committer email", func(v *commitPayload) { v.CommitterEmail = "bad" }},
		{"committed at", func(v *commitPayload) { v.CommittedAt = time.Time{} }},
		{"subject", func(v *commitPayload) { v.Subject = " " }},
	}
	for _, tc := range cases {
		v := base
		tc.edit(&v)
		if err := v.validate(); err == nil {
			t.Errorf("%s accepted", tc.name)
		}
	}
}

func TestCheckValidationRejectsEachScalar(t *testing.T) {
	base := checkPayload{Schema: "vera.witness.v1", RunID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Command: "make check", StartedAt: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC), FinishedAt: time.Date(2026, 8, 25, 12, 0, 1, 0, time.UTC), DurationMS: 1000, OutputSHA256: strings.Repeat("a", 64), GitSHA: strings.Repeat("b", 40), ToolVersions: toolPayload{Go: "go", GolangCILint: "lint", Make: "make"}}
	cases := []struct {
		name string
		edit func(*checkPayload)
	}{
		{"schema", func(v *checkPayload) { v.Schema = "bad" }}, {"run id", func(v *checkPayload) { v.RunID = "bad" }}, {"command", func(v *checkPayload) { v.Command = "bad" }},
		{"exit code", func(v *checkPayload) { v.ExitCode = 256 }}, {"started", func(v *checkPayload) { v.StartedAt = time.Time{} }}, {"finished", func(v *checkPayload) { v.FinishedAt = time.Time{} }},
		{"order", func(v *checkPayload) { v.FinishedAt = v.StartedAt.Add(-time.Second) }}, {"duration", func(v *checkPayload) { v.DurationMS = -1 }},
		{"output hash", func(v *checkPayload) { v.OutputSHA256 = "bad" }}, {"git hash", func(v *checkPayload) { v.GitSHA = "bad" }}, {"go", func(v *checkPayload) { v.ToolVersions.Go = " " }},
		{"lint", func(v *checkPayload) { v.ToolVersions.GolangCILint = " " }}, {"make", func(v *checkPayload) { v.ToolVersions.Make = " " }},
	}
	for _, tc := range cases {
		v := base
		tc.edit(&v)
		if err := v.validate(); err == nil {
			t.Errorf("%s accepted", tc.name)
		}
	}
}

func TestValidateSequence(t *testing.T) {
	if err := validateSequence(3, nil, 4); err != nil {
		t.Fatal(err)
	}
	if err := validateSequence(3, nil, 3); err == nil {
		t.Fatal("checkpoint duplicate accepted")
	}
	records := []store.Record{{Seq: 5}}
	if err := validateSequence(3, records, 5); err == nil {
		t.Fatal("duplicate accepted")
	}
	if err := validateSequence(3, records, 4); err == nil {
		t.Fatal("descending sequence accepted")
	}
}

func TestJSONColumnMatrix(t *testing.T) {
	for _, tc := range []struct {
		table string
		index int
		want  bool
	}{
		{"commits_view", 9, true}, {"commits_view", 10, true}, {"commits_view", 8, false},
		{"checks_view", 11, true}, {"checks_view", 10, false}, {"sessions_view", 8, false},
	} {
		if got := isJSONColumn(tc.table, tc.index); got != tc.want {
			t.Errorf("isJSONColumn(%q,%d)=%v want %v", tc.table, tc.index, got, tc.want)
		}
	}
}

func TestRequireFieldsDistinguishesMissingAndNullableArrays(t *testing.T) {
	base := map[string]any{}
	if err := json.Unmarshal(commitJSON(shaFor("required"), "subject"), &base); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"author_name", "files_touched"} {
		copy := map[string]any{}
		for key, value := range base {
			copy[key] = value
		}
		delete(copy, name)
		raw, _ := json.Marshal(copy)
		var payload commitPayload
		if err := decode(raw, &payload); err == nil {
			t.Fatalf("missing %s accepted", name)
		}
	}
	for _, name := range []string{"author_name", "files_touched"} {
		copy := map[string]any{}
		for key, value := range base {
			copy[key] = value
		}
		copy[name] = nil
		raw, _ := json.Marshal(copy)
		var payload commitPayload
		err := decode(raw, &payload)
		if name == "files_touched" {
			if err != nil {
				t.Fatalf("nullable %s rejected: %v", name, err)
			}
		} else if err == nil {
			t.Fatalf("nullable %s accepted", name)
		}
	}
}

func TestSnapshotPropagatesEnsureError(t *testing.T) {
	if _, err := New().Snapshot(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "snapshot ensure") {
		t.Fatal("nil store accepted")
	}
}

func TestSnapshotRowDigestPropagatesEncodingErrors(t *testing.T) {
	if _, err := snapshotRowDigest(map[string]any{"bad": func() {}}); err == nil {
		t.Fatal("marshal error suppressed")
	}
	if _, err := snapshotRowDigest(map[string]any{"bad": json.RawMessage(`{"`)}); err == nil {
		t.Fatal("canonicalization error suppressed")
	}
	if digest, err := snapshotRowDigest(map[string]any{"ok": "value"}); err != nil || digest == "" {
		t.Fatalf("valid row digest=%q err=%v", digest, err)
	}
}

type fakeSnapshotRows struct {
	next    bool
	scanErr error
	rowsErr error
	badJSON bool
}

func (r *fakeSnapshotRows) Next() bool { return r.next }
func (r *fakeSnapshotRows) Scan(dest ...any) error {
	if r.badJSON {
		*(dest[9].(*any)) = json.RawMessage(`{"`)
	}
	r.next = false
	return r.scanErr
}
func (r *fakeSnapshotRows) Err() error { return r.rowsErr }

func TestReadSnapshotRowsPropagatesDriverErrors(t *testing.T) {
	scanErr := errors.New("scan failed")
	if _, err := readSnapshotRows("commits_view", &fakeSnapshotRows{next: true, scanErr: scanErr}); !errors.Is(err, scanErr) {
		t.Fatalf("scan error=%v", err)
	}
	rowsErr := errors.New("rows failed")
	if _, err := readSnapshotRows("commits_view", &fakeSnapshotRows{rowsErr: rowsErr}); !errors.Is(err, rowsErr) {
		t.Fatalf("rows error=%v", err)
	}
	if _, err := readSnapshotRows("commits_view", &fakeSnapshotRows{next: true, badJSON: true}); err == nil {
		t.Fatal("canonical JSON error suppressed")
	}
}

func TestAppendSnapshotDigestsPropagatesDigestErrors(t *testing.T) {
	if err := appendSnapshotDigests(map[string][]string{}, "commits_view", []map[string]any{{"bad": func() {}}}); err == nil {
		t.Fatal("digest error suppressed")
	}
}

func TestSnapshotPropagatesDigestFunctionError(t *testing.T) {
	old := snapshotDigestRows
	snapshotDigestRows = func(map[string][]string, string, []map[string]any) error {
		return errors.New("digest rows failed")
	}
	defer func() { snapshotDigestRows = old }()
	s := testStore(t)
	defer s.Close()
	if _, err := New().Snapshot(context.Background(), s); err == nil || !strings.Contains(err.Error(), "digest rows failed") {
		t.Fatalf("digest function error=%v", err)
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

func TestSnapshotPropagatesQueryError(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	if err := New().Ensure(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `ALTER TABLE commits_view RENAME COLUMN sha TO sha_broken`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
			_, err := tx.Exec(ctx, `ALTER TABLE commits_view RENAME COLUMN sha_broken TO sha`)
			return err
		})
	}()
	if _, err := New().Snapshot(context.Background(), s); err == nil {
		t.Fatal("query error suppressed")
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
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	if _, err := pool.Exec(context.Background(), `TRUNCATE events, sync_runs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `DROP TABLE IF EXISTS projection_meta, commits_view, checks_view, sessions_view, reviews_view CASCADE`); err != nil {
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
	b, _ := json.Marshal(map[string]any{"sha": sha, "author_name": "A Author", "author_email": "author@example.test", "committer_name": "C Committer", "committer_email": "committer@example.test", "committed_at": time.Now().UTC(), "subject": subject, "files_touched": []string{}, "cited_decisions": []string{}})
	return b
}
func shaFor(label string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(label))) }
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
