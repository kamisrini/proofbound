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
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "proofbound.gate.v1"

type Definition struct {
	Schema      string      `json:"schema"`
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Expires     string      `json:"expires"`
	Mode        string      `json:"mode"`
	Source      core.Source `json:"source"`
	Kind        core.Kind   `json:"kind"`
	Selector    *Predicate  `json:"selector,omitempty"`
	Condition   Condition   `json:"condition"`
	Rule        string      `json:"rule,omitempty"`
	ScopeCommit string      `json:"scope_commit,omitempty"`
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
		return definition, errors.New("trailing JSON")
	}
	if err := definition.validate(); err != nil {
		return definition, err
	}
	return definition, nil
}

func (d Definition) validate() error {
	if d.Schema != Version || d.ID == "" || d.Description == "" || d.Expires == "" || (d.Mode != "canary" && d.Mode != "enforce") {
		return errors.New("invalid gate definition")
	}
	if _, err := time.Parse("2006-01-02", d.Expires); err != nil {
		return errors.New("invalid gate expiry")
	}
	if d.Rule != "" {
		if d.Source != "" || d.Kind != "" || d.Selector != nil || d.Condition.Field != "" || len(d.Condition.Equals) != 0 || len(d.Condition.All) != 0 {
			return errors.New("semantic rule cannot include event predicates")
		}
		if d.Rule != "intent-reference-integrity" && d.Rule != "intent-verdict-integrity" && d.Rule != "intent-delivery-readiness" {
			return errors.New("unknown semantic gate rule")
		}
		if d.Rule == "intent-delivery-readiness" {
			if d.ScopeCommit != "HEAD" && !gitCommit(d.ScopeCommit) {
				return errors.New("delivery readiness requires HEAD or an exact commit")
			}
		} else if d.ScopeCommit != "" {
			return errors.New("integrity rule cannot have a commit scope")
		}
	} else {
		if d.ScopeCommit != "" || !d.Source.WellFormed() || !d.Kind.Registered() || !validCondition(d.Condition) {
			return errors.New("invalid gate condition")
		}
		if d.Selector != nil && !validPredicate(*d.Selector) {
			return errors.New("invalid gate selector")
		}
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
	if definition.Rule != "" {
		return evaluateIntentRule(ctx, s, definition)
	}
	var latest *store.Record
	if err := s.ReadEvents(ctx, store.Filter{Source: definition.Source, Kind: definition.Kind}, func(record store.Record) error {
		if definition.Selector != nil {
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(record.Event.Payload, &payload); err != nil {
				return fmt.Errorf("gate %s: payload: %w", definition.ID, err)
			}
			state, _, err := evaluatePredicate(payload, *definition.Selector)
			if err != nil {
				return err
			}
			if state != StatePass {
				return nil
			}
		}
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

func gitCommit(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func evaluateIntentRule(ctx context.Context, s *store.Store, d Definition) (Result, error) {
	result := Result{GateID: d.ID, State: StateUnknown}
	latest, err := latestIntentProof(ctx, s, d.Rule, d.ScopeCommit)
	if err != nil {
		return result, err
	}
	if latest != nil {
		result.EventID, result.Seq = latest.Event.ID.String(), latest.Seq
	}
	if err := projections.New().Apply(ctx, s); err != nil {
		if latest == nil {
			return result, fmt.Errorf("gate %s: projection failed without proof: %w", d.ID, err)
		}
		result.State, result.WouldBlock = StateBlocked, true
		return result, nil
	}
	if latest == nil {
		return result, nil
	}
	if d.Rule == "intent-reference-integrity" || d.Rule == "intent-verdict-integrity" {
		result.State = StatePass
		return result, nil
	}
	var targets, satisfied, reviewed int
	err = s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `WITH targets AS (SELECT ci.commit_sha,t.requirement_source,t.requirement_id,t.requirement_artifact_sha256,t.obligation_id,t.intent_source,t.intent_id,t.intent_artifact_sha256 FROM commit_intents_view ci JOIN intent_targets_view t ON t.intent_source='intent.'||ci.provider AND t.intent_id=ci.intent_id AND t.intent_artifact_sha256=ci.intent_artifact_sha256 WHERE ci.commit_sha=$1), states AS (SELECT t.*, EXISTS(SELECT 1 FROM obligation_verdicts_view ov WHERE ov.commit_sha=t.commit_sha AND ov.intent_source=t.intent_source AND ov.intent_id=t.intent_id AND ov.intent_artifact_sha256=t.intent_artifact_sha256 AND ov.requirement_source=t.requirement_source AND ov.requirement_id=t.requirement_id AND ov.requirement_artifact_sha256=t.requirement_artifact_sha256 AND ov.obligation_id=t.obligation_id AND ov.outcome='SATISFIED') AS satisfied, EXISTS(SELECT 1 FROM requirement_reviews_view rr WHERE rr.requirement_source=t.requirement_source AND rr.requirement_id=t.requirement_id AND rr.requirement_artifact_sha256=t.requirement_artifact_sha256 AND rr.obligation_id=t.obligation_id AND rr.outcome='VERIFIABLE') AS reviewed FROM targets t) SELECT count(*),count(*) FILTER(WHERE satisfied),count(*) FILTER(WHERE reviewed) FROM states`, d.ScopeCommit).Scan(&targets, &satisfied, &reviewed)
	})
	if err != nil {
		return result, err
	}
	if targets > 0 && satisfied == targets && reviewed == targets {
		result.State = StatePass
	} else {
		result.State, result.WouldBlock = StateBlocked, true
	}
	return result, nil
}

func latestIntentProof(ctx context.Context, s *store.Store, rule, commit string) (*store.Record, error) {
	var latest *store.Record
	err := s.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		match := false
		switch rule {
		case "intent-reference-integrity":
			match = (record.Event.Source == core.SourceIntentRecords || record.Event.Source == core.SourceIntentSpecdir || record.Event.Source == core.SourceGit && record.Event.Kind == core.KindCommitRecorded)
		case "intent-verdict-integrity":
			match = record.Event.Source == core.SourceReviews && (record.Event.Kind == core.KindReviewVerdict || record.Event.Kind == core.KindRequirementReview)
		case "intent-delivery-readiness":
			match = record.Event.Source == core.SourceGit && record.Event.Kind == core.KindCommitRecorded && record.Event.NativeID == commit
		}
		if match {
			copy := record
			latest = &copy
		}
		return nil
	})
	return latest, err
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
