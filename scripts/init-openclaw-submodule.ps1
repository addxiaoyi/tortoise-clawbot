# Replaces openclaw-main/ with https://github.com/openclaw/openclaw as a git submodule.
# DESTRUCTIVE. Read docs/setup-openclaw-submodule.md first.

if ($env:TOHELP_CONFIRM_SUBMODULE_INIT -ne '1') {
    Write-Host "Refusing: set `$env:TOHELP_CONFIRM_SUBMODULE_INIT = '1' after backup." -ForegroundColor Red
    Write-Host "See docs/setup-openclaw-submodule.md"
    exit 1
}

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if (-not (Test-Path ".git")) {
    Write-Host "Error: not a git repository root (missing .git)." -ForegroundColor Red
    exit 1
}

if (Test-Path "openclaw-main") {
    Write-Host "Removing openclaw-main/ ..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force "openclaw-main"
}

Write-Host "Adding submodule openclaw-main ..." -ForegroundColor Cyan
git submodule add https://github.com/openclaw/openclaw.git openclaw-main
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
git submodule update --init --recursive
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Done. Next: install deps inside openclaw-main (see upstream README), then npm run doctor" -ForegroundColor Green
