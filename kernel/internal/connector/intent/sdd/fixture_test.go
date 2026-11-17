package sdd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// replayHistory materializes a commit-graph fixture (testdata/<name>/NN-label/{tree,MESSAGE}
// plus optional PARENTS and REF) into a fresh git repository, one commit per step in lexical step
// order, and returns the repository root and the commit SHAs in step order. Each step's tree
// replaces the whole working tree, so deletions and moves between steps are real git changes.
//
//   - PARENTS: whitespace-separated step labels, first parent first. Absent → the previous step
//     (none for the first step, which is a root commit). Present and empty → a root commit.
//   - REF: the branch the step's commit is written to. Absent → main. HEAD is main at the end.
//
// A fixture without PARENTS/REF files is therefore the linear sequence it always was. Lifecycle
// tests run over the whole graph, never a single snapshot (draft 4 §13).
func replayHistory(t *testing.T, name string) (string, []string) {
	t.Helper()
	src := filepath.Join("testdata", name)
	entries, err := os.ReadDir(src)
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
	root := t.TempDir()
	git := func(args ...string) string {
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
	git("init", "-q")
	var shas []string
	byLabel := map[string]string{}
	for i, step := range steps {
		clearTree(t, root)
		copyTree(t, filepath.Join(src, step, "tree"), root)
		git("add", "-A")
		tree := git("write-tree")
		var parents []string
		if raw, err := os.ReadFile(filepath.Join(src, step, "PARENTS")); err == nil {
			for _, label := range strings.Fields(string(raw)) {
				sha, ok := byLabel[label]
				if !ok {
					t.Fatalf("step %s: parent %q is not an earlier step", step, label)
				}
				parents = append(parents, sha)
			}
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		} else if i > 0 {
			parents = []string{shas[i-1]}
		}
		ref := "main"
		if raw, err := os.ReadFile(filepath.Join(src, step, "REF")); err == nil {
			ref = strings.TrimSpace(string(raw))
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		args := []string{"commit-tree", tree, "-F", filepath.Join(mustAbs(t, src), step, "MESSAGE")}
		for _, p := range parents {
			args = append(args, "-p", p)
		}
		sha := git(args...)
		git("update-ref", "refs/heads/"+ref, sha)
		byLabel[step] = sha
		shas = append(shas, sha)
	}
	git("symbolic-ref", "HEAD", "refs/heads/main")
	return root, shas
}

func clearTree(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			t.Fatal(err)
		}
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// requireProfiles fails unless every replayed commit carries the tenant profile. The fixture
// profiles live under a `.proofbound/` directory, which the repository's root .gitignore ignores as
// runtime state; before the re-include rule they were never committed, so on a fresh clone every
// fixture commit had no profile and every lifecycle test would have passed vacuously (draft 4 §13).
func requireProfiles(t *testing.T, root string, shas []string) {
	t.Helper()
	for i, sha := range shas {
		if exec.Command("git", "-C", root, "cat-file", "-e", sha+":.proofbound/sdd-profile.json").Run() != nil {
			t.Fatalf("replayed commit %d has no .proofbound/sdd-profile.json (is the fixture profile committed?)", i+1)
		}
	}
}

// TestFixturesReplayAsDocumented proves the fixtures hold what testdata/README.md says they hold,
// before any provider code exists: a fixture that silently lost its defining property would make
// every later lifecycle test vacuous (LESSON class fixture-not-what-it-claims).
func TestFixturesReplayAsDocumented(t *testing.T) {
	root, shas := replayHistory(t, "openspec-history")
	if len(shas) != 10 {
		t.Fatalf("openspec-history replayed %d commits, want 10", len(shas))
	}
	requireProfiles(t, root, shas)
	has := func(commit, path string) bool {
		return exec.Command("git", "-C", root, "cat-file", "-e", commit+":"+path).Run() == nil
	}
	// 02: proposal only — the change exists with zero requirement entries.
	if !has(shas[1], "openspec/changes/add-login/proposal.md") || has(shas[1], "openspec/changes/add-login/specs") {
		t.Fatal("step 02 is not a proposal-only change")
	}
	// 05: the Logout requirement is gone and the Login statement changed, within the active change.
	at := func(commit, path string) string {
		out, err := exec.Command("git", "-C", root, "show", commit+":"+path).Output()
		if err != nil {
			t.Fatalf("show %s:%s: %v", commit, path, err)
		}
		return string(out)
	}
	before, after := at(shas[3], "openspec/changes/add-login/specs/auth/spec.md"), at(shas[4], "openspec/changes/add-login/specs/auth/spec.md")
	if !strings.Contains(before, "Requirement: Logout") || strings.Contains(after, "Requirement: Logout") || before == after {
		t.Fatal("step 05 does not remove Logout and change Login within the change")
	}
	// 06: archived into a dated directory; the active directory is gone.
	if has(shas[5], "openspec/changes/add-login") || !has(shas[5], "openspec/changes/archive/2026-01-02-add-login/specs/auth/spec.md") {
		t.Fatal("step 06 is not an archive move")
	}
	// 07: squash merge — the archive directory is absent from the first parent and arrives with code.
	if has(shas[5], "openspec/changes/archive/2026-01-03-add-signup") || !has(shas[6], "openspec/changes/archive/2026-01-03-add-signup/specs/auth/spec.md") || !has(shas[6], "src/signup.ts") {
		t.Fatal("step 07 is not a squash-merged, already-archived change with code")
	}
	// 09: --skip-specs archive leaves the main spec untouched; 10: the main spec is deleted.
	if at(shas[7], "openspec/specs/auth/spec.md") != at(shas[8], "openspec/specs/auth/spec.md") || has(shas[9], "openspec/specs/auth/spec.md") {
		t.Fatal("steps 09/10 do not model --skip-specs and main-spec deletion")
	}

	sroot, sshas := replayHistory(t, "speckit-history")
	if len(sshas) != 3 {
		t.Fatalf("speckit-history replayed %d commits, want 3", len(sshas))
	}
	requireProfiles(t, sroot, sshas)
	if exec.Command("git", "-C", sroot, "cat-file", "-e", sshas[0]+":specs/001-export/tasks.md").Run() == nil {
		t.Fatal("speckit step 01 must have no tasks.md")
	}
	s2, _ := exec.Command("git", "-C", sroot, "show", sshas[1]+":specs/001-export/spec.md").Output()
	s3, _ := exec.Command("git", "-C", sroot, "show", sshas[2]+":specs/001-export/spec.md").Output()
	if !strings.Contains(string(s3), "FR-001") || string(s2) == string(s3) {
		t.Fatal("speckit step 03 is not an in-place FR-001 edit")
	}
}

// TestBranchFixtureReplaysAsDocumented proves the commit-graph fixture holds each round-3 shape
// (draft 4 §13) — including that a path-limited HEAD walk really misses the branch states, so a
// test of SDD-INV-31 over this fixture can kill the HEAD-only and path-limited implementations.
func TestBranchFixtureReplaysAsDocumented(t *testing.T) {
	root, shas := replayHistory(t, "branches")
	if len(shas) != 14 {
		t.Fatalf("branches replayed %d commits, want 14", len(shas))
	}
	requireProfiles(t, root, shas)
	out := func(args ...string) string {
		t.Helper()
		b, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return string(b)
	}
	has := func(commit, path string) bool {
		return exec.Command("git", "-C", root, "cat-file", "-e", commit+":"+path).Run() == nil
	}
	parents := func(i int) []string { return strings.Fields(out("rev-list", "--parents", "-n", "1", shas[i]))[1:] }
	set := func(s string) map[string]bool {
		m := map[string]bool{}
		for _, f := range strings.Fields(s) {
			m[f] = true
		}
		return m
	}
	// 01: a root commit that already carries an archived change.
	if len(parents(0)) != 0 || !has(shas[0], "openspec/changes/archive/2026-01-01-seed/specs/core/spec.md") {
		t.Fatal("step 01 is not a root commit with an archived change")
	}
	// 03-05 on a branch; 06 merges it --no-ff and is TREESAME to 02 on openspec/, so git's default
	// path-limited walk from HEAD prunes 03-05 while --all --full-history keeps them.
	if p := parents(5); len(p) != 2 || p[0] != shas[1] || p[1] != shas[4] {
		t.Fatal("step 06 is not a merge of 02 and 05")
	}
	if out("rev-parse", shas[1]+":openspec") != out("rev-parse", shas[5]+":openspec") {
		t.Fatal("step 06 is not TREESAME to 02 on openspec/")
	}
	limited, full := set(out("rev-list", "main", "--", "openspec/changes")), set(out("rev-list", "--all", "--full-history", "--", "openspec/changes"))
	for _, i := range []int{2, 4} { // 03 and 05 change openspec/changes; 04 changes only src/
		if limited[shas[i]] || !full[shas[i]] {
			t.Fatalf("step %02d: path-limited walk does not miss it, or full-history walk does not keep it", i+1)
		}
	}
	// 08: the merge's second parent (07) already archived add-x; its first parent (06) did not.
	if p := parents(7); len(p) != 2 || has(p[0], "openspec/changes/archive/2026-01-05-add-x") || !has(p[1], "openspec/changes/archive/2026-01-05-add-x") {
		t.Fatal("step 08 is not a merge of a side-branch archive")
	}
	// 09: an unmerged branch commit — on the --all walk, not reachable from main.
	if exec.Command("git", "-C", root, "merge-base", "--is-ancestor", shas[8], "main").Run() == nil || !set(out("rev-list", "--all"))[shas[8]] {
		t.Fatal("step 09 is not an unmerged branch commit")
	}
	// 11-12: add-z active and archived at once, the leftover byte-identical (OpenSpec 7ced2a87);
	// 13 removes it (df12368f).
	left, arch := "openspec/changes/add-z/specs/cli-list/spec.md", "openspec/changes/archive/2026-01-08-add-z/specs/cli-list/spec.md"
	for _, i := range []int{10, 11} {
		if !has(shas[i], left) || out("show", shas[i]+":"+left) != out("show", shas[i]+":"+arch) {
			t.Fatalf("step %02d is not the leftover-after-archive state", i+1)
		}
	}
	if has(shas[12], "openspec/changes/add-z") {
		t.Fatal("step 13 does not remove the leftover")
	}
	// 14: a ref tip ("stale") sitting on the leftover state, so a tip-fatal implementation fails
	// (draft 5 §10 D; OpenSpec has 4 remote-branch tips on its dual-presence state).
	if out("rev-parse", "stale") != out("rev-parse", shas[13]+"^{commit}") ||
		out("rev-parse", shas[13]+":openspec/changes") != out("rev-parse", shas[10]+":openspec/changes") {
		t.Fatal("step 14 is not a ref tip at the leftover state")
	}
	// Claims: 04 (branch, inside the edit window), 09 (unmerged), 12 (during the leftover).
	for i, want := range map[int]string{3: "Intent: sdd:add-x", 8: "Intent: sdd:add-y", 11: "Intent: sdd:add-z"} {
		if !strings.Contains(out("log", "-1", "--format=%B", shas[i]), want) {
			t.Fatalf("step %02d does not carry %q", i+1, want)
		}
	}

	// The fenced fixture (SDD-INV-39): a column-0 "## " line inside a fence in a MODIFIED section,
	// followed by two more requirements (OpenSpec improve-init-onboarding shape) — the case that
	// discriminates — plus the indented RENAMED example as the neutral control.
	read := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("testdata", "fenced", "openspec", "changes", rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	disc := read("improve-onboarding/specs/cli-init/spec.md")
	open, tldr, after := strings.Index(disc, "\n```markdown\n"), strings.Index(disc, "\n## TL;DR"), strings.Index(disc, "### Requirement: Success Output")
	if open < 0 || tldr < open || after < tldr || strings.Index(disc[tldr:], "\n```\n") < 0 {
		t.Fatal("fenced fixture lacks a column-0 header inside a fence before later requirements")
	}
	if neutral := read("document-renames/specs/conventions/spec.md"); !strings.Contains(neutral, "\n  ## RENAMED Requirements") {
		t.Fatal("fenced neutral control lost its indented RENAMED example")
	}
	if b, err := os.ReadFile(filepath.Join("testdata", "expected-skips.json")); err != nil || !json.Valid(b) {
		t.Fatalf("expected-skips.json missing or invalid: %v", err)
	}
}

// TestPinnedExpectationsMatchFixtures mechanises the pins M1's tests consume (draft 6 §13): every
// unit in expected-skips.json names a step and a path that exist, and every fenced case in
// expected-obligations.json exists, carries each pinned requirement header verbatim, and either
// discriminates (its OpenSpec-mask list differs from the naive or unmasked list) or is the one
// declared neutral control. The discrimination itself is computed by derive/history_probe.py's
// §4.2 parser and re-proved by M1's SDD-INV-39 test against these pins.
func TestPinnedExpectationsMatchFixtures(t *testing.T) {
	var skips map[string]json.RawMessage
	raw, err := os.ReadFile(filepath.Join("testdata", "expected-skips.json"))
	if err != nil || json.Unmarshal(raw, &skips) != nil {
		t.Fatalf("expected-skips.json: %v", err)
	}
	type skip struct{ Step, Unit, Cause string }
	for fixture, body := range skips {
		if fixture == "schema" || fixture == "note" {
			continue
		}
		var list []skip
		if err := json.Unmarshal(body, &list); err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
		if _, err := os.Stat(filepath.Join("testdata", fixture)); err != nil {
			t.Fatalf("expected-skips names missing fixture %s", fixture)
		}
		for _, s := range list {
			tree := filepath.Join("testdata", fixture)
			if s.Step != "" {
				tree = filepath.Join(tree, s.Step, "tree")
			}
			kind, name, _ := strings.Cut(s.Unit, " ")
			var probe string
			switch kind {
			case "requirement":
				probe = filepath.Join(tree, name)
			case "change":
				probe = filepath.Join(tree, "openspec", "changes", name)
			default:
				t.Fatalf("%s: unit %q has no known kind", fixture, s.Unit)
			}
			if _, err := os.Stat(probe); err != nil || s.Cause == "" {
				t.Fatalf("%s: pinned skip %+v does not name an existing unit (%s)", fixture, s, probe)
			}
		}
	}

	var obl struct {
		Cases map[string]struct {
			OpenSpecMask  [][2]string `json:"openspec_mask"`
			NaiveToggle   [][2]string `json:"naive_toggle"`
			Unmasked      [][2]string `json:"unmasked"`
			Discriminates bool        `json:"discriminates"`
		} `json:"cases"`
	}
	raw, err = os.ReadFile(filepath.Join("testdata", "expected-obligations.json"))
	if err != nil || json.Unmarshal(raw, &obl) != nil {
		t.Fatalf("expected-obligations.json: %v", err)
	}
	neutral := 0
	for path, c := range obl.Cases {
		text, err := os.ReadFile(filepath.Join("testdata", path))
		if err != nil {
			t.Fatalf("pinned case %s missing: %v", path, err)
		}
		if !strings.Contains(string(text), "\n## MODIFIED Requirements\n") && !strings.HasPrefix(string(text), "## MODIFIED Requirements\n") {
			t.Fatalf("%s: lost its column-0 MODIFIED section header", path)
		}
		for _, o := range c.OpenSpecMask {
			if !strings.Contains(string(text), "### Requirement: "+o[1]+"\n") {
				t.Fatalf("%s: pinned obligation %q has no requirement header", path, o[1])
			}
		}
		differs := fmt.Sprint(c.OpenSpecMask) != fmt.Sprint(c.NaiveToggle) || fmt.Sprint(c.OpenSpecMask) != fmt.Sprint(c.Unmasked)
		if differs != c.Discriminates || len(c.OpenSpecMask) == 0 {
			t.Fatalf("%s: discriminates=%v but lists say %v (or no obligations)", path, c.Discriminates, differs)
		}
		if !c.Discriminates {
			neutral++
		}
	}
	if neutral != 1 || len(obl.Cases) < 5 {
		t.Fatalf("want exactly 1 neutral control among >= 5 fenced cases, got %d of %d", neutral, len(obl.Cases))
	}

	// head-on-leftover: HEAD (main) itself ends on the dual-presence state, so a HEAD-fatal
	// implementation fails (round-5 F6).
	root, shas := replayHistory(t, "head-on-leftover")
	requireProfiles(t, root, shas)
	if len(shas) != 2 || exec.Command("git", "-C", root, "cat-file", "-e", "HEAD:openspec/changes/add-w").Run() != nil ||
		exec.Command("git", "-C", root, "cat-file", "-e", "HEAD:openspec/changes/archive/2026-02-01-add-w").Run() != nil {
		t.Fatal("head-on-leftover: HEAD is not on the dual-presence state")
	}
}
