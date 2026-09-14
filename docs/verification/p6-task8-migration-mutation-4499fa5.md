# P6 Task 8 — `internal/migration` mutation evidence

- Frozen implementation commit: `4499fa5`
- Exact package tree: `6f9cb480578e34f902d9940825d5f2ff70cc022c`
- Command: `PATH=/home/thamm/go/bin:$PATH make mutants PKG=internal/migration`
- Calibration: `neutral=survived invalid=invalid lethal=killed`
- Candidates: `33`
- Killed: `33`
- Invalid: `0`
- Survived: `0`

The sweep ran to its complete summary after the migration fixture-root correction and the
archive-validation invariant tests were committed. No interrupted or partial result is used.

This is package mutation evidence only. Task 8 remains open until the final package packet binds
all production packages at one frozen implementation commit and a non-author verdict for the same
commit and exact package tree is committed.
