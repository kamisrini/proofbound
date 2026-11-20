package warranty

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

// MVP step 3 (docs/plans/mvp-3-reviewer-signed-spec.md): reviewer proof by A (`.sig` files) and B
// (cosignatures), W-INV-3 amended and W-INV-9..12.

func seedKey(b byte) ed25519.PrivateKey { return ed25519.NewKeyFromSeed(append(make([]byte, 31), b)) }

func pubOf(k ed25519.PrivateKey) ed25519.PublicKey { return k.Public().(ed25519.PublicKey) }

func (x *fixture) commitAll(msg string) {
	x.t.Helper()
	x.git("add", "-A")
	x.git("commit", "-q", "--allow-empty", "-m", msg)
}

func (x *fixture) rm(rel string) { x.t.Helper(); x.git("rm", "-q", rel) }

func (x *fixture) read(rel string) []byte {
	x.t.Helper()
	data, err := os.ReadFile(filepath.Join(x.root, filepath.FromSlash(rel)))
	if err != nil {
		x.t.Fatal(err)
	}
	return data
}

func (x *fixture) verifyBytes(data []byte) error {
	x.t.Helper()
	f, err := Parse(data)
	if err != nil {
		return err
	}
	return Verify(context.Background(), f, x.root, "")
}

func (x *fixture) signed() []byte {
	x.t.Helper()
	data, err := Sign(x.body, x.key)
	if err != nil {
		x.t.Fatal(err)
	}
	return data
}

func wantUnreviewed(t *testing.T, label string, err error) {
	t.Helper()
	if !errors.Is(err, ErrUnreviewed) {
		t.Errorf("%s: %v, want UNREVIEWED", label, err)
	}
}

func wantInvalid(t *testing.T, label string, err error, substr string) {
	t.Helper()
	if err == nil || errors.Is(err, ErrUnreviewed) || errors.Is(err, ErrUnverified) || !strings.Contains(err.Error(), substr) {
		t.Errorf("%s: %v, want INVALID containing %q", label, err, substr)
	}
}

// W-INV-9 / W-INV-9b: every artifact signed by the reviewer's declared key is VALID and names the
// reviewer; any unsigned artifact, or no declared key and no signatures, is UNREVIEWED; INVALID wins.
func TestReviewerProofByArtifactSignatures(t *testing.T) {
	x := newFixture(t)
	f := signParse(t, x.body)
	proven, err := VerifyReport(context.Background(), f, x.root, "")
	if err != nil || len(proven) != 1 || proven[0].Name != "reviewer" || proven[0].Fingerprint != Fingerprint(pubOf(x.reviewer)) {
		t.Fatalf("calibration: %v %+v", err, proven)
	}
	head := x.git("rev-parse", "HEAD")
	reset := func() { x.git("reset", "-q", "--hard", head) }

	x.rm(x.body.Verdicts[0].ArtifactPath + ".sig")
	x.commitAll("one artifact unsigned")
	wantUnreviewed(t, "partly signed", x.verify(x.body))
	reset()

	x.rm(x.body.Verdicts[0].ArtifactPath + ".sig")
	x.rm(x.body.Reviews[0].ArtifactPath + ".sig")
	x.rm(KeysDir + "/reviewer.pub")
	x.commitAll("no key, no signatures")
	wantUnreviewed(t, "no key and no signatures", x.verify(x.body))
	reset()

	// INVALID wins: one artifact unsigned (unproven) and the other signed by a stranger.
	x.rm(x.body.Verdicts[0].ArtifactPath + ".sig")
	rr := x.body.Reviews[0].ArtifactPath
	x.write(rr+".sig", SignArtifact(x.read(rr), seedKey(9)))
	x.commitAll("one unsigned, one by a stranger")
	wantInvalid(t, "invalid beats unreviewed", x.verify(x.body), "not by the declared key")
	reset()
}

// W-INV-9a: a present `.sig` that is wrong is never ignored.
func TestBadArtifactSignaturesAreInvalid(t *testing.T) {
	x := newFixture(t)
	head := x.git("rev-parse", "HEAD")
	ov := x.body.Verdicts[0].ArtifactPath
	cases := map[string]struct {
		change func()
		want   string
	}{
		"another key":            {func() { x.write(ov+".sig", SignArtifact(x.read(ov), seedKey(9))) }, "not by the declared key"},
		"over other bytes":       {func() { x.write(ov+".sig", SignArtifact([]byte("other"), x.reviewer)) }, "does not verify"},
		"malformed":              {func() { x.write(ov+".sig", "ed25519 zz\n") }, "malformed"},
		"no trailing newline":    {func() { x.write(ov+".sig", strings.TrimSuffix(SignArtifact(x.read(ov), x.reviewer), "\n")) }, "malformed"},
		"signed, key removed":    {func() { x.rm(KeysDir + "/reviewer.pub") }, "not by the declared key"},
		"declared key not exact": {func() { x.write(KeysDir+"/reviewer.pub", " "+PubLine(pubOf(x.reviewer))) }, "not an ed25519 key line"},
		"sig is a symlink": {func() {
			x.rm(ov + ".sig")
			if err := os.Symlink("OV-elsewhere.md.sig", filepath.Join(x.root, filepath.FromSlash(ov+".sig"))); err != nil {
				t.Fatal(err)
			}
		}, "not a regular committed file"},
	}
	for name, c := range cases {
		c.change()
		x.commitAll(name)
		wantInvalid(t, name, x.verify(x.body), c.want)
		x.git("reset", "-q", "--hard", head)
		x.git("clean", "-qfdx")
	}
}

// W-INV-10: one key may not speak as reviewer and as exporter, owner, or another reviewer.
func TestReviewerKeysMustBeDistinct(t *testing.T) {
	x := newFixture(t)
	head := x.git("rev-parse", "HEAD")

	x.write(KeysDir+"/reviewer.pub", PubLine(pubOf(x.key)))
	x.commitAll("reviewer declared with the exporter's key")
	wantInvalid(t, "exporter key", x.verify(x.body), "exporter's key")
	x.git("reset", "-q", "--hard", head)

	x.write(KeysDir+"/spec-owner.pub", PubLine(pubOf(x.reviewer)))
	x.commitAll("owner declared with the reviewer's key")
	wantInvalid(t, "owner key", x.verify(x.body), "declared owner")
	x.git("reset", "-q", "--hard", head)

	// A second reviewer ("second") on the verdict, declared with the first reviewer's key.
	ov := x.body.Verdicts[0].ArtifactPath
	v := x.body.Verdicts[0]
	v.DeclaredReviewer = "second"
	x.write(ov, artifact(t, func(d string) any { v.ArtifactSHA256 = d; return v }))
	parsed, err := reviews.ParseObligationVerdict(ov, x.read(ov))
	if err != nil {
		t.Fatal(err)
	}
	b := x.body
	b.Verdicts = []reviews.ObligationVerdict{parsed}
	x.write(ov+".sig", SignArtifact(x.read(ov), x.reviewer))
	x.write(KeysDir+"/second.pub", PubLine(pubOf(x.reviewer)))
	x.commitAll("two reviewers, one key")
	wantInvalid(t, "shared key", x.verify(b), "share one declared key")

	x.write(KeysDir+"/second.pub", PubLine(pubOf(seedKey(11))))
	x.write(ov+".sig", SignArtifact(x.read(ov), seedKey(11)))
	x.commitAll("two reviewers, two keys")
	if proven, err := VerifyReport(context.Background(), signParse(t, b), x.root, ""); err != nil || len(proven) != 2 {
		t.Fatalf("calibration: two distinct reviewers: %v %+v", err, proven)
	}
}

// W-INV-3 amended: identity is KeyName.
func TestRuleUsesKeyNames(t *testing.T) {
	x := newFixture(t)
	if err := CheckRule(x.body); err != nil {
		t.Fatalf("calibration: %v", err)
	}
	edit := func(f func(b *Body)) Body {
		raw, _ := json.Marshal(x.body)
		var b Body
		if err := json.Unmarshal(raw, &b); err != nil {
			t.Fatal(err)
		}
		f(&b)
		return b
	}
	cases := map[string]struct {
		body Body
		want string
	}{
		"reviewer spells the owner differently": {edit(func(b *Body) { b.Verdicts[0].DeclaredReviewer = "Spec Owner" }), "declared owner"},
		"review spells the owner differently":   {edit(func(b *Body) { b.Reviews[0].DeclaredReviewer = "SPEC-OWNER." }), "declared owner"},
		"empty reviewer key name":               {edit(func(b *Body) { b.Verdicts[0].DeclaredReviewer = "!!!" }), "no usable key name"},
		"empty owner key name":                  {edit(func(b *Body) { b.Requirements[0].DeclaredOwner = "李" }), "no usable key name"},
		"two spellings of one reviewer":         {edit(func(b *Body) { b.Verdicts[0].DeclaredReviewer = "Reviewer" }), "one key name"},
	}
	for name, c := range cases {
		if err := CheckRule(c.body); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", name, err, c.want)
		}
	}
	if KeyName("  Alice S. ") != "alice-s" || KeyName("李") != "" {
		t.Fatalf("KeyName: %q %q", KeyName("  Alice S. "), KeyName("李"))
	}
}

// W-INV-11: cosignatures parse only in their one exact form.
func TestCosignatureParsingIsStrict(t *testing.T) {
	x := newFixture(t)
	good, err := Cosign(x.signed(), x.reviewer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(good); err != nil {
		t.Fatalf("calibration: %v", err)
	}
	var outer map[string]any
	if err := json.Unmarshal(good, &outer); err != nil {
		t.Fatal(err)
	}
	c := outer["cosignatures"].([]any)[0].(map[string]any)
	key, sig := c["key"].(string), c["sig"].(string)
	other, _ := Cosign(x.signed(), seedKey(12))
	var o2 map[string]any
	_ = json.Unmarshal(other, &o2)
	c2 := o2["cosignatures"].([]any)[0].(map[string]any)
	lo, hi := c, c2
	if c2["key"].(string) < key {
		lo, hi = c2, c
	}
	many := []any{}
	for i := 0; i < 17; i++ {
		many = append(many, map[string]any{"key": fmt.Sprintf("%064x", i+1), "sig": sig}) // sorted, unique: only the cap fails
	}
	cases := map[string]any{
		"uppercase key": []any{map[string]any{"key": strings.ToUpper(key), "sig": sig}},
		"short sig":     []any{map[string]any{"key": key, "sig": sig[:126]}},
		"unknown field": []any{map[string]any{"key": key, "sig": sig, "who": "x"}},
		"empty list":    []any{},
		"null":          nil,
		"unsorted":      []any{hi, lo},
		"duplicate":     []any{c, c},
		"more than 16":  many,
	}
	for name, v := range cases {
		outer["cosignatures"] = v
		data, _ := json.Marshal(outer)
		if _, err := Parse(data); err == nil || !strings.Contains(err.Error(), "malformed") {
			t.Errorf("%s: parsed: %v", name, err)
		}
	}
	delete(outer, "cosignatures")
	data, _ := json.Marshal(outer)
	if _, err := Parse(data); err != nil {
		t.Fatalf("absent cosignatures: %v", err)
	}
}

// W-INV-12: a cosignature by a reviewer's declared key proves the reviewer (B); one by any other key,
// over another exporter's file, or by the exporter itself is refused or INVALID.
func TestCosignatureProvesAReviewer(t *testing.T) {
	x := newFixture(t)
	x.rm(x.body.Verdicts[0].ArtifactPath + ".sig")
	x.rm(x.body.Reviews[0].ArtifactPath + ".sig")
	x.commitAll("no A signatures")
	base := x.signed()
	wantUnreviewed(t, "calibration: unsigned", x.verifyBytes(base))

	cosigned, err := Cosign(base, x.reviewer)
	if err != nil {
		t.Fatal(err)
	}
	if err := x.verifyBytes(cosigned); err != nil {
		t.Fatalf("a reviewer cosignature did not prove the reviewer: %v", err)
	}
	if _, err := Cosign(cosigned, x.reviewer); err == nil || !strings.Contains(err.Error(), "already cosigned") {
		t.Errorf("repeat cosign: %v", err)
	}
	if _, err := Cosign(base, x.key); err == nil || !strings.Contains(err.Error(), "exporter's key") {
		t.Errorf("exporter cosign: %v", err)
	}
	bystander, err := Cosign(base, seedKey(13))
	if err != nil {
		t.Fatal(err)
	}
	wantInvalid(t, "bystander cosignature", x.verifyBytes(bystander), "not the declared key of a reviewer")

	// The reviewer's cosignature carried onto the same body signed by another exporter key.
	var outer map[string]any
	_ = json.Unmarshal(cosigned, &outer)
	otherExporter := seedKey(14)
	x.write(KeysDir+"/other.pub", PubLine(pubOf(otherExporter)))
	x.commitAll("another exporter")
	resigned, err := Sign(x.body, otherExporter)
	if err != nil {
		t.Fatal(err)
	}
	var r map[string]any
	_ = json.Unmarshal(resigned, &r)
	r["cosignatures"] = outer["cosignatures"]
	moved, _ := json.Marshal(r)
	wantInvalid(t, "cosignature moved to another exporter's file", x.verifyBytes(moved), "does not verify")

	// Without a repository a bad cosignature is still INVALID, and a good one is UNVERIFIED.
	f, _ := Parse(moved)
	if err := Verify(context.Background(), f, "", ""); err == nil || errors.Is(err, ErrUnverified) {
		t.Errorf("no repo, moved cosignature: %v", err)
	}
	g, _ := Parse(cosigned)
	if err := Verify(context.Background(), g, "", ""); !errors.Is(err, ErrUnverified) {
		t.Errorf("no repo, good cosignature: %v", err)
	}
}

// W-INV-9a, clause by clause (mutation sweep kills): a `.sig` line is accepted only in its one exact
// form, and a short key is an error, never a panic in ed25519.Verify.
func TestArtifactSignatureLineIsExact(t *testing.T) {
	data := []byte("artifact bytes")
	good := SignArtifact(data, seedKey(3))
	if _, err := checkArtifactSig(data, []byte(good)); err != nil {
		t.Fatalf("calibration: %v", err)
	}
	f := strings.Fields(good)
	cases := map[string]string{
		"no trailing newline": strings.TrimSuffix(good, "\n"),
		"extra field":         f[0] + " " + f[1] + " " + f[2] + " x\n",
		"other algorithm":     "ed448 " + f[1] + " " + f[2] + "\n",
		"uppercase key hex":   f[0] + " " + strings.ToUpper(f[1]) + " " + f[2] + "\n",
		"short key":           f[0] + " " + f[1][:62] + " " + f[2] + "\n",
		"non-hex key":         f[0] + " " + "zz" + f[1][2:] + " " + f[2] + "\n",
		"uppercase sig hex":   f[0] + " " + f[1] + " " + strings.ToUpper(f[2]) + "\n",
		"short sig":           f[0] + " " + f[1] + " " + f[2][:126] + "\n",
		"non-hex sig":         f[0] + " " + f[1] + " " + "zz" + f[2][2:] + "\n",
	}
	for name, line := range cases {
		if _, err := checkArtifactSig(data, []byte(line)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// W-INV-9/9a, clause by clause: a declared key must be exactly a 32-byte key line, and a `.sig` that is
// a directory is not a regular committed file.
func TestDeclaredKeyAndSigFileShapes(t *testing.T) {
	x := newFixture(t)
	head := x.git("rev-parse", "HEAD")
	short := PubLine(pubOf(x.reviewer))
	short = "ed25519 " + short[len("ed25519 "):len(short)-3] + "\n" // 31 bytes of hex-encoded key
	x.write(KeysDir+"/reviewer.pub", short)
	x.commitAll("a 31-byte key line")
	wantInvalid(t, "short declared key", x.verify(x.body), "not an ed25519 key line")
	x.git("reset", "-q", "--hard", head)

	ov := x.body.Verdicts[0].ArtifactPath
	x.rm(ov + ".sig")
	x.write(ov+".sig/inner", "x\n")
	x.commitAll("a .sig directory")
	wantInvalid(t, "sig directory", x.verify(x.body), "not a regular committed file")
}
