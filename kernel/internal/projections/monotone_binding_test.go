package projections

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/migration"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// R2 M3 SPEC §7 row 10, M3-INV-10 (P-INV-34): bindings are monotone. A commit_intents_view row,
// once projected, survives every later event for its commit, so an obligation verdict that projected
// once keeps projecting. The ledger is this repository's own pinned historical archive (a records
// change intent, its requirement, and commit c29bb3b bound to it), plus a LATER event for the same
// commit WITHOUT the ref (what a later kernel version could append), plus this repository's own p5 v2
// verdict citing the binding. Untagged, on an ephemeral port, so it runs inside make check.
func TestBindingsAreMonotone(t *testing.T) {
	if os.Getenv("DATABASE_URL") != "" {
		t.Skip("needs a fresh embedded ledger: the archive imports only into an empty one")
	}
	ctx := context.Background()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: port, AllowHistoricalEvidenceImport: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ledger.Close() })
	repo, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	records, err := migration.Load(filepath.Join(repo, migration.ArchivePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Import(ctx, ledger, records); err != nil {
		t.Fatal(err)
	}

	// The archived binding event for c29bb3b, re-emitted with its claim removed.
	const sha = "c29bb3b78899f54588f9cbad6e6ecf50333d787c"
	var bound store.Record
	if err := ledger.ReadEvents(ctx, store.Filter{Source: core.SourceGit}, func(r store.Record) error {
		if r.Event.NativeID == sha {
			bound = r
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(bound.Event.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if refs, _ := payload["intent_refs"].([]any); len(refs) != 1 {
		t.Fatalf("calibration: the archived commit must carry exactly one claim, got %v", payload["intent_refs"])
	}
	payload["intent_refs"] = nil
	unbound, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	verdictPath := "docs/verification/verdicts/p5-obligation-verdict-round1.md"
	verdictBytes, err := os.ReadFile(filepath.Join(repo, verdictPath))
	if err != nil {
		t.Fatal(err)
	}
	verdict, err := connectorreviews.ParseObligationVerdict(verdictPath, verdictBytes)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.ReviewedCommit != sha {
		t.Fatalf("calibration: the p5 verdict reviews %s, not the archived commit", verdict.ReviewedCommit)
	}
	verdictPayload, err := json.Marshal(verdict)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	run, err := ledger.BeginSync(ctx, "m3-row-10")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []core.NewEventParams{
		{Source: core.SourceGit, NativeID: sha, Kind: core.KindCommitRecorded, OccurredAt: bound.Event.OccurredAt, Payload: unbound, ConnectorVersion: bound.Event.ConnectorVersion},
		{Source: core.SourceReviews, NativeID: verdict.VerdictID, Kind: core.KindReviewVerdict, OccurredAt: time.Unix(10, 0).UTC(), Payload: verdictPayload, ConnectorVersion: connectorreviews.Version},
	} {
		e, err := ids.NewEvent(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, inserted, err := run.Append(ctx, e); err != nil || !inserted {
			t.Fatalf("append %s: inserted=%v err=%v", p.Source, inserted, err)
		}
	}
	if err := run.Finish(ctx, json.RawMessage(`[]`), nil); err != nil {
		t.Fatal(err)
	}

	projector := New()
	if err := projector.Apply(ctx, ledger); err != nil {
		t.Fatalf("Apply after a later unbound event for a bound commit: %v", err)
	}
	if err := projector.Rebuild(ctx, ledger); err != nil {
		t.Fatalf("Rebuild after a later unbound event for a bound commit: %v", err)
	}
	var rows, verdicts int
	if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM commit_intents_view WHERE commit_sha=$1`, sha).Scan(&rows); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM obligation_verdicts_view WHERE verdict_id=$1`, verdict.VerdictID).Scan(&verdicts)
	}); err != nil {
		t.Fatal(err)
	}
	if rows != 1 || verdicts == 0 {
		t.Fatalf("binding rows=%d verdict rows=%d, want the binding kept and the verdict projected", rows, verdicts)
	}

	// P-INV-35 (docs/plans/needs-work-blocks-spec.md): every row carries its verdict's status, and a
	// view from before the column is migrated in place on the next Apply, with no new event.
	status := func(when string) {
		t.Helper()
		var wrong int
		if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM obligation_verdicts_view WHERE verdict_id=$1 AND verdict_status IS DISTINCT FROM $2`, verdict.VerdictID, verdict.Status).Scan(&wrong)
		}); err != nil {
			t.Fatal(err)
		}
		if wrong != 0 {
			t.Fatalf("%s: %d verdict rows lack status %q", when, wrong, verdict.Status)
		}
	}
	status("after the rebuild")
	if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `ALTER TABLE obligation_verdicts_view DROP COLUMN verdict_status`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Apply(ctx, ledger); err != nil {
		t.Fatal(err)
	}
	status("after the in-place migration")
}
