package warranty

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
)

// Reviewer proof (MVP step 3, docs/plans/mvp-3-reviewer-signed-spec.md, W-INV-9..11): a reviewer is
// proven by A, a `.sig` beside each of their committed artifacts, or B, a cosignature on the
// warranty file; either must be made by the key the repository declares for the reviewer's name.

const (
	reviewSigTag = "proofbound.review-signature.v1\n"
	cosigTag     = "proofbound.warranty-cosignature.v1\n"
)

// ErrUnreviewed is returned by Verify when every check passes but some named reviewer has not signed.
// Like ErrUnverified it is never a pass.
var ErrUnreviewed = errors.New("unreviewed")

var keyNameRE = regexp.MustCompile(`[^a-z0-9-]+`)

// KeyName is the file name stem under KeysDir that declares name's key (the `keys init --name` slug).
func KeyName(name string) string {
	return strings.Trim(keyNameRE.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-"), "-")
}

// SignArtifact returns the `.sig` line for an artifact's bytes (option A).
func SignArtifact(data []byte, key ed25519.PrivateKey) string {
	pub := key.Public().(ed25519.PublicKey)
	sig := ed25519.Sign(key, append([]byte(reviewSigTag), data...))
	return "ed25519 " + hex.EncodeToString(pub) + " " + hex.EncodeToString(sig) + "\n"
}

// checkArtifactSig verifies a `.sig` line over data and returns its key. The line must be exact.
func checkArtifactSig(data, line []byte) (ed25519.PublicKey, error) {
	fields := strings.Split(strings.TrimSuffix(string(line), "\n"), " ")
	if !bytes.HasSuffix(line, []byte("\n")) || len(fields) != 3 || fields[0] != "ed25519" {
		return nil, errors.New("malformed")
	}
	pub, err := hex.DecodeString(fields[1])
	if err != nil || len(pub) != ed25519.PublicKeySize || hex.EncodeToString(pub) != fields[1] {
		return nil, errors.New("malformed key")
	}
	// Canonical lowercase hex only; a wrong length is left to ed25519.Verify, which rejects it (unlike a
	// wrong-length key, which would make Verify panic, hence the key's explicit length check).
	sig, _ := hex.DecodeString(fields[2])
	if hex.EncodeToString(sig) != fields[2] {
		return nil, errors.New("malformed signature")
	}
	if !ed25519.Verify(pub, append([]byte(reviewSigTag), data...), sig) {
		return nil, errors.New("does not verify")
	}
	return pub, nil
}

// cosigBytes binds the exporter's key as well as the body, so a cosignature cannot be carried onto the
// same body re-signed by another exporter.
func cosigBytes(exporterHex string, canonical []byte) []byte {
	return append([]byte(cosigTag+exporterHex+"\n"), canonical...)
}

// checkCosignatures verifies every cosignature over the body (no repository needed).
func (f File) checkCosignatures() error {
	for _, c := range f.Cosignatures {
		pub, _ := hex.DecodeString(c.Key)
		sig, _ := hex.DecodeString(c.Sig)
		if !ed25519.Verify(pub, cosigBytes(f.Signature.Key, f.canonical), sig) {
			return fmt.Errorf("cosignature by %s does not verify", Fingerprint(pub))
		}
	}
	return nil
}

// Cosign returns the warranty file bytes with one more cosignature by key over the body (option B).
// It refuses a key that already cosigned or that is the exporter's key.
func Cosign(data []byte, key ed25519.PrivateKey) ([]byte, error) {
	f, err := Parse(data)
	if err != nil {
		return nil, err
	}
	pub := hex.EncodeToString(key.Public().(ed25519.PublicKey))
	if pub == f.Signature.Key {
		return nil, errors.New("cosign: this is the exporter's key; a reviewer cosigns with their own key")
	}
	for _, c := range f.Cosignatures {
		if c.Key == pub {
			return nil, errors.New("cosign: this key already cosigned the warranty")
		}
	}
	cosigs := append(append([]Signature{}, f.Cosignatures...), Signature{Key: pub, Sig: hex.EncodeToString(ed25519.Sign(key, cosigBytes(f.Signature.Key, f.canonical)))})
	sort.Slice(cosigs, func(i, j int) bool { return cosigs[i].Key < cosigs[j].Key })
	return marshalFile(f.canonical, f.Signature, cosigs)
}

// reviewersOf is every distinct declared reviewer in the body, sorted.
func reviewersOf(b Body) []string {
	seen := map[string]bool{}
	for _, r := range b.Reviews {
		seen[r.DeclaredReviewer] = true
	}
	for _, v := range b.Verdicts {
		seen[v.DeclaredReviewer] = true
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// declaredKey is the key committed at HEAD as KeysDir/<KeyName(name)>.pub, or nil when none is.
func declaredKey(ctx context.Context, r *gitcmd.Repo, head, name string) (ed25519.PublicKey, error) {
	stem := KeyName(name)
	if stem == "" {
		return nil, nil
	}
	data, err := optionalBlobAtHead(ctx, r, head, KeysDir+"/"+stem+".pub")
	if err != nil || data == nil {
		return nil, err
	}
	// Exactness is one comparison: the bytes must be PubLine of a 32-byte key (a decode error leaves a
	// prefix whose PubLine cannot equal the data).
	pub, _ := hex.DecodeString(strings.TrimSuffix(strings.TrimPrefix(string(data), "ed25519 "), "\n"))
	if len(pub) != ed25519.PublicKeySize || PubLine(pub) != string(data) {
		return nil, fmt.Errorf("%s/%s.pub is not an ed25519 key line", KeysDir, stem)
	}
	return ed25519.PublicKey(pub), nil
}

// ProvenReviewer is a reviewer and the fingerprint of the declared key that proved them.
type ProvenReviewer struct{ Name, Fingerprint string }

// reviewersProven is check 9: every named reviewer is proven by A or B with their declared key, which
// is neither the exporter's key nor a requirement owner's declared key. A present signature that
// fails is INVALID; a reviewer with no proof makes the result ErrUnreviewed.
func reviewersProven(ctx context.Context, r *gitcmd.Repo, head string, f File, exporter ed25519.PublicKey) ([]ProvenReviewer, error) {
	ownerKeys := map[string]string{}
	for _, req := range f.Body.Requirements {
		pub, err := declaredKey(ctx, r, head, req.DeclaredOwner)
		if err != nil {
			return nil, err
		}
		if pub != nil {
			ownerKeys[string(pub)] = req.DeclaredOwner
		}
	}
	names := reviewersOf(f.Body)
	keys := map[string]ed25519.PublicKey{}
	byKey := map[string]string{}
	for _, name := range names {
		pub, err := declaredKey(ctx, r, head, name)
		if err != nil {
			return nil, err
		}
		if pub == nil {
			continue
		}
		if bytes.Equal(pub, exporter) {
			return nil, fmt.Errorf("reviewer %q's declared key is the exporter's key", name)
		}
		if owner, ok := ownerKeys[string(pub)]; ok {
			return nil, fmt.Errorf("reviewer %q's declared key is the declared owner %q's key", name, owner)
		}
		if other, ok := byKey[string(pub)]; ok {
			return nil, fmt.Errorf("reviewers %q and %q share one declared key", other, name)
		}
		keys[name], byKey[string(pub)] = pub, name
	}
	cosigned := map[string]bool{}
	for _, c := range f.Cosignatures {
		pub, _ := hex.DecodeString(c.Key)
		name, ok := byKey[string(pub)]
		if !ok {
			return nil, fmt.Errorf("cosignature by key %s, which is not the declared key of a reviewer of this warranty", Fingerprint(pub))
		}
		cosigned[name] = true
	}
	signedAll := map[string]bool{}
	for _, name := range names {
		signedAll[name] = true
	}
	artifacts := make([][2]string, 0, len(f.Body.Reviews)+len(f.Body.Verdicts))
	for _, rv := range f.Body.Reviews {
		artifacts = append(artifacts, [2]string{rv.ArtifactPath, rv.DeclaredReviewer})
	}
	for _, v := range f.Body.Verdicts {
		artifacts = append(artifacts, [2]string{v.ArtifactPath, v.DeclaredReviewer})
	}
	for _, a := range artifacts {
		path, name := a[0], a[1]
		line, err := optionalBlobAtHead(ctx, r, head, path+".sig")
		if err != nil {
			return nil, err
		}
		if line == nil {
			signedAll[name] = false
			continue
		}
		data, err := blobAtHead(ctx, r, head, path)
		if err != nil {
			return nil, err
		}
		pub, err := checkArtifactSig(data, line)
		if err != nil {
			return nil, fmt.Errorf("review signature %s.sig: %v", path, err)
		}
		if keys[name] == nil || !bytes.Equal(pub, keys[name]) {
			return nil, fmt.Errorf("review signature %s.sig is not by the declared key of reviewer %q", path, name)
		}
	}
	var proven []ProvenReviewer
	var missing []string
	for _, name := range names {
		if cosigned[name] || (signedAll[name] && keys[name] != nil) {
			proven = append(proven, ProvenReviewer{Name: name, Fingerprint: Fingerprint(keys[name])})
		} else {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return proven, fmt.Errorf("%w: reviewer %s has not signed (review seal with a declared key, or warranty cosign)", ErrUnreviewed, strings.Join(missing, ", "))
	}
	return proven, nil
}

// optionalBlobAtHead is the blob at path in HEAD, nil when nothing is there; a symlink or a tree is an
// error, never "absent".
func optionalBlobAtHead(ctx context.Context, r *gitcmd.Repo, head, path string) ([]byte, error) {
	entries, err := r.TreeEntries(ctx, head, path, false, false)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	if len(entries) != 1 || entries[0].Type != "blob" || entries[0].Mode == "120000" {
		return nil, fmt.Errorf("%s is not a regular committed file", path)
	}
	return r.Blob(ctx, entries[0].OID)
}
