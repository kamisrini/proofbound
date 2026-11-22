package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
)

// Agent-native intent binding (CLI-INV-34..36, CI-mvp-agent-intent-…): an agent names the change it is
// working on once, with `proofbound task start <change>`, and the git hooks `init` installs stamp the
// claim on every commit that changes code. Nothing here is a trust boundary — an agent controls its own
// hooks — so the pull-request check (CLI-INV-33) stays the place where an unclaimed code commit fails.

const taskFile = ".proofbound/task"

var (
	taskNameRE   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
	trailerRE    = regexp.MustCompile(`^([A-Za-z0-9-]+): (.+)$`) // the kernel's trailer grammar (gitcmd)
	errNoTask    = errors.New("no task")
	errHookBlock = errors.New("commit refused")
)

func parseTask(args []string) (sub, name string, ok bool) {
	switch {
	case len(args) == 1 && args[0] == "task":
		return "show", "", true
	case len(args) == 2 && args[0] == "task" && args[1] == "stop":
		return "stop", "", true
	case len(args) == 3 && args[0] == "task" && args[1] == "start" && args[2] != "":
		return "start", args[2], true
	}
	return "", "", false
}

// parseHook accepts `hook commit-msg <file>` and `hook prepare-commit-msg <file> [<source> [<sha>]]`,
// the arguments git passes to those hooks.
func parseHook(args []string) (hook, file, source string, ok bool) {
	if len(args) < 3 || args[0] != "hook" || args[2] == "" {
		return "", "", "", false
	}
	switch {
	case args[1] == "commit-msg" && len(args) == 3:
		return args[1], args[2], "", true
	case args[1] == "prepare-commit-msg" && len(args) <= 5:
		if len(args) >= 4 {
			source = args[3]
		}
		return args[1], args[2], source, true
	}
	return "", "", "", false
}

func runTask(root, sub, name string, output io.Writer) error {
	path := filepath.Join(root, filepath.FromSlash(taskFile))
	switch sub {
	case "start":
		if !taskNameRE.MatchString(name) {
			return fmt.Errorf("task: %q is not a change name (the folder name under openspec/changes/ or specs/)", name)
		}
		if !changeExists(root, name) {
			return fmt.Errorf("task: no change %q: write the spec first (openspec/changes/%s/ or specs/%s/), then start the task", name, name, name)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte("sdd:"+name+"\n"), 0o644); err != nil {
			return err
		}
		_, err := fmt.Fprintf(output, "task sdd:%s: commits that change code now end with `Intent: sdd:%s`; `proofbound task stop` when done\n", name, name)
		return err
	case "stop":
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		_, err := fmt.Fprintln(output, "no task")
		return err
	default:
		claim, err := currentTask(root)
		if errors.Is(err, errNoTask) {
			_, err = fmt.Fprintln(output, "no task")
			return err
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "task %s\n", claim)
		return err
	}
}

func changeExists(root, name string) bool {
	for _, dir := range []string{"openspec/changes/" + name, "openspec/changes/archive/" + name, "specs/" + name} {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// currentTask is the active claim (`sdd:<name>`), or errNoTask. A task file that does not hold exactly
// one well-formed claim is an error, never silently ignored.
func currentTask(root string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(taskFile)))
	if errors.Is(err, os.ErrNotExist) {
		return "", errNoTask
	}
	if err != nil {
		return "", err
	}
	claim := strings.TrimSpace(string(raw))
	name, ok := strings.CutPrefix(claim, "sdd:")
	if !ok || !taskNameRE.MatchString(name) {
		return "", fmt.Errorf("task: %s holds %q, not sdd:<change>; run `proofbound task start <change>` or `proofbound task stop`", taskFile, claim)
	}
	return claim, nil
}

// stagedCode is every staged path (index against HEAD, or against the empty tree before the first
// commit) that housekeepingPath does not exempt.
func stagedCode(ctx context.Context, root string) ([]string, error) {
	args := []string{"-C", root, "diff", "--cached", "--name-only", "--no-renames", "-z"}
	if gitOutput(ctx, root, "rev-parse", "--verify", "-q", "HEAD") == "" {
		args = append(args, "4b825dc642cb6eb9a060e54bf8d69288fbee4904") // the empty tree
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	// git runs hooks with GIT_INDEX_FILE naming the index it is about to commit (a temporary one for
	// `commit -a` and `commit <path>`); dropping it, as repositoryGitEnv does, reads the wrong index.
	cmd.Env = repositoryGitEnv()
	if index := os.Getenv("GIT_INDEX_FILE"); index != "" {
		cmd.Env = append(cmd.Env, "GIT_INDEX_FILE="+index)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hook: cannot list staged paths: %s", strings.TrimSpace(stderr.String()))
	}
	var code []string
	for _, p := range strings.Split(string(raw), "\x00") {
		if p != "" && !housekeepingPath(p) {
			code = append(code, p)
		}
	}
	return code, nil
}

func inMerge(ctx context.Context, root string) bool {
	path := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--git-path", "MERGE_HEAD"))
	if path == "" {
		return false
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	_, err := os.Stat(path)
	return err == nil
}

// runHook is the body of both hooks `init` installs (CLI-INV-35, CLI-INV-36).
//   - prepare-commit-msg: with an active task, a message that claims nothing, and staged code, it adds
//     `Intent: sdd:<name>` to the message's final trailer block. Housekeeping-only commits (reviews,
//     keys, specs, docs) are never stamped, so a reviewer working beside an active task does not claim it.
//   - commit-msg: refuses a malformed claim, and a commit that stages code but claims no spec. Merges
//     pass (the pull-request check judges a merge by its own combined diff).
func runHook(ctx context.Context, root, hook, file, source string, output io.Writer) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	msg := string(raw)
	names, malformed := claimLines(msg)
	if hook == "prepare-commit-msg" {
		if len(names)+len(malformed) > 0 || source == "merge" || source == "squash" || inMerge(ctx, root) {
			return nil
		}
		claim, err := currentTask(root)
		if errors.Is(err, errNoTask) {
			return nil
		}
		if err != nil {
			return err
		}
		code, err := stagedCode(ctx, root)
		if err != nil || len(code) == 0 {
			return err
		}
		return os.WriteFile(file, []byte(withTrailer(msg, "Intent: "+claim)), 0o644)
	}
	if len(malformed) > 0 {
		fmt.Fprintf(output, "proofbound: malformed claim %s; write exactly `Intent: sdd:<change>` as the message's last line\n", strings.Join(malformed, ", "))
		return errHookBlock
	}
	if inMerge(ctx, root) {
		return nil
	}
	if len(names) > 0 {
		// The kernel binds only claims in the message's final trailer block; refuse one it would skip,
		// so the commit does not read as claimed here and UNBOUND in the pull-request check.
		if bound := gitcmd.IntentTrailers(uncommented(msg)); len(bound) == 0 {
			fmt.Fprintf(output, "proofbound: the claim is not in the message's final trailer block; put `Intent: sdd:<change>` on the last line, after a blank line\n")
			return errHookBlock
		}
		return nil
	}
	code, err := stagedCode(ctx, root)
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return nil
	}
	shown := code
	if len(shown) > 3 {
		shown = shown[:3]
	}
	fmt.Fprintf(output, "proofbound: this commit changes code (%s) but claims no spec.\n  Run `proofbound task start <change>` and commit again: the claim is added for you.\n  Specs, docs, gates, review records and Markdown need no claim.\n", strings.Join(shown, ", "))
	return errHookBlock
}

// withTrailer adds line to msg's final trailer block in the kernel's grammar (a last paragraph of
// `Key: value` lines after a blank line), or starts one. Trailing comment lines (`#`, as git's editor
// template writes them) stay after it, so git's cleanup leaves the trailer last.
func withTrailer(msg, line string) string {
	lines := strings.Split(strings.TrimRight(msg, "\n"), "\n")
	end := len(lines)
	for end > 0 && (strings.HasPrefix(lines[end-1], "#") || strings.TrimSpace(lines[end-1]) == "") {
		end--
	}
	body, tail := lines[:end], lines[end:]
	start := len(body)
	for start > 0 && trailerRE.MatchString(body[start-1]) {
		start--
	}
	var out []string
	if len(body) == 0 {
		// The editor flow: leave line 1 for the subject and a blank line before the trailer.
		out = []string{"", "", line}
	} else if start < len(body) && start > 0 && body[start-1] == "" && start > 1 {
		out = append(append(out, body...), line)
	} else {
		out = append(append(out, body...), "", line)
	}
	if len(body) == 0 && len(tail) > 0 {
		return strings.Join(append(out, trimLeadingBlank(tail)...), "\n") + "\n"
	}
	if len(tail) > 0 {
		out = append(append(out, ""), trimLeadingBlank(tail)...)
	}
	return strings.Join(out, "\n") + "\n"
}

// uncommented drops `#` comment lines, as git's default cleanup does after the commit-msg hook runs.
func uncommented(msg string) string {
	var kept []string
	for _, line := range strings.Split(msg, "\n") {
		if !strings.HasPrefix(line, "#") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func trimLeadingBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	return lines
}
