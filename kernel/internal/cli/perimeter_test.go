package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func perimeterRepo(t *testing.T) string {
	t.Helper()
	root := taskRepo(t)
	uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
	uiWrite(t, root, "specs/001-a/scope.paths", "src/a/\ntests/\n")
	uiWrite(t, root, "specs/002-b/spec.md", "# b\n")
	uiWrite(t, root, "src/b.go", "b\n")
	uiCommit(t, root, "owner", "specs")
	uiWrite(t, root, taskFile, "sdd:001-a\n")
	return root
}

func call(cwd, tool, path string) []byte {
	in := map[string]any{"cwd": cwd, "tool_name": tool, "tool_input": map[string]string{"file_path": path}}
	if tool == "NotebookEdit" {
		in["tool_input"] = map[string]string{"notebook_path": path}
	}
	b, _ := json.Marshal(in)
	return b
}

// CLI-INV-69: the PreToolUse hook refuses an edit outside the active task's declared area, read from
// HEAD, and allows everything it cannot judge.
func TestPerimeter(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	ctx := context.Background()
	root := perimeterRepo(t)
	allow := map[string][]byte{
		"inside the area":            call(root, "Write", filepath.Join(root, "src/a/x.go")),
		"a relative path inside":     call(filepath.Join(root, "src"), "Edit", "a/x.go"),
		"a test path in the area":    call(root, "Edit", filepath.Join(root, "tests/t.sh")),
		"housekeeping":               call(root, "Write", filepath.Join(root, "docs/n.md")),
		"a tool that is not an edit": call(root, "Bash", filepath.Join(root, "src/b.go")),
		"outside the repository":     call(root, "Write", filepath.Join(t.TempDir(), "x.go")),
		"malformed input":            []byte("{"),
	}
	for name, in := range allow {
		if ok, why := perimeterDecision(ctx, in); !ok {
			t.Errorf("%s was refused: %s", name, why)
		}
	}
	deny := map[string][]byte{
		"Write outside":        call(root, "Write", filepath.Join(root, "src/b.go")),
		"MultiEdit outside":    call(root, "MultiEdit", filepath.Join(root, "src/b.go")),
		"NotebookEdit outside": call(root, "NotebookEdit", filepath.Join(root, "nb/x.ipynb")),
		"dot-dot back out":     call(root, "Edit", filepath.Join(root, "src/a/../b.go")),
	}
	for name, in := range deny {
		ok, why := perimeterDecision(ctx, in)
		if ok || !strings.Contains(why, "is outside the area the spec for sdd:001-a allows (specs/001-a/scope.paths: src/a/ tests/)") {
			t.Errorf("%s: ok=%v %q", name, ok, why)
		}
	}
	t.Run("a symlink inside the area to a file outside it", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(root, "src", "a"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, "src"), filepath.Join(root, "src", "a", "up")); err != nil {
			t.Fatal(err)
		}
		if ok, _ := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/a/up/b.go"))); ok {
			t.Fatal("a symlink let an edit outside the area through")
		}
	})
	t.Run("an uncommitted wider area does not count", func(t *testing.T) {
		uiWrite(t, root, "specs/001-a/scope.paths", "**\n")
		defer uiWrite(t, root, "specs/001-a/scope.paths", "src/a/\ntests/\n")
		if ok, _ := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/b.go"))); ok {
			t.Fatal("the area was read from the working tree")
		}
	})
	t.Run("no task, or a spec without an area, allows", func(t *testing.T) {
		uiWrite(t, root, taskFile, "sdd:002-b\n")
		if ok, _ := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/b.go"))); !ok {
			t.Error("a spec without scope.paths refused")
		}
		if err := os.Remove(filepath.Join(root, taskFile)); err != nil {
			t.Fatal(err)
		}
		if ok, _ := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/b.go"))); !ok {
			t.Error("no task refused")
		}
		uiWrite(t, root, taskFile, "sdd:001-a\n")
	})
	t.Run("an invalid area refuses every code edit", func(t *testing.T) {
		bad := perimeterRepo(t)
		uiWrite(t, bad, "specs/001-a/scope.paths", "../x\n")
		uiCommit(t, bad, "owner", "break the area")
		uiWrite(t, bad, taskFile, "sdd:001-a\n")
		if ok, why := perimeterDecision(ctx, call(bad, "Write", filepath.Join(bad, "src/a/x.go"))); ok || !strings.Contains(why, "is invalid") {
			t.Fatalf("%v %q", ok, why)
		}
	})
	t.Run("a session whose cwd is outside the repository is still judged (review F2)", func(t *testing.T) {
		if ok, _ := perimeterDecision(ctx, call(t.TempDir(), "Write", filepath.Join(root, "src/b.go"))); ok {
			t.Fatal("an edit was allowed because the session ran outside the repository")
		}
	})
	t.Run("a symlink into a directory that does not exist yet (review F6)", func(t *testing.T) {
		if ok, _ := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/a/up/newdir/x.go"))); ok {
			t.Fatal("a new directory behind a symlink let an edit outside the area through")
		}
	})
	t.Run("through the command: exit 0 with Claude Code's deny JSON (review F1, F4)", func(t *testing.T) {
		old := hookStdin
		t.Cleanup(func() { hookStdin = old })
		hookStdin = bytes.NewReader(call(root, "Write", filepath.Join(root, "src/b.go")))
		var out, errb bytes.Buffer
		if code := run(ctx, []string{"hook", "pre-tool-use"}, &out, &errb); code != 0 {
			t.Fatalf("exit %d", code)
		}
		var got struct {
			HookSpecificOutput struct {
				HookEventName, PermissionDecision, PermissionDecisionReason string
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.HookSpecificOutput.HookEventName != "PreToolUse" || got.HookSpecificOutput.PermissionDecision != "deny" || !strings.Contains(got.HookSpecificOutput.PermissionDecisionReason, "outside the area") {
			t.Fatalf("%v %q", err, out.String())
		}
		hookStdin = bytes.NewReader(call(root, "Write", filepath.Join(root, "src/a/x.go")))
		out.Reset()
		if code := run(ctx, []string{"hook", "pre-tool-use"}, &out, &errb); code != 0 || out.Len() != 0 {
			t.Fatalf("an allowed edit: exit %d, stdout %q", code, out.String())
		}
	})
}

// CLI-INV-69 (init): .claude/settings.json gets the hook when absent; an existing one is never edited.
func TestInitWritesPerimeterHook(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	if err := writeClaudeSettings(&out, root); err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	data, _ := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	var parsed struct {
		Hooks struct {
			PreToolUse []struct {
				Matcher string                           `json:"matcher"`
				Hooks   []struct{ Type, Command string } `json:"hooks"`
			}
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil || len(parsed.Hooks.PreToolUse) != 1 || parsed.Hooks.PreToolUse[0].Matcher != "Edit|Write|MultiEdit|NotebookEdit" || len(parsed.Hooks.PreToolUse[0].Hooks) != 1 || parsed.Hooks.PreToolUse[0].Hooks[0].Type != "command" || parsed.Hooks.PreToolUse[0].Hooks[0].Command != "proofbound hook pre-tool-use || true" {
		t.Fatalf("%s", data)
	}
	// An old or broken proofbound (usage error exit 2, a crash) never blocks an edit (review F1).
	stub := t.TempDir()
	uiWrite(t, stub, "proofbound", "#!/bin/sh\necho usage >&2\nexit 2\n")
	if err := os.Chmod(filepath.Join(stub, "proofbound"), 0o755); err != nil {
		t.Fatal(err)
	}
	sh := exec.Command("sh", "-c", parsed.Hooks.PreToolUse[0].Hooks[0].Command)
	sh.Env = append(os.Environ(), "PATH="+stub+":/usr/bin:/bin")
	if err := sh.Run(); err != nil {
		t.Fatalf("the installed command blocks when proofbound fails: %v", err)
	}
	_ = settings
	other := t.TempDir()
	uiWrite(t, other, ".claude/settings.json", "{}\n")
	out.Reset()
	if err := writeClaudeSettings(&out, other); err != nil || !strings.Contains(out.String(), "without a Proofbound hook") || !strings.Contains(out.String(), `"PreToolUse"`) || !strings.Contains(out.String(), `"PostToolUse"`) {
		t.Fatalf("%v %s", err, out.String())
	}
	if data, _ := os.ReadFile(filepath.Join(other, ".claude", "settings.json")); string(data) != "{}\n" {
		t.Fatal("an existing settings.json was edited")
	}
	// The file init writes can be committed without a claim, and C6 lists every change to it.
	if !housekeepingPath(claudeSettingsPath) {
		t.Error(".claude/settings.json needs a claim to commit")
	}
	repo := taskRepo(t)
	uiCommit(t, repo, "owner", "start")
	uiWrite(t, repo, claudeSettingsPath, claudeSettings)
	sha := uiCommit(t, repo, "owner", "hook")
	if paths, err := trustPaths(context.Background(), repo, sha); err != nil || len(paths) != 1 || paths[0] != claudeSettingsPath {
		t.Errorf("C6 does not list the hook settings: %v %v", paths, err)
	}
}

// CLI-INV-69 (hook-guard spec): while a task file exists, an agent's file tool may not change the
// files that hold the guard on its own work, whatever the area says, in any case, through any link.
func TestPerimeterGuardsItsOwnFiles(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	ctx := context.Background()
	guarded := func(t *testing.T, in []byte, file, claim string) {
		t.Helper()
		ok, why := perimeterDecision(ctx, in)
		if ok || !strings.Contains(why, "proofbound: "+file+" holds the guard on this agent's work") || !strings.Contains(why, "the active task "+claim+")") {
			t.Errorf("not guarded: ok=%v %q", ok, why)
		}
	}
	// Written out, not guardFiles, so a file dropped from the code's list fails here.
	files := []string{".claude/settings.json", ".claude/settings.local.json", ".proofbound/task"}
	root := perimeterRepo(t)
	for _, tool := range []string{"Edit", "Write", "MultiEdit", "NotebookEdit"} {
		for _, f := range files {
			guarded(t, call(root, tool, filepath.Join(root, f)), f, "sdd:001-a")
		}
	}
	guarded(t, call(root, "Write", ".claude/settings.local.json"), ".claude/settings.local.json", "sdd:001-a")
	guarded(t, call(root, "Write", filepath.Join(root, ".CLAUDE/Settings.Local.json")), ".claude/settings.local.json", "sdd:001-a")
	t.Run("whatever the area says", func(t *testing.T) {
		uiWrite(t, root, taskFile, "sdd:002-b\n")
		defer uiWrite(t, root, taskFile, "sdd:001-a\n")
		for _, f := range files {
			guarded(t, call(root, "Write", filepath.Join(root, f)), f, "sdd:002-b")
		}
		wide := perimeterRepo(t)
		uiWrite(t, wide, "specs/001-a/scope.paths", "**\n")
		uiCommit(t, wide, "owner", "widen")
		uiWrite(t, wide, taskFile, "sdd:001-a\n")
		for _, f := range files {
			guarded(t, call(wide, "Edit", filepath.Join(wide, f)), f, "sdd:001-a")
		}
	})
	t.Run("a corrupt task file keeps the guard on", func(t *testing.T) {
		uiWrite(t, root, taskFile, "x\n")
		defer uiWrite(t, root, taskFile, "sdd:001-a\n")
		guarded(t, call(root, "Write", filepath.Join(root, ".claude/settings.json")), ".claude/settings.json", "x")
		uiWrite(t, root, taskFile, "  \n")
		guarded(t, call(root, "Write", filepath.Join(root, ".claude/settings.json")), ".claude/settings.json", "(unreadable)")
		uiWrite(t, root, taskFile, "  sdd:001-a  \nmore\n")
		guarded(t, call(root, "Write", filepath.Join(root, ".claude/settings.json")), ".claude/settings.json", "sdd:001-a")
	})
	t.Run("a linked .claude, either way round", func(t *testing.T) {
		out := t.TempDir()
		linked := perimeterRepo(t)
		if err := os.Symlink(out, filepath.Join(linked, ".claude")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(linked, "Write", filepath.Join(linked, ".claude/settings.json")), ".claude/settings.json", "sdd:001-a")
		if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, ".claude/settings.json", "{}\n")
		if err := os.MkdirAll(filepath.Join(root, "src", "a"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, ".claude/settings.json"), filepath.Join(root, "src/a/cfg.json")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(root, "Edit", filepath.Join(root, "src/a/cfg.json")), ".claude/settings.json", "sdd:001-a")
		// A link to a guard file not created yet: a Write through it would create the file.
		if err := os.Symlink(filepath.Join(root, ".claude/settings.local.json"), filepath.Join(root, "src/a/new.json")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(root, "Write", filepath.Join(root, "src/a/new.json")), ".claude/settings.local.json", "sdd:001-a")
	})
	t.Run("the session's own folder", func(t *testing.T) {
		sub := filepath.Join(root, "sub")
		guarded(t, call(sub, "Write", filepath.Join(sub, ".claude/settings.json")), ".claude/settings.json", "sdd:001-a")
		// Started at the root, sub/.claude is not loaded: judged by the area, as before.
		if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(sub, ".claude/settings.json"))); ok || strings.Contains(why, "holds the guard") {
			t.Errorf("ok=%v %q", ok, why)
		}
	})
	t.Run("the project directory after a cd (code review F1)", func(t *testing.T) {
		proj := filepath.Join(root, "src", "a")
		t.Setenv("CLAUDE_PROJECT_DIR", proj)
		guarded(t, call(filepath.Join(proj, "deep"), "Write", filepath.Join(proj, ".claude/settings.json")), ".claude/settings.json", "sdd:001-a")
	})
	t.Run("another name for the same file (code review F2)", func(t *testing.T) {
		if err := os.Link(filepath.Join(root, ".claude/settings.json"), filepath.Join(root, "src/a/hard.json")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(root, "Edit", filepath.Join(root, "src/a/hard.json")), ".claude/settings.json", "sdd:001-a")
		for in, want := range map[string]string{"C:/r/.claude/settings.json.": "C:/r/.claude/settings.json", "C:/r/.claude/settings.json::$DATA": "C:/r/.claude/settings.json", "C:/r/.claude/settings.json . ": "C:/r/.claude/settings.json", "C:/r/x.go": "C:/r/x.go"} {
			if got := windowsName(in); got != want {
				t.Errorf("windowsName(%q) = %q, want %q", in, got, want)
			}
		}
	})
	t.Run("a linked .claude seen from outside, a linked .claude inside, an unreadable task (code review F6)", func(t *testing.T) {
		out := t.TempDir()
		linked := perimeterRepo(t)
		if err := os.Symlink(out, filepath.Join(linked, ".claude")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(t.TempDir(), "Write", filepath.Join(linked, ".claude/settings.json")), ".claude/settings.json", "sdd:001-a")
		inside := perimeterRepo(t)
		if err := os.MkdirAll(filepath.Join(inside, "src/a/cfg"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(inside, "src/a/cfg"), filepath.Join(inside, ".claude")); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(inside, "Write", filepath.Join(inside, ".claude/settings.local.json")), ".claude/settings.local.json", "sdd:001-a")
		dir := perimeterRepo(t)
		if err := os.Remove(filepath.Join(dir, taskFile)); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, taskFile), 0o755); err != nil {
			t.Fatal(err)
		}
		guarded(t, call(dir, "Write", filepath.Join(dir, ".claude/settings.json")), ".claude/settings.json", "(unreadable)")
	})
	t.Run("with no task allowed; another .claude file is judged by the area, not the guard", func(t *testing.T) {
		// Amended 2026-11-27 (CLI-INV-83): a command file is an agent instruction file, so code.
		if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, ".claude/commands/x.md"))); ok || strings.Contains(why, "holds the guard") || !strings.Contains(why, "is outside the area") {
			t.Errorf("a command file: ok=%v %q", ok, why)
		}
		if err := os.Remove(filepath.Join(root, taskFile)); err != nil {
			t.Fatal(err)
		}
		defer uiWrite(t, root, taskFile, "sdd:001-a\n")
		for _, f := range files {
			if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, f))); !ok {
				t.Errorf("%s refused with no task: %s", f, why)
			}
		}
	})
}

// CLI-INV-84: while a task is active an agent may not change Proofbound's trust configuration, whatever
// the area says; CLI-INV-83: an agent instruction file is judged against the area like code.
func TestPerimeterRefusesTrustConfig(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	ctx := context.Background()
	// Written out, not trustConfigPath, so a path dropped from the code fails here.
	files := []string{".proofbound/keys/mallory.pub", ".proofbound/waivers/w.json", ".proofbound/tests.json", ".proofbound/agent-review.json",
		".proofbound/sdd-profile.json", "gates/ready-abc.yaml", ".github/workflows/proofbound.yml", ".PROOFBOUND/Keys/m.pub"}
	refused := func(t *testing.T, root, f string) {
		t.Helper()
		for _, tool := range []string{"Edit", "Write", "MultiEdit", "NotebookEdit"} {
			ok, why := perimeterDecision(ctx, call(root, tool, filepath.Join(root, f)))
			if ok || !strings.Contains(why, "proofbound: "+f+" is Proofbound's trust configuration") {
				t.Errorf("%s %s: ok=%v %q", tool, f, ok, why)
			}
		}
	}
	root := perimeterRepo(t)
	for _, f := range files {
		refused(t, root, f)
	}
	t.Run("whatever the area or the task file says", func(t *testing.T) {
		wide := perimeterRepo(t)
		uiWrite(t, wide, "specs/001-a/scope.paths", "**\n")
		uiCommit(t, wide, "owner", "widen")
		uiWrite(t, wide, taskFile, "garbage\n")
		for _, f := range files {
			refused(t, wide, f)
		}
	})
	t.Run("no task, no refusal", func(t *testing.T) {
		if err := os.Remove(filepath.Join(root, taskFile)); err != nil {
			t.Fatal(err)
		}
		defer uiWrite(t, root, taskFile, "sdd:001-a\n")
		for _, f := range files {
			if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, f))); !ok {
				t.Errorf("%s refused with no task: %s", f, why)
			}
		}
	})
	t.Run("an agent instruction file is judged against the area", func(t *testing.T) {
		for _, f := range []string{"CLAUDE.md", "AGENTS.md", "src/CLAUDE.md", ".claude/agents/r.md"} {
			ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, f)))
			if ok || !strings.Contains(why, f+" is outside the area the spec for sdd:001-a allows") {
				t.Errorf("%s: ok=%v %q", f, ok, why)
			}
		}
		if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "src/a/AGENTS.md"))); !ok {
			t.Errorf("an AGENTS.md inside the area was refused: %s", why)
		}
		if ok, why := perimeterDecision(ctx, call(root, "Write", filepath.Join(root, "docs/n.md"))); !ok {
			t.Errorf("docs stay housekeeping: %s", why)
		}
	})
}
