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
