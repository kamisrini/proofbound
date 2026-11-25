package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The Change page and Issue the warranty (docs/plans/ui-reviewer-spec.md R2.1, R2.2).

// changeCommit is one commit of a change, with every step the flow walks.
type changeCommit struct {
	Commit    auditCommit `json:"commit"`
	Gate      string      `json:"gate"` // the C1 line: "PASS: …", "NO REVIEW: …", "BLOCKED: …"
	State     string      `json:"state"`
	Reviewers []string    `json:"reviewers"`
	Warranty  string      `json:"warranty"` // the C3 line for this commit, or ""
	Issued    string      `json:"issued,omitempty"`
}

// changeView is one change (a claimed spec) through the period.
type changeView struct {
	Name    string         `json:"name"`
	Spec    []string       `json:"spec"`
	Commits []changeCommit `json:"commits"`
	// Action is the one thing this viewer may do now, and Link where it is done.
	Action string `json:"action,omitempty"`
	Link   string `json:"link,omitempty"`
	Note   string `json:"note"`
}

type changePage struct {
	V       uiView       `json:"-"`
	Changes []changeView `json:"changes"`
	One     *changeView  `json:"change,omitempty"`
}

// issuedWarranties maps a commit to the docs/warranties/<head12>/ folder whose index, as committed
// at HEAD (never an uncommitted file), lists it VALID (code review F3).
func issuedWarranties(ctx context.Context, root string) map[string]string {
	out := map[string]string{}
	for _, name := range strings.Split(strings.TrimSpace(gitOutput(ctx, root, "-c", "core.quotePath=false", "ls-tree", "-r", "--name-only", "HEAD", "--", "docs/warranties/")), "\n") {
		if path.Base(name) != "index.json" || strings.Count(name, "/") != 3 {
			continue
		}
		raw, err := gitBytes(ctx, root, "show", "HEAD:"+name)
		if err != nil {
			continue
		}
		var index bundleIndex
		if json.Unmarshal(raw, &index) != nil {
			continue
		}
		for _, e := range index.Entries {
			if e.State == bundleValid {
				out[e.Commit] = path.Dir(name)
			}
		}
	}
	return out
}

// releaseBlockers is why the period's check is not WARRANTED: a claiming commit that does not pass,
// or code with no spec claim that no waiver covers (code review F1). Empty when Issue may go ahead.
func releaseBlockers(rep auditReport) []string {
	var out []string
	claims := 0
	for _, c := range rep.Controls {
		if c.ID != "C1" {
			continue
		}
		for _, e := range c.Evidence {
			claims++
			if !strings.HasPrefix(e.Text, checkPass+":") {
				out = append(out, short(e.Commit, 12)+" "+printableUnit(e.Text))
			}
		}
	}
	for _, e := range rep.Exceptions {
		if e.Kind == "uncovered" {
			out = append(out, short(e.Commit, 12)+" "+printableUnit(e.Text))
		}
	}
	if claims == 0 {
		out = append(out, "no commit in this period claims a spec")
	}
	return out
}

// buildChanges groups the period's claiming commits by the change they claim.
func (s *uiServer) buildChanges(ctx context.Context, v uiView) []changeView {
	gate, warr := map[string]string{}, map[string]string{}
	for _, c := range v.Report.Controls {
		for _, e := range c.Evidence {
			switch c.ID {
			case "C1":
				gate[e.Commit] = e.Text
			case "C3":
				warr[e.Commit] = e.Text
			}
		}
	}
	reviewers := map[string][]string{}
	for _, r := range v.Report.Roles {
		reviewers[r.Commit] = r.Reviewers
	}
	issued := issuedWarranties(ctx, s.root)
	byName := map[string]*changeView{}
	var order []string
	for _, c := range v.Report.Commits {
		if c.Class != classClaimed {
			continue
		}
		for _, name := range commitClaims(ctx, s.root, c.Commit) {
			cv, ok := byName[name]
			if !ok {
				cv = &changeView{Name: name, Spec: []string{}, Commits: []changeCommit{}}
				if area, err := commitArea(ctx, s.root, c.Commit); err == nil {
					for _, f := range area.Files {
						cv.Spec = append(cv.Spec, path.Dir(f))
					}
				}
				byName[name] = cv
				order = append(order, name)
			}
			state, _, _ := strings.Cut(gate[c.Commit], ":")
			cv.Commits = append(cv.Commits, changeCommit{Commit: c, Gate: gate[c.Commit], State: state, Reviewers: reviewers[c.Commit], Warranty: warr[c.Commit], Issued: issued[c.Commit]})
		}
	}
	out := []changeView{}
	blockers := releaseBlockers(v.Report)
	for _, name := range order {
		cv := byName[name]
		s.changeAction(v, cv, blockers)
		out = append(out, *cv)
	}
	return out
}

// changeAction names the viewer's next action for a change (BR-audit-room-ui O-7): Review this when
// the viewer's key may review a commit of the change that waits for review; else Issue the warranty
// when the viewer may issue the period's warranty (a usable key that reviewed none of the period's
// changes, and a period whose check is WARRANTED); otherwise none. The note says where it stands.
func (s *uiServer) changeAction(v uiView, cv *changeView, blockers []string) {
	last := cv.Commits[len(cv.Commits)-1]
	switch last.State {
	case checkNoReview:
		cv.Note = "Waiting for a reviewer who did not write it."
	case checkBlocked:
		cv.Note = "Blocked: " + strings.TrimSpace(strings.TrimPrefix(last.Gate, checkBlocked+":")) + ". The author fixes it in a new commit."
	case checkPass:
		cv.Note = "Warranted by the check; no warranty issued yet."
		if last.Issued != "" {
			cv.Note = "Warranted, and issued in " + last.Issued + "."
		} else if len(blockers) > 0 {
			cv.Note += " The release waits for the rest of the period: " + strings.Join(blockers, "; ") + "."
		}
	default:
		cv.Note = last.Gate
	}
	if v.Actor.Name == "" || v.Actor.Problem != "" {
		return
	}
	for _, c := range cv.Commits {
		if c.State != checkNoReview {
			continue
		}
		if ok, _ := eligible(v.Actor.Name, c.Commit, nil); ok {
			cv.Action, cv.Link = "Review this", "/review?commit="+c.Commit.Commit
			return
		}
	}
	if !viewerMayIssue(v, blockers) || last.Issued != "" {
		return
	}
	cv.Action, cv.Link = "Issue the warranty", "/audit#issue"
}

// viewerMayIssue: a usable key that reviewed none of the period's changes, and a WARRANTED period.
func viewerMayIssue(v uiView, blockers []string) bool {
	if v.Actor.Name == "" || v.Actor.Problem != "" || len(blockers) > 0 {
		return false
	}
	for _, role := range v.Report.Roles {
		for _, r := range role.Reviewers {
			if pbwarranty.KeyName(r) == pbwarranty.KeyName(v.Actor.Name) {
				return false
			}
		}
	}
	return true
}

func (s *uiServer) serveChange(w http.ResponseWriter, r *http.Request) {
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
	p := changePage{V: v, Changes: s.buildChanges(r.Context(), v)}
	if name := r.URL.Query().Get("claim"); name != "" {
		for i := range p.Changes {
			if p.Changes[i].Name == name {
				p.One = &p.Changes[i]
			}
		}
		if p.One == nil {
			s.renderError(w, r, http.StatusNotFound, fmt.Errorf("no commit in this period claims %s", printableUnit(name)))
			return
		}
	}
	s.renderPage(w, r, "change.html", p, p)
}

// --- Issue the warranty (R2.2) ---

type issueRequest struct {
	Base    string `json:"base"`
	Confirm bool   `json:"confirm"`
}

// releasedAt says whether the release at this point was issued already: HEAD's tree holds
// docs/warranties/<head12>/index.json, or HEAD is a release commit of its parent, adding exactly
// docs/warranties/<parent12>/ and nothing else (code review F3: never a worktree file or a subject).
func releasedAt(ctx context.Context, root, head string) string {
	if _, err := gitBytes(ctx, root, "cat-file", "-e", head+":docs/warranties/"+head[:12]+"/index.json"); err == nil {
		return "docs/warranties/" + head[:12]
	}
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", head+"^1"))
	if len(parent) != 40 {
		return ""
	}
	folder := "docs/warranties/" + parent[:12] + "/"
	changed := strings.Fields(gitOutput(ctx, root, "-c", "core.quotePath=false", "diff-tree", "--no-commit-id", "--name-only", "-r", "--no-renames", parent, head))
	if len(changed) == 0 {
		return ""
	}
	for _, f := range changed {
		if !strings.HasPrefix(f, folder) {
			return ""
		}
	}
	if _, err := gitBytes(ctx, root, "cat-file", "-e", head+":"+folder+"index.json"); err != nil {
		return ""
	}
	return strings.TrimSuffix(folder, "/")
}

func (s *uiServer) serveIssue(w http.ResponseWriter, r *http.Request) {
	var req issueRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
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
	if v.Actor.Problem != "" || v.Actor.Name == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "no usable signing key: " + v.Actor.Problem})
		return
	}
	head := v.Report.Head
	if dirty := strings.TrimSpace(gitOutput(r.Context(), s.root, "status", "--porcelain", "--", "gates", ".proofbound/keys", "docs/warranties")); dirty != "" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "gates/, .proofbound/keys/ or docs/warranties/ differ from HEAD; commit them first:\n" + printableText(dirty)})
		return
	}
	if done := releasedAt(r.Context(), s.root, head); done != "" {
		writeJSON(w, http.StatusOK, map[string]any{"issued_already": done})
		return
	}
	// The period's check must read WARRANTED: every claiming commit passes and no code is uncovered
	// (code review F1); the bundle's own entries cover only the claiming commits.
	if blockers := releaseBlockers(v.Report); len(blockers) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "the period's check is not WARRANTED, so no warranty is issued", "not_warranted": blockers})
		return
	}
	index, files, err := s.issueCompute(r.Context(), v.Report.Base)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	// The browser issues only a whole release: any commit not VALID, and nothing is written (§2.2).
	var not []string
	for _, e := range index.Entries {
		if e.State != bundleValid {
			not = append(not, fmt.Sprintf("%s %s %s %s", e.Commit[:12], printableUnit(e.Claims), e.State, printableUnit(e.Reason)))
		}
	}
	if len(index.Entries) == 0 {
		not = append(not, "no commit in this period claims a spec")
	}
	if len(not) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "not every change is warranted, so no warranty is issued", "not_warranted": not})
		return
	}
	folder := "docs/warranties/" + index.Head[:12]
	var names []string
	for name := range files {
		names = append(names, folder+"/"+name)
	}
	sort.Strings(names)
	summary := map[string]any{"folder": folder, "files": names, "signer": v.Actor.Name + " (key " + v.Actor.Fingerprint + ")",
		"describes": index.Head, "commit_message": "release warranties at " + index.Head[:12], "warranted": len(index.Entries)}
	if !req.Confirm {
		writeJSON(w, http.StatusOK, summary)
		return
	}
	if err := writeDirAtomically(filepath.Join(s.root, filepath.FromSlash(folder)), files); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if now := strings.TrimSpace(gitOutput(r.Context(), s.root, "rev-parse", "HEAD")); now != index.Head {
		_ = os.RemoveAll(filepath.Join(s.root, filepath.FromSlash(folder)))
		writeJSON(w, http.StatusConflict, map[string]string{"error": "HEAD moved while the bundle was computed; nothing issued, try again"})
		return
	}
	committed, note, err := commitPaths(context.WithoutCancel(r.Context()), s.root, names, "release warranties at "+index.Head[:12])
	if err != nil {
		_ = os.RemoveAll(filepath.Join(s.root, filepath.FromSlash(folder)))
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	summary["committed"], summary["note"] = committed, note
	writeJSON(w, http.StatusOK, summary)
}

var errNoStore = errors.New("no record")

// issueComputeDefault opens the record and computes the bundle in memory.
func issueComputeDefault(root, databaseURL string) func(ctx context.Context, base string) (bundleIndex, map[string][]byte, error) {
	return func(ctx context.Context, base string) (bundleIndex, map[string][]byte, error) {
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return bundleIndex{}, nil, errors.Join(errNoStore, err)
		}
		index, files, err := bundleCompute(ctx, root, ledger, base)
		return index, files, errors.Join(err, closeStore(ledger))
	}
}
