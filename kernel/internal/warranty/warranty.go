// Package warranty is the portable signed warranty of R3 slice 1 (SPEC.md): a file that says a
// commit implements a spec change and was independently verified, signed by a key the repository
// declares, and checkable from a clone of the repository alone. It never opens a database (W-INV-7).
package warranty

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// Schema names the file format; it is also the domain tag of the signed bytes.
const Schema = "proofbound.warranty.v1"

// Body is everything the signature covers. Issuer, IssuedAt and RepositoryHint are signed but
// informational: nothing in a repository can confirm them.
type Body struct {
	Schema         string                        `json:"schema"`
	Commit         string                        `json:"commit"`
	ChangeName     string                        `json:"change_name"`
	IssuedAt       string                        `json:"issued_at"`
	Issuer         string                        `json:"issuer"`
	RepositoryHint string                        `json:"repository_hint"`
	ChangeIntent   connectorintent.ChangeIntent  `json:"change_intent"`
	Requirements   []connectorintent.Requirement `json:"requirements"`
	Reviews        []reviews.RequirementReview   `json:"reviews"`
	Verdicts       []reviews.ObligationVerdict   `json:"verdicts"`
}

// Signature is the ed25519 public key and signature, both hex.
type Signature struct {
	Key string `json:"key"`
	Sig string `json:"sig"`
}

// File is a parsed warranty: its body, the exporter's signature, any reviewer cosignatures (option
// B, sorted by key), and the exact canonical bytes the signatures cover.
type File struct {
	Body         Body
	Signature    Signature
	Cosignatures []Signature
	canonical    []byte
}

// outerFile is the on-disk shape; cosignatures is omitted when there are none, so a file without
// them is byte-for-byte the slice 1 format.
type outerFile struct {
	Body         json.RawMessage `json:"body"`
	Signature    Signature       `json:"signature"`
	Cosignatures []Signature     `json:"cosignatures,omitempty"`
}

func marshalFile(canonical []byte, sig Signature, cosigs []Signature) ([]byte, error) {
	out, err := json.MarshalIndent(outerFile{canonical, sig, cosigs}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

var hexRE = regexp.MustCompile(`^[0-9a-f]+$`)

const maxCosignatures = 16

// validCosignatures enforces W-INV-11: exact hex sizes, sorted by key, no duplicates, and present only
// when non-empty (an explicit empty list is not the canonical form).
func validCosignatures(cosigs []Signature, present bool) error {
	if present && len(cosigs) == 0 {
		return errors.New("empty cosignatures")
	}
	if len(cosigs) > maxCosignatures {
		return fmt.Errorf("more than %d cosignatures", maxCosignatures)
	}
	for i, c := range cosigs {
		if len(c.Key) != 2*ed25519.PublicKeySize || len(c.Sig) != 2*ed25519.SignatureSize || !hexRE.MatchString(c.Key) || !hexRE.MatchString(c.Sig) {
			return errors.New("cosignature is not hex ed25519")
		}
		if i > 0 && cosigs[i-1].Key >= c.Key {
			return errors.New("cosignatures are not sorted and unique by key")
		}
	}
	return nil
}

var commitRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

func signedBytes(canonical []byte) []byte {
	return append([]byte(Schema+"\n"), canonical...)
}

func canonicalBody(body Body) ([]byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return core.Canonicalize(raw)
}

// Fingerprint is the first 16 hex characters of the sha256 of the raw public key.
func Fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])[:16]
}

// Sign returns the warranty file bytes for body, signed by key.
func Sign(body Body, key ed25519.PrivateKey) ([]byte, error) {
	canonical, err := canonicalBody(body)
	if err != nil {
		return nil, err
	}
	pub, ok := key.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("warranty: key is not ed25519")
	}
	return marshalFile(canonical, Signature{Key: hex.EncodeToString(pub), Sig: hex.EncodeToString(ed25519.Sign(key, signedBytes(canonical)))}, nil)
}

func strictDecode(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if d.More() {
		return errors.New("trailing data")
	}
	return nil
}

// Parse reads a warranty strictly (W-INV-1): duplicate keys, unknown fields, trailing data, or a
// body that does not round-trip to the same canonical bytes are all malformed.
func Parse(data []byte) (File, error) {
	if _, err := core.Canonicalize(data); err != nil {
		return File{}, fmt.Errorf("malformed: %w", err)
	}
	var outer outerFile
	if err := strictDecode(data, &outer); err != nil {
		return File{}, fmt.Errorf("malformed: %w", err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return File{}, fmt.Errorf("malformed: %w", err)
	}
	_, present := keys["cosignatures"]
	if err := validCosignatures(outer.Cosignatures, present); err != nil {
		return File{}, fmt.Errorf("malformed: %w", err)
	}
	var f File
	if err := strictDecode(outer.Body, &f.Body); err != nil {
		return File{}, fmt.Errorf("malformed: body: %w", err)
	}
	canonical, err := core.Canonicalize(outer.Body)
	if err != nil {
		return File{}, fmt.Errorf("malformed: body: %w", err)
	}
	again, err := canonicalBody(f.Body)
	if err != nil || !bytes.Equal(again, canonical) {
		return File{}, errors.New("malformed: body is not in canonical warranty form")
	}
	if f.Body.Schema != Schema {
		return File{}, fmt.Errorf("malformed: schema %q", f.Body.Schema)
	}
	f.Signature, f.Cosignatures, f.canonical = outer.Signature, outer.Cosignatures, canonical
	return f, nil
}

// PublicKey decodes the file's signing key.
func (f File) PublicKey() (ed25519.PublicKey, error) {
	raw, err := hex.DecodeString(f.Signature.Key)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return nil, errors.New("signature: key is not a hex ed25519 public key")
	}
	return ed25519.PublicKey(raw), nil
}

// CheckSignature verifies the ed25519 signature over the domain-tagged canonical body.
func (f File) CheckSignature() error {
	pub, err := f.PublicKey()
	if err != nil {
		return err
	}
	sig, err := hex.DecodeString(f.Signature.Sig)
	if err != nil || len(sig) != ed25519.SignatureSize || !ed25519.Verify(pub, signedBytes(f.canonical), sig) {
		return errors.New("signature")
	}
	return nil
}

func key(id, digest string) string { return id + "@" + digest }

// CheckRule re-runs the readiness rule on the body alone (W-INV-3, SPEC §5.3): every target
// obligation SATISFIED by a verdict for this commit and change revision, and VERIFIABLE by a review
// of that requirement revision; no reviewer is the requirement's declared owner.
func CheckRule(b Body) error {
	if !commitRE.MatchString(b.Commit) {
		return errors.New("commit is not 40 hex")
	}
	ci := b.ChangeIntent
	if len(ci.Targets) == 0 {
		return errors.New("change intent has no targets")
	}
	owners := map[string]string{}
	obligations := map[string]map[string]bool{}
	for _, r := range b.Requirements {
		k := key(r.RequirementID, r.ArtifactSHA256)
		if _, dup := owners[k]; dup {
			return fmt.Errorf("requirement %s listed twice", r.RequirementID)
		}
		owners[k] = r.DeclaredOwner
		obligations[k] = map[string]bool{}
		for _, o := range r.Obligations {
			obligations[k][o.ID] = true
		}
	}
	// Identity is KeyName (W-INV-3 amended, MVP step 3): no empty key name, no reviewer whose key
	// name is an owner's, and one spelling per key name.
	for _, owner := range owners {
		if KeyName(owner) == "" {
			return fmt.Errorf("owner %q has no usable key name", owner)
		}
	}
	spelling := map[string]string{}
	for _, name := range reviewersOf(b) {
		k := KeyName(name)
		if k == "" {
			return fmt.Errorf("reviewer %q has no usable key name", name)
		}
		if other, ok := spelling[k]; ok {
			return fmt.Errorf("reviewers %q and %q are one key name", other, name)
		}
		spelling[k] = name
	}
	targeted := map[string]bool{}
	for _, t := range ci.Targets {
		k := key(t.RecordID, t.ArtifactSHA256)
		if _, ok := owners[k]; !ok {
			return fmt.Errorf("target %s has no requirement in the warranty", t.RecordID)
		}
		targeted[k] = true
		for _, id := range t.ObligationIDs {
			if !obligations[k][id] {
				return fmt.Errorf("target obligation %s is not in requirement %s", id, t.RecordID)
			}
		}
	}
	for k := range owners {
		if !targeted[k] {
			return fmt.Errorf("requirement %s is not targeted by the change", k)
		}
	}
	satisfied, reviewed := map[string]bool{}, map[string]bool{}
	for _, v := range b.Verdicts {
		if v.ReviewedCommit != b.Commit {
			return fmt.Errorf("verdict %s is for commit %s, not %s", v.VerdictID, v.ReviewedCommit, b.Commit)
		}
		if v.ChangeIntent.RecordID != ci.IntentID || v.ChangeIntent.ArtifactSHA256 != ci.ArtifactSHA256 {
			return fmt.Errorf("verdict %s evaluates another change intent revision", v.VerdictID)
		}
		for _, o := range v.Obligations {
			k := key(o.RequirementID, o.ArtifactSHA256)
			owner, ok := owners[k]
			if !ok {
				return fmt.Errorf("verdict %s cites untargeted requirement %s", v.VerdictID, o.RequirementID)
			}
			if v.DeclaredReviewer == owner || KeyName(v.DeclaredReviewer) == KeyName(owner) {
				return fmt.Errorf("verdict %s reviewer %q is the declared owner", v.VerdictID, owner)
			}
			if o.Outcome == "SATISFIED" {
				satisfied[k+"#"+o.ObligationID] = true
			}
		}
	}
	for _, r := range b.Reviews {
		k := key(r.Requirement.RecordID, r.Requirement.ArtifactSHA256)
		owner, ok := owners[k]
		if !ok {
			return fmt.Errorf("review %s cites untargeted requirement %s", r.ReviewID, r.Requirement.RecordID)
		}
		if r.DeclaredReviewer == owner || KeyName(r.DeclaredReviewer) == KeyName(owner) {
			return fmt.Errorf("review %s reviewer %q is the declared owner", r.ReviewID, owner)
		}
		for _, o := range r.Outcomes {
			if o.Outcome == "VERIFIABLE" {
				reviewed[k+"#"+o.ObligationID] = true
			}
		}
	}
	var missing []string
	for _, t := range ci.Targets {
		k := key(t.RecordID, t.ArtifactSHA256)
		for _, id := range t.ObligationIDs {
			if !satisfied[k+"#"+id] {
				missing = append(missing, fmt.Sprintf("obligation %s of %s is not SATISFIED", id, t.RecordID))
			}
			if !reviewed[k+"#"+id] {
				missing = append(missing, fmt.Sprintf("obligation %s of %s is not reviewed VERIFIABLE", id, t.RecordID))
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return errors.New(missing[0])
	}
	return nil
}
