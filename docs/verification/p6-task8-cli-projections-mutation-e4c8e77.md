# P6 Task 8 — CLI and projections mutation evidence at `e4c8e77`

This record contains the complete calibrated `internal/cli` and `internal/projections` reruns
against frozen implementation commit
`e4c8e77407699f7e089d5c1a2b3ce58df5871fbf`. The disposable databases ran PostgreSQL 16.4. Every
accepted invocation reported calibration `neutral=survived invalid=invalid lethal=killed`; every
candidate in the complete summaries was killed, with zero invalids and zero survivors.

## CLI pilot and full sweep

The one-candidate `internal/cli` integration pilot used a fresh disposable PostgreSQL instance. Its
calibration passed and `cli.go:73#1` was killed. The mutation invocation elapsed 15.22 seconds,
including tool startup and calibration; database provisioning was outside that timer. Based on this
measurement, the full sweep used a cap of two concurrent partitions, each with its own fresh
disposable PostgreSQL instance.

| Partition | Candidates | Killed | Invalid | Survived | Elapsed |
|---|---:|---:|---:|---:|---:|
| 1–117 | 117 | 117 | 0 | 0 | 707.11 s |
| 118–234 | 117 | 117 | 0 | 0 | 554.96 s |
| **Aggregate** | **234** | **234** | **0** | **0** | |

## Projections sweep

The frozen package has 404 candidates in the established source partitions. All four invocations
used independent calibration. The first and second core ranges ran the full tagged package test
suite; the intent/report range used the focused selector listed below. Each concurrent full-suite
partition had its own disposable PostgreSQL instance.

| Source partition | Runner range and test selection | Candidates | Killed | Invalid | Survived | Elapsed |
|---|---|---:|---:|---:|---:|---:|
| `intent_report.go` | 1–58; `^Test(CalibrationLethal\|Intent\|Obligation\|Requirement\|Derived\|SpecState\|Unverified)` | 58 | 58 | 0 | 0 | 949.16 s |
| `projections.go` | 59–202; full `integration` tagged package suite | 144 | 144 | 0 | 0 | 3,852.81 s |
| `projections.go` | 203–345; full `integration` tagged package suite | 143 | 143 | 0 | 0 | 3,311.33 s |
| `report.go` | 346–404; full `integration` tagged package suite | 59 | 59 | 0 | 0 | 701.70 s |
| **Aggregate** | | **404** | **404** | **0** | **0** | |

On the disposable PostgreSQL clusters, `checkpoint_timeout` was set to 30 minutes and
`max_wal_size` to 8 GB after checkpointer logs showed long file-sync pauses under temporary-schema
churn. Transaction and durability settings remained at their defaults. All task-local PostgreSQL
servers were stopped after their complete summaries.

## Excluded attempts

- The first filtered intent/report invocation omitted the harness-created lethal calibration test
  and stopped before candidate 1. The completed retry included it and passed calibration.
- An unbalanced 59–345 full-suite invocation was interrupted after 26 candidates to divide the
  work across the two database lanes. It produced no acceptance summary and is excluded.
- The first report attempt used an initialized cluster without the `proofbound` database. Its
  neutral control failed before candidate 1. After creating the database on the dedicated instance,
  the retry passed calibration and completed all 59 candidates.

Only the complete calibrated summaries in the tables above count toward package acceptance.
