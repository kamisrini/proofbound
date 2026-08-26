package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

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
		return err
	}); err != nil {
		return fmt.Errorf("report: read week: %w", err)
	}
	return renderWeekReport(output, start, end, report)
}

type weekReport struct {
	commits  []commitReportRow
	checks   []checkReportRow
	sessions []sessionReportRow
}

type commitReportRow struct {
	sha, eventID, subject string
	seq                   int64
	files                 int
	decisions             []string
	superseded            bool
	at                    time.Time
}

type checkReportRow struct {
	runID, eventID string
	seq            int64
	exitCode       int
	durationMS     int64
	at             time.Time
}

type sessionReportRow struct {
	sessionID, eventID string
	seq                int64
	messages, tools    int64
	files              int64
	coverage           float64
	at                 time.Time
}

func readCommitReportRows(ctx context.Context, tx *store.Tx, start, end time.Time, reachable map[string]bool) ([]commitReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT c.sha,c.event_id,c.seq,c.subject,c.files_touched,c.cited_decisions,c.committed_at
FROM commits_view c JOIN events e ON e.event_id=c.event_id
WHERE e.occurred_at >= $1 AND e.occurred_at < $2 ORDER BY e.occurred_at,c.seq,c.sha`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []commitReportRow
	for rows.Next() {
		var row commitReportRow
		var filesRaw, decisionsRaw []byte
		if err := rows.Scan(&row.sha, &row.eventID, &row.seq, &row.subject, &filesRaw, &decisionsRaw, &row.at); err != nil {
			return nil, err
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
	rows, err := tx.Query(ctx, `SELECT c.run_id,c.event_id,c.seq,c.exit_code,c.duration_ms,c.started_at
FROM checks_view c JOIN events e ON e.event_id=c.event_id
WHERE e.occurred_at >= $1 AND e.occurred_at < $2 ORDER BY e.occurred_at,c.seq,c.run_id`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []checkReportRow
	for rows.Next() {
		var row checkReportRow
		if err := rows.Scan(&row.runID, &row.eventID, &row.seq, &row.exitCode, &row.durationMS, &row.at); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func readSessionReportRows(ctx context.Context, tx *store.Tx, start, end time.Time) ([]sessionReportRow, error) {
	rows, err := tx.Query(ctx, `SELECT s.session_id,s.event_id,s.seq,s.message_count,s.tool_call_count,s.files_written_count,s.parse_coverage,e.occurred_at
FROM sessions_view s JOIN events e ON e.event_id=s.event_id
WHERE e.occurred_at >= $1 AND e.occurred_at < $2 ORDER BY e.occurred_at,s.seq,s.session_id`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []sessionReportRow
	for rows.Next() {
		var row sessionReportRow
		if err := rows.Scan(&row.sessionID, &row.eventID, &row.seq, &row.messages, &row.tools, &row.files, &row.coverage, &row.at); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
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
	return nil
}
