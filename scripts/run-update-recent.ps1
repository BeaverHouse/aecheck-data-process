# Opens the real update-recent flow in a visible interactive PowerShell window.

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path "$PSScriptRoot\.."
$aecheckPath = Join-Path $repoRoot "aecheck.exe"
$host.UI.RawUI.WindowTitle = "AECheck Manual Update"

Push-Location $repoRoot
try {
    Write-Host "Starting the AECheck manual update." -ForegroundColor Cyan
    Write-Host "Complete the CAPTCHA and LDPlayer image steps, then follow the CLI prompts."
    Write-Host "If this window is closed or the command fails, AECheck will retry at the next login."
    Write-Host ""

    & $aecheckPath update-recent
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        Write-Host ""
        Write-Host "The update failed with exit code $exitCode." -ForegroundColor Red
        Read-Host "Review the error, then press Enter to close this window"
        exit $exitCode
    }

    $confirmation = Read-Host "Did all intended DB and storage updates complete? (y/N)"
    if ($confirmation.Trim().ToLowerInvariant() -notin @("y", "yes")) {
        Write-Host "Completion was not confirmed. The update will remain pending." -ForegroundColor Yellow
        Read-Host "Press Enter to close this window"
        exit 2
    }

    Write-Host ""
    Write-Host "The update was confirmed complete." -ForegroundColor Green
    Read-Host "Press Enter to close this window"
    exit 0
} catch {
    Write-Host ""
    Write-Host "The update runner failed: $($_.Exception.Message)" -ForegroundColor Red
    Read-Host "Review the error, then press Enter to close this window"
    exit 1
} finally {
    Pop-Location
}
