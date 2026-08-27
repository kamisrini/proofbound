package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type mutant struct {
	file          string
	ordinal       int
	before, after string
	line, pos     int
}

var testTags string
var mutationPattern = "./..."

func main() {
	pkg := flag.String("pkg", "", "kernel package directory, e.g. internal/store")
	rootFlag := flag.String("root", "", "repository root")
	tags := flag.String("tags", "", "go test build tags")
	flag.Parse()
	testTags = *tags
	if *pkg == "" {
		fail("-pkg is required")
	}
	base := *rootFlag
	if base == "" {
		base, _ = os.Getwd()
	}
	root, err := filepath.Abs(filepath.Join(base, "kernel"))
	if err != nil {
		fail(err.Error())
	}
	target := filepath.Join(root, filepath.FromSlash(*pkg))
	if _, err := os.Stat(target); err != nil {
		fail(err.Error())
	}
	mutants := collect(target)
	if len(mutants) == 0 {
		fail("no mutation candidates found")
	}
	mutationPattern = "./" + filepath.ToSlash(*pkg)
	if err := calibrate(root, *pkg); err != nil {
		fail("calibration: " + err.Error())
	}
	mutationPattern = "./" + filepath.ToSlash(*pkg)
	fmt.Println("calibration neutral=survived invalid=invalid lethal=killed")
	var killed, invalid, survived int
	for _, m := range mutants {
		status := runMutant(root, m)
		switch status {
		case "killed":
			killed++
		case "invalid":
			invalid++
		default:
			survived++
		}
		fmt.Printf("%s:%d#%d %s\n", filepath.ToSlash(m.file), m.line, m.ordinal, status)
	}
	fmt.Printf("summary candidates=%d killed=%d invalid=%d survived=%d\n", len(mutants), killed, invalid, survived)
	if survived > 0 {
		os.Exit(1)
	}
}

func collect(dir string) []mutant {
	var out []mutant
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(b)
		ordinal := 0
		for _, op := range []struct{ before, after string }{{"==", "!="}, {"!=", "=="}, {"&&", "||"}, {"||", "&&"}} {
			start := 0
			for {
				i := strings.Index(text[start:], op.before)
				if i < 0 {
					break
				}
				i += start
				ordinal++
				out = append(out, mutant{file: path, ordinal: ordinal, before: op.before, after: op.after, line: 1 + bytes.Count(b[:i], []byte("\n")), pos: i})
				start = i + len(op.before)
			}
		}
		return nil
	})
	return out
}

func runMutant(root string, m mutant) string {
	tmp, err := os.MkdirTemp("", "vera-mutant-")
	if err != nil {
		return "invalid"
	}
	defer os.RemoveAll(tmp)
	copyTree(root, tmp)
	rel, _ := filepath.Rel(root, m.file)
	path := filepath.Join(tmp, rel)
	b, err := os.ReadFile(path)
	if err != nil {
		return "invalid"
	}
	needle := []byte(m.before)
	pos := m.pos
	if pos < 0 || pos+len(needle) > len(b) || string(b[pos:pos+len(needle)]) != m.before {
		return "invalid"
	}
	b = append(append(append([]byte{}, b[:pos]...), []byte(m.after)...), b[pos+len(needle):]...)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "invalid"
	}
	args := testArgs(mutationPattern)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = tmp
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/vera-mutant-cache")
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "killed"
	}
	if err == nil {
		return "survived"
	}
	if bytes.Contains(stderr.Bytes(), []byte("undefined")) || bytes.Contains(stderr.Bytes(), []byte("syntax error")) {
		return "invalid"
	}
	return "killed"
}

func calibrate(root, pkg string) error {
	tmp, err := os.MkdirTemp("", "vera-calibration-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	copyTree(root, tmp)
	kernel := tmp
	target := filepath.Join(kernel, filepath.FromSlash(pkg))
	neutral, packageName, err := calibrationSource(target)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(neutral)
	if err != nil {
		return err
	}
	original := append([]byte(nil), b...)
	if err = os.WriteFile(neutral, append(b, []byte("\n// neutral calibration mutant\n")...), 0o644); err != nil {
		return err
	}
	if status := runTests(tmp); status != "survived" {
		return fmt.Errorf("neutral mutant did not survive: %s", status)
	}
	invalid := neutral
	b, err = os.ReadFile(invalid)
	if err != nil {
		return err
	}
	if err = os.WriteFile(invalid, append(b, []byte("\nvar _ = undefinedCalibrationSymbol\n")...), 0o644); err != nil {
		return err
	}
	if status := runTests(tmp); status != "invalid" {
		return fmt.Errorf("invalid mutant was not classified invalid")
	}
	if err = os.WriteFile(invalid, original, 0o644); err != nil {
		return err
	}
	dead := filepath.Join(target, "zz_calibration_test.go")
	if err = os.WriteFile(dead, []byte("package "+packageName+"\nimport \"testing\"\nfunc TestCalibrationLethal(t *testing.T){ t.Fatal(\"lethal calibration\") }\n"), 0o644); err != nil {
		return err
	}
	if status := runTests(tmp); status != "killed" {
		return fmt.Errorf("lethal mutant was not killed")
	}
	return nil
}

func calibrationSource(dir string) (string, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == "package" {
				return path, fields[1], nil
			}
		}
	}
	return "", "", fmt.Errorf("no non-test Go source in %s", dir)
}
func runTests(dir string) string {
	args := testArgs(mutationPattern)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/vera-mutant-cache")
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return "killed"
	}
	if err == nil {
		return "survived"
	}
	if bytes.Contains(stderr.Bytes(), []byte("undefined")) || bytes.Contains(stderr.Bytes(), []byte("syntax error")) {
		return "invalid"
	}
	_ = os.WriteFile("/tmp/vera-mutant-last.log", stderr.Bytes(), 0o644)
	return "killed"
}

func testArgs(pattern string) []string {
	// Mutation and calibration runs must execute the copied tree, not reuse a
	// cached result from a previous scratch tree (especially the lethal control).
	args := []string{"test", "-count=1"}
	if testTags != "" {
		// Integration packages share the configured disposable database. Running them
		// concurrently makes one package's rows invalidate another package's assertions.
		args = append(args, "-p=1", "-tags", testTags)
	}
	return append(args, pattern)
}

func testTimeout() time.Duration {
	if testTags != "" {
		return 2 * time.Minute
	}
	return 1 * time.Minute
}

func copyTree(src, dst string) {
	_ = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			_ = os.MkdirAll(out, 0o755)
		} else {
			b, e := os.ReadFile(path)
			if e == nil {
				_ = os.WriteFile(out, b, 0o644)
			}
		}
		return nil
	})
	if filepath.Base(src) == "kernel" {
		if data, err := os.ReadFile(filepath.Join(filepath.Dir(src), "Makefile")); err == nil {
			_ = os.WriteFile(filepath.Join(dst, "Makefile"), data, 0o644)
		}
	}
}
func fail(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(2) }
