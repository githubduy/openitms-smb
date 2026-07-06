# core-patches CHANGELOG

Mỗi patch thêm/sửa/xóa phải có 1 entry ở đây (mới nhất lên đầu).
Format: `## <ngày> — <patch-file>` + WHY (vì sao cần) + WHAT (đổi gì, mức cao).

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
