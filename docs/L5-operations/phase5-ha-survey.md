---
level: L1
status: draft
owners: [maintainer]
updated: 2026-07-06
related-code: []
---

# Phase 5-C: Khảo sát HA/cluster (ghi chú kiến trúc)

Ngoài phạm vi v1 (plan 1.3: single-node trước). Ghi chú để mở sau, KHÔNG làm bây giờ.

## Điểm cần khảo sát khi tới
- DB: MariaDB single-node → replication/Galera; hoặc external managed DB (đổi dialect config).
- State: Semaphore core stateless phần lớn (state ở DB) → scale ngang sau LB được.
- Plugin Manager: mỗi node chạy plugin riêng → cần shared registry (đã có) + coordination.
- Certs: shared cert store (hiện local ./certs) → cần shared volume/secret manager.
- Cần ADR riêng khi quyết định làm HA.
