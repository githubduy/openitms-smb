---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [installer/, core-patches/]
---

# ADR-0005 — Bundle Gitea làm git server local mặc định

**Trạng thái:** approved (2026-07-07)

## Bối cảnh
Yêu cầu: mỗi khi tạo dự án mới trong OpenITMS-SMB, mặc định có sẵn **1 local git
repository** — không phải trỏ tới git bên ngoài. Kèm 1 project mặc định quản lý chính
máy host. Đúng triết lý "bundle mọi thứ, zero-config".

Semaphore gốc: mỗi project có "repository" (URL git bên ngoài) làm nguồn playbook. Ta cần
1 git server **tích hợp sẵn** để repo local tự sinh theo project.

## Quyết định
Bundle **Gitea** làm git server local mặc định (binary native trong dist, như MariaDB/pwsh).
- **Gitea = MIT license, single Go binary** → tương thích MIT của dự án, bundle hợp lệ, không
  Docker (đúng luật #5). Nhẹ (~256MB RAM).
- Chạy local (port nội bộ, vd 3080), data trong thư mục cài của tool.
- DB: dùng chung **MariaDB đã bundle** (database `gitea` riêng) cho nhất quán; fallback SQLite
  của Gitea cho bản Live USB minimal.
- Auto-provision lúc cài: admin Gitea, org mặc định `openitms`.

## Phương án đã cân nhắc
| Phương án | Ưu | Nhược | Vì sao (không) chọn |
|---|---|---|---|
| **Gitea (MIT)** | single binary, MIT, nhẹ, API đầy đủ | thêm 1 service | ✅ CHỌN — MIT bundle-friendly |
| Forgejo (fork Gitea) | cộng đồng | **GPLv3** → phức tạp khi đóng gói | loại (GPL) |
| GitLab | đầy đủ | rất nặng, cần Docker/Ruby | loại (nặng, trái luật #5) |
| Chỉ `git` bare repo + git-http-backend | siêu nhẹ | không UI/API/PR/quyền | loại (thiếu quản lý repo) |

## Tích hợp (core sạch — qua hook + integration)
1. **Bundle**: gitea binary vào `installer/vendor/gitea/` → dist (deps.lock pin sha256).
2. **Gitea Manager** (package Go ngoài core, như plugin-manager): start Gitea, health-check,
   provision admin + org lần đầu, expose API client tạo repo.
3. **Hook tạo project** (core-patch): khi project mới tạo trong OpenITMS → gọi Gitea API tạo
   repo `openitms/<project-slug>` → set repository của project trỏ tới `http://localhost:3080/...`.
   Repo hiện dưới project (khớp yêu cầu `/project/<id>/repositories`).
4. **Project mặc định "host"** (id 1): seed lúc cài — quản lý chính máy host, kèm repo riêng
   (chứa playbook/script quản trị host).

## Hệ quả
- Thêm 1 binary bundle + 1 service (systemd `openitms-gitea` after mariadb).
- Cần core-patch hook project-create (mỏng) + Gitea Manager (ngoài core).
- Zero-config: người dùng tạo project → có ngay repo local, không cấu hình git.
- Chi tiết thiết kế: `docs/L1-architecture/gitea-integration.md`.
