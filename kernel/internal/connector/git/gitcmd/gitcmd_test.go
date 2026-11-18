package gitcmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"io"
	"log/slog"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestParseScalars_RejectsMisframing(t *testing.T) {
	for _, data := range [][]byte{
		[]byte("too few\x00fields\x00"),
		[]byte("a\x00b\x00c\x00d\x00e\x00f\x00g\x00h\x00not-empty"),
		[]byte("a\x00b\x00c\x00d\x00e\x00f\x00g\x00h\x00\x00\x00"),
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

func TestCommits_ExtractsOnlyExplicitIntentTrailers(t *testing.T) {
	body := "subject mentions CI-fake-item-acde12\n\nordinary prose\n\nIntent: CI-real-item-acde12\nIntent: specdir:CI-other-item-acde12\nIntent: CI-real-item-acde12\nSigned-off-by: Test <test@example.invalid>\n"
	got := intentTrailers(body)
	want := []string{"CI-real-item-acde12", "specdir:CI-other-item-acde12"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for _, text := range []string{"subject CI-real-item-acde12\n", "subject\n\nIntent: CI-real-item-acde12\nthen prose\n", "Intent: CI-real-item-acde12\n"} {
		if got := intentTrailers(text); got != nil {
			t.Fatalf("text=%q got=%v", text, got)
		}
	}
	if got := intentTrailers("subject\nIntent: CI-real-item-acde12\nSigned-off-by: Test <test@example.invalid>\n"); got != nil {
		t.Fatalf("missing separator accepted: %v", got)
	}
}

// TestIntentTrailers_SharedCorpus pins the kernel grammar to the corpus the shell gate
// (scripts/intent-delivery-applicability.sh) is also tested against, so the two parsers cannot
// drift apart again (DEF-INTENT-TRAILER-BLOCK-001: the gate grepped the whole message while this
// parser reads only the final trailer block, so 17 governed commits passed the gate and bound
// nothing). Column 2 of cases.tsv is what this function must extract.
func TestIntentTrailers_SharedCorpus(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "..", "scripts", "tests", "fixtures", "intent-trailers")
	table, err := os.ReadFile(filepath.Join(dir, "cases.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, line := range strings.Split(strings.TrimRight(string(table), "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) != 3 {
			t.Fatalf("malformed corpus row %q", line)
		}
		msg, err := os.ReadFile(filepath.Join(dir, cols[0]+".msg"))
		if err != nil {
			t.Fatal(err)
		}
		got := "-"
		if values := intentTrailers(string(msg)); values != nil {
			got = strings.Join(values, ",")
		}
		if got != cols[1] {
			t.Errorf("case %s: kernel extracted %q, corpus expects %q", cols[0], got, cols[1])
		}
		cases++
	}
	// Calibration: a corpus that silently shrank to nothing would make this test vacuous.
	if cases < 12 {
		t.Fatalf("corpus has %d cases, want >= 12", cases)
	}
}

// corpusResolver resolves every well-formed claim, so a Sync failure can only come from the
// kernel's own grammar, framing, or value rule — the properties the corpus pins.
type corpusResolver struct{}

func (corpusResolver) Resolve(_ context.Context, _ string, provider, id string) (connectorgit.IntentRef, error) {
	return connectorgit.IntentRef{Provider: provider, RecordID: id, ArtifactSHA256: strings.Repeat("a", 64)}, nil
}

type corpusAppender struct{}

func (corpusAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	return store.Record{Event: e}, true, nil
}

// TestIntentTrailers_SharedCorpusKernelBinds is the end-to-end kernel side of column 3 ("the
// kernel binds a claim"): each corpus message is committed verbatim into a real repository, read
// through this real adapter (UTF-8 framing, replace-object and environment handling), and synced
// by the real connector (value rule). The shell gate's self-test asserts the same column, so the
// two cannot disagree on any corpus input (gate review round 1, F1).
func TestIntentTrailers_SharedCorpusKernelBinds(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "..", "scripts", "tests", "fixtures", "intent-trailers"))
	if err != nil {
		t.Fatal(err)
	}
	table, err := os.ReadFile(filepath.Join(dir, "cases.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	cases, pass, fail := 0, 0, 0
	for _, line := range strings.Split(strings.TrimRight(string(table), "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) != 3 {
			t.Fatalf("malformed corpus row %q", line)
		}
		name, want := cols[0], cols[2]
		// Build the commit object from the exact message bytes: `git commit` (even with
		// --cleanup=verbatim) and `git commit-tree` re-encode non-UTF-8 input, so only hash-object
		// stores what the corpus file holds.
		msg, err := os.ReadFile(filepath.Join(dir, name+".msg"))
		if err != nil {
			t.Fatal(err)
		}
		fixture := newFixture(t)
		fixture.write("scripts/case.sh", name)
		fixture.git("add", "-A")
		tree := strings.TrimSpace(fixture.git("write-tree"))
		object := append([]byte("tree "+tree+"\nauthor Fixture User <fixture@example.test> 1700000000 +0000\ncommitter Fixture User <fixture@example.test> 1700000000 +0000\n\n"), msg...)
		sha := strings.TrimSpace(commandInput(t, fixture.root, nil, object, "git", "hash-object", "-t", "commit", "-w", "--stdin"))
		fixture.git("update-ref", "refs/heads/master", sha)
		binds := false
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if commits, err := repo.Commits(context.Background()); err == nil && len(commits) == 1 && len(commits[0].IntentTrailers) > 0 {
			ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{5}, 4096)), Now: func() time.Time { return time.Unix(2, 0).UTC() }})
			if err != nil {
				t.Fatal(err)
			}
			c, err := connectorgit.New(&connectorgit.Deps{Repo: repo, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Resolver: corpusResolver{}})
			if err != nil {
				t.Fatal(err)
			}
			_, syncErr := c.Sync(context.Background(), corpusAppender{})
			binds = syncErr == nil
		}
		got := "fail"
		if binds {
			got, pass = "pass", pass+1
		} else {
			fail++
		}
		if got != want {
			t.Errorf("case %s: kernel binds=%v, corpus column 3 expects %s", name, binds, want)
		}
		cases++
	}
	// Calibration: both outcomes must be exercised, or a constant answer would pass.
	if cases < 19 || pass < 4 || fail < 10 {
		t.Fatalf("corpus too weak: cases=%d pass=%d fail=%d", cases, pass, fail)
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

func TestCommits_DetachedHEADAtCommitIsNotAnError(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	commitSHA := fixture.head()
	fixture.git("checkout", "--detach", "-q", commitSHA)
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil || !contains(allCommitSHAs(t, fixture), commitSHA) || len(commits) != 1 {
		t.Fatalf("commits=%v error=%v", commits, err)
	}
}

func TestReachable_EmptyRepositoryIsNotAnError(t *testing.T) {
	fixture := newFixture(t)
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	reachable, err := repo.Reachable(context.Background())
	if err != nil || len(reachable) != 0 {
		t.Fatalf("reachable=%v error=%v", reachable, err)
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
	paths := []string{" leading", "plain-slash"}
	if runtime.GOOS != "windows" {
		paths = append(paths, "trailing ", "line\nbreak", `slash\\`)
	}
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

func TestNULUTF8Strings_RejectsNonUTF8PathIdentity(t *testing.T) {
	for _, data := range [][]byte{{0xfe, 0}, {0xff, 0}, {'a', 0xfe, 0}} {
		if paths, err := nulUTF8Strings(data); err == nil || paths != nil {
			t.Fatalf("data=%x paths=%q error=%v", data, paths, err)
		}
	}
	paths, err := nulUTF8Strings([]byte("snowman-\xe2\x98\x83\x00"))
	if err != nil || !reflect.DeepEqual(paths, []string{"snowman-☃"}) {
		t.Fatalf("paths=%q error=%v", paths, err)
	}
}

func TestCommits_PathIdentityIsPreservedOrRefused(t *testing.T) {
	t.Run("valid UTF-8 survives", func(t *testing.T) {
		fixture := newFixture(t)
		path := " leading-☃"
		if runtime.GOOS != "windows" {
			path += "\n"
		}
		fixture.write(path, "content")
		fixture.commit("valid", "")
		if files := onlyCommit(t, fixture).FilesTouched; !reflect.DeepEqual(files, []string{path}) {
			t.Fatalf("files=%q", files)
		}
	})
	for _, nonRoot := range []bool{false, true} {
		name := "root"
		if nonRoot {
			name = "non-root"
		}
		t.Run(name+" invalid UTF-8 is refused", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("native Windows filenames cannot carry invalid UTF-8 bytes")
			}
			fixture := newFixture(t)
			if nonRoot {
				fixture.write("base", "base")
				fixture.commit("base", "")
			}
			invalid := string([]byte{'p', 0xfe})
			fixture.write(invalid, "content")
			fixture.commit("invalid", "")
			repo, err := New(fixture.root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := repo.Commits(context.Background()); err == nil {
				t.Fatal("adapter accepted a path whose byte identity JSON cannot preserve")
			}
		})
	}
}

func TestCommits_ResolvesCitationsAgainstTheCommitTree(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("docs/decisions/VD-local-only-backup-mjic4a.md", "decision")
	fixture.write("docs/decisions/VD-short-abcdef.md", "decision")
	fixture.write("docs/decisions/VD-no-extension", "not a decision")
	fixture.write("docs/decisions/not-vd.md", "not a decision")
	fixture.write("docs/decisions/archive/VD-nested-abcdef.md", "not a direct decision")
	fixture.commit("cite", "VD-short-abcdef VD-local-only-backup VD-fiction-aaaaaa VD-local-only-backup-mjic4a VD-short-abcdef VD-no-extension VD-nested-abcdef not-vd")
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
		if _, err := repo.Reachable(context.Background()); !errors.Is(err, ErrShallow) {
			t.Fatalf("reachable error=%v", err)
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
	if _, err := repo.Reachable(context.Background()); err == nil {
		t.Fatal("broken ref was treated as reachable history")
	}
}

func TestReachableReportsRelativePathFailure(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	want := errors.New("relative path failed")
	old := relativePath
	relativePath = func(string, string) (string, error) { return "", want }
	t.Cleanup(func() { relativePath = old })
	if _, err := repo.Reachable(context.Background()); !errors.Is(err, want) {
		t.Fatalf("reachable error=%v", err)
	}
}

func TestRepo_MissingObjectRoutesAreErrors(t *testing.T) {
	for _, packed := range []bool{false, true} {
		name := "loose"
		if packed {
			name = "packed"
		}
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.write("a", "a")
			fixture.commit("one", "")
			repo, err := New(fixture.root)
			if err != nil {
				t.Fatal(err)
			}
			missing := strings.Repeat("a", 40)
			if packed {
				content := "# pack-refs with: peeled fully-peeled sorted\n" + missing + " refs/heads/missing\n"
				if err := os.WriteFile(filepath.Join(fixture.root, ".git", "packed-refs"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(filepath.Join(fixture.root, ".git", "refs", "heads", "missing"), []byte(missing+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := repo.Commits(context.Background()); err == nil {
				t.Fatal("Commits accepted a ref whose target object is absent")
			}
			if _, err := repo.Tips(context.Background()); err == nil {
				t.Fatal("Tips accepted a ref whose target object is absent")
			}
		})
	}
	t.Run("detached HEAD", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fixture.root, ".git", "HEAD"), []byte(strings.Repeat("a", 40)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); err == nil {
			t.Fatal("Commits accepted detached HEAD whose target object is absent")
		}
		if _, err := repo.Tips(context.Background()); err == nil {
			t.Fatal("Tips accepted detached HEAD whose target object is absent")
		}
	})
	t.Run("detached HEAD at non-commit", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		blob := strings.TrimSpace(commandInput(t, fixture.root, nil, []byte("blob"), "git", "hash-object", "-w", "--stdin"))
		if err := os.WriteFile(filepath.Join(fixture.root, ".git", "HEAD"), []byte(blob+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); err == nil {
			t.Fatal("Commits accepted detached HEAD at a non-commit object")
		}
		if _, err := repo.Tips(context.Background()); err == nil {
			t.Fatal("Tips accepted detached HEAD at a non-commit object")
		}
	})
	t.Run("detached HEAD at annotated tag", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		fixture.git("tag", "-a", "tagged", "-m", "tagged")
		tagObject := strings.TrimSpace(fixture.git("rev-parse", "refs/tags/tagged"))
		if err := os.WriteFile(filepath.Join(fixture.root, ".git", "HEAD"), []byte(tagObject+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); err == nil {
			t.Fatal("Commits accepted detached HEAD at an annotated-tag object")
		}
		if _, err := repo.Tips(context.Background()); err == nil {
			t.Fatal("Tips accepted detached HEAD at an annotated-tag object")
		}
	})
	t.Run("annotated tag referent", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		missing := strings.Repeat("1", 40)
		tag := "object " + missing + "\ntype commit\ntag broken\ntagger Fixture User <fixture@example.test> 1787520000 +0000\n\nbroken\n"
		tagSHA := strings.TrimSpace(commandInput(t, fixture.root, nil, []byte(tag), "git", "hash-object", "-t", "tag", "-w", "--stdin"))
		fixture.git("update-ref", "refs/tags/broken", tagSHA)
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); err == nil {
			t.Fatal("Commits accepted an annotated tag with a missing referent")
		}
		if _, err := repo.Tips(context.Background()); err == nil {
			t.Fatal("Tips accepted an annotated tag with a missing referent")
		}
	})
	t.Run("valid non-commit ref", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.write("a", "a")
		fixture.commit("one", "")
		blob := strings.TrimSpace(commandInput(t, fixture.root, nil, []byte("blob"), "git", "hash-object", "-w", "--stdin"))
		fixture.git("update-ref", "refs/data/blob", blob)
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Commits(context.Background()); err != nil {
			t.Fatalf("valid non-commit ref broke Commits: %v", err)
		}
		tips, err := repo.Tips(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := tips["refs/data/blob"]; exists {
			t.Fatalf("non-commit ref appeared in tips: %v", tips)
		}
	})
}

func TestCommits_RenamePayloadIgnoresLocalConfig(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("old-name", "same content")
	fixture.commit("old", "")
	fixture.git("mv", "old-name", "new-name")
	fixture.commit("rename", "")
	readTip := func(setting string) connectorgit.Commit {
		fixture.git("config", "diff.renames", setting)
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		commits, err := repo.Commits(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return commits[len(commits)-1]
	}
	disabled := readTip("false")
	enabled := readTip("true")
	want := []string{"new-name", "old-name"}
	if disabled.SHA != enabled.SHA || !reflect.DeepEqual(disabled.FilesTouched, want) || !reflect.DeepEqual(enabled.FilesTouched, want) {
		t.Fatalf("disabled=%+v enabled=%+v", disabled, enabled)
	}
}

func TestRepo_RootCannotBeRedirectedByGitEnvironment(t *testing.T) {
	wanted := newFixture(t)
	wanted.write("wanted", "wanted")
	wanted.commit("wanted", "")
	wantedSHA := wanted.head()
	other := newFixture(t)
	other.write("other", "other")
	other.commit("other", "")
	linked := filepath.Join(t.TempDir(), "linked")
	wanted.git("worktree", "add", "-q", "-b", "linked", linked)
	t.Setenv("GIT_DIR", filepath.Join(other.root, ".git"))
	t.Setenv("GIT_WORK_TREE", wanted.root)
	repo, err := New(wanted.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil || len(commits) != 1 || commits[0].SHA != wantedSHA || !reflect.DeepEqual(commits[0].FilesTouched, []string{"wanted"}) {
		t.Fatalf("commits=%+v error=%v", commits, err)
	}

	linkedRepo, err := New(linked)
	if err != nil {
		t.Fatal(err)
	}
	linkedCommits, err := linkedRepo.Commits(context.Background())
	if err != nil || len(linkedCommits) == 0 || linkedCommits[len(linkedCommits)-1].SHA != wantedSHA {
		t.Fatalf("linked commits=%+v error=%v", linkedCommits, err)
	}
}

func TestGitEnvironment_RemovesEveryGitSelector(t *testing.T) {
	input := []string{"PATH=/bin", "GIT_DIR=/other", "git_work_tree=/wrong", "GIT_COMMON_DIR=/common", "HOME=/home"}
	want := []string{"PATH=/bin", "HOME=/home"}
	if got := gitEnvironment(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("environment=%v want=%v", got, want)
	}
}

func TestCommits_GitlinkPayloadIgnoresSubmoduleConfig(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("one", "one")
	fixture.commit("one", "")
	firstTarget := fixture.head()
	fixture.write("two", "two")
	fixture.commit("two", "")
	secondTarget := fixture.head()
	fixture.git("checkout", "-q", "--orphan", "links")
	fixture.git("rm", "-q", "-rf", ".")
	fixture.git("update-index", "--add", "--cacheinfo", "160000", firstTarget, "sub")
	fixture.git("commit", "-q", "-m", "gitlink root")
	rootSHA := fixture.head()
	fixture.git("update-index", "--cacheinfo", "160000", secondTarget, "sub")
	fixture.git("commit", "-q", "-m", "gitlink changed")
	tipSHA := fixture.head()

	read := func() map[string][]string {
		repo, err := New(fixture.root)
		if err != nil {
			t.Fatal(err)
		}
		commits, err := repo.Commits(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		result := make(map[string][]string)
		for _, commit := range commits {
			result[commit.SHA] = commit.FilesTouched
		}
		return result
	}
	fixture.git("config", "diff.ignoreSubmodules", "all")
	localAll := read()
	fixture.git("config", "diff.ignoreSubmodules", "none")
	localNone := read()
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[diff]\n\tignoreSubmodules = all\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fixture.git("config", "--unset", "diff.ignoreSubmodules")
	t.Setenv("HOME", home)
	globalAll := read()
	for _, got := range []map[string][]string{localAll, localNone, globalAll} {
		if !reflect.DeepEqual(got[rootSHA], []string{"sub"}) || !reflect.DeepEqual(got[tipSHA], []string{"sub"}) {
			t.Fatalf("root=%q tip=%q", got[rootSHA], got[tipSHA])
		}
	}
}

func TestCommits_RefusesInvalidUTF8Scalars(t *testing.T) {
	for index := 0; index < 9; index++ {
		fields := [][]byte{[]byte("sha"), []byte("author"), []byte("author@example.test"), []byte("committer"), []byte("committer@example.test"), []byte("2026-08-24T00:00:00Z"), []byte("subject"), []byte("body"), nil, nil}
		fields[index] = append(fields[index], 0xff)
		if _, err := parseScalars(bytes.Join(fields, []byte{0})); err == nil {
			t.Fatalf("scalar field %d accepted invalid UTF-8", index)
		}
	}

	fixture := newFixture(t)
	fixture.write("file", "content")
	fixture.git("add", "file")
	tree := strings.TrimSpace(fixture.git("write-tree"))
	raw := []byte("tree " + tree + "\nauthor Fixture User <fixture@example.test> 1787520000 +0000\ncommitter Fixture User <fixture@example.test> 1787520000 +0000\n\nbad-")
	raw = append(raw, 0xff, '\n')
	sha := strings.TrimSpace(commandInput(t, fixture.root, nil, raw, "git", "hash-object", "-t", "commit", "-w", "--stdin"))
	fixture.git("update-ref", "refs/heads/master", sha)
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Commits(context.Background()); err == nil {
		t.Fatal("adapter accepted invalid UTF-8 commit scalar content")
	}
}

func TestCommits_MergeFilesAreFirstParentDelta(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("base", "base")
	fixture.commit("base", "")
	fixture.git("checkout", "-q", "-b", "feature")
	fixture.write("feature", "feature")
	fixture.commit("feature", "")
	fixture.git("checkout", "-q", "master")
	fixture.write("main", "main")
	fixture.commit("main", "")
	fixture.git("merge", "--no-ff", "-q", "feature", "-m", "merge")
	mergeSHA := fixture.head()
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, commit := range commits {
		if commit.SHA == mergeSHA {
			if !reflect.DeepEqual(commit.FilesTouched, []string{"feature"}) {
				t.Fatalf("merge files=%q", commit.FilesTouched)
			}
			return
		}
	}
	t.Fatalf("merge %s absent", mergeSHA)
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
	return commandInput(t, dir, env, nil, name, args...)
}

func commandInput(t *testing.T, dir string, env []string, input []byte, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = bytes.NewReader(input)
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

// TestCommitSHAs_EqualsCommitsAndRefuses proves G-INV-23: CommitSHAs is exactly Commits' SHA list
// and repeats the shallow/graft refusal at every call.
func TestCommitSHAs_EqualsCommitsAndRefuses(t *testing.T) {
	fixture := newFixture(t)
	fixture.write("a", "a")
	fixture.commit("one", "")
	fixture.git("checkout", "-q", "-b", "side")
	fixture.write("b", "b")
	fixture.commit("two", "")
	fixture.git("checkout", "-q", "master")
	fixture.write("c", "c")
	fixture.commit("three", "")
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Commits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	shas, err := repo.CommitSHAs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, c := range commits {
		want = append(want, c.SHA)
	}
	if !reflect.DeepEqual(shas, want) || len(shas) != 3 {
		t.Fatalf("CommitSHAs=%v want %v", shas, want)
	}
	grafts := filepath.Join(fixture.root, ".git", "info", "grafts")
	if err := os.WriteFile(grafts, []byte(fixture.head()+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CommitSHAs(context.Background()); !errors.Is(err, ErrShallow) {
		t.Fatalf("grafted listing error=%v", err)
	}
}

// TestTreeEntries_TypedLiteralAndNeverDereferenced proves G-INV-24 and G-INV-25: a path's own
// entry is returned typed whatever its type (tree, blob, symlink, gitlink), an absent or
// magic-looking path is simply absent, children of a non-tree are empty, a literal "*x" name is
// read as itself, and a non-UTF-8 path is returned as raw bytes rather than refused.
func TestTreeEntries_TypedLiteralAndNeverDereferenced(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks and non-UTF-8 names need a POSIX filesystem")
	}
	fixture := newFixture(t)
	fixture.write("root/dir/f.md", "f")
	fixture.write("root/*x/g.md", "g")
	fixture.write("root/plain.md", "p")
	if err := os.Symlink("dir", filepath.Join(fixture.root, "root", "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "root", "caf\xe9.md"), []byte("c"), 0o644); err != nil {
		t.Fatal(err)
	}
	fixture.git("add", "-A")
	fixture.git("update-index", "--add", "--cacheinfo", "160000,"+strings.Repeat("a", 40)+",root/sub")
	fixture.git("commit", "-q", "-m", "tree") // not fixture.commit: `add -A` would drop the gitlink
	repo, err := New(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, head := context.Background(), fixture.head()
	own := func(path string) TreeEntry {
		t.Helper()
		e, err := repo.TreeEntries(ctx, head, path, false, false)
		if err != nil || len(e) != 1 {
			t.Fatalf("%s: entries=%v err=%v", path, e, err)
		}
		return e[0]
	}
	if e := own("root/dir"); e.Type != "tree" || e.Mode != "040000" {
		t.Fatalf("dir=%+v", e)
	}
	if e := own("root/plain.md"); e.Type != "blob" || e.Mode != "100644" {
		t.Fatalf("plain=%+v", e)
	}
	if e := own("root/link"); e.Mode != "120000" {
		t.Fatalf("symlink=%+v", e)
	}
	if e := own("root/sub"); e.Mode != "160000" || e.Type != "commit" {
		t.Fatalf("gitlink=%+v", e)
	}
	if e := own("root/*x"); e.Type != "tree" || e.Path != "root/*x" {
		t.Fatalf("literal star=%+v", e)
	}
	for _, absent := range []string{"root/nope", ":(glob)root/dir", ":!root"} {
		if e, err := repo.TreeEntries(ctx, head, absent, false, false); err != nil || len(e) != 0 {
			t.Fatalf("%q: entries=%v err=%v", absent, e, err)
		}
	}
	for _, nontree := range []string{"root/plain.md", "root/link", "root/sub", "root/nope"} {
		if e, err := repo.TreeEntries(ctx, head, nontree, true, true); err != nil || len(e) != 0 {
			t.Fatalf("children of %q: entries=%v err=%v", nontree, e, err)
		}
	}
	all, err := repo.TreeEntries(ctx, head, "root", true, true)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, e := range all {
		paths = append(paths, e.Path)
	}
	sort.Strings(paths)
	want := []string{"root/*x", "root/*x/g.md", "root/caf\xe9.md", "root/dir", "root/dir/f.md", "root/link", "root/plain.md", "root/sub"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("recursive children=%q want %q", paths, want)
	}
	nested, err := repo.TreeEntries(ctx, head, "root/dir", true, true)
	if err != nil || len(nested) != 1 || nested[0].Path != "root/dir/f.md" {
		t.Fatalf("children of a nested path must not include its ancestors: %+v %v", nested, err)
	}
	if b, err := repo.Blob(ctx, own("root/plain.md").OID); err != nil || string(b) != "p" {
		t.Fatalf("blob=%q err=%v", b, err)
	}
	if _, err := repo.TreeEntries(ctx, head, "", false, false); err == nil {
		t.Fatal("empty path must be refused, not read as the whole tree")
	}
}

// TestParents_IgnoresReplacementAndRoots proves G-INV-26: a replace --graft cannot turn a commit
// into a fabricated root, and a real root has no parents.
func TestParents_IgnoresReplacementAndRoots(t *testing.T) {
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
	if p, err := repo.Parents(context.Background(), second); err != nil || !reflect.DeepEqual(p, []string{first}) {
		t.Fatalf("parents=%v err=%v", p, err)
	}
	if p, err := repo.Parents(context.Background(), first); err != nil || len(p) != 0 {
		t.Fatalf("root parents=%v err=%v", p, err)
	}
	shas, err := repo.CommitSHAs(context.Background())
	if err != nil || !reflect.DeepEqual(shas, []string{first, second}) {
		t.Fatalf("shas=%v err=%v", shas, err)
	}
}
