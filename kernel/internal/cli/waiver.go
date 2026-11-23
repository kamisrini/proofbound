package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// Signed exception waivers (docs/plans/signed-waiver-spec.md, CLI-INV-56..59): a signed, path-scoped,
// expiring record, approved before the work by someone independent of the commit, turns an UNCOVERED
// commit into WAIVED, and every waived commit is shown wherever a reader looks.

const (
	waiverSchema  = "proofbound.waiver.v1"
	waiversDir    = ".proofbound/waivers"
	waiverMaxDays = 90
	// checkWaived is an unclaimed commit whose every code path a valid waiver covers (CLI-INV-57).
	checkWaived = "WAIVED"
)

var waiverIDRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// waiverUse is one waiver that covers (part of) a commit.
type waiverUse struct {
	ID      string `json:"waiver"`
	By      string `json:"by"`
	Expires string `json:"expires"`
	Reason  string `json:"reason"`
}

func waiverPath(id string) string { return waiversDir + "/" + id + ".json" }

// parseWaiverSign accepts `waiver sign <id> --path <p> [--path <p>]... --reason <r> --expires <d>`;
// --path may repeat, the other flags appear exactly once.
func parseWaiverSign(args []string) (id string, paths []string, reason, expires string, ok bool) {
	if len(args) < 3 || args[0] != "waiver" || args[1] != "sign" || args[2] == "" || strings.HasPrefix(args[2], "-") {
		return "", nil, "", "", false
	}
	id = args[2]
	rest := args[3:]
	if len(rest)%2 != 0 {
		return "", nil, "", "", false
	}
	seen := map[string]bool{}
	for i := 0; i < len(rest); i += 2 {
		flag, value := rest[i], rest[i+1]
		if value == "" {
			return "", nil, "", "", false
		}
		switch flag {
		case "--path":
			paths = append(paths, value)
		case "--reason", "--expires":
			if seen[flag] {
				return "", nil, "", "", false
			}
			seen[flag] = true
			if flag == "--reason" {
				reason = value
			} else {
				expires = value
			}
		default:
			return "", nil, "", "", false
		}
	}
	return id, paths, reason, expires, len(paths) > 0 && seen["--reason"] && seen["--expires"]
}

// waiverSign writes a new waiver signed with this machine's declared key (CLI-INV-56). It never
// overwrites a waiver and never commits.
func waiverSign(ctx context.Context, root, id string, paths []string, reason, expires string, now time.Time, output io.Writer) error {
	if len(id) > 64 || !waiverIDRE.MatchString(id) {
		return fmt.Errorf("waiver: id %q must be lowercase letters and digits joined by '-', at most 64 characters", id)
	}
	for _, p := range paths {
		if err := checkAreaPattern(p); err != nil {
			return fmt.Errorf("waiver: path %q %v", p, err)
		}
	}
	if err := checkReason(reason); err != nil {
		return fmt.Errorf("waiver: %v", err)
	}
	day, err := time.Parse("2006-01-02", expires)
	if err != nil || day.Format("2006-01-02") != expires {
		return fmt.Errorf("waiver: --expires %q is not a date (YYYY-MM-DD)", expires)
	}
	today := now.UTC().Truncate(24 * time.Hour)
	if day.Before(today) {
		return fmt.Errorf("waiver: --expires %s is in the past", expires)
	}
	if day.After(today.AddDate(0, 0, waiverMaxDays)) {
		return fmt.Errorf("waiver: --expires %s is more than %d days ahead; a waiver is never permanent", expires, waiverMaxDays)
	}
	a := localActor(ctx, root)
	if a.key == nil {
		return fmt.Errorf("waiver: %s", a.Problem)
	}
	data, err := signRecord(map[string]any{"schema": waiverSchema, "id": id, "paths": strings.Join(paths, "\n"), "reason": reason, "expires": expires,
		"by": a.Name, "key": a.Fingerprint, "at": now.UTC().Format(time.RFC3339)}, a.key)
	if err != nil {
		return err
	}
	rel := waiverPath(id)
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("waiver: %s exists; choose another id", rel)
		}
		return err
	}
	if _, err := f.Write(data); err != nil {
		return errors.Join(err, f.Close(), os.Remove(full))
	}
	if err := f.Close(); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(output, "wrote %s\nsigned as %s (key %s), until %s\ncommit this file before the work it waives: a waiver counts only from a commit's parent, and never for its own signer's commits\n", rel, a.Name, a.Fingerprint, expires); err != nil {
		return err
	}
	// A project set up before waivers existed ignores .proofbound/*: say how to commit it anyway.
	if strings.TrimSpace(gitOutput(ctx, root, "check-ignore", "--no-index", "--", rel)) != "" {
		_, err = fmt.Fprintf(output, "%s is ignored by .gitignore here: commit it with `git add -f %s`, or run `proofbound init` again to stop ignoring %s/\n", rel, rel, waiversDir)
	}
	return err
}

// waiverSubject is who made a commit, for the independence rule: every name git records as writing it
// (author and every co-author, by name and by e-mail local part) and its verified signer's key name.
// Git names are self-declared; the rule therefore also compares public keys (review F1).
type waiverSubject struct {
	writers map[string]string // key name -> role ("its author", "a co-author")
	signer  string
	// signed is true when the commit carries a signature verified against a declared key (C9), or, in
	// the hook, when git is configured to sign it (CLI-INV-63).
	signed bool
}

// newWaiverSubject builds the writer set from an author identity and co-author trailer values
// ("Name <e-mail>"), all of them, not only agents (review F2).
func newWaiverSubject(authorName, authorEmail string, coauthors []string, signer string) waiverSubject {
	w := waiverSubject{writers: map[string]string{}, signer: signer}
	add := func(name, role string) {
		if k := pbwarranty.KeyName(name); k != "" {
			if _, seen := w.writers[k]; !seen {
				w.writers[k] = role
			}
		}
	}
	add(authorName, "its author")
	add(emailLocal(authorEmail), "its author")
	for _, co := range coauthors {
		name, email, _ := strings.Cut(strings.TrimSpace(co), "<")
		add(strings.TrimSpace(name), "a co-author")
		add(emailLocal(strings.TrimSuffix(strings.TrimSpace(email), ">")), "a co-author")
	}
	if signer != "" {
		w.writers[signer] = "its signer"
	}
	return w
}

func emailLocal(email string) string {
	local, _, _ := strings.Cut(strings.TrimSpace(email), "@")
	return local
}

// independent reports whether a waiver signed by `by` with key pub is independent of the commit: `by`
// is no writer's key name, and pub is not the key any writer's name declares (one key under two names
// is one person, review F1).
func (w waiverSubject) independent(by string, pub ed25519.PublicKey, keys map[string]ed25519.PublicKey) (bool, string) {
	if role, ok := w.writers[pbwarranty.KeyName(by)]; ok {
		return false, "signer " + by + " is " + role
	}
	for name, role := range w.writers {
		if k, ok := keys[name]; ok && k.Equal(pub) {
			return false, "signer " + by + "'s key is also declared as " + name + ", " + role
		}
	}
	return true, ""
}

// waiversFor decides whether the waivers in sha's first parent cover its code paths (CLI-INV-57) for
// the range that starts at base. It returns the waivers used, the paths left unwaived, and, for those
// paths, why each waiver that names them did not count.
func waiversFor(ctx context.Context, root, base, sha string, code []string, who waiverSubject) (uses []waiverUse, unwaived, notes []string) {
	parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
	if len(parent) != 40 {
		return nil, code, nil
	}
	// The day a waiver must still be valid on: the commit's committer date, but never earlier than the
	// range base's, which the commit's author cannot move (review F3).
	floor := laterDay(commitDay(ctx, root, sha), commitDay(ctx, root, base))
	return waiversAt(ctx, root, parent, base, floor, code, who)
}

func commitDay(ctx context.Context, root, rev string) string {
	secs, err := strconv.ParseInt(strings.TrimSpace(gitOutput(ctx, root, "log", "-1", "--format=%ct", "--end-of-options", rev)), 10, 64)
	if err != nil {
		return ""
	}
	return time.Unix(secs, 0).UTC().Format("2006-01-02")
}

func laterDay(a, b string) string {
	if a > b {
		return a
	}
	return b
}

// waiversAt applies the waiver rule to a commit whose first parent is parent, checked on day floor
// or later: for a commit in history (waiversFor, base = the range base), or for the commit the
// commit-msg hook is about to make on top of HEAD (stagedWaived, base = "", floor = today). A waiver
// counts only when signed by a key declared in the parent and, when base is given, already declared
// with the same key at the base: a key added inside the range cannot waive anything in it (F1).
func waiversAt(ctx context.Context, root, parent, base, floor string, code []string, who waiverSubject) (uses []waiverUse, unwaived, notes []string) {
	listing := strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "--name-only", parent, "--", waiversDir+"/"))
	if listing == "" {
		return nil, code, nil
	}
	// CLI-INV-63: with the policy on (at the base or before the commit), only signed commits are waived.
	policyOn, policyNote := waiverPolicy(ctx, root, parent)
	if base != "" {
		if on, note := waiverPolicy(ctx, root, base); on {
			policyOn = true
			if policyNote == "" {
				policyNote = note
			}
		}
	}
	parentKeys := declaredKeys(ctx, root, parent)
	keys := parentKeys
	writerKeys := parentKeys
	if base != "" {
		baseKeys := declaredKeys(ctx, root, base)
		keys = map[string]ed25519.PublicKey{}
		for name, k := range parentKeys {
			if b, ok := baseKeys[name]; ok && b.Equal(k) {
				keys[name] = k
			}
		}
		writerKeys = map[string]ed25519.PublicKey{}
		for name, k := range baseKeys {
			writerKeys[name] = k
		}
		for name, k := range parentKeys {
			writerKeys[name] = k
		}
	}
	type candidate struct {
		use      waiverUse
		patterns []string
		problem  string
	}
	var cands []candidate
	for _, p := range strings.Split(listing, "\n") {
		p = strings.TrimSpace(p)
		id, ok := strings.CutSuffix(path.Base(p), ".json")
		if !ok || path.Dir(p) != waiversDir || !waiverIDRE.MatchString(id) || p == waiverPolicyPath {
			continue
		}
		raw, err := gitBytes(ctx, root, "show", parent+":"+p)
		if err != nil {
			continue
		}
		// The paths are read before verification only to decide whether a refusal is worth naming.
		var loose struct {
			Paths string `json:"paths"`
		}
		_ = json.Unmarshal(raw, &loose)
		patterns, perr := parseArea([]byte(loose.Paths))
		c := candidate{patterns: patterns}
		fields, err := verifyRecord(raw, waiverSchema, keys)
		switch {
		case err != nil:
			c.problem = "does not verify against a key declared both at the range base and before the commit: " + err.Error()
		case fields["id"] != id:
			c.problem = "its id does not match its file name"
		case perr != nil || len(patterns) == 0:
			c.problem = "its paths are invalid"
		default:
			c.use = waiverUse{ID: id, By: fields["by"], Expires: fields["expires"], Reason: printableUnit(fields["reason"])}
			signedDay, derr := waiverDates(fields["at"], fields["expires"])
			if derr != nil {
				c.problem = derr.Error()
			} else if ok, why := who.independent(fields["by"], keys[fields["by"]], writerKeys); !ok {
				c.problem = "is not independent of the commit: " + why
			} else if day := laterDay(floor, signedDay); fields["expires"] < day {
				c.problem = fmt.Sprintf("expired on %s (it must still be valid on %s: the latest of the commit's date, the range base's date and its own signing date)", fields["expires"], day)
			} else if policyOn && !signedByKnownKey(ctx, root, base, who, keys) {
				c.problem = "does not apply: this project waives only commits signed by a key declared at the range base (" + waiverPolicyPath + policyNote + "); this commit is not"
			}
		}
		c.use.ID = id
		cands = append(cands, c)
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].use.ID < cands[j].use.ID })
	used := map[string]bool{}
	noted := map[string]bool{}
	for _, p := range code {
		covered := false
		for _, c := range cands {
			if c.problem == "" && inArea(c.patterns, p) {
				covered, used[c.use.ID] = true, true
			}
		}
		if covered {
			continue
		}
		unwaived = append(unwaived, p)
		for _, c := range cands {
			if c.problem != "" && inArea(c.patterns, p) && !noted[c.use.ID] {
				noted[c.use.ID] = true
				notes = append(notes, "waiver "+c.use.ID+" "+c.problem)
			}
		}
	}
	for _, c := range cands {
		if used[c.use.ID] {
			uses = append(uses, c.use)
		}
	}
	if len(unwaived) > 0 {
		return nil, unwaived, notes
	}
	return uses, nil, nil
}

// waiverDates enforces, on every record read and not only at signing (review F4), that `at` is an
// RFC 3339 time, `expires` a real date, and that the waiver lasts at most waiverMaxDays from `at`. It
// returns the UTC day it was signed.
func waiverDates(at, expires string) (string, error) {
	signed, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return "", errors.New("its signing time is not an RFC 3339 time")
	}
	end, err := time.Parse("2006-01-02", expires)
	if err != nil || end.Format("2006-01-02") != expires {
		return "", errors.New("its expiry is not a date")
	}
	day := signed.UTC().Format("2006-01-02")
	if expires < day {
		return "", errors.New("it expires before it was signed")
	}
	start, _ := time.Parse("2006-01-02", day)
	if end.After(start.AddDate(0, 0, waiverMaxDays)) {
		return "", fmt.Errorf("it lasts more than %d days", waiverMaxDays)
	}
	return day, nil
}

// signedByKnownKey reports whether the commit counts as signed under the policy (CLI-INV-63). In
// history (base given), the signature C9 verified must be by a key declared, unchanged, at the range
// base, in the commit's parent and at HEAD (C9 verifies against HEAD): a key added or swapped inside
// the range cannot make a commit count as signed (review waiver-signed-commits-review-20261123 F1).
// In the hook (no base), git being configured to sign is the guidance.
func signedByKnownKey(ctx context.Context, root, base string, who waiverSubject, keys map[string]ed25519.PublicKey) bool {
	if !who.signed {
		return false
	}
	if base == "" {
		return true
	}
	k, ok := keys[who.signer]
	if !ok {
		return false
	}
	head, ok := declaredKeys(ctx, root, "HEAD")[who.signer]
	return ok && head.Equal(k)
}

// waiverPolicyPath is the opt-in switch: waivers cover only signed commits (CLI-INV-63).
const waiverPolicyPath = waiversDir + "/policy.json"

// waiverPolicy reports whether the policy file exists at rev. Any file there turns the policy on; a
// file that is not exactly the expected form still does (fail closed) and the note says so.
func waiverPolicy(ctx context.Context, root, rev string) (bool, string) {
	raw, err := gitBytes(ctx, root, "show", rev+":"+waiverPolicyPath)
	if err != nil {
		return false, ""
	}
	var p struct {
		Schema  string `json:"schema"`
		Require *bool  `json:"require_signed_commits"`
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if dec.Decode(&p) != nil || p.Schema != "proofbound.waiver-policy.v1" || p.Require == nil || !*p.Require {
		return true, ", malformed, so it applies"
	}
	return true, ""
}

// signingConfigured reports whether git will sign the commit the hook is checking, as `keys git-sign`
// configures it; read with the hook's own environment.
func signingConfigured(ctx context.Context, root string) bool {
	get := func(key string) string {
		cmd := exec.CommandContext(ctx, "git", "-C", root, "config", "--get", key)
		cmd.Env = os.Environ()
		out, _ := cmd.Output()
		return strings.TrimSpace(string(out))
	}
	return get("commit.gpgsign") == "true" && get("gpg.format") == "ssh" && get("user.signingkey") != ""
}

// waivedReason is the check line for a waived commit.
func waivedReason(uses []waiverUse) string {
	parts := make([]string, 0, len(uses))
	for _, u := range uses {
		parts = append(parts, fmt.Sprintf("waived by %s (waiver %s, until %s): %s", u.By, u.ID, u.Expires, u.Reason))
	}
	return strings.Join(parts, "; ")
}

func waivedPhrase(n int) string {
	if n == 1 {
		return "1 commit waived"
	}
	return fmt.Sprintf("%d commits waived", n)
}

// stagedWaived reports whether waivers at HEAD cover every staged code path of the commit the
// commit-msg hook is about to make, for its author and agents and today's date (CLI-INV-57, hook). The
// hook is guidance; the pull-request check judges the commit again from history.
func stagedWaived(ctx context.Context, root, msg string, code []string) ([]waiverUse, []string) {
	head := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", "HEAD"))
	if len(head) != 40 {
		return nil, nil
	}
	// The author git is about to record: read with the hook's own environment, since git passes
	// `-c user.name=…` and GIT_AUTHOR_NAME to its hooks there (repositoryGitEnv would drop them).
	cmd := exec.CommandContext(ctx, "git", "-C", root, "var", "GIT_AUTHOR_IDENT")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return nil, []string{"cannot tell who is committing: " + err.Error()}
	}
	ident := strings.TrimSpace(string(out))
	name, rest, _ := strings.Cut(ident, " <")
	email, _, _ := strings.Cut(rest, ">")
	var coauthors []string
	for _, line := range strings.Split(msg, "\n") {
		if v, ok := cutFold(strings.TrimSpace(line), "co-authored-by:"); ok {
			coauthors = append(coauthors, strings.TrimSpace(v))
		}
	}
	who := newWaiverSubject(name, email, coauthors, "")
	who.signed = signingConfigured(ctx, root)
	uses, _, notes := waiversAt(ctx, root, head, "", time.Now().UTC().Format("2006-01-02"), code, who)
	return uses, notes
}

func cutFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", false
	}
	return s[len(prefix):], true
}
