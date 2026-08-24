package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

func TestSurface_NilAndClosedHandles(t *testing.T) {
	ctx := context.Background()
	if err := (*Store)(nil).usable(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := (*Store)(nil).Close(); err != nil {
		t.Fatal(err)
	}
	if got := (*Store)(nil).Lock(); got != (LockInfo{}) {
		t.Fatalf("lock=%+v", got)
	}
	var noLockStore = &Store{}
	if got := noLockStore.Lock(); got != (LockInfo{}) {
		t.Fatalf("no-lock=%+v", got)
	}
	if _, _, err := (*Sync)(nil).Append(ctx, validSurfaceEvent(t)); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if (*Sync)(nil).Appended() != 0 {
		t.Fatal("nil sync count")
	}
	if err := (*Sync)(nil).Finish(ctx, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := (*Tx)(nil).Exec(ctx, "SELECT 1"); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := (*Tx)(nil).Query(ctx, "SELECT 1"); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := (&Tx{}).Query(ctx, "SELECT 1"); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := (*Tx)(nil).QueryRow(ctx, "SELECT 1").Scan(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := (&Tx{}).QueryRow(ctx, "SELECT 1").Scan(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	var rows *Rows
	if rows.Next() || !errors.Is(rows.Err(), ErrClosed) || !errors.Is(rows.Scan(), ErrClosed) {
		t.Fatal("nil rows")
	}
	emptyRows := &Rows{}
	if emptyRows.Next() || !errors.Is(emptyRows.Err(), ErrClosed) || !errors.Is(emptyRows.Scan(), ErrClosed) {
		t.Fatal("empty rows")
	}
	rows.Close()
	var noFile *ledgerLock
	if err := noFile.close(); err != nil {
		t.Fatal(err)
	}
	var row *Row
	if !errors.Is(row.Scan(), ErrClosed) {
		t.Fatal("nil row")
	}
}

func TestEmbeddedPort_DefaultAndExplicit(t *testing.T) {
	if got := embeddedPort(0); got != 55432 {
		t.Fatalf("default port=%d", got)
	}
	if got := embeddedPort(54321); got != 54321 {
		t.Fatalf("explicit port=%d", got)
	}
}

func TestOpen_FailureRoutesReleaseTheLock(t *testing.T) {
	t.Run("invalid database URL", func(t *testing.T) {
		root := t.TempDir()
		if s, err := Open(context.Background(), Config{Root: root, DatabaseURL: "://invalid"}); err == nil || s != nil {
			t.Fatalf("store=%v error=%v", s, err)
		}
		assertLockAvailable(t, Config{Root: root})
	})

	t.Run("ping failure", func(t *testing.T) {
		root := t.TempDir()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if s, err := Open(ctx, Config{Root: root, DatabaseURL: "postgres://vera:vera@127.0.0.1:1/vera?sslmode=disable"}); err == nil || s != nil {
			t.Fatalf("store=%v error=%v", s, err)
		}
		assertLockAvailable(t, Config{Root: root})
	})

	t.Run("embedded start failure", func(t *testing.T) {
		root := t.TempDir()
		binariesFile := filepath.Join(root, "not-a-directory")
		if err := os.WriteFile(binariesFile, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		s, err := Open(context.Background(), Config{Root: root, BinariesDir: binariesFile})
		if !errors.Is(err, ErrMigrate) || s != nil {
			t.Fatalf("store=%v error=%v", s, err)
		}
		assertLockAvailable(t, Config{Root: root, BinariesDir: binariesFile})
	})
}

func assertLockAvailable(t *testing.T, cfg Config) {
	t.Helper()
	c, err := cfg.normalized()
	if err != nil {
		t.Fatal(err)
	}
	l, err := acquireLock(c)
	if err != nil {
		t.Fatalf("lock remained held: %v", err)
	}
	if err := l.close(); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_DefaultsAndLockAssertion(t *testing.T) {
	c, err := (Config{Root: t.TempDir()}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	if c.MaxConns != 4 || c.Now == nil || c.LockPath == "" {
		t.Fatalf("defaults=%+v", c)
	}
	if _, err := (Config{Root: t.TempDir(), LockPath: "/wrong"}).normalized(); !errors.Is(err, ErrConfig) {
		t.Fatal(err)
	}
	root := t.TempDir()
	data := root + "/data"
	run := root + "/run"
	bin := root + "/bin"
	c, err = (Config{Root: root, DataDir: data, RuntimeDir: run, BinariesDir: bin}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	if c.DataDir != data || c.RuntimeDir != run || c.BinariesDir != bin {
		t.Fatalf("paths=%+v", c)
	}
}

func validSurfaceEvent(t *testing.T) core.Event { t.Helper(); return core.Event{} }
