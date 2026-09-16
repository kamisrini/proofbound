# P6 Task 8 — frozen-commit package reruns at `e4c8e77`

These are complete calibrated author-side mutation reruns against the frozen implementation
commit `e4c8e77407699f7e089d5c1a2b3ce58df5871fbf` (`fix: preserve isolated replay sequence
updates`). They are mechanical evidence only; they are not non-author semantic verdicts.

Every invocation reported `calibration neutral=survived invalid=invalid lethal=killed`. The
package-universe classification was rerun with
`scripts/tests/p6-task8-package-universe.test.sh` and passed. The rows below reached their
complete summaries with zero invalid and surviving mutants:

| Package | Exact tree object | Candidates | Killed | Invalid | Survived |
|---|---|---:|---:|---:|---:|
| `internal/core` | `9508eada402ff0231111cc11413a765311a361be` | 58 | 58 | 0 | 0 |
| `internal/connector/checks` | `200d781e872c063b0fd3ab2045b6760d685bd1f0` | 36 | 36 | 0 | 0 |
| `internal/connector/git` | `b4a7beeb7617625a3521e6393c7570bcdc73783c` | 34 | 34 | 0 | 0 |
| `internal/connector/git/gitcmd` | `1e72a5c0c52d731ff4d85005b820db8d1208d052` | 86 | 86 | 0 | 0 |
| `internal/connector/github` | `9b2aae50693322c21a987ccd5e119ef70eb66798` | 58 | 58 | 0 | 0 |
| `internal/connector/intent` | `6fad5e69fb203244d346233078e169a4b9bfc7c0` | 93 | 93 | 0 | 0 |
| `internal/connector/intent/records` | `9ac52619e76036e4c16ad6123e1679c7fece0570` | 52 | 52 | 0 | 0 |
| `internal/connector/intent/specdir` | `b50c85c27100da88ac713e2291fbb975d7b01a80` | 47 | 47 | 0 | 0 |
| `internal/connector/reviews` | `e5e00cb31d5e47037628da54b65cc2b3409d217d` | 114 | 114 | 0 | 0 |
| `internal/connector/sessions` | `5f984b21790002cf045c762865492767a491235a` | 62 | 62 | 0 | 0 |
| `internal/gates` | `b524b275ab4f6355638fcb6a7cea23a3d3075c3e` | 109 | 109 | 0 | 0 |
| `internal/migration` | `6f9cb480578e34f902d9940825d5f2ff70cc022c` | 33 | 33 | 0 | 0 |

The gates row used `MUTANT_TEST_TAGS=integration` and a disposable PostgreSQL instance. The
other rows used the package-local calibrated route without a test tag. The disposable database
was stopped after the run.

The CLI and projections reruns were started as database-backed partitions but interrupted for
operational reasons before complete summaries. Their partial output is deliberately excluded from
this record and is not acceptance evidence. Store and twin remain covered by
`docs/verification/p6-task8-store-twin-mutation-e4c8e77.md`.

Task 8 remains open: the final packet still needs complete frozen-commit CLI and projections
rows and a committed current-code non-author verdict for all 16 production packages.
