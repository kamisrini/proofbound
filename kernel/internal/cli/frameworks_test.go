package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

var updateFrameworkDoc = flag.Bool("update", false, "rewrite docs/compliance/FRAMEWORK-MAP.md")

// CLI-INV-65: every control maps to at least one criterion of the four frameworks, each criterion in
// its framework's ID format; nothing maps to EU AI Act Art. 12 or Art. 14; the page shows each entry
// with its reason and the caveat, and the transcript carries maps_to.
func TestFrameworkMap(t *testing.T) {
	formats := map[string]*regexp.Regexp{
		fwSOC2:  regexp.MustCompile(`^CC[1-9]\.[1-9]$`),
		fwISO:   regexp.MustCompile(`^A\.[5-8]\.[1-9][0-9]?$`),
		fwSSDF:  regexp.MustCompile(`^(PO|PS|PW|RV)\.[1-9]$`),
		fwEUAIA: regexp.MustCompile(`^(Art\. (11 and Annex IV\(6\)|17\(1\)\([a-m]\))|Annex IV\(2\)\(g\))$`),
	}
	if len(frameworkMap) != len(controlQuestions) {
		t.Fatalf("%d mapped controls, %d controls", len(frameworkMap), len(controlQuestions))
	}
	for id := range controlQuestions {
		entries := frameworkMap[id]
		if len(entries) == 0 {
			t.Errorf("%s maps to nothing", id)
		}
		seen := map[frameworkRef]bool{}
		for _, e := range entries {
			re, ok := formats[e.Framework]
			if !ok || !re.MatchString(e.Criterion) {
				t.Errorf("%s: %q %q is not a known framework's criterion", id, e.Framework, e.Criterion)
			}
			if strings.Contains(e.Criterion, "Art. 12") || strings.Contains(e.Criterion, "Art. 14") {
				t.Errorf("%s maps to %s, which the spec excludes", id, e.Criterion)
			}
			if strings.TrimSpace(e.Reason) == "" || seen[e.frameworkRef] {
				t.Errorf("%s: %s %s has no reason or is listed twice", id, e.Framework, e.Criterion)
			}
			seen[e.frameworkRef] = true
			// Review F1: a reason states the evidence, never the outcome.
			for _, w := range []string{"proves", "ensures", "protects", "guarantee", "complete", "compliant", "satisf"} {
				if strings.Contains(strings.ToLower(e.Reason), w) {
					t.Errorf("%s %s %s: reason claims an outcome (%q): %s", id, e.Framework, e.Criterion, w, e.Reason)
				}
			}
		}
	}

	root := taskRepo(t)
	uiCommit(t, root, "a", "one")
	head := uiCommit(t, root, "a", "two")
	rep := auditReport{Schema: auditSchema, Base: head, Head: head, Counts: auditCounts{Total: 1}, Exceptions: []auditException{}}
	for i := 1; i <= len(controlQuestions); i++ {
		id := fmt.Sprintf("C%d", i)
		rep.Controls = append(rep.Controls, auditControl{ID: id, Question: controlQuestions[id], State: auditHeld, Evidence: []auditEvidence{}, MapsTo: mapsTo(id)})
	}
	rep.Headline = auditHeadline(rep)
	s := testUIServer(t, root, rep)
	host := "127.0.0.1:4242"
	html := uiRequest(s, "GET", "/audit", host, "", s.token, "").Body.String()
	for _, c := range rep.Controls {
		for _, e := range frameworkMap[c.ID] {
			want := fmt.Sprintf(`<span title="%s">%s %s</span>`, templateEscape(e.Reason), e.Framework, e.Criterion)
			if !strings.Contains(html, want) {
				t.Errorf("the page lacks %s's entry %s", c.ID, want)
			}
		}
	}
	if !strings.Contains(html, templateEscape(frameworkCaveat)) || !strings.Contains(html, frameworkDocURL) {
		t.Error("the page lacks the caveat or the link to the mapping")
	}
	tr, err := transcript(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tr), `"maps_to":[{"criterion":"CC8.1","framework":"SOC 2"}`) {
		t.Errorf("the transcript does not carry maps_to: %s", tr)
	}
}

// templateEscape is html/template's escaping of text in an attribute or element.
func templateEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "'", "&#39;", "<", "&lt;", ">", "&gt;", `"`, "&#34;").Replace(s)
}

// CLI-INV-65: docs/compliance/FRAMEWORK-MAP.md is generated from the table; it differs only if the
// table changed without -update.
func TestFrameworkMapDoc(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "compliance", "FRAMEWORK-MAP.md")
	want := frameworkDoc()
	if *updateFrameworkDoc {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v; run `cd kernel && go test ./internal/cli -run TestFrameworkMapDoc -update`", err)
	}
	if string(got) != want {
		t.Fatal("docs/compliance/FRAMEWORK-MAP.md is stale; run `cd kernel && go test ./internal/cli -run TestFrameworkMapDoc -update`")
	}
}

// CLI-INV-65 wiring: computeAudit, the one producer of the Audit room's report, attaches each
// control's mapping.
func TestAuditCarriesFrameworkMap(t *testing.T) {
	ctx := context.Background()
	ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: ephemeralPort(t), DatabaseURL: os.Getenv("DATABASE_URL")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ledger.Close() })
	root := taskRepo(t)
	base := uiCommit(t, root, "a", "one")
	uiWrite(t, root, "README.md", "x\n")
	uiCommit(t, root, "a", "two")
	rep, err := computeAudit(ctx, root, ledger, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Controls) != len(controlQuestions) {
		t.Fatalf("%d controls", len(rep.Controls))
	}
	for _, c := range rep.Controls {
		if fmt.Sprint(c.MapsTo) != fmt.Sprint(mapsTo(c.ID)) || len(c.MapsTo) == 0 || c.Question != controlQuestions[c.ID] {
			t.Errorf("%s: maps_to %v, question %q", c.ID, c.MapsTo, c.Question)
		}
	}
}
