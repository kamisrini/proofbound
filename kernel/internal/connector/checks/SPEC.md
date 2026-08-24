# internal/connector/checks — SPEC

Contract for witnessed `make check` emission and ingestion. Written before implementation (Law 6).
The v1 JSON schema lives here; the plan is historical design input, not a second mutable schema.

## 1. Purpose and boundary

`kernel/scripts/check-witness.sh` runs the existing gate and writes one immutable witness file to
`.vera/spool/`. It is shell-only and never invokes `vera`, so evidence emission does not depend on
the product being built. This package reads those files, validates them, and mints `check.run`
events through an injected appender. It never opens a database and never deletes or rewrites spool
files.

Witnesses bind to their content, not filesystem position. The event idempotency key is
`(checks, run_id, content_sha)`: the same file content is absorbed on replay; changed content under
one `run_id` is a visible revision, never silently treated as the original observation. This follows
VD-verification-asymmetry-2dyjnd. Verdict artifacts remain a separate `review.verdict` source under
VD-verdicts-are-artifacts-rl0rab.

## 2. Witness v1 schema

The exact JSON object is:

```json
{
  "schema": "vera.witness.v1",
  "run_id": "<26-character Crockford ULID>",
  "command": "make check",
  "exit_code": 0,
  "started_at": "<RFC3339>",
  "finished_at": "<RFC3339>",
  "duration_ms": 12345,
  "output_sha256": "<64 lowercase hex characters>",
  "git_sha": "<40 or 64 lowercase hex characters>",
  "git_dirty": true,
  "tool_versions": {
    "go": "<non-empty version or unavailable>",
    "golangci_lint": "<non-empty version or unavailable>",
    "make": "<non-empty version or unavailable>"
  }
}
```

Unknown, missing, duplicate, or present-null fields are errors. `exit_code` is 0–255; timestamps are non-zero and
`finished_at >= started_at`; `duration_ms >= 0`. The filename is `<run_id>.json` and must match the
body. ULIDs use uppercase Crockford Base32, exclude I/L/O/U, and have a first character 0–7.

## 3. Go interface

```go
const Version = "checks/1"

type Witness struct { /* exact fields and JSON names from § 2 */ }

type Appender interface {
    Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct {
    SpoolDir string
    IDs      *core.IDGenerator
    Logger   *slog.Logger
}

type Connector struct { /* unexported */ }
func New(*Deps) (*Connector, error)

type Result struct {
    Listed   int
    Appended int
    Existing int
    Cursor   json.RawMessage
}

func (c *Connector) Sync(context.Context, Appender) (Result, error)
```

Files are listed lexically for deterministic diagnostics. `Cursor` is the sorted successfully
validated filename list for observability only and is never read as a seen-set.

## 4. Invariants

1. **C-INV-1 — The v1 schema is exact and strict.** Unknown/missing/duplicate/null fields, invalid
   UTF-8, trailing JSON, invalid ranges, malformed hashes, timestamps, tool versions, or ULIDs are
   refused.
2. **C-INV-2 — One witness mints one check event keyed by run id.** Source is `checks`, kind is
   `check.run`, native id is `run_id`, occurred_at is `started_at`, and version is `checks/1`.
3. **C-INV-3 — Evidence identity binds to content.** The full validated witness is the payload;
   changed content under one run id produces a different content sha.
4. **C-INV-4 — Re-ingestion is idempotent.** A second sync over unchanged spool files appends zero;
   the ledger/appender seen-set decides, never a cursor or filename cache.
5. **C-INV-5 — Malformed evidence fails closed.** Sync returns the named file's error and appends
   no later file; progress before the bad file remains explicit in Result.
6. **C-INV-6 — Spool evidence is never deleted or rewritten by ingestion.** A source scan and a real
   sync both pin file preservation.
7. **C-INV-7 — Filename and body run ids agree.** Moving bytes under another run-id filename cannot
   relabel their native identity.
8. **C-INV-8 — Every dependency is required at construction.** Empty spool dir, nil IDs, logger,
   or dependency object fails at wiring time.
9. **C-INV-9 — The cursor is observational and deterministic.** It is the sorted validated filename
   set and no previous cursor is read.
10. **C-INV-10 — Emission records the real gate result.** The wrapper writes exactly one valid v1
    file after either success or failure and exits with `make check`'s status.
11. **C-INV-11 — Output digest covers the full combined gate output.** The stored SHA-256 matches
    the exact stdout+stderr bytes emitted by the wrapper.
12. **C-INV-12 — Witness emission does not depend on VERA.** The script contains no `vera`
    invocation and succeeds in a fixture with no kernel binary.
13. **C-INV-13 — Emitted identity is self-consistent and sortable.** Filename and body contain the
    same newly minted ULID; two runs produce distinct ids.
14. **C-INV-14 — Repository and tool observations are present.** HEAD, dirty state, and non-empty
    Go/golangci-lint/make version strings are recorded even when a tool reports `unavailable`.
15. **C-INV-15 — Gate output and repository identity share one root.** The wrapper runs the gate at
    its derived repository root and removes every inherited `GIT_*` selector before Git observation.
16. **C-INV-16 — Repository observation fails before the gate.** An unavailable or malformed HEAD
    or failed dirty-state inspection emits no witness and never invokes `make check`.
17. **C-INV-17 — The committed Make target does not require executable script mode.** A 0644 script
    copied with the real Makefile runs through the explicit Bash recipe and emits valid evidence.
18. **C-INV-18 — Runtime dependencies are valid before evidence listing.** A typed-nil appender is
    refused even when the spool is empty or absent.
19. **C-INV-19 — Emitted JSON escapes tool control characters.** Legal command-output controls in
    version strings are represented with JSON escapes and round-trip through the strict reader.
20. **C-INV-20 — The gate process inherits the sanitized repository environment.** A real Git-using
    gate under hostile selectors observes the same repository SHA recorded by the wrapper.
21. **C-INV-21 — Inadmissible evidence bytes fail before identity.** NUL or invalid UTF-8 in a tool
    observation prevents the gate and witness; independently supplied invalid UTF-8 is refused
    before decoding, so distinct invalid bytes cannot collapse into one content identity.
22. **C-INV-22 — Publication failure is never reported as a witnessed gate.** Capture, hashing,
    serialization, and atomic publication failures return non-zero and leave no ingestible witness;
    interruption during pre-gate capture cleans all temporary files.
23. **C-INV-23 — Helper observation failure is never reported as a witnessed gate.** Failed byte
    scans, clocks, or run-id generation return non-zero and leave no ingestible witness; interrupted
    helper children and their temporary files are cleaned.
24. **C-INV-24 — Successful-empty helper output is never treated as observation.** Empty byte scans,
    timestamps, or entropy cannot produce a witness or a relabeled tool observation.

## 5. Invariant table

| Invariant | Statement | Proving test |
|---|---|---|
| C-INV-1 | Strict witness parsing and validation | checks_test.go::TestWitness_StrictValidation |
| C-INV-2 | One witness maps to the exact check event envelope | checks_test.go::TestSync_MintsCheckRunEvent |
| C-INV-3 | Changed evidence changes content identity | checks_test.go::TestSync_ContentBindsIdentity |
| C-INV-4 | Second unchanged sync appends zero | checks_test.go::TestSync_SecondSyncAppendsNothing |
| C-INV-5 | Malformed evidence aborts with explicit progress | checks_test.go::TestSync_MalformedFileFailsClosed |
| C-INV-6 | Ingestion preserves every spool file | checks_test.go::TestSync_NeverDeletesSpoolFiles |
| C-INV-7 | Filename cannot relabel the body run id | checks_test.go::TestSync_RequiresFilenameRunID |
| C-INV-8 | Every constructor dependency is required | checks_test.go::TestNew_RequiresEveryDependency |
| C-INV-9 | Cursor is sorted observation, never input | checks_test.go::TestSync_CursorIsSortedFilenames |
| C-INV-10 | Wrapper writes one witness and returns gate status | emitter_test.go::TestEmitter_RecordsSuccessAndFailure |
| C-INV-11 | Wrapper hashes exact combined output | emitter_test.go::TestEmitter_OutputDigestCoversCombinedBytes |
| C-INV-12 | Wrapper never invokes vera or requires its binary | emitter_test.go::TestEmitter_WorksWithoutVeraBinary |
| C-INV-13 | Wrapper emits distinct self-consistent ULIDs | emitter_test.go::TestEmitter_RunIDIsUniqueAndSelfConsistent |
| C-INV-14 | Wrapper records repository and tool observations | emitter_test.go::TestEmitter_RecordsRepositoryAndTools |
| C-INV-15 | Gate and Git observations bind to the script repository | emitter_test.go::TestEmitter_BindsGateAndGitToRepository |
| C-INV-16 | Failed repository observation prevents gate execution | emitter_test.go::TestEmitter_RepositoryObservationFailsBeforeGate |
| C-INV-17 | Real Make target works with a non-executable script | emitter_test.go::TestEmitter_FreshCheckoutMakeTarget |
| C-INV-18 | Typed-nil appender is rejected before listing | checks_test.go::TestSync_RejectsTypedNilAppenderBeforeListing |
| C-INV-19 | Tool-version controls remain valid JSON | emitter_test.go::TestEmitter_EscapesToolVersionControlCharacters |
| C-INV-20 | Real Git-using gate receives the sanitized environment | emitter_test.go::TestEmitter_GateGitUsesSanitizedRepository |
| C-INV-21 | NUL and invalid UTF-8 fail before evidence identity | emitter_test.go::TestEmitter_RejectsInadmissibleToolBytesBeforeGate |
| C-INV-22 | Publication failures are loud and capture temps are cleaned | emitter_test.go::TestEmitter_PublicationFailuresAreLoud |
| C-INV-23 | Helper observation failures fail closed without publishing a witness | emitter_test.go::TestEmitter_HelperFailuresAreLoud |
| C-INV-24 | Successful-empty helper output fails closed without publishing a witness | emitter_test.go::TestEmitter_EmptyHelperOutputIsLoud |

## 6. Non-goals and recovery

- The wrapper is not a replacement for `make check`; plain `make check` remains independent.
- The connector does not verify that a passing check asserted the right thing. It records the
  mechanical observation and its calibration context.
- Spool files are gitignored local state. Wiping `.vera/` before ingestion loses those witnesses;
  accepted for P1 because checks remain the gate and witnesses are recreatable evidence.
- No signatures or verifier identities in P1.
