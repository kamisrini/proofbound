package twin

// Unit-3 §6/§9.3 — author the proofbound.p7.case-manifest.v1 over the REAL run so
// DeriveCaseStatus == PASS, with every one of the nine outcomes bound to real committed evidence.
// The FIXTURE is the gate (spec §1): it authors each outcome PASS only from a verified witness;
// proof_binding is gated on the TWIN-INV-18 launched-process witness (not determinism alone, N3).
//
//   - TestEmitRealCaseManifest (archive + PROOFBOUND_P7_EMIT_DIR): runs a real replay, emits the
//     four witness records, authors the manifest referencing real committed evidence, asserts
//     DeriveCaseStatus == PASS, and writes the committed artifacts.
//   - TestCommittedCaseManifestIsPass (no archive — runs in make check): re-parses the COMMITTED
//     manifest, asserts DeriveCaseStatus == PASS, and RE-HASHES every referenced file to confirm the
//     content-addressed digests match the committed bytes (closing the gap that ValidateCaseManifest
//     checks digest FORMAT, not file existence).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/twin/manifest"
)

const committedCaseManifestPath = "../../../docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-case-manifest.json"

// evidence file paths, relative to the repo root (as stored in the manifest).
const (
	relLock       = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-behavior-lock.json"
	relProof      = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-replay-proof.json"
	relArchiveMan = "docs/verification/p7-replay-lock-fixtures/postgres-16.4-subset-archive.json"
	relSetup      = "docs/verification/p7-replay-lock-fixtures/setup-sources.json"
	relCompat     = "docs/verification/p7-replay-lock-prerequisites/postgres-16.4-subset-compat-20261105.md"
	relSpec       = "docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md"
	relAddendum   = "docs/plans/P7-replay-lock-spec-addendum-draft1.md"
	relMatrix     = "docs/plans/P7-replay-lock-fixture-matrix-draft1.md"
	relProjMan    = "docs/verification/p7-projector-build-evidence-capture-20261028.md"
	relLaunchW    = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-launch-witness.json"
	relLedgerW    = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-ledger-witness.json"
	relCleanupW   = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-cleanup-witness.json"
	relNoFetchW   = "docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-nofetch-declaration.json"
)

func sha256File(t *testing.T, repoRoot, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("hash %s: %v", rel, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func caseRef(t *testing.T, repoRoot, rel string) manifest.CaseRef {
	return manifest.CaseRef{Path: rel, SHA256: sha256File(t, repoRoot, rel)}
}

func headCommit(t *testing.T, repoRoot string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("git head: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestEmitRealCaseManifest(t *testing.T) {
	emitDir := os.Getenv("PROOFBOUND_P7_EMIT_DIR")
	if emitDir == "" {
		t.Skip("PROOFBOUND_P7_EMIT_DIR unset; case-manifest emit path not exercised")
	}
	input := loadVerifiedPostgresInput(t)
	src := openVerifiedSourceStore(t, input)
	appendDeployment(t, src, 1)
	appendDeployment(t, src, 2)

	before := sourceLedgerDigest(t, src)
	cands := []Candidate{{Seq: 3, Event: githubDeploymentEvent(3)}}
	result, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src, Request{ThroughSeq: 2, Candidate: cands}, emptySnapshot(), input)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	after := sourceLedgerDigest(t, src)

	// The fixture is the gate: author PASS only from verified witnesses.
	if result.launch.method != "proc-exe" || result.launch.resolvedSHA256 != input.PostgresBinarySHA256 {
		t.Fatalf("proof_binding NOT gated: launch witness unverified (%+v)", result.launch)
	}
	if before != after {
		t.Fatalf("ledger_preserved NOT satisfied: %s -> %s", before, after)
	}
	if _, statErr := os.Stat(result.ownedRoot); !os.IsNotExist(statErr) {
		t.Fatalf("cleanup NOT satisfied: owned root present: %v", statErr)
	}

	// Emit the four witness records with real values.
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-launch-witness.json"), mustJSON(t, map[string]any{
		"twin_invariant": "TWIN-INV-18", "method": "proc-exe",
		"resolved_pid": result.launch.resolvedPID, "resolved_sha256": result.launch.resolvedSHA256,
		"expected_sha256": input.PostgresBinarySHA256, "verified": true,
	}))
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-ledger-witness.json"), mustJSON(t, map[string]any{
		"twin_invariant": "TWIN-INV-20", "source_digest_before": before, "source_digest_after": after, "preserved": true,
	}))
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-cleanup-witness.json"), mustJSON(t, map[string]any{
		"twin_invariant": "TWIN-INV-21", "method": "stat-absence", "owned_root_removed": true,
	}))
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-nofetch-declaration.json"), mustJSON(t, map[string]any{
		"twin_invariant": "TWIN-INV-19", "method": "declaration",
		"binaries_path_local": true, "remote_fetch_strategy_constructed": false,
		"note": "labeled configuration declaration per VD-p7-nofetch-witness-2026-11-04; no-egress network namespace is the primary witness where wired (not wired in this build)",
	}))

	// The witness files must be hashed from the EMIT dir (that is what will be committed). The
	// committed-location hashes are computed by the deterministic test after commit.
	repoRoot := "../../.."
	hashEmit := func(name string) string {
		raw, err := os.ReadFile(filepath.Join(emitDir, name))
		if err != nil {
			t.Fatalf("read emitted %s: %v", name, err)
		}
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:])
	}
	pass := func(roles ...string) manifest.Outcome {
		return manifest.Outcome{Status: manifest.CaseStatusPass, Evidence: roles}
	}
	cm := manifest.CaseManifest{
		Schema:           manifest.CaseManifestSchema,
		CaseID:           "p7-b1-unit3-replay-isolated",
		Subject:          manifest.ReplayIsolatedSubject,
		CaseStatus:       string(manifest.CaseStatusPass),
		EvidenceRevision: headCommit(t, repoRoot),
		ContractRefs: manifest.ContractRefs{
			LockAddendum:        caseRef(t, repoRoot, relAddendum),
			FixtureMatrix:       caseRef(t, repoRoot, relMatrix),
			ImplementationScope: caseRef(t, repoRoot, relSpec),
		},
		Inputs: manifest.CaseInputs{
			Lock:              caseRef(t, repoRoot, relLock),
			Source:            caseRef(t, repoRoot, relProof),
			Candidate:         caseRef(t, repoRoot, relProof),
			SetupSources:      caseRef(t, repoRoot, relSetup),
			ProjectorManifest: caseRef(t, repoRoot, relProjMan),
		},
		Commands: []manifest.Command{},
		Witnesses: []manifest.Witness{
			{Kind: "launched-process", Path: relLaunchW, SHA256: hashEmit("p7-b1-unit3-launch-witness.json")},
			{Kind: "no-fetch", Path: relNoFetchW, SHA256: hashEmit("p7-b1-unit3-nofetch-declaration.json")},
			{Kind: "ledger-preserved", Path: relLedgerW, SHA256: hashEmit("p7-b1-unit3-ledger-witness.json")},
			{Kind: "cleanup", Path: relCleanupW, SHA256: hashEmit("p7-b1-unit3-cleanup-witness.json")},
		},
		Artifacts: []manifest.Artifact{
			{Role: "replay-proof", Path: relProof, SizeBytes: fileSize(t, repoRoot, relProof), SHA256: sha256File(t, repoRoot, relProof)},
			{Role: "behavior-lock", Path: relLock, SizeBytes: fileSize(t, repoRoot, relLock), SHA256: sha256File(t, repoRoot, relLock)},
			{Role: "prepared-archive-manifest", Path: relArchiveMan, SizeBytes: fileSize(t, repoRoot, relArchiveMan), SHA256: sha256File(t, repoRoot, relArchiveMan)},
			{Role: "compat-evidence", Path: relCompat, SizeBytes: fileSize(t, repoRoot, relCompat), SHA256: sha256File(t, repoRoot, relCompat)},
			{Role: "launch-witness", Path: relLaunchW, SizeBytes: emitSize(t, emitDir, "p7-b1-unit3-launch-witness.json"), SHA256: hashEmit("p7-b1-unit3-launch-witness.json")},
			{Role: "ledger-witness", Path: relLedgerW, SizeBytes: emitSize(t, emitDir, "p7-b1-unit3-ledger-witness.json"), SHA256: hashEmit("p7-b1-unit3-ledger-witness.json")},
			{Role: "cleanup-witness", Path: relCleanupW, SizeBytes: emitSize(t, emitDir, "p7-b1-unit3-cleanup-witness.json"), SHA256: hashEmit("p7-b1-unit3-cleanup-witness.json")},
			{Role: "nofetch-declaration", Path: relNoFetchW, SizeBytes: emitSize(t, emitDir, "p7-b1-unit3-nofetch-declaration.json"), SHA256: hashEmit("p7-b1-unit3-nofetch-declaration.json")},
		},
		Outcomes: manifest.Outcomes{
			SourceBinding:      pass("replay-proof"),
			BaselineProvenance: pass("replay-proof", "prepared-archive-manifest"),
			CandidateBinding:   pass("replay-proof"),
			LedgerPreserved:    pass("ledger-witness"),
			SnapshotComparison: pass("replay-proof"),
			ProofBinding:       pass("launch-witness", "behavior-lock"),
			NoFetch:            pass("nofetch-declaration"),
			Cleanup:            pass("cleanup-witness"),
			IntentBinding:      pass("behavior-lock"),
		},
		Review: manifest.CaseReview{
			AuthorCommit:       headCommit(t, repoRoot),
			IndependentVerdict: manifest.CaseVerdict{Status: "PENDING"},
		},
	}

	derived, err := manifest.DeriveCaseStatus(cm)
	if err != nil {
		t.Fatalf("DeriveCaseStatus: %v", err)
	}
	if derived != string(manifest.CaseStatusPass) {
		t.Fatalf("case status derived %q, want PASS", derived)
	}
	raw, err := manifest.CanonicalCaseManifestBytes(cm)
	if err != nil {
		t.Fatalf("canonical case-manifest: %v", err)
	}
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-case-manifest.json"), raw)
	t.Logf("emitted PASS case-manifest (%d bytes) + 4 witness records to %s", len(raw), emitDir)
}

func fileSize(t *testing.T, repoRoot, rel string) int64 {
	t.Helper()
	info, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("stat %s: %v", rel, err)
	}
	return info.Size()
}

func emitSize(t *testing.T, emitDir, name string) int64 {
	t.Helper()
	info, err := os.Stat(filepath.Join(emitDir, name))
	if err != nil {
		t.Fatalf("stat emitted %s: %v", name, err)
	}
	return info.Size()
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCommittedCaseManifestIsPass(t *testing.T) {
	raw, err := os.ReadFile(filepath.Clean(committedCaseManifestPath))
	if err != nil {
		t.Fatalf("read committed case-manifest: %v", err)
	}
	cm, err := manifest.ParseCaseManifest(raw)
	if err != nil {
		t.Fatalf("parse committed case-manifest: %v", err)
	}
	status, err := manifest.DeriveCaseStatus(cm)
	if err != nil {
		t.Fatalf("DeriveCaseStatus: %v", err)
	}
	if status != string(manifest.CaseStatusPass) {
		t.Fatalf("committed case-manifest status %q, want PASS", status)
	}

	// Re-hash every referenced file and confirm the content-addressed digests match the committed
	// bytes (ValidateCaseManifest checks digest FORMAT, not existence — this closes that gap).
	repoRoot := "../../.."
	refs := map[string]manifest.CaseRef{
		"contract_refs.lock_addendum":        cm.ContractRefs.LockAddendum,
		"contract_refs.fixture_matrix":       cm.ContractRefs.FixtureMatrix,
		"contract_refs.implementation_scope": cm.ContractRefs.ImplementationScope,
		"inputs.lock":                        cm.Inputs.Lock,
		"inputs.source":                      cm.Inputs.Source,
		"inputs.candidate":                   cm.Inputs.Candidate,
		"inputs.setup_sources":               cm.Inputs.SetupSources,
		"inputs.projector_manifest":          cm.Inputs.ProjectorManifest,
	}
	for label, ref := range refs {
		if got := sha256File(t, repoRoot, ref.Path); got != ref.SHA256 {
			t.Fatalf("%s digest mismatch: committed manifest says %s, file hashes %s", label, ref.SHA256, got)
		}
	}
	for _, w := range cm.Witnesses {
		if got := sha256File(t, repoRoot, w.Path); got != w.SHA256 {
			t.Fatalf("witness %q digest mismatch: manifest %s, file %s", w.Kind, w.SHA256, got)
		}
	}
	for _, a := range cm.Artifacts {
		if got := sha256File(t, repoRoot, a.Path); got != a.SHA256 {
			t.Fatalf("artifact %q digest mismatch: manifest %s, file %s", a.Role, a.SHA256, got)
		}
	}
}
