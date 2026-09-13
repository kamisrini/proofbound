# P6 native Windows platform run

**Date:** 2026-09-12

**Entry point:** `check-windows.ps1` under native Windows PowerShell

**Result:** DECISION REQUIRED

The repository entry point was invoked through real `powershell.exe` with `-NoProfile` and
`-ExecutionPolicy Bypass`; this was not a WSL Bash run. It printed:

```text
Corpus check: PASS
```

It then reached `make check` and failed because `make` is not installed or discoverable in the
native Windows PowerShell environment. No Windows green claim is made. The script itself correctly
fails rather than treating a text-corpus check as the repository gate.

Under `docs/plans/p6-task3-reproducibility-SPEC.md`, C8-002 can close only through either a real
green native run after installing the required toolchain or a founder-ratified support narrowing.
The P6 depth-before-width recommendation is to define Linux as the supported P6 execution platform
and retain the PowerShell files only as setup/migration helpers, without claiming native Windows
verification.
