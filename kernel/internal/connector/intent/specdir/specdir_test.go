package specdir

import (
	"context"
	"encoding/json"
	"errors"
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

func (r *fixtureReader) ReadSpecArtifacts(_ context.Context, ref string) ([]Artifact, error) {
	r.refs = append(r.refs, ref)
	return append([]Artifact(nil), r.artifacts...), nil
}
func fixtures(t *testing.T) []Artifact {
	t.Helper()
	var out []Artifact
	for _, name := range []string{"payments/requirements.md", "payments/change-intent.md"} {
		data, err := os.ReadFile(filepath.Join("testdata/valid", name))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, Artifact{Path: "specs/" + filepath.ToSlash(name), Bytes: data})
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
func TestProviderMethodsRejectInvalidInputs(t *testing.T) {
	var nilProvider *Provider
	if _, err := nilProvider.Revisions(context.Background()); err == nil {
		t.Fatal("nil provider accepted")
	}
	if _, err := (&Provider{}).Revisions(context.Background()); err == nil {
		t.Fatal("provider without reader accepted")
	}
	p := provider(t, &fixtureReader{artifacts: fixtures(t)})
	if _, err := nilProvider.Resolve(context.Background(), "HEAD", "CI-payments-receipts-acde12"); err == nil {
		t.Fatal("nil provider resolved intent")
	}
	if _, err := (&Provider{}).Resolve(context.Background(), "HEAD", "CI-payments-receipts-acde12"); err == nil {
		t.Fatal("provider without reader resolved intent")
	}
	if _, err := p.Resolve(context.Background(), "", "CI-payments-receipts-acde12"); err == nil {
		t.Fatal("empty tree accepted")
	}
	if _, err := p.Resolve(context.Background(), "HEAD", "CI-missing-item-acde12"); err == nil {
		t.Fatal("missing intent resolved")
	}
}
func TestValidVectors(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: fixtures(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 || revs[0].Kind != core.KindChangeIntent || revs[1].Kind != core.KindRequirement {
		t.Fatalf("revisions=%+v", revs)
	}
}
func TestAuthorizationRemainsUndeclared(t *testing.T) {
	revs, err := provider(t, &fixtureReader{artifacts: fixtures(t)}).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var br intent.Requirement
	for _, r := range revs {
		if r.Kind == core.KindRequirement {
			_ = json.Unmarshal(r.Payload, &br)
		}
	}
	if br.AuthorizedBy == nil || len(br.AuthorizedBy) != 0 {
		t.Fatalf("authorization fabricated: %+v", br.AuthorizedBy)
	}
}
func TestHostileFixturesFailClosed(t *testing.T) {
	valid := fixtures(t)[0]
	for name, a := range map[string]Artifact{"utf8": {Path: valid.Path, Bytes: []byte{0xff}}, "path": {Path: "specs/../escape/requirements.md", Bytes: valid.Bytes}, "metadata": {Path: valid.Path, Bytes: []byte(strings.Replace(string(valid.Bytes), "Owner: payments-owner", "Priority: high\nOwner: payments-owner", 1))}, "duplicate": {Path: valid.Path, Bytes: []byte(strings.Replace(string(valid.Bytes), "- [O-2]", "- [O-1]", 1))}} {
		t.Run(name, func(t *testing.T) {
			if _, err := parse(a, time.Unix(1, 0)); err == nil {
				t.Fatal("hostile fixture accepted")
			}
		})
	}
}
func TestRequirementShapeFieldsFailIndependently(t *testing.T) {
	valid := fixtures(t)[0]
	for name, replacement := range map[string][2]string{
		"header":      {"# Requirement:", "# Requirements:"},
		"status":      {"Status: active", "State: active"},
		"owner":       {"Owner: payments-owner", "Maintainer: payments-owner"},
		"separator":   {"Owner: payments-owner\n\n## Obligations", "Owner: payments-owner\nnot blank\n## Obligations"},
		"obligations": {"## Obligations", "## Acceptance"},
	} {
		t.Run(name, func(t *testing.T) {
			a := valid
			a.Bytes = []byte(strings.Replace(string(valid.Bytes), replacement[0], replacement[1], 1))
			if _, err := parse(a, time.Unix(1, 0)); err == nil {
				t.Fatal("malformed requirement shape accepted")
			}
		})
	}
}
func TestChangeIntentShapeFieldsFailIndependently(t *testing.T) {
	valid := fixtures(t)[1]
	for name, replacement := range map[string][2]string{
		"header":    {"# Change Intent:", "# Change:"},
		"status":    {"Status: accepted", "State: accepted"},
		"sponsor":   {"Sponsor: payments-sponsor", "Owner: payments-sponsor"},
		"separator": {"Sponsor: payments-sponsor\n\n## Targets", "Sponsor: payments-sponsor\nnot blank\n## Targets"},
		"targets":   {"## Targets", "## Requirements"},
	} {
		t.Run(name, func(t *testing.T) {
			a := valid
			a.Bytes = []byte(strings.Replace(string(valid.Bytes), replacement[0], replacement[1], 1))
			if _, err := parse(a, time.Unix(1, 0)); err == nil {
				t.Fatal("malformed change-intent shape accepted")
			}
		})
	}
}
func TestDuplicateTargetFailsClosed(t *testing.T) {
	a := fixtures(t)[1]
	line := "- implements intent.specdir:BR-payments-flow-acde12@820ccdd89938e169b063b95955b417eb90e206214f6f9086ce9951a148f64d1e#O-1,O-2\n"
	a.Bytes = append(a.Bytes, []byte(line)...)
	if _, err := parse(a, time.Unix(1, 0)); err == nil {
		t.Fatal("duplicate target accepted")
	}
}
func TestDanglingTargetFailsClosed(t *testing.T) {
	artifacts := fixtures(t)
	if _, err := provider(t, &fixtureReader{artifacts: artifacts[1:]}).Revisions(context.Background()); err == nil {
		t.Fatal("change intent without its requirement accepted")
	}
}
func TestCanonicalizationStability(t *testing.T) {
	a := fixtures(t)[0]
	r1, err := parse(a, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	r2, err := parse(a, time.Unix(2, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r1.ArtifactSHA256 != r2.ArtifactSHA256 || string(r1.Payload) != string(r2.Payload) {
		t.Fatal("mapping is unstable")
	}
}
func TestResolveUsesCommitTree(t *testing.T) {
	reader := &fixtureReader{artifacts: fixtures(t)}
	got, err := provider(t, reader).Resolve(context.Background(), "deadbeef", "CI-payments-receipts-acde12")
	if err != nil || got.Kind != core.KindChangeIntent || reader.refs[0] != "deadbeef" {
		t.Fatalf("got=%+v refs=%v err=%v", got, reader.refs, err)
	}
}
func TestUnmappableFixtureStops(t *testing.T) {
	data, err := os.ReadFile("testdata/unmappable/conditional-obligation.md")
	if err != nil {
		t.Fatal(err)
	}
	_, err = parse(Artifact{Path: "specs/runtime/requirements.md", Bytes: data}, time.Unix(1, 0))
	if !errors.Is(err, intent.ErrUnmappable) {
		t.Fatalf("got %v", err)
	}
}
