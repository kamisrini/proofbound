package gates

import "testing"

// GATE-INV-16: the readiness reason states the counts the rule decided on.
func TestReadinessReasonStatesTheCounts(t *testing.T) {
	for _, c := range []struct {
		targets, satisfied, reviewed int
		want                         string
	}{
		{0, 0, 0, "commit binds no change intent"},
		{4, 0, 0, "0 of 4 obligations verified SATISFIED; 0 of 4 requirement reviews VERIFIABLE"},
		{4, 4, 3, "4 of 4 obligations verified SATISFIED; 3 of 4 requirement reviews VERIFIABLE"},
		{1, 1, 1, "1 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"},
	} {
		if got := readinessReason(c.targets, c.satisfied, c.reviewed); got != c.want {
			t.Errorf("readinessReason(%d,%d,%d) = %q, want %q", c.targets, c.satisfied, c.reviewed, got, c.want)
		}
	}
}

// GATE-INV-17 (docs/plans/needs-work-blocks-spec.md): a current NEEDS_WORK verdict blocks the commit
// whatever the counts, naming the verdicts; an unknown verdict status fails closed; otherwise every
// target obligation must be verified and reviewed. The SQL that finds the verdicts is proven end to
// end by the sdd package's TestNeedsWorkBlocksEndToEnd.
func TestReadinessDecisionHonoursNeedsWork(t *testing.T) {
	for _, c := range []struct {
		targets, satisfied, reviewed int
		needsWork                    string
		unknown                      int
		state                        State
		reason                       string
	}{
		{1, 1, 1, "", 0, StatePass, "1 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"},
		{1, 1, 1, "OV-abc-1", 0, StateBlocked, "the reviewer recorded NEEDS_WORK (OV-abc-1); 1 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"},
		{2, 2, 2, "OV-abc-1, OV-abc-2", 0, StateBlocked, "the reviewer recorded NEEDS_WORK (OV-abc-1, OV-abc-2); 2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE"},
		{1, 1, 1, "", 1, StateBlocked, "verdict status unknown; run proofbound rebuild; 1 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"},
		{1, 0, 1, "", 0, StateBlocked, "0 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"},
		{0, 0, 0, "", 0, StateBlocked, "commit binds no change intent"},
	} {
		state, reason := readinessDecision(c.targets, c.satisfied, c.reviewed, c.needsWork, c.unknown)
		if state != c.state || reason != c.reason {
			t.Errorf("readinessDecision(%d,%d,%d,%q,%d) = %s %q, want %s %q", c.targets, c.satisfied, c.reviewed, c.needsWork, c.unknown, state, reason, c.state, c.reason)
		}
	}
}
