package sdd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
)

// These cases close the per-branch gaps the full scripts/mutants.sh sweep of this package found
// (252 candidates; operator flips that no test isolated). Each names the mutant it kills.

// build.go:116#31 — a provider with no reader errors instead of dereferencing it.
func TestRevisionsWithoutReaderErrors(t *testing.T) {
	var p Provider
	if _, err := p.Revisions(context.Background()); err == nil {
		t.Fatal("a provider without a reader listed revisions")
	}
}

// provider.go:44#51 — New refuses a missing reader and a zero observation time independently.
func TestNewRefusesEachMissingDependency(t *testing.T) {
	s := newSynth(t)
	s.commit("empty")
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewGitReader(repo)
	if _, err := New(nil, time.Now()); err == nil {
		t.Fatal("New accepted a nil reader")
	}
	if _, err := New(r, time.Time{}); err == nil {
		t.Fatal("New accepted a zero observation time")
	}
}

type countingReader struct {
	Reader
	reads int
}

func (c *countingReader) Read(ctx context.Context, commit string) ([]Artifact, error) {
	c.reads++
	return c.Reader.Read(ctx, commit)
}

// build.go:136#32 — §3: each distinct state key is mapped once, and commits without a profile are
// never read.
func TestEachDistinctStateIsReadOnce(t *testing.T) {
	s := newSynth(t)
	s.remove(ProfilePath)
	s.write("src/a.go", "package a\n")
	s.commit("before opting in")
	s.write(ProfilePath, testProfile)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.commit("state one")
	s.write("src/b.go", "package b\n")
	s.commit("same SDD state")
	s.write("openspec/changes/good/proposal.md", "# good")
	s.commit("state two")
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	gr, _ := NewGitReader(repo)
	cr := &countingReader{Reader: gr}
	p, _ := New(cr, timeOne)
	if _, err := p.Revisions(context.Background()); err != nil {
		t.Fatal(err)
	}
	if cr.reads != 2 {
		t.Fatalf("read %d states, want exactly the 2 distinct ones", cr.reads)
	}
}

// reader.go:111#35 — a commit without a profile has the empty state key.
func TestNoProfileStateKeyIsEmpty(t *testing.T) {
	s := newSynth(t)
	s.remove(ProfilePath)
	c := s.commit("no profile")
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewGitReader(repo)
	if k, err := r.StateKey(context.Background(), c); err != nil || k != "" {
		t.Fatalf("key=%q err=%v", k, err)
	}
}

// parse.go:51#19 — blank is the union of JavaScript and Go whitespace: a U+0085 line (Go-only) and
// a U+FEFF line (JavaScript-only) are both blank, so neither becomes a statement.
func TestBlankIsTheWhitespaceUnion(t *testing.T) {
	for _, line := range []string{"\u0085", "\ufeff\ufeff"} {
		found, _ := parseDelta("## ADDED Requirements\n### Requirement: A\n" + line + "\nReal.\n")
		if len(found) != 1 || found[0].Statement != "Real." {
			t.Fatalf("%q: %+v", line, found)
		}
	}
}

// parse.go:188#23 — the statement is the FIRST non-blank line; later lines never overwrite it.
func TestStatementIsTheFirstLine(t *testing.T) {
	found, _ := parseDelta("## ADDED Requirements\n### Requirement: A\nFirst.\nSecond.\n")
	if len(found) != 1 || found[0].Statement != "First." {
		t.Fatalf("%+v", found)
	}
}

// profile.go:62#10 — patterns match whole paths: an exact pattern matches only its own path.
func TestPatternMatchingIsWholePath(t *testing.T) {
	for _, c := range []struct {
		pattern, path string
		want          bool
	}{
		{"openspec/changes/x/specs/core/spec.md", "openspec/changes/x/specs/core/spec.md", true},
		{"openspec/changes/x", "openspec/changes/x/specs/core/spec.md", false},
		{"openspec/*", "openspec/changes", true},
		{"openspec/*", "openspec/changes/x", false},
		{"openspec/**", "openspec/changes/x", true},
	} {
		if got := matchPattern(c.pattern, c.path); got != c.want {
			t.Errorf("match(%q, %q)=%v", c.pattern, c.path, got)
		}
	}
}

// profile.go:32#47, :128#54, :129#58/#59, :130#62/#63, :145#72, :151#76/#77/#78 — every required
// field and every grammar clause rejects on its own, in a profile otherwise consistent.
func TestProfileClausesRejectIndependently(t *testing.T) {
	consistent := func(root string) string {
		return `{"schema": "proofbound.sdd-profile.v1", "kit": "openspec",
  "requirements": { "paths": ["` + root + `/*/specs/*/spec.md", "` + root + `/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "` + root + `", "archive": "` + root + `/archive", "status": "accepted" },
  "owners": [ { "pattern": "**", "owner": "o" } ], "sponsors": [ { "pattern": "**", "sponsor": "s" } ] }`
	}
	if _, err := parseProfile([]byte(consistent("openspec/changes"))); err != nil {
		t.Fatalf("control profile rejected: %v", err)
	}
	cases := map[string]string{
		"dot segment":          consistent("openspec/./changes"),
		"dot-dot segment":      consistent("openspec/../changes"),
		"space segment":        consistent("open spec/changes"),
		"missing schema":       profileWith(`"schema": "proofbound.sdd-profile.v1",`, ``),
		"missing requirements": profileWith(`"requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },`, ``),
		"missing changes":      profileWith(`"changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },`, ``),
		"missing mode":         profileWith(`"mode": "delta-sections", `, ``),
		"missing root":         profileWith(`"root": "openspec/changes", `, ``),
		"missing archive":      profileWith(`"archive": "openspec/changes/archive", `, ``),
		"missing status":       profileWith(`, "status": "accepted"`, ``),
		"owner no pattern":     profileWith(`"pattern": "openspec/**", "owner"`, `"owner"`),
		"sponsor no pattern":   profileWith(`"pattern": "openspec/changes/**", "sponsor"`, `"sponsor"`),
		"sponsor no value":     profileWith(`, "sponsor": "team-fixture"`, ``),
		"bad sponsor pattern":  profileWith(`"pattern": "openspec/changes/**", "sponsor"`, `"pattern": ":(glob)**", "sponsor"`),
	}
	for name, body := range cases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: parseProfile panicked: %v", name, r)
				}
			}()
			if _, err := parseProfile([]byte(body)); err == nil {
				t.Errorf("%s: accepted", name)
			}
		}()
	}
}

// provider.go:287#10, :357#38 — status follows the entry: an active change's record is proposed and
// its intent takes the profile's declared status; archived ones are active / accepted.
func TestStatusFollowsTheEntry(t *testing.T) {
	s := newSynth(t)
	s.write(ProfilePath, profileWith(`"status": "accepted"`, `"status": "proposed"`))
	s.write("openspec/changes/live/specs/core/spec.md", addedDelta("Live", "L."))
	s.write("openspec/changes/archive/2026-01-01-done/specs/core/spec.md", addedDelta("Done", "D."))
	s.commit("one active, one archived")
	revs, _ := syncAll(t, s.provider())
	status := func(key string) string {
		r, ok := revs[key]
		if !ok {
			t.Fatalf("missing %s", key)
		}
		var v struct{ Status string }
		if err := json.Unmarshal(r.Payload, &v); err != nil {
			t.Fatal(err)
		}
		return v.Status
	}
	for key, want := range map[string]string{
		reqKind + "openspec/changes/live/specs/core/spec.md":                    "proposed",
		reqKind + "openspec/changes/archive/2026-01-01-done/specs/core/spec.md": "active",
		ciKind + "openspec/changes/live":                                        "proposed",
		ciKind + "openspec/changes/archive/2026-01-01-done":                     "accepted",
	} {
		if got := status(key); got != want {
			t.Errorf("%s: status %s want %s", key, got, want)
		}
	}
}

// provider.go:136#41/#42/#43, :219#55 — a dated archive entry whose capture is out of the name
// domain (".", "..", or the archive's own name) is archive-name-invalid, never a change.
func TestOutOfDomainArchiveCaptures(t *testing.T) {
	s := newSynth(t)
	for _, n := range []string{"2026-01-01-.", "2026-01-01-..", "2026-01-01-archive"} {
		s.write("openspec/changes/archive/"+n+"/specs/core/spec.md", addedDelta("X", "X."))
	}
	s.commit("out-of-domain captures")
	revs, skips := syncAll(t, s.provider())
	for _, n := range []string{"2026-01-01-.", "2026-01-01-..", "2026-01-01-archive"} {
		if !skips["archive "+n+"|archive-name-invalid"] {
			t.Errorf("%s not archive-name-invalid: %s", n, fmtSkips(skips))
		}
	}
	for k := range revs {
		t.Errorf("out-of-domain capture produced %s", k)
	}
	var _ intent.Revision
}
