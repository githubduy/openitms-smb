---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [core-patches/0006-management-ui.patch]
---

# Spec patch 0006 — UI quản lý OpenITMS

## Mục đích (WHY)
Trang quản lý cho admin (Settings/Plugins/Registry là tính năng core theo plan). Thuần
frontend Vue, gọi API đã có (patch 0001 `GET /api/plugins`, 0005 `/api/registry`, hardening plugin).

## Thay đổi (chỉ `web/`)
| File | Loại | Nội dung |
|---|---|---|
| `web/src/views/OpenITMS.vue` | MỚI | `v-tabs`: **Plugins** (data-table từ `GET /api/plugins`), **Registry** (search `/api/registry/search` + install `/api/registry/install`), **Hardening** (scan `/api/plugins/hardening/scan` + fix). Vuetify 2 + axios, self-contained |
| `web/src/router/index.js` | +import +route | route `/openitms` → `OpenITMS.vue` |
| `web/src/App.vue` | +1 nav item | mục "OpenITMS" (admin-only) sau "runners" |

## Phụ thuộc backend
- `GET /api/plugins` (list) — thêm ở patch 0001 (registerQuickwinPlugins).
- `GET /api/registry/search`, `POST /api/registry/install` — patch 0005.
- `GET /api/plugins/hardening/scan`, `POST .../fix` — plugin hardening (API động).

## Hướng dẫn rebase
Nếu router/App.vue upstream đổi: thêm lại import + route `/openitms` + nav item.
OpenITMS.vue độc lập (chỉ dùng Vuetify + axios) → hiếm khi vỡ.

## Verify
- `apply-patches.sh` chuỗi 0001–0006 apply sạch (backend build PASS).
- **UI compile: CI job `ui-build`** (npm ci + npm run build — cần node; máy dev không có node).
