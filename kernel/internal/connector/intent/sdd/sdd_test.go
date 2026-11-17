package sdd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
)

// replayTree commits a single-tree fixture (testdata/<name>, no steps) and returns root and SHA.
func replayTree(t *testing.T, name string) (string, string) {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
			"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
			"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")
	copyTree(t, filepath.Join("testdata", name), root)
	run("add", "-A")
	run("commit", "-q", "-m", name)
	return root, run("rev-parse", "HEAD")
}

func newProvider(t *testing.T, root string) *Provider {
	t.Helper()
	repo, err := gitcmd.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewGitReader(repo)
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(r, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type pinnedSkip struct{ Step, Unit, Cause string }

func loadSkipPins(t *testing.T) map[string][]pinnedSkip {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-skips.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	out := map[string][]pinnedSkip{}
	for k, v := range doc {
		if k == "schema" || k == "note" {
			continue
		}
		var list []pinnedSkip
		if err := json.Unmarshal(v, &list); err != nil {
			t.Fatal(err)
		}
		out[k] = list
	}
	return out
}

func isHistory(name string) bool {
	_, err := os.Stat(filepath.Join("testdata", name, "01-"+firstStepSuffix(name)))
	return err == nil
}

func firstStepSuffix(name string) string {
	entries, _ := os.ReadDir(filepath.Join("testdata", name))
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "01-") {
			return strings.TrimPrefix(e.Name(), "01-")
		}
	}
	return "\x00"
}

func stepLabels(t *testing.T, name string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var steps []string
	for _, e := range entries {
		if e.IsDir() {
			steps = append(steps, e.Name())
		}
	}
	sort.Strings(steps)
	return steps
}

// TestSkipSetsArePinned proves SDD-INV-46: every fixture's Skip set equals the pin exactly, and
// every revision the provider returns passes intent.ValidateRevision.
func TestSkipSetsArePinned(t *testing.T) {
	for fixture, want := range loadSkipPins(t) {
		t.Run(fixture, func(t *testing.T) {
			var p *Provider
			stepOf := map[string]string{}
			if isHistory(fixture) {
				root, shas := replayHistory(t, fixture)
				for i, label := range stepLabels(t, fixture) {
					stepOf[shas[i]] = label
				}
				p = newProvider(t, root)
			} else {
				root, sha := replayTree(t, fixture)
				stepOf[sha] = ""
				p = newProvider(t, root)
			}
			revs, err := p.Revisions(context.Background())
			if err != nil {
				t.Fatalf("Revisions: %v", err)
			}
			for _, r := range revs {
				if err := intent.ValidateRevision(r); err != nil {
					t.Fatalf("%s: %v", r.ArtifactPath, err)
				}
			}
			var got []string
			for _, s := range p.Skipped() {
				if !errors.Is(s.Err, intent.ErrUnmappable) {
					t.Fatalf("skip %+v does not wrap ErrUnmappable", s)
				}
				got = append(got, stepOf[s.Commit]+"|"+s.Unit+"|"+s.Cause)
			}
			var exp []string
			for _, s := range want {
				exp = append(exp, s.Step+"|"+s.Unit+"|"+s.Cause)
			}
			sort.Strings(got)
			sort.Strings(exp)
			if strings.Join(got, "\n") != strings.Join(exp, "\n") {
				t.Fatalf("skips:\n got  %q\n want %q", got, exp)
			}
		})
	}
}

// TestResolveIsTotal proves SDD-INV-59 over the pinned rows: each (step, name) yields its pinned
// outcome, and a skipped change's Resolve cause equals its Skip cause.
func TestResolveIsTotal(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-resolve.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for history, body := range doc {
		if history == "schema" || history == "note" {
			continue
		}
		var rows []struct{ Step, Name, Outcome string }
		if err := json.Unmarshal(body, &rows); err != nil {
			t.Fatal(err)
		}
		root, shas := replayHistory(t, history)
		shaOf := map[string]string{}
		for i, label := range stepLabels(t, history) {
			shaOf[label] = shas[i]
		}
		p := newProvider(t, root)
		for _, row := range rows {
			_, err := p.Resolve(context.Background(), shaOf[row.Step], row.Name)
			if got := outcomeOf(err); got != row.Outcome {
				t.Errorf("%s %s %q: got %s want %s (err=%v)", history, row.Step, row.Name, got, row.Outcome, err)
			}
		}
	}
}

func outcomeOf(err error) string {
	var u *UnmappableError
	switch {
	case err == nil:
		return "revision"
	case errors.As(err, &u):
		return "ErrUnmappable:" + u.Cause
	case errors.Is(err, ErrNoChangeIntent):
		return "ErrNoChangeIntent"
	case errors.Is(err, ErrArchivedEarlier):
		return "ErrArchivedEarlier"
	case errors.Is(err, ErrUnknownChange):
		return "ErrUnknownChange"
	}
	return "reader-error: " + err.Error()
}
