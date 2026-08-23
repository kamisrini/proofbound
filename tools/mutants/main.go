package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type mutant struct {
	file          string
	ordinal       int
	before, after string
	line, pos     int
}

func main() {
	pkg := flag.String("pkg", "", "kernel package directory, e.g. internal/store")
	rootFlag := flag.String("root", "", "repository root")
	flag.Parse()
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
	if err := calibrate(root); err != nil {
		fail("calibration: " + err.Error())
	}
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
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = tmp
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/vera-mutant-cache")
	err = cmd.Run()
	if err == nil {
		return "survived"
	}
	if bytes.Contains(stderr.Bytes(), []byte("undefined")) || bytes.Contains(stderr.Bytes(), []byte("syntax error")) {
		return "invalid"
	}
	return "killed"
}

func calibrate(root string) error {
	tmp, err := os.MkdirTemp("", "vera-calibration-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	copyTree(root, tmp)
	kernel := tmp
	neutral := filepath.Join(kernel, "internal", "store", "doc.go")
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
	invalid := filepath.Join(kernel, "internal", "store", "doc.go")
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
	dead := filepath.Join(kernel, "internal", "store", "zz_calibration_test.go")
	if err = os.WriteFile(dead, []byte("package store\nimport \"testing\"\nfunc TestCalibrationLethal(t *testing.T){ t.Fatal(\"lethal calibration\") }\n"), 0o644); err != nil {
		return err
	}
	if status := runTests(tmp); status != "killed" {
		return fmt.Errorf("lethal mutant was not killed")
	}
	return nil
}
func runTests(dir string) string {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/vera-mutant-cache")
	if err := cmd.Run(); err == nil {
		return "survived"
	}
	if bytes.Contains(stderr.Bytes(), []byte("undefined")) || bytes.Contains(stderr.Bytes(), []byte("syntax error")) {
		return "invalid"
	}
	_ = os.WriteFile("/tmp/vera-mutant-last.log", stderr.Bytes(), 0o644)
	return "killed"
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
}
func fail(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(2) }
