---
level: L2
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [core-patches/0005-registry-client.patch, registry/]
---

# Spec patch 0005 — Registry client hook

## Mục đích (WHY)
Registry client là 1 trong 2 ngoại lệ core được duyệt (ADR-0003) — hạ tầng phân phối
mà mọi plugin/template phụ thuộc. Mount API registry trên subrouter đã authenticate.

## Thay đổi
| File | Loại | Nội dung |
|---|---|---|
| `api/quickwin_registry.go` | MỚI | `GET /api/registry/search?q=&type=` (mọi user) + `POST /api/registry/install` (admin — verify sig ed25519 + sha256, unpack plugin vào `QUICKWIN_PLUGINS_DIR`). Source từ env `QUICKWIN_REGISTRY_LOCAL/PUBLIC`, trusted key `QUICKWIN_REGISTRY_PUBKEYS`. Local chỉ thêm khi index.json tồn tại (fresh-install không lỗi) |
| `api/router.go` | +1 dòng | `registerQuickwinRegistry(authenticatedAPI)` sau `registerQuickwinPlugins` |
| `go.mod` | +require/replace | `quickwin.dev/registry` → `../registry`; go.sum generated |

## Hướng dẫn rebase
Chèn `registerQuickwinRegistry(authenticatedAPI)` ngay sau `registerQuickwinPlugins`.
go.sum conflict → `go mod tidy`. Helper `helpersUser` dùng `helpers.GetFromContext(r,"user")`.

## Semantics
- Verify: có trusted key → verify chữ ký index; không có (dev) → bỏ verify. Checksum tarball luôn verify.
- Install plugin → unpack vào plugins dir; plugin manager nhận ở lần scan kế (cần restart/rescan).
- Config mặc định 2 source: local (air-gapped) + public (nếu set env).

## Verify
- `apply-patches.sh` chuỗi 0001–0005 PASS; build PASS.
- E2E `tests/e2e/registry-through-core.sh`: build registry local (registryctl) → core search
  thấy artifact → registry mounted trong log. PASS 2026-07-06.
