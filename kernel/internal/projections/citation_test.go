package projections

import (
	"testing"
	"time"
)

// TestCommitValidationAcceptsBothDecisionSuffixForms guards the decision-citation validator against
// the real-data regression found by self-hosting: a commit citing a date-suffixed VD
// (e.g. VD-p5-intent-provenance-2026-09-10 — the /proofbound-decide phase-decision form) failed
// closed, crashing the projection apply and therefore `report week`, `rebuild`, and `verify` on this
// repo's own 17/31 date-suffixed decisions. This test is deliberately NOT integration-tagged so it
// runs under `make check` (kernel-check runs `go test ./...` without -tags integration).
func TestCommitValidationAcceptsBothDecisionSuffixForms(t *testing.T) {
	base := func(id string) commitPayload {
		return commitPayload{
			SHA:            "0123456789abcdef0123456789abcdef01234567",
			AuthorName:     "Author",
			AuthorEmail:    "author@example.test",
			CommitterName:  "Committer",
			CommitterEmail: "committer@example.test",
			CommittedAt:    time.Now(),
			Subject:        "subject",
			CitedDecisions: []string{id},
		}
	}
	// Both /proofbound-decide suffix conventions must be accepted: a YYYY-MM-DD phase-decision suffix
	// (the real repo form that previously failed closed) and a random 6-char standing-decision token.
	for _, id := range []string{
		"VD-p5-intent-provenance-2026-09-10",
		"VD-p7-def001-resolve-2026-11-05",
		"VD-release-line-kwdh4v",
		"VD-stack-go-fid9mi",
	} {
		if err := base(id).validate(); err != nil {
			t.Fatalf("valid decision citation %q rejected: %v", id, err)
		}
	}
	// Malformed citations still fail closed (P-INV-7): too-short token, an incomplete date tail, a
	// non-citation string, and an uppercase slug char.
	for _, id := range []string{"VD-bad-123", "VD-oops-2026-13", "not-a-decision", "VD-Upper-abc123"} {
		if err := base(id).validate(); err == nil {
			t.Fatalf("malformed decision citation %q accepted", id)
		}
	}
}
