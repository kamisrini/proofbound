# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-05

## Resume — 2026-11-05

**STANDING AUTHORIZATION IN FORCE** ([[VD-agent-standing-authorization-5ouz7q]], 2026-10-31): the
owner (Srinivas Kamisetty) authorized a product push. The agent self-authorizes governed commits
under the owner's name, works ACROSS the normally-dormant P7/deferral gates, and stops only at
genuine forks or irreversible/outward actions (shipping/redistributing a third-party binary, a
license/provenance go/no-go, deletion, publishing). Owner remains the non-author acceptor (Law 9);
adversarial rounds run via dispatched non-author reviewers. **This overrides the "P7 and OPEN
deferrals are dormant until owner authorizes" stop for the duration of the push** — do not re-apply
it blindly.

**2026-11-05 — OWNER ACCEPTED the change intent; unit-3 implementation is UNDERWAY.** The owner
("lets go", 2026-11-05) accepted `CI-p7-runtime-fixture-1a2b3c` as the non-author acceptor (Law 9);
the record was flipped `proposed`→`accepted` with its self-digest recomputed and committed
(`d04246d`, non-governed). **§9.2 is DONE and F5 is CLOSED** (`457cef5`): built PostgreSQL 16.4
minimal-permissive from the unit-1 SHA-256-verified source and proved embedded-postgres v1.31.0 runs
from the symlink-free SUBSET (initdb + loopback start + `select version()`=16.4 + clean stop, no
download). **Empirical §8.1 finding:** the spec's "expected `lib/` empty" is overturned — initdb
post-bootstrap loads `$libdir/dict_snowball`, so the subset includes `lib/postgresql/*.so` (server
backend modules); pgxs + client-only top-level libs excluded. Committed prepared-archive manifest
`docs/verification/p7-replay-lock-fixtures/postgres-16.4-subset-archive.json` (archive_sha256
`fa1ade9b…`, size 4145552, bin/postgres `ba2447e5…`, 681-entry inventory) + evidence
`docs/verification/p7-replay-lock-prerequisites/postgres-16.4-subset-compat-20261105.md`.

**ARCHIVE LOCATION (needed for the §9.3a live run; NOT committed — binary-exclusion):** the prepared
`.txz` is at `~/proofbound-p7-build/postgres-16.4-subset.txz` on this WSL box (full build tree under
`~/proofbound-p7-build/`). The §9.3a acceptance run MUST set `PROOFBOUND_P7_PG_ARCHIVE` to exactly
that file (its digests pin this prefix-specific build).

**§9.3 increments 1–2 DONE (the soundness core, end-to-end VERIFIED).**
- **Increment 1 (`7c2b6c7`):** `witnessLaunchedProcess(pid, expectedSHA256)` — single sole-producer
  of the launched-process digest; resolves `/proc/<pid>/exe`, rejects `(deleted)`, hashes the running
  image via the kernel magic symlink, returns the **distinct** sentinel
  `errLaunchedProcessDigestMismatch` (separate from the staging re-hash — N6). **Direct calibration
  unit test (N5)** against a live process: positive resolves, byte-divergent expected → distinct
  error; plus deleted-image + fail-closed controls. Lint clean.
- **Increment 2 (`44fa028`):** wired as the **sole producer** on `ReplayIsolatedWithVerifiedPostgres`
  (postmaster PID found under the replay-owned temp root via `findPostmasterPID` — no coupling to
  store's data-dir convention, Law 2; surfaced on `Result.launch`; replay FAILS on mismatch). Runtime
  fixture `runtime_fixture_test.go` (spec §4): reads `PROOFBOUND_P7_PG_ARCHIVE` (t.Skip when unset —
  UNVERIFIED, no PASS by absence), builds `VerifiedPostgresInput` from the committed manifest,
  launches a real from-source postgres for BOTH source + candidate stores (fetch-free). **VERIFIED
  end-to-end locally:** real postgres launched, `/proc` witness resolved `ba2447e5…` == declared
  `bin/postgres`, real `vera.replay.v1` proof produced. SPEC: TWIN-INV-18 added, TWIN-INV-14 note
  corrected, `invariants.lock` regenerated; all invariant/numbering lints pass.

**§9.3 increment 3 DONE (`9bda47f`) — TWIN-INV-19/20/21, all VERIFIED end-to-end against the real
archive.** TWIN-INV-20 whole-ledger preservation (before/after source digest; mutation control flips
it); TWIN-INV-21 cleanup absence (`replayIsolated` now surfaces `Result.ownedRoot`; witness stats it
for ABSENCE; suppressed-removal control leaves it present); TWIN-INV-19 no-fetch as the **labeled
declaration** per [[VD-p7-nofetch-witness-2026-11-04]] (staged `bin/pg_ctl` short-circuits the fetch;
no remote strategy constructed; recorded as a configuration assertion, NOT a namespace observation —
netns primary not wired). SPEC TWIN-INV-19/20/21 added, `invariants.lock` regenerated, skips
registered, all lints + `make kernel-check` green.

**So four runtime witnesses (18/19/20/21) are implemented and end-to-end verified against a real
from-source postgres.** The novel/risky core of unit 3 is done.

**§9.3 increments 6–9 DONE — unit-3 IMPLEMENTATION COMPLETE and self-verified (2026-11-05).**
- **(6) TWIN-INV-22 (`9f1a0f9`):** behavior-lock authored over the real proof → `EvaluateLock ==
  LOCKED`; committed `p7-b1-unit3-behavior-lock.json` + `…-replay-proof.json`; deterministic
  `TestCommittedBehaviorLockIsLocked` (runs in `make check`, no archive) + a negative control proving
  the intent-resolver gate has teeth.
- **(7) case-manifest (`9d07b5f`):** nine-outcome `p7-b1-unit3-case-manifest.json` over the real run →
  `DeriveCaseStatus == PASS`; four witness records committed; `TestCommittedCaseManifestIsPass`
  re-parses + RE-HASHES every referenced file against its committed digest.
- **(8) §6 evidence:** committed across (6)/(7) (proof, lock, case-manifest, 4 witness records).
- **(9) calibrated mutation sweep + packet (`64933c9`):** focused `launch_witness.go` sweep — 4/4
  killed, harness controls behaved (lethal killed, neutral survived); non-author review packet
  `docs/verification/p7-b1-unit3-review-packet-20261105.md`.

**Everything is gate-green (`make kernel-check` + lints), 14 commits (`d04246d`..`64933c9`), UNPUSHED.**

**Non-author verdict round-1 ACCEPTABLE (`1fe8146`, reviewed `15df687`).** An independent dispatched
non-author reviewer (Law 9) adversarially reviewed the implementation and returned **ACCEPTABLE, no
blocking findings** — verdict `docs/verification/verdicts/p7-b1-unit3-impl-review-round1-20261105.md`
(`vera.verdict.v1`, integrity gate green). Confirmed with evidence: N8 sole-producer, N3 gating, the
§5.1 controls load-bearing (empirical one-byte tamper → `TestCommittedCaseManifestIsPass` fails), the
resolver genuinely resolves (not a stub), TWIN-INV-19 honestly labeled, Build Laws honored. Two
**LOW/advisory** findings, non-blocking, recorded for optional cleanup: **F1** — the make-check gate
does not cross-check the committed launch-witness's internal consistency (by design per spec §1/§1.1,
records-not-self-proving); **F2** — the deterministic intent-gate negative control covers only the
missing-record branch. Applying F1/F2 would change the reviewed commit and need a re-verify, so leave
them unless the owner asks.

**§9.3a LIVE run DONE by the non-author (`58d8f52`) — unit-3 is now FULLY non-author accepted.** A
dispatched independent non-author ran the fixture with the archive present and witnessed LIVE
(nothing skipped): real from-source postgres launched; `/proc` witness resolved `ba2447e5…fcd5` ==
declared; LOCKED + PASS from the real run; §5.1 controls behaved; ledger/cleanup positive controls
fired. 5/6 live-emitted artifacts BYTE-IDENTICAL to the committed records (launch-witness differs
only in the per-run `resolved_pid`), so the committed LOCKED/PASS records reproduce from an
independent live run. Evidence `docs/verification/p7-b1-unit3-9.3a-live-run-20261105.md`. So unit 3
has **BOTH** a non-author ACCEPTABLE package verdict (`1fe8146`) AND non-author live-run evidence.
Full-package mutation sweep deferred (embedded-postgres per mutant, prohibitively slow; focused
`launch_witness.go` 4/4 + live witness-function controls stand).

**P7 B1 is COMPLETE. `DEF-P7-001` is RESOLVED (owner go/no-go 2026-11-05,
[[VD-p7-def001-resolve-2026-11-05]], committed `daf9fe4`).** The batch through unit-3 acceptance was
pushed to `origin/main` (`0397bcc..af0467b`, 2026-11-05); the DEF-P7-001 resolution commits sit after
that (push them too). The three B1 units are all non-author accepted; the only remaining OPEN
deferral is the unrelated [[DEF-KERNELCHECK-FLAKE-001]] (embedded-postgres port-contention flake).
RESOLVED closed the provenance/licensing-EVIDENCE deferral — it did NOT change the binary-exclusion
law and is NOT a release authorization; actual redistribution is still a separate operational step.

**#2b (witnessed regeneration) STARTED — owner authorized 2026-11-05 ("lets finish #2b next").** The
forward plan is [`docs/plans/product-completion-plan-2026-11-05.md`](../docs/plans/product-completion-plan-2026-11-05.md)
(indexed from ROADMAP P7+). First piece, spec-first: **detector-power metric + regeneration-trust
gate** — design spec draft 1 `docs/plans/P7-2b-detector-power-spec-draft1.md` (`d110465`, docs-only,
non-governed). Metric = `killed/(candidates−invalid)` so survivors (incl. declared-equivalent) lower
the score (an equivalent survivor is un-pinned latitude a regenerated impl could satisfy-and-reopen);
a lock is regeneration-trusted only at `detector_power ≥ owner-set THRESHOLD`. The regeneration LOOP
is deferred to a follow-on spec. **Round-1 non-author review NEEDS_CHANGES** (`282229d`,
`verdicts/p7-2b-detector-power-spec-review-round1-20261105.md`): F1 (HIGH) — draft-1's
all-survivors-in-denominator was a category error that rewarded deleting defense-in-depth; + F2-F8.
**Draft 2 (`986d6d7`) addresses all of them:** survivors classified E/R/U, `detector_power =
killed/(killed+U)` counts only unproven latitude (perfect lock can reach 1.0); DP-6 freshness bind;
necessary-not-sufficient; C/K/V/S re-derived from committed raw sweep stdout; loop assumptions named;
invariants DP-1..7. **Round-2 NEEDS_CHANGES** (`07270f5`, "narrow, near-acceptance"; all round-1
findings confirmed resolved, E/R/U core accepted): N1 (gate on the ratio is wrong — size-dependent,
paddable, and operator-unreachable blind spots → U=0 → 1.0; gate on ABSOLUTE U instead), N2
(freshness scope must include `_test.go` + register + SPEC rows), N3 (bind the raw sweep to a real
execution, not a fabricable JSON), N4 (R names its pinning line; exact bijection). **Draft 3
(`73108b9`) addresses all four:** trust gate keys on absolute class-U vs a per-tier `U_BUDGET`
(highest tier U=0), ratio demoted to diagnostic; DP-6 `lock_inputs_sha256` covers tests+register+SPEC
rows; DP-8 execution-binding (reproducible or named-attested + Law-4 expiry); DP-1..8. **Round-3
non-author re-verify IN FLIGHT** (reviewer expected round 3 to accept). **Exact next action:**
await/adjudicate round-3; on ACCEPTABLE the OWNER settles §6 (per-tier `U_BUDGET`; shell-gate vs
`kernel/internal/detectorpower` package) and approves a change intent before any governed code. The
#2b-spec commits + the review verdicts are LOCAL (unpushed) — push on owner instruction.

**2026-11-05 UPDATE — detector-power SPEC is ACCEPTED (`3994115`).** Round-3 non-author verdict
ACCEPTABLE at `73108b9` (`verdicts/p7-2b-detector-power-spec-review-round3-20261105.md`); the
three-round loop (round1 NEEDS_CHANGES → round2 NEEDS_CHANGES → round3 ACCEPTABLE) resolved all
findings, and the four non-blocking clarifications (CL-1 attested=independent non-author re-run; CL-2
lock_inputs pins the operator-set; CL-3 R needs a subsumption argument; CL-4 each budgeted class-U
argues its regeneration-safety) are applied. **SPEC acceptance only** — no gate code/record/sweep is
accepted. **OWNER §6 DECISIONS CAPTURED (2026-11-05):** §6.2 = **shell gate** (`scripts/detector-power.sh` +
`docs/gates.md` row, mirroring `meta-tax.sh`), NOT a kernel package; §6.1 = **measure-first with a
provisional `U=0` gate**, then the owner tunes per-tier `U_BUDGET` from the real classified sweeps
(with the CL-4 per-U safety argument for any non-zero budget).

**Exact next action — INTENT-PROVENANCE GATE (owner-level governance; agent cannot self-create a
business decision or self-accept a change intent, Law 9 + P5).** The governed implementation
(`scripts/*`, `docs/gates.md`, `docs/allowed-survivors.txt`) needs a change intent bound to a
requirement — but **#2b regeneration has NO business decision or requirement in the intent fabric**
(the only BD/BR lines are `BD/BR-proofbound-p5` and `BD/BR-replay-evidence`; detector-power is a
*different* obligation — regeneration trust, not replay reproducibility, so binding its CI to
`BR-replay-evidence` O-1 would be false provenance). So implementation is gated on a **BD → BR → CI
chain for #2b** that the owner authors/accepts. Offer standing: the agent can DRAFT the chain (BD +
BR with obligations + `CI-p7-detector-power`, all `proposed`) for owner acceptance; the owner remains
the acceptor. Only then: implement spec-first (shell gate + self-test; add E/R/U+`pins` to the
register; run the deferred full `internal/twin` sweep + classify to seed execution-bound records;
advisory `docs/gates.md` row; fold CL-1..4) → `make mutants` + non-author package verdict → the
follow-on regeneration-loop spec.

**2026-11-05 — intent chain DRAFTED (`d8fd031`), all `proposed`, awaiting owner acceptance.**
BD-witnessed-regeneration-9c4e21 → BR-regeneration-trust-5f8a03 (O-1 detector-power gate = CI target;
O-2 regeneration loop = follow-on) → CI-p7-detector-power-1a2b3c (shell-gate scope + the §6
decisions). Self-digests reproduce; chain non-dangling. **Exact next action — OWNER ACCEPTANCE** (the
agent flips statuses as scribe, owner is acceptor, Law 9): BD `proposed`→`accepted`, BR
`proposed`→`active`, CI `proposed`→`accepted` (recompute each self-digest + the downstream digest
refs after each flip, same scheme as the records connector). Then implement the shell gate spec-first
under the accepted CI. Census (not a make-check gate) flags the rejected `kernel/internal/detectorpower`
token that survives only in the committed round-3 verdict's history.
The detector-power gate's first evidence input wants the deferred full-package `internal/twin`
mutation sweep. Optional, non-blocking: verdict round-1 F1 (make-check cross-check of the
committed launch-witness internal consistency) + F2 (more intent-resolver negative-control branches),
each needing a re-verify since they change the reviewed commit; and the unit-2 round-2 three cosmetic
license-inventory nits. The full-package mutation sweep over `replay.go`/`verified_archive.go` stays
deferred (embedded-postgres per mutant; focused `launch_witness.go` 4/4 + live controls stand).

~~Superseded: the §9.3a live run is no longer "the gate left" — it is done. Historical command below.~~
Run on a host with the prepared archive at
`~/proofbound-p7-build/postgres-16.4-subset.txz` (`PROOFBOUND_P7_PG_ARCHIVE`): `cd kernel &&
PROOFBOUND_P7_PG_ARCHIVE=… go test ./internal/twin -run 'TestRuntimeFixture|TestEmitReal' -v` +
`go test ./internal/twin -run TestWitnessLaunchedProcess -v`, witnessing LOCKED+PASS live and
confirming by code inspection that `resolved_sha256` comes from the single `witnessLaunchedProcess`
call (N8). **Only on the non-author ACCEPTABLE verdict + the §9.3a run does `DEF-P7-001` move toward
RESOLVED.** Also pending: push the batch (outward action, owner instruction).

--- superseded authoring plan (kept for reference; now implemented) ---
- **(6) TWIN-INV-22 — author `proofbound.p7.behavior-lock.v1` over the real proof → `EvaluateLock ==
  LOCKED`.** Confirmed field rules (see `behaviorlock.go`): `ReplayProof` must satisfy
  `validateReplayProofConsistency` — `source_digest`/`baseline_snapshot_digest`/
  `candidate_snapshot_digest`/`through_seq` taken from the REAL `twin.Result.Proof`; `candidate_digest`
  = manifest-scheme `candidateDigest(Candidates)` (sha256 of newline-joined candidate `bytes_sha256`),
  which is NOT the twin candidate digest — the lock is its own artifact (matches the accepted 2026-11-01
  contract; TWIN-INV-22 is determinism, NOT a runtime witness). Needs a real source prefix
  (`through_seq ≥ 1`), `baseline.projector_revision` = a 40-hex git rev (`gitRevisionRE`) +
  `projector_revision_digest` 64-hex, and a DECLARED `intent_binding` tuple carrying the REAL
  `artifact_sha256` of `BR-replay-evidence-6e3d2b`. Committed proof bytes = `marshal(lock.ReplayProof)`
  (snake_case tags — NOT the twin.Proof JSON). **Resolver:** no concrete `RequirementResolver` exists
  outside tests — the deterministic CI re-validation test must supply one that genuinely resolves
  `BR-replay-evidence-6e3d2b` O-1 (read the committed intent record, check its artifact_sha256),
  otherwise `evaluateIntent` → UNVERIFIED and never LOCKED.
- **(7) author `proofbound.p7.case-manifest.v1` → `DeriveCaseStatus == PASS`.** Nine outcome keys
  (`source_binding, baseline_provenance, candidate_binding, ledger_preserved, snapshot_comparison,
  proof_binding, no_fetch, cleanup, intent_binding`); every `CaseRef` (ContractRefs, Inputs) +
  Artifact is content-addressed to a committed file; Commands carry `network_policy` ∈
  {none,loopback-only,offline}; the FIXTURE authors each outcome PASS only from a verified witness
  (`proof_binding` gated on TWIN-INV-18, not determinism — N3).
- **(8)** commit the §6 non-binary evidence (proof bytes, lock JSON, case-manifest JSON,
  witness/command records). Design idea: emit from the fixture (archive present) to a committed
  evidence dir, plus a SEPARATE deterministic test (no archive, runs in `make check`) that re-parses
  the committed lock/case-manifest and asserts `LOCKED`+`PASS` (TWIN-INV-22 determinism).
- **(9)** calibrated `make mutants` (every survivor pre-declared; state calibration controls) + the
  independent non-author review packet.

**Law 9: the owner (or a dispatched non-author) owns the §9.3a live acceptance run + the package
verdict — the agent cannot self-accept.** The archive for §9.3a is
`~/proofbound-p7-build/postgres-16.4-subset.txz` (its digests pin that prefix; set
`PROOFBOUND_P7_PG_ARCHIVE` to it). Run the fixture suite with that env var to witness `LOCKED`+`PASS`
live; the §9.3a checks are in spec §9 item 3a.

**Review-loop outcome (durable):** R1 — closure rested on `EvaluateLock → LOCKED` (pure digest
arithmetic, `replay.go:195-199`), not a real-run witness. R2 — the fix relocated the over-claim onto
`case_status=PASS`, also a pure fold (`casemanifest.go:197-222`/`398-425`), still forgeable. R3 —
verification relocated onto the FIXTURE (committed manifest is a *record*, not a self-proving gate);
found N4. R4 — N5: launch-layer swap unimplementable without a governed seam. R5 — direct
witness-function calibration + §8.4 settled; found N8 (observable value non-discriminating; fixture
not bound to the calibrated function). **R6 ACCEPTABLE** — single sole-producer
`witnessLaunchedProcess` on the `proof_binding` path + §9.3a code-inspection; two LOW cosmetic nits
applied. Net design: `case_status=PASS` entails a real from-source postgres being the launched
process, observed via a single calibrated, source-reviewed `/proc/<pid>/exe` witness; no accidental
false-PASS in the shipped config; the one residual deliberate-forgery path reduces to a contract
violation caught by §9.3a code inspection / Law-9 kernel review.

**Implementation sequence once the owner flips the CI** (spec §9): subset prep + compatibility proof
(§9.2, closes F5 — prove embedded-postgres runs from the symlink-free `bin/`+`share/` subset, not just
the full tree) → runtime fixture + TWIN-INV-18..22 (launched-process `/proc` witness with distinct
typed error + `resolved_sha256` observable; no-fetch namespace/`unshare -rn` primary + declaration
fallback per [[VD-p7-nofetch-witness-2026-11-04]]; ledger-preserved + cleanup positive controls;
ephemeral-port + short-socket isolation) → author behavior-lock + case-manifest over the real proof →
`make mutants` (calibrated) + non-author ACCEPTABLE package verdict + the **§9.3a live acceptance run**
(archive present, witness `LOCKED + PASS` live, run the witness-function control, inspect
`resolved_sha256` and its production call site). Only on that package acceptance can `DEF-P7-001` move
toward RESOLVED.
**§8.1 (highest risk) is RESOLVED 2026-11-03:** a throwaway probe proved `fergusstrange/embedded-postgres`
v1.31.0 runs our from-source 16.4 tree via `BinariesPath` end-to-end (initdb + start on 127.0.0.1 +
`select version()` = PostgreSQL 16.4 + clean stop), with **no download** (the `bin/pg_ctl` stat
short-circuits the fetch) and a short `/tmp` socket. The staged `bin/`+`share/` subset still needs
confirming, but the core approach is validated — the probe was deleted (never committed). The
unit also carries the P1 package-acceptance bar (`make mutants` + non-author ACCEPTABLE).
B1 unit 2 (build + license/notice inventory) is **non-author ACCEPTED 2026-11-03** (round-2
ACCEPTABLE at `a4642fe`). 2026-11-03: the owner installed the build toolchain
interactively (`build-essential bison flex libreadline-dev zlib1g-dev`; gcc 11.4.0) and approved the
**minimal-permissive** config `--without-icu --without-readline` (zlib default-on). Source re-verified
first (unit-1 verifier OK + self-test green), built outside the clone (configure 43s / make 35s /
install 4s); `ldd bin/postgres` links only zlib + glibc; functional proof = initdb (libc) + loopback
start + `select version()` = PostgreSQL 16.4 + clean stop + temp-root removal. The license inventory
survived an adversarial non-author review: **round-1 NEEDS_CHANGES** caught a real MED undercount
(the `ldd`-only method is blind to third-party source **vendored and statically compiled** into the
binary — Ryu/Boost-1.0, hyperloglog/MIT, Spencer regex, ISC, Gifford SHA-2, etc.), which I fixed by
rewriting the inventory to cover **two linkage dimensions** (dynamic + static vendored, enumerated
from a per-file header scan, verbatim notices for the obligation-bearing ones); **round-2
ACCEPTABLE**. The all-permissive / no-copyleft conclusion held throughout. Three LOW/cosmetic
findings recorded as optional cleanup (truncated HLL MIT tail; slightly over-inclusive CJK row;
unlisted public-domain IANA tz). Evidence: `postgres-source-16.4-build-20261103.md` +
`postgres-16.4-license-notice-inventory.md`; verdicts `…unit2-build-round{1,2}-20261103.md`.
**Unit-3 constraint (surfaced + independently re-confirmed during the build):** the replay fixture
MUST use a SHORT Unix-socket dir — the deep scratchpad/prefix path (>107 chars) overran PostgreSQL's
107-byte socket limit; TCP bind on 127.0.0.1 was fine (not a build defect).
B1 unit 1 (source provenance) is COMPLETE and ACCEPTED, and its two LOW findings have been applied
(`d9590da`) and the fix non-author-accepted (`10b4204`, round-2 ACCEPTABLE at `95dfb47`). The owner RESOLVED `DEF-P7-001` on
2026-11-01: **Option B1** — build PostgreSQL from upstream source. Recorded as [[VD-p7-postgres-source-route-2026-11-01]] (zonky
retired, dev/demo-only; Option C rejected). **Provenance mechanism corrected same day:** discovered
during unit 1 that postgresql.org ships source tarballs with SHA-256/MD5 checksums only — NO detached
`.asc` signature, no `KEYS` file — so "verify a signature against a published PGDG key" (as first
recorded) is impossible for source. Owner chose the least-expensive, light-touch mechanism: pin the
PUBLISHED SHA-256 and verify a pristine-upstream tarball fetched over HTTPS from the official domain
(integrity + official-domain TLS trust, NOT an offline signature — owner-accepted residual). The VD
and brief carry the dated correction. Slice plan:
[`docs/plans/P7-postgres-b1-source-build-slice-plan.md`](../docs/plans/P7-postgres-b1-source-build-slice-plan.md).

**B1 unit 1 fetch/verify is DONE** (2026-11-01): PostgreSQL 16.4 `.tar.bz2` fetched over HTTPS (HTTP
200, no redirect, requested URL == effective), measured SHA-256
`971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f` == published, size 24765786 ==
listed. Manifest `docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json`; evidence
`docs/verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md`. The
fail-closed verifier scripts/p7-postgres-source-verify.sh (not backticked here; the census treats a
backticked `scripts/…` token as a mechanism that must already exist) + its scripts/tests self-test
are DONE (self-test green, wired into hooks-test; verifier re-verifies the real download).
**Unit 1 ACCEPTED** by independent non-author verdict
`docs/verification/verdicts/p7-postgres-source-verify-round1-20261101.md` (status ACCEPTABLE at
`652bb5f`; reviewer independently re-fetched the upstream checksum and ran the real verifier; ~20
break attempts, no false-accept). Both LOW, fail-closed findings are now **APPLIED** in `d9590da` (owner said "do all 3"): (1) the
verifier rejects a symlinked tarball up front (was over-reject-only; now fails closed explicitly,
with a positive self-test control that only the new `-L` check can reject — Law 3); (2) the unused
duplicate `download_sha256` was dropped from the manifest (the verifier reads `published_sha256`
only — Law-2 drift seed removed). The first non-author re-verify of `d9590da` returned **NEEDS-CHANGES** and caught a real Law-3
harness defect: the symlink positive control did NOT isolate the `-L` guard (`stat -c %s` on a
symlink reports the target-path length, so the size check rejected the fixture on its own → the
guard could be deleted with the self-test staying green; the "test that could not fail" class). The
runtime behaviour was correct (no false-accept). Fixed in `10b4204`: the control now makes the
symlink target-path length equal the content size so only the guard can reject it — proven
load-bearing (guard stripped from a copy → fixture ACCEPTED). The **round-2 non-author re-verify
returned ACCEPTABLE** at `95dfb47`
(`docs/verification/verdicts/p7-postgres-source-verify-hardening-round2-20261101.md`; the round-1
NEEDS_WORK verdict is recorded alongside as `...-hardening-round1-...`): the reviewer independently
proved the control load-bearing (stripped-guard self-test exits non-zero) and found no false-accept
across an 11-case battery. One LOW/cosmetic finding (a leftover `download_sha256` in the test's
throwaway synthetic `mkman` manifest — not a drift seed, never read) is recorded as optional cleanup.
**The unit-1 hardening is ACCEPTED.** **Unit 2 (build + license/notice inventory)** is **non-author
ACCEPTED 2026-11-03** (build `a79d53d` + inventory undercount fix `a4642fe`; build host unblocked —
owner installed the toolchain interactively, [[proofbound-p7-b1-build-toolchain-gap]] RESOLVED;
round-1 NEEDS_CHANGES → fix → round-2 ACCEPTABLE). **Unit 3** = prepared-offline runtime fixture binding a real `vera.replay.v1` proof,
which lets case-manifest/behavior-lock derive a *verified* status. `DEF-P7-001` stays **OPEN/BLOCKED**
until units 1–3 are accepted. Binary-exclusion law unchanged. Snapshot provider stays skipped
(Gate 1). The retired dev/demo `twin replay` path still uses embedded-postgres ports 55461-55463.

**Shipped this session (2026-11-01):** two non-runtime P7 contracts added to the accepted
`internal/twin/manifest` package (production count stays 17) and non-author **ACCEPTABLE** —
`proofbound.p7.case-manifest.v1` (`casemanifest.go`, the `ReplayIsolated` case evidence index) and
`proofbound.p7.behavior-lock.v1` (`behaviorlock.go`, the authored lock + pure `EvaluateLock` over
committed `vera.replay.v1` proof bytes). SPEC extended (CASE-INV-1..6 / LOCK-INV-1..6). Calibrated
mutation sweep green (335/330/0/5, all survivors pre-declared, 0 new; calibration controls behaved).
Implementation `6dc9372` (governed, `Intent: CI-p7-case-manifest-1a2b3c` + `CI-p7-behavior-lock-1a2b3c`,
both bound to `BR-replay-evidence-6e3d2b` O-1). Verdict:
`docs/verification/verdicts/p7-case-behaviorlock-round1-20261101.md`. Owner brief:
`docs/plans/P7-postgres-provenance-go-no-go-brief.md`. See `notes/journal/2026-11-01.md`.

**Immediate blockers / honest gate state:**

- **2026-11-04 (gate re-verified green on current head `4cb6343`):** the 2026-11-03 "re-run
  `make check` at next wrap" item is now DONE. Bare `make check` fails only at `commit-cadence`
  (`dirty worktree outlived 90m commit window`) — the documented structural false-positive from the
  two permanent untracked files + an old HEAD, not a content defect; because `make` aborts there it
  never reaches `kernel-check`, so the kernel half was verified directly: **`make kernel-check`
  PASS** (golangci-lint 0 issues, all packages; `DEF-KERNELCHECK-FLAKE-001` did NOT fire),
  `p6-artifact-integrity --check` PASS, `deferrals --check` PASS, and every non-kernel gate green
  individually (incl. `state-freshness`, `index-check`). `gen-index.sh` produced no diff — the
  committed `docs/decisions/INDEX.md` is current (the `index stale` line during `make check` is the
  index-check hook's own negative self-test). This wrap's notes commit resets HEAD age, so the next
  bare `make check` on a freshly-committed tree clears cadence too (until the 90m window lapses
  again against the untracked files).
- **2026-11-03:** unit-2 work is **docs/evidence only** (`a79d53d`, non-governed — `docs/plans/*`
  + `docs/verification/*`, not in the governed path set); `make link-lint` green. No `kernel/*`,
  `scripts/*`, `gates/*`, or Makefile change, so no gate code restales and the last full-run green
  still stands.
- **`make check` VERIFIED GREEN at `6078563`** (full run; golangci-lint 0 issues; this notes pin
  commit sits one past it, non-governed, and does not touch any gate). The 2026-11-01 B1 work is
  docs/notes/evidence + four governed scripts commits (`edef420` INDEX regen; `00af5f3` verifier +
  self-test; `d9590da` symlink-guard hardening; `10b4204` self-test isolation fix) — no `kernel/*`,
  so the route-matrix code-commit pin does not restale; each governed
  commit is followed by a non-governed one so HEAD stays non-governed. The governed implementation
  `6dc9372` carries its `Intent:` trailers; invariant gates
  pass (the manifest SPEC uses the table form, which does not feed `invariants.lock`); the
  `@generated` verdict-integrity audit artifact and route-matrix pin were refreshed at `c83c807`.
- **FLAKE (seen TWICE 2026-11-01):** `make check` failed at `kernel-check` on two separate full runs
  while `twin/manifest` itself PASSED; `make kernel-check` alone and `go test ./...` standalone both
  pass every kernel package, and each full re-run was green. Likely embedded-postgres port contention
  (55461-55463) or parallel/timing sensitivity in `internal/twin`/`internal/cli` under the full-run
  profile. Not caused by any docs/shell change this session. Registered (Law 7, seen twice) as
  `DEF-KERNELCHECK-FLAKE-001` (OPEN/DEFERRED) — fix is to isolate the offending test and make it
  hermetic so the gate is deterministic.
- **Push:** B1 decision batch pushed to `origin/main` (`0f44df9..7ceb29e`) on owner instruction
  2026-11-01. The B1-unit-1 batch + the `d9590da` hardening after it are committed locally; the
  owner said "do all 3" (push) 2026-11-01; the full batch through this pin commit is pushed to
  `origin/main` this session.
- `DEF-P7-001` stays **OPEN/BLOCKED** (pre-GA): shippable PostgreSQL provenance/licensing. The pinned
  zonky.io embedded-postgres v16.4.0 Maven artifact fails provenance (detached GPG signature does not
  verify — expired encryption-only subkey, mismatched published key) and ships no license/notice
  coverage despite bundling PostgreSQL + OpenSSL/ICU/libxml2/libxslt/zlib/liblzma/OSSP-UUID. It runs
  for dev/demo (`kernel/cmd/replaydemo`) but is NOT safe to redistribute. See the go/no-go brief and
  [`p7-replay-lock-prerequisites/README.md`](../docs/verification/p7-replay-lock-prerequisites/README.md).
- `commit-cadence` fails whenever the tree is dirty and HEAD is older than 90m; the two preserved
  untracked user paths (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) keep the
  tree dirty between commits, so run `make check` promptly after a commit.
- Kernel toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `~/go/bin`, resolved by `kernel-check.sh`
  via a `$(go env GOPATH)/bin` fallback (no PATH edit needed). See [[proofbound-wsl-toolchain-gap]].

## Deferrals (see `docs/deferrals.tsv`)

- `DEF-P7-001` **RESOLVED 2026-11-05** ([[VD-p7-def001-resolve-2026-11-05]]) — pre-GA shippable
  PostgreSQL provenance/licensing. B1 route ([[VD-p7-postgres-source-route-2026-11-01]]) delivered:
  unit-1 source provenance ACCEPTED, unit-2 build + two-dimension license/notice inventory non-author
  ACCEPTED (2026-11-03), unit-3 runtime fixture non-author ACCEPTABLE package verdict + non-author
  §9.3a LIVE run (2026-11-05). RESOLVED closes the provenance/licensing-evidence deferral only; the
  binary-exclusion law is unchanged and it is not a release authorization. Residuals carried in the
  VD (prefix-embedded digests; dynamic-dep provenance out of scope; no-fetch declaration; SHA-256+TLS
  source provenance; deferred full-package mutation sweep).
- `DEF-P7-002` RESOLVED (Go-module evidence corpus).
- `DEF-P7-003` RESOLVED (twin/manifest accepted as the 17th package, 2026-10-31).
- `DEF-METATAX-001` RESOLVED (meta-tax is a diagnostic, not a gate).
- `DEF-KERNELCHECK-FLAKE-001` OPEN/DEFERRED — embedded-postgres port-contention flake in the full
  `make check` run; unrelated to B1; the only remaining open deferral.

## Branch status

Branch `main`, tracking `origin/main`. `origin/main` was last pushed at `0397bcc` (2026-11-04, the
round-1 verdict). **Local is ahead by the full 2026-11-04 spec-review-loop batch** — the round-2…6
verdicts, the round-2…6 spec revisions (ending `b10ea28`, spec marked ACCEPTED), the §8.4 decision
record (`56286fd`, index regenerated), and this EOD wrap's notes/journal. All
docs/decisions/verification (non-governed; no `kernel/*`, so no gate code restales). **Not pushed** —
push is an outward action outside the standing authorization's self-authorized scope; left for owner
instruction.

## P7 position

P1–P6 deep-complete (round-C ACCEPTABLE, zero-open census). P7 is a capability phase with no single
mechanical DoD. Status: the replay-lock / behavior-lock line's **non-runtime** contracts are now all
implemented and independently accepted — projector/build manifest + source-toolchain inventory +
evidence closure + skeleton + declaration (prior sessions), and now the case-manifest and
ReplayIsolated behavior-lock. Everything beyond is runtime/provider work gated on the owner
decisions above. The snapshot provider remains deliberately skipped (Gate 1).

## Tomorrow's first action

Run `/proofbound-next`. B1 units 1–2 non-author ACCEPTED; **unit-3 runtime-fixture SPEC ACCEPTED
2026-11-04** (round-6 non-author verdict; six-round loop — see Resume). **The one blocker is an owner
action: flip `CI-p7-runtime-fixture-1a2b3c` from `proposed` to `accepted`** (the agent cannot
self-accept the intent authorizing its own governed kernel work — Law 9). Once flipped, implement
spec-first per §9 (subset compatibility proof closing F5 → fixture + TWIN-INV-18..22 →
behavior-lock/case-manifest over the real proof → `make mutants` + non-author ACCEPTABLE + §9.3a live
run); only then does `DEF-P7-001` move toward RESOLVED. §8.4 is owner-SETTLED
([[VD-p7-nofetch-witness-2026-11-04]]). The 2026-11-04 batch is unpushed (push on owner instruction).
Optional: three LOW cosmetic license-inventory cleanups (unit-2 round-2 verdict).
