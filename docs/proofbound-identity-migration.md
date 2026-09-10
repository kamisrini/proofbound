# Proofbound identity migration

**Effective:** 2026-09-10  
**Compatibility removal date:** 2026-12-31 (or P6 start, whichever comes first)

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
the directory move does not destroy access to its ledger. That private storage identity is a
deprecated compatibility alias, not a live product surface; newly initialized stores use Proofbound
identity, and the alias is removed with the rest of the P5 compatibility window.

## Compatibility aliases

The live executable is `proofbound`. The `vera` executable remains a deprecated delegating alias
through the removal date above. It emits a dated advisory and executes the same command path.

Live environment names use the `PROOFBOUND_*` prefix. During the same compatibility window, these
old names remain fallback aliases and emit the advisory when consumed:

| Live name | Deprecated alias |
|---|---|
| `PROOFBOUND_CHECK_TARGET` | `VERA_CHECK_TARGET` |
| `PROOFBOUND_GITHUB_OWNER` | `VERA_GITHUB_OWNER` |
| `PROOFBOUND_GITHUB_REPOS` | `VERA_GITHUB_REPOS` |
| `PROOFBOUND_GITHUB_API_BASE_URL` | `VERA_GITHUB_API_BASE_URL` |
| `PROOFBOUND_VERIFY_TRACE` | `VERA_VERIFY_TRACE` |
| `PROOFBOUND_CLEANROOM_PATTERNS` | `VERA_CLEANROOM_PATTERNS` |

When both forms are present, the live name wins. Standard external names such as `DATABASE_URL` and
`GITHUB_TOKEN` are unchanged.

## Frozen bytes

The compatibility window does not permit rewriting historical events or artifacts. Pinned tests
continue to parse and reproduce the three v1 wire identities byte-for-byte. New schema families use
the `proofbound.*` namespace.
