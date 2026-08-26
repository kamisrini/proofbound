package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
	"github.com/oklog/ulid/v2"
)

type Kind string

const (
	KindCommitRecorded   Kind = "commit.recorded"
	KindCheckRun         Kind = "check.run"
	KindSessionObserved  Kind = "session.observed"
	KindReviewVerdict    Kind = "review.verdict"
	KindGitHubWorkflow   Kind = "github.workflow_run"
	KindGitHubDeployment Kind = "github.deployment"
)

func (k Kind) Registered() bool {
	switch k {
	case KindCommitRecorded, KindCheckRun, KindSessionObserved, KindReviewVerdict, KindGitHubWorkflow, KindGitHubDeployment:
		return true
	}
	return false
}
func Kinds() []Kind {
	return []Kind{KindCheckRun, KindCommitRecorded, KindGitHubDeployment, KindGitHubWorkflow, KindReviewVerdict, KindSessionObserved}
}

type Source string

const (
	SourceGit      Source = "git"
	SourceChecks   Source = "checks"
	SourceSessions Source = "sessions"
	SourceReviews  Source = "reviews"
	SourceGitHub   Source = "github"
)

var sourceRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

func (s Source) WellFormed() bool { return sourceRE.MatchString(string(s)) }

const ContentSHALen = 64
const MaxPayloadBytes = 1 << 20

var (
	ErrInvalidEvent   = errors.New("core: invalid event")
	ErrCanonicalJSON  = errors.New("core: canonical json")
	ErrUnsafeNumber   = errors.New("core: json number outside exact integer range")
	ErrInvalidEventID = errors.New("core: invalid event id")
	ErrInvalidConfig  = errors.New("core: invalid config")
)

func Canonicalize(raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil, ErrCanonicalJSON
	}
	if err := inspectJSON(raw); err != nil {
		return nil, err
	}
	out, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCanonicalJSON, err)
	}
	return out, nil
}

func inspectJSON(raw []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := rejectDuplicateKeys(dec); err != nil {
		return fmt.Errorf("%w: %v", ErrCanonicalJSON, err)
	}
	dec = json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("%w: %v", ErrCanonicalJSON, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return ErrCanonicalJSON
	}
	return inspectValue(v)
}

func rejectDuplicateKeys(dec *json.Decoder) error {
	return walkJSON(dec)
}

func walkJSON(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			seen := map[string]struct{}{}
			for dec.More() {
				key, err := dec.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return errors.New("object key is not a string")
				}
				if _, exists := seen[name]; exists {
					return fmt.Errorf("duplicate object key %q", name)
				}
				seen[name] = struct{}{}
				if err := walkJSON(dec); err != nil {
					return err
				}
			}
			_, err = dec.Token()
			return err
		case '[':
			for dec.More() {
				if err := walkJSON(dec); err != nil {
					return err
				}
			}
			_, err = dec.Token()
			return err
		}
	}
	return nil
}
func inspectValue(v any) error {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
			return fmt.Errorf("%w: %w", ErrCanonicalJSON, ErrUnsafeNumber)
		}
		if i, err := x.Int64(); err == nil && (i > 1<<53 || i < -(1<<53)) {
			return fmt.Errorf("%w: %w", ErrCanonicalJSON, ErrUnsafeNumber)
		}
	case []any:
		for _, e := range x {
			if err := inspectValue(e); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, e := range x {
			if err := inspectValue(e); err != nil {
				return err
			}
		}
	}
	return nil
}
func ContentSHA(raw json.RawMessage) (string, error) {
	b, err := Canonicalize(raw)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

type EventID [16]byte

func ParseEventID(s string) (EventID, error) {
	id, err := ulid.ParseStrict(s)
	if err != nil {
		return EventID{}, fmt.Errorf("%w: %v", ErrInvalidEventID, err)
	}
	var out EventID
	copy(out[:], id[:])
	return out, nil
}
func (id EventID) String() string { var u ulid.ULID; copy(u[:], id[:]); return u.String() }
func (id EventID) IsZero() bool   { return id == EventID{} }
func (id EventID) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrInvalidEventID
	}
	return []byte(id.String()), nil
}
func (id *EventID) UnmarshalText(b []byte) error {
	parsed, err := ParseEventID(string(b))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

type IDGeneratorConfig struct {
	Entropy io.Reader
	Now     func() time.Time
}
type IDGenerator struct {
	entropy io.Reader
	now     func() time.Time
	mu      sync.Mutex
}

func NewIDGenerator(cfg IDGeneratorConfig) (*IDGenerator, error) {
	if cfg.Entropy == nil {
		return nil, ErrInvalidConfig
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &IDGenerator{entropy: ulid.Monotonic(cfg.Entropy, 0), now: cfg.Now}, nil
}
func (g *IDGenerator) NewEventID() (EventID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	u, err := ulid.New(ulid.Timestamp(g.now()), g.entropy)
	if err != nil {
		return EventID{}, fmt.Errorf("%w: %v", ErrInvalidEventID, err)
	}
	var id EventID
	copy(id[:], u[:])
	return id, nil
}

type NewEventParams struct {
	Source           Source
	NativeID         string
	Kind             Kind
	OccurredAt       time.Time
	Payload          json.RawMessage
	ConnectorVersion string
}
type Event struct {
	ID               EventID         `json:"event_id"`
	Source           Source          `json:"source"`
	NativeID         string          `json:"native_id"`
	Kind             Kind            `json:"kind"`
	OccurredAt       time.Time       `json:"occurred_at"`
	RecordedAt       time.Time       `json:"recorded_at"`
	Payload          json.RawMessage `json:"payload"`
	ContentSHA       string          `json:"content_sha"`
	ConnectorVersion string          `json:"connector_version"`
}

func (g *IDGenerator) NewEvent(p NewEventParams) (Event, error) {
	id, err := g.NewEventID()
	if err != nil {
		return Event{}, err
	}
	payload, err := Canonicalize(p.Payload)
	if err != nil {
		return Event{}, err
	}
	hash := sha256.Sum256(payload)
	sha := hex.EncodeToString(hash[:])
	e := Event{ID: id, Source: p.Source, NativeID: p.NativeID, Kind: p.Kind, OccurredAt: p.OccurredAt.UTC().Truncate(time.Microsecond), RecordedAt: g.now().UTC().Truncate(time.Microsecond), Payload: payload, ContentSHA: sha, ConnectorVersion: p.ConnectorVersion}
	if err := e.Validate(); err != nil {
		return Event{}, err
	}
	return e, nil
}

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }
func (e *ValidationError) Unwrap() error { return ErrInvalidEvent }
func invalid(field, reason string) error { return &ValidationError{Field: field, Reason: reason} }
func (e Event) Validate() error {
	var es []error
	if e.ID.IsZero() {
		es = append(es, invalid("event_id", "missing"))
	}
	if !e.Source.WellFormed() {
		es = append(es, invalid("source", "malformed"))
	}
	if e.NativeID == "" || len(e.NativeID) > 512 || !utf8.ValidString(e.NativeID) || strings.TrimSpace(e.NativeID) != e.NativeID || strings.IndexFunc(e.NativeID, func(r rune) bool { return r < 32 || r == 127 }) >= 0 {
		es = append(es, invalid("native_id", "malformed"))
	}
	if e.Kind == "" || !e.Kind.Registered() {
		es = append(es, invalid("kind", "unregistered"))
	}
	if e.OccurredAt.IsZero() {
		es = append(es, invalid("occurred_at", "missing"))
	}
	if e.RecordedAt.IsZero() {
		es = append(es, invalid("recorded_at", "missing"))
	}
	if len(e.Payload) == 0 || len(e.Payload) > MaxPayloadBytes || !json.Valid(e.Payload) || e.Payload[0] != '{' {
		es = append(es, invalid("payload", "must be a non-empty object"))
	}
	if len(e.ContentSHA) != ContentSHALen || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(e.ContentSHA) {
		es = append(es, invalid("content_sha", "malformed"))
	}
	if e.ConnectorVersion == "" || len(e.ConnectorVersion) > 64 || strings.IndexFunc(e.ConnectorVersion, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r < 32 }) >= 0 {
		es = append(es, invalid("connector_version", "malformed"))
	}
	return errors.Join(es...)
}

type IdempotencyKey struct {
	Source     Source
	NativeID   string
	ContentSHA string
}

func (e Event) IdempotencyKey() IdempotencyKey {
	return IdempotencyKey{e.Source, e.NativeID, e.ContentSHA}
}
func (k IdempotencyKey) String() string {
	return string(k.Source) + "/" + k.NativeID + "/" + k.ContentSHA
}
func (k IdempotencyKey) SameSubject(o IdempotencyKey) bool {
	return k.Source == o.Source && k.NativeID == o.NativeID
}
func (k IdempotencyKey) IsRevisionOf(o IdempotencyKey) bool {
	return k.SameSubject(o) && k.ContentSHA != o.ContentSHA
}
