# P7 replay-lock prerequisite setup (proposed)

This README explains what a future `ReplayIsolated` lock fixture needs **before** its offline
run. It is not an implemented setup command or evidence of a passing replay. The proposed
[contract amendment](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
needs fresh non-author acceptance before these instructions replace the previously accepted
Git-tracked-byte contract. No PostgreSQL database, toolchain/Postgres binary archive, module
cache, or built executable belongs in Git.

## Before running a fixture

1. Start with a plain clone at the exact reviewed commit. Require a reviewed
   `docs/verification/p7-replay-lock-fixtures/setup-sources.json` and its SHA-256 in the case
   manifest. **That manifest and the runtime fixture corpus do not exist yet; stop as
   `UNVERIFIED` until they do.** Never infer a digest from a local cache filename.
2. On a network-enabled setup machine, obtain only the exact public Go toolchain archive,
   embedded-Postgres Maven distribution, and Go module `.mod`/`.zip` inputs named by that
   manifest. For a Maven container, verify its downloaded bytes, extract only its declared
   `txz`, and verify the nested archive separately. A private mirror or changed redirect is
   not an equivalent input without a new reviewed record.
3. Keep all downloaded and extracted bytes in a dedicated directory **outside the clone**.
   Verify every raw SHA-256 and size before use; safely extract into empty directories and
   compare complete toolchain and Postgres file inventories. For Postgres, require executable
   `bin/pg_ctl`, bind the actual server binary digest, and reject unsafe links. Build the
   projector executable from the pinned source and verified module inputs; compare its exact
   SHA-256 to the reviewed manifest. Record commands, resolved URLs, versions, digests, and
   effective database settings. Do not use an ambient module or embedded-Postgres cache.
4. Only after setup verification, run the replay fixture with outbound network denied and
   loopback available. Supply verified Postgres binaries via the temporary store's
   `BinariesDir`, never its cleared `RuntimeDir`. The fixture must prove zero fetch attempts,
   whole-ledger preservation, temporary-root cleanup, and byte-exact proof/result bindings.
   Missing or changed inputs, blocked setup downloads, or absent attempt-level witnesses are
   `UNVERIFIED`, not a passing negative control.

For this checkout, `kernel/go.mod` selects `embedded-postgres v1.31.0`; its default is
PostgreSQL 16.4.0 on Linux/amd64. A local cached `txz` measured 14,730,084 bytes, but its
origin was not independently matched to an upstream distribution because Maven DNS failed.
That local measurement is **not** a reviewed setup-source manifest or a license clearance.
The currently accepted [addendum](../../plans/P7-replay-lock-spec-addendum-draft1.md) and
[fixture matrix](../../plans/P7-replay-lock-fixture-matrix-draft1.md) still require committed
build/runtime bytes until the proposed amendment receives a fresh non-author verdict.
