---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0041-discovery-cmdb-inventory-picker.patch, core-patches/0038-unify-device-model.patch, plugins/device-inventory/]
---

# Spec patch 0041 — Discovery Add → CMDB + chọn IP từ CMDB trong WinRS inventory

## Mục đích (WHY)
Tiếp tục "CMDB là trung tâm":
- **(A)** Máy phát hiện qua Network Discovery khi bấm **Add** nên vào **CMDB device** (không chỉ WinRS
  inventory) → thấy ngay trong hub Thiết bị.
- **(B)** Editor **WinRS Endpoints** (máy đích task) đang gõ tay `<ip> cert=<file>` → nên **chọn host
  từ CMDB** cho nhanh, đỡ sai.

## Thay đổi (core patch — UI)
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/NetworkDiscovery.vue` | sửa | `addToInventory(ips)` +POST `/api/plugins/device-inventory/device` (host=ip, conn_type=winrs) cho mỗi IP → discovery Add vào cả CMDB (song song với inventory cũ) |
| `web/src/components/InventoryForm.vue` | sửa | Type WinRS: +`v-autocomplete` nạp host từ CMDB (`GET /plugins/device-inventory/devices`, lọc kind≠switch); chọn → `addHostFromCmdb` chèn `<ip> cert=<ip>.pem` vào `item.inventory` (bỏ trùng). +import axios |
| `web/src/lang/en.js`, `vi.js` | +key | `winrsPickFromCmdb`, `winrsCmdbEmpty` |

## Luồng hợp nhất
- Quét mạng → **Add** → máy vào CMDB (conn_type=winrs, cert đặt sau khi enroll) + WinRS inventory.
- Sửa WinRS Endpoints (task target) → dropdown **"Thêm host từ CMDB"** → chèn dòng, không gõ tay IP.
- CMDB là nguồn thiết bị; inventory task-target tham chiếu từ đó.

## Rebase
- Neo: `addToInventory` trong NetworkDiscovery.vue; block `item.type === 'winrs'` trong InventoryForm.vue.

## Verify (E2E)
- Discovery Add 1 IP → xuất hiện trong CMDB (`GET /devices`).
- Sửa WinRS inventory → chọn host từ dropdown → dòng `<ip> cert=<ip>.pem` được chèn. eslint sạch;
  chain 0001–0041 build.

## Liên quan
- Unify device model: [[0038-unify-device-model]]. Discovery inventory riêng: [[0030-discovery-separate-inventory]].
