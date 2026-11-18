package sdd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
)

// testProfile is the canonical M1 profile used by synthesized histories.
const testProfile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": "team-fixture" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "team-fixture" } ]
}
`

func addedDelta(name, statement string) string {
	return "## ADDED Requirements\n\n### Requirement: " + name + "\n" + statement + "\n"
}

// synth builds byte-level histories that committed fixtures cannot hold (symlinks, gitlinks,
// NUL bytes, invalid UTF-8), one commit at a time.
type synth struct {
	t    *testing.T
	root string
}

func newSynth(t *testing.T) *synth {
	t.Helper()
	s := &synth{t: t, root: t.TempDir()}
	s.git("init", "-q", "-b", "main")
	s.write(ProfilePath, testProfile)
	return s
}

func (s *synth) git(args ...string) string {
	s.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = s.root
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
		"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (s *synth) write(p, content string) {
	s.t.Helper()
	full := filepath.Join(s.root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		s.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		s.t.Fatal(err)
	}
}

func (s *synth) symlink(p, target string) {
	s.t.Helper()
	full := filepath.Join(s.root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		s.t.Fatal(err)
	}
	_ = os.RemoveAll(full)
	if err := os.Symlink(target, full); err != nil {
		s.t.Fatal(err)
	}
}

func (s *synth) remove(p string) {
	s.t.Helper()
	if err := os.RemoveAll(filepath.Join(s.root, filepath.FromSlash(p))); err != nil {
		s.t.Fatal(err)
	}
}

// commit stages everything and records extra gitlinks, then commits.
func (s *synth) commit(msg string, gitlinks ...string) string {
	s.t.Helper()
	s.git("add", "-A")
	for _, g := range gitlinks {
		s.git("update-index", "--add", "--cacheinfo", "160000,"+strings.Repeat("a", 40)+","+g)
	}
	s.git("commit", "-q", "--allow-empty", "-m", msg)
	return s.git("rev-parse", "HEAD")
}

func (s *synth) provider() *Provider { return newProvider(s.t, s.root) }

// sync runs Revisions and returns revisions by artifact path and skips as "unit|cause".
func syncAll(t *testing.T, p *Provider) (map[string]intent.Revision, map[string]bool) {
	t.Helper()
	revs, err := p.Revisions(context.Background())
	if err != nil {
		t.Fatalf("Revisions: %v", err)
	}
	byPath := map[string]intent.Revision{}
	for _, r := range revs {
		byPath[string(r.Kind)+" "+r.ArtifactPath] = r
	}
	skips := map[string]bool{}
	for _, s := range p.Skipped() {
		skips[s.Unit+"|"+s.Cause] = true
	}
	return byPath, skips
}
