package sdd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// Resolve's tenant outcomes (SPEC §6). Unmappable outcomes are *UnmappableError, which wraps
// intent.ErrUnmappable.
var (
	ErrUnknownChange   = errors.New("sdd: no such change at this commit")
	ErrArchivedEarlier = errors.New("sdd: change was archived before this commit")
	ErrNoChangeIntent  = errors.New("sdd: change exists at this commit but has no intent")
)

// UnmappableError carries one of SPEC §7's closed causes.
type UnmappableError struct{ Cause string }

func (e *UnmappableError) Error() string { return "sdd: unmappable: " + e.Cause }
func (e *UnmappableError) Unwrap() error { return intent.ErrUnmappable }

func unmappable(cause string) error { return &UnmappableError{Cause: cause} }

// Provider is the intent.sdd provider (SPEC §1-§7).
type Provider struct {
	reader     Reader
	observedAt time.Time
	skipped    []intent.Skip
}

func New(reader Reader, observedAt time.Time) (*Provider, error) {
	if reader == nil || observedAt.IsZero() {
		return nil, errors.New("sdd: reader and observation time are required")
	}
	return &Provider{reader: reader, observedAt: observedAt.UTC()}, nil
}

func (p *Provider) Name() string { return "sdd" }

// Skipped returns the Skips of the last Revisions call (SPEC §7).
func (p *Provider) Skipped() []intent.Skip { return append([]intent.Skip(nil), p.skipped...) }

var datedRE = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$`)

type unitSkip struct{ unit, cause, digest string }

// record is one record unit (SPEC §4.1): emitted (rev) or skipped (skip).
type record struct {
	change, key, unitPath string
	skip                  string
	rev                   *intent.Revision
	obligations           []string
	posLine               string
}

type changeResult struct {
	skip          string // a §7 cause, or ""
	absent        bool   // §5 rule 5
	singleArchive bool   // exactly one change entry, and it is an archive entry
	rev           *intent.Revision
}

type stateResult struct {
	stateSkip   string
	archiveBase string
	archive     string
	revisions   []intent.Revision
	skips       []unitSkip
	changes     map[string]*changeResult
}

func lineOf(a Artifact) string { return a.Mode + " " + a.OID + " " + a.Path }

func digestLines(lines []string) string {
	sort.Strings(lines)
	h := sha256.New()
	for _, l := range lines {
		h.Write([]byte(l + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func contentLine(a Artifact) string {
	if a.Regular {
		sum := sha256.Sum256(a.Bytes)
		return hex.EncodeToString(sum[:]) + "  " + a.Path
	}
	return a.Mode + " " + a.OID + "  " + a.Path
}

// manifest is SPEC §4.3's manifest digest over the given entries (trees are never lines).
func manifest(arts []Artifact) string {
	var lines []string
	for _, a := range arts {
		if a.Mode == "040000" {
			continue
		}
		lines = append(lines, contentLine(a))
	}
	sort.Slice(lines, func(i, j int) bool {
		return strings.SplitN(lines[i], "  ", 2)[1] < strings.SplitN(lines[j], "  ", 2)[1]
	})
	h := sha256.New()
	for _, l := range lines {
		h.Write([]byte(l + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalPayload(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return core.Canonicalize(raw)
}

func (p *Provider) build(kind core.Kind, id, artifactPath, digest string, payload []byte) intent.Revision {
	return intent.Revision{Source: core.SourceIntentSDD, Kind: kind, NativeID: id, ArtifactPath: artifactPath, ArtifactSHA256: digest, ObservedAt: p.observedAt, Payload: payload}
}

// inDomain is SPEC §6's change-name domain (the archive's own name is excluded by the caller).
func inDomain(n string) bool {
	return n != "" && n != "." && n != ".." && !strings.Contains(n, "/") && !strings.Contains(n, "\x00")
}

// mapState maps one SDD state (SPEC §3-§7). An error is a provider defect, never tenant content.
func (p *Provider) mapState(arts []Artifact) (*stateResult, error) {
	st := &stateResult{changes: map[string]*changeResult{}}
	if len(arts) == 0 {
		return st, nil
	}
	prof := arts[0]
	profLine := lineOf(prof)
	if !prof.Regular {
		st.stateSkip = "profile-invalid"
		st.skips = append(st.skips, unitSkip{"profile", "profile-invalid", digestLines([]string{profLine})})
		return st, nil
	}
	pr, err := parseProfile(prof.Bytes)
	if err != nil {
		st.stateSkip = "profile-invalid"
		st.skips = append(st.skips, unitSkip{"profile", "profile-invalid", digestLines([]string{profLine})})
		return st, nil
	}
	if pr.Kit == KitSpecKit {
		return p.mapSpecKit(st, arts, prof, profLine, pr)
	}
	st.archive, st.archiveBase = pr.Archive, path.Base(pr.Archive)
	byPath := map[string]Artifact{}
	children := map[string][]Artifact{}
	for _, a := range arts[1:] {
		byPath[a.Path] = a
		children[path.Dir(a.Path)] = append(children[path.Dir(a.Path)], a)
	}
	for _, own := range []string{pr.Root, pr.Archive} {
		if a, ok := byPath[own]; ok && a.Mode != "040000" {
			st.stateSkip = "non-regular-file"
			st.skips = append(st.skips, unitSkip{"profile", "non-regular-file", digestLines([]string{profLine, lineOf(a)})})
			return st, nil
		}
	}
	if _, ok := byPath[pr.Root]; !ok {
		return st, nil
	}
	under := func(dir string) []Artifact {
		var out []Artifact
		for _, a := range arts[1:] {
			if strings.HasPrefix(a.Path, dir+"/") {
				out = append(out, a)
			}
		}
		return out
	}
	unitDigest := func(entries []Artifact) string {
		lines := []string{profLine}
		for _, e := range entries {
			lines = append(lines, lineOf(e))
			for _, a := range under(e.Path) {
				lines = append(lines, lineOf(a))
			}
		}
		return digestLines(lines)
	}
	// Change entries (SPEC §5): trees, symlinks, gitlinks only; regular files are ignored.
	type entries struct {
		active   *Artifact
		archived []Artifact
	}
	names := map[string]*entries{}
	get := func(n string) *entries {
		if names[n] == nil {
			names[n] = &entries{}
		}
		return names[n]
	}
	for _, a := range children[pr.Root] {
		a := a
		base := path.Base(a.Path)
		if a.Path == pr.Archive || !isChangeEntryMode(a.Mode) || !inDomain(base) {
			continue
		}
		get(base).active = &a
	}
	for _, a := range children[pr.Archive] {
		if !isChangeEntryMode(a.Mode) {
			continue
		}
		m := datedRE.FindStringSubmatch(path.Base(a.Path))
		if m == nil || !inDomain(m[1]) || m[1] == st.archiveBase {
			st.skips = append(st.skips, unitSkip{"archive " + path.Base(a.Path), "archive-name-invalid", unitDigest([]Artifact{a})})
			continue
		}
		get(m[1]).archived = append(get(m[1]).archived, a)
	}
	var records []*record
	recordsOf := map[string][]*record{}
	addRecord := func(n string, r *record) {
		records = append(records, r)
		recordsOf[n] = append(recordsOf[n], r)
	}
	for n, es := range names {
		all := es.archived
		if es.active != nil {
			all = append([]Artifact{*es.active}, all...)
		}
		for _, ce := range all {
			if ce.Mode != "040000" {
				continue
			}
			specs, ok := byPath[ce.Path+"/specs"]
			if !ok {
				continue
			}
			if specs.Mode != "040000" {
				if !specs.Regular { // a regular file named specs is ignored, as OpenSpec ignores it
					addRecord(n, &record{change: n, unitPath: specs.Path, skip: "non-regular-file", posLine: lineOf(specs)})
				}
				continue
			}
			for _, capDir := range children[specs.Path] {
				if capDir.Mode != "040000" {
					// A regular file directly under specs/ (e.g. a "no spec changes" note) is ignored,
					// as OpenSpec's reader ignores it; only a symlink or gitlink is a cause.
					if !capDir.Regular {
						addRecord(n, &record{change: n, unitPath: capDir.Path, skip: "non-regular-file", posLine: lineOf(capDir)})
					}
					continue
				}
				spec, ok := byPath[capDir.Path+"/spec.md"]
				if !ok {
					continue
				}
				if !spec.Regular {
					addRecord(n, &record{change: n, unitPath: spec.Path, skip: "non-regular-file", posLine: lineOf(spec)})
					continue
				}
				r := &record{change: n, key: n + "/" + path.Base(capDir.Path), unitPath: spec.Path, posLine: lineOf(spec)}
				if !utf8.ValidString(spec.Path) || !utf8.Valid(spec.Bytes) {
					r.skip = "invalid-utf8"
					addRecord(n, r)
					continue
				}
				if bytes.IndexByte(spec.Bytes, 0) >= 0 {
					r.skip = "nul-byte"
					addRecord(n, r)
					continue
				}
				found, unpaired := parseDelta(string(spec.Bytes))
				if len(found) == 0 && unpaired == 0 {
					continue // not a record unit (SPEC §4.1)
				}
				addRecord(n, r)
				if unpaired > 0 {
					r.skip = "unpaired-rename"
					continue
				}
				obls, cause := openspecObligations(found)
				if cause != "" {
					r.skip = cause
					continue
				}
				if err := p.buildRecord(r.key, spec, prof, pr, obls, es.active != nil && ce.Path == es.active.Path, &r.skip, &r.rev, &r.obligations); err != nil {
					return nil, err
				}
			}
		}
	}
	// Record-id collisions within this state (SPEC §4.3).
	byID := map[string][]*record{}
	for _, r := range records {
		if r.rev != nil {
			byID[r.rev.NativeID] = append(byID[r.rev.NativeID], r)
		}
	}
	for _, group := range byID {
		distinct := map[string]bool{}
		for _, r := range group {
			distinct[r.key+"\x00"+r.rev.ArtifactSHA256] = true
		}
		if len(distinct) > 1 {
			for _, r := range group {
				r.skip, r.rev = "record-id-collision", nil
			}
		}
	}
	for _, r := range records {
		if r.skip != "" {
			st.skips = append(st.skips, unitSkip{"requirement " + r.unitPath, r.skip, digestLines([]string{profLine, r.posLine})})
		} else {
			st.revisions = append(st.revisions, *r.rev)
		}
	}
	// changeOutcome for every in-domain name with a change entry (SPEC §5).
	for n, es := range names {
		all := es.archived
		if es.active != nil {
			all = append([]Artifact{*es.active}, all...)
		}
		cr := &changeResult{singleArchive: es.active == nil && len(es.archived) == 1}
		st.changes[n] = cr
		cause := ""
		switch {
		case es.active != nil && len(es.archived) > 0 || len(es.archived) > 1:
			cause = "name-reuse"
		case !utf8.ValidString(n):
			cause = "invalid-utf8"
		}
		if cause == "" {
			if _, err := slug(n); err != nil {
				cause = "slug-too-short"
			}
		}
		if cause == "" {
			for _, e := range all {
				if e.Mode != "040000" {
					cause = "non-regular-file"
				}
			}
		}
		if cause == "" && len(recordsOf[n]) == 0 {
			cr.absent = true
			continue
		}
		if cause == "" {
			for _, r := range recordsOf[n] {
				if r.skip != "" {
					cause = "cascade"
				}
			}
		}
		if cause == "" {
			rev, c, err := p.buildChange(n, all[0], es.active != nil, under(all[0].Path), prof, pr, recordsOf[n])
			if err != nil {
				return nil, err
			}
			cause = c
			cr.rev = rev
		}
		if cause != "" {
			cr.skip, cr.rev = cause, nil
			st.skips = append(st.skips, unitSkip{"change " + n, cause, unitDigest(all)})
			continue
		}
		st.revisions = append(st.revisions, *cr.rev)
	}
	return finishState(st), nil
}

// finishState orders a state's skips and revisions deterministically.
func finishState(st *stateResult) *stateResult {
	sort.Slice(st.skips, func(i, j int) bool {
		a, b := st.skips[i], st.skips[j]
		return a.unit+"\x00"+a.cause+"\x00"+a.digest < b.unit+"\x00"+b.cause+"\x00"+b.digest
	})
	sort.Slice(st.revisions, func(i, j int) bool {
		a, b := st.revisions[i], st.revisions[j]
		return string(a.Kind)+"\x00"+a.NativeID+"\x00"+a.ArtifactSHA256 < string(b.Kind)+"\x00"+b.NativeID+"\x00"+b.ArtifactSHA256
	})
	return st
}
