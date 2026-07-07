---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0012-view-local-repo.patch, core-patches/0008-gitea-autorepo.patch]
---

# Spec patch 0012 — Xem repo local (Gitea) per project (G-05)

## Mục đích (WHY)
Patch 0008 tự tạo repo Gitea local cho mỗi project và đăng ký làm repository (GitURL có token
nhúng để clone HTTP không cần nhập mật khẩu). Token là secret → không hiện thẳng GitURL ra UI.
G-05: cho người dùng thấy + mở repo local của project (link tới Gitea web) mà KHÔNG lộ token.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_gitea_view.go` | MỚI | `GetProjectGiteaRepo`: duyệt repository của project, chọn repo local (host URL khớp `QUICKWIN_GITEA_ADDR` hoặc tên chứa "Gitea"), `u.User = nil` để **strip token/userinfo**, trả `{repo_id,name,branch,web_url}` (web_url đã bỏ `.git`); không có repo local → `{web_url:null}` |
| `api/router.go` | +1 route | `GET /project/{id}/gitea/repo` trên `projectUserAPI` |
| `web/src/views/project/Repositories.vue` | +banner +load | `created()` gọi `/gitea/repo`; banner "Repo local (Gitea): <name> <branch>" + nút "Mở trên Gitea" (`:href=web_url` target _blank rel noopener) |

## Bảo mật
- **Không lộ token**: handler strip `url.User` trước khi trả về; frontend chỉ nhận web_url sạch.
- Route gate `CanManageProjectResources` (projectUserAPI).
- Nút mở dùng `rel="noopener noreferrer"` + `target="_blank"`.

## Verify
- `go build ./api/...` OK; eslint Repositories.vue sạch.
- Chain 0001–0012 apply sạch + build.
- E2E: project có repo Gitea → banner hiện web_url không chứa token; project không có → không banner.

## Liên quan
- Auto-repo tạo repository local: [[0008-gitea-autorepo]].
