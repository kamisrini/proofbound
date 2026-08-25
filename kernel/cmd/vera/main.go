package main

import (
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
	found := false
	if err := ledger.ReadEvents(ctx, store.Filter{Source: core.SourceChecks, Kind: core.KindCheckRun}, func(store.Record) error { found = true; return nil }); err != nil {
		return err
	}
	if !found {
		return errors.New("verify: no check.run event found")
	}
	return nil
}
