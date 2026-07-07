---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0007-service-manager.patch, service-manager/]
---

# Spec patch 0007 — Service Manager (status + restart hạ tầng)

## Mục đích (WHY)
Trang Admin `/openitms` cần xem trạng thái + restart các service hạ tầng (MariaDB, Gitea,
core app). Logic ở `quickwin.dev/servicemanager` (ngoài core).

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/quickwin_services.go` | MỚI | `GET /api/services` (list status) + `POST /api/services/{name}/restart` (admin). Địa chỉ MariaDB từ config mysql host / env `QUICKWIN_MARIADB_ADDR`; Gitea từ env `QUICKWIN_GITEA_ADDR`; app = self port |
| `api/router.go` | +1 dòng | `registerQuickwinServices` sau `registerQuickwinRegistry` |
| `go.mod` | +require/replace | `quickwin.dev/servicemanager` |
| `web/src/views/OpenITMS.vue` | +tab | tab **Services** — data-table (label/status/addr) + nút Restart (admin, service `can_restart`) |

## servicemanager (package)
- **Status**: dial TCP (platform-agnostic — chạy cả Linux lẫn Windows dev). `up`/`down`.
- **Restart**: `systemctl restart <unit>` trên **Linux** (bản cài thật). Nền tảng khác → lỗi rõ
  "chỉ hỗ trợ Linux systemd". `app` (core) không tự restart (`can_restart=false`).
- Service mặc định: mariadb (unit openitms-db), gitea (openitms-gitea), app (openitms).

## Verify
- `apply-patches.sh` chuỗi 0001–0007 sạch + build.
- servicemanager unit test PASS (status up/down qua listener thật; restart lỗi đúng non-Linux).
- UI compile qua `ui-build` CI. E2E: `GET /api/services` trả status MariaDB/app.
