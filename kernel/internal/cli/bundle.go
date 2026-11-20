package cli

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// The shareable proof bundle (MVP step 4, docs/plans/mvp-4-shareable-proof-spec.md draft 2,
// CLI-INV-28..30). Only the warranty files are signed; index.json, VERIFY.md and badge.svg are the
// exporter's unsigned local result at head and say so.

const (
	bundleSchema        = "proofbound.warranty-bundle.v1"
	bundleValid         = "VALID"
	bundleUnreviewed    = "UNREVIEWED"
	bundleNotExportable = "NOT EXPORTABLE"
	bundleNote          = "Unsigned: the exporter's local result at head. Only the warranty files are signed; check each with `proofbound warranty verify <file> --repo <clone>`."
)

type bundleEntry struct {
	Commit           string   `json:"commit"`
	Claims           string   `json:"claims"`
	State            string   `json:"state"`
	File             string   `json:"file,omitempty"`
	FileSHA256       string   `json:"file_sha256,omitempty"`
	ReviewersProven  []string `json:"reviewers_proven,omitempty"`
	ReviewersMissing []string `json:"reviewers_missing,omitempty"`
	Reason           string   `json:"reason,omitempty"`
}

type bundleIndex struct {
	Schema             string        `json:"schema"`
	Unsigned           bool          `json:"unsigned"`
	Note               string        `json:"note"`
	Base               string        `json:"base"`
	Head               string        `json:"head"`
	ExporterKey        string        `json:"exporter_key"`
	NonClaimingCommits int           `json:"non_claiming_commits"`
	Entries            []bundleEntry `json:"entries"`
}

// errBundleIncomplete makes `warranty bundle` exit 3 after it has written a bundle in which some
// claiming commit is not VALID (exit 1 stays "error, nothing written").
var errBundleIncomplete = errors.New("bundle: written, but not every claiming commit is VALID")

func warrantyBundle(ctx context.Context, root string, ledger *store.Store, base, dir string, output io.Writer) error {
	if dir == "" {
		dir = "warranties"
	}
	if !filepath.IsAbs(dir) {
		wd, err := workingDirectory()
		if err != nil {
			return err
		}
		dir = filepath.Join(wd, dir)
	}
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return fmt.Errorf("bundle: %s exists and is not empty; nothing written", dir)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("bundle: %s: %w", dir, err)
	}
	if dirty := strings.TrimSpace(gitOutput(ctx, root, "status", "--porcelain", "--", "gates", ".proofbound/keys")); dirty != "" {
		return fmt.Errorf("bundle: gates/ or .proofbound/keys/ differ from HEAD; commit them first so a clean clone reproduces every state:\n%s", dirty)
	}
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return err
	}
	key, err := pbwarranty.LoadKey(keyPath)
	if err != nil {
		return err
	}
	baseSHA, err := resolveCommit(ctx, root, base)
	if err != nil {
		return err
	}
	headSHA, err := resolveCommit(ctx, root, "HEAD")
	if err != nil {
		return err
	}
	lines, err := classifiedRange(ctx, root, ledger, base)
	if err != nil {
		return err
	}
	all, err := commitsInRange(ctx, root, base)
	if err != nil {
		return err
	}
	index := bundleIndex{Schema: bundleSchema, Unsigned: true, Note: bundleNote, Base: baseSHA, Head: headSHA,
		ExporterKey: pbwarranty.Fingerprint(key.Public().(ed25519.PublicKey)), NonClaimingCommits: len(all) - len(lines), Entries: []bundleEntry{}}
	files := map[string][]byte{}
	for _, l := range lines {
		e := bundleEntry{Commit: l.commit, Claims: l.claims, State: l.state, Reason: l.reason}
		if l.state == checkPass {
			data, body, proven, unreviewed, err := buildWarranty(ctx, root, ledger, l.commit)
			if err != nil {
				e.State, e.Reason = bundleNotExportable, err.Error()
			} else {
				e.File = "warranty-" + l.commit[:12] + ".json"
				sum := sha256.Sum256(data)
				e.FileSHA256 = hex.EncodeToString(sum[:])
				files[e.File] = data
				e.State, e.Reason = bundleValid, ""
				if unreviewed != nil {
					e.State, e.Reason = bundleUnreviewed, unreviewed.Error()
				}
				provenNames := map[string]bool{}
				for _, p := range proven {
					e.ReviewersProven = append(e.ReviewersProven, p.Name+":"+p.Fingerprint)
					provenNames[p.Name] = true
				}
				for _, name := range bodyReviewers(body) {
					if !provenNames[name] {
						e.ReviewersMissing = append(e.ReviewersMissing, name)
					}
				}
				sort.Strings(e.ReviewersMissing)
			}
		}
		index.Entries = append(index.Entries, e)
	}
	raw, err := json.Marshal(index)
	if err != nil {
		return err
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return err
	}
	files["index.json"] = append(canonical, '\n')
	files["VERIFY.md"] = []byte(verifyPage(index))
	files["badge.svg"] = []byte(bundleBadge(index))
	if err := writeDirAtomically(dir, files); err != nil {
		return err
	}

	valid, unreviewedCount := 0, 0
	for _, e := range index.Entries {
		var line string
		switch e.State {
		case bundleValid, bundleUnreviewed:
			if e.State == bundleValid {
				valid++
			} else {
				unreviewedCount++
			}
			line = fmt.Sprintf("warranty %s commit=%s claims=%s %s", filepath.Join(dir, e.File), e.Commit, printableUnit(e.Claims), e.State)
			if len(e.ReviewersProven) > 0 {
				line += " reviewers=" + strings.Join(e.ReviewersProven, ",")
			}
			if len(e.ReviewersMissing) > 0 {
				line += " missing=" + printableUnit(strings.Join(e.ReviewersMissing, ","))
			}
		default:
			line = fmt.Sprintf("not warranted %s %s %s  %s", e.Commit[:12], printableUnit(e.Claims), e.State, printableUnit(e.Reason))
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "bundle %s (%s..%s): %d of %d claims verified (%d VALID, %d UNREVIEWED); VERIFY.md, index.json, badge.svg\n",
		dir, baseSHA[:12], headSHA[:12], valid, len(index.Entries), valid, unreviewedCount); err != nil {
		return err
	}
	if valid != len(index.Entries) {
		return errBundleIncomplete
	}
	return nil
}

// writeDirAtomically writes files into a fresh sibling of dir and renames it into place, replacing at
// most an empty dir; on any error nothing new remains (CLI-INV-30).
func writeDirAtomically(dir string, files map[string][]byte) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), ".warranty-bundle-*")
	if err != nil {
		return err
	}
	fail := func(err error) error { os.RemoveAll(tmp); return err }
	if err := os.Chmod(tmp, 0o755); err != nil {
		return fail(err)
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0o644); err != nil {
			return fail(err)
		}
	}
	if err := os.Remove(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fail(fmt.Errorf("bundle: %s: %w", dir, err))
	}
	if err := os.Rename(tmp, dir); err != nil {
		return fail(err)
	}
	return nil
}

// verifyPage is the human page a release carries. Every untrusted cell goes through mdCell (F3).
func verifyPage(ix bundleIndex) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Proofbound warranties for %s..%s\n\n", ix.Base[:12], ix.Head[:12])
	b.WriteString("> **This page and `index.json` are not signed.** They are the exporter's local result at commit\n")
	fmt.Fprintf(&b, "> `%s`. Only the `warranty-*.json` files are signed; trust a claim only after checking its file\n", ix.Head)
	b.WriteString("> yourself with `proofbound warranty verify` (below).\n\n")
	fmt.Fprintf(&b, "Range: `%s..%s`. Each warranty says one commit implements a spec change and was checked, obligation\n", ix.Base, ix.Head)
	b.WriteString("by obligation, by someone other than the spec's owner. Only commits that claim a spec (`Intent: sdd:`)\n")
	fmt.Fprintf(&b, "are covered: **%d commit(s) in this range claim none and are not covered by any warranty.**\n\n", ix.NonClaimingCommits)
	b.WriteString("| Commit | Claims | State | File | Reviewers who signed | Reviewers who have not |\n|---|---|---|---|---|---|\n")
	dash := func(xs []string) string {
		if len(xs) == 0 {
			return "—"
		}
		return mdCell(strings.Join(xs, ", "))
	}
	for _, e := range ix.Entries {
		file := "—"
		if e.File != "" {
			file = "`" + e.File + "`"
		}
		state := e.State
		if e.Reason != "" && e.State != bundleValid {
			state += ": " + e.Reason
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s | %s |\n", e.Commit[:12], mdCell(e.Claims), mdCell(state), file, dash(e.ReviewersProven), dash(e.ReviewersMissing))
	}
	if len(ix.Entries) == 0 {
		b.WriteString("| — | — | no commit in this range claims a spec | — | — | — |\n")
	}
	fmt.Fprintf(&b, `
## Verify it yourself

1. Install Proofbound: `+"`curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh`"+`
2. Clone the project **from a source you trust** (not from whoever sent you these files) and check out
   `+"`%s`"+`.
3. Get the exporter's key fingerprint **out of band** (from the project, not from this page). This page
   shows `+"`%s`"+` only so you can compare. Then, for each file:

   `+"```"+`
   proofbound warranty verify <warranty-file> --repo <clone> --trust-key <fingerprint you obtained>
   `+"```"+`

| Result | Meaning |
|---|---|
| `+"`VALID`"+` (exit 0) | the spec text, the review files and every signature check out against the clone, and each named reviewer's own declared key signed |
| `+"`UNREVIEWED`"+` (exit 3) | everything checks out except that a named reviewer has not signed |
| `+"`UNVERIFIED`"+` (exit 2) | you did not pass `+"`--repo`"+`; a signature alone proves nothing |
| `+"`INVALID`"+` (exit 1) | something does not match; the message names the first check that failed |

Reviewer keys are those the repository declares under `+"`.proofbound/keys/`"+` at that commit; a VALID
result proves "the key declared for this name signed", not who holds it. A later commit can change
what verifies; a later cosign can turn UNREVIEWED into VALID. Re-run the check rather than trusting
this page.
`, ix.Head, ix.ExporterKey)
	return b.String()
}

// bundleBadge is a deterministic flat SVG badge over the index states (CLI-INV-29).
func bundleBadge(ix bundleIndex) string {
	valid, onlyValidOrUnreviewed := 0, true
	for _, e := range ix.Entries {
		switch e.State {
		case bundleValid:
			valid++
		case bundleUnreviewed:
		default:
			onlyValidOrUnreviewed = false
		}
	}
	n := len(ix.Entries)
	text, color := fmt.Sprintf("%d of %d claims verified", valid, n), "#e05d44"
	switch {
	case n == 0:
		text, color = "no claims", "#9f9f9f"
	case valid == n:
		color = "#4c1"
	case onlyValidOrUnreviewed:
		color = "#dfb317"
	}
	const label = "proofbound"
	title := fmt.Sprintf("proofbound: %s for %s..%s (unsigned local result; verify the warranty files)", text, ix.Base[:12], ix.Head[:12])
	lw, rw := 7*len(label)+10, 7*len(text)+10
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s">`+
		`<title>%s</title><rect width="%d" height="20" fill="#555"/><rect x="%d" width="%d" height="20" fill="%s"/>`+
		`<g fill="#fff" font-family="Verdana,DejaVu Sans,sans-serif" font-size="11" text-anchor="middle">`+
		`<text x="%d" y="14">%s</text><text x="%d" y="14">%s</text></g></svg>`+"\n",
		lw+rw, html.EscapeString(title), html.EscapeString(title), lw, lw, rw, color, lw/2, label, lw+rw/2, html.EscapeString(text))
}
