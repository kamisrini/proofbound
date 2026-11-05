package twin

// Unit-3 §6/§9.3 — author a proofbound.p7.behavior-lock.v1 over the REAL replay proof so
// EvaluateLock == LOCKED (TWIN-INV-22, a determinism claim — NOT a runtime witness; the runtime
// fact is carried by TWIN-INV-18). Two tests:
//
//   - TestEmitRealBehaviorLock (needs the archive AND PROOFBOUND_P7_EMIT_DIR): runs a real replay,
//     authors the lock from the real proof values, proves EvaluateLock == LOCKED with a resolver
//     that GENUINELY resolves the committed requirement record, and writes the committed artifacts.
//   - TestCommittedBehaviorLockIsLocked (no archive — runs in make check): re-parses the COMMITTED
//     lock + proof bytes and asserts EvaluateLock == LOCKED deterministically. This is the
//     TWIN-INV-22 determinism gate.

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

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	"github.com/kamisrini/proofbound/kernel/internal/twin/manifest"
)

const (
	committedLockPath  = "../../../docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-behavior-lock.json"
	committedProofPath = "../../../docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-replay-proof.json"
	replayRequirementID = "BR-replay-evidence-6e3d2b"
	replayObligationID  = "O-1"
)

// committedRequirementResolver resolves a declared P5 requirement obligation against the committed
// intent record: it reads docs/intent/records/requirements/<id>/0001.md, recomputes the record's
// self-digest and requires it to equal the declared artifact_sha256, and requires the obligation to
// be present and active. A missing, stale (digest mismatch), or conflicting record fails closed —
// exactly BR-replay-evidence-6e3d2b O-1's "missing/stale/conflicting evidence must prevent the
// verified claim".
type committedRequirementResolver struct{ repoRoot string }

func (r committedRequirementResolver) ResolveRequirement(recordID, artifactSHA256, obligationID string) error {
	path := filepath.Join(r.repoRoot, "docs", "intent", "records", "requirements", recordID, "0001.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if got := recordSelfDigest(raw); got != artifactSHA256 {
		return &resolveError{"requirement self-digest mismatch: got " + got + " want " + artifactSHA256}
	}
	var rec struct {
		RequirementID string `json:"requirement_id"`
		Obligations   []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"obligations"`
	}
	if err := json.Unmarshal(jsonLine(raw), &rec); err != nil {
		return err
	}
	if rec.RequirementID != recordID {
		return &resolveError{"requirement_id mismatch"}
	}
	for _, o := range rec.Obligations {
		if o.ID == obligationID {
			if o.State != "active" {
				return &resolveError{"obligation " + obligationID + " not active"}
			}
			return nil
		}
	}
	return &resolveError{"obligation " + obligationID + " not found"}
}

type resolveError struct{ msg string }

func (e *resolveError) Error() string { return "resolve requirement: " + e.msg }

// jsonLine returns the proofbound-json object line (line 2) of a record file.
func jsonLine(raw []byte) []byte {
	lines := strings.SplitN(string(raw), "\n", 3)
	if len(lines) >= 2 {
		return []byte(lines[1])
	}
	return nil
}

// recordSelfDigest reproduces the record self-digest: zero the (unique) artifact_sha256 field, then
// sha256 the whole file. Mirrors scripts/p6-artifact-integrity.sh's normalization.
func recordSelfDigest(raw []byte) string {
	marker := []byte(`"artifact_sha256":"`)
	wi := strings.Index(string(raw), string(marker))
	if wi < 0 || wi+len(marker)+64 > len(raw) {
		return ""
	}
	ds := wi + len(marker)
	zeroed := make([]byte, 0, len(raw))
	zeroed = append(zeroed, raw[:ds]...)
	zeroed = append(zeroed, []byte(strings.Repeat("0", 64))...)
	zeroed = append(zeroed, raw[ds+64:]...)
	sum := sha256.Sum256(zeroed)
	return hex.EncodeToString(sum[:])
}

// requirementArtifactSHA256 reads the committed requirement record and returns its self-digest, so
// the lock binds the real digest without hardcoding it (Law 2: the datum's home is the record).
func requirementArtifactSHA256(t *testing.T, repoRoot, recordID string) string {
	t.Helper()
	path := filepath.Join(repoRoot, "docs", "intent", "records", "requirements", recordID, "0001.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read requirement record: %v", err)
	}
	return recordSelfDigest(raw)
}

func candidateBytesSHA256(ev core.Event) string {
	sum := sha256.Sum256(ev.Payload)
	return hex.EncodeToString(sum[:])
}

func manifestCandidateDigest(cands []manifest.LockCandidate) string {
	hexes := make([]string, 0, len(cands))
	for _, c := range cands {
		hexes = append(hexes, c.BytesSHA256)
	}
	sum := sha256.Sum256([]byte(strings.Join(hexes, "\n")))
	return hex.EncodeToString(sum[:])
}

// authorBehaviorLock builds the lock from the REAL replay result + the real source prefix and the
// authored candidate inputs. Source/baseline/candidate-snapshot digests + through_seq come from the
// real proof; the candidate digest is the manifest-scheme digest over the authored candidate byte
// identities (the lock is its own artifact — see the unit-3 spec and the accepted manifest
// contract).
func authorBehaviorLock(t *testing.T, result Result, src *store.Store, cands []Candidate, projRev, projRevDigest, requirementSHA string) manifest.BehaviorLock {
	t.Helper()
	var events []manifest.SourceEvent
	if err := src.ReadEvents(context.Background(), store.Filter{}, func(r store.Record) error {
		if r.Seq <= result.ThroughSeq {
			events = append(events, manifest.SourceEvent{EventID: r.Event.ID.String(), Seq: r.Seq})
		}
		return nil
	}); err != nil {
		t.Fatalf("read source prefix: %v", err)
	}
	lockCands := make([]manifest.LockCandidate, 0, len(cands))
	seqMap := make([]manifest.SequenceEntry, 0, len(cands))
	for _, c := range cands {
		lockCands = append(lockCands, manifest.LockCandidate{EventID: c.Event.ID.String(), Seq: c.Seq, BytesSHA256: candidateBytesSHA256(c.Event)})
		seqMap = append(seqMap, manifest.SequenceEntry{EventID: c.Event.ID.String(), Requested: c.Seq, Effective: c.Seq})
	}
	return manifest.BehaviorLock{
		Subject:      manifest.ReplayIsolatedSubject,
		LockRevision: manifest.LockRevision{Subject: manifest.ReplayIsolatedSubject, Number: 1, Parent: "genesis"},
		SourcePrefix: manifest.SourcePrefix{ThroughSeq: result.ThroughSeq, Events: events, SourceDigest: result.Proof.SourceDigest},
		Candidates:   lockCands,
		Baseline: manifest.LockBaseline{
			ThroughSeq:              result.ThroughSeq,
			SourceDigest:            result.Proof.SourceDigest,
			ProjectorRevision:       projRev,
			ProjectorRevisionDigest: projRevDigest,
			BaselineSnapshotDigest:  result.Proof.BaselineSnapshotDigest,
		},
		CandidateSnapshot: manifest.LockSnapshot{SnapshotDigest: result.Proof.CandidateSnapshotDigest},
		Verdict:           string(result.Verdict),
		SequenceMap:       seqMap,
		ReplayProof: manifest.ReplayProof{
			Schema:                  manifest.ReplayProofSchema,
			ThroughSeq:              result.ThroughSeq,
			SourceDigest:            result.Proof.SourceDigest,
			CandidateDigest:         manifestCandidateDigest(lockCands),
			BaselineSnapshotDigest:  result.Proof.BaselineSnapshotDigest,
			CandidateSnapshotDigest: result.Proof.CandidateSnapshotDigest,
		},
		IntentBinding: manifest.IntentBinding{
			RecordKind:     "requirement",
			Source:         "intent.records",
			RecordID:       replayRequirementID,
			ArtifactSHA256: requirementSHA,
			ObligationID:   replayObligationID,
		},
	}
}

func projectorRevision(t *testing.T, repoRoot string) (rev, digest string) {
	t.Helper()
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "HEAD:kernel/internal/projections").Output()
	if err != nil {
		t.Fatalf("projector tree rev: %v", err)
	}
	rev = strings.TrimSpace(string(out))
	tree, err := exec.Command("git", "-C", repoRoot, "ls-tree", "-r", "HEAD", "--", "kernel/internal/projections").Output()
	if err != nil {
		t.Fatalf("projector tree listing: %v", err)
	}
	sum := sha256.Sum256(tree)
	return rev, hex.EncodeToString(sum[:])
}

func TestEmitRealBehaviorLock(t *testing.T) {
	emitDir := os.Getenv("PROOFBOUND_P7_EMIT_DIR")
	if emitDir == "" {
		t.Skip("PROOFBOUND_P7_EMIT_DIR unset; emit path not exercised")
	}
	input := loadVerifiedPostgresInput(t)
	src := openVerifiedSourceStore(t, input)
	appendDeployment(t, src, 1)
	appendDeployment(t, src, 2)

	cands := []Candidate{{Seq: 3, Event: githubDeploymentEvent(3)}}
	result, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), src, Request{ThroughSeq: 2, Candidate: cands}, emptySnapshot(), input)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if result.launch.resolvedSHA256 != input.PostgresBinarySHA256 {
		t.Fatalf("launched-process witness did not verify: %s", result.launch.resolvedSHA256)
	}
	repoRoot := "../../.."
	projRev, projDigest := projectorRevision(t, repoRoot)
	reqSHA := requirementArtifactSHA256(t, repoRoot, replayRequirementID)
	lock := authorBehaviorLock(t, result, src, cands, projRev, projDigest, reqSHA)

	resolver := committedRequirementResolver{repoRoot: repoRoot}
	eval, err := manifest.EvaluateLock(lock, mustMarshalProof(t, lock), resolver)
	if err != nil {
		t.Fatalf("EvaluateLock: %v", err)
	}
	if eval.Overall != manifest.LockLocked {
		t.Fatalf("authored lock not LOCKED: %+v", eval)
	}

	lockBytes, err := manifest.CanonicalBehaviorLockBytes(lock)
	if err != nil {
		t.Fatalf("canonical lock: %v", err)
	}
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-behavior-lock.json"), lockBytes)
	writeEmit(t, filepath.Join(emitDir, "p7-b1-unit3-replay-proof.json"), mustMarshalProof(t, lock))
	t.Logf("emitted LOCKED behavior-lock (%d bytes) + replay proof to %s", len(lockBytes), emitDir)
}

func mustMarshalProof(t *testing.T, lock manifest.BehaviorLock) []byte {
	t.Helper()
	raw, err := json.Marshal(lock.ReplayProof)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func writeEmit(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestCommittedBehaviorLockIsLocked re-validates the COMMITTED lock deterministically (no archive,
// runs in make check): EvaluateLock over the committed proof bytes must be LOCKED (TWIN-INV-22).
func TestCommittedBehaviorLockIsLocked(t *testing.T) {
	lockRaw, err := os.ReadFile(filepath.Clean(committedLockPath))
	if err != nil {
		t.Fatalf("read committed lock: %v", err)
	}
	proofRaw, err := os.ReadFile(filepath.Clean(committedProofPath))
	if err != nil {
		t.Fatalf("read committed proof: %v", err)
	}
	lock, err := manifest.ParseBehaviorLock(lockRaw)
	if err != nil {
		t.Fatalf("parse committed lock: %v", err)
	}
	resolver := committedRequirementResolver{repoRoot: "../../.."}
	eval, err := manifest.EvaluateLock(lock, proofRaw, resolver)
	if err != nil {
		t.Fatalf("EvaluateLock: %v", err)
	}
	if eval.Overall != manifest.LockLocked {
		t.Fatalf("committed lock not LOCKED: %+v", eval)
	}
	// Negative control: a resolver that fails MUST drop it below LOCKED (intent gate has teeth).
	bad, err := manifest.EvaluateLock(lock, proofRaw, committedRequirementResolver{repoRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("EvaluateLock (bad resolver): %v", err)
	}
	if bad.Overall == manifest.LockLocked {
		t.Fatal("intent gate inert: LOCKED even when the requirement record cannot be resolved")
	}
}
