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

func appendDeployment(t *testing.T, s *store.Store, id int) {
	t.Helper()
	sy, err := s.BeginSync(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := sy.Append(context.Background(), githubDeploymentEvent(id)); err != nil {
		t.Fatal(err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func sourceLedgerDigest(t *testing.T, s *store.Store) string {
	t.Helper()
	var recs []store.Record
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(r store.Record) error {
		recs = append(recs, r)
		return nil
	}); err != nil {
		t.Fatalf("read source events: %v", err)
	}
	return digestRecords(recs)
}

// TestRuntimeFixtureLedgerPreservationWitness is TWIN-INV-20: the source store's event set is
// byte-identical before and after the isolated replay. The positive control proves the witness
// detects a real mutation (it does not merely restate the read-only invariant).
func TestRuntimeFixtureLedgerPreservationWitness(t *testing.T) {
	input := loadVerifiedPostgresInput(t)
	src := openVerifiedSourceStore(t, input)
	appendDeployment(t, src, 5)

	before := sourceLedgerDigest(t, src)
	req := Request{ThroughSeq: 5, Candidate: []Candidate{{Seq: 30, Event: githubDeploymentEvent(30)}}}
	if _, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src, req, emptySnapshot(), input); err != nil {
		t.Fatalf("replay: %v", err)
	}
	after := sourceLedgerDigest(t, src)
	if before != after {
		t.Fatalf("TWIN-INV-20 violated: source ledger changed across isolated replay: %s -> %s", before, after)
	}

	// Positive control: an actual mutation MUST change the digest.
	appendDeployment(t, src, 6)
	if sourceLedgerDigest(t, src) == before {
		t.Fatal("ledger-preservation control inert: digest unchanged after a real source mutation")
	}
}

// TestRuntimeFixtureCleanupWitness is TWIN-INV-21: the owned temporary root is removed on success.
// The witness stats the path for ABSENCE (not "RemoveAll was called"); the positive control
// suppresses removal and proves the absence check is load-bearing (the root then remains).
func TestRuntimeFixtureCleanupWitness(t *testing.T) {
	input := loadVerifiedPostgresInput(t)
	src := openVerifiedSourceStore(t, input)

	req := Request{ThroughSeq: 0, Candidate: []Candidate{{Seq: 30, Event: githubDeploymentEvent(30)}}}
	result, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src, req, emptySnapshot(), input)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if result.ownedRoot == "" {
		t.Fatal("owned root not surfaced")
	}
	if _, statErr := os.Stat(result.ownedRoot); !os.IsNotExist(statErr) {
		t.Fatalf("TWIN-INV-21: owned root still present after success: stat err=%v", statErr)
	}

	// Positive control: suppress removal; the owned root MUST then remain, proving the absence
	// check can fail.
	saved := removeTempTwinRoot
	removeTempTwinRoot = func(string) error { return nil }
	defer func() { removeTempTwinRoot = saved }()
	leaked, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src,
		Request{ThroughSeq: 0, Candidate: []Candidate{{Seq: 31, Event: githubDeploymentEvent(31)}}},
		emptySnapshot(), input)
	if err != nil {
		t.Fatalf("control replay: %v", err)
	}
	if _, statErr := os.Stat(leaked.ownedRoot); statErr != nil {
		t.Fatalf("cleanup control inert: owned root absent despite suppressed removal: %v", statErr)
	}
	_ = os.RemoveAll(leaked.ownedRoot) // leave no residue from the intentional leak
}

// TestRuntimeFixtureNoFetchDeclaration is TWIN-INV-19 in its LABELED DECLARATION mode (per
// VD-p7-nofetch-witness-2026-11-04: a no-egress network namespace is the primary witness where it
// is wired; this build uses the declaration fallback, recorded honestly as a configuration
// assertion, not a namespace observation). The declaration: provisioning uses only the
// caller-supplied archive bytes via BinariesDir; the staged tree contains bin/pg_ctl so
// embedded-postgres short-circuits its remote fetch, and store.Open constructs no remote fetch
// strategy — so no egress is attempted for binaries. This test asserts the fetch-short-circuit
// condition and that the store opens offline from exactly those bytes.
func TestRuntimeFixtureNoFetchDeclaration(t *testing.T) {
	input := loadVerifiedPostgresInput(t)
	binDir, err := stageVerifiedPostgres(context.Background(), t.TempDir(), input)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	if _, err := os.Stat(filepath.Join(binDir, "bin", "pg_ctl")); err != nil {
		t.Fatalf("no-fetch declaration: staged bin/pg_ctl absent (a missing pg_ctl is what triggers a fetch): %v", err)
	}
	port, err := temporaryPort()
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(context.Background(), store.Config{Root: t.TempDir(), Port: port, BinariesDir: binDir})
	if err != nil {
		t.Fatalf("offline open from staged bytes: %v", err)
	}
	_ = s.Close()
}
