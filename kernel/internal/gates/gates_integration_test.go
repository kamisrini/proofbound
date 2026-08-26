//go:build integration

package gates

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
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
	return appendCheckEventWithCommand(t, s, ids, native, "make check", exitCode)
}

func appendCheckEventWithCommand(t *testing.T, s *store.Store, ids *core.IDGenerator, native, command string, exitCode int) store.Record {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"schema": "vera.witness.v1", "run_id": native, "command": command, "exit_code": exitCode, "started_at": time.Now().UTC(), "finished_at": time.Now().UTC(), "duration_ms": 1, "output_sha256": "0000000000000000000000000000000000000000000000000000000000000000", "git_sha": "0000000000000000000000000000000000000000", "git_dirty": false, "tool_versions": map[string]string{"go": "go", "golangci_lint": "lint", "make": "make"}})
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

func TestLoadedIndexGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "index-check-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	appendCheckEventWithCommand(t, s, ids, "index-bad", "make index-check", 1)
	appendCheckEventWithCommand(t, s, ids, "other", "make check", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateUnknown || result.WouldBlock || result.EventID != "" {
		t.Fatalf("result=%+v", result)
	}
	pass := appendCheckEventWithCommand(t, s, ids, "index-good", "make index-check", 0)
	result, err = Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.EventID != pass.Event.ID.String() || result.Seq != pass.Seq {
		t.Fatalf("result=%+v pass=%+v", result, pass)
	}
}

func TestLoadedSpecNumberingGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "spec-numbering-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	wrong := appendCheckEventWithCommand(t, s, ids, "spec-wrong", "make law-citation-lint", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateUnknown || result.WouldBlock || result.EventID != "" {
		t.Fatalf("result=%+v wrong=%+v", result, wrong)
	}
	pass := appendCheckEventWithCommand(t, s, ids, "spec-good", "make spec-numbering-lint", 0)
	result, err = Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.EventID != pass.Event.ID.String() || result.Seq != pass.Seq {
		t.Fatalf("result=%+v pass=%+v", result, pass)
	}
}

func TestLoadedInvariantTableGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "invariant-table-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	wrong := appendCheckEventWithCommand(t, s, ids, "table-wrong", "make spec-numbering-lint", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateUnknown || result.WouldBlock || result.EventID != "" {
		t.Fatalf("result=%+v wrong=%+v", result, wrong)
	}
	pass := appendCheckEventWithCommand(t, s, ids, "table-good", "make invariant-table-lint", 0)
	result, err = Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.EventID != pass.Event.ID.String() || result.Seq != pass.Seq {
		t.Fatalf("result=%+v pass=%+v", result, pass)
	}
}

func TestLoadedLinkGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "link-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	wrong := appendCheckEventWithCommand(t, s, ids, "link-wrong", "make law-citation-lint", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateUnknown || result.WouldBlock || result.EventID != "" {
		t.Fatalf("result=%+v wrong=%+v", result, wrong)
	}
	pass := appendCheckEventWithCommand(t, s, ids, "link-good", "make link-lint", 0)
	result, err = Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.EventID != pass.Event.ID.String() || result.Seq != pass.Seq {
		t.Fatalf("result=%+v pass=%+v", result, pass)
	}
}

func TestLoadedKernelCheckGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "kernel-check-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	wrong := appendCheckEventWithCommand(t, s, ids, "kernel-wrong", "make check", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateUnknown || result.WouldBlock || result.EventID != "" {
		t.Fatalf("result=%+v wrong=%+v", result, wrong)
	}
	pass := appendCheckEventWithCommand(t, s, ids, "kernel-good", "make kernel-check", 0)
	result, err = Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StatePass || result.EventID != pass.Event.ID.String() || result.Seq != pass.Seq {
		t.Fatalf("result=%+v pass=%+v", result, pass)
	}
}

func gateDefinition() Definition {
	return Definition{Schema: Version, ID: "make-check-success", Description: "test", Expires: "2099-01-01", Mode: "canary", Source: core.SourceChecks, Kind: core.KindCheckRun, Condition: Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
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
	unknown, err := Evaluate(context.Background(), s, Definition{Schema: Version, ID: "other", Description: "other", Expires: "2099-01-01", Mode: "canary", Source: core.SourceGit, Kind: core.KindCommitRecorded, Condition: Condition{Field: "sha", Equals: json.RawMessage(`"x"`)}})
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

func TestCanaryThenEnforceRejectsBadWitness(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("Makefile", "kernel-check:\n\t@test ! -f bad.go\n")
	write("bad.go", "package bad\n\nvar Bad = true\n")
	for _, args := range [][]string{{"init"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "test"}, {"add", "Makefile", "bad.go"}, {"commit", "-m", "bad change"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	copyScript, err := os.ReadFile(filepath.Join(root, "..", "..", "scripts", "check-witness.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repo, "kernel", "scripts", "check-witness.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, copyScript, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", script)
	cmd.Dir = repo
	if output, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("bad kernel-check unexpectedly passed: %s", output)
	}
	connector, err := checks.New(&checks.Deps{SpoolDir: filepath.Join(repo, ".vera", "spool"), IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.BeginSync(context.Background(), "bad-change-proof")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connector.Sync(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := run.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "..", "..", "gates", "kernel-check-success.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	definition.Mode = "canary"
	canary, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if canary.State != StateBlocked || !canary.WouldBlock || canary.EventID == "" || canary.Seq == 0 {
		t.Fatalf("canary=%+v", canary)
	}
	definition.Mode = "enforce"
	if err := definition.EnforceReady(); err != nil {
		t.Fatal(err)
	}
	if err := Enforce(canary); err == nil {
		t.Fatal("enforce accepted a canary-blocked bad witness")
	}
}
