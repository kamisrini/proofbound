package sdd_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// LESSON class=embedded-default-port-contention (journal 2026-11-09 and 2026-11-19), compiled: the
// embedded database listens on ONE fixed default port (store.embeddedPort(0) == 55432, a product
// contract), and go test runs packages in parallel. Two packages whose untagged tests open the store
// at that port collide and make `make check` flaky. So every untagged test that can open the store at
// the default port must live in THIS package, where tests run one after another; tests elsewhere pick
// an ephemeral port or carry the integration build tag.

// defaultPortOpeners parses one test file and returns each syntactic way it can open the embedded
// store at the default port:
//   - cli.Run from another package (cli.Run always opens at the default port);
//   - forcing the embedded store with Setenv("DATABASE_URL", "");
//   - a store.Config literal (any layout) without a Port, or with Port: 0, outside package store;
//   - inside package cli, a bare Run or runCommand call in a file that never stubs openStore.
func defaultPortOpeners(name, src string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, src, 0)
	if err != nil {
		return nil, err
	}
	stubsOpenStore := false
	ast.Inspect(file, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "openStore" {
					stubsOpenStore = true
				}
			}
		}
		return true
	})
	var hits []string
	at := func(n ast.Node, what string) { hits = append(hits, fset.Position(n.Pos()).String()+": "+what) }
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CallExpr:
			switch fun := v.Fun.(type) {
			case *ast.SelectorExpr:
				if x, ok := fun.X.(*ast.Ident); ok && x.Name == "cli" && fun.Sel.Name == "Run" {
					at(v, "cli.Run")
				}
				if fun.Sel.Name == "Setenv" && len(v.Args) == 2 && stringLit(v.Args[0]) == "DATABASE_URL" && stringLit(v.Args[1]) == "" {
					at(v, `Setenv("DATABASE_URL", "")`)
				}
			case *ast.Ident:
				if file.Name.Name == "cli" && (fun.Name == "Run" || fun.Name == "runCommand") && !stubsOpenStore {
					at(v, "unstubbed "+fun.Name+" in package cli")
				}
			}
		case *ast.CompositeLit:
			if !isStoreConfig(v.Type, file.Name.Name) {
				return true
			}
			port := ""
			for _, elt := range v.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Port" {
						port = "set"
						if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Value == "0" {
							port = "zero"
						}
					}
				}
			}
			if port != "set" {
				at(v, "store.Config without a non-zero Port")
			}
		}
		return true
	})
	return hits, nil
}

// isStoreConfig matches store.Config used from another package. The store package's own tests are
// exempt: they construct Config to test Config itself (normalization, lock paths, failures before any
// server starts), and the 2026-11-09 fix gave the ones that start a server an ephemeral port.
func isStoreConfig(expr ast.Expr, _ string) bool {
	t, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := t.X.(*ast.Ident)
	return ok && x.Name == "store" && t.Sel.Name == "Config"
}

// stringLit returns a string literal's value, or a sentinel no literal can equal.
func stringLit(e ast.Expr) string {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "\x00not-a-literal"
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "\x00not-a-literal"
	}
	return s
}

// The detector's calibration, kept in the test so it re-runs with every make check (review
// r2-m3-impl-review-spec-laws-20261119 F5): every must-flag probe flags, every must-allow probe does not.
func TestDefaultPortDetectorCalibration(t *testing.T) {
	for name, c := range map[string]struct {
		src  string
		flag bool
	}{
		"cli.Run elsewhere":      {"package gates\nfunc f() { cli.Run(ctx, \"\", nil, nil, nil) }\n", true},
		"empty DATABASE_URL":     {"package cli\nfunc f(t *T) { openStore = nil; t.Setenv(\"DATABASE_URL\", \"\") }\n", true},
		"portless config":        {"package migration\nvar c = store.Config{Root: \"x\"}\n", true},
		"multi-line portless":    {"package migration\nvar c = store.Config{\n\tRoot: \"x\",\n\tDatabaseURL: \"\",\n}\n", true},
		"port zero":              {"package migration\nvar c = store.Config{Root: \"x\", Port: 0}\n", true},
		"store's own Config":     {"package store\nvar c = Config{Root: \"x\"}\n", false},
		"unstubbed runCommand":   {"package cli\nfunc f() { runCommand(ctx, commandSyncGit, nil, root, \"\", out) }\n", true},
		"unstubbed Run":          {"package cli\nfunc f() { Run(ctx, \"\", nil, nil, nil) }\n", true},
		"ported config":          {"package migration\nvar c = store.Config{Root: \"x\", Port: p}\n", false},
		"multi-line ported":      {"package migration\nvar c = store.Config{\n\tRoot: \"x\",\n\tPort: freePort(t),\n}\n", false},
		"stubbed runCommand":     {"package cli\nfunc f() { openStore = stub; runCommand(ctx, commandSyncGit, nil, root, \"\", out) }\n", false},
		"qualified other Run":    {"package cli\nfunc f() { openStore = stub; twin.Run(ctx) }\n", false},
		"non-empty DATABASE_URL": {"package gates\nfunc f(t *T) { t.Setenv(\"DATABASE_URL\", url) }\n", false},
		"other store type":       {"package migration\nvar c = store.Filter{Limit: 1}\n", false},
	} {
		hits, err := defaultPortOpeners(name+".go", c.src)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if (len(hits) > 0) != c.flag {
			t.Errorf("%s: hits=%v, want flagged=%v", name, hits, c.flag)
		}
	}
}

func TestDefaultPortStoreTestsLiveInOnePackage(t *testing.T) {
	kernel, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	home, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var offenders []string
	homeHits := 0
	err = filepath.WalkDir(kernel, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(b), "//go:build integration") {
			return nil
		}
		hits, err := defaultPortOpeners(p, string(b))
		if err != nil {
			return err
		}
		if filepath.Dir(p) == home {
			homeHits += len(hits)
			return nil
		}
		offenders = append(offenders, hits...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Fatalf("untagged tests outside %s can open the embedded store at the default port:\n%s", home, strings.Join(offenders, "\n"))
	}
	// Inertness guard: the detector must see this package's own default-port openers.
	if homeHits < 2 {
		t.Fatalf("detector inert: found %d default-port openers in this package", homeHits)
	}
}
