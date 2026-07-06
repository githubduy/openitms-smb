---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [plugin-manager/certstore/]
---

# Spec: Thư mục certificates `./certs` (certstore)

Package `plugin-manager/certstore` (`quickwin.dev/pluginmanager/certstore`).
Người dùng **ném file `.pem/.pfx` vào thư mục certs là dùng được ngay** — nạp nóng,
không restart (giá trị cốt lõi zero-config).

## Hành vi
- Watcher **polling** (mặc định 5s, không dep fsnotify — chạy mọi OS/filesystem/NFS).
- Nhận: `.pem .crt .key` (parse x509 + phát hiện PRIVATE KEY) · `.pfx .p12` (giữ raw —
  PKCS#12 thường có password, consumer decode lúc dùng).
- File đổi nội dung (SHA256 khác) → nạp lại + gọi callback `OnLoad`.
- File bị xóa → gỡ entry khỏi store.
- **File rác không làm crash** — bỏ qua + log warning.
- Thư mục chưa tồn tại lúc Start → vẫn watch, tạo sau vẫn nhận.

## API
`New(dir, interval, logger)` → `Start() / Stop() / Get(name) / List() / OnLoad(fn)`.
Consumer: plugin `winrs-cert` (P1-09) + inventory WinRM; core wire ở patch 0003+ khi
Settings UI cho đổi đường dẫn certs.

## Test (PASS 2026-07-06, Windows)
Hot-load PEM ≤ interval; PFX giữ raw; xóa file → gỡ; rác bị bỏ qua; dir tạo muộn vẫn nhận.
