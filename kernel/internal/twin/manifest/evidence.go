package manifest

import (
	"errors"
	"fmt"
)

// CommittedEvidenceReader supplies one already-committed regular file at an explicit revision.
// Implementations own the committed-tree boundary; ResolveInventory never reads the filesystem,
// invokes Git, fetches, executes, or resolves external artifacts itself.
type CommittedEvidenceReader interface {
	ReadCommittedFile(revision, path string) ([]byte, error)
}

// EvidenceResolution is the verified result for one inventory field. UNVERIFIED fields are
// preserved without a reader call or evidence claim.
type EvidenceResolution struct {
	Field            string
	State            InventoryState
	EvidencePath     string
	EvidenceRevision string
	EvidenceSHA256   string
	Reason           string
}

// EvidenceResolutionPending is permitted only as an explicit evaluator input; inventory
// declarations themselves remain limited to BOUND or UNVERIFIED.
const EvidenceResolutionPending InventoryState = "PENDING"

// EvidenceClosureState is the pure evaluator state for one required evidence field or the
// aggregate closure result.
type EvidenceClosureState string

const (
	EvidenceClosureBound      EvidenceClosureState = "BOUND"
	EvidenceClosureUnverified EvidenceClosureState = "UNVERIFIED"
	EvidenceClosurePending    EvidenceClosureState = "PENDING"
)

const EvidenceClosureSchema = "proofbound.p7.evidence-closure.v1"

// EvidenceClosure is the deterministic non-runtime completeness result. It is not a build or
// runtime acceptance claim.
type EvidenceClosure struct {
	Schema           string                 `json:"schema"`
	EvidenceRevision string                 `json:"evidence_revision"`
	Overall          EvidenceClosureState   `json:"overall"`
	Fields           []EvidenceClosureField `json:"fields"`
}

// EvidenceClosureField is one evaluated required evidence field.
type EvidenceClosureField struct {
	Field            string               `json:"field"`
	State            EvidenceClosureState `json:"state"`
	EvidencePath     string               `json:"evidence_path,omitempty"`
	EvidenceRevision string               `json:"evidence_revision,omitempty"`
	EvidenceSHA256   string               `json:"evidence_sha256,omitempty"`
	Reason           string               `json:"reason,omitempty"`
}

// ResolveInventory verifies BOUND inventory declarations against caller-supplied committed-tree
// bytes and preserves explicit UNVERIFIED fields. It performs no external operation itself.
func ResolveInventory(inventory EvidenceInventory, manifest Manifest, reader CommittedEvidenceReader) ([]EvidenceResolution, error) {
	if err := ValidateInventory(inventory, manifest); err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, fmt.Errorf("inventory: committed evidence reader is required")
	}
	resolved := make([]EvidenceResolution, 0, len(inventory.Fields))
	for _, field := range inventory.Fields {
		if field.State == InventoryUnverified {
			resolved = append(resolved, EvidenceResolution{Field: field.Field, State: InventoryUnverified, Reason: "inventory explicitly UNVERIFIED"})
			continue
		}
		bytes, err := reader.ReadCommittedFile(field.EvidenceRevision, field.EvidencePath)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: committed evidence %s: %w", field.Field, field.EvidencePath, err)
		}
		got := digestBytes(bytes)
		if got != field.EvidenceSHA256 {
			return nil, fmt.Errorf("inventory: %s: evidence digest mismatch: got %q want %q", field.Field, got, field.EvidenceSHA256)
		}
		resolved = append(resolved, EvidenceResolution{
			Field:            field.Field,
			State:            InventoryBound,
			EvidencePath:     field.EvidencePath,
			EvidenceRevision: field.EvidenceRevision,
			EvidenceSHA256:   field.EvidenceSHA256,
		})
	}
	return resolved, nil
}

// EvaluateEvidenceClosure classifies every required inventory field without performing I/O. The
// input resolution list must have exact field coverage. A pending or unverified field prevents an
// overall BOUND result; missing, extra, contradictory, stale, or malformed inputs fail closed.
func EvaluateEvidenceClosure(inventory EvidenceInventory, manifest Manifest, resolutions []EvidenceResolution) (EvidenceClosure, error) {
	if err := ValidateInventory(inventory, manifest); err != nil {
		return EvidenceClosure{}, err
	}
	if len(resolutions) != len(inventoryFields) {
		return EvidenceClosure{}, fmt.Errorf("closure: expected %d resolutions, got %d", len(inventoryFields), len(resolutions))
	}
	closure := EvidenceClosure{
		Schema:           EvidenceClosureSchema,
		EvidenceRevision: inventory.EvidenceRevision,
		Fields:           make([]EvidenceClosureField, 0, len(resolutions)),
		Overall:          EvidenceClosureBound,
	}
	for i, resolution := range resolutions {
		wantField := inventoryFields[i]
		if resolution.Field != wantField {
			return EvidenceClosure{}, fmt.Errorf("closure: field %d must be %q, got %q", i, wantField, resolution.Field)
		}
		if i > 0 && resolution.Field <= resolutions[i-1].Field {
			return EvidenceClosure{}, fmt.Errorf("closure: fields must be sorted and unique at %q", resolution.Field)
		}
		inventoryField := inventory.Fields[i]
		field, err := evaluateClosureField(inventory, inventoryField, resolution)
		if err != nil {
			return EvidenceClosure{}, fmt.Errorf("closure: %s: %w", resolution.Field, err)
		}
		closure.Fields = append(closure.Fields, field)
		switch field.State {
		case EvidenceClosurePending:
			closure.Overall = EvidenceClosurePending
		case EvidenceClosureUnverified:
			if closure.Overall == EvidenceClosureBound {
				closure.Overall = EvidenceClosureUnverified
			}
		}
	}
	return closure, nil
}

func evaluateClosureField(inventory EvidenceInventory, inventoryField InventoryField, resolution EvidenceResolution) (EvidenceClosureField, error) {
	switch resolution.State {
	case InventoryBound:
		if inventoryField.State != InventoryBound {
			return EvidenceClosureField{}, fmt.Errorf("contradictory BOUND resolution for %s inventory", inventoryField.State)
		}
		if err := validateEvidencePath(resolution.EvidencePath); err != nil {
			return EvidenceClosureField{}, err
		}
		if resolution.EvidenceRevision != inventory.EvidenceRevision {
			return EvidenceClosureField{}, fmt.Errorf("stale evidence revision %q", resolution.EvidenceRevision)
		}
		if err := validateSHA256("evidence", resolution.EvidenceSHA256); err != nil {
			return EvidenceClosureField{}, err
		}
		if resolution.Reason != "" {
			return EvidenceClosureField{}, errors.New("BOUND resolution must not have a reason")
		}
		return EvidenceClosureField{
			Field:            resolution.Field,
			State:            EvidenceClosureBound,
			EvidencePath:     resolution.EvidencePath,
			EvidenceRevision: resolution.EvidenceRevision,
			EvidenceSHA256:   resolution.EvidenceSHA256,
		}, nil
	case InventoryUnverified:
		if inventoryField.State != InventoryUnverified {
			return EvidenceClosureField{}, fmt.Errorf("contradictory UNVERIFIED resolution for %s inventory", inventoryField.State)
		}
		if resolution.EvidencePath != "" || resolution.EvidenceRevision != "" || resolution.EvidenceSHA256 != "" {
			return EvidenceClosureField{}, errors.New("UNVERIFIED resolution must not declare evidence")
		}
		if resolution.Reason == "" {
			return EvidenceClosureField{}, errors.New("UNVERIFIED resolution requires a reason")
		}
		return EvidenceClosureField{Field: resolution.Field, State: EvidenceClosureUnverified, Reason: resolution.Reason}, nil
	case EvidenceResolutionPending:
		if inventoryField.State != InventoryBound {
			return EvidenceClosureField{}, fmt.Errorf("contradictory PENDING resolution for %s inventory", inventoryField.State)
		}
		if resolution.EvidencePath != "" || resolution.EvidenceRevision != "" || resolution.EvidenceSHA256 != "" {
			return EvidenceClosureField{}, errors.New("PENDING resolution must not declare evidence")
		}
		if resolution.Reason == "" {
			return EvidenceClosureField{}, errors.New("PENDING resolution requires a reason")
		}
		return EvidenceClosureField{Field: resolution.Field, State: EvidenceClosurePending, Reason: resolution.Reason}, nil
	default:
		return EvidenceClosureField{}, fmt.Errorf("invalid resolution state %q", resolution.State)
	}
}
