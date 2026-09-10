package cli

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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorgithub "github.com/kamisrini/proofbound/kernel/internal/connector/github"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	intentrecords "github.com/kamisrini/proofbound/kernel/internal/connector/intent/records"
	intentspecdir "github.com/kamisrini/proofbound/kernel/internal/connector/intent/specdir"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	connectorsessions "github.com/kamisrini/proofbound/kernel/internal/connector/sessions"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const usage = "usage: proofbound sync {git|checks|sessions|reviews|github|all} | proofbound sync intent {records|specdir|all} | proofbound rebuild | proofbound verify | proofbound report {week|github|intent <id>|requirement <id>} | proofbound intent check --commit <sha> | proofbound gates {canary|enforce}"

const legacyAdvisory = "proofbound: deprecated VERA identity alias used; switch to Proofbound before 2026-12-31"

const verifyTimeout = 15 * time.Minute

// Run executes the shared command implementation for the live executable or its one-phase alias.
func Run(ctx context.Context, program string, args []string, stdout, stderr io.Writer) int {
	if strings.EqualFold(filepath.Base(program), "vera") {
		fmt.Fprintln(stderr, legacyAdvisory)
	}
	warnLegacyEnvironment(stderr)
	return run(ctx, args, stdout, stderr)
}

type command int

const (
	commandInvalid command = iota
	commandSyncGit
	commandSyncChecks
	commandSyncSessions
	commandSyncReviews
	commandSyncGitHub
	commandSyncAll
	commandSyncIntentRecords
	commandSyncIntentSpecdir
	commandSyncIntentAll
	commandRebuild
	commandVerify
	commandReportWeek
	commandReportGitHub
	commandReportIntent
	commandReportRequirement
	commandIntentCheck
	commandGatesCanary
	commandGatesEnforce
)

func parseCommand(args []string) command {
	switch {
	case len(args) == 2 && args[0] == "sync" && args[1] == "git":
		return commandSyncGit
	case len(args) == 2 && args[0] == "sync" && args[1] == "checks":
		return commandSyncChecks
	case len(args) == 2 && args[0] == "sync" && args[1] == "sessions":
		return commandSyncSessions
	case len(args) == 2 && args[0] == "sync" && args[1] == "reviews":
		return commandSyncReviews
	case len(args) == 2 && args[0] == "sync" && args[1] == "github":
		return commandSyncGitHub
	case len(args) == 2 && args[0] == "sync" && args[1] == "all":
		return commandSyncAll
	case len(args) == 3 && args[0] == "sync" && args[1] == "intent" && args[2] == "records":
		return commandSyncIntentRecords
	case len(args) == 3 && args[0] == "sync" && args[1] == "intent" && args[2] == "specdir":
		return commandSyncIntentSpecdir
	case len(args) == 3 && args[0] == "sync" && args[1] == "intent" && args[2] == "all":
		return commandSyncIntentAll
	case len(args) == 1 && args[0] == "rebuild":
		return commandRebuild
	case len(args) == 1 && args[0] == "verify":
		return commandVerify
	case len(args) == 2 && args[0] == "report" && args[1] == "week":
		return commandReportWeek
	case len(args) == 2 && args[0] == "report" && args[1] == "github":
		return commandReportGitHub
	case len(args) == 3 && args[0] == "report" && args[1] == "intent" && args[2] != "":
		return commandReportIntent
	case len(args) == 3 && args[0] == "report" && args[1] == "requirement" && args[2] != "":
		return commandReportRequirement
	case len(args) == 4 && args[0] == "intent" && args[1] == "check" && args[2] == "--commit" && args[3] != "":
		return commandIntentCheck
	case len(args) == 2 && args[0] == "gates" && args[1] == "canary":
		return commandGatesCanary
	case len(args) == 2 && args[0] == "gates" && args[1] == "enforce":
		return commandGatesEnforce
	default:
		return commandInvalid
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cmd := parseCommand(args)
	if cmd == commandInvalid {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	root, err := repositoryRoot()
	if err == nil {
		if cmd == commandVerify {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, verifyTimeout)
			defer cancel()
		}
		err = runCommand(ctx, cmd, args, root, os.Getenv("DATABASE_URL"), stdout)
	}
	if err != nil {
		if cmd == commandVerify && errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("verify: timed out after %s: %w", verifyTimeout, err)
		}
		fmt.Fprintf(stderr, "proofbound: %v\n", err)
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
	return store.Open(ctx, store.Config{Root: stateRoot(root), DatabaseURL: databaseURL})
}

func stateRoot(root string) string { return filepath.Join(root, ".proofbound") }

func productEnv(live, legacy string) string {
	if value, ok := os.LookupEnv(live); ok {
		return value
	}
	return os.Getenv(legacy)
}

func warnLegacyEnvironment(stderr io.Writer) {
	for _, names := range [][2]string{
		{"PROOFBOUND_GITHUB_OWNER", "VERA_GITHUB_OWNER"},
		{"PROOFBOUND_GITHUB_REPOS", "VERA_GITHUB_REPOS"},
		{"PROOFBOUND_GITHUB_API_BASE_URL", "VERA_GITHUB_API_BASE_URL"},
		{"PROOFBOUND_VERIFY_TRACE", "VERA_VERIFY_TRACE"},
	} {
		if _, liveSet := os.LookupEnv(names[0]); !liveSet {
			if _, legacySet := os.LookupEnv(names[1]); legacySet {
				fmt.Fprintf(stderr, "%s (%s -> %s)\n", legacyAdvisory, names[1], names[0])
			}
		}
	}
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
	reader := committedIntentReader{root: root}
	now := time.Now().UTC()
	recordsProvider, err := intentrecords.New(reader, now)
	if err != nil {
		return connectorgit.Result{}, err
	}
	specdirProvider, err := intentspecdir.New(reader, now)
	if err != nil {
		return connectorgit.Result{}, err
	}
	resolver := commitIntentResolver{records: recordsProvider, specdir: specdirProvider}
	connector, err := connectorgit.New(&connectorgit.Deps{Repo: repo, IDs: ids, Logger: logger(), Resolver: resolver})
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

type commitIntentResolver struct {
	records *intentrecords.Provider
	specdir *intentspecdir.Provider
}

func (r commitIntentResolver) Resolve(ctx context.Context, commitSHA, provider, recordID string) (connectorgit.IntentRef, error) {
	var revision connectorintent.Revision
	var err error
	switch provider {
	case "records":
		revision, err = r.records.Resolve(ctx, commitSHA, recordID)
	case "specdir":
		revision, err = r.specdir.Resolve(ctx, commitSHA, recordID)
	default:
		return connectorgit.IntentRef{}, fmt.Errorf("unknown intent provider %q", provider)
	}
	if err != nil {
		return connectorgit.IntentRef{}, err
	}
	return connectorgit.IntentRef{Provider: provider, RecordID: revision.NativeID, ArtifactSHA256: revision.ArtifactSHA256}, nil
}

type checksResult struct{ Listed, Appended, Existing int }

func syncChecksOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (checksResult, error) {
	connector, err := checks.New(&checks.Deps{SpoolDir: filepath.Join(stateRoot(root), "spool"), IDs: ids, Logger: logger()})
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

type sessionsResult struct{ Listed, Appended, Existing, Skipped int }

type committedVerdictReader struct{ root string }

func (r committedVerdictReader) ReadCommittedVerdicts(ctx context.Context) ([]connectorreviews.Artifact, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", r.root, "ls-tree", "-r", "--name-only", "HEAD", "--", "docs/verification/verdicts")
	cmd.Env = repositoryGitEnv()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list committed verdicts: %w", err)
	}
	var artifacts []connectorreviews.Artifact
	for _, name := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if name == "" || !strings.HasSuffix(name, ".md") {
			continue
		}
		show := exec.CommandContext(ctx, "git", "-C", r.root, "show", "HEAD:"+name)
		show.Env = repositoryGitEnv()
		data, err := show.Output()
		if err != nil {
			return nil, fmt.Errorf("read committed verdict %s: %w", name, err)
		}
		artifacts = append(artifacts, connectorreviews.Artifact{Path: name, Bytes: data})
	}
	return artifacts, nil
}

func repositoryGitEnv() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "GIT_") {
			env = append(env, value)
		}
	}
	return env
}

type reviewsResult struct{ Listed, Appended, Existing, Malformed, Documentary int }

type committedIntentReader struct{ root string }

func (r committedIntentReader) ReadIntentArtifacts(ctx context.Context, tree string) ([]intentrecords.Artifact, error) {
	items, err := r.read(ctx, tree, "docs/intent/records")
	if err != nil {
		return nil, err
	}
	out := make([]intentrecords.Artifact, len(items))
	for i, item := range items {
		out[i] = intentrecords.Artifact{Path: item.Path, Bytes: item.Bytes}
	}
	return out, nil
}

func (r committedIntentReader) ReadSpecArtifacts(ctx context.Context, tree string) ([]intentspecdir.Artifact, error) {
	items, err := r.read(ctx, tree, "specs")
	if err != nil {
		return nil, err
	}
	out := make([]intentspecdir.Artifact, len(items))
	for i, item := range items {
		out[i] = intentspecdir.Artifact{Path: item.Path, Bytes: item.Bytes}
	}
	return out, nil
}

func (r committedIntentReader) read(ctx context.Context, tree, prefix string) ([]connectorreviews.Artifact, error) {
	if tree == "" {
		return nil, errors.New("committed intent reader: tree is required")
	}
	cmd := exec.CommandContext(ctx, "git", "-C", r.root, "ls-tree", "-r", "--name-only", "-z", tree, "--", prefix)
	cmd.Env = repositoryGitEnv()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list committed intent artifacts at %s: %w", tree, err)
	}
	var artifacts []connectorreviews.Artifact
	for _, raw := range bytes.Split(output, []byte{0}) {
		name := string(raw)
		if name == "" || (!strings.HasSuffix(name, ".md")) {
			continue
		}
		show := exec.CommandContext(ctx, "git", "-C", r.root, "show", tree+":"+name)
		show.Env = repositoryGitEnv()
		data, err := show.Output()
		if err != nil {
			return nil, fmt.Errorf("read committed intent artifact %s: %w", name, err)
		}
		artifacts = append(artifacts, connectorreviews.Artifact{Path: name, Bytes: data})
	}
	return artifacts, nil
}

func syncIntentOnStore(ctx context.Context, root, selection string, ledger *store.Store, ids *core.IDGenerator) (connectorintent.Result, error) {
	reader := committedIntentReader{root: root}
	now := time.Now().UTC()
	recordsProvider, err := intentrecords.New(reader, now)
	if err != nil {
		return connectorintent.Result{}, err
	}
	specdirProvider, err := intentspecdir.New(reader, now)
	if err != nil {
		return connectorintent.Result{}, err
	}
	connector, err := connectorintent.New(&connectorintent.Deps{Providers: []connectorintent.Provider{recordsProvider, specdirProvider}, IDs: ids, Logger: logger()})
	if err != nil {
		return connectorintent.Result{}, err
	}
	run, err := ledger.BeginSync(ctx, "intent."+selection)
	if err != nil {
		return connectorintent.Result{}, err
	}
	result, syncErr := connector.Sync(ctx, selection, run)
	return result, errors.Join(syncErr, run.Finish(ctx, result.Cursor, syncErr))
}

func syncReviewsOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (reviewsResult, error) {
	connector, err := connectorreviews.New(&connectorreviews.Deps{Reader: committedVerdictReader{root: root}, IDs: ids, Logger: logger()})
	if err != nil {
		return reviewsResult{}, err
	}
	run, err := ledger.BeginSync(ctx, "reviews")
	if err != nil {
		return reviewsResult{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	finishErr := run.Finish(ctx, result.Cursor, syncErr)
	return reviewsResult{result.Listed, result.Appended, result.Existing, result.Malformed, result.Documentary}, errors.Join(syncErr, finishErr)
}

type githubResult struct{ Listed, Appended, Existing int }

func syncGitHubOnStore(ctx context.Context, ledger *store.Store, ids *core.IDGenerator) (githubResult, error) {
	owner := strings.TrimSpace(productEnv("PROOFBOUND_GITHUB_OWNER", "VERA_GITHUB_OWNER"))
	var repos []string
	for _, repo := range strings.Split(productEnv("PROOFBOUND_GITHUB_REPOS", "VERA_GITHUB_REPOS"), ",") {
		if repo = strings.TrimSpace(repo); repo != "" {
			repos = append(repos, repo)
		}
	}
	connector, err := connectorgithub.New(&connectorgithub.Deps{
		API:   &connectorgithub.HTTPClient{BaseURL: productEnv("PROOFBOUND_GITHUB_API_BASE_URL", "VERA_GITHUB_API_BASE_URL"), Token: os.Getenv("GITHUB_TOKEN")},
		Owner: owner, Repos: repos, IDs: ids, Logger: logger(),
	})
	if err != nil {
		return githubResult{}, err
	}
	run, err := ledger.BeginSync(ctx, "github")
	if err != nil {
		return githubResult{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	finishErr := run.Finish(ctx, result.Cursor, syncErr)
	return githubResult{result.Listed, result.Appended, result.Existing}, errors.Join(syncErr, finishErr)
}

func syncSessionsOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (sessionsResult, error) {
	connector, err := connectorsessions.New(&connectorsessions.Deps{Root: root, IDs: ids, Logger: logger()})
	if err != nil {
		return sessionsResult{}, err
	}
	run, err := ledger.BeginSync(ctx, "sessions")
	if err != nil {
		return sessionsResult{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	finishErr := run.Finish(ctx, result.Cursor, syncErr)
	return sessionsResult{result.Listed, result.Appended, result.Existing, result.Skipped}, errors.Join(syncErr, finishErr)
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

func runCommand(ctx context.Context, cmd command, args []string, root, databaseURL string, output io.Writer) (resultErr error) {
	if cmd == commandSyncChecks {
		return syncChecks(ctx, root, databaseURL, output)
	}
	if cmd == commandSyncSessions {
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
		result, err := syncSessionsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d skipped=%d\n", result.Listed, result.Appended, result.Existing, result.Skipped)
		return err
	}
	if cmd == commandSyncReviews {
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
		result, err := syncReviewsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d malformed=%d documentary=%d\n", result.Listed, result.Appended, result.Existing, result.Malformed, result.Documentary)
		return err
	}
	if cmd == commandSyncGitHub {
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
		result, err := syncGitHubOnStore(ctx, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing)
		return err
	}
	if cmd == commandSyncIntentRecords || cmd == commandSyncIntentSpecdir || cmd == commandSyncIntentAll {
		selection := map[command]string{commandSyncIntentRecords: "records", commandSyncIntentSpecdir: "specdir", commandSyncIntentAll: "all"}[cmd]
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
		result, err := syncIntentOnStore(ctx, root, selection, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing)
		return err
	}
	ledger, err := openStore(ctx, root, databaseURL)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, ledger.Close()) }()
	projector := projections.New()
	if cmd == commandGatesCanary || cmd == commandGatesEnforce {
		definitions, err := gates.LoadDir(filepath.Join(root, "gates"))
		if err != nil {
			return err
		}
		if cmd == commandGatesEnforce {
			if err := gates.RequireDefinitions(definitions); err != nil {
				return err
			}
		}
		if err := resolveIntentGateHeadScope(ctx, root, definitions); err != nil {
			return err
		}
		blocked := false
		results := make([]gates.Result, 0, len(definitions))
		for _, definition := range definitions {
			if cmd == commandGatesEnforce {
				if err := definition.EnforceReady(); err != nil {
					return err
				}
			}
			result, err := gates.Evaluate(ctx, ledger, definition)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(output, "gate=%s state=%s seq=%d proof=%s would_block=%t\n", result.GateID, result.State, result.Seq, result.EventID, result.WouldBlock); err != nil {
				return err
			}
			if cmd == commandGatesEnforce && gates.Enforce(result) != nil {
				blocked = true
			}
			results = append(results, result)
		}
		if cmd == commandGatesEnforce {
			if err := enforceGateResults(definitions, results); err != nil {
				return err
			}
		}
		if blocked {
			return errors.New("gate enforcement blocked: a gate is BLOCKED or UNKNOWN")
		}
		return nil
	}
	switch cmd {
	case commandRebuild:
		return projector.Rebuild(ctx, ledger)
	case commandSyncGit, commandSyncAll:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		var intentResult connectorintent.Result
		if cmd == commandSyncAll {
			intentResult, err = syncIntentOnStore(ctx, root, "all", ledger, ids)
			if err != nil {
				return err
			}
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
		sessionsResult, err := syncSessionsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		reviewsResult, err := syncReviewsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "intent appended=%d git appended=%d checks appended=%d sessions appended=%d reviews appended=%d\n", intentResult.Appended, gitResult.Appended, checksResult.Appended, sessionsResult.Appended, reviewsResult.Appended)
		return err
	case commandVerify:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		return verify(ctx, root, ledger, projector, ids)
	case commandReportWeek:
		repo, err := gitcmd.New(root)
		if err != nil {
			return err
		}
		reachable, err := repo.Reachable(ctx)
		if err != nil {
			return err
		}
		if err := projector.Apply(ctx, ledger); err != nil {
			return err
		}
		return projector.ReportWeek(ctx, ledger, time.Now(), reachable, output)
	case commandReportGitHub:
		return projector.ReportGitHub(ctx, ledger, time.Now(), output)
	case commandReportIntent:
		return projector.ReportIntent(ctx, ledger, args[2], time.Now(), output)
	case commandReportRequirement:
		return projector.ReportRequirement(ctx, ledger, args[2], output)
	case commandIntentCheck:
		return projector.CheckIntent(ctx, ledger, args[3], output)
	}
	return nil
}

func resolveIntentGateHeadScope(ctx context.Context, root string, definitions []gates.Definition) error {
	needsHead := false
	for _, definition := range definitions {
		needsHead = needsHead || (definition.Rule == "intent-delivery-readiness" && definition.ScopeCommit == "HEAD")
	}
	if !needsHead {
		return nil
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--verify", "HEAD^{commit}")
	cmd.Env = repositoryGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("resolve delivery gate HEAD: %w", err)
	}
	sha := strings.TrimSpace(string(out))
	for i := range definitions {
		if definitions[i].Rule == "intent-delivery-readiness" && definitions[i].ScopeCommit == "HEAD" {
			definitions[i].ScopeCommit = sha
		}
	}
	return nil
}

func enforceGateResults(definitions []gates.Definition, results []gates.Result) error {
	if err := gates.RequireDefinitions(definitions); err != nil {
		return err
	}
	if len(definitions) != len(results) {
		return errors.New("gate enforcement blocked: incomplete evaluation")
	}
	for i, definition := range definitions {
		if err := definition.EnforceReady(); err != nil {
			return err
		}
		if err := gates.Enforce(results[i]); err != nil {
			return err
		}
	}
	return nil
}

func verify(ctx context.Context, root string, ledger *store.Store, projector *projections.Projector, ids *core.IDGenerator) error {
	if err := verifyStep(ctx, "initial intent sync", func() error { _, err := syncIntentOnStore(ctx, root, "all", ledger, ids); return err }); err != nil {
		return err
	}
	if err := verifyStep(ctx, "initial git sync", func() error { _, err := syncGit(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	if err := verifyStep(ctx, "initial checks sync", func() error { _, err := syncChecksOnStore(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	if err := verifyStep(ctx, "initial sessions sync", func() error { _, err := syncSessionsOnStore(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	if err := verifyStep(ctx, "initial reviews sync", func() error { _, err := syncReviewsOnStore(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	var secondGit connectorgit.Result
	var secondIntent connectorintent.Result
	if err := verifyStep(ctx, "idempotence intent sync", func() error {
		var err error
		secondIntent, err = syncIntentOnStore(ctx, root, "all", ledger, ids)
		return err
	}); err != nil {
		return err
	}
	if err := verifyStep(ctx, "idempotence git sync", func() error { var err error; secondGit, err = syncGit(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	var secondChecks checksResult
	if err := verifyStep(ctx, "idempotence checks sync", func() error { var err error; secondChecks, err = syncChecksOnStore(ctx, root, ledger, ids); return err }); err != nil {
		return err
	}
	var secondSessions sessionsResult
	if err := verifyStep(ctx, "idempotence sessions sync", func() error {
		var err error
		secondSessions, err = syncSessionsOnStore(ctx, root, ledger, ids)
		return err
	}); err != nil {
		return err
	}
	var secondReviews reviewsResult
	if err := verifyStep(ctx, "idempotence reviews sync", func() error {
		var err error
		secondReviews, err = syncReviewsOnStore(ctx, root, ledger, ids)
		return err
	}); err != nil {
		return err
	}
	if secondIntent.Appended != 0 || secondGit.Appended != 0 || secondChecks.Appended != 0 || secondSessions.Appended != 0 || secondReviews.Appended != 0 {
		return fmt.Errorf("verify: second sync appended intent=%d git=%d checks=%d sessions=%d reviews=%d", secondIntent.Appended, secondGit.Appended, secondChecks.Appended, secondSessions.Appended, secondReviews.Appended)
	}
	if err := verifyStep(ctx, "projection apply", func() error { return projector.Apply(ctx, ledger) }); err != nil {
		return err
	}
	var before projections.Snapshot
	if err := verifyStep(ctx, "projection snapshot", func() error { var err error; before, err = projector.Snapshot(ctx, ledger); return err }); err != nil {
		return err
	}
	if err := verifyStep(ctx, "projection rebuild", func() error { return projector.Rebuild(ctx, ledger) }); err != nil {
		return err
	}
	var after projections.Snapshot
	if err := verifyStep(ctx, "rebuilt projection snapshot", func() error { var err error; after, err = projector.Snapshot(ctx, ledger); return err }); err != nil {
		return err
	}
	if err := projections.CompareSnapshots(before, after); err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	var witness checks.Witness
	if err := verifyStep(ctx, "witness read", func() error { var err error; witness, err = latestSpoolWitness(root); return err }); err != nil {
		return err
	}
	var latestLedgerRunID string
	if err := verifyStep(ctx, "ledger witness lookup", func() error { var err error; latestLedgerRunID, err = latestLedgerCheckRunID(ctx, ledger); return err }); err != nil {
		return err
	}
	if latestLedgerRunID != witness.RunID {
		return fmt.Errorf("verify: latest spool witness %s is not the latest ledger check.run %s", witness.RunID, latestLedgerRunID)
	}
	return nil
}

func verifyStep(ctx context.Context, name string, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("verify: %s: %w", name, err)
	}
	if productEnv("PROOFBOUND_VERIFY_TRACE", "VERA_VERIFY_TRACE") == "1" {
		fmt.Fprintf(os.Stderr, "verify: begin %s\n", name)
	}
	if err := fn(); err != nil {
		return fmt.Errorf("verify: %s: %w", name, err)
	}
	return nil
}

// latestSpoolWitness selects the lexically greatest witness filename. The
// witness script uses ULIDs, whose lexical order is chronological, and the
// checks connector ingests the same filename set in this order.
func latestSpoolWitness(root string) (checks.Witness, error) {
	spoolDir := filepath.Join(stateRoot(root), "spool")
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
