//go:build !js

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The page carries the loader and the compressed WebAssembly, and nothing is left unfilled.
func TestAssemble(t *testing.T) {
	dir := t.TempDir()
	wasm := append([]byte("\x00asm\x01\x00\x00\x00"), bytes.Repeat([]byte("x"), 1000)...)
	wp, ep := filepath.Join(dir, "v.wasm"), filepath.Join(dir, "wasm_exec.js")
	if err := os.WriteFile(wp, wasm, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ep, []byte("globalThis.Go = function () {};"), 0o644); err != nil {
		t.Fatal(err)
	}
	page, err := assemble(wp, ep, "v0.3.2<")
	if err != nil {
		t.Fatal(err)
	}
	s := string(page)
	if strings.Contains(s, "{{") || !strings.Contains(s, "globalThis.Go = function") || !strings.Contains(s, "Proofbound v0.3.2&lt;") {
		t.Fatal("the page is not filled")
	}
	m := regexp.MustCompile(`<script id="wasm" type="application/octet-stream">([A-Za-z0-9+/=]+)</script>`).FindStringSubmatch(s)
	if m == nil {
		t.Fatal("no WebAssembly in the page")
	}
	gz, _ := base64.StdEncoding.DecodeString(m[1])
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	if !bytes.Equal(got, wasm) {
		t.Error("the WebAssembly does not round-trip")
	}
	if err := os.WriteFile(wp, []byte("not wasm"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := assemble(wp, ep, "v"); err == nil {
		t.Error("a file that is not WebAssembly was assembled")
	}
}
