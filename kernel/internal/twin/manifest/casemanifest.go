package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// CaseManifestSchema identifies the documentary case evidence index contract.
const CaseManifestSchema = "proofbound.p7.case-manifest.v1"

// ReplayIsolatedSubject is the one bounded subject these P7 non-runtime contracts describe.
const ReplayIsolatedSubject = "ReplayIsolated"

// CaseOutcomeStatus is the per-outcome evaluation vocabulary.
type CaseOutcomeStatus string

const (
	CaseStatusPass       CaseOutcomeStatus = "PASS"
	CaseStatusFail       CaseOutcomeStatus = "FAIL"
	CaseStatusUnverified CaseOutcomeStatus = "UNVERIFIED"
	CaseStatusUndeclared CaseOutcomeStatus = "UNDECLARED"
	CaseStatusPending    CaseOutcomeStatus = "PENDING"
)

var (
	caseIDRE    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	verdictRE   = regexp.MustCompile(`^(PENDING|ACCEPTED|REJECTED)$`)
	netPolicyRE = regexp.MustCompile(`^(none|loopback-only|offline)$`)
)

// outcomeKeys is the exact, ordered set of required outcome keys.
var outcomeKeys = []string{
	"source_binding", "baseline_provenance", "candidate_binding", "ledger_preserved",
	"snapshot_comparison", "proof_binding", "no_fetch", "cleanup", "intent_binding",
}

// CaseManifest is the documentary evidence index for one bounded ReplayIsolated case.
// It indexes declarations and digests; it runs nothing and approves no runtime input.
type CaseManifest struct {
	Schema           string       `json:"schema"`
	CaseID           string       `json:"case_id"`
	Subject          string       `json:"subject"`
	CaseStatus       string       `json:"case_status"`
	EvidenceRevision string       `json:"evidence_revision"`
	ContractRefs     ContractRefs `json:"contract_refs"`
	Inputs           CaseInputs   `json:"inputs"`
	Commands         []Command    `json:"commands"`
	Witnesses        []Witness    `json:"witnesses"`
	Outcomes         Outcomes     `json:"outcomes"`
	Artifacts        []Artifact   `json:"artifacts"`
	Review           CaseReview   `json:"review"`
}

// CaseRef is a content-addressed reference to a committed evidence file.
type CaseRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// ContractRefs binds the exact planning contracts this case is evaluated against.
type ContractRefs struct {
	LockAddendum        CaseRef `json:"lock_addendum"`
	FixtureMatrix       CaseRef `json:"fixture_matrix"`
	ImplementationScope CaseRef `json:"implementation_scope"`
}

// CaseInputs binds the exact case inputs by content identity.
type CaseInputs struct {
	Lock              CaseRef `json:"lock"`
	Source            CaseRef `json:"source"`
	Candidate         CaseRef `json:"candidate"`
	SetupSources      CaseRef `json:"setup_sources"`
	ProjectorManifest CaseRef `json:"projector_manifest"`
}

// Command records one recorded invocation and its content-addressed output.
type Command struct {
	ID                string   `json:"id"`
	Argv              []string `json:"argv"`
	CWD               string   `json:"cwd"`
	NetworkPolicy     string   `json:"network_policy"`
	EnvironmentDigest string   `json:"environment_digest"`
	ExitCode          int      `json:"exit_code"`
	Stdout            CaseRef  `json:"stdout"`
	Stderr            CaseRef  `json:"stderr"`
}

// Witness is one recorded observation artifact.
type Witness struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// Outcome is one evaluated case outcome and its supporting evidence roles.
type Outcome struct {
	Status   CaseOutcomeStatus `json:"status"`
	Evidence []string          `json:"evidence"`
}

// Outcomes carries the exact nine required case outcomes.
type Outcomes struct {
	SourceBinding      Outcome `json:"source_binding"`
	BaselineProvenance Outcome `json:"baseline_provenance"`
	CandidateBinding   Outcome `json:"candidate_binding"`
	LedgerPreserved    Outcome `json:"ledger_preserved"`
	SnapshotComparison Outcome `json:"snapshot_comparison"`
	ProofBinding       Outcome `json:"proof_binding"`
	NoFetch            Outcome `json:"no_fetch"`
	Cleanup            Outcome `json:"cleanup"`
	IntentBinding      Outcome `json:"intent_binding"`
}

// ordered returns the nine outcomes in their canonical declaration order.
func (o Outcomes) ordered() []Outcome {
	return []Outcome{
		o.SourceBinding, o.BaselineProvenance, o.CandidateBinding, o.LedgerPreserved,
		o.SnapshotComparison, o.ProofBinding, o.NoFetch, o.Cleanup, o.IntentBinding,
	}
}

// Artifact is one produced evidence artifact referenced by outcomes.
type Artifact struct {
	Role      string `json:"role"`
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

// CaseReview records the author commit and the independent-review pointer.
type CaseReview struct {
	AuthorCommit       string      `json:"author_commit"`
	IndependentVerdict CaseVerdict `json:"independent_verdict"`
}

// CaseVerdict is the explicit review-status object.
type CaseVerdict struct {
	Status string `json:"status"`
}

// ParseCaseManifest validates strict canonical JSON and returns the typed case manifest.
func ParseCaseManifest(raw []byte) (CaseManifest, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return CaseManifest{}, fmt.Errorf("case-manifest: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return CaseManifest{}, errors.New("case-manifest: JSON is not canonical")
	}
	if err := validateCaseShape(raw); err != nil {
		return CaseManifest{}, err
	}
	var cm CaseManifest
	if err := json.Unmarshal(raw, &cm); err != nil {
		return CaseManifest{}, fmt.Errorf("case-manifest: decode: %w", err)
	}
	if err := ValidateCaseManifest(cm); err != nil {
		return CaseManifest{}, err
	}
	return cm, nil
}

// CanonicalCaseManifestBytes validates and encodes a case manifest as RFC 8785 canonical JSON.
func CanonicalCaseManifestBytes(cm CaseManifest) ([]byte, error) {
	if err := ValidateCaseManifest(cm); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(cm)
	if err != nil {
		return nil, fmt.Errorf("case-manifest: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("case-manifest: canonical encode: %w", err)
	}
	return canonical, nil
}

// CaseManifestDigest returns the SHA-256 of the validated canonical case manifest bytes.
func CaseManifestDigest(raw []byte) (string, error) {
	if _, err := ParseCaseManifest(raw); err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// DeriveCaseStatus returns the worst-wins fold over the nine outcomes.
// Absence of evidence can never fold up to PASS.
func DeriveCaseStatus(cm CaseManifest) (string, error) {
	anyPending, anyFail, anyUnverifiedOrUndeclared := false, false, false
	for _, outcome := range cm.Outcomes.ordered() {
		switch outcome.Status {
		case CaseStatusPending:
			anyPending = true
		case CaseStatusFail:
			anyFail = true
		case CaseStatusUnverified, CaseStatusUndeclared:
			anyUnverifiedOrUndeclared = true
		case CaseStatusPass:
		default:
			return "", fmt.Errorf("case-manifest: unknown outcome status %q", outcome.Status)
		}
	}
	switch {
	case anyPending:
		return string(CaseStatusPending), nil
	case anyFail:
		return string(CaseStatusFail), nil
	case anyUnverifiedOrUndeclared:
		return string(CaseStatusUnverified), nil
	default:
		return string(CaseStatusPass), nil
	}
}

// ValidateCaseManifest validates a typed case manifest. ParseCaseManifest is preferred for bytes.
func ValidateCaseManifest(cm CaseManifest) error {
	if cm.Schema != CaseManifestSchema {
		return fmt.Errorf("case-manifest: schema must be %q", CaseManifestSchema)
	}
	if !caseIDRE.MatchString(cm.CaseID) {
		return fmt.Errorf("case-manifest: invalid case_id %q", cm.CaseID)
	}
	if cm.Subject != ReplayIsolatedSubject {
		return fmt.Errorf("case-manifest: subject must be %q", ReplayIsolatedSubject)
	}
	if err := validateRevision("case-manifest evidence_revision", cm.EvidenceRevision); err != nil {
		return err
	}
	if err := validateCaseRef("contract_refs.lock_addendum", cm.ContractRefs.LockAddendum); err != nil {
		return err
	}
	if err := validateCaseRef("contract_refs.fixture_matrix", cm.ContractRefs.FixtureMatrix); err != nil {
		return err
	}
	if err := validateCaseRef("contract_refs.implementation_scope", cm.ContractRefs.ImplementationScope); err != nil {
		return err
	}
	if err := validateCaseRef("inputs.lock", cm.Inputs.Lock); err != nil {
		return err
	}
	if err := validateCaseRef("inputs.source", cm.Inputs.Source); err != nil {
		return err
	}
	if err := validateCaseRef("inputs.candidate", cm.Inputs.Candidate); err != nil {
		return err
	}
	if err := validateCaseRef("inputs.setup_sources", cm.Inputs.SetupSources); err != nil {
		return err
	}
	if err := validateCaseRef("inputs.projector_manifest", cm.Inputs.ProjectorManifest); err != nil {
		return err
	}
	if err := validateCommands(cm.Commands); err != nil {
		return err
	}
	if err := validateWitnesses(cm.Witnesses); err != nil {
		return err
	}
	roles, err := validateArtifacts(cm.Artifacts)
	if err != nil {
		return err
	}
	if err := validateOutcomes(cm.Outcomes, roles); err != nil {
		return err
	}
	if err := validateReview(cm.Review); err != nil {
		return err
	}
	derived, err := DeriveCaseStatus(cm)
	if err != nil {
		return err
	}
	if cm.CaseStatus != derived {
		return fmt.Errorf("case-manifest: case_status %q does not match derived %q", cm.CaseStatus, derived)
	}
	return nil
}

func validateCaseRef(label string, ref CaseRef) error {
	if err := validateRelativePath(ref.Path); err != nil {
		return fmt.Errorf("case-manifest: %s: %w", label, err)
	}
	if err := validateSHA256(label, ref.SHA256); err != nil {
		return err
	}
	return nil
}

func validateCommands(commands []Command) error {
	if commands == nil {
		return errors.New("case-manifest: commands must be an array")
	}
	seen := make(map[string]struct{}, len(commands))
	for i, command := range commands {
		if !caseIDRE.MatchString(command.ID) {
			return fmt.Errorf("case-manifest: command %d has invalid id %q", i, command.ID)
		}
		if _, dup := seen[command.ID]; dup {
			return fmt.Errorf("case-manifest: duplicate command id %q", command.ID)
		}
		seen[command.ID] = struct{}{}
		if len(command.Argv) == 0 {
			return fmt.Errorf("case-manifest: command %q argv must be non-empty", command.ID)
		}
		for _, arg := range command.Argv {
			if arg == "" {
				return fmt.Errorf("case-manifest: command %q argv has an empty element", command.ID)
			}
		}
		if command.CWD != "." {
			if err := validateRelativePath(command.CWD); err != nil {
				return fmt.Errorf("case-manifest: command %q cwd: %w", command.ID, err)
			}
		}
		if !netPolicyRE.MatchString(command.NetworkPolicy) {
			return fmt.Errorf("case-manifest: command %q network_policy %q is not allowed", command.ID, command.NetworkPolicy)
		}
		if err := validateSHA256("command "+command.ID+" environment_digest", command.EnvironmentDigest); err != nil {
			return err
		}
		if err := validateCaseRef("command "+command.ID+" stdout", command.Stdout); err != nil {
			return err
		}
		if err := validateCaseRef("command "+command.ID+" stderr", command.Stderr); err != nil {
			return err
		}
	}
	return nil
}

func validateWitnesses(witnesses []Witness) error {
	if witnesses == nil {
		return errors.New("case-manifest: witnesses must be an array")
	}
	seen := make(map[string]struct{}, len(witnesses))
	for i, witness := range witnesses {
		if witness.Kind == "" {
			return fmt.Errorf("case-manifest: witness %d requires a kind", i)
		}
		if _, dup := seen[witness.Kind]; dup {
			return fmt.Errorf("case-manifest: duplicate witness kind %q", witness.Kind)
		}
		seen[witness.Kind] = struct{}{}
		if err := validateRelativePath(witness.Path); err != nil {
			return fmt.Errorf("case-manifest: witness %q: %w", witness.Kind, err)
		}
		if err := validateSHA256("witness "+witness.Kind, witness.SHA256); err != nil {
			return err
		}
	}
	return nil
}

func validateArtifacts(artifacts []Artifact) (map[string]struct{}, error) {
	if artifacts == nil {
		return nil, errors.New("case-manifest: artifacts must be an array")
	}
	roles := make(map[string]struct{}, len(artifacts))
	for i, artifact := range artifacts {
		if artifact.Role == "" {
			return nil, fmt.Errorf("case-manifest: artifact %d requires a role", i)
		}
		if _, dup := roles[artifact.Role]; dup {
			return nil, fmt.Errorf("case-manifest: duplicate artifact role %q", artifact.Role)
		}
		roles[artifact.Role] = struct{}{}
		if err := validateRelativePath(artifact.Path); err != nil {
			return nil, fmt.Errorf("case-manifest: artifact %q: %w", artifact.Role, err)
		}
		if artifact.SizeBytes < 0 {
			return nil, fmt.Errorf("case-manifest: artifact %q size_bytes must be non-negative", artifact.Role)
		}
		if err := validateSHA256("artifact "+artifact.Role, artifact.SHA256); err != nil {
			return nil, err
		}
	}
	return roles, nil
}

func validateOutcomes(outcomes Outcomes, roles map[string]struct{}) error {
	for i, outcome := range outcomes.ordered() {
		if err := validateOutcome(outcomeKeys[i], outcome, roles); err != nil {
			return err
		}
	}
	return nil
}

func validateOutcome(key string, outcome Outcome, roles map[string]struct{}) error {
	if outcome.Evidence == nil {
		return fmt.Errorf("case-manifest: outcome %q evidence must be an array", key)
	}
	switch outcome.Status {
	case CaseStatusPass, CaseStatusFail:
		if len(outcome.Evidence) == 0 {
			return fmt.Errorf("case-manifest: outcome %q with status %q requires evidence", key, outcome.Status)
		}
		seen := make(map[string]struct{}, len(outcome.Evidence))
		for _, role := range outcome.Evidence {
			if _, ok := roles[role]; !ok {
				return fmt.Errorf("case-manifest: outcome %q cites unknown artifact role %q", key, role)
			}
			if _, dup := seen[role]; dup {
				return fmt.Errorf("case-manifest: outcome %q cites duplicate role %q", key, role)
			}
			seen[role] = struct{}{}
		}
	case CaseStatusUnverified, CaseStatusUndeclared, CaseStatusPending:
		if len(outcome.Evidence) != 0 {
			return fmt.Errorf("case-manifest: outcome %q with status %q must have empty evidence", key, outcome.Status)
		}
	default:
		return fmt.Errorf("case-manifest: outcome %q has unknown status %q", key, outcome.Status)
	}
	return nil
}

func validateReview(review CaseReview) error {
	if err := validateRevision("review author_commit", review.AuthorCommit); err != nil {
		return err
	}
	if !verdictRE.MatchString(review.IndependentVerdict.Status) {
		return fmt.Errorf("case-manifest: review verdict status %q is not allowed", review.IndependentVerdict.Status)
	}
	return nil
}

// validateCaseShape rejects unknown, missing, null, or non-object members at every level.
func validateCaseShape(raw json.RawMessage) error {
	top, err := exactObject(raw, "schema", "case_id", "subject", "case_status", "evidence_revision",
		"contract_refs", "inputs", "commands", "witnesses", "outcomes", "artifacts", "review")
	if err != nil {
		return err
	}
	if err := refsShape(top["contract_refs"], "lock_addendum", "fixture_matrix", "implementation_scope"); err != nil {
		return err
	}
	if err := refsShape(top["inputs"], "lock", "source", "candidate", "setup_sources", "projector_manifest"); err != nil {
		return err
	}
	if err := eachObject(top["commands"], func(entry json.RawMessage) error {
		cmd, cerr := exactObject(entry, "id", "argv", "cwd", "network_policy", "environment_digest", "exit_code", "stdout", "stderr")
		if cerr != nil {
			return cerr
		}
		if _, serr := exactObject(cmd["stdout"], "path", "sha256"); serr != nil {
			return serr
		}
		_, serr := exactObject(cmd["stderr"], "path", "sha256")
		return serr
	}); err != nil {
		return err
	}
	if err := eachObject(top["witnesses"], func(entry json.RawMessage) error {
		_, werr := exactObject(entry, "kind", "path", "sha256")
		return werr
	}); err != nil {
		return err
	}
	if err := eachObject(top["artifacts"], func(entry json.RawMessage) error {
		_, aerr := exactObject(entry, "role", "path", "size_bytes", "sha256")
		return aerr
	}); err != nil {
		return err
	}
	outcomes, err := exactObject(top["outcomes"], outcomeKeys...)
	if err != nil {
		return err
	}
	for _, key := range outcomeKeys {
		if _, oerr := exactObject(outcomes[key], "status", "evidence"); oerr != nil {
			return oerr
		}
	}
	review, err := exactObject(top["review"], "author_commit", "independent_verdict")
	if err != nil {
		return err
	}
	_, err = exactObject(review["independent_verdict"], "status")
	return err
}

// refsShape validates an object whose members are each a {path, sha256} reference.
func refsShape(raw json.RawMessage, keys ...string) error {
	object, err := exactObject(raw, keys...)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if _, rerr := exactObject(object[key], "path", "sha256"); rerr != nil {
			return rerr
		}
	}
	return nil
}

// eachObject unmarshals a JSON array and applies fn to each raw element.
func eachObject(raw json.RawMessage, fn func(json.RawMessage) error) error {
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("case-manifest: expected JSON array: %w", err)
	}
	for _, entry := range entries {
		if err := fn(entry); err != nil {
			return err
		}
	}
	return nil
}
