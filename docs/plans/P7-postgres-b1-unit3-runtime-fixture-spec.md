# P7 B1 unit 3 — prepared-offline runtime fixture spec (DRAFT round 4)

**Status:** DRAFT round 4 — spec-first per Law 6; scoping and contract only. No implementation, no
runtime `PASS`, and no `DEF-P7-001` movement is authorized by this document. Review history:
**round 1** NEEDS_CHANGES (closure rested on `LOCKED`, pure digest arithmetic); **round 2**
NEEDS_CHANGES (N1: the fix moved closure onto `case_status=PASS`, also a pure fold — forgeable by
declaration; N2 Law-1 precedent; N3 `proof_binding` key); **round 3** resolved N1/N2/N3 but found
**N4** (the §5.1 swapped-binary control trips at the step-2 archive-hash gate before launch, so it
never exercises the TWIN-INV-18 `/proc/<pid>/exe` witness it is meant to calibrate). This round-4
revision adds the **layered, launch-layer negative control** that calibrates TWIN-INV-18 (N4). It
needs a **round-4 non-author re-verify** (and, because it touches `kernel/*`, an owner-approved
change intent) before implementation.
**Date:** 2026-11-04 (revises the round-3 draft)
**Implements:** [`VD-p7-postgres-source-route-2026-11-01`](../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1), unit 3 of the [slice plan](P7-postgres-b1-source-build-slice-plan.md).
**Builds on (accepted):** units 1–2 — the SHA-256-verified 16.4 source
([verify evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md)),
the minimal-permissive build + all-permissive license inventory
([build evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-build-20261103.md),
round-2 ACCEPTABLE), and the already-accepted verified-archive staging seam
(`internal/twin` TWIN-INV-12..17; companion-archive-controls verdict).
**Deferral:** [`DEF-P7-001`](../deferrals.tsv) (stays OPEN/BLOCKED until unit 3 is accepted).

## 0. What the revisions changed (summary)

**Round 4 (this revision), after round-3 N4:** the §5.1 controls are now **layered** so the
launched-binary-swap control calibrates the TWIN-INV-18 `/proc/<pid>/exe` witness specifically — it
stages and verifies the real archive (so the step-2 archive-hash gate passes), then swaps the
*launched* executable, so **only** the `/proc` check can reject it; the tampered-archive control is
labeled as calibrating the step-2 gate, not the witness. §9.3a now requires the accepting non-author
to run the launched-binary-swap control specifically. The round-2/3 changes below still stand.

**Round 3, after round-2 N1/N2/N3:** the witness-verification claim is relocated out
of the committed case-manifest (which `DeriveCaseStatus` folds without re-hashing) and onto the
**fixture proving-test + §9.3a live non-author run + §5.1 negative control** — the committed
`case_status = PASS` is now stated as a *record* of a verified run, not a self-proving gate, and
`DEF-P7-001` closure rests on that acceptance process (§1, §1.1, §5, §6). The §6.1 Law-1 justification
is corrected: these artifacts get no automated integrity gate (the integrity script covers only
`verdicts/`), stated as a deliberate residual (N2). `proof_binding = PASS` is explicitly gated on the
launched-process witness, not determinism alone (N3). The round-2 changes below still stand.

**Round 2, after round-1 F1–F11:**

The round-1 review found that the draft's central claim was unsound: it rested `DEF-P7-001` closure
on `EvaluateLock → LOCKED`, but `LOCKED` is computed purely from event/snapshot **digests**
(`replay.go:195-199`) and witnesses nothing about whether a real PostgreSQL ran. This revision:

- **Separates two distinct claims** (F1). `LOCKED` is a *determinism* statement about the replay
  proof and is explicitly **not** a runtime witness. The runtime fact ("a real from-source postgres
  ran") is carried by the **case-manifest**, and `DEF-P7-001` closure rests on **`case_status = PASS`
  with cryptographically bound, independently-verifiable runtime witnesses** — never on `LOCKED`
  alone. The frozen `vera.replay.v1` proof shape is **unchanged** (§7 non-goal preserved).
- **Replaces the launched-binary witness** with launched-process (`/proc/<pid>/exe`) resolution and a
  **required negative control** (F2); drops the on-disk re-hash and `version()` options as
  non-witnessing.
- **Scopes the executable-image claim honestly** against ambient dynamic dependencies (F3).
- **Makes the zero-fetch witness a real network-namespace isolation**, with any configuration-level
  assertion labeled as a declaration, not a witness (F4).
- **Settles the symlink decision as drop-all** (F6) and **downgrades the §8.1 "RESOLVED" label**
  to full-tree-only, subset unproven (F5).
- **Resolves the Law-1 status** of the committed evidence as one-time documentary evidence with an
  explicit, reasoned regeneration exemption (F7); **cites rather than copies** the unit-2 toolchain
  identity (F8).
- **Adds positive/negative controls** so the ledger-preservation, cleanup, and overall `PASS`
  proving tests can actually fail (F9, F11), and makes a **live (non-skipped) run an acceptance
  precondition** plus isolates the fixture from the known port-contention flake (F10).

## 1. Purpose — what "verified" means here, and what it does NOT

Units 1–2 established that the PostgreSQL source is authentic and the produced binary's licensing is
clean. Unit 3 is the only unit that produces a **runtime** fact: it runs a real, from-source,
minimal-permissive PostgreSQL under `ReplayIsolated`, **witnesses the launched process**, and records
that evidence. Two artifacts result, carrying two **different** claims:

- `proofbound.p7.behavior-lock.v1` `EvaluateLock` derives **LOCKED** over the committed
  `vera.replay.v1` proof bytes. **LOCKED is a determinism claim only** (the replay reproduces the
  committed proof digests); it is **not** evidence that a real postgres ran, and `DEF-P7-001` closure
  does **not** rest on it.
- `proofbound.p7.case-manifest.v1` `DeriveCaseStatus` folds to **PASS** when every outcome is
  authored PASS. **The committed manifest is a *declaration*, not self-verifying** (round-2 N1):
  `DeriveCaseStatus` is a pure worst-wins fold over author-declared `outcome.Status` and
  `ValidateCaseManifest` only checks evidence-list well-formedness — neither re-hashes a witness nor
  launches anything (`casemanifest.go:197-222`, `398-425`). So PASS in the committed artifact alone
  does **not** prove a real run; taken in isolation it is forgeable by declaration.

### What actually establishes the runtime fact (N1 — the claim's real home)

The runtime fact is established **not by the committed artifacts but by the acceptance process that
produces and checks them** — exactly as every other evidence artifact in this repo (build-evidence,
verify-evidence, verdicts) draws its trust from the non-author who ran the verification, not from the
file validating itself. Concretely, three things, none of which is the manifest's own validation:

1. **The fixture proving-test (§4, §5).** The fixture compares `/proc/<pid>/exe` of the live
   postmaster to `PostgresBinarySHA256` **at run time** and **refuses to emit `proof_binding = PASS`**
   (and so refuses to author a PASS manifest) unless that and the other live witnesses verify. The
   manifest is an *output* of a run that already verified the witnesses; it is not trusted to
   re-verify them.
2. **The §9.3a live non-author run.** The accepting non-author runs the fixture **with the archive
   present** and watches `LOCKED + PASS` emerge live — a green `t.Skip` is not acceptance.
3. **The §5.1 negative control.** A swapped/absent binary must make the fixture refuse PASS (→
   `UNVERIFIED`/`FAIL`), proving the PASS is *caused by* the real run and cannot be reached without it.

`DEF-P7-001` closure therefore rests on **(the fixture + the §9.3a live run + the §5.1 negative
control)**, which together make the committed `case_status = PASS` a faithful record of a witnessed
real run — never on the committed manifest or `DeriveCaseStatus` re-verifying anything, and never on
`LOCKED`.

### 1.1 The soundness chain (explicit, after F1/N1)

The chain is anchored on the **fixture**, not on the committed artifact:

> **the fixture emitted `proof_binding = PASS`** ⟹ at run time the live postmaster's
> `/proc/<pid>/exe` digest == `PostgresBinarySHA256`, the no-fetch isolation held, the source ledger
> was byte-preserved, and the owned temp root was removed — i.e. **a real from-source `postgres`
> executable ran this replay** — because the fixture *refuses to emit that PASS otherwise*.

The committed `case_status = PASS` is then a faithful *record* of that fixture run; it is not itself
the proof (N1: `DeriveCaseStatus` re-verifies nothing). Two independent checks make the record
trustworthy to a non-author:
- the **§9.3a live run**: the accepting non-author re-runs the fixture with the archive present and
  sees `LOCKED + PASS` emerge live (a `t.Skip` is not acceptance); and
- the **§5.1 negative control** (F11): with the declared binary absent the fixture skips (case
  `UNVERIFIED`); with a *swapped/wrong* binary the launched-process witness FAILS and the fixture
  yields `proof_binding = FAIL` → `case_status = FAIL` — never `PASS`.

If the proving test cannot make the fixture refuse `PASS` when the real run is removed or tampered,
the acceptance proves nothing (Law 9 corollary).

## 2. The exact gap unit 3 closes

The `internal/twin` SPEC already ships and accepts the *preflight/staging* seam but explicitly defers
the runtime half:

- TWIN-INV-14: "A future runtime witness must additionally show that the launched `postgres` bytes
  equal the declared digest; **the current seam does not establish that runtime fact.**"
- SPEC §Interface note: "The current implementation seam does not observe the daemonized process's
  launched executable; launched-process observation and binding to the declared digest are **future
  runtime requirements** … real PostgreSQL runtime fixtures and no-fetch witnesses remain pending."
- `internal/twin/manifest` SPEC: the case-manifest/behavior-lock are "non-runtime authored" and
  "cannot be rendered as a runtime `PASS`"; a case with no produced artifacts "cannot be `PASS`."

Unit 3 supplies exactly those missing runtime artifacts. The **launched-process observation**
(TWIN-INV-18, §5) is what genuinely closes the TWIN-INV-14 deferral — the existing on-disk re-hash
(`verified_archive.go:267-273`) does **not**, and the twin-SPEC amendment must say so honestly (§5.2).
Unit 3 adds **no** new provider, event kind, feed, or PostgreSQL route, and does not touch the frozen
`ReplayIsolated` subject or the `vera.replay.v1` proof shape.

## 3. Prepared-offline archive production (the prep step)

The runtime input is a verified PostgreSQL binary tree. By the binary-exclusion law it is **never
committed**; it is produced out-of-band on a build host and staged offline. Only its *non-binary*
identity (digests + inventory) is committed.

**Source:** the unit-2 minimal-permissive build (`./configure --without-icu --without-readline`,
zlib default-on) of the SHA-256-verified 16.4 source, re-verified before each build.

**Packaging rules** (produce a `.tar.xz` / `.txz`, the format `extractVerifiedArchive` already
consumes):
1. **Minimal subset.** Include only what embedded-postgres needs to `initdb` and run the server:
   `bin/` (at minimum `postgres`, `initdb`, `pg_ctl` — all mode 0755) and `share/` (the `initdb`
   inputs: `postgres.bki`, timezone sets, `*.sample`, etc.). Exclude `include/` (~7.2M of headers),
   the static `*.a` archives, and client-only shared libraries not loaded by the server at runtime.
   `lib/` is included **only** if the replay projector loads a backend module (a PL, a contrib
   extension); the minimal replay loads none, so the expected `lib/` subset is **empty**. Whether any
   backend-loadable module is required is confirmed empirically when the subset is first packaged
   (§8.1) — if one is, its (regular-file) bytes enter the inventory like any other file.
2. **Drop all symlinks — no dereference** (F6, settles §8.2). The staging seam fails closed on
   symlinks (TWIN-INV-13). The unit-2 install tree contains 8 symlinks, all in `lib/`
   (`libpq.so.5 -> libpq.so.5.16`, `libecpg*`, `libpgtypes*`) — all **client** libraries the server
   does not link. The prep step **drops every symlink** (none is in the `bin/`+`share/` subset) and,
   after assembling the file set, **asserts the packaged set contains no symlink** before hashing.
   Dereferencing is **not** used: it would resolve a link *at prep time, outside the staging guard*,
   which inspects archive entry paths, not what a link target pointed at — so a link with an absolute
   or `../` target could become a safe-path regular file carrying arbitrary host bytes that the guard
   would pass. Drop-all removes that surface entirely; every retained entry is a regular file of mode
   0644 or 0755.
3. **Deterministic inventory.** Compute, over exactly the packaged files, the sorted-unique
   `Inventory []VerifiedArchiveEntry{Path, Mode, SHA256}`, the whole-archive `ArchiveSHA256`, and the
   `PostgresBinarySHA256` (`bin/postgres`). These feed `VerifiedPostgresInput`.

**Committed evidence (non-binary):** a prepared-archive manifest JSON (archive SHA-256, postgres
binary SHA-256, full per-file inventory, the prep command, and a **citation** — evidence path +
revision — of the unit-2 build evidence for the toolchain identity, **not** a copy of its values
(F8, Law 2)). The `.txz` bytes themselves stay out of Git. Honest-digest caveat from unit 2 applies:
PG embeds the prefix path, so the archive/binary digests pin **this** prepared build, not a
cross-host reproducibility anchor; the reproducibility anchor remains the unit-1 **source** digest.

**Ambient dynamic dependencies (F3, honest scope).** The from-source `postgres` dynamically links
host libraries (`libc`, zlib's system `libz`, …); `--without-icu --without-readline` removed those,
but zlib default-on links the *system* `libz.so`. The archive pins `bin/postgres` bytes only, so the
running process = the declared executable image **plus ambient host `.so` files that are not
inventoried**. The launched-process witness (§5) therefore proves the **`postgres` executable image**,
not the full behavior of the process; dynamic-dependency provenance is an **explicit residual of this
unit**, recorded like the honest-digest caveat (the `ldd bin/postgres` set is captured in the manifest
as *informational* evidence, not a witnessed claim). A future unit may pin dynamic deps; unit 3 does
not claim to.

**Staging location (prepared-offline):** the fixture reads the archive path from an environment
variable (proposed `PROOFBOUND_P7_PG_ARCHIVE`). No path baked into code, no network, no ambient
cache. The fixture hashes the bytes and requires `ArchiveSHA256` to equal the committed manifest
before use.

## 4. Fixture runtime contract (new behavior)

A runtime fixture (a Go acceptance test in `internal/twin`, plus any minimal prep/driver helper)
that:
1. Resolves the prepared archive from the env var. **If unset or absent → `t.Skip`**: the committed
   case stays `UNVERIFIED`; the fixture never fetches, never fabricates, never passes by absence. A
   `t.Skip` is **not** acceptance — see the §9 acceptance precondition (F10).
2. Loads the archive bytes, verifies `ArchiveSHA256` against the committed manifest, and builds a
   `VerifiedPostgresInput` from the committed inventory.
3. Assembles a **committed, synthetic** source-ledger prefix + candidate (fixtures, never presented
   as production observations — permitted by the behavior-lock SPEC) in a real source store.
4. Calls `ReplayIsolatedWithVerifiedPostgres`, which launches the real from-source server in a
   temporary embedded store on loopback, projects, compares snapshots, and returns a `Result` with a
   real `vera.replay.v1` `Proof`.
5. Produces the witnesses in §5 and the evidence artifacts in §6.
6. **Verifies every runtime witness itself (in the fixture)** and authors each outcome PASS only if
   its witness verified; then asserts `EvaluateLock == LOCKED` (determinism) and
   `DeriveCaseStatus == PASS` (the fold over the outcomes the fixture just authored), and commits the
   non-binary evidence at one evidence revision. The fixture — not `DeriveCaseStatus` — is the gate.

**Isolation & determinism (F10, short-socket from unit 2):** the fixture must not worsen
`DEF-KERNELCHECK-FLAKE-001` (embedded-postgres **port contention** on fixed ports 55461-55463). It
MUST (a) allocate an **ephemeral free port** per run (not a fixed port), (b) place the Unix socket in
a **short** directory (deep temp paths overrun PostgreSQL's 107-byte socket limit) or use loopback
TCP, and (c) run **serially**, isolated from the parallel embedded-postgres suite (e.g. its own
`go test` tag/target), so one real server never contends with another. Port/socket isolation is a
proving-test case.

## 5. Runtime witnesses (new invariants to add to `internal/twin` SPEC)

Each maps to a `case-manifest` outcome. **The fixture (not `DeriveCaseStatus`) is what verifies the
witnesses** (round-2 N1): it computes each witness at run time and authors an outcome as PASS **only**
if its witness verified; otherwise it authors FAIL / refuses to produce a PASS manifest. The committed
manifest then carries the witness records `{kind, path, sha256}` as the *evidence of* that fixture
verification — `DeriveCaseStatus` folds the author-declared statuses and does not re-hash them, so the
witness records are a durable record, not a self-executing gate. (Adding a committed-artifact gate
that cross-checks each witness `sha256` against its source would change the frozen `DeriveCaseStatus`
contract and is out of scope for this unit — it would be its own `CI-*` + acceptance; see §7.)

| New invariant | Statement | Required mechanism & control | case-manifest outcome |
|---|---|---|---|
| **TWIN-INV-18 — Launched-process witness** | the **running** postmaster's executable image equals the declared `PostgresBinarySHA256` | resolve the postmaster PID (`postmaster.pid` in the data dir), read `/proc/<pid>/exe`, **reject** if it carries the `(deleted)` marker, hash the target, require equality. The on-disk re-hash and `version()` string are **not** sufficient and are not used as the witness. **Negative control (required):** launching a *different/byte-swapped* binary MUST make this witness FAIL (and the case fold to FAIL). | `proof_binding` / provenance |
| **TWIN-INV-19 — No-fetch isolation** | the run performs no network fetch for binaries | **Primary witness:** the launch runs under a **network namespace with no egress** (`unshare -rn` / equivalent) so any fetch fails hard; `BinariesPath` is the verified local dir. **Declaration (labeled as such, not a witness):** the embedded-postgres remote-fetch strategy is never constructed. Where the platform cannot provide a namespace, the invariant is satisfied by the declaration **only if** the spec/evidence states plainly it is a configuration assertion, not an observation. | `no_fetch` |
| **TWIN-INV-20 — Whole-ledger preservation witness** | the source store's event set is byte-identical before and after the isolated run | before/after digest of the source store, recorded as a witness. **Positive control (required):** an injected operation that *does* mutate the source store MUST make the witness FAIL — proving it detects mutation, not merely restates TWIN-INV-1/-8. | `ledger_preserved` |
| **TWIN-INV-21 — Cleanup witness** | the owned temporary root (staged binaries dir + data dir) is removed on success and every failure path | the witness **stats the path and asserts absence** (not merely "RemoveAll was called"). **Positive control (required):** a path whose removal is suppressed MUST make the witness FAIL. | `cleanup` |
| **TWIN-INV-22 — Real-proof binding (determinism)** | the committed `vera.replay.v1` proof bytes reproduce under re-evaluation, making `EvaluateLock == LOCKED` | re-evaluate the lock over the committed proof bytes → LOCKED. **This is a determinism claim, NOT a runtime witness** (F1): LOCKED is a pure function of digests and does not establish that a real postgres ran — that fact is carried by TWIN-INV-18/19 and the case `PASS`. | (determinism check behind `proof_binding`, not a standalone runtime claim) |

**`proof_binding` gates on the launched-process witness, not determinism alone (round-2 N3).** The
nine case-outcome keys are frozen (CASE-INV-3), so both TWIN-INV-18 (launched-process provenance) and
TWIN-INV-22 (determinism re-eval) map to the single `proof_binding` key. The fixture MUST author
`proof_binding = PASS` **only when the launched-process witness (TWIN-INV-18) verifies**; the
determinism check (TWIN-INV-22 / `EvaluateLock == LOCKED`) is necessary but **not sufficient** for
`proof_binding = PASS`. A run that reproduces the proof digests (LOCKED) but whose
`/proc/<pid>/exe` does not equal `PostgresBinarySHA256` yields `proof_binding = FAIL`.

The remaining case outcomes (`source_binding`, `baseline_provenance`, `candidate_binding`,
`snapshot_comparison`, `intent_binding`) are produced by binding the committed source/candidate
fixtures, baseline provenance tuple, snapshot digests, and the P5 requirement
`BR-replay-evidence-6e3d2b` O-1 (resolved via the `RequirementResolver`) — the same bindings the
behavior-lock already validates, now carried into the case manifest with real evidence.

### 5.1 The overall negative control (F11)

A dedicated proving test demonstrates that **the fixture's emitted `PASS` depends on the real run**.
The controls are **layered** so each failure is attributed to the check that is actually being
calibrated (round-3 N4 — a control that fails at the wrong layer gives false confidence in the layer
it never reached):
- **binary absent** → fixture `t.Skip` → committed case stays `UNVERIFIED` (CASE-INV-5, no PASS by
  absence).
- **tampered archive** (archived bytes / inventory altered so the computed `ArchiveSHA256` ≠ the
  committed manifest) → rejected at **step 2** (archive-hash gate), *before launch*. This control
  calibrates the **step-2 archive gate**, **not** TWIN-INV-18 — it must be labeled as such and must
  **not** be relied on to show the launched-process witness has teeth.
- **launched-binary swap — the TWIN-INV-18 control (required, N4).** Stage and verify the **real**
  archive so step 2 and the staging checks pass on the correct bytes, then **swap the executable the
  postmaster actually runs for a *different, byte-divergent* postgres build at the launch layer**
  (e.g. replace `binariesDir/bin/postgres` after staging verification but before `pg_ctl start`, or
  redirect the launcher), keeping the committed archive/manifest internally consistent. Only
  TWIN-INV-18's `/proc/<pid>/exe` comparison against the pinned `PostgresBinarySHA256` can now reject
  it → the fixture authors `proof_binding = FAIL` → `case_status = FAIL`, never PASS. **Positive
  companion:** the swapped build must be one that *would otherwise run and pass every other check*
  (so the ONLY thing that fails is the `/proc` digest), and the unmodified run must reach and **pass**
  the `/proc` check — proving the witness is load-bearing, not skipped. A fixture that merely
  re-checks the on-disk/archive hash and never reads `/proc/<pid>/exe` MUST fail this control.
- **real declared binary** → the fixture verifies all witnesses (including the `/proc` check) →
  authors all outcomes PASS → `case_status = PASS`.

Because `intent_binding` uses an injected `RequirementResolver`, the proving tests MUST NOT let a
permissive stub mask a missing runtime witness: the real-run witnesses (TWIN-INV-18/19/20/21) are the
*varied key* in the controls, and a stubbed resolver may not flip them to PASS.

### 5.2 Honest amendment to the accepted twin SPEC (F2)

Amending TWIN-INV-14 must state the runtime fact is established **by the launched-process witness
(TWIN-INV-18)** — explicitly not by the pre-existing on-disk re-hash at `verified_archive.go:267-273`,
which the amendment must continue to describe as a staging (pre-launch) check only. The amendment may
not claim the runtime fact is established by any mechanism weaker than process-image resolution.

## 6. Evidence artifacts this unit commits (all non-binary)

- The prepared-archive manifest (archive + binary digests, per-file inventory, prep command,
  unit-2 toolchain **citation**, informational `ldd` set) — §3.
- The committed `vera.replay.v1` proof bytes from the real run.
- The authored `proofbound.p7.behavior-lock.v1` (LOCKED over those proof bytes — determinism) and the
  `proofbound.p7.case-manifest.v1` (`case_status=PASS`, nine outcomes with **bound** evidence).
- Witness records (launched-process, no-fetch, ledger-preserved, cleanup) as `{kind, path, sha256}`,
  each bound into the case manifest as the **record of** the fixture's run-time verification (the
  fixture computed and checked them; `DeriveCaseStatus` does not re-hash them — round-2 N1).
- Command logs (`initdb`/start/stop) with the no-fetch isolation method recorded (namespace vs.
  declaration — F4), exit codes, stdout/stderr digests.
- The runtime-fixture evidence record + the non-author verdict.

The archive `.txz`, the data directory, and any built executable are **excluded** from Git.

### 6.1 Law-1 status of the committed evidence (F7)

These artifacts are produced by a run that requires the **offline archive**, which is deliberately
absent from `make check` (the fixture `t.Skip`s without it). They are therefore **one-time captured
evidence, not regenerable derived state**, and carry **no `@generated` marker** — an **explicit,
reasoned exemption** from Law 1's regenerate-in-`make check` requirement, because the generating input
is intentionally out of CI. This is the same basis on which the existing **build-evidence and
verify-evidence** records under `docs/verification/p7-replay-lock-prerequisites/` are already
unmarked: they draw their trust from the non-author who ran and accepted the verification, not from a
regeneration.

**Honest correction (round-2 N2):** unlike the round-2 draft's claim, these artifacts are **not**
covered by `scripts/p6-artifact-integrity.sh` — that script walks only `docs/verification/verdicts/`
(`-maxdepth 1`), so build-evidence, verify-evidence, and the §6 artifacts here (prepared-archive
manifest JSON, proof bytes, lock/case-manifest JSON, witness records, command logs) get **no
automated integrity gate**; only the non-author verdict itself (which lands in `verdicts/`) is
self-digest-validated. This residual is stated plainly so it is a deliberate choice, not a silent
exemption of load-bearing artifacts: their integrity rests on the §9.3a live non-author run plus the
evidence revision recorded in the verdict, exactly as the accepted units 1–2 evidence already does.
*(Optional future hardening, out of scope here: extend `p6-artifact-integrity.sh` — or a sibling —
to self-digest-validate schema-bearing artifacts under `docs/verification/p7-replay-lock-prerequisites/`.
That would be its own governed `scripts/*` change + acceptance.)*

## 7. Non-goals

- No new provider, event kind, feed, writable truth, or PostgreSQL route; the snapshot provider
  stays skipped (Gate 1).
- **No change to the frozen `ReplayIsolated` subject, the `vera.replay.v1` proof shape, or the
  authored/observed boundary.** (The F1 fix is carried entirely by the case-manifest and witnesses,
  not by widening the proof.)
- No redistribution of the built binary; unit 3 establishes runtime evidence, not a release.
- No reproducible-across-hosts binary digest (unit-2 caveat); the source digest is the anchor.
- **No claim of dynamic-dependency provenance** (F3): the launched-process witness covers the
  `postgres` executable image, not ambient host `.so` files.
- **No claim that `LOCKED` witnesses a real run** (F1): LOCKED is determinism only.

## 8. Known risks / open decisions (settle before/within implementation)

1. **embedded-postgres from-source compatibility — RESOLVED FOR THE FULL TREE; SUBSET UNPROVEN**
   (F5, downgraded from "RESOLVED"). A throwaway probe (not committed) showed
   `fergusstrange/embedded-postgres` v1.31.0 with `BinariesPath` pointed at the **full** unit-2
   from-source install tree ran end-to-end — `initdb` succeeded, the server started on `127.0.0.1`,
   `select version()` returned `PostgreSQL 16.4 … gcc 11.4.0`, clean stop, **no download** (
   `embedded_postgres.go:155` stats `BinariesPath/bin/pg_ctl` and skips the fetch when present), unix
   socket on the compiled default `/tmp` (short). **What is NOT yet proven:** the shipped symlink-free
   `bin/`+`share/` **subset** runs `initdb`+start. The server needs only `bin/`+`share/` to
   `initdb`+start (it ignores the dropped `lib/` client libs), and `lib/` is needed only if the replay
   loads a backend module (the minimal replay loads none) — but this must be **confirmed on the
   subset** as the first implementation step (§9.2), not assumed from the full-tree probe.
2. **Symlink handling — SETTLED: drop all** (§3.2, F6). No dereference; prep asserts the packaged set
   is symlink-free before hashing. A proving test that staging rejects a symlinked archive already
   exists.
3. **Launched-process witness mechanism — SETTLED: `/proc/<pid>/exe` resolution** (TWIN-INV-18, F2)
   of the postmaster PID (from `postmaster.pid`), with the `(deleted)` marker rejected and a required
   **launch-layer** negative control that swaps the launched binary *after* the step-2 archive gate
   passes, so only the `/proc` check can reject it (§5.1, N4). The on-disk re-hash and `version()`
   string are rejected as non-witnessing. **Owner note:** this is the mechanism the review recommends;
   confirm the platform exposes `/proc/<pid>/exe` (Linux/WSL: yes).
4. **No-fetch witness mechanism — PROPOSED: network-namespace isolation** (TWIN-INV-19, F4), with the
   fetch-strategy-not-constructed assertion kept only as a labeled declaration. **Owner decision
   needed:** whether to require namespace isolation (strongest; needs user-namespace / `unshare -rn`
   support in the run environment) or accept the labeled configuration declaration where a namespace
   is unavailable. This is the one genuinely-open §8 mechanism decision for the owner.
5. **Governed paths / change intent.** New code in `kernel/*` is governed (SPEC.md is exempt). This
   unit needs an owner-approved change intent (a `CI-*` record) before the governed commit, plus the
   package-acceptance bar: `make mutants` green (survivors declared) AND a non-author ACCEPTABLE
   verdict (ROADMAP P1 bar, Law 9). The drafted `CI-p7-runtime-fixture-1a2b3c` **stays `proposed`**
   until this spec is re-accepted.

## 9. Sequence & gates

1. **Spec re-review (round 4)** — non-author re-verify of THIS revision + owner sign-off on the §8.4
   no-fetch mechanism decision; keep the change intent `proposed` until accepted.
2. **Subset prep + compatibility proof** — produce the symlink-free `bin/`+`share/` subset archive and
   **prove embedded-postgres runs from the SUBSET** (closes the F5 gap, not just the full tree).
   Commit the prepared-archive manifest.
3. **Fixture + witnesses** — implement the runtime fixture and TWIN-INV-18..22 (launched-**process**
   witness via `/proc/<pid>/exe`; no-fetch isolation; ledger-preserved + cleanup with their positive
   controls); extend the twin SPEC honestly (§5.2); author the behavior-lock + case-manifest over the
   real proof, recording the fixture-verified witnesses. **Gate:** the fixture verifies every witness
   at run time and emits `EvaluateLock == LOCKED` (determinism) AND `DeriveCaseStatus == PASS` (the
   fold over the outcomes the fixture authored from those verifications — the fixture is the gate, not
   `DeriveCaseStatus`); the §5.1 negative control proves the fixture's PASS collapses to UNVERIFIED
   (absent) / FAIL (swapped); `proof_binding` gated on the launched-process witness (N3);
   ephemeral-port + short-socket isolation holds; calibrated `make mutants` green.
3a. **Acceptance precondition (F10 + N4):** the non-author accepting this unit MUST run the fixture
   **with the archive present** and witness `LOCKED + PASS` **live** (a green `t.Skip` is not
   acceptance), and MUST independently run the §5.1 controls — in particular the **launched-binary
   swap control**, confirming it fails **only** at the TWIN-INV-18 `/proc/<pid>/exe` check (not at the
   step-2 archive gate) and that the unmodified run reaches and passes that check. An acceptance that
   exercises only the tampered-archive control has not calibrated the witness `proof_binding` depends
   on.
4. Only when unit 3 is accepted under 3a does the `ReplayIsolated` case leave `UNVERIFIED`, and only
   then can `DEF-P7-001` move toward RESOLVED.
