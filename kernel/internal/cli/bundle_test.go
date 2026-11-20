package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func bundleIx(states ...string) bundleIndex {
	ix := bundleIndex{Base: strings.Repeat("b", 40), Head: strings.Repeat("c", 40), ExporterKey: "0123456789abcdef"}
	for i, s := range states {
		ix.Entries = append(ix.Entries, bundleEntry{Commit: strings.Repeat(string(rune('d'+i)), 40), Claims: "sdd:x", State: s})
	}
	return ix
}

// CLI-INV-29: the badge's text and colour follow the index states; its title names the range and says
// it is unsigned; same index, same bytes.
func TestBundleBadge(t *testing.T) {
	cases := []struct {
		states      []string
		text, color string
	}{
		{nil, "no claims", "#9f9f9f"},
		{[]string{bundleValid, bundleValid}, "2 of 2 claims verified", "#4c1"},
		{[]string{bundleValid, bundleUnreviewed}, "1 of 2 claims verified", "#dfb317"},
		{[]string{bundleValid, bundleNotExportable}, "1 of 2 claims verified", "#e05d44"},
		{[]string{bundleUnreviewed, checkNoReview}, "0 of 2 claims verified", "#e05d44"},
		{[]string{checkBlocked}, "0 of 1 claims verified", "#e05d44"},
	}
	for _, c := range cases {
		ix := bundleIx(c.states...)
		svg := bundleBadge(ix)
		if !strings.Contains(svg, ">"+c.text+"</text>") || !strings.Contains(svg, `fill="`+c.color+`"`) {
			t.Errorf("%v: %s", c.states, svg)
		}
		if !strings.Contains(svg, "<title>proofbound: "+c.text+" for bbbbbbbbbbbb..cccccccccccc (unsigned local result; verify the warranty files)</title>") {
			t.Errorf("%v: title: %s", c.states, svg)
		}
		if bundleBadge(ix) != svg {
			t.Errorf("%v: not deterministic", c.states)
		}
	}
}

// CLI-INV-28 (page): untrusted cells cannot break the table or open a code span; the page says it is
// unsigned, names the range and the uncovered commits, and never asks the reader to trust its key.
func TestVerifyPageEscapesAndWarns(t *testing.T) {
	ix := bundleIx(checkBlocked)
	ix.NonClaimingCommits = 4
	ix.Entries[0].Claims = "sdd:x|`[run](http://evil)`"
	ix.Entries[0].Reason = "gate said\n| forged | row |"
	page := verifyPage(ix)
	var row string
	for _, line := range strings.Split(page, "\n") {
		if strings.HasPrefix(line, "| `dddddddddddd`") {
			row = line
		}
	}
	if row == "" || strings.Count(strings.ReplaceAll(row, `\|`, ""), "|") != 7 || strings.Count(row, "`") != 2 {
		t.Fatalf("row broke out of its cells: %q\n%s", row, page)
	}
	for _, want := range []string{"**This page and `index.json` are not signed.**", "`" + ix.Base + ".." + ix.Head + "`", "4 commit(s) in this range claim none", "**out of band**", "--trust-key <fingerprint you obtained>"} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	if verifyPage(ix) != page {
		t.Error("page not deterministic")
	}
}

// CLI-INV-30: refusals happen before anything is written, and a failed write leaves nothing behind.
func TestBundleRefusesAndWritesAtomically(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@example.invalid"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q")
	run("commit", "-q", "--allow-empty", "-m", "one")

	full := filepath.Join(root, "out")
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(full, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := warrantyBundle(context.Background(), root, nil, "HEAD", full, &out); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("non-empty dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "gates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "gates", "ready-x.yaml"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh := filepath.Join(root, "fresh")
	if err := warrantyBundle(context.Background(), root, nil, "HEAD", fresh, &out); err == nil || !strings.Contains(err.Error(), "differ from HEAD") {
		t.Fatalf("dirty gates: %v", err)
	}
	if _, err := os.Lstat(fresh); err == nil {
		t.Fatal("a refused bundle created its directory")
	}

	parent := t.TempDir()
	dir := filepath.Join(parent, "bundle")
	if err := writeDirAtomically(dir, map[string][]byte{"ok": []byte("1"), "missing/sub": []byte("2")}); err == nil {
		t.Fatal("a failing write succeeded")
	}
	if entries, _ := os.ReadDir(parent); len(entries) != 0 {
		t.Fatalf("a failed write left %v", entries)
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeDirAtomically(dir, map[string][]byte{"a": []byte("1")}); err != nil {
		t.Fatalf("an empty existing dir was not replaced: %v", err)
	}
	if entries, _ := os.ReadDir(parent); len(entries) != 1 {
		t.Fatalf("leftovers: %v", entries)
	}
}
