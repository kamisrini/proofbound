package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The declared change area (docs/plans/change-area-spec.md, CLI-INV-52..55): a spec folder's optional
// `scope.paths` names the files its change may touch; the lines a claiming commit changes outside it
// are listed, must be decided at seal, and are shown by audit control C10.

const (
	areaSchema       = "proofbound.area-review.v1"
	areaFileName     = "scope.paths"
	areaMaxPerFile   = 50
	areaMaxTotal     = 200
	areaRecordPrefix = "docs/verification/verdicts/AREA-"
)

// areaHunk is one changed line range (or whole-file change) outside the area.
type areaHunk struct {
	Path  string `json:"path"`
	Lines string `json:"lines"`
}

// areaDecision is the reviewer's decision on one areaHunk: accept (with a reason) or reject.
type areaDecision struct {
	Path     string `json:"path"`
	Lines    string `json:"lines"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// areaResult is a claiming commit's area and what it changed outside it.
type areaResult struct {
	Names    []string   // the claimed change names
	Declared bool       // every claimed change has a scope.paths in the commit's first parent
	Missing  []string   // claimed changes without one
	Files    []string   // the scope.paths files read (first parent)
	Invalid  string     // why a scope.paths is invalid; then nothing else is computed
	Hunks    []areaHunk // outside the area, in path order
}

func (h areaHunk) key() string { return h.Path + "\x00" + h.Lines }

// parseArea reads a scope.paths file: one pattern per line, blank lines and `#` comments ignored. Any
// invalid line makes the whole file invalid (spec §1): an invalid area never allows everything.
func parseArea(data []byte) ([]string, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("not UTF-8")
	}
	var out []string
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if err := checkAreaPattern(trimmed); err != nil {
			return nil, fmt.Errorf("line %d: %v", i+1, err)
		}
		out = append(out, trimmed)
	}
	return out, nil
}

func checkAreaPattern(p string) error {
	for _, r := range p {
		if unicode.IsControl(r) {
			return errors.New("holds a control character")
		}
	}
	switch {
	case strings.HasPrefix(p, "/"):
		return errors.New("is absolute")
	case strings.Contains(p, `\`):
		return errors.New("holds a backslash")
	}
	for _, seg := range strings.Split(strings.TrimSuffix(p, "/"), "/") {
		switch {
		case seg == "":
			return errors.New("has an empty path segment")
		case seg == "." || seg == "..":
			return errors.New("holds a . or .. segment")
		case seg != "**" && strings.Contains(seg, "**"):
			return errors.New("uses ** inside a segment; ** must be a whole segment")
		}
	}
	return nil
}

// areaMatch reports whether repository path p is inside pattern (spec §1): `*` within one segment,
// `**` any number of whole segments, a trailing `/` everything under that directory.
func areaMatch(pattern, p string) bool {
	if dir, ok := strings.CutSuffix(pattern, "/"); ok {
		// At least one segment below the directory: `src/cart/` does not match a file `src/cart`.
		pattern = dir + "/*/**"
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

func matchSegments(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(segs); i++ {
			if matchSegments(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	return len(segs) > 0 && matchStar(pat[0], segs[0]) && matchSegments(pat[1:], segs[1:])
}

// matchStar matches one segment where `*` is any run of characters and everything else is literal.
func matchStar(pat, s string) bool {
	parts := strings.Split(pat, "*")
	if len(parts) == 1 {
		return pat == s
	}
	if !strings.HasPrefix(s, parts[0]) {
		return false
	}
	s = s[len(parts[0]):]
	for _, mid := range parts[1 : len(parts)-1] {
		i := strings.Index(s, mid)
		if i < 0 {
			return false
		}
		s = s[i+len(mid):]
	}
	return strings.HasSuffix(s, parts[len(parts)-1])
}

// specDirsAt are the spec folders of change name at rev, in the order C7 looks for them.
func specDirsAt(ctx context.Context, root, rev, name string) []string {
	var out []string
	for _, dir := range []string{"openspec/changes/" + name, "specs/" + name, "openspec/changes/archive/" + name} {
		if strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "-d", "--name-only", rev, "--", dir)) != "" {
			out = append(out, dir)
		}
	}
	for _, entry := range strings.Split(gitOutput(ctx, root, "ls-tree", "-d", "--name-only", rev+":openspec/changes/archive"), "\n") {
		// Exactly OpenSpec's dated form for this name (the sdd connector's datedRE): a suffix match
		// would take `2026-01-01-add-cart` as an archive of `cart` and widen its area (review F3).
		if m := archivedRE.FindStringSubmatch(strings.TrimSpace(entry)); m != nil && m[1] == name {
			out = append(out, "openspec/changes/archive/"+m[0])
		}
	}
	return out
}

var archivedRE = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$`)

// areaCandidates are every path where a claimed change's scope.paths could be at rev, whether or not
// it exists there now: a deleted area file is a changed area too (review F5).
func areaCandidates(ctx context.Context, root, rev string, names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, name := range names {
		dirs := []string{"openspec/changes/" + name, "specs/" + name, "openspec/changes/archive/" + name}
		if rev != "" {
			dirs = append(dirs, specDirsAt(ctx, root, rev, name)...)
		}
		for _, d := range dirs {
			if f := d + "/" + areaFileName; !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	sort.Strings(out)
	return out
}

// commitClaims are the change names a commit's message claims (`Intent: sdd:<name>`).
func commitClaims(ctx context.Context, root, sha string) []string {
	names, _ := claimLines(gitOutput(ctx, root, "log", "-1", "--format=%B", sha))
	var out []string
	for _, claim := range names {
		if name := strings.TrimPrefix(claim, "sdd:"); taskNameRE.MatchString(name) {
			out = append(out, name)
		}
	}
	return out
}

// commitArea computes a claiming commit's area from its first parent and lists what it changed
// outside it (spec §1, §2). It reads only git.
func commitArea(ctx context.Context, root, sha string) (areaResult, error) {
	res := areaResult{Names: commitClaims(ctx, root, sha), Hunks: []areaHunk{}}
	if len(res.Names) == 0 {
		return res, nil
	}
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	var patterns []string
	for _, name := range res.Names {
		found := false
		if len(parent) == 40 {
			for _, dir := range specDirsAt(ctx, root, parent, name) {
				file := dir + "/" + areaFileName
				raw, err := gitBytes(ctx, root, "show", parent+":"+file)
				if err != nil {
					continue
				}
				found = true
				res.Files = append(res.Files, file)
				ps, err := parseArea(raw)
				if err != nil {
					res.Invalid = fmt.Sprintf("%s is invalid: %v", file, err)
					return res, nil
				}
				patterns = append(patterns, ps...)
			}
		}
		if !found {
			res.Missing = append(res.Missing, name)
		}
	}
	if len(res.Missing) > 0 {
		return res, nil
	}
	res.Declared = true
	from := parent
	if len(from) != 40 {
		from = emptyTree
	}
	raw, err := gitBytes(ctx, root, "diff", "--no-renames", "--no-ext-diff", "--ignore-submodules=none", "--name-status", "-z", from, sha, "--")
	if err != nil {
		return res, fmt.Errorf("area: cannot list the paths of %s", sha)
	}
	parts := strings.Split(string(raw), "\x00")
	for i := 0; i+1 < len(parts); i += 2 {
		status, p := parts[i], parts[i+1]
		if status == "" {
			break
		}
		if housekeepingPath(p) || inArea(patterns, p) {
			continue
		}
		hunks, err := pathHunks(ctx, root, from, sha, p, status)
		if err != nil {
			return res, err
		}
		res.Hunks = append(res.Hunks, hunks...)
	}
	if len(res.Hunks) > areaMaxTotal {
		// One last entry stands for the rest and is decided like any other (review F4).
		more := len(res.Hunks) - (areaMaxTotal - 1)
		res.Hunks = append(res.Hunks[:areaMaxTotal-1], areaHunk{Path: "", Lines: fmt.Sprintf("and %d more hunks in other files", more)})
	}
	return res, nil
}

// emptyTree is git's empty tree, the "parent" of a root commit.
const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func inArea(patterns []string, p string) bool {
	for _, pat := range patterns {
		if areaMatch(pat, p) {
			return true
		}
	}
	return false
}

// pathHunks lists one path's changed line ranges (spec §2), at most areaMaxPerFile of them.
func pathHunks(ctx context.Context, root, from, sha, p, status string) ([]areaHunk, error) {
	if status == "D" {
		return []areaHunk{{Path: p, Lines: "file deleted"}}, nil
	}
	raw, err := gitBytes(ctx, root, "--literal-pathspecs", "diff", "-U0", "--no-renames", "--no-ext-diff", "--ignore-submodules=none", "--no-color", from, sha, "--", p)
	if err != nil {
		return nil, fmt.Errorf("area: cannot diff %s in %s", p, sha)
	}
	var out []areaHunk
	binary, more := false, 0
	for _, line := range bytes.Split(raw, []byte("\n")) {
		s := string(line)
		if strings.HasPrefix(s, "Binary files ") {
			binary = true
		}
		if !strings.HasPrefix(s, "@@ ") {
			continue
		}
		lines, ok := hunkLines(s)
		if !ok {
			return nil, fmt.Errorf("area: unreadable hunk header in %s: %q", p, s)
		}
		if len(out) == areaMaxPerFile {
			more++
			continue
		}
		out = append(out, areaHunk{Path: p, Lines: lines})
	}
	if more > 0 {
		out = append(out, areaHunk{Path: p, Lines: fmt.Sprintf("and %d more hunks", more)})
	}
	if len(out) == 0 {
		what := "mode change"
		if binary {
			what = "binary file"
		}
		out = append(out, areaHunk{Path: p, Lines: what})
	}
	return out, nil
}

// hunkLines turns `@@ -o[,oc] +n[,nc] @@` into "lines n-m", or "removed lines o-p" for a pure deletion.
func hunkLines(header string) (string, bool) {
	f := strings.Fields(header)
	if len(f) < 4 || f[3] != "@@" && !strings.HasPrefix(f[3], "@@") {
		return "", false
	}
	oldStart, oldCount, ok1 := hunkRange(f[1], "-")
	newStart, newCount, ok2 := hunkRange(f[2], "+")
	if !ok1 || !ok2 {
		return "", false
	}
	span := func(word string, start, count int) string {
		if count == 1 {
			return fmt.Sprintf("%s %d", word, start)
		}
		return fmt.Sprintf("%ss %d-%d", word, start, start+count-1)
	}
	if newCount > 0 {
		return span("line", newStart, newCount), true
	}
	if oldCount > 0 {
		return span("removed line", oldStart, oldCount), true
	}
	return "", false
}

func hunkRange(s, sign string) (start, count int, ok bool) {
	s, ok = strings.CutPrefix(s, sign)
	if !ok {
		return 0, 0, false
	}
	startText, countText, has := strings.Cut(s, ",")
	start, err := strconv.Atoi(startText)
	if err != nil {
		return 0, 0, false
	}
	count = 1
	if has {
		if count, err = strconv.Atoi(countText); err != nil {
			return 0, 0, false
		}
	}
	return start, count, true
}

// withArea adds the commit's out-of-area list to a ledger draft (draft and the Review bench).
func withArea(ctx context.Context, root string, d draftFile) (draftFile, areaResult, error) {
	area, err := commitArea(ctx, root, d.Commit)
	if err != nil {
		return d, area, err
	}
	d.OutOfArea = areaDraft(area)
	return d, area, nil
}

// writeAreaLines prints what review draft found about the area.
func writeAreaLines(w io.Writer, area areaResult) error {
	var err error
	switch {
	case len(area.Names) == 0:
	case area.Invalid != "":
		_, err = fmt.Fprintf(w, "area: %s; seal will refuse until it is fixed\n", printableUnit(area.Invalid))
	case !area.Declared:
		_, err = fmt.Fprintf(w, "area: not declared by %s\n", strings.Join(area.Missing, ", "))
	case len(area.Hunks) == 0:
		_, err = fmt.Fprintln(w, "area: every changed line is inside the declared area")
	default:
		for _, h := range area.Hunks {
			if _, err = fmt.Fprintf(w, "outside area %s %s\n", printableUnit(h.Path), h.Lines); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(w, "decide each in out_of_area: \"accept\" with a reason, or \"reject\"")
	}
	return err
}

// areaNote is the Review bench's one line about the area.
func areaNote(area areaResult) string {
	switch {
	case len(area.Names) == 0:
		return ""
	case area.Invalid != "":
		return printableUnit(area.Invalid) + "; sealing is refused until it is fixed"
	case !area.Declared:
		return "Area not declared: " + strings.Join(area.Missing, ", ") + " has no scope.paths, so no line is checked against an area."
	case len(area.Hunks) == 0:
		return "Every changed line is inside the area its spec declares."
	}
	return fmt.Sprintf("%d line range(s) are outside the area its spec declares: accept each with a reason, or reject it.", len(area.Hunks))
}

// areaDraft is the undecided list review draft writes into draft.json.
func areaDraft(res areaResult) []areaDecision {
	out := []areaDecision{}
	for _, h := range res.Hunks {
		out = append(out, areaDecision{Path: h.Path, Lines: h.Lines})
	}
	return out
}

// decideArea matches the reviewer's decisions to the hunks git lists (spec §3). Every hunk needs
// `accept` with a reason or `reject`; anything else refuses the seal. It returns the decisions in
// hunk order and whether any was a reject.
func decideArea(res areaResult, edited []areaDecision) ([]areaDecision, bool, error) {
	if res.Invalid != "" {
		return nil, false, fmt.Errorf("seal: %s; fix the area before sealing", res.Invalid)
	}
	given := map[string]areaDecision{}
	for _, d := range edited {
		given[areaHunk{Path: d.Path, Lines: d.Lines}.key()] = d
	}
	var out []areaDecision
	var undecided []string
	rejected := false
	for _, h := range res.Hunks {
		d, ok := given[h.key()]
		d.Path, d.Lines, d.Reason = h.Path, h.Lines, strings.TrimSpace(d.Reason)
		switch {
		case !ok:
			undecided = append(undecided, printableUnit(h.Path)+" "+h.Lines)
		case d.Decision == "reject":
			rejected = true
		case d.Decision == "accept":
			if err := checkReason(d.Reason); err != nil {
				undecided = append(undecided, printableUnit(h.Path)+" "+h.Lines+" (accept needs a one-line reason)")
			}
		default:
			undecided = append(undecided, printableUnit(h.Path)+" "+h.Lines)
		}
		out = append(out, d)
	}
	if len(undecided) > 0 {
		return nil, false, fmt.Errorf("seal: lines outside the spec's area are not decided; set decision to accept (with a reason) or reject in out_of_area, or run `proofbound review draft` again: %s", strings.Join(undecided, "; "))
	}
	return out, rejected, nil
}

// areaRecordPath is where seal writes a commit's area decisions.
func areaRecordPath(sha string) string { return areaRecordPrefix + sha[:12] + ".json" }

// areaRecordBytes builds the AREA- record: signed like an exception outcome when the reviewer holds a
// declared key, unsigned otherwise (then the audit does not count it). `hunks` is the canonical JSON
// of the decisions, so every field stays a string and verifyRecord applies unchanged.
func areaRecordBytes(sha, by string, decisions []areaDecision, key ed25519.PrivateKey, at string) ([]byte, error) {
	hunks, err := canonicalJSON(decisions)
	if err != nil {
		return nil, err
	}
	fields := map[string]any{"schema": areaSchema, "commit": sha, "by": by, "at": at, "hunks": string(hunks), "key": ""}
	if key == nil {
		out, err := canonicalJSON(fields)
		return append(out, '\n'), err
	}
	fields["key"] = pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey))
	return signRecord(fields, key)
}

// readAreaRecord verifies HEAD's AREA- record for sha against HEAD's declared keys and returns its
// signer's key name and decisions.
func readAreaRecord(ctx context.Context, root, sha string) (string, []areaDecision, error) {
	raw, err := gitBytes(ctx, root, "show", "HEAD:"+areaRecordPath(sha))
	if err != nil {
		return "", nil, errNoAreaRecord
	}
	fields, err := verifyRecord(raw, areaSchema, declaredKeys(ctx, root, "HEAD"))
	if err != nil {
		return "", nil, fmt.Errorf("%s does not verify: %v", areaRecordPath(sha), err)
	}
	if fields["commit"] != sha {
		return "", nil, fmt.Errorf("%s judges another commit", areaRecordPath(sha))
	}
	var decisions []areaDecision
	dec := json.NewDecoder(strings.NewReader(fields["hunks"]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&decisions); err != nil {
		return "", nil, fmt.Errorf("%s: unreadable hunks", areaRecordPath(sha))
	}
	return fields["by"], decisions, nil
}

var errNoAreaRecord = errors.New("no area record")

// areaControl is audit control C10 (spec §4). owners maps a commit to its spec owners (from C8).
func areaControl(ctx context.Context, root, base string, rep *auditReport, except func(kind, control, commit, text string)) (auditControl, error) {
	c := auditControl{ID: "C10", Question: controlQuestions["C10"], Command: "proofbound review draft <commit>",
		Basis: "A spec may list in scope.paths the files its change may touch, read from the commit's parent. Lines changed outside it must be accepted by a reviewer who did not write the change, in a signed record; a spec without scope.paths is counted, not flagged."}
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
		if cm.Class != classClaimed {
			continue
		}
		res, err := commitArea(ctx, root, cm.Commit)
		if err != nil {
			return c, err
		}
		if len(res.Names) == 0 {
			continue
		}
		subj := cm.Subject
		// Who last set the area is checked first, whatever it is now: an agent that widened, broke
		// or deleted scope.paths must not turn its own change into "inside" or "not declared" (F5).
		parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", cm.Commit+"^1"))
		if setter := areaSetter(ctx, root, base, cm.Commit, areaCandidates(ctx, root, parent, res.Names), agents); setter != "" {
			declared++
			rep.trackFor(cm.Commit).AreaHunks = len(res.Hunks)
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: "area last changed by agent commit " + setter[:12]})
			except("area", "C10", cm.Commit, fmt.Sprintf("%q: its area was last changed by %s, a commit made by an agent", subj, setter[:12]))
			continue
		}
		if res.Invalid != "" {
			declared++
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: printableUnit(res.Invalid)})
			except("area", "C10", cm.Commit, fmt.Sprintf("%q: %s", subj, printableUnit(res.Invalid)))
			continue
		}
		if !res.Declared {
			missing++
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: "area not declared by " + strings.Join(res.Missing, ", ")})
			continue
		}
		declared++
		if len(res.Hunks) == 0 {
			good++
			rep.trackFor(cm.Commit)
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: "inside its area"})
			continue
		}
		problem := ""
		by, decisions, err := readAreaRecord(ctx, root, cm.Commit)
		accepted := map[string]areaDecision{}
		rejected := map[string]bool{}
		switch {
		case errors.Is(err, errNoAreaRecord):
			problem = "no reviewer decided them"
		case err != nil:
			problem = err.Error()
		default:
			if ok, note := independence(cm, []string{by}, owners[cm.Commit]); !ok {
				problem = "the decision is not independent: " + note
			}
			for _, d := range decisions {
				if d.Decision == "accept" && checkReason(d.Reason) == nil {
					accepted[areaHunk{Path: d.Path, Lines: d.Lines}.key()] = d
				}
				if d.Decision == "reject" {
					rejected[areaHunk{Path: d.Path, Lines: d.Lines}.key()] = true
				}
			}
		}
		// What this control trusted, for the review outcomes by agent (CLI-INV-73).
		tc := rep.trackFor(cm.Commit)
		tc.AreaHunks = len(res.Hunks)
		if problem == "" {
			for _, h := range res.Hunks {
				if _, ok := accepted[h.key()]; ok {
					tc.AreaAccepted++
				} else if rejected[h.key()] {
					tc.AreaRejected++
				}
			}
		}
		var open []string
		var reasons []string
		for _, h := range res.Hunks {
			if d, ok := accepted[h.key()]; ok && problem == "" {
				reasons = append(reasons, printableUnit(d.Reason))
				continue
			}
			open = append(open, printableUnit(h.Path)+" "+h.Lines)
		}
		if len(open) == 0 {
			good++
			c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: fmt.Sprintf("%d hunk(s) outside, accepted by %s: %s", len(res.Hunks), by, strings.Join(dedupe(reasons), "; "))})
			continue
		}
		if problem == "" {
			problem = "not accepted"
		}
		list := strings.Join(shortList(open, 10), ", ")
		c.Evidence = append(c.Evidence, auditEvidence{Commit: cm.Commit, Text: fmt.Sprintf("%d hunk(s) outside its area, %s: %s", len(open), problem, list)})
		except("area", "C10", cm.Commit, fmt.Sprintf("%q changed lines outside its spec's area (%s): %s", subj, printableUnit(problem), list))
	}
	switch {
	case declared == 0:
		c.State, c.Count = auditNone, fmt.Sprintf("no change declares an area; %d not declared", missing)
	case good == declared:
		c.State, c.Count = auditHeld, fmt.Sprintf("%d of %d declared; %d not declared", good, declared, missing)
	default:
		c.State, c.Count = auditExcepted, fmt.Sprintf("%d of %d declared; %d not declared", good, declared, missing)
	}
	return c, nil
}

// areaSetter is a commit in base..sha^ that an agent made and that was the last to change (add, edit
// or delete) one of files. Each file is checked on its own: a later human edit of one area file does
// not hide an agent's change to another (review F2).
func areaSetter(ctx context.Context, root, base, sha string, files []string, agents map[string]bool) string {
	for _, f := range files {
		last := strings.TrimSpace(gitOutput(ctx, root, "--literal-pathspecs", "log", "-1", "--format=%H", "--end-of-options", base+".."+sha+"^1", "--", f))
		if len(last) == 40 && agents[last] {
			return last
		}
	}
	return ""
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func shortList(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return append(append([]string{}, in[:n]...), fmt.Sprintf("and %d more", len(in)-n))
}
