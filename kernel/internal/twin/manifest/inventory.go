package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

const InventorySchema = "proofbound.p7.source-toolchain-inventory.v1"

// InventoryState records whether one manifest field has an exact committed evidence path.
type InventoryState string

const (
	InventoryBound      InventoryState = "BOUND"
	InventoryUnverified InventoryState = "UNVERIFIED"
)

// EvidenceInventory maps every manifest field to committed evidence or an explicit
// UNVERIFIED state. It does not read, retrieve, build, or execute any referenced artifact.
type EvidenceInventory struct {
	Schema           string           `json:"schema"`
	EvidenceRevision string           `json:"evidence_revision"`
	ManifestDigest   string           `json:"manifest_digest"`
	Fields           []InventoryField `json:"fields"`
}

// InventoryField is one exact manifest-field evidence declaration.
type InventoryField struct {
	Field            string         `json:"field"`
	State            InventoryState `json:"state"`
	EvidencePath     string         `json:"evidence_path"`
	EvidenceRevision string         `json:"evidence_revision"`
	EvidenceSHA256   string         `json:"evidence_sha256"`
}

var inventoryFields = []string{
	"build.cgo_enabled",
	"build.command",
	"build.gcflags",
	"build.ldflags",
	"build.tags",
	"evidence_revision",
	"modules",
	"projector_revision",
	"schema",
	"source.files",
	"toolchain.archive_format",
	"toolchain.archive_sha256",
	"toolchain.go_version",
	"toolchain.goarch",
	"toolchain.goos",
}

// ParseInventory validates strict canonical JSON against a manifest and returns its typed inventory.
func ParseInventory(raw []byte, manifest Manifest) (EvidenceInventory, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return EvidenceInventory{}, fmt.Errorf("inventory: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return EvidenceInventory{}, errors.New("inventory: JSON is not canonical")
	}
	top, err := exactObject(raw, "schema", "evidence_revision", "manifest_digest", "fields")
	if err != nil {
		return EvidenceInventory{}, err
	}
	var inventory EvidenceInventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		return EvidenceInventory{}, fmt.Errorf("inventory: decode: %w", err)
	}
	var fields []json.RawMessage
	if err := json.Unmarshal(top["fields"], &fields); err != nil {
		return EvidenceInventory{}, fmt.Errorf("inventory: fields: %w", err)
	}
	for i, field := range fields {
		if _, err := exactObject(field, "field", "state", "evidence_path", "evidence_revision", "evidence_sha256"); err != nil {
			return EvidenceInventory{}, fmt.Errorf("inventory: field %d: %w", i, err)
		}
	}
	if err := ValidateInventory(inventory, manifest); err != nil {
		return EvidenceInventory{}, err
	}
	return inventory, nil
}

// CanonicalInventoryBytes validates and encodes an inventory as RFC 8785 canonical JSON.
func CanonicalInventoryBytes(inventory EvidenceInventory, manifest Manifest) ([]byte, error) {
	if err := ValidateInventory(inventory, manifest); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(inventory)
	if err != nil {
		return nil, fmt.Errorf("inventory: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("inventory: canonical encode: %w", err)
	}
	return canonical, nil
}

// ValidateInventory validates exact field coverage and declared evidence identities.
// It compares only caller-supplied declarations and the canonical manifest digest.
func ValidateInventory(inventory EvidenceInventory, manifest Manifest) error {
	if inventory.Schema != InventorySchema {
		return fmt.Errorf("inventory: schema must be %q", InventorySchema)
	}
	if err := validateRevision("inventory evidence_revision", inventory.EvidenceRevision); err != nil {
		return err
	}
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		return fmt.Errorf("inventory: manifest: %w", err)
	}
	wantManifestDigest := digestBytes(manifestBytes)
	if inventory.ManifestDigest != wantManifestDigest {
		return fmt.Errorf("inventory: manifest digest mismatch: got %q want %q", inventory.ManifestDigest, wantManifestDigest)
	}
	if len(inventory.Fields) != len(inventoryFields) {
		return fmt.Errorf("inventory: expected %d fields, got %d", len(inventoryFields), len(inventory.Fields))
	}
	for i, field := range inventory.Fields {
		wantField := inventoryFields[i]
		if field.Field != wantField {
			return fmt.Errorf("inventory: field %d must be %q, got %q", i, wantField, field.Field)
		}
		if i > 0 && field.Field <= inventory.Fields[i-1].Field {
			return fmt.Errorf("inventory: fields must be sorted and unique at %q", field.Field)
		}
		if err := validateInventoryField(inventory, field); err != nil {
			return fmt.Errorf("inventory: %s: %w", field.Field, err)
		}
	}
	return nil
}

func validateInventoryField(inventory EvidenceInventory, field InventoryField) error {
	switch field.State {
	case InventoryBound:
		if err := validateEvidencePath(field.EvidencePath); err != nil {
			return err
		}
		if field.EvidenceRevision != inventory.EvidenceRevision {
			return fmt.Errorf("stale evidence revision %q", field.EvidenceRevision)
		}
		if err := validateSHA256("evidence", field.EvidenceSHA256); err != nil {
			return err
		}
	case InventoryUnverified:
		if field.EvidencePath != "" || field.EvidenceRevision != "" || field.EvidenceSHA256 != "" {
			return errors.New("UNVERIFIED field must not declare evidence")
		}
	default:
		return fmt.Errorf("invalid state %q", field.State)
	}
	return nil
}

func validateEvidencePath(value string) error {
	if value == "" || strings.ContainsRune(value, '\x00') {
		return errors.New("BOUND field requires a committed evidence path")
	}
	return validateRelativePath(value)
}
