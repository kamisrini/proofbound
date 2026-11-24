package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// Session-to-line provenance (docs/plans/session-provenance-spec.md, CLI-INV-70..72). A Claude Code
// PostToolUse hook records a hash of each line an agent's edit added; review matches every line a
// commit adds against those records and the reviewer signs what was found. It is not proof against a
// hostile agent, which runs on the same machine: it is the reviewer's signed statement of what the
// logs on their machine showed.

const (
	sessionEditSchema = "proofbound.session-edit.v1"
	provSchema        = "proofbound.provenance-review.v1"
	provRecordPrefix  = "docs/verification/verdicts/PROV-"
	// provRecordMax bounds one appended record, so records from parallel writers never interleave.
	provRecordMax = 3 << 10
)

var sessionIDRE = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9._-]{0,127}$`)

type postToolCall struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath  string `json:"file_path"`
		Content   string `json:"content"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
		Edits     []struct {
			OldString string `json:"old_string"`
			NewString string `json:"new_string"`
		} `json:"edits"`
	} `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

type toolResponse struct {
	Type            string `json:"type"`
	Success         *bool  `json:"success"`
	Error           any    `json:"error"`
	StructuredPatch []struct {
		Lines []string `json:"lines"`
	} `json:"structuredPatch"`
}

type sessionEdit struct {
	Schema  string   `json:"schema"`
	Session string   `json:"session"`
	At      string   `json:"at"`
	Tool    string   `json:"tool"`
	Path    string   `json:"path"`
	Task    string   `json:"task,omitempty"`
	Whole   bool     `json:"whole_file,omitempty"`
	Lines   []string `json:"lines"`
}

// provLine normalises a line before hashing: a trailing \r and trailing whitespace do not count.
func provLine(s string) string { return strings.TrimRight(s, " \t\r") }

// provHash is the hash a log records for one normalised line.
func provHash(s string) string {
	sum := sha256.Sum256([]byte(provLine(s)))
	return hex.EncodeToString(sum[:16])
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// addedLines is what the tool added, never the context it repeated (spec §1.1): the `+` lines of the
// structured patch Claude Code returns; else, for Edit and MultiEdit, new minus old as a multiset;
// else, for Write, every line, marked as a whole-file write. A new file's lines are all added.
func addedLines(call postToolCall, resp toolResponse) (lines []string, whole bool) {
	var plus []string
	for _, h := range resp.StructuredPatch {
		for _, l := range h.Lines {
			if strings.HasPrefix(l, "+") {
				plus = append(plus, l[1:])
			}
		}
	}
	if len(plus) > 0 {
		return plus, false
	}
	switch call.ToolName {
	case "Write":
		return splitLines(call.ToolInput.Content), resp.Type != "create"
	case "Edit":
		return multisetMinus(splitLines(call.ToolInput.NewString), splitLines(call.ToolInput.OldString)), false
	case "MultiEdit":
		var out []string
		for _, e := range call.ToolInput.Edits {
			out = append(out, multisetMinus(splitLines(e.NewString), splitLines(e.OldString))...)
		}
		return out, false
	}
	return nil, false
}

func multisetMinus(a, b []string) []string {
	n := map[string]int{}
	for _, l := range b {
		n[provLine(l)]++
	}
	var out []string
	for _, l := range a {
		if k := provLine(l); n[k] > 0 {
			n[k]--
			continue
		}
		out = append(out, l)
	}
	return out
}

// fileRepo finds the repository holding target (from the file, not the session's cwd: perimeter
// review F2) and target's slash path inside it.
func fileRepo(ctx context.Context, cwd, target string) (root, rel string, ok bool) {
	if target == "" || cwd == "" {
		return "", "", false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	target = resolvePath(target)
	root = strings.TrimSpace(gitOutput(ctx, existingAncestor(target), "rev-parse", "--show-toplevel"))
	if root == "" {
		return "", "", false
	}
	r, err := filepath.Rel(resolvePath(root), target)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", "", false
	}
	return root, filepath.ToSlash(r), true
}

// sessionsDir is where the session logs live: the git common folder, shared by every worktree and
// never tracked (spec §1.1, review F6).
func sessionsDir(ctx context.Context, root string) string {
	common := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	if common == "" {
		return ""
	}
	return filepath.Join(common, "proofbound", "sessions")
}

// provNow is the hook's clock (Claude Code's input carries no time; a variable for tests).
var provNow = func() time.Time { return time.Now().UTC() }

// recordSessionEdit appends the call's added-line hashes to its session log. Anything it cannot read
// or place is skipped: the hook never blocks or fails an edit.
func recordSessionEdit(ctx context.Context, raw []byte) error {
	var call postToolCall
	if err := json.Unmarshal(raw, &call); err != nil {
		return nil
	}
	switch call.ToolName {
	case "Edit", "Write", "MultiEdit":
	default:
		return nil
	}
	var resp toolResponse
	if len(call.ToolResponse) > 0 {
		_ = json.Unmarshal(call.ToolResponse, &resp)
	}
	// Claude Code sends PostToolUse only for a call that succeeded; this is a defensive check.
	if (resp.Success != nil && !*resp.Success) || (resp.Error != nil && resp.Error != false && resp.Error != "") {
		return nil
	}
	if !sessionIDRE.MatchString(call.SessionID) {
		return nil
	}
	root, rel, ok := fileRepo(ctx, call.Cwd, call.ToolInput.FilePath)
	if !ok || housekeepingPath(rel) {
		return nil
	}
	lines, whole := addedLines(call, resp)
	var hashes []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			hashes = append(hashes, provHash(l))
		}
	}
	if len(hashes) == 0 {
		return nil
	}
	dir := sessionsDir(ctx, root)
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	task, _ := currentTask(root)
	rec := sessionEdit{Schema: sessionEditSchema, Session: call.SessionID, At: provNow().Format(time.RFC3339), Tool: call.ToolName, Path: rel, Task: task, Whole: whole, Lines: hashes[:1]}
	if out, _ := json.Marshal(rec); len(out)+1 > provRecordMax {
		return nil // a path too long for one record: skip rather than interleave
	}
	logPath := filepath.Join(dir, call.SessionID+".jsonl")
	if info, err := os.Lstat(logPath); err == nil && !info.Mode().IsRegular() {
		return nil // never follow a link planted at the log path (review F9)
	}
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for len(hashes) > 0 {
		rec.Lines = nil
		for len(hashes) > 0 {
			rec.Lines = append(rec.Lines, hashes[0])
			if out, _ := json.Marshal(rec); len(out)+1 > provRecordMax {
				rec.Lines = rec.Lines[:len(rec.Lines)-1]
				break
			}
			hashes = hashes[1:]
		}
		out, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(out, '\n')); err != nil {
			return err
		}
	}
	return nil
}

// runPostToolHook is `proofbound hook post-tool-use`: it always exits 0 and prints nothing.
func runPostToolHook(ctx context.Context) int {
	raw, err := io.ReadAll(io.LimitReader(hookInput(), 8<<20))
	if err == nil {
		_ = recordSessionEdit(ctx, raw)
	}
	return 0
}

// The match (spec §1.2).

type provSession struct {
	Session string `json:"session"`
	Lines   int    `json:"lines"`
	// Whole is true when every match for this session came from a whole-file write.
	Whole bool `json:"whole_file,omitempty"`
}

type provFile struct {
	Path       string        `json:"path"`
	Sessions   []provSession `json:"sessions"`
	Unrecorded []int         `json:"unrecorded"`
	Common     int           `json:"common"`
}

type provLog struct {
	Session string `json:"session"`
	SHA256  string `json:"sha256"`
}

type provResult struct {
	Files    []provFile `json:"files"`
	Warnings []string   `json:"warnings"`
	Logs     []provLog  `json:"logs"`
}

const (
	provWarnNoLogs     = "no session logs here, so nothing can be traced (they are local to the machine the agent ran on)"
	provWarnNoHook     = "the edit-recording hook was probably not running for this commit"
	provWarnNoSettings = ".claude/settings.json in this checkout has no `proofbound hook post-tool-use` hook, so edits here are not being recorded"
)

type addedLine struct {
	n    int
	text string
}

// commitAddedLines are the lines sha adds to each traced path (not housekeeping), by new line number.
// A `+++ ` line is a file header only outside a hunk, whose length the `@@` header gives (review F3);
// paths are read with core.quotePath off, a trailing tab removed, and C-quoting undone (review F2).
func commitAddedLines(ctx context.Context, root, sha string) (map[string][]addedLine, []string, error) {
	if strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^{commit}")) == "" {
		// Not in this repository (as commitArea, which then has no claims): nothing to match.
		return map[string][]addedLine{}, nil, nil
	}
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	if parent == "" {
		parent = emptyTree
	}
	out, err := gitBytes(ctx, root, "-c", "core.quotePath=false", "diff", "--no-color", "--no-ext-diff", "--no-renames", "--unified=0", parent, sha, "--")
	if err != nil {
		return nil, nil, fmt.Errorf("provenance: cannot read the change of %s: %v", sha[:12], err)
	}
	files := map[string][]addedLine{}
	var order []string
	path, n, oldLeft, newLeft := "", 0, 0, 0
	for _, l := range strings.Split(string(out), "\n") {
		if oldLeft > 0 || newLeft > 0 {
			switch {
			case strings.HasPrefix(l, "+"):
				newLeft--
				if path != "" {
					if _, ok := files[path]; !ok {
						order = append(order, path)
					}
					files[path] = append(files[path], addedLine{n: n, text: l[1:]})
				}
				n++
			case strings.HasPrefix(l, "-"):
				oldLeft--
			case strings.HasPrefix(l, `\`):
			default:
				oldLeft--
				newLeft--
				n++
			}
			continue
		}
		switch {
		case strings.HasPrefix(l, "diff --git "):
			path = ""
		case strings.HasPrefix(l, "+++ "):
			path = diffPath(strings.TrimPrefix(l, "+++ "))
			if housekeepingPath(path) {
				path = ""
			}
		case strings.HasPrefix(l, "@@ "):
			if f := strings.Fields(l); len(f) >= 3 {
				_, oc, _ := hunkRange(f[1], "-")
				start, nc, _ := hunkRange(f[2], "+")
				n, oldLeft, newLeft = start, oc, nc
			}
		}
	}
	return files, order, nil
}

// diffPath is the repository path of a `+++` header's target, or "" for /dev/null.
func diffPath(p string) string {
	p = strings.TrimSuffix(p, "\t")
	if strings.HasPrefix(p, `"`) {
		u, err := strconv.Unquote(p)
		if err != nil {
			return ""
		}
		p = u
	}
	if !strings.HasPrefix(p, "b/") {
		return ""
	}
	return strings.TrimPrefix(p, "b/")
}

// commonLine is a line too common to be evidence: three or fewer non-space characters, or text the
// commit adds more than once to the same file (spec §1.2, review F4).
func commonLine(text string, seen map[string]int) bool {
	if len(strings.Join(strings.Fields(text), "")) <= 3 {
		return true
	}
	return seen[provLine(text)] > 1
}

type logIndex struct {
	// byPath maps a path and a line hash to the sessions that recorded it, and whether only by a
	// whole-file write.
	byPath map[string]map[string]map[string]bool
	logs   []provLog
	// problems are logs that could not be read: a warning, never a failed review (review F5).
	problems []string
}

func readSessionLogs(ctx context.Context, root string) (logIndex, error) {
	idx := logIndex{byPath: map[string]map[string]map[string]bool{}}
	dir := sessionsDir(ctx, root)
	if dir == "" {
		return idx, nil
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return idx, nil
	}
	if err != nil {
		idx.problems = append(idx.problems, "the session log folder cannot be read")
		return idx, nil
	}
	for _, e := range entries {
		name := e.Name()
		if !e.Type().IsRegular() || !strings.HasSuffix(name, ".jsonl") || !sessionIDRE.MatchString(strings.TrimSuffix(name, ".jsonl")) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			idx.problems = append(idx.problems, "the session log "+name+" cannot be read")
			continue
		}
		sum := sha256.Sum256(data)
		session := strings.TrimSuffix(name, ".jsonl")
		idx.logs = append(idx.logs, provLog{Session: session, SHA256: hex.EncodeToString(sum[:])})
		for _, line := range bytes.Split(data, []byte("\n")) {
			var rec sessionEdit
			if json.Unmarshal(line, &rec) != nil || rec.Schema != sessionEditSchema || rec.Session != session {
				continue
			}
			byHash := idx.byPath[rec.Path]
			if byHash == nil {
				byHash = map[string]map[string]bool{}
				idx.byPath[rec.Path] = byHash
			}
			for _, h := range rec.Lines {
				if byHash[h] == nil {
					byHash[h] = map[string]bool{}
				}
				// true means "only whole-file writes so far"; a precise record clears it.
				if prev, ok := byHash[h][session]; !ok || prev {
					byHash[h][session] = rec.Whole
				}
			}
		}
	}
	sort.Slice(idx.logs, func(a, b int) bool { return idx.logs[a].Session < idx.logs[b].Session })
	return idx, nil
}

// commitProvenance matches every line sha adds against the session logs on this machine.
func commitProvenance(ctx context.Context, root, sha string) (provResult, error) {
	res := provResult{Files: []provFile{}, Warnings: []string{}, Logs: []provLog{}}
	files, order, err := commitAddedLines(ctx, root, sha)
	if err != nil {
		return res, err
	}
	idx, err := readSessionLogs(ctx, root)
	if err != nil {
		return res, fmt.Errorf("provenance: cannot read the session logs: %v", err)
	}
	if idx.logs != nil {
		res.Logs = idx.logs
	}
	countable, traced := 0, 0
	for _, path := range order {
		f := provFile{Path: path, Sessions: []provSession{}, Unrecorded: []int{}}
		seen := map[string]int{}
		for _, l := range files[path] {
			seen[provLine(l.text)]++
		}
		counts := map[string]*provSession{}
		for _, l := range files[path] {
			if commonLine(l.text, seen) {
				f.Common++
				continue
			}
			countable++
			sessions := idx.byPath[path][provHash(l.text)]
			if len(sessions) == 0 {
				f.Unrecorded = append(f.Unrecorded, l.n)
				continue
			}
			traced++
			for s, whole := range sessions {
				c := counts[s]
				if c == nil {
					c = &provSession{Session: s, Whole: true}
					counts[s] = c
				}
				c.Lines++
				c.Whole = c.Whole && whole
			}
		}
		for _, c := range counts {
			f.Sessions = append(f.Sessions, *c)
		}
		sort.Slice(f.Sessions, func(a, b int) bool { return f.Sessions[a].Session < f.Sessions[b].Session })
		res.Files = append(res.Files, f)
	}
	res.Warnings = append(res.Warnings, idx.problems...)
	if len(idx.logs) == 0 {
		res.Warnings = append(res.Warnings, provWarnNoLogs)
	}
	if countable > 0 && traced == 0 && commitHasAgent(ctx, root, sha) && len(idx.logs) > 0 {
		res.Warnings = append(res.Warnings, provWarnNoHook)
	}
	if !settingsRecordEdits(root) {
		res.Warnings = append(res.Warnings, provWarnNoSettings)
	}
	return res, nil
}

// commitHasAgent is whether the commit's author or a Co-Authored-By names a coding agent.
func commitHasAgent(ctx context.Context, root, sha string) bool {
	out := gitOutput(ctx, root, "log", "-1", "--format=%an %ae%x00%(trailers:key=Co-authored-by,valueonly,separator=%x01)", sha)
	author, cos, _ := strings.Cut(out, "\x00")
	if agentName(author) {
		return true
	}
	for _, co := range strings.Split(cos, "\x01") {
		if co = strings.TrimSpace(co); co != "" && agentName(co) {
			return true
		}
	}
	return false
}

// settingsRecordEdits is whether the working tree's .claude/settings.json runs the recording hook.
func settingsRecordEdits(root string) bool {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(claudeSettingsPath)))
	return err == nil && bytes.Contains(data, []byte("proofbound hook post-tool-use"))
}

func shortSession(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// writeProvLines prints the draft's provenance lines (spec §1.2).
func writeProvLines(w io.Writer, res provResult) error {
	for _, f := range res.Files {
		var by []string
		total := 0
		for _, s := range f.Sessions {
			note := ""
			if s.Whole {
				note = " via a whole-file write"
			}
			by = append(by, fmt.Sprintf("%d by Claude Code session %s%s", s.Lines, shortSession(s.Session), note))
			total += s.Lines
		}
		if len(by) == 0 {
			by = []string{"0 by a session"}
		}
		un := fmt.Sprintf("%d with no recorded agent edit", len(f.Unrecorded))
		if len(f.Unrecorded) > 0 {
			un += " (lines " + lineList(f.Unrecorded, 10) + ")"
		}
		if _, err := fmt.Fprintf(w, "provenance %s: %s, %s, %d common\n", printableUnit(f.Path), strings.Join(by, ", "), un, f.Common); err != nil {
			return err
		}
	}
	for _, warn := range res.Warnings {
		if _, err := fmt.Fprintf(w, "provenance warning: %s\n", warn); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "provenance: docs, Markdown and records are not traced; lines moved by a shell or rewritten by a formatter show as no recorded agent edit")
	return err
}

func lineList(ns []int, max int) string {
	var s []string
	for i, n := range ns {
		if i == max {
			s = append(s, fmt.Sprintf("and %d more", len(ns)-max))
			break
		}
		s = append(s, strconv.Itoa(n))
	}
	return strings.Join(s, ", ")
}

func provRecordPath(sha string) string { return provRecordPrefix + sha[:12] + ".json" }

// provRecordBytes builds the PROV- record, signed like the area and tests records.
func provRecordBytes(sha, by string, res provResult, key ed25519.PrivateKey, at string) ([]byte, error) {
	files, err := canonicalJSON(res.Files)
	if err != nil {
		return nil, err
	}
	warnings, err := canonicalJSON(res.Warnings)
	if err != nil {
		return nil, err
	}
	logs, err := canonicalJSON(res.Logs)
	if err != nil {
		return nil, err
	}
	fields := map[string]any{"schema": provSchema, "commit": sha, "by": by, "at": at, "files": string(files), "warnings": string(warnings), "logs": string(logs), "key": ""}
	if key == nil {
		out, err := canonicalJSON(fields)
		return append(out, '\n'), err
	}
	fields["key"] = pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey))
	return signRecord(fields, key)
}

var errNoProvRecord = errors.New("no provenance record")

// readProvRecord verifies HEAD's PROV- record for sha against HEAD's declared keys.
func readProvRecord(ctx context.Context, root, sha string) (map[string]string, provResult, error) {
	var res provResult
	raw, err := gitBytes(ctx, root, "show", "HEAD:"+provRecordPath(sha))
	if err != nil {
		return nil, res, errNoProvRecord
	}
	fields, err := verifyRecord(raw, provSchema, declaredKeys(ctx, root, "HEAD"))
	if err != nil {
		return nil, res, fmt.Errorf("%s does not verify: %v", provRecordPath(sha), err)
	}
	if fields["commit"] != sha {
		return nil, res, fmt.Errorf("%s judges another commit", provRecordPath(sha))
	}
	for name, dst := range map[string]any{"files": &res.Files, "warnings": &res.Warnings, "logs": &res.Logs} {
		dec := json.NewDecoder(strings.NewReader(fields[name]))
		dec.DisallowUnknownFields()
		if err := dec.Decode(dst); err != nil {
			return nil, res, fmt.Errorf("%s: unreadable %s", provRecordPath(sha), name)
		}
	}
	return fields, res, nil
}

// provControl is audit control C12 (spec §1.4): a disclosure (LISTED). It is an exception only when a record
// does not verify or was signed by the commit's own author, agent or spec owner; the counts never are.
func provControl(ctx context.Context, root string, rep *auditReport, except func(kind, control, commit, text string)) (auditControl, error) {
	c := auditControl{ID: "C12", Question: controlQuestions["C12"], Command: "proofbound review draft <commit>",
		Basis: "A Claude Code hook records a hash of each line an agent's edit adds, in a log on the agent's machine. At review, every line a commit adds is matched against those logs, and the reviewer signs what was found: the lines each session recorded, and the lines no recorded agent edit wrote (a person, a shell command, a formatter, or an agent whose hook was not running). The agent runs on the same machine and could write to its log, so this is the reviewer's signed statement of what the logs showed, not proof against a hostile agent. Short and repeated lines are not counted, and a merge commit is compared with its first parent. Commits without a record are counted as not traced, not flagged."}
	owners := map[string][]string{}
	for _, r := range rep.Roles {
		owners[r.Commit] = r.Owners
	}
	records, good, untraced := 0, 0, 0
	for _, cm := range rep.Commits {
		if cm.Class != classClaimed || len(commitClaims(ctx, root, cm.Commit)) == 0 {
			continue
		}
		subj := cm.Subject
		fail := func(text string) {
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: text})
			except("provenance", "C12", cm.Commit, fmt.Sprintf("%q: %s", subj, text))
		}
		fields, res, err := readProvRecord(ctx, root, cm.Commit)
		switch {
		case errors.Is(err, errNoProvRecord):
			untraced++
			continue
		case err != nil:
			records++
			fail(printableUnit(err.Error()))
			continue
		}
		records++
		if ok, note := independence(cm, []string{fields["by"]}, owners[cm.Commit]); !ok {
			fail("the provenance record is not independent: " + note)
			continue
		}
		good++
		// What this control trusted, for the review outcomes by agent (CLI-INV-73).
		tc := rep.trackFor(cm.Commit)
		tc.Traced = true
		added, _, err := commitAddedLines(ctx, root, cm.Commit)
		if err != nil {
			return c, err
		}
		for _, f := range res.Files {
			for _, s := range f.Sessions {
				tc.Sessions[s.Session] += s.Lines
			}
			tc.LinesUnrecorded += len(f.Unrecorded)
			if n := len(added[f.Path]) - f.Common - len(f.Unrecorded); n > 0 {
				tc.LinesTraced += n
			}
		}
		c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: provSummary(res) + ", recorded by " + fields["by"]})
	}
	switch {
	case records == 0:
		c.State, c.Count = auditNone, fmt.Sprintf("no claimed change has a provenance record; %d not traced", untraced)
	case good == records:
		// Listed, never held: it discloses what the logs showed for the traced changes, and says how
		// many were not traced (review F4).
		c.State, c.Count = auditListed, fmt.Sprintf("%d traced; %d not traced", good, untraced)
	default:
		c.State, c.Count = auditExcepted, fmt.Sprintf("%d of %d records verify; %d not traced", good, records, untraced)
	}
	return c, nil
}

// provSummary is one line for the audit room: lines per session, unrecorded lines, warnings.
func provSummary(res provResult) string {
	per := map[string]int{}
	unrec, common := 0, 0
	var where []string
	for _, f := range res.Files {
		for _, s := range f.Sessions {
			per[s.Session] += s.Lines
		}
		unrec += len(f.Unrecorded)
		common += f.Common
		if len(f.Unrecorded) > 0 {
			where = append(where, fmt.Sprintf("%s (%d)", printableUnit(f.Path), len(f.Unrecorded)))
		}
	}
	var sessions []string
	for s := range per {
		sessions = append(sessions, s)
	}
	sort.Strings(sessions)
	var by []string
	for _, s := range sessions {
		by = append(by, fmt.Sprintf("%d line(s) by session %s", per[s], shortSession(s)))
	}
	if len(by) == 0 {
		by = []string{"no line by a recorded session"}
	}
	out := strings.Join(by, ", ") + fmt.Sprintf("; %d with no recorded agent edit", unrec)
	if len(where) > 0 {
		out += " in " + strings.Join(shortList(where, 5), ", ")
	}
	out += fmt.Sprintf("; %d common", common)
	for _, w := range res.Warnings {
		out += "; warning: " + w
	}
	return out
}
