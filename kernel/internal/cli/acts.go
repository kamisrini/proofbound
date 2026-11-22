package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The two signed records a person makes from the Audit room (U1 spec section 4, CLI-INV-39/40).

const (
	outcomeSchema     = "proofbound.exception-outcome.v1"
	attestationSchema = "proofbound.audit-attestation.v1"
	outcomesDir       = ".proofbound/outcomes"
	attestationsDir   = ".proofbound/attestations"
)

var (
	exceptionIDRE = regexp.MustCompile(`^(gate|uncovered|warranty|trust|order)-[0-9a-f]{12}$`)
	sha40RE       = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64RE       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	keyNameFileRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// The fields each schema carries, besides `signature`; nothing else is accepted.
var recordFields = map[string][]string{
	outcomeSchema:     {"schema", "exception", "commit", "outcome", "reason", "fixed_by", "by", "key", "at"},
	attestationSchema: {"schema", "base", "head", "transcript_sha256", "headline", "by", "key", "at"},
}

type outcomeView struct {
	Outcome  string `json:"outcome"`
	Reason   string `json:"reason"`
	FixedBy  string `json:"fixed_by,omitempty"`
	By       string `json:"by"`
	At       string `json:"at"`
	Verified bool   `json:"verified"`
	Problem  string `json:"problem,omitempty"`
	Path     string `json:"path"`
}

type attestationView struct {
	Path     string `json:"path"`
	Base     string `json:"base"`
	Head     string `json:"head"`
	Headline string `json:"headline"`
	By       string `json:"by"`
	At       string `json:"at"`
	Verified bool   `json:"verified"`
	Problem  string `json:"problem,omitempty"`
	// Moved counts the commits after its head; RecordsOnly counts those that only add Proofbound
	// records (outcomes, attestations). An attestation covers its own head, never a later one.
	Moved       int `json:"moved"`
	RecordsOnly int `json:"records_only"`
}

// actor is the person this machine signs as: the local key, when HEAD declares it.
type actor struct {
	Name        string `json:"name,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Problem     string `json:"problem,omitempty"`
	key         ed25519.PrivateKey
}

func recordTag(schema string) []byte { return []byte(schema + "\n") }

// declaredKeys are the keys commit rev declares, by name: `.proofbound/keys/<name>.pub` holding
// exactly `ed25519 <64 lowercase hex>\n`. Malformed files declare nothing.
func declaredKeys(ctx context.Context, root, rev string) map[string]ed25519.PublicKey {
	keys := map[string]ed25519.PublicKey{}
	for _, p := range strings.Split(gitOutput(ctx, root, "ls-tree", "--name-only", rev, "--", pbwarranty.KeysDir+"/"), "\n") {
		p = strings.TrimSpace(p)
		name, ok := strings.CutSuffix(path.Base(p), ".pub")
		if !ok || path.Dir(p) != pbwarranty.KeysDir || !keyNameFileRE.MatchString(name) {
			continue
		}
		raw, err := gitBytes(ctx, root, "show", rev+":"+p)
		if err != nil {
			continue
		}
		hexKey, ok := strings.CutPrefix(string(raw), "ed25519 ")
		hexKey, ok2 := strings.CutSuffix(hexKey, "\n")
		pub, err := hex.DecodeString(hexKey)
		if !ok || !ok2 || err != nil || len(pub) != ed25519.PublicKeySize || hex.EncodeToString(pub) != hexKey {
			continue
		}
		keys[name] = ed25519.PublicKey(pub)
	}
	return keys
}

// localActor loads the local signing key and names it from HEAD's declared keys.
func localActor(ctx context.Context, root string) actor {
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return actor{Problem: "no signing key location: " + err.Error()}
	}
	key, err := pbwarranty.LoadKey(keyPath)
	if err != nil {
		return actor{Problem: "this machine has no signing key: run `proofbound keys init --name <you>`, then commit the .pub file it writes"}
	}
	pub := key.Public().(ed25519.PublicKey)
	a := actor{Fingerprint: pbwarranty.Fingerprint(pub)}
	var names []string
	for name, declared := range declaredKeys(ctx, root, "HEAD") {
		if declared.Equal(pub) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > 1 {
		a.Problem = "your key " + a.Fingerprint + " is declared under more than one name (" + strings.Join(names, ", ") + "); keep one"
		return a
	}
	if len(names) == 0 {
		a.Problem = "your key " + a.Fingerprint + " is not declared in this repository: run `proofbound keys init --name <you>` (or copy your .pub file) and commit it under .proofbound/keys/"
		return a
	}
	a.Name, a.key = names[0], key
	return a
}

// signRecord returns the canonical record bytes with its signature.
func signRecord(fields map[string]any, key ed25519.PrivateKey) ([]byte, error) {
	schema, _ := fields["schema"].(string)
	unsigned, err := canonicalJSON(fields)
	if err != nil {
		return nil, err
	}
	signed := map[string]any{}
	for k, v := range fields {
		signed[k] = v
	}
	signed["signature"] = hex.EncodeToString(ed25519.Sign(key, append(recordTag(schema), unsigned...)))
	out, err := canonicalJSON(signed)
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func canonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return core.Canonicalize(raw)
}

// verifyRecord checks a record's bytes against the keys HEAD declares: exact canonical form, the
// expected schema, only that schema's fields, and an Ed25519 signature over the tag and the unsigned
// canonical bytes by the key declared for `by`, whose fingerprint is `key`.
func verifyRecord(data []byte, schema string, keys map[string]ed25519.PublicKey) (map[string]string, error) {
	var raw map[string]any
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&raw); err != nil {
		return nil, errors.New("not JSON")
	}
	fields := map[string]string{}
	for k, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("field %s is not a string", k)
		}
		fields[k] = s
	}
	if fields["schema"] != schema {
		return nil, fmt.Errorf("schema is %q, want %q", fields["schema"], schema)
	}
	allowed := map[string]bool{"signature": true}
	for _, f := range recordFields[schema] {
		allowed[f] = true
	}
	for k := range fields {
		if !allowed[k] {
			return nil, fmt.Errorf("unexpected field %s", k)
		}
	}
	canonical, err := canonicalJSON(raw)
	if err != nil || !bytes.Equal(append(canonical, '\n'), data) {
		return nil, errors.New("not in canonical form")
	}
	sig, err := hex.DecodeString(fields["signature"])
	if err != nil || hex.EncodeToString(sig) != fields["signature"] || len(sig) != ed25519.SignatureSize {
		return nil, errors.New("malformed signature")
	}
	pub, ok := keys[fields["by"]]
	if !ok {
		return nil, fmt.Errorf("signer %q has no key declared at the range's base", fields["by"])
	}
	if pbwarranty.Fingerprint(pub) != fields["key"] {
		return nil, fmt.Errorf("key %s is not the key declared for %s", fields["key"], fields["by"])
	}
	unsignedMap := map[string]any{}
	for k, v := range raw {
		if k != "signature" {
			unsignedMap[k] = v
		}
	}
	unsigned, err := canonicalJSON(unsignedMap)
	if err != nil {
		return nil, err
	}
	if !ed25519.Verify(pub, append(recordTag(schema), unsigned...), sig) {
		return nil, errors.New("signature does not verify")
	}
	return fields, nil
}

// headRecords are the files under dir in HEAD's tree, by path. The working tree is never read.
func headRecords(ctx context.Context, root, dir string) map[string][]byte {
	out := map[string][]byte{}
	for _, p := range strings.Split(gitOutput(ctx, root, "ls-tree", "--name-only", "HEAD", "--", dir+"/"), "\n") {
		p = strings.TrimSpace(p)
		if !strings.HasSuffix(p, ".json") || path.Dir(p) != dir {
			continue
		}
		if raw, err := gitBytes(ctx, root, "show", "HEAD:"+p); err == nil {
			out[p] = raw
		}
	}
	return out
}

// signerKeys are the keys a record about base..HEAD may be signed with: those already declared at
// base, so a key added inside the range cannot close the range's own exceptions (spec review F2).
func signerKeys(ctx context.Context, root string, rep auditReport) map[string]ed25519.PublicKey {
	return declaredKeys(ctx, root, rep.Base)
}

// authorKey is the key name a commit's author would have (`keys init` default: git user.name).
func authorKey(ctx context.Context, root, sha string) string {
	return pbwarranty.KeyName(gitOutput(ctx, root, "log", "-1", "--format=%an", sha))
}

// fixedByOK reports whether fix descends from commit and is in HEAD's history.
func fixedByOK(ctx context.Context, root, commit, fix string) bool {
	if commit == fix {
		return false
	}
	if !sha40RE.MatchString(commit) || !sha40RE.MatchString(fix) {
		return false
	}
	_, err1 := gitBytes(ctx, root, "merge-base", "--is-ancestor", commit, fix)
	_, err2 := gitBytes(ctx, root, "merge-base", "--is-ancestor", fix, "HEAD")
	return err1 == nil && err2 == nil
}

// attachOutcomes sets each exception's outcome from HEAD's committed records and recomputes the
// headline. A record that does not verify is shown, and closes nothing.
func attachOutcomes(ctx context.Context, root string, rep *auditReport) {
	keys := signerKeys(ctx, root, *rep)
	records := headRecords(ctx, root, outcomesDir)
	for i := range rep.Exceptions {
		e := &rep.Exceptions[i]
		p := outcomesDir + "/" + e.ID + ".json"
		data, ok := records[p]
		if !ok {
			continue
		}
		view := &outcomeView{Path: p}
		fields, err := verifyRecord(data, outcomeSchema, keys)
		switch {
		case err != nil:
		case fields["exception"] != e.ID || fields["commit"] != e.Commit:
			err = errors.New("the record names a different exception")
		case fields["by"] == authorKey(ctx, root, e.Commit):
			err = errors.New("signed by the commit's own author")
		case (fields["outcome"] == "fixed") != (fields["fixed_by"] != ""):
			err = errors.New("fixed_by must be present exactly when the outcome is fixed")
		case fields["outcome"] == "fixed" && (!sha40RE.MatchString(fields["fixed_by"]) || !fixedByOK(ctx, root, e.Commit, fields["fixed_by"])):
			err = errors.New("fixed_by is not a later commit in HEAD's history")
		case fields["outcome"] != "accepted" && fields["outcome"] != "fixed" && fields["outcome"] != "reverted":
			err = errors.New("unknown outcome")
		}
		if err != nil {
			view.Problem = "does not verify: " + err.Error()
		} else {
			view.Verified = true
			view.Outcome, view.Reason, view.FixedBy, view.By, view.At = fields["outcome"], printableUnit(fields["reason"]), fields["fixed_by"], fields["by"], fields["at"]
		}
		e.Outcome = view
	}
	rep.Headline = auditHeadline(*rep)
}

// readAttestations verifies HEAD's attestations and compares each with the transcript now.
func readAttestations(ctx context.Context, root string, rep auditReport) []attestationView {
	keys := signerKeys(ctx, root, rep)
	views := []attestationView{}
	records := headRecords(ctx, root, attestationsDir)
	paths := make([]string, 0, len(records))
	for p := range records {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		v := attestationView{Path: p}
		fields, err := verifyRecord(records[p], attestationSchema, keys)
		if err != nil {
			v.Problem = "does not verify: " + err.Error()
			views = append(views, v)
			continue
		}
		if !sha40RE.MatchString(fields["base"]) || !sha40RE.MatchString(fields["head"]) || !hex64RE.MatchString(fields["transcript_sha256"]) {
			v.Problem = "does not verify: malformed base, head or digest"
			views = append(views, v)
			continue
		}
		v.Verified = true
		v.Base, v.Head, v.Headline, v.By, v.At = fields["base"], fields["head"], printableUnit(fields["headline"]), fields["by"], fields["at"]
		for _, sha := range strings.Fields(gitOutput(ctx, root, "rev-list", "--end-of-options", v.Head+"..HEAD")) {
			v.Moved++
			if recordsOnly(ctx, root, sha) {
				v.RecordsOnly++
			}
		}
		views = append(views, v)
	}
	return views
}

// recordsOnly reports whether a commit only adds files under the two record directories.
func recordsOnly(ctx context.Context, root, sha string) bool {
	out := gitOutput(ctx, root, "diff-tree", "-r", "--no-commit-id", "--name-status", "--root", sha)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		status, p, ok := strings.Cut(line, "\t")
		if !ok || status != "A" || (!strings.HasPrefix(p, outcomesDir+"/") && !strings.HasPrefix(p, attestationsDir+"/")) {
			return false
		}
	}
	return len(lines) > 0
}

// recordDigest is the sha256 of a record's unsigned canonical bytes: preview shows it, sign must
// carry it back, so the record signed is exactly the record previewed (code review F1).
func recordDigest(fields map[string]any) (string, error) {
	unsigned, err := canonicalJSON(fields)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(unsigned)
	return hex.EncodeToString(sum[:]), nil
}

// actRequest is the body of /act/preview and /act/sign.
type actRequest struct {
	Kind      string `json:"kind"` // "outcome" or "attestation"
	Base      string `json:"base"`
	Exception string `json:"exception,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
	Reason    string `json:"reason,omitempty"`
	FixedBy   string `json:"fixed_by,omitempty"`
	// At is the time the preview showed; signing uses it, so the signed record is the one previewed.
	At string `json:"at,omitempty"`
	// Digest is the previewed record's recordDigest; sign refuses any other record.
	Digest string `json:"digest,omitempty"`
}

// signTime accepts the previewed time when it is well formed and from the last 15 minutes.
func signTime(at string, now time.Time) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, at)
	if err != nil || t.UTC().Format(time.RFC3339) != at {
		return time.Time{}, errors.New("preview the record before signing it")
	}
	if t.After(now.Add(time.Minute)) || now.Sub(t) > 15*time.Minute {
		return time.Time{}, errors.New("the preview is too old; review what will be written again")
	}
	return t, nil
}

// buildAct validates a request against a fresh audit and returns the record's path and unsigned
// fields. It never writes.
func buildAct(ctx context.Context, root string, rep auditReport, now []byte, a actor, req actRequest, at time.Time) (string, map[string]any, error) {
	if a.key == nil {
		return "", nil, errors.New(a.Problem)
	}
	if base, ok := signerKeys(ctx, root, rep)[a.Name]; !ok || !base.Equal(a.key.Public().(ed25519.PublicKey)) {
		return "", nil, fmt.Errorf("your key %s was not declared when this range began (%s), so it cannot sign for it", a.Fingerprint, rep.Base[:12])
	}
	fields := map[string]any{"by": a.Name, "key": a.Fingerprint, "at": at.UTC().Format(time.RFC3339)}
	switch req.Kind {
	case "outcome":
		if !exceptionIDRE.MatchString(req.Exception) {
			return "", nil, errors.New("not an exception id")
		}
		var found *auditException
		for i := range rep.Exceptions {
			if rep.Exceptions[i].ID == req.Exception {
				found = &rep.Exceptions[i]
			}
		}
		if found == nil {
			return "", nil, fmt.Errorf("%s is not an exception of this range now", req.Exception)
		}
		if a.Name == authorKey(ctx, root, found.Commit) {
			return "", nil, errors.New("you wrote this commit; someone else must record its outcome")
		}
		switch req.Outcome {
		case "accepted", "reverted":
			if req.FixedBy != "" {
				return "", nil, errors.New("fixed_by is only for a fixed outcome")
			}
		case "fixed":
			sha, err := resolveCommit(ctx, root, req.FixedBy)
			if err != nil || req.FixedBy == "" || !fixedByOK(ctx, root, found.Commit, sha) {
				return "", nil, errors.New("fixed needs the later commit, in HEAD's history, that fixed it")
			}
			fields["fixed_by"] = sha
		default:
			return "", nil, errors.New("outcome must be accepted, fixed or reverted")
		}
		if err := checkReason(req.Reason); err != nil {
			return "", nil, err
		}
		fields["schema"], fields["exception"], fields["commit"], fields["outcome"], fields["reason"] = outcomeSchema, found.ID, found.Commit, req.Outcome, req.Reason
		return outcomesDir + "/" + found.ID + ".json", fields, nil
	case "attestation":
		sum := sha256.Sum256(now)
		fields["schema"], fields["base"], fields["head"], fields["transcript_sha256"], fields["headline"] = attestationSchema, rep.Base, rep.Head, hex.EncodeToString(sum[:]), rep.Headline
		return attestationsDir + "/" + rep.Base[:12] + "-" + rep.Head[:12] + ".json", fields, nil
	}
	return "", nil, errors.New("kind must be outcome or attestation")
}

func checkReason(reason string) error {
	n := len([]rune(reason))
	if strings.TrimSpace(reason) == "" || n > 500 {
		return errors.New("the reason must be 1 to 500 characters")
	}
	for _, r := range reason {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return errors.New("the reason may not hold control or invisible formatting characters")
		}
	}
	return nil
}

// writeAndCommit writes a new record (never replacing one, in the working tree or at HEAD) and
// commits only that path. A failed commit leaves the file and is reported, not undone.
func writeAndCommit(ctx context.Context, root, rel string, data []byte, message string) (bool, string, error) {
	if _, err := gitBytes(ctx, root, "cat-file", "-e", "HEAD:"+rel); err == nil {
		return false, "", fmt.Errorf("%s is already committed; a record is never replaced", rel)
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return false, "", err
	}
	if err := writeExclusive(full, data); err != nil {
		return false, "", fmt.Errorf("%s exists; a record is never replaced", rel)
	}
	// A failed commit removes the file again, so nothing half-done blocks a retry (spec review F10).
	undo := func(out string) (bool, string, error) {
		_, _ = gitRun(ctx, root, "rm", "-q", "--cached", "--ignore-unmatch", "--", rel)
		_ = os.Remove(full)
		return false, "", fmt.Errorf("not recorded: git could not commit it: %s", out)
	}
	if out, err := gitRun(ctx, root, "add", "-f", "--", rel); err != nil {
		return undo(out)
	}
	if out, err := gitRun(ctx, root, "commit", "-q", "-m", message, "--", rel); err != nil {
		return undo(out)
	}
	return true, "committed", nil
}

func gitBytes(ctx context.Context, root string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	cmd.Env = repositoryGitEnv()
	return cmd.Output()
}

// gitRun runs git for a write, with the user's hooks, and returns its combined output.
func gitRun(ctx context.Context, root string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	cmd.Env = repositoryGitEnv()
	out, err := cmd.CombinedOutput()
	return printableUnit(strings.TrimSpace(string(out))), err
}
