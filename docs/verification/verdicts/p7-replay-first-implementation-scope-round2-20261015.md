# P7 first implementation slice scope review — round 2

**Reviewed commit:** `beea925a7b4ea4f5f599aabb7feef7738b11efd8`
**Reviewer identity:** fresh independent non-author context (`Kepler`), distinct from round-1
reviewer `Hume`
**Reviewed artifact and SHA-256:**

- `docs/plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md` — `32c74a68fa3a8f74bd7fb6d736acb40c3a9b84ade23d2bea382c08bcf10dfe39`

## VERDICT: ACCEPT — amended documentation scope only

The reviewer inspected the exact committed amended proposal and its round-1 diff, the round-1
`AMEND`, accepted P7 plan and decision, replay addendum and fixture matrix, round-4 documentary
verdict, P4 `twin`/`store` code and SPECs, and the pinned embedded-Postgres startup/cache/fetch
source. Read-only checks bound the commit and blob identity, file SHA-256, worktree status, and
diff whitespace. The reviewer edited no files. Findings: none at blocking, high, medium, or
low severity.

All three round-1 findings are resolved as proposed testable obligations:

- The complete staged-tree and executable `bin/pg_ctl` check precedes `store.Open`, with a
  zero-fetch-attempt witness and missing-`pg_ctl`/populated-cache negative control.
- A failure after temporary-store startup must show close/removal, no successful result, and
  byte-identical complete-ledger evidence including a record after `ThroughSeq`.
- Committed byte-exact `Result` and six-field `Proof` fixtures bind both legacy and companion
  calls with field comparisons, bytes, and SHA-256, without changing frozen proof identity.

The scope targets L17 and the Postgres prerequisite of L12, but does not claim an L12 `PASS`.
No new blocking contradiction or over-claim was found.

## Decision and non-claims

Accept the exact amended **scope proposal** only. Before product code, resolve the archive
size/provenance and clean-clone retrieval go/no-go, specify and review the exact API/SPEC and
fixture plan, and obtain separate implementation authorization. A later implementation needs
its own non-author verdict. This review does not accept product code, runtime fixtures, an L12
or L17 result, behavior-lock or regeneration implementation, P5 requirement satisfaction,
`make check`, or implementation acceptance. The reviewer ran no runtime tests or full gate.
