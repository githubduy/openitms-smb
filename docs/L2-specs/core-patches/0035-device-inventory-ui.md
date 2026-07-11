---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0035-device-inventory-ui.patch, plugins/device-inventory/, docs/L2-specs/plugins-device-inventory.md]
---

# Spec patch 0035 — UI "Devices" (CMDB) cho plugin device-inventory

## Mục đích (WHY)
Plugin `device-inventory` (Phase 1/2b) cung cấp API CMDB nhưng **frontend Semaphore không có cơ chế
render plugin UI** (`menu_title` trong manifest chưa được frontend đọc). User không thấy device/asset
trên UI. Patch thêm **view "Devices"** làm vỏ UI gọi API động của plugin.

Nguyên tắc: **logic ở plugin, core chỉ là vỏ UI** (view proxy dữ liệu qua `/api/plugins/...`). Đây là
ngoại lệ UI hợp lý — frontend chưa có plugin-UI framework nên phải thêm view chuyên biệt (giống các
view custom trước: WinRS Console 0010, Network Discovery 0018).

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/DeviceInventory.vue` | file mới | Bảng device (icon host/switch, host, name, vendor/model, os/fw, last_seen); dialog detail theo `kind` (host→software/services/patches; switch→interfaces/neighbors/fdb) + tab Changes; dialog "Collect switch (SNMP)" (v2c community / v3 user+auth+priv) → POST collect-switch |
| `web/src/App.vue` | sửa | +nav item `devices` (icon mdi-devices, `${base}/devices`) + tooltip `tooltipDevices` |
| `web/src/router/index.js` | sửa | +route `/project/:projectId/devices` → DeviceInventory |
| `web/src/lang/en.js`, `vi.js` | +key | `devicesTitle/Hint`, `dev*`, `tooltipDevices`, `refresh` |

## API plugin dùng
`GET /api/plugins/device-inventory/devices` · `/device?id=` · `/changes?id=` ·
`POST /collect-switch`. Core đã authenticate + proxy tới plugin (HandleRequest).

## Rebase
- Neo: mảng `navItems()` + map `navTooltips` trong `App.vue`; danh sách route trong `router/index.js`.
  Upstream đổi cấu trúc nav/router → gắn lại item `devices` + route.
- View độc lập (file mới) — hiếm conflict.

## Verify (E2E)
- Menu "Devices" (mdi-devices) hiện trong sidebar project; mở → view load, gọi `/devices`.
- Chưa có device → bảng rỗng + hint. Dialog "Collect switch" nhập host/community → POST.
- (Switch thật) sau collect → device xuất hiện, detail hiện interfaces/neighbors/fdb.
- eslint sạch; chain 0001–0035 build.

## Liên quan
- Plugin: [[plugins-device-inventory]]. View custom tương tự: [[0010-winrs-console]] (nếu có),
  [[0018-network-discovery]].
