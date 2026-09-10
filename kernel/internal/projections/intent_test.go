//go:build integration

package projections

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const (
	testBD = "BD-projection-chain-acde12"
	testBR = "BR-projection-chain-acde12"
	testCI = "CI-projection-chain-acde12"
)

func intentFixtureEvents(t *testing.T, s *store.Store, withCommit bool) []store.Record {
	t.Helper()
	bdDigest, brDigest, ciDigest := strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)
	bd := connectorintent.BusinessDecision{
		Schema: "proofbound.business-decision.v1", DecisionID: testBD, Status: "accepted",
		Outcome: "ship", DeclaredOwner: "owner", DeclaredApprover: "approver",
		DecidedAt: "2026-09-09T00:00:00Z", Supersedes: []connectorintent.Reference{},
		ArtifactPath: "docs/intent/records/business-decisions/" + testBD + "/0001.md", ArtifactSHA256: bdDigest,
	}
	br := connectorintent.Requirement{
		Schema: "proofbound.requirement.v1", RequirementID: testBR, Status: "active",
		AuthorizedBy:  []connectorintent.Reference{{RecordKind: "business_decision", Source: "intent.records", RecordID: testBD, ArtifactSHA256: bdDigest, Relation: "authorizes"}},
		DeclaredOwner: "owner", Obligations: []connectorintent.Obligation{{ID: "O-1", Statement: "observable outcome", State: "active"}},
		Supersedes: []connectorintent.Reference{}, ArtifactPath: "docs/intent/records/requirements/" + testBR + "/0001.md", ArtifactSHA256: brDigest,
	}
	ci := connectorintent.ChangeIntent{
		Schema: "proofbound.change-intent.v1", IntentID: testCI, Status: "accepted", DeclaredSponsor: "sponsor",
		Targets:       []connectorintent.Reference{{RecordKind: "requirement", Source: "intent.records", RecordID: testBR, ArtifactSHA256: brDigest, Relation: "implements", ObligationIDs: []string{"O-1"}}},
		ConstrainedBy: []connectorintent.Reference{}, Supersedes: []connectorintent.Reference{},
		ArtifactPath: "docs/intent/records/change-intents/" + testCI + "/0001.md", ArtifactSHA256: ciDigest,
	}
	marshal := func(v any) []byte {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := core.Canonicalize(raw)
		if err != nil {
			t.Fatal(err)
		}
		return canonical
	}
	records := []store.Record{
		appendRaw(t, s, core.SourceIntentRecords, core.KindBusinessDecision, testBD, marshal(bd), 1),
		appendRaw(t, s, core.SourceIntentRecords, core.KindRequirement, testBR, marshal(br), 2),
		appendRaw(t, s, core.SourceIntentRecords, core.KindChangeIntent, testCI, marshal(ci), 3),
	}
	if withCommit {
		sha := shaFor("intent-commit")
		payload := []byte(`{"sha":"` + sha + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:00:00Z","subject":"intent","files_touched":null,"cited_decisions":null,"intent_refs":[{"provider":"records","record_id":"` + testCI + `","artifact_sha256":"` + ciDigest + `"}]}`)
		records = append(records, appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, payload, 4))
	}
	return records
}

func TestIntentRowsRetainExactRevisionAndProof(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	records := intentFixtureEvents(t, s, true)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{`SELECT event_id,seq FROM business_decisions_view`, `SELECT event_id,seq FROM requirements_view`, `SELECT event_id,seq FROM requirement_obligations_view`, `SELECT event_id,seq FROM change_intents_view`, `SELECT event_id,seq FROM intent_targets_view`, `SELECT event_id,seq FROM commit_intents_view`} {
		var id string
		var seq int64
		if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error { return tx.QueryRow(ctx, query).Scan(&id, &seq) }); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, r := range records {
			found = found || (r.Event.ID.String() == id && r.Seq == seq)
		}
		if !found {
			t.Fatalf("unbound proof %s/%d", id, seq)
		}
	}
}
func TestIntentProjectionRejectsDanglingProof(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	ci := connectorintent.ChangeIntent{
		Schema: "proofbound.change-intent.v1", IntentID: testCI, Status: "accepted", DeclaredSponsor: "sponsor",
		Targets:       []connectorintent.Reference{{RecordKind: "requirement", Source: "intent.records", RecordID: testBR, ArtifactSHA256: strings.Repeat("b", 64), Relation: "implements", ObligationIDs: []string{"O-1"}}},
		ConstrainedBy: []connectorintent.Reference{}, Supersedes: []connectorintent.Reference{},
		ArtifactPath: "docs/intent/records/change-intents/" + testCI + "/0001.md", ArtifactSHA256: strings.Repeat("c", 64),
	}
	raw, _ := json.Marshal(ci)
	canonical, _ := core.Canonicalize(raw)
	appendRaw(t, s, core.SourceIntentRecords, core.KindChangeIntent, testCI, canonical, 1)
	if err := New().Apply(context.Background(), s); err == nil {
		t.Fatal("dangling target accepted")
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM change_intents_view`).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partial projection committed")
	}
}
func TestIntentProjectionRebuildMatchesIncremental(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	before, err := p.Snapshot(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Rebuild(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	after, err := p.Snapshot(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if err := CompareSnapshots(before, after); err != nil {
		t.Fatal(err)
	}
}
func TestIntentReportRendersComponentStates(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	var out bytes.Buffer
	if err := New().ReportIntent(context.Background(), s, testCI, time.Now(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"state=IMPLEMENTED_UNVERIFIED", "spec=unreviewed", "review=UNREVIEWED", "outcome=UNVERIFIED", "proof="} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, out.String())
		}
	}
}
func TestIntentReportProofFailsClosed(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, false)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE intent_targets_view SET event_id='01ARZ3NDEKTSV4RRFFQ69G5FAV'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.ReportIntent(context.Background(), s, testCI, time.Now(), &out); err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("output=%s err=%v", out.String(), err)
	}
}
