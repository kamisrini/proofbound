package warranty

import (
	"context"
	"crypto/ed25519"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

const (
	testCommit = "0123456789abcdef0123456789abcdef01234567"
	reqDigest  = "1111111111111111111111111111111111111111111111111111111111111111"
	ciDigest   = "2222222222222222222222222222222222222222222222222222222222222222"
)

func testKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	return ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
}

// sampleBody is a minimal warranty that passes the rule: one requirement with two obligations,
// both SATISFIED and VERIFIABLE by a reviewer who is not the owner.
func sampleBody() Body {
	req := connectorintent.Reference{RecordKind: "requirement", Source: "intent.sdd", RecordID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest, Relation: "implements", ObligationIDs: []string{"O-A", "O-B"}}
	return Body{
		Schema: Schema, Commit: testCommit, ChangeName: "add-x", IssuedAt: "2026-11-19T00:00:00Z", Issuer: "owner",
		ChangeIntent: connectorintent.ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: "CI-add-x-bbbbbb", Status: "proposed", Targets: []connectorintent.Reference{req}, ArtifactSHA256: ciDigest},
		Requirements: []connectorintent.Requirement{{Schema: "proofbound.requirement.v1", RequirementID: "BR-x-aaaaaa", DeclaredOwner: "owner", ArtifactSHA256: reqDigest, Obligations: []connectorintent.Obligation{{ID: "O-A", Statement: "A."}, {ID: "O-B", Statement: "B."}}}},
		Reviews: []reviews.RequirementReview{{ReviewID: "RR-1", Requirement: connectorintent.Reference{RecordID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest}, DeclaredReviewer: "reviewer",
			Outcomes: []reviews.RequirementReviewOutcome{{ObligationID: "O-A", Outcome: "VERIFIABLE"}, {ObligationID: "O-B", Outcome: "VERIFIABLE"}}}},
		Verdicts: []reviews.ObligationVerdict{{VerdictID: "OV-1", Status: "ACCEPTABLE", DeclaredReviewer: "reviewer", ReviewedCommit: testCommit,
			ChangeIntent: connectorintent.Reference{RecordID: "CI-add-x-bbbbbb", ArtifactSHA256: ciDigest},
			Obligations:  []reviews.ObligationOutcome{{RequirementID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest, ObligationID: "O-A", Outcome: "SATISFIED"}, {RequirementID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest, ObligationID: "O-B", Outcome: "SATISFIED"}}}},
	}
}

func signParse(t *testing.T, b Body) File {
	t.Helper()
	data, err := Sign(b, testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	f, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// W-INV-1: a signed file verifies; any change to the body or the signature does not; duplicate
// keys, unknown fields and non-canonical bodies are malformed.
func TestSignatureCoversEveryByte(t *testing.T) {
	data, err := Sign(sampleBody(), testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	f, err := Parse(data)
	if err != nil || f.CheckSignature() != nil {
		t.Fatalf("calibration: a freshly signed file does not verify: %v", err)
	}
	for name, edit := range map[string]func(string) string{
		"statement": func(s string) string { return strings.Replace(s, `"A."`, `"A!"`, 1) },
		"outcome":   func(s string) string { return strings.Replace(s, `"SATISFIED"`, `"INCONCLUSIVE"`, 1) },
		"signature": func(s string) string { i := strings.Index(s, `"sig": "`) + 8; return s[:i] + flip(s[i:i+1]) + s[i+1:] },
	} {
		f, err := Parse([]byte(edit(string(data))))
		if err == nil {
			err = f.CheckSignature()
		}
		if err == nil {
			t.Errorf("%s: an edited file still verifies", name)
		}
	}
	for name, bad := range map[string]string{
		"duplicate key": strings.Replace(string(data), `"body": {`, `"body": {"schema": "x",`, 1),
		"unknown field": strings.Replace(string(data), `"body": {`, `"body": {"extra": 1,`, 1),
		"outer field":   strings.Replace(string(data), `"signature": {`, `"note": 1, "signature": {`, 1),
	} {
		if _, err := Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), "malformed") {
			t.Errorf("%s: parsed: %v", name, err)
		}
	}
}

func flip(c string) string {
	if c == "0" {
		return "1"
	}
	return "0"
}

// W-INV-3: the readiness rule is re-run on the file. A body re-signed by the same key is still
// INVALID when any outcome is not SATISFIED/VERIFIABLE, a target lacks an outcome, the reviewer is
// the declared owner, a verdict is for another commit or change revision, or a requirement is
// missing or untargeted.
func TestRuleRejectsReSignedForgeries(t *testing.T) {
	if err := CheckRule(signParse(t, sampleBody()).Body); err != nil {
		t.Fatalf("calibration: the sample body fails the rule: %v", err)
	}
	cases := map[string]func(*Body){
		"inconclusive":     func(b *Body) { b.Verdicts[0].Obligations[1].Outcome = "INCONCLUSIVE" },
		"ambiguous":        func(b *Body) { b.Reviews[0].Outcomes[0].Outcome = "AMBIGUOUS" },
		"missing outcome":  func(b *Body) { b.Verdicts[0].Obligations = b.Verdicts[0].Obligations[:1] },
		"no review":        func(b *Body) { b.Reviews = nil },
		"owner reviews":    func(b *Body) { b.Reviews[0].DeclaredReviewer = "owner" },
		"owner verdict":    func(b *Body) { b.Verdicts[0].DeclaredReviewer = "owner" },
		"other commit":     func(b *Body) { b.Verdicts[0].ReviewedCommit = strings.Repeat("f", 40) },
		"other change rev": func(b *Body) { b.Verdicts[0].ChangeIntent.ArtifactSHA256 = reqDigest },
		"missing req":      func(b *Body) { b.Requirements = nil },
		"untargeted req": func(b *Body) {
			b.Requirements = append(b.Requirements, connectorintent.Requirement{RequirementID: "BR-y-cccccc"})
		},
		"obligation not req": func(b *Body) { b.ChangeIntent.Targets[0].ObligationIDs = []string{"O-A", "O-B", "O-C"} },
		"bad commit":         func(b *Body) { b.Commit = "HEAD" },
		// W-INV-3 amended (needs-work-blocks-spec.md): every obligation SATISFIED, but the reviewer
		// recorded NEEDS_WORK (a rejected out-of-area line or uncaught part).
		"needs work": func(b *Body) { b.Verdicts[0].Status = "NEEDS_WORK" },
	}
	for name, mutate := range cases {
		b := sampleBody()
		mutate(&b)
		if err := CheckRule(signParse(t, b).Body); err == nil {
			t.Errorf("%s: a re-signed forgery passed the rule", name)
		}
	}
}

// W-INV-4: without a repository, a good file is UNVERIFIED, never VALID; a pinned fingerprint is
// enforced even then.
func TestNoRepositoryIsNeverValid(t *testing.T) {
	f := signParse(t, sampleBody())
	if err := Verify(context.Background(), f, "", ""); !errors.Is(err, ErrUnverified) {
		t.Fatalf("without --repo: %v, want ErrUnverified", err)
	}
	pub, _ := f.PublicKey()
	if err := Verify(context.Background(), f, "", Fingerprint(pub)); !errors.Is(err, ErrUnverified) {
		t.Fatalf("matching pin: %v", err)
	}
	if err := Verify(context.Background(), f, "", "0000000000000000"); err == nil || errors.Is(err, ErrUnverified) {
		t.Fatalf("wrong pin: %v", err)
	}
}

// W-INV-7: the package's own imports exclude the store, pgx and the network packages.
func TestNoDatabaseOrNetworkImports(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range parsed.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasSuffix(path, "/internal/store") || strings.Contains(path, "pgx") || path == "net" || strings.HasPrefix(path, "net/") {
				t.Errorf("%s imports %s", name, path)
			}
		}
	}
}

// W-INV-8: the private key is never created inside a git work tree, and a key file that is a
// symlink or readable by others is refused.
func TestKeyFileSafety(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	if _, err := InitKey(filepath.Join(repo, "keys", "signing-key")); err == nil {
		t.Fatal("a key was created inside a git work tree")
	}
	path := filepath.Join(t.TempDir(), "cfg", "signing-key")
	if _, err := InitKey(path); err != nil {
		t.Fatal(err)
	}
	if _, err := InitKey(path); err == nil {
		t.Fatal("an existing key was replaced")
	}
	if _, err := LoadKey(path); err != nil {
		t.Fatalf("calibration: %v", err)
	}
	link := filepath.Join(filepath.Dir(path), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKey(link); err == nil {
		t.Fatal("a symlinked key was loaded")
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKey(path); err == nil {
		t.Fatal("a world-readable key was loaded")
	}
}

// Mutation-sweep kills for Parse and the signature (W-INV-1): a body that decodes but does not
// round-trip to its canonical form is malformed, and a key or signature with trailing junk after
// valid hex is refused, not decoded partially.
func TestParseAndSignatureRefuseNearMisses(t *testing.T) {
	data, err := Sign(sampleBody(), testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	nonCanonical := strings.Replace(s, `"relation": ""`, `"obligation_ids": [], "relation": ""`, 1)
	if nonCanonical == s {
		t.Fatal("calibration: no empty relation to extend")
	}
	if _, err := Parse([]byte(nonCanonical)); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("a non-round-tripping body parsed: %v", err)
	}
	for _, field := range []string{`"key": "`, `"sig": "`} {
		i := strings.Index(s, field) + len(field)
		j := i + strings.Index(s[i:], `"`)
		f, err := Parse([]byte(s[:j] + "zz" + s[j:]))
		if err != nil {
			t.Fatal(err)
		}
		if f.CheckSignature() == nil {
			t.Errorf("%s with trailing junk verified", field)
		}
	}
}

// Mutation-sweep kills for keys (W-INV-8): KeyPath honours XDG_CONFIG_HOME and falls back to
// $HOME/.config; a loaded key is the key that was created; a corrupt or short seed is refused; a
// path inside .git (not a work tree) is allowed.
func TestKeyPathAndLoadAreExact(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/x/cfg")
	if p, err := KeyPath(); err != nil || p != "/x/cfg/proofbound/signing-key" {
		t.Fatalf("XDG path: %q %v", p, err)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/x/home")
	if p, err := KeyPath(); err != nil || p != "/x/home/.config/proofbound/signing-key" {
		t.Fatalf("HOME path: %q %v", p, err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "k")
	pub, err := InitKey(path)
	if err != nil {
		t.Fatal(err)
	}
	key, err := LoadKey(path)
	if err != nil || !pub.Equal(key.Public()) {
		t.Fatalf("loaded a different key: %v", err)
	}
	seed := strings.Repeat("ab", ed25519.SeedSize)
	for name, content := range map[string]string{"trailing junk": seed + "zz", "short": seed[:60], "not hex": "zz"} {
		bad := filepath.Join(dir, strings.ReplaceAll(name, " ", "-"))
		if err := os.WriteFile(bad, []byte(content+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadKey(bad); err == nil {
			t.Errorf("%s seed loaded", name)
		}
	}
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	if _, err := InitKey(filepath.Join(repo, ".git", "proofbound", "k")); err != nil {
		t.Fatalf("a key inside .git (not a work tree) was refused: %v", err)
	}
}
