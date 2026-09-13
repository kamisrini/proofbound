# P6 native Windows platform run

**Date:** 2026-09-12

**Entry point:** `check-windows.ps1` under native Windows PowerShell

**Result:** OPEN — prerequisite remediation required under ratified dual-platform scope

The repository entry point was invoked through real `powershell.exe` with `-NoProfile` and
`-ExecutionPolicy Bypass`; this was not a WSL Bash run. It printed:

```text
Corpus check: PASS
```

It then reached `make check` and failed because `make` is not installed or discoverable in the
native Windows PowerShell environment. No Windows green claim is made. The script itself correctly
fails rather than treating a text-corpus check as the repository gate.

The founder has now ratified retaining both Linux and native Windows in
`docs/verification/verdicts/p6-dual-platform-ratification.md`, with semantic decision
`docs/decisions/VD-p6-dual-platform-2026-09-13.md`. C8-002 remains open until the missing Windows
toolchain is available and the complete native acceptance run passes. WSL will not substitute for
that evidence.
