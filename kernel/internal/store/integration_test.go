//go:build integration

package store

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

func TestStoreAppendDuplicateRevisionAndRead(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}
	resetIntegrationDatabase(t, url)
	s, err := Open(context.Background(), Config{Root: t.TempDir(), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	nativeID := "commit-" + time.Now().UTC().Format("20060102150405.000000000")
	e, err := g.NewEvent(core.NewEventParams{Source: core.SourceGit, NativeID: nativeID, Kind: core.KindCommitRecorded, OccurredAt: time.Unix(99, 0), Payload: json.RawMessage(`{"sha":"a"}`), ConnectorVersion: "test"})
	if err != nil {
		t.Fatal(err)
	}
	sy, err := s.BeginSync(context.Background(), "git")
	if err != nil {
		t.Fatal(err)
	}
	first, inserted, err := sy.Append(context.Background(), e)
	if err != nil || !inserted {
		t.Fatalf("first append: %+v %v %v", first, inserted, err)
	}
	dup, inserted, err := sy.Append(context.Background(), e)
	if err != nil || inserted || dup.Seq != first.Seq || dup.Event.ID != first.Event.ID {
		t.Fatalf("duplicate: %+v %v %v", dup, inserted, err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	var got []Record
	if err := s.ReadEvents(context.Background(), Filter{SinceSeq: first.Seq - 1}, func(r Record) error { got = append(got, r); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Seq != first.Seq {
		t.Fatalf("read=%+v", got)
	}
}

func TestStoreRevisionKindConflictAndFilters(t *testing.T) {
	s := integrationStore(t)
	defer s.Close()
	e := integrationEvent(t, "revision")
	sy, err := s.BeginSync(context.Background(), "git")
	if err != nil {
		t.Fatal(err)
	}
	first, inserted, err := sy.Append(context.Background(), e)
	if err != nil || !inserted {
		t.Fatalf("first=%v %v", err, inserted)
	}
	revision := integrationEvent(t, "revision")
	revision.NativeID = e.NativeID
	revision.Payload = json.RawMessage(`{"sha":"changed"}`)
	revision.ContentSHA, err = core.ContentSHA(revision.Payload)
	if err != nil {
		t.Fatal(err)
	}
	second, inserted, err := sy.Append(context.Background(), revision)
	if err != nil || !inserted || second.Seq <= first.Seq {
		t.Fatalf("revision=%+v %v %v", second, inserted, err)
	}
	conflict := e
	conflict.Kind = core.KindCheckRun
	if _, _, err := sy.Append(context.Background(), conflict); !errors.Is(err, ErrKindConflict) {
		t.Fatalf("kind conflict=%v", err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.ReadEvents(context.Background(), Filter{Source: core.SourceGit, Kind: core.KindCommitRecorded, SinceSeq: first.Seq - 1, Limit: 1}, func(Record) error { count++; return nil }); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("filter count=%d", count)
	}
}

func TestStoreWithTxCommitRollbackAndStop(t *testing.T) {
	s := integrationStore(t)
	defer s.Close()
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, "CREATE TEMP TABLE proofbound_tx_probe (n integer)")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		for _, q := range []string{"update events set source='x'", "delete from events", "truncate events"} {
			if _, err := tx.Exec(ctx, q); !errors.Is(err, ErrLedgerWrite) {
				return fmt.Errorf("%q: expected ErrLedgerWrite, got %v", q, err)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("stop")
	if err := s.WithTx(context.Background(), func(context.Context, *Tx) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatalf("rollback error=%v", err)
	}
	if err := s.ReadEvents(context.Background(), Filter{Limit: 1}, func(Record) error { return ErrStopIteration }); err != nil {
		t.Fatalf("stop iteration=%v", err)
	}
}

func TestStoreTransactionWrappersUseAndCloseLiveRows(t *testing.T) {
	s := integrationStore(t)
	defer s.Close()
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		row := tx.QueryRow(ctx, "SELECT 42")
		var answer int
		if err := row.Scan(&answer); err != nil || answer != 42 {
			return fmt.Errorf("row scan: answer=%d error=%v", answer, err)
		}

		rows, err := tx.Query(ctx, "SELECT generate_series(1, 3)")
		if err != nil {
			return err
		}
		if !rows.Next() {
			return fmt.Errorf("expected first row: %v", rows.Err())
		}
		var first int
		if err := rows.Scan(&first); err != nil || first != 1 {
			return fmt.Errorf("rows scan: first=%d error=%v", first, err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("rows error before close: %v", err)
		}
		rows.Close()
		if rows.Next() {
			return errors.New("rows remained open")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReadEventsPagingIsReentrant(t *testing.T) {
	s := integrationStoreWithConfig(t, Config{MaxConns: 1})
	defer s.Close()
	sy, err := s.BeginSync(context.Background(), "paging")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 260; i++ {
		e := integrationEvent(t, fmt.Sprintf("paging-%d", i))
		if _, _, err := sy.Append(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	count := 0
	if err := s.ReadEvents(context.Background(), Filter{Source: core.SourceGit}, func(r Record) error {
		count++
		return s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
			var n int
			return tx.QueryRow(ctx, "SELECT count(*) FROM events WHERE seq <= $1", r.Seq).Scan(&n)
		})
	}); err != nil {
		t.Fatal(err)
	}
	if count != 260 {
		t.Fatalf("count=%d", count)
	}
}

func TestStoreImportReplayRoutes(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		s := integrationStore(t)
		defer s.Close()
		if err := s.ImportReplayRecords(context.Background(), nil); !errors.Is(err, ErrConfig) {
			t.Fatalf("error=%v", err)
		}
	})
	t.Run("validation", func(t *testing.T) {
		s := integrationReplayStore(t, Config{})
		defer s.Close()
		e := integrationEvent(t, "replay-validation")
		bad := e
		bad.Payload = nil
		for name, records := range map[string][]Record{
			"invalid event":       {{Seq: 1, Event: bad}},
			"decreasing sequence": {{Seq: 2, Event: e}, {Seq: 1, Event: integrationEvent(t, "replay-decreasing")}},
			"duplicate event":     {{Seq: 1, Event: e}, {Seq: 2, Event: e}},
		} {
			t.Run(name, func(t *testing.T) {
				if err := s.ImportReplayRecords(context.Background(), records); !errors.Is(err, ErrConfig) {
					t.Fatalf("error=%v", err)
				}
			})
		}
	})
	t.Run("begin failure", func(t *testing.T) {
		s := integrationReplayStore(t, Config{})
		defer s.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := s.ImportReplayRecords(ctx, []Record{{Seq: 1, Event: integrationEvent(t, "replay-begin")}}); err == nil {
			t.Fatal("canceled transaction began")
		}
	})
	t.Run("insert failure rolls back", func(t *testing.T) {
		s := integrationReplayStore(t, Config{MaxConns: 1})
		defer s.Close()
		e := integrationEvent(t, "replay-insert")
		appendStoreEvent(t, s, e)
		if err := s.ImportReplayRecords(context.Background(), []Record{{Seq: 99, Event: e}}); err == nil {
			t.Fatal("duplicate event was accepted")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		count := 0
		if err := s.ReadEvents(ctx, Filter{}, func(Record) error { count++; return nil }); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("count=%d", count)
		}
	})
	t.Run("sequence update failure rolls back", func(t *testing.T) {
		s := integrationReplayStore(t, Config{MaxConns: 1})
		defer s.Close()
		defer dropEventSequence(t, s)()
		if err := s.ImportReplayRecords(context.Background(), []Record{{Seq: 1, Event: integrationEvent(t, "replay-sequence")}}); err == nil {
			t.Fatal("missing sequence was accepted")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.ReadEvents(ctx, Filter{}, func(Record) error { return nil }); err != nil {
			t.Fatal(err)
		}
	})
}

func TestStoreImportHistoricalEvidenceRoutes(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		s := integrationStore(t)
		defer s.Close()
		if err := s.ImportHistoricalEvidenceRecords(context.Background(), nil); !errors.Is(err, ErrConfig) {
			t.Fatalf("error=%v", err)
		}
	})
	t.Run("validation", func(t *testing.T) {
		s := integrationHistoricalStore(t, Config{})
		defer s.Close()
		e := integrationEvent(t, "historical-validation")
		bad := e
		bad.Payload = nil
		if err := s.ImportHistoricalEvidenceRecords(context.Background(), []Record{{Seq: 1, Event: bad}}); !errors.Is(err, ErrConfig) {
			t.Fatalf("error=%v", err)
		}
	})
	t.Run("begin failure", func(t *testing.T) {
		s := integrationHistoricalStore(t, Config{})
		defer s.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := s.ImportHistoricalEvidenceRecords(ctx, []Record{{Seq: 1, Event: integrationEvent(t, "historical-begin")}}); err == nil {
			t.Fatal("canceled transaction began")
		}
	})
	t.Run("insert failure rolls back", func(t *testing.T) {
		s := integrationHistoricalStore(t, Config{MaxConns: 1})
		defer s.Close()
		e := integrationEvent(t, "historical-insert")
		appendStoreEvent(t, s, e)
		if err := s.ImportHistoricalEvidenceRecords(context.Background(), []Record{{Seq: 99, Event: e}}); err == nil {
			t.Fatal("duplicate event was accepted")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.ReadEvents(ctx, Filter{}, func(Record) error { return nil }); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("sequence update failure rolls back", func(t *testing.T) {
		s := integrationHistoricalStore(t, Config{MaxConns: 1})
		defer s.Close()
		defer dropEventSequence(t, s)()
		if err := s.ImportHistoricalEvidenceRecords(context.Background(), []Record{{Seq: 1, Event: integrationEvent(t, "historical-sequence")}}); err == nil {
			t.Fatal("missing sequence was accepted")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.ReadEvents(ctx, Filter{}, func(Record) error { return nil }); err != nil {
			t.Fatal(err)
		}
	})
}

func TestStoreMigrationFailureRoutes(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required")
	}
	newPool := func(t *testing.T) *pgxpool.Pool {
		t.Helper()
		pool, err := pgxpool.New(context.Background(), url)
		if err != nil {
			t.Fatal(err)
		}
		return pool
	}
	t.Run("acquire failure", func(t *testing.T) {
		pool := newPool(t)
		pool.Close()
		if err := migrate(context.Background(), pool); err == nil {
			t.Fatal("migration acquired a closed pool")
		}
	})
	t.Run("advisory lock failure", func(t *testing.T) {
		pool := newPool(t)
		defer pool.Close()
		original := migrationAdvisoryLockSQL
		migrationAdvisoryLockSQL = "SELECT definitely_missing_migration_function()"
		defer func() { migrationAdvisoryLockSQL = original }()
		if err := migrate(context.Background(), pool); err == nil || !strings.Contains(err.Error(), "definitely_missing_migration_function") {
			t.Fatal("invalid advisory-lock query was accepted")
		}
	})
	t.Run("ledger SQL failure", func(t *testing.T) {
		pool := newPool(t)
		defer pool.Close()
		original := ledgerSQL
		ledgerSQL = []byte("SELECT definitely_missing_migration_function()")
		defer func() { ledgerSQL = original }()
		if err := migrate(context.Background(), pool); err == nil || !strings.Contains(err.Error(), "definitely_missing_migration_function") {
			t.Fatal("invalid ledger SQL was accepted")
		}
	})
}

func TestOpenEmbeddedIdentityWriteFailure(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	runtimeDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(data, 0o700); err != nil {
		t.Fatal(err)
	}
	binaries := os.Getenv("PROOFBOUND_TEST_REPO_ROOT")
	if binaries != "" {
		binaries = filepath.Join(binaries, "kernel", ".proofbound", "pgbin")
	} else {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		for {
			candidate := filepath.Join(wd, ".proofbound", "pgbin")
			if _, err := os.Stat(candidate); err == nil {
				binaries = candidate
				break
			}
			parent := filepath.Dir(wd)
			if parent == wd {
				t.Skip("embedded postgres binaries are unavailable")
			}
			wd = parent
		}
	}
	server := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(55441).DataPath(data).RuntimePath(runtimeDir).BinariesPath(binaries).Username("proofbound").Password("proofbound").Database("proofbound"))
	if err := server.Start(); err != nil {
		t.Skipf("embedded postgres unavailable: %v", err)
	}
	if err := server.Stop(); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(data, ".proofbound-identity")
	if err := os.WriteFile(marker, []byte(proofboundEmbeddedIdentity.marker+"\n"), 0o400); err != nil {
		t.Fatal(err)
	}
	s, err := Open(context.Background(), Config{Root: root, DataDir: data, BinariesDir: binaries, Port: 55441})
	if !errors.Is(err, ErrMigrate) || s != nil {
		t.Fatalf("store=%v error=%v", s, err)
	}
}

func integrationStore(t *testing.T) *Store {
	return integrationStoreWithConfig(t, Config{})
}

func integrationStoreWithConfig(t *testing.T, cfg Config) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required")
	}
	resetIntegrationDatabase(t, url)
	cfg.Root = t.TempDir()
	cfg.DatabaseURL = url
	s, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func integrationReplayStore(t *testing.T, cfg Config) *Store {
	t.Helper()
	root, err := os.MkdirTemp(t.TempDir(), "proofbound-twin-")
	if err != nil {
		t.Fatal(err)
	}
	return integrationStoreAtRoot(t, root, cfg, true, false)
}

func integrationHistoricalStore(t *testing.T, cfg Config) *Store {
	t.Helper()
	return integrationStoreAtRoot(t, t.TempDir(), cfg, false, true)
}

func integrationStoreAtRoot(t *testing.T, root string, cfg Config, replay, historical bool) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required")
	}
	resetIntegrationDatabase(t, url)
	cfg.Root, cfg.DatabaseURL = root, url
	cfg.AllowReplayImport, cfg.AllowHistoricalEvidenceImport = replay, historical
	s, err := Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func appendStoreEvent(t *testing.T, s *Store, e core.Event) {
	t.Helper()
	sy, err := s.BeginSync(context.Background(), "import-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := sy.Append(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func dropEventSequence(t *testing.T, s *Store) func() {
	t.Helper()
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, "DROP SEQUENCE IF EXISTS events_seq_seq CASCADE")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return func() {
		_ = s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
			if _, err := tx.Exec(ctx, "CREATE SEQUENCE IF NOT EXISTS events_seq_seq"); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "ALTER TABLE events ALTER COLUMN seq SET DEFAULT nextval('events_seq_seq')")
			return err
		})
	}
}

func resetIntegrationDatabase(t *testing.T, databaseURL string) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var exists bool
	if err := pool.QueryRow(context.Background(), `SELECT to_regclass('public.events') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		return
	}
	if _, err := pool.Exec(context.Background(), `TRUNCATE events, sync_runs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `DROP TABLE IF EXISTS projection_meta, commits_view, checks_view, sessions_view, reviews_view CASCADE`); err != nil {
		t.Fatal(err)
	}
}
func integrationEvent(t *testing.T, label string) core.Event {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	e, err := g.NewEvent(core.NewEventParams{Source: core.SourceGit, NativeID: "integration-" + label + "-" + time.Now().UTC().Format("150405.000000000"), Kind: core.KindCommitRecorded, OccurredAt: time.Now(), Payload: json.RawMessage(`{"sha":"original"}`), ConnectorVersion: "integration"})
	if err != nil {
		t.Fatal(err)
	}
	return e
}
