# P7 replay lock — first implementation slice proposal

**Status:** AMENDED PROPOSAL 2026-10-15; the owner authorized documentary amendments to the
round-1 [AMEND verdict](../verification/verdicts/p7-replay-first-implementation-scope-round1-20261015.md).
Fresh non-author review of this amended revision is pending. This document authorizes no
product code, runtime corpus, or acceptance claim.
The [round-4 verdict](../verification/verdicts/p7-replay-lock-slice-round4-20261015.md)
accepted the [addendum](P7-replay-lock-spec-addendum-draft1.md) and
[fixture matrix](P7-replay-lock-fixture-matrix-draft1.md) for documentary planning only.

## Recommended boundary

Make the first implementation slice the **offline embedded-Postgres provisioning seam** for
`twin.ReplayIsolated`, with a committed archive and a small adversarial fixture. This is the
Postgres portion of L12 and the L17 control, not a behavior-lock evaluator or an L12 pass.
Preserve the existing `ReplayIsolated` call and frozen `Result`/`Proof` behavior. Add a narrow,
explicit companion entry point for a caller that supplies the verified archive identity and
bytes. An internal shared replay path may avoid duplicate logic. Do not use a process-global
override, network fallback, or a test-only hook as the production route.

The companion path must verify that the archive is a regular Git-tracked artifact at the
addendum's content-addressed path in the reviewed evidence revision, hash its exact bytes,
reject unsafe or incomplete extraction, and stage the binaries in the replay's own temporary
root. Before `store.Open`, compare the complete staged file set, modes, digests, and any safe
link targets to the inventory obtained from safe extraction of the verified archive; reject
missing or extra entries, escaping/unsafe links, and unsupported file types. Require regular,
executable `bin/pg_ctl` and the declared server executable at their expected paths. A missing
`pg_ctl` fails before store open,
even if a library cache is populated. Pass the staged directory as `store.Config.BinariesDir`
to `store.Open`; do not place it in `RuntimeDir`, which embedded Postgres clears on startup.
Confirm the launched server executable matches the declared `postgres_binary_sha256` and
capture effective database settings. Any missing, changed, unsafe, or unbound input fails
closed before a successful replay result. The default legacy call remains unchanged; the
offline companion must neither use a cache fallback nor attempt a remote fetch.

Before coding, add the reviewed interface and invariants to the live
[`twin` SPEC](../../kernel/internal/twin/SPEC.md), plus a precise fixture plan. A measured archive
size, provenance/license check, and clean-clone retrieval check are a go/no-go for committing
the binary. The accepted contract requires real Git-tracked bytes, not LFS pointers or a digest
alone. If that cost is unacceptable, stop and seek an amended contract and fresh review rather
than weakening L12/L17 in code.

## Acceptance evidence for this slice

- Pin the exact source revision, evidence revision, archive path/size/SHA-256, extracted file
  inventory, launched `postgres` executable digest, database settings, commands, and outcomes.
  Use only a throwaway fixture ledger with an exclusive writer boundary; no production data.
- From a fresh clone with network disabled and empty Postgres runtime/cache directories, supply
  the committed archive to the temporary store and run one valid isolated replay. Capture the
  complete source ledger before and after, including a record beyond `ThroughSeq`, and require
  byte-for-byte equality. Require temporary-store removal on success and failure. The fixture
  records canonical before/after ledger bytes and SHA-256 values, not just event counts.
- Make the no-fetch claim observable: deny outbound access while allowing loopback Postgres
  traffic, and record attempted DNS resolution and non-loopback connection or fetch calls for
  the replay process and its children with a test-side witness. Pin the witness command,
  coverage, and trace artifact in the case manifest; if observation fails, the case is
  `UNVERIFIED`. The expected attempt count is zero on success and every failure path.
  Include missing `bin/pg_ctl` with an otherwise populated library cache; it must be rejected
  before `store.Open`, without using the cache or attempting DNS/network access. A failed
  request under blocked egress alone is not proof of zero fetch attempts.
- Run negative controls for missing/altered archive, unsafe extraction, and wrong launched
  binary. Also force a projection failure **after** the temporary store opens (for example,
  a validated event whose projector rejects its payload). Require no successful result, close
  and removal of the temporary store, zero fetch attempts, and byte-for-byte equality of the
  complete source ledger under the exclusive-writer boundary. Pin canonical before/after
  ledger bytes and SHA-256 values, including a record beyond `ThroughSeq`, for this failure.
- Commit test-only, byte-exact expected `Result` and `Proof` fixtures for the same fixed source,
  request, and baseline on both the legacy and companion calls. Pin the canonical test encoding
  of every existing `Result` field, including both exact snapshot maps and sequence bindings,
  before implementation; give those bytes and the separately encoded six-field proof bytes
  raw SHA-256 digests. The proof encoding is exactly the accepted addendum's encoding, not a
  new wire proof. Compare each call field-by-field to its fixture and to the other call, then
  compare both byte strings and digests. The legacy regression call need not itself establish
  clean-clone offline operation; that claim applies to the companion path only.
- Record a case manifest against L17 and the Postgres prerequisite of L12. Do **not** mark L12
  `PASS`: its committed Go toolchain, selected-module `.mod`/`.zip`, executable, projector
  manifest, lock bytes, snapshots, proof bytes, and full clean-clone reproduction are separate
  work. Nor does this slice prove L01–L16, P5 `O-1` satisfaction, or detector power.
- Run direct fixture checks and the bare `make check`; distinguish expected negative-test output
  from actual failures. A passing author-run gate is necessary but not acceptance. Pin the
  implementation commit and corpus digests for a fresh non-author implementation verdict of
  `ACCEPT`, `AMEND`, or `REJECT`.

## Stop points and exclusions

First obtain owner scope approval and a fresh non-author review of this exact proposal. Resolve
any `AMEND` and review the amended revision before code. Then settle the archive budget and
exact API/SPEC amendment before implementation. Stop if the embedded-Postgres library cannot
consume the verified staged bytes without a fetch, if the launched executable cannot be bound,
or if an exclusive ledger witness cannot be established. Do not infer a green L12 from a local
cache or a failed network test.

This slice adds no lock evaluator, regeneration-as-repair, snapshot provider, external feed,
event kind, wire identity, writable truth, production traffic, or requirement-satisfaction
verdict. A later separately authorized slice must close the remaining L01–L16 controls and
obtain its own independent implementation review before any behavior-lock acceptance.
