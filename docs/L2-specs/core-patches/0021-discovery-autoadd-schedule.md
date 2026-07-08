---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-08
related-code: [core-patches/0021-discovery-autoadd-schedule.patch, core-patches/0018-network-discovery.patch, core-patches/0009-winrs-app.patch]
---

# Spec patch 0021 — Discovery: quét định kỳ + add-to-inventory + auto-add WinRS

## Mục đích (WHY)
Nâng cấp Network Discovery (0018) theo yêu cầu: (1) quét subnet **định kỳ**, (2) nút **"Add all"** +
**"Add"** từng client vào inventory, (3) tuỳ chọn **tự động thêm** máy nào scan ra + **connect được
bằng WinRS** (cert mặc định) vào inventory.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_discovery.go` | sửa | `discoveryConfig` +`AutoScan`/`ScanInterval`(min)/`AutoAddWinRS`; `addHostsToWinRSInventory` (thêm IP vào inventory WinRS, cert = `QUICKWIN_WINRS_DEFAULT_CERT` hoặc `<ip>.pem`); `winrsReachable` (chạy `hostname` qua winrsexec bằng cert mặc định) + `autoAddReachable` (device unmanaged có cổng 5985/5986 → test WinRS → thêm nếu OK); handlers `SetDiscoveryConfig`, `AddDiscoveryToInventory`; `RunDueScans(store)` (scheduler); `managedHosts` đổi nhận `db.Store` |
| `api/router.go` | +2 route | `POST /discovery/config`, `POST /discovery/add-to-inventory` |
| `cli/cmd/root.go` | +goroutine | scheduler gọi `projects.RunDueScans(store)` mỗi 60s (quét project bật `auto_scan` + tới hạn) |
| `web/src/views/project/NetworkDiscovery.vue` | sửa | switch "Scan periodically" + interval + "Auto-add WinRS-reachable" (lưu qua /config); nút **"Add all to inventory"** (mọi unmanaged/gateway) + nút **"Add"** từng device |
| `web/src/lang/en.js` | +key | `discoveryAutoScan`, `discoveryIntervalMin`, `discoveryAutoAdd`, `discoveryAddAll`, `discoveryAddClient` |

## Hành vi
- **Add to inventory**: thêm dòng `<ip> cert=<cert>` vào inventory "WinRS Endpoints" (tạo nếu chưa có),
  bỏ qua IP đã có. cert = default cert (nếu set) hoặc `<ip>.pem` (placeholder tới khi enroll host đó).
  Sau khi thêm → device thành **managed**.
- **Auto-add WinRS**: chỉ chạy khi `QUICKWIN_WINRS_DEFAULT_CERT` set. Với device unmanaged có 5985/5986
  mở → thử `winrsexec.Run("hostname")` bằng cert đó; OK → thêm inventory + managed. (Máy đã enroll cùng
  cert sẽ tự được nhận + đưa vào quản lý.)
- **Quét định kỳ**: `RunDueScans` mỗi 60s duyệt project; project bật `auto_scan` + `interval>0` + tới hạn
  (`now - last_scan.ts >= interval*60`) → quét **subnet đầu** (nên đặt /24 để nhanh) + classify + auto-add.

## Bảo mật / giới hạn
- Route gate `CanManageProjectResources`. Auto-add cần cert mặc định (không bừa bãi).
- Định kỳ quét subnet đầu (tránh /16 lặp nặng) — UI khuyến nghị /24. Scheduler bounded bởi deadline scan.

## Verify (E2E app thật)
- `POST /discovery/config` → lưu auto_scan/interval/auto_add.
- `POST /discovery/add-to-inventory {ips}` → thêm vào inventory WinRS (added=N); scan loopback `/29` →
  "Add all" 5 IP → **re-scan hiện managed hết**.
- eslint sạch; `go build ./api/... ./cli/...` OK; chain 0001–0021 apply+build.

## Liên quan
- Discovery gốc: [[0018-network-discovery]]. WinRS engine: [[0009-winrs-app]].
