package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var (
	ErrConfig   = errors.New("store: invalid config")
	ErrLocked   = errors.New("store: ledger is locked by another process")
	ErrClosed   = errors.New("store: store is closed")
	ErrLockLost = errors.New("store: the ledger lock was lost")
)

type Config struct {
	Root        string
	DataDir     string
	RuntimeDir  string
	BinariesDir string
	LockPath    string
	Port        uint16
	DatabaseURL string
	MaxConns    int
	Now         func() time.Time
}

func (c Config) normalized() (Config, error) {
	if c.Root == "" {
		return Config{}, fmt.Errorf("%w: Root is required", ErrConfig)
	}
	root, err := filepath.Abs(c.Root)
	if err != nil {
		return Config{}, fmt.Errorf("%w: root: %v", ErrConfig, err)
	}
	if c.DataDir == "" {
		c.DataDir = filepath.Join(root, "db")
	}
	if c.RuntimeDir == "" {
		c.RuntimeDir = filepath.Join(root, "pgrun")
	}
	if c.BinariesDir == "" {
		c.BinariesDir = filepath.Join(root, "pgbin")
	}
	for name, path := range map[string]*string{"DataDir": &c.DataDir, "RuntimeDir": &c.RuntimeDir, "BinariesDir": &c.BinariesDir} {
		abs, err := filepath.Abs(*path)
		if err != nil {
			return Config{}, fmt.Errorf("%w: %s", ErrConfig, name)
		}
		*path = abs
	}
	if nested(c.DataDir, c.RuntimeDir) {
		return Config{}, fmt.Errorf("%w: DataDir and RuntimeDir overlap", ErrConfig)
	}
	want := c.DataDir + ".lock"
	if c.LockPath != "" {
		got, e := filepath.Abs(c.LockPath)
		if e != nil || got != want {
			return Config{}, fmt.Errorf("%w: LockPath must be %s", ErrConfig, want)
		}
	}
	c.LockPath = want
	if c.MaxConns == 0 {
		c.MaxConns = 4
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	return c, nil
}
func nested(a, b string) bool { return a == b || stringsHasPath(a, b) || stringsHasPath(b, a) }
func stringsHasPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && (len(rel) < 3 || rel[:3] != ".."+string(filepath.Separator))
}

type lockRecord struct {
	PID        int   `json:"pid"`
	AcquiredAt int64 `json:"acquired_at"`
}

type LockedError struct {
	Path  string
	PID   int
	Since time.Time
}

func (e *LockedError) Error() string { return fmt.Sprintf("%v: %s (pid %d)", ErrLocked, e.Path, e.PID) }
func (e *LockedError) Unwrap() error { return ErrLocked }

type ledgerLock struct {
	file *os.File
	path string
	info lockRecord
}

func acquireLock(c Config) (*ledgerLock, error) {
	if err := os.MkdirAll(filepath.Dir(c.LockPath), 0o755); err != nil {
		return nil, fmt.Errorf("%w: lock directory: %v", ErrConfig, err)
	}
	f, err := os.OpenFile(c.LockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("%w: open lock: %v", ErrConfig, err)
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, &LockedError{Path: c.LockPath}
		}
		return nil, err
	}
	l := &ledgerLock{file: f, path: c.LockPath, info: lockRecord{PID: os.Getpid(), AcquiredAt: c.Now().Unix()}}
	if err = writeLockRecord(f, l.info); err != nil {
		_ = f.Close()
		return nil, err
	}
	return l, nil
}

func writeLockRecord(f *os.File, info lockRecord) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(info)
}
func (l *ledgerLock) ownsPath() error {
	st, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("%w: descriptor: %v", ErrLockLost, err)
	}
	current, err := os.Stat(l.path)
	if err != nil || !os.SameFile(st, current) {
		return ErrLockLost
	}
	return nil
}
func (l *ledgerLock) close() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := l.ownsPath()
	closeErr := l.file.Close()
	l.file = nil
	return errors.Join(err, closeErr)
}
