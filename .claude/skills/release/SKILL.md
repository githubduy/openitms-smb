---
name: release
description: >
  Đóng gói + phát hành OpenITMS-SMB: version, release note, dist artifact, publish registry,
  cập nhật website. Trigger khi maintainer yêu cầu release. Enforce checklist publishing-policy;
  tag ký số CHỈ maintainer tạo (AI chuẩn bị, không tự tag/release).
---

# Skill: release

Bản thực thi runbook release. Chỉ maintainer bấm nút cuối (tag ký + publish).

## Bước (AI chuẩn bị — maintainer duyệt/thực thi phần ký)
1. Version: `quickwin-vX.Y.Z-sem<upstream>`. Release note từ Conventional Commits (git-cliff/release-please).
2. Build dist (`installer/package.sh`) — toàn binary native, không Docker. SBOM + SHA256 + licenses/.
3. **Checklist publishing-policy trước release:**
   - LICENSE, LICENSE-SEMAPHORE, NOTICE.md, THIRD_PARTY_LICENSES.md trong artifact.
   - gitleaks + secret scanning sạch từ release trước. banned-words sạch.
   - Artifact ký; registry index ký (REGISTRY_PRIVATE_KEY — maintainer/CI secret).
   - Release note đã người đọc lại (không lộ thông tin nội bộ).
   - Mirror đã sync tag.
4. Publish registry public (GitHub Pages) + cập nhật docs/website.
5. **Tag `quickwin-v*` ký số CHỈ maintainer/CI-sau-approve tạo** (trigger pipeline). AI KHÔNG tự tag.

## Cấm
- AI tự tạo tag/release. Bỏ qua checklist. Đụng private key.
