---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0013-menu-tooltips.patch]
---

# Spec patch 0013 — Tooltip giải thích menu sidebar (SMB-friendly)

## Mục đích (WHY)
Người dùng SMB không chuyên IT thấy các menu (Inventory, Environment, Key Store…) khó hiểu:
không rõ mỗi mục để làm gì, cần tạo gì. Thêm tooltip tiếng Việt ngắn gọn khi rê chuột vào mỗi
menu, giải thích ý nghĩa + việc cần làm.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `web/src/App.vue` | +computed +template | `navTooltips` map `key → mô tả` cho 10 menu (templates, schedule, inventory, winrs_console, environment, keys, repositories, integrations, team, runners); bọc 2 vòng render nav (pinned + unpinned) bằng `v-tooltip right` (open-delay 300ms, `:disabled` khi menu không có mô tả) |

## Nội dung tooltip (tiếng Việt, giải thích "cần tạo gì / ý nghĩa")
- **Inventory**: danh sách máy đích; "WinRS Endpoints" = máy Windows qua cert, "Ansible Inventory" = nhóm máy kiểu Ansible.
- **Task Templates**: mẫu "việc cần chạy" (script + máy đích), bấm 1 nút là chạy.
- **WinRS Console**: gõ nhanh 1 lệnh xuống 1 máy Windows, xem kết quả ngay.
- **Environment / Key Store / Repositories / Schedule / Integrations / Team / Runners**: mỗi mục 1 câu.

## Verify
- eslint App.vue sạch.
- Chain 0001–0013 apply sạch + build; bundle chứa chuỗi tooltip.
- UI: rê chuột menu → hiện tooltip bên phải.

## Ghi chú
- Chỉ frontend, không đụng backend. Tooltip dùng chuỗi cứng tiếng Việt (chưa i18n) — có thể chuyển
  sang `$t()` khi cần đa ngôn ngữ.
