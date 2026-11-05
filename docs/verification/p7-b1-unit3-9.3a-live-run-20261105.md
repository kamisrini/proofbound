# P7 B1 unit-3 — §9.3a LIVE acceptance run (non-author evidence) — 2026-11-05

**Documentary evidence record** (not a schema-bearing verdict). It records what an **independent
non-author** (Build Law 9) actually observed when running the unit-3 runtime fixture **with the
prepared PostgreSQL archive present**, discharging the live-run half of spec §9 item 3a that the
companion verdict
[`verdicts/p7-b1-unit3-impl-review-round1-20261105.md`](verdicts/p7-b1-unit3-impl-review-round1-20261105.md)
had to leave to the owner. The **owner remains the final acceptor**; this file supplies the live
evidence only. `DEF-P7-001` is **not** moved by this record.

- **Reviewed commit:** the unit-3 implementation at `15df68773d7908b08024e086eb6756e304a094f2`
  (the commit the companion verdict reviewed). The live run was executed with the working tree at
  `1bc91c7caed5b6dc198e84ecdcf215524893ca27`; `git diff 15df687..1bc91c7` touches only
  `docs/`+`notes/` (the verdict, the regenerated integrity audit, and the state note) and **no
  `kernel/` code**, so the implementation under test is byte-identical to the reviewed commit.
- **Spec:** `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (round-6 ACCEPTED), §9 item 3a.
- **Reviewer:** independent non-author; authored no unit-3 code, spec, or evidence. No implementation
  code was edited for this run. `notes/state.md` and the journal were not written.

## Archive binding (pre-condition)

- **Archive present at:** `~/proofbound-p7-build/postgres-16.4-subset.txz`.
- **Whole-file SHA-256:** `fa1ade9b4f9b7266add9495e0dc00be419f57857a9b3fc12632f24caaf6444e3`
  — **equals** the committed manifest `postgres-16.4-subset-archive.json` `archive.sha256`. MATCH.
- **Declared `postgres_binary_sha256`:**
  `ba2447e5a38e52d801d0faecc8a3ee46d4cb57792cabc6e80801af4d4888fcd5`.
- **Environment:** Linux (WSL2), `go` = go1.27; the staged server self-reported
  `PostgreSQL 16.4 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 11.4.0) 11.4.0, 64-bit`.

## Commands run (from `/home/thamm/projects/proofbound/kernel`)

```
PROOFBOUND_P7_PG_ARCHIVE=~/proofbound-p7-build/postgres-16.4-subset.txz \
PROOFBOUND_P7_EMIT_DIR=<tmp> \
  go test ./internal/twin -run 'TestRuntimeFixture|TestEmitReal' -v     # archive-gated live path
go test ./internal/twin -run TestWitnessLaunchedProcess -v              # §5.1 witness-function controls
```

## Observed results — ALL PASS (live, not skipped)

| Test | Result | What it establishes live |
|---|---|---|
| `TestRuntimeFixtureWitnessesRealLaunchedPostgres` | **PASS** (4.81s) | a REAL from-source postgres launched; `/proc/<pid>/exe` witness resolved the declared digest |
| `TestRuntimeFixtureLedgerPreservationWitness` | **PASS** (4.88s) | source ledger byte-identical across the isolated replay; mutation positive-control fired |
| `TestRuntimeFixtureCleanupWitness` | **PASS** (7.24s) | owned temp root absent after success; suppressed-removal positive-control left it present |
| `TestRuntimeFixtureNoFetchDeclaration` | **PASS** (2.11s) | offline open from staged bytes; `bin/pg_ctl` short-circuits fetch (labeled declaration) |
| `TestEmitRealBehaviorLock` | **PASS** (5.06s) | authored a `proofbound.p7.behavior-lock.v1` over the real proof → `EvaluateLock == LOCKED` |
| `TestEmitRealCaseManifest` | **PASS** (4.72s) | authored a `proofbound.p7.case-manifest.v1` over the real run → `DeriveCaseStatus == PASS` |
| `TestWitnessLaunchedProcessCalibration` | **PASS** | positive: resolved == pinned; negative: byte-divergent → distinct `errLaunchedProcessDigestMismatch` |
| `TestWitnessLaunchedProcessDeleted` | **PASS** | `(deleted)` marker on the running image is rejected |
| `TestWitnessLaunchedProcessFailClosed` | **PASS** | non-positive pid and malformed expected digest fail closed |

`go test` reported `ok github.com/kamisrini/proofbound/kernel/internal/twin` (exit 0) for both
invocations.

## Key live observation — the launched-process witness

```
runtime_fixture_test.go:115: REAL postgres launched; /proc witness resolved
  ba2447e5a38e52d801d0faecc8a3ee46d4cb57792cabc6e80801af4d4888fcd5 (pid 451432); proof vera.replay.v1
```

- **Observed `resolved_sha256`:**
  `ba2447e5a38e52d801d0faecc8a3ee46d4cb57792cabc6e80801af4d4888fcd5`
- This **equals** the declared `PostgresBinarySHA256` **and** the committed launch-witness
  `resolved_sha256`. The digest was produced by the single `witnessLaunchedProcess` call reading
  `/proc/<pid>/exe` of the live postmaster (N8 sole-producer path confirmed by code inspection in the
  companion verdict).

## Reproducibility — live-emitted vs. committed artifacts

Comparing the live `PROOFBOUND_P7_EMIT_DIR` output against the committed
`docs/verification/p7-replay-lock-fixtures/`:

- **Byte-identical:** `p7-b1-unit3-behavior-lock.json`, `p7-b1-unit3-replay-proof.json`,
  `p7-b1-unit3-ledger-witness.json`, `p7-b1-unit3-cleanup-witness.json`,
  `p7-b1-unit3-nofetch-declaration.json`.
- **`p7-b1-unit3-launch-witness.json`:** identical in every field except `resolved_pid` (live run:
  `451364`/`451432`; committed: `173025`) — the PID is legitimately non-deterministic per launch.
  `method`, `resolved_sha256`, `expected_sha256`, `verified`, and `twin_invariant` all match.

So the committed records faithfully reproduce from an independent live run on this host (modulo the
expected per-run PID), and the LOCKED/PASS outcomes are not forgeable declarations — they emerge from
a real launched postgres.

## Mutation sweep (full-package, archive-gated)

**Deferred (not run).** A full-package sweep exercises the replay/launch path per mutant, each run
launching an embedded postgres (~5-7s observed per live fixture), so a 100+-candidate sweep is
prohibitively slow and the coordinator directed not to block the live-run evidence on it. The
focused, calibrated `launch_witness.go` sweep (4/4 killed) is already recorded in the review packet;
the witness-function controls above independently demonstrate the witness discriminates a
byte-divergent image. Status: **attempted-scope-assessed, deferred as a separate CI/non-author task.**

## Conclusion

The **LIVE §9.3a run CONFIRMS the ACCEPTABLE verdict**
(`verdicts/p7-b1-unit3-impl-review-round1-20261105.md`). All four spec §9 item-3a observations held
live: (1) a real from-source postgres launched (not `t.Skip`); (2) the `/proc` witness
`resolved_sha256` == `ba2447e5…fcd5` == `PostgresBinarySHA256`; (3) a `LOCKED` behavior-lock and a
`PASS` case-manifest were produced from the real run and reproduce the committed records; (4) the
§5.1 witness-function controls behave (distinct mismatch on a divergent binary; deleted-image
rejected; fail-closed on bad inputs). The two LOW advisory findings in the companion verdict stand
unchanged and non-blocking.

**This record does not accept the unit or move `DEF-P7-001`** — that is the owner's action, now
supported by both a non-author ACCEPTABLE package verdict and this non-author live-run evidence.
