# P7 replay lock addendum and fixture planning review — round 3

**Reviewed commit:** `16c6189fd6b39680e8bc5acfd16286fa520cb9a6`
**Reviewer identity:** fresh independent non-author context (`Confucius`)
**Reviewed artifacts and SHA-256:**

- `docs/plans/P7-replay-lock-spec-addendum-draft1.md` — `153bdbc81eea4a073d65752513b94dd63bf0bdbd1cb82730154b3ed95b9022b9`
- `docs/plans/P7-replay-lock-fixture-matrix-draft1.md` — `1d84a0659af506df9ab4816147cfd3724bb49db32c0777d659fe91fc9da5a81b`

## VERDICT: AMEND

The reviewer inspected the exact commit, both prior addendum reviews, the accepted P7 decision,
and relevant P4 twin, projection, and store code and SPECs. Read-only checks included the committed
diff, `git diff --check`, SHA-256 of both reviewed files, and independent hashes of the snapshot
and proof serialization controls. The reviewer edited no files.

The round-2 snapshot finding is resolved in the proposal: evidence preserves the exact
`Snapshot.Tables` key set, including the difference between an absent table and a present empty
table, without changing frozen `digestSnapshot`. The proposed executable and external Go module
`.mod`/`.zip` byte bindings also address the specific round-2 dependency-byte finding.

## Findings

1. **HIGH — Clean-clone input set incomplete.** Matrix L12 requires rebuilding from committed
   artifacts and a committed toolchain, but the addendum binds only hashes of Go toolchain files,
   not retrievable toolchain bytes. It also binds a Postgres executable hash without requiring its
   bytes to be retrievable. `ReplayIsolated` opens a temporary embedded-Postgres store, and the
   store SPEC documents a cold-start binary download. Hashes can verify bytes obtained elsewhere
   but cannot supply the committed input set claimed by L12. Specify a retrievable, verified route
   for both runtime inputs or narrow L12 and its acceptance claim.
2. **MED — Nested toolchain manifest byte shape unspecified.** The addendum defines the main
   projector manifest's canonical JSON shape, but the separately hashed toolchain manifest has no
   exact JSON envelope and member names. Verifiers could enumerate the same files and hash
   different encodings. Pin its exact canonical byte representation.

The accepted P7 subject and frozen P4 semantics are unchanged. This `AMEND` does not accept a
runtime fixture corpus, requirement satisfaction, P7 implementation, regeneration, or the full
verification gate. Amend these findings and obtain another fresh non-author review of the exact
revision before implementation.
