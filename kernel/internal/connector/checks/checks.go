package checks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "checks/1"

var (
	ulidPattern = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)
	hex64       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	gitHash     = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
)

type ToolVersions struct {
	Go           string `json:"go"`
	GolangCILint string `json:"golangci_lint"`
	Make         string `json:"make"`
}

type Witness struct {
	Schema       string       `json:"schema"`
	RunID        string       `json:"run_id"`
	Command      string       `json:"command"`
	ExitCode     int          `json:"exit_code"`
	StartedAt    time.Time    `json:"started_at"`
	FinishedAt   time.Time    `json:"finished_at"`
	DurationMS   int64        `json:"duration_ms"`
	OutputSHA256 string       `json:"output_sha256"`
	GitSHA       string       `json:"git_sha"`
	GitDirty     bool         `json:"git_dirty"`
	ToolVersions ToolVersions `json:"tool_versions"`
}

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
	SpoolDir string
	IDs      *core.IDGenerator
	Logger   *slog.Logger
}

type Connector struct {
	spoolDir string
	ids      *core.IDGenerator
	logger   *slog.Logger
}

type Result struct {
	Listed   int
	Appended int
	Existing int
	Cursor   json.RawMessage
}

func New(deps *Deps) (*Connector, error) {
	if deps == nil {
		return nil, errors.New("checks connector: dependencies are required")
	}
	if deps.SpoolDir == "" {
		return nil, errors.New("checks connector: SpoolDir is required")
	}
	if deps.IDs == nil {
		return nil, errors.New("checks connector: IDs is required")
	}
	if deps.Logger == nil {
		return nil, errors.New("checks connector: Logger is required")
	}
	abs, err := filepath.Abs(deps.SpoolDir)
	if err != nil {
		return nil, fmt.Errorf("checks connector: spool directory: %w", err)
	}
	return &Connector{spoolDir: abs, ids: deps.IDs, logger: deps.Logger}, nil
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var result Result
	if c == nil || c.ids == nil || c.spoolDir == "" {
		return result, errors.New("checks connector: connector is not initialized")
	}
	if appender == nil || isNilAppender(appender) {
		return result, errors.New("checks connector: appender is required")
	}
	entries, err := os.ReadDir(c.spoolDir)
	if errors.Is(err, os.ErrNotExist) {
		result.Cursor = json.RawMessage("[]")
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("checks connector: read spool: %w", err)
	}
	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)
	result.Listed = len(filenames)
	validated := make([]string, 0, len(filenames))
	for _, filename := range filenames {
		witness, readErr := readWitness(filepath.Join(c.spoolDir, filename))
		if readErr != nil {
			result.Cursor = marshalCursor(validated)
			return result, fmt.Errorf("checks connector: %s: %w", filename, readErr)
		}
		if filename != witness.RunID+".json" {
			result.Cursor = marshalCursor(validated)
			return result, fmt.Errorf("checks connector: %s: filename does not match run_id", filename)
		}
		validated = append(validated, filename)
		payload, marshalErr := json.Marshal(witness)
		if marshalErr != nil {
			result.Cursor = marshalCursor(validated)
			return result, fmt.Errorf("checks connector: %s: marshal: %w", filename, marshalErr)
		}
		event, eventErr := c.ids.NewEvent(core.NewEventParams{
			Source:           core.SourceChecks,
			NativeID:         witness.RunID,
			Kind:             core.KindCheckRun,
			OccurredAt:       witness.StartedAt,
			Payload:          payload,
			ConnectorVersion: Version,
		})
		if eventErr != nil {
			result.Cursor = marshalCursor(validated)
			return result, fmt.Errorf("checks connector: %s: event: %w", filename, eventErr)
		}
		_, inserted, appendErr := appender.Append(ctx, event)
		if appendErr != nil {
			result.Cursor = marshalCursor(validated)
			return result, fmt.Errorf("checks connector: %s: append: %w", filename, appendErr)
		}
		if inserted {
			result.Appended++
		} else {
			result.Existing++
		}
	}
	result.Cursor = marshalCursor(validated)
	return result, nil
}

func readWitness(path string) (Witness, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Witness{}, err
	}
	if !utf8.Valid(data) {
		return Witness{}, errors.New("witness is not valid UTF-8")
	}
	if _, err := core.Canonicalize(data); err != nil {
		return Witness{}, fmt.Errorf("canonical JSON: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return Witness{}, err
	}
	required := [...]string{
		"schema", "run_id", "command", "exit_code", "started_at", "finished_at",
		"duration_ms", "output_sha256", "git_sha", "git_dirty", "tool_versions",
	}
	if len(fields) != len(required) {
		return Witness{}, errors.New("witness must contain exactly the v1 fields")
	}
	for _, name := range required {
		value, ok := fields[name]
		if !ok {
			return Witness{}, fmt.Errorf("missing field %q", name)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Witness{}, fmt.Errorf("field %q must not be null", name)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var witness Witness
	if err := decoder.Decode(&witness); err != nil {
		return Witness{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Witness{}, errors.New("trailing JSON value")
	} else if !errors.Is(err, io.EOF) {
		return Witness{}, err
	}
	if err := witness.validate(); err != nil {
		return Witness{}, err
	}
	return witness, nil
}

func isNilAppender(appender Appender) bool {
	value := reflect.ValueOf(appender)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (w Witness) validate() error {
	switch {
	case w.Schema != "vera.witness.v1":
		return errors.New("schema must be vera.witness.v1")
	case !ulidPattern.MatchString(w.RunID):
		return errors.New("run_id is not a strict ULID")
	case w.Command != "make check":
		return errors.New("command must be make check")
	case w.ExitCode < 0 || w.ExitCode > 255:
		return errors.New("exit_code is outside 0..255")
	case w.StartedAt.IsZero() || w.FinishedAt.IsZero() || w.FinishedAt.Before(w.StartedAt):
		return errors.New("timestamps are invalid")
	case w.DurationMS < 0:
		return errors.New("duration_ms is negative")
	case !hex64.MatchString(w.OutputSHA256):
		return errors.New("output_sha256 is malformed")
	case !gitHash.MatchString(w.GitSHA):
		return errors.New("git_sha is malformed")
	case strings.TrimSpace(w.ToolVersions.Go) == "" || strings.TrimSpace(w.ToolVersions.GolangCILint) == "" || strings.TrimSpace(w.ToolVersions.Make) == "":
		return errors.New("tool_versions are incomplete")
	}
	return nil
}

func marshalCursor(filenames []string) json.RawMessage {
	if filenames == nil {
		filenames = []string{}
	}
	data, _ := json.Marshal(filenames)
	return data
}
