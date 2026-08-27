// Package projections owns the rebuildable, ledger-derived read models.
package projections

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

var ErrSnapshotMismatch = errors.New("projections: snapshots differ")

var (
	shaPattern         = regexp.MustCompile(`^[0-9a-f]{40}$|^[0-9a-f]{64}$`)
	ulidPattern        = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)
	hexPattern         = regexp.MustCompile(`^[0-9a-f]{64}$`)
	gitPattern         = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
	emailPattern       = regexp.MustCompile(`^[^@\s]+@[^@\s]+$`)
	decisionPattern    = regexp.MustCompile(`^VD-[a-z0-9]+(?:-[a-z0-9]+)*-[a-z0-9]{6}$`)
	makeCommandPattern = regexp.MustCompile(`^make [A-Za-z][A-Za-z0-9_-]*$`)
)

type Projector struct{}

// Snapshot contains one canonical digest multiset per derived table.
type Snapshot struct{ Tables map[string][]string }

type snapshotRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}

var snapshotDigestRows = appendSnapshotDigests

func New() *Projector { return &Projector{} }

const ddl = `
CREATE TABLE IF NOT EXISTS projection_meta (
 projection_name TEXT PRIMARY KEY, projection_version INTEGER NOT NULL, last_seq BIGINT NOT NULL
);
INSERT INTO projection_meta(projection_name, projection_version, last_seq)
 VALUES('default', 1, 0) ON CONFLICT(projection_name) DO NOTHING;
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
 finding_id TEXT PRIMARY KEY, verdict_id TEXT NOT NULL, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 status TEXT NOT NULL, severity TEXT NOT NULL, reviewed_commit TEXT NOT NULL, defect_commit TEXT,
 artifact_path TEXT NOT NULL, artifact_sha TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS github_delivery_view (
 native_id TEXT PRIMARY KEY, event_id TEXT NOT NULL, seq BIGINT NOT NULL,
 kind TEXT NOT NULL, repository TEXT NOT NULL, commit_sha TEXT NOT NULL,
 workflow_id BIGINT NOT NULL, workflow TEXT NOT NULL, deployment_id BIGINT NOT NULL,
 environment TEXT NOT NULL, status TEXT NOT NULL, conclusion TEXT NOT NULL,
 occurred_at TIMESTAMPTZ NOT NULL, freshness_at TIMESTAMPTZ NOT NULL, url TEXT NOT NULL
);
ALTER TABLE reviews_view ADD COLUMN IF NOT EXISTS verdict_id TEXT;
ALTER TABLE reviews_view ADD COLUMN IF NOT EXISTS status TEXT;
ALTER TABLE reviews_view ADD COLUMN IF NOT EXISTS artifact_path TEXT;
ALTER TABLE reviews_view ADD COLUMN IF NOT EXISTS artifact_sha TEXT;`

func (p *Projector) Ensure(ctx context.Context, s *store.Store) error {
	return s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error { _, err := tx.Exec(ctx, ddl); return err })
}

func (p *Projector) Apply(ctx context.Context, s *store.Store) error {
	if err := p.Ensure(ctx, s); err != nil {
		return err
	}
	var checkpoint int64
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT last_seq FROM projection_meta WHERE projection_name='default' AND projection_version=1`).Scan(&checkpoint)
	}); err != nil {
		return err
	}
	var records []store.Record
	if err := s.ReadEvents(ctx, store.Filter{SinceSeq: checkpoint}, func(r store.Record) error {
		if err := validateSequence(checkpoint, records, r.Seq); err != nil {
			return err
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
		_, err := tx.Exec(ctx, `UPDATE projection_meta SET last_seq=$1 WHERE projection_name='default' AND projection_version=1`, records[len(records)-1].Seq)
		return err
	})
}

func validateSequence(checkpoint int64, records []store.Record, seq int64) error {
	if seq <= checkpoint || (len(records) > 0 && seq <= records[len(records)-1].Seq) {
		return fmt.Errorf("projections: non-increasing ledger sequence")
	}
	return nil
}

func (p *Projector) Rebuild(ctx context.Context, s *store.Store) error {
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `DROP TABLE IF EXISTS commits_view, checks_view, sessions_view, reviews_view, github_delivery_view, projection_meta`)
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
	case eventRoute(r.Event.Source, r.Event.Kind) == routeCommit:
		var v commitPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("commit seq %d: %w", r.Seq, err)
		}
		if err := v.validate(); err != nil {
			return fmt.Errorf("commit seq %d: %w", r.Seq, err)
		}
		files, _ := json.Marshal(v.FilesTouched)
		cited, _ := json.Marshal(v.CitedDecisions)
		_, err := tx.Exec(ctx, `INSERT INTO commits_view(sha,event_id,seq,author_name,author_email,committer_name,committer_email,committed_at,subject,files_touched,cited_decisions) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(sha) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,author_name=EXCLUDED.author_name,author_email=EXCLUDED.author_email,committer_name=EXCLUDED.committer_name,committer_email=EXCLUDED.committer_email,committed_at=EXCLUDED.committed_at,subject=EXCLUDED.subject,files_touched=EXCLUDED.files_touched,cited_decisions=EXCLUDED.cited_decisions WHERE commits_view.seq < EXCLUDED.seq`, v.SHA, r.Event.ID.String(), r.Seq, v.AuthorName, v.AuthorEmail, v.CommitterName, v.CommitterEmail, v.CommittedAt, v.Subject, files, cited)
		return err
	case eventRoute(r.Event.Source, r.Event.Kind) == routeCheck:
		var v checkPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("check seq %d: %w", r.Seq, err)
		}
		if err := v.validate(); err != nil {
			return fmt.Errorf("check seq %d: %w", r.Seq, err)
		}
		tools, _ := json.Marshal(v.ToolVersions)
		_, err := tx.Exec(ctx, `INSERT INTO checks_view(run_id,event_id,seq,command,exit_code,started_at,finished_at,duration_ms,output_sha256,git_sha,git_dirty,tool_versions) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(run_id) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,command=EXCLUDED.command,exit_code=EXCLUDED.exit_code,started_at=EXCLUDED.started_at,finished_at=EXCLUDED.finished_at,duration_ms=EXCLUDED.duration_ms,output_sha256=EXCLUDED.output_sha256,git_sha=EXCLUDED.git_sha,git_dirty=EXCLUDED.git_dirty,tool_versions=EXCLUDED.tool_versions WHERE checks_view.seq < EXCLUDED.seq`, v.RunID, r.Event.ID.String(), r.Seq, v.Command, v.ExitCode, v.StartedAt, v.FinishedAt, v.DurationMS, v.OutputSHA256, v.GitSHA, v.GitDirty, tools)
		return err
	case eventRoute(r.Event.Source, r.Event.Kind) == routeSession:
		var v sessionPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("session seq %d: %w", r.Seq, err)
		}
		if err := v.validate(); err != nil {
			return fmt.Errorf("session seq %d: %w", r.Seq, err)
		}
		_, err := tx.Exec(ctx, `INSERT INTO sessions_view(session_id,event_id,seq,started_at,finished_at,message_count,tool_call_count,files_written_count,parse_coverage) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(session_id) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,started_at=EXCLUDED.started_at,finished_at=EXCLUDED.finished_at,message_count=EXCLUDED.message_count,tool_call_count=EXCLUDED.tool_call_count,files_written_count=EXCLUDED.files_written_count,parse_coverage=EXCLUDED.parse_coverage WHERE sessions_view.seq < EXCLUDED.seq`, v.SessionID, r.Event.ID.String(), r.Seq, v.StartedAt, v.FinishedAt, v.MessageCount, v.ToolCallCount, v.FilesWrittenCount, v.ParseCoverage)
		return err
	case eventRoute(r.Event.Source, r.Event.Kind) == routeReview:
		var v reviewPayload
		if err := decode(r.Event.Payload, &v); err != nil {
			return fmt.Errorf("review seq %d: %w", r.Seq, err)
		}
		if err := v.validate(); err != nil {
			return fmt.Errorf("review seq %d: %w", r.Seq, err)
		}
		for _, finding := range v.Findings {
			if _, err := tx.Exec(ctx, `INSERT INTO reviews_view(finding_id,verdict_id,event_id,seq,status,severity,reviewed_commit,defect_commit,artifact_path,artifact_sha) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(finding_id) DO UPDATE SET verdict_id=EXCLUDED.verdict_id,event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,status=EXCLUDED.status,severity=EXCLUDED.severity,reviewed_commit=EXCLUDED.reviewed_commit,defect_commit=EXCLUDED.defect_commit,artifact_path=EXCLUDED.artifact_path,artifact_sha=EXCLUDED.artifact_sha WHERE reviews_view.seq < EXCLUDED.seq`, finding.FindingID, v.VerdictID, r.Event.ID.String(), r.Seq, v.Status, finding.Severity, v.ReviewedCommit, nullString(finding.DefectCommit), v.ArtifactPath, v.ArtifactSHA); err != nil {
				return err
			}
		}
		return nil
	case eventRoute(r.Event.Source, r.Event.Kind) == routeGitHub:
		return reduceGitHub(ctx, tx, r)
	default:
		return nil
	}
}

func supported(s core.Source, k core.Kind) bool {
	return eventRoute(s, k) != routeUnsupported
}

type route uint8

const (
	routeUnsupported route = iota
	routeCommit
	routeCheck
	routeSession
	routeReview
	routeGitHub
)

func eventRoute(s core.Source, k core.Kind) route {
	if s == core.SourceGit && k == core.KindCommitRecorded {
		return routeCommit
	}
	if s == core.SourceChecks && k == core.KindCheckRun {
		return routeCheck
	}
	if s == core.SourceSessions && k == core.KindSessionObserved {
		return routeSession
	}
	if s == core.SourceReviews && k == core.KindReviewVerdict {
		return routeReview
	}
	if s == core.SourceGitHub && (k == core.KindGitHubWorkflow || k == core.KindGitHubDeployment) {
		return routeGitHub
	}
	return routeUnsupported
}

type githubWorkflowPayload struct {
	Repository string    `json:"repository"`
	RunID      int64     `json:"run_id"`
	Workflow   string    `json:"workflow"`
	HeadSHA    string    `json:"head_sha"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type githubDeploymentPayload struct {
	Repository   string    `json:"repository"`
	DeploymentID int64     `json:"deployment_id"`
	Environment  string    `json:"environment"`
	SHA          string    `json:"sha"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type githubDeliveryRow struct {
	NativeID, Kind, Repository, CommitSHA, Workflow, Environment, Status, Conclusion, URL string
	WorkflowID, DeploymentID                                                              int64
	OccurredAt, FreshnessAt                                                               time.Time
}

func reduceGitHub(ctx context.Context, tx *store.Tx, r store.Record) error {
	var v githubDeliveryRow
	switch r.Event.Kind {
	case core.KindGitHubWorkflow:
		var payload githubWorkflowPayload
		if err := decode(r.Event.Payload, &payload); err != nil {
			return fmt.Errorf("github workflow seq %d: %w", r.Seq, err)
		}
		if err := payload.validate(); err != nil {
			return fmt.Errorf("github workflow seq %d: %w", r.Seq, err)
		}
		v = githubDeliveryRow{NativeID: r.Event.NativeID, Kind: string(r.Event.Kind), Repository: payload.Repository, CommitSHA: payload.HeadSHA, WorkflowID: payload.RunID, Workflow: payload.Workflow, Status: payload.Status, Conclusion: payload.Conclusion, OccurredAt: r.Event.OccurredAt, FreshnessAt: r.Event.RecordedAt, URL: payload.URL}
	case core.KindGitHubDeployment:
		var payload githubDeploymentPayload
		if err := decode(r.Event.Payload, &payload); err != nil {
			return fmt.Errorf("github deployment seq %d: %w", r.Seq, err)
		}
		if err := payload.validate(); err != nil {
			return fmt.Errorf("github deployment seq %d: %w", r.Seq, err)
		}
		v = githubDeliveryRow{NativeID: r.Event.NativeID, Kind: string(r.Event.Kind), Repository: payload.Repository, CommitSHA: payload.SHA, DeploymentID: payload.DeploymentID, Environment: payload.Environment, Status: "observed", OccurredAt: r.Event.OccurredAt, FreshnessAt: r.Event.RecordedAt, URL: payload.URL}
	default:
		return fmt.Errorf("github: unsupported kind %s", r.Event.Kind)
	}
	_, err := tx.Exec(ctx, `INSERT INTO github_delivery_view(native_id,event_id,seq,kind,repository,commit_sha,workflow_id,workflow,deployment_id,environment,status,conclusion,occurred_at,freshness_at,url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT(native_id) DO UPDATE SET event_id=EXCLUDED.event_id,seq=EXCLUDED.seq,kind=EXCLUDED.kind,repository=EXCLUDED.repository,commit_sha=EXCLUDED.commit_sha,workflow_id=EXCLUDED.workflow_id,workflow=EXCLUDED.workflow,deployment_id=EXCLUDED.deployment_id,environment=EXCLUDED.environment,status=EXCLUDED.status,conclusion=EXCLUDED.conclusion,occurred_at=EXCLUDED.occurred_at,freshness_at=EXCLUDED.freshness_at,url=EXCLUDED.url WHERE github_delivery_view.seq < EXCLUDED.seq`, v.NativeID, r.Event.ID.String(), r.Seq, v.Kind, v.Repository, v.CommitSHA, v.WorkflowID, v.Workflow, v.DeploymentID, v.Environment, v.Status, v.Conclusion, v.OccurredAt, v.FreshnessAt, v.URL)
	return err
}

func (v githubWorkflowPayload) validate() error {
	if !validRepository(v.Repository) || v.RunID <= 0 || strings.TrimSpace(v.Workflow) == "" || !gitPattern.MatchString(v.HeadSHA) || strings.TrimSpace(v.Status) == "" || (strings.EqualFold(v.Status, "completed") && strings.TrimSpace(v.Conclusion) == "") || v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() || v.UpdatedAt.Before(v.CreatedAt) {
		return errors.New("invalid or missing GitHub workflow field")
	}
	return nil
}

func (v githubDeploymentPayload) validate() error {
	if !validRepository(v.Repository) || v.DeploymentID <= 0 || strings.TrimSpace(v.Environment) == "" || !gitPattern.MatchString(v.SHA) || v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() || v.UpdatedAt.Before(v.CreatedAt) {
		return errors.New("invalid or missing GitHub deployment field")
	}
	return nil
}

func validRepository(value string) bool {
	parts := strings.Split(value, "/")
	return len(parts) == 2 && validExternalName(parts[0]) && validExternalName(parts[1])
}

func validExternalName(value string) bool {
	if value == "" || value == "." || value == ".." || len(value) > 100 {
		return false
	}
	for _, r := range value {
		valid := r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.'
		if !valid {
			return false
		}
	}
	return true
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

func (v commitPayload) validate() error {
	if !shaPattern.MatchString(v.SHA) || strings.TrimSpace(v.AuthorName) == "" || !emailPattern.MatchString(v.AuthorEmail) || strings.TrimSpace(v.CommitterName) == "" || !emailPattern.MatchString(v.CommitterEmail) || v.CommittedAt.IsZero() || strings.TrimSpace(v.Subject) == "" {
		return errors.New("invalid or missing commit field")
	}
	for _, file := range v.FilesTouched {
		if file == "" || strings.IndexByte(file, 0) >= 0 {
			return errors.New("invalid files_touched")
		}
	}
	for _, citation := range v.CitedDecisions {
		if !decisionPattern.MatchString(citation) {
			return errors.New("invalid cited_decisions")
		}
	}
	return nil
}

type toolPayload struct {
	Go           string `json:"go"`
	GolangCILint string `json:"golangci_lint"`
	Make         string `json:"make"`
}
type checkPayload struct {
	Schema       string      `json:"schema"`
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

type sessionPayload struct {
	SessionID         string     `json:"session_id"`
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	MessageCount      int64      `json:"message_count"`
	ToolCallCount     int64      `json:"tool_call_count"`
	FilesWrittenCount int64      `json:"files_written_count"`
	ParseCoverage     float64    `json:"parse_coverage"`
}

type reviewFinding struct {
	FindingID    string `json:"finding_id"`
	Severity     string `json:"severity"`
	DefectCommit string `json:"defect_commit,omitempty"`
}

type reviewPayload struct {
	Schema         string          `json:"schema"`
	VerdictID      string          `json:"verdict_id"`
	Status         string          `json:"status"`
	ReviewedCommit string          `json:"reviewed_commit"`
	Findings       []reviewFinding `json:"findings"`
	ArtifactPath   string          `json:"artifact_path"`
	ArtifactSHA    string          `json:"artifact_sha"`
}

func (v reviewPayload) validate() error {
	if v.Schema != "vera.verdict.v1" || strings.TrimSpace(v.VerdictID) == "" || (v.Status != "ACCEPTABLE" && v.Status != "NEEDS_WORK") || !gitPattern.MatchString(v.ReviewedCommit) || !strings.HasPrefix(v.ArtifactPath, "docs/verification/verdicts/") || !strings.HasSuffix(v.ArtifactPath, ".md") || !hexPattern.MatchString(v.ArtifactSHA) {
		return errors.New("invalid or missing review field")
	}
	if v.Status == "NEEDS_WORK" && len(v.Findings) == 0 {
		return errors.New("NEEDS_WORK review has no findings")
	}
	for _, finding := range v.Findings {
		if strings.TrimSpace(finding.FindingID) == "" || (finding.Severity != "HIGH" && finding.Severity != "MED" && finding.Severity != "LOW") || (finding.DefectCommit != "" && !gitPattern.MatchString(finding.DefectCommit)) {
			return errors.New("invalid review finding")
		}
	}
	return nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (v sessionPayload) validate() error {
	if strings.TrimSpace(v.SessionID) == "" || v.MessageCount < 0 || v.ToolCallCount < 0 || v.FilesWrittenCount < 0 || v.ParseCoverage < 0 || v.ParseCoverage > 1 {
		return errors.New("invalid or missing session field")
	}
	if v.StartedAt != nil && v.FinishedAt != nil && v.FinishedAt.Before(*v.StartedAt) {
		return errors.New("session timestamps are invalid")
	}
	return nil
}

func (v checkPayload) validate() error {
	if v.Schema != "vera.witness.v1" || !ulidPattern.MatchString(v.RunID) || !makeCommandPattern.MatchString(v.Command) || v.ExitCode < 0 || v.ExitCode > 255 || v.StartedAt.IsZero() || v.FinishedAt.IsZero() || v.FinishedAt.Before(v.StartedAt) || v.DurationMS < 0 || !hexPattern.MatchString(v.OutputSHA256) || !gitPattern.MatchString(v.GitSHA) || strings.TrimSpace(v.ToolVersions.Go) == "" || strings.TrimSpace(v.ToolVersions.GolangCILint) == "" || strings.TrimSpace(v.ToolVersions.Make) == "" {
		return errors.New("invalid or missing check field")
	}
	return nil
}

func decode(raw []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err == nil {
		return errors.New("trailing JSON")
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON: %w", err)
	}
	if err := requireFields(raw, out); err != nil {
		return err
	}
	return nil
}

func requireFields(raw []byte, out any) error {
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	var required []string
	switch out.(type) {
	case *commitPayload:
		required = []string{"sha", "author_name", "author_email", "committer_name", "committer_email", "committed_at", "subject", "files_touched", "cited_decisions"}
	case *checkPayload:
		required = []string{"schema", "run_id", "command", "exit_code", "started_at", "finished_at", "duration_ms", "output_sha256", "git_sha", "git_dirty", "tool_versions"}
	case *sessionPayload:
		required = []string{"session_id", "message_count", "tool_call_count", "files_written_count", "parse_coverage"}
	case *reviewPayload:
		required = []string{"schema", "verdict_id", "status", "reviewed_commit", "findings", "artifact_path", "artifact_sha"}
	case *githubWorkflowPayload:
		required = []string{"repository", "run_id", "workflow", "head_sha", "status", "conclusion", "url", "created_at", "updated_at"}
	case *githubDeploymentPayload:
		required = []string{"repository", "deployment_id", "environment", "sha", "url", "created_at", "updated_at"}
	}
	for _, name := range required {
		value, ok := fields[name]
		if !ok || (bytes.Equal(bytes.TrimSpace(value), []byte("null")) && name != "files_touched" && name != "cited_decisions") {
			return fmt.Errorf("missing or null field %q", name)
		}
	}
	return nil
}

func (p *Projector) Snapshot(ctx context.Context, s *store.Store) (Snapshot, error) {
	if err := p.Ensure(ctx, s); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot ensure: %w", err)
	}
	result := Snapshot{Tables: make(map[string][]string)}
	queries := map[string]string{"commits_view": "SELECT sha,event_id,seq,author_name,author_email,committer_name,committer_email,committed_at,subject,files_touched,cited_decisions FROM commits_view", "checks_view": "SELECT run_id,event_id,seq,command,exit_code,started_at,finished_at,duration_ms,output_sha256,git_sha,git_dirty,tool_versions FROM checks_view", "sessions_view": "SELECT session_id,event_id,seq,started_at,finished_at,message_count,tool_call_count,files_written_count,parse_coverage FROM sessions_view", "reviews_view": "SELECT finding_id,verdict_id,event_id,seq,status,severity,reviewed_commit,defect_commit,artifact_path,artifact_sha FROM reviews_view", "github_delivery_view": "SELECT native_id,event_id,seq,kind,repository,commit_sha,workflow_id,workflow,deployment_id,environment,status,conclusion,occurred_at,freshness_at,url FROM github_delivery_view"}
	for name, q := range queries {
		var rowsData []map[string]any
		if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
			rows, err := tx.Query(ctx, q)
			if err != nil {
				return err
			}
			defer rows.Close()
			rowsData, err = readSnapshotRows(name, rows)
			return err
		}); err != nil {
			return Snapshot{}, err
		}
		if err := snapshotDigestRows(result.Tables, name, rowsData); err != nil {
			return Snapshot{}, err
		}
	}
	return result, nil
}

func appendSnapshotDigests(tables map[string][]string, name string, rowsData []map[string]any) error {
	for _, row := range rowsData {
		digest, err := snapshotRowDigest(row)
		if err != nil {
			return err
		}
		tables[name] = append(tables[name], digest)
	}
	sort.Strings(tables[name])
	return nil
}

func readSnapshotRows(name string, rows snapshotRows) ([]map[string]any, error) {
	var rowsData []map[string]any
	for rows.Next() {
		var a []any
		switch name {
		case "commits_view":
			a = make([]any, 11)
		case "checks_view":
			a = make([]any, 12)
		case "sessions_view":
			a = make([]any, 9)
		case "reviews_view":
			a = make([]any, 10)
		case "github_delivery_view":
			a = make([]any, 15)
		default:
			a = make([]any, 6)
		}
		ptr := make([]any, len(a))
		for i := range a {
			ptr[i] = &a[i]
		}
		if err := rows.Scan(ptr...); err != nil {
			return nil, err
		}
		m := map[string]any{}
		for i, v := range a {
			if isJSONColumn(name, i) {
				canonical, err := canonicalJSONValue(v)
				if err != nil {
					return nil, err
				}
				v = json.RawMessage(canonical)
			}
			m[fmt.Sprintf("c%d", i)] = v
		}
		rowsData = append(rowsData, m)
	}
	return rowsData, rows.Err()
}

func snapshotRowDigest(row map[string]any) (string, error) {
	raw, err := json.Marshal(row)
	if err != nil {
		return "", err
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canonical)
	return hex.EncodeToString(h[:]), nil
}

func isJSONColumn(table string, index int) bool {
	return (table == "commits_view" && (index == 9 || index == 10)) || (table == "checks_view" && index == 11)
}

func canonicalJSONValue(value any) ([]byte, error) {
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("JSON column has unexpected type: %w", err)
		}
		raw = encoded
	}
	return core.Canonicalize(json.RawMessage(raw))
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
