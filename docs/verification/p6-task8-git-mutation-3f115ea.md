# P6 Task 8 — Git connector mutation evidence

This result is bound to implementation commit `3f115ea` (`test: harden CLI mutation contracts`).

- Package: `internal/connector/git`
- Package tree: `0d5f41c019a68eed08661a2e84d274f7b04a6479`
- Command: `PATH=/home/thamm/go/bin:$PATH make mutants PKG=internal/connector/git`
- Calibration: `neutral=survived invalid=invalid lethal=killed`
- Candidates: `34`
- Killed: `34`
- Invalid: `0`
- Survived: `0`

This is author-side mechanical evidence; it is not a non-author semantic verdict.
