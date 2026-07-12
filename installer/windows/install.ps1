# OpenITMS-SMB — cài đặt 1 lệnh trên Windows (chạy từ thư mục giải nén .zip).
# Chạy trong PowerShell "Run as Administrator". Idempotent: chạy lại KHÔNG phá dữ liệu.
# Không cần internet. Layout bundle: bin\ mariadb\ gitea\ plugins\ templates\ certs\ config\ install.ps1
#
# Tương đương install.sh (Linux) nhưng: MariaDB chạy TCP 127.0.0.1:3306 + service Windows,
# app + Gitea chạy bằng Scheduled Task lúc khởi động (thay cho systemd). Env cho app (gồm
# QUICKWIN_GITEA_*) đặt trong wrapper start.cmd để auto-repo (patch 0008) hoạt động.
[CmdletBinding()]
param(
  [string]$Prefix        = "C:\OpenITMS",
  [string]$DbPassword    = $env:OPENITMS_DB_PASSWORD,
  [string]$GiteaDbPass   = $env:OPENITMS_GITEA_DB_PASSWORD,
  [string]$GiteaAdminPass= $env:OPENITMS_GITEA_ADMIN_PASSWORD,
  [string]$AdminLogin    = "admin",
  [string]$AdminPass     = "quickwin123",       # banner UI ép đổi ở lần login đầu
  [int]$Port             = 3000
)

$ErrorActionPreference = "Stop"
$BinName = "openitms-app.exe"                    # tên binary trong bin\
if (-not $DbPassword)     { $DbPassword     = "OpenITMS@MariaDB#2026" }       # known default
if (-not $GiteaDbPass)    { $GiteaDbPass    = "OpenITMS_Gitea_2026_Secure" }  # INI-safe (không # ; ")
if (-not $GiteaAdminPass) { $GiteaAdminPass = "OpenITMS-Gitea-Admin-2026" }
$Here = Split-Path -Parent $MyInvocation.MyCommand.Path
$Data = Join-Path $Prefix "data"

function Need-Admin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  if (-not (New-Object Security.Principal.WindowsPrincipal($id)).IsInRole(
        [Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Cần chạy PowerShell bằng 'Run as Administrator'."
  }
}
function Rand-Key([int]$n = 32) {
  $b = New-Object 'System.Byte[]' 48
  [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b)
  ([Convert]::ToBase64String($b) -replace '[^A-Za-z0-9]','').Substring(0,$n)
}

Need-Admin

Write-Host "==> [1/8] Copy bundle vao $Prefix" -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $Prefix,"$Data\db","$Data\tmp","$Prefix\certs" | Out-Null
foreach ($d in "bin","mariadb","gitea","plugins","templates","config","licenses") {
  if (Test-Path "$Here\$d") { Copy-Item "$Here\$d" $Prefix -Recurse -Force }
}

$MariaBin = Join-Path $Prefix "mariadb\bin"
$Mysqld   = Join-Path $MariaBin "mariadbd.exe"
$Mysql    = Join-Path $MariaBin "mariadb.exe"
$InstallDb= Join-Path $MariaBin "mariadb-install-db.exe"

Write-Host "==> [2/8] MariaDB (bundled, 127.0.0.1:3306)" -ForegroundColor Cyan
if (-not (Test-Path "$Data\db\mysql")) {
  & $InstallDb "--datadir=$Data\db" | Out-Null
  Write-Host "    datadir khoi tao xong"
} else {
  Write-Host "    datadir da co - giu nguyen (idempotent)"
}
if (-not (Get-Service "OpenITMS-DB" -ErrorAction SilentlyContinue)) {
  & $Mysqld "--install" "OpenITMS-DB" "--datadir=$Data\db" | Out-Null
}
Start-Service "OpenITMS-DB" -ErrorAction SilentlyContinue
Write-Host "    cho MariaDB len..."
for ($i=0; $i -lt 30; $i++) {
  try { & $Mysql "-u" "root" "-e" "SELECT 1" 2>$null | Out-Null; if ($LASTEXITCODE -eq 0) { break } } catch {}
  Start-Sleep 1
}

Write-Host "==> [3/8] Database + user app (idempotent)" -ForegroundColor Cyan
@"
CREATE DATABASE IF NOT EXISTS openitms CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'openitms'@'localhost' IDENTIFIED BY '$DbPassword';
CREATE USER IF NOT EXISTS 'openitms'@'127.0.0.1' IDENTIFIED BY '$DbPassword';
ALTER USER 'openitms'@'localhost' IDENTIFIED BY '$DbPassword';
ALTER USER 'openitms'@'127.0.0.1' IDENTIFIED BY '$DbPassword';
GRANT ALL PRIVILEGES ON openitms.* TO 'openitms'@'localhost';
GRANT ALL PRIVILEGES ON openitms.* TO 'openitms'@'127.0.0.1';
FLUSH PRIVILEGES;
"@ | & $Mysql "-u" "root"

Write-Host "==> [4/8] Config lan dau" -ForegroundColor Cyan
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

Write-Host "==> [5/8] Admin mac dinh" -ForegroundColor Cyan
$App = Join-Path $Prefix "bin\$BinName"
& $App user add --admin --login $AdminLogin --name Admin --email admin@localhost `
  --password $AdminPass --config $cfgPath 2>$null
if ($LASTEXITCODE -ne 0) { Write-Host "    admin da ton tai - bo qua (idempotent)" }

# ------------------------------------------------------------------ Gitea
$GiteaExe   = Join-Path $Prefix "gitea\gitea.exe"
$GiteaAddr  = ""
$GiteaToken = ""
$GiteaOrg   = "openitms"
if (Test-Path $GiteaExe) {
  Write-Host "==> [6/8] Gitea (git server local, 127.0.0.1:3080)" -ForegroundColor Cyan
  $GDir = Join-Path $Prefix "gitea"
  $GConf = Join-Path $GDir "custom\conf\app.ini"
  New-Item -ItemType Directory -Force -Path "$GDir\custom\conf","$GDir\data","$GDir\repos","$GDir\log" | Out-Null

  @"
CREATE DATABASE IF NOT EXISTS gitea CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'gitea'@'localhost' IDENTIFIED BY '$GiteaDbPass';
CREATE USER IF NOT EXISTS 'gitea'@'127.0.0.1' IDENTIFIED BY '$GiteaDbPass';
ALTER USER 'gitea'@'localhost' IDENTIFIED BY '$GiteaDbPass';
ALTER USER 'gitea'@'127.0.0.1' IDENTIFIED BY '$GiteaDbPass';
GRANT ALL PRIVILEGES ON gitea.* TO 'gitea'@'localhost';
GRANT ALL PRIVILEGES ON gitea.* TO 'gitea'@'127.0.0.1';
FLUSH PRIVILEGES;
"@ | & $Mysql "-u" "root"

  if (-not (Test-Path $GConf)) {
    $secret = Rand-Key 40
    $internal = Rand-Key 40
    @"
APP_NAME = OpenITMS-SMB Git
RUN_MODE = prod
[server]
PROTOCOL = http
HTTP_ADDR = 127.0.0.1
HTTP_PORT = 3080
ROOT_URL = http://127.0.0.1:3080/
DISABLE_SSH = true
OFFLINE_MODE = true
[database]
DB_TYPE = mysql
HOST = 127.0.0.1:3306
NAME = gitea
USER = gitea
PASSWD = $GiteaDbPass
CHARSET = utf8mb4
[repository]
ROOT = $($GDir -replace '\\','/')/repos
[security]
INSTALL_LOCK = true
SECRET_KEY = $secret
INTERNAL_TOKEN = $internal
[service]
DISABLE_REGISTRATION = true
REQUIRE_SIGNIN_VIEW = true
[log]
ROOT_PATH = $($GDir -replace '\\','/')/log
LEVEL = Info
"@ | Set-Content -Path $GConf -Encoding ascii

    $env:GITEA_WORK_DIR = $GDir
    & $GiteaExe migrate --config $GConf | Out-Null
    & $GiteaExe admin user create --admin --username openitms-admin --password $GiteaAdminPass `
      --email admin@openitms.local --must-change-password=false --config $GConf 2>$null
    $GiteaToken = (& $GiteaExe admin user generate-access-token --username openitms-admin `
      --name openitms --scopes all --raw --config $GConf 2>$null | Select-Object -Last 1).Trim()
  } else {
    Write-Host "    Gitea da cau hinh - giu nguyen (idempotent)"
    # token da luu trong start.cmd cu (neu co) - doc lai o buoc [7]
  }
  $GiteaAddr = "127.0.0.1:3080"

  # Scheduled Task cho Gitea web
  $gAction = New-ScheduledTaskAction -Execute $GiteaExe `
    -Argument "web --config `"$GConf`" --work-path `"$GDir`"" -WorkingDirectory $GDir
  $gTrigger = New-ScheduledTaskTrigger -AtStartup
  $gPrincipal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
  $gSettings = New-ScheduledTaskSettingsSet -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -StartWhenAvailable
  Register-ScheduledTask -TaskName "OpenITMS-Gitea" -Action $gAction -Trigger $gTrigger `
    -Principal $gPrincipal -Settings $gSettings -Force | Out-Null
  Start-ScheduledTask -TaskName "OpenITMS-Gitea"
} else {
  Write-Host "==> [6/8] Gitea: khong co binary - bo qua (auto-repo se tat)" -ForegroundColor DarkGray
}

# ------------------------------------------------------------------ osquery (bundled, offline)
# osquery MSI di kem (vendor\osquery.msi) -> extract osqueryi.exe (msiexec /a, khong cai he thong)
# -> plugin device-inventory thu inventory KHONG can internet.
$OsqueryBin = ""
$OsqMsi = Join-Path $Prefix "vendor\osquery.msi"
if (Test-Path $OsqMsi) {
  Write-Host "==> osquery (bundled, offline inventory)" -ForegroundColor Cyan
  $osqDir = Join-Path $Prefix "osquery"
  New-Item -ItemType Directory -Force $osqDir | Out-Null
  $cand = Join-Path $osqDir "osquery\osqueryi.exe"
  if (-not (Test-Path $cand)) {
    Start-Process msiexec -ArgumentList '/a', "`"$OsqMsi`"", '/qn', "TARGETDIR=`"$osqDir`"" -Wait
  }
  if (Test-Path $cand) { $OsqueryBin = $cand }
}

Write-Host "==> [7/8] Wrapper start.cmd (env cho app + Gitea)" -ForegroundColor Cyan
$startCmd = Join-Path $cfgDir "start.cmd"
# Neu Gitea idempotent (khong sinh token moi) va da co start.cmd cu -> giu token cu.
if (-not $GiteaToken -and (Test-Path $startCmd)) {
  $old = Select-String -Path $startCmd -Pattern 'QUICKWIN_GITEA_TOKEN=(.+)' | Select-Object -First 1
  if ($old) { $GiteaToken = $old.Matches[0].Groups[1].Value.Trim() }
}
$lines = @(
  '@echo off',
  "set OPENITMS_PREFIX=$Prefix",
  "set QUICKWIN_CONFIG=$cfgPath",
  "set QUICKWIN_CERTS_DIR=$Prefix\certs",
  "set QUICKWIN_PLUGINS_DIR=$Prefix\plugins",
  'set NO_PROXY=127.0.0.1,localhost'
)
if ($OsqueryBin) {
  $lines += "set QUICKWIN_OSQUERY_BIN=$OsqueryBin"
}
if ($GiteaAddr -and $GiteaToken) {
  $lines += "set QUICKWIN_GITEA_ADDR=$GiteaAddr"
  $lines += "set QUICKWIN_GITEA_ORG=$GiteaOrg"
  $lines += "set QUICKWIN_GITEA_TOKEN=$GiteaToken"
}
$lines += "`"$App`" server --config `"$cfgPath`""
$lines -join "`r`n" | Set-Content -Path $startCmd -Encoding ascii

Write-Host "==> [8/8] Scheduled Task 'OpenITMS' (chay wrapper luc khoi dong)" -ForegroundColor Cyan
$action = New-ScheduledTaskAction -Execute "cmd.exe" -Argument "/c `"$startCmd`"" -WorkingDirectory $Prefix
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
if ($GiteaAddr) { Write-Host "  Git local: http://127.0.0.1:3080  (openitms-admin)" }
Write-Host "  Certs:     bo .pem vao $Prefix\certs la dung duoc ngay"
Write-Host "  Service:   Get-Service OpenITMS-DB ; Get-ScheduledTask OpenITMS*"
Write-Host "  Go cai:    installer\windows\uninstall.ps1"
Write-Host "===============================================================" -ForegroundColor Green
