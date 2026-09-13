# P6 native Windows platform run

**Date:** 2026-09-13

**Frozen commit:** `b3303100f649c3987003a959668a261ae9b7b3f6`

**Entry point:** `check-windows.ps1` under native Windows PowerShell, with a Windows-filesystem
checkout at `C:\Users\thamm\proofbound-c8-002-final` and `core.autocrlf=false`.

**Result:** PASS

This is native Windows evidence, not WSL evidence. The disposable checkout was transferred from the
verified repository bundle and checked out at the frozen commit. The explicit acceptance PATH
included Git `bin` and `usr\bin`, the Proofbound Go toolchain, GNU Make, golangci-lint, ripgrep, and
jq. The toolchain was:

| Tool | Version |
|---|---|
| Go | `go1.27.0 windows/amd64` |
| GNU Make | `3.81` |
| golangci-lint | `2.13.2` |
| ripgrep | `15.2.0` |
| jq | `1.8.2` |
| Git | `2.32.0.windows.2` |
| PowerShell | native `powershell.exe` |

## Commands and results

| Command | Exit | Result |
|---|---:|---|
| `check-windows.ps1` | 0 | Native corpus/platform check passed. |
| focused `go test ./internal/twin -run TestReplayRejectsDecreasingCandidates -count=1 -v` | 0 | Embedded PostgreSQL replay test passed; the earlier failure was a disposable fixed-port collision. |
| `make -f Makefile check` | 0 | Bare Windows gate passed; expected negative-fixture `index stale; run make index` output remained non-fatal; `golangci-lint` reported `0 issues.` |
| `make -f Makefile check-witnessed` | 0 | Witness `01M2ECQZRBW3P6T2AMYSNZMH12.json` published with `exit_code=0`, `git_sha=b3303100f649c3987003a959668a261ae9b7b3f6`, and `git_dirty=false`. |
| `go run ./cmd/proofbound migrate historical-evidence` | 0 | `imported=5`; exact three-record intent prerequisite closure plus two cited P5 evidence envelopes. |
| first `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=204 checks appended=0 sessions appended=0 reviews appended=23` |
| second `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=0 checks appended=0 sessions appended=0 reviews appended=0` |
| `go run ./cmd/proofbound verify` | 0 | Complete projection, snapshot, rebuild, witness lookup, and latest-witness binding passed. |
| `go run ./cmd/proofbound rebuild` | 0 | Explicit rebuild completed with no error after verify. |
| `go run ./cmd/proofbound report intent CI-implement-p5-0a1b2c` | 0 | Self-hosted intent chain rendered `state=SATISFIED`; exact intent/requirement/review/evidence proofs resolved. |
| `go run ./cmd/proofbound report requirement BR-intent-chain-d4e5f6` | 0 | Self-hosted active requirement rendered with exact authorization proof. |

The five-record archive is validated by exact line SHA-256, fixed sequence, identity, timestamps,
payload hash, and connector fields before append-only import. The migration does not alter normal
connectors or failed-closed dangling-reference validation. The native Windows filesystem path,
embedded PostgreSQL lifecycle, two-pass sync idempotence, verify, rebuild, and reports all passed.
C8-002 is evidence-bound by this record.

The checkout, database, and witness spool are disposable and are not themselves repository evidence.
