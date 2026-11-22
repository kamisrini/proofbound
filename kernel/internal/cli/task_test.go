package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func taskRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"config", "user.name", "a"}, {"config", "user.email", "a@example.com"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

func stage(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(rel+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "add", "--", rel).CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

// CLI-INV-34: `task start` needs an existing change folder and a well-formed name; `task` shows it;
// `task stop` clears it; a corrupt task file is an error, never ignored.
func TestTaskStartShowStop(t *testing.T) {
	root := taskRepo(t)
	var out bytes.Buffer
	if err := runTask(root, "start", "001-export", &out); err == nil || !strings.Contains(err.Error(), "write the spec first") {
		t.Fatalf("a change with no folder was accepted: %v", err)
	}
	if err := runTask(root, "start", "../x", &out); err == nil {
		t.Fatal("a path was accepted as a change name")
	}
	if err := os.MkdirAll(filepath.Join(root, "specs", "001-export"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runTask(root, "start", "001-export", &out); err != nil {
		t.Fatal(err)
	}
	if claim, err := currentTask(root); err != nil || claim != "sdd:001-export" {
		t.Fatalf("task = %q, %v", claim, err)
	}
	out.Reset()
	if err := runTask(root, "show", "", &out); err != nil || out.String() != "task sdd:001-export\n" {
		t.Fatalf("show: %q %v", out.String(), err)
	}
	if err := runTask(root, "stop", "", &out); err != nil {
		t.Fatal(err)
	}
	if _, err := currentTask(root); !errors.Is(err, errNoTask) {
		t.Fatalf("stop left a task: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".proofbound", "task"), []byte("sdd:a b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := currentTask(root); err == nil || errors.Is(err, errNoTask) {
		t.Fatalf("a corrupt task file was ignored: %v", err)
	}
}

// CLI-INV-35/36: prepare-commit-msg stamps the claim only on code commits under an active task; the
// commit-msg hook refuses unclaimed code and malformed claims, and passes housekeeping and merges.
func TestHooksStampAndRefuse(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	msgFile := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	hook := func(name, msg string) (string, error) {
		t.Helper()
		if err := os.WriteFile(msgFile, []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		var stderr bytes.Buffer
		err := runHook(ctx, root, name, msgFile, "message", &stderr)
		raw, _ := os.ReadFile(msgFile)
		if err != nil && !errors.Is(err, errHookBlock) {
			t.Fatalf("%s: %v", name, err)
		}
		return string(raw), err
	}

	// Housekeeping only, no task: nothing stamped, commit allowed.
	stage(t, root, "docs/notes.md")
	if got, err := hook("prepare-commit-msg", "notes\n"); got != "notes\n" || err != nil {
		t.Fatalf("housekeeping stamped or refused: %q %v", got, err)
	}
	if _, err := hook("commit-msg", "notes\n"); err != nil {
		t.Fatal("a housekeeping commit was refused")
	}

	// Code, no task: not stamped, refused.
	stage(t, root, "src/a.go")
	if got, _ := hook("prepare-commit-msg", "work\n"); got != "work\n" {
		t.Fatalf("stamped without a task: %q", got)
	}
	if _, err := hook("commit-msg", "work\n"); !errors.Is(err, errHookBlock) {
		t.Fatal("unclaimed code was accepted")
	}

	// Code with a task: stamped in the kernel's trailer grammar, then accepted.
	if err := os.MkdirAll(filepath.Join(root, "specs", "001-export"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runTask(root, "start", "001-export", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	got, _ := hook("prepare-commit-msg", "work\n")
	if got != "work\n\nIntent: sdd:001-export\n" {
		t.Fatalf("stamp = %q", got)
	}
	if _, err := hook("commit-msg", got); err != nil {
		t.Fatal("a stamped commit was refused")
	}
	if got, _ := hook("prepare-commit-msg", "work\n\nCo-Authored-By: x <x@y>\n"); got != "work\n\nCo-Authored-By: x <x@y>\nIntent: sdd:001-export\n" {
		t.Fatalf("stamp did not join the trailer block: %q", got)
	}
	if got, _ := hook("prepare-commit-msg", "work\n\nIntent: sdd:other\n"); got != "work\n\nIntent: sdd:other\n" {
		t.Fatalf("an explicit claim was overridden: %q", got)
	}
	if _, err := hook("commit-msg", "work\n\nintent: SDD:001-export\n"); !errors.Is(err, errHookBlock) {
		t.Fatal("a malformed claim was accepted")
	}
	// The retired opt-out no longer exempts code.
	t.Setenv("PROOFBOUND_NO_INTENT", "1")
	if _, err := hook("commit-msg", "work\n"); !errors.Is(err, errHookBlock) {
		t.Fatal("PROOFBOUND_NO_INTENT still exempts code")
	}
}

func TestWithTrailer(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"fix: a thing\n", "fix: a thing\n\nIntent: sdd:x\n"},
		{"subject\n\nbody line\n", "subject\n\nbody line\n\nIntent: sdd:x\n"},
		{"subject\n\nSigned-off-by: a <a@b>\n\n", "subject\n\nSigned-off-by: a <a@b>\nIntent: sdd:x\n"},
		{"subject\n\n# Please enter the commit message\n# Lines starting with '#' are ignored\n", "subject\n\nIntent: sdd:x\n\n# Please enter the commit message\n# Lines starting with '#' are ignored\n"},
	} {
		if got := withTrailer(c.in, "Intent: sdd:x"); got != c.want {
			t.Errorf("withTrailer(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// F1 of mvp-uncovered-and-agent-intent-review-20261122: `git commit -a` / `git commit <path>` run the
// hooks against a temporary index named by GIT_INDEX_FILE; the hook must read that one.
func TestHooksReadGitsIndexFile(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	alt := filepath.Join(t.TempDir(), "index")
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	add := exec.Command("git", "-C", root, "add", "a.go")
	add.Env = append(os.Environ(), "GIT_INDEX_FILE="+alt)
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	msg := filepath.Join(t.TempDir(), "msg")
	if err := os.WriteFile(msg, []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runHook(ctx, root, "commit-msg", msg, "", &bytes.Buffer{}); err != nil {
		t.Fatalf("the real index is empty, so this should pass: %v", err)
	}
	t.Setenv("GIT_INDEX_FILE", alt)
	if err := runHook(ctx, root, "commit-msg", msg, "", &bytes.Buffer{}); !errors.Is(err, errHookBlock) {
		t.Fatalf("code staged in GIT_INDEX_FILE was not seen: %v", err)
	}
}

// F2: a claim the kernel would not bind (not in the final trailer block) is refused.
func TestCommitMsgRefusesAClaimTheKernelSkips(t *testing.T) {
	root := taskRepo(t)
	msg := filepath.Join(t.TempDir(), "msg")
	try := func(text string) error {
		if err := os.WriteFile(msg, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return runHook(context.Background(), root, "commit-msg", msg, "", &bytes.Buffer{})
	}
	if !errors.Is(try("work\nIntent: sdd:x\n"), errHookBlock) {
		t.Fatal("a claim with no blank line before it was accepted")
	}
	if !errors.Is(try("work\n\nIntent: sdd:x\n\nmore body\n"), errHookBlock) {
		t.Fatal("a claim in the body was accepted")
	}
	if err := try("work\n\nIntent: sdd:x\n# Please enter the commit message\n"); err != nil {
		t.Fatalf("a trailer followed by git's comment lines was refused: %v", err)
	}
	if err := try("subject" + withTrailer("\n# Please enter the commit message\n", "Intent: sdd:x")); err != nil {
		t.Fatalf("a subject typed on line 1 of the stamped editor message was refused: %v", err)
	}
}
