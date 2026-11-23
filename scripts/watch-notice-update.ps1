# Run from Windows Task Scheduler at user logon.

$ErrorActionPreference = "Stop"

$noticeUrl = "https://api-ap.another-eden.games/asset/notice_v2/list?language=ko"
$stateDir = Join-Path $env:LOCALAPPDATA "aecheck"
$statePath = Join-Path $stateDir "notice-update-state.json"
$repoRoot = Resolve-Path "$PSScriptRoot\.."
$aecheckPath = Join-Path $repoRoot "aecheck.exe"
$runnerPath = Join-Path $PSScriptRoot "run-update-recent.ps1"

function Show-UserMessage([string]$message, [int]$timeoutSeconds = 300) {
    $msgPath = Join-Path $env:SystemRoot "System32\msg.exe"
    try {
        & $msgPath $env:USERNAME "/TIME:$timeoutSeconds" $message | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "msg.exe failed with exit code $LASTEXITCODE"
        }
    } catch {
        try {
            Add-Type -AssemblyName PresentationFramework
            [System.Windows.MessageBox]::Show($message, "AECheck") | Out-Null
        } catch {
            Write-Warning "Could not show user message: $message"
        }
    }
}

function Get-NoticeHash {
    $response = Invoke-WebRequest -Uri $noticeUrl -UseBasicParsing
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($response.Content)
    $sha256 = [System.Security.Cryptography.SHA256]::Create()
    try {
        $hashBytes = $sha256.ComputeHash($bytes)
    } finally {
        $sha256.Dispose()
    }
    return ([BitConverter]::ToString($hashBytes) -replace "-", "").ToLowerInvariant()
}

function Read-State {
    if (-not (Test-Path -LiteralPath $statePath)) {
        return [pscustomobject]@{}
    }

    return Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
}

function Set-StateValue($state, [string]$name, $value) {
    $property = $state.PSObject.Properties[$name]
    if ($null -eq $property) {
        $state | Add-Member -NotePropertyName $name -NotePropertyValue $value
        return
    }
    $property.Value = $value
}

function Write-State($state) {
    if (-not (Test-Path -LiteralPath $stateDir)) {
        New-Item -ItemType Directory -Path $stateDir | Out-Null
    }

    $state | ConvertTo-Json | Set-Content -LiteralPath $statePath -Encoding UTF8
}

function Invoke-NoticeUpdateWatcher {
    $now = Get-Date
    $today = $now.Date
    $hash = Get-NoticeHash
    $state = Read-State

    if (-not $state.lastHash) {
        Set-StateValue $state "lastHash" $hash
        Set-StateValue $state "lastCheckedAt" $now.ToString("o")
        Write-State $state
        Show-UserMessage "AECheck notice watcher started successfully."
        return
    }

    if ($state.lastHash -ne $hash) {
        $updateRecentAfter = $today.AddDays(2).ToString("yyyy-MM-dd")
        Set-StateValue $state "lastHash" $hash
        Set-StateValue $state "lastChangedAt" $now.ToString("o")
        Set-StateValue $state "updateRecentAfter" $updateRecentAfter
        Set-StateValue $state "lastCheckedAt" $now.ToString("o")
        Write-State $state
        Show-UserMessage "AECheck detected an update notice. The interactive update will open after $updateRecentAfter at login."
        return
    }

    Set-StateValue $state "lastCheckedAt" $now.ToString("o")

    if ($state.updateRecentAfter) {
        $updateRecentAfter = [DateTime]::ParseExact($state.updateRecentAfter, "yyyy-MM-dd", $null)
        if ($today -ge $updateRecentAfter) {
            if (-not (Test-Path -LiteralPath $aecheckPath)) {
                throw "aecheck.exe not found. Run scripts\setup-windows.ps1 first."
            }
            if (-not (Test-Path -LiteralPath $runnerPath)) {
                throw "Interactive update runner not found: $runnerPath"
            }

            $arguments = "-NoProfile -ExecutionPolicy Bypass -File `"$runnerPath`""
            $process = Start-Process `
                -FilePath "powershell.exe" `
                -ArgumentList $arguments `
                -WorkingDirectory $repoRoot `
                -Wait `
                -PassThru
            if ($process.ExitCode -ne 0) {
                throw "update-recent failed or was closed with exit code $($process.ExitCode)"
            }

            Set-StateValue $state "updateRecentAfter" $null
            Set-StateValue $state "lastTriggeredAt" $now.ToString("o")
            Set-StateValue $state "lastCompletedAt" (Get-Date).ToString("o")
            Set-StateValue $state "lastError" $null
            Write-State $state
            Show-UserMessage "AECheck manual update completed successfully."
            return
        }
    }

    Write-State $state
    if ($state.updateRecentAfter) {
        Show-UserMessage "AECheck notice check succeeded. Manual update is pending for $($state.updateRecentAfter)." 30
        return
    }
    Show-UserMessage "AECheck notice check succeeded. No update is pending." 30
}

try {
    Invoke-NoticeUpdateWatcher
} catch {
    $message = $_.Exception.Message
    try {
        $state = Read-State
        Set-StateValue $state "lastFailedAt" (Get-Date).ToString("o")
        Set-StateValue $state "lastError" $message
        Write-State $state
    } catch {
        # The user-facing error below remains the primary failure signal.
    }
    Show-UserMessage "AECheck notice watcher failed: $message"
    [Console]::Error.WriteLine("AECheck notice watcher failed: $message")
    exit 1
}
