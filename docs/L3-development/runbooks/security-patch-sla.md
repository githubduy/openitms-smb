---
level: L3
status: approved
owners: [maintainer]
updated: 2026-07-06
related-code: []
---

# Runbook: SLA vá bảo mật upstream (≤ 48h)

1. Theo dõi Semaphore security release (GitHub watch/RSS).
2. Có bản vá bảo mật → chạy runbook `sync-upstream` NGAY, ưu tiên cao nhất.
3. Nếu sync sạch → PR + merge + release patch trong ≤ 48h.
4. Nếu conflict phức tạp → maintainer trực tiếp rebase; vẫn giữ mốc 48h.
5. Ghi nhận CVE/bản vá vào release note.
