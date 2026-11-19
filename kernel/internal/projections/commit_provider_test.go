package projections

import (
	"strings"
	"testing"
	"time"
)

// P-INV-33 (R2 M3 SPEC draft 5 §4.5): commit claims accept exactly records, specdir and sdd. This
// test is untagged so it runs inside make check.
func TestCommitClaimProvidersAreClosed(t *testing.T) {
	payload := func(provider string) commitPayload {
		return commitPayload{
			SHA: strings.Repeat("c", 40), AuthorName: "Author", AuthorEmail: "author@example.test",
			CommitterName: "Committer", CommitterEmail: "committer@example.test", CommittedAt: time.Unix(1, 0).UTC(),
			Subject:    "subject",
			IntentRefs: []commitIntentRef{{Provider: provider, RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: strings.Repeat("b", 64)}},
		}
	}
	for _, provider := range []string{"records", "specdir", "sdd"} {
		v := payload(provider)
		if err := v.validate(); err != nil {
			t.Errorf("provider %s rejected: %v", provider, err)
		}
	}
	for _, provider := range []string{"", "github", "SDD", "intent.sdd", "sdd "} {
		v := payload(provider)
		if err := v.validate(); err == nil {
			t.Errorf("provider %q accepted", provider)
		}
	}
}
