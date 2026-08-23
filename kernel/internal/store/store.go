package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

var (
	ErrStopIteration = errors.New("store: stop iteration")
	ErrMigrate       = errors.New("store: migration failed")
	ErrLedgerWrite   = errors.New("store: ledger tables are append-only")
	ErrKindConflict  = errors.New("store: same subject and payload with a different kind")
)

type Record struct {
	Seq   int64
	Event core.Event
}
type Store struct {
	pool   *pgxpool.Pool
	lock   *ledgerLock
	cfg    Config
	closed bool
}
type Sync struct {
	store    *Store
	id       int64
	appended int64
	finished bool
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	cfg, err := cfg.normalized()
	if err != nil {
		return nil, err
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("%w: embedded postgres is not wired yet; set DatabaseURL", ErrConfig)
	}
	lock, err := acquireLock(cfg)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		_ = lock.close()
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		_ = lock.close()
		return nil, err
	}
	if err = migrate(ctx, pool); err != nil {
		pool.Close()
		_ = lock.close()
		return nil, fmt.Errorf("%w: %v", ErrMigrate, err)
	}
	return &Store{pool: pool, lock: lock, cfg: cfg}, nil
}
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS events (seq BIGSERIAL PRIMARY KEY,event_id TEXT NOT NULL UNIQUE,source TEXT NOT NULL,native_id TEXT NOT NULL,kind TEXT NOT NULL,occurred_at TIMESTAMPTZ NOT NULL,recorded_at TIMESTAMPTZ NOT NULL,payload JSON NOT NULL,content_sha TEXT NOT NULL,connector_version TEXT NOT NULL); CREATE UNIQUE INDEX IF NOT EXISTS events_idempotency ON events(source,native_id,content_sha); CREATE TABLE IF NOT EXISTS sync_runs (id BIGSERIAL PRIMARY KEY,connector TEXT NOT NULL,cursor_json JSON,started_at TIMESTAMPTZ NOT NULL,finished_at TIMESTAMPTZ,events_appended BIGINT NOT NULL DEFAULT 0,error TEXT)`)
	return err
}
func (s *Store) usable() error {
	if s == nil || s.closed {
		return ErrClosed
	}
	if err := s.lock.ownsPath(); err != nil {
		return err
	}
	return nil
}
func (s *Store) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	s.pool.Close()
	return s.lock.close()
}
func (s *Store) Lock() LockInfo {
	if s == nil || s.lock == nil || s.closed {
		return LockInfo{}
	}
	return LockInfo{Path: s.lock.path, PID: s.lock.info.PID, AcquiredAt: time.Unix(s.lock.info.AcquiredAt, 0)}
}

type LockInfo struct {
	Path       string
	PID        int
	AcquiredAt time.Time
}

func (s *Store) BeginSync(ctx context.Context, connector string) (*Sync, error) {
	if err := s.usable(); err != nil {
		return nil, err
	}
	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO sync_runs(connector,started_at) VALUES($1,$2) RETURNING id`, connector, s.cfg.Now().UTC()).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Sync{store: s, id: id}, nil
}
func (sy *Sync) Append(ctx context.Context, e core.Event) (Record, bool, error) {
	if sy == nil || sy.finished {
		return Record{}, false, ErrClosed
	}
	if err := sy.store.usable(); err != nil {
		return Record{}, false, err
	}
	if err := e.Validate(); err != nil {
		return Record{}, false, err
	}
	var r Record
	err := sy.store.pool.QueryRow(ctx, `INSERT INTO events(event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(source,native_id,content_sha) DO NOTHING RETURNING seq`, e.ID.String(), e.Source, e.NativeID, e.Kind, e.OccurredAt, e.RecordedAt, e.Payload, e.ContentSHA, e.ConnectorVersion).Scan(&r.Seq)
	inserted := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, err
	}
	if !inserted {
		err = sy.store.pool.QueryRow(ctx, `SELECT seq,event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version FROM events WHERE source=$1 AND native_id=$2 AND content_sha=$3`, e.Source, e.NativeID, e.ContentSHA).Scan(&r.Seq, &e.ID, &e.Source, &e.NativeID, &e.Kind, &e.OccurredAt, &e.RecordedAt, &e.Payload, &e.ContentSHA, &e.ConnectorVersion)
		if err != nil {
			return Record{}, false, err
		}
	} else {
		r.Event = e
		sy.appended++
	}
	if !inserted {
		r.Event = e
	}
	return r, inserted, nil
}
func (sy *Sync) Appended() int64 {
	if sy == nil {
		return 0
	}
	return sy.appended
}
func (sy *Sync) Finish(ctx context.Context, cursor json.RawMessage, cause error) error {
	if sy == nil || sy.finished {
		return nil
	}
	sy.finished = true
	var msg *string
	if cause != nil {
		s := cause.Error()
		msg = &s
	}
	_, err := sy.store.pool.Exec(ctx, `UPDATE sync_runs SET cursor_json=$1,finished_at=$2,events_appended=$3,error=$4 WHERE id=$5`, cursor, sy.store.cfg.Now().UTC(), sy.appended, msg, sy.id)
	return err
}

type Filter struct {
	Source        core.Source
	Kind          core.Kind
	SinceSeq      int64
	OccurredAfter time.Time
	Limit         int
}

func (s *Store) ReadEvents(ctx context.Context, f Filter, yield func(Record) error) error {
	if err := s.usable(); err != nil {
		return err
	}
	if f.Limit < 0 {
		return ErrConfig
	}
	q := `SELECT seq,event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version FROM events WHERE seq>$1`
	args := []any{f.SinceSeq}
	if f.Source != "" {
		q += fmt.Sprintf(` AND source=$%d`, len(args)+1)
		args = append(args, f.Source)
	}
	if f.Kind != "" {
		q += fmt.Sprintf(` AND kind=$%d`, len(args)+1)
		args = append(args, f.Kind)
	}
	if !f.OccurredAfter.IsZero() {
		q += fmt.Sprintf(` AND occurred_at>$%d`, len(args)+1)
		args = append(args, f.OccurredAfter)
	}
	q += ` ORDER BY seq`
	if f.Limit > 0 {
		q += fmt.Sprintf(` LIMIT %d`, f.Limit)
	}
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r Record
		var id string
		err = rows.Scan(&r.Seq, &id, &r.Event.Source, &r.Event.NativeID, &r.Event.Kind, &r.Event.OccurredAt, &r.Event.RecordedAt, &r.Event.Payload, &r.Event.ContentSHA, &r.Event.ConnectorVersion)
		if err != nil {
			return err
		}
		r.Event.ID, err = core.ParseEventID(id)
		if err != nil {
			return err
		}
		if err = yield(r); errors.Is(err, ErrStopIteration) {
			return nil
		} else if err != nil {
			return fmt.Errorf("read events: %w", err)
		}
	}
	return rows.Err()
}
