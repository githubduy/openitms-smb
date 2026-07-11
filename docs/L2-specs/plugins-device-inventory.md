---
level: L2
status: draft
owners: [maintainer]
updated: 2026-07-11
related-code: [plugins/device-inventory/]
---

# Spec: Plugin `device-inventory` (osquery + CMDB gọn)

## Mục đích (WHY)
Tách **asset/device inventory** ra khỏi core: device list + software/hardware do **osquery** (agentless
qua WinRS) thu, lưu thành **CMDB gọn trong MariaDB**. Thay cho inventory tự-viết trong core
(patch 0025–0028) sẽ được deprecate. Inventory **config** (host/cred) vẫn ở DB core (Semaphore native).

Quyết định kiến trúc (chốt với maintainer):
- Không dùng GLPI (PHP nặng) / NetBox / Fleet (đẻ DB engine mới).
- osquery = collector chuyên nghiệp, agentless; CMDB do plugin quản tối giản trong MariaDB sẵn có.

## Kiến trúc
- Plugin Go qua SDK (`quickwin.dev/sdk`), mount động ở `/api/plugins/device-inventory/*`.
- **DB**: dùng CHÍNH database app (`mysql.name` trong config.json) + **prefix bảng `di_`**.
  Lý do: user app **không có quyền CREATE DATABASE** trên bản cài mặc định → database riêng cần root
  (không portable). Prefix `di_` cho namespace rõ, backup chung, zero-privilege-issue.
- DSN đọc từ `QUICKWIN_CONFIG` (config.json) — plugin kế thừa env của app.

## Routes (Phase 1)
| Method | Path | Mô tả | Trạng thái |
|---|---|---|---|
| GET | `devices` | Danh sách device | ✅ Phase 1 |
| GET | `device?id=` | Chi tiết (software/services/patches) | ✅ Phase 1 |
| GET | `changes?id=` | Lịch sử thay đổi | ✅ Phase 1 (rỗng tới Phase 2) |
| POST | `collect` | Thu osquery/WinRS → upsert + diff | ⏳ Phase 2 |
| GET | `export?format=` | Export fleet | ⏳ Phase 3 |

## Schema (`di_*`)
`di_device` (host UNIQUE, hostname/os/os_version/os_build/domain, first_seen/last_seen/last_status),
`di_device_software` (name/version), `di_device_service` (name/state/start), `di_device_patch`
(kb/installed), `di_device_change` (ts/kind/detail — lịch sử diff mỗi lần quét). FK ON DELETE CASCADE.

## Roadmap
- **Phase 2**: `collect` — bundle osqueryi, đẩy qua WinRS chạy 1 lần, parse JSON (packs software/
  services/patches/hw/net), upsert + sinh `di_device_change` (diff so với lần trước).
- **Phase 3**: UI "Devices" trong OpenITMS (list/detail/changes/export).
- **Phase 4**: deprecate core inventory 0025–0028.
- **Phase 5**: bundle osqueryi trong installer.

## Verify (Phase 1, E2E máy thật)
- Plugin nạp (log core "plugin chạy device-inventory"); `GET /devices` → 200 `{"devices":[]}`.
- Schema `di_*` tự tạo trong database `openitms`. Unit test: queryID/jsonResp/Metadata-khớp-manifest.

## Liên quan
- Manifest/proto: [[plugin-manifest]], [[proto-contract]]. Mẫu: [[plugins-winrs-cert]].
