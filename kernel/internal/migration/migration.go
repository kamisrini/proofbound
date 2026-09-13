// Package migration contains explicitly authorized, one-shot repository migrations.
package migration

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const ArchivePath = "docs/verification/p6-historical-evidence.jsonl"

var ErrNonEmptyLedger = errors.New("historical evidence migration: ledger is not empty")

type archiveEnvelope struct {
	Seq   int64      `json:"seq"`
	Event core.Event `json:"event"`
}

type expectedRecord struct {
	seq, lineSHA string
	id, source   string
	native, kind string
	occurred     string
	recorded     string
	contentSHA   string
	connector    string
}

var expected = []expectedRecord{
	{
		seq: "1670", lineSHA: "8eff383aed595751e650b41593aebf948a39caea19ba2b800f419facc475e3ea",
		id: "01M28TPW9C8R7ND19MNDCJ9GDG", source: "checks", native: "01M28TP6CJB4FYQWESR6ZTEV73", kind: "check.run",
		occurred: "2026-09-11T14:11:44-04:00", recorded: "2026-09-11T14:12:07.084781-04:00",
		contentSHA: "47ff02e6019ec725e2cd8831edb3617c6ed79f59905ab5eed13e4f9cde372b0a", connector: "checks/1",
	},
	{
		seq: "1834", lineSHA: "8265193d72f3828564403309fcc1073837975049f624a6d8f52a6a21bea29d2c",
		id: "01M29HMPE5V977AR3VMW47DVDE", source: "git", native: "c29bb3b78899f54588f9cbad6e6ecf50333d787c", kind: "commit.recorded",
		occurred: "2026-09-11T20:51:50-04:00", recorded: "2026-09-11T20:52:52.805563-04:00",
		contentSHA: "5a4f0628a35eb06a6c47787538fa5e961b7d2563270b12ac6d0ac737e5f0830e", connector: "git/2",
	},
}

func Load(path string) ([]store.Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read historical evidence archive: %w", err)
	}
	return Parse(bytes.NewReader(data))
}

func Parse(input io.Reader) ([]store.Record, error) {
	if input == nil {
		return nil, errors.New("historical evidence archive: input is required")
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	lines := make([][]byte, 0, len(expected))
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("historical evidence archive: scan: %w", err)
	}
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			return nil, errors.New("historical evidence archive: blank line")
		}
	}
	if len(lines) != len(expected) {
		return nil, fmt.Errorf("historical evidence archive: got %d records, want %d", len(lines), len(expected))
	}
	records := make([]store.Record, 0, len(lines))
	for i, line := range lines {
		gotSHA := sha256.Sum256(line)
		if hex.EncodeToString(gotSHA[:]) != expected[i].lineSHA {
			return nil, fmt.Errorf("historical evidence archive: record %d digest mismatch", i)
		}
		var envelope archiveEnvelope
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&envelope); err != nil {
			return nil, fmt.Errorf("historical evidence archive: record %d: %w", i, err)
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			if err == nil {
				return nil, fmt.Errorf("historical evidence archive: record %d has trailing JSON", i)
			}
			return nil, fmt.Errorf("historical evidence archive: record %d trailing data: %w", i, err)
		}
		if err := validateExpected(i, envelope); err != nil {
			return nil, err
		}
		records = append(records, store.Record{Seq: envelope.Seq, Event: envelope.Event})
	}
	return records, nil
}

func validateExpected(index int, envelope archiveEnvelope) error {
	want := expected[index]
	if want.seq != fmt.Sprint(envelope.Seq) {
		return fmt.Errorf("historical evidence archive: record %d sequence is %d, want %s", index, envelope.Seq, want.seq)
	}
	e := envelope.Event
	if err := e.Validate(); err != nil {
		return fmt.Errorf("historical evidence archive: record %d event: %w", index, err)
	}
	if e.ID.String() != want.id || string(e.Source) != want.source || e.NativeID != want.native || string(e.Kind) != want.kind || e.ConnectorVersion != want.connector {
		return fmt.Errorf("historical evidence archive: record %d identity mismatch", index)
	}
	if e.OccurredAt.Format(time.RFC3339Nano) != want.occurred || e.RecordedAt.Format(time.RFC3339Nano) != want.recorded {
		return fmt.Errorf("historical evidence archive: record %d timestamp mismatch", index)
	}
	if e.ContentSHA != want.contentSHA {
		return fmt.Errorf("historical evidence archive: record %d content hash mismatch", index)
	}
	contentSHA, err := core.ContentSHA(e.Payload)
	if err != nil || contentSHA != e.ContentSHA {
		return fmt.Errorf("historical evidence archive: record %d payload hash mismatch", index)
	}
	canonical, err := core.Canonicalize(e.Payload)
	if err != nil || !bytes.Equal(canonical, e.Payload) {
		return fmt.Errorf("historical evidence archive: record %d payload is not canonical", index)
	}
	return nil
}

func Import(ctx context.Context, ledger *store.Store, records []store.Record) error {
	if ledger == nil {
		return errors.New("historical evidence migration: ledger is required")
	}
	count := 0
	if err := ledger.ReadEvents(ctx, store.Filter{}, func(store.Record) error { count++; return nil }); err != nil {
		return fmt.Errorf("historical evidence migration: inspect ledger: %w", err)
	}
	if count != 0 {
		return ErrNonEmptyLedger
	}
	return ledger.ImportHistoricalEvidenceRecords(ctx, records)
}
