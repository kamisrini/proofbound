//go:build integration

package store

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
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
