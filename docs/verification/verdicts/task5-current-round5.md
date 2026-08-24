NEEDS_WORK

## Round 5 scope and calibration

Independent adversarial review of frozen Task 5 remediation commit
`1e6a0bfc645577fbaf8e0f45e37019f0d1d90783`, including diff `7022fe6..1e6a0bf`.

The tracked implementation remained frozen throughout review. Hostile repositories, helper shims,
and the independent connector-ingestion test existed only under `/tmp`. This verdict file is the
only tracked review output.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` line was the expected index-check
  positive control.
- `GOCACHE=/tmp/proofbound-review-gocache go test -race ./... -count=1`: passed across every kernel
  package.
- Focused emitter tests for ordinary gate status, helper failures, publication failures,
  repository binding, and inadmissible bytes: passed.
- Production invariant-numbering/table checks and both corresponding self-tests: passed.
- `git diff --check 7022fe6..1e6a0bf`: passed.
- C-INV-23 is the sole invariant-lock addition and follows C-INV-22 in the checks namespace.
- The author's mutation, DB integration, and real witnessed-run results are treated as calibrated
  author evidence, not independent semantic proof. DB integration and mutation were not rerun in
  this round.

## Round 4 remedy adjudication

1. **Non-zero `od` failure during byte inspection: CLOSED.** The exact Round 4 shim now exits 1
   before the gate, emits no witness, and leaves no temporary files.
2. **Non-zero `date` failure: CLOSED.** A date command exiting 32 now fails at start-time
   observation before the gate and publishes no witness.
3. **Non-zero `od` failure during run-id generation: CLOSED.** `new_ulid` propagates the command
   failure, and the caller stops before gate invocation.
4. **Direct-PID TERM during a blocked version helper: CLOSED for the direct child.** The wrapper
   exited after direct TERM, killed the tracked helper PID, and removed both version temporary
   files. The same fixture that remained alive in Round 4 now reported
   `alive_after_direct_term=no temps=`.
5. **Successful but empty/malformed `od` byte-scan output: OPEN.** The command status is checked,
   but its output is not validated. Empty successful output again lets NUL reach Bash capture and
   become durable rewritten evidence. See HIGH-1.
6. **Successful but empty/malformed clock and entropy output: OPEN.** Millisecond time is shape
   checked, but RFC3339 timestamp output and the exact 16 entropy bytes are not. Empty successful
   output publishes invalid witnesses with the saved zero gate status. See MED-1.

## Closure against the governing HARM and prior routes

The governing HARM is that a failed, incomplete, differently sourced, or byte-collapsed run can
become durable successful evidence in an append-only ledger, or that the witnessed command can
report success without one valid published observation.

| Route | Verdict | Evidence |
|---|---|---|
| Present-null and strict reader routes | Closed | Reader is unchanged and full checks tests pass |
| Caller/root and hostile `GIT_*` binding | Closed | Independent two-repository real-Git fixture kept gate and witness on A |
| HEAD/status observation failure | Closed | Existing pre-gate failure tests remain green |
| NUL, `0xfe`, and `0xff` with working helpers | Closed | All stop before gate and witness |
| Non-zero `od` byte-scan failure | Closed | Loud before gate, no witness or temp |
| Empty successful `od` byte-scan output | **Open** | NUL is deleted and actual connector appends rewritten evidence |
| Non-zero clock failure | Closed | Loud before gate, no witness |
| Empty successful timestamp output | **Open** | Status 0 and invalid witness with empty `started_at` |
| Non-zero run-id entropy failure | Closed | Loud before gate, no witness |
| Empty successful entropy output | **Open** | Status 0 and invalid witness with ten-character run ID |
| mkdir and staged mktemp failures | Closed | Loud before gate, temps cleaned |
| cat failure | Closed | Loud after gate, no witness |
| Hash exit failure and malformed hash line | Closed | Loud after gate, no witness |
| Serialization I/O and spool-loss failure | Closed | Loud after gate, no witness |
| mv failure | Closed | Loud, no final witness, temp cleaned |
| Process-group interruption | Closed | Helper and version temps removed |
| Direct-PID TERM during version helper | Closed | Wrapper exits, direct helper killed, temps removed |
| Ordinary gate status and output digest | Closed | Valid witness carries actual success/failure and exact combined digest |
| Actual connector identity after helper-byte rewrite | **Open** | Collapsed-NUL witness appended successfully |

## Findings

### HIGH-1 — Successful but empty `od` output still disables NUL refusal and produces durable rewritten evidence

The remediation checks whether `od` exits non-zero at `kernel/scripts/check-witness.sh:100-103`.
It does not verify that successful output describes every byte in the first-line file. The loop at
lines 104-109 therefore treats empty output as “no NUL found.” `iconv` accepts NUL as valid UTF-8,
and Bash command substitution at line 114 deletes it.

Independent reproduction used a real repository and gate, `go version` output containing
`go<NUL>version fixture`, and an `od` shim which returned status 0 with empty output only for the
first byte scan, then delegated later calls to real `/usr/bin/od`. The wrapper reported:

```text
status=0 witnesses=1 marker=yes
warning: command substitution: ignored null byte in input
"tool_versions":{"go":"goversion fixture",...}
```

The later real entropy read produced a valid ULID, so the witness passed the strict schema. An
independent copy of the actual checks package ingested that exact emitted file through
`Connector.Sync`; the result was `Appended=1`, and the event payload contained
`"go":"goversion fixture"`.

This directly reaches the append-only false-evidence HARM: the tool emitted one byte sequence, the
successful wrapper recorded another, and the actual connector made the rewrite durable. It violates
C-INV-21's fail-closed byte policy and C-INV-23's helper-observation claim.

The finding is verified; a remedy is not. Checking process exit alone is insufficient. Any repair
must validate the byte-scan result against the source length/content or use a primitive whose
success semantically establishes the absence of NUL. Missing, non-zero, empty-success, truncated,
and malformed-success helper routes must be tested with actual NUL data through both emission and
connector ingestion.

### MED-1 — Successful but incomplete clock and entropy observations publish invalid witnesses with gate status zero

The same status-only assumption remains in two evidence-construction routes:

- `started_at` and `finished_at` at `kernel/scripts/check-witness.sh:180-183` and `211-214` check only
  `date`'s exit code, not RFC3339 shape or value.
- `new_ulid` at lines 66-73 checks `od`'s exit code but not that exactly 16 decimal bytes were
  returned before publishing the encoded value.

Independent real-wrapper reproductions demonstrated both routes:

1. A date shim delegated all calls except the start-timestamp call, where it returned status 0 and
   no bytes. The real gate ran and the wrapper returned 0 with:

   ```json
   {"started_at":"","finished_at":"2026-08-24T22:36:16Z",...}
   ```

2. An od shim delegated the three version scans, then returned status 0 and no bytes for entropy.
   The real gate ran and the wrapper returned 0 with a ten-character run ID:

   ```json
   {"run_id":"01M0TYNKGQ",...}
   ```

The strict reader correctly rejects both files, so these reproductions do not append false ledger
events. They nevertheless violate C-INV-10's exactly-one-valid-witness requirement and repeat the
successful-but-uningestible publication class. They also falsify C-INV-23's broad statement that
clock or run-id observation failure is never reported as a witnessed gate.

The finding is verified; a remedy is not. Timestamp bytes need real format/value validation, and
entropy output needs exact count and numeric-range validation before encoding. The final emitted
artifact should be exercised through the strict reader so a non-empty but invalid serialization
cannot satisfy the proving test.

### LOW

None.

## Invariant and mechanism audit

C-INV-23 is appended after C-INV-22, and the lock diff adds exactly its matching row without
changing an earlier row. Its proving-test citation has the enforced one-test-token shape and
resolves to `TestEmitter_HelperFailuresAreLoud`. Production invariant checks and their self-tests
pass.

The new test proves non-zero `od` and `date` exits are loud. It does not supply NUL to its od route,
distinguish byte-scan from entropy use, exercise successful malformed output, assert that the gate
did not run, inspect temporary cleanup, or pass a resulting file through the strict reader. Its
green result therefore does not establish the full C-INV-23 statement.

The direct-child mechanism is real: `first_line` stores the background helper PID, `wait` is
interruptible, and EXIT cleanup kills the tracked PID before deleting registered files. The exact
Round 4 direct-PID fixture independently closed.

## Verified strengths

- Explicit non-zero `od`, `date`, and run-id helper failures now stop cleanly.
- Direct-PID TERM closes the previously reported blocked-helper and temporary-file route.
- The full hostile uppercase `GIT_*` class remains stripped from direct observations and the real
  Git-using gate.
- With correct helper output, NUL and invalid UTF-8 fail before gate invocation.
- mkdir, staged mktemp, cat, hash, serialization-I/O, spool-loss, and mv failure routes remain loud.
- Ordinary successful and failing gates publish valid evidence with the actual gate status.
- Combined output hashing remains exact on the normal route.
- Present-null, typed-nil, filename/content identity, replay, malformed stopping, and spool
  preservation behavior did not regress in the full suite.
- C-INV-23 numbering and lock mechanics are append-only and mechanically green.

## Residual limits

- Shell behavior was exercised on Linux Bash and GNU userland, not BSD/macOS implementations.
- Cleanup tracks and kills the direct version helper PID; descendants created by a helper were not
  exhaustively modeled.
- Tool commands have no timeout absent an external signal.
- Concurrent mutation of the output capture between publication and hashing was not exhaustively
  modeled.
- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool evidence remains local and destructible before ingestion by accepted design.
- Mutation operators remain limited and exclude `_test.go`; green counts do not establish assertion
  meaning.
- DB integration and mutation were author evidence in this round, not independently rerun evidence.

## Acceptance rationale

Task 5 is not acceptable under Law 9 at commit
`1e6a0bfc645577fbaf8e0f45e37019f0d1d90783`.

The explicit Round 4 examples are fixed: non-zero helper exits now fail closed, and direct-PID TERM
now kills the tracked helper and cleans its temporary files. The invariant class is not closed.
Successful but empty helper output bypasses the same boundaries: the byte scan again permits NUL to
be deleted and appended as different evidence, while clock and entropy routes publish invalid
witnesses under a successful gate status.

The HIGH route reaches the exact durable false-evidence HARM through the actual connector. Green
author tests, mutation counts, full race results, and the repaired non-zero examples cannot prove a
helper-output contract they do not discriminate. Remediation must validate helper output rather
than only helper exit, verify final evidence through the strict reader, rerun the relevant
mutation/DB/gate evidence, and request another non-author verdict on a newly frozen commit.
