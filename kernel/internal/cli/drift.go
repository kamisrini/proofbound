package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

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

type intentRevision struct {
	digest string
	seq    int64
}

type driftChange struct{ kind, obligation, before, after string }

type driftEntry struct {
	intentID string
	stale    bool
	verified warranty
	current  string
	changes  []driftChange
}

// computeDrift is the whole drift rule, as a pure function of the projected rows.
func computeDrift(warranties []warranty, revisions map[string][]intentRevision, promises map[string]promise) []driftEntry {
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
		revs := revisions[id]
		if len(revs) == 0 {
			continue
		}
		current := revs[0]
		for _, r := range revs[1:] {
			if r.seq > current.seq {
				current = r
			}
		}
		now := promises[id+"\x00"+current.digest]
		ws := byIntent[id]
		sort.Slice(ws, func(a, b int) bool { return ws[a].seq > ws[b].seq || (ws[a].seq == ws[b].seq && ws[a].verdictID < ws[b].verdictID) })
		entry := driftEntry{intentID: id, current: current.digest, verified: ws[0]}
		covered := false
		for _, w := range ws {
			if samePromise(promises[id+"\x00"+w.digest], now) {
				covered, entry.verified = true, w
				break
			}
		}
		if !covered {
			entry.stale = true
			entry.changes = diffPromises(promises[id+"\x00"+ws[0].digest], now)
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

// reportDrift projects the ledger and prints the drift report.
func reportDrift(ctx context.Context, ledger *store.Store, output io.Writer) error {
	if err := applyProjection(ctx, projections.New(), ledger); err != nil {
		return err
	}
	var warranties []warranty
	revisions := map[string][]intentRevision{}
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
		rows, err = tx.Query(ctx, `SELECT intent_id, artifact_sha256, seq FROM change_intents_view WHERE source='intent.sdd'`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			var r intentRevision
			if err := rows.Scan(&id, &r.digest, &r.seq); err != nil {
				rows.Close()
				return err
			}
			revisions[id] = append(revisions[id], r)
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
	entries := computeDrift(warranties, revisions, promises)
	stale := 0
	for _, e := range entries {
		var err error
		if e.stale {
			stale++
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
		} else {
			_, err = fmt.Fprintf(output, "drift change=%s state=CURRENT verified_commit=%s verdict=%s current_revision=%s\n", e.intentID, short(e.verified.commit, 12), e.verified.verdictID, short(e.current, 8))
		}
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "summary warranted=%d current=%d stale=%d\n", len(entries), len(entries)-stale, stale)
	return err
}
