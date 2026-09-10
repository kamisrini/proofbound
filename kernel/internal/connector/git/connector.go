package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "git/2"

var intentIDRE = regexp.MustCompile(`^CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`)

type Repo interface {
	Commits(context.Context) ([]Commit, error)
	Tips(context.Context) (map[string]string, error)
}

type Commit struct {
	SHA            string      `json:"sha"`
	AuthorName     string      `json:"author_name"`
	AuthorEmail    string      `json:"author_email"`
	CommitterName  string      `json:"committer_name"`
	CommitterEmail string      `json:"committer_email"`
	CommittedAt    time.Time   `json:"committed_at"`
	Subject        string      `json:"subject"`
	FilesTouched   []string    `json:"files_touched"`
	CitedDecisions []string    `json:"cited_decisions"`
	IntentRefs     []IntentRef `json:"intent_refs"`
	IntentTrailers []string    `json:"-"`
}

type IntentRef struct {
	Provider       string `json:"provider"`
	RecordID       string `json:"record_id"`
	ArtifactSHA256 string `json:"artifact_sha256"`
}

type IntentResolver interface {
	Resolve(context.Context, string, string, string) (IntentRef, error)
}

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
	Repo     Repo
	IDs      *core.IDGenerator
	Logger   *slog.Logger
	Resolver IntentResolver
}

type Connector struct {
	repo     Repo
	ids      *core.IDGenerator
	logger   *slog.Logger
	resolver IntentResolver
}

type Result struct {
	Listed   int
	Appended int
	Existing int
	Cursor   json.RawMessage
}

func New(d *Deps) (*Connector, error) {
	if d == nil {
		return nil, errors.New("git connector: dependencies are required")
	}
	if isNilRepo(d.Repo) {
		return nil, errors.New("git connector: Repo is required")
	}
	if d.IDs == nil {
		return nil, errors.New("git connector: IDs is required")
	}
	if d.Logger == nil {
		return nil, errors.New("git connector: Logger is required")
	}
	return &Connector{repo: d.Repo, ids: d.IDs, logger: d.Logger, resolver: d.Resolver}, nil
}

func isNilRepo(repo Repo) bool {
	if repo == nil {
		return true
	}
	value := reflect.ValueOf(repo)
	if value.Kind() == reflect.Pointer {
		return value.IsNil()
	}
	return false
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var result Result
	if c == nil || c.repo == nil || c.ids == nil {
		return result, errors.New("git connector: connector is not initialized")
	}
	if appender == nil {
		return result, errors.New("git connector: appender is required")
	}
	commits, err := c.repo.Commits(ctx)
	if err != nil {
		return result, fmt.Errorf("git connector: list commits: %w", err)
	}
	result.Listed = len(commits)
	tips, err := c.repo.Tips(ctx)
	if err != nil {
		return result, fmt.Errorf("git connector: list tips: %w", err)
	}
	cursor, err := json.Marshal(tips)
	if err != nil {
		return result, fmt.Errorf("git connector: marshal cursor: %w", err)
	}
	result.Cursor = cursor

	for _, commit := range commits {
		commit, err = c.resolvedCommit(ctx, commit)
		if err != nil {
			return result, fmt.Errorf("git connector: commit %q intent: %w", commit.SHA, err)
		}
		payload, marshalErr := json.Marshal(commit)
		if marshalErr != nil {
			return result, fmt.Errorf("git connector: marshal commit %q: %w", commit.SHA, marshalErr)
		}
		event, eventErr := c.ids.NewEvent(core.NewEventParams{
			Source:           core.SourceGit,
			NativeID:         commit.SHA,
			Kind:             core.KindCommitRecorded,
			OccurredAt:       commit.CommittedAt,
			Payload:          payload,
			ConnectorVersion: Version,
		})
		if eventErr != nil {
			return result, fmt.Errorf("git connector: commit %q: %w", commit.SHA, eventErr)
		}
		_, inserted, appendErr := appender.Append(ctx, event)
		if appendErr != nil {
			return result, fmt.Errorf("git connector: append commit %q: %w", commit.SHA, appendErr)
		}
		if inserted {
			result.Appended++
		} else {
			result.Existing++
		}
	}
	return result, nil
}

func normalizedCommit(commit Commit) Commit {
	commit.CommittedAt = commit.CommittedAt.UTC()
	commit.FilesTouched = sortedUnique(commit.FilesTouched)
	commit.CitedDecisions = sortedUnique(commit.CitedDecisions)
	commit.IntentTrailers = sortedUnique(commit.IntentTrailers)
	if len(commit.IntentRefs) == 0 {
		commit.IntentRefs = nil
	} else {
		sort.Slice(commit.IntentRefs, func(i, j int) bool { return intentRefKey(commit.IntentRefs[i]) < intentRefKey(commit.IntentRefs[j]) })
		unique := commit.IntentRefs[:0]
		for _, ref := range commit.IntentRefs {
			if len(unique) == 0 || intentRefKey(unique[len(unique)-1]) != intentRefKey(ref) {
				unique = append(unique, ref)
			}
		}
		commit.IntentRefs = unique
	}
	return commit
}

func (c *Connector) resolvedCommit(ctx context.Context, commit Commit) (Commit, error) {
	commit = normalizedCommit(commit)
	if len(commit.IntentTrailers) == 0 {
		return commit, nil
	}
	if c.resolver == nil {
		return Commit{}, errors.New("resolver is required for an Intent trailer")
	}
	commit.IntentRefs = nil
	for _, raw := range commit.IntentTrailers {
		provider, id := "records", raw
		if strings.Contains(raw, ":") {
			parts := strings.Split(raw, ":")
			if len(parts) != 2 {
				return Commit{}, fmt.Errorf("malformed trailer %q", raw)
			}
			provider, id = parts[0], parts[1]
		}
		if provider != "records" && provider != "specdir" {
			return Commit{}, fmt.Errorf("unknown provider %q", provider)
		}
		if !intentIDRE.MatchString(id) {
			return Commit{}, fmt.Errorf("malformed change intent %q", id)
		}
		ref, err := c.resolver.Resolve(ctx, commit.SHA, provider, id)
		if err != nil {
			return Commit{}, err
		}
		if ref.Provider != provider || ref.RecordID != id || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(ref.ArtifactSHA256) {
			return Commit{}, errors.New("resolver returned a spoofed or malformed exact revision")
		}
		commit.IntentRefs = append(commit.IntentRefs, ref)
	}
	return normalizedCommit(commit), nil
}

func intentRefKey(r IntentRef) string {
	return r.Provider + "\x00" + r.RecordID + "\x00" + r.ArtifactSHA256
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
