---
level: L0
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [LICENSE, LICENSE-SEMAPHORE, NOTICE.md]
---

# Pháp lý — tóm tắt cho người phân phối

1. **MIT (Semaphore):** mọi bản phân phối (source/binary) PHẢI kèm `LICENSE-SEMAPHORE`
   (nguyên văn, không sửa) + `NOTICE.md`. CI nhúng `licenses/` vào mọi artifact.
2. **Trademark:** không dùng tên/logo "Semaphore" trong tên sản phẩm/domain/logo.
   Được ghi "based on / fork of Semaphore UI". Trang About giữ attribution.
3. **MariaDB (GPLv2):** phân phối dạng binary độc lập, IPC qua socket — mere aggregation,
   KHÔNG static-link. Kèm GPLv2 text + link source trong `licenses/`.
4. **Ansible (GPLv3):** chỉ gọi qua exec process (như upstream). Nếu bundle (Live USB):
   kèm GPLv3 text + offer source.
5. **Dependency mới:** CI chạy `go-licenses check` — GPL/AGPL link trực tiếp = build fail.
6. Code mới của dự án: MIT (`LICENSE`).

Chi tiết đầy đủ: `PLAN.md` mục 2.
