//go:build integration

package projections

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
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
