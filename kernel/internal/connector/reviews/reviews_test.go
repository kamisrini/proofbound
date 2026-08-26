package reviews

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type reader struct {
	artifacts []Artifact
	calls     int
}

func (r *reader) ReadCommittedVerdicts(context.Context) ([]Artifact, error) {
	r.calls++
	return append([]Artifact(nil), r.artifacts...), nil
}

type appender struct{ events []core.Event }

func (a *appender) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	for _, old := range a.events {
		if old.IdempotencyKey() == e.IdempotencyKey() {
			return store.Record{Event: old}, false, nil
		}
	}
	a.events = append(a.events, e)
	return store.Record{Event: e}, true, nil
}

const reviewed = "0123456789abcdef0123456789abcdef01234567"

func artifact(path, status string) Artifact {
	data := "---\n" +
		"schema: vera.verdict.v1\n" +
		"verdict_id: task8-current-round1\n" +
		"status: " + status + "\n" +
		"reviewed_commit: " + reviewed + "\n" +
		"findings:\n" +
		"  - finding_id: F-1\n" +
		"    severity: MED\n" +
		"    defect_commit: " + reviewed + "\n" +
		"artifact_path: " + path + "\n" +
		"artifact_sha: " + strings.Repeat("0", 64) + "\n" +
		"---\n\nA review.\n"
	digest := ArtifactSHA([]byte(data))
	data = strings.Replace(data, strings.Repeat("0", 64), digest, 1)
	return Artifact{Path: path, Bytes: []byte(data)}
}
func ids(t *testing.T) *core.IDGenerator {
	t.Helper()
	g, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{7}, 4096)), Now: func() time.Time { return time.Unix(100, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func connector(t *testing.T, r CommittedReader) *Connector {
	t.Helper()
	c, err := New(&Deps{Reader: r, IDs: ids(t), Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: func() time.Time { return time.Unix(200, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestParseStrictFrontMatter(t *testing.T) {
	a := artifact("docs/verification/verdicts/task8-current-round1.md", "ACCEPTABLE")
	v, err := Parse(a.Path, a.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if v.Schema != "vera.verdict.v1" || v.Status != "ACCEPTABLE" || len(v.Findings) != 1 || v.Findings[0].Severity != "MED" {
		t.Fatalf("parsed=%+v", v)
	}
	for name, data := range map[string][]byte{
		"unknown":    append([]byte("---\nunknown: x\n"), a.Bytes[strings.Index(string(a.Bytes), "---\n\n"):]...),
		"duplicate":  []byte(strings.Replace(string(a.Bytes), "status: ACCEPTABLE\n", "status: ACCEPTABLE\nstatus: NEEDS_WORK\n", 1)),
		"bad status": []byte(strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status: MAYBE", 1)),
		"bad utf8":   {0xff},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(a.Path, data); err == nil {
				t.Fatal("malformed artifact accepted")
			}
		})
	}
}

func TestSyncMintsReviewVerdictEvent(t *testing.T) {
	a := artifact("docs/verification/verdicts/task8-current-round1.md", "NEEDS_WORK")
	r := &reader{artifacts: []Artifact{a}}
	app := &appender{}
	result, err := connector(t, r).Sync(context.Background(), app)
	if err != nil || result.Appended != 1 || len(app.events) != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	e := app.events[0]
	if e.Source != core.SourceReviews || e.Kind != core.KindReviewVerdict || e.NativeID != "task8-current-round1" || e.ConnectorVersion != Version {
		t.Fatalf("event=%+v", e)
	}
	var got Verdict
	if err := json.Unmarshal(e.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "NEEDS_WORK" || got.ArtifactPath != a.Path || len(got.ArtifactSHA) != 64 {
		t.Fatalf("payload=%+v", got)
	}
}

func TestSyncIsIdempotentAndRevisionSafe(t *testing.T) {
	r := &reader{artifacts: []Artifact{artifact("docs/verification/verdicts/task8-current-round1.md", "ACCEPTABLE")}}
	c := connector(t, r)
	app := &appender{}
	first, err := c.Sync(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.Sync(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	if first.Appended != 1 || second.Existing != 1 || len(app.events) != 1 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	r.artifacts[0] = artifact("docs/verification/verdicts/task8-current-round1.md", "NEEDS_WORK")
	third, err := c.Sync(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	if third.Appended != 1 || len(app.events) != 2 {
		t.Fatalf("third=%+v events=%d", third, len(app.events))
	}
}

func TestSyncUsesOnlyInjectedCommittedReader(t *testing.T) {
	r := &reader{}
	result, err := connector(t, r).Sync(context.Background(), &appender{})
	if err != nil || result.Listed != 0 || r.calls != 1 {
		t.Fatalf("result=%+v calls=%d err=%v", result, r.calls, err)
	}
}

func TestSyncSortsAndFailsClosed(t *testing.T) {
	good := artifact("docs/verification/verdicts/a.md", "ACCEPTABLE")
	bad := artifact("docs/verification/verdicts/b.md", "NOPE")
	r := &reader{artifacts: []Artifact{bad, good}}
	result, err := connector(t, r).Sync(context.Background(), &appender{})
	if err == nil || result.Listed != 2 || result.Appended != 1 || result.Malformed != 1 || string(result.Cursor) != `["docs/verification/verdicts/a.md"]` {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestParseBindsPathAndDigest(t *testing.T) {
	a := artifact("docs/verification/verdicts/a.md", "ACCEPTABLE")
	if _, err := Parse("docs/verification/verdicts/other.md", a.Bytes); err == nil {
		t.Fatal("path mismatch accepted")
	}
	a.Bytes = append(a.Bytes, []byte("changed\n")...)
	if _, err := Parse(a.Path, a.Bytes); err == nil {
		t.Fatal("bad digest accepted")
	}
}

func TestNewRequiresDependencies(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := &reader{}
	for name, d := range map[string]*Deps{"nil": nil, "reader": {IDs: ids(t), Logger: log}, "ids": {Reader: r, Logger: log}, "logger": {Reader: r, IDs: ids(t)}} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(d); err == nil {
				t.Fatal("invalid dependencies accepted")
			}
		})
	}
}
