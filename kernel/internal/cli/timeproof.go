package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// U3 (docs/plans/ui-u3-time-proof-spec.md, CLI-INV-50/51): the time rail and proof in the browser.

// browserProof is what the page's script needs to verify one signature itself: the exact bytes,
// the signature, the declared key the server checked against, and the tag the signature covers.
type browserProof struct {
	Kind    string // "record" or "file"
	Label   string
	Bytes64 string
	Sig     string // record: empty (the signature is inside the record); file: the .sig line
	KeyHex  string
	Tag     string
}

func recordProof(label string, data []byte, key ed25519.PublicKey, schema string) browserProof {
	return browserProof{Kind: "record", Label: label, Bytes64: base64.StdEncoding.EncodeToString(data), KeyHex: hex.EncodeToString(key), Tag: schema + "\n"}
}

// reviewProofs are the review and verdict files of a claiming commit at HEAD with their .sig lines,
// each with the declared key (at HEAD) whose public half signed it, if any.
func reviewProofs(ctx context.Context, root, commit string) []browserProof {
	keys := declaredKeys(ctx, root, "HEAD")
	var out []browserProof
	for _, p := range strings.Split(gitOutput(ctx, root, "ls-tree", "--name-only", "HEAD", "--", "docs/verification/verdicts/"), "\n") {
		p = strings.TrimSpace(p)
		base := path.Base(p)
		if !strings.HasSuffix(base, ".md") || (!strings.HasPrefix(base, "RR-"+commit[:12]+"-") && !strings.HasPrefix(base, "OV-"+commit[:12]+"-")) {
			continue
		}
		data, err1 := gitBytes(ctx, root, "show", "HEAD:"+p)
		sig, err2 := gitBytes(ctx, root, "show", "HEAD:"+p+".sig")
		if err1 != nil || err2 != nil {
			continue
		}
		pr := browserProof{Kind: "file", Label: base, Bytes64: base64.StdEncoding.EncodeToString(data), Sig: strings.TrimSpace(string(sig)), Tag: reviewSigTagLine}
		if f := strings.Fields(pr.Sig); len(f) == 3 {
			for _, k := range keys {
				if hex.EncodeToString(k) == f[1] {
					pr.KeyHex = f[1]
				}
			}
		}
		out = append(out, pr)
	}
	return out
}

// reviewSigTagLine is the tag a review file's .sig covers (warranty package, option A).
const reviewSigTagLine = "proofbound.review-signature.v1\n"

// asOfInfo is the past moment the Audit room shows, and how it differs from now.
type asOfInfo struct {
	Commit   string        `json:"commit"`
	Date     string        `json:"date"`
	Changed  []controlDiff `json:"changed_controls"`
	OnlyThen []string      `json:"exceptions_only_then"`
	OnlyNow  []string      `json:"exceptions_only_now"`
	// SyncNote is set when the record at that moment could not be fully built; the page says so.
	SyncNote string `json:"sync_note,omitempty"`
}

type controlDiff struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Then     string `json:"then"`
	Now      string `json:"now"`
}

// compareMoments lists what differs between the audit then and now.
func compareMoments(then, now auditReport) ([]controlDiff, []string, []string) {
	nowByID := map[string]auditControl{}
	for _, c := range now.Controls {
		nowByID[c.ID] = c
	}
	var changed []controlDiff
	for _, c := range then.Controls {
		n := nowByID[c.ID]
		if c.State != n.State || c.Count != n.Count {
			changed = append(changed, controlDiff{ID: c.ID, Question: c.Question, Then: c.State + " · " + c.Count, Now: n.State + " · " + n.Count})
		}
	}
	thenIDs, nowIDs := map[string]bool{}, map[string]bool{}
	for _, e := range then.Exceptions {
		thenIDs[e.ID] = true
	}
	for _, e := range now.Exceptions {
		nowIDs[e.ID] = true
	}
	var onlyThen, onlyNow []string
	for _, e := range then.Exceptions {
		if !nowIDs[e.ID] {
			onlyThen = append(onlyThen, e.ID+": "+e.Text)
		}
	}
	for _, e := range now.Exceptions {
		if !thenIDs[e.ID] {
			onlyNow = append(onlyNow, e.ID+": "+e.Text)
		}
	}
	return changed, onlyThen, onlyNow
}

// withCheckout runs fn on a temporary detached worktree at commit and always removes it: no hooks run
// (core.hooksPath is emptied, so a GET never runs the user's post-checkout hook), and the worktree's
// metadata is pruned afterwards, even when fn fails or the request is cancelled (review
// ui-u3-review-20261122 F3).
func withCheckout(ctx context.Context, root, commit string, fn func(wt string) error) error {
	tmp, err := os.MkdirTemp("", "proofbound-asof-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	wt := filepath.Join(tmp, "checkout")
	if out, err := gitRun(ctx, root, "-c", "core.hooksPath="+os.DevNull, "worktree", "add", "--detach", "-q", wt, commit); err != nil {
		return fmt.Errorf("time rail: cannot check out %s: %s", short(commit, 12), out)
	}
	defer func() {
		_, _ = gitRun(context.Background(), root, "worktree", "remove", "--force", wt)
		_, _ = gitRun(context.Background(), root, "worktree", "prune")
	}()
	return fn(wt)
}

// asOfCompute builds the audit of base..commit from a temporary worktree at commit with its own
// record (CLI-INV-50). It never touches the repository's own record. A sync error is returned as a
// note with the result, which is then not cached (F2).
func asOfCompute(ctx context.Context, root, databaseURL, base, commit string) (ledgerView, string, error) {
	if databaseURL != "" {
		return ledgerView{}, "", errors.New("the time rail needs the embedded database; it is not available with DATABASE_URL set")
	}
	var lv ledgerView
	note := ""
	err := withCheckout(ctx, root, commit, func(wt string) error {
		var syncOut strings.Builder
		if err := runCommand(ctx, commandSyncAll, []string{"sync", "all"}, wt, "", &syncOut); err != nil {
			note = printableUnit(err.Error())
		}
		ledger, err := openStore(ctx, wt, "")
		if err != nil {
			return err
		}
		lv.Report, err = computeAudit(ctx, wt, ledger, base)
		if err == nil {
			var events []promiseEvent
			if events, err = readPromiseEvents(ctx, ledger); err == nil {
				lv.Promises = computePromises(events)
			}
		}
		if err == nil {
			attachOutcomes(ctx, wt, &lv.Report)
			withReviewProofs(ctx, wt, &lv.Report)
		}
		return errors.Join(err, closeStore(ledger))
	})
	return lv, note, err
}

// withReviewProofs attaches each role's review-file proofs (for the page only).
func withReviewProofs(ctx context.Context, root string, rep *auditReport) {
	for i := range rep.Roles {
		rep.Roles[i].Proofs = reviewProofs(ctx, root, rep.Roles[i].Commit)
	}
}
