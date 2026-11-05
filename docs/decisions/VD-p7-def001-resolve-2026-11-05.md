# VD-p7-def001-resolve-2026-11-05: resolve DEF-P7-001 — the B1 shippable-PostgreSQL provenance + licensing + runtime-evidence bar is met

**Status:** Accepted
**Date:** 2026-11-05
**Context:** [`DEF-P7-001`](../deferrals.tsv) has been OPEN/BLOCKED since 2026-10-17 — pre-GA,
customer-shippable PostgreSQL provenance and component licensing. The pinned zonky.io
embedded-postgres Maven artifact failed provenance and shipped no license/notice coverage, so it was
retired to dev/demo-only and the go-route was set to **B1 — build PostgreSQL from official
postgresql.org source** ([[VD-p7-postgres-source-route-2026-11-01]]). The B1 route was sliced into
three units, each with a defined acceptance bar: unit 1 source provenance, unit 2 build + component
license/notice inventory, unit 3 a prepared-offline runtime fixture that witnesses a real
from-source server actually running the replay. The owner reserved **resolving DEF-P7-001** as a
separate go/no-go (the unit-3 change intent `CI-p7-runtime-fixture-1a2b3c` explicitly excludes
DEF-P7-001 activation/resolution from unit-3 scope). As of 2026-11-05 all three units are complete
and independently accepted.

**Decision:** The **owner resolves `DEF-P7-001` (OPEN/BLOCKED → RESOLVED)** on 2026-11-05. The B1
provenance + licensing + runtime-evidence bar is met:
- **Unit 1 — source provenance (ACCEPTED):** PostgreSQL 16.4 `.tar.bz2` pinned by published SHA-256,
  fetched over official-domain HTTPS with no redirect; fail-closed verifier
  `scripts/p7-postgres-source-verify.sh` + non-author ACCEPTABLE verdict.
- **Unit 2 — build + licensing (ACCEPTED):** minimal-permissive `--without-icu --without-readline`
  build; the produced server links only permissive libraries (libc/libm/libz); a two-dimension
  (dynamic + statically-vendored) component license/notice inventory survived an adversarial
  non-author review (round-1 undercount fixed → round-2 ACCEPTABLE); all-permissive, no copyleft.
- **Unit 3 — runtime evidence (ACCEPTED):** the symlink-free `bin/`+`share/`+`lib/postgresql` subset
  runs embedded-postgres (F5 closed); a launched-process `/proc/<pid>/exe` witness (TWIN-INV-18)
  binds the running image to the declared `bin/postgres` digest; LOCKED behavior-lock + PASS
  case-manifest over a real `vera.replay.v1` proof; **non-author ACCEPTABLE package verdict**
  ([`verdicts/p7-b1-unit3-impl-review-round1-20261105.md`](../verification/verdicts/p7-b1-unit3-impl-review-round1-20261105.md))
  **and a non-author §9.3a LIVE run**
  ([`p7-b1-unit3-9.3a-live-run-20261105.md`](../verification/p7-b1-unit3-9.3a-live-run-20261105.md))
  that observed `resolved_sha256 == ba2447e5…fcd5` and reproduced the committed LOCKED/PASS records.

**What this does NOT change (residuals carried forward, stated so RESOLVED does not bury them):**
- The **binary-exclusion law is unchanged**; no PostgreSQL binary, archive, or data directory is
  committed. RESOLVED closes the provenance/**licensing-evidence** deferral — it is **not itself a
  release authorization**. Actually redistributing a built binary is a further operational step that
  this evidence supports but does not perform.
- **Source provenance is SHA-256 + official-domain TLS, not an offline signer signature**
  (postgresql.org ships no detached `.asc` for source) — the owner-accepted residual recorded in
  [[VD-p7-postgres-source-route-2026-11-01]].
- **Digests pin this prefix, not a cross-host reproducibility anchor** (PostgreSQL embeds the
  `--prefix` path); the reproducibility anchor is the unit-1 **source** digest.
- **Dynamic-dependency provenance is out of scope** (F3): the launched-process witness covers the
  `postgres` executable image, not ambient host `.so` files.
- **TWIN-INV-19 no-fetch is a labeled declaration** in this build, not a network-namespace
  observation, per [[VD-p7-nofetch-witness-2026-11-04]].
- The **full-package mutation sweep is deferred** (embedded-postgres per mutant is prohibitively
  slow); the focused `launch_witness.go` sweep (4/4 killed, calibration controls behaved) and the
  live witness-function controls stand in its place for the witness's discriminating power.

**Consequences:** The shippable-PostgreSQL provenance blocker is cleared for the B1 route; per unit-3
spec §9 item 4 the `ReplayIsolated` case may leave `UNVERIFIED`. The dev/demo zonky path and the
`Binary-exclusion` law are untouched. Reopen `DEF-P7-001` only if the binary route, the build
configuration, or the acceptance evidence changes (a new Postgres version, a different source, or a
decision to vendor the binary) — any of which needs fresh provenance + licensing evidence and a new
non-author verdict.
