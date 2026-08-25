package sessions

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "sessions/1"

var userHomeDir = os.UserHomeDir

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
	Root    string
	HomeDir string
	IDs     *core.IDGenerator
	Logger  *slog.Logger
	Now     func() time.Time
}

type Connector struct {
	root, home string
	ids        *core.IDGenerator
	logger     *slog.Logger
	now        func() time.Time
}

type Result struct {
	Listed, Appended, Existing, Skipped, Malformed int
	Cursor                                         json.RawMessage
}

type payload struct {
	SessionID         string     `json:"session_id"`
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	MessageCount      int        `json:"message_count"`
	ToolCallCount     int        `json:"tool_call_count"`
	FilesWrittenCount int        `json:"files_written_count"`
	ParseCoverage     float64    `json:"parse_coverage"`
}

func New(deps *Deps) (*Connector, error) {
	if deps == nil || strings.TrimSpace(deps.Root) == "" || deps.IDs == nil || deps.Logger == nil {
		return nil, errors.New("sessions connector: Root, IDs, and Logger are required")
	}
	home := deps.HomeDir
	if home == "" {
		var err error
		home, err = userHomeDir()
		if err != nil {
			return nil, fmt.Errorf("sessions connector: home directory: %w", err)
		}
	}
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	root, err := filepath.Abs(deps.Root)
	if err != nil {
		return nil, fmt.Errorf("sessions connector: root: %w", err)
	}
	return &Connector{root: root, home: home, ids: deps.IDs, logger: deps.Logger, now: now}, nil
}

func ProjectDir(home, root string) string {
	encoded := strings.NewReplacer("/", "-", ".", "-").Replace(root)
	return filepath.Join(home, ".claude", "projects", encoded)
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var result Result
	if c == nil || c.ids == nil || c.root == "" || c.home == "" {
		return result, errors.New("sessions connector: connector is not initialized")
	}
	if appender == nil || isNilAppender(appender) {
		return result, errors.New("sessions connector: appender is required")
	}
	dir := ProjectDir(c.home, c.root)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		result.Cursor = json.RawMessage("[]")
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("sessions connector: read project directory: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			return result, fmt.Errorf("sessions connector: stat %s: %w", name, err)
		}
		if info.ModTime().After(c.now().Add(-10 * time.Minute)) {
			result.Skipped++
			continue
		}
		result.Listed++
		p, err := parseFile(path, strings.TrimSuffix(name, ".jsonl"))
		if err != nil {
			result.Malformed++
			c.logger.Warn("sessions connector: skipped file", "file", name, "error", err)
			continue
		}
		if p.ParseCoverage < 0.5 {
			result.Skipped++
			continue
		}
		data, err := json.Marshal(p)
		if err != nil {
			return result, fmt.Errorf("sessions connector: marshal %s: %w", name, err)
		}
		occurred := c.now()
		if p.StartedAt != nil {
			occurred = *p.StartedAt
		}
		event, err := c.ids.NewEvent(core.NewEventParams{Source: core.SourceSessions, NativeID: p.SessionID, Kind: core.KindSessionObserved, OccurredAt: occurred, Payload: data, ConnectorVersion: Version})
		if err != nil {
			return result, fmt.Errorf("sessions connector: event %s: %w", name, err)
		}
		_, inserted, err := appender.Append(ctx, event)
		if err != nil {
			return result, fmt.Errorf("sessions connector: append %s: %w", name, err)
		}
		if inserted {
			result.Appended++
		} else {
			result.Existing++
		}
	}
	result.Cursor, _ = json.Marshal(names)
	return result, nil
}

func parseFile(path, fallbackID string) (payload, error) {
	f, err := os.Open(path)
	if err != nil {
		return payload{}, err
	}
	defer f.Close()
	var p payload
	p.SessionID = fallbackID
	var total, valid int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		total++
		var value map[string]any
		if json.Unmarshal([]byte(line), &value) != nil {
			continue
		}
		valid++
		observe(&p, value)
	}
	if err := scanner.Err(); err != nil {
		return payload{}, err
	}
	if total == 0 || valid == 0 {
		return payload{}, errors.New("no valid JSONL metadata")
	}
	if p.SessionID == "" {
		return payload{}, errors.New("session id is missing")
	}
	p.ParseCoverage = float64(valid) / float64(total)
	return p, nil
}

func observe(p *payload, value map[string]any) {
	p.MessageCount++
	id, _ := value["session_id"].(string)
	if id != "" {
		p.SessionID = id
	}
	id, _ = value["sessionId"].(string)
	if id != "" {
		p.SessionID = id
	}
	if ts := timestamp(value); ts != nil {
		if p.StartedAt == nil || ts.Before(*p.StartedAt) {
			p.StartedAt = ts
		}
		if p.FinishedAt == nil || ts.After(*p.FinishedAt) {
			p.FinishedAt = ts
		}
	}
	if isToolCall(value) {
		p.ToolCallCount++
	}
	if writesFile(value) {
		p.FilesWrittenCount++
	}
}

func timestamp(value map[string]any) *time.Time {
	for _, key := range []string{"timestamp", "created_at", "started_at", "ended_at"} {
		if raw, ok := value[key].(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
				return &t
			}
		}
	}
	return nil
}

func isToolCall(value map[string]any) bool {
	if typ, _ := value["type"].(string); typ == "tool_use" || typ == "tool_call" {
		return true
	}
	if _, ok := value["tool_calls"]; ok {
		return true
	}
	if message, ok := value["message"].(map[string]any); ok {
		if content, ok := message["content"].([]any); ok {
			for _, item := range content {
				if m, ok := item.(map[string]any); ok && (m["type"] == "tool_use" || m["type"] == "tool_call") {
					return true
				}
			}
		}
	}
	return false
}

func writesFile(value map[string]any) bool {
	name, _ := value["name"].(string)
	if name == "Write" || name == "Edit" || name == "write_file" || name == "edit_file" {
		return true
	}
	if tool, ok := value["tool_name"].(string); ok && (tool == "Write" || tool == "Edit") {
		return true
	}
	return false
}

func isNilAppender(appender Appender) bool {
	v := reflect.ValueOf(appender)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
