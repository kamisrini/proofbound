package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

// Review outcomes by agent (docs/plans/agent-track-record-spec.md, CLI-INV-73..74): how each agent's
// changes fared when someone checked them, from records committed at HEAD. Counts only, never a score.

const trackSchema = "proofbound.track-record.v1"

type trackAgent struct {
	Name           string `json:"name"`
	Commits        int    `json:"commits"`
	Claimed        int    `json:"claimed"`
	Uncovered      int    `json:"uncovered"`
	Acceptable     int    `json:"acceptable"`
	NeedsWork      int    `json:"needs_work"`
	NotIndependent int    `json:"reviewed_not_independent"`
	Unreviewed     int    `json:"unreviewed"`
	// Hunks changed outside the spec's area that git lists now, and how a trusted AREA record decided
	// them (the rest are undecided).
	AreaHunks    int `json:"area_hunks"`
	AreaRejected int `json:"area_hunks_rejected"`
	AreaAccepted int `json:"area_hunks_accepted"`
	// Parts of the change no test noticed when undone, from a trusted TESTS record.
	TestsNotCaught int `json:"uncaught_parts"`
	TestsRejected  int `json:"uncaught_parts_rejected"`
	TestsAccepted  int `json:"uncaught_parts_accepted"`
	WholeNotCaught int `json:"whole_change_not_caught"`
	// Lines from trusted PROV records: by a recorded session, and with no recorded agent edit.
	LinesBySession  int `json:"lines_by_a_session"`
	LinesUnrecorded int `json:"lines_no_recorded_agent_edit"`
	Open            int `json:"open_exceptions"`
}

type trackSession struct {
	Session    string   `json:"session"`
	Commits    []string `json:"commits"`
	Lines      int      `json:"lines"`
	Acceptable int      `json:"acceptable"`
	NeedsWork  int      `json:"needs_work"`
	// NotIndependent were reviewed, but not independently (C8).
	NotIndependent int `json:"reviewed_not_independent"`
	Unreviewed     int `json:"unreviewed"`
}

type trackReport struct {
	Schema   string         `json:"schema"`
	Base     string         `json:"base"`
	Head     string         `json:"head"`
	Agents   []trackAgent   `json:"agents"`
	Sessions []trackSession `json:"sessions"`
}

// reviewStatus is a claimed commit's review at HEAD: ACCEPTABLE, NEEDS_WORK (any verdict), or "" when
// no verdict for it is committed.
func reviewStatus(ctx context.Context, root, head, sha string, verdictFiles []string) (string, []string) {
	status := ""
	var reviewers []string
	prefix := "docs/verification/verdicts/OV-" + sha[:12] + "-"
	for _, path := range verdictFiles {
		if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, ".md") {
			continue
		}
		data, err := gitBytes(ctx, root, "show", head+":"+path)
		if err != nil {
			continue
		}
		v, err := connectorreviews.ParseObligationVerdict(path, data)
		if err != nil || v.ReviewedCommit != sha {
			continue
		}
		reviewers = append(reviewers, v.DeclaredReviewer)
		switch {
		case v.Status == "NEEDS_WORK":
			status = "NEEDS_WORK"
		case status == "":
			status = v.Status
		}
	}
	return status, reviewers
}

// computeTrack builds the review outcomes by agent for rep's range. Review outcomes come from the
// verdicts committed at HEAD for the commit, with C8's independence; the area, tests and provenance
// counts are exactly what C10, C11 and C12 trusted (rep.track), never a second reading of the records.
func computeTrack(ctx context.Context, root string, rep auditReport) trackReport {
	tr := trackReport{Schema: trackSchema, Base: rep.Base, Head: rep.Head, Agents: []trackAgent{}, Sessions: []trackSession{}}
	verdicts := strings.Split(strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "-r", "--name-only", rep.Head, "--", "docs/verification/verdicts/")), "\n")
	open := map[string]int{}
	for _, e := range rep.Exceptions {
		if e.Outcome == nil || !e.Outcome.Verified {
			open[e.Commit]++
		}
	}
	independent, owners := map[string]bool{}, map[string][]string{}
	for _, r := range rep.Roles {
		independent[r.Commit] = r.Independent || len(r.Reviewers) == 0
		owners[r.Commit] = r.Owners
	}
	agents := map[string]*trackAgent{}
	sessions := map[string]*trackSession{}
	for _, c := range rep.Commits {
		status := ""
		if c.Class == classClaimed {
			var reviewers []string
			status, reviewers = reviewStatus(ctx, root, rep.Head, c.Commit, verdicts)
			// Independence from C8's role, and from the verdicts' own declared reviewers, so a verdict
			// the ledger has not seen yet is still judged (review M2).
			if ok, _ := independence(c, reviewers, owners[c.Commit]); status != "" && (!ok || !independent[c.Commit]) {
				status = "NOT_INDEPENDENT"
			}
		}
		tc := rep.track[c.Commit]
		if tc == nil {
			tc = &commitTrack{}
		}
		for _, name := range dedupe(c.Agents) {
			a := agents[name]
			if a == nil {
				a = &trackAgent{Name: name}
				agents[name] = a
			}
			a.Commits++
			a.Open += open[c.Commit]
			switch c.Class {
			case classUncovered:
				a.Uncovered++
			case classClaimed:
				a.Claimed++
				switch status {
				case "ACCEPTABLE":
					a.Acceptable++
				case "NEEDS_WORK":
					a.NeedsWork++
				case "NOT_INDEPENDENT":
					a.NotIndependent++
				default:
					a.Unreviewed++
				}
			}
			a.AreaHunks += tc.AreaHunks
			a.AreaRejected += tc.AreaRejected
			a.AreaAccepted += tc.AreaAccepted
			a.TestsNotCaught += tc.TestsNotCaught
			a.TestsRejected += tc.TestsRejected
			a.TestsAccepted += tc.TestsAccepted
			if tc.WholeNotCaught {
				a.WholeNotCaught++
			}
			a.LinesBySession += tc.LinesTraced
			a.LinesUnrecorded += tc.LinesUnrecorded
		}
		for id, n := range tc.Sessions {
			t := sessions[id]
			if t == nil {
				t = &trackSession{Session: id, Commits: []string{}}
				sessions[id] = t
			}
			t.Lines += n
			t.Commits = append(t.Commits, c.Commit)
			switch status {
			case "ACCEPTABLE":
				t.Acceptable++
			case "NEEDS_WORK":
				t.NeedsWork++
			case "NOT_INDEPENDENT":
				t.NotIndependent++
			default:
				t.Unreviewed++
			}
		}
	}
	for _, a := range agents {
		tr.Agents = append(tr.Agents, *a)
	}
	sort.Slice(tr.Agents, func(i, j int) bool {
		if tr.Agents[i].Commits != tr.Agents[j].Commits {
			return tr.Agents[i].Commits > tr.Agents[j].Commits
		}
		return tr.Agents[i].Name < tr.Agents[j].Name
	})
	for _, s := range sessions {
		tr.Sessions = append(tr.Sessions, *s)
	}
	sort.Slice(tr.Sessions, func(i, j int) bool { return tr.Sessions[i].Session < tr.Sessions[j].Session })
	return tr
}

// trackNotes are the caveats printed with the review outcomes, in the terminal and on the page.
var trackNotes = []string{
	"Counts, not a score. Agent names are as git records them (proven only for commits signed by a declared key, C9).",
	"A commit counts once for every agent named on it, so the agents' counts do not add up to the range.",
	"NEEDS_WORK already includes a rejected out-of-area hunk or a rejected uncaught part.",
	"Area, tests and line counts come only from records the controls C10, C11 and C12 trusted; hunks are counted as git lists them now.",
}

// writeTrack prints the review outcomes by agent as text (`proofbound report agents <base>`).
func writeTrack(w io.Writer, tr trackReport) error {
	p := func(format string, a ...any) error { _, err := fmt.Fprintf(w, format, a...); return err }
	if err := p("Review outcomes by agent, %s..%s, from records committed at HEAD\n", tr.Base[:12], tr.Head[:12]); err != nil {
		return err
	}
	for _, n := range trackNotes {
		if err := p("  %s\n", n); err != nil {
			return err
		}
	}
	if err := p("\n"); err != nil {
		return err
	}
	if len(tr.Agents) == 0 {
		if err := p("No commit in this range names a coding agent.\n"); err != nil {
			return err
		}
	}
	for _, a := range tr.Agents {
		if err := p("%s (agent)\n  commits %d: %d claimed a spec, %d changed code without one\n  review of the claimed: %d ACCEPTABLE, %d NEEDS_WORK, %d reviewed but not independent, %d not reviewed\n  hunks outside the spec's area: %d (%d rejected, %d accepted)\n  parts no test caught when undone: %d (%d rejected, %d accepted); whole change not caught: %d\n  lines: %d by a recorded session, %d with no recorded agent edit\n  open exceptions: %d\n\n",
			printableUnit(a.Name), a.Commits, a.Claimed, a.Uncovered, a.Acceptable, a.NeedsWork, a.NotIndependent, a.Unreviewed, a.AreaHunks, a.AreaRejected, a.AreaAccepted, a.TestsNotCaught, a.TestsRejected, a.TestsAccepted, a.WholeNotCaught, a.LinesBySession, a.LinesUnrecorded, a.Open); err != nil {
			return err
		}
	}
	if len(tr.Sessions) > 0 {
		if err := p("Sessions in the provenance records (linked by line hashes, not by time; a line can count for more than one session):\n"); err != nil {
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "SESSION\tCOMMITS\tLINES\tACCEPTABLE\tNEEDS_WORK\tNOT INDEPENDENT\tNOT REVIEWED")
		for _, s := range tr.Sessions {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%d\n", printableUnit(s.Session), len(s.Commits), s.Lines, s.Acceptable, s.NeedsWork, s.NotIndependent, s.Unreviewed)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func writeTrackJSON(w io.Writer, tr trackReport) error {
	out, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}
