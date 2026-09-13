[CmdletBinding()]
param(
    [switch]$InstallTools
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '.')).Path
Set-Location $repo

Write-Host "Proofbound Windows setup: $repo"

if ($InstallTools) {
    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        throw 'winget is required for -InstallTools. Install App Installer from Microsoft Store first.'
    }
    $packages = @(
        @{ Id = 'GoLang.Go'; Command = 'go' },
        @{ Id = 'jqlang.jq'; Command = 'jq' },
        @{ Id = 'GnuWin32.Make'; Command = 'make' },
        @{ Id = 'golangci.golangci-lint'; Command = 'golangci-lint' },
        @{ Id = 'BurntSushi.ripgrep.MSVC'; Command = 'rg' }
    )
    foreach ($package in $packages) {
        if (-not (Get-Command $package.Command -ErrorAction SilentlyContinue)) {
            winget install --id $package.Id --exact --source winget --accept-source-agreements --accept-package-agreements
        } else {
            Write-Host "Already available: $($package.Command)"
        }
    }
    Write-Host 'Tools installed. Open a new PowerShell window so PATH changes take effect.'
}

$required = @('git', 'go', 'jq', 'make', 'golangci-lint', 'rg')
$missing = @($required | Where-Object { -not (Get-Command $_ -ErrorAction SilentlyContinue) })
$gitBash = Join-Path ${env:ProgramFiles} 'Git\bin\bash.exe'
if (-not (Test-Path $gitBash)) { $missing += 'Git Bash (Program Files\Git\bin\bash.exe)' }
if ($missing.Count -gt 0) {
    Write-Warning ('Missing tools: ' + ($missing -join ', ') + '. Re-run with -InstallTools or install them manually.')
} else {
    git --version
    go version
    jq --version
    make --version | Select-Object -First 1
    golangci-lint version
    rg --version | Select-Object -First 1
    & $gitBash --version | Select-Object -First 1
}

if (-not (Test-Path '.git')) {
    git init -b main
    git config core.autocrlf true
    git config core.eol crlf
    Write-Host 'Initialized Git repository with Windows line-ending normalization.'
}

Write-Host 'Migration corpus is present. Executable kernel/scripts are not present yet; they must be regenerated from the carried specs.'
Write-Host 'Next validation command: .\check-windows.ps1'
