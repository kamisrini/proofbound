package sdd_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/migration"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// R2 M3 SPEC end-to-end procedures: draft 6/7/8 §7.1 (row 8, M3-INV-8), §7.2 (row 5, M3-INV-5) and
// §7.3 (M3-INV-12), on the embedded store through cli.Run, untagged.

// archivedEvidence is a pinned event id in the repository's five-record historical archive (a
// checks event, seq 1670), so evidence citing it is valid on every ledger that imports the archive.
const archivedEvidence = "01M28TPW9C8R7ND19MNDCJ9GDG"

type m3Refs struct {
	ci          connectorintent.Reference // record_kind change_intent, relation evaluates
	requirement connectorintent.Reference // record_kind requirement
	obligations []string
}

// refsAt computes add-auth's change-intent and requirement references at HEAD WITHOUT a ledger, by
// running the sdd provider (its digests are content-derived, so they equal what sync intent sdd
// ledgers). skipCIDigest excludes an earlier revision.
func (m *m3Tenant) refsAt(skipCIDigest string) m3Refs {
	m.t.Helper()
	repo, err := gitcmd.New(m.root)
	if err != nil {
		m.t.Fatal(err)
	}
	reader, err := sdd.NewGitReader(repo)
	if err != nil {
		m.t.Fatal(err)
	}
	provider, err := sdd.New(reader, time.Unix(1, 0).UTC())
	if err != nil {
		m.t.Fatal(err)
	}
	revs, err := provider.Revisions(context.Background())
	if err != nil {
		m.t.Fatal(err)
	}
	var refs m3Refs
	for _, r := range revs {
		if r.Kind != core.KindChangeIntent || filepath.Base(r.ArtifactPath) != "add-auth" || r.ArtifactSHA256 == skipCIDigest {
			continue
		}
		var ci connectorintent.ChangeIntent
		if err := json.Unmarshal(r.Payload, &ci); err != nil {
			m.t.Fatal(err)
		}
		refs.ci = connectorintent.Reference{RecordKind: "change_intent", Source: string(core.SourceIntentSDD), RecordID: r.NativeID, ArtifactSHA256: r.ArtifactSHA256, Relation: "evaluates"}
		target := ci.Targets[0]
		refs.requirement = connectorintent.Reference{RecordKind: "requirement", Source: target.Source, RecordID: target.RecordID, ArtifactSHA256: target.ArtifactSHA256}
		refs.obligations = append([]string(nil), target.ObligationIDs...)
		sort.Strings(refs.obligations)
	}
	if refs.ci.RecordID == "" || len(refs.obligations) == 0 {
		m.m3Fatalf("no add-auth change intent at HEAD (skipping digest %q)", skipCIDigest)
	}
	return refs
}

func (m *m3Tenant) m3Fatalf(format string, args ...any) {
	m.t.Helper()
	m.t.Fatalf(format, args...)
}

// gitEventID returns the event id of the latest ledgered git event for sha.
func (m *m3Tenant) gitEventID(sha string) string {
	m.t.Helper()
	id := ""
	m.read(store.Filter{Source: core.SourceGit}, func(r store.Record) error {
		if r.Event.NativeID == sha {
			id = r.Event.ID.String()
		}
		return nil
	})
	if id == "" {
		m.t.Fatalf("no git event for %s", sha)
	}
	return id
}

const m3VerdictDir = "docs/verification/verdicts/"

// writeArtifact writes a v2/review artifact with its self-digest (reviews.ArtifactSHA256's rule).
func (m *m3Tenant) writeArtifact(rel string, marshal func(digest string) []byte) {
	m.t.Helper()
	wrap := func(front []byte) []byte {
		return append(append([]byte("---\n"), front...), []byte("\n---\n\nM3 end-to-end fixture.\n")...)
	}
	draft := wrap(marshal(strings.Repeat("0", 64)))
	m.write(rel, string(wrap(marshal(reviews.ArtifactSHA256(draft)))))
}

func (m *m3Tenant) writeReview(id string, refs m3Refs) {
	m.t.Helper()
	path := m3VerdictDir + id + ".md"
	requirement := refs.requirement
	requirement.Relation = "reviews"
	outcomes := []reviews.RequirementReviewOutcome{}
	for _, o := range refs.obligations {
		outcomes = append(outcomes, reviews.RequirementReviewOutcome{ObligationID: o, Outcome: "VERIFIABLE"})
	}
	m.writeArtifact(path, func(digest string) []byte {
		b, err := json.Marshal(reviews.RequirementReview{Schema: "proofbound.requirement-review.v1", ReviewID: id, Requirement: requirement, DeclaredReviewer: "reviewer-b", Outcomes: outcomes, ArtifactPath: path, ArtifactSHA256: digest})
		if err != nil {
			m.t.Fatal(err)
		}
		return b
	})
}

func (m *m3Tenant) writeVerdict(id, commit, evidence string, refs m3Refs) {
	m.t.Helper()
	path := m3VerdictDir + id + ".md"
	requirement := refs.requirement
	requirement.Relation = "evaluates"
	obligations := []reviews.ObligationOutcome{}
	for _, o := range refs.obligations {
		obligations = append(obligations, reviews.ObligationOutcome{Source: requirement.Source, RequirementID: requirement.RecordID, ArtifactSHA256: requirement.ArtifactSHA256, ObligationID: o, Outcome: "SATISFIED", EvidenceEventIDs: []string{evidence}})
	}
	m.writeArtifact(path, func(digest string) []byte {
		b, err := json.Marshal(reviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", VerdictID: id, Status: "ACCEPTABLE", DeclaredReviewer: "reviewer-b", ReviewedCommit: commit, ChangeIntent: refs.ci, Requirements: []connectorintent.Reference{requirement}, Obligations: obligations, Findings: []reviews.Finding{}, ArtifactPath: path, ArtifactSHA256: digest})
		if err != nil {
			m.t.Fatal(err)
		}
		return b
	})
}

func (m *m3Tenant) writeGate(id, sha string) {
	m.t.Helper()
	m.write("gates/"+id+".yaml", `{"schema":"proofbound.gate.v1","id":"`+id+`","description":"M3 readiness for `+id+`","expires":"2099-01-01","mode":"canary","rule":"intent-delivery-readiness","scope_commit":"`+sha+`"}`)
}

var gateLineRE = regexp.MustCompile(`(?m)^gate=(\S+) state=(\S+) `)

func (m *m3Tenant) canary() map[string]string {
	m.t.Helper()
	states := map[string]string{}
	for _, match := range gateLineRE.FindAllStringSubmatch(m.mustRun("gates", "canary"), -1) {
		states[match[1]] = match[2]
	}
	return states
}

func (m *m3Tenant) expectStates(want map[string]string) {
	m.t.Helper()
	got := m.canary()
	for id, state := range want {
		if got[id] != state {
			m.t.Fatalf("gate %s: state %q, want %q (all: %v)", id, got[id], state, got)
		}
	}
}

func mustContain(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("output lacks %q:\n%s", w, out)
		}
	}
}

// §7.1 (row 8, M3-INV-8): on ONE ledger the readiness gate passes X1 (bound, SATISFIED, VERIFIABLE by
// a non-owner) and blocks X2 (only claim skipped) and X3 (one bindable, one skipped claim, with a
// SATISFIED verdict committed for it, which is held because X3 binds no ref).
func TestM3ReadinessGateOnOneLedger(t *testing.T) {
	m := newM3Tenant(t)
	x1 := m.commit("add auth", "Intent: sdd:add-auth")
	m.write("src/x", "two\n")
	x2 := m.commit("unrelated", "Intent: sdd:nope")
	m.write("src/x", "three\n")
	x3 := m.commit("two claims", "Intent: sdd:add-auth", "Intent: sdd:nope")

	m.mustRun("sync", "intent", "sdd")
	mustContain(t, m.mustRun("sync", "git"), "claim-skip unknown-change sdd:nope at "+x2, "claim-skip unknown-change sdd:nope at "+x3)

	refs := m.refsAt("")
	if got := m.sddChangeIntents()[refs.ci.RecordID]; got != refs.ci.ArtifactSHA256 {
		t.Fatalf("the ledgered change intent %q differs from the computed %q", got, refs.ci.ArtifactSHA256)
	}
	m.writeReview("RR-m3-x", refs)
	m.writeVerdict("OV-m3-x1", x1, m.gitEventID(x1), refs)
	m.writeVerdict("OV-m3-x3", x3, m.gitEventID(x3), refs)
	m.commit("verdicts")
	m.writeGate("x1", x1)
	m.writeGate("x2", x2)
	m.writeGate("x3", x3)

	m.mustRun("sync", "git")
	out := m.mustRun("sync", "reviews")
	mustContain(t, out, "held claim-not-ledgered "+m3VerdictDir+"OV-m3-x3.md")
	if strings.Contains(out, "OV-m3-x1") || strings.Contains(out, "RR-m3-x") {
		t.Fatalf("X1's verdict or the review was held:\n%s", out)
	}
	m.expectStates(map[string]string{"x1": "PASS", "x2": "BLOCKED", "x3": "BLOCKED"})

	// GATE-INV-16, end to end: each readiness line says why.
	reasons := map[string]string{}
	for _, match := range gateReasonRE.FindAllStringSubmatch(m.mustRun("gates", "canary"), -1) {
		reasons[match[1]] = match[2]
	}
	if r := reasons["x2"]; r != "commit binds no change intent" {
		t.Fatalf("x2 reason %q", r)
	}
	// A PASS has every count equal and non-zero.
	counts := regexp.MustCompile(`^([1-9]\d*) of ([1-9]\d*) obligations verified SATISFIED; ([1-9]\d*) of ([1-9]\d*) requirement reviews VERIFIABLE$`).FindStringSubmatch(reasons["x1"])
	if counts == nil || counts[1] != counts[2] || counts[3] != counts[4] || counts[1] != counts[3] {
		t.Fatalf("x1 reason %q", reasons["x1"])
	}
}

var gateReasonRE = regexp.MustCompile(`(?m)^gate=(\S+) .* reason="([^"]*)"$`)

// m3ArchiveTenant is §7.2's tenant: X1 claims add-auth, the five-record archive is in the working
// tree, and X4 commits X1's verdict (evidence = the archived id) and the review, all authored
// before any sync.
func m3ArchiveTenant(t *testing.T) (*m3Tenant, string) {
	t.Helper()
	m := newM3Tenant(t)
	archive, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", migration.ArchivePath))
	if err != nil {
		t.Fatal(err)
	}
	m.write(migration.ArchivePath, string(archive))
	x1 := m.commit("add auth", "Intent: sdd:add-auth")
	refs := m.refsAt("")
	m.writeReview("RR-m3-x", refs)
	m.writeVerdict("OV-m3-x1", x1, archivedEvidence, refs)
	m.commit("verdicts")
	m.writeGate("x1", x1)
	return m, x1
}

// §7.2 interleaving A (row 5, M3-INV-5): `sync all` first holds both artifacts and the rebuild
// stays green; recovery appends them and the gate passes. Then §7.3 (M3-INV-12) continues on the
// same ledger: in-place revisions whose prerequisites are missing are refused loudly, the rebuild
// stays green, and recovery appends them.
func TestM3InterleavingAThenRefusal(t *testing.T) {
	m, x1 := m3ArchiveTenant(t)
	mustContain(t, m.mustRun("migrate", "historical-evidence"), "imported=5")
	out := m.mustRun("sync", "all")
	mustContain(t, out,
		"claim-skip not-ledgered sdd:add-auth at "+x1,
		"held claim-not-ledgered "+m3VerdictDir+"OV-m3-x1.md",
		"held requirement-not-ledgered "+m3VerdictDir+"RR-m3-x.md")
	m.mustRun("rebuild")
	m.mustRun("sync", "intent", "sdd")
	m.mustRun("sync", "git")
	if out := m.mustRun("sync", "reviews"); !strings.Contains(out, "appended=2 ") || strings.Contains(out, "held") {
		t.Fatalf("recovery did not append both:\n%s", out)
	}
	m.mustRun("rebuild")
	m.expectStates(map[string]string{"x1": "PASS"})

	// §7.3: a new requirement revision, a new claim, and both artifacts revised in place.
	old := m.refsAt("")
	m.write("openspec/changes/add-auth/specs/core/spec.md", "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users with a second factor.\n")
	x5 := m.commit("tighten auth", "Intent: sdd:add-auth")
	refs := m.refsAt(old.ci.ArtifactSHA256)
	if refs.requirement.RecordID == old.requirement.RecordID {
		t.Fatalf("calibration: the spec edit did not mint a new requirement revision")
	}
	m.writeReview("RR-m3-x", refs)
	m.writeVerdict("OV-m3-x1", x5, archivedEvidence, refs)
	m.commit("revise in place")
	m.writeGate("x5", x5)
	code, out := runCLI(t, m.root, "sync", "all")
	if code == 0 {
		t.Fatalf("a refused revision did not fail sync all:\n%s", out)
	}
	mustContain(t, out,
		"claim-skip not-ledgered sdd:add-auth at "+x5,
		"refused requirement-not-ledgered "+m3VerdictDir+"RR-m3-x.md",
		"refused claim-not-ledgered "+m3VerdictDir+"OV-m3-x1.md")
	m.mustRun("rebuild")
	m.mustRun("sync", "intent", "sdd")
	m.mustRun("sync", "git")
	if out := m.mustRun("sync", "reviews"); !strings.Contains(out, "appended=2 ") || strings.Contains(out, "refused") {
		t.Fatalf("recovery did not append both revisions:\n%s", out)
	}
	m.mustRun("rebuild")
	m.expectStates(map[string]string{"x5": "PASS"})
}

// §7.2 interleaving B (row 5, M3-INV-5): reviews before everything; the review appends once its
// requirement is ledgered, the verdict once its claim is; the rebuild is green after every step.
func TestM3InterleavingB(t *testing.T) {
	m, _ := m3ArchiveTenant(t)
	m.mustRun("migrate", "historical-evidence")
	mustContain(t, m.mustRun("sync", "reviews"), "held claim-not-ledgered "+m3VerdictDir+"OV-m3-x1.md", "held requirement-not-ledgered "+m3VerdictDir+"RR-m3-x.md")
	m.mustRun("rebuild")
	m.mustRun("sync", "intent", "sdd")
	m.mustRun("rebuild")
	out := m.mustRun("sync", "reviews")
	mustContain(t, out, "appended=1 ", "held claim-not-ledgered "+m3VerdictDir+"OV-m3-x1.md")
	m.mustRun("rebuild")
	m.mustRun("sync", "git")
	m.mustRun("rebuild")
	mustContain(t, m.mustRun("sync", "reviews"), "appended=1 ")
	m.mustRun("rebuild")
	m.expectStates(map[string]string{"x1": "PASS"})
}
