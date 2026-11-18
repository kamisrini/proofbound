package sdd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
)

// These tests kill the calibration mutants that survived the R2 M1 reviews
// (r2-m1-review-round1 F8 and r2-m1-recheck-round1 F1-F3), so the package enters the frozen
// P6 Task 8 production set with zero undeclared survivors.

// SDD-INV-65: a mode-only difference (an executable profile) changes the state key. Kills the
// oid-only key mutant.
func TestStateKeyModeOnlyDifference(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	plain := s.commit("plain profile")
	if err := os.Chmod(filepath.Join(s.root, ProfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	exec := s.commit("executable profile, same bytes")
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewGitReader(repo)
	a, _ := r.StateKey(context.Background(), plain)
	b, _ := r.StateKey(context.Background(), exec)
	if a == b {
		t.Fatalf("a mode-only change left the state key equal: %q", a)
	}
}

// SDD-INV-54: a symlink inside a change directory enters the manifest by git identity; retargeting
// it changes the change intent's digest. Kills hashing the (empty) symlink bytes as content.
func TestSymlinkManifestLineIsGitIdentity(t *testing.T) {
	digest := func(target string) string {
		s := newSynth(t)
		s.write("openspec/changes/designlink/specs/core/spec.md", addedDelta("D", "The system SHALL design."))
		s.symlink("openspec/changes/designlink/design.md", target)
		s.commit("link")
		revs, _ := syncAll(t, s.provider())
		r, ok := revs[ciKind+"openspec/changes/designlink"]
		if !ok {
			t.Fatal("no change intent")
		}
		return r.ArtifactSHA256
	}
	if digest("one.md") == digest("two.md") {
		t.Fatal("retargeting a symlink did not change the change manifest")
	}
}

// Re-check F2 of draft 9 (§6 line 5): a regular FILE in a parent's archive never "archives" a
// change. Kills dropping the parent entry type check.
func TestParentArchiveFileDoesNotArchive(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/archive/2026-01-01-late", "a stray file, not a change entry")
	s.commit("stray file in the archive")
	s.remove("openspec/changes/archive/2026-01-01-late")
	s.write("openspec/changes/archive/2026-01-01-late/specs/core/spec.md", addedDelta("Late", "L."))
	archived := s.commit("archive late as a directory")
	if _, err := s.provider().Resolve(context.Background(), archived, "late"); err != nil {
		t.Fatalf("archived by this commit, but got %v", err)
	}
}

// M1 re-check F1: a regular file named specs is ignored, so the change is simply spec-less.
// Kills reverting the specs-position half of the M1 F1 fix.
func TestRegularFileNamedSpecsIsIgnored(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/flat/specs", "a file, not a directory")
	c := s.commit("specs as a file")
	_, skips := syncAll(t, s.provider())
	if len(skips) != 0 {
		t.Fatalf("a regular file named specs became %s", fmtSkips(skips))
	}
	if _, err := s.provider().Resolve(context.Background(), c, "flat"); !errors.Is(err, ErrNoChangeIntent) {
		t.Fatalf("flat: %v", err)
	}
}

type parentsFailReader struct{ Reader }

func (parentsFailReader) Parents(context.Context, string) ([]string, error) {
	return nil, errors.New("parents unreadable")
}

// SDD-INV-59: a Parents failure is a reader error, never a tenant sentinel. Kills mapping it to
// ErrUnknownChange.
func TestParentsFailureIsNotASentinel(t *testing.T) {
	root, shas := replayHistory(t, "openspec-history")
	repo, err := gitcmd.New(root)
	if err != nil {
		t.Fatal(err)
	}
	gr, _ := NewGitReader(repo)
	p, _ := New(parentsFailReader{gr}, timeOne)
	_, err = p.Resolve(context.Background(), shas[5], "add-login") // archived by step 06: reads parents
	var u *UnmappableError
	if err == nil || errors.Is(err, ErrUnknownChange) || errors.Is(err, ErrArchivedEarlier) || errors.Is(err, ErrNoChangeIntent) || errors.As(err, &u) {
		t.Fatalf("a Parents failure surfaced as a tenant outcome: %v", err)
	}
}

// M1 review F9: within one state, Skipped() and Revisions() are in a fixed order. Kills removing
// either sort.
func TestOutputOrderIsDeterministic(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.write("openspec/changes/other/specs/core/spec.md", addedDelta("Other", "O."))
	s.write("openspec/changes/third/specs/core/spec.md", addedDelta("Third", "T."))
	hostile(s)
	s.commit("one state")
	p := s.provider()
	revs, err := p.Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !sort.SliceIsSorted(revs, func(i, j int) bool {
		a, b := revs[i], revs[j]
		return string(a.Kind)+"\x00"+a.NativeID+"\x00"+a.ArtifactSHA256 < string(b.Kind)+"\x00"+b.NativeID+"\x00"+b.ArtifactSHA256
	}) {
		t.Fatal("revisions of one state are not in their fixed order")
	}
	skips := p.Skipped()
	if len(skips) < 5 || !sort.SliceIsSorted(skips, func(i, j int) bool { return skips[i].Unit+"\x00"+skips[i].Cause < skips[j].Unit+"\x00"+skips[j].Cause }) {
		t.Fatalf("skips of one state are not in their fixed order: %v", skips)
	}
}

// SDD-INV-54 / §4.3: the change manifest's line for a symlink is exactly "<mode> <oid>  <path>",
// recomputed here independently from git. Kills any other symlink line format.
func TestSymlinkManifestLineFormatIsPinned(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/designlink/specs/core/spec.md", addedDelta("D", "The system SHALL design."))
	s.symlink("openspec/changes/designlink/design.md", "one.md")
	s.commit("link")
	revs, _ := syncAll(t, s.provider())
	got := revs[ciKind+"openspec/changes/designlink"].ArtifactSHA256
	lsLine := func(p string) (mode, oid string) {
		f := strings.Fields(s.git("ls-tree", "HEAD", "--", p))
		return f[0], f[2]
	}
	blobLine := func(p string) string {
		sum := sha256.Sum256([]byte(s.git("cat-file", "blob", "HEAD:"+p) + "\n"))
		return hex.EncodeToString(sum[:]) + "  " + p
	}
	mode, oid := lsLine("openspec/changes/designlink/design.md")
	lines := []string{
		blobLine(ProfilePath),
		mode + " " + oid + "  openspec/changes/designlink/design.md",
		blobLine("openspec/changes/designlink/specs/core/spec.md"),
	}
	sort.Slice(lines, func(i, j int) bool {
		return strings.SplitN(lines[i], "  ", 2)[1] < strings.SplitN(lines[j], "  ", 2)[1]
	})
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n") + "\n"))
	if want := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("change digest %s is not the independent manifest %s", got, want)
	}
}
