//go:build !js

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

const (
	testCommit = "0123456789abcdef0123456789abcdef01234567"
	reqDigest  = "1111111111111111111111111111111111111111111111111111111111111111"
	ciDigest   = "2222222222222222222222222222222222222222222222222222222222222222"
)

// body is a minimal warranty that passes the rule (the shape of the warranty package's sampleBody).
func body(status string) pbwarranty.Body {
	req := connectorintent.Reference{RecordKind: "requirement", Source: "intent.sdd", RecordID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest, Relation: "implements", ObligationIDs: []string{"O-A"}}
	return pbwarranty.Body{
		Schema: pbwarranty.Schema, Commit: testCommit, ChangeName: "add-x", IssuedAt: "2026-11-25T00:00:00Z", Issuer: "owner",
		ChangeIntent: connectorintent.ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: "CI-add-x-bbbbbb", Status: "proposed", Targets: []connectorintent.Reference{req}, ArtifactSHA256: ciDigest},
		Requirements: []connectorintent.Requirement{{Schema: "proofbound.requirement.v1", RequirementID: "BR-x-aaaaaa", DeclaredOwner: "owner", ArtifactSHA256: reqDigest, Obligations: []connectorintent.Obligation{{ID: "O-A", Statement: "A."}}}},
		Reviews: []reviews.RequirementReview{{ReviewID: "RR-1", Requirement: connectorintent.Reference{RecordID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest}, DeclaredReviewer: "reviewer",
			Outcomes: []reviews.RequirementReviewOutcome{{ObligationID: "O-A", Outcome: "VERIFIABLE"}}}},
		Verdicts: []reviews.ObligationVerdict{{VerdictID: "OV-1", Status: status, DeclaredReviewer: "reviewer", ReviewedCommit: testCommit,
			ChangeIntent: connectorintent.Reference{RecordID: "CI-add-x-bbbbbb", ArtifactSHA256: ciDigest},
			Obligations:  []reviews.ObligationOutcome{{RequirementID: "BR-x-aaaaaa", ArtifactSHA256: reqDigest, ObligationID: "O-A", Outcome: "SATISFIED"}}}},
	}
}

// §2.3: the page's answers are the warranty package's, for the same files.
func TestCheck(t *testing.T) {
	key := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	fp := pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey))
	good, err := pbwarranty.Sign(body("ACCEPTABLE"), key)
	if err != nil {
		t.Fatal(err)
	}
	if r := check(good, fp); r.State != "SIGNATURE ONLY" || r.Commit != testCommit || r.Key != fp || r.Claims != 1 {
		t.Fatalf("a good warranty: %+v", r)
	}
	if r := check(good, " "+strings.ToUpper(fp)+" "); r.State != "SIGNATURE ONLY" {
		t.Errorf("a pasted key with spaces and capitals: %+v", r)
	}
	other := pbwarranty.Fingerprint(ed25519.NewKeyFromSeed(bytes.Repeat([]byte{1}, ed25519.SeedSize)).Public().(ed25519.PublicKey))
	if r := check(good, other); r.State != "INVALID" || !strings.Contains(r.Reason, "fingerprint mismatch") {
		t.Errorf("another key: %+v", r)
	}
	// The key is never taken from the file: with none typed, nothing is checked.
	if r := check(good, ""); r.State != "INVALID" || !strings.Contains(r.Reason, "type the trust key") || r.Commit != "" {
		t.Errorf("no key typed: %+v", r)
	}
	altered := bytes.Replace(good, []byte(`"add-x"`), []byte(`"add-y"`), 1)
	if r := check(altered, fp); r.State != "INVALID" {
		t.Errorf("an altered byte: %+v", r)
	}
	if r := check([]byte("{}"), fp); r.State != "INVALID" {
		t.Errorf("not a warranty: %+v", r)
	}
	bad, err := pbwarranty.Sign(body("NEEDS_WORK"), key)
	if err != nil {
		t.Fatal(err)
	}
	if r := check(bad, fp); r.State != "INVALID" {
		t.Errorf("a re-signed NEEDS_WORK verdict: %+v", r)
	}
}

// The page carries the loader and the compressed WebAssembly, and nothing is left unfilled.
func TestAssemble(t *testing.T) {
	dir := t.TempDir()
	wasm := append([]byte("\x00asm\x01\x00\x00\x00"), bytes.Repeat([]byte("x"), 1000)...)
	wp, ep := filepath.Join(dir, "v.wasm"), filepath.Join(dir, "wasm_exec.js")
	if err := os.WriteFile(wp, wasm, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ep, []byte("globalThis.Go = function () {};"), 0o644); err != nil {
		t.Fatal(err)
	}
	page, err := assemble(wp, ep, "v0.3.2<")
	if err != nil {
		t.Fatal(err)
	}
	s := string(page)
	if strings.Contains(s, "{{") || !strings.Contains(s, "globalThis.Go = function") || !strings.Contains(s, "Proofbound v0.3.2&lt;") {
		t.Fatal("the page is not filled")
	}
	m := regexp.MustCompile(`<script id="wasm" type="application/octet-stream">([A-Za-z0-9+/=]+)</script>`).FindStringSubmatch(s)
	if m == nil {
		t.Fatal("no WebAssembly in the page")
	}
	gz, _ := base64.StdEncoding.DecodeString(m[1])
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	if !bytes.Equal(got, wasm) {
		t.Error("the WebAssembly does not round-trip")
	}
	if err := os.WriteFile(wp, []byte("not wasm"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := assemble(wp, ep, "v"); err == nil {
		t.Error("a file that is not WebAssembly was assembled")
	}
}
