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

The go/no-go picked B1 because it is the only route whose bytes this project can authenticate today.
The zonky binary failed on two counts — an unverifiable detached signature and no component
license/notice inventory. Building PostgreSQL from the official signed source tarball addresses both
at the root: the tarball is signed by a *published* PGDG release key (authenticity), and because we
control `./configure`, the set of bundled/linked components is known and small enough to inventory
completely (licensing). The slice's whole job is to turn that into committed, reviewable evidence.

## Scope (what this slice delivers)

1. **Setup-source manifest for the PostgreSQL source** — a pinned record, reviewed before any fetch
   is trusted, that names: the exact upstream tarball URL (`https://ftp.postgresql.org/pub/source/
   v<ver>/postgresql-<ver>.tar.bz2`), its detached-signature URL (`…​.tar.bz2.asc`), the expected
   tarball SHA-256 and byte size, and the full 40-hex fingerprint of the PGDG release signing key
   the signature must verify against. Target version: **PostgreSQL 16.4** (matches the retired
   zonky default so the replay's catalog/on-disk expectations are unchanged). This is a new
   role-typed section of / companion to
   [`setup-sources.json`](../verification/p7-replay-lock-fixtures/setup-sources.json), reusing its
   no-redirect, requested==final-URL, digest-and-size rules.
2. **A verifier script** (`scripts/p7-postgres-source-verify.sh`, governed) that, given the manifest
   and a locally fetched tarball + `.asc` + key, (a) imports only the pinned-fingerprint key and
   confirms the fingerprint, (b) runs `gpg --verify` and requires a good signature **by that exact
   key**, (c) recomputes the tarball SHA-256 and size and requires an exact manifest match, and
   fails CLOSED (non-zero, no partial trust) on any mismatch, expired/revoked key, or missing input.
   It has a self-test (Law 3): a positive control (tampered byte → reject) and a neutral control
   (pristine input → accept) wired into `make hooks-test`-style harness or `scripts/tests/`.
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

1. **Manifest + verifier (this is the provenance core).** Write the source manifest and
   `p7-postgres-source-verify.sh` with its self-test. Execute a network-permitted fetch of the
   tarball + `.asc` + PGDG key, run the verifier, and record the measured digests/size and the
   `gpg --verify` result as evidence. **Gate:** good signature by the pinned fingerprint AND exact
   digest/size match, or stop `UNVERIFIED`. Non-author verdict on the manifest + verifier + evidence.
2. **Build + inventory.** Configure/build outside the clone against the verified source; record
   build evidence and produced-binary digests; generate the component license/notice inventory.
   **Gate:** reproducible-enough build evidence + complete inventory; non-author verdict.
3. **Runtime fixture.** Wire the verified tree into the prepared-offline replay, produce the case
   manifest + behavior-lock over a real run, and bind the proof. **Gate:** zero-fetch witness,
   whole-ledger preservation, temp-root cleanup, byte-exact proof/result binding; non-author verdict.
   Only when unit 3 is accepted does any case leave `UNVERIFIED`, and only then can `DEF-P7-001`
   move toward RESOLVED.

## Evidence artifacts this slice will commit

- The source manifest (pinned URLs, digests, size, key fingerprint).
- `scripts/p7-postgres-source-verify.sh` + its self-test + a `scripts/tests/` case.
- The fetch/verify evidence record (requested==final URLs, measured digests, `gpg --verify` output,
  key fingerprint confirmation, timestamp).
- The build evidence record and the component license/notice inventory.
- The runtime-fixture evidence (zero-fetch witness, cleanup proof, proof/result binding) and the
  non-author verdicts for each unit.

## Open questions to settle at unit 1

- Exact PGDG signing key fingerprint to pin (confirm the current postgresql.org release key against
  its published fingerprint before pinning — the pin is only trustworthy if the fingerprint itself
  is obtained from an authenticated channel).
- Tarball compression to pin (`.tar.bz2` vs `.tar.gz`) — pick one and pin its digest.
- `./configure` feature minimization — the smallest feature set that still runs the replay, to keep
  the license/notice inventory small and the build deterministic.
