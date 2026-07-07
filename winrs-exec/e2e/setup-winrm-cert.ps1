# setup-winrm-cert.ps1 — CHUẨN BỊ 1 host Windows cho E2E WinRS (WinRM HTTPS + certificate auth).
#
# ⚠️  Script này THAY ĐỔI CẤU HÌNH BẢO MẬT hệ thống (WinRM listener HTTPS, bật Certificate auth,
#     map client cert vào 1 local user). CHỈ chạy trên máy LAB/test mà bạn toàn quyền, trong
#     PowerShell chạy AS ADMINISTRATOR. Đọc kỹ trước khi chạy. KHÔNG chạy trên máy production.
#
# Nó tạo:
#   - 1 local user (mặc định 'winrsuser') để cert map vào.
#   - 1 self-signed CA-ish client cert (UPN = winrsuser@localhost) → export cert+key ra PEM.
#   - 1 server cert cho HTTPS listener.
#   - WinRM HTTPS listener (5986) + bật Basic:$false, Certificate:$true.
#   - Cert mapping (client cert → local user).
#
# Sau khi chạy xong, dùng file PEM xuất ra với winrsexec (WINRS_E2E_CERT) để test:
#   go test ./winrs-exec/ -run TestE2E -v -count=1

[CmdletBinding()]
param(
  [string]$User = "winrsuser",
  [string]$Password = "WinRS-E2E-Pass!2026",
  [string]$OutDir = "$PSScriptRoot\..\..\run-local\certs",
  [string]$PemName = "host.pem"
)

$ErrorActionPreference = "Stop"

function Assert-Admin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  $p = New-Object Security.Principal.WindowsPrincipal($id)
  if (-not $p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Phải chạy PowerShell AS ADMINISTRATOR."
  }
}

Assert-Admin
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

Write-Host "==> 1. Tạo local user '$User'" -ForegroundColor Cyan
$secure = ConvertTo-SecureString $Password -AsPlainText -Force
if (-not (Get-LocalUser -Name $User -ErrorAction SilentlyContinue)) {
  New-LocalUser -Name $User -Password $secure -FullName "WinRS E2E" -Description "WinRS cert-auth E2E" | Out-Null
  Add-LocalGroupMember -Group "Administrators" -Member $User -ErrorAction SilentlyContinue
} else {
  Set-LocalUser -Name $User -Password $secure
}

Write-Host "==> 2. Tạo client cert (UPN=$User@localhost)" -ForegroundColor Cyan
$clientCert = New-SelfSignedCertificate `
  -Type Custom -KeyUsage DigitalSignature -KeySpec Signature `
  -Subject "CN=$User" -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.2", "2.5.29.17={text}upn=$User@localhost") `
  -CertStoreLocation "Cert:\CurrentUser\My"
Write-Host "    client thumbprint: $($clientCert.Thumbprint)"

Write-Host "==> 3. Tạo server cert (HTTPS listener)" -ForegroundColor Cyan
$serverCert = New-SelfSignedCertificate -DnsName $env:COMPUTERNAME, "localhost", "127.0.0.1" `
  -CertStoreLocation "Cert:\LocalMachine\My"

Write-Host "==> 4. Trust client cert (Root + TrustedPeople LocalMachine)" -ForegroundColor Cyan
$tmp = Join-Path $env:TEMP "winrs-client.cer"
Export-Certificate -Cert $clientCert -FilePath $tmp | Out-Null
Import-Certificate -FilePath $tmp -CertStoreLocation "Cert:\LocalMachine\Root" | Out-Null
Import-Certificate -FilePath $tmp -CertStoreLocation "Cert:\LocalMachine\TrustedPeople" | Out-Null

Write-Host "==> 5. WinRM HTTPS listener (5986)" -ForegroundColor Cyan
if (-not (Get-Service WinRM).Status -eq 'Running') { Start-Service WinRM }
Remove-WSManInstance -ResourceURI winrm/config/Listener -SelectorSet @{Address="*";Transport="HTTPS"} -ErrorAction SilentlyContinue
New-WSManInstance -ResourceURI winrm/config/Listener -SelectorSet @{Address="*";Transport="HTTPS"} `
  -ValueSet @{Hostname=$env:COMPUTERNAME;CertificateThumbprint=$serverCert.Thumbprint} | Out-Null

Write-Host "==> 6. Bật Certificate auth" -ForegroundColor Cyan
Set-Item -Path WSMan:\localhost\Service\Auth\Certificate -Value $true
New-NetFirewallRule -DisplayName "WinRM HTTPS 5986 (E2E)" -Direction Inbound -LocalPort 5986 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue | Out-Null

Write-Host "==> 7. Map client cert -> local user" -ForegroundColor Cyan
$issuerThumb = $clientCert.Thumbprint
New-Item -Path WSMan:\localhost\ClientCertificate `
  -Subject "$User@localhost" -URI * -Issuer $issuerThumb `
  -Credential (New-Object System.Management.Automation.PSCredential($User, $secure)) -Force | Out-Null

Write-Host "==> 8. Export cert+key ra PEM: $OutDir\$PemName" -ForegroundColor Cyan
# Export PFX rồi convert sang PEM (cert + key) qua .NET.
$pfxPass = ConvertTo-SecureString "pfx" -AsPlainText -Force
$pfxPath = Join-Path $env:TEMP "winrs-client.pfx"
Export-PfxCertificate -Cert $clientCert -FilePath $pfxPath -Password $pfxPass | Out-Null
$col = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2Collection
$col.Import($pfxPath, "pfx", "Exportable")
$c = $col[0]
$certB64 = [Convert]::ToBase64String($c.RawData, 'InsertLineBreaks')
$keyBytes = $c.GetRSAPrivateKey().ExportPkcs8PrivateKey()
$keyB64 = [Convert]::ToBase64String($keyBytes, 'InsertLineBreaks')
$pem = "-----BEGIN CERTIFICATE-----`n$certB64`n-----END CERTIFICATE-----`n-----BEGIN PRIVATE KEY-----`n$keyB64`n-----END PRIVATE KEY-----`n"
Set-Content -Path (Join-Path $OutDir $PemName) -Value $pem -Encoding ascii

Write-Host ""
Write-Host "HOÀN TẤT." -ForegroundColor Green
Write-Host "PEM (cert+key): $OutDir\$PemName"
Write-Host ""
Write-Host "Chạy E2E test:" -ForegroundColor Yellow
Write-Host "  `$env:WINRS_E2E_HOST = '127.0.0.1'"
Write-Host "  `$env:WINRS_E2E_CERT = '$OutDir\$PemName'"
Write-Host "  go test ./winrs-exec/ -run TestE2E -v -count=1"
Write-Host ""
Write-Host "GỠ sau khi test (dọn cấu hình):" -ForegroundColor Yellow
Write-Host "  Remove-WSManInstance winrm/config/Listener -SelectorSet @{Address='*';Transport='HTTPS'}"
Write-Host "  Set-Item WSMan:\localhost\Service\Auth\Certificate -Value `$false"
Write-Host "  Remove-LocalUser $User ; Remove-NetFirewallRule -DisplayName 'WinRM HTTPS 5986 (E2E)'"
