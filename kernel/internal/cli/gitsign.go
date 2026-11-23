package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// `keys git-sign` (CLI-INV-44): makes this person's or agent's declared Proofbound key sign their git
// commits, so the Audit room can prove who made each commit instead of trusting the author name.

// openSSHPrivateKey encodes an ed25519 key in OpenSSH's unencrypted "openssh-key-v1" format, the
// form ssh-keygen and git's SSH signing read. Written by hand so no dependency is added (Law 8).
func openSSHPrivateKey(key ed25519.PrivateKey, comment string) ([]byte, error) {
	str := func(b *bytes.Buffer, s []byte) {
		var n [4]byte
		binary.BigEndian.PutUint32(n[:], uint32(len(s)))
		b.Write(n[:])
		b.Write(s)
	}
	u32 := func(b *bytes.Buffer, v uint32) {
		var n [4]byte
		binary.BigEndian.PutUint32(n[:], v)
		b.Write(n[:])
	}
	pub := key.Public().(ed25519.PublicKey)
	var pubBlob bytes.Buffer
	str(&pubBlob, []byte("ssh-ed25519"))
	str(&pubBlob, pub)
	var check [4]byte
	if _, err := rand.Read(check[:]); err != nil {
		return nil, err
	}
	var priv bytes.Buffer
	u32(&priv, binary.BigEndian.Uint32(check[:]))
	u32(&priv, binary.BigEndian.Uint32(check[:]))
	str(&priv, []byte("ssh-ed25519"))
	str(&priv, pub)
	str(&priv, key) // seed followed by public key, 64 bytes
	str(&priv, []byte(comment))
	for i := byte(1); priv.Len()%8 != 0; i++ {
		priv.WriteByte(i)
	}
	var out bytes.Buffer
	out.WriteString("openssh-key-v1\x00")
	str(&out, []byte("none"))
	str(&out, []byte("none"))
	str(&out, nil)
	u32(&out, 1)
	str(&out, pubBlob.Bytes())
	str(&out, priv.Bytes())
	return pem.EncodeToMemory(&pem.Block{Type: "OPENSSH PRIVATE KEY", Bytes: out.Bytes()}), nil
}

// keysGitSign writes the local key in OpenSSH form beside it (`<key>-git-ssh`, mode 0600, and its
// `.pub`) and sets this repository's git config to sign every commit with it. The key must already
// be declared at HEAD, so every signature it makes verifies in the Audit room.
func keysGitSign(ctx context.Context, root string, output io.Writer) error {
	a := localActor(ctx, root)
	if a.key == nil {
		return errors.New("keys git-sign: " + a.Problem)
	}
	keyPath, err := pbwarranty.KeyPath()
	if err != nil {
		return err
	}
	sshPath := keyPath + "-git-ssh"
	data, err := openSSHPrivateKey(a.key, "proofbound:"+a.Name)
	if err != nil {
		return err
	}
	pubLine := sshWireKey(a.key.Public().(ed25519.PublicKey)) + " proofbound:" + a.Name + "\n"
	if info, err := os.Lstat(sshPath); err == nil {
		// Already exported: keep it only if it is a regular 0600 file whose contents are this key's
		// export (review ui-u2-review-20261122 F6), never replace it.
		existing, rerr := os.ReadFile(sshPath)
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || rerr != nil || !sameOpenSSHKey(existing, a.key) {
			return fmt.Errorf("keys git-sign: %s exists and is not this key's export; move it away first", sshPath)
		}
	} else {
		// Created 0600 and exclusively: never readable by others, not even for a moment (F2).
		f, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return fmt.Errorf("keys git-sign: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			os.Remove(sshPath)
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	if err := os.WriteFile(sshPath+".pub", []byte(pubLine), 0o644); err != nil {
		return err
	}
	// Per worktree (F3): in a linked worktree a plain --local would change every worktree's signer.
	if out, err := gitRun(ctx, root, "config", "--local", "extensions.worktreeConfig", "true"); err != nil {
		return fmt.Errorf("keys git-sign: git config: %s", out)
	}
	for _, kv := range [][2]string{{"gpg.format", "ssh"}, {"user.signingkey", sshPath}, {"commit.gpgsign", "true"}} {
		if out, err := gitRun(ctx, root, "config", "--worktree", kv[0], kv[1]); err != nil {
			return fmt.Errorf("keys git-sign: git config %s: %s", kv[0], out)
		}
	}
	_, err = fmt.Fprintf(output, "git-sign %s key=%s ssh=%s\nevery commit in this repository is now signed as %s; the Audit room shows it as verified\n", a.Name, a.Fingerprint, sshPath, a.Name)
	return err
}

// sameOpenSSHKey reports whether an OpenSSH private-key file holds exactly key (its random check
// bytes differ between exports, so only the key blob is compared).
func sameOpenSSHKey(file []byte, key ed25519.PrivateKey) bool {
	block, _ := pem.Decode(file)
	if block == nil || block.Type != "OPENSSH PRIVATE KEY" {
		return false
	}
	var want bytes.Buffer
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(key)))
	want.Write(n[:])
	want.Write(key)
	return bytes.Contains(block.Bytes, want.Bytes())
}
