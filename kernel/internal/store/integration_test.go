//go:build integration

package store

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

func TestStoreAppendDuplicateRevisionAndRead(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}
	s, err := Open(context.Background(), Config{Root: t.TempDir(), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	nativeID := "commit-" + time.Now().UTC().Format("20060102150405.000000000")
	e, err := g.NewEvent(core.NewEventParams{Source: core.SourceGit, NativeID: nativeID, Kind: core.KindCommitRecorded, OccurredAt: time.Unix(99, 0), Payload: json.RawMessage(`{"sha":"a"}`), ConnectorVersion: "test"})
	if err != nil {
		t.Fatal(err)
	}
	sy, err := s.BeginSync(context.Background(), "git")
	if err != nil {
		t.Fatal(err)
	}
	first, inserted, err := sy.Append(context.Background(), e)
	if err != nil || !inserted {
		t.Fatalf("first append: %+v %v %v", first, inserted, err)
	}
	dup, inserted, err := sy.Append(context.Background(), e)
	if err != nil || inserted || dup.Seq != first.Seq || dup.Event.ID != first.Event.ID {
		t.Fatalf("duplicate: %+v %v %v", dup, inserted, err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	var got []Record
	if err := s.ReadEvents(context.Background(), Filter{SinceSeq: first.Seq - 1}, func(r Record) error { got = append(got, r); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Seq != first.Seq {
		t.Fatalf("read=%+v", got)
	}
}

func TestStoreRevisionKindConflictAndFilters(t *testing.T) {
	s := integrationStore(t)
	defer s.Close()
	e := integrationEvent(t, "revision")
	sy, err := s.BeginSync(context.Background(), "git")
	if err != nil {
		t.Fatal(err)
	}
	first, inserted, err := sy.Append(context.Background(), e)
	if err != nil || !inserted {
		t.Fatalf("first=%v %v", err, inserted)
	}
	revision := integrationEvent(t, "revision")
	revision.NativeID = e.NativeID
	revision.Payload = json.RawMessage(`{"sha":"changed"}`)
	revision.ContentSHA, err = core.ContentSHA(revision.Payload)
	if err != nil {
		t.Fatal(err)
	}
	second, inserted, err := sy.Append(context.Background(), revision)
	if err != nil || !inserted || second.Seq <= first.Seq {
		t.Fatalf("revision=%+v %v %v", second, inserted, err)
	}
	conflict := e
	conflict.Kind = core.KindCheckRun
	if _, _, err := sy.Append(context.Background(), conflict); !errors.Is(err, ErrKindConflict) {
		t.Fatalf("kind conflict=%v", err)
	}
	if err := sy.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.ReadEvents(context.Background(), Filter{Source: core.SourceGit, Kind: core.KindCommitRecorded, SinceSeq: first.Seq - 1, Limit: 1}, func(Record) error { count++; return nil }); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("filter count=%d", count)
	}
}

func TestStoreWithTxCommitRollbackAndStop(t *testing.T) {
	s := integrationStore(t)
	defer s.Close()
	if err := s.WithTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, "CREATE TEMP TABLE vera_tx_probe (n integer)")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("stop")
	if err := s.WithTx(context.Background(), func(context.Context, *Tx) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatalf("rollback error=%v", err)
	}
	if err := s.ReadEvents(context.Background(), Filter{Limit: 1}, func(Record) error { return ErrStopIteration }); err != nil {
		t.Fatalf("stop iteration=%v", err)
	}
}

func integrationStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required")
	}
	s, err := Open(context.Background(), Config{Root: t.TempDir(), DatabaseURL: url})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func integrationEvent(t *testing.T, label string) core.Event {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	e, err := g.NewEvent(core.NewEventParams{Source: core.SourceGit, NativeID: "integration-" + label + "-" + time.Now().UTC().Format("150405.000000000"), Kind: core.KindCommitRecorded, OccurredAt: time.Now(), Payload: json.RawMessage(`{"sha":"original"}`), ConnectorVersion: "integration"})
	if err != nil {
		t.Fatal(err)
	}
	return e
}
