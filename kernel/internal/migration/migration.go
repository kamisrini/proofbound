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
		seq: "1227", lineSHA: "86fccc57e22472df3cdc3034f1293d3ebe2f2f2538f177d771948fa838d1f4b4",
		id: "01M25Y75PJCSGE5Q4ZX6JA9BH7", source: "intent.records", native: "BD-proofbound-p5-a1b2c3", kind: "business_decision.recorded",
		occurred: "2026-09-10T11:15:43.403356-04:00", recorded: "2026-09-10T11:15:43.442482-04:00",
		contentSHA: "b0dd1b659ad303852bdc1faf5f9071c0779199811e095a70e6323555655965e4", connector: "intent/1",
	},
	{
		seq: "1228", lineSHA: "8ef06dcbb7db771740e59267848b0f1e1cab9b15d3e4e0f7a0c78f916388cdc5",
		id: "01M25Y75PSMB0AM3N9XV8VG0F8", source: "intent.records", native: "BR-intent-chain-d4e5f6", kind: "requirement.recorded",
		occurred: "2026-09-10T11:15:43.403356-04:00", recorded: "2026-09-10T11:15:43.449557-04:00",
		contentSHA: "2590baa37f89a256fa3dc9ca19affc455e67a210007c6573f31ae5508cb4a864", connector: "intent/1",
	},
	{
		seq: "1229", lineSHA: "2960bc57c84a73b061960c3a978a48a1757c8156da5551c5845b80a5515d1e75",
		id: "01M25Y75PW3T47JK1QN9CR9G08", source: "intent.records", native: "CI-implement-p5-0a1b2c", kind: "change_intent.recorded",
		occurred: "2026-09-10T11:15:43.403356-04:00", recorded: "2026-09-10T11:15:43.452593-04:00",
		contentSHA: "b0cf6c8d4a13c536dbdee5d06c2b24b81ef603b3fa37dfe863bd12d55373ac37", connector: "intent/1",
	},
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
