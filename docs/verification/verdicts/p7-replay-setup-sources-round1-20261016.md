# P7 replay setup sources — round-1 documentary review

**Reviewed commit:** `8f5d12fc2340e274539e9ef19158beee773c668f`
**Reviewer identity:** fresh independent non-author context (`Kierkegaard`)
**Reviewed artifacts and SHA-256:**

- `docs/verification/p7-replay-lock-fixtures/setup-sources.json` — `e312bf28cea20d516e3b21704a53ba3ebc8ac84ff988eccb7da555317b900b33`
- `docs/verification/p7-replay-lock-prerequisites/README.md` — `716e459487fdb2b467a0585aab312c55a61429013171ee6e8d91781f0e753e06`

## VERDICT: AMEND — proposed setup documentation only

The reviewer checked the exact committed files against the accepted
[external-prerequisites amendment](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md)
and its [round-2 documentary verdict](p7-external-replay-prerequisites-round2-20261015.md).
The reviewer found the manifest's schema, canonical bytes for its ASCII/integer content,
two sorted roles, and `linux/amd64` platform consistent with the amendment. Both exact
source URLs returned HTTP 200 without redirects. Streamed downloads matched the listed
sizes and SHA-256 values; the [Go publisher metadata](https://go.dev/dl/?mode=json&include=all)
and [Maven checksum](https://repo1.maven.org/maven2/io/zonky/test/postgres/embedded-postgres-binaries-linux-amd64/16.4.0/embedded-postgres-binaries-linux-amd64-16.4.0.jar.sha256)
corroborated the outer hashes. The JAR held one regular `postgres-linux-x86_64.txz` with the recorded
size and SHA-256. Its 1,058 nested members included shared libraries and no
license/notice-named file. These byte checks do not authenticate the publisher.

1. **MEDIUM — incomplete human setup route.** The README's Phase-A instructions
   ([lines 73–80](../p7-replay-lock-prerequisites/README.md)) require inventory and digest
   checks but omit explicit rejection of unrecorded Go cache/workspace, `replace`, vendor,
   or build inputs. They also omit explicit recording of both requested and final URLs,
   time, size, extracted inventory, and build flags. The accepted amendment
   ([lines 79–84](../../plans/P7-replay-lock-external-prerequisites-amendment-draft1.md))
   requires those controls. Amend the README so its human setup route cannot produce an
   incomplete Phase-A record, then obtain a fresh non-author verdict on the amended bytes.

The manifest's source and byte claims passed this documentary check, but the pair is **not
accepted**. Publisher signing identity, component license/notice coverage, projector and
Go module evidence, companion API/live SPEC, fixtures, L12/L17 runtime evidence, and
independent implementation review remain open. This verdict does not clear provenance
or distribution permission and does not authorize product implementation or a `PASS` claim.
The reviewer edited no repository files; the two pre-existing untracked paths were
untouched.
