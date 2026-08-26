---
schema: vera.verdict.v1
verdict_id: task5-current-round7
status: ACCEPTABLE
reviewed_commit: 517470e4b5672c8925c4f79f1325ee4ee03a0541
findings: []
artifact_path: docs/verification/verdicts/task5-current-round7.md
artifact_sha: ee2217b9cfc3611a48aa58ccf9c94e7eed6a1044380ff39cc50ed2ca2fa35db0
---

ACCEPTABLE

## Scope and calibration

Independent non-author review of frozen Task 5 remediation commit `517470e`.

Reviewed:

- `CLAUDE.md`
- `notes/state.md`
- `notes/journal/2026-08-24.md`
- Prior Task 5 verdicts, Rounds 1–6
- `internal/connector/checks/SPEC.md`
- Emitter, connector, CLI, invariant, and test code
- Frozen remediation diff

The tracked worktree remained unmodified and clean.

Verification completed:

- `go test -race ./... -count=1`: passed
- Bare `make check`: passed
- Focused emitter, connector-binding, publication, helper-failure, and malformed-helper tests: passed
- New truncated `od`, malformed `od`, short entropy, impossible timestamp, and backward timestamp fixtures: passed
- Invariant numbering/table checks: passed
- `git diff --check`: passed

The mutation sweep calibrated successfully and began killing mutants but was stopped before completion at the user’s request. Mutation results are therefore not used as independent acceptance evidence.

## Findings

### HIGH

None.

### MED

None.

### LOW

None.

## Harm/routes closure

The governing harm is that incomplete, failed, differently sourced, or byte-collapsed gate observations could become durable false evidence in the append-only ledger.

All previously open Round 6 routes are closed:

- Truncated non-empty `od` output is rejected unless it contains exactly the expected complete byte count.
- Malformed, negative, non-decimal, and out-of-range byte tokens are rejected.
- Entropy requires exactly 16 valid decimal bytes.
- Impossible calendar timestamps are rejected.
- Backward timestamps are rejected.
- Negative measured duration is rejected.
- Valid malformed-helper failures publish no witness.
- Normal helper failures and successful-empty helper output remain loud.
- The real strict connector continues to reject malformed evidence.
- Existing repository binding, sanitized Git environment, byte admission, publication failure, cleanup, identity, replay, and spool-preservation routes remain closed.

## Invariant audit

C-INV-25 is appended after C-INV-24 without renumbering or modifying prior invariants.

The invariant lock and table citation are consistent:

`emitter_test.go::TestEmitter_MalformedHelperOutputIsLoud`

The implementation now enforces:

- Complete source-byte coverage
- Decimal byte parsing
- Byte range `0..255`
- Exact 16-byte entropy
- Calendar-valid timestamps
- Nondecreasing timestamp order
- Nonnegative duration

The proving tests cover each previously open malformed-helper class.

## Verified strengths

- The Round 6 HIGH route no longer permits NUL deletion and rewritten evidence.
- Short entropy can no longer generate an invalid short ULID.
- Impossible and backward timestamps fail before publication.
- Ordinary success and failure witnesses preserve the actual gate status.
- Combined output hashing remains exact.
- Repository and gate identity remain bound to one root.
- Inadmissible bytes fail before gate execution.
- Publication and cleanup failure routes remain fail-closed.
- Connector strictness, content identity, replay, malformed stopping, and spool preservation remain green.
- Full race tests and bare `make check` pass.

## Residual limits

- Shell behavior was tested on Linux Bash/GNU tools, not every BSD/macOS implementation.
- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool files remain local and destructible before ingestion.
- Concurrent mutation of captured output was not exhaustively modeled.
- Helper descendants and command timeouts remain outside the tested scope.
- Mutation operators exclude `_test.go`; the current sweep was not completed independently.

## Acceptance rationale

Task 5 is acceptable under Law 9 for frozen remediation commit `517470e`.

The three open Round 6 routes were independently reproduced as closed: truncated byte scans, incomplete entropy, and semantically invalid or backward timestamps now fail closed before witness publication. The exact malformed-helper tests pass, and the broader race, gate, connector, repository-binding, publication, and invariant checks remain green.

No HIGH, MED, or LOW findings remain.
