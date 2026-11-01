# P7 B1 — PostgreSQL-from-signed-source build slice plan

**Status:** DRAFT — scoping only; defines obligations, not an accepted result. No fetch, build,
runtime fixture, or `PASS` is authorized by this document alone.
**Date:** 2026-11-01
**Decision this implements:** [`VD-p7-postgres-source-route-2026-11-01`](../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(Option B1). **Deferral held open until the evidence below is produced and accepted:**
[`DEF-P7-001`](../deferrals.tsv).
**Supersedes nothing.** It extends the existing prerequisite route
([prerequisites README](../verification/p7-replay-lock-prerequisites/README.md)) by replacing the
*provenance-failing zonky Maven binary* input with a *provenance-verifiable source build*. Every
non-binary control from the accepted external-prerequisites overlay
([`VD-p7-external-replay-prerequisites-2026-10-15`](../decisions/VD-p7-external-replay-prerequisites-2026-10-15.md))
still holds: no runtime binary, database, or built executable in Git; prepared-offline replay;
digest-before-trust; `UNVERIFIED` on any missing/changed input.

## Why this slice exists

The go/no-go picked B1 because it is the route whose source bytes this project can authenticate.
The zonky binary failed on two counts — an unverifiable detached signature and no component
license/notice inventory. Building PostgreSQL from upstream source addresses the licensing count at
the root (we control `./configure`, so the bundled/linked component set is known and small enough to
inventory completely). **Authenticity note (corrected 2026-11-01):** postgresql.org ships source
tarballs with SHA-256/MD5 checksums only — there is NO detached `.asc` signature and no `KEYS` file
in the source tree, so "verify a signature against a published PGDG key" is not available. The owner
adopted the least-expensive, light-touch mechanism: **pin the published SHA-256 and verify a
pristine-upstream tarball fetched over HTTPS** from the official domain (integrity + official-domain
TLS trust, not an offline signature). See
[`VD-p7-postgres-source-route-2026-11-01`](../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(2026-11-01 amendment). The slice's job is to turn that into committed, reviewable evidence.

## Scope (what this slice delivers)

1. **Setup-source manifest for the PostgreSQL source** — a pinned record that names: the exact
   upstream tarball URL (`https://ftp.postgresql.org/pub/source/v<ver>/postgresql-<ver>.tar.bz2`),
   its published-checksum URL (`…​.tar.bz2.sha256`), the expected (published) tarball SHA-256 and
   byte size, and the no-redirect / requested==effective-URL rule. **No GPG fingerprint field** —
   source tarballs carry no detached signature (see the why-this-slice note). Target version:
   **PostgreSQL 16.4** (matches the retired zonky default so the replay's catalog/on-disk
   expectations are unchanged). Delivered as
   [`postgres-source-16.4.json`](../verification/p7-replay-lock-fixtures/postgres-source-16.4.json),
   a sibling to [`setup-sources.json`](../verification/p7-replay-lock-fixtures/setup-sources.json).
   **DONE 2026-11-01** — verified (digest + size match, no redirect); evidence at
   [`postgres-source-16.4-verify-20261101`](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md).
2. **A verifier script** (planned path scripts/p7-postgres-source-verify.sh, governed) that, given
   the manifest and a locally fetched tarball + its `.sha256`, (a) recomputes the tarball SHA-256 and
   size and requires an exact match to BOTH the published checksum file and the manifest pin, and
   (b) fails CLOSED (non-zero, no partial trust) on any mismatch or missing input. It carries a
   self-test: a positive control (tampered byte → reject) and a neutral control (pristine input →
   accept) under `scripts/tests/`. (The light-touch mechanism needs no `gpg` step; the fetch+verify
   was run by hand for unit 1's evidence, and this script makes it reproducible for review.)
3. **A build contract** — the exact `./configure` flags, the component set they pull in, the
   produced `bin/postgres`/`bin/initdb`/`bin/pg_ctl` the fixture needs, and the recorded build
   evidence (configure flags, compiler + libc versions, produced binary SHA-256s, timings). The
   built tree is kept **outside the clone** and never committed (binary-exclusion law unchanged).
4. **A complete component license/notice inventory** for the produced binary: the PostgreSQL license
   itself plus every library the chosen `./configure` actually links (minimized deliberately — e.g.
   decide per-component whether to enable ICU/OpenSSL/libxml for the replay fixture, and inventory
   exactly what is enabled). Committed as reviewable non-binary evidence.
5. **Runtime-fixture integration** — the replay entry point consumes a *verified* server tree:
   `initdb` into a fresh temp root on loopback only, run `ReplayIsolated`, bind the
   `vera.replay.v1` proof, and remove the entire root on success and failure. This is the piece that
   lets `proofbound.p7.behavior-lock.v1` / `proofbound.p7.case-manifest.v1` derive a **verified**
   (not `UNVERIFIED`) status for the real runtime case. Reuses the companion API's safe-staging
   discipline; replaces only the binary source.

## Non-goals (explicitly out of this slice)

- No change to the frozen `ReplayIsolated` subject, the `vera.replay.v1` proof shape, the
  authored/observed boundary, or the no-new-provider/feed/event exclusions.
- The deliberately-skipped snapshot provider stays skipped (Gate 1,
  [`VD-p7-snapshot-provider-skip-2026-10-01`](../decisions/VD-p7-snapshot-provider-skip-2026-10-01.md)).
- No redistribution of the built binary in this slice — that is a later release step; this slice
  establishes the *provenance + license evidence* that would make redistribution lawful, nothing
  more.
- No new production kernel package (the manifest/contract code already lives in the accepted
  `internal/twin/manifest`; the frozen 17-package universe is not disturbed).

## Sequence & gates (each unit spec-first, each result non-author-verified — Law 6 + Law 9)

1. **Manifest + verify (this is the provenance core).** Write the source manifest; execute a
   network-permitted fetch of the tarball + its published `.sha256` (no redirects, requested URL ==
   effective URL); require the measured SHA-256 and size to match BOTH the published checksum and the
   manifest pin; record the evidence. Add the reproducible verifier script + self-test.
   **Gate:** exact digest + size match (no detached signature exists to check), or stop `UNVERIFIED`.
   Non-author verdict on the manifest + verifier + evidence. **Fetch/verify DONE 2026-11-01** (manifest
   + evidence committed); the verifier script + self-test and the non-author verdict remain.
2. **Build + inventory.** Configure/build outside the clone against the verified source; record
   build evidence and produced-binary digests; generate the component license/notice inventory.
   **Gate:** reproducible-enough build evidence + complete inventory; non-author verdict.
3. **Runtime fixture.** Wire the verified tree into the prepared-offline replay, produce the case
   manifest + behavior-lock over a real run, and bind the proof. **Gate:** zero-fetch witness,
   whole-ledger preservation, temp-root cleanup, byte-exact proof/result binding; non-author verdict.
   Only when unit 3 is accepted does any case leave `UNVERIFIED`, and only then can `DEF-P7-001`
   move toward RESOLVED.

## Evidence artifacts this slice will commit

- The source manifest (pinned URLs, published + measured digests, size). **Done:**
  `docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json`.
- The verifier script (planned scripts/p7-postgres-source-verify.sh) + its self-test + a
  `scripts/tests/` case. **Pending.**
- The fetch/verify evidence record (requested==effective URLs, published vs measured digest/size,
  timestamp). **Done:** `docs/verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md`.
- The build evidence record and the component license/notice inventory. **Pending (unit 2).**
- The runtime-fixture evidence (zero-fetch witness, cleanup proof, proof/result binding) and the
  non-author verdicts for each unit. **Pending (unit 3 + verdicts).**

## Open questions to settle (remaining)

- **Build host** — the C toolchain is absent on this box (no gcc/bison/flex/PG headers). Resolve:
  install deps here (sudo apt) or build on a separate setup machine.
- `./configure` feature minimization — the smallest feature set that still runs the replay, to keep
  the license/notice inventory small and the build deterministic.
- *(Resolved 2026-11-01: provenance mechanism = published-SHA-256-over-HTTPS, light touch; tarball
  compression pinned to `.tar.bz2`; no GPG key to pin.)*
