# P7 replay setup sources — round-2 documentary review

**Reviewed commit:** `d07f3ffee70d1acd5f917d4fa06362aaac594718`
**Reviewer identity:** fresh independent non-author context (`Parfit`), distinct from
round-1 reviewer `Kierkegaard`
**Reviewed artifacts and SHA-256:**

- `docs/verification/p7-replay-lock-fixtures/setup-sources.json` — `e312bf28cea20d516e3b21704a53ba3ebc8ac84ff988eccb7da555317b900b33`
- `docs/verification/p7-replay-lock-prerequisites/README.md` — `6ef672f67702282c5f02fc19b2f1ab7c3d6897fe9caa44472931aa9b54f0c487`

## VERDICT: ACCEPT — exact proposed setup documentation only

The reviewer checked the committed bytes and amendment diff against the accepted
[external-prerequisites contract](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md),
its [documentary verdict](p7-external-replay-prerequisites-round2-20261015.md), and the
[round-1 `AMEND`](p7-replay-setup-sources-round1-20261016.md). The manifest was canonical
JSON with the required schema, two sorted roles, `linux/amd64` platform, and the specified
format and member values. The reviewer did not repeat round 1's live downloads or
authenticate either publisher. No new blocking, high, medium, or low findings were found.

The round-1 medium finding is closed. The README's
[Phase-A setup route](../p7-replay-lock-prerequisites/README.md) now explicitly rejects
unrecorded Go cache/workspace, `replace`, vendor, and other build inputs; verifies tracked
module bytes against the projector manifest and applicable `go.sum`; and requires recording
both requested and final URLs, commands, time, sizes, digests, complete extracted
inventories, build flags, versions, and effective database settings. The README still marks
the runtime fixture `UNVERIFIED`. Its statement that independent `ACCEPT` was pending was
true at the reviewed commit; this later verdict records acceptance of those exact bytes
without changing them.

This accepts only the proposed setup-source manifest and human setup documentation at the
reviewed revision. It does **not** authenticate the PostgreSQL signing key, clear component
licenses/notices or distribution permission, provide projector/module/toolchain/executable
or fixture evidence, approve the companion API or live SPEC, establish an L12/L17 runtime
`PASS`, or authorize implementation. Those remain separate blockers and review gates. The
reviewer edited no repository files; the two pre-existing untracked paths were untouched.
