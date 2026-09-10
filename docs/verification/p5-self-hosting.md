# P5 self-hosting packet

Status: implementation claim recorded; independent acceptance pending.

Semantic authority:

- `docs/decisions/VD-p5-intent-provenance-2026-09-10.md`
- `docs/verification/verdicts/p5-founder-ratification.md`
- `docs/verification/verdicts/p5-adjudication-round1.md`

Exact intent chain:

- BD: `intent.records:BD-proofbound-p5-a1b2c3@810c2932ed868ada9c0b9da389dc0c89880b1a6c658ba29fce5ab8aefc32745c`
- BR: `intent.records:BR-intent-chain-d4e5f6@913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562`
- CI: `intent.records:CI-implement-p5-0a1b2c@f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424`

Implementation durability checkpoints:

- `45af90d` — Proofbound identity migration
- `dd52348` — schemas and hostile fixtures
- `20d42d6` — intent providers
- `f88bb2f` — exact commit claims
- `b29351f` — projections and reports
- `930e05e` — verdict/review validation and deployment join
- `91cade9` — semantic gates in canary
- `d3c936c` — self-hosting records introduced before their first commit claim

The commit carrying this packet uses the explicit trailer
`Intent: CI-implement-p5-0a1b2c`. Its commit event, rather than this prose, supplies the exact
commit-to-CI revision binding. Evidence, requirement review, independent obligation verdict,
mutation results, final gate promotion, and final verification are appended without rewriting this
record.
