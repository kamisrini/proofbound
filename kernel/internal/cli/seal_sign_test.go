package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// CLI-INV-19/20 amended (MVP step 3, option A): `review seal` signs every artifact with the
// reviewer's declared key, writes unsigned (with a notice) when no key is declared, and refuses —
// writing nothing — when a key is declared that this machine does not hold or cannot load.
func TestReviewSealSignsWithTheDeclaredKey(t *testing.T) {
	short := sealCommit[:12]
	artifacts := []string{"docs/verification/verdicts/RR-" + short + "-1.md", "docs/verification/verdicts/OV-" + short + "-1.md"}
	declare := func(t *testing.T, root string, pub ed25519.PublicKey) {
		t.Helper()
		p := filepath.Join(root, ".proofbound", "keys", "proofbound-reviewer.pub")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(pbwarranty.PubLine(pub)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	localKey := func(t *testing.T) ed25519.PublicKey {
		t.Helper()
		path := filepath.Join(t.TempDir(), "reviewer-key")
		t.Setenv("PROOFBOUND_KEY", path)
		pub, err := pbwarranty.InitKey(path)
		if err != nil {
			t.Fatal(err)
		}
		return pub
	}
	seal := func(t *testing.T, root string) (string, error) {
		t.Helper()
		var out bytes.Buffer
		err := reviewSeal(context.Background(), root, writeDraft(t, root, sealDraft("SATISFIED", "VERIFIABLE")), ledgerTruth, &out)
		return out.String(), err
	}
	wroteNothing := func(t *testing.T, root, label string) {
		t.Helper()
		for _, rel := range append(artifacts, "gates/ready-"+short+".yaml", artifacts[0]+".sig") {
			if _, err := os.Lstat(filepath.Join(root, rel)); err == nil {
				t.Fatalf("%s: %s was written", label, rel)
			}
		}
	}

	t.Run("declared and held: signs", func(t *testing.T) {
		root := t.TempDir()
		pub := localKey(t)
		declare(t, root, pub)
		out, err := seal(t, root)
		if err != nil || !strings.Contains(out, "signed as proofbound-reviewer") {
			t.Fatalf("%v\n%s", err, out)
		}
		for _, rel := range artifacts {
			data, _ := os.ReadFile(filepath.Join(root, rel))
			sig, err := os.ReadFile(filepath.Join(root, rel+".sig"))
			if err != nil || string(sig) != pbwarranty.SignArtifact(data, mustLoad(t)) {
				t.Fatalf("%s.sig: %v %q", rel, err, sig)
			}
		}
	})
	t.Run("no key declared: unsigned with a notice", func(t *testing.T) {
		root := t.TempDir()
		localKey(t)
		out, err := seal(t, root)
		if err != nil || !strings.Contains(out, "unsigned: no key declared for proofbound-reviewer") {
			t.Fatalf("%v\n%s", err, out)
		}
		if _, err := os.Lstat(filepath.Join(root, artifacts[0]+".sig")); err == nil {
			t.Fatal("an unsigned seal wrote a .sig")
		}
	})
	t.Run("declared, another key here: refuses", func(t *testing.T) {
		root := t.TempDir()
		localKey(t)
		declare(t, root, ed25519.NewKeyFromSeed(make([]byte, 32)).Public().(ed25519.PublicKey))
		if _, err := seal(t, root); err == nil || !strings.Contains(err.Error(), "not the key declared") {
			t.Fatalf("mismatched key sealed: %v", err)
		}
		wroteNothing(t, root, "mismatch")
	})
	t.Run("declared, no key here: refuses", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("PROOFBOUND_KEY", filepath.Join(t.TempDir(), "missing"))
		declare(t, root, ed25519.NewKeyFromSeed(make([]byte, 32)).Public().(ed25519.PublicKey))
		if _, err := seal(t, root); err == nil || !strings.Contains(err.Error(), "cannot use one") {
			t.Fatalf("missing key sealed: %v", err)
		}
		wroteNothing(t, root, "missing key")
	})
	t.Run("an existing .sig: refuses", func(t *testing.T) {
		root := t.TempDir()
		declare(t, root, localKey(t))
		p := filepath.Join(root, filepath.FromSlash(artifacts[1]+".sig"))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := seal(t, root); err == nil || !strings.Contains(err.Error(), "exists") {
			t.Fatalf("sealed over an existing .sig: %v", err)
		}
		wroteNothing(t, root, "existing sig")
	})
}

func mustLoad(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	path, err := pbwarranty.KeyPath()
	if err != nil {
		t.Fatal(err)
	}
	k, err := pbwarranty.LoadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// CLI-INV-27 (in package): `warranty cosign` routes with an optional -o, and refuses a file that is
// missing or does not parse without writing anything.
func TestWarrantyCosignRefusesBadInput(t *testing.T) {
	for in, want := range map[string]command{
		"warranty cosign w.json":          commandWarrantyCosign,
		"warranty cosign w.json -o out":   commandWarrantyCosign,
		"warranty cosign":                 commandInvalid,
		"warranty cosign -o out":          commandInvalid,
		"warranty cosign w.json -o":       commandInvalid,
		"warranty cosign w.json --repo x": commandInvalid,
	} {
		if got := parseCommand(strings.Fields(in)); got != want {
			t.Errorf("%q: %v, want %v", in, got, want)
		}
	}
	root := t.TempDir()
	bad := filepath.Join(root, "w.json")
	if err := os.WriteFile(bad, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out.json")
	var buf bytes.Buffer
	for _, file := range []string{bad, filepath.Join(root, "missing.json")} {
		if err := warrantyCosign(context.Background(), root, file, out, &buf); err == nil || !strings.HasPrefix(err.Error(), "cosign:") {
			t.Fatalf("%s: %v", file, err)
		}
		if _, err := os.Lstat(out); err == nil {
			t.Fatalf("%s: a refused cosign wrote %s", file, out)
		}
	}
}
