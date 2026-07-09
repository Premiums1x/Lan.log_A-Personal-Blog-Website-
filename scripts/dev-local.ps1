$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$startScript = Join-Path $root "scripts\start.ps1"
$adminDir = Join-Path $root "web\admin"

Write-Host "Starting local blog backend..." -ForegroundColor Cyan
Start-Process powershell.exe -WorkingDirectory $root -ArgumentList @(
    "-NoProfile",
    "-ExecutionPolicy", "Bypass",
    "-NoExit",
    "-File", $startScript
)

Start-Sleep -Milliseconds 500

Write-Host "Starting admin Vite dev server..." -ForegroundColor Cyan
Start-Process powershell.exe -WorkingDirectory $adminDir -ArgumentList @(
    "-NoProfile",
    "-ExecutionPolicy", "Bypass",
    "-NoExit",
    "-File", $startScript,
    "-Admin"
)

Write-Host ""
Write-Host "Local services are starting in separate PowerShell windows." -ForegroundColor Green
Write-Host "Public site: http://localhost:8080"
Write-Host "Admin UI:    http://localhost:5174/admin"
