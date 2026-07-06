---
level: L5
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: []
---

# E2E trên hạ tầng thật (2026-07-07)

Ngoài CI Ubuntu, đã chạy E2E trên hardware thật để verify artifact + phát hiện bug thực tế.

## Linux — CentOS 7 (glibc, kernel 3.10) — ✅ PASS
- Cross-compile `GOOS=linux GOARCH=amd64 CGO_ENABLED=0` → binary **static** (chạy được glibc cũ).
- Upload core + plugin hello, chạy smoke (bolt, non-root, non-systemd — non-invasive trên server chung):
  - `./semaphore version` OK; tạo admin; `/api/ping` → pong; login OK.
  - **`GET /api/plugins` → hello RUNNING** (go-plugin gRPC launch trên Linux thật).
- Dọn sạch sau test.
- ⏳ **Chưa test**: full `install.sh` (systemd + MariaDB) — server chung không nên cài invasive;
  MariaDB binary không tải được (mạng chặn external ở cả server lẫn máy dev). Cần **VM throwaway**.

## Windows (winrs-cert) — ✅ PASS (một phần) + bug fix
- certstore nạp cert `.pem` thật qua request (`cn=e2e-client`, `has_key=true`).
- Plugin sống sót sau exec lỗi.
- 🐛 **Bug thật phát hiện + fix**: exec tới host chết TREO vô hạn (truyền `0` vào winrm timeout).
  Fix: hard dialTimeout + bọc goroutine/`select ctx.Done()` → handler LUÔN trả về trong timeout.
  Sau fix: exec host chết → "LỖI MẠNG: timeout" trong ~timeout giây. Verify PASS.
- ⏳ **Chưa test**: cert-auth exec THẬT thành công — cần WinRM HTTPS + cert-auth target, cần
  **admin** để cấu hình (máy dev không có admin). AC cuối P1-09.

## Kết luận
Core + plugin manager + go-plugin + API động + certstore: **verify trên Linux + Windows thật**.
Còn lại (full installer systemd+MariaDB, cert-auth exec thật) cần VM throwaway + admin/Win target.
