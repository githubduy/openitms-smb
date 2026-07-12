---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0039-auto-collect-ui.patch, plugins/device-inventory/, core-patches/0038-unify-device-model.patch]
---

# Spec patch 0039 — UI bật/tắt thu định kỳ tự động (scheduler)

## Mục đích (WHY)
Thu inventory định kỳ **không nhập creds lại**. Kết nối đã lưu cùng device (0038: conn_type + cert/
community), nên chỉ cần UI để **bật/tắt scheduler + đặt chu kỳ**.

## Thay đổi (core patch — UI)
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/DeviceInventory.vue` | sửa | +switch "Tự động thu" + ô interval (phút) trên trang Thiết bị → GET/POST `/plugins/device-inventory/config`; `loadSched`/`saveSched` |
| `web/src/lang/en.js`, `vi.js` | +key | `devAutoCollect`, `devInterval`, `devAutoCollectHint` |

## Backend (plugin — commit thẳng)
- `di_config` (key/value): `auto_enabled`, `interval_min` (mặc định 240 = 4h, tắt).
- `GET/POST /config` (loadConfig/saveConfig).
- `runScheduler` goroutine (start ở main): mỗi 60s kiểm tra; tới hạn → `collectAllDue` (thu mọi device
  có `conn_type` qua `doCollect`, bỏ qua lỗi từng máy, log stderr). `doCollect` dùng chung với HTTP
  collectByID (dispatch local/winrs/snmp theo kết nối đã lưu).

## Hành vi
- Bật "Tự động thu" + đặt phút → scheduler định kỳ thu lại toàn bộ device (host WinRS/local + switch
  SNMP) bằng kết nối đã lưu. Không cần thao tác/nhập creds.
- WinRS: cert name đã lưu (file ở ./certs). SNMP: community v2c đã lưu. Local: không cần creds.

## Giới hạn
- Chu kỳ chung cho mọi device (chưa per-device). SNMP v3 chưa lưu creds → scheduler bỏ qua (v2c only).

## Verify (E2E)
- GET /config → mặc định {enabled:false, interval_min:240}; POST {enabled:true, interval_min:60} → GET
  xác nhận đã lưu. Scheduler goroutine chạy (log khi thu). UI: switch + interval hiện. chain 0001–0039 build.

## Liên quan
- Device lưu kết nối: [[0038-unify-device-model]]. Plugin: [[plugins-device-inventory]].
