package twin

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// startLiveProcess copies srcBinary to a temp path, makes it executable, launches it so it blocks,
// and returns the running pid plus the copy's path. The process is killed on test cleanup.
func startLiveProcess(t *testing.T, srcBinary string) (pid int, copyPath string) {
	t.Helper()
	data, err := os.ReadFile(srcBinary)
	if err != nil {
		t.Skipf("cannot read %s: %v", srcBinary, err)
	}
	copyPath = filepath.Join(t.TempDir(), "proc-under-test")
	if err := os.WriteFile(copyPath, data, 0o755); err != nil {
		t.Fatalf("write copy: %v", err)
	}
	cmd := exec.Command(copyPath, "600") // sleep 600s
	if err := cmd.Start(); err != nil {
		t.Skipf("cannot start %s: %v", copyPath, err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	// Give the kernel a moment to point /proc/<pid>/exe at the image.
	deadline := time.Now().Add(2 * time.Second)
	exe := filepath.Join("/proc", itoaTest(cmd.Process.Pid), "exe")
	for time.Now().Before(deadline) {
		if _, err := os.Readlink(exe); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cmd.Process.Pid, copyPath
}

func itoaTest(p int) string {
	if p == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for p > 0 {
		i--
		b[i] = byte('0' + p%10)
		p /= 10
	}
	return string(b[i:])
}

// TestWitnessLaunchedProcessCalibration is the §5.1 direct calibration of TWIN-INV-18 (N5): it
// points the witness at a LIVE process and proves (a) the positive case returns the /proc-resolved
// digest with no error, and (b) a byte-divergent expected digest returns the DISTINCT mismatch
// error. No production swap seam is required.
func TestWitnessLaunchedProcessCalibration(t *testing.T) {
	const sleepBin = "/bin/sleep"
	pid, copyPath := startLiveProcess(t, sleepBin)

	wantSHA, err := fileSHA256(copyPath)
	if err != nil {
		t.Fatalf("hash copy: %v", err)
	}

	// Positive control: the witness resolves and matches the running image.
	resolved, err := witnessLaunchedProcess(pid, wantSHA)
	if err != nil {
		t.Fatalf("positive: unexpected error: %v", err)
	}
	if resolved != wantSHA {
		t.Fatalf("positive: resolved %s want %s", resolved, wantSHA)
	}

	// Negative control: a byte-divergent expected digest (a different real binary) must fail with
	// the DISTINCT launched-process mismatch error, proving the witness discriminates the running
	// image and does not inherit PASS.
	otherSHA := byteDivergentSHA(t, wantSHA)
	_, err = witnessLaunchedProcess(pid, otherSHA)
	if err == nil {
		t.Fatal("negative: expected mismatch error, got nil")
	}
	if !errors.Is(err, errLaunchedProcessDigestMismatch) {
		t.Fatalf("negative: expected errLaunchedProcessDigestMismatch, got %v", err)
	}
	// It must be DISTINCT from the staging re-hash error (N6): that error is not this sentinel.
	if errors.Is(err, errLaunchedProcessDeleted) {
		t.Fatal("negative: mismatch must not masquerade as the deleted-image error")
	}
}

// byteDivergentSHA returns the sha256 of some real binary whose digest differs from avoid, so the
// negative control is calibrated against a genuine second executable image.
func byteDivergentSHA(t *testing.T, avoid string) string {
	t.Helper()
	for _, cand := range []string{"/bin/true", "/usr/bin/true", "/bin/echo", "/usr/bin/env"} {
		if sha, err := fileSHA256(cand); err == nil && sha != avoid {
			return sha
		}
	}
	t.Skip("no byte-divergent second binary available")
	return ""
}

// TestWitnessLaunchedProcessDeleted proves the "(deleted)" marker is rejected: if the running
// image's on-disk entry is removed, the witness refuses rather than trusting it.
func TestWitnessLaunchedProcessDeleted(t *testing.T) {
	pid, copyPath := startLiveProcess(t, "/bin/sleep")
	if err := os.Remove(copyPath); err != nil {
		t.Fatalf("remove running image: %v", err)
	}
	wantSHA := "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := witnessLaunchedProcess(pid, wantSHA)
	if err == nil {
		t.Fatal("expected deleted-image error, got nil")
	}
	if !errors.Is(err, errLaunchedProcessDeleted) {
		t.Fatalf("expected errLaunchedProcessDeleted, got %v", err)
	}
}

// TestWitnessLaunchedProcessFailClosed proves the witness fails closed on bad inputs (TWIN-INV-7).
func TestWitnessLaunchedProcessFailClosed(t *testing.T) {
	if _, err := witnessLaunchedProcess(0, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("expected error for non-positive pid")
	}
	if _, err := witnessLaunchedProcess(1, "not-a-sha"); err == nil {
		t.Fatal("expected error for malformed expected digest")
	}
}
