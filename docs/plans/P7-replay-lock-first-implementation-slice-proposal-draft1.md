# P7 replay lock — first implementation slice proposal

**Status:** PROPOSED 2026-10-15; owner scope approval and fresh non-author scope review
are pending. This document authorizes no product code, runtime corpus, or acceptance claim.
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
root. Pass the staged directory as `store.Config.BinariesDir` to `store.Open`; do not place it in
`RuntimeDir`, which embedded Postgres clears on startup. Confirm the launched server executable
matches the declared `postgres_binary_sha256` and capture effective database settings. Any
missing, changed, unsafe, or unbound input fails closed before a successful replay result. The
default legacy call remains unchanged; the offline companion must never silently fetch.

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
  byte-for-byte equality. Require temporary-store removal on success and failure.
- Run negative controls for missing/altered archive, unsafe extraction, wrong launched binary,
  and attempted network dependency. Each must fail closed, preserve the complete source ledger,
  and leave no temporary store. Verify that the original `ReplayIsolated` route and
  `vera.replay.v1` proof fields/digests are unchanged for existing vectors.
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
