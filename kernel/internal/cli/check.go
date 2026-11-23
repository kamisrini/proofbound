package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// The pull-request check (MVP step 2, docs/plans/mvp-2-pr-check-spec.md, CLI-INV-26/27).

const (
	checkPass     = "PASS"
	checkBlocked  = "BLOCKED"
	checkNoReview = "NO REVIEW"
	checkUnbound  = "UNBOUND"
	// checkUncovered is a commit that claims no spec but changes a path outside housekeepingPath
	// (CLI-INV-33): it makes the pull request NOT WARRANTED whatever the claiming commits say.
	checkUncovered = "UNCOVERED"
)

type checkLine struct {
	commit, claims, state, reason, malformed string
	// waivers are the waivers that cover a WAIVED commit (CLI-INV-57).
	waivers []waiverUse
}

// errNotWarranted makes `check` exit 1 after it has printed its report.
var errNotWarranted = errors.New("check: not warranted")

// parseCheck accepts `check <base> [--markdown] [--require-claims]`, each flag at most once.
func parseCheck(args []string) (base string, markdown, requireClaims, ok bool) {
	if len(args) < 2 || args[0] != "check" || args[1] == "" || strings.HasPrefix(args[1], "-") {
		return "", false, false, false
	}
	for _, flag := range args[2:] {
		switch {
		case flag == "--markdown" && !markdown:
			markdown = true
		case flag == "--require-claims" && !requireClaims:
			requireClaims = true
		default:
			return "", false, false, false
		}
	}
	return args[1], markdown, requireClaims, true
}

// checkRange reports, for every commit in base..HEAD that claims an sdd change, whether its
// readiness gate is PASS, using only reads: git, the gate files at HEAD, and the ledger (CHK-1..3).
func checkRange(ctx context.Context, root string, ledger *store.Store, base string, markdown, requireClaims bool, output io.Writer) error {
	lines, err := classifiedRange(ctx, root, ledger, base)
	if err != nil {
		return err
	}
	listed, err := uncoveredCommits(ctx, root, base, lines)
	if err != nil {
		return err
	}
	var uncovered []checkLine
	waived := 0
	for _, l := range listed {
		if l.state == checkWaived {
			waived++
		} else {
			uncovered = append(uncovered, l)
		}
	}
	passed := 0
	for _, l := range lines {
		if l.state == checkPass {
			passed++
		}
	}
	var headline string
	warranted := false
	switch {
	case len(uncovered) > 0 && len(lines) == 0:
		headline = fmt.Sprintf("Proofbound: NOT WARRANTED (no commit claims a spec; %s)", uncoveredPhrase(len(uncovered)))
	case len(uncovered) > 0:
		headline = fmt.Sprintf("Proofbound: NOT WARRANTED (%d of %d claiming commits verified; %s)", passed, len(lines), uncoveredPhrase(len(uncovered)))
	case len(lines) == 0:
		headline = fmt.Sprintf("Proofbound: NO CLAIMS (no commit in %s..HEAD claims a spec)", base)
		warranted = !requireClaims
	case passed == len(lines):
		headline = fmt.Sprintf("Proofbound: WARRANTED (%d of %d claiming commits verified)", passed, len(lines))
		warranted = true
	default:
		headline = fmt.Sprintf("Proofbound: NOT WARRANTED (%d of %d claiming commits verified)", passed, len(lines))
	}
	if waived > 0 {
		// Every waived commit is named in the headline, whatever else it says (CLI-INV-58).
		headline = strings.TrimSuffix(headline, ")") + "; " + waivedPhrase(waived) + ")"
	}
	claimed := len(lines)
	lines = append(lines, listed...)
	var drift bytes.Buffer
	if claimed > 0 {
		if err := reportDrift(ctx, root, ledger, &drift); err != nil {
			return err
		}
	}
	if markdown {
		err = writeCheckMarkdown(output, headline, lines, drift.String())
	} else {
		err = writeCheckPlain(output, headline, lines, drift.String())
	}
	if err != nil {
		return err
	}
	if !warranted {
		return errNotWarranted
	}
	return nil
}

// classifiedRange is every claiming commit in base..HEAD with its CLI-INV-26 state, oldest first.
func classifiedRange(ctx context.Context, root string, ledger *store.Store, base string) ([]checkLine, error) {
	shas, err := commitsInRange(ctx, root, base)
	if err != nil {
		return nil, err
	}
	var lines []checkLine
	for _, sha := range shas {
		names, malformed := claimLines(gitOutput(ctx, root, "log", "-1", "--format=%B", sha))
		if len(names)+len(malformed) > 0 {
			lines = append(lines, checkLine{commit: sha, claims: strings.Join(append(names, malformed...), ","), malformed: strings.Join(malformed, ", ")})
		}
	}
	if len(lines) > 0 {
		if err := classifyClaims(ctx, root, ledger, lines); err != nil {
			return nil, err
		}
	}
	// A claim that does not have the exact form never counts: the commit is UNBOUND, so a typo fails
	// the pull request instead of hiding it as NO CLAIMS (review mvp-2-pr-check-review-20261120 F1).
	for i := range lines {
		if lines[i].malformed != "" {
			lines[i].state = checkUnbound
			lines[i].reason = fmt.Sprintf("malformed claim %s; write exactly `Intent: sdd:<name>` as the message's last line", lines[i].malformed)
		}
	}
	return lines, nil
}

var claimLikeRE = regexp.MustCompile(`(?i)^\s*intent\s*:\s*sdd\s*:`)

// claimLines returns a message's exact `Intent: sdd:<name>` claims as `sdd:<name>`, and, quoted, every
// other line that looks like an sdd claim (any case or spacing, an empty name).
func claimLines(msg string) (names, malformed []string) {
	for _, line := range strings.Split(msg, "\n") {
		t := strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(t, "Intent: sdd:"); ok && v != "" && v == strings.TrimSpace(v) && !strings.ContainsAny(v, " \t") {
			names = append(names, "sdd:"+v)
		} else if claimLikeRE.MatchString(t) {
			malformed = append(malformed, strconv.Quote(t))
		}
	}
	return names, malformed
}

func uncoveredPhrase(n int) string {
	if n == 1 {
		return "1 commit changes code without claiming a spec"
	}
	return fmt.Sprintf("%d commits change code without claiming a spec", n)
}

// uncoveredCommits is every commit in base..HEAD that claims no spec (no claim and no claim-like
// line, so not in claims) yet changes a path that housekeepingPath does not exempt, oldest first
// (CLI-INV-33). A commit's paths are its combined diff against all its parents (`--cc`), so a clean
// merge changes nothing of its own; renames count as a delete plus an add, so both paths are judged.
func uncoveredCommits(ctx context.Context, root, base string, claims []checkLine) ([]checkLine, error) {
	shas, err := commitsInRange(ctx, root, base)
	if err != nil {
		return nil, err
	}
	claiming := map[string]bool{}
	for _, l := range claims {
		claiming[l.commit] = true
	}
	var out []checkLine
	// Who made each commit, read only when a commit needs a waiver (CLI-INV-57): its author and every
	// co-author, and its verified signer (read once for the range).
	var sigs map[string][2]string
	subject := func(sha string) (waiverSubject, error) {
		if sigs == nil {
			var err error
			if sigs, err = commitSignatures(ctx, root, base); err != nil {
				return waiverSubject{}, err
			}
		}
		f := strings.SplitN(gitOutput(ctx, root, "log", "-1", "--format=%an%x00%ae%x00%(trailers:key=Co-authored-by,valueonly,separator=%x01)", "--end-of-options", sha), "\x00", 3)
		if len(f) != 3 {
			return waiverSubject{}, fmt.Errorf("check: cannot read who made %s", sha)
		}
		var coauthors []string
		for _, co := range strings.Split(strings.TrimSpace(f[2]), "\x01") {
			if co = strings.TrimSpace(co); co != "" {
				coauthors = append(coauthors, co)
			}
		}
		return newWaiverSubject(f[0], f[1], coauthors, sigs[sha][1]), nil
	}
	for _, sha := range shas {
		if claiming[sha] {
			continue
		}
		cmd := exec.CommandContext(ctx, "git", "-C", root, "diff-tree", "-r", "--cc", "--no-commit-id", "--name-only", "--no-renames", "--root", "-z", sha, "--")
		cmd.Env = repositoryGitEnv()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		raw, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("check: cannot list the paths of %s: %s", sha, strings.TrimSpace(stderr.String()))
		}
		var code []string
		for _, p := range strings.Split(string(raw), "\x00") {
			if p != "" && !housekeepingPath(p) {
				code = append(code, p)
			}
		}
		if len(code) == 0 {
			continue
		}
		var notes []string
		if strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "--name-only", sha+"^1", "--", waiversDir+"/")) != "" {
			who, err := subject(sha)
			if err != nil {
				return nil, err
			}
			uses, unwaived, why := waiversFor(ctx, root, base, sha, code, who)
			if len(uses) > 0 {
				out = append(out, checkLine{commit: sha, claims: "-", state: checkWaived, reason: waivedReason(uses), waivers: uses})
				continue
			}
			code, notes = unwaived, why
		}
		shown := code
		if len(shown) > 3 {
			shown = shown[:3]
		}
		reason := "claims no spec but changes " + strings.Join(shown, ", ")
		if len(code) > len(shown) {
			reason += fmt.Sprintf(" (+%d more)", len(code)-len(shown))
		}
		reason += "; run `proofbound task start <change>` before committing code, or end the message with `Intent: sdd:<change>`"
		if len(notes) > 0 {
			reason += "; " + strings.Join(notes, "; ")
		}
		out = append(out, checkLine{commit: sha, claims: "-", state: checkUncovered, reason: reason})
	}
	return out, nil
}

// housekeepingPath reports whether a commit may change path without claiming a spec (CLI-INV-33): the
// specs themselves, gates, Proofbound's own files and review records, documentation, and the files
// `init` writes. Everything else is code. The pull-request workflow is exempt because the check cannot
// guard its own workflow: GitHub runs the copy on the pull request's branch.
func housekeepingPath(path string) bool {
	for _, dir := range []string{"openspec/", "specs/", "gates/", ".proofbound/", "docs/"} {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	switch path {
	case ".gitignore", ".github/workflows/proofbound.yml":
		return true
	}
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

func commitsInRange(ctx context.Context, root, base string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", root, "rev-list", "--reverse", base+"..HEAD", "--")
	cmd.Env = repositoryGitEnv()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("check: cannot list %s..HEAD: %s", base, strings.TrimSpace(stderr.String()))
	}
	return strings.Fields(string(out)), nil
}

// classifyClaims sets each line's state: UNBOUND unless the ledger holds the commit with a bound
// sdd ref; else NO REVIEW without its readiness gate at HEAD; else the gate's PASS or BLOCKED.
func classifyClaims(ctx context.Context, root string, ledger *store.Store, lines []checkLine) error {
	want := map[string]bool{}
	for _, l := range lines {
		want[l.commit] = true
	}
	seen, bound := map[string]bool{}, map[string]bool{}
	err := ledger.ReadEvents(ctx, store.Filter{Source: core.SourceGit, Kind: core.KindCommitRecorded}, func(record store.Record) error {
		if !want[record.Event.NativeID] {
			return nil
		}
		var commit connectorgit.Commit
		if err := json.Unmarshal(record.Event.Payload, &commit); err != nil {
			return err
		}
		seen[record.Event.NativeID] = true
		for _, ref := range commit.IntentRefs {
			bound[record.Event.NativeID] = bound[record.Event.NativeID] || ref.Provider == "sdd"
		}
		return nil
	})
	if err != nil {
		return err
	}
	definitions, err := gates.LoadDir(filepath.Join(root, "gates"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for i := range lines {
		l := &lines[i]
		short := l.commit[:12]
		if !bound[l.commit] {
			l.state = checkUnbound
			if seen[l.commit] {
				l.reason = "the claim did not bind to a committed spec (see `proofbound sync git` output)"
			} else {
				l.reason = "commit not in the ledger; run `proofbound sync all`"
			}
			continue
		}
		var gate *gates.Definition
		for j := range definitions {
			d := &definitions[j]
			if d.ID == "ready-"+short && d.Rule == "intent-delivery-readiness" && d.ScopeCommit == l.commit {
				gate = d
			}
		}
		if gate == nil {
			l.state = checkNoReview
			l.reason = fmt.Sprintf("no readiness gate gates/ready-%s.yaml; a reviewer runs `proofbound review draft %s`", short, short)
			continue
		}
		result, err := gates.Evaluate(ctx, ledger, *gate)
		if err != nil {
			return err
		}
		if result.State == gates.StatePass {
			l.state = checkPass
		} else {
			l.state = checkBlocked
		}
		l.reason = result.Reason
		if l.reason == "" {
			l.reason = "gate state " + string(result.State)
		}
	}
	return nil
}

func writeCheckPlain(w io.Writer, headline string, lines []checkLine, drift string) error {
	var b strings.Builder
	b.WriteString(headline + "\n")
	for _, l := range lines {
		fmt.Fprintf(&b, "  %s %s %s  %s\n", l.commit[:12], l.claims, l.state, printableUnit(l.reason))
	}
	if drift != "" {
		b.WriteString("Drift:\n")
		for _, line := range strings.Split(strings.TrimRight(drift, "\n"), "\n") {
			b.WriteString("  " + line + "\n")
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeCheckMarkdown(w io.Writer, headline string, lines []checkLine, drift string) error {
	cell := mdCell
	var b strings.Builder
	b.WriteString("### " + headline + "\n\n")
	if len(lines) > 0 {
		b.WriteString("| Commit | Claims | State | Why |\n|---|---|---|---|\n")
		for _, l := range lines {
			fmt.Fprintf(&b, "| `%s` | `%s` | **%s** | %s |\n", l.commit[:12], cell(l.claims), l.state, cell(l.reason))
		}
	}
	if drift != "" {
		b.WriteString("\n<details><summary>Promise drift</summary>\n\n```\n" + strings.ReplaceAll(drift, "```", "'''") + "```\n</details>\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// mdCell escapes untrusted text (commit messages, gate reasons, artifact fields) for one Markdown table
// cell: no pipe, newline or backtick can break the table or open a code span (CLI-INV-26, CLI-INV-28).
func mdCell(s string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ", "`", "'").Replace(printableUnit(s))
}
