---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0034-repo-file-browser.patch, core-patches/0029-template-script-editor.patch]
---

# Spec patch 0034 — Duyệt file repo local để chọn script + tạo folder/file

## Mục đích (WHY)
Ở ô "script filename" user phải **gõ tay** đường dẫn file, không biết repo đang có gì. Cần **duyệt cây
thư mục git** để chọn file, và **tạo mới thư mục / file** ngay trong lúc duyệt.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_gitea_commit.go` | sửa | `ListTemplateRepoDir` (hook mỏng): GET `?repo_id=&path=` → `{local, path, branch, entries:[{name,path,type}]}` |
| `api/router.go` | +route | `GET /project/{id}/gitea/dir` |
| `web/src/components/TemplateForm.vue` | sửa | Nút Browse (mdi-folder-search) cạnh nút New/Edit; dialog duyệt (list folder/file, `..` lên cấp, click folder→vào, click file→điền `item.playbook`); nút "New folder" (mini-dialog tên → tạo `<dir>/.gitkeep`), "New file" (mở editor 0029 tại thư mục hiện tại); computed `sortedEntries` (folder trước) |
| `web/src/lang/en.js`, `vi.js` | +key | `scriptBrowse*` |

**Logic thật ngoài upstream** (commit thẳng, không patch): `gitea-manager/client.go` `ListDir` +
struct `DirEntry` (Gitea Contents API GET thư mục; 404 → rỗng).

## Luồng
1. Chọn Repository local → nút **Browse** bật → mở dialog, GET `/gitea/dir?path=` (gốc).
2. Click folder → điều hướng vào; `..` → lên cấp; click file → set `item.playbook = path`, đóng.
3. **New folder** → nhập tên → tạo `<path>/<tên>/.gitkeep` qua POST `/gitea/file` (commit) → reload.
4. **New file** → đóng browser, mở editor 0029 với filename mặc định tại thư mục đang xem.

## Bảo mật
- Dùng lại `resolveLocalGiteaRepo` (0029): chỉ duyệt/ghi repo local Gitea của project; repo ngoài →
  `{local:false}`. Token không lộ ra frontend. Endpoint dưới `projectUserAPI` (đã auth).

## Verify
- gitea-manager: `ListDir` (đã có PutFile/GetFile test 0029; ListDir cùng cơ chế Contents API).
- E2E: GET `/gitea/dir` liệt kê README + scripts/playbooks/deploys (từ scaffold 0033); tạo folder →
  xuất hiện; chọn file → filename điền vào form.
- eslint sạch; chain 0001–0034 build.

## Rebase
- Neo UI: slot `append-outer` ô playbook trong `TemplateForm.vue` (0029). Backend độc lập (file +route).

## Liên quan
- Editor + commit + PutFile: [[0029-template-script-editor]].
