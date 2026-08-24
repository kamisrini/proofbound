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
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "sync" || args[1] != "checks" {
		fmt.Fprintln(stderr, "usage: vera sync checks")
		return 2
	}
	root, err := repositoryRoot()
	if err == nil {
		err = syncChecks(ctx, root, os.Getenv("DATABASE_URL"), stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "vera sync checks: %v\n", err)
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
		if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr == nil {
			return dir, nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("repository root not found")
		}
		dir = parent
	}
}

func syncChecks(ctx context.Context, repositoryRoot, databaseURL string, output io.Writer) (resultErr error) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: crand.Reader, Now: time.Now})
	if err != nil {
		return err
	}
	connector, err := checks.New(&checks.Deps{
		SpoolDir: filepath.Join(repositoryRoot, ".vera", "spool"),
		IDs:      ids,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		return err
	}
	ledger, err := store.Open(ctx, store.Config{
		Root:        filepath.Join(repositoryRoot, ".vera"),
		DatabaseURL: databaseURL,
	})
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()

	syncRun, err := ledger.BeginSync(ctx, "checks")
	if err != nil {
		return err
	}
	result, syncErr := connector.Sync(ctx, syncRun)
	finishErr := syncRun.Finish(ctx, result.Cursor, syncErr)
	if syncErr != nil || finishErr != nil {
		return errors.Join(syncErr, finishErr)
	}
	_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing)
	return err
}
