---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-08
related-code: [core-patches/0018-network-discovery.patch, core-patches/0009-winrs-app.patch, winrs-exec/]
---

# Spec patch 0018 — Network Autodiscovery

## Mục đích (WHY)
Ngoài nhập subnet thủ công, cho phép **quét dãy mạng** (CIDR) để tự tìm thiết bị đang online và
phân loại: đã quản lý (managed), mới (unmanaged), bỏ qua (exception), gateway/router. Mặc định mỗi
project có sẵn subnet **192.168.0.0/16**. Giúp SMB thấy nhanh máy nào trong mạng, máy nào chưa quản lý.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_discovery.go` | MỚI | Storage JSON (`discovery/project_<id>.json`, env `QUICKWIN_DISCOVERY_DIR`); scanner `scanCIDR` (probe TCP các cổng 445/3389/22/5985/5986/80/443, concurrency 256, dial 300ms, deadline 60s, chặn >70000 host, bỏ network/broadcast); `classify` (ưu tiên exception > managed > gateway > unmanaged); handlers `GetDiscovery`, `AddDiscoverySubnet`, `RemoveDiscoverySubnet`, `ScanDiscovery`, `SetDiscoveryDevice` |
| `api/router.go` | +5 route | `GET /discovery`, `POST /discovery/subnet`, `/discovery/subnet/remove`, `/discovery/scan`, `/discovery/device` (projectUserAPI) |
| `web/src/views/project/NetworkDiscovery.vue` | MỚI | Quản lý subnet (chip + thêm/xoá), ô CIDR + nút Scan (cảnh báo range rộng), bảng device (IP/status chip/ports/note/ignore), legend, tiến độ scanned/total + partial |
| `web/src/router/index.js` | +route | `/project/:id/network_discovery` |
| `web/src/App.vue` | +nav +tooltip | menu "Network Discovery" (mdi-lan) + tooltip |
| `web/src/lang/en.js` | +key | `discovery*`, `status*`, `note`, `tooltipNetworkDiscovery` |

## Phân loại thiết bị
- **managed**: IP có trong inventory WinRS của project (parse qua `winrsexec.ParseHosts`).
- **exception**: IP người dùng đánh dấu bỏ qua (lưu trong config).
- **gateway**: heuristic octet cuối `.1` hoặc `.254` (router/gateway thường).
- **unmanaged**: online + không thuộc các nhóm trên → máy mới chưa quản lý.

## Online detection
Probe TCP nhiều cổng phổ biến; host **online** nếu bất kỳ cổng mở HOẶC bị `connection refused`
(host bật, cổng đóng vẫn trả RST). Host bị firewall drop hết → không phát hiện (giới hạn đã biết —
dùng thêm agent/enroll để chắc chắn).

## Hiệu năng / giới hạn
- Đồng bộ, bounded concurrency + deadline 60s. `/24` (256 host) ~ vài giây; **`/16` (65k host) sẽ
  quét PARTIAL trong 60s** → UI cảnh báo range rộng, khuyến nghị quét theo `/24`.
- Config lưu file JSON (không đổi schema DB, hợp mô hình patch). Default subnet `192.168.0.0/16`
  trả về cho MỌI project khi chưa có config → thoả yêu cầu "Demo Project có sẵn /16".

## Verify (E2E app thật)
- `GET /discovery` → default `["192.168.0.0/16"]`.
- `scan 127.0.0.0/29` → 6 host online, port phát hiện đúng, `.1` → gateway, completed=true.
- `device` mark `127.0.0.3` exception + note → lưu + reclassify khi GET lại.
- `subnet` add `10.10.10.0/24` OK; CIDR sai → 400.
- eslint sạch; chain 0001–0018 build; bundle chứa "Network Discovery".

## Còn lại (nâng cấp sau)
- Quét async + streaming tiến độ (thay đồng bộ) để `/16` full không chặn request.
- Reverse-DNS hostname; phát hiện gateway thật từ routing table; probe ICMP (cần quyền).
- Nút "enroll" trực tiếp từ device unmanaged → WinRS Console 1-click.

## Liên quan
- Inventory/engine WinRS (managed source): [[0009-winrs-app]].
