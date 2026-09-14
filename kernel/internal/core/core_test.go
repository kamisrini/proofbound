package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCanonicalize_KeyOrderAndWhitespaceIndependent(t *testing.T) {
	a, err := Canonicalize(json.RawMessage(`{ "b": 2, "a": 1 }`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Canonicalize(json.RawMessage(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("%s != %s", a, b)
	}
}
func TestCanonicalize_RejectsUnsafeNumbers(t *testing.T) {
	_, err := Canonicalize(json.RawMessage(`{"n":9007199254740993}`))
	if !errors.Is(err, ErrUnsafeNumber) || !errors.Is(err, ErrCanonicalJSON) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCanonicalize_RejectsDuplicateKeys(t *testing.T) {
	_, err := Canonicalize(json.RawMessage(`{"a":1,"a":2}`))
	if !errors.Is(err, ErrCanonicalJSON) {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestInspectJSON_RejectsNestedMalformedInput(t *testing.T) {
	for _, raw := range []string{`{"a"`, `{"a":[}`, `[}`} {
		if err := inspectJSON([]byte(raw)); err == nil {
			t.Fatalf("malformed JSON accepted: %q", raw)
		}
	}
}
func TestWalkJSON_PropagatesDecoderErrors(t *testing.T) {
	for _, raw := range []string{"", `{`, `{"a":`, `[`} {
		dec := json.NewDecoder(strings.NewReader(raw))
		if err := walkJSON(dec); err == nil {
			t.Fatalf("decoder error was swallowed: %q", raw)
		}
	}
	for _, raw := range []string{`{"outer":{"x":1,"x":2}}`, `[{"x":1,"x":2}]`} {
		dec := json.NewDecoder(strings.NewReader(raw))
		if err := walkJSON(dec); err == nil || !strings.Contains(err.Error(), "duplicate object key") {
			t.Fatalf("nested duplicate-key error was not propagated: %q: %v", raw, err)
		}
	}
}
func TestInspectValue_PropagatesNestedNumberErrors(t *testing.T) {
	bad := json.Number("9007199254740993")
	for _, value := range []any{[]any{bad}, map[string]any{"bad": bad}} {
		if err := inspectValue(value); !errors.Is(err, ErrUnsafeNumber) {
			t.Fatalf("nested unsafe number error lost: %v", err)
		}
	}
}
func TestContentSHA_Shape(t *testing.T) {
	s, err := ContentSHA(json.RawMessage(`{"ok":true}`))
	if err != nil || len(s) != ContentSHALen || strings.ToLower(s) != s {
		t.Fatalf("%q: %v", s, err)
	}
}
func TestKind_RegisteredOnlyForRegistryMembers(t *testing.T) {
	if !KindCommitRecorded.Registered() || !KindGitHubWorkflow.Registered() || !KindGitHubDeployment.Registered() || !KindBusinessDecision.Registered() || !KindRequirement.Registered() || !KindChangeIntent.Registered() || !KindRequirementReview.Registered() || Kind("unknown").Registered() {
		t.Fatal("registry mismatch")
	}
}
func TestSource_WellFormed(t *testing.T) {
	if Source("bad-value").WellFormed() || !Source("external_1").WellFormed() || !SourceGitHub.WellFormed() || !SourceIntentRecords.WellFormed() || !SourceIntentSpecdir.WellFormed() || Source("intent.bad.prefix").WellFormed() {
		t.Fatal("source shape mismatch")
	}
}
func TestEventID_TextRoundTrip(t *testing.T) {
	g, _ := NewIDGenerator(IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{1}, 64)), Now: func() time.Time { return time.Unix(1, 0) }})
	id, err := g.NewEventID()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseEventID(id.String())
	if err != nil || got != id {
		t.Fatalf("round trip: %v", err)
	}
}
func TestEventID_UnmarshalTextRejectsInvalidInput(t *testing.T) {
	var id EventID
	if err := id.UnmarshalText([]byte("not-an-event-id")); !errors.Is(err, ErrInvalidEventID) {
		t.Fatalf("invalid event ID accepted: %v", err)
	}
}
func TestIDGenerator_RequiresEntropy(t *testing.T) {
	if _, err := NewIDGenerator(IDGeneratorConfig{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatal(err)
	}
}
func TestIDGenerator_DefaultsClock(t *testing.T) {
	g, err := NewIDGenerator(IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{3}, 64))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.NewEventID(); err != nil {
		t.Fatal(err)
	}
}
func TestNewEvent_CanonicalPayloadAndValidation(t *testing.T) {
	g, _ := NewIDGenerator(IDGeneratorConfig{Entropy: bytes.NewReader(bytes.Repeat([]byte{2}, 64)), Now: time.Now})
	e, err := g.NewEvent(NewEventParams{Source: SourceGit, NativeID: "abc", Kind: KindCommitRecorded, OccurredAt: time.Now(), Payload: json.RawMessage(`{"z":1,"a":2}`), ConnectorVersion: "test/1"})
	if err != nil {
		t.Fatal(err)
	}
	if string(e.Payload) != `{"a":2,"z":1}` {
		t.Fatalf("payload=%s", e.Payload)
	}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_RejectsNativeIDControlCharacters(t *testing.T) {
	e := Event{ID: EventID{1}, Source: SourceGit, NativeID: "a\x00b", Kind: KindCommitRecorded, OccurredAt: time.Now(), RecordedAt: time.Now(), Payload: json.RawMessage(`{"ok":true}`), ContentSHA: strings.Repeat("a", 64), ConnectorVersion: "v1"}
	if !errors.Is(e.Validate(), ErrInvalidEvent) {
		t.Fatal("control character accepted")
	}
}
func TestValidate_RejectsEachMalformedField(t *testing.T) {
	base := Event{ID: EventID{1}, Source: SourceGit, NativeID: "native", Kind: KindCommitRecorded, OccurredAt: time.Unix(1, 0), RecordedAt: time.Unix(2, 0), Payload: json.RawMessage(`{"ok":true}`), ContentSHA: strings.Repeat("a", 64), ConnectorVersion: "core/1"}
	cases := map[string]func(*Event){
		"empty native ID":   func(e *Event) { e.NativeID = "" },
		"long native ID":    func(e *Event) { e.NativeID = strings.Repeat("x", 513) },
		"invalid native ID": func(e *Event) { e.NativeID = string([]byte{0xff}) },
		"spaced native ID":  func(e *Event) { e.NativeID = " native" },
		"empty kind":        func(e *Event) { e.Kind = "" },
		"unknown kind":      func(e *Event) { e.Kind = "unknown" },
		"empty payload":     func(e *Event) { e.Payload = nil },
		"large payload":     func(e *Event) { e.Payload = json.RawMessage(`{"x":"` + strings.Repeat("x", MaxPayloadBytes) + `"}`) },
		"invalid payload":   func(e *Event) { e.Payload = json.RawMessage(`not-json`) },
		"array payload":     func(e *Event) { e.Payload = json.RawMessage(`[]`) },
		"short content SHA": func(e *Event) { e.ContentSHA = "a" },
		"upper content SHA": func(e *Event) { e.ContentSHA = strings.Repeat("A", 64) },
		"empty connector":   func(e *Event) { e.ConnectorVersion = "" },
		"long connector":    func(e *Event) { e.ConnectorVersion = strings.Repeat("x", 65) },
		"spaced connector":  func(e *Event) { e.ConnectorVersion = "core 1" },
		"tabbed connector":  func(e *Event) { e.ConnectorVersion = "core\t1" },
		"control connector": func(e *Event) { e.ConnectorVersion = "core\x00" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			e := base
			mutate(&e)
			if !errors.Is(e.Validate(), ErrInvalidEvent) {
				t.Fatal("malformed field accepted")
			}
		})
	}
}
func TestIdempotencyKey_RevisionSemantics(t *testing.T) {
	a := IdempotencyKey{SourceGit, "x", "a"}
	b := IdempotencyKey{SourceGit, "x", "b"}
	if !a.IsRevisionOf(b) || a.SameSubject(b) == false {
		t.Fatal("revision mismatch")
	}
	if a.SameSubject(IdempotencyKey{SourceGitHub, "x", "a"}) || a.SameSubject(IdempotencyKey{SourceGit, "y", "a"}) {
		t.Fatal("different subjects were treated as equal")
	}
	if a.IsRevisionOf(IdempotencyKey{SourceGitHub, "x", "b"}) || a.IsRevisionOf(IdempotencyKey{SourceGit, "y", "b"}) {
		t.Fatal("different subjects were treated as revisions")
	}
}
