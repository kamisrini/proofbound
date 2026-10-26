package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

const Schema = "proofbound.p7.projector-manifest.v1"

var (
	gitRevisionRE = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256RE      = regexp.MustCompile(`^[0-9a-f]{64}$`)
	platformRE    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,15}$`)
	goVersionRE   = regexp.MustCompile(`^go[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)
)

// Manifest is the non-runtime identity contract for one projector/build input set.
// It declares identities; it does not retrieve or validate external bytes.
type Manifest struct {
	Schema            string    `json:"schema"`
	EvidenceRevision  string    `json:"evidence_revision"`
	ProjectorRevision string    `json:"projector_revision"`
	Source            Source    `json:"source"`
	Toolchain         Toolchain `json:"toolchain"`
	Modules           []Module  `json:"modules"`
	Build             Build     `json:"build"`
}

// Source binds the exact tracked source files consumed by the proposed build.
type Source struct {
	Files []File `json:"files"`
}

// File is a content-addressed source-tree entry.
type File struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	SHA256 string `json:"sha256"`
}

// Toolchain declares the toolchain identity without fetching or executing it.
type Toolchain struct {
	GoVersion     string `json:"go_version"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	ArchiveFormat string `json:"archive_format"`
	ArchiveSHA256 string `json:"archive_sha256"`
}

// Module binds the selected module metadata and source archive identities.
type Module struct {
	Module    string `json:"module"`
	Version   string `json:"version"`
	ModSHA256 string `json:"mod_sha256"`
	ZipSHA256 string `json:"zip_sha256"`
}

// Build records the deterministic build settings without running the build.
type Build struct {
	Command    string   `json:"command"`
	CGOEnabled bool     `json:"cgo_enabled"`
	GCFlags    string   `json:"gcflags"`
	LDFlags    string   `json:"ldflags"`
	Tags       []string `json:"tags"`
}

// Parse validates strict canonical JSON and returns its typed manifest.
func Parse(raw []byte) (Manifest, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return Manifest{}, errors.New("manifest: JSON is not canonical")
	}
	top, err := exactObject(raw, "schema", "evidence_revision", "projector_revision", "source", "toolchain", "modules", "build")
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("manifest: decode: %w", err)
	}
	if manifest.Schema != Schema {
		return Manifest{}, fmt.Errorf("manifest: schema must be %q", Schema)
	}
	if err := validateRevision("evidence_revision", manifest.EvidenceRevision); err != nil {
		return Manifest{}, err
	}
	if err := validateRevision("projector_revision", manifest.ProjectorRevision); err != nil {
		return Manifest{}, err
	}
	if err := validateSource(top["source"], manifest.Source); err != nil {
		return Manifest{}, err
	}
	if err := validateToolchain(top["toolchain"], manifest.Toolchain); err != nil {
		return Manifest{}, err
	}
	if err := validateModules(top["modules"], manifest.Modules); err != nil {
		return Manifest{}, err
	}
	if err := validateBuild(top["build"], manifest.Build); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// CanonicalBytes validates and encodes a manifest as RFC 8785 canonical JSON.
func CanonicalBytes(manifest Manifest) ([]byte, error) {
	if err := Validate(manifest); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("manifest: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("manifest: canonical encode: %w", err)
	}
	return canonical, nil
}

// Digest returns the SHA-256 of the validated canonical manifest bytes.
func Digest(raw []byte) (string, error) {
	if _, err := Parse(raw); err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// Validate validates a typed manifest. Parse should be preferred for untrusted bytes.
func Validate(manifest Manifest) error {
	if manifest.Schema != Schema {
		return fmt.Errorf("manifest: schema must be %q", Schema)
	}
	if err := validateRevision("evidence_revision", manifest.EvidenceRevision); err != nil {
		return err
	}
	if err := validateRevision("projector_revision", manifest.ProjectorRevision); err != nil {
		return err
	}
	if err := validateSourceValue(manifest.Source); err != nil {
		return err
	}
	if err := validateToolchainValue(manifest.Toolchain); err != nil {
		return err
	}
	if err := validateModuleValues(manifest.Modules); err != nil {
		return err
	}
	return validateBuildValue(manifest.Build)
}

func validateSource(raw json.RawMessage, source Source) error {
	if _, err := exactObject(raw, "files"); err != nil {
		return fmt.Errorf("manifest: source: %w", err)
	}
	var files []json.RawMessage
	if err := json.Unmarshal(rawField(raw, "files"), &files); err != nil {
		return fmt.Errorf("manifest: source files: %w", err)
	}
	for i, file := range files {
		if _, err := exactObject(file, "path", "mode", "sha256"); err != nil {
			return fmt.Errorf("manifest: source file %d: %w", i, err)
		}
	}
	return validateSourceValue(source)
}

func validateSourceValue(source Source) error {
	if len(source.Files) == 0 {
		return errors.New("manifest: source files must be non-empty")
	}
	previous := ""
	for i, file := range source.Files {
		if err := validateRelativePath(file.Path); err != nil {
			return fmt.Errorf("manifest: source file %d: %w", i, err)
		}
		if file.Path <= previous {
			return fmt.Errorf("manifest: source files must be sorted and unique at %q", file.Path)
		}
		previous = file.Path
		if file.Mode != "100644" && file.Mode != "100755" {
			return fmt.Errorf("manifest: source file %q has unsupported mode %q", file.Path, file.Mode)
		}
		if err := validateSHA256("source file "+file.Path, file.SHA256); err != nil {
			return err
		}
	}
	return nil
}

func validateToolchain(raw json.RawMessage, toolchain Toolchain) error {
	if _, err := exactObject(raw, "go_version", "goos", "goarch", "archive_format", "archive_sha256"); err != nil {
		return fmt.Errorf("manifest: toolchain: %w", err)
	}
	return validateToolchainValue(toolchain)
}

func validateToolchainValue(toolchain Toolchain) error {
	if !goVersionRE.MatchString(toolchain.GoVersion) {
		return fmt.Errorf("manifest: invalid go_version %q", toolchain.GoVersion)
	}
	if !platformRE.MatchString(toolchain.GOOS) || !platformRE.MatchString(toolchain.GOARCH) {
		return fmt.Errorf("manifest: invalid platform %q/%q", toolchain.GOOS, toolchain.GOARCH)
	}
	if toolchain.ArchiveFormat != "tar.gz" {
		return fmt.Errorf("manifest: toolchain archive_format must be %q", "tar.gz")
	}
	return validateSHA256("toolchain archive", toolchain.ArchiveSHA256)
}

func validateModules(raw json.RawMessage, modules []Module) error {
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("manifest: modules: %w", err)
	}
	for i, entry := range entries {
		if _, err := exactObject(entry, "module", "version", "mod_sha256", "zip_sha256"); err != nil {
			return fmt.Errorf("manifest: module %d: %w", i, err)
		}
	}
	return validateModuleValues(modules)
}

func validateModuleValues(modules []Module) error {
	if len(modules) == 0 {
		return errors.New("manifest: modules must be non-empty")
	}
	previous := ""
	for i, module := range modules {
		if module.Module == "" || module.Version == "" {
			return fmt.Errorf("manifest: module %d requires module and version", i)
		}
		key := module.Module + "\x00" + module.Version
		if key <= previous {
			return fmt.Errorf("manifest: modules must be sorted and unique at %q", key)
		}
		previous = key
		if err := validateSHA256("module .mod "+module.Module, module.ModSHA256); err != nil {
			return err
		}
		if err := validateSHA256("module .zip "+module.Module, module.ZipSHA256); err != nil {
			return err
		}
	}
	return nil
}

func validateBuild(raw json.RawMessage, build Build) error {
	if _, err := exactObject(raw, "command", "cgo_enabled", "gcflags", "ldflags", "tags"); err != nil {
		return fmt.Errorf("manifest: build: %w", err)
	}
	return validateBuildValue(build)
}

func validateBuildValue(build Build) error {
	if strings.TrimSpace(build.Command) == "" {
		return errors.New("manifest: build command must be non-empty")
	}
	if build.Tags == nil {
		return errors.New("manifest: build tags must be an array")
	}
	previous := ""
	for _, tag := range build.Tags {
		if tag == "" || tag <= previous {
			return fmt.Errorf("manifest: build tags must be sorted and unique at %q", tag)
		}
		previous = tag
		if strings.ContainsAny(tag, " \t\r\n") {
			return fmt.Errorf("manifest: build tag %q contains whitespace", tag)
		}
	}
	return nil
}

func validateRevision(label, value string) error {
	if !gitRevisionRE.MatchString(value) {
		return fmt.Errorf("manifest: %s must be a lowercase 40-hex Git revision", label)
	}
	return nil
}

func validateSHA256(label, value string) error {
	if !sha256RE.MatchString(value) {
		return fmt.Errorf("manifest: %s must be a lowercase 64-hex SHA-256", label)
	}
	return nil
}

func validateRelativePath(value string) error {
	if value == "" || strings.Contains(value, "\\") || strings.HasPrefix(value, "/") {
		return fmt.Errorf("manifest: unsafe relative path %q", value)
	}
	clean := path.Clean(value)
	if clean != value || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("manifest: unsafe relative path %q", value)
	}
	return nil
}

func exactObject(raw json.RawMessage, required ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, errors.New("manifest: expected JSON object")
	}
	expected := make(map[string]struct{}, len(required))
	for _, field := range required {
		expected[field] = struct{}{}
		value, ok := object[field]
		if !ok {
			return nil, fmt.Errorf("manifest: missing field %q", field)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("manifest: field %q must not be null", field)
		}
	}
	for field := range object {
		if _, ok := expected[field]; !ok {
			return nil, fmt.Errorf("manifest: unknown field %q", field)
		}
	}
	return object, nil
}

func rawField(raw json.RawMessage, field string) json.RawMessage {
	var object map[string]json.RawMessage
	_ = json.Unmarshal(raw, &object)
	return object[field]
}
