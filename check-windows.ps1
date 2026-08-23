[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Set-Location (Resolve-Path (Join-Path $PSScriptRoot '.'))

if (-not (Test-Path '.git')) { throw 'Not a Git repository. Run .\setup-windows.ps1 first.' }
$requiredDocs = @('CLAUDE.md', 'README.md', 'ROADMAP.md', 'notes/state.md', 'docs/laws.lock', 'docs/invariants.lock')
$missing = @($requiredDocs | Where-Object { -not (Test-Path $_) })
if ($missing.Count -gt 0) { throw ('Missing required corpus files: ' + ($missing -join ', ')) }

Write-Host 'Corpus check: PASS'
if (-not (Test-Path 'Makefile')) {
    Write-Warning 'Makefile is absent because this was migrated as text-only. Full make check is unavailable until the scaffold is regenerated.'
    exit 0
}
make check

