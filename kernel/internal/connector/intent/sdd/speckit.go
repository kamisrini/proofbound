package sdd

import (
	"bytes"
	"path"
	"regexp"
	"unicode/utf8"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
)

// Spec Kit mapping (docs/plans/R2-speckit-spec.md §3, SDD-INV-69..71). A feature is a directory
// directly under the profile's root; its spec.md is one requirement record; the feature has a
// change intent once its sibling (tasks.md) exists. There is no archive.

var specKitObligationRE = regexp.MustCompile(`^- \*\*((?:FR|SC|NFR)-[0-9]+)\*\*:? (.+)$`)

// parseSpecKit reads the obligation lines of a feature spec outside code fences. unit is false
// when the spec has no obligation line (it is then not a record unit); cause is a §7 cause or "".
func parseSpecKit(text string) (obls []intent.Obligation, cause string, unit bool) {
	lines := splitLines(text)
	mask := fenceMask(lines)
	seen := map[string]bool{}
	for i, line := range lines {
		if mask[i] {
			continue
		}
		m := specKitObligationRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		unit = true
		statement := jsTrim(m[2])
		id := "O-" + m[1]
		switch {
		case blank(statement):
			cause = "statement-blank"
		case seen[id]:
			cause = "obligation-collision"
		}
		if cause != "" {
			return nil, cause, true
		}
		seen[id] = true
		obls = append(obls, intent.Obligation{ID: id, Statement: statement, State: "active"})
	}
	return obls, "", unit
}

func (p *Provider) mapSpecKit(st *stateResult, arts []Artifact, prof Artifact, profLine string, pr profile) (*stateResult, error) {
	byPath := map[string]Artifact{}
	children := map[string][]Artifact{}
	for _, a := range arts[1:] {
		byPath[a.Path] = a
		children[path.Dir(a.Path)] = append(children[path.Dir(a.Path)], a)
	}
	rootEntry, ok := byPath[pr.Root]
	if !ok {
		return st, nil
	}
	if rootEntry.Mode != "040000" {
		st.stateSkip = "non-regular-file"
		st.skips = append(st.skips, unitSkip{"profile", "non-regular-file", digestLines([]string{profLine, lineOf(rootEntry)})})
		return st, nil
	}
	under := func(dir string) []Artifact {
		var out []Artifact
		for _, a := range arts[1:] {
			if len(a.Path) > len(dir) && a.Path[:len(dir)+1] == dir+"/" {
				out = append(out, a)
			}
		}
		return out
	}
	unitDigest := func(e Artifact) string {
		lines := []string{profLine, lineOf(e)}
		for _, a := range under(e.Path) {
			lines = append(lines, lineOf(a))
		}
		return digestLines(lines)
	}
	for _, fe := range children[pr.Root] {
		name := path.Base(fe.Path)
		if !isChangeEntryMode(fe.Mode) || !inDomain(name) {
			continue // regular files under the root are not features
		}
		cr := &changeResult{}
		st.changes[name] = cr
		fail := func(cause string) {
			cr.skip, cr.rev = cause, nil
			st.skips = append(st.skips, unitSkip{"change " + name, cause, unitDigest(fe)})
		}
		if !utf8.ValidString(name) {
			fail("invalid-utf8")
			continue
		}
		if fe.Mode != "040000" {
			fail("non-regular-file")
			continue
		}
		if _, err := slug(name); err != nil {
			fail("slug-too-short")
			continue
		}
		var rec *record
		if spec, ok := byPath[fe.Path+"/spec.md"]; ok {
			r := &record{change: name, key: name, unitPath: spec.Path, posLine: lineOf(spec)}
			switch {
			case !spec.Regular:
				r.skip, rec = "non-regular-file", r
			case !utf8.ValidString(spec.Path) || !utf8.Valid(spec.Bytes):
				r.skip, rec = "invalid-utf8", r
			case bytes.IndexByte(spec.Bytes, 0) >= 0:
				r.skip, rec = "nul-byte", r
			default:
				obls, cause, unit := parseSpecKit(string(spec.Bytes))
				if unit {
					rec = r
					r.skip = cause
					if cause == "" {
						if err := p.buildRecord(name, spec, prof, pr, obls, true, &r.skip, &r.rev, &r.obligations); err != nil {
							return nil, err
						}
					}
				}
			}
		}
		if rec != nil {
			if rec.skip != "" {
				st.skips = append(st.skips, unitSkip{"requirement " + rec.unitPath, rec.skip, digestLines([]string{profLine, rec.posLine})})
			} else {
				st.revisions = append(st.revisions, *rec.rev)
			}
		}
		sibling, planned := byPath[fe.Path+"/"+pr.Sibling]
		switch {
		case !planned || rec == nil:
			cr.absent = true // no sibling (not planned) or no requirement: no change intent
			continue
		case !sibling.Regular:
			fail("non-regular-file")
			continue
		case rec.skip != "":
			fail("cascade")
			continue
		}
		rev, cause, err := p.buildChange(name, fe, true, under(fe.Path), prof, pr, []*record{rec})
		if err != nil {
			return nil, err
		}
		if cause != "" {
			fail(cause)
			continue
		}
		cr.rev = rev
		st.revisions = append(st.revisions, *rev)
	}
	return finishState(st), nil
}
