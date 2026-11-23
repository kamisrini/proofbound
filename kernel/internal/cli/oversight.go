package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/store"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// Agent oversight (U1 spec section 11, CLI-INV-42..44): who reviewed each change, which commits are
// signed by a declared key, and the agent sessions in the period.

// Commit signature states.
const (
	sigVerified = "verified"  // a good signature by a key the repository declares at HEAD
	sigUnsigned = "unsigned"  // no signature
	sigBad      = "bad"       // a signature that does not verify
	sigOther    = "unchecked" // signed, but not by a declared key, or not checkable here
)

// sshWireKey is the OpenSSH public-key line for an ed25519 key: what git's allowed-signers file needs.
func sshWireKey(pub ed25519.PublicKey) string {
	var b bytes.Buffer
	for _, part := range [][]byte{[]byte("ssh-ed25519"), pub} {
		var n [4]byte
		binary.BigEndian.PutUint32(n[:], uint32(len(part)))
		b.Write(n[:])
		b.Write(part)
	}
	return "ssh-ed25519 " + base64.StdEncoding.EncodeToString(b.Bytes())
}

// commitSignatures checks every commit's git signature against the keys HEAD declares, using git's
// own SSH verification with an allowed-signers file built from those keys (one principal per key
// name). A GPG signature, or one by an undeclared key, is "unchecked"; without ssh-keygen nothing can
// be checked and every signed commit is "unchecked".
func commitSignatures(ctx context.Context, root, base string) (map[string][2]string, error) {
	keys := declaredKeys(ctx, root, "HEAD")
	dir, err := os.MkdirTemp("", "proofbound-signers-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	var lines []string
	for name, pub := range keys {
		lines = append(lines, name+` namespaces="git" `+sshWireKey(pub))
	}
	sort.Strings(lines)
	signers := filepath.Join(dir, "allowed_signers")
	if err := os.WriteFile(signers, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return nil, err
	}
	// gpg.program points nowhere, so a GPG signature can never read as a declared key (F5).
	cmd := exec.CommandContext(ctx, "git", "-C", root, "-c", "gpg.ssh.allowedSignersFile="+signers, "-c", "gpg.program=proofbound-no-gpg", "log", "--format=%H%x00%G?%x00%GS%x1e", base+"..HEAD", "--")
	cmd.Env = repositoryGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("audit: cannot read commit signatures: %v", err)
	}
	sigs := map[string][2]string{}
	for _, rec := range strings.Split(string(out), "\x1e") {
		f := strings.Split(strings.TrimLeft(rec, "\n"), "\x00")
		if len(f) != 3 {
			continue
		}
		state, signer := sigOther, ""
		switch f[1] {
		case "N":
			state = sigUnsigned
		case "B":
			state = sigBad
		case "G":
			if _, ok := keys[f[2]]; ok {
				state, signer = sigVerified, f[2]
			}
		}
		sigs[f[0]] = [2]string{state, signer}
	}
	return sigs, nil
}

// auditRole is who wrote, reviewed and owns one claiming commit (C8).
type auditRole struct {
	Commit      string   `json:"commit"`
	Author      string   `json:"author"`
	Agents      []string `json:"agents"`
	Signer      string   `json:"signer,omitempty"`
	Reviewers   []string `json:"reviewers"`
	Owners      []string `json:"owners"`
	Independent bool     `json:"independent"`
	Note        string   `json:"note"`
	// Proofs are the review files' signatures for the browser to check (page only).
	Proofs []browserProof `json:"-"`
}

// roleFor reads the reviewers and owners of a claiming commit from the record (the same body a
// warranty carries, without signing anything) and checks that no reviewer is the commit's author, one
// of its agents, its verified signer or a spec owner, by key name (W-INV-3's identity rule).
func roleFor(ctx context.Context, ledger *store.Store, c auditCommit) auditRole {
	r := auditRole{Commit: c.Commit, Author: c.Author, Agents: c.Agents, Signer: c.Signer, Reviewers: []string{}, Owners: []string{}}
	body, err := warrantyBody(ctx, ledger, c.Commit)
	if err != nil {
		r.Note = "no review in the record"
		return r
	}
	seen := map[string]bool{}
	for _, q := range body.Requirements {
		if o := printableUnit(q.DeclaredOwner); o != "" && !seen[o] {
			seen[o] = true
			r.Owners = append(r.Owners, o)
		}
	}
	for _, name := range bodyReviewers(body) {
		r.Reviewers = append(r.Reviewers, printableUnit(name))
	}
	if len(r.Reviewers) == 0 {
		r.Note = "no review in the record"
		return r
	}
	r.Independent, r.Note = independence(c, r.Reviewers, r.Owners)
	return r
}

// independence applies C8's rule: no reviewer may be, by key name, the commit's author, an agent
// named on it, its verified signer or a spec owner.
func independence(c auditCommit, reviewers, owners []string) (bool, string) {
	wrote := map[string]string{pbwarranty.KeyName(c.Author): "its author"}
	for _, a := range c.Agents {
		wrote[pbwarranty.KeyName(a)] = "its agent"
	}
	if c.Signer != "" {
		wrote[c.Signer] = "its signer"
	}
	for _, o := range owners {
		wrote[pbwarranty.KeyName(o)] = "a spec owner"
	}
	for _, rv := range reviewers {
		if k := pbwarranty.KeyName(rv); k != "" {
			if role, ok := wrote[k]; ok {
				return false, fmt.Sprintf("reviewer %s is %s", rv, role)
			}
		}
	}
	return true, "reviewed by someone who neither wrote nor owns it"
}

// auditSession is one observed agent session in the period and the commits made while it ran.
type auditSession struct {
	ID        string   `json:"id"`
	Started   string   `json:"started"`
	Finished  string   `json:"finished"`
	ToolCalls int      `json:"tool_calls"`
	Files     int      `json:"files_written"`
	Commits   []string `json:"commits"`
}

// sessionsInPeriod reads the `session.observed` events (Claude Code session metadata the sessions
// connector records: times and counts, never content) whose session ran after the period began, and
// links each commit whose committer time falls inside a session, with five minutes' grace.
func sessionsInPeriod(ctx context.Context, root string, ledger *store.Store, base string, commits []auditCommit) ([]auditSession, error) {
	since, err := time.Parse(time.RFC3339, strings.TrimSpace(gitOutput(ctx, root, "log", "-1", "--format=%cI", base)))
	if err != nil {
		since = time.Time{}
	}
	committed := map[string]time.Time{}
	for _, c := range commits {
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(gitOutput(ctx, root, "log", "-1", "--format=%cI", c.Commit))); err == nil {
			committed[c.Commit] = t
		}
	}
	// A session observed again (it grew, or a parser fix recounted it) is a new event; the latest wins.
	latest := map[string]auditSession{}
	var order []string
	err = ledger.ReadEvents(ctx, store.Filter{Source: core.SourceSessions, Kind: core.KindSessionObserved}, func(record store.Record) error {
		var p struct {
			SessionID     string  `json:"session_id"`
			StartedAt     *string `json:"started_at"`
			FinishedAt    *string `json:"finished_at"`
			ToolCallCount int     `json:"tool_call_count"`
			FilesWritten  int     `json:"files_written_count"`
		}
		if err := json.Unmarshal(record.Event.Payload, &p); err != nil || p.StartedAt == nil || p.FinishedAt == nil {
			return nil
		}
		start, err1 := time.Parse(time.RFC3339, *p.StartedAt)
		end, err2 := time.Parse(time.RFC3339, *p.FinishedAt)
		if err1 != nil || err2 != nil || end.Before(since) {
			return nil
		}
		s := auditSession{ID: printableUnit(p.SessionID), Started: start.UTC().Format("2006-01-02 15:04"), Finished: end.UTC().Format("15:04"), ToolCalls: p.ToolCallCount, Files: p.FilesWritten, Commits: []string{}}
		for _, c := range commits {
			if t, ok := committed[c.Commit]; ok && !t.Before(start) && !t.After(end.Add(5*time.Minute)) {
				s.Commits = append(s.Commits, c.Commit)
			}
		}
		if _, seen := latest[s.ID]; !seen {
			order = append(order, s.ID)
		}
		latest[s.ID] = s
		return nil
	})
	out := make([]auditSession, 0, len(order))
	for _, id := range order {
		s := latest[id]
		s.ID = short(s.ID, 8)
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Started < out[j].Started })
	return out, err
}
