package records

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

type fixtureReader struct {
	artifacts []Artifact
	refs      []string
}

func (r *fixtureReader) ReadIntentArtifacts(_ context.Context, ref string) ([]Artifact, error) {
	r.refs = append(r.refs, ref)
	return append([]Artifact(nil), r.artifacts...), nil
}
func validArtifacts(t *testing.T) []Artifact {
	t.Helper()
	root := "testdata/valid"
	paths := []string{"business-decisions/BD-proofbound-p5-a1b2c3/0001.md", "requirements/BR-intent-chain-d4e5f6/0001.md", "change-intents/CI-implement-p5-0a1b2c/0001.md"}
	var out []Artifact
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, Artifact{Path: "docs/intent/records/" + filepath.ToSlash(name), Bytes: data})
	}
	return out
}
func provider(t *testing.T, r Reader) *Provider {
	t.Helper()
	p, err := New(r, time.Unix(10, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNewRequiresReaderAndObservationTime(t *testing.T) {
	reader := &fixtureReader{}
	if _, err := New(nil, time.Unix(1, 0)); err == nil {
		t.Fatal("nil reader accepted")
	}
	if _, err := New(reader, time.Time{}); err == nil {
		t.Fatal("zero observation time accepted")
	}
}
func TestRevisionsRejectsUninitializedProvider(t *testing.T) {
	var nilProvider *Provider
	if _, err := nilProvider.Revisions(context.Background()); err == nil {
		t.Fatal("nil provider accepted")
	}
	if _, err := (&Provider{}).Revisions(context.Background()); err == nil {
		t.Fatal("provider without reader accepted")
	}
}
func TestResolveRejectsUninitializedProviderAndEmptyTree(t *testing.T) {
	var nilProvider *Provider
	if _, err := nilProvider.Resolve(context.Background(), "HEAD", "CI-item-change-acde12"); err == nil {
		t.Fatal("nil provider accepted")
	}
	if _, err := (&Provider{}).Resolve(context.Background(), "HEAD", "CI-item-change-acde12"); err == nil {
		t.Fatal("provider without reader accepted")
	}
	p := provider(t, &fixtureReader{artifacts: validArtifacts(t)})
	if _, err := p.Resolve(context.Background(), "", "CI-implement-p5-0a1b2c"); err == nil {
		t.Fatal("empty tree accepted")
	}
}
func rehash(data []byte) []byte {
	matches := digestFieldRE.FindAllSubmatchIndex(data, -1)
	self := matches[len(matches)-1]
	zero := append([]byte(nil), data...)
	copy(zero[self[2]:self[3]], strings.Repeat("0", 64))
	sum := sha256.Sum256(zero)
	copy(zero[self[2]:self[3]], hex.EncodeToString(sum[:]))
	return zero
}

func TestCommittedReaderBoundary(t *testing.T) {
	r := &fixtureReader{artifacts: validArtifacts(t)}
	revs, err := provider(t, r).Revisions(context.Background())
	if err != nil || len(revs) != 3 || len(r.refs) != 1 || r.refs[0] != "HEAD" {
		t.Fatalf("revisions=%d refs=%v err=%v", len(revs), r.refs, err)
	}
}
func TestHostileFixturesFailClosed(t *testing.T) {
	valid := validArtifacts(t)[0]
	cases := map[string]Artifact{"utf8": {Path: valid.Path, Bytes: []byte{0xff}}, "digest": {Path: valid.Path, Bytes: []byte(strings.Replace(string(valid.Bytes), "810c2932", "ffffffff", 1))}, "traversal": {Path: "docs/intent/records/business-decisions/../escape/0001.md", Bytes: valid.Bytes}, "id-path": {Path: strings.Replace(valid.Path, "BD-proofbound-p5-a1b2c3", "BD-other-record-ffffff", 1), Bytes: valid.Bytes}, "unknown": {Path: valid.Path, Bytes: rehash([]byte(strings.Replace(string(valid.Bytes), `"status":"accepted"`, `"status":"accepted","unknown":true`, 1)))}, "closing fence": {Path: valid.Path, Bytes: rehash([]byte(strings.Replace(string(valid.Bytes), "\n```\n", "\nnot-close\n", 1)))}, "spaced json": {Path: valid.Path, Bytes: rehash([]byte(strings.Replace(string(valid.Bytes), "```proofbound-json\n{", "```proofbound-json\n {", 1)))}}
	for name, a := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parse(a, time.Unix(1, 0)); err == nil {
				t.Fatal("hostile artifact accepted")
			}
		})
	}
}
func TestRecordKindDirectoryMustMatchSchema(t *testing.T) {
	a := validArtifacts(t)[0]
	a.Path = strings.Replace(a.Path, "business-decisions", "requirements", 1)
	a.Bytes = rehash([]byte(strings.Replace(string(a.Bytes), "business-decisions", "requirements", 1)))
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("business decision accepted from requirement directory")
	}
}
func TestRecordPathIDMustMatchPayloadID(t *testing.T) {
	a := validArtifacts(t)[0]
	a.Bytes = rehash([]byte(strings.Replace(string(a.Bytes), `"decision_id":"BD-proofbound-p5-a1b2c3"`, `"decision_id":"BD-other-record-acde12"`, 1)))
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("payload identity different from path accepted")
	}
}
func TestRecordIDPrefixMustMatchSchema(t *testing.T) {
	a := validArtifacts(t)[0]
	oldID, newID := "BD-proofbound-p5-a1b2c3", "CI-other-record-acde12"
	a.Path = strings.Replace(a.Path, oldID, newID, 1)
	data := strings.ReplaceAll(string(a.Bytes), oldID, newID)
	a.Bytes = rehash([]byte(data))
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("change-intent-prefixed id accepted for business decision")
	}
}
func TestRevisionZeroFailsClosed(t *testing.T) {
	a := validArtifacts(t)[0]
	a.Path = strings.Replace(a.Path, "0001.md", "0000.md", 1)
	a.Bytes = rehash([]byte(strings.Replace(string(a.Bytes), "0001.md", "0000.md", 1)))
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("revision zero accepted")
	}
}
func TestDeclaredArtifactPathMustMatchObservedPath(t *testing.T) {
	a := validArtifacts(t)[0]
	a.Bytes = rehash([]byte(strings.Replace(string(a.Bytes), `"artifact_path":"docs/intent/records/business-decisions/BD-proofbound-p5-a1b2c3/0001.md"`, `"artifact_path":"docs/intent/records/business-decisions/BD-proofbound-p5-a1b2c3/0002.md"`, 1)))
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("declared artifact path different from observed path accepted")
	}
}
func TestValidVectors(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: validArtifacts(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 3 || revs[0].Kind != core.KindBusinessDecision || revs[1].Kind != core.KindChangeIntent || revs[2].Kind != core.KindRequirement {
		t.Fatalf("revisions=%+v", revs)
	}
	for _, r := range revs {
		canonical, err := core.Canonicalize(r.Payload)
		if err != nil || string(canonical) != string(r.Payload) {
			t.Fatalf("payload not pinned canonical: %s %v", r.NativeID, err)
		}
	}
}
func TestRevisionIdentity(t *testing.T) {
	a := validArtifacts(t)[0]
	r1, err := parse(a, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	r2, err := parse(a, time.Unix(2, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r1.ArtifactSHA256 != r2.ArtifactSHA256 || string(r1.Payload) != string(r2.Payload) {
		t.Fatal("same bytes changed revision identity")
	}
}
func TestLifecycleTransitions(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: validArtifacts(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current := revs[2]
	var v intent.Requirement
	if err := json.Unmarshal(current.Payload, &v); err != nil {
		t.Fatal(err)
	}
	v.ArtifactSHA256 = strings.Repeat("e", 64)
	v.ArtifactPath = strings.Replace(v.ArtifactPath, "0001.md", "0002.md", 1)
	v.Supersedes = []intent.Reference{{RecordKind: "requirement", Source: string(current.Source), RecordID: current.NativeID, ArtifactSHA256: current.ArtifactSHA256, Relation: "supersedes"}}
	raw, _ := json.Marshal(v)
	canonical, _ := core.Canonicalize(raw)
	next := intent.Revision{Source: current.Source, Kind: current.Kind, NativeID: current.NativeID, ArtifactPath: v.ArtifactPath, ArtifactSHA256: v.ArtifactSHA256, ObservedAt: current.ObservedAt, Payload: canonical}
	if err := validateLineages(append(revs, next)); err != nil {
		t.Fatal(err)
	}
	v.Supersedes = nil
	raw, _ = json.Marshal(v)
	next.Payload, _ = core.Canonicalize(raw)
	if validateLineages(append(revs, next)) == nil {
		t.Fatal("missing predecessor accepted")
	}
}

func TestLineageRequiresExactPredecessorReference(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: validArtifacts(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current := revs[2]
	var base intent.Requirement
	if err := json.Unmarshal(current.Payload, &base); err != nil {
		t.Fatal(err)
	}
	base.ArtifactSHA256 = strings.Repeat("e", 64)
	base.ArtifactPath = strings.Replace(base.ArtifactPath, "0001.md", "0002.md", 1)
	base.Supersedes = []intent.Reference{{RecordKind: "requirement", Source: string(current.Source), RecordID: current.NativeID, ArtifactSHA256: current.ArtifactSHA256, Relation: "supersedes"}}

	cases := map[string]func(*intent.Reference){
		"relation":  func(r *intent.Reference) { r.Relation = "authorizes" },
		"source":    func(r *intent.Reference) { r.Source = "intent.specdir" },
		"record id": func(r *intent.Reference) { r.RecordID = "BR-other-record-acde12" },
		"digest":    func(r *intent.Reference) { r.ArtifactSHA256 = strings.Repeat("f", 64) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := base
			candidate.Supersedes = append([]intent.Reference(nil), base.Supersedes...)
			mutate(&candidate.Supersedes[0])
			raw, _ := json.Marshal(candidate)
			payload, _ := core.Canonicalize(raw)
			next := intent.Revision{Source: current.Source, Kind: current.Kind, NativeID: current.NativeID, ArtifactPath: candidate.ArtifactPath, ArtifactSHA256: candidate.ArtifactSHA256, ObservedAt: current.ObservedAt, Payload: payload}
			if validateLineages(append(revs, next)) == nil {
				t.Fatal("inexact predecessor accepted")
			}
		})
	}
}
func TestObligationLineage(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: validArtifacts(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current := revs[2]
	var v intent.Requirement
	_ = json.Unmarshal(current.Payload, &v)
	v.ArtifactSHA256 = strings.Repeat("e", 64)
	v.ArtifactPath = strings.Replace(v.ArtifactPath, "0001.md", "0002.md", 1)
	v.Supersedes = []intent.Reference{{RecordKind: "requirement", Source: string(current.Source), RecordID: current.NativeID, ArtifactSHA256: current.ArtifactSHA256, Relation: "supersedes"}}
	v.Obligations = v.Obligations[1:]
	raw, _ := json.Marshal(v)
	canonical, _ := core.Canonicalize(raw)
	next := intent.Revision{Source: current.Source, Kind: current.Kind, NativeID: current.NativeID, ArtifactPath: v.ArtifactPath, ArtifactSHA256: v.ArtifactSHA256, ObservedAt: current.ObservedAt, Payload: canonical}
	if validateLineages(append(revs, next)) == nil {
		t.Fatal("obligation disappearance accepted")
	}
}
func TestResolveUsesCommitTree(t *testing.T) {
	r := &fixtureReader{artifacts: validArtifacts(t)}
	got, err := provider(t, r).Resolve(context.Background(), "abc123", "CI-implement-p5-0a1b2c")
	if err != nil || got.Kind != core.KindChangeIntent || r.refs[0] != "abc123" {
		t.Fatalf("got=%+v refs=%v err=%v", got, r.refs, err)
	}
}
func TestResolveFailsClosed(t *testing.T) {
	p := provider(t, &fixtureReader{artifacts: validArtifacts(t)})
	if _, err := p.Resolve(context.Background(), "abc123", "CI-missing-item-acde12"); err == nil {
		t.Fatal("missing intent resolved")
	}
	if _, err := p.Resolve(context.Background(), "abc123", "BR-intent-chain-d4e5f6"); err == nil {
		t.Fatal("requirement resolved as intent")
	}
}

func TestDanglingNativeRelationFailsClosed(t *testing.T) {
	artifacts := validArtifacts(t)
	artifacts = artifacts[1:]
	if _, err := provider(t, &fixtureReader{artifacts: artifacts}).Revisions(context.Background()); err == nil {
		t.Fatal("native relation to absent revision accepted")
	}
}
