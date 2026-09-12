package gates

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadDirRequiresReadableDirectoryAndLoadsDefinitions(t *testing.T) {
	dir := t.TempDir()
	definition := []byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`)
	if err := os.WriteFile(filepath.Join(dir, "x.yaml"), definition, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("not a gate"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadDir(dir)
	if err != nil || len(loaded) != 1 || loaded[0].ID != "x" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	if _, err := LoadDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing gate directory accepted")
	}
}

func TestLoadRejectsInvalidDefinitions(t *testing.T) {
	valid := []byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`)
	if _, err := Parse(valid); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{
		[]byte(`{"schema":"proofbound.gate.v2"}`),
		[]byte(strings.Replace(string(valid), `"proofbound.gate.v1"`, `"proofbound.gate.v2"`, 1)),
		[]byte(strings.Replace(string(valid), `"id":"x"`, `"id":""`, 1)),
		[]byte(strings.Replace(string(valid), `"description":"d"`, `"description":""`, 1)),
		[]byte(strings.Replace(string(valid), `"expires":"2099-01-01"`, `"expires":""`, 1)),
		[]byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"experimental","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`),
		[]byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"not-a-date","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`),
		[]byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code"}}`),
		[]byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}} trailing`),
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

func TestEvaluatePayloadRequiresAllPredicates(t *testing.T) {
	condition := Condition{All: []Predicate{{Field: "command", Equals: json.RawMessage(`"make index-check"`)}, {Field: "exit_code", Equals: json.RawMessage("0")}}}
	state, blocked, err := evaluatePayload(map[string]json.RawMessage{"command": json.RawMessage(`"make index-check"`), "exit_code": json.RawMessage("1")}, condition)
	if err != nil || state != StateBlocked || !blocked {
		t.Fatalf("state=%s blocked=%v err=%v", state, blocked, err)
	}
	state, blocked, err = evaluatePayload(map[string]json.RawMessage{"command": json.RawMessage(`"make index-check"`), "exit_code": json.RawMessage("0")}, condition)
	if err != nil || state != StatePass || blocked {
		t.Fatalf("state=%s blocked=%v err=%v", state, blocked, err)
	}
}

func TestCompoundConditionRejectsScalarFields(t *testing.T) {
	predicates := []Predicate{{Field: "exit_code", Equals: json.RawMessage("0")}}
	for name, condition := range map[string]Condition{
		"field":  {Field: "exit_code", All: predicates},
		"equals": {Equals: json.RawMessage("0"), All: predicates},
	} {
		t.Run(name, func(t *testing.T) {
			if validCondition(condition) {
				t.Fatal("mixed scalar and compound condition accepted")
			}
		})
	}
}

func TestPredicateValidationRejectsEachInvalidDimension(t *testing.T) {
	if !validPredicate(Predicate{Field: "exit_code", Equals: json.RawMessage("0")}) {
		t.Fatal("valid predicate rejected")
	}
	for name, predicate := range map[string]Predicate{
		"empty field":  {Equals: json.RawMessage("0")},
		"nested field": {Field: "exit.code", Equals: json.RawMessage("0")},
		"empty value":  {Field: "exit_code"},
		"invalid JSON": {Field: "exit_code", Equals: json.RawMessage("not-json")},
	} {
		t.Run(name, func(t *testing.T) {
			if validPredicate(predicate) {
				t.Fatal("invalid predicate accepted")
			}
		})
	}
}

func TestIntentRuleDefinitionsFailClosed(t *testing.T) {
	valid := []byte(`{"schema":"proofbound.gate.v1","id":"intent","description":"d","expires":"2026-12-31","mode":"canary","rule":"intent-reference-integrity"}`)
	if _, err := Parse(valid); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"unknown":         []byte(strings.Replace(string(valid), "intent-reference-integrity", "intent-maybe", 1)),
		"source":          []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","source":"git"`, 1)),
		"kind":            []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","kind":"commit.recorded"`, 1)),
		"selector":        []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","selector":{"field":"sha","equals":"x"}`, 1)),
		"condition field": []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","condition":{"field":"sha"}`, 1)),
		"condition value": []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","condition":{"equals":"x"}`, 1)),
		"condition all":   []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-reference-integrity","condition":{"all":[{"field":"sha","equals":"x"}]}`, 1)),
		"scope":           []byte(strings.Replace(string(valid), `"rule":"intent-reference-integrity"`, `"rule":"intent-delivery-readiness","scope_commit":"branch"`, 1)),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(data); err == nil {
				t.Fatal("invalid semantic gate accepted")
			}
		})
	}
}

func TestGitCommitAcceptsOnlySupportedLowerHexLengths(t *testing.T) {
	for _, length := range []int{40, 64} {
		if !gitCommit(strings.Repeat("a", length)) {
			t.Fatalf("valid length %d rejected", length)
		}
	}
	for _, length := range []int{39, 41, 63, 65} {
		if gitCommit(strings.Repeat("a", length)) {
			t.Fatalf("invalid length %d accepted", length)
		}
	}
	if gitCommit(strings.Repeat("A", 40)) {
		t.Fatal("uppercase commit accepted")
	}
}

func TestParseIsReadOnly(t *testing.T) {
	data := []byte(`{"schema":"proofbound.gate.v1","id":"x","description":"d","expires":"2099-01-01","mode":"canary","source":"checks","kind":"check.run","condition":{"field":"exit_code","equals":0}}`)
	want := append([]byte(nil), data...)
	if _, err := Parse(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, want) {
		t.Fatal("Parse mutated its input")
	}
}

func TestEnforceRequiresPromotion(t *testing.T) {
	definition := Definition{Schema: Version, ID: "x", Description: "d", Expires: "2099-01-01", Mode: "canary", Source: "checks", Kind: "check.run", Condition: Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
	if err := definition.EnforceReady(); err == nil {
		t.Fatal("canary definition accepted for enforcement")
	}
	definition.Mode = "enforce"
	if err := definition.EnforceReady(); err != nil {
		t.Fatal(err)
	}
}

func TestEnforceRejectsExpiredDefinition(t *testing.T) {
	d := Definition{Expires: "2026-08-26"}
	if !d.Expired(time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expiry date should include the whole calendar day")
	}
	if d.Expired(time.Date(2026, 8, 26, 23, 59, 59, 0, time.UTC)) {
		t.Fatal("future expiry unexpectedly expired")
	}
	d = Definition{Schema: Version, ID: "x", Description: "d", Expires: "2026-08-25", Mode: "enforce", Source: "checks", Kind: "check.run", Condition: Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
	if err := d.EnforceReady(); err == nil {
		t.Fatal("expired definition accepted for enforcement")
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
