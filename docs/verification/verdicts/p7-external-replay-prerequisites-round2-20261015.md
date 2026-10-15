# P7 external replay prerequisites — round-2 documentary review

**Reviewed commit:** `a24fc19b29206d0bd2469f5a957af1edde0d9500`
**Reviewer identity:** fresh independent non-author context (`Locke`), distinct from round-1
reviewer `Godel`
**Reviewed artifacts and SHA-256:**

- `docs/decisions/VD-p7-external-replay-prerequisites-2026-10-15.md` — `29954655dcabe097b1cc8a8df3aebd30e93aed94f3e6a8661c5d2942b4ff0181`
- `docs/plans/P7-replay-lock-external-prerequisites-amendment-draft1.md` — `b506d45c6eb2d4fe29193baac0bb2dbbe5c450fd6d2b67dda739654dacad4de6`
- `docs/verification/p7-replay-lock-prerequisites/README.md` — `ffbefe79f1dcec36e85b6a1196b30b503af0717ae396607f6515b9a93e5ccd27`

## VERDICT: ACCEPT — amended documentary proposal only

The reviewer inspected the exact amended revision, the
[round-1 `AMEND`](p7-external-replay-prerequisites-round1-20261015.md), the accepted P7
plan, addendum, matrix, first-slice scope and verdict, and relevant P4 `twin`/`store` code and
SPECs. HEAD and all three SHA-256 values matched. The reviewed files had no worktree diff;
commit whitespace checks passed. Two pre-existing untracked paths were untouched. The
reviewer edited no files and found no new blocking, high, medium, or low findings.

All five round-1 findings are resolved as documentary obligations:

- The amendment explicitly overlays the accepted P7 plan's replay-from-committed-artifacts
  exit criterion; the old empty-cache/no-network claim is ineligible for `PASS`.
- Phase B passes verified `txz` bytes and expected digest to the companion, which rehashes,
  stages inside its own temporary root, uses `BinariesDir`, and removes the root on success
  and failure.
- The setup-source roles, platform, format, consumed member, and projector-manifest digest
  mappings are explicit.
- All HTTP redirects are rejected; using a direct final URL requires separate review.
- The excluded Go/Postgres binary distributions and built executable are named, while reviewed
  `.mod`/`.zip` module source artifacts remain Git-tracked.

This acceptance is for the exact amended **documentary proposal** only. It does not accept a
setup-source manifest, companion API, live `twin` SPEC change, fixture plan/corpus, runtime
L12/L17 result, provenance or license clearance, requirement satisfaction, full-gate success,
or product implementation. The reviewer ran no runtime tests or full gate. Each later step
needs its own authorization, evidence, and independent implementation verdict where applicable.
