package checks

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type emitterFixture struct {
	root   string
	script string
	binDir string
	caller string
}

func TestEmitter_RecordsSuccessAndFailure(t *testing.T) {
	success := newEmitterFixture(t)
	witness, _, err := success.run(t, 0)
	if err != nil || witness.ExitCode != 0 {
		t.Fatalf("success witness=%+v error=%v", witness, err)
	}

	failure := newEmitterFixture(t)
	witness, _, err = failure.run(t, 7)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 || witness.ExitCode != 7 {
		t.Fatalf("failure witness=%+v error=%v", witness, err)
	}
}

func TestEmitter_OutputDigestCoversCombinedBytes(t *testing.T) {
	fixture := newEmitterFixture(t)
	witness, output, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(output)
	if got := hex.EncodeToString(digest[:]); witness.OutputSHA256 != got {
		t.Fatalf("output=%q witness digest=%s actual=%s", output, witness.OutputSHA256, got)
	}
	if string(output) != "gate stdout\ngate stderr\n" {
		t.Fatalf("combined output=%q", output)
	}
}

func TestEmitter_WorksWithoutVeraBinary(t *testing.T) {
	fixture := newEmitterFixture(t)
	marker := filepath.Join(fixture.root, "vera-was-invoked")
	writeExecutable(t, filepath.Join(fixture.binDir, "vera"), "#!/usr/bin/env bash\nprintf invoked >\"$VERA_MARKER\"\nexit 99\n")
	t.Setenv("VERA_MARKER", marker)
	if _, _, err := fixture.run(t, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("vera invocation marker error=%v", err)
	}
}

func TestEmitter_RunIDIsUniqueAndSelfConsistent(t *testing.T) {
	fixture := newEmitterFixture(t)
	first, _, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.RunID == second.RunID || !ulidPattern.MatchString(first.RunID) || !ulidPattern.MatchString(second.RunID) {
		t.Fatalf("run ids %q and %q", first.RunID, second.RunID)
	}
	for _, runID := range []string{first.RunID, second.RunID} {
		if _, err := os.Stat(filepath.Join(fixture.root, ".vera", "spool", runID+".json")); err != nil {
			t.Fatalf("run id %s filename: %v", runID, err)
		}
	}
}

func TestEmitter_RecordsRepositoryAndTools(t *testing.T) {
	fixture := newEmitterFixture(t)
	witness, _, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	if witness.GitSHA != strings.Repeat("a", 40) || !witness.GitDirty {
		t.Fatalf("repository observation=%+v", witness)
	}
	if witness.ToolVersions.Go != "go version fixture" || witness.ToolVersions.GolangCILint != "golangci-lint fixture" || witness.ToolVersions.Make != "GNU Make fixture" {
		t.Fatalf("tool versions=%+v", witness.ToolVersions)
	}
}

func TestEmitter_BindsGateAndGitToRepository(t *testing.T) {
	fixture := newEmitterFixture(t)
	for _, variable := range []string{
		"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE", "GIT_INDEX_FILE",
		"GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_COUNT",
	} {
		t.Setenv(variable, t.TempDir())
	}
	witness, output, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	if witness.GitSHA != strings.Repeat("a", 40) || string(output) != "gate stdout\ngate stderr\n" {
		t.Fatalf("witness=%+v output=%q", witness, output)
	}
}

func TestEmitter_RepositoryObservationFailsBeforeGate(t *testing.T) {
	for name, environment := range map[string][]string{
		"head failure":   {"FAKE_GIT_HEAD_EXIT=2"},
		"malformed head": {"FAKE_GIT_HEAD_VALUE=not-a-hash"},
		"status failure": {"FAKE_GIT_STATUS_EXIT=2"},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newEmitterFixture(t)
			marker := filepath.Join(fixture.root, "make-was-invoked")
			before := witnessFiles(t, fixture.root)
			cmd := fixture.command(0)
			cmd.Env = append(cmd.Env, append(environment, "MAKE_MARKER="+marker)...)
			output, err := cmd.CombinedOutput()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || !strings.Contains(string(output), "cannot observe repository") {
				t.Fatalf("output=%q error=%v", output, err)
			}
			if after := witnessFiles(t, fixture.root); len(after) != len(before) {
				t.Fatalf("before=%v after=%v", before, after)
			}
			if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("make invocation marker error=%v", err)
			}
		})
	}
}

func TestEmitter_FreshCheckoutMakeTarget(t *testing.T) {
	fixture := newEmitterFixture(t)
	makefile, err := os.ReadFile(findRepositoryMakefile(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.root, "Makefile"), makefile, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fixture.script, 0o644); err != nil {
		t.Fatal(err)
	}
	actualMake, err := exec.LookPath("make")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(actualMake, "-C", fixture.root, "check-witnessed")
	cmd.Env = append(os.Environ(), "PATH="+fixture.binDir+string(os.PathListSeparator)+os.Getenv("PATH"), "EXPECTED_REPO_ROOT="+fixture.root, "TMPDIR="+fixture.root)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("output=%q error=%v", output, err)
	}
	files := witnessFiles(t, fixture.root)
	if len(files) != 1 {
		t.Fatalf("witness files=%v", files)
	}
	if _, err := readWitness(filepath.Join(fixture.root, ".vera", "spool", files[0])); err != nil {
		t.Fatal(err)
	}
}

func findRepositoryMakefile(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(dir, "Makefile")
		if data, readErr := os.ReadFile(candidate); readErr == nil && strings.Contains(string(data), "check-witnessed:") {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository Makefile not found")
		}
		dir = parent
	}
}

func TestEmitter_EscapesToolVersionControlCharacters(t *testing.T) {
	fixture := newEmitterFixture(t)
	t.Setenv("FAKE_GO_CONTROL", "1")
	witness, _, err := fixture.run(t, 0)
	if err != nil {
		t.Fatal(err)
	}
	if witness.ToolVersions.Go != "go\b\ffixture\t\r\x01\x1f" {
		t.Fatalf("go version=%q", witness.ToolVersions.Go)
	}
}

func newEmitterFixture(t *testing.T) emitterFixture {
	t.Helper()
	root := t.TempDir()
	scriptDir := filepath.Join(root, "kernel", "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "check-witness.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(scriptDir, "check-witness.sh")
	if err := os.WriteFile(script, source, 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "make"), "#!/usr/bin/env bash\nif [[ ${1:-} == --version ]]; then printf 'GNU Make fixture\\n'; exit 0; fi\nif [[ -n ${EXPECTED_REPO_ROOT:-} && $PWD != $EXPECTED_REPO_ROOT ]]; then printf 'wrong gate directory\\n' >&2; exit 42; fi\nif [[ -n ${MAKE_MARKER:-} ]]; then printf invoked >\"$MAKE_MARKER\"; fi\nprintf 'gate stdout\\n'\nprintf 'gate stderr\\n' >&2\nexit \"${FAKE_MAKE_EXIT:-0}\"\n")
	writeExecutable(t, filepath.Join(binDir, "git"), "#!/usr/bin/env bash\nfor variable in $(compgen -e); do if [[ $variable == GIT_* ]]; then printf '"+strings.Repeat("b", 40)+"\\n'; exit 0; fi; done\nif [[ ${3:-} == rev-parse ]]; then if [[ ${FAKE_GIT_HEAD_EXIT:-0} != 0 ]]; then exit \"$FAKE_GIT_HEAD_EXIT\"; fi; printf '%s\\n' \"${FAKE_GIT_HEAD_VALUE:-"+strings.Repeat("a", 40)+"}\"; exit 0; fi\nif [[ ${3:-} == status ]]; then if [[ ${FAKE_GIT_STATUS_EXIT:-0} != 0 ]]; then exit \"$FAKE_GIT_STATUS_EXIT\"; fi; printf ' M fixture\\n'; exit 0; fi\nexit 2\n")
	writeExecutable(t, filepath.Join(binDir, "go"), "#!/usr/bin/env bash\nif [[ -n ${FAKE_GO_CONTROL:-} ]]; then printf 'go\\b\\ffixture\\t\\r\\001\\037\\n'; else printf 'go version fixture\\n'; fi\n")
	writeExecutable(t, filepath.Join(binDir, "golangci-lint"), "#!/usr/bin/env bash\nprintf 'golangci-lint fixture\\n'\n")
	caller := filepath.Join(root, "caller")
	if err := os.Mkdir(caller, 0o755); err != nil {
		t.Fatal(err)
	}
	return emitterFixture{root: root, script: script, binDir: binDir, caller: caller}
}

func (f emitterFixture) run(t *testing.T, exitCode int) (Witness, []byte, error) {
	t.Helper()
	before := witnessFiles(t, f.root)
	cmd := f.command(exitCode)
	output, runErr := cmd.CombinedOutput()
	after := witnessFiles(t, f.root)
	if len(after) != len(before)+1 {
		t.Fatalf("before=%v after=%v", before, after)
	}
	var created string
	for _, name := range after {
		if !contains(before, name) {
			created = name
			break
		}
	}
	if created == "" {
		t.Fatalf("no new witness: before=%v after=%v", before, after)
	}
	witness, readErr := readWitness(filepath.Join(f.root, ".vera", "spool", created))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if created != witness.RunID+".json" {
		t.Fatalf("filename=%s run_id=%s", created, witness.RunID)
	}
	return witness, output, runErr
}

func (f emitterFixture) command(exitCode int) *exec.Cmd {
	cmd := exec.Command("bash", f.script)
	cmd.Dir = f.caller
	cmd.Env = append(os.Environ(),
		"PATH="+f.binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"FAKE_MAKE_EXIT="+strconv.Itoa(exitCode),
		"EXPECTED_REPO_ROOT="+f.root,
		"TMPDIR="+f.root,
	)
	return cmd
}

func witnessFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, ".vera", "spool"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, entry.Name())
		}
	}
	return names
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}
