---
schema: vera.verdict.v1
verdict_id: p7-projections-vd-citation-round1-20261111
status: ACCEPTABLE
reviewed_commit: 83c6157d45b64005014330c8e29a675ab8110e56
findings: []
artifact_path: docs/verification/verdicts/p7-projections-vd-citation-round1-20261111.md
artifact_sha: b06f572473e889057daef6f30b6cd8831212bae2443da512e7863326131890b9
---

# `internal/projections` date-suffixed VD citation fix — non-author verdict (round 1)

**ACCEPTABLE.** Independent non-author review of change intent `CI-fix-vd-citation-9958b1`
(obligation `BR-self-hosting-correctness-4607a3` O-1): the commit projection's `decisionPattern`
was broadened to accept a terminal `YYYY-MM-DD` phase-decision suffix in addition to the random
`[a-z0-9]{6}` token, with a new non-integration regression guard `citation_test.go`.

## Reviewer identity and Law 9 standing

Reviewed by this session (2026-11-11). The implementation (`projections.go`, `citation_test.go`)
was authored 2026-11-10 in commit `ac76f888782fdef460ee331218583b085376c52e`; **this session did
not author it**, so this is an **accepting** non-author verdict under Law 9, not a self pre-check.
The owner remains the final acceptor.

## HEAD under review

Reviewed at HEAD `83c6157d45b64005014330c8e29a675ab8110e56`. The two governed implementation files
(`kernel/internal/projections/projections.go`, `kernel/internal/projections/citation_test.go`) are
**byte-identical** between the implementation commit `ac76f88` and this HEAD (`git diff` empty); the
only intervening changes were the `@generated` route-matrix pin refresh (`9c54675`) and `notes/state.md`.
This verdict therefore covers `CI-fix-vd-citation-9958b1` exactly.

## What was verified

**Spec truth — P-INV-7 (malformed citations fail closed).** `commitPayload.validate()`
(`projections.go:715`) rejects any `cited_decisions` entry not matching `decisionPattern`. The new
pattern `^VD-[a-z0-9]+(?:-[a-z0-9]+)*-(?:[a-z0-9]{6}|[0-9]{4}-[0-9]{2}-[0-9]{2})$` accepts both
`/proofbound-decide` suffix conventions and still fails closed on malformed shapes. The intent-ref
`BD/BR/CI` ID patterns were untouched (they correctly remain 6-char-token-only; intent records never
carry a date suffix). SPEC §3 was updated to record the two accepted citation forms.

**Test discrimination (the fixture does not hold the key variable constant).** In
`TestCommitValidationAcceptsBothDecisionSuffixForms`, `base()` fills every other required field with
a valid value, so the cited-decision string is the only variable that can flip `validate()`. The
test both accepts the real date-suffixed IDs that previously crashed the apply and rejects four
malformed shapes.

**Calibrated focused mutation sweep (changed surface, make-check harness).** `internal/projections`
is validated under `make check` only by its non-integration tests (`citation_test.go`); the rest of
the package is integration-tagged, so a full-package AST sweep against the make-check harness would
report out-of-scope survivors in DB code this delta does not touch. The proportionate evidence for a
one-line regex broadening (connector/peripheral tier) is a focused calibrated sweep on the exact
changed surface. Calibration was declared before any count:

| Control | Mutation | Required | Observed |
|---|---|---|---|
| Baseline | none | pass | pass |
| A (positive) | drop the date branch (pre-fix regex) | die | FAIL (died) |
| B (positive) | token quantifier `{6}` → `{7}` | die | FAIL (died) |
| C (positive) | alternation → `.*` (weaken P-INV-7) | die | FAIL (died) |
| D (positive) | negate the guard (`!match` → `match`) | die | FAIL (died) |
| E (neutral) | swap alternation branch order (semantically equivalent) | survive | survive |
| F (lethal) | unconditional `t.Fatalf` in the test | killed | FAIL (killed) |

Positive controls 4/4 died; the neutral (declared-equivalent) control survived; the lethal
injected-failing-test control was killed, proving the harness actually runs the assertions.
`invalid` (uncompilable) mutants were not produced and none were counted. **Zero undeclared
survivors on the changed surface.** (Two controls initially reported a cached pass from a
no-op `perl` substitution; corrected with `-count=1` and confirmed the mutation was applied before
re-running — recorded here so the count is not believed on a stale harness.)

**Build Laws.**
- **Law 1 (derived state):** touching `kernel/internal/projections/` re-stales the `@generated`
  `p6-event-route-matrix.md`; it was refreshed in `9c54675` and `scripts/p6-route-matrix.sh --check`
  is clean at HEAD (exit 0).
- **Law 6 (spec before code):** no new invariant ID (the rule stays under the existing P-INV-7); the
  SPEC note is a documentation delta on an existing invariant, which is the correct treatment for a
  bug-fix delta.
- **Law 7 (lessons compile to enforcement):** the root-cause meta-lesson (synthetic fixtures used
  only 6-char IDs, so `make check` was green while the product could not process its own record) is
  compiled into a **non-integration** regression test that runs under `make check` — exactly the gap
  that let the defect ship.
- **Law 2 (one home):** no datum copied; the SPEC prose describes code behavior, it does not become a
  second source of truth.

## Non-blocking note for the author (advisory, not a defect)

The date branch `[0-9]{4}-[0-9]{2}-[0-9]{2}` is a *shape* guard: it accepts structurally-well-formed
but calendar-invalid dates (e.g. `VD-foo-2026-13-45`). This does **not** break P-INV-7 and is not a
regression — the validator was never a semantic/existence check (the pre-fix `[a-z0-9]{6}` token
likewise accepts meaningless tokens, and cross-reference existence is not this validator's job), so
the attack surface is unchanged. Recorded only as a possible future hardening, not a blocker; the
delivered obligation O-1 is met.

## Verdict

ACCEPTABLE. The fix is minimal and correct, P-INV-7 fail-closed semantics are preserved for malformed
citations, the regression guard discriminates the exact changed property under the make-check harness,
and the Build Laws hold. `CI-fix-vd-citation-9958b1` is non-author accepted on this revision; the
owner remains the final acceptor.
