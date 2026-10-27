package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

const EvidenceManifestSchema = "proofbound.p7.evidence-manifest.v1"

// EvidenceManifest is a deterministic documentary skeleton over accepted manifest, inventory,
// resolution, and closure inputs. It does not claim complete build or runtime evidence.
type EvidenceManifest struct {
	Schema           string                  `json:"schema"`
	EvidenceRevision string                  `json:"evidence_revision"`
	ManifestDigest   string                  `json:"manifest_digest"`
	InventoryDigest  string                  `json:"inventory_digest"`
	ClosureOverall   EvidenceClosureState    `json:"closure_overall"`
	Fields           []EvidenceManifestField `json:"fields"`
}

// EvidenceManifestField is one closure status row in the documentary skeleton.
type EvidenceManifestField struct {
	Field            string               `json:"field"`
	State            EvidenceClosureState `json:"state"`
	EvidencePath     string               `json:"evidence_path"`
	EvidenceRevision string               `json:"evidence_revision"`
	EvidenceSHA256   string               `json:"evidence_sha256"`
	Reason           string               `json:"reason"`
}

// BuildEvidenceManifest derives a skeleton and rejects any supplied closure that does not exactly
// match the accepted inventory and resolution inputs.
func BuildEvidenceManifest(manifest Manifest, inventory EvidenceInventory, resolutions []EvidenceResolution, closure EvidenceClosure) (EvidenceManifest, error) {
	derived, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		return EvidenceManifest{}, err
	}
	if !reflect.DeepEqual(closure, derived) {
		return EvidenceManifest{}, errors.New("skeleton: supplied closure does not match derived closure")
	}
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		return EvidenceManifest{}, err
	}
	inventoryBytes, err := CanonicalInventoryBytes(inventory, manifest)
	if err != nil {
		return EvidenceManifest{}, err
	}
	skeleton := EvidenceManifest{
		Schema:           EvidenceManifestSchema,
		EvidenceRevision: inventory.EvidenceRevision,
		ManifestDigest:   digestBytes(manifestBytes),
		InventoryDigest:  digestBytes(inventoryBytes),
		ClosureOverall:   closure.Overall,
		Fields:           make([]EvidenceManifestField, 0, len(closure.Fields)),
	}
	for _, field := range closure.Fields {
		skeleton.Fields = append(skeleton.Fields, EvidenceManifestField{
			Field:            field.Field,
			State:            field.State,
			EvidencePath:     field.EvidencePath,
			EvidenceRevision: field.EvidenceRevision,
			EvidenceSHA256:   field.EvidenceSHA256,
			Reason:           field.Reason,
		})
	}
	if err := ValidateEvidenceManifest(skeleton, manifest, inventory, resolutions, closure); err != nil {
		return EvidenceManifest{}, err
	}
	return skeleton, nil
}

// ValidateEvidenceManifest validates the skeleton and its exact derivation inputs.
func ValidateEvidenceManifest(skeleton EvidenceManifest, manifest Manifest, inventory EvidenceInventory, resolutions []EvidenceResolution, closure EvidenceClosure) error {
	derived, err := EvaluateEvidenceClosure(inventory, manifest, resolutions)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(closure, derived) {
		return errors.New("skeleton: closure does not match derivation inputs")
	}
	if skeleton.Schema != EvidenceManifestSchema {
		return fmt.Errorf("skeleton: schema must be %q", EvidenceManifestSchema)
	}
	if skeleton.EvidenceRevision != inventory.EvidenceRevision {
		return fmt.Errorf("skeleton: stale evidence revision %q", skeleton.EvidenceRevision)
	}
	manifestBytes, err := CanonicalBytes(manifest)
	if err != nil {
		return fmt.Errorf("skeleton: manifest: %w", err)
	}
	if skeleton.ManifestDigest != digestBytes(manifestBytes) {
		return errors.New("skeleton: manifest digest mismatch")
	}
	inventoryBytes, err := CanonicalInventoryBytes(inventory, manifest)
	if err != nil {
		return err
	}
	if skeleton.InventoryDigest != digestBytes(inventoryBytes) {
		return errors.New("skeleton: inventory digest mismatch")
	}
	if skeleton.ClosureOverall != closure.Overall {
		return errors.New("skeleton: closure overall mismatch")
	}
	if len(skeleton.Fields) != len(derived.Fields) {
		return fmt.Errorf("skeleton: expected %d fields, got %d", len(derived.Fields), len(skeleton.Fields))
	}
	for i, field := range skeleton.Fields {
		want := derived.Fields[i]
		if field.Field != want.Field {
			return fmt.Errorf("skeleton: field %d must be %q, got %q", i, want.Field, field.Field)
		}
		if i > 0 && field.Field <= skeleton.Fields[i-1].Field {
			return fmt.Errorf("skeleton: fields must be sorted and unique at %q", field.Field)
		}
		if field.State != want.State || field.EvidencePath != want.EvidencePath || field.EvidenceRevision != want.EvidenceRevision || field.EvidenceSHA256 != want.EvidenceSHA256 || field.Reason != want.Reason {
			return fmt.Errorf("skeleton: field %q does not match closure", field.Field)
		}
	}
	return nil
}

// ParseEvidenceManifest validates strict canonical JSON and exact derivation inputs.
func ParseEvidenceManifest(raw []byte, manifest Manifest, inventory EvidenceInventory, resolutions []EvidenceResolution, closure EvidenceClosure) (EvidenceManifest, error) {
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return EvidenceManifest{}, fmt.Errorf("skeleton: canonical JSON: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return EvidenceManifest{}, errors.New("skeleton: JSON is not canonical")
	}
	top, err := exactObject(raw, "schema", "evidence_revision", "manifest_digest", "inventory_digest", "closure_overall", "fields")
	if err != nil {
		return EvidenceManifest{}, err
	}
	var skeleton EvidenceManifest
	if err := json.Unmarshal(raw, &skeleton); err != nil {
		return EvidenceManifest{}, fmt.Errorf("skeleton: decode: %w", err)
	}
	var fields []json.RawMessage
	if err := json.Unmarshal(top["fields"], &fields); err != nil {
		return EvidenceManifest{}, fmt.Errorf("skeleton: fields: %w", err)
	}
	for i, field := range fields {
		if _, err := exactObject(field, "field", "state", "evidence_path", "evidence_revision", "evidence_sha256", "reason"); err != nil {
			return EvidenceManifest{}, fmt.Errorf("skeleton: field %d: %w", i, err)
		}
	}
	if err := ValidateEvidenceManifest(skeleton, manifest, inventory, resolutions, closure); err != nil {
		return EvidenceManifest{}, err
	}
	return skeleton, nil
}

// CanonicalEvidenceManifestBytes validates and encodes the skeleton as RFC 8785 canonical JSON.
func CanonicalEvidenceManifestBytes(skeleton EvidenceManifest, manifest Manifest, inventory EvidenceInventory, resolutions []EvidenceResolution, closure EvidenceClosure) ([]byte, error) {
	if err := ValidateEvidenceManifest(skeleton, manifest, inventory, resolutions, closure); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(skeleton)
	if err != nil {
		return nil, fmt.Errorf("skeleton: encode: %w", err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("skeleton: canonical encode: %w", err)
	}
	return canonical, nil
}
