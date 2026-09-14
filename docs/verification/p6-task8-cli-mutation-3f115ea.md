# P6 Task 8 — CLI mutation evidence

This result is bound to implementation commit `3f115ea` (`test: harden CLI mutation contracts`).

- Package: `internal/cli`
- Package tree: `6a1a1804018be77c5a17a3d70c024c2978511683`
- Command: `DATABASE_URL=postgres://proofbound:proofbound@127.0.0.1:55433/proofbound?sslmode=disable MUTANT_TEST_TAGS=integration PATH=/home/thamm/go/bin:$PATH make mutants PKG=internal/cli`
- Calibration: `neutral=survived invalid=invalid lethal=killed`
- Candidates: `234`
- Killed: `234`
- Invalid: `0`
- Survived: `0`

The database was a disposable embedded PostgreSQL instance started from the repository's cached
binary directory and stopped after the run. This is author-side mechanical evidence; it is not a
non-author semantic verdict.
