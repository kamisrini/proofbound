package store

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// stopOrphan stops a server left running on this DataDir by a process that died without
// closing its Store (INV-28, amended 2026-11-27). The caller holds the ledger lock, an OS
// file lock the kernel releases when its holder dies, so no live Proofbound owns this
// DataDir: a server still serving it is an orphan. The pid file is never trusted alone
// (a dead postmaster's pid may now belong to another process); the server is stopped only
// when a connection with this ledger's identity, on the port the pid file names, reports a
// data_directory that is this DataDir. Anything else is left alone, and Start then reports
// whatever is wrong in its own words.
func stopOrphan(ctx context.Context, cfg Config, identity embeddedIdentity) error {
	// The lock proves ownership only of the directory it sits beside. When DataDir is reached
	// through a symlinked leaf, another live process may hold the lock beside the real
	// directory, so the server could be its own: then nothing is stopped.
	if !holdsLockOfRealDir(cfg) {
		return nil
	}
	port, ok := pidFilePort(cfg.DataDir)
	if !ok {
		return nil
	}
	if !servesDataDir(ctx, port, identity, cfg.DataDir) {
		return nil
	}
	stop := exec.Command(filepath.Join(cfg.BinariesDir, "bin", "pg_ctl"), "stop", "-w", "-m", "fast", "-D", cfg.DataDir)
	log := serverLog(cfg.DataDir)
	_, _ = fmt.Fprintf(log, "proofbound: stopping a server left running on this data directory (port %d) by a process that did not close it\n", port)
	stop.Stdout, stop.Stderr = log, log
	if err := stop.Run(); err != nil {
		return fmt.Errorf("stop the server a killed proofbound left running on port %d: %v", port, err)
	}
	return nil
}

// holdsLockOfRealDir reports whether the lock this Store holds is the lock file beside the
// data directory's real (symlink-resolved) path.
func holdsLockOfRealDir(cfg Config) bool {
	resolved, err := filepath.EvalSymlinks(cfg.DataDir)
	if err != nil {
		return false
	}
	return sameFile(resolved+".lock", cfg.LockPath)
}

// pidFilePort reads the port from <DataDir>/postmaster.pid (its fourth line).
func pidFilePort(dataDir string) (uint16, bool) {
	data, err := os.ReadFile(filepath.Join(dataDir, "postmaster.pid"))
	if err != nil {
		return 0, false
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 4 {
		return 0, false
	}
	port, err := strconv.ParseUint(strings.TrimSpace(lines[3]), 10, 16)
	if err != nil || port == 0 {
		return 0, false
	}
	return uint16(port), true
}

// servesDataDir reports whether the server on port answers as one of this store's identities
// and serves dataDir itself. It tries every identity a ledger can have, not only the marker's:
// a run killed during its first start leaves a cluster without the marker, which then reads
// as legacy. It connects to the "postgres" database, which initdb always creates, because the
// ledger's own database may not exist yet.
func servesDataDir(ctx context.Context, port uint16, identity embeddedIdentity, dataDir string) bool {
	for _, id := range []embeddedIdentity{identity, proofboundEmbeddedIdentity, legacyEmbeddedIdentity} {
		if servesAs(ctx, port, id, dataDir) {
			return true
		}
	}
	return false
}

func servesAs(ctx context.Context, port uint16, identity embeddedIdentity, dataDir string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/postgres?sslmode=disable", identity.username, identity.password, port))
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close(context.Background()) }()
	var served string
	if err = conn.QueryRow(ctx, "SELECT current_setting('data_directory')").Scan(&served); err != nil {
		return false
	}
	return sameFile(served, dataDir)
}

func sameFile(a, b string) bool {
	ai, errA := os.Stat(a)
	bi, errB := os.Stat(b)
	if errA != nil || errB != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
