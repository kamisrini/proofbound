# P7 B1 unit-3 — independent non-author review packet (2026-11-05)

**Purpose:** package the implemented unit-3 prepared-offline runtime fixture for an **independent
non-author** verdict (Law 9) and the **§9.3a live acceptance run** (owner). The agent authored this
work and **cannot self-accept** it.
**Change intent:** `CI-p7-runtime-fixture-1a2b3c` (owner-accepted 2026-11-05).
**Spec:** [`docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md`](../plans/P7-postgres-b1-unit3-runtime-fixture-spec.md) (round-6 ACCEPTED).
**Deferral:** [`DEF-P7-001`](../deferrals.tsv) — stays OPEN/BLOCKED until this unit is accepted.

## What was implemented (commit range `d04246d`..HEAD)

| Area | Commit | What |
|---|---|---|
| intent acceptance | `d04246d` | owner flipped `CI-p7-runtime-fixture-1a2b3c` proposed→accepted |
| §9.2 F5 | `457cef5` | from-source build + subset compatibility proof + prepared-archive manifest |
| TWIN-INV-18 | `7c2b6c7`,`704418f` | launched-process witness + calibration; wired as sole producer; real-archive fixture |
| TWIN-INV-19/20/21 | `9bda47f` | no-fetch declaration, ledger-preservation, cleanup witnesses + controls |
| TWIN-INV-22 | `9f1a0f9` | behavior-lock authored over the real proof → `LOCKED` |
| case-manifest | `9d07b5f` | nine-outcome case-manifest over the real run → `PASS` |

## The soundness chain to verify (spec §1.1)

> **the fixture emitted `proof_binding = PASS`** ⟹ at run time the live postmaster's `/proc/<pid>/exe`
> digest == `PostgresBinarySHA256` ⟹ **a real from-source `postgres` ran this replay** — because the
> fixture refuses to author PASS otherwise.

The committed `case_status = PASS` and `LOCKED` are *records*; trust comes from (a) the fixture gate,
(b) the §9.3a live run, (c) the §5.1 controls — NOT from the committed artifacts self-validating.

## Review checklist (code — Law 9 governed-kernel review)

1. **`kernel/internal/twin/launch_witness.go`** — `witnessLaunchedProcess` is the single
   implementation and sole producer of `resolved_sha256`: it reads `/proc/<pid>/exe`, rejects
   `(deleted)`, hashes the ACTUAL running image via the kernel magic symlink, and returns the
   **distinct** `errLaunchedProcessDigestMismatch` (separate from the staging re-hash) on mismatch.
   Confirm the returned digest is always freshly computed from `/proc`, never copied from the
   expected value (**N8**).
2. **`kernel/internal/twin/replay.go`** — the witness is called on the
   `ReplayIsolatedWithVerifiedPostgres` path while the postmaster is live; the PID is found under the
   replay-owned temp root (`findPostmasterPID`); mismatch fails the replay; the resolved digest is
   surfaced on `Result.launch` and is the ONLY source of `proof_binding`'s `resolved_sha256` (**N8**
   code-inspection — confirm no literal / copy of the expected digest reaches the case manifest).
3. **`kernel/internal/twin/*_fixture_test.go`** — the fixture authors each outcome PASS only from a
   verified witness; `proof_binding` is gated on TWIN-INV-18, not determinism (**N3**).
4. **SPEC** `kernel/internal/twin/SPEC.md` TWIN-INV-18..22 + `docs/invariants.lock`.

## §9.3a live acceptance run (owner / non-author — the gate)

The prepared archive is NOT committed (binary-exclusion). On a host with the archive at
`$PROOFBOUND_P7_PG_ARCHIVE` (digests must match `postgres-16.4-subset-archive.json`):

```
cd kernel
PROOFBOUND_P7_PG_ARCHIVE=/path/to/postgres-16.4-subset.txz \
  go test ./internal/twin -run 'TestRuntimeFixture|TestEmitReal' -v
go test ./internal/twin -run TestWitnessLaunchedProcess -v   # calibration, no archive
```

MUST observe (spec §9 item 3a): a real postgres launched (not `t.Skip`); the `/proc` witness
`resolved_sha256` == `PostgresBinarySHA256`; the witness-function control returns the distinct
mismatch error on a byte-divergent binary and passes on the pinned one; and **by code inspection**
that `resolved_sha256` is produced by the single `witnessLaunchedProcess` call, not a literal (**N8**).
An acceptance that only runs the deterministic (no-archive) tests has NOT exercised the witness.

## Deterministic re-validation (runs in `make check`, no archive)

```
cd kernel
go test ./internal/twin -run 'TestCommittedBehaviorLockIsLocked|TestCommittedCaseManifestIsPass' -v
```
`TestCommittedBehaviorLockIsLocked` → `EvaluateLock == LOCKED` over the committed lock + a working
intent-resolver negative control. `TestCommittedCaseManifestIsPass` → `DeriveCaseStatus == PASS` AND
re-hashes every referenced evidence file against the committed digests.

## Mutation sweep (calibration)

**Focused, calibrated sweep of the new soundness file `launch_witness.go`** (2026-11-05):

```
MUTANT_TEST_RUN='TestWitnessLaunchedProcess|TestCalibration' make mutants PKG=internal/twin
```

- **`launch_witness.go`: 4 candidates, 4 killed, 0 survived.** The `witnessLaunchedProcess`
  conditionals (invalid pid, malformed expected digest, `(deleted)` rejection, `/proc` hash error)
  are all killed by the direct calibration tests (`TestWitnessLaunchedProcessCalibration`,
  `…Deleted`, `…FailClosed`).
- **Harness calibration behaved:** `lethal=killed`, `neutral=survived`, `invalid=invalid` — the
  sweep's own controls validate the harness (the stated positive/neutral-control requirement).
- **Scope note (honest):** the run restricts the per-mutant test set to the fast, no-postgres
  witness + calibration tests, so the package-wide summary (`candidates=110 killed=11 survived=99`)
  shows 99 survivors in the OTHER `twin` files (`replay.go`, `verified_archive.go`, `calibration.go`)
  that those tests do not exercise. Those are **out of scope for this focused sweep** — they are
  covered by the twin package's existing suite and by the archive-gated unit-3 fixtures, not by the
  witness-only run. A full-package sweep that exercises them must run the embedded-postgres suite
  (and, for the launch path, the archive fixtures) per mutant; that is a non-author/CI task, not a
  `make check` step. The new launch-path wiring in `replay.go` (`findPostmasterPID`, the witness
  call, `Result.launch`/`ownedRoot`) is exercised by the archive-gated fixtures + the §9.3a run.

## Known residuals (declared, not hidden)

- **TWIN-INV-19 no-fetch is a labeled DECLARATION**, not a no-egress network-namespace observation
  (per [`VD-p7-nofetch-witness-2026-11-04`](../decisions/VD-p7-nofetch-witness-2026-11-04.md); netns
  primary not wired in this build).
- **Dynamic-dependency provenance** is out of scope (F3): the witness proves the `postgres`
  executable image, not ambient host `.so` files.
- The behavior-lock's `candidate_digest` is the **manifest-scheme** digest over authored candidate
  byte identities, not the twin proof's candidate digest — the lock is its own artifact (TWIN-INV-22
  is determinism, not a runtime witness). The real-run binding is via source/baseline/candidate-
  snapshot digests + the launched-process witness.
- `findPostmasterPID` + the `replayIsolated` witness wiring are exercised only by the archive-gated
  fixtures (and the §9.3a run), not by `make check`.

## Only after acceptance
`DEF-P7-001` moves toward RESOLVED **only** on this unit's package acceptance (non-author ACCEPTABLE
verdict + the §9.3a live run). Spec acceptance alone did not close it; implementation acceptance is
required.
