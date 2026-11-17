package sdd

import (
	"regexp"
	"strings"
	"unicode"
)

// SPEC §4.2: OpenSpec's delta reader (parseDeltaSpec at f085c850) with ECMAScript regex semantics
// made explicit. WS is JavaScript's \s; DOT is JavaScript's `.`; the only case-insensitive match
// (the requirement header's "Requirement:") folds ASCII letters only.
const (
	ws  = `[\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`
	dot = `[^\n\r\x{2028}\x{2029}]`
	req = `[Rr][Ee][Qq][Uu][Ii][Rr][Ee][Mm][Ee][Nn][Tt]:`
)

var (
	sectionLineRE   = regexp.MustCompile(`^##` + ws + `+(` + dot + `+)$`)
	topRE           = regexp.MustCompile(`^##` + ws + `+`)
	reqRE           = regexp.MustCompile(`^###` + ws + `*` + req + ws + `*(` + dot + `+)` + ws + `*$`)
	removedBulletRE = regexp.MustCompile("^" + ws + "*[-*+]" + ws + "*`?###" + ws + "*Requirement:" + ws + "*(" + dot + "+?)`?" + ws + "*$")
	fromRE          = regexp.MustCompile("^" + ws + "*[-*+]?" + ws + "*FROM:" + ws + "*`?###" + ws + "*Requirement:" + ws + "*(" + dot + "+?)`?" + ws + "*$")
	toRE            = regexp.MustCompile("^" + ws + "*[-*+]?" + ws + "*TO:" + ws + "*`?###" + ws + "*Requirement:" + ws + "*(" + dot + "+?)`?" + ws + "*$")
	openRE          = regexp.MustCompile("^" + ws + "*(`{3,}|~{3,})(" + dot + "*)")
	closeRE         = regexp.MustCompile("^" + ws + "*(`{3,}|~{3,})" + ws + "*$")
	closingSeqRE    = regexp.MustCompile(`[ \t]+#+[ \t]*$`)
)

var deltaTitles = map[string]string{
	"added requirements":    "ADD",
	"modified requirements": "MOD",
	"removed requirements":  "REM",
	"renamed requirements":  "REN",
}

// entry is one obligation-bearing entry of a delta file: its section, name, and statement.
type entry struct{ Sec, Name, Statement string }

func jsSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

// blank is §4.2's blank: every code point is JavaScript whitespace or Go unicode.IsSpace.
func blank(s string) bool {
	for _, r := range s {
		if !jsSpace(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func jsTrim(s string) string { return strings.TrimFunc(s, jsSpace) }

func normalizeName(raw string) string { return jsTrim(closingSeqRE.ReplaceAllString(raw, "")) }

func splitLines(text string) []string {
	text = strings.TrimPrefix(text, string(rune(0xFEFF)))
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

// fenceMask is OpenSpec's buildCodeFenceMask (code-fence.ts).
func fenceMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	var marker byte
	length := 0
	for i, l := range lines {
		if marker == 0 {
			if m := openRE.FindStringSubmatch(l); m != nil && (m[1][0] != '`' || !strings.Contains(m[2], "`")) {
				marker, length, mask[i] = m[1][0], len(m[1]), true
			}
			continue
		}
		mask[i] = true
		if m := closeRE.FindStringSubmatch(l); m != nil && m[1][0] == marker && len(m[1]) >= length {
			marker = 0
		}
	}
	return mask
}

func titleKind(raw string) string {
	t := normalizeName(raw)
	for i := 0; i < len(t); i++ {
		if t[i] >= 0x80 {
			return ""
		}
	}
	return deltaTitles[strings.ToLower(t)]
}

// parseDelta returns a delta file's entries and the number of unpaired RENAMED lines (SPEC §4.2).
func parseDelta(text string) ([]entry, int) {
	lines := splitLines(text)
	mask := fenceMask(lines)
	type head struct {
		at   int
		kind string
	}
	var heads []head
	for i, l := range lines {
		if mask[i] {
			continue
		}
		if m := sectionLineRE.FindStringSubmatch(l); m != nil {
			heads = append(heads, head{i, titleKind(m[1])})
		}
	}
	var out []entry
	unpaired := 0
	for n, h := range heads {
		if h.kind == "" {
			continue
		}
		end := len(lines)
		if n+1 < len(heads) {
			end = heads[n+1].at
		}
		var body []string
		for i := h.at + 1; i < end; i++ {
			if !mask[i] {
				body = append(body, lines[i])
			}
		}
		if h.kind == "REN" {
			pending, has := "", false
			for _, l := range body {
				if m := fromRE.FindStringSubmatch(l); m != nil {
					if has {
						unpaired++
					}
					pending, has = normalizeName(m[1]), true
				} else if m := toRE.FindStringSubmatch(l); m != nil {
					if !has {
						unpaired++
						continue
					}
					out = append(out, entry{"REN", pending, l})
					has = false
				}
			}
			if has {
				unpaired++
			}
			continue
		}
		var cur *entry
		open := false
		flush := func() {
			if cur != nil {
				if cur.Statement == "" {
					cur.Statement = cur.Name
				}
				out = append(out, *cur)
				cur = nil
			}
		}
		for _, l := range body {
			if m := reqRE.FindStringSubmatch(l); m != nil {
				flush()
				cur, open = &entry{Sec: h.kind, Name: normalizeName(m[1])}, true
				continue
			}
			if topRE.MatchString(l) {
				open = false
				continue
			}
			if h.kind == "REM" {
				if m := removedBulletRE.FindStringSubmatch(l); m != nil {
					flush()
					open = false
					name := normalizeName(m[1])
					out = append(out, entry{"REM", name, name})
					continue
				}
			}
			if cur != nil && open && cur.Statement == "" {
				if strings.HasPrefix(l, "#") {
					open = false
				} else if !blank(l) {
					cur.Statement = l
				}
			}
		}
		flush()
	}
	return out, unpaired
}
