package cli

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// Tests that can fail (docs/plans/tests-can-fail-spec.md, CLI-INV-66..68): for a claiming commit,
// undo each hunk of code it changed, one at a time, run the project's declared tests, and record
// whether they noticed. Opt-in: nothing runs unless .proofbound/tests.json exists in the commit's
// parent, so the commit cannot weaken its own test command.

const (
	testsConfigPath   = ".proofbound/tests.json"
	testsConfigSchema = "proofbound.tests.v1"
	testsSchema       = "proofbound.tests-review.v1"
	testsRecordPrefix = "docs/verification/verdicts/TESTS-"
	testsMaxHunks     = 20
	testsMaxTimeout   = 3600

	testsCaught     = "caught"
	testsNotCaught  = "not caught"
	testsNotRun     = "not run"
	testsUndoFailed = "undo failed"
	testsPass       = "pass"
	testsFail       = "fail"
	testsTimeout    = "timeout"
	// testsUnreliable: the tests passed before the undo runs and failed when run again, unchanged,
	// after them: they depend on state or on having run before, so no result counts (review F1).
	testsUnreliable = "unreliable"
)

type testsConfig struct {
	Schema         string   `json:"schema"`
	Run            string   `json:"run"`
	Paths          []string `json:"paths"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// testsHunk is one undone hunk, its result, and (at seal) the reviewer's decision when not caught.
type testsHunk struct {
	Path     string `json:"path"`
	Lines    string `json:"lines"`
	Result   string `json:"result"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// testsResult is what review draft writes into draft.json and seal rebuilds by running again.
type testsResult struct {
	Declared bool        `json:"declared"`
	Invalid  string      `json:"invalid,omitempty"`
	Run      string      `json:"run,omitempty"`
	Baseline string      `json:"baseline,omitempty"`
	Whole    string      `json:"whole,omitempty"`
	Hunks    []testsHunk `json:"hunks"`
}

// testsUnit is one hunk of code to undo: its label and the one-hunk patch that undoes it.
type testsUnit struct {
	Path, Lines string
	patch       []byte // empty: cannot be undone hunk by hunk (binary, mode change, over the limit)
}

// readTestsConfig reads .proofbound/tests.json at rev; found is false when there is none.
func readTestsConfig(ctx context.Context, root, rev string) (testsConfig, bool, error) {
	var cfg testsConfig
	if len(rev) != 40 {
		return cfg, false, nil
	}
	raw, err := gitBytes(ctx, root, "show", rev+":"+testsConfigPath)
	if err != nil {
		return cfg, false, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, true, fmt.Errorf("%s is invalid: %v", testsConfigPath, err)
	}
	switch {
	case cfg.Schema != testsConfigSchema:
		return cfg, true, fmt.Errorf("%s is invalid: schema must be %q", testsConfigPath, testsConfigSchema)
	case strings.TrimSpace(cfg.Run) == "" || strings.ContainsAny(cfg.Run, "\n\r\x00"):
		return cfg, true, fmt.Errorf("%s is invalid: run must be one non-empty line", testsConfigPath)
	case cfg.TimeoutSeconds < 0 || cfg.TimeoutSeconds > testsMaxTimeout:
		return cfg, true, fmt.Errorf("%s is invalid: timeout_seconds must be 0..%d", testsConfigPath, testsMaxTimeout)
	}
	for _, p := range cfg.Paths {
		if err := checkAreaPattern(p); err != nil {
			return cfg, true, fmt.Errorf("%s is invalid: path %q: %v", testsConfigPath, p, err)
		}
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 300
	}
	return cfg, true, nil
}

// testsUnits lists the code hunks a claiming commit changed (not housekeeping, not test paths), each
// with the patch that undoes it, plus the patch that undoes all of them. It runs nothing.
func testsUnits(ctx context.Context, root, sha string, cfg testsConfig) ([]testsUnit, []byte, error) {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	from := parent
	if len(from) != 40 {
		from = emptyTree
	}
	raw, err := gitBytes(ctx, root, "diff", "--no-renames", "--no-ext-diff", "--ignore-submodules=none", "--name-status", "-z", from, sha, "--")
	if err != nil {
		return nil, nil, fmt.Errorf("tests: cannot list the paths of %s", sha)
	}
	var units []testsUnit
	var whole bytes.Buffer
	parts := strings.Split(string(raw), "\x00")
	for i := 0; i+1 < len(parts); i += 2 {
		status, p := parts[i], parts[i+1]
		if status == "" {
			break
		}
		if housekeepingPath(p) || inArea(cfg.Paths, p) {
			continue
		}
		diff, err := gitBytes(ctx, root, "--literal-pathspecs", "diff", "-U0", "--no-renames", "--no-ext-diff", "--no-color", "--binary", from, sha, "--", p)
		if err != nil {
			return nil, nil, fmt.Errorf("tests: cannot diff %s in %s", p, sha)
		}
		whole.Write(diff)
		header, hunks := splitHunks(diff)
		if len(hunks) == 0 {
			// A binary file, an empty file added or deleted, or a mode change: undone as one unit
			// with the whole file's patch (review F2).
			units = append(units, testsUnit{Path: p, Lines: wholeFileLabel(diff), patch: diff})
			continue
		}
		for _, h := range hunks {
			lines, ok := hunkLines(string(bytes.SplitN(h, []byte("\n"), 2)[0]))
			if !ok {
				return nil, nil, fmt.Errorf("tests: unreadable hunk header in %s", p)
			}
			patch := append(append([]byte{}, header...), h...)
			units = append(units, testsUnit{Path: p, Lines: lines, patch: patch})
		}
	}
	if len(units) > testsMaxHunks {
		for i := testsMaxHunks; i < len(units); i++ {
			units[i].patch = nil
		}
	}
	return units, whole.Bytes(), nil
}

func wholeFileLabel(diff []byte) string {
	switch {
	case bytes.Contains(diff, []byte("GIT binary patch")) || bytes.Contains(diff, []byte("Binary files ")):
		return "binary file"
	case bytes.Contains(diff, []byte("\nnew file mode ")):
		return "file added"
	case bytes.Contains(diff, []byte("\ndeleted file mode ")):
		return "file deleted"
	}
	return "mode change"
}

// splitHunks splits one file's -U0 diff into its header and its hunks (each starting at "@@ ").
func splitHunks(diff []byte) ([]byte, [][]byte) {
	lines := bytes.SplitAfter(diff, []byte("\n"))
	var header []byte
	var hunks [][]byte
	for _, l := range lines {
		switch {
		case bytes.HasPrefix(l, []byte("@@ ")):
			hunks = append(hunks, append([]byte{}, l...))
		case len(hunks) == 0:
			header = append(header, l...)
		default:
			hunks[len(hunks)-1] = append(hunks[len(hunks)-1], l...)
		}
	}
	return header, hunks
}

// testsRunner runs the declared command in a throwaway worktree of the commit. It is a variable so
// tests can count runs.
var testsRunner = func(ctx context.Context, dir, run string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", run)
	cmd.Dir = dir
	cmd.Env = testsEnv(dir)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	cmd.WaitDelay = 2 * time.Second
	killGroupOnCancel(cmd) // a timeout stops every process the tests started (review F4)
	err := cmd.Run()
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		return testsTimeout
	case err != nil:
		return testsFail
	}
	return testsPass
}

// testsEnv is the caller's environment without git's repository variables, and with git's
// repository search stopped above the copy, so the tests cannot find the reviewer's repository.
func testsEnv(dir string) []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_CEILING_DIRECTORIES="+filepath.Dir(dir))
}

// commitTests runs the tests with nothing undone, with each hunk undone, with the whole change
// undone, and with nothing undone again (spec §1, review F1). Every run gets a fresh copy of the
// commit's files with no .git, so the tests cannot ask git what was undone and nothing they leave
// behind reaches the next run.
func commitTests(ctx context.Context, root, sha string) (testsResult, error) {
	res := testsResult{Hunks: []testsHunk{}}
	if len(commitClaims(ctx, root, sha)) == 0 {
		return res, nil
	}
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	cfg, found, err := readTestsConfig(ctx, root, parent)
	if !found {
		return res, nil
	}
	res.Declared = true
	if err != nil {
		res.Invalid = err.Error()
		return res, nil
	}
	res.Run = cfg.Run
	units, whole, err := testsUnits(ctx, root, sha, cfg)
	if err != nil {
		return res, err
	}
	dir, err := os.MkdirTemp("", "proofbound-tests-")
	if err != nil {
		return res, err
	}
	defer os.RemoveAll(dir)
	pristine, run := filepath.Join(dir, "files"), filepath.Join(dir, "run")
	archive, err := gitBytes(ctx, root, "archive", "--format=tar", sha)
	if err != nil {
		return res, fmt.Errorf("tests: cannot export %s", sha[:12])
	}
	if err := extractTar(archive, pristine); err != nil {
		return res, fmt.Errorf("tests: cannot export %s: %v", sha[:12], err)
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	fresh := func() error {
		if err := os.RemoveAll(run); err != nil {
			return err
		}
		return copyTree(pristine, run)
	}
	undo := func(patch []byte) bool {
		cmd := exec.CommandContext(ctx, "git", "apply", "-R", "--unidiff-zero", "--whitespace=nowarn", "-")
		cmd.Dir = run
		cmd.Env = testsEnv(run)
		cmd.Stdin = bytes.NewReader(patch)
		return cmd.Run() == nil
	}
	attempt := func(patch []byte) (string, error) {
		if err := fresh(); err != nil {
			return "", err
		}
		if patch != nil && !undo(patch) {
			return testsUndoFailed, nil
		}
		return testsRunner(ctx, run, cfg.Run, timeout), nil
	}
	first, err := attempt(nil)
	if err != nil {
		return res, err
	}
	res.Baseline, res.Whole = first, testsNotRun
	for _, u := range units {
		h := testsHunk{Path: u.Path, Lines: u.Lines, Result: testsNotRun}
		if first == testsPass && len(u.patch) > 0 {
			got, err := attempt(u.patch)
			if err != nil {
				return res, err
			}
			h.Result = caughtWord(got)
		}
		res.Hunks = append(res.Hunks, h)
	}
	if first != testsPass {
		return res, nil
	}
	if len(whole) > 0 {
		got, err := attempt(whole)
		if err != nil {
			return res, err
		}
		res.Whole = caughtWord(got)
	}
	again, err := attempt(nil)
	if err != nil {
		return res, err
	}
	if again != testsPass {
		res.Baseline = testsUnreliable
	}
	return res, nil
}

// extractTar writes git archive's tar into dir: regular files, directories and symlinks only.
func extractTar(data []byte, dir string) error {
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.FromSlash(h.Name)
		if !filepath.IsLocal(name) {
			continue
		}
		target := filepath.Join(dir, name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(h.Mode)&0o755|0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return errors.Join(err, f.Close())
			}
			if err := f.Close(); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(h.Linkname, target); err != nil {
				return err
			}
		}
	}
}

// copyTree copies a directory of regular files, directories and symlinks, keeping file modes.
func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, 0o755)
		case info.Mode()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		case info.Mode().IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode().Perm())
		}
		return nil
	})
}

func caughtWord(run string) string {
	if run == testsPass {
		return testsNotCaught
	}
	return testsCaught
}

// draftTests runs commitTests for draft and seal, first saying that it runs the commit's code.
func draftTests(ctx context.Context, root, sha string, w io.Writer) (testsResult, error) {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	if cfg, found, err := readTestsConfig(ctx, root, parent); found && err == nil && len(commitClaims(ctx, root, sha)) > 0 {
		if _, err := fmt.Fprintf(w, "tests: running `%s` with each hunk of code undone (this runs the commit's own code)\n", printableUnit(cfg.Run)); err != nil {
			return testsResult{}, err
		}
	}
	return commitTests(ctx, root, sha)
}

// writeTestsLines prints what review draft found.
func writeTestsLines(w io.Writer, res testsResult) error {
	if !res.Declared {
		return nil
	}
	p := func(format string, a ...any) error { _, err := fmt.Fprintf(w, format, a...); return err }
	switch {
	case res.Invalid != "":
		return p("tests: %s; seal will refuse until it is fixed\n", printableUnit(res.Invalid))
	case res.Baseline == testsUnreliable:
		return p("tests: `%s` passed, then failed when run again unchanged after the undo runs: the results are unreliable; seal will record NEEDS_WORK\n", printableUnit(res.Run))
	case res.Baseline != testsPass:
		return p("tests: `%s` does not pass at this commit (%s); seal will record NEEDS_WORK\n", printableUnit(res.Run), res.Baseline)
	}
	if len(res.Hunks) == 0 {
		return p("tests: the commit changes no code outside the test paths\n")
	}
	for _, h := range res.Hunks {
		var err error
		switch h.Result {
		case testsCaught:
			err = p("tests caught undoing %s %s\n", printableUnit(h.Path), h.Lines)
		case testsNotCaught:
			err = p("tests did NOT catch undoing %s %s\n", printableUnit(h.Path), h.Lines)
		case testsUndoFailed:
			err = p("tests: could not undo %s %s\n", printableUnit(h.Path), h.Lines)
		default:
			err = p("tests not run for %s %s\n", printableUnit(h.Path), h.Lines)
		}
		if err != nil {
			return err
		}
	}
	if err := p("tests %s undoing the whole change\n", map[string]string{testsCaught: "caught", testsNotCaught: "did NOT catch", testsNotRun: "not run for"}[res.Whole]); err != nil {
		return err
	}
	for _, h := range res.Hunks {
		if h.Result != testsCaught {
			return p("decide each hunk not caught in tests.hunks: \"accept\" with a reason, or \"reject\"\n")
		}
	}
	return nil
}

// decideTests matches the reviewer's decisions to a fresh run (spec §3): every hunk that was not
// caught (not caught, not run, undo failed) needs accept with a reason, or reject (review F2).
// needsWork is true on any reject, a whole change not caught, or tests that do not reliably pass.
func decideTests(res testsResult, edited *testsResult) (testsResult, bool, error) {
	if !res.Declared {
		return res, false, nil
	}
	if res.Invalid != "" {
		return res, false, fmt.Errorf("seal: %s; fix it before sealing", res.Invalid)
	}
	given := map[string]testsHunk{}
	if edited != nil {
		for _, h := range edited.Hunks {
			given[areaHunk{Path: h.Path, Lines: h.Lines}.key()] = h
		}
	}
	needsWork := res.Baseline != testsPass || len(res.Hunks) > 0 && res.Whole != testsCaught
	var undecided []string
	for i := range res.Hunks {
		h := &res.Hunks[i]
		if h.Result == testsCaught {
			continue
		}
		d, ok := given[areaHunk{Path: h.Path, Lines: h.Lines}.key()]
		h.Decision, h.Reason = d.Decision, strings.TrimSpace(d.Reason)
		switch {
		case ok && h.Decision == "reject":
			needsWork = true
		case ok && h.Decision == "accept" && checkReason(h.Reason) == nil:
		default:
			undecided = append(undecided, printableUnit(h.Path)+" "+h.Lines)
		}
	}
	if len(undecided) > 0 {
		return res, false, fmt.Errorf("seal: hunks the tests did not catch are not decided; set decision to accept (with a reason) or reject in tests.hunks, or run `proofbound review draft` again: %s", strings.Join(undecided, "; "))
	}
	return res, needsWork, nil
}

func testsRecordPath(sha string) string { return testsRecordPrefix + sha[:12] + ".json" }

// testsRecordBytes builds the TESTS- record, signed like the area record.
func testsRecordBytes(sha, by string, res testsResult, key ed25519.PrivateKey, at string) ([]byte, error) {
	hunks, err := canonicalJSON(res.Hunks)
	if err != nil {
		return nil, err
	}
	fields := map[string]any{"schema": testsSchema, "commit": sha, "by": by, "at": at, "run": res.Run, "baseline": res.Baseline, "whole": res.Whole, "hunks": string(hunks), "key": ""}
	if key == nil {
		out, err := canonicalJSON(fields)
		return append(out, '\n'), err
	}
	fields["key"] = pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey))
	return signRecord(fields, key)
}

var errNoTestsRecord = errors.New("no tests record")

// readTestsRecord verifies HEAD's TESTS- record for sha against HEAD's declared keys.
func readTestsRecord(ctx context.Context, root, sha string) (map[string]string, []testsHunk, error) {
	raw, err := gitBytes(ctx, root, "show", "HEAD:"+testsRecordPath(sha))
	if err != nil {
		return nil, nil, errNoTestsRecord
	}
	fields, err := verifyRecord(raw, testsSchema, declaredKeys(ctx, root, "HEAD"))
	if err != nil {
		return nil, nil, fmt.Errorf("%s does not verify: %v", testsRecordPath(sha), err)
	}
	if fields["commit"] != sha {
		return nil, nil, fmt.Errorf("%s judges another commit", testsRecordPath(sha))
	}
	var hunks []testsHunk
	dec := json.NewDecoder(strings.NewReader(fields["hunks"]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&hunks); err != nil {
		return nil, nil, fmt.Errorf("%s: unreadable hunks", testsRecordPath(sha))
	}
	return fields, hunks, nil
}

// testsControl is audit control C11 (spec §3). It runs nothing: it checks each claiming commit's
// signed record against the hunks git lists now and the command declared in the commit's parent.
func testsControl(ctx context.Context, root, base string, rep *auditReport, except func(kind, control, commit, text string)) (auditControl, error) {
	c := auditControl{ID: "C11", Question: controlQuestions["C11"], Command: "proofbound review draft <commit>",
		Basis: "A project may declare its test command in .proofbound/tests.json, read from each commit's parent. For every claimed commit the reviewer's signed record must show the tests pass, fail when the whole change is undone, and fail when each hunk of code is undone, or a signed reason for each that does not. Undoing a hunk is not full mutation testing: a test can notice a removed hunk and still miss a subtle bug in it, and a test that only compares file contents notices every hunk, so the reviewer still reads the tests. Projects that declare no tests are counted, not flagged."}
	owners := map[string][]string{}
	for _, r := range rep.Roles {
		owners[r.Commit] = r.Owners
	}
	agents := map[string]bool{}
	for _, cm := range rep.Commits {
		agents[cm.Commit] = len(cm.Agents) > 0
	}
	declared, good, missing := 0, 0, 0
	for _, cm := range rep.Commits {
		if cm.Class != classClaimed || len(commitClaims(ctx, root, cm.Commit)) == 0 {
			continue
		}
		subj := cm.Subject
		fail := func(text string) {
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: text})
			except("tests", "C11", cm.Commit, fmt.Sprintf("%q: %s", subj, text))
		}
		if setter := areaSetter(ctx, root, base, cm.Commit, []string{testsConfigPath}, agents); setter != "" {
			declared++
			fail("its test declaration was last changed by " + setter[:12] + ", a commit made by an agent")
			continue
		}
		parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", cm.Commit+"^1"))
		cfg, found, err := readTestsConfig(ctx, root, parent)
		if !found {
			missing++
			continue
		}
		declared++
		if err != nil {
			fail(printableUnit(err.Error()))
			continue
		}
		units, _, err := testsUnits(ctx, root, cm.Commit, cfg)
		if err != nil {
			return c, err
		}
		fields, hunks, err := readTestsRecord(ctx, root, cm.Commit)
		switch {
		case errors.Is(err, errNoTestsRecord):
			fail("no reviewer recorded whether its tests catch the change")
			continue
		case err != nil:
			fail(printableUnit(err.Error()))
			continue
		}
		if ok, note := independence(cm, []string{fields["by"]}, owners[cm.Commit]); !ok {
			fail("the tests record is not independent: " + note)
			continue
		}
		if fields["run"] != cfg.Run {
			fail("the tests record ran a different command from the one declared")
			continue
		}
		recorded := map[string]testsHunk{}
		for _, h := range hunks {
			recorded[areaHunk{Path: h.Path, Lines: h.Lines}.key()] = h
		}
		// What this control trusted, for the review outcomes by agent (CLI-INV-73).
		tc := rep.trackFor(cm.Commit)
		tc.TestParts = len(units)
		tc.WholeNotCaught = len(units) > 0 && fields["whole"] != testsCaught
		for _, u := range units {
			h, ok := recorded[areaHunk{Path: u.Path, Lines: u.Lines}.key()]
			if !ok || h.Result == testsCaught {
				continue
			}
			tc.TestsNotCaught++
			switch {
			case h.Decision == "accept" && checkReason(h.Reason) == nil:
				tc.TestsAccepted++
			case h.Decision == "reject":
				tc.TestsRejected++
			}
		}
		var open []string
		for _, u := range units {
			h, ok := recorded[areaHunk{Path: u.Path, Lines: u.Lines}.key()]
			switch {
			case !ok:
				open = append(open, printableUnit(u.Path)+" "+u.Lines+" not in the record")
			case h.Result == testsCaught:
			case h.Decision == "accept" && checkReason(h.Reason) == nil:
			default:
				open = append(open, printableUnit(u.Path)+" "+u.Lines+" "+h.Result)
			}
		}
		if fields["baseline"] != testsPass {
			open = append([]string{"the tests do not pass at the commit"}, open...)
		}
		if len(units) > 0 && fields["whole"] != testsCaught {
			open = append([]string{"undoing the whole change: " + fields["whole"]}, open...)
		}
		if len(open) > 0 {
			fail("tests do not catch the change: " + strings.Join(shortList(open, 10), ", "))
			continue
		}
		good++
		c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: fmt.Sprintf("tests caught undoing each of %d hunk(s) and the whole change, recorded by %s", len(units), fields["by"])})
	}
	switch {
	case declared == 0:
		c.State, c.Count = auditNone, fmt.Sprintf("no claimed change declares tests; %d not declared", missing)
	case good == declared:
		c.State, c.Count = auditHeld, fmt.Sprintf("%d of %d declared; %d not declared", good, declared, missing)
	default:
		c.State, c.Count = auditExcepted, fmt.Sprintf("%d of %d declared; %d not declared", good, declared, missing)
	}
	return c, nil
}

// testsDeclared is whether the commit's parent declares tests (the Review bench refuses to seal then).
func testsDeclared(ctx context.Context, root, sha string) bool {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	_, found, _ := readTestsConfig(ctx, root, parent)
	return found
}
