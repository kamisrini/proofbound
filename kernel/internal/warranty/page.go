package warranty

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// PageResult is what the verify page (proofbound-verify.html, docs/plans/ui-reviewer-spec.md §2.3)
// shows for one warranty file.
type PageResult struct {
	State  string `json:"state"` // "SIGNATURE ONLY" or "INVALID"
	Reason string `json:"reason,omitempty"`
	Commit string `json:"commit,omitempty"`
	Change string `json:"change,omitempty"`
	Issuer string `json:"issuer,omitempty"`
	Key    string `json:"key,omitempty"`
	Claims int    `json:"obligations,omitempty"`
}

var trustKeyRE = regexp.MustCompile(`^[0-9a-f]{16}$`)

// PageCheck verifies one warranty file with no repository against the trust key the reader typed
// (W-INV-13): Parse, then Verify, so the page runs this package's rule and nothing else. The key is
// never taken from the file; with no valid key typed, nothing is checked. The best answer without a
// repository is SIGNATURE ONLY, never VALID (W-INV-4).
func PageCheck(data []byte, trustKey string) PageResult {
	trustKey = strings.ToLower(strings.TrimSpace(trustKey))
	if !trustKeyRE.MatchString(trustKey) {
		return PageResult{State: "INVALID", Reason: "type the trust key: the 16-character fingerprint the vendor gave you by another channel (never the one printed inside the bundle)"}
	}
	f, err := Parse(data)
	if err != nil {
		return PageResult{State: "INVALID", Reason: "not a well-formed warranty: " + err.Error()}
	}
	res := PageResult{Commit: f.Body.Commit, Change: f.Body.ChangeName, Issuer: f.Body.Issuer}
	if pub, err := f.PublicKey(); err == nil {
		res.Key = Fingerprint(pub)
	}
	for _, v := range f.Body.Verdicts {
		res.Claims += len(v.Obligations)
	}
	err = Verify(context.Background(), f, "", trustKey)
	if !errors.Is(err, ErrUnverified) {
		res.State, res.Reason = "INVALID", err.Error()
		return res
	}
	res.State = "SIGNATURE ONLY"
	return res
}
