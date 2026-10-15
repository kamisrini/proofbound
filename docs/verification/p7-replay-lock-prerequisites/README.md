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
source record, not a reviewed setup manifest or a passing fixture. The following remain
blocking before a case may report `PASS`:

- Fresh non-author review of the manifest's exact committed bytes, source URLs, role/member
  rules, and platform. No case manifest binds its SHA-256 and evidence revision yet.
- No reviewed projector/build manifest, complete Go module `.mod`/`.zip` byte corpus,
  toolchain-tree inventory, built executable digest, or runtime fixture corpus exists for
  this route. The current Go toolchain pin is 1.26.7; version suitability should be reviewed
  before binding a build, because later 1.26 patch releases exist.
- The Maven JAR's published checksum and nested `txz` matched the directly downloaded bytes,
  but its PGP signing key was not verified and the nested archive carried no license/notice
  files. The Maven POM declares Apache 2.0; that is not a clearance of every bundled
  component. Provenance, notices, and any necessary distribution permission remain
  unverified. No third-party binary is being committed or redistributed here.
- No companion API/live `twin` SPEC, safe-staging implementation, zero-fetch witness,
  complete-ledger/cleanup fixture, L12/L17 run, or independent implementation verdict exists.
  Missing any of these keeps the relevant case `UNVERIFIED`.

## Before running a fixture

1. Start with a plain clone at the exact reviewed commit. Require a reviewed
   `docs/verification/p7-replay-lock-fixtures/setup-sources.json` and its SHA-256 in the case
   manifest. **The current file is only a draft; it has no independent review or case-manifest
   binding, and the runtime corpus does not exist. Stop as `UNVERIFIED` until those gates
   close.** Never infer a digest from a local cache filename.
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
   `bin/pg_ctl`, bind the actual server binary digest, and reject unsafe links. Build the
   projector executable from the pinned source and Git-tracked verified module inputs;
   compare its exact SHA-256 to the reviewed manifest. Record commands, resolved URLs,
   versions, digests, and effective database settings. Do not use an ambient module or
   embedded-Postgres cache.
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
