---
schema: vera.verdict.v1
verdict_id: mvp-3-spec-review-round1-20261120
status: NEEDS_WORK
reviewed_commit: a793dd9629692b7f1b96278f0f587ff4f8a6adac
findings:
  - finding_id: mvp-3-spec-review-round1-20261120-F1
    severity: HIGH
  - finding_id: mvp-3-spec-review-round1-20261120-F2
    severity: MED
  - finding_id: mvp-3-spec-review-round1-20261120-F3
    severity: MED
  - finding_id: mvp-3-spec-review-round1-20261120-F4
    severity: MED
  - finding_id: mvp-3-spec-review-round1-20261120-F5
    severity: MED
  - finding_id: mvp-3-spec-review-round1-20261120-F6
    severity: MED
  - finding_id: mvp-3-spec-review-round1-20261120-F7
    severity: LOW
  - finding_id: mvp-3-spec-review-round1-20261120-F8
    severity: LOW
  - finding_id: mvp-3-spec-review-round1-20261120-F9
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-3-spec-review-round1-20261120.md
artifact_sha: 655f174f835b3d7107b5d0e06c597aa6fa70a855afbf3b05072440812147c048
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet), spec round 1 of 2 (Law 11) on
`docs/plans/mvp-3-reviewer-signed-spec.md` draft 2, under `CI-mvp-reviewer-sign-5tqv2c`.

## Findings (all answered in draft 3 §9)
- **F1 HIGH.** Reviewer identity compared as raw strings in the rule but by slug for keys: `Alice` could
  review `alice`'s requirement with alice's key; empty slugs collide. → identity is `KeyName`.
- **F2 MED.** Contradicted W-INV-2; VALID silently tightened for existing files; exit 2 overloaded. →
  W-INV-2 amended, exit 3, compatibility paragraph.
- **F3 MED.** Present-signature-without-declared-key undefined; INVALID/UNREVIEWED precedence undefined.
- **F4 MED.** Cosign preconditions inconsistent; no post-append check; repeat undefined.
- **F5 MED.** Cosignature parsing: uppercase hex duplicates, `[]`/`null`, no cap, bystander keys.
- **F6 MED.** Reviewer keys not required distinct from owner's or each other's.
- **F7-F9 LOW.** Replay scope, cosignature not binding the exporter key; key-loading edge cases;
  invariants too coarse. Taken where cheap.
