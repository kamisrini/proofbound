package sdd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
)

// TestSDDReaderSharesReadContract proves SDD-INV-51: the reader's commits and parents equal the
// git connector's under a replace graft (no fabricated root), and a grafted or shallow repository
// fails Revisions with ErrShallow instead of being read partially.
func TestSDDReaderSharesReadContract(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/archive/2026-01-01-seed/specs/core/spec.md", addedDelta("Seed", "S."))
	first := s.commit("one")
	s.write("src/a.go", "package a\n")
	second := s.commit("two")
	s.git("replace", "--graft", second)
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewGitReader(repo)
	ctx := context.Background()
	listed, err := r.Commits(ctx)
	if err != nil || !reflect.DeepEqual(listed, []string{first, second}) {
		t.Fatalf("commits=%v err=%v", listed, err)
	}
	if parents, err := r.Parents(ctx, second); err != nil || !reflect.DeepEqual(parents, []string{first}) {
		t.Fatalf("graft fabricated parents: %v %v", parents, err)
	}
	p, _ := New(r, timeOne)
	if _, err := p.Resolve(ctx, second, "seed"); !errors.Is(err, ErrArchivedEarlier) {
		t.Fatalf("a grafted commit must not be a root that archives seed: %v", err)
	}
	grafts := filepath.Join(s.root, ".git", "info", "grafts")
	if err := os.MkdirAll(filepath.Dir(grafts), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(grafts, []byte(second+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Revisions(ctx); !errors.Is(err, gitcmd.ErrShallow) {
		t.Fatalf("grafted repository read: %v", err)
	}
}
