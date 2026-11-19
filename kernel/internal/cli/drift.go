package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	intentsdd "github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// R2 M4 spec drift (docs/plans/R2-m4-drift-spec.md, CLI-INV-21): a warranty is STALE when the change
// now promises something other than what was verified. A promise is the (obligation id, statement)
// set of a change-intent revision's targets, never its digest: an archive move changes every digest
// and requirement id while usually promising exactly the same thing.

// promise maps obligation id to statement. Two targets sharing an obligation id keep both
// statements, sorted and joined, so a duplicate can neither hide nor invent a difference.
type promise map[string]string

type warranty struct {
	verdictID, commit, intentID, digest string
	seq                                 int64
}

// headState is what a warranted change is at HEAD, read from the repository by the sdd provider
// (SDD-INV-68), never from ledger order: a revert re-emits an old revision's exact bytes, which the
// ledger dedupes, so "latest ledger seq" can name a revision the tree no longer has (review
// r2-m4-drift-review-20261119 F1).
type headState struct {
	digest string // the change's revision digest at HEAD, when state == ""
	state  string // "" (compare), "REMOVED", "UNSYNCED" or "UNKNOWN"
	reason string
}

type driftChange struct{ kind, obligation, before, after string }

type driftEntry struct {
	intentID string
	state    string // CURRENT, STALE, REMOVED, UNSYNCED or UNKNOWN
	verified warranty
	current  string
	reason   string
	changes  []driftChange
}

// computeDrift is the whole drift rule, as a pure function of the projected rows and each warranted
// change's state at HEAD.
func computeDrift(warranties []warranty, heads map[string]headState, promises map[string]promise) []driftEntry {
	byIntent := map[string][]warranty{}
	for _, w := range warranties {
		byIntent[w.intentID] = append(byIntent[w.intentID], w)
	}
	ids := make([]string, 0, len(byIntent))
	for id := range byIntent {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var out []driftEntry
	for _, id := range ids {
		ws := byIntent[id]
		sort.Slice(ws, func(a, b int) bool {
			return ws[a].seq > ws[b].seq || (ws[a].seq == ws[b].seq && ws[a].verdictID < ws[b].verdictID)
		})
		head := heads[id]
		entry := driftEntry{intentID: id, verified: ws[0], current: head.digest, reason: head.reason}
		now, ledgered := promises[id+"\x00"+head.digest]
		switch {
		case head.state != "":
			entry.state = head.state
		case !ledgered:
			entry.state, entry.reason = "UNSYNCED", "the change's revision at HEAD is not in the ledger; run `proofbound sync intent sdd`"
		default:
			entry.state = "STALE"
			for _, w := range ws {
				if samePromise(promises[id+"\x00"+w.digest], now) {
					entry.state, entry.verified = "CURRENT", w
					break
				}
			}
			if entry.state == "STALE" {
				entry.changes = diffPromises(promises[id+"\x00"+ws[0].digest], now)
			}
		}
		out = append(out, entry)
	}
	return out
}

func samePromise(a, b promise) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

func diffPromises(before, after promise) []driftChange {
	keys := map[string]bool{}
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	var out []driftChange
	for _, k := range sorted {
		b, inBefore := before[k]
		a, inAfter := after[k]
		switch {
		case !inBefore:
			out = append(out, driftChange{"added", k, "", a})
		case !inAfter:
			out = append(out, driftChange{"removed", k, b, ""})
		case a != b:
			out = append(out, driftChange{"changed", k, b, a})
		}
	}
	return out
}

func short(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// quoteStatement keeps a tenant statement on one line (strconv.Quote escapes every control).
func quoteStatement(s string) string { return strconv.Quote(s) }

// sddStateProvider answers a change's revision in a commit's own state (SDD-INV-68). A variable so a
// test can stand one in.
var newSDDStateProvider = func(root string, now time.Time) (interface {
	StateRevision(ctx context.Context, commit, name string) (connectorintent.Revision, error)
}, error) {
	repo, err := gitcmd.New(root)
	if err != nil {
		return nil, err
	}
	reader, err := intentsdd.NewGitReader(repo)
	if err != nil {
		return nil, err
	}
	return intentsdd.New(reader, now)
}

// changeName is the native change name of a change intent revision's directory: the base name,
// without the archive's date prefix.
func changeName(artifactPath string) string {
	base := path.Base(artifactPath)
	if sddDatedRE.MatchString(base) {
		return base[len("YYYY-MM-DD-"):]
	}
	return base
}

// reportDrift projects the ledger, reads each warranted change's state at HEAD, and prints the
// drift report.
func reportDrift(ctx context.Context, root string, ledger *store.Store, output io.Writer) error {
	if err := applyProjection(ctx, projections.New(), ledger); err != nil {
		return err
	}
	var warranties []warranty
	names := map[string]string{}
	statements := map[string]map[string][]string{}
	err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT verdict_id, commit_sha, intent_id, intent_artifact_sha256, max(seq) FROM obligation_verdicts_view WHERE intent_source='intent.sdd' GROUP BY 1,2,3,4 HAVING bool_and(outcome='SATISFIED')`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var w warranty
			if err := rows.Scan(&w.verdictID, &w.commit, &w.intentID, &w.digest, &w.seq); err != nil {
				rows.Close()
				return err
			}
			warranties = append(warranties, w)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		rows, err = tx.Query(ctx, `SELECT intent_id, payload->>'artifact_path' FROM change_intents_view WHERE source='intent.sdd'`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id, artifactPath string
			if err := rows.Scan(&id, &artifactPath); err != nil {
				rows.Close()
				return err
			}
			names[id] = changeName(artifactPath)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		rows, err = tx.Query(ctx, `SELECT t.intent_id, t.intent_artifact_sha256, t.obligation_id, o.statement FROM intent_targets_view t JOIN requirement_obligations_view o ON o.source=t.requirement_source AND o.requirement_id=t.requirement_id AND o.artifact_sha256=t.requirement_artifact_sha256 AND o.obligation_id=t.obligation_id WHERE t.intent_source='intent.sdd'`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id, digest, obligation, statement string
			if err := rows.Scan(&id, &digest, &obligation, &statement); err != nil {
				rows.Close()
				return err
			}
			key := id + "\x00" + digest
			if statements[key] == nil {
				statements[key] = map[string][]string{}
			}
			statements[key][obligation] = append(statements[key][obligation], statement)
		}
		rows.Close()
		return rows.Err()
	})
	if err != nil {
		return err
	}
	promises := map[string]promise{}
	for key, byObligation := range statements {
		p := promise{}
		for obligation, list := range byObligation {
			sort.Strings(list)
			p[obligation] = strings.Join(list, " | ")
		}
		promises[key] = p
	}
	heads := map[string]headState{}
	if len(warranties) > 0 {
		head := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "HEAD^{commit}"))
		if len(head) != 40 {
			return errors.New("report drift: cannot resolve HEAD")
		}
		provider, err := newSDDStateProvider(root, time.Now())
		if err != nil {
			return err
		}
		for _, w := range warranties {
			if _, done := heads[w.intentID]; done {
				continue
			}
			revision, err := provider.StateRevision(ctx, head, names[w.intentID])
			var unmappable *intentsdd.UnmappableError
			switch {
			case err == nil:
				heads[w.intentID] = headState{digest: revision.ArtifactSHA256}
			case errors.Is(err, intentsdd.ErrUnknownChange):
				heads[w.intentID] = headState{state: "REMOVED", reason: "the change no longer exists at HEAD"}
			case errors.Is(err, intentsdd.ErrNoChangeIntent):
				heads[w.intentID] = headState{state: "UNKNOWN", reason: "the change has no change intent at HEAD"}
			case errors.As(err, &unmappable):
				heads[w.intentID] = headState{state: "UNKNOWN", reason: "unmappable at HEAD: " + unmappable.Cause}
			default:
				return fmt.Errorf("report drift: %s at HEAD: %w", w.intentID, err)
			}
		}
	}
	entries := computeDrift(warranties, heads, promises)
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.state]++
		var err error
		switch e.state {
		case "STALE":
			_, err = fmt.Fprintf(output, "drift change=%s state=STALE verified_commit=%s verdict=%s verified_revision=%s current_revision=%s\n", e.intentID, short(e.verified.commit, 12), e.verified.verdictID, short(e.verified.digest, 8), short(e.current, 8))
			for _, c := range e.changes {
				if err != nil {
					break
				}
				switch c.kind {
				case "changed":
					_, err = fmt.Fprintf(output, "  changed %s %s -> %s\n", c.obligation, quoteStatement(c.before), quoteStatement(c.after))
				case "added":
					_, err = fmt.Fprintf(output, "  added %s %s\n", c.obligation, quoteStatement(c.after))
				default:
					_, err = fmt.Fprintf(output, "  removed %s %s\n", c.obligation, quoteStatement(c.before))
				}
			}
		case "CURRENT":
			_, err = fmt.Fprintf(output, "drift change=%s state=CURRENT verified_commit=%s verdict=%s current_revision=%s\n", e.intentID, short(e.verified.commit, 12), e.verified.verdictID, short(e.current, 8))
		default:
			_, err = fmt.Fprintf(output, "drift change=%s state=%s verified_commit=%s verdict=%s reason=%s\n", e.intentID, e.state, short(e.verified.commit, 12), e.verified.verdictID, strconv.Quote(e.reason))
		}
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "summary warranted=%d current=%d stale=%d other=%d\n", len(entries), counts["CURRENT"], counts["STALE"], len(entries)-counts["CURRENT"]-counts["STALE"])
	return err
}
