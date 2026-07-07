# OpenITMS-SMB - go cai dat tren Windows. Chay AS ADMINISTRATOR.
# Mac dinh GIU du lieu (data\). Dung -PurgeData de xoa sach ca database.
# (Thong bao dung ASCII de PowerShell 5.1 doc file khong loi encoding.)
[CmdletBinding()]
param(
  [string]$Prefix = "C:\OpenITMS",
  [switch]$PurgeData
)
$ErrorActionPreference = "SilentlyContinue"

Write-Host "==> Dung + xoa Scheduled Task 'OpenITMS'"
Stop-ScheduledTask -TaskName "OpenITMS"
Unregister-ScheduledTask -TaskName "OpenITMS" -Confirm:$false

Write-Host "==> Dung + xoa Scheduled Task 'OpenITMS-Gitea'"
Stop-ScheduledTask -TaskName "OpenITMS-Gitea"
Unregister-ScheduledTask -TaskName "OpenITMS-Gitea" -Confirm:$false

Write-Host "==> Dung + xoa service 'OpenITMS-DB'"
Stop-Service "OpenITMS-DB"
$mysqld = Join-Path $Prefix "mariadb\bin\mariadbd.exe"
if (Test-Path $mysqld) { & $mysqld "--remove" "OpenITMS-DB" | Out-Null }

if ($PurgeData) {
  Write-Host "==> Xoa TOAN BO $Prefix (gom database)" -ForegroundColor Yellow
  Remove-Item -Recurse -Force $Prefix
} else {
  Write-Host "==> Giu du lieu tai $Prefix\data (dung -PurgeData de xoa han)"
  Remove-Item -Recurse -Force (Join-Path $Prefix "bin"),(Join-Path $Prefix "plugins"),(Join-Path $Prefix "gitea")
}
Write-Host "Da go OpenITMS-SMB." -ForegroundColor Green
