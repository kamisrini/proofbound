package reviews

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

type ObligationOutcome struct {
	Source           string   `json:"source"`
	RequirementID    string   `json:"requirement_id"`
	ArtifactSHA256   string   `json:"artifact_sha256"`
	ObligationID     string   `json:"obligation_id"`
	Outcome          string   `json:"outcome"`
	EvidenceEventIDs []string `json:"evidence_event_ids"`
}

type ObligationVerdict struct {
	Schema           string                      `json:"schema"`
	VerdictID        string                      `json:"verdict_id"`
	Status           string                      `json:"status"`
	DeclaredReviewer string                      `json:"declared_reviewer"`
	ReviewedCommit   string                      `json:"reviewed_commit"`
	ChangeIntent     connectorintent.Reference   `json:"change_intent"`
	Requirements     []connectorintent.Reference `json:"requirements"`
	Obligations      []ObligationOutcome         `json:"obligations"`
	Findings         []Finding                   `json:"findings"`
	ArtifactPath     string                      `json:"artifact_path"`
	ArtifactSHA256   string                      `json:"artifact_sha256"`
}

type RequirementReviewOutcome struct {
	ObligationID string `json:"obligation_id"`
	Outcome      string `json:"outcome"`
	Finding      string `json:"finding,omitempty"`
}

type RequirementReview struct {
	Schema           string                     `json:"schema"`
	ReviewID         string                     `json:"review_id"`
	Requirement      connectorintent.Reference  `json:"requirement"`
	DeclaredReviewer string                     `json:"declared_reviewer"`
	Outcomes         []RequirementReviewOutcome `json:"outcomes"`
	ArtifactPath     string                     `json:"artifact_path"`
	ArtifactSHA256   string                     `json:"artifact_sha256"`
}

func ParseObligationVerdict(path string, data []byte) (ObligationVerdict, error) {
	var v ObligationVerdict
	raw, err := strictJSONFrontMatter(path, data)
	if err != nil {
		return v, err
	}
	want := []string{"schema", "verdict_id", "status", "declared_reviewer", "reviewed_commit", "change_intent", "requirements", "obligations", "findings", "artifact_path", "artifact_sha256"}
	if err := decodeExactOrdered(raw, want, &v); err != nil {
		return v, err
	}
	if v.Schema != "proofbound.obligation-verdict.v2" {
		return v, errors.New("obligation verdict schema is invalid")
	}
	if !idRE.MatchString(v.VerdictID) {
		return v, errors.New("obligation verdict id is invalid")
	}
	if !oneOf(v.Status, "ACCEPTABLE", "NEEDS_WORK") {
		return v, errors.New("obligation verdict status is invalid")
	}
	if strings.TrimSpace(v.DeclaredReviewer) == "" {
		return v, errors.New("obligation verdict reviewer is invalid")
	}
	if !commitRE.MatchString(v.ReviewedCommit) {
		return v, errors.New("obligation verdict commit is invalid")
	}
	if v.ArtifactPath != path {
		return v, errors.New("obligation verdict path does not match")
	}
	if !validPath(v.ArtifactPath) {
		return v, errors.New("obligation verdict path is invalid")
	}
	if !digestRE.MatchString(v.ArtifactSHA256) {
		return v, errors.New("obligation verdict digest shape is invalid")
	}
	if v.ArtifactSHA256 != ArtifactSHA256(data) {
		return v, errors.New("obligation verdict digest is invalid")
	}
	if err := validateReference(v.ChangeIntent, "change_intent", "evaluates"); err != nil {
		return v, err
	}
	if len(v.Requirements) == 0 {
		return v, errors.New("obligation verdict requires revisions")
	}
	if len(v.Obligations) == 0 {
		return v, errors.New("obligation verdict requires outcomes")
	}
	last := ""
	requirements := map[string]bool{}
	for _, ref := range v.Requirements {
		if err := validateReference(ref, "requirement", "evaluates"); err != nil {
			return v, err
		}
		key := referenceKey(ref)
		if key <= last {
			return v, errors.New("requirements are not sorted and unique")
		}
		last = key
		requirements[ref.Source+"\x00"+ref.RecordID+"\x00"+ref.ArtifactSHA256] = true
	}
	last = ""
	allSatisfied := true
	for _, outcome := range v.Obligations {
		key := outcome.Source + "\x00" + outcome.RequirementID + "\x00" + outcome.ArtifactSHA256 + "\x00" + outcome.ObligationID
		if key <= last {
			return v, errors.New("obligation outcomes are not sorted and unique")
		}
		if !requirements[outcome.Source+"\x00"+outcome.RequirementID+"\x00"+outcome.ArtifactSHA256] {
			return v, errors.New("obligation outcome is not targeted")
		}
		if !digestRE.MatchString(outcome.ArtifactSHA256) || !idRE.MatchString(outcome.ObligationID) {
			return v, errors.New("obligation outcome identity is invalid")
		}
		if !oneOf(outcome.Outcome, "SATISFIED", "NOT_SATISFIED", "INCONCLUSIVE") {
			return v, errors.New("obligation outcome is invalid")
		}
		last = key
		if outcome.Outcome != "SATISFIED" {
			allSatisfied = false
		}
		if outcome.Outcome == "SATISFIED" && len(outcome.EvidenceEventIDs) == 0 {
			return v, errors.New("satisfied outcome requires evidence")
		}
		previous := ""
		for _, eventID := range outcome.EvidenceEventIDs {
			if _, err := core.ParseEventID(eventID); err != nil || eventID <= previous {
				return v, errors.New("invalid, unsorted, or duplicate evidence event id")
			}
			previous = eventID
		}
	}
	if v.Status == "ACCEPTABLE" && !allSatisfied {
		return v, errors.New("ACCEPTABLE verdict contains a non-satisfied obligation")
	}
	for _, finding := range v.Findings {
		if err := validateFinding(finding); err != nil {
			return v, err
		}
	}
	return v, nil
}

func ParseRequirementReview(path string, data []byte) (RequirementReview, error) {
	var v RequirementReview
	raw, err := strictJSONFrontMatter(path, data)
	if err != nil {
		return v, err
	}
	want := []string{"schema", "review_id", "requirement", "declared_reviewer", "outcomes", "artifact_path", "artifact_sha256"}
	if err := decodeExactOrdered(raw, want, &v); err != nil {
		return v, err
	}
	if v.Schema != "proofbound.requirement-review.v1" {
		return v, errors.New("requirement review schema is invalid")
	}
	if !idRE.MatchString(v.ReviewID) {
		return v, errors.New("requirement review id is invalid")
	}
	if strings.TrimSpace(v.DeclaredReviewer) == "" {
		return v, errors.New("requirement review reviewer is invalid")
	}
	if v.ArtifactPath != path {
		return v, errors.New("requirement review path does not match")
	}
	if !validPath(v.ArtifactPath) {
		return v, errors.New("requirement review path is invalid")
	}
	if !digestRE.MatchString(v.ArtifactSHA256) {
		return v, errors.New("requirement review digest shape is invalid")
	}
	if v.ArtifactSHA256 != ArtifactSHA256(data) {
		return v, errors.New("requirement review digest is invalid")
	}
	if err := validateReference(v.Requirement, "requirement", "reviews"); err != nil {
		return v, err
	}
	if len(v.Outcomes) == 0 {
		return v, errors.New("requirement review has no outcomes")
	}
	last := ""
	for _, outcome := range v.Outcomes {
		if !idRE.MatchString(outcome.ObligationID) || outcome.ObligationID <= last {
			return v, errors.New("requirement-review outcomes are invalid, unsorted, or duplicate")
		}
		if !oneOf(outcome.Outcome, "VERIFIABLE", "AMBIGUOUS", "UNTESTABLE", "CONTRADICTORY") {
			return v, errors.New("requirement-review outcome is unknown")
		}
		if strings.ContainsAny(outcome.Finding, "\r\n") {
			return v, errors.New("requirement-review finding is not one line")
		}
		last = outcome.ObligationID
	}
	return v, nil
}

func strictJSONFrontMatter(path string, data []byte) ([]byte, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("artifact is not valid UTF-8")
	}
	if !validPath(path) {
		return nil, errors.New("artifact path is invalid")
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		return nil, errors.New("JSON front matter is incomplete")
	}
	if lines[0] != "---" || lines[2] != "---" {
		return nil, errors.New("JSON front matter delimiters are invalid")
	}
	if !strings.HasPrefix(lines[1], "{") || strings.TrimSpace(lines[1]) != lines[1] {
		return nil, errors.New("JSON front matter must be one compact line between delimiters")
	}
	return []byte(lines[1]), nil
}
func decodeExactOrdered(raw []byte, want []string, target any) error {
	keys, err := topLevelKeys(raw)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(keys, want) {
		return fmt.Errorf("front matter fields or order are invalid: %v", keys)
	}
	if _, err := core.Canonicalize(raw); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON")
	}
	return nil
}
func topLevelKeys(raw []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if token != json.Delim('{') {
		return nil, errors.New("front matter must be an object")
	}
	var keys []string
	for dec.More() {
		token, err = dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("object key is not a string")
		}
		keys = append(keys, key)
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return keys, nil
}
func validateReference(ref connectorintent.Reference, kind, relation string) error {
	if ref.RecordKind != kind {
		return errors.New("exact record reference kind is invalid")
	}
	if ref.Relation != relation {
		return errors.New("exact record reference kind or relation is invalid")
	}
	if !core.Source(ref.Source).WellFormed() {
		return errors.New("exact record reference source is invalid")
	}
	if !idRE.MatchString(ref.RecordID) {
		return errors.New("exact record reference id is invalid")
	}
	if !digestRE.MatchString(ref.ArtifactSHA256) {
		return errors.New("exact record reference identity is invalid")
	}
	if len(ref.ObligationIDs) != 0 {
		return errors.New("exact record reference cannot name obligations")
	}
	return nil
}
func referenceKey(ref connectorintent.Reference) string {
	return ref.Source + "\x00" + ref.RecordID + "\x00" + ref.ArtifactSHA256
}
func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func ArtifactSHA256(data []byte) string {
	needle := []byte(`"artifact_sha256":"`)
	starts := []int{}
	for offset := 0; ; {
		i := bytes.Index(data[offset:], needle)
		if i < 0 {
			break
		}
		starts = append(starts, offset+i+len(needle))
		offset += i + len(needle)
	}
	normalized := append([]byte(nil), data...)
	if len(starts) > 0 {
		start := starts[len(starts)-1]
		if start+64 <= len(normalized) {
			copy(normalized[start:start+64], strings.Repeat("0", 64))
		}
	}
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:])
}
