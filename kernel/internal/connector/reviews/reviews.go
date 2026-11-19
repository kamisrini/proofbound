package reviews

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "reviews/1"

var (
	commitRE = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
	digestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
	idRE     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,127}$`)
)

type Artifact struct {
	Path  string
	Bytes []byte
}
type CommittedReader interface {
	ReadCommittedVerdicts(context.Context) ([]Artifact, error)
}
type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Finding struct {
	FindingID    string `json:"finding_id"`
	Severity     string `json:"severity"`
	DefectCommit string `json:"defect_commit,omitempty"`
}
type Verdict struct {
	Schema         string    `json:"schema"`
	VerdictID      string    `json:"verdict_id"`
	Status         string    `json:"status"`
	ReviewedCommit string    `json:"reviewed_commit"`
	Findings       []Finding `json:"findings"`
	ArtifactPath   string    `json:"artifact_path"`
	ArtifactSHA    string    `json:"artifact_sha"`
}

// LedgerView answers the order-only prerequisite questions the hold asks (R-INV-13). The CLI
// implements it over the store; it never reads tenant bytes.
type LedgerView interface {
	// Ledgered reports whether a reviews event of this payload schema and native id is ledgered.
	Ledgered(ctx context.Context, schema, nativeID string) (bool, error)
	// CommitClaimed reports whether a ledgered git commit event for sha carries exactly this ref.
	CommitClaimed(ctx context.Context, sha, provider, recordID, digest string) (bool, error)
	// RequirementLedgered reports whether this exact requirement revision is ledgered.
	RequirementLedgered(ctx context.Context, source core.Source, recordID, digest string) (bool, error)
}

// Hold reports one artifact that was not appended (R-INV-13), with its cause.
type Hold struct {
	Path  string
	Cause string
}

// ErrRevisionRefused is wrapped by the error Sync returns after processing every artifact when at
// least one was refused (R-INV-13).
var ErrRevisionRefused = errors.New("review artifact revision refused")

type Deps struct {
	Reader CommittedReader
	IDs    *core.IDGenerator
	Logger *slog.Logger
	Now    func() time.Time
	Ledger LedgerView
}
type Connector struct {
	reader CommittedReader
	ids    *core.IDGenerator
	logger *slog.Logger
	now    func() time.Time
	ledger LedgerView
}

type Result struct {
	Listed, Appended, Existing, Malformed, Documentary int
	Cursor                                             json.RawMessage
	Held, Refused                                      []Hold
}

func New(d *Deps) (*Connector, error) {
	if d == nil {
		return nil, errors.New("reviews connector: dependencies are required")
	}
	if d.Reader == nil || isNil(d.Reader) {
		return nil, errors.New("reviews connector: Reader is required")
	}
	if d.IDs == nil {
		return nil, errors.New("reviews connector: IDs is required")
	}
	if d.Logger == nil {
		return nil, errors.New("reviews connector: Logger is required")
	}
	if d.Ledger == nil || isNil(d.Ledger) {
		return nil, errors.New("reviews connector: Ledger is required")
	}
	now := d.Now
	if now == nil {
		now = time.Now
	}
	return &Connector{reader: d.Reader, ids: d.IDs, logger: d.Logger, now: now, ledger: d.Ledger}, nil
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var out Result
	if c == nil {
		return out, errors.New("reviews connector: connector is not initialized")
	}
	if c.reader == nil || c.ids == nil || c.now == nil || c.ledger == nil {
		return out, errors.New("reviews connector: connector is not initialized")
	}
	if appender == nil || isNil(appender) {
		return out, errors.New("reviews connector: appender is required")
	}
	items, err := c.reader.ReadCommittedVerdicts(ctx)
	if err != nil {
		return out, fmt.Errorf("reviews connector: read committed artifacts: %w", err)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	out.Listed = len(items)
	valid := make([]string, 0, len(items))
	duplicates := duplicateIdentities(items)
	// fail returns err joined with the refusal error when anything was refused (R-INV-13).
	fail := func(err error) (Result, error) {
		out.Cursor = cursor(valid)
		return out, errors.Join(err, refusedError(out.Refused))
	}
	for _, item := range items {
		candidate, err := wireCandidate(item.Path, item.Bytes)
		if err != nil {
			out.Malformed++
			return fail(fmt.Errorf("reviews connector: %s: %w", item.Path, err))
		}
		if !candidate {
			out.Documentary++
			valid = append(valid, item.Path)
			continue
		}
		schema, err := artifactSchema(item.Bytes)
		if err != nil {
			out.Malformed++
			return fail(fmt.Errorf("reviews connector: %s: %w", item.Path, err))
		}
		var value any
		var nativeID string
		var verdict *ObligationVerdict
		var review *RequirementReview
		kind := core.KindReviewVerdict
		switch schema {
		case "vera.verdict.v1":
			v, parseErr := Parse(item.Path, item.Bytes)
			if parseErr != nil {
				err = parseErr
			} else {
				value = v
				nativeID = v.VerdictID
			}
		case "proofbound.obligation-verdict.v2":
			v, parseErr := ParseObligationVerdict(item.Path, item.Bytes)
			if parseErr != nil {
				err = parseErr
			} else {
				value = v
				nativeID = v.VerdictID
				verdict = &v
			}
		case "proofbound.requirement-review.v1":
			v, parseErr := ParseRequirementReview(item.Path, item.Bytes)
			if parseErr != nil {
				err = parseErr
			} else {
				value = v
				nativeID = v.ReviewID
				kind = core.KindRequirementReview
				review = &v
			}
		default:
			err = fmt.Errorf("unknown review schema %q", schema)
		}
		if err != nil {
			out.Malformed++
			return fail(fmt.Errorf("reviews connector: %s: %w", item.Path, err))
		}
		if verdict != nil || review != nil {
			cause, err := c.holdCause(ctx, schema, nativeID, verdict, review, duplicates)
			if err != nil {
				return fail(fmt.Errorf("reviews connector: %s: ledger view: %w", item.Path, err))
			}
			if cause != "" {
				refused, err := c.isRevision(ctx, schema, nativeID, cause)
				if err != nil {
					return fail(fmt.Errorf("reviews connector: %s: ledger view: %w", item.Path, err))
				}
				if refused {
					out.Refused = append(out.Refused, Hold{Path: item.Path, Cause: cause})
				} else {
					out.Held = append(out.Held, Hold{Path: item.Path, Cause: cause})
				}
				continue
			}
		}
		valid = append(valid, item.Path)
		payload, err := json.Marshal(value)
		if err != nil {
			return fail(fmt.Errorf("reviews connector: %s: marshal: %w", item.Path, err))
		}
		e, err := c.ids.NewEvent(core.NewEventParams{Source: core.SourceReviews, NativeID: nativeID, Kind: kind, OccurredAt: c.now(), Payload: payload, ConnectorVersion: Version})
		if err != nil {
			return fail(fmt.Errorf("reviews connector: %s: event: %w", item.Path, err))
		}
		_, inserted, err := appender.Append(ctx, e)
		if err != nil {
			return fail(fmt.Errorf("reviews connector: %s: append: %w", item.Path, err))
		}
		if inserted {
			out.Appended++
		} else {
			out.Existing++
		}
	}
	out.Cursor = cursor(valid)
	return out, refusedError(out.Refused)
}

// identityKey is an artifact's identity for the hold: its payload schema and native id (R-INV-13).
func identityKey(schema, nativeID string) string { return schema + "\x00" + nativeID }

// duplicateIdentities returns the v2-verdict and requirement-review identities that more than one
// listed artifact carries. Artifacts that do not parse are ignored here; the main loop reaches them
// at their own position and fails as before (R-INV-5).
func duplicateIdentities(items []Artifact) map[string]bool {
	count := map[string]int{}
	for _, item := range items {
		if candidate, err := wireCandidate(item.Path, item.Bytes); err != nil || !candidate {
			continue
		}
		schema, err := artifactSchema(item.Bytes)
		if err != nil {
			continue
		}
		switch schema {
		case "proofbound.obligation-verdict.v2":
			if v, err := ParseObligationVerdict(item.Path, item.Bytes); err == nil {
				count[identityKey(schema, v.VerdictID)]++
			}
		case "proofbound.requirement-review.v1":
			if v, err := ParseRequirementReview(item.Path, item.Bytes); err == nil {
				count[identityKey(schema, v.ReviewID)]++
			}
		}
	}
	dup := map[string]bool{}
	for key, n := range count {
		if n > 1 {
			dup[key] = true
		}
	}
	return dup
}

// holdCause returns "" when the artifact may be appended, or the cause it is held or refused for:
// duplicate-identity first, then the order-only prerequisites in the order R-INV-13 lists them.
// Evidence is never consulted.
func (c *Connector) holdCause(ctx context.Context, schema, nativeID string, verdict *ObligationVerdict, review *RequirementReview, duplicates map[string]bool) (string, error) {
	if duplicates[identityKey(schema, nativeID)] {
		return "duplicate-identity", nil
	}
	requirements := []connectorintent.Reference{}
	if verdict != nil {
		provider := strings.TrimPrefix(verdict.ChangeIntent.Source, "intent.")
		claimed, err := c.ledger.CommitClaimed(ctx, verdict.ReviewedCommit, provider, verdict.ChangeIntent.RecordID, verdict.ChangeIntent.ArtifactSHA256)
		if err != nil {
			return "", err
		}
		if !claimed {
			return "claim-not-ledgered", nil
		}
		requirements = verdict.Requirements
	} else {
		requirements = append(requirements, review.Requirement)
	}
	for _, ref := range requirements {
		ok, err := c.ledger.RequirementLedgered(ctx, core.Source(ref.Source), ref.RecordID, ref.ArtifactSHA256)
		if err != nil {
			return "", err
		}
		if !ok {
			return "requirement-not-ledgered", nil
		}
	}
	return "", nil
}

// isRevision reports whether a not-appended artifact is refused rather than held: a duplicate
// identity is always refused; otherwise it is refused when its identity is already ledgered.
func (c *Connector) isRevision(ctx context.Context, schema, nativeID, cause string) (bool, error) {
	if cause == "duplicate-identity" {
		return true, nil
	}
	return c.ledger.Ledgered(ctx, schema, nativeID)
}

// refusedError lists every refused artifact, or is nil when none was refused.
func refusedError(refused []Hold) error {
	if len(refused) == 0 {
		return nil
	}
	parts := make([]string, 0, len(refused))
	for _, h := range refused {
		parts = append(parts, h.Cause+" "+h.Path)
	}
	return fmt.Errorf("reviews connector: %w: %s", ErrRevisionRefused, strings.Join(parts, "; "))
}

func artifactSchema(data []byte) (string, error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return "", errors.New("wire artifact front matter is incomplete")
	}
	if lines[0] != "---" {
		return "", errors.New("wire artifact must start with front matter")
	}
	if strings.HasPrefix(lines[1], "schema: ") {
		return strings.TrimPrefix(lines[1], "schema: "), nil
	}
	if strings.HasPrefix(lines[1], "{") {
		var probe struct {
			Schema string `json:"schema"`
		}
		if err := json.Unmarshal([]byte(lines[1]), &probe); err != nil {
			return "", err
		}
		if probe.Schema == "" {
			return "", errors.New("schema is required")
		}
		return probe.Schema, nil
	}
	return "", errors.New("schema declaration is invalid")
}

func wireCandidate(path string, data []byte) (bool, error) {
	if !utf8.Valid(data) {
		return false, errors.New("artifact is not valid UTF-8")
	}
	if !validPath(path) {
		return false, errors.New("artifact path is not under docs/verification/verdicts and does not end in .md")
	}
	first, _, _ := strings.Cut(string(data), "\n")
	return first == "---" || strings.HasPrefix(first, "schema:"), nil
}

// Parse validates the complete front matter and binds its declared path to path.
func Parse(path string, data []byte) (Verdict, error) {
	var v Verdict
	if !utf8.Valid(data) {
		return v, errors.New("artifact is not valid UTF-8")
	}
	if !validPath(path) {
		return v, errors.New("artifact path is not under docs/verification/verdicts and does not end in .md")
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		return v, errors.New("front matter is incomplete")
	}
	if lines[0] != "---" {
		return v, errors.New("front matter must start with ---")
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return v, errors.New("front matter has no closing ---")
	}
	if end == 1 {
		return v, errors.New("front matter is empty")
	}
	seen := map[string]bool{}
	var findingLines []string
	for _, line := range lines[1:end] {
		if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "    ") {
			findingLines = append(findingLines, line)
			continue
		}
		key, rawVal, ok := strings.Cut(line, ":")
		if !ok {
			return v, fmt.Errorf("invalid front matter line %q", line)
		}
		if strings.TrimSpace(key) != key {
			return v, fmt.Errorf("invalid front matter line %q", line)
		}
		if key == "" {
			return v, fmt.Errorf("invalid front matter line %q", line)
		}
		if rawVal == "" && key != "findings" {
			return v, fmt.Errorf("invalid front matter value %q", line)
		}
		if rawVal != "" {
			if !strings.HasPrefix(rawVal, " ") {
				return v, fmt.Errorf("invalid front matter value %q", line)
			}
			if strings.HasPrefix(rawVal, "  ") {
				return v, fmt.Errorf("invalid front matter value %q", line)
			}
		}
		val := strings.TrimPrefix(rawVal, " ")
		if strings.TrimSpace(val) != val {
			return v, fmt.Errorf("invalid front matter value %q", line)
		}
		if seen[key] {
			return v, fmt.Errorf("duplicate field %q", key)
		}
		seen[key] = true
		switch key {
		case "schema":
			v.Schema = val
		case "verdict_id":
			v.VerdictID = val
		case "status":
			v.Status = val
		case "reviewed_commit":
			v.ReviewedCommit = val
		case "artifact_path":
			v.ArtifactPath = val
		case "artifact_sha":
			v.ArtifactSHA = val
		case "findings":
			if val != "" && val != "[]" {
				return v, errors.New("findings must be [] or a list")
			}
			v.Findings = []Finding{}
		default:
			return v, fmt.Errorf("unknown field %q", key)
		}
	}
	if !seen["findings"] { // list form is parsed below; scalar form was not required
		for _, line := range findingLines {
			if strings.HasPrefix(line, "  - ") {
				seen["findings"] = true
				break
			}
		}
	}
	for _, required := range []string{"schema", "verdict_id", "status", "reviewed_commit", "findings", "artifact_path", "artifact_sha"} {
		if !seen[required] {
			return v, errors.New("front matter must contain exactly the required fields")
		}
	}
	if len(findingLines) > 0 {
		parsed, err := parseFindings(findingLines)
		if err != nil {
			return v, err
		}
		v.Findings = parsed
	}
	if v.ArtifactSHA != ArtifactSHA(data) {
		return v, errors.New("artifact_sha does not match committed artifact bytes")
	}
	if err := v.validate(path); err != nil {
		return v, err
	}
	return v, nil
}

func parseFindings(lines []string) ([]Finding, error) {
	var out []Finding
	var cur *Finding
	seen := map[string]bool{}
	for _, line := range lines {
		if strings.HasPrefix(line, "  - ") {
			if cur != nil {
				if err := validateFinding(*cur); err != nil {
					return nil, err
				}
				out = append(out, *cur)
			}
			cur = &Finding{}
			seen = map[string]bool{}
			line = strings.TrimPrefix(line, "  - ")
		} else {
			if cur == nil {
				return nil, errors.New("invalid findings list")
			}
			if !strings.HasPrefix(line, "    ") {
				return nil, errors.New("invalid findings list")
			}
			line = strings.TrimSpace(line)
		}
		key, rawVal, ok := strings.Cut(line, ":")
		if !ok {
			return nil, errors.New("invalid finding field")
		}
		if strings.TrimSpace(key) != key {
			return nil, errors.New("invalid finding field")
		}
		if rawVal != "" {
			if !strings.HasPrefix(rawVal, " ") {
				return nil, errors.New("invalid finding field")
			}
		}
		if seen[key] {
			return nil, errors.New("invalid or duplicate finding field")
		}
		val := strings.TrimPrefix(rawVal, " ")
		if strings.TrimSpace(val) != val {
			return nil, errors.New("invalid finding value")
		}
		seen[key] = true
		switch key {
		case "finding_id":
			cur.FindingID = val
		case "severity":
			cur.Severity = val
		case "defect_commit":
			cur.DefectCommit = val
		default:
			return nil, fmt.Errorf("unknown finding field %q", key)
		}
	}
	if cur == nil {
		return []Finding{}, nil
	}
	if err := validateFinding(*cur); err != nil {
		return nil, err
	}
	return append(out, *cur), nil
}

func (v Verdict) validate(path string) error {
	if v.Schema != "vera.verdict.v1" {
		return errors.New("schema is invalid")
	}
	if !idRE.MatchString(v.VerdictID) {
		return errors.New("verdict_id is invalid")
	}
	if v.Status != "ACCEPTABLE" && v.Status != "NEEDS_WORK" {
		return errors.New("status is invalid")
	}
	if !commitRE.MatchString(v.ReviewedCommit) {
		return errors.New("commit is invalid")
	}
	if !validPath(v.ArtifactPath) {
		return errors.New("artifact path is invalid")
	}
	if v.ArtifactPath != path {
		return errors.New("artifact path does not match")
	}
	if !digestRE.MatchString(v.ArtifactSHA) {
		return errors.New("artifact sha is invalid")
	}
	return nil
}
func validateFinding(f Finding) error {
	if !idRE.MatchString(f.FindingID) {
		return errors.New("finding id is invalid")
	}
	if f.Severity != "HIGH" && f.Severity != "MED" && f.Severity != "LOW" {
		return errors.New("finding severity is invalid")
	}
	if f.DefectCommit != "" && !commitRE.MatchString(f.DefectCommit) {
		return errors.New("finding is invalid")
	}
	return nil
}
func validPath(p string) bool {
	if !strings.HasPrefix(p, "docs/verification/verdicts/") {
		return false
	}
	if !strings.HasSuffix(p, ".md") {
		return false
	}
	if strings.Contains(p, "..") || strings.Contains(p, "\\") {
		return false
	}
	return true
}
func cursor(p []string) json.RawMessage { b, _ := json.Marshal(p); return b }
func isNil(v any) bool {
	if v == nil {
		return true
	}
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return x.IsNil()
	}
	return false
}

// ArtifactSHA computes the digest format required by the contract for callers
// producing committed-artifact metadata. The connector intentionally does not
// rewrite a self-referential front matter digest.
func ArtifactSHA(data []byte) string {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "artifact_sha: ") {
			lines[i] = "artifact_sha: " + strings.Repeat("0", 64)
			break
		}
	}
	h := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(h[:])
}
