---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0026-inventory-extra-schedule.patch, core-patches/0025-winrs-inventory.patch, core-patches/0021-discovery-autoadd-schedule.patch]
---

# Spec patch 0026 — Inventory: thêm services/tasks/patch level + auto-collect định kỳ

## Mục đích (WHY)
Mở rộng inventory (0025): thu thập thêm **services, scheduled tasks, patch level (hotfixes + os_build)**;
và **tự thu thập định kỳ** (mặc định **4h/lần**) cho mọi máy managed (không phải bấm tay từng máy).

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/collect-inventory.ps1` | sửa | +`services` (Win32_Service: name/state/start), +`tasks` (Get-ScheduledTask: name/path/state), +`hotfixes` (Get-HotFix: id/installed), +`os_build` |
| `api/projects/quickwin_winrs_inventory.go` | sửa | config +`collect_interval_min` (default 240) +`last_collect`; `collectHostInventory` (helper tách từ handler); `RunDueInventoryCollect(store)` — duyệt project bật + tới hạn → collect mọi host managed (host+cert từ inventory WinRS) |
| `api/router.go` | (0025) | route sẵn có |
| `cli/cmd/root.go` | +goroutine | check mỗi 5' → `RunDueInventoryCollect` (collect khi ≥ interval) |
| `web/src/views/project/WinRSConsole.vue` | sửa | field "Auto every (min)" (interval); hiển thị os_build + Hotfixes (KB+ngày) + Services (count+list, Running highlight) + Tasks (count+list) |
| `web/src/lang/en.js` | +key | `invIntervalMin`; cập nhật `invHint` |

## Auto-collect định kỳ
- Scheduler `RunDueInventoryCollect` chạy mỗi 5' (check), collect khi `now - last_collect >= interval*60`.
- Với mỗi project bật inventory: đọc host+cert từ inventory WinRS (`winrsexec.ParseHosts`), collect từng
  host (bỏ qua host offline/cert sai). Cập nhật `last_collect`.
- Default 240 phút (4h). Đặt qua UI field hoặc `POST /winrs/inventory/config {collect_interval_min}`.

## Verify (E2E máy thật đã managed)
- Collect → `os_build=22631`, `services=288`, `tasks=249`, `hotfixes=4` (KB5029921/KB5027397... + ngày),
  parse JSON OK.
- eslint sạch; `go build ./api/... ./cli/...` OK; chain 0001–0026 build.

## Liên quan
- Inventory gốc: [[0025-winrs-inventory]]. Pattern scheduler: [[0021-discovery-autoadd-schedule]].
