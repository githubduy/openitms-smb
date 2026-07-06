# E2E winrs-cert — cần Windows 11 lab (chưa tự động hóa trong CI)

Plugin `winrs-cert` gọi lệnh thật xuống Windows qua WinRM cert-auth. Unit test đã phủ
logic (không cần Windows); phần này verify **đường đi cert thật** — cần 1 máy/VM Win11.

## Chuẩn bị máy Win11 (chạy PowerShell admin)
```powershell
# 1. WinRM HTTPS listener + self-signed cho lab
$c = New-SelfSignedCertificate -DnsName $env:COMPUTERNAME -CertStoreLocation Cert:\LocalMachine\My
New-Item -Path WSMan:\localhost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $c.Thumbprint -Force
New-NetFirewallRule -DisplayName "WinRM HTTPS" -Direction Inbound -LocalPort 5986 -Protocol TCP -Action Allow

# 2. Bật Certificate auth
Set-Item WSMan:\localhost\Service\Auth\Certificate $true

# 3. Tạo client cert (trên máy OpenITMS) + map → user local
#    (xem template JEA/WinRS P3-05 để tự động hóa map cert→user)
```

## Chạy từ OpenITMS
1. Copy client cert `.pem` (cert+key) vào `./certs/win11-lab.pem`.
2. `POST /api/plugins/winrs-cert/exec`
   `{"host":"<ip-win11>","cert":"win11-lab.pem","command":"hostname"}`
3. Kỳ vọng: `exit_code=0`, `stdout` = hostname máy Win11.

## Điều kiện đóng AC P1-09
Chạy `hostname` qua API trả đúng + cert sai phân loại đúng nhóm lỗi.
Khi có lab: chuyển nội dung này thành `run.ps1` chạy được + đưa vào CI self-hosted Windows runner.
