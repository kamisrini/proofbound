//go:build integration

package main

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

func TestSyncChecksIngestsAndDeduplicates(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}
	resetIntegrationDatabase(t, databaseURL)
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	spool := filepath.Join(root, ".vera", "spool")
	if err := os.MkdirAll(spool, 0o755); err != nil {
		t.Fatal(err)
	}
	runID := ulid.MustNew(ulid.Timestamp(time.Now()), crand.Reader).String()
	witness := map[string]any{
		"schema": "vera.witness.v1", "run_id": runID, "command": "make check",
		"exit_code": 0, "started_at": "2026-08-24T12:00:00Z", "finished_at": "2026-08-24T12:00:01Z",
		"duration_ms": 1000, "output_sha256": strings.Repeat("a", 64), "git_sha": strings.Repeat("b", 40),
		"git_dirty": false, "tool_versions": map[string]string{"go": "go1.26", "golangci_lint": "v2", "make": "GNU Make 4"},
	}
	data, err := json.Marshal(witness)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(spool, runID+".json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	t.Setenv("DATABASE_URL", databaseURL)

	var output, stderr bytes.Buffer
	if code := run(context.Background(), []string{"sync", "checks"}, &output, &stderr); code != 0 || output.String() != "listed=1 appended=1 existing=0\n" || stderr.Len() != 0 {
		t.Fatalf("first code=%d output=%q stderr=%q", code, output.String(), stderr.String())
	}
	output.Reset()
	if code := run(context.Background(), []string{"sync", "checks"}, &output, &stderr); code != 0 || output.String() != "listed=1 appended=0 existing=1\n" || stderr.Len() != 0 {
		t.Fatalf("second code=%d output=%q stderr=%q", code, output.String(), stderr.String())
	}
}

func TestSyncChecksReportsMalformedWitness(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}
	resetIntegrationDatabase(t, databaseURL)
	root := t.TempDir()
	spool := filepath.Join(root, ".vera", "spool")
	if err := os.MkdirAll(spool, 0o755); err != nil {
		t.Fatal(err)
	}
	runID := ulid.MustNew(ulid.Timestamp(time.Now()), crand.Reader).String()
	if err := os.WriteFile(filepath.Join(spool, runID+".json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := syncChecks(context.Background(), root, databaseURL, &output); err == nil || output.Len() != 0 {
		t.Fatalf("output=%q error=%v", output.String(), err)
	}
}

func resetIntegrationDatabase(t *testing.T, databaseURL string) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var exists bool
	if err := pool.QueryRow(context.Background(), `SELECT to_regclass('public.events') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		return
	}
	if _, err := pool.Exec(context.Background(), `TRUNCATE events, sync_runs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `DROP TABLE IF EXISTS projection_meta, commits_view, checks_view, sessions_view, reviews_view CASCADE`); err != nil {
		t.Fatal(err)
	}
}
