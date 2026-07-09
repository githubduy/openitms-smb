---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0023-enroll-localaccount-token.patch, core-patches/0022-enroll-clientcert-negotiation.patch]
---

# Spec patch 0023 — Enroll: LocalAccountTokenFilterPolicy (fix WinRS AccessDenied)

## Mục đích (WHY)
Debug THẬT (tiếp theo 0022): sau khi bật client-cert negotiation, cert-auth chạy (hết 401), nhưng
WinRS trả `500` với `WSManFault AccessDenied (Code 5)`. Nguyên nhân: tài khoản **local** (openitms)
khi đăng nhập **qua mạng** (WinRM) bị UAC cấp **token đã lọc (standard-user)** dù thuộc Administrators
→ không đủ quyền tạo WinRM shell → Access Denied. (UAC remote restriction cho local account.)

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa | Sau khi thêm user vào Administrators: set `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System\LocalAccountTokenFilterPolicy = 1` (DWORD) → local-account remote logon nhận **full token** (không bị lọc) |

## Chuỗi lỗi WinRS cert-auth — TỔNG HỢP 4 tầng (tất cả đã fix)
| # | Lỗi | Triệu chứng | Fix |
|---|-----|-------------|-----|
| 1 | Proxy | "cannot connect" | winrs-exec/transport.go `Proxy: nil` |
| 2 | TLS 1.3 | "connection forcibly closed" | transport.go `MaxVersion: TLS12` |
| 3 | Client-cert negotiation | `401` | winrs-enroll.ps1 `clientcertnegotiation=enable` (0022) |
| 4 | UAC token filter | `AccessDenied Code 5` | winrs-enroll.ps1 `LocalAccountTokenFilterPolicy=1` (0023) |

## Fix máy đang chạy (không re-enroll)
```powershell
New-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name LocalAccountTokenFilterPolicy -Value 1 -PropertyType DWord -Force
```
Có hiệu lực ngay (không cần reboot).

## Verify
- Registry trước: `LocalAccountTokenFilterPolicy` rỗng (default → lọc token) → AccessDenied.
- `go build ./api/...` OK; PS 5.1 parse OK; chain 0001–0023 apply+build.
- Sau fix: local admin remote logon full token → WinRM shell OK → lệnh chạy (chờ user xác nhận).

## Liên quan
- Client-cert negotiation: [[0022-enroll-clientcert-negotiation]].
