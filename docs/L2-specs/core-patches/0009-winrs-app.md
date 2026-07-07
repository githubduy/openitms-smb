---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0009-winrs-app.patch, winrs-exec/, plugins/winrs-cert/]
---

# Spec patch 0009 — WinRS execution engine (inventory + app "winrs")

## Mục đích (WHY)
OpenITMS thực thi trực tiếp xuống Windows host qua WinRM (mô hình ADR: pwsh → SSH → winrs),
không cần Ansible. Semaphore gốc chỉ có inventory/app hướng-Ansible (static/file) + Terraform,
nên thiếu đường chạy pwsh thẳng xuống endpoint Windows. Patch thêm 1 inventory type + 1 app type
`winrs` làm engine execution: task/schedule/history của Semaphore chạy pwsh xuống nhiều host
qua WinRM certificate-auth.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `db/Inventory.go` | +1 dòng | `InventoryWinRS InventoryType = "winrs"` (host + cert) |
| `db/Template.go` | +3 dòng | `AppWinRS TemplateApp = "winrs"`; `InventoryTypes()` +case `AppWinRS → [InventoryWinRS]` |
| `db_lib/quickwin_winrs_app.go` | MỚI | `WinRSApp` (implements LocalApp): đọc script pwsh từ repo (`Template.Playbook`), parse host từ `Inventory.Inventory` qua `winrsexec.ParseHosts`, mã hoá script `-EncodedCommand` (base64 UTF-16LE), chạy mỗi host qua `winrsexec.Run` (WinRM cert-auth, ctx 5 phút + dial 60s), stream stdout/stderr vào Logger, fail nếu bất kỳ host lỗi |
| `db_lib/AppFactory.go` | +7 dòng | `case db.AppWinRS → &WinRSApp{Template, Repository, Inventory, Logger}` |
| `go.mod` | +require/replace | `quickwin.dev/winrsexec` (replace `../winrs-exec`) |
| `web/src/lib/constants.js` | +9 dòng | `winrs` vào APP_ICONS/APP_SHORT_TITLE/APP_TITLE/APP_INVENTORY_TITLE/APP_INVENTORY_TYPES |
| `web/src/views/project/Inventory.vue` | +5 dòng | `apps: ['ansible','winrs']` (dropdown NEW INVENTORY); `getAppByType('winrs')`; truyền `:app` cho form |
| `web/src/components/InventoryForm.vue` | ~50 dòng | prop `app`; `inventoryTypes` computed theo app; `getNewItem` mặc định type; editor host (codemirror plain) cho type `winrs`; ẩn ssh/sudo key khi `winrs` (dùng cert) |

## Package dùng chung `winrs-exec/` (module `quickwin.dev/winrsexec`)
Tách logic WinRM ra khỏi plugin `winrs-cert` để cả plugin lẫn app dùng chung:
- `Run(ctx, Params{Host,Port,CertPEM,KeyPEM,Command,Timeout}) (*Result{ExitCode,Stdout,Stderr}, error)`
  — WinRM HTTPS cert-auth (`ClientAuthRequest` transport), port mặc định 5986, goroutine+select
  trên `ctx.Done()` để không treo.
- `Classify(err) string` — phân loại: LỖI CHỨNG CHỈ / LỖI MẠNG / LỖI XÁC THỰC / LỖI WINRS.
- `ParseHosts(inventory, defaultCert) []Host` — mỗi dòng `host[:port] cert=<file.pem> [key=<file.key>]`;
  bỏ dòng trống / `#`.
Plugin `winrs-cert` đã refactor để delegate sang package này (không đổi hành vi).

## Cấu hình (env — runner đặt cert)
- `QUICKWIN_CERTS_DIR` (mặc định `certs`) — thư mục chứa file cert/key.
- `QUICKWIN_WINRS_DEFAULT_CERT` — cert mặc định cho host không khai `cert=`.

## Luồng thực thi (backend)
1. `LocalJob.installInventory()` — type `winrs` không khớp case nào → no-op (không ghi inventory file).
2. `getShellArgs` (nhánh default, không phải Ansible) — build args; WinRSApp bỏ qua CliArgs.
3. `App.Run` → `WinRSApp.Run`: đọc script + parse host + đọc cert theo host → `winrsexec.Run` từng host.
`ValidateInventory` chấp nhận `ssh_key_id = null` (winrs dùng cert, không cần SSH key).

## Định dạng inventory WinRS (ví dụ)
```
# host[:port] cert=<file.pem> [key=<file.key>]
10.0.0.5 cert=win11.pem
win-app.lab:5986 cert=app.pem key=app.key
```

## Verify
- `winrs-exec` unit test: `TestParseHosts` (4 host, port/cert/key), `TestClassify` — PASS.
- `plugins/winrs-cert` build + test PASS sau refactor.
- `scripts/build-all.sh`: chain 0001–0009 apply + build OK (binary `quickwin-dev-sem2.18.16`).
- `scripts/run-tests.sh` (đã thêm `winrs-exec` vào module list) — toàn bộ PASS.
- UI: eslint sạch (constants.js, InventoryForm.vue, Inventory.vue).

## Còn lại (ngoài patch này)
- Endpoint console: giao diện gõ lệnh winrs nhanh tới 1 host (user chọn "Cả hai" — làm sau).
- E2E thật xuống Windows host có WinRM+cert (unit đã cover parse/classify; đường winrm cần lab).
- G-04: seed project "host" mặc định. G-05: UI xem repo local.
