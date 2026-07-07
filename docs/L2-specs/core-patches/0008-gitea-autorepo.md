---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0008-gitea-autorepo.patch, gitea-manager/]
---

# Spec patch 0008 — Auto-repo Gitea khi tạo project

## Mục đích (WHY)
Mỗi project mới tự có 1 git repo local (ADR-0005 / gitea-integration.md). Hook trong
`AddProject`: sau khi tạo project + noneKey → tạo repo Gitea + đăng ký repository.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_gitea.go` | MỚI | `quickwinAutoRepo(store, project, noneKeyID)`: EnsureOrg + CreateRepo(org, slug(name), autoInit) + CreateRepository (GitURL = clone URL có token nhúng, SSHKey None, branch main). `injectToken` chèn token vào URL |
| `api/projects/projects.go` | +1 dòng | gọi `quickwinAutoRepo` sau khi tạo View, trước block Demo |
| `go.mod` | +require/replace | `quickwin.dev/gitea-manager` |

## Cấu hình (env — installer set sau khi provision Gitea)
- `QUICKWIN_GITEA_ADDR` (vd 127.0.0.1:3080), `QUICKWIN_GITEA_TOKEN` (admin token), `QUICKWIN_GITEA_ORG` (default openitms).
- **Non-fatal**: thiếu env / Gitea lỗi → log warning, KHÔNG chặn tạo project (không phải bản cài nào cũng bật Gitea).

## Verify (E2E thật 2026-07-07)
- Gitea bundle chạy (localhost:3080, DB mysql `gitea`), gitea-manager tạo org/repo thật.
- Tạo project "Demo Web App" qua API → tự có repository "Local (Gitea)" →
  `openitms/demo-web-app.git`; repo tồn tại trong Gitea org openitms. Idempotent.
- Chain 0001–0008 apply sạch + build.

## Còn lại
- G-04: seed project "host" mặc định (first-run) quản lý máy host + playbook quản trị.
- G-05: UI xem/link repo local trong project.
