package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

const (
	sdDigest = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	bdDigest = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	cdDigest = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	pdDigest = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	cb1      = "1111111111111111111111111111111111111111111111111111111111111111"
	cb2      = "2222222222222222222222222222222222222222222222222222222222222222"
)

type stubResolver struct{ err error }

func (s stubResolver) ResolveRequirement(_, _, _ string) error { return s.err }

func lockCandidates() []LockCandidate {
	return []LockCandidate{
		{EventID: "cand-1", Seq: 1, BytesSHA256: cb1},
		{EventID: "cand-2", Seq: 2, BytesSHA256: cb2},
	}
}

func validBehaviorLock() BehaviorLock {
	cands := lockCandidates()
	return BehaviorLock{
		Subject:      ReplayIsolatedSubject,
		LockRevision: LockRevision{Subject: ReplayIsolatedSubject, Number: 1, Parent: genesisParent},
		SourcePrefix: SourcePrefix{
			ThroughSeq:   10,
			Events:       []SourceEvent{{EventID: "evt-a", Seq: 3}, {EventID: "evt-b", Seq: 7}},
			SourceDigest: sdDigest,
		},
		Candidates: cands,
		Baseline: LockBaseline{
			ThroughSeq:              10,
			SourceDigest:            sdDigest,
			ProjectorRevision:       h40,
			ProjectorRevisionDigest: pdDigest,
			BaselineSnapshotDigest:  bdDigest,
		},
		CandidateSnapshot: LockSnapshot{SnapshotDigest: cdDigest},
		Verdict:           "preserved",
		SequenceMap:       []SequenceEntry{{EventID: "cand-1", Requested: 1, Effective: 1}, {EventID: "cand-2", Requested: 2, Effective: 2}},
		ReplayProof: ReplayProof{
			Schema:                  ReplayProofSchema,
			ThroughSeq:              10,
			SourceDigest:            sdDigest,
			CandidateDigest:         candidateDigest(cands),
			BaselineSnapshotDigest:  bdDigest,
			CandidateSnapshotDigest: cdDigest,
		},
		IntentBinding: IntentBinding{State: string(CaseStatusUndeclared)},
	}
}

func declaredIntent() IntentBinding {
	return IntentBinding{
		RecordKind:     "requirement",
		Source:         "intent.records",
		RecordID:       "BR-replay-evidence-6e3d2b",
		ArtifactSHA256: h64,
		ObligationID:   "O-1",
	}
}

func marshaledProof(t *testing.T, lock BehaviorLock) []byte {
	t.Helper()
	raw, err := json.Marshal(lock.ReplayProof)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestBehaviorLockCanonicalRoundTripAndDigest(t *testing.T) {
	for _, base := range []BehaviorLock{validBehaviorLock(), func() BehaviorLock {
		l := validBehaviorLock()
		l.IntentBinding = declaredIntent()
		return l
	}()} {
		raw, err := CanonicalBehaviorLockBytes(base)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseBehaviorLock(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(parsed, base) {
			t.Fatalf("round-trip differs:\n%+v\n%+v", parsed, base)
		}
		wantSum := sha256.Sum256(raw)
		if got, err := BehaviorLockDigest(raw); err != nil || got != hex.EncodeToString(wantSum[:]) {
			t.Fatalf("digest=%s err=%v", got, err)
		}
	}
}

func TestParseBehaviorLockRejectsShapeAndCanonicality(t *testing.T) {
	base, err := CanonicalBehaviorLockBytes(validBehaviorLock())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseBehaviorLock(append(append([]byte{}, base...), '\n')); err == nil {
		t.Fatal("non-canonical JSON accepted")
	}
	shape := map[string]func(map[string]any){
		"unknown top":        func(m map[string]any) { m["extra"] = 1 },
		"missing baseline":   func(m map[string]any) { delete(m, "baseline") },
		"null candidates":    func(m map[string]any) { m["candidates"] = nil },
		"revision extra key": func(m map[string]any) { m["lock_revision"].(map[string]any)["extra"] = 1 },
		"proof extra key":    func(m map[string]any) { m["replay_proof"].(map[string]any)["extra"] = 1 },
		"intent mixed shape": func(m map[string]any) {
			m["intent_binding"] = map[string]any{"state": "UNDECLARED", "record_id": "x"}
		},
	}
	for name, mutate := range shape {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseBehaviorLock(remarshal(t, base, mutate)); err == nil {
				t.Fatal("invalid shape accepted")
			}
		})
	}
	value := map[string]func(*BehaviorLock){
		"bad subject":       func(l *BehaviorLock) { l.Subject = "Other" },
		"bad proof schema":  func(l *BehaviorLock) { l.ReplayProof.Schema = "x" },
		"bad verdict":       func(l *BehaviorLock) { l.Verdict = "maybe" },
		"intent both forms": func(l *BehaviorLock) { l.IntentBinding = declaredIntent(); l.IntentBinding.State = "UNDECLARED" },
		"intent bad source": func(l *BehaviorLock) { l.IntentBinding = declaredIntent(); l.IntentBinding.Source = "github" },
		"intent bad kind":   func(l *BehaviorLock) { l.IntentBinding = declaredIntent(); l.IntentBinding.RecordKind = "business" },
	}
	for name, mutate := range value {
		t.Run(name, func(t *testing.T) {
			l := validBehaviorLock()
			mutate(&l)
			if err := ValidateBehaviorLock(l); err == nil {
				t.Fatal("invalid value accepted")
			}
		})
	}
}

func TestBehaviorLockRejectsOrderAncestryAndDigestViolations(t *testing.T) {
	mutations := map[string]func(*BehaviorLock){
		"revision number zero":    func(l *BehaviorLock) { l.LockRevision.Number = 0 },
		"revision bad parent":     func(l *BehaviorLock) { l.LockRevision.Parent = "root" },
		"revision parent genesis": func(l *BehaviorLock) { l.LockRevision.Parent = genesisParent }, // control: still valid
		"source seq out of range": func(l *BehaviorLock) { l.SourcePrefix.Events[1].Seq = 11 },
		"source not ascending":    func(l *BehaviorLock) { l.SourcePrefix.Events = []SourceEvent{{"a", 7}, {"b", 3}} },
		"source dup id":           func(l *BehaviorLock) { l.SourcePrefix.Events = []SourceEvent{{"a", 3}, {"a", 7}} },
		"candidates not ascending": func(l *BehaviorLock) {
			l.Candidates = []LockCandidate{{EventID: "c1", Seq: 2, BytesSHA256: cb1}, {EventID: "c2", Seq: 1, BytesSHA256: cb2}}
		},
		"candidate dup seq": func(l *BehaviorLock) {
			l.Candidates = []LockCandidate{{EventID: "c1", Seq: 1, BytesSHA256: cb1}, {EventID: "c2", Seq: 1, BytesSHA256: cb2}}
		},
		"baseline through mismatch":  func(l *BehaviorLock) { l.Baseline.ThroughSeq = 9 },
		"baseline source mismatch":   func(l *BehaviorLock) { l.Baseline.SourceDigest = bdDigest },
		"seqmap wrong count":         func(l *BehaviorLock) { l.SequenceMap = l.SequenceMap[:1] },
		"seqmap unknown candidate":   func(l *BehaviorLock) { l.SequenceMap[1].EventID = "ghost" },
		"proof candidate_digest off": func(l *BehaviorLock) { l.ReplayProof.CandidateDigest = cb1 },
		"proof snapshot mismatch":    func(l *BehaviorLock) { l.ReplayProof.CandidateSnapshotDigest = bdDigest },
		"proof baseline mismatch":    func(l *BehaviorLock) { l.ReplayProof.BaselineSnapshotDigest = cdDigest },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			l := validBehaviorLock()
			mutate(&l)
			err := ValidateBehaviorLock(l)
			if name == "revision parent genesis" {
				if err != nil {
					t.Fatalf("control rejected a valid lock: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("invalid lock accepted")
			}
		})
	}
}

func TestEvaluateLockLockedOnlyWhenEveryBindingVerifies(t *testing.T) {
	lock := validBehaviorLock()
	lock.IntentBinding = declaredIntent()
	eval, err := EvaluateLock(lock, marshaledProof(t, lock), stubResolver{nil})
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []string{eval.Identity, eval.SourceCandidate, eval.Baseline, eval.Proof, eval.Intent} {
		if got != LockVerified {
			t.Fatalf("expected every criterion VERIFIED, got %+v", eval)
		}
	}
	if eval.Overall != LockLocked {
		t.Fatalf("overall=%s want LOCKED", eval.Overall)
	}
}

func TestEvaluateLockDerivesUnverifiedAndFailsClosed(t *testing.T) {
	declared := validBehaviorLock()
	declared.IntentBinding = declaredIntent()
	goodProof := marshaledProof(t, declared)

	wrongProofLock := declared
	wrongProofLock.ReplayProof.CandidateSnapshotDigest = bdDigest // internally inconsistent → validation error, so build proof bytes separately
	wrongProofBytes := marshaledProof(t, wrongProofLock)

	unknownFieldProof := []byte(`{"schema":"vera.replay.v1","through_seq":10,"source_digest":"` + sdDigest +
		`","candidate_digest":"` + candidateDigest(lockCandidates()) + `","baseline_snapshot_digest":"` + bdDigest +
		`","candidate_snapshot_digest":"` + cdDigest + `","extra":true}`)

	cases := []struct {
		name     string
		lock     BehaviorLock
		proof    []byte
		resolver RequirementResolver
		want     string
	}{
		{"empty proof", declared, nil, stubResolver{nil}, LockUnverified},
		{"mismatched proof", declared, wrongProofBytes, stubResolver{nil}, LockUnverified},
		{"unknown proof field", declared, unknownFieldProof, stubResolver{nil}, LockUnverified},
		{"undeclared intent", validBehaviorLock(), goodProof, stubResolver{nil}, LockUnverified},
		{"resolver rejects", declared, goodProof, stubResolver{errResolver}, LockUnverified},
		{"nil resolver on declared", declared, goodProof, nil, LockUnverified},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eval, err := EvaluateLock(tc.lock, tc.proof, tc.resolver)
			if err != nil {
				t.Fatal(err)
			}
			if eval.Overall != tc.want {
				t.Fatalf("overall=%s want=%s (%+v)", eval.Overall, tc.want, eval)
			}
		})
	}

	// An invalid lock is an error, not a derived result.
	bad := validBehaviorLock()
	bad.Subject = "Other"
	if _, err := EvaluateLock(bad, goodProof, stubResolver{nil}); err == nil {
		t.Fatal("invalid lock evaluated without error")
	}
}

var errResolver = &resolverError{}

type resolverError struct{}

func (*resolverError) Error() string { return "unresolved" }

// TestBehaviorLockMutationKills defends the error-check and compound-condition branches the
// calibrated mutation sweep found undefended: each case makes exactly one operand decisive.
func TestBehaviorLockMutationKills(t *testing.T) {
	// Error-check flips (!= nil -> == nil) on baseline and intent digest validation.
	errCheck := map[string]func(*BehaviorLock){
		"baseline projector_revision invalid": func(l *BehaviorLock) { l.Baseline.ProjectorRevision = "short" },
		"baseline projector_digest invalid":   func(l *BehaviorLock) { l.Baseline.ProjectorRevisionDigest = "nothex" },
		"declared intent artifact invalid": func(l *BehaviorLock) {
			l.IntentBinding = declaredIntent()
			l.IntentBinding.ArtifactSHA256 = "nothex"
		},
	}
	for name, mutate := range errCheck {
		t.Run(name, func(t *testing.T) {
			l := validBehaviorLock()
			mutate(&l)
			if err := ValidateBehaviorLock(l); err == nil {
				t.Fatal("invalid digest field accepted")
			}
		})
	}

	// sequence_map Requested<1 || Effective<1: each operand true alone must still reject,
	// which kills the || -> && mutation.
	seqMap := map[string]func(*BehaviorLock){
		"requested below one": func(l *BehaviorLock) { l.SequenceMap[0].Requested = 0 },
		"effective below one": func(l *BehaviorLock) { l.SequenceMap[0].Effective = 0 },
	}
	for name, mutate := range seqMap {
		t.Run(name, func(t *testing.T) {
			l := validBehaviorLock()
			mutate(&l)
			if err := ValidateBehaviorLock(l); err == nil {
				t.Fatal("non-positive sequence endpoint accepted")
			}
		})
	}

	// undeclared(): a State of UNDECLARED with any single record field set must NOT read as
	// undeclared (and is not declared either), so the lock is rejected. This kills the && -> ||
	// mutation on each operand; the all-empty valid case is accepted by the round-trip test.
	undeclaredBleed := map[string]func(*IntentBinding){
		"record_kind set": func(b *IntentBinding) { b.RecordKind = "requirement" },
		"source set":      func(b *IntentBinding) { b.Source = "intent.records" },
		"record_id set":   func(b *IntentBinding) { b.RecordID = "BR-x" },
		"artifact set":    func(b *IntentBinding) { b.ArtifactSHA256 = h64 },
		"obligation set":  func(b *IntentBinding) { b.ObligationID = "O-1" },
	}
	for name, mutate := range undeclaredBleed {
		t.Run("undeclared bleed: "+name, func(t *testing.T) {
			l := validBehaviorLock()
			l.IntentBinding = IntentBinding{State: string(CaseStatusUndeclared)}
			mutate(&l.IntentBinding)
			if err := ValidateBehaviorLock(l); err == nil {
				t.Fatal("UNDECLARED with a record field accepted")
			}
			if l.IntentBinding.undeclared() {
				t.Fatal("undeclared() returned true with a record field set")
			}
		})
	}
}
