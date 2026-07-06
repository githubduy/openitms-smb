---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [plugin-manager/]
---

# Spec: Plugin Manager

Package `plugin-manager/` (module `quickwin.dev/pluginmanager`) — chạy NGOÀI cây upstream,
được hook vào core bằng patch 0001 (Phase 1). Chưa hook vẫn dùng standalone được (test/dev).

## Lifecycle
1. **Scan(dir)**: mỗi thư mục con có `plugin.yaml` = 1 plugin ứng viên.
   - Validate manifest theo JSON Schema (embedded `schema/plugin.schema.json`).
   - `protocol_version` phải = APP-PROTOCOL-VERSION của core (hiện 1).
   - `min_core_version` so semver với core — không đạt → bỏ qua.
   - Checksum: có khai → verify sha256 từng file, lệch → bỏ qua plugin; KHÔNG khai →
     warning, chỉ chấp nhận plugin dev local (registry Phase 3 sẽ bắt buộc).
   - Plugin hỏng bị bỏ qua + log error — KHÔNG chặn plugin khác, core không plugin vẫn chạy.
2. **Start**: launch từng plugin qua go-plugin (gRPC + mTLS tự động, handshake SDK):
   - `GetMetadata` ngay sau launch; **name/version/routes phải khớp manifest** — lệch → kill
     + từ chối load (chống manifest nói dối).
   - Entrypoint chọn theo `GOOS-GOARCH`; fallback `python` (cần python3 trên PATH).
   - stdout/stderr + hclog của plugin ghi vào `<plugin-dir>/plugin.log`.
3. **Health loop** (mỗi plugin 1 goroutine): interval 30s (config), timeout 5s/lần.
   - UNHEALTHY hoặc lỗi gRPC × 3 lần liên tiếp (config) → **restart**: kill → backoff
     (1s/5s/25s, lặp mức cuối) → launch lại. DEGRADED → log warning, không restart.
4. **Stop**: kill toàn bộ, chờ goroutine thoát.

## API động
`Manager.Handler(caller CallerFunc)` → `http.Handler` mount tại `/api/plugins/`:
- Path: `/api/plugins/<name>/<route...>`; match method + route pattern từ MANIFEST
  (segment literal + `{param}`).
- Route không khai trong manifest → **404**. Plugin không running → **503**.
- `CallerFunc` do core cung cấp (patch 0001) từ session đã authenticate; nil → **401**;
  route `require_admin` mà caller không admin → **403**.
- Body limit 10MB; timeout gọi plugin 30s; lỗi gRPC → **502**.

## Bảo mật
- Plugin là process riêng — crash không sập core; go-plugin tự cấp mTLS.
- Permissions trong manifest: hiện mới khai báo + hiển thị; enforce từng quyền cụ thể
  (certs:read, inventory:*) nối vào core ở patch 0001+ khi có API core tương ứng.

## Test (AC P1-04)
- `manifest_test.go`: schema bắt 6 case lỗi chuẩn.
- `integration_test.go`: build plugin hello THẬT → handshake go-plugin thật → API động
  (200/404/404) → RunTask stream (≥3 log + result SUCCESS) → kill process → tự restart
  với pid mới trong ≤10s.
