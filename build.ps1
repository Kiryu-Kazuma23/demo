# Build script for Fyne application
$ErrorActionPreference = "Stop"

Write-Host "Building Fyne application..." -ForegroundColor Green

# Check if Scoop is installed
if (-not (Get-Command scoop -ErrorAction SilentlyContinue)) {
    Write-Host "Scoop is not installed. Please install Scoop first." -ForegroundColor Red
    exit 1
}

# Ensure Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go is not installed. Please install Go first." -ForegroundColor Red
    exit 1
}

# Install NuGet provider if not present
if (-not (Get-PackageProvider -Name NuGet -ErrorAction SilentlyContinue)) {
    Write-Host "Installing NuGet provider..." -ForegroundColor Yellow
    Install-PackageProvider -Name NuGet -MinimumVersion 2.8.5.201 -Force | Out-Null
}

# Install BurntToast if not already installed
if (-not (Get-Module -ListAvailable -Name BurntToast)) {
    Write-Host "Installing BurntToast module..." -ForegroundColor Yellow
    Install-Module -Name BurntToast -Force -Scope CurrentUser | Out-Null
}

# Check and install required Scoop dependencies
$requiredPackages = @(
    "gcc",
    "make"
)

Write-Host "Checking Scoop dependencies..." -ForegroundColor Yellow
foreach ($package in $requiredPackages) {
    if (-not (scoop list | Select-String -Pattern $package)) {
        Write-Host "Installing $package..." -ForegroundColor Yellow
        scoop install $package
    }
}

# Set up environment variables
$env:Path = "C:\Users\$env:USERNAME\scoop\apps\gcc\current\bin;$env:Path"
$env:CGO_ENABLED = "1"
$env:CC = "gcc"

# Install required dependencies
Write-Host "Installing Go dependencies..." -ForegroundColor Yellow
go mod tidy

# Build the application
Write-Host "Building application..." -ForegroundColor Yellow
go build -o app.exe

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Starting the application..." -ForegroundColor Cyan
    # Play success sound
    [System.Media.SystemSounds]::Asterisk.Play()
    # Show Windows notification
    New-BurntToastNotification -Text "Build Complete", "Your Fyne application has been built successfully!" -Sound "Default"
    # Start the application
    Start-Process .\app.exe
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    # Play error sound
    [System.Media.SystemSounds]::Hand.Play()
    # Show Windows notification
    New-BurntToastNotification -Text "Build Failed", "There was an error building your Fyne application." -Sound "Default"
    exit 1
} 