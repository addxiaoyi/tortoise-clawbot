# Tortoise - Build Go Server
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "   Tortoise - Build Go Server" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
Write-Host ""

$Gobin = "$env:LOCALAPPDATA\Programs\Go\go\bin\go.exe"
if (-not (Test-Path $Gobin)) {
    Write-Host "ERROR: Go not found at $Gobin" -ForegroundColor Red
    exit 1
}

$env:Path = "$env:LOCALAPPDATA\Programs\Go\go\bin;$env:Path"

Write-Host "[1/3] Go version:" -ForegroundColor Cyan
go version

Write-Host ""
Write-Host "[2/3] Running go mod tidy..." -ForegroundColor Cyan
Set-Location "D:\qwq\项目\tohelp\server"
go mod tidy

Write-Host ""
Write-Host "[3/3] Building server..." -ForegroundColor Cyan
go build -o tortoise-server.exe .\cmd\server

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   Build Successful!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Binary: D:\qwq\项目\tohelp\server\tortoise-server.exe" -ForegroundColor Yellow
} else {
    Write-Host "ERROR: Build failed!" -ForegroundColor Red
}
