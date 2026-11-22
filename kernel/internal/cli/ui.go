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
	"runtime"
	"strconv"
	"strings"
	"sync"
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
}).ParseFS(uiKit, "uikit/audit.html"))

const uiMaxBody = 16 << 10

type uiOptions struct {
	base string
	port int
	open bool
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
	return opts, true
}

type uiServer struct {
	root, databaseURL, defaultBase, token string
	port                                  int
	mu                                    sync.Mutex
	// compute syncs and computes the audit of a resolved base; tests replace it.
	compute func(ctx context.Context, base string) (auditReport, string, error)
	now     func() time.Time
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
	transcript   []byte
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
	s := &uiServer{root: root, databaseURL: databaseURL, defaultBase: base, port: port, token: hex.EncodeToString(token), now: time.Now}
	s.compute = func(ctx context.Context, base string) (auditReport, string, error) {
		syncNote := ""
		var stderr bytes.Buffer
		if code := run(ctx, []string{"sync", "all"}, io.Discard, &stderr); code != 0 {
			syncNote = printableUnit(strings.TrimSpace(stderr.String()))
		}
		ledger, err := openStore(ctx, root, databaseURL)
		if err != nil {
			return auditReport{}, syncNote, err
		}
		rep, err := computeAudit(ctx, root, ledger, base)
		return rep, syncNote, errors.Join(err, closeStore(ledger))
	}
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
		http.Redirect(w, r, "/audit", http.StatusSeeOther)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/static/") {
		c, err := r.Cookie(s.cookieName())
		if err != nil || !constantEqual(c.Value, s.token) {
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
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		http.Redirect(w, r, "/audit", http.StatusSeeOther)
	case r.URL.Path == "/audit" && r.Method == http.MethodGet:
		s.serveAudit(w, r)
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

// view takes the one request lock (or answers busy at once) and computes the audit.
func (s *uiServer) view(ctx context.Context, baseRef string) (uiView, int, error) {
	if !s.mu.TryLock() {
		return uiView{}, http.StatusServiceUnavailable, errLedgerBusy
	}
	defer s.mu.Unlock()
	return s.viewLocked(ctx, baseRef)
}

// viewLocked syncs, opens the store, computes the audit and closes it; s.mu is held.
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
	rep, syncNote, err := s.compute(ctx, base)
	v.SyncError = syncNote
	if errors.Is(err, store.ErrLocked) {
		return v, http.StatusServiceUnavailable, errLedgerBusy
	}
	if err != nil {
		return v, http.StatusInternalServerError, err
	}
	attachOutcomes(ctx, s.root, &rep)
	v.Report = rep
	if v.transcript, err = transcript(rep); err != nil {
		return v, http.StatusInternalServerError, err
	}
	v.Attestations = readAttestations(ctx, s.root, rep)
	v.Actor = localActor(ctx, s.root)
	for _, t := range strings.Split(gitOutput(ctx, s.root, "tag", "--sort=-creatordate", "--merged", "HEAD"), "\n") {
		if t = strings.TrimSpace(t); t != "" && len(v.Tags) < 30 {
			v.Tags = append(v.Tags, t)
		}
	}
	return v, http.StatusOK, nil
}

func constantEqual(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

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
	v, status, err := s.view(r.Context(), r.URL.Query().Get("base"))
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
	if err := uiPage.Execute(&page, v); err != nil {
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
	if !s.mu.TryLock() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": errLedgerBusy.Error()})
		return
	}
	defer s.mu.Unlock()
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
