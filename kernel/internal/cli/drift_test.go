package cli

import (
	"reflect"
	"testing"
)

// CLI-INV-21: drift compares the promise at HEAD with the verified promise, never digests and never
// ledger order.
func TestDriftComparesPromises(t *testing.T) {
	a := promise{"O-A": "The system SHALL log in.", "O-B": "The system SHALL log out."}
	b := promise{"O-A": "The system SHALL log in with MFA.", "O-C": "The system SHALL lock out."}
	promises := map[string]promise{
		"CI-x\x00A":        a,
		"CI-x\x00archived": {"O-A": "The system SHALL log in.", "O-B": "The system SHALL log out."},
		"CI-x\x00B":        b,
	}
	onA := warranty{verdictID: "OV-A", commit: "ca", intentID: "CI-x", digest: "A", seq: 10}
	onB := warranty{verdictID: "OV-B", commit: "cb", intentID: "CI-x", digest: "B", seq: 30}
	aToB := []driftChange{
		{"changed", "O-A", "The system SHALL log in.", "The system SHALL log in with MFA."},
		{"removed", "O-B", "The system SHALL log out.", ""},
		{"added", "O-C", "", "The system SHALL lock out."},
	}
	bToA := []driftChange{
		{"changed", "O-A", "The system SHALL log in with MFA.", "The system SHALL log in."},
		{"added", "O-B", "", "The system SHALL log out."},
		{"removed", "O-C", "The system SHALL lock out.", ""},
	}
	cases := map[string]struct {
		warranties []warranty
		head       headState
		state      string
		verdict    string
		changes    []driftChange
	}{
		"head is the verified revision":    {[]warranty{onA}, headState{digest: "A"}, "CURRENT", "OV-A", nil},
		"archive move, same promise":       {[]warranty{onA}, headState{digest: "archived"}, "CURRENT", "OV-A", nil},
		"edited promise":                   {[]warranty{onA}, headState{digest: "B"}, "STALE", "OV-A", aToB},
		"re-verified after the edit":       {[]warranty{onA, onB}, headState{digest: "B"}, "CURRENT", "OV-B", nil},
		"verified A, edited, reverted":     {[]warranty{onA}, headState{digest: "A"}, "CURRENT", "OV-A", nil},
		"verified B only, reverted to A":   {[]warranty{onB}, headState{digest: "A"}, "STALE", "OV-B", bToA},
		"head revision not synced":         {[]warranty{onA}, headState{digest: "C"}, "UNSYNCED", "OV-A", nil},
		"change deleted at head":           {[]warranty{onA}, headState{state: "REMOVED"}, "REMOVED", "OV-A", nil},
		"verified both, head back on A":    {[]warranty{onA, onB}, headState{digest: "A"}, "CURRENT", "OV-A", nil},
		"verified both, head on unknown C": {[]warranty{onA, onB}, headState{digest: "C"}, "UNSYNCED", "OV-B", nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := computeDrift(c.warranties, map[string]headState{"CI-x": c.head}, promises)
			if len(got) != 1 {
				t.Fatalf("entries %+v", got)
			}
			e := got[0]
			if e.state != c.state || e.verified.verdictID != c.verdict || !reflect.DeepEqual(e.changes, c.changes) {
				t.Fatalf("got state=%s verdict=%s changes=%+v", e.state, e.verified.verdictID, e.changes)
			}
		})
	}
	if got := computeDrift(nil, map[string]headState{"CI-x": {digest: "A"}}, promises); len(got) != 0 {
		t.Fatalf("an unwarranted change was reported: %+v", got)
	}
}

func TestChangeNameStripsTheArchiveDate(t *testing.T) {
	for in, want := range map[string]string{
		"openspec/changes/add-auth":                    "add-auth",
		"openspec/changes/archive/2026-11-19-add-auth": "add-auth",
		"openspec/changes/2026-11-19":                  "2026-11-19",
	} {
		if got := changeName(in); got != want {
			t.Errorf("changeName(%q) = %q, want %q", in, got, want)
		}
	}
}
