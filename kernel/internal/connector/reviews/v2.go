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
	"sort"
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
	if v.Schema != "proofbound.obligation-verdict.v2" || !idRE.MatchString(v.VerdictID) || (v.Status != "ACCEPTABLE" && v.Status != "NEEDS_WORK") || strings.TrimSpace(v.DeclaredReviewer) == "" || !commitRE.MatchString(v.ReviewedCommit) || v.ArtifactPath != path || !validPath(v.ArtifactPath) || !digestRE.MatchString(v.ArtifactSHA256) || v.ArtifactSHA256 != ArtifactSHA256(data) {
		return v, errors.New("obligation verdict identity, status, commit, path, or digest is invalid")
	}
	if err := validateReference(v.ChangeIntent, "change_intent", "evaluates", false); err != nil {
		return v, err
	}
	if len(v.Requirements) == 0 || len(v.Obligations) == 0 {
		return v, errors.New("obligation verdict requires revisions and outcomes")
	}
	last := ""
	requirements := map[string]bool{}
	for _, ref := range v.Requirements {
		if err := validateReference(ref, "requirement", "evaluates", false); err != nil {
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
		if key <= last || !requirements[outcome.Source+"\x00"+outcome.RequirementID+"\x00"+outcome.ArtifactSHA256] || !digestRE.MatchString(outcome.ArtifactSHA256) || !idRE.MatchString(outcome.ObligationID) || !oneOf(outcome.Outcome, "SATISFIED", "NOT_SATISFIED", "INCONCLUSIVE") {
			return v, errors.New("invalid, untargeted, or duplicate obligation outcome")
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
	if v.Schema != "proofbound.requirement-review.v1" || !idRE.MatchString(v.ReviewID) || strings.TrimSpace(v.DeclaredReviewer) == "" || v.ArtifactPath != path || !validPath(v.ArtifactPath) || !digestRE.MatchString(v.ArtifactSHA256) || v.ArtifactSHA256 != ArtifactSHA256(data) {
		return v, errors.New("requirement review identity, path, or digest is invalid")
	}
	if err := validateReference(v.Requirement, "requirement", "reviews", false); err != nil {
		return v, err
	}
	if len(v.Outcomes) == 0 {
		return v, errors.New("requirement review has no outcomes")
	}
	last := ""
	for _, outcome := range v.Outcomes {
		if !idRE.MatchString(outcome.ObligationID) || outcome.ObligationID <= last || !oneOf(outcome.Outcome, "VERIFIABLE", "AMBIGUOUS", "UNTESTABLE", "CONTRADICTORY") || strings.ContainsAny(outcome.Finding, "\r\n") {
			return v, errors.New("invalid, unsorted, duplicate, or unknown requirement-review outcome")
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
	if len(lines) < 3 || lines[0] != "---" || lines[2] != "---" || !strings.HasPrefix(lines[1], "{") || strings.TrimSpace(lines[1]) != lines[1] {
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
	if err != nil || token != json.Delim('{') {
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
func validateReference(ref connectorintent.Reference, kind, relation string, obligations bool) error {
	if ref.RecordKind != kind || ref.Relation != relation || !core.Source(ref.Source).WellFormed() || !idRE.MatchString(ref.RecordID) || !digestRE.MatchString(ref.ArtifactSHA256) || (obligations && len(ref.ObligationIDs) == 0) || (!obligations && len(ref.ObligationIDs) != 0) {
		return errors.New("exact record reference is invalid")
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

func sortedUniqueStrings(values []string) []string {
	sort.Strings(values)
	out := values[:0]
	for _, value := range values {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}
