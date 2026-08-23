package store

import (
	"context"
	"errors"
	"testing"

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
	if err := (*Tx)(nil).QueryRow(ctx, "SELECT 1").Scan(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	var rows *Rows
	if rows.Next() || !errors.Is(rows.Err(), ErrClosed) || !errors.Is(rows.Scan(), ErrClosed) {
		t.Fatal("nil rows")
	}
	rows.Close()
	var row *Row
	if !errors.Is(row.Scan(), ErrClosed) {
		t.Fatal("nil row")
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
}

func validSurfaceEvent(t *testing.T) core.Event { t.Helper(); return core.Event{} }
