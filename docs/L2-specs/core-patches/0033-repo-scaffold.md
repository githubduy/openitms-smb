---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0033-repo-scaffold.patch, core-patches/0008-gitea-autorepo.patch, core-patches/0029-template-script-editor.patch]
---

# Spec patch 0033 — Auto-repo: scaffold README + folder mẫu

## Mục đích (WHY)
Repo local tạo tự động (0008) chỉ có README rỗng do `auto_init` → user mở ra thấy trống, không biết
đặt script ở đâu. Scaffold sẵn **README hướng dẫn + folder mẫu** (scripts/playbooks/deploys) để định
hướng ngay.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_gitea.go` | sửa | Sau `CreateRepo` gọi `scaffoldRepo(c, ctx, org, slug, projectName)`: `PutFile` README.md (giới thiệu + cấu trúc gợi ý) + `scripts/README.md` + `playbooks/README.md` + `deploys/README.md` |

- Dùng `PutFile` (thêm ở 0029, package giteamanager ngoài upstream) → không thêm phụ thuộc mới.
- Git không track folder rỗng → mỗi folder chứa 1 `README.md` để folder tồn tại + có mô tả.
- **Non-fatal**: mỗi file lỗi chỉ `log.Warn`, không chặn tạo project (giữ nguyên tinh thần 0008).
- Branch "main" (khớp branch repo đăng ký ở 0008).

## Verify (E2E)
- Tạo project mới → repo local có: `README.md`, `scripts/README.md`, `playbooks/README.md`,
  `deploys/README.md` (4 commit scaffold).
- Project cũ không bị ảnh hưởng (scaffold chỉ chạy lúc tạo repo).
- `go build ./api/projects/` OK; chain 0001–0033 build.

## Rebase
- Neo: hàm `quickwinAutoRepo` trong `quickwin_gitea.go` (0008). Nếu 0008 đổi luồng tạo repo → gọi lại
  `scaffoldRepo` sau khi có repo + trước/sau đăng ký repository.

## Liên quan
- Auto-repo: [[0008-gitea-autorepo]]. PutFile: [[0029-template-script-editor]].
