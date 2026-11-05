# P7 B1 unit-3 §9.2 — prepared-offline SUBSET compatibility proof (2026-11-05)

**Status:** evidence for B1 **unit 3** step §9.2 (subset prep + compatibility proof — closes the
round-5/round-6 **F5** gap: prove embedded-postgres runs from the symlink-free `bin/`+`share/` subset,
not just the full install tree). This is NOT yet an accepted unit — the fixture (§9.3) and an
independent non-author verdict + the §9.3a live run are still required (Law 9). `DEF-P7-001` stays
**OPEN/BLOCKED**.
**Change intent:** `CI-p7-runtime-fixture-1a2b3c` (owner-accepted 2026-11-05).
**Spec:** [`P7-postgres-b1-unit3-runtime-fixture-spec.md`](../../plans/P7-postgres-b1-unit3-runtime-fixture-spec.md) §3, §8.1, §9.2.
**Prepared-archive manifest (committed):** [`../p7-replay-lock-fixtures/postgres-16.4-subset-archive.json`](../p7-replay-lock-fixtures/postgres-16.4-subset-archive.json).
**Builds on (accepted):** unit-1 source provenance + unit-2 minimal-permissive build.

## Provenance precondition (digest-before-trust)

The build ran only after the unit-1 fail-closed verifier re-verified a freshly re-fetched tarball on
2026-11-05: `scripts/p7-postgres-source-verify.sh` printed
`OK role=postgres-source file=postgresql-16.4.tar.bz2 sha256=971766d6…3386d6f size=24765786`.
The tarball + its published `.sha256` were fetched over HTTPS with **no redirect** (`curl --proto
=https --max-redirs 0`) from `ftp.postgresql.org`. Nothing was built from an unverified byte.

## Build (outside the clone; never committed)

```
./configure --without-icu --without-readline --prefix=<out>
make -j$(nproc)
make install
```

Produced `postgres (PostgreSQL) 16.4`. `ldd bin/postgres` links only permissive libraries —
`ld-linux`, `libc`, `libm`, `libz`, `linux-vdso` (matches the accepted unit-2 build evidence;
toolchain identity is **cited**, not copied, in the manifest — F8/Law 2).

## Subset staging (§3 packaging rules)

Staged a symlink-free subset and **asserted** (not dereferenced — F6) that no symlink and no
non-`0644`/`0755` file remained, then packaged a deterministic `.txz`:

- **Included:** `bin/{postgres,initdb,pg_ctl}`, all of `share/`, and `lib/postgresql/*.so`.
- **Excluded:** `include/` (headers), static `*.a`, the top-level client shared libraries
  (`libpq`/`libecpg`/`libpgtypes`, all symlinked), `lib/postgresql/pgxs` (build-time only), and
  every symlink (the unit-2 tree's 8 symlinks are all client libs in `lib/`; none is in the subset).

### `lib/` finding — the spec's "expected empty" is empirically overturned (§8.1 confirmation)

Spec §3 point 1 expected `lib/` to be **empty** (the minimal replay loads no backend module). The
first subset (`bin/`+`share/` only) **failed `initdb`**: its post-bootstrap text-search setup loads
`$libdir/dict_snowball` at provisioning time —

```
performing post-bootstrap initialization ...
FATAL:  could not access file "$libdir/dict_snowball": No such file or directory
STATEMENT:  CREATE FUNCTION dsnowball_init(INTERNAL) ... AS '$libdir/dict_snowball', 'dsnowball_init'
child process exited with exit code 1
initdb: removing data directory ...
```

`$libdir` = `lib/postgresql` (`pg_config --pkglibdir`). The subset therefore includes the server
**backend-loadable** modules `lib/postgresql/*.so` (dict_snowball, the encoding-conversion modules,
plpgsql, pgoutput, libpqwalreceiver) so `initdb` and any on-demand encoding/plpgsql load during
replay succeed. These are server-loaded modules (not client-only libs); the build-time `pgxs/` dir is
excluded. This is exactly the empirical confirmation §3 point 1 / §8.1 called for — it is recorded,
not hidden.

## Prepared-archive identity (committed, non-binary)

| field | value |
|---|---|
| archive format | `txz` |
| `archive_sha256` | `fa1ade9b4f9b7266add9495e0dc00be419f57857a9b3fc12632f24caaf6444e3` |
| archive size (bytes) | `4145552` |
| `postgres_binary_sha256` (`bin/postgres`) | `ba2447e5a38e52d801d0faecc8a3ee46d4cb57792cabc6e80801af4d4888fcd5` |
| inventory entries | `681` (sorted-unique; `bin/pg_ctl` + `bin/postgres` present, mode `100755`) |

**Honest-digest caveat (unit-2 carried forward):** PostgreSQL embeds the `--prefix` path into built
artifacts, so these digests pin **this** prepared build at its prefix, not a cross-host
reproducibility anchor. The reproducibility anchor remains the unit-1 **source** digest. The §9.3a
acceptance run MUST use the exact `.txz` these digests pin.

## Compatibility proof — F5 CLOSED on the SUBSET

A **throwaway** probe (a `twin`-package test, never committed — per the §8.1 precedent) extracted the
actual committed-manifest `.txz` into a fresh temp dir and launched `embedded-postgres` v1.31.0 with
`BinariesPath` pointed at **that extracted subset**:

- **no download:** `bin/pg_ctl` present in the subset short-circuits the embedded-postgres fetch
  (`embedded_postgres.go` stats `BinariesPath/bin/pg_ctl`); the probe asserts its presence first.
- `initdb` succeeded (with `lib/postgresql` present), the server **started on an ephemeral loopback
  port**, `select version()` returned
  `PostgreSQL 16.4 on x86_64-pc-linux-gnu, compiled by gcc … 11.4.0, 64-bit`, and the server **stopped
  cleanly** (`database system is shut down`).
- socket: embedded-postgres' compiled default (`/tmp`, short — under the 107-byte limit).

**Result: PASS.** The symlink-free `bin/`+`share/`+`lib/postgresql` subset archive runs
embedded-postgres end-to-end. F5 is closed; the §9.3 fixture can build `VerifiedPostgresInput` from
this manifest.

## What this does NOT establish

Not a passing runtime replay, not the launched-process witness (TWIN-INV-18), not a `case_status=PASS`
— those are §9.3. Not acceptance (Law 9 non-author verdict + §9.3a live run pending). `DEF-P7-001`
unmoved.
