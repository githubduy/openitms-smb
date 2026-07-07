# OpenITMS-SMB — cài đặt 1 lệnh trên Windows (chạy từ thư mục giải nén .zip).
# Chạy trong PowerShell "Run as Administrator". Idempotent: chạy lại KHÔNG phá dữ liệu.
# Không cần internet. Layout bundle: bin\ mariadb\ plugins\ templates\ certs\ config\ install.ps1
#
# Tương đương install.sh (Linux) nhưng: MariaDB chạy TCP 127.0.0.1:3306 + service Windows,
# app chạy bằng Scheduled Task lúc khởi động (thay cho systemd).
[CmdletBinding()]
param(
  [string]$Prefix       = "C:\OpenITMS",
  [string]$DbPassword   = $env:OPENITMS_DB_PASSWORD,
  [string]$AdminLogin   = "admin",
  [string]$AdminPass    = "quickwin123",       # banner UI ép đổi ở lần login đầu
  [int]$Port            = 3000
)

$ErrorActionPreference = "Stop"
$BinName = "openitms-app.exe"                   # tên binary trong bin\
if (-not $DbPassword) { $DbPassword = "OpenITMS@MariaDB#2026" }  # known default, đủ an toàn
$Here = Split-Path -Parent $MyInvocation.MyCommand.Path
$Data = Join-Path $Prefix "data"

function Need-Admin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  if (-not (New-Object Security.Principal.WindowsPrincipal($id)).IsInRole(
        [Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Cần chạy PowerShell bằng 'Run as Administrator'."
  }
}
function Rand-Key {
  $b = New-Object 'System.Byte[]' 32
  [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
  ([Convert]::ToBase64String($b) -replace '[^A-Za-z0-9]','').Substring(0,32)
}

Need-Admin

Write-Host "==> [1/6] Copy bundle vao $Prefix" -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $Prefix,"$Data\db","$Data\tmp","$Prefix\certs" | Out-Null
foreach ($d in "bin","mariadb","plugins","templates","config","licenses") {
  if (Test-Path "$Here\$d") { Copy-Item "$Here\$d" $Prefix -Recurse -Force }
}

$MariaBin = Join-Path $Prefix "mariadb\bin"
$Mysqld   = Join-Path $MariaBin "mariadbd.exe"
$Mysql    = Join-Path $MariaBin "mariadb.exe"
$InstallDb= Join-Path $MariaBin "mariadb-install-db.exe"

Write-Host "==> [2/6] MariaDB (bundled, 127.0.0.1:3306)" -ForegroundColor Cyan
if (-not (Test-Path "$Data\db\mysql")) {
  & $InstallDb "--datadir=$Data\db" | Out-Null
  Write-Host "    datadir khoi tao xong"
} else {
  Write-Host "    datadir da co - giu nguyen (idempotent)"
}

# Service MariaDB Windows (OpenITMS-DB)
if (-not (Get-Service "OpenITMS-DB" -ErrorAction SilentlyContinue)) {
  & $Mysqld "--install" "OpenITMS-DB" "--datadir=$Data\db" | Out-Null
}
Start-Service "OpenITMS-DB" -ErrorAction SilentlyContinue
Write-Host "    cho MariaDB len..."
for ($i=0; $i -lt 30; $i++) {
  try { & $Mysql "-u" "root" "-e" "SELECT 1" 2>$null | Out-Null; if ($LASTEXITCODE -eq 0) { break } } catch {}
  Start-Sleep 1
}

Write-Host "==> [3/6] Database + user app (idempotent)" -ForegroundColor Cyan
$sql = @"
CREATE DATABASE IF NOT EXISTS openitms CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'openitms'@'localhost' IDENTIFIED BY '$DbPassword';
CREATE USER IF NOT EXISTS 'openitms'@'127.0.0.1' IDENTIFIED BY '$DbPassword';
ALTER USER 'openitms'@'localhost' IDENTIFIED BY '$DbPassword';
ALTER USER 'openitms'@'127.0.0.1' IDENTIFIED BY '$DbPassword';
GRANT ALL PRIVILEGES ON openitms.* TO 'openitms'@'localhost';
GRANT ALL PRIVILEGES ON openitms.* TO 'openitms'@'127.0.0.1';
FLUSH PRIVILEGES;
"@
$sql | & $Mysql "-u" "root"

Write-Host "==> [4/6] Config lan dau" -ForegroundColor Cyan
$cfgDir  = Join-Path $Prefix "config"
$cfgPath = Join-Path $cfgDir "config.json"
New-Item -ItemType Directory -Force -Path $cfgDir | Out-Null
if (-not (Test-Path $cfgPath)) {
  $cfg = [ordered]@{
    dialect = "mysql"
    mysql   = [ordered]@{ host = "127.0.0.1:3306"; user = "openitms"; pass = $DbPassword; name = "openitms" }
    port    = "$Port"
    tmp_path = "$Data\tmp"
    cookie_hash           = (Rand-Key)
    cookie_encryption     = (Rand-Key)
    access_key_encryption = (Rand-Key)
  }
  ($cfg | ConvertTo-Json -Depth 5) | Set-Content -Path $cfgPath -Encoding ascii
}

Write-Host "==> [5/6] Admin mac dinh" -ForegroundColor Cyan
$App = Join-Path $Prefix "bin\$BinName"
& $App user add --admin --login $AdminLogin --name Admin --email admin@localhost `
  --password $AdminPass --config $cfgPath 2>$null
if ($LASTEXITCODE -ne 0) { Write-Host "    admin da ton tai - bo qua (idempotent)" }

Write-Host "==> [6/6] Scheduled Task 'OpenITMS' (chay luc khoi dong)" -ForegroundColor Cyan
$action = New-ScheduledTaskAction -Execute $App -Argument "server --config `"$cfgPath`"" -WorkingDirectory $Prefix
$trigger = New-ScheduledTaskTrigger -AtStartup
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$settings = New-ScheduledTaskSettingsSet -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -StartWhenAvailable
Register-ScheduledTask -TaskName "OpenITMS" -Action $action -Trigger $trigger -Principal $principal `
  -Settings $settings -Force | Out-Null
Start-ScheduledTask -TaskName "OpenITMS"

$ip = (Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
       Where-Object { $_.IPAddress -notlike '169.*' -and $_.IPAddress -ne '127.0.0.1' } |
       Select-Object -First 1 -ExpandProperty IPAddress)
if (-not $ip) { $ip = "127.0.0.1" }

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Green
Write-Host "  OpenITMS-SMB cai dat XONG."
Write-Host "  URL:       http://${ip}:$Port"
Write-Host "  Dang nhap: $AdminLogin / $AdminPass   (DOI NGAY o lan login dau)"
Write-Host "  Certs:     bo .pem vao $Prefix\certs la dung duoc ngay"
Write-Host "  Service:   Get-Service OpenITMS-DB ; Get-ScheduledTask OpenITMS"
Write-Host "  Go cai:    installer\windows\uninstall.ps1"
Write-Host "===============================================================" -ForegroundColor Green
