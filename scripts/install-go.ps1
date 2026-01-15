#Requires -Version 5.1
<#
.SYNOPSIS
  Cài đặt Go (golang) tự động trên Windows: tải bản stable mới nhất, giải nén, thêm PATH.

.DESCRIPTION
  - Lấy danh sách phiên bản từ https://go.dev/dl/?mode=json
  - Chọn zip windows-amd64 hoặc windows-arm64 theo CPU
  - Mặc định cài vào %LOCALAPPDATA%\Programs\Go (không cần quyền Admin)
  - -UseProgramFiles: cài vào %ProgramFiles%\Go (cần Administrator, PATH máy)

.PARAMETER UseProgramFiles
  Cài vào C:\Program Files\Go (hoặc %ProgramFiles%\Go). Không dùng chung với -InstallDir.

.PARAMETER Version
  Ví dụ: 1.23.4 hoặc go1.23.4. Mặc định: stable mới nhất (bỏ qua beta/rc nếu có bản stable).

.PARAMETER InstallDir
  Thư mục cài (mặc định: $env:LOCALAPPDATA\Programs\Go).

.PARAMETER Force
  Cài lại dù đã có Go đủ mới.

.PARAMETER SkipPath
  Không chỉnh sửa biến môi trường PATH của user.

.EXAMPLE
  .\scripts\install-go.ps1

.EXAMPLE
  .\scripts\install-go.ps1 -Version 1.22.10 -InstallDir "D:\sdk\go"

.EXAMPLE
  # PowerShell (Run as administrator)
  .\scripts\install-go.ps1 -UseProgramFiles
#>

param(
    [string]$Version = "",
    [string]$InstallDir = "",
    [switch]$Force,
    [switch]$SkipPath,
    [switch]$UseProgramFiles,
    # 内部使用：提权后二次启动时传入，避免重复弹 UAC
    [switch]$ReElevated
)

$ErrorActionPreference = "Stop"

# 使用 Host 自带 Write-Host，避免与 Write-Information / Write-Warning 等产生歧义或重复输出
function Write-GoInfo($msg) { Microsoft.PowerShell.Host\Write-Host "[install-go] $msg" -ForegroundColor Cyan }
function Write-GoOk($msg) { Microsoft.PowerShell.Host\Write-Host "[install-go] $msg" -ForegroundColor Green }
function Write-GoWarn($msg) { Microsoft.PowerShell.Host\Write-Host "[install-go] $msg" -ForegroundColor Yellow }

function Normalize-GoVersion([string]$v) {
    $v = $v.Trim()
    if ($v -match '^go(\d+\.\d+\.\d+)') { return $matches[1] }
    if ($v -match '^(\d+\.\d+\.\d+)') { return $matches[1] }
    return $null
}

function Compare-GoVersion([string]$a, [string]$b) {
    # Returns: -1 if a < b, 0 if equal, 1 if a > b
    $pa = $a -split '\.' | ForEach-Object { [int]$_ }
    $pb = $b -split '\.' | ForEach-Object { [int]$_ }
    for ($i = 0; $i -lt 3; $i++) {
        $da = if ($i -lt $pa.Count) { $pa[$i] } else { 0 }
        $db = if ($i -lt $pb.Count) { $pb[$i] } else { 0 }
        if ($da -lt $db) { return -1 }
        if ($da -gt $db) { return 1 }
    }
    return 0
}

function Get-WindowsArchZipName {
    $proc = $env:PROCESSOR_ARCHITECTURE
    if ($proc -eq "ARM64") { return "windows-arm64" }
    return "windows-amd64"
}

function Get-LatestStableVersion {
    $json = Invoke-RestMethod -Uri "https://go.dev/dl/?mode=json" -UseBasicParsing
    $stable = $json | Where-Object {
        $_.stable -eq $true -and ($_.version -match '^go\d+\.\d+\.\d+$')
    } | Select-Object -First 1
    if (-not $stable) {
        $stable = $json | Where-Object { $_.version -match '^go\d+\.\d+\.\d+$' } | Select-Object -First 1
    }
    if (-not $stable) { throw "Không đọc được danh sách phiên bản Go từ go.dev." }
    return $stable.version.TrimStart('go')
}

function Test-IsAdministrator {
    $p = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
    return $p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-InstallDirNeedsAdmin([string]$dir) {
    if (-not $dir) { return $false }
    try {
        $full = [System.IO.Path]::GetFullPath($dir)
    } catch {
        return $false
    }
    $pf = [Environment]::GetFolderPath("ProgramFiles")
    if ($full.StartsWith($pf, [StringComparison]::OrdinalIgnoreCase)) { return $true }
    $pf86 = ${env:ProgramFiles(x86)}
    if ($pf86 -and $full.StartsWith($pf86, [StringComparison]::OrdinalIgnoreCase)) { return $true }
    return $false
}

function Get-ExistingGoVersion {
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) { return $null }
    try {
        $out = & go env GOVERSION 2>$null
        if ($LASTEXITCODE -ne 0 -or -not $out) { return $null }
        return Normalize-GoVersion $out
    } catch {
        return $null
    }
}

if ($UseProgramFiles -and $InstallDir) {
    throw "Không dùng đồng thời -UseProgramFiles và -InstallDir."
}
if ($UseProgramFiles) {
    $InstallDir = Join-Path $env:ProgramFiles "Go"
} elseif (-not $InstallDir) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\Go"
}

if (Test-InstallDirNeedsAdmin $InstallDir) {
    if (-not (Test-IsAdministrator)) {
        if (-not $ReElevated) {
            Write-GoInfo "需要管理员权限（Program Files / 系统 PATH）。正在弹出 UAC，请在对话框中点「是」…"
            $pwshExe = (Get-Process -Id $PID -ErrorAction SilentlyContinue).Path
            if (-not $pwshExe) { $pwshExe = "powershell.exe" }
            $argList = @(
                "-NoProfile", "-ExecutionPolicy", "Bypass",
                "-File", $PSCommandPath,
                "-ReElevated"
            )
            if ($UseProgramFiles) { $argList += "-UseProgramFiles" }
            elseif ($PSBoundParameters.ContainsKey("InstallDir")) { $argList += "-InstallDir", $InstallDir }
            if ($PSBoundParameters.ContainsKey("Version")) { $argList += "-Version", $Version }
            if ($Force) { $argList += "-Force" }
            if ($SkipPath) { $argList += "-SkipPath" }
            try {
                $proc = Start-Process -FilePath $pwshExe -Verb RunAs -ArgumentList $argList -PassThru -Wait -ErrorAction Stop
            } catch {
                Write-GoWarn "无法启动提升权限的窗口: $_"
                Microsoft.PowerShell.Host\Write-Host "  请右键 PowerShell → 以管理员身份运行，再执行本脚本。" -ForegroundColor Yellow
                exit 1
            }
            if ($null -eq $proc) {
                Write-GoWarn "未启动提升进程（可能已取消 UAC）。"
                exit 1
            }
            exit $proc.ExitCode
        }
        Write-GoWarn "仍无管理员权限，无法写入: $InstallDir"
        Microsoft.PowerShell.Host\Write-Host "  请右键 PowerShell → 以管理员身份运行后再执行。" -ForegroundColor Yellow
        exit 1
    }
}

$archSuffix = Get-WindowsArchZipName
$targetVer = $Version
if (-not $targetVer) {
    $targetVer = Get-LatestStableVersion
} else {
    $n = Normalize-GoVersion $targetVer
    if (-not $n) { throw "Version không hợp lệ: $Version" }
    $targetVer = $n
}

$existing = Get-ExistingGoVersion
if ($existing -and -not $Force) {
    $cmp = Compare-GoVersion $existing $targetVer
    if ($cmp -ge 0) {
        Write-GoOk "Đã có Go $existing (>= $targetVer). Bỏ qua. Dùng -Force để cài lại."
        exit 0
    }
    Write-GoInfo "Go hiện tại: $existing — sẽ nâng lên $targetVer"
}

$zipName = "go$targetVer.$archSuffix.zip"
$baseUrl = "https://go.dev/dl/"
$zipUrl = $baseUrl + $zipName
$tempRoot = Join-Path $env:TEMP "go-install-$(New-Guid)"
$zipPath = Join-Path $tempRoot $zipName

try {
    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
    Write-GoInfo "Đang tải: $zipUrl"
    Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath -UseBasicParsing

    if (Test-Path $InstallDir) {
        Write-GoInfo "Gỡ thư mục cũ: $InstallDir"
        Remove-Item -LiteralPath $InstallDir -Recurse -Force
    }
    $parent = Split-Path -Parent $InstallDir
    if (-not (Test-Path $parent)) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }

    Write-GoInfo "Đang giải nén vào $InstallDir ..."
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::ExtractToDirectory($zipPath, $parent)

    $extracted = Join-Path $parent "go"
    if (-not (Test-Path $extracted)) {
        throw "Giải nén xong nhưng không thấy thư mục 'go' trong $parent"
    }
    if ($extracted -ne $InstallDir) {
        Move-Item -LiteralPath $extracted -Destination $InstallDir -Force
    }
} finally {
    if (Test-Path $tempRoot) {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

$goBin = Join-Path $InstallDir "bin"
if (-not $SkipPath) {
    $pathScope = if (Test-InstallDirNeedsAdmin $InstallDir) { "Machine" } else { "User" }
    $curPath = [Environment]::GetEnvironmentVariable("Path", $pathScope)
    $parts = @()
    if ($curPath) { $parts = $curPath -split ';' | ForEach-Object { $_.TrimEnd('\') } }
    $normBin = $goBin.TrimEnd('\')
    $has = $false
    foreach ($p in $parts) {
        if ($p -ieq $normBin) { $has = $true; break }
    }
    if (-not $has) {
        $newPath = if ($curPath) { "$curPath;$goBin" } else { $goBin }
        [Environment]::SetEnvironmentVariable("Path", $newPath, $pathScope)
        $env:Path = "$env:Path;$goBin"
        $scopeLabel = if ($pathScope -eq "Machine") { "PATH máy (Machine)" } else { "PATH user" }
        Write-GoOk "Đã thêm vào ${scopeLabel}: $goBin"
    } else {
        Write-GoInfo "PATH ($pathScope) đã có sẵn: $goBin"
    }
}

$verOut = & "$goBin\go.exe" version 2>&1
Write-GoOk "Hoàn tất: $verOut"
Write-GoInfo "Mở terminal mới nếu IDE chưa thấy lệnh 'go'."
