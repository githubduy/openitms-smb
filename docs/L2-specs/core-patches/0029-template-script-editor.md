---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-11
related-code: [core-patches/0029-template-script-editor.patch, core-patches/0008-gitea-autorepo.patch, core-patches/0012-view-local-repo.patch]
---

# Spec patch 0029 — Template: tạo/sửa script file trực tiếp → commit vào repo local

## Mục đích (WHY)
SMB IT tạo Task Template phải nhập "script filename" (playbook/bash/ps1/...) trỏ tới file trong
git repo. Nhưng lúc đầu repo trống — user phải clone repo, tạo file, commit, push bằng git bên
ngoài rồi mới quay lại điền tên. Rào cản lớn cho người không rành git.

Patch này cho phép **tạo/sửa file ngay trên UI**: bấm "New/Edit file" cạnh ô filename → editor mở
ra (điền sẵn sample theo loại app khi tạo mới) → **Lưu là commit thẳng vào repo local (Gitea)**.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_gitea_commit.go` | file mới | Hook mỏng: `resolveLocalGiteaRepo` (tìm repository theo id, xác nhận là repo local Gitea, tách org/repo/branch/token từ GitURL); `GetTemplateScriptFile` (GET đọc file để edit); `CommitTemplateScriptFile` (POST tạo/sửa → commit) |
| `api/router.go` | +route | `GET/POST /project/{id}/gitea/file` |
| `web/src/components/TemplateForm.vue` | sửa | Nút "New/Edit file" (append-outer ô playbook); dialog editor (codemirror + mode theo app); `openFileEditor`/`saveFileEditor`/`sampleContentForApp`/`defaultFilenameForApp`; import yaml/shell/python modes + `getErrorMessage` |
| `web/src/lang/en.js` | +key | `scriptEditor*` (Btn/NewTitle/EditTitle/Filename/Content/SaveCommit/NotLocal) |

**Logic thật ngoài upstream** (không phải patch — commit thẳng): `gitea-manager/client.go`
`GetFile`/`PutFile` (Gitea Contents API: POST tạo / PUT update kèm sha; content base64) + test.

## Luồng
1. Form Template, chọn Repository = "Local (Gitea)" → nút "New/Edit file" bật.
2. Bấm nút: `openFileEditor` lấy filename hiện tại (hoặc default theo app: site.yml / script.sh /
   script.ps1 / main.py / main.tf), GET `/gitea/file?repo_id=&path=`:
   - `{local:false}` → cảnh báo "không phải repo local" (nút Lưu tắt).
   - file đã tồn tại → nạp nội dung để sửa; chưa có → điền sample theo app.
3. Sửa trong codemirror → "Save & commit" → POST `/gitea/file` → `PutFile` commit 1 lần →
   set `item.playbook = path`, đóng dialog.

## Bảo mật
- Chỉ commit được vào **repo local (Gitea)** của chính project (verify host==QUICKWIN_GITEA_ADDR
  hoặc tên repo chứa "Gitea"); repo ngoài (GitHub...) bị từ chối với message rõ ràng.
- Token Gitea lấy từ GitURL (nhúng bởi 0008) hoặc env — **không trả ra frontend**.
- Endpoint dưới `projectUserAPI` (đã auth + gắn project context).

## Rebase (khi upstream đổi)
- Neo UI: ô `v-text-field v-model="item.playbook"` trong `TemplateForm.vue`. Nếu upstream đổi field
  này → chèn lại slot `append-outer` + dialog. Backend độc lập upstream (file mới + 1 route).

## Verify
- gitea-manager test: `GetFile` found/not-found; `PutFile` create(POST không sha)→update(PUT kèm sha).
- E2E: tạo file mới vào repo local qua UI → thấy commit trong Gitea; mở lại sửa → PUT update.
- eslint sạch; chain 0001–0029 build.

## Liên quan
- Repo local tự tạo: [[0008-gitea-autorepo]]. Xem repo trong UI: [[0012-view-local-repo]].
