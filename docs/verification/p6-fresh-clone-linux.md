# P6 fresh-clone Linux reproducibility run

**Date:** 2026-09-13

**Frozen commit:** `b330310`

**Clone mode:** local `git clone --no-local`, detached HEAD, no copied `.proofbound` state

**Result:** PASS

The disposable clone path was `/tmp/proofbound-p6-accept.lST67s/clone`. The path and database are
not evidence; the frozen commit and command results below are. A fresh witness was produced before
the migration and first sync.

| Command | Exit | Result |
|---|---:|---|
| `PATH=/home/thamm/go/bin:$PATH make check` | 0 | All shell/Go tests passed; `golangci-lint` reported `0 issues.` Expected negative-fixture diagnostics were emitted separately by the test harness. |
| `PATH=/home/thamm/go/bin:$PATH make check-witnessed` | 0 | Fresh `vera.witness.v1` check evidence published to the disposable spool. |
| `go run ./cmd/proofbound migrate historical-evidence` | 0 | `imported=5`; exact three-record intent prerequisite closure plus two cited P5 evidence envelopes. |
| first `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=204 checks appended=0 sessions appended=0 reviews appended=23` |
| second `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=0 checks appended=0 sessions appended=0 reviews appended=0` |
| `go run ./cmd/proofbound verify` | 0 | Projection apply, snapshot, rebuild, snapshot equality, witness lookup, and latest-witness binding passed. |
| `go run ./cmd/proofbound rebuild` | 0 | Explicit rebuild completed with no error after verify. |
| `go run ./cmd/proofbound report intent CI-implement-p5-0a1b2c` | 0 | Self-hosted intent chain rendered `state=SATISFIED`; exact intent/requirement/review/evidence proofs resolved. |
| `go run ./cmd/proofbound report requirement BR-intent-chain-d4e5f6` | 0 | Self-hosted active requirement rendered with exact authorization proof. |

The archive is validated by exact line SHA-256, fixed sequence, identity, timestamps, payload hash,
and connector fields before the append-only import. The migration does not alter connectors or
failed-closed dangling-reference validation. C8-001 is evidence-bound by this run.
