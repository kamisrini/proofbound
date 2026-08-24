package git_test

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func TestSync_AmendAppendsTheNewShaAndKeepsTheOld(t *testing.T) {
	fixture := newRewriteFixture(t)
	fixture.write("a", "one")
	fixture.commit("one", nil)
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 1)
	old := fixture.head()
	fixture.write("a", "amended")
	fixture.git("add", "a")
	fixture.git("commit", "-q", "--amend", "-m", "amended")
	newSHA := fixture.head()
	if old == newSHA {
		t.Fatal("amend did not change sha")
	}
	assertSync(t, connector, appender, 1)
	assertSync(t, connector, appender, 0)
	if !appender.has(old) || !appender.has(newSHA) || len(appender.events) != 2 {
		t.Fatalf("events=%v", appender.nativeIDs())
	}
}

func TestSync_RebaseAppendsEveryNewSha(t *testing.T) {
	fixture := newRewriteFixture(t)
	fixture.write("base", "base")
	fixture.commit("base", nil)
	fixture.git("checkout", "-q", "-b", "feature")
	fixture.write("feature-one", "one")
	fixture.commit("feature one", nil)
	fixture.write("feature-two", "two")
	fixture.commit("feature two", nil)
	fixture.git("checkout", "-q", "master")
	fixture.write("main", "main")
	fixture.commit("main", nil)
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 4)
	fixture.git("checkout", "-q", "feature")
	fixture.git("rebase", "master")
	assertSync(t, connector, appender, 2)
	assertSync(t, connector, appender, 0)
	if len(appender.events) != 6 {
		t.Fatalf("events=%v", appender.nativeIDs())
	}
}

func TestSync_SecondBranchIsIngestedRegardlessOfCheckout(t *testing.T) {
	fixture := newRewriteFixture(t)
	fixture.write("main", "main")
	fixture.commit("main", nil)
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 1)
	fixture.git("checkout", "-q", "-b", "other")
	fixture.write("other", "other")
	fixture.commit("other", nil)
	other := fixture.head()
	fixture.git("checkout", "-q", "master")
	assertSync(t, connector, appender, 1)
	assertSync(t, connector, appender, 0)
	if !appender.has(other) {
		t.Fatalf("events=%v", appender.nativeIDs())
	}
}

func TestSync_ACommitOlderThanAnyAlreadySeenIsIngested(t *testing.T) {
	fixture := newRewriteFixture(t)
	fixture.write("newer", "newer")
	fixture.commit("newer", []string{"GIT_AUTHOR_DATE=2026-08-24T12:00:00Z", "GIT_COMMITTER_DATE=2026-08-24T12:00:00Z"})
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 1)
	fixture.write("older", "older")
	fixture.commit("older", []string{"GIT_AUTHOR_DATE=2020-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2020-01-01T00:00:00Z"})
	older := fixture.head()
	assertSync(t, connector, appender, 1)
	assertSync(t, connector, appender, 0)
	if !appender.has(older) {
		t.Fatalf("events=%v", appender.nativeIDs())
	}
}

func TestSync_CitedDecisionsAreReadFromTheCommitBody(t *testing.T) {
	fixture := newRewriteFixture(t)
	fixture.write("docs/decisions/VD-fixture-aaaaaa.md", "decision")
	fixture.git("add", "-A")
	fixture.git("commit", "-q", "-m", "subject without citation", "-m", "VD-fixture-aaaaaa")
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 1)
	var payload connectorgit.Commit
	if err := json.Unmarshal(appender.events[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(payload.CitedDecisions, []string{"VD-fixture-aaaaaa"}) {
		t.Fatalf("citations=%v", payload.CitedDecisions)
	}
}

func TestSync_EmptyRepositoryIsNotAnError(t *testing.T) {
	fixture := newRewriteFixture(t)
	connector, appender := rewriteConnector(t, fixture.root)
	assertSync(t, connector, appender, 0)
	if len(appender.events) != 0 {
		t.Fatalf("events=%v", appender.nativeIDs())
	}
}

type rewriteAppender struct {
	events []core.Event
	seen   map[string]struct{}
}

func (a *rewriteAppender) Append(_ context.Context, event core.Event) (store.Record, bool, error) {
	key := event.IdempotencyKey().String()
	if _, ok := a.seen[key]; ok {
		return store.Record{Event: event}, false, nil
	}
	a.seen[key] = struct{}{}
	a.events = append(a.events, event)
	return store.Record{Seq: int64(len(a.events)), Event: event}, true, nil
}

func (a *rewriteAppender) has(sha string) bool {
	for _, event := range a.events {
		if event.NativeID == sha {
			return true
		}
	}
	return false
}

func (a *rewriteAppender) nativeIDs() []string {
	result := make([]string, len(a.events))
	for i, event := range a.events {
		result[i] = event.NativeID
	}
	return result
}

func rewriteConnector(t *testing.T, root string) (*connectorgit.Connector, *rewriteAppender) {
	t.Helper()
	repo, err := gitcmd.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	connector, err := connectorgit.New(&connectorgit.Deps{
		Repo: repo, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return connector, &rewriteAppender{seen: make(map[string]struct{})}
}

func assertSync(t *testing.T, connector *connectorgit.Connector, appender *rewriteAppender, appended int) {
	t.Helper()
	result, err := connector.Sync(context.Background(), appender)
	if err != nil {
		t.Fatal(err)
	}
	if result.Appended != appended || result.Appended+result.Existing != result.Listed {
		t.Fatalf("result=%+v want appended=%d", result, appended)
	}
}

type rewriteFixture struct {
	t    *testing.T
	root string
}

func newRewriteFixture(t *testing.T) *rewriteFixture {
	t.Helper()
	root := t.TempDir()
	rewriteCommand(t, root, nil, "init", "-q")
	rewriteCommand(t, root, nil, "symbolic-ref", "HEAD", "refs/heads/master")
	rewriteCommand(t, root, nil, "config", "user.name", "Fixture User")
	rewriteCommand(t, root, nil, "config", "user.email", "fixture@example.test")
	return &rewriteFixture{t: t, root: root}
}

func (f *rewriteFixture) write(path, content string) {
	f.t.Helper()
	full := filepath.Join(f.root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *rewriteFixture) commit(subject string, env []string) {
	f.t.Helper()
	f.git("add", "-A")
	rewriteCommand(f.t, f.root, env, "commit", "-q", "-m", subject)
}

func (f *rewriteFixture) git(args ...string) string {
	f.t.Helper()
	return rewriteCommand(f.t, f.root, nil, args...)
}

func (f *rewriteFixture) head() string {
	f.t.Helper()
	return strings.TrimSpace(f.git("rev-parse", "HEAD"))
}

func rewriteCommand(t *testing.T, root string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
