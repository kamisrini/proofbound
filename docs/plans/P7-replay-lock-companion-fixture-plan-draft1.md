# P7 replay lock — proposed companion fixture plan

**Status:** ACCEPTED FOR DOCUMENTARY PLANNING — round-1 non-author review on 2026-10-19;
implementation separately gated. The rows below define the fixture and evidence work needed
after explicit implementation authorization. No row has been run, and no runtime result is claimed.

**Companion:** [proposed `twin` API/SPEC amendment](P7-replay-lock-twin-api-amendment-draft1.md)
and the accepted [P7 fixture matrix](P7-replay-lock-fixture-matrix-draft1.md). This plan is a
companion execution map, not a replacement for the accepted lock contract or readiness manifest.

## Fixture rules

- Every case is bound to one exact evidence revision, implementation-scope digest, setup-source
  digest, and input-artifact digest set.
- The case manifest uses the accepted `proofbound.p7.case-manifest.v1` planning shape. Until a
  case is actually run, its outcome is `PENDING` or `UNVERIFIED`, never `PASS`.
- Positive and negative controls use the same source ledger, request, baseline, and exact archive
  identity where applicable. Expected rejection is a passing negative control only when the gate
  observes the required rejection and all preservation/cleanup witnesses.
- The third-party binary provenance/licensing item remains `DEF-P7-001`, dormant under owner
  policy. Cases requiring that input cannot be run or labeled beyond `UNVERIFIED` until the owner
  explicitly activates that deferral.
- No fixture may use a private feed, ambient module/PostgreSQL cache, pre-staged binary path,
  network fallback, or test-only global override.

## Proposed case matrix

| ID | Fixture / control | Setup and mutation | Expected outcome | Required evidence | Current state |
|---|---|---|---|---|---|
| C01 | Legacy regression | Run existing `ReplayIsolated` with its current call and fixed P4 corpus | Existing behavior remains byte-compatible; no P7 claim | Result/proof bytes, source before/after, existing test output | `PENDING`; no companion implementation |
| C02 | Valid verified archive | Supply exact Phase-A `txz` bytes, archive digest, binary digest, and complete inventory | Companion replays successfully with existing `Result`/`Proof` meanings | Input digests, extracted inventory, launched executable digest, result/proof fixtures, cleanup witness | `UNVERIFIED`; binary prerequisite dormant |
| C03 | Archive digest mismatch | Flip one archive byte while retaining declared archive digest | Reject before store open; no fetch/cache fallback | Error, zero-fetch witness, no-store-open witness, cleanup witness | `PENDING` |
| C04 | Unsafe archive | Include absolute, traversal, duplicate, unsafe-link, or unsupported entry | Reject before store open | Archive bytes, extraction error, no-store-open and cleanup witnesses | `PENDING` |
| C05 | Missing `bin/pg_ctl` | Remove required executable while leaving a populated ambient library cache | Reject before store open; cache is not used | Staged inventory, cache state, no-store-open, zero-fetch witness | `PENDING` |
| C06 | Missing or wrong server executable | Remove server binary or substitute bytes while retaining declared digest | Reject before successful replay; launched identity cannot be inferred | Inventory, actual executable digest, error, cleanup witness | `PENDING` |
| C07 | No archive / cache-only control | Omit archive bytes while populating a cache and denying network | Reject closed; no cache fallback or network attempt | Input error, cache state, DNS/non-loopback witness, no-store-open | `PENDING` |
| C08 | Wrong declared binary digest | Keep archive valid but change `PostgresBinarySHA256` | Reject when actual launched binary cannot match | Declared/actual digest pair, error, cleanup witness | `PENDING` |
| C09 | Complete ledger preservation | Valid companion replay with a source record after `ThroughSeq` | Complete ledger bytes before/after are identical, not only selected prefix | Canonical ledger bytes/digests, exclusive-writer witness, result/proof | `UNVERIFIED`; binary prerequisite dormant |
| C10 | Post-open projector failure | Force a validated event to fail after temporary store opens | No successful result; source ledger unchanged; root removed | Projector error, ledger before/after, cleanup, zero-fetch, store lifecycle trace | `PENDING` |
| C11 | Legacy/companion result lock | Same fixed source/request/baseline through both calls | Existing fields, snapshots, sequence map, verdict, and six proof fields match their fixtures and each other | Canonical result/proof bytes and separate SHA-256 values | `PENDING` |
| C12 | Network observation | Run success and failure paths with non-loopback egress denied and loopback allowed | Zero DNS/non-loopback fetch attempts; blocked egress alone is insufficient | Declared policy, process/child trace, DNS trace, command/output digests | `PENDING` |
| C13 | Cleanup closure | Exercise success, malformed archive, store-open failure, projector failure, cancellation | Entire owned temporary root is absent after each case; cleanup failure is unverified | Before/after filesystem inventory and cleanup trace | `PENDING` |
| C14 | `BinariesDir` route | Instrument/configure store with staged binaries and distinct runtime directory | Staged binaries arrive through `BinariesDir`; runtime directory is not the source | Effective store config, path inventory, startup trace | `PENDING` |
| C15 | Setup-source binding | Bind case manifest to exact setup-source manifest and evidence revision | Missing, stale, or mismatched binding stays `UNVERIFIED` | Manifest bytes/digests, contract refs, review identifiers | `PENDING`; setup documentation accepted, runtime not run |
| C16 | Inventory substitution | Keep archive bytes fixed but alter one inventory path, mode, or file digest | Reject complete-set mismatch before store open | Archive digest, declared inventory, extracted inventory, error | `PENDING` |
| C17 | Cancellation closure | Cancel during extraction, store open, and projector execution | No successful result; owned root removed; source ledger preserved | Cancellation timing, lifecycle trace, ledger bytes, cleanup witness | `PENDING` |

## Evidence packet per case

Each implemented case must provide, at minimum:

1. The stable case ID, exact reviewed evidence revision, implementation-scope digest, and
   contract-reference digests.
2. Exact command argv, working directory, relevant environment digest, network policy, exit code,
   stdout/stderr digests, and expected-negative-test classification.
3. Input archive and manifest identities, extracted inventory, required executable paths and
   modes, and actual launched server digest where the case reaches launch.
4. Complete source-ledger bytes and SHA-256 immediately before and after the call, including
   records beyond `ThroughSeq`, plus the exclusive-writer witness.
5. Exact `Result` and `Proof` evidence bytes for successful calls, with field-by-field comparison
   against the legacy fixture and the accepted six-field proof encoding.
6. Store lifecycle, temporary-root before/after inventory, cleanup outcome, and network-observation
   evidence for every success and failure path.
7. The case outcome and a non-author review against immutable artifact digests. A passing author
   test is necessary but cannot be the independent implementation verdict.

## Mapping to accepted P7 controls

| Accepted control | Proposed companion cases |
|---|---|
| L05 — baseline/build provenance | C11, C15, plus the projector/build manifest supplied by the accepted addendum |
| L06 — isolated candidate and source preservation | C02, C09, C10, C13, C17 |
| L08 — proof bytes | C11 |
| L11 — hostile input | C03, C04, C06, C08, C16 |
| L12 — prepared offline replay | C02, C07, C09, C12, C15 |
| L15/L16 — build/toolchain closure | External accepted plans; not implemented by this fixture plan |
| L17 — embedded-Postgres archive closure | C02–C08, C12–C14, C16–C17 |

The cases do not close L12 or L17 by themselves. A case is not eligible for `PASS` until the
binary provenance/licensing prerequisite, exact API/SPEC amendment, setup-source binding,
implementation review, and all required witnesses are separately satisfied.

## Review and implementation gates

1. Review this fixture plan and the companion API/SPEC proposal as documentary artifacts.
2. Record any `AMEND` or `ACCEPT` verdict against exact bytes and a full commit ID.
3. Obtain explicit implementation authorization; documentary acceptance does not authorize code.
4. Resolve or explicitly activate `DEF-P7-001` before running cases that consume the pinned
   PostgreSQL input. Until then, leave those cases dormant and `UNVERIFIED`.
5. Build the fixture corpus and case manifests, run the direct controls, then run the appropriate
   full verification gate.
6. Obtain a separate independent implementation verdict. This gate is now satisfied for the
   preflight/staging companion slice by
   [`p7-replay-lock-companion-implementation-round1-20261019.md`](../verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md);
   do not treat it as runtime fixture or behavior-lock acceptance.

## Explicit non-goals

This plan does not create fixture files, download or approve third-party binaries, alter the live
`twin` API/SPEC, implement the companion, add a parser, run L12/L17, add a provider or external
feed, or reopen any dormant deferral.

## Review outcome

Round-1 non-author documentary review by Srini K accepted this fixture plan with no findings. See
[`p7-replay-lock-companion-round1-20261019.md`](../verification/verdicts/p7-replay-lock-companion-round1-20261019.md).
The acceptance is limited to documentary planning. The authorized preflight/staging implementation
was independently accepted with no findings by Srini K in
[`p7-replay-lock-companion-implementation-round1-20261019.md`](../verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md).
No runtime fixture has been run; runtime evidence, L12/L17, and binary provenance/licensing remain
separately gated.
