---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: []
---

# ADR-0002 — MariaDB bundled, chỉ listen unix socket/localhost

**Trạng thái:** approved (2026-07-05)

## Bối cảnh
Cần DB performance tốt đóng gói sẵn, zero-config, không vi phạm GPLv2.

## Quyết định
Bundle MariaDB LTS binary độc lập (mere aggregation), init tự động, KHÔNG mở TCP ra ngoài, driver go-sql-driver/mysql (MPL-2.0) - không link GPL.

## Phương án đã loại
Loại - BoltDB mặc định upstream (không đủ performance/công cụ cho SMB); PostgreSQL (tốt nhưng đã chốt yêu cầu MariaDB); SQLite (giới hạn concurrent).

## Hệ quả
Xem PLAN.md mục 3, 5 và GUIDELINE (luật cứng) — mọi thay đổi ngược quyết định này cần ADR mới supersede.
