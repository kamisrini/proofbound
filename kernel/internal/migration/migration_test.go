package migration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func committedArchive(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "verification", "p6-historical-evidence.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadCommittedArchive(t *testing.T) {
	records, err := Parse(bytes.NewReader(committedArchive(t)))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 5 || records[0].Seq != 1227 || records[1].Seq != 1228 || records[2].Seq != 1229 || records[3].Seq != 1670 || records[4].Seq != 1834 {
		t.Fatalf("records=%+v", records)
	}
	if records[0].Event.ID.String() != "01M25Y75PJCSGE5Q4ZX6JA9BH7" || records[1].Event.ID.String() != "01M25Y75PSMB0AM3N9XV8VG0F8" || records[2].Event.ID.String() != "01M25Y75PW3T47JK1QN9CR9G08" || records[3].Event.ID.String() != "01M28TPW9C8R7ND19MNDCJ9GDG" || records[4].Event.ID.String() != "01M29HMPE5V977AR3VMW47DVDE" {
		t.Fatalf("ids=%s,%s,%s,%s,%s", records[0].Event.ID, records[1].Event.ID, records[2].Event.ID, records[3].Event.ID, records[4].Event.ID)
	}
}

func TestParseRejectsArchiveMutations(t *testing.T) {
	base := committedArchive(t)
	cases := map[string][]byte{
		"altered envelope":  bytes.Replace(base, []byte(`"seq":1670`), []byte(`"seq":1671`), 1),
		"additional record": append(append([]byte(nil), base...), base[:bytes.IndexByte(base, '\n')+1]...),
		"reordered records": append(append([]byte(nil), base[bytes.IndexByte(base, '\n')+1:]...), base[:bytes.IndexByte(base, '\n')+1]...),
		"unknown field":     bytes.Replace(base, []byte(`{"seq":1670`), []byte(`{"extra":true,"seq":1670`), 1),
		"malformed JSON":    bytes.Replace(base, []byte(`{"seq":1670`), []byte(`{"seq":`), 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(bytes.NewReader(data)); err == nil {
				t.Fatal("mutated archive accepted")
			}
		})
	}
	if _, err := Parse(bytes.NewReader(append(append([]byte(nil), base...), '\n'))); err != nil {
		t.Fatal("final newline should be accepted:", err)
	}
}

func TestImportRefusesNonEmptyLedger(t *testing.T) {
	s, err := store.Open(context.Background(), store.Config{Root: filepath.Join(t.TempDir(), ".proofbound"), DatabaseURL: os.Getenv("DATABASE_URL"), AllowHistoricalEvidenceImport: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r, err := s.BeginSync(context.Background(), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := Parse(bytes.NewReader(committedArchive(t)))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Append(context.Background(), archive[0].Event); err != nil {
		t.Fatal(err)
	}
	if err := r.Finish(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := Import(context.Background(), s, archive); err != ErrNonEmptyLedger {
		t.Fatalf("err=%v want=%v", err, ErrNonEmptyLedger)
	}
	count := 0
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(store.Record) error { count++; return nil }); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("non-empty ledger changed: count=%d", count)
	}
}

func TestImportExactArchive(t *testing.T) {
	s, err := store.Open(context.Background(), store.Config{Root: filepath.Join(t.TempDir(), ".proofbound"), DatabaseURL: os.Getenv("DATABASE_URL"), AllowHistoricalEvidenceImport: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	archive, err := Parse(bytes.NewReader(committedArchive(t)))
	if err != nil {
		t.Fatal(err)
	}
	if err := Import(context.Background(), s, archive); err != nil {
		t.Fatal(err)
	}
	var got []store.Record
	if err := s.ReadEvents(context.Background(), store.Filter{}, func(r store.Record) error { got = append(got, r); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 || got[0].Seq != 1227 || got[1].Seq != 1228 || got[2].Seq != 1229 || got[3].Seq != 1670 || got[4].Seq != 1834 || got[0].Event.ID != archive[0].Event.ID || got[4].Event.ID != archive[4].Event.ID {
		t.Fatalf("got=%+v", got)
	}
}
