package twin

// Unit-3 runtime fixture (spec §4). Runs ONLY when PROOFBOUND_P7_PG_ARCHIVE points at the prepared
// verified-from-source PostgreSQL subset archive whose digests match the committed manifest; it
// t.Skips otherwise (the committed case stays UNVERIFIED — no PASS by absence, §5.1). A green skip
// is NOT acceptance — the §9.3a acceptance run sets the env var and witnesses LOCKED+PASS live.
//
// This increment asserts the TWIN-INV-18 launched-process witness against the REAL running
// postgres: the resolved /proc/<pid>/exe digest equals the declared bin/postgres digest, proving a
// real from-source postgres ran this replay. The behavior-lock/case-manifest authoring
// (TWIN-INV-22 determinism + the nine case outcomes) lands in the next increment.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type preparedArchiveManifest struct {
	PostgresBinarySHA256 string `json:"postgres_binary_sha256"`
	Archive              struct {
		SHA256 string `json:"sha256"`
	} `json:"archive"`
	Inventory []struct {
		Path   string `json:"path"`
		Mode   string `json:"mode"`
		SHA256 string `json:"sha256"`
	} `json:"inventory"`
}

// preparedArchiveManifestPath is the committed non-binary identity of the prepared archive.
const preparedArchiveManifestPath = "../../../docs/verification/p7-replay-lock-fixtures/postgres-16.4-subset-archive.json"

func loadVerifiedPostgresInput(t *testing.T) VerifiedPostgresInput {
	t.Helper()
	archivePath := os.Getenv("PROOFBOUND_P7_PG_ARCHIVE")
	if archivePath == "" {
		t.Skip("PROOFBOUND_P7_PG_ARCHIVE unset; runtime fixture stays UNVERIFIED (no PASS by absence)")
	}
	raw, err := os.ReadFile(filepath.Clean(preparedArchiveManifestPath))
	if err != nil {
		t.Fatalf("read committed prepared-archive manifest: %v", err)
	}
	var m preparedArchiveManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	bytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive %s: %v", archivePath, err)
	}
	inv := make([]VerifiedArchiveEntry, 0, len(m.Inventory))
	for _, e := range m.Inventory {
		inv = append(inv, VerifiedArchiveEntry{Path: e.Path, Mode: e.Mode, SHA256: e.SHA256})
	}
	return VerifiedPostgresInput{
		ArchiveBytes:         bytes,
		ArchiveSHA256:        m.Archive.SHA256,
		PostgresBinarySHA256: m.PostgresBinarySHA256,
		Inventory:            inv,
	}
}

// openVerifiedSourceStore opens a source store backed by the SAME verified-from-source archive, so
// the whole fixture stays fetch-free (no zonky download for the source half).
func openVerifiedSourceStore(t *testing.T, input VerifiedPostgresInput) *store.Store {
	t.Helper()
	binDir, err := stageVerifiedPostgres(context.Background(), t.TempDir(), input)
	if err != nil {
		t.Fatalf("stage source binaries: %v", err)
	}
	port, err := temporaryPort()
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(context.Background(), store.Config{Root: t.TempDir(), Port: port, BinariesDir: binDir})
	if err != nil {
		t.Fatalf("open verified source store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestRuntimeFixtureWitnessesRealLaunchedPostgres(t *testing.T) {
	input := loadVerifiedPostgresInput(t)
	src := openVerifiedSourceStore(t, input)

	// Synthetic candidate over an empty projectable prefix (fixtures, never production
	// observations — permitted by the behavior-lock SPEC). The projector supports the GitHub
	// deployment pair.
	req := Request{ThroughSeq: 0, Candidate: []Candidate{{Seq: 30, Event: githubDeploymentEvent(30)}}}
	result, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src, req, emptySnapshot(), input)
	if err != nil {
		t.Fatalf("replay with verified postgres: %v", err)
	}

	// TWIN-INV-18: the running postgres image equals the declared digest, surfaced via the sole
	// producer witnessLaunchedProcess (not a literal).
	if result.launch.method != "proc-exe" {
		t.Fatalf("launch witness method = %q, want proc-exe", result.launch.method)
	}
	if result.launch.resolvedPID <= 0 {
		t.Fatalf("launch witness resolvedPID = %d, want > 0", result.launch.resolvedPID)
	}
	if result.launch.resolvedSHA256 != input.PostgresBinarySHA256 {
		t.Fatalf("launch witness resolved %s, want declared %s", result.launch.resolvedSHA256, input.PostgresBinarySHA256)
	}
	if result.Proof.Schema != "vera.replay.v1" {
		t.Fatalf("proof schema = %q, want vera.replay.v1", result.Proof.Schema)
	}
	t.Logf("REAL postgres launched; /proc witness resolved %s (pid %d); proof %s",
		result.launch.resolvedSHA256, result.launch.resolvedPID, result.Proof.Schema)
}
