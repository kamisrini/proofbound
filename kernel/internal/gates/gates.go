package gates

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "vera.gate.v1"

type Definition struct {
	Schema      string      `json:"schema"`
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Expires     string      `json:"expires"`
	Mode        string      `json:"mode"`
	Source      core.Source `json:"source"`
	Kind        core.Kind   `json:"kind"`
	Condition   Condition   `json:"condition"`
}

type Condition struct {
	Field  string          `json:"field"`
	Equals json.RawMessage `json:"equals"`
	All    []Predicate     `json:"all,omitempty"`
}

type Predicate struct {
	Field  string          `json:"field"`
	Equals json.RawMessage `json:"equals"`
}

type State string

const (
	StateUnknown State = "UNKNOWN"
	StatePass    State = "PASS"
	StateBlocked State = "BLOCKED"
)

type Result struct {
	GateID     string
	State      State
	EventID    string
	Seq        int64
	WouldBlock bool
}

func LoadDir(dir string) ([]Definition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			paths = append(paths, entry.Name())
		}
	}
	sort.Strings(paths)
	definitions := make([]Definition, 0, len(paths))
	seen := map[string]bool{}
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("gate %s: %w", name, err)
		}
		definition, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("gate %s: %w", name, err)
		}
		if seen[definition.ID] {
			return nil, fmt.Errorf("gate %s: duplicate id %q", name, definition.ID)
		}
		seen[definition.ID] = true
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func Parse(data []byte) (Definition, error) {
	var definition Definition
	canonical, err := core.Canonicalize(data)
	if err != nil {
		return definition, fmt.Errorf("gate JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&definition); err != nil {
		return definition, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return definition, errors.New("trailing JSON")
		}
		return definition, fmt.Errorf("trailing JSON: %w", err)
	}
	if err := definition.validate(); err != nil {
		return definition, err
	}
	return definition, nil
}

func (d Definition) validate() error {
	if d.Schema != Version || d.ID == "" || d.Description == "" || d.Expires == "" || (d.Mode != "canary" && d.Mode != "enforce") || !d.Source.WellFormed() || !d.Kind.Registered() {
		return errors.New("invalid gate definition")
	}
	if _, err := time.Parse("2006-01-02", d.Expires); err != nil {
		return errors.New("invalid gate expiry")
	}
	if !validCondition(d.Condition) {
		return errors.New("invalid gate condition")
	}
	return nil
}

func validPredicate(p Predicate) bool {
	return p.Field != "" && !strings.ContainsAny(p.Field, ".[]\\\x00") && len(p.Equals) > 0 && json.Valid(p.Equals)
}

func validCondition(c Condition) bool {
	if len(c.All) > 0 {
		if c.Field != "" || len(c.Equals) != 0 {
			return false
		}
		for _, p := range c.All {
			if !validPredicate(p) {
				return false
			}
		}
		return true
	}
	return validPredicate(Predicate{Field: c.Field, Equals: c.Equals})
}

func (d Definition) EnforceReady() error {
	if err := d.validate(); err != nil {
		return err
	}
	if d.Mode != "enforce" {
		return fmt.Errorf("gate %s is not promoted to enforce mode", d.ID)
	}
	if d.Expired(time.Now()) {
		return fmt.Errorf("gate %s expired on %s", d.ID, d.Expires)
	}
	return nil
}

func (d Definition) Expired(now time.Time) bool {
	if _, err := time.Parse("2006-01-02", d.Expires); err != nil {
		return false
	}
	return now.Format("2006-01-02") > d.Expires
}

func Enforce(result Result) error {
	if result.State != StatePass {
		return fmt.Errorf("gate %s is %s", result.GateID, result.State)
	}
	return nil
}

func RequireDefinitions(definitions []Definition) error {
	if len(definitions) == 0 {
		return errors.New("gate enforcement blocked: no gate definitions found")
	}
	return nil
}

func Evaluate(ctx context.Context, s *store.Store, definition Definition) (Result, error) {
	if err := definition.validate(); err != nil {
		return Result{}, err
	}
	result := Result{GateID: definition.ID, State: StateUnknown}
	var latest *store.Record
	if err := s.ReadEvents(ctx, store.Filter{Source: definition.Source, Kind: definition.Kind}, func(record store.Record) error {
		copy := record
		latest = &copy
		return nil
	}); err != nil {
		return Result{}, err
	}
	if latest == nil {
		return result, nil
	}
	result.EventID, result.Seq = latest.Event.ID.String(), latest.Seq
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(latest.Event.Payload, &payload); err != nil {
		return Result{}, fmt.Errorf("gate %s: payload: %w", definition.ID, err)
	}
	state, wouldBlock, err := evaluatePayload(payload, definition.Condition)
	if err != nil {
		return Result{}, err
	}
	result.State, result.WouldBlock = state, wouldBlock
	return result, nil
}

func evaluatePayload(payload map[string]json.RawMessage, condition Condition) (State, bool, error) {
	if len(condition.All) > 0 {
		for _, predicate := range condition.All {
			state, blocked, err := evaluatePredicate(payload, predicate)
			if err != nil || state != StatePass {
				return state, blocked, err
			}
		}
		return StatePass, false, nil
	}
	return evaluatePredicate(payload, Predicate{Field: condition.Field, Equals: condition.Equals})
}

func evaluatePredicate(payload map[string]json.RawMessage, predicate Predicate) (State, bool, error) {
	value, ok := payload[predicate.Field]
	if !ok {
		return StateBlocked, true, nil
	}
	if bytes.Equal(bytes.TrimSpace(value), bytes.TrimSpace(predicate.Equals)) {
		return StatePass, false, nil
	}
	return StateBlocked, true, nil
}

// FilesystemGateDir is a small helper used by the CLI and keeps path handling
// out of the evaluator. It is intentionally not part of the gate contract.
func FilesystemGateDir(root string) fs.FS { return os.DirFS(filepath.Join(root, "gates")) }
