# P7 B1 unit 3 — prepared-offline runtime fixture spec (DRAFT)

**Status:** DRAFT — spec-first per Law 6; scoping and contract only. No implementation, no runtime
`PASS`, and no `DEF-P7-001` movement is authorized by this document. Needs a non-author spec review
(and, because it touches `kernel/*`, an owner-approved change intent) before implementation.
**Date:** 2026-11-03
**Implements:** [`VD-p7-postgres-source-route-2026-11-01`](../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1), unit 3 of the [slice plan](P7-postgres-b1-source-build-slice-plan.md).
**Builds on (accepted):** units 1–2 — the SHA-256-verified 16.4 source
([verify evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md)),
the minimal-permissive build + all-permissive license inventory
([build evidence](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-build-20261103.md),
round-2 ACCEPTABLE), and the already-accepted verified-archive staging seam
(`internal/twin` TWIN-INV-12..17; companion-archive-controls verdict).
**Deferral:** [`DEF-P7-001`](../deferrals.tsv) (stays OPEN/BLOCKED until unit 3 is accepted).

## 1. Purpose — what "verified" means here

Units 1–2 established that the PostgreSQL source is authentic and the produced binary's licensing is
clean. Unit 3 is the only unit that produces a **runtime** fact: it runs a real, from-source,
minimal-permissive PostgreSQL under `ReplayIsolated`, witnesses that run, and binds its
`vera.replay.v1` proof so that:

- `proofbound.p7.behavior-lock.v1` `EvaluateLock` derives **LOCKED** over the real committed proof
  bytes (today it derives `UNVERIFIED` because no real proof exists); and
- `proofbound.p7.case-manifest.v1` `DeriveCaseStatus` folds to **PASS** (today it is `UNVERIFIED`
  because the nine outcomes carry no evidence).

Only when a non-author accepts that runtime evidence can `DEF-P7-001` move toward RESOLVED.

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

Unit 3 supplies exactly those missing runtime artifacts. It adds **no** new provider, event kind,
feed, or PostgreSQL route, and does not touch the frozen `ReplayIsolated` subject or the
`vera.replay.v1` proof shape.

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
   The exact `lib/` subset (if any backend-loadable module is needed by the replay projector) is a
   decision the implementation must settle empirically (see §8) — the minimal replay loads no
   contrib modules, so the expected `lib/` subset is empty or near-empty.
2. **No symlinks.** The staging seam fails closed on symlinks (TWIN-INV-13). The unit-2 install tree
   contains 8 symlinks in `lib/` (`libpq.so.5 -> libpq.so.5.16`, `libecpg*`, `libpgtypes*`) — all
   **client** libraries the server does not link. The prep step must drop them (if not in the
   subset) or **dereference** any retained symlink into a regular file. The packaged archive must
   contain only regular files of mode 0644 or 0755.
3. **Deterministic inventory.** Compute, over exactly the packaged files, the sorted-unique
   `Inventory []VerifiedArchiveEntry{Path, Mode, SHA256}`, the whole-archive `ArchiveSHA256`, and the
   `PostgresBinarySHA256` (`bin/postgres`). These feed `VerifiedPostgresInput`.

**Committed evidence (non-binary):** a prepared-archive manifest JSON (archive SHA-256, postgres
binary SHA-256, full per-file inventory, the prep command + toolchain identity from the unit-2 build
evidence). The `.txz` bytes themselves stay out of Git. Honest-digest caveat from unit 2 applies: PG
embeds the prefix path, so the archive/binary digests pin **this** prepared build, not a cross-host
reproducibility anchor; the reproducibility anchor remains the unit-1 **source** digest.

**Staging location (prepared-offline):** the fixture reads the archive path from an environment
variable (proposed `PROOFBOUND_P7_PG_ARCHIVE`). No path baked into code, no network, no ambient
cache. The fixture hashes the bytes and requires `ArchiveSHA256` to equal the committed manifest
before use.

## 4. Fixture runtime contract (new behavior)

A runtime fixture (a Go acceptance test in `internal/twin`, plus any minimal prep/driver helper)
that:
1. Resolves the prepared archive from the env var. **If unset or absent → `t.Skip`**: the committed
   case stays `UNVERIFIED`; the fixture never fetches, never fabricates, never passes by absence.
2. Loads the archive bytes, verifies `ArchiveSHA256` against the committed manifest, and builds a
   `VerifiedPostgresInput` from the committed inventory.
3. Assembles a **committed, synthetic** source-ledger prefix + candidate (fixtures, never presented
   as production observations — permitted by the behavior-lock SPEC) in a real source store.
4. Calls `ReplayIsolatedWithVerifiedPostgres`, which launches the real from-source server in a
   temporary embedded store on loopback, projects, compares snapshots, and returns a `Result` with a
   real `vera.replay.v1` `Proof`.
5. Produces the witnesses in §5 and the evidence artifacts in §6.
6. Authors the behavior-lock + case-manifest over the real proof, asserts `EvaluateLock == LOCKED`
   and `DeriveCaseStatus == PASS`, and commits the non-binary evidence at one evidence revision.

**Short-socket-dir constraint (from unit 2):** the temporary replay root must place its Unix socket
in a short directory — a deep temp path overruns PostgreSQL's 107-byte socket limit. The fixture
must use a short socket dir (or loopback TCP) and this is a proving-test case.

## 5. Runtime witnesses (new invariants to add to `internal/twin` SPEC)

Each maps to a `case-manifest` outcome that must become `PASS` with evidence:

| New invariant | Statement | case-manifest outcome |
|---|---|---|
| **TWIN-INV-18 — Launched-binary witness** | the launched `postgres` executable's bytes equal the declared `PostgresBinarySHA256` (closes the TWIN-INV-14 runtime deferral — re-hash the staged `bin/postgres` across the run / observe the launched process), not just the pre-launch staged bytes | `proof_binding` / provenance |
| **TWIN-INV-19 — Zero-fetch witness** | the run performs no network fetch for binaries: `BinariesPath` is the verified local dir, no remote/cache fetch strategy is taken; a test seam asserts the download path is never invoked | `no_fetch` |
| **TWIN-INV-20 — Whole-ledger preservation witness** | the source store's event set is byte-identical before and after the isolated run (extends TWIN-INV-1 with a recorded witness) | `ledger_preserved` |
| **TWIN-INV-21 — Cleanup witness** | the owned temporary root (incl. the staged binaries dir and data dir) is removed on success and every failure path, witnessed (extends TWIN-INV-17) | `cleanup` |
| **TWIN-INV-22 — Real-proof binding** | the committed `vera.replay.v1` proof bytes are produced by this real run and make `EvaluateLock == LOCKED` | `proof_binding` |

The remaining case outcomes (`source_binding`, `baseline_provenance`, `candidate_binding`,
`snapshot_comparison`, `intent_binding`) are produced by binding the committed source/candidate
fixtures, baseline provenance tuple, snapshot digests, and the P5 requirement
`BR-replay-evidence-6e3d2b` O-1 (resolved via the `RequirementResolver`) — the same bindings the
behavior-lock already validates, now carried into the case manifest with real evidence.

## 6. Evidence artifacts this unit commits (all non-binary)

- The prepared-archive manifest (archive + binary digests, per-file inventory, prep command) — §3.
- The committed `vera.replay.v1` proof bytes from the real run.
- The authored `proofbound.p7.behavior-lock.v1` (LOCKED over those proof bytes) and the
  `proofbound.p7.case-manifest.v1` (`case_status=PASS`, nine outcomes with evidence).
- Witness records (launched-binary, no-fetch, ledger-preserved, cleanup) as `{kind, path, sha256}`.
- Command logs (`initdb`/start/stop) with `network_policy=loopback-only`, exit codes, stdout/stderr
  digests.
- The runtime-fixture evidence record + the non-author verdict.

The archive `.txz`, the data directory, and any built executable are **excluded** from Git.

## 7. Non-goals

- No new provider, event kind, feed, writable truth, or PostgreSQL route; the snapshot provider
  stays skipped (Gate 1).
- No change to the frozen `ReplayIsolated` subject, the `vera.replay.v1` proof shape, or the
  authored/observed boundary.
- No redistribution of the built binary; unit 3 establishes runtime evidence, not a release.
- No reproducible-across-hosts binary digest (unit-2 caveat); the source digest is the anchor.

## 8. Known risks / open decisions (settle before/within implementation)

1. **embedded-postgres from-source compatibility (highest risk).** `fergusstrange/embedded-postgres`
   v1.31.0 with `BinariesPath` set was designed around the zonky archive layout. It must be proven
   that it will `initdb` + `Start` from our from-source `bin/`+`share/` subset. The implementation
   must demonstrate a green run before any evidence is authored; if the lib needs a layout tweak,
   that is a finding to resolve, not to paper over.
2. **Symlink dereference** (§3.2) — the prep must produce a symlink-free archive; proving test that
   staging rejects a symlinked archive already exists.
3. **Launched-process witness mechanism** (TWIN-INV-18) — options: re-hash the staged `bin/postgres`
   immediately after `Start` (proves no swap), assert the server `version()` string equals the pinned
   `PostgreSQL 16.4 …`, and/or resolve the launched PID's executable. Pick the strongest practical
   witness; the current seam only checks pre-launch bytes.
4. **Zero-fetch witness mechanism** (TWIN-INV-19) — how to assert no download path was taken (a seam
   that fails if the remote fetch strategy is constructed, plus `BinariesPath` set).
5. **Governed paths / change intent.** New code in `kernel/*` is governed (SPEC.md is exempt). This
   unit needs an owner-approved change intent (a `CI-*` record) before the governed commit, plus the
   package-acceptance bar: `make mutants` green (survivors declared) AND a non-author ACCEPTABLE
   verdict (ROADMAP P1 bar, Law 9).

## 9. Sequence & gates

1. **Spec review** (this document) — non-author review + owner sign-off on the approach and the open
   decisions in §8; raise a change intent for the governed kernel work.
2. **Prep + compatibility proof** — produce the archive, prove embedded-postgres runs from it
   (risk §8.1). Commit the prepared-archive manifest.
3. **Fixture + witnesses** — implement the runtime fixture and TWIN-INV-18..22; extend the twin
   SPEC; author the behavior-lock + case-manifest over the real proof. **Gate:** `EvaluateLock ==
   LOCKED`, `DeriveCaseStatus == PASS`, zero-fetch + ledger-preserved + cleanup witnessed,
   calibrated `make mutants` green, non-author ACCEPTABLE.
4. Only when unit 3 is accepted does the `ReplayIsolated` case leave `UNVERIFIED`, and only then can
   `DEF-P7-001` move toward RESOLVED.
