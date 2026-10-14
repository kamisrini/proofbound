# P7 replay lock addendum and fixture planning review — round 2

**Reviewed commit:** `730a474c633d0ea320f6aa1eeec4a809b4481b83`
**Reviewer identity:** fresh independent non-author context (`Mendel`)
**Reviewed artifacts and SHA-256:**

- `docs/plans/P7-replay-lock-spec-addendum-draft1.md` — `30d6b9ffdee93b5fbae1976de11f5d17f397ee58b8e55fc2c381c5e6b510fe21`
- `docs/plans/P7-replay-lock-fixture-matrix-draft1.md` — `8945233301dd87b199f6fd260ba7bb8fa22e29dc979701e274e71eabd5b96d9e`

## VERDICT: AMEND

The reviewer inspected the committed addendum and matrix against the accepted P7 planning
decision, the round-1 `AMEND`, relevant package code and SPECs. Checks were read-only: exact
commit and file SHA-256s, revision diff whitespace, and the proof serialization control hash.
No files were edited by the reviewer.

The round-1 complete-source-record, whole-ledger-before/after, and candidate-payload-integrity
findings are resolved in the **proposal**. The proposed six-field proof encoding is precise and
its no-newline control hashes to the declared
`ec7960eda35be3798ac6cb25e0853608937c6d24a2025e8135a146ac8771ca79`.
Evidence-byte pinning remains incomplete for two reasons:

1. **HIGH — Snapshot evidence conflicts with actual snapshot maps.** The addendum requires empty
   tables to be present in snapshot evidence. `projections.Snapshot.Tables` only acquires a table
   key when a row is processed (`kernel/internal/projections/projections.go`,
   `appendSnapshotDigests`). Adding an absent empty table while decoding evidence changes the map
   hashed by frozen `twin.digestSnapshot` (`kernel/internal/twin/replay.go`). Define an encoding
   that round-trips the actual map, including absent tables, without changing the P4 digest.
2. **MED — Build identity is incomplete.** The exact projector manifest shape has no executable
   digest or bytes field, though its prose claims to capture the executable's exact bytes. It
   identifies `kernel/` blobs and the Go toolchain, but not the exact downloaded dependency
   source bytes used by the build (`kernel/go.mod` declares external modules). Pin the built
   executable and verify dependency bytes, or explicitly narrow the reproducibility claim.

The accepted round-3 P7 subject and planning boundary remain accepted. This round-2 `AMEND` does
not accept the later addendum, runtime fixture corpus, requirement satisfaction, implementation,
regeneration, or the full verification gate. Amend these two findings and obtain a fresh non-author
review of the amended exact revision before implementation.
