# OpenITMS-SMB

Nền tảng mã nguồn mở cho **SMB quản lý hạ tầng IT**: tự động hóa triển khai, chạy lệnh/script
xuống **Windows 11 + Linux** từ một Web UI. Fork của
[Semaphore UI](https://github.com/semaphoreui/semaphore) (MIT) — xem [NOTICE.md](NOTICE.md).

> **Trạng thái: Phase 0 — đang dựng nền móng.** Chưa có bản phát hành.

**Cài trong 10 phút, 1 lệnh, zero-config.** Toàn binary native (core + MariaDB + PowerShell
Core đóng gói sẵn) — **không Docker**, không bước cấu hình.

## Tài liệu
- [PLAN.md](PLAN.md) — kế hoạch tổng (mục 0 = tóm tắt 1 trang)
- [TASKS.md](TASKS.md) — backlog chi tiết
- [docs/](docs/) — tài liệu L0→L5; bắt đầu từ [vision](docs/L0-overview/vision.md)
- **Tham gia phát triển (người & AI):** đọc [docs/L3-development/AI-ENGINEER-GUIDELINE.md](docs/L3-development/AI-ENGINEER-GUIDELINE.md) TRƯỚC TIÊN

## Build từ source (dev)
```bash
scripts/setup-toolchain.sh   # Go pin vào ./Go/ (không dùng Go hệ thống)
scripts/apply-patches.sh     # apply core-patches/series vào upstream/
scripts/build-all.sh         # → dist/bin/semaphore
scripts/reset-upstream.sh    # trả upstream/ về sạch
```

## Giấy phép
MIT — [LICENSE](LICENSE). Chứa mã của Semaphore UI © Denis Gukov / Castaway Labs LLC,
MIT — [LICENSE-SEMAPHORE](LICENSE-SEMAPHORE). Không liên kết/không được chứng thực bởi Semaphore UI.
