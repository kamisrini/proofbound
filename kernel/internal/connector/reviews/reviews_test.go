package reviews

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
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

func TestSyncDistinguishesDocumentaryArtifacts(t *testing.T) {
	documentary := Artifact{Path: "docs/verification/verdicts/p5-adjudication-round1.md", Bytes: []byte("# Adjudication\n\nDocumentary evidence.\n")}
	r := &reader{artifacts: []Artifact{documentary}}
	result, err := connector(t, r).Sync(context.Background(), &appender{})
	if err != nil || result.Listed != 1 || result.Documentary != 1 || result.Appended != 0 || string(result.Cursor) != `["docs/verification/verdicts/p5-adjudication-round1.md"]` {
		t.Fatalf("result=%+v error=%v", result, err)
	}

	for name, candidate := range map[string]Artifact{
		"front matter": {Path: documentary.Path, Bytes: []byte("---\nbroken\n")},
		"schema line":  {Path: documentary.Path, Bytes: []byte("schema: broken\n")},
		"bad utf8":     {Path: documentary.Path, Bytes: []byte{0xff}},
		"bad path":     {Path: "../outside.md", Bytes: documentary.Bytes},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := connector(t, &reader{artifacts: []Artifact{candidate}}).Sync(context.Background(), &appender{})
			if err == nil || result.Malformed != 1 || result.Documentary != 0 {
				t.Fatalf("result=%+v error=%v", result, err)
			}
		})
	}
}

func readFixture(t *testing.T, name, path string) Artifact {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return Artifact{Path: path, Bytes: data}
}
func mutateJSONArtifact(t *testing.T, a Artifact, old, new string) Artifact {
	t.Helper()
	a.Bytes = []byte(strings.Replace(string(a.Bytes), old, new, 1))
	needle := []byte(`"artifact_sha256":"`)
	start := bytes.LastIndex(a.Bytes, needle)
	if start < 0 {
		t.Fatal("self digest missing")
	}
	start += len(needle)
	copy(a.Bytes[start:start+64], strings.Repeat("0", 64))
	digest := ArtifactSHA256(a.Bytes)
	copy(a.Bytes[start:start+64], digest)
	return a
}

func TestSchemaDispatchAndV1Compatibility(t *testing.T) {
	v1 := artifact("docs/verification/verdicts/v1.md", "ACCEPTABLE")
	v2 := readFixture(t, "obligation-verdict-v2-valid.md", "docs/verification/verdicts/p5-obligation-verdict-round1.md")
	rr := readFixture(t, "requirement-review-valid.md", "docs/verification/verdicts/p5-requirement-review-round1.md")
	app := &appender{}
	result, err := connector(t, &reader{artifacts: []Artifact{rr, v2, v1}}).Sync(context.Background(), app)
	if err != nil || result.Appended != 3 || len(app.events) != 3 {
		t.Fatalf("result=%+v events=%d err=%v", result, len(app.events), err)
	}
	kinds := map[core.Kind]int{}
	for _, event := range app.events {
		kinds[event.Kind]++
	}
	if kinds[core.KindReviewVerdict] != 2 || kinds[core.KindRequirementReview] != 1 {
		t.Fatalf("kinds=%v", kinds)
	}
	parsed, err := Parse(v1.Path, v1.Bytes)
	if err != nil || parsed.Schema != "vera.verdict.v1" {
		t.Fatalf("v1 changed: %+v %v", parsed, err)
	}
}

func TestObligationVerdictAggregation(t *testing.T) {
	a := readFixture(t, "obligation-verdict-v2-valid.md", "docs/verification/verdicts/p5-obligation-verdict-round1.md")
	if _, err := ParseObligationVerdict(a.Path, a.Bytes); err != nil {
		t.Fatal(err)
	}
	bad := mutateJSONArtifact(t, a, `"outcome":"SATISFIED"`, `"outcome":"INCONCLUSIVE"`)
	if _, err := ParseObligationVerdict(bad.Path, bad.Bytes); err == nil {
		t.Fatal("ACCEPTABLE with inconclusive obligation accepted")
	}
}

func TestObligationVerdictEvidenceValidation(t *testing.T) {
	a := readFixture(t, "obligation-verdict-v2-valid.md", "docs/verification/verdicts/p5-obligation-verdict-round1.md")
	bad := mutateJSONArtifact(t, a, `"evidence_event_ids":["01ARZ3NDEKTSV4RRFFQ69G5FAV"]`, `"evidence_event_ids":[]`)
	if _, err := ParseObligationVerdict(bad.Path, bad.Bytes); err == nil {
		t.Fatal("satisfied outcome without evidence accepted")
	}
}

func TestRequirementReviewValidation(t *testing.T) {
	a := readFixture(t, "requirement-review-valid.md", "docs/verification/verdicts/p5-requirement-review-round1.md")
	v, err := ParseRequirementReview(a.Path, a.Bytes)
	if err != nil || len(v.Outcomes) != 2 {
		t.Fatalf("review=%+v err=%v", v, err)
	}
	for name, bad := range map[string]Artifact{"outcome": mutateJSONArtifact(t, a, `"outcome":"VERIFIABLE"`, `"outcome":"MAYBE"`), "obligation": mutateJSONArtifact(t, a, `"obligation_id":"O-2"`, `"obligation_id":"O-1"`), "digest": mutateJSONArtifact(t, a, `"artifact_sha256":"913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562"`, `"artifact_sha256":"bad"`)} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRequirementReview(bad.Path, bad.Bytes); err == nil {
				t.Fatal("hostile review accepted")
			}
		})
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
