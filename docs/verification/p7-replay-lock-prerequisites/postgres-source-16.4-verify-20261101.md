# P7 B1 — PostgreSQL 16.4 source fetch + verify evidence (2026-11-01)

**Status:** fetch/verify evidence for B1 unit 1. This records a reproducible provenance check of the
pristine-upstream PostgreSQL source; it is NOT yet an accepted unit (a non-author verdict is still
required — Law 9) and NOT a passing runtime replay.
**Decision:** [`VD-p7-postgres-source-route-2026-11-01`](../../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1), as corrected 2026-11-01. **Manifest:** [`postgres-source-16.4.json`](../p7-replay-lock-fixtures/postgres-source-16.4.json).
**Deferral:** [`DEF-P7-001`](../../deferrals.tsv) (stays OPEN/BLOCKED).

## Provenance mechanism (as adopted)

postgresql.org distributes source tarballs with **SHA-256 and MD5 checksum files only — there is no
detached `.asc` GPG signature and no `KEYS` file in the source tree** (confirmed directly against the
v16.4 folder listing). So the originally-recorded "verify a detached signature against a published
PGDG key" step is not available for the source tarball. The owner adopted the least-expensive,
light-touch route: **pin the published SHA-256 and verify a tarball fetched over HTTPS from the
official domain.**

**Honest guarantee of this mechanism:** byte-for-byte integrity of the tarball against the
checksum *published by the project*, plus official-domain TLS trust at fetch time. It is **not** an
offline, independently-authenticated signer-identity signature: the checksum and the tarball come
from the same origin, so a compromise of that origin/TLS path is not independently detectable. This
is a deliberate, owner-accepted residual, materially stronger than the retired zonky route (whose
detached signature actively failed to verify) but weaker than a true signature chain.

## What was done (reproducible)

Downloads kept in a scratch directory **outside the clone**. No redirects followed
(`--max-redirs 0`); the requested URL equaled the effective URL for every fetch (HTTP 200).

```
BASE=https://ftp.postgresql.org/pub/source/v16.4
curl -sS --proto '=https' --max-redirs 0 $BASE/postgresql-16.4.tar.bz2.sha256   # 200, no redirect
curl -sS --proto '=https' --max-redirs 0 $BASE/postgresql-16.4.tar.bz2          # 200, no redirect
sha256sum -c postgresql-16.4.tar.bz2.sha256                                     # OK
```

## Measured vs published

| field | value |
|---|---|
| source_url | `https://ftp.postgresql.org/pub/source/v16.4/postgresql-16.4.tar.bz2` |
| effective_url | identical to source_url (no 3xx) |
| published SHA-256 (from `.sha256` over HTTPS) | `971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f` |
| measured SHA-256 (`sha256sum` of the download) | `971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f` |
| SHA-256 match | **yes** (`sha256sum -c` → OK) |
| published size (folder listing) | 24765786 bytes |
| measured size | 24765786 bytes |
| MD5 (informational) | `bdcc1e350b473c13d132d190c0c18499` (matched) |

## Reproducible verifier

The hand-run above is now reproducible via `scripts/p7-postgres-source-verify.sh` (fail-closed:
reads the pinned manifest, requires the published `.sha256` file to equal the manifest pin, the
tarball bytes to hash to that pin, and the size to match; exits non-zero on any mismatch or missing
input). Against the real download it printed:

```
p7-postgres-source-verify: OK role=postgres-source file=postgresql-16.4.tar.bz2 \
  sha256=971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f size=24765786
```

Its self-test `scripts/tests/p7-postgres-source-verify.test.sh` (auto-run by `make hooks-test`)
asserts a neutral control (pristine input accepted) and positive controls that must be rejected:
tampered byte, wrong pinned size, manifest/checksum divergence, bytes-vs-pin mismatch, and missing
input.

## Outcome and what remains

Unit 1's provenance check **passes** under the adopted mechanism: the pristine-upstream
PostgreSQL 16.4 source tarball is pinned and verifies against its published SHA-256 and size, via a
fail-closed verifier with a self-test. What still remains before any case may leave `UNVERIFIED`:

- A non-author verdict accepting this manifest + verifier + evidence (Law 9).
- Unit 2: build from this verified source (the C toolchain is absent on this box — see the build-host
  decision) and produce a complete component license/notice inventory.
- Unit 3: prepared-offline runtime fixture binding a real `vera.replay.v1` proof.
