## MODIFIED Requirements

### Requirement: Init Flow
The init command SHALL print a quick checklist after setup.

#### Scenario: Checklist shown
- **WHEN** init completes
- **THEN** the output includes:
```markdown
## TL;DR Quick Checklist
- Run `openspec list`
- Propose your first change
```

### Requirement: Interactive Mode
The init command SHALL prompt for each tool when run interactively.

### Requirement: Success Output
The init command SHALL report every file it created.
