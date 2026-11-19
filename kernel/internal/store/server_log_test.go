package store

import (
	"context"
	"io"
	"net"
	"os"
	"strings"
	"testing"
)

// INV-52: the embedded server writes its output to <DataDir>.log, not to stdout.
func TestEmbeddedServerOutputGoesToLogFile(t *testing.T) {
	if os.Getenv("DATABASE_URL") != "" {
		t.Skip("needs the embedded server")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}

	// Capture the process's stdout while the server starts and stops.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	captured := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		captured <- string(b)
	}()
	root := t.TempDir()
	s, openErr := Open(context.Background(), Config{Root: root, Port: port})
	var closeErr error
	if openErr == nil {
		closeErr = s.Close()
	}
	os.Stdout = stdout
	_ = w.Close()
	out := <-captured
	if openErr != nil || closeErr != nil {
		t.Fatalf("open=%v close=%v", openErr, closeErr)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("the embedded server wrote to stdout:\n%s", out)
	}
	log, err := os.ReadFile(serverLogPath(s.cfg.DataDir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "server started") {
		t.Fatalf("the server log lacks the server's own output:\n%s", log)
	}
}
