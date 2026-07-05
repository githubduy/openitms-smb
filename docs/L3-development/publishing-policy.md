---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [.github/workflows/, .gitignore]
---

# Chính sách publish (Git hosting)

Nguồn: `PLAN.md` mục 9.5. Tóm tắt thực thi:

## Nền tảng
- **GitHub = primary** (Issues, PR, CI, Releases, Pages). **Codeberg (Forgejo) = mirror**
  read-only tự động (main + tags); README mirror ghi "đóng góp tại GitHub".
- Forgejo/Gitea self-host: tùy chọn cho khách air-gapped.

## Cấm publish
Secrets, private key (nhất là khóa ký registry), cert thật, `servers.json`, IP nội bộ,
credentials khách. 3 lớp chặn: GitHub push protection + gitleaks (CI) + `.gitignore`.
**Lỡ push secret = ĐÃ LỘ → rotate ngay, không chỉ xóa commit.**

## Branch & tag
- `main` protected: PR + ≥1 human review + CI xanh; cấm force-push.
- AI chỉ push `ai/*`; người: `feat/* fix/* docs/* chore/*`. Squash-merge PR của AI.
- Tag release `quickwin-v*` ký số, CHỈ maintainer (hoặc CI sau approve) tạo — tag = trigger release.
- Maintainer bật 2FA; admin repo ≥2 và tối thiểu hóa.

## Checklist trước mọi release
1. `LICENSE`, `LICENSE-SEMAPHORE`, `NOTICE.md`, `THIRD_PARTY_LICENSES.md` trong artifact.
2. gitleaks + secret scanning sạch từ release trước.
3. SHA256 + SBOM + chữ ký artifact.
4. Release note đã được người đọc lại.
5. Mirror Codeberg đã sync tag.
