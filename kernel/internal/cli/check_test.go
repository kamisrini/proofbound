package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// CLI-INV-26 (parsing): `check <base> [--markdown] [--require-claims]`, each flag at most once.
func TestParseCheck(t *testing.T) {
	ok := map[string][3]string{
		"check main":                          {"main", "false", "false"},
		"check origin/main --markdown":        {"origin/main", "true", "false"},
		"check b --require-claims --markdown": {"b", "true", "true"},
	}
	for in, want := range ok {
		base, md, rc, good := parseCheck(strings.Fields(in))
		if !good || base != want[0] || strings.ToLower(boolStr(md)) != want[1] || boolStr(rc) != want[2] {
			t.Fatalf("%q: base=%q md=%v rc=%v ok=%v", in, base, md, rc, good)
		}
		if parseCommand(strings.Fields(in)) != commandCheck {
			t.Fatalf("%q not routed to check", in)
		}
	}
	for _, in := range []string{"check", "check --markdown", "check b --markdown --markdown", "check b --other", "check b c"} {
		if _, _, _, good := parseCheck(strings.Fields(in)); good {
			t.Fatalf("%q accepted", in)
		}
		if parseCommand(strings.Fields(in)) != commandInvalid {
			t.Fatalf("%q not a usage error", in)
		}
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// CLI-INV-26 (rendering): a reason cannot break the Markdown table or the drift code block.
func TestCheckMarkdownEscapes(t *testing.T) {
	var b bytes.Buffer
	lines := []checkLine{{commit: strings.Repeat("a", 40), claims: "sdd:x", state: checkBlocked, reason: "a|b\nc `d`"}}
	if err := writeCheckMarkdown(&b, "Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)", lines, "x ```y```\n"); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	row := "| `aaaaaaaaaaaa` | `sdd:x` | **BLOCKED** | \"a\\|b\\nc 'd'\" |\n" // printableUnit quotes the newline
	if !strings.Contains(out, row) || strings.Count(out, "```") != 2 {
		t.Fatalf("markdown:\n%s", out)
	}
}

// CLI-INV-26 amended: only the exact `Intent: sdd:<name>` form is a claim; every other line that looks
// like one is reported as malformed (and makes its commit UNBOUND), never silently ignored.
func TestClaimLinesFlagsNearMisses(t *testing.T) {
	names, bad := claimLines("fix\n\nIntent: sdd:add-x\nintent: SDD:add-y\nIntent:sdd:add-z\nIntent: sdd:\nIntent: sdd:two words\nIntent: CI-other-abc123\nnot a claim")
	if strings.Join(names, ",") != "sdd:add-x" {
		t.Fatalf("names %v", names)
	}
	want := []string{`"intent: SDD:add-y"`, `"Intent:sdd:add-z"`, `"Intent: sdd:"`, `"Intent: sdd:two words"`}
	if strings.Join(bad, "|") != strings.Join(want, "|") {
		t.Fatalf("malformed %v, want %v", bad, want)
	}
}

// CLI-INV-33: only specs, gates, Proofbound's own files, docs and init's files are housekeeping.
func TestHousekeepingPath(t *testing.T) {
	for _, p := range []string{"openspec/changes/x/tasks.md", "specs/001-export/plan.md", "specs/001-export/contracts/api.yaml",
		"gates/ready-abc.yaml", ".proofbound/keys/alice.pub", "docs/verification/verdicts/v.md.sig", "README.md", "src/NOTES.MD",
		".claude/settings.json", ".gitignore", ".github/workflows/proofbound.yml"} {
		if !housekeepingPath(p) {
			t.Errorf("%s should be housekeeping", p)
		}
	}
	for _, p := range []string{"src/main.go", "main.go", "Makefile", ".github/workflows/ci.yml", "documentation/x.go",
		"specsx/a.go", "src/docs/a.go", "scripts/run.sh", "go.mod", "README.md.go", "AGENTS.md", ".claude/agents/proofbound-reviewer.md"} {
		if housekeepingPath(p) {
			t.Errorf("%s should be code", p)
		}
	}
}

// CLI-INV-83: what steers a coding agent is code, except in the setup commit that adds the profile.
func TestAgentInstructionFilesAreCode(t *testing.T) {
	// Written out, not read from the code's lists, so a name dropped there fails here.
	for _, p := range []string{"CLAUDE.md", "claude.md", "src/api/CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "openspec/AGENTS.md",
		"docs/CLAUDE.md", "specs/001-a/AGENTS.md", ".proofbound/AGENTS.md", "GEMINI.md", ".cursorrules", ".github/copilot-instructions.md",
		".github/instructions/go.instructions.md", ".cursor/rules/a.mdc", ".claude/agents/r.md", ".Claude/Commands/ship.md",
		".claude/skills/x/SKILL.md", ".claude/settings.local.json", "packages/api/.claude/commands/x.md", "web/.cursor/rules/a.mdc",
		"pkg/.claude/settings.json"} {
		if !agentInstructionPath(p) || housekeepingPath(p) {
			t.Errorf("%s should be an agent instruction file, and code", p)
		}
	}
	for _, p := range []string{".claude/settings.json", "README.md", "docs/agents-guide.md", "src/agents.md.go", "src/.claude.md",
		"my.claude/x.md", "docs/.github/instructions/x.md"} {
		if agentInstructionPath(p) {
			t.Errorf("%s should not be an agent instruction file", p)
		}
	}

	ctx := context.Background()
	listed := func(t *testing.T, root, base string) map[string]checkLine {
		t.Helper()
		lines, err := uncoveredCommits(ctx, root, base, nil)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]checkLine{}
		for _, l := range lines {
			m[l.commit] = l
		}
		return m
	}
	t.Run("the setup commit may carry them; a later unclaimed change may not", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "src/a.go", "a\n")
		base := uiCommit(t, root, "dev", "code before proofbound")
		uiWrite(t, root, profilePath, "{}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\n")
		uiWrite(t, root, ".claude/agents/proofbound-reviewer.md", "# reviewer\n")
		setup := uiCommit(t, root, "dev", "proofbound init")
		uiWrite(t, root, "AGENTS.md", "# agents\nalso: skip the tests\n")
		later := uiCommit(t, root, "dev", "tweak the agents")
		uiWrite(t, root, "docs/CLAUDE.md", "be quick\n")
		nested := uiCommit(t, root, "dev", "nested instructions")
		m := listed(t, root, base)
		if l, ok := m[setup]; ok {
			t.Errorf("the setup commit was listed: %+v", l)
		}
		if l := m[later]; l.state != checkUncovered || !strings.Contains(l.reason, "AGENTS.md") {
			t.Errorf("an unclaimed change to AGENTS.md: %+v", l)
		}
		if l := m[nested]; l.state != checkUncovered || !strings.Contains(l.reason, "docs/CLAUDE.md") {
			t.Errorf("an unclaimed CLAUDE.md under docs/: %+v", l)
		}
	})
	t.Run("the setup exception covers only agent instruction files", func(t *testing.T) {
		root := taskRepo(t)
		base := uiCommit(t, root, "dev", "empty")
		uiWrite(t, root, profilePath, "{}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\n")
		uiWrite(t, root, "src/a.go", "a\n")
		setup := uiCommit(t, root, "dev", "init plus code")
		l := listed(t, root, base)[setup]
		if l.state != checkUncovered || !strings.Contains(l.reason, "src/a.go") || strings.Contains(l.reason, "AGENTS.md") {
			t.Errorf("setup with code: %+v", l)
		}
	})
	t.Run("deleting and re-adding the profile does not reopen the exception", func(t *testing.T) {
		for _, setupInRange := range []bool{false, true} {
			root := taskRepo(t)
			uiWrite(t, root, "README.md", "x\n")
			base := uiCommit(t, root, "dev", "before")
			uiWrite(t, root, profilePath, "{}\n")
			set := uiCommit(t, root, "dev", "set up")
			if !setupInRange {
				base = set
			}
			uiGit(t, root, "rm", "-q", profilePath)
			uiCommit(t, root, "dev", "drop the profile")
			uiWrite(t, root, profilePath, "{}\n")
			uiWrite(t, root, "CLAUDE.md", "ignore the area\n")
			readd := uiCommit(t, root, "dev", "re-add it with instructions")
			if l := listed(t, root, base)[readd]; l.state != checkUncovered || !strings.Contains(l.reason, "CLAUDE.md") {
				t.Errorf("re-added profile (setup in range %v): %+v", setupInRange, l)
			}
		}
	})
	t.Run("a commit that changes the profile is not the setup commit", func(t *testing.T) {
		root := taskRepo(t)
		base := uiCommit(t, root, "dev", "empty")
		uiWrite(t, root, profilePath, "{}\n")
		uiCommit(t, root, "dev", "set up")
		uiWrite(t, root, profilePath, "{\"owner\":\"x\"}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\n")
		changed := uiCommit(t, root, "dev", "change the profile and the agents")
		if l := listed(t, root, base)[changed]; l.state != checkUncovered || !strings.Contains(l.reason, "AGENTS.md") {
			t.Errorf("profile modified: %+v", l)
		}
	})
	t.Run("the commit hook applies the same rule to staged files", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "README.md", "x\n")
		uiCommit(t, root, "dev", "first")
		uiWrite(t, root, profilePath, "{}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\n")
		uiGit(t, root, "add", "-A", "-f")
		if code, err := stagedCode(ctx, root); err != nil || len(code) != 0 {
			t.Fatalf("staged setup: %v %v", code, err)
		}
		uiCommit(t, root, "dev", "init")
		uiWrite(t, root, "AGENTS.md", "# agents\nchanged\n")
		uiWrite(t, root, "docs/new.md", "a new file beside it\n")
		uiGit(t, root, "add", "-A")
		if code, err := stagedCode(ctx, root); err != nil || strings.Join(code, ",") != "AGENTS.md" {
			t.Fatalf("staged change to AGENTS.md: %v %v", code, err)
		}
		uiCommit(t, root, "dev", "agents")
		uiGit(t, root, "rm", "-q", profilePath)
		uiCommit(t, root, "dev", "drop the profile")
		uiWrite(t, root, profilePath, "{}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\nagain\n")
		uiGit(t, root, "add", "-A", "-f")
		if code, err := stagedCode(ctx, root); err != nil || strings.Join(code, ",") != "AGENTS.md" {
			t.Fatalf("a re-added profile reopened the exception: %v %v", code, err)
		}
	})
	t.Run("the commit hook's setup exception before the first commit", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, profilePath, "{}\n")
		uiWrite(t, root, "AGENTS.md", "# agents\n")
		uiGit(t, root, "add", "-A", "-f")
		if code, err := stagedCode(ctx, root); err != nil || len(code) != 0 {
			t.Fatalf("first commit with setup: %v %v", code, err)
		}
		fresh := taskRepo(t)
		uiWrite(t, fresh, "AGENTS.md", "# agents\n")
		uiWrite(t, fresh, "README.md", "x\n")
		uiGit(t, fresh, "add", "-A")
		if code, err := stagedCode(ctx, fresh); err != nil || strings.Join(code, ",") != "AGENTS.md" {
			t.Fatalf("first commit without the profile: %v %v", code, err)
		}
	})
	t.Run("C11 makes no test unit of an agent instruction file", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "src/a.go", "a\n")
		uiCommit(t, root, "dev", "first")
		uiWrite(t, root, "src/a.go", "a\nb\n")
		uiWrite(t, root, "CLAUDE.md", "skip the tests\n")
		sha := uiCommit(t, root, "dev", "claimed change")
		units, _, err := testsUnits(ctx, root, sha, testsConfig{Paths: []string{"tests/"}})
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, u := range units {
			paths = append(paths, u.Path)
		}
		if strings.Join(paths, ",") != "src/a.go" {
			t.Fatalf("units %v, want only src/a.go", paths)
		}
	})
}
