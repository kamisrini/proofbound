# P7-SPEC-ADOPT-1 — round 1 independent review verdict

**Reviewer identity:** independent non-author reviewer

**Review role:** live SPEC/API adoption reviewer

**Review date:** 2026-10-24

**Reviewed commit:** `7d1c81debb66b2d03300c079f68e9a11bbd50e39`

**Reviewed files:**

- `kernel/internal/twin/SPEC.md` — SHA-256 `54da96fe0150854a67d56a05cc272f412eedc37144ae009e94ca3db1dde31630`
- `docs/intent/records/change-intents/CI-p7-spec-adopt-1a2b3c/0001.md` — SHA-256 `535d5d3384285747f401c76e20a9ec71f02b0120896fb107248b2583369805ec`

## VERDICT: REJECT

The SPEC wording matches the authorized slice: staged executable binding is adopted, launched-
process observation is explicitly future/unadopted, and no runtime claim is made.

The authorization record's native self-digest is invalid. Replacing only its declared
`artifact_sha256` with 64 zeroes, as required by the records contract, produces
`5725e5795368609d886b77341e1465ca205eb7ac2cb3bf565b6076a08093fe9a`, but the record declares
`fa7a1715910edb7bd3fa4c70c289c1b03776578d571aafaab3486312ec34b476`. The authorization cannot be
cryptographically accepted until corrected and freshly reviewed.

## Scope limits

No production-code changes, runtime fixtures, binary validation, PostgreSQL approval, no-fetch or
cleanup witnesses, L12/L17, behavior-lock acceptance, clean-clone proof, or runtime verdict were
reviewed or claimed. `DEF-P7-001` remains dormant and the pinned PostgreSQL JAR remains unapproved.
