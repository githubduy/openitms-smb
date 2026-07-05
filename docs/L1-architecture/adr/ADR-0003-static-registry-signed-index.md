---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: []
---

# ADR-0003 — Registry = static index ký số (không server động)

**Trạng thái:** approved (2026-07-05)

## Bối cảnh
Cần registry local (air-gapped) + public rẻ, an toàn, dễ vận hành.

## Quyết định
index.json + tarball + chữ ký (cosign/minisign) host tĩnh (GitHub Pages/thư mục local). Client verify chữ ký trước khi cài. Private key chỉ maintainer/CI-secret giữ.

## Phương án đã loại
Loại - registry server động (thêm bề mặt tấn công + chi phí vận hành, SMB không cần).

## Hệ quả
Xem PLAN.md mục 3, 5 và GUIDELINE (luật cứng) — mọi thay đổi ngược quyết định này cần ADR mới supersede.
