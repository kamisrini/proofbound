package warranty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// KeysDir is where a repository declares the public keys that speak for it (D3).
const KeysDir = ".proofbound/keys"

// ErrUnverified is returned by Verify when no repository was given: the signature and rule hold, but
// nothing ties the claims to real bytes (W-INV-4). It is never a pass.
var ErrUnverified = errors.New("signature only; pass --repo to check the claims")

// Verify runs SPEC §5 on a parsed file. Without a repository it stops after the file-only checks and
// returns ErrUnverified. trustKey, when set, pins the signing key's fingerprint.
func Verify(ctx context.Context, f File, repo, trustKey string) error {
	_, err := VerifyReport(ctx, f, repo, trustKey)
	return err
}

// VerifyReport is Verify that also returns the reviewers check 9 proved (W-INV-9).
func VerifyReport(ctx context.Context, f File, repo, trustKey string) ([]ProvenReviewer, error) {
	if err := f.verifyThrough8(ctx, repo, trustKey); err != nil {
		return nil, err
	}
	r, err := gitcmd.New(repo)
	if err != nil {
		return nil, err
	}
	head, err := git(repo, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return nil, errors.New("clone has no HEAD commit")
	}
	pub, err := f.PublicKey()
	if err != nil {
		return nil, err
	}
	return reviewersProven(ctx, r, head, f, pub)
}

func (f File) verifyThrough8(ctx context.Context, repo, trustKey string) error {
	if err := f.CheckSignature(); err != nil {
		return err
	}
	if err := f.checkCosignatures(); err != nil {
		return err
	}
	if err := CheckRule(f.Body); err != nil {
		return err
	}
	pub, err := f.PublicKey()
	if err != nil {
		return err
	}
	if trustKey != "" && Fingerprint(pub) != trustKey {
		return fmt.Errorf("fingerprint mismatch: key %s, expected %s", Fingerprint(pub), trustKey)
	}
	if repo == "" {
		return ErrUnverified
	}
	r, err := gitcmd.New(repo)
	if err != nil {
		return err
	}
	commit := f.Body.Commit
	if _, err := r.Parents(ctx, commit); err != nil {
		return fmt.Errorf("commit %s not in clone", commit)
	}
	head, err := git(repo, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return errors.New("clone has no HEAD commit")
	}
	if _, err := git(repo, "merge-base", "--is-ancestor", commit, head); err != nil {
		return fmt.Errorf("HEAD does not descend from commit %s", commit)
	}
	if err := keyDeclared(ctx, r, head, PubLine(pub)); err != nil {
		return err
	}
	if err := rederive(ctx, r, f.Body); err != nil {
		return err
	}
	return artifactsAtHead(ctx, r, head, f.Body)
}

// git runs one read-only git command against the clone with the caller's GIT_* variables removed,
// so the clone, not the environment, decides what is read.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func keyDeclared(ctx context.Context, r *gitcmd.Repo, head, line string) error {
	entries, err := r.TreeEntries(ctx, head, KeysDir, true, false)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Type != "blob" || e.Mode == "120000" || !strings.HasSuffix(e.Path, ".pub") {
			continue
		}
		data, err := r.Blob(ctx, e.OID)
		if err != nil {
			return err
		}
		if string(data) == line {
			return nil
		}
	}
	return errors.New("key not declared by the repository")
}

func canonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return core.Canonicalize(raw)
}

func sameJSON(a, b any) bool {
	x, errA := canonicalJSON(a)
	y, errB := canonicalJSON(b)
	return errA == nil && errB == nil && bytes.Equal(x, y)
}

// rederive maps the commit's own state from git and requires the body's change intent and
// requirements to be exactly those records (W-INV-5, SDD-INV-72).
func rederive(ctx context.Context, r *gitcmd.Repo, b Body) error {
	reader, err := sdd.NewGitReader(r)
	if err != nil {
		return err
	}
	p, err := sdd.New(reader, time.Unix(1, 0).UTC())
	if err != nil {
		return err
	}
	change, reqs, err := p.StateRecords(ctx, b.Commit, b.ChangeName)
	if err != nil {
		return fmt.Errorf("spec at %s: %v", b.Commit, err)
	}
	var ci connectorintent.ChangeIntent
	if err := json.Unmarshal(change.Payload, &ci); err != nil {
		return err
	}
	if !sameJSON(ci, b.ChangeIntent) {
		return fmt.Errorf("change intent differs from the spec at %s", b.Commit)
	}
	derived := make([]connectorintent.Requirement, 0, len(reqs))
	for _, rev := range reqs {
		var req connectorintent.Requirement
		if err := json.Unmarshal(rev.Payload, &req); err != nil {
			return err
		}
		derived = append(derived, req)
	}
	body := append([]connectorintent.Requirement(nil), b.Requirements...)
	sort.Slice(body, func(i, j int) bool { return body[i].RequirementID < body[j].RequirementID })
	if !sameJSON(derived, body) {
		return fmt.Errorf("requirements differ from the spec at %s", b.Commit)
	}
	return nil
}

func blobAtHead(ctx context.Context, r *gitcmd.Repo, head, path string) ([]byte, error) {
	entries, err := r.TreeEntries(ctx, head, path, false, false)
	if err != nil {
		return nil, err
	}
	if len(entries) != 1 || entries[0].Type != "blob" || entries[0].Mode == "120000" {
		return nil, fmt.Errorf("artifact %s is not committed at HEAD", path)
	}
	return r.Blob(ctx, entries[0].OID)
}

// artifactsAtHead requires every cited review and verdict to be committed at HEAD and to parse to
// exactly the record the body carries (W-INV-5).
func artifactsAtHead(ctx context.Context, r *gitcmd.Repo, head string, b Body) error {
	for _, want := range b.Reviews {
		data, err := blobAtHead(ctx, r, head, want.ArtifactPath)
		if err != nil {
			return err
		}
		got, err := reviews.ParseRequirementReview(want.ArtifactPath, data)
		if err != nil {
			return fmt.Errorf("artifact %s: %v", want.ArtifactPath, err)
		}
		if !sameJSON(got, want) {
			return fmt.Errorf("artifact %s differs from the warranty", want.ArtifactPath)
		}
	}
	for _, want := range b.Verdicts {
		data, err := blobAtHead(ctx, r, head, want.ArtifactPath)
		if err != nil {
			return err
		}
		got, err := reviews.ParseObligationVerdict(want.ArtifactPath, data)
		if err != nil {
			return fmt.Errorf("artifact %s: %v", want.ArtifactPath, err)
		}
		if !sameJSON(got, want) {
			return fmt.Errorf("artifact %s differs from the warranty", want.ArtifactPath)
		}
	}
	return nil
}
