---
schema: vera.verdict.v1
verdict_id: r3-warranty-spec-review-20261119
status: NEEDS_WORK
reviewed_commit: 254983a0c29b8805f37aeea6385f19f033cc1513
findings:
  - finding_id: r3-warranty-spec-review-20261119-F1
    severity: HIGH
  - finding_id: r3-warranty-spec-review-20261119-F2
    severity: HIGH
  - finding_id: r3-warranty-spec-review-20261119-F3
    severity: HIGH
  - finding_id: r3-warranty-spec-review-20261119-F4
    severity: MED
  - finding_id: r3-warranty-spec-review-20261119-F5
    severity: MED
  - finding_id: r3-warranty-spec-review-20261119-F6
    severity: MED
  - finding_id: r3-warranty-spec-review-20261119-F7
    severity: MED
  - finding_id: r3-warranty-spec-review-20261119-F8
    severity: LOW
  - finding_id: r3-warranty-spec-review-20261119-F9
    severity: LOW
artifact_path: docs/verification/verdicts/r3-warranty-spec-review-20261119.md
artifact_sha: b3de3eb1a8fef68bd35388d5876ebc18bcf073b13eac322301e9b3c0f7ed1532
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet) that did not write the spec. Spec
round 1 of 2 (Law 11) on `docs/plans/R3-warranty-export-spec.md` draft 1 at `reviewed_commit`. Read
only: it read the cited code and ran no tests. Findings are condensed from its report. Each keeps its
severity, its harm and its code evidence.

## Status
NEEDS_WORK (3 HIGH, 4 MED, 2 LOW).

## Findings
- **F1 HIGH.** `StateRevision` (sdd `build.go:205`) returns only the change intent. The requirement
  revisions are private to `stateResult.revisions`, so `verify --repo` cannot re-derive the obligation
  statements. The "word for word" payoff has no path.
- **F2 HIGH.** Without `--repo`, a signer can sign fabricated obligations, reviews and verdicts, and
  W-INV-3's internal-consistency check passes them as VALID.
- **F3 HIGH.**
  - The trust anchor is whatever clone the verifier is handed, and it is based on HEAD: a key
    committed after the warranted commit still validates.
  - `export` ("at the warranted commit or later") contradicts W-INV-5 ("at HEAD").
  - The `.pub` format is unspecified, and there is no fingerprint pinning.
- **F4 MED.** A requirement review carries no commit (`v2.go:165`), so the body's "reviewed commit"
  for reviews is wrong. The bindings are undefined.
- **F5 MED.** The readiness rule is not re-runnable from the body alone: it needs the change targets
  and each requirement's `declared_owner`. The owner check lives at ingest (`projections.go:555`), not
  in the gate SQL.
- **F6 MED.** "Exists in the commit history" is vague. Artifacts are committed after the warranted
  commit. The check must use `ArtifactSHA256()`, then re-parse and compare fields.
- **F7 MED.**
  - Replay across commits is not excluded.
  - A lenient parse leaves canonicalization ambiguity.
  - The signature has no domain separator.
  - Ledger-local gate ids are unverifiable.
- **F8 LOW.** Key-file handling at export is unspecified (missing, permissions, symlink), and so is
  the fallback when XDG is unset.
- **F9 LOW.** Demo step 6 needs a second, uncommitted keystore. W-INV-6 needs a forbidden-import list.

## Author's disposition (not part of the verdict)
All findings are answered in draft 2, §10.
