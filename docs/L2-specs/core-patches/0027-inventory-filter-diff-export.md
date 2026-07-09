---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0027-inventory-filter-diff-export.patch, core-patches/0025-winrs-inventory.patch, core-patches/0026-inventory-extra-schedule.patch]
---

# Spec patch 0027 — Inventory: lọc Microsoft + so sánh 2 lần quét + export CSV/JSON

## Mục đích (WHY)
Làm inventory dễ audit hơn:
- **(A) Lọc Microsoft** khỏi services/tasks (mặc định ẩn — chỉ hiện bên thứ 3). Checkbox "Show
  Microsoft" hiện lại; checkbox "Only running services" chỉ hiện service đang chạy.
- **(B) So sánh** giữa 2 lần quét: phát hiện software mới cài / gỡ, service đổi state (vd Running→Stopped).
- **(C) Export** CSV/JSON inventory toàn fleet.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/collect-inventory.ps1` | sửa | Service +cờ `ms` = `PathName -like '*\Windows\*'` (built-in/Microsoft) để client lọc |
| `api/projects/quickwin_winrs_inventory.go` | sửa | `winrsInvHost` +`Prev`/`PrevTS`; `setHostData` (data mới → cũ chuyển sang prev, giữ 1 mốc); `ExportWinRSInventory` (`?format=json` raw / `csv` 1 dòng/host tóm tắt: ip/hostname/domain/os/build/ips/dns/gateway + đếm users/groups/software/services/tasks/hotfixes + collected) |
| `api/router.go` | +route | `GET /winrs/inventory/export` |
| `web/src/views/project/WinRSConsole.vue` | sửa | checkbox `showMicrosoft`/`onlyRunning`; `filteredServices`/`filteredTasks` (lọc ms + Running); panel "Changes since last scan" (`changes(h)` diff data vs prev: swAdded/swRemoved/service state change); nút Export CSV + JSON (`<a :href download>`) |
| `web/src/lang/en.js` | +key | `invShowMs`, `invOnlyRunning`, `invChangesTitle`, `invSwAdded`, `invSwRemoved` |

## Lọc Microsoft
- **Services**: cờ `ms` (chạy từ `%WINDIR%`) → ẩn mặc định. Heuristic (một số 3rd-party cài vào System32
  vẫn bị coi ms) — chấp nhận cho audit; checkbox hiện lại toàn bộ.
- **Tasks**: `path` bắt đầu `\Microsoft\` → ẩn mặc định.

## So sánh (client-side)
- Lưu 1 mốc `prev` mỗi host (data trước đó). `changes(h)`: software theo `name` (added/removed);
  service theo `name` với `state` khác nhau. Alert "Changes since last scan" chỉ hiện khi có thay đổi.

## Verify (E2E máy thật)
- Re-collect → **259/288 service có cờ `ms`** (còn ~29 third-party khi ẩn Microsoft).
- Export CSV → header + 1 dòng/host (software 142/services 288/tasks 249/hotfixes 4).
- Export JSON → Content-Disposition attachment.
- eslint sạch; `go build ./api/...` OK; chain 0001–0027 build.

## Liên quan
- Inventory: [[0025-winrs-inventory]], [[0026-inventory-extra-schedule]].
