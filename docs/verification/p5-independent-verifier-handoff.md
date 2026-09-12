# P5 independent-verifier handoff

This packet requests two non-author artifacts for the frozen P5 implementation tip. It does not
prescribe either outcome. The verifier owns every finding and conclusion; returned files are
committed verbatim on receipt under `docs/verification/verdicts/`.

## Authority and reviewed scope

Read these before evaluating the implementation:

1. `docs/plans/P5-PROOFBOUND-INTENT-PROVENANCE-v3.md`
2. `docs/verification/verdicts/p5-founder-ratification.md`
3. `docs/verification/verdicts/p5-adjudication-round1.md`
4. `docs/decisions/VD-p5-intent-provenance-2026-09-10.md`

The reviewed commit is the exact result of `git rev-parse HEAD` in the frozen handoff checkout. It
must carry `Intent: CI-implement-p5-0a1b2c`. The exact canonical chain is:

- `intent.records:BD-proofbound-p5-a1b2c3@810c2932ed868ada9c0b9da389dc0c89880b1a6c658ba29fce5ab8aefc32745c`
- `intent.records:BR-intent-chain-d4e5f6@913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562`
- `intent.records:CI-implement-p5-0a1b2c@f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424`

The requirement has obligations `O-1` and `O-2`. The requirement author/owner is
`proofbound-maintainer`; a review declaring that same identity must fail closed.

## Required independent artifacts

First return `docs/verification/verdicts/p5-requirement-review-round1.md` using the frozen
`proofbound.requirement-review.v1` contract. It must bind the exact BR digest above, cover both
obligations exactly once, name the verifier's distinct declared identity, and record the verifier's
actual closed verifiability outcomes.

Then return `docs/verification/verdicts/p5-obligation-verdict-round1.md` using the frozen
`proofbound.obligation-verdict.v2` contract. It must bind the frozen reviewed commit, exact CI and
BR digests above, cover both obligations exactly once, and cite only actual ledger evidence event
IDs at or before the verdict observation. `ACCEPTABLE` is valid only if both independently reached
outcomes are `SATISFIED`.

Both artifacts must cite the founder ratification and round-1 adjudication in their prose. Their
top-level `artifact_sha256` values are the SHA-256 of the complete file after replacing only that
self-digest value with 64 zeroes, as specified by `internal/connector/reviews/SPEC.md`.

## Reproduction commands

From the repository root:

```text
git log -1 --oneline
make check
cd kernel && go run ./cmd/proofbound verify
cd kernel && go run ./cmd/proofbound report intent CI-implement-p5-0a1b2c
cd kernel && go run ./cmd/proofbound intent check --commit "$(git -C .. rev-parse HEAD)"
```

Mutation summaries and the controlled-boundary BLOCKED reproduction are recorded separately in
`docs/verification/p5-mutation-evidence.md` and `docs/verification/p5-delivery-readiness.md` before
the handoff tip is frozen.

