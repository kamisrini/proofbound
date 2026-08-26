package projections

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderWeekReport_ProofAndSupersededFixture(t *testing.T) {
	var output bytes.Buffer
	report := weekReport{
		commits:  []commitReportRow{{sha: "deadbeef", subject: "old", files: 2, decisions: []string{"VD-example-abc123"}, eventID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", superseded: true}},
		checks:   []checkReportRow{{runID: "check-1", exitCode: 0, durationMS: 42, eventID: "01ARZ3NDEKTSV4RRFFQ69G5FAW"}},
		sessions: []sessionReportRow{{sessionID: "session-1", messages: 3, tools: 4, files: 1, coverage: 0.75, eventID: "01ARZ3NDEKTSV4RRFFQ69G5FAX"}},
	}
	start := time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)
	if err := renderWeekReport(&output, start, start.Add(7*24*time.Hour), report); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{
		"commits count=1",
		"checks count=1",
		"sessions count=1",
		"deadbeef old [superseded] files=2 decisions=VD-example-abc123 proof=01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"check-1 pass duration_ms=42 proof=01ARZ3NDEKTSV4RRFFQ69G5FAW",
		"session-1 messages=3 tool_calls=4 files_written=1 coverage=0.750 proof=01ARZ3NDEKTSV4RRFFQ69G5FAX",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report missing %q:\n%s", want, text)
		}
	}
}

func TestRenderGitHubReport_StatesMissingFailedAndStaleExplicitly(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	rows := []githubReportRow{
		{repository: "github/docs", commitSHA: "a", freshness: now.Add(-48 * time.Hour), deployments: []string{"production:observed"}, deployed: true, proofs: []string{"deploy-proof/2"}},
		{repository: "github/docs", commitSHA: "b", freshness: now.Add(-time.Hour), workflows: []string{"CI:failed"}, tested: true, testFailed: true, proofs: []string{"workflow-proof/1"}},
		{repository: "github/docs", commitSHA: "c", freshness: now.Add(-time.Hour), proofs: []string{"commit-proof/3"}},
	}
	var output bytes.Buffer
	if err := renderGitHubReport(&output, now, rows); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{
		"github_deliveries count=3",
		"commit=a tested=missing deployed=production:observed freshness=stale proof=deploy-proof/2",
		"commit=b tested=CI:failed deployed=missing freshness=fresh proof=workflow-proof/1",
		"commit=c tested=missing deployed=missing freshness=fresh proof=commit-proof/3",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report missing %q:\n%s", want, text)
		}
	}
}

func TestWorkflowState(t *testing.T) {
	for _, tc := range []struct {
		status, conclusion, want string
	}{
		{"completed", "success", "passed"},
		{"completed", "failure", "failed"},
		{"in_progress", "", "running"},
	} {
		if got := workflowState(tc.status, tc.conclusion); got != tc.want {
			t.Errorf("workflowState(%q,%q)=%q want %q", tc.status, tc.conclusion, got, tc.want)
		}
	}
}
