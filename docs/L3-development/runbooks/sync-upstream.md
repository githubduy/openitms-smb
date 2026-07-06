---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [scripts/sync-upstream.sh, .claude/skills/sync-upstream]
---

# Runbook: Sync upstream

Skill `sync-upstream`. Mục tiêu: bản vá bảo mật Semaphore → ≤ 48h.

## Các bước
1. `scripts/sync-upstream.sh <new-tag>` — fetch + checkout + apply series + build + test.
2. Patch conflict → đọc `# WHY:` header + spec L2 → rebase GIỮ Ý ĐỊNH (điểm hook tương đương,
   không dán nguyên dòng cũ). go.sum conflict → `go mod tidy` (đừng đọc diff).
3. Full build + test + SonarQube. Rebase làm **ĐỔI HÀNH VI** → DỪNG, mở issue hỏi maintainer.
4. Bump baseline: submodule gitlink mới + cập nhật ADR-0004 (#2) + NOTICE.md.
5. PR liệt kê: patch giữ nguyên / rebase / diff hành vi. Maintainer merge.

## Diễn tập (2026-07-06)
`sync-upstream.sh v2.18.16` PASS end-to-end; tag không tồn tại → exit 128, lỗi rõ.
