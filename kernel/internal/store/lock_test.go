package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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

func TestConfig_RejectsNestedDataAndRuntime(t *testing.T) {
	root := t.TempDir()
	_, err := (Config{Root: root, DataDir: filepath.Join(root, "db"), RuntimeDir: filepath.Join(root, "db", "runtime")}).normalized()
	if !errors.Is(err, ErrConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}
