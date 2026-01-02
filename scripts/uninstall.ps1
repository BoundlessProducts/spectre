# Uninstallation script for Spectre on Windows
# Run with: powershell -ExecutionPolicy Bypass -File uninstall.ps1

$ErrorActionPreference = "Stop"

$InstallDir = "$env:ProgramFiles\Spectre"

Write-Host "Uninstalling Spectre..." -ForegroundColor Green

if (Test-Path "$InstallDir\spectre.exe") {
    Remove-Item -Recurse -Force $InstallDir
    Write-Host "✓ Spectre uninstalled successfully!" -ForegroundColor Green
    
    # Remove from PATH
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $NewPath = ($CurrentPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
    Write-Host "Removed from PATH." -ForegroundColor Green
} else {
    Write-Host "Spectre not found at $InstallDir" -ForegroundColor Yellow
    exit 1
}

