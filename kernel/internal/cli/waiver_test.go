package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// waiverRepo declares this machine's key as "approver" and returns the repository, its base and now.
func waiverRepo(t *testing.T) (string, string, ed25519.PrivateKey) {
	t.Helper()
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/approver.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	uiWrite(t, root, "README.md", "x\n")
	base := uiCommit(t, root, "owner", "start")
	return root, base, key
}

func day(offset int) string { return time.Now().UTC().AddDate(0, 0, offset).Format("2006-01-02") }

func signWaiver(t *testing.T, root, id string, paths []string, expires string) {
	t.Helper()
	var out bytes.Buffer
	if err := waiverSign(context.Background(), root, id, paths, "vendored third-party code", expires, time.Now(), &out); err != nil {
		t.Fatalf("waiver sign %s: %v", id, err)
	}
}

// CLI-INV-56: the waiver record and `waiver sign`.
func TestWaiverSign(t *testing.T) {
	ctx := context.Background()
	root, _, _ := waiverRepo(t)
	var out bytes.Buffer
	if err := waiverSign(ctx, root, "vendor-lib", []string{"vendor/", "third_party/*.c"}, "vendored", day(30), time.Now(), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "commit this file before the work it waives") {
		t.Errorf("output = %s", out.String())
	}
	raw, err := os.ReadFile(filepath.Join(root, ".proofbound", "waivers", "vendor-lib.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(uiGit(t, root, "status", "--porcelain", "--", ".proofbound/waivers")) == "" {
		t.Error("waiver sign committed or wrote nothing")
	}
	fields, err := verifyRecord(raw, waiverSchema, declaredKeys(ctx, root, "HEAD"))
	if err != nil || fields["by"] != "approver" || fields["paths"] != "vendor/\nthird_party/*.c" || fields["expires"] != day(30) || fields["id"] != "vendor-lib" {
		t.Fatalf("record = %v, %v", fields, err)
	}
	if err := waiverSign(ctx, root, "vendor-lib", []string{"x/"}, "again", day(1), time.Now(), &out); err == nil || !strings.Contains(err.Error(), "exists") {
		t.Errorf("an existing waiver was overwritten: %v", err)
	}
	for _, bad := range []struct {
		id      string
		paths   []string
		reason  string
		expires string
		want    string
	}{
		{"Bad_ID", []string{"x/"}, "r", day(1), "id"},
		{"ok", []string{"/abs"}, "r", day(1), "absolute"},
		{"ok", []string{"a/../b"}, "r", day(1), ".."},
		{"ok", []string{"x/"}, " ", day(1), "reason"},
		{"ok", []string{"x/"}, "r", day(-1), "in the past"},
		{"ok", []string{"x/"}, "r", day(91), "more than 90 days"},
		{"ok", []string{"x/"}, "r", "2026-13-01", "not a date"},
	} {
		if err := waiverSign(ctx, root, bad.id, bad.paths, bad.reason, bad.expires, time.Now(), &out); err == nil || !strings.Contains(err.Error(), bad.want) {
			t.Errorf("waiverSign(%q, %q, %q, %q) = %v, want %q", bad.id, bad.paths, bad.reason, bad.expires, err, bad.want)
		}
	}
	if err := waiverSign(ctx, root, "ok", []string{"x/"}, "r", day(90), time.Now(), &out); err != nil {
		t.Errorf("90 days ahead was refused: %v", err)
	}
	// An undeclared local key cannot sign.
	other := taskRepo(t)
	uiCommit(t, other, "owner", "start")
	if err := waiverSign(ctx, other, "x", []string{"x/"}, "r", day(1), time.Now(), &out); err == nil || !strings.Contains(err.Error(), "not declared") {
		t.Errorf("an undeclared key signed: %v", err)
	}
	// Parsing: --path repeats; --reason and --expires exactly once.
	if id, paths, reason, expires, ok := parseWaiverSign([]string{"waiver", "sign", "w", "--path", "a/", "--reason", "r", "--path", "b/", "--expires", "2026-12-01"}); !ok || id != "w" || !reflect.DeepEqual(paths, []string{"a/", "b/"}) || reason != "r" || expires != "2026-12-01" {
		t.Errorf("parse = %q %q %q %q %v", id, paths, reason, expires, ok)
	}
	for _, args := range [][]string{
		{"waiver", "sign", "w", "--reason", "r", "--expires", "2026-12-01"},
		{"waiver", "sign", "w", "--path", "a/", "--reason", "r", "--reason", "s", "--expires", "2026-12-01"},
		{"waiver", "sign", "--path", "a/", "--reason", "r", "--expires", "2026-12-01"},
		{"waiver", "sign", "w", "--path", "a/", "--expires", "2026-12-01"},
		{"waiver", "sign", "w", "--path", "a/", "--reason", "r", "--expires", "2026-12-01", "--other", "x"},
	} {
		if _, _, _, _, ok := parseWaiverSign(args); ok || parseCommand(args) != commandInvalid {
			t.Errorf("parsed %q", args)
		}
	}
}

// CLI-INV-57: when waivers cover a commit.
func TestWaiverCovers(t *testing.T) {
	ctx := context.Background()
	root, base, _ := waiverRepo(t)
	signWaiver(t, root, "vendor-lib", []string{"vendor/"}, day(30))
	uiCommit(t, root, "approver", "waive vendor/")
	dev := newWaiverSubject("dev", "dev@example.com", nil, "")

	uiWrite(t, root, "vendor/a.c", "a\n")
	x := uiCommit(t, root, "dev", "vendor a")
	uses, unwaived, _ := waiversFor(ctx, root, base, x, []string{"vendor/a.c"}, dev)
	if len(uses) != 1 || uses[0].ID != "vendor-lib" || uses[0].By != "approver" || uses[0].Expires != day(30) || unwaived != nil {
		t.Fatalf("valid waiver: %+v %q", uses, unwaived)
	}
	// The deferral's done-when: a waiver signed by the commit's own author does not count.
	uses, unwaived, notes := waiversFor(ctx, root, base, x, []string{"vendor/a.c"}, newWaiverSubject("approver", "a@example.com", nil, ""))
	if uses != nil || len(unwaived) != 1 || len(notes) != 1 || !strings.Contains(notes[0], "signer approver is its author") {
		t.Errorf("author-signed waiver: %+v %q %q", uses, unwaived, notes)
	}
	if uses, _, notes := waiversFor(ctx, root, base, x, []string{"vendor/a.c"}, newWaiverSubject("dev", "dev@example.com", []string{"Approver <noreply@anthropic.com>"}, "")); uses != nil || !strings.Contains(strings.Join(notes, ""), "is a co-author") {
		t.Errorf("agent-signed waiver: %+v %q", uses, notes)
	}
	if uses, _, notes := waiversFor(ctx, root, base, x, []string{"vendor/a.c"}, newWaiverSubject("dev", "dev@example.com", nil, "approver")); uses != nil || !strings.Contains(strings.Join(notes, ""), "its signer") {
		t.Errorf("waiver signed by the commit's signer: %+v %q", uses, notes)
	}
	// Partly covered: the rest is named.
	uses, unwaived, _ = waiversFor(ctx, root, base, x, []string{"vendor/a.c", "src/x.c"}, dev)
	if uses != nil || !reflect.DeepEqual(unwaived, []string{"src/x.c"}) {
		t.Errorf("partly covered: %+v %q", uses, unwaived)
	}
	// Two waivers together cover everything.
	signWaiver(t, root, "src-hotfix", []string{"src/"}, day(10))
	uiCommit(t, root, "approver", "waive src/")
	uiWrite(t, root, "vendor/b.c", "b\n")
	uiWrite(t, root, "src/x.c", "x\n")
	both := uiCommit(t, root, "dev", "both")
	if uses, unwaived, _ := waiversFor(ctx, root, base, both, []string{"src/x.c", "vendor/b.c"}, dev); len(uses) != 2 || unwaived != nil {
		t.Errorf("two waivers: %+v %q", uses, unwaived)
	}
	// Only in the commit itself, not in its parent: no.
	signWaiver(t, root, "late", []string{"lib/"}, day(10))
	uiWrite(t, root, "lib/l.c", "l\n")
	late := uiCommit(t, root, "dev", "waiver and code together")
	if uses, unwaived, _ := waiversFor(ctx, root, base, late, []string{"lib/l.c"}, dev); uses != nil || len(unwaived) != 1 {
		t.Errorf("a waiver written with its code counted: %+v", uses)
	}
	// Expired by the commit's date.
	uiWrite(t, root, "vendor/c.c", "c\n")
	uiGit(t, root, "add", "-A")
	later := exec.Command("git", "-C", root, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "--no-verify", "-m", "after expiry")
	later.Env = append(repositoryGitEnv(), "GIT_COMMITTER_DATE="+time.Now().UTC().AddDate(0, 0, 40).Format(time.RFC3339))
	if out, err := later.CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	expired := uiGit(t, root, "rev-parse", "HEAD")
	if uses, _, notes := waiversFor(ctx, root, base, expired, []string{"vendor/c.c"}, dev); uses != nil || !strings.Contains(strings.Join(notes, ""), "expired on "+day(30)) {
		t.Errorf("an expired waiver counted: %+v %q", uses, notes)
	}
	// Signed by a key the repository does not declare under that name.
	forged, err := signRecord(map[string]any{"schema": waiverSchema, "id": "forged", "paths": "forged/", "reason": "r", "expires": day(5), "by": "approver", "key": "0000000000000000", "at": "2026-11-23T00:00:00Z"}, ed25519.NewKeyFromSeed(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	uiWrite(t, root, waiverPath("forged"), string(forged))
	uiCommit(t, root, "dev", "a forged waiver")
	uiWrite(t, root, "forged/f.c", "f\n")
	f := uiCommit(t, root, "dev", "forged code")
	if uses, _, notes := waiversFor(ctx, root, base, f, []string{"forged/f.c"}, dev); uses != nil || !strings.Contains(strings.Join(notes, ""), "does not verify") {
		t.Errorf("a forged waiver counted: %+v %q", uses, notes)
	}
}

// CLI-INV-58 (check): a waived commit is listed as WAIVED, named in the headline, and does not fail
// the pull request; an unwaived one still does.
func TestCheckWaived(t *testing.T) {
	ctx := context.Background()
	root, base, _ := waiverRepo(t)
	signWaiver(t, root, "vendor-lib", []string{"vendor/"}, day(30))
	uiCommit(t, root, "approver", "waive vendor/")
	uiWrite(t, root, "vendor/a.c", "a\n")
	waived := uiCommit(t, root, "dev", "vendor a")
	var out bytes.Buffer
	if err := checkRange(ctx, root, nil, base, false, false, &out); err != nil {
		t.Fatalf("a waived range is not warranted: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "; 1 commit waived)") || !strings.Contains(out.String(), waived[:12]+" - WAIVED  waived by approver (waiver vendor-lib, until "+day(30)+"): vendored third-party code") {
		t.Errorf("plain output:\n%s", out.String())
	}
	out.Reset()
	_ = checkRange(ctx, root, nil, base, true, false, &out)
	if !strings.Contains(out.String(), "| **WAIVED** | waived by approver") {
		t.Errorf("markdown output:\n%s", out.String())
	}
	// The waiver's own signer committing the same kind of change is not waived.
	uiWrite(t, root, "vendor/b.c", "b\n")
	uiCommit(t, root, "approver", "vendor b")
	out.Reset()
	if err := checkRange(ctx, root, nil, base, false, false, &out); err == nil || !strings.Contains(out.String(), "NOT WARRANTED") || !strings.Contains(out.String(), "UNCOVERED") || !strings.Contains(out.String(), "signer approver is its author") {
		t.Errorf("a self-waived commit passed: %v\n%s", err, out.String())
	}
}

// CLI-INV-58 (bundle): index.json lists waived commits, VERIFY.md shows them, the badge counts them; no
// `waived` field without waivers.
func TestBundleWaived(t *testing.T) {
	ctx := context.Background()
	root, base, _ := waiverRepo(t)
	dir := filepath.Join(t.TempDir(), "none")
	var out bytes.Buffer
	if err := warrantyBundle(ctx, root, nil, base, dir, &out); err != nil {
		t.Fatal(err)
	}
	if ix, _ := os.ReadFile(filepath.Join(dir, "index.json")); strings.Contains(string(ix), "waived") {
		t.Errorf("an index without waivers mentions them: %s", ix)
	}
	signWaiver(t, root, "vendor-lib", []string{"vendor/"}, day(30))
	uiCommit(t, root, "approver", "waive vendor/")
	uiWrite(t, root, "vendor/a.c", "a\n")
	waived := uiCommit(t, root, "dev", "vendor a")
	dir = filepath.Join(t.TempDir(), "w")
	if err := warrantyBundle(ctx, root, nil, base, dir, &out); err != nil {
		t.Fatal(err)
	}
	ix, _ := os.ReadFile(filepath.Join(dir, "index.json"))
	verify, _ := os.ReadFile(filepath.Join(dir, "VERIFY.md"))
	badge, _ := os.ReadFile(filepath.Join(dir, "badge.svg"))
	if !strings.Contains(string(ix), `"waived":[{"by":"approver","commit":"`+waived+`","expires":"`+day(30)+`","reason":"vendored third-party code","waiver":"vendor-lib"}]`) {
		t.Errorf("index = %s", ix)
	}
	if !strings.Contains(string(verify), "## Waived commits") || !strings.Contains(string(verify), "| `"+waived[:12]+"` | `vendor-lib` | approver |") {
		t.Errorf("VERIFY.md = %s", verify)
	}
	if !strings.Contains(string(badge), "no claims · 1 waived") {
		t.Errorf("badge = %s", badge)
	}
}

// CLI-INV-59: C2 shows a waived commit as evidence, not an exception; C6 lists waiver changes.
func TestAuditWaived(t *testing.T) {
	waived := checkLine{commit: strings.Repeat("a", 40), state: checkWaived, reason: "waived by approver (waiver vendor-lib, until 2026-12-01): vendored"}
	uncovered := checkLine{commit: strings.Repeat("b", 40), state: checkUncovered, reason: "claims no spec but changes src/x.c; run ..."}
	var ex []auditException
	except := func(kind, control, commit, text string) {
		ex = append(ex, auditException{Kind: kind, Commit: commit, Text: text})
	}
	subject := func(string) string { return "s" }
	c := uncoveredControl("base", 3, nil, []checkLine{waived}, subject, except)
	if c.State != auditHeld || len(ex) != 0 || !strings.Contains(c.Count, "1 commit waived") || c.Evidence[0].Text != waived.reason {
		t.Errorf("waived only: %+v %+v", c, ex)
	}
	c = uncoveredControl("base", 3, []checkLine{uncovered}, []checkLine{waived}, subject, except)
	if c.State != auditExcepted || len(ex) != 1 || ex[0].Commit != uncovered.commit {
		t.Errorf("with an uncovered commit: %+v %+v", c, ex)
	}
	ctx := context.Background()
	root, _, _ := waiverRepo(t)
	signWaiver(t, root, "vendor-lib", []string{"vendor/"}, day(30))
	sha := uiCommit(t, root, "approver", "waive vendor/")
	paths, err := trustPaths(ctx, root, sha)
	if err != nil || !reflect.DeepEqual(paths, []string{".proofbound/waivers/vendor-lib.json"}) {
		t.Errorf("C6 paths = %q %v", paths, err)
	}
}

// CLI-INV-57 (hook): the commit-msg hook lets through staged code that waivers at HEAD cover for this
// author, and still refuses the waiver's own signer.
func TestHookHonoursWaiver(t *testing.T) {
	ctx := context.Background()
	root, _, _ := waiverRepo(t)
	signWaiver(t, root, "vendor-lib", []string{"vendor/"}, day(30))
	uiCommit(t, root, "approver", "waive vendor/")
	uiWrite(t, root, "vendor/a.c", "a\n")
	uiGit(t, root, "add", "vendor/a.c")
	msg := filepath.Join(t.TempDir(), "MSG")
	if err := os.WriteFile(msg, []byte("vendor a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runHook(ctx, root, "commit-msg", msg, "", &out); err != nil || !strings.Contains(out.String(), "waived by approver (waiver vendor-lib") {
		t.Fatalf("a waived commit was refused: %v %s", err, out.String())
	}
	uiGit(t, root, "config", "user.name", "approver")
	out.Reset()
	if err := runHook(ctx, root, "commit-msg", msg, "", &out); err == nil || !strings.Contains(out.String(), "signer approver is its author") {
		t.Errorf("the waiver's signer passed the hook: %v %s", err, out.String())
	}
	uiGit(t, root, "config", "user.name", "a")
	// git hands a hook `-c user.name=…` and GIT_AUTHOR_NAME in the environment; the hook must see them.
	t.Setenv("GIT_AUTHOR_NAME", "approver")
	out.Reset()
	if err := runHook(ctx, root, "commit-msg", msg, "", &out); err == nil || !strings.Contains(out.String(), "signer approver is its author") {
		t.Errorf("the hook did not see the author git passed it: %v %s", err, out.String())
	}
	t.Setenv("GIT_AUTHOR_NAME", "")
	os.Unsetenv("GIT_AUTHOR_NAME")
	if err := os.WriteFile(msg, []byte("vendor a\n\nCo-Authored-By: Approver <noreply@anthropic.com>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runHook(ctx, root, "commit-msg", msg, "", &out); err == nil || !strings.Contains(out.String(), "is a co-author") {
		t.Errorf("an agent named on the commit as the signer passed the hook: %v %s", err, out.String())
	}
	uiWrite(t, root, "src/x.c", "x\n")
	uiGit(t, root, "add", "src/x.c")
	if err := os.WriteFile(msg, []byte("vendor and src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runHook(ctx, root, "commit-msg", msg, "", &out); err == nil {
		t.Error("partly waived staged code passed the hook")
	}
}

// handWaiver writes a waiver signed by key outside `waiver sign`, as an attacker could.
func handWaiver(t *testing.T, root string, key ed25519.PrivateKey, id, by, paths, at, expires string) {
	t.Helper()
	data, err := signRecord(map[string]any{"schema": waiverSchema, "id": id, "paths": paths, "reason": "r", "expires": expires, "by": by,
		"key": pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey)), "at": at}, key)
	if err != nil {
		t.Fatal(err)
	}
	uiWrite(t, root, waiverPath(id), string(data))
}

// Review signed-waiver-review-20261123 F1..F4: each way an author could waive their own work.
func TestWaiverReviewFindings(t *testing.T) {
	ctx := context.Background()
	notesOf := func(root, base, sha string, code []string, who waiverSubject) ([]waiverUse, string) {
		uses, _, notes := waiversFor(ctx, root, base, sha, code, who)
		return uses, strings.Join(notes, " | ")
	}
	t.Run("F1 a key name unlike the git name: the e-mail local part still identifies the author", func(t *testing.T) {
		root, base, _ := waiverRepo(t)
		signWaiver(t, root, "w", []string{"vendor/"}, day(30))
		uiCommit(t, root, "approver", "waive")
		uiWrite(t, root, "vendor/a.c", "a\n")
		sha := uiCommit(t, root, "dev", "a")
		if uses, notes := notesOf(root, base, sha, []string{"vendor/a.c"}, newWaiverSubject("Approver Smith", "approver@example.com", nil, "")); uses != nil || !strings.Contains(notes, "is its author") {
			t.Errorf("%+v %s", uses, notes)
		}
	})
	t.Run("F1 one key declared under a second name is the same person", func(t *testing.T) {
		key := uiKey(t)
		root := taskRepo(t)
		pub := pbwarranty.PubLine(key.Public().(ed25519.PublicKey))
		uiWrite(t, root, ".proofbound/keys/approver.pub", pub)
		uiWrite(t, root, ".proofbound/keys/alice.pub", pub)
		base := uiCommit(t, root, "owner", "keys")
		handWaiver(t, root, key, "w", "approver", "vendor/", time.Now().UTC().Format(time.RFC3339), day(30))
		uiCommit(t, root, "alice", "waive")
		uiWrite(t, root, "vendor/a.c", "a\n")
		sha := uiCommit(t, root, "alice", "a")
		if uses, notes := notesOf(root, base, sha, []string{"vendor/a.c"}, newWaiverSubject("alice", "alice@example.com", nil, "")); uses != nil || !strings.Contains(notes, "also declared as alice") {
			t.Errorf("%+v %s", uses, notes)
		}
	})
	t.Run("F1 a key added inside the range cannot waive anything in it", func(t *testing.T) {
		root, base, _ := waiverRepo(t)
		puppet := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, 32))
		uiWrite(t, root, ".proofbound/keys/puppet.pub", pbwarranty.PubLine(puppet.Public().(ed25519.PublicKey)))
		handWaiver(t, root, puppet, "w", "puppet", "vendor/", time.Now().UTC().Format(time.RFC3339), day(30))
		uiCommit(t, root, "dev", "a key and a waiver")
		uiWrite(t, root, "vendor/a.c", "a\n")
		sha := uiCommit(t, root, "dev", "a")
		if uses, notes := notesOf(root, base, sha, []string{"vendor/a.c"}, newWaiverSubject("dev", "dev@example.com", nil, "")); uses != nil || !strings.Contains(notes, "declared both at the range base") {
			t.Errorf("%+v %s", uses, notes)
		}
	})
	t.Run("F2 a human co-author who signed the waiver is a writer", func(t *testing.T) {
		root, base, _ := waiverRepo(t)
		signWaiver(t, root, "w", []string{"vendor/"}, day(30))
		uiCommit(t, root, "approver", "waive")
		uiWrite(t, root, "vendor/a.c", "a\n")
		sha := uiCommit(t, root, "dev", "a\n\nCo-authored-by: Approver <approver@example.com>")
		var out bytes.Buffer
		if err := checkRange(ctx, root, nil, base, false, false, &out); err == nil || !strings.Contains(out.String(), "is a co-author") {
			t.Errorf("a co-author's own waiver passed check: %v\n%s", err, out.String())
		}
		_ = sha
	})
	t.Run("F3 a backdated commit cannot use an expired waiver", func(t *testing.T) {
		root, _, key := waiverRepo(t)
		at := time.Now().UTC().AddDate(0, 0, -60)
		handWaiver(t, root, key, "w", "approver", "vendor/", at.Format(time.RFC3339), at.AddDate(0, 0, 20).Format("2006-01-02"))
		uiCommit(t, root, "approver", "an old waiver")
		base := uiGit(t, root, "rev-parse", "HEAD~1")
		uiWrite(t, root, "vendor/a.c", "a\n")
		uiGit(t, root, "add", "-A")
		back := exec.Command("git", "-C", root, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "--no-verify", "-m", "backdated")
		back.Env = append(repositoryGitEnv(), "GIT_COMMITTER_DATE="+at.AddDate(0, 0, 5).Format(time.RFC3339))
		if out, err := back.CombinedOutput(); err != nil {
			t.Fatalf("%v %s", err, out)
		}
		sha := uiGit(t, root, "rev-parse", "HEAD")
		if uses, notes := notesOf(root, base, sha, []string{"vendor/a.c"}, newWaiverSubject("dev", "dev@example.com", nil, "")); uses != nil || !strings.Contains(notes, "expired on") {
			t.Errorf("%+v %s", uses, notes)
		}
	})
	t.Run("F4 a hand-signed record must keep the date rules", func(t *testing.T) {
		root, base, key := waiverRepo(t)
		now := time.Now().UTC().Format(time.RFC3339)
		handWaiver(t, root, key, "forever", "approver", "a/", now, "9999-12-31")
		handWaiver(t, root, key, "notadate", "approver", "b/", now, "zzzz")
		handWaiver(t, root, key, "noat", "approver", "c/", "yesterday", day(5))
		uiCommit(t, root, "approver", "bad waivers")
		uiWrite(t, root, "a/x.c", "x\n")
		uiWrite(t, root, "b/x.c", "x\n")
		uiWrite(t, root, "c/x.c", "x\n")
		sha := uiCommit(t, root, "dev", "x")
		uses, notes := notesOf(root, base, sha, []string{"a/x.c", "b/x.c", "c/x.c"}, newWaiverSubject("dev", "dev@example.com", nil, ""))
		if uses != nil || !strings.Contains(notes, "lasts more than 90 days") || !strings.Contains(notes, "expiry is not a date") || !strings.Contains(notes, "not an RFC 3339 time") {
			t.Errorf("%+v %s", uses, notes)
		}
	})
}
