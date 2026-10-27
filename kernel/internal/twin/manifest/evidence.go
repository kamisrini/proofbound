package manifest

import (
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
			resolved = append(resolved, EvidenceResolution{Field: field.Field, State: InventoryUnverified})
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
