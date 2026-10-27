package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

const ManifestDeclarationSchema = "proofbound.p7.projector-manifest-declaration.v1"

// CommittedSourceDeclaration names one already-committed source file to bind.
type CommittedSourceDeclaration struct {
	Path string
	Mode string
}

// CommittedModuleDeclaration names already-committed module metadata and archive bytes.
type CommittedModuleDeclaration struct {
	Module  string
	Version string
	ModPath string
	ZipPath string
}

// ManifestDeclarationInput describes the exact committed paths and build declarations used to
// derive one manifest. The reader supplies bytes at the declared evidence revision.
type ManifestDeclarationInput struct {
	EvidenceRevision  string
	ProjectorRevision string
	SourceFiles       []CommittedSourceDeclaration
	Toolchain         Toolchain
	Modules           []CommittedModuleDeclaration
	Build             Build
}

// ManifestDeclaration is a concrete non-runtime manifest plus explicit evidence incompleteness.
// It does not claim that the toolchain or the full projector/build evidence is available.
type ManifestDeclaration struct {
	Schema            string                        `json:"schema"`
	EvidenceRevision  string                        `json:"evidence_revision"`
	ProjectorRevision string                        `json:"projector_revision"`
	Manifest          Manifest                      `json:"manifest"`
	Evidence          []ManifestDeclarationEvidence `json:"evidence"`
}

// ManifestDeclarationEvidence records which declaration fields have committed-byte support.
type ManifestDeclarationEvidence struct {
	Field         string               `json:"field"`
	State         EvidenceClosureState `json:"state"`
	EvidencePaths []string             `json:"evidence_paths"`
	Reason        string               `json:"reason"`
}

// DeclareCommittedManifest reads only caller-supplied committed bytes and derives a validated
// manifest. Toolchain/archive evidence is explicit UNVERIFIED and other completeness evidence is
// explicit PENDING; neither state can become complete by absence.
func DeclareCommittedManifest(input ManifestDeclarationInput, reader CommittedEvidenceReader) (ManifestDeclaration, error) {
	if reader == nil {
		return ManifestDeclaration{}, errors.New("declaration: committed evidence reader is required")
	}
	if err := validateManifestDeclarationInput(input); err != nil {
		return ManifestDeclaration{}, err
	}
	sources := make([]SourceInput, 0, len(input.SourceFiles))
	for _, declaration := range input.SourceFiles {
		content, err := reader.ReadCommittedFile(input.EvidenceRevision, declaration.Path)
		if err != nil {
			return ManifestDeclaration{}, fmt.Errorf("declaration: source %s: %w", declaration.Path, err)
		}
		sources = append(sources, SourceInput{Path: declaration.Path, Mode: declaration.Mode, Bytes: content})
	}
	modules := make([]ModuleInput, 0, len(input.Modules))
	for _, declaration := range input.Modules {
		mod, err := reader.ReadCommittedFile(input.EvidenceRevision, declaration.ModPath)
		if err != nil {
			return ManifestDeclaration{}, fmt.Errorf("declaration: module %s .mod: %w", declaration.Module, err)
		}
		zip, err := reader.ReadCommittedFile(input.EvidenceRevision, declaration.ZipPath)
		if err != nil {
			return ManifestDeclaration{}, fmt.Errorf("declaration: module %s .zip: %w", declaration.Module, err)
		}
		modules = append(modules, ModuleInput{Module: declaration.Module, Version: declaration.Version, Mod: mod, Zip: zip})
	}
	manifest, err := Bind(BindingInput{
		EvidenceRevision:  input.EvidenceRevision,
		ProjectorRevision: input.ProjectorRevision,
		SourceFiles:       sources,
		Toolchain:         input.Toolchain,
		Modules:           modules,
		Build:             input.Build,
	})
	if err != nil {
		return ManifestDeclaration{}, err
	}
	declaration := ManifestDeclaration{
		Schema:            ManifestDeclarationSchema,
		EvidenceRevision:  input.EvidenceRevision,
		ProjectorRevision: input.ProjectorRevision,
		Manifest:          manifest,
		Evidence:          declarationEvidence(input),
	}
	if err := ValidateManifestDeclaration(declaration, input); err != nil {
		return ManifestDeclaration{}, err
	}
	return declaration, nil
}

// ValidateManifestDeclaration validates exact declaration inputs and explicit incomplete states.
func ValidateManifestDeclaration(declaration ManifestDeclaration, input ManifestDeclarationInput) error {
	if err := validateManifestDeclarationInput(input); err != nil {
		return err
	}
	if declaration.Schema != ManifestDeclarationSchema {
		return fmt.Errorf("declaration: schema must be %q", ManifestDeclarationSchema)
	}
	if declaration.EvidenceRevision != input.EvidenceRevision || declaration.ProjectorRevision != input.ProjectorRevision {
		return errors.New("declaration: revision mismatch")
	}
	if err := Validate(declaration.Manifest); err != nil {
		return fmt.Errorf("declaration: manifest: %w", err)
	}
	if declaration.Manifest.EvidenceRevision != input.EvidenceRevision || declaration.Manifest.ProjectorRevision != input.ProjectorRevision {
		return errors.New("declaration: manifest revisions do not match input")
	}
	wantEvidence := declarationEvidence(input)
	if !reflect.DeepEqual(declaration.Evidence, wantEvidence) {
		return errors.New("declaration: evidence does not match input")
	}
	return nil
}

// ParseManifestDeclaration validates strict canonical JSON against the exact declaration input.
func ParseManifestDeclaration(raw []byte, input ManifestDeclarationInput) (ManifestDeclaration, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return ManifestDeclaration{}, fmt.Errorf("declaration: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return ManifestDeclaration{}, errors.New("declaration: JSON is not canonical")
	}
	top, err := exactObject(raw, "schema", "evidence_revision", "projector_revision", "manifest", "evidence")
	if err != nil {
		return ManifestDeclaration{}, err
	}
	if _, err := Parse(top["manifest"]); err != nil {
		return ManifestDeclaration{}, fmt.Errorf("declaration: manifest: %w", err)
	}
	var fields []json.RawMessage
	if err := json.Unmarshal(top["evidence"], &fields); err != nil {
		return ManifestDeclaration{}, fmt.Errorf("declaration: evidence: %w", err)
	}
	for i, field := range fields {
		if _, err := exactObject(field, "field", "state", "evidence_paths", "reason"); err != nil {
			return ManifestDeclaration{}, fmt.Errorf("declaration: evidence field %d: %w", i, err)
		}
	}
	var declaration ManifestDeclaration
	if err := json.Unmarshal(raw, &declaration); err != nil {
		return ManifestDeclaration{}, fmt.Errorf("declaration: decode: %w", err)
	}
	if err := ValidateManifestDeclaration(declaration, input); err != nil {
		return ManifestDeclaration{}, err
	}
	return declaration, nil
}

// CanonicalManifestDeclarationBytes validates and encodes a declaration as RFC 8785 canonical JSON.
func CanonicalManifestDeclarationBytes(declaration ManifestDeclaration, input ManifestDeclarationInput) ([]byte, error) {
	if err := ValidateManifestDeclaration(declaration, input); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(declaration)
	if err != nil {
		return nil, fmt.Errorf("declaration: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("declaration: canonical encode: %w", err)
	}
	return canonical, nil
}

func validateManifestDeclarationInput(input ManifestDeclarationInput) error {
	if err := validateRevision("declaration evidence_revision", input.EvidenceRevision); err != nil {
		return err
	}
	if err := validateRevision("declaration projector_revision", input.ProjectorRevision); err != nil {
		return err
	}
	if len(input.SourceFiles) == 0 {
		return errors.New("declaration: source files must be non-empty")
	}
	previous := ""
	for i, source := range input.SourceFiles {
		if err := validateRelativePath(source.Path); err != nil {
			return fmt.Errorf("declaration: source file %d: %w", i, err)
		}
		if source.Path <= previous {
			return fmt.Errorf("declaration: source files must be sorted and unique at %q", source.Path)
		}
		previous = source.Path
		if source.Mode != "100644" && source.Mode != "100755" {
			return fmt.Errorf("declaration: source file %q has unsupported mode %q", source.Path, source.Mode)
		}
	}
	if len(input.Modules) == 0 {
		return errors.New("declaration: modules must be non-empty")
	}
	previous = ""
	for i, module := range input.Modules {
		if module.Module == "" || module.Version == "" {
			return fmt.Errorf("declaration: module %d requires module and version", i)
		}
		key := module.Module + "\x00" + module.Version
		if key <= previous {
			return fmt.Errorf("declaration: modules must be sorted and unique at %q", key)
		}
		previous = key
		if err := validateEvidencePath(module.ModPath); err != nil {
			return fmt.Errorf("declaration: module %s .mod: %w", module.Module, err)
		}
		if err := validateEvidencePath(module.ZipPath); err != nil {
			return fmt.Errorf("declaration: module %s .zip: %w", module.Module, err)
		}
		if module.ModPath == module.ZipPath {
			return fmt.Errorf("declaration: module %s reuses one path for .mod and .zip", module.Module)
		}
	}
	if err := validateToolchainValue(input.Toolchain); err != nil {
		return err
	}
	return validateBuildValue(input.Build)
}

func declarationEvidence(input ManifestDeclarationInput) []ManifestDeclarationEvidence {
	sourcePaths := make([]string, 0, len(input.SourceFiles))
	for _, source := range input.SourceFiles {
		sourcePaths = append(sourcePaths, source.Path)
	}
	modulePaths := make([]string, 0, len(input.Modules)*2)
	for _, module := range input.Modules {
		modulePaths = append(modulePaths, module.ModPath, module.ZipPath)
	}
	sort.Strings(modulePaths)
	bound := map[string][]string{
		"source.files": sourcePaths,
		"modules":      modulePaths,
	}
	toolchain := map[string]bool{
		"toolchain.archive_format": true,
		"toolchain.archive_sha256": true,
		"toolchain.go_version":     true,
		"toolchain.goarch":         true,
		"toolchain.goos":           true,
	}
	result := make([]ManifestDeclarationEvidence, 0, len(inventoryFields))
	for _, field := range inventoryFields {
		if paths, ok := bound[field]; ok {
			result = append(result, ManifestDeclarationEvidence{Field: field, State: EvidenceClosureBound, EvidencePaths: paths})
			continue
		}
		if toolchain[field] {
			result = append(result, ManifestDeclarationEvidence{
				Field:         field,
				State:         EvidenceClosureUnverified,
				EvidencePaths: []string{},
				Reason:        "toolchain/archive bytes are unavailable in the committed repository slice",
			})
			continue
		}
		result = append(result, ManifestDeclarationEvidence{
			Field:         field,
			State:         EvidenceClosurePending,
			EvidencePaths: []string{},
			Reason:        "independent completeness evidence is not established by this declaration",
		})
	}
	return result
}
