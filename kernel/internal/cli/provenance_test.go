package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// postCall is the JSON Claude Code sends a PostToolUse hook.
func postCall(t *testing.T, session, cwd, tool string, input, response map[string]any) []byte {
	t.Helper()
	m := map[string]any{"session_id": session, "cwd": cwd, "hook_event_name": "PostToolUse", "tool_name": tool, "tool_input": input}
	if response != nil {
		m["tool_response"] = response
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func logRecords(t *testing.T, root, session string) []sessionEdit {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(sessionsDir(context.Background(), root), session+".jsonl"))
	if err != nil {
		return nil
	}
	var out []sessionEdit
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		var r sessionEdit
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatalf("a log line is not JSON: %s", line)
		}
		if len(line)+1 > provRecordMax {
			t.Fatalf("a record of %d bytes exceeds %d", len(line)+1, provRecordMax)
		}
		out = append(out, r)
	}
	return out
}

func hashes(lines ...string) []string {
	var out []string
	for _, l := range lines {
		out = append(out, provHash(l))
	}
	return out
}

// CLI-INV-70: the PostToolUse hook logs only the lines each edit added, in the git common folder.
func TestSessionLog(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	file := filepath.Join(root, "src", "a.sh")
	rec := func(raw []byte) {
		t.Helper()
		if err := recordSessionEdit(ctx, raw); err != nil {
			t.Fatal(err)
		}
	}
	// The structured patch decides: context lines are not credited.
	rec(postCall(t, "s1", root, "Edit", map[string]any{"file_path": file, "old_string": "echo one\necho two", "new_string": "echo one\necho TWO"},
		map[string]any{"structuredPatch": []any{map[string]any{"lines": []string{" echo one", "-echo two", "+echo PATCHED"}}}}))
	// Without a patch: new minus old as a multiset.
	rec(postCall(t, "s1", root, "Edit", map[string]any{"file_path": "src/a.sh", "old_string": "keep\nold", "new_string": "keep\nnew line\nkeep"}, nil))
	rec(postCall(t, "s1", root, "MultiEdit", map[string]any{"file_path": file, "edits": []any{map[string]any{"old_string": "x", "new_string": "x\nmulti"}}}, nil))
	// A new file: every line, not a whole-file write; an existing file without a patch: whole-file.
	rec(postCall(t, "s1", root, "Write", map[string]any{"file_path": file, "content": "w1\n\nw2\n"}, map[string]any{"type": "create", "structuredPatch": []any{}}))
	rec(postCall(t, "s1", root, "Write", map[string]any{"file_path": file, "content": "u1\n"}, map[string]any{"type": "update"}))
	got := logRecords(t, root, "s1")
	want := [][]string{hashes("echo PATCHED"), hashes("new line", "keep"), hashes("multi"), hashes("w1", "w2"), hashes("u1")}
	if len(got) != len(want) {
		t.Fatalf("%d records, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if strings.Join(got[i].Lines, ",") != strings.Join(want[i], ",") || got[i].Path != "src/a.sh" || got[i].Session != "s1" || got[i].At == "" {
			t.Errorf("record %d: %+v, want lines %v", i, got[i], want[i])
		}
	}
	if got[3].Whole || !got[4].Whole {
		t.Errorf("whole-file marks: create %v, update %v", got[3].Whole, got[4].Whole)
	}

	// Nothing is logged for what is not an agent's successful edit of a traced file in a repository.
	for name, raw := range map[string][]byte{
		"a failed edit (defensive)": postCall(t, "s2", root, "Write", map[string]any{"file_path": file, "content": "z\n"}, map[string]any{"success": false}),
		"an error (defensive)":      postCall(t, "s2", root, "Write", map[string]any{"file_path": file, "content": "z\n"}, map[string]any{"error": "boom"}),
		"not an edit":               postCall(t, "s2", root, "Bash", map[string]any{"command": "echo z > src/a.sh"}, nil),
		"a notebook":                postCall(t, "s2", root, "NotebookEdit", map[string]any{"notebook_path": file, "new_source": "z"}, nil),
		"outside the repository":    postCall(t, "s2", root, "Write", map[string]any{"file_path": filepath.Join(t.TempDir(), "x"), "content": "z\n"}, nil),
		"housekeeping":              postCall(t, "s2", root, "Write", map[string]any{"file_path": filepath.Join(root, "docs", "n.md"), "content": "z\n"}, nil),
		"a bad session id":          postCall(t, "../s2", root, "Write", map[string]any{"file_path": file, "content": "z\n"}, nil),
		"malformed":                 []byte("{not json"),
	} {
		if err := recordSessionEdit(ctx, raw); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if got := logRecords(t, root, "s2"); len(got) != 0 {
			t.Errorf("%s was logged", name)
		}
	}
	if _, err := os.Stat(filepath.Join(sessionsDir(ctx, root), "..", "s2.jsonl")); err == nil {
		t.Error("a session id escaped the log folder")
	}

	// The hook never fails the edit, whatever it reads.
	old := hookStdin
	t.Cleanup(func() { hookStdin = old })
	hookStdin = strings.NewReader("garbage")
	if code := run(ctx, []string{"hook", "post-tool-use"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("exit %d", code)
	}

	// A large write is split into records of at most provRecordMax bytes (logRecords checks), none lost.
	var big []string
	for i := 0; i < 500; i++ {
		big = append(big, "line number "+strings.Repeat("x", i%7)+string(rune('a'+i%26))+strings.Repeat("y", i/26))
	}
	rec(postCall(t, "s3", root, "Write", map[string]any{"file_path": file, "content": strings.Join(big, "\n")}, map[string]any{"type": "create"}))
	n := 0
	recs := logRecords(t, root, "s3")
	for _, r := range recs {
		n += len(r.Lines)
	}
	if n != 500 || len(recs) < 2 {
		t.Fatalf("%d hashes in %d records", n, len(recs))
	}

	// A worktree's log lands in the common git folder, where the main checkout reads it.
	uiWrite(t, root, "src/a.sh", "base\n")
	uiCommit(t, root, "owner", "base")
	wt := filepath.Join(t.TempDir(), "wt")
	uiGit(t, root, "worktree", "add", "-q", wt)
	rec(postCall(t, "s4", wt, "Write", map[string]any{"file_path": filepath.Join(wt, "src", "b.sh"), "content": "from the worktree\n"}, map[string]any{"type": "create"}))
	if got := logRecords(t, root, "s4"); len(got) != 1 || got[0].Path != "src/b.sh" {
		t.Fatalf("worktree log: %+v", got)
	}
}

// provRepo: a person's commit, then an agent's commit whose lines come partly from logged edits.
func provRepo(t *testing.T) (root, claim string) {
	t.Helper()
	ctx := context.Background()
	root = taskRepo(t)
	uiWrite(t, root, "src/a.sh", "echo human one\n")
	uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
	uiCommit(t, root, "owner", "start")
	file := filepath.Join(root, "src", "a.sh")
	for _, raw := range [][]byte{
		postCall(t, "sess-aaaaaaaa1", root, "Edit", map[string]any{"file_path": file, "old_string": "echo human one", "new_string": "echo human one\necho agent two\n}\necho dup line\necho dup line"}, nil),
		postCall(t, "sess-bbbbbbbb2", root, "Edit", map[string]any{"file_path": file, "old_string": "x", "new_string": "echo agent two\necho crlf line\r"}, nil),
		// The same text in another file does not count for src/a.sh.
		postCall(t, "sess-aaaaaaaa1", root, "Write", map[string]any{"file_path": filepath.Join(root, "src", "other.sh"), "content": "echo by shell\n"}, map[string]any{"type": "create"}),
		postCall(t, "sess-aaaaaaaa1", root, "Write", map[string]any{"file_path": filepath.Join(root, "src", "w.sh"), "content": "echo typed first\necho agent rewrote\n"}, map[string]any{"type": "update"}),
	} {
		if err := recordSessionEdit(ctx, raw); err != nil {
			t.Fatal(err)
		}
	}
	// What reaches the commit: the logged lines, a shell-written line, a CRLF line and no final newline.
	uiWrite(t, root, "src/a.sh", "echo human one\necho agent two\n}\necho dup line\necho dup line\necho by shell\necho crlf line\r\n")
	uiWrite(t, root, "src/w.sh", "echo typed first\necho agent rewrote")
	uiWrite(t, root, "docs/n.md", "echo agent two\n")
	claim = uiCommit(t, root, "Claude", "change a\n\nIntent: sdd:001-a")
	return root, claim
}

// CLI-INV-71: each added line is by the sessions that logged it, has no recorded agent edit, or is
// too common to count; with the three warnings.
func TestProvenanceMatch(t *testing.T) {
	ctx := context.Background()
	root, claim := provRepo(t)
	res, err := commitProvenance(ctx, root, claim)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]provFile{}
	for _, f := range res.Files {
		files[f.Path] = f
	}
	if _, ok := files["docs/n.md"]; ok || len(files) != 2 {
		t.Fatalf("traced files: %+v", res.Files)
	}
	a := files["src/a.sh"]
	// echo agent two: both sessions; echo crlf line: session b (CRLF ignored); echo by shell: line 6
	// unrecorded (another file's log does not count); "}" and the repeated line: common.
	if len(a.Sessions) != 2 || a.Sessions[0].Session != "sess-aaaaaaaa1" || a.Sessions[0].Lines != 1 || a.Sessions[1].Lines != 2 || a.Sessions[0].Whole {
		t.Errorf("src/a.sh sessions: %+v", a.Sessions)
	}
	if len(a.Unrecorded) != 1 || a.Unrecorded[0] != 6 || a.Common != 3 {
		t.Errorf("src/a.sh: unrecorded %v common %d", a.Unrecorded, a.Common)
	}
	// A person's line followed by an agent's whole-file write matches, but only "via a whole-file write".
	w := files["src/w.sh"]
	if len(w.Sessions) != 1 || w.Sessions[0].Lines != 2 || !w.Sessions[0].Whole || len(w.Unrecorded) != 0 {
		t.Errorf("src/w.sh: %+v", w)
	}
	if len(res.Logs) != 2 || len(res.Logs[0].SHA256) != 64 {
		t.Errorf("logs: %+v", res.Logs)
	}
	if strings.Join(res.Warnings, "|") != provWarnNoSettings {
		t.Errorf("warnings: %v", res.Warnings)
	}
	var out bytes.Buffer
	if err := writeProvLines(&out, res); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"provenance src/a.sh: 1 by Claude Code session sess-aaa, 2 by Claude Code session sess-bbb, 1 with no recorded agent edit (lines 6), 3 common\n",
		"provenance src/w.sh: 2 by Claude Code session sess-aaa via a whole-file write, 0 with no recorded agent edit, 0 common\n",
		"provenance warning: " + provWarnNoSettings + "\n",
		"provenance: docs, Markdown and records are not traced",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("draft lines lack %q:\n%s", want, out.String())
		}
	}

	// No time cut-off: an edit logged after the commit (an amend) still matches.
	if err := recordSessionEdit(ctx, postCall(t, "sess-ccccccccc3", root, "Write", map[string]any{"file_path": filepath.Join(root, "src", "a.sh"), "content": "echo by shell\n"}, map[string]any{"type": "create"})); err != nil {
		t.Fatal(err)
	}
	uiWrite(t, root, ".claude/settings.json", claudeSettings)
	if res, _ = commitProvenance(ctx, root, claim); len(res.Files[0].Unrecorded) != 0 || len(res.Warnings) != 0 {
		t.Errorf("after a later edit: %+v %v", res.Files[0], res.Warnings)
	}

	// No logs at all; an agent commit with logs that match nothing.
	t.Run("warnings", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, ".claude/settings.json", claudeSettings)
		uiWrite(t, root, "src/a.sh", "echo something long\n")
		c := uiCommit(t, root, "Claude", "x")
		res, err := commitProvenance(ctx, root, c)
		if err != nil || strings.Join(res.Warnings, "|") != provWarnNoLogs {
			t.Fatalf("%v %v", err, res.Warnings)
		}
		if err := recordSessionEdit(ctx, postCall(t, "s9", root, "Write", map[string]any{"file_path": filepath.Join(root, "src", "z.sh"), "content": "unrelated text\n"}, map[string]any{"type": "create"})); err != nil {
			t.Fatal(err)
		}
		if res, _ := commitProvenance(ctx, root, c); strings.Join(res.Warnings, "|") != provWarnNoHook {
			t.Fatalf("%v", res.Warnings)
		}
		uiWrite(t, root, "src/a.sh", "echo a person's line\n")
		person := uiCommit(t, root, "owner", "y")
		if res, _ := commitProvenance(ctx, root, person); len(res.Warnings) != 0 {
			t.Fatalf("a person's commit warned: %v", res.Warnings)
		}
	})
}

// CLI-INV-71 (seal): seal matches again, ignores the draft's provenance, and writes a signed PROV-.
func TestSealWritesProv(t *testing.T) {
	ctx := context.Background()
	keyPath := filepath.Join(t.TempDir(), "reviewer-key")
	t.Setenv("PROOFBOUND_KEY", keyPath)
	pub, err := pbwarranty.InitKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	root, claim := provRepo(t)
	uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(pub))
	uiCommit(t, root, "owner", "declare the reviewer")
	d := sealDraft("SATISFIED", "VERIFIABLE")
	d.Commit = claim
	d.Provenance = &provResult{Files: []provFile{{Path: "src/a.sh", Sessions: []provSession{{Session: "forged", Lines: 99}}}}}
	var out bytes.Buffer
	if err := reviewSeal(ctx, root, writeDraft(t, root, d), areaTruth(claim), &out); err != nil || !strings.Contains(out.String(), "wrote "+provRecordPath(claim)) {
		t.Fatalf("%v\n%s", err, out.String())
	}
	uiCommit(t, root, "proofbound-reviewer", "review")
	fields, res, err := readProvRecord(ctx, root, claim)
	if err != nil || fields["by"] != "proofbound-reviewer" || fields["key"] == "" {
		t.Fatalf("%v %v", err, fields)
	}
	if strings.Contains(fields["files"], "forged") || len(res.Files) != 2 || len(res.Files[0].Unrecorded) != 1 || len(res.Logs) != 2 {
		t.Fatalf("the record: %+v", res)
	}
}

// CLI-INV-72: C12 discloses; it is an exception only for a record that does not verify or is not
// independent, never for the counts.
func TestAuditC12(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	control := func(root, claim string) (auditControl, []auditException) {
		t.Helper()
		rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "change a", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
		var ex []auditException
		c, err := provControl(ctx, root, &rep, func(kind, ctl, commit, text string) {
			ex = append(ex, auditException{Kind: kind, Control: ctl, Commit: commit, Text: text})
		})
		if err != nil {
			t.Fatal(err)
		}
		return c, ex
	}
	record := func(t *testing.T, root, claim, by string, k ed25519.PrivateKey, edit func([]byte) []byte) {
		t.Helper()
		uiWrite(t, root, ".proofbound/keys/"+by+".pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
		res, err := commitProvenance(ctx, root, claim)
		if err != nil {
			t.Fatal(err)
		}
		data, err := provRecordBytes(claim, by, res, k, "2026-11-24T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		if edit != nil {
			data = edit(data)
		}
		uiWrite(t, root, provRecordPath(claim), string(data))
		uiCommit(t, root, by, "provenance review")
	}
	t.Run("no record: not applicable, counted as not traced", func(t *testing.T) {
		root, claim := provRepo(t)
		if c, ex := control(root, claim); c.State != auditNone || len(ex) != 0 || !strings.Contains(c.Count, "1 not traced") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("a good record holds and discloses the unrecorded lines", func(t *testing.T) {
		root, claim := provRepo(t)
		record(t, root, claim, "rev", key, nil)
		c, ex := control(root, claim)
		if c.State != auditListed || len(ex) != 0 || len(c.Evidence) != 1 || !strings.Contains(c.Evidence[0].Text, "1 with no recorded agent edit in src/a.sh (1)") || !strings.Contains(c.Evidence[0].Text, "line(s) by session sess-aaa") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	for name, tc := range map[string]struct {
		by   string
		k    bool
		edit func([]byte) []byte
		want string
	}{
		"unsigned": {by: "rev", want: "does not verify"},
		"tampered": {by: "rev", k: true, edit: func(b []byte) []byte {
			return bytes.Replace(b, []byte(`\"unrecorded\":[6]`), []byte(`\"unrecorded\":[]`), 1)
		}, want: "does not verify"},
		"signed by the agent": {by: "claude", k: true, want: "not independent"},
	} {
		t.Run(name, func(t *testing.T) {
			root, claim := provRepo(t)
			var k ed25519.PrivateKey
			if tc.k {
				k = key
			}
			record(t, root, claim, tc.by, k, tc.edit)
			c, ex := control(root, claim)
			if c.State != auditExcepted || len(ex) != 1 || ex[0].Kind != "provenance" || ex[0].Control != "C12" || !strings.Contains(ex[0].Text, tc.want) {
				t.Fatalf("%+v %+v", c, ex)
			}
		})
	}
}

// CLI-INV-70 (init): both hooks are written; an older settings file is told exactly what to add, and
// the installed recording command never blocks an edit when proofbound is old or broken.
func TestInitHooks(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	if err := writeClaudeSettings(&out, root); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	var parsed struct {
		Hooks struct {
			PostToolUse []struct {
				Matcher string                           `json:"matcher"`
				Hooks   []struct{ Type, Command string } `json:"hooks"`
			}
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil || len(parsed.Hooks.PostToolUse) != 1 || parsed.Hooks.PostToolUse[0].Matcher != "Edit|Write|MultiEdit" || len(parsed.Hooks.PostToolUse[0].Hooks) != 1 || parsed.Hooks.PostToolUse[0].Hooks[0].Command != "proofbound hook post-tool-use || true" {
		t.Fatalf("%s", data)
	}
	stub := t.TempDir()
	uiWrite(t, stub, "proofbound", "#!/bin/sh\necho usage >&2\nexit 2\n")
	if err := os.Chmod(filepath.Join(stub, "proofbound"), 0o755); err != nil {
		t.Fatal(err)
	}
	sh := exec.Command("sh", "-c", parsed.Hooks.PostToolUse[0].Hooks[0].Command)
	sh.Env = append(os.Environ(), "PATH="+stub+":/usr/bin:/bin")
	if err := sh.Run(); err != nil {
		t.Fatalf("the installed command fails when proofbound fails: %v", err)
	}
	// The file v0.3.2's init wrote: only the perimeter hook.
	old := t.TempDir()
	v032 := strings.Replace(claudeSettings, "proofbound hook post-tool-use", "removed", 1)
	uiWrite(t, old, ".claude/settings.json", v032)
	out.Reset()
	if err := writeClaudeSettings(&out, old); err != nil || !strings.Contains(out.String(), `"PostToolUse": [ { "matcher": "Edit|Write|MultiEdit", "hooks": [ { "type": "command", "command": "proofbound hook post-tool-use || true" } ] } ]`) || strings.Contains(out.String(), `"PreToolUse"`) {
		t.Fatalf("%v %s", err, out.String())
	}
	if data, _ := os.ReadFile(filepath.Join(old, ".claude", "settings.json")); string(data) != v032 {
		t.Fatal("an existing settings.json was edited")
	}
	out.Reset()
	if err := writeClaudeSettings(&out, root); err != nil || !strings.HasPrefix(out.String(), "exists ") || strings.Contains(out.String(), "without") {
		t.Fatalf("%v %s", err, out.String())
	}
}

// The code review's findings (session-provenance code review F1..F6), each with a regression test.
func TestProvenanceReviewFindings(t *testing.T) {
	ctx := context.Background()
	write := func(t *testing.T, root, session, rel, content string) {
		t.Helper()
		if err := recordSessionEdit(ctx, postCall(t, session, root, "Write", map[string]any{"file_path": filepath.Join(root, filepath.FromSlash(rel)), "content": content}, map[string]any{"type": "create"})); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("F2 a space or non-ASCII in the path", func(t *testing.T) {
		root := taskRepo(t)
		uiCommit(t, root, "owner", "start")
		for _, rel := range []string{"src/a b.sh", "src/café.sh"} {
			write(t, root, "s1", rel, "echo long enough line\n")
			uiWrite(t, root, rel, "echo long enough line\n")
		}
		c := uiCommit(t, root, "Claude", "x")
		res, err := commitProvenance(ctx, root, c)
		if err != nil || len(res.Files) != 2 {
			t.Fatalf("%v %+v", err, res.Files)
		}
		for _, f := range res.Files {
			if (f.Path != "src/a b.sh" && f.Path != "src/café.sh") || len(f.Sessions) != 1 || len(f.Unrecorded) != 0 {
				t.Errorf("%q: %+v", f.Path, f)
			}
		}
	})
	t.Run("F3 an added line starting with ++ is not a file header", func(t *testing.T) {
		root := taskRepo(t)
		uiCommit(t, root, "owner", "start")
		body := "first real line\n++ looks like a header\nafter the plus line\n"
		write(t, root, "s1", "src/p.txt", body)
		uiWrite(t, root, "src/p.txt", body)
		c := uiCommit(t, root, "Claude", "x")
		res, err := commitProvenance(ctx, root, c)
		if err != nil || len(res.Files) != 1 || len(res.Files[0].Sessions) != 1 || res.Files[0].Sessions[0].Lines != 3 {
			t.Fatalf("%v %+v", err, res.Files)
		}
	})
	t.Run("F5 an unreadable log or log folder warns, never fails", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "src/a.sh", "echo something long\n")
		c := uiCommit(t, root, "Claude", "x")
		dir := sessionsDir(ctx, root)
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dir, []byte("not a folder"), 0o644); err != nil {
			t.Fatal(err)
		}
		if res, err := commitProvenance(ctx, root, c); err != nil || !strings.Contains(strings.Join(res.Warnings, "|"), "folder cannot be read") {
			t.Fatalf("%v %v", err, res.Warnings)
		}
		if err := os.Remove(dir); err != nil {
			t.Fatal(err)
		}
		write(t, root, "s1", "src/a.sh", "echo something long\n")
		if err := os.Chmod(filepath.Join(dir, "s1.jsonl"), 0); err != nil {
			t.Fatal(err)
		}
		if _, err := os.ReadFile(filepath.Join(dir, "s1.jsonl")); err == nil {
			t.Skip("running as a user who can read any file")
		}
		if res, err := commitProvenance(ctx, root, c); err != nil || !strings.Contains(strings.Join(res.Warnings, "|"), "s1.jsonl cannot be read") {
			t.Fatalf("%v %v", err, res.Warnings)
		}
	})
	t.Run("F6 a path too long for one record is skipped, not split or failed", func(t *testing.T) {
		root := taskRepo(t)
		var segs []string
		for i := 0; i < 16; i++ {
			segs = append(segs, strings.Repeat(string(rune('a'+i)), 200))
		}
		write(t, root, "s1", strings.Join(segs, "/")+"/x.sh", "echo something long\n")
		if got := logRecords(t, root, "s1"); len(got) != 0 {
			t.Fatalf("logged %d records", len(got))
		}
	})
	t.Run("F9 a link planted at the log path is not followed", func(t *testing.T) {
		root := taskRepo(t)
		dir := sessionsDir(ctx, root)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "victim")
		if err := os.WriteFile(target, []byte("keep\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(dir, "s1.jsonl")); err != nil {
			t.Fatal(err)
		}
		write(t, root, "s1", "src/a.sh", "echo something long\n")
		if data, _ := os.ReadFile(target); string(data) != "keep\n" {
			t.Fatalf("the hook wrote through a link: %q", data)
		}
	})
	t.Run("F1 a seal without a declared key writes no PROV record", func(t *testing.T) {
		keyPath := filepath.Join(t.TempDir(), "reviewer-key")
		t.Setenv("PROOFBOUND_KEY", keyPath)
		if _, err := pbwarranty.InitKey(keyPath); err != nil {
			t.Fatal(err)
		}
		root, claim := provRepo(t)
		d := sealDraft("SATISFIED", "VERIFIABLE")
		d.Commit = claim
		var out bytes.Buffer
		if err := reviewSeal(ctx, root, writeDraft(t, root, d), areaTruth(claim), &out); err != nil {
			t.Fatalf("%v\n%s", err, out.String())
		}
		if _, err := os.Lstat(filepath.Join(root, provRecordPath(claim))); err == nil {
			t.Fatal("an unsigned PROV record was written")
		}
	})
}

// CLI-INV-72: a record signed for another commit, placed at this commit's path, is an exception.
func TestAuditC12JudgesItsCommit(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	root, claim := provRepo(t)
	other := uiGit(t, root, "rev-parse", claim+"^1")
	uiWrite(t, root, ".proofbound/keys/rev.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	data, err := provRecordBytes(other, "rev", provResult{Files: []provFile{}, Warnings: []string{}, Logs: []provLog{}}, key, "2026-11-24T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	uiWrite(t, root, provRecordPath(claim), string(data))
	uiCommit(t, root, "rev", "provenance review")
	rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "change a", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
	var ex []auditException
	c, err := provControl(ctx, root, &rep, func(kind, ctl, commit, text string) { ex = append(ex, auditException{Text: text}) })
	if err != nil || c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, "judges another commit") {
		t.Fatalf("%v %+v %+v", err, c, ex)
	}
}
