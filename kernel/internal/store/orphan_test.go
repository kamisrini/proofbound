package store

import (
	"bufio"
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const orphanChildEnv = "PROOFBOUND_STORE_ORPHAN_CHILD"

// TestOrphanChild is not a test: run as a child process by TestOpen_StopsAnOrphanOfThisDataDir,
// it opens a Store, says so, and blocks until it is killed.
func TestOrphanChild(t *testing.T) {
	spec := os.Getenv(orphanChildEnv)
	if spec == "" {
		return // not running as the child: nothing to do
	}
	root, portText, _ := strings.Cut(spec, "|")
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(context.Background(), Config{Root: root, Port: uint16(port)})
	if err != nil {
		t.Fatalf("child open: %v", err)
	}
	_ = s
	os.Stdout.WriteString("child-ready\n")
	// A deadline, so a parent that dies before killing it does not leave it holding the store.
	time.Sleep(5 * time.Minute)
	os.Exit(1)
}

func listening(port uint16) bool {
	c, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(int(port)), time.Second)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// INV-28 (amended 2026-11-27).
func TestOpen_StopsAnOrphanOfThisDataDir(t *testing.T) {
	if os.Getenv("DATABASE_URL") != "" {
		t.Skip("needs the embedded server")
	}
	ctx := context.Background()

	t.Run("a killed process's server is stopped and the next Open serves", func(t *testing.T) {
		root, port := t.TempDir(), freePort(t)
		killedChild(t, root, port)
		s, err := Open(ctx, Config{Root: root, Port: port})
		if err != nil {
			t.Fatalf("Open after a killed run: %v", err)
		}
		if err = s.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		if listening(port) {
			t.Fatal("after Close no server should remain on the port")
		}
		logText, _ := os.ReadFile(serverLogPath(filepath.Join(root, "db")))
		if !strings.Contains(string(logText), "stopping a server left running on this data directory") {
			t.Fatalf("the server log should say the orphan was stopped:\n%s", logText)
		}
	})

	// A run killed during its first start leaves no identity marker, so the next Open reads the
	// cluster as legacy; the orphan must still be recognized and stopped. (That Open then fails on
	// the half-made cluster, a separate first-start window: what matters here is the port.)
	t.Run("an orphan with no identity marker is still stopped", func(t *testing.T) {
		root, port := t.TempDir(), freePort(t)
		killedChild(t, root, port)
		if err := os.Remove(filepath.Join(root, "db", ".proofbound-identity")); err != nil {
			t.Fatal(err)
		}
		s, err := Open(ctx, Config{Root: root, Port: port})
		if err == nil {
			_ = s.Close()
		} else if strings.Contains(err.Error(), "already listening") {
			t.Fatalf("the orphan was not recognized without its marker: %v", err)
		}
		logText, _ := os.ReadFile(serverLogPath(filepath.Join(root, "db")))
		if !strings.Contains(string(logText), "stopping a server left running on this data directory") {
			t.Fatalf("the server log should say the orphan was stopped:\n%s", logText)
		}
		if listening(port) {
			t.Fatal("no server should remain on the port")
		}
	})

	// The lock sits beside the path it was given, so a symlinked data directory is not
	// protected by the real directory's lock: a live server reached that way is never stopped.
	t.Run("a live server reached through a symlinked data directory is never stopped", func(t *testing.T) {
		rootA, portA := t.TempDir(), freePort(t)
		a, err := Open(ctx, Config{Root: rootA, Port: portA})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = a.Close() }()
		rootB := t.TempDir()
		if err = os.Symlink(filepath.Join(rootA, "db"), filepath.Join(rootB, "db")); err != nil {
			t.Skipf("cannot create a symlink here: %v", err)
		}
		// B shares A's binaries, so a wrong decision to stop A would be carried out, not
		// fail for want of pg_ctl.
		if b, err := Open(ctx, Config{Root: rootB, Port: freePort(t), BinariesDir: filepath.Join(rootA, "pgbin")}); err == nil {
			_ = b.Close()
		}
		if !listening(portA) {
			t.Fatal("A's live server was stopped through a symlinked data directory")
		}
		if err = a.WithTx(ctx, func(context.Context, *Tx) error { return nil }); err != nil {
			t.Fatalf("A must still serve: %v", err)
		}
	})

	t.Run("a server for another data directory is never stopped", func(t *testing.T) {
		rootA, portA := t.TempDir(), freePort(t)
		a, err := Open(ctx, Config{Root: rootA, Port: portA})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = a.Close() }()

		// B is an initialized ledger whose postmaster.pid (planted) names A's live server, as a
		// stale pid file can after its pid and port are reused.
		rootB, portB := t.TempDir(), freePort(t)
		b, err := Open(ctx, Config{Root: rootB, Port: portB})
		if err != nil {
			t.Fatal(err)
		}
		if err = b.Close(); err != nil {
			t.Fatal(err)
		}
		pid, err := os.ReadFile(filepath.Join(rootA, "db", "postmaster.pid"))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(rootB, "db", "postmaster.pid"), pid, 0o600); err != nil {
			t.Fatal(err)
		}
		if b2, err := Open(ctx, Config{Root: rootB, Port: portB}); err == nil {
			_ = b2.Close()
		}
		if !listening(portA) {
			t.Fatal("A's server was stopped by B's Open: a server for another data directory must never be stopped")
		}
		if err = a.WithTx(ctx, func(context.Context, *Tx) error { return nil }); err != nil {
			t.Fatalf("A must still serve: %v", err)
		}
	})
}

// killedChild opens a Store on root in a child process and kills it mid-run: no Close, no
// deferred cleanup, exactly as SIGKILL or SIGPIPE leaves it. It fails unless the child's server
// is still listening afterwards, or the caller's test would prove nothing.
func killedChild(t *testing.T, root string, port uint16) {
	t.Helper()
	child := exec.Command(os.Args[0], "-test.run=^TestOrphanChild$", "-test.count=1")
	child.Env = append(os.Environ(), orphanChildEnv+"="+root+"|"+strconv.Itoa(int(port)))
	out, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	child.Stderr = os.Stderr
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	// Whatever the test does next, stop the server it leaves behind before TempDir is removed
	// (cleanups run last-registered first, and root's TempDir was registered before this).
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = exec.Command(filepath.Join(root, "pgbin", "bin", "pg_ctl"), "stop", "-w", "-m", "immediate", "-D", filepath.Join(root, "db")).Run()
	})
	ready := make(chan bool, 1)
	go func() {
		sc := bufio.NewScanner(out)
		for sc.Scan() {
			if sc.Text() == "child-ready" {
				ready <- true
				return
			}
		}
		ready <- false
	}()
	select {
	case ok := <-ready:
		if !ok {
			_ = child.Process.Kill()
			t.Fatal("child exited before opening its store")
		}
	case <-time.After(3 * time.Minute):
		_ = child.Process.Kill()
		t.Fatal("child did not open its store")
	}
	if err = child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	if !listening(port) {
		t.Fatal("precondition: the killed child's server should still be listening")
	}
}

func freePort(t *testing.T) uint16 {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}
