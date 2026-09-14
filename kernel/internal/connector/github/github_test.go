package github

import (
	"context"
	"crypto/rand"
	"errors"
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

func TestSyncRetainsIdentityAcrossRepositories(t *testing.T) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader, Now: func() time.Time { return time.Unix(10, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(&Deps{API: fakeAPI{}, Owner: "github", Repos: []string{"docs", "roadmap"}, IDs: ids, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: func() time.Time { return time.Unix(20, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	var a appendFake
	got, err := c.Sync(context.Background(), &a)
	if err != nil || got.Appended != 4 || len(a.events) != 4 {
		t.Fatalf("result=%+v events=%d err=%v", got, len(a.events), err)
	}
	want := []string{
		"github/docs/workflow/7", "github/docs/deployment/9",
		"github/roadmap/workflow/7", "github/roadmap/deployment/9",
	}
	for i, event := range a.events {
		if event.NativeID != want[i] {
			t.Fatalf("event %d native_id=%q want=%q", i, event.NativeID, want[i])
		}
		repository := strings.Split(want[i], "/workflow/")[0]
		repository = strings.Split(repository, "/deployment/")[0]
		if !strings.Contains(string(event.Payload), `"repository":"`+repository+`"`) {
			t.Fatalf("event %d lost repository identity: %s", i, event.Payload)
		}
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

func TestNewChecksEveryRequiredDependencyAndDefaultsClock(t *testing.T) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	base := func() *Deps {
		return &Deps{API: fakeAPI{}, Owner: "github", Repos: []string{"docs"}, IDs: ids, Logger: slog.Default()}
	}
	for name, mutate := range map[string]func(*Deps){
		"api":    func(d *Deps) { d.API = nil },
		"ids":    func(d *Deps) { d.IDs = nil },
		"logger": func(d *Deps) { d.Logger = nil },
		"owner":  func(d *Deps) { d.Owner = "" },
	} {
		t.Run(name, func(t *testing.T) {
			d := base()
			mutate(d)
			if _, err := New(d); err == nil {
				t.Fatal("accepted incomplete dependencies")
			}
		})
	}
	if _, err := New(nil); err == nil {
		t.Fatal("accepted nil dependencies")
	}
	d, err := New(base())
	if err != nil || d.now == nil {
		t.Fatalf("default clock present=%t err=%v", d.now != nil, err)
	}
}

func TestSyncChecksEveryUninitializedConnectorField(t *testing.T) {
	ids, err := core.NewIDGenerator(core.IDGeneratorConfig{Entropy: rand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	for name, connector := range map[string]*Connector{
		"nil connector": nil,
		"api":           {ids: ids, now: time.Now},
		"ids":           {api: fakeAPI{}, now: time.Now},
		"clock":         {api: fakeAPI{}, ids: ids},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := connector.Sync(context.Background(), &appendFake{}); err == nil {
				t.Fatal("accepted uninitialized connector")
			}
		})
	}
}

func TestValidNameChecksEveryCharacterClassAndBoundary(t *testing.T) {
	for _, name := range []string{"A", "a", "0", "-", "_", "a.b"} {
		if !validName(name) {
			t.Fatalf("valid name %q rejected", name)
		}
	}
	for _, name := range []string{"", ". .", "..", strings.Repeat("a", 101), "@", "[", "`", "{", "/", ":"} {
		if validName(name) {
			t.Fatalf("invalid name %q accepted", name)
		}
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

func TestHTTPClientRejectsInvalidSchemesAndTransportErrors(t *testing.T) {
	for _, base := range []string{"%", "ftp://example.test"} {
		c := &HTTPClient{BaseURL: base, Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
		})}}
		if err := c.request(context.Background(), "endpoint", &map[string]any{}); err == nil {
			t.Fatalf("accepted base URL %q", base)
		}
	}
	want := errors.New("transport failed")
	c := &HTTPClient{Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, want })}}
	if err := c.request(context.Background(), "endpoint", &map[string]any{}); !errors.Is(err, want) {
		t.Fatalf("transport error=%v", err)
	}
}

func TestHTTPClientRejectsBothNonSuccessStatusBoundaries(t *testing.T) {
	for _, status := range []int{199, 300} {
		c := &HTTPClient{Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Status: "test status", Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
		})}}
		if err := c.request(context.Background(), "endpoint", &map[string]any{}); err == nil {
			t.Fatalf("accepted status %d", status)
		}
	}
}

func TestHTTPClientClampsCollectionLimit(t *testing.T) {
	if got, err := boundedLimit(1000); err != nil || got != maxItemsPerCollection {
		t.Fatalf("bounded limit=%d err=%v", got, err)
	}
	if _, err := boundedLimit(0); err == nil {
		t.Fatal("accepted zero collection limit")
	}
	if _, err := (&HTTPClient{}).Deployments(context.Background(), "github", "docs", 0); err == nil {
		t.Fatal("Deployments accepted zero collection limit")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
