---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [.github/workflows/ai-triage.yml, .claude/skills/]
---

# Quy trình AI-driven dev

Issue → AI đọc → AI dev → quality gate → người merge. Cộng đồng nạp yêu cầu, AI phát triển.

## Pipeline
1. **Nạp**: GitHub Issue (Issue Forms: feature/bug/plugin-proposal — field máy-đọc).
2. **Triage** (`ai-triage.yml`, tự động): gắn label `area:*` + `severity:*` theo keyword;
   thiếu repro/acceptance → comment hỏi. Bản keyword chạy ngay, không cần secret.
3. **AI dev** (maintainer bật): issue đủ điều kiện (label `ai-approved` HOẶC bug có repro) →
   spawn **Claude Code headless** trên branch `ai/<issue>-<slug>`, chạy đúng **AI Skill** theo area:
   - `area:core-patch` → skill `dev-core-patch`
   - `area:plugin` → `dev-plugin`
   - `area:registry` → `dev-registry`
   - `area:template` → `dev-template`
   - sync upstream → `sync-upstream` · release → `release`
   (Cần self-hosted runner + `ANTHROPIC_API_KEY` — maintainer cấu hình; AI không tự có.)
4. **Gate** (CI): build + test + smoke + plugin-through-core + patch-hygiene + license +
   gitleaks + banned-words + proto + ui-build + python-plugin + SonarQube (fail-when-red).
5. **Merge**: maintainer review & merge. **AI KHÔNG merge, KHÔNG push main, KHÔNG tag** (luật #9).

## Ràng buộc (mọi AI agent)
Đọc `AI-ENGINEER-GUIDELINE.md` làm ngữ cảnh đầu tiên. 10 luật cứng enforce bởi hook + CI.
Gặp quyết định kiến trúc chưa có ADR → DỪNG, mở issue (escalation).
