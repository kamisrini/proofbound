#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

setup=setup-windows.ps1
check=check-windows.ps1
rg -q 'BurntSushi\.ripgrep\.MSVC' "$setup"
for tool in "'git'" "'go'" "'jq'" "'make'" "'golangci-lint'" "'rg'"; do
  rg -q --fixed-strings "$tool" "$setup"
done
rg -q '\$gitBash' "$setup"
rg -q '\$packages = @\(' "$setup"
rg -q 'if \(-not \(Get-Command \$package\.Command' "$setup"
rg -q "ProgramFiles.*Git\\\\bin\\\\bash\.exe" "$check"
rg -q "WSL is not native Windows evidence" "$check"
rg -q "make check" "$check"
echo 'ok windows-contract'
