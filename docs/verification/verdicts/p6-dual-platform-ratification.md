# Founder ratification — P6 dual-platform execution

**Date:** 2026-09-13

**Founder statement on receipt:** “okay, i like the idea.”

**Applies to:** P6 Task 3 C8-002 native-platform acceptance

This is a declared-authority record. The statement approves the immediately preceding
recommendation to retain both Linux and native Windows support, using Git Bash/MSYS2 as the Bash
execution layer behind the PowerShell entry point. It does not claim that Windows has passed yet.

## Ratified platform contract

1. Linux remains a supported execution platform.
2. Native Windows remains a supported execution platform. WSL is not native Windows evidence.
3. The Windows operator path is PowerShell setup/launch plus a Bash-compatible environment capable
   of running the repository's GNU Make recipes (`SHELL := /usr/bin/env bash`).
4. The Windows prerequisite set is Git, Go >=1.26, `jq`, GNU Make, `golangci-lint`, Bash, and the
   standard command-line tools used by the checked scripts.
5. C8-002 closes only after a real native Windows run records tool versions, `check-windows.ps1`,
   `make check`, and the required witnessed/ledger checks with exit status 0.

This decision preserves the existing platform path represented by the tracked PowerShell setup and
check entry points; it does not authorize a new product capability, provider, event kind, or vision
ladder phase. It supersedes only the proposed Linux-only narrowing in the unaccepted Windows evidence
record.
