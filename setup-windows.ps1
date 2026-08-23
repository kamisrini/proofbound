[CmdletBinding()]
param(
    [switch]$InstallTools
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '.')).Path
Set-Location $repo

Write-Host "VERA Windows setup: $repo"

if ($InstallTools) {
    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        throw 'winget is required for -InstallTools. Install App Installer from Microsoft Store first.'
    }
    winget install --id GoLang.Go --exact --source winget --accept-source-agreements --accept-package-agreements
    winget install --id jqlang.jq --exact --source winget --accept-source-agreements --accept-package-agreements
    winget install --id GnuWin32.Make --exact --source winget --accept-source-agreements --accept-package-agreements
    winget install --id golangci.golangci-lint --exact --source winget --accept-source-agreements --accept-package-agreements
    Write-Host 'Tools installed. Open a new PowerShell window so PATH changes take effect.'
}

$required = @('git', 'go', 'jq', 'make', 'golangci-lint')
$missing = @($required | Where-Object { -not (Get-Command $_ -ErrorAction SilentlyContinue) })
if ($missing.Count -gt 0) {
    Write-Warning ('Missing tools: ' + ($missing -join ', ') + '. Re-run with -InstallTools or install them manually.')
} else {
    git --version
    go version
    jq --version
    make --version | Select-Object -First 1
    golangci-lint version
}

if (-not (Test-Path '.git')) {
    git init -b main
    git config core.autocrlf true
    git config core.eol crlf
    Write-Host 'Initialized Git repository with Windows line-ending normalization.'
}

Write-Host 'Migration corpus is present. Executable kernel/scripts are not present yet; they must be regenerated from the carried specs.'
Write-Host 'Next validation command: .\check-windows.ps1'

