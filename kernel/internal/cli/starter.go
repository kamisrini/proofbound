package cli

import (
	"context"
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The R2 starter kit and reviewer workflow (docs/plans/R2-starter-kit-spec.md, CLI-INV-18..20).

//go:embed starterkit/agents-block.md
var agentsBlock string

//go:embed starterkit/reviewer.md
var reviewerAgent string

//go:embed starterkit/commit-msg
var commitMsgHook string

//go:embed starterkit/prepare-commit-msg
var prepareCommitMsgHook string

//go:embed starterkit/proofbound.yml
var prCheckWorkflow string

const agentsBegin = "<!-- proofbound:begin -->"

// initKit writes the starter kit for kit ("openspec" or "speckit") into root. It never overwrites a
// file; the .gitignore lines and the AGENTS.md block are appended only when missing, so a second run
// changes nothing (CLI-INV-18).
func initKit(ctx context.Context, root, kit, owner string, output io.Writer) error {
	if strings.TrimSpace(owner) == "" {
		owner = strings.TrimSpace(gitOutput(ctx, root, "config", "user.name"))
	}
	if owner == "" {
		owner = "project-owner"
	}
	quoted, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	var profile string
	switch kit {
	case "openspec":
		profile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": ` + string(quoted) + ` } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": ` + string(quoted) + ` } ]
}
`
	case "speckit":
		profile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"] },
  "changes": { "mode": "feature-sibling", "root": "specs", "sibling": "tasks.md", "status": "accepted" },
  "owners":   [ { "pattern": "specs/**", "owner": ` + string(quoted) + ` } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": ` + string(quoted) + ` } ]
}
`
	default:
		return fmt.Errorf("init: unknown kit %q (openspec or speckit)", kit)
	}
	hooks := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--git-path", "hooks"))
	if hooks == "" {
		return errors.New("init: cannot locate the git hooks directory")
	}
	if !filepath.IsAbs(hooks) {
		hooks = filepath.Join(root, hooks)
	}
	steps := []func() error{
		func() error { return writeNew(output, root, ".proofbound/sdd-profile.json", profile, 0o644) },
		func() error {
			return appendMissingLines(output, root, ".gitignore", []string{".proofbound/*", "!.proofbound/sdd-profile.json", "!.proofbound/keys/", "!.proofbound/waivers/", "!.proofbound/tests.json", "!" + agentConfigPath})
		},
		func() error { return appendBlock(output, root, "AGENTS.md", agentsBlock) },
		func() error {
			return writeNew(output, root, ".claude/agents/proofbound-reviewer.md", reviewerAgent, 0o644)
		},
		func() error { return writeClaudeSettings(output, root) },
		func() error { return writeNew(output, hooks, "commit-msg", commitMsgHook, 0o755) },
		func() error { return writeNew(output, hooks, "prepare-commit-msg", prepareCommitMsgHook, 0o755) },
		func() error {
			return writeNew(output, root, ".github/workflows/proofbound.yml", prCheckWorkflow, 0o644)
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func gitOutput(ctx context.Context, root string, args ...string) string {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	cmd.Env = repositoryGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func writeNew(output io.Writer, dir, rel, content string, mode os.FileMode) error {
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Lstat(path); err == nil {
		_, err = fmt.Fprintf(output, "exists %s\n", path)
		return err
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return err
	}
	_, err := fmt.Fprintf(output, "wrote %s\n", path)
	return err
}

func appendMissingLines(output io.Writer, root, rel string, lines []string) error {
	path := filepath.Join(root, rel)
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	have := map[string]bool{}
	for _, line := range strings.Split(string(existing), "\n") {
		have[strings.TrimSpace(line)] = true
	}
	var add []string
	for _, line := range lines {
		if !have[line] {
			add = append(add, line)
		}
	}
	if len(add) == 0 {
		_, err = fmt.Fprintf(output, "unchanged %s\n", path)
		return err
	}
	return appendText(output, path, existing, strings.Join(add, "\n")+"\n")
}

func appendBlock(output io.Writer, root, rel, block string) error {
	path := filepath.Join(root, rel)
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if strings.Contains(string(existing), agentsBegin) {
		_, err = fmt.Fprintf(output, "unchanged %s\n", path)
		return err
	}
	return appendText(output, path, existing, block)
}

func appendText(output io.Writer, path string, existing []byte, text string) error {
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		text = "\n" + text
	}
	if len(existing) > 0 {
		text = "\n" + text
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		return errors.Join(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "appended %s\n", path)
	return err
}

// draftFile is what `review draft` writes and the reviewer edits. Only the outcome, finding and
// declared_reviewer fields are the reviewer's; seal rebuilds everything else from them.
type draftFile struct {
	Commit           string         `json:"commit"`
	DeclaredReviewer string         `json:"declared_reviewer"`
	Reviews          []draftReview  `json:"reviews"`
	Verdicts         []draftVerdict `json:"verdicts"`
	// OutOfArea are the lines the commit changed outside its spec's declared area, each to be
	// accepted (with a reason) or rejected (change-area-spec.md §3). Built from git, never trusted.
	OutOfArea []areaDecision `json:"out_of_area"`
	// Tests is what undoing each code hunk did to the declared tests (tests-can-fail-spec.md); absent
	// when the project declares no tests. Seal runs them again and takes only the decisions.
	Tests *testsResult `json:"tests,omitempty"`
	// Provenance is which recorded agent sessions wrote the lines the commit adds
	// (session-provenance-spec.md); informational: seal matches again and trusts none of it.
	Provenance *provResult `json:"provenance,omitempty"`
	// Agent is the agent reviewer's checked answer (agent-reviewer-spec.md); advice only: it never
	// decides a question, and the seal records it with the person's differences (AGENT- record).
	Agent *agentReview `json:"agent,omitempty"`
}

type draftReview struct {
	Requirement   connectorintent.Reference `json:"requirement"`
	DeclaredOwner string                    `json:"declared_owner"`
	Obligations   []draftReviewObligation   `json:"obligations"`
}

type draftReviewObligation struct {
	ObligationID string `json:"obligation_id"`
	Statement    string `json:"statement"`
	Outcome      string `json:"outcome"`
	Finding      string `json:"finding"`
}

type draftVerdict struct {
	ChangeIntent connectorintent.Reference   `json:"change_intent"`
	Requirements []connectorintent.Reference `json:"requirements"`
	Obligations  []draftVerdictObligation    `json:"obligations"`
}

type draftVerdictObligation struct {
	Source         string   `json:"source"`
	RequirementID  string   `json:"requirement_id"`
	ArtifactSHA256 string   `json:"artifact_sha256"`
	ObligationID   string   `json:"obligation_id"`
	Statement      string   `json:"statement"`
	Outcome        string   `json:"outcome"`
	EvidenceRefs   []string `json:"evidence_refs"`
}

func draftDir(root, sha string) string { return filepath.Join(root, ".proofbound", "drafts", sha) }

func resolveCommit(ctx context.Context, root, arg string) (string, error) {
	sha := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--end-of-options", arg+"^{commit}"))
	if len(sha) != 40 {
		return "", fmt.Errorf("%q is not a commit", arg)
	}
	return sha, nil
}

// reviewDraft writes .proofbound/drafts/<sha>/draft.json for every sdd change the commit's latest
// ledgered git event binds, with every obligation statement, and prints them.
func reviewDraft(ctx context.Context, root string, ledger *store.Store, arg string, output io.Writer) error {
	sha, err := resolveCommit(ctx, root, arg)
	if err != nil {
		return err
	}
	draft, err := ledgerDraft(ctx, ledger, sha)
	if err != nil {
		return err
	}
	draft, area, err := withArea(ctx, root, draft)
	if err != nil {
		return err
	}
	tests, err := draftTests(ctx, root, sha, output)
	if err != nil {
		return err
	}
	if tests.Declared {
		draft.Tests = &tests
	}
	prov, err := commitProvenance(ctx, root, sha)
	if err != nil {
		return err
	}
	draft.Provenance = &prov
	dir := draftDir(root, sha)
	path := filepath.Join(dir, "draft.json")
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s exists; edit it, or delete it to start over", path)
	}
	body, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o644); err != nil {
		return err
	}
	for _, v := range draft.Verdicts {
		for _, o := range v.Obligations {
			if _, err := fmt.Fprintf(output, "obligation %s %s: %s\n", v.ChangeIntent.RecordID, o.ObligationID, printableUnit(o.Statement)); err != nil {
				return err
			}
		}
	}
	if err := writeAreaLines(output, area); err != nil {
		return err
	}
	if err := writeTestsLines(output, tests); err != nil {
		return err
	}
	if err := writeProvLines(output, prov); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "draft %s\n", path)
	return err
}

// ledgerDraft builds a commit's undecided draft from the ledger alone: the commit's latest git event,
// its bound sdd change intents, their target requirement revisions and owners, and the git event as
// evidence. Seal rebuilds it too, so nothing but the reviewer's decisions comes from the editable
// draft (review r2-starter-kit-review-20261119 F1).
func ledgerDraft(ctx context.Context, ledger *store.Store, sha string) (draftFile, error) {
	var gitEventID string
	var commit connectorgit.Commit
	changeIntents := map[string]connectorintent.ChangeIntent{}
	requirements := map[string]connectorintent.Requirement{}
	err := ledger.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		e := record.Event
		switch {
		case e.Source == core.SourceGit && e.Kind == core.KindCommitRecorded && e.NativeID == sha:
			commit = connectorgit.Commit{}
			if err := json.Unmarshal(e.Payload, &commit); err != nil {
				return fmt.Errorf("git event seq %d: %w", record.Seq, err)
			}
			gitEventID = e.ID.String()
		case e.Source == core.SourceIntentSDD && e.Kind == core.KindChangeIntent:
			var ci connectorintent.ChangeIntent
			if err := json.Unmarshal(e.Payload, &ci); err != nil {
				return fmt.Errorf("change intent seq %d: %w", record.Seq, err)
			}
			changeIntents[ci.IntentID+"\x00"+ci.ArtifactSHA256] = ci
		case e.Kind == core.KindRequirement && strings.HasPrefix(string(e.Source), "intent."):
			var r connectorintent.Requirement
			if err := json.Unmarshal(e.Payload, &r); err != nil {
				return fmt.Errorf("requirement seq %d: %w", record.Seq, err)
			}
			requirements[string(e.Source)+"\x00"+r.RequirementID+"\x00"+r.ArtifactSHA256] = r
		}
		return nil
	})
	if err != nil {
		return draftFile{}, err
	}
	if gitEventID == "" {
		return draftFile{}, fmt.Errorf("commit %s is not in the ledger; run `proofbound sync all` first", sha)
	}
	draft := draftFile{Commit: sha, Reviews: []draftReview{}, Verdicts: []draftVerdict{}}
	reviewed := map[string]bool{}
	for _, ref := range commit.IntentRefs {
		if ref.Provider != "sdd" {
			continue
		}
		ci, ok := changeIntents[ref.RecordID+"\x00"+ref.ArtifactSHA256]
		if !ok {
			return draftFile{}, fmt.Errorf("commit %s binds %s, which is not in the ledger", sha, ref.RecordID)
		}
		verdict := draftVerdict{ChangeIntent: connectorintent.Reference{RecordKind: "change_intent", Source: string(core.SourceIntentSDD), RecordID: ci.IntentID, ArtifactSHA256: ci.ArtifactSHA256, Relation: "evaluates"}}
		for _, target := range ci.Targets {
			key := target.Source + "\x00" + target.RecordID + "\x00" + target.ArtifactSHA256
			requirement, ok := requirements[key]
			if !ok {
				return draftFile{}, fmt.Errorf("change intent %s targets %s, which is not in the ledger", ci.IntentID, target.RecordID)
			}
			statements := map[string]string{}
			for _, o := range requirement.Obligations {
				statements[o.ID] = o.Statement
			}
			ref := connectorintent.Reference{RecordKind: "requirement", Source: target.Source, RecordID: target.RecordID, ArtifactSHA256: target.ArtifactSHA256, Relation: "evaluates"}
			verdict.Requirements = append(verdict.Requirements, ref)
			for _, id := range target.ObligationIDs {
				verdict.Obligations = append(verdict.Obligations, draftVerdictObligation{Source: target.Source, RequirementID: target.RecordID, ArtifactSHA256: target.ArtifactSHA256, ObligationID: id, Statement: statements[id], EvidenceRefs: []string{connectorreviews.CommitEvidenceRef(sha)}})
			}
			if !reviewed[key] {
				reviewed[key] = true
				review := draftReview{Requirement: ref, DeclaredOwner: requirement.DeclaredOwner}
				review.Requirement.Relation = "reviews"
				for _, o := range requirement.Obligations {
					review.Obligations = append(review.Obligations, draftReviewObligation{ObligationID: o.ID, Statement: o.Statement})
				}
				draft.Reviews = append(draft.Reviews, review)
			}
		}
		draft.Verdicts = append(draft.Verdicts, verdict)
	}
	if len(draft.Verdicts) == 0 {
		return draftFile{}, fmt.Errorf("commit %s binds no spec change: does its message end with `Intent: sdd:<name>`, and has `proofbound sync all` run since?", sha)
	}
	return draft, nil
}

// reviewSeal turns an edited draft into committed-ready review and verdict artifacts plus the
// reviewed commit's readiness gate. It validates everything, and re-parses each artifact with the
// reviews connector's own parser, before writing anything (CLI-INV-19, CLI-INV-20).
func reviewSeal(ctx context.Context, root, dir string, fresh func(context.Context, string) (draftFile, error), output io.Writer) error {
	return reviewSealWith(ctx, root, dir, nil, fresh, output)
}

// sealTestsRun is a test run made by the caller before the record was opened (ui-reviewer-spec §1.1,
// review F2): the tests may take long, and nothing should hold the record meanwhile.
type sealTestsRun struct {
	Commit string
	Result testsResult
}

// readSealDraft resolves a seal argument (the draft folder or its draft.json) and reads the draft.
func readSealDraft(dir string) (string, draftFile, error) {
	if !filepath.IsAbs(dir) {
		wd, err := workingDirectory()
		if err != nil {
			return dir, draftFile{}, err
		}
		dir = filepath.Join(wd, dir)
	}
	// `review draft` prints the draft.json path, so seal accepts it as well as its folder (CLI-INV-24).
	if info, err := os.Stat(dir); err == nil && !info.IsDir() && filepath.Base(dir) == "draft.json" {
		dir = filepath.Dir(dir)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "draft.json"))
	if err != nil {
		return dir, draftFile{}, err
	}
	var edited draftFile
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&edited); err != nil {
		return dir, draftFile{}, fmt.Errorf("draft.json: %w", err)
	}
	return dir, edited, nil
}

// reviewSealWith seals; ran, when given, is the caller's fresh test run of the draft's commit, made
// just before this call, and replaces the run seal would make itself.
func reviewSealWith(ctx context.Context, root, dir string, ran *sealTestsRun, fresh func(context.Context, string) (draftFile, error), output io.Writer) error {
	_, edited, err := readSealDraft(dir)
	if err != nil {
		return err
	}
	reviewer := strings.TrimSpace(edited.DeclaredReviewer)
	if reviewer == "" || strings.ContainsAny(reviewer, "\n\r") {
		return errors.New("seal: set declared_reviewer to your name (one line)")
	}
	if len(edited.Commit) != 40 {
		return errors.New("seal: draft.json has no commit; regenerate it with `proofbound review draft`")
	}
	// Everything but the reviewer's decisions comes from the ledger, never from the editable draft:
	// an edited reference, digest, owner, evidence id or obligation set would otherwise produce
	// artifacts that ledger fine and then fail projection on every later gate run (F1).
	truth, err := fresh(ctx, edited.Commit)
	if err != nil {
		return fmt.Errorf("seal: %w", err)
	}
	draft, err := mergeDecisions(truth, edited)
	if err != nil {
		return err
	}
	area, err := commitArea(ctx, root, draft.Commit)
	if err != nil {
		return err
	}
	areaDecisions, areaRejected, err := decideArea(area, edited.OutOfArea)
	if err != nil {
		return err
	}
	var run testsResult
	switch {
	case ran == nil:
		if run, err = draftTests(ctx, root, draft.Commit, output); err != nil {
			return err
		}
	case ran.Commit != draft.Commit:
		return errors.New("seal: the test run is for another commit; nothing written")
	default:
		run = ran.Result
	}
	tests, testsNeedWork, err := decideTests(run, edited.Tests)
	if err != nil {
		return err
	}
	// Provenance is matched again here, never read from the draft (spec §1.3).
	prov, err := commitProvenance(ctx, root, draft.Commit)
	if err != nil {
		return err
	}
	// The agent's answer is checked against its digest and this commit's questions (spec §4): the
	// questions as the bench asked them, with the parts outside the area rebuilt from git (the
	// ledger's draft holds none) and the draft's test run (live run 2026-11-25).
	asked := draft
	asked.OutOfArea = areaDecisions
	// An answer edited after the agent gave it is refused; one that no longer fits the questions (the
	// tests ran again) is left out with a line saying so, never sealed (code review F4).
	agentNote := ""
	if err := checkDraftAgent(draft.Agent, asked); errors.Is(err, errAgentStale) {
		draft.Agent, asked.Agent = nil, nil
		agentNote = "agent: its answer no longer fits this commit's questions, so it is not recorded; ask the agent again to record one\n"
	} else if err != nil {
		return err
	}
	signer, err := reviewSigner(root, reviewer)
	if err != nil {
		return err
	}
	short := draft.Commit[:12]
	files := map[string][]byte{}
	var order []string
	add := func(rel string, data []byte) error {
		if _, ok := files[rel]; ok {
			return fmt.Errorf("seal: two artifacts would be written to %s", rel)
		}
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
			return fmt.Errorf("seal: %s exists; nothing written", rel)
		}
		files[rel] = data
		order = append(order, rel)
		return nil
	}
	// MVP step 3, option A: each sealed artifact gets a `.sig` by the reviewer's declared key, written
	// in the same all-or-nothing batch (CLI-INV-19/20 amended).
	addSig := func(path string, data []byte) error {
		if signer == nil {
			return nil
		}
		return add(path+".sig", []byte(pbwarranty.SignArtifact(data, signer)))
	}
	for i, review := range draft.Reviews {
		if reviewer == review.DeclaredOwner || pbwarranty.KeyName(reviewer) == pbwarranty.KeyName(review.DeclaredOwner) {
			return fmt.Errorf("seal: the reviewer %q is the declared owner of %s; a different reviewer must review it", reviewer, review.Requirement.RecordID)
		}
		outcomes := []connectorreviews.RequirementReviewOutcome{}
		for _, o := range review.Obligations {
			switch o.Outcome {
			case "VERIFIABLE", "AMBIGUOUS", "UNTESTABLE", "CONTRADICTORY":
			default:
				return fmt.Errorf("seal: review of %s obligation %s has outcome %q; use VERIFIABLE, AMBIGUOUS, UNTESTABLE or CONTRADICTORY", review.Requirement.RecordID, o.ObligationID, o.Outcome)
			}
			outcomes = append(outcomes, connectorreviews.RequirementReviewOutcome{ObligationID: o.ObligationID, Outcome: o.Outcome, Finding: strings.TrimSpace(o.Finding)})
		}
		sort.Slice(outcomes, func(a, b int) bool { return outcomes[a].ObligationID < outcomes[b].ObligationID })
		id := fmt.Sprintf("RR-%s-%d", short, i+1)
		path := "docs/verification/verdicts/" + id + ".md"
		data, err := sealArtifact(path, func(digest string) any {
			return connectorreviews.RequirementReview{Schema: "proofbound.requirement-review.v1", ReviewID: id, Requirement: review.Requirement, DeclaredReviewer: reviewer, Outcomes: outcomes, ArtifactPath: path, ArtifactSHA256: digest}
		}, func(data []byte) error { _, err := connectorreviews.ParseRequirementReview(path, data); return err }, draft.Commit, reviewer)
		if err != nil {
			return err
		}
		if err := add(path, data); err != nil {
			return err
		}
		if err := addSig(path, data); err != nil {
			return err
		}
	}
	for i, verdict := range draft.Verdicts {
		obligations := []connectorreviews.ObligationOutcome{}
		status := "ACCEPTABLE"
		if areaRejected || testsNeedWork {
			// A rejected out-of-area change fails the review of the whole commit (§3).
			status = "NEEDS_WORK"
		}
		for _, o := range verdict.Obligations {
			switch o.Outcome {
			case "SATISFIED":
				if len(o.EvidenceRefs) == 0 {
					return fmt.Errorf("seal: %s obligation %s is SATISFIED with no evidence", verdict.ChangeIntent.RecordID, o.ObligationID)
				}
			case "NOT_SATISFIED", "INCONCLUSIVE":
				status = "NEEDS_WORK"
			default:
				return fmt.Errorf("seal: %s obligation %s has outcome %q; use SATISFIED, NOT_SATISFIED or INCONCLUSIVE", verdict.ChangeIntent.RecordID, o.ObligationID, o.Outcome)
			}
			// Portable evidence (CLI-INV-20 amended): the reviewed commit by its SHA, never a ledger-local
			// event id, so the verdict projects on a fresh CI ledger.
			evidence := append([]string{}, o.EvidenceRefs...)
			sort.Strings(evidence)
			obligations = append(obligations, connectorreviews.ObligationOutcome{Source: o.Source, RequirementID: o.RequirementID, ArtifactSHA256: o.ArtifactSHA256, ObligationID: o.ObligationID, Outcome: o.Outcome, EvidenceEventIDs: []string{}, EvidenceRefs: evidence})
		}
		sort.Slice(obligations, func(a, b int) bool {
			return outcomeKey(obligations[a]) < outcomeKey(obligations[b])
		})
		requirements := append([]connectorintent.Reference{}, verdict.Requirements...)
		sort.Slice(requirements, func(a, b int) bool { return referenceSortKey(requirements[a]) < referenceSortKey(requirements[b]) })
		id := fmt.Sprintf("OV-%s-%d", short, i+1)
		path := "docs/verification/verdicts/" + id + ".md"
		data, err := sealArtifact(path, func(digest string) any {
			return connectorreviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", VerdictID: id, Status: status, DeclaredReviewer: reviewer, ReviewedCommit: draft.Commit, ChangeIntent: verdict.ChangeIntent, Requirements: requirements, Obligations: obligations, Findings: []connectorreviews.Finding{}, ArtifactPath: path, ArtifactSHA256: digest}
		}, func(data []byte) error { _, err := connectorreviews.ParseObligationVerdict(path, data); return err }, draft.Commit, reviewer)
		if err != nil {
			return err
		}
		if err := add(path, data); err != nil {
			return err
		}
		if err := addSig(path, data); err != nil {
			return err
		}
	}
	if len(areaDecisions) > 0 {
		data, err := areaRecordBytes(draft.Commit, pbwarranty.KeyName(reviewer), areaDecisions, signer, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return err
		}
		if err := add(areaRecordPath(draft.Commit), data); err != nil {
			return err
		}
	}
	if tests.Declared {
		data, err := testsRecordBytes(draft.Commit, pbwarranty.KeyName(reviewer), tests, signer, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return err
		}
		if err := add(testsRecordPath(draft.Commit), data); err != nil {
			return err
		}
	}
	// PROV is written only when the reviewer signs: an unsigned record never verifies, and C12 would
	// turn every keyless seal into an exception (review F1).
	if signer != nil {
		data, err := provRecordBytes(draft.Commit, pbwarranty.KeyName(reviewer), prov, signer, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return err
		}
		if err := add(provRecordPath(draft.Commit), data); err != nil {
			return err
		}
	}
	// Like PROV, the AGENT record is written only when the reviewer signs (spec review F9).
	if draft.Agent != nil && signer != nil {
		mine := asked
		mine.Tests = &tests
		data, err := agentRecordBytes(draft.Commit, pbwarranty.KeyName(reviewer), draft.Agent, agentDifferences(mine), signer, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return err
		}
		if err := add(agentRecordPath(draft.Commit), data); err != nil {
			return err
		}
	}
	gate := fmt.Sprintf(`{"schema":"proofbound.gate.v1","id":"ready-%s","description":"readiness of commit %s","expires":"2099-01-01","mode":"canary","rule":"intent-delivery-readiness","scope_commit":"%s"}`+"\n", short, short, draft.Commit)
	if err := add("gates/ready-"+short+".yaml", []byte(gate)); err != nil {
		return err
	}
	var written []string
	undo := func(err error) error {
		for _, path := range written {
			_ = os.Remove(path)
		}
		return err
	}
	for _, rel := range order {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return undo(err)
		}
		// O_EXCL: never follow or truncate a file that appeared after the existence check.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return undo(err)
		}
		written = append(written, path)
		if _, err := f.Write(files[rel]); err != nil {
			return undo(errors.Join(err, f.Close()))
		}
		if err := f.Close(); err != nil {
			return undo(err)
		}
	}
	for _, rel := range order {
		if _, err := fmt.Fprintf(output, "wrote %s\n", rel); err != nil {
			return err
		}
	}
	if agentNote != "" {
		if _, err := io.WriteString(output, agentNote); err != nil {
			return err
		}
	}
	keyFile := pbwarranty.KeysDir + "/" + pbwarranty.KeyName(reviewer) + ".pub"
	if signer == nil {
		_, err = fmt.Fprintf(output, "unsigned: no key declared for %s (%s); warranties for this commit will verify UNREVIEWED until the reviewer cosigns\n", reviewer, keyFile)
	} else {
		_, err = fmt.Fprintf(output, "signed as %s (key %s); commit %s with the review if it is not committed yet\n", reviewer, pbwarranty.Fingerprint(signer.Public().(ed25519.PublicKey)), keyFile)
	}
	return err
}

// reviewSigner decides how `review seal` signs (mvp-3 spec §3): nil when no key is declared for the
// reviewer in the working tree; the local key when it is exactly the declared one; an error (nothing
// written) for an unusable name, a declaration that is not an exact key line, or a declared key this
// machine does not hold.
func reviewSigner(root, reviewer string) (ed25519.PrivateKey, error) {
	name := pbwarranty.KeyName(reviewer)
	if name == "" {
		return nil, fmt.Errorf("seal: reviewer %q has no usable key name; use letters, digits or '-'", reviewer)
	}
	pubPath := filepath.Join(root, filepath.FromSlash(pbwarranty.KeysDir), name+".pub")
	info, err := os.Lstat(pubPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("seal: %s is not a regular file", pubPath)
	}
	declared, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, err
	}
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return nil, err
	}
	key, err := pbwarranty.LoadKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("seal: a key is declared for %s but this machine cannot use one: %v", reviewer, err)
	}
	if pbwarranty.PubLine(key.Public().(ed25519.PublicKey)) != string(declared) {
		return nil, fmt.Errorf("seal: this machine's key is not the key declared for %s in %s (set PROOFBOUND_KEY to %s's key)", reviewer, pubPath, reviewer)
	}
	return key, nil
}

// mergeDecisions copies the reviewer's decisions (outcome, finding) from the edited draft onto the
// ledger's draft. Every ledger obligation must be decided, and the edited draft may hold nothing
// the ledger does not (an edited digest, id or extra obligation is a mismatch, not a new fact).
func mergeDecisions(truth, edited draftFile) (draftFile, error) {
	type decision struct{ outcome, finding string }
	reviews := map[string]decision{}
	for _, r := range edited.Reviews {
		for _, o := range r.Obligations {
			reviews[referenceSortKey(r.Requirement)+"\x00"+o.ObligationID] = decision{o.Outcome, o.Finding}
		}
	}
	verdicts := map[string]decision{}
	for _, v := range edited.Verdicts {
		for _, o := range v.Obligations {
			verdicts[referenceSortKey(v.ChangeIntent)+"\x00"+o.Source+"\x00"+o.RequirementID+"\x00"+o.ArtifactSHA256+"\x00"+o.ObligationID] = decision{o.Outcome, ""}
		}
	}
	mismatch := func(what string) error {
		return fmt.Errorf("seal: draft.json does not match the ledger (%s); delete the draft and run `proofbound review draft` again", what)
	}
	used := 0
	for i := range truth.Reviews {
		r := &truth.Reviews[i]
		for j := range r.Obligations {
			d, ok := reviews[referenceSortKey(r.Requirement)+"\x00"+r.Obligations[j].ObligationID]
			if !ok {
				return draftFile{}, mismatch("review of " + r.Requirement.RecordID + " " + r.Obligations[j].ObligationID + " is missing")
			}
			r.Obligations[j].Outcome, r.Obligations[j].Finding = d.outcome, d.finding
			used++
		}
	}
	if used != len(reviews) {
		return draftFile{}, mismatch("the reviews hold obligations the ledger does not")
	}
	used = 0
	for i := range truth.Verdicts {
		v := &truth.Verdicts[i]
		for j := range v.Obligations {
			o := &v.Obligations[j]
			d, ok := verdicts[referenceSortKey(v.ChangeIntent)+"\x00"+o.Source+"\x00"+o.RequirementID+"\x00"+o.ArtifactSHA256+"\x00"+o.ObligationID]
			if !ok {
				return draftFile{}, mismatch("verdict obligation " + o.RequirementID + " " + o.ObligationID + " is missing")
			}
			o.Outcome = d.outcome
			used++
		}
	}
	if used != len(verdicts) {
		return draftFile{}, mismatch("the verdicts hold obligations the ledger does not")
	}
	// Area decisions are matched by path and lines; the list itself is rebuilt from git (§3).
	given := map[string]areaDecision{}
	for _, d := range edited.OutOfArea {
		given[areaHunk{Path: d.Path, Lines: d.Lines}.key()] = d
	}
	for i := range truth.OutOfArea {
		if d, ok := given[areaHunk{Path: truth.OutOfArea[i].Path, Lines: truth.OutOfArea[i].Lines}.key()]; ok {
			truth.OutOfArea[i].Decision, truth.OutOfArea[i].Reason = d.Decision, d.Reason
		}
	}
	// The test run is never in the record; the draft's run and its decisions are carried so the
	// Review bench shows them (ui-reviewer-spec §1.1, review F1). Seal runs the tests again.
	truth.Tests = edited.Tests
	truth.Agent = edited.Agent
	truth.DeclaredReviewer = edited.DeclaredReviewer
	return truth, nil
}

// sealArtifact renders one artifact with its self-digest and proves the reviews connector accepts it.
func sealArtifact(path string, build func(digest string) any, parse func([]byte) error, commit, reviewer string) ([]byte, error) {
	render := func(digest string) ([]byte, error) {
		front, err := json.Marshal(build(digest))
		if err != nil {
			return nil, err
		}
		body := fmt.Sprintf("\nReview of commit %s by %s, written by `proofbound review seal`.\n", commit, reviewer)
		return []byte("---\n" + string(front) + "\n---\n" + body), nil
	}
	draft, err := render(strings.Repeat("0", 64))
	if err != nil {
		return nil, err
	}
	data, err := render(connectorreviews.ArtifactSHA256(draft))
	if err != nil {
		return nil, err
	}
	if err := parse(data); err != nil {
		return nil, fmt.Errorf("seal: %s would not be accepted: %w", path, err)
	}
	return data, nil
}

func outcomeKey(o connectorreviews.ObligationOutcome) string {
	return o.Source + "\x00" + o.RequirementID + "\x00" + o.ArtifactSHA256 + "\x00" + o.ObligationID
}

func referenceSortKey(r connectorintent.Reference) string {
	return r.RecordKind + "\x00" + r.Source + "\x00" + r.RecordID + "\x00" + r.ArtifactSHA256 + "\x00" + r.Relation
}
