//go:build integration

package gates

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func gateIntegrationStore(t *testing.T) *store.Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required")
	}
	s, err := store.Open(context.Background(), store.Config{Root: filepath.Join(t.TempDir(), ".vera"), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func appendCheckEvent(t *testing.T, s *store.Store, ids *core.IDGenerator, native string, exitCode int) store.Record {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"schema": "vera.witness.v1", "run_id": native, "command": "make check", "exit_code": exitCode, "started_at": time.Now().UTC(), "finished_at": time.Now().UTC(), "duration_ms": 1, "output_sha256": "0000000000000000000000000000000000000000000000000000000000000000", "git_sha": "0000000000000000000000000000000000000000", "git_dirty": false, "tool_versions": map[string]string{"go": "go", "golangci_lint": "lint", "make": "make"}})
	e, err := ids.NewEvent(core.NewEventParams{Source: core.SourceChecks, NativeID: native, Kind: core.KindCheckRun, OccurredAt: time.Now(), Payload: payload, ConnectorVersion: "test/1"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.BeginSync(context.Background(), "gates-test")
	if err != nil {
		t.Fatal(err)
	}
	record, _, err := run.Append(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	return record
}

func gateDefinition() Definition {
	return Definition{Schema: Version, ID: "make-check-success", Description: "test", Mode: "canary", Source: core.SourceChecks, Kind: core.KindCheckRun, Condition: Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
}

func testIDs(t *testing.T) *core.IDGenerator {
	t.Helper()
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func TestEvaluateUsesLatestMatchingEvent(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	appendCheckEvent(t, s, ids, "bad", 1)
	latest := appendCheckEvent(t, s, ids, "good", 0)
	result, err := Evaluate(context.Background(), s, gateDefinition())
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.Seq != latest.Seq || result.EventID != latest.Event.ID.String() {
		t.Fatalf("result=%+v latest=%+v", result, latest)
	}
}

func TestEvaluateStatesAndProof(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	blocked := appendCheckEvent(t, s, ids, "bad", 1)
	result, err := Evaluate(context.Background(), s, gateDefinition())
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateBlocked || !result.WouldBlock || result.Seq != blocked.Seq || result.EventID != blocked.Event.ID.String() {
		t.Fatalf("result=%+v blocked=%+v", result, blocked)
	}
	unknown, err := Evaluate(context.Background(), s, Definition{Schema: Version, ID: "other", Description: "other", Mode: "canary", Source: core.SourceGit, Kind: core.KindCommitRecorded, Condition: Condition{Field: "sha", Equals: json.RawMessage(`"x"`)}})
	if err != nil {
		t.Fatal(err)
	}
	if unknown.State != StateUnknown || unknown.Seq != 0 || unknown.EventID != "" {
		t.Fatalf("unknown=%+v", unknown)
	}
}

func TestEvaluateIsReadOnly(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	appendCheckEvent(t, s, ids, "good", 0)
	var before int
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(store.Record) error { before++; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(context.Background(), s, gateDefinition()); err != nil {
		t.Fatal(err)
	}
	var after int
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(store.Record) error { after++; return nil }); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("event count changed: before=%d after=%d", before, after)
	}
}
