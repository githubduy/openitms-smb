---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-08
related-code: [core-patches/0022-enroll-clientcert-negotiation.patch, core-patches/0014-endpoint-enroll-script.patch, winrs-exec/transport.go]
---

# Spec patch 0022 — Enroll: bật client-cert negotiation + dọn cert cũ (fix 401)

## Mục đích (WHY)
Debug THẬT trên máy Windows: sau khi fix TLS 1.3→1.2 (winrs-exec), WinRS vẫn `401`. Chẩn đoán bằng
Go TLS callback + `netsh http show sslcert`:
- Binding `0.0.0.0:5986`: **`Negotiate Client Certificate : Disabled`** → server **KHÔNG yêu cầu**
  client cert trong TLS handshake → client không gửi cert → WinRM không có cert để map → **401**.
- `New-WSManInstance` tạo listener nhưng KHÔNG bật cờ này (lỗi kinh điển WinRM cert-auth).
- Chạy enroll nhiều lần để lại nhiều cert `CN=openitms` ở Root → nhiễu.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | sửa | (1) Sau tạo listener + mapping: `netsh http update sslcert ipport=0.0.0.0:5986 clientcertnegotiation=enable` (fallback: parse appid từ `show sslcert` → `delete` + `add` với `clientcertnegotiation=enable`). (2) Trước tạo cert: dọn cert `CN=$User` ở CurrentUser\My + LocalMachine\{My,Root,TrustedPeople} + WSMan client cert mapping cũ (match `$User@localhost`) → idempotent, hết nhiễu |

## Chuỗi lỗi WinRS cert-auth (tổng hợp — 3 tầng, đều đã fix)
1. **Proxy** (winrs-exec/transport.go): `HTTPS_PROXY` route WinRM qua proxy → "cannot connect". Fix: `Proxy: nil`.
2. **TLS 1.3** (winrs-exec/transport.go): WinRM cert-auth không hỗ trợ TLS 1.3 → "forcibly closed". Fix: `MaxVersion: TLS12`.
3. **Client-cert negotiation** (0022): binding không hỏi client cert → 401. Fix: `clientcertnegotiation=enable`.

## Fix máy đang chạy (không cần re-enroll)
```powershell
netsh http update sslcert ipport=0.0.0.0:5986 clientcertnegotiation=enable
```

## Verify
- `netsh http show sslcert ipport=0.0.0.0:5986` → `Negotiate Client Certificate : Disabled` (trước fix).
- Go TLS callback: server KHÔNG gọi `GetClientCertificate` → không request cert (bằng chứng root cause).
- Parser PS 5.1: winrs-enroll.ps1 PARSE OK; `go build ./api/...` OK.
- Sau fix (enable negotiation): server request cert → client gửi → map → auth OK (chờ user xác nhận trên máy thật).

## Liên quan
- Script gốc: [[0014-endpoint-enroll-script]]. Relaunch/PKCS#1: [[0019-enroll-ps51-compat]], [[0020-enroll-pem-pkcs1]].
