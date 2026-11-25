package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// The Review bench's background job and reading aids (docs/plans/ui-reviewer-spec.md R1).

// benchJob is the one job a server runs at a time: the tests of a commit, or a seal (which runs them
// first). It lives on the server, not on a request: closing the tab does not stop it (§1.1, F6).
type benchJob struct {
	Kind    string         `json:"kind"` // "tests" or "seal"
	Commit  string         `json:"commit"`
	Running bool           `json:"running"`
	Lines   []string       `json:"lines"`
	Error   string         `json:"error,omitempty"`
	Result  map[string]any `json:"result,omitempty"`
}

type benchJobs struct {
	mu  sync.Mutex
	job *benchJob
	wg  sync.WaitGroup
}

var errJobRunning = errors.New("a test run or a seal is already running on this server; wait for it to finish")

// start begins fn as the server's job, refusing while another runs.
func (b *benchJobs) start(life context.Context, kind, commit string, fn func(ctx context.Context, say func(string)) (map[string]any, error)) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.job != nil && b.job.Running {
		return errJobRunning
	}
	job := &benchJob{Kind: kind, Commit: commit, Running: true, Lines: []string{}}
	b.job = job
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		say := func(line string) {
			b.mu.Lock()
			job.Lines = append(job.Lines, printableText(line))
			b.mu.Unlock()
		}
		result, err := fn(life, say)
		b.mu.Lock()
		defer b.mu.Unlock()
		job.Running, job.Result = false, result
		if err != nil {
			job.Error = printableText(err.Error())
		}
	}()
	return nil
}

// snapshot is the current job for commit (any commit when commit is empty), copied.
func (b *benchJobs) snapshot(commit string) (benchJob, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.job == nil || commit != "" && b.job.Commit != commit {
		return benchJob{}, false
	}
	out := *b.job
	out.Lines = append([]string{}, b.job.Lines...)
	return out, true
}

// serveJob is GET /review/job?commit=: read-only (§1.1, F6).
func (s *uiServer) serveJob(w http.ResponseWriter, r *http.Request) {
	job, ok := s.jobs.snapshot(r.URL.Query().Get("commit"))
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"none": true})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// keepTestsDecisions copies the reviewer's decisions from old onto a fresh run, matched by path and
// lines; a hunk that is now caught needs none.
func keepTestsDecisions(fresh testsResult, old *testsResult) testsResult {
	if old == nil {
		return fresh
	}
	given := map[string]testsHunk{}
	for _, h := range old.Hunks {
		given[areaHunk{Path: h.Path, Lines: h.Lines}.key()] = h
	}
	for i := range fresh.Hunks {
		h := &fresh.Hunks[i]
		if h.Result == testsCaught {
			continue
		}
		if d, ok := given[areaHunk{Path: h.Path, Lines: h.Lines}.key()]; ok {
			h.Decision, h.Reason = d.Decision, d.Reason
		}
	}
	return fresh
}

// writeLocalDraft writes the bench's draft.json for sha.
func writeLocalDraft(root, sha string, d draftFile) error {
	body, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	dir := draftDir(root, sha)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".draft-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(append(body, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	_ = os.Chmod(tmp.Name(), 0o644)
	return os.Rename(tmp.Name(), filepath.Join(dir, "draft.json"))
}

// --- the flags on the code (§1.3) ---

const (
	flagArea  = "outside the area"
	flagTests = "tests did not notice"
	flagProv  = "no recorded agent edit"
)

// diffLine is one line of "The change", with the flags of an added line.
type diffLine struct {
	Text  string   `json:"text"`
	Kind  string   `json:"kind"` // "add", "del", "hunk", "file", ""
	Flags []string `json:"flags,omitempty"`
}

// lineSpan reads the hunk labels of the area and tests checks: "line 3", "lines 1-39", or a whole-file
// label ("file added"); removed lines and summaries carry no added line.
func lineSpan(label string) (lo, hi int, whole, ok bool) {
	switch label {
	case "file added", "binary file", "mode change":
		return 0, 0, true, true
	}
	rest, found := strings.CutPrefix(label, "lines ")
	if !found {
		if rest, found = strings.CutPrefix(label, "line "); !found {
			return 0, 0, false, false
		}
	}
	a, b, has := strings.Cut(rest, "-")
	lo, err := strconv.Atoi(a)
	if err != nil {
		return 0, 0, false, false
	}
	hi = lo
	if has {
		if hi, err = strconv.Atoi(b); err != nil {
			return 0, 0, false, false
		}
	}
	return lo, hi, false, true
}

// lineFlags maps path → new line number (0: every line) → flags.
type lineFlags map[string]map[int][]string

func (f lineFlags) add(path, label, flag string) {
	lo, hi, whole, ok := lineSpan(label)
	if !ok {
		return
	}
	if f[path] == nil {
		f[path] = map[int][]string{}
	}
	if whole {
		f[path][0] = appendOnce(f[path][0], flag)
		return
	}
	for n := lo; n <= hi && n-lo < 100000; n++ {
		f[path][n] = appendOnce(f[path][n], flag)
	}
}

func (f lineFlags) at(path string, n int) []string {
	var out []string
	for _, fl := range f[path][0] {
		out = appendOnce(out, fl)
	}
	for _, fl := range f[path][n] {
		out = appendOnce(out, fl)
	}
	return out
}

func appendOnce(list []string, s string) []string {
	if containsString(list, s) {
		return list
	}
	return append(list, s)
}

// benchFlags gathers the flags of a draft: area hunks, tests hunks not caught, lines no session wrote.
func benchFlags(d draftFile) lineFlags {
	f := lineFlags{}
	for _, a := range d.OutOfArea {
		f.add(a.Path, a.Lines, flagArea)
	}
	if d.Tests != nil {
		for _, h := range d.Tests.Hunks {
			if h.Result != testsCaught {
				f.add(h.Path, h.Lines, flagTests)
			}
		}
	}
	if d.Provenance != nil {
		for _, pf := range d.Provenance.Files {
			for _, n := range pf.Unrecorded {
				f.add(pf.Path, "line "+strconv.Itoa(n), flagProv)
			}
		}
	}
	return f
}

// benchDiff is the change a reviewer reads, line by line with flags; flagged added lines past the
// 1,500-line cut are listed instead (§1.3, F11).
func benchDiff(ctx context.Context, root, sha string, flags lineFlags) ([]diffLine, []string) {
	// The reviewer's own git settings must not change what is parsed (code review F3).
	out := gitOutput(ctx, root, "-c", "core.quotePath=false", "show", "--stat", "--patch", "--no-color", "--no-ext-diff", "--no-renames", "--src-prefix=a/", "--dst-prefix=b/", "--format=commit %H%nAuthor: %an%nDate:   %ad%n%n    %s%n", "--end-of-options", sha)
	lines := strings.Split(out, "\n")
	var view []diffLine
	var hidden []string
	file, n, inHunk := "", 0, false
	for i, l := range lines {
		d := diffLine{Text: printableText(l)}
		switch {
		case strings.HasPrefix(l, "diff --git "):
			file, inHunk = "", false
			d.Kind = "file"
		case !inHunk && strings.HasPrefix(l, "+++ "):
			file = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(l, "+++ "), "b/"), "\t")
			d.Kind = "file"
		case !inHunk && strings.HasPrefix(l, "--- "):
			d.Kind = "file"
		case strings.HasPrefix(l, "@@ "):
			inHunk = true
			if f := strings.Fields(l); len(f) >= 3 {
				if start, _, ok := hunkRange(f[2], "+"); ok {
					n = start
				}
			}
			d.Kind = "hunk"
		case inHunk && strings.HasPrefix(l, "+"):
			d.Kind, d.Flags = "add", flags.at(file, n)
			n++
		case inHunk && strings.HasPrefix(l, "-"):
			d.Kind = "del"
		case inHunk && strings.HasPrefix(l, " "):
			n++
		}
		if i < 1500 {
			view = append(view, d)
		} else if len(d.Flags) > 0 && len(hidden) < 200 {
			hidden = append(hidden, fmt.Sprintf("%s line %d: %s", printableUnit(file), n-1, strings.Join(d.Flags, ", ")))
		}
	}
	if len(lines) > 1500 {
		view = append(view, diffLine{Text: fmt.Sprintf("… %d more lines; run `git show %s` to read the rest", len(lines)-1500, short(sha, 12))})
	}
	// A flagged path the patch never shows (binary, or not in it) is listed, never silently dropped.
	shown := map[string]bool{}
	for _, l := range lines {
		if p, ok := strings.CutPrefix(l, "+++ b/"); ok {
			shown[strings.TrimSuffix(p, "\t")] = true
		}
	}
	var paths []string
	for p := range flags {
		if !shown[p] {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		hidden = append(hidden, printableUnit(p)+": flagged, not shown in the patch")
	}
	return view, hidden
}

// waitingForReview is whether the period's readiness gate reads NO REVIEW for commit.
func waitingForReview(rep auditReport, commit string) bool {
	for _, c := range rep.Controls {
		if c.ID != "C1" {
			continue
		}
		for _, e := range c.Evidence {
			if e.Commit == commit && strings.HasPrefix(e.Text, checkNoReview+":") {
				return true
			}
		}
	}
	return false
}

// testsConfigProblem is why the tests file at the commit's parent is invalid, or "".
func testsConfigProblem(ctx context.Context, root, sha string) string {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	if _, found, err := readTestsConfig(ctx, root, parent); found && err != nil {
		return printableUnit(err.Error())
	}
	return ""
}

// --- the spec beside the promises (§1.4) ---

// specText is one file of the claimed spec, as the area check reads it.
type specText struct {
	Path string `json:"path"`
	Text string `json:"text"`
	// Changed: the commit under review edits this file too (code review F4); the text shown is the
	// one before it, which is what the area check enforces.
	Changed bool `json:"changed_by_commit,omitempty"`
}

// benchSpec reads the Markdown files and scope.paths beside each scope.paths the area check read,
// from the commit's first parent: the spec the commit was written against, which is also the area
// the check enforces. A spec folder with no scope.paths shows nothing here.
func benchSpec(ctx context.Context, root, sha string, area areaResult) []specText {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	if parent == "" {
		return nil
	}
	var out []specText
	for _, scope := range area.Files {
		dir := path.Dir(scope)
		names := strings.Split(strings.TrimSpace(gitOutput(ctx, root, "-c", "core.quotePath=false", "ls-tree", "--name-only", parent, "--", dir+"/")), "\n")
		for _, name := range names {
			if name == "" || !strings.HasSuffix(name, ".md") && name != scope || len(out) >= 8 {
				continue
			}
			raw, err := gitBytes(ctx, root, "show", parent+":"+name)
			if err != nil {
				continue
			}
			if len(raw) > 64<<10 {
				raw = append(raw[:64<<10], []byte("\n…")...)
			}
			t := specText{Path: name, Text: printableText(string(bytes.ToValidUTF8(raw, []byte("?"))))}
			if strings.TrimSpace(gitOutput(ctx, root, "diff", "--name-only", parent, sha, "--", name)) != "" {
				t.Changed = true
			}
			out = append(out, t)
		}
	}
	return out
}
