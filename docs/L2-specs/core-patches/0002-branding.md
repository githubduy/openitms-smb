---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [core-patches/0002-branding.patch]
---

# Spec patch 0002 — Branding OpenITMS-SMB (v1)

## Mục đích (WHY — dùng khi rebase)
MIT không cấp quyền trademark (plan mục 2.3): thay tên/logo "Semaphore" bằng
**OpenITMS-SMB** (ADR-0004 #1). Chỉ dùng **asset-replace + thay string literal**,
không đổi cấu trúc code — rebase khi conflict = áp lại cùng phép thế.

## Thay đổi (v1 — phần verify được không cần node)
| File | Phép thế |
|---|---|
| `web/public/index.html` | `<title>` → `OpenITMS-SMB` |
| `web/src/assets/logo.svg` | Thay bằng logo OpenITMS-SMB (SVG text, 1 dòng) |
| `web/public/favicon.svg` | Cùng logo trên |
| `web/src/lang/{de,en,nl,uk,zh_tw}.js` | key `ansibleSemaphore: 'Semaphore UI'` → `'OpenITMS-SMB'` (key này là tên app hiển thị trên UI) |

## Còn lại (P1-07b — cần node/FULL_UI để build + verify UI thật)
- `favicon.png` (binary — thay khi có pipeline asset).
- Các chuỗi i18n mô tả dài còn nhắc "Semaphore" (19 file) — thay + build UI verify.
- Trang About: attribution "Based on Semaphore UI (MIT)" (nghĩa vụ plan 2.1).
- Đổi các nhãn "Upgrade to Semaphore PRO" (App.vue) — quyết định ẩn/hiện tính năng PRO.

## Verify
- `apply-patches.sh` chuỗi 0001+0002 từ upstream sạch PASS; build backend PASS.
- UI thật: verify bằng CI job FULL_UI (cần node) — chưa chạy local (máy dev không có node).
