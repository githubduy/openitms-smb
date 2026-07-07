# Phát triển OpenITMS-SMB

Trang này dành cho **lập trình viên** (người & AI) muốn build từ mã nguồn hoặc đóng góp.
Người dùng cuối xem [README.md](README.md).

## Kiến trúc tóm tắt

OpenITMS-SMB là **bản fork của [Semaphore UI](https://github.com/semaphoreui/semaphore)** (Go +
Vue, MIT) theo mô hình **"fork mỏng"**:

- `upstream/` — submodule Semaphore, ghim tag (hiện `v2.18.16`). **KHÔNG sửa trực tiếp.**
- `core-patches/` — mọi thay đổi vào core là các patch có đánh số, apply lên `upstream/`.
- `plugins/`, `winrs-exec/`, `registry/`, `gitea-manager/`, … — các package riêng của OpenITMS.
- `installer/` — bộ cài Linux (`linux/install.sh`) và Windows (`windows/install.ps1`).

Mỗi core patch tuân thủ **"bộ-4"**: file `.patch` + 1 dòng trong `core-patches/series` +
entry trong `core-patches/CHANGELOG.md` + spec trong `docs/L2-specs/core-patches/`.

## Build từ source

```bash
scripts/setup-toolchain.sh   # Go pin vào ./Go/ (không dùng Go hệ thống)
scripts/apply-patches.sh     # apply core-patches/series vào upstream/
scripts/build-all.sh         # → dist/bin/semaphore  (FULL_UI=1 để build cả frontend)
scripts/reset-upstream.sh    # trả upstream/ về sạch
scripts/run-tests.sh         # test các package của OpenITMS
```

Thêm/sửa patch: `scripts/export-patch.sh <NNNN-tên>` (xem hướng dẫn trong chính script).

## Đóng gói bản cài

```bash
installer/fetch-deps.sh      # tải MariaDB + pwsh (pin checksum trong deps.lock)
installer/package.sh         # → dist/openitms-smb-<ver>-linux-amd64.tar.gz
```

Bộ cài Windows: `installer/windows/install.ps1` (+ `uninstall.ps1`).

## Tài liệu

- [PLAN.md](PLAN.md) — kế hoạch tổng (mục 0 = tóm tắt 1 trang)
- [TASKS.md](TASKS.md) — backlog chi tiết
- [docs/](docs/) — tài liệu phân tầng L0→L5 (bắt đầu từ [vision](docs/L0-overview/vision.md))
- **Bắt buộc đọc trước khi code (người & AI):**
  [docs/L3-development/AI-ENGINEER-GUIDELINE.md](docs/L3-development/AI-ENGINEER-GUIDELINE.md)

## Đóng góp

Xem [CONTRIBUTING.md](CONTRIBUTING.md) và [GOVERNANCE.md](GOVERNANCE.md). Tuân thủ
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
