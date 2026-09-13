# P6 fresh-clone Linux reproducibility run

**Date:** 2026-09-12

**Frozen commit:** `a5383b947ecc304d49d6969aea83821bc828ab63`

**Clone mode:** local `git clone --no-local`, detached HEAD, no copied `.proofbound` state

**Result:** BLOCKED — historical verdict evidence is not source-recoverable

The disposable clone path was `/tmp/proofbound-p6-clone.CgVz69`. The path and database are not
evidence; the frozen commit and command results below are.

| Command | Exit | Result |
|---|---:|---|
| `PATH=/home/thamm/go/bin:$PATH make check` | 0 | All shell/Go tests passed; `golangci-lint` reported `0 issues.` |
| first `go run ./cmd/proofbound sync all` | 0 | `intent appended=3 git appended=173 checks appended=0 sessions appended=0 reviews appended=23` |
| second `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=0 checks appended=0 sessions appended=0 reviews appended=0` |
| `go run ./cmd/proofbound verify` | 1 | `projection apply: v2 verdict evidence 01M28TPW9C8R7ND19MNDCJ9GDG is dangling: no rows in result set` |

To exclude a missing local witness as the cause, `make check-witnessed` was then run successfully
and the new witness was ingested (`checks appended=1`, every other connector appended zero).
`proofbound verify` still failed on the same historical event ID. The failing reference is committed
verbatim in `docs/verification/verdicts/p5-obligation-verdict-round1.md`; its source check event is
present in the original machine ledger but neither its exact event envelope nor a replayable identity
mapping exists in Git. Connector-created event IDs are random ULIDs, so rerunning the check cannot
recreate the cited ID.

Rebuild equality and the self-hosted intent/requirement reports were not claimed: projection apply
fails closed before those checks. Closing C8-001 therefore requires a founder decision between a
strict, minimal portable historical-evidence mechanism and narrowing the ratified fresh-empty-ledger
claim. Silently accepting dangling evidence would weaken the P5 proof model and is not proposed.
