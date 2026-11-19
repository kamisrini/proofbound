package cli

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
)

type namedResolver struct {
	commits []string
}

func (r *namedResolver) Resolve(_ context.Context, commit, name string) (connectorintent.Revision, error) {
	r.commits = append(r.commits, commit)
	if name == "add-auth" {
		return connectorintent.Revision{NativeID: "CI-add-auth-xuuno2"}, nil
	}
	return connectorintent.Revision{}, errors.New("unknown change")
}

// CLI-INV-17: `report intent sdd:<name>` resolves the name at HEAD; other ids pass through.
func TestReportIntentResolvesAnSDDChangeName(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "--allow-empty", "-m", "x"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	head, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	resolver := &namedResolver{}
	old := newSDDClaimProvider
	t.Cleanup(func() { newSDDClaimProvider = old })
	newSDDClaimProvider = func(string, time.Time) (sddClaimProvider, error) { return resolver, nil }

	ctx := context.Background()
	if id, err := reportIntentID(ctx, root, "sdd:add-auth"); err != nil || id != "CI-add-auth-xuuno2" {
		t.Fatalf("sdd:add-auth -> %q, %v", id, err)
	}
	if len(resolver.commits) != 1 || resolver.commits[0] != strings.TrimSpace(string(head)) {
		t.Fatalf("resolved at %v, want HEAD %s", resolver.commits, head)
	}
	if id, err := reportIntentID(ctx, root, "sdd:nope"); err == nil || id != "" {
		t.Fatalf("an unresolvable name reported %q, %v", id, err)
	}
	if id, err := reportIntentID(ctx, root, "CI-plain-abcdef"); err != nil || id != "CI-plain-abcdef" || len(resolver.commits) != 2 {
		t.Fatalf("a plain id was not passed through untouched: %q, %v, %d resolves", id, err, len(resolver.commits))
	}
}
