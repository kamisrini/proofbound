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
func TestContentSHA_Shape(t *testing.T) {
	s, err := ContentSHA(json.RawMessage(`{"ok":true}`))
	if err != nil || len(s) != ContentSHALen || strings.ToLower(s) != s {
		t.Fatalf("%q: %v", s, err)
	}
}
func TestKind_RegisteredOnlyForRegistryMembers(t *testing.T) {
	if !KindCommitRecorded.Registered() || Kind("unknown").Registered() {
		t.Fatal("registry mismatch")
	}
}
func TestSource_WellFormed(t *testing.T) {
	if Source("bad-value").WellFormed() || !Source("external_1").WellFormed() {
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
func TestIDGenerator_RequiresEntropy(t *testing.T) {
	if _, err := NewIDGenerator(IDGeneratorConfig{}); !errors.Is(err, ErrInvalidConfig) {
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
func TestIdempotencyKey_RevisionSemantics(t *testing.T) {
	a := IdempotencyKey{SourceGit, "x", "a"}
	b := IdempotencyKey{SourceGit, "x", "b"}
	if !a.IsRevisionOf(b) || a.SameSubject(b) == false {
		t.Fatal("revision mismatch")
	}
}
