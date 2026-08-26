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

type Deps struct {
	Reader CommittedReader
	IDs    *core.IDGenerator
	Logger *slog.Logger
	Now    func() time.Time
}
type Connector struct {
	reader CommittedReader
	ids    *core.IDGenerator
	logger *slog.Logger
	now    func() time.Time
}
type Result struct {
	Listed, Appended, Existing, Malformed int
	Cursor                                json.RawMessage
}

func New(d *Deps) (*Connector, error) {
	if d == nil || d.Reader == nil || isNil(d.Reader) {
		return nil, errors.New("reviews connector: Reader is required")
	}
	if d.IDs == nil {
		return nil, errors.New("reviews connector: IDs is required")
	}
	if d.Logger == nil {
		return nil, errors.New("reviews connector: Logger is required")
	}
	now := d.Now
	if now == nil {
		now = time.Now
	}
	return &Connector{reader: d.Reader, ids: d.IDs, logger: d.Logger, now: now}, nil
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var out Result
	if c == nil || c.reader == nil || c.ids == nil || c.now == nil {
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
	for _, item := range items {
		v, err := Parse(item.Path, item.Bytes)
		if err != nil {
			out.Malformed++
			out.Cursor = cursor(valid)
			return out, fmt.Errorf("reviews connector: %s: %w", item.Path, err)
		}
		valid = append(valid, item.Path)
		payload, err := json.Marshal(v)
		if err != nil {
			out.Cursor = cursor(valid)
			return out, fmt.Errorf("reviews connector: %s: marshal: %w", item.Path, err)
		}
		e, err := c.ids.NewEvent(core.NewEventParams{Source: core.SourceReviews, NativeID: v.VerdictID, Kind: core.KindReviewVerdict, OccurredAt: c.now(), Payload: payload, ConnectorVersion: Version})
		if err != nil {
			out.Cursor = cursor(valid)
			return out, fmt.Errorf("reviews connector: %s: event: %w", item.Path, err)
		}
		_, inserted, err := appender.Append(ctx, e)
		if err != nil {
			out.Cursor = cursor(valid)
			return out, fmt.Errorf("reviews connector: %s: append: %w", item.Path, err)
		}
		if inserted {
			out.Appended++
		} else {
			out.Existing++
		}
	}
	out.Cursor = cursor(valid)
	return out, nil
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
	if len(lines) < 3 || lines[0] != "---" {
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
		if !ok || strings.TrimSpace(key) != key || key == "" {
			return v, fmt.Errorf("invalid front matter line %q", line)
		}
		if (rawVal == "" && key != "findings") || (rawVal != "" && (!strings.HasPrefix(rawVal, " ") || strings.HasPrefix(rawVal, "  "))) {
			return v, fmt.Errorf("invalid front matter value %q", line)
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
	if !seen["schema"] || !seen["verdict_id"] || !seen["status"] || !seen["reviewed_commit"] || !seen["findings"] || !seen["artifact_path"] || !seen["artifact_sha"] {
		return v, errors.New("front matter must contain exactly the required fields")
	}
	if len(findingLines) > 0 && len(v.Findings) == 0 {
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
		} else if cur == nil || !strings.HasPrefix(line, "    ") {
			return nil, errors.New("invalid findings list")
		} else {
			line = strings.TrimSpace(line)
		}
		key, rawVal, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != key || (rawVal != "" && !strings.HasPrefix(rawVal, " ")) || seen[key] {
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
	if v.Schema != "vera.verdict.v1" || !idRE.MatchString(v.VerdictID) {
		return errors.New("schema or verdict_id is invalid")
	}
	if v.Status != "ACCEPTABLE" && v.Status != "NEEDS_WORK" {
		return errors.New("status is invalid")
	}
	if !commitRE.MatchString(v.ReviewedCommit) || !validPath(v.ArtifactPath) || v.ArtifactPath != path || !digestRE.MatchString(v.ArtifactSHA) {
		return errors.New("commit, artifact path, or artifact sha is invalid")
	}
	for _, f := range v.Findings {
		if err := validateFinding(f); err != nil {
			return err
		}
	}
	return nil
}
func validateFinding(f Finding) error {
	if !idRE.MatchString(f.FindingID) || (f.Severity != "HIGH" && f.Severity != "MED" && f.Severity != "LOW") || (f.DefectCommit != "" && !commitRE.MatchString(f.DefectCommit)) {
		return errors.New("finding is invalid")
	}
	return nil
}
func validPath(p string) bool {
	return strings.HasPrefix(p, "docs/verification/verdicts/") && strings.HasSuffix(p, ".md") && !strings.Contains(p, "..") && !strings.Contains(p, "\\")
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
