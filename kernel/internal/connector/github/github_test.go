package github

import (
	"context"
	"crypto/rand"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type fakeAPI struct{ conclusion string }

func (f fakeAPI) WorkflowRuns(context.Context, string, string, int) ([]WorkflowRun, error) {
	conclusion := f.conclusion
	if conclusion == "" {
		conclusion = "success"
	}
	return []WorkflowRun{{ID: 7, Name: "CI", HeadSHA: "0123456789012345678901234567890123456789", Status: "completed", Conclusion: conclusion, CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(2, 0)}}, nil
}
func (fakeAPI) Deployments(context.Context, string, string, int) ([]Deployment, error) {
	return []Deployment{{ID: 9, Environment: "production", SHA: "0123456789012345678901234567890123456789", CreatedAt: time.Unix(3, 0), UpdatedAt: time.Unix(4, 0)}}, nil
}

type appendFake struct{ events []core.Event }

func (a *appendFake) Append(_ context.Context, e core.Event) (store.Record, bool, error) {
	for _, prior := range a.events {
		if prior.Source == e.Source && prior.NativeID == e.NativeID && prior.ContentSHA == e.ContentSHA {
			return store.Record{Event: prior}, false, nil
		}
	}
	a.events = append(a.events, e)
	return store.Record{Event: e}, true, nil
}

func TestSyncEmitsQualifiedWorkflowAndDeploymentEvents(t *testing.T) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader, Now: func() time.Time { return time.Unix(10, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(&Deps{API: fakeAPI{}, Owner: "github", Repos: []string{"docs"}, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: func() time.Time { return time.Unix(20, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	var a appendFake
	got, err := c.Sync(context.Background(), &a)
	if err != nil {
		t.Fatal(err)
	}
	if got.Listed != 2 || got.Appended != 2 || len(a.events) != 2 {
		t.Fatalf("result=%+v events=%d", got, len(a.events))
	}
	if a.events[0].Source != core.SourceGitHub || a.events[0].NativeID != "github/docs/workflow/7" || a.events[1].NativeID != "github/docs/deployment/9" {
		t.Fatalf("events=%+v", a.events)
	}
	for _, e := range a.events {
		if string(e.Payload) == "" || e.ContentSHA == "" {
			t.Fatalf("invalid event=%+v", e)
		}
	}
	got, err = c.Sync(context.Background(), &a)
	if err != nil || got.Appended != 0 || got.Existing != 2 {
		t.Fatalf("replay result=%+v err=%v", got, err)
	}
}

func TestNewRejectsUnsafeRepository(t *testing.T) {
	ids, _ := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader})
	if _, err := New(&Deps{API: fakeAPI{}, Owner: "github", Repos: []string{"docs/actions"}, IDs: ids, Logger: slog.Default()}); err == nil {
		t.Fatal("accepted unsafe repository")
	}
	if _, err := New(&Deps{API: fakeAPI{}, Owner: "..", Repos: []string{"docs"}, IDs: ids, Logger: slog.Default()}); err == nil {
		t.Fatal("accepted unsafe owner")
	}
}

func TestSyncChangedUpstreamRecordCreatesRevision(t *testing.T) {
	ids, _ := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader})
	now := func() time.Time { return time.Unix(20, 0) }
	makeConnector := func(api API) *Connector {
		c, err := New(&Deps{API: api, Owner: "github", Repos: []string{"docs"}, IDs: ids, Logger: slog.Default(), Now: now})
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	var a appendFake
	if _, err := makeConnector(fakeAPI{}).Sync(context.Background(), &a); err != nil {
		t.Fatal(err)
	}
	got, err := makeConnector(fakeAPI{conclusion: "failure"}).Sync(context.Background(), &a)
	if err != nil || got.Appended != 1 || got.Existing != 1 || len(a.events) != 3 {
		t.Fatalf("revision result=%+v events=%d err=%v", got, len(a.events), err)
	}
}

func TestHTTPClientPreservesQueryAndUsesHeaderAuth(t *testing.T) {
	var gotURL string
	var gotAuth string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotURL = r.URL.String()
		gotAuth = r.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"workflow_runs":[]}`)),
			Header:     make(http.Header),
		}, nil
	})
	client := &http.Client{Transport: transport}
	base, _ := url.Parse("https://example.test/api")
	c := &HTTPClient{BaseURL: base.String(), Token: "secret", Client: client}
	if _, err := c.WorkflowRuns(context.Background(), "github", "docs", 100); err != nil {
		t.Fatal(err)
	}
	if gotURL != "https://example.test/api/repos/github/docs/actions/runs?per_page=100" {
		t.Fatalf("request URL=%s", gotURL)
	}
	if gotAuth != "Bearer secret" {
		t.Fatalf("authorization header missing")
	}
	if strings.Contains(gotURL, "secret") {
		t.Fatal("token leaked into URL")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
