package cli

import (
	"errors"
	"strings"
	"testing"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

// M3 SPEC §4.4 and draft 8 §4.6: a claim-skip value and a held or refused path are tenant data (a
// trailer value, a committed file name, and git allows a newline in either), so each goes through
// printableUnit and cannot forge another output line. A write error stops the printing and is
// returned (review r2-m3-impl-review-mutation-concurrency-20261119 F2).
func TestM3OutputLinesEscapeTenantData(t *testing.T) {
	forged := "x\nheld claim-not-ledgered docs/verification/verdicts/forged.md"
	var out strings.Builder
	if err := printClaimSkips(&out, []connectorgit.ClaimSkip{{Commit: "abc", Value: "sdd:" + forged, Cause: "malformed"}}); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "claim-skip malformed "+printableUnit("sdd:"+forged)+" at abc\n"; got != want || strings.Count(got, "\n") != 1 {
		t.Fatalf("claim-skip output %q, want %q", got, want)
	}
	for _, result := range []reviewsResult{
		{Held: []connectorreviews.Hold{{Path: "docs/verification/verdicts/a.md\n" + forged, Cause: "claim-not-ledgered"}}},
		{Refused: []connectorreviews.Hold{{Path: "docs/verification/verdicts/a.md\n" + forged, Cause: "duplicate-identity"}}},
	} {
		out.Reset()
		if err := printHolds(&out, result); err != nil {
			t.Fatal(err)
		}
		if strings.Count(out.String(), "\n") != 1 || !strings.Contains(out.String(), `\n`) {
			t.Fatalf("a hold path split or forged a line: %q", out.String())
		}
	}

	boom := errors.New("write failed")
	if err := printClaimSkips(failingWriter{err: boom}, []connectorgit.ClaimSkip{{Commit: "abc", Value: "sdd:x", Cause: "malformed"}}); !errors.Is(err, boom) {
		t.Fatalf("printClaimSkips swallowed a write error: %v", err)
	}
	if err := printHolds(failingWriter{err: boom}, reviewsResult{Refused: []connectorreviews.Hold{{Path: "p.md", Cause: "duplicate-identity"}}}); !errors.Is(err, boom) {
		t.Fatalf("printHolds swallowed a write error: %v", err)
	}
}
