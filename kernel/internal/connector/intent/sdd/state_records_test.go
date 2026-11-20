package sdd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// SDD-INV-72: on every committed fixture row, StateRecords returns StateRevision's change intent
// (same outcome, same bytes) plus exactly the requirement revisions its targets name, byte-identical
// to the ones a full Revisions walk (what the ledger holds) emits for those (id, digest) pairs.
func TestStateRecordsMatchesTheLedgeredRevisions(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-resolve.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	checked := 0
	for history, body := range doc {
		if history == "schema" || history == "note" {
			continue
		}
		var expected []struct{ Step, Name, Outcome string }
		if err := json.Unmarshal(body, &expected); err != nil {
			t.Fatal(err)
		}
		root, shas := replayHistory(t, history)
		shaOf := map[string]string{}
		for i, label := range stepLabels(t, history) {
			shaOf[label] = shas[i]
		}
		p := newProvider(t, root)
		all, err := p.Revisions(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ledger := map[string][]byte{}
		for _, rev := range all {
			if rev.Kind == core.KindRequirement {
				ledger[rev.NativeID+"@"+rev.ArtifactSHA256] = rev.Payload
			}
		}
		for _, row := range expected {
			state, serr := p.StateRevision(ctx, shaOf[row.Step], row.Name)
			change, reqs, rerr := p.StateRecords(ctx, shaOf[row.Step], row.Name)
			if outcomeOf(serr) != outcomeOf(rerr) || !bytes.Equal(state.Payload, change.Payload) {
				t.Errorf("%s %s %q: StateRecords %s, StateRevision %s", history, row.Step, row.Name, outcomeOf(rerr), outcomeOf(serr))
				continue
			}
			if rerr != nil {
				continue
			}
			var ci intent.ChangeIntent
			if err := json.Unmarshal(change.Payload, &ci); err != nil {
				t.Fatal(err)
			}
			if len(reqs) != len(ci.Targets) {
				t.Errorf("%s %s %q: %d requirements for %d targets", history, row.Step, row.Name, len(reqs), len(ci.Targets))
			}
			for _, rev := range reqs {
				if got, ok := ledger[rev.NativeID+"@"+rev.ArtifactSHA256]; !ok || !bytes.Equal(got, rev.Payload) {
					t.Errorf("%s %s %q: requirement %s differs from the ledgered revision", history, row.Step, row.Name, rev.NativeID)
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("calibration: no fixture row yielded a change intent")
	}
}
