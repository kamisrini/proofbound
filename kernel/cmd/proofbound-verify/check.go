// Command proofbound-verify is the warranty verify page (docs/plans/ui-reviewer-spec.md §2.3).
//
// Built for the browser (GOOS=js GOARCH=wasm) it is the check itself: the same Go code as
// `proofbound warranty verify`, with no repository. Built natively it assembles the page: one
// self-contained HTML file that carries the compressed WebAssembly, published with each release and
// never inside a bundle, so a bundle cannot vouch for itself.
package main

import (
	"context"
	"errors"
	"regexp"
	"strings"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// pageResult is what the page shows for one warranty file.
type pageResult struct {
	State  string `json:"state"` // "SIGNATURE ONLY" or "INVALID"
	Reason string `json:"reason,omitempty"`
	Commit string `json:"commit,omitempty"`
	Change string `json:"change,omitempty"`
	Issuer string `json:"issuer,omitempty"`
	Key    string `json:"key,omitempty"`
	Claims int    `json:"obligations,omitempty"`
}

var trustKeyRE = regexp.MustCompile(`^[0-9a-f]{16}$`)

// check verifies one warranty file against the trust key the reader typed: parse (strict), the
// signature, the cosignatures and the rule, and the key's fingerprint. Without a repository the best
// answer is SIGNATURE ONLY, never VALID (W-INV-4). The key is never taken from the file.
func check(data []byte, trustKey string) pageResult {
	trustKey = strings.ToLower(strings.TrimSpace(trustKey))
	if !trustKeyRE.MatchString(trustKey) {
		return pageResult{State: "INVALID", Reason: "type the trust key: the 16-character fingerprint the vendor gave you by another channel (never the one printed inside the bundle)"}
	}
	f, err := pbwarranty.Parse(data)
	if err != nil {
		return pageResult{State: "INVALID", Reason: "not a well-formed warranty: " + err.Error()}
	}
	res := pageResult{Commit: f.Body.Commit, Change: f.Body.ChangeName, Issuer: f.Body.Issuer}
	if pub, err := f.PublicKey(); err == nil {
		res.Key = pbwarranty.Fingerprint(pub)
	}
	for _, v := range f.Body.Verdicts {
		res.Claims += len(v.Obligations)
	}
	err = pbwarranty.Verify(context.Background(), f, "", trustKey)
	if !errors.Is(err, pbwarranty.ErrUnverified) {
		res.State, res.Reason = "INVALID", err.Error()
		return res
	}
	res.State = "SIGNATURE ONLY"
	return res
}
