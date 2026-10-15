# P7 first implementation slice scope review — round 1

**Reviewed commit:** `61298348e6ddcf443fbebac84873b65306c84f6a`
**Reviewer identity:** fresh independent non-author context (`Hume`)
**Reviewed artifact and SHA-256:**

- `docs/plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md` — `f52e55333e2821cb0e388db345d60b967051fdec88e2ae4c5548b887b9377702`

## VERDICT: AMEND — proposed scope only

The reviewer inspected the exact committed proposal, accepted P7 planning boundary and decision,
the accepted replay addendum and fixture matrix, the round-4 documentary verdict, relevant P4
`twin` and `store` code and SPECs, and the locally pinned embedded-Postgres source. Read-only
checks included commit/blob and SHA-256 identity, status, whitespace, and source inspection. No
files were edited by the reviewer. The recommended offline `BinariesDir` seam is coherent and
bounded, but the following acceptance controls need amendment before the scope is accepted.

## Findings

1. **HIGH — “never fetch” is not yet demonstrable.** The embedded-Postgres startup route may
   use cache or fetch when `BinariesDir/bin/pg_ctl` is missing. A replay failure with egress
   blocked does not prove that no fetch was attempted. Require verification of the complete
   staged binary tree, including `bin/pg_ctl`, before `store.Open`, and a witness of zero remote
   fetch attempts. Test a cold start with outbound access denied but loopback available,
   including a missing-`pg_ctl` case with an otherwise populated cache.
2. **MED — failure cleanup after store startup is not exercised.** Archive and extraction
   negatives can all fail before the temporary store opens. Add a post-open import or
   projection failure and require no successful result, byte-identical complete source-ledger
   evidence under the exclusive-writer boundary, and temporary-root removal. Pin before/after
   ledger bytes and digests including a record beyond `ThroughSeq`.
3. **MED — frozen-proof regression evidence needs exact bytes.** Existing vector assertions do
   not establish byte identity for the in-memory P4 `Proof`. Require committed byte-exact
   result/proof fixtures for legacy and companion calls, the addendum's six-field proof
   encoding and SHA-256, and field-by-field comparison of both outputs.

The reviewer found the other scope boundaries coherent: the proposal preserves the existing
call and P4 result semantics, uses `BinariesDir` rather than cleared `RuntimeDir`, requires
complete source preservation, leaves full L12 unpassed, and separates scope approval from
independent implementation acceptance.

## Decision and non-claims

Amend the exact scope proposal for all three findings and obtain a fresh non-author review of
the amended revision before product code. This verdict does not accept a runtime corpus, L12 or
L17 result, P5 requirement satisfaction, full-gate success, behavior-lock or regeneration code,
or any implementation. The reviewer ran no runtime tests or `make check` and graded no code.
