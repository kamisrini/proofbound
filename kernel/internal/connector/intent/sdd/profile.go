package sdd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ProfilePath is where a tenant commits its profile (SPEC §2.1).
const ProfilePath = ".proofbound/sdd-profile.json"

// profile is the parsed proofbound.sdd-profile.v1 (SPEC §2.1). M1 accepts only kit "openspec".
type profile struct {
	Root, Archive, ChangeStatus string
	Owners, Sponsors            []rule
}

type rule struct{ Pattern, Value string }

var segmentRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// literalPath reports whether p follows SPEC §2.1's literal path grammar.
func literalPath(p string) bool {
	if p == "" {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if !segmentRE.MatchString(seg) || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

// patternPath reports whether p is an owners/sponsors pattern: literal segments, `*`, or `**`.
func patternPath(p string) bool {
	if p == "" {
		return false
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "*" || seg == "**" {
			continue
		}
		if !segmentRE.MatchString(seg) || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

// matchPattern matches a path against a pattern whose `*` spans one segment and `**` any number.
func matchPattern(pattern, path string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(path, "/"))
}

func matchSegments(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(segs); i++ {
			if matchSegments(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	if len(segs) == 0 {
		return false
	}
	if pat[0] != "*" && pat[0] != segs[0] {
		return false
	}
	return matchSegments(pat[1:], segs[1:])
}

func firstMatch(rules []rule, path string) (string, bool) {
	for _, r := range rules {
		if matchPattern(r.Pattern, path) {
			return r.Value, true
		}
	}
	return "", false
}

var errProfile = errors.New("profile-invalid")

// parseProfile applies every SPEC §2.1 rule; any violation is errProfile (cause profile-invalid).
func parseProfile(raw []byte) (profile, error) {
	if !utf8.Valid(raw) || bytes.IndexByte(raw, 0) >= 0 {
		return profile{}, errProfile
	}
	if err := strictShape(raw); err != nil {
		return profile{}, errProfile
	}
	var doc struct {
		Schema       *string `json:"schema"`
		Kit          *string `json:"kit"`
		Requirements *struct {
			Paths []string `json:"paths"`
		} `json:"requirements"`
		Changes *struct {
			Mode    *string `json:"mode"`
			Root    *string `json:"root"`
			Archive *string `json:"archive"`
			Status  *string `json:"status"`
		} `json:"changes"`
		Owners []struct {
			Pattern *string `json:"pattern"`
			Owner   *string `json:"owner"`
		} `json:"owners"`
		Sponsors []struct {
			Pattern *string `json:"pattern"`
			Sponsor *string `json:"sponsor"`
		} `json:"sponsors"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil || dec.More() {
		return profile{}, errProfile
	}
	if doc.Schema == nil || *doc.Schema != "proofbound.sdd-profile.v1" || doc.Kit == nil || *doc.Kit != "openspec" ||
		doc.Requirements == nil || doc.Changes == nil || doc.Changes.Mode == nil || *doc.Changes.Mode != "delta-sections" ||
		doc.Changes.Root == nil || doc.Changes.Archive == nil || doc.Changes.Status == nil ||
		(*doc.Changes.Status != "proposed" && *doc.Changes.Status != "accepted") ||
		len(doc.Owners) == 0 || len(doc.Sponsors) == 0 {
		return profile{}, errProfile
	}
	root, archive := *doc.Changes.Root, *doc.Changes.Archive
	if !literalPath(root) || !literalPath(archive) || !strings.HasPrefix(archive, root+"/") || strings.Contains(archive[len(root)+1:], "/") {
		return profile{}, errProfile
	}
	want := []string{root + "/*/specs/*/spec.md", archive + "/*/specs/*/spec.md"}
	if len(doc.Requirements.Paths) != 2 || doc.Requirements.Paths[0] != want[0] || doc.Requirements.Paths[1] != want[1] {
		return profile{}, errProfile
	}
	p := profile{Root: root, Archive: archive, ChangeStatus: *doc.Changes.Status}
	for _, o := range doc.Owners {
		if o.Pattern == nil || o.Owner == nil || !patternPath(*o.Pattern) || blank(*o.Owner) {
			return profile{}, errProfile
		}
		p.Owners = append(p.Owners, rule{*o.Pattern, *o.Owner})
	}
	for _, o := range doc.Sponsors {
		if o.Pattern == nil || o.Sponsor == nil || !patternPath(*o.Pattern) || blank(*o.Sponsor) {
			return profile{}, errProfile
		}
		p.Sponsors = append(p.Sponsors, rule{*o.Pattern, *o.Sponsor})
	}
	return p, nil
}

// strictShape rejects duplicate object keys and null values anywhere in the document, which
// encoding/json would otherwise accept silently.
func strictShape(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var walk func() error
	walk = func() error {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case nil:
			return errors.New("null")
		case json.Delim:
			switch t {
			case '{':
				seen := map[string]bool{}
				for dec.More() {
					k, err := dec.Token()
					if err != nil {
						return err
					}
					key, ok := k.(string)
					if !ok || seen[key] {
						return fmt.Errorf("duplicate key %v", k)
					}
					seen[key] = true
					if err := walk(); err != nil {
						return err
					}
				}
				_, err = dec.Token()
				return err
			case '[':
				for dec.More() {
					if err := walk(); err != nil {
						return err
					}
				}
				_, err = dec.Token()
				return err
			}
		}
		return nil
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := dec.Token(); err == nil {
		return errors.New("trailing data")
	}
	return nil
}
