package sdd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// hostile writes one unit per M1-reachable state-local cause into s (not profile-invalid, which
// masks a whole state).
func hostile(s *synth) {
	s.write("openspec/changes/nul-one/specs/core/spec.md", addedDelta("Nul", "The system SHALL \x00 stop."))
	s.write("openspec/changes/bad-utf/specs/core/spec.md", "## ADDED Requirements\n### Requirement: Bad\n\xff\xfe\n")
	s.write("openspec/changes/rename-only/specs/core/spec.md", "## RENAMED Requirements\n- FROM: `### Requirement: Old`\n")
	s.write("openspec/changes/bang/specs/core/spec.md", addedDelta("!!!", "The system SHALL bang."))
	s.write("openspec/changes/dup/specs/core/spec.md", addedDelta("Login!", "A.")+"\n### Requirement: Login?\nB.\n")
	s.write("openspec/changes/reused/specs/core/spec.md", addedDelta("Reuse", "The system SHALL reuse."))
	s.write("openspec/changes/archive/2026-01-01-reused/specs/core/spec.md", addedDelta("Reuse", "The system SHALL reuse."))
	s.write("openspec/changes/a/specs/core/spec.md", addedDelta("Short", "The system SHALL be short."))
	s.write("openspec/changes/archive/undated/specs/core/spec.md", addedDelta("Undated", "U."))
	s.symlink("openspec/changes/linked", "good")
	s.write("openspec/changes/huge/specs/core/spec.md", addedDelta("Huge", strings.Repeat("h", core.MaxPayloadBytes)))
}

var hostileSkips = []string{
	"requirement openspec/changes/nul-one/specs/core/spec.md|nul-byte", "change nul-one|cascade",
	"requirement openspec/changes/bad-utf/specs/core/spec.md|invalid-utf8", "change bad-utf|cascade",
	"requirement openspec/changes/rename-only/specs/core/spec.md|unpaired-rename", "change rename-only|cascade",
	"requirement openspec/changes/bang/specs/core/spec.md|empty-name-part", "change bang|cascade",
	"requirement openspec/changes/dup/specs/core/spec.md|obligation-collision", "change dup|cascade",
	"change reused|name-reuse",
	"change a|slug-too-short",
	"archive undated|archive-name-invalid",
	"change linked|non-regular-file",
	"requirement openspec/changes/huge/specs/core/spec.md|payload-too-large", "change huge|cascade",
}

// TestHistoricalFailClosedIsRecoverable proves SDD-INV-35: every state-local cause, placed (a) at
// a non-tip commit later fixed, (b) at a non-HEAD ref tip, and (c) at HEAD, leaves Revisions
// without error and every clean unit emitted; profile-invalid likewise at a state of its own.
func TestHistoricalFailClosedIsRecoverable(t *testing.T) {
	for _, where := range []string{"history", "tip", "head"} {
		t.Run(where, func(t *testing.T) {
			s := newSynth(t)
			s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "The system SHALL be good."))
			s.commit("clean")
			hostile(s)
			bad := s.commit("hostile")
			s.write(ProfilePath, `{"schema": "proofbound.sdd-profile.v1"}`)
			s.commit("profile broken")
			if where != "head" {
				s.write(ProfilePath, testProfile)
				for _, d := range []string{"nul-one", "bad-utf", "rename-only", "bang", "dup", "reused", "a", "archive", "linked", "huge"} {
					s.remove("openspec/changes/" + d)
				}
				s.commit("fixed")
			}
			if where == "tip" {
				s.git("branch", "stale", bad)
			}
			revs, skips := syncAll(t, s.provider())
			if _, ok := revs[ciKind+"openspec/changes/good"]; !ok {
				t.Fatalf("clean change not emitted: %v", revs)
			}
			for _, want := range append(hostileSkips, "profile|profile-invalid") {
				if !skips[want] {
					t.Errorf("missing skip %s; got %s", want, fmtSkips(skips))
				}
			}
		})
	}
}

// TestObligationDefectsArePerRecord proves SDD-INV-43.
func TestObligationDefectsArePerRecord(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/dup/specs/core/spec.md", addedDelta("Login!", "A.")+"\n### Requirement: Login?\nB.\n")
	s.write("openspec/changes/dup/specs/other/spec.md", addedDelta("Fine", "F."))
	s.write("openspec/changes/rename-only/specs/core/spec.md", "## RENAMED Requirements\n- FROM: `### Requirement: Old`\n")
	s.write("openspec/changes/zero/specs/core/spec.md", "# no sections at all\n")
	s.commit("defects")
	revs, skips := syncAll(t, s.provider())
	for _, want := range []string{"requirement openspec/changes/dup/specs/core/spec.md|obligation-collision", "change dup|cascade", "requirement openspec/changes/rename-only/specs/core/spec.md|unpaired-rename", "change rename-only|cascade"} {
		if !skips[want] {
			t.Errorf("missing %s; got %s", want, fmtSkips(skips))
		}
	}
	if _, ok := revs[reqKind+"openspec/changes/dup/specs/other/spec.md"]; !ok {
		t.Fatal("the clean sibling record was not emitted")
	}
	for k := range skips {
		if strings.Contains(k, "zero") {
			t.Fatalf("zero-entry file produced %s", k)
		}
	}
}

// TestReusedNameIsPerState proves SDD-INV-45.
func TestReusedNameIsPerState(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/archive/2026-01-01-again/specs/core/spec.md", addedDelta("Old", "O."))
	s.write("openspec/changes/again/specs/core/spec.md", addedDelta("New", "N."))
	reuse := s.commit("reuse")
	s.remove("openspec/changes/archive/2026-01-01-again")
	clean := s.commit("clean")
	p := s.provider()
	revs, skips := syncAll(t, p)
	if !skips["change again|name-reuse"] {
		t.Fatalf("no name-reuse skip: %s", fmtSkips(skips))
	}
	if _, ok := revs[reqKind+"openspec/changes/archive/2026-01-01-again/specs/core/spec.md"]; !ok {
		t.Fatal("the reused name's archived record was not emitted")
	}
	if _, ok := revs[ciKind+"openspec/changes/again"]; !ok {
		t.Fatal("the change did not map again at the next state")
	}
	if _, err := p.Resolve(context.Background(), reuse, "again"); outcomeOf(err) != "ErrUnmappable:name-reuse" {
		t.Fatalf("reuse state: %v", err)
	}
	if _, err := p.Resolve(context.Background(), clean, "again"); err != nil {
		t.Fatalf("clean state: %v", err)
	}
}

func profileWith(old, repl string) string { return strings.Replace(testProfile, old, repl, 1) }

func expectProfileInvalid(t *testing.T, name, body string) {
	t.Helper()
	s := newSynth(t)
	s.write(ProfilePath, body)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.commit(name)
	revs, skips := syncAll(t, s.provider())
	if len(revs) != 0 || !skips["profile|profile-invalid"] || len(skips) != 1 {
		t.Errorf("%s: revs=%d skips=%s", name, len(revs), fmtSkips(skips))
	}
}

// TestProfileIsStrict proves SDD-INV-40.
func TestProfileIsStrict(t *testing.T) {
	cases := map[string]string{
		"unknown field":   profileWith(`"kit": "openspec",`, `"kit": "openspec", "extra": 1,`),
		"duplicate field": profileWith(`"kit": "openspec",`, `"kit": "openspec", "kit": "openspec",`),
		"null field":      profileWith(`"kit": "openspec",`, `"kit": null,`),
		"missing field":   profileWith(`"kit": "openspec",`, ``),
		"unknown enum":    profileWith(`"status": "accepted"`, `"status": "approved"`),
		"speckit in M1":   profileWith(`"kit": "openspec"`, `"kit": "speckit"`),
		"mode mismatch":   profileWith(`"mode": "delta-sections"`, `"mode": "feature-sibling"`),
		"blank owner":     profileWith(`"owner": "team-fixture"`, `"owner": "  "`),
		"nul owner":       profileWith(`"owner": "team-fixture"`, `"owner": "a\u0000b"`),
		"non-utf8":        profileWith(`"owner": "team-fixture"`, "\"owner\": \"\xff\""),
		"no owners":       profileWith(`[ { "pattern": "openspec/**", "owner": "team-fixture" } ]`, `[]`),
		"trailing data":   testProfile + "{}",
		"not json":        "not json",
	}
	for name, body := range cases {
		expectProfileInvalid(t, name, body)
	}
}

// TestProfilePathsAreLiteral proves SDD-INV-67.
func TestProfilePathsAreLiteral(t *testing.T) {
	root := `"root": "openspec/changes"`
	for name, repl := range map[string][2]string{
		"pathspec magic":    {root, `"root": ":(glob)openspec/changes"`},
		"exclude magic":     {root, `"root": ":!x"`},
		"parent":            {root, `"root": "../openspec/changes"`},
		"absolute":          {root, `"root": "/openspec/changes"`},
		"trailing slash":    {root, `"root": "openspec/changes/"`},
		"dot segment":       {root, `"root": "./openspec/changes"`},
		"double slash":      {root, `"root": "openspec//changes"`},
		"space":             {root, `"root": "open spec/changes"`},
		"nested archive":    {`"archive": "openspec/changes/archive"`, `"archive": "openspec/changes/x/archive"`},
		"archive not under": {`"archive": "openspec/changes/archive"`, `"archive": "openspec/archive"`},
		"nonstandard glob":  {`"openspec/changes/*/specs/*/spec.md",`, `"openspec/changes/*/spec.md",`},
		"glob without star": {`"openspec/changes/*/specs/*/spec.md",`, `"openspec/changes/x/specs/core/spec.md",`},
		"bad owner pattern": {`"pattern": "openspec/**", "owner"`, `"pattern": ":(glob)**", "owner"`},
	} {
		expectProfileInvalid(t, name, profileWith(repl[0], repl[1]))
	}
}

// TestM1IsOpenSpecOnly proves SDD-INV-64.
func TestM1IsOpenSpecOnly(t *testing.T) {
	root, _ := replayHistory(t, "speckit-history")
	revs, skips := syncAll(t, newProvider(t, root))
	if len(revs) != 0 || len(skips) != 1 || !skips["profile|profile-invalid"] {
		t.Fatalf("revs=%d skips=%s", len(revs), fmtSkips(skips))
	}
}

// TestIdentityOnlyFromProfile proves SDD-INV-41.
func TestIdentityOnlyFromProfile(t *testing.T) {
	s := newSynth(t)
	s.write(ProfilePath, profileWith(`[ { "pattern": "openspec/**", "owner": "team-fixture" } ]`,
		`[ { "pattern": "openspec/changes/first/**", "owner": "owner-one" }, { "pattern": "openspec/changes/first/**", "owner": "owner-two" } ]`))
	s.write("openspec/changes/first/specs/core/spec.md", addedDelta("One", "O."))
	s.write("openspec/changes/second/specs/core/spec.md", addedDelta("Two", "T."))
	s.commit("owners")
	revs, skips := syncAll(t, s.provider())
	if r, ok := revs[reqKind+"openspec/changes/first/specs/core/spec.md"]; !ok || !bytes.Contains(r.Payload, []byte(`"declared_owner":"owner-one"`)) {
		t.Fatalf("first match not used: %s", r.Payload)
	}
	if !skips["requirement openspec/changes/second/specs/core/spec.md|owner-uncovered"] || !skips["change second|cascade"] {
		t.Fatalf("uncovered owner: %s", fmtSkips(skips))
	}
	s2 := newSynth(t)
	s2.write(ProfilePath, profileWith(`"pattern": "openspec/changes/**", "sponsor"`, `"pattern": "elsewhere/**", "sponsor"`))
	s2.write("openspec/changes/first/specs/core/spec.md", addedDelta("One", "O."))
	s2.commit("sponsors")
	if _, skips := syncAll(t, s2.provider()); !skips["change first|sponsor-uncovered"] {
		t.Fatalf("uncovered sponsor: %s", fmtSkips(skips))
	}
}

// TestPayloadStringsAreStoreSafe proves SDD-INV-53.
func TestPayloadStringsAreStoreSafe(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/nul-statement/specs/core/spec.md", addedDelta("Ok", "The system \x00 SHALL."))
	s.write("openspec/changes/nul-name/specs/core/spec.md", "## ADDED Requirements\n### Requirement: Na\x00me\n")
	s.write("openspec/changes/nul-to/specs/core/spec.md", "## RENAMED Requirements\n- FROM: `### Requirement: A`\n- TO: `### Requirement: B\x00`\n")
	s.commit("nul")
	_, skips := syncAll(t, s.provider())
	for _, n := range []string{"nul-statement", "nul-name", "nul-to"} {
		if !skips["requirement openspec/changes/"+n+"/specs/core/spec.md|nul-byte"] {
			t.Errorf("%s not nul-byte: %s", n, fmtSkips(skips))
		}
	}
	expectProfileInvalid(t, "nul sponsor", profileWith(`"sponsor": "team-fixture"`, `"sponsor": "\u0000"`))
}

// TestNonRegularFilesAreUnits proves SDD-INV-54.
func TestNonRegularFilesAreUnits(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/target/specs/core/spec.md", addedDelta("T", "The system SHALL target."))
	s.write("openspec/changes/sym-delta/specs/core/real.md", addedDelta("R", "R."))
	s.symlink("openspec/changes/sym-delta/specs/core/spec.md", "real.md")
	s.write("openspec/changes/tree-delta/specs/core/spec.md/inner.md", "a tree where spec.md belongs")
	s.symlink("openspec/changes/sym-specs/specs", "../target/specs")
	s.symlink("openspec/changes/sym-cap/specs/core", "../../target/specs/core")
	s.symlink("openspec/changes/sym-change", "target")
	s.symlink("openspec/changes/archive/2026-01-01-sym-archived", "../target")
	s.write("openspec/changes/designlink/specs/core/spec.md", addedDelta("D", "The system SHALL design."))
	s.symlink("openspec/changes/designlink/design.md", "../target/specs/core/spec.md")
	s.commit("links", "openspec/changes/gitlinked", "openspec/changes/designlink/vendor")
	revs, skips := syncAll(t, s.provider())
	for _, want := range []string{
		"requirement openspec/changes/sym-delta/specs/core/spec.md|non-regular-file",
		"requirement openspec/changes/tree-delta/specs/core/spec.md|non-regular-file",
		"requirement openspec/changes/sym-specs/specs|non-regular-file",
		"requirement openspec/changes/sym-cap/specs/core|non-regular-file",
		"change sym-change|non-regular-file",
		"change gitlinked|non-regular-file",
		"change sym-archived|non-regular-file",
	} {
		if !skips[want] {
			t.Errorf("missing %s; got %s", want, fmtSkips(skips))
		}
	}
	if _, ok := revs[ciKind+"openspec/changes/designlink"]; !ok {
		t.Fatalf("a symlinked design.md and a nested gitlink must not be causes: %s", fmtSkips(skips))
	}
	if _, ok := revs[ciKind+"openspec/changes/target"]; !ok {
		t.Fatal("regular sibling not emitted")
	}
}

// TestStateLevelObjectTypes proves SDD-INV-61.
func TestStateLevelObjectTypes(t *testing.T) {
	s := newSynth(t)
	s.write("real/x/specs/core/spec.md", addedDelta("X", "X."))
	s.symlink("openspec/changes", "../real")
	s.commit("symlinked root")
	s.remove("openspec/changes")
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.write("openspec/changes/archive", "a blob where the archive belongs")
	s.commit("blob archive")
	s.remove("openspec/changes/archive")
	s.write("openspec/changes/good/diagram.png", "\x89PNG\x00\xff\xfe binary")
	s.commit("clean with binary")
	revs, skips := syncAll(t, s.provider())
	n := 0
	for k := range skips {
		if strings.HasPrefix(k, "profile|non-regular-file") {
			n++
		}
	}
	if n != 1 || len(skips) != 1 {
		t.Fatalf("want one state-level non-regular-file skip per distinct defect (symlinked root, blob archive dedupe only if identical); got %s", fmtSkips(skips))
	}
	if _, ok := revs[ciKind+"openspec/changes/good"]; !ok {
		t.Fatalf("a binary file in a change directory must not be a cause: %s", fmtSkips(skips))
	}
	d := newSynth(t)
	d.remove(ProfilePath)
	d.write(ProfilePath+"/inner", "a directory at the profile path")
	d.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	d.commit("profile dir")
	if revs, skips := syncAll(t, d.provider()); len(revs) != 0 || !skips["profile|profile-invalid"] {
		t.Fatalf("directory profile: %s", fmtSkips(skips))
	}
}

// TestForeignPathsAreIgnored proves SDD-INV-55.
func TestForeignPathsAreIgnored(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.write("docs/caf\xe9.md", "outside the roots")
	s.write("openspec/changes/good/caf\xe9.md", "a non-parsed file inside a change")
	s.write("openspec/changes/caf\xe9/specs/core/spec.md", addedDelta("C", "C."))
	s.commit("latin-1 names")
	revs, skips := syncAll(t, s.provider())
	if _, ok := revs[ciKind+"openspec/changes/good"]; !ok {
		t.Fatalf("good change: %s", fmtSkips(skips))
	}
	if !skips["change caf\xe9|invalid-utf8"] {
		t.Fatalf("non-UTF-8 change name: %s", fmtSkips(skips))
	}
	for k := range skips {
		if strings.Contains(k, "good") || strings.Contains(k, "docs") {
			t.Fatalf("unexpected skip %s", k)
		}
	}
}

// TestPlainFilesAreNotChanges proves SDD-INV-66.
func TestPlainFilesAreNotChanges(t *testing.T) {
	root, shas := replayHistory(t, "resolve-edges")
	p := newProvider(t, root)
	_, skips := syncAll(t, p)
	for k := range skips {
		if strings.Contains(k, "IMPLEMENTATION_ORDER") || strings.Contains(k, "README") {
			t.Fatalf("plain file became a unit: %s", k)
		}
	}
	for _, n := range []string{"IMPLEMENTATION_ORDER.md", "README.md"} {
		if _, err := p.Resolve(context.Background(), shas[6], n); !errors.Is(err, ErrUnknownChange) {
			t.Fatalf("%s: %v", n, err)
		}
	}
}

// TestStateKeyCarriesModes proves SDD-INV-65.
func TestStateKeyCarriesModes(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	regular := s.commit("regular profile")
	s.symlink(ProfilePath, testProfile)
	linked := s.commit("profile as a symlink to identical text")
	repo, err := gitcmd.New(s.root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := NewGitReader(repo)
	ctx := context.Background()
	ka, _ := r.StateKey(ctx, regular)
	kb, _ := r.StateKey(ctx, linked)
	if ka == kb || ka == "" || kb == "" {
		t.Fatalf("keys equal or empty: %q %q", ka, kb)
	}
	if _, skips := syncAll(t, s.provider()); !skips["profile|profile-invalid"] {
		t.Fatalf("symlinked profile mapped: %s", fmtSkips(skips))
	}
	// (b) equal keys ⇒ equal changeOutcome and skips, over every committed history.
	for _, h := range []string{"openspec-history", "branches", "head-on-leftover", "resolve-edges"} {
		root, shas := replayHistory(t, h)
		repo, err := gitcmd.New(root)
		if err != nil {
			t.Fatal(err)
		}
		gr, _ := NewGitReader(repo)
		p, _ := New(gr, timeOne)
		seen := map[string]string{}
		for _, c := range shas {
			k, _ := gr.StateKey(ctx, c)
			arts, _ := gr.Read(ctx, c)
			st, err := p.mapState(arts)
			if err != nil {
				t.Fatal(err)
			}
			sig := stateSignature(st)
			if prev, ok := seen[k]; ok && prev != sig {
				t.Fatalf("%s: equal keys, different outcomes at %s", h, c)
			}
			seen[k] = sig
		}
	}
}

func stateSignature(st *stateResult) string {
	var parts []string
	for n, c := range st.changes {
		parts = append(parts, n+"="+c.skip+"/"+map[bool]string{true: "absent"}[c.absent])
	}
	for _, s := range st.skips {
		parts = append(parts, s.unit+"|"+s.cause+"|"+s.digest)
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// TestResolveIsMemberOfRevisions proves SDD-INV-31.
func TestResolveIsMemberOfRevisions(t *testing.T) {
	for _, h := range []string{"branches", "resolve-edges", "openspec-history"} {
		root, _ := replayHistory(t, h)
		repo, err := gitcmd.New(root)
		if err != nil {
			t.Fatal(err)
		}
		commits, err := repo.Commits(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		gr, _ := NewGitReader(repo)
		listed, _ := gr.Commits(context.Background())
		var want []string
		for _, c := range commits {
			want = append(want, c.SHA)
		}
		if !reflect.DeepEqual(listed, want) {
			t.Fatalf("%s: Commits differs from the git connector's list", h)
		}
		p, _ := New(gr, timeOne)
		revs, err := p.Revisions(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		member := map[string]bool{}
		for _, r := range revs {
			member[r.NativeID+"@"+r.ArtifactSHA256] = true
		}
		names := map[string]bool{}
		_ = filepath.WalkDir(filepath.Join("testdata", h), func(path string, d os.DirEntry, err error) error {
			if err == nil && d.IsDir() && strings.Contains(path, "/openspec/changes/") {
				base := filepath.Base(path)
				if m := datedRE.FindStringSubmatch(base); m != nil {
					base = m[1]
				}
				names[base] = true
			}
			return nil
		})
		checked := 0
		for _, c := range listed {
			for n := range names {
				r, err := p.Resolve(context.Background(), c, n)
				if err != nil {
					continue
				}
				checked++
				if !member[""+r.NativeID+"@"+r.ArtifactSHA256] {
					t.Fatalf("%s: Resolve(%s, %s) = %s not in Revisions", h, c, n, r.NativeID)
				}
			}
		}
		if checked == 0 {
			t.Fatalf("%s: no successful Resolve checked", h)
		}
	}
}

// TestNativeNameResolution proves SDD-INV-32 on explicit cases (the full table is SDD-INV-59's).
func TestNativeNameResolution(t *testing.T) {
	root, shas := replayHistory(t, "openspec-history")
	p := newProvider(t, root)
	ctx := context.Background()
	check := func(step int, n, want string) {
		t.Helper()
		if _, err := p.Resolve(ctx, shas[step], n); outcomeOf(err) != want {
			t.Errorf("step %02d %s: got %s want %s", step+1, n, outcomeOf(err), want)
		}
	}
	check(3, "add-login", "revision")           // active
	check(1, "add-login", "ErrNoChangeIntent")  // proposal only
	check(5, "add-login", "revision")           // archived by this commit
	check(6, "add-login", "ErrArchivedEarlier") // archived earlier
	check(6, "add-signup", "revision")          // squash-merged: archived by its first commit
	check(3, "nope", "ErrUnknownChange")        // unknown
	check(3, "archive", "ErrUnknownChange")     // the archive's own name
}

// recordingReader fails on any read outside the allowed set (SDD-INV-33).
type recordingReader struct {
	Reader
	commit  string
	parents map[string]bool
	archive string
	t       *testing.T
}

func (r *recordingReader) Read(ctx context.Context, c string) ([]Artifact, error) {
	if c != r.commit {
		r.t.Fatalf("Read(%s) outside the named commit", c)
	}
	return r.Reader.Read(ctx, c)
}
func (r *recordingReader) StateKey(context.Context, string) (string, error) {
	r.t.Fatal("Resolve must not need state keys")
	return "", nil
}
func (r *recordingReader) Commits(context.Context) ([]string, error) {
	r.t.Fatal("Resolve must not list history")
	return nil, nil
}
func (r *recordingReader) List(ctx context.Context, c, dir string) ([]Entry, error) {
	if !r.parents[c] || dir != r.archive {
		r.t.Fatalf("List(%s, %s) is not a parent archive listing", c, dir)
	}
	return r.Reader.List(ctx, c, dir)
}

// TestResolveReadSet proves SDD-INV-33.
func TestResolveReadSet(t *testing.T) {
	root, shas := replayHistory(t, "branches")
	repo, err := gitcmd.New(root)
	if err != nil {
		t.Fatal(err)
	}
	gr, _ := NewGitReader(repo)
	ctx := context.Background()
	for i, c := range shas {
		parents, _ := gr.Parents(ctx, c)
		allowed := map[string]bool{}
		for _, p := range parents {
			allowed[p] = true
		}
		rr := &recordingReader{Reader: gr, commit: c, parents: allowed, archive: "openspec/changes/archive", t: t}
		p, _ := New(rr, timeOne)
		for _, n := range []string{"add-x", "add-y", "add-z", "seed", "nope"} {
			if _, err := p.Resolve(ctx, c, n); err != nil && strings.HasPrefix(outcomeOf(err), "reader-error") {
				t.Fatalf("step %d %s: %v", i+1, n, err)
			}
		}
	}
}

// TestProviderDefectsFailNotSkip proves SDD-INV-49: a provider defect (here, a payload that is
// not canonical JSON) fails Revisions; it is never a Skip.
func TestProviderDefectsFailNotSkip(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	s.commit("good")
	defectHook = func(b []byte) []byte { return append([]byte(" "), b...) }
	defer func() { defectHook = nil }()
	p := s.provider()
	if _, err := p.Revisions(context.Background()); err == nil {
		t.Fatalf("defect became %v", p.Skipped())
	}
}

// TestBatchSyncsEndToEnd proves SDD-INV-47: every fixture history, the hostile synthesized one
// included, passes the intent connector's Sync; the oversize unit never reaches append.
func TestBatchSyncsEndToEnd(t *testing.T) {
	providers := map[string]*Provider{}
	for _, h := range []string{"openspec-history", "branches", "head-on-leftover", "resolve-edges", "speckit-history"} {
		root, _ := replayHistory(t, h)
		providers[h] = newProvider(t, root)
	}
	s := newSynth(t)
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "G."))
	hostile(s)
	s.commit("hostile")
	providers["hostile"] = s.provider()
	for name, p := range providers {
		app := &memAppender{}
		c, err := intent.New(&intent.Deps{Providers: []intent.Provider{p}, IDs: testIDs(t), Logger: discard()})
		if err != nil {
			t.Fatal(err)
		}
		res, err := c.Sync(context.Background(), "sdd", app)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, e := range app.events {
			if len(e.Payload) > core.MaxPayloadBytes {
				t.Fatalf("%s: oversize payload appended", name)
			}
		}
		if name == "hostile" && len(res.ByProvider["sdd"].Skipped) == 0 {
			t.Fatal("hostile history reported no skips")
		}
	}
}

// TestRecordIDCollisionIsPerState proves SDD-INV-48 with the pinned colliding pair: at one state
// both records are unmappable (record-id-collision); at two different states they are two
// revisions of one id, bound by digest.
func TestRecordIDCollisionIsPerState(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "record-id-collision.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pin struct {
		Names  []string `json:"names"`
		Suffix string   `json:"suffix"`
	}
	if err := json.Unmarshal(raw, &pin); err != nil || len(pin.Names) != 2 {
		t.Fatalf("pin: %v", err)
	}
	delta := addedDelta("Do", "The system SHALL do.")
	together := newSynth(t)
	for _, n := range pin.Names {
		together.write("openspec/changes/"+n+"/specs/core/spec.md", delta)
	}
	together.commit("both at one state")
	revs, skips := syncAll(t, together.provider())
	for _, n := range pin.Names {
		if !skips["requirement openspec/changes/"+n+"/specs/core/spec.md|record-id-collision"] {
			t.Fatalf("%s not skipped: %s", n, fmtSkips(skips))
		}
	}
	if len(revs) != 0 {
		t.Fatalf("colliding records emitted: %v", revs)
	}
	apart := newSynth(t)
	apart.write("openspec/changes/"+pin.Names[0]+"/specs/core/spec.md", delta)
	apart.commit("first alone")
	apart.remove("openspec/changes/" + pin.Names[0])
	apart.write("openspec/changes/"+pin.Names[1]+"/specs/core/spec.md", delta)
	apart.commit("second alone")
	all, err := apart.provider().Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ids, digests := map[string]bool{}, map[string]bool{}
	for _, r := range all {
		if r.Kind == core.KindRequirement {
			ids[r.NativeID], digests[r.ArtifactSHA256] = true, true
		}
	}
	if len(ids) != 1 || len(digests) != 2 {
		t.Fatalf("want one id with two digests: ids=%v digests=%d", ids, len(digests))
	}
	for id := range ids {
		if !strings.HasSuffix(id, "-"+pin.Suffix) {
			t.Fatalf("pinned suffix drifted: %s", id)
		}
	}
}
