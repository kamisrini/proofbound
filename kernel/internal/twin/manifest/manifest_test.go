package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func TestBindComputesByteIdentitiesAndSorts(t *testing.T) {
	base := validManifest()
	bound, err := Bind(BindingInput{
		EvidenceRevision:  base.EvidenceRevision,
		ProjectorRevision: base.ProjectorRevision,
		SourceFiles: []SourceInput{
			{Path: "kernel/internal/twin/replay.go", Mode: "100644", Bytes: []byte("replay")},
			{Path: "kernel/go.mod", Mode: "100644", Bytes: []byte("module")},
		},
		Toolchain: base.Toolchain,
		Modules: []ModuleInput{
			{Module: "example.com/z", Version: "v1.0.0", Mod: []byte("z-mod"), Zip: []byte("z-zip")},
			{Module: "example.com/a", Version: "v1.0.0", Mod: []byte("a-mod"), Zip: []byte("a-zip")},
		},
		Build: base.Build,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := bound.Source.Files[0].Path; got != "kernel/go.mod" {
		t.Fatalf("source order=%q", got)
	}
	if got := bound.Source.Files[0].SHA256; got != digestBytes([]byte("module")) {
		t.Fatalf("source digest=%s", got)
	}
	if got := bound.Modules[0].Module; got != "example.com/a" {
		t.Fatalf("module order=%q", got)
	}
	if got := bound.Modules[0].ZipSHA256; got != digestBytes([]byte("a-zip")) {
		t.Fatalf("module zip digest=%s", got)
	}
}

func TestBindRejectsDuplicateAndInvalidInputs(t *testing.T) {
	base := validManifest()
	tests := map[string]func(*BindingInput){
		"duplicate source": func(input *BindingInput) {
			input.SourceFiles = append(input.SourceFiles, input.SourceFiles[0])
		},
		"duplicate module": func(input *BindingInput) {
			input.Modules = append(input.Modules, input.Modules[0])
		},
		"invalid source mode": func(input *BindingInput) { input.SourceFiles[0].Mode = "100600" },
		"invalid revision":    func(input *BindingInput) { input.EvidenceRevision = "working-tree" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			input := BindingInput{
				EvidenceRevision:  base.EvidenceRevision,
				ProjectorRevision: base.ProjectorRevision,
				SourceFiles:       []SourceInput{{Path: "kernel/go.mod", Mode: "100644", Bytes: []byte("module")}},
				Toolchain:         base.Toolchain,
				Modules:           []ModuleInput{{Module: "example.com/module", Version: "v1.0.0", Mod: []byte("mod"), Zip: []byte("zip")}},
				Build:             base.Build,
			}
			mutate(&input)
			if _, err := Bind(input); err == nil {
				t.Fatal("invalid binding accepted")
			}
		})
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

func TestInventoryBindsManifestAndEvidence(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	inventory.Fields[len(inventory.Fields)-1] = InventoryField{
		Field: inventory.Fields[len(inventory.Fields)-1].Field,
		State: InventoryUnverified,
	}
	raw, err := CanonicalInventoryBytes(inventory, manifest)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseInventory(raw, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, inventory) {
		t.Fatalf("parsed inventory differs:\n%+v\n%+v", parsed, inventory)
	}
}

func TestInventoryRejectsMissingExtraDuplicateStaleAndDigestInputs(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*EvidenceInventory){
		"missing field": func(inventory *EvidenceInventory) {
			inventory.Fields = inventory.Fields[:len(inventory.Fields)-1]
		},
		"extra field": func(inventory *EvidenceInventory) {
			inventory.Fields = append(inventory.Fields, inventory.Fields[len(inventory.Fields)-1])
		},
		"duplicate field": func(inventory *EvidenceInventory) {
			inventory.Fields[1].Field = inventory.Fields[0].Field
		},
		"stale revision": func(inventory *EvidenceInventory) {
			inventory.Fields[0].EvidenceRevision = strings.Repeat("b", 40)
		},
		"digest mismatch": func(inventory *EvidenceInventory) {
			inventory.ManifestDigest = strings.Repeat("0", 64)
		},
		"unverified evidence": func(inventory *EvidenceInventory) {
			inventory.Fields[0].State = InventoryUnverified
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			inventory := validInventory(manifest, manifestBytes)
			mutate(&inventory)
			if err := ValidateInventory(inventory, manifest); err == nil {
				t.Fatal("invalid inventory accepted")
			}
		})
	}
}

func TestParseInventoryRejectsUnknownAndNonCanonicalShape(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := CanonicalInventoryBytes(validInventory(manifest, manifestBytes), manifest)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object["unexpected"] = true
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(mutated)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseInventory(canonical, manifest); err == nil {
		t.Fatal("unknown inventory field accepted")
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseInventory(pretty.Bytes(), manifest); err == nil {
		t.Fatal("non-canonical inventory accepted")
	}
}

func TestResolveInventoryVerifiesCommittedBytesAndPreservesUnverified(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	content := []byte("committed evidence")
	contentDigest := digestBytes(content)
	for i := range inventory.Fields {
		inventory.Fields[i].EvidenceSHA256 = contentDigest
	}
	last := len(inventory.Fields) - 1
	inventory.Fields[last] = InventoryField{Field: inventory.Fields[last].Field, State: InventoryUnverified}
	reader := &memoryCommittedReader{
		revision: manifest.EvidenceRevision,
		files:    map[string][]byte{"docs/evidence/p7-source-toolchain.md": content},
	}
	resolved, err := ResolveInventory(inventory, manifest, reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != len(inventory.Fields) {
		t.Fatalf("resolved fields=%d want=%d", len(resolved), len(inventory.Fields))
	}
	if len(reader.calls) != len(inventory.Fields)-1 {
		t.Fatalf("reader calls=%d want=%d", len(reader.calls), len(inventory.Fields)-1)
	}
	if resolved[last].State != InventoryUnverified || resolved[last].EvidencePath != "" {
		t.Fatalf("unverified resolution=%+v", resolved[last])
	}
	for _, call := range reader.calls {
		if call.revision != manifest.EvidenceRevision || call.path != "docs/evidence/p7-source-toolchain.md" {
			t.Fatalf("reader call=%+v", call)
		}
	}
}

func TestResolveInventoryRejectsMissingRevisionUnsafeAndDigestEvidence(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("committed evidence")
	tests := map[string]func(*EvidenceInventory, *memoryCommittedReader){
		"missing path": func(inventory *EvidenceInventory, reader *memoryCommittedReader) {
			reader.files = map[string][]byte{}
		},
		"revision mismatch": func(inventory *EvidenceInventory, reader *memoryCommittedReader) {
			reader.revision = strings.Repeat("b", 40)
		},
		"unsafe path": func(inventory *EvidenceInventory, reader *memoryCommittedReader) {
			inventory.Fields[0].EvidencePath = "../escape"
		},
		"digest mismatch": func(inventory *EvidenceInventory, reader *memoryCommittedReader) {
			reader.files["docs/evidence/p7-source-toolchain.md"] = []byte("different bytes")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			inventory := validInventory(manifest, manifestBytes)
			contentDigest := digestBytes(content)
			for i := range inventory.Fields {
				inventory.Fields[i].EvidenceSHA256 = contentDigest
			}
			reader := &memoryCommittedReader{
				revision: manifest.EvidenceRevision,
				files:    map[string][]byte{"docs/evidence/p7-source-toolchain.md": content},
			}
			mutate(&inventory, reader)
			if _, err := ResolveInventory(inventory, manifest, reader); err == nil {
				t.Fatal("invalid committed evidence accepted")
			}
		})
	}
}

func TestEvaluateEvidenceClosureClassifiesWithoutClaimingAcceptance(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	resolutions := closureResolutions(inventory)
	closure, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	if closure.Overall != EvidenceClosureBound {
		t.Fatalf("complete closure=%q want %q", closure.Overall, EvidenceClosureBound)
	}

	inventory.Fields[len(inventory.Fields)-1] = InventoryField{Field: inventory.Fields[len(inventory.Fields)-1].Field, State: InventoryUnverified}
	resolutions = closureResolutions(inventory)
	closure, err = EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	if closure.Overall != EvidenceClosureUnverified {
		t.Fatalf("unverified closure=%q want %q", closure.Overall, EvidenceClosureUnverified)
	}

	inventory = validInventory(manifest, manifestBytes)
	resolutions = closureResolutions(inventory)
	resolutions[0] = EvidenceResolution{Field: resolutions[0].Field, State: EvidenceResolutionPending, Reason: "committed evidence is not yet available"}
	closure, err = EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	if closure.Overall != EvidenceClosurePending {
		t.Fatalf("pending closure=%q want %q", closure.Overall, EvidenceClosurePending)
	}
}

func TestEvaluateEvidenceClosureRejectsMissingExtraContradictoryAndStaleInputs(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*[]EvidenceResolution){
		"missing": func(resolutions *[]EvidenceResolution) { *resolutions = (*resolutions)[:len(*resolutions)-1] },
		"extra": func(resolutions *[]EvidenceResolution) {
			*resolutions = append(*resolutions, (*resolutions)[len(*resolutions)-1])
		},
		"duplicate": func(resolutions *[]EvidenceResolution) { (*resolutions)[1].Field = (*resolutions)[0].Field },
		"contradictory": func(resolutions *[]EvidenceResolution) {
			(*resolutions)[0] = EvidenceResolution{Field: (*resolutions)[0].Field, State: InventoryUnverified, Reason: "contradictory"}
		},
		"stale": func(resolutions *[]EvidenceResolution) {
			(*resolutions)[0].EvidenceRevision = strings.Repeat("b", 40)
		},
		"pending without reason": func(resolutions *[]EvidenceResolution) {
			(*resolutions)[0] = EvidenceResolution{Field: (*resolutions)[0].Field, State: EvidenceResolutionPending}
		},
		"invalid state": func(resolutions *[]EvidenceResolution) {
			(*resolutions)[0].State = InventoryState("MAYBE")
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			inventory := validInventory(manifest, manifestBytes)
			resolutions := closureResolutions(inventory)
			mutate(&resolutions)
			if _, err := EvaluateEvidenceClosure(inventory, manifest, resolutions); err == nil {
				t.Fatal("invalid closure input accepted")
			}
		})
	}
}

func TestEvidenceManifestBuildsCanonicalSkeleton(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	resolutions := closureResolutions(inventory)
	closure, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	skeleton, err := BuildEvidenceManifest(manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := CanonicalEvidenceManifestBytes(skeleton, manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseEvidenceManifest(raw, manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, skeleton) {
		t.Fatalf("parsed skeleton differs:\n%+v\n%+v", parsed, skeleton)
	}
}

func TestEvidenceManifestPreservesUnverifiedAndPending(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	last := len(inventory.Fields) - 1
	inventory.Fields[last] = InventoryField{Field: inventory.Fields[last].Field, State: InventoryUnverified}
	resolutions := closureResolutions(inventory)
	closure, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	skeleton, err := BuildEvidenceManifest(manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	if skeleton.ClosureOverall != EvidenceClosureUnverified || skeleton.Fields[last].State != EvidenceClosureUnverified {
		t.Fatalf("unverified skeleton=%+v", skeleton)
	}

	inventory = validInventory(manifest, manifestBytes)
	resolutions = closureResolutions(inventory)
	resolutions[0] = EvidenceResolution{Field: resolutions[0].Field, State: EvidenceResolutionPending, Reason: "evidence is pending owner review"}
	closure, err = EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	skeleton, err = BuildEvidenceManifest(manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	if skeleton.ClosureOverall != EvidenceClosurePending || skeleton.Fields[0].State != EvidenceClosurePending {
		t.Fatalf("pending skeleton=%+v", skeleton)
	}
}

func TestEvidenceManifestRejectsMismatchedAndNonCanonicalInputs(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inventory := validInventory(manifest, manifestBytes)
	resolutions := closureResolutions(inventory)
	closure, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		t.Fatal(err)
	}
	skeleton, err := BuildEvidenceManifest(manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*EvidenceManifest){
		"missing field":    func(value *EvidenceManifest) { value.Fields = value.Fields[:len(value.Fields)-1] },
		"extra field":      func(value *EvidenceManifest) { value.Fields = append(value.Fields, value.Fields[len(value.Fields)-1]) },
		"stale revision":   func(value *EvidenceManifest) { value.EvidenceRevision = strings.Repeat("b", 40) },
		"manifest digest":  func(value *EvidenceManifest) { value.ManifestDigest = strings.Repeat("0", 64) },
		"inventory digest": func(value *EvidenceManifest) { value.InventoryDigest = strings.Repeat("0", 64) },
		"closure overall":  func(value *EvidenceManifest) { value.ClosureOverall = EvidenceClosureUnverified },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := skeleton
			value.Fields = append([]EvidenceManifestField(nil), skeleton.Fields...)
			mutate(&value)
			if err := ValidateEvidenceManifest(value, manifest, inventory, resolutions, closure); err == nil {
				t.Fatal("invalid skeleton accepted")
			}
		})
	}
	badClosure := closure
	badClosure.Overall = EvidenceClosurePending
	if _, err := BuildEvidenceManifest(manifest, inventory, resolutions, badClosure); err == nil {
		t.Fatal("mismatched closure accepted")
	}
	staleResolutions := closureResolutions(inventory)
	staleResolutions[0].EvidenceRevision = strings.Repeat("b", 40)
	if _, err := BuildEvidenceManifest(manifest, inventory, staleResolutions, closure); err == nil {
		t.Fatal("stale resolution accepted")
	}
	raw, err := CanonicalEvidenceManifestBytes(skeleton, manifest, inventory, resolutions, closure)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object["unexpected"] = true
	mutated, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(mutated)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseEvidenceManifest(canonical, manifest, inventory, resolutions, closure); err == nil {
		t.Fatal("unknown skeleton field accepted")
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseEvidenceManifest(pretty.Bytes(), manifest, inventory, resolutions, closure); err == nil {
		t.Fatal("non-canonical skeleton accepted")
	}
}

func closureResolutions(inventory EvidenceInventory) []EvidenceResolution {
	resolutions := make([]EvidenceResolution, 0, len(inventory.Fields))
	for _, field := range inventory.Fields {
		if field.State == InventoryUnverified {
			resolutions = append(resolutions, EvidenceResolution{Field: field.Field, State: InventoryUnverified, Reason: "inventory explicitly UNVERIFIED"})
			continue
		}
		resolutions = append(resolutions, EvidenceResolution{
			Field:            field.Field,
			State:            InventoryBound,
			EvidencePath:     field.EvidencePath,
			EvidenceRevision: field.EvidenceRevision,
			EvidenceSHA256:   field.EvidenceSHA256,
		})
	}
	return resolutions
}

type memoryCommittedReader struct {
	revision string
	files    map[string][]byte
	calls    []memoryCommittedCall
}

type memoryCommittedCall struct {
	revision string
	path     string
}

func (reader *memoryCommittedReader) ReadCommittedFile(revision, path string) ([]byte, error) {
	reader.calls = append(reader.calls, memoryCommittedCall{revision: revision, path: path})
	if revision != reader.revision {
		return nil, fmt.Errorf("revision %q is not available", revision)
	}
	content, ok := reader.files[path]
	if !ok {
		return nil, fmt.Errorf("path %q is missing", path)
	}
	return append([]byte(nil), content...), nil
}

func validInventory(manifest Manifest, manifestBytes []byte) EvidenceInventory {
	fields := make([]InventoryField, 0, len(inventoryFields))
	for i, name := range inventoryFields {
		fields = append(fields, InventoryField{
			Field:            name,
			State:            InventoryBound,
			EvidencePath:     "docs/evidence/p7-source-toolchain.md",
			EvidenceRevision: manifest.EvidenceRevision,
			EvidenceSHA256:   fmt.Sprintf("%064x", i+1),
		})
	}
	return EvidenceInventory{
		Schema:           InventorySchema,
		EvidenceRevision: manifest.EvidenceRevision,
		ManifestDigest:   digestBytes(manifestBytes),
		Fields:           fields,
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
