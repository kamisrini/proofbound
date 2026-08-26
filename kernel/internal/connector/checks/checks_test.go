package checks

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const fixtureRunID = "01KZFAPQ00NENTQAXBNENTQAXB"

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
	if _, exists := a.seen[key]; exists {
		return store.Record{Event: event}, false, nil
	}
	a.seen[key] = struct{}{}
	a.events = append(a.events, event)
	return store.Record{Seq: int64(len(a.events)), Event: event}, true, nil
}

func TestWitness_StrictValidation(t *testing.T) {
	valid := testWitness(fixtureRunID)
	mutations := map[string]func(*Witness){
		"schema":       func(w *Witness) { w.Schema = "other" },
		"run id":       func(w *Witness) { w.RunID = "invalid" },
		"command":      func(w *Witness) { w.Command = "make -n" },
		"exit low":     func(w *Witness) { w.ExitCode = -1 },
		"exit high":    func(w *Witness) { w.ExitCode = 256 },
		"start":        func(w *Witness) { w.StartedAt = time.Time{} },
		"finish":       func(w *Witness) { w.FinishedAt = w.StartedAt.Add(-time.Second) },
		"duration":     func(w *Witness) { w.DurationMS = -1 },
		"output hash":  func(w *Witness) { w.OutputSHA256 = "ABC" },
		"git hash":     func(w *Witness) { w.GitSHA = "abc" },
		"go version":   func(w *Witness) { w.ToolVersions.Go = "" },
		"lint version": func(w *Witness) { w.ToolVersions.GolangCILint = " " },
		"make version": func(w *Witness) { w.ToolVersions.Make = "" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			witness := valid
			mutate(&witness)
			if err := witness.validate(); err == nil {
				t.Fatalf("accepted %+v", witness)
			}
		})
	}
	dir := t.TempDir()
	path := filepath.Join(dir, fixtureRunID+".json")
	writeRaw(t, path, append(mustJSON(t, valid), []byte(` {}`)...))
	if _, err := readWitness(path); err == nil {
		t.Fatal("accepted trailing JSON")
	}
	data := strings.TrimSuffix(string(mustJSON(t, valid)), "}") + `,"unknown":true}`
	writeRaw(t, path, []byte(data))
	if _, err := readWitness(path); err == nil {
		t.Fatal("accepted unknown field")
	}
	data = `{ "schema":"other",` + strings.TrimPrefix(string(mustJSON(t, valid)), "{")
	writeRaw(t, path, []byte(data))
	if _, err := readWitness(path); err == nil {
		t.Fatal("accepted duplicate field")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(mustJSON(t, valid), &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		t.Run("missing "+name, func(t *testing.T) {
			copy := make(map[string]json.RawMessage, len(fields)-1)
			for key, value := range fields {
				if key != name {
					copy[key] = value
				}
			}
			writeRaw(t, path, mustJSON(t, copy))
			if _, err := readWitness(path); err == nil {
				t.Fatalf("accepted witness without %s", name)
			}
		})
	}
	for name := range fields {
		t.Run("null "+name, func(t *testing.T) {
			copy := make(map[string]json.RawMessage, len(fields))
			for key, value := range fields {
				copy[key] = value
			}
			copy[name] = json.RawMessage("null")
			writeRaw(t, path, mustJSON(t, copy))
			if _, err := readWitness(path); err == nil {
				t.Fatalf("accepted null %s", name)
			}
		})
	}
}

func TestSync_MintsCheckRunEvent(t *testing.T) {
	dir := t.TempDir()
	witness := testWitness(fixtureRunID)
	writeWitness(t, dir, witness)
	connector := testConnector(t, dir)
	appender := &memoryAppender{}
	result, err := connector.Sync(context.Background(), appender)
	if err != nil || result.Listed != 1 || result.Appended != 1 || result.Existing != 0 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	event := appender.events[0]
	if event.Source != core.SourceChecks || event.Kind != core.KindCheckRun || event.NativeID != fixtureRunID || event.ConnectorVersion != Version || !event.OccurredAt.Equal(witness.StartedAt) {
		t.Fatalf("event=%+v", event)
	}
}

func TestSync_ContentBindsIdentity(t *testing.T) {
	dir := t.TempDir()
	witness := testWitness(fixtureRunID)
	writeWitness(t, dir, witness)
	connector := testConnector(t, dir)
	appender := &memoryAppender{}
	if _, err := connector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	witness.ExitCode = 1
	writeWitness(t, dir, witness)
	if _, err := connector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	if len(appender.events) != 2 || appender.events[0].NativeID != appender.events[1].NativeID || appender.events[0].ContentSHA == appender.events[1].ContentSHA {
		t.Fatalf("events=%+v", appender.events)
	}
}

func TestSync_SecondSyncAppendsNothing(t *testing.T) {
	dir := t.TempDir()
	writeWitness(t, dir, testWitness(fixtureRunID))
	connector := testConnector(t, dir)
	appender := &memoryAppender{}
	if _, err := connector.Sync(context.Background(), appender); err != nil {
		t.Fatal(err)
	}
	result, err := connector.Sync(context.Background(), appender)
	if err != nil || result.Appended != 0 || result.Existing != 1 || len(appender.events) != 1 {
		t.Fatalf("result=%+v events=%d error=%v", result, len(appender.events), err)
	}
}

func TestSync_MalformedFileFailsClosed(t *testing.T) {
	dir := t.TempDir()
	first := testWitness("01KZFAPQ00NENTQAXBNENTQAXA")
	writeWitness(t, dir, first)
	writeRaw(t, filepath.Join(dir, "01KZFAPQ00NENTQAXBNENTQAXB.json"), []byte(`{}`))
	writeWitness(t, dir, testWitness("01KZFAPQ00NENTQAXBNENTQAXC"))
	appender := &memoryAppender{}
	result, err := testConnector(t, dir).Sync(context.Background(), appender)
	if err == nil || result.Listed != 3 || result.Appended != 1 || len(appender.events) != 1 || string(result.Cursor) != `["01KZFAPQ00NENTQAXBNENTQAXA.json"]` {
		t.Fatalf("result=%+v events=%d error=%v", result, len(appender.events), err)
	}
}

func TestSync_NeverDeletesSpoolFiles(t *testing.T) {
	dir := t.TempDir()
	witness := testWitness(fixtureRunID)
	writeWitness(t, dir, witness)
	before, err := os.ReadFile(filepath.Join(dir, witness.RunID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testConnector(t, dir).Sync(context.Background(), &memoryAppender{}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, witness.RunID+".json"))
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Fatalf("after=%s error=%v", after, err)
	}
	source, err := os.ReadFile("checks.go")
	if err != nil || strings.Contains(string(source), "os.Remove") || strings.Contains(string(source), "os.Rename") {
		t.Fatalf("source deletion route error=%v", err)
	}
}

func TestSync_RequiresFilenameRunID(t *testing.T) {
	dir := t.TempDir()
	writeRaw(t, filepath.Join(dir, "01KZFAPQ00NENTQAXBNENTQAXC.json"), mustJSON(t, testWitness(fixtureRunID)))
	if _, err := testConnector(t, dir).Sync(context.Background(), &memoryAppender{}); err == nil {
		t.Fatal("filename relabelled witness")
	}
}

func TestNew_RequiresEveryDependency(t *testing.T) {
	ids := testIDs(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for name, deps := range map[string]*Deps{
		"deps":   nil,
		"spool":  {IDs: ids, Logger: logger},
		"ids":    {SpoolDir: t.TempDir(), Logger: logger},
		"logger": {SpoolDir: t.TempDir(), IDs: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if connector, err := New(deps); err == nil || connector != nil {
				t.Fatalf("connector=%v error=%v", connector, err)
			}
		})
	}
}

func TestSync_RejectsEveryUninitializedRoute(t *testing.T) {
	for name, connector := range map[string]*Connector{
		"nil connector": nil,
		"nil ids":       {spoolDir: t.TempDir()},
		"empty spool":   {ids: testIDs(t)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := connector.Sync(context.Background(), &memoryAppender{}); err == nil {
				t.Fatal("accepted uninitialized connector")
			}
		})
	}
}

func TestSync_RejectsTypedNilAppenderBeforeListing(t *testing.T) {
	var appender *memoryAppender
	for _, spoolDir := range []string{t.TempDir(), filepath.Join(t.TempDir(), "absent")} {
		result, err := testConnector(t, spoolDir).Sync(context.Background(), appender)
		if err == nil || result.Listed != 0 || result.Appended != 0 || result.Existing != 0 || result.Cursor != nil {
			t.Fatalf("spool=%s result=%+v error=%v", spoolDir, result, err)
		}
	}
}

func TestSync_CursorIsSortedFilenames(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "ignored.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeRaw(t, filepath.Join(dir, "ignored.txt"), []byte("not evidence"))
	for _, id := range []string{"01KZFAPQ00NENTQAXBNENTQAXC", "01KZFAPQ00NENTQAXBNENTQAXA", "01KZFAPQ00NENTQAXBNENTQAXB"} {
		writeWitness(t, dir, testWitness(id))
	}
	result, err := testConnector(t, dir).Sync(context.Background(), &memoryAppender{})
	if err != nil || string(result.Cursor) != `["01KZFAPQ00NENTQAXBNENTQAXA.json","01KZFAPQ00NENTQAXBNENTQAXB.json","01KZFAPQ00NENTQAXBNENTQAXC.json"]` {
		t.Fatalf("cursor=%s error=%v", result.Cursor, err)
	}
}

func testWitness(runID string) Witness {
	return Witness{
		Schema:       "vera.witness.v1",
		RunID:        runID,
		Command:      "make check",
		ExitCode:     0,
		StartedAt:    time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
		FinishedAt:   time.Date(2026, 8, 24, 12, 0, 2, 0, time.UTC),
		DurationMS:   2000,
		OutputSHA256: strings.Repeat("a", 64),
		GitSHA:       strings.Repeat("b", 40),
		GitDirty:     true,
		ToolVersions: ToolVersions{Go: "go1.26", GolangCILint: "v2", Make: "GNU Make 4"},
	}
}

func testConnector(t *testing.T, dir string) *Connector {
	t.Helper()
	connector, err := New(&Deps{SpoolDir: dir, IDs: testIDs(t), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	return connector
}

func testIDs(t *testing.T) *core.IDGenerator {
	t.Helper()
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func writeWitness(t *testing.T, dir string, witness Witness) {
	t.Helper()
	writeRaw(t, filepath.Join(dir, witness.RunID+".json"), mustJSON(t, witness))
}

func writeRaw(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
