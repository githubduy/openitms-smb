---
level: L1
status: approved
owners: [maintainer]
updated: 2026-07-05
related-code: [upstream/, .github/workflows/]
---

# ADR-0004 — Các quyết định khởi động dự án

**Trạng thái:** approved (chốt bởi maintainer, 2026-07-05)

| # | Quyết định | Chốt | Hệ quả |
|---|---|---|---|
| 1 | Tên sản phẩm + domain | **CHƯA CHỐT** — "QuickWin" là tên tạm | Chặn P1-07 (branding) + URL registry public. Đổi tên sau = sửa patch 0002 + rename repo (chi phí thấp, đã thiết kế cô lập) |
| 2 | Baseline upstream | **v2.18.16** (tag stable mới nhất tại ngày init; dòng 2.19.x còn beta) | Submodule `upstream/` pin tag này; mọi patch viết trên baseline này |
| 3 | Registry public | **GitHub Pages** (static index + artifact, ký số) | CI publish lên nhánh gh-pages / Pages workflow |
| 4 | Chất lượng code | **SonarQube self-host** | Cần dựng + vận hành instance (task P4-03); CI runner phải reach được instance |
| 5 | Live USB base | **Alpine**, control từng package (danh sách tường minh, có review) | musl libc → binary bundle phải build musl-compatible/static; pwsh trên musl phải kiểm chứng sớm (P5-B) |

Bối cảnh & các phương án đã loại: xem `PLAN.md` mục 13.
