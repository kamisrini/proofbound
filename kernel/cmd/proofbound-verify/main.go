//go:build !js

package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

//go:embed page.html
var pageTemplate string

// main assembles the page: proofbound-verify -wasm <verify.wasm> -exec <wasm_exec.js> -version <v> -o <page.html>.
func main() {
	if err := build(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "proofbound-verify:", err)
		os.Exit(1)
	}
}

func build(args []string) error {
	fs := flag.NewFlagSet("proofbound-verify", flag.ContinueOnError)
	wasm := fs.String("wasm", "", "the GOOS=js GOARCH=wasm build of this command")
	exec := fs.String("exec", "", "wasm_exec.js from the same Go toolchain ($(go env GOROOT)/lib/wasm/wasm_exec.js)")
	version := fs.String("version", "", "the Proofbound release this page belongs to")
	out := fs.String("o", "", "the page to write")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *wasm == "" || *exec == "" || *version == "" || *out == "" {
		return errors.New("usage: proofbound-verify -wasm <file> -exec <wasm_exec.js> -version <v> -o <page.html>")
	}
	page, err := assemble(*wasm, *exec, *version)
	if err != nil {
		return err
	}
	return os.WriteFile(*out, page, 0o644)
}

// assemble puts the gzipped, base64 WebAssembly and Go's loader into the page template.
func assemble(wasmPath, execPath, version string) ([]byte, error) {
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(wasm, []byte("\x00asm")) {
		return nil, fmt.Errorf("%s is not WebAssembly", wasmPath)
	}
	loader, err := os.ReadFile(execPath)
	if err != nil {
		return nil, err
	}
	if bytes.Contains(bytes.ToLower(loader), []byte("</script")) {
		return nil, errors.New("the loader contains </script")
	}
	var gz bytes.Buffer
	w, _ := gzip.NewWriterLevel(&gz, gzip.BestCompression)
	if _, err := w.Write(wasm); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	page := strings.NewReplacer(
		"{{VERSION}}", htmlText(version),
		"{{LOADER}}", string(loader),
		"{{WASM}}", base64.StdEncoding.EncodeToString(gz.Bytes()),
	).Replace(pageTemplate)
	return []byte(page), nil
}

func htmlText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
