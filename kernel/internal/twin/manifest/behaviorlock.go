package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// BehaviorLockSchema identifies the authored ReplayIsolated behavior-lock contract.
// The lock carries no top-level "schema" field; the schema name identifies the artifact family.
const BehaviorLockSchema = "proofbound.p7.behavior-lock.v1"

// ReplayProofSchema is the frozen replay-proof wire identity the lock binds against.
const ReplayProofSchema = "vera.replay.v1"

// genesisParent is the sentinel parent of a first lock revision.
const genesisParent = "genesis"

// Behavior-lock evaluation criterion and overall states.
const (
	LockVerified   = "VERIFIED"
	LockUnverified = "UNVERIFIED"
	LockLocked     = "LOCKED"
)

var (
	eventIDRE       = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
	intentSourceRE  = regexp.MustCompile(`^intent\.[a-z0-9][a-z0-9_-]*$`)
	intentRecordRE  = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
	obligationIDRE  = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
	lockVerdictEnum = map[string]struct{}{"preserved": {}, "changed": {}}
)

// BehaviorLock is the authored, content-addressed ReplayIsolated behavior-lock artifact.
// Validating it and cross-binding it to supplied committed proof bytes are non-runtime operations.
type BehaviorLock struct {
	Subject           string          `json:"subject"`
	LockRevision      LockRevision    `json:"lock_revision"`
	SourcePrefix      SourcePrefix    `json:"source_prefix"`
	Candidates        []LockCandidate `json:"candidates"`
	Baseline          LockBaseline    `json:"baseline"`
	CandidateSnapshot LockSnapshot    `json:"candidate_snapshot"`
	Verdict           string          `json:"verdict"`
	SequenceMap       []SequenceEntry `json:"sequence_map"`
	ReplayProof       ReplayProof     `json:"replay_proof"`
	IntentBinding     IntentBinding   `json:"intent_binding"`
}

// LockRevision is the lock's identity and ancestry.
type LockRevision struct {
	Subject string `json:"subject"`
	Number  int64  `json:"number"`
	Parent  string `json:"parent"`
}

// SourcePrefix binds the observed append-only ledger prefix.
type SourcePrefix struct {
	ThroughSeq   int64         `json:"through_seq"`
	Events       []SourceEvent `json:"events"`
	SourceDigest string        `json:"source_digest"`
}

// SourceEvent is one observed source event identity and sequence number.
type SourceEvent struct {
	EventID string `json:"event_id"`
	Seq     int64  `json:"seq"`
}

// LockCandidate is one authored candidate event, declared as candidate input, not observation.
type LockCandidate struct {
	EventID     string `json:"event_id"`
	Seq         int64  `json:"seq"`
	BytesSHA256 string `json:"bytes_sha256"`
}

// LockBaseline is the full baseline provenance tuple.
type LockBaseline struct {
	ThroughSeq              int64  `json:"through_seq"`
	SourceDigest            string `json:"source_digest"`
	ProjectorRevision       string `json:"projector_revision"`
	ProjectorRevisionDigest string `json:"projector_revision_digest"`
	BaselineSnapshotDigest  string `json:"baseline_snapshot_digest"`
}

// LockSnapshot is the candidate snapshot digest.
type LockSnapshot struct {
	SnapshotDigest string `json:"snapshot_digest"`
}

// SequenceEntry maps a candidate's requested sequence to its effective sequence.
type SequenceEntry struct {
	EventID   string `json:"event_id"`
	Requested int64  `json:"requested"`
	Effective int64  `json:"effective"`
}

// ReplayProof mirrors the frozen vera.replay.v1 proof identity the lock protects.
type ReplayProof struct {
	Schema                  string `json:"schema"`
	ThroughSeq              int64  `json:"through_seq"`
	SourceDigest            string `json:"source_digest"`
	CandidateDigest         string `json:"candidate_digest"`
	BaselineSnapshotDigest  string `json:"baseline_snapshot_digest"`
	CandidateSnapshotDigest string `json:"candidate_snapshot_digest"`
}

// IntentBinding is either the explicit UNDECLARED state or an exact requirement tuple.
type IntentBinding struct {
	State          string `json:"state,omitempty"`
	RecordKind     string `json:"record_kind,omitempty"`
	Source         string `json:"source,omitempty"`
	RecordID       string `json:"record_id,omitempty"`
	ArtifactSHA256 string `json:"artifact_sha256,omitempty"`
	ObligationID   string `json:"obligation_id,omitempty"`
}

// declared reports whether the intent binding is an exact requirement tuple.
func (b IntentBinding) declared() bool {
	return b.State == "" && b.RecordKind != "" && b.Source != "" &&
		b.RecordID != "" && b.ArtifactSHA256 != "" && b.ObligationID != ""
}

// undeclared reports whether the intent binding is the explicit UNDECLARED state.
func (b IntentBinding) undeclared() bool {
	return b.State == string(CaseStatusUndeclared) && b.RecordKind == "" && b.Source == "" &&
		b.RecordID == "" && b.ArtifactSHA256 == "" && b.ObligationID == ""
}

// RequirementResolver resolves a declared P5 requirement obligation to its exact revision.
// A nil return means the exact record, artifact digest, and obligation resolve; any error fails closed.
type RequirementResolver interface {
	ResolveRequirement(recordID, artifactSHA256, obligationID string) error
}

// LockEvaluation is the derived, non-runtime evaluation of a lock against supplied proof bytes.
type LockEvaluation struct {
	Identity        string
	SourceCandidate string
	Baseline        string
	Proof           string
	Intent          string
	Overall         string
}

// ParseBehaviorLock validates strict canonical JSON and returns the typed behavior lock.
func ParseBehaviorLock(raw []byte) (BehaviorLock, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return BehaviorLock{}, fmt.Errorf("behavior-lock: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return BehaviorLock{}, errors.New("behavior-lock: JSON is not canonical")
	}
	if err := validateLockShape(raw); err != nil {
		return BehaviorLock{}, err
	}
	var lock BehaviorLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		return BehaviorLock{}, fmt.Errorf("behavior-lock: decode: %w", err)
	}
	if err := ValidateBehaviorLock(lock); err != nil {
		return BehaviorLock{}, err
	}
	return lock, nil
}

// CanonicalBehaviorLockBytes validates and encodes a behavior lock as RFC 8785 canonical JSON.
func CanonicalBehaviorLockBytes(lock BehaviorLock) ([]byte, error) {
	if err := ValidateBehaviorLock(lock); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		return nil, fmt.Errorf("behavior-lock: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("behavior-lock: canonical encode: %w", err)
	}
	return canonical, nil
}

// BehaviorLockDigest returns the SHA-256 of the validated canonical lock bytes. It is the lock identity.
func BehaviorLockDigest(raw []byte) (string, error) {
	if _, err := ParseBehaviorLock(raw); err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// ValidateBehaviorLock validates a typed lock, including internal proof/baseline consistency.
func ValidateBehaviorLock(lock BehaviorLock) error {
	if lock.Subject != ReplayIsolatedSubject {
		return fmt.Errorf("behavior-lock: subject must be %q", ReplayIsolatedSubject)
	}
	if err := validateLockRevision(lock.LockRevision); err != nil {
		return err
	}
	if err := validateSourcePrefix(lock.SourcePrefix); err != nil {
		return err
	}
	candidateIDs, err := validateLockCandidates(lock.Candidates)
	if err != nil {
		return err
	}
	if err := validateLockBaseline(lock.Baseline, lock.SourcePrefix); err != nil {
		return err
	}
	if err := validateSHA256("candidate_snapshot snapshot_digest", lock.CandidateSnapshot.SnapshotDigest); err != nil {
		return err
	}
	if _, ok := lockVerdictEnum[lock.Verdict]; !ok {
		return fmt.Errorf("behavior-lock: verdict must be preserved or changed, got %q", lock.Verdict)
	}
	if err := validateSequenceMap(lock.SequenceMap, candidateIDs); err != nil {
		return err
	}
	if err := validateIntentBinding(lock.IntentBinding); err != nil {
		return err
	}
	return validateReplayProofConsistency(lock)
}

func validateLockRevision(revision LockRevision) error {
	if revision.Subject != ReplayIsolatedSubject {
		return fmt.Errorf("behavior-lock: lock_revision subject must be %q", ReplayIsolatedSubject)
	}
	if revision.Number < 1 {
		return fmt.Errorf("behavior-lock: lock_revision number must be positive, got %d", revision.Number)
	}
	if revision.Parent != genesisParent && !sha256RE.MatchString(revision.Parent) {
		return fmt.Errorf("behavior-lock: lock_revision parent must be %q or a 64-hex digest", genesisParent)
	}
	return nil
}

func validateSourcePrefix(prefix SourcePrefix) error {
	if prefix.ThroughSeq < 1 {
		return fmt.Errorf("behavior-lock: source_prefix through_seq must be positive, got %d", prefix.ThroughSeq)
	}
	if len(prefix.Events) == 0 {
		return errors.New("behavior-lock: source_prefix events must be non-empty")
	}
	seenID := make(map[string]struct{}, len(prefix.Events))
	var prevSeq int64
	for i, event := range prefix.Events {
		if !eventIDRE.MatchString(event.EventID) {
			return fmt.Errorf("behavior-lock: source event %d has invalid event_id %q", i, event.EventID)
		}
		if _, dup := seenID[event.EventID]; dup {
			return fmt.Errorf("behavior-lock: duplicate source event_id %q", event.EventID)
		}
		seenID[event.EventID] = struct{}{}
		if event.Seq < 1 || event.Seq > prefix.ThroughSeq {
			return fmt.Errorf("behavior-lock: source event %q seq %d out of range [1,%d]", event.EventID, event.Seq, prefix.ThroughSeq)
		}
		if event.Seq <= prevSeq {
			return fmt.Errorf("behavior-lock: source events must be ascending and unique by seq at %q", event.EventID)
		}
		prevSeq = event.Seq
	}
	return validateSHA256("source_prefix source_digest", prefix.SourceDigest)
}

func validateLockCandidates(candidates []LockCandidate) ([]string, error) {
	if len(candidates) == 0 {
		return nil, errors.New("behavior-lock: candidates must be non-empty")
	}
	ids := make([]string, 0, len(candidates))
	seenID := make(map[string]struct{}, len(candidates))
	seenSeq := make(map[int64]struct{}, len(candidates))
	var prevSeq int64
	for i, candidate := range candidates {
		if !eventIDRE.MatchString(candidate.EventID) {
			return nil, fmt.Errorf("behavior-lock: candidate %d has invalid event_id %q", i, candidate.EventID)
		}
		if _, dup := seenID[candidate.EventID]; dup {
			return nil, fmt.Errorf("behavior-lock: duplicate candidate event_id %q", candidate.EventID)
		}
		seenID[candidate.EventID] = struct{}{}
		if candidate.Seq < 1 {
			return nil, fmt.Errorf("behavior-lock: candidate %q seq must be positive", candidate.EventID)
		}
		if _, dup := seenSeq[candidate.Seq]; dup {
			return nil, fmt.Errorf("behavior-lock: duplicate candidate seq %d", candidate.Seq)
		}
		seenSeq[candidate.Seq] = struct{}{}
		if candidate.Seq <= prevSeq {
			return nil, fmt.Errorf("behavior-lock: candidates must be ascending by seq at %q", candidate.EventID)
		}
		prevSeq = candidate.Seq
		if err := validateSHA256("candidate "+candidate.EventID+" bytes_sha256", candidate.BytesSHA256); err != nil {
			return nil, err
		}
		ids = append(ids, candidate.EventID)
	}
	return ids, nil
}

func validateLockBaseline(baseline LockBaseline, prefix SourcePrefix) error {
	if baseline.ThroughSeq != prefix.ThroughSeq {
		return fmt.Errorf("behavior-lock: baseline through_seq %d must equal source_prefix %d", baseline.ThroughSeq, prefix.ThroughSeq)
	}
	if baseline.SourceDigest != prefix.SourceDigest {
		return errors.New("behavior-lock: baseline source_digest must equal source_prefix source_digest")
	}
	if err := validateRevision("baseline projector_revision", baseline.ProjectorRevision); err != nil {
		return err
	}
	if err := validateSHA256("baseline projector_revision_digest", baseline.ProjectorRevisionDigest); err != nil {
		return err
	}
	return validateSHA256("baseline baseline_snapshot_digest", baseline.BaselineSnapshotDigest)
}

func validateSequenceMap(entries []SequenceEntry, candidateIDs []string) error {
	if len(entries) != len(candidateIDs) {
		return fmt.Errorf("behavior-lock: sequence_map must cover exactly %d candidates, got %d", len(candidateIDs), len(entries))
	}
	want := make(map[string]struct{}, len(candidateIDs))
	for _, id := range candidateIDs {
		want[id] = struct{}{}
	}
	seen := make(map[string]struct{}, len(entries))
	var prevRequested int64
	for i, entry := range entries {
		if _, ok := want[entry.EventID]; !ok {
			return fmt.Errorf("behavior-lock: sequence_map entry %q is not a candidate", entry.EventID)
		}
		if _, dup := seen[entry.EventID]; dup {
			return fmt.Errorf("behavior-lock: duplicate sequence_map entry %q", entry.EventID)
		}
		seen[entry.EventID] = struct{}{}
		if entry.Requested < 1 || entry.Effective < 1 {
			return fmt.Errorf("behavior-lock: sequence_map %q requires positive requested and effective", entry.EventID)
		}
		if i > 0 && entry.Requested <= prevRequested {
			return fmt.Errorf("behavior-lock: sequence_map must be ascending by requested at %q", entry.EventID)
		}
		prevRequested = entry.Requested
	}
	return nil
}

func validateIntentBinding(binding IntentBinding) error {
	switch {
	case binding.undeclared():
		return nil
	case binding.declared():
		if binding.RecordKind != "requirement" {
			return fmt.Errorf("behavior-lock: intent_binding record_kind must be %q", "requirement")
		}
		if !intentSourceRE.MatchString(binding.Source) {
			return fmt.Errorf("behavior-lock: intent_binding source %q must be intent.<provider>", binding.Source)
		}
		if !intentRecordRE.MatchString(binding.RecordID) {
			return fmt.Errorf("behavior-lock: intent_binding record_id %q is invalid", binding.RecordID)
		}
		if err := validateSHA256("intent_binding artifact_sha256", binding.ArtifactSHA256); err != nil {
			return err
		}
		if !obligationIDRE.MatchString(binding.ObligationID) {
			return fmt.Errorf("behavior-lock: intent_binding obligation_id %q is invalid", binding.ObligationID)
		}
		return nil
	default:
		return errors.New("behavior-lock: intent_binding must be exactly UNDECLARED or a complete requirement tuple")
	}
}

// validateReplayProofConsistency requires the lock's declared replay proof to agree with its own
// source, candidate, baseline, and candidate-snapshot members.
func validateReplayProofConsistency(lock BehaviorLock) error {
	proof := lock.ReplayProof
	if proof.Schema != ReplayProofSchema {
		return fmt.Errorf("behavior-lock: replay_proof schema must be %q", ReplayProofSchema)
	}
	if proof.ThroughSeq != lock.SourcePrefix.ThroughSeq {
		return errors.New("behavior-lock: replay_proof through_seq must equal source_prefix")
	}
	if proof.SourceDigest != lock.SourcePrefix.SourceDigest {
		return errors.New("behavior-lock: replay_proof source_digest must equal source_prefix")
	}
	if proof.BaselineSnapshotDigest != lock.Baseline.BaselineSnapshotDigest {
		return errors.New("behavior-lock: replay_proof baseline_snapshot_digest must equal baseline")
	}
	if proof.CandidateSnapshotDigest != lock.CandidateSnapshot.SnapshotDigest {
		return errors.New("behavior-lock: replay_proof candidate_snapshot_digest must equal candidate_snapshot")
	}
	if proof.CandidateDigest != candidateDigest(lock.Candidates) {
		return errors.New("behavior-lock: replay_proof candidate_digest must equal the derived candidate digest")
	}
	return nil
}

// candidateDigest is the SHA-256 of the newline-joined ordered candidate bytes_sha256 values.
func candidateDigest(candidates []LockCandidate) string {
	hexes := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		hexes = append(hexes, candidate.BytesSHA256)
	}
	sum := sha256.Sum256([]byte(strings.Join(hexes, "\n")))
	return hex.EncodeToString(sum[:])
}

// EvaluateLock derives the non-runtime lock evaluation over supplied committed proof bytes.
// It never runs a replay; LOCKED requires retrievable matching proof bytes and a resolved intent.
func EvaluateLock(lock BehaviorLock, proofBytes []byte, resolver RequirementResolver) (LockEvaluation, error) {
	if err := ValidateBehaviorLock(lock); err != nil {
		return LockEvaluation{}, err
	}
	evaluation := LockEvaluation{
		Identity:        LockVerified,
		SourceCandidate: LockVerified,
		Baseline:        LockVerified,
		Proof:           evaluateProof(lock, proofBytes),
		Intent:          evaluateIntent(lock.IntentBinding, resolver),
	}
	if evaluation.Proof == LockVerified && evaluation.Intent == LockVerified {
		evaluation.Overall = LockLocked
	} else {
		evaluation.Overall = LockUnverified
	}
	return evaluation, nil
}

// evaluateProof verifies the supplied bytes parse to vera.replay.v1 and equal the lock's proof.
func evaluateProof(lock BehaviorLock, proofBytes []byte) string {
	if len(proofBytes) == 0 {
		return LockUnverified
	}
	var proof ReplayProof
	decoder := json.NewDecoder(bytes.NewReader(proofBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil {
		return LockUnverified
	}
	if proof.Schema != ReplayProofSchema {
		return LockUnverified
	}
	if proof != lock.ReplayProof {
		return LockUnverified
	}
	return LockVerified
}

// evaluateIntent resolves a declared requirement; UNDECLARED is never passed to the resolver.
func evaluateIntent(binding IntentBinding, resolver RequirementResolver) string {
	if binding.undeclared() {
		return LockUnverified
	}
	if resolver == nil {
		return LockUnverified
	}
	if err := resolver.ResolveRequirement(binding.RecordID, binding.ArtifactSHA256, binding.ObligationID); err != nil {
		return LockUnverified
	}
	return LockVerified
}

// validateLockShape rejects unknown, missing, null, or non-object members at every level.
func validateLockShape(raw json.RawMessage) error {
	top, err := exactObject(raw, "subject", "lock_revision", "source_prefix", "candidates",
		"baseline", "candidate_snapshot", "verdict", "sequence_map", "replay_proof", "intent_binding")
	if err != nil {
		return err
	}
	if _, rerr := exactObject(top["lock_revision"], "subject", "number", "parent"); rerr != nil {
		return rerr
	}
	prefix, perr := exactObject(top["source_prefix"], "through_seq", "events", "source_digest")
	if perr != nil {
		return perr
	}
	if eerr := eachObject(prefix["events"], func(entry json.RawMessage) error {
		_, serr := exactObject(entry, "event_id", "seq")
		return serr
	}); eerr != nil {
		return eerr
	}
	if cerr := eachObject(top["candidates"], func(entry json.RawMessage) error {
		_, serr := exactObject(entry, "event_id", "seq", "bytes_sha256")
		return serr
	}); cerr != nil {
		return cerr
	}
	if _, berr := exactObject(top["baseline"], "through_seq", "source_digest", "projector_revision",
		"projector_revision_digest", "baseline_snapshot_digest"); berr != nil {
		return berr
	}
	if _, serr := exactObject(top["candidate_snapshot"], "snapshot_digest"); serr != nil {
		return serr
	}
	if merr := eachObject(top["sequence_map"], func(entry json.RawMessage) error {
		_, serr := exactObject(entry, "event_id", "requested", "effective")
		return serr
	}); merr != nil {
		return merr
	}
	if _, rerr := exactObject(top["replay_proof"], "schema", "through_seq", "source_digest",
		"candidate_digest", "baseline_snapshot_digest", "candidate_snapshot_digest"); rerr != nil {
		return rerr
	}
	return validateIntentShape(top["intent_binding"])
}

// validateIntentShape accepts exactly the UNDECLARED object or the full requirement tuple.
func validateIntentShape(raw json.RawMessage) error {
	if _, err := exactObject(raw, "state"); err == nil {
		return nil
	}
	if _, err := exactObject(raw, "record_kind", "source", "record_id", "artifact_sha256", "obligation_id"); err == nil {
		return nil
	}
	return errors.New("behavior-lock: intent_binding must be exactly {state} or a requirement tuple")
}
