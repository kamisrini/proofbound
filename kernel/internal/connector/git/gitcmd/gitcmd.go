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
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
)

const gitCommandTimeout = 10 * time.Second

var ErrShallow = errors.New("gitcmd: repository history is incomplete")

type Repo struct {
	root string
}

var relativePath = filepath.Rel

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
	shas, err := r.CommitSHAs(ctx)
	if err != nil {
		return nil, err
	}
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

// CommitSHAs lists exactly the commits Commits ingests, oldest first, after the same shallow,
// graft and ref refusals, without reading any commit payload or file list (G-INV-23). It is the
// one home of the history walk's arguments: Commits and the intent.sdd reader both use it.
func (r *Repo) CommitSHAs(ctx context.Context) ([]string, error) {
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
	return strings.Fields(string(out)), nil
}

// Parents returns commit's recorded parents in order; a root commit has none. Replace refs and
// grafts cannot alter them (G-INV-26).
func (r *Repo) Parents(ctx context.Context, commit string) ([]string, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	out, err := r.run(ctx, "rev-list", "--parents", "-n", "1", commit, "--")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: parents of %s: %w", commit, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 || fields[0] != commit {
		return nil, fmt.Errorf("gitcmd: parents of %s: unexpected output", commit)
	}
	return fields[1:], nil
}

// TreeEntry is one entry of a commit's tree. Path is the full repository path as raw bytes (it may
// be invalid UTF-8; callers decide). Mode is git's octal mode ("040000", "100644", "100755",
// "120000", "160000") and Type is "tree", "blob" or "commit".
type TreeEntry struct{ Mode, Type, OID, Path string }

// TreeEntries looks paths up in commit's tree without ever dereferencing "<commit>:<path>"
// (G-INV-24). With children=false it returns path's own entry (none when absent). With
// children=true it returns the entries under path (recursive=true walks the whole subtree, trees
// included); a path that is absent or not a tree has no children. Every lookup is literal.
func (r *Repo) TreeEntries(ctx context.Context, commit, path string, children, recursive bool) ([]TreeEntry, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	if path == "" || strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return nil, fmt.Errorf("gitcmd: tree path %q is not a repository path", path)
	}
	if children {
		own, err := r.TreeEntries(ctx, commit, path, false, false)
		if err != nil {
			return nil, err
		}
		if len(own) != 1 || own[0].Type != "tree" {
			return nil, nil
		}
	}
	args := []string{"--literal-pathspecs", "ls-tree", "-z", "--full-tree"}
	if recursive {
		args = append(args, "-r", "-t")
	}
	target := path
	if children {
		target += "/"
	}
	out, err := r.run(ctx, append(args, commit, "--", target)...)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: tree of %s at %q: %w", commit, path, err)
	}
	var entries []TreeEntry
	for _, record := range strings.Split(string(out), "\x00") {
		if record == "" {
			continue
		}
		meta, name, ok := strings.Cut(record, "\t")
		fields := strings.Fields(meta)
		if !ok || len(fields) != 3 {
			return nil, fmt.Errorf("gitcmd: tree of %s at %q: misframed entry", commit, path)
		}
		if !children && name != path || children && !strings.HasPrefix(name, path+"/") {
			continue // own-entry mode wants only path; children mode only what lies under it (-t also lists ancestors)
		}
		entries = append(entries, TreeEntry{Mode: fields[0], Type: fields[1], OID: fields[2], Path: name})
	}
	return entries, nil
}

// Blob returns the exact bytes of a blob object.
func (r *Repo) Blob(ctx context.Context, oid string) ([]byte, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	out, err := r.run(ctx, "cat-file", "blob", oid)
	if err != nil {
		return nil, fmt.Errorf("gitcmd: blob %s: %w", oid, err)
	}
	return out, nil
}

// Reachable returns the commit object names currently reachable from the
// repository's non-excluded refs. It deliberately does not read commit
// payloads or file trees; callers that only need supersession membership
// should not pay the full ingestion cost.
func (r *Repo) Reachable(ctx context.Context) (map[string]bool, error) {
	if r == nil {
		return nil, errors.New("gitcmd: repository is nil")
	}
	if err := r.refusePartial(ctx); err != nil {
		return nil, err
	}
	if err := r.validateRefs(ctx); err != nil {
		return nil, err
	}
	out, err := r.run(ctx, "rev-list", "--exclude=refs/stash", "--exclude=refs/notes/*", "--exclude=refs/replace/*", "--all")
	if err != nil {
		return nil, fmt.Errorf("gitcmd: list reachable commits: %w", err)
	}
	reachable := make(map[string]bool)
	for _, sha := range strings.Fields(string(out)) {
		reachable[sha] = true
	}
	if head, headErr := r.peelCommit(ctx, "HEAD"); headErr == nil {
		reachable[head] = true
	}
	return reachable, nil
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
		sha, isCommit, resolveErr := r.tipCommit(ctx, ref)
		if resolveErr != nil {
			return nil, fmt.Errorf("gitcmd: resolve tip %s: %w", ref, resolveErr)
		}
		if !isCommit {
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
	// Include parents in the metadata query. This removes one Git subprocess
	// per commit during verification on mounted workspaces.
	format := "%H%x00%an%x00%ae%x00%cn%x00%ce%x00%cI%x00%s%x00%B%x00%P%x00"
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
	files, err := r.files(ctx, sha, string(fields[8]))
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
		IntentTrailers: intentTrailers(string(fields[7])),
	}, nil
}

var trailerLineRE = regexp.MustCompile(`^([A-Za-z0-9-]+): (.+)$`)

func intentTrailers(body string) []string {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) < 3 {
		return nil
	}
	var values []string
	i := len(lines) - 1
	for ; i >= 0; i-- {
		match := trailerLineRE.FindStringSubmatch(lines[i])
		if match == nil {
			break
		}
		if match[1] == "Intent" {
			values = append(values, match[2])
		}
	}
	if i < 0 || lines[i] != "" {
		return nil
	}
	sort.Strings(values)
	if len(values) == 0 {
		return nil
	}
	unique := values[:0]
	for _, value := range values {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func parseScalars(out []byte) ([][]byte, error) {
	out = bytes.TrimSuffix(out, []byte{'\n'})
	fields := bytes.Split(out, []byte{0})
	if len(fields) != 10 {
		return nil, errors.New("output framing is invalid")
	}
	if len(fields[9]) != 0 {
		return nil, errors.New("output framing is invalid")
	}
	for _, field := range fields[:9] {
		if !utf8.Valid(field) {
			return nil, errors.New("scalar is not valid UTF-8")
		}
	}
	return fields, nil
}

func (r *Repo) files(ctx context.Context, sha, parentText string) ([]string, error) {
	parents := strings.Fields(parentText)
	var out []byte
	var err error
	if len(parents) == 0 {
		out, err = r.run(ctx, "diff-tree", "--root", "--ignore-submodules=none", "--no-ext-diff", "--no-textconv", "--no-commit-id", "--name-only", "-r", "-z", sha)
	} else {
		out, err = r.run(ctx, "diff", "--ignore-submodules=none", "--no-ext-diff", "--no-textconv", "--name-only", "-z", parents[0], sha)
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
		rel, relErr := relativePath(gitDir, path)
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
		if _, referentErr := r.run(ctx, "cat-file", "-e", ref+"^{}"); referentErr != nil {
			return fmt.Errorf("gitcmd: validate ref referent %s: %w", ref, referentErr)
		}
	}
	if _, symbolicErr := r.run(ctx, "symbolic-ref", "-q", "HEAD"); symbolicErr != nil {
		headOut, headErr := r.run(ctx, "rev-parse", "--verify", "HEAD^{object}")
		if headErr != nil {
			return fmt.Errorf("gitcmd: validate detached HEAD: %w", headErr)
		}
		typeOut, typeErr := r.run(ctx, "cat-file", "-t", strings.TrimSpace(string(headOut)))
		if typeErr != nil {
			return fmt.Errorf("gitcmd: validate detached HEAD type: %w", typeErr)
		}
		if strings.TrimSpace(string(typeOut)) != "commit" {
			return errors.New("gitcmd: detached HEAD does not name a commit object")
		}
	}
	// `show-ref` exits 1 both for a legitimate unborn repository and for a valid
	// detached HEAD, which has no named ref. Loose/packed refs and detached HEAD
	// were already validated above, so there is no additional failure to report.
	return nil
}

func (r *Repo) peelCommit(ctx context.Context, ref string) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(out))
	return sha, nil
}

func (r *Repo) tipCommit(ctx context.Context, ref string) (string, bool, error) {
	out, err := r.run(ctx, "rev-parse", "--verify", ref+"^{}")
	if err != nil {
		return "", false, err
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", false, errors.New("gitcmd: empty resolved ref")
	}
	typeOut, err := r.run(ctx, "cat-file", "-t", sha)
	if err != nil {
		return "", false, err
	}
	if strings.TrimSpace(string(typeOut)) != "commit" {
		return "", false, nil
	}
	return sha, true, nil
}

func (r *Repo) run(ctx context.Context, args ...string) ([]byte, error) {
	commandCtx, cancel := context.WithTimeout(ctx, gitCommandTimeout)
	defer cancel()
	base := []string{"--no-replace-objects", "-c", "core.quotepath=true", "-c", "diff.renames=false", "-c", "diff.ignoreSubmodules=none", "-C", r.root}
	cmd := exec.CommandContext(commandCtx, "git", append(base, args...)...)
	cmd.Env = gitEnvironment(os.Environ())
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

func gitEnvironment(environment []string) []string {
	clean := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if len(name) >= len("GIT_") && strings.EqualFold(name[:len("GIT_")], "GIT_") {
			continue
		}
		clean = append(clean, entry)
	}
	return clean
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
