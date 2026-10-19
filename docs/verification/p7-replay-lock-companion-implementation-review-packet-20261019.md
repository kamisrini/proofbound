# P7 replay lock companion — implementation review packet

**Packet status:** REVIEW REQUEST — no verdict recorded

**Review type:** fresh non-author implementation review

**Requested outcome:** `ACCEPT`, `AMEND`, or `REJECT`

**Reviewed commit:** `fd4a90122a62820ae897141e4755d7a386281367`

**Implementation commit:** `d4a2000084849eb625d16e6646b93c04d86a7ff7`

**Derived route-matrix refresh:** `849b1e7a34ba9ed2ae4a16eba18af46f9f4df131`

This packet asks an independent reviewer to assess the authorized P7 companion implementation.
The implementation adds the verified-archive preflight/staging seam and preserves the existing
legacy replay call. It does not claim a real PostgreSQL runtime result, L12/L17, binary provenance,
or behavior-lock acceptance.

## Immutable implementation corpus

| Role | Path | SHA-256 |
|---|---|---|
| live twin SPEC | `kernel/internal/twin/SPEC.md` | `5d4dd09c146463eea558cff51056ee9689e8a8391006041b97153a31f4c24fbb` |
| replay integration | `kernel/internal/twin/replay.go` | `46b343f794bed897ce502ce1ef3a2b373b8a3d6561e89959bb4f37d140d41592` |
| verified archive implementation | `kernel/internal/twin/verified_archive.go` | `1524bf56a2dc4186ec9b828e1fc67fc31526b366827a98b1463d311782872dc1` |
| verified archive tests | `kernel/internal/twin/verified_archive_test.go` | `4307d5f424a50be210d46eaa07a085d4a6e1b2bcded94741a3454d69b4e024c4` |
| implementation intent | `docs/intent/records/change-intents/CI-p7-companion-1a2b3c/0001.md` | `9a93de46ede0edd4c69f5bcabde095e1be3b6537ec6f82dcb82deff39ed4f08f` |
| route-matrix evidence | `docs/verification/p6-event-route-matrix.md` | `e364f9fafb1ebdd9ee9c3c2aad93b9af868aa9d218d95b0c59c4d55b55dd23bf` |
| accepted API/SPEC planning | `docs/plans/P7-replay-lock-twin-api-amendment-draft1.md` | `bb467a319f2689b3dc2b24c0465c4308fb386c51a372f6430bc0203ee6fea2f2` |
| accepted fixture planning | `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` | `fb9da66a993bc1c0fe8f39c0a6468755a313a29fbcac3bfc3ead47a930603bd5` |
| documentary planning verdict | `docs/verification/verdicts/p7-replay-lock-companion-round1-20261019.md` | `071ea9959729b6ea731f4ae88348007988fc745d90b46aeea7d604e4b0f93ca0` |

The review must bind its verdict to the full reviewed commit and these exact bytes, not to a later
working tree or conversation summary. Any changed byte requires a new packet and fresh review.

## Implementation summary

- Added `ReplayIsolatedWithVerifiedPostgres` with exact archive bytes, archive digest, executable
  digest, and complete extracted-file inventory.
- Rejects malformed digests, unsorted/duplicate/unsafe paths, unsupported archive types, missing
  `bin/pg_ctl` or `bin/postgres`, mode/digest mismatches, and executable substitution.
- Extracts only regular files into a fresh replay-owned root, passes the staged directory through
  `store.Config.BinariesDir`, and leaves the legacy `ReplayIsolated` path unchanged.
- Checks context cancellation during staging and relies on the existing isolated replay defer for
  store close and full temporary-root cleanup on all return paths.
- Added focused positive and negative preflight tests. Tests use a small synthetic archive only;
  no real PostgreSQL distribution was downloaded or executed.

## Review questions

1. Does the implementation match the accepted companion API/SPEC planning boundary and preserve
   the legacy `ReplayIsolated` behavior?
2. Are archive bytes, declared archive digest, inventory paths/modes/digests, and the launched
   `postgres` executable bound strongly enough to fail closed?
3. Are path traversal, unsafe links, unsupported file types, duplicates, missing files, and
   unexpected extracted files rejected before store open?
4. Does the implementation route binaries through `BinariesDir`, avoid `RuntimeDir`, and avoid
   URL/path/cache/network fallback under the available embedded-Postgres API?
5. Do context cancellation, store-open failure, projector failure, and cleanup errors preserve the
   source store and remove the replay-owned temporary root as required?
6. Are the focused tests and full pinned `make check` result sufficient for an implementation
   verdict, or are additional tests required before acceptance?
7. Are the unverified boundaries explicit: no real PostgreSQL runtime, no network witness, no
   case-manifest binding, no L12/L17 result, and no independent behavior-lock acceptance?

## Verification recorded for review

- `go test ./internal/twin`: PASS under Go 1.26.7.
- Focused verified-archive tests: PASS.
- `scripts/intent-delivery-applicability.sh --commit bee757fc223970e57ac883b12cdbc2c84594b4c6`:
  applicable with explicit `CI-p7-companion-1a2b3c` trailer.
- Final pinned `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check`: PASS with
  `golangci-lint` reporting `0 issues`; expected `index stale; run make index` negative-test
  output was not an actual failure.
- `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant. No PostgreSQL input was downloaded,
  approved, or validated.

## Acceptance bar

Return `ACCEPT` only if the implementation is correct for the authorized preflight/staging seam
and has no blocking, high, medium, or low findings. Return `AMEND` with exact findings and
required changes if any behavior, test, or evidence is incomplete. Return `REJECT` if the
implementation boundary should be abandoned.

The verdict must identify the reviewer as non-author, include the reviewed commit and this packet's
SHA-256, and state its scope. An `ACCEPT` here accepts this implementation slice only; it does not
accept runtime fixtures, L12/L17, PostgreSQL provenance/licensing, a behavior lock, regeneration,
or any provider.

## Verdict slot

No verdict yet. The authoring session must not fill this section with an acceptance decision.
