package warranty

import (
	"bytes"
	"crypto/ed25519"
	"strings"
	"testing"
)

// W-INV-13: the verify page's answers are this package's, for the same files; the key is never taken
// from the file.
func TestPageCheck(t *testing.T) {
	key := testKey(t)
	fp := Fingerprint(key.Public().(ed25519.PublicKey))
	good, err := Sign(sampleBody(), key)
	if err != nil {
		t.Fatal(err)
	}
	if r := PageCheck(good, fp); r.State != "SIGNATURE ONLY" || r.Commit != testCommit || r.Key != fp || r.Claims != 2 {
		t.Fatalf("a good warranty: %+v", r)
	}
	if r := PageCheck(good, " "+strings.ToUpper(fp)+" "); r.State != "SIGNATURE ONLY" {
		t.Errorf("a pasted key with spaces and capitals: %+v", r)
	}
	other := Fingerprint(ed25519.NewKeyFromSeed(bytes.Repeat([]byte{1}, ed25519.SeedSize)).Public().(ed25519.PublicKey))
	if r := PageCheck(good, other); r.State != "INVALID" || !strings.Contains(r.Reason, "fingerprint mismatch") {
		t.Errorf("another key: %+v", r)
	}
	if r := PageCheck(good, ""); r.State != "INVALID" || !strings.Contains(r.Reason, "type the trust key") || r.Commit != "" {
		t.Errorf("no key typed: %+v", r)
	}
	if r := PageCheck(bytes.Replace(good, []byte(`"add-x"`), []byte(`"add-y"`), 1), fp); r.State != "INVALID" {
		t.Errorf("an altered byte: %+v", r)
	}
	if r := PageCheck([]byte("{}"), fp); r.State != "INVALID" {
		t.Errorf("not a warranty: %+v", r)
	}
	b := sampleBody()
	b.Verdicts[0].Status = "NEEDS_WORK"
	bad, err := Sign(b, key)
	if err != nil {
		t.Fatal(err)
	}
	if r := PageCheck(bad, fp); r.State != "INVALID" {
		t.Errorf("a re-signed NEEDS_WORK verdict: %+v", r)
	}
}
