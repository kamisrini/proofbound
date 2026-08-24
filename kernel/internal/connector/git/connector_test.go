package git

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type fakeRepo struct {
	commits []Commit
	tips    map[string]string
	err     error
}

func (r *fakeRepo) Commits(context.Context) ([]Commit, error) {
	return append([]Commit(nil), r.commits...), r.err
}
func (r *fakeRepo) Tips(context.Context) (map[string]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.tips, nil
}

type memoryAppender struct {
	events []core.Event
	seen   map[string]struct{}
	failAt int
	err    error
}

func (a *memoryAppender) Append(_ context.Context, event core.Event) (store.Record, bool, error) {
	if a.failAt > 0 && len(a.events)+1 == a.failAt {
		return store.Record{}, false, a.err
	}
	if a.seen == nil {
		a.seen = make(map[string]struct{})
	}
	key := event.IdempotencyKey().String()
	if _, ok := a.seen[key]; ok {
		return store.Record{Event: event}, false, nil
	}
	a.seen[key] = struct{}{}
	a.events = append(a.events, event)
	return store.Record{Seq: int64(len(a.events)), Event: event}, true, nil
}

func TestSync_MintsOneCommitRecordedEventPerCommit(t *testing.T) {
	connector := testConnector(t, &fakeRepo{commits: []Commit{testCommit("abc")}, tips: map[string]string{"HEAD": "abc"}})
	appender := &memoryAppender{}
	result, err := connector.Sync(context.Background(), appender)
	if err != nil {
		t.Fatal(err)
	}
	if result.Listed != 1 || result.Appended != 1 || result.Existing != 0 || string(result.Cursor) != `{"HEAD":"abc"}` {
		t.Fatalf("result=%+v", result)
	}
	event := appender.events[0]
	if event.NativeID != "abc" || event.Source != core.SourceGit || event.Kind != core.KindCommitRecorded || event.ConnectorVersion != Version {
		t.Fatalf("event=%+v", event)
	}
}

func TestSync_SecondSyncAppendsNothing(t *testing.T) {
	connector := testConnector(t, &fakeRepo{commits: []Commit{testCommit("abc")}, tips: map[string]string{"HEAD": "abc"}})
	appender := &memoryAppender{}
	if _, err := connector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	result, err := connector.Sync(context.Background(), appender)
	if err != nil || result.Appended != 0 || result.Existing != 1 || len(appender.events) != 1 {
		t.Fatalf("result=%+v events=%d error=%v", result, len(appender.events), err)
	}
}

func TestSync_ResultCountsPartitionTheListingExactly(t *testing.T) {
	repo := &fakeRepo{commits: []Commit{testCommit("a"), testCommit("b"), testCommit("c")}, tips: map[string]string{"HEAD": "c"}}
	connector := testConnector(t, repo)
	appender := &memoryAppender{seen: map[string]struct{}{}}
	firstConnector := testConnector(t, &fakeRepo{commits: []Commit{testCommit("a")}, tips: repo.tips})
	if _, err := firstConnector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	result, err := connector.Sync(context.Background(), appender)
	if err != nil || result.Listed != 3 || result.Appended != 2 || result.Existing != 1 || result.Appended+result.Existing != result.Listed {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestPayload_IsStableAcrossReads(t *testing.T) {
	first := testCommit("abc")
	first.FilesTouched = []string{"b", "a", "b"}
	first.CitedDecisions = []string{"VD-z-aaaaaa", "VD-a-bbbbbb", "VD-z-aaaaaa"}
	second := first
	second.FilesTouched = []string{"a", "b"}
	second.CitedDecisions = []string{"VD-a-bbbbbb", "VD-z-aaaaaa"}
	a := syncOne(t, first)
	b := syncOne(t, second)
	if a.ContentSHA != b.ContentSHA || !bytes.Equal(a.Payload, b.Payload) {
		t.Fatalf("a=%s/%s b=%s/%s", a.ContentSHA, a.Payload, b.ContentSHA, b.Payload)
	}
}

func TestPayload_FieldChangeChangesTheContentSHA(t *testing.T) {
	first := testCommit("abc")
	second := first
	second.Subject = "changed"
	if syncOne(t, first).ContentSHA == syncOne(t, second).ContentSHA {
		t.Fatal("changed payload retained content sha")
	}
}

func TestPayload_EmptySlicesArePinnedToNull(t *testing.T) {
	commit := testCommit("abc")
	commit.FilesTouched = []string{}
	commit.CitedDecisions = []string{}
	event := syncOne(t, commit)
	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["files_touched"] != nil || payload["cited_decisions"] != nil {
		t.Fatalf("payload=%s", event.Payload)
	}
}

func TestPayload_PinnedVector(t *testing.T) {
	commit := Commit{
		SHA:            "9f2b7c1d8e5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c",
		AuthorName:     "A Author",
		AuthorEmail:    "author@example.test",
		CommitterName:  "C Committer",
		CommitterEmail: "committer@example.test",
		CommittedAt:    time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC),
		Subject:        "pin the vector",
		FilesTouched:   []string{"b.go", "a.go", "b.go"},
		CitedDecisions: []string{"VD-fixture-aaaaaa", "VD-fixture-aaaaaa"},
	}
	event := syncOne(t, commit)
	wantPayload := `{"author_email":"author@example.test","author_name":"A Author","cited_decisions":["VD-fixture-aaaaaa"],"committed_at":"2026-08-12T09:00:00Z","committer_email":"committer@example.test","committer_name":"C Committer","files_touched":["a.go","b.go"],"sha":"9f2b7c1d8e5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c","subject":"pin the vector"}`
	if string(event.Payload) != wantPayload || event.ContentSHA != "eaaf4869ccfa6491f2da707db289c4f3a72aa851a9d2c7db8dcb07e9af6be7ad" {
		t.Fatalf("payload=%s\nsha=%s", event.Payload, event.ContentSHA)
	}
}

func TestSync_ListingOrderDoesNotChangeTheOutcome(t *testing.T) {
	a := []Commit{testCommit("a"), testCommit("b"), testCommit("c")}
	b := []Commit{a[2], a[0], a[1]}
	one := syncAll(t, a)
	two := syncAll(t, b)
	if !reflect.DeepEqual(eventSHAs(one), eventSHAs(two)) {
		t.Fatalf("one=%v two=%v", eventSHAs(one), eventSHAs(two))
	}
}

func TestSync_AppendFailureAbortsAndReportsProgress(t *testing.T) {
	sentinel := errors.New("append failed")
	connector := testConnector(t, &fakeRepo{commits: []Commit{testCommit("a"), testCommit("b"), testCommit("c")}, tips: map[string]string{}})
	appender := &memoryAppender{failAt: 2, err: sentinel}
	result, err := connector.Sync(context.Background(), appender)
	if !errors.Is(err, sentinel) || result.Listed != 3 || result.Appended != 1 || len(appender.events) != 1 {
		t.Fatalf("result=%+v events=%d error=%v", result, len(appender.events), err)
	}
}

func TestSync_MalformedCommitIsRefusedByCore(t *testing.T) {
	commit := testCommit("bad\x00sha")
	connector := testConnector(t, &fakeRepo{commits: []Commit{commit}, tips: map[string]string{}})
	result, err := connector.Sync(context.Background(), &memoryAppender{})
	if !errors.Is(err, core.ErrInvalidEvent) || result.Appended != 0 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestSync_CursorIsTheRepositoryTips(t *testing.T) {
	tips := map[string]string{"refs/heads/main": "a", "HEAD": "a"}
	connector := testConnector(t, &fakeRepo{tips: tips})
	result, err := connector.Sync(context.Background(), &memoryAppender{})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(result.Cursor, &got); err != nil || !reflect.DeepEqual(got, tips) {
		t.Fatalf("cursor=%s error=%v", result.Cursor, err)
	}
}

func TestVersion_IsPinned(t *testing.T) {
	if Version != "git/1" {
		t.Fatalf("version=%q", Version)
	}
}

func TestNew_RequiresEveryDependency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	valid := &Deps{Repo: &fakeRepo{}, IDs: ids, Logger: logger}
	var typedNil *fakeRepo
	for name, deps := range map[string]*Deps{
		"deps":      nil,
		"repo":      {IDs: ids, Logger: logger},
		"typed-nil": {Repo: typedNil, IDs: ids, Logger: logger},
		"ids":       {Repo: valid.Repo, Logger: logger},
		"log":       {Repo: valid.Repo, IDs: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if connector, newErr := New(deps); newErr == nil || connector != nil {
				t.Fatalf("connector=%v error=%v", connector, newErr)
			}
		})
	}
}

func TestSync_RejectsEveryUninitializedConnectorShape(t *testing.T) {
	appender := &memoryAppender{}
	var nilConnector *Connector
	for name, connector := range map[string]*Connector{
		"nil":       nilConnector,
		"empty":     {},
		"repo-only": {repo: &fakeRepo{}},
	} {
		t.Run(name, func(t *testing.T) {
			if result, err := connector.Sync(context.Background(), appender); err == nil || !reflect.DeepEqual(result, Result{}) {
				t.Fatalf("result=%+v error=%v", result, err)
			}
		})
	}
}

func testConnector(t *testing.T, repo Repo) *Connector {
	t.Helper()
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: func() time.Time { return time.Unix(100, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	connector, err := New(&Deps{Repo: repo, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	return connector
}

func testCommit(sha string) Commit {
	return Commit{SHA: sha, AuthorName: "Author", AuthorEmail: "author@example.test", CommitterName: "Committer", CommitterEmail: "committer@example.test", CommittedAt: time.Unix(99, 0), Subject: "subject"}
}

func syncOne(t *testing.T, commit Commit) core.Event {
	t.Helper()
	events := syncAll(t, []Commit{commit})
	if len(events) != 1 {
		t.Fatalf("events=%d", len(events))
	}
	return events[0]
}

func syncAll(t *testing.T, commits []Commit) []core.Event {
	t.Helper()
	connector := testConnector(t, &fakeRepo{commits: commits, tips: map[string]string{}})
	appender := &memoryAppender{}
	if _, err := connector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	sorted := append([]core.Event(nil), appender.events...)
	sortEvents(sorted)
	return sorted
}

func sortEvents(events []core.Event) {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].NativeID < events[i].NativeID {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

func eventSHAs(events []core.Event) []string {
	result := make([]string, len(events))
	for i, event := range events {
		result[i] = event.NativeID
	}
	return result
}
