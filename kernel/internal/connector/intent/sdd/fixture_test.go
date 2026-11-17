package sdd

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// replayHistory materializes a commit-sequence fixture (testdata/<name>/NN-label/{tree,MESSAGE})
// into a fresh git repository, one commit per step in lexical step order, and returns the
// repository root and the commit SHAs in order. Each step's tree replaces the whole working tree,
// so deletions and moves between steps are real git changes. Lifecycle tests run over the whole
// sequence, never a single snapshot (draft 3 §12).
func replayHistory(t *testing.T, name string) (string, []string) {
	t.Helper()
	src := filepath.Join("testdata", name)
	entries, err := os.ReadDir(src)
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
	root := t.TempDir()
	git := func(args ...string) string {
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
	git("init", "-q")
	var shas []string
	for _, step := range steps {
		clearTree(t, root)
		copyTree(t, filepath.Join(src, step, "tree"), root)
		git("add", "-A")
		git("commit", "-q", "--allow-empty", "-F", filepath.Join(mustAbs(t, src), step, "MESSAGE"))
		shas = append(shas, git("rev-parse", "HEAD"))
	}
	return root, shas
}

func clearTree(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			t.Fatal(err)
		}
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// TestFixturesReplayAsDocumented proves the fixtures hold what testdata/README.md says they hold,
// before any provider code exists: a fixture that silently lost its defining property would make
// every later lifecycle test vacuous (LESSON class fixture-not-what-it-claims).
func TestFixturesReplayAsDocumented(t *testing.T) {
	root, shas := replayHistory(t, "openspec-history")
	if len(shas) != 10 {
		t.Fatalf("openspec-history replayed %d commits, want 10", len(shas))
	}
	has := func(commit, path string) bool {
		return exec.Command("git", "-C", root, "cat-file", "-e", commit+":"+path).Run() == nil
	}
	// 02: proposal only — the change exists with zero requirement entries.
	if !has(shas[1], "openspec/changes/add-login/proposal.md") || has(shas[1], "openspec/changes/add-login/specs") {
		t.Fatal("step 02 is not a proposal-only change")
	}
	// 05: the Logout requirement is gone and the Login statement changed, within the active change.
	at := func(commit, path string) string {
		out, err := exec.Command("git", "-C", root, "show", commit+":"+path).Output()
		if err != nil {
			t.Fatalf("show %s:%s: %v", commit, path, err)
		}
		return string(out)
	}
	before, after := at(shas[3], "openspec/changes/add-login/specs/auth/spec.md"), at(shas[4], "openspec/changes/add-login/specs/auth/spec.md")
	if !strings.Contains(before, "Requirement: Logout") || strings.Contains(after, "Requirement: Logout") || before == after {
		t.Fatal("step 05 does not remove Logout and change Login within the change")
	}
	// 06: archived into a dated directory; the active directory is gone.
	if has(shas[5], "openspec/changes/add-login") || !has(shas[5], "openspec/changes/archive/2026-01-02-add-login/specs/auth/spec.md") {
		t.Fatal("step 06 is not an archive move")
	}
	// 07: squash merge — the archive directory is absent from the first parent and arrives with code.
	if has(shas[5], "openspec/changes/archive/2026-01-03-add-signup") || !has(shas[6], "openspec/changes/archive/2026-01-03-add-signup/specs/auth/spec.md") || !has(shas[6], "src/signup.ts") {
		t.Fatal("step 07 is not a squash-merged, already-archived change with code")
	}
	// 09: --skip-specs archive leaves the main spec untouched; 10: the main spec is deleted.
	if at(shas[7], "openspec/specs/auth/spec.md") != at(shas[8], "openspec/specs/auth/spec.md") || has(shas[9], "openspec/specs/auth/spec.md") {
		t.Fatal("steps 09/10 do not model --skip-specs and main-spec deletion")
	}

	sroot, sshas := replayHistory(t, "speckit-history")
	if len(sshas) != 3 {
		t.Fatalf("speckit-history replayed %d commits, want 3", len(sshas))
	}
	if exec.Command("git", "-C", sroot, "cat-file", "-e", sshas[0]+":specs/001-export/tasks.md").Run() == nil {
		t.Fatal("speckit step 01 must have no tasks.md")
	}
	s2, _ := exec.Command("git", "-C", sroot, "show", sshas[1]+":specs/001-export/spec.md").Output()
	s3, _ := exec.Command("git", "-C", sroot, "show", sshas[2]+":specs/001-export/spec.md").Output()
	if !strings.Contains(string(s3), "FR-001") || string(s2) == string(s3) {
		t.Fatal("speckit step 03 is not an in-place FR-001 edit")
	}
}
