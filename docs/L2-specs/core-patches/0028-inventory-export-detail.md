---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-09
related-code: [core-patches/0028-inventory-export-detail.patch, core-patches/0027-inventory-filter-diff-export.patch, core-patches/0025-winrs-inventory.patch]
---

# Spec patch 0028 — Inventory: export CSV chi tiết (long-format)

## Mục đích (WHY)
Bản export CSV ở 0027 là **tóm tắt** (1 dòng/host, chỉ đếm số lượng software/services/tasks/hotfixes).
Audit thực tế cần **chi tiết từng mục** để lọc/pivot trong Excel (vd: "máy nào cài AnyDesk?",
"service X đang ở state gì trên toàn fleet?"). Patch này thêm CSV long-format 1 dòng/mục.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_winrs_inventory.go` | sửa | Thay `invDataSummary` (dùng `json.RawMessage` + đếm) bằng `invDataFull` parse đủ trường (`invNameVer`, `invService`, `invTask`, `invHotfix`). `ExportWinRSInventory` nhận `?detail=`; switch header + rows theo loại |
| `web/src/views/project/WinRSConsole.vue` | sửa | Nút CSV đơn → `v-menu` dropdown 5 mục; method `invCsvUrl(detail)` build URL |
| `web/src/lang/en.js` | +key | `invExpSummary/Software/Services/Tasks/Hotfixes` |

## API
`GET /winrs/inventory/export`
- `?format=json` — raw (không đổi).
- `?format=csv` (không `detail`) — **tóm tắt** 1 dòng/host (như 0027).
- `?format=csv&detail=software` — cột `ip,hostname,software,version` (1 dòng/app).
- `?format=csv&detail=services` — cột `ip,hostname,service,state,start,microsoft`.
- `?format=csv&detail=tasks` — cột `ip,hostname,task_path,task,state`.
- `?format=csv&detail=hotfixes` — cột `ip,hostname,hotfix,installed`.

Filename theo loại: `inventory-<detail>.csv` (hoặc `inventory.csv` cho tóm tắt).
Không tạo route mới — dùng lại route export của 0027, chỉ thêm query param.

## Verify (E2E máy thật)
- `detail=software` → header đúng + 1 dòng/app (name+version) cho từng host.
- `detail=services` → cột `microsoft` = true/false (client có thể lọc trong Excel như UI).
- Bản tóm tắt (không detail) vẫn giữ nguyên format cũ.
- eslint sạch; `go build ./api/projects/` OK; chain 0001–0028 build.

## Liên quan
- Export tóm tắt + filter Microsoft: [[0027-inventory-filter-diff-export]].
- Inventory gốc: [[0025-winrs-inventory]], [[0026-inventory-extra-schedule]].
