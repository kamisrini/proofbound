package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The four other workspaces (U2, docs/plans/ui-u2-workspaces-spec.md, CLI-INV-45..49).

// Promise states (U2 spec section 1).
const (
	promiseProven     = "PROVEN"
	promiseNotKept    = "NOT KEPT"
	promiseUnsure     = "INCONCLUSIVE"
	promiseStale      = "CHANGED AFTER PROOF"
	promiseUnreviewed = "CLAIMED, NOT REVIEWED"
	promiseUnclaimed  = "NOT CLAIMED"
)

type promiseView struct {
	ID        string   `json:"id"`
	Statement string   `json:"statement"`
	State     string   `json:"state"`
	Reviewer  string   `json:"reviewer,omitempty"`
	Unclear   string   `json:"unclear,omitempty"` // AMBIGUOUS, UNTESTABLE or CONTRADICTORY
	Finding   string   `json:"finding,omitempty"`
	FoundBy   string   `json:"found_by,omitempty"`
	Commits   []string `json:"commits"`
}

type promiseChange struct {
	Change   string        `json:"change"`
	IntentID string        `json:"intent_id"`
	Owners   []string      `json:"owners"`
	Promises []promiseView `json:"promises"`
}

// promiseEvent is the slice of a ledger event the promise rule needs.
type promiseEvent struct {
	seq     int64
	source  core.Source
	kind    core.Kind
	payload json.RawMessage
}

// readPromiseEvents reads the events computePromises needs in one ledger pass.
func readPromiseEvents(ctx context.Context, ledger *store.Store) ([]promiseEvent, error) {
	var out []promiseEvent
	err := ledger.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		e := record.Event
		switch {
		case e.Source == core.SourceIntentSDD && e.Kind == core.KindChangeIntent,
			e.Kind == core.KindRequirement && strings.HasPrefix(string(e.Source), "intent."),
			e.Kind == core.KindReviewVerdict, e.Kind == core.KindRequirementReview,
			e.Source == core.SourceGit && e.Kind == core.KindCommitRecorded:
			out = append(out, promiseEvent{seq: record.Seq, source: e.Source, kind: e.Kind, payload: e.Payload})
		}
		return nil
	})
	return out, err
}

// computePromises applies the U2 promise rule to the events (CLI-INV-45). Pure, so it is tested on
// built events.
func computePromises(events []promiseEvent) []promiseChange {
	sort.Slice(events, func(i, j int) bool { return events[i].seq < events[j].seq })
	latest := map[string]connectorintent.ChangeIntent{}
	requirements := map[string]connectorintent.Requirement{}
	type decided struct{ outcome, reviewer string }
	verdicts := map[string]decided{} // requirement source/id/sha + obligation -> latest outcome
	type proof struct{ reqKey, outcome, reviewer string }
	latestAny := map[string]proof{}   // change intent + obligation -> its latest verdict, on any revision
	reviewsOf := map[string]decided{} // requirement source/id/sha + obligation -> review outcome
	findings := map[string]string{}   // same key -> finding
	claims := map[string][]string{}   // change intent id -> claiming commits
	for _, ev := range events {
		switch ev.kind {
		case core.KindChangeIntent:
			var ci connectorintent.ChangeIntent
			if json.Unmarshal(ev.payload, &ci) == nil && ci.IntentID != "" {
				latest[ci.IntentID] = ci
			}
		case core.KindRequirement:
			var r connectorintent.Requirement
			if json.Unmarshal(ev.payload, &r) == nil {
				requirements[string(ev.source)+"\x00"+r.RequirementID+"\x00"+r.ArtifactSHA256] = r
			}
		case core.KindReviewVerdict:
			var v connectorreviews.ObligationVerdict
			if json.Unmarshal(ev.payload, &v) != nil || v.Schema != "proofbound.obligation-verdict.v2" {
				continue
			}
			// A verdict belongs to one change intent: it never proves another change's promise.
			ci := v.ChangeIntent.RecordID
			for _, o := range v.Obligations {
				verdicts[ci+"\x00"+o.Source+"\x00"+o.RequirementID+"\x00"+o.ArtifactSHA256+"\x00"+o.ObligationID] = decided{o.Outcome, v.DeclaredReviewer}
				latestAny[ci+"\x00"+o.ObligationID] = proof{o.Source + "\x00" + o.RequirementID + "\x00" + o.ArtifactSHA256, o.Outcome, v.DeclaredReviewer}
			}
		case core.KindRequirementReview:
			var r connectorreviews.RequirementReview
			if json.Unmarshal(ev.payload, &r) != nil {
				continue
			}
			for _, o := range r.Outcomes {
				k := r.Requirement.Source + "\x00" + r.Requirement.RecordID + "\x00" + r.Requirement.ArtifactSHA256 + "\x00" + o.ObligationID
				reviewsOf[k] = decided{o.Outcome, r.DeclaredReviewer}
				findings[k] = o.Finding
			}
		case core.KindCommitRecorded:
			var c connectorgit.Commit
			if json.Unmarshal(ev.payload, &c) != nil {
				continue
			}
			for _, ref := range c.IntentRefs {
				if ref.Provider == "sdd" && !containsString(claims[ref.RecordID], c.SHA) {
					claims[ref.RecordID] = append(claims[ref.RecordID], c.SHA)
				}
			}
		}
	}
	ids := make([]string, 0, len(latest))
	for id := range latest {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := []promiseChange{}
	for _, id := range ids {
		ci := latest[id]
		pc := promiseChange{Change: changeName(ci.ArtifactPath), IntentID: id, Owners: []string{}, Promises: []promiseView{}}
		for _, t := range ci.Targets {
			req := requirements[t.Source+"\x00"+t.RecordID+"\x00"+t.ArtifactSHA256]
			if o := printableUnit(req.DeclaredOwner); o != "" && !containsString(pc.Owners, o) {
				pc.Owners = append(pc.Owners, o)
			}
			statements := map[string]string{}
			for _, o := range req.Obligations {
				statements[o.ID] = o.Statement
			}
			for _, oid := range t.ObligationIDs {
				k := t.Source + "\x00" + t.RecordID + "\x00" + t.ArtifactSHA256 + "\x00" + oid
				p := promiseView{ID: oid, Statement: printableUnit(statements[oid]), Commits: append([]string{}, claims[id]...)}
				switch v, ok := verdicts[id+"\x00"+k]; {
				case ok && v.outcome == "SATISFIED":
					p.State, p.Reviewer = promiseProven, printableUnit(v.reviewer)
				case ok && v.outcome == "NOT_SATISFIED":
					p.State, p.Reviewer = promiseNotKept, printableUnit(v.reviewer)
				case ok:
					p.State, p.Reviewer = promiseUnsure, printableUnit(v.reviewer)
				case latestAny[id+"\x00"+oid].outcome != "":
					// Reviewed on an earlier revision (a spec edit gives the requirement a new id and digest).
					// Only the latest verdict counts (review ui-u2-review-20261122 F4): with the same words it
					// still stands; with other words a proof is stale and anything else is unreviewed.
					last := latestAny[id+"\x00"+oid]
					same := false
					for _, o := range requirements[last.reqKey].Obligations {
						same = same || (o.ID == oid && o.Statement == statements[oid])
					}
					switch {
					case same && last.outcome == "SATISFIED":
						p.State, p.Reviewer = promiseProven, printableUnit(last.reviewer)
					case same && last.outcome == "NOT_SATISFIED":
						p.State, p.Reviewer = promiseNotKept, printableUnit(last.reviewer)
					case same:
						p.State, p.Reviewer = promiseUnsure, printableUnit(last.reviewer)
					case last.outcome == "SATISFIED":
						p.State = promiseStale
					case len(claims[id]) > 0:
						p.State = promiseUnreviewed
					default:
						p.State = promiseUnclaimed
					}
				case len(claims[id]) > 0:
					p.State = promiseUnreviewed
				default:
					p.State = promiseUnclaimed
				}
				if r, ok := reviewsOf[k]; ok && r.outcome != "VERIFIABLE" {
					p.Unclear, p.Finding, p.FoundBy = r.outcome, printableUnit(findings[k]), printableUnit(r.reviewer)
				}
				pc.Promises = append(pc.Promises, p)
			}
		}
		out = append(out, pc)
	}
	return out
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// nowItem is one thing that needs a person (CLI-INV-46).
type nowItem struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Why     string `json:"why"`
	Command string `json:"command"`
	Link    string `json:"link"`
	Commit  string `json:"commit,omitempty"`
}

// nowItems derives the attention list from the audit and the promises.
func nowItems(rep auditReport, promises []promiseChange) []nowItem {
	items := []nowItem{}
	for _, c := range rep.Controls {
		if c.ID != "C1" {
			continue
		}
		for _, e := range c.Evidence {
			state, reason, _ := strings.Cut(e.Text, ": ")
			switch state {
			case checkNoReview:
				items = append(items, nowItem{Kind: "review", Commit: e.Commit, Title: "Waiting for a reviewer who did not write it", Why: "Nothing it claims counts as proven until someone independent reviews it.", Command: "proofbound review draft " + short(e.Commit, 12), Link: "/review?commit=" + e.Commit})
			case checkBlocked, checkUnbound:
				items = append(items, nowItem{Kind: "blocked", Commit: e.Commit, Title: "Claims a spec but is " + state, Why: reason, Command: "proofbound check " + short(rep.Base, 12), Link: "/audit"})
			}
		}
	}
	for _, e := range rep.Exceptions {
		if e.Kind == "uncovered" && (e.Outcome == nil || !e.Outcome.Verified) {
			items = append(items, nowItem{Kind: "uncovered", Commit: e.Commit, Title: "Code without a spec", Why: e.Text, Command: "proofbound task start <change>   # then commit code with the claim", Link: "/audit"})
		}
	}
	for _, pc := range promises {
		for _, p := range pc.Promises {
			switch {
			case p.State == promiseStale:
				items = append(items, nowItem{Kind: "drift", Title: pc.Change + " " + p.ID + " changed after it was proven", Why: "The words changed; nobody has checked the code against the new words.", Command: "proofbound report drift", Link: "/promises"})
			case p.State == promiseNotKept:
				items = append(items, nowItem{Kind: "not-kept", Title: pc.Change + " " + p.ID + " is not kept", Why: "Its reviewer found the code does not keep it.", Command: "proofbound task start " + pc.Change, Link: "/promises"})
			case p.Unclear != "":
				items = append(items, nowItem{Kind: "unclear", Title: pc.Change + " " + p.ID + " is " + strings.ToLower(p.Unclear) + " to its reviewer", Why: p.Finding, Command: "edit the spec, then ask for a re-review", Link: "/promises"})
			}
		}
	}
	return items
}

// worktreeTask is one worktree and the agent task it holds.
type worktreeTask struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
	Task   string `json:"task"`
	Head   string `json:"head"`
}

func worktreeTasks(ctx context.Context, root string) []worktreeTask {
	var out []worktreeTask
	var cur *worktreeTask
	for _, line := range strings.Split(gitOutput(ctx, root, "worktree", "list", "--porcelain"), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			out = append(out, worktreeTask{Path: strings.TrimPrefix(line, "worktree ")})
			cur = &out[len(out)-1]
		case cur != nil && strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case cur != nil && strings.HasPrefix(line, "HEAD "):
			cur.Head = short(strings.TrimPrefix(line, "HEAD "), 12)
		}
	}
	for i := range out {
		task, err := currentTask(out[i].Path)
		switch err {
		case nil:
			out[i].Task = task
		case errNoTask:
			out[i].Task = ""
		default:
			out[i].Task = "unreadable task file"
		}
		out[i].Path = printableUnit(filepath.Base(out[i].Path))
		out[i].Branch = printableUnit(out[i].Branch)
	}
	return out
}

// eligible reports whether reviewer may review a commit (CLI-INV-47): not its author, an agent named
// on it, its verified signer, or an owner of a requirement it targets, by key name.
func eligible(reviewer string, c auditCommit, owners []string) (bool, string) {
	ok, note := independence(c, []string{reviewer}, owners)
	if ok {
		return true, ""
	}
	return false, strings.Replace(note, "reviewer "+reviewer+" is", "you are", 1)
}

// draftOwners are the requirement owners a ledger draft names.
func draftOwners(d draftFile) []string {
	var owners []string
	for _, r := range d.Reviews {
		if !containsString(owners, r.DeclaredOwner) {
			owners = append(owners, r.DeclaredOwner)
		}
	}
	return owners
}

// localDraft reads a saved local draft for sha, if any.
func localDraft(root, sha string) (draftFile, bool) {
	raw, err := os.ReadFile(filepath.Join(draftDir(root, sha), "draft.json"))
	if err != nil {
		return draftFile{}, false
	}
	var d draftFile
	if json.Unmarshal(raw, &d) != nil {
		return draftFile{}, false
	}
	return d, true
}

// agentStep is one next step for an agent (CLI-INV-48).
type agentStep struct {
	Do      string `json:"do"`
	Command string `json:"command,omitempty"`
	Why     string `json:"why"`
	For     string `json:"for,omitempty"`
}

type agentBlock struct {
	Kind   string `json:"kind"`
	Commit string `json:"commit,omitempty"`
	Who    string `json:"who"`
	YouMay bool   `json:"you_may"`
}

type agentAnswer struct {
	Head      string         `json:"as_of_head"`
	Task      string         `json:"task"`
	Change    *promiseChange `json:"change"`
	Commits   []checkView    `json:"commits"`
	Uncovered []string       `json:"uncovered"`
	Blocking  []agentBlock   `json:"blocking"`
	Next      []agentStep    `json:"next"`
}

type checkView struct {
	Commit string `json:"commit"`
	State  string `json:"state"`
	Reason string `json:"reason"`
}

// agentNext is the next-step rule for an agent holding task (CLI-INV-48).
func agentNext(rep auditReport, promises []promiseChange, task string) agentAnswer {
	ans := agentAnswer{Head: rep.Head, Task: task, Commits: []checkView{}, Uncovered: []string{}, Blocking: []agentBlock{}, Next: []agentStep{}}
	name := strings.TrimPrefix(task, "sdd:")
	for i := range promises {
		if promises[i].Change == name {
			ans.Change = &promises[i]
		}
	}
	for _, c := range rep.Controls {
		if c.ID != "C1" {
			continue
		}
		for _, e := range c.Evidence {
			state, reason, _ := strings.Cut(e.Text, ": ")
			if state == checkPass {
				reason = ""
			}
			if ans.Change == nil || containsString(changeCommits(*ans.Change), e.Commit) {
				ans.Commits = append(ans.Commits, checkView{Commit: e.Commit, State: state, Reason: reason})
			}
		}
	}
	for _, e := range rep.Exceptions {
		if e.Kind == "uncovered" {
			ans.Uncovered = append(ans.Uncovered, e.Commit)
		}
	}
	if task == "" {
		ans.Next = append(ans.Next, agentStep{Do: "start_task", Command: "proofbound task start <change>", Why: "no task is active in this worktree; every commit that changes code must claim a spec"})
		return ans
	}
	if ans.Change == nil {
		ans.Next = append(ans.Next, agentStep{Do: "stop_and_ask_spec_owner", Why: "the task names a change Proofbound does not know; its spec may not be synced or may not exist"})
		return ans
	}
	if len(changeCommits(*ans.Change)) == 0 {
		ans.Next = append(ans.Next, agentStep{Do: "commit_code", Why: "no commit claims this change yet; with the task active, the hook adds the claim"})
	}
	unreviewed := false
	for _, cv := range ans.Commits {
		switch cv.State {
		case checkNoReview:
			unreviewed = true
			ans.Blocking = append(ans.Blocking, agentBlock{Kind: "needs_independent_review", Commit: cv.Commit, Who: "not the author, not an agent on it, not the spec owner", YouMay: false})
		case checkBlocked:
			// GATE-INV-17: a reviewed commit can be BLOCKED with every promise proven, when the reviewer
			// recorded NEEDS_WORK (a rejected line outside the area or a part no test caught). Never "done".
			ans.Next = append(ans.Next, agentStep{Do: "fix_code", Why: short(cv.Commit, 12) + " is BLOCKED: " + cv.Reason, Command: "commit a fix with the task active"})
		}
	}
	for _, p := range ans.Change.Promises {
		switch {
		case p.Unclear != "":
			ans.Next = append(ans.Next, agentStep{Do: "stop_and_ask_spec_owner", Why: p.ID + " is " + strings.ToLower(p.Unclear) + ": " + p.Finding})
		case p.State == promiseNotKept:
			ans.Next = append(ans.Next, agentStep{Do: "fix_code", Why: p.ID + " was reviewed as not kept", Command: "commit a fix with the task active"})
		case p.State == promiseStale:
			ans.Next = append(ans.Next, agentStep{Do: "hand_off_for_re_review", Why: p.ID + " changed after it was proven"})
		}
	}
	if unreviewed {
		for _, b := range ans.Blocking {
			ans.Next = append(ans.Next, agentStep{Do: "hand_off", Command: "proofbound review draft " + short(b.Commit, 12), For: "a reviewer agent or person in another session", Why: "you may not review your own change"})
		}
	}
	if len(ans.Next) == 0 {
		ans.Next = append(ans.Next, agentStep{Do: "done", Why: "every promise of the change is proven by an independent review"})
	}
	return ans
}

// changeCommits are the commits that claim a change (every promise of it lists the same ones).
func changeCommits(pc promiseChange) []string {
	var out []string
	for _, p := range pc.Promises {
		for _, c := range p.Commits {
			if !containsString(out, c) {
				out = append(out, c)
			}
		}
	}
	return out
}

// reviewerFor checks that the local actor may review commit c, given the ledger draft's owners.
func reviewerFor(a actor, c auditCommit, d draftFile) error {
	if a.key == nil {
		return fmt.Errorf("%s", a.Problem)
	}
	if ok, why := eligible(a.Name, c, draftOwners(d)); !ok {
		return fmt.Errorf("you may not review this change: %s", why)
	}
	if pbwarranty.KeyName(a.Name) == "" {
		return fmt.Errorf("your key name is not usable")
	}
	return nil
}
