---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [core-patches/0001-plugin-manager-hook.patch, plugin-manager/]
---

# Spec patch 0001 — Plugin Manager hook

## Mục đích (WHY — dùng khi rebase)
Một điểm hook DUY NHẤT trong core để:
1. Start Plugin Manager khi service lên (Scan + Start, degrade gracefully khi không có plugins).
2. Mount API động `/api/plugins/` **trên subrouter đã qua middleware `authentication`**
   → plugin thừa hưởng authn/session của Semaphore, không tự mở cửa riêng.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/quickwin_plugins.go` | **MỚI** (~60 dòng) | Keo: env `QUICKWIN_PLUGINS_DIR` (default `plugins`), `pluginmanager.New/Scan/Start`, `quickwinCaller` map `helpers.GetFromContext(r,"user") (*db.User)` → `pluginv1.Caller{UserId,Username,IsAdmin}` |
| `api/router.go` | +1 dòng | `registerQuickwinPlugins(authenticatedAPI)` ngay sau `authenticatedAPI.Use(StoreMiddleware, JSONMiddleware, authentication)` |
| `go.mod` | +6 dòng | require `quickwin.dev/pluginmanager v0.0.0` + replace 3 module `quickwin.dev/*` → path tương đối (`../plugin-manager`, `../proto/gen/go`, `../sdk/go`) |
| `go.sum` | generated | **Conflict khi rebase → đừng đọc diff, chạy `go mod tidy` trong upstream rồi export lại** |

Sửa tay vào file upstream có sẵn: **6 dòng** (đạt tiêu chí hook mỏng < 30 dòng; go.sum là lockfile generated, không tính).

## Hướng dẫn rebase khi upstream đổi
1. Tìm nơi tạo `authenticatedAPI` trong `api/router.go` (dấu hiệu: `PathPrefix(webPath + "api")` + `Use(...authentication...)`) → gọi `registerQuickwinPlugins(authenticatedAPI)` ngay sau `Use`.
2. Nếu helper user đổi (`helpers.GetFromContext` / key `"user"` / struct `db.User`): cập nhật `quickwinCaller` — hợp đồng cần: user id, username, cờ admin.
3. `go mod tidy` → export lại patch bằng `scripts/export-patch.sh`.

## Semantics
- Không có thư mục plugins → log info, core chạy bình thường (AC verify bằng `tests/e2e/smoke.sh`).
- Chưa authenticate → middleware Semaphore chặn từ ngoài (401). Caller không map được → Handler trả 401 (phòng thủ 2 lớp).
- `CoreVersion` để trống (check `min_core_version` tắt) cho tới khi version string release chuẩn semver — TODO gắn ở P4-04.

## Verify
- `scripts/apply-patches.sh` từ upstream sạch PASS; `scripts/build-all.sh` PASS.
- E2E `tests/e2e/plugin-through-core.sh`: core load hello → unauthenticated 401 → login session thật → `info`/`echo` trả đúng kèm `"caller":"admin"` — PASS 2026-07-06.
