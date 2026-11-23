package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

// CLI-INV-64: a warranty cites only the requirement reviews whose artifact is a regular file
// committed at HEAD: not one only in the working tree, on another branch, deleted at HEAD, a
// symlink, or a directory.
func TestReviewsAtHead(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	dir := "docs/verification/verdicts/"
	uiWrite(t, root, dir+"RR-a-1.md", "a\n")
	uiWrite(t, root, dir+"RR-gone-1.md", "g\n")
	uiWrite(t, root, dir+"RR-dir-1.md/x", "d\n")
	if err := os.Symlink("RR-a-1.md", filepath.Join(root, dir, "RR-link-1.md")); err != nil {
		t.Fatal(err)
	}
	uiCommit(t, root, "a", "reviews")
	uiGit(t, root, "checkout", "-q", "-b", "other")
	uiWrite(t, root, dir+"RR-other-1.md", "o\n")
	uiCommit(t, root, "a", "other branch")
	uiGit(t, root, "checkout", "-q", "main")
	uiGit(t, root, "rm", "-q", dir+"RR-gone-1.md")
	uiCommit(t, root, "a", "drop one")
	uiWrite(t, root, dir+"RR-untracked-1.md", "u\n")

	var all []reviews.RequirementReview
	for _, name := range []string{"RR-a-1.md", "RR-gone-1.md", "RR-dir-1.md", "RR-link-1.md", "RR-other-1.md", "RR-untracked-1.md", "RR-missing-1.md"} {
		all = append(all, reviews.RequirementReview{ReviewID: name, ArtifactPath: dir + name})
	}
	kept := reviewsAtHead(ctx, root, all)
	if len(kept) != 1 || kept[0].ReviewID != "RR-a-1.md" {
		t.Fatalf("kept %+v, want only RR-a-1.md", kept)
	}
}
