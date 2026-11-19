package sdd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SDD-INV-68: StateRevision agrees with Resolve on every committed fixture row, except that a change
// Resolve refuses as archived before the commit still yields its revision in that commit's state
// (an archived path), because drift asks what the change promises there, not whether it may be
// claimed.
func TestStateRevisionAgreesWithResolveExceptArchivedEarlier(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-resolve.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rows, archivedEarlier := 0, 0
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
		for _, row := range expected {
			rows++
			resolved, rerr := p.Resolve(ctx, shaOf[row.Step], row.Name)
			state, serr := p.StateRevision(ctx, shaOf[row.Step], row.Name)
			if outcomeOf(rerr) == "ErrArchivedEarlier" {
				archivedEarlier++
				if serr != nil || !strings.Contains(state.ArtifactPath, "/archive/") {
					t.Errorf("%s %s %q: archived-earlier change has no state revision: %+v %v", history, row.Step, row.Name, state, serr)
				}
				continue
			}
			if outcomeOf(serr) != outcomeOf(rerr) || state.ArtifactSHA256 != resolved.ArtifactSHA256 || state.NativeID != resolved.NativeID {
				t.Errorf("%s %s %q: StateRevision %s/%s, Resolve %s/%s", history, row.Step, row.Name, outcomeOf(serr), state.ArtifactSHA256, outcomeOf(rerr), resolved.ArtifactSHA256)
			}
		}
	}
	if rows == 0 || archivedEarlier == 0 {
		t.Fatalf("calibration: %d rows, %d archived-earlier rows; the fixture must exercise both", rows, archivedEarlier)
	}
}
