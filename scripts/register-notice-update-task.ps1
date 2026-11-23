# Registers the AE Check notice watcher to run at user logon.

$ErrorActionPreference = "Stop"

# Ensure the script is running with Administrator privileges
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Error "This script must be run as Administrator. Please open PowerShell as Administrator and try again."
}

$taskName = "AECheckNoticeUpdateWatcher"
$scriptPath = Resolve-Path "$PSScriptRoot\watch-notice-update.ps1"
$workingDirectory = Resolve-Path "$PSScriptRoot\.."

$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`"" `
    -WorkingDirectory $workingDirectory

$trigger = New-ScheduledTaskTrigger -AtLogOn
$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive -RunLevel Limited

Register-ScheduledTask `
    -TaskName $taskName `
    -Action $action `
    -Trigger $trigger `
    -Principal $principal `
    -Description "Checks Another Eden notice hash and runs AE Check update-recent when due." `
    -Force -ErrorAction Stop | Out-Null

wevtutil.exe set-log "Microsoft-Windows-TaskScheduler/Operational" /enabled:true
if ($LASTEXITCODE -ne 0) {
    Write-Warning "Scheduled Task history could not be enabled. User notifications will still work."
}

Write-Host "Registered scheduled task: $taskName"
