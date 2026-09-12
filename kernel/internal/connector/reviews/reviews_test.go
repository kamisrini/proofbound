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

func TestParseAcceptsEmptyFindingsScalar(t *testing.T) {
	a := artifact("docs/verification/verdicts/empty.md", "ACCEPTABLE")
	text := string(a.Bytes)
	start := strings.Index(text, "findings:\n")
	end := strings.Index(text[start:], "artifact_path:") + start
	text = text[:start] + "findings: []\n" + text[end:]
	text = strings.Replace(text, regexpArtifactSHA(text), strings.Repeat("0", 64), 1)
	digest := ArtifactSHA([]byte(text))
	text = strings.Replace(text, strings.Repeat("0", 64), digest, 1)
	if _, err := Parse(a.Path, []byte(text)); err != nil {
		t.Fatal(err)
	}
}

func regexpArtifactSHA(text string) string {
	marker := "artifact_sha: "
	start := strings.Index(text, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	return text[start : start+64]
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

func obligationArtifact(t *testing.T, mutate func(*ObligationVerdict)) Artifact {
	t.Helper()
	base := readFixture(t, "obligation-verdict-v2-valid.md", "docs/verification/verdicts/p5-obligation-verdict-round1.md")
	v, err := ParseObligationVerdict(base.Path, base.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&v)
	v.ArtifactSHA256 = strings.Repeat("0", 64)
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	base.Bytes = append(append([]byte("---\n"), raw...), []byte("\n---\n\nGenerated hostile fixture.\n")...)
	v.ArtifactSHA256 = ArtifactSHA256(base.Bytes)
	raw, err = json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	base.Bytes = append(append([]byte("---\n"), raw...), []byte("\n---\n\nGenerated hostile fixture.\n")...)
	return base
}

func requirementReviewArtifact(t *testing.T, mutate func(*RequirementReview)) Artifact {
	t.Helper()
	base := readFixture(t, "requirement-review-valid.md", "docs/verification/verdicts/p5-requirement-review-round1.md")
	v, err := ParseRequirementReview(base.Path, base.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&v)
	v.ArtifactSHA256 = strings.Repeat("0", 64)
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	base.Bytes = append(append([]byte("---\n"), raw...), []byte("\n---\n\nGenerated hostile fixture.\n")...)
	v.ArtifactSHA256 = ArtifactSHA256(base.Bytes)
	raw, err = json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	base.Bytes = append(append([]byte("---\n"), raw...), []byte("\n---\n\nGenerated hostile fixture.\n")...)
	return base
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

func TestObligationVerdictClosedValuesAreAccepted(t *testing.T) {
	for _, status := range []string{"ACCEPTABLE", "NEEDS_WORK"} {
		t.Run("status "+status, func(t *testing.T) {
			a := obligationArtifact(t, func(v *ObligationVerdict) { v.Status = status })
			if _, err := ParseObligationVerdict(a.Path, a.Bytes); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, outcome := range []string{"SATISFIED", "NOT_SATISFIED", "INCONCLUSIVE"} {
		t.Run("outcome "+outcome, func(t *testing.T) {
			a := obligationArtifact(t, func(v *ObligationVerdict) {
				v.Status = "NEEDS_WORK"
				v.Obligations[0].Outcome = outcome
				if outcome != "SATISFIED" {
					v.Obligations[0].EvidenceEventIDs = nil
				}
			})
			if _, err := ParseObligationVerdict(a.Path, a.Bytes); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestObligationVerdictEvidenceValidation(t *testing.T) {
	a := readFixture(t, "obligation-verdict-v2-valid.md", "docs/verification/verdicts/p5-obligation-verdict-round1.md")
	bad := mutateJSONArtifact(t, a, `"evidence_event_ids":["01ARZ3NDEKTSV4RRFFQ69G5FAV"]`, `"evidence_event_ids":[]`)
	if _, err := ParseObligationVerdict(bad.Path, bad.Bytes); err == nil {
		t.Fatal("satisfied outcome without evidence accepted")
	}
}

func TestObligationVerdictRejectsEveryInvalidIdentityAndReference(t *testing.T) {
	badDigest := strings.Repeat("b", 64)
	cases := map[string]func(*ObligationVerdict){
		"schema":                  func(v *ObligationVerdict) { v.Schema = "proofbound.obligation-verdict.v3" },
		"verdict id":              func(v *ObligationVerdict) { v.VerdictID = "" },
		"status":                  func(v *ObligationVerdict) { v.Status = "MAYBE" },
		"reviewer":                func(v *ObligationVerdict) { v.DeclaredReviewer = " \t" },
		"commit":                  func(v *ObligationVerdict) { v.ReviewedCommit = "bad" },
		"artifact path":           func(v *ObligationVerdict) { v.ArtifactPath = "../outside.md" },
		"ci kind":                 func(v *ObligationVerdict) { v.ChangeIntent.RecordKind = "requirement" },
		"ci relation":             func(v *ObligationVerdict) { v.ChangeIntent.Relation = "claims" },
		"ci source":               func(v *ObligationVerdict) { v.ChangeIntent.Source = "bad source" },
		"ci id":                   func(v *ObligationVerdict) { v.ChangeIntent.RecordID = "" },
		"ci digest":               func(v *ObligationVerdict) { v.ChangeIntent.ArtifactSHA256 = "bad" },
		"ci obligations":          func(v *ObligationVerdict) { v.ChangeIntent.ObligationIDs = []string{"O-1"} },
		"no requirements":         func(v *ObligationVerdict) { v.Requirements = nil },
		"no outcomes":             func(v *ObligationVerdict) { v.Obligations = nil },
		"requirement kind":        func(v *ObligationVerdict) { v.Requirements[0].RecordKind = "change_intent" },
		"requirement relation":    func(v *ObligationVerdict) { v.Requirements[0].Relation = "implements" },
		"requirement source":      func(v *ObligationVerdict) { v.Requirements[0].Source = "bad source" },
		"requirement id":          func(v *ObligationVerdict) { v.Requirements[0].RecordID = "" },
		"requirement digest":      func(v *ObligationVerdict) { v.Requirements[0].ArtifactSHA256 = "bad" },
		"requirement obligations": func(v *ObligationVerdict) { v.Requirements[0].ObligationIDs = []string{"O-1"} },
		"duplicate requirement":   func(v *ObligationVerdict) { v.Requirements = append(v.Requirements, v.Requirements[0]) },
		"untargeted source":       func(v *ObligationVerdict) { v.Obligations[0].Source = "intent.specdir" },
		"untargeted id":           func(v *ObligationVerdict) { v.Obligations[0].RequirementID = "BR-other-acde12" },
		"untargeted digest":       func(v *ObligationVerdict) { v.Obligations[0].ArtifactSHA256 = badDigest },
		"bad outcome digest":      func(v *ObligationVerdict) { v.Obligations[0].ArtifactSHA256 = "bad" },
		"bad obligation id":       func(v *ObligationVerdict) { v.Obligations[0].ObligationID = "" },
		"bad outcome":             func(v *ObligationVerdict) { v.Obligations[0].Outcome = "MAYBE" },
		"duplicate outcome":       func(v *ObligationVerdict) { v.Obligations = append(v.Obligations[:1], v.Obligations...) },
		"bad event id":            func(v *ObligationVerdict) { v.Obligations[0].EvidenceEventIDs[0] = "bad" },
		"duplicate event id": func(v *ObligationVerdict) {
			id := v.Obligations[0].EvidenceEventIDs[0]
			v.Obligations[0].EvidenceEventIDs = []string{id, id}
		},
		"invalid finding id": func(v *ObligationVerdict) { v.Findings = []Finding{{FindingID: "", Severity: "LOW"}} },
		"invalid severity":   func(v *ObligationVerdict) { v.Findings = []Finding{{FindingID: "F-1", Severity: "MAYBE"}} },
		"invalid defect commit": func(v *ObligationVerdict) {
			v.Findings = []Finding{{FindingID: "F-1", Severity: "LOW", DefectCommit: "bad"}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			a := obligationArtifact(t, mutate)
			if _, err := ParseObligationVerdict(a.Path, a.Bytes); err == nil {
				t.Fatal("invalid obligation verdict accepted")
			}
		})
	}

	a := obligationArtifact(t, func(v *ObligationVerdict) { v.ArtifactSHA256 = badDigest })
	a.Bytes = []byte(strings.Replace(string(a.Bytes), `"artifact_sha256":"`+v2SelfDigest(t, a)+`"`, `"artifact_sha256":"`+badDigest+`"`, 1))
	if _, err := ParseObligationVerdict(a.Path, a.Bytes); err == nil {
		t.Fatal("wrong self digest accepted")
	}
}

func v2SelfDigest(t *testing.T, a Artifact) string {
	t.Helper()
	var v ObligationVerdict
	if err := json.Unmarshal(bytes.Split(a.Bytes, []byte("\n"))[1], &v); err != nil {
		t.Fatal(err)
	}
	return v.ArtifactSHA256
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

func TestRequirementReviewClosedOutcomesAreAccepted(t *testing.T) {
	for _, outcome := range []string{"VERIFIABLE", "AMBIGUOUS", "UNTESTABLE", "CONTRADICTORY"} {
		t.Run(outcome, func(t *testing.T) {
			a := requirementReviewArtifact(t, func(v *RequirementReview) { v.Outcomes[0].Outcome = outcome })
			if _, err := ParseRequirementReview(a.Path, a.Bytes); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRequirementReviewRejectsEveryInvalidIdentityAndOutcome(t *testing.T) {
	cases := map[string]func(*RequirementReview){
		"schema":                func(v *RequirementReview) { v.Schema = "proofbound.requirement-review.v2" },
		"review id":             func(v *RequirementReview) { v.ReviewID = "" },
		"reviewer":              func(v *RequirementReview) { v.DeclaredReviewer = " \t" },
		"artifact path":         func(v *RequirementReview) { v.ArtifactPath = "../outside.md" },
		"reference kind":        func(v *RequirementReview) { v.Requirement.RecordKind = "change_intent" },
		"reference relation":    func(v *RequirementReview) { v.Requirement.Relation = "evaluates" },
		"reference source":      func(v *RequirementReview) { v.Requirement.Source = "bad source" },
		"reference id":          func(v *RequirementReview) { v.Requirement.RecordID = "" },
		"reference digest":      func(v *RequirementReview) { v.Requirement.ArtifactSHA256 = "bad" },
		"reference obligations": func(v *RequirementReview) { v.Requirement.ObligationIDs = []string{"O-1"} },
		"no outcomes":           func(v *RequirementReview) { v.Outcomes = nil },
		"obligation id":         func(v *RequirementReview) { v.Outcomes[0].ObligationID = "" },
		"duplicate outcome":     func(v *RequirementReview) { v.Outcomes = append(v.Outcomes[:1], v.Outcomes...) },
		"outcome":               func(v *RequirementReview) { v.Outcomes[0].Outcome = "MAYBE" },
		"multiline finding":     func(v *RequirementReview) { v.Outcomes[0].Finding = "line one\nline two" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			a := requirementReviewArtifact(t, mutate)
			if _, err := ParseRequirementReview(a.Path, a.Bytes); err == nil {
				t.Fatal("invalid requirement review accepted")
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

func TestStrictJSONFrontMatterRejectsEachMalformedBoundary(t *testing.T) {
	path := "docs/verification/verdicts/test.md"
	for name, data := range map[string][]byte{
		"utf8":            {0xff},
		"short":           []byte("---\n{}"),
		"open delimiter":  []byte("--\n{}\n---\n"),
		"close delimiter": []byte("---\n{}\n--\n"),
		"not object":      []byte("---\n[]\n---\n"),
		"outer space":     []byte("---\n {}\n---\n"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := strictJSONFrontMatter(path, data); err == nil {
				t.Fatal("malformed JSON front matter accepted")
			}
		})
	}
	if _, err := strictJSONFrontMatter("../outside.md", []byte("---\n{}\n---\n")); err == nil {
		t.Fatal("hostile path accepted")
	}
}

func TestDecodeExactOrderedRejectsMalformedValue(t *testing.T) {
	var target map[string]any
	if err := decodeExactOrdered([]byte(`{"schema":`), []string{"schema"}, &target); err == nil {
		t.Fatal("malformed JSON accepted")
	}
	var typed struct {
		Schema string `json:"schema"`
	}
	if err := decodeExactOrdered([]byte(`{"schema":1}`), []string{"schema"}, &typed); err == nil {
		t.Fatal("wrong JSON value type accepted")
	}
}

func TestLegacyMultipleFindingsValidateEachEntry(t *testing.T) {
	valid := []string{
		"  - finding_id: F-1",
		"    severity: LOW",
		"  - finding_id: F-2",
		"    severity: HIGH",
	}
	if findings, err := parseFindings(valid); err != nil || len(findings) != 2 {
		t.Fatalf("findings=%+v err=%v", findings, err)
	}
	invalidFirst := append([]string(nil), valid...)
	invalidFirst[1] = "    severity: MAYBE"
	if _, err := parseFindings(invalidFirst); err == nil {
		t.Fatal("invalid non-final finding accepted")
	}
	if _, err := parseFindings([]string{"  - finding_id: F-1", "    severity: LOW", "    defect_commit:"}); err != nil {
		t.Fatalf("empty optional defect commit rejected: %v", err)
	}
}

func TestLegacyVerdictRejectsEveryMalformedBoundary(t *testing.T) {
	a := artifact("docs/verification/verdicts/a.md", "ACCEPTABLE")
	for name, text := range map[string]string{
		"short":                   "---\n",
		"open delimiter":          strings.Replace(string(a.Bytes), "---\n", "--\n", 1),
		"empty front matter":      "---\n---\n",
		"missing colon":           strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status ACCEPTABLE", 1),
		"space in key":            strings.Replace(string(a.Bytes), "status: ACCEPTABLE", " status: ACCEPTABLE", 1),
		"empty key":               strings.Replace(string(a.Bytes), "status: ACCEPTABLE", ": ACCEPTABLE", 1),
		"empty value":             strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status:", 1),
		"no value space":          strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status:ACCEPTABLE", 1),
		"two value spaces":        strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status:  ACCEPTABLE", 1),
		"trailing value":          strings.Replace(string(a.Bytes), "status: ACCEPTABLE", "status: ACCEPTABLE ", 1),
		"findings scalar":         strings.Replace(string(a.Bytes), "findings:\n", "findings: nope\n", 1),
		"bad finding indent":      strings.Replace(string(a.Bytes), "    severity: MED", "   severity: MED", 1),
		"finding no colon":        strings.Replace(string(a.Bytes), "    severity: MED", "    severity MED", 1),
		"finding key space":       strings.Replace(string(a.Bytes), "    severity: MED", "     severity: MED", 1),
		"finding no value space":  strings.Replace(string(a.Bytes), "    severity: MED", "    severity:MED", 1),
		"duplicate finding field": strings.Replace(string(a.Bytes), "    severity: MED", "    severity: MED\n    severity: LOW", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(a.Path, []byte(text)); err == nil {
				t.Fatal("malformed legacy verdict accepted")
			}
		})
	}
	for _, field := range []string{"schema", "verdict_id", "status", "reviewed_commit", "findings", "artifact_path", "artifact_sha"} {
		t.Run("missing "+field, func(t *testing.T) {
			lines := strings.Split(string(a.Bytes), "\n")
			filtered := lines[:0]
			for _, line := range lines {
				if !strings.HasPrefix(line, field+":") {
					filtered = append(filtered, line)
				}
			}
			if _, err := Parse(a.Path, []byte(strings.Join(filtered, "\n"))); err == nil {
				t.Fatal("legacy verdict with missing field accepted")
			}
		})
	}
}

func TestLegacyValidationClosedRegistriesAndPaths(t *testing.T) {
	for _, severity := range []string{"HIGH", "MED", "LOW"} {
		if err := validateFinding(Finding{FindingID: "F-1", Severity: severity}); err != nil {
			t.Fatalf("severity %s: %v", severity, err)
		}
	}
	for name, finding := range map[string]Finding{
		"id":       {Severity: "LOW"},
		"severity": {FindingID: "F-1", Severity: "MAYBE"},
		"commit":   {FindingID: "F-1", Severity: "LOW", DefectCommit: "bad"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateFinding(finding); err == nil {
				t.Fatal("invalid finding accepted")
			}
		})
	}
	for name, path := range map[string]string{
		"prefix":    "elsewhere/a.md",
		"suffix":    "docs/verification/verdicts/a.txt",
		"traversal": "docs/verification/verdicts/../a.md",
		"backslash": `docs/verification/verdicts/a\\b.md`,
	} {
		t.Run(name, func(t *testing.T) {
			if validPath(path) {
				t.Fatal("invalid path accepted")
			}
		})
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

func TestNewRejectsTypedNilReaderAndSyncRejectsInvalidState(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	var nilReader *reader
	if _, err := New(&Deps{Reader: nilReader, IDs: ids(t), Logger: log}); err == nil {
		t.Fatal("typed-nil reader accepted")
	}
	valid := connector(t, &reader{})
	for name, c := range map[string]*Connector{
		"nil":    nil,
		"reader": {ids: valid.ids, now: valid.now},
		"ids":    {reader: valid.reader, now: valid.now},
		"now":    {reader: valid.reader, ids: valid.ids},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Sync(context.Background(), &appender{}); err == nil {
				t.Fatal("invalid connector accepted")
			}
		})
	}
	var nilAppender *appender
	if _, err := valid.Sync(context.Background(), nilAppender); err == nil {
		t.Fatal("typed-nil appender accepted")
	}
	c, err := New(&Deps{Reader: &reader{}, IDs: ids(t), Logger: log})
	if err != nil || c.now == nil {
		t.Fatalf("default clock not installed: %v", err)
	}
}
