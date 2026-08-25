package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type fakeAppender struct{ events []core.Event }

func (a *fakeAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	for _, old := range a.events {
		if old.IdempotencyKey() == e.IdempotencyKey() {
			return store.Record{Event: old}, false, nil
		}
	}
	a.events = append(a.events, e)
	return store.Record{Event: e}, true, nil
}

func testConnector(t *testing.T, root, home string) *Connector {
	t.Helper()
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{1}, 1024)), Now: func() time.Time { return time.Unix(100, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(&Deps{Root: root, HomeDir: home, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func writeFixture(t *testing.T, home, root, name, data string, age time.Duration) {
	t.Helper()
	dir := ProjectDir(home, root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func TestProjectDir(t *testing.T) {
	got := ProjectDir("/home/test", "/tmp/a.b/repo")
	want := filepath.Join("/home/test", ".claude", "projects", "-tmp-a-b-repo")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSyncRecordsParseCoverage(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	data := `{"sessionId":"s-1","timestamp":"2026-08-01T00:00:00Z","type":"user","message":{"content":"hello"}}
not json
{"sessionId":"s-1","timestamp":"2026-08-01T00:01:00Z","type":"tool_use","name":"Write"}
`
	writeFixture(t, home, root, "s-1.jsonl", data, 11*time.Minute)
	a := &fakeAppender{}
	r, err := testConnector(t, root, home).Sync(context.Background(), a)
	if err != nil || r.Appended != 1 {
		t.Fatalf("result=%+v err=%v", r, err)
	}
	var got payload
	if err := json.Unmarshal(a.events[0].Payload, &got); err != nil {
		t.Fatal(err)
	}
	if got.ParseCoverage != 2.0/3.0 || got.MessageCount != 2 || got.ToolCallCount != 1 || got.FilesWrittenCount != 1 {
		t.Fatalf("payload=%+v", got)
	}
	if strings.Contains(string(a.events[0].Payload), "hello") {
		t.Fatal("message content leaked")
	}
}

func TestSyncDropsLowCoverage(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	writeFixture(t, home, root, "bad.jsonl", "not json\n{\"sessionId\":\"x\"}\nnot json\n", 11*time.Minute)
	a := &fakeAppender{}
	r, err := testConnector(t, root, home).Sync(context.Background(), a)
	if err != nil || r.Appended != 0 || r.Skipped != 1 {
		t.Fatalf("result=%+v err=%v", r, err)
	}
}

func TestSyncSkipsLiveFiles(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	writeFixture(t, home, root, "live.jsonl", `{"sessionId":"live"}
`, time.Minute)
	a := &fakeAppender{}
	r, err := testConnector(t, root, home).Sync(context.Background(), a)
	if err != nil || r.Appended != 0 || r.Skipped != 1 {
		t.Fatalf("result=%+v err=%v", r, err)
	}
}

func TestSyncDeduplicates(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	writeFixture(t, home, root, "s.jsonl", `{"sessionId":"s"}
`, 11*time.Minute)
	a := &fakeAppender{}
	c := testConnector(t, root, home)
	first, err := c.Sync(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.Sync(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	if first.Appended != 1 || second.Existing != 1 || len(a.events) != 1 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
}

func TestNewValidation(t *testing.T) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{2}, 1024)), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	valid := &Deps{Root: t.TempDir(), HomeDir: t.TempDir(), IDs: ids, Logger: log}
	for name, deps := range map[string]*Deps{
		"nil":        nil,
		"empty root": {HomeDir: valid.HomeDir, IDs: ids, Logger: log},
		"nil ids":    {Root: valid.Root, HomeDir: valid.HomeDir, Logger: log},
		"nil logger": {Root: valid.Root, HomeDir: valid.HomeDir, IDs: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(deps); err == nil {
				t.Fatal("invalid dependencies accepted")
			}
		})
	}
	if c, err := New(valid); err != nil || c.now == nil {
		t.Fatalf("valid dependencies: connector=%v err=%v", c, err)
	}
	oldHome := userHomeDir
	userHomeDir = func() (string, error) { return "", os.ErrNotExist }
	defer func() { userHomeDir = oldHome }()
	if _, err := New(&Deps{Root: valid.Root, IDs: ids, Logger: log}); err == nil {
		t.Fatal("home directory error suppressed")
	}
}

func TestSyncValidationAndFileFiltering(t *testing.T) {
	var nilConnector *Connector
	if _, err := nilConnector.Sync(context.Background(), &fakeAppender{}); err == nil {
		t.Fatal("nil connector accepted")
	}
	c := testConnector(t, t.TempDir(), t.TempDir())
	for name, broken := range map[string]*Connector{
		"nil ids":    {root: c.root, home: c.home},
		"empty root": {ids: c.ids, home: c.home},
		"empty home": {ids: c.ids, root: c.root},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := broken.Sync(context.Background(), &fakeAppender{}); err == nil {
				t.Fatal("invalid connector accepted")
			}
		})
	}
	var nilAppender *fakeAppender
	if _, err := c.Sync(context.Background(), nilAppender); err == nil {
		t.Fatal("typed nil appender accepted")
	}
	dir := ProjectDir(c.home, c.root)
	if err := os.MkdirAll(filepath.Join(dir, "nested.jsonl"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, c.home, c.root, "ignored.txt", "{}\n", 11*time.Minute)
	writeFixture(t, c.home, c.root, "good.jsonl", `{"sessionId":"good"}
`, 11*time.Minute)
	r, err := c.Sync(context.Background(), &fakeAppender{})
	if err != nil || r.Listed != 1 || r.Appended != 1 {
		t.Fatalf("result=%+v err=%v", r, err)
	}
}

func TestParseFileRejectsEmptyAndMissingID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseFile(path, "fallback"); err == nil {
		t.Fatal("empty file accepted")
	}
	if err := os.WriteFile(path, []byte(`{"message":"metadata"}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseFile(path, ""); err == nil {
		t.Fatal("missing session id accepted")
	}
	if err := os.WriteFile(path, []byte("not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseFile(path, "fallback"); err == nil {
		t.Fatal("zero valid lines accepted")
	}
}

func TestObserveMetadataBranches(t *testing.T) {
	var p payload
	observe(&p, map[string]any{"session_id": "explicit"})
	if p.SessionID != "explicit" {
		t.Fatalf("session_id=%q", p.SessionID)
	}
	observe(&p, map[string]any{"session_id": "first", "timestamp": "2026-08-01T00:02:00Z", "type": "tool_use", "name": "Write"})
	observe(&p, map[string]any{"sessionId": "second", "timestamp": "2026-08-01T00:01:00Z", "tool_calls": []any{}})
	observe(&p, map[string]any{"timestamp": "2026-08-01T00:03:00Z", "message": map[string]any{"content": []any{map[string]any{"type": "tool_call"}}}, "tool_name": "Edit"})
	if p.SessionID != "second" || p.MessageCount != 4 || p.ToolCallCount != 3 || p.FilesWrittenCount != 2 || p.StartedAt == nil || p.FinishedAt == nil {
		t.Fatalf("payload=%+v", p)
	}
	if timestamp(map[string]any{"timestamp": "not-time"}) != nil {
		t.Fatal("invalid timestamp accepted")
	}
	if isToolCall(map[string]any{"message": map[string]any{"content": []any{map[string]any{"type": "text"}}}}) {
		t.Fatal("text classified as tool call")
	}
	if writesFile(map[string]any{"name": "Read"}) {
		t.Fatal("read classified as write")
	}
	if !writesFile(map[string]any{"tool_name": "Write"}) || !writesFile(map[string]any{"tool_name": "Edit"}) || writesFile(map[string]any{"tool_name": "Read"}) {
		t.Fatal("tool_name write classification incorrect")
	}
	if observeID := func() string { var q payload; observe(&q, map[string]any{"session_id": ""}); return q.SessionID }(); observeID != "" {
		t.Fatal("empty session id accepted")
	}
}
