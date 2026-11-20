package warranty

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KeyPath is where the signing key lives: $PROOFBOUND_KEY verbatim when set (a second identity on one
// machine, MVP step 3), else ${XDG_CONFIG_HOME:-$HOME/.config}/proofbound/signing-key.
func KeyPath() (string, error) {
	if p := os.Getenv("PROOFBOUND_KEY"); p != "" {
		return p, nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "proofbound", "signing-key"), nil
}

// PubLine is the committed form of a public key: `ed25519 <64 hex>\n`.
func PubLine(pub ed25519.PublicKey) string { return "ed25519 " + hex.EncodeToString(pub) + "\n" }

// insideWorkTree reports whether dir is inside a git work tree (W-INV-8).
func insideWorkTree(dir string) bool {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(d); err == nil {
			out, err := exec.Command("git", "-C", d, "rev-parse", "--is-inside-work-tree").Output()
			return err == nil && strings.TrimSpace(string(out)) == "true"
		}
		if filepath.Dir(d) == d {
			return false
		}
	}
}

// InitKey creates a new signing key at path. It refuses an existing file and any path inside a git
// work tree, so the private key can never be committed (W-INV-8).
func InitKey(path string) (ed25519.PublicKey, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if insideWorkTree(filepath.Dir(abs)) {
		return nil, fmt.Errorf("keys: %s is inside a git work tree; set XDG_CONFIG_HOME elsewhere", abs)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("keys: %s exists; refusing to replace a signing key", abs)
		}
		return nil, err
	}
	if _, err := f.WriteString(hex.EncodeToString(priv.Seed()) + "\n"); err != nil {
		f.Close()
		os.Remove(abs)
		return nil, err
	}
	return pub, f.Close()
}

// LoadKey reads the signing key, refusing a missing file, a symlink, or one readable by others.
func LoadKey(path string) (ed25519.PrivateKey, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("keys: no signing key at %s; run `proofbound keys init`", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("keys: %s is not a regular file", path)
	}
	if info.Mode().Perm() != 0o600 {
		return nil, fmt.Errorf("keys: %s has mode %o; it must be 0600", path, info.Mode().Perm())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("keys: %s is not a hex ed25519 seed", path)
	}
	return ed25519.NewKeyFromSeed(seed), nil
}
