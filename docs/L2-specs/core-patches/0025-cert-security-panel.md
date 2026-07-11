---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0025-cert-security-panel.patch]
---

# Spec patch 0025 — Panel cảnh báo bảo mật cert (.pem) trên WinRS Console

## Mục đích (WHY)
File cert `.pem` (vd `openitms-<HOST>.pem`) chứa **cả private key** — là credential đăng nhập admin
lên mọi máy Windows tin cert đó. Nếu lộ = kiểm soát toàn fleet. Cần cảnh báo rõ ngay trên trang WinRS
Console (nơi user đặt/dùng cert).

> Patch này **tách ra từ `0025-winrs-inventory` cũ**. Phần inventory tự-viết trong core (thu thập/lưu/
> so sánh/export software/services…) đã **bỏ ở Phase 4** vì chuyển sang plugin `device-inventory`
> (osquery + SNMP + CMDB). Chỉ giữ lại phần cảnh báo bảo mật cert.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/project/WinRSConsole.vue` | sửa | +expansion panel cảnh báo (viền cam, icon shield): mô tả rủi ro nếu `.pem` lộ + cách bảo vệ |
| `web/src/lang/en.js` | +key | `certSecurityTitle/Body/Risk1..3/Do` |

vi.js đã có sẵn các key `certSecurity*` (patch 0031 dịch toàn bộ).

## Verify
- Mở WinRS Console → thấy panel "Security: protect the .pem certificate file" (expand xem chi tiết).
- eslint sạch; chain build.

## Liên quan
- Plugin thay thế inventory core: [[plugins-device-inventory]].
- Tiếng Việt: [[0031-vietnamese-language]].
