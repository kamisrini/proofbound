# P7 replay-lock companion — synthetic archive-control review packet

**Review type:** independent implementation review of the next bounded test slice

**Implementation commit:** `bd5459592cc2ea53a95845a0c61fc8bb6f5e3e93`

**Parent accepted implementation:** `c5f65a478bcb3191f08ec567f9c22633e01a129b`

**Intent:** `CI-p7-companion-1a2b3c`

**Review date:** 2026-10-19

## Scope

This packet freezes the synthetic archive-control test expansion added after the independent
acceptance of the verified-archive preflight/staging implementation. It covers only local,
synthetic validation and documentation alignment. It does not consume a real PostgreSQL archive,
run PostgreSQL, validate a third-party binary, run L12/L17, create runtime fixture evidence, or
reopen `DEF-P7-001`.

## Corpus

The reviewer must inspect the exact implementation commit and these immutable files:

| Path | SHA-256 |
|---|---|
| `kernel/internal/twin/verified_archive_test.go` | `916e5f5e25e9bbddb4397951e87cfd2da53198376437a5c7f3f923d0d7b41c2a` |
| `kernel/internal/twin/SPEC.md` | `fec7f8947807e03139fc0ddb79f5fbe2c88f1cd90da207446038727254437424` |
| `docs/plans/P7-readiness-manifest-plan-draft1.md` | `faea631a07820a797c9a7f3855902f91b2d4f0926138acadc9e8e11299eb4915` |
| `docs/plans/P7-replay-lock-companion-fixture-plan-draft1.md` | `093179a9a69e90d90292607ad70c50eb8a671658cb21ea38b77bbc0df69f4b27` |
| `docs/verification/p7-replay-lock-prerequisites/README.md` | `5d200e59ac1b58e613e9fdc5946a77cec3b8afc7a7a57426c4134811e87c29e9` |
| `ROADMAP.md` | `7f2d4c566c855c1a39b444d8e83b4ee74f7cc353d139592b60c99139df5daff6` |

## Added controls

The test expansion adds coverage for:

1. sorted/unique inventory and required executable shape;
2. unsupported and non-executable modes;
3. unsafe archive path forms;
4. cancellation before staging creates a binaries directory; and
5. unexpected staged files and post-stage mode drift.

The aligned documents explicitly distinguish these synthetic controls from runtime evidence. The
existing accepted implementation boundary and dormant binary deferral are unchanged.

## Verification evidence

- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go test ./internal/twin -count=1`: PASS.
- Focused verbose archive-control tests: PASS.
- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check`: PASS; `golangci-lint` reported
  `0 issues`.
- Expected `index stale; run make index` output was a negative-test message, not a gate failure.
- `scripts/intent-delivery-applicability.sh --commit bd5459592cc2ea53a95845a0c61fc8bb6f5e3e93`:
  applicable with explicit `Intent:` trailer.
- `scripts/deferrals.sh --check`: PASS; `DEF-P7-001` remains `OPEN`/`POSTPONED` and dormant.
- No PostgreSQL distribution was downloaded, approved, or executed.

## Review questions

1. Do the new tests actually exercise the claimed fail-closed archive and inventory boundaries?
2. Are the tests deterministic, platform-appropriate, and free of hidden real-binary/runtime
   claims?
3. Does the SPEC map each new control to one machine-readable proving test without weakening the
   existing invariant contract?
4. Do the planning and prerequisite documents accurately distinguish accepted preflight work from
   unverified runtime work?
5. Does the change preserve the dormant status of `DEF-P7-001` and avoid third-party binary
   validation?

## Verdict requirement

Return `ACCEPT` only for this exact commit and corpus if there are no blocking, high, medium, or low
findings. Any verdict must remain limited to this synthetic archive-control test expansion and its
documentation alignment. It must not be relabeled as runtime fixture acceptance, L12/L17, behavior
lock acceptance, or binary provenance/licensing clearance.
