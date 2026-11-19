package sdd

import (
	"context"
	"errors"
	"path"
	"sync"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
)

// Entry is one tree entry directly under a directory (SPEC §2).
type Entry struct{ Name, Mode, OID string }

// Artifact is one tree entry at a commit; Bytes is set only for regular files (SPEC §2).
type Artifact struct {
	Path    string
	Bytes   []byte
	Regular bool
	Mode    string
	OID     string
}

// Reader is the provider's only view of the tenant repository (SPEC §2).
type Reader interface {
	Commits(ctx context.Context) ([]string, error)
	Parents(ctx context.Context, commit string) ([]string, error)
	StateKey(ctx context.Context, commit string) (string, error)
	Read(ctx context.Context, commit string) ([]Artifact, error)
	List(ctx context.Context, commit, dir string) ([]Entry, error)
}

func regularMode(mode string) bool { return mode == "100644" || mode == "100755" }

// GitReader implements Reader over the git connector's own gitcmd.Repo, sharing its read
// contract (SPEC §3): the same history listing and refusals, typed literal tree lookups, and no
// per-commit file lists.
type GitReader struct {
	repo  *gitcmd.Repo
	mu    sync.Mutex
	blobs map[string][]byte
}

func NewGitReader(repo *gitcmd.Repo) (*GitReader, error) {
	if repo == nil {
		return nil, errors.New("sdd: repository is required")
	}
	return &GitReader{repo: repo, blobs: map[string][]byte{}}, nil
}

func (g *GitReader) Commits(ctx context.Context) ([]string, error) { return g.repo.CommitSHAs(ctx) }

func (g *GitReader) Parents(ctx context.Context, commit string) ([]string, error) {
	return g.repo.Parents(ctx, commit)
}

func (g *GitReader) blob(ctx context.Context, oid string) ([]byte, error) {
	g.mu.Lock()
	b, ok := g.blobs[oid]
	g.mu.Unlock()
	if ok {
		return b, nil
	}
	b, err := g.repo.Blob(ctx, oid)
	if err != nil {
		return nil, err
	}
	g.mu.Lock()
	g.blobs[oid] = b
	g.mu.Unlock()
	return b, nil
}

func (g *GitReader) own(ctx context.Context, commit, p string) (*gitcmd.TreeEntry, error) {
	e, err := g.repo.TreeEntries(ctx, commit, p, false, false)
	if err != nil || len(e) == 0 {
		return nil, err
	}
	return &e[0], nil
}

func entryLine(e *gitcmd.TreeEntry) string {
	if e == nil {
		return "-"
	}
	return e.Mode + " " + e.OID
}

// profileAt returns the profile entry and, when it is a regular file that parses, the profile.
func (g *GitReader) profileAt(ctx context.Context, commit string) (*gitcmd.TreeEntry, []byte, *profile, error) {
	pe, err := g.own(ctx, commit, ProfilePath)
	if err != nil || pe == nil || !regularMode(pe.Mode) {
		return pe, nil, nil, err
	}
	raw, err := g.blob(ctx, pe.OID)
	if err != nil {
		return nil, nil, nil, err
	}
	p, perr := parseProfile(raw)
	if perr != nil {
		return pe, raw, nil, nil
	}
	return pe, raw, &p, nil
}

// StateKey is SPEC §3's key: "<mode> <oid>" of the profile entry, then changes.root, then
// changes.archive ("-" when absent); "" without a profile entry. An unparseable profile's key is
// its own line alone, so roots are never read from it.
func (g *GitReader) StateKey(ctx context.Context, commit string) (string, error) {
	pe, _, p, err := g.profileAt(ctx, commit)
	if err != nil || pe == nil {
		return "", err
	}
	if p == nil {
		return entryLine(pe), nil
	}
	re, err := g.own(ctx, commit, p.Root)
	if err != nil {
		return "", err
	}
	var ae *gitcmd.TreeEntry // Spec Kit has no archive: its key line is "-"
	if p.Archive != "" {
		if ae, err = g.own(ctx, commit, p.Archive); err != nil {
			return "", err
		}
	}
	return entryLine(pe) + "\n" + entryLine(re) + "\n" + entryLine(ae), nil
}

func (g *GitReader) artifact(ctx context.Context, e gitcmd.TreeEntry) (Artifact, error) {
	a := Artifact{Path: e.Path, Mode: e.Mode, OID: e.OID, Regular: regularMode(e.Mode)}
	if a.Regular {
		b, err := g.blob(ctx, e.OID)
		if err != nil {
			return Artifact{}, err
		}
		a.Bytes = b
	}
	return a, nil
}

// Read returns the profile entry, and, when the profile parses, the root's and archive's own
// entries plus every entry under the root (trees included; bytes for regular files only).
func (g *GitReader) Read(ctx context.Context, commit string) ([]Artifact, error) {
	pe, raw, p, err := g.profileAt(ctx, commit)
	if err != nil || pe == nil {
		return nil, err
	}
	out := []Artifact{{Path: ProfilePath, Bytes: raw, Regular: regularMode(pe.Mode), Mode: pe.Mode, OID: pe.OID}}
	if p == nil {
		return out, nil
	}
	for _, own := range []string{p.Root, p.Archive} {
		if own == "" {
			continue // Spec Kit has no archive
		}
		e, err := g.own(ctx, commit, own)
		if err != nil {
			return nil, err
		}
		if e != nil {
			a, err := g.artifact(ctx, *e)
			if err != nil {
				return nil, err
			}
			out = append(out, a)
		}
	}
	under, err := g.repo.TreeEntries(ctx, commit, p.Root, true, true)
	if err != nil {
		return nil, err
	}
	for _, e := range under {
		if p.Archive != "" && e.Path == p.Archive {
			continue // already listed as the archive's own entry
		}
		a, err := g.artifact(ctx, e)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (g *GitReader) List(ctx context.Context, commit, dir string) ([]Entry, error) {
	under, err := g.repo.TreeEntries(ctx, commit, dir, true, false)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(under))
	for _, e := range under {
		out = append(out, Entry{Name: path.Base(e.Path), Mode: e.Mode, OID: e.OID})
	}
	return out, nil
}

var _ Reader = (*GitReader)(nil)

func isChangeEntryMode(mode string) bool {
	return mode == "040000" || mode == "120000" || mode == "160000"
}
