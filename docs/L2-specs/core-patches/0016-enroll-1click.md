---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0016-enroll-1click.patch, core-patches/0014-endpoint-enroll-script.patch, core-patches/0009-winrs-app.patch]
---

# Spec patch 0016 — Enrollment 1-click (tự nạp cert từ máy đích)

## Mục đích (WHY)
0014 cho tải script chuẩn bị endpoint nhưng vẫn phải **copy cert thủ công** lên server + tự thêm
host vào inventory. 0016 khép kín thành **1-click**: máy đích chạy script → script tự nạp cert lên
OpenITMS + tự thêm host vào inventory WinRS. Giảm ma sát cho người dùng SMB.

## Kiến trúc token (stateless)
- Token = `base64url("projectID:exp") + "." + base64url(HMAC-SHA256(msg, key))`.
- `key` = `base64decode(util.Config.CookieHash)` (secret sẵn có của server) → **không cần lưu token**.
- TTL 30 phút. Verify: so HMAC (hmac.Equal, chống timing) + check hạn → trả projectID.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_enroll_token.go` | MỚI | `signEnrollToken`/`verifyEnrollToken`; `GetEnrollToken` (authed) trả token+expiry; `EnrollWinRS` (public) verify token → lưu cert vào certs dir (`safeCertName` lọc traversal, `..`→`.`, bỏ ký tự đầu; giới hạn 64KB; check `-----BEGIN`) → thêm `host cert=<name>.pem` vào inventory WinRS (tạo "WinRS Endpoints" nếu chưa có, idempotent theo host); `enrollScriptInject` |
| `api/projects/quickwin_enroll.go` | sửa (0014) | `GetEnrollScript` chèn `?server=&token=` vào script qua placeholder |
| `api/projects/scripts/winrs-enroll.ps1` | sửa (0014) | `$OpenITMSServer`/`$EnrollToken` (placeholder); block cuối tự `Invoke-RestMethod POST /api/enroll/winrs` (tolerate self-signed cho LAN), ẩn hướng dẫn thủ công khi đã auto |
| `api/router.go` | +2 route | `GET /project/{id}/winrs/enroll-token` (projectUserAPI, authed); `POST /enroll/winrs` (publicAPIRouter, auth bằng token) |
| `web/src/views/project/WinRSConsole.vue` | +nút | "Enroll 1-click" → lấy token → tải script nhúng token+`window.location.origin` |
| `web/src/lang/en.js` | +key | `enroll1click*`, `enrollGenerating`, `enrollError`, `enrollManualNote` |

## Bảo mật
- Endpoint upload **public nhưng auth bằng token HMAC** (không giả mạo được nếu không có CookieHash).
- Token ngắn hạn (30'), scope 1 project; kẻ có token chỉ thêm được cert/host vào project đó.
- `safeCertName`: chỉ `[A-Za-z0-9._-]`, triệt `..`, bỏ ký tự đầu → **không path-traversal** (đã test `../../evil` → `evil.pem`).
- Cert giới hạn 64KB + phải chứa `-----BEGIN`. Ghi file mode 0600.
- Script tolerate self-signed **chỉ cho lần POST cert** (LAN enrollment); revert callback sau đó.

## Luồng 1-click
1. UI WinRS Console → "Enroll 1-click" → `GET /winrs/enroll-token` → token.
2. Tải `enroll-script/winrs?token=…&server=<origin>` → script nhúng sẵn.
3. Chạy AS ADMIN trên máy Windows → bật WinRM+cert → **tự POST cert** lên `/enroll/winrs`.
4. Server lưu cert + thêm host vào inventory WinRS → endpoint sẵn sàng, không thao tác tay.

## Verify (E2E app thật)
- `GET /winrs/enroll-token` (authed) → token; script tải về có `$OpenITMSServer`/`$EnrollToken` đã chèn (0 placeholder sót).
- `POST /enroll/winrs` token hợp lệ + PEM giả → `{ok:true, cert_file}`; token giả → 401; `name=../../evil` → `evil.pem` (không escape).
- Inventory "WinRS Endpoints" tự tạo + thêm cả 2 host; cert files nằm trong certs dir.
- Chain 0001–0016 apply + build.

## Liên quan
- Script + tải: [[0014-endpoint-enroll-script]]. Engine + inventory WinRS: [[0009-winrs-app]].
