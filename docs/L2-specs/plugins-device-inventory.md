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

## Routes
| Method | Path | Mô tả | Trạng thái |
|---|---|---|---|
| GET | `devices` | Danh sách device (host + switch) | ✅ Phase 1 |
| GET | `device?id=` | Chi tiết (host: software/services/patches; switch: iface/neighbor/fdb) | ✅ |
| GET | `changes?id=` | Lịch sử thay đổi | ✅ |
| POST | `collect` | Thu host qua osquery/WinRS → upsert + diff | ✅ Phase 2 |
| POST | `collect-switch` | Thu switch qua SNMP (v2c/v3) → upsert + diff | ✅ Phase 2b |
| GET | `export?format=` | Export fleet | ⏳ Phase 3 |

## Network switch qua SNMP (Phase 2b)
Switch/router không chạy osquery → thu qua **SNMP** (thư viện OSS **gosnmp**, BSD). Hỗ trợ **v2c**
(community) + **v3** (user/auth/priv: MD5/SHA + DES/AES). `device.kind='switch'`.

Thu (full + topology):
- **System**: sysName/sysDescr/location/contact/uptime; đoán vendor từ sysDescr.
- **Hardware** (ENTITY-MIB): model/serial/firmware.
- **Interfaces** (IF-MIB + ifXTable): if_index/name/alias/type/speed/oper/mac.
- **LLDP neighbors** (LLDP-MIB): local_port ↔ remote_name/port/chassis (sơ đồ kết nối).
- **MAC/FDB** (BRIDGE-MIB dot1dTpFdb): mac ↔ port (endpoint nào ở cổng nào).

Bảng: `di_switch_iface`, `di_switch_neighbor`, `di_switch_fdb` + cột switch trên `di_device`
(kind/vendor/model/serial/firmware/location/descr/uptime). Diff số cổng/láng giềng → `di_device_change`.

Verify (Phase 2b): parse PDU unit-test (mac/oid-index/uptime/vendor/operstatus); route `collect-switch`
xử lý host không SNMP → 502 gọn (timeout). **Thu switch thật cần phần cứng của người dùng.**

## Schema (`di_*`)
`di_device` (host UNIQUE, hostname/os/os_version/os_build/domain, first_seen/last_seen/last_status),
`di_device_software` (name/version), `di_device_service` (name/state/start), `di_device_patch`
(kb/installed), `di_device_change` (ts/kind/detail — lịch sử diff mỗi lần quét). FK ON DELETE CASCADE.

## Host qua osquery (Phase 2)
Host Windows: chạy **osqueryi** trên máy đích qua **WinRS cert-auth** (dùng lại `winrsexec` +
`certstore`). 1 lệnh PowerShell chạy nhiều query, phân section bằng marker `@@<name>`; plugin parse
JSON từng bảng: `system_info` (hostname/vendor/model), `os_version`, `programs` (software),
`services`, `patches` (hotfix). Upsert `kind='host'` + software/services/patches; diff phần mềm
mới cài/gỡ → `di_device_change`.

- **osqueryi phải có sẵn** trên máy đích (PATH / Program Files / ProgramData) — nếu không → lỗi rõ
  "cần deploy osquery (Phase 5)". Bundle + auto-deploy osqueryi là Phase 5.
- Verify: parse unit-test (system/os/software/services/patches); WinRS transport E2E (kết nối +
  cert-auth handshake tới máy thật, nhận WinRM response). Green-path đủ dữ liệu cần máy enroll đúng
  + osquery cài.

## Roadmap còn lại
- **Phase 4**: deprecate core inventory 0025–0028 (gỡ khỏi core, device data do plugin quản).
- **Phase 5**: bundle osqueryi + đẩy-qua-WinRS-mỗi-lần trong installer; store SNMP/WinRS creds.

## Verify (Phase 1, E2E máy thật)
- Plugin nạp (log core "plugin chạy device-inventory"); `GET /devices` → 200 `{"devices":[]}`.
- Schema `di_*` tự tạo trong database `openitms`. Unit test: queryID/jsonResp/Metadata-khớp-manifest.

## Liên quan
- Manifest/proto: [[plugin-manifest]], [[proto-contract]]. Mẫu: [[plugins-winrs-cert]].
