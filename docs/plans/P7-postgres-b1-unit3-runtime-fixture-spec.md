# P7 B1 unit 3 — prepared-offline runtime fixture spec (DRAFT round 2)

**Status:** DRAFT round 2 — spec-first per Law 6; scoping and contract only. No implementation, no
runtime `PASS`, and no `DEF-P7-001` movement is authorized by this document. **Round-1 non-author
review returned NEEDS_CHANGES** (`docs/verification/verdicts/p7-b1-unit3-spec-review-round1-20261104.md`,
HEAD `d468d9c`); this revision addresses findings F1–F11. It needs a **round-2 non-author re-verify**
(and, because it touches `kernel/*`, an owner-approved change intent) before implementation.
**Date:** 2026-11-04 (revises the 2026-11-03 draft)
**Implements:** [`VD-p7-postgres-source-route-2026-11-01`](../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1), unit 3 of the [slice plan](P7-postgres-b1-source-build-slice-plan.md).
**Builds on (accepted):** units 1–2 — the SHA-256-verified 16.4 source
([verify evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md)),
the minimal-permissive build + all-permissive license inventory
([build evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-build-20261103.md),
round-2 ACCEPTABLE), and the already-accepted verified-archive staging seam
(`internal/twin` TWIN-INV-12..17; companion-archive-controls verdict).
**Deferral:** [`DEF-P7-001`](../deferrals.tsv) (stays OPEN/BLOCKED until unit 3 is accepted).

## 0. What round 1 changed (summary of the revision)

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
that evidence so that:

- `proofbound.p7.case-manifest.v1` `DeriveCaseStatus` folds to **PASS** — and PASS is reachable
  **only** when the bound runtime witnesses (launched-process, no-fetch, ledger-preserved, cleanup)
  are present and verify. This is the fact `DEF-P7-001` closure rests on.
- `proofbound.p7.behavior-lock.v1` `EvaluateLock` derives **LOCKED** over the committed
  `vera.replay.v1` proof bytes. **LOCKED is a determinism claim only** (the replay reproduces the
  committed proof digests); it is **not** evidence that a real postgres ran, and `DEF-P7-001` closure
  does **not** rest on it. LOCKED and PASS are reported together but carry different claims.

Only when a non-author **runs the fixture live (archive present) and accepts that `case_status = PASS`
with verifying witnesses** can `DEF-P7-001` move toward RESOLVED.

### 1.1 The soundness chain (explicit, after F1)

> `case_status = PASS` ⟹ every runtime outcome is PASS ⟹ the launched-process witness
> (`/proc/<pid>/exe` digest == `PostgresBinarySHA256`) verified, the no-fetch isolation held, the
> source ledger was byte-preserved, and the owned temp root was removed — i.e. **a real from-source
> `postgres` executable ran this replay**.

This chain must be demonstrated by a **negative control** (§5.1, F11): with the declared binary
absent the case is `UNVERIFIED`; with a *swapped/wrong* binary the launched-process witness FAILS and
the case folds to `FAIL` — never `PASS`. If a proving test cannot make `PASS` collapse when the real
run is removed, the acceptance proves nothing (Law 9 corollary).

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
6. Authors the behavior-lock + case-manifest, asserts `EvaluateLock == LOCKED` (determinism) **and**
   `DeriveCaseStatus == PASS` with all runtime witnesses verifying, and commits the non-binary
   evidence at one evidence revision.

**Isolation & determinism (F10, short-socket from unit 2):** the fixture must not worsen
`DEF-KERNELCHECK-FLAKE-001` (embedded-postgres **port contention** on fixed ports 55461-55463). It
MUST (a) allocate an **ephemeral free port** per run (not a fixed port), (b) place the Unix socket in
a **short** directory (deep temp paths overrun PostgreSQL's 107-byte socket limit) or use loopback
TCP, and (c) run **serially**, isolated from the parallel embedded-postgres suite (e.g. its own
`go test` tag/target), so one real server never contends with another. Port/socket isolation is a
proving-test case.

## 5. Runtime witnesses (new invariants to add to `internal/twin` SPEC)

Each maps to a `case-manifest` outcome that must become `PASS` **with a bound, re-verifiable witness
record** — not a self-declared state. `DeriveCaseStatus` verifies each witness digest before folding
the outcome to PASS.

| New invariant | Statement | Required mechanism & control | case-manifest outcome |
|---|---|---|---|
| **TWIN-INV-18 — Launched-process witness** | the **running** postmaster's executable image equals the declared `PostgresBinarySHA256` | resolve the postmaster PID (`postmaster.pid` in the data dir), read `/proc/<pid>/exe`, **reject** if it carries the `(deleted)` marker, hash the target, require equality. The on-disk re-hash and `version()` string are **not** sufficient and are not used as the witness. **Negative control (required):** launching a *different/byte-swapped* binary MUST make this witness FAIL (and the case fold to FAIL). | `proof_binding` / provenance |
| **TWIN-INV-19 — No-fetch isolation** | the run performs no network fetch for binaries | **Primary witness:** the launch runs under a **network namespace with no egress** (`unshare -rn` / equivalent) so any fetch fails hard; `BinariesPath` is the verified local dir. **Declaration (labeled as such, not a witness):** the embedded-postgres remote-fetch strategy is never constructed. Where the platform cannot provide a namespace, the invariant is satisfied by the declaration **only if** the spec/evidence states plainly it is a configuration assertion, not an observation. | `no_fetch` |
| **TWIN-INV-20 — Whole-ledger preservation witness** | the source store's event set is byte-identical before and after the isolated run | before/after digest of the source store, recorded as a witness. **Positive control (required):** an injected operation that *does* mutate the source store MUST make the witness FAIL — proving it detects mutation, not merely restates TWIN-INV-1/-8. | `ledger_preserved` |
| **TWIN-INV-21 — Cleanup witness** | the owned temporary root (staged binaries dir + data dir) is removed on success and every failure path | the witness **stats the path and asserts absence** (not merely "RemoveAll was called"). **Positive control (required):** a path whose removal is suppressed MUST make the witness FAIL. | `cleanup` |
| **TWIN-INV-22 — Real-proof binding (determinism)** | the committed `vera.replay.v1` proof bytes reproduce under re-evaluation, making `EvaluateLock == LOCKED` | re-evaluate the lock over the committed proof bytes → LOCKED. **This is a determinism claim, NOT a runtime witness** (F1): LOCKED is a pure function of digests and does not establish that a real postgres ran — that fact is carried by TWIN-INV-18/19 and the case `PASS`. | (determinism check behind `proof_binding`, not a standalone runtime claim) |

The remaining case outcomes (`source_binding`, `baseline_provenance`, `candidate_binding`,
`snapshot_comparison`, `intent_binding`) are produced by binding the committed source/candidate
fixtures, baseline provenance tuple, snapshot digests, and the P5 requirement
`BR-replay-evidence-6e3d2b` O-1 (resolved via the `RequirementResolver`) — the same bindings the
behavior-lock already validates, now carried into the case manifest with real evidence.

### 5.1 The overall negative control (F11)

A dedicated proving test demonstrates that `DeriveCaseStatus → PASS` **depends on the real run**:
- **binary absent** → fixture `t.Skip` → committed case stays `UNVERIFIED` (CASE-INV-5, no PASS by
  absence);
- **binary swapped / bytes altered** → TWIN-INV-18 launched-process witness FAILS →
  `proof_binding` outcome FAIL → worst-wins fold → `case_status = FAIL`, never PASS;
- **real declared binary** → all witnesses verify → `case_status = PASS`.

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
  each digest bound into the case manifest and re-verified by `DeriveCaseStatus`.
- Command logs (`initdb`/start/stop) with the no-fetch isolation method recorded (namespace vs.
  declaration — F4), exit codes, stdout/stderr digests.
- The runtime-fixture evidence record + the non-author verdict.

The archive `.txz`, the data directory, and any built executable are **excluded** from Git.

### 6.1 Law-1 status of the committed evidence (F7)

These artifacts are produced by a run that requires the **offline archive**, which is deliberately
absent from `make check` (the fixture `t.Skip`s without it). They are therefore **one-time captured
evidence, not regenerable derived state**, and are treated exactly as the existing schema-bearing
evidence artifacts are — verdicts, build-evidence, verify-evidence — which are **documentary**,
carry **no `@generated` marker**, and are validated **structurally** (self-digest + path resolution
via `scripts/p6-artifact-integrity.sh`), not by regeneration. This is an **explicit, reasoned
exemption** from Law 1's regenerate-in-`make check` requirement, on the same basis those artifacts
already enjoy: their generating input is intentionally out of CI. They are not hand-editable
narrative either — they are machine-produced records validated by self-digest. The spec records this
exemption so the choice is deliberate, not silent.

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
   negative control (swapped binary ⟹ FAIL). The on-disk re-hash and `version()` string are rejected
   as non-witnessing. **Owner note:** this is the mechanism the review recommends; confirm the
   platform exposes `/proc/<pid>/exe` (Linux/WSL: yes).
4. **No-fetch witness mechanism — PROPOSED: network-namespace isolation** (TWIN-INV-19, F4), with the
   fetch-strategy-not-constructed assertion kept only as a labeled declaration. **Owner decision
   needed:** whether to require namespace isolation (strongest; needs user-namespace / `unshare -rn`
   support in the run environment) or accept the labeled configuration declaration where a namespace
   is unavailable. This is the one genuinely-open §8 mechanism decision for the owner.
5. **Governed paths / change intent.** New code in `kernel/*` is governed (SPEC.md is exempt). This
   unit needs an owner-approved change intent (a `CI-*` record) before the governed commit, plus the
   package-acceptance bar: `make mutants` green (survivors declared) AND a non-author ACCEPTABLE
   verdict (ROADMAP P1 bar, Law 9). The drafted `CI-p7-runtime-fixture-1a2b3c` **stays `proposed`**
   until this spec is re-accepted at round 2.

## 9. Sequence & gates

1. **Spec re-review (round 2)** — non-author re-verify of THIS revision + owner sign-off on the §8.4
   no-fetch mechanism decision; keep the change intent `proposed` until accepted.
2. **Subset prep + compatibility proof** — produce the symlink-free `bin/`+`share/` subset archive and
   **prove embedded-postgres runs from the SUBSET** (closes the F5 gap, not just the full tree).
   Commit the prepared-archive manifest.
3. **Fixture + witnesses** — implement the runtime fixture and TWIN-INV-18..22 (launched-**process**
   witness via `/proc/<pid>/exe`; no-fetch isolation; ledger-preserved + cleanup with their positive
   controls); extend the twin SPEC honestly (§5.2); author the behavior-lock + case-manifest over the
   real proof with **bound** witnesses. **Gate:** `EvaluateLock == LOCKED` (determinism) AND
   `DeriveCaseStatus == PASS` with every witness verifying; the §5.1 negative control proves PASS
   collapses to UNVERIFIED (absent) / FAIL (swapped); ephemeral-port + short-socket isolation holds;
   calibrated `make mutants` green.
3a. **Acceptance precondition (F10):** the non-author accepting this unit MUST run the fixture **with
   the archive present** and witness `LOCKED + PASS` **live** (a green `t.Skip` is not acceptance),
   and MUST independently run the §5.1 negative control.
4. Only when unit 3 is accepted under 3a does the `ReplayIsolated` case leave `UNVERIFIED`, and only
   then can `DEF-P7-001` move toward RESOLVED.
