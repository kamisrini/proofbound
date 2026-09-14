# P6 Task 8 — checks connector mutation evidence

This result is bound to implementation commit `3f115ea` (`test: harden CLI mutation contracts`).

- Package: `internal/connector/checks`
- Package tree: `200d781e872c063b0fd3ab2045b6760d685bd1f0`
- Command: `PATH=/home/thamm/go/bin:$PATH make mutants PKG=internal/connector/checks`
- Calibration: `neutral=survived invalid=invalid lethal=killed`
- Candidates: `36`
- Killed: `36`
- Invalid: `0`
- Survived: `0`

This is author-side mechanical evidence; it is not a non-author semantic verdict.
