//go:build integration

package projections

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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
