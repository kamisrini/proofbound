# P6 intent self-coverage

**Historical measurement:** 2026-09-13

**Current census recheck:** 2026-09-18

**Anchor:** immediately after `f426ca8`

**Current canary head:** `e02dbb1c0569f6f892584370767159153dc2e567`

## Historical coverage at the 2026-09-13 baseline

The complete path-only canary is committed at
[`p6-intent-applicability-canary.md`](p6-intent-applicability-canary.md). It examines 69 commits,
classifies 40 as applicable, and finds zero explicit exact `Intent:` claims. The resulting coverage
is `0 / 40 = 0.0%`. This is an observation-only result before enforcement; historical commits are
not retroactively rejected. No known false negative was found, and the ratified greater-than-10%
false-positive redesign threshold did not fire under the exact path definition.

| Datum | Result | Provenance |
|---|---|---|
| Post-anchor commits examined | 69 | `scripts/p6-intent-canary.sh --after f426ca8` |
| Applicable commits | 40 | `scripts/intent-applicability.sh` for each commit |
| Valid exact Intent claims on applicable commits | 0 | committed Git objects and canary artifact |
| Coverage | `0.0%` | numerator / denominator above |
| Known false negatives | 0 | hostile matcher tests and canary output |
| False-positive rate | `0.0%` under the ratified path definition | founder-ratified semantic VD |

## Current census interval

The current census generator input is
`8bef29020d0799c7a71efba1230c85e0860a3e4f`. Its C6 commit interval contains every commit after
`f426ca8` through that input commit; the registry represents every one of those commits with an
`intent-commit:<full-sha>` row. A refreshed canary at current HEAD
`e02dbb1c05695451961f0cad4a669501c47c95b3` examines 135 commits, including and extending past that
interval. The full path and claim classification is in
[`p6-intent-applicability-canary-task8-20260918.md`](p6-intent-applicability-canary-task8-20260918.md).

| Datum | Result |
|---|---:|
| Commits examined by the refreshed canary | 135 |
| Applicable commits | 81 |
| Applicable commits with valid exact Intent claims | 0 |
| Coverage | 0.0% |
| Known false negatives | 0 |
| False-positive rate under the ratified path definition | 0.0% |

No commit in this interval carried an explicit `Intent:` trailer, so there was no new claim to
resolve against the ledger. The refreshed result remains observation-only; it does not reject any
historical commit or widen the existing enforcement boundary.

## Current active intent chain

The native provider has one active requirement revision and two active obligations:

| Kind | Exact revision | State | Evidence |
|---|---|---|---|
| BD | `intent.records:BD-proofbound-p5-a1b2c3@810c2932ed868ada9c0b9da389dc0c89880b1a6c658ba29fce5ab8aefc32745c` | authorizes | `docs/intent/records/business-decisions/BD-proofbound-p5-a1b2c3/0001.md` |
| BR | `intent.records:BR-intent-chain-d4e5f6@913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562` | active | `docs/intent/records/requirements/BR-intent-chain-d4e5f6/0001.md` |
| CI | `intent.records:CI-implement-p5-0a1b2c@f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424` | accepted | `docs/intent/records/change-intents/CI-implement-p5-0a1b2c/0001.md` |

`O-1` and `O-2` are both active and both have exact-revision non-author `VERIFIABLE` review
outcomes in `docs/verification/verdicts/p5-requirement-review-round1.md`. The accepted P5
obligation verdict binds both outcomes to the exact evidence events
`01M28TPW9C8R7ND19MNDCJ9GDG` and `01M29HMPE5V977AR3VMW47DVDE`.

The self-hosted intent report rendered `state=SATISFIED` and the requirement report rendered the
exact authorization proof. The controlled `make delivery-enforce` boundary remains the only
enforcement point; its new applicability check blocks an applicable commit with no explicit
`Intent:` trailer, while the existing intent gates retain exact-reference validation.

## Decision

This packet closes the C6 evidence rows by recording the historical gap, the refreshed census
interval, and the current exact chain. It does not claim retroactive intent for applicable
historical commits, promote external consumers, or alter any frozen wire identity or vector.
