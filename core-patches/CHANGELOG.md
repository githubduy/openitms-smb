# core-patches CHANGELOG

Mỗi patch thêm/sửa/xóa phải có 1 entry ở đây (mới nhất lên đầu).
Format: `## <ngày> — <patch-file>` + WHY (vì sao cần) + WHAT (đổi gì, mức cao).

## 2026-07-07 — 0007-service-manager.patch
**WHY:** trang Admin xem status + restart service hạ tầng (MariaDB/Gitea/app).
**WHAT:** api/quickwin_services.go (/api/services list+restart); router +1; go.mod +servicemanager;
OpenITMS.vue +tab Services. Status via TCP (mọi OS); restart via systemctl (Linux). Spec 0007.

## 2026-07-06 — 0006-management-ui.patch
**WHY:** UI quản lý OpenITMS (Plugins/Registry/Hardening) — tính năng core theo plan.
**WHAT:** view mới OpenITMS.vue (3 tab, axios gọi /api/plugins + /api/registry + hardening);
router +route /openitms; App.vue +nav item. Thuần frontend. Verify: ui-build CI (cần node).

## 2026-07-06 — 0005-registry-client.patch
**WHY:** registry client = ngoại lệ core được duyệt (ADR-0003), hạ tầng plugin/template phụ thuộc.
**WHAT:** file mới api/quickwin_registry.go (search + install, verify sig+checksum, unpack plugin);
router.go +1 dòng; go.mod +require/replace quickwin.dev/registry. E2E: registry-through-core.sh.

## 2026-07-06 — 0003-default-password-banner.patch
**WHY:** admin mặc định admin/quickwin123 (yêu cầu gốc) → ép nhắc đổi mật khẩu lần đầu (plan 7.3).
**WHAT:** App.vue thêm 1 v-alert (banner cam) + computed showDefaultPasswordWarning +
data defaultPasswordDismissed (localStorage) + 2 method. Chỉ frontend, không đụng backend.
Spec: 0003-default-password-banner.md.

## 2026-07-06 — 0002-branding.patch
**WHY:** MIT không cấp quyền trademark — thay tên/logo "Semaphore" bằng OpenITMS-SMB (ADR-0004 #1).
**WHAT:** title index.html; logo.svg + favicon.svg (SVG text); key i18n `ansibleSemaphore`
→ 'OpenITMS-SMB' (5 file lang). Asset-replace thuần, không đổi code. Phần cần node
(favicon.png, chuỗi i18n dài, trang About attribution) → P1-07b. Spec: 0002-branding.md.

## 2026-07-06 — 0001-plugin-manager-hook.patch
**WHY:** cần 1 điểm hook duy nhất để start Plugin Manager + mount API động /api/plugins/
sau middleware authentication của Semaphore (plugin thừa hưởng authn core).
**WHAT:** file mới `api/quickwin_plugins.go` (keo ~60 dòng); `api/router.go` +1 dòng;
`go.mod` +require/replace `quickwin.dev/*`; `go.sum` generated (conflict → `go mod tidy`).
Sửa tay vào file upstream: **6 dòng** (1 router.go + 5 go.mod). E2E: tests/e2e/plugin-through-core.sh.
