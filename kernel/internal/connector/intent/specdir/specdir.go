package specdir

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
	ReadSpecArtifacts(context.Context, string) ([]Artifact, error)
}
type Provider struct {
	reader     Reader
	observedAt time.Time
}

var (
	pathRE              = regexp.MustCompile(`^specs/([a-z0-9][a-z0-9-]{0,62})/(requirements|change-intent)\.md$`)
	requirementHeaderRE = regexp.MustCompile(`^# Requirement: ((?:BR)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6})$`)
	intentHeaderRE      = regexp.MustCompile(`^# Change Intent: ((?:CI)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6})$`)
	obligationLineRE    = regexp.MustCompile(`^- \[(O-[A-Z0-9][A-Z0-9-]{0,31})\] (.+)$`)
	targetLineRE        = regexp.MustCompile(`^- (implements|modifies|repairs|retires) (intent\.(?:records|specdir)):((?:BR)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6})@([0-9a-f]{64})#(O-[A-Z0-9][A-Z0-9-]{0,31}(?:,O-[A-Z0-9][A-Z0-9-]{0,31})*)$`)
)

func New(reader Reader, observedAt time.Time) (*Provider, error) {
	if reader == nil || observedAt.IsZero() {
		return nil, errors.New("specdir: reader and observation time are required")
	}
	return &Provider{reader: reader, observedAt: observedAt.UTC()}, nil
}
func (p *Provider) Name() string { return "specdir" }
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) {
	if p == nil || p.reader == nil {
		return nil, errors.New("specdir: provider is not initialized")
	}
	a, err := p.reader.ReadSpecArtifacts(ctx, "HEAD")
	if err != nil {
		return nil, err
	}
	return p.parseAll(a)
}
func (p *Provider) Resolve(ctx context.Context, tree, recordID string) (intent.Revision, error) {
	if p == nil || p.reader == nil || tree == "" {
		return intent.Revision{}, errors.New("specdir: provider and tree are required")
	}
	a, err := p.reader.ReadSpecArtifacts(ctx, tree)
	if err != nil {
		return intent.Revision{}, err
	}
	revisions, err := p.parseAll(a)
	if err != nil {
		return intent.Revision{}, err
	}
	var found []intent.Revision
	for _, r := range revisions {
		if r.NativeID == recordID && r.Kind == core.KindChangeIntent {
			found = append(found, r)
		}
	}
	if len(found) != 1 {
		return intent.Revision{}, fmt.Errorf("specdir: intent %q resolves to %d artifacts", recordID, len(found))
	}
	var ci intent.ChangeIntent
	_ = json.Unmarshal(found[0].Payload, &ci)
	if ci.Status == "withdrawn" || ci.Status == "superseded" {
		return intent.Revision{}, fmt.Errorf("specdir: intent %q is %s", recordID, ci.Status)
	}
	return found[0], nil
}

func (p *Provider) parseAll(artifacts []Artifact) ([]intent.Revision, error) {
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	var out []intent.Revision
	for _, a := range artifacts {
		r, err := parse(a, p.observedAt)
		if err != nil {
			return nil, fmt.Errorf("specdir: %s: %w", a.Path, err)
		}
		out = append(out, r)
	}
	type key struct{ source, id, digest string }
	available := map[key]intent.Revision{}
	obligations := map[key]map[string]bool{}
	for _, r := range out {
		available[key{string(r.Source), r.NativeID, r.ArtifactSHA256}] = r
		if r.Kind == core.KindRequirement {
			var br intent.Requirement
			_ = json.Unmarshal(r.Payload, &br)
			ids := map[string]bool{}
			for _, o := range br.Obligations {
				ids[o.ID] = true
			}
			obligations[key{string(r.Source), r.NativeID, r.ArtifactSHA256}] = ids
		}
	}
	for _, r := range out {
		if r.Kind != core.KindChangeIntent {
			continue
		}
		var ci intent.ChangeIntent
		_ = json.Unmarshal(r.Payload, &ci)
		for _, target := range ci.Targets {
			if target.Source != string(core.SourceIntentSpecdir) {
				continue
			}
			k := key{target.Source, target.RecordID, target.ArtifactSHA256}
			if _, ok := available[k]; !ok {
				return nil, fmt.Errorf("specdir: dangling target %s", target.RecordID)
			}
			for _, id := range target.ObligationIDs {
				if !obligations[k][id] {
					return nil, fmt.Errorf("specdir: target obligation %s is absent", id)
				}
			}
		}
	}
	return out, nil
}

func parse(a Artifact, observed time.Time) (intent.Revision, error) {
	m := pathRE.FindStringSubmatch(a.Path)
	if m == nil || path.Clean(a.Path) != a.Path || strings.ContainsAny(a.Path, "\\\x00") {
		return intent.Revision{}, errors.New("invalid artifact path")
	}
	if !utf8.Valid(a.Bytes) {
		return intent.Revision{}, errors.New("artifact is not valid UTF-8")
	}
	text := strings.TrimSuffix(string(a.Bytes), "\n")
	lines := strings.Split(text, "\n")
	if strings.Contains(text, "[when feature flag") {
		return intent.Revision{}, intent.ErrUnmappable
	}
	sum := sha256.Sum256(a.Bytes)
	digest := hex.EncodeToString(sum[:])
	if m[2] == "requirements" {
		return parseRequirement(a.Path, digest, lines, observed)
	}
	return parseIntent(a.Path, digest, lines, observed)
}
func parseRequirement(artifactPath, digest string, lines []string, observed time.Time) (intent.Revision, error) {
	if len(lines) < 6 {
		return intent.Revision{}, errors.New("requirement shape is incomplete")
	}
	h := requirementHeaderRE.FindStringSubmatch(lines[0])
	if h == nil || !strings.HasPrefix(lines[1], "Status: ") || !strings.HasPrefix(lines[2], "Owner: ") || lines[3] != "" || lines[4] != "## Obligations" {
		return intent.Revision{}, errors.New("requirement metadata is malformed or out of order")
	}
	status := strings.TrimPrefix(lines[1], "Status: ")
	owner := strings.TrimPrefix(lines[2], "Owner: ")
	if !oneOf(status, "proposed", "active", "superseded", "retired") || strings.TrimSpace(owner) == "" {
		return intent.Revision{}, errors.New("invalid requirement metadata")
	}
	var obligations []intent.Obligation
	for _, line := range lines[5:] {
		match := obligationLineRE.FindStringSubmatch(line)
		if match == nil {
			return intent.Revision{}, errors.New("invalid obligation")
		}
		obligations = append(obligations, intent.Obligation{ID: match[1], Statement: match[2], State: "active"})
	}
	sort.Slice(obligations, func(i, j int) bool { return obligations[i].ID < obligations[j].ID })
	for i := 1; i < len(obligations); i++ {
		if obligations[i].ID == obligations[i-1].ID {
			return intent.Revision{}, errors.New("duplicate obligation")
		}
	}
	v := intent.Requirement{Schema: "proofbound.requirement.v1", RequirementID: h[1], Status: status, AuthorizedBy: []intent.Reference{}, DeclaredOwner: owner, Obligations: obligations, Supersedes: []intent.Reference{}, ArtifactPath: artifactPath, ArtifactSHA256: digest}
	raw, _ := json.Marshal(v)
	canonical, _ := core.Canonicalize(raw)
	r := intent.Revision{Source: core.SourceIntentSpecdir, Kind: core.KindRequirement, NativeID: h[1], ArtifactPath: artifactPath, ArtifactSHA256: digest, ObservedAt: observed, Payload: canonical}
	if err := intent.ValidateRevision(r); err != nil {
		return intent.Revision{}, err
	}
	return r, nil
}
func parseIntent(artifactPath, digest string, lines []string, observed time.Time) (intent.Revision, error) {
	if len(lines) < 6 {
		return intent.Revision{}, errors.New("change intent shape is incomplete")
	}
	h := intentHeaderRE.FindStringSubmatch(lines[0])
	if h == nil || !strings.HasPrefix(lines[1], "Status: ") || !strings.HasPrefix(lines[2], "Sponsor: ") || lines[3] != "" || lines[4] != "## Targets" {
		return intent.Revision{}, errors.New("change intent metadata is malformed or out of order")
	}
	status := strings.TrimPrefix(lines[1], "Status: ")
	sponsor := strings.TrimPrefix(lines[2], "Sponsor: ")
	if !oneOf(status, "proposed", "accepted", "superseded", "withdrawn") || strings.TrimSpace(sponsor) == "" {
		return intent.Revision{}, errors.New("invalid change intent metadata")
	}
	var targets []intent.Reference
	for _, line := range lines[5:] {
		match := targetLineRE.FindStringSubmatch(line)
		if match == nil {
			return intent.Revision{}, errors.New("invalid target")
		}
		ids := strings.Split(match[5], ",")
		sort.Strings(ids)
		targets = append(targets, intent.Reference{RecordKind: "requirement", Source: match[2], RecordID: match[3], ArtifactSHA256: match[4], Relation: match[1], ObligationIDs: ids})
	}
	sort.Slice(targets, func(i, j int) bool { return refKey(targets[i]) < refKey(targets[j]) })
	for i := 1; i < len(targets); i++ {
		if refKey(targets[i]) == refKey(targets[i-1]) {
			return intent.Revision{}, errors.New("duplicate target")
		}
	}
	v := intent.ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: h[1], Status: status, DeclaredSponsor: sponsor, Targets: targets, ConstrainedBy: []intent.Reference{}, Supersedes: []intent.Reference{}, ArtifactPath: artifactPath, ArtifactSHA256: digest}
	raw, _ := json.Marshal(v)
	canonical, _ := core.Canonicalize(raw)
	r := intent.Revision{Source: core.SourceIntentSpecdir, Kind: core.KindChangeIntent, NativeID: h[1], ArtifactPath: artifactPath, ArtifactSHA256: digest, ObservedAt: observed, Payload: canonical}
	if err := intent.ValidateRevision(r); err != nil {
		return intent.Revision{}, err
	}
	return r, nil
}
func oneOf(v string, values ...string) bool {
	for _, x := range values {
		if v == x {
			return true
		}
	}
	return false
}
func refKey(r intent.Reference) string {
	return r.Source + "\x00" + r.RecordID + "\x00" + r.ArtifactSHA256 + "\x00" + r.Relation + "\x00" + strings.Join(r.ObligationIDs, ",")
}
