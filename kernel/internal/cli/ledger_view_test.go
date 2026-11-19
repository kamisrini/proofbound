package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/migration"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// R2 M3 SPEC §7 row 12 (draft 6, as amended by drafts 7-8): the REAL store-backed LedgerView over a
// mixed ledger. The store opens on an ephemeral port, so this package never contends for the
// embedded database's default port (port_home_test.go in the sdd package).

func ephemeralPort(t *testing.T) uint16 {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// TestStoreLedgerViewSeesEveryProvider builds a ledger from this repository's pinned historical
// archive (a records requirement and a git commit carrying a records claim) plus appended specdir
// and sdd requirements, git refs of each provider, and v1/v2 reviews events sharing one id; then
// asks every method about each.
func TestStoreLedgerViewSeesEveryProvider(t *testing.T) {
	ctx := context.Background()
	ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: ephemeralPort(t), DatabaseURL: os.Getenv("DATABASE_URL"), AllowHistoricalEvidenceImport: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ledger.Close() })
	if os.Getenv("DATABASE_URL") != "" {
		t.Skip("needs a fresh embedded ledger: the archive imports only into an empty one")
	}
	records, err := migration.Load(filepath.Join(repoRoot(t), migration.ArchivePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Import(ctx, ledger, records); err != nil {
		t.Fatal(err)
	}
	ids, err := newIDs()
	if err != nil {
		t.Fatal(err)
	}
	run, err := ledger.BeginSync(ctx, "m3-row-12")
	if err != nil {
		t.Fatal(err)
	}
	digest := func(c string) string { return strings.Repeat(c, 64) }
	appendEvent := func(source core.Source, kind core.Kind, nativeID string, payload any) {
		t.Helper()
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		e, err := ids.NewEvent(core.NewEventParams{Source: source, NativeID: nativeID, Kind: kind, OccurredAt: time.Unix(10, 0).UTC(), Payload: raw, ConnectorVersion: "m3-test/1"})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := run.Append(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	requirement := func(id, d string) map[string]string {
		return map[string]string{"requirement_id": id, "artifact_sha256": d}
	}
	appendEvent(core.SourceIntentSpecdir, core.KindRequirement, "BR-spec-item-aaaaaa", requirement("BR-spec-item-aaaaaa", digest("1")))
	appendEvent(core.SourceIntentSDD, core.KindRequirement, "BR-sdd-item-bbbbbb", requirement("BR-sdd-item-bbbbbb", digest("2")))
	sha := strings.Repeat("e", 40)
	appendEvent(core.SourceGit, core.KindCommitRecorded, sha, connectorgit.Commit{SHA: sha, IntentRefs: []connectorgit.IntentRef{
		{Provider: "sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: digest("3")},
		{Provider: "specdir", RecordID: "CI-spec-item-cccccc", ArtifactSHA256: digest("4")},
	}})
	appendEvent(core.SourceReviews, core.KindReviewVerdict, "V-shared", map[string]string{"schema": "vera.verdict.v1"})
	appendEvent(core.SourceReviews, core.KindRequirementReview, "RR-one", map[string]string{"schema": "proofbound.requirement-review.v1"})
	if err := run.Finish(ctx, json.RawMessage(`[]`), nil); err != nil {
		t.Fatal(err)
	}

	view := &storeLedgerView{ledger: ledger}
	check := func(name string, got bool, err error, want bool) {
		t.Helper()
		if err != nil || got != want {
			t.Errorf("%s: got %v (err %v), want %v", name, got, err, want)
		}
	}
	// This repository's own live v2 verdict cites the archived records claim and requirement.
	archived := "c29bb3b78899f54588f9cbad6e6ecf50333d787c"
	got, err := view.CommitClaimed(ctx, archived, "records", "CI-implement-p5-0a1b2c", "f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424")
	check("imported records claim", got, err, true)
	got, err = view.RequirementLedgered(ctx, core.SourceIntentRecords, "BR-intent-chain-d4e5f6", "913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562")
	check("imported records requirement", got, err, true)
	got, err = view.CommitClaimed(ctx, sha, "sdd", "CI-add-auth-xuuno2", digest("3"))
	check("sdd claim", got, err, true)
	got, err = view.CommitClaimed(ctx, sha, "specdir", "CI-spec-item-cccccc", digest("4"))
	check("specdir claim", got, err, true)
	got, err = view.RequirementLedgered(ctx, core.SourceIntentSpecdir, "BR-spec-item-aaaaaa", digest("1"))
	check("specdir requirement", got, err, true)
	got, err = view.RequirementLedgered(ctx, core.SourceIntentSDD, "BR-sdd-item-bbbbbb", digest("2"))
	check("sdd requirement", got, err, true)
	got, err = view.Ledgered(ctx, "vera.verdict.v1", "V-shared")
	check("v1 identity", got, err, true)
	got, err = view.Ledgered(ctx, "proofbound.obligation-verdict.v2", "V-shared")
	check("v2 identity sharing a v1 id", got, err, false)
	got, err = view.Ledgered(ctx, "proofbound.requirement-review.v1", "RR-one")
	check("review identity", got, err, true)
	// Negatives: a wrong digest, a wrong source, an unknown id.
	got, err = view.CommitClaimed(ctx, sha, "sdd", "CI-add-auth-xuuno2", digest("9"))
	check("claim with a wrong digest", got, err, false)
	got, err = view.RequirementLedgered(ctx, core.SourceIntentRecords, "BR-spec-item-aaaaaa", digest("1"))
	check("requirement under a wrong source", got, err, false)
	got, err = view.Ledgered(ctx, "vera.verdict.v1", "V-unknown")
	check("unknown identity", got, err, false)
}

// Row 12, lazy read, through the real CLI path: syncReviewsOnStore over a repository whose committed
// verdicts are v1 only never reads the ledger; once v2 artifacts are committed it reads it exactly
// once. The store opens on an ephemeral port.
func TestStoreLedgerViewReadsLazilyOnce(t *testing.T) {
	ctx := context.Background()
	passes := 0
	old := readLedgerView
	t.Cleanup(func() { readLedgerView = old })
	readLedgerView = func(ctx context.Context, ledger *store.Store) (*ledgerViewSets, error) {
		passes++
		return old(ctx, ledger)
	}
	ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: ephemeralPort(t), DatabaseURL: os.Getenv("DATABASE_URL")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ledger.Close() })
	ids, err := newIDs()
	if err != nil {
		t.Fatal(err)
	}
	src := repoRoot(t)
	tenant := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = tenant
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	commitVerdicts := func(names ...string) {
		t.Helper()
		for _, name := range names {
			rel := filepath.Join("docs", "verification", "verdicts", name)
			b, err := os.ReadFile(filepath.Join(src, rel))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(tenant, filepath.Dir(rel)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tenant, rel), b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		git("add", "-A", "docs")
		git("commit", "-q", "-m", "verdicts")
	}
	git("init", "-q", "-b", "main")
	commitVerdicts("r2-m3-spec-review-round7-20261119.md")
	if _, err := syncReviewsOnStore(ctx, tenant, ledger, ids); err != nil {
		t.Fatal(err)
	}
	if passes != 0 {
		t.Fatalf("a v1-only reviews sync read the ledger %d times", passes)
	}
	commitVerdicts("p5-obligation-verdict-round1.md", "p5-requirement-review-round1.md", "p7-replay-requirement-review-round1-20261014.md")
	result, err := syncReviewsOnStore(ctx, tenant, ledger, ids)
	if err != nil {
		t.Fatal(err)
	}
	if passes != 1 {
		t.Fatalf("a reviews sync with v2 artifacts read the ledger %d times, want 1", passes)
	}
	if len(result.Held) != 3 {
		t.Fatalf("on a ledger without their prerequisites, the three v2/review artifacts must be held: %+v", result.Held)
	}
}

// selfDigested rewrites a v2 artifact's artifact_path and recomputes its self-digest.
func selfDigested(t *testing.T, b []byte, oldPath, newPath string) []byte {
	t.Helper()
	s := strings.Replace(string(b), `"artifact_path":"`+oldPath+`"`, `"artifact_path":"`+newPath+`"`, 1)
	needle := `"artifact_sha256":"`
	i := strings.LastIndex(s, needle) + len(needle)
	zeroed := s[:i] + strings.Repeat("0", 64) + s[i+64:]
	sum := sha256.Sum256([]byte(zeroed))
	return []byte(s[:i] + hex.EncodeToString(sum[:]) + s[i+64:])
}

// M3 draft 8 §4.6 (CLI-INV-16): `sync reviews` prints its counts line, then held and refused lines
// (paths through printableUnit), exits 0 on holds alone and fails on a refusal, printing first.
func TestSyncReviewsPrintsHoldsAndFailsOnRefusal(t *testing.T) {
	ctx := context.Background()
	port := ephemeralPort(t)
	old := openStore
	t.Cleanup(func() { openStore = old })
	openStore = func(ctx context.Context, root, databaseURL string) (*store.Store, error) {
		return store.Open(ctx, store.Config{Root: stateRoot(root), Port: port, DatabaseURL: databaseURL})
	}
	src := repoRoot(t)
	tenant := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = tenant
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	dir := filepath.Join("docs", "verification", "verdicts")
	write := func(rel string, b []byte) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(tenant, filepath.Dir(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tenant, rel), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	verdict, err := os.ReadFile(filepath.Join(src, dir, "p5-obligation-verdict-round1.md"))
	if err != nil {
		t.Fatal(err)
	}
	git("init", "-q", "-b", "main")
	write(filepath.Join(dir, "p5-obligation-verdict-round1.md"), verdict)
	git("add", "-A", "docs")
	git("commit", "-q", "-m", "one v2 verdict")
	var out strings.Builder
	if err := runCommand(ctx, commandSyncReviews, nil, tenant, "", &out); err != nil {
		t.Fatalf("holds alone failed the sync: %v\n%s", err, out.String())
	}
	held := "held claim-not-ledgered docs/verification/verdicts/p5-obligation-verdict-round1.md\n"
	if !strings.HasPrefix(out.String(), "listed=1 ") || !strings.HasSuffix(out.String(), held) {
		t.Fatalf("output:\n%s", out.String())
	}

	copyPath := "docs/verification/verdicts/p5-obligation-verdict-copy.md"
	write(filepath.FromSlash(copyPath), selfDigested(t, verdict, "docs/verification/verdicts/p5-obligation-verdict-round1.md", copyPath))
	git("add", "-A", "docs")
	git("commit", "-q", "-m", "a second file with the same verdict_id")
	out.Reset()
	err = runCommand(ctx, commandSyncReviews, nil, tenant, "", &out)
	if !errors.Is(err, connectorreviews.ErrRevisionRefused) {
		t.Fatalf("a refusal did not fail the sync: %v\n%s", err, out.String())
	}
	for _, want := range []string{"listed=2 ", "refused duplicate-identity " + copyPath + "\n", "refused duplicate-identity docs/verification/verdicts/p5-obligation-verdict-round1.md\n"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}

	// Draft 8 row 11, CLI half: a malformed artifact AFTER the refusals stops the sync, the error
	// joins both, and the refused lines are still printed.
	write(filepath.Join(dir, "zz-malformed.md"), []byte("---\nschema: vera.verdict.v1\nnot yaml\n"))
	git("add", "-A", "docs")
	git("commit", "-q", "-m", "a malformed artifact after the refusals")
	out.Reset()
	err = runCommand(ctx, commandSyncReviews, nil, tenant, "", &out)
	if !errors.Is(err, connectorreviews.ErrRevisionRefused) || !strings.Contains(err.Error(), "zz-malformed.md") {
		t.Fatalf("malformed after refusal: %v\n%s", err, out.String())
	}
	for _, want := range []string{"listed=3 ", "malformed=1 ", "refused duplicate-identity " + copyPath + "\n"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}
}

// Draft 7 §4.6 "View errors" (row 12): a ledgered git payload that does not decode as
// connectorgit.Commit, or a reviews payload without a schema, fails the view (fail-closed), rather
// than being skipped — skipping a schema-less reviews event would turn a later revision of it into a
// quietly held "first" revision.
func TestStoreLedgerViewFailsClosedOnUndecodablePayloads(t *testing.T) {
	ctx := context.Background()
	for name, event := range map[string]core.NewEventParams{
		"git payload":        {Source: core.SourceGit, NativeID: strings.Repeat("f", 40), Kind: core.KindCommitRecorded, Payload: []byte(`{"sha":1}`)},
		"schema-less review": {Source: core.SourceReviews, NativeID: "V-no-schema", Kind: core.KindReviewVerdict, Payload: []byte(`{"verdict_id":"V-no-schema"}`)},
		"requirement":        {Source: core.SourceIntentSDD, NativeID: "R-bad", Kind: core.KindRequirement, Payload: []byte(`{"requirement_id":1}`)},
	} {
		t.Run(name, func(t *testing.T) {
			ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: ephemeralPort(t), DatabaseURL: os.Getenv("DATABASE_URL")})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = ledger.Close() })
			ids, err := newIDs()
			if err != nil {
				t.Fatal(err)
			}
			run, err := ledger.BeginSync(ctx, "m3-view-errors")
			if err != nil {
				t.Fatal(err)
			}
			event.OccurredAt, event.ConnectorVersion = time.Unix(10, 0).UTC(), "m3-test/1"
			e, err := ids.NewEvent(event)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := run.Append(ctx, e); err != nil {
				t.Fatal(err)
			}
			if err := run.Finish(ctx, json.RawMessage(`[]`), nil); err != nil {
				t.Fatal(err)
			}
			// Every method fails, not only the first to load: a view that kept the partial sets after
			// a failed load would answer the other two from them (mutation review F1, C49).
			for method, ask := range map[string]func(*storeLedgerView) error{
				"Ledgered": func(v *storeLedgerView) error {
					_, err := v.Ledgered(ctx, "proofbound.obligation-verdict.v2", "x")
					return err
				},
				"CommitClaimed": func(v *storeLedgerView) error {
					_, err := v.CommitClaimed(ctx, strings.Repeat("f", 40), "records", "CI-x-abcdef", strings.Repeat("a", 64))
					return err
				},
				"RequirementLedgered": func(v *storeLedgerView) error {
					_, err := v.RequirementLedgered(ctx, core.SourceIntentSDD, "R-bad", strings.Repeat("a", 64))
					return err
				},
			} {
				view := &storeLedgerView{ledger: ledger}
				if err := ask(view); err == nil {
					t.Fatalf("%s skipped an undecodable payload", method)
				}
				if err := ask(view); err == nil {
					t.Fatalf("%s answered from partial sets on a second call", method)
				}
			}
		})
	}
}

// M3 SPEC §4.3: L is read fail-closed too. A ledgered intent.sdd change intent whose payload does not
// decode fails the read, rather than being left out of L (which would quietly turn every claim on it
// into a not-ledgered skip) (review r2-m3-impl-review-mutation-concurrency-20261119 F1, C22).
func TestSDDLedgeredSetFailsClosedOnUndecodablePayload(t *testing.T) {
	ctx := context.Background()
	ledger, err := store.Open(ctx, store.Config{Root: t.TempDir(), Port: ephemeralPort(t), DatabaseURL: os.Getenv("DATABASE_URL")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ledger.Close() })
	ids, err := newIDs()
	if err != nil {
		t.Fatal(err)
	}
	run, err := ledger.BeginSync(ctx, "m3-l-errors")
	if err != nil {
		t.Fatal(err)
	}
	e, err := ids.NewEvent(core.NewEventParams{Source: core.SourceIntentSDD, NativeID: "CI-bad-abcdef", Kind: core.KindChangeIntent, OccurredAt: time.Unix(10, 0).UTC(), ConnectorVersion: "m3-test/1", Payload: []byte(`{"intent_id":1}`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := run.Append(ctx, e); err != nil {
		t.Fatal(err)
	}
	if err := run.Finish(ctx, json.RawMessage(`[]`), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := readSDDLedgered(ctx, ledger); err == nil {
		t.Fatal("L skipped an undecodable change intent")
	}
}
