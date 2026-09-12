//go:build integration

package projections

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type failWriteAt struct {
	at, calls int
	err       error
}

func (w *failWriteAt) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.at {
		return 0, w.err
	}
	return len(p), nil
}

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

func TestCommitIntentInsertFailureStopsProjection(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, false)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `
CREATE OR REPLACE FUNCTION reject_commit_intent() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'forced commit intent failure'; END
$$;
CREATE TRIGGER reject_commit_intent BEFORE INSERT ON commit_intents_view
FOR EACH ROW EXECUTE FUNCTION reject_commit_intent()`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	sha := shaFor("failed-intent-commit")
	payload := []byte(`{"sha":"` + sha + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:00:00Z","subject":"intent","files_touched":null,"cited_decisions":null,"intent_refs":[{"provider":"records","record_id":"` + testCI + `","artifact_sha256":"` + strings.Repeat("c", 64) + `"}]}`)
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, sha, payload, 4)
	if err := p.Apply(context.Background(), s); err == nil || !strings.Contains(err.Error(), "forced commit intent failure") {
		t.Fatalf("commit intent insert error=%v", err)
	}
}

func TestSupersedingBusinessDecisionIsProjectedAfterExactReferenceCheck(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, false)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	const nextID = "BD-projection-next-acde12"
	next := connectorintent.BusinessDecision{
		Schema: "proofbound.business-decision.v1", DecisionID: nextID, Status: "accepted",
		Outcome: "supersede", DeclaredOwner: "owner", DeclaredApprover: "approver",
		DecidedAt:    "2026-09-10T00:00:00Z",
		Supersedes:   []connectorintent.Reference{{RecordKind: "business_decision", Source: "intent.records", RecordID: testBD, ArtifactSHA256: strings.Repeat("a", 64), Relation: "supersedes"}},
		ArtifactPath: "docs/intent/records/business-decisions/" + nextID + "/0001.md", ArtifactSHA256: strings.Repeat("d", 64),
	}
	raw, err := json.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	appendRaw(t, s, core.SourceIntentRecords, core.KindBusinessDecision, nextID, canonical, 4)
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM business_decisions_view WHERE decision_id=$1`, nextID).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("superseding decision rows=%d", count)
	}
}

func TestRequirementProjectsEveryObligation(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, false)
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	const requirementID = "BR-projection-two-acde12"
	requirement := connectorintent.Requirement{
		Schema: "proofbound.requirement.v1", RequirementID: requirementID, Status: "active",
		AuthorizedBy:  []connectorintent.Reference{{RecordKind: "business_decision", Source: "intent.records", RecordID: testBD, ArtifactSHA256: strings.Repeat("a", 64), Relation: "authorizes"}},
		DeclaredOwner: "owner",
		Obligations:   []connectorintent.Obligation{{ID: "O-1", Statement: "first outcome", State: "active"}, {ID: "O-2", Statement: "second outcome", State: "active"}},
		Supersedes:    []connectorintent.Reference{},
		ArtifactPath:  "docs/intent/records/requirements/" + requirementID + "/0001.md", ArtifactSHA256: strings.Repeat("d", 64),
	}
	raw, err := json.Marshal(requirement)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	appendRaw(t, s, core.SourceIntentRecords, core.KindRequirement, requirementID, canonical, 4)
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM requirement_obligations_view WHERE requirement_id=$1`, requirementID).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("obligation rows=%d", count)
	}
	const intentID = "CI-projection-two-acde12"
	intent := connectorintent.ChangeIntent{
		Schema: "proofbound.change-intent.v1", IntentID: intentID, Status: "accepted", DeclaredSponsor: "sponsor",
		Targets:       []connectorintent.Reference{{RecordKind: "requirement", Source: "intent.records", RecordID: requirementID, ArtifactSHA256: strings.Repeat("d", 64), Relation: "implements", ObligationIDs: []string{"O-1", "O-2"}}},
		ConstrainedBy: []connectorintent.Reference{}, Supersedes: []connectorintent.Reference{},
		ArtifactPath: "docs/intent/records/change-intents/" + intentID + "/0001.md", ArtifactSHA256: strings.Repeat("e", 64),
	}
	raw, err = json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err = core.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	appendRaw(t, s, core.SourceIntentRecords, core.KindChangeIntent, intentID, canonical, 5)
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM intent_targets_view WHERE intent_id=$1`, intentID).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("intent target rows=%d", count)
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

func TestExactRecordLookupPreservesDatabaseFailure(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	if err := New().Ensure(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `DROP TABLE requirements_view`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	ref := connectorintent.Reference{RecordKind: "requirement", Source: "intent.records", RecordID: testBR, ArtifactSHA256: strings.Repeat("b", 64), Relation: "implements"}
	err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return requireExactRecord(ctx, tx, ref)
	})
	if err == nil || !strings.Contains(err.Error(), "requirements_view") || strings.Contains(err.Error(), "dangling exact record") {
		t.Fatalf("lookup error=%v", err)
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

func TestIntentReportsPropagateEveryOutputFailure(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	appendDeployment(t, s, 1, "production", shaFor("intent-commit"))
	p := New()
	want := errors.New("write failed")
	for at := 1; at <= 5; at++ {
		writer := &failWriteAt{at: at, err: want}
		if err := p.ReportIntent(context.Background(), s, testCI, time.Now(), writer); !errors.Is(err, want) {
			t.Fatalf("intent write %d error=%v calls=%d", at, err, writer.calls)
		}
	}
	writer := &failWriteAt{at: 1, err: want}
	if err := p.ReportRequirement(context.Background(), s, testBR, writer); !errors.Is(err, want) {
		t.Fatalf("requirement write error=%v", err)
	}
	writer = &failWriteAt{at: 1, err: want}
	if err := p.CheckIntent(context.Background(), s, shaFor("intent-commit"), writer); !errors.Is(err, want) {
		t.Fatalf("intent check write error=%v", err)
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

func TestRequirementReportRendersAuthorityAndFailsClosedWithoutProof(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	records := intentFixtureEvents(t, s, false)
	p := New()
	var out bytes.Buffer
	if err := p.ReportRequirement(context.Background(), s, testBR, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"requirement=" + testBR, "authorization=declared", "artifact_sha256=" + strings.Repeat("b", 64), "proof=" + records[1].Event.ID.String()} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, out.String())
		}
	}
	var missing bytes.Buffer
	if err := p.ReportRequirement(context.Background(), s, "BR-missing-item-acde12", &missing); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("missing requirement output=%q err=%v", missing.String(), err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE requirements_view SET event_id='01ARZ3NDEKTSV4RRFFQ69G5FAV'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := p.ReportRequirement(context.Background(), s, testBR, &out); err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("output=%s err=%v", out.String(), err)
	}
}

func TestIntentCheckRendersExactClaimAndFailsClosedWithoutProof(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	records := intentFixtureEvents(t, s, true)
	sha := shaFor("intent-commit")
	p := New()
	var out bytes.Buffer
	if err := p.CheckIntent(context.Background(), s, sha, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"commit=" + sha, "intent=records:" + testCI + "@" + strings.Repeat("c", 64), "proof=" + records[3].Event.ID.String()} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "intent=none") {
		t.Fatalf("claimed commit also rendered as unclaimed: %s", out.String())
	}
	var none bytes.Buffer
	if err := p.CheckIntent(context.Background(), s, shaFor("no-intent"), &none); err != nil || none.String() != "commit="+shaFor("no-intent")+" intent=none\n" {
		t.Fatalf("unclaimed output=%q err=%v", none.String(), err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE commit_intents_view SET event_id='01ARZ3NDEKTSV4RRFFQ69G5FAV'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := p.CheckIntent(context.Background(), s, sha, &out); err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("output=%s err=%v", out.String(), err)
	}
}

func TestIntentReportKeepsMultipleCommitsAndDeploymentsDistinct(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	firstSHA := shaFor("intent-commit")
	secondSHA := shaFor("intent-commit-second")
	payload := []byte(`{"sha":"` + secondSHA + `","author_name":"A","author_email":"a@example.test","committer_name":"C","committer_email":"c@example.test","committed_at":"2026-09-10T00:01:00Z","subject":"intent second","files_touched":null,"cited_decisions":null,"intent_refs":[{"provider":"records","record_id":"` + testCI + `","artifact_sha256":"` + strings.Repeat("c", 64) + `"}]}`)
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, secondSHA, payload, 5)
	appendDeployment(t, s, 1, "staging", firstSHA)
	appendDeployment(t, s, 2, "production", secondSHA)
	var out bytes.Buffer
	if err := New().ReportIntent(context.Background(), s, testCI, time.Now(), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"commit=" + firstSHA, "staging:observed", "commit=" + secondSHA, "production:observed"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, out.String())
		}
	}
}

func TestSpecStateUsesWorstReviewRegardlessOfOrder(t *testing.T) {
	for name, tc := range map[string]struct {
		reviews []string
		want    string
	}{
		"verifiable":              {[]string{"VERIFIABLE"}, "verifiable"},
		"unreviewed":              {[]string{"VERIFIABLE", "UNREVIEWED"}, "unreviewed"},
		"ambiguous after missing": {[]string{"UNREVIEWED", "AMBIGUOUS"}, "ambiguous"},
		"missing after ambiguous": {[]string{"AMBIGUOUS", "UNREVIEWED"}, "ambiguous"},
		"untestable":              {[]string{"AMBIGUOUS", "UNTESTABLE", "UNREVIEWED"}, "untestable"},
		"contradictory":           {[]string{"UNTESTABLE", "CONTRADICTORY", "AMBIGUOUS"}, "contradictory"},
	} {
		t.Run(name, func(t *testing.T) {
			targets := make([]intentTargetReport, len(tc.reviews))
			for i, review := range tc.reviews {
				targets[i].review = review
			}
			if got := specState(targets); got != tc.want {
				t.Fatalf("state=%s reviews=%v", got, tc.reviews)
			}
		})
	}
}

func TestDerivedIntentStatesRequireBothSatisfactionAndReview(t *testing.T) {
	revision := intentReportRevision{status: "accepted"}
	commit := intentCommitReport{sha: shaFor("state")}
	deployed := commit
	deployed.deployments = []string{"production:observed:fresh"}
	for name, tc := range map[string]struct {
		revision intentReportRevision
		targets  []intentTargetReport
		commits  []intentCommitReport
		want     string
	}{
		"declared":             {revision, nil, nil, "DECLARED"},
		"implemented empty":    {revision, nil, []intentCommitReport{commit}, "IMPLEMENTED_UNVERIFIED"},
		"satisfied reviewed":   {revision, []intentTargetReport{{outcome: "SATISFIED", review: "VERIFIABLE"}}, []intentCommitReport{commit}, "SATISFIED"},
		"satisfied unreviewed": {revision, []intentTargetReport{{outcome: "SATISFIED", review: "UNREVIEWED"}}, []intentCommitReport{commit}, "IMPLEMENTED_UNVERIFIED"},
		"unverified reviewed":  {revision, []intentTargetReport{{outcome: "UNVERIFIED", review: "VERIFIABLE"}}, []intentCommitReport{commit}, "IMPLEMENTED_UNVERIFIED"},
		"not satisfied":        {revision, []intentTargetReport{{outcome: "NOT_SATISFIED", review: "VERIFIABLE"}}, []intentCommitReport{commit}, "NOT_SATISFIED"},
		"inconclusive":         {revision, []intentTargetReport{{outcome: "INCONCLUSIVE", review: "VERIFIABLE"}}, []intentCommitReport{commit}, "INCONCLUSIVE"},
		"superseded":           {intentReportRevision{status: "superseded"}, nil, []intentCommitReport{commit}, "SUPERSEDED"},
		"withdrawn":            {intentReportRevision{status: "withdrawn"}, nil, []intentCommitReport{commit}, "SUPERSEDED"},
		"deployed verified":    {revision, []intentTargetReport{{outcome: "SATISFIED", review: "VERIFIABLE"}}, []intentCommitReport{deployed}, "DEPLOYED_VERIFIED"},
		"deployed unreviewed":  {revision, []intentTargetReport{{outcome: "SATISFIED", review: "UNREVIEWED"}}, []intentCommitReport{deployed}, "DEPLOYED_UNVERIFIED"},
		"one deployed commit":  {revision, []intentTargetReport{{outcome: "UNVERIFIED", review: "UNREVIEWED"}}, []intentCommitReport{commit, deployed}, "DEPLOYED_UNVERIFIED"},
	} {
		t.Run(name, func(t *testing.T) {
			state, _ := deriveIntentState(tc.revision, tc.targets, tc.commits)
			if state != tc.want {
				t.Fatalf("state=%s want=%s", state, tc.want)
			}
		})
	}
}

func TestIntentReportEntryPointsRejectEachInvalidArgument(t *testing.T) {
	p := New()
	ctx := context.Background()
	var output bytes.Buffer
	if err := p.ReportIntent(ctx, nil, "CI-valid-item-acde12", time.Time{}, nil); err == nil || err.Error() != "intent report: id and output are required" {
		t.Fatalf("intent report nil-output error=%v", err)
	}
	if err := p.ReportIntent(ctx, nil, " \t", time.Time{}, &output); err == nil || err.Error() != "intent report: id and output are required" {
		t.Fatalf("intent report blank-id error=%v", err)
	}
	if err := p.ReportRequirement(ctx, nil, "BR-valid-item-acde12", nil); err == nil || err.Error() != "requirement report: id and output are required" {
		t.Fatalf("requirement report nil-output error=%v", err)
	}
	if err := p.ReportRequirement(ctx, nil, "", &output); err == nil || err.Error() != "requirement report: id and output are required" {
		t.Fatalf("requirement report empty-id error=%v", err)
	}
	if err := p.CheckIntent(ctx, nil, strings.Repeat("a", 40), nil); err == nil || err.Error() != "intent check: valid commit and output are required" {
		t.Fatalf("intent check nil-output error=%v", err)
	}
	if err := p.CheckIntent(ctx, nil, "bad", &output); err == nil || err.Error() != "intent check: valid commit and output are required" {
		t.Fatalf("intent check bad-commit error=%v", err)
	}
}
