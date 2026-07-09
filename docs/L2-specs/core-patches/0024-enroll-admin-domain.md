---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0024-enroll-admin-domain.patch, core-patches/0023-enroll-localaccount-token.patch]
---

# Spec patch 0024 — Enroll: add-to-Administrators tin cậy trên máy domain (fix AccessDenied #2)

## Mục đích (WHY)
Debug thật (tiếp 0023): dù `LocalAccountTokenFilterPolicy=1`, WinRS vẫn `AccessDenied`. Kiểm tra
`Get-LocalGroupMember Administrators` → **user `openitms` KHÔNG có trong Administrators**. Máy
**domain-joined**: `Add-LocalGroupMember -Member "openitms"` resolve tên sang domain
principal → fail; `-ErrorAction SilentlyContinue` **nuốt lỗi** → user bị bỏ ở mức standard → WinRS
tạo shell = AccessDenied.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa | Thêm vào Administrators bằng `net localgroup Administrators <user> /add` (tin cậy cho local account trên máy domain) + **verify** membership; nếu chưa vào → fallback `Add-LocalGroupMember -Member "$env:COMPUTERNAME\<user>"` (tên máy-qualified, không nhầm domain) |

## Chuỗi lỗi WinRS — 5 tầng (tất cả đã fix qua debug thật)
1. Proxy → `Proxy: nil` (winrs-exec)
2. TLS 1.3 → `MaxVersion: TLS12` (winrs-exec)
3. Client-cert negotiation → `clientcertnegotiation=enable` (0022)
4. UAC token filter → `LocalAccountTokenFilterPolicy=1` (0023)
5. **User không vào Administrators (domain)** → `net localgroup` (0024)

## Fix máy đang chạy
```powershell
net localgroup Administrators openitms /add
```

## Verify
- `Get-LocalGroupMember Administrators` trước fix: không có openitms → AccessDenied.
- PS 5.1 parse OK; `go build ./api/...` OK; chain 0001–0024 build.
- Sau `net localgroup add`: openitms là admin + token filter=1 → WinRM shell OK (chờ user xác nhận).

## Liên quan
- Token filter: [[0023-enroll-localaccount-token]].
