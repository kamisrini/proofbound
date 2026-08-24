package gitcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
)

var ErrShallow = errors.New("gitcmd: repository history is incomplete")

type Repo struct {
	root string
}

func New(root string) (*Repo, error) {
	if root == "" {
		return nil, errors.New("gitcmd: repository root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: repository root: %w", err)
	}
	r := &Repo{root: abs}
	inside, err := r.run(context.Background(), "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: not a work tree: %w", err)
	}
	if strings.TrimSpace(string(inside)) != "true" {
		return nil, errors.New("gitcmd: repository is not a work tree")
	}
	if err := r.refusePartial(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repo) Commits(ctx context.Context) ([]connectorgit.Commit, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	if err := r.refusePartial(ctx); err != nil {
		return nil, err
	}
	if err := r.validateRefs(ctx); err != nil {
		return nil, err
	}
	out, err := r.run(ctx, "rev-list", "--reverse", "--date-order",
		"--exclude=refs/stash", "--exclude=refs/notes/*", "--exclude=refs/replace/*", "--all")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list commits: %w", err)
	}
	shas := strings.Fields(string(out))
	commits := make([]connectorgit.Commit, 0, len(shas))
	for _, sha := range shas {
		commit, readErr := r.readCommit(ctx, sha)
		if readErr != nil {
			return nil, readErr
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (r *Repo) Tips(ctx context.Context) (map[string]string, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	if err := r.refusePartial(ctx); err != nil {
		return nil, err
	}
	if err := r.validateRefs(ctx); err != nil {
		return nil, err
	}
	out, err := r.run(ctx, "for-each-ref", "--format=%(refname)%00")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list refs: %w", err)
	}
	tips := make(map[string]string)
	for _, raw := range bytes.Split(out, []byte{0}) {
		ref := strings.TrimSpace(string(raw))
		if ref == "" || excludedRef(ref) {
			continue
		}
		sha, resolveErr := r.peelCommit(ctx, ref)
		if resolveErr != nil {
			continue
		}
		tips[ref] = sha
	}
	if sha, resolveErr := r.peelCommit(ctx, "HEAD"); resolveErr == nil {
		tips["HEAD"] = sha
	}
	return tips, nil
}

func (r *Repo) readCommit(ctx context.Context, sha string) (connectorgit.Commit, error) {
	format := "%H%x00%an%x00%ae%x00%cn%x00%ce%x00%cI%x00%s%x00%B%x00"
	out, err := r.run(ctx, "show", "-s", "--format="+format, sha)
	if err != nil {
		return connectorgit.Commit{}, fmt.Errorf("gitcmd: read commit %q: %w", sha, err)
	}
	fields, err := parseScalars(out)
	if err != nil {
		return connectorgit.Commit{}, fmt.Errorf("gitcmd: commit %q: %w", sha, err)
	}
	committedAt, err := time.Parse(time.RFC3339, string(fields[5]))
	if err != nil {
		return connectorgit.Commit{}, fmt.Errorf("gitcmd: commit %q committer date: %w", sha, err)
	}
	files, err := r.files(ctx, sha)
	if err != nil {
		return connectorgit.Commit{}, err
	}
	citations, err := r.citations(ctx, sha, string(fields[7]))
	if err != nil {
		return connectorgit.Commit{}, err
	}
	return connectorgit.Commit{
		SHA:            string(fields[0]),
		AuthorName:     string(fields[1]),
		AuthorEmail:    string(fields[2]),
		CommitterName:  string(fields[3]),
		CommitterEmail: string(fields[4]),
		CommittedAt:    committedAt,
		Subject:        string(fields[6]),
		FilesTouched:   files,
		CitedDecisions: citations,
	}, nil
}

func parseScalars(out []byte) ([][]byte, error) {
	out = bytes.TrimSuffix(out, []byte{'\n'})
	fields := bytes.Split(out, []byte{0})
	if len(fields) != 9 || len(fields[8]) != 0 {
		return nil, errors.New("output framing is invalid")
	}
	return fields, nil
}

func (r *Repo) files(ctx context.Context, sha string) ([]string, error) {
	parentsOut, err := r.run(ctx, "show", "-s", "--format=%P", sha)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list parents for %q: %w", sha, err)
	}
	parents := strings.Fields(string(parentsOut))
	var out []byte
	if len(parents) == 0 {
		out, err = r.run(ctx, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", "-z", sha)
	} else {
		out, err = r.run(ctx, "diff", "--name-only", "-z", parents[0], sha)
	}
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list files for %q: %w", sha, err)
	}
	paths, err := nulUTF8Strings(out)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list files for %q: %w", sha, err)
	}
	return paths, nil
}

func (r *Repo) citations(ctx context.Context, sha, body string) ([]string, error) {
	out, err := r.run(ctx, "ls-tree", "-r", "--name-only", "-z", sha, "--", "docs/decisions")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list decisions for %q: %w", sha, err)
	}
	paths, err := nulUTF8Strings(out)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list decisions for %q: %w", sha, err)
	}
	var found []string
	for _, path := range paths {
		if pathpkg.Dir(path) != "docs/decisions" {
			continue
		}
		base := pathpkg.Base(path)
		if !strings.HasPrefix(base, "VD-") || !strings.HasSuffix(base, ".md") {
			continue
		}
		id := strings.TrimSuffix(base, ".md")
		if containsToken(body, id) {
			found = append(found, id)
		}
	}
	sort.Strings(found)
	return found, nil
}

func (r *Repo) refusePartial(ctx context.Context) error {
	out, err := r.run(ctx, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return fmt.Errorf("gitcmd: inspect shallow state: %w", err)
	}
	if strings.TrimSpace(string(out)) == "true" {
		return ErrShallow
	}
	graftsPath, err := r.run(ctx, "rev-parse", "--git-path", "info/grafts")
	if err != nil {
		return fmt.Errorf("gitcmd: locate grafts: %w", err)
	}
	path := strings.TrimSpace(string(graftsPath))
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.root, path)
	}
	info, statErr := os.Stat(path)
	if statErr == nil {
		if info.Size() > 0 {
			return ErrShallow
		}
		return nil
	}
	if errors.Is(statErr, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("gitcmd: inspect grafts: %w", statErr)
}

func (r *Repo) validateRefs(ctx context.Context) error {
	gitDirOut, err := r.run(ctx, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("gitcmd: locate git directory: %w", err)
	}
	gitDir := strings.TrimSpace(string(gitDirOut))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(r.root, gitDir)
	}
	refsDir := filepath.Join(gitDir, "refs")
	walkErr := filepath.WalkDir(refsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(gitDir, path)
		if relErr != nil {
			return relErr
		}
		ref := filepath.ToSlash(rel)
		if _, resolveErr := r.run(ctx, "rev-parse", "--verify", ref); resolveErr != nil {
			return fmt.Errorf("invalid ref %s: %w", ref, resolveErr)
		}
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, os.ErrNotExist) {
		return fmt.Errorf("gitcmd: validate loose refs: %w", walkErr)
	}
	refsOut, err := r.run(ctx, "for-each-ref", "--format=%(refname) %(objectname)")
	if err != nil {
		return fmt.Errorf("gitcmd: enumerate refs for validation: %w", err)
	}
	fields := strings.Fields(string(refsOut))
	if len(fields)%2 != 0 {
		return errors.New("gitcmd: validate refs: invalid ref listing")
	}
	for i := 0; i < len(fields); i += 2 {
		ref, object := fields[i], fields[i+1]
		if _, objectErr := r.run(ctx, "cat-file", "-e", object+"^{object}"); objectErr != nil {
			return fmt.Errorf("gitcmd: validate ref %s: %w", ref, objectErr)
		}
	}
	if _, headErr := r.run(ctx, "rev-parse", "--verify", "HEAD^{object}"); headErr != nil {
		if _, symbolicErr := r.run(ctx, "symbolic-ref", "-q", "HEAD"); symbolicErr != nil {
			return fmt.Errorf("gitcmd: validate detached HEAD: %w", headErr)
		}
	}
	if _, err := r.run(ctx, "show-ref"); err == nil {
		return nil
	} else if _, headErr := r.peelCommit(ctx, "HEAD"); headErr == nil {
		return fmt.Errorf("gitcmd: validate refs: %w", err)
	}
	// `show-ref` exits 1 with no refs in a legitimate unborn repository.
	return nil
}

func (r *Repo) peelCommit(ctx context.Context, ref string) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("gitcmd: empty peeled ref")
	}
	return sha, nil
}

func (r *Repo) run(ctx context.Context, args ...string) ([]byte, error) {
	base := []string{"--no-replace-objects", "-c", "core.quotepath=true", "-C", r.root}
	cmd := exec.CommandContext(ctx, "git", append(base, args...)...)
	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitErr.Stderr)))
	}
	return nil, err
}

func nulUTF8Strings(data []byte) ([]string, error) {
	parts := bytes.Split(data, []byte{0})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		if !utf8.Valid(part) {
			return nil, errors.New("path is not valid UTF-8")
		}
		result = append(result, string(part))
	}
	return result, nil
}

func containsToken(text, token string) bool {
	for start := 0; ; {
		index := strings.Index(text[start:], token)
		if index < 0 {
			return false
		}
		index += start
		beforeOK := index == 0 || !idByte(text[index-1])
		end := index + len(token)
		afterOK := end == len(text) || !idByte(text[end])
		if beforeOK && afterOK {
			return true
		}
		start = index + 1
	}
}

func idByte(b byte) bool {
	return b == '-' || b >= '0' && b <= '9' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
}

func excludedRef(ref string) bool {
	return ref == "refs/stash" || strings.HasPrefix(ref, "refs/notes/") || strings.HasPrefix(ref, "refs/replace/")
}
