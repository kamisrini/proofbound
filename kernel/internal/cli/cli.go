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
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorgithub "github.com/kamisrini/proofbound/kernel/internal/connector/github"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	intentrecords "github.com/kamisrini/proofbound/kernel/internal/connector/intent/records"
	intentsdd "github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	intentspecdir "github.com/kamisrini/proofbound/kernel/internal/connector/intent/specdir"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	connectorsessions "github.com/kamisrini/proofbound/kernel/internal/connector/sessions"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
	"github.com/kamisrini/proofbound/kernel/internal/migration"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	"github.com/kamisrini/proofbound/kernel/internal/twin"
)

const usage = "usage: proofbound version | proofbound task [start <change>|stop] | proofbound ui [--base <ref>] [--port <n>] [--open] | proofbound ui --bundle <dir> [--trust-key <fingerprint>] [--repo <clone>] [--port <n>] [--open] | proofbound check <base> [--markdown] [--require-claims] | proofbound sync {git|checks|sessions|reviews|github|all} | proofbound sync intent {records|specdir|sdd|all} | proofbound migrate historical-evidence | proofbound rebuild | proofbound verify | proofbound report {week|github|drift|intent <id>|requirement <id>} | proofbound report agents <base> [--json] | proofbound intent check --commit <sha> | proofbound gates {canary|enforce} | proofbound twin replay [--through <seq>] [--candidate <file>] | proofbound init {openspec|speckit} [--owner <name>] | proofbound review {draft <commit>|seal <draft-dir>} | proofbound keys init [--name <name>] | proofbound keys git-sign | proofbound waiver sign <id> --path <pattern> [--path <pattern>]... --reason <text> --expires <YYYY-MM-DD> | proofbound warranty export <commit> [-o <file>] | proofbound warranty verify <file> [--repo <path>] [--trust-key <fingerprint>] | proofbound warranty cosign <file> [-o <file>] | proofbound warranty bundle <base> [-o <dir>]"

const verifyTimeout = 15 * time.Minute

// Version is stamped at release build time with -ldflags "-X …/internal/cli.Version=<v>"
// (scripts/release-build.sh); a plain go build reports "dev" (CLI-INV-25).
var Version = "dev"

var workingDirectory = os.Getwd

// Run executes the shared command implementation for the live Proofbound executable.
func Run(ctx context.Context, _ string, args []string, stdout, stderr io.Writer) int {
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
	commandMigrateHistoricalEvidence
	commandTwinReplay
	commandSyncIntentSDD
	commandInitOpenSpec
	commandReviewDraft
	commandReviewSeal
	commandReportDrift
	commandKeysInit
	commandWarrantyExport
	commandWarrantyVerify
	commandVersion
	commandCheck
	commandWarrantyCosign
	commandWarrantyBundle
	commandTask
	commandHook
	commandUI
	commandKeysGitSign
	commandWaiverSign
	commandReportAgents
)

func parseCommand(args []string) command {
	switch {
	case len(args) == 1 && args[0] == "version":
		return commandVersion
	case len(args) >= 1 && args[0] == "task":
		if _, _, ok := parseTask(args); ok {
			return commandTask
		}
		return commandInvalid
	case len(args) >= 1 && args[0] == "hook":
		if _, _, _, ok := parseHook(args); ok {
			return commandHook
		}
		return commandInvalid
	case len(args) >= 1 && args[0] == "ui":
		if _, ok := parseUI(args); ok {
			return commandUI
		}
		return commandInvalid
	case len(args) >= 1 && args[0] == "check":
		if _, _, _, ok := parseCheck(args); ok {
			return commandCheck
		}
		return commandInvalid
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
	case len(args) == 3 && args[0] == "sync" && args[1] == "intent" && args[2] == "sdd":
		return commandSyncIntentSDD
	case len(args) == 3 && args[0] == "sync" && args[1] == "intent" && args[2] == "all":
		return commandSyncIntentAll
	case len(args) == 1 && args[0] == "rebuild":
		return commandRebuild
	case len(args) == 1 && args[0] == "verify":
		return commandVerify
	case len(args) == 2 && args[0] == "report" && args[1] == "week":
		return commandReportWeek
	case len(args) == 2 && args[0] == "report" && args[1] == "drift":
		return commandReportDrift
	case (len(args) == 3 || len(args) == 4 && args[3] == "--json") && args[0] == "report" && args[1] == "agents" && args[2] != "" && !strings.HasPrefix(args[2], "-"):
		return commandReportAgents
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
	case len(args) == 2 && args[0] == "migrate" && args[1] == "historical-evidence":
		return commandMigrateHistoricalEvidence
	case len(args) >= 2 && args[0] == "twin" && args[1] == "replay":
		return commandTwinReplay
	case (len(args) == 2 || (len(args) == 4 && args[2] == "--owner" && args[3] != "")) && args[0] == "init" && (args[1] == "openspec" || args[1] == "speckit"):
		return commandInitOpenSpec
	case len(args) == 3 && args[0] == "review" && args[1] == "draft" && args[2] != "":
		return commandReviewDraft
	case len(args) == 3 && args[0] == "review" && args[1] == "seal" && args[2] != "":
		return commandReviewSeal
	case args0(args, "waiver", "sign"):
		if _, _, _, _, ok := parseWaiverSign(args); ok {
			return commandWaiverSign
		}
		return commandInvalid
	case len(args) == 2 && args[0] == "keys" && args[1] == "git-sign":
		return commandKeysGitSign
	case args0(args, "keys", "init") && flagsOK(args[2:], "--name"):
		return commandKeysInit
	case args0(args, "warranty", "export") && len(args) >= 3 && args[2] != "" && !strings.HasPrefix(args[2], "-") && flagsOK(args[3:], "-o"):
		return commandWarrantyExport
	case args0(args, "warranty", "bundle") && len(args) >= 3 && args[2] != "" && !strings.HasPrefix(args[2], "-") && flagsOK(args[3:], "-o"):
		return commandWarrantyBundle
	case args0(args, "warranty", "cosign") && len(args) >= 3 && args[2] != "" && !strings.HasPrefix(args[2], "-") && flagsOK(args[3:], "-o"):
		return commandWarrantyCosign
	case args0(args, "warranty", "verify") && len(args) >= 3 && args[2] != "" && !strings.HasPrefix(args[2], "-") && flagsOK(args[3:], "--repo", "--trust-key"):
		return commandWarrantyVerify
	default:
		return commandInvalid
	}
}

func args0(args []string, a, b string) bool { return len(args) >= 2 && args[0] == a && args[1] == b }

// flagsOK accepts zero or more `<flag> <value>` pairs, each flag at most once and from allowed.
func flagsOK(rest []string, allowed ...string) bool {
	if len(rest)%2 != 0 {
		return false
	}
	seen := map[string]bool{}
	for i := 0; i < len(rest); i += 2 {
		ok := false
		for _, a := range allowed {
			ok = ok || rest[i] == a
		}
		if !ok || seen[rest[i]] || rest[i+1] == "" {
			return false
		}
		seen[rest[i]] = true
	}
	return true
}

// flagValue returns the value of flag in rest (validated by flagsOK), or "".
func flagValue(rest []string, flag string) string {
	for i := 0; i+1 < len(rest); i += 2 {
		if rest[i] == flag {
			return rest[i+1]
		}
	}
	return ""
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cmd := parseCommand(args)
	if cmd == commandInvalid {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if cmd == commandVersion {
		// Runs anywhere: no repository root, no store, no network (CLI-INV-25).
		fmt.Fprintf(stdout, "proofbound %s\n", Version)
		return 0
	}
	if cmd == commandWarrantyVerify {
		// A verifier may be anywhere: no repository root, no database (W-INV-7).
		return warrantyVerify(ctx, args[2], flagValue(args[3:], "--repo"), flagValue(args[3:], "--trust-key"), stdout)
	}
	if opts, _ := parseUI(args); cmd == commandUI && opts.bundle != "" {
		// A bundle is read anywhere: no repository root, no store (CLI-INV-60).
		if err := runUIBundle(ctx, opts, stdout); err != nil {
			fmt.Fprintf(stderr, "proofbound: %v\n", err)
			return 1
		}
		return 0
	}
	if cmd == commandHook && len(args) == 2 && args[1] == "pre-tool-use" {
		// The write-time perimeter reads its repository from Claude Code's input, not the cwd.
		return runPerimeterHook(ctx, stdout)
	}
	if cmd == commandHook && len(args) == 2 && args[1] == "post-tool-use" {
		// Session-to-line provenance (CLI-INV-70): records the lines an agent's edit added.
		return runPostToolHook(ctx)
	}
	root, err := repositoryRoot()
	if err == nil && cmd == commandHook {
		// The git hooks `init` installs: no store, messages to stderr, exit 1 refuses the commit.
		hook, file, source, _ := parseHook(args)
		if err = runHook(ctx, root, hook, file, source, stderr); errors.Is(err, errHookBlock) {
			return 1
		}
	} else if err == nil {
		if isVerificationCommand(cmd) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, verifyTimeout)
			defer cancel()
		}
		err = runCommand(ctx, cmd, args, root, os.Getenv("DATABASE_URL"), stdout)
	}
	if errors.Is(err, errBundleIncomplete) {
		// Written, but some claiming commit is not VALID (mvp-4 spec §2.6): exit 3, not 1.
		fmt.Fprintf(stderr, "proofbound: %v\n", err)
		return 3
	}
	if err != nil {
		if isVerificationCommand(cmd) && errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("verify: timed out after %s: %w", verifyTimeout, err)
		}
		fmt.Fprintf(stderr, "proofbound: %v\n", err)
		return 1
	}
	return 0
}

func isVerificationCommand(cmd command) bool { return cmd == commandVerify }

func repositoryRoot() (string, error) {
	dir, err := workingDirectory()
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

var openStore = func(ctx context.Context, root, databaseURL string) (*store.Store, error) {
	return store.Open(ctx, store.Config{Root: stateRoot(root), DatabaseURL: databaseURL})
}

var openHistoricalEvidenceStore = func(ctx context.Context, root, databaseURL string) (*store.Store, error) {
	return store.Open(ctx, store.Config{Root: stateRoot(root), DatabaseURL: databaseURL, AllowHistoricalEvidenceImport: true})
}

var beginSync = func(ctx context.Context, ledger *store.Store, connector string) (*store.Sync, error) {
	return ledger.BeginSync(ctx, connector)
}

var evaluateGate = func(ctx context.Context, ledger *store.Store, definition gates.Definition) (gates.Result, error) {
	return gates.Evaluate(ctx, ledger, definition)
}

var applyProjection = func(ctx context.Context, projector *projections.Projector, ledger *store.Store) error {
	return projector.Apply(ctx, ledger)
}

func closeStore(ledger *store.Store) error {
	if ledger == nil {
		return nil
	}
	return ledger.Close()
}

func stateRoot(root string) string { return filepath.Join(root, ".proofbound") }

func productEnv(live string) string { return os.Getenv(live) }

var newIDs = func() (*core.IDGenerator, error) {
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
	resolver := &commitIntentResolver{records: recordsProvider, specdir: specdirProvider, root: root, now: now, ledger: ledger}
	connector, err := connectorgit.New(&connectorgit.Deps{Repo: repo, IDs: ids, Logger: logger(), Resolver: resolver})
	if err != nil {
		return connectorgit.Result{}, err
	}
	run, err := beginSync(ctx, ledger, "git")
	if err != nil {
		return connectorgit.Result{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	return result, errors.Join(syncErr, run.Finish(ctx, result.Cursor, syncErr))
}

// sddClaimProvider is the one sdd surface a claim needs; *intentsdd.Provider satisfies it.
type sddClaimProvider interface {
	Resolve(ctx context.Context, commit, name string) (connectorintent.Revision, error)
}

// newSDDClaimProvider builds the sdd provider for claim resolution. It is a variable so a test can
// count constructions (R2 M3 SPEC draft 5 §7 row 6, M3-INV-6).
var newSDDClaimProvider = func(root string, now time.Time) (sddClaimProvider, error) {
	repo, err := gitcmd.New(root)
	if err != nil {
		return nil, err
	}
	reader, err := intentsdd.NewGitReader(repo)
	if err != nil {
		return nil, err
	}
	return intentsdd.New(reader, now)
}

// readSDDLedgered reads the set L of (change-intent id, digest) for every ledgered intent.sdd
// change intent, keyed id + "\x00" + digest (M3 SPEC §4.3). A variable so a test can count reads.
var readSDDLedgered = func(ctx context.Context, ledger *store.Store) (map[string]bool, error) {
	set := map[string]bool{}
	err := ledger.ReadEvents(ctx, store.Filter{Source: core.SourceIntentSDD, Kind: core.KindChangeIntent}, func(record store.Record) error {
		var payload connectorintent.ChangeIntent
		if err := json.Unmarshal(record.Event.Payload, &payload); err != nil {
			return fmt.Errorf("ledgered sdd change intent seq %d: %w", record.Seq, err)
		}
		set[payload.IntentID+"\x00"+payload.ArtifactSHA256] = true
		return nil
	})
	return set, err
}

// sddDatedRE is the archive-date prefix of an archived change's directory (M3 SPEC §4.2). The id and
// digest grammars are not repeated here: connectorintent.ValidateRevision is their home (Law 2).
var sddDatedRE = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}-`)

// commitIntentResolver resolves one Intent claim at its commit. The sdd provider and the ledger set
// L are built lazily, on the first sdd claim, so a history without one does neither (M3-INV-6).
type commitIntentResolver struct {
	records *intentrecords.Provider
	specdir *intentspecdir.Provider
	root    string
	now     time.Time
	ledger  *store.Store

	sdd      sddClaimProvider
	ledgered map[string]bool
}

func (r *commitIntentResolver) Resolve(ctx context.Context, commitSHA, provider, recordID string) (connectorgit.IntentRef, error) {
	var revision connectorintent.Revision
	var err error
	switch provider {
	case "records":
		revision, err = r.records.Resolve(ctx, commitSHA, recordID)
	case "specdir":
		revision, err = r.specdir.Resolve(ctx, commitSHA, recordID)
	case "sdd":
		return r.resolveSDD(ctx, commitSHA, recordID)
	default:
		return connectorgit.IntentRef{}, fmt.Errorf("unknown intent provider %q", provider)
	}
	if err != nil {
		return connectorgit.IntentRef{}, err
	}
	return connectorgit.IntentRef{Provider: provider, RecordID: revision.NativeID, ArtifactSHA256: revision.ArtifactSHA256}, nil
}

// resolveSDD maps an sdd claim to a ref or a *ClaimUnboundError (M3 SPEC §4.2-§4.4). Tenant outcomes
// are unbound claims; a reader failure or a revision failing the anti-spoofing checks is an error.
func (r *commitIntentResolver) resolveSDD(ctx context.Context, commitSHA, name string) (connectorgit.IntentRef, error) {
	if r.sdd == nil {
		provider, err := newSDDClaimProvider(r.root, r.now)
		if err != nil {
			return connectorgit.IntentRef{}, err
		}
		r.sdd = provider
	}
	revision, err := r.sdd.Resolve(ctx, commitSHA, name)
	var unmappable *intentsdd.UnmappableError
	switch {
	case errors.Is(err, intentsdd.ErrUnknownChange):
		return connectorgit.IntentRef{}, &connectorgit.ClaimUnboundError{Cause: "unknown-change"}
	case errors.Is(err, intentsdd.ErrArchivedEarlier):
		return connectorgit.IntentRef{}, &connectorgit.ClaimUnboundError{Cause: "archived-earlier"}
	case errors.Is(err, intentsdd.ErrNoChangeIntent):
		return connectorgit.IntentRef{}, &connectorgit.ClaimUnboundError{Cause: "no-change-intent"}
	case errors.As(err, &unmappable):
		return connectorgit.IntentRef{}, &connectorgit.ClaimUnboundError{Cause: "unmappable:" + unmappable.Cause}
	case err != nil:
		return connectorgit.IntentRef{}, err
	}
	segment := path.Base(revision.ArtifactPath)
	if revision.Source != core.SourceIntentSDD || revision.Kind != core.KindChangeIntent ||
		connectorintent.ValidateRevision(revision) != nil || !segmentNamesChange(segment, name) {
		return connectorgit.IntentRef{}, fmt.Errorf("sdd resolver returned a revision that does not name change %q", name)
	}
	if r.ledgered == nil {
		set, err := readSDDLedgered(ctx, r.ledger)
		if err != nil {
			return connectorgit.IntentRef{}, err
		}
		r.ledgered = set
	}
	if !r.ledgered[revision.NativeID+"\x00"+revision.ArtifactSHA256] {
		return connectorgit.IntentRef{}, &connectorgit.ClaimUnboundError{Cause: "not-ledgered"}
	}
	return connectorgit.IntentRef{Provider: "sdd", RecordID: revision.NativeID, ArtifactSHA256: revision.ArtifactSHA256}, nil
}

// segmentNamesChange reports whether an artifact path's last segment names the claimed change: the
// active `<name>`, or the archived `YYYY-MM-DD-<name>` (M3 SPEC §4.2; defence in depth only).
func segmentNamesChange(segment, name string) bool {
	if segment == name {
		return true
	}
	return sddDatedRE.MatchString(segment) && segment[len("YYYY-MM-DD-"):] == name
}

// printClaimSkips prints one line per sdd claim that did not bind, after a counts line (M3 §4.4).
func printClaimSkips(output io.Writer, skips []connectorgit.ClaimSkip) error {
	for _, skip := range skips {
		if _, err := fmt.Fprintf(output, "claim-skip %s %s at %s\n", skip.Cause, printableUnit(skip.Value), skip.Commit); err != nil {
			return err
		}
	}
	return nil
}

type checksResult struct{ Listed, Appended, Existing int }

func syncChecksOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (checksResult, error) {
	connector, err := checks.New(&checks.Deps{SpoolDir: filepath.Join(stateRoot(root), "spool"), IDs: ids, Logger: logger()})
	if err != nil {
		return checksResult{}, err
	}
	run, err := beginSync(ctx, ledger, "checks")
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
		if !strings.HasSuffix(name, ".md") {
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

type reviewsResult struct {
	Listed, Appended, Existing, Malformed, Documentary int
	Held, Refused                                      []connectorreviews.Hold
}

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
	providers, err := intentProviders(root, selection, time.Now().UTC())
	if err != nil {
		return connectorintent.Result{}, err
	}
	connector, err := connectorintent.New(&connectorintent.Deps{Providers: providers, IDs: ids, Logger: logger()})
	if err != nil {
		return connectorintent.Result{}, err
	}
	run, err := beginSync(ctx, ledger, "intent."+selection)
	if err != nil {
		return connectorintent.Result{}, err
	}
	result, syncErr := connector.Sync(ctx, selection, run)
	return result, errors.Join(syncErr, run.Finish(ctx, result.Cursor, syncErr))
}

// intentProviders constructs the intent providers a selection needs. intent.sdd is constructed only
// when selected by name: `all` keeps meaning records + specdir, so `sync all`, `verify`, and
// delivery-enforce never walk full history (sdd SPEC §12, SDD-INV-56).
func intentProviders(root, selection string, now time.Time) ([]connectorintent.Provider, error) {
	if selection == "sdd" {
		repo, err := gitcmd.New(root)
		if err != nil {
			return nil, err
		}
		reader, err := intentsdd.NewGitReader(repo)
		if err != nil {
			return nil, err
		}
		provider, err := intentsdd.New(reader, now)
		if err != nil {
			return nil, err
		}
		return []connectorintent.Provider{provider}, nil
	}
	reader := committedIntentReader{root: root}
	recordsProvider, err := intentrecords.New(reader, now)
	if err != nil {
		return nil, err
	}
	if specKitOwnsSpecs(root) {
		// The tenant's committed profile declares specs/ as Spec Kit's, so the P5 specdir
		// convention (also rooted at specs/) does not apply and would reject every Spec Kit file
		// (CLI-INV-22).
		return []connectorintent.Provider{recordsProvider}, nil
	}
	specdirProvider, err := intentspecdir.New(reader, now)
	if err != nil {
		return nil, err
	}
	return []connectorintent.Provider{recordsProvider, specdirProvider}, nil
}

// hasSDDProfile reports whether HEAD commits the sdd profile. It reads one committed path and
// never constructs the sdd provider.
func hasSDDProfile(root string) bool {
	return gitOutput(context.Background(), root, "rev-parse", "--verify", "--quiet", "HEAD:"+intentsdd.ProfilePath) != ""
}

// specKitOwnsSpecs reports whether HEAD's committed sdd profile declares kit "speckit" with a root
// at or overlapping specs/. It only reads one committed file; it never constructs the sdd provider
// (SDD-INV-56). A missing or unreadable profile leaves the P5 providers exactly as before.
func specKitOwnsSpecs(root string) bool {
	raw := gitOutput(context.Background(), root, "show", "HEAD:"+intentsdd.ProfilePath)
	if raw == "" {
		return false
	}
	var profile struct {
		Kit     string `json:"kit"`
		Changes struct {
			Root string `json:"root"`
		} `json:"changes"`
	}
	if err := json.Unmarshal([]byte(raw), &profile); err != nil || profile.Kit != intentsdd.KitSpecKit {
		return false
	}
	r := strings.TrimSuffix(profile.Changes.Root, "/")
	return r == "specs" || strings.HasPrefix(r, "specs/")
}

// printableUnit keeps a tenant-controlled unit name on one line: a name with a C0 or C1 control,
// DEL, U+2028/U+2029, or invalid UTF-8 is printed Go-quoted (strconv.Quote escapes every one of
// them), so it can neither split nor forge a skip line (sdd re-check
// r2-m1-recheck-round1-20261118 F3).
func printableUnit(unit string) string {
	if !utf8.ValidString(unit) {
		return strconv.Quote(unit)
	}
	for _, r := range unit {
		// C0, DEL and C1 (U+0080-U+009F, incl. NEL and CSI), and the Unicode line and paragraph
		// separators, which some terminals treat as line breaks or escape introducers.
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) || r == 0x2028 || r == 0x2029 {
			return strconv.Quote(unit)
		}
	}
	return unit
}

// writeSkips prints one line per Skip after the counts line (sdd SPEC §7, SDD-INV-60).
func writeSkips(output io.Writer, result connectorintent.Result) error {
	names := make([]string, 0, len(result.ByProvider))
	for name := range result.ByProvider {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, skip := range result.ByProvider[name].Skipped {
			if _, err := fmt.Fprintf(output, "skip %s %s at %s\n", skip.Cause, printableUnit(skip.Unit), skip.Commit); err != nil {
				return err
			}
		}
	}
	return nil
}

func syncReviewsOnStore(ctx context.Context, root string, ledger *store.Store, ids *core.IDGenerator) (reviewsResult, error) {
	connector, err := connectorreviews.New(&connectorreviews.Deps{Reader: committedVerdictReader{root: root}, IDs: ids, Logger: logger(), Ledger: &storeLedgerView{ledger: ledger}})
	if err != nil {
		return reviewsResult{}, err
	}
	run, err := beginSync(ctx, ledger, "reviews")
	if err != nil {
		return reviewsResult{}, err
	}
	result, syncErr := connector.Sync(ctx, run)
	finishErr := run.Finish(ctx, result.Cursor, syncErr)
	return reviewsResult{result.Listed, result.Appended, result.Existing, result.Malformed, result.Documentary, result.Held, result.Refused}, errors.Join(syncErr, finishErr)
}

// ledgerViewSets are what one pass over the ledger tells the reviews hold (M3 SPEC draft 8 §4.6).
type ledgerViewSets struct {
	identities   map[string]bool // payload schema \x00 native id, of every reviews event
	claims       map[string]bool // sha \x00 provider \x00 record id \x00 digest, of every git ref
	requirements map[string]bool // source \x00 requirement id \x00 digest, of every intent requirement
}

// readLedgerView is one ReadEvents pass with an empty filter. It is a variable so a test can count
// passes (row 12). An undecodable git payload, or a reviews payload without a schema, fails it: the
// ledger is then unprojectable anyway, and failing is fail-closed.
var readLedgerView = func(ctx context.Context, ledger *store.Store) (*ledgerViewSets, error) {
	sets := &ledgerViewSets{identities: map[string]bool{}, claims: map[string]bool{}, requirements: map[string]bool{}}
	err := ledger.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		e := record.Event
		switch {
		case e.Source == core.SourceReviews:
			var probe struct {
				Schema string `json:"schema"`
			}
			if err := json.Unmarshal(e.Payload, &probe); err != nil || probe.Schema == "" {
				return fmt.Errorf("ledgered reviews event seq %d has no schema", record.Seq)
			}
			sets.identities[probe.Schema+"\x00"+e.NativeID] = true
		case e.Source == core.SourceGit && e.Kind == core.KindCommitRecorded:
			var commit connectorgit.Commit
			if err := json.Unmarshal(e.Payload, &commit); err != nil {
				return fmt.Errorf("ledgered git event seq %d: %w", record.Seq, err)
			}
			for _, ref := range commit.IntentRefs {
				sets.claims[commit.SHA+"\x00"+ref.Provider+"\x00"+ref.RecordID+"\x00"+ref.ArtifactSHA256] = true
			}
		case e.Kind == core.KindRequirement && strings.HasPrefix(string(e.Source), "intent."):
			var requirement connectorintent.Requirement
			if err := json.Unmarshal(e.Payload, &requirement); err != nil {
				return fmt.Errorf("ledgered requirement seq %d: %w", record.Seq, err)
			}
			sets.requirements[string(e.Source)+"\x00"+requirement.RequirementID+"\x00"+requirement.ArtifactSHA256] = true
		}
		return nil
	})
	return sets, err
}

// storeLedgerView answers the reviews hold from the store, reading the ledger lazily (at most once
// per sync, and never for a sync whose artifacts are all v1 or documentary).
type storeLedgerView struct {
	ledger *store.Store
	sets   *ledgerViewSets
}

func (v *storeLedgerView) load(ctx context.Context) (*ledgerViewSets, error) {
	if v.sets == nil {
		sets, err := readLedgerView(ctx, v.ledger)
		if err != nil {
			return nil, err
		}
		v.sets = sets
	}
	return v.sets, nil
}

func (v *storeLedgerView) Ledgered(ctx context.Context, schema, nativeID string) (bool, error) {
	sets, err := v.load(ctx)
	if err != nil {
		return false, err
	}
	return sets.identities[schema+"\x00"+nativeID], nil
}

func (v *storeLedgerView) CommitClaimed(ctx context.Context, sha, provider, recordID, digest string) (bool, error) {
	sets, err := v.load(ctx)
	if err != nil {
		return false, err
	}
	return sets.claims[sha+"\x00"+provider+"\x00"+recordID+"\x00"+digest], nil
}

func (v *storeLedgerView) RequirementLedgered(ctx context.Context, source core.Source, recordID, digest string) (bool, error) {
	sets, err := v.load(ctx)
	if err != nil {
		return false, err
	}
	return sets.requirements[string(source)+"\x00"+recordID+"\x00"+digest], nil
}

// printHolds prints the reviews hold's held and refused lines, after a counts line (M3 §4.6).
func printHolds(output io.Writer, result reviewsResult) error {
	for _, group := range []struct {
		word  string
		holds []connectorreviews.Hold
	}{{"held", result.Held}, {"refused", result.Refused}} {
		for _, h := range group.holds {
			if _, err := fmt.Fprintf(output, "%s %s %s\n", group.word, h.Cause, printableUnit(h.Path)); err != nil {
				return err
			}
		}
	}
	return nil
}

type githubResult struct{ Listed, Appended, Existing int }

func syncGitHubOnStore(ctx context.Context, ledger *store.Store, ids *core.IDGenerator) (githubResult, error) {
	owner := strings.TrimSpace(productEnv("PROOFBOUND_GITHUB_OWNER"))
	var repos []string
	for _, repo := range strings.Split(productEnv("PROOFBOUND_GITHUB_REPOS"), ",") {
		if repo = strings.TrimSpace(repo); repo != "" {
			repos = append(repos, repo)
		}
	}
	connector, err := connectorgithub.New(&connectorgithub.Deps{
		API:   &connectorgithub.HTTPClient{BaseURL: productEnv("PROOFBOUND_GITHUB_API_BASE_URL"), Token: os.Getenv("GITHUB_TOKEN")},
		Owner: owner, Repos: repos, IDs: ids, Logger: logger(),
	})
	if err != nil {
		return githubResult{}, err
	}
	run, err := beginSync(ctx, ledger, "github")
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
	run, err := beginSync(ctx, ledger, "sessions")
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
	switch cmd {
	case commandInitOpenSpec:
		owner := ""
		if len(args) == 4 {
			owner = args[3]
		}
		return initKit(ctx, root, args[1], owner, output)
	case commandReviewSeal:
		// The tests run before the record is opened (ui-reviewer-spec §1.1, review F2).
		var ran *sealTestsRun
		if _, edited, err := readSealDraft(args[2]); err == nil && sha40RE.MatchString(edited.Commit) {
			res, err := draftTests(ctx, root, edited.Commit, output)
			if err != nil {
				return err
			}
			ran = &sealTestsRun{Commit: edited.Commit, Result: res}
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		fresh := func(ctx context.Context, sha string) (draftFile, error) { return ledgerDraft(ctx, ledger, sha) }
		return reviewSealWith(ctx, root, args[2], ran, fresh, output)
	case commandKeysInit:
		return keysInit(ctx, root, flagValue(args[2:], "--name"), output)
	case commandUI:
		return runUI(ctx, root, databaseURL, args, output)
	case commandKeysGitSign:
		return keysGitSign(ctx, root, output)
	case commandWaiverSign:
		id, paths, reason, expires, _ := parseWaiverSign(args)
		return waiverSign(ctx, root, id, paths, reason, expires, time.Now(), output)
	case commandTask:
		sub, name, _ := parseTask(args)
		return runTask(root, sub, name, output)
	case commandWarrantyExport:
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		return warrantyExport(ctx, root, ledger, args[2], flagValue(args[3:], "-o"), output)
	case commandWarrantyCosign:
		return warrantyCosign(ctx, root, args[2], flagValue(args[3:], "-o"), output)
	case commandWarrantyBundle:
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		return warrantyBundle(ctx, root, ledger, args[2], flagValue(args[3:], "-o"), output)
	case commandCheck:
		base, markdown, requireClaims, _ := parseCheck(args)
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		return checkRange(ctx, root, ledger, base, markdown, requireClaims, output)
	case commandReportAgents:
		// The agent track record (CLI-INV-74): computed from the same audit as the Audit room.
		base, err := resolveCommit(ctx, root, args[2])
		if err != nil {
			return err
		}
		// Sync first, as the Audit room does, so reviews and claims committed since the last sync count
		// (review M2); a failed sync is reported, not fatal.
		var syncErr bytes.Buffer
		if code := run(ctx, []string{"sync", "all"}, io.Discard, &syncErr); code != 0 {
			if _, err := fmt.Fprintf(output, "note: sync failed, so the record may be behind git: %s\n", printableUnit(strings.TrimSpace(syncErr.String()))); err != nil {
				return err
			}
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		rep, err := computeAudit(ctx, root, ledger, base)
		if err != nil {
			return err
		}
		// The same exceptions as the Audit room: a verified signed outcome closes one (review M1).
		attachOutcomes(ctx, root, &rep)
		tr := computeTrack(ctx, root, rep)
		if len(args) == 4 {
			return writeTrackJSON(output, tr)
		}
		return writeTrack(output, tr)
	case commandReportDrift:
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		return reportDrift(ctx, root, ledger, output)
	case commandReviewDraft:
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		return reviewDraft(ctx, root, ledger, args[2], output)
	case commandSyncChecks:
		return syncChecks(ctx, root, databaseURL, output)
	case commandSyncSessions:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		result, err := syncSessionsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d skipped=%d\n", result.Listed, result.Appended, result.Existing, result.Skipped)
		return err
	case commandSyncReviews:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		result, syncErr := syncReviewsOnStore(ctx, root, ledger, ids)
		// Counts, held and refused lines are printed before ANY error is returned (M3 draft 8 §4.6).
		if _, err := fmt.Fprintf(output, "listed=%d appended=%d existing=%d malformed=%d documentary=%d\n", result.Listed, result.Appended, result.Existing, result.Malformed, result.Documentary); err != nil {
			return errors.Join(syncErr, err)
		}
		return errors.Join(syncErr, printHolds(output, result))
	case commandSyncGitHub:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		result, err := syncGitHubOnStore(ctx, ledger, ids)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing)
		return err
	case commandSyncIntentRecords, commandSyncIntentSpecdir, commandSyncIntentSDD, commandSyncIntentAll:
		selection := map[command]string{commandSyncIntentRecords: "records", commandSyncIntentSpecdir: "specdir", commandSyncIntentSDD: "sdd", commandSyncIntentAll: "all"}[cmd]
		ids, err := newIDs()
		if err != nil {
			return err
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		result, err := syncIntentOnStore(ctx, root, selection, ledger, ids)
		if err != nil {
			return err
		}
		if _, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", result.Listed, result.Appended, result.Existing); err != nil {
			return err
		}
		return writeSkips(output, result)
	case commandMigrateHistoricalEvidence:
		ledger, err := openHistoricalEvidenceStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
		records, err := migration.Load(filepath.Join(root, migration.ArchivePath))
		if err != nil {
			return err
		}
		if err := migration.Import(ctx, ledger, records); err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "imported=%d\n", len(records))
		return err
	}
	ledger, err := openStore(ctx, root, databaseURL)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, closeStore(ledger)) }()
	projector := projections.New()
	switch cmd {
	case commandGatesCanary, commandGatesEnforce:
		definitions, err := gates.LoadDir(filepath.Join(root, "gates"))
		if err != nil {
			return err
		}
		definitions, err = gateDefinitionsForCommand(cmd, definitions)
		if err != nil {
			return fmt.Errorf("select gate definitions: %w", err)
		}
		if err := resolveIntentGateHeadScope(ctx, root, definitions); err != nil {
			return err
		}
		results := make([]gates.Result, 0, len(definitions))
		for _, definition := range definitions {
			result, err := evaluateGate(ctx, ledger, definition)
			if err != nil {
				return err
			}
			reason := ""
			if result.Reason != "" {
				reason = " reason=" + strconv.Quote(result.Reason)
			}
			if _, err := fmt.Fprintf(output, "gate=%s state=%s seq=%d proof=%s would_block=%t%s\n", result.GateID, result.State, result.Seq, result.EventID, result.WouldBlock, reason); err != nil {
				return err
			}
			results = append(results, result)
		}
		switch cmd {
		case commandGatesEnforce:
			if err := enforceGateResults(definitions, results); err != nil {
				return err
			}
		}
		return nil
	}
	switch cmd {
	case commandRebuild:
		return projector.Rebuild(ctx, ledger)
	case commandTwinReplay:
		return twinReplay(ctx, ledger, args, output)
	case commandSyncGit, commandSyncAll:
		ids, err := newIDs()
		if err != nil {
			return err
		}
		var intentResult connectorintent.Result
		var sddErr error
		if cmd == commandSyncAll {
			intentResult, err = syncIntentOnStore(ctx, root, "all", ledger, ids)
			if err != nil {
				return err
			}
			// A project that committed an sdd profile opted in to spec sync, so `sync all` reads
			// its specs too (CLI-INV-23); without one, nothing sdd is constructed (SDD-INV-56).
			// A failing spec sync never stops the other steps: they run, and `sync all` then fails
			// with the spec sync's error joined to theirs (CLI-INV-23).
			if hasSDDProfile(root) {
				sddResult, err := syncIntentOnStore(ctx, root, "sdd", ledger, ids)
				if err != nil {
					sddErr = fmt.Errorf("sync intent sdd: %w", err)
					if _, err := fmt.Fprintf(output, "sdd error=%q\n", sddErr.Error()); err != nil {
						return errors.Join(sddErr, err)
					}
				} else {
					if _, err = fmt.Fprintf(output, "sdd listed=%d appended=%d existing=%d\n", sddResult.Listed, sddResult.Appended, sddResult.Existing); err != nil {
						return err
					}
					if err := writeSkips(output, sddResult); err != nil {
						return err
					}
					intentResult.Appended += sddResult.Appended
				}
			}
		}
		gitResult, err := syncGit(ctx, root, ledger, ids)
		if err != nil {
			return errors.Join(sddErr, err)
		}
		switch cmd {
		case commandSyncGit:
			if _, err = fmt.Fprintf(output, "listed=%d appended=%d existing=%d\n", gitResult.Listed, gitResult.Appended, gitResult.Existing); err != nil {
				return err
			}
			return printClaimSkips(output, gitResult.ClaimSkips)
		case commandSyncAll:
			break
		}
		checksResult, err := syncChecksOnStore(ctx, root, ledger, ids)
		if err != nil {
			return errors.Join(sddErr, err)
		}
		sessionsResult, err := syncSessionsOnStore(ctx, root, ledger, ids)
		if err != nil {
			return errors.Join(sddErr, err)
		}
		reviewsResult, syncErr := syncReviewsOnStore(ctx, root, ledger, ids)
		syncErr = errors.Join(sddErr, syncErr)
		// Reviews is the last step; its counts, claim-skip, held and refused lines are printed before
		// its error is returned (M3 draft 8 §4.6).
		if _, err = fmt.Fprintf(output, "intent appended=%d git appended=%d checks appended=%d sessions appended=%d reviews appended=%d\n", intentResult.Appended, gitResult.Appended, checksResult.Appended, sessionsResult.Appended, reviewsResult.Appended); err != nil {
			return errors.Join(syncErr, err)
		}
		if err := printClaimSkips(output, gitResult.ClaimSkips); err != nil {
			return errors.Join(syncErr, err)
		}
		return errors.Join(syncErr, printHolds(output, reviewsResult))
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
		if err := applyProjection(ctx, projector, ledger); err != nil {
			return err
		}
		return projector.ReportWeek(ctx, ledger, time.Now(), reachable, output)
	case commandReportGitHub:
		return projector.ReportGitHub(ctx, ledger, time.Now(), output)
	case commandReportIntent:
		id, err := reportIntentID(ctx, root, args[2])
		if err != nil {
			return err
		}
		return projector.ReportIntent(ctx, ledger, id, time.Now(), output)
	case commandReportRequirement:
		return projector.ReportRequirement(ctx, ledger, args[2], output)
	case commandIntentCheck:
		return projector.CheckIntent(ctx, ledger, args[3], output)
	}
	return nil
}

func enforceDefinitions(definitions []gates.Definition) []gates.Definition {
	promoted := make([]gates.Definition, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Mode == "enforce" {
			promoted = append(promoted, definition)
		}
	}
	return promoted
}

func gateDefinitionsForCommand(cmd command, definitions []gates.Definition) ([]gates.Definition, error) {
	if cmd != commandGatesEnforce {
		return definitions, nil
	}
	promoted := enforceDefinitions(definitions)
	if err := gates.RequireDefinitions(promoted); err != nil {
		return nil, err
	}
	for _, definition := range promoted {
		if err := definition.EnforceReady(); err != nil {
			return nil, err
		}
	}
	return promoted, nil
}

// reportIntentID maps `report intent sdd:<change-name>` to the change intent that name resolves
// to at HEAD, through the same sdd provider commit claims bind through (CLI-INV-17). Any other
// argument is already an intent id.
func reportIntentID(ctx context.Context, root, arg string) (string, error) {
	name, ok := strings.CutPrefix(arg, "sdd:")
	if !ok {
		return arg, nil
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--verify", "HEAD^{commit}")
	cmd.Env = repositoryGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("report intent %s: resolve HEAD: %w", arg, err)
	}
	provider, err := newSDDClaimProvider(root, time.Now())
	if err != nil {
		return "", fmt.Errorf("report intent %s: %w", arg, err)
	}
	revision, err := provider.Resolve(ctx, strings.TrimSpace(string(out)), name)
	if err != nil {
		return "", fmt.Errorf("report intent %s: no change intent at HEAD: %w", arg, err)
	}
	return revision.NativeID, nil
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

// twinReplay runs twin.ReplayIsolated against the committed projection in a
// throwaway embedded store and prints the verdict and vera.replay.v1 proof. It
// never mutates the source ledger: the isolated path is read-only over the
// committed events, and the temporary store is created and removed internally.
// With no --candidate it re-derives and proves the committed state (verdict
// preserved); with --candidate it evaluates one counterfactual event.
func twinReplay(ctx context.Context, ledger *store.Store, args []string, output io.Writer) error {
	through, candidatePath, err := parseTwinReplayFlags(args)
	if err != nil {
		return err
	}
	if through < 0 {
		if through, err = latestLedgerSeq(ctx, ledger); err != nil {
			return err
		}
	}
	// The baseline is the isolated projection of the committed prefix THROUGH the
	// same seq, so the verdict reflects only the candidate's effect — never the
	// events dropped by a lower --through. A no-candidate replay then proves the
	// prefix re-derives deterministically.
	base, err := twin.ReplayIsolated(ctx, ledger, twin.Request{ThroughSeq: through}, projections.Snapshot{})
	if err != nil {
		return fmt.Errorf("twin replay: baseline projection: %w", err)
	}
	baseline := base.Candidate
	var candidates []twin.Candidate
	if candidatePath != "" {
		candidate, err := loadTwinCandidate(candidatePath)
		if err != nil {
			return err
		}
		candidates = []twin.Candidate{candidate}
	}
	result, err := twin.ReplayIsolated(ctx, ledger, twin.Request{ThroughSeq: through, Candidate: candidates}, baseline)
	if err != nil {
		return fmt.Errorf("twin replay: %w", err)
	}
	_, err = fmt.Fprintf(output,
		"verdict=%s through_seq=%d source_events=%d candidate_events=%d\nproof schema=%s source_digest=%s candidate_digest=%s baseline_snapshot=%s candidate_snapshot=%s\n",
		result.Verdict, result.ThroughSeq, result.SourceEvents, result.CandidateEvents,
		result.Proof.Schema, result.Proof.SourceDigest, result.Proof.CandidateDigest,
		result.Proof.BaselineSnapshotDigest, result.Proof.CandidateSnapshotDigest)
	return err
}

// parseTwinReplayFlags parses the optional --through and --candidate flags that
// follow `twin replay`. through is -1 when unset (meaning: the latest seq).
func parseTwinReplayFlags(args []string) (through int64, candidatePath string, err error) {
	through = -1
	rest := args[2:]
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--through":
			i++
			if i >= len(rest) {
				return 0, "", errors.New("twin replay: --through requires a value")
			}
			through, err = strconv.ParseInt(rest[i], 10, 64)
			if err != nil || through < 0 {
				return 0, "", fmt.Errorf("twin replay: invalid --through %q", rest[i])
			}
		case "--candidate":
			i++
			if i >= len(rest) {
				return 0, "", errors.New("twin replay: --candidate requires a value")
			}
			candidatePath = rest[i]
		default:
			return 0, "", fmt.Errorf("twin replay: unexpected argument %q", rest[i])
		}
	}
	return through, candidatePath, nil
}

func latestLedgerSeq(ctx context.Context, ledger *store.Store) (int64, error) {
	var maxSeq int64
	if err := ledger.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		if record.Seq > maxSeq {
			maxSeq = record.Seq
		}
		return nil
	}); err != nil {
		return 0, fmt.Errorf("twin replay: read ledger: %w", err)
	}
	return maxSeq, nil
}

// loadTwinCandidate reads a counterfactual event from a JSON file shaped
// {"seq": <int>, "event": {<core.Event canonical JSON>}}. Unknown fields are
// rejected; twin.ReplayIsolated validates the event itself.
func loadTwinCandidate(path string) (twin.Candidate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return twin.Candidate{}, fmt.Errorf("twin replay: read candidate: %w", err)
	}
	var parsed struct {
		Seq   int64      `json:"seq"`
		Event core.Event `json:"event"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return twin.Candidate{}, fmt.Errorf("twin replay: parse candidate %s: %w", path, err)
	}
	if decoder.More() {
		return twin.Candidate{}, fmt.Errorf("twin replay: candidate %s has trailing data after the JSON object", path)
	}
	return twin.Candidate{Seq: parsed.Seq, Event: parsed.Event}, nil
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
	if err := requireNoSecondSyncAppends(secondIntent.Appended, secondGit.Appended, secondChecks.Appended, secondSessions.Appended, secondReviews.Appended); err != nil {
		return err
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

func requireNoSecondSyncAppends(intent, git, checks, sessions, reviews int) error {
	counts := []int{intent, git, checks, sessions, reviews}
	for _, count := range counts {
		if count != 0 {
			return fmt.Errorf("verify: second sync appended intent=%d git=%d checks=%d sessions=%d reviews=%d", intent, git, checks, sessions, reviews)
		}
	}
	return nil
}

func verifyStep(ctx context.Context, name string, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("verify: %s: %w", name, err)
	}
	if productEnv("PROOFBOUND_VERIFY_TRACE") == "1" {
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
	if !validMakeCheckWitness(filename, witness) {
		return checks.Witness{}, fmt.Errorf("verify: latest spool file %s is not a valid make check witness", filename)
	}
	return witness, nil
}

func validMakeCheckWitness(filename string, witness checks.Witness) bool {
	return filename == witness.RunID+".json" && witness.Schema == "vera.witness.v1" && witness.Command == "make check" && witness.RunID != ""
}

type eventReader interface {
	ReadEvents(context.Context, store.Filter, func(store.Record) error) error
}

func latestLedgerCheckRunID(ctx context.Context, ledger eventReader) (string, error) {
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
