package sdd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

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
