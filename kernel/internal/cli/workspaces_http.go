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
	// OutOfArea and Tests say what needs a decision (ui-reviewer-spec §1.5).
	OutOfArea int    `json:"out_of_area"`
	Tests     string `json:"tests,omitempty"`
}

// benchNote is a reviewed commit the gate blocks (§1.9), or the result of a seal (§1.8).
type benchNote struct {
	Commit auditCommit `json:"commit"`
	Text   string      `json:"text"`
	Pass   bool        `json:"pass"`
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
	Refusal string       `json:"refusal,omitempty"`
	// Area says whether the commit's spec declares an area (change-area-spec.md §3).
	Area string `json:"area,omitempty"`
	// ui-reviewer-spec R1: who signs, the tests, the spec, the flagged change, the needs-work list,
	// the result of a seal, and a job still running for this commit.
	Signer        string      `json:"signer,omitempty"`
	TestsDeclared bool        `json:"tests_declared"`
	TestsInvalid  string      `json:"tests_invalid,omitempty"`
	Spec          []specText  `json:"spec,omitempty"`
	Lines         []diffLine  `json:"-"`
	FlaggedPast   []string    `json:"flagged_past_cut,omitempty"`
	ProvError     string      `json:"provenance_error,omitempty"`
	NeedsWork     []benchNote `json:"needs_work"`
	Sealed        *benchNote  `json:"sealed,omitempty"`
	Job           *benchJob   `json:"job,omitempty"`
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
	p := reviewPage{V: v, Head: v.Report.Head, Queue: []benchItem{}, Hidden: []string{}, NeedsWork: []benchNote{}}
	if v.Actor.Fingerprint != "" && v.Actor.Problem == "" {
		p.Signer = v.Actor.Name + " (key " + v.Actor.Fingerprint + ")"
	}
	gateText := map[string]string{}
	for _, c := range v.Report.Controls {
		if c.ID == "C1" {
			for _, e := range c.Evidence {
				gateText[e.Commit] = e.Text
			}
		}
	}
	for _, c := range v.Report.Commits {
		if t := gateText[c.Commit]; strings.HasPrefix(t, checkBlocked+":") && strings.Contains(t, "NEEDS_WORK") {
			p.NeedsWork = append(p.NeedsWork, benchNote{Commit: c, Text: t})
		}
		if sealed := r.URL.Query().Get("sealed"); sealed != "" && sealed == c.Commit && gateText[c.Commit] != "" {
			p.Sealed = &benchNote{Commit: c, Text: gateText[c.Commit], Pass: strings.HasPrefix(gateText[c.Commit], checkPass+":")}
		}
	}
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
		item := benchItem{Commit: c, Promises: n}
		if area, err := commitArea(r.Context(), s.root, c.Commit); err == nil {
			item.OutOfArea = len(area.Hunks)
		}
		if testsDeclared(r.Context(), s.root, c.Commit) {
			item.Tests = "tests not run yet"
			if local, has := localDraft(s.root, c.Commit); has && local.Tests != nil && local.Tests.Declared {
				missed := 0
				for _, h := range local.Tests.Hunks {
					if h.Result != testsCaught {
						missed++
					}
				}
				item.Tests = fmt.Sprintf("tests run: %d part(s) not caught", missed)
				if missed == 0 {
					item.Tests = "tests run: every part caught"
				}
			}
		}
		p.Queue = append(p.Queue, item)
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
		p.TestsDeclared = testsDeclared(r.Context(), s.root, sha)
		p.TestsInvalid = testsConfigProblem(r.Context(), s.root, sha)
		if prov, err := commitProvenance(r.Context(), s.root, sha); err == nil {
			d.Provenance = &prov
		} else {
			p.ProvError = printableUnit(err.Error())
		}
		p.Commit, p.Draft = found, &d
		if err := reviewerFor(v.Actor, *found, d); err != nil {
			p.Refusal = err.Error()
		}
		p.Spec = benchSpec(r.Context(), s.root, sha, area)
		p.Lines, p.FlaggedPast = benchDiff(r.Context(), s.root, sha, benchFlags(d))
		if job, ok := s.jobs.snapshot(sha); ok {
			p.Job = &job
		}
	}
	s.renderPage(w, r, "review.html", p, p)
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
	if d.Tests != nil {
		t := *d.Tests
		t.Hunks = append([]testsHunk{}, t.Hunks...)
		d.Tests = &t
	}
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
		case "tests":
			if d.Tests == nil || dec.Index < 0 || dec.Index >= len(d.Tests.Hunks) {
				return d, errors.New("a decision names no tested part of this commit; run the tests first")
			}
			h := &d.Tests.Hunks[dec.Index]
			if h.Path != dec.Path || h.Lines != dec.Lines {
				return d, errors.New("the commit's test results changed; reload the page")
			}
			if h.Result == testsCaught {
				return d, errors.New("a part the tests caught needs no decision")
			}
			if err := checkReasonOrEmpty(dec.Finding); err != nil {
				return d, err
			}
			h.Decision, h.Reason = dec.Outcome, dec.Finding
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
func undecided(d draftFile, testsDeclared bool) []string {
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
	if testsDeclared {
		switch t := d.Tests; {
		case t == nil || !t.Declared:
			out = append(out, "run the tests first")
		case t.Invalid != "":
			out = append(out, "fix "+testsConfigPath+": "+printableUnit(t.Invalid))
		default:
			for _, h := range t.Hunks {
				switch {
				case h.Result == testsCaught, h.Decision == "reject", h.Decision == "accept" && checkReason(h.Reason) == nil:
				case h.Decision == "accept":
					out = append(out, "why accept "+printableUnit(h.Path)+" "+h.Lines+"?")
				default:
					out = append(out, "accept or reject "+printableUnit(h.Path)+" "+h.Lines+" (the tests did not notice it undone)?")
				}
			}
		}
	}
	return out
}

// serveReviewAct handles draft (save the local draft), preview (what sealing writes), tests (run the
// declared tests in the background) and seal (in the background when tests are declared: they run
// first, with no lock held; ui-reviewer-spec §1.1).
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
	// Only a commit waiting for review: a reviewed one's code is never run, nor its draft overwritten,
	// from the bench (code review F2).
	if !waitingForReview(v.Report, req.Commit) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "that commit is not waiting for review"})
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
	// The test run is never in the record: it is carried from the local draft, and the request only
	// decides its hunks (review F1).
	declared := testsDeclared(r.Context(), s.root, req.Commit)
	if local, has := localDraft(s.root, req.Commit); has && local.Commit == req.Commit {
		truth.Tests = local.Tests
	}
	edited, err := applyDecisions(truth, v.Actor.Name, req.Decisions)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	dir := draftDir(s.root, req.Commit)
	short12 := req.Commit[:12]
	switch act {
	case "draft":
		if err := writeLocalDraft(s.root, req.Commit, edited); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"saved": filepath.ToSlash(filepath.Join(".proofbound", "drafts", req.Commit, "draft.json"))})
	case "preview":
		if missing := undecided(edited, declared); len(missing) > 0 {
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
		if declared {
			files = append(files, testsRecordPath(req.Commit)+" (signed)")
		}
		files = append(files, provRecordPath(req.Commit)+" (signed)", "gates/ready-"+short12+".yaml")
		writeJSON(w, http.StatusOK, map[string]any{"reviewer": v.Actor.Name + " (key " + v.Actor.Fingerprint + ")", "files": files, "commit_message": "review " + short12, "warnings": sealWarnings(edited)})
	case "tests":
		if !declared {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "this project declares no tests (" + testsConfigPath + ")"})
			return
		}
		if why := testsConfigProblem(r.Context(), s.root, req.Commit); why != "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": why})
			return
		}
		if err := writeLocalDraft(s.root, req.Commit, edited); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		sha := req.Commit
		err := s.jobs.start(s.life, "tests", sha, func(ctx context.Context, say func(string)) (map[string]any, error) {
			res, err := s.runTests(ctx, sha, say)
			if err != nil {
				return nil, err
			}
			// The merge and write take the request gate, so a Save cannot land in between (F7).
			select {
			case s.gate <- struct{}{}:
				defer s.release()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			d := edited
			if local, has := localDraft(s.root, sha); has && local.Commit == sha {
				d = local
			}
			kept := keepTestsDecisions(res, d.Tests)
			d.Tests = &kept
			if err := writeLocalDraft(s.root, sha, d); err != nil {
				return nil, err
			}
			return map[string]any{"saved": true}, nil
		})
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": "tests"})
	case "seal":
		if missing := undecided(edited, declared); len(missing) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "decide every promise first", "missing": missing})
			return
		}
		if err := writeLocalDraft(s.root, req.Commit, edited); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if !declared {
			result, err := s.sealAndCommit(r.Context(), req.Commit, dir, nil)
			if err != nil {
				writeJSON(w, sealStatus(err), map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, result)
			return
		}
		// The decisions submitted with this click are sealed, even if the draft is saved again
		// meanwhile: they are copied to a folder of the job's own.
		snapshot, err := os.MkdirTemp("", "proofbound-seal-")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		body, _ := json.MarshalIndent(edited, "", "  ")
		if err := os.WriteFile(filepath.Join(snapshot, "draft.json"), append(body, '\n'), 0o644); err != nil {
			_ = os.RemoveAll(snapshot)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		sha := req.Commit
		err = s.jobs.start(s.life, "seal", sha, func(ctx context.Context, say func(string)) (map[string]any, error) {
			defer os.RemoveAll(snapshot)
			say("tests: running them again before sealing (a seal never trusts an earlier run)")
			res, err := s.runTests(ctx, sha, say)
			if err != nil {
				return nil, err
			}
			return s.sealAndCommit(ctx, sha, snapshot, &sealTestsRun{Commit: sha, Result: res})
		})
		if err != nil {
			_ = os.RemoveAll(snapshot)
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": "seal"})
	}
}

// errCommitRefused marks a seal whose files git would not commit (409).
type errCommitRefused struct{ error }

func sealStatus(err error) int {
	var refused errCommitRefused
	if errors.As(err, &refused) {
		return http.StatusConflict
	}
	return http.StatusUnprocessableEntity
}

// sealAndCommit seals dir and commits exactly the written files. Without ran (no tests declared) it is
// called under the request lock; with ran it takes the lock itself, after the tests.
func (s *uiServer) sealAndCommit(ctx context.Context, sha, dir string, ran *sealTestsRun) (map[string]any, error) {
	if ran != nil {
		select {
		case s.gate <- struct{}{}:
			defer s.release()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	// Once sealing starts it finishes, even if the server is stopping: a half-made commit would leave
	// staged files behind (code review F8).
	ctx = context.WithoutCancel(ctx)
	var out bytes.Buffer
	if err := s.seal(ctx, dir, ran, &out); err != nil {
		return nil, err
	}
	var written []string
	for _, line := range strings.Split(out.String(), "\n") {
		if rel, ok := strings.CutPrefix(line, "wrote "); ok {
			written = append(written, rel)
		}
	}
	committed, note, err := commitPaths(ctx, s.root, written, "review "+sha[:12])
	if err != nil {
		return nil, errCommitRefused{err}
	}
	return map[string]any{"committed": committed, "note": note, "files": written, "seal": printableText(out.String())}, nil
}

// sealWarnings says, before sealing, what will make the review NEEDS_WORK whatever the promises say.
func sealWarnings(d draftFile) []string {
	var out []string
	for _, a := range d.OutOfArea {
		if a.Decision == "reject" {
			out = append(out, "you reject "+printableUnit(a.Path)+" "+a.Lines+" (outside the area): the review will be NEEDS_WORK")
		}
	}
	if t := d.Tests; t != nil && t.Declared {
		switch {
		case t.Baseline == testsUnreliable:
			out = append(out, "the tests passed, then failed run again unchanged: the review will be NEEDS_WORK")
		case t.Baseline != "" && t.Baseline != testsPass:
			out = append(out, "the tests do not pass at this commit ("+t.Baseline+"): the review will be NEEDS_WORK")
		case len(t.Hunks) > 0 && t.Whole != testsCaught:
			out = append(out, "the tests did not catch the whole change undone: the review will be NEEDS_WORK")
		}
		for _, h := range t.Hunks {
			if h.Result != testsCaught && h.Decision == "reject" {
				out = append(out, "you reject "+printableUnit(h.Path)+" "+h.Lines+" (the tests did not notice it): the review will be NEEDS_WORK")
			}
		}
	}
	return out
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
		return false, "", fmt.Errorf("not recorded: git could not commit %s: %s", message, out)
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
