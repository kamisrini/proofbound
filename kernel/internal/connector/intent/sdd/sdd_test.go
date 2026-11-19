package sdd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// replayTree commits a single-tree fixture (testdata/<name>, no steps) and returns root and SHA.
func replayTree(t *testing.T, name string) (string, string) {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
			"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
			"GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")
	copyTree(t, filepath.Join("testdata", name), root)
	run("add", "-A")
	run("commit", "-q", "-m", name)
	return root, run("rev-parse", "HEAD")
}

func newProvider(t *testing.T, root string) *Provider {
	t.Helper()
	repo, err := gitcmd.New(root)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewGitReader(repo)
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(r, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type pinnedSkip struct{ Step, Unit, Cause string }

func loadSkipPins(t *testing.T) map[string][]pinnedSkip {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-skips.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	out := map[string][]pinnedSkip{}
	for k, v := range doc {
		if k == "schema" || k == "note" {
			continue
		}
		var list []pinnedSkip
		if err := json.Unmarshal(v, &list); err != nil {
			t.Fatal(err)
		}
		out[k] = list
	}
	return out
}

func isHistory(name string) bool {
	_, err := os.Stat(filepath.Join("testdata", name, "01-"+firstStepSuffix(name)))
	return err == nil
}

func firstStepSuffix(name string) string {
	entries, _ := os.ReadDir(filepath.Join("testdata", name))
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "01-") {
			return strings.TrimPrefix(e.Name(), "01-")
		}
	}
	return "\x00"
}

func stepLabels(t *testing.T, name string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var steps []string
	for _, e := range entries {
		if e.IsDir() {
			steps = append(steps, e.Name())
		}
	}
	sort.Strings(steps)
	return steps
}

// TestSkipSetsArePinned proves SDD-INV-46: every fixture's Skip set equals the pin exactly, and
// every revision the provider returns passes intent.ValidateRevision.
func TestSkipSetsArePinned(t *testing.T) {
	for fixture, want := range loadSkipPins(t) {
		t.Run(fixture, func(t *testing.T) {
			var p *Provider
			stepOf := map[string]string{}
			if isHistory(fixture) {
				root, shas := replayHistory(t, fixture)
				for i, label := range stepLabels(t, fixture) {
					stepOf[shas[i]] = label
				}
				p = newProvider(t, root)
			} else {
				root, sha := replayTree(t, fixture)
				stepOf[sha] = ""
				p = newProvider(t, root)
			}
			revs, err := p.Revisions(context.Background())
			if err != nil {
				t.Fatalf("Revisions: %v", err)
			}
			for _, r := range revs {
				if err := intent.ValidateRevision(r); err != nil {
					t.Fatalf("%s: %v", r.ArtifactPath, err)
				}
			}
			var got []string
			for _, s := range p.Skipped() {
				if !errors.Is(s.Err, intent.ErrUnmappable) {
					t.Fatalf("skip %+v does not wrap ErrUnmappable", s)
				}
				got = append(got, stepOf[s.Commit]+"|"+s.Unit+"|"+s.Cause)
			}
			var exp []string
			for _, s := range want {
				exp = append(exp, s.Step+"|"+s.Unit+"|"+s.Cause)
			}
			sort.Strings(got)
			sort.Strings(exp)
			if strings.Join(got, "\n") != strings.Join(exp, "\n") {
				t.Fatalf("skips:\n got  %q\n want %q", got, exp)
			}
		})
	}
}

// TestResolveIsTotal proves SDD-INV-59 over the pinned rows: each (step, name) yields its pinned
// outcome, and a skipped change's Resolve cause equals its Skip cause.
func TestResolveIsTotal(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-resolve.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for history, body := range doc {
		if history == "schema" || history == "note" {
			continue
		}
		var rows []struct{ Step, Name, Outcome string }
		if err := json.Unmarshal(body, &rows); err != nil {
			t.Fatal(err)
		}
		root, shas := replayHistory(t, history)
		shaOf := map[string]string{}
		for i, label := range stepLabels(t, history) {
			shaOf[label] = shas[i]
		}
		p := newProvider(t, root)
		for _, row := range rows {
			_, err := p.Resolve(context.Background(), shaOf[row.Step], row.Name)
			if got := outcomeOf(err); got != row.Outcome {
				t.Errorf("%s %s %q: got %s want %s (err=%v)", history, row.Step, row.Name, got, row.Outcome, err)
			}
		}
	}
	t.Run("corners", testResolveTotalityCorners)
}

// failingReader fails every read, to prove a reader failure is never a tenant sentinel.
type failingReader struct{ Reader }

func (failingReader) Read(context.Context, string) ([]Artifact, error) {
	return nil, errors.New("disk on fire")
}

// testResolveTotalityCorners (a TestResolveIsTotal subtest) completes SDD-INV-59's row: the synthesized corners the replay
// helper cannot commit, the Skipped()/Resolve agreement sweep, and a failing reader (M1 review F2).
func testResolveTotalityCorners(t *testing.T) {
	ctx := context.Background()
	s := newSynth(t)
	s.remove(ProfilePath)
	s.write("openspec/changes/early/specs/core/spec.md", addedDelta("Early", "E."))
	noProfile := s.commit("before opting in")
	s.write(ProfilePath, testProfile)
	s.remove("openspec/changes")
	s.write("real/x/specs/core/spec.md", addedDelta("X", "X."))
	s.symlink("openspec/changes", "../real")
	symRoot := s.commit("symlinked root")
	s.remove("openspec/changes")
	s.write("openspec/changes/archive", "a blob archive")
	blobArchive := s.commit("blob archive")
	s.remove("openspec/changes/archive")
	s.write("openspec/changes/target/specs/core/spec.md", addedDelta("T", "T."))
	s.symlink("openspec/changes/linked", "target")
	s.write("openspec/changes/symdelta/specs/core/real.md", addedDelta("R", "R."))
	s.symlink("openspec/changes/symdelta/specs/core/spec.md", "real.md")
	clean := s.commit("change-level corners")
	p := s.provider()
	for _, c := range []struct {
		commit, name, want string
	}{
		{noProfile, "early", "ErrUnknownChange"},
		{symRoot, "x", "ErrUnmappable:non-regular-file"},
		{symRoot, "archive", "ErrUnknownChange"},
		{blobArchive, "target", "ErrUnmappable:non-regular-file"},
		{blobArchive, "archive", "ErrUnknownChange"},
		{clean, "linked", "ErrUnmappable:non-regular-file"},
		{clean, "symdelta", "ErrUnmappable:cascade"},
		{clean, "target", "revision"},
	} {
		if _, err := p.Resolve(ctx, c.commit, c.name); outcomeOf(err) != c.want {
			t.Errorf("Resolve(%s): got %s want %s", c.name, outcomeOf(err), c.want)
		}
	}
	// Agreement sweep: at every commit of every committed history and this one, each change Skip
	// mapState reports is exactly Resolve's ErrUnmappable cause for that name.
	type history struct {
		p      *Provider
		r      Reader
		commit []string
	}
	var hs []history
	for _, h := range []string{"openspec-history", "branches", "head-on-leftover", "resolve-edges"} {
		root, shas := replayHistory(t, h)
		repo, err := gitcmd.New(root)
		if err != nil {
			t.Fatal(err)
		}
		gr, _ := NewGitReader(repo)
		hp, _ := New(gr, timeOne)
		hs = append(hs, history{hp, gr, shas})
	}
	repo, _ := gitcmd.New(s.root)
	gr, _ := NewGitReader(repo)
	sp, _ := New(gr, timeOne)
	hs = append(hs, history{sp, gr, []string{noProfile, symRoot, blobArchive, clean}})
	checked := 0
	for _, h := range hs {
		for _, c := range h.commit {
			arts, err := h.r.Read(ctx, c)
			if err != nil {
				t.Fatal(err)
			}
			st, err := h.p.mapState(arts)
			if err != nil {
				t.Fatal(err)
			}
			for _, sk := range st.skips {
				n, ok := strings.CutPrefix(sk.unit, "change ")
				if !ok {
					continue
				}
				checked++
				if _, err := h.p.Resolve(ctx, c, n); outcomeOf(err) != "ErrUnmappable:"+sk.cause {
					t.Errorf("%s at %s: Skipped says %s, Resolve says %s", n, c, sk.cause, outcomeOf(err))
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("agreement sweep checked nothing")
	}
	// A failing reader yields none of the tenant sentinels.
	fp, _ := New(failingReader{gr}, timeOne)
	_, err := fp.Resolve(ctx, clean, "target")
	var u *UnmappableError
	if err == nil || errors.Is(err, ErrUnknownChange) || errors.Is(err, ErrArchivedEarlier) || errors.Is(err, ErrNoChangeIntent) || errors.As(err, &u) {
		t.Fatalf("a reader failure surfaced as a tenant outcome: %v", err)
	}
}

func outcomeOf(err error) string {
	var u *UnmappableError
	switch {
	case err == nil:
		return "revision"
	case errors.As(err, &u):
		return "ErrUnmappable:" + u.Cause
	case errors.Is(err, ErrNoChangeIntent):
		return "ErrNoChangeIntent"
	case errors.Is(err, ErrArchivedEarlier):
		return "ErrArchivedEarlier"
	case errors.Is(err, ErrUnknownChange):
		return "ErrUnknownChange"
	}
	return "reader-error: " + err.Error()
}

const (
	reqKind = "requirement.recorded "
	ciKind  = "change_intent.recorded "
)

// TestRecordIDsAreDeterministic proves SDD-INV-7: every vector in testdata/id-vectors.json is
// reproduced exactly, non-ASCII inputs included (kills strings.ToUpper / ToLower).
func TestRecordIDsAreDeterministic(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "id-vectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Vectors []struct {
			Kind, Section, Name, Key, ID, Error string
			Digest                              string `json:"artifact_sha256"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, v := range doc.Vectors {
		var got string
		var err error
		switch v.Kind {
		case "obligation":
			got, err = obligationID(v.Section, v.Name)
		case "change-intent":
			got, err = changeIntentID(v.Name)
		case "requirement":
			got, err = requirementID(v.Key, v.Digest)
		default:
			t.Fatalf("unknown vector kind %q", v.Kind)
		}
		if v.Error != "" {
			if err == nil || err.Error() != v.Error {
				t.Errorf("%+v: got %q err=%v, want error %s", v, got, err, v.Error)
			}
			continue
		}
		if err != nil || got != v.ID {
			t.Errorf("%+v: got %q err=%v", v, got, err)
		}
	}
	if len(doc.Vectors) < 10 {
		t.Fatal("too few vectors")
	}
}

// TestObligationIDsAreNotPositional proves SDD-INV-5.
func TestObligationIDsAreNotPositional(t *testing.T) {
	ids := func(text string) []string {
		found, _ := parseDelta(text)
		var out []string
		for _, e := range found {
			id, err := obligationID(e.Sec, e.Name)
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, id)
		}
		sort.Strings(out)
		return out
	}
	base := "## ADDED Requirements\n\n### Requirement: Alpha\nA.\n\n### Requirement: Beta\nB.\n"
	inserted := "## ADDED Requirements\n\n### Requirement: Zeta\nZ.\n\n### Requirement: Alpha\nA.\n\n### Requirement: Beta\nB.\n"
	reordered := "## ADDED Requirements\n\n### Requirement: Beta\nB.\n\n### Requirement: Alpha\nA.\n"
	if !reflect.DeepEqual(ids(base), ids(reordered)) {
		t.Fatalf("reordering changed ids: %v vs %v", ids(base), ids(reordered))
	}
	got := ids(inserted)
	for _, id := range ids(base) {
		if sort.SearchStrings(got, id) == len(got) || got[sort.SearchStrings(got, id)] != id {
			t.Fatalf("inserting an entry changed id %s: %v", id, got)
		}
	}
}

func naiveToggle(lines []string) []bool {
	mask := make([]bool, len(lines))
	inside := false
	for i, l := range lines {
		t := strings.TrimLeft(l, " \t")
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			inside = !inside
			mask[i] = true
			continue
		}
		mask[i] = inside
	}
	return mask
}

func noMask(lines []string) []bool { return make([]bool, len(lines)) }

// TestCodeFencesAreMasked proves SDD-INV-39: every fenced case reproduces all three pinned
// lists (OpenSpec mask, naive toggle, no mask), so a stale pin and an unmasked or naive parser
// both fail.
func TestCodeFencesAreMasked(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-obligations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Cases map[string]struct {
			OpenSpecMask [][2]string `json:"openspec_mask"`
			NaiveToggle  [][2]string `json:"naive_toggle"`
			Unmasked     [][2]string `json:"unmasked"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	long := map[string]string{"ADD": "ADDED", "MOD": "MODIFIED", "REM": "REMOVED", "REN": "RENAMED"}
	list := func(text string, mask func([]string) []bool) [][2]string {
		found, _ := parseDeltaMasked(text, mask)
		out := [][2]string{}
		for _, e := range found {
			out = append(out, [2]string{long[e.Sec], e.Name})
		}
		return out
	}
	for path, c := range doc.Cases {
		b, err := os.ReadFile(filepath.Join("testdata", path))
		if err != nil {
			t.Fatal(err)
		}
		text := string(b)
		for name, pair := range map[string]struct {
			mask func([]string) []bool
			want [][2]string
		}{"openspec": {fenceMask, c.OpenSpecMask}, "naive": {naiveToggle, c.NaiveToggle}, "none": {noMask, c.Unmasked}} {
			if got := list(text, pair.mask); !reflect.DeepEqual(got, pair.want) && (len(got) != 0 || len(pair.want) != 0) {
				t.Errorf("%s %s: got %v want %v", path, name, got, pair.want)
			}
		}
	}
}

// TestParserMatchesOpenSpecReader proves SDD-INV-57 on round 5's divergence table.
func TestParserMatchesOpenSpecReader(t *testing.T) {
	cases := []struct {
		name, text string
		want       []string
		unpaired   int
	}{
		{"removed bullet", "## REMOVED Requirements\n- `### Requirement: Legacy Export`\n", []string{"REM Legacy Export"}, 0},
		{"two-space title", "## ADDED  Requirements\n\n### Requirement: Foo\nThe system SHALL foo.\n", nil, 0},
		{"nbsp after ##", "## MODIFIED Requirements\n\n### Requirement: Foo\nThe system SHALL foo.\n", []string{"MOD Foo"}, 0},
		{"one-space ## line", "## ADDED Requirements\n### Requirement: A\nx\n## \n### Requirement: B\ny\n", []string{"ADD A", "ADD B"}, 0},
		{"long s title", "## ADDED Requirementſ\n### Requirement: A\nx\n", nil, 0},
		{"interleaved renames", "## RENAMED Requirements\n- FROM: `### Requirement: a`\n- FROM: `### Requirement: b`\n- TO: `### Requirement: x`\n- TO: `### Requirement: y`\n", []string{"REN b"}, 2},
		{"closing sequence", "## ADDED Requirements ##\n### Requirement: Foo ###\nx\n", []string{"ADD Foo"}, 0},
		{"crlf and bom", "\ufeff## ADDED Requirements\r\n### Requirement: Foo\r\nThe system SHALL foo.\r\n", []string{"ADD Foo"}, 0},
	}
	for _, c := range cases {
		found, unpaired := parseDelta(c.text)
		var got []string
		for _, e := range found {
			got = append(got, e.Sec+" "+e.Name)
		}
		if !reflect.DeepEqual(got, c.want) || unpaired != c.unpaired {
			t.Errorf("%s: got %v (unpaired %d) want %v (unpaired %d)", c.name, got, unpaired, c.want, c.unpaired)
		}
	}
	if found, _ := parseDelta("## ADDED Requirements\r\n### Requirement: Foo\r\nThe system SHALL foo.\r\n"); found[0].Statement != "The system SHALL foo." {
		t.Fatalf("CRLF statement = %q", found[0].Statement)
	}
}

// TestNoBusinessDecisionIsEmitted proves SDD-INV-2.
func TestNoBusinessDecisionIsEmitted(t *testing.T) {
	root, _ := replayHistory(t, "openspec-history")
	revs, err := newProvider(t, root).Revisions(context.Background())
	if err != nil || len(revs) == 0 {
		t.Fatalf("revs=%d err=%v", len(revs), err)
	}
	for _, r := range revs {
		switch r.Kind {
		case core.KindRequirement:
			var v intent.Requirement
			if err := json.Unmarshal(r.Payload, &v); err != nil || len(v.AuthorizedBy) != 0 {
				t.Fatalf("%s: authorized_by=%v err=%v", r.ArtifactPath, v.AuthorizedBy, err)
			}
		case core.KindChangeIntent:
		default:
			t.Fatalf("emitted kind %s", r.Kind)
		}
	}
}

// TestDeltaSectionsMapToClosedRelations proves SDD-INV-9: every entry of every section is an
// `implements` target of its change intent (REMOVED included).
func TestDeltaSectionsMapToClosedRelations(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/rework/specs/core/spec.md", "## ADDED Requirements\n### Requirement: New\nN.\n\n## MODIFIED Requirements\n### Requirement: Kept\nK.\n\n## REMOVED Requirements\n### Requirement: Gone\n**Reason**: obsolete.\n\n## RENAMED Requirements\n- FROM: `### Requirement: Old`\n- TO: `### Requirement: Fresh`\n")
	s.commit("rework")
	revs, _ := syncAll(t, s.provider())
	ci, ok := revs[ciKind+"openspec/changes/rework"]
	if !ok {
		t.Fatalf("no change intent: %v", revs)
	}
	var v intent.ChangeIntent
	if err := json.Unmarshal(ci.Payload, &v); err != nil || len(v.Targets) != 1 {
		t.Fatalf("targets=%v err=%v", v.Targets, err)
	}
	want := []string{"O-ADD-NEW", "O-MOD-KEPT", "O-REM-GONE", "O-REN-OLD"}
	if v.Targets[0].Relation != "implements" || !reflect.DeepEqual(v.Targets[0].ObligationIDs, want) {
		t.Fatalf("target=%+v want %v", v.Targets[0], want)
	}
}

// TestCanonicalizationStability proves SDD-INV-12.
func TestCanonicalizationStability(t *testing.T) {
	root, _ := replayHistory(t, "branches")
	a, err := newProvider(t, root).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, err := newProvider(t, root).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) || len(a) == 0 {
		t.Fatalf("len %d vs %d", len(a), len(b))
	}
	for i := range a {
		if !bytes.Equal(a[i].Payload, b[i].Payload) {
			t.Fatalf("payload %d differs", i)
		}
		c, err := core.Canonicalize(a[i].Payload)
		if err != nil || !bytes.Equal(c, a[i].Payload) {
			t.Fatalf("payload %d is not canonical", i)
		}
	}
}

// TestChangeManifestDigest proves SDD-INV-18: a change's digest moves with any file of its
// directory or the profile; its record's digest does not move with proposal.md.
func TestChangeManifestDigest(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/add-x/specs/core/spec.md", addedDelta("Do", "The system SHALL do."))
	s.write("openspec/changes/add-x/proposal.md", "one")
	s.commit("one")
	first, _ := syncAll(t, s.provider())
	s.write("openspec/changes/add-x/proposal.md", "two")
	s.commit("two")
	second, _ := syncAll(t, s.provider())
	var cis []string
	for k, r := range second {
		if strings.HasPrefix(k, ciKind) {
			cis = append(cis, r.ArtifactSHA256)
		}
	}
	reqs := 0
	for k := range second {
		if strings.HasPrefix(k, reqKind) {
			reqs++
		}
	}
	if len(cis) != 1 || cis[0] == first[ciKind+"openspec/changes/add-x"].ArtifactSHA256 {
		t.Fatalf("change digest did not move: %v", cis)
	}
	if reqs != 1 || second[reqKind+"openspec/changes/add-x/specs/core/spec.md"].ArtifactSHA256 != first[reqKind+"openspec/changes/add-x/specs/core/spec.md"].ArtifactSHA256 {
		t.Fatal("record digest moved with proposal.md")
	}
}

// TestArchiveNeverBreaksSync proves SDD-INV-20 over the leftover-after-archive, --skip-specs,
// and deleted-main-spec histories.
func TestArchiveNeverBreaksSync(t *testing.T) {
	for _, h := range []string{"openspec-history", "branches", "head-on-leftover", "resolve-edges"} {
		root, _ := replayHistory(t, h)
		if _, err := newProvider(t, root).Revisions(context.Background()); err != nil {
			t.Fatalf("%s: %v", h, err)
		}
	}
}

// TestPayloadIsPinnedByManifest proves SDD-INV-26.
func TestPayloadIsPinnedByManifest(t *testing.T) {
	build := func(owner string, noise bool) map[string]intent.Revision {
		s := newSynth(t)
		s.write(ProfilePath, strings.Replace(testProfile, `"owner": "team-fixture"`, `"owner": "`+owner+`"`, 1))
		s.write("openspec/changes/x/specs/core/spec.md", addedDelta("Do", "The system SHALL do."))
		if noise {
			s.write("openspec/specs/core/spec.md", "# main spec, never ingested\n")
			s.write("openspec/changes/other/proposal.md", "another change")
			s.write("src/app.go", "package app\n")
			s.git("config", "user.name", "someone else")
		}
		s.commit("x")
		revs, _ := syncAll(t, s.provider())
		return revs
	}
	base, renamed, noisy := build("team-a", false), build("team-b", false), build("team-a", true)
	k := reqKind + "openspec/changes/x/specs/core/spec.md"
	if base[k].ArtifactSHA256 == renamed[k].ArtifactSHA256 {
		t.Fatal("owner change did not change the digest")
	}
	for key, r := range base {
		if !bytes.Equal(noisy[key].Payload, r.Payload) || noisy[key].ArtifactSHA256 != r.ArtifactSHA256 {
			t.Fatalf("%s moved with a non-manifest input", key)
		}
	}
}

// TestEmptyChangeIsAbsent proves SDD-INV-28.
func TestEmptyChangeIsAbsent(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/draft/proposal.md", "# draft")
	s.commit("draft")
	revs, skips := syncAll(t, s.provider())
	if len(revs) != 0 || len(skips) != 0 {
		t.Fatalf("revs=%v skips=%v", revs, skips)
	}
}

// TestMainSpecsAreNotIngested proves SDD-INV-30.
func TestMainSpecsAreNotIngested(t *testing.T) {
	root, _ := replayHistory(t, "openspec-history")
	revs, err := newProvider(t, root).Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range revs {
		if strings.HasPrefix(r.ArtifactPath, "openspec/specs/") {
			t.Fatalf("main spec ingested: %s", r.ArtifactPath)
		}
	}
}

// TestDigestIsManifestOfExactBytes proves SDD-INV-34.
func TestDigestIsManifestOfExactBytes(t *testing.T) {
	delta := addedDelta("Do", "The system SHALL do.")
	s := newSynth(t)
	s.write("openspec/changes/x/specs/core/spec.md", delta)
	s.commit("x")
	revs, _ := syncAll(t, s.provider())
	r := revs[reqKind+"openspec/changes/x/specs/core/spec.md"]
	show := func(p string) []byte {
		out, err := exec.Command("git", "-C", s.root, "cat-file", "blob", "HEAD:"+p).Output()
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	line := func(p string) string {
		sum := sha256.Sum256(show(p))
		return hex.EncodeToString(sum[:]) + "  " + p + "\n"
	}
	lines := []string{line(".proofbound/sdd-profile.json"), line("openspec/changes/x/specs/core/spec.md")}
	sum := sha256.Sum256([]byte(strings.Join(lines, "")))
	if r.ArtifactSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("digest %s is not the independent manifest %x", r.ArtifactSHA256, sum)
	}
	crlf := newSynth(t)
	crlf.write("openspec/changes/x/specs/core/spec.md", strings.ReplaceAll(delta, "\n", "\r\n"))
	crlf.commit("x")
	crevs, _ := syncAll(t, crlf.provider())
	c := crevs[reqKind+"openspec/changes/x/specs/core/spec.md"]
	var a, b intent.Requirement
	_ = json.Unmarshal(r.Payload, &a)
	_ = json.Unmarshal(c.Payload, &b)
	if c.ArtifactSHA256 == r.ArtifactSHA256 || !reflect.DeepEqual(a.Obligations, b.Obligations) {
		t.Fatalf("CRLF: digest equal=%v obligations %v vs %v", c.ArtifactSHA256 == r.ArtifactSHA256, a.Obligations, b.Obligations)
	}
}

// TestRequirementRecordsAreContentAddressed proves SDD-INV-44.
func TestRequirementRecordsAreContentAddressed(t *testing.T) {
	s := newSynth(t)
	s.write("openspec/changes/x/specs/core/spec.md", addedDelta("Do", "The system SHALL do."))
	s.commit("one")
	s.write("openspec/changes/x/specs/core/spec.md", addedDelta("Do", "The system SHALL do it well."))
	s.commit("two")
	revs, err := s.provider().Revisions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]int{}
	for _, r := range revs {
		if r.Kind == core.KindRequirement {
			ids[r.NativeID]++
		}
	}
	if len(ids) != 2 {
		t.Fatalf("want two record ids, one revision each: %v", ids)
	}
	for id, n := range ids {
		if n != 1 {
			t.Fatalf("%s has %d revisions", id, n)
		}
	}
}

func fmtSkips(m map[string]bool) string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return fmt.Sprint(out)
}

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

// collisionNames returns the pinned record-id collision pair (testdata/record-id-collision.json);
// the pair collides only under testProfile's exact bytes.
func collisionNames(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "record-id-collision.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pin struct {
		Names []string `json:"names"`
	}
	if err := json.Unmarshal(raw, &pin); err != nil || len(pin.Names) != 2 {
		t.Fatalf("pin: %v", err)
	}
	return pin.Names
}

// dataState writes, under testProfile, one unit per state-local cause that profile can carry:
// hostile()'s set plus the record-id collision pair, beside one clean change (fresh-data) that no
// clean state carries.
func dataState(t *testing.T, s *synth) []string {
	s.write(ProfilePath, testProfile)
	hostile(s)
	s.write("openspec/changes/fresh-data/specs/core/spec.md", addedDelta("Fresh", "The system SHALL be fresh."))
	want := append([]string(nil), hostileSkips...)
	for _, n := range collisionNames(t) {
		s.write("openspec/changes/"+n+"/specs/core/spec.md", addedDelta("Do", "The system SHALL do."))
		want = append(want, "requirement openspec/changes/"+n+"/specs/core/spec.md|record-id-collision")
	}
	return want
}

// narrowState writes the two coverage causes, which need a profile narrower than testProfile, and
// gives the covered change good a statement no other state carries.
func narrowState(s *synth) []string {
	p := profileWith(`[ { "pattern": "openspec/**", "owner": "team-fixture" } ]`,
		`[ { "pattern": "openspec/changes/good/**", "owner": "team-fixture" }, { "pattern": "openspec/changes/sponsorless/**", "owner": "team-fixture" } ]`)
	s.write(ProfilePath, strings.Replace(p, `"pattern": "openspec/changes/**", "sponsor"`, `"pattern": "openspec/changes/good/**", "sponsor"`, 1))
	s.write("openspec/changes/ownerless/specs/core/spec.md", addedDelta("Owner", "O."))
	s.write("openspec/changes/sponsorless/specs/core/spec.md", addedDelta("Sponsor", "S."))
	s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "The system SHALL be narrow."))
	return []string{"requirement openspec/changes/ownerless/specs/core/spec.md|owner-uncovered", "change sponsorless|sponsor-uncovered"}
}

// TestHistoricalFailClosedIsRecoverable proves SDD-INV-35: every state-local cause (hostile()'s
// set, the record-id collision pair, owner- and sponsor-uncovered, and profile-invalid), placed
// (a) at a non-tip commit later fixed, (b) at a non-HEAD ref tip, and (c) at HEAD, leaves Revisions
// without error and every clean unit emitted. The causes need three profiles, so they span three
// states, and each of the three "head" variants puts a different one of them at HEAD (round-1 F8:
// previously only profile-invalid was ever at HEAD).
func TestHistoricalFailClosedIsRecoverable(t *testing.T) {
	broken := func(s *synth) []string {
		s.write(ProfilePath, `{"schema": "proofbound.sdd-profile.v1"}`)
		return []string{"profile|profile-invalid"}
	}
	for _, v := range []struct {
		where string
		order []string // the three defect states, in commit order
		fixed bool
	}{
		{"history", []string{"data", "narrow", "broken"}, true},
		{"tip", []string{"data", "narrow", "broken"}, true},
		{"head-data", []string{"narrow", "broken", "data"}, false},
		{"head-narrow", []string{"data", "broken", "narrow"}, false},
		{"head-profile", []string{"data", "narrow", "broken"}, false},
	} {
		t.Run(v.where, func(t *testing.T) {
			s := newSynth(t)
			s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "The system SHALL be good."))
			s.commit("clean")
			var want []string
			bad := ""
			for _, st := range v.order {
				switch st {
				case "data":
					want = append(want, dataState(t, s)...)
				case "narrow":
					want = append(want, narrowState(s)...)
				case "broken":
					want = append(want, broken(s)...)
				}
				c := s.commit(st)
				if st == "data" {
					bad = c
				}
			}
			if v.fixed {
				s.write(ProfilePath, testProfile)
				for _, d := range append([]string{"nul-one", "bad-utf", "rename-only", "bang", "dup", "reused", "a", "archive", "linked", "huge", "ownerless", "sponsorless", "fresh-data"}, collisionNames(t)...) {
					s.remove("openspec/changes/" + d)
				}
				s.write("openspec/changes/good/specs/core/spec.md", addedDelta("Good", "The system SHALL be good."))
				s.commit("fixed")
			}
			if v.where == "tip" {
				s.git("branch", "stale", bad)
			}
			p := s.provider()
			all, err := p.Revisions(context.Background())
			if err != nil {
				t.Fatalf("Revisions: %v", err)
			}
			skips := map[string]bool{}
			for _, sk := range p.Skipped() {
				skips[sk.Unit+"|"+sk.Cause] = true
			}
			// "Every other unit emitted" (round lowfix F2): each of these exists ONLY in a state
			// that also carries Skips, so dropping a defect state's clean revisions fails here.
			fresh, narrow := false, false
			for _, r := range all {
				switch {
				case r.Kind == core.KindChangeIntent && r.ArtifactPath == "openspec/changes/fresh-data":
					fresh = true
				case r.Kind == core.KindRequirement && r.ArtifactPath == "openspec/changes/good/specs/core/spec.md" && strings.Contains(string(r.Payload), "SHALL be narrow."):
					narrow = true
				}
			}
			if !fresh || !narrow {
				t.Fatalf("a defect state's clean units were not emitted: fresh-data=%v narrow good=%v", fresh, narrow)
			}
			for _, w := range want {
				if !skips[w] {
					t.Errorf("missing skip %s; got %s", w, fmtSkips(skips))
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
		"upper-case key":  profileWith(`"kit": "openspec",`, `"KIT": "openspec",`),
		"case-twin keys":  profileWith(`"kit": "openspec",`, `"kit": "speckit", "Kit": "openspec",`),
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
	s.write("openspec/changes/notes-only/specs/no-changes.md", "No spec files are needed for this fix.\n")
	s.write("openspec/changes/notes-only/proposal.md", "# notes only")
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
	// A regular file directly under specs/ (OpenSpec's real "no-changes.md" shape, M1 review F1) is
	// ignored, as OpenSpec ignores it: no unit, no cause, and the change is simply spec-less.
	for k := range skips {
		if strings.Contains(k, "notes-only") {
			t.Fatalf("a plain note under specs/ became %s", k)
		}
	}
	if _, err := s.provider().Resolve(context.Background(), s.git("rev-parse", "HEAD"), "notes-only"); !errors.Is(err, ErrNoChangeIntent) {
		t.Fatalf("notes-only change: %v", err)
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
