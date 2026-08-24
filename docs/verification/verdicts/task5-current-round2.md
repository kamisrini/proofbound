NEEDS_WORK

## Round 2 scope and calibration

Independent adversarial review of frozen remediation commit `b634982ec80d08911efc3a3d4447abc48e7cdb37`, including diff `95f9cfc..b634982`.

The tracked repository remained frozen:

- `git rev-parse HEAD` remained exactly the tested commit.
- `git status --short` remained empty.
- `git diff --exit-code` and `git diff --check 95f9cfc..b634982` passed.
- Hostile tests and repositories existed only under `/tmp`.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` line was the expected index-check positive control.
- `go test -race ./internal/connector/checks ./cmd/vera -count=1`: passed.
- DB-backed `go test -tags=integration ./cmd/vera -count=1`: passed against the disposable PostgreSQL 16 container.
- Production invariant-numbering and table checks passed.
- Both corresponding mechanism self-tests passed.
- The mutation scratch-copy self-test passed and confirmed the repository Makefile is copied.
- The author’s 37/37 connector and 17/17 DB-aware CLI mutation results are treated as calibrated author evidence, not semantic proof. Mutation was not rerun independently.

## Round 1 remedy adjudication

1. **Present-null evidence: CLOSED.** All eleven top-level fields were independently replaced with whitespace-surrounded `null` and passed through the actual connector. Every route failed before append. Nested `tool_versions.go`, `golangci_lint`, and `make` nulls also failed before append. No event was emitted.
2. **Caller working directory: CLOSED.** A wrapper launched from an unrelated directory ran the gate from its derived repository root.
3. **Inherited Git environment: PARTIAL.** The wrapper’s own HEAD/status observations remove the dynamically enumerated uppercase `GIT_*` environment class. The gate process still inherits that class and can inspect a different repository; see HIGH-1.
4. **Repository observation failures: CLOSED for the Round 1 routes.** Real unborn HEAD, malformed HEAD, and unreadable-index/status-failure repositories all exited before `make`, produced no witness, and left the gate marker absent.
5. **Stored 0644 emitter mode: CLOSED.** A fresh Linux clone stored `kernel/scripts/check-witness.sh` as mode 0644, and the real Make target invoked it through Bash. The underlying imported gate later failed on unrelated legacy script modes, but the wrapper correctly wrote a valid witness with that gate’s exit code 2. The Round 1 emitter-execution defect itself is closed.
6. **JSON control escaping: PARTIAL.** Representable controls—backspace, form feed, tab, carriage return, U+0001, U+001F, and the generic non-NUL control branch—round-trip. NUL cannot enter a Bash variable and is silently deleted, while invalid UTF-8 is silently normalized during ingestion. See HIGH-2.
7. **Typed-nil appender: CLOSED.** Independently supplied typed-nil appenders were rejected before listing both empty and absent spool directories.
8. **Scratch Makefile mechanism: CLOSED for the stated repair.** Both mutation-copy routes use `copyTree`; the helper copies the repository Makefile when the source is `kernel`, and its self-test verifies exact copied content.

## Closure against HARM and ROUTES

The governing HARM remains: an incomplete, failed, or differently sourced gate can become durable evidence of a successful check, with false repository or content identity that cannot be repaired in place.

| Route | Verdict | Evidence |
|---|---|---|
| Gate status and combined-output digest | Closed | Success/failure status and exact combined digest remain correct |
| Missing/duplicate/unknown/trailing/null schema fields | Closed | All top-level and nested null routes failed before append |
| Caller-directory repository mismatch | Closed | Gate changes to the derived root |
| Git selector redirection | Open | Gate inherits selectors even though wrapper observations do not |
| HEAD/status observation failure | Closed | Real hostile repositories stopped before gate invocation |
| Filename/run-id relabel | Closed | Existing actual route still fails |
| Changed content revision | Open for invalid bytes | Distinct invalid UTF-8 bytes collapse to one content identity |
| Cursor/cache seen-set | Closed | Connector continues to reread spool and use appender/store dedupe |
| Malformed evidence and preservation | Closed for exercised routes | Malformed input aborts and files remain |
| `.json` directories and initialized state | Closed | Directories ignored; typed nil rejected before listing |
| CLI sync journal/error/count behavior | Closed | DB integration remains green |
| Mutation harness DB serialization/timeout/scratch Makefile | Closed for current mechanics | Helper paths and self-tests passed |

## Findings

### HIGH-1 — The gate still inherits hostile `GIT_*` selectors and can check a different repository than the witness names

The remediation builds `git_env_args` and applies it only inside `git_repo` at `kernel/scripts/check-witness.sh:8-17`. The actual gate invocation at `kernel/scripts/check-witness.sh:103` is:

```bash
(cd "$repo_root" && make check)
```

It inherits the original environment unchanged.

Independent real-repository reproduction:

1. Repository A contained the wrapper and a successful `check` target that printed the SHA visible to Git.
2. Repository B had a different commit.
3. The wrapper was launched from an unrelated caller directory with `GIT_DIR=<B>/.git` and `GIT_WORK_TREE=<A>`.
4. The wrapper’s sanitized observation correctly recorded A’s SHA:
   `6d5b3376e9a8c2fbb3bb3f092aef58d5964bbe21`.
5. The gate inherited the selectors and successfully printed B’s SHA:
   `66a4bb7b27fc1eaeb2c310a62518f4010b5d56a9`.
6. The witness had `exit_code:0`, named A, and hashed output proving that the gate actually inspected B.

The new authored fixture cannot detect this route. Its fake Git fails if a selector reaches the wrapper’s direct Git process, but its fake Make executable does not inspect the environment or invoke Git (`kernel/internal/connector/checks/emitter_test.go:99-115`, `224-226`).

This directly violates C-INV-15’s “Gate output and repository identity share one root” claim (`kernel/internal/connector/checks/SPEC.md:111-112`) and reaches the central different-source durable-evidence HARM.

The finding is verified; a remedy is not. Passing the sanitized environment to the gate is a plausible direction, but must be tested using a gate that itself invokes real Git under the full selector class. A fixture whose Make process ignores Git cannot verify the invariant.

### HIGH-2 — NUL and invalid UTF-8 evidence bytes are deleted or collapsed instead of preserved or refused

The new `json_escape` loop operates on a Bash string (`kernel/scripts/check-witness.sh:58-80`). Bash variables cannot contain NUL. Separately, `readWitness` canonicalizes and decodes raw JSON without first checking UTF-8 validity (`kernel/internal/connector/checks/checks.go:162-205`).

Two independent routes were reproduced.

#### NUL deletion

A real tool fixture emitted:

```text
go<NUL>version fixture
```

Bash warned that command substitution ignored the null byte. The wrapper still exited with the successful gate status and emitted valid evidence containing:

```json
"go":"goversion fixture"
```

The observed byte was deleted rather than represented as `\u0000` or refused. The authored control test begins at U+0001 and never exercises NUL (`kernel/internal/connector/checks/emitter_test.go:193-203`, `226`), despite the remediation journal claiming U+0000–U+001F coverage.

#### Invalid UTF-8 collapse

A tool emitted byte `0xff` inside its version. The wrapper wrote that raw invalid byte into the JSON file. The actual reader accepted it and produced:

```text
Go: "go�version fixture"
error: nil
```

More decisively, two witnesses with the same run ID differing only by byte `0xfe` versus `0xff` were ingested sequentially. The first appended; the second returned `Existing=1`. Both invalid inputs collapsed to the same replacement-character payload and therefore the same `content_sha`.

This violates:

- C-INV-1’s exact/strict schema and malformed-evidence refusal at `kernel/internal/connector/checks/SPEC.md:83-84`;
- C-INV-3’s requirement that changed content under one run ID changes content identity at `SPEC.md:87-88`;
- C-INV-19’s claim that legal tool-output controls are represented and round-trip at `SPEC.md:119-120`.

It directly reaches the append-only HARM: different observed bytes become one durable fact, and the ledger cannot later recover which bytes were seen.

The finding is verified; a remedy is not. Shell escaping alone cannot preserve NUL in a Bash variable. Any remediation must choose and verify a complete byte policy—preserve through an encoding capable of representing every admitted byte, or fail closed before publishing—and the reader must reject invalid UTF-8 before any decoder can replace it.

## MED

None beyond the HIGH byte-fidelity class above.

## LOW

None.

## Mechanism and lock audit

C-INV-15 through C-INV-19 are appended after C-INV-14 in `docs/invariants.lock:18-22`; no earlier checks invariant moved or changed.

The production numbering/table checks and their self-tests passed. The new table cells retain the enforced one-test-token shape.

The mutation scratch repair is real:

- `copyTree` copies the repository Makefile into a kernel scratch root at `tools/mutants/main.go:247-250`.
- `TestCopyTreeIncludesRepositoryMakefile` verifies exact content at `tools/mutants/main_test.go:32-46`.
- Both calibration and mutant execution use `copyTree`.

This establishes the scratch artifact repair. It does not compensate for the two semantic findings above.

## Verified strengths

The findings do not erase the following independently supported behavior:

- All present-null schema routes now fail before append.
- Missing, unknown, duplicate, trailing, range-invalid, malformed-hash, timestamp, and ULID routes remain refused.
- Ordinary success and failure preserve gate status.
- Combined stdout/stderr hashing remains exact.
- Caller cwd no longer selects the gate root.
- The wrapper’s own Git observations are protected from the inherited uppercase `GIT_*` class.
- Missing/malformed HEAD and failed status observation stop before the gate.
- The 0644 emitter script runs through the real Make target.
- Representable non-NUL JSON controls round-trip.
- Typed-nil appenders fail before empty or absent spool inspection.
- Replay, filename binding, malformed-file stopping, cursor observation, and spool preservation remain intact.
- DB-backed CLI append/replay/malformed-journal behavior remains green.
- C-INV numbering and table mechanisms remain operational.

## Residual limits

- No signatures or verifier identities exist in P1 by explicit non-goal.
- Spool files remain local and can be lost before ingestion by accepted design.
- A fresh Linux clone showed unrelated imported gate scripts still stored non-executable; the emitter correctly recorded that gate failure. This is repository mechanism debt outside the repaired emitter-mode route.
- Concurrent repository changes between the pre-gate observation and gate execution were not exhaustively modeled.
- Mutation operators remain limited and exclude `_test.go`; 37/37 and 17/17 do not prove assertion meaning.
- The DB integration used the existing disposable PostgreSQL container and unique test fixtures.
- Cross-platform shell behavior beyond the exercised Linux Bash environment remains unproven.

## Acceptance rationale

Task 5 remains unaccepted under Law 9 at commit `b634982ec80d08911efc3a3d4447abc48e7cdb37`.

Five of the six Round 1 findings are closed at their stated routes, and the repository-binding repair closes caller cwd plus the wrapper’s own Git observation. It does not close the invariant: the gate still receives the hostile selectors and can successfully inspect repository B while the witness names repository A.

The JSON remedy likewise closes only controls representable in a Bash variable. NUL is silently removed, and invalid UTF-8 bytes are accepted and collapsed through replacement characters; distinct bytes under one run ID become one durable content identity.

Both remaining HIGH findings reach the exact append-only HARM. Remediation must address the invariant classes rather than these example fixtures, independently verify any proposed fixes against real Git-using gates and hostile byte streams, append any new invariant identities rather than inserting numbers, regenerate the lock, rerun bare `make check`, race and DB integration tests, rerun both mutation sweeps, and request another non-author verdict on a new frozen commit.
