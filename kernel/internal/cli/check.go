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
)

type checkLine struct{ commit, claims, state, reason string }

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
	shas, err := commitsInRange(ctx, root, base)
	if err != nil {
		return err
	}
	var lines []checkLine
	for _, sha := range shas {
		var names []string
		for _, line := range strings.Split(gitOutput(ctx, root, "log", "-1", "--format=%B", sha), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Intent: sdd:"); ok {
				names = append(names, "sdd:"+strings.TrimSpace(v))
			}
		}
		if len(names) > 0 {
			lines = append(lines, checkLine{commit: sha, claims: strings.Join(names, ",")})
		}
	}
	if len(lines) > 0 {
		if err := classifyClaims(ctx, root, ledger, lines); err != nil {
			return err
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
	case len(lines) == 0:
		headline = fmt.Sprintf("Proofbound: NO CLAIMS (no commit in %s..HEAD claims a spec)", base)
		warranted = !requireClaims
	case passed == len(lines):
		headline = fmt.Sprintf("Proofbound: WARRANTED (%d of %d claiming commits verified)", passed, len(lines))
		warranted = true
	default:
		headline = fmt.Sprintf("Proofbound: NOT WARRANTED (%d of %d claiming commits verified)", passed, len(lines))
	}
	var drift bytes.Buffer
	if len(lines) > 0 {
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
	cell := func(s string) string {
		return strings.NewReplacer("|", `\|`, "\n", " ", "`", "'").Replace(printableUnit(s))
	}
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
