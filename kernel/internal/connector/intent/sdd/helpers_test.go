package sdd

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

var timeOne = time.Unix(100, 0).UTC()

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func testIDs(t *testing.T) *core.IDGenerator {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{4}, 1<<16)), Now: func() time.Time { return time.Unix(2, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

type memAppender struct{ events []core.Event }

func (a *memAppender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	for _, old := range a.events {
		if old.IdempotencyKey() == e.IdempotencyKey() {
			return store.Record{Event: old}, false, nil
		}
	}
	a.events = append(a.events, e)
	return store.Record{Event: e}, true, nil
}
