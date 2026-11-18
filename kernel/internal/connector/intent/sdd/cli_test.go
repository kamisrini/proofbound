package sdd_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/cli"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// tenant commits a single-tree fixture into a fresh repository and returns its root and HEAD.
func tenant(t *testing.T, fixture string) (string, string) {
	t.Helper()
	root := t.TempDir()
	gitIn(t, root, "init", "-q", "-b", "main")
	copyDir(t, filepath.Join("testdata", fixture), root)
	gitIn(t, root, "add", "-A")
	gitIn(t, root, "commit", "-q", "-m", fixture)
	return root, gitIn(t, root, "rev-parse", "HEAD")
}

func runCLI(t *testing.T, dir string, args ...string) (int, string) {
	t.Helper()
	t.Chdir(dir)
	t.Setenv("DATABASE_URL", "")
	var out, errb bytes.Buffer
	code := cli.Run(context.Background(), "", args, &out, &errb)
	return code, out.String() + errb.String()
}

// TestSyncIntentSDDPrintsSkips proves SDD-INV-60: `sync intent sdd` prints every Skip after the
// counts line and exits 0.
func TestSyncIntentSDDPrintsSkips(t *testing.T) {
	root, head := tenant(t, "unmappable/reused-name")
	code, out := runCLI(t, root, "sync", "intent", "sdd")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	want := "skip name-reuse change add-login at " + head
	if !strings.Contains(out, "listed=") || !strings.Contains(out, want) {
		t.Fatalf("output lacks %q:\n%s", want, out)
	}
}

// TestIntentAllExcludesSDD proves SDD-INV-56: on a shallow clone with no profile, `sync intent all`
// behaves as before M1 (it succeeds: it never constructs or walks sdd), while `sync intent sdd`
// refuses the shallow history.
func TestIntentAllExcludesSDD(t *testing.T) {
	source, _ := tenant(t, "fenced")
	gitIn(t, source, "commit", "-q", "--allow-empty", "-m", "second")
	clone := filepath.Join(t.TempDir(), "clone")
	gitIn(t, filepath.Dir(clone), "clone", "-q", "--depth=1", "file://"+filepath.ToSlash(source), clone)
	if code, out := runCLI(t, clone, "sync", "intent", "all"); code != 0 {
		t.Fatalf("sync intent all on a shallow clone: exit %d: %s", code, out)
	}
	if code, out := runCLI(t, clone, "sync", "intent", "sdd"); code == 0 || !strings.Contains(out, "incomplete") {
		t.Fatalf("sync intent sdd on a shallow clone: exit %d: %s", code, out)
	}
}
