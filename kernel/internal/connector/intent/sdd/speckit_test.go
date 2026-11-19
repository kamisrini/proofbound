package sdd

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// R2 Spec Kit (docs/plans/R2-speckit-spec.md): SDD-INV-69..71.

const specKitProfile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"] },
  "changes": { "mode": "feature-sibling", "root": "specs", "sibling": "tasks.md", "status": "proposed" },
  "owners":   [ { "pattern": "specs/**", "owner": "team-fixture" } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "team-fixture" } ]
}
`

// stateAt maps one commit's state.
func stateAt(t *testing.T, p *Provider, sha string) *stateResult {
	t.Helper()
	arts, err := p.reader.Read(context.Background(), sha)
	if err != nil {
		t.Fatal(err)
	}
	st, err := p.mapState(arts)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func promiseOf(t *testing.T, st *stateResult, kind core.Kind) (intent.Revision, map[string]string) {
	t.Helper()
	var found []intent.Revision
	for _, r := range st.revisions {
		if r.Kind == kind {
			found = append(found, r)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want exactly one %s, got %d", kind, len(found))
	}
	obls := map[string]string{}
	if kind == core.KindRequirement {
		var req intent.Requirement
		if err := decodePayload(found[0].Payload, &req); err != nil {
			t.Fatal(err)
		}
		for _, o := range req.Obligations {
			obls[o.ID] = o.Statement
		}
	}
	return found[0], obls
}

func TestSpecKitHistoryMapsExactly(t *testing.T) {
	root, shas := replayHistory(t, "speckit-history")
	p := newProvider(t, root)

	st1 := stateAt(t, p, shas[0])
	req1, obls1 := promiseOf(t, st1, core.KindRequirement)
	want := map[string]string{"O-FR-001": "System MUST export reports as CSV.", "O-FR-002": "System MUST include a header row.", "O-SC-001": "An export of 10k rows completes in under 2 seconds."}
	if len(obls1) != 3 || obls1["O-FR-001"] != want["O-FR-001"] || obls1["O-FR-002"] != want["O-FR-002"] || obls1["O-SC-001"] != want["O-SC-001"] {
		t.Fatalf("step 01 obligations %v", obls1)
	}
	if len(st1.revisions) != 1 || len(st1.skips) != 0 || !st1.changes["001-export"].absent {
		t.Fatalf("step 01: a change intent without tasks.md, or a skip: %+v", st1)
	}

	st2 := stateAt(t, p, shas[1])
	ci2, _ := promiseOf(t, st2, core.KindChangeIntent)
	req2, _ := promiseOf(t, st2, core.KindRequirement)
	var ci intent.ChangeIntent
	if err := decodePayload(ci2.Payload, &ci); err != nil {
		t.Fatal(err)
	}
	if req2.NativeID != req1.NativeID || len(ci.Targets) != 1 || ci.Targets[0].RecordID != req1.NativeID || strings.Join(ci.Targets[0].ObligationIDs, ",") != "O-FR-001,O-FR-002,O-SC-001" || ci.Status != "proposed" || ci.ArtifactPath != "specs/001-export" {
		t.Fatalf("step 02 change intent %+v", ci)
	}

	st3 := stateAt(t, p, shas[2])
	ci3, _ := promiseOf(t, st3, core.KindChangeIntent)
	req3, obls3 := promiseOf(t, st3, core.KindRequirement)
	if req3.NativeID == req1.NativeID || ci3.NativeID != ci2.NativeID || ci3.ArtifactSHA256 == ci2.ArtifactSHA256 {
		t.Fatalf("step 03: an in-place edit must give a new record and a new change-intent revision")
	}
	if obls3["O-FR-001"] != "System MUST export reports as CSV and JSON." || obls3["O-FR-002"] != want["O-FR-002"] || obls3["O-SC-001"] != want["O-SC-001"] || len(obls3) != 3 {
		t.Fatalf("step 03 obligations %v", obls3)
	}
	if _, skips := syncAll(t, p); len(skips) != 0 {
		t.Fatalf("the Spec Kit history has skips: %s", fmtSkips(skips))
	}

	// Fenced example lines are never obligations; a repeated id and a blank statement are causes.
	s := newSynth(t)
	s.write(ProfilePath, specKitProfile)
	s.write("specs/002-fenced/spec.md", "## Requirements\n\n- **FR-001**: System MUST log in.\n\n```markdown\n- **FR-009**: Example only.\n```\n")
	s.write("specs/003-dup/spec.md", "- **FR-001**: One.\n- **FR-001**: Two.\n")
	s.write("specs/004-prose/spec.md", "# Just prose, no obligations\n")
	sha := s.commit("cases")
	st := stateAt(t, s.provider(), sha)
	var fenced map[string]string
	for _, r := range st.revisions {
		if r.Kind == core.KindRequirement && r.ArtifactPath == "specs/002-fenced/spec.md" {
			var req intent.Requirement
			if err := decodePayload(r.Payload, &req); err != nil {
				t.Fatal(err)
			}
			fenced = map[string]string{}
			for _, o := range req.Obligations {
				fenced[o.ID] = o.Statement
			}
		}
	}
	if len(fenced) != 1 || fenced["O-FR-001"] != "System MUST log in." {
		t.Fatalf("fenced example read as an obligation: %v", fenced)
	}
	skips := map[string]bool{}
	for _, k := range st.skips {
		skips[k.unit+"|"+k.cause] = true
	}
	if len(skips) != 1 || !skips["requirement specs/003-dup/spec.md|obligation-collision"] {
		t.Fatalf("skips %v", skips)
	}
}

func TestSpecKitProfileIsStrict(t *testing.T) {
	mapped := func(name, profile string) int {
		t.Helper()
		s := newSynth(t)
		s.write(ProfilePath, profile)
		s.write("specs/001-export/spec.md", "- **FR-001**: System MUST export.\n")
		s.write("specs/001-export/tasks.md", "- [ ] T001\n")
		s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
		s.commit(name)
		revs, skips := syncAll(t, s.provider())
		if len(revs) == 0 && (!skips["profile|profile-invalid"] || len(skips) != 1) {
			t.Fatalf("%s: no revisions but not profile-invalid: %s", name, fmtSkips(skips))
		}
		return len(revs)
	}
	if n := mapped("valid", specKitProfile); n != 2 {
		t.Fatalf("calibration: the valid Spec Kit profile mapped %d revisions, want 2", n)
	}
	with := func(old, repl string) string {
		if !strings.Contains(specKitProfile, old) {
			t.Fatalf("calibration: %q not in the profile", old)
		}
		return strings.Replace(specKitProfile, old, repl, 1)
	}
	for name, body := range map[string]string{
		"two-level paths":   with(`["specs/*/spec.md"]`, `["specs/*/*/spec.md"]`),
		"paths not root":    with(`["specs/*/spec.md"]`, `["docs/*/spec.md"]`),
		"non-literal root":  with(`"root": "specs"`, `"root": "../specs"`),
		"sibling with path": with(`"sibling": "tasks.md"`, `"sibling": "x/tasks.md"`),
		"no sibling":        with(`, "sibling": "tasks.md"`, ``),
		"archive key":       with(`"root": "specs",`, `"root": "specs", "archive": "specs/archive",`),
		"openspec mode":     with(`"mode": "feature-sibling"`, `"mode": "delta-sections"`),
		"unknown kit":       with(`"kit": "speckit"`, `"kit": "kiro"`),
		"openspec+sibling":  strings.Replace(testProfile, `"mode": "delta-sections",`, `"mode": "delta-sections", "sibling": "tasks.md",`, 1),
	} {
		if n := mapped(name, body); n != 0 {
			t.Errorf("%s: accepted (%d revisions)", name, n)
		}
	}
}

func TestSpecKitResolveAndIDs(t *testing.T) {
	root, shas := replayHistory(t, "speckit-history")
	p := newProvider(t, root)
	ctx := context.Background()
	if _, err := p.Resolve(ctx, shas[0], "001-export"); !errors.Is(err, ErrNoChangeIntent) {
		t.Fatalf("unplanned feature: %v", err)
	}
	rev, err := p.Resolve(ctx, shas[1], "001-export")
	if err != nil || rev.Kind != core.KindChangeIntent || rev.ArtifactPath != "specs/001-export" {
		t.Fatalf("planned feature: %+v %v", rev, err)
	}
	if state, err := p.StateRevision(ctx, shas[1], "001-export"); err != nil || state.ArtifactSHA256 != rev.ArtifactSHA256 {
		t.Fatalf("StateRevision disagrees: %+v %v", state, err)
	}
	for _, sha := range shas {
		if _, err := p.Resolve(ctx, sha, "002-missing"); !errors.Is(err, ErrUnknownChange) {
			t.Fatalf("missing feature: %v", err)
		}
	}
	os, err := changeIntentID(Kit, "install-skills-globally")
	if err != nil || os != "CI-install-skills-globally-57r52f" {
		t.Fatalf("an OpenSpec id changed: %q %v", os, err)
	}
	sk, err := changeIntentID(KitSpecKit, "install-skills-globally")
	if err != nil || sk == os || !strings.HasPrefix(sk, "CI-install-skills-globally-") {
		t.Fatalf("the Spec Kit id does not differ by kit: %q", sk)
	}
}

func decodePayload(raw []byte, v any) error { return json.Unmarshal(raw, v) }
