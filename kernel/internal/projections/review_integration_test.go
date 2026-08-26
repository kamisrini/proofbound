//go:build integration

package projections

import (
	"context"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestApply_ReviewFindingsRetainProofAndRevision(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	first := []byte(`{"schema":"vera.verdict.v1","verdict_id":"task9-round1","status":"NEEDS_WORK","reviewed_commit":"0123456789012345678901234567890123456789","findings":[{"finding_id":"F-1","severity":"MED","defect_commit":""}],"artifact_path":"docs/verification/verdicts/task9-round1.md","artifact_sha":"0000000000000000000000000000000000000000000000000000000000000000"}`)
	appendRaw(t, s, core.SourceReviews, core.KindReviewVerdict, "task9-round1", first, 1)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	second := []byte(`{"schema":"vera.verdict.v1","verdict_id":"task9-round2","status":"ACCEPTABLE","reviewed_commit":"0123456789012345678901234567890123456789","findings":[{"finding_id":"F-1","severity":"LOW","defect_commit":""}],"artifact_path":"docs/verification/verdicts/task9-round2.md","artifact_sha":"0000000000000000000000000000000000000000000000000000000000000000"}`)
	appendRaw(t, s, core.SourceReviews, core.KindReviewVerdict, "task9-round2", second, 2)
	if err := New().Apply(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	var status, severity, eventID string
	var seq int64
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT status,severity,event_id,seq FROM reviews_view WHERE finding_id='F-1'`).Scan(&status, &severity, &eventID, &seq)
	}); err != nil {
		t.Fatal(err)
	}
	if status != "ACCEPTABLE" || severity != "LOW" || eventID == "" || seq != 2 {
		t.Fatalf("row=(%s,%s,%s,%d)", status, severity, eventID, seq)
	}
}

func TestRedVerdictChainRequiresStrictLedgerInterval(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	sha := "0123456789012345678901234567890123456789"
	verdict := func(id, status string) []byte {
		return []byte(`{"schema":"vera.verdict.v1","verdict_id":"` + id + `","status":"` + status + `","reviewed_commit":"` + sha + `","findings":[{"finding_id":"` + id + `-finding","severity":"MED","defect_commit":""}],"artifact_path":"docs/verification/verdicts/` + id + `.md","artifact_sha":"0000000000000000000000000000000000000000000000000000000000000000"}`)
	}
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, "before", commitJSON(sha, "before"), 1)
	appendRaw(t, s, core.SourceReviews, core.KindReviewVerdict, "red", verdict("red", "NEEDS_WORK"), 2)
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, "between", commitJSON(sha, "between"), 3)
	appendRaw(t, s, core.SourceReviews, core.KindReviewVerdict, "next", verdict("next", "ACCEPTABLE"), 4)
	appendRaw(t, s, core.SourceGit, core.KindCommitRecorded, "after", commitJSON(sha, "after"), 5)
	chains, err := readRedVerdictChains(context.Background(), s, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 1 || len(chains[0].changeEventIDs) != 1 {
		t.Fatalf("chains=%+v", chains)
	}
}
