# P6 Task 8 — `internal/store` and `internal/twin` mutation evidence

This is calibrated author-side mechanical evidence for frozen implementation commit
`e4c8e77407699f7e089d5c1a2b3ce58df5871fbf` (`fix: preserve isolated replay sequence updates`). It
is not the required non-author semantic verdict.

The frozen package universe at this commit contains 16 production packages returned by
`go list ./internal/...` that contain non-test Go files. `internal/specfirst` is returned by
`go list` but is test-only and remains excluded under C3-017.

| Package | Tree object | Candidates | Killed | Invalid | Survived |
|---|---|---:|---:|---:|---:|
| `internal/store` | `66d8f61e08561c075b30b05fe5e1e6c33228692a` | 133 | 133 | 0 | 0 |
| `internal/twin` | `bbc59b6e7cf3e4fb1b8f48130355e7a505a3f1d3` | 31 | 31 | 0 | 0 |

Every mutation invocation reported `calibration neutral=survived invalid=invalid lethal=killed`.

## Store partitions

The active Linux mutation universe was partitioned exactly once: lock/config 1–34, equality
guards 35–58, open/startup 59–70, migration 71–74, transaction/read 75–102, imports 103–114,
compound guards 115–119, and OR guards 120–133.

| Partition | Candidates | Killed | Invalid | Survived |
|---|---:|---:|---:|---:|
| 1–34 | 34 | 34 | 0 | 0 |
| 35–58 | 24 | 24 | 0 | 0 |
| 59–70 | 12 | 12 | 0 | 0 |
| 71–74 | 4 | 4 | 0 | 0 |
| 75–102 | 28 | 28 | 0 | 0 |
| 103–114 | 12 | 12 | 0 | 0 |
| 115–119 | 5 | 5 | 0 | 0 |
| 120–133 | 14 | 14 | 0 | 0 |
| **Aggregate** | **133** | **133** | **0** | **0** |

The standalone tagged store suite passed in 4.890s. The standalone tagged twin suite passed in
11.954s. Both used the disposable PostgreSQL service, which was stopped and removed afterward.

The complete tagged twin sweep reported:

```text
summary candidates=31 killed=31 invalid=0 survived=0
```

These results contain no interrupted mutation run. They close the mechanical store/twin work only;
C3 remains open until every package result is bound to this frozen commit and a current-code
non-author verdict is committed.
