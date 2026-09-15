//go:build integration

package projections

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func reviewEvent(t *testing.T, kind core.Kind, native string, payload any) core.Event {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	event, err := g.NewEvent(core.NewEventParams{Source: core.SourceReviews, NativeID: native, Kind: kind, OccurredAt: time.Now(), Payload: raw, ConnectorVersion: "reviews/1"})
	if err != nil {
		t.Fatal(err)
	}
	return event
}
func evidenceEvent(t *testing.T) core.Event {
	t.Helper()
	payload := []byte(`{"session_id":"evidence","started_at":null,"finished_at":null,"message_count":1,"tool_call_count":0,"files_written_count":0,"parse_coverage":1}`)
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	event, err := g.NewEvent(core.NewEventParams{Source: core.SourceSessions, NativeID: "evidence", Kind: core.KindSessionObserved, OccurredAt: time.Now(), Payload: payload, ConnectorVersion: "sessions/1"})
	if err != nil {
		t.Fatal(err)
	}
	return event
}
func appendEvents(t *testing.T, s *store.Store, events ...core.Event) {
	t.Helper()
	sync, err := s.BeginSync(context.Background(), "review-fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if _, _, err := sync.Append(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}
	if err := sync.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}
func requirementReviewFixture(reviewer, digest, obligation string) connectorreviews.RequirementReview {
	return connectorreviews.RequirementReview{Schema: "proofbound.requirement-review.v1", ReviewID: "RR-projection-round-acde12", Requirement: connectorintent.Reference{RecordKind: "requirement", Source: "intent.records", RecordID: testBR, ArtifactSHA256: digest, Relation: "reviews"}, DeclaredReviewer: reviewer, Outcomes: []connectorreviews.RequirementReviewOutcome{{ObligationID: obligation, Outcome: "VERIFIABLE"}}, ArtifactPath: "docs/verification/verdicts/rr.md", ArtifactSHA256: strings.Repeat("d", 64)}
}
func obligationVerdictFixture(commit, evidence string) connectorreviews.ObligationVerdict {
	return connectorreviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", VerdictID: "OV-projection-round-acde12", Status: "ACCEPTABLE", DeclaredReviewer: "independent", ReviewedCommit: commit, ChangeIntent: connectorintent.Reference{RecordKind: "change_intent", Source: "intent.records", RecordID: testCI, ArtifactSHA256: strings.Repeat("c", 64), Relation: "evaluates"}, Requirements: []connectorintent.Reference{{RecordKind: "requirement", Source: "intent.records", RecordID: testBR, ArtifactSHA256: strings.Repeat("b", 64), Relation: "evaluates"}}, Obligations: []connectorreviews.ObligationOutcome{{Source: "intent.records", RequirementID: testBR, ArtifactSHA256: strings.Repeat("b", 64), ObligationID: "O-1", Outcome: "SATISFIED", EvidenceEventIDs: []string{evidence}}}, Findings: []connectorreviews.Finding{}, ArtifactPath: "docs/verification/verdicts/ov.md", ArtifactSHA256: strings.Repeat("e", 64)}
}

func TestObligationVerdictProjectionValidation(t *testing.T) {
	t.Run("valid exact chain", func(t *testing.T) {
		s := testStore(t)
		defer s.Close()
		intentFixtureEvents(t, s, true)
		evidence := evidenceEvent(t)
		review := requirementReviewFixture("independent", strings.Repeat("b", 64), "O-1")
		verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
		appendEvents(t, s, evidence, reviewEvent(t, core.KindRequirementReview, review.ReviewID, review), reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
		if err := New().Apply(context.Background(), s); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM obligation_verdicts_view`).Scan(&count)
		}); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("rows=%d", count)
		}
		var out bytes.Buffer
		if err := New().ReportIntent(context.Background(), s, testCI, time.Now(), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "state=SATISFIED") || !strings.Contains(out.String(), "spec=verifiable") {
			t.Fatalf("report=%s", out.String())
		}
	})
	for _, tc := range []struct {
		name   string
		mutate func(*connectorreviews.ObligationVerdict)
	}{
		{"missing requirement reference", func(v *connectorreviews.ObligationVerdict) { v.Requirements = nil }},
		{"missing obligation target", func(v *connectorreviews.ObligationVerdict) { v.Obligations[0].ObligationID = "O-404" }},
	} {
		t.Run("rejects incomplete exact chain "+tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, true)
			evidence := evidenceEvent(t)
			appendEvents(t, s, evidence)
			verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
			tc.mutate(&verdict)
			appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
			if err := New().Apply(context.Background(), s); err == nil {
				t.Fatal("incomplete v2 exact chain accepted")
			}
		})
	}
	t.Run("future evidence", func(t *testing.T) {
		s := testStore(t)
		defer s.Close()
		intentFixtureEvents(t, s, true)
		evidence := evidenceEvent(t)
		verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
		appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict), evidence)
		if err := New().Apply(context.Background(), s); err == nil || !strings.Contains(err.Error(), "future") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("wrong commit", func(t *testing.T) {
		s := testStore(t)
		defer s.Close()
		intentFixtureEvents(t, s, true)
		evidence := evidenceEvent(t)
		appendEvents(t, s, evidence)
		verdict := obligationVerdictFixture(shaFor("other"), evidence.ID.String())
		appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
		if err := New().Apply(context.Background(), s); err == nil {
			t.Fatal("wrong commit accepted")
		}
	})
	for _, tc := range []struct {
		name    string
		status  string
		wantErr bool
	}{{"acceptable aggregate mismatch", "ACCEPTABLE", true}, {"needs work preserves outcome", "NEEDS_WORK", false}} {
		t.Run(tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, true)
			evidence := evidenceEvent(t)
			appendEvents(t, s, evidence)
			verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
			verdict.Status = tc.status
			verdict.Obligations[0].Outcome = "INCONCLUSIVE"
			verdict.Obligations[0].EvidenceEventIDs = nil
			appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
			err := New().Apply(context.Background(), s)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v wantErr=%t", err, tc.wantErr)
			}
		})
	}
	for _, tc := range []struct {
		name    string
		outcome string
		status  string
		wantErr bool
	}{
		{"satisfied", "SATISFIED", "ACCEPTABLE", false},
		{"not satisfied", "NOT_SATISFIED", "NEEDS_WORK", false},
		{"inconclusive", "INCONCLUSIVE", "NEEDS_WORK", false},
		{"unknown", "UNKNOWN", "NEEDS_WORK", true},
	} {
		t.Run("outcome domain "+tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, true)
			evidence := evidenceEvent(t)
			appendEvents(t, s, evidence)
			verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
			verdict.Status = tc.status
			verdict.Obligations[0].Outcome = tc.outcome
			appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
			if err := New().Apply(context.Background(), s); (err != nil) != tc.wantErr {
				t.Fatalf("error=%v wantErr=%t", err, tc.wantErr)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*connectorreviews.ObligationVerdict)
	}{
		{"schema", func(v *connectorreviews.ObligationVerdict) { v.Schema = "proofbound.obligation-verdict.v1" }},
		{"verdict id", func(v *connectorreviews.ObligationVerdict) { v.VerdictID = "other" }},
		{"reviewed commit", func(v *connectorreviews.ObligationVerdict) { v.ReviewedCommit = "not-a-commit" }},
		{"status", func(v *connectorreviews.ObligationVerdict) { v.Status = "UNKNOWN" }},
		{"reviewer", func(v *connectorreviews.ObligationVerdict) { v.DeclaredReviewer = "" }},
		{"artifact digest", func(v *connectorreviews.ObligationVerdict) { v.ArtifactSHA256 = "not-a-digest" }},
	} {
		t.Run("invalid envelope "+tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, true)
			evidence := evidenceEvent(t)
			appendEvents(t, s, evidence)
			verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
			nativeID := verdict.VerdictID
			tc.mutate(&verdict)
			appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, nativeID, verdict))
			if err := New().Apply(context.Background(), s); err == nil {
				t.Fatal("invalid v2 envelope accepted")
			}
		})
	}
}

func TestObligationVerdictProjectionPropagatesFindingInsertError(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	evidence := evidenceEvent(t)
	review := requirementReviewFixture("independent", strings.Repeat("b", 64), "O-1")
	appendEvents(t, s, evidence, reviewEvent(t, core.KindRequirementReview, review.ReviewID, review))
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `ALTER TABLE reviews_view ADD CONSTRAINT p6_finding_insert_failure CHECK (false)`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	verdict := obligationVerdictFixture(shaFor("intent-commit"), evidence.ID.String())
	verdict.Status = "NEEDS_WORK"
	verdict.Findings = []connectorreviews.Finding{{FindingID: "F-1", Severity: "MED"}}
	appendEvents(t, s, reviewEvent(t, core.KindReviewVerdict, verdict.VerdictID, verdict))
	if err := New().Apply(context.Background(), s); err == nil {
		t.Fatal("finding insert error was swallowed")
	}
}

func TestRequirementReviewProjectionValidation(t *testing.T) {
	for _, tc := range []struct{ name, reviewer, digest, obligation, outcome string }{{"author equals reviewer", "owner", strings.Repeat("b", 64), "O-1", "VERIFIABLE"}, {"wrong revision", "independent", strings.Repeat("f", 64), "O-1", "VERIFIABLE"}, {"absent obligation", "independent", strings.Repeat("b", 64), "O-404", "VERIFIABLE"}, {"unknown outcome", "independent", strings.Repeat("b", 64), "O-1", "MAYBE"}} {
		t.Run(tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, false)
			review := requirementReviewFixture(tc.reviewer, tc.digest, tc.obligation)
			review.Outcomes[0].Outcome = tc.outcome
			appendEvents(t, s, reviewEvent(t, core.KindRequirementReview, review.ReviewID, review))
			if err := New().Apply(context.Background(), s); err == nil {
				t.Fatal("hostile requirement review accepted")
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*connectorreviews.RequirementReview)
	}{
		{"invalid schema", func(r *connectorreviews.RequirementReview) { r.Schema = "proofbound.requirement-review.v0" }},
		{"invalid review id", func(r *connectorreviews.RequirementReview) { r.ReviewID = "other" }},
		{"invalid reviewer", func(r *connectorreviews.RequirementReview) { r.DeclaredReviewer = "" }},
		{"invalid requirement artifact digest", func(r *connectorreviews.RequirementReview) { r.Requirement.ArtifactSHA256 = "not-a-digest" }},
		{"invalid review artifact digest", func(r *connectorreviews.RequirementReview) { r.ArtifactSHA256 = "not-a-digest" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := testStore(t)
			defer s.Close()
			intentFixtureEvents(t, s, false)
			review := requirementReviewFixture("independent", strings.Repeat("b", 64), "O-1")
			nativeID := review.ReviewID
			tc.mutate(&review)
			appendEvents(t, s, reviewEvent(t, core.KindRequirementReview, nativeID, review))
			if err := New().Apply(context.Background(), s); err == nil {
				t.Fatal("invalid requirement review digest accepted")
			}
		})
	}
}
func TestUnverifiedIsDerived(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM obligation_verdicts_view WHERE outcome='UNVERIFIED'`).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("UNVERIFIED was stored")
	}
	var out bytes.Buffer
	if err := New().ReportIntent(context.Background(), s, testCI, time.Now(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "outcome=UNVERIFIED") {
		t.Fatalf("report=%s", out.String())
	}
}

func appendDeployment(t *testing.T, s *store.Store, id int, environment, sha string) store.Record {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"repository":"proofbound/proofbound","deployment_id":%d,"environment":"%s","sha":"%s","url":"https://example.invalid/deploy/%d","created_at":"2026-09-10T00:00:00Z","updated_at":"2026-09-10T00:01:00Z"}`, id, environment, sha, id))
	return appendRaw(t, s, core.SourceGitHub, core.KindGitHubDeployment, fmt.Sprintf("proofbound/proofbound/deployment/%d", id), payload, int64(id))
}

func TestIntentDeploymentJoinsExactCommit(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	sha := shaFor("intent-commit")
	a := appendDeployment(t, s, 1, "staging", sha)
	b := appendDeployment(t, s, 2, "production", sha)
	appendDeployment(t, s, 3, "wrong", shaFor("other"))
	var out bytes.Buffer
	if err := New().ReportIntent(context.Background(), s, testCI, time.Now().Add(time.Hour), &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"staging:observed:fresh", "production:observed:fresh", a.Event.ID.String(), b.Event.ID.String()} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "wrong:observed") {
		t.Fatalf("wrong commit joined: %s", text)
	}
}
func TestIntentDeploymentStates(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	appendDeployment(t, s, 1, "production", shaFor("intent-commit"))
	var out bytes.Buffer
	if err := New().ReportIntent(context.Background(), s, testCI, time.Now().Add(48*time.Hour), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "state=DEPLOYED_UNVERIFIED") || !strings.Contains(out.String(), "production:observed:stale") {
		t.Fatalf("report=%s", out.String())
	}
}
func TestIntentDeploymentProofFailsClosed(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	intentFixtureEvents(t, s, true)
	appendDeployment(t, s, 1, "production", shaFor("intent-commit"))
	p := New()
	if err := p.Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE github_delivery_view SET event_id='01ARZ3NDEKTSV4RRFFQ69G5FAV'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.ReportIntent(context.Background(), s, testCI, time.Now(), &out); err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("out=%s err=%v", out.String(), err)
	}
}
