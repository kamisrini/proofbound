package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// ReportWeek renders the projection data whose source events occurred in the
// seven-day window ending at now. reachable contains the commit SHAs currently
// reachable from repository refs; a missing SHA is retained and marked
// superseded. The map is intentionally supplied by the Git adapter because
// projections do not own repository observation.
func (p *Projector) ReportWeek(ctx context.Context, s *store.Store, now time.Time, reachable map[string]bool, output io.Writer) error {
	if output == nil {
		return fmt.Errorf("projections: report output is required")
	}
	if err := p.Ensure(ctx, s); err != nil {
		return fmt.Errorf("report: ensure projections: %w", err)
	}
	end := now.UTC()
	start := end.Add(-7 * 24 * time.Hour)
	var report weekReport
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		var err error
		report.commits, err = readCommitReportRows(ctx, tx, start, end, reachable)
		if err != nil {
			return err
		}
		report.checks, err = readCheckReportRows(ctx, tx, start, end)
		if err != nil {
			return err
		}
		report.sessions, err = readSessionReportRows(ctx, tx, start, end)
		if err != nil {
			return err
		}
		report.reviews, err = readReviewReportRows(ctx, tx, start, end)
		return err
	}); err != nil {
		return fmt.Errorf("report: read week: %w", err)
	}
	chains, err := readRedVerdictChains(ctx, s, start, end)
	if err != nil {
		return fmt.Errorf("report: read review chains: %w", err)
	}
	report.chains = chains
	return renderWeekReport(output, start, end, report)
}

type weekReport struct {
	commits  []commitReportRow
	checks   []checkReportRow
	sessions []sessionReportRow
	reviews  []reviewReportRow
	chains   []redVerdictChain
}

type commitReportRow struct {
	sha, eventID, subject string
	seq                   int64
	files                 int
	decisions             []string
	superseded            bool
}

type checkReportRow struct {
	runID, eventID string
	seq            int64
	exitCode       int
	durationMS     int64
}

type sessionReportRow struct {
	sessionID, eventID string
	seq                int64
	messages, tools    int64
	files              int64
	coverage           float64
}

type reviewReportRow struct {
	findingID, verdictID, eventID, status, severity, reviewedCommit, defectCommit string
	seq                                                                           int64
}

type redVerdictChain struct {
	redEventID, nextEventID string
	redSeq, nextSeq         int64
	changeEventIDs          []string
}

func readCommitReportRows(ctx context.Context, tx *store.Tx, start, end time.Time, reachable map[string]bool) ([]commitReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT c.sha,c.event_id,c.seq,c.subject,c.files_touched,c.cited_decisions,e.occurred_at
FROM commits_view c LEFT JOIN events e ON e.event_id=c.event_id
WHERE e.event_id IS NULL OR (e.occurred_at >= $1 AND e.occurred_at < $2) ORDER BY e.occurred_at,c.seq,c.sha`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []commitReportRow
	for rows.Next() {
		var row commitReportRow
		var filesRaw, decisionsRaw []byte
		var occurredAt *time.Time
		if err := rows.Scan(&row.sha, &row.eventID, &row.seq, &row.subject, &filesRaw, &decisionsRaw, &occurredAt); err != nil {
			return nil, err
		}
		if occurredAt == nil {
			return nil, fmt.Errorf("commit %s: missing event proof %s", row.sha, row.eventID)
		}
		var files, decisions []string
		if err := json.Unmarshal(filesRaw, &files); err != nil {
			return nil, fmt.Errorf("commit %s files: %w", row.sha, err)
		}
		if err := json.Unmarshal(decisionsRaw, &decisions); err != nil {
			return nil, fmt.Errorf("commit %s decisions: %w", row.sha, err)
		}
		row.files, row.decisions = len(files), decisions
		row.superseded = reachable != nil && !reachable[row.sha]
		result = append(result, row)
	}
	return result, rows.Err()
}

func readCheckReportRows(ctx context.Context, tx *store.Tx, start, end time.Time) ([]checkReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT c.run_id,c.event_id,c.seq,c.exit_code,c.duration_ms,e.occurred_at
FROM checks_view c LEFT JOIN events e ON e.event_id=c.event_id
WHERE e.event_id IS NULL OR (e.occurred_at >= $1 AND e.occurred_at < $2) ORDER BY e.occurred_at,c.seq,c.run_id`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []checkReportRow
	for rows.Next() {
		var row checkReportRow
		var occurredAt *time.Time
		if err := rows.Scan(&row.runID, &row.eventID, &row.seq, &row.exitCode, &row.durationMS, &occurredAt); err != nil {
			return nil, err
		}
		if occurredAt == nil {
			return nil, fmt.Errorf("check %s: missing event proof %s", row.runID, row.eventID)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func readSessionReportRows(ctx context.Context, tx *store.Tx, start, end time.Time) ([]sessionReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT s.session_id,s.event_id,s.seq,s.message_count,s.tool_call_count,s.files_written_count,s.parse_coverage,e.occurred_at
FROM sessions_view s LEFT JOIN events e ON e.event_id=s.event_id
WHERE e.event_id IS NULL OR (e.occurred_at >= $1 AND e.occurred_at < $2) ORDER BY e.occurred_at,s.seq,s.session_id`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []sessionReportRow
	for rows.Next() {
		var row sessionReportRow
		var occurredAt *time.Time
		if err := rows.Scan(&row.sessionID, &row.eventID, &row.seq, &row.messages, &row.tools, &row.files, &row.coverage, &occurredAt); err != nil {
			return nil, err
		}
		if occurredAt == nil {
			return nil, fmt.Errorf("session %s: missing event proof %s", row.sessionID, row.eventID)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func readReviewReportRows(ctx context.Context, tx *store.Tx, start, end time.Time) ([]reviewReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT r.finding_id,r.verdict_id,r.event_id,r.seq,r.status,r.severity,r.reviewed_commit,COALESCE(r.defect_commit,''),e.occurred_at FROM reviews_view r LEFT JOIN events e ON e.event_id=r.event_id WHERE e.event_id IS NULL OR (e.occurred_at >= $1 AND e.occurred_at < $2) ORDER BY e.occurred_at,r.seq,r.finding_id`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []reviewReportRow
	for rows.Next() {
		var row reviewReportRow
		var occurredAt *time.Time
		if err := rows.Scan(&row.findingID, &row.verdictID, &row.eventID, &row.seq, &row.status, &row.severity, &row.reviewedCommit, &row.defectCommit, &occurredAt); err != nil {
			return nil, err
		}
		if occurredAt == nil {
			return nil, fmt.Errorf("review finding %s: missing event proof %s", row.findingID, row.eventID)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

type chainMarker struct {
	seq           int64
	id, kind, ref string
	red           bool
	occurredAt    time.Time
}

func readRedVerdictChains(ctx context.Context, s *store.Store, start, end time.Time) ([]redVerdictChain, error) {
	var markers []chainMarker
	if err := s.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		if record.Event.Source == core.SourceReviews && record.Event.Kind == core.KindReviewVerdict {
			if record.Event.OccurredAt.Before(start) || !record.Event.OccurredAt.Before(end) {
				return nil
			}
			var payload reviewPayload
			if err := json.Unmarshal(record.Event.Payload, &payload); err != nil {
				return fmt.Errorf("review event %s: %w", record.Event.ID, err)
			}
			if err := payload.validate(); err != nil {
				return fmt.Errorf("review event %s: %w", record.Event.ID, err)
			}
			markers = append(markers, chainMarker{record.Seq, record.Event.ID.String(), "review", payload.ReviewedCommit, payload.Status == "NEEDS_WORK", record.Event.OccurredAt})
		} else if record.Event.Source == core.SourceGit && record.Event.Kind == core.KindCommitRecorded {
			var payload commitPayload
			if err := json.Unmarshal(record.Event.Payload, &payload); err != nil {
				return err
			}
			markers = append(markers, chainMarker{record.Seq, record.Event.ID.String(), "commit", payload.SHA, false, record.Event.OccurredAt})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(markers, func(i, j int) bool { return markers[i].seq < markers[j].seq })
	var chains []redVerdictChain
	for i, marker := range markers {
		if !marker.red {
			continue
		}
		var changes []string
		for _, next := range markers[i+1:] {
			if next.kind == "commit" {
				changes = append(changes, next.id)
				continue
			}
			if len(changes) > 0 {
				chains = append(chains, redVerdictChain{marker.id, next.id, marker.seq, next.seq, changes})
			}
			break
		}
	}
	return chains, nil
}

func renderWeekReport(output io.Writer, start, end time.Time, report weekReport) error {
	if _, err := fmt.Fprintf(output, "week %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "commits count=%d\n", len(report.commits)); err != nil {
		return err
	}
	for _, row := range report.commits {
		status := ""
		if row.superseded {
			status = " [superseded]"
		}
		decisions := strings.Join(row.decisions, ",")
		if _, err := fmt.Fprintf(output, "- %s %s%s files=%d decisions=%s proof=%s\n", row.sha, row.subject, status, row.files, decisions, row.eventID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "checks count=%d\n", len(report.checks)); err != nil {
		return err
	}
	for _, row := range report.checks {
		status := "pass"
		if row.exitCode != 0 {
			status = "fail"
		}
		if _, err := fmt.Fprintf(output, "- %s %s duration_ms=%d proof=%s\n", row.runID, status, row.durationMS, row.eventID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "sessions count=%d\n", len(report.sessions)); err != nil {
		return err
	}
	for _, row := range report.sessions {
		if _, err := fmt.Fprintf(output, "- %s messages=%d tool_calls=%d files_written=%d coverage=%.3f proof=%s\n", row.sessionID, row.messages, row.tools, row.files, row.coverage, row.eventID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "reviews count=%d\n", len(report.reviews)); err != nil {
		return err
	}
	for _, row := range report.reviews {
		if _, err := fmt.Fprintf(output, "- %s verdict=%s status=%s severity=%s reviewed=%s proof=%s\n", row.findingID, row.verdictID, row.status, row.severity, row.reviewedCommit, row.eventID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "red_verdict_chains count=%d\n", len(report.chains)); err != nil {
		return err
	}
	for _, chain := range report.chains {
		if _, err := fmt.Fprintf(output, "- red_proof=%s red_seq=%d changes=%s next_verdict_proof=%s next_verdict_seq=%d\n", chain.redEventID, chain.redSeq, strings.Join(chain.changeEventIDs, ","), chain.nextEventID, chain.nextSeq); err != nil {
			return err
		}
	}
	return nil
}
