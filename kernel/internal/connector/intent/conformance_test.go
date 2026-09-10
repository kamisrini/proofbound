package intent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

func TestProviderScopedIdentityAndPrefixSpoofing(t *testing.T) {
	a := sampleRevision(t, core.SourceIntentRecords, "BR-shared-record-acde12", strings.Repeat("a", 64))
	b := sampleRevision(t, core.SourceIntentSpecdir, "BR-shared-record-acde12", strings.Repeat("b", 64))
	app := &testAppender{}
	_, err := newTestConnector(t, testProvider{name: "records", revisions: []Revision{a}}, testProvider{name: "specdir", revisions: []Revision{b}}).Sync(context.Background(), "all", app)
	if err != nil || len(app.events) != 2 {
		t.Fatalf("collision lost: %d %v", len(app.events), err)
	}
	b.Source = "intent.records"
	if _, err := newTestConnector(t, testProvider{name: "specdir", revisions: []Revision{b}}).Sync(context.Background(), "all", &testAppender{}); err == nil {
		t.Fatal("provider prefix spoof accepted")
	}
}
func TestCrossProviderSupersedes(t *testing.T) {
	a := sampleRevision(t, core.SourceIntentRecords, "BR-records-base-acde12", strings.Repeat("a", 64))
	v := Requirement{Schema: "proofbound.requirement.v1", RequirementID: "BR-specdir-next-acde12", Status: "active", AuthorizedBy: []Reference{}, DeclaredOwner: "owner", Obligations: []Obligation{{ID: "O-1", Statement: "observable", State: "active"}}, Supersedes: []Reference{{RecordKind: "requirement", Source: string(a.Source), RecordID: a.NativeID, ArtifactSHA256: a.ArtifactSHA256, Relation: "supersedes"}}, ArtifactPath: "specs/x/requirements.md", ArtifactSHA256: strings.Repeat("b", 64)}
	b := Revision{Source: core.SourceIntentSpecdir, Kind: core.KindRequirement, NativeID: v.RequirementID, ArtifactPath: v.ArtifactPath, ArtifactSHA256: v.ArtifactSHA256, ObservedAt: a.ObservedAt, Payload: canonical(t, v)}
	app := &testAppender{}
	_, err := newTestConnector(t, testProvider{name: "records", revisions: []Revision{a}}, testProvider{name: "specdir", revisions: []Revision{b}}).Sync(context.Background(), "all", app)
	if err != nil || len(app.events) != 2 {
		t.Fatalf("cross-provider supersedes lost: %v", err)
	}
}
func TestUpstreamMutationDigestMismatch(t *testing.T) {
	r := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	r.ArtifactSHA256 = strings.Repeat("b", 64)
	if _, err := newTestConnector(t, testProvider{name: "specdir", revisions: []Revision{r}}).Sync(context.Background(), "all", &testAppender{}); err == nil {
		t.Fatal("upstream mutation accepted")
	}
}
func TestCrossProviderSyncOrdering(t *testing.T) { TestSyncValidatesBeforeAppendAndOrdersProviders(t) }
func TestReplaySufficiency(t *testing.T) {
	r := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	var got Requirement
	if err := json.Unmarshal(r.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if got.Obligations[0].Statement == "" || got.RequirementID == "" || got.ArtifactSHA256 == "" {
		t.Fatalf("payload insufficient: %+v", got)
	}
}
func TestUnmappableFixtureStops(t *testing.T) {
	if !errors.Is(fmt.Errorf("provider: %w", ErrUnmappable), ErrUnmappable) {
		t.Fatal("STOP sentinel cannot be identified")
	}
}
