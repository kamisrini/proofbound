package store

import (
	"errors"
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
	if err := l.ownsPath(); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_RejectsNestedDataAndRuntime(t *testing.T) {
	root := t.TempDir()
	_, err := (Config{Root: root, DataDir: filepath.Join(root, "db"), RuntimeDir: filepath.Join(root, "db", "runtime")}).normalized()
	if !errors.Is(err, ErrConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}
