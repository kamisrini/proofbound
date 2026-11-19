package sdd_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// R2 M3 end-to-end proofs (docs/plans/R2-sdd-m3-git-binding-spec-draft5.md §7, as amended by drafts
// 6-8). They live here, beside this package's other cli.Run tests, because every embedded-store test
// must share one package: the embedded database listens on one fixed port and go test runs packages
// in parallel.

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
	gitIn(t, m.root, "init", "-q", "-b", "main")
	m.write(sdd.ProfilePath, m3Profile)
	m.write("openspec/changes/add-auth/specs/core/spec.md", "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users.\n")
	return m
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

// commit stages only the tenant's own paths: the state root's ledger files share .proofbound/ with
// the profile and must never be committed.
func (m *m3Tenant) commit(subject string, trailers ...string) string {
	m.t.Helper()
	gitIn(m.t, m.root, "add", "-A", "--", sdd.ProfilePath, "openspec")
	for _, extra := range []string{"src", "docs", "gates"} {
		if _, err := os.Stat(filepath.Join(m.root, extra)); err == nil {
			gitIn(m.t, m.root, "add", "-A", "--", extra)
		}
	}
	msg := subject
	if len(trailers) > 0 {
		msg += "\n\n" + strings.Join(trailers, "\n")
	}
	gitIn(m.t, m.root, "commit", "-q", "--allow-empty", "-m", msg)
	return gitIn(m.t, m.root, "rev-parse", "HEAD")
}

func (m *m3Tenant) mustRun(args ...string) string {
	m.t.Helper()
	code, out := runCLI(m.t, m.root, args...)
	if code != 0 {
		m.t.Fatalf("%v: exit %d\n%s", args, code, out)
	}
	return out
}

// read opens the tenant's state root between commands, exactly as the CLI does.
func (m *m3Tenant) read(f store.Filter, yield func(store.Record) error) {
	m.t.Helper()
	ctx := context.Background()
	ledger, err := store.Open(ctx, store.Config{Root: filepath.Join(m.root, ".proofbound")})
	if err != nil {
		m.t.Fatal(err)
	}
	defer func() {
		if err := ledger.Close(); err != nil {
			m.t.Fatal(err)
		}
	}()
	if err := ledger.ReadEvents(ctx, f, yield); err != nil {
		m.t.Fatal(err)
	}
}

// latestRefs returns, per commit sha, the intent refs of its latest ledgered git event.
func (m *m3Tenant) latestRefs() map[string][]connectorgit.IntentRef {
	refs := map[string][]connectorgit.IntentRef{}
	m.read(store.Filter{Source: core.SourceGit}, func(r store.Record) error {
		var c connectorgit.Commit
		if err := json.Unmarshal(r.Event.Payload, &c); err != nil {
			return err
		}
		refs[c.SHA] = c.IntentRefs
		return nil
	})
	return refs
}

// sddChangeIntents returns the ledgered intent.sdd change-intent revisions as id -> digest.
func (m *m3Tenant) sddChangeIntents() map[string]string {
	out := map[string]string{}
	m.read(store.Filter{Source: core.SourceIntentSDD, Kind: core.KindChangeIntent}, func(r store.Record) error {
		var p struct {
			IntentID       string `json:"intent_id"`
			ArtifactSHA256 string `json:"artifact_sha256"`
		}
		if err := json.Unmarshal(r.Event.Payload, &p); err != nil {
			return err
		}
		out[p.IntentID] = p.ArtifactSHA256
		return nil
	})
	return out
}

// M3-INV-1 (cli half) and M3-INV-5 (claim half): an sdd claim is not-ledgered until `sync intent
// sdd` ledgers its revision; a rebuild succeeds at every step; after recovery the commit binds
// exactly the ledgered revision, and the projected claim row exists.
func TestM3ClaimBindsOnlyALedgeredRevision(t *testing.T) {
	m := newM3Tenant(t)
	x1 := m.commit("add auth", "Intent: sdd:add-auth")

	out := m.mustRun("sync", "git")
	if !strings.Contains(out, "claim-skip not-ledgered sdd:add-auth at "+x1) {
		t.Fatalf("sync git before sync intent sdd:\n%s", out)
	}
	if refs := m.latestRefs(); refs[x1] != nil {
		t.Fatalf("an un-ledgered revision was bound: %+v", refs[x1])
	}
	m.mustRun("rebuild")

	m.mustRun("sync", "intent", "sdd")
	cis := m.sddChangeIntents()
	if len(cis) != 1 {
		t.Fatalf("change intents=%v", cis)
	}
	if out := m.mustRun("sync", "git"); strings.Contains(out, "claim-skip") {
		t.Fatalf("a ledgered revision was skipped:\n%s", out)
	}
	refs := m.latestRefs()[x1]
	if len(refs) != 1 || refs[0].Provider != "sdd" || cis[refs[0].RecordID] != refs[0].ArtifactSHA256 {
		t.Fatalf("bound %+v, ledgered %v", refs, cis)
	}
	m.mustRun("rebuild")
	if out := m.mustRun("intent", "check", "--commit", x1); !strings.Contains(out, "intent=sdd:"+refs[0].RecordID+"@"+refs[0].ArtifactSHA256) {
		t.Fatalf("the projected claim row is missing:\n%s", out)
	}
}
