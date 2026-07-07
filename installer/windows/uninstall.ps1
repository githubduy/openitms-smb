# OpenITMS-SMB — gỡ cài đặt trên Windows. Chạy AS ADMINISTRATOR.
# Mặc định GIỮ dữ liệu (data\). Dùng -PurgeData để xoá sạch cả database.
[CmdletBinding()]
param(
  [string]$Prefix = "C:\OpenITMS",
  [switch]$PurgeData
)
$ErrorActionPreference = "SilentlyContinue"

Write-Host "==> Dừng + xoá Scheduled Task 'OpenITMS'"
Stop-ScheduledTask -TaskName "OpenITMS"
Unregister-ScheduledTask -TaskName "OpenITMS" -Confirm:$false

Write-Host "==> Dừng + xoá service 'OpenITMS-DB'"
Stop-Service "OpenITMS-DB"
$mysqld = Join-Path $Prefix "mariadb\bin\mariadbd.exe"
if (Test-Path $mysqld) { & $mysqld "--remove" "OpenITMS-DB" | Out-Null }

# Gitea (nếu có)
Stop-ScheduledTask -TaskName "OpenITMS-Gitea"
Unregister-ScheduledTask -TaskName "OpenITMS-Gitea" -Confirm:$false

if ($PurgeData) {
  Write-Host "==> Xoá TOÀN BỘ $Prefix (gồm database)" -ForegroundColor Yellow
  Remove-Item -Recurse -Force $Prefix
} else {
  Write-Host "==> Giữ dữ liệu tại $Prefix\data (dùng -PurgeData để xoá hẳn)"
  Remove-Item -Recurse -Force (Join-Path $Prefix "bin"),(Join-Path $Prefix "plugins")
}
Write-Host "Đã gỡ OpenITMS-SMB." -ForegroundColor Green
