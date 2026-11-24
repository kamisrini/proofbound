package sdd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CLI-INV-66/67 end to end through cli.Run: with tests declared in the parent, `review draft`
// prints which hunks the tests did not catch and writes them into draft.json; seal with them
// rejected records NEEDS_WORK and a TESTS- record.
func TestTestsCanFailEndToEnd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := removeHooks(m.root); err != nil {
		t.Fatal(err)
	}
	m.write(".proofbound/tests.json", `{"schema":"proofbound.tests.v1","run":"sh tests/t.sh","paths":["tests/"]}`+"\n")
	gitIn(t, m.root, "add", "-f", "--", ".proofbound/tests.json")
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	m.commit("specify 001-export")
	m.write("src/x.sh", "echo a,b\n")
	m.write("tests/t.sh", "sh src/x.sh > /dev/null; true\n")
	gitIn(t, m.root, "add", "--", "tests")
	x := m.commit("implement 001-export", "Intent: sdd:001-export")
	m.mustRun("sync", "all")
	out := m.mustRun("review", "draft", x)
	for _, want := range []string{"this runs the commit's own code", "tests did NOT catch undoing src/x.sh line 1", "tests did NOT catch undoing the whole change"} {
		if !strings.Contains(out, want) {
			t.Fatalf("review draft lacks %q:\n%s", want, out)
		}
	}
	path := filepath.Join(m.root, ".proofbound", "drafts", x, "draft.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var draft map[string]any
	if err := json.Unmarshal(raw, &draft); err != nil {
		t.Fatal(err)
	}
	tests, ok := draft["tests"].(map[string]any)
	if !ok || tests["whole"] != "not caught" {
		t.Fatalf("draft.json has no tests result: %s", raw)
	}
	for _, h := range tests["hunks"].([]any) {
		h.(map[string]any)["decision"] = "reject"
	}
	edited, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROOFBOUND_KEY", filepath.Join(t.TempDir(), "reviewer-key"))
	m.mustRun("keys", "init", "--name", "proofbound-reviewer")
	m.review(x, "SATISFIED")
	ov, _ := filepath.Glob(filepath.Join(m.root, "docs", "verification", "verdicts", "OV-*.md"))
	data, _ := os.ReadFile(ov[0])
	if !strings.Contains(string(data), `"status":"NEEDS_WORK"`) {
		t.Fatalf("rejected hunks did not fail the review:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(m.root, "docs", "verification", "verdicts", "TESTS-"+x[:12]+".json")); err != nil {
		t.Fatal("no TESTS- record")
	}
}
