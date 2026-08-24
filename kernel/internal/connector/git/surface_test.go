package git

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestExportedSurfaceMatchesTheSpec(t *testing.T) {
	files := packageFiles(t, ".")
	exports := make(map[string]struct{})
	for _, file := range files {
		for _, decl := range file.Decls {
			switch item := decl.(type) {
			case *ast.FuncDecl:
				if ast.IsExported(item.Name.Name) {
					exports[item.Name.Name] = struct{}{}
				}
			case *ast.GenDecl:
				for _, spec := range item.Specs {
					switch value := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(value.Name.Name) {
							exports[value.Name.Name] = struct{}{}
						}
					case *ast.ValueSpec:
						for _, name := range value.Names {
							if ast.IsExported(name.Name) {
								exports[name.Name] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	got := sortedKeys(exports)
	want := []string{"Appender", "Commit", "Connector", "Deps", "New", "Repo", "Result", "Sync", "Version"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("exports=%v want=%v", got, want)
	}
}

func TestExportedStructFieldsMatchTheSpec(t *testing.T) {
	files := packageFiles(t, ".")
	want := map[string][]string{
		"Commit": {"AuthorEmail", "AuthorName", "CitedDecisions", "CommittedAt", "CommitterEmail", "CommitterName", "FilesTouched", "SHA", "Subject"},
		"Deps":   {"IDs", "Logger", "Repo"},
		"Result": {"Appended", "Cursor", "Existing", "Listed"},
	}
	got := make(map[string][]string)
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok || !ast.IsExported(typeSpec.Name.Name) {
				return true
			}
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if ast.IsExported(name.Name) {
						got[typeSpec.Name.Name] = append(got[typeSpec.Name.Name], name.Name)
					}
				}
			}
			sort.Strings(got[typeSpec.Name.Name])
			return false
		})
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields=%v want=%v", got, want)
	}
}

func TestPackagePurity(t *testing.T) {
	for _, file := range packageFiles(t, ".") {
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if path == "database/sql" || path == "os/exec" || strings.Contains(path, "pgx") {
				t.Fatalf("banned import %q", path)
			}
		}
	}
}

func TestNoCursorIsEverRead(t *testing.T) {
	for _, dir := range []string{".", "gitcmd"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			text := string(body)
			for _, banned := range []string{"readCursor", "watermark", `"--since`, `"--after`, `"--until`, `"--before`, `"--max-count`} {
				if strings.Contains(text, banned) {
					t.Fatalf("%s contains %q", entry.Name(), banned)
				}
			}
		}
	}
}

func TestGitcmdDoesNotReachTheDatabase(t *testing.T) {
	entries, err := os.ReadDir("gitcmd")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join("gitcmd", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, banned := range []string{"database/sql", "pgx", "internal/store"} {
			if strings.Contains(string(body), banned) {
				t.Fatalf("%s contains %q", entry.Name(), banned)
			}
		}
	}
}

func packageFiles(t *testing.T, dir string) []*ast.File {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var result []*ast.File
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, file)
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
