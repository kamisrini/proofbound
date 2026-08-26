// Package github ingests a bounded, read-only slice of GitHub delivery data.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

const Version = "github/1"

type WorkflowRun struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	HeadSHA    string    `json:"head_sha"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	HTMLURL    string    `json:"html_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Deployment struct {
	ID          int64     `json:"id"`
	Environment string    `json:"environment"`
	SHA         string    `json:"sha"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type API interface {
	WorkflowRuns(context.Context, string, string, int) ([]WorkflowRun, error)
	Deployments(context.Context, string, string, int) ([]Deployment, error)
}

type Appender interface {
	Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
	API    API
	Owner  string
	Repos  []string
	IDs    *core.IDGenerator
	Logger *slog.Logger
	Now    func() time.Time
}

type Connector struct {
	api    API
	owner  string
	repos  []string
	ids    *core.IDGenerator
	logger *slog.Logger
	now    func() time.Time
}

type Result struct {
	Listed, Appended, Existing int
	Cursor                     json.RawMessage
}

func New(d *Deps) (*Connector, error) {
	if d == nil || d.API == nil || d.IDs == nil || d.Logger == nil || !validName(d.Owner) {
		return nil, errors.New("github connector: API, Owner, IDs, and Logger are required")
	}
	if len(d.Repos) == 0 {
		return nil, errors.New("github connector: at least one repository is required")
	}
	repos := append([]string(nil), d.Repos...)
	sort.Strings(repos)
	for _, repo := range repos {
		if !validName(repo) {
			return nil, fmt.Errorf("github connector: invalid repository %q", repo)
		}
	}
	now := d.Now
	if now == nil {
		now = time.Now
	}
	return &Connector{api: d.API, owner: d.Owner, repos: repos, ids: d.IDs, logger: d.Logger, now: now}, nil
}

func (c *Connector) Sync(ctx context.Context, appender Appender) (Result, error) {
	var out Result
	if c == nil || c.api == nil || c.ids == nil || c.now == nil {
		return out, errors.New("github connector: connector is not initialized")
	}
	if appender == nil {
		return out, errors.New("github connector: appender is required")
	}
	for _, repo := range c.repos {
		runs, err := c.api.WorkflowRuns(ctx, c.owner, repo, 100)
		if err != nil {
			return out, fmt.Errorf("github connector: %s workflow runs: %w", repo, err)
		}
		deployments, err := c.api.Deployments(ctx, c.owner, repo, 100)
		if err != nil {
			return out, fmt.Errorf("github connector: %s deployments: %w", repo, err)
		}
		out.Listed += len(runs) + len(deployments)
		for _, run := range runs {
			payload, err := json.Marshal(map[string]any{"repository": c.owner + "/" + repo, "run_id": run.ID, "workflow": run.Name, "head_sha": run.HeadSHA, "status": run.Status, "conclusion": run.Conclusion, "url": run.HTMLURL, "created_at": run.CreatedAt.UTC(), "updated_at": run.UpdatedAt.UTC()})
			if err != nil {
				return out, err
			}
			if err := c.append(ctx, appender, core.KindGitHubWorkflow, fmt.Sprintf("%s/%s/workflow/%d", c.owner, repo, run.ID), run.UpdatedAt, payload, &out); err != nil {
				return out, err
			}
		}
		for _, deployment := range deployments {
			payload, err := json.Marshal(map[string]any{"repository": c.owner + "/" + repo, "deployment_id": deployment.ID, "environment": deployment.Environment, "sha": deployment.SHA, "url": deployment.URL, "created_at": deployment.CreatedAt.UTC(), "updated_at": deployment.UpdatedAt.UTC()})
			if err != nil {
				return out, err
			}
			if err := c.append(ctx, appender, core.KindGitHubDeployment, fmt.Sprintf("%s/%s/deployment/%d", c.owner, repo, deployment.ID), deployment.UpdatedAt, payload, &out); err != nil {
				return out, err
			}
		}
	}
	out.Cursor, _ = json.Marshal(map[string]any{"owner": c.owner, "repositories": c.repos, "synced_at": c.now().UTC()})
	return out, nil
}

func (c *Connector) append(ctx context.Context, appender Appender, kind core.Kind, native string, occurred time.Time, payload []byte, out *Result) error {
	e, err := c.ids.NewEvent(core.NewEventParams{Source: core.SourceGitHub, NativeID: native, Kind: kind, OccurredAt: occurred, Payload: payload, ConnectorVersion: Version})
	if err != nil {
		return err
	}
	_, inserted, err := appender.Append(ctx, e)
	if err != nil {
		return err
	}
	if inserted {
		out.Appended++
	} else {
		out.Existing++
	}
	return nil
}

func validName(s string) bool {
	if s == "" || s == "." || s == ".." || len(s) > 100 {
		return false
	}
	for _, r := range s {
		valid := r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.'
		if !valid {
			return false
		}
	}
	return true
}

type HTTPClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

const maxItemsPerCollection = 100

func (c *HTTPClient) request(ctx context.Context, endpoint string, out any) error {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = "https://api.github.com"
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("github client: invalid base URL")
	}
	ep, err := url.Parse(endpoint)
	if err != nil {
		return errors.New("github client: invalid endpoint")
	}
	u.Path = path.Join(u.Path, ep.Path)
	u.RawQuery = ep.RawQuery
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("github client: HTTP %s", resp.Status)
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("github client: decode: %w", err)
	}
	return nil
}

func (c *HTTPClient) WorkflowRuns(ctx context.Context, owner, repo string, limit int) ([]WorkflowRun, error) {
	var err error
	limit, err = boundedLimit(limit)
	if err != nil {
		return nil, err
	}
	var v struct {
		Runs []WorkflowRun `json:"workflow_runs"`
	}
	err = c.request(ctx, fmt.Sprintf("repos/%s/%s/actions/runs?per_page=%d", url.PathEscape(owner), url.PathEscape(repo), limit), &v)
	return v.Runs, err
}
func (c *HTTPClient) Deployments(ctx context.Context, owner, repo string, limit int) ([]Deployment, error) {
	var err error
	limit, err = boundedLimit(limit)
	if err != nil {
		return nil, err
	}
	var v []Deployment
	err = c.request(ctx, fmt.Sprintf("repos/%s/%s/deployments?per_page=%d", url.PathEscape(owner), url.PathEscape(repo), limit), &v)
	return v, err
}

func boundedLimit(limit int) (int, error) {
	if limit <= 0 {
		return 0, errors.New("github client: limit must be positive")
	}
	if limit > maxItemsPerCollection {
		return maxItemsPerCollection, nil
	}
	return limit, nil
}
