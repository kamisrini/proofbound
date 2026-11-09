//go:build integration

package gates

import (
	"net"
	"testing"
)

// freeTestPort reserves an ephemeral TCP port so this package's embedded-postgres
// integration test instances do not contend on the fixed product default
// (store.embeddedPort(0) == 55432, a stable local-CLI DB port asserted by
// store/surface_test.go). Resolves DEF-KERNELCHECK-FLAKE-001 under
// BR-gate-integrity-4b2e08 O-1 (CI-flake-ephemeral-port-7d5f10); product contract unchanged.
func freeTestPort(t *testing.T) uint16 {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve free port: %v", err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatalf("release free port: %v", err)
	}
	return port
}
