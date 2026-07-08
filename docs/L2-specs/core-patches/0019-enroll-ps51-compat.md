---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-08
related-code: [core-patches/0019-enroll-ps51-compat.patch, core-patches/0014-endpoint-enroll-script.patch]
---

# Spec patch 0019 — Enroll scripts tương thích PowerShell 5.1

## Mục đích (WHY)
Người dùng chạy `openitms-winrs-enroll.ps1` bằng **PowerShell 7 (pwsh 7.6)** → `New-LocalUser` lỗi
`Could not load type 'Microsoft.PowerShell.Telemetry.Internal.TelemetryAPI'`. Nguyên nhân: module
`Microsoft.PowerShell.LocalAccounts` (và một số cmdlet cert/WSMan) **không tương thích PowerShell 7**;
chỉ chạy ổn trên **Windows PowerShell 5.1**. Thêm nữa, chuỗi tiếng Việt (UTF-8 không BOM) làm PS 5.1
đọc theo ANSI → parse lỗi (đã gặp ở uninstall.ps1).

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa | (1) Guard đầu script: nếu `$PSVersionTable.PSVersion.Major -ge 6` → relaunch `powershell.exe` (WinPS 5.1) `-File $PSCommandPath @PSBoundParameters` rồi `exit`; (2) toàn bộ chuỗi/comment → ASCII/English (bỏ dấu + ký tự non-ASCII). Giữ nguyên logic + placeholder `@@OPENITMS_SERVER@@`/`@@ENROLL_TOKEN@@` (1-click) |
| `api/projects/scripts/ssh-enroll.ps1` | sửa | Cùng guard relaunch + ASCII/English |

## Vì sao relaunch (thay vì thay cmdlet)
Script dùng nhiều module hướng Windows-PowerShell: `LocalAccounts` (New/Get/Set-LocalUser,
Add-LocalGroupMember), `PKI` (New-SelfSignedCertificate, Export/Import-Certificate, Export-Pfx),
`WSMan` provider + `New-WSManInstance`, `NetSecurity`, `DISM` (Add-WindowsCapability). Chạy toàn bộ
dưới WinPS 5.1 (có sẵn trên mọi Windows 10/11) là cách chắc chắn nhất, thay vì thay từng cmdlet.

## Vì sao ASCII
Sau relaunch, script chạy dưới PS 5.1 — vốn đọc file `.ps1` UTF-8-không-BOM theo ANSI → ký tự tiếng
Việt/emoji trong **chuỗi** làm hỏng terminator → parse fail. ASCII-only đảm bảo parse sạch ở cả
PS 5.1 lẫn 7 (đã verify AST parser 5.1).

## Verify
- Parser PS 5.1: `winrs-enroll.ps1` + `ssh-enroll.ps1` **PARSE OK** (không lỗi).
- `go build ./api/...` OK (go:embed nội dung mới).
- Workaround trước fix: `powershell.exe -ExecutionPolicy Bypass -File .\openitms-winrs-enroll.ps1`.
- Chain 0001–0019 apply + build.

## Liên quan
- Script gốc + tải: [[0014-endpoint-enroll-script]].
