package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

func TestDigestAndRevisionBinding(t *testing.T) {
	base := validManifest()
	raw, err := CanonicalBytes(base)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, base) {
		t.Fatalf("parsed manifest differs:\n%+v\n%+v", parsed, base)
	}
	wantSum := sha256.Sum256(raw)
	want := hex.EncodeToString(wantSum[:])
	got, err := Digest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("digest=%s want=%s", got, want)
	}

	changed := base
	changed.EvidenceRevision = strings.Repeat("c", 40)
	changedRaw, err := CanonicalBytes(changed)
	if err != nil {
		t.Fatal(err)
	}
	changedDigest, err := Digest(changedRaw)
	if err != nil {
		t.Fatal(err)
	}
	if changedDigest == got {
		t.Fatal("changing the evidence revision did not change the digest")
	}
}

func TestParseRejectsShapeAndCanonicality(t *testing.T) {
	raw, err := CanonicalBytes(validManifest())
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(map[string]any){
		"unknown field": func(object map[string]any) { object["unexpected"] = true },
		"missing field": func(object map[string]any) { delete(object, "build") },
		"null field":    func(object map[string]any) { object["source"] = nil },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			mutated := canonicalObject(t, raw, mutate)
			if _, err := Parse(mutated); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(pretty.Bytes()); err == nil {
		t.Fatal("non-canonical manifest accepted")
	}
}

func TestValidateRejectsBindingOrderAndDigest(t *testing.T) {
	tests := map[string]func(*Manifest){
		"source order": func(manifest *Manifest) {
			manifest.Source.Files[0], manifest.Source.Files[1] = manifest.Source.Files[1], manifest.Source.Files[0]
		},
		"unsafe source path": func(manifest *Manifest) { manifest.Source.Files[0].Path = "../escape" },
		"source digest":      func(manifest *Manifest) { manifest.Source.Files[0].SHA256 = "bad" },
		"module order": func(manifest *Manifest) {
			manifest.Modules = append(manifest.Modules, Module{Module: "a.example/module", Version: "v1.0.0", ModSHA256: strings.Repeat("a", 64), ZipSHA256: strings.Repeat("b", 64)})
		},
		"toolchain digest": func(manifest *Manifest) { manifest.Toolchain.ArchiveSHA256 = strings.Repeat("A", 64) },
		"tag order":        func(manifest *Manifest) { manifest.Build.Tags = []string{"z", "a"} },
		"revision":         func(manifest *Manifest) { manifest.ProjectorRevision = strings.Repeat("g", 40) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := validManifest()
			mutate(&manifest)
			if err := Validate(manifest); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func validManifest() Manifest {
	return Manifest{
		Schema:            Schema,
		EvidenceRevision:  strings.Repeat("a", 40),
		ProjectorRevision: strings.Repeat("b", 40),
		Source: Source{Files: []File{
			{Path: "kernel/go.mod", Mode: "100644", SHA256: strings.Repeat("c", 64)},
			{Path: "kernel/internal/twin/replay.go", Mode: "100644", SHA256: strings.Repeat("d", 64)},
		}},
		Toolchain: Toolchain{
			GoVersion:     "go1.26.7",
			GOOS:          "linux",
			GOARCH:        "amd64",
			ArchiveFormat: "tar.gz",
			ArchiveSHA256: strings.Repeat("e", 64),
		},
		Modules: []Module{{
			Module:    "example.com/module",
			Version:   "v1.0.0",
			ModSHA256: strings.Repeat("f", 64),
			ZipSHA256: strings.Repeat("0", 64),
		}},
		Build: Build{
			Command:    "go build ./...",
			CGOEnabled: false,
			GCFlags:    "",
			LDFlags:    "",
			Tags:       []string{"reproducible"},
		},
	}
}

func canonicalObject(t *testing.T, raw []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	mutate(object)
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(mutated)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
