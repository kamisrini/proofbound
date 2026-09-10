package intent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "intent/1"

var (
	ErrUnmappable = errors.New("intent: source cannot map without changing the canonical model")
	digestRE      = regexp.MustCompile(`^[0-9a-f]{64}$`)
	recordIDRE    = regexp.MustCompile(`^(BD|BR|CI)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`)
	obligationRE  = regexp.MustCompile(`^O-[A-Z0-9][A-Z0-9-]{0,31}$`)
)

type Revision struct {
	Source         core.Source
	Kind           core.Kind
	NativeID       string
	ArtifactPath   string
	ArtifactSHA256 string
	ObservedAt     time.Time
	Payload        json.RawMessage
}

type Provider interface {
	Name() string
	Revisions(context.Context) ([]Revision, error)
}

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Reference struct {
	RecordKind     string   `json:"record_kind"`
	Source         string   `json:"source"`
	RecordID       string   `json:"record_id"`
	ArtifactSHA256 string   `json:"artifact_sha256"`
	Relation       string   `json:"relation"`
	ObligationIDs  []string `json:"obligation_ids,omitempty"`
}

type Obligation struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	State     string `json:"state"`
}

type BusinessDecision struct {
	Schema           string      `json:"schema"`
	DecisionID       string      `json:"decision_id"`
	Status           string      `json:"status"`
	Outcome          string      `json:"outcome"`
	DeclaredOwner    string      `json:"declared_owner"`
	DeclaredApprover string      `json:"declared_approver"`
	DecidedAt        string      `json:"decided_at"`
	Supersedes       []Reference `json:"supersedes"`
	ArtifactPath     string      `json:"artifact_path"`
	ArtifactSHA256   string      `json:"artifact_sha256"`
}

type Requirement struct {
	Schema         string       `json:"schema"`
	RequirementID  string       `json:"requirement_id"`
	Status         string       `json:"status"`
	AuthorizedBy   []Reference  `json:"authorized_by"`
	DeclaredOwner  string       `json:"declared_owner"`
	Obligations    []Obligation `json:"obligations"`
	Supersedes     []Reference  `json:"supersedes"`
	ArtifactPath   string       `json:"artifact_path"`
	ArtifactSHA256 string       `json:"artifact_sha256"`
}

type ChangeIntent struct {
	Schema          string      `json:"schema"`
	IntentID        string      `json:"intent_id"`
	Status          string      `json:"status"`
	DeclaredSponsor string      `json:"declared_sponsor"`
	Targets         []Reference `json:"targets"`
	ConstrainedBy   []Reference `json:"constrained_by"`
	Supersedes      []Reference `json:"supersedes"`
	ArtifactPath    string      `json:"artifact_path"`
	ArtifactSHA256  string      `json:"artifact_sha256"`
}

type Deps struct {
	Providers []Provider
	IDs       *core.IDGenerator
	Logger    *slog.Logger
}

type ProviderResult struct{ Listed, Appended, Existing int }

type Result struct {
	Listed, Appended, Existing int
	ByProvider                 map[string]ProviderResult
	Cursor                     json.RawMessage
}

type Connector struct {
	providers map[string]Provider
	ids       *core.IDGenerator
	logger    *slog.Logger
}

type batchItem struct {
	provider string
	revision Revision
}

func New(d *Deps) (*Connector, error) {
	if d == nil || d.IDs == nil || d.Logger == nil || len(d.Providers) == 0 {
		return nil, errors.New("intent connector: providers, IDs, and Logger are required")
	}
	providers := make(map[string]Provider, len(d.Providers))
	for _, p := range d.Providers {
		if p == nil || reflect.ValueOf(p).Kind() == reflect.Pointer && reflect.ValueOf(p).IsNil() {
			return nil, errors.New("intent connector: nil provider")
		}
		name := p.Name()
		if name != "records" && name != "specdir" {
			return nil, fmt.Errorf("intent connector: production provider %q is not configured", name)
		}
		if _, exists := providers[name]; exists {
			return nil, fmt.Errorf("intent connector: duplicate provider %q", name)
		}
		providers[name] = p
	}
	return &Connector{providers: providers, ids: d.IDs, logger: d.Logger}, nil
}

func (c *Connector) Sync(ctx context.Context, selection string, app Appender) (Result, error) {
	result := Result{ByProvider: map[string]ProviderResult{}}
	if c == nil || c.ids == nil || len(c.providers) == 0 {
		return result, errors.New("intent connector: connector is not initialized")
	}
	if app == nil || reflect.ValueOf(app).Kind() == reflect.Pointer && reflect.ValueOf(app).IsNil() {
		return result, errors.New("intent connector: appender is required")
	}
	names := make([]string, 0, len(c.providers))
	if selection == "all" {
		for name := range c.providers {
			names = append(names, name)
		}
		sort.Strings(names)
	} else if _, ok := c.providers[selection]; ok {
		names = append(names, selection)
	} else {
		return result, fmt.Errorf("intent connector: unknown provider %q", selection)
	}
	var batch []batchItem
	for _, name := range names {
		revisions, err := c.providers[name].Revisions(ctx)
		if err != nil {
			return result, fmt.Errorf("intent connector: provider %s: %w", name, err)
		}
		result.ByProvider[name] = ProviderResult{Listed: len(revisions)}
		result.Listed += len(revisions)
		for _, rev := range revisions {
			batch = append(batch, batchItem{name, rev})
		}
	}
	sort.Slice(batch, func(i, j int) bool {
		a, b := batch[i].revision, batch[j].revision
		ak := string(a.Source) + "\x00" + string(a.Kind) + "\x00" + a.NativeID + "\x00" + a.ArtifactSHA256 + "\x00" + a.ArtifactPath
		bk := string(b.Source) + "\x00" + string(b.Kind) + "\x00" + b.NativeID + "\x00" + b.ArtifactSHA256 + "\x00" + b.ArtifactPath
		return ak < bk
	})
	if err := validateBatch(batch); err != nil {
		return result, err
	}
	batch, err := orderBatch(batch)
	if err != nil {
		return result, err
	}
	for _, entry := range batch {
		rev := entry.revision
		e, err := c.ids.NewEvent(core.NewEventParams{Source: rev.Source, NativeID: rev.NativeID, Kind: rev.Kind, OccurredAt: rev.ObservedAt, Payload: rev.Payload, ConnectorVersion: Version})
		if err != nil {
			return result, fmt.Errorf("intent connector: %s: %w", rev.NativeID, err)
		}
		_, inserted, err := app.Append(ctx, e)
		if err != nil {
			return result, fmt.Errorf("intent connector: append %s: %w", rev.NativeID, err)
		}
		pr := result.ByProvider[entry.provider]
		if inserted {
			result.Appended++
			pr.Appended++
		} else {
			result.Existing++
			pr.Existing++
		}
		result.ByProvider[entry.provider] = pr
	}
	cursor, _ := json.Marshal(names)
	result.Cursor = cursor
	return result, nil
}

func ValidateRevision(rev Revision) error {
	if rev.Source != core.SourceIntentRecords && rev.Source != core.SourceIntentSpecdir {
		return fmt.Errorf("intent: invalid source %q", rev.Source)
	}
	if rev.Kind != core.KindBusinessDecision && rev.Kind != core.KindRequirement && rev.Kind != core.KindChangeIntent {
		return fmt.Errorf("intent: invalid kind %q", rev.Kind)
	}
	if !recordIDRE.MatchString(rev.NativeID) || !digestRE.MatchString(rev.ArtifactSHA256) || rev.ArtifactPath == "" || rev.ObservedAt.IsZero() {
		return errors.New("intent: invalid revision envelope")
	}
	canonical, err := core.Canonicalize(rev.Payload)
	if err != nil || string(canonical) != string(rev.Payload) {
		return errors.New("intent: payload is not canonical JSON")
	}
	var path, digest, id string
	switch rev.Kind {
	case core.KindBusinessDecision:
		var v BusinessDecision
		if err := strict(rev.Payload, &v); err != nil {
			return err
		}
		if v.Schema != "proofbound.business-decision.v1" || !oneOf(v.Status, "proposed", "accepted", "superseded", "withdrawn") || empty(v.Outcome, v.DeclaredOwner, v.DeclaredApprover) {
			return errors.New("intent: invalid business decision")
		}
		if _, err := time.Parse(time.RFC3339, v.DecidedAt); err != nil {
			return errors.New("intent: invalid decided_at")
		}
		id, path, digest = v.DecisionID, v.ArtifactPath, v.ArtifactSHA256
		if err := validateReferences(v.Supersedes); err != nil {
			return err
		}
	case core.KindRequirement:
		var v Requirement
		if err := strict(rev.Payload, &v); err != nil {
			return err
		}
		if v.Schema != "proofbound.requirement.v1" || !oneOf(v.Status, "proposed", "active", "superseded", "retired") || v.DeclaredOwner == "" || len(v.Obligations) == 0 {
			return errors.New("intent: invalid requirement")
		}
		id, path, digest = v.RequirementID, v.ArtifactPath, v.ArtifactSHA256
		if err := validateReferences(append(append([]Reference{}, v.AuthorizedBy...), v.Supersedes...)); err != nil {
			return err
		}
		if err := validateObligations(v.Obligations); err != nil {
			return err
		}
	case core.KindChangeIntent:
		var v ChangeIntent
		if err := strict(rev.Payload, &v); err != nil {
			return err
		}
		if v.Schema != "proofbound.change-intent.v1" || !oneOf(v.Status, "proposed", "accepted", "superseded", "withdrawn") || v.DeclaredSponsor == "" || len(v.Targets) == 0 {
			return errors.New("intent: invalid change intent")
		}
		id, path, digest = v.IntentID, v.ArtifactPath, v.ArtifactSHA256
		refs := append(append(append([]Reference{}, v.Targets...), v.ConstrainedBy...), v.Supersedes...)
		if err := validateReferences(refs); err != nil {
			return err
		}
	}
	if id != rev.NativeID || path != rev.ArtifactPath || digest != rev.ArtifactSHA256 {
		return errors.New("intent: revision envelope does not bind payload")
	}
	return nil
}

func strict(raw []byte, target any) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("intent: payload: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("intent: trailing JSON")
	}
	return nil
}
func empty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}
func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
func validateObligations(values []Obligation) error {
	last := ""
	for _, o := range values {
		if !obligationRE.MatchString(o.ID) || strings.TrimSpace(o.Statement) == "" || !oneOf(o.State, "active", "retired") || o.ID <= last {
			return errors.New("intent: invalid obligation")
		}
		last = o.ID
	}
	return nil
}
func validateReferences(refs []Reference) error {
	last := ""
	for _, ref := range refs {
		key := ref.Source + "\x00" + ref.RecordKind + "\x00" + ref.RecordID + "\x00" + ref.ArtifactSHA256 + "\x00" + ref.Relation
		if key <= last || !core.Source(ref.Source).WellFormed() || !recordIDRE.MatchString(ref.RecordID) || !digestRE.MatchString(ref.ArtifactSHA256) {
			return errors.New("intent: invalid reference")
		}
		if !oneOf(ref.RecordKind, "business_decision", "requirement", "change_intent", "verification_decision") || !oneOf(ref.Relation, "authorizes", "implements", "modifies", "repairs", "retires", "supersedes", "constrained_by", "reviews") {
			return errors.New("intent: invalid reference type")
		}
		if ref.RecordKind == "requirement" && oneOf(ref.Relation, "implements", "modifies", "repairs", "retires") {
			if len(ref.ObligationIDs) == 0 {
				return errors.New("intent: target has no obligations")
			}
			for i, id := range ref.ObligationIDs {
				if !obligationRE.MatchString(id) || i > 0 && id <= ref.ObligationIDs[i-1] {
					return errors.New("intent: invalid target obligation")
				}
			}
		} else if len(ref.ObligationIDs) != 0 {
			return errors.New("intent: non-target reference has obligations")
		}
		last = key
	}
	return nil
}
func references(raw []byte, kind core.Kind) []Reference {
	switch kind {
	case core.KindBusinessDecision:
		var v BusinessDecision
		_ = json.Unmarshal(raw, &v)
		return v.Supersedes
	case core.KindRequirement:
		var v Requirement
		_ = json.Unmarshal(raw, &v)
		return append(v.AuthorizedBy, v.Supersedes...)
	case core.KindChangeIntent:
		var v ChangeIntent
		_ = json.Unmarshal(raw, &v)
		return append(append(v.Targets, v.ConstrainedBy...), v.Supersedes...)
	}
	return nil
}
func validateBatch(batch []batchItem) error {
	type key struct{ source, id, digest string }
	available := make(map[key]struct{}, len(batch))
	loadedSources := map[string]bool{}
	for _, entry := range batch {
		if err := ValidateRevision(entry.revision); err != nil {
			return fmt.Errorf("intent connector: %s: %w", entry.revision.NativeID, err)
		}
		if string(entry.revision.Source) != "intent."+entry.provider {
			return fmt.Errorf("intent connector: provider %s spoofed source %s", entry.provider, entry.revision.Source)
		}
		k := key{string(entry.revision.Source), entry.revision.NativeID, entry.revision.ArtifactSHA256}
		if _, duplicate := available[k]; duplicate {
			return fmt.Errorf("intent connector: duplicate revision %s", entry.revision.NativeID)
		}
		available[k] = struct{}{}
		loadedSources[string(entry.revision.Source)] = true
	}
	for _, entry := range batch {
		for _, ref := range references(entry.revision.Payload, entry.revision.Kind) {
			if !strings.HasPrefix(ref.Source, "intent.") || !loadedSources[ref.Source] {
				continue
			}
			if _, ok := available[key{ref.Source, ref.RecordID, ref.ArtifactSHA256}]; !ok {
				return fmt.Errorf("intent connector: dangling exact revision %s:%s@%s", ref.Source, ref.RecordID, ref.ArtifactSHA256)
			}
		}
	}
	return nil
}

func orderBatch(batch []batchItem) ([]batchItem, error) {
	type key struct{ source, id, digest string }
	positions := map[key]int{}
	for i, entry := range batch {
		positions[key{string(entry.revision.Source), entry.revision.NativeID, entry.revision.ArtifactSHA256}] = i
	}
	done := map[int]bool{}
	out := make([]batchItem, 0, len(batch))
	for len(out) < len(batch) {
		progress := false
		for i, entry := range batch {
			if done[i] {
				continue
			}
			ready := true
			for _, ref := range references(entry.revision.Payload, entry.revision.Kind) {
				if dep, ok := positions[key{ref.Source, ref.RecordID, ref.ArtifactSHA256}]; ok && !done[dep] {
					ready = false
					break
				}
			}
			if ready {
				out = append(out, entry)
				done[i] = true
				progress = true
			}
		}
		if !progress {
			return nil, errors.New("intent connector: cyclic exact-revision relations")
		}
	}
	return out, nil
}
