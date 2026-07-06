---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [plugins/winrs-cert/]
---

# Spec: Plugin winrs-cert

Chạy lệnh xuống Windows qua **WinRS (WinRM) + certificate authentication** (không mật khẩu).
Cert lấy từ thư mục `./certs` của OpenITMS-SMB (certstore) — ném `.pem` vào là dùng.

## API động
| Route | Mô tả |
|---|---|
| `POST /api/plugins/winrs-cert/exec` (admin) | body `{host, port?=5986, cert, key?, command, timeout?=60}` → `{exit_code, stdout, stderr}` |
| `GET /api/plugins/winrs-cert/certs` | liệt kê cert khả dụng trong `./certs` (name, kind, has_key, cn) |

Cũng hỗ trợ `RunTask` (task runner core) với params `host/command/cert/key`.

## Chứng chỉ
- `.pem` chứa cả CERTIFICATE + PRIVATE KEY → dùng trực tiếp.
- cert/key tách → truyền thêm `key` (tên file .pem khác trong ./certs).
- `.pfx/.p12`: **v1 chưa hỗ trợ** (cần password decode) → báo lỗi rõ. TODO v2.
- Transport WinRM HTTPS, hiện `insecure=true` (bỏ verify server cert — lab/self-signed).
  TODO: option verify CA (đã chừa import tls).

## Phân loại lỗi (giúp người dùng biết sửa đâu)
`classifyError`: **CHỨNG CHỈ** (x509/tls/certificate) · **MẠNG** (refused/timeout/dial) ·
**XÁC THỰC** (401/403 — WinRM chưa bật cert auth / chưa map cert→user) · **WINRS** (khác).

## Yêu cầu phía Windows đích (để cert auth chạy)
1. WinRM listener HTTPS (port 5986).
2. Bật Certificate authentication: `Set-Item WSMan:\localhost\Service\Auth\Certificate $true`.
3. Map client cert → user local/JEA (`New-Item WSMan:\localhost\ClientCertificate ...`).
Template JEA/WinRS (P3-05) sẽ tự động hóa các bước này.

## Timeout (robustness)
`runWinRSCert` bọc lời gọi winrm trong goroutine + `select` trên `ctx.Done()` — handler
LUÔN trả về trong `timeout` request dù `masterzen/winrm` dial không tôn trọng context cho
host chết. Bug này (trước truyền `0` = treo vô hạn) phát hiện + fix qua E2E 2026-07-07.

## Test
- Unit (không cần Windows): metadata↔manifest routes; classifyError 4 nhóm; resolveCertKey
  (combo.pem OK, pfx từ chối); handleExec validation (400/404); handleListCerts.
- **E2E qua core (2026-07-07)**: certstore nạp cert thật (`cn`, `has_key`); exec tới host chết
  → lỗi phân loại "LỖI MẠNG: timeout" trong ~timeout giây, plugin sống sót. PASS.
- **Cert-auth THẬT (exec thành công) cần WinRM HTTPS + cert-auth target** — cần admin cấu hình
  (tests/e2e/winrs/README.md). Chưa chạy: máy dev không có admin. AC cuối P1-09.
