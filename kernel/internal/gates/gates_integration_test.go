//go:build integration

package gates

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func gateIntegrationStore(t *testing.T) *store.Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	s, err := store.Open(context.Background(), store.Config{Root: filepath.Join(t.TempDir(), ".proofbound"), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if url != "" {
		pool, err := pgxpool.New(context.Background(), url)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		if _, err := pool.Exec(context.Background(), `TRUNCATE events, sync_runs RESTART IDENTITY CASCADE`); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(context.Background(), `DROP TABLE IF EXISTS projection_meta, requirement_reviews_view, obligation_verdicts_view, commit_intents_view, intent_targets_view, requirement_obligations_view, change_intents_view, requirements_view, business_decisions_view, commits_view, checks_view, sessions_view, reviews_view, github_delivery_view CASCADE`); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func appendGateRaw(t *testing.T, s *store.Store, source core.Source, kind core.Kind, native string, payload []byte) store.Record {
	t.Helper()
	ids := testIDs(t)
	event, err := ids.NewEvent(core.NewEventParams{Source: source, NativeID: native, Kind: kind, OccurredAt: time.Now(), Payload: payload, ConnectorVersion: "test/1"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.BeginSync(context.Background(), "intent-gate-test")
	if err != nil {
		t.Fatal(err)
	}
	record, _, err := run.Append(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	return record
}

func TestIntentIntegrityGateStatesAndProof(t *testing.T) {
	definition := Definition{Schema: Version, ID: "intent-reference-integrity", Description: "test", Expires: "2099-01-01", Mode: "canary", Rule: "intent-reference-integrity"}
	t.Run("good", func(t *testing.T) {
		s := gateIntegrationStore(t)
		sha := strings.Repeat("a", 40)
		payload := []byte(`{"sha":"` + sha + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:00:00Z","subject":"plain","files_touched":null,"cited_decisions":null,"intent_refs":null}`)
		proof := appendGateRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, payload)
		result, err := Evaluate(context.Background(), s, definition)
		if err != nil || result.State != StatePass || result.EventID != proof.Event.ID.String() {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	})
	t.Run("bad", func(t *testing.T) {
		s := gateIntegrationStore(t)
		sha := strings.Repeat("b", 40)
		payload := []byte(`{"sha":"` + sha + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:00:00Z","subject":"bad","files_touched":null,"cited_decisions":null,"intent_refs":[{"provider":"records","record_id":"CI-missing-item-acde12","artifact_sha256":"` + strings.Repeat("c", 64) + `"}]}`)
		proof := appendGateRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, payload)
		result, err := Evaluate(context.Background(), s, definition)
		if err != nil || result.State != StateBlocked || !result.WouldBlock || result.EventID != proof.Event.ID.String() {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	})
}

func TestLatestIntentReferenceProofSelectsRecordsAndExcludesUnrelatedEvents(t *testing.T) {
	s := gateIntegrationStore(t)
	intent := appendGateRaw(t, s, core.SourceIntentRecords, core.KindBusinessDecision, "BD-proof-acde12", []byte(`{}`))
	appendGateRaw(t, s, core.SourceSessions, core.KindSessionObserved, "unrelated", []byte(`{}`))
	appendGateRaw(t, s, core.SourceGit, core.KindCheckRun, "wrong-git-kind", []byte(`{}`))
	appendGateRaw(t, s, core.SourceChecks, core.KindCommitRecorded, "wrong-commit-source", []byte(`{}`))
	latest, err := latestIntentProof(context.Background(), s, "intent-reference-integrity", "")
	if err != nil {
		t.Fatal(err)
	}
	if latest == nil || latest.Event.ID != intent.Event.ID {
		t.Fatalf("latest=%+v intent=%+v", latest, intent)
	}
}

func TestLatestIntentVerdictProofSelectsReviewsAndExcludesUnrelatedEvents(t *testing.T) {
	s := gateIntegrationStore(t)
	appendGateRaw(t, s, core.SourceReviews, core.KindReviewVerdict, "OV-proof-acde12", []byte(`{}`))
	review := appendGateRaw(t, s, core.SourceReviews, core.KindRequirementReview, "RR-proof-acde12", []byte(`{}`))
	appendGateRaw(t, s, core.SourceSessions, core.KindSessionObserved, "unrelated", []byte(`{}`))
	appendGateRaw(t, s, core.SourceChecks, core.KindRequirementReview, "wrong-review-source", []byte(`{}`))
	latest, err := latestIntentProof(context.Background(), s, "intent-verdict-integrity", "")
	if err != nil {
		t.Fatal(err)
	}
	if latest == nil || latest.Event.ID != review.Event.ID {
		t.Fatalf("latest=%+v review=%+v", latest, review)
	}
}

func TestCurrentGateEstateEventRoutes(t *testing.T) {
	pairs := []struct {
		source  core.Source
		kind    core.Kind
		consume bool
	}{
		{core.SourceGit, core.KindCommitRecorded, true},
		{core.SourceChecks, core.KindCheckRun, true},
		{core.SourceSessions, core.KindSessionObserved, false},
		{core.SourceReviews, core.KindReviewVerdict, true},
		{core.SourceReviews, core.KindRequirementReview, true},
		{core.SourceGitHub, core.KindGitHubWorkflow, false},
		{core.SourceGitHub, core.KindGitHubDeployment, false},
		{core.SourceIntentRecords, core.KindBusinessDecision, true},
		{core.SourceIntentRecords, core.KindRequirement, true},
		{core.SourceIntentRecords, core.KindChangeIntent, true},
		{core.SourceIntentSpecdir, core.KindBusinessDecision, true},
		{core.SourceIntentSpecdir, core.KindRequirement, true},
		{core.SourceIntentSpecdir, core.KindChangeIntent, true},
	}
	definitions, err := LoadDir("../../../gates")
	if err != nil {
		t.Fatal(err)
	}
	s := gateIntegrationStore(t)
	for i, pair := range pairs {
		record := appendGateRaw(t, s, pair.source, pair.kind, fmt.Sprintf("route-%d", i), []byte(`{"route":true}`))
		selected := false
		for _, definition := range definitions {
			if definition.Rule == "" && definition.Source == pair.source && definition.Kind == pair.kind {
				selected = true
			}
			if definition.Rule != "" {
				latest, err := latestIntentProof(context.Background(), s, definition.Rule, record.Event.NativeID)
				if err != nil {
					t.Fatal(err)
				}
				if latest != nil && latest.Event.ID == record.Event.ID {
					selected = true
				}
			}
		}
		if selected != pair.consume {
			t.Errorf("route %s/%s consume=%v want %v", pair.source, pair.kind, selected, pair.consume)
		}
	}
}

func TestLatestDeliveryProofRequiresExactSourceKindAndCommit(t *testing.T) {
	s := gateIntegrationStore(t)
	commit := appendGateRaw(t, s, core.SourceGit, core.KindCommitRecorded, "target-commit", []byte(`{"case":"target"}`))
	appendGateRaw(t, s, core.SourceChecks, core.KindCommitRecorded, "target-commit", []byte(`{"case":"source"}`))
	appendGateRaw(t, s, core.SourceGit, core.KindCheckRun, "target-commit", []byte(`{"case":"kind"}`))
	appendGateRaw(t, s, core.SourceGit, core.KindCommitRecorded, "other-commit", []byte(`{"case":"native"}`))
	latest, err := latestIntentProof(context.Background(), s, "intent-delivery-readiness", "target-commit")
	if err != nil {
		t.Fatal(err)
	}
	if latest == nil || latest.Event.ID != commit.Event.ID {
		t.Fatalf("latest=%+v commit=%+v", latest, commit)
	}
}

func TestIntentDeliveryReadiness(t *testing.T) {
	s := gateIntegrationStore(t)
	sha := strings.Repeat("d", 40)
	payload := []byte(`{"sha":"` + sha + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:00:00Z","subject":"delivery","files_touched":null,"cited_decisions":null,"intent_refs":null}`)
	proof := appendGateRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, payload)
	definition := Definition{Schema: Version, ID: "intent-delivery-readiness", Description: "test", Expires: "2099-01-01", Mode: "canary", Rule: "intent-delivery-readiness", ScopeCommit: sha}
	blocked, err := Evaluate(context.Background(), s, definition)
	if err != nil || blocked.State != StateBlocked || blocked.EventID != proof.Event.ID.String() {
		t.Fatalf("blocked=%+v err=%v", blocked, err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		statements := []struct {
			sql  string
			args []any
		}{{`INSERT INTO intent_targets_view(intent_source,intent_id,intent_artifact_sha256,requirement_source,requirement_id,requirement_artifact_sha256,obligation_id,relation,event_id,seq) VALUES('intent.records','CI-ready-item-acde12',$1,'intent.records','BR-ready-item-acde12',$2,'O-1','implements',$3,$4)`, []any{strings.Repeat("e", 64), strings.Repeat("f", 64), proof.Event.ID.String(), proof.Seq}}, {`INSERT INTO commit_intents_view(commit_sha,provider,intent_id,intent_artifact_sha256,event_id,seq) VALUES($1,'records','CI-ready-item-acde12',$2,$3,$4)`, []any{sha, strings.Repeat("e", 64), proof.Event.ID.String(), proof.Seq}}, {`INSERT INTO obligation_verdicts_view(verdict_id,commit_sha,intent_source,intent_id,intent_artifact_sha256,requirement_source,requirement_id,requirement_artifact_sha256,obligation_id,outcome,evidence_event_ids,event_id,seq) VALUES('OV-ready',$1,'intent.records','CI-ready-item-acde12',$2,'intent.records','BR-ready-item-acde12',$3,'O-1','SATISFIED','[]',$4,$5)`, []any{sha, strings.Repeat("e", 64), strings.Repeat("f", 64), proof.Event.ID.String(), proof.Seq}}, {`INSERT INTO requirement_reviews_view(review_id,requirement_source,requirement_id,requirement_artifact_sha256,obligation_id,outcome,finding,declared_reviewer,event_id,seq) VALUES('RR-ready','intent.records','BR-ready-item-acde12',$1,'O-1','VERIFIABLE','','reviewer',$2,$3)`, []any{strings.Repeat("f", 64), proof.Event.ID.String(), proof.Seq}}}
		for _, statement := range statements {
			if _, err := tx.Exec(ctx, statement.sql, statement.args...); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	pass, err := Evaluate(context.Background(), s, definition)
	if err != nil || pass.State != StatePass {
		t.Fatalf("pass=%+v err=%v", pass, err)
	}
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

func testRepositoryRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("PROOFBOUND_TEST_REPO_ROOT")
	if root == "" {
		workingDirectory, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		root = filepath.Join(workingDirectory, "..", "..", "..")
	}
	return root
}

func loadedGateDefinition(t *testing.T, name string) Definition {
	t.Helper()
	root := testRepositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "gates", name))
	if err != nil {
		t.Fatal(err)
	}
	definition, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

func TestLoadedIndexGateMatchesCommandAndExitCode(t *testing.T) {
	s := gateIntegrationStore(t)
	ids := testIDs(t)
	definition := loadedGateDefinition(t, "index-check-success.yaml")
	appendCheckEventWithCommand(t, s, ids, "index-bad", "make index-check", 1)
	appendCheckEventWithCommand(t, s, ids, "other", "make check", 0)
	result, err := Evaluate(context.Background(), s, definition)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != StateBlocked || !result.WouldBlock || result.EventID == "" {
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
	definition := loadedGateDefinition(t, "spec-numbering-success.yaml")
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
	definition := loadedGateDefinition(t, "invariant-table-success.yaml")
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
	definition := loadedGateDefinition(t, "link-success.yaml")
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
	definition := loadedGateDefinition(t, "kernel-check-success.yaml")
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
	root := testRepositoryRoot(t)
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
	copyScript, err := os.ReadFile(filepath.Join(root, "kernel", "scripts", "check-witness.sh"))
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
	cmd.Env = append(os.Environ(), "PROOFBOUND_CHECK_TARGET=kernel-check")
	if output, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("bad kernel-check unexpectedly passed: %s", output)
	}
	connector, err := checks.New(&checks.Deps{SpoolDir: filepath.Join(repo, ".proofbound", "spool"), IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
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
	definition := loadedGateDefinition(t, "kernel-check-success.yaml")
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
