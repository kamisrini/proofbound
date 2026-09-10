package records

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

type Artifact struct {
	Path  string
	Bytes []byte
}
type Reader interface {
	ReadIntentArtifacts(context.Context, string) ([]Artifact, error)
}
type Provider struct {
	reader     Reader
	observedAt time.Time
}

var (
	pathRE        = regexp.MustCompile(`^docs/intent/records/(business-decisions|requirements|change-intents)/((BD|BR|CI)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6})/([0-9]{4})\.md$`)
	digestFieldRE = regexp.MustCompile(`"artifact_sha256":"([0-9a-f]{64})"`)
)

func New(reader Reader, observedAt time.Time) (*Provider, error) {
	if reader == nil || observedAt.IsZero() {
		return nil, errors.New("records: reader and observation time are required")
	}
	return &Provider{reader: reader, observedAt: observedAt.UTC()}, nil
}
func (p *Provider) Name() string { return "records" }
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) {
	if p == nil || p.reader == nil {
		return nil, errors.New("records: provider is not initialized")
	}
	artifacts, err := p.reader.ReadIntentArtifacts(ctx, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("records: read committed artifacts: %w", err)
	}
	return p.parseAll(artifacts)
}
func (p *Provider) Resolve(ctx context.Context, tree, recordID string) (intent.Revision, error) {
	if p == nil || p.reader == nil || tree == "" {
		return intent.Revision{}, errors.New("records: provider and tree are required")
	}
	artifacts, err := p.reader.ReadIntentArtifacts(ctx, tree)
	if err != nil {
		return intent.Revision{}, fmt.Errorf("records: read tree %s: %w", tree, err)
	}
	revisions, err := p.parseAll(artifacts)
	if err != nil {
		return intent.Revision{}, err
	}
	var matches []intent.Revision
	for _, revision := range revisions {
		if revision.NativeID == recordID {
			matches = append(matches, revision)
		}
	}
	if len(matches) == 0 {
		return intent.Revision{}, fmt.Errorf("records: intent %q is missing", recordID)
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].ArtifactPath < matches[j].ArtifactPath })
	latest := matches[len(matches)-1]
	if latest.Kind != core.KindChangeIntent {
		return intent.Revision{}, fmt.Errorf("records: %q is not a change intent", recordID)
	}
	var ci intent.ChangeIntent
	_ = json.Unmarshal(latest.Payload, &ci)
	if ci.Status == "withdrawn" || ci.Status == "superseded" {
		return intent.Revision{}, fmt.Errorf("records: intent %q is %s", recordID, ci.Status)
	}
	return latest, nil
}

func (p *Provider) parseAll(artifacts []Artifact) ([]intent.Revision, error) {
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	revisions := make([]intent.Revision, 0, len(artifacts))
	for _, artifact := range artifacts {
		revision, err := parse(artifact, p.observedAt)
		if err != nil {
			return nil, fmt.Errorf("records: %s: %w", artifact.Path, err)
		}
		revisions = append(revisions, revision)
	}
	if err := validateLineages(revisions); err != nil {
		return nil, err
	}
	return revisions, nil
}

func parse(artifact Artifact, observedAt time.Time) (intent.Revision, error) {
	match := pathRE.FindStringSubmatch(artifact.Path)
	if match == nil || path.Clean(artifact.Path) != artifact.Path || strings.ContainsAny(artifact.Path, "\\\x00") {
		return intent.Revision{}, errors.New("invalid artifact path")
	}
	if !utf8.Valid(artifact.Bytes) {
		return intent.Revision{}, errors.New("artifact is not valid UTF-8")
	}
	prefix := "```proofbound-json\n"
	if !strings.HasPrefix(string(artifact.Bytes), prefix) {
		return intent.Revision{}, errors.New("metadata block must be first")
	}
	rest := string(artifact.Bytes[len(prefix):])
	line, tail, ok := strings.Cut(rest, "\n")
	if !ok || !strings.HasPrefix(tail, "```\n") || strings.Contains(line, "\n") || strings.TrimSpace(line) != line {
		return intent.Revision{}, errors.New("metadata block must contain one compact JSON object")
	}
	if strings.Contains(tail[len("```\n"):], "```proofbound-json") {
		return intent.Revision{}, errors.New("second metadata block")
	}
	digestMatches := digestFieldRE.FindAllSubmatchIndex(artifact.Bytes, -1)
	if len(digestMatches) == 0 {
		return intent.Revision{}, errors.New("artifact digest field is missing")
	}
	self := digestMatches[len(digestMatches)-1]
	declared := string(artifact.Bytes[self[2]:self[3]])
	normalized := append([]byte(nil), artifact.Bytes...)
	copy(normalized[self[2]:self[3]], strings.Repeat("0", 64))
	sum := sha256.Sum256(normalized)
	if hex.EncodeToString(sum[:]) != declared {
		return intent.Revision{}, errors.New("artifact_sha256 does not match normalized source bytes")
	}
	var probe struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal([]byte(line), &probe); err != nil {
		return intent.Revision{}, fmt.Errorf("metadata JSON: %w", err)
	}
	var kind core.Kind
	var nativeID, payloadPath, payloadDigest string
	switch probe.Schema {
	case "proofbound.business-decision.v1":
		kind = core.KindBusinessDecision
		var v intent.BusinessDecision
		if err := strict([]byte(line), &v); err != nil {
			return intent.Revision{}, err
		}
		nativeID, payloadPath, payloadDigest = v.DecisionID, v.ArtifactPath, v.ArtifactSHA256
	case "proofbound.requirement.v1":
		kind = core.KindRequirement
		var v intent.Requirement
		if err := strict([]byte(line), &v); err != nil {
			return intent.Revision{}, err
		}
		nativeID, payloadPath, payloadDigest = v.RequirementID, v.ArtifactPath, v.ArtifactSHA256
	case "proofbound.change-intent.v1":
		kind = core.KindChangeIntent
		var v intent.ChangeIntent
		if err := strict([]byte(line), &v); err != nil {
			return intent.Revision{}, err
		}
		nativeID, payloadPath, payloadDigest = v.IntentID, v.ArtifactPath, v.ArtifactSHA256
	default:
		return intent.Revision{}, fmt.Errorf("unknown schema %q", probe.Schema)
	}
	expectedDir, expectedPrefix := map[core.Kind][2]string{core.KindBusinessDecision: {"business-decisions", "BD"}, core.KindRequirement: {"requirements", "BR"}, core.KindChangeIntent: {"change-intents", "CI"}}[kind][0], map[core.Kind][2]string{core.KindBusinessDecision: {"business-decisions", "BD"}, core.KindRequirement: {"requirements", "BR"}, core.KindChangeIntent: {"change-intents", "CI"}}[kind][1]
	if match[1] != expectedDir || match[2] != nativeID || match[3] != expectedPrefix || match[4] == "0000" || payloadPath != artifact.Path || payloadDigest != declared {
		return intent.Revision{}, errors.New("path, schema, identity, or digest mismatch")
	}
	canonical, err := core.Canonicalize([]byte(line))
	if err != nil {
		return intent.Revision{}, err
	}
	revision := intent.Revision{Source: core.SourceIntentRecords, Kind: kind, NativeID: nativeID, ArtifactPath: artifact.Path, ArtifactSHA256: declared, ObservedAt: observedAt, Payload: canonical}
	if err := intent.ValidateRevision(revision); err != nil {
		return intent.Revision{}, err
	}
	return revision, nil
}

func strict(raw []byte, target any) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("records: strict metadata: %w", err)
	}
	return nil
}

func validateLineages(revisions []intent.Revision) error {
	type key struct{ source, id, digest string }
	available := map[key]intent.Revision{}
	byID := map[string][]intent.Revision{}
	for _, revision := range revisions {
		available[key{string(revision.Source), revision.NativeID, revision.ArtifactSHA256}] = revision
		byID[revision.NativeID] = append(byID[revision.NativeID], revision)
	}
	for _, revision := range revisions {
		for _, ref := range revisionRefs(revision) {
			if strings.HasPrefix(ref.Source, "intent.") {
				if _, ok := available[key{ref.Source, ref.RecordID, ref.ArtifactSHA256}]; !ok && ref.Source == string(core.SourceIntentRecords) {
					return fmt.Errorf("records: dangling relation %s", ref.RecordID)
				}
			}
		}
	}
	for id, lineage := range byID {
		sort.Slice(lineage, func(i, j int) bool { return lineage[i].ArtifactPath < lineage[j].ArtifactPath })
		for i := 1; i < len(lineage); i++ {
			prev, current := lineage[i-1], lineage[i]
			if !hasSupersedes(current, prev) {
				return fmt.Errorf("records: %s revision does not supersede exact predecessor", id)
			}
			if prev.Kind == core.KindRequirement {
				var a, b intent.Requirement
				_ = json.Unmarshal(prev.Payload, &a)
				_ = json.Unmarshal(current.Payload, &b)
				old := map[string]intent.Obligation{}
				for _, o := range a.Obligations {
					old[o.ID] = o
				}
				for oid, earlier := range old {
					found := false
					for _, later := range b.Obligations {
						if later.ID == oid {
							found = true
							if later.Statement != earlier.Statement {
								return fmt.Errorf("records: obligation %s changed meaning", oid)
							}
							break
						}
					}
					if !found {
						return fmt.Errorf("records: obligation %s disappeared", oid)
					}
				}
			}
		}
	}
	return nil
}
func revisionRefs(r intent.Revision) []intent.Reference {
	switch r.Kind {
	case core.KindBusinessDecision:
		var v intent.BusinessDecision
		_ = json.Unmarshal(r.Payload, &v)
		return v.Supersedes
	case core.KindRequirement:
		var v intent.Requirement
		_ = json.Unmarshal(r.Payload, &v)
		return append(v.AuthorizedBy, v.Supersedes...)
	case core.KindChangeIntent:
		var v intent.ChangeIntent
		_ = json.Unmarshal(r.Payload, &v)
		return append(append(v.Targets, v.ConstrainedBy...), v.Supersedes...)
	}
	return nil
}
func hasSupersedes(current, previous intent.Revision) bool {
	for _, r := range revisionRefs(current) {
		if r.Relation == "supersedes" && r.Source == string(previous.Source) && r.RecordID == previous.NativeID && r.ArtifactSHA256 == previous.ArtifactSHA256 {
			return true
		}
	}
	return false
}
