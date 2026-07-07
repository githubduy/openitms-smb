---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-07
related-code: [core-patches/0011-seed-host-project.patch, core-patches/0008-gitea-autorepo.patch, core-patches/0009-winrs-app.patch]
---

# Spec patch 0011 — Seed project "Host" mặc định (G-04)

## Mục đích (WHY)
Yêu cầu gốc: "mặc định sẽ có 1 project quản lý chỉnh máy chính host". Sau khi cài, người dùng
nên có sẵn 1 project trỏ vào chính máy đang chạy OpenITMS — không phải tự tạo project + inventory
từ đầu để bắt đầu quản trị host.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/projects/quickwin_seed.go` | MỚI | `SeedHostProjectIfEmpty(store, accessKeyService)`: nếu `GetAllProjects()` rỗng và có ≥1 user → tạo project "Host — máy chính OpenITMS" + gán owner (admin đầu tiên) + None key + view "All" + `quickwinAutoRepo` (repo Gitea, patch 0008) + inventory WinRS "Máy host (WinRS)" trỏ `127.0.0.1 cert=host.pem` |
| `cli/cmd/root.go` | +import +1 dòng | gọi `projects.SeedHostProjectIfEmpty(store, accessKeyService)` trong `runService` sau khi tạo `accessKeyService` (DB đã migrate qua `createStore`) |

## Hành vi
- **Idempotent**: chỉ seed khi CHƯA có project nào. Đã có ≥1 project → return ngay (rẻ, 1 query/lần start).
- **Cần user trước**: nếu chưa có user (setup chưa xong) → bỏ qua, seed ở lần start sau khi đã có admin.
- Chọn owner = admin đầu tiên; không có admin → user[0].
- Mọi bước lỗi → log warning, KHÔNG chặn server start (non-fatal).

## Thứ tự cài (lưu ý)
- **install.sh / CLI**: admin tạo trước khi start server (`semaphore user add`) → lần start đầu đã có
  user, 0 project → seed chạy ngay. (Đường chuẩn.)
- **Web-setup-first**: nếu admin tạo qua wizard SAU khi server đã start, seed lần đầu bỏ qua (0 user);
  Host project xuất hiện sau lần restart kế tiếp. Chấp nhận được.

## Verify
- `go build ./cli/... ./api/...` OK.
- Skip-path: start app khi đã có project → không tạo trùng, không crash (log skip).
- Create-path: cần DB trống (fresh install) → seed tạo Host project + inventory WinRS localhost.
- Chain 0001–0011 apply sạch + build.

## Liên quan
- Auto-repo Gitea: [[0008-gitea-autorepo]]. Inventory/engine WinRS: [[0009-winrs-app]].
