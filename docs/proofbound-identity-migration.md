# Proofbound identity migration

**Effective:** 2026-09-10  
**External compatibility removed:** P6 start, 2026-09-12

Proofbound is the live product name. Historical artifacts and the frozen wire schemas
`vera.witness.v1`, `vera.verdict.v1`, and `vera.replay.v1` retain their original bytes and meaning.

## Local state

The live state directory is `.proofbound/`. Move an existing directory once, while no Proofbound
process is running:

```bash
test ! -e .proofbound
mv .vera .proofbound
```

If both paths exist, stop and resolve which ledger is authoritative; never merge or delete either
directory automatically. The directory is local derived/runtime state and remains gitignored. The
store lock remains derived from the data directory, so moving the root does not weaken lock
ownership.

An already-initialized embedded PostgreSQL cluster retains its legacy internal database/user name so
the directory move does not destroy access to its ledger. Per
`VD-p6-private-storage-compat-2026-09-12`, that private identity is frozen migration compatibility,
not a live product surface. New stores use only the Proofbound private identity.

## Removed external aliases

The live executable is `proofbound`; the former delegating executable was removed at P6 start.
Product environment names use only `PROOFBOUND_*`; former product-prefixed fallbacks are ignored.
Standard external names such as `DATABASE_URL` and `GITHUB_TOKEN` are unchanged.

## Frozen bytes

The compatibility window does not permit rewriting historical events or artifacts. Pinned tests
continue to parse and reproduce the three v1 wire identities byte-for-byte. New schema families use
the `proofbound.*` namespace.
