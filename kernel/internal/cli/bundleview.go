package cli

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// `proofbound ui --bundle <dir>` (docs/plans/ui-bundle-spec.md, CLI-INV-60..62): a release bundle as a
// read-only local page, with no clone, store or repository, each warranty checked as `warranty verify`
// checks it and labelled with how far it was checked.

const bundleMaxFile = 4 << 20

// Results a warranty can have on the bundle page.
const (
	bundleSignatureOnly = "SIGNATURE ONLY"
	bundleInvalid       = "INVALID"
	bundleMissing       = "MISSING"
	bundleAltered       = "ALTERED"
	bundleNoWarranty    = "NO WARRANTY"
)

var warrantyFileRE = regexp.MustCompile(`^warranty-[0-9a-f]{12}\.json$`)

// bundleView is everything the page shows, computed once when the server starts.
type bundleView struct {
	Dir         string           `json:"dir"`
	Base        string           `json:"base"`
	Head        string           `json:"head"`
	ExporterKey string           `json:"index_exporter_key"`
	TrustKey    string           `json:"trust_key,omitempty"`
	Repo        string           `json:"repo,omitempty"`
	Counts      []bundleCount    `json:"counts"`
	Warranties  []bundleWarranty `json:"warranties"`
	Waived      []bundleWaived   `json:"waived_unsigned"`
	NonClaiming int              `json:"non_claiming_commits_unsigned"`
	Checked     string           `json:"checked_at"`
}

type bundleCount struct {
	Result string `json:"result"`
	N      int    `json:"n"`
}

// bundleWarranty is one row: a warranty file (or an index entry without one) and its check.
type bundleWarranty struct {
	File       string `json:"file,omitempty"`
	InIndex    bool   `json:"in_index"`
	IndexState string `json:"index_state_unsigned,omitempty"`
	Result     string `json:"result"`
	Problem    string `json:"problem,omitempty"`
	Key        string `json:"key,omitempty"`
	// From the signed body only (empty when the file does not parse).
	Commit    string          `json:"commit"`
	Change    string          `json:"change,omitempty"`
	Issuer    string          `json:"issuer,omitempty"`
	IssuedAt  string          `json:"issued_at,omitempty"`
	RepoHint  string          `json:"repository_hint,omitempty"`
	Promises  []bundlePromise `json:"promises,omitempty"`
	Reviews   []bundleReview  `json:"reviews,omitempty"`
	Reviewers []string        `json:"reviewers,omitempty"`
	Proven    []string        `json:"reviewers_proven,omitempty"`
	Cosigners []string        `json:"cosigners,omitempty"`
	Proof     *browserProof   `json:"-"`
	// Signed is true when the commit and everything below it come from a signed body that checked out.
	Signed bool `json:"signed_body_shown"`
}

type bundlePromise struct {
	Requirement string             `json:"requirement"`
	Digest      string             `json:"digest"`
	Obligations []bundleObligation `json:"obligations"`
}

type bundleObligation struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Outcome   string `json:"outcome"`
}

type bundleReview struct {
	Requirement string   `json:"requirement"`
	Reviewer    string   `json:"reviewer"`
	Outcomes    []string `json:"outcomes"`
}

// readBundleFile reads one regular file directly inside dir, refusing symlinks and anything larger
// than bundleMaxFile.
func readBundleFile(dir, name string) ([]byte, error) {
	p := filepath.Join(dir, name)
	info, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", name)
	}
	if info.Size() > bundleMaxFile {
		return nil, fmt.Errorf("%s is larger than %d bytes", name, bundleMaxFile)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, bundleMaxFile+1))
}

// loadBundle reads and checks a bundle folder (CLI-INV-61). A missing or unreadable index fails.
func loadBundle(ctx context.Context, dir, trustKey, repo string, now time.Time) (bundleView, error) {
	v := bundleView{Dir: dir, TrustKey: trustKey, Repo: repo, Warranties: []bundleWarranty{}, Waived: []bundleWaived{}, Checked: now.UTC().Format(time.RFC3339)}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return v, fmt.Errorf("ui --bundle: %s is not a folder", dir)
	}
	raw, err := readBundleFile(dir, "index.json")
	if err != nil {
		return v, fmt.Errorf("ui --bundle: %s has no readable index.json: %v", dir, err)
	}
	var ix bundleIndex
	if err := json.Unmarshal(raw, &ix); err != nil || ix.Schema != bundleSchema || !sha40RE.MatchString(ix.Base) || !sha40RE.MatchString(ix.Head) {
		return v, fmt.Errorf("ui --bundle: %s/index.json is not a Proofbound bundle index", dir)
	}
	v.Base, v.Head, v.ExporterKey, v.NonClaiming = ix.Base, ix.Head, ix.ExporterKey, ix.NonClaimingCommits
	if ix.Waived != nil {
		v.Waived = ix.Waived
	}
	named := map[string]bool{}
	seenCommit := map[string]bool{}
	for _, e := range ix.Entries {
		row := bundleWarranty{InIndex: true, IndexState: e.State, Commit: e.Commit}
		if seenCommit[e.Commit] {
			row.Result, row.Problem = bundleAltered, "the index lists this commit twice"
			v.Warranties = append(v.Warranties, row)
			continue
		}
		seenCommit[e.Commit] = true
		if e.File == "" {
			row.Result, row.Problem = bundleNoWarranty, e.Reason
			v.Warranties = append(v.Warranties, row)
			continue
		}
		row.File = e.File
		if !warrantyFileRE.MatchString(e.File) {
			row.Result, row.Problem = bundleAltered, "the index names a file that is not a warranty file name"
			v.Warranties = append(v.Warranties, row)
			continue
		}
		if named[e.File] {
			row.Result, row.Problem = bundleAltered, "the index names this file for a second commit"
			v.Warranties = append(v.Warranties, row)
			continue
		}
		named[e.File] = true
		data, err := readBundleFile(dir, e.File)
		if errors.Is(err, os.ErrNotExist) {
			row.Result, row.Problem = bundleMissing, "the index names this file but the folder does not hold it"
			v.Warranties = append(v.Warranties, row)
			continue
		}
		if err != nil {
			row.Result, row.Problem = bundleInvalid, err.Error()
			v.Warranties = append(v.Warranties, row)
			continue
		}
		row.Commit = ""
		checkWarranty(ctx, &row, data, trustKey, repo)
		bindRow(&row, e.Commit, v.Base, v.Head, repo)
		if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != e.FileSHA256 {
			problem := "its bytes differ from the index's file_sha256"
			if row.Problem != "" && row.Result != bundleSignatureOnly {
				problem += "; " + row.Problem
			}
			row.Result, row.Problem = bundleAltered, problem
		}
		if row.Commit == "" {
			row.Commit = e.Commit
		}
		v.Warranties = append(v.Warranties, row)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return v, err
	}
	for _, d := range entries {
		if !warrantyFileRE.MatchString(d.Name()) || named[d.Name()] {
			continue
		}
		row := bundleWarranty{File: d.Name()}
		data, err := readBundleFile(dir, d.Name())
		if err != nil {
			row.Result, row.Problem = bundleInvalid, err.Error()
		} else {
			checkWarranty(ctx, &row, data, trustKey, repo)
			bindRow(&row, "", v.Base, v.Head, repo)
		}
		v.Warranties = append(v.Warranties, row)
	}
	counts := map[string]int{}
	for _, w := range v.Warranties {
		counts[w.Result]++
	}
	for _, r := range []string{bundleValid, bundleUnreviewed, bundleSignatureOnly, bundleInvalid, bundleAltered, bundleMissing, bundleNoWarranty} {
		if counts[r] > 0 {
			v.Counts = append(v.Counts, bundleCount{Result: r, N: counts[r]})
		}
	}
	return v, nil
}

// checkWarranty runs `warranty verify` on one file's bytes. The signed body is shown only when the file
// itself checks out (SIGNATURE ONLY, VALID, UNREVIEWED): an INVALID file's contents are not evidence,
// so nothing from it is displayed but its key (review ui-bundle-review-20261123 F1).
func checkWarranty(ctx context.Context, row *bundleWarranty, data []byte, trustKey, repo string) {
	f, err := pbwarranty.Parse(data)
	if err != nil {
		row.Result, row.Problem = bundleInvalid, "not a well-formed warranty file: "+err.Error()
		return
	}
	if pub, err := f.PublicKey(); err == nil {
		row.Key = pbwarranty.Fingerprint(pub)
	}
	proven, err := pbwarranty.VerifyReport(ctx, f, repo, trustKey)
	switch {
	case err == nil:
		row.Result = bundleValid
	case errors.Is(err, pbwarranty.ErrUnverified):
		row.Result, row.Problem = bundleSignatureOnly, "the file is intact and signed by key "+row.Key+"; its claims were not checked against the source (open with --repo <clone> to check them)"
	case errors.Is(err, pbwarranty.ErrUnreviewed):
		row.Result, row.Problem = bundleUnreviewed, err.Error()
	case err.Error() == "signature":
		row.Result, row.Problem = bundleInvalid, "the signature does not match the file's contents: it was edited or forged, so none of it is shown"
		return
	default:
		row.Result, row.Problem = bundleInvalid, err.Error()+"; its contents are not shown"
		return
	}
	for _, p := range proven {
		row.Proven = append(row.Proven, p.Name+":"+p.Fingerprint)
	}
	row.Signed = true
	b := f.Body
	row.Commit, row.Change, row.Issuer, row.IssuedAt, row.RepoHint = b.Commit, b.ChangeName, b.Issuer, b.IssuedAt, b.RepositoryHint
	row.Reviewers = bodyReviewers(b)
	if pub, err := f.PublicKey(); err == nil {
		if proof, ok := warrantyProof(row.File, data, pub, f.Signature.Sig); ok {
			row.Proof = &proof
		}
	}
	for _, c := range f.Cosignatures {
		if raw, err := hex.DecodeString(c.Key); err == nil && len(raw) == ed25519.PublicKeySize {
			row.Cosigners = append(row.Cosigners, pbwarranty.Fingerprint(raw))
		}
	}
	row.Promises, row.Reviews = bundlePromises(b)
}

// shown reports whether a row's signed body checked out and is displayed.
func (w bundleWarranty) shown() bool {
	return w.Result == bundleSignatureOnly || w.Result == bundleValid || w.Result == bundleUnreviewed
}

// bindRow ties a checked file to what names it (F2, F3): the signed commit must be the one its file
// name and its index entry name, and, with a clone, inside the index's range. A mismatch is ALTERED;
// the signed body stays visible, since it is what the signature covers.
func bindRow(row *bundleWarranty, indexCommit, base, head, repo string) {
	if !row.shown() {
		return
	}
	var problems []string
	if strings.TrimSuffix(strings.TrimPrefix(row.File, "warranty-"), ".json") != short(row.Commit, 12) {
		problems = append(problems, "its file name does not match the commit it is signed for ("+short(row.Commit, 12)+")")
	}
	if indexCommit != "" && indexCommit != row.Commit {
		problems = append(problems, "the index lists it for commit "+short(printableUnit(indexCommit), 12)+", but it is signed for "+short(row.Commit, 12))
	}
	if repo != "" && !inRange(repo, base, head, row.Commit) {
		problems = append(problems, "in the clone, its commit is not in the index's range "+short(base, 12)+".."+short(head, 12))
	}
	if len(problems) > 0 {
		row.Result, row.Problem = bundleAltered, strings.Join(problems, "; ")
	}
}

// inRange reports whether commit is in base..head in the clone: an ancestor of head, not of base.
func inRange(repo, base, head, commit string) bool {
	anc := func(a, b string) bool {
		cmd := exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", "--end-of-options", a, b)
		cmd.Env = repositoryGitEnv()
		return cmd.Run() == nil
	}
	return anc(commit, head) && !anc(commit, base)
}

// warrantyProof is what the browser needs to re-verify the warranty's signature: the canonical body
// with the schema tag line in front, and the signature as a `.sig`-style line (the "file" proof kind).
func warrantyProof(label string, data []byte, pub ed25519.PublicKey, sig string) (browserProof, bool) {
	var outer struct {
		Body json.RawMessage `json:"body"`
	}
	if json.Unmarshal(data, &outer) != nil {
		return browserProof{}, false
	}
	canonical, err := core.Canonicalize(outer.Body)
	if err != nil {
		return browserProof{}, false
	}
	keyHex := hex.EncodeToString(pub)
	return browserProof{Kind: "file", Label: label, Bytes64: base64.StdEncoding.EncodeToString(canonical), Sig: "ed25519 " + keyHex + " " + sig, KeyHex: keyHex, Tag: pbwarranty.Schema + "\n"}, true
}

// bundlePromises lists, from the signed body, each targeted obligation with its statement and its
// verdict outcomes, and each requirement review's outcomes.
func bundlePromises(b pbwarranty.Body) ([]bundlePromise, []bundleReview) {
	outcomes := map[string][]string{}
	for _, v := range b.Verdicts {
		for _, o := range v.Obligations {
			k := o.RequirementID + "\x00" + o.ObligationID
			outcomes[k] = append(outcomes[k], o.Outcome+" ("+printableUnit(v.DeclaredReviewer)+")")
		}
	}
	targeted := map[string][]string{}
	for _, t := range b.ChangeIntent.Targets {
		targeted[t.RecordID] = append(targeted[t.RecordID], t.ObligationIDs...)
	}
	var promises []bundlePromise
	for _, r := range b.Requirements {
		statements := map[string]string{}
		for _, o := range r.Obligations {
			statements[o.ID] = o.Statement
		}
		p := bundlePromise{Requirement: r.RequirementID, Digest: r.ArtifactSHA256}
		ids := append([]string{}, targeted[r.RequirementID]...)
		sort.Strings(ids)
		for _, id := range ids {
			out := strings.Join(outcomes[r.RequirementID+"\x00"+id], ", ")
			if out == "" {
				out = "no verdict"
			}
			p.Obligations = append(p.Obligations, bundleObligation{ID: id, Statement: printableUnit(statements[id]), Outcome: out})
		}
		promises = append(promises, p)
	}
	var reviews []bundleReview
	for _, rv := range b.Reviews {
		r := bundleReview{Requirement: rv.Requirement.RecordID, Reviewer: printableUnit(rv.DeclaredReviewer)}
		for _, o := range rv.Outcomes {
			line := o.ObligationID + " " + o.Outcome
			if o.Finding != "" {
				line += ": " + printableUnit(o.Finding)
			}
			r.Outcomes = append(r.Outcomes, line)
		}
		reviews = append(reviews, r)
	}
	return promises, reviews
}

// serveBundle is the bundle page and its JSON twin.
func (s *uiServer) serveBundle(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, r, "bundle.html", s.bundle, s.bundle)
}

// runUIBundle serves a bundle folder (CLI-INV-60): no repository root, no store, no discovery file.
func runUIBundle(ctx context.Context, opts uiOptions, output io.Writer) error {
	dir, err := filepath.Abs(opts.bundle)
	if err != nil {
		return err
	}
	view, err := loadBundle(ctx, dir, opts.trustKey, opts.repo, time.Now())
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(opts.port)))
	if err != nil {
		return fmt.Errorf("ui: %w", err)
	}
	s := newUIServer("", "", "", listener.Addr().(*net.TCPAddr).Port)
	s.bundle = &view
	url := fmt.Sprintf("http://127.0.0.1:%d/?k=%s", s.port, s.token)
	if _, err := fmt.Fprintf(output, "ui %s (bundle %s; Ctrl-C to stop; the link is your key to this page, do not share it)\n", url, dir); err != nil {
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
