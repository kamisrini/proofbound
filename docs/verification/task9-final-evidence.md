# Task 9 / P1 close evidence

Date: 2026-08-26
Frozen implementation: `2457245`

## Scope

- `vera.verdict.v1` front matter and committed-only review ingestion for 19 historical verdict artifacts.
- `review.verdict` ledger events and `reviews_view` finding/event/artifact proof rows.
- `vera sync reviews` and review participation in `sync all` and `vera verify`.
- Weekly review rendering and ledger-ordered red-verdict/change/next-verdict chain output.
- Blocking spec-first gate at `kernel/internal/specfirst` under bare `make check`.

## Commands and results

- `make check` — PASS; emits the known non-failing `index stale; run make index` diagnostic.
- `make check-witnessed` — PASS; emitted the latest `make check` witness.
- `go test ./... -count=1` — PASS.
- Fresh empty PostgreSQL database `vera_task9_final2`: `vera sync reviews` — `listed=19 appended=19 existing=0 malformed=0`.
- Fresh empty PostgreSQL database: `go run ./cmd/vera verify` — PASS. The verifier completed its two-pass sync, projection apply/rebuild snapshot comparison, and latest witness assertion.
- Fresh empty PostgreSQL database `vera_task9_final3`: strict review projection and chain integration tests — PASS.
- Fresh empty PostgreSQL database `vera_task9_final4`: `vera sync reviews` — `listed=19 appended=19 existing=0 malformed=0`; `go run ./cmd/vera verify` — PASS.
- Fresh PostgreSQL database: `go run ./cmd/vera report week` — PASS; review findings rendered with event proof IDs. The chain count was zero because the committed-only run ingested verdicts contiguously before Git events, leaving no commit sequence strictly between adjacent verdict events; this is ledger ordering, not an omitted implementation.

## Known limitation

The first integration attempt used the retained shared database and was contaminated by prior fixtures. It reported missing projection rows and non-idempotent pre-existing session/check data. Those results were not used as acceptance evidence; the verifier was rerun against the fresh database above.
