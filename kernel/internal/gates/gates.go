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

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "vera.gate.v1"

type Definition struct {
	Schema      string      `json:"schema"`
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Mode        string      `json:"mode"`
	Source      core.Source `json:"source"`
	Kind        core.Kind   `json:"kind"`
	Condition   Condition   `json:"condition"`
}

type Condition struct {
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
	decoder := json.NewDecoder(bytes.NewReader(data))
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
	if d.Schema != Version || d.ID == "" || d.Description == "" || d.Mode != "canary" || !d.Source.WellFormed() || !d.Kind.Registered() || d.Condition.Field == "" || len(d.Condition.Equals) == 0 || !json.Valid(d.Condition.Equals) {
		return errors.New("invalid gate definition")
	}
	if strings.ContainsAny(d.Condition.Field, ".[]\\\x00") {
		return errors.New("invalid condition field")
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
	value, ok := payload[condition.Field]
	if !ok {
		return StateBlocked, true, nil
	}
	if bytes.Equal(bytes.TrimSpace(value), bytes.TrimSpace(condition.Equals)) {
		return StatePass, false, nil
	}
	return StateBlocked, true, nil
}

// FilesystemGateDir is a small helper used by the CLI and keeps path handling
// out of the evaluator. It is intentionally not part of the gate contract.
func FilesystemGateDir(root string) fs.FS { return os.DirFS(filepath.Join(root, "gates")) }
