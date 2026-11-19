package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	intentsdd "github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// These tests derive from the R2 M3 SPEC (docs/plans/R2-sdd-m3-git-binding-spec-draft5.md §7, as
// amended by drafts 6-8). They open no store: every embedded-store M3 proof lives in the sdd
// package beside its other cli.Run tests, because the embedded database listens on one fixed port
// and go test runs packages in parallel.

const m3Profile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": "owner-a" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "sponsor-a" } ]
}
`

type m3Tenant struct {
	t    *testing.T
	root string
}

func newM3Tenant(t *testing.T) *m3Tenant {
	t.Helper()
	m := &m3Tenant{t: t, root: t.TempDir()}
	m.git("init", "-q", "-b", "main")
	m.write(intentsdd.ProfilePath, m3Profile)
	m.write("openspec/changes/add-auth/specs/core/spec.md", "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users.\n")
	return m
}

func (m *m3Tenant) git(args ...string) string {
	m.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = m.root
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		m.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (m *m3Tenant) write(rel, body string) {
	m.t.Helper()
	p := filepath.Join(m.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		m.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		m.t.Fatal(err)
	}
}

// commit stages the tenant's own paths (never the state root's ledger files, which share
// .proofbound/ with the profile) and commits with the given trailers.
func (m *m3Tenant) commit(subject string, trailers ...string) string {
	m.t.Helper()
	m.git("add", "-A", "--", intentsdd.ProfilePath, "openspec")
	for _, extra := range []string{"src", "docs", "gates"} {
		if _, err := os.Stat(filepath.Join(m.root, extra)); err == nil {
			m.git("add", "-A", "--", extra)
		}
	}
	msg := subject
	if len(trailers) > 0 {
		msg += "\n\n" + strings.Join(trailers, "\n")
	}
	m.git("commit", "-q", "--allow-empty", "-m", msg)
	return m.git("rev-parse", "HEAD")
}

type fakeSDDClaims struct {
	rev connectorintent.Revision
	err error
}

func (f fakeSDDClaims) Resolve(context.Context, string, string) (connectorintent.Revision, error) {
	return f.rev, f.err
}

// M3-INV-2 (cli half) and the §4.4 cause table: each §4.2 clause rejects on its own, each sdd
// sentinel maps to its cause, and a correct revision binds only once it is in L.
func TestSDDResolverChecksAndCauses(t *testing.T) {
	digest := strings.Repeat("d", 64)
	good := connectorintent.Revision{Source: core.SourceIntentSDD, Kind: core.KindChangeIntent, NativeID: "CI-add-auth-xuuno2", ArtifactSHA256: digest, ArtifactPath: "openspec/changes/add-auth"}
	ledgered := map[string]bool{"CI-add-auth-xuuno2\x00" + digest: true}
	resolve := func(rev connectorintent.Revision, err error, set map[string]bool) (connectorgit.IntentRef, error) {
		r := &commitIntentResolver{sdd: fakeSDDClaims{rev: rev, err: err}, ledgered: set}
		return r.Resolve(context.Background(), "abc", "sdd", "add-auth")
	}
	if ref, err := resolve(good, nil, ledgered); err != nil || ref != (connectorgit.IntentRef{Provider: "sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: digest}) {
		t.Fatalf("control: ref=%+v err=%v", ref, err)
	}
	archived := good
	archived.ArtifactPath = "openspec/changes/archive/2026-01-01-add-auth"
	if _, err := resolve(archived, nil, ledgered); err != nil {
		t.Fatalf("a dated archive segment was refused: %v", err)
	}
	spoofs := map[string]func(*connectorintent.Revision){
		"source":       func(r *connectorintent.Revision) { r.Source = core.SourceIntentRecords },
		"kind":         func(r *connectorintent.Revision) { r.Kind = core.KindRequirement },
		"id":           func(r *connectorintent.Revision) { r.NativeID = "add-auth" },
		"digest":       func(r *connectorintent.Revision) { r.ArtifactSHA256 = "bad" },
		"segment":      func(r *connectorintent.Revision) { r.ArtifactPath = "openspec/changes/add-auth-two" },
		"undated":      func(r *connectorintent.Revision) { r.ArtifactPath = "openspec/changes/archive/x-add-auth" },
		"undated-long": func(r *connectorintent.Revision) { r.ArtifactPath = "openspec/changes/archive/abcd-ef-gh-add-auth" },
		"dated-other":  func(r *connectorintent.Revision) { r.ArtifactPath = "openspec/changes/archive/2026-01-01-other" },
		"empty-path":   func(r *connectorintent.Revision) { r.ArtifactPath = "" },
	}
	for name, mutate := range spoofs {
		rev := good
		mutate(&rev)
		_, err := resolve(rev, nil, ledgered)
		if err == nil || errors.Is(err, connectorgit.ErrClaimUnbound) {
			t.Errorf("%s: a spoofed revision was not a sync failure: %v", name, err)
		}
	}
	causes := map[string]error{
		"unknown-change":            intentsdd.ErrUnknownChange,
		"archived-earlier":          intentsdd.ErrArchivedEarlier,
		"no-change-intent":          intentsdd.ErrNoChangeIntent,
		"unmappable:slug-too-short": &intentsdd.UnmappableError{Cause: "slug-too-short"},
	}
	for cause, sentinel := range causes {
		_, err := resolve(connectorintent.Revision{}, sentinel, ledgered)
		var unbound *connectorgit.ClaimUnboundError
		if !errors.As(err, &unbound) || unbound.Cause != cause {
			t.Errorf("%s: got %v", cause, err)
		}
	}
	_, err := resolve(good, nil, map[string]bool{})
	var unbound *connectorgit.ClaimUnboundError
	if !errors.As(err, &unbound) || unbound.Cause != "not-ledgered" {
		t.Errorf("not in L: got %v", err)
	}
	if _, err := resolve(connectorintent.Revision{}, errors.New("reader on fire"), ledgered); err == nil || errors.Is(err, connectorgit.ErrClaimUnbound) {
		t.Errorf("a reader error was read as a skip: %v", err)
	}
}

// sddByName resolves the names it holds and reports every other name as unknown.
type sddByName map[string]connectorintent.Revision

func (f sddByName) Resolve(_ context.Context, _ string, name string) (connectorintent.Revision, error) {
	if rev, ok := f[name]; ok {
		return rev, nil
	}
	return connectorintent.Revision{}, intentsdd.ErrUnknownChange
}

// countingAppender counts appends without a store.
type countingAppender struct{ n int }

func (a *countingAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	a.n++
	return store.Record{Seq: int64(a.n), Event: e}, true, nil
}

// M3-INV-6 (§7 row 6): a history without an sdd value neither constructs the sdd provider nor reads
// L; one sdd value (even on several claims) does each exactly once per sync. No store is opened, so
// this test never contends for the embedded database (the end-to-end proofs live in the sdd package).
func TestSDDClaimWorkIsLazy(t *testing.T) {
	digest := strings.Repeat("d", 64)
	builds, reads := 0, 0
	oldBuild, oldRead := newSDDClaimProvider, readSDDLedgered
	t.Cleanup(func() { newSDDClaimProvider, readSDDLedgered = oldBuild, oldRead })
	newSDDClaimProvider = func(string, time.Time) (sddClaimProvider, error) {
		builds++
		return sddByName{"add-auth": {Source: core.SourceIntentSDD, Kind: core.KindChangeIntent, NativeID: "CI-add-auth-xuuno2", ArtifactSHA256: digest, ArtifactPath: "openspec/changes/add-auth"}}, nil
	}
	readSDDLedgered = func(context.Context, *store.Store) (map[string]bool, error) {
		reads++
		return map[string]bool{"CI-add-auth-xuuno2\x00" + digest: true}, nil
	}
	ids, err := newIDs()
	if err != nil {
		t.Fatal(err)
	}
	m := newM3Tenant(t)
	m.commit("plain")
	m.commit("no claim", "Refs: none")
	// syncGit builds its resolver before it opens a sync run; a nil ledger stops it right there.
	if _, err := syncGit(context.Background(), m.root, nil, ids); err == nil {
		t.Fatal("syncGit with a nil ledger succeeded")
	}
	sync := func() {
		t.Helper()
		repo, err := gitcmd.New(m.root)
		if err != nil {
			t.Fatal(err)
		}
		resolver := &commitIntentResolver{root: m.root, now: time.Now()}
		c, err := connectorgit.New(&connectorgit.Deps{Repo: repo, IDs: ids, Logger: logger(), Resolver: resolver})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Sync(context.Background(), &countingAppender{}); err != nil {
			t.Fatal(err)
		}
	}
	sync()
	if builds != 0 || reads != 0 {
		t.Fatalf("no sdd value: builds=%d reads=%d", builds, reads)
	}
	m.commit("claim", "Intent: sdd:add-auth")
	m.commit("claim again", "Intent: sdd:add-auth", "Intent: sdd:nope")
	sync()
	if builds != 1 || reads != 1 {
		t.Fatalf("sdd values: builds=%d reads=%d", builds, reads)
	}
}
