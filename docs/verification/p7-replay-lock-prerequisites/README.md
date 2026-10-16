# P7 replay-lock prerequisite setup (proposed)

This README explains what a future `ReplayIsolated` lock fixture needs **before** its offline
run. It is not an implemented setup command or evidence of a passing replay. The
[contract amendment](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
received [fresh non-author documentary acceptance](../verdicts/p7-external-replay-prerequisites-round2-20261015.md)
for its exact reviewed revision; that does not accept this later source manifest or a runtime
result. No PostgreSQL database, toolchain/Postgres binary archive, ambient
module cache, or built executable belongs in Git. Source files, configurations, READMEs,
manifests, and other nonbinary evidence may be committed; the accepted addendum still calls
for exact `.mod`/`.zip` module bytes in the reviewed Git revision.
Those module archives are source inputs, not permission to commit executable/runtime payloads.

## Current fixture gate: UNVERIFIED

The [candidate setup-source manifest](../p7-replay-lock-fixtures/setup-sources.json) records
the two directly fetched Linux/amd64 archives measured on 2026-10-15. It is a **proposed**
source record, not an accepted setup manifest or a passing fixture. Its
[round-1 non-author review](../verdicts/p7-replay-setup-sources-round1-20261016.md) returned
`AMEND`. The following remain blocking before a case may report `PASS`:

- Fresh non-author `ACCEPT` of the amended setup documentation's exact committed bytes,
  source URLs, role/member rules, and platform. No case manifest binds the setup-source
  manifest's SHA-256 and evidence revision yet.
- No reviewed projector/build manifest, complete Go module `.mod`/`.zip` byte corpus,
  toolchain-tree inventory, built executable digest, or runtime fixture corpus exists for
  this route. The current Go toolchain pin is 1.26.7; version suitability should be reviewed
  before binding a build, because later 1.26 patch releases exist.
- The Maven JAR's published checksum and nested `txz` matched the directly downloaded bytes,
  but the detached JAR signature did **not** verify. The public key returned for signature
  key ID `6B02ED0ED53F9FD3` has a self-declared `Developers Zonky` identity; that ID is an
  encryption-only subkey (`C8C07C1C159A1E0087DF0F4E6B02ED0ED53F9FD3`) that expired in
  January 2020, before the August 2024 signature. GPG reported `Can't check signature: No
  public key`; neither a valid signature nor an independently authenticated publisher
  identity has been established. The JAR and its 1,058-member nested `txz` had no
  license/copyright/notice-named files. Its Maven POM declares Apache 2.0, but the archive
  also contains PostgreSQL and other shared libraries; the POM alone does not establish
  their license or notice coverage. Obtain a trusted publisher key or alternative
  authenticated provenance and a component-license/notice inventory before clearing this
  input. Distribution permission remains unverified. No third-party binary is being
  committed or redistributed here.
- No companion API/live `twin` SPEC, safe-staging implementation, zero-fetch witness,
  complete-ledger/cleanup fixture, L12/L17 run, or independent implementation verdict exists.
  Missing any of these keeps the relevant case `UNVERIFIED`.

Upstream source check (2026-10-15): the publisher's tagged
[v16.4.0 build configuration](https://github.com/zonkyio/embedded-postgres-binaries/blob/v16.4.0/build.gradle)
selects the Debian build route for Linux/amd64. Its
[build script](https://github.com/zonkyio/embedded-postgres-binaries/blob/v16.4.0/scripts/build-postgres-debian.sh)
uses PostgreSQL source and copies system libraries including OpenSSL, ICU, libxml2, libxslt,
zlib, liblzma, and OSSP UUID into the bundle. This source route is not a reproducible-build
attestation for the downloaded JAR. The tagged repository has a project Apache license, but
no component-license/notice inventory was located there. The
[release workflow](https://github.com/zonkyio/embedded-postgres-binaries/blob/v16.4.0/.github/workflows/release.yml)
supplies signing parameters from secrets without publishing a fingerprint; the maintainer's
[public GitHub GPG key record](https://api.github.com/users/tomix26/gpg_keys) lists a different
key ID and does not authenticate the Maven signature. These findings do not close the
provenance or license-coverage blockers above.

## Before running a fixture

1. Start with a plain clone at the exact reviewed commit. Require an independently accepted
   `docs/verification/p7-replay-lock-fixtures/setup-sources.json` and its SHA-256 in the case
   manifest. **The current file remains a draft: round-1 review returned `AMEND`, no
   independent `ACCEPT` or case-manifest binding exists, and the runtime corpus does not
   exist. Stop as `UNVERIFIED` until those gates close.** Never infer a digest from a local
   cache filename.
2. On a network-enabled setup machine, obtain only the exact public Go toolchain archive and
   embedded-Postgres Maven distribution named by that manifest. Disable HTTP redirect
   following; any 3xx response fails setup. The requested and final URLs must be identical.
   Resolve the selected Go module `.mod`/`.zip` bytes from the reviewed Git revision. Require
   the Go entry to use `tar.gz`/`self` with equal download/consumed size and SHA-256, and the
   Postgres entry to use `jar+txz` with one exact, safe, regular nested `.txz` member. Require
   each entry's platform and consumed digest to match the projector manifest as specified in
   the amendment. Verify the JAR before opening it and the nested member before extracting it.
3. Keep all downloaded and extracted bytes in a dedicated directory **outside the clone**.
   Verify every raw SHA-256 and size before use; safely extract into empty directories and
   compare complete toolchain and Postgres file inventories. For Postgres, require executable
   `bin/pg_ctl`, bind the actual server binary digest, and reject unsafe links. Verify the
   Git-tracked module bytes against the projector manifest and applicable `go.sum`; reject
   unrecorded Go cache/workspace, `replace`, vendor, or other build inputs. Build the
   projector executable from the verified toolchain, pinned source, and verified modules;
   compare its exact SHA-256 to the reviewed manifest. Record the exact requested and final
   URLs (which must match), commands, time, sizes, digests, complete extracted inventories,
   build flags, versions, and effective database settings. Do not trust the setup directory
   by path alone or use an ambient module or embedded-Postgres cache.
4. Only after setup verification, run the replay fixture with outbound network denied and
   loopback available. Supply the exact verified Postgres `txz` bytes and reviewed digest to
   the companion entry point. It must rehash, safely extract inside its own fresh temporary
   root, recheck the complete staged inventory, and use that root's `BinariesDir`, never its
   cleared `RuntimeDir` or an external pre-staged path. It must remove the entire root on
   success and failure. The fixture must prove zero fetch attempts, whole-ledger preservation,
   temporary-root cleanup, and byte-exact proof/result bindings.
   Missing or changed inputs, blocked setup downloads, or absent attempt-level witnesses are
   `UNVERIFIED`, not a passing negative control.

For this checkout, `kernel/go.mod` selects `embedded-postgres v1.31.0`; its default is
PostgreSQL 16.4.0 on Linux/amd64. The earlier local cached `txz` was 14,730,084 bytes; a
subsequent direct Maven download exposed exactly one `postgres-linux-x86_64.txz` member with
the same size and SHA-256. The [accepted overlay](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
supersedes only the old binary-commit and empty-cache/no-network clean-clone obligations of
the earlier [addendum](../../plans/P7-replay-lock-spec-addendum-draft1.md) and
[matrix](../../plans/P7-replay-lock-fixture-matrix-draft1.md). Other controls remain; the old
claim cannot be relabeled `PASS`, and the new prepared-offline case has not been run.
