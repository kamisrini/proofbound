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
