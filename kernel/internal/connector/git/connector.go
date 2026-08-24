package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "git/1"

type Repo interface {
	Commits(context.Context) ([]Commit, error)
	Tips(context.Context) (map[string]string, error)
}

type Commit struct {
	SHA            string    `json:"sha"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CommittedAt    time.Time `json:"committed_at"`
	Subject        string    `json:"subject"`
	FilesTouched   []string  `json:"files_touched"`
	CitedDecisions []string  `json:"cited_decisions"`
}

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
	Repo   Repo
	IDs    *core.IDGenerator
	Logger *slog.Logger
}

type Connector struct {
	repo   Repo
	ids    *core.IDGenerator
	logger *slog.Logger
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
	if d.Repo == nil {
		return nil, errors.New("git connector: Repo is required")
	}
	if d.IDs == nil {
		return nil, errors.New("git connector: IDs is required")
	}
	if d.Logger == nil {
		return nil, errors.New("git connector: Logger is required")
	}
	return &Connector{repo: d.Repo, ids: d.IDs, logger: d.Logger}, nil
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
		commit = normalizedCommit(commit)
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
	return commit
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
