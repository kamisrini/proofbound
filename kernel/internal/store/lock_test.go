package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLock_ExclusiveAndPathBound(t *testing.T) {
	root := t.TempDir()
	c, err := (Config{Root: root}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	l, err := acquireLock(c)
	if err != nil {
		t.Fatal(err)
	}
	defer l.close()
	if _, err := acquireLock(c); !errors.Is(err, ErrLocked) {
		t.Fatalf("second lock error=%v", err)
	}
	var locked *LockedError
	if _, err := acquireLock(c); !errors.As(err, &locked) {
		t.Fatalf("missing LockedError: %v", err)
	}
	if err := l.ownsPath(); err != nil {
		t.Fatal(err)
	}
}

func TestLock_PathReplacementIsLoss(t *testing.T) {
	root := t.TempDir()
	c, err := (Config{Root: root}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	l, err := acquireLock(c)
	if err != nil {
		t.Fatal(err)
	}
	defer l.file.Close()
	if err := os.Remove(c.LockPath); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(l.ownsPath(), ErrLockLost) {
		t.Fatal("path replacement was not detected")
	}
}

func TestStore_UsableRejectsALostLock(t *testing.T) {
	c, err := (Config{Root: t.TempDir()}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	l, err := acquireLock(c)
	if err != nil {
		t.Fatal(err)
	}
	defer l.file.Close()
	if err := os.Remove(c.LockPath); err != nil {
		t.Fatal(err)
	}
	if err := (&Store{lock: l}).usable(); !errors.Is(err, ErrLockLost) {
		t.Fatalf("usable error=%v", err)
	}
}

func TestConfig_RejectsNestedDataAndRuntime(t *testing.T) {
	root := t.TempDir()
	_, err := (Config{Root: root, DataDir: filepath.Join(root, "db"), RuntimeDir: filepath.Join(root, "db", "runtime")}).normalized()
	if !errors.Is(err, ErrConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestConfig_RejectsDeeplyNestedDataAndRuntime(t *testing.T) {
	root := t.TempDir()
	_, err := (Config{
		Root:       root,
		DataDir:    filepath.Join(root, "db"),
		RuntimeDir: filepath.Join(root, "db", "one", "two"),
	}).normalized()
	if !errors.Is(err, ErrConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestConfig_AcceptsTheDerivedLockPathExplicitly(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	c, err := (Config{Root: root, DataDir: data, LockPath: data + ".lock"}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	if c.LockPath != data+".lock" {
		t.Fatalf("lock path=%q", c.LockPath)
	}
}

func TestLock_RecordCloseAndReplacementRoutes(t *testing.T) {
	now := time.Unix(1234, 0)
	c, err := (Config{Root: t.TempDir(), Now: func() time.Time { return now }}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	l, err := acquireLock(c)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(c.LockPath)
	if err != nil {
		t.Fatal(err)
	}
	var record lockRecord
	if err := json.Unmarshal(b, &record); err != nil {
		t.Fatalf("lock record: %v (%q)", err, b)
	}
	if record.PID != os.Getpid() || record.AcquiredAt != now.Unix() {
		t.Fatalf("lock record=%+v", record)
	}
	if err := os.Remove(c.LockPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.LockPath, []byte("replacement"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(l.ownsPath(), ErrLockLost) {
		t.Fatal("replacement inode was not detected")
	}
	if err := l.close(); !errors.Is(err, ErrLockLost) {
		t.Fatalf("close error=%v", err)
	}

	if err := os.Remove(c.LockPath); err != nil {
		t.Fatal(err)
	}
	l, err = acquireLock(c)
	if err != nil {
		t.Fatalf("lock was not released by close: %v", err)
	}
	if err := l.close(); err != nil {
		t.Fatal(err)
	}
}
