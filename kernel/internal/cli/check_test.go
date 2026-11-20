package cli

import (
	"bytes"
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
