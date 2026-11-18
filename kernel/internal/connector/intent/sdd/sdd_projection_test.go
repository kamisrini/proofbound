//go:build integration

package sdd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
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
	// The pinned record-id collision pair collides only under testProfile's exact bytes.
	for _, n := range []string{"a.._-.--_..b", "a---__-___-_b"} {
		s.write("openspec/changes/"+n+"/specs/core/spec.md", addedDelta("Do", "The system SHALL do."))
	}
	s.commit("hostile")
	// The four causes hostile() cannot carry (M1 review F3): profile-invalid at a state of its own,
	// owner-uncovered and sponsor-uncovered under a narrower profile, and the pinned record-id
	// collision pair.
	s.write(ProfilePath, `{"schema": "proofbound.sdd-profile.v1"}`)
	s.commit("profile invalid")
	s.write(ProfilePath, profileWith(`[ { "pattern": "openspec/**", "owner": "team-fixture" } ]`,
		`[ { "pattern": "openspec/changes/good/**", "owner": "team-fixture" }, { "pattern": "openspec/changes/sponsorless/**", "owner": "team-fixture" }, { "pattern": "openspec/changes/a.._-.--_..b/**", "owner": "team-fixture" }, { "pattern": "openspec/changes/a---__-___-_b/**", "owner": "team-fixture" } ]`))
	s.write(ProfilePath, strings.Replace(readFile(t, filepath.Join(s.root, ProfilePath)), `"pattern": "openspec/changes/**", "sponsor"`, `"pattern": "openspec/changes/good/**", "sponsor"`, 1))
	s.write("openspec/changes/ownerless/specs/core/spec.md", addedDelta("Owner", "O."))
	s.write("openspec/changes/sponsorless/specs/core/spec.md", addedDelta("Sponsor", "S."))
	s.commit("narrow profile")
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
		if i == 1 {
			causes := map[string]bool{}
			for _, sk := range res.ByProvider["sdd"].Skipped {
				causes[sk.Cause] = true
			}
			for _, c := range []string{"profile-invalid", "invalid-utf8", "nul-byte", "non-regular-file", "owner-uncovered", "sponsor-uncovered", "obligation-collision", "empty-name-part", "slug-too-short", "unpaired-rename", "record-id-collision", "name-reuse", "archive-name-invalid", "payload-too-large", "cascade"} {
				if !causes[c] {
					t.Errorf("hostile history lacks a %s unit", c)
				}
			}
		}
	}
	if err := projections.New().Rebuild(ctx, ledger); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	// Every clean unit has its row, per unit (not merely a non-empty view).
	for _, want := range []struct{ view, path string }{
		{"requirements_view", "openspec/changes/good/specs/core/spec.md"},
		{"change_intents_view", "openspec/changes/good"},
		{"requirements_view", "openspec/changes/sponsorless/specs/core/spec.md"},
		{"requirements_view", "openspec/changes/archive/2026-01-02-add-login/specs/auth/spec.md"},
		{"change_intents_view", "openspec/changes/archive/2026-01-03-add-signup"},
	} {
		var n int
		if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM `+want.view+` WHERE source = 'intent.sdd' AND payload->>'artifact_path' = $1`, want.path).Scan(&n)
		}); err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			t.Errorf("%s has no row for %s", want.view, want.path)
		}
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
