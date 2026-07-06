---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: [.github/workflows/release.yml, .claude/skills/release]
---

# Runbook: Release

Skill `release` chuẩn bị; maintainer bấm nút cuối (tag ký). Pipeline `release.yml`.

## Các bước
1. Chốt version `quickwin-vX.Y.Z-sem<upstream>`.
2. AI/maintainer chạy skill `release`: build dist native, release note (Conventional Commits),
   SBOM + SHA256SUMS + licenses/.
3. **Checklist trước tag** (publishing-policy): license files trong artifact; gitleaks +
   banned-words sạch; artifact + registry index ký; release note đã đọc lại; mirror synced.
4. **Maintainer tạo tag ký số** `quickwin-v*` → push → `release.yml` chạy: GitHub Release +
   publish registry public (GitHub Pages, cần secret REGISTRY_PRIVATE_KEY) + cập nhật docs.
5. Verify: tải bản release trên máy sạch, cài, login.

## Cấm
AI tự tạo tag/release. Bỏ checklist. Đụng private key (maintainer/CI secret giữ).
