//go:build liveacceptance

package sessions

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestLiveQuiescentJSONLAcceptance(t *testing.T) {
	source := os.Getenv("PROOFBOUND_SESSIONS_LIVE_FILE")
	if source == "" {
		t.Fatal("PROOFBOUND_SESSIONS_LIVE_FILE is required for explicit live acceptance")
	}
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	quiescent := time.Since(info.ModTime())
	if info.Size() == 0 || quiescent < 10*time.Minute {
		t.Fatalf("source must be nonempty and quiescent: size=%d age=%s", info.Size(), quiescent)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	total, valid := jsonlCounts(t, data)
	if total == 0 || valid == 0 || float64(valid)/float64(total) < 0.5 {
		t.Fatalf("insufficient live parse coverage: total=%d valid=%d", total, valid)
	}

	root, home := t.TempDir(), t.TempDir()
	dir := ProjectDir(home, root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(dir, "live.jsonl")
	if err := os.WriteFile(destination, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(destination, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}

	appender := &fakeAppender{}
	connector := testConnector(t, root, home)
	first, err := connector.Sync(context.Background(), appender)
	if err != nil || first.Appended != 1 || first.Existing != 0 || len(appender.events) != 1 {
		t.Fatalf("first sync result=%+v events=%d err=%v", first, len(appender.events), err)
	}
	var payload map[string]any
	if err := json.Unmarshal(appender.events[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	wantKeys := []string{"files_written_count", "finished_at", "message_count", "parse_coverage", "session_id", "started_at", "tool_call_count"}
	if len(keys) != len(wantKeys) {
		t.Fatalf("payload keys=%v", keys)
	}
	for i := range keys {
		if keys[i] != wantKeys[i] {
			t.Fatalf("payload keys=%v", keys)
		}
	}

	second, err := connector.Sync(context.Background(), appender)
	if err != nil || second.Appended != 0 || second.Existing != 1 || len(appender.events) != 1 {
		t.Fatalf("replay result=%+v events=%d err=%v", second, len(appender.events), err)
	}
	digest := sha256.Sum256(data)
	t.Logf("live_session source_sha256=%s quiescent_seconds=%d lines_total=%d lines_valid=%d lines_skipped=%d first_appended=%d replay_appended=%d replay_existing=%d payload_keys=%v",
		hex.EncodeToString(digest[:]), int64(quiescent.Seconds()), total, valid, total-valid, first.Appended, second.Appended, second.Existing, keys)
}

func jsonlCounts(t *testing.T, data []byte) (total, valid int) {
	t.Helper()
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		total++
		var value map[string]any
		if json.Unmarshal(line, &value) == nil {
			valid++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return total, valid
}
