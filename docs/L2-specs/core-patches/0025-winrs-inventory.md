---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0025-winrs-inventory.patch, core-patches/0010-winrs-console.patch, core-patches/0009-winrs-app.patch]
---

# Spec patch 0025 — Cảnh báo bảo mật cert + thu thập inventory máy managed

## Mục đích (WHY)
- **(A) Cảnh báo `.pem`**: file cert+key là **credential admin toàn fleet** — lộ = kẻ tấn công chạy
  lệnh admin trên mọi máy tin cert đó. Hiển thị rõ rủi ro + cách bảo vệ trên WinRS Console.
- **(B) Inventory**: máy đã managed có tuỳ chọn (mặc định BẬT) thu thập: software, IP, DNS, route,
  hostname, user/group, domain — bằng script collect+parse gọn.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/collect-inventory.ps1` | MỚI | Thu thập qua CIM (Win32_ComputerSystem/OperatingSystem) + registry Uninstall (software) + Net*/DnsClient (IP/DNS/route) + LocalUser/LocalGroup → **1 dòng JSON compact** (`ConvertTo-Json -Compress`) để server parse |
| `api/projects/quickwin_winrs_inventory.go` | MỚI | `//go:embed` script; `WinRSInventoryCollect` (chạy qua winrsexec trên host → validate JSON → lưu `inventory_<pid>.json`); `GetWinRSInventory` + `SetWinRSInventoryConfig` (enabled, mặc định true) |
| `api/router.go` | +3 route | `GET /winrs/inventory`, `POST /winrs/inventory/config`, `POST /winrs/inventory/collect` |
| `web/src/views/project/WinRSConsole.vue` | +panel +section | Panel cảnh báo bảo mật `.pem` (mở rộng, liệt kê rủi ro + cách bảo vệ); section Inventory (switch bật/tắt + nút Collect + expansion mỗi host: hostname/domain/OS/IPs/DNS/gateway/users/groups/software list) |
| `web/src/lang/en.js` | +key | `certSecurity*` (5), `inv*` (7) |

## Rủi ro nếu .pem lộ (nội dung cảnh báo)
- Chạy BẤT KỲ lệnh nào với quyền admin trên mọi máy Windows tin cert → toàn quyền điều khiển fleet.
- Đọc/sửa file, cài phần mềm, tạo tài khoản, tắt bảo mật — tương đương quyền admin vật lý.
- Lateral movement trong mạng.
- **Xử lý**: giữ như root password (chỉ trong `certs/` server, không share/commit); lộ → re-enroll máy
  (sinh cert mới, cert cũ hết tác dụng).

## Bảo mật / thiết kế
- Script **chỉ đọc** (không sửa gì trên máy đích). Route gate `CanManageProjectResources`.
- Chặn path-traversal cert/key. Output phải là JSON hợp lệ mới lưu (json.Valid).
- Inventory lưu file runtime (không đổi schema DB). Toggle mặc định bật.

## Verify (E2E máy thật đã managed)
- `POST /winrs/inventory/collect` {host, cert} → chạy script qua WinRS → JSON: hostname, domain, OS,
  ips, dns, gateway, users(7), groups(19), **software(142)** — parse OK, lưu file.
- `GET /winrs/inventory` → enabled + hosts; toggle config off/on OK.
- eslint sạch; `go build ./api/...` OK; chain 0001–0025 build.

## Liên quan
- WinRS Console: [[0010-winrs-console]]. Engine: [[0009-winrs-app]].
