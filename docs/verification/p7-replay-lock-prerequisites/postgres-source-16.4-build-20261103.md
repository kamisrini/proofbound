# P7 B1 — PostgreSQL 16.4 source build evidence (2026-11-03)

**Status:** build evidence for B1 unit 2 (build + component license/notice inventory). This records
a reproducible-enough build of the SHA-256-verified pristine-upstream source and proves the produced
server runs; it is NOT yet an accepted unit (a non-author verdict is still required — Law 9) and NOT
a passing runtime replay (that is unit 3).
**Decision:** [`VD-p7-postgres-source-route-2026-11-01`](../../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1). **Slice plan:** [`P7-postgres-b1-source-build-slice-plan.md`](../../plans/P7-postgres-b1-source-build-slice-plan.md)
(unit 2). **Source provenance (unit 1):** [`postgres-source-16.4-verify-20261101.md`](postgres-source-16.4-verify-20261101.md).
**Inventory (companion):** [`postgres-16.4-license-notice-inventory.md`](postgres-16.4-license-notice-inventory.md).
**Deferral:** [`DEF-P7-001`](../../deferrals.tsv) (stays OPEN/BLOCKED).

## Provenance precondition (digest-before-trust)

The build ran only after the unit-1 fail-closed verifier re-verified a freshly re-fetched tarball on
2026-11-03: `scripts/p7-postgres-source-verify.sh` printed `OK … sha256=971766d6…3386d6f size=24765786`
and its self-test passed. The build tree was extracted from **that** verified tarball. Nothing was
built from an unverified byte.

## Build host and feature-minimization decision

The C toolchain was absent on this WSL box until the owner installed it interactively on 2026-11-03
(`sudo apt-get install -y build-essential bison flex libreadline-dev zlib1g-dev`). The owner approved
the **minimal-permissive** `./configure` feature set for a replay fixture that needs only the server
driven over libpq/pgx (not interactive `psql`):

- `--without-icu` — ICU is the heaviest dependency and largest license surface; `initdb` defaults the
  collation provider to **libc**, so the replay does not need ICU. (Also avoids a `libicu-dev`
  install.)
- `--without-readline` — GNU Readline is **GPLv3** and links only into `psql`, which the fixture
  never uses; dropping it keeps the produced binary's license set **all-permissive**.
- zlib kept (default-on) — permissive (zlib license), trivial notice.

Everything else is off by `./configure`'s own defaults for 16.4: OpenSSL/TLS, libxml/libxslt, lz4,
zstd, gssapi, pam, ldap, selinux, systemd, perl/python/tcl, and NLS. (Help-text convention: a flag
shown as `--without-X` is on by default; `--with-X` is off by default — confirmed against the
verified source's `./configure --help`.)

## Toolchain (this build)

| tool | version |
|---|---|
| host | `Linux 5.15.167.4-microsoft-standard-WSL2 x86_64` (WSL2) |
| gcc | `gcc (Ubuntu 11.4.0-1ubuntu1~22.04.3) 11.4.0` |
| glibc | `GLIBC 2.35-0ubuntu3.15` |
| GNU Bison | `3.8.2` |
| flex | `2.6.4` |
| GNU Make | `4.3` |

## Build (reproducible enough)

Built **outside the clone**, in a session scratch directory; the built tree is never committed
(binary-exclusion law unchanged). `<out>` denotes that scratch prefix.

```
./configure --without-icu --without-readline --prefix=<out>
make -j<nproc>
make install
```

| step | result |
|---|---|
| configure | OK, 43 s (`using compiler=gcc … 11.4.0`; CFLAGS `-O2 … -fwrapv`; zlib detected) |
| make (`-j28`) | OK, 35 s |
| make install | OK, 4 s |

### Produced binaries the fixture needs — SHA-256 (this build)

```
11d734c405f16d21ff85fc63b0c30549bd0203016fb27af9bc3efb098f09f55e  bin/postgres
5efd446c39862f1c25182d65119608c13a3a3d8c6ede7da9d815ac2aee3376bc  bin/initdb
59ddc27882e9aaa02d7840c786bc53ccef93557fed660636f72f9893b0a1cfd7  bin/pg_ctl
f4bbd564d72a4b45530a5fbe84160f776307a947d25466885475093f0a4741bc  bin/psql
ef7a9b5b0fb77cb1b49b825194aa5467852ed156fe9012b285eabb92b581133b  bin/pg_config
```

**Honest limit on these digests:** PostgreSQL embeds the configure `--prefix` path and build paths
into its binaries, so these SHA-256s attest *this* build on *this* host/prefix produced *these
bytes* — they are **not** a cross-environment reproducibility pin (a rebuild under a different prefix
or compiler yields different bytes). The reproducible-across-environments anchor is the unit-1
**source** digest (`971766d6…3386d6f`); these binary digests pin the artifact of one build run.

## What actually linked into the server binary (grounds the inventory)

`ldd bin/postgres`:

```
linux-vdso.so.1
libz.so.1 => /lib/x86_64-linux-gnu/libz.so.1
libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6
libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
/lib64/ld-linux-x86-64.so.2
```

Grep for the minimized/opt-in libraries confirmed **none** of
`libicu / libreadline / libedit / libssl / libcrypto / libxml / liblz4 / libzstd` is linked into the
server. The only non-system shared library is **zlib** (`libz.so.1`); the rest (`libm`, `libc`,
loader) are the glibc runtime. This grounds the inventory's **dynamic-linkage** set in fact. Note
that `ldd` covers dynamic linkage only — third-party source vendored into the PostgreSQL tree and
statically compiled into the binary is enumerated separately by the companion inventory (a per-file
license-header scan); see the companion inventory's Method section (corrected after round-1 review).

## Functional proof (the minimal build runs)

`initdb` + loopback start + `select version()` + clean stop + temp-root removal, all in a scratch
root:

| check | result |
|---|---|
| initdb (`--locale-provider=libc -E UTF8`) | OK |
| server start (loopback `127.0.0.1`, private socket dir) | OK |
| `select version()` | `PostgreSQL 16.4 on x86_64-pc-linux-gnu, compiled by gcc … 11.4.0, 64-bit` |
| `datlocprovider` of `postgres` db | `c` (libc — confirms the `--without-icu` path works) |
| `pg_ctl … stop` + `rm -rf` temp root | OK (nothing left running, root removed) |

**Unit-3 implementation note (surfaced here):** the server failed its first start only because the
long scratchpad path overran PostgreSQL's **107-byte Unix-domain socket path limit** (`LOG: Unix-domain
socket path … is too long`). This is an environment constraint, not a build defect — TCP bind on
`127.0.0.1` succeeded. The unit-3 replay fixture MUST place its Unix socket in a short directory
(e.g. a short `mktemp -d` path), not under the deep scratchpad tree.

## Outcome and what remains

Unit 2's build + link-set + functional checks **pass** under the minimal-permissive configuration:
the verified 16.4 source builds cleanly, the produced server links only permissive/system libraries,
and it runs on libc collation. What still remains before any case may leave `UNVERIFIED`:

- A non-author verdict accepting this build evidence + the companion license/notice inventory (Law 9).
- Unit 3: a prepared-offline runtime fixture that binds a real `vera.replay.v1` proof over a run of
  this server tree (honoring the short-socket-dir note above), producing a *verified*
  `proofbound.p7.case-manifest.v1` / `proofbound.p7.behavior-lock.v1` status.
