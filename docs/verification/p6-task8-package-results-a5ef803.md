# P6 Task 8 — current package mutation results at `a5ef803`

These author-side calibrated results bind the same frozen implementation commit
`a5ef803` (`test: close gitcmd mutation survivors`). They are mechanical evidence, not non-author
semantic verdicts.

| Package | Tree object | Candidates | Killed | Invalid | Survived | Calibration |
|---|---|---:|---:|---:|---:|---|
| `internal/cli` | `6a1a1804018be77c5a17a3d70c024c2978511683` | 234 | 234 | 0 | 0 | neutral=survived invalid=invalid lethal=killed |
| `internal/connector/checks` | `200d781e872c063b0fd3ab2045b6760d685bd1f0` | 36 | 36 | 0 | 0 | neutral=survived invalid=invalid lethal=killed |
| `internal/connector/git` | `b4a7beeb7617625a3521e6393c7570bcdc73783c` | 34 | 34 | 0 | 0 | neutral=survived invalid=invalid lethal=killed |
| `internal/connector/git/gitcmd` | `1e72a5c0c52d731ff4d85005b820db8d1208d052` | 86 | 86 | 0 | 0 | neutral=survived invalid=invalid lethal=killed |

The CLI result used a disposable embedded PostgreSQL service on port 55433; it was stopped after
the run. The other listed package runs used their package-local test suites.
