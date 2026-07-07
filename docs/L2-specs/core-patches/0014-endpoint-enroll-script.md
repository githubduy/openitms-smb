---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0014-endpoint-enroll-script.patch, core-patches/0009-winrs-app.patch, winrs-exec/e2e/]
---

# Spec patch 0014 — Tải script chuẩn bị endpoint (WinRS / SSH) từ web UI

## Mục đích (WHY)
Để OpenITMS quản lý 1 máy Windows, máy đó phải bật WinRM HTTPS + cert auth (hoặc OpenSSH). Trước
đây người dùng phải tự làm thủ công. Patch cho **tải script cài đặt ngay trên web**: người dùng
(kể cả không chuyên) tải về, chạy AS ADMIN trên máy đích; script tự bật kết nối + xuất chứng chỉ
+ in hướng dẫn nạp lên OpenITMS.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/scripts/winrs-enroll.ps1` | MỚI | Tạo user cục bộ + cert, bật WinRM HTTPS (5986) + cert auth + firewall + map cert→user, xuất PEM (cert+key) ra Desktop, in IP + dòng inventory gợi ý |
| `api/projects/scripts/ssh-enroll.ps1` | MỚI | Cài + bật OpenSSH Server (22), firewall, đặt pwsh làm default shell, thêm `-PublicKey` vào administrators_authorized_keys |
| `api/projects/quickwin_enroll.go` | MỚI | `//go:embed` 2 script; `GetEnrollScript` trả script theo `{kind}=winrs\|ssh` dạng attachment (Content-Disposition, nosniff) |
| `api/router.go` | +1 route | `GET /project/{id}/endpoint/enroll-script/{kind}` trên `projectUserAPI` |
| `web/src/views/project/WinRSConsole.vue` | +panel | Expansion "Tải script chuẩn bị endpoint" + 2 nút (`<a download>` WinRS / SSH) |
| `web/src/components/InventoryForm.vue` | +link | Editor WinRS thêm link "Tải script chuẩn bị máy Windows" |

## Bảo mật
- Route gate `CanManageProjectResources` (projectUserAPI). Tải qua `<a download>` same-origin → gửi cookie auth.
- Script nhúng binary (go:embed) → không đọc file ngoài lúc chạy, không path-injection.
- Script **cảnh báo rõ** đổi cấu hình bảo mật hệ thống, chỉ chạy trên máy nội bộ có quyền admin.
- Không sinh/không nhúng secret trong script tải về (cert sinh TRÊN máy đích, không qua server).

## Luồng người dùng (enrollment)
1. Web UI → tải `openitms-winrs-enroll.ps1`.
2. Chạy AS ADMIN trên máy Windows đích → script bật WinRM+cert, xuất `openitms-<host>.pem`.
3. Copy PEM vào `certs/` của máy chủ OpenITMS.
4. Inventory > WinRS Endpoints: thêm `<ip> cert=<file.pem>`.
5. Chạy WinRS Console / template xuống máy đó.

## Verify
- `go build ./api/...` OK (go:embed script). eslint WinRSConsole.vue + InventoryForm.vue sạch.
- Chain 0001–0014 apply sạch + build; API `/endpoint/enroll-script/winrs` tải đúng nội dung .ps1.

## Liên quan
- Engine + cert model: [[0009-winrs-app]]. Setup lab E2E tương tự: `winrs-exec/e2e/setup-winrm-cert.ps1`.
