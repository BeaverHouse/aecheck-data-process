# Another Eden 이미지 복사 스크립트 (LDPlayer ADB 사용)
# 실행: PowerShell에서 .\copy_ae_images.ps1
# LDPlayer가 켜져 있고, ADB가 정상적으로 연결되어 있어야 합니다. (기타 설정에서 활성화)

$LDPlayerPath = "C:\LDPlayer\LDPlayer9"
$ADB = "$LDPlayerPath\adb.exe"
$DestPath = "$env:USERPROFILE\Pictures\aecheck"

# 로컬 폴더 생성
if (!(Test-Path $DestPath)) {
    New-Item -ItemType Directory -Path $DestPath | Out-Null
}

Write-Host "=== Another Eden 이미지 복사 ===" -ForegroundColor Cyan

# ADB 연결 확인
Write-Host "`n[1/4] ADB 연결 확인..." -ForegroundColor Yellow
& $ADB devices

# 안드로이드 내 임시 폴더 생성 및 이미지 복사 (flat하게)
Write-Host "`n[2/4] 루트 권한으로 이미지 수집 중..." -ForegroundColor Yellow

$SRC = "/data/data/net.wrightflyer.anothereden/files/contents"
$DST = "/sdcard/Pictures/aecheck"

& $ADB shell "su -c 'mkdir -p $DST'"
& $ADB shell "su -c 'find $SRC/*/files/character/stella_panel -type f \( -name *.png -o -name *.jpg -o -name *.webp \) -exec cp {} $DST/ \;'"
& $ADB shell "su -c 'find $SRC/*/files/character/command -type f \( -name *.png -o -name *.jpg -o -name *.webp \) -exec cp {} $DST/ \;'"
& $ADB shell "su -c 'find $SRC/*/files/buddy/command -type f \( -name *.png -o -name *.jpg -o -name *.webp \) -exec cp {} $DST/ \;'"
& $ADB shell "su -c 'ls $DST | wc -l'"

# PC로 pull
Write-Host "`n[3/4] PC로 파일 가져오는 중..." -ForegroundColor Yellow
& $ADB pull /sdcard/Pictures/aecheck/. $DestPath

# 안드로이드 임시 폴더 정리
Write-Host "`n[4/4] 안드로이드 임시 폴더 정리..." -ForegroundColor Yellow
& $ADB shell "rm -rf /sdcard/Pictures/aecheck"

# 결과 출력
$fileCount = (Get-ChildItem $DestPath -File).Count
Write-Host "`n=== 완료 ===" -ForegroundColor Green
Write-Host "복사된 파일: $fileCount 개"
Write-Host "저장 위치: $DestPath"
