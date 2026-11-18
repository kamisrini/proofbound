package cli

import (
	"strings"
	"testing"
	"time"
)

// TestPrintableUnitQuotesEveryLineBreaker extends sdd SDD-INV-60 (re-check
// r2-m1-recheck-round1-20261118 F3): C1 controls (NEL, CSI) and U+2028/U+2029 are quoted like C0
// and DEL, and the quoted form carries no raw line breaker; ordinary names, non-ASCII letters
// included, print unchanged.
func TestPrintableUnitQuotesEveryLineBreaker(t *testing.T) {
	for _, raw := range []string{"a\nb", "a\x7fb", "a\u0085b", "a\u009bb", "a\u0080b", "a\u009fb", "a b", "a b", "a\xffb"} {
		got := printableUnit(raw)
		if got == raw || !strings.HasPrefix(got, `"`) {
			t.Errorf("%q printed unquoted: %q", raw, got)
		}
		for _, r := range got {
			if r < 0x20 || (r >= 0x7f && r <= 0x9f) || r == 0x2028 || r == 0x2029 {
				t.Errorf("%q: quoted form %q still carries %U", raw, got, r)
			}
		}
	}
	for _, plain := range []string{"change add-login", "archive 2026-01-01-x", "change café", "change  nbsp"} {
		if got := printableUnit(plain); got != plain {
			t.Errorf("%q altered to %q", plain, got)
		}
	}
}

// TestAllSelectionNeverConstructsSDD extends sdd SDD-INV-56 to `proofbound verify` (round-1 F8):
// verify's intent syncs call syncIntentOnStore with "all" only, and the "all" selection builds
// records and specdir without touching git — it succeeds outside any repository — and never sdd.
func TestAllSelectionNeverConstructsSDD(t *testing.T) {
	providers, err := intentProviders(t.TempDir(), "all", time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatalf("all outside a repository: %v", err)
	}
	var names []string
	for _, p := range providers {
		names = append(names, p.Name())
	}
	if strings.Join(names, ",") != "records,specdir" {
		t.Fatalf("all built %v, want records,specdir", names)
	}
	if _, err := intentProviders(t.TempDir(), "sdd", time.Unix(1, 0).UTC()); err == nil {
		t.Fatal("calibration: sdd outside a repository must fail to construct")
	}
}
