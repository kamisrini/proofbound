package store

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fergusstrange/embedded-postgres"
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

//go:embed migrations/001_ledger.sql
var ledgerMigration embed.FS

var ledgerSQL = func() []byte {
	sql, err := ledgerMigration.ReadFile("migrations/001_ledger.sql")
	if err != nil {
		panic(err)
	}
	return sql
}()

type Store struct {
	pool     *pgxpool.Pool
	lock     *ledgerLock
	embedded *embeddedpostgres.EmbeddedPostgres
	cfg      Config
	closed   bool
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
	lock, err := acquireLock(cfg)
	if err != nil {
		return nil, err
	}
	var server *embeddedpostgres.EmbeddedPostgres
	stopServer := func() {}
	if cfg.DatabaseURL == "" {
		port := embeddedPort(cfg.Port)
		server = embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(uint32(port)).DataPath(cfg.DataDir).RuntimePath(cfg.RuntimeDir).BinariesPath(cfg.BinariesDir).Username("vera").Password("vera").Database("vera"))
		if err = server.Start(); err != nil {
			_ = lock.close()
			return nil, fmt.Errorf("%w: embedded postgres: %v", ErrMigrate, err)
		}
		stopServer = func() { _ = server.Stop() }
		cfg.DatabaseURL = fmt.Sprintf("postgres://vera:vera@127.0.0.1:%d/vera?sslmode=disable", port)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		stopServer()
		_ = lock.close()
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		stopServer()
		_ = lock.close()
		return nil, err
	}
	if err = migrate(ctx, pool); err != nil {
		pool.Close()
		stopServer()
		_ = lock.close()
		return nil, fmt.Errorf("%w: %v", ErrMigrate, err)
	}
	return &Store{pool: pool, lock: lock, cfg: cfg, embedded: server}, nil
}
func embeddedPort(port uint16) uint16 {
	if port == 0 {
		return 55432
	}
	return port
}
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Multiple vera processes may open the same externally managed database at
	// once (for example, package-level integration tests). Serialize the
	// create-if-not-exists migration on the database connection.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('proofbound:ledger-migration'))`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, string(ledgerSQL)); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
	return errors.Join(s.lock.close(), func() error {
		if s.embedded != nil {
			return s.embedded.Stop()
		}
		return nil
	}())
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
	requestedKind := e.Kind
	var r Record
	err := sy.store.pool.QueryRow(ctx, `INSERT INTO events(event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(source,native_id,content_sha) DO NOTHING RETURNING seq`, e.ID.String(), e.Source, e.NativeID, e.Kind, e.OccurredAt, e.RecordedAt, e.Payload, e.ContentSHA, e.ConnectorVersion).Scan(&r.Seq)
	inserted := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, err
	}
	if !inserted {
		var id string
		err = sy.store.pool.QueryRow(ctx, `SELECT seq,event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version FROM events WHERE source=$1 AND native_id=$2 AND content_sha=$3`, e.Source, e.NativeID, e.ContentSHA).Scan(&r.Seq, &id, &e.Source, &e.NativeID, &e.Kind, &e.OccurredAt, &e.RecordedAt, &e.Payload, &e.ContentSHA, &e.ConnectorVersion)
		if err != nil {
			return Record{}, false, err
		}
		e.ID, err = core.ParseEventID(id)
		if err != nil {
			return Record{}, false, err
		}
		if requestedKind != e.Kind {
			return Record{}, false, ErrKindConflict
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

type Tx struct{ tx pgx.Tx }
type Rows struct{ rows pgx.Rows }
type Row struct{ row pgx.Row }

func (tx *Tx) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	if tx == nil || tx.tx == nil {
		return 0, ErrClosed
	}
	if strings.Contains(strings.ToUpper(sql), "UPDATE EVENTS") || strings.Contains(strings.ToUpper(sql), "DELETE FROM EVENTS") || strings.Contains(strings.ToUpper(sql), "TRUNCATE EVENTS") {
		return 0, ErrLedgerWrite
	}
	tag, err := tx.tx.Exec(ctx, sql, args...)
	return tag.RowsAffected(), err
}
func (tx *Tx) Query(ctx context.Context, sql string, args ...any) (*Rows, error) {
	if tx == nil || tx.tx == nil {
		return nil, ErrClosed
	}
	rows, err := tx.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}
func (tx *Tx) QueryRow(ctx context.Context, sql string, args ...any) *Row {
	if tx == nil || tx.tx == nil {
		return &Row{}
	}
	return &Row{row: tx.tx.QueryRow(ctx, sql, args...)}
}
func (r *Rows) Next() bool { return r != nil && r.rows != nil && r.rows.Next() }
func (r *Rows) Scan(dest ...any) error {
	if r == nil || r.rows == nil {
		return ErrClosed
	}
	return r.rows.Scan(dest...)
}
func (r *Rows) Err() error {
	if r == nil || r.rows == nil {
		return ErrClosed
	}
	return r.rows.Err()
}
func (r *Rows) Close() {
	if r != nil && r.rows != nil {
		r.rows.Close()
	}
}
func (r *Row) Scan(dest ...any) error {
	if r == nil || r.row == nil {
		return ErrClosed
	}
	return r.row.Scan(dest...)
}
func (s *Store) WithTx(ctx context.Context, fn func(context.Context, *Tx) error) (err error) {
	if err = s.usable(); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	wrapper := &Tx{tx: tx}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback(ctx)
			panic(v)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()
	return fn(ctx, wrapper)
}

func (s *Store) ReadEvents(ctx context.Context, f Filter, yield func(Record) error) error {
	if err := s.usable(); err != nil {
		return err
	}
	if f.Limit < 0 {
		return ErrConfig
	}
	const pageSize = 256
	cursor, emitted := f.SinceSeq, 0
	for {
		pageLimit := pageSize
		if f.Limit > 0 && f.Limit-emitted < pageLimit {
			pageLimit = f.Limit - emitted
		}
		if pageLimit <= 0 {
			return nil
		}
		q := `SELECT seq,event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version FROM events WHERE seq>$1`
		args := []any{cursor}
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
		q += fmt.Sprintf(` ORDER BY seq LIMIT %d`, pageLimit)
		rows, err := s.pool.Query(ctx, q, args...)
		if err != nil {
			return err
		}
		page := make([]Record, 0, pageLimit)
		for rows.Next() {
			var r Record
			var id string
			err = rows.Scan(&r.Seq, &id, &r.Event.Source, &r.Event.NativeID, &r.Event.Kind, &r.Event.OccurredAt, &r.Event.RecordedAt, &r.Event.Payload, &r.Event.ContentSHA, &r.Event.ConnectorVersion)
			if err != nil {
				rows.Close()
				return err
			}
			r.Event.ID, err = core.ParseEventID(id)
			if err != nil {
				rows.Close()
				return err
			}
			page = append(page, r)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, r := range page {
			if err = yield(r); errors.Is(err, ErrStopIteration) {
				return nil
			} else if err != nil {
				return fmt.Errorf("read events: %w", err)
			}
			cursor, emitted = r.Seq, emitted+1
		}
		if len(page) < pageLimit {
			return nil
		}
	}
}

// ImportReplayRecords imports an explicitly sequenced stream into a store
// created solely for an isolated replay. Ordinary stores cannot use this seam.
func (s *Store) ImportReplayRecords(ctx context.Context, records []Record) error {
	if err := s.usable(); err != nil {
		return err
	}
	if !s.cfg.AllowReplayImport {
		return fmt.Errorf("%w: replay import disabled", ErrConfig)
	}
	seenIDs := make(map[string]struct{}, len(records))
	seenKeys := make(map[string]struct{}, len(records))
	var last int64
	for i, r := range records {
		if r.Seq <= last {
			return fmt.Errorf("%w: replay sequence %d is not increasing", ErrConfig, i)
		}
		if err := r.Event.Validate(); err != nil {
			return fmt.Errorf("%w: replay event %d: %v", ErrConfig, i, err)
		}
		id := r.Event.ID.String()
		key := string(r.Event.Source) + "\x00" + r.Event.NativeID + "\x00" + r.Event.ContentSHA
		if _, ok := seenIDs[id]; ok {
			return fmt.Errorf("%w: duplicate replay event id at %d", ErrConfig, i)
		}
		if _, ok := seenKeys[key]; ok {
			return fmt.Errorf("%w: duplicate replay idempotency key at %d", ErrConfig, i)
		}
		seenIDs[id], seenKeys[key], last = struct{}{}, struct{}{}, r.Seq
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	for _, r := range records {
		_, err = tx.Exec(ctx, `INSERT INTO events(seq,event_id,source,native_id,kind,occurred_at,recorded_at,payload,content_sha,connector_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, r.Seq, r.Event.ID.String(), r.Event.Source, r.Event.NativeID, r.Event.Kind, r.Event.OccurredAt, r.Event.RecordedAt, r.Event.Payload, r.Event.ContentSHA, r.Event.ConnectorVersion)
		if err != nil {
			return err
		}
	}
	if len(records) > 0 {
		_, err = tx.Exec(ctx, `SELECT setval(pg_get_serial_sequence('events','seq'), $1, true)`, records[len(records)-1].Seq)
		if err != nil {
			return err
		}
	}
	err = tx.Commit(ctx)
	return err
}
