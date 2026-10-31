package manifest

import (
	"strings"
	"testing"
)

// These tests close the per-branch coverage gaps found by the calibrated mutation
// sweep (scripts/mutants.sh internal/twin/manifest) during P6-Task-8-style
// acceptance of this package. The existing reject-tests trip compound `||`/`&&`
// conditions as a whole; each case below makes exactly ONE sub-condition true so
// that flipping that operand changes the accept/reject outcome and the mutant dies.

// Validate: each surviving operand in validateSourceValue, validateToolchainValue,
// and validateModuleValues (manifest.go:277,305,370,374).
func TestValidateRejectsEachUnsafeSubBranch(t *testing.T) {
	tests := map[string]func(*Manifest){
		"toolchain goos invalid":     func(m *Manifest) { m.Toolchain.GOOS = "" },
		"toolchain goarch invalid":   func(m *Manifest) { m.Toolchain.GOARCH = "" },
		"module empty name":          func(m *Manifest) { m.Modules[0].Module = "" },
		"module empty version":       func(m *Manifest) { m.Modules[0].Version = "" },
		"source path empty":          func(m *Manifest) { m.Source.Files[0].Path = "" },
		"source path backslash":      func(m *Manifest) { m.Source.Files[0].Path = "a\\b" },
		"source path leading slash":  func(m *Manifest) { m.Source.Files[0].Path = "/x" },
		"source path non-canonical":  func(m *Manifest) { m.Source.Files[0].Path = "a/b/../c" },
		"source path dot":            func(m *Manifest) { m.Source.Files[0].Path = "." },
		"source path dotdot":         func(m *Manifest) { m.Source.Files[0].Path = ".." },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			m := validManifest()
			mutate(&m)
			if err := Validate(m); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

// ValidateInventory: each surviving operand in the BOUND and UNVERIFIED branches
// (inventory.go:151,155,165).
func TestValidateInventoryRejectsEachEvidenceSubBranch(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*EvidenceInventory){
		"bound invalid sha": func(inv *EvidenceInventory) { inv.Fields[0].EvidenceSHA256 = "bad" },
		"bound empty path":  func(inv *EvidenceInventory) { inv.Fields[0].EvidencePath = "" },
		"bound nul path":    func(inv *EvidenceInventory) { inv.Fields[0].EvidencePath = "a\x00b" },
		"unverified declares path": func(inv *EvidenceInventory) {
			inv.Fields[0] = InventoryField{Field: inv.Fields[0].Field, State: InventoryUnverified, EvidencePath: "docs/x.md"}
		},
		"unverified declares revision": func(inv *EvidenceInventory) {
			inv.Fields[0] = InventoryField{Field: inv.Fields[0].Field, State: InventoryUnverified, EvidenceRevision: manifest.EvidenceRevision}
		},
		"unverified declares sha": func(inv *EvidenceInventory) {
			inv.Fields[0] = InventoryField{Field: inv.Fields[0].Field, State: InventoryUnverified, EvidenceSHA256: strings.Repeat("1", 64)}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			inv := validInventory(manifest, manifestBytes)
			mutate(&inv)
			if err := ValidateInventory(inv, manifest); err == nil {
				t.Fatal("invalid inventory accepted")
			}
		})
	}
}

// EvaluateEvidenceClosure: a resolution that declares evidence it must not, for
// each of the UNVERIFIED and PENDING branches (evidence.go:166,177).
func TestEvaluateEvidenceClosureRejectsResolutionEvidenceSubBranch(t *testing.T) {
	manifest := validManifest()
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}

	// UNVERIFIED resolution (inventory field must also be UNVERIFIED) must not declare evidence.
	unverified := map[string]func(*EvidenceResolution){
		"path":     func(r *EvidenceResolution) { r.EvidencePath = "docs/x.md" },
		"revision": func(r *EvidenceResolution) { r.EvidenceRevision = strings.Repeat("a", 40) },
		"sha":      func(r *EvidenceResolution) { r.EvidenceSHA256 = strings.Repeat("1", 64) },
	}
	for name, mutate := range unverified {
		t.Run("unverified declares "+name, func(t *testing.T) {
			inv := validInventory(manifest, manifestBytes)
			last := len(inv.Fields) - 1
			inv.Fields[last] = InventoryField{Field: inv.Fields[last].Field, State: InventoryUnverified}
			res := closureResolutions(inv)
			mutate(&res[last])
			if _, err := EvaluateEvidenceClosure(inv, manifest, res); err == nil {
				t.Fatal("UNVERIFIED resolution with evidence accepted")
			}
		})
	}

	// PENDING resolution (inventory field BOUND) must not declare evidence.
	pending := map[string]func(*EvidenceResolution){
		"path":     func(r *EvidenceResolution) { r.EvidencePath = "docs/x.md" },
		"revision": func(r *EvidenceResolution) { r.EvidenceRevision = strings.Repeat("a", 40) },
		"sha":      func(r *EvidenceResolution) { r.EvidenceSHA256 = strings.Repeat("1", 64) },
	}
	for name, mutate := range pending {
		t.Run("pending declares "+name, func(t *testing.T) {
			inv := validInventory(manifest, manifestBytes)
			res := closureResolutions(inv)
			res[0] = EvidenceResolution{Field: res[0].Field, State: EvidenceResolutionPending, Reason: "pending"}
			mutate(&res[0])
			if _, err := EvaluateEvidenceClosure(inv, manifest, res); err == nil {
				t.Fatal("PENDING resolution with evidence accepted")
			}
		})
	}
}

// ValidateEvidenceManifest: each attribute in the per-field match against the
// derived closure (skeleton.go:114).
func TestValidateEvidenceManifestRejectsEachFieldAttribute(t *testing.T) {
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
	tests := map[string]func(*EvidenceManifestField){
		"state":    func(f *EvidenceManifestField) { f.State = EvidenceClosureUnverified },
		"path":     func(f *EvidenceManifestField) { f.EvidencePath = "docs/other.md" },
		"revision": func(f *EvidenceManifestField) { f.EvidenceRevision = strings.Repeat("9", 40) },
		"sha":      func(f *EvidenceManifestField) { f.EvidenceSHA256 = strings.Repeat("9", 64) },
		"reason":   func(f *EvidenceManifestField) { f.Reason = "tampered" },
	}
	for name, mutate := range tests {
		t.Run("field "+name, func(t *testing.T) {
			value := skeleton
			value.Fields = append([]EvidenceManifestField(nil), skeleton.Fields...)
			mutate(&value.Fields[0])
			if err := ValidateEvidenceManifest(value, manifest, inventory, resolutions, closure); err == nil {
				t.Fatal("tampered skeleton field accepted")
			}
		})
	}
}

// A source file with the executable mode 100755 must be ACCEPTED, pinning the
// "100755" literal in the mode check (declaration.go:207, manifest.go:256) so a
// string->"" mutation that would wrongly reject executable sources dies.
func TestManifestDeclarationAcceptsExecutableSourceMode(t *testing.T) {
	base := validManifest()
	input := ManifestDeclarationInput{
		EvidenceRevision:  base.EvidenceRevision,
		ProjectorRevision: base.ProjectorRevision,
		SourceFiles:       []CommittedSourceDeclaration{{Path: "kernel/go.mod", Mode: "100755"}},
		Toolchain:         base.Toolchain,
		Modules:           []CommittedModuleDeclaration{{Module: "example.com/module", Version: "v1.0.0", ModPath: "modules/a.mod", ZipPath: "modules/a.zip"}},
		Build:             base.Build,
	}
	reader := &memoryCommittedReader{revision: input.EvidenceRevision, files: map[string][]byte{
		"kernel/go.mod": []byte("module committed"), "modules/a.mod": []byte("mod"), "modules/a.zip": []byte("zip"),
	}}
	if _, err := DeclareCommittedManifest(input, reader); err != nil {
		t.Fatalf("executable source mode 100755 was rejected: %v", err)
	}
}

// Declaration: each surviving operand in validateManifestDeclarationInput
// (declaration.go:189,192,207,216,234) and the revision-mismatch checks in
// ValidateManifestDeclaration (declaration.go:121,127).
func TestManifestDeclarationRejectsEachInputSubBranch(t *testing.T) {
	base := validManifest()
	input := ManifestDeclarationInput{
		EvidenceRevision:  base.EvidenceRevision,
		ProjectorRevision: base.ProjectorRevision,
		SourceFiles:       []CommittedSourceDeclaration{{Path: "kernel/go.mod", Mode: "100644"}},
		Toolchain:         base.Toolchain,
		Modules:           []CommittedModuleDeclaration{{Module: "example.com/module", Version: "v1.0.0", ModPath: "modules/a.mod", ZipPath: "modules/a.zip"}},
		Build:             base.Build,
	}
	reader := &memoryCommittedReader{revision: input.EvidenceRevision, files: map[string][]byte{
		"kernel/go.mod": []byte("module committed"), "modules/a.mod": []byte("mod"), "modules/a.zip": []byte("zip"),
	}}

	inputCases := map[string]func(*ManifestDeclarationInput){
		"empty evidence revision":   func(v *ManifestDeclarationInput) { v.EvidenceRevision = "" },
		"invalid projector revision": func(v *ManifestDeclarationInput) { v.ProjectorRevision = "nothex" },
		"invalid source mode":       func(v *ManifestDeclarationInput) { v.SourceFiles[0].Mode = "100600" },
		"empty module name":         func(v *ManifestDeclarationInput) { v.Modules[0].Module = "" },
		"empty module version":      func(v *ManifestDeclarationInput) { v.Modules[0].Version = "" },
		"invalid toolchain":         func(v *ManifestDeclarationInput) { v.Toolchain.GoVersion = "nope" },
	}
	for name, mutate := range inputCases {
		t.Run(name, func(t *testing.T) {
			value := input
			value.SourceFiles = append([]CommittedSourceDeclaration(nil), input.SourceFiles...)
			value.Modules = append([]CommittedModuleDeclaration(nil), input.Modules...)
			mutate(&value)
			if _, err := DeclareCommittedManifest(value, reader); err == nil {
				t.Fatal("invalid declaration input accepted")
			}
		})
	}

	declaration, err := DeclareCommittedManifest(input, reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("declaration evidence revision mismatch", func(t *testing.T) {
		tampered := declaration
		tampered.EvidenceRevision = strings.Repeat("9", 40)
		if err := ValidateManifestDeclaration(tampered, input); err == nil {
			t.Fatal("declaration revision mismatch accepted")
		}
	})
	t.Run("declaration manifest revision mismatch", func(t *testing.T) {
		tampered := declaration
		tampered.Manifest.EvidenceRevision = strings.Repeat("9", 40)
		if err := ValidateManifestDeclaration(tampered, input); err == nil {
			t.Fatal("declaration manifest revision mismatch accepted")
		}
	})
}
