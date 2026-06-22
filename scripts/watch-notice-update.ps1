# Run from Windows Task Scheduler at user logon.

$ErrorActionPreference = "Stop"

$noticeUrl = "https://api-ap.another-eden.games/asset/notice_v2/list?language=ko"
$stateDir = Join-Path $env:LOCALAPPDATA "aecheck"
$statePath = Join-Path $stateDir "notice-update-state.json"
$repoRoot = Resolve-Path "$PSScriptRoot\.."
$aecheckPath = Join-Path $repoRoot "aecheck.exe"

function Get-NoticeHash {
    $response = Invoke-WebRequest -Uri $noticeUrl -UseBasicParsing
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($response.Content)
    $hashBytes = [System.Security.Cryptography.SHA256]::HashData($bytes)
    return [Convert]::ToHexString($hashBytes).ToLowerInvariant()
}

function Read-State {
    if (-not (Test-Path -LiteralPath $statePath)) {
        return [pscustomobject]@{}
    }

    return Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
}

function Write-State($state) {
    if (-not (Test-Path -LiteralPath $stateDir)) {
        New-Item -ItemType Directory -Path $stateDir | Out-Null
    }

    $state | ConvertTo-Json | Set-Content -LiteralPath $statePath -Encoding UTF8
}

$now = Get-Date
$today = $now.Date
$hash = Get-NoticeHash
$state = Read-State

if (-not $state.lastHash) {
    $state = [pscustomobject]@{
        lastHash = $hash
        lastCheckedAt = $now.ToString("o")
    }
    Write-State $state
    Write-Host "Initial notice hash saved: $hash"
    exit 0
}

if ($state.lastHash -ne $hash) {
    $state.lastHash = $hash
    $state.lastChangedAt = $now.ToString("o")
    $state.updateRecentAfter = $today.AddDays(2).ToString("yyyy-MM-dd")
    $state.lastCheckedAt = $now.ToString("o")
    Write-State $state
    Write-Host "Notice hash changed. update-recent scheduled after $($state.updateRecentAfter)."
    exit 0
}

$state.lastCheckedAt = $now.ToString("o")

if ($state.updateRecentAfter) {
    $updateRecentAfter = [DateTime]::ParseExact($state.updateRecentAfter, "yyyy-MM-dd", $null)
    if ($today -ge $updateRecentAfter) {
        if (-not (Test-Path -LiteralPath $aecheckPath)) {
            throw "aecheck.exe not found. Run scripts\setup-windows.ps1 first."
        }

        Push-Location $repoRoot
        try {
            & $aecheckPath update-recent
            if ($LASTEXITCODE -ne 0) {
                throw "update-recent failed with exit code $LASTEXITCODE"
            }
        } finally {
            Pop-Location
        }

        $state.updateRecentAfter = $null
        $state.lastTriggeredAt = $now.ToString("o")
        Write-State $state
        Write-Host "update-recent completed."
        exit 0
    }
}

Write-State $state
Write-Host "No update-recent trigger needed."
