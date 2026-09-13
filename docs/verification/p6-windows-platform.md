# P6 native Windows platform run

**Date:** 2026-09-13

**Frozen commit:** `a2e3aedddfd30edb14ec2d513647c0cfe428b0a1`

**Entry point:** `check-windows.ps1` under native Windows PowerShell, with a Windows-filesystem
checkout and `core.autocrlf=false`

**Result:** PLATFORM GREEN THROUGH SYNC; C8-002 OPEN ON SHARED HISTORICAL-VERIFY BLOCKER

This is native Windows evidence, not WSL evidence. The checkout was created from the verified
complete bundle at the frozen commit. The toolchain was:

| Tool | Version |
|---|---|
| Go | `go1.27.0 windows/amd64` |
| GNU Make | `3.81` |
| golangci-lint | `2.13.2` |
| ripgrep | `15.2.0` |
| jq | `1.8.2` |
| Git | `2.32.0.windows.2` |
| PowerShell | native `powershell.exe`, `-NoProfile -ExecutionPolicy Bypass` |

## Commands and results

| Command | Exit | Result |
|---|---:|---|
| `check-windows.ps1` | 0 | Corpus check and bare `make -f Makefile check` passed; `golangci-lint` reported `0 issues.` |
| `make -f Makefile check-witnessed` | 0 | One witness published: `01M2DXM7YBVT7S1FXJEBFW6ZXY.json`; full check remained green. |
| first `go run ./cmd/proofbound sync all` | 0 | `intent appended=3 git appended=197 checks appended=1 sessions appended=0 reviews appended=23` |
| second `go run ./cmd/proofbound sync all` | 0 | `intent appended=0 git appended=0 checks appended=0 sessions appended=0 reviews appended=0` |
| `go run ./cmd/proofbound verify` | 1 | Failed closed at `projection apply`: v2 verdict evidence `01M28TPW9C8R7ND19MNDCJ9GDG` is dangling. |

The detached checkout is valid and the Git connector now admits it after validating its `HEAD`
commit object. This was directly regression-tested and is the reason the native sync completed;
the prior failure was a real validator defect fixed in `01e0dc9`.

The verify failure is the same source-recoverability blocker recorded by the Linux fresh-clone
run in `docs/verification/p6-fresh-clone-linux.md`: the accepted P5 obligation verdict cites an
event ULID not recoverable from Git. It is not a Windows portability failure. Rebuild equality and
self-hosted intent/requirement reports were not run because verify fails closed before those stages.
C8-002 therefore remains open under the Task 3 rule requiring a complete native verify; closing it
requires resolving C8-001's founder decision on strict historical evidence portability.

The native checkout and database are disposable and are not themselves repository evidence. The
durable evidence is this record, the frozen commit, the command results above, and the committed
dual-platform ratification and semantic VD.
