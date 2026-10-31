package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// TestParseTwinReplayFlags pins the flag grammar without needing a store.
func TestParseTwinReplayFlags(t *testing.T) {
	ok := map[string]struct {
		args      []string
		through   int64
		candidate string
	}{
		"bare":           {[]string{"twin", "replay"}, -1, ""},
		"through":        {[]string{"twin", "replay", "--through", "7"}, 7, ""},
		"candidate":      {[]string{"twin", "replay", "--candidate", "c.json"}, -1, "c.json"},
		"both":           {[]string{"twin", "replay", "--through", "0", "--candidate", "c.json"}, 0, "c.json"},
	}
	for name, tc := range ok {
		t.Run(name, func(t *testing.T) {
			through, candidate, err := parseTwinReplayFlags(tc.args)
			if err != nil || through != tc.through || candidate != tc.candidate {
				t.Fatalf("got through=%d candidate=%q err=%v", through, candidate, err)
			}
		})
	}
	bad := map[string][]string{
		"missing through value":   {"twin", "replay", "--through"},
		"missing candidate value": {"twin", "replay", "--candidate"},
		"negative through":        {"twin", "replay", "--through", "-3"},
		"non-numeric through":     {"twin", "replay", "--through", "x"},
		"unexpected argument":     {"twin", "replay", "--nope"},
	}
	for name, args := range bad {
		t.Run(name, func(t *testing.T) {
			if _, _, err := parseTwinReplayFlags(args); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// TestTwinReplayIsolatesAndProves (CLI-INV-11): with no candidate, `twin replay`
// re-derives the committed projection in an isolated store, prints a preserved
// verdict with a vera.replay.v1 proof, and leaves the source ledger unchanged.
func TestTwinReplayIsolatesAndProves(t *testing.T) {
	ledger := openTwinReplayTestStore(t, 55461)
	seedDeploymentEvents(t, ledger, 2)
	before := ledgerSeqs(t, ledger)

	var out bytes.Buffer
	if err := twinReplay(context.Background(), ledger, []string{"twin", "replay"}, &out); err != nil {
		t.Fatalf("twin replay: %v", err)
	}
	got := out.String()
	for _, want := range []string{"verdict=preserved", "candidate_events=0", "schema=vera.replay.v1", "source_digest=", "candidate_snapshot="} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
	if after := ledgerSeqs(t, ledger); !equalInt64(before, after) {
		t.Fatalf("source ledger changed: before=%v after=%v", before, after)
	}
}

// TestTwinReplayCounterfactualVerdict (CLI-INV-12): a --candidate file injects one
// counterfactual event; a state-changing candidate yields verdict=changed with a
// distinct candidate-snapshot digest, and the source ledger is still unchanged.
func TestTwinReplayCounterfactualVerdict(t *testing.T) {
	ledger := openTwinReplayTestStore(t, 55462)
	seedDeploymentEvents(t, ledger, 2)
	before := ledgerSeqs(t, ledger)

	path := filepath.Join(t.TempDir(), "candidate.json")
	candidate := struct {
		Seq   int64      `json:"seq"`
		Event core.Event `json:"event"`
	}{Seq: 3, Event: deploymentEvent(3)}
	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := twinReplay(context.Background(), ledger, []string{"twin", "replay", "--candidate", path}, &out); err != nil {
		t.Fatalf("twin replay --candidate: %v", err)
	}
	got := out.String()
	for _, want := range []string{"verdict=changed", "candidate_events=1", "schema=vera.replay.v1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
	// baseline (committed) and candidate snapshots must differ for a changed verdict.
	baseline := fieldValue(got, "baseline_snapshot=")
	candidateSnapshot := fieldValue(got, "candidate_snapshot=")
	if baseline == "" || candidateSnapshot == "" || baseline == candidateSnapshot {
		t.Fatalf("expected distinct snapshot digests, got baseline=%q candidate=%q", baseline, candidateSnapshot)
	}
	if after := ledgerSeqs(t, ledger); !equalInt64(before, after) {
		t.Fatalf("source ledger changed: before=%v after=%v", before, after)
	}
}

// TestTwinReplayThroughBelowLatestIsCoherent locks the fix for the review's
// finding 1: --through below the latest seq must compare against the projection
// of the SAME prefix (verdict preserved), not against the full committed
// projection (which would spuriously report changed with no candidate).
func TestTwinReplayThroughBelowLatestIsCoherent(t *testing.T) {
	ledger := openTwinReplayTestStore(t, 55463)
	seedDeploymentEvents(t, ledger, 3)

	var out bytes.Buffer
	if err := twinReplay(context.Background(), ledger, []string{"twin", "replay", "--through", "2"}, &out); err != nil {
		t.Fatalf("twin replay --through 2: %v", err)
	}
	got := out.String()
	for _, want := range []string{"verdict=preserved", "through_seq=2", "candidate_events=0"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q (prefix replay should be coherent):\n%s", want, got)
		}
	}
}

func openTwinReplayTestStore(t *testing.T, port uint16) *store.Store {
	t.Helper()
	ledger, err := store.Open(context.Background(), store.Config{Root: t.TempDir(), Port: port, DatabaseURL: os.Getenv("DATABASE_URL")})
	if err != nil {
		t.Skipf("embedded postgres unavailable: %v", err)
	}
	if os.Getenv("DATABASE_URL") != "" {
		if err := ledger.WithTx(context.Background(), func(ctx context.Context, tx *store.Tx) error {
			_, err := tx.Exec(ctx, "TRUNCATE TABLE events, sync_runs RESTART IDENTITY CASCADE")
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = ledger.Close() })
	return ledger
}

func seedDeploymentEvents(t *testing.T, ledger *store.Store, n int) {
	t.Helper()
	sync, err := ledger.BeginSync(context.Background(), "twin-replay-test")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= n; i++ {
		if _, _, err := sync.Append(context.Background(), deploymentEvent(i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := sync.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func deploymentEvent(id int) core.Event {
	payload := json.RawMessage(fmt.Sprintf(
		`{"repository":"github/docs","deployment_id":%d,"environment":"preview","sha":"077df53b0a1887aa6e941814e384fba00bceae8b","url":"https://example.test/deploy/%d","created_at":"2026-08-26T00:00:00Z","updated_at":"2026-08-26T00:01:00Z"}`,
		id, id))
	sum := sha256.Sum256(payload)
	return core.Event{
		ID:               core.EventID{byte(id)},
		Source:           core.SourceGitHub,
		NativeID:         fmt.Sprintf("deployment-%d", id),
		Kind:             core.KindGitHubDeployment,
		OccurredAt:       time.Unix(int64(id), 0).UTC(),
		RecordedAt:       time.Unix(int64(id), 0).UTC(),
		Payload:          payload,
		ContentSHA:       hex.EncodeToString(sum[:]),
		ConnectorVersion: "twin-replay-test/1",
	}
}

func ledgerSeqs(t *testing.T, ledger *store.Store) []int64 {
	t.Helper()
	var seqs []int64
	if err := ledger.ReadEvents(context.Background(), store.Filter{}, func(record store.Record) error {
		seqs = append(seqs, record.Seq)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return seqs
}

func equalInt64(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fieldValue(output, key string) string {
	idx := strings.Index(output, key)
	if idx < 0 {
		return ""
	}
	rest := output[idx+len(key):]
	if end := strings.IndexAny(rest, " \n"); end >= 0 {
		return rest[:end]
	}
	return rest
}
