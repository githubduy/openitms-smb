---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0010-winrs-console.patch, winrs-exec/, core-patches/0009-winrs-app.patch]
---

# Spec patch 0010 — WinRS Console (gõ lệnh nhanh xuống 1 host Windows)

## Mục đích (WHY)
Bổ sung cho engine 0009: nhiều thao tác IT chỉ cần "chạy 1 lệnh xuống 1 máy Windows và xem
kết quả ngay" (kiểm tra service, đọc event log, ping cấu hình…) — không đáng để tạo template +
inventory + chạy task có lịch sử. Console cung cấp đường đồng bộ: nhập host/cert/lệnh → chạy →
xem stdout/stderr/exit_code.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_winrs.go` | MỚI | `GetWinRSCerts` (liệt kê `*.pem`/`*.crt` trong certs dir); `WinRSExec` (bind `{host,port,cert,key,command}`, đọc cert theo tên file, chạy `winrsexec.Run` với `-EncodedCommand`, timeout 90s ctx / 60s dial, trả `{ok,exit_code,stdout,stderr}` hoặc `{ok:false,error}` đã Classify); helper `encodePowerShellConsole` (base64 UTF-16LE) |
| `api/router.go` | +2 route | `GET /winrs/certs`, `POST /winrs/exec` trên `projectUserAPI` |
| `web/src/views/project/WinRSConsole.vue` | MỚI | Form host/port/cert (autocomplete từ /winrs/certs) + editor lệnh (codemirror) + nút Chạy + hiển thị exit chip / stdout / stderr |
| `web/src/router/index.js` | +import +route | `/project/:projectId/winrs_console` |
| `web/src/App.vue` | +nav item | "WinRS Console" (icon windows) sau Inventory trong sidebar project |

## Bảo mật / gate
- Route nằm trên `projectUserAPI` → qua `ProjectMiddleware` + `GetMustCanMiddleware(CanManageProjectResources)`:
  chỉ thành viên project có quyền quản trị tài nguyên mới gọi được (thừa hưởng authn/authz core).
- **Chặn path-traversal**: `cert`/`key` không được chứa `/` hoặc `\` → chỉ đọc file trong certs dir
  (`QUICKWIN_CERTS_DIR`, mặc định `certs`). Không nhận đường dẫn tuyệt đối.
- Đồng bộ, **không lưu lịch sử** (khác Task Template app WinRS của 0009 — cái đó có history/schedule).

## Luồng
1. UI load `GET /winrs/certs` → điền dropdown cert.
2. User nhập host/port/lệnh, chọn cert, bấm Chạy → `POST /winrs/exec`.
3. Backend đọc cert → `winrsexec.Run` (WinRM HTTPS cert-auth) → trả kết quả JSON.
4. UI hiển thị exit code (chip xanh/đỏ) + stdout/stderr.

## Verify
- `go build ./api/...` OK (handler compile, winrsexec resolve từ go.mod 0009).
- eslint sạch (WinRSConsole.vue, router/index.js, App.vue).
- Chain 0001–0010 apply sạch + build (xác nhận ở commit).
- Đường winrm thật cần lab Windows + cert (E2E riêng — xem 0009 spec "còn lại").

## Liên quan
- Engine + inventory/app type: [[0009-winrs-app]].
