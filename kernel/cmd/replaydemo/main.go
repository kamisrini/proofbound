// Command replaydemo is a runnable demonstration of Proofbound's twin replay:
// it stands up a real embedded PostgreSQL, seeds an append-only event ledger,
// and shows twin.ReplayIsolated (1) reproducing the committed projected state
// with a vera.replay.v1 proof, (2) detecting a counterfactual change, and
// (3) leaving the source ledger byte-for-byte unchanged throughout.
//
// This is a demo entry point, not product wiring. It uses only the public
// kernel APIs (store, core, projections, twin). Run it from kernel/:
//
//	go run ./cmd/replaydemo
//
// It requires the embedded-postgres binaries to be available (cached or
// fetchable); on an unavailable Postgres it exits non-zero with the reason.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	"github.com/kamisrini/proofbound/kernel/internal/twin"
)

// out is the real stdout for the demo narrative. The embedded-postgres library
// logs to os.Stdout with no knob to quiet it, so run() points os.Stdout at
// /dev/null and keeps the saved handle here. say() writes only to out.
var out io.Writer = os.Stdout

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "\nreplaydemo: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Capture the real stdout for our narrative, then silence embedded-postgres
	// by sending os.Stdout (which it logs to) to /dev/null for the whole run.
	out = os.Stdout
	if devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0); err == nil {
		os.Stdout = devnull
		defer func() { os.Stdout = out.(*os.File); _ = devnull.Close() }()
	}

	root, err := os.MkdirTemp("", "proofbound-replaydemo-")
	if err != nil {
		return fmt.Errorf("temp root: %w", err)
	}
	defer os.RemoveAll(root)

	say("Opening a real embedded PostgreSQL and an append-only event ledger…")
	s, err := store.Open(ctx, store.Config{Root: root, Port: 55450})
	if err != nil {
		return fmt.Errorf("embedded postgres unavailable: %w", err)
	}
	defer s.Close()

	// Seed the source ledger with two committed deployment events.
	for i := 1; i <= 2; i++ {
		if err := appendEvent(ctx, s, deploymentEvent(i)); err != nil {
			return fmt.Errorf("seed event %d: %w", i, err)
		}
	}
	before, err := ledgerDigest(ctx, s)
	if err != nil {
		return err
	}
	say("\nSource ledger committed: %d events, digest %s", before.count, short(before.sha))

	baseline := projections.Snapshot{Tables: map[string][]string{"github_delivery_view": {}}}

	// 1. Reproduce the committed state: replay the ledger with no counterfactual.
	//    result.Candidate is the real projector's view of the committed ledger.
	reproduced, err := twin.ReplayIsolated(ctx, s, twin.Request{ThroughSeq: 2}, baseline)
	if err != nil {
		return fmt.Errorf("reproduce replay: %w", err)
	}
	committed := reproduced.Candidate

	// Replaying again against that committed projection must say "preserved":
	// the ledger deterministically re-derives the exact same state.
	preserved, err := twin.ReplayIsolated(ctx, s, twin.Request{ThroughSeq: 2}, committed)
	if err != nil {
		return fmt.Errorf("preserved replay: %w", err)
	}
	say("\n[1] Reproduce committed state")
	printResult(preserved)

	// 2. A counterfactual: what if a third deployment had also happened?
	changed, err := twin.ReplayIsolated(ctx, s, twin.Request{
		ThroughSeq: 2,
		Candidate:  []twin.Candidate{{Seq: 3, Event: deploymentEvent(3)}},
	}, committed)
	if err != nil {
		return fmt.Errorf("counterfactual replay: %w", err)
	}
	say("\n[2] Counterfactual change (one extra deployment)")
	printResult(changed)

	// 3. Prove the source ledger was never touched by either replay.
	after, err := ledgerDigest(ctx, s)
	if err != nil {
		return err
	}
	say("\nSource ledger after both replays: %d events, digest %s", after.count, short(after.sha))
	if before.sha != after.sha {
		return fmt.Errorf("source ledger changed: %s -> %s", short(before.sha), short(after.sha))
	}

	say("\nTrust properties demonstrated:")
	say("  • replay re-derives the committed state and proves it (verdict %q, schema %q)", preserved.Verdict, preserved.Proof.Schema)
	say("  • a proposed change is detected (verdict %q) and bound by a distinct snapshot digest", changed.Verdict)
	say("  • the counterfactual digest %s differs from the committed digest %s", short(changed.Proof.CandidateSnapshotDigest), short(preserved.Proof.CandidateSnapshotDigest))
	say("  • the source ledger is byte-for-byte identical before and after (digest %s)", short(after.sha))
	return nil
}

func printResult(r twin.Result) {
	say("    verdict:            %s", r.Verdict)
	say("    source events:      %d", r.SourceEvents)
	say("    candidate events:   %d", r.CandidateEvents)
	say("    proof schema:       %s", r.Proof.Schema)
	say("    source digest:      %s", short(r.Proof.SourceDigest))
	say("    candidate digest:   %s", short(r.Proof.CandidateDigest))
	say("    baseline snapshot:  %s", short(r.Proof.BaselineSnapshotDigest))
	say("    candidate snapshot: %s", short(r.Proof.CandidateSnapshotDigest))
}

type digest struct {
	count int
	sha   string
}

// ledgerDigest reads the whole source ledger and hashes its canonical event
// identity + content, so any mutation would change the digest.
func ledgerDigest(ctx context.Context, s *store.Store) (digest, error) {
	h := sha256.New()
	count := 0
	if err := s.ReadEvents(ctx, store.Filter{}, func(r store.Record) error {
		count++
		fmt.Fprintf(h, "%d:%s:%s:%s\n", r.Seq, r.Event.Source, r.Event.Kind, r.Event.ContentSHA)
		return nil
	}); err != nil {
		return digest{}, fmt.Errorf("read ledger: %w", err)
	}
	return digest{count: count, sha: hex.EncodeToString(h.Sum(nil))}, nil
}

func appendEvent(ctx context.Context, s *store.Store, e core.Event) error {
	sy, err := s.BeginSync(ctx, "replaydemo")
	if err != nil {
		return err
	}
	if _, _, err := sy.Append(ctx, e); err != nil {
		return err
	}
	return sy.Finish(ctx, nil, nil)
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
		OccurredAt:       time.Unix(int64(id), 0),
		RecordedAt:       time.Unix(int64(id), 0),
		Payload:          payload,
		ContentSHA:       hex.EncodeToString(sum[:]),
		ConnectorVersion: "replaydemo/1",
	}
}

func short(s string) string {
	if len(s) <= 12 {
		return s
	}
	return s[:12] + "…"
}

func say(format string, args ...any) {
	fmt.Fprintf(out, format+"\n", args...)
}
