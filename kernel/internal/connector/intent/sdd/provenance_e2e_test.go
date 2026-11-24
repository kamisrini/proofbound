package sdd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hookCall runs `proofbound hook post-tool-use` through cli.Run with Claude Code's JSON on stdin.
func (m *m3Tenant) hookCall(call map[string]any) {
	m.t.Helper()
	raw, err := json.Marshal(call)
	if err != nil {
		m.t.Fatal(err)
	}
	in := filepath.Join(m.t.TempDir(), "stdin.json")
	if err := os.WriteFile(in, raw, 0o644); err != nil {
		m.t.Fatal(err)
	}
	f, err := os.Open(in)
	if err != nil {
		m.t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old; _ = f.Close() }()
	m.mustRun("hook", "post-tool-use")
}

// CLI-INV-70/71/72 end to end through cli.Run: a Write and an Edit through the hook and a
// shell-written line, then draft, seal and audit; and an agent commit whose edits were never recorded
// warns that the hook was probably not running.
func TestSessionProvenanceEndToEnd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := removeHooks(m.root); err != nil {
		t.Fatal(err)
	}
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	m.commit("specify 001-export")
	file := filepath.Join(m.root, "src", "x.sh")
	m.write("src/x.sh", "echo header row\necho first row\n")
	m.hookCall(map[string]any{"session_id": "e2e-session-1", "cwd": m.root, "tool_name": "Write", "tool_input": map[string]any{"file_path": file, "content": "echo header row\necho first row\n"}, "tool_response": map[string]any{"type": "create"}})
	m.write("src/x.sh", "echo header row\necho first row\necho second row\n")
	m.hookCall(map[string]any{"session_id": "e2e-session-1", "cwd": m.root, "tool_name": "Edit", "tool_input": map[string]any{"file_path": file, "old_string": "echo first row", "new_string": "echo first row\necho second row"}})
	m.write("src/x.sh", "echo header row\necho first row\necho second row\necho from the shell\n")
	x := m.commit("implement 001-export", "Co-Authored-By: Claude <noreply@anthropic.com>", "Intent: sdd:001-export")
	m.mustRun("sync", "all")
	out := m.mustRun("review", "draft", x)
	if want := "provenance src/x.sh: 3 by Claude Code session e2e-sess, 1 with no recorded agent edit (lines 4), 0 common"; !strings.Contains(out, want) {
		t.Fatalf("review draft lacks %q:\n%s", want, out)
	}
	raw, err := os.ReadFile(filepath.Join(m.root, ".proofbound", "drafts", x, "draft.json"))
	if err != nil || !strings.Contains(string(raw), `"provenance"`) {
		t.Fatalf("draft.json has no provenance: %v %s", err, raw)
	}
	t.Setenv("PROOFBOUND_KEY", filepath.Join(t.TempDir(), "reviewer-key"))
	m.mustRun("keys", "init", "--name", "proofbound-reviewer")
	m.review(x, "SATISFIED")
	rec, err := os.ReadFile(filepath.Join(m.root, "docs", "verification", "verdicts", "PROV-"+x[:12]+".json"))
	if err != nil || !strings.Contains(string(rec), "e2e-session-1") {
		t.Fatalf("no PROV- record: %v %s", err, rec)
	}

	// An agent commit none of whose lines any recorded edit wrote: the hook was probably not running.
	m.write("src/y.sh", "echo never recorded\n")
	y := m.commit("more export", "Co-Authored-By: Claude <noreply@anthropic.com>", "Intent: sdd:001-export")
	m.mustRun("sync", "all")
	if out := m.mustRun("review", "draft", y); !strings.Contains(out, "provenance warning: the edit-recording hook was probably not running for this commit") {
		t.Fatalf("no hook warning:\n%s", out)
	}
}
