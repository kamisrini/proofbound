// Package projections owns the rebuildable, ledger-derived read models.
package projections

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

var ErrSnapshotMismatch = errors.New("projections: snapshots differ")

type Projector struct{}

// Snapshot contains one canonical digest multiset per derived table.
type Snapshot struct{ Tables map[string][]string }

func New() *Projector { return &Projector{} }

const ddl = `
CREATE TABLE IF NOT EXISTS projection_meta (last_seq BIGINT NOT NULL);
INSERT INTO projection_meta(last_seq) SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM projection_meta);
CREATE TABLE IF NOT EXISTS commits_view (
 sha TEXT PRIMARY KEY, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 author_name TEXT NOT NULL, author_email TEXT NOT NULL, committer_name TEXT NOT NULL,
 committer_email TEXT NOT NULL, committed_at TIMESTAMPTZ NOT NULL, subject TEXT NOT NULL,
 files_touched JSON NOT NULL, cited_decisions JSON NOT NULL
);
CREATE TABLE IF NOT EXISTS checks_view (
 run_id TEXT PRIMARY KEY, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 command TEXT NOT NULL, exit_code INTEGER NOT NULL, started_at TIMESTAMPTZ NOT NULL,
 finished_at TIMESTAMPTZ NOT NULL, duration_ms BIGINT NOT NULL, output_sha256 TEXT NOT NULL,
 git_sha TEXT NOT NULL, git_dirty BOOLEAN NOT NULL, tool_versions JSON NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions_view (
 session_id TEXT PRIMARY KEY, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 started_at TIMESTAMPTZ, finished_at TIMESTAMPTZ, message_count BIGINT NOT NULL,
 tool_call_count BIGINT NOT NULL, files_written_count BIGINT NOT NULL, parse_coverage NUMERIC NOT NULL
);
CREATE TABLE IF NOT EXISTS reviews_view (
 finding_id TEXT PRIMARY KEY, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 severity TEXT NOT NULL, reviewed_commit TEXT NOT NULL, defect_commit TEXT
);`

func (p *Projector) Ensure(ctx context.Context, s *store.Store) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error { _, err := tx.Exec(ctx, ddl); return err })
}

func (p *Projector) Apply(ctx context.Context, s *store.Store) error {
	if err := p.Ensure(ctx, s); err != nil {
		return err
	}
	var checkpoint int64
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT last_seq FROM projection_meta`).Scan(&checkpoint)
	}); err != nil {
		return err
	}
	var records []store.Record
	if err := s.ReadEvents(ctx, store.Filter{SinceSeq: checkpoint}, func(r store.Record) error {
		if r.Seq <= checkpoint || (len(records) > 0 && r.Seq <= records[len(records)-1].Seq) {
			return fmt.Errorf("projections: non-increasing ledger sequence")
		}
		records = append(records, r)
		return nil
	}); err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	return s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		for _, r := range records {
			if err := reduce(ctx, tx, r); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `UPDATE projection_meta SET last_seq=$1`, records[len(records)-1].Seq)
		return err
	})
}

func (p *Projector) Rebuild(ctx context.Context, s *store.Store) error {
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `DROP TABLE IF EXISTS commits_view, checks_view, sessions_view, reviews_view, projection_meta`)
		return err
	}); err != nil {
		return err
	}
	return p.Apply(ctx, s)
}

func reduce(ctx context.Context, tx *store.Tx, r store.Record) error {
	if !supported(r.Event.Source, r.Event.Kind) {
		return fmt.Errorf("projections: unsupported event %s/%s", r.Event.Source, r.Event.Kind)
	}
	switch {
	case r.Event.Source == core.SourceGit && r.Event.Kind == core.KindCommitRecorded:
		var v commitPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("commit seq %d: %w", r.Seq, err)
		}
		if v.SHA == "" || v.CommittedAt.IsZero() {
			return fmt.Errorf("commit seq %d: missing required field", r.Seq)
		}
		files, _ := json.Marshal(v.FilesTouched)
		cited, _ := json.Marshal(v.CitedDecisions)
		_, err := tx.Exec(ctx, `INSERT INTO commits_view(sha,event_id,seq,author_name,author_email,committer_name,committer_email,committed_at,subject,files_touched,cited_decisions) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(sha) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,author_name=EXCLUDED.author_name,author_email=EXCLUDED.author_email,committer_name=EXCLUDED.committer_name,committer_email=EXCLUDED.committer_email,committed_at=EXCLUDED.committed_at,subject=EXCLUDED.subject,files_touched=EXCLUDED.files_touched,cited_decisions=EXCLUDED.cited_decisions WHERE commits_view.seq < EXCLUDED.seq`, v.SHA, r.Event.ID.String(), r.Seq, v.AuthorName, v.AuthorEmail, v.CommitterName, v.CommitterEmail, v.CommittedAt, v.Subject, files, cited)
		return err
	case r.Event.Source == core.SourceChecks && r.Event.Kind == core.KindCheckRun:
		var v checkPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("check seq %d: %w", r.Seq, err)
		}
		if v.RunID == "" || v.StartedAt.IsZero() || v.FinishedAt.IsZero() {
			return fmt.Errorf("check seq %d: missing required field", r.Seq)
		}
		tools, _ := json.Marshal(v.ToolVersions)
		_, err := tx.Exec(ctx, `INSERT INTO checks_view(run_id,event_id,seq,command,exit_code,started_at,finished_at,duration_ms,output_sha256,git_sha,git_dirty,tool_versions) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(run_id) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,command=EXCLUDED.command,exit_code=EXCLUDED.exit_code,started_at=EXCLUDED.started_at,finished_at=EXCLUDED.finished_at,duration_ms=EXCLUDED.duration_ms,output_sha256=EXCLUDED.output_sha256,git_sha=EXCLUDED.git_sha,git_dirty=EXCLUDED.git_dirty,tool_versions=EXCLUDED.tool_versions WHERE checks_view.seq < EXCLUDED.seq`, v.RunID, r.Event.ID.String(), r.Seq, v.Command, v.ExitCode, v.StartedAt, v.FinishedAt, v.DurationMS, v.OutputSHA256, v.GitSHA, v.GitDirty, tools)
		return err
	default:
		return nil
	}
}

func supported(s core.Source, k core.Kind) bool {
	return (s == core.SourceGit && k == core.KindCommitRecorded) || (s == core.SourceChecks && k == core.KindCheckRun) || (s == core.SourceSessions && k == core.KindSessionObserved) || (s == core.SourceReviews && k == core.KindReviewVerdict)
}

type commitPayload struct {
	SHA            string    `json:"sha"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedAt    time.Time `json:"committed_at"`
	Subject        string    `json:"subject"`
	FilesTouched   []string  `json:"files_touched"`
	CitedDecisions []string  `json:"cited_decisions"`
}
type toolPayload struct {
	Go           string `json:"go"`
	GolangCILint string `json:"golangci_lint"`
	Make         string `json:"make"`
}
type checkPayload struct {
	RunID        string      `json:"run_id"`
	Command      string      `json:"command"`
	ExitCode     int         `json:"exit_code"`
	StartedAt    time.Time   `json:"started_at"`
	FinishedAt   time.Time   `json:"finished_at"`
	DurationMS   int64       `json:"duration_ms"`
	OutputSHA256 string      `json:"output_sha256"`
	GitSHA       string      `json:"git_sha"`
	GitDirty     bool        `json:"git_dirty"`
	ToolVersions toolPayload `json:"tool_versions"`
}

func decode(raw []byte, out any) error {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err == nil {
		return errors.New("trailing JSON")
	}
	return nil
}

func (p *Projector) Snapshot(ctx context.Context, s *store.Store) (Snapshot, error) {
	if err := p.Ensure(ctx, s); err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{Tables: make(map[string][]string)}
	queries := map[string]string{"commits_view": "SELECT sha,event_id,seq,author_name,author_email,committer_name,committer_email,committed_at,subject,files_touched,cited_decisions FROM commits_view", "checks_view": "SELECT run_id,event_id,seq,command,exit_code,started_at,finished_at,duration_ms,output_sha256,git_sha,git_dirty,tool_versions FROM checks_view", "sessions_view": "SELECT session_id,event_id,seq,started_at,finished_at,message_count,tool_call_count,files_written_count,parse_coverage FROM sessions_view", "reviews_view": "SELECT finding_id,event_id,seq,severity,reviewed_commit,defect_commit FROM reviews_view"}
	for name, q := range queries {
		var rowsData []map[string]any
		if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
			rows, err := tx.Query(ctx, q)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var a []any
				switch name {
				case "commits_view":
					a = make([]any, 11)
				case "checks_view":
					a = make([]any, 12)
				case "sessions_view":
					a = make([]any, 9)
				default:
					a = make([]any, 6)
				}
				ptr := make([]any, len(a))
				for i := range a {
					ptr[i] = &a[i]
				}
				if err := rows.Scan(ptr...); err != nil {
					return err
				}
				m := map[string]any{}
				for i, v := range a {
					m[fmt.Sprintf("c%d", i)] = v
				}
				rowsData = append(rowsData, m)
			}
			return rows.Err()
		}); err != nil {
			return Snapshot{}, err
		}
		for _, row := range rowsData {
			raw, err := json.Marshal(row)
			if err != nil {
				return Snapshot{}, err
			}
			canonical, err := core.Canonicalize(raw)
			if err != nil {
				return Snapshot{}, err
			}
			h := sha256.Sum256(canonical)
			result.Tables[name] = append(result.Tables[name], hex.EncodeToString(h[:]))
		}
		sort.Strings(result.Tables[name])
	}
	return result, nil
}

func CompareSnapshots(a, b Snapshot) error {
	if len(a.Tables) != len(b.Tables) {
		return ErrSnapshotMismatch
	}
	for name, av := range a.Tables {
		bv, ok := b.Tables[name]
		if !ok || len(av) != len(bv) {
			return fmt.Errorf("%w: table %s", ErrSnapshotMismatch, name)
		}
		x, y := append([]string(nil), av...), append([]string(nil), bv...)
		sort.Strings(x)
		sort.Strings(y)
		for i := range x {
			if x[i] != y[i] {
				return fmt.Errorf("%w: table %s", ErrSnapshotMismatch, name)
			}
		}
	}
	return nil
}
