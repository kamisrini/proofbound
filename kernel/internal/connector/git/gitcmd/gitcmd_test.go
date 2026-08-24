package gitcmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
)

func TestParseScalars_RejectsMisframing(t *testing.T) {
	for _, data := range [][]byte{
		[]byte("too few\x00fields\x00"),
		[]byte("a\x00b\x00c\x00d\x00e\x00f\x00g\x00h\x00not-empty"),
		[]byte("a\x00b\x00c\x00d\x00e\x00f\x00g\x00h\x00\x00"),
	} {
		if _, err := parseScalars(data); err == nil {
			t.Fatalf("accepted %q", data)
		}
	}
}

func TestTokenBoundariesAndRefClasses(t *testing.T) {
	if !containsToken("VD-real-abcdef", "VD-real-abcdef") ||
		!containsToken("(VD-real-abcdef)", "VD-real-abcdef") ||
		containsToken("xVD-real-abcdef", "VD-real-abcdef") ||
		containsToken("VD-real-abcdefx", "VD-real-abcdef") ||
		containsToken("VD-real-abcdef-more", "VD-real-abcdef") {
		t.Fatal("token boundary mismatch")
	}
	for _, char := range []byte{'-', '0', '9', 'A', 'Z', 'a', 'z'} {
		if !idByte(char) {
			t.Fatalf("id byte %q rejected", char)
		}
	}
	for _, char := range []byte{'_', '.', '/', ' '} {
		if idByte(char) {
			t.Fatalf("non-id byte %q accepted", char)
		}
	}
	for _, ref := range []string{"refs/stash", "refs/notes/x", "refs/replace/x"} {
		if !excludedRef(ref) {
			t.Fatalf("ref %q included", ref)
		}
	}
	for _, ref := range []string{"refs/stashed", "refs/note/x", "refs/replaced/x"} {
		if excludedRef(ref) {
			t.Fatalf("ref %q excluded", ref)
		}
	}
}

func TestCommits_EmptyRepositoryIsNotAnError(t *testing.T) {
	fixture := newFixture(t)
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil || len(commits) != 0 {
		t.Fatalf("commits=%v error=%v", commits, err)
	}
}

func TestCommits_MapsEveryScalar(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("file.txt", "content")
	fixture.gitEnv([]string{
		"GIT_AUTHOR_NAME=A Author", "GIT_AUTHOR_EMAIL=author@example.test",
		"GIT_COMMITTER_NAME=C Committer", "GIT_COMMITTER_EMAIL=committer@example.test",
		"GIT_AUTHOR_DATE=2026-08-20T10:00:00Z", "GIT_COMMITTER_DATE=2026-08-21T11:12:13Z",
	}, "add", "file.txt")
	fixture.gitEnv([]string{
		"GIT_AUTHOR_NAME=A Author", "GIT_AUTHOR_EMAIL=author@example.test",
		"GIT_COMMITTER_NAME=C Committer", "GIT_COMMITTER_EMAIL=committer@example.test",
		"GIT_AUTHOR_DATE=2026-08-20T10:00:00Z", "GIT_COMMITTER_DATE=2026-08-21T11:12:13Z",
	}, "commit", "-m", "the subject", "-m", "the body")
	commit := onlyCommit(t, fixture)
	if commit.AuthorName != "A Author" || commit.AuthorEmail != "author@example.test" ||
		commit.CommitterName != "C Committer" || commit.CommitterEmail != "committer@example.test" ||
		commit.CommittedAt.Format("2006-01-02T15:04:05Z07:00") != "2026-08-21T11:12:13Z" || commit.Subject != "the subject" {
		t.Fatalf("commit=%+v", commit)
	}
}

func TestCommits_ContentCannotBreakFraming(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("file.txt", "content")
	fixture.commit("subject\x1ewith delimiter", "body\x1fwith delimiter")
	commit := onlyCommit(t, fixture)
	if commit.Subject != "subject\x1ewith delimiter" {
		t.Fatalf("subject=%q", commit.Subject)
	}
}

func TestCommits_PreservesHostilePaths(t *testing.T) {
	fixture := newFixture(t)
	paths := []string{" leading", "trailing ", "line\nbreak", `quote"slash\\`}
	for _, path := range paths {
		fixture.write(path, path)
	}
	fixture.commit("paths", "")
	commit := onlyCommit(t, fixture)
	sort.Strings(paths)
	sort.Strings(commit.FilesTouched)
	if !reflect.DeepEqual(commit.FilesTouched, paths) {
		t.Fatalf("files=%q want=%q", commit.FilesTouched, paths)
	}
}

func TestCommits_ResolvesCitationsAgainstTheCommitTree(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("docs/decisions/VD-local-only-backup-mjic4a.md", "decision")
	fixture.write("docs/decisions/VD-short-abcdef.md", "decision")
	fixture.write("docs/decisions/VD-no-extension", "not a decision")
	fixture.write("docs/decisions/not-vd.md", "not a decision")
	fixture.commit("cite", "VD-short-abcdef VD-local-only-backup VD-fiction-aaaaaa VD-local-only-backup-mjic4a VD-short-abcdef VD-no-extension not-vd")
	commit := onlyCommit(t, fixture)
	want := []string{"VD-local-only-backup-mjic4a", "VD-short-abcdef"}
	if !reflect.DeepEqual(commit.CitedDecisions, want) {
		t.Fatalf("citations=%v", commit.CitedDecisions)
	}
}

func TestRepo_RefusesPartialHistory(t *testing.T) {
	t.Run("grafts at construction and listing", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		grafts := filepath.Join(fixture.root, ".git", "info", "grafts")
		if err := os.MkdirAll(filepath.Dir(grafts), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(grafts, []byte(fixture.head()+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); !errors.Is(err, ErrShallow) {
			t.Fatalf("listing error=%v", err)
		}
		if _, err := New(fixture.root); !errors.Is(err, ErrShallow) {
			t.Fatalf("construction error=%v", err)
		}
	})

	t.Run("shallow", func(t *testing.T) {
		source := newFixture(t)
		source.write("a", "a")
		source.commit("one", "")
		source.write("b", "b")
		source.commit("two", "")
		clone := filepath.Join(t.TempDir(), "clone")
		command(t, "", nil, "git", "clone", "--quiet", "--depth=1", "file://"+filepath.ToSlash(source.root), clone)
		if _, err := New(clone); !errors.Is(err, ErrShallow) {
			t.Fatalf("error=%v", err)
		}
	})
}

func TestCommits_BrokenRepositoryIsAnError(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, ".git", "refs", "heads", "broken"), []byte("not-a-sha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Commits(context.Background()); err == nil {
		t.Fatal("broken ref was treated as an empty repository")
	}
}

func TestCommits_IncludesEveryHistoryRoute(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("main", "main")
	fixture.commit("main", "")
	mainSHA := fixture.head()
	fixture.git("checkout", "-q", "-b", "other")
	fixture.write("branch", "branch")
	fixture.commit("branch", "")
	branchSHA := fixture.head()
	fixture.git("tag", "only-tag", branchSHA)
	fixture.git("checkout", "-q", "master")
	fixture.git("branch", "-D", "other")
	fixture.git("checkout", "-q", "--detach", mainSHA)
	fixture.write("detached", "detached")
	fixture.commit("detached", "")
	detachedSHA := fixture.head()
	commits := allCommitSHAs(t, fixture)
	for _, sha := range []string{mainSHA, branchSHA, detachedSHA} {
		if !contains(commits, sha) {
			t.Fatalf("missing %s from %v", sha, commits)
		}
	}
}

func TestCommits_ExcludesNonHistoryRefs(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	fixture.write("a", "changed")
	fixture.git("stash", "push", "-q")
	fixture.git("notes", "add", "-m", "note")
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 {
		t.Fatalf("commits=%v", commits)
	}
	tips, err := repo.Tips(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for ref := range tips {
		if excludedRef(ref) {
			t.Fatalf("excluded tip %q in %v", ref, tips)
		}
	}
}

func TestTips_CoverCommitsAndPeelTags(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	head := fixture.head()
	fixture.git("tag", "-a", "release", "-m", "release")
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	tips, err := repo.Tips(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tips["HEAD"] != head || tips["refs/heads/master"] != head || tips["refs/tags/release"] != head {
		t.Fatalf("tips=%v head=%s", tips, head)
	}
}

func TestCommits_IgnoresReplacementObjects(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	first := fixture.head()
	fixture.write("b", "b")
	fixture.commit("two", "")
	second := fixture.head()
	fixture.git("replace", "--graft", second)
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].SHA != first || commits[1].SHA != second || !reflect.DeepEqual(commits[1].FilesTouched, []string{"b"}) {
		t.Fatalf("commits=%+v", commits)
	}
}

type fixtureRepo struct {
	t    *testing.T
	root string
}

func newFixture(t *testing.T) *fixtureRepo {
	t.Helper()
	root := t.TempDir()
	command(t, root, nil, "git", "init", "-q")
	command(t, root, nil, "git", "symbolic-ref", "HEAD", "refs/heads/master")
	command(t, root, nil, "git", "config", "user.name", "Fixture User")
	command(t, root, nil, "git", "config", "user.email", "fixture@example.test")
	return &fixtureRepo{t: t, root: root}
}

func (f *fixtureRepo) write(path, content string) {
	f.t.Helper()
	full := filepath.Join(f.root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixtureRepo) commit(subject, body string) {
	f.t.Helper()
	f.git("add", "-A")
	args := []string{"commit", "-q", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}
	f.git(args...)
}

func (f *fixtureRepo) head() string {
	f.t.Helper()
	return strings.TrimSpace(f.git("rev-parse", "HEAD"))
}

func (f *fixtureRepo) git(args ...string) string {
	f.t.Helper()
	return f.gitEnv(nil, args...)
}

func (f *fixtureRepo) gitEnv(env []string, args ...string) string {
	f.t.Helper()
	return command(f.t, f.root, env, "git", args...)
}

func command(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return string(out)
}

func onlyCommit(t *testing.T, fixture *fixtureRepo) connectorgit.Commit {
	t.Helper()
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil || len(commits) != 1 {
		t.Fatalf("commits=%v error=%v", commits, err)
	}
	return commits[0]
}

func allCommitSHAs(t *testing.T, fixture *fixtureRepo) []string {
	t.Helper()
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, len(commits))
	for i, commit := range commits {
		result[i] = commit.SHA
	}
	return result
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
