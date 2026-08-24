NEEDS_WORK

## Round 3 scope and calibration

Independent adversarial review of frozen Task 5 remediation commit `f7096511e351a318fd68bc4b58fd5ea039e21837`, including diff `718f28a..f709651`.

The tracked repository remained frozen:

- `git rev-parse HEAD` remained exactly the tested commit.
- `git status --short` remained empty.
- `git diff --exit-code` and `git diff --check 718f28a..f709651` passed.
- All hostile fixtures and the independent raw-byte test copy existed only under `/tmp`.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` line was the expected index-check positive control.
- `GOCACHE=/tmp/proofbound-review-gocache go test -race ./... -count=1`: passed across all kernel packages.
- Focused checks tests covering strict parsing, repository observation, real Make invocation, controls, and inadmissible bytes passed.
- `make invariant-table-lint`, `make spec-numbering-lint`, and both corresponding mechanism self-tests passed.
- The mutation scratch-copy self-test passed and confirmed the repository Makefile is copied.
- The author’s 37/37 connector and 17/17 CLI mutation counts remain calibrated author evidence, not semantic proof. Mutation was not rerun independently.

## Round 2 remedy adjudication

1. **Gate inherits hostile Git selectors: CLOSED.** `kernel/scripts/check-witness.sh:123` applies the dynamically generated `env -u GIT_*` arguments to the actual Make process. An independent fixture used two real repositories and a real Make target invoking real Git while setting `GIT_DIR`, `GIT_WORK_TREE`, `GIT_COMMON_DIR`, object and alternate-object directories, namespace, index, configuration injection, discovery, replacement, and pathspec variables. Both gate output and witness named A’s SHA; B’s SHA was absent.
2. **NUL deletion before Bash capture: CLOSED.** Tool output containing `go<NUL>different` was captured to a file, rejected before command substitution, invoked no gate, and emitted no witness.
3. **Invalid UTF-8 emitter collapse: CLOSED on exercised bytes.** First-line `0xfe` and `0xff` tool output independently failed before Make and witness creation.
4. **Invalid UTF-8 connector collapse: CLOSED.** `readWitness` checks `utf8.Valid(data)` at `kernel/internal/connector/checks/checks.go:168-170`, before canonicalization and JSON decoding. Independent actual `Sync` routes for `0xfe` and `0xff` appended nothing.
5. **Fake-gate proof gap: CLOSED.** The authored C-INV-20 test now uses two real repositories, real Make, and real Git. Its selector set is narrower than the independent fixture, but the independent extended-class reproduction confirms the production behavior.
6. **Fail-closed byte ordering: CLOSED for normal execution.** NUL inspection and `iconv` validation occur while bytes are still in files at `kernel/scripts/check-witness.sh:53-64`; only admitted first-line bytes enter a Bash variable at line 65. Missing `iconv` also refuses before the gate instead of weakening validation.

## Closure against the governing HARM and routes

The governing HARM is that a failed, incomplete, differently sourced, or byte-collapsed run can become durable successful evidence in an append-only ledger.

| Route | Verdict | Evidence |
|---|---|---|
| Present-null schema fields | Closed | Prior actual routes remain covered; reader logic is unchanged except for earlier UTF-8 refusal |
| Caller-directory mismatch | Closed | Gate continues to change to the derived repository root |
| Direct Git selector redirection | Closed | HEAD/status use the complete dynamically enumerated uppercase `GIT_*` removal |
| Gate Git selector redirection | Closed | Independent extended selector fixture kept real gate Git on A |
| HEAD/status observation failure | Closed | Existing hostile tests still stop before Make |
| Stored 0644 script | Closed | Real Make target still explicitly invokes Bash |
| JSON controls | Closed for admitted UTF-8 | Existing controls round-trip; NUL and invalid UTF-8 are now refused before capture |
| Raw invalid UTF-8 identity collapse | Closed | `0xfe` and `0xff` fail before canonicalization, decoding, or append |
| Typed-nil appender | Closed | Validation remains before empty/absent spool listing |
| Filename/run-id and changed-content identity | Closed for valid evidence | Existing actual connector routes remain green |
| Cursor/cache, malformed stopping, and ingestion preservation | Closed for exercised routes | Connector behavior is unchanged and full tests pass |
| CLI journaling and replay counts | Not regressed by diff; DB not rerun | CLI/store code is unchanged; race and ordinary tests pass |
| Successful gate followed by publication failure | **Open** | Wrapper can exit 0 with no witness or an invalid witness; see MED-1 |
| Pre-gate tool-capture interruption cleanup | **Open** | Signal interruption leaves both version temporary files; see LOW-1 |

The Round 2 central false-source and byte-collapse routes are closed. Acceptance nevertheless remains open because C-INV-10’s witness-publication mechanism has an independently reproduced failure route.

## Findings

### HIGH

None.

### MED-1 — Post-gate publication failures are swallowed, so `check-witnessed` can succeed without one valid witness

C-INV-10 requires the wrapper to write exactly one valid v1 file after either gate outcome (`kernel/internal/connector/checks/SPEC.md:102-103`). The post-gate output, digest, JSON write, and rename operations at `kernel/scripts/check-witness.sh:128-143` are not checked. The unconditional `exit "$exit_code"` at line 145 then replaces their failure status with the gate’s status.

Two independent real-wrapper routes reproduced the class:

1. A successful real Make target removed `.vera/spool` during the gate. JSON redirection and `mv` failed, but the wrapper returned status 0 and no witness existed:

   ```text
   status=0
   gate succeeded
   .../.vera/spool/.witness...: No such file or directory
   mv: cannot stat .../.witness...: No such file or directory
   spool=absent
   ```

2. A hostile `sha256sum` exited 23 without output. The wrapper returned status 0 and published a JSON file containing `"output_sha256":""`, which the strict connector correctly rejects:

   ```text
   status=0
   gate succeeded
   "output_sha256":""
   ```

The connector’s refusal prevents these particular files from becoming durable false ledger events. However, the canonical `make check-witnessed` operation itself reports success despite producing no ingestible observation. This is the same severity class as Round 1’s successful-but-invalid JSON finding: it falsifies the witness mechanism and leaves a successful gate indistinguishable from successfully witnessed execution.

The finding is verified; a remedy is not. Any repair must preserve the real gate result inside a successfully published witness while making capture, hashing, serialization, and atomic publication failures loud. It should be tested with actual failing hash commands and a spool path removed or made unusable after the gate, not solely with ordinary success/failure gates.

### LOW-1 — Version-capture temporary files leak when the wrapper is interrupted before the cleanup trap exists

`first_line` creates `vera-version-output.*` and `vera-version-line.*` at `kernel/scripts/check-witness.sh:46-47`. The script does not install its cleanup trap until line 121, after all three version commands have completed. Ordinary return paths explicitly remove the files, but signals during a version command bypass those paths.

An independent fixture made `go version` block and terminated the wrapper with `timeout`. Both files remained:

```text
status=124
leftover=vera-version-line.oZYddh
leftover=vera-version-output.ELfsps
```

The leak does not create false evidence and is therefore LOW. It is a verified finding; a specific remedy is not verified. Cleanup must cover the pre-gate capture interval and should be tested with an interrupted version command while retaining the existing ordinary failure behavior.

## Mechanism and lock audit

C-INV-20 and C-INV-21 are appended after C-INV-19 in the checks namespace. The lock diff adds exactly two rows at that boundary; no pre-existing lock row changes or moves.

The invariant declarations and table entries retain the enforced shapes:

- C-INV-20 maps to `emitter_test.go::TestEmitter_GateGitUsesSanitizedRepository`.
- C-INV-21 maps to `emitter_test.go::TestEmitter_RejectsInadmissibleToolBytesBeforeGate`.

The production numbering and table checks passed, as did both self-tests. Relevant scripts and self-tests are executable. This verifies the numbering/table mechanism and exact lock change, not the semantic sufficiency of every authored assertion.

## Verified strengths

The findings do not erase the following independently supported behavior:

- The complete inherited uppercase `GIT_*` class is removed from direct Git observations and the real gate process.
- A real Git-using gate and witness share repository A under selectors aimed at B.
- NUL, `0xfe`, and `0xff` tool observations fail before Make and witness publication.
- Raw invalid UTF-8 is refused before canonicalization and JSON decoding, preventing replacement-character identity collapse.
- Ordinary successful and failing gates retain their status when publication succeeds.
- Combined gate output hashing remains exact on the normal route.
- Representable JSON controls still round-trip.
- Repository observation failures still prevent gate invocation.
- The real Make target still supports the Git-stored 0644 script.
- Present-null, typed-nil, filename binding, changed valid content, replay, malformed stopping, and spool-preservation behavior did not regress in the exercised suite.
- C-INV-20/21 and their lock rows are mechanically consistent.

## Residual limits

- Shell portability was exercised on Linux Bash with GNU userland. BSD/macOS behavior of `sed`, `od`, `iconv`, `mktemp`, and date fallback was not independently executed.
- `iconv` is now a fail-closed runtime dependency. Its absence stops the gate; no silent byte-policy downgrade was observed.
- Tool commands can still hang indefinitely; no timeout is specified.
- Concurrent repository mutation between pre-gate identity observation and gate execution was not exhaustively modeled.
- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool evidence remains local and can be lost before ingestion by accepted design; MED-1 concerns the stronger defect that this loss can occur during emission while the witnessed command still returns success.
- Mutation operators remain limited and exclude `_test.go`.
- DB-backed CLI integration was not rerun in this round because Docker access was unavailable.

## Acceptance rationale

Task 5 is not acceptable under Law 9 at commit `f7096511e351a318fd68bc4b58fd5ea039e21837`.

The Round 2 HIGH classes are genuinely closed: real gate Git is bound to the observed repository under hostile selectors, and NUL/invalid UTF-8 can no longer be deleted or collapsed into content identity. C-INV-20/21 and their lock mechanism are also structurally sound.

A separate C-INV-10 route remains reproducibly open. After a successful real gate, a hash or publication failure can be overwritten by the saved gate status, causing `make check-witnessed` to return zero with either no witness or a witness the connector must reject. The pre-gate capture mechanism also leaks temporary files on interruption.

Green tests, race results, mutation counts, and the successfully closed Round 2 routes do not prove publication completeness. Remediation must restate the successful-but-unwitnessed HARM, enumerate post-gate hash/serialization/publication and pre-trap interruption routes, verify proposed remedies against real failures, rerun the connector and DB-aware mutation sweeps, bare `make check`, race tests, and request another non-author verdict on a new frozen commit.
