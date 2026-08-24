NEEDS_WORK

## Round 4 scope and calibration

Independent adversarial review of frozen Task 5 remediation commit
`7022fe6a2130b0cf979ef1d3db68c10cb89d570e`, including diff `f709651..7022fe6`.

The tracked implementation remained frozen throughout review. Hostile repositories, tool shims,
and the independent connector-ingestion test existed only under `/tmp`. This verdict file is the
only tracked review output.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` line was the index self-test's
  expected positive control.
- `GOCACHE=/tmp/proofbound-review-gocache go test -race ./... -count=1`: passed across every kernel
  package.
- Focused emitter tests, including `TestEmitter_PublicationFailuresAreLoud`: passed.
- Production invariant-numbering/table checks and both corresponding self-tests: passed.
- `git diff --check f709651..7022fe6`: passed.
- C-INV-22 is the sole lock addition and follows C-INV-21 in the checks namespace.
- The author's 37/37 connector and 17/17 DB-aware CLI mutation results, DB integration, and real
  witnessed run are treated as calibrated author evidence, not independent semantic proof. DB
  integration and mutation were not rerun in this round.

## Round 3 remedy adjudication

1. **Initial spool creation: CLOSED.** A real repository with `.vera` replaced by a regular file
   exited 1, did not invoke the gate, and left no temporary file.
2. **Version, output, and witness `mktemp` failure: CLOSED for ordinary command failure.** Independent
   shims failed the second, seventh, and eighth `mktemp` calls. Every route exited 1 before the gate
   and cleanup removed previously created temporary files.
3. **Gate-output publication: CLOSED for `cat` failure.** A failing `cat` produced exit 1, no witness,
   and no temporary residue after the successful real gate.
4. **Hash failure and malformed hash output: CLOSED.** A hash command exiting 23 and one returning a
   successful but malformed line both produced exit 1 and no witness.
5. **Serialization after spool loss: CLOSED.** Removing the spool during a successful real gate made
   JSON redirection fail; the wrapper returned 1 and published no witness.
6. **Atomic publication: CLOSED for `mv` failure.** A failing `mv` returned 1, published no final
   witness, and the EXIT trap removed the temporary witness.
7. **Saved gate status: CLOSED on successful publication and the exercised publication failures.**
   Ordinary successful and failing gates still return their own status after a valid witness is
   published. Exercised publication failures override a saved zero gate status with a non-zero
   wrapper result.
8. **Process-group interruption cleanup: CLOSED.** Terminating the wrapper's process group while
   `go version` blocked left no version-capture temporary files.
9. **Complete fail-closed helper behavior: OPEN.** An `od` failure during NUL inspection is ignored,
   allowing Bash to delete NUL and publish durable rewritten evidence. See HIGH-1.
10. **Direct-process interruption cleanup: OPEN.** Sending TERM to the wrapper PID alone leaves it
    blocked in the child and leaves both temporary files present. See LOW-1.

## Closure against the governing HARM and routes

The governing HARM is that a failed, incomplete, differently sourced, or byte-collapsed run can
become durable successful evidence in an append-only ledger, or that the witnessed command can
report success without one valid published observation.

| Route | Verdict | Evidence |
|---|---|---|
| Present-null and strict schema routes | Closed | Reader logic is unchanged and full checks tests pass |
| Caller/root and complete `GIT_*` binding | Closed | Round 3's real two-repository result is unaffected by this diff |
| HEAD/status observation failures | Closed | Existing pre-gate failure tests remain green |
| NUL/invalid UTF-8 with working helpers | Closed | NUL, `0xfe`, and `0xff` still fail before the gate |
| NUL with failed byte-inspection helper | **Open** | Failed `od` is ignored and NUL is deleted; see HIGH-1 |
| mkdir failure | Closed | Loud before gate, no witness or temp |
| version/output/witness mktemp failure | Closed | Loud before gate, registered temps cleaned |
| cat failure | Closed | Loud after gate, no witness |
| hash exit failure or malformed output | Closed | Loud after gate, no witness |
| JSON redirection/serialization failure | Closed for exercised I/O failure | Spool removal is loud and leaves no witness |
| Evidence-construction helper failure | **Open** | Failed `date` publishes an invalid witness while returning zero; see MED-1 |
| mv failure | Closed | Loud, no final witness, temp cleaned |
| Process-group interruption | Closed | Version temps removed |
| Direct-PID TERM while a tool blocks | **Open** | Wrapper remains alive and temps remain; see LOW-1 |
| Actual connector identity after byte rewrite | **Open** | The emitted collapsed-NUL witness appended successfully |

## Findings

### HIGH-1 — Failure of `od` disables the fail-closed NUL policy and produces durable rewritten evidence

The NUL scan at `kernel/scripts/check-witness.sh:84` is a command substitution feeding a `for`
loop. The exit status of `od` is never tested. If `od` fails, the loop simply has no iterations and
execution proceeds to `iconv`. NUL is valid UTF-8, so `iconv` accepts it. Bash then captures the
line at `kernel/scripts/check-witness.sh:94`, deletes NUL, and continues with altered evidence.

Independent reproduction used a real repository and gate, a `go version` emitting
`go<NUL>version fixture`, and an `od` shim which failed only its first invocation before delegating
all later calls to the real `/usr/bin/od`. The result was:

```text
status=0 witnesses=1 marker=yes
warning: command substitution: ignored null byte in input
"tool_versions":{"go":"goversion fixture",...}
```

The later real `od` call produced a valid ULID, so the witness was otherwise valid. An independent
copy of the actual checks package passed that emitted file through `Connector.Sync`; it returned
`Appended=1`, and the event payload contained `"go":"goversion fixture"`.

This is not merely an unavailable-tool or invalid-spool route. The gate and wrapper report success,
the connector accepts and appends the event, and the append-only ledger permanently records bytes
different from the tool observation. It directly reopens the Round 2 central byte-fidelity HARM and
violates C-INV-21's claim that NUL prevents gate and witness.

The finding is verified; a remedy is not. Any fix must make byte-inspection execution and output
validity explicit before Bash capture, and must prove missing, failing, and malformed `od` routes
with actual NUL data through emission and connector ingestion. The separate `od` use for ULID
entropy also needs failure/byte-count validation rather than assuming a successful 16-byte read.

### MED-1 — Failed timestamp observation still publishes an invalid witness and returns the successful gate status

`now_ms` deliberately suppresses the first `date` failure at
`kernel/scripts/check-witness.sh:42-46`, while `started_at` and `finished_at` at lines 157 and 179
do not test `date` at all. The new publication checks establish only that `printf` wrote non-empty
bytes; they do not establish that those bytes are a valid v1 witness before `mv`.

With a `date` shim that always exited 32, a successful real gate produced status 0 and published:

```json
{"run_id":"0000000000DGNJD7B0T2B8J30W","started_at":"","finished_at":"","duration_ms":0,...}
```

The strict connector correctly rejects this file, so the reproduction did not append false ledger
evidence. It nevertheless violates C-INV-10's requirement to write one valid v1 file and repeats the
successful-but-uningestible publication class. It also shows that C-INV-22's serialization wording
is broader than its `printf`/non-empty check.

The finding is verified; a remedy is not. Timestamp acquisition failure and malformed timestamp
output must be checked as evidence-construction failures, and the publication route needs a test
which proves the final bytes satisfy the real v1 reader rather than merely being non-empty.

### LOW-1 — TERM sent directly to the wrapper does not interrupt a blocked version child or promptly clean its files

The EXIT and TERM traps are installed before temporary creation at
`kernel/scripts/check-witness.sh:19-22`, which fixes cleanup when the whole process group is
terminated. Bash defers a trapped TERM while it waits for a foreground child, however, and the
handler does not propagate the signal to that child.

An independent fixture started the wrapper in its own process group, blocked forever in
`go version`, and sent `SIGTERM` only to the wrapper PID. After 300 ms the wrapper was still alive
and both files remained:

```text
alive_after_direct_term=yes
temps=vera-version-line.Udmjob vera-version-output.UvjZWG
```

Sending TERM to the full process group then allowed the authored cleanup path to complete. This is
LOW because it does not create false evidence, but C-INV-22 says interruption during pre-gate
capture cleans all temporary files without restricting interruption to process-group delivery. The
proving test at `kernel/internal/connector/checks/emitter_test.go:320-334` exercises only the broader
process-group signal and therefore cannot detect this route.

The finding is verified; a remedy is not. The contract could explicitly narrow supported signal
semantics, or the wrapper could manage/terminate the active child, but either remedy must be checked
with direct-PID TERM as well as process-group interruption.

## Invariant and mechanism audit

C-INV-22 is appended after C-INV-21, and the lock diff adds exactly its one matching row without
changing an earlier row. Its table citation has the enforced one-test-token shape and resolves to
`TestEmitter_PublicationFailuresAreLoud`. Production invariant checks and their self-tests pass.

The proving test has real value: it catches spool-loss and hash-command failure, and its
process-group signal correctly proves that cleanup route. It does not prove the full C-INV-22
statement. It omits mkdir, staged mktemp, cat, malformed hash output, timestamp/evidence validity,
mv, helper failure, and direct-PID signal routes. Independent fixtures closed several of those in
the product, but HIGH-1, MED-1, and LOW-1 remain.

## Verified strengths

- The Round 3 spool-loss and hash-exit findings are fixed at their real routes.
- mkdir, every staged mktemp, cat, malformed hash, JSON redirection, and mv failures tested here are
  loud and leave no final witness.
- Temporary files are cleaned on ordinary failures and process-group interruption.
- Successfully published ordinary success/failure witnesses preserve the gate's actual status.
- Normal combined-output digest behavior remains exact.
- With functioning byte helpers, NUL and invalid UTF-8 are refused before gate invocation.
- Repository binding, present-null refusal, typed-nil refusal, filename/content identity, replay,
  malformed stopping, and store-independent connector behavior did not regress in the full suite.
- C-INV-22 numbering and lock mechanics are append-only and mechanically green.

## Residual limits

- Shell behavior was exercised on Linux Bash and GNU userland, not BSD/macOS implementations.
- Concurrent mutation of the output capture between `cat` and hashing was not exhaustively modeled.
- Tool commands have no timeout; direct-PID termination of a blocked child is the concrete route
  reported above.
- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool evidence remains local and destructible before ingestion by accepted design.
- Mutation operators remain limited and exclude `_test.go`; green counts do not establish assertion
  meaning.
- DB integration and mutation were author evidence in this round, not independently rerun evidence.

## Acceptance rationale

Task 5 is not acceptable under Law 9 at commit
`7022fe6a2130b0cf979ef1d3db68c10cb89d570e`.

The explicit Round 3 publication routes are substantially improved and independently close under
mkdir, staged mktemp, cat, hash, serialization-I/O, and mv failure. The mechanism still fails at a
more fundamental boundary: its fail-closed byte policy assumes `od` succeeds. One failed helper
causes Bash to delete NUL, the wrapper to return zero, and the actual connector to append the
rewritten observation. That directly reaches the governing append-only false-evidence HARM.

The timestamp and direct-signal routes separately show that non-empty serialization and a
process-group-only test do not establish the full C-INV-22 claim. Green author tests, mutation
counts, race results, and the closed publication fixtures cannot compensate for those missing
discriminations. Remediation must enumerate helper failure and malformed output as part of the byte
and evidence-construction routes, verify the final witness through the actual reader, resolve or
narrow direct-signal semantics, rerun the relevant mutation/DB/gate evidence, and request another
non-author verdict on a new frozen commit.
