## MODIFIED Requirements

### Requirement: Rename Format
Delta specs SHALL express a rename as an explicit FROM/TO pair.

#### Scenario: Documenting a rename
- **WHEN** an author renames a requirement
- **AND** specify both old and new names explicitly:
  ```markdown
  ## RENAMED Requirements
  - FROM: `### Requirement: Old Name`
  - TO: `### Requirement: New Name`
  ```
- **THEN** the archive applies the rename
