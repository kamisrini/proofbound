# Task 6 final verification evidence

Frozen implementation: `d1f2f57` (`test: cover projection snapshot error paths`).

Executed on 2026-08-25 against `postgres:16-alpine` container
`proofbound-task6-postgres` at `127.0.0.1:55433`:

- `MUTANT_TEST_TAGS=integration DATABASE_URL=... PKG=internal/projections make mutants`
  - calibration: `neutral=survived invalid=invalid lethal=killed`
  - `summary candidates=83 killed=83 invalid=0 survived=0`
- `GOCACHE=/tmp/proofbound-gocache go test -race ./...`: passed
- `make check`: passed (`0 issues`; the index checker reports the environment's read-only filesystem warning)
- Fresh schema, build, and `/tmp/vera-task6-final verify`: exit `0`
- PostgreSQL projection integration tests: passed
