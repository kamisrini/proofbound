package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// The write-time perimeter (docs/plans/write-perimeter-spec.md, CLI-INV-69): a Claude Code PreToolUse
// hook that refuses an agent's Edit, Write, MultiEdit or NotebookEdit of a file outside the area the
// active task's spec declares. It is guidance at the moment of writing, not a trust boundary: the
// agent controls its own environment, and C10 still judges every committed line.

// hookStdin is where `hook pre-tool-use` reads Claude Code's JSON (a variable for tests).
var hookStdin io.Reader = os.Stdin

type toolCall struct {
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
	} `json:"tool_input"`
}

// perimeterDecision says whether the call may proceed; when it may not, why, in words for the agent.
// Everything it cannot judge (no repository, no task, no area declared, a path outside the
// repository, a tool that writes no file) is allowed: C10 judges the commit later.
func perimeterDecision(ctx context.Context, raw []byte) (bool, string) {
	var call toolCall
	if err := json.Unmarshal(raw, &call); err != nil {
		return true, ""
	}
	target := call.ToolInput.FilePath
	if call.ToolName == "NotebookEdit" {
		target = call.ToolInput.NotebookPath
	}
	switch call.ToolName {
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
	default:
		return true, ""
	}
	if target == "" || call.Cwd == "" {
		return true, ""
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(call.Cwd, target)
	}
	// The repository is the one holding the file, not the session's cwd (review F2): a session started
	// above the repository, or after a `cd`, must still be judged, and a nested repository by its own.
	target = resolvePath(target)
	root := strings.TrimSpace(gitOutput(ctx, existingAncestor(target), "rev-parse", "--show-toplevel"))
	if root == "" {
		return true, ""
	}
	rel, err := filepath.Rel(resolvePath(root), target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true, ""
	}
	rel = filepath.ToSlash(rel)
	if housekeepingPath(rel) {
		return true, ""
	}
	claim, err := currentTask(root)
	if err != nil {
		return true, ""
	}
	name := strings.TrimPrefix(claim, "sdd:")
	var files, patterns []string
	for _, dir := range specDirsAt(ctx, root, "HEAD", name) {
		file := dir + "/" + areaFileName
		data, err := gitBytes(ctx, root, "show", "HEAD:"+file)
		if err != nil {
			continue
		}
		ps, err := parseArea(data)
		if err != nil {
			// An invalid area allows nothing, as in C10 (fail closed).
			return false, fmt.Sprintf("proofbound: %s is invalid (%v), so no code edit for %s can be checked against its area. Ask the spec owner to fix it, or stop.", file, err, claim)
		}
		files = append(files, file)
		patterns = append(patterns, ps...)
	}
	if len(files) == 0 || inArea(patterns, rel) {
		return true, ""
	}
	return false, fmt.Sprintf("proofbound: %s is outside the area the spec for %s allows (%s: %s). Do not edit it for this task. If the change truly needs it, stop and ask the spec owner to widen %s and commit that first; otherwise the reviewer must accept or reject every line outside the area (control C10).",
		rel, claim, strings.Join(files, ", "), strings.Join(patterns, " "), areaFileName)
}

// resolvePath follows symlinks through the nearest existing ancestor and appends the rest, so a link
// cannot place a file, even in a directory not created yet, outside the comparison (review F6).
func resolvePath(p string) string {
	p = filepath.Clean(p)
	rest := ""
	for dir := p; ; dir = filepath.Dir(dir) {
		if r, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(r, rest)
		}
		if parent := filepath.Dir(dir); parent == dir {
			return p
		}
		rest = filepath.Join(filepath.Base(dir), rest)
	}
}

// existingAncestor is p, or its nearest ancestor that exists as a directory.
func existingAncestor(p string) string {
	for dir := p; ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		if filepath.Dir(dir) == dir {
			return dir
		}
	}
}

// runPerimeterHook is `proofbound hook pre-tool-use`. It always exits 0 and refuses with Claude
// Code's PreToolUse JSON on stdout (permissionDecision "deny"), never with exit 2: exit 2 is also what
// an older proofbound returns for an unknown command and what a crash returns, and either must not
// block every edit (review F1). The installed command adds `|| true` for the same reason.
func runPerimeterHook(ctx context.Context, stdout io.Writer) int {
	raw, err := io.ReadAll(io.LimitReader(hookStdin, 1<<20))
	if err != nil {
		return 0
	}
	ok, why := perimeterDecision(ctx, raw)
	if ok {
		return 0
	}
	out, err := json.Marshal(map[string]any{"hookSpecificOutput": map[string]string{"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": why}})
	if err == nil {
		_, _ = fmt.Fprintln(stdout, string(out))
	}
	return 0
}

// claudeSettingsPath holds the agent hook: housekeeping (no claim needed) and listed by C6.
const claudeSettingsPath = ".claude/settings.json"

// claudeSettings is the .claude/settings.json `init` writes when there is none: the PreToolUse hook.
const claudeSettings = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit|NotebookEdit",
        "hooks": [ { "type": "command", "command": "proofbound hook pre-tool-use || true" } ]
      }
    ]
  }
}
`

// writeClaudeSettings writes .claude/settings.json, or, when one exists without the hook, says how to
// add it (init never edits a file it did not create).
func writeClaudeSettings(output io.Writer, root string) error {
	path := filepath.Join(root, ".claude", "settings.json")
	existing, err := os.ReadFile(path)
	if err == nil {
		if strings.Contains(string(existing), "proofbound hook pre-tool-use") {
			_, err = fmt.Fprintf(output, "exists %s\n", path)
			return err
		}
		_, err = fmt.Fprintf(output, "exists %s without the Proofbound hook; add a PreToolUse hook running `proofbound hook pre-tool-use || true` for Edit|Write|MultiEdit|NotebookEdit\n", path)
		return err
	}
	return writeNew(output, root, ".claude/settings.json", claudeSettings, 0o644)
}
