package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The Audit room's computation (U1, docs/plans/ui-u1-audit-room-spec.md section 3, CLI-INV-38). It
// only reads: git, the files at HEAD, and the ledger. Warranties are built in memory and discarded.

const auditSchema = "proofbound.audit.v1"

// Control states.
const (
	auditHeld     = "HELD"
	auditExcepted = "EXCEPTION"
	auditListed   = "LISTED"
	auditNone     = "NONE"
)

// Commit classes (C5).
const (
	classClaimed      = "claimed"
	classUncovered    = "uncovered"
	classHousekeeping = "housekeeping"
)

type auditCommit struct {
	Commit  string `json:"commit"`
	Subject string `json:"subject"`
	Class   string `json:"class"`
	// Author, Date and Agents are as git records them: self-declared, never signed. Agents are the
	// author and Co-Authored-By names that look like a coding agent (agentName).
	Author string   `json:"author"`
	Date   string   `json:"date"`
	Agents []string `json:"agents"`
	// Signature is verified, unsigned, bad or unchecked (commitSignatures); Signer is the declared key
	// name of a verified signature. Unlike Author, a verified Signer is proven.
	Signature string `json:"signature"`
	Signer    string `json:"signer,omitempty"`
}

type auditEvidence struct {
	Commit string `json:"commit,omitempty"`
	Text   string `json:"text"`
}

type auditControl struct {
	ID       string          `json:"id"`
	Question string          `json:"question"`
	State    string          `json:"state"`
	Count    string          `json:"count"`
	Basis    string          `json:"basis"`
	Command  string          `json:"command"`
	Evidence []auditEvidence `json:"evidence"`
}

type auditException struct {
	ID      string       `json:"id"`
	Kind    string       `json:"kind"`
	Control string       `json:"control"`
	Commit  string       `json:"commit"`
	Text    string       `json:"text"`
	Outcome *outcomeView `json:"outcome,omitempty"`
}

type auditCounts struct {
	Total        int `json:"total"`
	Claimed      int `json:"claimed"`
	Uncovered    int `json:"uncovered"`
	Housekeeping int `json:"housekeeping"`
}

// auditReport is the transcript (everything here) plus, outside the transcript, what only this
// machine knows: the actor, the attestations read back and a sync error.
type auditReport struct {
	Schema     string           `json:"schema"`
	Base       string           `json:"base"`
	Head       string           `json:"head"`
	Headline   string           `json:"headline"`
	Counts     auditCounts      `json:"counts"`
	Commits    []auditCommit    `json:"commits"`
	Controls   []auditControl   `json:"controls"`
	Exceptions []auditException `json:"exceptions"`
	Roles      []auditRole      `json:"roles"`
	Sessions   []auditSession   `json:"sessions"`
}

// computeAudit answers C1..C7 for base..HEAD. base must already resolve to a commit.
func computeAudit(ctx context.Context, root string, ledger *store.Store, base string) (auditReport, error) {
	var rep auditReport
	rep.Schema = auditSchema
	var err error
	if rep.Base, err = resolveCommit(ctx, root, base); err != nil {
		return rep, err
	}
	if rep.Head, err = resolveCommit(ctx, root, "HEAD"); err != nil {
		return rep, err
	}
	shas, err := commitsInRange(ctx, root, rep.Base)
	if err != nil {
		return rep, err
	}
	infos, err := commitInfos(ctx, root, rep.Base)
	if err != nil {
		return rep, err
	}
	sigs, err := commitSignatures(ctx, root, rep.Base)
	if err != nil {
		return rep, err
	}
	claims, err := classifiedRange(ctx, root, ledger, rep.Base)
	if err != nil {
		return rep, err
	}
	uncovered, err := uncoveredCommits(ctx, root, rep.Base, claims)
	if err != nil {
		return rep, err
	}

	// C5 classes first: every other control refers to them.
	class := map[string]string{}
	for _, l := range claims {
		class[l.commit] = classClaimed
	}
	for _, l := range uncovered {
		class[l.commit] = classUncovered
	}
	for _, sha := range shas {
		c := class[sha]
		if c == "" {
			c = classHousekeeping
		}
		switch c {
		case classClaimed:
			rep.Counts.Claimed++
		case classUncovered:
			rep.Counts.Uncovered++
		default:
			rep.Counts.Housekeeping++
		}
		info := infos[sha]
		rep.Commits = append(rep.Commits, auditCommit{Commit: sha, Subject: printableUnit(info.subject), Class: c, Author: printableUnit(info.author), Date: info.date, Agents: info.agents, Signature: sigs[sha][0], Signer: sigs[sha][1]})
	}
	rep.Counts.Total = len(shas)
	rep.Exceptions = []auditException{}
	except := func(kind, control, commit, text string) {
		rep.Exceptions = append(rep.Exceptions, auditException{ID: kind + "-" + commit[:12], Kind: kind, Control: control, Commit: commit, Text: text})
	}
	subject := func(sha string) string { return printableUnit(infos[sha].subject) }
	rangeArg := rep.Base[:12]

	// C1: readiness gates.
	c1 := auditControl{ID: "C1", Question: "Every commit that claims a spec passed its readiness gate.", Command: "proofbound check " + rangeArg,
		Basis: "The readiness gate passes only when every promise the commit claims has a sealed review in the record."}
	passed := 0
	for _, l := range claims {
		if l.state == checkPass {
			passed++
			c1.Evidence = append(c1.Evidence, auditEvidence{Commit: l.commit, Text: "PASS: " + printableUnit(l.claims)})
			continue
		}
		text := l.state + ": " + printableUnit(l.reason)
		c1.Evidence = append(c1.Evidence, auditEvidence{Commit: l.commit, Text: text})
		except("gate", "C1", l.commit, fmt.Sprintf("%q claims %s but is %s: %s", subject(l.commit), printableUnit(l.claims), l.state, printableUnit(l.reason)))
	}
	switch {
	case len(claims) == 0:
		c1.State, c1.Count = auditNone, "no commit in the range claims a spec"
	case passed == len(claims):
		c1.State, c1.Count = auditHeld, fmt.Sprintf("%d of %d", passed, len(claims))
	default:
		c1.State, c1.Count = auditExcepted, fmt.Sprintf("%d of %d passed", passed, len(claims))
	}

	// C2: uncovered code.
	c2 := auditControl{ID: "C2", Question: "No commit changed code without claiming a spec.", Command: "proofbound check " + rangeArg,
		Basis: "A commit that claims no spec may change only specs, docs, gates and Proofbound's own records."}
	for _, l := range uncovered {
		c2.Evidence = append(c2.Evidence, auditEvidence{Commit: l.commit, Text: printableUnit(l.reason)})
		except("uncovered", "C2", l.commit, fmt.Sprintf("%q changed code with no spec claim: %s", subject(l.commit), uncoveredPaths(l.reason)))
	}
	if len(uncovered) == 0 {
		c2.State, c2.Count = auditHeld, fmt.Sprintf("%d of %d commits claim a spec or change no code", len(shas), len(shas))
	} else {
		c2.State, c2.Count = auditExcepted, fmt.Sprintf("%d commit(s) changed code with no claim", len(uncovered))
	}

	// C3: warranties, built in memory as `warranty bundle` does.
	c3 := auditControl{ID: "C3", Question: "Every passing claim carries a warranty that verifies against this repository.", Command: "proofbound warranty bundle " + rangeArg,
		Basis: "A warranty is VALID only when a reviewer who is not the change's owner signed it with a key the repository declares."}
	if passed == 0 {
		c3.State, c3.Count = auditNone, "no passing claim to warrant"
	} else if a := localActor(ctx, root); a.key == nil {
		// A warranty must be signed by a key HEAD declares, so without one every warranty would read as
		// an exception that says nothing about the change (spec review F7).
		c3.State, c3.Count = auditNone, "not checked"
		c3.Basis = "Not checked on this machine: " + a.Problem + "."
	} else {
		valid, checked := 0, 0
		for _, l := range claims {
			if l.state != checkPass {
				continue
			}
			// A warranty's issuer may not be its reviewer, so a machine signing as one of the change's
			// reviewers cannot check it; that says nothing about the change (owner try-out).
			if body, err := warrantyBody(ctx, ledger, l.commit); err == nil && reviewedBy(body, a.Name) {
				c3.Evidence = append(c3.Evidence, auditEvidence{Commit: l.commit, Text: "not checked here: this machine signs as " + a.Name + ", one of its reviewers; check it from another key"})
				continue
			}
			checked++
			_, _, _, unreviewed, err := buildWarranty(ctx, root, ledger, l.commit)
			switch {
			case err != nil:
				c3.Evidence = append(c3.Evidence, auditEvidence{Commit: l.commit, Text: "NOT EXPORTABLE: " + printableUnit(err.Error())})
				except("warranty", "C3", l.commit, fmt.Sprintf("%q passed its gate but its warranty cannot be built: %s", subject(l.commit), printableUnit(err.Error())))
			case unreviewed != nil:
				c3.Evidence = append(c3.Evidence, auditEvidence{Commit: l.commit, Text: "UNREVIEWED: " + printableUnit(unreviewed.Error())})
				except("warranty", "C3", l.commit, fmt.Sprintf("%q has a warranty no reviewer signed: %s", subject(l.commit), printableUnit(unreviewed.Error())))
			default:
				valid++
				c3.Evidence = append(c3.Evidence, auditEvidence{Commit: l.commit, Text: "VALID"})
			}
		}
		switch {
		case checked == 0:
			c3.State, c3.Count = auditNone, "not checked from a reviewer's key"
		case valid == checked:
			c3.State, c3.Count = auditHeld, fmt.Sprintf("%d of %d VALID", valid, checked)
		default:
			c3.State, c3.Count = auditExcepted, fmt.Sprintf("%d of %d VALID", valid, checked)
		}
	}

	// C4: drift disclosed (disclosure is the control, so drift is not an exception).
	c4 := auditControl{ID: "C4", Question: "Promises that changed after they were verified are disclosed.", Command: "proofbound report drift",
		Basis: "A promise whose words changed after its review is shown as STALE, never as proven."}
	var drift bytes.Buffer
	if err := reportDrift(ctx, root, ledger, &drift); err != nil {
		// One control that cannot be computed must not blank the other six: say why, count nothing
		// (owner try-out: Proofbound's own ledger holds a dangling intent revision).
		c4.State, c4.Count = auditNone, "not computed"
		c4.Basis = "Not computed: `proofbound report drift` fails on this record: " + printableUnit(err.Error()) + "."
	} else if c4.Evidence = driftEvidence(drift.String()); len(c4.Evidence) == 0 {
		c4.State, c4.Count = auditHeld, "no verified promise has changed"
	} else {
		c4.State, c4.Count = auditListed, fmt.Sprintf("%d change(s) disclosed", len(c4.Evidence))
	}

	// C5: every commit accounted for.
	c5 := auditControl{ID: "C5", Question: "Every commit in the range is accounted for.", Command: "git rev-list --count " + rangeArg + "..HEAD",
		Basis:    "Each commit git lists for the range is classed exactly once: claimed a spec, changed code without one, or changed only specs, docs and records.",
		Evidence: []auditEvidence{{Text: fmt.Sprintf("%d claimed, %d uncovered, %d specs, docs and records only", rep.Counts.Claimed, rep.Counts.Uncovered, rep.Counts.Housekeeping)}}}
	// C5 is a disclosure, not a test: it cannot fail, so it is never counted as held (spec review F8).
	c5.State, c5.Count = auditListed, fmt.Sprintf("%d commits, each in one class", rep.Counts.Total)

	// C6: changes to who may sign and to the check itself.
	c6 := auditControl{ID: "C6", Question: "Changes to who may sign, and to the check itself, are listed for review.", Command: "git log --stat " + rangeArg + "..HEAD -- .proofbound/keys .proofbound/sdd-profile.json .github/workflows gates",
		Basis: "Declared keys decide whose signature counts; the profile decides which specs are governed; the workflow runs the check; a changed or deleted gate changes what a review must prove. Reviews add gates; that is not listed."}
	for _, sha := range shas {
		paths, err := trustPaths(ctx, root, sha)
		if err != nil {
			return rep, err
		}
		if len(paths) == 0 {
			continue
		}
		shown := strings.Join(paths, ", ")
		c6.Evidence = append(c6.Evidence, auditEvidence{Commit: sha, Text: printableUnit(shown)})
		except("trust", "C6", sha, fmt.Sprintf("%q changed %s", subject(sha), printableUnit(shown)))
	}
	if len(c6.Evidence) == 0 {
		c6.State, c6.Count = auditHeld, "none"
	} else {
		c6.State, c6.Count = auditListed, fmt.Sprintf("%d commit(s)", len(c6.Evidence))
	}

	// C7: the claimed spec existed before its code.
	c7 := auditControl{ID: "C7", Question: "Each claimed spec existed before the code that claims it.", Command: "git ls-tree -d <commit>^ -- specs/<change> openspec/changes/<change>",
		Basis: "The change's spec folder must already be in the commit's first parent: permission before the work, not written with it."}
	checked, before := 0, 0
	for _, l := range claims {
		names, _ := claimLines(gitOutput(ctx, root, "log", "-1", "--format=%B", l.commit))
		if len(names) == 0 {
			continue
		}
		checked++
		var missing []string
		for _, claim := range names {
			name := strings.TrimPrefix(claim, "sdd:")
			if !taskNameRE.MatchString(name) || !specInParent(ctx, root, l.commit, name) {
				missing = append(missing, printableUnit(claim))
			}
		}
		if len(missing) == 0 {
			before++
			c7.Evidence = append(c7.Evidence, auditEvidence{Commit: l.commit, Text: printableUnit(l.claims) + ": spec in the parent"})
			continue
		}
		list := strings.Join(missing, ", ")
		c7.Evidence = append(c7.Evidence, auditEvidence{Commit: l.commit, Text: list + ": not in the parent"})
		except("order", "C7", l.commit, fmt.Sprintf("%q claims %s, whose spec was not committed before it", subject(l.commit), list))
	}
	switch {
	case checked == 0:
		c7.State, c7.Count = auditNone, "no claiming commit to check"
	case before == checked:
		c7.State, c7.Count = auditHeld, fmt.Sprintf("%d of %d", before, checked)
	default:
		c7.State, c7.Count = auditExcepted, fmt.Sprintf("%d of %d", before, checked)
	}

	// C8: no change was reviewed by its own author, agent, signer or spec owner.
	c8 := auditControl{ID: "C8", Question: "No change was reviewed by its own author, agent or spec owner.", Command: "proofbound warranty export <commit>",
		Basis: "Each claiming commit's reviewers, read from its review records, are compared by key name with its git author, the agents named on it, its verified signer and its spec owners."}
	rep.Roles = []auditRole{}
	reviewed, independent := 0, 0
	for _, c := range rep.Commits {
		if c.Class != classClaimed {
			continue
		}
		role := roleFor(ctx, ledger, c)
		rep.Roles = append(rep.Roles, role)
		if len(role.Reviewers) == 0 {
			continue
		}
		reviewed++
		if role.Independent {
			independent++
			c8.Evidence = append(c8.Evidence, auditEvidence{Commit: c.Commit, Text: "reviewed by " + strings.Join(role.Reviewers, ", ")})
			continue
		}
		c8.Evidence = append(c8.Evidence, auditEvidence{Commit: c.Commit, Text: role.Note})
		except("self-review", "C8", c.Commit, fmt.Sprintf("%q: %s", subject(c.Commit), role.Note))
	}
	switch {
	case reviewed == 0:
		c8.State, c8.Count = auditNone, "no reviewed change in the range"
	case independent == reviewed:
		c8.State, c8.Count = auditHeld, fmt.Sprintf("%d of %d", independent, reviewed)
	default:
		c8.State, c8.Count = auditExcepted, fmt.Sprintf("%d of %d independent", independent, reviewed)
	}

	// C9: which commits are signed by a declared key (a disclosure; a bad signature is an exception).
	c9 := auditControl{ID: "C9", Question: "Commits are signed by a key the repository declares.", Command: "git log --show-signature " + rangeArg + "..HEAD",
		Basis: "A git author name is self-declared; a commit signed by a declared key proves which person or agent made it. Commits are checked with git's SSH signature verification against .proofbound/keys at HEAD."}
	verified, bad := 0, 0
	for _, c := range rep.Commits {
		switch c.Signature {
		case sigVerified:
			verified++
		case sigBad:
			bad++
			c9.Evidence = append(c9.Evidence, auditEvidence{Commit: c.Commit, Text: "the signature does not verify"})
			except("signature", "C9", c.Commit, fmt.Sprintf("%q carries a signature that does not verify", subject(c.Commit)))
		case sigOther:
			c9.Evidence = append(c9.Evidence, auditEvidence{Commit: c.Commit, Text: "signed, but not by a declared key (or not checkable here)"})
		}
	}
	c9.Evidence = append([]auditEvidence{{Text: fmt.Sprintf("%d of %d commits signed by a declared key", verified, len(shas))}}, c9.Evidence...)
	if bad > 0 {
		c9.State, c9.Count = auditExcepted, fmt.Sprintf("%d bad signature(s)", bad)
	} else {
		c9.State, c9.Count = auditListed, fmt.Sprintf("%d of %d signed by a declared key", verified, len(shas))
	}

	// C10: the declared change area (docs/plans/change-area-spec.md §4).
	c10, err := areaControl(ctx, root, rep.Base, &rep, except)
	if err != nil {
		return rep, err
	}

	if rep.Sessions, err = sessionsInPeriod(ctx, root, ledger, rep.Base, rep.Commits); err != nil {
		return rep, err
	}
	if rep.Sessions == nil {
		rep.Sessions = []auditSession{}
	}

	rep.Controls = []auditControl{c1, c2, c3, c4, c5, c6, c7, c8, c9, c10}
	if rep.Counts.Total == 0 {
		// An empty range holds nothing: no control can be held in it (spec review F5).
		for i := range rep.Controls {
			rep.Controls[i].State, rep.Controls[i].Count = auditNone, "the range holds no commits"
		}
	}
	for i := range rep.Controls {
		if rep.Controls[i].Evidence == nil {
			rep.Controls[i].Evidence = []auditEvidence{}
		}
	}
	rep.Headline = auditHeadline(rep)
	return rep, nil
}

// auditHeadline is "<h> of <c> controls held for this range." plus the exception count, where c counts
// the controls that are tests (HELD or EXCEPTION); LISTED disclosures and NONE are not counted.
func auditHeadline(rep auditReport) string {
	if rep.Counts.Total == 0 {
		return "The range holds no commits."
	}
	held, counted := 0, 0
	for _, c := range rep.Controls {
		if c.State == auditNone || c.State == auditListed {
			continue
		}
		counted++
		if c.State == auditHeld {
			held++
		}
	}
	h := fmt.Sprintf("%d of %d controls held for this range.", held, counted)
	if len(rep.Exceptions) == 0 {
		return h + " No exceptions."
	}
	closed := 0
	for _, e := range rep.Exceptions {
		if e.Outcome != nil && e.Outcome.Verified {
			closed++
		}
	}
	return fmt.Sprintf("%s %d exception(s), %d with a signed outcome.", h, len(rep.Exceptions), closed)
}

// transcript is the canonical JSON of the report: what an attestation signs and /transcript serves.
func transcript(rep auditReport) ([]byte, error) {
	raw, err := json.Marshal(rep)
	if err != nil {
		return nil, err
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}

type commitInfo struct {
	subject, author, date string
	agents                []string
}

// agentMarkers are the name or e-mail fragments that mark a coding agent in git authorship. Git
// identities are self-declared, so this is a disclosure, not an identification.
var agentMarkers = []string{"anthropic.com", "claude", "copilot", "[bot]", "cursor", "codex", "openai", "devin", "aider", "gemini", "windsurf", "amazon q"}

func agentName(ident string) bool {
	l := strings.ToLower(ident)
	for _, m := range agentMarkers {
		if strings.Contains(l, m) {
			return true
		}
	}
	return false
}

// commitInfos reads each commit's subject, author, date and agent co-authors in one git call.
func commitInfos(ctx context.Context, root, base string) (map[string]commitInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", root, "log", "--format=%H%x00%an%x00%ae%x00%as%x00%s%x00%(trailers:key=Co-authored-by,valueonly,separator=%x01)%x1e", base+"..HEAD", "--")
	cmd.Env = repositoryGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("audit: cannot read the commits of %s..HEAD: %v", base, err)
	}
	infos := map[string]commitInfo{}
	for _, rec := range strings.Split(string(out), "\x1e") {
		f := strings.Split(strings.TrimLeft(rec, "\n"), "\x00")
		if len(f) != 6 {
			continue
		}
		info := commitInfo{subject: f[4], author: f[1], date: f[3], agents: []string{}}
		seen := map[string]bool{}
		add := func(name string) {
			name = printableUnit(strings.TrimSpace(name))
			if name != "" && !seen[name] {
				seen[name] = true
				info.agents = append(info.agents, name)
			}
		}
		if agentName(f[1] + " " + f[2]) {
			add(f[1])
		}
		for _, co := range strings.Split(strings.TrimSpace(f[5]), "\x01") {
			if co = strings.TrimSpace(co); co != "" && agentName(co) {
				name, _, _ := strings.Cut(co, "<")
				add(name)
			}
		}
		infos[f[0]] = info
	}
	return infos, nil
}

// uncoveredPaths is the path list of an UNCOVERED reason (CLI-INV-33 wording), without its advice.
func uncoveredPaths(reason string) string {
	paths, _, _ := strings.Cut(strings.TrimPrefix(reason, "claims no spec but changes "), "; ")
	return printableUnit(paths)
}

// driftEvidence is one row per drift entry that is not CURRENT, from the `report drift` lines.
func driftEvidence(out string) []auditEvidence {
	var rows []auditEvidence
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "  ") && len(rows) > 0 {
			rows[len(rows)-1].Text += "; " + strings.TrimSpace(line)
			continue
		}
		if !strings.HasPrefix(line, "drift ") || strings.Contains(line, " state=CURRENT ") {
			continue
		}
		fields := map[string]string{}
		for _, f := range strings.Fields(line) {
			if k, v, ok := strings.Cut(f, "="); ok {
				fields[k] = v
			}
		}
		rows = append(rows, auditEvidence{Commit: fields["verified_commit"], Text: printableUnit(fields["change"] + " is " + fields["state"])})
	}
	for i := range rows {
		rows[i].Text = printableUnit(rows[i].Text)
	}
	return rows
}

// trustPaths are the paths of sha's combined diff that change who may sign or what the check proves
// (C6): anything under .proofbound/keys/ or .github/workflows/, and a gates/ file modified or deleted
// against every parent (an added gate is a review's own output).
func trustPaths(ctx context.Context, root, sha string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", root, "diff-tree", "-r", "--cc", "--no-commit-id", "--name-status", "--no-renames", "--root", "-z", sha, "--")
	cmd.Env = repositoryGitEnv()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("audit: cannot list the paths of %s: %s", sha, strings.TrimSpace(stderr.String()))
	}
	parts := strings.Split(string(raw), "\x00")
	var out []string
	for i := 0; i+1 < len(parts); i += 2 {
		status, path := parts[i], parts[i+1]
		if status == "" {
			break
		}
		switch {
		case strings.HasPrefix(path, ".proofbound/keys/"), strings.HasPrefix(path, ".github/workflows/"), path == ".proofbound/sdd-profile.json":
			out = append(out, path)
		case strings.HasPrefix(path, "gates/") && strings.Trim(status, "A") != "":
			out = append(out, path)
		}
	}
	return out, nil
}

// specInParent reports whether change name's spec folder is in sha's first parent (C7).
func specInParent(ctx context.Context, root, sha, name string) bool {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	return len(parent) == 40 && len(specDirsAt(ctx, root, parent, name)) > 0
}

// reviewedBy reports whether name (a key name) reviewed the warranty body's change.
func reviewedBy(body pbwarranty.Body, name string) bool {
	for _, r := range bodyReviewers(body) {
		if pbwarranty.KeyName(r) == name {
			return true
		}
	}
	return false
}
