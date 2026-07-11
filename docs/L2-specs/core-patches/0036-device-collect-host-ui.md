---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0036-device-collect-host-ui.patch, core-patches/0035-device-inventory-ui.patch, plugins/device-inventory/]
---

# Spec patch 0036 — UI "Collect host" (osquery) trên trang Devices

## Mục đích (WHY)
View Devices (0035) mới có nút thu **switch** (SNMP). Cần nút thu **host** (osquery/WinRS) từ giao
diện để dùng Phase 5 (plugin tự cài osquery khi máy đích chưa có) mà không phải gọi API tay.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/DeviceInventory.vue` | sửa | +nút "Collect host"; dialog nhập host + chọn cert (nạp từ `/api/project/{id}/winrs/certs`) + checkbox "auto-deploy osquery" → POST `/api/plugins/device-inventory/collect` |
| `web/src/lang/en.js`, `vi.js` | +key | `devCollectHost`, `devAutoDeploy`, `devAutoDeployHint` |

## Luồng
1. Bấm "Collect host" → dialog; `openCollectHost` nạp danh sách cert từ core `/winrs/certs`.
2. Nhập host/IP, chọn cert, bật/tắt auto-deploy (mặc định bật).
3. Submit → POST `/collect` {host, cert, auto_deploy}. Plugin chạy osquery qua WinRS (tự cài nếu
   thiếu + auto_deploy). Xong → reload danh sách device.

## Rebase
- Neo: toolbar + block `collect` trong `DeviceInventory.vue` (0035). Backend endpoint độc lập (plugin).

## Verify (E2E)
- Trang Devices có 2 nút: "Collect host" + "Collect switch".
- Dialog host nạp cert dropdown; POST /collect gọi plugin. eslint sạch; chain 0001–0036 build.

## Liên quan
- UI Devices: [[0035-device-inventory-ui]]. Auto-deploy osquery: [[plugins-device-inventory]] (Phase 5).
