NEEDS_WORK

## Round 6 scope and calibration

Independent adversarial review of frozen Task 5 remediation commit
`98617e45b23819c86932bf99c6f62f3218aae863`, including diff `1e6a0bf..98617e4`.

The tracked implementation remained frozen throughout review. Hostile repositories, helper shims,
and the independent connector-ingestion test existed only under `/tmp`. This verdict file is the
only tracked review output.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` line was the expected index-check
  positive control.
- `GOCACHE=/tmp/proofbound-review-gocache go test -race ./... -count=1`: passed across every kernel
  package.
- Focused emitter tests for ordinary status, helper/empty-helper failures, publication failures,
  repository binding, and inadmissible bytes: passed.
- Production invariant-numbering/table checks and both corresponding self-tests: passed.
- `git diff --check 1e6a0bf..98617e4`: passed.
- C-INV-24 is the sole invariant-lock addition and follows C-INV-23 in the checks namespace.
- Author mutation, DB integration, and real witnessed-run results remain calibration evidence, not
  independent semantic proof. DB integration and mutation were not rerun in this round.

## Round 5 remedy adjudication

1. **Successful-empty `od` byte scan: CLOSED.** The exact prior NUL fixture now exits 1 before the
   gate, publishes no witness, and leaves no temporary file.
2. **Successful-empty entropy output: CLOSED.** The exact prior run-id fixture now exits 1 before
   the gate and publishes no witness.
3. **Successful-empty timestamp output: CLOSED.** The exact prior date fixture now exits 1 before
   the gate and publishes no witness.
4. **Direct-PID TERM cleanup: CLOSED.** The wrapper exits, the tracked direct helper is killed, and
   both version temporary files are removed.
5. **Non-empty truncated/malformed byte-scan output: OPEN.** A single nonzero token is treated as a
   complete scan, allowing NUL deletion and durable rewritten evidence. See HIGH-1.
6. **Non-empty incomplete entropy output: OPEN.** One byte is treated as the requested 16-byte
   entropy observation and produces an invalid short run ID. See MED-1.
7. **Shape-valid but semantically invalid clock output: OPEN.** Impossible or backward timestamps
   pass the shell regex and produce invalid witnesses under status 0. See MED-1.

## Closure against the governing HARM and prior routes

The governing HARM is that a failed, incomplete, differently sourced, or byte-collapsed run can
become durable successful evidence in an append-only ledger, or that the witnessed command can
report success without one valid published observation.

| Route | Verdict | Evidence |
|---|---|---|
| Present-null and strict reader routes | Closed | Reader unchanged; full checks suite passes |
| Caller/root and hostile `GIT_*` binding | Closed | Independent two-repository real-Git fixture kept gate and witness on A |
| HEAD/status observation failure | Closed | Existing pre-gate routes remain green |
| NUL, `0xfe`, `0xff` with real helpers | Closed | Each stops before gate and witness |
| Non-zero and empty-success `od` scan | Closed | Loud before gate, no witness |
| Non-empty truncated `od` scan | **Open** | NUL deleted and actual connector appends rewritten evidence |
| Non-zero and empty-success clock output | Closed | Loud before gate |
| Invalid-shaped clock output | Closed | Shell shape check refuses it |
| Shape-valid impossible/backward clock output | **Open** | Status 0 and invalid witness |
| Non-zero and empty-success entropy output | Closed | Loud before gate |
| Non-empty short/malformed entropy output | **Open** | Status 0 and invalid short run ID |
| mkdir and staged mktemp failure | Closed | Loud before gate, temps cleaned |
| cat, hash exit, and malformed hash failure | Closed | Loud after gate, no witness |
| Serialization I/O/spool loss and mv failure | Closed | Loud, no final witness, temp cleaned |
| Process-group and direct-PID helper interruption | Closed | Direct helper and capture temps removed |
| Ordinary gate status and combined digest | Closed | Valid witness records actual status and exact output digest |
| Actual connector identity after malformed scan | **Open** | Collapsed-NUL emitted witness appended successfully |

## Findings

### HIGH-1 — A non-empty truncated `od` result still disables NUL refusal and produces durable rewritten evidence

The byte scan at `kernel/scripts/check-witness.sh:100-110` now checks command status and non-empty
output. It does not establish that `od` described every byte in the source file. Any non-empty list
without zero is accepted as a complete scan. `iconv` then accepts NUL as valid UTF-8, and Bash
capture at line 116 deletes it.

Independent reproduction used a real repository and gate, `go version` containing
`go<NUL>version fixture`, and an `od` shim that returned status 0 with the single token `1` for the
first scan before delegating later invocations to `/usr/bin/od`. The result was:

```text
status=0 witnesses=1 marker=yes
warning: command substitution: ignored null byte in input
"tool_versions":{"go":"goversion fixture",...}
```

Later real entropy produced a valid ULID. An independent copy of the actual checks package ingested
that exact emitted file through `Connector.Sync`; it returned `Appended=1`, and the event payload
contained `"go":"goversion fixture"`.

This directly reaches the durable append-only false-evidence HARM and violates C-INV-21/C-INV-23.
C-INV-24's narrower successful-empty statement is true, but it does not close the helper-output
invariant class identified in Round 5.

The finding is verified; a remedy is not. The scan must be validated against the source rather than
against “non-empty output.” Tests must distinguish full, empty, truncated, malformed-token,
out-of-range, and extra-token results while carrying actual NUL data through emission and connector
ingestion.

### MED-1 — Incomplete entropy and semantically invalid clock observations still publish invalid witnesses with status zero

`new_ulid` at `kernel/scripts/check-witness.sh:66-74` requires only non-empty `od` output. It does not
require exactly 16 decimal bytes in range 0..255. The timestamp checks at lines 182 and 213 validate
only digit placement, not calendar validity or the relationship between start and finish.

Independent real-wrapper routes showed:

1. After three real version scans, an entropy `od` returned status 0 with one byte. The real gate
   ran, the wrapper returned 0, and the witness contained an eleven-character run ID:

   ```json
   {"run_id":"01M0TZ31841",...}
   ```

2. A date shim returned the regex-shaped but impossible timestamp
   `2026-99-99T99:99:99Z`. The gate ran and the wrapper returned 0 with that value published.

3. A date shim returned individually well-shaped timestamps and millisecond values whose finish
   preceded start. The wrapper returned 0 with:

   ```json
   {"started_at":"2033-05-18T03:33:20Z","finished_at":"2033-05-18T03:33:19Z","duration_ms":-1000,...}
   ```

The strict reader rejects all three, so they do not append false ledger events. They nevertheless
violate C-INV-10's exactly-one-valid-witness requirement and report successful witnessing where
helper output was incomplete or inconsistent.

The finding is verified; a remedy is not. Entropy requires exact count and numeric-range checks;
clock observations require semantic parsing or equivalent calendar/range/order validation. The
final emitted artifact should be exercised through the real strict reader.

### LOW

None.

## Invariant and mechanism audit

C-INV-24 is appended after C-INV-23, and the lock diff adds exactly its matching row without
changing any earlier row. The citation resolves to `TestEmitter_EmptyHelperOutputIsLoud`, and the
production numbering/table mechanisms plus their self-tests pass.

The new test proves exactly its title: globally empty successful `od` and `date` output are loud.
It does not distinguish scan versus entropy calls, provide actual NUL, exercise non-empty malformed
output, assert gate non-invocation or cleanup, or pass any produced artifact through the strict
reader. It therefore cannot establish the broader helper-observation and valid-publication claims.

## Verified strengths

- Exact successful-empty `od`, entropy, and timestamp routes now fail closed.
- Explicit non-zero helper failures remain loud.
- Direct-PID and process-group helper interruption clean the tracked helper and capture files.
- Hostile uppercase `GIT_*` variables remain stripped from direct observation and real gate Git.
- Normal NUL and invalid UTF-8 are refused before the gate.
- mkdir, staged mktemp, cat, hash, serialization-I/O, spool-loss, and mv routes remain loud.
- Ordinary successful/failing gates publish valid status and exact combined-output digest.
- Reader strictness, typed-nil refusal, filename/content identity, replay, malformed stopping, and
  spool preservation did not regress.
- C-INV-24 and the matching lock row are append-only and mechanically green.

## Residual limits

- Shell behavior was exercised on Linux Bash and GNU userland, not BSD/macOS implementations.
- Cleanup tracks the direct version helper; descendants created by a helper were not exhaustively
  modeled.
- Tool commands have no timeout absent an external signal.
- Concurrent mutation of the output capture between publication and hashing was not exhaustively
  modeled.
- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool evidence remains local and destructible before ingestion by accepted design.
- Mutation operators remain limited and exclude `_test.go`.
- DB integration and mutation were author evidence in this round, not independently rerun evidence.

## Acceptance rationale

Task 5 is not acceptable under Law 9 at commit
`98617e45b23819c86932bf99c6f62f3218aae863`.

The exact empty-output examples and direct-PID cleanup are fixed. The mechanism still equates
“non-empty” with “complete and valid.” A one-token scan again changes observed NUL-containing bytes
and appends the rewrite through the actual connector. Short entropy and syntactically shaped but
invalid/backward clocks separately cause successful-but-uningestible publication.

The HIGH route reaches the central durable false-evidence HARM. Green author tests, mutation counts,
race results, and C-INV-24's closed empty cases do not establish a malformed-output contract they do
not discriminate. Remediation must validate helper output completeness and semantics, verify final
evidence through the strict reader, rerun the relevant mutation/DB/gate evidence, and request
another non-author verdict on a newly frozen commit.
