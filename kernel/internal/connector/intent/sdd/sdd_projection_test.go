//go:build integration

package sdd

import (
	"context"
	"net"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func embeddedStore(t *testing.T) *store.Store {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()
	s, err := store.Open(context.Background(), store.Config{Root: t.TempDir(), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestSDDEventsProject proves SDD-INV-52: a clean OpenSpec-shaped history plus a synthesized one
// carrying a unit for each M1-reachable cause syncs into a real ledger, and a projection rebuild
// then succeeds (no U+0000 or oversize value reaches a TEXT column), with requirement and
// change-intent rows present for the clean units.
func TestSDDEventsProject(t *testing.T) {
	ctx := context.Background()
	ledger := embeddedStore(t)
	root, _ := replayHistory(t, "openspec-history")
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "The system SHALL be good."))
	hostile(s)
	s.commit("hostile")
	ids := testIDs(t) // one generator: its entropy and clock are fixed, so two would mint equal ids
	for i, p := range []*Provider{newProvider(t, root), s.provider()} {
		c, err := intent.New(&intent.Deps{Providers: []intent.Provider{p}, IDs: ids, Logger: discard()})
		if err != nil {
			t.Fatal(err)
		}
		run, err := ledger.BeginSync(ctx, "intent.sdd")
		if err != nil {
			t.Fatal(err)
		}
		res, syncErr := c.Sync(ctx, "sdd", run)
		if err := run.Finish(ctx, res.Cursor, syncErr); err != nil || syncErr != nil {
			t.Fatalf("history %d: sync=%v finish=%v", i, syncErr, err)
		}
		if res.Appended == 0 {
			t.Fatalf("history %d appended nothing", i)
		}
	}
	if err := projections.New().Rebuild(ctx, ledger); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	var reqs, cis int
	if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM requirements_view WHERE source = 'intent.sdd'`).Scan(&reqs); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM change_intents_view WHERE source = 'intent.sdd'`).Scan(&cis)
	}); err != nil {
		t.Fatal(err)
	}
	if reqs == 0 || cis == 0 {
		t.Fatalf("rows: requirements=%d change_intents=%d", reqs, cis)
	}
}
