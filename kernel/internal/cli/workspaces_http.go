package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// HTTP handlers of the U2 workspaces (CLI-INV-45..49).

// renderPage writes one workspace as HTML, or its JSON twin.
func (s *uiServer) renderPage(w http.ResponseWriter, r *http.Request, name string, page any, twin any) {
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, twin)
		return
	}
	var buf bytes.Buffer
	if err := uiPage.ExecuteTemplate(&buf, name, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (s *uiServer) renderError(w http.ResponseWriter, r *http.Request, status int, err error) {
	if wantsJSON(r) || r.Method == http.MethodPost || strings.HasPrefix(r.URL.Path, "/agent/") {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = uiPage.ExecuteTemplate(w, "error", map[string]string{"Message": err.Error()})
}

// nowPage is the developer's home (O-1).
type nowPage struct {
	V         uiView         `json:"-"`
	Head      string         `json:"as_of_head"`
	Items     []nowItem      `json:"items"`
	Worktrees []worktreeTask `json:"worktrees"`
	Commits   []auditCommit  `json:"commits"`
}

func (s *uiServer) serveNow(w http.ResponseWriter, r *http.Request) {
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		s.renderError(w, r, status, err)
		return
	}
	p := nowPage{V: v, Head: v.Report.Head, Items: nowItems(v.Report, v.Promises), Worktrees: worktreeTasks(r.Context(), s.root), Commits: v.Report.Commits}
	if p.Worktrees == nil {
		p.Worktrees = []worktreeTask{}
	}
	s.renderPage(w, r, "now.html", p, p)
}

// promisesPage is the spec owner's map (O-3).
type promisesPage struct {
	V       uiView          `json:"-"`
	Head    string          `json:"as_of_head"`
	Changes []promiseChange `json:"changes"`
	Counts  map[string]int  `json:"counts"`
}

func (s *uiServer) servePromises(w http.ResponseWriter, r *http.Request) {
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		s.renderError(w, r, status, err)
		return
	}
	p := promisesPage{V: v, Head: v.Report.Head, Changes: v.Promises, Counts: map[string]int{}}
	for _, c := range v.Promises {
		for _, pr := range c.Promises {
			p.Counts[pr.State]++
			if pr.Unclear != "" {
				p.Counts["UNCLEAR"]++
			}
		}
	}
	s.renderPage(w, r, "promises.html", p, p)
}

// benchItem is one commit in the review queue.
type benchItem struct {
	Commit   auditCommit `json:"commit"`
	Promises int         `json:"promises"`
}

// reviewPage is the review bench (O-2): the queue, or one commit.
type reviewPage struct {
	V       uiView       `json:"-"`
	Head    string       `json:"as_of_head"`
	Queue   []benchItem  `json:"queue"`
	Hidden  []string     `json:"hidden"`
	Commit  *auditCommit `json:"commit,omitempty"`
	Draft   *draftFile   `json:"draft,omitempty"`
	Saved   bool         `json:"saved_draft"`
	Diff    string       `json:"-"`
	Refusal string       `json:"refusal,omitempty"`
	// Area says whether the commit's spec declares an area (change-area-spec.md §3).
	Area string `json:"area,omitempty"`
}

func (s *uiServer) serveReview(w http.ResponseWriter, r *http.Request) {
	if !s.acquire(r.Context()) {
		s.renderError(w, r, http.StatusServiceUnavailable, errLedgerBusy)
		return
	}
	defer s.release()
	v, status, err := s.viewLocked(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		s.renderError(w, r, status, err)
		return
	}
	waiting := map[string]bool{}
	for _, c := range v.Report.Controls {
		if c.ID == "C1" {
			for _, e := range c.Evidence {
				if strings.HasPrefix(e.Text, checkNoReview+":") {
					waiting[e.Commit] = true
				}
			}
		}
	}
	var shas []string
	for sha := range waiting {
		shas = append(shas, sha)
	}
	drafts, err := s.drafts(r.Context(), shas)
	if err != nil {
		s.renderError(w, r, http.StatusInternalServerError, err)
		return
	}
	p := reviewPage{V: v, Head: v.Report.Head, Queue: []benchItem{}, Hidden: []string{}}
	for _, c := range v.Report.Commits {
		d, ok := drafts[c.Commit]
		if !waiting[c.Commit] || !ok {
			continue
		}
		if v.Actor.Name != "" {
			if ok, why := eligible(v.Actor.Name, c, draftOwners(d)); !ok {
				p.Hidden = append(p.Hidden, short(c.Commit, 12)+": "+why)
				continue
			}
		}
		n := 0
		for _, vd := range d.Verdicts {
			n += len(vd.Obligations)
		}
		p.Queue = append(p.Queue, benchItem{Commit: c, Promises: n})
	}
	if sha := r.URL.Query().Get("commit"); sha != "" {
		var found *auditCommit
		for i := range v.Report.Commits {
			if v.Report.Commits[i].Commit == sha {
				found = &v.Report.Commits[i]
			}
		}
		d, ok := drafts[sha]
		if found == nil || !ok {
			s.renderError(w, r, http.StatusNotFound, fmt.Errorf("commit %s is not waiting for review in this period", printableUnit(short(sha, 12))))
			return
		}
		d, area, err := withArea(r.Context(), s.root, d)
		if err != nil {
			s.renderError(w, r, http.StatusInternalServerError, err)
			return
		}
		p.Area = areaNote(area)
		if saved, has := localDraft(s.root, sha); has {
			if merged, err := mergeDecisions(d, saved); err == nil {
				d, p.Saved = merged, true
			}
		}
		p.Commit, p.Draft = found, &d
		if err := reviewerFor(v.Actor, *found, d); err != nil {
			p.Refusal = err.Error()
		}
		p.Diff = commitDiff(r.Context(), s.root, sha)
	}
	s.renderPage(w, r, "review.html", p, p)
}

// commitDiff is the change a reviewer reads: stat and patch, at most 1,500 lines.
func commitDiff(ctx context.Context, root, sha string) string {
	out := gitOutput(ctx, root, "show", "--stat", "--patch", "--format=commit %H%nAuthor: %an%nDate:   %ad%n%n    %s%n", "--end-of-options", sha)
	lines := strings.Split(out, "\n")
	if len(lines) > 1500 {
		lines = append(lines[:1500], fmt.Sprintf("… %d more lines; run `git show %s` to read the rest", len(lines)-1500, short(sha, 12)))
	}
	return printableText(strings.Join(lines, "\n"))
}

// printableText keeps newlines and tabs but replaces other control characters.
func printableText(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r >= 0x20 && r != 0x7f {
			return r
		}
		return '?'
	}, s)
}

// reviewDecision is one choice from the bench, addressed by its place in the ledger draft and
// checked against the obligation id there.
type reviewDecision struct {
	Part         string `json:"part"` // "review" or "verdict"
	Group        int    `json:"group"`
	Index        int    `json:"index"`
	ObligationID string `json:"obligation_id"`
	Outcome      string `json:"outcome"`
	Finding      string `json:"finding,omitempty"`
	// Path and Lines address an out-of-area entry (part "area"); Outcome is accept or reject and
	// Finding the reason.
	Path  string `json:"path,omitempty"`
	Lines string `json:"lines,omitempty"`
}

type reviewRequest struct {
	Base      string           `json:"base"`
	Commit    string           `json:"commit"`
	Decisions []reviewDecision `json:"decisions"`
}

// applyDecisions copies the bench's choices onto the ledger draft and names the reviewer.
func applyDecisions(d draftFile, reviewer string, decisions []reviewDecision) (draftFile, error) {
	for _, dec := range decisions {
		switch dec.Part {
		case "review":
			if dec.Group < 0 || dec.Group >= len(d.Reviews) || dec.Index < 0 || dec.Index >= len(d.Reviews[dec.Group].Obligations) {
				return d, errors.New("a decision names no obligation of this commit")
			}
			o := &d.Reviews[dec.Group].Obligations[dec.Index]
			if o.ObligationID != dec.ObligationID {
				return d, errors.New("the commit's promises changed; reload the page")
			}
			if err := checkReasonOrEmpty(dec.Finding); err != nil {
				return d, err
			}
			o.Outcome, o.Finding = dec.Outcome, dec.Finding
		case "verdict":
			if dec.Group < 0 || dec.Group >= len(d.Verdicts) || dec.Index < 0 || dec.Index >= len(d.Verdicts[dec.Group].Obligations) {
				return d, errors.New("a decision names no obligation of this commit")
			}
			o := &d.Verdicts[dec.Group].Obligations[dec.Index]
			if o.ObligationID != dec.ObligationID {
				return d, errors.New("the commit's promises changed; reload the page")
			}
			o.Outcome = dec.Outcome
		case "area":
			if dec.Index < 0 || dec.Index >= len(d.OutOfArea) {
				return d, errors.New("a decision names no line range of this commit")
			}
			a := &d.OutOfArea[dec.Index]
			if a.Path != dec.Path || a.Lines != dec.Lines {
				return d, errors.New("the commit's lines outside its area changed; reload the page")
			}
			if err := checkReasonOrEmpty(dec.Finding); err != nil {
				return d, err
			}
			a.Decision, a.Reason = dec.Outcome, dec.Finding
		default:
			return d, errors.New("a decision has no part")
		}
	}
	d.DeclaredReviewer = reviewer
	return d, nil
}

func checkReasonOrEmpty(s string) error {
	if s == "" {
		return nil
	}
	return checkReason(s)
}

// undecided lists the obligations without a valid outcome; seal would refuse them.
func undecided(d draftFile) []string {
	var out []string
	for _, r := range d.Reviews {
		for _, o := range r.Obligations {
			switch o.Outcome {
			case "VERIFIABLE", "AMBIGUOUS", "UNTESTABLE", "CONTRADICTORY":
			default:
				out = append(out, "is "+o.ObligationID+" clear enough to verify?")
			}
		}
	}
	for _, v := range d.Verdicts {
		for _, o := range v.Obligations {
			switch o.Outcome {
			case "SATISFIED", "NOT_SATISFIED", "INCONCLUSIVE":
			default:
				out = append(out, "does the code keep "+o.ObligationID+"?")
			}
		}
	}
	for _, a := range d.OutOfArea {
		switch {
		case a.Decision == "reject", a.Decision == "accept" && checkReason(a.Reason) == nil:
		case a.Decision == "accept":
			out = append(out, "why accept "+printableUnit(a.Path)+" "+a.Lines+"?")
		default:
			out = append(out, "accept or reject "+printableUnit(a.Path)+" "+a.Lines+" (outside the area)?")
		}
	}
	return out
}

// serveReviewAct handles draft (save the local draft), preview (what sealing writes) and seal.
func (s *uiServer) serveReviewAct(w http.ResponseWriter, r *http.Request, act string) {
	var req reviewRequest
	// Up to areaMaxTotal out-of-area decisions with 500-character reasons (review F4).
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, uiMaxReviewBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || !sha40RE.MatchString(req.Commit) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed request"})
		return
	}
	if !s.acquire(r.Context()) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": errLedgerBusy.Error()})
		return
	}
	defer s.release()
	v, status, err := s.viewLocked(r.Context(), req.Base)
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	var commit *auditCommit
	for i := range v.Report.Commits {
		if v.Report.Commits[i].Commit == req.Commit && v.Report.Commits[i].Class == classClaimed {
			commit = &v.Report.Commits[i]
		}
	}
	if commit == nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "that commit does not claim a spec in this period"})
		return
	}
	drafts, err := s.drafts(r.Context(), []string{req.Commit})
	truth, ok := drafts[req.Commit]
	if err != nil || !ok {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "the record holds no review draft for that commit"})
		return
	}
	if truth, _, err = withArea(r.Context(), s.root, truth); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := reviewerFor(v.Actor, *commit, truth); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	edited, err := applyDecisions(truth, v.Actor.Name, req.Decisions)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	dir := draftDir(s.root, req.Commit)
	saveDraft := func() error {
		body, err := json.MarshalIndent(edited, "", "  ")
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "draft.json"), append(body, '\n'), 0o644)
	}
	short12 := req.Commit[:12]
	switch act {
	case "draft":
		if err := saveDraft(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"saved": filepath.ToSlash(filepath.Join(".proofbound", "drafts", req.Commit, "draft.json"))})
	case "preview":
		if missing := undecided(edited); len(missing) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "decide every promise first", "missing": missing})
			return
		}
		var files []string
		for i := range edited.Reviews {
			files = append(files, fmt.Sprintf("docs/verification/verdicts/RR-%s-%d.md (+ .sig)", short12, i+1))
		}
		for i := range edited.Verdicts {
			files = append(files, fmt.Sprintf("docs/verification/verdicts/OV-%s-%d.md (+ .sig)", short12, i+1))
		}
		if len(edited.OutOfArea) > 0 {
			files = append(files, areaRecordPath(req.Commit)+" (signed)")
		}
		files = append(files, "gates/ready-"+short12+".yaml")
		writeJSON(w, http.StatusOK, map[string]any{"reviewer": v.Actor.Name + " (key " + v.Actor.Fingerprint + ")", "files": files, "commit_message": "review " + short12})
	case "seal":
		if missing := undecided(edited); len(missing) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "decide every promise first", "missing": missing})
			return
		}
		if err := saveDraft(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		var out bytes.Buffer
		if err := s.seal(r.Context(), dir, &out); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		var written []string
		for _, line := range strings.Split(out.String(), "\n") {
			if rel, ok := strings.CutPrefix(line, "wrote "); ok {
				written = append(written, rel)
			}
		}
		committed, note, err := commitPaths(r.Context(), s.root, written, "review "+short12)
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"committed": committed, "note": note, "files": written, "seal": printableText(out.String())})
	}
}

// commitPaths commits exactly these new paths; on a git failure they are unstaged and removed, so a
// retry starts clean (as writeAndCommit does for one record).
func commitPaths(ctx context.Context, root string, rels []string, message string) (bool, string, error) {
	if len(rels) == 0 {
		return false, "", errors.New("the seal wrote nothing")
	}
	undo := func(out string) (bool, string, error) {
		_, _ = gitRun(ctx, root, append([]string{"rm", "-q", "--cached", "--ignore-unmatch", "--"}, rels...)...)
		for _, rel := range rels {
			_ = os.Remove(filepath.Join(root, filepath.FromSlash(rel)))
		}
		return false, "", fmt.Errorf("not recorded: git could not commit the review: %s", out)
	}
	if out, err := gitRun(ctx, root, append([]string{"add", "-f", "--"}, rels...)...); err != nil {
		return undo(out)
	}
	if out, err := gitRun(ctx, root, append([]string{"commit", "-q", "-m", message, "--"}, rels...)...); err != nil {
		return undo(out)
	}
	return true, "committed", nil
}

func (s *uiServer) serveAgentDoc(w http.ResponseWriter, r *http.Request) {
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		s.renderError(w, r, status, err)
		return
	}
	task, _ := currentTask(s.root)
	ans := agentNext(v.Report, v.Promises, task)
	raw, _ := json.MarshalIndent(ans, "", "  ")
	s.renderPage(w, r, "agent.html", map[string]any{"V": v, "Task": task, "Example": string(raw)}, ans)
}

func (s *uiServer) serveAgentNext(w http.ResponseWriter, r *http.Request) {
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	task := r.URL.Query().Get("task")
	if task == "" {
		task, _ = currentTask(s.root)
	} else if name, ok := strings.CutPrefix(task, "sdd:"); !ok || !taskNameRE.MatchString(name) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "task must be sdd:<change>"})
		return
	}
	writeJSON(w, http.StatusOK, agentNext(v.Report, v.Promises, task))
}
