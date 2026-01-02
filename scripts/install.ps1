# Installation script for Spectre on Windows
# Run with: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

$Version = "0.1.0"
$InstallDir = "$env:ProgramFiles\Spectre"
$RepoUrl = "https://github.com/spectre-lang/spectre"

Write-Host "Installing Spectre $Version..." -ForegroundColor Green

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "Found Go: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go is not installed. Please install Go first." -ForegroundColor Red
    Write-Host "Visit: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

# Create temporary directory
$TempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

try {
    # Clone repository
    Write-Host "Downloading Spectre..." -ForegroundColor Green
    Set-Location $TempDir
    git clone $RepoUrl spectre
    Set-Location spectre

    # Build
    Write-Host "Building Spectre..." -ForegroundColor Green
    go build -o spectre.exe ./cmd/spectre

    # Install
    Write-Host "Installing to $InstallDir..." -ForegroundColor Green
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item spectre.exe "$InstallDir\spectre.exe"

    # Add to PATH if not already present
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Host "Adding Spectre to PATH..." -ForegroundColor Green
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "Machine")
        $env:Path += ";$InstallDir"
    }

    Write-Host "✓ Spectre installed successfully!" -ForegroundColor Green
    Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
    
    # Verify installation
    if (Get-Command spectre -ErrorAction SilentlyContinue) {
        spectre --version
    } else {
        Write-Host "Installation complete. Run 'spectre --help' for usage." -ForegroundColor Green
    }
} finally {
    Remove-Item -Recurse -Force $TempDir
}

