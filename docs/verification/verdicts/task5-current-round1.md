---
schema: vera.verdict.v1
verdict_id: task5-current-round1
status: NEEDS_WORK
reviewed_commit: baf5c116356f9cb77d0782089cfb8b64f4781dec
findings:
  - finding_id: task5-current-round1-finding
    severity: MED
artifact_path: docs/verification/verdicts/task5-current-round1.md
artifact_sha: d2a42f627253d1008d0585c6ed21c10914168ff11644621465eca81e2b2bda36
---

NEEDS_WORK

## Scope and calibration

Independent adversarial review of frozen Task 5 author commit `baf5c116356f9cb77d0782089cfb8b64f4781dec`, including the full diff from `02626b9` through the tested commit.

The tracked repository remained frozen:

- `git rev-parse HEAD` remained exactly `baf5c116356f9cb77d0782089cfb8b64f4781dec`.
- `git status --short` remained empty.
- `git diff --exit-code` passed.
- A real wrapper invocation created only its designed, gitignored `.vera/spool` output.
- Hostile code additions and repositories existed only under `/tmp`.

Commands and results:

- Bare `make check`: passed. The `index stale; run make index` text was the index self-test’s expected positive control.
- `go test -race ./internal/connector/checks ./cmd/vera -count=1`: passed.
- `go test -tags=integration ./cmd/vera -count=1` against the disposable PostgreSQL 16 container: passed after rerunning outside the network-restricted sandbox.
- A direct SQL check showed the malformed-witness CLI route produced a finished `sync_runs` row with `events_appended=0` and the connector error recorded.
- `make spec-numbering-lint` and `make invariant-table-lint`: passed.
- `git diff --check 02626b9..baf5c11`: passed.
- The C-INV-1 through C-INV-14 lock addition and generalized prefixed-namespace table mechanism are structurally consistent.
- The author’s checks 36/36 and DB-aware CLI 17/17 mutation results are treated as calibrated author evidence, not semantic proof. Mutation was not rerun independently.

## Closure against the stated HARM and ROUTES

The stated HARM is durable false verification evidence: a failed, incomplete, or differently sourced run can be recorded as a successful observation, or evidence can acquire the wrong identity in the append-only ledger.

Route adjudication:

1. **Wrapper status, output, identity, and repository binding: OPEN.** Success/failure status propagation, combined-output hashing, absence of a VERA invocation, ULID uniqueness, and filename/body agreement worked in controlled fixtures. However, the gate and repository observations are not bound to one repository, and a failed dirty-state observation becomes a valid clean observation. See HIGH-2 and HIGH-3.
2. **Strict v1 decoding: OPEN.** Missing, unknown, duplicate, trailing, and many malformed values are refused, but JSON `null` for three scalar fields is accepted and rewritten into zero/false evidence. See HIGH-1.
3. **Content and filename identity: CLOSED for exercised semantic changes.** Moving a witness under another run-id filename fails; changing `exit_code` under one run ID creates a different content identity.
4. **Seen-set, malformed evidence, and spool preservation: CLOSED for current code.** The connector rereads the lexical `.json` set, does not consume a cursor, aborts at malformed evidence with explicit progress, and did not delete or rewrite tested spool files.
5. **Directories and partially initialized state: PARTIAL.** Real `.json` directories are excluded and the authored connector-state routes fail. A typed-nil appender is nevertheless accepted when the spool is empty or absent. See LOW-1.
6. **CLI journaling and counts: CLOSED for exercised database routes.** Initial ingestion reported one append, replay reported zero appends/one existing event, malformed evidence returned an error, and SQL confirmed the failed sync journal was finished with the error recorded. The sync and finish errors are joined rather than one replacing the other.
7. **Tagged mutation harness mechanics: CLOSED for current implementation, with mutation limits retained.** Both mutation execution routes call the common tagged argument/timeout helpers; tagged tests receive `-p=1` and a 30-second ceiling, and the helper self-test runs under `make check`. This establishes the current command shape, not Task 5 semantic correctness.

## Findings

### HIGH-1 — JSON `null` is converted into successful zero/false evidence and appended

The strict reader first verifies only that top-level field names exist, then decodes directly into non-pointer Go scalar fields ([checks.go:169](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks.go:169), [checks.go:185](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks.go:185)). Go’s JSON decoder accepts `null` for `int`, `int64`, and `bool`, leaving their zero values. The subsequent validation accepts those zero values as legitimate ([checks.go:203](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks.go:203)).

Independent hostile witnesses proved all three routes:

- `"exit_code":null` became `ExitCode:0`;
- `"duration_ms":null` became `DurationMS:0`;
- `"git_dirty":null` became `GitDirty:false`.

The `exit_code:null` witness then passed through the actual connector and appended this payload:

```json
{"command":"make check","duration_ms":2000,"exit_code":0,"finished_at":"2026-08-24T12:00:02Z","git_dirty":true,"git_sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","output_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","run_id":"01KZFAPQ00NENTQAXBNENTQAXB","schema":"vera.witness.v1","started_at":"2026-08-24T12:00:00Z","tool_versions":{"go":"go1.26","golangci_lint":"v2","make":"GNU Make 4"}}
```

This directly reaches the central HARM: incomplete evidence with no exit status becomes durable successful-check evidence. It violates the exact typed schema ([checks/SPEC.md:24](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/SPEC.md:24)) and C-INV-1’s strictness claim ([checks/SPEC.md:83](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/SPEC.md:83)).

The authored test checks missing fields but not present `null` fields ([checks_test.go:87](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks_test.go:87)). That fixture therefore holds the relevant distinction constant.

The finding is verified; a remedy is not. Pointer decoding, raw-token type validation, or another strict representation could address it, but any proposed change must be tested through the actual reader and connector for every nullable scalar route.

### HIGH-2 — The wrapper can bind one repository’s identity to another repository’s successful gate

The wrapper derives `repo_root` from its own location ([check-witness.sh:4](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:4)) but invokes bare `make check` in the caller’s current directory ([check-witness.sh:61](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:61)). Its Git commands separately use `git -C "$repo_root"` while inheriting all repository-selecting Git environment variables ([check-witness.sh:77](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:77)).

Two independent reproductions reached the same wrong-source class:

1. The script lived in repository A, but was launched with repository B as its current directory. B’s `check` target printed `WRONG-REPOSITORY-GATE-PASSED` and exited zero. The witness was written to A’s spool and named A’s configured Git identity.
2. The intended repository A’s gate ran and printed `INTENDED-REPOSITORY-GATE-PASSED`, while inherited `GIT_DIR=<B>/.git` and `GIT_WORK_TREE=<A>` caused the valid witness to record repository B’s SHA, `1a99fd1cc7dfd06022a83ed35816e499fd8a7b5c`.

Both witnesses had `exit_code:0` and passed the v1 shape. The output digest correctly bound the wrong gate output, which demonstrates why a correct digest alone does not bind the gate to the claimed repository.

This violates the emitter purpose ([checks/SPEC.md:8](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/SPEC.md:8)), C-INV-10’s “real gate result” claim ([checks/SPEC.md:101](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/SPEC.md:101)), and C-INV-14’s repository-observation claim ([checks/SPEC.md:109](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/SPEC.md:109)). It directly reaches the stated different-bytes durable-evidence HARM.

The test harness replaces `make` with a fixture executable that ignores its working directory and replaces `git` with a fixture that ignores hostile environment ([emitter_test.go:117](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/emitter_test.go:117)). It therefore cannot discriminate repository binding.

The finding is verified; a remedy is not. Anchoring the gate invocation and controlling repository-selection environment are plausible directions, but must be verified together against normal repositories, hostile caller directories, and the full Git selector class.

### HIGH-3 — Failure to inspect dirty state is silently recorded as `git_dirty:false`

The wrapper initializes `git_dirty=false`, then changes it only when command substitution produces non-empty stdout ([check-witness.sh:78](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:78)). It does not inspect `git status`’s exit status.

A hostile Git fixture returned a valid HEAD SHA, then made `git status` print an error to stderr and exit 2. The wrapper suppressed that error, exited zero after the successful gate, and emitted a fully valid witness containing:

```json
"git_dirty":false
```

This is not fail-closed “unavailable” evidence; it is a fabricated clean observation. A real equivalent can arise when HEAD is readable but status cannot inspect the index or work tree.

The authored repository/tool test covers only a successful dirty status command ([emitter_test.go:84](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/emitter_test.go:84), [emitter_test.go:118](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/emitter_test.go:118)). It cannot distinguish “clean” from “inspection failed.”

This directly reaches the central HARM because unknown repository bytes become durable evidence explicitly claiming clean bytes. The finding is verified. The current boolean schema has no unknown state, so the correct remedy may require a schema decision rather than a local shell edit; no remedy is pre-verified.

### MED-1 — A fresh checkout cannot execute the canonical `make check-witnessed` target

The Git object records `kernel/scripts/check-witness.sh` as mode `100644`, while the Make target invokes it directly rather than through Bash ([Makefile:4](/mnt/d/Users/thamm/Desktop/Projects/Vera/Makefile:4)).

A fresh tree materialized from `git archive baf5c11` produced:

```text
-rw-r--r-- kernel/scripts/check-witness.sh
bash: kernel/scripts/check-witness.sh: Permission denied
make: *** [Makefile:5: check-witnessed] Error 126
```

The current Windows-mounted working tree reports permissive executable bits independently of Git’s stored mode, masking the defect. The emitter test also copies the script with explicit mode `0755` and launches it as `bash script` ([emitter_test.go:105](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/emitter_test.go:105), [emitter_test.go:127](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/emitter_test.go:127)), so it cannot detect the committed-mode failure.

This is MED rather than HIGH because it prevents emission instead of writing false ledger evidence. The finding is verified; changing mode or invocation style is not considered verified until a fresh-checkout test exercises the actual Make target.

### MED-2 — Tool-version control characters produce invalid JSON while the wrapper reports gate success

`json_escape` escapes only backslashes and quotation marks ([check-witness.sh:47](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:47)). The tool-version strings are interpolated directly into JSON ([check-witness.sh:86](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/scripts/check-witness.sh:86)).

A fixture whose `go version` first line contained a tab caused the successful wrapper to write a spool file with an unescaped JSON control character. JSON parsing failed at that character, while the wrapper still returned the gate’s zero status.

This does fail closed during ingestion, so it did not reach durable false evidence in the reproduction. It nevertheless violates C-INV-10’s promise that successful or failing emission writes exactly one valid v1 file. Carriage returns and other legal command-output controls are routes through the same class.

The finding is verified; a complete JSON encoder is a possible remedy, but is not verified by this finding.

### LOW-1 — A typed-nil appender is accepted when there is no evidence to iterate

`Sync` checks only `appender == nil` ([checks.go:98](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks.go:98)). An `Appender` interface containing a typed-nil pointer is non-nil. With an empty or absent spool, execution returns success before any method call exposes the invalid appender ([checks.go:101](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/checks/checks.go:101)).

An independent test passed `(*memoryAppender)(nil)` and received:

```text
Result{Listed:0 Appended:0 Existing:0 Cursor:[]}, error=nil
```

With evidence present, the typed nil eventually fails or panics according to its concrete method; acceptance therefore depends on spool contents rather than dependency validity.

This did not produce false evidence in the exercised route, so severity is LOW. The finding is verified; a particular reflection-based or interface-design remedy is not.

## Verified strengths

The findings do not erase behavior that was independently supported:

- Exit-code propagation works for ordinary successful and failing `make` processes.
- The output digest matched the exact combined stdout/stderr bytes.
- VERA was not invoked by emission.
- Ordinary emitted IDs were distinct and filename/body consistent.
- Unknown, missing, duplicate, trailing, range-invalid, hash-invalid, timestamp-invalid, and ULID-invalid evidence was refused outside the nullable-scalar hole.
- Filename relabeling failed.
- Changed semantic witness content created a new content identity.
- Replay deduplicated through the store rather than a cursor.
- Malformed evidence stopped later ingestion while preserving earlier explicit progress.
- Current ingestion did not remove or rewrite spool evidence.
- The real CLI appended once, reported replay as existing, and recorded malformed-ingestion failure in `sync_runs`.
- Sync and finish errors are joined in the CLI.
- Tagged mutation commands currently serialize integration packages and use the tested longer timeout.

## Residual limits

- No signature or verifier identity exists in P1 by explicit non-goal; manually fabricated but schema-valid spool evidence remains possible.
- Spool evidence remains local, gitignored, and destructible before ingestion by explicit accepted design.
- The mutation operator set is small and does not mutate `_test.go`; green counts do not establish assertion meaning.
- The DB verification used an already-running disposable PostgreSQL container and unique fixtures, not a newly isolated server.
- Concurrent repository mutation during a gate was not exhaustively modeled. The verified repository-binding findings already prevent acceptance without relying on that additional race.
- Cross-platform emission was tested through a fresh Git archive and hostile shell fixtures, not every supported shell/tool implementation.

## Acceptance rationale

Task 5 is not acceptable under Law 9 at commit `baf5c116356f9cb77d0782089cfb8b64f4781dec`.

The central HARM remains directly reachable. `exit_code:null` becomes a durable successful `check.run`; repository identity can be taken from different bytes than the gate that passed; and a failed dirty-state observation is silently converted into a clean claim. These are independent HIGH routes, not variations of one fixture. The fresh-checkout execution failure and invalid JSON emission additionally falsify the claimed emitter mechanism, and the typed-nil route leaves defensive initialization incomplete.

The author’s green tests, mutation counts, race run, and DB wiring are useful calibration evidence, but they do not discriminate these properties. Remediation must first restate the HARM in these findings’ terms, enumerate the full nullable-scalar, repository-binding, observation-failure, execution-mode, and JSON-encoding routes, verify each proposed remedy against real hostile data, rerun the connector and DB-aware mutation sweeps, run bare `make check` and race tests, and request another non-author verdict on a new frozen commit.
