package cli

import (
	"reflect"
	"testing"
)

// CLI-INV-21: drift compares promises, not digests.
func TestDriftComparesPromises(t *testing.T) {
	verified := promise{"O-A": "The system SHALL log in.", "O-B": "The system SHALL log out."}
	promises := map[string]promise{
		"CI-x\x00v1":       verified,
		"CI-x\x00archived": {"O-A": "The system SHALL log in.", "O-B": "The system SHALL log out."},
		"CI-x\x00edited":   {"O-A": "The system SHALL log in with MFA.", "O-C": "The system SHALL lock out."},
		"CI-x\x00v1-again": verified,
	}
	w1 := warranty{verdictID: "OV-1", commit: "c1", intentID: "CI-x", digest: "v1", seq: 10}
	cases := map[string]struct {
		warranties []warranty
		revisions  []intentRevision
		stale      bool
		verdict    string
		changes    []driftChange
	}{
		"only the verified revision": {[]warranty{w1}, []intentRevision{{"v1", 1}}, false, "OV-1", nil},
		"archive move, same promise": {[]warranty{w1}, []intentRevision{{"v1", 1}, {"archived", 20}}, false, "OV-1", nil},
		"edited promise": {[]warranty{w1}, []intentRevision{{"v1", 1}, {"edited", 20}}, true, "OV-1", []driftChange{
			{"changed", "O-A", "The system SHALL log in.", "The system SHALL log in with MFA."},
			{"removed", "O-B", "The system SHALL log out.", ""},
			{"added", "O-C", "", "The system SHALL lock out."},
		}},
		"re-verified after the edit": {[]warranty{w1, {verdictID: "OV-2", commit: "c2", intentID: "CI-x", digest: "edited", seq: 30}}, []intentRevision{{"v1", 1}, {"edited", 20}}, false, "OV-2", nil},
		"edit reverted":              {[]warranty{w1}, []intentRevision{{"v1", 1}, {"edited", 20}, {"v1-again", 40}}, false, "OV-1", nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := computeDrift(c.warranties, map[string][]intentRevision{"CI-x": c.revisions}, promises)
			if len(got) != 1 {
				t.Fatalf("entries %+v", got)
			}
			e := got[0]
			if e.stale != c.stale || e.verified.verdictID != c.verdict || !reflect.DeepEqual(e.changes, c.changes) {
				t.Fatalf("got stale=%v verdict=%s changes=%+v", e.stale, e.verified.verdictID, e.changes)
			}
		})
	}
	if got := computeDrift(nil, map[string][]intentRevision{"CI-x": {{"v1", 1}}}, promises); len(got) != 0 {
		t.Fatalf("an unwarranted change was reported: %+v", got)
	}
}
