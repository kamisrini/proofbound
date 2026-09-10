package intent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type testProvider struct {
	name      string
	revisions []Revision
	err       error
}

func (p testProvider) Name() string { return p.name }
func (p testProvider) Revisions(context.Context) ([]Revision, error) {
	return append([]Revision(nil), p.revisions...), p.err
}

type testAppender struct{ events []core.Event }

func (a *testAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	for _, old := range a.events {
		if old.IdempotencyKey() == e.IdempotencyKey() {
			return store.Record{Event: old}, false, nil
		}
	}
	a.events = append(a.events, e)
	return store.Record{Event: e}, true, nil
}
func testIDs(t *testing.T) *core.IDGenerator {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{4}, 4096)), Now: func() time.Time { return time.Unix(2, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func canonical(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	out, err := core.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func sampleRevision(t *testing.T, source core.Source, id, digest string) Revision {
	t.Helper()
	v := Requirement{Schema: "proofbound.requirement.v1", RequirementID: id, Status: "active", AuthorizedBy: []Reference{}, DeclaredOwner: "owner", Obligations: []Obligation{{ID: "O-1", Statement: "observable", State: "active"}}, Supersedes: []Reference{}, ArtifactPath: "specs/x/requirements.md", ArtifactSHA256: digest}
	return Revision{Source: source, Kind: core.KindRequirement, NativeID: id, ArtifactPath: v.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, v)}
}
func newTestConnector(t *testing.T, providers ...Provider) *Connector {
	t.Helper()
	c, err := New(&Deps{Providers: providers, IDs: testIDs(t), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestValidateRevisionRegistry(t *testing.T) {
	r := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	if err := ValidateRevision(r); err != nil {
		t.Fatal(err)
	}
	r.Source = "intent.fake"
	if ValidateRevision(r) == nil {
		t.Fatal("unregistered provider source accepted")
	}
	r = sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	r.Kind = core.KindCheckRun
	if ValidateRevision(r) == nil {
		t.Fatal("non-intent kind accepted")
	}
}
func TestTwoDigestsAreIndependentlyRequired(t *testing.T) {
	r := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	r.ArtifactSHA256 = strings.Repeat("b", 64)
	if ValidateRevision(r) == nil {
		t.Fatal("artifact digest mismatch accepted")
	}
	r = sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	r.Payload = append(r.Payload, ' ')
	if ValidateRevision(r) == nil {
		t.Fatal("noncanonical payload accepted")
	}
}
func TestReferenceValidationFailsClosed(t *testing.T) {
	r := Reference{RecordKind: "requirement", Source: "intent.records", RecordID: "BR-sample-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "implements"}
	if validateReferences([]Reference{r}) == nil {
		t.Fatal("target without obligation accepted")
	}
	r.ObligationIDs = []string{"O-1"}
	if err := validateReferences([]Reference{r}); err != nil {
		t.Fatal(err)
	}
	r.Source = "intent.records.spoof"
	if validateReferences([]Reference{r}) == nil {
		t.Fatal("prefix spoof accepted")
	}
}
func TestSyncValidatesBeforeAppendAndOrdersProviders(t *testing.T) {
	a := sampleRevision(t, core.SourceIntentRecords, "BR-records-item-acde12", strings.Repeat("a", 64))
	b := sampleRevision(t, core.SourceIntentSpecdir, "BR-specdir-item-acde12", strings.Repeat("b", 64))
	bad := b
	bad.ArtifactSHA256 = "bad"
	app := &testAppender{}
	c := newTestConnector(t, testProvider{name: "specdir", revisions: []Revision{bad}}, testProvider{name: "records", revisions: []Revision{a}})
	if _, err := c.Sync(context.Background(), "all", app); err == nil || len(app.events) != 0 {
		t.Fatalf("batch was partially appended: %d %v", len(app.events), err)
	}
	c = newTestConnector(t, testProvider{name: "specdir", revisions: []Revision{b}}, testProvider{name: "records", revisions: []Revision{a}})
	if _, err := c.Sync(context.Background(), "all", app); err != nil {
		t.Fatal(err)
	}
	if app.events[0].Source != core.SourceIntentRecords {
		t.Fatalf("order=%v", app.events)
	}
}
func TestSyncIdempotenceAndRevision(t *testing.T) {
	r := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	p := testProvider{name: "specdir", revisions: []Revision{r}}
	c := newTestConnector(t, p)
	app := &testAppender{}
	first, err := c.Sync(context.Background(), "specdir", app)
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.Sync(context.Background(), "specdir", app)
	if err != nil || first.Appended != 1 || second.Existing != 1 {
		t.Fatalf("first=%+v second=%+v err=%v", first, second, err)
	}
	r2 := sampleRevision(t, core.SourceIntentSpecdir, r.NativeID, strings.Repeat("b", 64))
	p.revisions = []Revision{r2}
	c = newTestConnector(t, p)
	third, err := c.Sync(context.Background(), "specdir", app)
	if err != nil || third.Appended != 1 {
		t.Fatalf("third=%+v err=%v", third, err)
	}
}
func TestSyntheticProviderIsTestOnly(t *testing.T) {
	_, err := New(&Deps{Providers: []Provider{testProvider{name: "synthetic"}}, IDs: testIDs(t), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err == nil {
		t.Fatal("synthetic provider accepted by production constructor")
	}
}
