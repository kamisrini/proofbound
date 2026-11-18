package sdd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// These cases close the test-adequacy gaps the M1 non-author reviews left open beyond the seven
// named survivors that survivors_test.go kills: re-check r2-m1-recheck-round1-20261118 F2's List
// path, and round-1 r2-m1-review-round1-20261117 F8's narrower-than-the-row tests.

func gitReaderFor(t *testing.T, s *synth) *GitReader {
	t.Helper()
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewGitReader(repo)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

type failingList struct{ Reader }

func (failingList) List(context.Context, string, string) ([]Entry, error) {
	return nil, errors.New("listing on fire")
}

// Re-check F2, the List twin of V-parents-error-as-unknown (the Parents half is
// TestParentsFailureIsNotASentinel): SDD-INV-59's "any other error comes only from the reader"
// holds when listing a parent's archive fails.
func TestResolveParentListFailureIsAReaderError(t *testing.T) {
	ctx := context.Background()
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.commit("base")
	s.write("openspec/changes/archive/2026-01-01-done/specs/core/spec.md", addedDelta("Done", "D."))
	c := s.commit("archive done")
	gr := gitReaderFor(t, s)
	control, _ := New(gr, timeOne)
	if _, err := control.Resolve(ctx, c, "done"); err != nil {
		t.Fatalf("control: a singly-archived change must resolve: %v", err)
	}
	for name, r := range map[string]Reader{"List": failingList{gr}} {
		p, _ := New(r, timeOne)
		_, err := p.Resolve(ctx, c, "done")
		if got := outcomeOf(err); !strings.HasPrefix(got, "reader-error") {
			t.Errorf("a failing %s surfaced as %s", name, got)
		}
	}
}

// Round-1 F8, SDD-INV-61: a symlinked root, a blob archive, and a gitlink root are three
// state-level Skips (counted as Skips, not as a unit|cause set), and a later clean state maps.
func TestStateLevelSkipsAreCounted(t *testing.T) {
	s := newSynth(t)
	s.write("real/x/specs/core/spec.md", addedDelta("X", "X."))
	s.symlink("openspec/changes", "../real")
	s.commit("symlinked root")
	s.remove("openspec/changes")
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.write("openspec/changes/archive", "a blob where the archive belongs")
	s.commit("blob archive")
	s.remove("openspec/changes")
	s.commit("gitlink root", "openspec/changes")
	s.git("rm", "-q", "--cached", "openspec/changes")
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.commit("clean")
	p := s.provider()
	revs, err := p.Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, sk := range p.Skipped() {
		if sk.Unit != "profile" || sk.Cause != "non-regular-file" {
			t.Errorf("unexpected skip %s|%s", sk.Unit, sk.Cause)
		}
		n++
	}
	if n != 3 {
		t.Fatalf("want 3 state-level skips, got %d", n)
	}
	for _, r := range revs {
		if r.Kind == core.KindChangeIntent && r.ArtifactPath == "openspec/changes/good" {
			return
		}
	}
	t.Fatal("the later clean state was not emitted")
}

// Round-1 F8, SDD-INV-47: a requirement payload of exactly core.MaxPayloadBytes is emitted and
// appended end to end; one byte more is payload-too-large and never reaches the appender.
func TestExactBoundPayloadSyncs(t *testing.T) {
	payloadFor := func(n int) (*synth, int) {
		s := newSynth(t)
		s.write("openspec/changes/big/specs/core/spec.md", addedDelta("Big", strings.Repeat("a", n)))
		s.commit("big")
		revs, _ := syncAll(t, s.provider())
		r, ok := revs[reqKind+"openspec/changes/big/specs/core/spec.md"]
		if !ok {
			return s, -1
		}
		return s, len(r.Payload)
	}
	_, base := payloadFor(1000)
	if base < 0 {
		t.Fatal("calibration: small payload not emitted")
	}
	exactN := 1000 + core.MaxPayloadBytes - base
	s, got := payloadFor(exactN)
	if got != core.MaxPayloadBytes {
		t.Fatalf("calibration: payload is %d bytes, want exactly %d", got, core.MaxPayloadBytes)
	}
	app := &memAppender{}
	c, err := intent.New(&intent.Deps{Providers: []intent.Provider{s.provider()}, IDs: testIDs(t), Logger: discard()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sync(context.Background(), "sdd", app); err != nil {
		t.Fatalf("exact-bound sync: %v", err)
	}
	found := false
	for _, e := range app.events {
		if len(e.Payload) == core.MaxPayloadBytes {
			found = true
		}
	}
	if !found {
		t.Fatal("the exact-bound payload was not appended")
	}
	over, _ := payloadFor(exactN + 1)
	if _, skips := syncAll(t, over.provider()); !skips["requirement openspec/changes/big/specs/core/spec.md|payload-too-large"] {
		t.Fatalf("one byte over the bound: %s", fmtSkips(skips))
	}
}
