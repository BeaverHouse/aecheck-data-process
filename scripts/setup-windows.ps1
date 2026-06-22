# Run in PowerShell

$ErrorActionPreference = "Stop"

Write-Host "Building aecheck for Windows..." -ForegroundColor Cyan
Push-Location "$PSScriptRoot\.."
try {
    $outputPath = Join-Path (Get-Location) "aecheck.exe"
    if (Test-Path -LiteralPath $outputPath) {
        Remove-Item -LiteralPath $outputPath -Force
    }

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -o aecheck.exe .
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE"
    }
    if (-not (Test-Path -LiteralPath $outputPath)) {
        throw "go build completed but aecheck.exe was not created"
    }
} finally {
    Pop-Location
}
Write-Host "Build complete. Run '.\aecheck.exe' from this directory (with .env file)." -ForegroundColor Green
