#requires -RunAsAdministrator
<#
  jea-winrs-setup — chuẩn hóa Windows 11 để OpenITMS-SMB gọi lệnh qua WinRS + certificate auth.
  Chạy TRÊN máy Windows đích (qua bootstrap ban đầu hoặc thủ công lần đầu).
  Params inject bởi OpenITMS-SMB task runner từ template.yaml inputs.
#>
param(
  [Parameter(Mandatory)] [string]$TargetHost,
  [Parameter(Mandatory)] [string]$ClientCertThumbprint,
  [string]$MappedUser = "openitms-runner"
)
$ErrorActionPreference = "Stop"

Write-Host "==> [1/4] WinRM HTTPS listener (self-signed cho lab; CA-signed cho production)"
$srvCert = New-SelfSignedCertificate -DnsName $env:COMPUTERNAME -CertStoreLocation Cert:\LocalMachine\My
New-Item -Path WSMan:\localhost\Listener -Transport HTTPS -Address * `
  -CertificateThumbPrint $srvCert.Thumbprint -Force | Out-Null
New-NetFirewallRule -DisplayName "OpenITMS WinRM HTTPS" -Direction Inbound `
  -LocalPort 5986 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue | Out-Null

Write-Host "==> [2/4] Bật Certificate authentication"
Set-Item WSMan:\localhost\Service\Auth\Certificate $true

Write-Host "==> [3/4] Map client cert → user $MappedUser (JEA)"
if (-not (Get-LocalUser -Name $MappedUser -ErrorAction SilentlyContinue)) {
  $pw = ConvertTo-SecureString ([guid]::NewGuid().ToString()) -AsPlainText -Force
  New-LocalUser -Name $MappedUser -Password $pw -PasswordNeverExpires -AccountNeverExpires | Out-Null
}
$cred = New-Object System.Management.Automation.PSCredential($MappedUser,
  (ConvertTo-SecureString ([guid]::NewGuid().ToString()) -AsPlainText -Force))
$issuer = (Get-Item "Cert:\LocalMachine\My\$ClientCertThumbprint" -ErrorAction SilentlyContinue).Issuer
if (-not $issuer) { throw "Không tìm thấy client cert thumbprint $ClientCertThumbprint trên máy này (import trước)." }
New-Item -Path WSMan:\localhost\ClientCertificate `
  -Subject "*" -URI * -Issuer $issuer -Credential $cred -Force | Out-Null

Write-Host "==> [4/4] Xong. Test từ OpenITMS: POST /api/plugins/winrs-cert/exec"
Write-Host "    { host: '$TargetHost', cert: '<file .pem>', command: 'hostname' }"
