package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// keysInit creates the signing key outside any work tree and writes its public half to
// .proofbound/keys/<name>.pub for the owner to commit (warranty SPEC §3).
func keysInit(ctx context.Context, root, name string, output io.Writer) error {
	if name == "" {
		name = gitOutput(ctx, root, "config", "user.name")
	}
	name = pbwarranty.KeyName(name)
	if name == "" {
		return errors.New("keys init: no name; pass --name <name>")
	}
	pubPath := filepath.Join(root, filepath.FromSlash(pbwarranty.KeysDir), name+".pub")
	if _, err := os.Lstat(pubPath); err == nil {
		return fmt.Errorf("keys init: %s exists", pubPath)
	}
	path, err := pbwarranty.KeyPath()
	if err != nil {
		return err
	}
	pub, err := pbwarranty.InitKey(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pubPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(pubPath, []byte(pbwarranty.PubLine(pub)), 0o644); err != nil {
		return err
	}
	if err := appendMissingLines(output, root, ".gitignore", []string{"!.proofbound/keys/"}); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "key %s\nprivate %s (never commit it)\npublic %s (commit it: it declares the key for this repository)\n", pbwarranty.Fingerprint(pub), path, pubPath)
	return err
}

// claimedChange is the one `Intent: sdd:<name>` claim in the commit's message.
func claimedChange(ctx context.Context, root, sha string) (string, error) {
	var names []string
	for _, line := range strings.Split(gitOutput(ctx, root, "log", "-1", "--format=%B", sha), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Intent: sdd:"); ok {
			names = append(names, strings.TrimSpace(v))
		}
	}
	if len(names) != 1 {
		return "", fmt.Errorf("warranty: commit %s must claim exactly one sdd change (found %d)", sha, len(names))
	}
	return names[0], nil
}

// warrantyBody assembles the body from ledgered events only: the commit's one bound sdd change
// intent, its targeted requirements, and the latest revision of every review and verdict artifact
// that binds to them, which are the records the readiness gate counted.
func warrantyBody(ctx context.Context, ledger *store.Store, sha string) (pbwarranty.Body, error) {
	var commit *connectorgit.Commit
	changes := map[string]connectorintent.ChangeIntent{}
	requirements := map[string]connectorintent.Requirement{}
	verdicts := map[string]reviews.ObligationVerdict{}
	reviewsByPath := map[string]reviews.RequirementReview{}
	var order []string
	err := ledger.ReadEvents(ctx, store.Filter{}, func(record store.Record) error {
		e := record.Event
		switch {
		case e.Source == core.SourceGit && e.Kind == core.KindCommitRecorded && e.NativeID == sha:
			commit = &connectorgit.Commit{}
			return json.Unmarshal(e.Payload, commit)
		case e.Source == core.SourceIntentSDD && e.Kind == core.KindChangeIntent:
			var ci connectorintent.ChangeIntent
			if err := json.Unmarshal(e.Payload, &ci); err != nil {
				return err
			}
			changes[ci.IntentID+"@"+ci.ArtifactSHA256] = ci
		case e.Source == core.SourceIntentSDD && e.Kind == core.KindRequirement:
			var r connectorintent.Requirement
			if err := json.Unmarshal(e.Payload, &r); err != nil {
				return err
			}
			requirements[r.RequirementID+"@"+r.ArtifactSHA256] = r
		case e.Source == core.SourceReviews && e.Kind == core.KindReviewVerdict:
			var head struct {
				Schema string `json:"schema"`
			}
			if json.Unmarshal(e.Payload, &head) != nil || head.Schema != "proofbound.obligation-verdict.v2" {
				return nil
			}
			var v reviews.ObligationVerdict
			if err := json.Unmarshal(e.Payload, &v); err != nil {
				return err
			}
			if _, seen := verdicts[v.ArtifactPath]; !seen {
				order = append(order, "v"+v.ArtifactPath)
			}
			verdicts[v.ArtifactPath] = v
		case e.Source == core.SourceReviews && e.Kind == core.KindRequirementReview:
			var r reviews.RequirementReview
			if err := json.Unmarshal(e.Payload, &r); err != nil {
				return err
			}
			if _, seen := reviewsByPath[r.ArtifactPath]; !seen {
				order = append(order, "r"+r.ArtifactPath)
			}
			reviewsByPath[r.ArtifactPath] = r
		}
		return nil
	})
	if err != nil {
		return pbwarranty.Body{}, err
	}
	if commit == nil {
		return pbwarranty.Body{}, fmt.Errorf("warranty: commit %s is not in the ledger; run `proofbound sync all`", sha)
	}
	var bound []connectorgit.IntentRef
	for _, ref := range commit.IntentRefs {
		if ref.Provider == "sdd" {
			bound = append(bound, ref)
		}
	}
	if len(bound) != 1 {
		return pbwarranty.Body{}, fmt.Errorf("warranty: commit %s binds %d sdd changes; slice 1 warrants exactly one", sha, len(bound))
	}
	ci, ok := changes[bound[0].RecordID+"@"+bound[0].ArtifactSHA256]
	if !ok {
		return pbwarranty.Body{}, fmt.Errorf("warranty: change intent %s is not in the ledger", bound[0].RecordID)
	}
	body := pbwarranty.Body{Schema: pbwarranty.Schema, Commit: sha, ChangeIntent: ci, Requirements: []connectorintent.Requirement{}, Reviews: []reviews.RequirementReview{}, Verdicts: []reviews.ObligationVerdict{}}
	targeted := map[string]bool{}
	for _, t := range ci.Targets {
		k := t.RecordID + "@" + t.ArtifactSHA256
		r, ok := requirements[k]
		if !ok {
			return pbwarranty.Body{}, fmt.Errorf("warranty: requirement %s is not in the ledger", t.RecordID)
		}
		targeted[k] = true
		body.Requirements = append(body.Requirements, r)
	}
	for _, o := range order {
		if o[0] == 'v' {
			v := verdicts[o[1:]]
			if v.ReviewedCommit == sha && v.ChangeIntent.RecordID == ci.IntentID && v.ChangeIntent.ArtifactSHA256 == ci.ArtifactSHA256 {
				body.Verdicts = append(body.Verdicts, v)
			}
		} else {
			r := reviewsByPath[o[1:]]
			if targeted[r.Requirement.RecordID+"@"+r.Requirement.ArtifactSHA256] {
				body.Reviews = append(body.Reviews, r)
			}
		}
	}
	return body, nil
}

// warrantyExport writes a signed warranty for a commit whose readiness gate is PASS, after verifying
// it against this repository exactly as a recipient would; it never writes one that would not
// verify (W-INV-2).
func warrantyExport(ctx context.Context, root string, ledger *store.Store, arg, out string, output io.Writer) error {
	sha, err := resolveCommit(ctx, root, arg)
	if err != nil {
		return err
	}
	gateID := "ready-" + sha[:12]
	definitions, err := gates.LoadDir(filepath.Join(root, "gates"))
	if err != nil {
		return err
	}
	var gate *gates.Definition
	for i := range definitions {
		if definitions[i].ID == gateID && definitions[i].Rule == "intent-delivery-readiness" && definitions[i].ScopeCommit == sha {
			gate = &definitions[i]
		}
	}
	if gate == nil {
		return fmt.Errorf("warranty: no readiness gate gates/%s.yaml; review and seal the commit first", gateID)
	}
	result, err := gates.Evaluate(ctx, ledger, *gate)
	if err != nil {
		return err
	}
	if result.State != gates.StatePass {
		return fmt.Errorf("warranty: gate %s is %s (%s); only a PASS commit can be warranted", gateID, result.State, result.Reason)
	}
	body, err := warrantyBody(ctx, ledger, sha)
	if err != nil {
		return err
	}
	if body.ChangeName, err = claimedChange(ctx, root, sha); err != nil {
		return err
	}
	body.IssuedAt = time.Now().UTC().Format(time.RFC3339)
	body.Issuer = strings.TrimSpace(gitOutput(ctx, root, "config", "user.name"))
	body.RepositoryHint = strings.TrimSpace(gitOutput(ctx, root, "remote", "get-url", "origin"))
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return err
	}
	key, err := pbwarranty.LoadKey(keyPath)
	if err != nil {
		return err
	}
	data, err := pbwarranty.Sign(body, key)
	if err != nil {
		return err
	}
	parsed, err := pbwarranty.Parse(data)
	if err != nil {
		return err
	}
	// W-INV-2 amended: VALID or UNREVIEWED (a reviewer can still cosign it); never anything else.
	if err := pbwarranty.Verify(ctx, parsed, root, ""); errors.Is(err, pbwarranty.ErrUnreviewed) {
		if _, werr := fmt.Fprintf(output, "warning: %v; it will verify UNREVIEWED until the reviewer runs `proofbound warranty cosign`\n", err); werr != nil {
			return werr
		}
	} else if err != nil {
		return fmt.Errorf("warranty: would not verify against this repository: %v", err)
	}
	if out == "" {
		out = "warranty-" + sha[:12] + ".json"
	}
	f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("warranty: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(out)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	pub, _ := parsed.PublicKey()
	_, err = fmt.Fprintf(output, "warranty %s commit=%s change=sdd:%s key=%s obligations=%d\n", out, sha, body.ChangeName, pbwarranty.Fingerprint(pub), countTargets(body))
	return err
}

func countTargets(b pbwarranty.Body) int {
	n := 0
	for _, t := range b.ChangeIntent.Targets {
		n += len(t.ObligationIDs)
	}
	return n
}

// warrantyVerify prints exactly one verdict line and returns its exit code: 0 VALID, 2 UNVERIFIED
// (no --repo), 3 UNREVIEWED (a named reviewer has not signed), 1 INVALID (warranty SPEC §1).
func warrantyVerify(ctx context.Context, file, repo, trustKey string, output io.Writer) int {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(output, "INVALID: %v\n", err)
		return 1
	}
	f, err := pbwarranty.Parse(data)
	if err != nil {
		fmt.Fprintf(output, "INVALID: %v\n", err)
		return 1
	}
	fpr := "?"
	if pub, err := f.PublicKey(); err == nil {
		fpr = pbwarranty.Fingerprint(pub)
	}
	switch proven, err := pbwarranty.VerifyReport(ctx, f, repo, trustKey); {
	case err == nil:
		names := make([]string, 0, len(proven))
		for _, p := range proven {
			names = append(names, p.Name+":"+p.Fingerprint)
		}
		fmt.Fprintf(output, "VALID key=%s commit=%s change=sdd:%s obligations=%d reviewers=%s\n", fpr, f.Body.Commit, f.Body.ChangeName, countTargets(f.Body), strings.Join(names, ","))
		return 0
	case errors.Is(err, pbwarranty.ErrUnverified):
		fmt.Fprintf(output, "UNVERIFIED key=%s: %v\n", fpr, err)
		return 2
	case errors.Is(err, pbwarranty.ErrUnreviewed):
		fmt.Fprintf(output, "UNREVIEWED key=%s: %v\n", fpr, err)
		return 3
	default:
		fmt.Fprintf(output, "INVALID: %v\n", err)
		return 1
	}
}

// warrantyCosign adds the local key's cosignature to a warranty (option B, CLI-INV-27): the file must
// verify VALID or UNREVIEWED against this repository, the key must be the declared key of a reviewer
// the body names, and the result must still verify VALID or UNREVIEWED; it never writes a partial file.
func warrantyCosign(ctx context.Context, root, file, out string, output io.Writer) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cosign: %w", err)
	}
	f, err := pbwarranty.Parse(data)
	if err != nil {
		return fmt.Errorf("cosign: %w", err)
	}
	if err := pbwarranty.Verify(ctx, f, root, ""); err != nil && !errors.Is(err, pbwarranty.ErrUnreviewed) {
		return fmt.Errorf("cosign: the warranty does not verify against this repository: %v", err)
	}
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return err
	}
	key, err := pbwarranty.LoadKey(keyPath)
	if err != nil {
		return err
	}
	pub := key.Public().(ed25519.PublicKey)
	reviewer := ""
	for _, name := range bodyReviewers(f.Body) {
		declared, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(pbwarranty.KeysDir), pbwarranty.KeyName(name)+".pub"))
		if err == nil && string(declared) == pbwarranty.PubLine(pub) {
			reviewer = name
		}
	}
	if reviewer == "" {
		return fmt.Errorf("cosign: key %s is not the declared key of a reviewer of this warranty (%s)", pbwarranty.Fingerprint(pub), strings.Join(bodyReviewers(f.Body), ", "))
	}
	signed, err := pbwarranty.Cosign(data, key)
	if err != nil {
		return err
	}
	again, err := pbwarranty.Parse(signed)
	if err != nil {
		return err
	}
	result := "VALID"
	if err := pbwarranty.Verify(ctx, again, root, ""); errors.Is(err, pbwarranty.ErrUnreviewed) {
		result = "UNREVIEWED (another reviewer has not signed)"
	} else if err != nil {
		return fmt.Errorf("cosign: the cosigned warranty would not verify: %v", err)
	}
	if out == "" {
		out = file
		tmp, err := os.CreateTemp(filepath.Dir(file), ".warranty-cosign-*")
		if err != nil {
			return err
		}
		if _, err := tmp.Write(signed); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return err
		}
		if err := tmp.Close(); err != nil {
			os.Remove(tmp.Name())
			return err
		}
		if err := os.Rename(tmp.Name(), file); err != nil {
			os.Remove(tmp.Name())
			return err
		}
	} else if err := writeExclusive(out, signed); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "cosigned %s reviewer=%s key=%s now=%s\n", out, reviewer, pbwarranty.Fingerprint(pub), result)
	return err
}

func bodyReviewers(b pbwarranty.Body) []string {
	seen := map[string]bool{}
	var names []string
	for _, r := range b.Reviews {
		if !seen[r.DeclaredReviewer] {
			seen[r.DeclaredReviewer] = true
			names = append(names, r.DeclaredReviewer)
		}
	}
	for _, v := range b.Verdicts {
		if !seen[v.DeclaredReviewer] {
			seen[v.DeclaredReviewer] = true
			names = append(names, v.DeclaredReviewer)
		}
	}
	return names
}

func writeExclusive(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}
