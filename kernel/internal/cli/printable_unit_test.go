package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestPrintableUnitQuotesEveryLineBreaker extends sdd SDD-INV-60 (re-check
// r2-m1-recheck-round1-20261118 F3): C1 controls (NEL, CSI) and U+2028/U+2029 are quoted like C0
// and DEL, and the quoted form carries no raw line breaker; ordinary names, non-ASCII letters
// included, print unchanged.
func TestPrintableUnitQuotesEveryLineBreaker(t *testing.T) {
	for _, raw := range []string{"a\nb", "a\x7fb", "a\u0085b", "a\u009bb", "a\u0080b", "a\u009fb", "a b", "a b", "a\xffb"} {
		got := printableUnit(raw)
		if got == raw || !strings.HasPrefix(got, `"`) {
			t.Errorf("%q printed unquoted: %q", raw, got)
		}
		for _, r := range got {
			if r < 0x20 || (r >= 0x7f && r <= 0x9f) || r == 0x2028 || r == 0x2029 {
				t.Errorf("%q: quoted form %q still carries %U", raw, got, r)
			}
		}
	}
	for _, plain := range []string{"change add-login", "archive 2026-01-01-x", "change café", "change  nbsp"} {
		if got := printableUnit(plain); got != plain {
			t.Errorf("%q altered to %q", plain, got)
		}
	}
}

// TestAllSelectionNeverConstructsSDD extends sdd SDD-INV-56 to `proofbound verify` (round-1 F8):
// verify's intent syncs call syncIntentOnStore with "all" only, and the "all" selection builds
// records and specdir without touching git — it succeeds outside any repository — and never sdd.
func TestAllSelectionNeverConstructsSDD(t *testing.T) {
	providers, err := intentProviders(t.TempDir(), "all", time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatalf("all outside a repository: %v", err)
	}
	var names []string
	for _, p := range providers {
		names = append(names, p.Name())
	}
	if strings.Join(names, ",") != "records,specdir" {
		t.Fatalf("all built %v, want records,specdir", names)
	}
	if _, err := intentProviders(t.TempDir(), "sdd", time.Unix(1, 0).UTC()); err == nil {
		t.Fatal("calibration: sdd outside a repository must fail to construct")
	}
}

// TestAllSelectionExcludesSDDWhereItCouldBeBuilt closes re-check
// r2-m1-lowfix-recheck-20261118 F1: in a full (non-shallow) repository, where sdd CAN be
// constructed, "all" still builds exactly records and specdir.
func TestAllSelectionExcludesSDDWhereItCouldBeBuilt(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "commit", "-q", "--allow-empty", "-m", "root"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	now := time.Unix(1, 0).UTC()
	if _, err := intentProviders(root, "sdd", now); err != nil {
		t.Fatalf("calibration: sdd must be constructible here: %v", err)
	}
	providers, err := intentProviders(root, "all", now)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range providers {
		names = append(names, p.Name())
	}
	if strings.Join(names, ",") != "records,specdir" {
		t.Fatalf("all built %v in a full repository, want records,specdir", names)
	}
}

// TestVerifySyncsIntentWithAllOnly closes the verify half of re-check
// r2-m1-lowfix-recheck-20261118 F1 structurally (running verify needs a ledger): every
// syncIntentOnStore call inside verify passes the literal selection "all", so verify can
// never construct sdd.
func TestVerifySyncsIntentWithAllOnly(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "cli.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "verify" || fn.Recv != nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "syncIntentOnStore" {
				return true
			}
			calls++
			lit, ok := call.Args[2].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || lit.Value != `"all"` {
				t.Errorf("%s: verify syncs intent with selection %s, want \"all\"", fset.Position(call.Pos()), exprText(call.Args[2]))
			}
			return true
		})
	}
	if calls == 0 {
		t.Fatal("calibration: no syncIntentOnStore call found in verify")
	}
}

func exprText(e ast.Expr) string {
	if lit, ok := e.(*ast.BasicLit); ok {
		return lit.Value
	}
	return "a non-literal expression"
}
