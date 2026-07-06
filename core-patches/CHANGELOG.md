# core-patches CHANGELOG

Mỗi patch thêm/sửa/xóa phải có 1 entry ở đây (mới nhất lên đầu).
Format: `## <ngày> — <patch-file>` + WHY (vì sao cần) + WHAT (đổi gì, mức cao).

## 2026-07-06 — 0001-plugin-manager-hook.patch
**WHY:** cần 1 điểm hook duy nhất để start Plugin Manager + mount API động /api/plugins/
sau middleware authentication của Semaphore (plugin thừa hưởng authn core).
**WHAT:** file mới `api/quickwin_plugins.go` (keo ~60 dòng); `api/router.go` +1 dòng;
`go.mod` +require/replace `quickwin.dev/*`; `go.sum` generated (conflict → `go mod tidy`).
Sửa tay vào file upstream: **6 dòng** (1 router.go + 5 go.mod). E2E: tests/e2e/plugin-through-core.sh.
