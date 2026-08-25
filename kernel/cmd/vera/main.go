package main

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const usage = "usage: vera sync {git|checks|all} | vera rebuild | vera verify"

func main() { os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)) }

type command int

const (
	commandInvalid command = iota
	commandSyncGit
	commandSyncChecks
	commandSyncAll
	commandRebuild
	commandVerify
)

func parseCommand(args []string) command {
	switch {
	case len(args) == 2 && args[0] == "sync" && args[1] == "git":
		return commandSyncGit
	case len(args) == 2 && args[0] == "sync" && args[1] == "checks":
		return commandSyncChecks
	case len(args) == 2 && args[0] == "sync" && args[1] == "all":
		return commandSyncAll
	case len(args) == 1 && args[0] == "rebuild":
		return commandRebuild
	case len(args) == 1 && args[0] == "verify":
		return commandVerify
	default:
		return commandInvalid
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cmd := parseCommand(args)
	if cmd == commandInvalid {
		if len(args) == 2 && args[0] == "sync" && args[1] == "sessions" {
			fmt.Fprintln(stderr, "vera sync sessions: not implemented (Task 7)")
			return 1
		}
		fmt.Fprintln(stderr, usage)
		return 2
	}
	root, err := repositoryRoot()
	if err == nil {
		err = runCommand(ctx, cmd, root, os.Getenv("DATABASE_URL"), stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "vera: %v\n", err)
		return 1
	}
	return 0
}

func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return repositoryRootFrom(dir)
}

func repositoryRootFrom(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("repository root not found")
		}
		dir = parent
	}
}

func openStore(ctx context.Context, root, databaseURL string) (*store.Store, error) {
	return store.Open(ctx, store.Config{Root: filepath.Join(root, ".vera"), DatabaseURL: databaseURL})
}

func newIDs() (*core.IDGenerator, error) {
	return core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
}

func logger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func syncGit(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (connectorgit.Result, error) {
	repo, err := gitcmd.New(root)
	if err != nil {
		return connectorgit.Result{}, err
	}
	connector, err := connectorgit.New(&connectorgit.Deps{Repo: repo, IDs: ids, Logger: logger()})
	if err != nil {
		return connectorgit.Result{}, err
	}
	run, err := ledger.BeginSync(ctx, "git")
	if err != nil {
		return connectorgit.Result{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	return result, errors.Join(syncErr, run.Finish(ctx, result.Cursor, syncErr))
}

type checksResult struct{ Listed, Appended, Existing int }

func syncChecksOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (checksResult, error) {
	connector, err := checks.New(&checks.Deps{SpoolDir: filepath.Join(root, ".vera", "spool"), IDs: ids, Logger: logger()})
	if err != nil {
		return checksResult{}, err
	}
	run, err := ledger.BeginSync(ctx, "checks")
	if err != nil {
		return checksResult{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	finishErr := run.Finish(ctx, result.Cursor, syncErr)
	return checksResult{result.Listed, result.Appended, result.Existing}, errors.Join(syncErr, finishErr)
}

func syncChecks(ctx context.Context, root, databaseURL string, output io.Writer) (resultErr error) {
	ids, err := newIDs()
	if err != nil {
		return err
	}
	ledger, err := openStore(ctx, root, databaseURL)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
	result, err := syncChecksOnStore(ctx, root, ledger, ids)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing)
	return err
}

func runCommand(ctx context.Context, cmd command, root, databaseURL string, output io.Writer) (resultErr error) {
	if cmd == commandSyncChecks {
		return syncChecks(ctx, root, databaseURL, output)
	}
	ledger, err := openStore(ctx, root, databaseURL)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
	projector := projections.New()
	switch cmd {
	case commandRebuild:
		return projector.Rebuild(ctx, ledger)
	case commandSyncGit, commandSyncAll:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		gitResult, err := syncGit(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		if cmd == commandSyncGit {
			_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", gitResult.Listed, gitResult.Appended, gitResult.Existing)
			return err
		}
		checksResult, err := syncChecksOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "git appended=%d checks appended=%d\n", gitResult.Appended, checksResult.Appended)
		return err
	case commandVerify:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		return verify(ctx, root, ledger, projector, ids)
	}
	return nil
}

func verify(ctx context.Context, root string, ledger *store.Store, projector *projections.Projector, ids *core.IDGenerator) error {
	if _, err := syncGit(ctx, root, ledger, ids); err != nil {
		return err
	}
	if _, err := syncChecksOnStore(ctx, root, ledger, ids); err != nil {
		return err
	}
	secondGit, err := syncGit(ctx, root, ledger, ids)
	if err != nil {
		return err
	}
	secondChecks, err := syncChecksOnStore(ctx, root, ledger, ids)
	if err != nil {
		return err
	}
	if secondGit.Appended != 0 || secondChecks.Appended != 0 {
		return fmt.Errorf("verify: second sync appended git=%d checks=%d", secondGit.Appended, secondChecks.Appended)
	}
	if err := projector.Apply(ctx, ledger); err != nil {
		return err
	}
	before, err := projector.Snapshot(ctx, ledger)
	if err != nil {
		return err
	}
	if err := projector.Rebuild(ctx, ledger); err != nil {
		return err
	}
	after, err := projector.Snapshot(ctx, ledger)
	if err != nil {
		return err
	}
	if err := projections.CompareSnapshots(before, after); err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	witness, err := latestSpoolWitness(root)
	if err != nil {
		return err
	}
	latestLedgerRunID, err := latestLedgerCheckRunID(ctx, ledger)
	if err != nil {
		return err
	}
	if latestLedgerRunID != witness.RunID {
		return fmt.Errorf("verify: latest spool witness %s is not the latest ledger check.run %s", witness.RunID, latestLedgerRunID)
	}
	return nil
}

// latestSpoolWitness selects the lexically greatest witness filename. The
// witness script uses ULIDs, whose lexical order is chronological, and the
// checks connector ingests the same filename set in this order.
func latestSpoolWitness(root string) (checks.Witness, error) {
	spoolDir := filepath.Join(root, ".vera", "spool")
	entries, err := os.ReadDir(spoolDir)
	if errors.Is(err, os.ErrNotExist) {
		return checks.Witness{}, errors.New("verify: no make check witness found in spool")
	}
	if err != nil {
		return checks.Witness{}, fmt.Errorf("verify: read witness spool: %w", err)
	}
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)
	if len(filenames) == 0 {
		return checks.Witness{}, errors.New("verify: no make check witness found in spool")
	}
	filename := filenames[len(filenames)-1]
	data, err := os.ReadFile(filepath.Join(spoolDir, filename))
	if err != nil {
		return checks.Witness{}, fmt.Errorf("verify: read latest witness %s: %w", filename, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var witness checks.Witness
	if err := decoder.Decode(&witness); err != nil {
		return checks.Witness{}, fmt.Errorf("verify: decode latest witness %s: %w", filename, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return checks.Witness{}, fmt.Errorf("verify: latest witness %s contains trailing JSON", filename)
		}
		return checks.Witness{}, fmt.Errorf("verify: latest witness %s has invalid trailing data: %w", filename, err)
	}
	if filename != witness.RunID+".json" || witness.Schema != "vera.witness.v1" || witness.Command != "make check" || witness.RunID == "" {
		return checks.Witness{}, fmt.Errorf("verify: latest spool file %s is not a valid make check witness", filename)
	}
	return witness, nil
}

func latestLedgerCheckRunID(ctx context.Context, ledger *store.Store) (string, error) {
	var latest string
	if err := ledger.ReadEvents(ctx, store.Filter{Source: core.SourceChecks, Kind: core.KindCheckRun}, func(record store.Record) error {
		latest = record.Event.NativeID
		return nil
	}); err != nil {
		return "", err
	}
	if latest == "" {
		return "", errors.New("verify: no check.run event found")
	}
	return latest, nil
}
