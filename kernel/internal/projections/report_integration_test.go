//go:build integration

package projections

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestReportWeek_RendersLedgerRowsAndSupersededCommit(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	r := appendCommit(t, s, "historical", "historical commit", 1)
	weekEnd := time.Now().UTC().Add(time.Hour)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := New().ReportWeek(context.Background(), s, weekEnd, map[string]bool{}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "historical commit [superseded]") || !strings.Contains(output.String(), "proof="+r.Event.ID.String()) {
		t.Fatalf("report=%q", output.String())
	}
}

func TestReportWeek_FailsClosedWhenProofEventIsMissing(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	appendCommit(t, s, "missing-proof", "missing proof", 1)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE commits_view SET event_id=$1`, "01ARZ3NDEKTSV4RRFFQ69G5FAV")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := New().ReportWeek(context.Background(), s, time.Now().UTC().Add(time.Hour), map[string]bool{}, &output)
	if err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("output=%q error=%v", output.String(), err)
	}
}

func TestReportGitHub_RendersJoinStatesFreshnessAndProof(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	sha := shaFor("github-report")
	workflow := []byte(`{"repository":"github/docs","run_id":7,"workflow":"CI","head_sha":"` + sha + `","status":"completed","conclusion":"failure","url":"https://github.test/run/7","created_at":"2026-08-25T12:00:00Z","updated_at":"2026-08-25T12:01:00Z"}`)
	deployment := []byte(`{"repository":"github/docs","deployment_id":9,"environment":"production","sha":"` + sha + `","url":"https://github.test/deploy/9","created_at":"2026-08-25T12:02:00Z","updated_at":"2026-08-25T12:03:00Z"}`)
	w := appendRaw(t, s, core.SourceGitHub, core.KindGitHubWorkflow, "github/docs/workflow/7", workflow, 1)
	d := appendRaw(t, s, core.SourceGitHub, core.KindGitHubDeployment, "github/docs/deployment/9", deployment, 2)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	now := time.Now().UTC().Add(time.Hour)
	if err := New().ReportGitHub(context.Background(), s, now, &output); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{
		"github_deliveries count=1",
		"tested=CI:failed",
		"deployed=production:observed",
		"freshness=fresh",
		"proof=" + w.Event.ID.String() + "/" + fmt.Sprint(w.Seq),
		fmt.Sprintf("%s/%d", d.Event.ID, d.Seq),
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report missing %q:\n%s", want, text)
		}
	}
}

func TestReportGitHub_FailsClosedWhenProofEventIsMissing(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	sha := shaFor("github-missing-proof")
	workflow := []byte(`{"repository":"github/docs","run_id":7,"workflow":"CI","head_sha":"` + sha + `","status":"completed","conclusion":"success","url":"https://github.test/run/7","created_at":"2026-08-25T12:00:00Z","updated_at":"2026-08-25T12:01:00Z"}`)
	appendRaw(t, s, core.SourceGitHub, core.KindGitHubWorkflow, "github/docs/workflow/7", workflow, 1)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE github_delivery_view SET event_id=$1`, "01ARZ3NDEKTSV4RRFFQ69G5FAV")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := New().ReportGitHub(context.Background(), s, time.Now().UTC().Add(time.Hour), &output)
	if err == nil || !strings.Contains(err.Error(), "missing event proof") {
		t.Fatalf("output=%q error=%v", output.String(), err)
	}
}
