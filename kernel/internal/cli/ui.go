package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// `proofbound ui` (U1, docs/plans/ui-u1-audit-room-spec.md, CLI-INV-37..41): a loopback server for the
// Audit room. It holds the store only while it answers one request.

//go:embed uikit
var uiKit embed.FS

var uiPage = template.Must(template.New("audit.html").Funcs(template.FuncMap{
	"short": func(s string) string { return short(s, 12) },
	"lower": strings.ToLower,
	// label is the reader's word for a control state; the JSON keeps the state itself.
	"label": func(state string) string {
		return map[string]string{auditHeld: "Passed", auditExcepted: "Exception", auditListed: "Disclosed", auditNone: "Not assessed"}[state]
	},
	"outcomeLabel": func(o string) string {
		return map[string]string{"accepted": "Risk accepted", "fixed": "Fixed", "reverted": "Reverted"}[o]
	},
	"kindLabel": func(k string) string {
		return map[string]string{"gate": "Unreviewed change", "uncovered": "Code without a spec", "warranty": "Warranty does not verify", "trust": "Change to signing or the check", "order": "Spec written with or after its code", "self-review": "Reviewed by its own author or owner", "signature": "Signature does not verify", "area": "Change outside its declared area", "tests": "Change its tests do not catch", "provenance": "Provenance record does not verify"}[k]
	},
	"join":         func(xs []string) string { return strings.Join(xs, ", ") },
	"jumpLine":     jumpLineOf,
	"testsSummary": testsSummary,
	"agentWord":    agentWord,
	"seenWords":    seenWords,
	// agentAt is the agent's answer to a question key, if any (agent-reviewer-spec §1.3).
	"agentAt": func(by map[string]agentItem, format string, a ...any) *agentItem {
		if it, ok := by[fmt.Sprintf(format, a...)]; ok {
			return &it
		}
		return nil
	},
	// mapReason is the one-line reason a control maps to a criterion (frameworks.go, CLI-INV-65).
	"mapReason": func(control string, ref frameworkRef) string {
		for _, e := range frameworkMap[control] {
			if e.frameworkRef == ref {
				return e.Reason
			}
		}
		return ""
	},
	"frameworkCaveat": func() string { return frameworkCaveat },
	"frameworkDocURL": func() string { return frameworkDocURL },
	"dict": func(kv ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
		return m
	},
	"stateClass": func(state string) string {
		switch state {
		case promiseProven:
			return "passed"
		case promiseNotKept:
			return "exception"
		case promiseStale, promiseUnsure:
			return "disclosed"
		}
		return "none"
	},
}).ParseFS(uiKit, "uikit/*.html"))

const uiMaxBody = 16 << 10

// uiMaxReviewBody bounds a Review bench request: promises plus at most areaMaxTotal area decisions.
const uiMaxReviewBody = 1 << 20

type uiOptions struct {
	base string
	port int
	open bool
	// bundle, trustKey and repo are `ui --bundle <dir> [--trust-key <fp>] [--repo <clone>]` (CLI-INV-60).
	bundle, trustKey, repo string
	// agentReview is `--agent-review`: the person allows the bench to ask the project's agent
	// reviewer, which may cost money and sends the change out (agent-reviewer-spec.md §3).
	agentReview bool
}

// parseUI accepts `ui [--base <ref>] [--port <n>] [--open]`, each flag at most once.
func parseUI(args []string) (uiOptions, bool) {
	opts := uiOptions{}
	if len(args) == 0 || args[0] != "ui" {
		return opts, false
	}
	seen := map[string]bool{}
	rest := args[1:]
	for len(rest) > 0 {
		flag := rest[0]
		if seen[flag] {
			return opts, false
		}
		seen[flag] = true
		switch flag {
		case "--open":
			opts.open = true
			rest = rest[1:]
		case "--agent-review":
			opts.agentReview = true
			rest = rest[1:]
		case "--bundle", "--trust-key", "--repo":
			if len(rest) < 2 || rest[1] == "" || strings.HasPrefix(rest[1], "-") {
				return opts, false
			}
			switch flag {
			case "--bundle":
				opts.bundle = rest[1]
			case "--trust-key":
				opts.trustKey = rest[1]
			default:
				opts.repo = rest[1]
			}
			rest = rest[2:]
		case "--base", "--port":
			if len(rest) < 2 || rest[1] == "" || strings.HasPrefix(rest[1], "-") {
				return opts, false
			}
			if flag == "--base" {
				opts.base = rest[1]
			} else {
				n, err := strconv.Atoi(rest[1])
				if err != nil || n < 0 || n > 65535 || strconv.Itoa(n) != rest[1] {
					return opts, false
				}
				opts.port = n
			}
			rest = rest[2:]
		default:
			return opts, false
		}
	}
	// --trust-key and --repo only mean something for a bundle; a bundle has no period (CLI-INV-60).
	if opts.bundle == "" && (opts.trustKey != "" || opts.repo != "") || opts.bundle != "" && (opts.base != "" || opts.agentReview) {
		return opts, false
	}
	return opts, true
}

type uiServer struct {
	// token opens the page for the person (the terminal link); agentToken only reads, by bearer, and
	// is what .proofbound/ui.json holds (review ui-u2-review-20261122 F1).
	root, databaseURL, defaultBase, token, agentToken string
	port                                              int
	// agentReview: started with --agent-review (agent-reviewer-spec.md §3).
	agentReview bool
	// gate lets one request compute at a time; a later one waits up to wait for it, then gets 503,
	// so a reload during a slow first sync shows the answer instead of "busy" (owner try-out).
	gate chan struct{}
	wait time.Duration
	// compute syncs and computes the audit and the promises of a resolved base; tests replace it.
	compute func(ctx context.Context, base string) (ledgerView, string, error)
	// drafts builds the ledger review drafts of commits (one store open); tests replace it.
	drafts func(ctx context.Context, shas []string) (map[string]draftFile, error)
	// seal runs `review seal` on a draft folder with an open store, with the caller's fresh test run
	// when given; tests replace it.
	seal func(ctx context.Context, dir string, ran *sealTestsRun, output io.Writer) error
	// runTests runs a commit's declared tests, reporting each run; tests replace it.
	runTests func(ctx context.Context, sha string, say func(string)) (testsResult, error)
	// jobs is the bench's one background job; life is the server's own context, cancelled when it
	// stops, so a job outlives its request but not the server (ui-reviewer-spec §1.1, F6).
	// issueCompute builds the release bundle in memory for Issue (R2.2); tests replace it.
	issueCompute func(ctx context.Context, base string) (bundleIndex, map[string][]byte, error)
	jobs         *benchJobs
	life         context.Context
	stopLife     context.CancelFunc
	// asOf computes the audit of base..commit from a checkout of commit; tests replace it.
	asOf    func(ctx context.Context, base, commit string) (ledgerView, string, error)
	asOfMem map[string]ledgerView
	now     func() time.Time
	// bundle, when set, makes this a read-only bundle server: only / and /static/ (CLI-INV-60).
	bundle *bundleView
}

// ledgerView is what one store opening computes.
type ledgerView struct {
	Report   auditReport
	Promises []promiseChange
}

// uiCookie names the cookie per port, so this server's token is never offered to another (code
// review F8: cookies are not port-scoped).
func (s *uiServer) cookieName() string { return "proofbound_ui_" + strconv.Itoa(s.port) }

// uiView is what one audit request computed: the transcript report plus this machine's view.
type uiView struct {
	Report       auditReport       `json:"report"`
	BaseRef      string            `json:"base_ref"`
	Actor        actor             `json:"actor"`
	Attestations []attestationView `json:"attestations"`
	SyncError    string            `json:"sync_error,omitempty"`
	Tags         []string          `json:"-"`
	// Page-only: the repository's folder name, when this page was computed, and the summary tiles.
	Repo      string `json:"-"`
	BaseLabel string `json:"-"` // "v0.3.2", or "the first commit (6ad247c8a3d9)", or a short id
	// AsOf is set when the page shows a past moment (the time rail); Report is then that moment's.
	AsOf *asOfInfo `json:"as_of,omitempty"`
	// Rail is the period's commits for the time rail, oldest first (page only).
	Rail       []auditCommit   `json:"-"`
	Generated  string          `json:"-"`
	Summary    viewSummary     `json:"-"`
	Actors     []actorRow      `json:"-"`
	Track      trackReport     `json:"-"`
	TrackNotes []string        `json:"-"`
	Promises   []promiseChange `json:"-"`
	transcript []byte
}

type viewSummary struct{ Held, Tests, Exceptions, Closed, Disclosed, NotAssessed int }

func summarize(rep auditReport) viewSummary {
	var sum viewSummary
	for _, c := range rep.Controls {
		switch c.State {
		case auditHeld:
			sum.Held++
			sum.Tests++
		case auditExcepted:
			sum.Tests++
		case auditListed:
			sum.Disclosed++
		default:
			sum.NotAssessed++
		}
	}
	sum.Exceptions = len(rep.Exceptions)
	for _, e := range rep.Exceptions {
		if e.Outcome != nil && e.Outcome.Verified {
			sum.Closed++
		}
	}
	return sum
}

// actorRow is one author or agent's share of the period (the "Who did the work" section).
type actorRow struct {
	Name, Kind                                            string
	Commits, Claimed, Uncovered, Signed, Exceptions, Open int
}

// actorRows counts each commit for its author and for each agent named on it; agents first, then by
// commits. Names are git's self-declared identities.
func actorRows(rep auditReport) []actorRow {
	rows := map[string]*actorRow{}
	get := func(name, kind string) *actorRow {
		key := kind + "\x00" + name
		if rows[key] == nil {
			rows[key] = &actorRow{Name: name, Kind: kind}
		}
		return rows[key]
	}
	exc, open := map[string]int{}, map[string]int{}
	for _, e := range rep.Exceptions {
		exc[e.Commit]++
		if e.Outcome == nil || !e.Outcome.Verified {
			open[e.Commit]++
		}
	}
	for _, c := range rep.Commits {
		var who []*actorRow
		agentAuthor := false
		for _, a := range c.Agents {
			who = append(who, get(a, "agent"))
			agentAuthor = agentAuthor || a == c.Author
		}
		if !agentAuthor && c.Author != "" {
			who = append(who, get(c.Author, "person"))
		}
		for _, r := range who {
			r.Commits++
			if c.Signature == sigVerified {
				r.Signed++
			}
			r.Exceptions += exc[c.Commit]
			r.Open += open[c.Commit]
			switch c.Class {
			case classClaimed:
				r.Claimed++
			case classUncovered:
				r.Uncovered++
			}
		}
	}
	out := make([]actorRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind == "agent"
		}
		if out[i].Commits != out[j].Commits {
			return out[i].Commits > out[j].Commits
		}
		return out[i].Name < out[j].Name
	})
	return out
}

var errLedgerBusy = errors.New("the ledger is busy: another proofbound command is running; reload in a few seconds")

func runUI(ctx context.Context, root, databaseURL string, args []string, output io.Writer) error {
	opts, _ := parseUI(args)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(opts.port)))
	if err != nil {
		return fmt.Errorf("ui: %w", err)
	}
	s := newUIServer(root, databaseURL, opts.base, listener.Addr().(*net.TCPAddr).Port)
	s.agentReview = opts.agentReview
	url := fmt.Sprintf("http://127.0.0.1:%d/?k=%s", s.port, s.token)
	if _, err := fmt.Fprintf(output, "ui %s (Ctrl-C to stop; the link is your key to this page, do not share it)\n", url); err != nil {
		listener.Close()
		return err
	}
	if opts.open {
		if err := openBrowser(url); err != nil {
			fmt.Fprintf(output, "note: could not open a browser (%v); open %s yourself\n", err, url)
		}
	}
	// The discovery file lets this user's agents find the server and read it (U2 spec section 4).
	discovery := filepath.Join(stateRoot(root), "ui.json")
	if raw, err := json.Marshal(map[string]any{"url": fmt.Sprintf("http://127.0.0.1:%d/", s.port), "token": s.agentToken, "pid": os.Getpid()}); err == nil {
		if err := os.MkdirAll(stateRoot(root), 0o755); err == nil && os.WriteFile(discovery, append(raw, '\n'), 0o600) == nil {
			_ = os.Chmod(discovery, 0o600)
			defer os.Remove(discovery)
		}
	}
	// Stopping the server stops a running test job (its processes killed, its copy removed).
	defer func() {
		s.stopLife()
		s.jobs.wg.Wait()
	}()
	server := &http.Server{Handler: s, ReadHeaderTimeout: 10 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newUIServer(root, databaseURL, base string, port int) *uiServer {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		panic(err)
	}
	agentToken := make([]byte, 32)
	if _, err := rand.Read(agentToken); err != nil {
		panic(err)
	}
	s := &uiServer{root: root, databaseURL: databaseURL, defaultBase: base, port: port, token: hex.EncodeToString(token), agentToken: hex.EncodeToString(agentToken), now: time.Now,
		gate: make(chan struct{}, 1), wait: 2 * time.Minute}
	s.compute = func(ctx context.Context, base string) (ledgerView, string, error) {
		syncNote := ""
		var stderr bytes.Buffer
		if code := run(ctx, []string{"sync", "all"}, io.Discard, &stderr); code != 0 {
			syncNote = printableUnit(strings.TrimSpace(stderr.String()))
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return ledgerView{}, syncNote, err
		}
		var lv ledgerView
		lv.Report, err = computeAudit(ctx, root, ledger, base)
		if err == nil {
			var events []promiseEvent
			if events, err = readPromiseEvents(ctx, ledger); err == nil {
				lv.Promises = computePromises(events)
			}
		}
		return lv, syncNote, errors.Join(err, closeStore(ledger))
	}
	s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return nil, err
		}
		out := map[string]draftFile{}
		for _, sha := range shas {
			if d, err := ledgerDraft(ctx, ledger, sha); err == nil {
				out[sha] = d
			}
		}
		return out, closeStore(ledger)
	}
	s.asOfMem = map[string]ledgerView{}
	s.asOf = func(ctx context.Context, base, commit string) (ledgerView, string, error) {
		return asOfCompute(ctx, root, databaseURL, base, commit)
	}
	s.seal = func(ctx context.Context, dir string, ran *sealTestsRun, output io.Writer) error {
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return err
		}
		fresh := func(ctx context.Context, sha string) (draftFile, error) { return ledgerDraft(ctx, ledger, sha) }
		return errors.Join(reviewSealWith(ctx, root, dir, ran, fresh, output), closeStore(ledger))
	}
	s.runTests = func(ctx context.Context, sha string, say func(string)) (testsResult, error) {
		parent := strings.TrimSpace(gitOutput(ctx, root, "rev-parse", "--verify", "--quiet", sha+"^1"))
		if cfg, found, err := readTestsConfig(ctx, root, parent); found && err == nil {
			say(fmt.Sprintf("tests: running `%s` with each hunk of code undone (this runs the commit's own code)", printableUnit(cfg.Run)))
		}
		return commitTestsProgress(ctx, root, sha, say)
	}
	s.issueCompute = issueComputeDefault(root, databaseURL)
	s.jobs = &benchJobs{}
	s.life, s.stopLife = context.WithCancel(context.Background())
	return s
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (s *uiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Cache-Control", "no-store")
	port := strconv.Itoa(s.port)
	if r.Host != "127.0.0.1:"+port && r.Host != "localhost:"+port {
		http.Error(w, "this server answers only on its loopback address", http.StatusMisdirectedRequest)
		return
	}
	// The token reaches the browser only through the link printed in the terminal; it is then kept in
	// a SameSite=Strict, HttpOnly cookie and never written into a page (spec review F1).
	if k := r.URL.Query().Get("k"); k != "" && r.Method == http.MethodGet {
		if !constantEqual(k, s.token) {
			http.Error(w, "this link is not the one printed in the terminal", http.StatusForbidden)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: s.cookieName(), Value: s.token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		target := "/audit"
		if s.bundle != nil {
			target = "/"
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/static/") {
		c, err := r.Cookie(s.cookieName())
		bearer, hasBearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		// Agents read with the token from .proofbound/ui.json (GET only: agents read, people act).
		readByAgent := hasBearer && r.Method == http.MethodGet && s.bundle == nil && constantEqual(bearer, s.agentToken)
		if !readByAgent && (err != nil || !constantEqual(c.Value, s.token)) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "Open the link `proofbound ui` printed in your terminal.\n")
			return
		}
	}
	if r.Method == http.MethodPost {
		origin := r.Header.Get("Origin")
		if origin != "http://127.0.0.1:"+port && origin != "http://localhost:"+port {
			http.Error(w, "refused: not a request from this page", http.StatusForbidden)
			return
		}
	}
	if s.bundle != nil {
		// A bundle page is read-only: no act, no other workspace (CLI-INV-60).
		switch {
		case r.URL.Path == "/" && r.Method == http.MethodGet:
			s.serveBundle(w, r)
		case strings.HasPrefix(r.URL.Path, "/static/") && r.Method == http.MethodGet:
			sub, _ := fs.Sub(uiKit, "uikit/static")
			http.StripPrefix("/static/", http.FileServer(http.FS(sub))).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		http.Redirect(w, r, "/audit", http.StatusSeeOther)
	case r.URL.Path == "/audit" && r.Method == http.MethodGet:
		s.serveAudit(w, r)
	case r.URL.Path == "/now" && r.Method == http.MethodGet:
		s.serveNow(w, r)
	case r.URL.Path == "/promises" && r.Method == http.MethodGet:
		s.servePromises(w, r)
	case r.URL.Path == "/review" && r.Method == http.MethodGet:
		s.serveReview(w, r)
	case r.URL.Path == "/agent" && r.Method == http.MethodGet:
		s.serveAgentDoc(w, r)
	case r.URL.Path == "/agent/next" && r.Method == http.MethodGet:
		s.serveAgentNext(w, r)
	case r.URL.Path == "/change" && r.Method == http.MethodGet:
		s.serveChange(w, r)
	case r.URL.Path == "/act/issue" && r.Method == http.MethodPost:
		s.serveIssue(w, r)
	case r.URL.Path == "/review/job" && r.Method == http.MethodGet:
		s.serveJob(w, r)
	case (r.URL.Path == "/act/review-draft" || r.URL.Path == "/act/review-preview" || r.URL.Path == "/act/review-tests" || r.URL.Path == "/act/review-agent" || r.URL.Path == "/act/review-seal") && r.Method == http.MethodPost:
		s.serveReviewAct(w, r, strings.TrimPrefix(r.URL.Path, "/act/review-"))
	case r.URL.Path == "/transcript" && r.Method == http.MethodGet:
		s.serveTranscript(w, r)
	case (r.URL.Path == "/act/preview" || r.URL.Path == "/act/sign") && r.Method == http.MethodPost:
		s.serveAct(w, r, r.URL.Path == "/act/sign")
	case strings.HasPrefix(r.URL.Path, "/static/") && r.Method == http.MethodGet:
		sub, _ := fs.Sub(uiKit, "uikit/static")
		http.StripPrefix("/static/", http.FileServer(http.FS(sub))).ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}

// acquire takes the request gate, waiting at most s.wait and never past the request's own end.
func (s *uiServer) acquire(ctx context.Context) bool {
	timer := time.NewTimer(s.wait)
	defer timer.Stop()
	select {
	case s.gate <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	case <-timer.C:
		return false
	}
}

func (s *uiServer) release() { <-s.gate }

// view takes the request gate and computes the audit.
func (s *uiServer) view(ctx context.Context, baseRef string) (uiView, int, error) {
	if !s.acquire(ctx) {
		return uiView{}, http.StatusServiceUnavailable, errLedgerBusy
	}
	defer s.release()
	v, status, err := s.viewLocked(ctx, baseRef)
	// Another proofbound command may hold the ledger for a moment: retry for a short while.
	for tries := 0; errors.Is(err, errLedgerBusy) && tries < 10 && ctx.Err() == nil; tries++ {
		time.Sleep(time.Second)
		v, status, err = s.viewLocked(ctx, baseRef)
	}
	return v, status, err
}

// viewLocked syncs, opens the store, computes the audit and closes it; the gate is held.
func (s *uiServer) viewLocked(ctx context.Context, baseRef string) (uiView, int, error) {
	if baseRef == "" {
		baseRef = s.defaultBase
	}
	if baseRef == "" {
		baseRef = defaultAuditBase(ctx, s.root)
	}
	v := uiView{BaseRef: baseRef, Attestations: []attestationView{}}
	base, err := resolveCommit(ctx, s.root, baseRef)
	if err != nil {
		return v, http.StatusBadRequest, fmt.Errorf("base %q is not a commit in this repository", printableUnit(baseRef))
	}
	if _, err := gitBytes(ctx, s.root, "merge-base", "--is-ancestor", base, "HEAD"); err != nil {
		return v, http.StatusBadRequest, fmt.Errorf("base %q is not an ancestor of HEAD", printableUnit(baseRef))
	}
	lv, syncNote, err := s.compute(ctx, base)
	rep := lv.Report
	v.Promises = lv.Promises
	if v.Promises == nil {
		v.Promises = []promiseChange{}
	}
	v.SyncError = syncNote
	if errors.Is(err, store.ErrLocked) {
		return v, http.StatusServiceUnavailable, errLedgerBusy
	}
	if err != nil {
		return v, http.StatusInternalServerError, err
	}
	attachOutcomes(ctx, s.root, &rep)
	withReviewProofs(ctx, s.root, &rep)
	v.Report = rep
	v.Rail = rep.Commits
	if v.transcript, err = transcript(rep); err != nil {
		return v, http.StatusInternalServerError, err
	}
	v.Attestations = readAttestations(ctx, s.root, rep)
	v.Actor = localActor(ctx, s.root)
	v.Repo = filepath.Base(s.root)
	v.BaseLabel = printableUnit(baseRef)
	if roots := strings.Fields(gitOutput(ctx, s.root, "rev-list", "--max-parents=0", "HEAD")); containsString(roots, rep.Base) {
		v.BaseLabel = "the first commit (" + short(rep.Base, 12) + ")"
	} else if sha40RE.MatchString(baseRef) {
		v.BaseLabel = short(baseRef, 12)
	}
	v.Generated = s.now().Format("2 Jan 2006, 15:04 MST")
	v.Summary = summarize(rep)
	v.Actors = actorRows(rep)
	v.Track, v.TrackNotes = computeTrack(ctx, s.root, rep), trackNotes
	for _, t := range strings.Split(gitOutput(ctx, s.root, "tag", "--sort=-creatordate", "--merged", "HEAD"), "\n") {
		if t = strings.TrimSpace(t); t != "" && len(v.Tags) < 30 {
			v.Tags = append(v.Tags, t)
		}
	}
	return v, http.StatusOK, nil
}

func constantEqual(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

// viewAsOf computes now (for the rail and the comparison) and the moment asof, a commit of the
// period, from its own checkout and record (CLI-INV-50); cached per base and commit.
func (s *uiServer) viewAsOf(ctx context.Context, baseRef, asof string) (uiView, int, error) {
	if !s.acquire(ctx) {
		return uiView{}, http.StatusServiceUnavailable, errLedgerBusy
	}
	defer s.release()
	v, status, err := s.viewLocked(ctx, baseRef)
	if err != nil {
		return v, status, err
	}
	commit, err := resolveCommit(ctx, s.root, asof)
	inPeriod := false
	if err == nil {
		for _, c := range v.Report.Commits {
			inPeriod = inPeriod || c.Commit == commit
		}
	}
	if !inPeriod {
		return v, http.StatusBadRequest, fmt.Errorf("%q is not a commit of this period", printableUnit(asof))
	}
	key := v.Report.Base + ".." + commit
	then, ok := s.asOfMem[key]
	note := ""
	if !ok {
		if then, note, err = s.asOf(ctx, v.Report.Base, commit); err != nil {
			return v, http.StatusInternalServerError, err
		}
		// Only a moment whose record was built without error is kept (review ui-u3-review-20261122 F2).
		if note == "" {
			s.asOfMem[key] = then
		}
	}
	changed, onlyThen, onlyNow := compareMoments(then.Report, v.Report)
	info := &asOfInfo{Commit: commit, Changed: changed, OnlyThen: onlyThen, OnlyNow: onlyNow, SyncNote: note}
	for _, c := range v.Report.Commits {
		if c.Commit == commit {
			info.Date = c.Date
		}
	}
	v.AsOf = info
	v.Report = then.Report
	v.Promises = then.Promises
	v.Summary = summarize(then.Report)
	v.Actors = actorRows(then.Report)
	v.Track, v.TrackNotes = computeTrack(ctx, s.root, then.Report), trackNotes
	v.Attestations = []attestationView{}
	if v.transcript, err = transcript(then.Report); err != nil {
		return v, http.StatusInternalServerError, err
	}
	return v, http.StatusOK, nil
}

// defaultAuditBase is the newest tag reachable from HEAD^, else the root commit.
func defaultAuditBase(ctx context.Context, root string) string {
	if tag := strings.TrimSpace(gitOutput(ctx, root, "describe", "--tags", "--abbrev=0", "HEAD^")); tag != "" {
		return tag
	}
	roots := strings.Fields(gitOutput(ctx, root, "rev-list", "--max-parents=0", "HEAD"))
	if len(roots) > 0 {
		return roots[len(roots)-1]
	}
	return "HEAD"
}

func wantsJSON(r *http.Request) bool {
	return r.URL.Query().Get("format") == "json" || strings.Contains(r.Header.Get("Accept"), "application/json")
}

func (s *uiServer) serveAudit(w http.ResponseWriter, r *http.Request) {
	var v uiView
	var status int
	var err error
	if asof := r.URL.Query().Get("asof"); asof != "" {
		v, status, err = s.viewAsOf(r.Context(), r.URL.Query().Get("base"), asof)
	} else {
		v, status, err = s.view(r.Context(), r.URL.Query().Get("base"))
	}
	if err != nil {
		if wantsJSON(r) {
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		_ = uiPage.ExecuteTemplate(w, "error", map[string]string{"Message": err.Error()})
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, v)
		return
	}
	var page bytes.Buffer
	if err := uiPage.ExecuteTemplate(&page, "audit.html", v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page.Bytes())
}

func (s *uiServer) serveTranscript(w http.ResponseWriter, r *http.Request) {
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="proofbound-transcript-%s-%s.json"`, v.Report.Base[:12], v.Report.Head[:12]))
	_, _ = w.Write(v.transcript)
}

func (s *uiServer) serveAct(w http.ResponseWriter, r *http.Request, sign bool) {
	var req actRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, uiMaxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed request"})
		return
	}
	if !s.acquire(r.Context()) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": errLedgerBusy.Error()})
		return
	}
	defer s.release()
	// The audit, the validation and the write happen under one lock, so HEAD cannot move between them
	// from this server (spec review F10).
	v, status, err := s.viewLocked(r.Context(), req.Base)
	if err != nil {
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	at := s.now()
	if sign {
		if at, err = signTime(req.At, s.now()); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
	}
	rel, fields, err := buildAct(r.Context(), s.root, v.Report, v.transcript, v.Actor, req, at)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	digest, err := recordDigest(fields)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if sign && req.Digest != digest {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "the record changed since you reviewed it (the repository moved); review what will be written again"})
		return
	}
	if !sign {
		unsigned, err := canonicalJSON(fields)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		var pretty bytes.Buffer
		_ = json.Indent(&pretty, unsigned, "", "  ")
		writeJSON(w, http.StatusOK, map[string]string{"path": rel, "record": pretty.String(), "at": fields["at"].(string), "digest": digest, "signer": v.Actor.Name + " (key " + v.Actor.Fingerprint + ")"})
		return
	}
	data, err := signRecord(fields, v.Actor.key)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	kind := "exception outcome " + req.Exception
	if req.Kind == "attestation" {
		kind = "audit attestation " + v.Report.Base[:12] + ".." + v.Report.Head[:12]
	}
	committed, note, err := writeAndCommit(r.Context(), s.root, rel, data, "proofbound: "+kind)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": rel, "committed": committed, "note": note})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
