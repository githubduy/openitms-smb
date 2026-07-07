---
level: L5
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [installer/]
---

# Test cài đặt mới từ đầu (fresh install) — 2026-07-07

Chạy trên Windows (dev) với MariaDB bundle từ binary local — mô phỏng install.sh từ zero.

## Quy trình (từ zero, sau khi wipe sạch data)
1. MariaDB binary từ bundle local (`installer/vendor/mariadb-winx64.zip` → giải nén).
2. `mariadb-install-db` → datadir MỚI.
3. `mariadbd` chạy → tạo DB `openitms` + user với **mật khẩu mặc định** `OpenITMS@MariaDB#2026`.
4. Build plugins (hello, hardening) vào thư mục plugins.
5. Config app dialect `mysql` (không bolt).
6. Migration chạy trên MariaDB + tạo admin `admin/quickwin123`.
7. Start app + verify.

## Kết quả (PASS)
| Kiểm | Kết quả |
|---|---|
| `/api/ping` | pong |
| login admin/quickwin123 | HTTP 204 |
| **boltdb_used** | **False** (chạy MariaDB, KHÔNG banner deprecated) |
| `/api/services` | app=up, mariadb=up |
| `/api/plugins` | hardening=running, hello=running |
| hardening scan (có QUICKWIN_CONFIG) | **phát hiện default-db-password** (WARN medium) + ui-tls |

## Ghi nhận từ test → fix production
- Hardening plugin cần biết đường dẫn config → systemd unit `openitms.service` nay export
  `OPENITMS_PREFIX`, `QUICKWIN_CONFIG`, `QUICKWIN_CERTS_DIR`, `QUICKWIN_GITEA_ADDR` cho core
  (plugin kế thừa env) — để hardening tìm được config + services tab thấy Gitea.
- MariaDB default password mặc định đủ mạnh, KNOWN (tiện quản trị, DB socket/localhost-only),
  hardening cảnh báo nếu chưa đổi.

## Còn lại (chưa test ở đây)
- Full `install.sh` systemd trên VM Linux throwaway (server chung không cài invasive).
- Gitea bundle (chờ binary staging — G-01).
