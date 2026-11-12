---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-impl-review-round1-20261105
status: ACCEPTABLE
reviewed_commit: 15df68773d7908b08024e086eb6756e304a094f2
findings:
  - finding_id: p7-b1-unit3-impl-review-round1-F1
    severity: LOW
  - finding_id: p7-b1-unit3-impl-review-round1-F2
    severity: LOW
artifact_path: docs/verification/verdicts/p7-b1-unit3-impl-review-round1-20261105.md
artifact_sha: a1c3882575efcd951500a020d32cd20e242fbb6fca62b59b3d3ba6bd5721def9
---

# P7 B1 unit-3 runtime-fixture — implementation review (round 1) — non-author verdict — ACCEPTABLE

## Provenance

- **Under review:** the implemented P7 B1 unit-3 prepared-offline PostgreSQL runtime fixture
  (`CI-p7-runtime-fixture-1a2b3c`), commit range `d04246d..HEAD`, at **HEAD
  `15df68773d7908b08024e086eb6756e304a094f2`** (unchanged at dispatch and at verdict write).
- **Against:** the round-6 ACCEPTED spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (§1.1 soundness chain, §5/§5.1
  witnesses + N3/N8, §9 gates) and the Build Laws (CLAUDE.md).
- **Reviewer identity:** an independent **non-author** dispatched under Build Law 9. The reviewer
  did not author any unit-3 code, spec, or evidence. This is a Law-9 governed-kernel review of the
  fixture source plus deterministic re-validation; the acceptance capacity is the non-author's.
- **Scope boundary:** this verdict accepts the **deterministic half** that `make check` can reach
  (the committed behavior-lock/case-manifest records, the `witnessLaunchedProcess` function and its
  calibration, the N8 sole-producer code path, and Law compliance). It does **NOT** substitute for
  the §9.3a **live acceptance run**, which only the owner can perform because the prepared archive
  is binary-excluded and absent here. See "What remains for the owner's §9.3a run" below. Package
  acceptance + `DEF-P7-001` movement require BOTH this verdict AND that live run.

## What was verified (and how to reproduce)

### 1. N8 — `witnessLaunchedProcess` is the sole producer of `resolved_sha256`. CONFIRMED.

Data-flow trace (code inspection, the §9.3a-required N8 check performed here):
- `kernel/internal/twin/launch_witness.go:32-55` — the single implementation. On every success it
  returns `resolved`, freshly computed by `fileSHA256("/proc/<pid>/exe")` (line 47), after reading
  the magic symlink (line 40) and rejecting the `(deleted)` marker (line 44). On
  `resolved != expectedSHA256` it returns the distinct sentinel `errLaunchedProcessDigestMismatch`
  (line 52) and **no digest**. The returned value is never the expected literal — it is the hash of
  the running image.
- `kernel/internal/twin/replay.go:197-201` — the one call site on the
  `ReplayIsolatedWithVerifiedPostgres` path. `witness.resolvedSHA256` is assigned **only** from the
  function's return `resolved` (line 201); on `wErr != nil` the replay returns an error (line 199)
  and no `Result` is produced. `expectPostgresSHA256` is `input.PostgresBinarySHA256` (line 150),
  validated non-empty by `validateVerifiedPostgresInput`, so the witness always runs on that path.
- `kernel/internal/twin/case_manifest_fixture_test.go:104` — the committed `resolved_sha256` is
  emitted from `result.launch.resolvedSHA256`, **not** `input.PostgresBinarySHA256` (which is
  emitted separately as `expected_sha256`). The gate at line 91 additionally refuses to author PASS
  unless `result.launch.method == "proc-exe"` and the resolved digest matches.
- A repo-wide grep (`resolved_sha256` / `resolvedSHA256`) finds exactly one producer
  (`replay.go:201`) and read-only consumers in tests. `verified_archive.go:271-272` is the
  **pre-launch staging re-hash** (TWIN-INV-14) with a *different* error string — it is not the
  witness, and the SPEC (TWIN-INV-14/18) says so.

The `resolved_sha256` value is, as the spec concedes, non-discriminating on the positive path (it
is tautologically equal to the expected digest because the function errors out otherwise). The N8
guarantee is the single bound, calibrated, source-reviewed call site — which holds. No literal and
no second path exist.

### 2. N3 — `proof_binding = PASS` is gated on the launched-process witness, not determinism. CONFIRMED.

A wrong `/proc` image cannot reach PASS: `replayIsolated` (replay.go:191-200) runs the witness while
the postmaster is live and returns an **error** on mismatch, so no `Result`, no proof, no manifest
is authored. Determinism (`EvaluateLock == LOCKED`, TWIN-INV-22) is evaluated separately in the
behavior-lock test and is necessary-but-not-sufficient: a run that reproduces the proof digests but
whose `/proc/<pid>/exe` diverges fails before any PASS is authored. The committed case-manifest
routes TWIN-INV-18 (launch-witness) into the `proof_binding` outcome (`case_manifest_fixture_test.go:174`),
not determinism alone.

### 3. "Cannot fail" / forgery controls — load-bearing, confirmed where reachable.

- **Witness-function calibration** (`launch_witness_test.go:62-95`): launches a LIVE process over a
  copy of `/bin/sleep`, asserts the positive case returns the `/proc`-resolved digest with no error,
  and the negative case (a byte-divergent real binary) returns the DISTINCT
  `errLaunchedProcessDigestMismatch` (not the deleted-image sentinel). This is a genuine positive +
  neutral/negative calibration. A mutant that returned `expectedSHA256` instead of `resolved` would
  make the negative control pass with nil error → test fails → mutant killed (matches the packet's
  4/4-killed focused sweep). Deleted-marker and fail-closed branches are directly tested
  (`launch_witness_test.go:112-135`).
  Reproduce: `cd kernel && go test ./internal/twin -run TestWitnessLaunchedProcess -v`.
- **Committed re-hash is load-bearing** (confirmed empirically): appending one byte to
  `docs/verification/p7-replay-lock-fixtures/p7-b1-unit3-ledger-witness.json` makes
  `TestCommittedCaseManifestIsPass` FAIL at `case_manifest_fixture_test.go:264`
  (`witness "ledger-preserved" digest mismatch`). The tree was reverted with `git checkout --`;
  `git status` is clean. So the deterministic test is not a vacuous pass — it binds the committed
  bytes.
  Reproduce: tamper a referenced evidence file by one byte, run
  `go test ./internal/twin -run TestCommittedCaseManifestIsPass`, then `git checkout --` the file.
- **Ledger-preservation / cleanup positive controls** (`runtime_fixture_test.go:148-204`): code
  inspection shows a real mutation (`appendDeployment`) that must change the digest, and a suppressed
  `removeTempTwinRoot` that must leave the owned root present (stat-for-absence, not "RemoveAll was
  called"). Both are genuine — but they are **archive-gated** and were NOT executed here (§9.3a).

### 4. Resolver (N4) — genuinely resolves, with teeth. CONFIRMED (one coverage gap, F2).

`committedRequirementResolver` (`behavior_lock_fixture_test.go:45-76`) reads the committed
`docs/intent/records/requirements/BR-replay-evidence-6e3d2b/0001.md`, recomputes the record
self-digest and requires it to equal the declared `artifact_sha256`, checks `requirement_id`, and
requires obligation `O-1` to be `active` — failing closed on any miss. It is not a stub: the LOCKED
positive path exercises the full read+digest+obligation logic (the committed lock binds the real
record digest via `requirementArtifactSHA256`, Law 2 — not hardcoded). The negative control
(`behavior_lock_fixture_test.go:285-291`) points the resolver at an empty dir and asserts the
evaluation drops below LOCKED, proving the intent gate has teeth (spec §5 requirement satisfied).
F2 notes the other fail-closed branches lack a dedicated deterministic negative control.
Reproduce: `go test ./internal/twin -run TestCommittedBehaviorLockIsLocked -v`.

### 5. Honest labeling (N5) — CONFIRMED.

TWIN-INV-19 no-fetch is consistently labeled a **declaration**, not a network-namespace observation:
the committed `p7-b1-unit3-nofetch-declaration.json` (`method: "declaration"`, explicit note citing
`VD-p7-nofetch-witness-2026-11-04`), the SPEC TWIN-INV-19 text, and
`runtime_fixture_test.go:206-232` (asserts only the fetch-short-circuit condition + offline open).
No over-claim against spec §7 non-goals was found: dynamic-dependency provenance is excluded, the
behavior-lock `candidate_digest` is honestly described as the manifest-scheme digest (not the twin
proof's), and TWIN-INV-22 is labeled determinism-only in both SPEC and test comments.

### 6. Build Laws — CONFIRMED.

- **Law 1 (generated freshness):** `bash scripts/gen-invariants-lock.sh` produces NO diff
  (`git diff --stat docs/invariants.lock` empty); TWIN-INV-18..22 are present
  (`docs/invariants.lock:234-238`). The §6 evidence artifacts correctly carry no `@generated` marker
  under the spec §6.1 reasoned exemption (offline-archive input out of CI).
- **Law 6 (spec before code):** the spec was ACCEPTED (round 6, `9393af4`) before the implementation
  commits; TWIN-INV-18..22 are in the SPEC with proving-test derivations.
- **Skip registrations:** all six archive/emit-gated tests are registered in
  `docs/allowed-skips.txt:44-49` with honest reasons, so skip-lint does not flag them.
- **Law 2 (no datum copied):** the requirement digest is read from the record, the toolchain
  identity is cited (not copied), and `findPostmasterPID` deliberately avoids coupling to the store
  package's data-dir convention.
- **`make kernel-check` is green** (exit 0, "0 issues") — golangci-lint + the kernel suite pass.
  Reproduce: `make kernel-check`.

## Calibration (which controls I confirmed load-bearing)

- `witnessLaunchedProcess` negative control (byte-divergent live binary → distinct error): **run and
  load-bearing.** Positive control (pinned image → resolved, no error): **run and load-bearing.**
- Deleted-marker + fail-closed (invalid pid, malformed digest) branches: **run and load-bearing.**
- `TestCommittedCaseManifestIsPass` evidence re-hash: **empirically load-bearing** (one-byte tamper
  → FAIL, reverted clean).
- `TestCommittedBehaviorLockIsLocked` intent negative control (unresolvable requirement → not
  LOCKED): **run and load-bearing.**
- Ledger-preservation and cleanup positive controls: **load-bearing by code inspection only** —
  archive-gated, NOT executed here.

## What remains for the owner's §9.3a live acceptance run (NOT verified by me)

The prepared archive (`PROOFBOUND_P7_PG_ARCHIVE`) is binary-excluded and absent, so I could NOT:
1. Launch a real from-source postgres (the fixtures `t.Skip` — a green skip is not acceptance). The
   launch wiring — `findPostmasterPID`, the `witnessLaunchedProcess` call in `replayIsolated`,
   `Result.launch`/`ownedRoot`, and `TestRuntimeFixture*` / `TestEmitReal*` — is exercised ONLY by
   the archive-gated path, never by `make check`.
2. Observe live that the `/proc` witness `resolved_sha256` == `PostgresBinarySHA256`
   (`ba2447e5…fcd5` in the committed manifest).
3. Exercise the §5.1 live controls end-to-end: binary-absent → `UNVERIFIED`; tampered-archive →
   step-2 rejection; and the ledger/cleanup positive controls firing against a real run.
4. Run the full-package mutation sweep over `replay.go` / `verified_archive.go` (the packet's 99
   package-wide survivors are in archive-gated code, declared out of scope for the focused sweep).

The owner MUST, per spec §9.3a: set the env var and witness `LOCKED + PASS` emerge live; run
`TestWitnessLaunchedProcess` and confirm the distinct mismatch error on a divergent binary + pass on
the pinned one; and confirm the committed `resolved_sha256` equals `PostgresBinarySHA256`. The N8
code-inspection half of §9.3a I have performed here and confirm (section 1 above).

## Overall

**ACCEPTABLE.** The implementation faithfully realizes the six-round-hardened spec: `proof_binding`
is gated on a single, calibrated, sole-producer `/proc` witness with no literal or second path (N8);
a wrong launched image cannot reach PASS (N3); the deterministic committed records are non-vacuous
(empirically load-bearing re-hash) and honestly labeled as records, not self-proving gates; the
intent gate has teeth; and the Build Laws (generated freshness, spec-before-code, skip
registrations, Law 2, kernel-check) hold. The two findings are LOW/advisory and do not block
acceptance; neither introduces a new unsound claim. Package acceptance and `DEF-P7-001` movement
remain gated on the owner's §9.3a live run, which this non-author review does not and cannot
substitute for.


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - finding_id: p7-b1-unit3-impl-review-round1-F1
    severity: LOW
    summary: the deterministic make-check gate (TestCommittedCaseManifestIsPass) re-hashes evidence files and folds authored statuses but never cross-checks the committed launch-witness's internal consistency (resolved_sha256 == expected_sha256 == archive manifest postgres_binary_sha256, verified==true) — a cheap, in-scope hardening; BY DESIGN per spec §1/§1.1/§5 (committed artifacts are records, not self-proving), so advisory, not a defect
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-impl-review-round1-F2
    severity: LOW
    summary: the deterministic intent-gate negative control (TestCommittedBehaviorLockIsLocked) exercises only the missing-record fail-closed branch of committedRequirementResolver (empty TempDir); the self-digest-mismatch, inactive-obligation, and requirement_id-mismatch branches are not covered by a dedicated deterministic negative control, though all execute on the LOCKED positive path — spec required only "an unresolvable requirement must drop below LOCKED", which is satisfied
    verdict: CONFIRMED
```