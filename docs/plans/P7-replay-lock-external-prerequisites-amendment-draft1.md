# P7 replay lock — external-prerequisites evidence amendment

**Status:** ACCEPTED FOR DOCUMENTARY PLANNING by fresh non-author
[round-2 review](../verification/verdicts/p7-external-replay-prerequisites-round2-20261015.md)
of exact commit `a24fc19b29206d0bd2469f5a957af1edde0d9500`. This status annotation is
post-review; the normative contract bytes are those at that commit. No product code, binary
corpus, runtime result, or implementation acceptance is authorized. The
[owner direction](../decisions/VD-p7-external-replay-prerequisites-2026-10-15.md) excludes
binary artifacts from Git, not non-code text or source files. Read with the prior
[addendum](P7-replay-lock-spec-addendum-draft1.md),
[matrix](P7-replay-lock-fixture-matrix-draft1.md), and
[accepted first-slice scope](P7-replay-lock-first-implementation-slice-proposal-draft1.md).

## Exact replacement boundary

If and only if this amendment receives a fresh non-author `ACCEPT`, it replaces the prior
[P7 plan's evidence-controls exit criterion](P7-behavior-locks-regeneration-plan-draft1.md)
only where it demands clean-clone replay **from committed artifacts**, plus the addendum,
matrix, and accepted first-slice requirements that the Go toolchain `tar.gz`, embedded-Postgres
distribution/`txz`, and built executable bytes be Git-tracked. The downloaded module
`.mod`/`.zip` source bytes remain Git-tracked under the accepted evidence route. The new L12
claim is **prepared offline replay from a plain clean clone**: Phase A fetches and verifies
reviewed public binary prerequisites with network permitted; Phase B replays with egress
denied. The original empty-cache/no-network clean-clone case and replay solely from committed
artifacts remain ineligible for `PASS`; record them as superseded claims, not passing cases.
The plan's other evidence controls and exit criteria remain required. The old round-4
addendum/matrix and round-2 first-slice verdicts remain valid records of what was reviewed
then, not approval of this change. Until a new verdict, their original contract is accepted
and implementation remains blocked.

Keep Git-tracked, content-addressed *small evidence* in the reviewed revision: lock and parent
bytes, canonical full events and whole-ledger witnesses, baseline/candidate snapshots, exact
`vera.replay.v1` proof bytes, projector and toolchain manifests, setup-source manifest,
downloaded module `.mod`/`.zip` byte artifacts, fixture expectations, case outcomes, and
independent reviews. Configurations, READMEs, and other appropriate nonbinary files remain
eligible for Git. No database data directory, third-party binary archive, ambient module
cache, or built executable is committed. The accepted canonical lock shape,
full-event/ledger evidence, proof encoding, projector-manifest field
shape and hash algorithms, P5 intent binding, and frozen P4 behavior do not change.

## Preparation and evidence route

The companion [README](../verification/p7-replay-lock-prerequisites/README.md) is the human
setup route. Before a case can pass, commit and review a small canonical
`docs/verification/p7-replay-lock-fixtures/setup-sources.json` listing each external input:
its exact public HTTPS source and both downloaded and consumed byte identities. Its bytes are
UTF-8 RFC 8785 canonical JSON with exactly one top-level `inputs` array. Each entry has exactly
`role`, `platform`, `format`, `source_url`, `download_size_bytes`, `download_sha256`,
`consumed_member`, `consumed_size_bytes`, and `consumed_sha256`; no nulls or unknown members.
The `role` is exactly `go-toolchain` or `postgres`, each occurring once, sorted by UTF-8
`role` bytes. `platform` is the exact `<goos>/<goarch>` pair and must equal the projector
manifest's `build.goos`/`build.goarch`. Both sizes are positive integers, and SHA-256 values
are lowercase hex. The `go-toolchain` entry requires `format: "tar.gz"`,
`consumed_member: "self"`, and identical download/consumed sizes and SHA-256s; that SHA-256
must equal `build.toolchain_archive_sha256`, and `build.toolchain_archive_format` must be
`tar.gz`. The `postgres` entry requires `format: "jar+txz"` and the exact, unique, regular,
safely addressed `.txz` member path inside the distribution JAR; its consumed size and SHA-256
describe those extracted member bytes. That SHA-256 must equal
`runtime.postgres_archive_sha256`, and `runtime.postgres_archive_format` must be `txz`.
`source_url` is the exact direct public HTTPS download URL for the outer bytes. **Reject every
HTTP redirect**, including one to another public origin; do not follow a 3xx response. If a
public endpoint requires redirection, obtain and separately review a direct final URL or amend
this schema before using it. Verify the downloaded bytes against the manifest before opening
the JAR or archive; verify the nested member separately before extraction.
The selected module graph determines the required Git-tracked `.mod`/`.zip` artifact set under
the existing addendum. Reject duplicate roles, missing/extra inputs, private URLs, any
redirect, or a digest-only record with no obtainable source. Bind the manifest's SHA-256 and
evidence revision in the case manifest. This is test evidence, not a new product wire identity.

Phase A, network permitted: from a plain clone, fetch the exact Go archive and the exact
embedded-Postgres distribution container and nested `txz` into a dedicated external setup
directory outside the clone. Disable redirect following. Resolve the selected Go module
`.mod` and `.zip` bytes from the reviewed Git revision, not an ambient module cache. Verify
each external download before extraction and
every consumed byte against the reviewed setup-source and projector manifests; separately
verify the Git-tracked module bytes against the projector manifest and applicable `go.sum`.
Safely unpack into empty directories; check the complete extracted Go toolchain tree and
the Postgres staged tree, including safe links, executable `bin/pg_ctl`, and the actual server
executable. Reject unrecorded Go cache/workspace,
`replace`, vendor, or build inputs. Build the executable from the verified toolchain, source,
and modules; require `executable_sha256` to match and run that exact output. Record the exact
requested and final URL (which must match), command, time, size, digest, extracted inventory,
build flags, and effective database settings. The setup directory is not committed and is not
trusted by path alone.

Phase B, network denied but loopback permitted: the separately reviewed companion entry point
receives the exact verified Postgres `txz` bytes from Phase A plus their expected
`runtime.postgres_archive_sha256`; it rehashes them before use and rejects a mismatch. It
does not receive or trust a pre-staged tree path, ambient cache, or network URL. The companion
safely extracts the verified archive into its own fresh replay temporary root, rechecks the
complete staged file set, modes, digests, and safe links against the Phase-A verified inventory,
and passes that root's binaries directory through `store.Config.BinariesDir`. It must not
use `RuntimeDir`, which store startup clears. Before `store.Open`, missing or changed
`bin/pg_ctl` or server executable must fail even if a cache is populated. Close the store and
remove the entire temporary root, including staged binaries, on success and every failure
path; no external Phase-A setup directory is cleaned up by replay. Observe zero DNS/non-loopback
fetch attempts by replay and child processes, not merely a failed network request. Require
complete source-ledger before/after byte equality, temporary-store cleanup on success and
post-open failure, and exact result/proof fixtures as in the accepted first-slice scope.

This preserves the accepted companion-entry-point and temporary-root staging/cleanup shape,
but **replaces** its Git-tracked-archive-path and clean-clone retrieval obligations with the
reviewed Phase-A external source and byte identity. The exact companion API, live `twin` SPEC
amendment, and fixture plan still require separate review and authorization before code.

No product runtime fetch is authorized. Network use belongs only to explicit setup; a replay
that downloads a byte, falls back to an ambient cache, or cannot prove zero attempts is
`UNVERIFIED`. A setup failure or absence of reviewed setup-source manifest is also
`UNVERIFIED`, not a pass. The setup phase cannot be relabeled as an offline clean-clone run.

## Case changes proposed for fresh review

| Case | Replacement obligation |
|---|---|
| L12 | Rename the claim to *prepared offline replay from a plain clean clone*. Fetch and verify only the pinned public binary prerequisites in Phase A, use committed module bytes, then run the companion in Phase B with egress denied and loopback allowed. Pin both phase records and all result/proof digests. The old empty-cache/no-network and committed-artifacts-only claims are not eligible for `PASS`. |
| L15 | Change/miss a committed `.mod`/`.zip` or locally built executable output while source Git blobs are fixed; require `UNVERIFIED`. Raw SHA-256 and `go.sum` checks both remain. |
| L16 | Fetch the pinned public Go archive in Phase A; verify its raw digest, safe extraction, exact nested toolchain manifest, complete file set, modes, and hashes. Missing or changed bytes and unrecorded cache inputs yield `UNVERIFIED`. No Go binary archive in Git. |
| L17 | Fetch the pinned public Maven distribution in Phase A and verify its container and nested Postgres `txz`. In Phase B, give the verified `txz` bytes and digest to the companion; it stages binaries inside its temporary root and uses `BinariesDir`. Missing/mismatched bytes, unsafe links, unbound server executable, incomplete cleanup, cache fallback, or any fetch attempt in Phase B yields `UNVERIFIED`. No Postgres binary archive in Git. |

All other L01–L11, L13–L14, full-ledger witnesses, frozen proof checks, P5 obligation review,
calibrated mutation controls, and independent implementation review remain required. This is
not an L12/L17 result and does not satisfy a behavior lock by itself.

## Stop points

Stop and seek another decision/review if a public artifact cannot be fetched without a
redirect or byte-matched, if redistribution/provenance cannot be established, if the selected
build is nondeterministic under pinned inputs, if the Postgres library cannot use the
companion's verified staging path without fetch/cache fallback, or if the setup manifest
cannot enumerate every behavior-affecting input.
Do not substitute a private feed, a second writable truth, or a looser digest-only claim.
