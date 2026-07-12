---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0038-unify-device-model.patch, plugins/device-inventory/, core-patches/0035-device-inventory-ui.patch]
---

# Spec patch 0038 — Gộp Inventory + CMDB: device là thực thể trung tâm

## Mục đích (WHY)
Menu **Inventory** (Semaphore native: máy đích chạy task) và **Thiết bị (CMDB)** (tài sản thu thập) dễ
nhầm. Maintainer muốn **1 model quản lý**: device = **identity + cách kết nối + tài sản**. Trang Thiết
bị thành hub trung tâm; Inventory native de-emphasize (giữ cho task-targeting/Ansible — không phá).

## Thay đổi (core patch — UI)
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/DeviceInventory.vue` | sửa | Thay 2 dialog collect-host/switch bằng **1 dialog "Thêm thiết bị"** (conn_type local/winrs/snmp + trường tương ứng) → lưu kết nối rồi thu ngay. Bảng +cột hành động per-row **Collect** (thu lại bằng kết nối đã lưu) + **Delete**. Giữ nút "Thu server này (local)" |
| `web/src/App.vue` | sửa | Đổi menu Inventory native: title `navTaskTargets` ("Máy đích (task)") + icon mdi-target + tooltip `tooltipTaskTargets` (rõ là để chạy task, không phải CMDB) |
| `web/src/lang/en.js`, `vi.js` | +key | `devAddDevice/devSaveCollect/devConnType/devConn*/devAddHint/delete/navTaskTargets/tooltipTaskTargets` |

## Backend (plugin — commit thẳng, không patch)
- `di_device` +cột `conn_type/conn_cert/conn_port/snmp_version/snmp_community` (lưu cách kết nối cùng device).
- `POST /device` (thêm/sửa device+kết nối), `DELETE /device`, `POST /collect {id}` (thu theo kết nối
  đã lưu, dispatch local→osquery local / winrs→osquery WinRS / snmp→SNMP).

## Mô hình hợp nhất
- **Thiết bị (CMDB)** = hub: thêm device (kèm kết nối) → thu (dùng kết nối đã lưu, không nhập lại) →
  xem tài sản + lịch sử. 1 chỗ quản lý host (WinRS/local) + switch (SNMP).
- **Máy đích (task)** (Inventory native) = riêng cho Task Template chạy lên (WinRS Endpoints/Ansible).
  Không bị xoá (task-targeting còn nguyên), chỉ đổi tên cho hết nhầm.

Giới hạn: SNMP lưu kết nối chỉ v2c (community); v3 để sau. conn_cert = tên file trong ./certs (không
phải secret). Community lưu plaintext (như các secret khác của app) — mã hoá là cải tiến tương lai.

## Verify (E2E)
- Thiết bị: nút "Thêm thiết bị" → chọn local/winrs/snmp → Lưu & thu; per-row Collect/Delete chạy.
- Menu đổi "Máy đích (task)"; "Thiết bị (CMDB)" là hub. eslint sạch; chain 0001–0038 build.

## Liên quan
- UI Devices gốc: [[0035-device-inventory-ui]]. Local: [[0037-local-self-collect]]. Plugin: [[plugins-device-inventory]].
