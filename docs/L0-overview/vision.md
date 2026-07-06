---
level: L0
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: []
---

# Tầm nhìn

**OpenITMS-SMB** — nền tảng mã nguồn mở cho **SMB quản lý hạ tầng IT**: tự động hóa
triển khai và chạy lệnh/script xuống Windows 11 + Linux từ một Web UI, fork từ
[Semaphore UI](https://github.com/semaphoreui/semaphore) (MIT).

**Giá trị cốt lõi:** cài trong 10 phút, 1 lệnh, zero-config — không cần biết Docker/Ansible.

**4 trụ cột:** Core sạch (patch mỏng, sync upstream ≤ 48h) · Plugin-first (go-plugin
gRPC/Protobuf, Go + Python, registry ký số) · Zero-config install (toàn binary native,
KHÔNG Docker: core + MariaDB + pwsh) · AI-driven dev (issue → AI dev → gate → người merge).

**Người dùng mục tiêu:** IT admin của doanh nghiệp 10–500 nhân sự, quản lý hỗn hợp
máy trạm Windows 11 + server Linux, ít người, ít thời gian, cần "chạy được ngay".

Bản đầy đủ: `PLAN.md` mục 0 (primary plan 1 trang).
