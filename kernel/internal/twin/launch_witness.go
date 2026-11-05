package twin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// errLaunchedProcessDigestMismatch is the distinct, assertable sentinel returned by
// witnessLaunchedProcess when the running executable image does not equal the declared digest.
// It is deliberately SEPARATE from the staging re-hash error in verifyStagedPostgres
// (verified_archive.go: "launched postgres executable digest mismatch") so a caller can tell a
// runtime launched-process mismatch (TWIN-INV-18) from a pre-launch staged-file mismatch
// (TWIN-INV-14). Unit-3 spec §5.1 N6.
var errLaunchedProcessDigestMismatch = errors.New("twin: launched process executable digest mismatch")

// errLaunchedProcessDeleted is returned when /proc/<pid>/exe carries the "(deleted)" marker: the
// running image's on-disk entry is gone, so it cannot be a trustworthy witness of launched bytes.
var errLaunchedProcessDeleted = errors.New("twin: launched process executable image is deleted")

// witnessLaunchedProcess is the SINGLE implementation and SOLE PRODUCER of resolved_sha256 on the
// proof_binding path (unit-3 spec §5 TWIN-INV-18, §5.1 N8). It resolves /proc/<pid>/exe of the
// running postmaster, rejects the "(deleted)" marker, hashes the ACTUAL running image (via the
// kernel magic symlink, which resolves to the executed inode itself — robust to on-disk path
// games), and compares it to the declared PostgresBinarySHA256. On mismatch it returns
// errLaunchedProcessDigestMismatch, which is distinct from the staging re-hash error (N6). The
// returned digest is always the one just computed from /proc — it is never copied from the
// expected value.
func witnessLaunchedProcess(pid int, expectedSHA256 string) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("twin: launched process witness: invalid pid %d", pid)
	}
	if err := validateSHA256("expected launched process", expectedSHA256); err != nil {
		return "", err
	}
	exe := filepath.Join("/proc", strconv.Itoa(pid), "exe")
	target, err := os.Readlink(exe)
	if err != nil {
		return "", fmt.Errorf("twin: launched process witness: readlink %s: %w", exe, err)
	}
	if strings.HasSuffix(target, " (deleted)") {
		return "", fmt.Errorf("%w: pid %d -> %q", errLaunchedProcessDeleted, pid, target)
	}
	resolved, err := fileSHA256(exe)
	if err != nil {
		return "", fmt.Errorf("twin: launched process witness: hash %s: %w", exe, err)
	}
	if resolved != expectedSHA256 {
		return "", fmt.Errorf("%w: pid %d got %s want %s", errLaunchedProcessDigestMismatch, pid, resolved, expectedSHA256)
	}
	return resolved, nil
}
