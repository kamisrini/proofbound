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

type valueAppender struct{}

func (valueAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
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

func TestNewRequiresEveryDependency(t *testing.T) {
	provider := testProvider{name: "specdir"}
	ids := testIDs(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for name, deps := range map[string]*Deps{
		"nil deps":  nil,
		"providers": {IDs: ids, Logger: logger},
		"ids":       {Providers: []Provider{provider}, Logger: logger},
		"logger":    {Providers: []Provider{provider}, IDs: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(deps); err == nil {
				t.Fatal("incomplete dependencies accepted")
			}
		})
	}
}

func TestNewRejectsNilProviders(t *testing.T) {
	ids := testIDs(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var typedNil *testProvider
	for name, provider := range map[string]Provider{"nil interface": nil, "typed nil": typedNil} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(&Deps{Providers: []Provider{provider}, IDs: ids, Logger: logger}); err == nil {
				t.Fatal("nil provider accepted")
			}
		})
	}
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
func TestValidateRevisionRejectsMalformedChangeIntent(t *testing.T) {
	digest := strings.Repeat("c", 64)
	target := Reference{RecordKind: "requirement", Source: "intent.records", RecordID: "BR-sample-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "implements", ObligationIDs: []string{"O-1"}}
	value := ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: "CI-sample-change-acde12", Status: "accepted", DeclaredSponsor: "sponsor", Targets: []Reference{target}, ConstrainedBy: []Reference{}, Supersedes: []Reference{}, ArtifactPath: "docs/intent.md", ArtifactSHA256: digest}
	revision := Revision{Source: core.SourceIntentRecords, Kind: core.KindChangeIntent, NativeID: value.IntentID, ArtifactPath: value.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, value)}
	if err := ValidateRevision(revision); err != nil {
		t.Fatal(err)
	}
	value.DeclaredSponsor = ""
	revision.Payload = canonical(t, value)
	if ValidateRevision(revision) == nil {
		t.Fatal("empty sponsor accepted")
	}
	value.DeclaredSponsor = "sponsor"
	value.Targets = []Reference{}
	revision.Payload = canonical(t, value)
	if ValidateRevision(revision) == nil {
		t.Fatal("empty targets accepted")
	}
	value.Targets = []Reference{target}
	value.Targets[0].ArtifactSHA256 = "bad"
	revision.Payload = canonical(t, value)
	if ValidateRevision(revision) == nil {
		t.Fatal("malformed target accepted")
	}
}
func TestValidateRevisionRejectsBlankBusinessFields(t *testing.T) {
	digest := strings.Repeat("d", 64)
	value := BusinessDecision{Schema: "proofbound.business-decision.v1", DecisionID: "BD-sample-record-acde12", Status: "accepted", Outcome: "outcome", DeclaredOwner: "owner", DeclaredApprover: "approver", DecidedAt: "2026-09-09T00:00:00Z", Supersedes: []Reference{}, ArtifactPath: "docs/intent.md", ArtifactSHA256: digest}
	revision := Revision{Source: core.SourceIntentRecords, Kind: core.KindBusinessDecision, NativeID: value.DecisionID, ArtifactPath: value.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, value)}
	if err := ValidateRevision(revision); err != nil {
		t.Fatal(err)
	}
	value.Outcome = " "
	revision.Payload = canonical(t, value)
	if ValidateRevision(revision) == nil {
		t.Fatal("blank outcome accepted")
	}
	value.Outcome = "outcome"
	value.Supersedes = []Reference{{RecordKind: "business_decision", Source: "bad source", RecordID: "BD-prior-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "supersedes"}}
	revision.Payload = canonical(t, value)
	if ValidateRevision(revision) == nil {
		t.Fatal("malformed supersedes reference accepted")
	}
}
func TestBusinessDecisionClosedFieldsFailIndependently(t *testing.T) {
	digest := strings.Repeat("d", 64)
	base := BusinessDecision{Schema: "proofbound.business-decision.v1", DecisionID: "BD-sample-record-acde12", Status: "accepted", Outcome: "outcome", DeclaredOwner: "owner", DeclaredApprover: "approver", DecidedAt: "2026-09-09T00:00:00Z", Supersedes: []Reference{}, ArtifactPath: "docs/intent.md", ArtifactSHA256: digest}
	for name, mutate := range map[string]func(*BusinessDecision){
		"schema":   func(v *BusinessDecision) { v.Schema = "proofbound.business-decision.v2" },
		"status":   func(v *BusinessDecision) { v.Status = "maybe" },
		"outcome":  func(v *BusinessDecision) { v.Outcome = " " },
		"owner":    func(v *BusinessDecision) { v.DeclaredOwner = " " },
		"approver": func(v *BusinessDecision) { v.DeclaredApprover = " " },
	} {
		t.Run(name, func(t *testing.T) {
			value := base
			mutate(&value)
			revision := Revision{Source: core.SourceIntentRecords, Kind: core.KindBusinessDecision, NativeID: value.DecisionID, ArtifactPath: value.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, value)}
			if ValidateRevision(revision) == nil {
				t.Fatal("invalid business decision accepted")
			}
		})
	}
}
func TestRequirementClosedFieldsFailIndependently(t *testing.T) {
	digest := strings.Repeat("a", 64)
	base := Requirement{Schema: "proofbound.requirement.v1", RequirementID: "BR-sample-record-acde12", Status: "active", AuthorizedBy: []Reference{}, DeclaredOwner: "owner", Obligations: []Obligation{{ID: "O-1", Statement: "observable", State: "active"}}, Supersedes: []Reference{}, ArtifactPath: "docs/intent.md", ArtifactSHA256: digest}
	for name, mutate := range map[string]func(*Requirement){
		"schema":      func(v *Requirement) { v.Schema = "proofbound.requirement.v2" },
		"status":      func(v *Requirement) { v.Status = "maybe" },
		"owner":       func(v *Requirement) { v.DeclaredOwner = "" },
		"obligations": func(v *Requirement) { v.Obligations = nil },
	} {
		t.Run(name, func(t *testing.T) {
			value := base
			mutate(&value)
			revision := Revision{Source: core.SourceIntentRecords, Kind: core.KindRequirement, NativeID: value.RequirementID, ArtifactPath: value.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, value)}
			if ValidateRevision(revision) == nil {
				t.Fatal("invalid requirement accepted")
			}
		})
	}
}
func TestChangeIntentClosedFieldsFailIndependently(t *testing.T) {
	digest := strings.Repeat("c", 64)
	target := Reference{RecordKind: "requirement", Source: "intent.records", RecordID: "BR-sample-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "implements", ObligationIDs: []string{"O-1"}}
	base := ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: "CI-sample-change-acde12", Status: "accepted", DeclaredSponsor: "sponsor", Targets: []Reference{target}, ConstrainedBy: []Reference{}, Supersedes: []Reference{}, ArtifactPath: "docs/intent.md", ArtifactSHA256: digest}
	for name, mutate := range map[string]func(*ChangeIntent){
		"schema":  func(v *ChangeIntent) { v.Schema = "proofbound.change-intent.v2" },
		"status":  func(v *ChangeIntent) { v.Status = "maybe" },
		"sponsor": func(v *ChangeIntent) { v.DeclaredSponsor = "" },
		"targets": func(v *ChangeIntent) { v.Targets = nil },
	} {
		t.Run(name, func(t *testing.T) {
			value := base
			mutate(&value)
			revision := Revision{Source: core.SourceIntentRecords, Kind: core.KindChangeIntent, NativeID: value.IntentID, ArtifactPath: value.ArtifactPath, ArtifactSHA256: digest, ObservedAt: time.Unix(1, 0).UTC(), Payload: canonical(t, value)}
			if ValidateRevision(revision) == nil {
				t.Fatal("invalid change intent accepted")
			}
		})
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
func TestRevisionEnvelopeFieldsFailIndependently(t *testing.T) {
	base := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	for name, mutate := range map[string]func(*Revision){
		"native id": func(r *Revision) {
			var value Requirement
			_ = json.Unmarshal(r.Payload, &value)
			value.RequirementID, r.NativeID = "bad", "bad"
			r.Payload = canonical(t, value)
		},
		"digest": func(r *Revision) {
			var value Requirement
			_ = json.Unmarshal(r.Payload, &value)
			value.ArtifactSHA256, r.ArtifactSHA256 = "bad", "bad"
			r.Payload = canonical(t, value)
		},
		"path": func(r *Revision) {
			var value Requirement
			_ = json.Unmarshal(r.Payload, &value)
			value.ArtifactPath, r.ArtifactPath = "", ""
			r.Payload = canonical(t, value)
		},
		"observed at": func(r *Revision) { r.ObservedAt = time.Time{} },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := base
			mutate(&candidate)
			if ValidateRevision(candidate) == nil {
				t.Fatal("invalid revision envelope accepted")
			}
		})
	}
}
func TestRevisionEnvelopeBindsPayloadFieldsIndependently(t *testing.T) {
	base := sampleRevision(t, core.SourceIntentSpecdir, "BR-sample-record-acde12", strings.Repeat("a", 64))
	for name, mutate := range map[string]func(*Revision){
		"native id": func(r *Revision) { r.NativeID = "BR-other-record-acde12" },
		"path":      func(r *Revision) { r.ArtifactPath = "specs/other/requirements.md" },
		"digest":    func(r *Revision) { r.ArtifactSHA256 = strings.Repeat("b", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := base
			mutate(&candidate)
			if ValidateRevision(candidate) == nil {
				t.Fatal("payload/envelope mismatch accepted")
			}
		})
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
	r.ObligationIDs = []string{"bad"}
	if validateReferences([]Reference{r}) == nil {
		t.Fatal("invalid target obligation id accepted")
	}
	r.ObligationIDs = []string{"O-1"}
	r.Source = "intent.records.spoof"
	if validateReferences([]Reference{r}) == nil {
		t.Fatal("prefix spoof accepted")
	}
	for name, candidate := range map[string]Reference{
		"record kind": {RecordKind: "other", Source: "intent.records", RecordID: "BR-sample-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "supersedes"},
		"relation":    {RecordKind: "business_decision", Source: "intent.records", RecordID: "BD-sample-record-acde12", ArtifactSHA256: strings.Repeat("a", 64), Relation: "other"},
	} {
		t.Run(name, func(t *testing.T) {
			if validateReferences([]Reference{candidate}) == nil {
				t.Fatal("invalid reference vocabulary accepted")
			}
		})
	}
}
func TestObligationValidationFieldsFailIndependently(t *testing.T) {
	base := Obligation{ID: "O-1", Statement: "observable", State: "active"}
	for name, mutate := range map[string]func(*Obligation){
		"id":        func(o *Obligation) { o.ID = "bad" },
		"statement": func(o *Obligation) { o.Statement = " " },
		"state":     func(o *Obligation) { o.State = "maybe" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := base
			mutate(&candidate)
			if validateObligations([]Obligation{candidate}) == nil {
				t.Fatal("invalid obligation accepted")
			}
		})
	}
	if validateObligations([]Obligation{{ID: "O-2", Statement: "second", State: "active"}, base}) == nil {
		t.Fatal("out-of-order obligations accepted")
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

func TestSyncAcceptsNonPointerAppender(t *testing.T) {
	c := newTestConnector(t, testProvider{name: "specdir"})
	if _, err := c.Sync(context.Background(), "specdir", valueAppender{}); err != nil {
		t.Fatal(err)
	}
}
func TestSyncRejectsNilAppenders(t *testing.T) {
	c := newTestConnector(t, testProvider{name: "specdir"})
	var typedNil *testAppender
	for name, appender := range map[string]Appender{"nil interface": nil, "typed nil": typedNil} {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Sync(context.Background(), "specdir", appender); err == nil {
				t.Fatal("nil appender accepted")
			}
		})
	}
}
func TestSyncRejectsEveryUninitializedConnectorShape(t *testing.T) {
	ids := testIDs(t)
	provider := testProvider{name: "specdir"}
	for name, connector := range map[string]*Connector{
		"nil":          nil,
		"missing ids":  {providers: map[string]Provider{"specdir": provider}},
		"no providers": {providers: map[string]Provider{}, ids: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := connector.Sync(context.Background(), "specdir", valueAppender{}); err == nil {
				t.Fatal("uninitialized connector accepted")
			}
		})
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
