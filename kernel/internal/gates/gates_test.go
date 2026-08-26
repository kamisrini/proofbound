package gates

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLoadRejectsInvalidDefinitions(t *testing.T) {
	valid := []byte(`{"schema":"vera.gate.v1","id":"x","description":"d","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`)
	if _, err := Parse(valid); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{
		[]byte(`{"schema":"vera.gate.v2"}`),
		[]byte(`{"schema":"vera.gate.v1","id":"x","description":"d","mode":"experimental","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`),
		[]byte(`{"schema":"vera.gate.v1","id":"x","description":"d","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code"}}`),
		[]byte(`{"schema":"vera.gate.v1","id":"x","description":"d","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}} trailing`),
	} {
		if _, err := Parse(bad); err == nil {
			t.Fatal("invalid definition accepted")
		}
	}
}

func TestEvaluatePayloadPassesMatchingValue(t *testing.T) {
	condition := Condition{Field: "exit_code", Equals: json.RawMessage("0")}
	state, blocked, err := evaluatePayload(map[string]json.RawMessage{"exit_code": json.RawMessage("0")}, condition)
	if err != nil || state != StatePass || blocked {
		t.Fatalf("state=%s blocked=%v err=%v", state, blocked, err)
	}
}

func TestEvaluatePayloadBlocksMismatchAndMissing(t *testing.T) {
	condition := Condition{Field: "exit_code", Equals: json.RawMessage("0")}
	for _, tc := range []struct {
		name    string
		payload map[string]json.RawMessage
		want    State
	}{
		{"blocked", map[string]json.RawMessage{"exit_code": json.RawMessage("1")}, StateBlocked},
		{"missing", map[string]json.RawMessage{}, StateBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state, blocked, err := evaluatePayload(tc.payload, condition)
			if err != nil || state != tc.want || !blocked {
				t.Fatalf("state=%s blocked=%v err=%v", state, blocked, err)
			}
		})
	}
}

func TestParseIsReadOnly(t *testing.T) {
	data := []byte(`{"schema":"vera.gate.v1","id":"x","description":"d","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`)
	want := append([]byte(nil), data...)
	if _, err := Parse(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, want) {
		t.Fatal("Parse mutated its input")
	}
}

func TestEnforceRequiresPromotion(t *testing.T) {
	definition := Definition{Schema: Version, ID: "x", Description: "d", Mode: "canary", Source: "checks", Kind: "check.run", Condition: Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
	if err := definition.EnforceReady(); err == nil {
		t.Fatal("canary definition accepted for enforcement")
	}
	definition.Mode = "enforce"
	if err := definition.EnforceReady(); err != nil {
		t.Fatal(err)
	}
}

func TestEnforceRejectsNonPass(t *testing.T) {
	for _, state := range []State{StateBlocked, StateUnknown, State("INVALID"), State("")} {
		if err := Enforce(Result{GateID: "x", State: state}); err == nil {
			t.Fatalf("state %s was accepted", state)
		}
	}
	if err := Enforce(Result{GateID: "x", State: StatePass}); err != nil {
		t.Fatal(err)
	}
}

func TestRequireDefinitions(t *testing.T) {
	if err := RequireDefinitions(nil); err == nil {
		t.Fatal("empty definition set accepted")
	}
	if err := RequireDefinitions([]Definition{{ID: "x"}}); err != nil {
		t.Fatal(err)
	}
}
