package sdd

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// Kit is the OpenSpec kit, M1's (VD-sdd-m1-openspec-first-2026-11-17); KitSpecKit is Spec Kit's
// (docs/plans/R2-speckit-spec.md). suffix hashes the kit, so two kits never share an id.
const (
	Kit        = "openspec"
	KitSpecKit = "speckit"
)

var (
	errEmptyNamePart = errors.New("empty-name-part")
	errSlugTooShort  = errors.New("slug-too-short")
)

// asciiMap is SPEC §4.2/§4.3's byte-wise ASCII mapping: never strings.ToUpper/ToLower, whose
// Unicode folding would turn "ſ" into "S" (SDD-INV-7).
func asciiMap(s string, upper bool) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case upper && c >= 'a' && c <= 'z':
			c -= 'a' - 'A'
		case !upper && c >= 'A' && c <= 'Z':
			c += 'a' - 'A'
		}
		keep := c >= '0' && c <= '9' || upper && c >= 'A' && c <= 'Z' || !upper && c >= 'a' && c <= 'z'
		if !keep {
			c = '-'
		}
		if c == '-' && len(b) > 0 && b[len(b)-1] == '-' {
			continue
		}
		b = append(b, c)
	}
	return strings.Trim(string(b), "-")
}

func namepart(name string) (string, error) {
	np := asciiMap(name, true)
	if np == "" {
		return "", errEmptyNamePart
	}
	if len(np) > 28 {
		sum := sha256.Sum256([]byte(name))
		np = strings.TrimRight(np[:21], "-") + "-" + strings.ToUpper(hex.EncodeToString(sum[:])[:6])
	}
	return np, nil
}

func slug(s string) (string, error) {
	out := asciiMap(s, false)
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	if len(out) < 2 {
		return "", errSlugTooShort
	}
	return out, nil
}

func suffix(kit, kind, key string) string {
	sum := sha256.Sum256([]byte(kit + "\x00" + kind + "\x00" + key))
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = alphabet[int(sum[i])%36]
	}
	return string(b)
}

func changeIntentID(kit, name string) (string, error) {
	s, err := slug(name)
	if err != nil {
		return "", err
	}
	return "CI-" + s + "-" + suffix(kit, "change-intent", name), nil
}

func requirementID(kit, key, digest string) (string, error) {
	s, err := slug(key)
	if err != nil {
		return "", err
	}
	return "BR-" + s + "-" + suffix(kit, "requirement", key+"\x00"+digest), nil
}

var secCode = map[string]string{"ADD": "ADD", "MOD": "MOD", "REM": "REM", "REN": "REN"}

func obligationID(sec, name string) (string, error) {
	np, err := namepart(name)
	if err != nil {
		return "", err
	}
	return "O-" + secCode[sec] + "-" + np, nil
}
